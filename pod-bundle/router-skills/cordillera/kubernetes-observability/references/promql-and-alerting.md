# PromQL & Alerting

Enough PromQL to write useful queries and alerts, plus PrometheusRule, Alertmanager routing, SLO/error
budgets, and runbooks.

## Contents
- [PromQL essentials](#promql-essentials)
- [Rates: rate vs irate vs increase](#rates-rate-vs-irate-vs-increase)
- [Aggregation](#aggregation)
- [Histograms & quantiles](#histograms--quantiles)
- [Golden-signal queries](#golden-signal-queries)
- [Recording rules](#recording-rules)
- [PrometheusRule: alerting](#prometheusrule-alerting)
- [Alertmanager: routing, receivers, silences](#alertmanager-routing-receivers-silences)
- [SLOs, SLIs & error budgets](#slos-slis--error-budgets)
- [Runbooks](#runbooks)
- [Prometheus alerts vs Grafana alerts](#prometheus-alerts-vs-grafana-alerts)

## PromQL essentials

A query starts with a metric and optional label matchers, then applies functions/operators:

```promql
kube_pod_info                                  # all series
kube_pod_info{namespace="prod"}                # filter by label (=, !=, =~ regex, !~ neg-regex)
container_memory_working_set_bytes{namespace="monitoring"}
```

- **Instant vector** — one value per series at one instant (`kube_pod_info`).
- **Range vector** — a window of samples per series, needed by `rate()` etc. (`http_requests_total[5m]`).
- Math uses infix notation and matches on labels:
  `(used / total) * 100`.

Cluster CPU utilization % (combining metrics + math), from the kube-prometheus dashboards:

```promql
(sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate)
  / max(count without(cpu,mode,pod) (node_cpu_seconds_total{mode="idle"}))) * 100
```

## Rates: rate vs irate vs increase

Counters only go up, so you almost always wrap them:
- `rate(http_requests_total[5m])` — average per-second rate over the window. Use for alerting/graphs;
  smooth, resilient to scrape gaps.
- `irate(http_requests_total[5m])` — instantaneous rate from the last two samples. Spiky; for
  fast-moving graphs only.
- `increase(http_requests_total[1h])` — total increase over the window (= `rate * window`).

Rule of thumb: alerts and dashboards → `rate`; the window should be ≥ 4× the scrape interval.

## Aggregation

Collapse series with aggregation operators, controlling labels with `by` (keep these) or `without`
(drop these):

```promql
sum by (namespace) (kube_pod_info)                       # pods per namespace
sum by (pod) (rate(container_cpu_usage_seconds_total{namespace="monitoring"}[5m]))
avg by (instance) (node_load1)
max without (cpu) (node_cpu_seconds_total{mode="idle"})
topk(5, sum by (pod) (rate(container_cpu_usage_seconds_total[5m])))   # noisiest 5 pods
count(up == 0)                                            # how many targets are down
```

Operators: `sum`, `avg`, `min`, `max`, `count`, `stddev`, `topk`, `bottomk`, `quantile`.

## Histograms & quantiles

Histogram metrics expose `<name>_bucket{le="..."}` cumulative buckets. Compute a quantile by summing
bucket rates by the `le` label:

```promql
# p95 request latency across all pods of a service
histogram_quantile(0.95,
  sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# p99 per route
histogram_quantile(0.99,
  sum(rate(http_request_duration_seconds_bucket[5m])) by (le, route))
```

Always `rate()` the buckets first, always keep `le` in the `by` clause. p50/p95/p99 tell you the
distribution; the average hides tail latency.

## Golden-signal queries

```promql
# Traffic — requests/sec
sum(rate(http_requests_total[5m]))

# Errors — 5xx error ratio
sum(rate(http_requests_total{code=~"5.."}[5m]))
  / sum(rate(http_requests_total[5m]))

# Latency — p95
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# Saturation — CPU throttling (pod is hitting its CPU limit)
sum by (pod) (rate(container_cpu_cfs_throttled_periods_total[5m]))
  / sum by (pod) (rate(container_cpu_cfs_periods_total[5m]))

# Saturation — memory vs limit
sum by (pod) (container_memory_working_set_bytes)
  / sum by (pod) (kube_pod_container_resource_limits{resource="memory"})

# Restarts (CrashLoop signal)
sum by (pod) (increase(kube_pod_container_status_restarts_total[15m]))
```

## Recording rules

Pre-compute expensive/recurring expressions so dashboards and alerts read cheap, stable series. Naming
convention: `level:metric:operation`.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: myapp-recording
  namespace: monitoring
  labels:
    release: kube-prometheus-stack        # so the operator adopts it
spec:
  groups:
    - name: myapp.recording
      interval: 30s
      rules:
        - record: job:http_requests:rate5m
          expr: sum by (job) (rate(http_requests_total[5m]))
        - record: job:http_error_ratio:rate5m
          expr: |
            sum by (job) (rate(http_requests_total{code=~"5.."}[5m]))
              / sum by (job) (rate(http_requests_total[5m]))
```

## PrometheusRule: alerting

Alerts are PromQL with a `for:` (must stay true this long before firing — kills flapping), `labels`
(for routing/severity), and `annotations` (human text + runbook link).

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: myapp-alerts
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  groups:
    - name: myapp.alerts
      rules:
        - alert: HighErrorRate
          expr: |
            sum(rate(http_requests_total{job="myapp",code=~"5.."}[2m]))
              / sum(rate(http_requests_total{job="myapp"}[2m])) > 0.05
          for: 1m
          labels:
            severity: warning
            service: myapp
          annotations:
            summary: "High error rate in myapp"
            description: "5xx error ratio >5% for over 1m."
            runbook_url: "https://git.example.com/runbooks/myapp-errors.md"

        - alert: HighLatencyP95
          expr: |
            histogram_quantile(0.95,
              sum(rate(http_request_duration_seconds_bucket{job="myapp"}[2m])) by (le)) > 1
          for: 1m
          labels:
            severity: critical
            service: myapp
          annotations:
            summary: "High p95 latency in myapp"
            description: "p95 latency exceeds 1s for over 1m."

        - alert: ServiceDown            # absent() catches a metric vanishing entirely
          expr: absent(up{job="myapp"} == 1)
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: "myapp is not reporting any healthy targets"
```

`absent()` is the right tool when "the metric is gone" is the failure (e.g. the app stopped exporting
`active_sessions`) — `active_sessions < 1` would never fire because there's no series to evaluate.

Test every `expr` in the Prometheus UI **before** committing it; the alert `expr` and a graph query are
identical PromQL. Confirm rules loaded at `http://<prometheus>:9090/rules`.

## Alertmanager: routing, receivers, silences

Prometheus fires alerts; **Alertmanager** dedupes, groups, routes, and notifies. Point Prometheus at it
(in `prometheus.yml`, or the Prometheus CR sets this automatically in the stack):

```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets: ["alertmanager-operated.monitoring.svc:9093"]
```

Alertmanager config (`alertmanager.yml`, via a Secret/ConfigMap or the `Alertmanager` CR):

```yaml
global:
  resolve_timeout: 5m
route:
  receiver: default
  group_by: ['alertname', 'service']
  group_wait: 30s          # wait to batch the first notification of a new group
  group_interval: 5m       # wait before sending an updated group
  repeat_interval: 2h      # re-notify if still firing
  routes:
    - matchers: ['severity = "critical"']
      receiver: pager
receivers:
  - name: default
    slack_configs:
      - api_url: "https://hooks.slack.com/services/XXX/YYY/ZZZ"
        channel: "#alerts"
        title: "{{ .CommonAnnotations.summary }}"
        text: "{{ .CommonAnnotations.description }}"
  - name: pager
    pagerduty_configs:
      - routing_key: "<integration-key>"
    email_configs:
      - to: "oncall@example.com"
        from: "alertmanager@example.com"
        smarthost: "smtp.example.com:587"
        auth_username: "alertmanager@example.com"
        auth_password: "<password>"
```

With the operator you can scope routing per namespace using **AlertmanagerConfig** (the object's
namespace is auto-added to its matchers):

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: AlertmanagerConfig
metadata:
  name: critical-alerts
  namespace: kube-system
  labels:
    alertmanagerConfig: critical
spec:
  route:
    receiver: webhook
    groupBy: ['namespace']
    groupWait: 30s
    groupInterval: 5m
    repeatInterval: 30s
    matchers:
      - name: severity
        matchType: "="
        value: critical
  receivers:
    - name: webhook
      webhookConfigs:
        - sendResolved: true
          url: http://my-webhook.svc/webhook
```

**Silences** are created in the Alertmanager UI by label matchers, with a duration and author — use them
for known outages, upstream issues outside your control, or while actively fixing. Note: silences are
*not* Kubernetes objects, so you can't audit them via the API. (A silence can also hide an attacker's
tracks — restrict UI access.)

Receivers exist for Slack, email, PagerDuty, Opsgenie, MS Teams, generic webhook, and more — you rarely
need to write your own.

## SLOs, SLIs & error budgets

- **SLI** (indicator) — a measured ratio of good events, e.g. `successful_requests / total_requests`
  or `fast_requests / total_requests`.
- **SLO** (objective) — the target, e.g. "99.9% of requests succeed over 30 days."
- **Error budget** — `1 − SLO` = how much failure you're allowed (99.9% → 0.1% → ~43 min/month). Spend
  it on releases; when it's exhausted, freeze risky changes and stabilize.

Alert on **burn rate** (how fast you're consuming the budget), not just an instantaneous threshold —
this catches both fast outages and slow leaks:

```promql
# Fast burn: error budget being consumed 14.4x too fast over 1h → page
(1 - (sum(rate(http_requests_total{code!~"5.."}[1h])) / sum(rate(http_requests_total[1h]))))
  > (14.4 * 0.001)            # 0.001 = 1 - 0.999 SLO
```

DORA framing for delivery health: deployment frequency, lead time, change-failure rate, time-to-restore
— pair with SLOs to balance velocity against reliability.

## Runbooks

Every alert should link a runbook. A good entry is short and actionable:
- **Summary** — what fired.
- **Impact** — what users/systems feel.
- **Immediate checks** — confirm the alert; open the dashboard; check downstream deps.
- **Potential causes** — recent deploy, dependency timeout, resource exhaustion.
- **Remediation** — check logs/traces, roll back if post-deploy, scale or restart, then verify recovery.
- **References** — dashboard, trace, related runbooks.

Link it from the alert annotation (`runbook_url`) so the signal carries its next step. Review and update
runbooks after every incident.

## Prometheus alerts vs Grafana alerts

| | Prometheus + Alertmanager | Grafana unified alerting |
|---|---|---|
| Defined as | YAML rules (versioned) | Visually on panels, or provisioned |
| Scope | Global, all metrics | Often per-dashboard/team |
| Routing | Centralized via Alertmanager | Grafana contact points |
| Use for | Production, versioned alerts | Dashboard-level/experimental |

Many teams use both: Prometheus for critical versioned alerts, Grafana for supplementary visual ones.
See `grafana-dashboards.md` for Grafana-side alerting.
