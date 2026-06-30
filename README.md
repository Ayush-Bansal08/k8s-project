# k8s-project — Production-Grade Internal Developer Platform

> **📌 NOTE TO THE NEXT CLAUDE CODE SESSION (read this first):**
> This README is a **complete context handoff**. The user is **new to DevOps**, just finished a 12-day "Kubernetes in Hindi" course (YAML/Linux/Docker → architecture → pods → deployments → security → observability → runtime security), and wants to build **one impressive portfolio project to get hired in a DevOps / Platform / Cloud / SRE role.**
> The user prefers **simple, plain-language explanations with real-world analogies** (not jargon dumps), wants to keep the whole project **100% free**, and is building this step-by-step as a beginner.
> This document contains **everything** discussed while scoping the project: the plan, the tools, the reasoning, the analogies used to explain each concept, and ready-made answers to the questions a senior/recruiter will ask. Use it to be "just like" the previous session. Don't assume prior K8s depth — explain things simply and build incrementally (Tier 1 first).

---

## Table of contents
1. [The one-line pitch](#the-one-line-pitch)
2. [What this is (and isn't)](#what-this-is-and-isnt)
3. [Who uses it / why it matters](#who-uses-it--why-it-matters)
4. [Simple explanation with analogies](#simple-explanation-with-analogies-how-it-was-taught)
5. [What the platform does (by layer)](#what-the-platform-does-by-layer)
6. [The full tool stack](#the-full-tool-stack)
7. [Every concept explained simply](#every-concept-explained-simply)
8. [The automation framework](#the-automation-framework)
9. [Where it runs — 100% free](#where-it-runs--100-free)
10. [Build order (tiers)](#build-order-dont-try-to-do-it-all-at-once)
11. [Q&A — how to answer your senior / recruiter](#qa--how-to-answer-your-senior--recruiter)
12. [The polished project statement](#the-polished-project-statement)
13. [Proof artifacts](#proof-artifacts-make-it-visible)
14. [Status](#status)

---

## The one-line pitch

> I built a platform that can run **any standard microservice app**. The app runs *on* the platform, which **auto-deploys** it, **monitors** it in real time, **collects all its logs and errors**, and **automatically detects, diagnoses, and recovers** from problems. Because it's app-agnostic, I can swap in a completely different microservice app and everything still works.

**Why this project (the recruiter truth):** "I deployed an app to Kubernetes" is on 80% of junior resumes and impresses no one — everyone can `kubectl apply`. What makes a hiring manager say *"get this person on a call"* is proof you can run software **the way a real company does** (GitOps, observability, security, autoscaling, self-healing) — plus the killer bonus that you can **extend Kubernetes itself** (a custom operator). This project does all of that and uses **every day** of the course, then pushes two steps beyond it.

---

## What this is (and isn't)

- **It IS** an **Internal Developer Platform (IDP)** — the system developers use to ship apps safely onto a cluster.
- **It is NOT** Kubernetes-as-a-Service (KaaS hands you whole clusters, like GKE/EKS).
- **The app is a prop.** The demo app (Google's **Online Boutique**, an 11-service polyglot microservices demo) is **downloaded, not built**. ALL the engineering is the **platform around it**.
- **It must stay app-agnostic** — proven by deploying a **2nd unrelated app** onto the same platform with no changes. The "contract": any app that is **(1) containerized** and **(2) has standard Kubernetes manifests (Helm/Kustomize)** plugs right in.

**Important reframe (the user worried about this a lot):** In DevOps, *the app is never the project — the system that runs the app is the project.* Real DevOps engineers don't write the product's code; they build the infrastructure that runs *other people's* code. So using a ready-made app is **correct and professional, not cheating.** The app runs *on* the platform (like a guest), it is **not** "connected to" or "baked into" it.

---

## Who uses it / why it matters

- **Users (in a real company):** developers — e.g. "Priya," who wrote a feature but **doesn't know Kubernetes** and doesn't want to. She just does `git push`; the platform builds, scans, deploys, gives her a preview URL, monitors it, and auto-blocks/rolls-back bad changes. She ships in **minutes instead of days**, safely.
- **Builder:** the **DevOps / Platform engineer** — this is the job this project proves you can do.
- **For YOU right now:** the real "user" is the **recruiter/interviewer**, who uses it to verify you can do the job. A portfolio project doesn't need real users — its job is to **prove you could build the thing real users depend on.**
- **Company value:** faster shipping · fewer outages · less 3 AM firefighting · security built-in.

---

## Simple explanation with analogies (how it was taught)

**The big idea (McDonald's analogy):** Cooking one burger = running one app on Kubernetes (boring, anyone can do it). Running an entire **McDonald's franchise** = this project. The franchise is everything *around* the burger: the assembly line that restocks/rebuilds automatically, the CCTV/dashboards, the health inspectors and security guards, the manager who instantly replaces a sick worker. **You build the franchise machinery. The burger (app) is just a prop.**

**A day in the life (online shoe store):**
- **Scene 1 — You change the "Buy" button to green.** Instead of editing the live server by hand, you edit a text file in **Git**. A robot (**ArgoCD**) constantly compares Git to the live store and makes reality match. If someone changes the live store by hand, it resets it back. *Like a thermostat — set what you want in one place, it forces reality to match.*
- **Scene 2 — Before it goes live, 3 guards inspect it.** **Trivy** = airport baggage scanner (scans the container for known vulnerabilities). **Kyverno** = bouncer with a rulebook (e.g. "no app may run as admin"; rejects rule-breakers at the door). **Falco** = silent alarm inside the building (screams if a running app behaves suspiciously).
- **Scene 3 — Watching its health.** **Prometheus + Grafana** = TV screens of live graphs, like a hospital heart-rate monitor for the website (slow? errors? traffic?).
- **Scene 4 — Black Friday rush.** **Autoscaling** automatically spins up more copies to share the load, then removes them when quiet (proven by blasting fake traffic with **k6**).
- **Scene 5 — A computer crashes at 3 AM.** **Self-healing** instantly revives the dead part on another machine, no human (proven by deliberately killing pods with **Chaos Mesh**).
- **Scene 6 — A coworker wants to test their own idea (THE SHOWSTOPPER).** Your **custom Operator** (a robot *you* build) automatically creates a complete private copy of the store at a temporary URL for their change, then deletes it when done. Almost no beginner can build this — it proves you understand Kubernetes deeply enough to *extend* it.

**Other analogies used:**
- **Monitoring & app-agnostic (hospital heart-rate monitor):** the monitor isn't built around one patient — it reads *whoever is plugged in*. Online Boutique is just the first patient. Swap apps, it monitors the new one. Zero changes.
- **You build the system, not the app (movie theater):** a theater owner doesn't make the movies. They build/run the theater (screens, sound, ticketing, security). The movies (apps) are just content playing inside the system they built.
- **RBAC (office keycards):** the intern's keycard opens only the supply closet; the manager's opens everything; the cleaning robot's opens only what it needs. RBAC = handing out those keycards. Rule: **least privilege** (minimum keys needed).
- **Terraform (furniture instruction sheet):** instead of assembling furniture by hand every time, you write an instruction sheet (code) and a robot builds it identically every time — and can tear it down.

---

## What the platform does (by layer)

| # | Layer | What it does |
|---|---|---|
| 1 | **Run** | Runs any containerized microservice app (demo = Online Boutique) |
| 2 | **Auto-deploy** | GitOps: push code → auto-build → auto-deploy, with canary + auto-rollback |
| 3 | **Observe** | Metrics, logs, and traces in one place (the 3 pillars) |
| 4 | **Secure** | Image scanning, admission policies, runtime detection, RBAC, network policies, TLS, secrets |
| 5 | **Survive** | Autoscaling, self-healing, HA — proven with load tests + chaos experiments |
| 6 | **Self-service** ⭐ | Custom Operator: ephemeral per-PR preview environments (the standout) |
| 7 | **Self-driving** | Agentic auto-remediation: detect → diagnose → recover (capstone, user's own idea) |
| 8 | **Cost** | Cost visibility via OpenCost |

---

## The full tool stack

| Concern | Tool | Notes |
|---|---|---|
| Containers | **Docker** | Build images (multi-stage, distroless, non-root). Foundation layer. |
| Cluster | **k3s** | Lightweight Kubernetes |
| Infra as Code | **Terraform** | Provisions the server (optional polish, high-value skill) |
| CI (build/scan) | **GitHub Actions** | Free, builds + scans on every push |
| CD (GitOps) | **ArgoCD** | Keeps cluster in sync with Git |
| Packaging | **Helm + Kustomize** | Templating + dev/staging/prod overlays |
| Metrics | **Prometheus** | "Is something wrong?" |
| Dashboards | **Grafana** | Visualize everything |
| Logs | **Loki** | "What exactly happened?" |
| Traces | **Tempo + OpenTelemetry** | "Where/why did it break?" |
| Image scanning | **Trivy** | Vulnerabilities before deploy |
| Admission policy | **Kyverno** (+ OPA) | Block unsafe deployments |
| Runtime security | **Falco** | Detect suspicious behavior at runtime |
| Access control | **RBAC** | Least-privilege "keycards" |
| Network security | **Network Policies** | Control pod-to-pod traffic |
| TLS / HTTPS | **cert-manager + Let's Encrypt** | Free real certificates |
| Secrets | **External Secrets Operator** | Safe secret management |
| Autoscaling | **HPA** | Scale with traffic |
| Reliability | **PodDisruptionBudgets, affinity, taints/tolerations, QoS** | High availability |
| Load testing | **k6** | Prove autoscaling works |
| Chaos testing | **Chaos Mesh** | Prove self-healing works |
| Custom operator ⭐ | **Kubebuilder (Go)** or **Kopf (Python)** | Preview environments |
| Auto-remediation | **n8n + Alertmanager + LLM (Ollama)** | Agentic, human-in-the-loop |
| Cost | **OpenCost** | Cluster cost visibility |
| (Optional KaaS direction) | **vCluster** | Hand out virtual clusters instead of namespaces |

**Course-day → project mapping (uses everything learned):** Docker/YAML/kubectl (build & run) · Deployments/ReplicaSets/Services/Ingress (run + expose) · Namespaces/RBAC/ServiceAccounts (access + operator perms) · Secrets/External Secrets (safe keys) · Volumes/Storage Classes (database storage) · Taints/Tolerations/Affinity/QoS (smart placement) · cert-manager/TLS/Gateway API (real HTTPS) · Trivy/Kyverno/OPA/Falco/Network Policies/Security Context (security) · Prometheus/Grafana/Jaeger/OTel (observability) · OpenCost (cost) · HA/PDBs (survive crashes) · Admission Controllers/control-plane (what the operator builds on).

---

## Every concept explained simply

**GitOps (ArgoCD):** You write what you *want* in Git (a "notebook"). A robot keeps the live cluster matching the notebook automatically — and undoes manual changes. You never run `kubectl apply` by hand.

**The 3 pillars of observability (why Loki AND Tempo):** In a microservices app, one click ("Buy") hops through many services. Metrics alone leave you blind. So:
- **Metrics (Prometheus)** → "Is something wrong?" (error rate jumped to 8%)
- **Logs (Loki)** → "What exactly happened?" (the actual error text)
- **Traces (Tempo + OpenTelemetry)** → "Where in the chain, and what was slow?" (request stuck 4.2s in payment)
> Honest scope note: Prometheus+Grafana is the non-negotiable core; **Loki (logs) is high-value, low-effort — add it**; **Tempo (traces) is the most impressive but most setup** (nearly free here because Online Boutique is already OTel-instrumented). If ever cutting scope, drop Tempo last — distributed tracing is a senior skill.

**RBAC:** "Who is allowed to do what" — least-privilege keycards. Real uses here: the **operator** needs a least-privilege ServiceAccount to create namespaces/deployments; **ArgoCD** needs scoped permissions; each app part gets its own limited identity; you can model a "dev" role blocked from prod.

**Docker (the foundation):** Every service runs as a Docker container; CI builds Docker images on every push; you write Dockerfiles for the parts you create (the operator, the 2nd app); Trivy scans your images. Skills built: multi-stage builds, distroless/non-root images, tagging, pushing to ghcr.io.
> Sharp nuance: Modern Kubernetes uses **containerd** as the runtime, not Docker — but you **still use Docker to BUILD images.** "You build with Docker → Kubernetes runs the resulting image."

**Terraform:** Creates/manages your cloud infrastructure (servers, networks, firewalls) from a file instead of clicking in a web console. `terraform apply` builds it, edits update only what changed, `terraform destroy` deletes it. Infrastructure as Code. Optional to start (you can click-create the VM manually), worth adding later — it's a highly-requested skill.

**The agentic auto-remediation loop (user's own idea — Tier 4 capstone):**
1. **Detect** — Prometheus Alertmanager or a Loki log-alert fires (e.g. "checkout error rate > 5%" or "OOMKilled") → webhook to **n8n**.
2. **Enrich** — n8n gathers context: logs (Loki), metrics (Prometheus), pod status + recent deploys (K8s API).
3. **Reason** — feed the bundle to an **LLM agent** (Ollama locally for free) → diagnose root cause + propose action + confidence.
4. **Act (tiered):** safe/reversible → auto-fix (restart, scale, rollback via ArgoCD); risky/low-confidence → post diagnosis + proposed fix to Slack with **Approve/Reject** buttons (**human-in-the-loop**).
> **Senior-level caveat (must say this):** never let an LLM take destructive, irreversible actions on prod unsupervised. Start with **diagnose-and-suggest**, graduate to auto-act only for safe, reversible, well-understood cases with guardrails (rate limits, blast-radius caps, dry-run). This solves **MTTR** (mean time to resolve) and **toil** — a "self-driving platform."

---

## The automation framework

**Backbone:** GitOps = **GitHub Actions (CI)** + **ArgoCD (CD)**, on **Terraform**-provisioned infrastructure.

**Auto-deploy flow:**
```
You: git push   (the only manual step)
   ↓
GitHub Actions:  build → security-scan (Trivy) → push image
   ↓
ArgoCD:  detects the change → auto-deploys to the cluster
   ↓
Website live with the update ✅  (canary rollout + auto-rollback if it breaks)
```
You never run `kubectl apply` by hand or log into a server.

**Operator language choice:** know/can-learn **Go → Kubebuilder** (most credible, what real operators use); prefer **Python → Kopf** (easier, still legit).

---

## Where it runs — 100% free

- **Develop on:** `kind` or `k3d` on the laptop (free)
- **Go live on:** **Oracle Cloud Always Free** (4 ARM cores / 24 GB RAM, **free forever**) → install **k3s**
- **Domain + HTTPS:** **DuckDNS** (free subdomain, works with Let's Encrypt) + **Let's Encrypt** (free real certs)
- **LLM (Tier 4 only):** **Ollama** (local, free)
- **Everything else:** open source, self-hosted

**Honest gotchas:** (1) cloud signup needs a credit/debit card for **verification** — never charged on the free tier. (2) Oracle's free ARM servers are popular and sometimes show "out of capacity" — retry, pick a less busy region, or use an auto-retry script (`oci-arm-host-capacity`); Plan B is a ~$4–5/mo VM. (3) **Students:** the **GitHub Student Developer Pack** gives a free real domain for a year + cloud credits.

**Oracle vs Google Cloud:** same *kind* of thing — both are cloud providers (rent servers over the internet). We use Oracle ONLY because its free tier is free forever (Google's is $300 for 90 days, then paid). Kubernetes skills transfer 100% across clouds; only the signup/console looks different.

---

## Build order (don't try to do it all at once)

### Tier 1 — already beats most resumes (start here)
Docker images → run Online Boutique on local `kind` → GitOps with ArgoCD → Ingress + TLS → Prometheus + Grafana.

### Tier 2 — senior signal
Security (Trivy, Kyverno, Falco, RBAC, Network Policies) → logs (Loki) + traces (Tempo) → autoscaling (HPA) + k6 load test → move to Oracle Cloud + real DuckDNS domain.

### Tier 3 — unforgettable
Custom Operator + per-PR preview environments → canary deploys (Argo Rollouts) → Chaos Mesh → **prove app-agnostic by deploying a 2nd app**.

### Tier 4 — capstone
Agentic auto-remediation (n8n + Alertmanager + LLM, human-in-the-loop).

### Polish
Terraform (Infrastructure as Code), OpenCost, architecture diagram, demo video, LinkedIn writeup.

---

## Q&A — how to answer your senior / recruiter

**"Are you building Kubernetes-as-a-Service?"**
> "No. KaaS means provisioning whole *clusters* as a product (like GKE/EKS). I'm building an **Internal Developer Platform** — a layer on top of one existing cluster that gives developers a smooth, safe way to ship apps: GitOps, preview environments, observability, security guardrails. Preview environments are isolated per-namespace, not separate clusters." *(If pushed: the line is what you hand the user — KaaS hands a cluster, my platform hands a paved path to deploy onto a cluster that already exists. If I ever wanted KaaS, the clean path is vCluster — virtual clusters inside one real cluster.)*

**"Does it solve bugs automatically?"** (be precise — a senior will test this)
> "It doesn't rewrite the app's code. For **operational problems** (crashes, slowness, overload, bad deploys) it detects, diagnoses, and can **auto-recover** (restart, roll back, scale). For **code bugs** it hands the developer the exact logs, traces, and metrics showing *where* and *why* it broke, so they fix it fast." *(Don't say "it solves all bugs automatically.")*

**"What's your automation framework?"**
> "It's GitOps-based — GitHub Actions for CI and ArgoCD for continuous deployment — with Terraform automating the infrastructure underneath."
> *Humanized:* "The whole thing is automated in three layers. Terraform sets up the servers automatically. Every time I push code, GitHub Actions builds it and checks it for problems. And ArgoCD deploys it to the cluster on its own. So once I push, everything from building to going live happens automatically — I don't touch anything manually."

**"Isn't it tied to that one app?"**
> "No — it's app-agnostic. The app runs *on* the platform, not baked into it. I proved it by deploying a second unrelated app onto the same platform unchanged — deployment, monitoring, and preview environments all just worked." *(The monitoring watches whatever pods run; ArgoCD deploys whatever's in Git; policies apply cluster-wide.)*

**"Doesn't Kubernetes not use Docker anymore?"**
> "Right — K8s uses containerd as the runtime, but I still use Docker to build and test images locally. You build with Docker, Kubernetes runs the resulting image."

**"Did you actually build anything, or just use a pre-made app?"**
> "The app is the workload, not the project. Everything around it is mine: the GitOps setup, monitoring/logging/tracing, security policies, autoscaling config, CI/CD pipelines, and a **custom Kubernetes Operator I wrote** — that's real code. In DevOps you don't write the product code; you build the platform that runs it."

---

## The polished project statement

**Full version:**
> In this project, I take a pre-built website (Google's Online Boutique) and run it on top of a complete platform that I build around it. The website runs *inside* my system, and my platform handles everything around it. It automatically deploys the website whenever the code changes, monitors the whole thing in real time, and collects all its logs, errors, and performance data in one place. When something goes wrong, my system detects it, shows exactly *where* and *why* it happened, and can even recover on its own — for example by restarting a crashed service or rolling back a bad deployment. It works with **any standard microservice app**, not just this one. So in short, the platform's job is to deploy the website automatically, keep it healthy, and make any problem easy to find and fix.

**Short version:**
> I take a ready-made website (Google's Online Boutique) and run it on a platform I built around it. The platform auto-deploys the site, monitors it in real time, collects all its logs and errors, and automatically detects, diagnoses, and recovers from problems — and it works with any standard microservice app, not just this one.

---

## Proof artifacts (make it visible)

- [ ] Public GitHub repo with this README + one-command bootstrap (`make up`)
- [ ] Architecture diagram (Excalidraw)
- [ ] Live demo URL with valid HTTPS
- [ ] 2–3 min demo video (Grafana dashboards + a chaos experiment showing self-healing)
- [ ] One LinkedIn / blog writeup per milestone (this is how recruiters *find you*)

---

## Status

- [ ] Tier 1 — not started yet (next step: Docker + run Online Boutique on local `kind`)
- [ ] Tier 2
- [ ] Tier 3
- [ ] Tier 4
- [ ] Polish
