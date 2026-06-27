# JobFlow & DAG Workflows with Volcano

While workflows like Argo or Tekton orchestrate sequences at the individual pod or container layer, Volcano's `JobFlow` engine orchestrates workflows at the **Volcano Job layer**. This is crucial for multi-stage machine learning and big data pipelines where whole distributed task groups must be chained in a dependency graph.

---

## 1. JobFlow Architecture

A `JobFlow` consists of **JobTemplates** (reusable Volcano Job specifications) and a **JobFlow DAG** (specifying the execution steps, patch rules, and dependency logic).

```
 ┌───────────────┐        ┌─────────────────┐
 │  etl-job      │ ──────►│ data-validation │
 │  (Task A)     │        │ (Task B)        │
 └───────────────┘        └────────┬────────┘
                                   │
                                   ▼
                          ┌─────────────────┐
                          │ model-training  │
                          │ (Task C)        │
                          └─────────────────┘
```

---

## 2. Reusable Templates: `JobTemplate`

Define the structure of a job, excluding dynamic parameters like specific queue mappings or storage targets:

```yaml
apiVersion: flow.volcano.sh/v1alpha1
kind: JobTemplate
metadata:
  name: trainer-template
  namespace: ml-pipeline
spec:
  minAvailable: 2
  schedulerName: volcano
  tasks:
    - name: worker
      replicas: 2
      template:
        spec:
          containers:
            - name: training-container
              image: training-image:latest
              resources:
                requests:
                  cpu: "4"
                  memory: "16Gi"
```

---

## 3. The `JobFlow` DAG (`flow.volcano.sh/v1alpha1`)

Define the execution pipeline, connecting steps via `dependsOn` arrays and applying custom runtime patches.

```yaml
apiVersion: flow.volcano.sh/v1alpha1
kind: JobFlow
metadata:
  name: end-to-end-ml-flow
  namespace: ml-pipeline
spec:
  # The global queue where all workflow stages run unless patched
  queue: pipeline-queue
  flows:
    # --- Step 1: Data Extraction ---
    - name: etl-stage
      jobTemplateName: etl-template
      patch:
        # Patch execution details dynamically for this run
        jobSpec:
          queue: etl-low-priority-queue

    # --- Step 2: Data Validation (runs after Step 1 finishes) ---
    - name: validation-stage
      jobTemplateName: validate-template
      dependsOn:
        targets:
          - etl-stage
        # Run only if the ETL stage successfully completed
        probe:
          taskStatusList:
            - taskName: extractor
              phase: Completed

    # --- Step 3: Model Training (runs after Step 2 completes) ---
    - name: training-stage
      jobTemplateName: trainer-template
      dependsOn:
        targets:
          - validation-stage
      patch:
        jobSpec:
          queue: gpu-training-queue
```

---

## 4. Execution & Status Tracking

The `JobFlow` controller in `vc-controller-manager` actively watches flow states:

```bash
# Query active workflows
kubectl get jobflow -n ml-pipeline

# View detailed execution path and completed steps
kubectl describe jobflow end-to-end-ml-flow -n ml-pipeline
```

- `status.runningJobs`: Shows which stage in the DAG is currently running.
- `status.completedJobs`: Lists successfully processed stages.
- `status.pendingJobs`: Stages waiting for pre-requisite dependencies to finish.
- `status.conditions`: Captures transitions, timing, and errors.
