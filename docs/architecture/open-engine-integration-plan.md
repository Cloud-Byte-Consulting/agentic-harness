# Plan: Implement AIR Consolidation + Open Engine as one governed platform

## Context

Two efforts exist as separate docs in `agentic-harness/docs/`:

1. **AIR + Omnigent consolidation** (`docs/architecture/`) — collapse ~10 federated repos into one governed platform with a single owner per concern: Omnigent owns the runtime, OPA/Rego owns all security allow/deny, satellites (Sentry, ARD, Cachy, teo, token-dashboard) sit at protocol boundaries only, `air`/`aos`/`role-router` shrink or archive. Phases 0–5 (+3b) are already defined.
2. **Open Engine** (`docs/open-engine/`) — a Linear-based shared task queue so work handoffs between agents (Codex, Claude, Cursor) become task records with receipts, not human copy-paste. Today it is **documentation only** (1 main guide, 19 setup prompts, 16 templates, 29 skill cards) with **zero connection** to Omnigent/Sentry/OPA.

**The problem:** Open Engine describes *who does the next step and proves it happened* but enforces nothing — its `<boundaries>` ("never publish, deploy, delete…") are advisory prose the LLM is *asked* to honor. The consolidation builds deterministic governance (OPA) but has no task/handoff layer on top. Neither doc references the other.

**Intended outcome:** one plan where a Linear task becomes a **governed Omnigent session** — OPA-gated tools, ASK-driven human holds, receipts and status ledger feeding the three-plane audit trail. Open Engine becomes Omnigent's task-queue front-end; the consolidation is the backbone it runs on.

**Ground truth (verified this session):** `omnigent` (506 py files, sessions + `native_policy_hook.py` + 10 harness bridges + sandbox + ASK), `Agentic-Sentry` (live `queryOPA` gateway + Entra), `Cachy`, `agentic-harness` (air CLI), `copilot-role-router`, `teo`, `token-dashboard` are all **built and production-ready**. `ARD` is spec + conformance CLI (partial). `aos` and `open-engine`/`open-engine-data` are **docs only**. Repos are individual git repos; `go.work` = `agentic-harness/air` + `teo`.

## Strategy (decided with user)

| Decision | Choice |
| :--- | :--- |
| Sequencing | **Open Engine first, consolidation underneath** — ship the queue now on existing agents; consolidation runs as the parallel backbone; converge on the governed bridge |
| Integration depth | **Full governed bridge** (Linear task → Omnigent session → OPA → three-plane audit) as the end-state |
| First proof loop | **A dev loop** (`[agent instructions][alex-claude][task]` → Omnigent session → review back to Linear) |
| Runner automation | **Manual copy-paste now → Omnigent-native later** (cron shim as interim; no daemon until enterprise) |

**Key reframing:** Open Engine is a prompt + Linear-convention contract, not a service. "Combining" it with the stack is mostly *point the existing queue-runner prompt at the `omnigent` CLI* and let the session govern. The only genuinely new engineering is `opa_delegate` — already planned in consolidation Phase 3. Everything else is prompt + config + one Linear label.

## Architecture: the three planes

```
OPEN ENGINE   coordination/handoff  — Linear queue, task record, receipts, runner loop
OMNIGENT      runtime/session       — session wrapper, sandbox, ASK, native policy hook
air satellites contract/governance  — Sentry(OPA) authz · ARD · Cachy · teo · token-dashboard
HOST CLIs     kernel runtime        — Claude / Codex / Cursor / ...
```

A task flows top→down: Open Engine routes it → Omnigent runs it as a governed session → OPA decides every tool call → receipts + cost + outputs flow back up as evidence.

### The join key
`task_ref (<provider>:<id>) ↔ session_id ↔ subject_id`. At claim, stamp the Omnigent session with label `openengine.issue=<provider>:<id>` (e.g. `linear:ENG-123`, `github:owner/repo#123`, `jira:PROJ-45`), and stamp `session_id` into the `AGENT CLAIMED` receipt. That single shared key ties the tracker, runtime events, and OPA decision logs together — no new datastore. The label KEY `openengine.issue` is fixed (already wired in the stack profile and `sessions.py`); only the VALUE becomes provider-qualified. Linear is live today; GitHub and Jira arrive via OE-1b.

