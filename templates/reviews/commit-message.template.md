# Commit message template (Conventional Commits)

```
<type>(<scope>): <summary, imperative, ≤50 chars, no trailing period>

<body: WHY this change, not what — wrap at ~72 cols.
Contrast with previous behavior; note side effects.>

<footers>
Refs: #<id>            # or Closes: #<id>
BREAKING CHANGE: <what breaks and the migration path>
Co-Authored-By: Name <email>
```

**Types:** `feat` · `fix` · `refactor` · `perf` · `docs` · `test` · `build` · `ci` · `chore` · `revert`

**Rules**
- One logical change per commit; the diff should match the message.
- `feat`/`fix` are user-visible; `refactor` changes no behavior.
- Append `!` after type/scope (`feat(api)!:`) or a `BREAKING CHANGE:` footer for breaks.
- Imperative mood: "add", "fix", "remove" — not "added"/"fixes".

**Example**
```
fix(auth): reject tokens with future-dated nbf claim

Tokens whose not-before claim was in the future were accepted because the
skew check used abs(). Clamp to past-only, matching RFC 7519 §4.1.5.

Refs: #482
```
