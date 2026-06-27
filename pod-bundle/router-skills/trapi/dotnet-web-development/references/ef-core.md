# Entity Framework Core: DbContext, migrations, LINQ, relationships

EF Core is the standard ORM for .NET — map C# entities to relational tables, query with LINQ, and persist changes without hand-writing SQL. It uses ADO.NET internally (e.g. `Microsoft.Data.SqlClient` for SQL Server). Examples target EF Core 10 with ASP.NET Core.

Table of contents:
1. Packages & providers
2. Entities & DbContext
3. Registering with DI
4. Migrations
5. Querying with LINQ
6. Loading related data
7. Inserting, updating, deleting
8. Relationships & configuration
9. Change tracking
10. Transactions & concurrency
11. Performance & gotchas

---

## 1. Packages & providers

Reference the provider for your database plus the design/tools packages for migrations:

- SQL Server: `Microsoft.EntityFrameworkCore.SqlServer`
- SQLite: `Microsoft.EntityFrameworkCore.Sqlite`
- PostgreSQL: `Npgsql.EntityFrameworkCore.PostgreSQL`
- Design-time (migrations): `Microsoft.EntityFrameworkCore.Design` and (for the `Package Manager Console`) `Microsoft.EntityFrameworkCore.Tools`
- CLI tool: `dotnet tool install --global dotnet-ef`

Keep entity models (POCOs) in a class library with no provider dependency, and the `DbContext` (with the provider) in a separate library, so the same entities can be reused client-side.

---

## 2. Entities & DbContext

An entity is a POCO; the context exposes `DbSet<T>` per table and configures the model.

```csharp
public class Product
{
    public int ProductId { get; set; }
    public string ProductName { get; set; } = null!;
    public decimal? UnitPrice { get; set; }
    public int? CategoryId { get; set; }
    public Category? Category { get; set; }          // navigation
}

public class NorthwindContext : DbContext
{
    public NorthwindContext(DbContextOptions<NorthwindContext> options) : base(options) { }
    public DbSet<Product> Products => Set<Product>();
    public DbSet<Category> Categories => Set<Category>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Product>(e =>
        {
            e.ToTable("Products");
            e.Property(p => p.ProductName).HasMaxLength(40).IsRequired();
            e.HasOne(p => p.Category).WithMany(c => c.Products)
             .HasForeignKey(p => p.CategoryId);
        });
    }
}
```

You can configure via Fluent API (`OnModelCreating`, preferred for complex mapping) or data annotations (`[Key]`, `[Required]`, `[StringLength]`, `[Column]`, `[Table]`, `[ForeignKey]`, `[InverseProperty]`). For an existing database, scaffold with `dotnet ef dbcontext scaffold "<connection>" Microsoft.EntityFrameworkCore.SqlServer`.

---

## 3. Registering with DI

Register the context as **Scoped** (one per HTTP request):

```csharp
builder.Services.AddDbContext<NorthwindContext>(options =>
    options.UseSqlServer(builder.Configuration.GetConnectionString("Northwind")));
```

Inject it into controllers/repositories via the constructor. In Blazor Server (where a scope = circuit, not request) or in singletons/background services, register `AddDbContextFactory<T>` and create a short-lived context per operation with `IDbContextFactory<T>.CreateDbContext()`.

---

## 4. Migrations

Migrations evolve the schema from the model. From the project folder containing the context (use `--project`/`--startup-project` if split):

```bash
dotnet ef migrations add InitialCreate     # generate a migration from current model
dotnet ef database update                  # apply pending migrations to the DB
dotnet ef migrations remove                # undo the last (unapplied) migration
dotnet ef database update PreviousMigration # roll back to a named migration
dotnet ef migrations script                # produce idempotent SQL for deployment
```

Each migration is a C# class with `Up`/`Down`; applied migrations are tracked in `__EFMigrationsHistory`. ASP.NET Core Identity ships its own migration (`CreateIdentitySchema`) creating `AspNetUsers`, `AspNetRoles`, etc. For production, prefer generating a SQL script and running it through your deployment pipeline over calling `Database.Migrate()` at startup (the dev-only `UseMigrationsEndPoint()` middleware can apply migrations via HTTP during development).

---

## 5. Querying with LINQ

`DbSet<T>` is `IQueryable<T>`; LINQ translates to SQL and executes when enumerated. Use **async** terminal operators in web apps so the thread returns to the pool:

```csharp
// Filtering, ordering, projection
var costly = await _db.Products
    .Where(p => p.UnitPrice > 50)
    .OrderByDescending(p => p.UnitPrice)
    .Select(p => new { p.ProductName, p.UnitPrice })   // projection -> efficient SELECT
    .ToListAsync();

var product = await _db.Products.SingleOrDefaultAsync(p => p.ProductId == id);
var byKey   = await _db.Products.FindAsync(id);        // checks the change tracker first
bool any    = await _db.Products.AnyAsync(p => p.Discontinued);
int count   = await _db.Products.CountAsync();
```

