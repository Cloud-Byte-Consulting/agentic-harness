# Cluster Troubleshooting

Triage *cluster* health — nodes, control-plane components, runtime. (App/Pod debugging beyond the
node level belongs to kubernetes-workloads.)

## Contents
- [Triage order](#triage-order)
- [Where the logs are](#where-the-logs-are)
- [NotReady nodes](#notready-nodes)
- [API server down](#api-server-down)
- [etcd problems](#etcd-problems)
- [kubelet down](#kubelet-down)
- [Certificate expiry](#certificate-expiry)
- [crictl: inspecting the runtime directly](#crictl-inspecting-the-runtime-directly)
- [Node add / remove](#node-add--remove)
- [Capacity planning basics](#capacity-planning-basics)

## Triage order

Work from the outside in:
1. `kubectl get nodes` — any `NotReady`? `kubectl get pods -A` — control-plane pods healthy?
2. `kubectl describe node <bad-node>` — read **Conditions** (MemoryPressure, DiskPressure,
   PIDPressure, Ready) and **Events**.
3. `kubectl get --raw='/readyz?verbose'` and `/livez?verbose` — which API server check fails.
4. SSH to the affected node → `systemctl status kubelet`, `journalctl -u kubelet -e`.
5. Control-plane component logs (see below).
6. If kubelet/API path is broken, drop to **crictl** on the node to see the runtime truth.

## Where the logs are

**Control-plane nodes** (kubeadm runs these as static Pods, so prefer `kubectl logs` when the API is
up, else the container logs on disk):
```bash
kubectl -n kube-system logs kube-apiserver-<node>
kubectl -n kube-system logs kube-controller-manager-<node>
kubectl -n kube-system logs kube-scheduler-<node>
kubectl -n kube-system logs etcd-<node>
# On disk (API down): /var/log/pods/... and /var/log/containers/...
# Some setups: /var/log/kube-apiserver.log, kube-scheduler.log, kube-controller-manager.log
```
**Worker nodes:** `journalctl -u kubelet`, `journalctl -u kube-proxy` (or
`/var/log/kubelet.log`, `/var/log/kube-proxy.log`). Cluster-wide dump for offline analysis:
```bash
kubectl cluster-info dump --output-directory=/tmp/clusterdump --all-namespaces
```

## NotReady nodes

A node reports `NotReady` when its kubelet stops heartbeating or a node condition fires. Check, in
order:
- **kubelet alive?** `systemctl status kubelet` on the node. Restart: `systemctl restart kubelet`.
- **Logs** `journalctl -u kubelet -e` — common causes: container runtime down, swap re-enabled,
  expired/invalid certs, full disk (`DiskPressure`), wrong cgroup driver, CNI not installed/healthy.
- **Runtime up?** `systemctl status containerd` (or crio); `crictl info`.
- **CNI present?** A fresh cluster's nodes stay `NotReady` until a CNI is applied. Check CNI pods:
  `kubectl get pods -n kube-system` (calico/cilium/flannel).
- **Resource pressure?** `kubectl describe node` Conditions + `kubectl top node`.
- After fixing, the node returns to `Ready` automatically. Pods that were on a `NotReady` node aren't
  rescheduled until the node is marked failed or you `kubectl delete pod` them (if not managed by a
  controller that will recreate elsewhere).

## API server down

If `kubectl` returns connection refused / TLS errors:
- Hitting the right endpoint? `kubectl config view --minify` → check `server:`; confirm the LB/FQDN
  resolves and 6443 is reachable.
- On a control-plane node: `crictl ps -a | grep apiserver`; `crictl logs <id>`. Common causes:
  etcd unreachable, expired certs, bad flag in the static-Pod manifest, port conflict.
- The static-Pod manifest is `/etc/kubernetes/manifests/kube-apiserver.yaml`; a syntax error there
  stops the kubelet from starting it. Validate and let the kubelet recreate it.
- In HA, one API server down is survivable — the LB routes to the others; fix the unhealthy one.

## etcd problems

- Health: `ETCDCTL_API=3 etcdctl endpoint health --endpoints=https://127.0.0.1:2379 \
  --cacert=.../ca.crt --cert=.../server.crt --key=.../server.key`
- Members & leader: `etcdctl member list` / `etcdctl endpoint status --write-out=table`.
- **Lost quorum** (majority of members down) → cluster is read-only/unavailable. Recover members or
  restore from snapshot (see etcd-backup-restore.md).
- Disk full or slow disk causes etcd to misbehave; etcd needs low-latency storage.
- Never edit etcd data directly.

## kubelet down

- `systemctl status kubelet`; `journalctl -u kubelet -e`.
- Frequent culprits: swap on (`swapoff -a`), cgroup-driver mismatch between kubelet and containerd
  (both should be `systemd`), bad `/var/lib/kubelet/config.yaml`, CRI socket wrong/missing,
  expired kubelet client cert.
- Restart: `systemctl daemon-reload && systemctl restart kubelet`.
- Simulating failure for testing (kind/lab): `docker exec <node> systemctl stop kubelet` → node goes
  `NotReady`; running Pods keep running but nothing new schedules there.

## Certificate expiry

Symptom: API/kubelet suddenly refuse TLS after ~1 year on a cluster that wasn't upgraded.
```bash
kubeadm certs check-expiration
sudo kubeadm certs renew all
# restart control-plane static pods (move manifests out/in, or reboot kubelet)
```
Regular `kubeadm upgrade apply` renews certs as a side effect — staying current avoids surprise
expiry. See kubeadm-and-ha.md.

## crictl: inspecting the runtime directly

When the kubelet/API path is broken, `crictl` (CRI client) talks straight to containerd/CRI-O. Point
it at the socket once: `export CONTAINER_RUNTIME_ENDPOINT=unix:///run/containerd/containerd.sock`
(or `/var/run/crio/crio.sock`), or use `crictl --runtime-endpoint`.
```bash
crictl pods                 # list pod sandboxes the runtime knows about
crictl ps -a                # all containers (running + exited)
crictl logs <container-id>  # logs even when kubectl can't reach the API
crictl inspect <container-id>
crictl images               # images pulled on this node
crictl info                 # runtime status & config
```
Useful for confirming "is the container actually running on this node?" independent of what the API
server thinks.

## Node add / remove

- **Add a worker (kubeadm):** prep prerequisites, then run a fresh join command:
  `kubeadm token create --print-join-command` on a control-plane node → run on the new node.
- **Gracefully remove a node:**
  ```bash
  kubectl drain <node> --ignore-daemonsets --delete-emptydir-data
  kubectl delete node <node>
  # on the node itself: sudo kubeadm reset && clean up CNI/iptables
  ```
- **Replace a dead control-plane node:** also remove its stale etcd member
  (`etcdctl member remove <id>`) before joining the replacement.

## Capacity planning basics

- Right-size node pools by workload class: **standard** (web/middleware), **memory-intensive**
  (in-memory DBs, large JVMs), **CPU-intensive**, **special** (GPU/high-bandwidth NIC). Use distinct
  node pools/labels and schedule with selectors/taints (scheduling itself → kubernetes-autoscaling-
  scheduling).
- A Pod stays **Pending** when no node can satisfy its CPU/memory **requests** — add capacity or
  lower requests. `kubectl describe pod` Events say "Insufficient cpu/memory".
- Plan headroom: control-plane components, the CNI, monitoring/logging agents (DaemonSets) all
  consume node resources before your apps do.
- `metrics-server` enables `kubectl top nodes/pods`; for real observability wire Prometheus + Grafana
  and ship logs (node logging agent reading `/var/log`, or a per-Pod sidecar) to a central platform.
- Plan for failure domains: spread across zones/data centers (hot/hot replication preferred), avoid a
  single point of failure, keep distance between sites.
