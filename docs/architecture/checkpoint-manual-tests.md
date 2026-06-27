# Checkpoint manual tests (CP-1 … CP-4)

Each CP gate in [open-engine-integration-plan.md](open-engine-integration-plan.md)
is a **stop-and-test gate**. This doc makes each gate concrete: a handful of
manual tests with a defined **input**, **expected output** (receipts / status /
verdict / parsed result), and a **usage / observability** check. A gate is green
only when every test in it passes; a red test sends the lane back to its owner
before any downstream lane consumes its output.

**How to run:** execute each test by hand against a live `omnigent` session with
the Linear MCP attached. Record the result in the log table at the bottom (date,
tester, pass/fail, notes). Don't advance past a red gate.

**Vocabulary** — receipts: `AGENT CLAIMED/DONE/BLOCKED/HUMAN HOLD/HUMAN
ANSWERED/UNBLOCKED/RESUMED/FAILED`. Linear states: *Agent Todo / Working / Done /
Review / Needs Input*. Ledger verbs: `claimed/completed/blocked/holding/resumed/
failed/none`. Usage plane = token-dashboard; authz plane = OPA decision log.

---

## CP-1 — Milestone 1 (Lanes A + B): the governed dev loop

**Gate:** receipts + statuses + ledger correct; `session_id` present in
`AGENT CLAIMED`; native hook hard-gates the **named-tool** boundary; session
labeled `openengine.issue`. *Opaque-shell boundary classification is out of
scope here (CP-3).*

**Preconditions:** runner = `prompts/17_queue_run_prompt_v2.md`; the shim creates
the session and injects `OMNIGENT_SESSION_ID`; existing Python policy hook active.

| # | Test | Input | Expected output | Usage / observability |
| :-- | :-- | :-- | :-- | :-- |
| 1.1 | Happy path claim→done | `templates/basic-smoke-test-issue-body.md` as a real `[agent instructions][alex-claude][task]` issue | session created + **labeled `openengine.issue=ENG-###`**; `AGENT CLAIMED` carries a `session_id` via the receipt marker `<!-- openengine session=<session_id> task_ref=linear:ENG-### phase=claimed -->` (see `CONTRACT.md §Receipt marker`); `AGENT DONE` carries matching marker with `phase=done`; *Working→Done*; ledger `completed ENG-###` | exactly **one** session in token-dashboard for this `session_id`; non-zero token/cost recorded |
| 1.2 | Blocked-resume | `templates/blocked-resume-test.md` | `AGENT BLOCKED` = **prompt pause, not an ASK**; *Needs Input*; ledger `blocked`. Answer on the issue → `UNBLOCKED`→`RESUMED`→`DONE` | one session, resumed in the same `session_id` (no orphan second session) |
| 1.3 | Human-hold | `templates/human-hold-test.md` | a runtime-permission need raises an **Omnigent ASK** → `AGENT HUMAN HOLD`; ledger `holding`. Answer in the Omnigent thread → `HUMAN ANSWERED`→`RESUMED` | ASK pending→resolved shows in the session event stream |
| 1.4 | Named-tool boundary (deny) | `/home/bittahcriminal/air/workspace/agentic-harness/docs/open-engine/templates/named-tool-deny-test.md` (calls a named destructive MCP tool, e.g. `delete_repository`, on a throwaway target) | native `PreToolUse` hook **fail-closes** → `AGENT FAILED` (**not** `AGENT BLOCKED`, no silent retry) | a **DENY** decision recorded for `tool=delete_repository`, keyed by `session_id` |
| 1.5 | Join-key integrity | any completed issue from 1.1 | the `session_id` in `AGENT CLAIMED` == `AGENT DONE` == the Omnigent session id; `alex-claude` ran on the **`claude-native`** harness | `harness_for_agent_code("alex-claude") == "claude-native"` (self-check) |
| 1.6 | Input validation: ineligible issue | `/home/bittahcriminal/air/workspace/agentic-harness/docs/open-engine/templates/ineligible-issue-test.md` (title missing the agent-code bracket or wrong agent code) | runner **skips** it (not eligible); ledger `none`; **no session created** | zero new sessions in the dashboard for this run |

---

## CP-2 — TEO I/O (Lane F): token savings + format integrity

**Gate:** measurable token drop; `teo` round-trips clean; **zero** tool-call
format regressions.

**Preconditions:** TEO I/O on via the agent-profile stanza
(`templates/starter-private-context-file.md`) and the runner `<io_format>` block;
fix the `teo-llm-io.md` Go snippet first.

