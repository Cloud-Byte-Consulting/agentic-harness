# Cross-cutting concerns

Contents:
- The principle: decoupled, standardized seams
- Structured logging (ILogger + Serilog)
- Correlation, tracing, observability
- Validation (FluentValidation)
- Resilience with Polly (v8 pipelines)
- Resilience for HTTP clients
- Externalizing business rules (when it pays off)
- Externalizing authorization (RBAC, ABAC, OPA)

## The principle: decoupled, standardized seams

Cross-cutting concerns (logging, validation, resilience, authorization, monitoring) touch every part of the system, so they are the easiest place to accidentally create coupling. The rule: each concern is reached through a **standardized interface** (a framework-local or international standard) so its implementation can change without touching business code. Logging behind `ILogger`; validation behind `IValidator`; resilience as a policy/pipeline wrapped around calls; authorization behind a policy abstraction. Then a Serilog→Elastic swap, or a retry-policy tweak, is local.

In a mediator-based service, most of these live as **pipeline behaviors** (see `cqrs-and-mediator.md`); in ASP.NET Core, as **middleware**; for outbound calls, as **delegating handlers/policies**.

## Structured logging (ILogger + Serilog)

Log against the abstraction `Microsoft.Extensions.Logging.ILogger<T>` so the sink is swappable. Use **Serilog** as the implementation for structured (semantic) logging — log events with named properties, not interpolated strings, so logs are queryable.

```csharp
// Program.cs
builder.Host.UseSerilog((ctx, cfg) => cfg
    .ReadFrom.Configuration(ctx.Configuration)
    .Enrich.FromLogContext()
    .WriteTo.Console(new Serilog.Formatting.Compact.CompactJsonFormatter()));

// Usage — message template with structured properties (NOT $"...{id}")
_logger.LogInformation("Book {BookId} moved to status {Status}", book.Id, book.Status);
```

Separate the *responsibilities* behind "logging": client-side trace, server-side trace, end-to-end interaction tracing (needs a shared correlation id), resource monitoring, and Business Activity Monitoring (BAM — usage statistics). Keep them loosely coupled: you should be able to run BAM continuously without turning on verbose traces, and you shouldn't be locked into a monitoring vendor's proprietary alerting. Centralizing logs/metrics aids cross-service analysis but is itself a form of technical coupling — get the semantics and metadata right so the central store stays useful.

## Correlation, tracing, observability

A single user interaction crosses the client, the API, and several services. Thread a **correlation id** (a unique interaction code) through every hop so a centralized system can reconstruct the whole flow. Use `System.Diagnostics.Activity` / **OpenTelemetry** for distributed tracing — it propagates trace context across HTTP and messaging automatically and exports to Jaeger/Tempo/etc.

```csharp
builder.Services.AddOpenTelemetry()
    .WithTracing(t => t.AddAspNetCoreInstrumentation().AddHttpClientInstrumentation())
    .WithMetrics(m => m.AddAspNetCoreInstrumentation().AddRuntimeInstrumentation());
```

The three pillars: structured **logs**, **metrics** (counters/histograms), and **traces** (spans across services). Standardize on OpenTelemetry so backends stay swappable.

## Validation (FluentValidation)

Validate at the boundary, as early as possible, in business terms. **FluentValidation** keeps rules out of the model and the controller:

```csharp
public sealed class CreateBookValidator : AbstractValidator<CreateBook>
{
    public CreateBookValidator()
    {
        RuleFor(x => x.Title).NotEmpty().MaximumLength(300);
        RuleFor(x => x.Isbn).Must(BeValidIsbn).When(x => x.Isbn is not null)
            .WithMessage("ISBN must be 10 or 13 digits.");
    }
    private static bool BeValidIsbn(string? v) => (v?.Replace("-", "").Length is 10 or 13);
}
```

Wire it as a MediatR `ValidationBehavior` (see `cqrs-and-mediator.md`) so every command is validated uniformly, or call it in an endpoint filter. Distinguish **input validation** (shape/format, at the boundary) from **domain invariants** (enforced inside the aggregate — never bypassable). Both matter; don't let the controller's validation be the only guard.

## Resilience with Polly (v8 pipelines)

