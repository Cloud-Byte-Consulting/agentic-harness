# Authentication, authorization & security

Authentication = "who are you?" (verify identity, then issue a token/cookie). Authorization = "what can you do?" (gate resources by role/claim/policy). Examples target current .NET (8/9/10).

Table of contents:
1. ASP.NET Core Identity (cookie auth for sites)
2. Roles, claims & programmatic management
3. Authorization: [Authorize], roles, policies
4. JWT bearer for APIs
5. External providers (OAuth / OIDC)
6. Filters
7. Cookie security
8. Password guidance
9. Hardening: CSRF, XSS, over-posting, CORS, headers

---

## 1. ASP.NET Core Identity (cookie auth for sites)

Identity is the framework for managing users, passwords, roles, claims, and tokens, backed (usually) by EF Core. The MVC/Razor templates scaffold it with `--auth Individual` (or "Individual Accounts" in the IDE), which adds registration, login, account management, and an EF Core store (SQLite `app.db` by default, or SQL Server LocalDB with `--use-local-db`).

The pieces:

```csharp
// DbContext for Identity tables (AspNetUsers, AspNetRoles, ...).
public class ApplicationDbContext : IdentityDbContext<IdentityUser>
{
    public ApplicationDbContext(DbContextOptions<ApplicationDbContext> options) : base(options) { }
}

// Program.cs
builder.Services.AddDbContext<ApplicationDbContext>(o =>
    o.UseSqlServer(builder.Configuration.GetConnectionString("DefaultConnection")));
builder.Services.AddDefaultIdentity<IdentityUser>(o => o.SignIn.RequireConfirmedAccount = true)
    .AddEntityFrameworkStores<ApplicationDbContext>();
builder.Services.AddControllersWithViews();
// ...pipeline...
app.UseAuthentication();   // before UseAuthorization
app.UseAuthorization();
app.MapRazorPages();       // keep this: Identity's login/register UI is Razor Pages (Microsoft.AspNetCore.Identity.UI)
```

Create the Identity database with `dotnet ef database update` (runs the `CreateIdentitySchema` migration). The template follows double-opt-in: registration sends a confirmation email (`SignIn.RequireConfirmedAccount = true`); without an email provider you click the simulated confirmation link. Identity hashes passwords with PBKDF2 by default. The default login/register/access-denied pages are Razor Pages in the `Microsoft.AspNetCore.Identity.UI` package; scaffold them into your project to customize. `ManageUsers`/`SignInManager`/`UserManager` are the programmatic APIs.

---

## 2. Roles, claims & programmatic management

Enable role management (off by default) with `.AddRoles<IdentityRole>()` in the Identity setup. Manage users/roles with the injected `UserManager<IdentityUser>` and `RoleManager<IdentityRole>`:

```csharp
public class RolesController : Controller
{
    private readonly RoleManager<IdentityRole> _roles;
    private readonly UserManager<IdentityUser> _users;
    public RolesController(RoleManager<IdentityRole> roles, UserManager<IdentityUser> users)
        { _roles = roles; _users = users; }

    public async Task<IActionResult> Seed()
    {
        if (!await _roles.RoleExistsAsync("Administrators"))
            await _roles.CreateAsync(new IdentityRole("Administrators"));

        var user = await _users.FindByEmailAsync("test@example.com");
        if (user is not null && !await _users.IsInRoleAsync(user, "Administrators"))
            await _users.AddToRoleAsync(user, "Administrators");
        return Redirect("/");
    }
}
```

After a role is granted, the user must **log out and back in** for the new membership to load into their claims. `IdentityResult.Succeeded`/`.Errors` report outcomes. Add claims with `_users.AddClaimAsync(user, new Claim("Department", "Finance"))`.

---

## 3. Authorization: [Authorize], roles, policies

Gate controllers/actions/Razor Pages/Blazor components/minimal endpoints:

```csharp
[Authorize]                              // any authenticated user
[Authorize(Roles = "Sales,Marketing")]   // member of either role
[Authorize(Policy = "Over18Only")]       // a named policy
public IActionResult AdminOnly() => View();

[AllowAnonymous]                          // opt out (e.g. on a [Authorize] controller)
public IActionResult PublicPage() => View();
```

Minimal API: `app.MapGet("/secure", () => ...).RequireAuthorization("Over18Only");`

**Policies** encapsulate complex rules:

```csharp
builder.Services.AddAuthorization(options =>
{
    options.AddPolicy("RequireAdmin", p => p.RequireRole("Administrators"));
    options.AddPolicy("Over18Only", p => p.RequireClaim("Age", "Over18"));
    options.AddPolicy("FinanceHours", p => p.RequireClaim("Department", "Finance")
        .AddRequirements(new WorkingHoursRequirement()));
});
```

For logic that simple requirements can't express, implement `AuthorizationHandler<TRequirement>` (override `HandleRequirementAsync`, call `context.Succeed(requirement)` when satisfied) and register it as a service. Role-based is simpler; claims/policy-based is finer-grained and preferred for real apps.

---

## 4. JWT bearer for APIs

For stateless APIs, validate JSON Web Tokens (header.payload.signature, base64url) supplied in the `Authorization: Bearer <token>` header. Reference `Microsoft.AspNetCore.Authentication.JwtBearer`:

```csharp
builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
    .AddJwtBearer(options =>
    {
        options.TokenValidationParameters = new TokenValidationParameters
        {
            ValidateIssuer = true,
            ValidateAudience = true,
            ValidateLifetime = true,
            ValidateIssuerSigningKey = true,
            ValidIssuer = builder.Configuration["Jwt:Issuer"],
            ValidAudience = builder.Configuration["Jwt:Audience"],
            IssuerSigningKey = new SymmetricSecurityKey(
                Encoding.UTF8.GetBytes(builder.Configuration["Jwt:Key"]!))
        };
    });
```

JWTs are self-contained (claims travel in the token, no server session lookup), signed (tamper-evident), compact, and ideal for distributed/microservice/SPA/mobile scenarios. Issue them from a login endpoint after validating credentials (build with `JwtSecurityTokenHandler`/`JsonWebTokenHandler`). Keep the signing key in configuration/secrets, use short lifetimes plus refresh tokens, and always serve over HTTPS.

---

## 5. External providers (OAuth / OIDC)

Identity supports external sign-in via Google, Microsoft, Facebook, etc., using OAuth 2.0 / OpenID Connect — better UX (existing accounts), stronger security (provider MFA), faster onboarding. Register the app with the provider to get a client id + secret, reference the provider package (e.g. `Microsoft.AspNetCore.Authentication.Google`), and configure:

```csharp
builder.Services.AddAuthentication()
    .AddGoogle(o => { o.ClientId = cfg["Google:ClientId"]!; o.ClientSecret = cfg["Google:ClientSecret"]!; });
```

For full OIDC (e.g. Entra ID, Auth0, Okta), use `AddOpenIdConnect` with authority/client settings. Identity-management platforms like Auth0 offload session/password/token handling and follow security best practices for you.

---

## 6. Filters

Filters inject cross-cutting logic into the MVC request pipeline at action/controller/global scope. Execution order: **Authorization → Resource → Action → (action runs) → Result**, with Exception filters wrapping the lot. Register globally:

```csharp
builder.Services.AddControllersWithViews(o => o.Filters.Add(typeof(MyExceptionFilter)));
```

Implement `IAuthorizationFilter` (runs first, before model binding — `context.Result = new UnauthorizedResult()` to short-circuit), `IResourceFilter` (caching/short-circuit before binding), `IActionFilter` (`OnActionExecuting`/`OnActionExecuted` — logging, param tweaks), `IExceptionFilter` (centralized error handling — set `context.ExceptionHandled = true`), `IResultFilter` (tweak the response). Filters keep controllers thin and apply logic consistently. Note: a filter registered as a singleton can't constructor-inject scoped services — use `ServiceFilterAttribute`/`TypeFilterAttribute` or resolve within the filter method.

---

## 7. Cookie security

Identity uses cookie auth by default. Harden the cookie:

```csharp
builder.Services.ConfigureApplicationCookie(o =>
{
    o.Cookie.HttpOnly = true;                       // JS can't read it -> mitigates XSS theft
    o.Cookie.SecurePolicy = CookieSecurePolicy.Always; // HTTPS only -> mitigates MITM
    o.Cookie.SameSite = SameSiteMode.Strict;        // (or Lax) -> mitigates CSRF
    o.ExpireTimeSpan = TimeSpan.FromMinutes(30);    // short lifetime
    o.SlidingExpiration = true;
    o.LoginPath = "/Account/Login";
    o.AccessDeniedPath = "/Account/AccessDenied";
});
```

Cookie risks: XSS theft (use `HttpOnly`), CSRF (use `SameSite` + anti-forgery tokens), insecure transmission (use `Secure`/HTTPS), session hijacking/fixation (regenerate the cookie on login), and persistent-cookie exposure on shared devices. Combine with MFA and strong random session IDs.

---

## 8. Password guidance

Current NIST guidance (SP 800-63B): require **≥ 8** characters (recommend **≥ 15**); do **not** impose composition rules (no forced mix of character types); do **not** force periodic rotation; **do** force a change on evidence of compromise; check against breach/common-password lists. Identity hashes passwords automatically — never store or log plaintext.

---

## 9. Hardening: CSRF, XSS, over-posting, CORS, headers

- **CSRF** — anti-forgery tokens. `@Html.AntiForgeryToken()` (or the Form Tag Helper, which adds it automatically) emits a hidden `__RequestVerificationToken` field tied to a paired cookie; `[ValidateAntiForgeryToken]` on the POST action validates them. ASP.NET Core auto-disables response caching when anti-forgery is in play. For SPAs/AJAX, send the token in a header.
- **XSS** — Razor `@` HTML-encodes output by default. Only use `@Html.Raw` for trusted content. Set a Content-Security-Policy header.
- **Over-posting / mass assignment** — bind to DTOs, not domain entities; or restrict with `[Bind(nameof(...))]`. Never expose fields like `IsAdmin` to the model binder.
- **SQL injection** — EF Core parameterizes queries; for raw SQL use `FromSqlInterpolated`/parameters, never string concatenation.
- **CORS** — see web-apis.md; place `UseCors` after `UseRouting`, before `UseAuthorization`.
- **Transport** — `UseHttpsRedirection()` + `UseHsts()` (production) to force HTTPS; use secure headers (HSTS, CSP, X-Content-Type-Options).
