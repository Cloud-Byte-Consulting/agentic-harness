# Caching & performance

Caching stores expensive-to-produce, infrequently-changing data closer to the consumer to cut latency and backend load. ASP.NET Core has several layers; pick the right one(s). Examples target current .NET (8/9/10).

Table of contents:
1. Guidelines & layers
2. Response caching (browser/CDN)
3. Output caching (server)
4. In-memory object caching
5. Distributed object caching
6. Hybrid caching
7. The Cache Tag Helper
8. Cache invalidation
9. Async & scalability

---

## 1. Guidelines & layers

Rules: cache data that costs a lot to generate and rarely changes; code must always be able to fall back to the source on a cache miss; cache is a limited resource — always set expirations and size limits and monitor hit rates. Don't over-cache (it hides bugs as stale data).

A request can be served from several caches, outermost first:

| Type | Where | Enable with |
| --- | --- | --- |
| Response | Browser, CDN, proxies | `[ResponseCache]` / `Cache-Control` headers |
| Output | Web server | `AddOutputCache` + `UseOutputCache` + `CacheOutput()` |
| Object | Web server / external | `IMemoryCache`, `IDistributedCache`, or `HybridCache` |

Note: CDN response caching often serves the large majority of real-world traffic. Configure your CDN before micro-optimizing code.

---

## 2. Response caching (browser/CDN)

`[ResponseCache]` adds `Cache-Control` headers telling clients/intermediaries how long they *may* cache (advisory — they can ignore it):

```csharp
[ResponseCache(Duration = 3600, Location = ResponseCacheLocation.Any)]  // public, max-age=3600
public IActionResult Index() => View();
```