### Governance mapping (advisory prose → deterministic OPA)

| Open Engine boundary | Rego verdict | Enforced at | Open Engine surface |
| :--- | :--- | :--- | :--- |
| delete / change credentials / billing (irreversible) | `deny` | native hook→`opa_delegate` + Sentry (MCP) | `AGENT FAILED` (never silent retry) |
| publish / email / deploy (ask-first) | `require_approval` | same two points; Omnigent renders ASK | `AGENT HUMAN HOLD` → `HUMAN ANSWERED` → `RESUMED` |
| ordinary scoped work | `allow` | both | proceeds |
| missing task *information* (not a permission) | n/a (no OPA) | prompt-level pause | `AGENT BLOCKED` (answered on the issue) |

Rego decides `allow | deny | require_approval`; Omnigent owns the ASK UX. Same bundle Sentry already queries (`/v1/data/mcp/auth/decision`), so MCP tools and native tools hit identical policy. **New rule to document:** a hard OPA `deny` ≠ `AGENT BLOCKED` (it is `AGENT FAILED`, or `AGENT HUMAN HOLD` if the fix is a human authority change).

### Receipts → three-plane audit

| Plane | Existing source | Keyed by |
| :--- | :--- | :--- |
| authz | OPA logs (Sentry + native hook→opa_delegate) | subject_id + session_id + tool + verdict |
| runtime | Omnigent session events (`session_lifecycle.py`, phases, ASK) | session_id |
| artifacts/cost | token-dashboard logs + teo outputs | session_id |

`AGENT CLAIMED`→runtime start (carries session_id, the join); `AGENT DONE`→complete + teo artifact + dashboard deep-link; `BLOCKED`→runtime pause; `HUMAN HOLD/ANSWERED`→ASK pending/resolved; `FAILED`→error or OPA deny. The status ledger is the human-facing projection of the runtime plane.

### LLM I/O format: TEO (token savings, with a tool-format carve-out)

All **model-facing** text uses TEO (`teo/`) to cut tokens on *both* sides of the LLM call:
- **Input** — the context assembled for the model (task record, status-ledger state, retrieved context, standing skills) is rendered as dense TEO instead of verbose prose/JSON.
- **Output** — the agent emits its freeform results (reviews, summaries, `AGENT DONE` evidence) as TEO; the `teo` library validates/parses them.

**Carve-out — respect native formats at every protocol boundary.** TEO never wraps a payload a tool or protocol mandates:

| Surface | Format (NOT TEO) | Why |
| :--- | :--- | :--- |
| MCP `tools/call` (Linear MCP, Sentry gateway) | JSON-RPC | the MCP server's contract |
| OPA decision input/output | JSON (`data.mcp.auth.decision`) | Rego input schema |
| Omnigent session API | JSON (`POST /v1/sessions`) | server contract |
| Tool / function-call arguments | harness-native JSON schema | the model's tool interface |
| Entra claims | JWT | identity standard |

**Rule of thumb:** TEO on the context window and the agent's prose artifacts; native schema on the wire to any tool. Implementation is **prompt + library, not new infra** — instruct TEO I/O in the queue-runner prompt and agent profile; `teo` validates output; *(optional, API-key mode only)* Cachy adds wire compression on the LLM stream. This elevates the consolidation's existing `teo` owner role into the default I/O grammar rather than just an `AGENT DONE` format.

## Phased plan (Open Engine sub-phases anchored to consolidation phases)

