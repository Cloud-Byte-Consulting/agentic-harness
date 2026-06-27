# Pods, Lifecycle, and Health Probes

Deep reference for authoring Pod specs and diagnosing Pod failures.

## Contents
- Pod spec anatomy
- Pod lifecycle: phases and conditions
- Container states
- restartPolicy
- Health probes: liveness, readiness, startup
- Probe handlers and tuning fields
- Pod-state debugging tree
- Useful kubectl commands

## Pod spec anatomy

A complete single-container Pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: web
  labels:
    app: web
    tier: frontend
  annotations:
    team: platform@example.com   # contextual metadata, not used for selection
spec:
  restartPolicy: Always          # Always (default) | OnFailure | Never
  containers:
    - name: web
      image: nginx:1.27
      command: ["nginx"]          # overrides image ENTRYPOINT (optional)
      args: ["-g", "daemon off;"] # overrides image CMD (optional)
      ports:
        - containerPort: 80
      env:
        - name: LOG_LEVEL
          value: "info"
      resources:
        requests: { cpu: 100m, memory: 128Mi }
        limits:   { cpu: 500m, memory: 256Mi }
```

Key points:
- All containers in a Pod share the network namespace (reach each other on `localhost`)
  and the same port space, and can share volumes. A Pod is always scheduled to one node.
- `command` maps to Docker `ENTRYPOINT`; `args` maps to `CMD`. Omit one and Kubernetes
  uses the image's value. Most of the time you only override `args`.
- The Pod is the unit of deletion: deleting a Pod deletes all its containers. Container
  lifecycle is bound to the Pod.

### Labels & selectors (recap)
Labels are key/value pairs ≤63 chars used for identification and selection
(`kubectl get pods -l environment=prod`). Controllers and Services find Pods by label.
Annotations hold arbitrary non-identifying metadata and can be longer; some controllers
and tools read annotations as config (e.g. `kubernetes.io/change-cause`).

### ownerReferences & cascading deletion
Controller-created Pods carry an `ownerReferences` entry pointing at their controller
(visible as `Controlled By: ReplicaSet/...` in `kubectl describe pod`). Deleting the
owner cascades to its children by default (foreground/background). Use
`--cascade=orphan` to delete just the controller and leave Pods running (they can then
be adopted by a controller with a matching selector).

## Pod lifecycle: phases

`status.phase` is a coarse summary:

| Phase | Meaning |
|---|---|
| `Pending` | Accepted but not all containers running yet — image pulling, or unschedulable. |
| `Running` | Bound to a node; at least one container is running/starting/restarting. |
| `Succeeded` | All containers terminated successfully and won't restart. |
| `Failed` | All containers terminated; at least one failed (non-zero exit) and won't restart. |
| `Unknown` | Node state can't be obtained (often node/network problem). |

`Terminating` is **not** a phase — it's what `kubectl` shows while a Pod is in its
deletion grace period (default 30s, `terminationGracePeriodSeconds`). Force with
`--grace-period=0 --force` only when you understand the risk (the Pod may keep running
on a partitioned node).

Pod **conditions** (in `describe`) give finer detail: `PodScheduled`, `Initialized`,
`ContainersReady`, `Ready`. `Ready` gates Service endpoint membership.

## Container states

Each container reports one of:
- **Waiting** — not yet running; the `Reason` is the useful part
  (`ContainerCreating`, `ImagePullBackOff`, `CrashLoopBackOff`).
- **Running** — executing.
- **Terminated** — finished or killed; includes exit code and reason (e.g. `OOMKilled`,
  `Completed`, `Error`).

The READY column `m/n` in `kubectl get pods` counts containers passing their readiness
checks vs total containers. A multi-container Pod showing `1/2` has one container not
ready (failed image, crash, or failing readiness probe).

## restartPolicy

Applies to all containers in the Pod:
- `Always` (default) — restart on any exit. Required for long-running services;
  Deployments/DaemonSets effectively require it.
- `OnFailure` — restart only on non-zero exit. Common for Jobs you want retried.
- `Never` — never restart. Useful for Jobs you want to debug (failed Pods stick around).

Restarts back off exponentially (10s, 20s, 40s … capped at 5 min) — this is the
`CrashLoopBackOff` you see. Restarts are per-container and happen on the same node; the
Pod itself is not rescheduled by a restart.

## Health probes

Three independent probes, each configurable per container. By default **none** are set:
Kubernetes only knows a container "started" and restarts it if the process exits.

### Readiness probe — "can this Pod receive traffic right now?"
On failure, the Pod is removed from Service endpoints but **not** restarted. Use whenever
a container may be running yet temporarily unable to serve (warming caches, waiting on a
dependency, finishing startup). This is the probe you'll add most often.

### Liveness probe — "is this container wedged and in need of a restart?"
On failure (past the threshold), kubelet kills and restarts the container. Use sparingly —
only when the process can deadlock without exiting. **Check the process, not its
dependencies**: a liveness probe that pings the database will restart every replica during
a DB blip, causing a cascading outage.

### Startup probe — "has a slow-starting app finished booting?"
While it's running, liveness and readiness probes are disabled. Once it succeeds, the
others take over. Use for apps with long/variable init so you can keep liveness intervals
tight without premature kills. Give it a generous `failureThreshold × periodSeconds`
budget equal to worst-case startup time.

### Probe handlers (all three probe types accept one of)

```yaml
livenessProbe:
  httpGet:                 # 2xx/3xx = success
    path: /healthz
    port: 8080
    httpHeaders:
      - name: X-Probe
        value: "true"
