# End-to-End Pipelines on Kubernetes

Two complete worked pipelines wiring the engines together: a **batch** pipeline (Airflow → Spark → object store → catalog → Trino) and a **real-time** pipeline (Kafka Connect CDC → Spark Structured Streaming → Elasticsearch). The glue YAML/RBAC is the point here; engine-specific depth lives in the per-engine references.

## Contents
- Prerequisites checklist
- Batch pipeline: ingest → Spark transform → catalog → Trino
- Streaming pipeline: CDC → Spark Streaming → Elasticsearch
- Cross-namespace & RBAC glue
- Verification & teardown

## Prerequisites checklist

Before assembling, confirm each operator/engine is healthy:

```bash
kubectl get pods -n spark-operator     # Spark Operator controller + webhook
kubectl get pods -n kafka              # Strimzi operator (+ Kafka cluster if up)
kubectl get pods -n elastic            # ECK operator (+ ES/Kibana)
kubectl get pods -n trino              # Trino coordinator + workers
kubectl get pods -n airflow            # Airflow scheduler/webserver
```

Per-namespace plumbing that pipelines depend on:
- A `spark` ServiceAccount with pod-management RBAC in any namespace that runs Spark.
- An `aws-credentials` Secret (or IRSA/Workload Identity) in each namespace whose pods touch object storage.
- Airflow worker SA bound to a ClusterRole that can manage `sparkapplications` (see `airflow-on-kubernetes.md`).

## Batch pipeline: ingest → Spark transform → catalog → Trino

Goal: download source data, land it in S3 (bronze), transform TSV→Parquet, join into a consolidated "one big table" (silver), catalog it, and query in Trino. **Airflow orchestrates; Spark does the compute.**

**DAG shape** (TaskFlow + Spark operator/sensor + a catalog crawler step):

```python
from airflow.decorators import dag, task
from airflow.utils.task_group import TaskGroup
from airflow.providers.cncf.kubernetes.operators.spark_kubernetes import SparkKubernetesOperator
from airflow.providers.cncf.kubernetes.sensors.spark_kubernetes import SparkKubernetesSensor
from airflow.providers.amazon.aws.operators.glue_crawler import GlueCrawlerOperator
from datetime import datetime

@dag(schedule="@daily", start_date=datetime(2024, 1, 1), catchup=False, tags=["imdb"])
def imdb_batch():

    @task
    def acquire():
        # download sources, upload to s3://bucket/landing/imdb/  (bronze)
        return True

    with TaskGroup("tsv_to_parquet") as g1:
        submit1 = SparkKubernetesOperator(
            task_id="tsv_to_parquet",
            namespace="airflow",
            application_file="spark_tsv_parquet.yaml",
            kubernetes_conn_id="kubernetes_default",
            do_xcom_push=True,
        )
        wait1 = SparkKubernetesSensor(
            task_id="tsv_to_parquet_sensor",
            namespace="airflow",
            application_name="{{ task_instance.xcom_pull(task_ids='tsv_to_parquet.tsv_to_parquet')['metadata']['name'] }}",
            kubernetes_conn_id="kubernetes_default",
        )
        submit1 >> wait1

    with TaskGroup("transform") as g2:
        submit2 = SparkKubernetesOperator(
            task_id="consolidated_table",
            namespace="airflow",
            application_file="spark_consolidated.yaml",
            kubernetes_conn_id="kubernetes_default",
            do_xcom_push=True,
        )
        wait2 = SparkKubernetesSensor(
            task_id="consolidated_table_sensor",
            namespace="airflow",
            application_name="{{ task_instance.xcom_pull(task_ids='transform.consolidated_table')['metadata']['name'] }}",
            kubernetes_conn_id="kubernetes_default",
        )
        submit2 >> wait2

    catalog = GlueCrawlerOperator(
        task_id="crawl_consolidated",
        region_name="us-east-1",
        aws_conn_id="aws_conn",
        wait_for_completion=True,
        config={"Name": "imdb_consolidated_crawler"},
    )

    acquire() >> g1 >> g2 >> catalog

imdb_batch()
```

The `SparkApplication` YAMLs (`spark_tsv_parquet.yaml`, `spark_consolidated.yaml`) live in the `dags/` folder beside the DAG so the operator can read them. Each is a normal `SparkApplication` (see `spark-on-kubernetes.md`) whose PySpark `mainApplicationFile` sits in S3.

