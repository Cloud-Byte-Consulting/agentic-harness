---
name: csharp-dotnet-fundamentals
description: >-
  Write, structure, and reason about C# 12 code on the .NET 8 runtime. Use for any C# task
  (.cs, .csproj, .sln files) covering language syntax, types (class, struct, record, enum,
  interface), generics, delegates and events, pattern matching, nullable reference types,
  async/await and Task, LINQ, collections, exceptions, and the base class library. Use for the
  dotnet CLI (dotnet new/build/run/test/publish/pack), NuGet packages, project and solution
  layout, target frameworks (net8.0, netstandard2.0), configuration (appsettings.json), and
  unit testing with xUnit or NUnit. Triggers include "C#", "dotnet", ".NET", "csharp", a
  NullReferenceException, CS-prefixed compiler errors, questions about async deadlocks,
  generic constraints, IDisposable, records vs classes, or how to package a class library. For
  ASP.NET and web see dotnet-web-development; for DDD and enterprise architecture see
  dotnet-enterprise-architecture.
---

# C# 12 and .NET 8 Fundamentals

Equips Claude to write idiomatic, correct, modern C# targeting the .NET 8 runtime — from language constructs and the type system to the base class library, the dotnet CLI, NuGet, project structure, and unit testing.

## When to use this skill

- Writing or reviewing any C# code (`*.cs`), project files (`*.csproj`), or solutions (`*.sln`).
- Choosing between language constructs: `class` vs `struct` vs `record`; field vs property; interface vs abstract class; array vs `List<T>` vs `Dictionary<TKey,TValue>`.
- Writing generics, delegates, events, pattern matching, or LINQ queries.
- Diagnosing `NullReferenceException`, nullable-reference warnings (CS8600–CS8625), `async`/`await` problems, boxing, overflow, or `IDisposable`/leak issues.
- Using the `dotnet` CLI (`new`, `build`, `run`, `test`, `publish`, `pack`), managing NuGet packages, picking a target framework, or reading `appsettings.json`.
- Setting up unit tests with xUnit or NUnit.
- Any prompt naming "C#", "dotnet", ".NET", or showing a `CS####` compiler error.

Boundary: ASP.NET Core, Razor, Blazor, web services → **dotnet-web-development**. Domain-driven design, layered/enterprise architecture, EF Core data modeling at scale → **dotnet-enterprise-architecture**. This skill covers the language, runtime, BCL, tooling, and testing that all of those build on.

## Core concepts

**C# is a statically typed, compiled language on .NET.** The Roslyn compiler turns `.cs` source into IL stored in an assembly (`.dll`/`.exe`); at runtime CoreCLR JIT-compiles IL to native code. C# keywords like `int` and `string` are aliases for BCL types (`System.Int32`, `System.String`). A project targets a framework moniker (`net8.0` for current apps, `netstandard2.0` for libraries shared with legacy platforms).

**Modern project defaults.** New SDK-style projects enable `ImplicitUsings` (auto-imports common namespaces like `System`, `System.Linq`) and `Nullable` (nullable-reference-type analysis) and use top-level statements — the compiler synthesizes the `Program` class and `<Main>$` method around your statements. Put functions and types at the bottom of `Program.cs` or, better, in a separate file with `partial class Program`.

```csharp
// Program.cs — top-level statements; net8.0; ImplicitUsings + Nullable enabled
Console.WriteLine("Hello, C#!");
int total = Add(2, 3);                 // local/static helper defined below
static int Add(int a, int b) => a + b; // expression-bodied method
```

**Value vs reference types.** `struct`/`record struct` are value types (data lives on the stack or inline; copied by value; compared by value). `class`/`record` are reference types (data on the heap; a variable holds a reference; `==` compares references — except `string` and `record`, which compare by value). Prefer `struct` only for small (≤16 bytes), immutable, value-like data; otherwise use `class`. See `references/types-and-oop.md`.

**Nullable reference types (NRT).** With `<Nullable>enable</Nullable>`, the compiler warns when a reference type that isn't declared `?` might be null. `string` is non-null-by-intent; `string?` may be null. NRT is *static analysis only* — it does not stop nulls at runtime, so still guard inputs. Key operators: `?.` (null-conditional), `??` / `??=` (null-coalescing), `!` (null-forgiving — suppresses the warning, no runtime effect). See `references/csharp-language.md`.

**Sequences and LINQ.** Anything implementing `IEnumerable<T>` (arrays, `List<T>`, query results) can be queried with LINQ extension methods (`Where`, `Select`, `OrderBy`, `GroupBy`, `Sum`...). LINQ uses **deferred execution**: building a query doesn't run it — enumeration (`foreach`) or a materializing method (`ToList`, `ToArray`, `Count`, `First`) does. See `references/linq.md`.

**Async.** `async`/`await` frees the calling thread while awaiting I/O. Methods return `Task`, `Task<T>`, or `ValueTask<T>`. See `references/async-await.md` for the rules that prevent deadlocks and avoid `async void`.

## Workflow: how to approach C# tasks

