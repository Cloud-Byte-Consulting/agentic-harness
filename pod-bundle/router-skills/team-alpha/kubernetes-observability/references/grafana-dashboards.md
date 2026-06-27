# Grafana Dashboards

Connecting data sources, provisioning dashboards as code, picking useful panels, and Grafana-side
alerting.

## Contents
- [Accessing Grafana in kube-prometheus-stack](#accessing-grafana-in-kube-prometheus-stack)
- [Data sources as code](#data-sources-as-code)
- [Dashboards as code (ConfigMap sidecar)](#dashboards-as-code-configmap-sidecar)
- [Building a panel](#building-a-panel)
- [Useful panels & queries](#useful-panels--queries)
- [Reusing the bundled dashboards](#reusing-the-bundled-dashboards)
- [Grafana alerting](#grafana-alerting)

## Accessing Grafana in kube-prometheus-stack

The stack ships Grafana pre-wired to Prometheus with curated dashboards. Default login is
`admin` / `prom-operator` (the stack's default) — **change it**. Reach it via port-forward:

```bash
kubectl -n monitoring port-forward svc/kube-prometheus-stack-grafana 3000:80
# then http://localhost:3000
```

Grafana has its own auth model but the community edition has only Admin/Viewer roles. For real SSO,
configure OIDC or front Grafana with an auth proxy (the stack's chart supports proxy auth via an HTTP
header). Never expose Grafana publicly with the default password.

## Data sources as code

Provision data sources via a ConfigMap so they survive restarts and are reproducible. Prometheus is the
default; add Loki and Tempo/Jaeger for logs and traces.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
  namespace: monitoring
  labels:
    grafana_datasource: "1"        # the stack's sidecar imports labeled ConfigMaps
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
      - uid: prometheus
        name: Prometheus
        type: prometheus
        access: proxy
        url: http://kube-prometheus-stack-prometheus.monitoring.svc:9090
        isDefault: true
      - uid: loki
        name: Loki
        type: loki
        access: proxy
        url: http://loki-gateway.monitoring.svc
      - uid: tempo
        name: Tempo
        type: tempo
        access: proxy
        url: http://tempo.monitoring.svc:3200
```

Self-managed Grafana (e.g. the simple Compose/standalone install) sets the data source by hand:
Connections → Data sources → Add → Prometheus, URL `http://prometheus:9090`, Save & Test.

## Dashboards as code (ConfigMap sidecar)

A dashboard is JSON (a dataset + visualization rules). Don't hand-click and lose it on restart — store
the JSON in a ConfigMap labeled `grafana_dashboard: "1"`; the stack's sidecar auto-loads it within
seconds.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-dashboard
  namespace: monitoring
  labels:
    grafana_dashboard: "1"          # the magic label Grafana watches for
data:
  myapp.json: |-
    {
      "title": "myapp – RED",
      "panels": [ /* ... exported dashboard JSON ... */ ],
      "schemaVersion": 39,
      "version": 1
    }
```

Workflow: build the dashboard in the Grafana UI → **Export → Save to file (JSON)** → drop the JSON into
a ConfigMap with the label → apply. Now it's versioned and rolls out anywhere (this is exactly how
GitOps distributes dashboards across many clusters — see **kubernetes-gitops-cicd**).

## Building a panel

To find the query behind any panel, open the dashboard, edit the panel, and read its PromQL — then test
that same query in the Prometheus UI. To build one:
1. Add visualization → pick the Prometheus data source.
2. Enter a PromQL query (see below).
3. Choose a visualization (time series, stat, gauge, table, heatmap).
4. Set unit, legend, thresholds; Apply; Save.

Note: dashboard queries often omit a `cluster` label because dashboards are built to span multiple
clusters, whereas a single Prometheus only sees one cluster — so the same query may behave differently
pasted into Prometheus vs Grafana.

## Useful panels & queries

- **RED row** (per service): traffic `sum(rate(http_requests_total[5m]))`, error ratio
  `sum(rate(http_requests_total{code=~"5.."}[5m]))/sum(rate(http_requests_total[5m]))`, p95
  `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))`.
- **Compute resources by namespace** (the bundled dashboard): CPU
  `sum by (namespace) (rate(container_cpu_usage_seconds_total[5m]))`, memory
  `sum by (namespace) (container_memory_working_set_bytes)`.
- **Node USE** (from node-exporter): CPU busy
  `1 - avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m]))`, mem available
  `node_memory_MemAvailable_bytes`, disk free `node_filesystem_avail_bytes`.
- **Pod health**: restarts `sum by (pod) (increase(kube_pod_container_status_restarts_total[15m]))`,
  not-ready `kube_pod_status_ready{condition="false"}`.
- **Saturation**: CPU throttle ratio
  `sum by (pod)(rate(container_cpu_cfs_throttled_periods_total[5m]))/sum by (pod)(rate(container_cpu_cfs_periods_total[5m]))`.

Use a **stat** panel for current-value SLI/error budget, **time series** for trends, **heatmap** for
latency histograms, **table** for top-N.

## Reusing the bundled dashboards

`kube-prometheus-stack` installs dashboards for cluster/namespace/pod compute resources, node-exporter,
the control plane, etc. Browse them (Dashboards → Browse), pick e.g. **Kubernetes / Compute Resources /
Namespace (Pods)**, and switch the namespace variable. You can also import community dashboards from
grafana.com by ID. Edit a panel to learn its query, then build your own from that pattern.

## Grafana alerting

Grafana's unified alerting evaluates a query on a schedule and fires when a condition holds for a
duration — same principle as Prometheus rules, but defined on panels and able to reuse dashboard
queries. Good for dashboard-level/experimental alerts or visual warnings; use Prometheus +
Alertmanager for production-grade, versioned alerts (comparison table in `promql-and-alerting.md`).

To create one in the UI: edit a panel → **Alert** tab → **Create alert rule** → define the condition
(e.g. `WHEN last() OF query(A) IS ABOVE 1000`), evaluation interval, and `for:` duration → pick a
contact point (Slack/email). Grafana can also be **provisioned** declaratively (mounted ConfigMaps/JSON)
so alerts are code, not clicks:

```yaml
apiVersion: 1
groups:
  - orgId: 1
    name: orders-latency
    interval: 30s
    rules:
      - uid: latency_warn
        title: "Orders p95 latency above 1s"
        condition: C
        data:
          - refId: A
            datasourceUid: prometheus
            model:
              expr: histogram_quantile(0.95, sum(rate(orders_request_seconds_bucket[2m])) by (le))
          - refId: C
            type: condition
            conditions:
              - evaluator: { type: gt, params: [1] }
                reducer: { type: last }
        for: 1m
        labels: { severity: warning }
        annotations:
          summary: "p95 latency above 1s"
          runbook_url: "https://git.example.com/runbooks/orders-latency.md"
```

When an alert fires, the panel border changes color — an at-a-glance cue for anyone watching the
dashboard.