**Transform job 1** (bronze → silver, schema-defined read, write Parquet):

```python
schema_names = "nconst string, primaryName string, birthYear int, deathYear int, knownForTitles string"
names = (spark.read.schema(schema_names)
         .options(header=True, delimiter="\t")
         .csv("s3a://bucket/landing/imdb/names.tsv.gz"))
names.write.mode("overwrite").parquet("s3a://bucket/bronze/imdb/names")
```

**Transform job 2** (join into the consolidated table — filter/explode before joins to shrink shuffles):

```python
from pyspark.sql import functions as f
names = spark.read.parquet("s3a://bucket/bronze/imdb/names")
basics = spark.read.parquet("s3a://bucket/bronze/imdb/basics")
ratings = spark.read.parquet("s3a://bucket/bronze/imdb/ratings")

names = names.select("nconst", "primaryName",
                     f.explode(f.split("knownForTitles", ",")).alias("knownForTitles"))
basics_ratings = basics.join(ratings, on="tconst", how="inner")
obt = basics_ratings.join(names, basics_ratings.tconst == names.knownForTitles, "inner").dropDuplicates()
obt.write.mode("overwrite").parquet("s3a://bucket/silver/imdb/consolidated")
```

After the crawler registers the table in the catalog, query in Trino:

```sql
SELECT primaryTitle, startYear, averageRating
FROM hive."bdok-database"."imdb_consolidated"
WHERE primaryName = 'Keanu Reeves'
ORDER BY averageRating DESC;
```

Flow recap: **Airflow** (download to bronze) → **Spark** (bronze→silver TSV→Parquet) → **Spark** (join→OBT silver) → **catalog crawler** → **Trino** (serve). Airflow never holds the data; it triggers pods and waits on sensors.

## Streaming pipeline: CDC → Spark Streaming → Elasticsearch

Goal: capture row changes from Postgres in real time, transform them with Spark Structured Streaming, and land them in Elasticsearch for live Kibana dashboards. All components share the **`kafka` namespace** so Kafka Connect can reach Elasticsearch internally.

Flow: **Postgres** → **Kafka Connect JDBC/Debezium source** → topic `src-customers` → **Spark Structured Streaming** (transform) → topic `customers-transformed` → **Kafka Connect Elasticsearch sink** → **Elasticsearch index** → **Kibana**.

**1. Kafka + Kafka Connect + Elasticsearch** — deploy Strimzi Kafka and an ECK Elasticsearch *into the `kafka` namespace* (see `kafka-strimzi.md`, `data-lake-and-query.md`). Because ECK forces TLS, build a keystore so Connect can reach ES over HTTPS:

```bash
# pull ES CA + cert + key
kubectl get secret elastic-es-http-certs-public  -n kafka -o go-template='{{index .data "ca.crt"  | base64decode}}' > ca.crt
kubectl get secret elastic-es-http-certs-public  -n kafka -o go-template='{{index .data "tls.crt" | base64decode}}' > tls.crt
kubectl get secret elastic-es-http-certs-internal -n kafka -o go-template='{{index .data "tls.key" | base64decode}}' > tls.key
# build a JKS keystore/truststore
openssl pkcs12 -export -in tls.crt -inkey tls.key -CAfile ca.crt -caname root \
  -out keystore.p12 -password pass:changeit -name es
keytool -importkeystore -srckeystore keystore.p12 -srcstoretype PKCS12 -srcstorepass changeit \
  -deststorepass changeit -destkeypass changeit -destkeystore keystore.jks -alias es
# mount it into Connect as a secret
kubectl create secret generic es-keystore --from-file=keystore.jks -n kafka
```

Reference the keystore secret via `KafkaConnect.spec.externalConfiguration.volumes` so it mounts at `/opt/kafka/external-configuration/...`.

**2. JDBC source connector** — `KafkaConnector` capturing the `customers` table into `src-customers` (full YAML in `kafka-strimzi.md`). Verify messages arrive:

```bash
kubectl exec my-cluster-broker-0 -n kafka -c kafka -it -- \
  bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic src-customers
```

**3. Spark Structured Streaming** — read the topic, transform, write back to a new topic. Run as a long-lived `SparkApplication` (`restartPolicy: Always`) in `kafka`:

