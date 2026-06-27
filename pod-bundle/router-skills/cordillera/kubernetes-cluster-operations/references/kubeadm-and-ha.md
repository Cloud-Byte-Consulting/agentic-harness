# Self-Managed Clusters with kubeadm (incl. HA)

kubeadm is the official tool for bootstrapping best-practice, production-capable clusters on your
own machines (bare metal, VMs, on-prem, or cloud instances). You own everything the cloud would
otherwise manage: control-plane HA, etcd, certificates, the API load balancer, networking, storage,
and upgrades.

## Contents
- [Prerequisites on every node](#prerequisites-on-every-node)
- [Single control-plane bootstrap](#single-control-plane-bootstrap)
- [Joining worker nodes](#joining-worker-nodes)
- [Installing a CNI](#installing-a-cni)
- [HA: multiple control-plane nodes](#ha-multiple-control-plane-nodes)
- [Certificates](#certificates)
- [On-prem realities](#on-prem-realities)

## Prerequisites on every node

Do these on **all** control-plane and worker nodes before `kubeadm init/join`:

1. **Disable swap** (the kubelet expects it off):
   ```bash
   sudo swapoff -a
   # and comment the swap line out of /etc/fstab so it stays off after reboot
   ```
2. **Kernel modules + sysctl** for the container network:
   ```bash
   sudo modprobe overlay
   sudo modprobe br_netfilter
   cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
   net.bridge.bridge-nf-call-ip6tables = 1
   net.bridge.bridge-nf-call-iptables  = 1
   net.ipv4.ip_forward                 = 1
   EOF
   sudo sysctl --system
   ```
3. **Install a CRI container runtime** (containerd or CRI-O) and ensure its socket is up
   (`systemctl enable --now containerd`). Configure containerd to use the systemd cgroup driver,
   matching the kubelet.
4. **Install kube tools** from the community repo `pkgs.k8s.io` (the old `apt.kubernetes.io` /
   `yum.kubernetes.io` repos were frozen Sept 2023). Pin the minor version you want, e.g.:
   ```bash
   # Debian/Ubuntu (v1.30 line shown)
   curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key \
     | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
   echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] \
     https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
   sudo apt-get update && sudo apt-get install -y kubelet kubeadm kubectl
   sudo apt-mark hold kubelet kubeadm kubectl
   ```
5. Open required ports between nodes: control plane needs **6443** (API), **2379-2380** (etcd),
   **10250** (kubelet), **10257/10259** (controller-manager/scheduler); workers need **10250** and
   the NodePort range **30000-32767**.

## Single control-plane bootstrap

On the first control-plane node:
```bash
sudo kubeadm init --pod-network-cidr=10.244.0.0/16
```
- `--pod-network-cidr` must match your CNI's expectation (10.244.0.0/16 suits Flannel; Calico can
  auto-detect). 
- On success, kubeadm prints (a) the admin kubeconfig setup and (b) a `kubeadm join` command with a
  token and CA hash — **save the join command**, you'll need it for workers.

Set up kubectl for your user:
```bash
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
kubeadm runs the control-plane components (API server, controller-manager, scheduler, etcd) as
**static Pods** defined by manifests in `/etc/kubernetes/manifests/` — the kubelet watches that
directory and keeps them running.

## Joining worker nodes

After prerequisites, run the printed join command on each worker:
```bash
sudo kubeadm join <control-plane-host>:6443 --token <token> \
  --discovery-token-ca-cert-hash sha256:<hash>
```
If the token expired (24h default), mint a fresh join command on a control-plane node:
```bash
kubeadm token create --print-join-command
```
Verify from the control plane: `kubectl get nodes` (workers go `Ready` once the CNI is installed).

## Installing a CNI

The cluster's nodes stay `NotReady` until a pod network is installed. Apply one CNI (pick to match
your `--pod-network-cidr`), e.g.:
```bash
# Calico (operator-based)
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.28.0/manifests/tigera-operator.yaml
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.28.0/manifests/custom-resources.yaml
# or Flannel (expects 10.244.0.0/16)
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml
```
CNI selection and NetworkPolicy are covered in depth by kubernetes-networking.

## HA: multiple control-plane nodes

For production, run **3 (or 5) control-plane nodes** behind a load balancer.

1. **Stand up an API load balancer first.** An L4 TCP LB (HAProxy + keepalived, MetalLB, or a cloud
   LB) on a stable FQDN, balancing port 6443 across all control-plane nodes. Note: MetalLB cannot
   load-balance the API server in HA, so for bare metal use HAProxy/keepalived for the API VIP.
2. **Init the first control-plane node pointing at the LB** and upload certs so peers can pull them:
   ```bash
   sudo kubeadm init --control-plane-endpoint "k8s-api.example.com:6443" \
     --upload-certs --pod-network-cidr=10.244.0.0/16
   ```
   `--control-plane-endpoint` is **mandatory for HA and must be set at init time** — retrofitting it
   onto an existing single-node cluster is painful. The output now includes **two** join commands:
   one for control-plane nodes (with `--control-plane --certificate-key ...`) and one for workers.
3. **Join the other control-plane nodes:**
   ```bash
   sudo kubeadm join k8s-api.example.com:6443 --token <token> \
     --discovery-token-ca-cert-hash sha256:<hash> \
     --control-plane --certificate-key <cert-key>
   ```
   (The `--certificate-key` from step 2 expires in ~2h; regenerate with
   `kubeadm init phase upload-certs --upload-certs` if needed.)
4. **Join workers** with the worker join command as above.

This yields the **stacked etcd** topology (an etcd member on each control-plane node). For
**external etcd**, build a separate etcd cluster first and pass its endpoints/certs via a kubeadm
config file (`etcd.external` section) at init. See architecture.md for the trade-off.

## Certificates

kubeadm generates a PKI under `/etc/kubernetes/pki/` (CA, API server, etcd, front-proxy certs).
Client/serving certs are valid **1 year**; the CA is valid 10 years.

```bash
kubeadm certs check-expiration          # list all certs and expiry dates
sudo kubeadm certs renew all            # renew everything (then restart control-plane static pods)
```
A normal `kubeadm upgrade apply` **also renews** the certs automatically — staying on a regular
upgrade cadence keeps certs fresh. If certs expire on an idle cluster, the API server stops
accepting connections; renew, then restart the static Pods (e.g. move the manifests out and back,
or reboot the kubelet). kubelet client certs can auto-rotate (`rotateCertificates: true`).

## On-prem realities

Self-managed/on-prem means you also own:
- **Infrastructure provisioning** — automate node images with Packer; provision with Terraform /
  Cluster API / Kubespray (which itself uses kubeadm under the hood).
- **Load balancing & external access** — NodePort/LoadBalancer aren't enough on bare metal; use
  MetalLB for service LBs (but not the API VIP), plus an Ingress controller.
- **Persistent storage** — wire PVs/PVCs to real storage (Longhorn, Ceph/Rook, NFS, iSCSI/block).
- **Upgrades & scaling** — test new versions in staging; add nodes by joining them.
- **Monitoring & logging** — Prometheus/Grafana, a logging stack; control-plane logs land in
  `/var/log/` and component static-Pod logs.
- **Security/compliance** — e.g. FIPS where required (hardening lives in kubernetes-security-rbac).

Fleet/lifecycle managers like **Rancher** or **OpenShift** reduce this toil; on huge fleets,
private clouds like **OpenStack** (with Magnum) are common.
