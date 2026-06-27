# Resource Management and QoS Classes

Sizing containers with requests/limits, understanding QoS, and debugging OOM/eviction.

## Contents
- requests vs limits
- CPU and memory units
- How limits are enforced (CPU throttling vs OOMKill)
- QoS classes and how they're derived
- Eviction order
- Setting requests/limits in a workload
- LimitRange and ResourceQuota (pointers)
- Sizing guidance and pitfalls

## requests vs limits

Each container can declare resource `requests` and `limits` for CPU and memory:

- **requests** — the amount the scheduler reserves. A Pod is only placed on a node with
  enough unreserved capacity to satisfy the sum of its containers' requests. Requests also
  determine the container's QoS class and its share of contended CPU.
- **limits** — the hard ceiling the kubelet/runtime enforces at runtime.

```yaml
resources:
  requests:
    cpu: 250m
    memory: 256Mi
  limits:
    cpu: "1"
    memory: 512Mi
```

If you set a limit but no request, Kubernetes defaults the request to equal the limit.

## CPU and memory units

**CPU** is measured in cores; `1` = one vCPU/core. Millicores: `500m` = 0.5 core,
`100m` = 0.1 core. CPU is compressible — it can be throttled.

**Memory** is in bytes; use suffixes. Power-of-two (binary): `Ki`, `Mi`, `Gi`
(`256Mi` = 256 × 2²⁰ bytes). Decimal: `K`, `M`, `G`. Prefer the binary suffixes for
memory. Memory is incompressible — you can't throttle it, so exceeding the limit means the
container is killed.

## How limits are enforced

- **CPU over limit → throttled**, not killed. The container is simply not scheduled more
  CPU time than its limit; latency rises but it keeps running.
- **Memory over limit → OOMKilled**. The kernel OOM killer terminates the container; you
  see `Reason: OOMKilled` in `kubectl describe pod` and the container restarts per
  `restartPolicy` (often into `CrashLoopBackOff` if it immediately re-exceeds).

## QoS classes

Kubernetes assigns each Pod a Quality-of-Service class, derived automatically from its
containers' requests/limits. QoS drives eviction priority when a node is under resource
pressure.

| QoS | Condition | Eviction priority |
|---|---|---|
| **Guaranteed** | Every container sets both CPU and memory `limits`, and each request **equals** its limit | Evicted last |
| **Burstable** | At least one container has a request or limit, but not meeting the Guaranteed rule | Evicted after BestEffort |
| **BestEffort** | No container sets any requests or limits | Evicted first |

Examples:
```yaml
# Guaranteed: requests == limits for both resources, on every container
resources:
  requests: { cpu: 500m, memory: 256Mi }
  limits:   { cpu: 500m, memory: 256Mi }

# Burstable: has requests, limits higher (or partially set)
resources:
  requests: { cpu: 100m, memory: 128Mi }
  limits:   { cpu: 500m, memory: 256Mi }

# BestEffort: nothing set anywhere in the Pod  -> first to be killed under pressure
```

Guaranteed Pods get the strongest scheduling and eviction protection; reserve the class
for latency-sensitive or critical workloads. Most app Pods are sensibly Burstable. Avoid
BestEffort in production — those Pods are the first casualties when a node runs low.

## Eviction order

Under node memory/disk pressure the kubelet evicts in this order: BestEffort first, then
Burstable Pods that exceed their memory **requests** (most-over-request first), and
Guaranteed last. This is why setting realistic requests matters: a Burstable Pod staying
under its memory request is far safer than one running well above it.

## Setting requests/limits in a workload

They go on each container in the Pod template, so the same block applies whether the Pod
is managed by a Deployment, StatefulSet, DaemonSet, Job, or CronJob:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  replicas: 3
  selector: { matchLabels: { app: api } }
  template:
    metadata:
      labels: { app: api }
    spec:
      containers:
        - name: api
          image: my-api:1.4
          resources:
            requests: { cpu: 250m, memory: 256Mi }
            limits:   { cpu: "1",  memory: 512Mi }
```

## LimitRange and ResourceQuota (pointers)

- **LimitRange** (namespaced) sets default requests/limits for containers that omit them,
  and min/max bounds. Useful to ensure no BestEffort Pods slip into a namespace.
- **ResourceQuota** (namespaced) caps the aggregate requests/limits (and object counts)
  for a whole namespace — multi-tenancy guardrails.

These are namespace-administration tools; full coverage belongs with cluster
administration/multi-tenancy material rather than this workloads skill, but know they exist
and that a LimitRange can silently inject the requests/limits you didn't specify.

## Sizing guidance and pitfalls

- **Always set requests and limits** on production containers. Omitting them → BestEffort
  → first evicted, and the scheduler can't place the Pod well.
- **Right-size from observed usage**, not guesses. Over-requesting wastes capacity and
  reduces packing density; under-requesting causes OOMKills and CPU starvation. Measure
  with metrics (Prometheus/metrics-server) and adjust.
- **Memory limit too low → OOMKilled / CrashLoopBackOff.** Raise the limit or fix the leak;
  check `kubectl describe pod` for the `OOMKilled` reason and the restart count.
- **CPU limit too low → throttling and latency**, not crashes. If a service is slow under
  load but not restarting, suspect CPU throttling.
- **For HPA**, requests are the denominator for CPU/memory-percentage targets — wrong
  requests make autoscaling misbehave. HPA itself is covered in
  **kubernetes-autoscaling-scheduling**.
- **Guaranteed for critical pods**: set request == limit for both CPU and memory to get the
  class and its eviction protection.
