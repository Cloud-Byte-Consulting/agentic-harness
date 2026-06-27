# HorizontalPodAutoscaler (HPA) & Metrics Server

HPA scales the **replica count** of a Deployment/StatefulSet/ReplicaSet (anything with a `scale`
subresource) toward a target metric. API: **`autoscaling/v2`** (use this; `v1` only supported CPU
and stored v2 fields as annotations).

## Contents
- Metrics Server (required)
- HPA basics & the scaling algorithm
- Resource metrics (Utilization / AverageValue)
- Custom & external metrics
- Scaling behavior (stabilization, policies)
- Troubleshooting

---

## Metrics Server (required)

HPA's `Resource` metrics, VPA, and `kubectl top` all depend on Metrics Server. It scrapes each
kubelet's cAdvisor (~every 15s), aggregates CPU/memory into the Metrics API, keeps only the
**current** value in memory (no history — it is *not* a monitoring system).

```bash
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm repo update
helm install metrics-server metrics-server/metrics-server
# Local / self-signed kubelet certs (kind, bare-metal): disable kubelet TLS verification
helm install metrics-server metrics-server/metrics-server --set args="{--kubelet-insecure-tls}"
```

Verify:

```bash
kubectl get pods -n kube-system | grep metrics-server
kubectl top nodes          # shows CPU(cores)/% and MEMORY(bytes)/%
kubectl top pods -n NS     # shows per-pod cores/bytes (no % for pods)
```

If `kubectl top` works, HPA resource metrics will work. Many managed clusters (EKS/GKE/AKS) ship
it pre-installed.

---

## HPA basics & the scaling algorithm

```
desiredReplicas = ceil[ currentReplicas × (currentMetricValue / desiredMetricValue) ]
```

- Polls metrics ~every 15s. **Tolerance band 10%**: if `current/desired` is in `[0.9, 1.1]`, no
  action (prevents churn on minor noise).
- **Multiple metrics**: desiredReplicas is computed per metric; HPA uses the **largest** result.
- Scale-down uses a stabilization window (default 300s) so a brief dip doesn't drop replicas that
  are needed again seconds later.
- HPA never goes below `minReplicas` (≥1 for standard HPA — only KEDA / the alpha `HPAScaleToZero`
  gate allow 0).

Example: target 70% CPU, 5 replicas at 80% → `ceil[5 × 80/70] = 6`. Later at 60% →
`ceil[5 × 60/70] = 5`.

Imperative shortcut (equivalent to a minimal CPU HPA):

```bash
kubectl autoscale deployment montecarlo-pi --cpu-percent=70 --min=1 --max=10
```

---

## Resource metrics (CPU / memory)

The target requires the pod's containers to declare `requests` for that resource (the % is
relative to requests).

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: montecarlo-pi
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: montecarlo-pi
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization        # percentage of the pod's CPU request
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: AverageValue       # absolute value instead of a percentage
        averageValue: 500Mi
```

- `Utilization` → percent of `requests` averaged across pods (needs requests set).
- `AverageValue` → absolute per-pod average (does not need requests).

---

## Custom & external metrics

Three Metrics APIs feed HPA:
- **Resource Metrics API** — CPU/memory, from Metrics Server (above).
- **Custom Metrics API** — application metrics (e.g. p95 latency, requests/sec). Needs a metric
  store **and** an adapter exposing `custom.metrics.k8s.io`. Common: Prometheus + **Prometheus
  Adapter**, which translates a Prometheus query into a k8s metric.
- **External Metrics API** — off-cluster sources (cloud queues, etc.). KEDA is usually the simpler
  path here (see `keda.md`).

**Pods** metric (per-pod, always `AverageValue`):

```yaml
  metrics:
  - type: Pods
    pods:
      metric:
        name: monte_carlo_latency_seconds   # exposed via the custom-metrics adapter
      target:
        type: AverageValue
        averageValue: 500m                   # 0.5s average latency per pod
```

**Object** (scale on a metric describing one object, e.g. Ingress requests/sec):

```yaml
  - type: Object
    object:
      describedObject:
        apiVersion: networking.k8s.io/v1
        kind: Ingress
        name: main-ingress
      metric:
        name: requests_per_second
      target:
        type: Value
        value: "2k"
```

**External**:

```yaml
  - type: External
    external:
      metric:
        name: queue_messages_ready
        selector:
          matchLabels: {queue: "worker_tasks"}
      target:
        type: AverageValue
        averageValue: "30"
```

Verify a custom metric is being served:

```bash
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pods/*/monte_carlo_latency_seconds"
```

To make Prometheus scrape an app, create a `ServiceMonitor` (CRD from kube-prometheus-stack):

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: montecarlo-pi
  labels: {release: prometheus}   # must match the Prometheus instance's serviceMonitorSelector
spec:
  selector:
    matchLabels: {app: montecarlo-pi}
  endpoints:
  - port: http
```

> Boundary: deep Prometheus/Grafana setup belongs to the observability skill. Here, expose just
> the metric the HPA needs.

---

## Scaling behavior (stabilization & policies)

`behavior` controls *how fast* HPA adds/removes replicas. Defaults: scale-up is immediate-ish;
scale-down has a 300s stabilization window.

```yaml
spec:
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 0        # react immediately on spikes
      policies:
      - type: Percent
        value: 100                         # may double replicas...
        periodSeconds: 15                  # ...every 15s
      - type: Pods
        value: 4                           # ...or add up to 4 pods / 15s
      selectPolicy: Max                    # pick the policy allowing the larger change
    scaleDown:
      stabilizationWindowSeconds: 60       # wait 60s of sustained low load before scaling down
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
```

Guidance: scale **up fast** (users feel under-provisioning immediately), scale **down slowly**
(avoid flapping; a re-spike right after scale-down is costly). Some apps misbehave if scaled too
fast — and aggressive scale-up forces more node launches. Tune per workload and test.

---

## Inspecting & troubleshooting

```bash
kubectl get hpa                 # TARGETS column shows current/target
kubectl describe hpa NAME       # Conditions + Events explain decisions
```

`status.conditions` types: `AbleToScale`, `ScalingActive`, `ScalingLimited`. Typical failures:

| Symptom | Likely cause | Fix |
|---|---|---|
| `TARGETS: <unknown>/70%` | Metrics Server down/missing, or no `requests` on container | `kubectl top pods`; install/fix Metrics Server; add CPU/memory requests |
| `AbleToScale False / FailedGetScale` | `scaleTargetRef` name/kind/namespace wrong | Correct the target ref; confirm Deployment exists |
| Custom metric unavailable | Adapter not installed, bad query, ServiceMonitor not matched | Check adapter logs, the raw custom-metrics API, the Prometheus query format/units |
| Replicas flapping | Stabilization window too short / noisy metric | Lengthen `scaleDown.stabilizationWindowSeconds`; smooth the metric |
| Won't scale past N | `maxReplicas` hit, or no node capacity (pods Pending) | Raise `maxReplicas`; add node autoscaling (Karpenter/CA) |

Check Metrics Server logs when resource metrics fail:

```bash
kubectl logs -n kube-system -l app.kubernetes.io/name=metrics-server --all-containers=true --tail=20
```

Common Metrics Server failures: network/RBAC to the Metrics API, kubelet TLS (add
`--kubelet-insecure-tls` on self-signed clusters), or the pod hitting its own memory limit.
