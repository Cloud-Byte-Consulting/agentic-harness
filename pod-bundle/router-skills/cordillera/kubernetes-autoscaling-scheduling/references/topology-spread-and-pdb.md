# Topology Spread, PodDisruptionBudgets, Disruption Budgets & Priorities

These controls govern *resilience* and *the pace of disruption*: how evenly pods spread across
failure domains, how many pods/nodes may be taken down at once, and which workloads win when
capacity is scarce.

## Contents
- topologySpreadConstraints (maxSkew, minDomains)
- PodDisruptionBudget (PDB)
- Karpenter consolidation, drift, expiration
- Karpenter NodePool disruption budgets (NDB)
- PDB + NDB together; Spot caveats
- PriorityClass & preemption; overprovisioning

---

## topologySpreadConstraints

Declaratively spread pods across domains (zones, nodes, capacity-type) — more general and usually
better than pod anti-affinity for "balance my replicas."

```yaml
spec:
  topologySpreadConstraints:
  - topologyKey: "topology.kubernetes.io/zone"
    maxSkew: 1
    whenUnsatisfiable: DoNotSchedule       # hard: pod waits rather than break the spread
    minDomains: 3                          # require pods across ≥3 zones (else scheduler uses what exists)
    labelSelector:
      matchLabels: {app: web}
  - topologyKey: "karpenter.sh/capacity-type"
    maxSkew: 2
    whenUnsatisfiable: ScheduleAnyway      # soft: prefer balance, don't block
    labelSelector:
      matchLabels: {app: web}
```

- `maxSkew` — max allowed difference in pod count between the busiest and emptiest domain.
- `whenUnsatisfiable`: `DoNotSchedule` (hard) or `ScheduleAnyway` (soft preference).
- `minDomains` — the scheduler only knows about domains that currently have nodes; without
  `minDomains` it may pack into 2 zones even if a 3rd is possible. Set `minDomains=N` to force
  spreading across the intended count.

**Karpenter** is topology-aware: a `DoNotSchedule` constraint with no node in a required domain
makes Karpenter launch a node *in that domain* (if the NodePool allows it). So the NodePool must
span the targeted zones/capacity-types — otherwise pods stay Pending. With zonal volumes (EBS),
Karpenter may add zonal affinity that skews spread; the `nodeAffinityPolicy` setting can tell the
scheduler to ignore volume-derived affinity when computing spread.

---

## PodDisruptionBudget (PDB)

A PDB limits how many pods of an app may be **voluntarily** evicted at once (drains,
consolidation, node upgrades) — it guards application availability. It does **not** stop
*involuntary* disruptions (node crash, Spot interruption).

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: web-pdb
spec:
  minAvailable: 2            # or maxUnavailable: "25%"
  selector:
    matchLabels: {app: web}
```

- Use **either** `minAvailable` **or** `maxUnavailable` (absolute number or percentage).
- **Too strict** (`minAvailable` = replica count) blocks *all* voluntary disruption — Karpenter can
  never drain that node to consolidate/upgrade it. You'll see node events like
  `Unconsolidatable - pdb default/web-pdb prevents pod evictions`.
- PDBs don't help against involuntary loss; for that, raise replica count.

---

## Karpenter disruption: consolidation, drift, expiration

Karpenter actively manages node lifecycle. Disruption types:
- **Voluntary** (Karpenter-initiated, respects PDBs + budgets + `do-not-disrupt`): consolidation,
  drift.
- **Involuntary** (can't be deferred/throttled): Spot interruption, health events, `expireAfter`.

**Consolidation** (`spec.disruption.consolidationPolicy`):
- `WhenEmpty` — remove only nodes with no app pods (DaemonSet/static pods ignored).
- `WhenEmptyOrUnderutilized` (default) — also remove/replace underutilized nodes:
  - *remove empty* nodes (in parallel, no replacement),
  - *consolidate underutilized* (pack pods onto fewer nodes),
  - *downsize* (replace a node with a cheaper/smaller one),
  - *spot-to-spot* (feature-gated; needs ≥15 viable instance types in the Fleet request).
- `consolidateAfter` — wait this long after pod churn before treating a node as a candidate
  (prevents thrashing; `0s` reacts immediately).

Karpenter **launches replacements before draining** the old node, and verifies via simulation that
every pod can reschedule before disrupting. It **skips** consolidation for: blocking PDBs;
`karpenter.sh/do-not-disrupt: "true"` (on pod or node); required/preferred affinity that can't be
satisfied elsewhere (preferred treated like required for consolidation; tunable via
`PREFERENCE_POLICY=Ignore`); no cheaper single-node replacement (Karpenter only consolidates *into
one* replacement node); or a disruption budget at its limit.

**Drift** — when a node no longer matches the NodePool/NodeClass spec (new AMI, changed instance
families/security groups, k8s version), Karpenter marks it drifted and does a rolling replacement
(respecting PDBs/NDBs). **Pin AMIs** (`bottlerocket@1.38.0`, not `@latest`) so a new image doesn't
drift every node mid-day; gate rollouts with budgets.

**Expiration** (`expireAfter`, e.g. `720h`) — caps node lifetime (node freshness, limits long-lived
Spot risk). Involuntary — *not* throttled by NDBs. Pair with `terminationGracePeriod` for a hard
drain cap.

```bash
alias kl='kubectl -n karpenter logs -l app.kubernetes.io/name=karpenter --all-containers=true -f --tail=20'
kl   # watch "disrupting node(s)", "reason":"underutilized"/"drifted", "tainted node", "deleted node"
```

---

## Karpenter NodePool disruption budgets (NDB)

NDBs throttle **voluntary** disruptions per NodePool (consolidation + drift). They do **not** stop
involuntary disruption (Spot, expiration, manual delete). Default (if unset): 10% of the NodePool
at a time.

```yaml
spec:
  disruption:
    consolidationPolicy: WhenEmptyOrUnderutilized
    consolidateAfter: 0s
    budgets:
      - nodes: "25%"                       # general cap (percentage scales with pool size)
      - nodes: "5"                         # absolute global ceiling
      - nodes: "0"                         # block these reasons entirely...
        reasons: ["Drifted"]               # ...e.g. do drift/upgrades manually for now
      - nodes: "0"                         # time-boxed block during business hours (UTC only)
        schedule: "0 9 * * 1-5"
        duration: 8h
