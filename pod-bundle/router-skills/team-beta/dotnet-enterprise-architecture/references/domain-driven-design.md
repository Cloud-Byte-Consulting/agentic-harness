# Domain-driven design in .NET

Contents:
- The mindset: functional before technical
- Ubiquitous language
- Bounded contexts and context mapping
- Subdomains: core, supporting, generic
- Tactical building blocks (entity, value object, aggregate, repository)
- Domain events
- Entity life cycle and time
- Semantics: the most expensive bugs
- A worked C# aggregate

## The mindset: functional before technical

DDD's last D is **Design**, not Development. The single most valuable habit: understand the problem functionally for as long as you can, and only then translate to code. Once a model is solidified into code — and worse, into APIs and databases consumed by many clients — design errors become extremely hard to correct. Think of data, not databases; of models and business rules, not attributes and methods.

The essentials of the method (per Eric Evans):
- Creative collaboration of domain experts and software experts.
- Exploration and experimentation.
- Emerging models that shape and reshape the ubiquitous language.
- Explicit context boundaries.
- Focus on the core domain.

## Ubiquitous language

Different roles use different words for the same thing — editors say "work", salespeople say "product", both mean a book. DDD does not abolish local jargon; it agrees on **one shared term** (e.g. `Book`) used wherever a misunderstanding could cause harm, while letting local jargon survive inside its own context. The agreed terms become the names of your classes, methods, endpoints, and events. If the code says `Book`, the conversation says "book", and the API path is `/books`.

## Bounded contexts and context mapping

A **bounded context** is the perimeter inside which the vocabulary is consistent and a single model holds. The same word can mean different things in different contexts; the boundary is where you must be explicit about translation.

- Bounded contexts often align to business subdomains, but not always — entity *life cycle* also drives where boundaries fall.
- Each bounded context ideally maps to one module (in a modular monolith) or one service (in microservices), with one ubiquitous language and one published contract.
- A **context map** records the relationships between contexts: partnership, shared kernel, customer/supplier, conformist, anti-corruption layer (ACL). Use an **anti-corruption layer** when integrating with a legacy or external model you don't want leaking into your clean domain — it translates the foreign model into your terms at the boundary.

## Subdomains: core, supporting, generic

- **Core domain** — where the business differentiates itself (for a publisher: authoring and selling books). Invest your best modeling here; this is rarely off-the-shelf software.
- **Supporting subdomains** — necessary but not differentiating (HR, accounting). Model lightly or buy.
- **Generic subdomains** — solved problems (identity, document management, email, payments). Use standards and existing products; don't build.

## Tactical building blocks

- **Entity** — has identity and a life cycle; equality is by id, not by attributes (a Book, an Author). Identity persists even as attributes change.
- **Value object** — defined entirely by its values, immutable, no identity (a `Money`, an `Isbn`, an `Address`). Replace, don't mutate. C# `record` types are an excellent fit.
- **Aggregate** — a cluster of entities and value objects with one **aggregate root**; the root enforces invariants and is the only entry point. External references point to the root by id only. Transactions and consistency boundaries are *per aggregate* — keep aggregates small.
- **Repository** — a collection-like abstraction for loading/saving an aggregate root (one repository per aggregate, not per table). See `persistence-patterns.md`.
- **Domain service** — behavior that doesn't belong to a single entity/value object (e.g. a pricing calculation spanning several aggregates). Stateless, expressed in domain terms.

Rule for minor vs major entities: an entity is **major** (its own aggregate, maybe its own context/service) when it has an independent life cycle (Book, Author). It is **minor** when it only exists within a parent and dies with it (an author's address, a book's tag). Minor entities live inside the aggregate.

A value object example:

```csharp
public readonly record struct Isbn
{
    public string Value { get; }
    public Isbn(string value)
    {
        if (!IsValid(value))
            throw new ArgumentException($"Invalid ISBN: {value}", nameof(value));
        Value = value;
    }
    private static bool IsValid(string v) => v.Replace("-", "").Length is 10 or 13;
    public override string ToString() => Value;
}

public sealed record Money(decimal Amount, string CurrencyCode) // ISO 4217, e.g. "EUR"
{
    public static Money Eur(decimal amount) => new(amount, "EUR");
}
```

