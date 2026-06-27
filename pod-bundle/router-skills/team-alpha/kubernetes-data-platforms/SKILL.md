---
name: kubernetes-data-platforms
description: >-
  Run big-data, stateful, and AI workloads on Kubernetes. Use for the operator pattern for
  data systems, Apache Spark on Kubernetes (Spark Operator and SparkApplication, spark-submit
  with a k8s master, driver and executor pods, dynamic allocation), Apache Airflow (Helm
  chart, KubernetesExecutor, KubernetesPodOperator), Apache Kafka via the Strimzi operator
  (Kafka, KafkaTopic, KafkaUser, KRaft), object-storage data lakes (MinIO, S3) and table
  formats (Iceberg, Delta), query and consumption layers (Trino), end-to-end
  ingest-process-store-consume pipelines, and GenAI/LLM workloads with GPUs (NVIDIA device
  plugin and GPU Operator, MIG and time-slicing, model serving with vLLM/KServe/Triton).
  Trigger whenever the user runs Spark, Airflow, Kafka, Trino, a data lake, or model serving
  on Kubernetes, schedules GPUs, or builds a data pipeline on a cluster - even without saying
  Kubernetes. For generic PV/PVC see kubernetes-storage; for generic autoscaling see
  kubernetes-autoscaling-scheduling.
---

# Kubernetes Data Platforms

This skill equips Claude to run data, streaming, analytics, and AI workloads on Kubernetes the way a senior data-platform engineer would: pick the right operator, author correct CRDs and Helm values for Spark/Airflow/Kafka/Trino/Elasticsearch, wire them into end-to-end batch and real-time pipelines, schedule GPU/LLM serving, and isolate heavy jobs so they don't trample the cluster.

## When to use this skill

- Deciding **whether** to run a data engine on K8s at all, vs a managed service (EMR, MSK, Confluent Cloud, Dataproc, Databricks).
- Deploying or debugging **Apache Spark** on Kubernetes — `spark-submit` with a `k8s://` master, the Spark Operator and `SparkApplication`/`ScheduledSparkApplication` CRDs, driver/executor pod tuning, dynamic allocation, S3/object-store access.
- Deploying **Apache Airflow** — official Helm chart, `KubernetesExecutor`, `KubernetesPodOperator`/`SparkKubernetesOperator`, DAGs via gitSync, remote logging to S3.
- Deploying **Apache Kafka** with the **Strimzi** operator — `Kafka`, `KafkaNodePool`, `KafkaTopic`, `KafkaUser`, `KafkaConnect`, `KafkaConnector` CRDs; KRaft vs ZooKeeper; JBOD storage; partition rebalancing with Cruise Control.
- Standing up a **data lake / lakehouse** — MinIO or cloud object storage, Parquet, open table formats (Iceberg/Delta/Hudi), the bronze/silver/gold medallion layout.
- Adding a **consumption/query layer** — Trino/Presto over the lake, Elasticsearch + Kibana via the ECK operator.
- Building a **complete pipeline** that chains ingestion → processing → storage → serving, batch or streaming.
- Scheduling **GPU and LLM workloads** — the NVIDIA device plugin, `nvidia.com/gpu` requests, model servers (vLLM, KServe, Triton), and the trade-offs of self-hosting vs a managed inference API.
- Controlling **storage throughput, resource isolation, and noisy-neighbor** behavior for heavy jobs.

## Core concepts

**Why Kubernetes for data, and when not to.** K8s gives data engines one substrate: unified resource management, bin-packing, dynamic scaling, declarative GitOps deploys, and cloud-agnostic portability. The cost is real operational complexity — clustered stateful systems need careful storage, networking, and lifecycle handling that a managed service hides. Rule of thumb: run it on K8s when you already operate K8s and want one control plane, need portability/cost control, or want tight coupling between jobs and your other workloads. Reach for a **managed service** when the team is small, the data system is your single biggest operational risk (a large production Kafka cluster), or you have no appetite for on-call storage debugging. A frequent, well-judged split: **keep object storage off K8s** (S3/GCS/Azure Blob are cheap, durable, and effortless) and run only the *compute* engines (Spark, Trino, Airflow) and *brokers* (Kafka) in the cluster.

**The operator pattern is how stateful data systems run on K8s.** A bare Deployment/StatefulSet can't safely manage a Kafka rebalance, a Spark driver→executor topology, or an Elasticsearch shard migration. An **operator** = a set of **Custom Resource Definitions (CRDs)** that add new object types to the API (e.g. `SparkApplication`, `Kafka`, `Elasticsearch`) + a **controller** that watches those objects and drives the real cluster toward the declared spec (creating pods, mounting volumes, handling failover, exposing metrics). You declare *what* you want; the controller does the *how*. Every engine in this skill is deployed through an operator (Spark Operator, Strimzi, ECK) or a mature Helm chart (Airflow, Trino). Operators are usually installed via Helm, then driven with `kubectl apply` of CRs.

