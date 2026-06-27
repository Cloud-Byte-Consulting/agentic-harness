# Web APIs: minimal APIs, controllers, binding, validation, OpenAPI, testing

Table of contents:
1. REST & HTTP background
2. Minimal APIs
3. Controller-based Web APIs
4. Model binding & validation
5. Status codes & ProblemDetails
6. Content negotiation
7. OpenAPI / Scalar / clients
8. Versioning & CORS
9. Integration testing
10. OData & FastEndpoints (when to reach for them)

---

## 1. REST & HTTP background

REST (Fielding, 2000) is an architectural style usually over HTTP: stateless requests, resource-oriented URIs, a uniform interface via HTTP methods, cacheability, layering, and (ideally) HATEOAS. Not every HTTP service is RESTful (SOAP, gRPC, GraphQL are HTTP services too).

Method semantics: **GET** read (safe, idempotent), **POST** create/action, **PUT** create-or-replace (idempotent), **PATCH** partial update, **DELETE** remove (idempotent). Default response format in ASP.NET Core is JSON (`application/json`). Caching uses `ETag`/`Last-Modified` + conditional `If-None-Match`/`If-Modified-Since` → `304 Not Modified`.

Status codes you'll return: `200 OK`, `201 Created` (+ `Location`), `202 Accepted`, `204 No Content`, `301/302/307` redirects, `400 Bad Request`, `401 Unauthorized` (not authenticated), `403 Forbidden` (authenticated but not allowed), `404 Not Found`, `405 Method Not Allowed`, `406 Not Acceptable`, `409 Conflict`, `415 Unsupported Media Type`, `422 Unprocessable Entity`, `429 Too Many Requests`, `500`, `503`.

---

## 2. Minimal APIs

Lean, low-ceremony HTTP endpoints. Default `dotnet new webapi` template since .NET 8. No controller classes; handlers are lambdas/methods with DI via parameters.

```csharp
var builder = WebApplication.CreateBuilder(args);
builder.Services.AddDbContext<NorthwindContext>(o => o.UseSqlServer(cs));
builder.Services.AddScoped<ICustomerRepository, CustomerRepository>();
builder.Services.AddOpenApi();
var app = builder.Build();
if (app.Environment.IsDevelopment()) app.MapOpenApi();

var customers = app.MapGroup("/api/customers").WithTags("Customers");

customers.MapGet("/", async (ICustomerRepository repo, string? country) =>
    Results.Ok(string.IsNullOrWhiteSpace(country)
        ? await repo.RetrieveAllAsync()
        : (await repo.RetrieveAllAsync()).Where(c => c.Country == country)));

customers.MapGet("/{id}", async (string id, ICustomerRepository repo) =>
    await repo.RetrieveAsync(id) is { } c ? Results.Ok(c) : Results.NotFound())
    .WithName("GetCustomer");

customers.MapPost("/", async (Customer c, ICustomerRepository repo) =>
{
    var added = await repo.CreateAsync(c);
    return added is null ? Results.BadRequest()
        : Results.CreatedAtRoute("GetCustomer", new { id = added.CustomerId.ToLower() }, added);
});

customers.MapPut("/{id}", async (string id, Customer c, ICustomerRepository repo) =>
{
    if (!string.Equals(id, c.CustomerId, StringComparison.OrdinalIgnoreCase)) return Results.BadRequest();
    return await repo.RetrieveAsync(id) is null ? Results.NotFound()
        : (await repo.UpdateAsync(c), Results.NoContent()).Item2;
});

customers.MapDelete("/{id}", async (string id, ICustomerRepository repo) =>
    await repo.DeleteAsync(id) == true ? Results.NoContent() : Results.NotFound());

app.Run();
```

