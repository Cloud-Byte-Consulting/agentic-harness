# etcd Backup & Restore

etcd holds **all** cluster state. No etcd backup = no disaster recovery. Snapshot regularly, store
off-cluster, and **test restores**. (On managed clusters the provider backs up etcd — this applies
to self-managed/kubeadm clusters.)

## Contents
- [Where etcd lives](#where-etcd-lives)
- [Taking a snapshot](#taking-a-snapshot)
- [Verifying a snapshot](#verifying-a-snapshot)
- [Restoring a snapshot](#restoring-a-snapshot)
- [Scheduling backups](#scheduling-backups)
- [HA / multi-member notes](#ha--multi-member-notes)

## Where etcd lives

With kubeadm, etcd runs as a **static Pod** on each control-plane node (manifest in
`/etc/kubernetes/manifests/etcd.yaml`), with data under `/var/lib/etcd` and certs under
`/etc/kubernetes/pki/etcd/`. It listens on **2379** (client) and **2380** (peer).

Install `etcdctl` (matching your etcd version) — it ships in the etcd release, or you can
`kubectl exec` into the etcd static Pod which already has it. Always use the v3 API.

## Taking a snapshot

Run on a control-plane node, pointing at the local etcd member with its client certs:
```bash
ETCDCTL_API=3 etcdctl snapshot save /backup/etcd-snapshot-$(date +%F-%H%M).db \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key
```
Or via the static Pod:
```bash
kubectl -n kube-system exec etcd-<cp-node> -- sh -c \
  'ETCDCTL_API=3 etcdctl snapshot save /var/lib/etcd-backup.db \
   --endpoints=https://127.0.0.1:2379 \
   --cacert=/etc/kubernetes/pki/etcd/ca.crt \
   --cert=/etc/kubernetes/pki/etcd/server.crt \
   --key=/etc/kubernetes/pki/etcd/server.key'
```
**Copy the `.db` file off the node** (object storage / another host). A snapshot stranded on the
node that died is no backup.

## Verifying a snapshot

```bash
ETCDCTL_API=3 etcdctl --write-out=table snapshot status /backup/etcd-snapshot-*.db
```
Shows hash, revision, total keys, size — confirm it's non-empty and the key count looks sane.

## Restoring a snapshot

Restore is a **deliberate, cluster-down** operation. The cluster will revert to the snapshot's
state; anything created after the snapshot is lost.

1. **Stop the control plane.** Move the static-Pod manifests out so the kubelet stops api-server and
   etcd:
   ```bash
   sudo mv /etc/kubernetes/manifests/{etcd.yaml,kube-apiserver.yaml} /tmp/
   ```
2. **Restore into a fresh data dir:**
   ```bash
   ETCDCTL_API=3 etcdctl snapshot restore /backup/etcd-snapshot-2026-06-12.db \
     --data-dir=/var/lib/etcd-restored
   ```
   (For a multi-member cluster, also pass `--name`, `--initial-cluster`,
   `--initial-cluster-token`, and `--initial-advertise-peer-urls` matching that member.)
3. **Point etcd at the restored data dir.** Edit `/etc/kubernetes/manifests/etcd.yaml` so the
   `hostPath` volume for `/var/lib/etcd` maps to `/var/lib/etcd-restored` (or move the directory
   into place).
4. **Restart the control plane** by moving the manifests back:
   ```bash
   sudo mv /tmp/{etcd.yaml,kube-apiserver.yaml} /etc/kubernetes/manifests/
   ```
   The kubelet recreates the static Pods.
5. **Verify:** `kubectl get nodes`, `kubectl get pods -A`, `kubectl get --raw='/readyz?verbose'`.

For an **HA cluster**, restore on **one** member, bring it up as a single-member cluster, then have
the other control-plane nodes re-join (rebuild their etcd from this restored leader) rather than
restoring each independently.

## Scheduling backups

Automate and monitor:
- A systemd timer or cron on a control-plane node running the `snapshot save` + off-node copy.
- Or a Kubernetes `CronJob` (`batch/v1`) mounting the etcd certs and pushing to object storage.
- Retain a rolling window (e.g. hourly for a day, daily for a month).
- **Periodically test a restore in a throwaway cluster** — an untested backup is a hope, not a plan.

## HA / multi-member notes

- Snapshot from any one healthy member; the snapshot is the full keyspace.
- Quorum requires a majority of members; etcd member counts should be **odd** (3 or 5). Losing
  quorum (e.g. 2 of 3 members down) makes the cluster read-only/unavailable until quorum returns or
  you restore.
- Manage members with `etcdctl member list / add / remove` (same cert flags). Replacing a dead
  control-plane node means removing its stale etcd member and joining a new one.
