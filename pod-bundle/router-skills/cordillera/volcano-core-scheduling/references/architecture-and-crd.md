# Volcano Component Architecture & Core CRD Reference

This reference deep-dives into the physical components of a Volcano deployment, the relationship between controllers and webhooks, and the exact specifications of the core custom resource definitions (CRDs): `Job`, `Queue`, `PodGroup`, and `Command`.

---

## 1. Component Architecture

A running Volcano cluster deploys three core binaries inside the `volcano-system` namespace:

```
                  ┌──────────────────────────────────────────────┐
                  │            KUBERNETES API SERVER             │
                  └──────▲──────────────────▲───────────────▲────┘
                         │                  │               │
        Read/Write Pods/ │                  │ Mutate/       │ Read Config/
        PodGroups        │                  │ Validate      │ Bind Pods
                         │                  │               │
┌────────────────────────┴─┐       ┌────────┴─────────┐     │ ┌────────────────────────┐
│  vc-controller-manager   │       │  vc-webhook      │     │ │      vc-scheduler      │
│  (pkg/controllers/)      │       │  (pkg/webhooks/) │     │ │      (pkg/scheduler/)  │
│                          │       │                  │     │ │                        │
│ • Reconciles Volcano Jobs│       │ • Mutating:      │     │ │ • Cycle (default 1s)   │
│ • Manages PodGroups      │       │   injects PG env,│     │ │ • Actions: Enqueue,    │
│ • Handles Command bus    │       │   sets default PG│     │ │   Allocate, Preempt... │
│ • Job lifecycle policies │       │ • Validating:    │     │ │ • Plugins: gang, drf,  │
│   (eviction, restarts)   │       │   schema checks  │     │ │   predicates...        │
└──────────────────────────┘       └──────────────────┘     │ └────────────────────────┘
                                                            │
                                                     vcctl  │ (Command signaling)
                                                     CLI ───┘
```

### vc-scheduler (Scheduler)
The core scheduling engine. Instead of scheduling pods one-by-one, it works in **cycles** (default 1s). It reads the entire state of PodGroups, Queues, Nodes, and Pods, builds a local cache, and runs pipeline actions to make batch placement decisions. Once decisions are made, it sends `Bind` requests to the API Server.

### vc-controller-manager (Controller Manager)
Handles the non-scheduling lifecycle of Volcano resources. It reconciles `batch.volcano.sh/v1alpha1` `Job` and `CronJob` resources, creating the underlying Pods and the matching `PodGroup`. It is also responsible for executing job-level policies (e.g., restarting jobs, completing jobs when a task completes) and managing the command bus (`bus.volcano.sh/v1alpha1`).

### vc-webhook-manager (Admission Webhook)
A mutating and validating webhook. It does three critical things:
1. **Validation**: Rejects invalid queue paths, mismatched task replicas, or missing required fields before they persist in etcd.
2. **PodGroup Injection**: Automatically injects a default `PodGroup` for any pod scheduled with `schedulerName: volcano` if it doesn't already have one, preventing the scheduler from getting stuck.
3. **Environment Injection**: Injects specific environment variables (like master endpoints or task indexes) into distributed training pods (e.g., PyTorch, TensorFlow).

---

## 2. Core Custom Resource Definitions (CRDs)

Volcano's APIs are grouped under `volcano.sh`. The core groups are:
- `batch.volcano.sh/v1alpha1` (Jobs, CronJobs)
- `scheduling.volcano.sh/v1beta1` (Queues, PodGroups)
- `bus.volcano.sh/v1alpha1` (Commands)

---

### A. Queue (`scheduling.volcano.sh/v1beta1`)
Queues are cluster-scoped resources used to partition capacity and establish fair-share boundaries across tenants.

```yaml
apiVersion: scheduling.volcano.sh/v1beta1
kind: Queue
metadata:
  name: ml-team
spec:
  weight: 2                   # Relative share weight (used by proportion/drf plugins)
  priority: 10                # Higher priority queues get evaluated first
  reclaimable: true           # If true, other queues can reclaim resources when idle
  parent: root                # Parent queue name (for hierarchical queues)
  capability:                 # HARD limit on resources this queue can consume
    cpu: "64"
    memory: "256Gi"
    nvidia.com/gpu: "8"
  deserved:                   # SOFT target (shared if unused, reclaimed if needed)
    cpu: "32"
    memory: "128Gi"
    nvidia.com/gpu: "4"
  guarantee:                  # RESERVED resources (never shared, always available)
    resource:
      cpu: "10"
      memory: "32Gi"
status:
  state: Open                 # Open, Closed, Closing, Unknown
  allocated:                  # Current resource usage
    cpu: "24"
    memory: "96Gi"
    nvidia.com/gpu: "4"
```