Note how `Money` carries the currency, and weight would be `WeightInGrams` (an explicit, standardized unit) — never bare `decimal`s whose unit is implicit. Things you think are primitives (country, currency, status) are usually small objects.

## Domain events

A **domain event** records something meaningful that happened in the domain ("BookReadyToPrint", "AuthorEnrolled"), named in the past tense. Aggregates raise them; the application layer dispatches them after a successful commit. They decouple "what happened" from "what reacts": printing, notifications, and cache refreshes subscribe without the aggregate knowing about them.

```csharp
public abstract record DomainEvent(DateTimeOffset OccurredAt);
public sealed record BookReadyToPrint(BookId BookId, DateTimeOffset OccurredAt) : DomainEvent(OccurredAt);

public sealed class Book
{
    private readonly List<DomainEvent> _events = new();
    public IReadOnlyList<DomainEvent> DomainEvents => _events;
    public void ClearEvents() => _events.Clear();

    public void MarkReadyToPrint(IClock clock)
    {
        if (Status != BookStatus.Reviewed)
            throw new InvalidOperationException("Only a reviewed book can be marked ready to print.");
        Status = BookStatus.ReadyToPrint;
        _events.Add(new BookReadyToPrint(Id, clock.Now));
    }
}
```

In-process, dispatch via MediatR notifications (see `cqrs-and-mediator.md`); across services, publish to a broker after persisting (see `microservices-and-messaging.md`, outbox pattern).

## Entity life cycle and time

A major entity is not just a bag of attributes — it is a living object with states over time (a book: Idea → AuthorChosen → Writing → Reviewed → ReadyToPrint → Available → Retired → Archived). Model the life cycle explicitly:

- **Status is often a business rule, not a stored field.** "ReadyToPrint" may *derive* from conditions (two editors approved, contract signed, printer accepted files), not a value someone toggles. Model the rule; cache the computed status only for performance, and be explicit when a state, once reached (e.g. Archived), can never be reversed.
- **History is part of the model.** Important entities frequently need every past state, not just the latest, plus the *functional reason* a change happened (not just "addresses[1] changed" but "the author moved"). This pushes you toward storing operations/deltas — see `persistence-patterns.md`.
- Time also drives boundaries: tags have no life cycle outside a book → not a major entity; authors do (created, change details, eventually GDPR-erased) → a major entity.

## Semantics: the most expensive bugs

The costliest design errors are semantic and get frozen into schemas and APIs:

- Modeling **customer** and **supplier** as separate entity tables when a company can be both — leading to address/data duplication and "fix-up" triggers that loop and crash the database. The correct model: one `Actor` (individual or organization) plus *business rules* — an actor is a customer if an order points at it within N months; a supplier if an incoming-order/equipment record does. Change the rule (18 months → 12) without migrating any data. The schema doesn't move when the rule moves.
- Confusing a derived state with stored data (prospect vs customer being a stage in a sales pipeline, not an entity type).

When in doubt, ask: "is this a thing, or a rule about things?" Run modeling workshops with product owners; be wary of letting a technical mindset prematurely constrain the model.

## A worked C# aggregate

```csharp
public enum BookStatus { Idea, AuthorChosen, Writing, Reviewed, ReadyToPrint, Available, Retired, Archived }

public sealed class Book // aggregate root
{
    public BookId Id { get; }
    public Isbn? Isbn { get; private set; }          // null until officially registered — nullability models reality
    public string? Title { get; private set; }
    public BookStatus Status { get; private set; }
    private readonly List<DomainEvent> _events = new();
    public IReadOnlyList<DomainEvent> DomainEvents => _events;

    public Book(BookId id) { Id = id; Status = BookStatus.Idea; }

    public void AssignIsbn(Isbn isbn)
    {
        if (Status is BookStatus.Archived)
            throw new InvalidOperationException("Cannot modify an archived book.");
        Isbn = isbn;
    }
    // invariants and transitions live here; external code never sets Status directly
}
```

Notes: nullable properties are *correct* for a business model — most attributes are genuinely unknown at some life-cycle stage (an ISBN isn't assigned until registration). This differs from algorithmic code where you fight nulls. Keep invariant enforcement inside the root; expose intent-revealing methods, not setters.
