# 🎬 Demo Video — Complete Production Guide

> **Goal:** a 5–6 minute video that makes a hiring manager think *"we need to talk to this person."*
> **Method:** record **11 short parts** separately, then stitch them together.
> **Rule:** show things **working**, don't explain how they're built. The README explains; the video *proves*.

---

## 📋 PART 0 — Preparation (do this before recording anything)

### A. Technical setup
| Item | Setting | Why |
|---|---|---|
| **Screen resolution** | 1920×1080 | standard, crisp on YouTube |
| **Terminal font size** | **18–22pt** (big!) | text must be readable on a phone |
| **Terminal theme** | dark bg, high contrast | looks sharp on video |
| **Browser zoom** | 110–125% | dashboards readable |
| **Notifications** | **OFF** (Windows Focus Assist) | no popups mid-take |
| **Recording tool** | **OBS Studio** (free) or Windows Game Bar (`Win+G`) | OBS is better; free |
| **Mic** | any headset/phone earbuds | clear audio > fancy video |

### B. Get the platform running (before you hit record)
```bash
# 1. Local cluster + everything
bash setup.sh                       # wait until it finishes

# 2. The operator (leave running in its own terminal)
cd operator && make install && make run

# 3. Port-forwards (each in the background or its own terminal)
kubectl port-forward svc/argocd-server -n argocd 8080:443 &
kubectl port-forward svc/monitoring-grafana -n monitoring 3000:80 &
kubectl argo rollouts dashboard &

# 4. Start the CLOUD VM (Azure Portal -> idp-server -> Start)
#    so https://ayush-idp.duckdns.org is live for Part 2
```

### C. Browser tabs to pre-open (in this order)
1. `https://ayush-idp.duckdns.org` — the live cloud site
2. `http://shop.localtest.me` — the local shop
3. `https://localhost:8080` — ArgoCD (logged in already)
4. `http://localhost:3000` — Grafana (on the Namespace/Pods dashboard, namespace `default`)
5. `http://localhost:3100` — Argo Rollouts dashboard
6. Your GitHub repo page

### D. Get passwords ready (so you don't fumble on camera)
```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo
# Grafana: admin / prom-operator
```

> ✅ **Pre-flight check:** every tab loads, the operator is running, the cloud site is up. *Then* record.

---

## 🎬 THE 11 PARTS

**Format for each part:** 🖥️ *what's on screen* · ⌨️ *commands* · 🗣️ *what to say*

---

### PART 1 — The Hook (~20 sec)

🖥️ **Show:** your GitHub repo README (top of the page, showing the title + badges).

🗣️ **Say:**
> "Hi, I'm Ayush. Most Kubernetes projects stop at *'I deployed an app.'*
> This one is the **platform around the app** — the system a real company builds so developers can ship safely.
> It runs any microservice app, deploys it automatically from Git, monitors it, secures it, scales it —
> and it extends Kubernetes itself with a custom operator I wrote in Go.
> Let me show you."

💡 *Tip:* Energy matters most in the first 10 seconds. Sit up, smile, speak a bit faster than feels natural.

---

### PART 2 — It's LIVE on the internet (~30 sec)

🖥️ **Show:** browser → `https://ayush-idp.duckdns.org` → then **click the padlock** → "Connection is secure" → certificate details.

🗣️ **Say:**
> "First, this isn't only running on my laptop.
> This is **live on the internet** — a cloud Kubernetes cluster I deployed on Azure running **k3s**,
> behind a real domain, with an automatic **Let's Encrypt** certificate.
> [click padlock] Real TLS — issued and auto-renewed by **cert-manager** through the ACME protocol.
> Same manifests as my local cluster; only the host changed."

💡 *Tip:* Slowly hover/click the padlock — the visual of a valid cert is powerful.

---

### PART 3 — ⭐ The Custom Operator (the showstopper) (~75 sec)

🖥️ **Show:** split view — left: terminal running the operator (`make run`), right: a second terminal. Optionally show the CRD file first.