**Modern data architecture in two shapes.** *Lambda* runs separate **batch** and **speed (streaming)** layers and merges them in a serving layer — proven, flexible, but two systems to operate. *Kappa* collapses everything into one stream-processing path over an immutable log (Kafka) and replays history when needed — simpler in theory, harder to operate at scale. Most real platforms blend them. The **lakehouse** unifies cheap object-store data lakes with warehouse-grade SQL and ACID via open table formats, organized in **medallion** layers: *bronze* (raw landed data), *silver* (cleaned/conformed), *gold* (aggregated, BI-ready).

**The canonical open-source stack** these patterns map to: ingest with **Kafka/Kafka Connect** (streaming, CDC) or **Spark/custom Python** (batch); process with **Spark** (batch via Spark SQL/DataFrames, streaming via Structured Streaming micro-batches); store on **object storage** in Parquet + a table format; orchestrate with **Airflow**; serve with **Trino** (SQL over the lake) and **Elasticsearch/Kibana** (real-time search/dashboards). Airflow **orchestrates** — it must never *do* the heavy data processing itself (it triggers Spark jobs as pods).

**Heavy jobs need explicit isolation.** Data engines are resource-hungry and bursty. Without `requests`/`limits`, ResourceQuotas, dedicated node pools, and PriorityClasses, a Spark executor storm or a Kafka rebalance will starve everything else. Treat resource isolation as a first-class design input, not an afterthought (see the noisy-neighbor section in `references/data-on-k8s-and-operators.md`).

## Workflow / how to approach data-platform tasks

### 1. Choose the engine and the deployment mechanism
Match the need to the tool, then to its operator/chart:
- Distributed batch/stream processing → **Spark** via the **Spark Operator** (`SparkApplication` CRD). Use raw `spark-submit --master k8s://...` only for quick one-offs or CI.
- Workflow orchestration / scheduling / dependencies → **Airflow** via the official Helm chart with `KubernetesExecutor`.
- Durable pub/sub log, CDC, real-time ingestion → **Kafka** via **Strimzi**.
- Interactive SQL over the lake → **Trino** Helm chart.
- Real-time search / log analytics / dashboards → **Elasticsearch + Kibana** via the **ECK** operator.
- LLM/ML serving → GPU node pool + **KServe**/vLLM/Triton.

### 2. Install the operator, then declare the workload
The pattern is identical across engines:
1. `helm repo add` + `helm install` the operator/chart into its own namespace.
2. Verify the controller pod is `Running` (`kubectl get pods -n <ns>`).
3. `kubectl apply -f` the custom resource (`SparkApplication`, `Kafka`, `Elasticsearch`, …).
4. Watch the CR's status, not just pods: `kubectl get sparkapplication`, `kubectl get kafka`, `kubectl get elastic`.

### 3. Wire object-store / credentials access
Most jobs read/write S3-compatible storage. Create credentials as a `Secret` and inject as env vars (`AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY`) — **prefer IRSA / Workload Identity / pod-level cloud identity over long-lived keys** in production. For Spark, also set the `s3a` Hadoop properties (`fs.s3a.impl`, V4 signing, fast upload). Details and current YAML in `references/spark-on-kubernetes.md` and `references/data-lake-and-query.md`.

### 4. Build the pipeline
Compose the engines. Airflow is the spine: a DAG triggers Spark jobs as `SparkApplication`s (via `SparkKubernetesOperator` + `SparkKubernetesSensor`), waits for completion, then triggers cataloging/serving steps. For streaming: Kafka Connect (CDC source) → topic → Spark Structured Streaming → topic/sink connector → Elasticsearch or object store. Full worked batch + streaming pipelines in `references/end-to-end-pipeline.md`.

### 5. Right-size, isolate, and observe
Set `requests`/`limits` on every driver/executor/broker. Use dedicated node pools (taints + tolerations + nodeSelector) for big jobs, ResourceQuotas per namespace, and PriorityClasses so latency-sensitive serving outranks batch. Scrape the operators' Prometheus metrics. For HPA/cluster-autoscaler specifics defer to **kubernetes-autoscaling-scheduling**; for StorageClass/PVC mechanics defer to **kubernetes-storage**.

## Common pitfalls & anti-patterns