```

- `nodes` — percentage (`roundup(total × pct) − deleting − notready`) or absolute
  (`N − deleting − notready`). Counting NotReady nodes means a bad rollout self-throttles.
- `reasons` — limit a budget to `Empty`, `Underutilized`, and/or `Drifted`. If a reason has **no**
  matching budget, it is **unconstrained** (can disrupt unlimited) — always cover the reasons you
  care about, or add a catch-all.
- `schedule` + `duration` — time windows (cron, **UTC only**, `@daily` macros). `nodes: "0"`
  without a schedule blocks that reason forever.
- **Multiple budgets → most restrictive wins** (minimum across applicable budgets). Common pattern:
  a percentage that scales + an absolute hard ceiling.

---

## PDB + NDB together; Spot caveats

- **PDB** = application level (which pods stay up). **NDB** = platform level (how many nodes drain
  at once). Karpenter always respects PDBs during drain; NDBs add a global cap so it never floods
  the cluster with parallel drains.
- Set NDBs as a safety net even where apps lack PDBs (e.g. `nodes: "10%"`).
- **Spot interruptions are involuntary** (~2-min warning, then forced). PDBs/NDBs can't block them.
  Karpenter watches the interruption queue, cordons/drains, and pre-launches a replacement. For
  Spot-heavy clusters: define narrow PDBs for critical apps, keep NDBs conservative (e.g.
  `nodes: "1"`), and add enough replicas that simultaneous interruptions don't breach the PDB
  minimum.

---

## PriorityClass & preemption

Pod priority decides who gets scarce capacity. A higher-priority Pending pod can **preempt**
(evict) lower-priority pods to schedule.

```yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000000
globalDefault: false
preemptionPolicy: PreemptLowerPriority    # or Never (high priority, but won't evict others)
description: "Critical workloads"
```
```yaml
# Pod
spec:
  priorityClassName: high-priority
```

**Overprovisioning / warm capacity** pattern: run low-priority "pause" placeholder pods
(`preemptionPolicy` low value, e.g. negative) that reserve headroom. When real workloads spike,
they preempt the placeholders and schedule instantly while Karpenter/CA launches new nodes (which
then re-host the placeholders). This hides cold-start node-launch latency for predictable surges.

---

## Graceful scale-down & node-launch timing

(See `rightsizing-and-patterns.md` for the full graceful-shutdown lab.) Key levers when nodes are
removed (consolidation, drift, Spot): `terminationGracePeriodSeconds` (pod), `preStop` hooks, and a
readiness probe that **fails on SIGTERM** so load balancers deregister the pod before it exits.
Karpenter's termination flow taints the node `NoSchedule`, evicts via the Eviction API (honoring
PDBs, skipping DaemonSet/static pods), waits for volume detach and connection draining, then
terminates the instance and removes the finalizer.