⌨️ **Commands:**
```bash
# 1. Show the request (the custom resource YOU invented)
cat operator/preview-sample.yaml

# 2. Nothing exists yet
kubectl get ns | grep preview

# 3. Apply it — the magic
kubectl apply -f operator/preview-sample.yaml

# 4. Show what the operator built, automatically
kubectl get ns | grep preview
kubectl get all,ingress -n preview-pr-42

# 5. (optional) open the preview URL in a browser
curl -I http://pr-42.localtest.me

# 6. Delete it → the finalizer cleans everything up
kubectl delete -f operator/preview-sample.yaml
kubectl get ns | grep preview        # gone
```

🗣️ **Say:**
> "This is the piece I'm most proud of. I **extended the Kubernetes API with my own resource type** —
> a `PreviewEnvironment` — and wrote the controller that reconciles it, in **Go with Kubebuilder**.
>
> [cat the file] This is all a developer writes: a pull-request number and an image.
>
> [apply] Watch what my operator does automatically — [show] it creates a **namespace**,
> deploys the **app**, creates a **Service**, and exposes it at its **own URL**.
> That's a complete, isolated preview environment — the kind a developer gets per pull request,
> so they can test a change without touching production.
>
> And when the pull request closes — [delete] a **finalizer** tears the whole thing down. No orphaned resources.
>
> This is the same reconciliation pattern Kubernetes uses internally.
> **I'm not just using Kubernetes — I'm extending it.**"

💡 *Tip:* This is your best 75 seconds. Rehearse it twice. Show the operator's **log lines** appearing as it creates things — that's the "wow."

---

### PART 4 — GitOps: git push = deploy (~45 sec)

🖥️ **Show:** split — VS Code (the deployment.yaml) + ArgoCD UI + a terminal.

⌨️ **Commands:**
```bash
# (edit apps/hello/deployment.yaml: replicas 4 -> 5 in VS Code, on camera)
git add apps/hello/deployment.yaml
git commit -m "Scale hello to 5 replicas"
git push

# then in ArgoCD UI: click REFRESH on the hello app, watch it sync
kubectl get pods -l app=hello
```

🗣️ **Say:**
> "Deployments here are **GitOps**. I never run `kubectl apply` by hand.
> **Git is the single source of truth**, and **ArgoCD** continuously reconciles the cluster to match it.
>
> Watch — I change the replica count and push. [push]
> ArgoCD detects the change and deploys it automatically. [show the app go OutOfSync → Synced, new pod appears]
>
> And it works in reverse too: **self-heal**. If someone changes the cluster by hand,
> ArgoCD reverts it back to what's in Git. The cluster can't drift."

💡 *Tip:* The ArgoCD UI going yellow → green is great visual. Zoom into the app tile.

---

### PART 5 — Automated canary with auto-rollback (~50 sec)

🖥️ **Show:** the Argo Rollouts dashboard (`localhost:3100`) or the terminal `get rollout` output.

⌨️ **Commands:**
```bash
# deploy a new version
kubectl argo rollouts set image podinfo-canary podinfo=ghcr.io/stefanprodan/podinfo:6.6.2

# watch it progress + the AnalysisRun
kubectl argo rollouts get rollout podinfo-canary
```

🗣️ **Say:**
> "New versions don't go out all at once. This is a **canary rollout** with Argo Rollouts.
>
> I deploy a new version — [set image] — and it goes to **25% of traffic** first.
> Then, the key part: it **queries Prometheus** for the live success rate. [point at the AnalysisRun]
>
> If the metrics are healthy, it **auto-promotes** to 50, 75, 100 percent.
> If they degrade, it **automatically rolls back**. No human in the loop.
>
> That's **progressive delivery driven by real observability data** — the deploy pipeline
> consuming the monitoring pipeline to make its own decisions."

💡 *Tip:* The `AnalysisRun ... ✔3` line is the money shot — zoom in on it.

---

### PART 6 — Autoscaling under real load (~50 sec)

🖥️ **Show:** split — k6 running (left) + `watch kubectl get hpa` (right) + Grafana in a third pane if possible.

⌨️ **Commands:**
```bash
# terminal 1
k6 run loadtest/shop-test.js

# terminal 2
watch kubectl get hpa
```

