# Volume expansion, volumeMode, snapshots & cloning

Contents:
- Expanding a PVC
- volumeMode: Filesystem vs Block
- VolumeSnapshot / VolumeSnapshotClass / restore
- Cloning a PVC
- Prerequisites & gotchas

All of these features require a **CSI driver that supports them**. Legacy in-tree types
cannot expand or snapshot. Verify driver support before promising any of this.

## Expanding a PVC

1. The PVC's StorageClass must have `allowVolumeExpansion: true`.
2. Edit the PVC and increase `spec.resources.requests.storage`. **Only growing is allowed;
   shrinking is rejected.**

```bash
kubectl patch pvc data -n app -p '{"spec":{"resources":{"requests":{"storage":"50Gi"}}}}'
```

The `external-resizer` sidecar grows the backing volume. For a `Filesystem` volume the
filesystem is then expanded — modern CSI drivers do this online (while mounted). If the PVC
shows the condition `FileSystemResizePending`, some drivers complete the filesystem grow only
after the Pod is restarted. Track progress with `kubectl describe pvc data -n app` (watch
`status.capacity` and conditions).

## volumeMode: Filesystem vs Block

Set on the PV/PVC (`spec.volumeMode`):
- **`Filesystem`** (default) — the volume is formatted and mounted at a directory path
  (`volumeMounts.mountPath`). Use for almost everything.
- **`Block`** — the raw block device is exposed to the container with no filesystem, surfaced
  via `volumeDevices.devicePath`. Use for databases/apps that manage their own block I/O.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: raw-block
  namespace: app
spec:
  accessModes: ["ReadWriteOnce"]
  volumeMode: Block
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 20Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: block-consumer
  namespace: app
spec:
  containers:
    - name: app
      image: busybox:1.36
      command: ["sh", "-c", "sleep infinity"]
      volumeDevices:                # note: volumeDevices, not volumeMounts
        - name: data
          devicePath: /dev/xvda
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: raw-block
```

## VolumeSnapshots

A point-in-time copy of a PVC's data. Three CRD objects mirror the PV/PVC pattern:
- **`VolumeSnapshotClass`** — like a StorageClass for snapshots; names the CSI driver and a
  `deletionPolicy` (`Delete`|`Retain`).
- **`VolumeSnapshot`** — the user's request to snapshot a source PVC (namespaced).
- **`VolumeSnapshotContent`** — the actual snapshot on the storage backend (cluster-scoped),
  bound to the VolumeSnapshot by the snapshot controller.

These are `snapshot.storage.k8s.io/v1` and are **not built in** — the cluster needs the
external snapshot CRDs and the snapshot-controller installed, plus a CSI driver advertising
snapshot support (it runs the `csi-snapshotter` sidecar). Verify with
`kubectl get crd | grep snapshot` and `kubectl get volumesnapshotclass`.

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: csi-snapshot
driver: ebs.csi.aws.com
deletionPolicy: Delete
---
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: data-snap-1
  namespace: app
spec:
  volumeSnapshotClassName: csi-snapshot
  source:
    persistentVolumeClaimName: data       # the PVC to snapshot
```

Check readiness: `kubectl get volumesnapshot -n app` → `READYTOUSE` should be `true`.

### Restore a snapshot to a new PVC

Create a PVC whose `dataSource` points at the snapshot. The new PVC must request at least the
snapshot's size and use a compatible StorageClass.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-restored
  namespace: app
spec:
  accessModes: ["ReadWriteOnce"]
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 20Gi
  dataSource:
    name: data-snap-1
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
```

## Cloning a PVC

Duplicate an existing PVC directly (no snapshot step). Source and destination must use the
same StorageClass and the driver must support cloning. Handy for spinning up QA/troubleshooting
copies of production data without touching the original.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-clone
  namespace: app
spec:
  accessModes: ["ReadWriteOnce"]
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 20Gi
  dataSource:
    kind: PersistentVolumeClaim
    name: data                    # source PVC, same namespace & StorageClass
```

## Prerequisites & gotchas

- **Snapshots ≠ backups.** A snapshot usually lives on the same storage backend as the source;
  if the backend dies, so does the snapshot. For real disaster recovery, move data off-cluster
  (see `backup-and-restore-velero.md`).
- **Application consistency.** A raw snapshot is crash-consistent, not necessarily
  app-consistent. Quiesce/flush the app (or use a pre-snapshot hook) for databases.
- **Expansion needs `allowVolumeExpansion: true`** on the class *before* the PVC was created
  matters less than at resize time, but you can't shrink and you can't expand legacy in-tree
  volumes.
- **Snapshot CRDs/controller missing** is the most common reason a `VolumeSnapshot` does
  nothing — the objects are accepted but never become `READYTOUSE`.