- Return `Results.*` / `TypedResults.*` (`Ok`, `Created`, `CreatedAtRoute`, `NoContent`, `BadRequest`, `NotFound`, `Problem`, `ValidationProblem`, `File`, `Stream`). `TypedResults` gives compile-time types and better OpenAPI inference.
- Parameter binding source is inferred: route values → route, simple query string → query, complex types → body (JSON), `IFormFile`/`IFormFileCollection` → form, registered services → DI. Override with `[FromRoute]`/`[FromQuery]`/`[FromBody]`/`[FromHeader]`/`[FromServices]`/`[FromForm]`/`[AsParameters]`.
- `MapGroup` shares a prefix; chain `.WithTags`, `.WithName`, `.RequireAuthorization`, `.WithOpenApi`, `.AddEndpointFilter`, `.CacheOutput`.
- **Limitation:** minimal APIs do not perform content negotiation — clients get JSON unless you implement it. They also have no built-in model-state validation (use `MapToApiVersion`/filters or a library, or validate manually). The `webapiaot` template builds native-AOT-compatible minimal APIs.

---

## 3. Controller-based Web APIs

Richer feature set (content negotiation, filters, conventions, model binding metadata). Use `dotnet new webapi --use-controllers`.

```csharp
[Route("api/[controller]")]
[ApiController]
public class CustomersController : ControllerBase
{
    private readonly ICustomerRepository _repo;
    public CustomersController(ICustomerRepository repo) => _repo = repo;

    // GET api/customers?country=Germany
    [HttpGet]
    [ProducesResponseType<IEnumerable<Customer>>(StatusCodes.Status200OK)]
    public async Task<IEnumerable<Customer>> GetCustomers(string? country) =>
        string.IsNullOrWhiteSpace(country)
            ? await _repo.RetrieveAllAsync()
            : (await _repo.RetrieveAllAsync()).Where(c => c.Country == country);

    // GET api/customers/ALFKI
    [HttpGet("{id}", Name = nameof(GetCustomer))]
    [ProducesResponseType<Customer>(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetCustomer(string id) =>
        await _repo.RetrieveAsync(id, default) is { } c ? Ok(c) : NotFound();

    // POST api/customers
    [HttpPost]
    [ProducesResponseType<Customer>(StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> Create([FromBody] Customer c)
    {
        var added = await _repo.CreateAsync(c);
        return added is null ? BadRequest("Failed to create.")
            : CreatedAtRoute(nameof(GetCustomer), new { id = added.CustomerId.ToLower() }, added);
    }

    [HttpPut("{id}")]
    [ProducesResponseType(StatusCodes.Status204NoContent)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> Update(string id, [FromBody] Customer c)
    {
        if (!string.Equals(id, c.CustomerId, StringComparison.OrdinalIgnoreCase)) return BadRequest();
        if (await _repo.RetrieveAsync(id, default) is null) return NotFound();
        await _repo.UpdateAsync(c);
        return NoContent();
    }

    [HttpDelete("{id}")]
    public async Task<IActionResult> Delete(string id) =>
        await _repo.RetrieveAsync(id, default) is null ? NotFound()
            : (await _repo.DeleteAsync(id) == true ? NoContent() : BadRequest());
}
```

- `[ApiController]` enables API conventions: automatic `400` with `ValidationProblemDetails` on invalid `ModelState`, binding-source inference, attribute-routing requirement. **Always apply it** to Web API controllers.
- Action return types: a POCO/collection (serialized to the negotiated format), `IActionResult` (varied results), or `ActionResult<T>` (one type, varied status codes).
- Document with `[ProducesResponseType]` (the generic form `[ProducesResponseType<T>(StatusCodes.Status200OK)]` is current; on .NET 10 these attributes accept a `Description`).
- Map HTTP methods with `[HttpGet]`/`[HttpPost]`/`[HttpPut]`/`[HttpPatch]`/`[HttpDelete]`/`[HttpHead]`/`[HttpOptions]`, optionally with a route template + constraints.
- **DI efficiency tip:** constructor-injecting every dependency means all are instantiated for every action. For controllers with many actions, method-inject per-action with `[FromServices]` so only needed services resolve.

---

## 4. Model binding & validation

