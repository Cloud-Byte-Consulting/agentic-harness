---
name: dotnet-enterprise-architecture
description: >-
  Architect enterprise systems on .NET (C#/ASP.NET Core). Use when choosing or applying an
  architecture style (layered/N-tier, clean, onion, hexagonal/ports-and-adapters, modular
  monolith, microservices), doing domain-driven design (entities, value objects, aggregates,
  bounded contexts, ubiquitous language, domain events), applying CQRS, MediatR or a
  mediator/handler pipeline, designing messaging and event-driven integration (RabbitMQ, AMQP,
  MassTransit, webhooks, outbox, eventual consistency), persistence (repository, unit of work,
  event sourcing), cross-cutting concerns (Serilog/ILogger, FluentValidation, Polly
  resilience, retries, circuit breakers), REST/OpenAPI integration design, quality-attribute
  trade-offs, or ADRs. Triggers: structuring a .NET solution, aggregate/bounded-context
  modeling, splitting a monolith, coupling review, SQL vs NoSQL choice. For C# language
  features use csharp-dotnet-fundamentals; for Blazor/Razor/minimal-API web mechanics use
  dotnet-web-development.
---

# .NET Enterprise Architecture

This skill equips Claude to design, evaluate, and evolve enterprise systems on .NET: choosing an architecture style, modeling the domain, separating concerns cleanly, integrating services, and recording the trade-offs behind each decision.

## When to use this skill

- Structuring a new .NET solution or service and deciding on a style (layered, clean/onion/hexagonal, modular monolith, microservices).
- Modeling a business domain with DDD: finding bounded contexts, defining aggregates/entities/value objects, settling the ubiquitous language.
- Applying CQRS, a mediator/handler pipeline (e.g. MediatR), or splitting read and write models.
- Designing messaging / event-driven integration (RabbitMQ, AMQP, MassTransit, webhooks) and handling eventual consistency, the outbox pattern, and idempotency.
- Choosing persistence (EF Core vs a document store, repository / unit-of-work, event sourcing) and deciding SQL vs NoSQL per service.
- Adding cross-cutting concerns: structured logging, validation, resilience (Polly), correlation/tracing.
- Designing or reviewing REST APIs, OpenAPI contracts, versioning, and inter-service contracts.
- Reasoning about quality attributes (scalability, evolvability, testability) and the trade-offs between them.
- Writing an architecture decision record (ADR) or critiquing an existing architecture for coupling/cohesion problems.

Boundary: C# language features, async, LINQ, generics belong to the **csharp-dotnet-fundamentals** skill. ASP.NET Core web mechanics (minimal APIs, MVC/Razor, Blazor, middleware plumbing, auth wiring) belong to the **dotnet-web-development** skill. This skill is about *system shape and decisions*, not language or framework mechanics.

## Core mental model

Architecture is the set of decisions that are expensive to change. Optimize for **evolvability**: keep the part that captures the business (the domain) free of technical dependencies, and put indirection (interfaces, contracts, messages) around everything that touches the outside world. Almost every good rule below is a corollary of two ideas:

1. **The business model is the core and depends on nothing.** Data structures representing business concepts plus the rules over them sit at the center. Frameworks, databases, UI, and messaging are details that depend *on* the core, never the reverse (dependency inversion). This is what clean / onion / hexagonal all say — the same idea drawn differently. Don't argue about the picture; enforce the dependency direction.

2. **Decide functionally before you decide technically.** Model the domain in business terms first; commit to a database, an ORM, or a transport as late as possible. A design bound to `VARCHAR(50)`, to SQL identity columns, or to a particular broker has *baked in* accidental constraints that fight every future change. Think of data, not databases; think of models and rules, not tables and columns.

### Architecture styles at a glance

- **Layered / N-tier** — presentation → application → domain → infrastructure. Simple and familiar. The classic failure mode is layers (or whole apps) integrating *through* the database, which is catastrophic coupling. Use it, but enforce that nothing skips the application/domain layer to reach data directly.
- **Clean / Onion / Hexagonal (ports & adapters)** — domain-centric; dependencies point inward; outer rings (UI, persistence, messaging) implement interfaces (ports) defined by the core. The default for any service whose business logic will outlive its tech stack.
- **Modular monolith** — one deployable, strict module boundaries (ideally aligned to bounded contexts), in-process calls across well-defined interfaces. Usually the *right starting point*: most of the decoupling benefit, almost none of the distributed-systems cost.
- **Microservices** — independently deployable services per bounded context, each owning its data, communicating via APIs/messages. Worth it only under specific conditions (high volume, frequent independent change, clear team boundaries). The "micro" is a red herring — the real question is **granularity**, and DDD bounded contexts are how you find it. Don't reflexively swing "back to the monolith" either; find the right grain.

> Rule of thumb: start as a modular monolith with clean internal boundaries; extract a service only when a concrete force (scaling, team autonomy, independent release cadence) demands it. Premature microservices cost more than they save.

See `references/architecture-styles.md` for diagrams, decision criteria, C4 modeling, and the monolith-vs-microservices analysis.

## How to approach an enterprise architecture task

1. **Understand the domain in business language first.** Gather the experts, build the ubiquitous language, sketch the business capabilities. Resist jumping to entities-and-tables. Watch for *semantic* bugs — e.g. modeling "customer" and "supplier" as separate tables when they are really *roles* (business rules) over a single "actor" entity. Such mistakes get frozen into schemas and APIs and become almost impossible to fix later. (See `references/domain-driven-design.md`.)

2. **Find the bounded contexts and aggregates.** Decompose by business subdomain, not by technical layer. Each major entity with its own life cycle (a Book, an Author) is a candidate aggregate and a candidate service/module boundary. Minor entities (an address, a tag) live *inside* an aggregate and disappear with it. One bounded context → one ubiquitous language → ideally one module/service → one API contract → one repo → one deployable.

3. **Design the contracts (ports) before the implementations.** Define the API/OpenAPI contract and the domain interfaces first (contract-first), not generated-from-code-last. Reuse standards wherever they exist (ISO codes, RFCs, OpenID Connect, OData, JSON Patch, CMIS, BPMN/DMN). A standard is "technically written but not technically restrictive" — precise, yet not bound to your stack. Where no standard exists, define a **pivotal format** (a neutral, business-described schema) and fall back on existing norms for its inner fields. (See `references/api-and-integration.md`.)

4. **Choose persistence to fit the data shape, not habit.** Most business entities are naturally tree-shaped documents → a document store often fits better than relational tables and avoids the transaction/lock/ORM tax. Use SQL where the data is genuinely tabular or where reporting/BI needs it. Consider storing **operations (deltas), not states**, when history/traceability matters — it gives you time-travel, audit, and lock-free writes, with a cached "best-so-far" state for fast reads. (See `references/persistence-patterns.md`.)

5. **Separate reads from writes when they diverge (CQRS).** When read and write workloads differ in shape, volume, or consistency needs, split them. Pair with a mediator/handler pipeline so each command/query is a single-responsibility handler and cross-cutting concerns (validation, logging, transactions) become pipeline behaviors. Reach for event sourcing only when the audit trail / replay is a real requirement. (See `references/cqrs-and-mediator.md`.)

6. **Integrate with the least coupling that works.** Prefer direct REST calls behind an interface/URL (dependency inversion via configuration) for query-style reads; prefer **events/messages** (webhooks, a message broker) for "something happened" notifications and to decouple producers from consumers. Standardize message payloads functionally so the broker doesn't have to translate. Plan for **eventual consistency**, idempotency, and delivery robustness (retries, outbox, dead-letter). (See `references/microservices-and-messaging.md`.)

7. **Add cross-cutting concerns as decoupled, standardized seams.** Structured logging behind `ILogger` (Serilog as a sink), validation at the boundary (FluentValidation / data annotations), resilience with Polly (retry, circuit breaker, timeout) on every outbound call, and a correlation id threaded through every hop for distributed tracing. (See `references/cross-cutting-concerns.md`.)

8. **Name the quality attributes and trade-offs, then record the decision.** State which attributes you are optimizing (evolvability, scalability, availability, consistency, security) and what you are trading away. Capture each significant choice in an **ADR** so the *why* survives. (See `references/adrs-and-quality-attributes.md`.)

## SOLID at system scale

The SOLID principles apply to services and modules, not just classes:

- **Single Responsibility** → each module/service owns exactly one business reason to change (one bounded context). If a class or service changes for both "author management" and "book management" reasons, split it.
- **Open/Closed** → extend behavior without breaking existing consumers. For APIs: add new fields/endpoints or a new versioned contract; never silently change existing semantics.
- **Liskov** → a v1.1 API must behave like v1.0 for v1.0 calls (`/print` with no color still prints black). Backward compatibility is Liskov for services. Beware **Hyrum's Law**: every observable behavior, even undocumented, *will* be depended on.
- **Interface Segregation** → split fat contracts (`/api/books` vs `/api/books/sales`) so a consumer/implementer isn't forced to depend on operations it doesn't use; it also eases independent versioning.
- **Dependency Inversion** → both the caller and the implementation depend on a *contract* (an interface in code, an OpenAPI/pivotal-format contract between services), not on each other. In services this indirection is the URL-from-config or a `/subscribe` callback registration.

## Common pitfalls & anti-patterns

- **Designing the database first.** Letting `VARCHAR(n)`, SQL identity counters, or date-vs-instant SQL types leak into the domain model. Model functionally; pick storage after.
- **Bad semantics frozen into schema.** Treating role-derived states (customer, prospect, supplier) as entity types instead of business rules over a shared entity — then "fixing" it with DB triggers that loop and crash. Model the rule, not the snapshot.
- **Integrating at the database layer / shared database.** The single biggest cause of legacy systems that can't evolve. Every cross-app interaction goes through an explicit contract (API or message), never a shared table. Nothing but the owning service touches a referential's database.
- **Premature microservices.** Splitting too fine, then spending more effort on plumbing (endpoints, contracts, integration tests) than the monolith ever cost. Find the grain with DDD; default to a modular monolith.
- **Recreating coupling after a clean split.** A reporting/aggregation service legitimately reads from many services; the anti-pattern is an atomic service then *consuming back* from the aggregator — circular dependency and hidden coupling. Keep data flow one-directional.
- **Generated-last API contracts.** Auto-generating OpenAPI from controllers produces monolithic, un-segregated contracts. Go contract-first.
- **`PUT`-replacing whole resources when history matters.** Destroys eventual-consistency benefits and the audit trail. Use `PATCH` (JSON Patch, RFC 6902) to record deltas; treat create as a patch from empty; treat delete as a status change (archive), with true deletion only where regulation (e.g. GDPR) demands it.
- **Over-engineering small services.** A ~1000-line service does not need full hexagonal ceremony or dependency inversion on a standardized driver. Use interfaces where mocking/swapping is plausible; otherwise keep it simple. Match the ceremony to the size (YAGNI).
- **Externalizing rules/processes you don't need.** A full BRMS or BPMN engine is overkill ~99% of the time — most business rules belong in a well-placed function or a small handler. Externalize only when rules change very frequently or carry heavy regulatory/traceability needs.

## Reference files

- `references/architecture-styles.md` — Layered, clean, onion, hexagonal, modular monolith, microservices; C4 modeling; decision criteria; granularity and the monolith debate.
- `references/domain-driven-design.md` — Ubiquitous language, bounded contexts, entities/value objects/aggregates, domain events, entity life cycle and time, semantics pitfalls.
- `references/cqrs-and-mediator.md` — Command/query separation, mediator + handler pipeline (MediatR), pipeline behaviors, event sourcing, when to use each.
- `references/microservices-and-messaging.md` — Service granularity, REST vs messaging, RabbitMQ/AMQP, MassTransit, webhooks, outbox, eventual consistency, idempotency, orchestration vs choreography.
- `references/persistence-patterns.md` — Repository & unit of work, EF Core vs document stores, SQL-vs-NoSQL, operation-vs-state storage, master data management, identifiers.
- `references/api-and-integration.md` — REST/OpenAPI contract-first design, versioning & compatibility, pivotal formats and standards, history in APIs (value dates, JSON Patch), API gateways.
- `references/cross-cutting-concerns.md` — Logging (ILogger/Serilog), validation (FluentValidation), resilience (Polly v8 pipelines), correlation/tracing, externalizing rules/authorization.
- `references/adrs-and-quality-attributes.md` — Quality attributes & trade-offs, ADR template and workflow, fitness functions, evaluating an existing architecture.
