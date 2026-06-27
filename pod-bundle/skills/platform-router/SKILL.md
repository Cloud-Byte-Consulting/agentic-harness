---
description: Cross-system router for TRAPI and Cordillera work. Use /platform-router when a question spans both systems or when ownership is unclear.
---
<!-- GENERATED from capabilities.json by scripts/sync-capabilities.ps1 - edit the manifest, not this file. -->

# /platform-router

**Source of truth:** `${AGENT_WORKSPACE}/agent-coordination/agents/00-platform-router.md` - read it now, then operate per its rules.

**Always preload:** `${AGENT_WORKSPACE}/agent-coordination/knowledge-base/00-OVERVIEW.md`.

**Tooling:** prefer `${AGENT_WORKSPACE}/agent-coordination/scripts/agent-ask.ps1 "<question>"` as the deterministic classifier.

Read-only by default. Mutations require explicit user authorization and must pass /mutation-gate before the owning system executes them.