The default model binder maps request data → action parameters / model properties by name, from (priority high→low) **form fields**, **route values**, **query string**, and the **body** (for `[FromBody]`). It binds simple types (`int`, `string`, `DateTime`, `bool`), complex types (class/record/struct), and collections. Binding/validation results land in `ControllerBase.ModelState`.

Validation via data annotations on the model:

```csharp
public record Thing(
    [Range(1, 10)] int? Id,
    [Required] string? Color,
    [EmailAddress] string? Email);
```

Common attributes: `[Required]`, `[StringLength(n)]`/`[MaxLength]`/`[MinLength]`, `[Range(min,max)]`, `[EmailAddress]`, `[Phone]`, `[Url]`, `[RegularExpression]`, `[Compare]`, `[CreditCard]`. For complex rules, implement `IValidatableObject` or a custom `ValidationAttribute`. With `[ApiController]`, invalid `ModelState` auto-returns 400; in MVC, check `if (!ModelState.IsValid) return View(model);`. Client-side validation: Tag Helpers emit `data-val-*` attributes consumed by jQuery Unobtrusive Validation (`_ValidationScriptsPartial`).

**Security:** validation is also defense — prevents injection/XSS. Avoid **over-posting (mass assignment)**: never bind directly to a domain entity that has sensitive fields (e.g. `IsAdmin`); bind to a purpose-built DTO or restrict with `[Bind(nameof(...))]`. EF Core parameterizes queries, mitigating SQL injection.

---

## 5. Status codes & ProblemDetails

Return precise codes via helpers (`Ok`, `Created`, `CreatedAtRoute`, `Accepted`, `NoContent`, `BadRequest`, `NotFound`, `Conflict`, `UnprocessableEntity`, `StatusCode`, `Problem`, `ValidationProblem`). `[ApiController]` controllers auto-emit RFC 7807 `ProblemDetails` for 4xx. Customize:

```csharp
return BadRequest(new ProblemDetails {
    Status = StatusCodes.Status400BadRequest,
    Type = "https://example.com/errors/failed-to-delete",
    Title = $"Customer {id} found but failed to delete.",
    Detail = "...", Instance = HttpContext.Request.Path });
```

Wire a global handler with `app.UseExceptionHandler()` + `builder.Services.AddProblemDetails()` to convert unhandled exceptions to ProblemDetails responses.

---

## 6. Content negotiation

Controllers negotiate the response format from the client's `Accept` header. JSON via `System.Text.Json` is default. Add XML:

```csharp
builder.Services.AddControllers()
    .AddXmlSerializerFormatters();        // XmlSerializer (use [XmlIgnore] on interface/collection props it can't serialize)
//  .AddXmlDataContractSerializerFormatters();  // alternative; uses [DataContract]/[DataMember]
```

`XmlSerializer` cannot serialize interface-typed members (e.g. `ICollection<T>` navigation properties) — decorate them `[XmlIgnore]`. Default output formatters: `HttpNoContentOutputFormatter`, `StringOutputFormatter` (text/plain), `StreamOutputFormatter`, `SystemTextJsonOutputFormatter`. You can write custom formatters (e.g. CSV).

---

## 7. OpenAPI / Scalar / clients