| OE phase | Rides | Deliverable |
| :--- | :--- | :--- |
| **OE-0** | Phase 0 | Docs: declare Open Engine = Omnigent's task-queue front-end contract; fold `open-engine` docs into the architecture set; add `session_id` stamp to receipt templates; add `subject_id` column to routing map; add "deny ≠ blocked" rule to receipt vocab. Generalize the join-key label value to a provider-qualified task ref: `openengine.issue=<provider>:<id>` (e.g. `linear:ENG-123`, `github:owner/repo#123`, `jira:PROJ-45`) so the queue is tracker-neutral from day one. *Docs only.* |
| **OE-1** | Phase 1 + 3b | **Task→Session bridge.** Queue-runner prompt targets `omnigent` (not bare CLI); agent-code→Omnigent harness/profile map; Linear MCP injected into the session; `openengine.issue` label stamped at create; native `PreToolUse` hook enforces (existing Python policy). **← Milestone 1.** |
| **OE-1b** | Phase 1 (after CP-1) | **Multi-tracker front-end.** Provider adapter = the tracker MCP injected via the stack profile + a status-model map + a runner-prompt variant. No central router; one queue-runner/cron per tracker. GitHub Issues first (MCP already in use; labels-as-status cheapest adapter), then Jira alongside OE-3 (enterprise push). OE-2/OE-3 need no rework — governance gates tool calls and audit keys on `session_id`; only the join-key field value widens. **GitHub adapter landed:** `github-status-labels.md`, runner prompt `20_queue_run_prompt_github.md`, `claude-code-github-mcp-setup.md`, `openengine_stack_github.yaml`, shim provider-qualified value. **Jira adapter landed:** `jira-status-transitions.md`, runner prompt `21_queue_run_prompt_jira.md`, `claude-code-jira-mcp-setup.md`, `openengine_stack_jira.yaml` (gated transitions, no shim change). **OE-1b multi-tracker complete: Linear (live) + GitHub + Jira.** |
| **OE-2** | Phase 3 (+2) | **Governance determinism.** Open Engine boundaries → shared Rego bundle; `opa_delegate` makes the native hook evaluate OPA (same bundle as Sentry); ask-first → ASK = `AGENT HUMAN HOLD`. ASK rendering of `require_approval` rides Phase 2 supervisor YAML (after role-router Judge/hill-climb is ported). |
| **OE-3** | Phase 5 | **Enterprise.** Entra `subject_id` binding; propagate `session_id` to Sentry/OPA logs; three-plane correlation query; teo+cost rollup per issue; BECU pilot. |

### Provider status models (OE-1b)

Eligibility, receipts, and governance are provider-neutral. Only the STATUS MODEL is provider-specific:

| Provider | Status mechanism | Notes |
| :--- | :--- | :--- |
| **Linear** | Native workflow states (e.g. Agent Todo → Agent Working → Agent Done) | Live today. |
| **GitHub Issues** | `open`/`closed` only → status via labels (`oe:working`, `oe:done`, …) or Projects v2 `Status` field (GraphQL) | Cheapest adapter; GitHub MCP already in use. |
| **Jira** | Native workflow transitions (gated; requires transition IDs + a matching workflow configured per project) | Arrives with OE-3 enterprise push. |

This does **not** replace the existing consolidation phases — it threads onto them. The consolidation backbone (absorb air bootstrap → Phase 1; port role-router → Phase 2; opa_delegate + Sentry same bundle → Phase 3; opa-hook + agy 2.0 verify → Phase 3b; shrink air + archive aos → Phase 4; Entra + audit pipeline → Phase 5) proceeds as already planned; Open Engine consumes each phase's output.

### Parallel execution

The repos are independent by design (each builds and tests standalone — the only sanctioned source dependency is `air → teo`), so most work runs concurrently across repos. **Five streams can start at once**; only **two sync points** gate the rest.

**Wave 0 — all start immediately, no cross-stream dependencies (different repos/files, assignable to different agents/people now):**

| Lane | Repo / area | Work | Owner |
| :--- | :--- | :--- | :--- |
| **A — OE docs & conventions** | `agentic-harness/docs` | OE-0: contract declaration; `session_id` in receipt templates; `subject_id` column in routing map; "deny ≠ blocked" rule; retarget queue-runner prompt to `omnigent` | docs |
| **B — Omnigent bring-up** | `omnigent` | `harness_aliases.py` agent-code→harness map; `openengine.issue` label at session create (`sessions.py`) | Omnigent |
| **C — Authz / Rego** | `Agentic-Sentry` | author shared Rego bundle (OE boundary → `allow`/`deny`/`require_approval`) + `opa test`; `session_id` ↔ `MCP-Session-Id` header binding (`main.go`) | Sentry |
| **D — Consolidation P0 → P1** | `omnigent` + `agentic-harness` | Phase 0 declarative (ownership matrix, deprecation notices, merge `aos` docs); Phase 1 Omnigent stack profile absorbing `air bootstrap`/MCP-inject | Omnigent |
| **E — Audit design** | `docs` | spec the three-plane correlation query + Entra subject mapping — **design only, no code** | docs |
| **F — TEO I/O** | `agentic-harness/docs` + `teo` | instruct TEO for model input/output in the runner prompt + agent profile; wire `teo` validate/parse; document the tool-format carve-out | docs + teo |

