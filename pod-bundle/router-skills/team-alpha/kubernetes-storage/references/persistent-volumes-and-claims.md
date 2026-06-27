# PersistentVolumes & PersistentVolumeClaims

Contents:
- PV/PVC model and lifecycle
- Access modes
- Reclaim policies
- Static provisioning (PV + PVC with selector)
- Dynamic provisioning (PVC only)
- Statuses cheat sheet
- Finalizers & deletion ordering
- Troubleshooting: stuck Pending / Terminating / Released
- StatefulSet volumeClaimTemplates (per-replica PVCs)

## The model

- **PersistentVolume (PV)** — cluster-scoped object representing a real piece of storage
  (cloud disk, NFS export, CSI volume). It is *not* the disk itself, just a pointer +
  metadata. Created by an admin (static) or by a provisioner (dynamic).
- **PersistentVolumeClaim (PVC)** — namespaced request for storage (size, access mode,
  class). A Pod mounts a PVC, never a PV directly.
- **Binding** — Kubernetes matches a PVC to a suitable PV (capacity ≥ request, compatible
  access modes, same `storageClassName`, satisfies any `selector`) and binds them 1:1.

**Rule:** the PVC must be in the **same namespace** as the Pod that mounts it. PVs are not
namespaced.

### Lifecycle stages

Provision → Available (PV) → Bound (PV+PVC) → in use by Pod → PVC deleted →
Released/Deleted per reclaim policy.

## Access modes

| Mode | Short | Meaning |
|---|---|---|
| `ReadWriteOnce` | RWO | read/write by a single **node** (many Pods on that node OK) |
| `ReadOnlyMany` | ROX | read-only by many nodes |
| `ReadWriteMany` | RWX | read/write by many nodes — needs shared/file storage (NFS, CephFS, EFS) |
| `ReadWriteOncePod` | RWOP | read/write by exactly one **Pod** cluster-wide (stable 1.29+) |

Block/cloud-disk backends are RWO/RWOP only. Requesting RWX from block storage is a classic
reason a PVC never binds. RWOP is the tool when you must *guarantee* a single writer (e.g.
a non-clustered database) even across nodes.

## Reclaim policies

Set on the PV (`spec.persistentVolumeReclaimPolicy`) or inherited from the StorageClass
(`reclaimPolicy`). Controls what happens when the bound PVC is deleted:

- **`Delete`** — PV object and the backing disk are removed. Default for dynamically
  provisioned volumes. Convenient but destroys data; back up first if it matters.
- **`Retain`** — PV is kept, moves to `Released`. Data is preserved for manual recovery.
  The PV will **not** rebind to a new PVC until an admin clears `spec.claimRef`.
- **`Recycle`** — deprecated, do not use. Use dynamic provisioning instead.

Reclaim policy is mutable on an existing PV:

```bash
kubectl patch pv my-pv -p '{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}'
```

To recover a `Retain`ed PV for reuse: `kubectl patch pv my-pv -p '{"spec":{"claimRef":null}}'`
(it returns to `Available`).

## Static provisioning

Admin creates the PV; developer's PVC targets it via matching attributes (and optionally a
label selector). Useful for pre-existing NFS exports or hand-managed disks.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
  labels:
    tier: shared
spec:
  capacity:
    storage: 5Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  storageClassName: nfs-manual      # any string; "" means no class
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    server: nfs.example.com
    path: /exports/app
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-claim
  namespace: app
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: nfs-manual      # must match the PV
  resources:
    requests:
      storage: 5Gi
  selector:
    matchLabels:
      tier: shared
```

Setting `storageClassName: ""` on both PV and PVC opts out of dynamic provisioning entirely
(prevents the default StorageClass from interfering).

## Dynamic provisioning (the common case)

Author only the PVC; the StorageClass + CSI provisioner creates the PV and backing disk.
**Do not create a PV.** Omit `storageClassName` to use the cluster default.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data
  namespace: app
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: standard        # omit to use the default StorageClass
  resources:
    requests:
      storage: 20Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: app
  namespace: app
spec:
  containers:
    - name: app
      image: nginx:1.27
      volumeMounts:
        - name: data
          mountPath: /var/lib/app
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: data
```

`kubectl get pvc,pv -n app` — the PVC should be `Bound` to an auto-named PV
(`pvc-<uuid>`).

## Statuses cheat sheet