- **Doing real data processing inside Airflow tasks.** Airflow orchestrates; it is not a compute engine. A `PythonOperator` that loads a big DataFrame will OOM the worker. Trigger Spark (or a `KubernetesPodOperator`) instead and have Airflow only wait on it.
- **Running object storage on K8s without a reason.** Self-hosting MinIO means you now own its scalability and durability. Unless you need on-prem/air-gapped or S3-compatible local dev, use the cloud object store and keep only compute on K8s.
- **No resource requests/limits on Spark executors or Kafka brokers.** Guarantees noisy-neighbor incidents and unschedulable pods. Always set them; cap Spark with `spark.cores.max` / `spark.executor.instances`.
- **Treating Kafka as durable primary storage.** Topics have a retention window (default 7 days). Sink important data to object storage; don't make Kafka your system of record.
- **`inferSchema=True` on large Spark reads.** Spark scans the data twice (once to infer, once to read). Define the schema explicitly for big text/CSV/TSV inputs.
- **Ignoring narrow-vs-wide transformation order.** Filter/project (narrow) *before* joins/groupBy/sort (wide) to shrink the shuffle. Shuffles are the dominant Spark cost.
- **`deleteClaim: true` (or default) on Kafka/stateful storage you care about.** Deleting the CR then nukes the PVCs and your data. Set `deleteClaim: false` for anything durable.
- **Exposing Trino/Elasticsearch/Kibana via a public LoadBalancer with no auth.** Fine for a throwaway demo, a breach in production. Keep them private (VPC/ClusterIP + ingress with auth) and enable authentication.
- **Self-hosting an LLM on K8s when an API would do.** GPUs are scarce, expensive, and operationally heavy. Self-host only when you need data residency, fine-tuned/private models, or predictable high-volume cost — otherwise call a hosted inference API and skip the GPU ops.
- **Forgetting the GPU device plugin / `runtimeClass`.** Without the NVIDIA device plugin DaemonSet, `nvidia.com/gpu` is an unschedulable resource and GPU pods sit `Pending` forever.
- **Using removed/legacy API versions.** Strimzi has dropped ZooKeeper-based and `v1beta1` Kafka in current releases (KRaft + `KafkaNodePool` are the norm); Spark Operator moved to the `kubeflow/spark-operator` project with the `v1beta2` API. Always check the chart/CRD version you actually installed.

## Reference files

- **`references/data-on-k8s-and-operators.md`** — when to run data on K8s vs managed; the operator pattern in depth; Helm-vs-operator; resource isolation, node pools, quotas, PriorityClasses, noisy-neighbor control; Lambda/Kappa/lakehouse architecture. Read for "should this even be on K8s?" and platform-design questions.
- **`references/spark-on-kubernetes.md`** — Spark Operator install, full `SparkApplication`/`ScheduledSparkApplication` YAML, `spark-submit` with `k8s://`, driver/executor pods, dynamic allocation, S3/`s3a` config, resource tuning, transformation/join internals, debugging failed jobs. Read for any Spark task.
- **`references/airflow-on-kubernetes.md`** — Airflow Helm chart values, `KubernetesExecutor`, gitSync DAGs, remote logging, `KubernetesPodOperator`/`SparkKubernetesOperator`, RBAC for launching Spark, DAG authoring best practices. Read for any Airflow task.
- **`references/kafka-strimzi.md`** — Strimzi install, `Kafka` (KRaft + `KafkaNodePool` and legacy ZooKeeper), `KafkaTopic`, `KafkaUser`, listeners/TLS, JBOD storage, `KafkaConnect`/`KafkaConnector` (JDBC CDC, S3/ES sinks), Cruise Control rebalancing, CLI. Read for any Kafka task.
- **`references/data-lake-and-query.md`** — object storage / MinIO, Parquet + Iceberg/Delta/Hudi table formats, medallion layers, Trino deployment + catalogs (Hive/Glue/Iceberg), Elasticsearch via ECK + Kibana. Read for storage-layer and query-engine tasks.
- **`references/end-to-end-pipeline.md`** — full worked batch pipeline (Airflow → Spark → S3 → catalog → Trino) and streaming pipeline (Kafka Connect CDC → Spark Structured Streaming → Elasticsearch), with the glue YAML and RBAC. Read when assembling multiple engines.
- **`references/genai-and-gpu-workloads.md`** — NVIDIA device plugin, `nvidia.com/gpu` requests/limits, GPU node pools/MIG/time-slicing, model serving (vLLM, KServe, Triton), RAG/agent app deployment, self-host-vs-managed trade-offs. Read for GPU/LLM tasks.