**Sync point 1 → Milestone 1 (manual governed loop):** needs only **Lane A + Lane B**. Independent of C, D, E. Ship as soon as both land — the existing Python policy hook enforces; no OPA / Entra / absorption required.

**Sync point 2 → OE-2 (deterministic governance):** needs **Lane C** (Rego bundle) + the **`opa_delegate`** change in Lane B (`native_policy_hook.py` → OPA). Phase 2 (port role-router → supervisor YAML) is its own parallel track and gates only the ASK-rendering *polish* of OE-2, not its core.

**After the sync points (can overlap):**
- **Phase 4** (shrink `air`, archive `aos`) — needs Phase 1 (Lane D); otherwise independent, can land late.
- **OE-3 / Phase 5** (Entra binding, `session_id`→OPA logs, correlation query) — needs B + C landed; its design (Lane E) is already done in Wave 0.

**Dependency graph (what blocks what):**

```
Lane A ─┐
        ├─► Milestone 1 ─[CP-1]─┐
Lane B ─┘                       │
Lane F (TEO I/O) ──────[CP-2]───┤
                                ├─► OE-2 ─[CP-3]─► OE-3 / Phase 5 ─[CP-4]
Lane C ─► opa_delegate ─────────┘
Lane D (P0→P1) ─► Phase 4
Phase 2 (role-router→YAML) ─────┘  (ASK-rendering polish only)
Lane E (audit design) ──────────┘
```

**Critical path:** Lane C (Rego) → `opa_delegate` → OE-2 → Phase 5. Everything else has slack. **Milestone 1 is off the critical path** — it lands early on the short A + B path while C, D, E, F proceed. `[CP-n]` are human test gates (see Checkpoints).

**OE-1b (multi-tracker front-end) is parallel to OE-2/OE-3 and OFF the governance critical path.** It starts after CP-1 (prove the loop on one tracker first) and can run alongside OE-2 and OE-3. OE-2 (governance) and OE-3 (audit) need no rework: governance gates tool calls (tracker-agnostic) and audit keys on `session_id` — only the join-key field value widens from `ENG-123` to `<provider>:<id>`.

**Dogfood it:** the six Wave-0 lanes map 1:1 onto Open Engine's own routing map — one Linear task per lane, assigned per owner. The plan's own coordination layer schedules the plan's own work.

## Checkpoints (human test gates)

Each checkpoint is a **stop-and-test gate**: you exercise the functionality by hand, and any regression is **fixed before the next wave starts**. No advancing past a red checkpoint — a failed gate sends the lane back to its owner before any downstream lane consumes its output.

The table below is the summary; the **runnable per-gate tests** (a handful each, with defined input / expected output / usage check) live in [checkpoint-manual-tests.md](checkpoint-manual-tests.md). A gate is green only when every test in its set passes.

| Gate | After | You test | Pass criteria |
| :--- | :--- | :--- | :--- |
| **CP-1** | Milestone 1 (A + B) | run the dev loop: claim→done, blocked-resume, human-hold smoke issues (`templates/*-test.md`) | receipts + statuses + ledger correct; native hook enforces boundaries; session labeled `openengine.issue` |
| **CP-2** | TEO I/O (F) | run a task with TEO I/O on; compare token usage vs prose baseline (token-dashboard); confirm outputs parse and tool calls still use JSON | measurable token drop; `teo` round-trips clean; **zero** tool-call format regressions |
| **CP-3** | OE-2 (C + `opa_delegate`) | attempt a `deny` tool and an `ask-first` tool from a task | deny → `AGENT FAILED` (fail-closed, not silent retry); ask-first → Omnigent ASK → `AGENT HUMAN HOLD`; Sentry (MCP) + native return the **same** verdict; `opa test` green |
| **CP-4** | OE-3 / Phase 5 | complete one issue, query all three planes by `session_id` | single correlated trace joins to `task_ref` + `subject_id`; satellites still build standalone |

