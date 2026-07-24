<div align="center">

# ☸️ Internal Developer Platform on Kubernetes

### The platform *around* the app — not just another deployed app.

**GitOps delivery · self-driving canary releases · full observability · layered security ·<br/>chaos-tested resilience · and a custom Kubernetes Operator written in Go.**

<br/>

[**🌍 Live Demo**](https://ayush-idp.duckdns.org) · [**🎥 Video Walkthrough**](https://drive.google.com/file/d/14-x45ra9iRBMOevZ0CqfLridM_HQqiSa/view?usp=sharing)

*(live demo runs on an Azure k3s cluster with real Let's Encrypt TLS — the VM is stopped when idle to control cost; happy to bring it up on request)*

<br/>

![Kubernetes](https://img.shields.io/badge/Kubernetes-k3s%20%7C%20kind-326CE5?logo=kubernetes&logoColor=white)
![GitOps](https://img.shields.io/badge/GitOps-ArgoCD-EF7B4D?logo=argo&logoColor=white)
![Progressive Delivery](https://img.shields.io/badge/Delivery-Argo%20Rollouts-EF7B4D?logo=argo&logoColor=white)
![Go](https://img.shields.io/badge/Operator-Go%20%2F%20Kubebuilder-00ADD8?logo=go&logoColor=white)
![Observability](https://img.shields.io/badge/Observability-Prometheus%20%7C%20Grafana%20%7C%20Loki-E6522C?logo=prometheus&logoColor=white)
![Security](https://img.shields.io/badge/DevSecOps-Trivy%20%7C%20Kyverno%20%7C%20Falco-4B32C3)
![Cloud](https://img.shields.io/badge/Cloud-Azure%20%7C%20Let's%20Encrypt-0078D4?logo=microsoftazure&logoColor=white)

</div>

---

## 💡 The idea

Most Kubernetes projects stop at *"I deployed an app."*

This project is everything a real company builds **around** the app — the internal platform that lets a whole team ship safely: automated deployment from Git, releases that judge themselves against live metrics, centralized monitoring and logging, policy enforcement at the cluster door, autoscaling proven under real load, and on-demand preview environments powered by an operator I wrote myself.

To prove the platform is **app-agnostic**, it runs two completely unrelated workloads with zero platform changes:

| Workload | What it is | Why it's here |
|---|---|---|
| 🛒 **Online Boutique** | Google's 11-service polyglot microservices demo | A realistic, non-trivial app *I deliberately didn't write* — in platform engineering, the app is the workload; **the platform is the project** |
| 📟 **Podinfo** | A tiny Go web service | Proof that a second, unrelated app onboards with no platform changes |

---

## 🏗️ Architecture

```mermaid
flowchart TB
    dev(["👨‍💻 Developer"]) == "① git push" ==> git[("📦 GitHub<br/>single source of truth")]
    git == "② ArgoCD pulls & syncs" ==> argocd
    users(["🌍 Users"]) == "③ HTTPS 🔒 Let's Encrypt" ==> nginx

    subgraph k8s["☸️ KUBERNETES CLUSTER — kind on my laptop · k3s on Azure"]
        direction TB
        argocd["🐙 ArgoCD<br/><b>GitOps engine</b><br/>self-heal · drift correction"]
        nginx["🌐 NGINX Ingress<br/><b>the front door</b>"]

        subgraph apps["WORKLOADS"]
            direction LR
            shop["🛒 Online Boutique<br/>11 microservices"]
            pi["📟 Podinfo"]
            prev["🔬 Preview env<br/>pr-42.localtest.me"]
        end

        op["⭐ CUSTOM GO OPERATOR<br/>PreviewEnvironment CRD"]

        subgraph ship["🚀 SHIP SAFELY"]
            roll["Argo Rollouts<br/>canary 25→50→75→100%"]
        end

        subgraph obs["📊 OBSERVE"]
            direction LR
            prom["Prometheus<br/>metrics"]
            graf["Grafana<br/>dashboards"]
            loki["Loki<br/>logs"]
        end

        subgraph sec["🛡️ SECURE"]
            direction LR
            kyv["Kyverno<br/>admission policy"]
            falco["Falco<br/>runtime alerts"]
            rbac["RBAC<br/>least privilege"]
        end

        subgraph res["📈 SCALE & SURVIVE"]
            direction LR
            hpa["HPA autoscaling<br/>+ k6 load tests"]
            chaos["Chaos Mesh<br/>failure injection"]
        end

        argocd == "deploys" ==> apps
        nginx ==> apps
        op -- "creates & tears down" --> prev
        apps -. "metrics & logs" .-> obs
        roll -. "asks: is the new version healthy?" .-> prom
    end

    classDef person fill:#1a7f37,stroke:#116329,color:#ffffff,font-weight:bold
    classDef gitN fill:#24292f,stroke:#57606a,color:#ffffff
    classDef argoN fill:#EF7B4D,stroke:#c4562a,color:#ffffff,font-weight:bold
    classDef nginxN fill:#009639,stroke:#00702b,color:#ffffff
    classDef appN fill:#326CE5,stroke:#2557b8,color:#ffffff
    classDef opN fill:#eac54f,stroke:#b08800,color:#24292f,font-weight:bold
    classDef obsN fill:#E6522C,stroke:#b93815,color:#ffffff
    classDef secN fill:#8250df,stroke:#6639ba,color:#ffffff
    classDef resN fill:#1f883d,stroke:#116329,color:#ffffff
    classDef zone fill:#f6f8fa,stroke:#d0d7de,color:#24292f

    class dev,users person
    class git gitN
    class argocd,roll argoN
    class nginx nginxN
    class shop,pi,prev appN
    class op opN
    class prom,graf,loki obsN
    class kyv,falco,rbac secN
    class hpa,chaos resN
    class apps,obs,ship,sec,res zone
    style k8s fill:none,stroke:#326CE5,stroke-width:2px,stroke-dasharray:8 6,color:#326CE5
```

**Same manifests run locally (kind) and in the cloud (k3s on Azure)** — only the host changes. That's the point of declarative infrastructure.

---

## ⚡ What it does

| Layer | Capability | Tools |
|---|---|---|
| **Run** | Runs any containerized microservice app | Kubernetes (kind / k3s) |
| **Deploy** | `git push` → auto-deploy, drift correction, self-heal | **ArgoCD** (GitOps) |
| **Release** | Canary rollouts with **metric-driven auto-rollback** | **Argo Rollouts** + Prometheus analysis |
| **Observe** | Live metrics, dashboards, centralized logs | **Prometheus · Grafana · Loki** |
| **Secure** | Image scanning · admission policy · runtime detection · least privilege | **Trivy · Kyverno · Falco · RBAC** |
| **Scale** | Autoscales under load, proven with load tests | **HPA** + **k6** |
| **Survive** | Chaos experiments prove self-healing | **Chaos Mesh** |
| **Self-service** ⭐ | On-demand ephemeral preview environments | **Custom Operator (Go / Kubebuilder)** |
| **Go live** | Public HTTPS on a cloud cluster | **Azure · k3s · cert-manager · Let's Encrypt** |

---

## ⭐ The standout: a custom Kubernetes Operator

I extended the Kubernetes API with my own resource type — and wrote the Go controller that reconciles it.

**A developer writes 7 lines:**

```yaml
apiVersion: platform.myproject.io/v1
kind: PreviewEnvironment
metadata:
  name: pr-42
spec:
  prNumber: 42
  image: nginx:1.27
```

**…and the operator does the rest:**

```mermaid
sequenceDiagram
    autonumber
    actor Dev as 👨‍💻 Developer
    participant API as ☸️ Kubernetes API
    participant Op as ⭐ Operator (Go)

    rect rgba(46, 160, 67, 0.15)
        Note over Dev,Op: 🟢 CREATE — developer asks, operator builds everything
        Dev->>API: kubectl apply → PreviewEnvironment "pr-42"
        API-->>Op: 🔔 watch event → Reconcile()
        Op->>API: create Namespace preview-pr-42
        Op->>API: create Deployment (image from spec)
        Op->>API: create Service
        Op->>API: create Ingress → pr-42.localtest.me
        Note over Op: 🌐 full isolated environment,<br/>live at its own URL — in seconds
    end

    rect rgba(248, 81, 73, 0.15)
        Note over Dev,Op: 🔴 DELETE — the finalizer guarantees cleanup
        Dev->>API: kubectl delete PreviewEnvironment "pr-42"
        API-->>Op: deletion paused — finalizer present
        Op->>API: delete namespace (everything inside cascades)
        Op->>API: remove finalizer → object fully deleted
        Note over Op: 🧹 zero orphaned resources,<br/>zero cost leaks
    end
```

This is the **same reconciliation pattern Kubernetes itself is built on** — a declared desired state, and a control loop that makes reality match it. Idempotent reconcile, finalizer-based cleanup, least-privilege RBAC generated from Kubebuilder markers.

> The real-world use case: every pull request gets its own isolated, disposable copy of the app at its own URL — the feature you know from Vercel/Netlify previews, built here on raw Kubernetes.

📁 [`operator/`](operator/) · 📖 [`operator/README.md`](operator/README.md)

---

## 🚀 Releases that judge themselves

New versions don't ship all at once — and no human decides whether they're healthy. **Prometheus does.**

```mermaid
flowchart LR
    A["🚢 New version<br/>deployed"] ==> B["🐤 Canary gets<br/><b>25%</b> of traffic"]
    B ==> C{"🔬 Prometheus check<br/>success rate ≥ 95%?<br/><i>4 checks, 20s apart</i>"}
    C == "✅ healthy" ==> D["📈 Auto-promote<br/>50% → 75% → 100%"]
    C == "❌ degraded" ==> E["🤖 Auto-rollback<br/>to last good version"]
    D ==> F(["✔ Stable — no human involved"])
    E ==> F

    classDef start fill:#316dca,stroke:#255ab2,color:#ffffff
    classDef canary fill:#EF7B4D,stroke:#c4562a,color:#ffffff
    classDef check fill:#eac54f,stroke:#b08800,color:#24292f,font-weight:bold
    classDef good fill:#1f883d,stroke:#116329,color:#ffffff
    classDef bad fill:#cf222e,stroke:#a40e26,color:#ffffff
    class A start
    class B canary
    class C check
    class D,F good
    class E bad
    linkStyle 2 stroke:#1f883d,stroke-width:3px
    linkStyle 3 stroke:#cf222e,stroke-width:3px
```

Two independent safety nets:
- **Metric gate** — an `AnalysisTemplate` runs a live PromQL query (non-5xx ÷ total requests) against real traffic; one failed check aborts and rolls back
- **Health gate** — `progressDeadlineAbort` + readiness probes catch versions that never even become healthy

📁 [`apps/podinfo-canary/`](apps/podinfo-canary/)

---

## 🛡️ Security: defense in depth

Four layers, at four different stages of a container's life — no single control has to be perfect, because there's another one behind it.

```mermaid
flowchart LR
    img["📦 Container image"] ==> trivy["🔍 <b>TRIVY</b><br/>CVE scan<br/><i>build time</i>"]
    trivy == "✅ no critical CVEs" ==> kyv["🚪 <b>KYVERNO</b><br/>admission policy<br/><i>deploy time</i>"]
    kyv == "✅ policy passed" ==> pod["🟢 Running pod"]
    bad["😈 nginx:latest<br/>untagged image"] -- "❌ REJECTED at the door" --> kyv
    falco["👁️ <b>FALCO</b><br/>runtime detection<br/><i>shells, weird syscalls</i>"] -. "watches 24/7" .-> pod
    rbac["🔑 <b>RBAC</b><br/>least privilege<br/><i>read-only dev identity</i>"] -. "guards the API" .-> pod

    classDef stage fill:#316dca,stroke:#255ab2,color:#ffffff
    classDef gate fill:#8250df,stroke:#6639ba,color:#ffffff,font-weight:bold
    classDef ok fill:#1f883d,stroke:#116329,color:#ffffff
    classDef threat fill:#ffebe9,stroke:#cf222e,color:#cf222e,font-weight:bold
    classDef watch fill:#57606a,stroke:#424a53,color:#ffffff
    class img stage
    class trivy,kyv gate
    class pod ok
    class bad threat
    class falco,rbac watch
    linkStyle 3 stroke:#cf222e,stroke-width:3px
```

The Kyverno policy is enforced live — `kubectl run test --image=nginx` (untagged → `:latest`) is **rejected at admission** and never reaches the cluster.

📁 [`security/`](security/)

---

## 📈 Proven, not assumed

Claims are cheap. Every resilience property here is **demonstrated**:

- **Autoscaling** — a k6 load test ramps 100 virtual users against the shop; the HPA scales the frontend from 1 → 10 pods as CPU crosses the 50% target, then scales back down after the stabilization window. Watched live.
- **Self-healing** — Chaos Mesh kills the frontend pod on purpose; Kubernetes schedules a replacement in seconds while the Service keeps routing to healthy pods.
- **Slow-dependency chaos** — injected 2s network latency (the failure mode far more common than a clean crash) to observe degradation behaviour.
- **Disaster recovery** — when my local cluster's certificates corrupted, I rebuilt the *entire* platform in minutes with one script. Cattle, not pets.

📁 [`loadtest/`](loadtest/) · [`chaos/`](chaos/)

---

## 🔥 Debugging war stories (the part tutorials skip)

Real things that broke while building this — and how I fixed them:

| Incident | Root cause | Fix |
|---|---|---|
| **Grafana crash-looping, 432 restarts** | Two datasources (Prometheus + Loki) both provisioned with `isDefault: true` — Grafana refuses to start | Read the container logs → found the provisioning error → patched the Loki ConfigMap to `isDefault: false` |
| **Entire local cluster unreachable** (`x509: certificate signed by unknown authority`) | kind pins the node IP; Docker network changed it after a reboot | Rebuilt the whole platform from `setup.sh` in minutes — the recovery story that proves reproducibility |
| **ArgoCD install failed** (`metadata.annotations: Too long`) | Huge CRDs exceed the annotation limit used by client-side apply | `kubectl apply --server-side` |
| **Canary analysis inconclusive** | Deployed before Prometheus had scraped any traffic — `rate(...[1m])` over no data | Generate load first, wait a scrape cycle, then roll out |

> Every one of these taught me more than the happy path did. **Read the logs, don't guess.**

---

## 🖥️ Quick start — one command

```bash
bash setup.sh
```

Provisions the entire platform from scratch: kind cluster → ingress-nginx → metrics-server → ArgoCD → all apps (via GitOps) → Prometheus + Grafana → Loki → Kyverno → Falco → Chaos Mesh.

Then start the operator:

```bash
cd operator && make install && make run
```

---

## 📁 Repository structure

```
├── infra/kind-config.yaml      # local cluster, as code
├── apps/
│   ├── online-boutique/        # 11-service demo workload + HPA
│   ├── podinfo/                # 2nd app — proves the platform is app-agnostic
│   ├── podinfo-canary/         # automated canary + Prometheus analysis
│   └── hello/                  # minimal reference app
├── bootstrap/                  # ArgoCD Applications (GitOps, as code)
├── operator/                   # ⭐ custom Kubernetes Operator (Go / Kubebuilder)
├── security/                   # Kyverno policies + RBAC
├── chaos/                      # Chaos Mesh experiments
├── loadtest/                   # k6 load tests
└── setup.sh                    # one-command platform bootstrap
```

---

## ☁️ Cloud deployment

The platform also runs on a **k3s cluster on an Azure VM**, behind a real domain, with **automatic Let's Encrypt TLS** via cert-manager (ACME HTTP-01). Same manifests as local — only the ingress host changes.

🌍 **https://ayush-idp.duckdns.org**

---

## 🎓 Key concepts demonstrated

**Kubernetes** — Deployments · ReplicaSets · Services · Ingress · namespaces · RBAC & ServiceAccounts · CRDs & controllers · reconciliation loops · finalizers · readiness probes · resource requests & limits · HPA

**Platform engineering** — GitOps · progressive delivery · observability (metrics + logs) · policy-as-code · admission control · chaos engineering · load testing · reproducible infrastructure · TLS automation · cost-aware cloud operations

---

## 🧭 What this is — and isn't

**This is a portfolio platform, not a production system with real users** — and I think saying that clearly matters.

What it *is*: an end-to-end platform I designed, wired, broke, and fixed myself. Every tool is here for a reason I can defend, every capability is demonstrated live rather than claimed, and the whole thing rebuilds from scratch with one script.

---

<div align="center">
<sub>Built as a deep dive into platform engineering — every component built and understood from first principles.<br/>
<b>Ayush Bansal</b> · <a href="https://github.com/Ayush-Bansal08">GitHub</a> · <a href="https://ayush-idp.duckdns.org">Live demo</a></sub>
</div>
