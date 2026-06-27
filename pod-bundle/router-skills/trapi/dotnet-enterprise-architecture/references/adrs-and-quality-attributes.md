# Quality attributes, trade-offs, and ADRs

Contents:
- Why decisions, not diagrams, are the architecture
- Quality attributes that matter
- Naming trade-offs explicitly
- Architecture decision records (ADRs)
- An ADR template
- Worked ADR example
- Fitness functions: enforcing decisions over time
- Evaluating an existing architecture
- Technical debt

## Why decisions, not diagrams, are the architecture

Architecture is the set of decisions that are expensive to reverse. Diagrams describe a snapshot; the durable value is *why* each significant choice was made and what was traded away. Capture that reasoning, or it evaporates and the next team re-litigates (or silently violates) it.

## Quality attributes that matter

Name the attributes you are optimizing for; they conflict, so you cannot maximize all of them:

- **Evolvability / maintainability** — the headline goal for enterprise systems. Achieved by a dependency-inverted domain core and contracts at boundaries. Most other choices are in service of this.
- **Scalability** — handle more load (vertically/horizontally). CQRS read replicas, stateless services, async messaging help; stateful sessions and DB-counter ids hurt.
- **Availability / reliability** — stays up under failure. Redundancy, resilience (Polly), graceful degradation, no single hard point of failure on the hot path.
- **Consistency** — strong (immediate) vs eventual. Distributed systems trade consistency for availability/scalability (CAP). Be explicit about where eventual consistency is acceptable.
- **Performance / latency** — caching, indexing, best-so-far snapshots; beware premature optimization ("the root of all evil") — design for change first, optimize measured hot paths later. (Security and multitenancy are the exceptions: hard to retrofit, decide early.)
- **Security** — confidentiality, integrity, availability, traceability. Zero-trust between modules; auth at every boundary; design in, don't bolt on.
- **Testability** — a decoupled domain and contract-first APIs make unit, integration, and contract tests cheap.
- **Operability / observability** — can you run, monitor, and debug it (logs, metrics, traces, correlation)?
- **Cost** — engineering effort and runtime spend; a "temporary" intermediate architecture can cost as much as the target — sometimes go straight to the target.

## Naming trade-offs explicitly

Every choice gives something up. State both sides:

- Microservices → independent scaling/deploy **but** distributed-systems complexity, eventual consistency, ops overhead.
- Eventual consistency → availability/scalability **but** stale reads, idempotency burden.
- Document store → shape-fit, simple writes **but** weaker cross-entity transactions, less mature reporting tooling.
- Event sourcing → full audit/replay **but** event versioning, snapshotting, learning curve.
- Externalized rules/authorization (BRMS/OPA) → centralized, auditable, business-editable **but** performance cost and an extra moving part.

A trade-off made *consciously and documented* is fine; an accidental one is debt.

## Architecture decision records (ADRs)

An **ADR** is a short, immutable, numbered markdown file capturing one decision: the context, the choice, the alternatives, and the consequences. Keep them in the repo (e.g. `docs/adr/0001-use-modular-monolith.md`) so the history of *why* lives with the code. Never edit a decided ADR's substance — supersede it with a new one that links back.

Lifecycle status: `Proposed` → `Accepted` → (later) `Superseded by ADR-NNNN` or `Deprecated`.

## An ADR template

```markdown
# ADR-0007: Store catalog data in a document database

## Status
Accepted (2026-06-13). Supersedes none.

## Context
The Catalog service owns Book and Author aggregates. These are tree-shaped
(nested editing/sales sub-objects, arrays of tags and links) and need full
change history for audit. The team knows SQL; no one has run a document store
in production. Expected volume is modest (low millions of documents).

## Decision
Use a document database for the Catalog write store, persisting changes as
JSON Patch deltas plus a cached best-so-far state. Expose a SQL/OData read
endpoint for reporting consumers.

## Alternatives considered
- Relational + EF Core: familiar, but multi-table mapping of nested documents,
  transactions, and locks add accidental complexity for tree-shaped data.
- Event sourcing (Marten): full replay, but more machinery than the audit
  requirement justifies right now.

## Consequences
+ Writes are single-document, lock-free; history and time-travel come for free.
+ Read model matches API JSON; no ORM impedance.
- Team needs ~3 days of training; ops must learn backup/restore for the store.
- Cross-aggregate transactional invariants must be avoided by design (small aggregates).
- Reporting relies on the secondary SQL/OData endpoint, kept eventually consistent.
```

## Worked ADR example (decision summary)

For "should Catalog and Sales be separate services?": context = one team, shared release cadence, frequent cross-reads between book and sales data. Decision = **keep them as two modules in one modular monolith**, not two services, because they share a deployment rhythm and the cross-reads would be chatty network calls. Consequence = lower latency and no distributed transactions now; an explicit module boundary (separate assemblies, `internal` visibility, architecture tests) so either can be extracted later if scaling or team boundaries change. Trade-off accepted: less independent scalability today in exchange for far less complexity.

## Fitness functions: enforcing decisions over time

Decisions decay unless enforced. **Fitness functions** are automated checks that fail the build when an architectural property is violated:

- **Dependency-direction tests** (e.g. NetArchTest, ArchUnitNET): assert Domain references no Infrastructure; module A doesn't reference module B's internals.
- **Contract tests** against OpenAPI schemas: catch backward-incompatible API changes.
- **Performance budgets** / load tests in CI for latency-sensitive paths.
- **Dependency/license scans**: catch disallowed packages.

```csharp
// NetArchTest example: the domain must not depend on EF Core
var result = Types.InAssembly(typeof(Book).Assembly)
    .That().ResideInNamespace("Acme.Catalog.Domain")
    .ShouldNot().HaveDependencyOn("Microsoft.EntityFrameworkCore")
    .GetResult();
Assert.True(result.IsSuccessful, string.Join(", ", result.FailingTypeNames ?? []));
```

## Evaluating an existing architecture

A quick health check when reviewing or inheriting a system:

1. **Dependency direction** — does anything in the domain depend on frameworks/DB? Fix the arrows.
2. **Coupling at the seams** — shared databases, DB-level integration, circular service dependencies, chatty sync calls on hot paths? These are the evolvability killers.
3. **Boundary alignment** — do module/service boundaries match bounded contexts, or do features routinely span many?
4. **Cohesion** — does each module/service have a single business reason to change (SRP at scale)?
5. **Contracts** — contract-first or generated-last? Versioned? Backward-compatible?
6. **Cross-cutting consistency** — logging/validation/resilience/authorization done uniformly or ad hoc?
7. **The "why" trail** — are significant decisions recorded (ADRs), or is the rationale lost?

## Technical debt

Debt is the gap between the current design and one aligned with the business. Some is deliberate and sound (ship simple now, evolve later — *if* the domain model is correct so evolution stays cheap). Debt becomes dangerous when it is accidental (DB coupling, frozen-in semantic errors) or when "temporary" intermediate states ossify. Track it, attach it to ADRs where relevant, and pay it down before it compounds into a system that cannot evolve. The whole point of the domain-centric, contract-at-the-boundary approach is to keep the cost of change low so debt stays serviceable.
