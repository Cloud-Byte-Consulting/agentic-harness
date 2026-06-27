---
name: volcano-ecosystem-workflows
description: >-
  Orchestrate complex distributed training and data workflows on Kubernetes using Volcano's ecosystem.
  Use for JobFlow DAG automation, CronVolcanoJobs, and operators/framework integrations (Kubeflow PyTorch/TF,
  MPI Operator, Apache Spark, Ray, Flink, PaddlePaddle).
  Trigger when users discuss running Spark on Volcano, Kubeflow operators, MPI distributed setups (SSH/headless service),
  JobFlow DAG dependencies, CronVolcanoJob schedules, or Ray cluster integration.
---

# Volcano Ecosystem & Workflow Orchestration

This skill equips Claude with specialized expertise in orchestrating complex, distributed data pipelines and machine learning workloads using the broader Volcano ecosystem. It covers multi-job dependency mapping (JobFlows), scheduled batch jobs (CronVolcanoJobs), and native integrations with standard data and AI runtimes like Kubeflow, Apache Spark, Ray, and MPI.

## When to use this skill

- Designing multi-stage ML pipelines where model training depends on completed data extraction and validation.
- Setting up scheduled, periodic batch jobs that require custom lifecycle and concurrency policies (CronVolcanoJobs).
- Integrating Volcano with standard operators (Kubeflow PyTorchJob/TFJob, Spark Operator, Ray Operator, Argo Workflows).
- Configuring complex MPI training runs requiring passwordless SSH keys and dynamic hostfile DNS setup via job plugins.
- Handling distributed framework-specific environment configurations (setting up RANK, WORLD_SIZE, MASTER_ADDR, and MASTER_PORT).
- Implementing self-healing DAG workflows where tasks trigger conditional downstream rollouts.

## Core concepts

- **JobFlow (`flow.volcano.sh/v1alpha1`)**: Volcano's native workflow DAG (Directed Acyclic Graph) engine. While Argo Workflows operates at the pod/container layer, JobFlow operates at the *Volcano Job layer*. It allows developers to chain complex, multi-task, multi-replica Volcano Jobs together, enforcing dependencies, execution conditions, and automatic rollouts.
- **JobTemplate**: A reusable schema defining a Volcano Job. Under a JobFlow, task nodes reference JobTemplates and define a custom `patch` to modify execution params (like queues, storage mounts, or replica counts) dynamically at runtime.
- **CronVolcanoJob**: A scheduled CRD modeled after standard Kubernetes CronJobs, but running `batch.volcano.sh` Volcano Jobs. It introduces crucial batch safeguards like `concurrencyPolicy: Forbid` (preventing overlapping runs) and custom eviction/restart policy inheritance.
- **Distributed Framework Plugins**: Specialized controllers inside `vc-controller-manager` that augment Volcano Jobs to fit specific framework requirements:
  - `ssh`: Generates passwordless SSH keypairs and mounts them into all pods of an MPI job.
  - `svc`: Creates a headless service matching the job name, and generates host lists (e.g. `/etc/volcano/mpiworker.host`) so masters can locate workers instantly.
  - `pytorch` / `tensorflow`: Automatically resolves and injects master network details and coordination variables.

## Workflow / how to approach tasks

1. **Leverage Native Plugins**: For MPI or Distributed PyTorch jobs, never manage SSH keys or hostfiles manually. Always include the `ssh` and `svc` plugins under the Job's `spec.plugins` block to let Volcano handle host discovery and secure inter-pod comms.
2. **Build Multi-stage Pipelines**: Define your workflows using the `JobFlow` CRD. Specify dependencies using the `dependsOn` blocks.
3. **Opt-in Framework Operators**: When deploying Spark or Ray, specify `schedulerName: volcano` in the operator configurations. Volcano's gang-scheduling and queue capacity management will seamlessly overlay to protect resources and prevent deadlocks.
4. **Schedule Periodic Work**: Use `CronJob` (`batch.volcano.sh`) for regular runs. Set history limits (`successfulJobsHistoryLimit`, `failedJobsHistoryLimit`) to prevent resource clutter.

## Common pitfalls & anti-patterns

- **Overlapping Cron Job Executions**: Leaving `concurrencyPolicy: Allow` on heavy, resource-intensive CronVolcanoJobs can lead to duplicate concurrent runs that saturate queues and block other critical cluster workloads. Always default to `Forbid` or `Replace`.
- **Manually Generating SSH Keys for MPI**: Hand-crafting SSH keys, saving them as K8s Secrets, and writing manual mounting scripts is error-prone, insecure, and unnecessary. Use Volcano's native `ssh` plugin.
- **Spark Operator without Gang Scheduling**: Deploying a Spark job without setting a matching `PodGroup` or `minAvailable` constraint allows driver pods to start up while executors stay Pending, wasting driver resources and potentially starving other jobs.
- **Mismatched JobFlow Dependency Types**: Specifying task dependencies in a JobFlow based on simple completion without checking task exit codes can cause down-stream steps to run using corrupted or missing data from a failed pre-requisite step. Always enforce strict dependency validation.

## Reference files

- `references/framework-integrations.md` — Kubeflow operators (PyTorch, TensorFlow), MPI with SSH/SVC plugins, native Spark on Volcano, and Ray integrations.
- `references/jobflow-and-dag.md` — `JobFlow` and `JobTemplate` CRD details, task dependencies, patch schemas, and pipeline DAG examples.
- `references/cron-volcano-job.md` — `CronJob` specifications, concurrency management, scheduling parameters, and cron pipeline manifests.
