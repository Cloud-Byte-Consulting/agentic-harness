---
name: kubernetes-autoscaling-scheduling
description: >-
  Scale Kubernetes workloads and nodes and control pod placement. Use for
  HorizontalPodAutoscaler (autoscaling/v2: resource, custom, and external metrics, scaling
  behavior and stabilization), VerticalPodAutoscaler modes, KEDA event-driven scaling and
  scale-to-zero, the Metrics Server, node autoscaling with Cluster Autoscaler and Karpenter
  (NodePools, consolidation, disruption), rightsizing, and scheduling controls: nodeSelector,
  node and pod affinity and anti-affinity, taints and tolerations, topologySpreadConstraints,
  PodDisruptionBudgets, PriorityClasses and preemption, and the descheduler. Trigger whenever
  the user wants an app to scale on CPU/memory/queue depth or events, sees pods stuck Pending
  or nodes not scaling, tunes HPA/VPA/KEDA, adds or right-sizes nodes, spreads replicas across
  zones, pins or repels pods from nodes, or protects availability during drains - even without
  saying Kubernetes. For the workload objects themselves see kubernetes-workloads; for
  Prometheus depth see kubernetes-observability.
---

# Kubernetes Autoscaling & Scheduling

This skill equips Claude to design and debug the full Kubernetes autoscaling stack — pod-level
(HPA, VPA, KEDA) and node-level (Cluster Autoscaler, Karpenter) — and the scheduling controls
(affinity, taints, topology spread, PDBs, priorities) that decide *where* pods land and *whether*
nodes can be reclaimed. The goal is an efficient cluster: high utilization, no Pending pods, no
idle nodes, predictable behavior under load.

## When to use this skill

- Authoring or reviewing an `HorizontalPodAutoscaler`, `VerticalPodAutoscaler`, KEDA
  `ScaledObject`/`ScaledJob`, Karpenter `NodePool`/`EC2NodeClass`, or `PodDisruptionBudget`.
- Choosing a scaling strategy: HPA vs VPA vs KEDA; Cluster Autoscaler vs Karpenter.
- Setting resource `requests`/`limits` or rightsizing an over/under-provisioned workload.
- Debugging: pods stuck `Pending`, HPA shows `<unknown>` targets, nodes never scale down,
  Karpenter not launching/consolidating nodes, flapping replicas, eviction loops.
- Scheduling work: pin pods to nodes, spread across zones, isolate workloads (GPU, Spot),
  guarantee availability during disruptions, set pod priority/preemption.
- Cost optimization: scale-to-zero, Spot/Graviton, consolidation, off-hours shutdown.

## Core concepts

**Requests are the foundation of everything.** `requests` (not `limits`) drive kube-scheduler
placement, HPA's utilization percentage denominator, VPA recommendations, and node autoscaler
capacity math. Without correct requests, *every* autoscaler makes bad decisions. Rule of thumb:

- **CPU** (compressible — throttled, not killed): set `requests` near the 95th-percentile usage.
  Usually **leave CPU `limits` unset** so pods can burst into spare capacity; only set them for
  noisy-neighbor isolation or licensing. kube-scheduler ignores limits anyway.
- **Memory** (incompressible — OOM-killed): set `requests == limits` for predictable, stable
  scheduling. Large gaps risk node overcommit and OOM events.
- If teams don't set requests, enforce defaults with a `LimitRange` (per-container) or a mutating
  webhook (Kyverno). Cap namespace totals with `ResourceQuota`.

**Two scaling dimensions, complementary, not competing:**
- *Workload autoscaling* (pods): HPA / VPA / KEDA react to utilization, custom metrics, or events.
- *Data-plane autoscaling* (nodes): Cluster Autoscaler / Karpenter react **only to unschedulable
  (Pending) pods** — never to CPU/memory utilization. They provision capacity; **kube-scheduler
  still schedules the pods.** This is the single most common misconception.

The loop: load rises → HPA/KEDA add replicas → cluster runs out of room → pods go Pending →
Karpenter/CA add nodes → kube-scheduler places the pods. Load falls → replicas removed → nodes
emptied → nodes removed. You need *both* layers.

