# ASP.NET Core fundamentals: hosting, pipeline, DI, routing, config, MVC/Razor

Table of contents:
1. Hosting and `Program.cs`
2. The middleware / HTTP request pipeline
3. Dependency injection
4. Configuration and the Options pattern
5. Routing
6. MVC: controllers, actions, results
7. Razor Pages
8. Razor syntax, views, layouts
9. Tag Helpers & HTML Helpers
10. Localization & globalization
11. Project structure

---

## 1. Hosting and `Program.cs`

Current ASP.NET Core uses a single top-level `Program.cs` (the `Startup` class with `ConfigureServices`/`Configure` is legacy — you'll still see it in older code, but don't write it for new apps).

```csharp
var builder = WebApplication.CreateBuilder(args);
// --- Service registration (DI container) ---
builder.Services.AddControllersWithViews();
var app = builder.Build();
// --- HTTP pipeline (order matters) ---
if (!app.Environment.IsDevelopment())
{
    app.UseExceptionHandler("/Home/Error");
    app.UseHsts();
}
app.UseHttpsRedirection();
app.UseRouting();
app.UseAuthentication();
app.UseAuthorization();
app.MapStaticAssets();
app.MapControllerRoute(name: "default", pattern: "{controller=Home}/{action=Index}/{id?}");
app.Run();   // blocks; Kestrel listens for requests.
```

`WebApplication.CreateBuilder` wires up Kestrel, default configuration sources, logging, and DI. `builder` exposes:

- `Services` (`IServiceCollection`) — register dependencies here.
- `Configuration` (`ConfigurationManager`, implements `IConfiguration`) — merged settings.
- `Environment` (`IWebHostEnvironment`) — `IsDevelopment()`, `EnvironmentName`, `ContentRootPath`, `WebRootPath`.
- `Logging`, `Metrics`, `WebHost`, `Host`.

The environment is read from `DOTNET_ENVIRONMENT`/`ASPNETCORE_ENVIRONMENT` (set per-launch-profile in `launchSettings.json` during development — that file is dev-only and never deployed). On .NET 6+, the developer exception page is shown automatically in Development, so you no longer call `UseDeveloperExceptionPage()` yourself.

Kestrel is the cross-platform web server; in production it usually sits behind a reverse proxy (IIS, Nginx) or runs directly in a container.

`MapStaticAssets()` (.NET 9+) pre-compresses `wwwroot` assets at build/publish time and fingerprints ETags. On .NET 8 and earlier, use `app.UseStaticFiles()`. `MapStaticAssets` can conflict with tools that inject `<script>` into static HTML at runtime (Browser Link, Hot Reload) → `ERR_CONTENT_DECODING_FAILED`; use `UseStaticFiles` for such HTML or disable those VS features.

---

## 2. The middleware / HTTP request pipeline

A request flows through a chain of middleware (`RequestDelegate`); each can run code, then call the next, then run more code as the response flows back. Building blocks:

- `app.Use(async (context, next) => { /* before */ await next(); /* after */ });` — pass-through middleware.
- `app.Run(handler)` — terminal middleware (no `next`).
- `app.Map("/path", branch)` / `MapWhen` — branch the pipeline.
- `app.UseMiddleware<T>()` — a class with a `RequestDelegate` ctor param and an `InvokeAsync(HttpContext)` method.

**Order matters.** Canonical order: exception handling → HSTS/HTTPS redirect → static files → routing → CORS → authentication → authorization → custom → endpoints. Common bugs come from violating it (auth before authentication, custom middleware before `UseRouting` when it needs `context.GetEndpoint()`, which returns null before routing).

Inline endpoint-aware middleware example:

```csharp
app.Use(async (HttpContext context, Func<Task> next) =>
{
    var endpoint = context.GetEndpoint() as RouteEndpoint;   // null if before UseRouting
    if (context.Request.Path == "/health-lite")
    {
        await context.Response.WriteAsync("OK");             // terminal: returns, no next()
        return;
    }
    await next();
});
```

`HttpContext` exposes `Request`, `Response`, `User` (claims principal), `Connection`, `RequestServices` (the per-request service provider), and `Features`.

---

## 3. Dependency injection

