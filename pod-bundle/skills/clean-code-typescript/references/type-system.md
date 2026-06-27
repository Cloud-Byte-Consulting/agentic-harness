# Leveraging the Type System

The type system is a design tool, not decoration. Used well, it makes illegal states unrepresentable and moves whole classes of bugs from runtime to build time. This file covers the practical toolkit and the patterns that matter for clean, maintainable code.

## Table of contents
- `type` vs `interface`
- Unions, intersections, literals
- Generics and constraints
- Narrowing and type guards
- Discriminated unions and exhaustiveness
- Mapped, conditional, and utility types
- Strict mode and tsconfig
- Modeling the domain and the runtime gate

## `type` vs `interface`

Both name object shapes; they overlap heavily. Guidance:

- **`interface`** — best for object/class shapes you may extend or implement, and for public API surfaces. Supports declaration merging (multiple declarations combine) and reads naturally with `extends`/`implements`.
- **`type`** — needed for anything that isn't a plain object shape: unions, intersections, tuples, function types, mapped/conditional types, and aliasing primitives or literals.

```typescript
interface User { name: string; age: number; }          // object shape, extendable
type Id = string | number;                              // union — must be `type`
type Point = readonly [number, number];                 // tuple — must be `type`
type Handler = (e: Event) => void;                       // function type — `type` reads better
```

A reasonable default: `interface` for object shapes, `type` for everything else. Consistency within a codebase matters more than the exact rule.

## Unions, intersections, literals

- **Union (`A | B`)** — a value is one of several types. Great for "this can be a string or a number," and the backbone of modeling variants.
- **Intersection (`A & B`)** — a value has *all* properties of several types. Use to compose shapes and avoid duplication (DRY).
- **Literal types** — exact values, not just primitives. The most under-used clean-code tool in TypeScript.

```typescript
type Person = { name: string };
type Employee = Person & { badgeNumber: number };       // intersection composes shapes

type Direction = "up" | "down" | "left" | "right";       // string-literal union
function move(d: Direction) {}                            // move("forward") is a compile error
```

Prefer a **string-literal union over a string** (and often over an `enum`) for a fixed set of options — it's lighter, fully type-checked, and erases at compile time. (Numeric `enum`s exist and are fine for interop, but a literal union or `as const` object is usually cleaner and avoids `enum`'s runtime surprises.)

## Generics and constraints

Generics let one implementation work across types while preserving the relationship between input and output — reuse **without** sacrificing type safety. Reach for them when the same logic applies to many types (containers, utilities, data structures).

```typescript
function wrapInArray<T>(value: T): T[] { return [value]; } // return type tied to input type
const nums = wrapInArray(42);   // number[]
const strs = wrapInArray("hi"); // string[]

class Stack<T> {
  private items: T[] = [];
  push(item: T) { this.items.push(item); }
  pop(): T | undefined { return this.items.pop(); }
}
```

**Constrain generics** with `extends` when you need certain properties — this keeps the generic flexible but safe:

```typescript
function logLength<T extends { length: number }>(item: T): void {
  console.log(item.length); // safe: T is guaranteed to have .length
}
logLength("hello"); logLength([1, 2, 3]);
// logLength(42); // Error: number has no 'length'
```

Don't over-genericize. A generic that's only ever used with one type is just noise (YAGNI) — write the concrete type.

## Narrowing and type guards

TypeScript narrows a broad type to a specific one inside a checked branch. Use built-in guards and custom ones to safely work with unions and `unknown`.

```typescript
function format(value: string | number): string {
  if (typeof value === "string") return value.toUpperCase(); // narrowed to string
  return value.toFixed(2);                                    // narrowed to number
}
```

Narrowing tools: `typeof` (primitives), `instanceof` (classes), the `in` operator (property presence), truthiness checks, and equality. For complex shapes, write a **user-defined type guard** returning `value is T`:

```typescript
function isUser(value: unknown): value is User {
  return typeof value === "object" && value !== null
    && "name" in value && typeof (value as any).name === "string";
}
```

Use **optional chaining** (`user?.name`) and **nullish coalescing** (`name ?? "guest"`) for nullable access instead of manual `&&` ladders.

## Discriminated unions and exhaustiveness

This is the highest-leverage clean-code pattern in TypeScript. Instead of one fuzzy object with many optional fields ("maybe everything exists"), model the **valid variants explicitly** with a shared discriminant field. Then a `switch` narrows automatically, and a `never` check enforces that you handle every case.

```typescript
type Citation = { sourceTitle: string; url: string };

type ChatResponse =
  | { kind: "answer"; reply: string }
  | { kind: "answer-with-citations"; reply: string; citations: Citation[] };

function render(response: ChatResponse): string {
  switch (response.kind) {
    case "answer":
      return response.reply;
    case "answer-with-citations":
      return `${response.reply} (${response.citations.length} sources)`;
    default: {
      const _exhaustive: never = response; // compile error if a variant is unhandled
      return _exhaustive;
    }
  }
}
```