- `Duration` → `max-age` (seconds). `Location`: `Any` → `public`, `Client` → `private`, `None` → `no-cache`. `NoStore = true` → `no-store` (sensitive data).
- Key `Cache-Control` directives: `public`/`private`, `no-cache` (revalidate before use), `no-store` (never store), `max-age`/`s-maxage` (shared caches), `must-revalidate`, `immutable`, `stale-while-revalidate`, `stale-if-error`. Static assets often use `public, max-age=31536000, immutable`.
- Anti-forgery automatically overrides caching to `no-store` for authenticated responses (you'll see a warning logged). So response caching mostly helps anonymous traffic.

---

## 3. Output caching (server)

Output caching (`Microsoft.AspNetCore.OutputCaching`, ASP.NET Core 7+) stores the rendered response on the server, skipping regeneration on a hit. Works in any ASP.NET Core app.

```csharp
builder.Services.AddOutputCache(o =>
{
    o.DefaultExpirationTimeSpan = TimeSpan.FromSeconds(10);     // default is 1 minute
    o.AddPolicy("views", p => p.SetVaryByQuery("alertstyle")); // vary only by named query keys
});
var app = builder.Build();
app.UseOutputCache();                                          // before endpoint mapping
app.MapGet("/cached", () => DateTime.Now.ToString()).CacheOutput();
app.MapControllerRoute("default", "{controller=Home}/{action=Index}/{id?}")
   .CacheOutput(policyName: "views");
```

By default output caching keys on the full path **including query string**, so `?color=red` vs `?color=blue` are separate entries. Use `SetVaryByQuery("...")` (or `""` for none) so irrelevant query params don't fragment the cache. Authenticated requests bypass it. Enable verbose logging via `"Microsoft.AspNetCore.OutputCaching": "Information"` to see hits/misses. Remember to disable it while debugging unexpected behavior.

---

## 4. In-memory object caching

`IMemoryCache` (`Microsoft.Extensions.Caching.Memory`) stores live objects in the server's memory. Avoid the legacy `System.Runtime.Caching`.

```csharp
builder.Services.AddSingleton<IMemoryCache>(new MemoryCache(new MemoryCacheOptions
{
    TrackStatistics = true,
    SizeLimit = 50          // in your chosen unit; entries must declare a Size
}));

// In a controller:
if (!_memoryCache.TryGetValue($"PROD{id}", out Product? model))
{
    model = await _db.Products.FindAsync(id);
    if (model is null) return NotFound();
    _memoryCache.Set($"PROD{id}", model, new MemoryCacheEntryOptions
    {
        SlidingExpiration = TimeSpan.FromSeconds(10),
        Size = 1
    });
}
```

**Expirations:** *Absolute* (`AbsoluteExpiration`/`AbsoluteExpirationRelativeToNow`) — evict at a fixed time; *Sliding* (`SlidingExpiration`) — reset on each access (good for popular items, but may never expire alone — combine with an absolute cap); set neither, both, or `CacheItemPriority.NeverRemove`. `RegisterPostEvictionCallback` runs logic on eviction. With multiple servers and in-memory cache you must enable sticky sessions so a client hits the same server's memory.

---

## 5. Distributed object caching

`IDistributedCache` survives restarts, is shared across servers (no sticky sessions), and frees local memory — but **only stores byte arrays**, so objects must be serialized.

```csharp
builder.Services.AddDistributedMemoryCache();         // dev/test in-process; not truly distributed
// Production: AddStackExchangeRedisCache(...) or AddDistributedSqlServerCache(...)

byte[]? bytes = await _cache.GetAsync("CATEGORIES");
List<Category>? categories = bytes is null ? null
    : JsonSerializer.Deserialize<List<Category>>(bytes);
if (categories is null)
{
    categories = await _db.Categories.ToListAsync();
    await _cache.SetAsync("CATEGORIES",
        JsonSerializer.SerializeToUtf8Bytes(categories),
        new DistributedCacheEntryOptions {
            SlidingExpiration = TimeSpan.FromMinutes(1),
            AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(20) });
}
```

Methods: `Get/Set/Remove/Refresh` (+ async). Implementations: **Redis** (`Microsoft.Extensions.Caching.StackExchangeRedis`), **SQL Server**, **NCache**, or a custom provider. Use a real distributed cache (Redis) in production.

---

## 6. Hybrid caching

`HybridCache` (`Microsoft.Extensions.Caching.Hybrid`, GA since .NET 9.x) unifies in-memory + distributed: it reads/writes the fast L1 in-memory cache first and falls through to an `IDistributedCache` L2 when registered. It adds **stampede protection** (concurrent requests for the same missing key wait for one factory call instead of all hammering the source) and configurable serialization (`System.Text.Json` by default; pluggable Protobuf/XML). It targets .NET Standard 2.0, so it runs on older runtimes too.

```csharp
builder.Services.AddHybridCache(o => o.DefaultEntryOptions = new HybridCacheEntryOptions
{
    Expiration = TimeSpan.FromSeconds(60),        // overall (L2)
    LocalCacheExpiration = TimeSpan.FromSeconds(30) // L1
});

// Get-or-create with a factory (the common, recommended pattern):
public Task<Customer?> RetrieveAsync(string id, CancellationToken ct = default) =>
    _cache.GetOrCreateAsync(
        key: id.ToUpper(),
        factory: async cancel => await _db.Customers.FirstOrDefaultAsync(c => c.CustomerId == id, ct),
        cancellationToken: ct).AsTask();
// Write-through on create/update; RemoveAsync on delete:
await _cache.SetAsync(c.CustomerId, c);
await _cache.RemoveAsync(c.CustomerId);
```

Prefer `HybridCache` for new code — it can replace direct `IMemoryCache`/`IDistributedCache` usage. Note: don't cache a `DbContext` (it has its own internal state); cache the entities it returns.

---

## 7. The Cache Tag Helper

Cache a fragment of a Razor View's HTML on the server:

```cshtml
<cache expires-after="@TimeSpan.FromSeconds(10)" vary-by-user="true">
  UTC: @DateTime.UtcNow.ToLongTimeString()
</cache>
```

Attributes: `enabled`, `expires-after` (default 20 min), `expires-on`, `expires-sliding`, and `vary-by-{header|user|route|cookie|query}`/`vary-by`. `<distributed-cache>` is identical but stores in the configured distributed cache (best for web farms/cloud); `<cache>` uses in-memory (single server or sticky-session farm).

---

## 8. Cache invalidation

Three strategies: **time-based** (absolute/sliding expiration — for data valid for a window), **dependency-based** (tie an entry to a `CancellationToken`/change token, invalidate on an external event like a file/DB change), and **manual** (explicitly `Remove`/overwrite after a write or admin update). Distributed invalidation must be deliberate — manage keys/expirations in the central store (e.g. Redis). When caching complex objects, ensure reference-data changes trigger invalidation.

---

## 9. Async & scalability

Scalability = handling more load without degrading. Beyond caching: use **async I/O** in request handlers (DB, files, HTTP) so the thread returns to the pool while the OS completes the I/O, increasing throughput. In ASP.NET Core each request starts on a thread-pool thread; synchronous blocking I/O can cause **thread-pool starvation** and timeouts. Async does **not** speed up CPU-bound work — keep heavy computation synchronous or offload it to a separate service/queue; don't `Task.Run` it to "scale" (extra threads burn memory/CPU; the thread pool grows adaptively via hill-climbing). Make controller actions `async Task<IActionResult>` and `await` async EF/HTTP calls. Scale vertically (bigger machine) or horizontally (more machines + distributed cache, no sticky sessions).
