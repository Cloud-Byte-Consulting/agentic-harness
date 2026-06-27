# Error Handling and Debugging

Robust, clean error handling keeps applications stable and user-friendly. This file covers the categories of errors, synchronous and asynchronous handling, the `unknown` catch type, secure/graceful practices, and debugging tools.

## Categories of errors

Knowing the kind of error tells you where and how to catch it:

- **Syntax errors** — violate the language grammar (missing `)`, invalid name). The compiler catches them; code won't build. Read the `tsc` message; the editor underlines the spot.
- **Type errors** — a value used incompatibly with its type (assigning `string` to `number`, calling `.push()` on a string, accessing a property on a possibly-`null` value). TypeScript catches these at compile time — its core value. Fix by correcting the type, adding a guard, or enabling `strict`.
- **Runtime errors** — code compiles and passes type checks but fails while running. Common causes: trusting external data, `any`/forced casts that bypass safety, operating on unexpected values. These are the ones error handling and validation exist to manage.
- **Logical errors** — code runs without crashing but produces the wrong result (off-by-one, wrong operator precedence, `=` instead of `===`). No error message — caught only by testing, especially boundary/edge cases.

The general principle of TypeScript is to **shift failures earlier** — from production runtime to build time, where they're cheaper to fix. `strict` mode, precise types, and discriminated unions do most of that work (see `references/type-system.md`).

## Synchronous error handling

Synchronous code runs top-to-bottom; an unhandled throw stops execution and can crash the program or leave the UI in a broken state.

### try/catch

Wrap the risky operation; recover or report in `catch`. Keep the happy path in `try`, handle failure in `catch`:

```typescript
function handleCalculation(input: string): string {
  try {
    return `Result: ${calculateSquare(input)}`;
  } catch (error: unknown) {
    return error instanceof Error ? error.message : "An unexpected error occurred.";
  }
}
```

### Validate input early (fail fast)

The best error handling prevents the error. Validate at the boundary before the value propagates deeper, and exit immediately on bad input:

```typescript
function getUserData(id: number): User | null {
  if (typeof id !== "number" || id <= 0) return null; // reject early, before any work
  // proceed knowing id is valid
}
```

Validating early improves UX (instant feedback) and is a security measure — unchecked input is the root of injection and XSS. For complex/structured input, use a schema validation library (Zod, Valibot, Joi) rather than hand-rolled checks; they give you both runtime validation and inferred static types.

## Asynchronous error handling

Async operations (network, file I/O, timers) complete later, so their errors surface outside the original call's stack — they need explicit handling or they fail silently.

### Promises: `.then().catch().finally()`

```typescript
function fetchData(url: string): Promise<void> {
  return fetch(url)
    .then(res => {
      if (!res.ok) throw new Error("Network response was not ok");
      return res.json();
    })
    .then(data => console.log(data))
    .catch(error => console.error("Failed:", error))   // catches anywhere in the chain
    .finally(() => console.log("done"));               // always runs, success or failure
}
```

### async/await with try/catch (preferred for readability)

`async/await` lets asynchronous code read like synchronous code and uses the same `try/catch` you already know:

```typescript
async function fetchData(url: string): Promise<void> {
  try {
    const res = await fetch(url);
    if (!res.ok) throw new Error("Network response was not ok");
    const data = await res.json();
    console.log(data);
  } catch (error: unknown) {
    if (error instanceof Error) console.error("Failed:", error.message);
  } finally {
    console.log("done");
  }
}
```

**Never leave a promise unhandled.** Always `await` inside a `try`, or attach `.catch`. A floating promise (called but not awaited or caught) swallows errors and causes hard-to-trace bugs and unhandled-rejection warnings.

## The `unknown` catch type

In modern TypeScript, the variable in `catch (error)` is typed `unknown`, not `any`. This is correct and deliberate: **JavaScript can throw anything** — a string, a number, a plain object, not just an `Error`. Third-party libraries, browser APIs, and database drivers don't all throw proper `Error` instances.

Because the thrown value could be anything, you must narrow before accessing properties like `.message`:

```typescript
try {
  doSomething();
} catch (error: unknown) {
  if (error instanceof Error) {
    console.error("Error message:", error.message); // safe after narrowing
  } else {
    console.error("Unexpected thrown value:", error);
  }
}
```

This forces defensive handling at system boundaries. Don't reflexively cast `error as Error` — it reintroduces the unsafety `unknown` is protecting you from.

## Clean and secure error-handling practices

- **Graceful degradation** — when something fails, keep the app usable and show the user a clear, friendly message rather than crashing or freezing.
- **Generic messages to users, details to logs** — never surface stack traces, SQL, or internal paths to end users (they leak information to attackers). Show "Something went wrong, please try again," and log the full detail to a secure sink.
- **Centralized logging** — route errors to one place (and in production, a monitoring service such as Sentry or Datadog) for observability.
- **Don't log secrets** — keep passwords, tokens, and PII out of logs.
- **Timeouts and retries** — for network calls, add timeouts and bounded retry/backoff so transient failures don't hang the app.

```typescript
try {
  await processPayment(order);
} catch (error: unknown) {
  showToast("Sorry, something went wrong. Please try again later."); // safe for the user
  logError(error);                                                    // full detail, secure sink
}
```

## Debugging tools

When an error does reach runtime, these tools find it fast:

- **Source maps** — set `"sourceMap": true` in `tsconfig.json` so the debugger maps compiled JS back to your original TypeScript. You then debug the code you actually wrote.
- **Editor debugger (e.g. VS Code)** — set a `launch.json`, add **breakpoints** (click the gutter), and step through: step over (next line), step into (enter a call), step out (leave the function), continue. Inspect variables in the Variables pane and watch expressions. Use **conditional breakpoints** (pause only when `count > 10`) and **logpoints** (log a message without pausing) to debug without editing code.
- **Browser DevTools** — for frontend code: inspect runtime state, network, and console; the Performance and Memory tabs find slow functions and leaks.
- **`console` methods** — `console.log` is fine for quick tracing; `console.error`/`console.warn`/`console.table` add severity and structure. Remove or gate debug logs before production — they clutter output and can leak data.

Use the type system as your first line of defense (catch at build time), validation as the second (guard the boundary), error handling as the third (recover gracefully), and the debugger as the tool when something still slips through.
