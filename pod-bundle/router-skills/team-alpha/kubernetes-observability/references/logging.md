# Logging

Cluster logging architecture, the collector agents, backend choices (EFK/OpenSearch vs Loki),
structured logging, and retention.

## Contents
- [How container logs work](#how-container-logs-work)
- [Cluster logging architecture](#cluster-logging-architecture)
- [Fluent Bit vs Fluentd](#fluent-bit-vs-fluentd)
- [A Fluent Bit DaemonSet](#a-fluent-bit-daemonset)
- [Backend choice: EFK/OpenSearch vs Loki](#backend-choice-efkopensearch-vs-loki)
- [Loki + Promtail](#loki--promtail)
- [Structured logging & correlation](#structured-logging--correlation)
- [Retention & smart logging](#retention--smart-logging)
- [Querying logs](#querying-logs)
- [Node-level log rotation (containerd/Docker)](#node-level-log-rotation-containerddocker)

## How container logs work

Containers should write to **stdout/stderr**, not to files. The container runtime captures those
streams and writes them as JSON files on the node, conventionally under
`/var/log/containers/*.log` (symlinks into `/var/log/pods/...`, backed by the runtime's storage). The
kubelet exposes them through the API, which is what `kubectl logs` reads:

```bash
kubectl logs <pod> -n <ns> -c <container>
kubectl logs <pod> -n <ns> --previous          # the crashed/previous instance
kubectl logs -f deploy/<name> -n <ns>          # follow a Deployment's pods
```

This is simple but limited: no rotation guarantees, logs vanish when pods are deleted, and you can't
search across pods. Hence centralization.

## Cluster logging architecture

The standard **node-agent** pattern: a **DaemonSet** runs one collector pod per node; it tails every
container log file on that node, enriches each line with Kubernetes metadata (namespace, pod, labels,
container), and ships it to a central backend. Two other patterns exist:
- **Application forwarding** — the app sends logs directly to the backend (more coupling).
- **Sidecar** — a per-pod sidecar container ships that pod's logs (more overhead, useful for apps that
  insist on writing to files).

The node-agent DaemonSet is the default for cluster-wide logging. Its pods are **privileged**: they
`hostPath`-mount `/var/log` and run with a permissive securityContext. Protect the logging namespace —
restrict who can deploy there and use admission policy (Pod Security Standards / Kyverno / Gatekeeper)
to limit what runs there. (Pod Security: see **kubernetes-security-rbac**.)

## Fluent Bit vs Fluentd

Both are CNCF log processors; pick based on weight vs power:
- **Fluent Bit** — small, fast, low memory (C). Tails logs, adds metadata, ships. The default choice
  for most clusters and the lighter half of the "EFK" stack.
- **Fluentd** — heavier (Ruby), far richer parsing/transformation/routing and a huge plugin ecosystem.
  Use when you need complex pipeline processing.

(Logstash is the original ELK shipper but is heavyweight; Fluent Bit/Fluentd are the common Kubernetes
choices and are compatible with Elasticsearch/OpenSearch.)

## A Fluent Bit DaemonSet

Concept (the `fluent/fluent-bit` Helm chart wires this for you, including RBAC and tolerations):

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluent-bit
  namespace: logging
spec:
  selector:
    matchLabels: { app: fluent-bit }
  template:
    metadata:
      labels: { app: fluent-bit }
    spec:
      serviceAccountName: fluent-bit       # needs RBAC to read pod metadata
      tolerations:
        - operator: Exists                 # run on every node incl. control plane/tainted
      containers:
        - name: fluent-bit
          image: fluent/fluent-bit:3.1
          volumeMounts:
            - { name: varlog, mountPath: /var/log, readOnly: true }
            - { name: config, mountPath: /fluent-bit/etc/ }
      volumes:
        - { name: varlog, hostPath: { path: /var/log } }
        - { name: config, configMap: { name: fluent-bit-config } }
```

Config tails the container logs, applies the `kubernetes` filter (adds namespace/pod/labels), and
outputs to a backend:

```ini
[INPUT]
    Name              tail
    Path              /var/log/containers/*.log
    Parser            cri
    Tag               kube.*

[FILTER]
    Name              kubernetes
    Match             kube.*
    Merge_Log         On

[OUTPUT]
    Name              loki                 # or 'es' for Elasticsearch/OpenSearch
    Match             *
    Host              loki-gateway.monitoring.svc
    Labels            job=fluent-bit
```

Prefer the chart over hand-rolled DaemonSets — it keeps the RBAC, tolerations, parsers, and OOM-safe
resource limits correct.

## Backend choice: EFK/OpenSearch vs Loki

| | EFK / OpenSearch | Loki (Grafana) |
|---|---|---|
| Indexes | Full-text on log content | Labels only; content not indexed |
| Storage cost | Higher (heavy indices) | Lower (object storage, chunked) |
| Query | Rich full-text (Lucene/PPL/KQL) | LogQL, label-first then grep |
| UI | Kibana / OpenSearch Dashboards | Grafana (unified with metrics) |
| Best for | Deep ad-hoc search, compliance | Cost-efficient, Grafana-native, correlate with metrics |

**OpenSearch** is the open-source fork of Elasticsearch 7.0 (they've since diverged); it ships Kibana-
style Dashboards and supports OIDC SSO out of the box. Components: **masters** (index the data),
**nodes** (host the query/ingest API), **Dashboards** (UI), and a **Fluent Bit/Fluentd DaemonSet** to
feed it. Secure it with OIDC (via OpenUnison/Dex/Keycloak) rather than the single hard-coded admin
password — log data is sensitive.

**Loki** indexes only labels (like Prometheus) and stores compressed log chunks in object storage,
making it cheap at scale and native to Grafana — you can pivot from a metric spike straight to the logs
on the same dashboard. Choose Loki for cost + Grafana integration; choose OpenSearch/EFK when you need
heavy full-text search or compliance-grade retention.

## Loki + Promtail

A minimal Loki path uses Promtail (or Fluent Bit's Loki output) as the agent:

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm install loki grafana/loki-stack \
  --namespace monitoring \
  --set promtail.enabled=true \
  --set loki.persistence.enabled=true
```

Add Loki as a Grafana data source (see `grafana-dashboards.md`) and query with LogQL:

```logql
{namespace="prod", app="myapp"} |= "ERROR"
{namespace="prod"} | json | level="error" | line_format "{{.message}}"
sum by (level) (rate({app="myapp"}[5m]))     # log-rate metrics from logs
```

## Structured logging & correlation

Emit logs as **JSON** with consistent fields so the pipeline can parse and filter them — don't make
operators regex free-form text. Include correlation IDs to stitch a request across services:

```json
{"level":"error","ts":"2026-06-12T10:00:00Z","service":"payments",
 "trace_id":"abc123","user_id":"u-42","msg":"payment failed","code":"PAY001"}
```

`trace_id` lets you jump from a log line to the distributed trace (see `tracing-opentelemetry.md`);
`user_id`/`request_id`/`transaction_id` let you follow one user's journey across the user, product,
payment, and delivery services. Add identifying fields (e.g. `service.name`) at the collector when the
runtime metadata is unavailable (e.g. Docker Desktop on macOS/Windows, where `/var/lib/docker/
containers` isn't host-accessible and apps instead write to a shared volume the agent tails).

## Retention & smart logging

"Capture everything" is an anti-pattern: it buries signal and burns storage. Be deliberate:
- Set **log levels** appropriately — DEBUG in dev, INFO+ in prod; don't ship a flood of OK/heartbeat
  lines.
- Keep **actionable** logs (WARN/ERROR, business-critical events); treat noise as cost.
- Set **retention** per backend deliberately (e.g. 7–30 days hot, archive cold for compliance). In
  OpenSearch use Index State Management; in Loki set `retention_period` and use object-storage
  lifecycle rules.
- Regularly prune what you collect — audit and retire log streams nobody queries.

## Querying logs

- **kubectl** (ad hoc, single source): `kubectl logs`, `--previous`, `-f`, `--since=1h`,
  `--tail=100`, `-l app=myapp` (across a label).
- **OpenSearch/Kibana**: create an index pattern (e.g. `logstash-*` / `filebeat-*`), pick `@timestamp`,
  then Discover; filter by `kubernetes.namespace_name`, `container.name`, `log.level`. PPL example:
  `source = logstash-* | where kubernetes.namespace_name="ingress-nginx" | fields log`.
- **Loki/Grafana**: LogQL as above, often started from a metric panel.

## Node-level log rotation (containerd/Docker)

If you operate the node runtime directly, cap on-disk log growth. For Docker, set the default driver
and rotation in `/etc/docker/daemon.json`:

```json
{
  "log-driver": "json-file",
  "log-opts": { "max-size": "10m", "max-file": "5" }
}
```

`max-size` rotates each file at that size; `max-file` caps how many to keep. containerd/CRI clusters set
the kubelet's `containerLogMaxSize`/`containerLogMaxFiles` instead. Either way, centralize first — once
logs are shipped, node retention can be short.
