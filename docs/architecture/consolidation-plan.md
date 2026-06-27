# AIR + Omnigent consolidation plan

**Status:** Approved direction; execution pending Phase 0+.  
**Principle:** Independence of tooling, **zero overlap of functionality** — each concern has exactly one owner.

---

## Goal

Absorb or shrink `air` and `aos` so they do not compete with the runtime. Satellite repos stay independently shippable but must not re-implement Omnigent (or another designated owner).

**Surviving runtime:** [omnigent](../../../omnigent) — sessions, harness adapters, sandboxes, session policies, server/UI.

**Surviving satellites (boundaries only):**

| Satellite | Owns exclusively |
| :--- | :--- |
| [Agentic-Sentry](../../../Agentic-Sentry) | MCP `tools/call` OPA/RBAC gateway (pre-call authz) |
| [Agentic-Resource-Discovery](../../../Agentic-Resource-Discovery) | Federated discovery spec + registry |
| [Cachy](../../../Cachy) | LLM HTTP proxy + compression/CCR (token plane) |
| [teo](../../../teo) | Token-efficient output grammar |
| [token-dashboard](../../../token-dashboard) | Local cost / activity measurement |

**Shrunk to content + docs:** [agentic-harness](../../) — pod-bundle, personas, skills, stack docs; `air` CLI becomes **content-only** (no session/bootstrap/runtime).

**Deprecated / archived:**

| Repo | Disposition |
| :--- | :--- |
| [aos](../../../aos) | Archive after doc merge into this folder |
| [copilot-role-router](../../../copilot-role-router) | Orchestration → Omnigent supervisor YAML; optional thin shim, then archive |
| `air` bootstrap, install providers, session paths | → Omnigent stack profile |

---

## Single-owner matrix

| Concern | Owner | Not owned by |
| :--- | :--- | :--- |
| Live agent sessions, resume, events, artifacts | **Omnigent** | air, aos, role-router |
| Session server / share URL / cloud runner | **Omnigent** | air |
| OS sandbox + network interception for agent compute | **Omnigent** | Sentry, Cachy |
| Session-level cost budgets + ASK pause | **Omnigent** | token-dashboard (measure only) |
| MCP tool authorization (deterministic) | **OPA/Rego via Sentry** | LLM PromptPolicy, role-router mutation rules for security |
| Native tool authorization (Bash, run_command, etc.) | **OPA via `opa_delegate` + `opa-hook`** | Sentry alone (MCP-only) |
| Federated discovery | **ARD** | Omnigent catalog duplication |
| Token-plane proxy / compression | **Cachy** | Omnigent network layer |
| Output wire format | **TEO** | ad-hoc JSON in each repo |
| Cost analytics (exact / activity / estimate) | **token-dashboard** | Omnigent inline only |
| Personas, pods, skills, stack manifest content | **agentic-harness** | Omnigent |
| Judge / hill-climb / multi-role orchestration | **Omnigent supervisor YAML** (ported from role-router) | role-router core long-term |

**Policy rule:** Omnigent engine stays open (phases, ASK UX). **All allow/deny for tools and sensitive actions = OPA/Rego.** No LLM-judged security gates.

**Post-tool redaction:** Omnigent deterministic `tool_result` / `output` filtering — not role-router output-guard for security.

---

## Phased execution

| Phase | Deliverable |
| :--- | :--- |
| **0** | This doc set + ownership matrix; merge `aos` docs; deprecations listed in README/manifest |
| **1** | Omnigent stack profile replaces `air bootstrap` / MCP-inject / compose overlap |
| **2** | Judge/hill-climb → Omnigent supervisor YAML; role-router orchestration deprecated — **done:** config port (`omnigent/examples/role_router/config.yaml` modeled on `polly`) + the two code-enforced pieces (`hillclimb_budget` nessie policy, Judge skill); role-router archived in the manifest + README. Follow-up: sweep the ~15 generated persona `surfaces`/`workflows.md` refs (regen from `capabilities.json`). |
| **3** | Sentry MCP gateway + `opa_delegate` for native tools + shared Rego bundle |
| **3b** | Standalone `opa-hook` CLI; PreToolUse for Claude/Copilot/Antigravity (verify agy 2.0) — **opa-hook built** (`agentic-harness/tools/opa_hook.py`); agy 2.0 PreToolUse still to verify |
| **4** | Content-only harness; archive `aos` — **done:** removed air runtime (`bootstrap`/`install`/`doctor` + `internal/{mcpinject,profile}`, absorbed into the omnigent stack profile); air is content-only (agents/skills/personas/package/targets/status/version), build+tests green. `aos` archived (deprecation-notice commit pushed; flip repo to Archived in GitHub). |
| **5** | Enterprise profile: Entra, audit pipeline, [enterprise-pitch.md](../enterprise-pitch.md), BECU pilots — **audit-correlation built:** `tools/oe_correlate.py` (three-plane read-side query, CP-4 invariants in code) + the producers it needs (env-gated `OE_AUDIT_LOG` authz emit in Sentry+omnigent; receipt-marker contract). Entra subject/groups + session binding already done (OE-3). **CP-4 fixture proven**; live CP-4 runbook in checkpoint-manual-tests. Follow-ups: OPA-unavailable deny emits no authz row; native subject_id="" pre-OE-3; ASK-event persistence; label-query API; OTel/SIEM forwarding; BECU pilot. |

See [policy-opa-hooks-audit.md](policy-opa-hooks-audit.md) for authz gap closure and audit correlation.

---

## What we cancel

- `air install` provider orchestration duplicating Omnigent stack
- Parallel session APIs (`air session` as runtime)
- AOS-native kernel (placeholder only)
- LLM `PromptPolicy` for security allow/deny
- Sibling-to-sibling ingestion seams without Omnigent/ARD boundary

---

## Open decisions

1. **Omnigent fork vs upstream** for enterprise patches (Entra, audit export).
2. **role-router archive timing** — immediate vs one release cycle with adapter shim.
3. **Antigravity PreToolUse** — Omnigent verified non-firing on agy 1.0.8; re-verify on agy 2.0 before editor-only tier.

---

## Prior art merged here

- [air-vs-omnigent.md](air-vs-omnigent.md) — overlap tables (pre-decision to standardize on Omnigent runtime)
- [aos-vs-air.md](aos-vs-air.md) — AOS pillar mapping
- [integration-plan.md](../integration-plan.md) — original wiring (superseded for runtime ownership)

---

## Nice to have (deferred — do last)

- **ponytail statusline badge** — wire the `statusLine` block into `~/.claude/settings.json` so the active mode (`[PONYTAIL]`) shows in the prompt. Dev-environment polish, unrelated to consolidation.
