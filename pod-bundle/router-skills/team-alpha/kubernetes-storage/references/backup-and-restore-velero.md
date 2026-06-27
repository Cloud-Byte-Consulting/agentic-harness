# Backup & restore: etcd snapshots and Velero

Contents:
- What needs backing up
- etcd snapshots
- Velero overview & components
- Installing Velero
- Backing up (one-time, scheduled, filtered)
- Backing up PV data: CSI snapshots vs filesystem (Kopia/Restic)
- opt-in vs opt-out PV selection
- Restoring (same cluster, namespace, new cluster / migration)
- Velero CLI cheat sheet
- Gotchas

## What needs backing up

A complete Kubernetes DR strategy covers three things:
1. **etcd** — the cluster's state of record (all API objects). Lose every etcd member with no
   snapshot and the cluster is gone.
2. **Kubernetes API objects** — namespaces, Deployments, Services, RBAC, ConfigMaps, etc.
3. **Persistent volume data** — the bytes inside PVs.

`etcdctl` covers (1). **Velero** covers (2) and (3). They are complementary — Velero is *not*
an etcd backup, and an etcd snapshot does *not* capture PV data.

## etcd snapshots

etcd ships `etcdctl` with a built-in snapshot command. Run it against the API endpoint with
the client certs (typically under `/etc/kubernetes/pki/etcd/`). etcd 3.4+ defaults to API v3,
so no `ETCDCTL_API=3` is needed.

```bash
etcdctl snapshot save etcd-snapshot.db \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt \
  --key=/etc/kubernetes/pki/etcd/healthcheck-client.key

# verify
etcdctl --write-out=table snapshot status etcd-snapshot.db
```

An `x509: certificate signed by unknown authority` error means wrong certs — re-check the
`--cacert/--cert/--key` paths. In production, schedule this (e.g. CronJob/systemd timer) and
ship the snapshot to secure off-host storage. Also back up the control-plane PKI in
`/etc/kubernetes/pki`. (Managed control planes — EKS/AKS/GKE — handle etcd for you; you only
need Velero there.)

## Velero overview

Open-source (originally Heptio Ark, now under VMware Tanzu) backup/restore/migration tool.
Components:
- **velero CLI** — installs Velero and drives all backup/restore operations (there is no GUI).
- **Velero server** — controllers running in-cluster (the `velero` namespace) that execute jobs.
- **Object storage** — an S3-compatible bucket holds backups. Use AWS S3, GCS, Azure Blob, or
  self-hosted **MinIO** for on-prem/labs.
- **Provider plugin** — e.g. `velero-plugin-for-aws` (works with any S3-compatible store
  including MinIO).
- **Node agent (optional)** — a DaemonSet that performs filesystem-level PV backups using a
  data mover (**Kopia** by default, or Restic).

Velero stores backups as CRDs in-cluster (`Backup`, `Schedule`, `Restore`,
`BackupStorageLocation`, `VolumeSnapshotLocation`) plus the data in the bucket.

## Installing Velero

Provide bucket credentials in a file (`credentials-velero`), then install. Example targeting
MinIO with filesystem PV backup enabled (Kopia) and opt-out volume backup as default:

```bash
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.10.0 \
  --bucket velero \
  --secret-file ./credentials-velero \
  --use-volume-snapshots=false \
  --backup-location-config region=minio,s3ForcePathStyle="true",s3Url=http://minio.velero.svc:9000 \
  --use-node-agent \
  --default-volumes-to-fs-backup
```

Key flags:
- `--provider` / `--plugins` — storage provider and its plugin image (`aws` works for MinIO).
- `--bucket` / `--secret-file` — target bucket and its credentials.
- `--use-volume-snapshots` — enable native CSI/cloud volume snapshots (set `false` if using
  filesystem backup instead, or if the backend has no snapshot support).
- `--use-node-agent` — deploy the filesystem-backup DaemonSet (needed for Kopia/Restic).
- `--default-volumes-to-fs-backup` — make filesystem backup the default for all PVs (opt-out
  mode). Omit to use opt-in mode.

`credentials-velero` (S3-style):
```
[default]
aws_access_key_id = <key>
aws_secret_access_key = <secret>
```

## Backing up

```bash
# One-time backup of the entire cluster (all namespaces + PV data per the install flags)
velero backup create initial-backup

# Watch status
velero backup describe initial-backup
velero backup logs initial-backup
velero get backups
```

A successful backup shows `Phase: Completed` and `Items backed up == Total items`. If
fewer items were backed up than expected, investigate before relying on it.

