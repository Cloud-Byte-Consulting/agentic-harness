# Azure DevOps work item template

> CLI: `az boards work-item create --title "…" --type "User Story" --area "Proj\\Team" --iteration "Proj\\Sprint 5"`
> Fields below mirror the ADO work-item form (Agile process).

**Work Item Type:** Epic | Feature | User Story | Task | Bug
**Title:** `<concise summary>`
**Area Path:** `<Project\\Team>` · **Iteration Path:** `<Project\\Sprint N>`
**Assigned To:** `<user>` · **State:** New | Active | Resolved | Closed
**Tags:** `<tag; tag>`
**Story Points / Effort:** `<n>` · **Priority:** 1–4 · **Severity (Bug):** 1-Critical … 4-Low

---

## Description
<context and intent, with links to the source of truth>

## Acceptance Criteria
- [ ] <observable, testable condition>
- [ ] <…>

## Repro Steps  *(Bug)*
1. …
2. …
- **Expected:** …
- **Actual:** …
- **System info:** build `<id>`, environment `<…>`

## Links
- Parent: `#<id>` (Feature/Epic) · Related: `#<id>` · PR: `<url>`
