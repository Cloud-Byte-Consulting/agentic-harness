# Audit correlation design (three-plane trace)

**Status:** Design only (Lane E, Wave 0). Implemented by OE-3 / Phase 5.
**Owes:** the correlation-query spec OE-3 builds. No code in this lane.

This refines the three-plane table in
[`policy-opa-hooks-audit.md`](./policy-opa-hooks-audit.md) and the
"Receipts → three-plane audit" section of
[`open-engine-integration-plan.md`](./open-engine-integration-plan.md) into a
concrete join: which existing log carries which key, and the query that ties
one Linear issue to one correlated trace.

No new datastore. The trace is a **query over three existing surfaces** joined
by a single shared key (per the plan's "No new audit datastore/UI early" cut).

---

## 1. The join key

```
task_ref (<provider>:<id>)  ↔  session_id  ↔  subject_id
   (Open Engine)               (Omnigent)      (Entra/OPA)
```

| Key | Meaning | Stamped where |
| :--- | :--- | :--- |
| `task_ref` | the task record (`linear:ENG-123`) | tracker label `openengine.issue=<provider>:<id>`, applied at session create — already in `agentic-harness/docs/open-engine/prompts/17_queue_run_prompt_v2.md:6` |
| `session_id` | the Omnigent run | created by `POST /v1/sessions`; in Omnigent `session_id == conversation_id` (`omnigent/omnigent/server/routes/sessions.py:461`, `_publish_collaboration_mode(session_id)` → `conversation_id=session_id`) |
| `subject_id` | the human/agent identity | Entra `sub` claim via `subjectFromClaims` (`Agentic-Sentry/internal/auth/auth.go:139`) |

The label + the `session_id` stamp in the `AGENT CLAIMED` receipt are what make
the join exist — no field is added to any plane's storage. The receipt carries
the session_id as a machine-parseable HTML comment (the **receipt marker**,
specified in [`CONTRACT.md §Receipt marker`](../open-engine/CONTRACT.md)):

```
<!-- openengine session=<session_id> task_ref=<provider>:<id> phase=claimed|done -->
```

Parser regex: `<!--\s*openengine\s+(.*?)\s*-->`. Correlation is:
**resolve `task_ref` → `session_id` (from the receipt marker / session label),
then fan out to all three planes keyed by `session_id`, and attach
`subject_id` from the authz plane.**

---

## 2. The three planes — exact source + join fields

### Plane A — authz (OPA / Sentry decision logs)

| | |
| :--- | :--- |
| **MCP `tools/call`** | Sentry gateway `queryOPA` → `data.mcp.auth.decision`. Source file `Agentic-Sentry/cmd/gateway/main.go`. The session key arrives as the **`MCP-Session-Id` HTTP header** (`main.go:131-138`). |
| **Native host tools** | `opa_delegate` in `omnigent/omnigent/native_policy_hook.py` (PreToolUse) → same bundle. |
| **Join fields** | `subject_id` (from `subjectFromClaims`, `auth.go:139`, = Entra `sub`); `session_id` (the `MCP-Session-Id` header); `tool`; `verdict` (`allow`/`deny`/`require_approval` once Lane C lands the tri-state). |

**Honest gap (real work, not config):** `main.go:131-138` reads
`MCP-Session-Id` but the `ponytail:` comment there states it is **not bound to a
subject** ("MCP-Session-Id is required but not validated/bound to a subject").
So today the authz plane has `subject_id` (from the JWT) and a *raw*
`session_id` header that is **not yet asserted to belong to that subject**.
Closing this is risk 2 / Lane C's `session_id ↔ MCP-Session-Id` binding — the
correlation query is only trustworthy once that binding lands. Until then the
join is advisory (header value, attacker-spoofable), not authenticated.

### Plane B — runtime (Omnigent session events + ASK)

| | |
| :--- | :--- |
| **Lifecycle markers** | `omnigent/omnigent/session_lifecycle.py` — label/close helpers (`labels_with_closed_status`, `is_session_closed`). This file holds the **label** machinery the `openengine.issue` stamp rides on, not the event stream. |
| **Event stream** | `omnigent/omnigent/server/routes/sessions.py` — emits `session.usage` SSE (`_EXTERNAL_SESSION_USAGE_TYPE`, `sessions.py:387-388`) and ASK/elicitation events (`ElicitationRequestEvent`, `build_elicitation_request_event`, `sessions.py:131`). All keyed by `conversation_id == session_id`. |
| **Join field** | `session_id`. Every runtime event carries it as `conversation_id` (`sessions.py:461-476`). |

This plane is authoritative for the human-in-the-loop proof: ASK raised →
`AGENT HUMAN HOLD`, ASK resolved → `RESUMED`. (Per risk 4, the Omnigent runtime
plane — not the Linear receipt — is the source of truth; Linear is the human
projection.)

### Plane C — artifacts / cost (token-dashboard + teo)

| | |
| :--- | :--- |
| **Per-session cost** | Omnigent's own usage accounting: `session.usage` event (above) + cost labels namespace `cost_control.` (`omnigent/omnigent/cost_plan.py:44`), `load_session_usage` (`sessions.py:134`). This **is** keyed by `session_id`. |
| **Aggregate cost** | `token-dashboard` — `BurnRow` in `token-dashboard/lib/burn-data.ts` is keyed by **date + source** (`codex_tokens`, `claude_code_tokens`, `api_tokens`, `azure_openai_tokens`), **not by `session_id`**. |
| **Artifacts** | `teo` outputs — `teo/teo.go` validates/parses the TEO format; a TEO payload carries **no embedded session_id** (it is a content format, not a record). |

**Honest gap:** the cost join is **two-tier**. Per-session cost is recoverable
from Omnigent (`session.usage` / `cost_plan`), so the *runtime* plane already
carries cost keyed by `session_id`. `token-dashboard` is a date/source rollup
with **no `session_id` column**, and `teo` artifacts have no embedded key —
both join to the issue **by deep-link, not by field**. This matches the plan's
explicit cuts: "attach cost by link", "token-dashboard feed … deferred (no
producers/consumers yet)". So the artifacts/cost plane is:
`session_id` → Omnigent per-session usage (real join) **+** a dashboard
deep-link and a teo artifact reference stamped in the `AGENT DONE` receipt
(link-attached, not queried). Do not claim a `session_id` join into
token-dashboard until that feed exists.

