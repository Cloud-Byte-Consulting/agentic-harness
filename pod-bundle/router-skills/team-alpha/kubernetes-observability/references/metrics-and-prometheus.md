# Metrics & Prometheus

How Kubernetes exposes metrics, how Prometheus collects them, the exporters that matter, and how to
wire your own apps in via the Prometheus Operator.

## Contents
- [Where Kubernetes metrics come from](#where-kubernetes-metrics-come-from)
- [Metrics Server vs Prometheus](#metrics-server-vs-prometheus)
- [The Prometheus exposition format & metric types](#the-prometheus-exposition-format--metric-types)
- [The exporters you actually run](#the-exporters-you-actually-run)
- [Scraping your app: ServiceMonitor & PodMonitor](#scraping-your-app-servicemonitor--podmonitor)
- [Operator RBAC for cross-namespace scraping](#operator-rbac-for-cross-namespace-scraping)
- [Securing the /metrics endpoint](#securing-the-metrics-endpoint)
- [Retention, storage & remote-write](#retention-storage--remote-write)
- [Troubleshooting a scrape that isn't working](#troubleshooting-a-scrape-that-isnt-working)

## Where Kubernetes metrics come from

The API server exposes a `/metrics` URI, but it requires auth — never make it anonymous (escalation
risk). To read it manually:

```bash
kubectl create sa getmetrics
kubectl create clusterrole get-metrics --non-resource-url=/metrics --verb=get
kubectl create clusterrolebinding get-metrics --clusterrole=get-metrics \
  --serviceaccount=default:getmetrics
TOKEN=$(kubectl create token getmetrics -n default)
curl -sk -H "Authorization: Bearer $TOKEN" https://<apiserver>:6443/metrics | head
```

In practice you never scrape these by hand — Prometheus and the exporters do it. The key insight: most
useful cluster metrics (`kube_pod_info`, `kube_deployment_status_replicas`, …) do **not** come from the
API server's own `/metrics`. They come from **kube-state-metrics**.

## Metrics Server vs Prometheus

| | Metrics Server | Prometheus |
|---|---|---|
| Purpose | Feed `kubectl top` + HPA/VPA | Monitoring, alerting, debugging |
| Data | Current CPU/mem only | Any metric, full history |
| Storage | In-memory, no history | On-disk TSDB |
| API | `metrics.k8s.io` (Metrics API) | PromQL / HTTP API |
| Source | kubelet/cAdvisor, ~15s | Scrapes all `/metrics` endpoints |

Run **both**. Metrics Server is a focused add-on for autoscaling; Prometheus is your observability
backend. Installing Metrics Server (only if your distro doesn't ship it):

```bash
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm install metrics-server metrics-server/metrics-server -n kube-system
# Local/self-hosted clusters with self-signed kubelet certs (NOT production):
#   --set args="{--kubelet-insecure-tls}"
```

If `kubectl top` fails with a TLS error on kubeadm clusters, the node IPs aren't in the kubelet serving
cert SAN. Fix by enabling `serverTLSBootstrap: true` in the kubeadm config (`kubectl edit cm -n
kube-system kubeadm-config`) and in each node's `/var/lib/kubelet/config.yaml`, then approve the CSRs.
(Metrics Server's role in HPA/VPA: see **kubernetes-autoscaling-scheduling**.)

## The Prometheus exposition format & metric types

Endpoints serve plain text; `#` lines are metadata:

```
# HELP kube_pod_info Information about pod.
# TYPE kube_pod_info gauge
kube_pod_info{namespace="prod",pod="api-7d9",node="worker-1"} 1
```

Form: `metric_name{label1="v1",label2="v2"} value`. Labels are how you filter and aggregate.

Four types:
- **Counter** — only increases (or resets to 0 on restart). E.g. `http_requests_total`. Always query
  with `rate()`/`increase()`, never raw.
- **Gauge** — goes up and down. E.g. `node_memory_MemAvailable_bytes`, open connections.
- **Histogram** — pre-bucketed observations, emits `_bucket{le="..."}`, `_sum`, `_count`. Compute
  quantiles with `histogram_quantile()`. Use for latency/size distributions — far cheaper than a
  sample per request.
- **Summary** — client-side quantiles. Quantiles **cannot be aggregated** across instances. Prefer
  histograms unless you specifically need a client-computed exact quantile.

## The exporters you actually run

- **node-exporter** — DaemonSet; per-node OS/hardware metrics (`node_cpu_seconds_total`,
  `node_memory_*`, `node_filesystem_*`, disk/network). The USE-method source.
- **kube-state-metrics** — Deployment; turns API object state into metrics (`kube_pod_info`,
  `kube_deployment_status_replicas_available`, `kube_pod_container_status_restarts_total`,
  `kube_pod_status_phase`). Object health, not resource usage.
- **kubelet / cAdvisor** — built into the kubelet; per-container usage
  (`container_cpu_usage_seconds_total`, `container_memory_working_set_bytes`).
- **API server / etcd / scheduler / controller-manager** — control-plane health
  (`apiserver_request_total`, etcd latency). Monitor these on the control plane.

`kube-prometheus-stack` deploys node-exporter and kube-state-metrics for you and wires their
ServiceMonitors automatically.

## Scraping your app: ServiceMonitor & PodMonitor

Expose `/metrics` in your app, put a Service in front of the pods, then declare a ServiceMonitor. The
operator generates scrape config from it — pods are scraped on their **Pod IP** (resolved via the
Service's Endpoints), not via the Service VIP.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: myapp
  namespace: monitoring
  labels:
    release: kube-prometheus-stack   # MUST match the operator's serviceMonitorSelector
spec:
  namespaceSelector:
    matchNames: ["myapp-ns"]          # where the target Service lives
  selector:
    matchLabels:
      app: myapp                      # selects the Service (and thus its pods)
  endpoints:
    - port: http-metrics              # named port on the Service
      path: /metrics
      interval: 30s
      # For an HTTPS endpoint protected by the pod's own SA token:
      # scheme: https
      # bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
      # tlsConfig: { insecureSkipVerify: true }
```

The corresponding Service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp
  namespace: myapp-ns
  labels:
    app: myapp
spec:
  selector:
    app: myapp
  ports:
    - name: http-metrics              # the name the ServiceMonitor references
      port: 8080
      targetPort: 8080
```

**PodMonitor** is the same idea without a Service — scrape pods directly by pod label. Use it for pods
that have no Service:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: myapp-pods
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  namespaceSelector:
    matchNames: ["myapp-ns"]
  selector:
    matchLabels:
      app: myapp
  podMetricsEndpoints:
    - port: metrics
      interval: 30s
```

The `release: kube-prometheus-stack` label is critical: Prometheus is not multi-tenant, so the operator
only adopts monitors whose labels match its `serviceMonitorSelector`/`podMonitorSelector`. Check what
your Prometheus expects:

```bash
kubectl get prometheus -n monitoring -o jsonpath='{.items[0].spec.serviceMonitorSelector}'
```

## Operator RBAC for cross-namespace scraping

To scrape a Service in another namespace, Prometheus's ServiceAccount needs read access to
endpoints/pods/services there, or the operator can't resolve targets (it fails quietly):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: prometheus-discovery
  namespace: myapp-ns
rules:
  - apiGroups: [""]
    resources: ["services", "endpoints", "pods"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: prometheus-discovery
  namespace: myapp-ns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: prometheus-discovery
subjects:
  - kind: ServiceAccount
    name: kube-prometheus-stack-prometheus   # the Prometheus pod's SA
    namespace: monitoring
```

(General RBAC patterns: see **kubernetes-security-rbac**.)

## Securing the /metrics endpoint

Metrics leak structure: workload counts, traffic patterns, when nodes are busiest — "attackers think in
graphs." Protect app metrics endpoints with token validation: the Prometheus pod has a Kubernetes
identity, so configure the ServiceMonitor's `bearerTokenFile` to send the pod's SA token and have the
app validate it (e.g. via `SubjectAccessReview`). Add NetworkPolicies to restrict who can reach
`/metrics`. (NetworkPolicy: see **kubernetes-networking**.)

Prometheus and Alertmanager themselves have **no authentication**. Either keep them reachable only via
`kubectl port-forward`, or front them with an authenticating reverse proxy (oauth2-proxy, OpenUnison).
Grafana has a user model but ships a known default admin password — change it.

## Retention, storage & remote-write

Prometheus stores its TSDB locally (hence a StatefulSet + PVC). Configure retention on the Prometheus
CR:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: kube-prometheus-stack-prometheus
  namespace: monitoring
spec:
  retention: 15d              # delete samples older than this
  retentionSize: 50GB         # or cap by size (whichever hits first)
  storage:
    volumeClaimTemplate:
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 100Gi
  # Ship to long-term/global storage (Thanos, Mimir, Cortex, or a vendor):
  remoteWrite:
    - url: https://remote-store.example.com/api/v1/write
```

Local Prometheus is single-cluster and short-retention by design. For long retention, global query
across clusters, or HA, use **remote-write** to Thanos/Mimir/Cortex or a managed backend (Grafana
Cloud, Datadog, Amazon Managed Prometheus — all speak the Prometheus format). Storage sizing details
and PVC provisioning: see **kubernetes-storage**.

## Troubleshooting a scrape that isn't working

Metrics not showing up after a few minutes — check in this order:
1. **Operator logs** — did it fail to load the Service/Endpoint/Pod? (RBAC, selector mismatch)
   `kubectl logs -n monitoring deploy/kube-prometheus-stack-operator`
2. **config-reloader sidecar** in the Prometheus pod — did the new config load?
   `kubectl logs -n monitoring prometheus-...-0 -c config-reloader`
3. **Prometheus container** — is the target listed and what's its error?
   `kubectl logs -n monitoring prometheus-...-0 -c prometheus`
   then **Status → Target health** in the UI for the per-target scrape error.

Most common causes: missing `release:` label on the ServiceMonitor; missing operator RBAC in the target
namespace; wrong port name; app not actually serving `/metrics`; `/metrics` only served on one hostname
while scrape hits the Pod IP.
