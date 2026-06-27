# Cluster Upgrades & Version-Skew Policy

Upgrading a cluster wrong breaks it. This is the rulebook plus the runbooks.

## Contents
- [Version-skew policy](#version-skew-policy)
- [Golden rules](#golden-rules)
- [kubeadm upgrade runbook](#kubeadm-upgrade-runbook)
- [Managed-cluster upgrades](#managed-cluster-upgrades)
- [Pre-upgrade checklist](#pre-upgrade-checklist)

## Version-skew policy

Kubernetes versions are `major.minor.patch` (e.g. v1.30.2). Supported skew between components:

| Component | Allowed skew vs kube-apiserver |
|---|---|
| Other control-plane components (scheduler, controller-manager) | within **1 minor** below; same patch line in HA |
| **kubelet** | up to **3 minor versions below** the API server; **never newer** than the API server |
| **kube-proxy** | matches the kubelet's allowance (up to 3 minor below; never newer than the API server) |
| **kubectl** | within **±1 minor** of the API server |
| In HA, a second API server vs another | within **1 minor** of each other |

Consequence: the **control plane is always upgraded first**, then kubelets/workers catch up. Because
the kubelet can trail by 3 minors, you have room, but you still upgrade the control plane ahead of
the nodes — never the reverse.

## Golden rules

1. **Control plane first, nodes second.** Upgrade all control-plane nodes to the new minor, then
   roll the worker nodes.
2. **One minor version at a time.** v1.29 → v1.30 → v1.31. **Never skip a minor** (no 1.29 → 1.31).
   Patch bumps within a minor (1.30.2 → 1.30.6) are fine to do directly.
3. **kubelet/kubectl never ahead of the API server.**
4. **Back up etcd before you start** (see etcd-backup-restore.md).
5. **Drain nodes before upgrading their kubelet**, uncordon after.
6. **Test the target version in staging first**, especially for removed/deprecated APIs.

## kubeadm upgrade runbook

Upgrading from, say, v1.29 to v1.30. Replace versions as appropriate.

### 1. First control-plane node
```bash
# Find the available target patch for the desired minor
sudo apt-get update
apt-cache madison kubeadm | grep 1.30        # e.g. 1.30.2-1.1

# Upgrade kubeadm itself
sudo apt-mark unhold kubeadm
sudo apt-get install -y kubeadm=1.30.2-1.1
sudo apt-mark hold kubeadm

# Plan, then apply (this upgrades control-plane static pods AND renews certs)
sudo kubeadm upgrade plan
sudo kubeadm upgrade apply v1.30.2
```

### 2. Other control-plane nodes
Same `kubeadm` package upgrade, but apply with:
```bash
sudo kubeadm upgrade node
```

### 3. Upgrade kubelet & kubectl on each control-plane node
```bash
kubectl drain <cp-node> --ignore-daemonsets   # from a machine with cluster access
sudo apt-mark unhold kubelet kubectl
sudo apt-get install -y kubelet=1.30.2-1.1 kubectl=1.30.2-1.1
sudo apt-mark hold kubelet kubectl
sudo systemctl daemon-reload && sudo systemctl restart kubelet
kubectl uncordon <cp-node>
```

### 4. Worker nodes (one at a time)
```bash
# On a control-plane / admin machine:
kubectl drain <worker> --ignore-daemonsets --delete-emptydir-data

# On the worker:
sudo apt-mark unhold kubeadm && sudo apt-get install -y kubeadm=1.30.2-1.1 && sudo apt-mark hold kubeadm
sudo kubeadm upgrade node
sudo apt-mark unhold kubelet kubectl && sudo apt-get install -y kubelet=1.30.2-1.1 kubectl=1.30.2-1.1 && sudo apt-mark hold kubelet kubectl
sudo systemctl daemon-reload && sudo systemctl restart kubelet

# Back on the admin machine:
kubectl uncordon <worker>
```

### 5. Verify
```bash
kubectl get nodes        # all Ready, all showing the new VERSION
kubectl version
kubectl get pods -A      # nothing stuck/crashlooping
```

## Managed-cluster upgrades

The provider upgrades the control plane; you trigger it and then upgrade node pools. Same
control-plane-first, one-minor-at-a-time discipline applies, and node pools must stay within skew of
the control plane.

```bash
# EKS
eksctl upgrade cluster --name prod --region us-west-2 --approve         # control plane
eksctl upgrade nodegroup --cluster prod --name ng-prod --kubernetes-version 1.30

# AKS
az aks get-upgrades --resource-group rg-prod --name aks-prod
az aks upgrade --resource-group rg-prod --name aks-prod --kubernetes-version 1.30.2  # control plane + nodes
az aks nodepool upgrade --resource-group rg-prod --cluster-name aks-prod --name np1 --kubernetes-version 1.30.2

# GKE
gcloud container clusters upgrade prod --master --cluster-version 1.30 --zone us-central1-a  # control plane
gcloud container clusters upgrade prod --node-pool default-pool --zone us-central1-a          # nodes
```
GKE Autopilot and "auto-upgrade"-enabled node pools handle this for you on a release channel.

## Pre-upgrade checklist

- [ ] etcd snapshot taken and verified.
- [ ] Target is exactly one minor above current (or a patch bump).
- [ ] Checked release notes for removed/deprecated APIs; ran `kubectl` (or `kubent`/`pluto`) to find
      manifests using soon-removed `apiVersion`s and migrated them.
- [ ] PodDisruptionBudgets exist so `drain` keeps services up (and aren't so strict they deadlock the
      drain).
- [ ] Validated in staging.
- [ ] Maintenance window / on-call aware.
