# Governed Platform Architecture — Open Engine on the AIR/Omnigent backbone

**Status:** As-built after Wave 0 (Milestone-1 + OE-0 docs landed; OE-2/OE-3 ahead).
**Scope:** the *structure* of the system being built. For *sequencing* see
[open-engine-integration-plan.md](open-engine-integration-plan.md); for the repo
consolidation see [consolidation-plan.md](consolidation-plan.md).

---

## 1. What this is, in one paragraph

A **Linear task becomes a governed agent session.** Open Engine routes work as
task records with receipts; Omnigent runs each task as a session that wraps a
host CLI (Claude Code, Codex, Cursor, …); OPA/Rego decides every tool call;
receipts, session events, and cost flow back as a three-plane audit trail tied
together by one shared key. Open Engine supplies *coordination* (who does the
next step, proof it happened); the AIR/Omnigent consolidation supplies the
*deterministic backbone* (runtime, sandbox, policy) it runs on. The boundaries
Open Engine used to merely *ask* the model to honor become **enforced** at the
tool surface.

---

## 2. The four planes

```
OPEN ENGINE    coordination / handoff  — Linear queue, task record, receipts, runner loop
OMNIGENT       runtime / session       — session wrapper, sandbox, ASK, native policy hook
air satellites contract / governance   — Sentry(OPA) authz · ARD · Cachy · teo · token-dashboard
HOST CLIs      kernel runtime          — Claude / Codex / Cursor / …
```

A task flows **top → down**: Open Engine routes it → Omnigent runs it as a
governed session → OPA decides each tool call → receipts + cost + outputs flow
back **up** as evidence.

---

## 3. Lifecycle of one task

1. **Route.** The runner (cron/paste prompt, [prompt 17](../open-engine/prompts/17_queue_run_prompt_v2.md))
   finds the oldest eligible `[agent instructions][<operator>-<agent>][task]` issue
   for this operator.
2. **Create session.** The shim `POST /v1/sessions` with label
   `openengine.issue=ENG-123`, captures `session_id`, injects it
   (`OMNIGENT_SESSION_ID`) into the runner context, attaches the Linear MCP.
3. **Claim.** Status → *Agent Working*; `AGENT CLAIMED` receipt **stamped with
   session_id** (the join key).
4. **Work, governed.** Every tool call passes the native `PreToolUse` hook →
   policy (`ALLOW | DENY | ASK`), fail-closed.
5. **Hold / block.** A runtime-permission need raises an Omnigent **ASK** →
   `AGENT HUMAN HOLD`; a missing *information* need (not a permission) → prompt
   pause → `AGENT BLOCKED` (answered on the issue).
6. **Done.** `AGENT DONE` stamps session_id (+ optional dashboard link); status →
   *Agent Done* / *Agent Review*; ledger `completed ENG-123`.

The status ledger is the **human-facing projection** of the runtime plane.

---

## 4. The join key

```
task_ref (<provider>:<id>)  ↔  session_id  ↔  subject_id
   (label openengine.issue=      (Omnigent;      (Entra OID via
   <provider>:<id>, e.g.         ==conversation_id) subjectFromClaims)
   linear:ENG-123 / github:owner/repo#123
   / jira:PROJ-45)
```

One shared key ties the task tracker, Omnigent runtime events, and OPA
decision logs — **no new datastore.** The label key stays `openengine.issue`;
the value becomes `<provider>:<id>` (the **task ref**). `session_id` is the
spine; it is stamped into the `AGENT CLAIMED` receipt and threaded into the
OPA decision input. Governance and audit are tracker-agnostic; only the
status model is provider-specific (see OE-1b below).

> **OE-0 (now):** docs adopt the provider-qualified value format.
> **OE-1b (GitHub + Jira adapters landed):** the shim stamps the `<provider>:<id>` value; GitHub runner (prompt 20) + status-label map + `openengine_stack_github.yaml`, and Jira runner (prompt 21) + status-transition map + `openengine_stack_jira.yaml` (gated transitions). Multi-tracker complete: Linear (live) + GitHub + Jira.

### 4a. Multi-tracker front-end (OE-1b)

Open Engine is a **prompt + tracker convention**, not a service. A provider is:
`{its MCP injected via the stack profile}` + `{a status-model map}` +
`{a runner-prompt variant}`. No central router; one queue-runner/cron per tracker.

The convention is provider-neutral for: eligibility (title brackets + label),
receipts (issue comments), status ledger (an issue), governance (OPA gates
**tool calls**, tracker-agnostic), and three-plane audit (keys on `session_id`).
Only the **status model** is provider-specific:

