# Open Engine — decision log & follow-up register

**For review.** Consolidates the decisions made and the follow-ups still open as
of 2026-06-26. Working branch: `feat/open-engine-wave0` (agentic-harness +
Agentic-Sentry → Gitea; omnigent → fork `Cloud-Byte-Consulting/omnigent`; `aos`
skipped). Canonical detail lives in
`/home/bittahcriminal/air/workspace/agentic-harness/docs/architecture/governed-platform-architecture.md`
and `/home/bittahcriminal/air/workspace/agentic-harness/docs/architecture/open-engine-integration-plan.md`.

---

## 1. Decisions made

| # | Decision | Choice | Why | Encoded in |
| :-- | :-- | :-- | :-- | :-- |
| D1 | Harness for OE agent codes | **`claude → claude-native`** (native CLI bridge, not the SDK orchestrator) | OE agents are persistent vendor CLIs claiming tasks, not in-process SDK calls | `/home/bittahcriminal/air/workspace/omnigent/omnigent/harness_aliases.py` (`OPENENGINE_AGENT_HARNESS`) + the stack profile |
| D2 | Ask-first vs RBAC-deny overlap | **Fail-closed** — RBAC-denied `publish/email/deploy` → `deny`, not `require_approval` | A denied action must not become an approval prompt | `mcp_auth.rego` (`oe_boundary_approval` gated on `allow`) |
| D3 | OE boundaries vs human admins | **Admin carve-out** (role-based: `is_admin` bypasses) | Boundaries exist to constrain unattended agents, not human admins | `mcp_auth.rego` + `governed-platform-architecture.md` §5. **Caveat:** role-based → an agent under an admin identity inherits the bypass; run OE agents under a non-admin service identity |
| D4 | `opa_delegate` rollout | **Default-off**, `off`→`shadow`→`enforce`; **deny-wins via engine policy composition**; OPA-unreachable in enforce → **fail-closed DENY** | Zero behaviour change until promoted; never fail open | server-side builtin `/home/bittahcriminal/air/workspace/omnigent/omnigent/policies/builtins/opa.py` |
| D5 | Native-plane policy query | The OPA policy builtin queries **`oe_decision`** (default-allow + OE boundaries only), NOT the gateway's full-RBAC `decision` | Host tools (`Bash`) have no MCP allow-list entry; the gateway default-deny would blanket-deny them | `mcp_auth.rego` (`oe_decision`) + `policies/builtins/opa.py` |
| D13 | OPA delegation location | **Server-side PolicyEngine builtin** (`policies/builtins/opa.py`), not the hook | Reuses the existing ASK elicitation gate so native-plane `require_approval` renders a human prompt; deny-wins is free via policy composition; hook stays simple | `/home/bittahcriminal/air/workspace/omnigent/omnigent/policies/builtins/opa.py` + the 3 stack profiles' `guardrails.policies` |
| D14 | Profile→session activation (Lane B) | Stack profile selected by the **`openengine.profile` label**; its `guardrails.policies` attached as **session policies** (not merged into the shared spec, not `default_policies`) | The AgentSpec isn't built at create time (the engine is lazy); session policies are the per-session, additive seam every engine build already reads — zero build-site change | `/home/bittahcriminal/air/workspace/omnigent/omnigent/server/profiles.py` + the `sessions.py` create handler |
| D15 | Native-plane subject groups (OE-3) | Plumb Entra `groups` from the verified id_token → session JWT → policy event → OPA input; fail-safe to `[]` (strict) on any absent/malformed/unverified/header-mode source | Unblocks the rego admin carve-out on the native plane; groups only ever *relax* for admins (never tighten) and can't come from any client-controllable input | `routes/auth.py`, `oidc.py`, `auth.py`, `policies/{types,schema,function,builtins/opa}.py`, `sessions.py` |
| D16 | Audit session↔subject binding (OE-3) | The Sentry gateway binds `MCP-Session-Id` → subject **trust-on-first-use**; a session id replayed under a different subject is rejected (403) | Makes the three-plane correlation join non-spoofable (plan risk 2) without touching the authz path; assumes one subject per session (the OE model) | `/home/bittahcriminal/air/workspace/Agentic-Sentry/cmd/gateway/main.go` (`bindSession`) |
| D17 | Agent identity model (research) | Run each OE agent-code under its **own dedicated non-admin agent identity** (Entra Agent ID / AWS AgentCore Identity / GCP Agent Identity), using the *autonomous* (own-rights) model, not delegated/on-behalf-of | Cloud vendors converged on per-agent least-privilege identities precisely to keep agents out of admin roles; this resolves our "agent-under-admin inherits the bypass" caveat — groups attach to the agent, not the operator | `decision-log.md` (this); deployment guidance |
| D6 | omnigent remote | Push to the **fork** `Cloud-Byte-Consulting/omnigent`, not upstream `omnigent-ai/omnigent`; `origin` stays upstream for sync | Can't/shouldn't push WIP to third-party upstream | git remote `fork` on the omnigent repo |
| D7 | Multi-tracker support | **Yes** — Linear + GitHub Issues + Jira; new phase **OE-1b**, after CP-1, parallel to OE-2/OE-3, off the critical path | Open Engine is a tracker convention; the trackers all provide issues/labels/comments/status/query | `open-engine-integration-plan.md` (OE-1b row) + `governed-platform-architecture.md` §4a |
| D8 | Tracker order | **Linear (live) → GitHub Issues (first) → Jira (enterprise/OE-3)** | GitHub MCP already in use, labels-as-status is the cheapest adapter; Jira is the enterprise tracker | same |
| D9 | Join key | Generalized to **`task_ref` = `<provider>:<id>`**; label KEY `openengine.issue` stays, VALUE becomes provider-qualified (`linear:ENG-123` / `github:<owner>/<repo>#<n>` / `jira:PROJ-45`) | One key, tracker-neutral; OE-2/OE-3 need no rework | all four architecture docs |
| D10 | GitHub status model | **Labels-as-status** (`oe:todo/working/review/needs-input/done`); Projects v2 deliberately skipped for v1 | GitHub has no native workflow states; labels are the cheapest map | OE-1b GitHub adapter (landed) — `github-status-labels.md`, runner prompt 20, `openengine_stack_github.yaml` |
| D11 | Security: issue prose | **Verdict supersedes issue prose** — approval only from an Omnigent ASK, never from issue text | Closes a prompt-injection hole (a task body can't self-grant) | `17_queue_run_prompt_v2.md` `<boundaries>` |
| D12 | Workflow model tiering | Tier model+effort per task; don't default every subagent to session Opus | Cost/appropriateness | saved preference (memory) |

---

## 2. Open follow-ups (action register)

### Needs you (human)
- **CP-1 manual gate** — the dev-loop smoke test; not yet run. Prereq: a Linear/Open Engine workspace set up (status ledger, routing map, `agent-instructions` label, states/labels, `alex-claude` identity). Run sheet: `/home/bittahcriminal/air/workspace/agentic-harness/docs/architecture/checkpoint-manual-tests.md`.
- **Branch / PR decision** — merge `feat/open-engine-wave0` to `main` or open PRs across the three repos. (`gh` is not authenticated here — needs `! gh auth login` or web-UI PRs.)
- **omnigent fork PR** — open at `https://github.com/Cloud-Byte-Consulting/omnigent/pull/new/feat/open-engine-wave0` when ready.
- **`aos`** — empty repo (unborn HEAD), archive-bound; decide init/disposition.

### Code follow-ups (before `enforce` is fully usable)
- ~~Native-plane `require_approval` ASK rendering~~ — **RESOLVED (2026-06-26).** OPA delegation moved to a server-side PolicyEngine builtin (`/home/bittahcriminal/air/workspace/omnigent/omnigent/policies/builtins/opa.py`); it returns `ASK` on `require_approval`, which the existing server elicitation gate renders as a human prompt. The hook-side branch was removed. A governed CLI agent can now request runtime permission interactively instead of failing closed. **Activation (Lane B loader landed 2026-06-26):** a session opts in via the `openengine.profile=<name>` label on `POST /v1/sessions`; the create handler (`/home/bittahcriminal/air/workspace/omnigent/omnigent/server/profiles.py`) attaches the profile's `guardrails.policies` as session policies. So enforcement is AND-gated on (`OMNIGENT_OPA_DELEGATE_MODE=enforce`) AND (the `openengine.profile` label) — and whoever creates the session (the runner/shim) must set the label. (Also, before enabling `enforce`: confirm connector-MCP tools aren't double-gated on both Sentry and the native hook — pre-existing, flagged by the adversarial review.)
- **Profile attach covers only the JSON create path** — `_apply_openengine_profile_if_requested` is wired into `POST /v1/sessions` (the OE runner/shim path). The multipart **bundle-import** and **terminal** create paths don't apply the profile, so an OE session born there is silently ungoverned. Root-cause fix: hoist the call to a point both paths route through, or wire `_create_session_from_bundle` too. (Adversarial review finding #5; low risk today since the shim uses the JSON path. The runner/shim must also actually *set* the `openengine.profile` label.)
- ~~Groups binding~~ — **RESOLVED (OE-3, 2026-06-26).** The authenticated subject's Entra `groups` are now plumbed from the verified id_token → minted session JWT → policy event → the OPA builtin (`build_opa_input(groups=)`), so the admin carve-out applies on the native plane. Fail-safe (no/overage/malformed/header-mode → `[]` → strict); escalation-audited (groups only from cryptographically verified tokens; no client-controllable path; the native hook's POST body can't override them). **Operational note:** admin-group *revocation* lags by up to the session TTL (~8h) — token-lifetime cache + JWT design. Still gated behind `enforce` + the `openengine.profile` label.
- **`session_id` ↔ `MCP-Session-Id` subject binding** — `Agentic-Sentry/cmd/gateway/main.go:131`; the audit join is spoofable until bound. Pick the header.
- **Deploy coupling** — ship the OPA bundle + Sentry gateway together; a stale bundle silently falls back to boolean `allow`.

