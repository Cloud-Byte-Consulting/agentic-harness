---
name: kubernetes-storage
description: >-
  Provision and manage stateful storage in Kubernetes. Use for volumes (emptyDir, hostPath,
  configMap/secret, projected, downwardAPI, ephemeral), PersistentVolumes and
  PersistentVolumeClaims, access modes (RWO/ROX/RWX/RWOP), reclaim policies, StorageClasses
  and dynamic provisioning, volumeBindingMode (Immediate vs WaitForFirstConsumer), CSI
  drivers, volume expansion, volumeMode block vs filesystem, StatefulSet volumeClaimTemplates,
  VolumeSnapshots and restore, PVC cloning, and backup/restore with Velero plus etcd
  snapshots. Trigger whenever the user attaches storage to a pod, writes a PVC or
  StorageClass, debugs a PVC stuck Pending or Terminating or a volume that will not mount or
  attach, resizes a volume, sets up snapshots, or plans cluster/namespace backups - even
  without saying Kubernetes. For StatefulSet controller behavior see kubernetes-workloads; for
  storage of big-data engines see kubernetes-data-platforms.
---

# Kubernetes Storage & Data Protection

This skill equips Claude to design, provision, debug, and back up persistent storage in Kubernetes: picking the right volume type, wiring PVs/PVCs and StorageClasses, driving CSI dynamic provisioning, expanding and snapshotting volumes, and protecting data with Velero.

## When to use this skill

- Choosing a volume type for a Pod (scratch space, config injection, persistent data, temporary per-Pod volumes).
- Authoring `PersistentVolume`, `PersistentVolumeClaim`, or `StorageClass` manifests.
- Setting up or debugging **dynamic provisioning** and CSI drivers.
- A **PVC is stuck `Pending`**, a PVC/PV won't delete (`Terminating`/`Released`), or a Pod can't mount a volume (`FailedAttachVolume`, `FailedMount`).
- Expanding a volume, switching `volumeMode`, or controlling where binding happens (`Immediate` vs `WaitForFirstConsumer`).
- Taking and restoring **VolumeSnapshots**, or cloning a PVC.
- Designing per-replica storage with StatefulSet `volumeClaimTemplates`.
- **Backup & disaster recovery**: Velero cluster/namespace/PV backups and schedules, etcd snapshots, restoring to the same or a new cluster.

## Core concepts

**Two layers of storage abstraction.** A *Volume* is declared inside a Pod spec and is (mostly) tied to the Pod's lifecycle. A *PersistentVolume (PV)* is a cluster-scoped object representing a real piece of storage with its own lifecycle, independent of any Pod. A Pod never references a PV directly — it references a *PersistentVolumeClaim (PVC)*, a namespaced request that Kubernetes binds to a matching PV.

**The PV/PVC split mirrors org roles.** PVs (and StorageClasses) are the cluster admin's concern — they map to real hardware/cloud resources. PVCs are the app developer's concern — "I want 10Gi RWO storage". This decoupling is the whole point of the extra object. **A PVC must live in the same namespace as the Pod that mounts it; PVs are not namespaced.**

**Static vs dynamic provisioning.** *Static*: an admin pre-creates the storage and a PV pointing at it; a PVC then binds to it. *Dynamic*: a `StorageClass` + CSI provisioner creates the backing disk **and** the PV on the fly when a PVC is created. Dynamic is the norm today — at scale, hand-managing PVs is untenable. With dynamic provisioning you usually only author the PVC.

**StorageClass is the bridge.** It names a `provisioner` (a CSI driver), default `parameters` (disk type, IOPS, filesystem), a `reclaimPolicy`, a `volumeBindingMode`, and `allowVolumeExpansion`. A cluster can have many StorageClasses; exactly one can be the default (annotation `storageclass.kubernetes.io/is-default-class: "true"`), used when a PVC omits `storageClassName`.

