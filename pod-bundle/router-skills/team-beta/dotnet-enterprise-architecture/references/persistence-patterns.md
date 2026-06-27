# Persistence patterns for .NET enterprise systems

Contents:
- Repository and unit of work
- EF Core as a unit of work
- SQL vs NoSQL: choosing by data shape
- Storing operations, not states (delta persistence)
- Master data management (MDM) and the "referential"
- Identifiers
- Deletion, archiving, and regulation
- In-memory persistence (object prevalence)

## Repository and unit of work

- **Repository** — a collection-like abstraction over loading and saving an **aggregate root**. One repository per aggregate, expressed in domain terms (`IBookRepository.GetAsync(BookId)`), not one per table and not a generic CRUD bag. The domain/application layer defines the interface (a port); infrastructure implements it.
- **Unit of work** — tracks changes across one logical operation and commits them atomically, so a single business action that touches several entities either fully succeeds or fully fails.

Keep repositories focused on persistence; don't smuggle business rules into them (those belong in the aggregate or a domain service).

## EF Core as a unit of work

`DbContext` *is* a unit of work, and `DbSet<T>` *is* a repository. For many services you can use them directly and skip extra abstraction — wrapping EF Core in a hand-written generic repository often adds indirection without benefit. Add your own repository interface only when you want to (a) keep the domain free of EF, (b) constrain the API to aggregate-level operations, or (c) enable easy unit-test mocking.

```csharp
public sealed class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
{
    public DbSet<Book> Books => Set<Book>();
    protected override void OnModelCreating(ModelBuilder b)
    {
        b.Entity<Book>(e =>
        {
            e.HasKey(x => x.Id);
            e.OwnsOne(x => x.Isbn);          // value object as owned type
            e.Property(x => x.Status).HasConversion<string>(); // store enum as text, not int
        });
    }
}

// A unit-of-work-style command boundary
public async Task<BookId> Handle(CreateBook cmd, CancellationToken ct)
{
    var book = new Book(BookId.New());
    book.Rename(cmd.Title);
    db.Books.Add(book);
    await db.SaveChangesAsync(ct);   // single atomic commit = unit of work
    return book.Id;
}
```

