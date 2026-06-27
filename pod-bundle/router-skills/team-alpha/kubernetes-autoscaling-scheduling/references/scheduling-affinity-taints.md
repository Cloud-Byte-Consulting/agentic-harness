# Pod Scheduling — affinity, taints/tolerations, selectors

kube-scheduler assigns each pending pod (empty `nodeName`) to a node in two phases: **Filtering**
(which nodes *can* run it — resources, selectors, taints, affinity) then **Scoring** (rank the
survivors; bind to the highest). Decisions are made **once, at schedule time** — a running pod is
not moved if rules later stop matching (use the Descheduler for that). With Karpenter, scheduling
is **layered**: cloud universe → NodePool constraints → pod constraints; a pod's constraints must
fall within the NodePool or it stays Pending.

## Contents
- nodeName (avoid)
- nodeSelector
- Node affinity (required/preferred)
- Pod affinity & anti-affinity
- Taints & tolerations
- Requesting specific hardware (GPU, local NVMe) with Karpenter
- PersistentVolume zonal placement
- Static pods & scheduler config

---

## nodeName (avoid in production)

Setting `spec.nodeName` pins a pod to one node and **bypasses the scheduler entirely**. Brittle
(node names change, no capacity check). Only useful for debugging.

```yaml
spec:
  nodeName: worker-2
```

## nodeSelector (simple, hard)

Exact label match — pod schedules only on nodes carrying *all* listed labels. Equality only (no
set logic). It is a **hard** requirement; no matching node ⇒ Pending.

```yaml
spec:
  nodeSelector:
    disktype: ssd
    topology.kubernetes.io/zone: eu-west-1a
```

Use well-known labels (`kubernetes.io/os`, `kubernetes.io/arch`, zone, instance-type) or your own
(`kubectl label node NODE disktype=ssd`). With Karpenter, label nodes via the NodePool
`template.metadata.labels`; well-known labels are set automatically.

## Node affinity (`spec.affinity.nodeAffinity`)

Richer than nodeSelector: operators (`In`, `NotIn`, `Exists`, `DoesNotExist`, `Gt`, `Lt`), and
soft preferences with weights. Two kinds:
- `requiredDuringSchedulingIgnoredDuringExecution` — hard rule (like nodeSelector + operators).
- `preferredDuringSchedulingIgnoredDuringExecution` — soft; scheduler tries, falls back if it
  can't. Weighted (`weight: 1–100`).

There is no separate node *anti*-affinity field — use `NotIn`/`DoesNotExist` to repel.

```yaml
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:    # must NOT be an extremelyslow node
        nodeSelectorTerms:
        - matchExpressions:
          - {key: node-type, operator: NotIn, values: ["extremelyslow"]}
      preferredDuringSchedulingIgnoredDuringExecution:   # prefer spot in these zones
      - weight: 100
        preference:
          matchExpressions:
          - {key: karpenter.sh/capacity-type, operator: In, values: ["spot"]}
```

**Karpenter note:** it treats `preferred` affinity as **required initially** when provisioning, so
heavy use of preferences can cause *more* nodes than expected (it adds capacity to satisfy the
preference rather than bin-packing tighter). Because kube-scheduler may relax a preference before
Karpenter provisions, there are sharp edges — keep preferences minimal if you want tight packing.

## Pod affinity & anti-affinity (`podAffinity` / `podAntiAffinity`)

Place pods relative to *other pods*, within a `topologyKey` domain (hostname, zone, region…).
`requiredDuringScheduling…` (hard) or `preferredDuringScheduling…` (soft).

Anti-affinity — keep replicas off the same node:

```yaml
spec:
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchLabels: {app: web}
        topologyKey: "kubernetes.io/hostname"
```

Affinity — co-locate frontend in a zone that already runs a backend:

```yaml
    podAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
          - {key: system, operator: In, values: ["backend"]}
        topologyKey: "topology.kubernetes.io/zone"
```

Caveats: required pod anti-affinity **scales poorly** (the scheduler must consider all existing
pods and their rules). Karpenter will launch a new node to satisfy a required anti-affinity (e.g.
a 4th `app:web` replica when one-per-node and 3 nodes exist). For "spread my replicas," prefer
**topologySpreadConstraints** (see `topology-spread-and-pdb.md`) over anti-affinity.

## Taints & tolerations

Taints **repel** pods from a node unless the pod **tolerates** the taint. Structure:
`<key>=<value>:<effect>`.

