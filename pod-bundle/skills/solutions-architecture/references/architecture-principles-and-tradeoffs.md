# Architecture Principles & Pattern Catalog

The pervasive design principles plus the catalog of structural and resilience patterns. Read this when choosing or critiquing a pattern, or when you need the "why" behind a design rule.

## Contents
- The pervasive design principles
- Choosing a structural pattern (decision guide)
- n-tier (layered) architecture
- Multi-tenant SaaS architecture
- SOA and RESTful services
- Microservices
- Event-driven and serverless
- Saga and fan-out/fan-in
- MVC
- Domain-Driven Design (DDD)
- Resilience patterns: circuit breaker, bulkhead, floating IP
- Caching patterns
- Containers
- Anti-patterns

## The pervasive design principles

Apply these across every layer; they are not separate sections of a design.

- **Design for failure, and nothing will fail.** Assume every component fails. Add redundancy, remove single points of failure, health-check everything, and degrade gracefully.
- **Loose coupling.** Put an intermediary (load balancer, queue, event bus, well-defined API) between components so they scale, deploy, and fail independently. Loose coupling shrinks the blast radius — an error is contained to one instance/service.
- **Think service, not server.** Service-oriented thinking (REST/microservices) avoids hardware dependency and tight coupling.
- **Cattle, not pets / immutable infrastructure.** Servers are provisioned quickly, managed consistently, and replaced — not hand-tuned. Build from a golden image; never patch live; spin up new instances for upgrades and troubleshooting (keep logs for RCA before disposing).
- **Statelessness.** Externalize session and config state (e.g. to a NoSQL store) so any instance can serve any request and you can scale horizontally and self-heal.
- **Elasticity over fixed capacity.** Scale out to meet demand and scale in to save cost; don't guess capacity.
- **Security everywhere (defense in depth).** Secure every layer and isolate each from the others; least privilege; reduce blast radius.
- **Automate everything.** Testing, deployment, infrastructure (IaC), monitoring, security, and recovery should be code.
- **Data-driven design.** Most systems revolve around data; work backward from the data model and access patterns.
- **Design for operation.** Plan deployment, monitoring/alerting, runbooks/playbooks, patching, and DR from the start.
- **Build future-proof, extensible, reusable, interoperable, portable architecture.** Favor loosely coupled, API-based modules; choose portable tech (e.g. JSON/XML interchange, cross-platform languages/frameworks) where it earns its keep.

## Choosing a structural pattern (decision guide)

- **Start simple.** A modular monolith or n-tier is the right default. Justify before going distributed.
- Go **microservices** when independent scaling, independent deployment, team autonomy, or fault isolation are concrete needs — and you can pay the operational tax (distributed monitoring, network calls, eventual consistency, service coordination).
- Go **event-driven** when components must react to events asynchronously and you want strong decoupling.
- Go **serverless** when workloads are spiky/intermittent, you want to minimize ops, and per-request billing fits.
- Go **multi-tenant SaaS** when one codebase serves many customer organizations.

## n-tier (layered) architecture

