---
name: kubernetes-cluster-operations
description: >-
  Stand up, operate, and troubleshoot Kubernetes clusters. Use for cluster architecture
  (kube-apiserver, etcd, scheduler, controller-manager, kubelet, kube-proxy, CRI runtime),
  local and dev clusters (KinD, K3s, minikube), production bootstrap with kubeadm including HA
  control planes, managed clusters (EKS, AKS, GKE and Autopilot, LKE, DigitalOcean, OpenShift)
  and how to choose, kubectl mastery (contexts, kubeconfig, jsonpath, krew), cluster upgrades
  and version-skew policy, cordon/drain and node maintenance, certificate rotation, etcd
  snapshot backup and restore, and troubleshooting NotReady nodes or down control-plane
  components (crictl, logs). Trigger whenever the user creates or upgrades a cluster, adds or
  drains nodes, picks a distro or managed service, backs up etcd, switches kube contexts, or
  debugs apiserver/etcd/kubelet failures - even without saying Kubernetes. For node
  autoscaling see kubernetes-autoscaling-scheduling; for app objects see kubernetes-workloads.
---

# Kubernetes Cluster Operations

This skill equips Claude to choose, bootstrap, operate, upgrade, back up, and troubleshoot
Kubernetes clusters — the control plane, the nodes, and the lifecycle around them — across
local, on-prem, and managed-cloud environments.

## When to use this skill

- Picking a cluster tool: local dev (kind / k3s / minikube), self-managed prod (kubeadm),
  or managed (EKS / AKS / GKE / LKE / DigitalOcean / OpenShift).
- Bootstrapping a cluster: `kubeadm init/join`, `kind create cluster`, `eksctl create cluster`,
  `az aks create`, `gcloud container clusters create`, GKE Autopilot.
- Standing up an HA control plane (stacked vs external etcd) and a load balancer in front of the API.
- Day-2 ops: upgrades and version-skew, `kubectl drain`/`cordon`, node add/remove, cert rotation.
- etcd disaster recovery: `etcdctl snapshot save` / `snapshot restore`.
- Debugging: `NotReady` nodes, apiserver/etcd/kubelet down, control-plane logs, `crictl`.
- kubectl/kubeconfig fluency: contexts, namespaces, jsonpath/custom-columns, `kubectl explain`, krew.

## Core concepts

**Two node roles.** A cluster has **control-plane nodes** (run the management components) and
**worker / compute nodes** (run your Pods). You never talk to workers directly — you submit
desired state to the control plane, which reconciles reality toward it.

**Control-plane components** (all reachable only through, or only talking to, the API server):

| Component | Role |
|---|---|
| `kube-apiserver` | The only entry point. A stateless REST API; the **only** component that reads/writes etcd. Horizontally scalable behind an L7/L4 load balancer. |
| `etcd` | Distributed key-value store holding **all** cluster state. Lose etcd without a backup → cluster is unrecoverable. Listens on 2379 (client) / 2380 (peer). Uses Raft consensus; needs an **odd** member count (3 or 5) for HA. |
| `kube-scheduler` | Watches for unscheduled Pods (empty `.spec.nodeName`), picks a node honoring requests/affinity/taints, writes the binding. |
| `kube-controller-manager` | Runs reconciliation loops (node, replicaset, deployment, job, endpoints, serviceaccount controllers, …) to drive actual → desired. |
| `cloud-controller-manager` | Cloud-specific loops (node lifecycle, routes, LoadBalancer services). Absent on bare-metal/local clusters. |

**Node components** (on every node, including control-plane nodes):

| Component | Role |
|---|---|
| `kubelet` | Node agent. Runs as a host systemd service (never in a container). Watches the API for Pods bound to its node and drives the container runtime via the CRI socket. |
| `kube-proxy` | Programs iptables/IPVS rules implementing the Service abstraction. (Some CNIs, e.g. Cilium, replace it.) |
| Container runtime | containerd or CRI-O via the **CRI**. Docker's `dockershim` was removed in **v1.24** — use a CRI runtime. Container *images* built with Docker still run fine (OCI). |

**Request flow** (e.g. `kubectl run nginx`): kubectl forges an HTTPS request → kube-apiserver
(authn/authz/admission) → writes the object to etcd → scheduler assigns a node → that node's
kubelet sees the bound Pod → pulls image & starts container via CRI. Everything is a watch/reconcile
loop, not a push.

