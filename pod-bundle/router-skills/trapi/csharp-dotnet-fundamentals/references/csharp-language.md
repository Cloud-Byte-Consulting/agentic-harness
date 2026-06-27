# C# Language Reference

Core syntax and semantics of C# 12. Lead with the recipe; explanation follows.

## Contents
- Variables, types, and literals
- Operators
- Control flow (selection and iteration)
- Pattern matching
- Type conversion
- Methods and parameters
- Lambdas and local functions
- Exception handling
- Overflow checking
- Null handling

## Variables, types, and literals

```csharp
// Built-in value types (aliases for System.* types):
int i = -23;            // System.Int32; also sbyte/short/long, byte/ushort/uint/ulong
uint u = 23;            // unsigned (>= 0)
float f = 2.3f;         // single precision — 'f' suffix required
double d = 2.3;         // default for a literal with a decimal point
decimal money = 0.1m;   // 'm' suffix; use for money/exact reals, NOT double
bool ok = true;
char c = 'A';           // single quotes; one UTF-16 code unit
string s = "Bob";       // double quotes; reference type, compared by value

// Inference and target-typed new:
var name = "Alice";          // compiler infers string
List<int> nums = new();      // target-typed new() (C# 9+)

// Numeric legibility and bases:
int million = 1_000_000;     // digit separators
int bin = 0b_0001_1110;      // binary literal
int hex = 0x_001E_8480;      // hex literal
```

