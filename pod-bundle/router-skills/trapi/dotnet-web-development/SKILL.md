---
name: dotnet-web-development
description: >-
  Build web apps and HTTP APIs with ASP.NET Core on current .NET (8/9/10). Use for any ASP.NET
  Core work: minimal APIs and controller-based Web APIs, MVC, Razor Pages, Blazor (Server,
  WebAssembly, render modes, components), model binding and validation,
  OpenAPI/Swagger/Scalar, Entity Framework Core (DbContext, migrations, LINQ, relationships,
  change tracking), auth (ASP.NET Core Identity, cookies, JWT bearer, OAuth/OIDC, policies,
  roles, claims), the middleware/HTTP request pipeline, dependency injection and service
  lifetimes, configuration and the Options pattern, output/response/distributed/hybrid
  caching, Kestrel hosting, and Docker deployment. Triggers: Program.cs, builder.Services,
  app.MapGet, app.MapControllers, .cshtml, .razor, appsettings.json, ApiController, Authorize,
  DbContext, dotnet ef, 401/403/404/CORS/CSRF issues, "web API", "endpoint", "controller",
  "migration". Language basics belong to csharp-dotnet-fundamentals; architecture to
  dotnet-enterprise-architecture.
---

# ASP.NET Core Web Development

This skill equips Claude to build, secure, optimize, and ship web applications and HTTP APIs with ASP.NET Core on current .NET (examples target .NET 10; almost everything applies to .NET 8/9). It covers the shared platform (hosting, the middleware pipeline, DI, configuration), every UI/API model (minimal APIs, controllers, MVC, Razor Pages, Blazor), data access with EF Core, security, caching, and deployment.

## When to use this skill

- Creating or modifying any ASP.NET Core project: `dotnet new web`, `webapi`, `mvc`, `razor`, `blazor`; anything with `Program.cs`, `builder.Services`, `app.Map*`, `appsettings.json`, `launchSettings.json`.
- Building HTTP APIs — minimal API endpoints (`app.MapGet/MapPost`) or controllers (`[ApiController]`, `ControllerBase`); model binding, validation, content negotiation, OpenAPI docs.
- Server-rendered UI: MVC controllers + Razor Views, Razor Pages, Tag Helpers, layouts, localization.
- Blazor: components (`.razor`), Server vs WebAssembly, the unified render modes (`InteractiveServer`, `InteractiveWebAssembly`, `InteractiveAuto`).
- Data access with EF Core: `DbContext`, migrations (`dotnet ef`), LINQ queries, relationships, eager/lazy loading, change tracking.
- Auth: registering/logging in users (Identity), securing endpoints (`[Authorize]`, policies, roles, claims), JWT bearer for APIs, external OAuth/OIDC providers.
- Performance: response/output/distributed/hybrid caching, async I/O, scalability.
- Diagnosing symptoms: 401/403, 404 on a route, CORS errors, CSRF token failures, DI "Unable to resolve service", `NullReferenceException` from a view, middleware ordering bugs, `ObjectDisposedException` from lifetime mismatch.
- Configuration, the Options pattern, environments, and deploying via Docker/containers.

## Choosing the model

ASP.NET Core is a shared set of components (Kestrel, the pipeline, DI, config, routing) under several programming models. Learn the shared parts once; then pick a model:

| Need | Use |
| --- | --- |
| HTTP API, lean and fast | **Minimal APIs** (`app.MapGet`) — default `webapi` template since .NET 8 |
| HTTP API, large surface, content negotiation, filters, conventions | **Controllers** (`[ApiController]`, `webapi --use-controllers`) |
| Complex server-rendered site, testable separation | **MVC** (controllers + Razor Views) |
| Simple page-focused server-rendered site | **Razor Pages** |
| Rich interactive UI in C# instead of JS | **Blazor** (Server / WebAssembly / Auto) |

MVC and Razor Pages are mature, fully supported, and not deprecated. Blazor is Microsoft's recommended UI for new interactive apps, but there is no mandate to migrate. Minimal APIs and controllers are both first-class; minimal APIs do **not** do content negotiation out of the box (clients get JSON). Don't migrate a working app between models without a reason.

## Core concepts

**Hosting & `Program.cs`.** Modern apps use top-level statements with `WebApplication`:

```csharp
var builder = WebApplication.CreateBuilder(args);
// 1. Register services into the DI container.
builder.Services.AddControllers();          // or AddControllersWithViews, AddRazorPages, AddRazorComponents
builder.Services.AddDbContext<NorthwindContext>(o => o.UseSqlServer(cs));
var app = builder.Build();
// 2. Configure the HTTP request pipeline (middleware ORDER matters).
if (!app.Environment.IsDevelopment()) { app.UseExceptionHandler("/Error"); app.UseHsts(); }
app.UseHttpsRedirection();
app.UseRouting();
app.UseAuthentication();   // who are you?  (before authorization)
app.UseAuthorization();    // what can you do?
app.MapStaticAssets();     // .NET 9+; pre-compresses wwwroot assets. Use UseStaticFiles on .NET 8-.
app.MapControllers();      // or MapControllerRoute, MapRazorPages, MapRazorComponents<App>()
app.Run();                 // thread-blocking; starts Kestrel listening.
```

