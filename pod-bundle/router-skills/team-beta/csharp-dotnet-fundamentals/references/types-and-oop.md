# Types and Object-Oriented Programming

Designing custom types in C# 12: classes, structs, records, enums, members, inheritance, interfaces, delegates/events, and memory semantics.

## Contents
- Choosing a type kind
- Fields, properties, indexers
- Constructors and `required`/`init`
- Access modifiers
- Static members
- Enums and `[Flags]`
- Records
- Value vs reference types and memory
- Inheritance and polymorphism
- Interfaces
- Operator overloading
- Delegates and events
- IDisposable and finalizers

## Choosing a type kind

| Need | Use |
|---|---|
| Data + behavior, identity matters, will inherit | `class` |
| Immutable data, value equality, "non-destructive" copies | `record` (class) |
| Small (≤16 bytes), value-like, won't inherit | `struct` |
| Small immutable value with auto equality | `record struct` |
| Fixed set of named constants / bit flags | `enum` |
| Contract for unrelated implementers | `interface` |

Every type ultimately inherits from `System.Object` (`object`), giving `ToString()`, `Equals()`, `GetHashCode()`, `GetType()`.

## Fields, properties, indexers

Fields store data; properties control access to it (the preferred public surface). Make fields `private`/`protected`; expose properties.

```csharp
public class Person
{
    public string? Name { get; set; }                 // auto-property
    public DateTimeOffset Born { get; init; }          // set only during init
    public bool IsAdult => Age >= 18;                  // computed (get-only) property
    private readonly List<Person> _children = new();   // backing field
    public IReadOnlyList<Person> Children => _children;

    // full property with backing field + validation:
    private int _age;
    public int Age
    {
        get => _age;
        set => _age = value < 0 ? throw new ArgumentOutOfRangeException(nameof(value)) : value;
    }

    // indexer — array-style access:
    public Person this[int i] => _children[i];
}
```
`{ get; set; }` mutable; `{ get; init; }` set-once (during object initializer or constructor) then immutable; `{ get; }` get-only (settable only in the constructor). Use `DateTimeOffset` for moments in time, `DateOnly`/`TimeOnly` (.NET 6+) for calendar dates/clock times.

## Constructors and required/init

```csharp
public class Book
{
    public required string Title { get; init; }   // C# 11: must be set at construction
    public required string Isbn  { get; init; }
    public string? Author { get; set; }

    public Book() { }                              // for object-initializer use
    [SetsRequiredMembers]                          // tells compiler this ctor sets all required members
    public Book(string isbn, string title) { Isbn = isbn; Title = title; }
}

var b = new Book { Isbn = "978-...", Title = "C# 12" };   // object initializer
```
A class with no constructor gets a public parameterless one. Multiple constructors can chain: `public Address(string city) : this() { City = city; }`. Set fields/properties via constructor parameters or object-initializer `{ }` syntax.