# --- or ---
  tcpSocket:               # success if TCP connect succeeds
    port: 5432
# --- or ---
  exec:                    # success if command exits 0
    command: ["cat", "/tmp/ready"]
# --- or (1.24+) ---
  grpc:                    # uses gRPC health checking protocol
    port: 9000
```

### Tuning fields (with defaults)

```yaml
readinessProbe:
  httpGet: { path: /ready, port: 80 }
  initialDelaySeconds: 5    # wait before first probe (default 0)
  periodSeconds: 10         # probe interval (default 10)
  timeoutSeconds: 1         # per-probe timeout (default 1)
  successThreshold: 1       # consecutive successes to flip healthy (default 1; must be 1 for liveness/startup)
  failureThreshold: 3       # consecutive failures to flip unhealthy (default 3)
```

Worked example (readiness with a deliberately conservative failure budget):

```yaml
readinessProbe:
  httpGet: { path: /ready, port: 80 }
  initialDelaySeconds: 5
  periodSeconds: 2
  timeoutSeconds: 10
  successThreshold: 1
  failureThreshold: 2       # must fail twice (≈4s) before being pulled from endpoints
```

Best practices:
- Prefer a dedicated lightweight endpoint (e.g. `/healthz`, `/ready`) over `/`.
- Keep readiness checks cheap — they run for the whole Pod lifecycle.
- If you check shared dependencies in readiness, set `timeoutSeconds` greater than the
  dependency's own timeout to avoid synchronized cascading failures.
- Use conservative `initialDelaySeconds`/startup probes to avoid restart loops.
- If your process simply exits on unrecoverable error, you may not need a liveness probe
  at all.

## Pod-state debugging tree

1. `kubectl get pods -o wide` — note STATUS, READY, RESTARTS, NODE.
2. `kubectl describe pod <name>` — read the **Events** section bottom-up.
3. `kubectl logs <name> [-c <container>] [--previous]` — `--previous` shows the last
   crashed container's logs.

| Symptom | Likely cause | Next step |
|---|---|---|
| `Pending` | Unschedulable: insufficient CPU/mem, node selector/affinity, taints | `describe pod` → FailedScheduling event; see kubernetes-autoscaling-scheduling |
| `ContainerCreating` (stuck) | Volume/ConfigMap/Secret mount failing, image pulling | `describe pod` → FailedMount or pull events |
| `ImagePullBackOff` / `ErrImagePull` | Wrong image/tag, private registry auth missing | Fix tag; check `imagePullSecrets` |
| `CrashLoopBackOff` | Container starts then exits | `logs --previous`; check command/args, missing config, failing liveness |
| `Init:0/n` (stuck) | Init container failing/blocking | `logs <pod> -c <init>`; init containers are mandatory and run in order |
| `Running` but `0/1` READY | Readiness probe failing | `describe pod` → Unhealthy event; Pod pulled from Service endpoints |
| `OOMKilled` (Terminated reason) | Memory limit exceeded | Raise `limits.memory` or fix leak; see resource-management-and-qos.md |
| `Error` / non-zero exit | App crashed on start | `logs`; check env/config/dependencies |

## Useful kubectl commands

```bash
kubectl get pods -w                          # watch state transitions live
kubectl get pod <p> -o yaml                   # full live spec (good for backup)
kubectl get pod <p> -o jsonpath='{.spec.containers[*].name}'   # list container names
kubectl describe pod <p>                      # events, conditions, container states
kubectl logs <p> -c <container> --tail=50 -f  # stream a specific container's logs
kubectl logs <p> --previous                   # logs from the prior crashed instance
kubectl exec -it <p> -c <container> -- sh      # shell into a specific container
kubectl port-forward pod/<p> 8080:80           # reach a container locally for testing
```