ASP.NET Core has a built-in DI container implementing inversion of control. Register an interface→implementation with a lifetime:

```csharp
builder.Services.AddTransient<IEmailService, EmailService>();   // new instance every resolve
builder.Services.AddScoped<ICustomerRepository, CustomerRepository>(); // one per HTTP request
builder.Services.AddSingleton<IClock, SystemClock>();           // one for app lifetime
builder.Services.AddSingleton(new PreBuiltService());           // provide an instance
builder.Services.AddKeyedSingleton<ICache, BigCache>("big");    // keyed services
```

**Lifetimes:** Transient = lightweight/stateless. Scoped = per-request, e.g. `DbContext`. Singleton = shared resource/config.

**Injection mechanisms:**
- *Constructor injection* (preferred): dependencies as ctor params → stored in `readonly` fields. Makes dependencies explicit and mockable.
- *Method injection*: in MVC actions via `[FromServices]`; in minimal API lambdas the params are resolved from DI automatically; in middleware via `InvokeAsync` params (required for Scoped/Transient — the middleware itself is effectively a singleton).
- *Property injection*: not natively supported (3rd-party containers like Autofac do it).

**Captive dependency rule:** you may only consume a Scoped service from within a Scoped (or Transient resolved in a scope) service. Injecting Scoped into Singleton throws or causes `ObjectDisposedException`. In a `BackgroundService`/`IHostedService` (singletons), create a scope:

```csharp
using var scope = serviceScopeFactory.CreateScope();
var repo = scope.ServiceProvider.GetRequiredService<ICustomerRepository>();
```

Group feature registrations into extension methods (`this IServiceCollection`) to keep `Program.cs` clean — the framework does this (`AddControllersWithViews`). Disposable services created by the container are disposed by it automatically; never dispose them yourself.

`AddControllers` (APIs) ⊂ `AddControllersWithViews` (MVC) ⊂ adds views/Razor engine; `AddRazorPages` for page-based sites; `AddRazorComponents` for Blazor. `AddMvc` exists only for backward compatibility.

---

## 4. Configuration and the Options pattern

`IConfiguration` is a merged key-value view over all providers (in override order): `appsettings.json` → `appsettings.{Environment}.json` → user secrets (Development) → environment variables → command-line args. Hierarchy uses `:` (`Section:Key`) or, for env vars, `__` (`Section__Key`).

Read directly: `config["ConnectionStrings:Default"]` or `config.GetConnectionString("Default")` or `config.GetSection("Logging")`.

**Prefer the Options pattern** — bind a section to a typed POCO:

```csharp
public class NorthwindOptions { public string SiteTitle { get; set; } = ""; public int PagerSize { get; set; } = 10; }

builder.Services.Configure<NorthwindOptions>(builder.Configuration.GetSection("Northwind"));
// Validation:
builder.Services.AddOptions<NorthwindOptions>().Bind(builder.Configuration.GetSection("Northwind"))
    .Validate(o => o.PagerSize > 0, "PagerSize must be > 0").ValidateOnStart();
```

Inject `IOptions<NorthwindOptions>` (singleton snapshot), `IOptionsSnapshot<T>` (reloaded per scope/request), or `IOptionsMonitor<T>` (push notifications via `OnChange`). In production, override settings with environment variables (ideal for Docker/Kubernetes/serverless) — no rebuild needed, secrets stay out of source control.

Custom providers: implement `IConfigurationSource` + `ConfigurationProvider` and add with `builder.Configuration.Add(new MySource())`.

---

## 5. Routing

Endpoint routing builds a tree of endpoints; `UseRouting` marks where the match decision is made and endpoint middleware (`MapControllers`, etc.) executes it.

**MVC conventional route** `{controller=Home}/{action=Index}/{id?}`: `/Products/Detail/3` → `ProductsController.Detail(3)`; `/` → `HomeController.Index()`. Defaults via `=`, optional via `?`.

**Attribute routing** (Web API, also usable in MVC): `[Route("api/[controller]")]` on the class, `[HttpGet("{id:int}")]` on actions. `[controller]` token = class name minus the `Controller` suffix.

