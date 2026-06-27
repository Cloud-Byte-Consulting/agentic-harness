# Tracing with OpenTelemetry

Distributed tracing: instrument with OpenTelemetry, ship OTLP to a Collector, and visualize in Jaeger or
Tempo.

## Contents
- [Traces and spans](#traces-and-spans)
- [OpenTelemetry building blocks](#opentelemetry-building-blocks)
- [Instrumenting an app](#instrumenting-an-app)
- [Deploying the OpenTelemetry Collector](#deploying-the-opentelemetry-collector)
- [Jaeger](#jaeger)
- [Tempo](#tempo)
- [Context propagation](#context-propagation)
- [Sampling](#sampling)
- [OpenTelemetry vs service-mesh tracing](#opentelemetry-vs-service-mesh-tracing)
- [Instrumentation best practices](#instrumentation-best-practices)

## Traces and spans

A **trace** is one request's journey through your system. Each unit of work (a service handling part of
the request) is a **span**, with start/end timing and context. A trace is a tree of spans, so you can
see exactly which service/call contributed the latency. Where metrics say "p95 is high" and logs say
"this errored," traces say **where** in the call graph it happened — the missing third pillar.

## OpenTelemetry building blocks

OpenTelemetry (OTel) is the vendor-neutral standard that merged the old OpenTracing (traces) and
OpenCensus (metrics) projects. Three pieces:
1. **Instrumentation libraries / SDK** — embedded in your app; capture spans (and metrics/logs)
   automatically for common frameworks, or manually via the API.
2. **OpenTelemetry Collector** — receives, processes (batch, filter, enrich), and exports telemetry. It
   decouples your apps from the backend: apps only know the Collector, the Collector knows Jaeger/Tempo/
   Prometheus/etc.
3. **Exporters** — define where data goes. Prefer **OTLP** (the native protocol) for interoperability.

This pipeline means you can swap backends without touching application code.

## Instrumenting an app

Auto-instrumentation captures spans for HTTP frameworks and outgoing calls with almost no code. Python
(Flask) sketch:

```python
from opentelemetry import trace
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor

trace.set_tracer_provider(TracerProvider(
    resource=Resource.create({SERVICE_NAME: "orders-service"})))
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(OTLPSpanExporter(
        endpoint="http://otel-collector.observability.svc:4318/v1/traces")))

app = Flask(__name__)
FlaskInstrumentor().instrument_app(app)   # spans for incoming requests
RequestsInstrumentor().instrument()       # spans for outgoing HTTP calls
```

Node.js uses `@opentelemetry/sdk-node` + `getNodeAutoInstrumentations()` and the OTLP HTTP exporter the
same way. Point every service at the Collector via `OTEL_EXPORTER_OTLP_ENDPOINT` (env var) so the
endpoint is config, not code. Use OTLP/HTTP on `:4318` or OTLP/gRPC on `:4317`.

## Deploying the OpenTelemetry Collector

ConfigMap (receivers → processors → exporters, wired into a `traces` pipeline; here exporting to Jaeger
over OTLP):

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config
  namespace: observability
data:
  config.yaml: |
    receivers:
      otlp:
        protocols:
          http:
          grpc:
    processors:
      batch: {}
    exporters:
      otlp:
        endpoint: jaeger-collector.observability.svc.cluster.local:4317
        tls:
          insecure: true
    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [batch]
          exporters: [otlp]
```

Deployment + Service (exposing both OTLP ports; the Collector also serves its own Prometheus metrics on
`:8888`, which you can scrape):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-collector
  namespace: observability
spec:
  replicas: 1
  selector:
    matchLabels: { app: otel-collector }
  template:
    metadata:
      labels: { app: otel-collector }
    spec:
      containers:
        - name: otel-collector
          image: otel/opentelemetry-collector-contrib:0.103.1
          args: ["--config=/etc/otel/config.yaml"]
          ports:
            - { name: otlp-grpc, containerPort: 4317 }
            - { name: otlp-http, containerPort: 4318 }
          volumeMounts:
            - { name: otel-config, mountPath: /etc/otel }
      volumes:
        - name: otel-config
          configMap: { name: otel-collector-config }
---
apiVersion: v1
kind: Service
metadata:
  name: otel-collector
  namespace: observability
spec:
  selector: { app: otel-collector }
  ports:
    - { name: otlp-grpc, port: 4317, targetPort: 4317 }
    - { name: otlp-http, port: 4318, targetPort: 4318 }
```

The OpenTelemetry Operator can also manage Collectors via an `OpenTelemetryCollector` CR and inject
auto-instrumentation into pods — useful at scale, but the plain Deployment above is the clearest mental
model.

## Jaeger

For learning/small setups, the Jaeger all-in-one image bundles ingest + storage + query/UI. Enable OTLP
ingest so the Collector can export to it:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger
  namespace: observability
spec:
  replicas: 1
  selector:
    matchLabels: { app: jaeger }
  template:
    metadata:
      labels: { app: jaeger }
    spec:
      containers:
        - name: jaeger
          image: jaegertracing/all-in-one:1.58
          args: ["--collector.otlp.enabled=true"]
          ports:
            - { name: ui, containerPort: 16686 }       # web UI
            - { name: otlp-grpc, containerPort: 4317 }  # OTLP ingest
            - { name: otlp-http, containerPort: 4318 }
```

View traces: `kubectl port-forward svc/jaeger-query 16686:16686`, open `http://localhost:16686`, pick
the service and operation, **Find Traces**. In production, run the Collector, query, and storage
(Elasticsearch/Cassandra) separately or use a managed backend — all-in-one keeps no durable storage.
Helm: `helm install jaeger jaegertracing/jaeger`.

## Tempo

**Grafana Tempo** is a high-scale, cost-efficient tracing backend that stores traces in object storage
and indexes minimally (find traces by ID, or via metrics/logs exemplars). It's the natural choice with a
Grafana + Loki + Prometheus stack because you can pivot metric → log → trace in one UI. Point the
Collector's OTLP exporter at Tempo's OTLP endpoint and add Tempo as a Grafana data source (see
`grafana-dashboards.md`). Choose Jaeger for a standalone tracing UI; choose Tempo for Grafana-native,
cheap-at-scale tracing.

## Context propagation

Distributed tracing only works if trace context is **propagated** across service boundaries —
auto-instrumentation injects/extracts W3C `traceparent` headers on HTTP calls so a downstream span joins
the same trace. If you make raw calls or cross a queue, propagate context manually. Put the `trace_id`
into your structured logs (`tracing-opentelemetry.md` ↔ `logging.md`) so a log line links back to its
trace.

## Sampling

You rarely keep 100% of traces at scale. Strategies:
- **Head sampling** — decide at the start (e.g. keep 10%). Simple, cheap, may miss rare errors.
- **Tail sampling** — Collector buffers a whole trace, then keeps it based on outcome (e.g. always keep
  errors and slow traces, sample the rest). Richer, more memory. Configure with the Collector's
  `tail_sampling` processor.

Balance trace volume/cost against the granularity you need for debugging.

## OpenTelemetry vs service-mesh tracing

A service mesh (Linkerd, Istio) can produce traces and service-to-service observability **without
modifying app code** — the sidecar proxies emit spans for every hop. OpenTelemetry gives **in-process**
detail (custom spans, business context) but requires instrumentation. Common split: mesh for
zero-code service-graph/latency visibility (great for legacy apps and packet-loss/bottleneck hunting);
OTel SDK where you want rich, app-level spans and custom attributes. Meshes can export their spans to the
same Jaeger/Tempo backend. (Service-mesh networking itself: see **kubernetes-networking**.)

## Instrumentation best practices

- **Instrument early** — add telemetry during development, not after the first incident; run it in
  staging/load tests too (shift left).
- **Standardize service naming** — consistent `service.name` makes dashboards and traces coherent.
- **Include business context** — record user/transaction IDs on spans where appropriate.
- **Correlate** — link latency spikes (metrics) to trace spans, and spans to logs via `trace_id`.
- **Export OTLP** — maximum interoperability and backend portability.
- **Trace the whole journey** — every hop on critical paths (frontend → backend → third-party APIs),
  or the real bottleneck stays invisible. Partial tracing is the "fragmentation" anti-pattern.