**CSI (Container Storage Interface).** The vendor-neutral plugin standard that replaced the old "in-tree" volume drivers. CSI drivers are containerized (a controller Deployment for provision/attach + a node DaemonSet for mount), referenced by name in a StorageClass `provisioner`. As of recent Kubernetes, in-tree cloud drivers (`awsElasticBlockStore`, `azureDisk`, `gcePersistentDisk`, `cinder`, `vsphereVolume`, etc.) are removed/migrated to CSI — never author manifests using those legacy in-tree types.

**Access modes** (a property of the PV/PVC, constrained by the backend):
- `ReadWriteOnce` (RWO) — read/write by a single **node**.
- `ReadOnlyMany` (ROX) — read-only by many nodes.
- `ReadWriteMany` (RWX) — read/write by many nodes (needs file/shared storage like NFS, CephFS, EFS).
- `ReadWriteOncePod` (RWOP) — read/write by exactly one **Pod** cluster-wide (stable since 1.29; use when you must guarantee a single writer).

Block storage (most cloud disks) is RWO/RWOP only. File/network storage can do RWX.

**Reclaim policy** (what happens to the PV when its PVC is deleted): `Delete` (PV + backing disk wiped — dynamic-provisioning default) or `Retain` (PV kept in `Released` state, data preserved for manual recovery). `Recycle` is deprecated — do not use it.

## Workflow / how to approach storage tasks

### 1. Pick the volume type
- Scratch space shared between containers in a Pod, dies with the Pod → **`emptyDir`** (add `medium: Memory` for a tmpfs RAM disk).
- Inject config/credentials → **`configMap`**, **`secret`**, or **`projected`** (combine several sources, including a short-lived `serviceAccountToken`). Limited to ~1.5 MB (etcd-backed).
- Expose Pod/container metadata as files → **`downwardAPI`**.
- Node-local path (node-bound, fragile) → **`hostPath`** — demo/system-agent only, never for app data.
- Per-Pod disposable volume from a real driver, auto-deleted with the Pod → **generic ephemeral volume** (`ephemeral.volumeClaimTemplate`).
- Data that must survive Pod restarts/rescheduling → **PVC** (almost always via dynamic provisioning).

See `references/volumes.md` for every type with YAML.

### 2. Provision persistent storage (prefer dynamic)
1. Confirm a suitable StorageClass exists: `kubectl get sc`. Look at `PROVISIONER`, `RECLAIMPOLICY`, `VOLUMEBINDINGMODE`, `ALLOWVOLUMEEXPANSION`, and which is `(default)`.
2. Author a PVC requesting size + access mode, referencing the StorageClass (or omit `storageClassName` to use the default). **Do not pre-create a PV for dynamic provisioning.**
3. Mount the PVC in the Pod via `volumes[].persistentVolumeClaim.claimName`.
4. Verify: `kubectl get pvc,pv` — PVC should reach `Bound`.

For static provisioning, label the PV and use a `selector` on the PVC (plus matching `capacity`/`accessModes`/`storageClassName`) so the right PV binds. Full catalogs and binding rules: `references/persistent-volumes-and-claims.md` and `references/storage-classes-and-csi.md`.

### 3. Choose `volumeBindingMode` deliberately
- `Immediate` — PV is provisioned/bound as soon as the PVC is created. Risk: the volume may land in a zone where the Pod can't be scheduled.
- `WaitForFirstConsumer` (WFFC) — binding is delayed until a Pod using the PVC is scheduled, so the volume is created in the right zone/node. **Prefer WFFC** for zonal block storage and for node-local provisioners (LVM, local-path). It's also why a PVC can sit `Pending` "by design" until a consumer Pod exists.

### 4. Expand, snapshot, or clone
- **Expand**: StorageClass must have `allowVolumeExpansion: true`; then edit the PVC's `spec.resources.requests.storage` to a larger value. Shrinking is not allowed. (CSI drivers only — legacy types can't expand.)
- **Snapshot**: needs the snapshot CRDs + a snapshot controller + a CSI driver that supports snapshots. Create a `VolumeSnapshotClass`, then a `VolumeSnapshot` of a source PVC; restore by creating a new PVC with `spec.dataSource` pointing at the snapshot.
- **Clone**: create a PVC with `spec.dataSource` referencing an existing PVC (same StorageClass).

