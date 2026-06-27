# Apache Kafka on Kubernetes with Strimzi

Deploying and operating Kafka via the Strimzi operator: the `Kafka`/`KafkaNodePool`/`KafkaTopic`/`KafkaUser` CRDs, KRaft vs ZooKeeper, listeners/TLS, JBOD storage, Kafka Connect for CDC/sinks, rebalancing, and the CLI.

## Contents
- Kafka mental model
- Why Strimzi
- Installing the operator
- KRaft vs ZooKeeper (version note)
- Kafka cluster: KafkaNodePool + Kafka (current) and legacy form
- Listeners, TLS & external access
- Storage: JBOD and persistent claims
- KafkaTopic and KafkaUser CRDs
- Kafka Connect & KafkaConnector (JDBC CDC, sinks)
- Rebalancing with Cruise Control
- CLI, monitoring, gotchas

## Kafka mental model

Kafka is a distributed, partitioned, replicated **publish-subscribe log**.
- **Producers** write **records** to **topics**; **consumers** read from topics.
- A topic is split into **partitions** — each an ordered, immutable, append-only sequence. Partitions are the unit of parallelism and ordering (order is guaranteed *within* a partition, not across).
- Each record in a partition has a monotonic **offset**; consumers track their offset to resume exactly where they left off (offsets are themselves stored in the `__consumer_offsets` topic).
- Each partition has a **leader** broker and **follower** replicas; all reads/writes go to the leader, followers replicate it, and a follower is promoted on leader failure. **Replication factor** controls copies (production: 3 brokers, RF 3, `min.insync.replicas=2`).
- **Consumer groups** load-balance partitions across consumers and preserve per-partition ordering — the mechanism for scaling consumption.
- **Retention is a window** (default 7 days), not permanent storage. Sink durable data elsewhere.

## Why Strimzi

Hand-running Kafka StatefulSets on K8s means manually handling broker config, rolling restarts that respect partition leadership, storage, listeners, certs, and rebalancing. **Strimzi** (a CNCF project) is the operator that turns all of this declarative: you describe the desired Kafka cluster as CRs and the operator reconciles it — provisioning brokers, generating TLS certs and a CA, managing rolling updates safely, and exposing metrics.

## Installing the operator

```bash
helm repo add strimzi https://strimzi.io/charts/
helm repo update

helm install strimzi strimzi/strimzi-kafka-operator \
  --namespace kafka --create-namespace
```

Verify and check which CRDs/API versions you got:

```bash
kubectl get pods -n kafka                  # strimzi-cluster-operator-... Running
kubectl get crd | grep strimzi
```

## KRaft vs ZooKeeper (version note)

- **Current Strimzi (and Kafka 4.x) are KRaft-only** — Kafka manages its own metadata quorum via **controller** nodes; **ZooKeeper is removed**. Clusters are defined with **`KafkaNodePool`** resources (separate pools for `controller` and `broker` roles) plus the `Kafka` CR, and require the annotation `strimzi.io/node-pools: enabled` (and `strimzi.io/kraft: enabled`).
- **Older books/tutorials show a ZooKeeper-based `Kafka` CR** with a `spec.zookeeper` block and brokers inline. That form is **deprecated/removed in current releases** — only use it if you're pinned to an old Strimzi version, and plan to migrate.

**Always match your manifests to the Strimzi version you installed.** Below shows the current KRaft+NodePool form first, then the legacy form for reference.

## Kafka cluster: KafkaNodePool + Kafka (current)

```yaml
# Controllers (KRaft metadata quorum)
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: controller
  namespace: kafka
  labels:
    strimzi.io/cluster: my-cluster
spec:
  replicas: 3
  roles:
    - controller
  storage:
    type: persistent-claim
    size: 10Gi
    deleteClaim: false
    class: gp3
---
# Brokers
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: broker
  namespace: kafka
  labels:
    strimzi.io/cluster: my-cluster
spec:
  replicas: 3
  roles:
    - broker
  storage:
    type: jbod                 # multiple disks per broker = more parallel throughput
    volumes:
      - id: 0
        type: persistent-claim
        size: 100Gi
        deleteClaim: false
        class: gp3
      - id: 1
        type: persistent-claim
        size: 100Gi
        deleteClaim: false
        class: gp3
  resources:
    requests: { memory: 2Gi, cpu: "1" }
    limits:   { memory: 4Gi, cpu: "2" }
  template:
    pod:
      affinity:                # spread brokers across nodes for HA
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchExpressions:
                  - key: strimzi.io/pool-name
                    operator: In
                    values: ["broker"]
              topologyKey: kubernetes.io/hostname
---
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
  annotations:
    strimzi.io/node-pools: enabled
    strimzi.io/kraft: enabled
spec:
  kafka:
    version: 3.7.0
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
      - name: tls
        port: 9093
        type: internal
        tls: true
      - name: external
        port: 9094
        type: loadbalancer       # or 'route'/'ingress'/'nodeport'
        tls: true
    config:
      default.replication.factor: 3
      min.insync.replicas: 2
      offsets.topic.replication.factor: 3
      transaction.state.log.replication.factor: 3
      transaction.state.log.min.isr: 2
      num.partitions: 6
      log.retention.hours: 168    # 7 days; raise carefully — costs disk
  entityOperator:                 # enables KafkaTopic/KafkaUser management
    topicOperator: {}
    userOperator: {}
```

