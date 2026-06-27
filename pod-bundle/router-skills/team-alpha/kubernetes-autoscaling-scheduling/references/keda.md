# KEDA — Kubernetes Event-Driven Autoscaling

KEDA extends HPA with event-driven scaling and **scale-to-zero**. You declare a `ScaledObject`
(or `ScaledJob`); KEDA creates and manages an HPA under the hood and feeds it metrics directly
from the source (queue, Prometheus, cron, cloud service) via its own metrics server — no separate
custom-metrics adapter needed. CNCF-graduated. CRD group: `keda.sh/v1alpha1`.

## Contents
- Architecture & install
- ScaledObject (Deployments/StatefulSets)
- ScaledJob (batch)
- Common scalers (cpu/memory, prometheus, cron, queues)
- Scale-to-zero
- TriggerAuthentication & cloud auth
- Scaling speed, caching, fallback, pausing, scaling modifiers
- Monitoring & troubleshooting

---

## Architecture & install

Three components: **Operator** (watches CRDs, manages HPAs, queries sources directly), **Metrics
Server** (exposes event metrics via the External Metrics API to the HPA), **Admission Webhooks**
(validate; prevent two ScaledObjects targeting one workload).

Division of labor: **KEDA owns activation 0↔1 (and 1↔0); HPA owns scaling 1↔N.** Don't run a
manual HPA on a workload KEDA manages — they'd conflict.

```bash
helm repo add kedacore https://kedacore.github.io/charts
helm repo update
helm install keda kedacore/keda --namespace keda --create-namespace
kubectl get pods -n keda   # keda-operator, keda-operator-metrics-apiserver, keda-admission-webhooks
```

---

## ScaledObject (Deployments / StatefulSets)

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: consumer-app-scaler
spec:
  scaleTargetRef:
    name: consumer-app          # add apiVersion+kind for StatefulSet/CRD targets
  pollingInterval: 30           # how often KEDA checks the trigger (s); default 30
  cooldownPeriod: 300           # wait after last active trigger before scaling to 0; default 300
  minReplicaCount: 0            # default 0 (scale to zero)
  maxReplicaCount: 30           # default 100
  triggers:
  - type: rabbitmq
    metadata:
      queueName: orders
      mode: QueueLength
      value: "20"               # target ~20 messages per replica
    authenticationRef:
      name: rabbitmq-auth
```

Multiple triggers → like HPA, KEDA computes replicas per trigger and uses the **max**. You can
pause autoscaling via annotations (see below).

---

## ScaledJob (batch / run-to-completion)

For work where interrupting a pod mid-processing is costly, use `ScaledJob`: KEDA creates **one
Job per unit of work** and lets each finish cleanly (instead of guessing a `cooldownPeriod` for a
Deployment).

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledJob
metadata:
  name: queue-job-consumer
spec:
  jobTargetRef:                 # a full Job spec template (required)
    template:
      spec:
        containers:
        - name: worker
          image: my-worker:1.0
        restartPolicy: Never
  pollingInterval: 10
  maxReplicaCount: 15           # max Jobs created per poll (running Jobs are subtracted)
  successfulJobsHistoryLimit: 1
  failedJobsHistoryLimit: 2
  triggers:
  - type: rabbitmq
    metadata: {queueName: orders, mode: QueueLength, value: "20"}
    authenticationRef: {name: rabbitmq-auth}
```

> ScaledJob does **not** support `useCachedMetrics`, `fallback`, or `scalingModifiers`.

---

## Common scalers (70+ available)

**CPU / memory** (same role as HPA, lets you standardize on KEDA):

```yaml
  triggers:
  - type: cpu
    metricType: Utilization      # or AverageValue
    metadata: {value: "70"}
  - type: memory
    metricType: Utilization
    metadata: {value: "80"}
```

**Prometheus** (go straight to the source — no Prometheus Adapter, supports multiple Prometheus
instances, no 1MB ConfigMap limit):

```yaml
  - type: prometheus
    metadata:
      serverAddress: http://prometheus-kube-prometheus-prometheus.monitoring.svc:9090
      metricName: monte_carlo_latency_seconds
      threshold: "0.5"           # seconds — express target in base units, not "500m"
      query: >
        sum(histogram_quantile(0.95,
          rate(monte_carlo_latency_seconds_bucket{namespace="default",pod=~"app-.*"}[2m])))
```

**Cron** (time windows — e.g. off-hours scale-to-zero). It defines *windows*, not recurring jobs;
outside the window the target drops to `minReplicaCount`:

```yaml
  - type: cron
    metadata:
      timezone: US/Pacific       # IANA tz
      start: "0 8 * * 1-5"       # 8am Mon–Fri
      end:   "0 18 * * 1-5"      # 6pm Mon–Fri
      desiredReplicas: "1"
```

Combine cron + cpu to keep ≥1 replica during business hours and still burst on load, while
scaling to 0 overnight (set `minReplicaCount: 0`).

**Queue scalers** (RabbitMQ, AWS SQS, Azure Service Bus, Kafka, etc.) — the canonical
scale-to-zero use case. SQS example:

```yaml
  - type: aws-sqs-queue
    authenticationRef: {name: aws-credentials}
    metadata:
      queueURL: https://sqs.eu-west-1.amazonaws.com/123456789/orders
      queueLength: "5"
      awsRegion: "eu-west-1"
```