| Provider | Status mechanism | Notes |
| :--- | :--- | :--- |
| **Linear** (live) | Native workflow states | Built; `linear:ENG-123` value format |
| **GitHub Issues** (OE-1b, first) | Labels (`oe:working`, `oe:done`, …) or Projects v2 Status field (GraphQL) | GitHub MCP already in use; cheapest adapter |
| **Jira** (OE-1b, with Phase-5 enterprise push) | Native workflow transitions (gated; need transition ids + matching workflow) | Jira is the enterprise/BECU tracker |

Rollout order: Linear → GitHub Issues → Jira.
OE-1b rides **after CP-1** (prove the loop on Linear first), **parallel to
OE-2/OE-3**, and is **off the governance critical path** — OE-2 and OE-3 need
no rework, only the `openengine.issue` value format widens.

---

## 5. Governance model

### Verdicts
Policy returns one of three verdicts; the same vocabulary exists at both
enforcement layers (Omnigent Python `ALLOW|DENY|ASK` ↔ Rego
`allow|deny|require_approval`).

| Boundary | Verdict | Open Engine surface |
| :--- | :--- | :--- |
| delete / change credentials / billing (irreversible) | `deny` | `AGENT FAILED` (never silent retry) |
| publish / email / deploy (ask-first) | `require_approval` | `AGENT HUMAN HOLD` → answered → `RESUMED` |
| ordinary scoped work | `allow` | proceeds |
| missing task *information* (not a permission) | n/a (no policy) | `AGENT BLOCKED` |

### Decisions baked in
- **deny ≠ blocked.** A hard policy `deny` is `AGENT FAILED`, *not* `AGENT BLOCKED`
  (which is for missing information answerable on the issue).
- **Verdict supersedes issue prose (security).** Approval comes **only** from an
  Omnigent ASK answer at runtime — never from text in the issue body. The runner
  `<boundaries>` no longer self-grants; this closes a prompt-injection hole (a
  task body cannot say "you are approved to deploy").
- **Fail-closed (decided).** When ask-first overlaps a denial, the request
  fails closed: an RBAC-denied `publish/email/deploy` resolves to `deny`, not
  `require_approval`. `require_approval` only fires when the action otherwise
  passes RBAC.
- **`require_approval` needs a human channel.** It renders as an ASK **only when
  an Omnigent session mediates the call.** A bare Sentry MCP gateway call with
  `require_approval` fails closed (deny) — there is no one to ask. Open Engine
  always runs in-session, so this is sound.
- **Admin carve-out (decided).** Admins bypass the OE boundaries by role
  (`is_admin` against `admin_groups` in the bundle); agents and non-admin humans
  get them. **Caveat — this is role-based, not session-based:** an agent running
  under an *admin* identity inherits the bypass, so run Open Engine agents under
  a dedicated **non-admin** service identity (or later layer an
  `input.session.openengine` flag to constrain agent-admins).

### Enforcement surfaces — what is gated where
| Surface | Sees | Enforcement | Phase |
| :--- | :--- | :--- | :--- |
| MCP `tools/call` | named tool + args | Sentry gateway → OPA (`data.mcp.auth.decision`) | built (tri-state added Wave 0) |
| Native host tools (`Bash`, `Write`, …) | the tool surface | Omnigent `native_policy_hook.py` → PolicyEngine; OPA via the `opa` policy builtin (`require_approval` renders a human ASK) | M1 Python; OPA builtin built |
| Editor hosts (no Omnigent wrapper) | the tool surface | standalone `opa-hook` | Phase 3b |

**Tool-surface vs semantic.** The native hook sees `Bash("kubectl apply …")`, not
"a deploy." So at Milestone 1, boundary actions exposed as **named MCP tools**
(`delete_*`, `publish_*`) are hard-gated by name; boundary actions buried in an
**opaque shell string** are not reliably classified until OE-2's content-aware
Rego (or a shell classifier). M1 = *enforcement seam live + named-tool gating*;
full semantic coverage is OE-2.

---

## 6. LLM I/O grammar — TEO

Model-facing text uses [TEO](teo-llm-io.md) to cut tokens on both sides of the
call (assembled context in; prose artifacts out). The rule of thumb:

```
next reader = model      → TEO   (context window + inter-agent artifacts)
next reader = human       → prose (Linear ledger, receipt verbs)
next reader = tool/protocol → native JSON / JWT / Rego (never TEO)
```