## Milestone 1 — smallest viable governed loop (the dev loop)

A Linear `[agent instructions][alex-claude][task]` issue, end to end:

1. cron (or manual paste) runs the queue-runner prompt inside an `omnigent` claude-native session with the Linear MCP attached;
2. session created bound to alex-claude's persistent `agent_id`, labeled `openengine.issue=ENG-123`;
3. claim: status→Agent Working, `AGENT CLAIMED` **stamped with session_id**;
4. every tool call passes the native `PreToolUse` hook → Omnigent policy, fail-closed — boundaries are now **enforced, not advised**;
5. a runtime-permission need raises Omnigent ASK → `AGENT HUMAN HOLD` + ledger `holding ENG-123`;
6. `AGENT DONE` stamps session_id (+ optional dashboard link); status→Agent Done; ledger `completed`.

**New work for Milestone 1 is small:** the agent-code→profile map and pointing the hook target at the session (config-scale). `opa_delegate` and Entra are explicitly **not** required here — the existing Python policy hook governs the first loop; Rego parity is the OE-2 upgrade.

## Critical files

Reuse existing rails — do not build new services.

- `omnigent/omnigent/native_policy_hook.py` — the `opa_delegate` seam; where boundaries become deterministic (OE-2). Today POSTs to Omnigent Python policy registry; Phase 3 makes it delegate to OPA.
- `omnigent/omnigent/server/routes/sessions.py` — `POST /v1/sessions` create + `openengine.issue` label + ASK/elicitation `inputRequests` (Milestone 1, Seam HOLD).
- `omnigent/omnigent/session_lifecycle.py` — runtime-plane events + label flow.
- `omnigent/omnigent/harness_aliases.py` — add agent-code suffix → harness mapping (`*-claude`→claude bridge, etc.).
- `Agentic-Sentry/cmd/gateway/main.go` — `queryOPA` (built); add Open Engine boundary rules to the shared Rego bundle; bind `session_id` via `MCP-Session-Id` header (existing ponytail TODO at `main.go:131`) for connector-MCP authz correlation.
- `Agentic-Sentry/mcp-policies/*.rego` — the shared bundle: add the Open Engine boundary → verdict rules (OE-2).
- `Agentic-Sentry/internal/auth/auth.go` — `subjectFromClaims` (built, no change) yields Entra subject for OE-3.
- `agentic-harness/air/cmd/bootstrap.go` — MCP-inject the Linear MCP + agent-code→harness wiring (Phase 1 / per `plan.md` Step B). This logic moves into the Omnigent stack profile when air bootstrap is absorbed.
- `agentic-harness/docs/open-engine/prompts/17_queue_run_prompt_v2.md` — retarget runner to `omnigent`; add the `session_id` stamp + `openengine.issue` label instructions.
- `agentic-harness/docs/open-engine/templates/starter-private-context-file.md` — per-operator agent-code→profile + `subject_id`.
- `agentic-harness/docs/open-engine/prompts/15_agent_routing_map.md` — add `subject_id` (Entra OID) column.
- `agentic-harness/docs/open-engine/open-engine.md` §13 / receipt vocab — add the "deny ≠ blocked" rule.
- `teo/` — output validate/parse + input rendering for TEO LLM I/O (Lane F); instructed from the runner prompt + agent profile, no new infra.

## What NOT to build (ponytail / YAGNI)

