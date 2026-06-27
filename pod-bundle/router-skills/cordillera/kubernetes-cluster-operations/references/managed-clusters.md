# Managed Kubernetes Clusters

Selection guidance and operational commands for the major managed offerings. With all of these the
provider runs, patches, and backs up the **control plane and etcd**; you manage worker node pools,
workloads, networking choices, and observability — unless you choose a serverless tier that hides
the nodes too.

## Contents
- [Selection matrix](#selection-matrix)
- [What the provider manages vs you](#what-the-provider-manages-vs-you)
- [Amazon EKS (eksctl)](#amazon-eks)
- [Azure AKS (az aks)](#azure-aks)
- [Google GKE (gcloud) + Autopilot](#google-gke)
- [Linode LKE](#linode-lke)
- [DigitalOcean Kubernetes](#digitalocean-kubernetes)
- [OpenShift / ROSA](#openshift--rosa)
- [Serverless tiers](#serverless-tiers)

## Selection matrix

| Offering | Pick when | Notes |
|---|---|---|
| **GKE** | Deepest, most polished managed K8s; Autopilot for hands-off | Originated at Google; great console integration |
| **EKS** | You're standardized on AWS | Less console-integrated; `eksctl` strongly recommended; control-plane fee (~$0.10/hr) |
| **AKS** | You're on Azure | Free control plane (standard tier); tight Azure Monitor integration |
| **LKE** (Linode/Akamai) | Predictable low cost, simplicity | **Free control plane**; fewer surrounding services (no native IAM/RBAC tie-in, registry, etc.) |
| **DigitalOcean** | Simple, cheap, dev-friendly | Sometimes trails on newest K8s versions — check before relying on a specific API version |
| **OpenShift / ROSA** | Enterprise PaaS, vendor support, compliance | Expensive (licensing + infra); adds dev tooling, etcd encryption, integrated CI/CD; still plain K8s underneath |

Decision drivers: existing cloud, level of control wanted, team expertise, scaling needs, and cost
(control-plane fee vs free, node hours, support contracts). Multi-cloud is real but advanced — it
needs cross-cloud networking, auth, and security work; research thoroughly before committing.

## What the provider manages vs you

| Layer | Managed (EKS/AKS/GKE/LKE/DO) | You |
|---|---|---|
| API server, scheduler, controller-manager, etcd | ✅ provider | — |
| Control-plane patching, HA, backups | ✅ provider | — |
| Worker node pools (size, count, OS, autoscaling) | partial | ✅ you (except serverless tiers) |
| Workloads, namespaces, RBAC | — | ✅ you |
| Monitoring/logging/alerting | optional add-on | ✅ you choose & wire |
| CNI / network policy choice | sometimes default | ✅ you |

Node *autoscaling* (Cluster Autoscaler / Karpenter) belongs to kubernetes-autoscaling-scheduling —
cross-link rather than duplicate here.

## Amazon EKS

Use the official **eksctl** CLI; raw AWS CLI/console for EKS requires ~8 manual steps (VPC, subnets,
IAM roles, node groups). Configure the AWS CLI with an IAM user that has EKS permissions first.

### Install eksctl
```bash
brew install eksctl                 # macOS
choco install eksctl                # Windows
# Linux: download the tarball from github.com/eksctl-io/eksctl/releases, extract, move to /usr/local/bin
eksctl version
```

### Create / inspect / delete
```bash
# Simplest sandbox (defaults: ~2 nodes, managed node group, official EKS AMI)
eksctl create cluster

# Explicit
eksctl create cluster --name prod --region us-west-2 \
  --nodes 3 --node-type t3.large --managed --nodegroup-name ng-prod

eksctl get cluster --name prod --region us-west-2
eksctl delete cluster --name prod --region us-west-2
```
eksctl provisions everything via **CloudFormation** (VPC, subnets, security groups, IAM, node group)
and configures kubectl automatically. To point another machine's kubectl at the cluster:
```bash
aws eks update-kubeconfig --name prod --region us-west-2
```

### IAM specifics
- The EKS **control-plane role** needs `AmazonEKSClusterPolicy` (+ `AmazonEC2ContainerRegistryReadOnly`).
- **Worker node role** needs `AmazonEKSWorkerNodePolicy`, `AmazonEKS_CNI_Policy`, and
  `AmazonEC2ContainerRegistryReadOnly`.
- API endpoint access can be **public**, **public+private** (control plane public, node traffic
  internal), or **private** (VPC-only).
- eksctl grants cluster access to the IAM identity that created it; viewing workloads in the console
  may require logging in as that user (or configuring EKS access entries / aws-auth).
- Enable control-plane logging (`api,audit,authenticator,controllerManager,scheduler`) for auditability.

## Azure AKS

Use the **az** CLI (or Azure Cloud Shell with mounted storage so kubeconfig persists).

```bash
az account show                                   # confirm subscription
az group create --name rg-prod --location eastus  # AKS inherits the group's location

az aks create --resource-group rg-prod --name aks-prod \
  --node-count 3 --enable-addons monitoring --generate-ssh-keys \
  --node-vm-size Standard_DS3_v2

# Point kubectl at it
az aks get-credentials --resource-group rg-prod --name aks-prod
kubectl get nodes

# Scale a node pool
az aks scale --resource-group rg-prod --name aks-prod --node-count 5
# (Cluster Autoscaler: --enable-cluster-autoscaler --min-count --max-count — see autoscaling skill)

# Teardown
az aks delete --resource-group rg-prod --name aks-prod
az group delete --name rg-prod                    # removes everything in the group
```
Node pools are just Azure VM scale sets running as workers. AKS abstracts the control plane/API
server; you own worker scaling, monitoring (Azure Monitor / Container Insights, or Prometheus+Grafana
for multi-cloud), and apps.

## Google GKE

Use **gcloud** (`gcloud init` to authenticate; install the auth plugin so kubectl can authenticate).

```bash
gcloud config set project PROJECT_ID

# Zonal cluster (nodes in one zone)
gcloud container clusters create prod --num-nodes=2 --zone=us-central1-a

# Regional cluster (nodes replicated across zones in the region — more HA)
gcloud container clusters create prod --num-nodes=2 --region=us-central1 \
  --node-locations us-central1-a,us-central1-b,us-central1-c

# Configure kubectl
gcloud container clusters get-credentials prod --zone=us-central1-a
kubectl get nodes

# Resize / update node locations
gcloud container clusters update prod --min-nodes=2 --region=us-central1 \
  --node-locations us-central1-a,us-central1-b,us-central1-c

# Delete (match the zone/region flag used at create time)
gcloud container clusters delete prod --zone=us-central1-a
```
GKE is the most integrated of the three with its cloud console. Regional clusters spread workers and
control-plane replicas across zones for higher availability.

### GKE Autopilot
Fully managed *nodes too* — you only deploy workloads; Google sizes, scales, secures, and bills per
Pod resource. Create with `gcloud container clusters create-auto NAME --region=us-central1`. Choose
Autopilot when you don't want to manage node pools at all.

## Linode LKE

Developer-friendly, transparent bundled pricing, **free control plane**. Create via the cloud
manager UI or Terraform (`linode_lke_cluster` resource; pools are a list of `{type, count}`; output
the `kubeconfig`). Enable the **HA Control Plane** option for production. Trade-off: fewer ecosystem
services than the big three (no built-in IAM/RBAC integration, registry, etc.), which can be a
dealbreaker for some security teams.

## DigitalOcean Kubernetes

Simple and inexpensive; UI, `doctl`, or Terraform (`digitalocean_kubernetes_cluster`; node pools are
DO Droplets, support `auto_scale` with `min_nodes`/`max_nodes`). Always enable the **HA Control
Plane** for production. Watch the available Kubernetes versions — DO has historically lagged the
newest releases, so confirm the API version you need is offered.

## OpenShift / ROSA

Red Hat's enterprise Kubernetes PaaS. It **is** Kubernetes underneath — your Pod/Deployment/Service
manifests work unchanged — but adds opinionated management, security (RBAC + SCC), developer tooling,
image builds, and CI/CD. It's enterprise-oriented (licensing + infra cost; you pay for support).

- **Local/dev:** OpenShift Local (CodeReady Containers / CRC): `crc setup` then `crc start`. Or the
  hosted Developer Sandbox. (CRC isn't supported on Apple Silicon/ARM.)
- **Production on AWS (ROSA):** from the Red Hat console, create a ROSA cluster; associate your AWS
  account, set cluster name/version/region, machine pool (EC2 size + autoscaling), networking
  (public/private API, new vs existing VPC, Pod/Service CIDRs), roles/policies (auto or manual), and
  update strategy. ROSA can encrypt etcd and PVs with customer keys.

## Serverless tiers

No worker nodes to manage:
- **GKE Autopilot** — Google manages and bills per-Pod.
- **EKS Fargate** — define Fargate profiles; pods run on AWS-managed capacity (Virtual Kubelet under
  the hood, automatic).
- **AKS Virtual Kubelet → ACI** — "ACI bursting": send Pods to Azure Container Instances instead of
  scaling node pools (Azure Container Apps is the newer serverless option).