`builder` exposes `Configuration`, `Services`, `Environment`, `Logging`. The default builder already loads `appsettings.json`, `appsettings.{Environment}.json`, user secrets (Development), environment variables, and command-line args, in that override order.

**Middleware pipeline.** Each middleware is a `RequestDelegate` chained in order; it can short-circuit (return a response) or call `next()`. Order is load-bearing: `UseRouting` before endpoint-aware middleware, `UseAuthentication` before `UseAuthorization`, exception handling first. See `references/aspnet-core-fundamentals.md`.

**Dependency injection.** Register with a lifetime: **Transient** (new each request), **Scoped** (one per HTTP request — e.g. `DbContext`), **Singleton** (one for app lifetime). Prefer constructor injection. Never inject a Scoped service into a Singleton (causes captive-dependency / `ObjectDisposedException`); use `IServiceScopeFactory` in singletons/background services, and method injection (`[FromServices]`, or minimal-API lambda parameters) in middleware/filters/minimal endpoints.

**Routing.** MVC default route `{controller=Home}/{action=Index}/{id?}` maps a URL to a controller + action. Web API and minimal APIs use attribute/endpoint routing with constraints (`{id:int}`, `{id:guid}`, `{n:range(1,100)}`). A failed route constraint returns **404**, not 400 — matching happens before model binding.

**Configuration & Options.** Read settings via `IConfiguration["Section:Key"]`. Prefer the strongly-typed **Options pattern**: bind a section to a POCO with `builder.Services.Configure<MyOptions>(config.GetSection("My"))`, inject `IOptions<MyOptions>` (or `IOptionsSnapshot`/`IOptionsMonitor` for reload). Override in production with environment variables (`Section__Key` or `Section:Key`).

## Workflow / how to approach tasks

**Build an HTTP API.** Decide minimal vs controllers. Define DTOs/entities, register `DbContext` + services, map endpoints, return the right status codes (`Results.Ok`/`Created`/`NotFound`/`BadRequest`/`Problem` for minimal; `Ok()`/`CreatedAtRoute()`/`NotFound()` for controllers). Decorate controller actions with `[ProducesResponseType]` and let `[ApiController]` auto-return 400 with `ProblemDetails` on invalid models. Enable OpenAPI (`AddOpenApi` + `MapOpenApi`) and a UI (Scalar) in development. Make I/O-bound actions `async`. Full recipe: `references/web-apis.md`.

**Build a server-rendered site (MVC).** Create the project (`dotnet new mvc`), define a controller whose actions build a view model and `return View(model)`. Write Razor Views (`.cshtml`, `@model T`) using a shared `_Layout.cshtml`. Use Tag Helpers (`asp-for`, `asp-controller`, `asp-action`, `<form>`) over HTML Helpers. For data entry, define a GET action (show form) and a `[HttpPost]` action with `[ValidateAntiForgeryToken]`. Annotate models with validation attributes and check `ModelState.IsValid`. Details: `references/aspnet-core-fundamentals.md`.

**Build interactive UI (Blazor).** Create `.razor` components; pick a render mode per component (`@rendermode InteractiveServer` etc.) or globally. Server mode runs C# on the server over a SignalR circuit; WebAssembly runs in the browser; Auto picks at runtime. Use `[Parameter]`, `@bind`, `EventCallback`, lifecycle methods (`OnInitializedAsync`), and DI via `[Inject]`. Reference: `references/blazor.md`.

**Access data with EF Core.** Define entities and a `DbContext` with `DbSet<T>` properties; register it Scoped. Create the schema via migrations: `dotnet ef migrations add Initial` then `dotnet ef database update`. Query with LINQ; use `Include`/`ThenInclude` for related data, async terminal operators (`ToListAsync`, `SingleOrDefaultAsync`), and `AsNoTracking()` for read-only queries. Modify via `Add`/`Update`/`Remove` + `SaveChangesAsync()`. Reference: `references/ef-core.md`.

**Secure the app.** For sites, scaffold Identity (`--auth Individual`) — registration, login, cookies. Protect with `[Authorize]`, `[Authorize(Roles="Admin")]`, or policies (`AddPolicy(... RequireClaim/RequireRole)`). For APIs, use JWT bearer (`AddAuthentication().AddJwtBearer(...)`) and validate issuer/audience/lifetime/signing key. Mitigate CSRF with anti-forgery tokens, XSS with output encoding (Razor `@` encodes by default), over-posting with DTOs/`[Bind]`, SQL injection via parameterized EF Core queries. Reference: `references/auth-and-identity.md`.

