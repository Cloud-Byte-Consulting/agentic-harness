# Multi-Tenancy, Quotas & vClusters

Sharing a cluster safely: namespaces as the boundary, ResourceQuota/LimitRange for fair-share, soft vs hard
isolation, and vClusters for tenants who need cluster-admin without owning a cluster.

## Table of contents
- Namespaces are the security boundary
- ResourceQuota
- LimitRange
- requests vs limits (and what enforces them)
- Soft vs hard isolation
- vClusters: architecture
- vCluster identity & secrets
- vCluster HA & upgrades
- Self-service tenancy

## Namespaces are the security boundary

Namespaces give resource isolation, name scoping, and the anchor for RBAC + quotas + policy. Cross-namespace pod
traffic is allowed by default (services resolve at `<svc>.<ns>.svc.cluster.local`) — restrict it with
NetworkPolicy (mechanics in the **kubernetes-networking** skill). Recommended: one namespace per app/team/env. The
default namespaces (`default`, `kube-system`, `kube-public`, `kube-node-lease`) — leave `kube-system` alone;
don't deploy workloads there.

```bash
kubectl create ns team-a
kubectl config set-context --current --namespace=team-a   # avoid forgetting -n
```

## ResourceQuota

Caps aggregate consumption and object counts in a namespace (enforced by the `ResourceQuota` admission
controller). Always set one per tenant namespace so no tenant starves the cluster:
```yaml
apiVersion: v1
kind: ResourceQuota
metadata: { name: team-a-quota, namespace: team-a }
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    pods: "50"
    services: "10"
    configmaps: "30"
    persistentvolumeclaims: "10"
    requests.storage: 100Gi
    gold.storageclass.storage.k8s.io/requests.storage: 50Gi   # per-StorageClass quota
```
With a ResourceQuota on compute, every pod **must** declare requests/limits or it's rejected. A violation is
explicit:
```
Error from server (Forbidden): ... exceeded quota: team-a-quota, requested: requests.memory=3Gi,
used: requests.memory=7Gi, limited: requests.memory=8Gi
```

## LimitRange

Sets per-container/pod **defaults** and **min/max bounds** in a namespace — the safety net for when someone
forgets requests/limits:
```yaml
apiVersion: v1
kind: LimitRange
metadata: { name: team-a-limits, namespace: team-a }
spec:
  limits:
    - type: Container
      default:            # applied as limit if omitted
        cpu: 500m
        memory: 256Mi
      defaultRequest:     # applied as request if omitted
        cpu: 250m
        memory: 128Mi
      max:
        cpu: "1"
        memory: 1Gi
      min:
        cpu: 100m
        memory: 128Mi
```
If `default`/`defaultRequest` are omitted, `max`/`min` act as the defaults. Use ResourceQuota (namespace total) +
LimitRange (per-object bounds/defaults) together.

## requests vs limits (and what enforces them)

- **requests** — the minimum the scheduler reserves; a pod requesting more than any node can offer stays
  **Pending** (`FailedScheduling … Insufficient memory`). Pods can't span nodes.
- **limits** — the ceiling. Exceed **CPU** → throttled (slow). Exceed **memory** → **OOM-killed** (memory isn't
  compressible). Always set both; memory limit too low = crash loop.
```yaml
resources:
  requests: { cpu: 250m, memory: 512Mi }   # 250m = 1/4 core
  limits:   { cpu: "1",  memory: 1Gi }
```

## Soft vs hard isolation

- **Soft (namespace) isolation** — namespace + RBAC + ResourceQuota + NetworkPolicy + Pod Security. Cheap, but
  tenants share one API server, one set of CRDs (only one version cluster-wide), and node kernels — a container
  breakout or a noisy CRD upgrade has a cross-tenant blast radius. Granting a tenant the `admin` ClusterRole lets
  them edit RBAC and potentially escalate.
- **Hard isolation** — separate clusters (max isolation, max cost/sprawl) or **vClusters** (virtual clusters
  inside namespaces — most of the isolation, far less sprawl).

## vClusters: architecture

A vCluster (Loft Labs) runs a full Kubernetes control plane (API server + datastore + a syncer) inside an
*unprivileged namespace* of a host cluster. The tenant gets their own API server, can install their own CRDs/
operators, and even have cluster-admin on the virtual cluster — without touching the host. Pods created in the
vCluster are **synced by the syncer into the host namespace** and actually scheduled there, so they're still
subject to the host's ResourceQuota, Pod Security, and admission policies. You apply policy once, on the host.
vClusters share the host's Ingress controller (Ingress objects sync up).

```bash
vcluster create myvc --distro k3s -n tenant1     # default datastore is k3s (SQLite/relational)
vcluster connect myvc -n tenant1                 # port-forward + temp kubeconfig context
vcluster disconnect
```
The synced pod's env/DNS is rewritten so it believes `kubernetes.default.svc` is the vCluster API server, not the
host's — from the pod's view it's a private cluster.

## vCluster identity & secrets

A pod in a vCluster has two identities: the **vCluster SA token** (signed by the vCluster's unique keys, scoped to
the vCluster API server) and the **host identity** (it actually runs in the host). Gotchas:
- vCluster-issued tokens can have very long expiry (a known issue) — mitigate by validating the token's backing
  pod (Vault's TokenReview rejects tokens from dead pods).
- Enable `--service-account-token-secrets=true` so projected tokens land in host Secrets instead of inline pod
  annotations (avoids leaking them in audit logs).
- For external secrets (Vault), you either onboard each vCluster as its own auth path (clean namespace-based
  policies, but per-vCluster onboarding) or use host tokens (less onboarding, but host SA names embed the
  namespace so generic per-namespace policies are awkward).

## vCluster HA & upgrades

- For production, back the vCluster with a shared datastore (etcd or a relational DB like MySQL) instead of local
  SQLite, run ≥2–3 replicas of the control-plane pods, and set a `PodDisruptionBudget`.
- Upgrading is a Helm upgrade of the vCluster values with a new k3s/distro image:
  `vcluster create myvc --upgrade -f values.yaml -n tenant1`. Keep vCluster and host versions from drifting far.

## Self-service tenancy

Repeatable onboarding (often a "platform engineering" portal, e.g. OpenUnison's namespace-as-a-service, or
Terraform/Pulumi/Crossplane) typically, per tenant: create namespace + RoleBinding to an IdP group → ResourceQuota/
LimitRange → deploy a vCluster → wire central auth (host IdP federates to the vCluster) → register the vCluster
with Vault. Centralize authentication and secret management on the host so every tenant inherits compliance
without re-integration. For policy-driven multi-tenancy frameworks, see also Capsule and HyperShift.
