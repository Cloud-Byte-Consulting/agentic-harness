# StatefulSets, DaemonSets, Jobs, and CronJobs

The controllers for stateful apps, per-node daemons, and batch/scheduled work.

## Contents
- StatefulSet: identity, ordering, headless Service, volumeClaimTemplates
- StatefulSet scaling, updates, partitions/canary, deletion
- DaemonSet: per-node pods, scheduling, update strategy
- Job: completions, parallelism, backoffLimit, deadlines, TTL
- CronJob: schedule, concurrency, history limits

---

# StatefulSet

Use a StatefulSet when each replica needs a **stable identity** and/or its **own
persistent storage** that survives rescheduling: databases (MySQL, MongoDB), clustered
stores (etcd, Cassandra, Kafka), or anything addressed individually. If replicas are
interchangeable, use a Deployment instead.

What a StatefulSet guarantees that a Deployment does not:
- **Stable, sticky Pod names**: `<name>-0`, `<name>-1`, … (not random hashes).
- **Stable DNS names** per Pod via a headless Service (the Pod IP may change, the DNS
  name does not).
- **Ordered creation/scaling**: Pods come up `0,1,2…`; each waits for its predecessor to
  be Running and Ready. Scale-down terminates in reverse: `N-1 … 1, 0`.
- **Per-Pod storage**: `volumeClaimTemplates` creates one PVC per Pod, bound persistently
  to that Pod's ordinal across restarts/reschedules.

## StatefulSet manifest

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mysql
  labels: { app: mysql }
spec:
  serviceName: mysql-headless        # MUST reference an existing headless Service
  replicas: 3
  podManagementPolicy: OrderedReady  # OrderedReady (default) | Parallel
  updateStrategy:
    type: RollingUpdate              # RollingUpdate (default) | OnDelete
    rollingUpdate:
      partition: 0                   # ordinals >= partition get updated (0 = all)
  selector:
    matchLabels: { app: mysql }
  template:
    metadata:
      labels: { app: mysql }
    spec:
      containers:
        - name: mysql
          image: mysql:8.4
          ports: [{ containerPort: 3306 }]
          env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef: { name: mysql-secret, key: MYSQL_ROOT_PASSWORD }
          volumeMounts:
            - name: data
              mountPath: /var/lib/mysql
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
        # storageClassName: standard   # omit to use the cluster default SC
```

The required **headless Service** (`clusterIP: None`) provides per-Pod DNS A records:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: mysql-headless
spec:
  clusterIP: None
  selector:
    app: mysql
  ports:
    - port: 3306
```

PVC naming follows `<volumeClaimTemplateName>-<statefulSetName>-<ordinal>`, e.g.
`data-mysql-0`. Per-Pod DNS is `<pod>.<headless-service>.<namespace>.svc.cluster.local`,
e.g. `mysql-0.mysql-headless.default.svc.cluster.local`. Connect to a specific replica via
its DNS name (the IP can change after a restart; the name and its PVC do not).

Storage depth (PV/PVC/StorageClass internals, access modes, provisioners) lives in
**kubernetes-storage** — this file shows only the StatefulSet wiring.

## Scaling

`OrderedReady` (default): scaling is sequential and ordered — scale-up waits for each
predecessor to be Ready; scale-down removes highest ordinal first. `Parallel` relaxes this
(Pods come/go together) but does not change update behavior. Scale declaratively
(`spec.replicas` + apply) or `kubectl scale statefulset/mysql --replicas=4`.

## Updates: RollingUpdate, OnDelete, partitions

- **RollingUpdate** (default): on a template change, Pods are recreated one at a time in
  **reverse ordinal order** (`N-1 … 0`), each waiting for the prior to be Ready. A broken
  Pod halts the rollout for manual intervention.
- **OnDelete**: the controller does **not** auto-recreate Pods on a template change; you
  delete a Pod manually to pick up the new template. Useful when each replica needs manual
  verification (e.g. confirming a node rejoined a DB cluster).

**Partitioned / canary rollouts** via `rollingUpdate.partition`: only Pods with ordinal
`>= partition` are updated. Workflow:
1. Set `partition` = `replicas` (e.g. 3) and apply with the new image — *stages* the
   change; nothing updates yet.
2. Lower to `replicas-1` (e.g. 2) — only the highest-ordinal Pod (`mysql-2`) updates =
   your canary. Verify it.
3. Lower further for phased rollout; set to `0` to update all.

## Deletion

```bash
kubectl delete statefulset mysql                  # deletes Pods (unordered) then the STS
kubectl delete statefulset mysql --cascade=orphan # delete STS only; leave Pods
```

Important: PVCs/PVs are **not** deleted by default — your data is preserved; clean up
manually if intended. To automate, set (1.27+):
```yaml
spec:
  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Retain     # or Delete
    whenScaled: Delete      # or Retain
```

Best practices: for a clean reuse, **scale to 0 before deleting** (ordered shutdown);
don't set `terminationGracePeriodSeconds: 0` (skips graceful cleanup/preStop hooks → state
corruption); use remote/network storage so a rescheduled Pod can reattach its volume on any
node; be careful with rollbacks that downgrade across a state-schema change.

---

# DaemonSet

Runs **one Pod per (eligible) node**. As nodes join, Pods are added; as nodes leave, Pods
are removed. There is **no `replicas` field** — the node set determines the count. Used for
node-level infrastructure: log collectors (Fluentd), metrics exporters (Prometheus
node-exporter), CNI agents (Calico/Flannel), kube-proxy, storage daemons (Ceph OSD), and
sometimes Ingress controllers on bare metal.

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
  namespace: logging
  labels: { k8s-app: fluentd-logging }