Tips: map value objects with `OwnsOne`/`OwnsMany`; store enums and statuses as **strings** (adding a value shouldn't force a migration, and text survives reordering); use `AsNoTracking()` for queries; keep read queries projecting to DTOs.

## SQL vs NoSQL: choosing by data shape

Decide by the natural shape of your data, not by team habit:

- **Document store (NoSQL, e.g. MongoDB, Cosmos DB, Marten on PostgreSQL)** — most business entities are tree-shaped documents (a Book with nested editing/sales sub-objects and arrays). A document store matches that shape, removing the ORM/transaction/join tax: an insert is one operation, a read is one document. Schemaless flexibility lets related-but-varied items (individuals and organizations as "actors") coexist with type-specific fields — *use that freedom responsibly*; it is not licence for a junk drawer.
- **Relational (SQL)** — genuinely tabular data, strong multi-entity transactional invariants, and reporting/BI consumers who expect SQL. SQL's tabular constraint, ACID transactions, and lock management are strengths *for the right data* and overhead for the wrong data.

The historical reason SQL dominates is largely inertia (fixed-length rows for spinning-disk block addressing), no longer technically compelling for document-shaped data. A good service can also expose a SQL endpoint for reporting tools even when it stores documents.

Anti-pattern: choosing relational tables for tree-shaped business entities *just* to reuse a familiar query language — it imports needless complexity (multi-table joins, transactions, optimistic/pessimistic locks).

## Storing operations, not states (delta persistence)

When history, traceability, or audit matters, store the **changes** that happened, not just the current snapshot. This is the lightweight cousin of event sourcing and a core master-data technique. Three collections/tables:

1. **changes** — every modification as a JSON Patch (RFC 6902) delta, with entity id and value date. Nothing is ever overwritten; you only append.
2. **states** — the full state after each change (so you don't replay all deltas every read). Optionally store only every Nth state and replay a few patches to reach an exact historical point.
3. **best-so-far** — the latest state per entity, a cache for the common "give me the current value" read. Looks like a normal table, but no data is lost.

Benefits: complete history and audit, time-travel reads (`?valueDate=`), and **lock-free writes** (appends never conflict, so no optimistic/pessimistic locks, no compensation). Create becomes a patch from empty; delete becomes a status change (archive). See `api-and-integration.md` for the API surface (`PATCH` with JSON Patch, value dates, history endpoints).

```csharp
// PATCH handler core (document store): append the delta, recompute and cache best-so-far
public async Task<IActionResult> Patch(string id, [FromBody] JsonPatchDocument<Book> patch,
                                        [FromQuery] DateTimeOffset? valueDate = null)
{
    if (patch is null) return BadRequest();
    var when = valueDate ?? DateTimeOffset.UtcNow;

    var current = await _bestSoFar.Find(b => b.EntityId == id).FirstOrDefaultAsync()
                  ?? new Book { EntityId = id };   // create == patch from empty
    await _changes.InsertOneAsync(new ChangeUnit(id, when, patch));   // 1. append delta
    patch.ApplyTo(current);
    await _states.InsertOneAsync(new ObjectState<Book>(id, when, current)); // 2. record state
    await _bestSoFar.ReplaceOneAsync(b => b.EntityId == id, current, new ReplaceOptions { IsUpsert = true }); // 3. cache
    return new ObjectResult(current);
}
```

Note: ASP.NET Core's `JsonPatchDocument` support historically required the `Newtonsoft.Json`-based path (`AddNewtonsoftJson()` + `Microsoft.AspNetCore.Mvc.NewtonsoftJson`); recent .NET versions add `System.Text.Json` JSON Patch support — check your target framework and prefer the built-in serializer where available.

## Master data management (MDM) and the "referential"

A **data referential** is a service that owns the single source of truth for a business entity (Books, Authors) — more than a database: it owns persistence + history + metadata + validation + authorization + governance for that entity. Architectural patterns:

- **Centralized** — one owning service; everyone reads/writes through its API. Simplest; the single point of failure is a solved technical problem (replication, scaling) and is more about org ownership.
- **Clone** — applications keep a local cache/copy synced from the centralized source (via events/webhooks ideally, batch as fallback). The central store stays the source of truth.
- **Consolidated/distributed** — the referential exposes a full entity but some parts ("petals") are owned by other apps and fetched/cached on read. Watch performance (pagination across sources is genuinely hard).

Governance is non-technical but essential: name a **data owner** (defines rules, who can access) and a **data steward** (daily quality/availability). Shared or absent ownership leads to uncontrolled model drift — and if IT owns the data definition by default, expect modeling mistakes (one address only, no product/article distinction).

The cardinal rule: **nothing but the owning service touches its database.** Feeding a referential's DB via ETL bypasses its validation/business rules and is a recipe for corruption.

## Identifiers

- **System-wide identifier** — globally understood, stable, often a URN (`urn:com:acme:library:books:978-...`). Avoid leaking transport (a URL implies HTTP access).
- **Local/technical identifier** — store-specific (a Mongo ObjectId, a SQL surrogate key). Never expose it to other modules; changing the store would break their links.
- **Business identifiers** — ISBN, VAT number, etc. Use these where a recognized standard exists.
- **External identifiers** — keep a dictionary of other systems' ids keyed by a *generic* module key (`urn:org:acme:accounting → BK4648`), not the vendor product name, so a vendor swap doesn't pollute the mapping.

Avoid SQL identity counters for distributed systems (they don't scale and force centralization). Prefer GUIDs or business keys; if users love gap-free sequential numbers, assign them **asynchronously** after creation (modulo-per-server or pre-allocated ranges) to avoid contention — and consider convincing them to drop the requirement.

## Deletion, archiving, and regulation

With delta storage you never physically remove data by default — delete sets a status (`Archived`) and reads of archived data return 404 unless the caller has an archive role. *True* erasure is reserved for regulation: GDPR's right-to-be-forgotten requires actually removing personal data everywhere (including backups and any denormalized copies in other services — e.g. replacing an author's name embedded in book links with `N/A (GDPR)`).

## In-memory persistence (object prevalence)

For small-volume referentials needing very high performance, very complex queries (easier in LINQ than SQL), or a fast-evolving model, consider **object prevalence**: keep the model in RAM, persist a disk log of commands for durability (rebuild on restart). Reads/queries run directly against your object graph with LINQ — nothing is faster, with no ORM/serialization. Niche, but powerful for the right case. Match the persistence mechanism to the business need, as always.
