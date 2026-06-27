# StorageClasses & CSI

Contents:
- StorageClass anatomy
- The default StorageClass
- volumeBindingMode (Immediate vs WaitForFirstConsumer)
- allowVolumeExpansion & reclaimPolicy
- Dynamic provisioning flow
- CSI driver architecture
- Multiple classes / tiers
- Inspecting & changing the default

## StorageClass anatomy

A StorageClass tells Kubernetes *how* to dynamically provision storage. It is the bridge
between a developer's abstract PVC and a concrete CSI driver.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: ebs.csi.aws.com         # the CSI driver name
parameters:                          # driver-specific knobs
  type: gp3
  iops: "3000"
  throughput: "125"
  encrypted: "true"
reclaimPolicy: Delete                # Delete (default) | Retain
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
mountOptions:
  - noatime
# allowedTopologies:                 # optionally restrict zones
#   - matchLabelExpressions:
#       - key: topology.ebs.csi.aws.com/zone
#         values: ["us-east-1a", "us-east-1b"]
```

Field notes:
- `provisioner` — the CSI driver that handles this class. Examples: `ebs.csi.aws.com`,
  `disk.csi.azure.com`, `pd.csi.storage.gke.io`, `nfs.csi.k8s.io`, `rook-ceph.rbd.csi.ceph.com`,
  `local.csi.openebs.io`. `kubernetes.io/no-provisioner` is used for purely static/local PVs.
- `parameters` — opaque to Kubernetes; each driver defines its own keys (disk type, fs type,
  IOPS, replication factor, encryption). **Immutable** after creation.
- `reclaimPolicy` — inherited by PVs this class provisions. `Delete` by default.
- `volumeBindingMode` — see below.
- `allowVolumeExpansion` — must be `true` for PVCs of this class to be resizable.

> Legacy in-tree provisioners like `kubernetes.io/aws-ebs`, `kubernetes.io/gce-pd`,
> `kubernetes.io/azure-disk`, `kubernetes.io/cinder` are removed/migrated to CSI in current
> Kubernetes. Author StorageClasses with the **CSI driver name**, not these.

## The default StorageClass

The class marked default is used by any PVC that omits `storageClassName`. It's identified by
an annotation:

```yaml
metadata:
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
```

- A cluster should have **exactly one** default. Zero → PVCs without a class stay `Pending`.
  Two → unpredictable selection (and a warning).
- The default differs by distribution: minikube `standard` (`k8s.io/minikube-hostpath`), AKS
  `default`/`managed-csi` (`disk.csi.azure.com`), EKS typically `gp2`/`gp3` (`ebs.csi.aws.com`),
  GKE `standard-rwo` (`pd.csi.storage.gke.io`).

## volumeBindingMode

- **`Immediate`** — the PV is provisioned and bound the moment the PVC is created, before any
  Pod is scheduled. Risk: with zonal storage the volume may be created in a zone where the
  consuming Pod can't run, causing unschedulable Pods.
- **`WaitForFirstConsumer` (WFFC)** — provisioning/binding waits until a Pod referencing the
  PVC is scheduled, so topology (node/zone) is known and the volume is placed correctly.
  **Prefer WFFC** for zonal block storage and node-local provisioners. Side effect: a PVC
  legitimately sits `Pending` ("waiting for first consumer") until a Pod uses it — not a bug.

## Dynamic provisioning flow

1. Admin installs a CSI driver and creates one or more StorageClasses.
2. Developer creates a PVC referencing a class (or the default).
3. The external-provisioner sidecar sees the PVC and calls the CSI driver to create a volume
   in the backend.
4. The driver creates the disk and a PV is generated and bound to the PVC.
5. When a Pod using the PVC is scheduled, the CSI controller **attaches** the volume to the
   node and the CSI node plugin **mounts** it into the Pod.

## CSI driver architecture

CSI (Container Storage Interface) is the vendor-neutral plugin standard. A driver typically
ships as:
- A **controller** component (Deployment/StatefulSet) running sidecars: `external-provisioner`
  (Create/DeleteVolume), `external-attacher` (Controller Publish/attach), `external-resizer`
  (expansion), `external-snapshotter` (snapshots) — plus the vendor's controller logic.
- A **node** component (DaemonSet) that runs on every node to stage/publish (mount) volumes
  into Pods.

You never call CSI directly; you reference the driver by name in a StorageClass `provisioner`
and consume volumes through PVCs (or inline ephemeral volumes). Installed drivers appear in
`kubectl get csidrivers`. Check a driver's snapshot/expansion/RWX support before relying on it.

## Multiple classes / tiers

Offer different tiers so developers pick by name without knowing the backend. Example: a
standard HDD-style class and a fast SSD class.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
provisioner: pd.csi.storage.gke.io
parameters:
  type: pd-standard
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast
provisioner: pd.csi.storage.gke.io
parameters:
  type: pd-ssd
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

Multiple StorageClasses can share one provisioner with different `parameters`.

## Inspecting & changing the default

```bash
kubectl get sc
# NAME                 PROVISIONER            RECLAIMPOLICY  VOLUMEBINDINGMODE     ALLOWVOLUMEEXPANSION
# standard (default)   pd.csi.storage.gke.io  Delete         WaitForFirstConsumer  true

kubectl get sc fast-ssd -o yaml          # full details

# Make a class the default
kubectl patch sc fast-ssd -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
# Remove default from another
kubectl patch sc standard -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
```