**Constraints** restrict matches by type/pattern: `{id:int}`, `{id:guid}`, `{price:decimal}`, `{flag:bool}`, `{n:range(1,100)}`, `{name:minlength(3)}`, `{name:alpha}`, `{id:regex(...)}`, `{x:required}`. Combine with colons: `{years:int:min(1)}`. A failed constraint → **404** (matching precedes binding/validation).

---

## 6. MVC: controllers, actions, results

A controller derives from `Controller` (view support) or `ControllerBase` (API, no views). An action returns `IActionResult`/`ActionResult<T>`/a POCO. Keep controllers thin — business logic lives in injected services.

```csharp
public class SuppliersController : Controller
{
    private readonly NorthwindContext _db;
    public SuppliersController(NorthwindContext db) => _db = db;

    public async Task<IActionResult> Index()
    {
        var model = await _db.Suppliers.OrderBy(s => s.Country).ToListAsync();
        return View(model);   // -> Views/Suppliers/Index.cshtml
    }

    public async Task<IActionResult> Edit(int? id)            // GET: show form
    {
        var s = await _db.Suppliers.FindAsync(id);
        return s is null ? NotFound() : View(s);
    }

    [HttpPost]
    [ValidateAntiForgeryToken]
    public async Task<IActionResult> Edit(Supplier supplier)  // POST: process form
    {
        if (!ModelState.IsValid) return View(supplier);
        _db.Suppliers.Update(supplier);
        await _db.SaveChangesAsync();
        return RedirectToAction(nameof(Index));
    }
}
```

`[HttpGet]`/`[HttpPost]`/etc. disambiguate same-named actions for the same route. `ControllerBase` helpers: `Ok`, `NotFound`, `BadRequest`, `Unauthorized`, `Forbid`, `Conflict`, `UnprocessableEntity`, `StatusCode`, `Problem`, `ValidationProblem`, `RedirectToAction`, `CreatedAtRoute`, `File`, `Json`. `Controller` adds `View`, `PartialView`, `ViewComponent`, plus `ViewData`/`ViewBag` (request-scoped dictionary) and `TempData` (survives one redirect).

Make I/O-bound actions `async Task<IActionResult>` and `await` async EF/HTTP calls so the thread returns to the pool.

---

## 7. Razor Pages

Page-focused alternative to MVC. A `.cshtml` page starts with `@page` and pairs with a `PageModel` (`.cshtml.cs`) exposing `OnGet`/`OnPost` handlers and `[BindProperty]`. Register with `AddRazorPages()` + `MapRazorPages()`. Even MVC-only apps keep `MapRazorPages()` because ASP.NET Core Identity's UI ships as Razor Pages.

```csharp
public class SuppliersModel : PageModel
{
    [BindProperty] public Supplier Input { get; set; } = new();
    public void OnGet() { }
    public async Task<IActionResult> OnPostAsync()
    {
        if (!ModelState.IsValid) return Page();
        // save...
        return RedirectToPage("Index");
    }
}
```

---

## 8. Razor syntax, views, layouts

A Razor View renders a model to HTML. `@` switches from HTML to C#; `@{ }` is a code block; `@expr` outputs (HTML-encoded by default). `@model T` declares the strongly-typed model, accessed via `Model`.

```cshtml
@model IEnumerable<Order>
@{ ViewData["Title"] = "Orders"; }
<h1>@ViewData["Title"]</h1>
@foreach (var o in Model) { <p>@o.OrderId — @o.OrderDate?.ToString("D")</p> }
@if (Model is null) { <div>None.</div> }
```

**Layouts** centralize chrome. `_ViewStart.cshtml` sets `Layout = "_Layout";` for all views in its folder tree. `_Layout.cshtml` defines the shell with `@RenderBody()` and optional `@await RenderSectionAsync("Scripts", required: false)`. A view fills a section with `@section Scripts { ... }`. `_ViewImports.cshtml` holds shared `@using` and `@addTagHelper` directives.

**File types:** Razor View / Layout / `_ViewStart` / `_ViewImports` = `.cshtml` (no `@page`); Razor Page = `.cshtml` + `@page`; Blazor component = `.razor`. **Putting `@page` on an MVC view makes it a Razor Page — the controller won't pass the model and `Model` will be null.**

