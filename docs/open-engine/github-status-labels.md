# Open Engine | GitHub Status Labels (OE-1b)

GitHub Issues has no native workflow states — issues are open or closed. Open Engine models the five abstract workflow states using a single `oe:*` label per issue. This document is authoritative for the OE-1b GitHub adapter.

---

## The oe:* label set

Exactly one `oe:*` status label must be on any Open Engine issue at a time. Remove the current label before adding the next one.

| Open Engine state | GitHub label | Issue state | Notes |
|---|---|---|---|
| Agent Todo | `oe:todo` | open | Eligible to claim |
| Agent Working | `oe:working` | open | Claim lock |
| Agent Review | `oe:review` | open | Needs human judgment |
| Agent Needs Input | `oe:needs-input` | open | Blocked or on hold — see receipt |
| Agent Done | `oe:done` | open or closed | Optionally also close the issue |

### vs. Linear native states

Linear encodes these states as first-class workflow statuses in the team's status column. GitHub has no equivalent — the `oe:*` labels are the full state machine. Everything else (open/closed, milestones, assignees) is orthogonal and unchanged by runner transitions.

---

## Transition rule

A status transition = **remove** the current `oe:*` label, **add** the new one.

The GitHub MCP supports `add_issue_labels` and `remove_issue_labels`. Both calls are required for every transition. Never leave two `oe:*` labels on the same issue.

```
# example transition: Todo → Working
remove_issue_labels: ["oe:todo"]
add_issue_labels:    ["oe:working"]
```

---

## Eligibility rule

The runner may claim a GitHub issue only when **all** of the following are true:

1. Issue is **open**.
2. Issue has the `agent-instructions` label.
3. Issue has the `oe:todo` label.
4. Issue title matches `[agent instructions][<operator>-<agent>][task] <outcome>`, where the second bracket matches this runtime's agent code.
5. Issue is assigned to this operator.

Claim = remove `oe:todo`, add `oe:working`, leave `AGENT CLAIMED` comment, re-read the issue.

---

## Receipt vocabulary

Receipts are issue **comments** — the same vocabulary as the Linear runner. The receipt comment carries the semantic distinction that the label cannot:

- `oe:needs-input` covers **both** `AGENT BLOCKED` (missing answer belongs on the GitHub issue) **and** `AGENT HUMAN HOLD` (answer belongs in the human's own agent thread or app). Read the comment to know which.
- `AGENT UNBLOCKED` → `AGENT RESUMED` clears a blocked issue once the answer appears on the same issue.
- `AGENT HUMAN ANSWERED` → `AGENT RESUMED` clears a human hold after the operator answers in their own thread.

No new receipt tokens are introduced for GitHub — the vocabulary is tracker-agnostic.

---

## Join key

The Omnigent session label key is `openengine.issue`. The value for GitHub issues is provider-qualified:

```
github:<owner>/<repo>#<number>
```

Example: `github:Cloud-Byte-Consulting/agentic-harness#42`

Linear uses `linear:<ID>`. Jira uses `jira:<KEY>`. The key is always `openengine.issue`.

At task claim the runner stamps this value onto the Omnigent session and records the `session_id` in the `AGENT CLAIMED` comment. This ties the tracker queue, Omnigent runtime events, and OPA audit log together with no new datastore.

---

## Labels to create in the repo

Create these labels in every GitHub repo that runs Open Engine. Color is arbitrary; the names are exact.

| Label | Suggested color | Purpose |
|---|---|---|
| `agent-instructions` | `#0075ca` | Eligibility filter — runner ignores issues without this |
| `oe:todo` | `#e4e669` | Agent Todo |
| `oe:working` | `#f9d0c4` | Agent Working (claim lock) |
| `oe:review` | `#c5def5` | Agent Review |
| `oe:needs-input` | `#d93f0b` | Agent Needs Input (blocked or human hold) |
| `oe:done` | `#0e8a16` | Agent Done |

Quick setup via `gh`:

```sh
gh label create "agent-instructions" --color "0075ca" --description "Open Engine eligibility"
gh label create "oe:todo"        --color "e4e669" --description "OE: Agent Todo"
gh label create "oe:working"     --color "f9d0c4" --description "OE: Agent Working"
gh label create "oe:review"      --color "c5def5" --description "OE: Agent Review"
gh label create "oe:needs-input" --color "d93f0b" --description "OE: Agent Needs Input"
gh label create "oe:done"        --color "0e8a16" --description "OE: Agent Done"
```

---

## Projects v2 — deliberately skipped for v1

GitHub Projects v2 has a native Status field (single-select column) that maps cleanly to the five Open Engine states and is queryable via GraphQL. It would replace the `oe:*` label convention with first-class project-item status, eliminating the label-swap step.

We skip it for OE-1b because:

- Projects v2 status lives on the **project item**, not the issue. The GitHub MCP's issue tools do not surface it; GraphQL mutations are required.
- Not every repo or org has Projects v2 enabled or wants the overhead.
- Labels work today with zero GraphQL, zero project setup, and the existing MCP surface.

If throughput or query complexity makes the label approach painful, Projects v2 (Status field + `updateProjectV2ItemFieldValue` mutation) is the natural upgrade path.
