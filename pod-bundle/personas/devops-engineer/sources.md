# Sources

List all source systems this Pod consumes and the minimum fields required.

## Example Source Table

| Source | Query/Scope | Required fields | Notes |
|---|---|---|---|
| GitHub | org/repo/project | id, title, status, assignee, updated_at | read-only |
| Azure DevOps | project + area path | id, title, state, assignedTo, changedDate | read-only |
| WorkIQ | decision records | title, date, author, artifact link | cite artifact |

## Evidence Policy

1. Cite primary source links for decisions.
2. Do not infer missing fields.
3. Mark inaccessible sources explicitly.
