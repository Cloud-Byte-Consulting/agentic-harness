# Tech Design Doc: <feature / change>

> Engineering design doc (a.k.a. RFC) — the detailed "how" once the approach is
> approved. Lighter than a [PDR](pdr.template.md); deeper than an [ADR](adr.template.md).

- **Author:** <name> · **Status:** Draft | In review | Approved | Implemented
- **Reviewers:** <names> · **Date:** YYYY-MM-DD · **Tracking:** #<id>

## Context
The problem and why now. Link the PRD/issue. State current behavior.

## Goals / non-goals
- **Goals:** <what success looks like>
- **Non-goals:** <explicitly excluded>

## Proposed design
The approach in detail: data model, APIs/interfaces, control flow, key types.
Include a diagram and the critical code/contract sketches.

## Alternatives considered
- <alternative> — why not.

## Impact
- **Backward compatibility / migration:** …
- **Security & privacy:** …
- **Performance:** expected cost, hot paths.
- **Observability:** metrics/logs/traces added.

## Test plan
- Unit: <…> · Integration: <…> · E2E / manual: <…>
- Rollout: flag / phased / canary · Rollback: <…>

## Open questions
- <question> → owner / resolution
