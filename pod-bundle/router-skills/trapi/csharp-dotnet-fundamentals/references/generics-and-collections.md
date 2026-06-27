# Generics and Collections

Type-safe reuse with generics, plus choosing and using the right collection. Includes spans, indexes, and ranges.

## Contents
- Generics
- Generic constraints
- Arrays
- Picking a collection
- List
- Dictionary
- HashSet (sets)
- Stack and Queue
- Sorted, immutable, concurrent, specialized collections
- Read-only views
- Spans, indexes, and ranges

## Generics

Generics let a type or method work with a type parameter supplied by the caller — type-safe, no boxing, no casts.

```csharp
Dictionary<int, string> lookup = new();   // compiler enforces int keys, string values
lookup.Add(1, "Alpha");
// lookup.Add(person, "x");   // compile error — wrong key type

T Echo<T>(T value) => value;              // generic method; T inferred from the argument
int n = Echo(42);                          // T = int
```
Naming: a single parameter is `T` (`List<T>`); multiple use `T`-prefixed names (`Dictionary<TKey, TValue>`). Always prefer generic collections over the legacy non-generic `System.Collections` types (`ArrayList`, `Hashtable`) which box value types and are weakly typed.

## Generic constraints

Restrict what `T` can be with `where`:
```csharp
T Max<T>(T a, T b) where T : IComparable<T>          // T must be comparable
    => a.CompareTo(b) >= 0 ? a : b;

T New<T>() where T : new() => new T();                // T has a public parameterless ctor
void Use<T>(T x) where T : class { }                  // reference type
void Use2<T>(T x) where T : struct { }                // non-nullable value type
void Use3<T>(T x) where T : notnull { }               // non-nullable
void Use4<T>(T x) where T : Person { }                // T is Person or derived
```
Variance on interfaces/delegates: `in T` (contravariant — accepts a less-derived type), `out T` (covariant — returns a more-derived type), e.g. `IComparable<in T>`, `IEnumerable<out T>`.

## Arrays

Fixed size, contiguous, fastest indexed access — use when the count won't change.
```csharp
string[] names = new string[4];                 // allocate
names[0] = "Kate";
string[] init = { "Kate", "Jack", "Rebecca" };  // array initializer
int len = init.Length;                          // use Length, not Count()
string[,] grid = new string[3, 4];              // rectangular multi-dimensional
string[][] jagged = { new[] { "a" }, new[] { "b", "c" } }; // array of arrays
```
Indexes are zero-based. Use `Array.Empty<int>()` for an empty array (no allocation). `Array.Sort(arr)` sorts in place (elements must implement `IComparable`, or pass an `IComparer`).

## Picking a collection

| Need | Type | Namespace |
|---|---|---|
| Ordered, indexable, growable list | `List<T>` | `System.Collections.Generic` |
| Key→value lookup | `Dictionary<TKey,TValue>` | same |
| Unique items / set algebra | `HashSet<T>` | same |
| LIFO (undo stack) | `Stack<T>` | same |
| FIFO (work queue) | `Queue<T>` | same |
| Prioritized FIFO | `PriorityQueue<TElement,TPriority>` (.NET 6+) | same |
| Always-sorted | `SortedDictionary`, `SortedList`, `SortedSet` | same |
| Never changes / snapshot | `ImmutableArray`, `ImmutableList`, ... | `System.Collections.Immutable` |
| Thread-safe shared access | `ConcurrentDictionary`, `ConcurrentQueue`, ... | `System.Collections.Concurrent` |

All collections implement `ICollection<T>` (`Count`, `Add`, `Clear`, `Contains`, `Remove`) and `IEnumerable<T>` (iterable with `foreach`). `IList<T>` adds indexing (`this[int]`, `Insert`, `RemoveAt`, `IndexOf`).

## List

