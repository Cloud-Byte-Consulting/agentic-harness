# Rightsizing Methodology & Layered Autoscaling Patterns

Efficiency = useful work / resources paid for. The two failure modes are **over-provisioning**
(nodes at 30% utilization — paying for 70% idle) and **under-provisioning** (contention, OOM,
throttling). The target is steady ~70–80% utilization with no Pending pods and no idle nodes.
Everything below depends on **correct resource requests** — they drive scheduling, HPA percentages,
VPA, and node-autoscaler capacity math.

## Contents
- Rightsizing methodology
- Establishing default requests/limits
- The end-to-end scaling timeline & choosing metrics
- Cost optimization: Spot, Graviton/ARM, consolidation
- Graceful shutdown (Spot & consolidation safety)
- Scale-to-zero (pods and nodes)

---

## Rightsizing methodology

Rightsizing is **continuous**, not one-off — apps change, so re-measure regularly.

1. **Get visibility.** Install Metrics Server (`kubectl top`) for quick checks, and
   Prometheus/Grafana (e.g. the `Kubernetes / Compute Resources / Workload` dashboard) for
   history. (Deep observability setup → kubernetes-observability skill.)
2. **Measure under representative load.** Run a load test (`ab`, `hey`, or real traffic) for
   several minutes and read actual CPU/memory.
3. **Set requests from data:**
   - **CPU**: request ≈ 95th-percentile usage **+ ~20% headroom** for spikes. Leave **CPU limits
     unset** (let it burst into spare capacity) unless you need noisy-neighbor isolation or
     licensing caps. Remember the scheduler only uses *requests*; CPU limits cause runtime
     throttling even on an idle node.
   - **Memory**: set **requests == limits** for predictable, stable scheduling. Avoid large
     request/limit gaps (node overcommit → OOM). Memory over limit = OOM-killed (incompressible);
     CPU over limit = throttled (compressible).
   - "Efficiency %" = used / requested. ~80% is a good target; tune without hurting performance.
4. **Automate where useful.** **KRR** (Robusta Kubernetes Resource Recommender) reads Prometheus
   history and suggests CPU/memory requests/limits as a CLI (no in-cluster agent). VPA in `Off`
   mode gives recommendations too. Paid tools (Kubecost, etc.) can rightsize continuously.

Example workflow:

```bash
kubectl top pods -n NS                  # current usage
# load test, then read Grafana / kubectl top for sustained usage, then patch requests:
kubectl patch deployment montecarlo-pi --patch \
  '{"spec":{"template":{"spec":{"containers":[{"name":"montecarlo-pi","resources":{"requests":{"cpu":"1200m"}}}]}}}}'
```

---

## Establishing default requests/limits

When teams won't set requests, enforce safe defaults so the scheduler/autoscalers still behave:

- **LimitRange** (per-namespace, applies to **new** pods/containers): defaults + min/max bounds.

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: cpu-memory-defaults
  namespace: team-a
spec:
  limits:
  - type: Container
    defaultRequest: {cpu: 900m, memory: 512Mi}   # applied if the container omits requests
    default:        {cpu: 1200m}                  # default limits