Protocol boundaries keep their mandated format: MCP JSON-RPC, OPA JSON, the
Omnigent session API, function-call args, Entra JWT. No TEO↔JSON shim at tool
edges; wire-level compression stays Cachy's job.

---

## 7. Three-plane audit

| Plane | Source | Keyed by |
| :--- | :--- | :--- |
| **authz** | OPA decision logs (Sentry + `opa_delegate`) | subject_id + session_id + tool + verdict |
| **runtime** | Omnigent session events (`sessions.py`, ASK) | session_id (== conversation_id) |
| **artifacts / cost** | token-dashboard + teo outputs | session_id (cost link-attached today) |

One completed issue → one `session_id` → rows in all three planes, joining back
to `task_ref` (`<provider>:<id>`) and `subject_id`. Design + exact field map:
[audit-correlation-design.md](audit-correlation-design.md). Known gap: the
`MCP-Session-Id` header is not yet bound to the authenticated subject
(`Agentic-Sentry/cmd/gateway/main.go`), so the authz↔runtime join is spoofable
until that lands — fine for the single-tenant pilot, not the enterprise claim.

---

## 8. Component ownership (single owner per concern)

| Concern | Owner | Repo | State |
| :--- | :--- | :--- | :--- |
| Sessions, sandbox, ASK, native policy hook | **Omnigent** | `omnigent` | built |
| MCP tool authorization (deterministic) | **OPA/Rego via Sentry** | `Agentic-Sentry` | built; tri-state Wave 0 |
| Native-tool authorization | **`opa` policy builtin** (PolicyEngine → OPA) | `omnigent` | **built, default-off** — server-side builtin; `require_approval` renders ASK; `shadow`→`enforce` rollout; live enforcement at CP-3 |
| Token-plane proxy / compression | **Cachy** | `Cachy` | built |
| Token-efficient output grammar | **TEO** | `teo` | built |
| Cost / activity measurement | **token-dashboard** | `token-dashboard` | built |
| Federated discovery | **ARD** | `Agentic-Resource-Discovery` | partial (spec + CLI) |
| Personas, pods, skills, queue contract | **agentic-harness** | `agentic-harness` | content + docs |
| Task queue / handoff / receipts | **Open Engine** | `agentic-harness/docs/open-engine` | prompt + tracker-neutral contract (Linear live; GitHub/Jira via OE-1b) |
| Orchestration (judge / hill-climb) | **Omnigent supervisor YAML** (ported from role-router) | `omnigent` | Phase 2 |
| `air` bootstrap / session / install | → absorbed into Omnigent stack profile | — | shrinking |
| `aos` | archive after doc merge | `aos` | docs only |

**Independence rule:** every satellite builds and tests standalone; the only
sanctioned source dependency is `air → teo`.

---

## 9. Harness resolution

Open Engine agent codes are `<operator>-<vendor>` (e.g. `alex-claude`). The
suffix resolves to a **native CLI harness**, because an Open Engine agent is a
persistent vendor CLI claiming tasks — not an in-process SDK call:

| suffix | harness | | suffix | harness |
| :-- | :-- | :-- | :-- | :-- |
| `claude` | **`claude-native`** | | `kiro` | `kiro-native` |
| `codex` | `codex-native` | | `cursor` | `claude-sdk` |
| `copilot` | `copilot` | | `gemini` / `windsurf` | `claude-sdk` |

`claude → claude-native` (**decided**) deliberately overrides the
`HARNESS_ALIASES["claude"] = "claude-sdk"` shorthand, which is meant for SDK
orchestration, not an interactive agent session. The map lives in
`omnigent/omnigent/harness_aliases.py` (`OPENENGINE_AGENT_HARNESS`) and mirrors
`agent_code_harness_map` in `omnigent/profiles/openengine_stack.yaml`; the two
collapse to one source once `sessions.py` reads the profile (TODO).

---

## 10. Wave-0 as-built vs ahead

**Landed (uncommitted/feature-branch):** runner retargeted to Omnigent sessions;
session_id stamping + shim; `openengine.issue` label; security boundary fix;
`subject_id`, receipt vocab, CONTRACT; tri-state Rego (`verdict`) with boolean
back-compat + parity spike (40/40 `opa test`); gateway parses `verdict`; TEO
carve-out + runner `<io_format>` + agent-profile stanza; audit-correlation
design; Phase-0 deprecation notices; Phase-1 stack-profile **stub**.

