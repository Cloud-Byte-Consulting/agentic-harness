# SOLID, Composition, and Class/Module Design

The five SOLID principles in idiomatic TypeScript, plus the design decisions that make code flexible and testable: composition over inheritance, interfaces vs classes, dependency injection, and encapsulation.

## Table of contents
- The five SOLID principles
- Composition over inheritance
- Interfaces vs classes: when to use each
- Dependency injection and testability
- Encapsulation and access modifiers

## The five SOLID principles

SOLID is a set of object-oriented design guidelines. They are not TypeScript-specific, but TypeScript's interfaces, generics, and access modifiers make them clean to express.

### S — Single Responsibility Principle (SRP)

*A class or module should have one reason to change.* If a class handles persistence **and** formatting **and** business rules, a change to any one risks the others. Split responsibilities into focused units.

```typescript
// Smell: this class fetches, formats, and persists — three reasons to change.
class UserReport {
  load() { /* DB access */ }
  format() { /* HTML generation */ }
  save() { /* file I/O */ }
}

// Clean: one responsibility each, composed by a coordinator.
class UserRepository { load(id: string) { /* ... */ } }
class ReportFormatter { format(user: User): string { /* ... */ return ""; } }
class ReportWriter { save(html: string) { /* ... */ } }
```

### O — Open/Closed Principle (OCP)

*Open for extension, closed for modification.* You should add new behavior without editing existing, tested code. In TypeScript, achieve this with interfaces/strategies rather than growing `switch`/`if-else` chains.

```typescript
// Smell: every new shape edits this function.
function area(shape: { kind: string; /* ... */ }): number {
  if (shape.kind === "circle") { /* ... */ }
  else if (shape.kind === "square") { /* ... */ }
  // adding "triangle" means editing here, risking the others
  return 0;
}

// Clean: add a new class, touch nothing existing.
interface Shape { area(): number; }
class Circle implements Shape { constructor(private r: number) {} area() { return Math.PI * this.r ** 2; } }
class Square implements Shape { constructor(private s: number) {} area() { return this.s ** 2; } }
class Triangle implements Shape { constructor(private b: number, private h: number) {} area() { return 0.5 * this.b * this.h; } }
const totalArea = (shapes: Shape[]) => shapes.reduce((sum, s) => sum + s.area(), 0);
```

(For data, a discriminated union with an exhaustive `switch` is the type-driven equivalent — see `references/type-system.md`.)

### L — Liskov Substitution Principle (LSP)

*Subtypes must be usable wherever their base type is expected, without surprises.* A subclass must honor the base class's contract — not strengthen preconditions, weaken postconditions, or throw where the base wouldn't. The classic violation: `Square extends Rectangle` where setting width also changes height, breaking code that assumed they were independent. When substitution would break expectations, prefer composition or a shared interface over inheritance.

```typescript
// All implementations must truly behave like a PaymentMethod —
// same contract, no surprising side effects or thrown "not supported".
interface PaymentMethod { process(amount: number): void; }
class CreditCard implements PaymentMethod { process(amount: number) { /* ... */ } }
class DebitCard implements PaymentMethod { process(amount: number) { /* ... */ } }

const methods: PaymentMethod[] = [new CreditCard(), new DebitCard()];
methods.forEach(m => m.process(100)); // any method substitutes safely
```

### I — Interface Segregation Principle (ISP)

*Don't force clients to depend on methods they don't use.* Prefer several small, role-specific interfaces over one fat interface. A class can implement many.

```typescript
// Smell: a printer that doesn't fax is forced to implement fax().
interface Machine { print(): void; scan(): void; fax(): void; }

// Clean: segregated roles; implement only what applies.
interface Printer { print(): void; }
interface Scanner { scan(): void; }
class SimplePrinter implements Printer { print() { /* ... */ } }
class AllInOne implements Printer, Scanner { print() {} scan() {} }
```

### D — Dependency Inversion Principle (DIP)

*Depend on abstractions, not concretions.* High-level code shouldn't import low-level details directly; both should depend on an interface. This is what makes code swappable and testable.