```

- **ResourceQuota** caps the *aggregate* CPU/memory of all pods in a namespace (LimitRange is
  per-pod; ResourceQuota is per-namespace total).
- **Mutating webhook** (Kyverno / OPA Gatekeeper) for richer policy than LimitRange.

LimitRange is a guardrail, not a substitute for measured per-container requests.

---

## End-to-end scaling timeline & choosing metrics

Realistic latency from "load rises" to "new capacity serving" with KEDA→HPA→Karpenter:
Prometheus scrape (~15s) + KEDA poll (~30s) + HPA decision (~15s) + node launch (~40s) + pod
start/health/LB registration (~30–60s) ≈ **90–150s**. Implications:
- By the time **CPU%** spikes enough to trigger, users may already feel it (CPU% can be misleading —
  includes memory-stall cycles). Scale on **leading** signals: requests/sec, queue depth, p95
  latency.
- HPA/KEDA take the **max** across multiple metrics — use a primary business metric (RPS, queue
  depth) plus a CPU/latency **safety net**. Beware thresholds that cause over-scaling.
- Tune `behavior` to scale up fast / down slow (see `hpa.md`, `keda.md`).
- Keep node bootstrap time short (lean AMI/agents) so node-launch latency is minimal; consider
  overprovisioning placeholder pods (see `topology-spread-and-pdb.md`).

There is no single "best" metric — model what actually degrades *your* app (RPS+connections for a
web API; queue depth+processing time for a pipeline; tokens/sec for ML inference) and balance
cost, latency SLA, and reliability.

---

## Cost optimization: Spot, Graviton/ARM, consolidation

Typical optimization path (Karpenter implements each via consolidation in real time;
`eks-node-viewer` visualizes cost):
1. **Rightsize** over-provisioned pods (above) — fewer/smaller nodes.
2. **ARM / Graviton** — allow `kubernetes.io/arch: ["amd64","arm64"]` in the NodePool *and* build
   multi-arch images; Graviton is cheaper per-performance for many workloads.
3. **Spot** — add `karpenter.sh/capacity-type: ["spot"]` (or `["on-demand","spot"]`). Big discounts
   for interruptible, fault-tolerant workloads. "Spot is the reward for a good architecture."
4. **Consolidation** — `WhenEmptyOrUnderutilized` keeps the fleet packed; Karpenter downsizes and
   does spot-to-spot replacement automatically.

Use NodePool `weight` for "prefer spot, fall back to on-demand" (or "reserved first"). Cap spend
with NodePool `limits`.

---

## Graceful shutdown (Spot & consolidation safety)

Any node removal (Spot interruption ~2-min warning, consolidation, drift, expiration) evicts pods.
Make workloads tolerate it with Kubernetes primitives — no app code change required:

```yaml
spec:
  terminationGracePeriodSeconds: 90        # > preStop sleep, < 120s Spot window
  containers:
  - name: app
    lifecycle:
      preStop:
        exec:
          command: ["/bin/sh","-c","sleep 75"]   # drain in-flight work, flush, close conns
    readinessProbe:                         # fails on SIGTERM → LB deregisters before exit
      httpGet: {path: /healthz, port: 8080}
      periodSeconds: 10
      failureThreshold: 3
```

Eviction sequence: pod removed from Service endpoints → SIGTERM + `preStop` runs → grace timer →
SIGKILL if still alive. A readiness probe that flips to `503` on SIGTERM is what makes **external**
load balancers (ALB/NLB) stop routing promptly — without it you get 5xx during shutdown. For
deeper control, handle SIGTERM in app code (mark `/ready` unhealthy, finish in-flight work, then
exit). Karpenter pre-launches the replacement node as soon as it sees the interruption, so keep
bootstrap fast.

---

## Scale-to-zero

**Pods** (KEDA, see `keda.md`): `minReplicaCount: 0` for queue/event/cron workloads; keep ≥1 for
synchronous HTTP. Cron scaler turns dev/staging off out of hours.

**Nodes** (Karpenter): nodes go to zero automatically when empty (consolidation). For a *guaranteed*
off-hours shutdown of non-prod, a robust pattern uses two CronJobs (RBAC-scoped) + a ConfigMap
listing target NodePools:
- **Scale-down (e.g. weekdays 17:00):** for each listed NodePool, annotate it with its current
  `limits`, set `limits: {cpu: 0}` (so Karpenter launches nothing), then delete its NodeClaims
  (the finalizer terminates the instances) → zero nodes.
- **Scale-up (e.g. weekdays 09:00):** restore `limits` from the annotation → Karpenter relaunches
  capacity for the now-Pending pods.

Run the jobs manually for incidents via `kubectl create job --from=cronjob/...`. If using GitOps
(Argo CD / Flux), exclude these NodePools from reconciliation so the temporary `limits`/annotation
changes aren't reverted.