**Primary constructors** (C# 12) put parameters on the type declaration. For a `class`, the parameters are in scope for the whole body but are *not* auto-exposed as properties (unlike records):
```csharp
public class Headset(string productName, decimal price)
{
    public string ProductName { get; } = productName;   // surface explicitly
    public decimal Price { get; set; } = price;
}
```

## Access modifiers

| Modifier | Visible to |
|---|---|
| `private` | the type only (default for members) |
| `internal` | the type + same assembly (default for top-level types) |
| `protected` | the type + derived types |
| `public` | everywhere |
| `protected internal` | same assembly **or** derived types |
| `private protected` | derived types **in the same assembly** |
| `file` (C# 11) | the same source file only |

Always state the modifier explicitly. Make a class `public` to use it from another assembly.

## Static members

`static` members belong to the type, not an instance — one shared copy. `const` fields are compile-time literals (copied into callers; never change). `readonly` fields are set once (at declaration or in a constructor), can be computed at runtime, and are referenced live.
```csharp
public const string Species = "Homo Sapiens";     // compile-time constant
public static decimal InterestRate { get; set; }  // one value shared by all instances
public readonly DateTime CreatedAt = DateTime.Now; // per-instance, set once
```
Prefer `readonly` over `const` for anything that might change between versions.

## Enums and [Flags]

```csharp
public enum Wonder { Pyramid, Gardens, Zeus }      // backed by int starting at 0

[Flags]                                            // allow bitwise combinations
public enum Days : byte                            // pick a small backing type
{
    None = 0, Mon = 1, Tue = 2, Wed = 4, Thu = 8, Fri = 16
}
Days work = Days.Mon | Days.Wed;                   // combine
bool hasMon = work.HasFlag(Days.Mon);
```
With `[Flags]`, assign power-of-two values so they occupy distinct bits; `ToString()` then returns a comma-separated list. Choose backing type by option count: `byte`≤8, `ushort`≤16, `uint`≤32, `ulong`≤64.

## Records

A `record` (reference type) gives **value equality**, a readable `ToString()`, deconstruction, and non-destructive copies via `with` — ideal for DTOs/immutable data.
```csharp
public record Customer(string FirstName, string LastName);   // positional record
var a = new Customer("Ada", "Byron");
var b = a with { LastName = "Lovelace" };       // copy with one change
bool eq = a == new Customer("Ada", "Byron");    // true — value equality
var (first, last) = a;                          // deconstruction

public record class Animal(string Name);        // explicit 'class' recommended
public record struct Point(int X, int Y);       // value type; mutable unless `readonly record struct`
```
Two records are equal when all their members are equal; two plain classes are equal only if they reference the same object. `init`-only properties give partial immutability without the full record machinery.

## Value vs reference types and memory

- **Value types** (`struct`, `record struct`, all numeric types, `bool`, `char`, `DateTime`, `Guid`): the data lives where the variable lives (stack for locals, inline within an owning object on the heap). Copied by value; `==` compares values (you must implement it for a plain `struct`, or use `record struct`). Cannot be inherited from. Always have a default (all-zero) value.
- **Reference types** (`class`, `record`, `string`, arrays, delegates): the object lives on the heap; the variable holds its address. Assignment copies the reference, so two variables can point to the same object. `==` compares references (except `string`/`record`). The GC reclaims unreferenced heap memory.
- **Boxing**: implicitly wrapping a value type in an `object` moves it to the heap (e.g. `object o = 42;`). Unboxing (`(int)o`) is explicit. Boxing is slow — avoid it in hot loops; generics (`List<int>`) avoid it.

Microsoft guidance: prefer `struct` only when total field size ≤ ~16 bytes, all fields are value types, and you won't inherit; otherwise use `class`.

## Inheritance and polymorphism

```csharp
public class Employee : Person                  // single base class
{
    public string? EmployeeCode { get; set; }

    public override string ToString()           // override a virtual base member
        => $"{Name} ({EmployeeCode})";
}
```
- Mark a base member `virtual` to allow overriding; `override` it in the derived class; `sealed override` to stop further overriding.
- `abstract` class can't be instantiated and may have `abstract` (unimplemented) members subclasses must implement. `Stream` is abstract; `FileStream`/`MemoryStream` are concrete.
- `new` *hides* (doesn't override) an inherited member — generally avoid; use `virtual`/`override`.
- `this` = current instance; `base` = base-class implementation (`base.ToString()`).
- Cast down with pattern matching: `if (person is Employee e) { ... }` or `as` (`var e = person as Employee;` → null if not that type).

## Interfaces

```csharp
public interface IComparable<in T> { int CompareTo(T? other); }

public class Person : IComparable<Person?>
{
    public int CompareTo(Person? other)         // <0 before, 0 equal, >0 after
        => string.Compare(Name, other?.Name, StringComparison.Ordinal);
}
```
Implementing an interface is a contract promising specific members. Common BCL interfaces: `IComparable<T>` (sortable), `IComparer<T>` (external comparison), `IEnumerable<T>` (iterable), `IDisposable` (resource cleanup), `IEquatable<T>`, `IFormattable`. A type can implement many interfaces. Use **explicit implementation** (`int IGamePlayer.Lose() { ... }`) only when two interfaces collide on a member name. Default interface methods (C# 8, needs .NET Core 3+/.NET Standard 2.1) let you add members with a body later — use sparingly.

## Operator overloading

```csharp
public static DisplacementVector operator +(DisplacementVector a, DisplacementVector b)
    => new(a.X + b.X, a.Y + b.Y);
```
Define operators as `public static`. Always provide a named-method equivalent too (operators aren't shown in IntelliSense and aren't usable from every .NET language).

## Delegates and events

A delegate is a type-safe reference to a method. Events are delegates used for one-to-many notification.
```csharp
public class Button
{
    public event EventHandler? Click;                 // EventHandler(object?, EventArgs)
    public void Press() => Click?.Invoke(this, EventArgs.Empty);  // raise if subscribers exist
}

button.Click += (sender, e) => Console.WriteLine("clicked");   // subscribe with +=
button.Click -= handler;                                       // unsubscribe with -=
```
Use the predefined `EventHandler` / `EventHandler<TEventArgs>` delegates for events. The `event` keyword restricts callers to `+=`/`-=` (they can't overwrite or invoke the delegate). Delegates are multicast — assigning with `=` replaces all handlers, so always use `+=`. Naming convention for handlers: `Object_Event`.

## IDisposable and finalizers

Implement `IDisposable` to deterministically release unmanaged resources (files, sockets, OS handles). Callers use `using` to guarantee cleanup even on exceptions:
```csharp
using (var stream = File.CreateText(path)) { /* ... */ }   // Dispose() called automatically
using var reader = new StreamReader(path);                 // declaration form (C# 8+)
await using var conn = OpenAsync();                         // IAsyncDisposable
```
Standard dispose pattern: a public `Dispose()` that calls `Dispose(true)` then `GC.SuppressFinalize(this)`, a `protected virtual void Dispose(bool disposing)` that releases unmanaged (always) and managed (when `disposing`) resources guarded by a `disposed` flag, and an optional finalizer `~T()` calling `Dispose(false)`. Most code only *consumes* `IDisposable` types; you rarely write finalizers.