#### Queue Specs Explained:
- `weight`: Expresses the ratio of fair-share allocation. If Queue A has weight 1 and Queue B has weight 2, Queue B will receive twice the resources of Queue A under heavy contention.
- `deserved`: Under the `capacity` plugin, this acts as the "soft ceiling". If a queue runs above its `deserved` allocation using idle cluster space, other queues can **evict** its pods to reclaim their deserved share.
- `guarantee`: Absolute reservation. Standard pods scheduled to this queue cannot be preempted below this resource threshold.

---

### B. PodGroup (`scheduling.volcano.sh/v1beta1`)
A PodGroup represents a co-scheduling group (a "gang"). It is the unit evaluated by the Volcano scheduler.

```yaml
apiVersion: scheduling.volcano.sh/v1beta1
kind: PodGroup
metadata:
  name: tf-distributed-training
  namespace: ml-workloads
spec:
  minMember: 5                # Minimum number of pods that must be schedulable
  queue: ml-team              # The queue this PodGroup belongs to
  priorityClassName: high-priority
  minResources:               # Minimum aggregate resources required to run the gang
    cpu: "10"
    memory: "40Gi"
    nvidia.com/gpu: "4"
status:
  phase: Pending              # Pending, Inqueue, Running, Succeeded, Failed, Unknown
  conditions:
    - type: Scheduled
      status: "True"
      lastTransitionTime: "2026-06-13T13:01:00Z"
  running: 0
  succeeded: 0
  failed: 0
```

#### PodGroup Phases:
1. `Pending`: Created, awaiting admission by the scheduler into a queue.
2. `Inqueue`: Admitted into a queue. Ready to be allocated nodes once resources are free.
3. `Running`: At least `minMember` pods have been successfully bound to nodes and are running.
4. `Succeeded`: All pods in the group finished with exit code 0.
5. `Failed`: One or more pods failed, or the group failed to meet its gang requirements within the timeout.

---

### C. Job (`batch.volcano.sh/v1alpha1`)
The `vcjob` provides rich batch job control, task-specific partitioning, plugins, and lifecycle policy definitions.

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: distributed-training-job
  namespace: default
spec:
  minAvailable: 3             # Gang scheduling constraint (Job level)
  queue: ml-team              # Queue to submit to
  schedulerName: volcano      # Must be set to volcano
  maxRetry: 3                 # Retries before marking job failed
  policies:                   # Lifecycle events and actions
    - event: PodFailed
      action: RestartJob
    - event: PodEvicted
      action: RestartJob
  plugins:                    # Workload augmentations
    env: []                   # Injects VOLCANO_TEST, etc.
    svc: []                   # Generates hostnames and headless services
  tasks:
    - name: ps                # Parameter Server task
      replicas: 1
      template:
        spec:
          containers:
            - name: tensorflow
              image: tensorflow/tensorflow:latest-gpu
              resources:
                requests:
                  cpu: "2"
                  memory: "4Gi"
          restartPolicy: OnFailure
    - name: worker            # Worker task
      replicas: 2
      template:
        spec:
          containers:
            - name: tensorflow
              image: tensorflow/tensorflow:latest-gpu
              resources:
                requests:
                  cpu: "4"
                  memory: "8Gi"
                  nvidia.com/gpu: "1"
          restartPolicy: OnFailure
```

#### Job Lifecycle Policies:
You can map specific events or exit codes to actions to build self-healing distributed workflows:
- **Events**: `PodFailed`, `PodEvicted`, `TaskCompleted`.
- **Actions**: `RestartJob` (recreates all pods), `AbortJob` (stops job, leaves resource allocations), `CompleteJob` (gracefully stops remaining tasks), `TerminateJob` (hard delete).
- **Example**: In a PyTorch job, if the master task finishes, you can trigger `CompleteJob` so that worker tasks (which loop indefinitely waiting for work) are gracefully stopped.

---

### D. Command (`bus.volcano.sh/v1alpha1`)
Used for event signaling and control plane communication. It allows users or controllers to request state changes on running Jobs.

```yaml
apiVersion: bus.volcano.sh/v1alpha1
kind: Command
metadata:
  name: abort-job-command
  namespace: default
action: AbortJob              # Action: AbortJob, RestartJob, ResumeJob
target:
  apiVersion: batch.volcano.sh/v1alpha1
  kind: Job
  name: distributed-training-job
  namespace: default
reason: "OperatorInitiated"
message: "Manual cancellation by cluster administrator"
```
