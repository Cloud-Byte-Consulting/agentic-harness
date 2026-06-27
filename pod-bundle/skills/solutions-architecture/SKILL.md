---
name: solutions-architecture
description: >-
  Practice cloud solution architecture: design, evaluate, and document systems and make
  defensible technical trade-offs. Use when translating requirements into a design, choosing a
  pattern (n-tier, microservices, event-driven, serverless, SOA), reasoning about quality
  attributes (scalability, availability, reliability, performance, security, cost) or
  well-architected reviews, picking storage or databases (SQL vs NoSQL, cache, object store,
  data lake), planning resilience and disaster recovery (RTO, RPO, multi-AZ, failover),
  planning a cloud migration (the 6 Rs, rehost, replatform, refactor), setting up DevOps, IaC,
  or CI/CD, optimizing cost or FinOps (right-sizing, savings plans, tagging, show-back), or
  writing an ADR, solution architecture document, or HLD. Triggers include "design a system",
  "architecture review", "should we use microservices", "how do we scale this", "migrate to
  the cloud", "reduce our cloud bill", "non-functional requirements". Cloud-broad with
  AWS-leaning examples.
---

# Solutions Architecture

This skill equips Claude to act as a senior solutions architect: turn business needs into a sound technical design, reason explicitly about quality-attribute trade-offs, choose patterns and technologies that fit constraints, and communicate the result so stakeholders can act on it.

## When to use this skill

- Translating business or functional requirements into a technical design, or surfacing the **non-functional requirements (NFRs)** a team forgot.
- Choosing or critiquing an **architecture pattern** (n-tier, microservices, event-driven, serverless, SOA, MVC, DDD) for a given problem.
- Reasoning about a **quality attribute** — scalability, availability, reliability, performance, security, cost — or doing a well-architected-style review.
- Selecting **data storage**: relational vs NoSQL, caching, object storage, data lake, read replicas, sharding.
- Planning **resilience / disaster recovery**: RTO/RPO targets, multi-AZ/multi-region, failover, backup strategies.
- Planning a **cloud migration**: portfolio discovery, the 6 Rs, hybrid connectivity, cutover.
- Setting up **DevOps**: CI/CD, infrastructure as code, configuration management, DevSecOps.
- **Cost optimization / FinOps**: right-sizing, pricing models, tagging, show-back/charge-back, TCO.
- Producing **architecture documentation**: a solution architecture document (SAD), ADR, or HLD; communicating a design to mixed audiences.

This is design-and-decision work. For deep implementation of a specific tool (e.g. authoring Kubernetes manifests, writing a Terraform module), defer to the relevant specialist skill and keep this skill focused on the architecture-level decision.

## Core concepts

**The architect's job is to manage constraints, not eliminate them.** Every design balances competing forces: cost, quality, time, scope, technology, risk, resources, and compliance. Pushing on one almost always inflates another (more caching layers → more cost; tighter schedule with fewer people → lower quality → more rework). Make the trade-off *explicit* and tie it back to the business goal. There is no objectively "best" architecture — only the best fit for the stated constraints and quality-attribute priorities.

**Separate functional requirements from non-functional requirements.** FRs say *what* the system does (place an order, reset a password). NFRs say *how well* it must do it — performance, scalability, availability, reliability, security, maintainability, usability, portability. NFRs are invisible to users until they're missing, and they're the first thing dropped under schedule pressure. A senior architect's signature move is asking the NFR questions nobody else asked: "What's the acceptable downtime?" "What load at peak?" "What data loss can we tolerate?" Quantify them (99.9% availability ≈ 43 min/month; catalog page < 3 s; tolerate 15 min of data loss).

**The quality attributes interact and have costs.** High availability (the app is up) is not fault tolerance (the app is up *at full performance*) is not reliability (the app *recovers* from failure). Each level costs more. Don't over-architect: a revenue-driving e-commerce site may need 100% fault tolerance; an internal payroll tool can tolerate degraded performance for an hour. See `references/scalability-and-reliability.md`.

