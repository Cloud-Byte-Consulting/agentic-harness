# kube-prometheus-stack

Installing and operating the standard Helm-based monitoring stack: what it bundles, what to tune, how to
verify it, and how to access the UIs.

## Contents
- [What it is](#what-it-is)
- [Install](#install)
- [What gets deployed](#what-gets-deployed)
- [Common values to tune](#common-values-to-tune)
- [Verifying the install](#verifying-the-install)
- [Accessing the UIs](#accessing-the-uis)
- [Adding your own monitors, rules, dashboards](#adding-your-own-monitors-rules-dashboards)
- [Upgrades & the curated rules](#upgrades--the-curated-rules)

## What it is

`kube-prometheus-stack` (from the `prometheus-community` Helm repo) is the de-facto way to deploy a
complete monitoring stack on Kubernetes. It packages the **Prometheus Operator** plus a working,
opinionated set of components, CRDs, ServiceMonitors, alert rules, and Grafana dashboards. Start here
rather than assembling Prometheus/Grafana/Alertmanager by hand.

## Install

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace
# pin a version in real use: --version <x.y.z>
```

Idempotent upgrade form:

```bash
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  -f values.yaml
```

## What gets deployed

- **Prometheus Operator** — reconciles CRDs into running config.
- **Prometheus** — StatefulSet with local TSDB (it's a StatefulSet because it persists data).
- **Alertmanager** — StatefulSet; routing/grouping/silences.
- **Grafana** — pre-wired to Prometheus, with curated dashboards.
- **node-exporter** — DaemonSet; per-node OS metrics.
- **kube-state-metrics** — Deployment; Kubernetes object-state metrics.
- **CRDs** — `Prometheus`, `Alertmanager`, `ServiceMonitor`, `PodMonitor`, `PrometheusRule`,
  `AlertmanagerConfig`, `ThanosRuler`, `Probe`.
- **~40 PrometheusRules** and ServiceMonitors for the control plane, kubelet, nodes, and the stack
  itself.

The bundled rules are maintained upstream and updated with experience — **don't edit them in place**;
add your own PrometheusRule objects instead.

## Common values to tune

```yaml
# values.yaml
prometheus:
  prometheusSpec:
    retention: 15d
    retentionSize: 50GB
    # Adopt monitors/rules even without the release label (use cautiously):
    # serviceMonitorSelectorNilUsesHelmValues: false
    # ruleSelectorNilUsesHelmValues: false
    storageSpec:
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          resources: { requests: { storage: 100Gi } }
    resources:
      requests: { cpu: 500m, memory: 2Gi }
      limits:   { memory: 4Gi }
    # remoteWrite:
    #   - url: https://thanos-receive.example.com/api/v1/receive

alertmanager:
  alertmanagerSpec:
    storage:
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          resources: { requests: { storage: 10Gi } }

grafana:
  adminPassword: "<set-a-strong-one>"     # default is 'prom-operator' — change it
  defaultDashboardsEnabled: true
  persistence:
    enabled: true
    size: 10Gi
```

Key knobs: **retention/storage** (Prometheus is short-retention single-cluster by default — use
`remoteWrite` to Thanos/Mimir/Cortex/managed for long-term/global; see `metrics-and-prometheus.md`),
**resources** (Prometheus is memory-hungry; size for your series count), **selectors** (by default it
only adopts monitors/rules carrying the `release: kube-prometheus-stack` label — the
`*SelectorNilUsesHelmValues: false` flags relax that, but explicit labels are cleaner), and **Grafana
admin password**.

## Verifying the install

```bash
kubectl get pods -n monitoring
kubectl get servicemonitor,prometheusrule -n monitoring
kubectl top nodes                       # confirms Metrics Server too (separate component)
```

Then in the Prometheus UI, **Status → Target health**: every intended target should be `UP`. Some
alerts may fire by default on local clusters (KinD/minikube) due to networking quirks — that's expected,
not a real problem.

## Accessing the UIs

Neither Prometheus nor Alertmanager has auth — reach them via port-forward, or front them with an auth
proxy for shared access (see `metrics-and-prometheus.md` → securing).

```bash
kubectl -n monitoring port-forward svc/kube-prometheus-stack-prometheus 9090:9090
kubectl -n monitoring port-forward svc/kube-prometheus-stack-alertmanager 9093:9093
kubectl -n monitoring port-forward svc/kube-prometheus-stack-grafana 3000:80
```

- Prometheus `http://localhost:9090` — Graph, Alerts, Status→Targets/Rules.
- Alertmanager `http://localhost:9093` — active alerts, silences.
- Grafana `http://localhost:3000` — login `admin` / your configured password.

## Adding your own monitors, rules, dashboards

Everything you add is a labeled object the operator adopts (full examples in the topic references):
- **ServiceMonitor / PodMonitor** with `release: kube-prometheus-stack` → `metrics-and-prometheus.md`.
- **PrometheusRule** (alerts + recording rules) with the same label → `promql-and-alerting.md`.
- **Grafana dashboard** as a ConfigMap labeled `grafana_dashboard: "1"` → `grafana-dashboards.md`.

Verify adoption: the new ServiceMonitor's targets appear in Status→Targets; the new alerts appear at
`/rules` and in the Alerts tab.

## Upgrades & the curated rules

`helm upgrade` bumps component images and the bundled rules/dashboards. Because CRDs are **not** upgraded
by `helm upgrade`, apply the chart's CRD updates manually when the release notes call for it. Keep your
customizations in your own `values.yaml` and your own PrometheusRule/dashboard objects — never patch the
bundled rules, so upgrades stay clean. (Managing this stack declaratively via GitOps/Argo CD: see
**kubernetes-gitops-cicd**.)