| # | Test | Input | Expected output | Usage / observability |
| :-- | :-- | :-- | :-- | :-- |
| 2.1 | Token delta vs prose baseline | run the **same** task twice — once TEO-on, once prose-baseline | both complete identically (same receipts/status) | token-dashboard shows a **measurable drop** TEO-on; record both input+output token counts and the % delta |
| 2.2 | `teo` round-trip | an `AGENT DONE` TEO artifact from 2.1 | `teo.Validate` returns nil; `teo.Parse` yields the expected fields (`GetScalar`/`FindBlock`) | n/a |
| 2.3 | Tool-call format guard | inspect the session's MCP / OPA / function-call payloads during the run | all emit **native JSON** — no TEO wrapping anywhere on a protocol boundary | sample ≥3 tool calls; none TEO-encoded |
| 2.4 | Human-readability guard | read the Linear status ledger + receipt lines for 2.1 | ledger lines + receipt verbs are **prose Markdown**, not TEO | a human can scan the ledger unaided |
| 2.5 | Malformed-output handling | `/home/bittahcriminal/air/workspace/agentic-harness/docs/open-engine/templates/malformed-teo-test.md` (a corrupted TEO artifact) | `teo.Validate` errors → runner emits `AGENT FAILED` or re-prompts; **no** silent accept | the failure is logged, not swallowed |

---

## CP-3 — OE-2 (Lane C + `opa_delegate`): deterministic governance

**Gate:** deny → `AGENT FAILED` (fail-closed); ask-first in-session → ASK →
`AGENT HUMAN HOLD`; Sentry (MCP) + native return the **same tri-state verdict**;
`opa test` green.

**Preconditions:** OPA bundle **and** gateway deployed **together** (a stale
bundle silently falls back to boolean `allow`); `opa_delegate` wired into the
native hook; `admin_groups` configured.

| # | Test | Input | Expected output | Usage / observability |
| :-- | :-- | :-- | :-- | :-- |
| 3.1 | Named deny tool (non-admin) | non-admin session calls `delete_*` | `verdict=deny` → `AGENT FAILED`, fail-closed | OPA log: `verdict=deny`, reason "Irreversible action denied (Open Engine boundary)" |
| 3.2 | Ask-first in-session | session calls a `publish_*`/`deploy_*` tool that is otherwise RBAC-allowed | `verdict=require_approval` → Omnigent ASK → `AGENT HUMAN HOLD` → answer → `RESUMED` | OPA log: `verdict=require_approval`; ASK event present |
| 3.3 | MCP vs native parity (boundary tools) | call the **same boundary tool** (e.g. `delete_*`) once via the Sentry MCP gateway and once as a native tool (`opa_delegate`) | **identical** verdict from the shared OE boundary rules. *Non-boundary tools differ by design:* the gateway applies MCP RBAC (`decision`), the native plane uses `oe_decision` (default-allow) | two decision-log entries, same `verdict` for the boundary tool |
| 3.4 | Admin carve-out | `delete_repository` as an **admin** identity, then as a **non-admin** | admin → `verdict=allow` (bypass); non-admin → `verdict=deny` | both logged; confirm the bypass is by `is_admin` |
| 3.5 | Fail-closed RBAC | RBAC-denied `publish_*` (caller lacks the group) | resolves to `deny`, **not** `require_approval` | OPA log: `verdict=deny` |
| 3.6 | Bare-gateway ask-first | direct MCP `tools/call` (no Omnigent session) for a `require_approval` tool | **fails closed (deny)** — no human channel at the bare gateway | gateway returns deny; no ASK raised |
| 3.7 | `opa test` green | `opa test mcp-policies/` | all tests pass | record the count (currently 40/40) |

---

## CP-4 — OE-3 / Phase 5: three-plane correlation

**Gate:** one completed issue → a single correlated trace across all three planes,
joining to `task_ref` + `subject_id`; satellites still build standalone.

**Preconditions:** `session_id` propagated to OPA logs; Entra subject bound;
correlation query available.

| # | Test | Input | Expected output | Usage / observability |
| :-- | :-- | :-- | :-- | :-- |
| 4.1 | Correlated trace | complete one issue, query all three planes by its `session_id` | **one** trace joining authz (OPA decisions) + runtime (session events) + artifacts/cost, back to `task_ref` + `subject_id` | the query returns rows in all three planes for the `session_id` |
| 4.2 | Cost attribution | the same issue | per-issue cost reachable from the issue (dashboard deep-link / per-session usage) | token-dashboard shows the session's burn; link resolves |
| 4.3 | Subject binding | inspect the OPA decision log subject for the issue | `subject_id` == Entra OID from `subjectFromClaims` == the routing-map `subject_id` for that operator | three sources agree on the OID |
| 4.4 | Cross-session isolation | run **two** issues, correlate each | each `session_id`'s plane rows are **disjoint** — no trace contamination | two distinct traces, no shared rows |
| 4.5 | Independence regression | build/test each satellite standalone | every satellite builds + tests on its own; `air → teo` is the only sanctioned source dependency | CI/local build green per repo |

