# CQRS and the mediator pattern in .NET

Contents:
- What CQRS is, and when it pays off
- The mediator pattern and MediatR
- Commands, queries, handlers
- Pipeline behaviors (validation, logging, transactions)
- Separate read models
- Event sourcing
- Decision guide

## What CQRS is, and when it pays off

**Command Query Responsibility Segregation** separates the model that *changes* state (commands) from the model that *reads* state (queries). Reads and writes in line-of-business apps differ profoundly: typically ~80% reads / ~20% writes; reads need no locks and want denormalized, query-shaped data; writes need consistency and invariants. Separating them lets each side be optimized and scaled independently (you can add read replicas without limit since the write side remains the single source of truth).

Use CQRS when:
- Read and write workloads diverge in shape, volume, or consistency.
- You want different storage for reads (a denormalized view, a search index) than for writes.
- Complex domains where command handlers should be small and focused.

Do **not** reach for full CQRS (separate databases, async projections) by default — it adds eventual-consistency complexity. The lightweight, hugely common version is: command and query *handlers* in one service against one database, mediated by MediatR. That alone gives single-responsibility handlers and a clean pipeline.

## The mediator pattern and MediatR

A **mediator** decouples the caller (a controller, a message consumer) from the handler. The caller sends a request object; the mediator routes it to exactly one handler. This keeps controllers thin and makes each use case an isolated, testable unit.

MediatR is the de-facto .NET library. (Note: MediatR moved to a commercial license for newer major versions; confirm the license terms for your version, or use an alternative such as a hand-rolled dispatcher, Wolverine, or `Mediator` source generator if licensing is a concern. The pattern below is identical regardless of library.)

```csharp
// Program.cs
builder.Services.AddMediatR(cfg =>
    cfg.RegisterServicesFromAssembly(typeof(CreateBook).Assembly));
```

## Commands, queries, handlers

A **command** expresses intent to change state and returns little (an id, or nothing). A **query** returns data and must not mutate state.

```csharp
// Command
public sealed record CreateBook(string Title, string? Isbn) : IRequest<BookId>;

public sealed class CreateBookHandler(IBookRepository repo, IClock clock)
    : IRequestHandler<CreateBook, BookId>
{
    public async Task<BookId> Handle(CreateBook cmd, CancellationToken ct)
    {
        var book = new Book(BookId.New());
        book.Rename(cmd.Title);
        if (cmd.Isbn is not null) book.AssignIsbn(new Isbn(cmd.Isbn));
        await repo.SaveAsync(book, ct);
        return book.Id;
    }
}

// Query — returns a read DTO, never a domain entity
public sealed record GetBookById(Guid Id) : IRequest<BookDto?>;

public sealed class GetBookByIdHandler(IReadDb db) : IRequestHandler<GetBookById, BookDto?>
{
    public Task<BookDto?> Handle(GetBookById q, CancellationToken ct) =>
        db.Books.Where(b => b.Id == q.Id)
                .Select(b => new BookDto(b.Id, b.Title, b.Isbn, b.Status))
                .FirstOrDefaultAsync(ct);
}
```

Controller/endpoint stays minimal:

```csharp
app.MapPost("/books", async (CreateBook cmd, ISender mediator, CancellationToken ct) =>
{
    var id = await mediator.Send(cmd, ct);
    return Results.Created($"/books/{id.Value}", new { id });
});
```

Queries return DTOs (read models), not domain aggregates — never leak the write model out of a query, and don't reuse a query result as an input to a command.

## Pipeline behaviors (validation, logging, transactions)

Cross-cutting concerns wrap every request as **pipeline behaviors** — the mediator equivalent of middleware. This is where validation, logging, transactions, caching, and retries live, so handlers stay pure business logic.

```csharp
public sealed class ValidationBehavior<TReq, TResp>(IEnumerable<IValidator<TReq>> validators)
    : IPipelineBehavior<TReq, TResp> where TReq : notnull
{
    public async Task<TResp> Handle(TReq request, RequestHandlerDelegate<TResp> next, CancellationToken ct)
    {
        var context = new ValidationContext<TReq>(request);
        var failures = validators
            .Select(v => v.Validate(context))
            .SelectMany(r => r.Errors)
            .Where(f => f is not null)
            .ToList();
        if (failures.Count != 0) throw new ValidationException(failures);
        return await next();
    }
}

// Registration (order matters — behaviors run in registration order)
builder.Services.AddTransient(typeof(IPipelineBehavior<,>), typeof(LoggingBehavior<,>));
builder.Services.AddTransient(typeof(IPipelineBehavior<,>), typeof(ValidationBehavior<,>));
builder.Services.AddTransient(typeof(IPipelineBehavior<,>), typeof(TransactionBehavior<,>));
```

A `TransactionBehavior` can open a unit of work for commands only (check a marker interface like `ICommand`) and commit on success / roll back on exception — see `persistence-patterns.md`.

## Separate read models

The next step up: write and read use different stores. Commands update the normalized write store and raise domain events; **projections** subscribe to those events and update denormalized read models (a SQL view table, a document, an Elasticsearch index) shaped exactly for the queries. Reads become trivial and fast; the cost is **eventual consistency** between write and read — the read model lags by milliseconds to seconds.

Handle the consistency gap honestly: tell the UI a value may be momentarily stale, or read-your-own-writes from the write store for the acting user immediately after a command.

## Event sourcing

Instead of storing current state, store the **sequence of events** that produced it; rebuild state by replaying events. It pairs naturally with CQRS: the event stream is the write side; projections build read models.

Benefits: complete audit/history, temporal queries (state at any past date), and lock-free appends. Costs: schema/versioning of events, snapshotting for performance, and a steeper learning curve. .NET options include Marten (events on PostgreSQL) and EventStoreDB.

Adopt event sourcing only when the audit trail / temporal replay is a genuine requirement — not for novelty. A lighter middle ground that captures most of the benefit: store **operations (deltas) plus a cached best-so-far state** (see `persistence-patterns.md`), which gives history and lock-free writes without full event-sourcing machinery.

## Decision guide

| Need | Approach |
|------|----------|
| Thin controllers, isolated use cases | Mediator + handlers (one DB) |
| Reads diverge from writes in shape/volume | CQRS with separate read DTOs/queries |
| Reads need a different store/index | CQRS with projections (accept eventual consistency) |
| Full audit trail + temporal replay | Event sourcing |
| Audit + lock-free writes, modest scope | Operation/delta storage + best-so-far cache |
| Cross-cutting concerns around use cases | Pipeline behaviors |

Start at the top row; move down only when a concrete force justifies the added complexity.
