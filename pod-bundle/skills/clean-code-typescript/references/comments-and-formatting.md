# Comments, Formatting, and Project Structure

When comments help versus hurt, how to make code self-documenting, and how to enforce consistent formatting and structure automatically so the team never argues about style.

## Comments: the principle

**Comments are not a substitute for clean code.** Well-named, well-structured code needs few comments. A comment is a small admission that the code couldn't explain itself — sometimes warranted, often a signal to refactor instead. Before writing a comment, ask: can a better name, a smaller function, or a clearer type make this comment unnecessary?

### Comments that earn their place

- **Why, not what.** Explain intent, trade-offs, and non-obvious decisions the code can't express. `// Stripe rounds half-up; we match that to avoid reconciliation drift.`
- **Warnings and gotchas.** `// NOTE: this list is 1-indexed because the upstream API is.`
- **TSDoc on public API.** Document the contract of exported functions, classes, and types (see `references/naming-and-functions.md` for tags). These power editor tooltips and generated docs.
- **TODO/FIXME with context.** `// TODO(auth): replace in-memory store with the DB once schema lands.`

### Comments that hurt

- **Restating the code.** `i++; // increment i` — pure noise. `// loop over users` above an obvious `for...of`.
- **Commented-out code.** Delete it. Version control remembers; dead code rots and confuses.
- **Stale comments.** A comment that no longer matches the code is worse than none — it actively misleads. If you change behavior, update or delete the comment.
- **Compensating for bad names.** `const d = 86400; // seconds in a day` should be `const SECONDS_PER_DAY = 86400;` with no comment.

## Self-documenting code: the Stranger Test

A useful heuristic: if someone who didn't know how to code looked at your file, could they roughly guess what it does? Code that passes this test rarely needs narration. `class ShoppingCart` with methods `addItem`, `calculateTotal`, `checkout` tells the story by itself. Achieve this through the practices in `references/naming-and-functions.md`: descriptive names, small single-purpose functions, expressive types, and avoiding side effects.

## Formatting: automate it, don't argue it

Consistent formatting improves readability and removes a whole category of review noise. The rule: **let tools enforce it** so humans never hand-format or debate style in PRs. The modern TypeScript toolchain is **ESLint** (correctness/quality rules) + **typescript-eslint** (TS-aware rules) + **Prettier** (formatting).

### ESLint with typescript-eslint (flat config, current)

ESLint moved to "flat config" (`eslint.config.mjs`). Install:

```bash
npm install --save-dev eslint typescript typescript-eslint
```

```javascript
// eslint.config.mjs
import tseslint from 'typescript-eslint';

export default tseslint.config(
  ...tseslint.configs.recommended,
  {
    rules: {
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': 'error',
      '@typescript-eslint/explicit-function-return-type': 'off', // allow inference for locals
    },
  },
);
```

Run it: `npx eslint .`. ESLint catches real bugs (unused vars, typos like `naame`, unsafe `any`, floating promises) — not just style.

> Note: older projects use TSLint (deprecated since 2019 — migrate to typescript-eslint) and `.eslintrc.json` (the legacy config format). Prefer flat config on new work.

### Prettier for formatting

```bash
npm install --save-dev prettier
```

```json
// .prettierrc
{
  "semi": true,
  "singleQuote": true,
  "trailingComma": "all",
  "printWidth": 80
}
```

Format: `npx prettier --write .`. To stop ESLint and Prettier from fighting over formatting rules, install `eslint-config-prettier` and add it last in your ESLint config so it disables stylistic ESLint rules that Prettier owns.

## Enforce on commit with Git hooks

Run the checks automatically so unformatted/lint-failing code can't enter history. Use **Husky** (manages Git hooks) + **lint-staged** (runs commands only on staged files, keeping it fast).

```bash
npm install --save-dev husky lint-staged
npx husky init
```

```bash
# .husky/pre-commit
npx lint-staged
```

```json
// .lintstagedrc.json
{
  "*.{js,ts,tsx}": ["eslint --fix", "prettier --write"],
  "*.{json,md}": ["prettier --write"]
}
```

Now every commit auto-fixes formatting and blocks on unfixable lint errors. Use **pre-commit** for fast checks (lint, format, type-check staged files) and **pre-push** for heavier ones (full test suite). Consider `commitlint` to enforce conventional commit messages.

## Project and module structure

Organize for findability and change-locality. Three common strategies:

- **By feature** — group everything for a capability together (`cart/Cart.ts`, `cart/CartService.ts`, `cart/cart.test.ts`). High cohesion; changes to a feature stay in one folder; tests colocate with code. Downside: shared utilities can get duplicated.
- **By function/layer** — group by kind (`components/`, `services/`, `utils/`). Centralizes reusable code; encourages layered architecture. Downside: one feature's code is scattered across folders.
- **Hybrid (what most teams converge on)** — feature folders at the top level, layered subfolders within (`cart/components/`, `cart/services/`). Balances cohesion with clear internal structure and scales well.

A typical TS project also has: `src/` (source), `dist/` (compiled output, git-ignored), a root `index.ts` barrel for the public surface, and `tsconfig.json`. Colocate tests next to the code they test (`Button.tsx` + `Button.test.tsx`).

### Use ES modules

Prefer ES modules (`import`/`export`) over CommonJS (`require`/`module.exports`) in new code:
- **Tree-shaking** — bundlers statically analyze ES module imports and drop unused exports, shrinking bundles.
- **Static analysis** — imports are resolved at build time, enabling better tooling, optimization, and ahead-of-time error detection.
- **Encapsulation** — only what you `export` is visible; the rest stays module-private, preventing global namespace pollution and naming conflicts.

```typescript
// utils/math.ts
export function add(a: number, b: number): number { return a + b; }

// app.ts
import { add } from './utils/math';
```

Import only what you use (`import { add } from './math'`, not `import * as math`) so tree-shaking can do its job.
