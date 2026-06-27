# Async and await

Asynchronous programming with `async`/`await`, `Task`, and `ValueTask` for responsive, scalable code that doesn't block threads on I/O.

## Contents
- The model
- Writing an async method
- Returning values
- Async Main
- Cancellation
- Running work in parallel
- ValueTask
- Deadlock avoidance and anti-patterns

## The model

`await`ing an operation frees the current thread while waiting (for I/O, a timer, a network call); when the awaited operation completes, execution resumes. This improves UI responsiveness and server scalability — fewer threads handle more concurrent work. An `async` method returns a `Task` (no result), `Task<T>` (a result), or `ValueTask`/`ValueTask<T>`.

## Writing an async method

```csharp
public async Task<string> FetchAsync(string url)
{
    using HttpClient client = new();
    HttpResponseMessage response = await client.GetAsync(url); // yields the thread
    return await response.Content.ReadAsStringAsync();
}
```
Rules:
- `await` is only legal inside a method marked `async`.
- Name async methods with the `Async` suffix by convention.
- An `async` method's return type is `Task`, `Task<T>`, `ValueTask`, `ValueTask<T>`, or `IAsyncEnumerable<T>` (for `await foreach`). **Avoid `async void`** except for event handlers — its exceptions can't be awaited or caught and crash the process.
- `await` unwraps the result and re-throws any exception that occurred, so wrap awaits in `try`/`catch` as usual.

## Returning values

```csharp
public async Task<int> CountLinesAsync(string path)
{
    string[] lines = await File.ReadAllLinesAsync(path);
    return lines.Length;                 // returned as Task<int>
}

int n = await CountLinesAsync("log.txt"); // caller awaits to get the int
```

## Async Main

Modern console templates generate an async `<Main>$`, so top-level statements can use `await` directly:
```csharp
// Program.cs (top-level statements)
HttpResponseMessage r = await new HttpClient().GetAsync("https://example.com");
Console.WriteLine(r.StatusCode);
```
If you write an explicit entry point, declare it `static async Task Main(string[] args)`.

## Cancellation

Accept a `CancellationToken` and pass it down; cancel via `CancellationTokenSource`.
```csharp
public async Task ProcessAsync(CancellationToken ct = default)
{
    for (int i = 0; i < 100; i++)
    {
        ct.ThrowIfCancellationRequested();      // observe cancellation
        await Task.Delay(100, ct);              // also cancellable
    }
}

using CancellationTokenSource cts = new(TimeSpan.FromSeconds(5)); // auto-cancel after 5s
await ProcessAsync(cts.Token);
```

## Running work in parallel

```csharp
// Start tasks, then await them together:
Task<string> a = FetchAsync(url1);
Task<string> b = FetchAsync(url2);
string[] results = await Task.WhenAll(a, b);     // wait for all; aggregates results

Task<string> first = await Task.WhenAny(a, b);   // the first to complete

// Offload CPU-bound work to the thread pool:
int sum = await Task.Run(() => HeavyCompute(data));
```
`Task.WhenAll` collects results/exceptions; `Task.WhenAny` returns the first finished task. Use `Task.Run` for CPU-bound work (don't use `Task.Run` to make synchronous I/O "async" — call the I/O's own async API). For CPU-bound data parallelism, consider `Parallel.ForEach`/PLINQ instead.

## ValueTask

`ValueTask<T>` avoids allocating a `Task` object when a method often completes synchronously (e.g. a cache hit). Use it in hot paths; otherwise prefer plain `Task<T>` for simplicity. Don't await a `ValueTask` more than once and don't store it — convert with `.AsTask()` if you need to.

## Deadlock avoidance and anti-patterns

- **Never block on async code** with `.Result` or `.Wait()` from a context with a synchronization context (classic UI/ASP.NET) — it can deadlock. `await` all the way up.
- **`async void`** swallows exceptions — use `async Task`. Only event handlers should be `async void`.
- **Forgetting to await** a `Task` (fire-and-forget) drops its exceptions silently. Await it, or deliberately handle the task.
- In library code, consider `await task.ConfigureAwait(false)` to avoid capturing the caller's context (improves performance and avoids some deadlocks); app-level code usually doesn't need it.
- Remember LINQ's deferred execution interacts with tasks: a `Select(_ => Task.Run(...))` builds tasks lazily, so re-enumerating (e.g. via `Count()`) can start them again. Materialize the task sequence once (`ToArray()`) before awaiting.