**HA topologies** (see `references/kubeadm-and-ha.md`):
- *Single-node* (kind/minikube default): control plane + worker on one host. Dev only.
- *Single control-plane, many workers*: HA for workloads, **but the control plane and its single
  etcd are a single point of failure.** Not production-grade.
- *Multi control-plane (3+)* behind a load balancer: each control-plane node runs an etcd member
  forming a Raft quorum. This is the production target. Stacked etcd (etcd on the control-plane
  nodes) is simplest; external etcd (dedicated etcd cluster) isolates failure domains.

**Managed vs self-managed.** Managed services (EKS/AKS/GKE/LKE/DO/OpenShift) run and patch the
control plane and etcd for you; you manage worker node pools, workloads, and observability.
"Serverless" tiers (GKE Autopilot, EKS Fargate, AKS Virtual Kubelet/ACI) hide the nodes too.

## Workflow / how to approach cluster tasks

### 1. Choosing the right cluster

Decision order — match the *job* to the tool:

- **Local dev / CI, single laptop** → **kind** (multi-node via Docker-in-Docker, fast, scriptable;
  best for testing failover, Ingress, RBAC) or **minikube** (VM or container driver, addons, HA flag)
  or **k3s** (single binary, real cluster, great for edge/IoT/CI; defaults to SQLite, swappable).
  Never expose any of these to the internet or run them in production.
- **Self-managed prod / on-prem / bare metal** → **kubeadm** (official, best-practice bootstrap).
  Wrap with kops/Kubespray/Cluster API for fleet automation. You own upgrades, certs, etcd, LB, storage.
- **Cloud, want it managed** → **EKS** (AWS, use `eksctl`), **AKS** (Azure, `az aks`),
  **GKE** (GCP, `gcloud`; most polished, Autopilot for hands-off). **LKE/DigitalOcean** for
  simple, cheap, predictable billing (often free control plane). **OpenShift/ROSA** for
  enterprise PaaS with vendor support (costly; etcd encryption, integrated CI/CD).
- See `references/managed-clusters.md` and `references/local-clusters-kind-k3s-minikube.md`.

### 2. Bootstrapping

Pick the matching quickstart from the reference files. Key one-liners:

```bash
# kind: multi-node from a config file (control plane + 2 workers)
kind create cluster --name dev --config kind-cluster.yaml

# k3s: single-node, world-readable kubeconfig for dev
curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644

# minikube: HA, 3 control-plane + 2 workers
minikube start --nodes 5 --ha=true --cni calico --container-runtime=containerd

# kubeadm: init first control-plane node (note the printed join commands!)
sudo kubeadm init --control-plane-endpoint "LB_FQDN:6443" --upload-certs \
  --pod-network-cidr=10.244.0.0/16

# EKS / AKS / GKE
eksctl create cluster --name prod --region us-west-2 --nodes 3 --managed
az aks create -g rg-prod -n aks-prod --node-count 3 --generate-ssh-keys
gcloud container clusters create prod --num-nodes=2 --region=us-central1
```

After bootstrap **always** verify: `kubectl get nodes` (all `Ready`),
`kubectl get --raw='/readyz?verbose'`, and `kubectl cluster-info`.

### 3. Configuring kubectl / kubeconfig

- kubeconfig (default `~/.kube/config`, override with `$KUBECONFIG` or `--kubeconfig`) has three
  sections that bind together: **clusters** (API URL + CA), **users** (credentials), **contexts**
  (cluster+user+namespace tuple). Managed tools write a context for you.
- Daily commands: `kubectl config get-contexts`, `... use-context NAME`,
  `... set-context --current --namespace=NS`. Pull fresh managed creds with
  `aws eks update-kubeconfig` / `az aks get-credentials` / `gcloud container clusters get-credentials`.
- Master output flags: `-o wide`, `-o yaml`, `-o jsonpath=...`, `-o custom-columns=...`,
  `kubectl explain <res>.<field>`. Full catalog in `references/kubectl-and-kubeconfig.md`.

### 4. Upgrades (respect version skew!)

Skew rules (current Kubernetes): kubelet may be **up to 3 minor versions behind** the API server
(never ahead); kubectl within **±1** minor of the API server; control-plane components within
1 minor of the API server. **Upgrade order: control plane first, then worker nodes. Never skip a
minor version** (1.29 → 1.30 → 1.31, not 1.29 → 1.31). For kubeadm: `kubeadm upgrade plan`,
`kubeadm upgrade apply vX.Y.Z` on the first control-plane node, `kubeadm upgrade node` on the rest,
then drain → upgrade kubelet/kubectl → uncordon each node. Full runbook in
`references/upgrades-and-version-skew.md`.

