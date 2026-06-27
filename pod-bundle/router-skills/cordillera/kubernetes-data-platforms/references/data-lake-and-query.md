# Data Lake, Table Formats & Query Layer

Object storage (MinIO/S3) as the lake, file/table formats (Parquet + Iceberg/Delta/Hudi), the medallion layout, and the query/consumption layer: Trino over the lake and Elasticsearch+Kibana via ECK.

## Contents
- Object storage as the lake (and when to self-host MinIO)
- File formats and open table formats
- Medallion layers
- Trino: deploy, catalogs, query
- Elasticsearch + Kibana via ECK

## Object storage as the lake

A data lake is centralized storage holding data **as-is** in native formats, with schema applied at read time (schema-on-read) rather than at write time. It separates storage from compute, so you scale and pay for each independently. Cloud object storage — **Amazon S3, Google Cloud Storage, Azure Blob** — is the default substrate: cheap, effectively infinite, durable, and zero-ops.

**Recommendation: keep the lake's storage layer off Kubernetes.** Cloud object storage already solves durability and scaling; running your own takes those problems back. The one good reason to self-host **MinIO** (an S3-compatible object store) on K8s is on-prem/air-gapped clusters or S3-compatible local development.

If you do run MinIO, use its **operator** (or Helm chart) for a distributed, erasure-coded `Tenant`; expose an S3 endpoint and point engines at it with `fs.s3a.endpoint` + `path.style.access=true`:

```yaml
# Minimal MinIO Tenant (MinIO Operator CRD) — illustrative
apiVersion: minio.min.io/v2
kind: Tenant
metadata:
  name: lake
  namespace: minio
spec:
  pools:
    - servers: 4
      volumesPerServer: 4
      volumeClaimTemplate:
        metadata:
          name: data
        spec:
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 100Gi
          storageClassName: gp3
  requestAutoCert: true
```

For all PVC/StorageClass mechanics, defer to **kubernetes-storage**.

## File formats and open table formats

**File format — use Parquet** for analytical data: columnar, compressed, splittable, with predicate pushdown. Avoid CSV/TSV/JSON for large datasets (row-oriented, slow, not splittable when gzipped). ORC is a fine alternative; Avro suits row-oriented streaming/serialization.

**Open table formats** sit *on top of* Parquet files in object storage and add the warehouse guarantees a raw file lake lacks — turning a lake into a **lakehouse**:

| Format | Strengths |
|---|---|
| **Apache Iceberg** | Vendor-neutral, broad engine support (Spark, Trino, Flink), hidden partitioning, schema/partition evolution, snapshots/time-travel. Often the default choice today. |
| **Delta Lake** | Strong in the Spark/Databricks ecosystem; ACID, time-travel, `MERGE` upserts. |
| **Apache Hudi** | Optimized for streaming upserts/incremental pulls and CDC ingestion. |

All three provide: **ACID transactions**, **schema enforcement + evolution**, **row-level upserts/deletes** (raw Parquet would require rewriting whole files), **time-travel / point-in-time** queries, and consistent reads while writing. Pick one and standardize — mixing them across one dataset is pain. Iceberg is the safest neutral default for a Trino-centric lakehouse.

A **catalog/metastore** maps table names to their files + schema: **AWS Glue Data Catalog**, **Hive Metastore**, or an Iceberg REST catalog. Trino and Spark read tables through it.

## Medallion layers

Organize the lake into quality tiers (typically separate prefixes/buckets):
- **bronze** — raw landed data, untouched (the source of truth).
- **silver** — cleaned, deduplicated, conformed, joined; consistent schemas; Parquet / table-format tables.
- **gold** — aggregated metrics, KPIs, BI-ready models.

A Spark batch job typically reads bronze → writes silver; another aggregates silver → gold; a crawler/catalog registers each so Trino can query it.

## Trino: deploy, catalogs, query

**Trino** (formerly PrestoSQL) is a distributed, MPP SQL engine that queries data *in place* across the lake and other sources — no ingestion step, no separate warehouse. One **coordinator** parses/plans/coordinates; many **workers** scan data and exchange intermediate results in parallel. It separates storage from compute and federates many sources via **connectors/catalogs** (Hive, Iceberg, Delta, PostgreSQL, Elasticsearch, …).

Deploy via the official Helm chart:

```bash
helm repo add trino https://trinodb.github.io/charts
helm repo update
helm install trino trino/trino -n trino --create-namespace -f trino-values.yaml
```

```yaml
# trino-values.yaml
server:
  workers: 3                 # scale for query concurrency / data size
coordinator:
  jvm:
    maxHeapSize: "8G"
worker:
  jvm:
    maxHeapSize: "8G"
service:
  type: ClusterIP            # expose via authenticated ingress in prod; LoadBalancer for demo only

# Catalogs define data sources:
catalogs:
  lake: |
    connector.name=iceberg
    iceberg.catalog.type=glue
    hive.metastore=glue
    fs.native-s3.enabled=true
    s3.region=us-east-1
  # A Hive catalog over Parquet via Glue (book-style):
  hive: |
    connector.name=hive
    hive.metastore=glue
    fs.native-s3.enabled=true
  # Federate a relational DB:
  pg: |
    connector.name=postgresql
    connection-url=jdbc:postgresql://db:5432/analytics
    connection-user=trino
    connection-password=...
```

