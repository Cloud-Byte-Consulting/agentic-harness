# Node Autoscaling — Cluster Autoscaler vs Karpenter

Node (data-plane) autoscalers add capacity when pods are **Pending** (unschedulable) and remove
underutilized nodes. They react **only to unschedulable pods** — never to CPU/memory utilization —
and they **provision capacity; kube-scheduler still schedules the pods**. Pods must declare
`requests` or the math is unreliable.

## Contents
- How to choose
- Cluster Autoscaler (CA) on AWS + best practices
- Karpenter: resources (NodePool / EC2NodeClass / NodeClaim)
- NodePool requirements & well-known labels
- How Karpenter launches nodes (batching, bin-packing)
- Other autoscalers: Descheduler, CPA, CPVA

---

## How to choose

| | Cluster Autoscaler | Karpenter |
|---|---|---|
| Model | Infrastructure-first: scales predefined node groups (ASGs) | Application-first: reads pending pods, calls cloud API directly |
| Instance types | Needs **homogeneous** types per group (uses first type to size) | Diverse types/sizes from broad constraints; bin-packs |
| Node groups | Many groups → slower decisions | No groups; few NodePools |
| Consolidation | Removes empty/underutilized nodes (10-min default delay) | Continuous consolidation: empty, underutilized, downsize, spot-to-spot |
| Portability | Many clouds + on-prem (varies) | AWS (mature), Azure (prod), GCP (active); no on-prem yet |

Prefer **Karpenter** for flexibility, bin-packing, and cost (it picks right-sized, diverse,
cheapest-fit instances and consolidates aggressively). **CA** is fine for simple, stable fleets or
where Karpenter has no provider. The skill defaults examples to AWS/EKS.

---

## Cluster Autoscaler on AWS

Runs as a Deployment; integrates with EC2 Auto Scaling groups (adjusts *desired* capacity, leaves
min/max). Prefer **Auto-Discovery** mode (tag ASGs) and **managed node groups**. Needs IAM perms
to describe/modify ASGs.

```bash
kubectl get pods -n kube-system -l app.kubernetes.io/instance=cluster-autoscaler
kubectl logs  -n kube-system -l app.kubernetes.io/instance=cluster-autoscaler --all-containers=true -f --tail=20
```

### CA best practices (also explain *why* Karpenter differs)
- **Homogeneous instance types within a node group.** CA sizes scale-out using the **first** type
  in the list; mixed specs cause waste (bigger nodes) or starvation (smaller nodes). Add many
  *same-spec* types for capacity/Spot diversification.
- **Limit the number of node groups.** CA loads all groups into memory every scan (~10s) to
  simulate scheduling; more groups = slower scale-out. Isolate workloads by namespace, not groups.
- **Sane scan speed.** Don't set an aggressive `--scan-interval`; instance launch takes minutes and
  frequent API calls hit cloud rate limits. ~1 minute is a good balance. Tune scale-down with
  `--scale-down-delay-after-add`, `--scale-down-unneeded-time`, etc. (default scale-down after
  empty: 10 min).
- **Priority expander** for ordered group preference (`--expander=priority`) via a ConfigMap (e.g.
  prefer spot-large → spot-xlarge → spot-2xlarge).
- **Single-AZ node groups for EBS-backed workloads** (EBS is AZ-bound) to avoid pods scheduling
  away from their volume — but weigh the AZ-resilience trade-off for replicated stateful systems.

---

## Karpenter resources

Three CRDs. **NodeClass** = cloud-specific infra (on AWS: `EC2NodeClass`). **NodePool** =
constraints/boundaries for what Karpenter may launch. **NodeClaim** = immutable, read-only record
of a launched node (inspect for troubleshooting).

### EC2NodeClass (`karpenter.k8s.aws/v1`)

```yaml
apiVersion: karpenter.k8s.aws/v1
kind: EC2NodeClass
metadata:
  name: default
spec:
  role: "KarpenterNodeRole-mycluster"     # IAM role for launched instances; immutable
  amiSelectorTerms:
    - alias: "bottlerocket@1.38.0"         # pin a version — NOT @latest in prod (see drift)
  subnetSelectorTerms:
    - tags: {karpenter.sh/discovery: "mycluster"}
  securityGroupSelectorTerms:
    - tags: {karpenter.sh/discovery: "mycluster"}
```

Tag-based selectors (vs static IDs) make the NodeClass portable across regions/clusters and let
you add subnets without editing it. EC2NodeClasses are multi-arch by default.

### NodePool (`karpenter.sh/v1`)

