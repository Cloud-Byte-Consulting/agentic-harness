# NUMA-Aware Scheduling with Volcano

NUMA (Non-Uniform Memory Access) systems group memory and CPU cores into physical "sockets". Accessing local socket memory is extremely fast; crossing sockets via the interconnect bus (QPI/UPI) introduces latency and memory bottlenecks. This reference guide details Volcano's NUMA-aware scheduling stack.

---

## 1. Prerequisites: Kubelet Configuration

For NUMA scheduling to function, the host's **Kubelet** must be configured to manage CPU and topology boundaries statically. By default, Kubelet's CPU Manager is set to `none`, which allows pods to float across any CPU socket.

### `/var/lib/kubelet/config.yaml`
On all high-performance NUMA worker nodes, apply:
```yaml
cpuManagerPolicy: "static"
topologyManagerPolicy: "single-numa-node" # Rejects pods that cannot be aligned to 1 NUMA node
# Reserved CPU cores for system/kubelet tasks (important, must be excluded from static allocation)
reservedSystemCPUs: "0,1"
```

After modifying the configuration, clear Kubelet's state and restart:
```bash
sudo systemctl stop kubelet
sudo rm -f /var/lib/kubelet/cpu_manager_state
sudo systemctl start kubelet
```

---

## 2. Numatopology Discovery (`nodeinfo.volcano.sh/v1alpha1`)

Volcano deploys the `volcano-agent` daemonset to discover the underlying physical socket structure of nodes and report them as `Numatopology` custom resources.

```yaml
apiVersion: nodeinfo.volcano.sh/v1alpha1
kind: Numatopology
metadata:
  name: gpu-worker-01
spec:
  policies:
    CPUManagerPolicy: "static"
    TopologyManagerPolicy: "single-numa-node"
  resReserved:
    cpu: "2"
    memory: "4Gi"
  numares:
    cpu:
      allocatable: "48"
      capacity: 48
    memory:
      allocatable: "256Gi"
      capacity: 256Gi
  cpuDetail:
    "0":
      numa: 0
      socket: 0
      core: 0
    "1":
      numa: 0
      socket: 0
      core: 1
    "24":
      numa: 1
      socket: 1
      core: 12
    "25":
      numa: 1
      socket: 1
      core: 13
```

- `cpuDetail`: Maps every logical core thread ID (OS index) to its physical `numa` ID and physical `socket`.
- `resReserved`: Tells the scheduler which core IDs are reserved by Kubelet and must not be used for high-performance scheduling.

---

## 3. Configuring the Scheduler

Enable the `numaaware` plugin in your `volcano-scheduler.conf` configmap under the `allocate` actions tier.

```yaml
actions: "enqueue, allocate, backfill"
tiers:
  - plugins:
    - name: priority
    - name: gang
  - plugins:
    - name: numaaware          # Enabled for NUMA scoring
    - name: predicates
    - name: nodeorder
```

---

## 4. Deploying a NUMA-Aligned Job

To request NUMA-aware allocation, pods must belong to the `Guaranteed` Quality of Service (QoS) class (where `requests == limits` for both CPU and Memory) and specify the Volcano topology policy.

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: numa-hpc-job
spec:
  minAvailable: 1
  schedulerName: volcano
  tasks:
    - name: solver
      replicas: 1
      topologyPolicy: restricted # restricted, best-effort, single-numa-node
      template:
        spec:
          containers:
            - name: compute
              image: hpc-solver:latest
              resources:
                requests:
                  cpu: "8"            # Must request integer CPUs
                  memory: "32Gi"
                limits:
                  cpu: "8"            # Limits must match requests (Guaranteed QoS)
                  memory: "32Gi"
```

### Topology Policies:
- `restricted`: Standard strict alignment. The scheduler searches for a node that can fit the CPU/Memory entirely within a single NUMA socket. If unavailable, it falls back to a node where they fit on multiple sockets without failing.
- `single-numa-node`: Extreme strict alignment. Rejects scheduling if the container's requests cannot be completely satisfied within a single NUMA socket.

---

## 5. Verification & Troubleshooting

Check if cores are bound correctly on the node:
```bash
# Describe the numatopology of the node to see depleted core capacity
kubectl describe numatopology gpu-worker-01

# Run in-container to check assigned CPU masks
kubectl exec -it <pod-name> -- cat /sys/fs/cgroup/cpuset/cpuset.cpus
# If successfully NUMA-aligned, the returned core IDs will all reside under the same NUMA node index.
```
