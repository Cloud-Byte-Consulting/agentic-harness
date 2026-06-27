# Deployments, ReplicaSets, and Rollouts

Managing stateless workloads: replication, controlled updates, rollback, and scaling.

## Contents
- ReplicaSet vs Deployment
- The Deployment spec
- Rollout strategies: RollingUpdate vs Recreate
- maxSurge / maxUnavailable / minReadySeconds
- Triggering and watching a rollout
- Rollback and revision history
- Pause / resume
- Scaling
- Canary deployments
- Deleting a Deployment
- Best practices

## ReplicaSet vs Deployment

A **ReplicaSet** keeps a fixed number of identical Pods running, matched by a label
selector (set-based or equality-based). It self-heals: delete a Pod and it makes a new
one; create a bare Pod matching its selector and it deletes one to hold the count.

You rarely create ReplicaSets directly. A **Deployment** owns a ReplicaSet and adds
declarative rollouts, rollback, and versioned revision history. The ownership chain is
`Deployment → ReplicaSet → Pods`. ReplicationController is the obsolete predecessor of
ReplicaSet — never use it.

When you change a Deployment's Pod template, the Deployment creates a **new** ReplicaSet
(identified by a `pod-template-hash` label) and shifts Pods from old to new per the
strategy. Old ReplicaSets are kept (scaled to 0) for rollback.

## The Deployment spec

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  labels: { app: web }
spec:
  replicas: 3
  revisionHistoryLimit: 10        # old ReplicaSets kept for rollback (default 10)
  minReadySeconds: 10             # Pod must stay ready this long to count as available
  selector:
    matchLabels: { app: web }     # immutable after creation
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1           # or a percentage, e.g. 25%
      maxSurge: 1                 # or a percentage
  template:
    metadata:
      labels: { app: web }        # MUST match selector
    spec:
      containers:
        - name: web
          image: nginx:1.27
          ports: [{ containerPort: 80 }]
          resources:
            requests: { cpu: 100m, memory: 128Mi }
            limits:   { cpu: 500m, memory: 256Mi }
          readinessProbe:
            httpGet: { path: /, port: 80 }
            initialDelaySeconds: 5
            periodSeconds: 2
```

The selector is **immutable** once created (changing it would orphan ReplicaSets). The two
operations you perform on a live Deployment are: change the **template** (triggers a
rollout) and change **replicas** (triggers scaling).

Generate a skeleton fast:
```bash
kubectl create deployment web --image=nginx:1.27 --replicas=3 --port=80 \
  --dry-run=client -o yaml > web-deployment.yaml
```

## Rollout strategies

**RollingUpdate (default)** — zero-downtime. Kubernetes brings up new-version Pods while
draining old ones gradually, using two ReplicaSets. Combined with readiness probes and a
Service, traffic only goes to ready Pods. This is what you want for production.

**Recreate** — kills all old Pods, then creates new ones. Causes downtime. Use only when
old and new versions cannot run simultaneously (e.g. an incompatible schema lock) or in
dev. Do not use for production web apps.

```yaml
spec:
  strategy:
    type: Recreate
