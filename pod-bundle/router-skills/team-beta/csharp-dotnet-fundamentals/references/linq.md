# LINQ

Language Integrated Query: declarative filtering, projection, sorting, joining, grouping, and aggregation over any `IEnumerable<T>` (and `IQueryable<T>` for providers like EF Core).

## Contents
- What LINQ is and why
- Deferred execution
- Operator catalog
- Filtering and sorting
- Projection
- Filtering by type
- Set operations
- Joining, grouping, lookups
- Aggregation
- Paging
- Query (comprehension) syntax
- IEnumerable vs IQueryable

## What LINQ is and why

Instead of imperative loops (you specify *how*), LINQ is declarative (you specify *what*) — more concise and less bug-prone. Its parts:
- **Extension methods** (required): `Where`, `Select`, `OrderBy`, etc., added by the static `Enumerable` class to anything implementing `IEnumerable<T>`. Import `System.Linq` (implicit by default).
- **Providers** (required): LINQ to Objects (in-memory), LINQ to Entities (EF Core → SQL), LINQ to XML.
- **Lambda expressions** (optional): the predicate/selector logic.
- **Query comprehension syntax** (optional): SQL-like `from`/`where`/`select` keywords.

```csharp
var matches = names.Where(n => n.EndsWith("m"))   // method syntax
                   .OrderBy(n => n.Length);
```

## Deferred execution

Most operators build a query (a *question*), they don't run it. Execution happens when you enumerate (`foreach`) or call a materializing method (`ToList`, `ToArray`, `ToDictionary`, `ToHashSet`, `ToLookup`, or aggregates like `Count`, `First`, `Sum`).

```csharp
var query = names.Where(n => n.EndsWith("m"));  // nothing runs yet
names[2] = "Jimmy";                              // source changed before enumeration...
foreach (var n in query) Console.WriteLine(n);   // ...so this reflects the change
```
Consequence: enumerating twice re-runs the whole query (re-hitting a DB, re-running side effects). Materialize once with `ToList()` if you'll iterate multiple times.

## Operator catalog

| Category | Methods |
|---|---|
| Filter | `Where`, `OfType<T>`, `Distinct`, `DistinctBy` |
| Project | `Select`, `SelectMany` (flatten), `Cast<T>`, `Chunk` |
| Sort | `OrderBy`, `OrderByDescending`, `ThenBy`, `ThenByDescending`, `Order`, `OrderDescending` (.NET 7+), `Reverse` |
| Element | `First`, `FirstOrDefault`, `Last`, `LastOrDefault`, `Single`, `SingleOrDefault`, `ElementAt(OrDefault)` |
| Quantify | `Any`, `All`, `Contains` |
| Partition | `Skip`, `SkipWhile`, `Take`, `TakeWhile`, `Take(range)` |
| Set | `Union`, `Intersect`, `Except`, `Concat`, `UnionBy`/`IntersectBy`/`ExceptBy` |
| Join/group | `Join`, `GroupJoin`, `GroupBy`, `Zip` |
| Aggregate | `Count`, `LongCount`, `Sum`, `Min`, `Max`, `Average`, `Aggregate`, `MinBy`, `MaxBy` |
| Materialize | `ToArray`, `ToList`, `ToDictionary`, `ToHashSet`, `ToLookup` |
| Convert (no alloc) | `AsEnumerable`, `AsQueryable` |
| Generate (static) | `Enumerable.Range`, `Enumerable.Repeat`, `Enumerable.Empty<T>` |

`As*` methods just re-type the sequence (fast, no allocation). `To*` methods allocate a new collection (force execution). The `*OrDefault` variants return `default(T)` (0 / null) instead of throwing when there's no match.

## Filtering and sorting

```csharp
var query = names
    .Where(name => name.Length > 4)     // predicate: Func<string,bool>
    .OrderBy(name => name.Length)       // primary sort key
    .ThenBy(name => name);              // tie-breaker
```
`OrderBy`/`ThenBy` take a key selector. `Order()` / `OrderDescending()` (.NET 7+) sort by the item itself (the item type must be `IComparable`). Format each chained call on its own line for readability.

## Projection

