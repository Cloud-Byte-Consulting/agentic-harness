# Apache Spark on Kubernetes

How Spark runs on K8s, the Spark Operator and `SparkApplication` CRD, `spark-submit` with a `k8s://` master, resource tuning, object-store access, and debugging.

## Contents
- Spark execution model on K8s
- Two ways to run: spark-submit vs the Spark Operator
- Installing the Spark Operator
- The SparkApplication CRD (full YAML)
- ScheduledSparkApplication
- Object storage (S3 / s3a) configuration
- Resource tuning & dynamic allocation
- Spark internals that affect performance (lazy eval, narrow/wide, joins)
- Submitting, monitoring, debugging

## Spark execution model on K8s

A Spark application is one **driver** plus N **executors**. The driver hosts the `SparkSession`, builds the execution plan, and schedules **tasks**; executors are worker processes that run tasks and hold data in memory. A *job* splits into *stages* (separated by shuffles); each stage is a set of parallel *tasks*, one per data partition, run across executor *slots* (cores).

On Kubernetes, **K8s itself is the cluster manager** (replacing YARN/Mesos/standalone). The driver and every executor each run as their **own pod**. The driver pod talks to the API server to request executor pods. Two submission modes:
- **`cluster` mode** (the norm on K8s) — the driver runs *inside* the cluster as a pod. Use this for the Operator and production.
- **`client` mode** — the driver runs where you launched it (e.g. a notebook/JupyterLab); executors still run as pods. Handy for interactive dev.

## Two ways to run: spark-submit vs the Spark Operator

**Raw `spark-submit` with a `k8s://` master** — no operator needed; Spark's own k8s backend creates the pods. Good for CI, ad-hoc runs, or environments where you can't install an operator:

```bash
spark-submit \
  --master k8s://https://<API_SERVER>:6443 \
  --deploy-mode cluster \
  --name titanic-job \
  --class org.apache.spark.examples.SparkPi \
  --conf spark.kubernetes.namespace=spark-jobs \
  --conf spark.kubernetes.container.image=apache/spark:3.5.1 \
  --conf spark.kubernetes.authenticate.driver.serviceAccountName=spark \
  --conf spark.executor.instances=3 \
  --conf spark.executor.cores=2 \
  --conf spark.executor.memory=4g \
  --conf spark.driver.memory=2g \
  local:///opt/spark/work-dir/job.py
```

`local://` means the path is inside the container image. The `spark` ServiceAccount needs RBAC to create pods (see Airflow reference for the equivalent binding).

**The Spark Operator (`SparkApplication` CRD)** — declarative, GitOps-friendly, with scheduling, retries, metrics, and lifecycle handling. **Prefer this for anything recurring or production.** You define the job once as YAML and `kubectl apply` it.

## Installing the Spark Operator

The maintained project is **`kubeflow/spark-operator`** (formerly GoogleCloudPlatform). Install via Helm:

```bash
helm repo add spark-operator https://kubeflow.github.io/spark-operator
helm repo update

kubectl create namespace spark-operator

helm install spark-operator spark-operator/spark-operator \
  --namespace spark-operator \
  --set webhook.enable=true \
  --set "spark.jobNamespaces={spark-jobs}"
```

The mutating webhook (`webhook.enable=true`) is required for the operator to inject volumes, env, and pod customizations into driver/executor pods. Verify:

```bash
kubectl get pods -n spark-operator
# spark-operator-controller-...   1/1   Running
# spark-operator-webhook-...      1/1   Running
```

Each namespace that runs Spark jobs needs a `spark` ServiceAccount with pod-management RBAC (the chart can create it, or create it yourself):

```bash
kubectl create serviceaccount spark -n spark-jobs
kubectl create clusterrolebinding spark-role \
  --clusterrole=edit --serviceaccount=spark-jobs:spark
```

> The CRD API group is `sparkoperator.k8s.io/v1beta2`. If you inherit an older `GoogleCloudPlatform/spark-on-k8s-operator` install, the kind and group are the same but the Helm repo/release differs — check `kubectl get crd | grep spark`.

## The SparkApplication CRD

A complete, applyable PySpark job reading from and writing to S3:

```yaml
apiVersion: sparkoperator.k8s.io/v1beta2
kind: SparkApplication
metadata:
  name: imdb-tsv-to-parquet
  namespace: spark-jobs
spec:
  type: Python
  pythonVersion: "3"
  mode: cluster
  image: "myrepo/spark-aws:3.5.1"     # image with hadoop-aws + your deps baked in
  imagePullPolicy: Always
  mainApplicationFile: "s3a://my-bucket/spark-jobs/tsv_to_parquet.py"
  sparkVersion: "3.5.1"
  restartPolicy:
    type: Never                        # batch job: run once
  # Spark + Hadoop config:
  sparkConf:
    "spark.kubernetes.allocation.batch.size": "10"
    "spark.jars.ivy": "/tmp/ivy"       # cache deps on the mounted volume
  hadoopConf:
    "fs.s3a.impl": "org.apache.hadoop.fs.s3a.S3AFileSystem"
    "fs.s3a.aws.credentials.provider": "com.amazonaws.auth.EnvironmentVariableCredentialsProvider"
  volumes:
    - name: ivy
      emptyDir: {}
  driver:
    cores: 1
    memory: "2g"
    serviceAccount: spark
    labels:
      app: imdb-batch
    env:
      - name: AWS_ACCESS_KEY_ID
        valueFrom:
          secretKeyRef: { name: aws-credentials, key: aws_access_key_id }
      - name: AWS_SECRET_ACCESS_KEY
        valueFrom:
          secretKeyRef: { name: aws-credentials, key: aws_secret_access_key }
    volumeMounts:
      - name: ivy
        mountPath: /tmp
  executor:
    instances: 3
    cores: 2
    memory: "4g"
    labels:
      app: imdb-batch
    env:
      - name: AWS_ACCESS_KEY_ID
        valueFrom:
          secretKeyRef: { name: aws-credentials, key: aws_access_key_id }
      - name: AWS_SECRET_ACCESS_KEY
        valueFrom:
          secretKeyRef: { name: aws-credentials, key: aws_secret_access_key }
    volumeMounts:
      - name: ivy
        mountPath: /tmp
```

Create the credentials secret first:

```bash
kubectl create secret generic aws-credentials \
  --from-literal=aws_access_key_id=<KEY> \
  --from-literal=aws_secret_access_key=<SECRET> \
  -n spark-jobs
```

> **Production credentials:** prefer **IRSA (EKS)** / **Workload Identity (GKE)** / **AAD Workload Identity (AKS)** — attach a cloud IAM role to the `spark` ServiceAccount and drop the static keys entirely. The secret-env approach above is fine for dev/training.

Key spec fields:
- `type`: `Python` | `Scala` | `Java` | `R`.
- `mode`: `cluster` (almost always).
- `image`: must contain Spark **and** any connectors your job needs (e.g. `hadoop-aws`, `spark-sql-kafka`). Bake them in rather than fetching at runtime for reliability.
- `mainApplicationFile`: `s3a://…`, `local://…` (in-image), or `http(s)://…`.
- `restartPolicy.type`: `Never` (batch) or `Always`/`OnFailure` with `restartPolicy.onFailureRetries` for streaming/resilient jobs.
- `deps.jars` / `deps.packages`: add Maven coordinates if not baked into the image, e.g. `org.apache.spark:spark-sql-kafka-0-10_2.12:3.5.1`.

The driver writes its name to `.status`; `restartPolicy.type: Never` means K8s won't let you resubmit the same name — delete first to rerun.

## ScheduledSparkApplication

For cron-style recurring jobs, use `ScheduledSparkApplication` (the operator manages the schedule; no Airflow needed for simple cases):

```yaml
apiVersion: sparkoperator.k8s.io/v1beta2
kind: ScheduledSparkApplication
metadata:
  name: nightly-aggregation
  namespace: spark-jobs
spec:
  schedule: "0 2 * * *"               # 02:00 daily (cron)
  concurrencyPolicy: Forbid           # don't overlap runs
  successfulRunHistoryLimit: 3
  failedRunHistoryLimit: 3
  template:                            # a full SparkApplication spec
    type: Python
    mode: cluster
    image: "myrepo/spark-aws:3.5.1"
    mainApplicationFile: "s3a://my-bucket/jobs/agg.py"
    sparkVersion: "3.5.1"
    restartPolicy:
      type: Never
    driver: { cores: 1, memory: "2g", serviceAccount: spark }
    executor: { instances: 2, cores: 2, memory: "4g" }
```

## Object storage (S3 / s3a) configuration

Spark talks to S3-compatible stores through the **`s3a`** Hadoop connector. Set in `hadoopConf` (operator) or `spark.hadoop.*` confs (in code / `spark-submit`):

```python
conf = (
    SparkConf()
    .set("spark.hadoop.fs.s3a.impl", "org.apache.hadoop.fs.s3a.S3AFileSystem")
    .set("spark.hadoop.fs.s3a.aws.credentials.provider",
         "com.amazonaws.auth.EnvironmentVariableCredentialsProvider")
    .set("spark.hadoop.fs.s3a.fast.upload", "true")
    # For MinIO / non-AWS S3:
    # .set("spark.hadoop.fs.s3a.endpoint", "http://minio.minio.svc:9000")
    # .set("spark.hadoop.fs.s3a.path.style.access", "true")
)
```

- Use a Spark/Hadoop build where `hadoop-aws` and the matching `aws-java-sdk` versions line up — mismatches cause `NoSuchMethodError`. Bake them into the image (e.g. Hadoop 3.3.x with Spark 3.5.x), don't mix arbitrary versions.
- Read/write **Parquet** (columnar, splittable, compressed) for lake storage; avoid CSV/TSV for anything large.
- Use `s3a://` paths everywhere (not `s3://` or `s3n://`).
- For MinIO: set `fs.s3a.endpoint` and `fs.s3a.path.style.access=true`.

## Resource tuning & dynamic allocation

