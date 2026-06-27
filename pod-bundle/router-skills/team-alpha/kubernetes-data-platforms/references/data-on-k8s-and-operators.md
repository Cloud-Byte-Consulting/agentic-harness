# Data on Kubernetes, Operators & Resource Isolation

Platform-level decisions: whether to run data engines on K8s at all, how operators work, and how to keep heavy jobs from destroying the cluster.

## Contents
- Should this run on Kubernetes?
- The operator pattern (CRDs + controllers)
- Operator vs Helm chart
- Architecture patterns: Lambda, Kappa, lakehouse
- Resource isolation & noisy-neighbor control
- Storage throughput for data engines

## Should this run on Kubernetes?

Run a data engine **on K8s** when:
- You already operate Kubernetes and want a single control plane, scheduler, and GitOps workflow for everything.
- You need portability across clouds / on-prem, or want to avoid managed-service lock-in and markup.
- You want jobs co-located and tightly integrated with your other containerized workloads.
- The workload is *compute* (Spark, Trino, Airflow): stateless-ish, bursty, scales to zero between jobs — an ideal K8s fit.

Prefer a **managed service** (EMR/Glue, MSK/Confluent Cloud, Dataproc, Databricks, OpenSearch Service, BigQuery/Snowflake) when:
- The team is small and the data system is your single largest operational risk (a big production Kafka or a multi-TB OLAP cluster).
- You have no appetite to debug PVC attach failures, broker rebalances, or shard recovery at 3 a.m.
- The managed offering's cost is acceptable relative to the engineering time saved.

**The pragmatic hybrid most teams land on:** keep **object storage off K8s** (S3/GCS/Azure Blob are durable, cheap, infinitely scalable, and zero-ops) and run the *compute* engines and *brokers* in the cluster. Self-hosting MinIO on K8s only pays off for on-prem/air-gapped environments or S3-compatible local development — otherwise you've taken on durability and scaling problems the cloud already solved.

State maturity matters: stateless query/compute (Spark, Trino) is low-risk on K8s today. Stateful clustered systems (Kafka, Elasticsearch, databases) are viable **through a mature operator** but demand solid storage (fast CSI, correct `volumeBindingMode`), anti-affinity across nodes/zones, and tested backup/restore.

## The operator pattern

A plain `Deployment` or `StatefulSet` only knows how to keep N identical pods running. It has no idea how to add a Kafka broker and trigger a partition rebalance, scale Spark executors under a driver, or migrate Elasticsearch shards before draining a node. That domain knowledge lives in an **operator**.

An operator has two parts:

1. **Custom Resource Definitions (CRDs)** — extend the Kubernetes API with new object kinds. After installing Strimzi you can `kubectl get kafka`, `kubectl get kafkatopic`; after the Spark Operator you can `kubectl get sparkapplication`. These behave like built-in objects (RBAC, `kubectl`, GitOps all work).

2. **A controller** — a pod running a reconcile loop that *watches* those custom resources and continuously drives the real system toward the declared `spec`: creating/deleting pods, mounting volumes, configuring the app, handling failover, and writing observed state back to the CR's `status`.

Example: you `kubectl apply` a `SparkApplication`. The Spark Operator controller sees it and:
- creates the driver pod from the spec,
- the driver requests executors, which the operator/Spark schedules as pods,
- mounts configured volumes and secrets,
- monitors the app and surfaces status (`SUBMITTED` → `RUNNING` → `COMPLETED`/`FAILED`),
- emits metrics/logs for Prometheus/Grafana.

You declared *what* (a Python Spark job with these resources); the controller handled *how*.