### 5. Node maintenance

```bash
kubectl cordon <node>                                   # stop new scheduling
kubectl drain <node> --ignore-daemonsets --delete-emptydir-data  # evict gracefully (honors PDBs)
# ... patch / reboot / upgrade the node ...
kubectl uncordon <node>                                 # return to service
```
`drain` respects PodDisruptionBudgets — it can block, which is correct. See
`references/cluster-troubleshooting.md` for capacity planning and node add/remove.

### 6. etcd backup & restore

Treat etcd as the crown jewels. Take regular snapshots; test restores.

```bash
ETCDCTL_API=3 etcdctl snapshot save /backup/etcd-$(date +%F).db \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key
```
Restore is a deliberate, cluster-down operation — see `references/etcd-backup-restore.md`.

### 7. Troubleshooting

Triage from the outside in: `kubectl get nodes` → describe the bad node → SSH and check
`systemctl status kubelet` + `journalctl -u kubelet` → control-plane static-pod logs in
`/var/log/` or `kubectl logs -n kube-system <component-pod>` → `crictl ps -a` / `crictl logs`
to inspect the runtime directly when the kubelet/API path is broken. Decision trees, common
failures (NotReady, apiserver/etcd down, cert expiry), and crictl recipes live in
`references/cluster-troubleshooting.md`.

## Common pitfalls & anti-patterns

- **Running a dev tool in production.** kind/k3s/minikube and KinD's bundled HAProxy are not
  production-grade. Use kubeadm or a managed service for real workloads.
- **Single etcd / single control plane in prod.** One corrupted etcd with no backup = total loss.
  Run 3+ (odd) control-plane nodes and snapshot etcd on a schedule.
- **No load balancer in front of multiple API servers.** kubeconfig points at one host; without an
  LB + stable FQDN you lose HA. Set `--control-plane-endpoint` to the LB at `kubeadm init` time
  (you cannot easily add it later).
- **Skipping minor versions or upgrading nodes before the control plane.** Both violate skew policy
  and break the cluster. Control plane first, one minor at a time.
- **Editing etcd directly or mutating objects outside the API server.** The API server is the sole
  writer to etcd; bypassing it corrupts state.
- **Mounting the container-runtime socket into a Pod.** Bypasses Kubernetes security and is a classic
  attacker target. (Hardening details belong to kubernetes-security-rbac.)
- **Assuming Docker is still the runtime.** dockershim was removed in v1.24; clusters use containerd/CRI-O.
- **Setting a default StorageClass in prod by reflex.** A silent default can bind PVCs to the wrong
  backend; many teams prefer no default so misconfigured PVCs fail loudly.
- **Treating `kubectl get componentstatuses` as truth.** It's deprecated; prefer `/readyz`, `/livez`,
  and direct component logs.

## Reference files

- `references/architecture.md` — deep dive on every component, request flow, watch/reconcile model,
  HA topologies, stacked vs external etcd. Read when explaining internals or designing a topology.
- `references/local-clusters-kind-k3s-minikube.md` — full configs and command sets for kind, k3s,
  minikube, including multi-node and HA. Read when standing up a dev/CI cluster.
- `references/kubeadm-and-ha.md` — kubeadm prerequisites, single + HA bootstrap, join flows,
  CNI install, certificates. Read for self-managed / on-prem clusters.
- `references/managed-clusters.md` — EKS/eksctl, AKS/az, GKE/gcloud + Autopilot, LKE, DigitalOcean,
  OpenShift/ROSA: selection matrix, what the provider manages, IAM/auth, teardown. Read for cloud.
- `references/kubectl-and-kubeconfig.md` — kubeconfig structure, contexts/namespaces, output &
  jsonpath, verbs, plugins/krew, completion. Read for kubectl fluency questions.
- `references/upgrades-and-version-skew.md` — exact skew policy and the kubeadm/managed upgrade
  runbooks. Read before any version change.
- `references/etcd-backup-restore.md` — snapshot save/restore, verification, scheduling. Read for
  backup or disaster recovery.
- `references/cluster-troubleshooting.md` — triage trees, component-down playbooks, crictl,
  capacity planning, on-prem realities. Read when something is broken.