---

## Scale-to-zero

KEDA's headline feature: `minReplicaCount: 0` lets a Deployment go to **zero** when idle and wake
on the next event. Great for queue/event workers and dev environments. **Not** for synchronous
HTTP APIs — at zero there's nothing to serve the first request (cold start / timeout). For HTTP,
keep ≥1 replica, or use KEDA's **HTTP add-on** (`HTTPScaledObject`; beta) or a cron floor.

---

## TriggerAuthentication & cloud auth

Decouple credentials from the trigger and reuse them. Namespaced `TriggerAuthentication` or
cluster-wide `ClusterTriggerAuthentication`:

```yaml
apiVersion: keda.sh/v1alpha1
kind: TriggerAuthentication
metadata:
  name: rabbitmq-auth
spec:
  secretTargetRef:
  - parameter: host
    name: rabbitmq-consumer-secret
    key: RabbitMqHost
```

Cloud providers — prefer **pod/workload identity**, and give the *workload's* identity (not KEDA's
operator) the metric-read permissions. KEDA assumes the workload identity to read the scaler:

```yaml
# AWS (IRSA / EKS Pod Identity)
spec:
  podIdentity:
    provider: aws
    identityOwner: workload      # KEDA uses the scaled workload's IAM role
---
# Azure
spec:
  podIdentity:
    provider: azure-workload
    identityId: $MI_CLIENT_ID
---
# GCP
spec:
  podIdentity:
    provider: gcp                # impersonates the workload's GSA
```

Security principle: don't grant KEDA's operator broad cloud permissions; let each workload's
identity carry only the permissions its scaler needs (e.g. SQS `ReceiveMessage`,
`GetQueueAttributes`).

---

## Scaling speed, caching, fallback, pausing, modifiers

**Speed** — same `behavior` as HPA, under `advanced.behavior`:

```yaml
spec:
  advanced:
    behavior:
      scaleUp:   {stabilizationWindowSeconds: 0,  policies: [{type: Percent, value: 100, periodSeconds: 15}]}
      scaleDown: {stabilizationWindowSeconds: 60, policies: [{type: Percent, value: 100, periodSeconds: 15}]}
```

**Cache metrics** to reduce calls to a throttled source (per trigger). Not for cpu/memory/cron or
ScaledJobs:

```yaml
  triggers:
  - type: rabbitmq
    useCachedMetrics: true
    metadata: {...}
```

**Fallback** — hold a safe replica count if the scaler errors repeatedly (not for ScaledJobs,
cpu/memory, or `AverageValue` scalers):

```yaml
spec:
  fallback:
    failureThreshold: 3
    replicas: 6
```

**Pause** (maintenance/incident) — add to the ScaledObject's `metadata.annotations`:

```yaml
    autoscaling.keda.sh/paused-replicas: "0"   # pin to this count and pause
    autoscaling.keda.sh/paused: "true"         # pause, keep current count
```

To resume, **remove** the annotation (setting `paused: "false"` does not resume).

**Scaling modifiers** (GA 2.15) — combine multiple trigger metrics into one composite via a
formula (Expr language); the composite **replaces** all individual metrics sent to HPA. Each
referenced trigger needs a unique `name`. Not for cpu/memory or ScaledJobs.

```yaml
spec:
  advanced:
    scalingModifiers:
      target: "5"
      metricType: "AverageValue"
      formula: "db_write_latency < 100 ? queue : 5"   # pause queue scaling when DB is stressed
  triggers:
  - type: rabbitmq
    name: queue
    metadata: {queueName: orders, mode: QueueLength, value: "5"}
  - type: metrics-api
    name: db_write_latency
    metadata: {targetValue: "100", url: "http://dummyapi.default.svc/", valueLocation: "database.metrics.write_latency"}
```

---

## Monitoring & troubleshooting

KEDA exposes Prometheus metrics on `:8080/metrics` (enable via Helm `prometheus.operator.enabled`
+ ServiceMonitor). Useful series: `keda_scaler_active`, `keda_scaler_metrics_value`,
`keda_scaler_errors_total`, `keda_scaled_object_errors_total`, `keda_scaler_metrics_latency_seconds`.
There's an official Grafana dashboard in the KEDA repo.

```bash
kubectl describe scaledobject NAME            # Events: ScaledObjectCheckFailed, KEDAScalerFailed, ...
kubectl get hpa keda-hpa-NAME                 # KEDA names the generated HPA keda-hpa-<so-name>
kubectl logs -n keda -l app=keda-operator --all-containers=true --tail=20   # the "brain"
```

Common issues:
- **Target workload doesn't exist** → `ScaledObjectCheckFailed: Target resource doesn't exist`.
- **Scaler auth/connectivity** → check `keda-operator` logs and `keda_scaler_errors_total`.
- **CPU/memory scaler not working** → Metrics Server missing.
- **NetworkPolicy** blocking KEDA → source unreachable.
- Two ScaledObjects on one workload → blocked by the admission webhook.

**Upgrades:** pin the version, read release notes (CRDs can change). Helm
`helm upgrade keda kedacore/keda -n keda`, or GitOps with `kubectl replace -f keda-<ver>.yaml`.
