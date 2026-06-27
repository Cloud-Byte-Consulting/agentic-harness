# Volume types

Pod-level volumes are declared in `spec.volumes` and surfaced into containers with
`spec.containers[].volumeMounts`. Most are tied to the Pod lifecycle (data dies with the
Pod); PVC-backed and snapshot-backed storage is the exception (see
`persistent-volumes-and-claims.md`). Pick the lightest type that meets the durability and
sharing requirements.

## Quick chooser

| Need | Use |
|---|---|
| Scratch space shared by containers in one Pod, OK to lose | `emptyDir` |
| In-memory (tmpfs) scratch | `emptyDir` with `medium: Memory` |
| Inject config files | `configMap` |
| Inject credentials/certs | `secret` |
| Combine config + secret + token + downward API in one mount | `projected` |
| Expose Pod/container metadata as files | `downwardAPI` |
| Node-local directory (agents/demos only) | `hostPath` |
| Per-Pod disposable volume from a real CSI driver, auto-deleted | generic `ephemeral` |
| Data that must survive Pod restart/reschedule | `persistentVolumeClaim` |
| Shared network filesystem mounted directly | `nfs` (prefer a CSI NFS driver + PVC) |

## emptyDir

Created empty when the Pod is assigned to a node; deleted permanently when the Pod is
removed from the node. Survives container crashes/restarts (the Pod stays), not Pod
deletion. Great for caches, scratch, and sharing files between containers in a Pod.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: scratch-demo
spec:
  containers:
    - name: app
      image: busybox:1.36
      command: ["sh", "-c", "sleep infinity"]
      volumeMounts:
        - name: cache
          mountPath: /cache
  volumes:
    - name: cache
      emptyDir:
        sizeLimit: 1Gi          # optional cap
        # medium: Memory        # uncomment for a tmpfs RAM-backed dir
```

## hostPath

Mounts a file or directory from the node's filesystem. **Node-bound and a security risk** —
do not use for application data. Legitimate uses: node-level agents (log shippers, CNI,
monitoring) that genuinely need host access. Set `type` to make intent explicit and safe.

```yaml
  volumes:
    - name: varlog
      hostPath:
        path: /var/log
        type: Directory          # DirectoryOrCreate, File, Socket, ...
```

## configMap

Mounts ConfigMap keys as files. Updates to the ConfigMap propagate to the mounted files
(eventually; not for `subPath` mounts). Cap ~1.5 MB (etcd-backed). Use `items` to project
specific keys, or `defaultMode`/`mode` to set file permissions.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  app.conf: |
    log_level = info
    workers = 4
---
apiVersion: v1
kind: Pod
metadata:
  name: cm-demo
spec:
  containers:
    - name: app
      image: nginx:1.27
      volumeMounts:
        - name: config
          mountPath: /etc/app          # app.conf appears as /etc/app/app.conf
          readOnly: true
  volumes:
    - name: config
      configMap:
        name: app-config
        # items:                       # optional: project only some keys
        #   - key: app.conf
        #     path: app.conf
```

## secret

Identical mechanics to configMap but for `Secret` objects (base64 in the API, plaintext in
the mount). Mounted secrets are stored in tmpfs (RAM), never written to node disk. Always
mount `readOnly`.

```yaml
  containers:
    - name: app
      image: nginx:1.27
      volumeMounts:
        - name: tls
          mountPath: /etc/tls
          readOnly: true
  volumes:
    - name: tls
      secret:
        secretName: app-tls
        defaultMode: 0400
```

## projected

Combines multiple sources (configMap, secret, downwardAPI, and a short-lived, audience-bound
`serviceAccountToken`) into a single directory. The serviceAccountToken source is the modern
way to get a rotating, scoped token for workload identity / OIDC federation.

```yaml
  volumes:
    - name: combined
      projected:
        sources:
          - configMap:
              name: app-config
          - secret:
              name: app-tls
          - downwardAPI:
              items:
                - path: labels
                  fieldRef:
                    fieldPath: metadata.labels
          - serviceAccountToken:
              path: token
              audience: vault
              expirationSeconds: 3600
```

## downwardAPI

Exposes Pod/container fields and resource limits as files (parallel to the env-var form).

```yaml
  volumes:
    - name: podinfo
      downwardAPI:
        items:
          - path: namespace
            fieldRef:
              fieldPath: metadata.namespace
          - path: cpu_limit
            resourceFieldRef:
              containerName: app
              resource: limits.cpu
```

## Ephemeral volumes (generic, per-Pod, CSI-backed)

A **generic ephemeral volume** gives a Pod its own freshly provisioned volume from a real
StorageClass that is created and **deleted together with the Pod**. It's an inline PVC
template — you get CSI features (snapshots/expansion if the driver supports them) without a
standalone PVC object. Ideal for per-Pod scratch that's too big for `emptyDir`/node disk.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: ephemeral-demo
spec:
  containers:
    - name: app
      image: busybox:1.36
      command: ["sh", "-c", "sleep infinity"]
      volumeMounts:
        - name: scratch
          mountPath: /scratch
  volumes:
    - name: scratch
      ephemeral:
        volumeClaimTemplate:
          metadata:
            labels:
              type: scratch
          spec:
            accessModes: ["ReadWriteOnce"]
            storageClassName: standard
            resources:
              requests:
                storage: 10Gi
```

There is also a **CSI ephemeral volume** type (`csi:` inline in the Pod) for drivers that
expose lightweight, driver-managed ephemeral data (e.g. secrets-store CSI). It does not use
a PVC and is only available if the specific driver advertises ephemeral support.

## Network volumes mounted directly (NFS / FC)

You *can* reference network storage inline in a Pod, but this hard-codes infrastructure into
the workload and breaks the config/resource decoupling. **Prefer a PV/PVC** (ideally via a
CSI NFS driver). Inline NFS, shown for completeness:

```yaml
  volumes:
    - name: shared
      nfs:
        server: nfs.example.com
        path: /exports/app
        readOnly: true
```

Fibre Channel and iSCSI follow the same inline pattern (`fc:` with `targetWWNs`/`lun`,
`iscsi:` with `targetPortal`/`iqn`) but are almost always wired through a PV instead.