**Design for failure, and nothing will fail.** Assume every component fails. Build in redundancy, remove single points of failure, make servers replaceable ("cattle, not pets"), keep components loosely coupled so one failure doesn't cascade. Loose coupling (load balancers, queues, well-defined APIs) and statelessness are what make scaling, self-healing, and graceful degradation possible.

**Work backward from data and from the customer.** Almost every system is fundamentally about collecting, storing, and serving data — let the data shape the design. And start from customer needs: deliver a minimum viable product (use MoSCoW prioritization), get it in users' hands, and iterate, rather than building everything speculatively.

**Prefer managed and cloud-native where it earns its keep.** On a cloud platform, lean into elasticity (pay-as-you-go, scale to demand), managed services (offload undifferentiated heavy lifting), automation everywhere, and global infrastructure — but weigh provider lock-in deliberately.

## Workflow: how to approach an architecture task

1. **Establish context and constraints first.** Who are the stakeholders (internal and external) and what does each one care about? What are the business goals, the budget, the timeline, the compliance regime, the existing systems you must integrate with? Capture constraints before proposing anything. Watch for **scope creep** — the silent killer of all other constraints.

2. **Gather and quantify requirements.** Separate FRs from NFRs. For each NFR, ask the pointed question and pin a number to it. Surface assumptions, dependencies (upstream and downstream), and risks explicitly. See `references/architect-role-and-process.md`.

3. **Prioritize the quality attributes.** You can't max everything. Decide — with the business — what ranks highest (e.g. "cost over performance for this internal tool"; "availability and security over time-to-market for this banking feature"). This ranking drives every subsequent decision.

4. **Choose patterns and technologies against the criteria.** Map requirements to a pattern (see `references/architecture-principles-and-tradeoffs.md` and `references/data-architecture.md`). Define explicit selection criteria (integration needs, performance, security, team skills, scalability headroom). Build a small **proof of concept** to de-risk an unproven choice before committing. Pick technology that satisfies today's requirements *and* leaves room to scale.

5. **Apply the cross-cutting principles to every layer.** Scalability/elasticity, loose coupling, statelessness, immutability, security-in-depth, automation, observability, and operability are not separate sections — apply them throughout. Reduce the blast radius of any failure or breach by isolating layers and granting least privilege.

6. **Analyze trade-offs out loud and decide.** For each significant decision, state the options, the pros/cons, why you chose what you chose, and what you gave up. This *is* the value you add. Record it (an ADR or the "key architecture decisions" section of a SAD).

7. **Design for operation and the long term.** How is it deployed, monitored, patched, recovered, and scaled in production? Define RTO/RPO and a DR strategy. Plan logging, monitoring, alerting, runbooks, and playbooks. Plan for change: small increments, rollback, blue-green/canary deploys via CI/CD.

8. **Document and communicate.** Produce the right artifact for the audience — a SAD with multiple views (business, logical, process, deployment, implementation, data, operational), an ADR for a single decision, or a focused diagram. Write so a non-technical stakeholder can follow the *why* and a developer can follow the *what*. See `references/documentation-and-communication.md`.

## Common pitfalls & anti-patterns