Effects:
- `NoSchedule` — scheduler won't place new pods (≈ hard node anti-affinity).
- `PreferNoSchedule` — tries to avoid (≈ soft).
- `NoExecute` — won't schedule **and evicts** running pods that don't tolerate it; `tolerationSeconds`
  controls how long a tolerating pod stays before eviction.

```bash
kubectl taint node NODE machine-check-exception=memory:NoSchedule
kubectl taint node NODE machine-check-exception=memory:NoSchedule-   # remove (note trailing -)
```

```yaml
spec:
  tolerations:
  - key: "nvidia.com/gpu"
    operator: "Exists"        # Exists = match key only; Equal = match key+value
    effect: "NoSchedule"
```

Key points:
- A toleration **allows** scheduling on a tainted node; it does **not force** the pod there. To
  *force*, pair the taint/toleration with a `nodeSelector`/affinity for a label only those nodes
  have.
- A pod must tolerate **all** of a node's taints to land on it.
- Built-in `NoExecute` taints managed by the system: `node.kubernetes.io/not-ready`,
  `unreachable`, `memory-pressure`, `disk-pressure`, `network-unavailable`, `unschedulable`,
  `node.cloudprovider.kubernetes.io/uninitialized`. New pods auto-get `not-ready`/`unreachable`
  tolerations with `tolerationSeconds: 300`.
- **Always pair `NoExecute` with a `NoSchedule` taint** when using `tolerationSeconds`, or you get
  eviction→reschedule→eviction loops (the `NoExecute` taint is "ignored" during the toleration
  window, so it doesn't prevent re-scheduling).
- **Common mistake:** taint a NodePool's nodes but forget the matching toleration on the workload →
  pods Pending forever and Karpenter won't launch nodes for them.
- **`startupTaints`** (Karpenter NodePool): if a CNI/agent adds a temporary taint at boot, declare
  it as a `startupTaint` so Karpenter knows it's expected and waits, instead of over-provisioning
  more nodes thinking the node is unusable.

## Requesting specific hardware (Karpenter)

**GPU** — dedicate a NodePool to GPU instances and taint it; pods request the GPU resource and
tolerate the taint:

```yaml
# NodePool
spec:
  template:
    spec:
      requirements:
        - {key: node.kubernetes.io/instance-type, operator: In, values: ["p3.8xlarge","p3.16xlarge"]}
      taints:
        - {key: nvidia.com/gpu, value: "true", effect: NoSchedule}
```
```yaml
# Pod
spec:
  tolerations: [{key: nvidia.com/gpu, operator: Exists, effect: NoSchedule}]
  containers:
  - name: trainer
    resources: {limits: {nvidia.com/gpu: 1}}
```

**Local NVMe** — require via the well-known label (with `Gt` for a minimum size); ensure the
NodeClass mounts the instance store:

```yaml
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - {key: "karpenter.k8s.aws/instance-local-nvme", operator: Gt, values: ["99"]}  # ≥100GB
```

Only restrict instance types for true hardware/perf needs — over-restriction hurts
availability/cost. Note: ephemeral-storage *requests* don't reliably map to local-NVMe capacity in
Karpenter; express the need via labels + taints/affinity instead.

## PersistentVolume zonal placement

Storage imposes implicit scheduling constraints. An EBS volume lives in one AZ, so a pod using it
is auto-pinned (via PV node affinity) to that AZ — even without an explicit zone rule. Karpenter
reads this and provisions a node in the right zone. With `WaitForFirstConsumer` StorageClasses,
the volume is created *after* the pod schedules; Karpenter picks an allowed zone and brings up a
node there. **Ensure your NodePools cover every zone your volumes use**, or stateful pods get
stuck Pending.

## Static pods & scheduler config

- **Static pods** are managed directly by a node's kubelet (from a file/URL), not the API server
  (which only sees read-only mirror pods). Used to bootstrap control-plane components. For
  per-node workloads, prefer DaemonSets.
- **Scheduler configuration** (`KubeSchedulerConfiguration`, v1.23+) replaces old scheduling
  policies: define profiles, plugins, and extension points. `percentageOfNodesToScore` trades
  scheduling speed vs placement quality in large clusters (default scales with size: ~50% at 100
  nodes, ~10% at 5000). `NodeRestriction` admission + Node authorizer prevent a kubelet from
  mislabeling its node (restricted-prefix labels for isolation). On managed clusters
  (EKS/GKE/AKS) you can't edit the scheduler — control placement via the pod fields above.