🗣️ **Say:**
> "Let's see it handle traffic. I'm running a **k6 load test** — 100 virtual users hitting the shop.
>
> [point at HPA] Watch the autoscaler: as CPU crosses the 50% target,
> Kubernetes **automatically adds pods** — one, two, three — up to ten, to absorb the load.
>
> [Grafana] And here it is in Grafana — you can see CPU climbing across the microservices.
>
> When the load stops, it **scales back down** after a stabilization window, so it doesn't flap.
> That's autoscaling **proven with a real load test**, not just configured."

💡 *Tip:* Speed this section up 2× in editing — the ramp takes a minute, but nobody wants to watch it in real time.

---

### PART 7 — Observability: metrics + logs (~35 sec)

🖥️ **Show:** Grafana dashboard (namespace `default`), then switch to Explore → Loki query.

⌨️ **Loki query to run on camera:**
```
{namespace="default"} |= "checkout"
```

🗣️ **Say:**
> "Everything is observable. **Prometheus** scrapes metrics from every service,
> and **Grafana** visualizes them — CPU and memory per microservice, in real time.
>
> Logs are centralized in **Loki** — [run the query] here I'm filtering the shop's logs for checkout requests.
>
> That's the workflow that matters in an incident: **metrics tell me *something* is wrong;
> logs tell me *what*.** I can go from a spike on a graph straight to the exact lines that explain it."

---

### PART 8 — Security: defense in depth (~40 sec)

🖥️ **Show:** terminal. Three quick demos back-to-back.

⌨️ **Commands:**
```bash
# 1. Trivy — scan an image for vulnerabilities
trivy image --severity CRITICAL,HIGH --ignore-unfixed nginx:1.27

# 2. Kyverno — try to deploy something that breaks policy (gets BLOCKED)
kubectl run test-bad --image=nginx

# 3. RBAC — least privilege in action
kubectl auth can-i list pods   --as=system:serviceaccount:default:dev
kubectl auth can-i delete pods --as=system:serviceaccount:default:dev
kubectl auth can-i get secrets --as=system:serviceaccount:default:dev
```

🗣️ **Say:**
> "Security is layered — **defense in depth**, at three different stages.
>
> **Build time:** Trivy scans container images for known vulnerabilities before anything ships.
>
> **Deploy time:** Kyverno enforces policy at admission. Watch — if I try to deploy an image
> with an unpinned tag — [run it] — **it's rejected at the door.** It never reaches the cluster.
>
> **Runtime:** Falco watches running containers and alerts on suspicious behaviour, like a shell
> being opened inside a pod.
>
> And **RBAC** enforces least privilege — [run can-i] this dev identity can *view* resources,
> but can't delete anything or read secrets."

💡 *Tip:* The Kyverno **rejection message** in red is very satisfying — pause on it for a beat.

---

### PART 9 — Chaos engineering: prove it self-heals (~30 sec)

🖥️ **Show:** split — `kubectl get pods -l app=frontend -w` (left) + apply chaos (right).

⌨️ **Commands:**
```bash
# terminal 1
kubectl get pods -l app=frontend -w

# terminal 2
kubectl apply -f chaos/kill-frontend.yaml
```

🗣️ **Say:**
> "Finally — does it survive failure? I use **Chaos Mesh** to break things on purpose.
>
> Here I **kill the frontend pod** — [apply] — and watch: Kubernetes immediately
> schedules a replacement, and the Service keeps routing to healthy pods.
>
> I also inject **network latency** to test how the system degrades when a dependency gets slow —
> the failure mode that's far more common than a clean crash.
>
> **Self-healing, proven — not assumed.**"

---

### PART 10 — Reproducibility (~25 sec)

🖥️ **Show:** `setup.sh` open in VS Code, scrolling slowly.

🗣️ **Say:**
> "And the whole platform is **reproducible**. Cluster, ingress, GitOps, monitoring, security, chaos —
> all of it rebuilds from **one script**.
>
> When my local cluster's certificates corrupted, I didn't debug for hours —
> I rebuilt the entire environment in minutes, because every layer is declarative and in Git.
>
> **Cattle, not pets.**"

💡 *Tip:* This quietly signals maturity — interviewers love the disaster-recovery story.

---

### PART 11 — The Close (~25 sec)

