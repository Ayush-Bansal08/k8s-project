/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	platformv1 "github.com/Ayush-Bansal08/preview-operator/api/v1"
)

// PreviewEnvironmentReconciler reconciles a PreviewEnvironment object
type PreviewEnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=platform.myproject.io,resources=previewenvironments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.myproject.io,resources=previewenvironments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.myproject.io,resources=previewenvironments/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile runs every time a PreviewEnvironment is created or changed.
// Its job: make the cluster match what the PreviewEnvironment asks for.
func (r *PreviewEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 1. Fetch the PreviewEnvironment that triggered this call (so we can read its prNumber).
	var preview platformv1.PreviewEnvironment
	if err := r.Get(ctx, req.NamespacedName, &preview); err != nil {
		// If it's gone (deleted), there's nothing to do — just stop cleanly.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Work out the namespace name from the PR number, e.g. "preview-pr-42".
	nsName := fmt.Sprintf("preview-pr-%d", preview.Spec.PRNumber)

	// The finalizer marks "don't fully delete this PreviewEnvironment until I've cleaned up".
	const previewFinalizer = "platform.myproject.io/finalizer"

	// Is this PreviewEnvironment being deleted?
	if !preview.DeletionTimestamp.IsZero() {
		// Yes — if our finalizer is still on it, run cleanup first.
		if controllerutil.ContainsFinalizer(&preview, previewFinalizer) {
			// CLEANUP: delete the namespace (which removes the app inside it too).
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
			if err := r.Delete(ctx, ns); err != nil && !apierrors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			log.Info("Cleaned up preview namespace", "namespace", nsName)

			// Remove the finalizer so Kubernetes can finish deleting the object.
			controllerutil.RemoveFinalizer(&preview, previewFinalizer)
			if err := r.Update(ctx, &preview); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil // stop — nothing to build for something being deleted
	}

	// Not being deleted: make sure our finalizer is attached, so future deletes trigger cleanup.
	if !controllerutil.ContainsFinalizer(&preview, previewFinalizer) {
		controllerutil.AddFinalizer(&preview, previewFinalizer)
		if err := r.Update(ctx, &preview); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Small helper: create the object only if it doesn't already exist (idempotent).
	ensure := func(obj client.Object, kind string) error {
		if err := r.Create(ctx, obj); err != nil {
			if apierrors.IsAlreadyExists(err) {
				return nil // already there — fine
			}
			return err
		}
		log.Info("Created "+kind, "namespace", nsName)
		return nil
	}

	labels := map[string]string{"app": "preview-app"}

	// Decide which image to deploy: from the request, or fall back to nginx.
	image := preview.Spec.Image
	if image == "" {
		image = "nginx:1.27"
	}

	// 3. The namespace to hold this preview.
	if err := ensure(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: nsName},
	}, "namespace"); err != nil {
		return ctrl.Result{}, err
	}

	// 4. The Deployment — runs the app image inside the namespace.
	replicas := int32(1)
	if err := ensure(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "preview-app", Namespace: nsName},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "app",
						Image: image, // <-- from the request's spec!
						Ports: []corev1.ContainerPort{{ContainerPort: 80}},
					}},
				},
			},
		},
	}, "deployment"); err != nil {
		return ctrl.Result{}, err
	}

	// 5. The Service — a stable in-cluster address for the app.
	if err := ensure(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "preview-app", Namespace: nsName},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Port:       80,
				TargetPort: intstr.FromInt(80),
			}},
		},
	}, "service"); err != nil {
		return ctrl.Result{}, err
	}

	// 6. The Ingress — a public URL for this preview, e.g. pr-42.localtest.me.
	pathType := networkingv1.PathTypePrefix
	ingressClass := "nginx"
	host := fmt.Sprintf("pr-%d.localtest.me", preview.Spec.PRNumber)
	if err := ensure(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: "preview-app", Namespace: nsName},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClass,
			Rules: []networkingv1.IngressRule{{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: "preview-app",
									Port: networkingv1.ServiceBackendPort{Number: 80},
								},
							},
						}},
					},
				},
			}},
		},
	}, "ingress"); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Preview environment ready", "namespace", nsName, "url", "http://"+host)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PreviewEnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1.PreviewEnvironment{}).
		Named("previewenvironment").
		Complete(r)
}