```csharp
List<string> cities = new() { "London", "Paris", "Milan" };
cities.Add("Rome");
cities.AddRange(new[] { "Oslo", "Bonn" });
cities.Insert(0, "Sydney");      // shifts later items; their indexes change
cities.RemoveAt(1);
cities.Remove("Milan");
string first = cities[0];
int count = cities.Count;        // use Count property, not LINQ Count()
cities.Sort();                   // in place; custom types need IComparable or an IComparer
```
A `List<T>` is ideal while adding/removing; convert to an array (`ToArray()`) once stable if you want lower memory and contiguous storage.

## Dictionary

Fast key→value lookup; keys must be unique.
```csharp
Dictionary<string, string> kw = new()
{
    ["int"] = "32-bit integer",
    ["long"] = "64-bit integer",
};
kw.Add("float", "single precision");          // throws if key exists
if (kw.TryGetValue("long", out string? def))  // safe lookup, no exception
    Console.WriteLine(def);
bool has = kw.ContainsKey("int");
foreach (KeyValuePair<string, string> kvp in kw)
    Console.WriteLine($"{kvp.Key}: {kvp.Value}");
```
Items are `KeyValuePair<TKey,TValue>` (a value type). Indexer `dict[key]` throws `KeyNotFoundException` if missing — prefer `TryGetValue`. Most dictionaries in practice are built from data via LINQ `ToDictionary`/`ToLookup` (see `linq.md`).

## HashSet (sets)

Unique items, fast membership tests, set algebra.
```csharp
HashSet<string> names = new();
bool added = names.Add("Adam");      // false if already present
names.UnionWith(other);              // add all from other
names.IntersectWith(other);          // keep only those also in other
names.ExceptWith(other);             // remove those in other
bool subset = names.IsSubsetOf(other);
```

## Stack and Queue

```csharp
Stack<string> undo = new();
undo.Push("a"); string last = undo.Pop();  string top = undo.Peek();  // LIFO

Queue<string> work = new();
work.Enqueue("job1"); string next = work.Dequeue(); string front = work.Peek(); // FIFO

PriorityQueue<string, int> pq = new();      // lower priority value dequeues first
pq.Enqueue("Pamela", 1);
pq.Enqueue("Rebecca", 3);
string served = pq.Dequeue();                // "Pamela"
```
Stacks/queues aren't sortable and don't expose an index — that's intentional.

## Sorted, immutable, concurrent, specialized collections

- **Auto-sorting**: `SortedDictionary<TKey,TValue>` (binary tree), `SortedList<TKey,TValue>` (sorted array — less memory, slower inserts on unsorted data), `SortedSet<T>`.
- **Immutable** (`System.Collections.Immutable`): operations return a new collection; the original never changes — safe to share across threads.
- **Concurrent** (`System.Collections.Concurrent`): `ConcurrentDictionary`, `ConcurrentQueue`, `BlockingCollection` — safe for multi-threaded producers/consumers.
- **Specialized**: `LinkedList<T>` (doubly linked — fast mid-list insert/remove), `BitArray` (compact bit flags).

## Read-only views

Pass a collection without allowing mutation by accepting `IReadOnlyList<T>`, `IReadOnlyCollection<T>`, or `IReadOnlyDictionary<TKey,TValue>`. `ICollection<T>.IsReadOnly` reports whether a wrapper forbids changes. .NET 8 adds **frozen** collections (`FrozenDictionary`, `FrozenSet`) optimized for create-once-read-many.

## Spans, indexes, and ranges

`Index` and `Range` give concise slicing of arrays/strings/spans:
```csharp
int[] a = { 10, 20, 30, 40, 50 };
int last = a[^1];          // ^1 = last element (from the end) → 50
int[] mid = a[1..4];       // range [1,4)  → 20,30,40
int[] tail = a[2..];       // from index 2 to end
int[] head = a[..3];       // start to index 3 (exclusive)
```
`Span<T>`/`ReadOnlySpan<T>` are stack-only views over contiguous memory (arrays, strings, stack-allocated buffers) that allow slicing without copying — high-performance, low-allocation. Many BCL APIs (`int.TryParse`, `Regex.IsMatch`) accept `ReadOnlySpan<char>`. Use spans for hot-path parsing/processing where avoiding allocations matters.