```

## maxSurge / maxUnavailable / minReadySeconds

For RollingUpdate:
- `maxUnavailable` — how many Pods (below the desired count) may be unavailable during the
  rollout. `1` of `3` means at least 2 must always be ready.
- `maxSurge` — how many extra Pods (above the desired count) may be created during the
  rollout. `1` of `3` means up to 4 Pods may exist transiently.
- Both accept absolute numbers or percentages (`25%`), useful with autoscaling.
- `minReadySeconds` — a new Pod must stay ready this long before it's counted as
  available and the rollout proceeds. Guards against flapping Pods.

Setting `maxUnavailable: 0` with `maxSurge: 1` gives a strictly additive rollout (never
drops below capacity), at the cost of needing room for one extra Pod.

## Triggering and watching a rollout

Edit the image (declaratively, preferred) and apply:
```bash
# change template image to nginx:1.28 in YAML, then:
kubectl apply -f web-deployment.yaml
kubectl rollout status deployment/web        # blocks until complete
kubectl get rs                                # see old RS at 0, new RS at desired count
kubectl describe deployment web               # ScalingReplicaSet events show the shuffle
```

Record why a revision changed (the `--record` flag is deprecated; annotate manually):
```bash
kubectl annotate deployment/web kubernetes.io/change-cause='Update image to 1.28'
```

Imperative image update (dev/debug only):
```bash
kubectl set image deployment/web web=nginx:1.28
```

## Rollback and revision history

Each template change creates a revision with a matching retained ReplicaSet
(`revisionHistoryLimit`, default 10).

```bash
kubectl rollout history deployment/web                 # list revisions
kubectl rollout history deployment/web --revision=2    # inspect a specific revision
kubectl rollout undo deployment/web                    # roll back to previous revision
kubectl rollout undo deployment/web --to-revision=2    # roll back to a specific revision
```

After an undo, the target revision number is retired and a new revision is created (e.g.
history goes `1,2,3` → `1,3,4`). With a GitOps/declarative model, rolling back is usually
just reverting the commit and re-applying.

## Pause / resume

Pause to batch several changes into a single rollout, or to hold mid-rollout while you
investigate:
```bash
kubectl rollout pause deployment/web
# make multiple edits / kubectl set image ... (no rollout happens yet)
kubectl rollout resume deployment/web
```

## Scaling

Declarative (preferred): change `spec.replicas` and `kubectl apply`. Imperative:
```bash
kubectl scale deployment/web --replicas=10
```
Behind a Service, scaled-up Pods are auto-added as endpoints once ready and removed on
scale-down. For automatic scaling use a HorizontalPodAutoscaler — see
**kubernetes-autoscaling-scheduling** (don't hard-set `replicas` in the manifest when an
HPA owns it).

## Canary deployments

Kubernetes doesn't have a built-in canary primitive; you compose one. Simplest approach:
two Deployments sharing the labels a Service selects on, with replica counts setting the
traffic split.

```yaml
# stable: 3 replicas, labels app=myapp, tier=frontend, version=stable, image v1.0
# canary: 1 replica,  labels app=myapp, tier=frontend, version=canary, image v2.0
---
apiVersion: v1
kind: Service
metadata:
  name: frontend
spec:
  selector:                # matches BOTH deployments (no version label)
    app: myapp
    tier: frontend
  ports:
    - port: 80
      targetPort: 8080
```

The Service load-balances across both, so ~1 in 4 requests hits the canary. Monitor error
rate/latency, then either promote (bump stable's image, delete canary) or roll back
(delete canary). For percentage-precise splits independent of replica counts, use an
Ingress controller or service mesh (see kubernetes-networking). Blue/green is a related
technique: run both versions and flip the Service's selector between them.

## Deleting a Deployment

```bash
kubectl delete deployment web                 # deletes Deployment, its ReplicaSets, and Pods
kubectl delete deployment web --cascade=orphan  # delete only the Deployment; leave Pods running
```

## Best practices

- **Declarative everything.** `kubectl apply` from version-controlled YAML; reserve
  imperative commands for dev/debug. (Deleting via imperative command is fine.)
- **Pin meaningful image tags** (semantic versions, or digests). `:latest` makes rollback
  impossible because the tag moves.
- **Set resource requests/limits** on every container (see resource-management-and-qos.md).
- **Configure probes deliberately** — readiness to gate traffic; liveness only with cause.
  Misconfigured probes cause outages.
- **Avoid overlapping selectors** between Deployments or with bare Pods — Kubernetes won't
  stop you, and the result is controllers fighting over Pods. Use semantic labels.
- **Use RollingUpdate for production**; Recreate only when downtime is acceptable.
- **Store secrets/config in Secrets and ConfigMaps**, never inline in the template.