Apply and watch:

```bash
kubectl apply -f kafka-cluster.yaml -n kafka
kubectl get kafka -n kafka
kubectl get kafkanodepool -n kafka
kubectl get pods -n kafka
```

Key config choices: **RF 3 + `min.insync.replicas=2`** for production durability; **JBOD** for broker throughput; **`deleteClaim: false`** so deleting the `Kafka` CR doesn't destroy data; **podAntiAffinity** so a node loss can't take the quorum.

## Kafka cluster: legacy ZooKeeper form (reference only)

```yaml
# DEPRECATED in current Strimzi — only for old pinned versions.
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: my-cluster
spec:
  kafka:
    version: 3.5.0
    replicas: 3
    listeners:
      - { name: plain, port: 9092, type: internal, tls: false }
      - { name: tls,   port: 9093, type: internal, tls: true }
    config:
      default.replication.factor: 3
      min.insync.replicas: 2
    storage:
      type: jbod
      volumes:
        - { id: 0, type: persistent-claim, size: 100Gi, deleteClaim: false }
  zookeeper:
    replicas: 3
    storage: { type: persistent-claim, size: 10Gi, deleteClaim: false }
  entityOperator:
    topicOperator: {}
    userOperator: {}
```

## Listeners, TLS & external access

A listener defines how clients reach brokers:
- `type: internal` — in-cluster only (use the bootstrap Service `my-cluster-kafka-bootstrap:9092`). Prefer this for apps inside the cluster.
- `type: loadbalancer` / `nodeport` / `ingress` / `route` — external access. Strimzi provisions per-broker addresses + a bootstrap address. **Enable `tls: true` and authentication for anything external.**
- Add `authentication` (e.g. `type: tls` for mTLS, or `scram-sha-512`) on the listener and an `authorization` block (`type: simple`) on `spec.kafka` to enforce ACLs.

Strimzi auto-generates a cluster CA and per-component certs. Clients pull the CA from the secret `my-cluster-cluster-ca-cert`.

## Storage: JBOD and persistent claims

- **`type: persistent-claim`** — one PVC per node from a StorageClass (`class:`). Use a fast SSD/NVMe class.
- **`type: jbod`** — multiple independent volumes per broker ("Just a Bunch Of Disks"); Kafka spreads partitions across them for higher aggregate throughput. Preferred for brokers.
- **`deleteClaim: false`** keeps PVCs when the CR is deleted — essential for not losing data. Set `true` only for ephemeral/test clusters.
- For pure throwaway testing, `type: ephemeral` uses `emptyDir` (data lost on restart) — never for anything you care about.

## KafkaTopic CRD

Manage topics declaratively (requires the `topicOperator`):

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: orders
  namespace: kafka
  labels:
    strimzi.io/cluster: my-cluster
spec:
  partitions: 6
  replicas: 3
  config:
    retention.ms: "604800000"     # 7 days
    cleanup.policy: "delete"      # or "compact" for changelog/keyed topics
```

`partitions` can be increased later but **never decreased**; `replicas` ≤ broker count. Plan partition count for target parallelism (consumers in a group ≤ partitions to all stay busy).

## KafkaUser CRD

Provision client credentials + ACLs (requires `userOperator` and an authenticating listener):

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: orders-app
  namespace: kafka
  labels:
    strimzi.io/cluster: my-cluster
spec:
  authentication:
    type: scram-sha-512           # or 'tls' for mTLS
  authorization:
    type: simple
    acls:
      - resource: { type: topic, name: orders, patternType: literal }
        operations: [Read, Write, Describe]
      - resource: { type: group, name: orders-consumers, patternType: literal }
        operations: [Read]
```

The operator writes the user's credentials/cert into a Secret named after the user, which the client mounts.

## Kafka Connect & KafkaConnector (CDC, sinks)

