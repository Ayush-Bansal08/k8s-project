# ☸️ Internal Developer Platform on Kubernetes

> A production-grade platform that runs **any** containerized microservice app —
> auto-deploying it via GitOps, monitoring it, securing it, scaling it, and
> **extending Kubernetes itself** with a custom Operator.

**🌍 Live demo:** [https://ayush-idp.duckdns.org](https://ayush-idp.duckdns.org) — running on a cloud k3s cluster with real Let's Encrypt TLS 🔒
*(the demo VM is stopped when idle to control cost — happy to bring it up on request)*

![Kubernetes](https://img.shields.io/badge/Kubernetes-k3s%20%7C%20kind-326CE5?logo=kubernetes&logoColor=white)
![GitOps](https://img.shields.io/badge/GitOps-ArgoCD-EF7B4D?logo=argo&logoColor=white)
![Go](https://img.shields.io/badge/Operator-Go%20%2F%20Kubebuilder-00ADD8?logo=go&logoColor=white)
![Observability](https://img.shields.io/badge/Observability-Prometheus%20%7C%20Grafana%20%7C%20Loki-E6522C?logo=prometheus&logoColor=white)
![Security](https://img.shields.io/badge/DevSecOps-Trivy%20%7C%20Kyverno%20%7C%20Falco-4B32C3)
![Cloud](https://img.shields.io/badge/Cloud-Azure%20%7C%20Let's%20Encrypt-0078D4?logo=microsoftazure&logoColor=white)

---

## What this is

Most Kubernetes projects stop at *"I deployed an app."* This one is the **platform around the app** — the system a real company builds so developers can ship safely.

The demo workload is **Google's Online Boutique** (an 11-service polyglot microservices app) — deliberately *not written by me*. In platform engineering, **the app is the workload; the platform is the project.** To prove the platform is app-agnostic, a second unrelated app (**Podinfo**) runs on it with **zero platform changes**.

---

## Architecture

```
                          ┌──────────────────────────────┐
   git push ──────────────►  GitHub  (single source of truth)
                          └──────────────┬───────────────┘
                                         │  ArgoCD syncs (GitOps)
                                         ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │                       Kubernetes cluster                          │
   │                                                                   │
   │   Ingress (nginx) ──► Online Boutique · Podinfo · preview envs    │
   │                                                                   │
   │   ┌───────────────┐  ┌───────────────┐  ┌─────────────────────┐  │
   │   │ Observability │  │   Security     │  │  Delivery           │  │
   │   │ Prometheus    │  │  Trivy (scan)  │  │  Argo Rollouts      │  │
   │   │ Grafana       │  │  Kyverno       │  │  canary + metric-   │  │
   │   │ Loki          │  │  Falco         │  │  driven rollback    │  │
   │   └───────────────┘  │  RBAC          │  └─────────────────────┘  │
   │                      └───────────────┘                            │
   │   ┌───────────────────────────────────────────────────────────┐  │
   │   │ ⭐ Custom Operator (Go / Kubebuilder)                       │  │
   │   │    PreviewEnvironment CRD → namespace + app + URL on demand │  │
   │   └───────────────────────────────────────────────────────────┘  │
   │                                                                   │
   │   Resilience: HPA autoscaling · Chaos Mesh · self-healing         │
   └──────────────────────────────────────────────────────────────────┘
        Local: kind          │          Cloud: k3s on Azure + TLS 🔒
```

---

## What it does

| Layer | Capability | Tools |
|---|---|---|
| **Run** | Runs any containerized microservice app | Kubernetes (kind / k3s) |
| **Deploy** | `git push` → auto-deploy, with drift correction & self-heal | **ArgoCD** (GitOps) |
| **Release** | Canary rollouts with **metric-driven auto-rollback** | **Argo Rollouts** + Prometheus analysis |
| **Observe** | Live metrics, dashboards, centralized logs | **Prometheus · Grafana · Loki** |
| **Secure** | Image scanning · admission policy · runtime detection · least privilege | **Trivy · Kyverno · Falco · RBAC** |
| **Scale** | Autoscales under load, proven with load tests | **HPA** + **k6** |
| **Survive** | Chaos experiments prove self-healing | **Chaos Mesh** |
| **Self-service** ⭐ | On-demand ephemeral preview environments | **Custom Operator (Go/Kubebuilder)** |
| **Go live** | Public HTTPS on a cloud cluster | **Azure · k3s · cert-manager · Let's Encrypt** |

---

## ⭐ The standout: a custom Kubernetes Operator

I extended the Kubernetes API with my own resource type and wrote the controller that reconciles it.

Apply this:
```yaml
apiVersion: platform.myproject.io/v1
kind: PreviewEnvironment
metadata:
  name: pr-42
spec:
  prNumber: 42
  image: nginx:1.27
```

…and the operator automatically provisions a **namespace**, a **Deployment** (from the requested image), a **Service**, and an **Ingress** at `pr-42.<domain>` — then **tears it all down** on delete via a **finalizer**.

**Built with:** Go + Kubebuilder · a `PreviewEnvironment` **CRD** · a reconciliation **controller** · **finalizer**-based cleanup · least-privilege **RBAC** markers.

📁 [`operator/`](operator/) · 📖 [`operator/README.md`](operator/README.md)

---

## Automated canary deployments

New versions roll out **gradually** (25% → 50% → 75% → 100%), and the rollout **judges itself**:

```
   deploy new version → 25% traffic → query Prometheus for success rate
        ├─ healthy (≥95%)  → auto-promote to 100%   ✅
        └─ degraded        → auto-rollback           🤖
```

Health-based rollback (`progressDeadlineAbort` + readiness probes) catches crashes; metric-based analysis (`AnalysisTemplate` → Prometheus) catches bad *behaviour*.

📁 [`apps/podinfo-canary/`](apps/podinfo-canary/)

---

## Quick start — one command

The entire platform rebuilds from scratch with a single script:

```bash
bash setup.sh
```

Provisions: kind cluster → ingress-nginx → metrics-server → ArgoCD → all apps (via GitOps) → Prometheus/Grafana → Loki → Kyverno → Falco → Chaos Mesh.

Then start the operator:
```bash
cd operator && make install && make run
```

**Why this matters:** when the local cluster's certificates corrupted, the entire environment was rebuilt in minutes — because every layer is declarative and in Git. *Cattle, not pets.*

---

## Repository structure

```
├── infra/kind-config.yaml      # local cluster, as code
├── apps/
│   ├── online-boutique/        # the 11-service demo workload + HPA
│   ├── podinfo/                # 2nd app — proves the platform is app-agnostic
│   ├── podinfo-canary/         # automated canary + Prometheus analysis
│   └── hello/                  # minimal reference app
├── bootstrap/                  # ArgoCD Applications (GitOps, as code)
├── operator/                   # ⭐ custom Kubernetes Operator (Go/Kubebuilder)
├── security/                   # Kyverno policies + RBAC
├── chaos/                      # Chaos Mesh experiments
├── loadtest/                   # k6 load tests
├── setup.sh                    # one-command platform bootstrap
└── docs/
    ├── azure-cloud-go-live-guide.md   # full cloud + HTTPS runbook
    └── project-context.md             # original project brief
```

---

## Key concepts demonstrated

**Kubernetes** — Deployments · ReplicaSets · Services · Ingress · namespaces · RBAC & ServiceAccounts · CRDs & controllers · finalizers · reconciliation loops · readiness probes · resource requests & limits

**Platform engineering** — GitOps · progressive delivery · observability (metrics + logs) · policy-as-code · chaos engineering · reproducible infrastructure · cost-aware cloud operations

---

## Cloud deployment

The platform also runs on a **cloud k3s cluster** (Azure VM) behind a real domain with **automatic Let's Encrypt TLS**. The same manifests run in both places — only the host changes.

📖 Full step-by-step runbook: **[docs/azure-cloud-go-live-guide.md](docs/azure-cloud-go-live-guide.md)**

---

<sub>Built as a deep dive into platform engineering — every component was built and understood from first principles.</sub>
