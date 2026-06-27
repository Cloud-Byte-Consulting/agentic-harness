# Scalability, Performance, Reliability & Disaster Recovery

Everything for "how do we scale / stay up / perform / recover" questions. Read this when designing the scalability, availability, reliability, performance, and DR aspects of a system.

## Contents
- Scaling: horizontal vs vertical, predictive vs reactive
- Elasticity and the elastic three-tier example
- Scaling each tier (static content, sessions, database)
- High availability vs fault tolerance vs reliability
- Redundancy, self-healing, and cascading-failure mitigation
- Performance: latency, throughput, concurrency, caching
- Reliability design principles
- RTO and RPO
- Data replication (sync vs async, methods)
- The four DR strategies
- DR best practices

## Scaling: horizontal vs vertical, predictive vs reactive

- **Horizontal scaling** — add more instances behind a load balancer. The default for stateless tiers; cheaper at scale as compute commoditizes. (1,000 req/s on 2 instances → double to 4 instances for ~2,000 req/s.)
- **Vertical scaling** — add CPU/memory to the same server. Cost rises non-linearly and hits a hardware ceiling; commonly used for relational databases until you must shard/partition.
- **Predictive scaling** — for known/seasonal patterns (e.g. e-commerce holidays); pre-provision from historical data.
- **Reactive scaling** — for sudden, unforeseen spikes (flash sales); auto-scaling policies add/remove instances on metrics (e.g. add a server when CPU > 60%, min 3 / max 6). Always set a sane **maximum** — it caps cost and limits damage from a DDoS-driven scale-out.

Both modes require monitoring and data collection to plan.

**Sharding** scales databases by partitioning data across servers on a **shard key** (e.g. usernames A–E on shard 1, F–I on shard 2). The application routes records to the right shard. Plan it before the DB hits its vertical ceiling; it's significant work.

## Elasticity and the elastic three-tier example

Elasticity = scale out under load **and** scale in when load drops, to right-size cost. A typical elastic three-tier design: Route 53 (DNS) → CloudFront (CDN for static content) → Elastic Load Balancing → auto-scaling web tier → auto-scaling app tier → RDS primary + read replica + multi-AZ standby. It expands and contracts automatically across Availability Zones for both elasticity and high availability.

## Scaling each tier

- **Static content** — offload images/video/static HTML to **object storage** (e.g. S3) and serve via a **CDN** (CloudFront/Akamai/Azure CDN/Google CDN) that caches near users, cutting latency and web-server load. Object storage scales independently of compute and is cheap.
- **Sessions** — decouple session state from the app server into an independent layer (e.g. a NoSQL store like DynamoDB/MongoDB) so the app tier can scale horizontally without resetting user progress (cart on mobile, checkout on desktop).
- **Database** — protect the master: move sessions to NoSQL, static content to object storage, and hot queries to a cache (Redis/Memcached). Keep the master for writes; use **read replicas** for reads (e.g. RDS MySQL supports up to 15 read replicas; expect millisecond replication lag). Shard when you outgrow vertical scaling.

A solutions architect evaluates: CDN for static content, load balancing + auto-scaling for servers, and the storage mix (cache, object store, NoSQL, read replicas, sharding) for data.

## High availability vs fault tolerance vs reliability

These are distinct:
- **High availability (HA)** — the application stays *up* (services remain accessible). Achieved by redundancy across isolated locations and automated failover.
- **Fault tolerance** — the application stays up *at full performance* during a failure. Requires full redundancy. Example: 4 servers needed, split 2+2 across two zones → on a zone loss you're 100% available but only ~50% fault-tolerant (degraded). Full fault tolerance needs 2N (4+4), doubling cost.
- **Reliability** — the system *recovers* from failure within a target time and keeps serving correctly.

Match the target to criticality: an e-commerce checkout may justify 100% fault tolerance (degradation = lost revenue); an internal payroll tool can tolerate short degradation. Don't over-architect — HA is directly tied to cost.

## Redundancy, self-healing, and cascading-failure mitigation

Multi-layered redundancy: clusters across racks → multiple data centers in a region → multiple regions/geographies. Add DNS-based global routing with health checks (route to the optimal/healthy location), CDN for static content, load balancers within a region, auto-scaling, and a standby database. **Shallow health checks** catch local host failures; **deep health checks** also catch dependency failures but cost more.

**Self-healing**: define KPIs (requests/sec, page-load time, CPU/memory thresholds), monitor them, and automate remediation (e.g. spin up servers as CPU nears max; replace unhealthy instances).

**Cascading-failure mitigations**: timeouts, traffic rejection under overload, idempotent operations, and circuit breakers (see architecture-principles-and-tradeoffs.md).

## Performance: latency, throughput, concurrency, caching

Performance directly affects revenue; design it at every layer and monitor continuously.

