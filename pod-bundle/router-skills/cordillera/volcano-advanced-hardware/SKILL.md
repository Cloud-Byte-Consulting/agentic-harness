---
name: volcano-advanced-hardware
description: >-
  Schedule high-performance workloads (AI/ML, HPC) using hardware-level optimizations in Volcano.
  Use for NUMA awareness (Numatopology), fractional GPU and device sharing (deviceshare),
  network-topology-aware scheduling (HyperNode, RoCE/UFM, tiering), and colocation/oversubscription QoS.
  Trigger when users discuss GPU sharing, vGPU, node NUMA topology, high-throughput network alignment (RoCE),
  HyperNodes, or running online/offline workloads concurrently on Volcano.
---

# Volcano Advanced Clusters & Hardware Scheduling

This skill equips Claude with deep expertise in optimizing Volcano scheduling at the physical hardware layer. It covers node-level CPU socket alignment (NUMA), fractional and shared GPU architectures, high-performance network-aware topologies (HyperNodes), and cluster colocation strategies that maximize hardware utilization and efficiency.

## When to use this skill

- Designing or reviewing topologies for latency-critical HPC or large-scale AI/ML workloads (LLM pre-training, deep learning).
- Writing and configuring `Numatopology` resources for CPU/Memory socket pin-point scheduling.
- Configuring `HyperNode` resources for network-aware, low-latency pod grouping (RoCE/InfiniBand fabrics).
- Setting up shared or fractional GPU allocations (vGPUs, vNPUs) via the `deviceshare` plugin.
- Authoring `ColocationConfiguration` policies to run high-priority (online) and low-priority (offline batch) workloads on the same nodes safely.
- Troubleshooting memory latency spikes, inter-socket bus bottlenecks, network switch hops, or GPU resource contention.

## Core concepts

- **NUMA (Non-Uniform Memory Access)**: Modern multi-socket servers have memory controllers dedicated to specific CPU sockets. Accessing local socket memory is extremely fast; accessing another socket's memory over the interconnect bus (UPI/QPI) introduces substantial latency. Volcano uses `Numatopology` CRDs populated by `volcano-agent` and the `numaaware` plugin to align pod CPU and memory allocations strictly to the same NUMA node.
- **Fractional GPU / Device Sharing**: Standard Kubernetes allocates GPUs on an integer basis (1 Pod gets 1 whole GPU). Volcano's `deviceshare` plugin allows fractional allocation (e.g., `nvidia.com/gpu: 0.25`), enabling smaller ML inference or test tasks to share a single GPU. It tracks GPU memory and cores, preventing out-of-memory (OOM) failures.
- **HyperNode**: A custom resource representing a physical network topology grouping (e.g., a rack, an InfiniBand leaf switch, or a cloud region zone). By constructing a hierarchical tree of HyperNodes, Volcano can place tightly-coupled distributed tasks (like MPI or deep learning) within the same low-latency network switch boundaries.
- **Colocation & Oversubscription**: Running batch (offline) jobs alongside latency-critical microservices (online) optimizes hardware utilization. Volcano's `volcano-agent` and `ColocationConfiguration` monitor real-time node pressure. If the online service spikes, the agent dynamically throttles, suspends, or evicts the offline batch workloads to maintain online SLAs.

## Workflow / how to approach tasks

1. **NUMA Alignment**: Ensure the `volcano-agent` daemonset is deployed to discover node NUMA structures. Enable the `numaaware` plugin in the scheduler configuration. Specify `topologyPolicy: restricted` on tasks requiring strict socket alignment.
2. **GPU Sharing**: Enable the `deviceshare` plugin in `volcano-scheduler.conf`. Set resource requests using decimals (e.g., `volcano.sh/vgpu-memory: "4Gi"`, `volcano.sh/vgpu-cores: "30"`).
3. **Network Topology Tree**: Map the physical data center network into a hierarchy of `HyperNode` manifests. Specify the `networkTopology` constraint in your Job spec to dictate hard or soft placement constraints within specific switch tiers.
4. **Colocation Policies**: Deploy the `ColocationConfiguration` to define memory and CPU eviction thresholds (e.g., trigger throttling when CPU usage exceeds 80%, evict batch pods if memory usage exceeds 90%).

## Common pitfalls & anti-patterns

- **Mismatched Kubelet CPU Manager & Volcano NUMA policies**: If the Kubelet CPU Manager is set to `none` (default), it will ignore socket boundaries, rendering Volcano's `numaaware` scheduler decisions useless. Always set Kubelet's CPU Manager to `static`.
- **GPU Oversubscription without Memory Safeguards**: Specifying fractional GPU usage without defining `vgpu-memory` limits allows pods to exceed real VRAM capacity, triggering silent GPU crashes (CUDA Out of Memory) across all sharing pods.
- **Oversimplifying HyperNode Trees**: Creating a flat `HyperNode` list with no nested members disables tier-based routing, forcing the scheduler to fallback to default (random node) placement. Always design a multi-tier network tree (`Tier 1: Rack, Tier 2: Switch`).
- **Aggressive Colocation Eviction Thresholds**: Setting colocation trigger levels too low (e.g., 50% CPU pressure) causes constant, disruptive restarts of batch jobs, ruining training runs. Set trigger thresholds based on real 95th-percentile microservice utilization.

## Reference files

- `references/numa-aware-scheduling.md` — `Numatopology` specs, Kubelet config prerequisites, the `numaaware` plugin, and verification.
- `references/gpu-and-device-sharing.md` — Fractional GPU allocations, the `deviceshare` plugin, vGPUs, and custom device configurations.
- `references/network-topology-aware.md` — `HyperNode` definitions, multi-tier network trees, `networkTopology` spec modes, and RoCE/UFM low-latency fabrics.
- `references/colocation-and-qos.md` — `ColocationConfiguration` YAML, online/offline resource oversubscription, and the `volcano-agent` throttling daemon.
