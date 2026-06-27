# PR / code review comment templates

Lead with a label so the author knows whether they're blocked. Cite evidence
(`file:line`), give the *why*, and propose a concrete fix.

**Labels:** `blocking` · `non-blocking` · `nit` · `question` · `praise`

---

### Single review comment

```
[<label>] <one-line observation>

Why: <impact — bug, security, perf, readability, contract>
Evidence: <path/to/file.ext:line> (or test output / spec link)
Suggestion:
```<lang>
<proposed code>
```
```

### Inline suggestion (GitHub/GitLab apply-able)
````
```suggestion
<exact replacement lines>
```
````

### Review summary (top-level)
```
**Verdict:** Approve | Approve with nits | Request changes

**Blocking (must fix)**
- <file:line> — <issue>

**Non-blocking / nits**
- <file:line> — <suggestion>

**Questions**
- <file:line> — <question>

Strengths: <what was done well>
```

> Discipline: distinguish blocking from taste; never request changes without an
> evidence reference; prefer a suggested diff over prose.
