# GitLab issue template

> CLI: `glab issue create --title "…" --description "$(cat issue.md)"`
> For project-wide forms, save as `.gitlab/issue_templates/<Name>.md`.
> Quick actions (each on its own line) set metadata when the issue is created.

```
/assign @user
/label ~"type::bug" ~"priority::high" ~"area::api"
/milestone %"v1.4"
/weight 3
/estimate 1d
/epic &12
/due 2026-07-01
```

---

## Summary
<what and why, with links to evidence>

## Steps to reproduce  *(bug)*
1. …
2. …

**Expected vs. actual**
- Expected: …
- Actual: …

<details><summary>Logs</summary>

```
<paste>
```
</details>

## Acceptance criteria
- [ ] <observable, testable condition>

## Tasks
- [ ] <subtask>

## Related
- Related to #<id> · Blocked by #<id> · MR: !<id>