Partial views: `<partial name="_Product" model="item" />` or `@await Html.PartialAsync("_Product", item)`. Display/editor templates live in `DisplayTemplates`/`EditorTemplates` folders.

---

## 9. Tag Helpers & HTML Helpers

Tag Helpers add server-side behavior to HTML-looking attributes — cleaner than HTML Helpers and friendlier to front-end devs. Enable with `@addTagHelper *, Microsoft.AspNetCore.Mvc.TagHelpers` (already in the template `_ViewImports.cshtml`).

```cshtml
<a asp-controller="Home" asp-action="Orders" asp-route-id="ALFKI">Orders</a>
<form asp-controller="Suppliers" asp-action="Edit" method="post">
  @Html.AntiForgeryToken()
  <label asp-for="CompanyName"></label>
  <input asp-for="CompanyName" class="form-control" />
  <span asp-validation-for="CompanyName" class="text-danger"></span>
  <button type="submit">Save</button>
</form>
```

Common Tag Helpers: **Anchor** (`asp-controller/action/route-*/fragment/protocol`), **Form** (auto-adds `method="post"` + anti-forgery token), **Label/Input/Select/TextArea** (`asp-for` generates `id`/`name`/`for` + client validation `data-*` from data annotations; honors `[Display(Name=...)]`), **Validation** (`asp-validation-for`, `asp-validation-summary`), **Cache** (`<cache expires-after="..." vary-by-*>` — server fragment cache), **Environment** (`<environment names="Development,Staging">`), **Image/Link/Script** (`asp-append-version` for cache-busting).

HTML Helpers (`@Html.ActionLink`, `@Html.DisplayFor`, `@Html.EditorFor`, `@Html.Raw`, `@Html.AntiForgeryToken`) are the older API, retained for cases Tag Helpers can't cover (e.g. rendering nested markup). Tag Helpers cannot be used inside Blazor components.

---

## 10. Localization & globalization

Globalization = formatting dates/numbers/currency per culture (e.g. `fr-CA` vs `fr-FR`); localization = translating UI text. Culture codes are ISO `language-REGION`.

1. Store translated strings in `.resx` resource files under a `Resources` folder, with culture suffixes (`Orders.resx` invariant, `Orders.fr.resx` neutral French, `Orders.fr-FR.resx`). Fallback: specific → neutral → invariant. Keep the **Name** keys identical across files (only translate Values); use human-readable English keys so a missing translation degrades gracefully.
2. Register: `builder.Services.AddLocalization(o => o.ResourcesPath = "Resources");` and `builder.Services.AddControllersWithViews().AddViewLocalization();`
3. Enable request localization early in the pipeline:

```csharp
string[] cultures = ["en-US", "en-GB", "fr", "fr-FR"];
var options = new RequestLocalizationOptions()
    .SetDefaultCulture(cultures[0]).AddSupportedCultures(cultures).AddSupportedUICultures(cultures);
app.UseRequestLocalization(options);
```

4. In a view, `@inject IViewLocalizer Localizer` then `@Localizer["Order ID"]`. The browser signals preference via the `Accept-Language` header (with `q` weights), a `?culture=` query param, or a culture cookie.

---

## 11. Project structure

Two common layouts:

- **Technical concerns** (template default): `/Controllers`, `/Models`, `/Views`, `/Services`. Familiar, IDE-friendly; scales poorly as related files scatter.
- **Feature folders / Vertical Slice Architecture**: group everything for a feature (`/Features/Catalog/{Controller,Service,Model,Views}`). Better modularity/isolation; easier to extract into separate projects later.

`wwwroot` holds static assets (css/js/lib/images). `appsettings.json`/`appsettings.Development.json` hold settings. `Properties/launchSettings.json` is dev-only launch config (URLs, env vars, launch profiles). Project file (`.csproj`) controls SDK, target framework, package references; with **Central Package Management** versions live in a root `Directory.Packages.props` (set `<ManagePackageVersionsCentrally>true`), and `<PackageReference>` elements omit versions. Consider `<TreatWarningsAsErrors>true</TreatWarningsAsErrors>` for discipline.
