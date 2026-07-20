#!/usr/bin/env bash
# One-command rebuild of the entire IDP platform.
# Usage:  bash setup.sh
set -uo pipefail

CLUSTER_NAME="idp"
REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"

echo "==> [1/6] Deleting any existing '$CLUSTER_NAME' cluster..."
kind delete cluster --name "$CLUSTER_NAME" || true

echo "==> [2/6] Creating a fresh cluster from infra/kind-config.yaml..."
kind create cluster --config "$REPO_ROOT/infra/kind-config.yaml"

echo "==> [3/6] Installing the ingress-nginx controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller --timeout=180s || true

echo "==> [4/6] Installing ArgoCD..."
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
kubectl apply --server-side -n argocd \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n argocd || true

echo "==> [4b] Installing metrics-server (for kubectl top + autoscaling)..."
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl patch -n kube-system deployment metrics-server --type=json \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]' || true

echo "==> [5/6] Deploying apps via ArgoCD (bootstrap)..."
kubectl apply -f "$REPO_ROOT/bootstrap/"

echo "==> [6/6] Installing the monitoring stack (Prometheus + Grafana)..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null 2>&1 || true
helm repo update >/dev/null
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace

echo "==> Installing Loki + Promtail (log aggregation)..."
helm repo add grafana https://grafana.github.io/helm-charts >/dev/null 2>&1 || true
helm repo update >/dev/null
helm upgrade --install loki grafana/loki-stack \
  --namespace monitoring --create-namespace \
  --set promtail.enabled=true --set loki.persistence.enabled=false

echo "==> [7/7] Installing the security stack (Kyverno + Falco)..."
helm repo add kyverno https://kyverno.github.io/kyverno/ >/dev/null 2>&1 || true
helm repo add falcosecurity https://falcosecurity.github.io/charts >/dev/null 2>&1 || true
helm repo update >/dev/null
helm upgrade --install kyverno kyverno/kyverno --namespace kyverno --create-namespace
helm upgrade --install falco falcosecurity/falco --namespace falco --create-namespace \
  --set driver.kind=modern_ebpf --set tty=true

echo "==> Applying Kyverno security policies..."
kubectl wait --for=condition=available --timeout=180s \
  deployment/kyverno-admission-controller -n kyverno || true
kubectl apply -f "$REPO_ROOT/security/" || true

echo "==> Installing Chaos Mesh (resilience testing)..."
helm repo add chaos-mesh https://charts.chaos-mesh.org >/dev/null 2>&1 || true
helm repo update >/dev/null
helm upgrade --install chaos-mesh chaos-mesh/chaos-mesh \
  --namespace chaos-mesh --create-namespace \
  --set chaosDaemon.runtime=containerd \
  --set chaosDaemon.socketPath=/run/containerd/containerd.sock

echo ""
echo "======================================================================"
echo " DONE! Your platform is rebuilt. Access it with these port-forwards:"
echo ""
echo "  Shop:      http://shop.localtest.me   (no port-forward needed)"
echo "  ArgoCD:    kubectl port-forward svc/argocd-server    -n argocd    8080:443"
echo "             -> https://localhost:8080  (user: admin)"
echo "  Grafana:   kubectl port-forward svc/monitoring-grafana -n monitoring 3000:80"
echo "             -> http://localhost:3000   (user: admin / prom-operator)"
echo "======================================================================"