---

## 3. Entra subject mapping

`subjectFromClaims(provider, claims)` (`Agentic-Sentry/internal/auth/auth.go:139`)
yields the `Subject` whose `SubjectID = claims["sub"]` — for Entra (provider
`"azure"`) this is the **Entra Object ID (OID)**, the stable `subject_id` for
correlation. The same function fills `SubjectEmail` from `email` /
`preferred_username` / `upn` (`auth.go:144-149`) and `Groups` from the `groups`
claim (`groupsFromClaims`, `auth.go:155`).

This `Subject` is what the gateway forwards to OPA, so **every authz decision
log line already carries `subject_id`** — no new identity plumbing. OE-3 binds
it to the runtime plane by ensuring the session-create call authenticates as
the same Entra subject (the `MCP-Session-Id` ↔ subject binding, §2 Plane A
gap). `subjectFromClaims` itself needs **no change** (confirmed: it already
extracts `sub`); the work is binding, not extraction.

Note the Entra **groups-overage** edge: >~200 groups omits the `groups` claim
and `groupsFromClaims` returns the `_overage` sentinel (`auth.go:163-167`) —
group-gated correlation fails closed for those subjects. `subject_id` (`sub`)
is unaffected.

---

## 4. The correlation query and CP-4

### Inputs
One `task_ref` (e.g. `linear:ENG-123`).

### Resolution order
1. `task_ref` → `session_id`: read the session whose label
   `openengine.issue=linear:ENG-123` (Plane B label machinery,
   `session_lifecycle.py`) — equivalently, parse the receipt marker
   `<!-- openengine session=<session_id> task_ref=... phase=claimed -->` from
   the issue's `AGENT CLAIMED` comment (format: `CONTRACT.md §Receipt marker`).
2. `session_id` → authz rows: filter OPA/Sentry decision logs where
   `MCP-Session-Id == session_id` (Plane A); each row carries `subject_id`,
   `tool`, `verdict`.