- No Omnigent-native scheduled runner daemon until enterprise multi-tenant — cron + copy-paste prompt + `omnigent` CLI covers personal/team. The runner is a prompt, not a service.
- No Linear→Omnigent webhook bridge — polling via the runner is correct at low volume.
- Don't parse issue `<boundaries>` grants into Rego — fire ASK at runtime; the human grant *is* the ASK answer (`AGENT HUMAN HOLD`).
- No new audit datastore/UI early — three existing surfaces + shared `session_id` are the pipeline; build a unified query only for the BECU/Phase-5 pilot.
- Don't auto-create an Omnigent agent per task — reuse persistent per-agent-code profiles.
- No receipt-emitter code — receipts are MCP comment writes the agent already makes; only *add* the `session_id` stamp instruction.
- **Don't TEO-encode tool/protocol payloads** — TEO is for the model's context and prose artifacts only; MCP/OPA/session/JWT/function-call payloads keep their mandated schema. No TEO↔JSON shim at tool boundaries.
- No bespoke TEO compressor service — TEO is prompt + `teo` library; wire-level compression is Cachy's job (API-key mode, already deferred). Don't duplicate it.
- Honor `plan.md` cuts: token-dashboard "feed" and ARD runtime resolution stay deferred (no producers/consumers yet) — attach cost by link; hardcode gateway URL.

## Open decisions / risks

1. **opa_delegate Rego parity.** Omnigent's Python policy builtins (`omnigent/policies/builtins`) must be expressible in the shared Rego bundle, else native and MCP planes diverge. Decision: port builtins → Rego, or keep Python as a thin pre-filter with Rego as authority?
2. **session_id ↔ MCP-Session-Id binding** (`Agentic-Sentry/cmd/gateway/main.go:131` TODO) — needed for connector-MCP authz correlation; native-tool authz already correlates. Decide header (`MCP-Session-Id` reuse vs `X-Omnigent-Session`).
3. **Headless cron + HOLD.** An unattended runner needs a non-interactive Entra credential; `AGENT HUMAN HOLD` needs a human channel. Resolution likely: HOLD pauses, human answers async in the Omnigent thread, next cron tick resumes (fits the existing loop) — confirm.
4. **Audit source of truth.** Receipts can drift from session events. Recommend: Omnigent runtime plane authoritative; Linear is the human projection.
5. **agent-code→profile location.** Per-operator private SKILL.md vs a central map — drift risk if both. Pick one.
6. Carried from consolidation: Omnigent fork vs upstream; role-router archive timing; agy 2.0 PreToolUse verification (gates Phase 3b editor tier); BECU first workload choice.

## Verification

Each bullet is gated by the matching checkpoint (CP-1..CP-4) — a green automated check **and** your hands-on confirmation before the next wave proceeds.

- **OE-0:** docs build/links resolve; receipt + routing-map templates contain `session_id` / `subject_id`; "deny ≠ blocked" rule present in `open-engine.md`.
- **Milestone 1 (OE-1):** create a real `[agent instructions][alex-claude][task]` smoke issue (reuse `templates/basic-smoke-test-issue-body.md`). Run the retargeted runner prompt in an `omnigent` session. Assert: session created with `openengine.issue` label; `AGENT CLAIMED` carries a session_id; status moves Working→Done; ledger shows `completed ENG-123`. Then the **blocked-resume** test (`templates/blocked-resume-test.md`) and **human-hold** test (`templates/human-hold-test.md`) — assert `AGENT BLOCKED`/`HUMAN HOLD` map to a prompt pause vs Omnigent ASK respectively.
- **Governance (OE-2):** add a `deny` Rego rule for a destructive tool; run a task that attempts it; assert the native hook fail-closes and the runner emits `AGENT FAILED` (not `AGENT BLOCKED`). Add a `require_approval` rule for `publish`; assert it raises Omnigent ASK → `AGENT HUMAN HOLD`. Confirm Sentry (MCP) and the native hook return the same verdict from the same bundle (`opa test` green).
- **TEO I/O (CP-2):** run one task with TEO I/O on; token-dashboard shows a measurable drop vs the prose baseline; `teo` parse/validate round-trips the agent output; tool calls still emit native JSON (no format regression).
- **Audit (OE-3):** for one completed issue, query all three planes by `session_id` and confirm a single correlated trace (OPA decisions + session events + teo/cost) joins back to `task_ref` and `subject_id`.
- **Consolidation regression:** each satellite still builds/tests standalone (independence rule); `air → teo` remains the only sanctioned source dependency.