Every call that crosses a process boundary can fail transiently. Wrap outbound calls in resilience strategies with **Polly**. Modern Polly (v8+) uses **resilience pipelines** via `Microsoft.Extensions.Resilience` / `Polly.Core`:

```csharp
var pipeline = new ResiliencePipelineBuilder<HttpResponseMessage>()
    .AddRetry(new HttpRetryStrategyOptions
    {
        MaxRetryAttempts = 5,
        BackoffType = DelayBackoffType.Exponential,
        UseJitter = true,
        Delay = TimeSpan.FromMilliseconds(200)
    })
    .AddCircuitBreaker(new HttpCircuitBreakerStrategyOptions
    {
        FailureRatio = 0.5,
        SamplingDuration = TimeSpan.FromSeconds(30),
        MinimumThroughput = 10,
        BreakDuration = TimeSpan.FromSeconds(15)
    })
    .AddTimeout(TimeSpan.FromSeconds(10))
    .Build();
```

Strategy menu:
- **Retry** with exponential backoff + jitter — for transient faults. Only retry **idempotent** operations (or use idempotency keys), or you risk duplicate side effects.
- **Circuit breaker** — stop hammering a failing dependency; fail fast while it recovers.
- **Timeout** — bound how long you wait.
- **Fallback** — a degraded but acceptable response (a cached value, a default).
- **Hedging / rate limiter / bulkhead** — advanced isolation strategies.

## Resilience for HTTP clients

Prefer `Microsoft.Extensions.Http.Resilience` to attach a standard pipeline to a typed/named `HttpClient`:

```csharp
builder.Services.AddHttpClient<AuthorsApiClient>()
    .AddStandardResilienceHandler();   // retry + circuit breaker + timeout, sensible defaults
```

The legacy approach (`Microsoft.Extensions.Http.Polly` + `HttpPolicyExtensions.HandleTransientHttpError().WaitAndRetryAsync(...)` with `AddPolicyHandler`) still works and is common in older codebases, but new code should use the `AddStandardResilienceHandler`/pipeline API. As with webhooks, retries assume idempotency — use PUT/upsert so a replayed call is safe.

## Externalizing business rules (when it pays off)

Most business rules belong **in code** — a well-named function or a command handler:

```csharp
public Money TotalLinePrice(Money unitPrice, int quantity, decimal taxRate) =>
    new(unitPrice.Amount * quantity * (1 + taxRate), unitPrice.CurrencyCode);
```

Centralize rules likely to change in one place (a domain service, a `CommonBusinessValues` constant), don't scatter copies. A dedicated **Business Rules Management System** (BRMS, e.g. Drools/DMN) is overkill ~99% of the time — its cost (performance, complexity, harder-to-read logic) only pays off when rules change *very* frequently, carry heavy regulatory/traceability needs, or require business-user editing and simulation. The DMN standard (decision tables/graphs) is the right model if you do externalize, callable over REST. Default: keep the rule in a function; the hard part is usually just *recognizing* you've written a business rule.

## Externalizing authorization (RBAC, ABAC, OPA)

Authorization is the most common large set of business rules. Match the mechanism to the need:

- **RBAC (role-based)** — permissions by role (admin/editor/reader). Sufficient for most apps; implement with ASP.NET Core authorization policies and claims from the identity provider (OIDC/JWT). Don't reach for more.
- **ABAC (attribute-based)** — decisions from attributes of subject/resource/context (an author may edit only their own record, verified by an email claim). Use ASP.NET Core policy handlers or resource-based authorization.
- **Externalized policy (OPA / XACML)** — when authorization rules are complex, change often, must be auditable/justifiable, and should live outside application code. **Open Policy Agent** (policies in Rego, queried over REST, deployable as a sidecar) is the modern choice; XACML is the older, heavier XML standard. OPA is to XACML roughly as REST is to SOAP.

```csharp
// RBAC + ABAC in ASP.NET Core
builder.Services.AddAuthorization(o =>
{
    o.AddPolicy("EditorsOnly", p => p.RequireRole("editor"));
    o.AddPolicy("OwnAuthorRecord", p => p.Requirements.Add(new SameEmailRequirement()));
});
```

Start with RBAC; add ABAC for ownership/context rules; externalize to OPA only when complexity and audit requirements truly justify the extra moving part.
