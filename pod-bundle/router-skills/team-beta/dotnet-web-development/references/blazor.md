# Blazor: hosting models, render modes, components

Blazor builds interactive web UI with C# and Razor components (`.razor`) instead of a JavaScript framework. It shares the ASP.NET Core platform (DI, config, routing, auth). Examples target current .NET (8/9/10), where Blazor uses a unified full-stack model with per-component render modes.

Table of contents:
1. Hosting models & render modes
2. Project setup
3. Components
4. Parameters & cascading values
5. Data binding & events
6. Lifecycle
7. Routing & layouts
8. Forms & validation
9. DI, HttpClient & state
10. JS interop

---

## 1. Hosting models & render modes

Two execution targets, selectable per component via render mode (the early "pick one hosting model for the whole app" choice is gone since .NET 8):

- **Interactive Server** (`InteractiveServer`): component logic runs **on the server**; UI diffs stream to the browser over a SignalR circuit. Fast initial load, full .NET on the server, no code shipped to the client; but needs a persistent connection and latency per interaction.
- **Interactive WebAssembly** (`InteractiveWebAssembly`): component logic runs **in the browser** on the .NET WASM runtime. Works offline, no server round-trip per interaction; but a larger download and the browser's resource limits.
- **Interactive Auto** (`InteractiveAuto`): starts on the server (fast first paint) and transparently switches to WebAssembly on later visits once the runtime is cached.
- **Static SSR** (default, no render mode): the component renders to HTML once on the server with no interactivity — ideal for content pages; supports streaming rendering and enhanced navigation/form posts.

Set render mode globally on `<Routes>`/`HeadOutlet` or per component:

```razor
@rendermode InteractiveServer
```

or at the call site: `<Counter @rendermode="InteractiveWebAssembly" />`. Components that only render statically need no mode. Be deliberate: interactive modes have costs; use static SSR where you can.

---

## 2. Project setup

`dotnet new blazor` creates a Blazor Web App; flags select interactivity: `-int Server | WebAssembly | Auto` and `-ai` (per-page/global). `Program.cs`:

```csharp
var builder = WebApplication.CreateBuilder(args);
builder.Services.AddRazorComponents()
    .AddInteractiveServerComponents()        // for Server
    .AddInteractiveWebAssemblyComponents();  // for WebAssembly / Auto
var app = builder.Build();
app.UseStaticFiles();                         // or MapStaticAssets (.NET 9+)
app.UseAntiforgery();
app.MapRazorComponents<App>()
   .AddInteractiveServerRenderMode()
   .AddInteractiveWebAssemblyRenderMode();
app.Run();
```

A WebAssembly or Auto app has a separate `.Client` project for components that run in the browser; shared/static components live in the server project. `_Imports.razor` holds shared `@using` directives.

---

## 3. Components

A component is a `.razor` file combining markup and C# in an `@code` block; the class name is the file name. Use it as an element `<Counter />`.

```razor
@* Counter.razor *@
<h3>Counter</h3>
<p>Current count: @currentCount</p>
<button class="btn btn-primary" @onclick="IncrementCount">Click me</button>

@code {
    private int currentCount = 0;
    private void IncrementCount() => currentCount++;
}
```

`@` switches to C# (same Razor syntax as MVC views). Code-behind is also supported via a partial class (`Counter.razor.cs`). Scoped CSS goes in `Counter.razor.css`.

---

## 4. Parameters & cascading values

Inputs are `[Parameter]` properties; pass them as attributes.

```razor
@* ProductCard.razor *@
<div class="card">@Name — @Price.ToString("C")</div>
@code {
    [Parameter] public string Name { get; set; } = "";
    [Parameter] public decimal Price { get; set; }
    [Parameter] public EventCallback<string> OnSelected { get; set; }
}
```

`<ProductCard Name="Chai" Price="18.0m" OnSelected="HandleSelect" />`. Use `[Parameter(CaptureUnmatchedValues = true)] public IDictionary<string, object>? Attributes` to splat extra attributes. For data shared down a subtree, use `<CascadingValue Value="theme">...</CascadingValue>` and receive with `[CascadingParameter]`. Route segments bind via `[Parameter]` + `@page "/product/{Id:int}"`.