```python
from pyspark.sql import functions as f
df = (spark.readStream.format("kafka")
      .option("kafka.bootstrap.servers", "my-cluster-kafka-bootstrap:9092")
      .option("subscribe", "src-customers")
      .option("startingOffsets", "earliest")
      .load())

# parse the Connect JSON envelope (schema/payload) -> business fields, then enrich:
enriched = (parsed
    .withColumn("today", f.to_date(f.current_timestamp()))
    .withColumn("age", f.round(f.datediff("today", f.to_date("birthdate")) / 365.25, 0)))

(enriched.selectExpr("to_json(struct(*)) AS value")
    .writeStream.format("kafka")
    .option("kafka.bootstrap.servers", "my-cluster-kafka-bootstrap:9092")
    .option("topic", "customers-transformed")
    .option("checkpointLocation", "s3a://bucket/spark-checkpoint/customers/")   # exactly-once recovery
    .start().awaitTermination())
```

The job's `SparkApplication` image needs the Kafka connector (`org.apache.spark:spark-sql-kafka-0-10_2.12:<spark-version>`) — bake it in or list it in `deps.packages`. **`checkpointLocation` on object storage is mandatory** for fault-tolerant streaming.

**4. Elasticsearch sink connector** — `KafkaConnector` reading `customers-transformed` into an ES index over TLS using the mounted keystore:

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnector
metadata:
  name: es-sink
  namespace: kafka
  labels:
    strimzi.io/cluster: my-connect
spec:
  class: io.confluent.connect.elasticsearch.ElasticsearchSinkConnector
  tasksMax: 1
  config:
    topics: "customers-transformed"
    connection.url: "https://elastic-es-http.kafka:9200"
    connection.username: "elastic"
    connection.password: "${secrets:kafka/es-creds:password}"
    key.ignore: "true"
    schema.ignore: "false"
    elastic.security.protocol: "SSL"
    elastic.https.ssl.keystore.location: "/opt/kafka/external-configuration/es-keystore-volume/keystore.jks"
    elastic.https.ssl.keystore.password: "changeit"
    elastic.https.ssl.keystore.type: "JKS"
    elastic.https.ssl.truststore.location: "/opt/kafka/external-configuration/es-keystore-volume/keystore.jks"
    elastic.https.ssl.truststore.password: "changeit"
    elastic.https.ssl.truststore.type: "JKS"
```

**5. Verify in Kibana** — `GET _cat/indices` should show `customers-transformed`; create a data view on it (timestamp `dt_update`) and build dashboards. Run more source inserts and watch them flow through end-to-end.

## Cross-namespace & RBAC glue

- **Co-locate Connect and its sink target.** Kafka Connect in `kafka` reaches the ES service `elastic-es-http.kafka:9200` by in-cluster DNS only because both are in `kafka`. Cross-namespace works too (`<svc>.<ns>.svc`) but watch NetworkPolicies.
- **Spark in the `kafka` namespace** needs its own `spark` SA + edit binding there:
  ```bash
  kubectl create serviceaccount spark -n kafka
  kubectl create clusterrolebinding spark-role-kafka --clusterrole=edit --serviceaccount=kafka:spark
  ```
- **Secrets are namespaced** — recreate `aws-credentials` / `es-creds` in every namespace that needs them; a Secret in `spark-jobs` is invisible to pods in `kafka`.
- **Don't hardcode passwords in connector CRs** — mount via `KafkaConnect.spec.externalConfiguration` and reference with `${secrets:...}` config providers.

## Verification & teardown

```bash
# batch
kubectl get sparkapplication -n airflow
kubectl logs <driver-pod> -n airflow

# streaming
kubectl get kafkaconnector -n kafka
kubectl describe kafkaconnector es-sink -n kafka
kubectl get sparkapplication -n kafka

# teardown (note deleteClaim:false keeps Kafka/ES PVCs)
kubectl delete sparkapplication --all -n kafka
kubectl delete kafkaconnector --all -n kafka
helm uninstall airflow -n airflow
```

When tearing down stateful clusters, remember `deleteClaim: false` (Kafka) and `volumeClaimDeletePolicy` (ES) govern whether data volumes survive — and that deleting a namespace deletes its PVCs.