Details and YAML: `references/snapshots-and-expansion.md`.

### 5. Per-replica storage with StatefulSets
Use `volumeClaimTemplates` so each replica gets its own PVC named `<template>-<statefulset>-<ordinal>` (e.g. `data-mysql-0`), bound persistently to that ordinal — a restarted `mysql-0` remounts the same PVC. PVCs from a StatefulSet are **not** deleted automatically on scale-down or delete; clean them up manually or set `persistentVolumeClaimRetentionPolicy` (`whenScaled`/`whenDeleted`: `Retain`|`Delete`, available 1.27+). Controller rollout/scaling mechanics live in **kubernetes-workloads** — this skill covers the storage angle. See `references/persistent-volumes-and-claims.md` for the StatefulSet storage section.

### 6. Back up and protect data
- **etcd** is the cluster state of record — snapshot it with `etcdctl snapshot save` (separate from workload backups).
- **Velero** backs up Kubernetes objects (to S3-compatible object storage) plus PV data (via CSI snapshots, or filesystem backup with Kopia/Restic). Use it for namespace/cluster backups, scheduled backups, and cross-cluster restore/migration.
Full runbook: `references/backup-and-restore-velero.md`.

## Common pitfalls & anti-patterns

- **Using `hostPath` for application data.** It binds the Pod to one node and isn't portable; if the Pod reschedules it starts with no data (split-brain risk). Use a CSI-backed PVC. `hostPath` is for node agents/demos only.
- **Pre-creating a PV when using a StorageClass.** With dynamic provisioning the PV is created for you. A hand-made PV plus a dynamic PVC leads to confusing binding behavior.
- **PVC stuck `Pending`.** Usual causes: no matching/default StorageClass; requested access mode the backend can't satisfy (e.g. RWX on block storage); `WaitForFirstConsumer` and no Pod scheduled yet; provisioner/CSI controller not running; quota exhausted. Check `kubectl describe pvc` events first.
- **PVC stuck `Terminating`.** Almost always the `kubernetes.io/pvc-protection` finalizer: a Pod is still using it. Delete/scale down the consumer Pod and the PVC clears. Don't blindly strip the finalizer — that orphans the volume.
- **Forgetting `Retain` means manual cleanup.** A `Retained` PV goes to `Released` (not `Available`) after its PVC is deleted and will *not* rebind to a new PVC until an admin clears `spec.claimRef`. With `Delete`, deleting the PVC destroys the data — back it up first.
- **Assuming snapshots/expansion "just work".** Both require CSI driver support and (for snapshots) the snapshot controller + CRDs installed. Verify before promising it.
- **Treating Velero as an etcd backup.** Velero backs up API objects and PV *data*, not the etcd database itself. For full cluster DR you need both.
- **`emptyDir` for anything that must survive a restart.** It's deleted when the Pod is removed from the node.
- **Not setting a single default StorageClass** (or having two defaults). PVCs without `storageClassName` then bind unpredictably or stay `Pending`.

## Reference files

- `references/volumes.md` — every volume type (emptyDir, hostPath, configMap/secret, projected, downwardAPI, ephemeral, NFS/FC) with YAML and when to use each.
- `references/persistent-volumes-and-claims.md` — PV/PVC lifecycle, binding, statuses, access modes, reclaim policies, finalizers, stuck-PVC troubleshooting, StatefulSet `volumeClaimTemplates`.
- `references/storage-classes-and-csi.md` — StorageClass fields, default class, `volumeBindingMode`, dynamic provisioning flow, CSI driver architecture.
- `references/snapshots-and-expansion.md` — volume expansion, `volumeMode` (Filesystem vs Block), VolumeSnapshot/VolumeSnapshotClass/restore, PVC cloning.
- `references/backup-and-restore-velero.md` — etcd snapshots, Velero install/backup/schedule/restore, opt-in/opt-out PV backup, Kopia/Restic, cross-cluster migration, gotchas.