**Strings.** Literal `"a\tb"` honors escapes (`\t`, `\n`, `\\`, `\"`). Verbatim `@"C:\path"` disables escapes and allows newlines. Raw string literals (C# 11) use three-or-more `"""` and need no escaping — great for JSON/XML/SQL; the closing delimiter's indentation is stripped from each line. Interpolation: `$"Hi {name}, {price:C}"`. Combine: `$$"""{{expr}}"""` (the dollar count = number of braces that mark an interpolation; fewer braces are literal).

```csharp
string json = $$"""
    { "name": "{{name}}", "age": {{age}} }
    """;
```

Standard format specifiers in interpolation/`string.Format`: `N0` (number, thousands, 0 decimals), `C` (currency, culture-aware), `X`/`X2` (hex), `B8` (binary, 8 digits, C# 8+), `D`/`d` (long/short date), `{x,-5}` left-align in 5, `{x,7}` right-align in 7.

## Operators

- Arithmetic: `+ - * / %` (integer `/` truncates; `%` is remainder). Unary `++ --` (postfix returns then increments; prefix increments then returns — don't combine with assignment).
- Assignment: `=` and compound `+= -= *= /= %=`.
- Comparison: `== != < > <= >=`.
- Logical (bool): `& | ^` (always evaluate both operands), `&& ||` (short-circuit), `!`.
- Bitwise (integers): `& | ^ ~`; shifts `<< >>` (left shift by n ≈ multiply by 2ⁿ).
- Null operators: `?.` (null-conditional), `??` (coalesce), `??=` (coalesce-assign).
- Ternary conditional: `cond ? a : b`.
- Misc: `nameof(x)` → "x" as a string (great for exception messages); `sizeof(int)` → byte size; `typeof(T)` → `Type`; member access `.`, invocation `()`, indexer `[]`.

```csharp
int maxLen = input?.Length ?? 30;   // 30 if input is null
authorName ??= "unknown";           // assign only if currently null
```

## Control flow

### Selection
```csharp
if (password.Length < 8) { /* ... */ } else { /* ... */ }   // always use braces

// switch statement: each case ends with break / goto case / return, or is empty (fall-through label)
switch (n)
{
    case 1: Console.WriteLine("one"); break;
    case 2:
    case 3: Console.WriteLine("two or three"); break;   // shared section
    default: break;
}

// switch EXPRESSION (C# 8+): concise, returns a value; `_` is the discard/default
string size = n switch
{
    < 0 => "negative",
    0   => "zero",
    < 10 => "small",
    _   => "large"
};
```

### Iteration
```csharp
while (cond) { }                 // test at top
do { } while (cond);             // test at bottom — runs at least once
for (int i = 0; i < 10; i++) { } // counter loop
foreach (var item in sequence) { } // read-only; works on any IEnumerable/IEnumerable<T>
```
`foreach` requires a `GetEnumerator()` returning an object with `Current` and `MoveNext()` — formalized by `IEnumerable`/`IEnumerable<T>`. The loop variable is read-only. `break` exits the loop, `continue` skips to the next iteration, `return` exits the method.

## Pattern matching

```csharp
// type + declaration pattern
if (o is int n) Console.WriteLine(n * 2);

// property pattern + when guard in a switch
string describe = animal switch
{
    Cat { Legs: 4 } c => $"{c.Name} has four legs",
    Cat c when !c.IsDomestic => $"wild cat {c.Name}",
    Spider s when s.IsPoisonous => "run!",
    null => "no animal",
    _ => animal.GetType().Name
};

// list patterns (C# 11) — needs Length/Count + int indexer
string shape = arr switch
{
    [] => "empty",
    [var only] => $"one: {only}",
    [1, 2, .., 10] => "starts 1,2 … ends 10",
    [var first, .., var last] => $"{first}…{last}",
    [..] => "any"
};
```
Order list/relational patterns most-specific-first; a broad pattern (`[..]`, `_`) earlier makes later ones unreachable (compiler error).

## Type conversion

- **Implicit** (widening, lossless): `double d = anInt;`.
- **Explicit cast** (may lose data — you accept the risk): `int n = (int)aDouble;` (truncates).
- **Parsing text:** `int.Parse(s)` throws `FormatException`/`OverflowException` on bad input; prefer the `Try` pattern, which returns `bool` and sets an `out`:
```csharp
if (int.TryParse(input, out int count)) { /* use count */ }
else { /* handle invalid */ }
```
The `Try*` convention (`int.TryParse`, `Uri.TryCreate`, `dict.TryGetValue`) avoids exceptions for expected-bad input — exceptions are comparatively expensive.

## Methods and parameters

```csharp
static decimal CalculateTax(decimal amount, string region = "US") // optional param last
    => amount * RateFor(region);                                  // expression body

CalculateTax(amount: 149, region: "FR");  // named args — order-independent, self-documenting
```
- **Parameter passing:** by value (default; in-only), `out` (out-only, must be set in method), `ref` (in-and-out), `in` (read-only reference). `out`/`ref`/`in` can't have defaults.
- **Overloading:** same name, different parameter-type lists (return type alone is not enough to distinguish).
- **XML doc comments:** type `///` above a method to scaffold `<summary>`, `<param>`, `<returns>`. Local functions can't have them.

## Lambdas and local functions

```csharp
Func<string, bool> longName = name => name.Length > 4;     // expression lambda
Func<int, int, int> add = (a, b) => a + b;
Action<string> log = msg => Console.WriteLine(msg);        // returns void
var q = names.Where(n => n.Length > 4);                    // lambda as a delegate arg

static int Factorial(int n) =>                              // local function (recursion)
    n < 0 ? throw new ArgumentOutOfRangeException(nameof(n))
          : n == 0 ? 1 : n * Factorial(n - 1);
```
`=>` ("goes to") indicates the return expression. `Func<...,TResult>` returns a value; `Action<...>` returns void; `Predicate<T>` returns bool. C# 12 allows default values on lambda parameters.

## Exception handling

```csharp
try
{
    int age = int.Parse(input);
}
catch (OverflowException)              // most-specific first
{
    Console.WriteLine("Number too big/small.");
}
catch (FormatException) when (input.Contains('$'))  // exception filter
{
    Console.WriteLine("No dollar signs.");
}
catch (Exception ex)                   // most-general last
{
    Console.WriteLine($"{ex.GetType()}: {ex.Message}");
    throw;                             // rethrow preserving stack trace (NOT `throw ex;`)
}
finally
{
    // always runs — release resources here (or use `using`)
}
```
Rules of thumb: catch only what you can handle; let the rest propagate to a layer with enough context. Throw the right BCL type for usage errors: `ArgumentNullException`, `ArgumentException`, `ArgumentOutOfRangeException`, `InvalidOperationException`, `NotSupportedException`. Use **guard clauses**:
```csharp
ArgumentNullException.ThrowIfNull(manager);
ArgumentException.ThrowIfNullOrWhiteSpace(accountName);
ArgumentOutOfRangeException.ThrowIfNegativeOrZero(amount);
```

## Overflow checking

Integer arithmetic overflows silently by default (for speed). Force exceptions with `checked`; disable compile-time overflow detection with `unchecked`.
```csharp
checked
{
    int x = int.MaxValue;
    x++;                  // throws OverflowException instead of wrapping to int.MinValue
}
```

## Null handling

```csharp
string? maybe = GetName();          // ? = may be null (no warning when null)
string definite = "x";              // no ? = compiler warns if it might be null

if (maybe is not null)              // preferred null test (== can be overloaded)
    Console.WriteLine(maybe.Length);

int? len = maybe?.Length;           // null-conditional → int? (null if maybe is null)
int safeLen = maybe?.Length ?? 0;   // coalesce to a fallback
Console.WriteLine(maybe!.Length);   // null-forgiving: silence the warning (no runtime change)
```
A `?` on a **value** type changes the type (`int?` is `Nullable<int>`, with `.HasValue`, `.Value`, `.GetValueOrDefault()`). A `?` on a **reference** type does *not* change the type — it only adjusts compiler warnings. Project-level switch: `<Nullable>enable</Nullable>`; file-level `#nullable enable/disable`; suppress specific warnings with `#pragma warning disable CS8602` … `restore`.