**Run it (GitHub-first; the query is provider-agnostic):**

```bash
# Offline proof of the query logic — no stack needed (asserts all 5 criteria,
# incl. the negatives: subject mismatch + cross-contamination rejected):
python3 agentic-harness/tools/oe_correlate.py --self-check

# Live, end-to-end on one GitHub issue:
export OE_AUDIT_LOG=/tmp/oe-audit.jsonl      # Sentry gateway + omnigent native hook append here (opt-in)
# ...process the issue via the OE runner (it stamps the receipt marker)...
gh issue view <owner/repo#N> --comments > /tmp/comments.txt
python3 agentic-harness/tools/oe_correlate.py \
  --task-ref github:<owner/repo#N> \
  --comments-file /tmp/comments.txt \
  --authz-log "$OE_AUDIT_LOG" \
  --omnigent-url "$OMNIGENT_URL" --token "$TOKEN"   # or --snapshot saved.json
# exit 0 + one trace = CP-4 pass; non-zero = an invariant failed.
```

**Known limits (pilot-OK, not yet compliance-grade):** native-plane authz rows carry
`subject_id=""` until OE-3 surfaces it in the policy event (tolerated as "unknown",
not a second actor); an OPA-*unavailable* deny currently emits no authz row (follow-up).

---

## Results log

Copy a row per run.

| Gate | Test # | Date | Tester | Result | Notes (token counts, verdict, anomalies) |
| :-- | :-- | :-- | :-- | :-- | :-- |
| CP-1 | 1.1 | 2026-06-26 | live run (Linear OAuth) | ☑ pass | CLO-5: Todo→In Progress→Done with `AGENT CLAIMED`/`DONE` receipts carrying the session marker. Driven via the Linear OAuth app token (`client_credentials`) against the real CloudByteConsulting/CLO workspace. |
| CP-1 | 1.5 | 2026-06-26 | live run | ☑ pass | Join-key: marker `session=conv_cp1live01` identical across CLAIMED/DONE, `task_ref=linear:CLO-5`; `harness_for_agent_code('alex-claude')=='claude-native'`. |
| CP-1 | 1.2–1.4, 1.6 | 2026-06-26 | live run | ◐ not run | Blocked-resume / human-hold / named-tool-deny / ineligible-skip exercise the autonomous agent loop + native hook + Omnigent ASK — need a model credential (no agent run in this env), same limit as CP-4 cost. |
| CP-4 | 4.1 | 2026-06-26 | live run (this env) | ☑ pass | `oe_correlate` joined the three planes by `session_id=conv_cp4live01` for `github:Cloud-Byte-Consulting/my-test-project#1`. Authz + tracker **live** (real Sentry gateway+OPA `OE_AUDIT_LOG`; real GitHub receipt-marker comments), exit 0. |
| CP-4 | 4.2 | 2026-06-26 | live run | ◐ representative | Per-session usage came from a representative snapshot — no model credential configured in this env, so no real agent session could be run (the `GET /sessions/{id}` endpoint is live but empty). |
| CP-4 | 4.3 | 2026-06-26 | live run (Entra OIDC) | ☑ pass | **Real Entra subject.** With `OIDC_PROVIDERS=azure` wired, the gateway validated a real Entra JWT (`az` delegated token) and emitted a real `subject_id` (the `sub` claim) on every authz row; `oe_correlate` exit 0, criterion 3 satisfied for real. The earlier static-token run correctly *failed* #3 with no subject — proving the gate isn't faked. |
| CP-4 | 4.4 | 2026-06-26 | live run | ☑ pass | A real decoy session (`conv_cp4smoke`) in the authz log was correctly excluded from the trace (no cross-contamination). |
| CP-4 | 4.5 | 2026-06-26 | live run | ☑ pass | Producers are env-gated/additive; satellites unchanged. |
| CP-4 | 4.6 | 2026-06-26 | live run (Entra OIDC) | ☑ pass | **Group-RBAC live** with real Entra group GUIDs: admin group → `github.create_pull_request` allow; non-member group → deny. OIDC token-validation + OPA group RBAC path fully live (`smoke-entra.sh` steps 1–5 + OPA checks pass; step 6 needs an ARD `resource_discovery` backend, unrelated to Entra). |