🖥️ **Show:** the GitHub repo README, then slowly scroll the repo structure.

🗣️ **Say:**
> "So that's the platform: **GitOps** delivery, full **observability**, layered **security**,
> **autoscaling**, **chaos-tested** resilience, **automated canary** releases,
> and a **custom Kubernetes operator** — running **live on the cloud with real HTTPS**.
>
> Everything is documented in the repo — link in the description.
> I'd love to talk about it. Thanks for watching."

💡 *End card (optional):* your name · GitHub URL · live demo URL · LinkedIn. Hold it for 3–4 seconds.

---

## ⏱️ Runtime plan

| Part | Content | Target |
|---|---|---|
| 1 | Hook | 0:20 |
| 2 | Live cloud + HTTPS | 0:30 |
| 3 | ⭐ Custom Operator | 1:15 |
| 4 | GitOps | 0:45 |
| 5 | Canary + auto-rollback | 0:50 |
| 6 | Autoscaling + k6 | 0:50 |
| 7 | Observability | 0:35 |
| 8 | Security | 0:40 |
| 9 | Chaos | 0:30 |
| 10 | Reproducibility | 0:25 |
| 11 | Close | 0:25 |
| | **Total** | **~6:05** |

> 🎯 If you need it shorter, cut Part 10 and trim Part 8 — never cut Part 3.

---

## 🎥 Recording tips (these make it look professional)

1. **Record each part separately.** If you fumble, redo just that part. Way less stressful.
2. **Do a silent take first.** Run the commands once without talking to check timing, then record with narration.
3. **Narrate, don't read.** Glance at the script, then say it in your own words. Slight imperfection sounds *human*; robotic reading sounds worse than a stumble.
4. **Cut every wait.** Speed up (2–4×) anything that loads, pulls, or ramps. No dead air, ever.
5. **Zoom in on the money shots** — the Kyverno rejection, the AnalysisRun, the HPA replica count climbing, the padlock.
6. **Type slowly and deliberately**, or pre-type commands and just hit Enter. Nothing worse than typos on camera.
7. **Speak with confidence.** You built this. Say "I built", "I wrote", "I designed" — not "I tried to".

---

## ✂️ Editing & assembly

- **Free editors:** DaVinci Resolve (powerful, free) · CapCut (easy) · Clipchamp (built into Windows)
- **Add a title card** at the start: *"Internal Developer Platform on Kubernetes — Ayush Bansal"* (3 s)
- **Add section labels** as small text overlays when each part starts (e.g. *"Custom Operator"*) — helps skimmers
- **Background music:** optional, very quiet (-25 dB) or skip it. Never let it compete with your voice.
- **Captions:** YouTube auto-captions are fine; correct the tech terms (Kubernetes, ArgoCD, Kubebuilder).

---

## 📤 Publishing

1. **YouTube → Unlisted** (not private — unlisted lets anyone with the link watch)
2. **Title:** `Internal Developer Platform on Kubernetes — GitOps, Custom Operator, Observability & Security`
3. **Description:** one paragraph + GitHub link + live demo link + a timestamped chapter list
4. **Chapters** (paste in the description so YouTube makes clickable sections):
   ```
   0:00 Intro
   0:20 Live cloud deployment with HTTPS
   0:50 Custom Kubernetes Operator
   2:05 GitOps with ArgoCD
   2:50 Automated canary deployments
   3:40 Autoscaling under load
   4:30 Observability — metrics & logs
   5:05 Security — Trivy, Kyverno, Falco, RBAC
   5:45 Chaos engineering
   6:15 Reproducibility
   ```
5. **Then put the link:** in your **README** (top), your **resume**, and a **LinkedIn post**.

---

## ✅ Final checklist before you publish

- [ ] Audio is clear and consistent across all parts
- [ ] Text is readable when watched **on a phone**
- [ ] No secrets/passwords/tokens visible on screen (check ArgoCD password, kubeconfig, `.pem` paths!)
- [ ] Every claim in the video is true and demonstrated
- [ ] Under 7 minutes
- [ ] Link works from an incognito window (unlisted, not private)

---

*Record it in parts, stitch it together, and you'll have the artifact that turns months of work into interviews. You've got this. 🚀*
