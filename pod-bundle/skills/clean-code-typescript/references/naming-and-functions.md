# Naming, Functions, and Immutability

How to name things so code explains itself, design small focused functions, manage parameters and side effects, and keep data immutable. Lead with the rule, then the why.

## Naming conventions

| Kind | Convention | Examples |
|------|-----------|----------|
| Variables, functions, methods | camelCase | `userName`, `calculateArea`, `fetchUser` |
| Types, interfaces, classes, enums | PascalCase | `ShoppingCart`, `PaymentMethod`, `Role` |
| Constants (true compile-time consts) | camelCase or UPPER_SNAKE for module-level config | `maxRetries`, `API_BASE_URL` |
| Type parameters | single uppercase, or descriptive PascalCase | `T`, `TKey`, `TValue` |
| Private fields | `private` keyword (preferred) or `#name` for hard-private | `private items`, `#secret` |

TypeScript does **not** need Hungarian/type-prefix names (`strName`, `IUser`) — the type system already tells you the type. The `I`-prefix on interfaces is a legacy convention; modern TS style drops it (`User`, not `IUser`).

## Naming heuristics

- **Be descriptive, not clever.** `getUserData()` beats `doStuff()`. `elapsedTimeInDays` beats `d`. The name should let a reader skip the implementation.
- **Functions are verbs / verb phrases.** `renderButton`, `applyDiscount`, `parseConfig`.
- **Booleans read as assertions.** Prefix with `is`/`has`/`can`/`should`: `isLoggedIn`, `hasPermission`, `canEdit`, `shouldRetry`. Then `if (user.isActive)` reads like English.
- **Accessors use `get`/`set`.** `getProductDetails()`, `setActiveUser()`.
- **Avoid noise words.** `data`, `info`, `manager`, `value`, `obj`, `temp` carry no meaning. `userRecord` is rarely better than `user`.
- **No magic numbers/strings.** Name them: `const MAX_LOGIN_ATTEMPTS = 5;` or a union/enum for a fixed set.
- **Be consistent.** Pick one term per concept (`fetch` vs `get` vs `retrieve`) and use it everywhere. Inconsistent vocabulary forces readers to wonder whether the difference is meaningful.
- **Length tracks scope.** A loop index `i` is fine; a module-level export deserves a full, precise name. Don't pad short-lived locals; don't abbreviate long-lived public names.

**The Stranger Test:** if someone who couldn't code sat at your screen, could they roughly guess what the file does? `class ShoppingCart` passes; `class Cart` (cart of what?) is weaker; `class Mgr` fails.

## Small, focused functions

A clean function does **one thing** and has **one reason to change** (the Single Responsibility Principle applied at function level). Benefits: readability, lower complexity, reusability, and testability — small functions are trivial to isolate and test.

How to find the seams: write the function's purpose in one sentence. If it does anything beyond that sentence, extract it.

```typescript
type CartItem = { price: number; quantity: number };

// Smell: one function calculates, discounts, AND prints.
function checkout(cart: CartItem[], discount: number): void {
  let total = 0;
  for (const item of cart) total += item.price * item.quantity;
  total = total * (1 - discount / 100);
  console.log(`Your total is: $${total.toFixed(2)}`);
}

// Clean: each function has a single responsibility.
function calculateTotal(cart: CartItem[]): number {
  return cart.reduce((sum, item) => sum + item.price * item.quantity, 0);
}

function applyDiscount(total: number, discountPercent: number): number {
  return total * (1 - discountPercent / 100);
}

function formatReceipt(total: number): string {
  return `Your total is: $${total.toFixed(2)}`;
}

function checkout(cart: CartItem[], discount: number): string {
  const total = applyDiscount(calculateTotal(cart), discount);
  return formatReceipt(total); // printing happens at the edge, by the caller
}
```

Now each piece is independently testable, and `checkout` reads as a summary of its steps.

## Parameters

- **Aim for ≤3 parameters.** More than that usually signals the function does too much, or that related arguments belong together.
- **Group related parameters into an object type.** This also makes call sites self-documenting and order-independent.