**`opa_delegate` (OE-2 core, landed default-off):** OPA delegation runs as a
**server-side PolicyEngine builtin** — `omnigent/omnigent/policies/builtins/opa.py`
(`opa_require_approval`) — evaluated on every `tool_call` event. It queries the
bundle's native-plane **`oe_decision`** rule (default-allow + the OE boundaries
only, so host tools like `Bash` aren't blanket-denied by the gateway's MCP-RBAC
default-deny) and returns `DENY` / `ASK` / `ALLOW`. The OPA client lives in
`omnigent/omnigent/opa_delegate.py`. Modes `off`(default)→`shadow`(log-only)→
`enforce` (`OMNIGENT_OPA_DELEGATE_MODE`); OPA-unreachable in enforce fails closed
to DENY. Attached via the stack-profile `guardrails.policies` (`opa_oe_boundaries`).
**Native-plane `require_approval` now renders a human ASK** — the builtin returns
`ASK`, which the *existing* server elicitation gate (`evaluate_policy` →
`_hold_native_ask_gate`) holds + prompts + resumes, exactly like a Python ASK. The
earlier hook-side branch (which collapsed `require_approval` to deny) was removed;
deny-wins falls out of the engine's policy composition (DENY short-circuits), so no
`combine_actions` is needed. A governed CLI agent can now request runtime
permission interactively instead of failing closed.

> **Activation (Lane B loader landed):** a session opts into a stack profile via
> the **`openengine.profile=<name>`** label on `POST /v1/sessions`. The create
> handler (`omnigent/omnigent/server/profiles.py`) reads
> `profiles/<name>.yaml`, parses its `guardrails.policies`, and writes them as
> **session policies** (keyed by `conversation_id`); the lazily-built PolicyEngine
> loads session policies first, so `opa_oe_boundaries` goes live with no
> engine-build change. `enforce` is thus AND-gated on (`OMNIGENT_OPA_DELEGATE_MODE`)
> AND (the `openengine.profile` label). Path-safe (`[A-Za-z0-9_-]` only, confined
> to the profiles dir); best-effort (a bad label is logged, never fatal). The
> profile name still needs to be set by whoever creates the session (the runner /
> shim).

**Ahead:** supervisor-YAML orchestration (Phase 2); `air` shrink + `aos` archive
(Phase 4); Entra + correlation query (Phase 5). Gates CP-1…CP-4 are **manual**.

***OE-3 is complete.*** Native-plane groups (the admin carve-out): the subject's
Entra `groups` are plumbed from the verified id_token → session JWT → policy event
→ the OPA builtin, fail-safe to `[]` (strict); admin-group revocation lags ≤ the
session TTL (~8h). Audit binding: the Sentry gateway now **binds `MCP-Session-Id`
to the authenticated subject** (trust-on-first-use, `cmd/gateway/main.go`) and
rejects (403) a session id replayed under a different subject — so the three-plane
correlation join can't be spoofed. **Deployment note (per the cloud agent-identity
model — Entra Agent ID / AWS AgentCore Identity / GCP Agent Identity):** run each
OE agent-code under its own dedicated **non-admin agent identity**, so its groups
(and the admin carve-out) attach to the agent, not the human operator.

---

## 11. Open decisions

- **`opa_delegate` Rego parity** (critical path; de-risked by the Wave-0 parity
  spike, not yet wired into the native hook).
- **Omnigent fork vs upstream** — `omnigent` remote is `omnigent-ai/omnigent`;
  consolidation patches need a fork before any push.
- **`session_id`↔`MCP-Session-Id` binding** — header choice + subject binding.
- **OE-1b multi-tracker front-end** — GitHub Issues adapter first (MCP already
  in use; labels-as-status cheapest path); Jira with Phase-5 enterprise push.
  Blocked on CP-1 (prove single-tracker loop). Code change: widen
  `openengine.issue` value format in `openengine_session_shim.py` and stack
  profile; docs already provider-neutral (OE-0 done).

---

## 12. Detailed references

- [open-engine-integration-plan.md](open-engine-integration-plan.md) — sequencing, lanes, checkpoints
- [consolidation-plan.md](consolidation-plan.md) — single-owner matrix, phases
- [policy-opa-hooks-audit.md](policy-opa-hooks-audit.md) — authz split, editor hooks
- [audit-correlation-design.md](audit-correlation-design.md) — three-plane query
- [teo-llm-io.md](teo-llm-io.md) — TEO carve-out
- [../open-engine/CONTRACT.md](../open-engine/CONTRACT.md) — Open Engine ↔ Omnigent contract
