# Local & Dev Clusters: kind, k3s, minikube

Practical configs and commands for the three common local/dev tools, plus when to pick each.
**None of these is for production** — don't expose them to the internet.

## Contents
- [Picking one](#picking-one)
- [kind](#kind-kubernetes-in-docker)
- [k3s](#k3s)
- [minikube](#minikube)
- [Quick teardown reference](#quick-teardown-reference)

## Picking one

| Need | Best choice |
|---|---|
| Multi-node, test failover/Ingress/RBAC, scriptable in CI | **kind** |
| Real lightweight cluster, edge/IoT, single binary, low overhead | **k3s** |
| Beginner-friendly, addons (dashboard, ingress, registry), VM isolation | **minikube** |
| Just need *a* cluster fast on a dev box | any; kind if Docker is present |

All three are free and open source. kind and k3s are lighter than minikube; k3s and kind give you
real multi-node behavior, minikube can too (`--nodes`).

## kind (Kubernetes in Docker)

Runs each node as a Docker (or Podman) container via Docker-in-Docker. Great for multi-node testing
on one host.

### Install
```bash
# Linux
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-amd64
chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind
# macOS
brew install kind
# Windows
choco install kind
```

### Single-node (quick)
```bash
kind create cluster --name dev          # one container = control plane + worker
```

### Multi-node from a config file
`kind-cluster.yaml`:
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
```
```bash
kind create cluster --name dev --config kind-cluster.yaml
```
kind writes the kubeconfig context (`kind-dev`) and sets it current. Verify with
`kubectl get nodes` and `kubectl cluster-info`.

### Pin a Kubernetes version (use the node image digest)
```bash
kind create cluster --name dev --config kind-cluster.yaml \
  --image kindest/node:v1.30.0
```

### HA control plane + port mappings + custom CNI
Multiple `control-plane` roles make kind auto-deploy an internal **HAProxy** container to load
balance the API across them. `extraPortMappings` expose host ports to a worker (e.g. for an Ingress
controller). `disableDefaultCNI: true` lets you install Calico/Cilium instead of the bundled kindnet.
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: "0.0.0.0"
  disableDefaultCNI: true          # then `kubectl apply` your CNI
  podSubnet: "10.240.0.0/16"
  serviceSubnet: "10.96.0.0/16"
nodes:
- role: control-plane
- role: control-plane
- role: control-plane
- role: worker
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
  - containerPort: 443
    hostPort: 443
```

### Customize control-plane / kubelet flags
kind reuses kubeadm config, so you can patch API server flags (e.g. OIDC) via `kubeadmConfigPatches`
with a `ClusterConfiguration` `apiServer.extraArgs` block.

### Notes & gotchas
- kind's bundled HAProxy load-balances **only the API server**, not worker traffic. To load balance
  Ingress across workers you run your own HAProxy/LB container on the `kind` Docker network
  (`--net=kind`).
- Includes Rancher's `local-path-provisioner` (default StorageClass `standard`) so PVCs auto-bind to
  host-path PVs — handy for testing stateful workloads.
- Simulate a node failure: `docker exec <node> systemctl stop kubelet` → node goes `NotReady`.
- Delete: `kind delete cluster --name dev`.

## k3s

A single-binary, CNCF-sandbox lightweight Kubernetes. Server + agent in one binary; small memory
footprint; ideal for edge/IoT/CI. Defaults to SQLite as the datastore (swappable for etcd/MySQL/
Postgres). Linux-first (use WSL or a Linux VM on Windows/macOS).

### Install (single node)
```bash
curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644
```
`--write-kubeconfig-mode 644` makes `/etc/rancher/k3s/k3s.yaml` readable without sudo (dev only).

### Verify
```bash
k3s --version
k3s check-config          # kernel/module/dependency diagnostics
kubectl get nodes
kubectl cluster-info
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml   # or copy into ~/.kube/config
```

### Multi-node (server + agents)
On the **server** node, grab the node token from `/var/lib/rancher/k3s/server/node-token`, then on
each **agent**:
```bash
curl -sfL https://get.k3s.io | K3S_URL=https://<server-ip>:6443 \
  K3S_TOKEN=<node-token> sh -
```

### Notes
- k3s strips non-essential components and bundles its own (Traefik ingress, ServiceLB, local-path
  storage) — you can disable them at install (`--disable traefik`, etc.).
- Uninstall: `/usr/local/bin/k3s-uninstall.sh` (server) or `k3s-agent-uninstall.sh` (agent).

## minikube

Wraps a full single-node (optionally multi-node) cluster in a VM or container. Beginner-friendly
with a rich addon system. **Note:** the minikube version and the Kubernetes version it ships are
independent (e.g. minikube 1.32 defaults to k8s 1.28); pin with `--kubernetes-version`.

### Install
```bash
# Linux
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
# Windows
winget install minikube      # or: choco install minikube
# macOS: brew install minikube
```

### Drivers
Container drivers (`docker`, `podman`) or VM drivers (KVM2, VirtualBox, Hyper-V, QEMU, Hyperkit,
Parallels, VMware). Set a default or pass per-start:
```bash
minikube config set driver docker
minikube start --driver=docker --kubernetes-version=v1.30.0 --memory=8000m --cpus=2
```

### Multi-node and HA
```bash
minikube start --nodes 3 --driver=docker          # 1 control plane + 2 workers
minikube start --nodes 5 --ha=true --cni calico \  # 3 control planes + 2 workers
  --cpus=2 --memory=2g --kubernetes-version=v1.30.0 --container-runtime=containerd
```

### Multiple independent clusters (profiles)
```bash
minikube start --profile cluster-a --driver=docker
minikube start --profile cluster-b --driver=virtualbox
minikube stop   --profile cluster-a
minikube delete --profile cluster-a
```

### Lifecycle
```bash
minikube status
minikube pause / minikube unpause        # freeze/resume quickly
minikube stop                            # keep state
minikube delete                          # destroy (irrecoverable)
```
Config changes (`minikube config set cpus 4`) take effect only after `minikube delete` + `start`.
minikube auto-writes the `minikube` kubeconfig context and enables `storage-provisioner` +
`default-storageclass` addons.

## Quick teardown reference

| Tool | Delete |
|---|---|
| kind | `kind delete cluster --name <name>` |
| k3s | `/usr/local/bin/k3s-uninstall.sh` |
| minikube | `minikube delete [--profile <name>]` |
