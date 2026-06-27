# Volcano Operations, CLI, & Troubleshooting

This reference is an operational guide for cluster administrators. It details installation via Helm, full usage of the `vcctl` CLI, a comprehensive troubleshooting checklist for common failures, and performance tuning configurations.

---

## 1. Installation & Upgrades

Volcano is best installed via Helm.

### Install via Helm:
```bash
# Add the helm repository
helm repo add volcano-sh https://volcano-sh.github.io/helm-charts
helm repo update

# Install the Volcano chart into its own namespace
helm install volcano volcano-sh/volcano \
  --namespace volcano-system \
  --create-namespace \
  --set custom.scheduler_kube_api_qps=2000 \
  --set custom.scheduler_kube_api_burst=2000
```

### Supported Kubernetes Versions:
Volcano strictly follows a "Kube-3" support window (the current version plus the two previous stable Kubernetes releases). Ensure your target cluster matches:
- **Volcano v1.15**: Supports Kubernetes 1.33 through 1.35.

---

## 2. The `vcctl` CLI

`vcctl` is the specialized command-line utility for managing Volcano resources.

### Installation:
```bash
# Download and install the binary matching your Volcano release version
tar -xvf vcctl-v1.15.0-linux-amd64.tar.gz
sudo mv vcctl /usr/local/bin/
```

### Core Commands:

#### A. Job Management:
```bash
# List all volcano jobs in the current namespace
vcctl job list

# View detailed status of a specific job
vcctl job show --name=distributed-training-job

# Suspend a running job (frees cluster resources, keeps job state)
vcctl job suspend --name=distributed-training-job

# Resume a suspended job
vcctl job resume --name=distributed-training-job

# Delete a job
vcctl job delete --name=distributed-training-job
```

#### B. Queue Management:
```bash
# List all queues
vcctl queue list

# View queue allocations
vcctl queue show --name=ml-team
```

---

## 3. Comprehensive Troubleshooting Checklist

Use this structured guide when workloads are misbehaving.

### A. Problem: Job is stuck in `Pending`

#### 1. Check the PodGroup Status:
Since standard Pods are bound to PodGroups, the `PodGroup` is the scheduler's entry point.
```bash
kubectl get podgroup -n default
# If the phase is "Pending", the scheduler has not admitted the group.
# If "Inqueue", the group is admitted but cannot find node capacity.
```

#### 2. Check the PodGroup Description:
Look for scheduling conditions and messages:
```bash
kubectl describe podgroup distributed-training-job -n default
```

#### 3. Analyze Scheduler Log Events:
Query the scheduler pod logs specifically filtering for your job's namespace or name:
```bash
kubectl logs -n volcano-system -l app=volcano-scheduler | grep -E "(distributed-training-job|invalid|not ready|minAvailable)"
```
- **Common Log Message**: `rejection reason: gang constraint not met`. This indicates the cluster does not have enough resources to run the *entire* minimum number of pods (`minAvailable`) at once.

#### 4. Verify Node Selectors & Taints:
Ensure that any `nodeSelector`, `affinity`, or `tolerations` on your task templates match the actual labels/taints of the cluster nodes. A single un-tolerated taint on a worker node will prevent the entire gang from starting.

---

### B. Problem: Preemption or Reclamation is not working

#### 1. Inspect the Scheduler Configuration Map:
Verify that `preempt` and `reclaim` actions are defined in your `volcano-scheduler.conf` map:
```bash
kubectl get configmap -n volcano-system volcano-scheduler-conf -o yaml
# Look for actions line. Correct value:
# actions: "enqueue, allocate, preempt, reclaim, backfill"
```

#### 2. Verify Queue Metrics & Status:
```bash
kubectl describe queue ml-team
# Confirm that the state is "Open" and resource metrics like "deserved" and "capability" are correct.
```

---

## 4. Performance Tuning Reference

For larger clusters (100+ nodes), the default scheduler and controller parameters must be tuned to prevent CPU bottlenecks and API starvation.

### Configurable Helm Parameters (`values.yaml`):

```yaml
custom:
  # --- API Client QPS & Burst ---
  # Maximum requests per second to the API Server from the scheduler
  scheduler_kube_api_qps: 2000
  # Burst tolerance for scheduler API client
  scheduler_kube_api_burst: 2000
  # Controller Manager API limits
  controller_kube_api_qps: 500
  controller_kube_api_burst: 1000

  # --- Scheduling Threading & Periods ---
  # Duration of the scheduling cycle (lower = faster turnaround, higher CPU)
  scheduler_schedule_period: 500ms
  # Worker threads for evaluating nodes in parallel
  scheduler_node_worker_threads: 30
  # Percentage of nodes to evaluate per pod allocation pass
  scheduler_percentage_nodes_to_find: 50

  # --- Controller Workers ---
  # Threads for processing Volcano Job lifecycle events
  controller_worker_threads: 10
  # Threads for garbage collection processing
  controller_worker_threads_for_gc: 10
  # Threads for managing PodGroup events
  controller_worker_threads_for_podgroup: 10
```

---

## 5. Volcano vs. Default Kubernetes Scheduler

| Feature | Default kube-scheduler | Volcano Scheduler |
|---------|------------------------|-------------------|
| **Scheduling Unit** | Single Pod | PodGroup (Batch Job / Gang) |
| **Atomic (Gang) Scheduling** | No (requires custom plugins) | Yes (`minAvailable` / `minMember` native) |
| **Dominant Resource Fairness** | No | Yes (Hierarchical DRF) |
| **Hierarchical Queues** | No | Yes (`root/parent/child` with weights) |
| **Cross-Queue Reclaim** | No | Yes (reclaims idle resource shares) |
| **Distributed Plugins** | No | Yes (PyTorch, TF, MPI, SSH integrations) |
| **Fractional GPU Scheduling** | No (requires third-party webhooks) | Yes (`deviceshare` fractional device allocation) |
| **NUMA Sockets Alignment** | No | Yes (via `Numatopology` and `volcano-agent`) |
| **Custom Lifecycle Policies** | No | Yes (TaskCompleted -> CompleteJob, etc.) |
