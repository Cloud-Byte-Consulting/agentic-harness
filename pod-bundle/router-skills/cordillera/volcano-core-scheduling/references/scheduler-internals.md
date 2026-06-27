# Volcano Scheduler Internals & Configuration

This reference covers the innermost workings of the Volcano scheduler pipeline: the 6 fundamental Actions, the 20+ plugins that implement decision policies, and the exact configuration syntax of `volcano-scheduler.conf`.

---

## 1. The Scheduling Cycle & Actions

The Volcano scheduler does not schedule pods individually. Instead, it runs a periodic **scheduling cycle** (default every 1 second). Each cycle starts by opening a `Session`, fetching a snapshot of the cluster state, and running a series of sequential pipeline steps called **Actions**.

```
Session Open (Cache snapshot of Cluster)
  │
  ▼
┌────────────────────────┐
│ 1. ENQUEUE Action      │ ──► Moves PodGroups from Pending to Inqueue (validates queues)
└────────────────────────┘
  │
  ▼
┌────────────────────────┐
│ 2. ALLOCATE Action     │ ──► Binds pods to nodes using plugin filters & scoring
└────────────────────────┘
  │
  ▼
┌────────────────────────┐
│ 3. PREEMPT Action      │ ──► Evicts low-priority pods to fit high-priority ones in same queue
└────────────────────────┘
  │
  ▼
┌────────────────────────┐
│ 4. RECLAIM Action      │ ──► Evicts pods from over-capacity queues to return resources to deserved queues
└────────────────────────┘
  │
  ▼
┌────────────────────────┐
│ 5. BACKFILL Action     │ ──► Allocates remaining idle resources to small/short jobs (no preemption)
└────────────────────────┘
  │
  ▼
Session Close (Batch commit bindings and states to K8s API Server)
```

### 1. Enqueue
Validates if a `PodGroup` is allowed to enter the active scheduling pool. It checks if the target queue has space and calls plugins' `OnJobEnqueue` methods. If passed, the PodGroup transitions from `Pending` to `Inqueue`.

### 2. Allocate
The primary scheduling step. It iterates through all `Inqueue` jobs and their pods, runs filtering and scoring plugins, and binds pods to the best-fit nodes. It updates the local cycle cache so subsequent pods in the same cycle recognize the depleted capacity.

### 3. Preempt
Operates *within* a single queue. If a high-priority job is Pending due to lack of space, the scheduler will locate lower-priority running pods *in that same queue*, evict them, and allocate those resources to the high-priority job.

### 4. Reclaim
Operates *across* different queues. If Queue A is running above its `deserved` capacity using idle cluster space, and Queue B (which is currently under-capacity) receives a new job, the `reclaim` action will evict running pods from Queue A to return those resources to Queue B.

### 5. Backfill
A low-overhead pass at the end of the cycle. It schedules small, short-running pods into any remaining tiny resource "gaps" in the cluster. It will never trigger preemption or eviction; it only fills empty slots.

---

## 2. Core Scheduler Plugins

Plugins implement the actual policies invoked during Actions. Volcano ships with 20+ built-in plugins:

| Plugin Name | Primary Purpose | Key Methods / Functions |
|-------------|-----------------|-------------------------|
| **gang** | Enforces atomic co-scheduling. | Rejects jobs in `Allocate` if the number of allocatable pods is less than `minAvailable`. |
| **drf** | Implements Dominant Resource Fairness. | Orders jobs/queues based on resource usage shares, prioritizing the most resource-starved tenants. |
| **proportion** | Distributes queue capacity proportionally. | Allocates resources to queues based on their configured `weight`. |
| **binpack** | Maximizes node resource packing density. | Scores nodes higher if they have existing workloads, packing pods tightly to allow other nodes to idle/power down. |
| **priority** | Sorts jobs and tasks. | Orders scheduling priority based on standard `PriorityClass` values. |
| **predicates** | Standard Kubernetes filtering. | Runs kube-scheduler filters like `NodeAffinity`, `Taints/Tolerations`, and `VolumeLimits`. |
| **nodeorder** | Scores nodes. | Evaluates and scores nodes based on resource fit, affinity, and topology. |
| **capacity** | Enforces Queue capabilities. | Limits queue allocations to their hard `capability` specs, and triggers `Reclaim` relative to `deserved` specs. |
| **sla** | Enforces Job SLAs. | Prioritizes jobs based on waiting times or latency metrics to meet Service Level Agreements. |
| **numa-aware** | Allocates NUMA sockets. | Coordinates with `volcano-agent` to align high-performance pods with matching CPU socket topologies. |
| **deviceshare** | Shares GPUs/VPU/NPU. | Enables fractional device scheduling (e.g., allocating `0.5` of a GPU to a container). |
| **conformance** | Protects system/critical pods. | Prevents critical namespace pods (e.g., `kube-system` or daemonsets) from being preempted. |

---

## 3. Configuration Reference (`volcano-scheduler.conf`)

The scheduler configuration is loaded from a ConfigMap (typically `volcano-scheduler-conf` in `volcano-system`). It is structured into **actions**, **tiers**, and **configurations**.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: volcano-scheduler-conf
  namespace: volcano-system
data:
  volcano-scheduler.conf: |
    actions: "enqueue, allocate, preempt, reclaim, backfill"
    tiers:
      - plugins:
        - name: priority
        - name: gang
          arguments:
            # Rejection behavior: if gang fails, how long to backoff before re-evaluation
            gang-backoff-period: 15s
        - name: conformance
      - plugins:
        - name: drf
        - name: predicates
        - name: proportion
        - name: nodeorder
        - name: binpack
          arguments:
            # Weights for resource types in bin-packing score calculation
            binpack.resources: "cpu: 10, memory: 5, nvidia.com/gpu: 20"
    configurations:
      - name: capacity
        arguments:
          # Enable cross-queue preemption (reclaim) under capacity guidelines
          reclaimable: true
      - name: proportion
        arguments:
          # Automatically open queues if they meet conditions
          vqueue-enabled: true
```

### Key Parameters:
- `actions`: Space/comma-separated list of actions to execute in order. **Order matters**. If you place `preempt` before `allocate`, preemption logic will evaluate before regular allocation, which is highly inefficient.
- `tiers`: Layers of plugins. The scheduler runs the first tier of plugins to filter and sort. If multiple nodes/jobs are equal, it moves to the next tier (e.g., sorting first by `priority` and `gang` compliance, and then scoring via `drf` and `binpack`).
- `arguments`: Plugin-specific parameters. For example, `binpack.resources` lets you tune which resources the binpacker should prioritize when packing nodes.