- **Latency** — time for a request/response round trip; influenced by distance, medium, congestion, and per-component delays (CPU↔RAM, disk seek, slow queries). Reduce with CDNs/edge locations, fiber connectivity, DB partitioning/sharding, and caching. You won't hit zero — target within the user's tolerance.
- **Throughput** — data successfully transferred per unit time; bounded by bandwidth, connection quality, and protocols. Lower latency → higher throughput. Disk throughput = average I/O size × IOPS (e.g. 4 KB × 20,000 IOPS ≈ 78 MB/s). Choose the right IOPS for write-intensive workloads.
- **Concurrency vs parallelism** — concurrency handles many tasks by interleaving (single thread, like a traffic signal); parallelism truly runs tasks at once across cores/nodes. Databases must handle concurrent reads/writes safely (commit before another update).
- **Caching** — apply at every layer (CPU cache, OS page cache, DNS cache, browser cache, CDN, app cache, DB query cache). See caching patterns in architecture-principles-and-tradeoffs.md.

Choose the right server size (memory/compute) and storage (IOPS) for the workload; load-test to confirm you meet concurrency and UX targets, noting that auto-scaling introduces a brief latency dip while it reacts.

## Reliability design principles

- **Self-healing via automation** — monitor KPIs and auto-remediate.
- **Quality assurance via automation** — reproduce environments with scripts/IaC, not manual steps (the script *is* the documentation and is reproducible).
- **Distributed systems** — decompose so one component's failure doesn't take down the whole (payment failure shouldn't block order placement). Communication is harder (latency, delivery guarantees, consistency, partial failure) — use circuit breakers.
- **Monitor and add capacity** — don't guess capacity; scale on demand in the cloud.
- **Recovery validation** — test how the system *fails* and recovers, not just the happy path; simulate failures.

## RTO and RPO

Driven by the SLA (e.g. 99.9% availability ≈ 43 min/month downtime).
- **RPO (Recovery Point Objective)** — maximum tolerable *data loss* (e.g. 15 minutes). Drives backup frequency / replication.
- **RTO (Recovery Time Objective)** — maximum tolerable *downtime* to restore service (e.g. 30 minutes). Drives recovery approach.

Example: failure at 10:00, last backup 09:00, restore takes 30 min → RPO = 1 hour, RTO = 30 min. Lower RTO/RPO = higher cost; reserve the lowest values for mission-critical systems (stock trading can lose no data; railway signaling can't be down). Aim to reduce RTO/RPO over time.

## Data replication

Replication copies the primary site to a secondary so the system can fail over. Also speeds up creating test/dev environments.
- **Synchronous** — real-time copy; lowest RPO; expensive (continuous resource use). (e.g. RDS multi-AZ standby is synchronous.)
- **Asynchronous** — copy with lag or on a schedule; cheaper; for systems tolerant of a longer RPO. (e.g. RDS read replicas are asynchronous.)

**Methods**: array-based (built-in, homogeneous arrays; large enterprises), network-based (between heterogeneous arrays via an appliance; costlier), host-based (agent on the host; SMB-friendly but consumes host compute), hypervisor-based (VM-aware, scalable, efficient DR).

## The four DR strategies

Ordered from highest to lowest RTO/RPO (and lowest to highest cost):

1. **Backup and restore** — store machine images and DB snapshots in the DR site/cloud; restore on disaster. Cheapest; highest RTO/RPO. Apply a backup lifecycle (e.g. 90-day active, then archive to cold storage like Glacier, then delete; note PCI-DSS may require 7-year retention).
2. **Pilot light** — keep minimal core services running (e.g. a small replicated database, Active Directory) with machine images ready; on disaster, start and scale the rest. Lower RTO/RPO than backup/restore; cost-effective.
3. **Warm standby** — a full but low-capacity copy of the environment runs 24/7 (e.g. handling 1–5% of traffic for continuous testing); on disaster, route all traffic over and scale up (vertical for DB, horizontal for servers). Faster recovery for critical workloads; can double as a test/staging environment.
4. **Multi-site (hot standby)** — a full-capacity replica actively serves traffic with continuous replication and load balancing across sites; near-zero RTO/RPO; most expensive. Justified for financial, healthcare, and e-commerce where downtime cost is high.

## DR best practices

- Start small — bring up the most business-critical workloads first.
- Apply a data backup lifecycle (archive/delete) to control cost.
- Track software licenses (per install/CPU/user) so scaling doesn't break licensing.
- Plan scaling (horizontal adds licensed instances; vertical adds CPU/memory).
- **Test often** — an untested DR plan will fail; SLA breaches cost money and trust.
- **Play game days** — simulate a disaster on a low-traffic day; let the team recover; verify backups/snapshots/images work.
- Continuously monitor resources to trigger automated failover and avoid resource saturation.

The cloud improves reliability: data centers across geographies, easy backup/image tracking, built-in metrics and change management (e.g. CloudWatch, Systems Manager), easy DR testing, and scalable, self-healing infrastructure.