**Metrics Server is mandatory** for HPA resource metrics, VPA, and `kubectl top`. It scrapes
kubelet/cAdvisor (~every 15s), holds only current CPU/memory in memory (no history, not a
monitoring tool). HPA custom/external metrics need a different source (Prometheus Adapter, or
KEDA's metrics server). Install via Helm; on local/self-signed clusters add
`--kubelet-insecure-tls`. See `references/hpa.md`.

**HPA target types** (API `autoscaling/v2`):
- `Resource` + `Utilization` (% of requests) or `AverageValue` (absolute).
- `Pods` (per-pod custom metric, always `AverageValue`), `Object` (single k8s object),
  `External` (off-cluster). Multiple metrics → HPA computes desired replicas for each and **takes
  the max**.
- Algorithm: `desiredReplicas = ceil[currentReplicas × (currentMetric / desiredMetric)]`, with a
  10% tolerance band and `behavior` policies / `stabilizationWindowSeconds` to damp flapping.

**VPA** rightsizes a pod's CPU/memory requests (vertical), it does **not** add replicas. Modes:
`Off` (recommend only — safest), `Initial` (apply at creation), `Recreate`/`Auto` (evict to
resize — disruptive unless in-place resize is available). Do **not** run VPA and HPA on the *same*
metric (CPU/memory) — they conflict. VPA-Off + HPA, or VPA + HPA-on-custom-metrics, is fine.

**KEDA** is event-driven autoscaling — it manages an HPA under the hood and extends it with 70+
scalers (queues, Prometheus, cron, cloud services) and, crucially, **scale-to-zero**. KEDA owns
0↔1 activation; HPA owns 1↔N. Use `ScaledObject` for long-running Deployments/StatefulSets,
`ScaledJob` for batch (one Job per unit of work, finishes cleanly). Don't run KEDA and a manual
HPA on the same workload.

**Karpenter vs Cluster Autoscaler:** CA scales predefined node groups (Auto Scaling groups) and
needs *homogeneous* instance types per group; it's infrastructure-first. Karpenter is
application-first — no node groups; it reads pending pods' real requirements and calls the cloud
API directly for right-sized, diverse instances, then **consolidates** (removes/replaces nodes)
to cut cost. Prefer Karpenter for flexibility and bin-packing; CA is fine for simple/stable
fleets. See `references/cluster-autoscaler-and-karpenter.md`.

## Workflow / how to approach autoscaling & scheduling tasks

1. **Always start with requests.** Confirm every container has CPU + memory `requests`. If
   rightsizing, measure real usage (Prometheus/Grafana, `kubectl top`, KRR) over a representative
   window, then set CPU requests ~95th pct (+~20% headroom), memory requests = limits. Wrong
   requests invalidate every downstream autoscaler. See `references/rightsizing-and-patterns.md`.

2. **Pick the pod autoscaler:**
   - CPU/memory tracks load linearly → **HPA** with `Resource`/`Utilization` (start at 70%).
   - Event/queue-driven, bursty, or needs scale-to-zero → **KEDA** (`ScaledObject`, or
     `ScaledJob` for batch). KEDA can also do CPU/memory, so you can standardize on it.
   - Hard-to-scale-horizontally / legacy / stateful, or you just want rightsizing → **VPA**
     (`Off` mode for recommendations, `Auto` only where pod restarts are acceptable).
   - Scaling on latency/RPS/queue depth → HPA custom metrics (Prometheus Adapter) or, simpler,
     KEDA's `prometheus` scaler going straight to the source. See `references/hpa.md`,
     `references/keda.md`, `references/vpa.md`.

3. **Tune scaling speed** with `behavior`: scale up fast (`stabilizationWindowSeconds: 0`,
   `Percent 100 / 15s`), scale down conservatively (longer window) to avoid flapping. KEDA exposes
   the same under `advanced.behavior`. Consider multiple metrics as a safety net (HPA takes the
   max). See `references/hpa.md`.

4. **Add node autoscaling** so Pending pods get capacity. Deploy Karpenter (or CA). For Karpenter,
   define a lean `NodePool` (broad instance categories `c,m,r`, recent generations, capacity-type
   on-demand/spot) + `EC2NodeClass` (AMI/role/subnet/SG selectors). Enable consolidation
   (`WhenEmptyOrUnderutilized`). See `references/cluster-autoscaler-and-karpenter.md`.

5. **Control scheduling & protect availability:**
   - Place/repel pods: `nodeSelector` (hard label match), `nodeAffinity` (required/preferred,
     operators), `podAffinity`/`podAntiAffinity` (co-locate/separate by topology), `taints` +
     `tolerations` (dedicate nodes). See `references/scheduling-affinity-taints.md`.
   - Spread for resilience: `topologySpreadConstraints` (zones, capacity-type) with `maxSkew` and
     `minDomains`. See `references/topology-spread-and-pdb.md`.
   - Guard during disruptions: `PodDisruptionBudget` (`minAvailable`/`maxUnavailable`) + Karpenter
     `NodePool` disruption budgets. Use `PriorityClass` + preemption for critical workloads and
     overprovisioning. See `references/topology-spread-and-pdb.md`.

6. **Layer it together and verify.** Confirm pod constraints fall *within* what the NodePool can
   provision, watch `kubectl get hpa/scaledobject/nodeclaim`, read events and controller logs.
   For graceful scale-down (esp. Spot): `terminationGracePeriodSeconds`, `preStop` hooks, readiness
   probes that fail on SIGTERM. See `references/rightsizing-and-patterns.md`.

## Common pitfalls & anti-patterns

- **No resource requests.** Then HPA can't compute utilization (`<unknown>`), the scheduler
  best-effort-evicts your pods first, and node autoscalers can't size capacity. Fix: set requests.
- **HPA `TARGETS: <unknown>/70%`.** Metrics Server missing/broken, or custom-metrics adapter
  misconfigured. Check `kubectl top pods`, Metrics Server logs, and the adapter/Prometheus query.
- **Running HPA and VPA on the same CPU/memory metric** → they fight (race condition). Use VPA
  `Off` for recommendations, or HPA on custom metrics alongside VPA.
- **Expecting node autoscalers to scale on CPU%.** They scale on **Pending pods** only. If nodes
  aren't growing, check whether pods are actually Pending and *why* (describe the pod's events).
