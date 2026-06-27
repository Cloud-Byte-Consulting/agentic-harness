# Testing and Test-Driven Development

Tests make code reliable, document intended behavior, and let you refactor without fear. This file covers testing vocabulary, the levels of testing, setup with modern runners, mocking, and the TDD cycle with best practices.

## Table of contents
- Why tests matter
- Testing vocabulary
- Levels of testing
- Test runners and setup (Vitest, Jest)
- Mocks, stubs, spies, and dependency injection
- Test-Driven Development (Red-Green-Refactor)
- Best practices

## Why tests matter

Without tests, defects slip to production where they're expensive to fix. Well-written tests catch regressions early, give you confidence to change code, and serve as living documentation of how a unit is supposed to behave. They also pressure your design: code that's hard to test is usually too coupled — testability and clean design go hand in hand.

## Testing vocabulary

- **Test case** — a scenario checking a specific unit with specific inputs against an expected outcome.
- **Assertion** — the statement that verifies the result (`expect(add(2, 2)).toBe(4)`).
- **Test suite** — a group of related test cases (a `describe` block).
- **Test fixture** — a fixed setup/state a test runs against.
- **Mock** — a stand-in for a real dependency, programmed with expected behavior, used to isolate the unit under test.
- **Stub** — a minimal canned implementation that returns fixed values.
- **Spy** — wraps a real or fake function to record how it was called (arguments, call count).

## Levels of testing

- **Unit testing** — tests the smallest isolated piece (a function, a method, a class) without external systems. The foundation; fast and precise.
- **Integration testing** — tests how units work *together* and across boundaries (e.g. controller → service → storage, or an HTTP endpoint end to end). TypeScript's compile-time checks ensure types line up; integration tests verify that independently correct components actually interact correctly at runtime.
- (System and acceptance testing exist above these; unit and integration give the most leverage day to day.)

## Test runners and setup

A test runner executes your suite and reports pass/fail. Popular choices for TypeScript:

- **Vitest** — fast, Vite-native, first-class TypeScript and ESM support; the modern default for new projects.
- **Jest** — mature, huge ecosystem, built-in mocking and snapshots; still extremely common.
- **Mocha** — flexible/lightweight, bring-your-own assertions.
- **Cypress / Playwright** — primarily end-to-end and browser testing.

### Vitest setup

```bash
npm install --save-dev vitest typescript @types/node
```

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
export default defineConfig({
  test: { globals: true, environment: 'node' }, // use 'jsdom' for DOM/React code
});
```

```typescript
// square.ts
export function square(n: number): number { return n * n; }

// square.test.ts
import { describe, it, expect } from 'vitest';
import { square } from './square';

describe('square', () => {
  it('squares a positive number', () => expect(square(3)).toBe(9));
  it('squares a negative number', () => expect(square(-4)).toBe(16));
});
```

Add `"test": "vitest"` to `package.json` scripts and run `npm test`. (Vitest and recent tooling expect Node 20+.)

Jest is similar (`describe`/`it`/`expect`); with TS it's typically configured via `ts-jest` or Babel. The test *concepts* below apply to any runner.

## Mocks, stubs, spies, and dependency injection

To unit-test code that talks to a database, API, or other external system, replace the dependency so tests stay fast, deterministic, and isolated.

### Preferred: inject the dependency (manual mocks)

The cleanest approach uses plain objects and dependency injection — no module-system magic. This is why DIP (see `references/solid-principles.md`) makes code testable.

```typescript
// Under test: dependency is a parameter.
async function getUserInfo(userId: string, db: { getUserById(id: string): Promise<User> }) {
  return db.getUserById(userId);
}

// Test: pass a fake.
import { vi, test, expect } from 'vitest';
test('fetches user info', async () => {
  const fakeDb = {
    getUserById: vi.fn().mockResolvedValue({ id: '123', name: 'John Doe' }),
  };
  const user = await getUserInfo('123', fakeDb);
  expect(user.name).toBe('John Doe');
  expect(fakeDb.getUserById).toHaveBeenCalledWith('123'); // spy assertion
});
```

(`vi.fn()` is Vitest; `jest.fn()` is the Jest equivalent.)

### When you must: mock the imported module

If the dependency is imported directly rather than injected, intercept the import:

```typescript
import { vi, test, expect } from 'vitest';
import * as dbService from '../src/databaseService';
import { getUserInfo } from '../src/getUserInfo';

vi.mock('../src/databaseService'); // replace the real module

test('uses a mocked module', async () => {
  vi.mocked(dbService.getUserById).mockResolvedValue({ name: 'Jane Doe' } as User);
  const user = await getUserInfo('123');
  expect(user.name).toBe('Jane Doe');
});
```

Reserve module mocking for imported singletons you can't inject — injection keeps tests simpler and the design cleaner.

### Integration test example (HTTP endpoint)

For integration tests of an API, drive the app with a request library (e.g. Supertest) without binding a real port — export the app instance and let the test framework start/stop it:

```typescript
import { describe, it, expect } from 'vitest';
import request from 'supertest';
import app from '../src/app';

describe('Posts API', () => {
  it('creates then retrieves a post', async () => {
    const created = await request(app).post('/posts').send({ title: 'Hi', content: '...' });
    expect(created.status).toBe(201);
    const fetched = await request(app).get(`/posts/${created.body.id}`);
    expect(fetched.status).toBe(200);
    expect(fetched.body.title).toBe('Hi');
  });
});
```

## Test-Driven Development (Red-Green-Refactor)

TDD reverses the usual order: write the test *first*, then the code. The cycle:

1. **Red** — write a test for the behavior you want. Run it; it fails (no implementation yet). This forces you to specify the expected outcome before coding.
2. **Green** — write the *minimum* code to make the test pass. Resist gold-plating.
3. **Refactor** — clean up the implementation (and tests) while keeping them green.

Repeat per small behavior. TDD improves design (you think about the interface and outcomes before the internals), guarantees coverage, and prevents whole classes of logical errors — you've stated the expectation up front.

```typescript
// Red: write this first; Calculator.add doesn't exist yet.
describe('Calculator', () => {
  it('adds two numbers', () => {
    expect(Calculator.add(2, 3)).toBe(5);
  });
});
// Green: implement just enough -> static add(a, b) { return a + b; }
// Refactor: tidy up, keep the test green.
```

## Best practices

- **Descriptive test names** — state the expected behavior, not the mechanics: `it('returns 0 for an empty cart')`, not `it('test1')`.
- **One behavior per test** — a failing test should point to exactly one thing. Split addition and subtraction into separate tests.
- **Arrange-Act-Assert (AAA)** — structure each test in three clear phases: set up inputs, run the unit, verify the result. Readable and consistent.

  ```typescript
  it('adds two numbers', () => {
    const a = 2, b = 3;                 // Arrange
    const result = Calculator.add(a, b); // Act
    expect(result).toBe(5);              // Assert
  });
  ```

- **Independent tests** — no test should depend on state left by another. Tests must pass in any order; reset shared state in `beforeEach`/`afterEach`.
- **Test edge cases and boundaries** — not just the happy path. For a range of 18–65, test 17, 18, 65, 66; also empty inputs, nulls, and extremes. Boundary bugs hide where typical inputs don't reach.
- **Keep tests DRY** — extract repeated setup/assertions into helpers or `beforeEach` hooks; but don't over-abstract to the point where a test is hard to read.
- **Mock external dependencies, not the unit** — isolate the thing under test; don't mock so much that you test the mocks instead of the code.

Good tests are themselves clean code: clear names, small scope, no duplication, easy to read.
