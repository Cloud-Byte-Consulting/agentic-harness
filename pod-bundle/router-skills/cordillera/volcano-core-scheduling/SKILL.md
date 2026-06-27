---
name: volcano-core-scheduling
description: >-
  Design, configure, and troubleshoot Volcano batch scheduling workloads, queues, and PodGroups on Kubernetes.
  Use for gang/atomic scheduling (minAvailable, minMember), dominant resource fairness (DRF), preemption/reclamation,
  hierarchical queue configurations, scheduler plugin setup, and vcctl command-line management.
  Trigger whenever a user mentions batch scheduling, ML distributed training (PyTorch, TensorFlow, MPI),
  gang scheduling, queue capability/reclaim, or scheduler actions (enqueue, allocate, preempt, reclaim).
---

# Volcano Core Batch Scheduling

This skill equips Claude with deep, production-grade expertise in Volcano — the CNCF-hosted Kubernetes-native batch scheduling engine. It covers the core batch scheduling mechanics, hierarchical queues, PodGroups, dominant resource fairness (DRF), and gang scheduling that make Volcano the industry standard for running AI/ML training, big data pipelines, and HPC workloads on Kubernetes.

## When to use this skill

- Authoring or reviewing Volcano core manifests (`Job`, `Queue`, `PodGroup`).
- Configuring the Volcano scheduler via `volcano-scheduler.conf` configmaps (defining Actions, Plugins, and Tiers).
- Planning or implementing multi-tenant Kubernetes clusters with queue-based fair-share resource allocation.
- Troubleshooting jobs stuck `Pending` (determining if it is a gang constraint, queue saturation, or PodGroup mismatch).
- Tuning scheduler throughput, API request limits (QPS/burst), or scheduling cycles for large-scale clusters.
- Working with the `vcctl` CLI for job lifecycle control (suspending, resuming, deleting).

## Core concepts

- **Gang (Atomic) Scheduling**: AI/ML distributed training requires all replicas (e.g., parameter servers and workers) to start together. Standard `kube-scheduler` schedules pods individually, leading to deadlock (where Job A holds 2 of 4 GPUs, Job B holds 2, and neither can proceed). Volcano guarantees **all-or-nothing** scheduling via `minAvailable` (specifies the minimum schedulable pods before binding).
- **Dominant Resource Fairness (DRF)**: In a multi-resource environment (CPU, Memory, GPU), traditional max-min fairness is insufficient. DRF equalizes the *share of the dominant resource* allocated to each queue/tenant. The dominant resource of a user is the resource type they have the largest share of relative to cluster capacity.
- **Hierarchical Queues**: Organize cluster capacity along organizational boundaries. Teams have a tree of sub-queues (e.g., `root/eng/prod`, `root/eng/dev`). Excess capacity in one queue is shared and can be **reclaimed** based on priority, weight, and capability.
- **PodGroup**: The scheduling abstraction that wraps a group of pods belonging to the same batch job. Standard Pods can opt-in to Volcano scheduling by specifying `schedulerName: volcano` and adding the annotation `scheduling.volcano.sh/group-name: <podgroup-name>`.
- **Actions vs. Plugins**: Volcano decouples scheduling logic. **Actions** define the phases of a single scheduling cycle (run in sequence: `enqueue`, `allocate`, `preempt`, `reclaim`, `backfill`). **Plugins** implement the decision-making rules (e.g., `gang` validates minAvailable; `drf` scores job priority based on fair-share; `predicates` runs standard node fit tests).

## Workflow / how to approach tasks

1. **Verify the scheduling cycle actions**: Always confirm that the scheduler configuration (`volcano-scheduler.conf`) has the appropriate actions enabled. For multi-tenant preemption and reclamation, the actions list must include `preempt` and `reclaim` in sequence: `enqueue, allocate, preempt, reclaim, backfill`.
2. **Define the queue topology**: When setting up hierarchical queues, use annotations to declare the path (e.g., `volcano.sh/hierarchy: root/team/subteam`) and set appropriate values for:
   - `weight`: Relative share denominator.
   - `capability`: Hard upper resource limit.
   - `deserved`: Soft entitlement limit (reclaim trigger floor).
3. **Configure Gang constraints**: For any distributed batch job, define `minAvailable` at the Job level. Ensure it matches the sum of critical task replicas (e.g., 1 master + 3 workers = 4).
4. **Tune performance**: On large clusters (>100 nodes), tune the scheduler by increasing api QPS/burst, adjusting the scheduling cycle period (default 1s), and increasing parallel worker threads.

## Common pitfalls & anti-patterns

- **Omitting `minAvailable` or setting it to 1 for multi-pod training**: Disables gang scheduling, risking cluster-wide deadlocks under high load.
- **Mismatch between `minAvailable` and Node Capacity**: If `minAvailable` is set to 8 but the largest single node only fits 4, and inter-node scheduling is blocked by topology, the job stays Pending forever. Ensure node groups or autoscalers can satisfy the gang requirement.
- **Strict PDBs on Node-Draining or Preemption**: Setting a `PodDisruptionBudget` that matches the exact replica count prevents Volcano from evicting lower-priority pods to reclaim capacity, blocking high-priority jobs.
- **Excluding `preempt` or `reclaim` Actions in Conf**: If they are omitted from the configuration file, queue prioritization and resource reclamation will not function, regardless of how queues are configured.

## Reference files

- `references/architecture-and-crd.md` — Volcano components (vc-scheduler, vc-controller-manager, vc-webhook-manager, vcctl), CRD specifications for Queue, PodGroup, and Command.
- `references/scheduler-internals.md` — The 6 Actions, 20+ Plugins, scheduler cycle execution timeline, and configuration syntax (`volcano-scheduler.conf`).
- `references/troubleshooting-and-ops.md` — CLI commands, Helm installation, troubleshooting checklist, logs analysis, performance tuning, and Volcano vs Default Scheduler comparison.
