# Data Architecture & Storage Selection

Choosing the right data store and designing the data/analytics pipeline. Read this when selecting databases/storage or designing big-data/analytics.

## Contents
- How to choose a data store (the questions)
- Relational databases (OLTP) and ACID
- Data warehouses (OLAP)
- NoSQL: types and SQL-vs-NoSQL trade-offs
- Search, object, streaming, vector, blockchain stores
- Caching engines
- The big-data pipeline (ingest → store → process → visualize)
- FLAIR principles and batch vs stream
- Data lake vs data warehouse

## How to choose a data store

There is no single store for all needs — combine stores, balancing latency against cost ("right tool for the right job"). Decide by:
- **How structured is the data?** Well-formed schema (weblogs, protocols) → relational/columnar; arbitrary binary (images/audio/video/PDF) → object storage; semi-structured with variability (JSON/CSV) → NoSQL.
- **How fast must new data be queryable?** Real-time (recommendations, fraud) vs near-real-time (engagement emails) vs batch (monthly reports, model training).
- **Ingest size?** Per-record (REST payloads), large batches (system integrations/feeds), or micro-batches (clickstream).
- **Total volume and growth?** GB/TB vs PB/EB; rolling window vs full history.
- **Cost to store and query** at that location — performance + resilience cost more.
- **Query type?** Fixed dashboard metrics, large numerical aggregations, or full-text search/pattern analysis.

## Relational databases (OLTP) and ACID

Row-based, best for **online transaction processing** with complex joins (e-commerce, banking, bookings). Examples: Oracle, MySQL, MariaDB, PostgreSQL, SQL Server, Amazon Aurora/RDS. Must satisfy **ACID**: Atomicity (all-or-nothing), Consistency (committed valid state), Isolation (concurrent transactions don't interfere), Durability (survive interruption). Scale vertically; add read replicas and a cache for read load; shard when needed. Row format = fast writes, slower analytic reads (scans irrelevant columns).

## Data warehouses (OLAP)

Central repositories of current + historical structured data, optimized for **online analytical processing** — large reads, aggregation, and summarization for BI. Modern warehouses use **columnar storage** (better compression, selective column reads, faster aggregation) and **massively parallel processing (MPP)** across nodes. Examples: Amazon Redshift, Snowflake, Google BigQuery (columnar); Netezza, Teradata, Greenplum (older row-based). Loaded in batches; not for high-concurrency writes or real-time hot data. Traditional warehouses struggle with diverse data (text, images, IoT, audio/video) and with ML's direct, non-SQL data access — hence data lakes.

## NoSQL: types and SQL-vs-NoSQL trade-offs

Non-relational; no enforced schema (rows can have different attributes); accessed via a **partition key**; highly distributed, replicated, horizontally scalable. Solves the scaling/performance limits of relational DBs. Examples: DynamoDB, Cassandra, MongoDB. May trade some ACID guarantees for horizontal scale and flexibility.

**Types**:
- **Columnar** (Cassandra, HBase) — scan a column, not the whole row; good for wide tables with targeted aggregates.
- **Document** (MongoDB, Couchbase, DynamoDB, DocumentDB) — semi-structured JSON/XML.
- **Graph** (Amazon Neptune, Neo4j, JanusGraph, OrientDB) — vertices and edges; relationship-heavy queries.
- **In-memory key-value** (Redis, Memcached) — heavy-read caching, sessions, hot profiles.

**SQL vs NoSQL summary**: SQL = normalized relational model, full ACID, vertical scale (sharding for distribution), schema-on-write. NoSQL = flexible schema, often relaxed ACID, horizontal scale on commodity clusters, performance tied to cluster size/network/access pattern.

## Search, object, streaming, vector, blockchain stores

- **Search** (Elasticsearch / Amazon OpenSearch; ML search like Amazon Kendra) — full-text/log search and analysis over warm data; ad hoc queries across many attributes including string tokens; Kibana for visualization.
- **Object storage** (Amazon S3, Azure Blob, Google Cloud Storage) — objects in buckets with a flat namespace, accessed via API (GET/PUT); data + metadata together; effectively unlimited; the go-to foundation for cloud **data lakes** because it decouples storage from compute. Not a filesystem (latency, no file locking).
- **Streaming** (Apache Kafka, Flink, Spark Structured Streaming, Samza; Amazon Kinesis/MSK) — continuous, high-velocity data with no defined end; decouples producers from consumers and provides a replayable buffer. Kinesis: Data Streams (raw stream), Data Firehose (deliver to S3/Redshift/OpenSearch/Splunk), Data Analytics (Flink-based analytics).
- **Vector (VectorDB)** — stores high-dimensional embeddings for similarity/nearest-neighbor search; powers semantic search, recommendations, and RAG/GenAI. Pros: fast similarity search, scales with data, integrates with ML pipelines. Cons: complexity, resource-intensive, emerging ecosystem.
- **Blockchain / ledger** (Amazon QLDB, Managed Blockchain, Hyperledger, Ethereum/Corda) — immutable, cryptographically verifiable, decentralized records; public/private/consortium networks; for integrity-critical use (land registry, healthcare records, supply chain).

## Caching engines

Redis vs Memcached: Memcached is multithreaded, simple key-value, fast, no persistence, easy ops; Redis is single-threaded, supports rich data structures, persistence/replication, more complex. Choose Redis for persistence/advanced types (leaderboards, live voting); Memcached for simple high-performance string/JSON caching. Set TTL and eviction; aim for a high cache-hit ratio. (More caching patterns in architecture-principles-and-tradeoffs.md.)

## The big-data pipeline

General flow: data → **ingest** → **store** → **process/analyze** → **visualize/serve** → insight. **Decouple** the stages (don't run the whole pipeline on one tool) for fault tolerance and the right cost/throughput balance per stage.

- **Ingest** — categories: databases (OLTP sources), streams (clickstream/IoT via Kafka/Fluentd/Kinesis), logs, files. Tools: Apache DistCp, Sqoop, Flume; cloud: AWS DMS/Direct Connect/Snowball, GCP Storage Transfer/Pub-Sub/Dataflow, Azure Data Factory/Event Hubs.
- **Store** — pick per the questions above; data lake on object storage decouples compute/storage.
- **Process/analyze** — **batch** (large cold data, hours — Hadoop/MapReduce/EMR, Hive, Pig, Presto, Spark, HBase) vs **stream** (small hot data, real-time — Kafka/Flink/Spark Streaming/Kinesis). Spark is in-memory, DAG-based, partition-aware. ETL/data-lake pipeline example: sources → S3 → EMR (Hive/Pig/Spark) transform → Redshift, with Athena for ad hoc S3 queries and QuickSight for visualization.
- **Visualize** — Amazon QuickSight, Kibana, Tableau, Power BI, Spotfire, Jaspersoft.

## FLAIR principles and batch vs stream

When designing a data architecture, apply **FLAIR**: **F**indability (locate assets + metadata), **L**ineage (trace origin and flow), **A**ccessibility (credentials + network to reach data), **I**nteroperability (formats usable across systems), **R**eusability (documented schema, attributed source, MDM). Start from the **user personas** (teams, proficiency, tools, employee/customer/partner) and their access patterns and retention needs.

## Data lake vs data warehouse

A **data warehouse** handles only structured relational data for BI; a **data lake** (object-storage-backed) stores structured *and* unstructured data (JSON logs, CSV, images, audio, video) and serves analytics, ML, and ad hoc querying — decoupling compute and storage for flexibility and cost. Use both: warehouse for governed structured reporting, lake for the breadth of raw and processed data.
