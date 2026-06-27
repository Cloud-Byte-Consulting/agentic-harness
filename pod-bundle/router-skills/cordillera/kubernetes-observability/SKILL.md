---
name: kubernetes-observability
description: >-
  Instrument, monitor, and debug Kubernetes clusters and apps. Use for the three pillars
  (metrics, logs, traces) and the four golden signals, the Metrics Server vs a full Prometheus
  pipeline, Prometheus (scrape model, node-exporter, kube-state-metrics,
  ServiceMonitor/PodMonitor via the Prometheus Operator, kube-prometheus-stack), PromQL for
  queries and recording/alerting rules, Alertmanager and SLO/SLI error budgets, Grafana
  dashboards as code, cluster logging (Fluent Bit/Fluentd, EFK vs Loki, structured logs),
  distributed tracing (OpenTelemetry, Jaeger, Tempo), and a practical debugging workflow
  (kubectl top/logs/events/describe, ephemeral debug containers). Trigger whenever the user
  sets up monitoring or logging, writes PromQL or alerts, builds Grafana dashboards, ships or
  queries logs, adds tracing, defines SLOs, or debugs why an app is slow or erroring
  in-cluster - even without saying observability. Metrics consumed for autoscaling are in
  kubernetes-autoscaling-scheduling.
---

# Kubernetes Observability

This skill equips Claude to design, deploy, and operate observability for Kubernetes — metrics, logs,
and traces — and to debug live clusters and workloads using signals correlated across all three
pillars.

## When to use this skill

- Standing up a monitoring stack: `kube-prometheus-stack`, Prometheus, Grafana, Alertmanager.
- Writing or debugging **PromQL** queries, recording rules, or **PrometheusRule** alerts.
- Wiring app metrics into Prometheus via **ServiceMonitor** / **PodMonitor** (Prometheus Operator).
- Deciding between **Metrics Server** (for `kubectl top`/HPA) and a **full Prometheus pipeline** — they are not interchangeable.
- Building Grafana dashboards or choosing the right panels and queries.
- Centralizing logs: Fluent Bit / Fluentd DaemonSets, EFK / OpenSearch, or **Loki**; structured logging and retention.
- Adding distributed tracing with **OpenTelemetry** (Collector + instrumentation) and **Jaeger** / **Tempo**.
- Defining **SLOs/SLIs** and **error budgets**; routing alerts via Alertmanager.
- Debugging production: `kubectl top/logs/events/describe`, **ephemeral debug containers**, correlating golden signals.

## Core concepts

