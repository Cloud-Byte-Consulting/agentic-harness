# Open Engine | Contract Declaration

**Open Engine = Omnigent's task-queue front-end contract.**

## Tracker neutrality

Open Engine is tracker-neutral over the abstract tracker surface `{issues, labels, comments, status, query}`. Supported providers: **Linear** (live), **GitHub Issues** (next, OE-1b), **Jira** (enterprise, OE-1b/OE-3). A provider is `{its MCP injected via the stack profile}` + `{a status-model map}` + `{a runner-prompt variant}` — no central router, one queue-runner/cron per tracker. The status vocabulary (`AGENT CLAIMED` / `AGENT DONE` / `NEEDS INPUT`, mapping to the abstract states Todo / Working / Review / Needs-Input / Done) is tracker-agnostic; only the implementation is provider-specific: Linear native workflow states, GitHub Issues labels (`oe:working`, `oe:done`, …), Jira workflow transitions (gated, require transition IDs + matching workflow).

## The join key

| Component | Value |
| --- | --- |
| Tracker task ref | `tracker_issue_id` — provider-qualified: `<provider>:<id>` |
| Omnigent session | `session_id` |
| Subject (human/operator) | `subject_id` (Entra OID) |

The label **key** is fixed: `openengine.issue`. The label **value** is provider-qualified: `linear:ENG-123` / `github:owner/repo#123` / `jira:PROJ-45`. Docs call this the abstract **task ref**. The shim (`/home/bittahcriminal/air/workspace/omnigent/scripts/openengine_session_shim.py`) now stamps this provider-qualified value (`<provider>:<id>`) — the GitHub adapter (OE-1b) landed the widening.

**Binding rule:** at task claim, the runner stamps:
- Tracker label `openengine.issue=<provider>:<id>` on the Omnigent session
- Omnigent `session_id` into the `AGENT CLAIMED` receipt on the tracker issue
- `session_id` into the status ledger (`Last session ID` field)

This single shared key ties the three planes together — tracker work queue, Omnigent runtime events, and OPA/policy decision logs — with no new datastore.

## Contract surfaces

| Plane | Authority | Join key |
| --- | --- | --- |
| Tracker (queue) | Task issue tracking | `tracker_issue_id` (`<provider>:<id>`) |
| Omnigent (runtime) | Session lifecycle, ASK, native policy | `session_id` |
| OPA / Sentry (authz) | OPA Rego bundle, policy verdicts | `session_id` ↔ `subject_id` |

All three planes are keyed by `session_id`; `session_id` is bound to `tracker_issue_id` by the label `openengine.issue=<provider>:<id>` and to `subject_id` via the Omnigent session context. OPA gates tool calls tracker-agnostically; the audit three-plane correlation keys on `session_id` only — OE-2 and OE-3 need no rework for multi-tracker.

## Receipt marker

Every `AGENT CLAIMED` and `AGENT DONE` comment MUST include the following HTML comment on its own line (invisible in rendered Markdown, regex-parseable by the correlation query):

```
<!-- openengine session=<session_id> task_ref=<provider>:<id> phase=claimed|done -->
```

**Parser regex:** `<!--\s*openengine\s+(.*?)\s*-->` — then split whitespace-separated `key=value` tokens.

| Field | Value |
| :--- | :--- |
| `session` | the `OMNIGENT_SESSION_ID` for this run |
| `task_ref` | the provider-qualified issue ref (`linear:ENG-123`, `github:owner/repo#42`, `jira:PROJ-45`) |
| `phase` | `claimed` in `AGENT CLAIMED`; `done` in `AGENT DONE` |

**Rule:** `session` and `task_ref` MUST be identical between the `AGENT CLAIMED` and `AGENT DONE` markers for the same task; only `phase` differs. The `session` value is the canonical join key that ties the tracker receipt to Omnigent runtime events and OPA decision logs (see [audit-correlation-design.md §1](../architecture/audit-correlation-design.md)).
