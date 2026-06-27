# Configuration in production, containerization & deployment

How to configure ASP.NET Core for production and package it for deployment. Examples target current .NET (8/9/10).

Table of contents:
1. Environments
2. Configuration & secrets in production
3. Kestrel & reverse proxies
4. Publishing
5. Containerizing with Docker
6. Overriding config in containers/Kubernetes
7. Health checks
8. Aspire (local orchestration of multi-project solutions)

---

## 1. Environments

`IWebHostEnvironment.EnvironmentName` comes from `ASPNETCORE_ENVIRONMENT` (or `DOTNET_ENVIRONMENT`): `Development`, `Staging`, `Production` (the default if unset). Branch behavior on it:

```csharp
if (app.Environment.IsDevelopment())
    app.MapOpenApi();                          // dev-only tooling
else
{
    app.UseExceptionHandler("/Home/Error");    // friendly error page in prod
    app.UseHsts();
}
```

Environment-specific settings load by convention: `appsettings.Production.json` overrides `appsettings.json` only when the environment is Production. The `<environment>` Tag Helper renders content per environment. `launchSettings.json` is **development-only** (sets URLs/env vars for `dotnet run` and IDEs) and is never deployed.

---

## 2. Configuration & secrets in production

Configuration sources merge in override order: `appsettings.json` → `appsettings.{Environment}.json` → user secrets (Development only) → environment variables → command-line args. So later sources win — ideal for overriding per environment without touching files.

Use the **Options pattern** for typed settings (see aspnet-core-fundamentals.md §4): `Configure<T>(GetSection("..."))`, inject `IOptions<T>`/`IOptionsSnapshot<T>`/`IOptionsMonitor<T>`, and add `.Validate(...).ValidateOnStart()` to fail fast on bad config.

**Never commit secrets** (connection strings, API keys, JWT signing keys). In development use `dotnet user-secrets set "ConnectionStrings:Northwind" "..."`. In production use environment variables and/or a vault (Azure Key Vault, AWS Secrets Manager). Map hierarchy in env vars with `__` (works cross-platform) or `:`:

```
ConnectionStrings__Northwind=Server=...;Database=Northwind;...
Logging__LogLevel__Default=Warning
```

---

## 3. Kestrel & reverse proxies

Kestrel is the built-in cross-platform server. In production it usually runs behind a reverse proxy (Nginx, IIS, a cloud load balancer) or directly in a container. When behind a proxy that terminates TLS, add forwarded-headers handling so the app sees the real scheme/IP:

```csharp
app.UseForwardedHeaders(new ForwardedHeadersOptions {
    ForwardedHeaders = ForwardedHeaders.XForwardedFor | ForwardedHeaders.XForwardedProto });
```

Bind URLs/ports via `ASPNETCORE_URLS` (e.g. `http://+:8080`), the `--urls` arg, or `appsettings` Kestrel config — not the dev-only `launchSettings.json`. `HTTP.sys` is a Windows-only alternative for Windows-auth edge cases.

---

## 4. Publishing

```bash
dotnet publish -c Release -o ./publish
```

This compiles, gathers static assets (with `MapStaticAssets`, pre-compresses them as gzip+brotli), and produces a framework-dependent deployment. Options: `--self-contained` (bundle the runtime), `-r linux-x64` (RID-specific), and native AOT for the `webapiaot` template (smaller/faster startup, minimal-API only). Run pending EF Core migrations as part of deployment by generating an idempotent SQL script (`dotnet ef migrations script --idempotent -o migrate.sql`) and applying it in the pipeline, rather than at app startup.

---

## 5. Containerizing with Docker

Containers give portability, consistency, isolation, efficiency, fast deploys; the trade-offs are shared-kernel security, orchestration complexity, and ephemeral-storage management. Use a multi-stage Dockerfile so the final image carries only the runtime, not the SDK:

```dockerfile
# Build stage
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src
COPY . .
RUN dotnet restore Northwind.WebApi/Northwind.WebApi.csproj
RUN dotnet publish Northwind.WebApi/Northwind.WebApi.csproj -c Release -o /app

# Runtime stage
FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS final
WORKDIR /app
COPY --from=build /app .
ENV ASPNETCORE_URLS=http://+:8080
EXPOSE 8080
ENTRYPOINT ["dotnet", "Northwind.WebApi.dll"]
```

`docker build -t northwind-api .` then `docker run --rm -p 8000:8080 northwind-api` (host 8000 → container 8080). The .NET SDK can also build images without a Dockerfile via `dotnet publish -t:PublishContainer`. Containerized apps deploy unchanged to Azure App Service, AKS, EKS, etc.

---

## 6. Overriding config in containers/Kubernetes

Inject configuration as environment variables at runtime so the immutable image stays generic and secrets stay out of it. In a Dockerfile (`ENV Logging__LogLevel__Default=Debug`), via `docker run -e`, or in a Kubernetes Deployment:

```yaml
spec:
  template:
    spec:
      containers:
      - name: api
        image: northwind-api:latest
        env:
        - name: Logging__LogLevel__Default
          value: Debug
        - name: ConnectionStrings__Northwind
          valueFrom: { secretKeyRef: { name: db, key: connection } }
```

These map to `Logging:LogLevel:Default` and `ConnectionStrings:Northwind` and override `appsettings.json`. Prefer secret stores for sensitive values.

---

## 7. Health checks

Expose liveness/readiness endpoints for orchestrators:

```csharp
builder.Services.AddHealthChecks()
    .AddDbContextCheck<NorthwindContext>();
// pipeline:
app.MapHealthChecks("/healthz");
```

Kubernetes probes / load balancers hit these to decide routing and restarts.

---

## 8. Aspire (local orchestration of multi-project solutions)

Aspire is a developer-time stack for building/running distributed cloud-native solutions locally — it does **not** run in production. It's opinionated, resilient (Polly), observable (OpenTelemetry + a dashboard), configurable, and assumes container deployment (needs Docker Desktop or Podman). Aspire 13 ships alongside .NET 10, is polyglot, and is versioned separately from .NET (only the latest release is supported).

An Aspire solution adds two projects: **AppHost** (a console app that orchestrates startup of all projects/containers/executables and wires service discovery) and **ServiceDefaults** (a class library centralizing telemetry, health checks, resilience, and service-discovery config that each project references). In the AppHost:

```csharp
var builder = DistributedApplication.CreateBuilder(args);
var cache = builder.AddRedis("cache");
var api   = builder.AddProject<Projects.Northwind_WebApi>("apiservice");
builder.AddProject<Projects.Northwind_Mvc>("web")
    .WithExternalHttpEndpoints()
    .WithReference(cache)
    .WithReference(api);                       // service discovery: refer to http://apiservice, not a real URL
builder.Build().Run();
```

Methods: `AddProject`/`AddContainer`/`AddExecutable`/`AddRedis`/etc. to compose; `WithReference`/`WithEnvironment`/`WithHttpEndpoint` to connect/configure. Run with F5/Ctrl+F5 (IDE) or `dotnet watch`/`aspire run`; the dashboard shows resources, console + structured logs, distributed traces, and metrics — using the same OpenTelemetry standards you'd wire to Grafana/Prometheus in production. Install templates: `dotnet new install Aspire.ProjectTemplates`.