On .NET 9+, use the first-party `Microsoft.AspNetCore.OpenApi` (Swashbuckle is effectively abandoned; don't add it for new projects).

```csharp
builder.Services.AddOpenApi();            // optionally AddOpenApi(o => o.OpenApiVersion = ...)
if (app.Environment.IsDevelopment())
{
    app.MapOpenApi();                      // serves /openapi/v1.json (default OpenAPI 3.1 on .NET 10)
    app.MapScalarApiReference();           // interactive UI (Scalar.AspNetCore) — the OpenApi pkg has no UI
}
```

The generated document describes paths, parameters, schemas, and validation. Serve YAML with `app.MapOpenApi("/openapi/{documentName}.yaml")`. Generate strongly-typed C# clients from the spec with **NSwag**. For "try it out", **Scalar** is the modern choice; **HTTP Editor** (Visual Studio) and the **REST Client** extension (VS Code) let you author `.http` files:

```http
@base = https://localhost:5091/api/customers/
GET {{base}}
###
GET {{base}}?country=USA
Accept: application/xml
###
POST {{base}}
Content-Type: application/json

{ "customerID": "ABCXY", "companyName": "ABC Corp", "country": "USA" }
```

Separate requests with `###`. Read env vars with `{{$processEnv MY_SQL_PWD}}`.

---

## 8. Versioning & CORS

**Versioning** with `Asp.Versioning.Mvc` (formerly `Microsoft.AspNetCore.Mvc.Versioning`): `builder.Services.AddApiVersioning(...).AddApiExplorer(...)`, then `[ApiVersion("1.0")]` and version-aware routes/query/header readers.

**CORS** for browser clients on other origins:

```csharp
builder.Services.AddCors(o => o.AddPolicy("spa", p =>
    p.WithOrigins("https://localhost:3000").AllowAnyHeader().AllowAnyMethod()));
// pipeline: AFTER UseRouting, BEFORE UseAuthorization:
app.UseCors("spa");
// or per-endpoint: endpoint.RequireCors("spa");
```

---

## 9. Integration testing

Test the full pipeline with `Microsoft.AspNetCore.Mvc.Testing` + xUnit. `WebApplicationFactory<TEntryPoint>` boots an in-memory `TestServer` and gives an `HttpClient`.

```csharp
public class CustomersApiTests : IClassFixture<WebApplicationFactory<Program>>
{
    private readonly HttpClient _client;
    public CustomersApiTests(WebApplicationFactory<Program> factory) => _client = factory.CreateClient();

    [Fact]
    public async Task Get_All_Returns200_Json()
    {
        var resp = await _client.GetAsync("/api/customers");
        resp.EnsureSuccessStatusCode();
        Assert.Equal("application/json", resp.Content.Headers.ContentType?.MediaType);
    }
}
```

On .NET 9 and earlier, the implicit `Program` class is internal — expose it with `public partial class Program { }` at the end of `Program.cs` (or `[assembly: InternalsVisibleTo(...)]`). .NET 10 streamlines this so test projects can reference `Program` from top-level statements directly. Override services for tests via `factory.WithWebHostBuilder(b => b.ConfigureServices(...))`. Prefer testing against a real database (containerized) over in-memory providers — only a real store is a true integration test; use transaction rollback or a SQL reset script to manage state. Use the test pyramid: many unit tests, fewer integration tests, fewest E2E.

---

## 10. OData & FastEndpoints (when to reach for them)

These are optional layers on ASP.NET Core covered for completeness:

**OData** (`Microsoft.AspNetCore.OData`) exposes data with a standardized query language in the URL — `$filter`, `$select`, `$expand`, `$orderby`, `$top`, `$skip`, `$count`. Register an Entity Data Model with `ODataConventionModelBuilder` and `.AddOData(o => o.Select().Expand().Filter().OrderBy().Count().SetMaxTop(100))`; decorate controller `Get` methods returning `IQueryable<T>` with `[EnableQuery]` so OData translates the query options into an optimized LINQ/SQL query. Powerful for ad-hoc client querying; downsides are query complexity and exposure of the data shape — restrict allowed operations.

**FastEndpoints** (`FastEndpoints`) is a fast third-party alternative to controllers using the REPR (Request-Endpoint-Response) pattern — one class per endpoint. `builder.Services.AddFastEndpoints()` + `app.UseFastEndpoints()`. Define an endpoint by deriving from `Endpoint<TRequest, TResponse>`, overriding `Configure()` (`Get("/...")`, `AllowAnonymous()`, `Roles(...)`) and `HandleAsync(req, ct)` (`await Send.OkAsync(...)`). Validation integrates with FluentValidation. Choose it when you want minimal-API-like performance with stronger structure than controllers.
