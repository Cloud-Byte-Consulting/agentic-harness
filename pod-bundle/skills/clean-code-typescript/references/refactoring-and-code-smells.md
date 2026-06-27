# DRY/KISS/YAGNI, Code Smells, and Refactoring

The principles that govern *how much* structure to add, a catalog of code smells with TypeScript fixes, the refactoring techniques that fix them, and a safe workflow for applying them.

## Table of contents
- DRY, KISS, YAGNI
- Code smell catalog (with fixes)
- Refactoring techniques
- A safe refactoring workflow

## DRY, KISS, YAGNI

These three principles pull in tension; clean code balances them.

### DRY — Don't Repeat Yourself

Every piece of *knowledge* should have one authoritative representation. Duplicated logic means a change has to be made in N places, and someone will miss one. Remove genuine duplication by extracting a function, type, constant, or module.

```typescript
// Smell: the discount formula lives in two places — change one, forget the other.
const cartTotal = subtotal * (1 - userDiscount / 100);
const previewTotal = preview * (1 - userDiscount / 100);

// DRY: one source of truth.
const applyDiscount = (amount: number, pct: number) => amount * (1 - pct / 100);
const cartTotal = applyDiscount(subtotal, userDiscount);
const previewTotal = applyDiscount(preview, userDiscount);
```

At the type level, DRY means deriving types from one source (`Omit`, `Pick`, mapped types) instead of maintaining parallel shapes — see `references/type-system.md`.

**The caveat:** DRY is about duplicated *knowledge*, not duplicated *text*. Two pieces of code that look alike today but change for different reasons are *coincidental* duplication — merging them couples unrelated things and creates a worse mess later. Don't abstract until the duplication is real and stable. Premature DRY is its own smell.

### KISS — Keep It Simple

Prefer the simplest design that solves the problem. Complexity is a cost: clever one-liners, deep abstraction layers, and elaborate type gymnastics make code harder to read, debug, and change. A plain function often beats a design pattern. Readability and maintainability usually matter more than micro-optimizations — choose the clear approach unless profiling proves you need the fast one.

```typescript
// Over-engineered for a fixed config: a Singleton + factory for one object.
// KISS: a plain exported const is enough.
export const config = { apiUrl: "/api", retries: 3 } as const;
```

### YAGNI — You Aren't Gonna Need It

Don't build for imagined future requirements. Speculative generality — extra parameters "just in case," abstractions with one implementation, flags nobody uses — adds complexity that you pay for now and may never benefit from. Build what the current requirement needs; add structure when a real second case arrives.

A generic with a single concrete use, an interface with one implementer that will never have another, or a config option that's always the same value are all YAGNI violations. Introduce abstraction when the second case shows up, not before.

**Together:** YAGNI and KISS keep you from over-building; DRY keeps you from under-building (copy-paste). The skill is knowing which pressure applies. When unsure, prefer simple and concrete — it's cheaper to extract an abstraction later than to unwind a wrong one.

## Code smell catalog

A *code smell* is a surface symptom that hints at a deeper design problem. Each below pairs the smell with its idiomatic TypeScript fix.

| Smell | What it looks like | Fix |
|-------|-------------------|-----|
| **Long function** | A function spanning many responsibilities/screens | Extract Function; split by responsibility (SRP) |
| **Long parameter list** | 4+ positional params; booleans you can't tell apart | Introduce Parameter Object; split the function |
| **Boolean flag argument** | `render(true, false)` | Split into named functions or pass an options object |
| **Deep nesting / arrow code** | Nested `if`s drifting rightward | Guard clauses + early return; extract helpers |
| **God class / large class** | One class doing many unrelated things | Split by responsibility (SRP); extract collaborators |
| **Primitive obsession** | Raw `string`/`number` for domain concepts | Replace with literal union, branded type, or small type |
| **Optional-field soup** | Many `?:` fields encoding invalid combinations | Discriminated union; make illegal states unrepresentable |
| **`any` at boundaries** | `any`/`as` on external data | `unknown` + type guard at the boundary |
| **Feature envy** | A method uses another object's data more than its own | Move the method to the object that owns the data |
| **Shotgun surgery** | One change forces edits across many files | Consolidate the responsibility into one module |
| **Duplicated code** | Same logic in several places | Extract Function/Type/Module (DRY) — if truly the same knowledge |
| **Magic numbers/strings** | Unexplained literals | Name them; use a const or literal union |
| **Dead code / commented-out code** | Unreachable code, `//`-ed blocks | Delete it; git remembers |
| **Comment explaining *what*** | Narrating obvious code | Rename/extract so the code is self-explanatory |
| **Mutating shared state** | Reassigning args/globals; in-place mutation | Return new values; `readonly`, spread, `as const` |
| **Switch/if-else on a type tag, repeated** | The same conditional in many spots | Replace Conditional with Polymorphism/Strategy (OCP) |
| **Floating promises** | Async calls not awaited/caught | `await` in `try/catch` or attach `.catch` |
| **Over-abstraction** | Layers/generics with one user | Inline it (YAGNI/KISS) |