spec:
  selector:
    matchLabels: { name: fluentd }
  updateStrategy:
    type: RollingUpdate            # RollingUpdate (default) | OnDelete
    rollingUpdate:
      maxUnavailable: 1            # at most N nodes' pods down at once during update
  template:
    metadata:
      labels: { name: fluentd }
    spec:
      # nodeSelector / tolerations control WHICH nodes get a pod (see
      # kubernetes-autoscaling-scheduling for affinity/taints depth)
      nodeSelector:
        kubernetes.io/os: linux
      containers:
        - name: fluentd
          image: quay.io/fluentd_elasticsearch/fluentd:v4.7
          resources:
            requests: { cpu: 100m, memory: 200Mi }
            limits:   { memory: 200Mi }
          volumeMounts:
            - name: varlog
              mountPath: /var/log
      terminationGracePeriodSeconds: 30
      volumes:
        - name: varlog
          hostPath:
            path: /var/log
```

Node targeting: use `nodeSelector` (or affinity/tolerations) on the template to restrict
to a subset of nodes — e.g. Linux-only in a hybrid cluster, or GPU nodes only. Change a
node's labels/taints so it newly matches → a Pod is created there; stop matching → its Pod
is removed. The DaemonSet controller uses node affinity internally to place exactly one Pod
per matching node.

Updates work like Deployments: `RollingUpdate` replaces Pods node-by-node bounded by
`maxUnavailable`; `OnDelete` requires manually deleting a node's Pod to pick up the new
template. `kubectl rollout status/undo ds/<name>` apply. Prioritize critical DaemonSets
with a high `priorityClassName` (system DaemonSets use `system-node-critical`).

Draining a node with `kubectl drain` requires `--ignore-daemonsets` because DaemonSet Pods
are expected to stay until the node leaves.

Alternatives to consider: sidecar pattern (for log gathering tied to one app), CronJob (for
periodic tasks that needn't run on every node), static Pods, or OS-level daemons.

---

# Job

Runs Pods to **completion** (batch work: backups, queue draining, one-off processing),
then stops — unlike a Deployment which runs forever.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: data-export
spec:
  completions: 5            # total successful Pod completions required
  parallelism: 2            # how many Pods run at once
  backoffLimit: 4           # retries before the Job is marked Failed (default 6)
  activeDeadlineSeconds: 600  # hard wall-clock cap; Job fails if exceeded
  ttlSecondsAfterFinished: 3600  # auto-delete Job (and its Pods) 1h after finishing
  template:
    spec:
      restartPolicy: OnFailure   # Job Pods must use OnFailure or Never (NOT Always)
      containers:
        - name: export
          image: busybox:1.36
          command: ["/bin/sh", "-c", "echo exporting; sleep 5"]
```

Knobs:
- **completions** — how many successful runs are needed. With only `completions` set, Pods
  run sequentially (one after another).
- **parallelism** — how many Pods run concurrently. Set `parallelism` without
  `completions` for a work-queue style Job (Pods run in parallel until enough succeed).
- **backoffLimit** — retry budget for failed Pods before the Job gives up. Set low when
  debugging.
- **activeDeadlineSeconds** — terminates the Job after a wall-clock duration regardless of
  progress (takes precedence over `backoffLimit`).
- **ttlSecondsAfterFinished** — garbage-collects the finished Job and its Pods after the
  delay; without it, completed Jobs linger so you can read logs.
- **restartPolicy** — must be `OnFailure` (container restarted in place on failure) or
  `Never` (failed Pod left for inspection; a new Pod created up to `backoffLimit`).
  `Always` is invalid for Jobs.

```bash
kubectl get jobs                 # COMPLETIONS column shows e.g. 5/5
kubectl logs job/data-export     # logs from the Job's pod(s)
kubectl delete job data-export   # cascades to its Pods (--cascade=orphan to keep Pods)
```

---

# CronJob

A controller that creates **Jobs on a schedule** (`CronJob → Job → Pod`). It's a wrapper
around Job using Unix cron syntax.

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: nightly-backup
spec:
  schedule: "0 1 * * *"            # 01:00 every day
  timeZone: "America/New_York"     # optional (1.27+ stable); default is controller's TZ
  concurrencyPolicy: Forbid        # Allow (default) | Forbid | Replace
  startingDeadlineSeconds: 120     # skip a run if it can't start within this window
  successfulJobsHistoryLimit: 3    # keep last 3 successful Jobs (default 3)
  failedJobsHistoryLimit: 1        # keep last 1 failed Job (default 1)
  jobTemplate:                     # a full Job spec
    spec:
      backoffLimit: 2
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: backup
              image: busybox:1.36
              command: ["/bin/sh", "-c", "echo backing up"]
```

Cron schedule fields, left to right: minute (0-59), hour (0-23), day-of-month (1-31),
month (1-12), day-of-week (0-6, Sun=0). `*` = every, `,` = list, `-` = range, `*/n` =
step. Examples: `*/5 * * * *` every 5 min; `0 1 * * 0` 01:00 every Sunday; `30 8 * * 1-5`
08:30 on weekdays. Validate with crontab.guru rather than guessing.

Concurrency policy controls overlapping runs:
- `Allow` (default) — concurrent Jobs may run.
- `Forbid` — skip a new run if the previous is still running.
- `Replace` — cancel the running Job and start the new one.

`startingDeadlineSeconds` bounds how late a missed run may start before it's skipped.
History limits cap how many finished Jobs (and their Pods) are retained.

```bash
kubectl get cronjobs            # SCHEDULE, LAST SCHEDULE, ACTIVE, SUSPEND
kubectl get jobs                # Jobs spawned by the CronJob (named <cronjob>-<timestamp>)
```

Use a CronJob for recurring maintenance; for a one-off, use a Job. For periodic per-node
work, weigh CronJob vs DaemonSet.