**Optimize.** Apply caching at the right layer: `[ResponseCache]`/`Cache-Control` (browser/CDN), output caching (`AddOutputCache`/`CacheOutput`, server), object caching (`IMemoryCache`, `IDistributedCache`, or `HybridCache` for best-of-both with stampede protection). Cache data that's expensive to produce and changes rarely; always set expirations and size limits, and never depend on cached data being present. Reference: `references/caching-and-performance.md`.

**Configure & deploy.** Use the Options pattern for typed settings; keep secrets in environment variables / user secrets / a vault, never in source. Containerize with a multi-stage Dockerfile (`dotnet publish` into `mcr.microsoft.com/dotnet/aspnet`), override config via env vars (`ConnectionStrings__Default=...`). Reference: `references/deployment.md`.

## Common pitfalls & anti-patterns

- **`@page` in a Razor View** — adding `@page` to an MVC view turns it into a Razor Page; the controller no longer passes the model, so `Model` is null and you get a `NullReferenceException`. MVC views must NOT have `@page`.
- **Middleware in the wrong order** — `UseAuthorization` before `UseAuthentication`, or endpoint middleware before `UseRouting`, silently breaks auth/routing. Exception handling must be near the top.
- **Captive dependencies** — injecting a Scoped service (like `DbContext`) into a Singleton. Resolve a scope per use instead.
- **Blocking on async** — calling `.Result`/`.Wait()` on async I/O, or doing synchronous DB/HTTP/file I/O in a request handler, causes thread-pool starvation. Use `async`/`await` for I/O; keep CPU-bound work synchronous (don't `Task.Run` to "scale" it).
- **Over-posting / mass assignment** — binding directly to domain entities lets attackers set fields like `IsAdmin`. Bind to purpose-built DTOs or use `[Bind]`.
- **Missing `[ValidateAntiForgeryToken]`** on POST actions, or forgetting `@Html.AntiForgeryToken()` in forms.
- **Forgetting `[ApiController]`** on Web API controllers — you lose automatic 400/`ProblemDetails` and `[FromBody]` inference.
- **`MapStaticAssets` + runtime script injection** — Browser Link/Hot Reload injecting into pre-compressed static HTML causes `ERR_CONTENT_DECODING_FAILED`. Use `UseStaticFiles` for dynamically-modified HTML, or disable those VS features.
- **Trusting cached responses unconditionally** — response caching is advisory; CDNs/browsers may ignore it. Anti-forgery automatically disables caching for authenticated responses.
- **Confusing N+1 queries** — looping and lazy-loading related data per row. Use `Include`/projection. Use `AsNoTracking()` for reads.
- **Deprecated/legacy advice** — avoid `Startup.cs`/`ConfigureServices`+`Configure` (use top-level `Program.cs`), `System.Runtime.Caching` (use `Microsoft.Extensions.Caching.Memory`), Swashbuckle (use `Microsoft.AspNetCore.OpenApi` + Scalar/NSwag on .NET 9+), ASP.NET Web Forms (`.aspx`), and `System.Web`.

## Reference files

- `references/aspnet-core-fundamentals.md` — hosting, the middleware pipeline, DI lifetimes and injection mechanisms, routing, configuration & Options, MVC/Razor Pages, Razor syntax, Tag Helpers, localization. Open for any structural/pipeline/server-UI question.
- `references/web-apis.md` — minimal APIs vs controllers, model binding & validation, status codes & `ProblemDetails`, content negotiation, OpenAPI/Scalar, versioning, CORS, integration testing with `WebApplicationFactory`, and a note on OData & FastEndpoints. Open when building or testing an HTTP API.
- `references/blazor.md` — hosting models, render modes, components, parameters, binding, events, lifecycle, forms/validation, DI, JS interop. Open for any Blazor/`.razor` work.
- `references/ef-core.md` — `DbContext`, providers, migrations CLI, LINQ querying, loading related data, relationships/configuration, change tracking, transactions, performance. Open for any data-access work.
- `references/auth-and-identity.md` — ASP.NET Core Identity, cookies, JWT bearer, OAuth/OIDC, roles/claims/policies, filters, and security hardening (CSRF/XSS/over-posting/CORS). Open for any auth/security work.
- `references/caching-and-performance.md` — response/output/in-memory/distributed/hybrid caching, expirations, invalidation, async scalability. Open for performance work.
- `references/deployment.md` — environments, Kestrel/reverse proxy, the Options pattern in production, Dockerfile, container config override, health checks, Aspire orchestration. Open for shipping/ops work.