### Monitoring vs observability
*Monitoring* answers "is a known thing broken?" against pre-defined metrics and thresholds.
*Observability* answers "why is it broken?" — including failures you didn't anticipate — by
correlating the **three pillars**: **metrics** (numeric time series, trends), **logs** (discrete
events, the "why"), and **traces** (one request's path across services). You need all three; metrics
alone tell you something is slow, traces tell you *where*, logs tell you *what* happened.

### The four golden signals and RED/USE
Instrument and alert on **symptoms users feel**, not internal causes:
- **Golden signals** (Google SRE): **latency**, **traffic**, **errors**, **saturation**.
- **RED** (per request-serving service): **R**ate, **E**rrors, **D**uration. Best for app services.
- **USE** (per resource: node, disk, CPU): **U**tilization, **S**aturation, **E**rrors. Best for infra.

Alert on symptoms (high error rate, p95 latency), diagnose with causes (CPU, memory, GC). "Users care
about slow responses, not which container ran out of CPU."

### Metrics Server is NOT a monitoring system
- **Metrics Server**: a lean, in-memory, cluster-wide aggregator of *current* CPU/memory for nodes
  and pods. It scrapes each kubelet (via cAdvisor) ~every 15s and serves the **Metrics API**
  (`metrics.k8s.io`). It powers `kubectl top` and **HPA/VPA**. No history, no disk, no app metrics.
- **Prometheus**: a full time-series database with history, PromQL, alerting, and arbitrary app
  metrics. Use it for monitoring and debugging — **not** Metrics Server.
- They coexist: Metrics Server for autoscaling, Prometheus for everything else. Don't try to make one
  do the other's job. (Autoscaling details: see **kubernetes-autoscaling-scheduling**.)

### Prometheus model in one paragraph
Prometheus **pulls** (scrapes) metrics over HTTP from `/metrics` endpoints exposing a simple text
format: `metric_name{label="value",...} 12.3`. Four metric types: **counter** (monotonically up,
e.g. `http_requests_total`), **gauge** (up/down, e.g. memory in use), **histogram** (bucketed
observations → quantiles, e.g. request latency), **summary** (client-computed quantiles; prefer
histograms). Cluster data comes from exporters: **node-exporter** (per-node OS metrics),
**kube-state-metrics** (object state like `kube_pod_info`, replica counts — NOT from the API server's
own `/metrics`), the **kubelet/cAdvisor** (`container_*`), and the API server. Labels are how you slice
and aggregate. Prometheus runs as a StatefulSet because it persists its TSDB locally.

### The Operator pattern (kube-prometheus-stack)
The **Prometheus Operator** turns config into CRDs. You don't edit `prometheus.yml`; you create
objects the operator reconciles into scrape config:
- **ServiceMonitor** — scrape pods selected via a **Service** (the cloud-native default).
- **PodMonitor** — scrape pods directly, no Service needed.
- **PrometheusRule** — recording rules and alerting rules.
- **Alertmanager / AlertmanagerConfig** — routing, grouping, receivers, silences.

`kube-prometheus-stack` (Helm) bundles Prometheus + Operator + Alertmanager + Grafana + node-exporter
+ kube-state-metrics + ~40 curated alert rules and dashboards. It is the standard starting point.

## Workflow / how to approach observability tasks

### 1. Stand up the metrics stack
Deploy `kube-prometheus-stack` and Metrics Server (Metrics Server only if your distro lacks it):

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace
# Metrics Server (skip on managed clusters that ship it; --kubelet-insecure-tls only for local/dev)
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm install metrics-server metrics-server/metrics-server -n kube-system
```

Verify: `kubectl top nodes`, `kubectl top pods -A`, then check **Status → Target health** in the
Prometheus UI — every intended target should be `UP`. See `references/kube-prometheus-stack.md`.

### 2. Scrape your application
Expose `/metrics` in the app (Prometheus client lib), front it with a Service, then create a
**ServiceMonitor** carrying the label the operator selects on (commonly `release: <helm-release>`,
e.g. `release: kube-prometheus-stack`). The operator needs RBAC to list services/endpoints/pods in the
target namespace. If metrics don't appear, check three places in order: operator logs, the
config-reloader sidecar, then the Prometheus container. Full manifests + the RBAC Role in
`references/metrics-and-prometheus.md`.

### 3. Write queries and alerts
Use PromQL: `rate()` for per-second counter rates, `sum by (label)` to aggregate,
`histogram_quantile()` over `_bucket` series for p95/p99 latency. Test the expression in the
Prometheus UI **before** putting it in a PrometheusRule (the `expr` is identical). Alert on golden
signals with a `for:` duration to avoid flapping. PromQL recipes, golden-signal queries, and
PrometheusRule/SLO examples in `references/promql-and-alerting.md`.

### 4. Route alerts and define SLOs
Connect Prometheus → Alertmanager; configure routes/receivers (Slack, email, PagerDuty), grouping,
and silences. Derive alert thresholds from **SLOs** (e.g. 99.9% success) and spend from the **error
budget**. Every alert should link a runbook. See `references/promql-and-alerting.md`.

### 5. Build dashboards
Grafana reads Prometheus (and Loki, Tempo, Jaeger) as data sources. Provision dashboards as code via
ConfigMaps labeled `grafana_dashboard: "1"` and data sources via a provisioning ConfigMap — never rely
on hand-clicked, unpersisted dashboards. Panel/query guidance in `references/grafana-dashboards.md`.

### 6. Centralize logs
Run a **Fluent Bit** (lightweight) or **Fluentd** (heavier, more transforms) DaemonSet that tails
`/var/log/containers/*.log` on each node, enriches with Kubernetes metadata, and ships to a backend:
**EFK/OpenSearch** (full search, heavier) or **Loki** (label-indexed, cheaper, Grafana-native). Emit
**structured (JSON) logs** with correlation IDs; set retention deliberately. Architecture, manifests,
and the EFK-vs-Loki decision in `references/logging.md`.

### 7. Add tracing
Instrument services with **OpenTelemetry** SDKs (auto-instrumentation where possible), export **OTLP**
to an **OpenTelemetry Collector**, which fans out to **Jaeger** or **Tempo**. A trace is a tree of
spans; propagate context across service calls. Collector/Jaeger manifests and sampling guidance in
`references/tracing-opentelemetry.md`.

### 8. Debug a live incident
Start at symptoms, drill to cause, correlate pillars:

```bash
kubectl get events -n <ns> --sort-by=.lastTimestamp     # what just happened
kubectl describe pod <pod> -n <ns>                       # state, probes, restarts, OOMKilled
kubectl top pod <pod> -n <ns> --containers               # live CPU/mem (needs Metrics Server)
kubectl logs <pod> -n <ns> -c <container> --previous     # logs of the crashed instance
kubectl debug -it <pod> -n <ns> --image=busybox \
  --target=<container>                                    # ephemeral container, shares namespaces
```

Then: Grafana for the trend, traces for the slow span, logs (filtered by the trace/request ID) for the
"why". Full troubleshooting tree, golden-signal correlation, and `kubectl debug` patterns in
`references/debugging-and-golden-signals.md`.

## Common pitfalls & anti-patterns

- **Confusing Metrics Server with monitoring.** Metrics Server has no history and no app metrics; never
  build dashboards or alerts on it. Use Prometheus.
- **"Let's capture everything."** Indiscriminate logging/metric-dumping buries signal in noise and
  burns storage. Log with intent (set levels DEBUG/INFO/WARN/ERROR), prune metrics, alert only on
  actionable thresholds. More data ≠ more insight.
- **Alerting on causes, not symptoms.** A CPU-high alert pages people at 3 a.m. for a non-problem.
  Alert on user-facing symptoms (errors, latency); keep causes for diagnosis. Add `for:` to kill flaps.
- **Alert fatigue.** Too many noisy alerts → desensitization → missed real ones. Group related alerts,
  tune thresholds, automate known fixes, review after every incident. "Alert fatigue is worse than
  missing a single event."
- **Forgetting the ServiceMonitor selector label.** Prometheus isn't multi-tenant; the operator only
  picks up ServiceMonitors carrying the expected label (e.g. `release: ...`). Missing label = silent
  no-scrape.
- **Missing operator RBAC.** Without a Role to list services/endpoints/pods in the target namespace,
  the operator can't resolve the ServiceMonitor — fails quietly. Scrape happens on the **Pod IP**, not
  the Service, so per-host routing apps must serve `/metrics` on all hostnames.
- **Unsecured Prometheus/Alertmanager.** Neither has any auth model ("Security is Not My Problem").
  Metrics map your whole environment for an attacker. Front with an auth proxy (oauth2-proxy/OpenUnison)
  or restrict to `kubectl port-forward`; never expose raw. Grafana ships a hard-coded admin password —
  change it. (RBAC/identity: see **kubernetes-security-rbac**.)
- **Summaries when you want aggregatable quantiles.** Summary quantiles can't be re-aggregated across
  pods; use **histograms** + `histogram_quantile()`.
- **Fragmented / partial tracing.** Tracing only some services leaves blind spots — the real
  bottleneck (often a third-party API) stays invisible. Instrument every hop on critical paths.
- **No persistent dashboards/rules.** Hand-clicked Grafana dashboards and ad-hoc Prometheus config are
  lost on restart. Define them as code (ConfigMaps/CRDs).
- **Observability only in prod.** Shift left — run the same metrics/alerts in staging and load tests so
  latency regressions surface before users hit them.

## Reference files

- **`references/metrics-and-prometheus.md`** — Prometheus architecture, the scrape/text format, metric
  types, exporters (node-exporter, kube-state-metrics, kubelet/cAdvisor), ServiceMonitor/PodMonitor +
  the operator RBAC Role, securing the `/metrics` endpoint, retention & remote-write. Read when wiring
  apps into Prometheus or explaining how scraping works.
- **`references/promql-and-alerting.md`** — PromQL essentials (rate/irate, aggregation, `by`/`without`,
  histogram quantiles, math), golden-signal queries, PrometheusRule examples, Alertmanager routing +
  receivers + silences, SLO/SLI and error budgets, runbooks. Read when writing queries or alerts.
- **`references/grafana-dashboards.md`** — Grafana data sources, dashboard-as-code (ConfigMap label),
  provisioning, useful panels and queries, Grafana vs Prometheus alerting. Read when building or
  provisioning dashboards.
- **`references/logging.md`** — cluster logging architecture, Fluent Bit vs Fluentd DaemonSets, EFK/
  OpenSearch vs Loki, structured logging, retention, Docker log drivers/rotation. Read for any logging
  pipeline task.
- **`references/tracing-opentelemetry.md`** — OpenTelemetry (Collector, SDK instrumentation, OTLP,
  exporters), Jaeger and Tempo, spans/traces, context propagation, sampling, full Kubernetes manifests.
  Read for distributed tracing tasks.
- **`references/kube-prometheus-stack.md`** — installing/configuring the Helm stack, what it bundles,
  values you'll commonly tune, verifying targets, accessing UIs. Read when deploying the stack.
- **`references/debugging-and-golden-signals.md`** — the production debugging workflow, `kubectl
  top/logs/events/describe`, ephemeral debug containers, correlating metrics/logs/traces, impact vs
  diagnostic metrics, common failure signatures (OOMKilled, CrashLoopBackOff, throttling). Read when
  debugging a live workload or cluster.