Why this matters for you as an author:
- **Always reconcile through the CR**, never by hand-editing the pods/StatefulSets the operator manages — the controller will revert you.
- **Check the CR `status` and `kubectl describe`** for debugging, not just `kubectl get pods`.
- **Mind the CRD/operator version** — APIs evolve (e.g. Strimzi `v1beta1`→`v1beta2`, ZooKeeper→KRaft+`KafkaNodePool`; the Spark Operator's move to the `kubeflow/spark-operator` project). Manifests must match the version actually installed.

Benefits operators give you: simplified lifecycle (high-level abstraction over K8s internals), built-in Prometheus metrics, day-2 automation (upgrades, scaling, rebalancing), and cloud-agnostic operation.

## Operator vs Helm chart

Both are common; they solve different things and often combine.
- **Helm** is a *package manager*: templated YAML (a "chart") with values for install-time configuration. Great for deploying an app once. Many operators are themselves *installed* via Helm.
- **Operators** add *ongoing* domain-aware reconciliation and lifecycle automation that a static Helm release can't.

In practice: `helm install` the **operator** (Spark Operator, Strimzi, ECK), then drive workloads with **CRDs**. For apps without a dedicated operator (Airflow, Trino), a well-maintained **Helm chart** is the deployment path and you configure via `values.yaml`.

## Architecture patterns

**Lambda** — separate **batch** layer (high throughput over the full master dataset, e.g. nightly Spark) and **speed** layer (low-latency over recent data, e.g. Spark Structured Streaming off Kafka), merged at a **serving** layer for unified queries. Proven and flexible — you can even run batch-only if you don't need streaming — at the cost of operating two processing paths.

**Kappa** — one **stream-processing** path over an immutable append-only log (Kafka). Historical reprocessing is done by *replaying* the log. Conceptually simpler (one system), but harder to operate well at very large scale and costs can climb.

**Lakehouse** — combine the cheap, any-format scale of a **data lake** (object storage, schema-on-read) with the SQL performance and **ACID** guarantees of a **warehouse**, via open table formats (Delta Lake, Apache Iceberg, Apache Hudi) that add transactions, schema evolution, time-travel, and row-level upserts/deletes on top of object storage. Organized in **medallion** layers:
- **Bronze** — raw, as-landed data (any format: logs, CSV, JSON, images). The source of truth; no transformation.
- **Silver** — cleaned, deduplicated, conformed, joined; consistent schemas; query-ready (Parquet / table-format tables).
- **Gold** — aggregated metrics, KPIs, BI-ready models.
- (Optional **landing** zone before bronze for truly raw drops.)

Most production platforms blend Lambda + lakehouse: a medallion lake on object storage, batch via Spark+Airflow, streaming via Kafka+Spark, served by Trino (SQL) and Elasticsearch (real-time).

## Resource isolation & noisy-neighbor control

Data jobs are bursty and hungry; one unbounded Spark stage or Kafka rebalance can starve latency-sensitive services. Defend the cluster:

**Always set `requests` and `limits`** on every driver, executor, broker, and worker. Requests drive scheduling and protect the pod; limits cap blast radius. For Spark, additionally bound parallelism with `spark.cores.max`, `spark.executor.instances`, and per-pod `cores`/`memory`.

**Dedicated node pools** for heavy/spiky workloads. Taint the pool and tolerate it on the data pods so general workloads don't land there:

```yaml
# On the SparkApplication driver/executor spec, or any data pod:
tolerations:
  - key: "workload"
    operator: "Equal"
    value: "data"
    effect: "NoSchedule"
nodeSelector:
  node-pool: "data"
```

Taint nodes with `kubectl taint nodes <node> workload=data:NoSchedule` (or set it on the managed node group). GPU pools work the same way (see `genai-and-gpu-workloads.md`).

**ResourceQuota per namespace** caps total consumption so one team's batch can't eat the cluster:

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: data-quota
  namespace: spark-jobs
spec:
  hard:
    requests.cpu: "40"
    requests.memory: 128Gi
    limits.cpu: "60"
    limits.memory: 192Gi
    pods: "100"
```

**LimitRange** sets per-pod defaults/maximums inside a namespace so a forgotten request doesn't sneak through.

**PriorityClass** lets latency-sensitive serving outrank batch and survive preemption:

```yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: data-serving
value: 1000000
globalDefault: false
description: "Trino/Elasticsearch/model-serving — preempt batch under pressure."
---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: data-batch
value: 10000
globalDefault: false
description: "Spark/Airflow batch — preemptible."
```

Reference the class via `priorityClassName` on the pod/CR spec.

**Pod anti-affinity** spreads brokers/replicas across nodes and zones so a node loss doesn't take out a quorum (Strimzi and ECK expose this in their CRDs; see those references).

> For HPA, VPA, cluster-autoscaler, and the scheduler internals behind all of the above, defer to **kubernetes-autoscaling-scheduling**. This skill only covers the data-engine-specific knobs (`spark.cores.max`, broker resources, executor counts).

## Storage throughput for data engines

Heavy jobs are I/O-bound as often as CPU-bound:
- **Use a fast CSI StorageClass** (NVMe/SSD-backed, e.g. cloud `gp3`/premium SSD) for Kafka logs, Elasticsearch data, and Spark shuffle/spill. Spinning disks throttle throughput.
- **`volumeBindingMode: WaitForFirstConsumer`** so a volume is provisioned in the same zone as the pod (avoids cross-zone latency and unschedulable pods). See **kubernetes-storage**.
- **Spark shuffle/spill** can use large `emptyDir` (or local NVMe) for scratch — size it; default node ephemeral storage is small and an over-spill evicts the pod.
- **Kafka prefers JBOD** (multiple independent persistent volumes per broker) over one big disk — more spindles, more parallel throughput; set `deleteClaim: false`.
- **Set `requests.ephemeral-storage`** on jobs that write large temp files so the scheduler accounts for it.
- **RWX vs RWO**: most cloud block disks are `ReadWriteOnce` (one node). Don't design a job that assumes many pods share one block PVC — use object storage for shared data, or a real RWX filesystem (EFS/CephFS) only when genuinely needed.

For all StorageClass/PVC/volume mechanics (provisioning, expansion, snapshots, access modes), defer to **kubernetes-storage** — this file only covers what's specific to feeding data engines.