**PV:** `Available` (free) · `Bound` (claimed) · `Released` (PVC gone, Retain policy, awaiting
admin) · `Failed` (provisioner/storage error) · `Unknown` (lost contact with backend).

**PVC:** `Pending` (no matching PV yet / waiting for consumer) · `Bound` (in use) ·
`Terminating` (delete requested, finalizer holding).

## Finalizers & deletion ordering

- PVCs carry the `kubernetes.io/pvc-protection` finalizer; PVs carry
  `kubernetes.io/pv-protection`. These block deletion while the object is in use, so data
  isn't yanked out from under a running Pod.
- Correct teardown order: **delete the Pod(s) first, then the PVC.** Deleting the PVC while a
  Pod uses it leaves the PVC `Terminating` until the Pod is gone.

## Troubleshooting

Always start with `kubectl describe pvc <name>` and read the Events.

**PVC stuck `Pending`:**
- No default StorageClass and the PVC omitted `storageClassName` → set a default or name a class.
- Requested access mode unsupported by backend (e.g. RWX on block storage) → use RWO or a file backend.
- `volumeBindingMode: WaitForFirstConsumer` and no Pod scheduled yet → **expected**; create/schedule the consumer Pod.
- CSI controller / provisioner Pod not running → check the driver's namespace.
- ResourceQuota on storage exhausted → check `kubectl describe quota -n <ns>`.
- Static binding: no PV matches capacity/accessModes/selector/class → fix the PV or the request.

**PVC stuck `Terminating`:** a Pod is still mounting it (pvc-protection finalizer). Find and
remove the consumer (`kubectl get pods -o jsonpath` for the claim name, or scale down the
workload). Only as a last resort, and knowing it orphans the disk, clear the finalizer:
`kubectl patch pvc <name> -n <ns> -p '{"metadata":{"finalizers":null}}'`.

**PV stuck `Released` (won't rebind):** Retain policy left a `claimRef`. Clear it:
`kubectl patch pv <name> -p '{"spec":{"claimRef":null}}'` → returns to `Available`.

**Pod `FailedAttachVolume` / `FailedMount`:** node can't attach/mount — often an RWO volume
still attached to a previous node (stuck `VolumeAttachment`), a zone mismatch (volume in a
different AZ than the node — use `WaitForFirstConsumer`), or a missing CSI node plugin on
that node. Check `kubectl get volumeattachment` and the node's CSI DaemonSet.

## StatefulSet volumeClaimTemplates (per-replica PVCs)

A StatefulSet's `volumeClaimTemplates` creates a **separate PVC per replica**, named
`<templateName>-<statefulSetName>-<ordinal>` (e.g. `data-web-0`, `data-web-1`). Each PVC is
bound persistently to its ordinal: when `web-0` is rescheduled it remounts the **same** PVC
and data. This is the storage half of StatefulSets — the controller/rollout mechanics
(ordered create/scale/update, headless Service identity) belong to the **kubernetes-workloads**
skill.

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  namespace: app
spec:
  serviceName: web-headless        # headless Service governs network identity
  replicas: 3
  selector:
    matchLabels: { app: web }
  template:
    metadata:
      labels: { app: web }
    spec:
      containers:
        - name: web
          image: nginx:1.27
          volumeMounts:
            - name: data
              mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: standard    # omit -> default StorageClass
        resources:
          requests:
            storage: 1Gi
```

Storage facts to know:
- PVCs from `volumeClaimTemplates` are **not deleted** on scale-down or StatefulSet delete —
  data is protected by default. Leftover PVCs accumulate cost; clean up manually:
  `kubectl delete pvc data-web-2 -n app`.
- Control automatic cleanup with `persistentVolumeClaimRetentionPolicy` (Kubernetes 1.27+):

  ```yaml
  spec:
    persistentVolumeClaimRetentionPolicy:
      whenScaled: Delete      # delete PVCs of removed replicas on scale-down
      whenDeleted: Retain     # keep PVCs when the StatefulSet is deleted
  ```
- For data to survive Pod rescheduling across nodes, the StorageClass must provide storage
  reachable from any node (network/replicated). A zonal disk pins the replica to its zone.
- Editing `volumeClaimTemplates` on an existing StatefulSet is largely immutable — plan
  capacity up front, or expand the individual PVCs (see `snapshots-and-expansion.md`).
