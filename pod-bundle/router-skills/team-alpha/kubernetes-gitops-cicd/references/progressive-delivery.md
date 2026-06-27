# Progressive delivery (Argo Rollouts & Flagger)

Progressive delivery shifts traffic to a new version gradually, gates promotion on metrics, and auto-rolls-back on
failure — turning a deploy into a measured experiment. Two controllers dominate on Kubernetes: **Argo Rollouts**
(a `Rollout` CRD replacing `Deployment`) and **Flagger** (drives canary on top of your existing Deployment + a
service mesh/ingress). Read this to pick a strategy, wire metric gates, and configure rollback.

## Strategy selection

| Strategy | How it works | Rollback | Cost | Use when |
|---|---|---|---|---|
| **Rolling** (default Deployment) | Replace pods N at a time in place | Roll forward / `kubectl rollout undo` | 1× | Low-risk changes, no special tooling |
| **Blue-green** | Two full envs; cut traffic over at once after testing | Instant (flip back) | 2× | You need instant cutover + instant rollback and can pay for double capacity |
| **Canary** | Route a small % to new version, ramp up while watching metrics | Gradual / automatic on bad metrics | ~1×+ | Best blast-radius control; want metric-gated, incremental rollout |

Related techniques: **A/B testing** (route by header/cookie to segments), **feature toggles/flags** (ship code
dark, enable at runtime — decouples deploy from release), **dark launches** (send shadow traffic). These compose
with the above.

## Argo Rollouts — canary with analysis

A `Rollout` is a drop-in for `Deployment` that adds a `strategy.canary` (or `blueGreen`) with steps and analysis:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: api
spec:
  replicas: 5
  selector:
    matchLabels: { app: api }
  template:                       # identical to a Deployment pod template
    metadata: { labels: { app: api } }
    spec:
      containers:
        - name: api
          image: registry.example.com/api:1.4.0
  strategy:
    canary:
      canaryService: api-canary    # services the controller manages for traffic split
      stableService: api-stable
      steps:
        - setWeight: 20
        - pause: { duration: 2m }
        - analysis:                # run metric checks; abort+rollback if they fail
            templates:
              - templateName: success-rate
        - setWeight: 50
        - pause: { duration: 5m }
        - setWeight: 100
```

**AnalysisTemplate** queries a metrics provider (Prometheus shown) and fails the rollout if the value breaches the
condition — the controller then auto-aborts and reverts to the stable version:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
spec:
  metrics:
    - name: success-rate
      interval: 1m
      successCondition: result[0] >= 0.95
      failureLimit: 3
      provider:
        prometheus:
          address: http://prometheus.monitoring:9090
          query: |
            sum(rate(http_requests_total{app="api",status!~"5.."}[2m]))
            / sum(rate(http_requests_total{app="api"}[2m]))
```

Blue-green is `strategy.blueGreen` with `activeService` / `previewService` and `autoPromotion`/`prePromotionAnalysis`.
Drive rollouts with `kubectl argo rollouts get rollout api --watch`, `… promote api`, `… abort api`. The
controller integrates with ingress/mesh (NGINX, Istio, SMI, Gateway API) for the actual traffic split.

## Flagger — canary on your existing Deployment

Flagger watches a normal `Deployment` and a `Canary` CR; it creates the canary/primary services, ramps traffic
via the mesh/ingress, and runs metric checks:

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: api
  namespace: prod
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api
  service:
    port: 80
  analysis:
    interval: 1m
    threshold: 5            # consecutive failed checks before rollback
    maxWeight: 50
    stepWeight: 10          # +10% traffic each interval
    metrics:
      - name: request-success-rate
        thresholdRange: { min: 99 }
        interval: 1m
      - name: request-duration
        thresholdRange: { max: 500 }   # ms (p99 latency)
        interval: 1m
    webhooks:
      - name: load-test
        url: http://flagger-loadtester.test/
        metadata: { cmd: "hey -z 1m -q 10 http://api-canary.prod/" }
```

When you push a new image to the Deployment, Flagger detects it, ramps `stepWeight` until `maxWeight` while
metrics pass, then promotes to primary; if `threshold` checks fail it rolls back automatically.

## GitOps fit

Progressive-delivery CRs are just manifests — commit the `Rollout`/`Canary` + `AnalysisTemplate` to the config
repo and let Argo CD/Flux apply them. CI pushes the new image tag (into Git or the Deployment spec); the rollout
controller then performs the metered, metric-gated promotion inside the cluster. This keeps the safety logic
declarative and version-controlled. Metrics setup (Prometheus/Grafana) belongs to `kubernetes-observability` —
this skill only consumes the metrics endpoint in analysis queries.

## Pitfalls

- **No analysis = just a slow rolling update.** The value is in the metric gate; without it you only delay the
  bad version reaching 100%.
- **Metrics that don't reflect user pain** (e.g. only CPU) won't catch a broken release. Gate on success rate +
  latency + error budget.
- **Forgetting the mesh/ingress integration.** Traffic split needs NGINX/Istio/SMI/Gateway API; a bare cluster
  can't shift weights.
- **Blue-green doubles cost** — fine briefly, but don't leave the idle stack running indefinitely.
- **Stateful apps / DB migrations** complicate canary (two versions hitting one schema). Use expand-contract
  migrations and PreSync hooks; don't naively canary a breaking schema change.
