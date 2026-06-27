# Cluster Colocation & QoS with Volcano

Running latency-critical online services (like HTTP microservices or databases) and resource-intensive batch jobs (like ML training or data rendering) on separate clusters results in under-utilized hardware. Volcano's colocation framework enables online and offline workloads to share the same physical nodes safely through active oversubscription and resource reclamation.

---

## 1. How Colocation Works

Volcano employs a split-level colocation strategy:

```
┌──────────────────────────────────────────────────────────────┐
│                  KUBERNETES CONTROL PLANE                    │
│   • Core Microservice Pods (High Priority / Online QoS)     │
│   • Volcano Batch Jobs     (Low Priority / Offline QoS)      │
└──────────────────────────────┬───────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────┐
│                       PHYSICAL NODE                          │
│                                                              │
│  ┌────────────────────────┐      ┌────────────────────────┐  │
│  │   Online Pod (HTTP)    │      │   Offline Batch Pod    │  │
│  │   • Guaranteed CPU     │      │   • Oversubscribed CPU │  │
│  └───────────▲────────────┘      └───────────▲────────────┘  │
│              │ (Protects QoS)                │ (Throttles)   │
│              └───────────────┬───────────────┘               │
│                              │                               │
│                 ┌────────────┴────────────┐                  │
│                 │      volcano-agent      │                  │
│                 │ (pkg/agent/ colocation) │                  │
│                 └─────────────────────────┘                  │
└──────────────────────────────────────────────────────────────┘
```

1. **Active Throttling**: The `volcano-agent` daemonset runs on every node, continuously tracking kernel CPU/Memory pressure.
2. **Resource Reclamation**: High-priority online pods are scheduled normally. Low-priority batch pods request oversubscribed, reclaimed resources.
3. **Guaranteed SLAs**: If an online microservice experiences a traffic spike and requires more CPU/Memory, `volcano-agent` immediately throttles, freezes, or evicts the offline batch pods, protecting the online service's SLA.

---

## 2. The `ColocationConfiguration` CRD

The colocation policies are defined namespace-wide via the `ColocationConfiguration` custom resource.

```yaml
apiVersion: config.volcano.sh/v1alpha1
kind: ColocationConfiguration
metadata:
  name: production-colocation-policy
  namespace: default
spec:
  selector:
    matchLabels:
      workload-class: collocated
  configuration:
    cpuQos:
      cpuReclaimedRatio: 60   # Allocates up to 60% of idle CPU capacity to offline jobs
      evictionCpuThreshold: 85 # Evict offline pods if total node CPU exceeds 85%
      freezeCpuThreshold: 75   # Freeze (cgroup freeze) offline pods if CPU exceeds 75%
    memoryQos:
      evictionMemThreshold: 90 # Evict offline pods if total node Memory exceeds 90%
      highRatio: 80            # Throttling memory ratio
      lowRatio: 50
      minRatio: 30
```

- `evictionCpuThreshold: 85`: If host CPU usage exceeds 85%, the `volcano-agent` begins killing (evicting) offline batch pods starting with the lowest priority to free cores immediately.
- `freezeCpuThreshold: 75`: Under moderate node pressure (75%-84%), offline batch processes are suspended in the cgroup tree without being killed, resuming automatically once the pressure drops.

---

## 3. Deploying a Collocated Workload

To deploy a low-priority batch job that utilizes oversubscribed resources, label the job tasks to trigger colocation:

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: collocated-render-job
spec:
  minAvailable: 1
  schedulerName: volcano
  tasks:
    - name: renderer
      replicas: 4
      template:
        metadata:
          labels:
            # Matches the ColocationConfiguration selector
            workload-class: collocated
            volcano.sh/qos-level: offline # Declares offline batch priority
        spec:
          containers:
            - name: blender
              image: blender:latest
              resources:
                requests:
                  # Request oversubscribed, virtual resources
                  volcano.sh/reclaimed-cpu: "4"
                  volcano.sh/reclaimed-memory: "8Gi"
                limits:
                  volcano.sh/reclaimed-cpu: "4"
                  volcano.sh/reclaimed-memory: "8Gi"
```

- `volcano.sh/qos-level: offline`: Informs the scheduler and node agents that this pod runs on reclaimed capacity and can be frozen or evicted at any moment.
- `volcano.sh/reclaimed-cpu`: Reclaimed virtual resources managed by the colocation agent.