---

## 5. Data binding & events

Two-way binding with `@bind`:

```razor
<input @bind="searchText" @bind:event="oninput" />
<p>You typed: @searchText</p>
@code { private string searchText = ""; }
```

`@bind` defaults to the `onchange` event; `@bind:event="oninput"` updates per keystroke. Event handlers: `@onclick`, `@onchange`, `@onsubmit`, etc., bound to a method or lambda (`@onclick="() => Select(item)"`). Component-to-parent communication uses `EventCallback`/`EventCallback<T>` (`await OnSelected.InvokeAsync(value)`). Calling `StateHasChanged()` re-renders; it's invoked automatically after event handlers, but call it manually after async work that updates state outside the normal flow (use `InvokeAsync(StateHasChanged)` from a non-UI thread).

---

## 6. Lifecycle

Override these (sync and `...Async` variants exist): `OnInitialized` / `OnInitializedAsync` (once, after parameters first set — load data here), `OnParametersSet` / `OnParametersSetAsync` (whenever parameters change), `OnAfterRender(firstRender)` / `OnAfterRenderAsync` (after render; do JS interop here, gated on `firstRender`), `ShouldRender` (skip re-render), and `IDisposable`/`IAsyncDisposable` for cleanup.

```razor
@code {
    private List<Product>? products;
    protected override async Task OnInitializedAsync()
        => products = await Repo.GetProductsAsync();
}
```

---

## 7. Routing & layouts

A routable component declares `@page "/route/{Param:int?}"`. The Blazor `Router` matches the URL. Navigate in code with the injected `NavigationManager` (`Nav.NavigateTo("/products")`). Layouts derive from `LayoutComponentBase` with `@Body`; set per component with `@layout MainLayout` or app-wide via `_Imports`/`Router`. `<NavLink href="...">` adds an active CSS class automatically.

---

## 8. Forms & validation

Use `EditForm` with a model and data annotations:

```razor
<EditForm Model="@supplier" OnValidSubmit="Save">
    <DataAnnotationsValidator />
    <ValidationSummary />
    <InputText @bind-Value="supplier.CompanyName" class="form-control" />
    <ValidationMessage For="@(() => supplier.CompanyName)" />
    <button type="submit">Save</button>
</EditForm>
@code {
    private Supplier supplier = new();
    private async Task Save() => await Repo.UpdateAsync(supplier);
}
```

Built-in inputs: `InputText`, `InputTextArea`, `InputNumber`, `InputCheckbox`, `InputSelect`, `InputDate`, `InputRadioGroup`, `InputFile`. With static SSR, `EditForm` posts the form (requires `app.UseAntiforgery()`); with interactive modes it handles submission client-side.

---

## 9. DI, HttpClient & state

Inject services with `@inject IMyService Svc` (in markup) or `[Inject] private IMyService Svc { get; set; }` (in `@code`). Register them in `Program.cs` as usual. **Lifetime caveat:** in Interactive Server, a "scope" is the SignalR circuit (lives as long as the connection), not a single HTTP request — be careful sharing `DbContext` (prefer `IDbContextFactory<T>` and create a context per operation). In WebAssembly there's no per-request scope; scoped ≈ singleton within the app instance.

For data, WebAssembly components call APIs via `HttpClient` (register a typed/named client pointing at your API base address); Server components can use repositories/`DbContext` directly. Share state with scoped services or cascading values; for cross-circuit/global state use a singleton (Server) carefully.

---

## 10. JS interop

Call JS from C# via injected `IJSRuntime`:

```razor
@inject IJSRuntime JS
@code {
    protected override async Task OnAfterRenderAsync(bool firstRender)
    {
        if (firstRender)
            await JS.InvokeVoidAsync("console.log", "Component rendered");
    }
}
```

Use `InvokeAsync<T>` for return values. Call .NET from JS via `[JSInvokable]` methods and `DotNetObjectReference`. Do JS interop in `OnAfterRender(Async)` (the DOM exists by then), not in `OnInitialized`. Prefer `IJSObjectReference` + ES modules (`JS.InvokeAsync<IJSObjectReference>("import", "./script.js")`) to keep JS scoped to the component.