- **Pods Pending forever after a taint/affinity/topology change.** A taint with no matching
  toleration, a `nodeSelector`/required affinity no node (or NodePool) can satisfy, or a
  `DoNotSchedule` topology spread with no node in the required domain. The pod's events tell you;
  ensure a NodePool covers the requested zones/labels.
- **Over-constraining NodePools / using a single instance type.** Starves Karpenter's choice,
  hurts availability and cost. Constrain only what the workload truly needs (let it pick from
  `c,m,r`); push hardware specifics to the pod, not the pool.
- **Scale-to-zero on a synchronous HTTP API.** First request after idle gets a cold start /
  timeout. Use scale-to-zero for queue/event workloads; keep ≥1 replica for always-on APIs (or use
  KEDA's HTTP add-on / a cron floor).
- **`NoExecute` taint without a paired `NoSchedule` taint** when tolerations use
  `tolerationSeconds` → eviction-reschedule loops. Add both.
- **PDB too strict (`minAvailable` = replica count)** blocks all voluntary disruption — Karpenter
  can never drain/consolidate the node. PDBs don't stop *involuntary* disruptions (Spot
  interruptions); add replicas for that.
- **AMI `@latest` in production** → uncontrolled Karpenter drift replaces every node when a new
  image ships. Pin a version; gate rollouts with disruption budgets.

## Reference files

- `references/hpa.md` — Metrics Server install; HPA `autoscaling/v2` full YAML; resource/custom/
  external metrics; target types; the scaling algorithm; `behavior`/stabilization tuning;
  troubleshooting `<unknown>` targets and conditions.
- `references/vpa.md` — VPA install; modes (`Off`/`Initial`/`Recreate`/`Auto`); recommender bounds
  (Lower/Target/Upper/Uncapped); `resourcePolicy`; HPA+VPA coexistence; troubleshooting.
- `references/keda.md` — KEDA architecture; `ScaledObject` vs `ScaledJob`; common scalers
  (cpu/memory, prometheus, cron, queue/SQS/RabbitMQ); scale-to-zero; `TriggerAuthentication`;
  caching/fallback/pause; scaling modifiers; cloud auth; monitoring.
- `references/cluster-autoscaler-and-karpenter.md` — CA on AWS + best practices; Karpenter
  `NodePool`/`EC2NodeClass`/`NodeClaim`; requirements & well-known labels; bin-packing; how to
  choose; descheduler/CPA/CPVA.
- `references/scheduling-affinity-taints.md` — `nodeName`, `nodeSelector`, node affinity, pod
  (anti-)affinity, taints/tolerations (effects, built-ins), static pods, scheduler config.
- `references/topology-spread-and-pdb.md` — `topologySpreadConstraints` (`maxSkew`, `minDomains`,
  Karpenter behavior), `PodDisruptionBudget`, Karpenter NodePool disruption budgets +
  consolidation/drift, `PriorityClass` & preemption, overprovisioning.
- `references/rightsizing-and-patterns.md` — rightsizing methodology & tools; end-to-end scaling
  timeline & metric choice; Spot/Graviton cost optimization; graceful shutdown
  (`terminationGracePeriodSeconds`, `preStop`, readiness on SIGTERM); scale-to-zero nodes.
