# Network-Topology-Aware Scheduling with Volcano

Large-scale distributed deep learning and MPI workloads are highly network-bound. To achieve maximum throughput, pods must be placed as close together as possible on the network topology (e.g., same rack, same InfiniBand leaf switch) to minimize hop count and maximize inter-node RoCE/InfiniBand bandwidth. This reference covers Volcano's network-topology-aware scheduling.

---

## 1. The `HyperNode` CRD (`topology.volcano.sh/v1alpha1`)

Volcano models the network topology of a cluster as a hierarchical tree of cluster-scoped `HyperNode` resources.

```
                    Region (Tier 3)
                   /               \
            Zone-A (Tier 2)       Zone-B (Tier 2)
              /         \            /         \
         Rack-1          Rack-2   Rack-3       Rack-4
        (Tier 1)        (Tier 1) (Tier 1)     (Tier 1)
```

### Example Manifest: `racks.yaml`
```yaml
apiVersion: topology.volcano.sh/v1alpha1
kind: HyperNode
metadata:
  name: rack-1
spec:
  tier: 1                     # Lowest level: Rack boundary
  tierName: intra-rack
  members:
    - type: Node
      selector:
        exactMatch: worker-node-01
    - type: Node
      selector:
        exactMatch: worker-node-02
---
apiVersion: topology.volcano.sh/v1alpha1
kind: HyperNode
metadata:
  name: zone-a
spec:
  tier: 2                     # Middle level: Switch boundary
  tierName: leaf-switch
  members:
    - type: HyperNode         # References child HyperNodes
      selector:
        exactMatch: rack-1
    - type: HyperNode
      selector:
        exactMatch: rack-2
```

---

## 2. Scheduler Configuration Prerequisites

Add the `network-topology-aware` (or `topology` in older versions) plugin to your actions tier in `volcano-scheduler.conf`:

```yaml
actions: "enqueue, allocate, backfill"
tiers:
  - plugins:
    - name: priority
  - plugins:
    - name: predicates
    - name: nodeorder
    - name: numaaware
    - name: network-topology-aware # Enabled for network-aware scoring
```

---

## 3. Deploying a Network-Aware Job

Specify your network topology requirements under the Job's `spec.networkTopology` block:

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: distributed-llm-pretrain
spec:
  minAvailable: 8             # Requires 8 worker pods
  schedulerName: volcano
  networkTopology:
    mode: hard                # hard: fail if constraints cannot be met; soft: score best-fit
    highestTierAllowed: 1     # Limits placement to Rack boundary (Tier 1)
    highestTierName: "intra-rack"
  tasks:
    - name: gpu-worker
      replicas: 8
      template:
        spec:
          containers:
            - name: trainer
              image: pytorch-distributed:latest
              resources:
                requests:
                  cpu: "8"
                  memory: "32Gi"
                  nvidia.com/gpu: "4"
```

- `mode: hard`: If the scheduler cannot find a single `intra-rack` (Tier 1) `HyperNode` that has enough resources to run all 8 worker pods, the job is **rejected** and stays `Pending`. This is crucial for performance: running distributed training across a slow network tier ruins scaling efficiency.
- `highestTierAllowed: 1`: Limits the maximum network distance between any two worker pods to the lowest tier (Tier 1: Rack).

---

## 4. RoCE & InfiniBand UFM Integration

When running over RDMA over Converged Ethernet (RoCE) or InfiniBand, Volcano integrates with Unified Fabric Manager (UFM) to discover link health:
- If a leaf switch or network port undergoes packet loss, UFM labels the corresponding nodes.
- Volcano's topology-aware scheduler automatically routes pods **away** from degraded network domains.