Divide the system into independently implementable, scalable layers — commonly **web/presentation, application/logic, database/data**. Each layer scales independently, can adopt new tech without disturbing others, and is network-isolated (compromise of one layer doesn't grant the others). Most common variant is three-tier. Add a data-access or domain layer for four/five tiers when complexity warrants.

- **Web/presentation** — UI (HTML/CSS, React/Angular, JSP/ASP); collects input, renders responses; UX and page-load performance live here.
- **Application/logic** — business logic, algorithms, recommendations (C++/Java/.NET/Node.js).
- **Database/data** — relational (PostgreSQL, MySQL, SQL Server, Aurora, RDS) and/or NoSQL (DynamoDB, MongoDB, Cassandra); sessions, config, transactions; encrypt at rest and in transit; give it the most security attention.

## Multi-tenant SaaS architecture

One instance of software + infrastructure serves many customers (tenants), each isolated by configuration, identity, and data, while sharing the product for economies of scale. Tenants are identified by a tenant ID. Data isolation options (choose by compliance, security, cost, and contract):
- **Database-level isolation** — a separate database per tenant (strongest isolation; needed when a shared DB is unacceptable).
- **Table-level isolation** — separate tables per tenant (e.g. tenant-ID-prefixed).
- **Row-level isolation** — shared tables with a tenant-ID column; the data-access layer filters by tenant.
SaaS reduces buyer responsibility (the vendor owns maintenance/updates) but limits customization — assess fit for enterprise needs, and compare total cost of ownership for build-vs-buy.

## SOA and RESTful services

SOA spreads operations into independent services communicating over a network protocol; each provides end-to-end functionality. Benefits: parallel development/deployment, independent scaling. Costs: stronger governance and more tooling/automation for monitoring/deployment/scaling. SOAP (XML-only, heavyweight) is now legacy; **REST** is the modern default.

**REST principles**: stateless (each request self-contained), client-server separation, uniform interface (HTTP verbs map to CRUD — GET read, POST create, PUT update, DELETE remove), resource-based (each resource has a URL), multiple representations (JSON/XML/HTML), layered system (caches/load balancers transparent to the client), optional code-on-demand. JSON is the lightweight default interchange format.

## Microservices

Each component is an independent service with its own framework and (ideally) its own database, scaled independently. Convert a monolith by carving small, independent components; modularization reduces the cost, size, and risk of change. Trade-offs: smaller code surface and self-containment vs. more granular, distributed monitoring and operational overhead. Decompose along **bounded contexts** (see DDD), not arbitrarily. Deploy in containers or as serverless functions to limit infra overhead.

## Event-driven and serverless

**Serverless (FaaS)** — no servers to manage; the platform auto-scales and you pay per execution (e.g. AWS Lambda + API Gateway + DynamoDB + S3). Best for event-driven, spiky workloads. Considerations: design functions to trigger on events; respect execution-time/memory limits; emphasize **statelessness**; add distributed tracing/monitoring; manage cold-start and cost; lean on managed services for databases/queues/auth rather than writing functions for everything.

**Event-driven architecture** — components emit and react to events asynchronously via a queue or event bus, decoupling producers from consumers. **Queue-based decoupling** lets producers and consumers work independently and scale the consumer fleet (and auto-scale down when the queue is empty) — e.g. an image-processing pipeline.

## Saga and fan-out/fan-in

- **Saga** — manages a distributed transaction across microservices as a sequence of local steps, each with a compensating action to undo it on failure. Coordination is either **orchestration** (a central coordinator drives the steps) or **choreography** (each service emits an event that triggers the next, e.g. `OrderCreated` → charge → update inventory → notify). Use to maintain consistency without distributed locks.
- **Fan-out/fan-in** — data fans out to multiple parallel processors, then results fan back in for aggregation. Good for collecting and consolidating data from many sources (e.g. real-time analytics over posts, comments, likes). Benefits: parallelism, scalability, modularity (separate collection from aggregation).

## MVC

Separates **Model** (data + business rules), **View** (presentation), and **Controller** (handles input, updates model and view). Benefits: separation of concerns, reusability, maintainability, flexibility (change UI without touching data logic). Widely used from web frameworks to desktop apps.

## Domain-Driven Design (DDD)

A methodology for taming complexity by modeling software around the business domain. Key concepts:
- **Domain** — the problem area (e.g. healthcare management).
- **Ubiquitous language** — shared vocabulary between developers and domain experts.
- **Bounded context** — an explicit boundary encapsulating one part of the domain (e.g. Patient Management, Appointment Scheduling, Billing); maps well to a microservice.
- **Entities** (identity over time), **value objects** (immutable, no identity), **aggregates** (a cluster treated as a unit, accessed only through an aggregate root), **repositories** (retrieve aggregates from storage), **factories** (create complex objects in a valid state), **services** (operations not belonging to an entity/value object), **domain events** (something significant happened), **anti-corruption layer** (translate between models/external systems).

## Resilience patterns

- **Circuit breaker** — wrap calls to a downstream dependency; if a threshold of failures occurs over an interval, "open" the circuit and fail fast (throwing immediately) for a timeout, rather than retrying and exhausting threads (which causes cascading failure). Periodically let a few requests through to test recovery, then "close" when healthy. Track state in a low-latency store (DynamoDB/Redis/Memcached).
- **Bulkhead** — partition the system (like watertight ship compartments) so failure or overload in one partition doesn't sink the whole system. Isolate high-dependency services into pools; pick a useful granularity; monitor each partition's SLA; isolate critical consumers from standard ones.
- **Floating IP / ENI** — for legacy apps with hardcoded IP/DNS, move the network interface (e.g. AWS Elastic IP / Elastic Network Interface) to a replacement instance so it assumes the old server's identity — enabling upgrades and fast rollback without changing the address.
- **Cascading-failure mitigations** generally: timeouts, traffic rejection under overload, idempotent operations, circuit breakers.

## Caching patterns

Cache temporarily stores data between requester and source to speed responses and cut bandwidth/DB load. Apply at every layer: client/browser (HTTP cache-control, TTL, cookies), DNS, CDN/edge (static content near users), application, and database. Use a **dedicated, centralized caching layer** (independent of the app lifecycle) so a server crash doesn't lose the cache and all horizontally scaled servers share it.

- **Cache distribution** — caching across web/app/DB tiers (e.g. CloudFront edge + ElastiCache app cache + DynamoDB session store).
- **Rename distribution** — to update CDN content before TTL expires, publish the new file under a new name and update the URL (cheaper than invalidation).
- **Cache proxy / rewrite proxy** — a caching/rewriting proxy (NGINX/Apache) in front of the app servers; deliver or redirect content without modifying the app.
- **App caching: lazy (cache-aside) vs write-through.** Lazy fills the cache on a miss — best for read-heavy data tolerant of slight staleness (product catalog). Write-through writes cache and DB together — best for write-heavy data needing immediate consistency (user reviews shown instantly).
- **Engines**: Memcached (multithreaded, simple key-value, fast, no persistence, easy ops) vs Redis (single-threaded, rich data structures, persistence/replication, more complex). Choose Redis for persistence/advanced types (leaderboards), Memcached for simple high-performance string/JSON caching. Always set TTL and an eviction policy; aim for a high cache-hit ratio.

## Containers

Containers package an app with its dependencies and isolate at the kernel level (vs VMs at the OS level), enabling many apps per host with quick startup and efficient resource use. Benefits: portable runtime, fast dev/deploy cycles, dependency packaging, multiple app versions side by side, automatable, better utilization. Orchestrate with Kubernetes (or Docker tooling); managed control planes (e.g. EKS/ECS, with serverless node options like Fargate) remove infra burden. Watch for stateful apps that store files locally or rely on sticky sessions when containerizing — externalize state (e.g. session state in Redis).

## Anti-patterns

See the SKILL.md "Common pitfalls" section. In short: skipping NFRs; over-architecting; premature microservices; tight coupling and hardcoded endpoints; stateful app servers; pet servers / config drift; untested DR; lift-and-shift then stop; cost as an afterthought; caching without TTL/eviction; accidental vendor lock-in.