3. `session_id` → runtime events: filter Omnigent session/ASK events where
   `conversation_id == session_id` (Plane B).
4. `session_id` → cost/artifacts: Omnigent `session.usage` for per-session
   cost (Plane C, real join) + the dashboard deep-link / teo artifact refs from
   the `AGENT DONE` receipt (link-attached).
5. Assert `subject_id` is identical across every authz row (one actor per
   session) and attach it to the trace.

### Correlated-trace shape (illustrative)

```jsonc
{
  "task_ref": "linear:ENG-123",
  "session_id": "conv_abc123",
  "subject_id": "e1f2...-entra-oid",        // claims.sub via subjectFromClaims
  "subject_email": "alex@example.com",
  "authz": [                                 // Plane A: OPA/Sentry + opa_delegate
    { "tool": "linear.add_comment", "verdict": "allow", "ts": "..." },
    { "tool": "github.delete_repo", "verdict": "deny",  "ts": "..." }
  ],
  "runtime": [                               // Plane B: sessions.py events
    { "event": "session.created", "label": "openengine.issue=linear:ENG-123" },
    { "event": "AGENT_CLAIMED",   "session_id": "conv_abc123" },
    { "event": "ask.raised",      "verdict": "require_approval" },   // HUMAN HOLD
    { "event": "ask.resolved" },                                     // RESUMED
    { "event": "session.completed" }
  ],
  "cost": {                                  // Plane C
    "session_usage": { "tokens": 41200 },    // Omnigent cost_plan (real join)
    "dashboard_link": "https://.../burn?...",// link-attached (no session_id col)
    "teo_artifacts": ["receipt:AGENT_DONE"]  // referenced, no embedded key
  }
}
```

### CP-4 — the test this must satisfy
From the plan's checkpoint table:

> **CP-4** — complete one issue, query all three planes by `session_id`:
> a **single correlated trace** joins to `task_ref` + `subject_id`;
> satellites still build standalone.

Concretely, CP-4 passes when, for one completed `[agent
instructions][alex-claude][task]` issue:

1. exactly one `session_id` resolves from `openengine.issue=linear:ENG-123`;
2. all three planes return rows for that `session_id` (authz ≥1 decision,
   runtime ≥ claimed+done, cost ≥ per-session usage);
3. every authz row's `subject_id` equals the Entra `sub` of the claiming
   identity, and that single `subject_id` is on the trace;
4. no plane returns rows for a *different* session under that issue (one issue
   → one trace, no cross-contamination);
5. each satellite (Sentry, omnigent, token-dashboard, teo) still builds/tests
   standalone — correlation is a read-side query, it adds no source dependency.

**Pre-req for a *trustworthy* CP-4 (call it out, do not paper over):** the
authz↔runtime join is only authenticated once the `MCP-Session-Id` ↔ subject
binding (`main.go:131`, Lane C / risk 2) lands. Until then CP-4 can demonstrate
a correlated trace but on a spoofable header key — acceptable for the
single-tenant pilot, not for the enterprise compliance claim.

---

## Confirmed real files (cited above)
- `Agentic-Sentry/internal/auth/auth.go` — `subjectFromClaims` at line 139, `SubjectID` at 16.
- `Agentic-Sentry/cmd/gateway/main.go` — `MCP-Session-Id` at 131-138; `queryOPA` gateway.
- `omnigent/omnigent/session_lifecycle.py` — label/close helpers (86 lines).
- `omnigent/omnigent/server/routes/sessions.py` — `session.usage` (387-388), elicitation/ASK (131), `conversation_id==session_id` (461-476), `load_session_usage` (134).
- `omnigent/omnigent/cost_plan.py` — `COST_CONTROL_LABEL_NAMESPACE` at 44.
- `omnigent/omnigent/native_policy_hook.py` — `opa_delegate` seam (per plan/audit doc).
- `token-dashboard/lib/burn-data.ts` — `BurnRow` keyed by date+source (no session_id).
- `teo/teo.go` — TEO format validate/parse (no embedded session_id).
- `agentic-harness/docs/open-engine/prompts/17_queue_run_prompt_v2.md` — `openengine.issue` label at line 6.
