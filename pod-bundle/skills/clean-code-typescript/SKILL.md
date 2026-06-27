---
name: clean-code-typescript
description: >-
  Write clean, maintainable, idiomatic TypeScript and review or refactor existing TS/TSX for
  quality. Use when writing or editing .ts/.tsx files, naming variables, types, functions, or
  classes, designing functions, classes, modules, or interfaces, choosing types vs interfaces,
  applying generics, unions, narrowing, discriminated unions, or strict mode, handling
  sync/async errors and the unknown catch type, applying SOLID, DRY, KISS, or YAGNI, deciding
  composition vs inheritance, making data immutable, writing unit or integration tests or
  doing TDD with Jest or Vitest, or recognizing and fixing code smells (long functions, deep
  nesting, primitive obsession, feature envy, shotgun surgery, any-typed boundaries). Also
  triggers on "clean code", "refactor this", "code review", "make this more maintainable", "is
  this idiomatic", "too many parameters", "this function is too long", reducing complexity, or
  improving readability and type safety in a TypeScript codebase.
---

# Clean Code in TypeScript

This skill equips Claude to write, review, and refactor TypeScript that is readable, correct, and maintainable — using the type system as a design tool, not just an annotation layer. It covers naming, function and class design, the SOLID principles in TS idiom, leveraging types, error handling, testing/TDD, and recognizing and fixing code smells.

## When to use this skill

- Writing or editing any `.ts` / `.tsx` file, or designing a new module, class, function, or type.
- Naming things; deciding `type` vs `interface`; reaching for generics, unions, narrowing, or strict mode.
- Applying SOLID, DRY, KISS, YAGNI; choosing composition over inheritance; making data immutable.
- Handling synchronous and asynchronous errors; dealing with `unknown` in `catch`; validating untyped JSON boundaries.
- Writing unit/integration tests or practicing TDD (Jest, Vitest).
- A user asks to "clean up", "refactor", "review", or "make idiomatic" TypeScript, or flags symptoms: a function that's too long, too many parameters, deep nesting, `any` everywhere, duplicated logic, a class doing too much.

## Core principles

**Clean code reads like prose and localizes change.** The test: could a stranger guess what a file does at a glance? Optimize for the reader — code is read far more than written.

**Use the type system to make illegal states unrepresentable.** A type is a design decision. Model your *domain* (what the product needs), not a vendor's wire format. Push `unknown` data through a validation gate at every network/JSON boundary, then rely on strong types inside. See `references/type-system.md`.

**Names carry the meaning.** Functions are verbs (`calculateTotal`, `isWeekend`, `getUser`); booleans read as predicates (`isValid`, `hasItems`); types/classes are nouns (`ShoppingCart`, not `Cart`). camelCase for values/functions, PascalCase for types/classes/enums. No abbreviations, no `data`/`info`/`temp`/`doStuff`. A good name removes the need for a comment. See `references/naming-and-functions.md`.

**Small, single-purpose functions.** One reason to change. Aim for ≤3 parameters — group related ones into an object type. Prefer pure functions (same input → same output, no side effects). Isolate side effects (I/O, mutation, globals) at the edges. See `references/naming-and-functions.md`.

**Prefer composition over inheritance, and program to interfaces.** Inheritance models a strict *is-a*; composition (*has-a*) is more flexible and keeps coupling loose. Define behavior with interfaces, implement with classes. See `references/solid-principles.md`.

**DRY, KISS, YAGNI — in tension, in balance.** Remove genuine duplication of *knowledge*, but don't over-abstract coincidental similarity. Keep the simplest design that works. Don't build for imagined future requirements. See `references/refactoring-and-code-smells.md`.

## How to approach common tasks

### Writing new code
1. Name the responsibility in one sentence. If the sentence needs "and", split it.
2. Define the types first — model the domain, make invalid states impossible (unions, discriminated unions, `readonly`, branded/literal types). Turn on `strict`.
3. Write the smallest function/class that satisfies the contract. Keep parameters few; return explicit, well-typed values.
4. Push side effects to the boundary; keep the core pure and testable.
5. Validate external input (`req.body`, `res.json()`, env, user input) with type guards before trusting it.
6. Let the name do the explaining; add a comment only for *why*, never *what*.

### Reviewing / refactoring existing code
1. Read for the smell first (long function, many params, deep nesting, `any` at boundaries, duplicated logic, a god class, primitive obsession, feature envy). See `references/refactoring-and-code-smells.md`.
2. Add a test characterizing current behavior before you change it (especially if untested).
3. Apply the smallest safe transformation: Extract Function, Extract Type, Replace Conditional with Polymorphism/Strategy, Introduce Parameter Object, Replace Primitive with Type, Replace `any` with a precise type or `unknown` + guard.
4. Re-run tests after each step. Refactor in small, reversible moves.
5. Tighten types as you go: narrow `any` → `unknown` → precise; add `readonly`; convert loose strings to unions/literals.

