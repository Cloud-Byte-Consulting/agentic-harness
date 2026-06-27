---
description: How chief-of-staff (cockpit), agent-coordination (domain brain + guardrails), and the Omnigent supervisor (execution engine) compose into one safe sense->decide->execute->record loop. Use when driving work end-to-end, not just routing a question.
---
<!-- GENERATED from capabilities.json by scripts/sync-capabilities.ps1 - edit the manifest, not this file. -->

# /co-operating-model

**Source of truth:** `${AGENT_WORKSPACE}/agent-coordination/agents/40-co-operating-model.md` - read it now, then operate per its rules.

**Always preload:** `${AGENT_WORKSPACE}/agent-coordination/knowledge-base/00-OVERVIEW.md`.

**Tooling:** prefer `${AGENT_WORKSPACE}/agent-coordination/scripts/co-dispatch.ps1 "<intent>"` as the deterministic classifier.

**Execution engine (Phase 2+):** copilot-role-router is archived. Execute through the Omnigent supervisor (`omnigent/examples/role_router/config.yaml`) and require Judge PASS via the ported Judge skill (`omnigent/examples/role_router/skills/judge/`). Hill-climb budget is enforced by the nessie policy (not a free-spin loop).

Recon and Judge are read-only. Tier-2/3 writes require a /mutation-gate ALLOW, not just an approval click. Default ESCALATE on unclear ownership or low gate confidence. Grande (Team-alpha) and TEAM-BETA prod always require explicit authorization.