### Scheduled backups (cron)

```bash
velero schedule create cluster-daily-1am --schedule="0 1 * * *"
velero schedule create cluster-daily     --schedule="@daily"     # shorthand
velero get schedules
```

Shorthands: `@yearly`, `@monthly`, `@weekly`, `@daily`, `@hourly`. Scheduled runs create
backups named `<schedule>-<YYYYMMDDhhmmss>` (UTC). Set retention with `--ttl` (default
`720h0m0s` = 30 days).

### Filtering what's backed up

```bash
velero backup create app-only --include-namespaces app,app-staging
velero backup create no-system --exclude-namespaces kube-system,kube-node-lease
velero backup create labeled --selector app.kubernetes.io/name=ingress-nginx
velero backup create no-sc --exclude-resources storageclasses.storage.k8s.io
```

## Backing up PV data

Two mechanisms:
- **Volume snapshots** — Velero asks the CSI driver / cloud provider to snapshot the PV.
  Fast, storage-side, and application-consistent when supported. Enable with
  `--use-volume-snapshots=true` (and a `VolumeSnapshotLocation`).
- **Filesystem backup (FSB)** — the node agent copies file contents using the data mover
  (Kopia default, or Restic via `--data-mover restic`). Used when the backend lacks snapshot
  support. Data lands in the `kopia`/`restic` folder of the bucket; objects go in `backups`.

### opt-in vs opt-out volume selection

- **opt-out** (installed with `--default-volumes-to-fs-backup`, or per-backup with the same
  flag): *all* PVs are backed up except those excluded via a Pod annotation:
  ```
  backup.velero.io/backup-volumes-excludes=volume2,volume3
  ```
- **opt-in** (default without that flag): only volumes named in an annotation are backed up:
  ```bash
  kubectl -n app annotate pod/<pod> backup.velero.io/backup-volumes=data,logs
  ```

### hostPath / local-path limitation

Velero's filesystem backup **cannot back up `hostPath` volumes**. The common
`local-path-provisioner` maps PVs to hostPath by default. Work around it by annotating PVCs to
use the `local` volume type so FSB can read them:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data
  namespace: app
  annotations:
    volumeType: local
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: 1Gi
```

## Restoring

A restore is created *from* a backup; the restore object is named `<backup>-<timestamp>`.

```bash
# Restore everything from a backup
velero restore create --from-backup initial-backup

# Restore a single namespace (e.g. a dev deleted their namespace)
velero restore create --from-backup initial-backup --include-namespaces app

velero restore describe <restore-name>     # InProgress -> Completed
velero restore logs <restore-name>
velero get restores
```

### Cross-cluster restore / migration

1. Stand up the new cluster and install Velero pointing at the **same** bucket (use the
   externally reachable S3 URL, not an in-cluster service name).
2. Velero syncs and discovers the existing backups: `velero get backups` lists them.
3. Restore the desired backup: `velero restore create --from-backup <name>`.

This is how you migrate workloads between clusters or build a DR/dev clone. Velero restores
namespaces, objects, and (with FSB/snapshots) PV data into the new cluster.

## Velero CLI cheat sheet

```bash
velero install ...                              # deploy server components
velero backup create <NAME> [--include-namespaces ...] [--default-volumes-to-fs-backup]
velero backup describe <NAME>                   # details / phase
velero backup logs <NAME>
velero backup delete <NAME>
velero backup get
velero schedule create <NAME> --schedule="<cron>"
velero schedule get
velero restore create --from-backup <BACKUP>
velero restore describe <NAME>
velero restore get
velero get backup-locations
velero version
```

You can also inspect the CRDs directly when the CLI isn't handy:
`kubectl describe backups <name> -n velero`, `kubectl get schedules -n velero`.

## Gotchas

- **Velero is not an etcd backup.** Pair it with `etcdctl` snapshots (on self-managed
  control planes) for full cluster DR.
- **Targeting MinIO by in-cluster service name** (`minio.velero.svc:9000`) makes some
  `velero describe` queries error from outside the cluster, and breaks cross-cluster restore
  — expose an external URL for those cases.
- **Changing the node-agent repository password after backups exist** makes old backups
  unreadable. Set it before the first backup, if at all.
- **Default ServiceAccount tokens / ConfigMaps mounted as volumes** are backed up as objects,
  not as volume data — don't expect FSB to capture them.
- **Always verify** `Phase: Completed` and item counts; treat a failed/partial backup as no
  backup. Wire alerts (e.g. Alertmanager) off backup status.