### Designing classes and modules
- One class, one responsibility. Favor `private`/`protected`; expose a small surface. Use getters/setters only to enforce invariants, not as reflexive boilerplate.
- Depend on abstractions (interfaces), not concretions — inject dependencies rather than `new`-ing them inside. This is what makes code testable.
- Organize files by feature first, then by layer within (the hybrid layout most teams converge on). Use ES modules (`import`/`export`) for tree-shaking and static analysis. See `references/solid-principles.md`.

### Error handling
- Choose deliberately: `try/catch` for synchronous and `async/await` flows; `.then().catch().finally()` for promise chains. Always handle rejections.
- In `catch (e)`, `e` is `unknown` (correct — JS can throw anything). Narrow with `e instanceof Error` before reading `.message`.
- Validate input early ("fail fast"); return or throw with a clear, user-safe message; log details securely (never leak stack traces or secrets to users). See `references/error-handling.md`.

### Testing & TDD
- Red → Green → Refactor. Arrange-Act-Assert. One behavior per test. Descriptive test names that state the expected behavior.
- Keep tests independent (no shared mutable state), test edge cases and boundaries, and keep tests DRY with helpers/hooks. Mock external dependencies via injection where possible; reserve module mocks for imported singletons. See `references/testing-and-tdd.md`.

## Common pitfalls & anti-patterns

- **`any` at the boundary.** It silently disables the compiler exactly where data is least trustworthy. Use `unknown` + a type guard. A passing build over `any` is a false sense of safety.
- **Comments that restate code.** `// increment i` is noise. Delete dead/commented-out code; rename instead of explaining; comment only non-obvious *why*.
- **God functions/classes.** Doing calculation + formatting + I/O in one place. Split by responsibility (SRP).
- **Long parameter lists / boolean flags.** `render(true, false, null)` is unreadable. Use a parameter object or split the function.
- **Primitive obsession.** Passing raw `string`/`number` where a union, literal, or small type would prevent mixups (e.g. `type Role = 'admin' | 'user' | 'guest'`).
- **Optional-field soup.** Many `?:` fields encoding "maybe everything exists." Use discriminated unions so only valid states compile; add a `never` exhaustiveness check in switches.
- **Inheritance for code reuse.** Reaching for `extends` to share helpers creates rigid hierarchies. Compose instead.
- **Swallowed async errors / floating promises.** Unhandled rejections and un-awaited promises. Always `await` or `.catch`.
- **Over-engineering.** Applying a design pattern, abstraction, or generic where a plain function suffices — violates KISS/YAGNI. Introduce structure only when it earns its keep.
- **Mutating shared data.** Reassigning inputs or globals causes spooky action at a distance. Return new values; use `readonly`, `as const`, and copy (`{...obj}`, `[...arr]`).

## Reference files

- `references/naming-and-functions.md` — naming conventions and heuristics; small-function design; parameters and parameter objects; pure functions; side-effect management; immutability; TSDoc/JSDoc for documentation.
- `references/comments-and-formatting.md` — when comments help vs. hurt; self-documenting code; the "Stranger Test"; ESLint + Prettier + typescript-eslint setup; pre-commit hooks (Husky, lint-staged); module/file organization.
- `references/solid-principles.md` — the five SOLID principles with idiomatic TypeScript examples; composition over inheritance; interfaces vs classes; dependency injection; encapsulation and access modifiers.
- `references/type-system.md` — `type` vs `interface`; unions, intersections, literals; generics and constraints; narrowing and type guards; discriminated unions and exhaustiveness; mapped/conditional/utility types; `strict` mode and `tsconfig`; modeling the domain and the runtime-gate pattern.
- `references/error-handling.md` — error categories; synchronous `try/catch` and input validation; async errors with promises and `async/await`; the `unknown` catch type; fail-fast, graceful degradation, secure logging; debugging tools (source maps, breakpoints, console).
- `references/testing-and-tdd.md` — testing vocabulary; unit vs integration; Jest and Vitest setup; mocks/stubs/spies and dependency injection; the Red-Green-Refactor TDD cycle; AAA pattern and test best practices.
- `references/refactoring-and-code-smells.md` — DRY/KISS/YAGNI in depth; a catalog of code smells with TypeScript fixes; concrete refactoring techniques; safe refactoring workflow.
