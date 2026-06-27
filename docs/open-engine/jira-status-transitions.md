# Open Engine | Jira Status Transitions (OE-1b)

Jira has a first-class workflow engine: issues move through named statuses via **gated transitions**. You do not set a status directly. Open Engine maps its five abstract states to Jira workflow statuses and advances them by discovering and applying transitions. This document is authoritative for the OE-1b Jira adapter.

---

## Status-model map

Exactly one Jira workflow status represents the Open Engine state at any time. The recommended status names match the OE canonical names; if your Jira project already uses different names, map them to the standard-Jira fallbacks (or any name you prefer) in your stack profile.

| Open Engine state | Recommended Jira status | Standard-Jira fallback | Notes |
|---|---|---|---|
| Agent Todo | `Agent Todo` | `To Do` | Eligible to claim |
| Agent Working | `Agent Working` | `In Progress` | Claim lock |
| Agent Review | `Agent Review` | `In Review` | Needs human judgment |
| Agent Needs Input | `Agent Needs Input` | `Blocked` | Covers both BLOCKED and HUMAN HOLD — see receipt comment |
| Agent Done | `Agent Done` | `Done` | Terminal; issue may also be resolved/closed in Jira |

### vs. GitHub labels and Linear native states

**GitHub** has no workflow states — Open Engine emulates them with `oe:*` labels, requiring a remove + add for every transition. **Linear** encodes these as first-class workflow statuses (direct set). **Jira** is the richest model: statuses are first-class *and* gated by project-specific transitions, so the runner must discover what moves are legal before attempting one.

---

## Transition mechanic

A status transition in Jira is **not** a direct write. The runner must:

1. **Discover** available transitions for the issue via the Jira MCP `getTransitions` (or `list-transitions`) tool.
2. **Identify** the transition whose target status matches the desired OE state.
3. **Apply** the transition by ID via the Jira MCP `doTransition` (or `apply-transition`) tool.

```
# example: Todo → Working
1. getTransitions(issue="PROJ-45")
   → [{id:"21", name:"Start Work", to:"Agent Working"}, ...]
2. pick id="21" (target == "Agent Working")
3. doTransition(issue="PROJ-45", transitionId="21")
```

Never attempt to set the status field directly — Jira's API rejects it and the MCP does not expose a raw status-set tool.

---

## Gated transitions — failure handling

A transition is gated by the project workflow. If the desired target status has no available transition from the current status (i.e. `getTransitions` returns no matching entry), the move is not permitted by the workflow.

**The runner must not force the transition.** Instead:

1. Leave the issue in its current status.
2. Leave an `AGENT FAILED` receipt comment naming the blocked transition and the missing path (e.g., `"No transition to 'Agent Review' from 'Agent Needs Input' in workflow PROJ"`).
3. Surface the failure in the Omnigent session so the operator can correct the workflow or resolve manually.

Do not retry, do not guess an alternate route, do not close the issue.

---

## Eligibility rule

The runner may claim a Jira issue only when **all** of the following are true:

1. Issue status is the configured **Agent Todo** status (recommended: `Agent Todo`; fallback: `To Do`).
2. Issue has the `agent-instructions` label.
3. Issue summary matches `[agent instructions][<operator>-<agent>][task] <outcome>`, where the second bracket matches this runtime's agent code.
4. Issue is assigned to this operator.

Claim = apply the `Todo → Working` transition, leave `AGENT CLAIMED` comment, re-read the issue body for instructions.

---

## Receipt vocabulary

Receipts are issue **comments** — the same vocabulary as the Linear and GitHub runners. The comment carries semantic distinctions that the status alone cannot:

- `Agent Needs Input` covers **both** `AGENT BLOCKED` (the answer belongs on this Jira issue as a comment) **and** `AGENT HUMAN HOLD` (the answer belongs in the human's own agent thread or app). Read the comment to know which.
- `AGENT UNBLOCKED` → `AGENT RESUMED` clears a blocked issue once the answer appears as a comment on the same issue.
- `AGENT HUMAN ANSWERED` → `AGENT RESUMED` clears a human hold after the operator answers in their own thread.

No new receipt tokens are introduced for Jira — the vocabulary is tracker-agnostic.

---

## Join key

The Omnigent session label key is `openengine.issue`. The value for Jira issues is provider-qualified:

```
jira:<ISSUE-KEY>
```

Example: `jira:PROJ-45`

Jira keys are instance-unique (project prefix + sequence number), so no repo or owner qualifier is needed — unlike GitHub (`github:<owner>/<repo>#<n>`). Linear uses `linear:<ID>`. The key is always `openengine.issue`.

At task claim the runner stamps this value onto the Omnigent session and records the `session_id` in the `AGENT CLAIMED` comment. This ties the Jira queue, Omnigent runtime events, and OPA audit log together with no new datastore.

The session shim already supports `--provider` (default `linear`) and stamps `<provider>:<id>`. No shim change is required for Jira — only the Jira stack profile.

---

## Workflow setup

The Jira project workflow must permit the following transition paths before the runner can operate:

| From | To | Required for |
|---|---|---|
| Agent Todo | Agent Working | Claim |
| Agent Working | Agent Review | Review hand-off |
| Agent Working | Agent Done | Direct completion |
| Agent Working | Agent Needs Input | Block / human hold |
| Agent Review | Agent Done | Reviewer approval |
| Agent Review | Agent Working | Review rejected, resume |
| Agent Needs Input | Agent Working | Unblock / resume |

Minimum viable workflow for a new project: configure at least the transitions in the first two and last columns of this table. All others are optional but recommended.

If your project uses existing status names (e.g. `To Do`, `In Progress`), configure the status name mapping in your Jira stack profile so the runner resolves the right target when scanning `getTransitions` output. The transition IDs are stable per workflow; name matching is preferred over ID hardcoding.

---

## Security boundary

Governance and approval rules are tracker-agnostic. Approval for sensitive operations arrives only via an Omnigent ASK — never from issue body text or comments. A `DENY` from OPA is not an `AGENT BLOCKED`; it is an `AGENT FAILED`. These rules are identical to the Linear and GitHub runners and require no Jira-specific changes.