Strimzi runs Connect as its own CR; connectors are `KafkaConnector` CRs (or REST). Build a Connect image with the plugins you need (JDBC, S3, Elasticsearch):

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
  annotations:
    strimzi.io/use-connector-resources: "true"   # manage connectors via KafkaConnector CRs
spec:
  version: 3.7.0
  replicas: 1
  bootstrapServers: my-cluster-kafka-bootstrap:9093
  tls:
    trustedCertificates:
      - secretName: my-cluster-cluster-ca-cert
        certificate: ca.crt
  config:
    group.id: connect-cluster
    key.converter: org.apache.kafka.connect.json.JsonConverter
    value.converter: org.apache.kafka.connect.json.JsonConverter
    config.storage.replication.factor: 3
    offset.storage.replication.factor: 3
    status.storage.replication.factor: 3
  build:                          # let Strimzi build the plugin image for you
    output:
      type: docker
      image: myregistry/my-connect:latest
      pushSecret: regcred
    plugins:
      - name: debezium-postgres
        artifacts:
          - type: tgz
            url: https://repo1.maven.org/.../debezium-connector-postgres-...-plugin.tar.gz
      - name: confluent-s3
        artifacts:
          - type: zip
            url: https://.../confluentinc-kafka-connect-s3-...zip
```

A **JDBC source connector** (CDC from Postgres into a topic):

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnector
metadata:
  name: jdbc-source
  namespace: kafka
  labels:
    strimzi.io/cluster: my-connect
spec:
  class: io.confluent.connect.jdbc.JdbcSourceConnector
  tasksMax: 1
  config:
    connection.url: "jdbc:postgresql://<db-host>:5432/postgres"
    connection.user: "postgres"
    connection.password: "${secrets:kafka/db-creds:password}"   # via KafkaConnect externalConfiguration
    mode: "timestamp"               # incremental capture by a timestamp column
    timestamp.column.name: "dt_update"
    query: "SELECT * FROM public.customers"
    topic.prefix: "src-customers"
    poll.interval.ms: "1000"
    value.converter: org.apache.kafka.connect.json.JsonConverter
    value.converter.schemas.enable: "true"
```

> **Don't hardcode DB passwords in the CR.** Use `KafkaConnect.spec.externalConfiguration` to mount a Secret and reference it with the config provider (`${secrets:...}`), or Strimzi's `KafkaConnector` config providers. For true log-based CDC (captures deletes, no polling) prefer a **Debezium** source over the JDBC polling connector.

Sink connectors (e.g. `io.confluent.connect.s3.S3SinkConnector`, `io.confluent.connect.elasticsearch.ElasticsearchSinkConnector`) follow the same `KafkaConnector` shape with `topics:` listing the source topics. A full CDC→Spark→Elasticsearch pipeline is in `end-to-end-pipeline.md`.

## Rebalancing with Cruise Control

Adding brokers doesn't move existing partitions automatically. Strimzi integrates **Cruise Control** for balanced partition placement. Enable it and request a rebalance:

```yaml
# In the Kafka CR:
spec:
  cruiseControl: {}
---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaRebalance
metadata:
  name: rebalance
  namespace: kafka
  labels:
    strimzi.io/cluster: my-cluster
spec:
  mode: full          # or add-brokers / remove-brokers
```

Approve the generated proposal: `kubectl annotate kafkarebalance rebalance strimzi.io/rebalance=approve -n kafka`.

## CLI, monitoring, gotchas

Run client tools inside a broker pod:

```bash
# create a topic
kubectl exec my-cluster-broker-0 -n kafka -c kafka -it -- \
  bin/kafka-topics.sh --create --bootstrap-server localhost:9092 \
  --topic test --partitions 3 --replication-factor 3

# console consumer
kubectl exec my-cluster-broker-0 -n kafka -c kafka -it -- \
  bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 \
  --topic src-customers --from-beginning

kubectl get kafkaconnector -n kafka
kubectl describe kafkaconnector jdbc-source -n kafka      # see connector state/errors
```

Gotchas:
- **`replication-factor` ≤ broker count**, and partitions ≤ brokers when you want one replica set per broker.
- **`deleteClaim: false`** or you'll lose data when deleting the cluster.
- **External listeners need TLS + auth** — never expose a plaintext `loadbalancer` listener publicly.
- Strimzi exposes Prometheus metrics via JMX exporter config in the CR — wire it to your monitoring (see **kubernetes-observability**).
- A `KafkaConnector` stuck `NotReady` usually means a missing plugin (rebuild the Connect image), bad credentials, or an unreachable source — check `kubectl describe kafkaconnector`.
