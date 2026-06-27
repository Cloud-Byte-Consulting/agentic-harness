---
name: kubernetes-workloads
description: >-
  Author, debug, and operate Kubernetes workload objects - Pods, Deployments, ReplicaSets,
  StatefulSets, DaemonSets, Jobs, and CronJobs - plus the Pod-spec building blocks inside
  them: init and sidecar containers, liveness/readiness/startup probes, ConfigMaps, Secrets,
  resource requests/limits and QoS classes, labels, selectors, and ownerReferences. Use
  whenever the user writes or reviews workload YAML; asks how to roll out, roll back, scale,
  or update an app; mentions apps/v1 or batch/v1 kinds; debugs Pod states like
  CrashLoopBackOff, ImagePullBackOff, Pending, Init, or OOMKilled; or configures probes,
  envFrom, volume-mounted config, completions/parallelism, cron schedules,
  RollingUpdate/maxSurge, headless Services for stateful pods, or volumeClaimTemplates - even
  without saying Kubernetes. For affinity/taints/topology and HPA autoscaling see
  kubernetes-autoscaling-scheduling; for PV/PVC/StorageClass depth see kubernetes-storage; for
  Service/Ingress see kubernetes-networking.
---

# Kubernetes Workloads

This skill equips Claude to author correct, production-grade Kubernetes workload
manifests and to diagnose why workloads misbehave. It covers the Pod and every
controller that manages Pods, plus the configuration and resource plumbing that
lives inside a Pod spec.

## When to use this skill

- Writing or reviewing YAML for a `Pod`, `Deployment`, `ReplicaSet`, `StatefulSet`,
  `DaemonSet`, `Job`, or `CronJob`.
- Choosing the right controller: "should this be a Deployment or StatefulSet?",
  "I need one pod per node", "run this nightly", "process a queue in parallel".
- Designing multi-container Pods: init containers, sidecars (incl. native sidecars),
  ambassador, adapter.
- Configuring health probes (liveness/readiness/startup) and explaining why a probe
  is restarting or de-registering a Pod.
- Injecting configuration: ConfigMaps and Secrets via `env`, `envFrom`, volume mounts,
  `subPath`, projected volumes, immutability.
- Setting resource `requests`/`limits` and reasoning about QoS (Guaranteed/Burstable/
  BestEffort) and OOMKills.
- Rollouts: RollingUpdate vs Recreate, `maxSurge`/`maxUnavailable`, pause/resume,
  rollback, `revisionHistoryLimit`, canary.
- Debugging Pod phases and statuses: `Pending`, `ImagePullBackOff`, `CrashLoopBackOff`,
  `Init:0/1`, `OOMKilled`, `Terminating`.
- Labels, selectors, annotations, and `ownerReferences` / cascading deletion.

Boundary — defer and cross-link, don't duplicate:
- Pod scheduling (nodeAffinity, taints/tolerations, topologySpreadConstraints) and
  autoscaling (HPA/VPA/cluster autoscaler) → **kubernetes-autoscaling-scheduling**.
- PersistentVolume / PersistentVolumeClaim / StorageClass internals → **kubernetes-storage**.
  (This skill shows StatefulSet `volumeClaimTemplates` *mechanics* only.)
- Service / Ingress / Gateway / NetworkPolicy → **kubernetes-networking**.

## Core concepts

**The Pod is the only thing that runs containers.** You never create a container
directly; you create a Pod, and the kubelet on the scheduled node turns it into one or
more containers sharing a network namespace (same localhost, same port space) and able
to share volumes. A Pod never spans nodes. Treat Pods as cattle: disposable, recreatable,
stateless where possible (push state to external storage).

**Everything above a Pod is a controller** running a reconcile loop that drives current
state toward desired state. The ownership chain for a typical stateless app is
`Deployment → ReplicaSet → Pod`; for a scheduled batch task it's `CronJob → Job → Pod`.
You almost never create bare Pods or bare ReplicaSets in production — use a controller so
you get self-healing, rollouts, and scaling.

**Pick the controller by workload shape:**

