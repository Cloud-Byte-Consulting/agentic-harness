---
description: Authorization gate for any proposed write across TRAPI/Cordillera/GitHub/Grafana/ADO. Use /mutation-gate before executing a mutation to decide ALLOW / BLOCK / REVISE / ESCALATE.
---
<!-- GENERATED from capabilities.json by scripts/sync-capabilities.ps1 - edit the manifest, not this file. -->

# /mutation-gate

**Source of truth:** `${AGENT_WORKSPACE}/agent-coordination/agents/20-mutation-gate.md` - read it now, then operate per its rules.

**Always preload:** `${AGENT_WORKSPACE}/agent-coordination/knowledge-base/00-OVERVIEW.md`.

Require a structured Action Proposal; evaluate authorization -> evidence -> exposure -> policy; default to ESCALATE under uncertainty. You return a decision; the owning system executes only after ALLOW. Read-only by default.