Trino needs IAM permission to read the catalog (Glue) and the underlying S3 data — on EKS grant it via **IRSA on the Trino ServiceAccount** (preferred) rather than node-instance-role wildcards.

Verify and query:

```bash
kubectl get pods,svc -n trino
# pod/trino-coordinator-...  Running ; pod/trino-worker-...  Running x3
```

```sql
-- catalog.schema.table
SELECT pclass, sex, COUNT(1) AS people, AVG(age) AS avg_age
FROM hive."bdok-database".titanic
GROUP BY pclass, sex
ORDER BY sex, pclass;

-- Iceberg table with time-travel:
SELECT * FROM lake.silver.orders FOR VERSION AS OF 1234567890;
```

Connect a BI/SQL client (DBeaver, Superset, Tableau) to the Trino endpoint with user `trino` (set real auth in prod). **Apache Superset** (Helm-deployable on K8s) is a common open-source dashboard layer over Trino.

Production notes: **enable authentication** (password file, OAuth2, LDAP, or mTLS), keep the endpoint **private** (VPC + authenticated ingress, not a public LoadBalancer), set `query.max-memory` / `query.max-memory-per-node` to bound runaway queries, and scale workers for concurrency. For autoscaling Trino workers, see **kubernetes-autoscaling-scheduling**.

## Elasticsearch + Kibana via ECK

For real-time search, log/metric analytics, and live dashboards over semi/unstructured data, use **Elasticsearch + Kibana** managed by **ECK (Elastic Cloud on Kubernetes)**, the official Elastic operator. Elasticsearch stores JSON documents in **indices**, sharded across nodes (each index split into **primary shards** + **replica shards** for parallelism and HA). Choose shard count at index creation (can't reduce later) — a few shards per index is a good start; thousands of shards is an anti-pattern.

Install the operator:

```bash
helm repo add elastic https://helm.elastic.co
helm repo update
helm install elastic-operator elastic/eck-operator -n elastic --create-namespace
```

Elasticsearch cluster CR:

```yaml
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: elastic
  namespace: elastic
spec:
  version: 8.13.0
  volumeClaimDeletePolicy: DeleteOnScaledownAndClusterDeletion
  nodeSets:
    - name: default
      count: 3
      config:
        node.store.allow_mmap: false       # or set vm.max_map_count via the init container below
      podTemplate:
        spec:
          containers:
            - name: elasticsearch
              resources:
                requests: { memory: 2Gi, cpu: 1 }
                limits:   { memory: 2Gi }
          initContainers:
            - name: sysctl                  # production: raise mmap limit
              securityContext:
                privileged: true
                runAsUser: 0
              command: ['sh', '-c', 'sysctl -w vm.max_map_count=262144']
      volumeClaimTemplates:
        - metadata:
            name: elasticsearch-data
          spec:
            accessModes: ["ReadWriteOnce"]
            resources:
              requests:
                storage: 20Gi
            storageClassName: gp3
```

Kibana CR (references the ES cluster):

```yaml
apiVersion: kibana.k8s.elastic.co/v1
kind: Kibana
metadata:
  name: kibana
  namespace: elastic
spec:
  version: 8.13.0
  count: 1
  elasticsearchRef:
    name: elastic
  http:
    service:
      spec:
        type: ClusterIP          # demo: LoadBalancer; prod: ingress + auth
  podTemplate:
    spec:
      containers:
        - name: kibana
          env:
            - name: NODE_OPTIONS
              value: "--max-old-space-size=2048"
          resources:
            requests: { memory: 1Gi, cpu: 0.5 }
            limits:   { memory: 2Gi, cpu: 2 }
```

```bash
kubectl apply -f elastic.yaml -n elastic
kubectl apply -f kibana.yaml -n elastic
kubectl get elasticsearch -n elastic        # HEALTH green, PHASE Ready
# default 'elastic' user password:
kubectl get secret elastic-es-elastic-user -n elastic \
  -o go-template='{{.data.elastic | base64decode}}'
```

ECK enables TLS by default — **Kibana and the ES HTTP API are HTTPS-only** (use `https://…:5601` for Kibana). When a downstream client (e.g. a Kafka Connect Elasticsearch sink) must connect, extract the CA/cert from the ES `*-http-certs-public` secret and build a keystore/truststore for it (worked example in `end-to-end-pipeline.md`).

Trino answers *SQL over the lake*; Elasticsearch answers *real-time search/aggregation over recent semi-structured data*. A Lambda platform commonly uses both: Trino for the batch/serving layer, Elasticsearch+Kibana for the speed/real-time layer.