```yaml
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: default
spec:
  template:
    metadata:
      labels: {intent: apps}              # custom labels propagate to nodes (for nodeSelector)
    spec:
      requirements:
        - {key: kubernetes.io/arch, operator: In, values: ["amd64"]}
        - {key: kubernetes.io/os,   operator: In, values: ["linux"]}
        - {key: karpenter.sh/capacity-type, operator: In, values: ["on-demand","spot"]}
        - {key: karpenter.k8s.aws/instance-category,   operator: In, values: ["c","m","r"]}
        - {key: karpenter.k8s.aws/instance-generation, operator: Gt, values: ["4"]}
      nodeClassRef: {group: karpenter.k8s.aws, kind: EC2NodeClass, name: default}
      expireAfter: 720h                    # max node lifetime; "Never" to disable (stateful)
      terminationGracePeriod: 24h          # hard cap on drain before forced termination
  limits:
    cpu: "1000"                            # cap total resources this NodePool may manage
  disruption:
    consolidationPolicy: WhenEmptyOrUnderutilized   # or WhenEmpty
    consolidateAfter: 30s
    budgets:
      - nodes: "10%"                       # see topology-spread-and-pdb.md for full budget syntax
```

Field notes:
- `requirements` — the heart of the NodePool: which instances Karpenter may launch. A pod's
  constraints must intersect these or it stays Pending.
- `limits` — total CPU/memory/GPU ceiling (not a node count, since sizes vary).
- `disruption` — consolidation policy + budgets (see `topology-spread-and-pdb.md`).
- `expireAfter` / `terminationGracePeriod` — node freshness & graceful-drain cap. Changing
  `expireAfter` marks existing nodes drifted (gradual replacement), not instant.
- `weight` (0–100) — when a pod matches multiple NodePools, the highest weight wins (overrides
  cost). Use for "prefer spot, fall back to on-demand" patterns. Avoid equal-weight overlap.

`capacity-type` priority order in Karpenter's simulation: **reserved → spot → on-demand.** On AWS,
on-demand/reserved use the `lowest-price` allocation strategy; spot uses
`price-capacity-optimized`.

### NodeClaim

Immutable; created by Karpenter after a scheduling decision, tracks the node lifecycle. Inspect to
see which instance types were considered and why a node did/didn't launch:

```bash
kubectl get nodeclaim
kubectl describe nodeclaim NAME    # Requirements (instance-type list), Events (launch/registration)
```

---

## NodePool requirements & well-known labels

Operators: `In`, `NotIn`, `Exists`, `DoesNotExist`, `Gt`, `Lt`. Values are always arrays (even a
single value). Less restrictive = more options for Karpenter = better cost/availability.

Well-known Kubernetes labels: `kubernetes.io/arch` (`amd64`/`arm64`), `kubernetes.io/os`,
`topology.kubernetes.io/zone`, `node.kubernetes.io/instance-type`.

Karpenter AWS labels: `karpenter.sh/capacity-type`, `karpenter.k8s.aws/instance-category`,
`-instance-generation`, `-instance-family`, `-instance-size`, `-instance-cpu`, `-instance-memory`,
`-instance-local-nvme`, `-instance-gpu-*`.

`minValues` (alpha) — require a *minimum diversity* of options for a key (resilience for Spot):

```yaml
      requirements:
        - {key: karpenter.k8s.aws/instance-family, operator: Exists, minValues: 5}
        - {key: node.kubernetes.io/instance-type,  operator: Exists, minValues: 15}
```

Under the default `Strict` `min-values-policy`, a NodePool that can't meet `minValues` is skipped;
`BestEffort` relaxes it to keep scheduling.

**Recommendation:** keep NodePool requirements broad; push hardware specifics to the *pod*
(nodeSelector/affinity/tolerations). Over-restricting the pool causes Pending pods and prevents
consolidation. See `scheduling-affinity-taints.md` for GPU/local-NVMe/zonal patterns.

---

## How Karpenter launches nodes

1. **Scheduling** — kube-scheduler marks a pod unschedulable → Karpenter's trigger.
2. **Batching** — expanding window (1s idle, 10s max) groups co-arriving pods so one node can host
   many (env vars `BATCH_IDLE_DURATION`, `BATCH_MAX_DURATION`; rarely change).
3. **Bin-packing** — first-fit-decreasing: sort candidate instance types by cost, sort pods by
   CPU then memory descending, pack onto virtual nodes. Adding a pod with a new constraint (e.g.
   ARM) shrinks the compatible instance list ("resizable bins"). **Karpenter also reserves room for
   DaemonSets** that will land on the new node — so requested totals exceed your app's requests
   alone.
4. **Launch** — pass the diversified instance list to the cloud API (EC2 Fleet on AWS); the
   provider's allocation strategy picks the actual type. Karpenter cleans up failed launches and
   waits for registration/initialization.

Disruption (consolidation, drift, expiration, spot interruption) and disruption budgets are in
`topology-spread-and-pdb.md`.

---

## Other relevant autoscalers

- **Descheduler** — re-balances *already-running* pods (kube-scheduler never re-evaluates). Evicts
  pods from over-utilized nodes so they reschedule onto under-utilized ones. Strategies:
  `LowNodeUtilization`, `RemoveDuplicates`, etc. Use PDBs and start conservatively.
- **Cluster Proportional Autoscaler (CPA)** — scales a workload's replicas proportional to cluster
  size (node/core count), not metrics. Good for `CoreDNS`/system services. Beta.
- **Cluster Proportional Vertical Autoscaler (CPVA)** — vertically scales system components
  proportional to cluster size. Beta.

These complement, not replace, CA/Karpenter.