```typescript
// Smell: positional, easy to mix up, boolean flags are opaque at the call site.
function createUser(name: string, email: string, admin: boolean, verified: boolean) { /* ... */ }
createUser("Ada", "ada@x.com", true, false); // which flag is which?

// Clean: a parameter object. Call site is self-explanatory.
type CreateUserOptions = {
  name: string;
  email: string;
  isAdmin?: boolean;
  isVerified?: boolean;
};
function createUser(options: CreateUserOptions) { /* ... */ }
createUser({ name: "Ada", email: "ada@x.com", isAdmin: true });
```

- **Avoid boolean flag parameters** — a function that takes a `boolean` to choose between two behaviors is really two functions. Split it, or pass a named option.
- **Use optional (`?`) and default parameters** instead of overloads or sentinel values where natural: `function greet(name: string, greeting = "Hello") {}`.

## Function signatures and return types

A signature is a contract: parameter types, return type, and (by documentation) what it may throw. Annotate return types on exported/public functions — it documents intent and catches accidental drift, even though TS can infer them.

```typescript
function calculateCircleArea(radius: number): number {
  return Math.PI * radius * radius;
}
```

Return explicit, well-typed values. Prefer returning data over mutating a parameter. If a function can fail, make that visible (throw a typed error, or return a result/union) rather than returning `undefined` silently.

## Pure functions and side effects

A **side effect** is any change a function makes outside its own scope: mutating a global or argument, writing a file, hitting the network, logging, updating the DOM. Side effects make code unpredictable and hard to test, and cause race conditions under concurrency.

A **pure function** depends only on its inputs and returns only a value — same input, same output, no observable effect. Pure functions are reliable, trivially testable, and safe under concurrency.

Strategy:
- Keep the core logic pure; push side effects (I/O, mutation, console, DOM) to the edges/boundary of the system.
- Don't read or write module-level mutable state from inside reusable functions; pass what you need as arguments and return results.

```typescript
// Smell: depends on and mutates shared global state — output varies with timing.
let currentUser = "Alice";
async function fetchUserData(userId: string): Promise<void> {
  const res = await fetch(`/api/users/${userId}`);
  currentUser = (await res.json()).name; // hidden side effect
}

// Clean: pure-ish — returns the result; caller decides what to do with it.
async function fetchUserName(userId: string): Promise<string> {
  const res = await fetch(`/api/users/${userId}`);
  const data = (await res.json()) as { name: string };
  return data.name;
}
```

## Immutability

Treat data as immutable by default; create new values instead of mutating existing ones. This prevents action-at-a-distance bugs where one part of the app changes data another part relied on.

Tools TypeScript gives you:
- `readonly` on properties and `ReadonlyArray<T>` / `readonly T[]` for arrays.
- `as const` to freeze a literal into its narrowest, immutable type.
- `Readonly<T>` utility type to make all properties read-only.
- Non-mutating operations: spread (`{ ...obj, x: 1 }`, `[...arr, item]`), `map`/`filter`/`reduce` instead of in-place `push`/`splice` where practical, `toSorted`/`toReversed` (ES2023) instead of `sort`/`reverse` which mutate.

```typescript
type Point = { readonly x: number; readonly y: number };
const p: Point = { x: 1, y: 2 };
// p.x = 5; // Error: Cannot assign to 'x' because it is a read-only property.
const moved = { ...p, x: 5 }; // new object, original untouched

const directions = ["up", "down"] as const; // readonly ["up", "down"], type is the literals
```

## Documenting functions with TSDoc

Modern TypeScript uses **TSDoc** (the `/** ... */` comment standard), and **TypeDoc** generates HTML docs from it. Document the *contract and intent*, especially non-obvious parameters — but remember: good names and types reduce how much you need to document. Document *why*, and clarify ambiguous units/meaning; don't restate the signature.

```typescript
/**
 * Applies a percentage discount to a total.
 * @param total - The pre-discount total, in the cart's currency.
 * @param discountPercent - The discount as a percentage (e.g. 10 for 10% off).
 * @returns The discounted total.
 */
function applyDiscount(total: number, discountPercent: number): number {
  return total * (1 - discountPercent / 100);
}
```

Common TSDoc tags: `@param`, `@returns`, `@remarks` (design notes), `@deprecated` (with the replacement), `@example`, `@link`. The payoff is editor hover tooltips and generated docs that make onboarding and reuse easier.
