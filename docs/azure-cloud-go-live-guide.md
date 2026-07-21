# ☁️ Cloud Go-Live Guide — Deploy the Platform on Azure with HTTPS

> **Purpose:** A complete, beginner-friendly runbook to take a Kubernetes platform
> from *local* to a *live, HTTPS-secured website on the internet* — on a free Azure
> Student VM. Written so you can redo the whole thing a year from now, with zero help.
>
> **What you'll end up with:** a real cloud server running Kubernetes (k3s), an app
> reachable at a real domain over `https://` with a valid padlock 🔒.
>
> **Reference values used in this guide (yours will differ):**
> - Public IP: `20.244.31.87`
> - Domain: `ayush-idp.duckdns.org`
> - VM user: `azureuser`
> - SSH key file: `idp-server_key.pem`
>
> Whenever you see `<YOUR-IP>`, `<YOUR-DOMAIN>`, `<YOUR-EMAIL>` → replace with your own.

---

## Table of Contents
1. [The big picture](#1-the-big-picture)
2. [Concepts glossary (read once)](#2-concepts-glossary-read-once)
3. [Phase 1 — Create the cloud server (Azure VM)](#3-phase-1--create-the-cloud-server-azure-vm)
4. [Phase 2 — SSH in + install k3s](#4-phase-2--ssh-in--install-k3s)
5. [Phase 3 — Install tooling + the Ingress controller](#5-phase-3--install-tooling--the-ingress-controller)
6. [Phase 4 — Deploy an app + get a real domain](#6-phase-4--deploy-an-app--get-a-real-domain)
7. [Phase 5 — Real HTTPS (cert-manager + Let's Encrypt)](#7-phase-5--real-https-cert-manager--lets-encrypt)
8. [Deploying additional apps](#8-deploying-additional-apps)
9. [Stopping & restarting the server (save money)](#9-stopping--restarting-the-server-save-money)
10. [The full traffic pipeline (how a request flows)](#10-the-full-traffic-pipeline-how-a-request-flows)
11. [Troubleshooting](#11-troubleshooting)

---

## 1. The big picture

**Goal:** run the same Kubernetes manifests you built locally, but on a *real server*
on the internet, at a *real domain*, with *real HTTPS*.

```
   LOCAL (before):  runs on your laptop (kind), reachable only at *.localtest.me
   CLOUD (after):   runs on an Azure server (k3s), at your domain, with HTTPS 🔒
```

**The 5 phases:**
```
   Phase 1 — Create a cloud server (Azure VM)
   Phase 2 — SSH in + install k3s (real Kubernetes)
   Phase 3 — Install tools + the ingress controller (front door)
   Phase 4 — Deploy an app + point a domain at the server
   Phase 5 — Real HTTPS (cert-manager + Let's Encrypt)
```

**Key idea:** almost nothing about your *work* changes. Your YAML manifests, GitOps,
etc. all run unchanged — you're just running them on a better machine. That's the
whole payoff of "everything as code."

---

## 2. Concepts glossary (read once)

| Term | Simple meaning |
|---|---|
| **VM (Virtual Machine)** | A computer you rent in a datacenter. Always on, has a public IP. |
| **Public IP** | The server's address on the internet (e.g. `20.244.31.87`). |
| **Private IP** | The server's address inside the cloud's internal network (e.g. `10.0.0.4`). |
| **SSH** | Secure Shell — securely log into a remote server's terminal from your laptop. |
| **SSH key (`.pem`)** | A secret file used instead of a password to log in. Guard it. |
| **NSG / Security List** | The cloud **firewall** — controls which ports are open (22, 80, 443). |
| **k3s** | A lightweight, production-grade Kubernetes for real servers (kind was for local Docker). |
| **kubectl** | The command to control Kubernetes. |
| **KUBECONFIG** | Tells `kubectl` which cluster/credentials file to use. |
| **Helm** | The "app store" for Kubernetes — installs bundled apps with one command. |
| **Ingress** | A smart front door that routes web traffic by hostname to the right app. |
| **ingress-nginx** | The program (controller) that enforces Ingress rules using nginx. |
| **DNS** | The internet's phone book: translates a name → an IP address. |
| **A record** | A DNS rule mapping a hostname → an IPv4 address. |
| **DuckDNS** | A free dynamic-DNS service; gives you a free subdomain you can point at any IP. |
| **TLS / HTTPS** | Encrypts web traffic + proves the server's identity via a certificate (the 🔒). |
| **Certificate Authority (CA)** | A trusted org that issues certificates. Browsers trust them. |
| **Let's Encrypt** | A free, trusted, automated CA. |
| **cert-manager** | A Kubernetes controller that auto-gets & auto-renews certificates. |
| **ACME / HTTP-01 challenge** | How cert-manager proves to Let's Encrypt that you own the domain. |
| **Port 80 / 443** | 80 = HTTP (unencrypted). 443 = HTTPS (encrypted). |

---

## 3. Phase 1 — Create the cloud server (Azure VM)

### 3.1 Sign up
- **Azure for Students:** https://azure.microsoft.com/free/students → sign in with a
  student email. Gives **$100 credit, no credit card required.**
- *(Bonus)* **GitHub Student Pack:** https://education.github.com/pack → gives a free
  real domain for a year + more credits.

### 3.2 Create the Virtual Machine
Azure Portal → search **"Virtual machines"** → **+ Create → Azure virtual machine**.

| Field | Value | Why |
|---|---|---|
| Subscription | Azure for Students | your free credit |
| Resource group | **Create new** → `idp-rg` | a "folder" for all related resources |
| VM name | `idp-server` | a label |
| Region | Central India (or nearest) | which datacenter |
| Image | **Ubuntu Server 24.04 LTS - x64 Gen2** | the OS |
| VM architecture | **x64** | standard architecture |
| Size | **Standard_B2as_v2** (2 vCPU, 8 GiB, ~$36/mo) | cheapest 8 GB burstable option |
| Authentication type | **SSH public key** | key-based login (secure) |
| Username | `azureuser` | your login name on the server |
| SSH key source | **Generate new key pair**, name `idp-server_key` | Azure makes the key for you |
| Public inbound ports | **Allow selected** → **22, 80, 443** | firewall: SSH + HTTP + HTTPS |
| Azure Spot | **unchecked** | spot VMs can be shut down randomly |

→ **Review + create** → **Create**.

### 3.3 ⚠️ Download the private key
A popup appears: **"Download private key and create resource"** → click it and **SAVE
the `.pem` file** (e.g. `Downloads\idp-server_key.pem`). **You can NEVER re-download it.**

### 3.4 Get your details
VM **Overview** page → note the **Public IP address** (`20.244.31.87`).

### 3.5 (Recommended) Make the public IP Static
By default Azure IPs can change on restart, which breaks your DNS. **Standard-SKU IPs
are static by default** — check: Portal → "Public IP addresses" → your IP → **SKU:
Standard** means it's already static ✅. If it's Basic/Dynamic:
Configuration → **Assignment: Static** → Save.

---

## 4. Phase 2 — SSH in + install k3s

### 4.1 What is SSH?
SSH securely logs you into the server's terminal from your laptop. You authenticate
with your **private key** (the `.pem`) instead of a password.

### 4.2 Fix the key permissions
SSH refuses a key that other users can read. Lock it down.

**From Windows PowerShell:**
```powershell
icacls "C:\Users\<YOU>\Downloads\idp-server_key.pem" /inheritance:r /grant:r "$($env:USERNAME):R"
```
*(`/inheritance:r` removes inherited perms; `/grant:r you:R` gives only you read access.)*

**OR from WSL/Linux:**
```bash
cp "/mnt/c/Users/<YOU>/Downloads/idp-server_key.pem" ~/idp-server_key.pem
chmod 600 ~/idp-server_key.pem
```
*(`chmod 600` = only the owner can read/write. SSH requires this.)*

### 4.3 SSH into the server
```bash
ssh -i "<path-to>/idp-server_key.pem" azureuser@<YOUR-IP>
```
- `-i <file>` → the identity (private key) file.
- `azureuser` → the username; `@<YOUR-IP>` → the server address.
- First time it asks `Are you sure you want to continue connecting?` → type **`yes`**.
- You land at `azureuser@idp-server:~$` = **you're inside the server.** Every command
  now runs *on the server*.

**Verify:**
```bash
hostname       # -> idp-server
free -h        # -> ~8 GB RAM
```

### 4.4 Install k3s (real Kubernetes)
```bash
curl -sfL https://get.k3s.io | sh -s - --disable traefik
```
- Installs k3s as a background service in ~30s (a full Kubernetes cluster).
- `--disable traefik` → k3s ships its own ingress (Traefik); we disable it because we
  use **ingress-nginx** (your apps say `ingressClassName: nginx`). Avoids a port 80/443 clash.

### 4.5 Set up kubectl for your user
k3s's config is root-only by default. Copy it so `azureuser` can use `kubectl`:
```bash
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
export KUBECONFIG=~/.kube/config
echo 'export KUBECONFIG=~/.kube/config' >> ~/.bashrc
```
- Copies k3s's cluster config to `~/.kube/config` and makes you the owner.
- `export KUBECONFIG=...` tells kubectl to use that file (k3s's kubectl otherwise
  defaults to the root-only `/etc/rancher/k3s/k3s.yaml` → "permission denied").
- The `echo ... >> ~/.bashrc` makes it permanent across logins.

### 4.6 Verify the cluster is alive
```bash
kubectl get nodes        # -> idp-server   Ready   control-plane,master
kubectl get pods -A      # -> k3s's core components, all Running
```

---

## 5. Phase 3 — Install tooling + the Ingress controller

### 5.1 Install git + helm (on the server)
```bash
sudo apt-get update
sudo apt-get install -y git
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```
- `git` → to clone your project repo. `helm` → to install packaged apps. (kubectl came with k3s.)

### 5.2 Clone your project onto the server
```bash
git clone https://github.com/<YOUR-USER>/k8s-project.git
cd k8s-project
```
*(Your manifests are now on the server — the same files you built locally.)*

### 5.3 Install the Ingress controller (the front door)
```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
```
- This is the **cloud** variant (not the kind one). It creates a **LoadBalancer** service.
- On k3s, the built-in **ServiceLB (klipper)** binds the server's host ports 80/443 to
  ingress-nginx — so traffic to your public IP:80/443 reaches your apps.

**Verify:**
```bash
kubectl get pods -n ingress-nginx     # controller should be Running
kubectl get svc  -n ingress-nginx     # controller Service shows an EXTERNAL-IP (the node IP)
```

---

## 6. Phase 4 — Deploy an app + get a real domain

### 6.1 Deploy an app (example: podinfo)
```bash
kubectl apply -f apps/podinfo/deployment.yaml
kubectl apply -f apps/podinfo/service.yaml
kubectl get pods -l app=podinfo       # wait for Running
```

### 6.2 (Optional) Instant URL via nip.io — before you have a domain
`nip.io` is a free DNS trick: `anything.<YOUR-IP>.nip.io` resolves to `<YOUR-IP>`.
Create an Ingress using it:
```bash
kubectl apply -f - <<'EOF'
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: podinfo
spec:
  ingressClassName: nginx
  rules:
    - host: podinfo.<YOUR-IP>.nip.io
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: podinfo
                port:
                  number: 9898         # <-- podinfo's Service port
EOF
```
- `kubectl apply -f - <<'EOF' ... EOF` = a **heredoc**: pipes the YAML straight into
  kubectl, no file needed.
- `number: 9898` = the port podinfo's **Service** exposes. The Ingress always points at
  the **Service's** port (not the pod directly). Different apps use different numbers.

Visit: `http://podinfo.<YOUR-IP>.nip.io`

### 6.3 Get a real domain (DuckDNS)
1. https://www.duckdns.org → sign in.
2. Create a subdomain, e.g. `ayush-idp` → gives `ayush-idp.duckdns.org`.
3. Set its **current ip** to your **server's** IP (`<YOUR-IP>`) → **update ip**.
   - This creates an **A record**: name → IP.
   - **Wildcard bonus:** `*.ayush-idp.duckdns.org` (any subdomain) also resolves to your
     IP automatically — so you can add `shop.`, `grafana.`, etc. with no extra DNS setup.

**Verify (from the server):**
```bash
nslookup <YOUR-DOMAIN>        # should return <YOUR-IP>
```

### 6.4 Update the Ingress to use the real domain
```bash
kubectl apply -f - <<'EOF'
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: podinfo
spec:
  ingressClassName: nginx
  rules:
    - host: <YOUR-DOMAIN>
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: podinfo
                port:
                  number: 9898
EOF
```
Visit: `http://<YOUR-DOMAIN>` (still plain HTTP — HTTPS is next).

---

## 7. Phase 5 — Real HTTPS (cert-manager + Let's Encrypt)

### 7.1 How it works (simple)
- **cert-manager** asks **Let's Encrypt** (a free trusted CA) for a certificate.
- Let's Encrypt says "prove you own the domain" → **HTTP-01 challenge**: cert-manager
  serves a secret token at `http://<domain>/.well-known/acme-challenge/...`.
- Let's Encrypt visits the domain, sees the token → confirms ownership → issues the cert.
- cert-manager stores the cert in a Kubernetes **secret**; **ingress-nginx** uses it to
  serve HTTPS on port **443**. The browser trusts it → shows the 🔒.

### 7.2 Install cert-manager
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
kubectl get pods -n cert-manager      # 3 pods, all Running
```

### 7.3 Create a ClusterIssuer (connect to Let's Encrypt)
Replace `<YOUR-EMAIL>` with your real email.
```bash
kubectl apply -f - <<'EOF'
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: <YOUR-EMAIL>
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
EOF
```
- **`letsencrypt-prod`** = the name we give this issuer. "prod" because the `server:` URL
  is Let's Encrypt's **production** endpoint (real, trusted certs).
- There's also a **staging** endpoint (`acme-staging-v02...`) for testing — it issues
  *untrusted* certs but has generous rate limits. Use staging to debug, prod for real.
- ⚠️ **Production is rate-limited (~5 certs/domain/week).** Don't retry repeatedly — fix
  errors first (use staging if you need to iterate a lot).

### 7.4 Update the Ingress to request HTTPS
Replace `<YOUR-DOMAIN>` (both places).
```bash
kubectl apply -f - <<'EOF'
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: podinfo
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - <YOUR-DOMAIN>
      secretName: podinfo-tls
  rules:
    - host: <YOUR-DOMAIN>
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: podinfo
                port:
                  number: 9898
EOF
```
- The **annotation** tells cert-manager "get a cert for this ingress."
- The **`tls:` block** enables HTTPS and says "store the cert in the `podinfo-tls` secret."

### 7.5 Watch the cert get issued (~1–2 min)
```bash
kubectl get certificate               # wait for READY: True
# if slow, inspect:
kubectl describe certificate podinfo-tls
```

### 7.6 Visit your secure site 🎉
```
https://<YOUR-DOMAIN>
```
The 🔒 padlock = real, trusted, encrypted HTTPS.

**Note on ports:** HTTPS uses **443** (HTTP uses 80). The browser auto-connects to 443
for `https://`. ingress-nginx **terminates TLS** on 443 (decrypts with the cert), then
forwards plain HTTP to the app internally. It also **redirects http (80) → https (443)**,
so all traffic ends up encrypted.

---

## 8. Deploying additional apps

**The same 3 steps for ANY app** (thanks to the platform + DuckDNS wildcard):

```
   1. Apply its manifests            (deployment + service)
   2. Create an Ingress              (a subdomain + cert-manager annotation + tls)
   3. cert-manager auto-issues HTTPS (visit https://<subdomain>.<YOUR-DOMAIN>)
```

**Example — the Online Boutique shop:**
```bash
# 1. Deploy its manifests (already in your repo)
kubectl apply -f apps/online-boutique/kubernetes-manifests.yaml

# 2. Ingress for its frontend (subdomain + HTTPS)
kubectl apply -f - <<'EOF'
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: online-boutique
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - shop.<YOUR-DOMAIN>
      secretName: boutique-tls
  rules:
    - host: shop.<YOUR-DOMAIN>
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: frontend
                port:
                  number: 80
EOF
```
Then visit `https://shop.<YOUR-DOMAIN>`.

> ⚠️ **Resource note:** an 8 GB server can't run *everything* at once. Heavy apps (like
> the 11-service shop) may leave pods `Pending` (out of memory). Check with
> `kubectl get pods`; free memory by deleting other apps if needed.

---

## 9. Stopping & restarting the server (save money)

**Billing is per-hour while the VM is RUNNING.** Stop it when idle.

### Stop (save credit)
Azure Portal → VM `idp-server` → **Overview** → **Stop** (top toolbar).
- Status becomes **"Stopped (deallocated)"** = compute billing stops. ✅
- ⚠️ Use the **portal** Stop button — NOT `sudo poweroff` inside the VM (that keeps billing).

### Restart
Azure Portal → VM → **Start**. Wait ~2 min.
- **Everything auto-comes-back**: k3s is a systemd service (auto-starts), and all your
  apps/ingress/cert are saved on disk. **Nothing to reinstall.**
- If your IP is **Static** (Standard SKU) → the domain still works, do nothing.
- If your IP is **Dynamic** → it may change; update DuckDNS with the new IP after each start.

### While stopped
`https://<YOUR-DOMAIN>` won't load (server is off). Start the VM before sharing the link.

**Cost math (B2as_v2):** running 24/7 ≈ $36/mo. Stopped when idle → credit lasts months.

---

## 10. The full traffic pipeline (how a request flows)

```
   Browser: https://<YOUR-DOMAIN>   (port 443, the HTTPS default)
        │
        │  ① DNS: DuckDNS resolves <YOUR-DOMAIN> -> <YOUR-IP>
        ▼
   Azure public IP (<YOUR-IP>)
        │  ② Azure NSG (firewall) allows 443; NAT -> VM private IP (10.0.0.4)
        ▼
   The server (host ports 80/443)
        │  ③ k3s ServiceLB (klipper) binds host 80/443 -> ingress-nginx
        ▼
   ingress-nginx
        │  ④ TLS termination: decrypts using the cert (from the podinfo-tls secret)
        │  ⑤ routes by hostname/path to the right Service
        ▼
   App Service (e.g. podinfo :9898) -> Pod -> the page 🎉
```

Two "gatekeepers" must both allow a port: the **Azure firewall (NSG)** and **k3s klipper**.

---

## 11. Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `ssh: permissions are too open` | key file readable by others | re-run `chmod 600` (Linux) or `icacls ... /inheritance:r` (Windows) |
| `ssh: connection timed out` | firewall/port 22 closed, or wrong IP | ensure NSG allows 22; check the public IP |
| `kubectl: permission denied /etc/rancher/k3s/k3s.yaml` | using k3s's root-only config | `export KUBECONFIG=~/.kube/config` (after copying it) |
| ingress-nginx Service has no EXTERNAL-IP | LoadBalancer not wired | on k3s it uses klipper; wait a minute; check `kubectl get pods -n ingress-nginx` |
| `nslookup <domain>` doesn't return your IP | DNS not propagated / wrong IP in DuckDNS | wait 1–2 min; re-check the IP you set in DuckDNS |
| Certificate stuck `READY: False` | ACME challenge failing | `kubectl describe certificate <name>`; ensure port 80 open + domain resolves to server; check you didn't hit Let's Encrypt rate limits (use staging) |
| Site works on `http` but not `https` | cert not issued yet | wait for `kubectl get certificate` → `READY: True` |
| Pods stuck `Pending` | out of memory (8 GB limit) | delete unused apps; deploy fewer/lighter workloads |
| Domain breaks after VM restart | dynamic public IP changed | make the IP **Static**, or update DuckDNS with the new IP |
| Azure keeps billing after "shutdown" | shut down inside VM, not deallocated | use the **portal Stop** button → status must say "Stopped (deallocated)" |

---

## Quick command cheat-sheet (the whole thing, in order)

```bash
# --- On your laptop: connect ---
icacls "<path>\idp-server_key.pem" /inheritance:r /grant:r "$($env:USERNAME):R"   # Windows
ssh -i "<path>/idp-server_key.pem" azureuser@<YOUR-IP>

# --- On the server: install Kubernetes ---
curl -sfL https://get.k3s.io | sh -s - --disable traefik
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
echo 'export KUBECONFIG=~/.kube/config' >> ~/.bashrc && export KUBECONFIG=~/.kube/config
kubectl get nodes

# --- Tools + repo + ingress ---
sudo apt-get update && sudo apt-get install -y git
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
git clone https://github.com/<YOUR-USER>/k8s-project.git && cd k8s-project
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml

# --- Deploy an app ---
kubectl apply -f apps/podinfo/deployment.yaml -f apps/podinfo/service.yaml
# (then create the Ingress via the heredoc in Phase 4/5)

# --- HTTPS ---
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
# (then the ClusterIssuer + TLS Ingress heredocs in Phase 5)
kubectl get certificate     # wait for READY: True
```

---

*End of guide. If future-you follows Phases 1→5, you'll have a live, HTTPS-secured app
on a cloud Kubernetes cluster — no help needed. 🚀*