1. **Create or locate the project.** `dotnet new console -o MyApp` (or `classlib`, `xunit`), add to a solution with `dotnet sln add`. Reference siblings with `dotnet add reference`, packages with `dotnet add package`. See `references/dotnet-runtime-and-cli.md`.
2. **Pick the right type.** Data + behavior → `class`. Immutable data with value equality → `record`. Small value-like data → `struct`/`record struct`. Fixed set of named options → `enum` (add `[Flags]` for bit combinations). Contract for multiple implementers → `interface`.
3. **Encapsulate.** Make fields `private`; expose `public` auto-properties (`{ get; set; }`, or `{ get; init; }` for set-once immutability). Validate constructor arguments with guard clauses (`ArgumentNullException.ThrowIfNull(x)`, `ArgumentException.ThrowIfNullOrWhiteSpace(s)`).
4. **Prefer expressive constructs.** Switch expressions and pattern matching over long `if`/`switch` chains; LINQ over manual loops for filter/transform/aggregate; `var` for obvious local types.
5. **Handle errors honestly.** Catch only exceptions you can act on; otherwise let them propagate. Use `try`/`catch` ordered most-specific-first. Use `using`/`await using` (or declaration form `using var x = ...;`) for anything `IDisposable`/`IAsyncDisposable`. See `references/csharp-language.md`.
6. **Make I/O and CPU-bound waits async** where it helps responsiveness/scalability (`references/async-await.md`).
7. **Test it.** Write xUnit/NUnit tests in a separate project (Arrange–Act–Assert), run with `dotnet test`. See `references/testing.md`.
8. **Package/publish.** `dotnet pack` for a library NuGet package; `dotnet publish -c Release -r <rid> --self-contained` for a deployable app. See `references/dotnet-runtime-and-cli.md`.

### A representative type

```csharp
namespace MyApp.Domain;

public class BankAccount
{
    public string AccountName { get; }
    public decimal Balance { get; private set; }
    public static decimal InterestRate { get; set; }   // shared across instances

    public BankAccount(string accountName)
        => AccountName = accountName ?? throw new ArgumentNullException(nameof(accountName));

    public void Deposit(decimal amount)
    {
        if (amount <= 0)
            throw new ArgumentOutOfRangeException(nameof(amount), "Must be positive.");
        Balance += amount;
    }
}

// Immutable value-like data with built-in value equality + non-destructive `with`:
public record Customer(string FirstName, string LastName, DateOnly DateOfBirth);

var c1 = new Customer("Ada", "Lovelace", new(1815, 12, 10));
var c2 = c1 with { LastName = "Byron" };   // copy with one change
bool same = c1 == new Customer("Ada", "Lovelace", new(1815, 12, 10)); // true
```

## Common pitfalls & anti-patterns

- **Swallowing exceptions** with an empty `catch {}`. At minimum log and rethrow, or don't catch. Catch specific types, ordered most-derived first.
- **Assuming NRT prevents nulls.** It's compile-time analysis. Still validate parameters; `!` only hides the warning. Check with `is null` / `is not null`, not `== null` (the operator can be overloaded).
- **Re-enumerating a LINQ query.** Each `foreach`/`Count()` re-runs a deferred query (and may re-hit a database or re-run side effects). Materialize once with `ToList()` if you'll iterate multiple times. Prefer `Any()` over `Count() > 0`, and an array's `Length` or a collection's `Count` property over `Count()`.
- **`async void`** (except event handlers) — exceptions can't be awaited or caught; use `async Task`. Don't block on async with `.Result`/`.Wait()` — it can deadlock. See `references/async-await.md`.
- **Using non-generic collections** (`ArrayList`, `Hashtable`) — they box value types and lose type safety. Use `List<T>`, `Dictionary<TKey,TValue>`.
- **`string +=` in a loop** — allocates a new string each time. Use `StringBuilder` or `string.Join`.
- **Comparing reference types with `==` expecting value semantics** — only `string` and `record` do that by default. For other types implement equality or use a `record`.
- **`int.Parse` on untrusted input** — throws on bad input. Use `int.TryParse(s, out var n)` (the `Try*` pattern).
- **Wildcard NuGet versions** (`1.2.*`, `beta`) in production — pin to a fixed version that matches your target framework.
- **Floating-point for money** — use `decimal`, not `double`/`float`.

## Reference files

- `references/csharp-language.md` — syntax, operators, variables and literals, control flow, pattern matching, exceptions, null handling, methods/parameters, lambdas. Read when writing core logic or diagnosing CS-errors.
- `references/types-and-oop.md` — classes, structs, records, enums, fields/properties/indexers, constructors, access modifiers, inheritance, polymorphism, interfaces, delegates and events, value vs reference semantics, `IDisposable`. Read when designing types.
- `references/generics-and-collections.md` — generic types/methods/constraints, and the collection families (`List`, `Dictionary`, `HashSet`, `Queue`, `Stack`, sorted/immutable/concurrent), spans/indexes/ranges. Read when choosing or writing a data structure.
- `references/linq.md` — LINQ providers, deferred execution, the full operator catalog, query vs method syntax, projection, joins/groups/lookups, aggregation, paging. Read when querying or transforming sequences.
- `references/async-await.md` — `Task`/`ValueTask`, async method rules, cancellation, parallelism, and deadlock avoidance. Read for any concurrent or I/O-bound code.
- `references/dotnet-runtime-and-cli.md` — runtime/SDK versions, the dotnet CLI, project & solution structure, target frameworks, NuGet, publishing/AOT/trimming, configuration. Read for tooling, build, and packaging tasks.
- `references/testing.md` — xUnit and NUnit setup, Arrange–Act–Assert, assertions, fixtures, parameterized tests, mocking, `dotnet test`. Read when adding or running tests.