### Worked examples

**Deep nesting → guard clauses:**

```typescript
// Smell
function getPrice(user?: User): number {
  if (user) {
    if (user.isActive) {
      if (user.subscription) {
        return user.subscription.price;
      }
    }
  }
  return 0;
}

// Fixed: invert conditions, return early, flatten.
function getPrice(user?: User): number {
  if (!user?.isActive) return 0;
  if (!user.subscription) return 0;
  return user.subscription.price;
}
```

**Primitive obsession → literal union:**

```typescript
// Smell: any string passes; typos compile.
function setRole(role: string) {}
setRole("amin"); // oops, no error

// Fixed: only valid roles compile.
type Role = "admin" | "user" | "guest";
function setRole(role: Role) {}
// setRole("amin"); // compile error
```

**Repeated conditional → Strategy (OCP):**

```typescript
// Smell: every new method edits this switch.
function pay(method: string, amount: number) {
  if (method === "paypal") { /* ... */ }
  else if (method === "card") { /* ... */ }
}

// Fixed: add a class, touch nothing existing.
interface PaymentStrategy { pay(amount: number): void; }
class PayPal implements PaymentStrategy { pay(a: number) { /* ... */ } }
class Card implements PaymentStrategy { pay(a: number) { /* ... */ } }
const checkout = (strategy: PaymentStrategy, amount: number) => strategy.pay(amount);
```

## Refactoring techniques

Refactoring changes structure **without changing behavior**. The core moves:

- **Extract Function** — pull a cohesive chunk into a well-named function. The most common, highest-value refactor; cures long functions and duplication.
- **Extract Type / Interface** — name a recurring object shape; share it (DRY).
- **Introduce Parameter Object** — bundle related parameters into one typed object.
- **Replace Primitive with Type** — swap a raw `string`/`number` for a union, literal, or branded type.
- **Replace Conditional with Polymorphism / Strategy** — turn a type-tag `switch` into interface implementations.
- **Replace Inheritance with Composition** — break a rigid hierarchy into composed parts (see `references/solid-principles.md`).
- **Inline** — remove an abstraction that no longer earns its keep (YAGNI cleanup).
- **Replace `any` with `unknown` + guard** — tighten an untrusted boundary.
- **Rename** — the cheapest, most underrated refactor; a better name often removes the need for a comment.

## A safe refactoring workflow

Refactor in small, reversible steps with a safety net:

1. **Pin behavior with a test.** If the code isn't tested, write a characterization test that captures what it does *now* before you touch it. This is your guardrail.
2. **Make one small change.** Apply a single refactoring move (extract one function, rename one thing).
3. **Re-run tests.** They must stay green. If they go red, you changed behavior — revert and try a smaller step.
4. **Tighten types as you go.** Narrow `any` → `unknown` → precise; add `readonly`; turn loose strings into unions. Let the compiler find missed call sites — that's TypeScript's refactoring superpower.
5. **Commit the small step.** Frequent commits make any misstep trivial to undo.
6. **Repeat.** Many tiny safe moves beat one big risky rewrite.

Lean on the toolchain: the type checker flags every place a changed signature ripples to; ESLint flags new smells; pre-commit hooks keep the result clean (see `references/comments-and-formatting.md`). When a system evolves and a pattern no longer fits, refactoring *toward* simplicity is itself good practice — patterns and abstractions aren't sacred.