| Need | Use |
|---|---|
| Stateless replicas, rolling updates, rollback | `Deployment` |
| Stable identity, ordered start/stop, per-pod storage | `StatefulSet` |
| Exactly one pod per (eligible) node | `DaemonSet` |
| Run-to-completion batch task | `Job` |
| Scheduled / recurring task | `CronJob` |
| Raw replica count, no rollout (rarely direct) | `ReplicaSet` |

**Stable API versions (use these — older groups like `extensions/v1beta1` are removed):**
- `apps/v1` → Deployment, ReplicaSet, StatefulSet, DaemonSet
- `batch/v1` → Job, CronJob (CronJob graduated to `batch/v1` in 1.21; `batch/v1beta1` is gone)
- `v1` → Pod, ConfigMap, Secret, Service
- `scheduling.k8s.io/v1` → PriorityClass

**Labels are load-bearing, not decoration.** Controllers find their Pods via
`spec.selector` matching `spec.template.metadata.labels`. A mismatch is rejected by the
API server. Overlapping selectors between two controllers cause silent fighting over
Pods — use disciplined, semantic labels (`app`, `tier`, `environment`, `version`).
Annotations carry non-identifying metadata (contact info, tooling config, change-cause).

**Minimal correct Deployment** (the shape you'll write most often):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  labels: { app: web }
spec:
  replicas: 3
  selector:
    matchLabels: { app: web }
  template:
    metadata:
      labels: { app: web }      # MUST match the selector
    spec:
      containers:
        - name: web
          image: nginx:1.27      # pin a real tag, never :latest in prod
          ports:
            - containerPort: 80
          resources:
            requests: { cpu: 100m, memory: 128Mi }
            limits:   { cpu: 500m, memory: 256Mi }
          readinessProbe:
            httpGet: { path: /healthz, port: 80 }
            initialDelaySeconds: 5
            periodSeconds: 5
```

## Workflow / how to approach workload tasks

**1. Identify the workload shape first.** Stateless web/API → Deployment. Needs stable
network name or per-replica disk (databases, clustered stores) → StatefulSet. Node-level
agent (logs, metrics, CNI, storage daemon) → DaemonSet. One-shot or scheduled job →
Job / CronJob. When unsure between Deployment and StatefulSet, ask: does each replica
need its own identity/storage that must survive a reschedule? If no → Deployment.

**2. Write declarative YAML, not imperative commands.** Use `kubectl apply -f` and keep
manifests in source control (GitOps). Imperative `kubectl run`/`create`/`scale`/`set image`
/`rollout undo` are for dev, debugging, and learning only. Generate a starting skeleton
with `--dry-run=client -o yaml`:
```bash
kubectl create deployment web --image=nginx:1.27 --replicas=3 --port=80 \
  --dry-run=client -o yaml > web-deployment.yaml
```

**3. Always set resource requests and limits.** Requests drive scheduling and QoS;
limits cap usage. Omitting them yields BestEffort Pods that are killed first under
pressure. See `references/resource-management-and-qos.md`.

**4. Add the right probes.** Readiness gates traffic; liveness restarts a wedged
container; startup protects slow-booting apps from premature liveness kills. Don't add a
liveness probe "just because" — a misconfigured one causes restart loops and cascading
failures. See `references/pods-and-probes.md`.

**5. Decouple configuration.** Never bake env-specific config or secrets into the image.
Use ConfigMaps for non-sensitive data and Secrets for sensitive data, injected via
`env`/`envFrom` or volume mounts. See `references/config-and-secrets.md`.

**6. Choose and tune the rollout strategy.** Default `RollingUpdate` with
`maxSurge`/`maxUnavailable` for zero-downtime; `Recreate` only when downtime is fine and
versions can't coexist. Verify with `kubectl rollout status`; roll back with
`kubectl rollout undo`. See `references/deployments-and-rollouts.md`.

**7. Verify the result.** `kubectl get`/`describe` the object, watch Pods with `-w`,
read container logs with `kubectl logs`, and check `kubectl get events`. The READY column
(e.g. `2/2`) tells you how many containers are up; STATUS plus the `describe` Events
section is your primary debugging surface.

**Debugging quick map** (full trees in the reference files):
- `Pending` → unschedulable (resources, selectors) — `kubectl describe pod`.
- `ImagePullBackOff` / `ErrImagePull` → bad image name/tag or registry auth.
- `CrashLoopBackOff` → container starts then exits; check `kubectl logs` (and
  `--previous`); often bad command/args, missing config, or failing liveness probe.
- `Init:0/n` stuck → an init container is failing or blocking; check its logs with
  `kubectl logs <pod> -c <init-container>`.
- `OOMKilled` (in `describe`) → exceeded memory limit; raise limit or fix the leak.
- `0/1` READY but Running → readiness probe failing; Pod is up but pulled from Service
  endpoints.

## Common pitfalls & anti-patterns

- **Selector ↔ template label mismatch.** The API server rejects it; double-check both
  blocks carry the same labels.
- **Treating containers like VMs.** One concern per container; use multi-container Pods
  (sidecar/adapter/ambassador) for helpers, not one fat container running several daemons.
- **`:latest` image tags in production.** Breaks reproducibility and makes rollback
  impossible (the tag moves). Pin semantic versions or digests (`@sha256:...`).
- **No resource requests/limits.** Yields BestEffort Pods and noisy-neighbor problems;
  the scheduler can't place them well and they're evicted first.
- **Liveness probe that checks dependencies** (DB, downstream service). A blip restarts
  every replica → cascading failure. Liveness checks the process; readiness checks
  dependencies.
- **Secrets in environment variables when they could be files.** Env Secrets can leak via
  logs/`describe`/child processes; prefer volume mounts for sensitive values. And
  remember base64 in a Secret is encoding, not encryption.
- **Using `Recreate` for production web apps.** Guarantees downtime; use `RollingUpdate`.
- **`kubectl delete` on a StatefulSet you intend to reuse.** Pods die unordered; scale to
  0 first for a clean, ordered shutdown. PVCs are *not* deleted by default — clean them up
  explicitly (or set `persistentVolumeClaimRetentionPolicy`).
- **Forgetting the headless Service for a StatefulSet.** `spec.serviceName` must point to
  an existing `clusterIP: None` Service or per-pod DNS won't work.
- **Bare Pods with labels matching a controller's selector.** The controller adopts or
  deletes them unexpectedly. Give standalone Pods unique labels.
- **`restartPolicy: Always` on a Job Pod.** Jobs require `Never` or `OnFailure`; `Always`
  is invalid for Jobs/CronJobs.

## Reference files

- **references/pods-and-probes.md** — Pod spec anatomy, lifecycle phases & container
  states, `restartPolicy`, the three probe types with full tuning fields, and a Pod-state
  debugging tree. Read when authoring a Pod spec or diagnosing why a Pod won't run/serve.
- **references/multi-container-patterns.md** — init containers, sidecar (classic + native
  1.28+ restartable init), ambassador, adapter; shared `emptyDir` volumes; accessing logs
  and exec-ing into a specific container. Read for any multi-container Pod design.
- **references/deployments-and-rollouts.md** — ReplicaSet vs Deployment, rollout
  strategies, `maxSurge`/`maxUnavailable`, `minReadySeconds`, pause/resume, rollback,
  revision history, canary, and scaling. Read for stateless app management.
- **references/statefulsets-daemonsets-jobs.md** — StatefulSet identity/ordering/headless
  Service/`volumeClaimTemplates`/partitioned & canary rollouts; DaemonSet update strategy
  and node targeting; Job `completions`/`parallelism`/`backoffLimit`/`activeDeadlineSeconds`/
  `ttlSecondsAfterFinished`; CronJob schedule, concurrency, history limits. Read when the
  workload has state, runs per-node, or runs to completion.
- **references/config-and-secrets.md** — creating/consuming ConfigMaps & Secrets (env,
  envFrom, volume, `subPath`, projected volumes), immutability, Secret types, and security
  guidance. Read for any configuration-injection task.
- **references/resource-management-and-qos.md** — requests vs limits, CPU/memory units,
  the three QoS classes and how they're derived, OOMKill behavior, LimitRange/ResourceQuota
  pointers. Read when sizing containers or debugging eviction/OOM.
