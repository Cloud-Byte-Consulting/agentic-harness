# PR description template

**Title:** `<type>(<scope>): <summary>`  (Conventional Commits — mirrors the squash commit)

## Summary

What this PR does and why, in 1–3 sentences.

## Context / motivation

The problem being solved. Link the driving issue/ADR/PRD.

- Closes #<id>   <!-- or: Refs #<id> -->

## What changed

- <key change 1>
- <key change 2>

## How it was verified

Evidence, not assertions — paste commands/output, link CI, attach screenshots.

- [ ] Unit / integration tests added or updated
- [ ] Manually verified: <what you ran and observed>
- [ ] No behavior change (refactor only)

## Risk & rollback

- **Risk level:** low | medium | high
- **Blast radius:** <what could break>
- **Rollback:** <revert / flag off / migration down>

## Reviewer notes

Where to start, areas wanting scrutiny, anything intentionally deferred.

## Checklist

- [ ] Scoped to one concern
- [ ] Docs / changelog updated if user-facing
- [ ] No secrets, no debug code
- [ ] Breaking changes called out (with migration)
