# Architecture styles for .NET enterprise systems

Contents:
- The one principle behind all domain-centric styles
- Layered / N-tier
- Clean, onion, hexagonal (ports & adapters)
- Modular monolith
- Microservices, and the granularity question
- The monolith-vs-microservices debate
- Project/folder layout for a clean .NET solution
- C4 for describing what you built
- Choosing a style: a checklist

## The one principle behind all domain-centric styles

Clean, onion, and hexagonal architectures are visually different drawings of the **same two rules**:

1. The business model (entities + the rules over them) is at the core and **depends on nothing** — no framework, no database, no UI, no transport.
2. Everything that touches the outside world depends on the core through **indirection** (interfaces/ports), so it can be swapped without disturbing the business logic.

Arguing over circles vs hexagons is wasted effort. What produces quality is keeping the dependency arrows pointing toward the domain and controlling coupling at every boundary. Pick whichever drawing your team thinks in, or mix them.

## Layered / N-tier

```
Presentation  →  Application  →  Domain  →  Infrastructure
```

Each tier calls the one beneath it. Familiar, easy to onboard. Two failure modes:

- **DB-level integration.** Apps (or layers) reaching straight into the database, bypassing the application layer. This is the dominant reason legacy systems can't evolve — the schema becomes a shared, un-versioned contract. Enforce that data access goes only through the domain/infrastructure boundary the layer above owns.
- **Anemic domain.** Logic drains into "service" classes and the domain becomes bags of getters/setters. Keep behavior with the data it guards.

A *strict* layered approach still respects rule (1) above if Domain has no reference to Infrastructure — that is essentially onion architecture.

## Clean, onion, hexagonal (ports & adapters)

Concentric (onion/clean) or center-and-sides (hexagonal) — same dependency rule. A typical .NET realization:

- **Domain** (innermost): entities, value objects, aggregates, domain events, domain-service interfaces. No NuGet dependencies beyond the BCL.
- **Application**: use cases / command & query handlers, orchestrating the domain; defines **ports** — interfaces like `IBookRepository`, `IClock`, `IEmailSender`.
- **Infrastructure** (adapters): EF Core / Mongo repositories, HTTP clients, message-broker publishers, file/email/blob adapters — *implements* the ports.
- **Presentation** (adapters): ASP.NET Core API, Blazor, gRPC, a console host.

Composition root (the host's `Program.cs`) wires adapters to ports via DI. The arrow from Infrastructure/Presentation *into* Application/Domain is the whole point.

```csharp
// Application layer — a PORT (the domain/app owns this interface)
public interface IBookRepository
{
    Task<Book?> GetAsync(BookId id, CancellationToken ct);
    Task SaveAsync(Book book, CancellationToken ct);
}

// Infrastructure layer — an ADAPTER implementing the port
public sealed class EfBookRepository(AppDbContext db) : IBookRepository
{
    public Task<Book?> GetAsync(BookId id, CancellationToken ct) =>
        db.Books.FirstOrDefaultAsync(b => b.Id == id, ct);

    public Task SaveAsync(Book book, CancellationToken ct)
    {
        db.Books.Update(book);
        return db.SaveChangesAsync(ct);
    }
}
```

When to *not* go full hexagonal: a small service (~1000 lines) over a standardized driver (e.g. a Mongo client whose API is itself effectively a standard) gains little from extracting every persistence call behind a port. Use an interface where mocking or swapping is plausible; otherwise keep it direct. Match ceremony to size.

## Modular monolith

One deployable process; internally split into **modules** with hard boundaries (ideally one module per bounded context). Modules talk to each other only through published interfaces or in-process messages — never by reaching into each other's tables or internals.

Why it's usually the right start:
- You get cohesion, clear ownership, and an evolvable structure.
- You avoid network calls, distributed transactions, eventual-consistency bugs, and per-service ops overhead.
- A well-bounded module is straightforward to later extract into a service if a real force demands it.

Enforce boundaries with separate projects/assemblies, `internal` visibility, and (optionally) architecture tests (e.g. NetArchTest) that fail the build if module A references module B's internals.

## Microservices, and the granularity question

Independently deployable services, each owning its data and exposing an API/message contract. Real prerequisites before choosing them:

- High and *independent* scaling needs across capabilities.
- Frequent, independent release cadences per capability.
- Clear team boundaries (Conway's law working *for* you).
- Tolerance for eventual consistency and distributed-systems complexity.

The "micro" prefix misleads. The actual decision is **granularity**: how coarse or fine the service boundaries are. Too fine and you spend more on plumbing (endpoints, contracts, integration tests, ops) than you ever saved. DDD bounded contexts are the tool for finding the right grain — one context, one service, one data store, one contract.

Cross-service data: each service owns its store. For relationships across services, store a *link/identifier* plus a small cached copy of frequently needed fields (e.g. a book stores the author id + name), and decide functionally whether that cached value should track changes or stay frozen. For cross-cutting reads, build a dedicated read/reporting service that *consumes from* the others — and keep that flow one-directional (never let an atomic service consume back from the aggregator; that re-introduces circular coupling).

## The monolith-vs-microservices debate

Both extremes are wrong as dogma. "We're not Amazon so never microservices" and "Amazon does it so we must" are equally unthinking. The engineer's answer is: analyze the business, find the right granularity, and place yourself on the spectrum accordingly. A "return to the monolith" after over-splitting usually means the grain was wrong, not that boundaries are bad. Re-grain; don't collapse to one blob.

## Project/folder layout for a clean .NET solution

```
src/
  Acme.Catalog.Domain/          # entities, value objects, aggregates, domain events, port interfaces
  Acme.Catalog.Application/     # use cases / command+query handlers, DTOs, validators
  Acme.Catalog.Infrastructure/  # EF Core/Mongo, message bus, external HTTP clients (adapters)
  Acme.Catalog.Api/             # ASP.NET Core host = composition root
tests/
  Acme.Catalog.Domain.Tests/
  Acme.Catalog.Application.Tests/
  Acme.Catalog.Architecture.Tests/   # dependency-direction guard rails
```

Project references point inward only: Api → Infrastructure → Application → Domain. Domain references nothing in the solution.

## C4 for describing what you built

C4 describes a system at four zoom levels — use only the levels that add value, not as a box-ticking exercise:

1. **Context** — the system and the people/external systems it talks to.
2. **Container** — the deployable units (a web app, an API, a database, a broker). In a containerized deploy, one C4 container ≈ one Docker container ≈ one process.
3. **Component** — the major building blocks inside a container (a controller, a repository, a cache). In .NET, components often map to assemblies/namespaces.
4. **Code** — class diagrams; usually skip unless a specific design needs it.

C4 is top-down and complements the domain-centric styles, which best describe the *component* level of a container.

## Choosing a style: a checklist

- Single small app, one team, modest scale → **clean/onion inside a single deployable**.
- Multiple capabilities, one team, want evolvability without ops cost → **modular monolith**.
- Independent scaling/release per capability, multiple teams, eventual consistency acceptable → **microservices** sized by bounded context.
- Legacy with DB-level coupling you must tame → introduce a referential/owning service per data domain, route all access through its API, and strangle the old integrations over time.

In every case: domain at the core, contracts at the boundaries, decisions recorded as ADRs.
