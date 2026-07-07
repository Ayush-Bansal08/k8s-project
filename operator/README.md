# PreviewEnvironment Operator

A custom Kubernetes operator (built with **Kubebuilder / Go**) that provisions
**ephemeral per-PR preview environments**. Apply a `PreviewEnvironment` and the
operator creates a namespace, a Deployment (from the requested image), a Service,
and an Ingress (`pr-<N>.localtest.me`). Delete it and a **finalizer** cleans it all up.

---

## 1. How this operator was created (one-time scaffolding)

Run from **inside** the `operator/` folder:

```bash
# Initialize the Kubebuilder project (creates the skeleton)
kubebuilder init --domain myproject.io --repo github.com/Ayush-Bansal08/preview-operator

# Create the custom API type + controller (answer 'y' to both prompts)
kubebuilder create api --group platform --version v1 --kind PreviewEnvironment
```

This generated the two files we actually edit:
- `api/v1/previewenvironment_types.go` -> the **CRD** ("what a request looks like")
- `internal/controller/previewenvironment_controller.go` -> the **Controller** ("what to do")

---

## 2. Develop & run locally (day-to-day workflow)

Run from inside `operator/` after editing the code:

```bash
make manifests        # regenerate CRD + RBAC YAML from the Go code/markers
make generate         # regenerate helper (deepcopy) code
make install          # install the CRD into the cluster (teach K8s the new type)
make run              # compile & run the operator locally (watches the cluster)
```

> `make run` runs the operator on your machine using your kubeconfig.
> Leave it running — it's the operator "on duty". Re-run it after code changes.

---

## 3. Test it

In a second terminal:

```bash
# Create a preview environment
kubectl apply -f preview-sample.yaml

# See everything the operator built
kubectl get all,ingress -n preview-pr-42

# Visit the preview's URL
curl -I http://pr-42.localtest.me

# Delete it -> the finalizer cleans up the whole namespace
kubectl delete -f preview-sample.yaml
```

`preview-sample.yaml`:
```yaml
apiVersion: platform.myproject.io/v1
kind: PreviewEnvironment
metadata:
  name: pr-42
spec:
  prNumber: 42
  image: nginx:1.27
```

---

## 4. Deploy in-cluster (production style) — optional

Instead of `make run`, package the operator as an image and run it as a pod:

```bash
make docker-build IMG=preview-operator:latest
kind load docker-image preview-operator:latest --name idp   # load image into kind
make deploy IMG=preview-operator:latest                     # deploy as a Deployment
```

---

## What each field does

| Field       | Purpose                                              |
|-------------|------------------------------------------------------|
| `prNumber`  | The PR number; used to name the namespace + URL      |
| `image`     | Container image to deploy (defaults to `nginx:1.27`) |