Use `AsNoTracking()` for read-only queries to skip change-tracking overhead. Avoid pulling whole tables into memory before filtering — keep the query composable until the terminal call. Raw SQL when needed: `FromSqlInterpolated($"SELECT * FROM Products WHERE UnitPrice > {price}")` (parameterized) or `ExecuteUpdate`/`ExecuteDelete` for set-based bulk operations without loading entities.

---

## 6. Loading related data

- **Eager** (recommended): `Include` / `ThenInclude`.

  ```csharp
  var orders = await _db.Orders
      .Include(o => o.Customer)
      .Include(o => o.OrderDetails).ThenInclude(d => d.Product)
      .ToListAsync();
  ```

- **Explicit**: `_db.Entry(product).Reference(p => p.Category).Load();` / `.Collection(...).Load();`
- **Lazy** (opt-in): install `Microsoft.EntityFrameworkCore.Proxies`, call `UseLazyLoadingProxies()`, mark navigations `virtual`. Convenient but causes the **N+1 problem** — a query per accessed navigation in a loop. Prefer eager loading or projection in web apps.

---

## 7. Inserting, updating, deleting

EF Core tracks loaded entities and writes changes on `SaveChanges`/`SaveChangesAsync` (returns the count of affected rows):

```csharp
// Insert
var entry = await _db.Suppliers.AddAsync(supplier);
int affected = await _db.SaveChangesAsync();
int newId = entry.Entity.SupplierId;        // DB-assigned key available after save

// Update (load, mutate, save)
var s = await _db.Suppliers.FindAsync(id);
if (s is not null) { s.Phone = newPhone; await _db.SaveChangesAsync(); }
// or attach a disconnected entity:
_db.Suppliers.Update(supplier);             // marks all properties modified
await _db.SaveChangesAsync();

// Delete
var toDelete = await _db.Suppliers.FindAsync(id);
if (toDelete is not null) { _db.Suppliers.Remove(toDelete); await _db.SaveChangesAsync(); }
```

Deleting a parent that still has FK-referencing children throws a referential-integrity error unless cascade delete is configured (or you delete children first). In web apps, only update the properties present in the submitted form to avoid clobbering or over-posting.

---

## 8. Relationships & configuration

EF Core models one-to-many, one-to-one, and many-to-many:

```csharp
// one-to-many
modelBuilder.Entity<Product>()
    .HasOne(p => p.Category).WithMany(c => c.Products).HasForeignKey(p => p.CategoryId);
// many-to-many (EF Core auto-creates the join table)
modelBuilder.Entity<Post>().HasMany(p => p.Tags).WithMany(t => t.Posts);
```

Configure delete behavior with `.OnDelete(DeleteBehavior.Cascade | Restrict | SetNull)`. Required vs optional relationship is inferred from FK nullability. Use `[InverseProperty]` to disambiguate multiple navigations between the same two types.

---

## 9. Change tracking

The tracker records the `EntityState` of each loaded entity: `Added`, `Modified`, `Deleted`, `Unchanged`, `Detached`. `SaveChanges` generates the SQL from these states. Inspect/override via `_db.Entry(entity).State` and `_db.ChangeTracker`. For read-only paths use `AsNoTracking()` (or set `ChangeTracker.QueryTrackingBehavior`) to avoid retaining and diffing entities — important for performance and to avoid stale cached state when combined with object caching.

---

## 10. Transactions & concurrency

`SaveChanges` is itself transactional (all-or-nothing for one call). For multiple operations, wrap in a transaction:

```csharp
await using var tx = await _db.Database.BeginTransactionAsync();
try { /* multiple SaveChangesAsync calls */ await tx.CommitAsync(); }
catch { await tx.RollbackAsync(); throw; }
```

Optimistic concurrency: add a `[Timestamp] byte[] RowVersion` (or `.IsConcurrencyToken()`) property; EF includes it in the `WHERE` clause and throws `DbUpdateConcurrencyException` if another write changed the row, which you handle (reload + merge or surface a 409 Conflict).

---

## 11. Performance & gotchas

- **N+1 queries** — eager-load (`Include`) or project (`Select`) instead of lazy-loading per row in a loop.
- **Tracking on reads** — add `AsNoTracking()`.
- **Pulling too much** — project to DTOs/anonymous types; never `ToList()` before filtering.
- **Async** — always `await` async terminal operators in request handlers; never mix sync and async on the same context, and never use one `DbContext` instance concurrently (it's not thread-safe).
- **Migrations drift** — keep the model and DB in sync; review generated migrations before applying.
- **Set-based bulk** — use `ExecuteUpdateAsync`/`ExecuteDeleteAsync` for large updates/deletes instead of loading thousands of entities.
- **Connection strings/secrets** — keep them out of source; load from configuration/environment (see deployment reference).