```typescript
// Smell: OrderService is welded to a concrete logger and email client.
class OrderService {
  private mailer = new SmtpMailer(); // hard dependency, can't substitute in tests
}

// Clean: depend on an interface; inject the implementation.
interface Mailer { send(to: string, body: string): Promise<void>; }
class OrderService {
  constructor(private mailer: Mailer) {} // any Mailer works, including a fake
  async placeOrder(/* ... */) { await this.mailer.send("x@y.com", "Order placed"); }
}
```

## Composition over inheritance

Inheritance models a strict **is-a** relationship (a `Dog` is an `Animal`) and is rigid — the hierarchy is fixed at design time and a change to the base ripples to every subclass. Composition models **has-a** (a `Band` has a `Singer`, `Guitarist`, `Drummer`) and is far more flexible: you assemble behavior from parts and can swap them independently.

**Default to composition.** Reach for inheritance only when there's a genuine, stable is-a relationship and you want to share both interface and implementation.

```typescript
// Inheritance: only when the is-a relationship is real and stable.
class Animal { breathe() { console.log("breathing"); } }
class Dog extends Animal { bark() { console.log("woof"); } }

// Composition: assemble behavior; each part is independently replaceable.
class Singer { sing() {} }
class Guitarist { play() {} }
class Band {
  constructor(
    private singer: Singer = new Singer(),
    private guitarist: Guitarist = new Guitarist(),
  ) {}
  perform() { this.singer.sing(); this.guitarist.play(); }
}
// Swap a member, add a keyboardist — no base-class surgery, nothing else breaks.
```

A common, effective TypeScript pattern: **define behavior with interfaces, implement it with classes, and compose objects out of those implementations.** This keeps coupling loose and units replaceable.

## Interfaces vs classes: when to use each

Both can enable polymorphism. Choose based on what you need:

**Use an interface when** you want to define a contract or shape that multiple, possibly unrelated, implementations follow. Interfaces exist only at compile time — zero runtime cost, no runtime coupling. Ideal for: a common API, enabling polymorphism without forcing inheritance, keeping implementations loosely coupled and swappable.

**Use a class when** you need runtime behavior and shared implementation: constructors, instance state, methods, inheritance/overriding, and objects that are actually instantiated and carry behavior.

```typescript
// Interface = the contract (compile-time only).
interface PaymentMethod { process(amount: number): void; }

// Classes = the runtime implementations.
class CreditCard implements PaymentMethod { process(amount: number) { /* ... */ } }
class Stripe implements PaymentMethod { process(amount: number) { /* ... */ } }
```

Rule of thumb: **interfaces to define behavior, classes to implement it.**

## Dependency injection and testability

Don't `new` your dependencies inside the class that uses them — accept them through the constructor (or a parameter). Injecting dependencies is what makes a unit testable in isolation: in tests you pass a fake/mock; in production you pass the real thing. This is DIP in practice and is the single biggest lever for testability.

```typescript
interface Database { getUserById(id: string): Promise<User>; }

// Inject the dependency — production passes the real DB, tests pass a fake.
async function getUserInfo(userId: string, db: Database): Promise<User> {
  return db.getUserById(userId);
}
```

(See `references/testing-and-tdd.md` for how this enables clean mocking without module-level magic.)

## Encapsulation and access modifiers

Encapsulation bundles data with the methods that operate on it and hides internals behind a small, controlled surface. Expose only what callers need.

- `public` (default) — accessible anywhere.
- `private` — only within the class. Use it for internal state by default.
- `protected` — within the class and its subclasses.
- `#field` — true runtime-private (ECMAScript private fields), inaccessible even via bracket access; stronger than `private`, which is only compile-time.

Use **getters/setters to enforce invariants**, not as reflexive boilerplate. A setter that just assigns adds nothing; a setter that validates earns its place.

```typescript
class Cake {
  private flavor: string;
  constructor(flavor: string) { this.flavor = flavor; }

  get flavorLabel(): string { return this.flavor.toUpperCase(); } // formatting on read

  set newFlavor(value: string) {                                   // validation on write
    if (value.length < 3) throw new Error("Flavor must be at least 3 characters.");
    this.flavor = value;
  }
}
```

TypeScript also offers **parameter properties** to cut boilerplate — declaring and assigning a field in the constructor signature:

```typescript
class User {
  constructor(public readonly id: number, private name: string) {} // declares + assigns both
}
```

Keep the public surface minimal: the less you expose, the less can be misused, and the more freely you can refactor internals without breaking callers.