**Static sizing** (predictable, simplest): set `executor.instances`, `executor.cores`, `executor.memory`. Total parallelism = `instances × cores` task slots.

**Cap maximum resources** so a job can't blow the cluster: `spark.cores.max`, plus per-pod `cores`/`memory` and a ResourceQuota on the namespace.

**Dynamic allocation** scales executors with workload — Spark adds executors when there's a task backlog and removes idle ones. On K8s this needs **shuffle tracking** (no external shuffle service required on modern Spark):

```yaml
sparkConf:
  "spark.dynamicAllocation.enabled": "true"
  "spark.dynamicAllocation.shuffleTracking.enabled": "true"
  "spark.dynamicAllocation.minExecutors": "1"
  "spark.dynamicAllocation.maxExecutors": "10"
  "spark.dynamicAllocation.initialExecutors": "2"
```

Use dynamic allocation for variable/interactive workloads; static sizing for steady, well-understood batch. Memory tuning: leave headroom — executor pod memory = `spark.executor.memory` + `spark.executor.memoryOverhead` (default ~10%, raise for PySpark/UDF-heavy jobs that use off-heap/Python memory). If pods get OOMKilled, raise `memoryOverhead` before raising `memory`.

## Spark internals that affect performance

These determine whether a job is fast or a shuffle disaster — author jobs with them in mind.

**Lazy evaluation.** *Transformations* (`select`, `filter`, `groupBy`, `join`, `orderBy`) build a plan but execute nothing. *Actions* (`show`, `count`, `collect`, `write`) trigger execution. Spark optimizes the whole transformation chain (Catalyst optimizer: predicate pushdown, projection pruning, join selection) only when an action fires.

**Narrow vs wide transformations.**
- *Narrow* (`map`, `filter`, `select`, `where`) act per-partition with no data movement — cheap, parallel.
- *Wide* (`groupBy`, `join`, `orderBy`, window functions) require a **shuffle** — moving/repartitioning data across the network. Shuffles are the dominant cost.
- **Filter and project before you join/aggregate.** Shrinking the data first shrinks the shuffle. This is the single highest-leverage Spark optimization.

**Join strategies** (Spark picks automatically; you can hint):
- **Broadcast hash join** — small side is broadcast to every executor; no shuffle of the big side. Fastest *when one side fits in executor memory*. Force with `f.broadcast(small_df)`.
- **Sort-merge join** — both sides shuffled and sorted on the key, then merged. Spark's default for two large tables.
- **Shuffle hash join** — both sides hash-partitioned, build+probe per partition; no sort. Hint with `.hint("shuffle_hash")`.

**Schema, not inference, for big reads.** `inferSchema=True` reads the data twice. Define the schema:

```python
schema_names = "nconst string, primaryName string, birthYear int, deathYear int, knownForTitles string"
df = spark.read.schema(schema_names).options(header=True, delimiter="\t").csv("s3a://bucket/names.tsv.gz")
```

**`cache()`** a DataFrame reused across multiple actions (e.g. a join result queried several ways) to avoid recomputing the lineage. Don't cache things used once.

**Structured Streaming** turns a Kafka topic (or files) into an unbounded table processed as **micro-batches** (low latency + batch scalability). Use checkpointing (`checkpointLocation` on object storage) for fault-tolerant exactly-once recovery. See `end-to-end-pipeline.md` for a full Kafka→Spark→sink example.

## Submitting, monitoring, debugging

```bash
kubectl apply -f spark_job.yaml -n spark-jobs

kubectl get sparkapplication -n spark-jobs                 # high-level state
kubectl describe sparkapplication imdb-tsv-to-parquet -n spark-jobs
kubectl get pods -n spark-jobs                              # driver + executor pods
kubectl logs imdb-tsv-to-parquet-driver -n spark-jobs       # driver logs (the useful ones)

kubectl delete sparkapplication imdb-tsv-to-parquet -n spark-jobs   # required before re-running same name
```

Spark UI: `kubectl port-forward <driver-pod> 4040:4040 -n spark-jobs` then open `localhost:4040` — inspect Jobs/Stages/SQL tabs and the physical plan (look for unexpected `SortMergeJoin`/`Exchange` = shuffle).

Common failures:
- **`SparkApplication` stuck `SUBMITTED`, no driver pod** → ServiceAccount lacks pod-create RBAC, or webhook not running. Check operator logs and the `spark` SA binding.
- **Driver `Pending`** → no node fits the requested resources, or a missing toleration for a tainted node pool.
- **Executors OOMKilled** → raise `memoryOverhead`, then `memory`; reduce per-executor `cores` (less concurrent task memory pressure); fix skew (one huge partition).
- **`ClassNotFoundException` / `NoClassDefFoundError` for `S3AFileSystem` or Kafka** → connector jar missing from the image; add to `deps.packages` or bake into the image with matching versions.
- **`403`/`AccessDenied` on S3** → bad creds, wrong region, or missing IAM/bucket-policy permission.
- **`Never` restart + same name** → delete the old `SparkApplication` before resubmitting.