`Select` reshapes each item into a new type — often an anonymous type to grab just the needed fields:
```csharp
var shaped = products
    .Where(p => p.UnitPrice < 10m)
    .Select(p => new { p.ProductId, p.ProductName, p.UnitPrice }); // anonymous type
foreach (var p in shaped) Console.WriteLine($"{p.ProductName}: {p.UnitPrice:C}");
```
Against EF Core, projecting fewer columns generates a narrower SQL `SELECT`. `SelectMany` flattens nested sequences (e.g. each order → its lines → a single flat stream of lines).

## Filtering by type

`OfType<T>()` keeps only items assignable to `T` (respecting inheritance) — handy with heterogeneous sequences:
```csharp
IEnumerable<ArithmeticException> arith = exceptions.OfType<ArithmeticException>();
```

## Set operations

```csharp
cohort.Distinct();                 // remove duplicates
cohort.DistinctBy(n => n[..2]);    // dedupe by a key (first 2 chars)
a.Union(b);                        // unique items from both (a set)
a.Concat(b);                       // all items, keeping duplicates
a.Intersect(b);                    // in both
a.Except(b);                       // in a, not in b
a.Zip(b, (x, y) => $"{x}-{y}");    // pair by position (extra items dropped)
```

## Joining, grouping, lookups

```csharp
// Join: one flat row per match (inner join)
var joined = categories.Join(
    inner: products,
    outerKeySelector: c => c.CategoryId,
    innerKeySelector: p => p.CategoryId,
    resultSelector: (c, p) => new { c.CategoryName, p.ProductName });

// GroupBy: group items by a key
var byCat = products.GroupBy(p => p.CategoryId);
foreach (var g in byCat)
    Console.WriteLine($"{g.Key}: {g.Count()} products");

// ToLookup: a reusable, dictionary-like grouped structure (one-to-many)
ILookup<int, Product> lookup = products.ToLookup(p => p.CategoryId);
IEnumerable<Product> beverages = lookup[1];   // all products in category 1
```
`GroupJoin` is like `Join` but groups the inner matches under each outer item. With EF Core, some operators can't translate to SQL — call `.AsEnumerable()` first to switch to in-memory LINQ to Objects (less efficient, so push as much as possible into the DB).

## Aggregation

```csharp
int total = products.Count();
int discontinued = products.Count(p => p.Discontinued);   // filtered count
decimal max = products.Max(p => p.UnitPrice);
decimal avg = products.Average(p => p.UnitPrice);
decimal stockValue = products.Sum(p => p.UnitPrice * p.UnitsInStock);
var cheapest = products.MinBy(p => p.UnitPrice);          // the item, not the value
```
**Counting efficiently**: prefer an array's `Length` or a collection's `Count` property over `Count()`. Use `Any()` (not `Count() > 0`) to test for any items. `TryGetNonEnumeratedCount(out n)` (.NET 6+) gets a count only if cheap. `Count()` enumerates the whole sequence when there's no `Count`/`ICollection` — which can re-run side effects (e.g. a `Select(_ => Task.Run(...))` would launch the tasks again).

## Paging

```csharp
var page = products
    .OrderBy(p => p.ProductId)          // ALWAYS order before paging
    .Skip(currentPage * pageSize)
    .Take(pageSize);
```
Order first: providers don't guarantee a stable order otherwise, so pages could repeat or skip rows. Against EF Core this generates `ORDER BY ... LIMIT ... OFFSET ...`. `Chunk(size)` splits a sequence into fixed-size batches.

## Query (comprehension) syntax

Equivalent SQL-like sugar the compiler rewrites to method calls:
```csharp
var q = from name in names
        where name.Length > 4
        orderby name.Length, name
        select name;
```
`select` is required in query syntax; the `Select` method is optional in method syntax (the whole item is selected if omitted). Not every operator has a keyword (`Skip`/`Take` don't) — wrap query syntax in parentheses and continue with methods: `(from ... select x).Skip(80).Take(10)`. Learn both styles; you'll maintain code using each.

## IEnumerable vs IQueryable

- `IEnumerable<T>` / `IOrderedEnumerable<T>`: in-memory (LINQ to Objects). Lambdas are compiled delegates run by the CLR.
- `IQueryable<T>` / `IOrderedQueryable<T>`: a provider (e.g. EF Core) builds an **expression tree** that it translates to another language (SQL) and executes remotely. Inspect the generated SQL with `.ToQueryString()`.

Switch from queryable to in-memory with `.AsEnumerable()` when you need an operator the provider can't translate — but only after filtering/projecting in the database as much as possible.