If someone later adds `{ kind: "rate-limited" }` and forgets to handle it, the `never` assignment fails to compile — correctness is enforced structurally. This is "type safety under change": entire categories of bugs become hard to ship.

**Avoid optional-field soup.** Many `?:` fields encode invalid combinations the compiler can't catch. A discriminated union says "only these states exist," which is far stronger.

## Mapped, conditional, and utility types

**Mapped types** transform every property of a type — the DRY way to derive related types:

```typescript
type ReadonlyAll<T> = { readonly [K in keyof T]: T[K] };
type Optional<T>    = { [K in keyof T]?: T[K] };
```

**Conditional types** branch on a type relationship (`T extends U ? X : Y`), often with `infer` to extract a piece:

```typescript
type ElementType<T> = T extends (infer U)[] ? U : T;
type A = ElementType<string[]>; // string
```

**Built-in utility types** cover most everyday needs — use them instead of rolling your own:

| Utility | Effect |
|---------|--------|
| `Partial<T>` | all properties optional |
| `Required<T>` | all properties required |
| `Readonly<T>` | all properties read-only |
| `Pick<T, K>` | keep only keys `K` |
| `Omit<T, K>` | drop keys `K` |
| `Record<K, V>` | object with keys `K`, values `V` |
| `Exclude<T, U>` / `Extract<T, U>` | remove / keep union members |
| `NonNullable<T>` | strip `null`/`undefined` |
| `ReturnType<T>` / `Parameters<T>` | a function's return / parameter types |

```typescript
type CreateUser = Omit<User, "id">;            // input shape derived from User
type UserSummary = Pick<User, "name" | "email">;
type Roles = "admin" | "user";
type Perms = Record<Roles, boolean>;           // { admin: boolean; user: boolean }
```

Deriving types from a single source (rather than hand-maintaining parallel shapes) keeps them in sync automatically. Keep type-level programming readable — deeply nested conditionals and excessive `infer` become as hard to maintain as clever runtime code (KISS applies to types too).

## Strict mode and tsconfig

**Turn on `strict`.** It's the foundation of getting value from TypeScript — it enables the whole family of strict checks. The most important members:

- `strictNullChecks` — `null`/`undefined` are not silently assignable; you must handle them. Eliminates a huge swath of runtime "cannot read property of undefined" errors.
- `noImplicitAny` — forbids accidental `any`; forces you to type (or deliberately widen) values.

```json
// tsconfig.json — sensible modern baseline
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "sourceMap": true,
    "outDir": "./dist"
  }
}
```

Other useful flags: `sourceMap` (debug the original TS, not compiled JS), `noUnusedLocals`/`noUnusedParameters` (catch dead code), `declaration` (emit `.d.ts` for libraries). Use `extends` to share a base config across packages in a monorepo.

> A skill-aware note on `enum`: `preserveConstEnums` and numeric enums have runtime quirks; modern style often prefers `as const` objects + literal unions over enums for fully type-safe, tree-shakeable constants.

## Modeling the domain and the runtime gate

Two principles tie the type system to clean architecture:

**1. Types describe your *domain*, not a vendor's wire format.** If your shared types mirror an external API's raw response, every upstream change becomes a change throughout your app. Define types for what *your product* needs; adapt vendor responses into your domain shape behind an interface (Strategy/Adapter). This isolates volatility and keeps boundaries stable.

```typescript
// Shared domain type — what the app actually uses.
interface ChatReply { message: string; references: string[] }

// Each provider adapts its own SDK into the domain type behind one interface.
interface LlmProvider { generateReply(prompt: string): Promise<ChatReply>; }
// Swapping providers never touches the rest of the app.
```

**2. Types don't validate runtime data — guard the boundary.** Everything crossing the network (`req.body`, `res.json()`, env vars, user input) is `unknown` until proven otherwise. A passing build over an `as SomeType` cast or `any` is a *false* sense of safety — the compiler had nothing to check. Treat boundary data as `unknown` and pass it through a type guard before trusting it; fail fast with a clear error if it doesn't conform.

```typescript
function isChatReply(v: unknown): v is ChatReply {
  return typeof v === "object" && v !== null
    && typeof (v as any).message === "string"
    && Array.isArray((v as any).references);
}

async function fetchReply(prompt: string): Promise<ChatReply> {
  const res = await fetch("/api/chat", { method: "POST", body: JSON.stringify({ prompt }) });
  const data: unknown = await res.json();          // unknown, not trusted
  if (!isChatReply(data)) throw new Error("Invalid ChatReply shape from server");
  return data;                                     // now safely typed
}
```

Inside the gate, lean fully on the types. At the gate, never lie to the compiler with `as` casts on untrusted data — that's where a "silent type lie" turns into a production bug.
