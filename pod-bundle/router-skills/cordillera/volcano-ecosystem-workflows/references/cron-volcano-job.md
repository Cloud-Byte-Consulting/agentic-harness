# Scheduled Jobs with CronVolcanoJob

To run periodic data extraction, model retraining, or nightly report generation with Volcano batch features (like queue capability routing and gang constraints), use the `CronJob` resource in the `batch.volcano.sh` group.

---

## 1. Why `CronJob` (`batch.volcano.sh`) vs standard K8s CronJob

- **Queue Alignment**: Standard K8s CronJobs create default Pods that cannot be placed into Volcano's hierarchical queue structure. CronVolcanoJobs create native `batch.volcano.sh/v1alpha1` `Jobs`, adhering to team fair-share boundaries.
- **Co-Scheduling Support**: CronVolcanoJobs can run gang-constrained tasks. For example, a nightly distributed simulation run can be scheduled to run *only* if all workers are available.
- **Advanced Lifecycle Actions**: Incorporates Volcano's rich event/action mapping (e.g., restart job on eviction, abort on pod failure).

---

## 2. CronVolcanoJob Specification

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: CronJob
metadata:
  name: nightly-model-retrain
  namespace: ml-pipeline
spec:
  schedule: "0 2 * * *"         # Run every day at 2:00 AM
  concurrencyPolicy: Forbid     # Forbid: skip if previous is still running; Allow; Replace
  startingDeadlineSeconds: 300  # Max time allowed to start if schedule is missed
  suspend: false                # Toggle to true to temporarily pause scheduling
  successfulJobsHistoryLimit: 3 # Retain last 3 successful jobs
  failedJobsHistoryLimit: 1     # Retain last failed job for debugging
  jobTemplate:                  # The standard Volcano Job Spec template
    spec:
      minAvailable: 3           # Requires all 3 replicas to start together
      queue: nightly-batch-queue
      schedulerName: volcano
      tasks:
        - name: trainer
          replicas: 3
          template:
            spec:
              containers:
                - name: train
                  image: pytorch-retrain:latest
                  command: ["python", "retrain.py"]
                  resources:
                    requests:
                      cpu: "4"
                      memory: "16Gi"
              restartPolicy: Never
      policies:
        - event: PodEvicted
          action: RestartJob
```

---

## 3. Concurrency Policies Explained

- `Forbid`: If a nightly retraining job takes longer than 24 hours, the scheduler will **skip** the next 2:00 AM run. This protects the queue from becoming saturated with duplicate running pipelines.
- `Replace`: If a new schedule is triggered, the scheduler will **delete** the currently running job and spawn a fresh replacement, useful for sliding-window data aggregators.
- `Allow`: Spawns runs concurrently. Avoid using on resource-heavy ML clusters.

---

## 4. Monitoring Scheduled Jobs

```bash
# List all cron jobs
kubectl get cronjobs.batch.volcano.sh -n ml-pipeline

# Check history and latest execution times
kubectl describe cronjob.batch.volcano.sh nightly-model-retrain -n ml-pipeline
```
- `status.active`: Lists the API resource IDs of currently running Volcano Job instances created by the Cron job.
- `status.lastScheduleTime`: Timestamp of the last run trigger.
- `status.lastSuccessfulTime`: Timestamp of the last successful completion.