- **Ignoring NFRs until late.** Teams over-focus on features; the architect must champion scalability, availability, security, and operability from day one. Retrofitting them is expensive or impossible.
- **Over-architecting.** Building 100% fault tolerance, multi-region DR, or microservices for a low-traffic internal app burns money and time for value nobody needs. Match the investment to the actual business criticality.
- **Premature microservices.** Microservices add real operational overhead (distributed monitoring, network calls, eventual consistency, service coordination). Don't decompose a small, well-understood app just because microservices are fashionable; a modular monolith is often the right call. Decompose along **bounded contexts**, not arbitrarily.
- **Tight coupling and hardcoded endpoints.** Hardcoded IPs/DNS names, direct server-to-server dependencies, and shared mutable state block scaling and make failures cascade. Insert load balancers, queues, and APIs; keep state in an external store.
- **Stateful application servers.** Storing session state on the server forces sticky sessions and breaks horizontal scaling. Externalize session state (e.g. a NoSQL store) and keep servers stateless.
- **Treating servers as pets.** Manually tuned, irreplaceable servers cause configuration drift and slow, risky changes. Use immutable infrastructure built from golden images via IaC.
- **Single point of failure / no recovery testing.** Redundancy you never fail-test is a guess. Validate failure and recovery paths regularly; a DR plan that's never exercised will fail when you need it.
- **Lift-and-shift then stop.** Rehosting realizes some cloud value, but skipping subsequent optimization leaves cost and agility on the table. Migrate first if needed, but plan the optimization phase.
- **Cost as an afterthought.** Cost is everyone's responsibility across the whole lifecycle. Calculate TCO (CapEx + OpEx), not just upfront price; tag resources; monitor against budget/forecast. Cost optimization ≠ blind cost cutting that harms the customer experience.
- **Caching without an eviction/TTL strategy.** Caches that never expire serve stale data; caches without a clear hit-strategy add complexity for little gain. Decide TTL, eviction, and lazy vs write-through deliberately.
- **Vendor lock-in by accident.** Cloud-native proprietary services are powerful but bind you to one provider. Make the lock-in trade-off consciously, not by default.

## Reference files

- **`references/architect-role-and-process.md`** — The solutions architect role and its variants, the solution delivery lifecycle, requirement gathering, FR vs NFR, stakeholder engagement, constraints, POCs, MVP/MoSCoW, Agile architecture. Read when scoping a role, a process, or a requirements exercise.
- **`references/architecture-principles-and-tradeoffs.md`** — The design principles (loose coupling, statelessness, immutability, service-not-server, automation, design-for-operation) and the catalog of patterns (n-tier, SOA, REST, microservices, event-driven, serverless, queue-based, CQRS-style, circuit breaker, bulkhead, MVC, DDD, multi-tenant SaaS, containers). Read when choosing or critiquing a pattern.
- **`references/scalability-and-reliability.md`** — Scaling (horizontal/vertical, predictive/reactive, elasticity, sharding), high availability vs fault tolerance vs reliability, redundancy, self-healing, caching, RTO/RPO, replication, and the four DR strategies. Read for any "how do we scale / stay up / recover" question.
- **`references/security-and-compliance.md`** — Security design principles, identity (IAM, SSO, FIM, SAML, OAuth/OIDC, Kerberos, AD), defense in depth, data protection (rest/transit/in-use), shared responsibility model, threat modeling, DevSecOps, compliance regimes. Read for any security or compliance question.
- **`references/data-architecture.md`** — Choosing storage (relational, NoSQL, object, in-memory cache, data warehouse, data lake), the big-data pipeline (ingest/store/process/analyze/visualize), FLAIR principles, batch vs stream, Redis vs Memcached. Read when selecting data stores or designing analytics.
- **`references/cloud-migration.md`** — Public/private/hybrid/multi-cloud, IaaS/PaaS/SaaS/FaaS, the 6 Rs, the migration lifecycle (discover→optimize), data and server migration techniques, cutover strategies, hybrid connectivity, CloudOps. Read for any migration question.
- **`references/cost-optimization.md`** — Cost design principles, TCO, budget vs forecast, demand/service-catalog management, show-back/charge-back, right-sizing, cloud pricing models (on-demand, savings plans/reserved, spot), tagging and account structure, FinOps, green IT. Read for any cost/FinOps question.
- **`references/documentation-and-communication.md`** — The SAD (purpose, the seven views, structure), ADRs, IT procurement docs (RFP/RFI/RFQ), diagramming, and communicating to mixed audiences. Read when producing or reviewing architecture documentation.
- **`references/devops-and-operations.md`** — DevOps and automation for the architect: CI/CD pipelines, infrastructure as code, deployment strategies (blue-green, canary, rolling), monitoring and observability, and running systems in production (day-2 operations). Read for any DevOps, CI/CD, IaC, or operational-readiness question.