### Ops / validation
- **Run the Python test suites in a real omnigent env** — no venv/`uv`/`httpx` here; `opa test` (44/44) and offline logic checks ran, but `pytest` for `test_opa_delegate.py` needs the omnigent deps. (`opa test` and offline validation were run successfully.)
- **Enable `opa_delegate`** — stand up an OPA server with the bundle, set `OMNIGENT_OPA_DELEGATE_MODE=shadow`, observe parity, then `enforce`.

### Roadmap (next phases)
- **OE-1b Jira adapter** — with the OE-3/Phase-5 enterprise push (native workflow transitions).
- **Full builtin Rego parity** — the Wave-0 spike ported 2 builtins; full parity before declaring "Rego is the one authority".
- Carried consolidation items: role-router → supervisor YAML (Phase 2); `air` shrink + `aos` archive (Phase 4); Entra + correlation query (Phase 5).

---

## 3. Build status (what's real)

| Area | State |
| :-- | :-- |
| OE-0 docs + conventions | ✅ landed |
| Milestone 1 (task→session bridge, Linear) | ✅ built (CP-1 manual gate pending) |
| Tri-state Rego + parity spike | ✅ `opa test` 44/44 |
| `opa_delegate` (OE-2 core) | ✅ built, default-off, adversarially reviewed (no fail-open); **server-side builtin — `require_approval` renders a human ASK**; **activatable via the `openengine.profile` label (Lane B loader)**; groups (admin carve-out on native plane) still a follow-up |
| OE-1b Lane B profile loader | ✅ built — `openengine.profile` label attaches a profile's `guardrails.policies` as session policies; path-safe, fail-safe; offline + CI tests |
| OE-3 native subject-groups | ✅ built — Entra groups plumbed auth→JWT→event→OPA; admin carve-out works on the native plane; fail-safe + escalation-audited (no fail-open-to-admin); gated behind `enforce`. Revocation lags ≤ session TTL. |
| OE-3 Sentry session binding | ✅ built — the gateway binds `MCP-Session-Id` to the authenticated subject (trust-on-first-use) and rejects (403) a replay under a different subject; build/vet/test green. **OE-3 complete.** |
| Phase 3b opa-hook CLI | ✅ built — standalone editor `PreToolUse` gate (`tools/opa_hook.py`); queries `oe_decision`; deny/ask/allow; mode-gated, fail-closed; self-check green. Claude Code/Codex wired; **agy 2.0 PreToolUse still to verify**. |
| Phase 4 content-only air + aos archive | ✅ done — removed air runtime (`bootstrap`/`install`/`doctor`, `internal/{mcpinject,profile}`) absorbed by the omnigent stack profile; air is content-only, `go build`/`vet`/`test` green, no external importers. `aos` archived (notice commit pushed; GitHub Archived-flag flip is manual). |
| Phase 2 role-router → supervisor | ✅ done — config port (`omnigent/examples/role_router/config.yaml` on the `polly` engine) + 2 code-enforced pieces: `hillclimb_budget` nessie policy (sticky rework budget) + Judge skill (tripwire floor + anti-gaming merge). role-router archived (manifest + README banner). **Adversarial verify caught a fail-open:** the budget was first written against `session_state`, which `RunnerToolPolicyGate` never provides — rewrote to **closure** state (sibling of `spawn_bounds`) so the cap actually fires; test now drives the real callable. 87 nessie tests green. Follow-up: sweep ~15 generated persona refs (regen from `capabilities.json`). |
| TEO LLM-I/O | ✅ docs + carve-out + agent profile |
| Phase 5 audit correlation | ✅ built — three-plane query `tools/oe_correlate.py` (CP-4 invariants enforced in code: empty native subject tolerated as "unknown", two non-empty subjects rejected, decoy/ambiguous rejected) + env-gated `OE_AUDIT_LOG` authz emit (Sentry MCP plane + omnigent native plane) + the receipt-marker contract (one canonical string across CONTRACT + 3 prompts). **CP-4 fixture self-check PASS** incl. both negatives; Sentry + omnigent producer tests green. Live CP-4 runbook in checkpoint-manual-tests; **CP-4 + CP-1 validated live** (real Entra subject + group-RBAC; Linear OAuth happy-path). Adversarial verify caught a live-breaking bug (native `subject_id=""` would falsely trip the single-actor check) — fixed. **Resolved 2026-06-26:** OPA-unavailable now emits a fail-closed `deny` authz row on both planes (no silent audit gap); `subject_id` prefers the stable Entra `oid` over the pairwise `sub`. **Resolved 2026-06-27 (loop):** native-plane `session_id`/`subject_id` now plumbed into the policy event (`EvaluationContext`→`_build_event`→`event["context"]`, mirroring OE-3 groups) so native authz rows correlate; persona role-router refs swept to the Omnigent supervisor; Lane-B loader coverage added; ponytail statusline wired. **Resolved 2026-06-27 (loop 2):** **bundle-create governance gap CLOSED** — the multipart/bundle path now applies the OE profile (positive coverage tests); `aos` GitHub repo **Archived**; OTel/SIEM forwarding **decided** (the `OE_AUDIT_LOG` JSONL is the integration point — a standard collector tails it, no bespoke shipper); label-query API **deferred** with rationale (receipt-marker path is canonical; a labels-JOIN across store impls isn't worth the low value). **ASK-event persistence DONE** — elicitation raise/resolve persisted as `ConversationItem`s (`NON_CONTENT_ITEM_TYPES`, best-effort, exactly-once dedup; +tests), so the audit runtime plane has a durable human-in-the-loop record. **All code-actionable follow-ups resolved.** Remaining is decision/credential-gated only: a model credential (→ CP-1 1.2–1.4 + CP-4 cost), agy 2.0 verify, the Jira MCP endpoint, the fork-vs-upstream + archive-timing decisions, CP-2/CP-3, and the PR merge. |
| OE-1b GitHub adapter | ✅ landed — status-label map, runner prompt 20, MCP setup, GitHub stack profile, shim provider-qualified value (Opus verify 5/5 PASS) |
| OE-1b Jira adapter | ✅ landed — status-transition map (gated transitions), runner prompt 21, MCP setup, Jira stack profile; no shim change (Opus verify 5/5 PASS) |
| CP-1…CP-4 | manual gates — **none run yet** |

---

*Living document — update as decisions/follow-ups change.*
