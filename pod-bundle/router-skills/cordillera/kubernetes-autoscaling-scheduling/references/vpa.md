# VerticalPodAutoscaler (VPA)

VPA adjusts a pod's **CPU/memory `requests`** (and optionally limits) to match real usage — it
rightsizes pods, it does **not** change replica count. Use it for workloads that are hard to scale
horizontally (legacy monoliths, some databases) or purely to get rightsizing recommendations.

VPA is **not built into Kubernetes** — install the CRDs and three controllers from the
`kubernetes/autoscaler` repo. API group: `autoscaling.k8s.io/v1`.

## Architecture (three controllers)

- **Recommender** — watches usage (from Metrics Server) + history, computes optimal CPU/memory
  requests per container.
- **Updater** — decides when to apply recommendations to running pods; in active modes it **evicts**
  pods so they're recreated with new values (unless in-place resize is available).
- **Admission Controller** — a mutating webhook that rewrites requests on pod creation so new pods
  start already rightsized.

> Pods must be terminated & recreated to change requests unless the `InPlacePodVerticalScaling`
> feature gate is on (alpha-introduced in 1.27; default-on in newer releases) — then VPA can resize
> in place without a restart.

## Install

```bash
git clone https://github.com/kubernetes/autoscaler.git
cd autoscaler/vertical-pod-autoscaler
./hack/vpa-up.sh
kubectl get pods -n kube-system | grep vpa   # recommender, updater, admission-controller
```

## A VPA object

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: montecarlo-pi-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: montecarlo-pi
  updatePolicy:
    updateMode: "Off"            # see modes below
  resourcePolicy:
    containerPolicies:
    - containerName: '*'         # or a specific container; use Off here to exclude a container
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1200m
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
```

### Update modes (`spec.updatePolicy.updateMode`)
- **`Off`** — recommend only; nothing is changed. **Safest** and the most common production choice.
  Read the recommendation with `kubectl describe vpa`.
- **`Initial`** — apply recommendations only when pods are *created* (next rollout/restart). No
  evictions of running pods.
- **`Recreate`** — apply at creation and **evict** running pods to resize them. Disruptive.
- **`Auto`** — currently equivalent to `Recreate`; will use in-place resize where the feature gate
  is enabled, avoiding restarts.

`resourcePolicy.containerPolicies`:
- `minAllowed` / `maxAllowed` — bound the recommendation (always set `maxAllowed` so an app can't
  scale toward an unschedulable size).
- `controlledResources` — which resources VPA manages (`cpu`, `memory`).
- Set `mode: "Off"` on a specific `containerName` to exclude a sidecar from resizing.

## Reading recommendations

```bash
kubectl get vpa            # MODE, CPU, MEM, PROVIDED columns
kubectl describe vpa montecarlo-pi-vpa
```

The recommendation has four values per container:
- **Lower Bound** — minimum that's still reasonable; below this the pod is under-provisioned.
- **Target** — what VPA will set on the container.
- **Uncapped Target** — what Target would be ignoring `min/maxAllowed`.
- **Upper Bound** — maximum reasonable value.

Notes that surprise people:
- The recommender enforces its own **minimum memory** (default ~256Mi) that overrides a lower
  `minAllowed` (tunable via `pod-recommendation-min-memory-mb`).
- Active modes won't evict a pod until it has run long enough — default **12h** (tunable via
  `in-recommendation-bounds-eviction-lifetime-threshold` on the updater). Force a refresh with
  `kubectl rollout restart deployment NAME`.
- By default VPA only acts when there is **>1 replica** (to limit downtime). Add `minReplicas` at
  the VPA-rule level to override per-rule.

## HPA + VPA coexistence (important)

Do **not** run HPA and VPA on the **same metric** (CPU or memory) — they create a race: HPA adds
replicas while VPA raises per-pod requests, causing oscillation and poor efficiency. Workable
patterns:
- **HPA for scaling + VPA in `Off` mode** for sizing recommendations you apply manually.
- **HPA on custom/external metrics + VPA on CPU/memory** — they manage different things, no
  conflict.
- **VPA only** for workloads that can't scale horizontally (often `Auto` in pre-prod, `Off` in
  prod so you plan resize windows).

## Troubleshooting

```bash
kubectl describe vpa NAME      # Conditions (RecommendationProvided) + Recommendation
kubectl logs -n kube-system -l app=vpa-recommender --all-containers=true --tail=20
kubectl logs -n kube-system -l app=vpa-updater     --all-containers=true --tail=20
kubectl logs -n kube-system -l app=vpa-admission-controller --all-containers=true --tail=20
```

VPA not applying changes? Check:
- Components running and CRDs installed; Metrics Server healthy.
- `updateMode` is `Auto`/`Recreate` (not `Off`/`Initial`).
- `maxAllowed` or a namespace `LimitRange`/`ResourceQuota` isn't capping the recommendation.
- Target pods are owned by a controller (Deployment/RS/STS) — VPA won't evict bare pods.
- The cluster has capacity for the larger pod (otherwise the resized pod goes Pending — pair with
  node autoscaling).
- The 12h eviction-lifetime threshold hasn't elapsed yet.
