# Multi-Container Pods and Design Patterns

When and how to put more than one container in a Pod, plus the four canonical patterns.

## Contents
- When to use a multi-container Pod
- Sharing data with volumes (emptyDir)
- Init containers
- Sidecar pattern (classic + native restartable sidecars)
- Ambassador pattern
- Adapter pattern
- Working with individual containers (logs, exec)

## When to use a multi-container Pod

Group containers in one Pod only when they are **tightly coupled** and must share a node,
network namespace, and/or volume. Typical reasons:
- A helper that augments the main app without changing it (log shipper, metrics exporter).
- A local proxy that brokers the main app's outbound connections.
- A format/protocol translator between the app and the outside world.
- One-time setup that must finish before the app starts.

If two components don't need to share fate/volumes/localhost, put them in **separate**
Pods (separate Deployments) and connect them via a Service. Decoupling (e.g. app vs
database) improves stability, availability, and independent scaling.

Multi-container Pods must be created declaratively — `kubectl run` only creates
single-container Pods.

## Sharing data with volumes (emptyDir)

The patterns below rely on containers sharing files. `emptyDir` is an ephemeral volume
created when the Pod is scheduled and deleted when the Pod is removed; it's bound to the
**Pod** lifecycle, so it survives a container restart but not Pod deletion. Mount it into
each container that needs the shared directory:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: shared-dir
spec:
  containers:
    - name: app
      image: alpine:3.20
      command: ["sh", "-c", "while true; do date >> /data/out.log; sleep 5; done"]
      volumeMounts:
        - name: shared
          mountPath: /data
    - name: reader
      image: alpine:3.20
      command: ["sh", "-c", "tail -F /data/out.log"]
      volumeMounts:
        - name: shared
          mountPath: /data
  volumes:
    - name: shared
      emptyDir: {}
```

(`hostPath` mounts a node directory but is an anti-pattern: it ties the Pod to a node and
carries security risks. For real persistence use PVCs — see kubernetes-storage.)

## Init containers

`initContainers` run to completion **in order, before** any app container starts. The next
init container starts only after the previous one exits successfully; if any init container
fails, the Pod restarts it (per `restartPolicy`) and the app containers never start. Use
for setup: fetching config/assets, waiting on a dependency, running migrations, fixing
permissions, seeding a shared volume. They can use a different (heavier) image than the
app, keeping the app image small.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: web-with-init
spec:
  initContainers:
    - name: fetch-content
      image: busybox:1.36
      command:
        - sh
        - -c
        - |
          wget -O /usr/share/nginx/html/index.html https://example.com/index.html
      volumeMounts:
        - name: html
          mountPath: /usr/share/nginx/html
  containers:
    - name: nginx
      image: nginx:1.27
      volumeMounts:
        - name: html
          mountPath: /usr/share/nginx/html
  volumes:
    - name: html
      emptyDir: {}
```

Status during init shows `Init:0/1`, then `PodInitializing`, then `Running`. Read init
logs with `kubectl logs <pod> -c <init-container-name>`.

## Sidecar pattern

A sidecar extends or assists the main container without modifying it — the main container
may not even know it exists. Classic use cases: log collection/forwarding (Fluentd,
Filebeat), metrics exporters, service-mesh proxies (Envoy/Istio), credential agents.
The sidecar usually shares a volume with the main container to read its output.

Classic sidecar (both are normal containers):

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: web-with-logging
spec:
  containers:
    - name: nginx
      image: nginx:1.27
      ports: [{ containerPort: 80 }]
      volumeMounts:
        - name: logs
          mountPath: /var/log/nginx
    - name: log-forwarder            # sidecar
      image: fluent/fluentd:v1.17
      volumeMounts:
        - name: logs
          mountPath: /var/log/nginx
  volumes:
    - name: logs
      emptyDir: {}
```

### Native (restartable) sidecars — Kubernetes 1.28+

Classic sidecars have drawbacks: they can keep running after the main container exits
(wasting resources, blocking Job completion), and Kubernetes doesn't understand their
relationship to the app. Native sidecars fix this by declaring the sidecar as an **init
container with `restartPolicy: Always`**. Such a container starts before the main
containers, keeps running alongside them, and is terminated after them — and it does
**not** block Job completion.

```yaml
spec:
  initContainers:
    - name: logshipper
      image: alpine:3.20
      restartPolicy: Always          # <-- makes this a native sidecar
      command: ["sh", "-c", "tail -F /opt/logs.txt"]
      volumeMounts:
        - name: data
          mountPath: /opt
  containers:
    - name: app
      image: alpine:3.20
      command: ["sh", "-c", "while true; do echo logging >> /opt/logs.txt; sleep 1; done"]
      volumeMounts:
        - name: data
          mountPath: /opt
  volumes:
    - name: data
      emptyDir: {}
```

Prefer native sidecars on 1.28+ for log shippers and especially for sidecars used with
Jobs (a classic sidecar would keep the Job's Pod alive forever).

## Ambassador pattern

An ambassador container proxies the main container's **outbound** connections to an
external service, so the app just talks to `localhost`. The ambassador handles connection
details: SQL proxying, SSL/TLS termination, sharding, retries, endpoint config. The app
is simplified and the connection logic is centralized and swappable per environment.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-ambassador
spec:
  containers:
    - name: app
      image: my-app:1.0           # connects to 127.0.0.1:3306
    - name: db-ambassador         # proxy to the real DB
      image: my-sql-proxy:latest
      ports: [{ containerPort: 3306 }]
      env:
        - name: DB_HOST
          value: db.prod.us-east-1.rds.amazonaws.com
```

Note: the order of containers in the YAML implies no hierarchy — "main" vs "ambassador"
is just convention. Use an ambassador for outbound calls to *external* systems; for
in-cluster calls between Pods, use a Service (see kubernetes-networking).

## Adapter pattern

An adapter transforms the main container's output from format A into format B that some
downstream consumer expects — e.g. normalizing app logs into JSON, or exposing metrics in
Prometheus format. Like the sidecar, it works through a shared volume.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-adapter
spec:
  containers:
    - name: app                    # writes raw logs
      image: alpine:3.20
      command: ["sh", "-c", "while true; do echo \"$(date) raw\" >> /var/log/app/app.log; sleep 5; done"]
      volumeMounts:
        - name: logs
          mountPath: /var/log/app
    - name: log-adapter            # rewrites into the target format
      image: alpine:3.20
      command: ["sh", "-c", "while true; do sed 's/$/ PROCESSED/' /logs/app.log > /logs/processed.log; sleep 10; done"]
      volumeMounts:
        - name: logs
          mountPath: /logs
  volumes:
    - name: logs
      emptyDir: {}
```

## Pattern selection

| Goal | Pattern |
|---|---|
| Run setup before the app, once | init container |
| Add a capability beside the app (logs/metrics/mesh) | sidecar (native on 1.28+) |
| Broker the app's outbound connections | ambassador |
| Convert the app's output to another format | adapter |

## Working with individual containers

In a multi-container Pod you must name the container for log/exec/attach operations:

```bash
kubectl logs <pod> -c <container> --tail=30 -f
kubectl logs <pod> -c <container> --since=2h
kubectl exec -it <pod> -c <container> -- /bin/sh
kubectl get pod <pod> -o jsonpath='{.spec.containers[*].name}'   # discover names
```

Omitting `-c` targets the first container, and `kubectl logs` without `-c` on a
multi-container Pod returns an error prompting you to choose. The READY column `2/2`
confirms both containers are up.
