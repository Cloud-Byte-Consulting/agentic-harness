# Microsoft Entra Agent ID (`azure`)

**Source:** <https://learn.microsoft.com/en-us/entra/agent-id/what-are-agent-identities>
(doc dated 2025-11-06, updated 2026-06-15).

## What it is

A first-class identity construct in Microsoft Entra ID built specifically for AI
agents — **not** a human user and **not** a classic application service
principal. Designed for **scale and ephemerality**: agents are often created
dynamically (Copilot Studio, API orchestration), may live for minutes, and can be
created/destroyed thousands of times a day. Organizations create them in bulk,
apply consistent policies, and retire them without orphaned credentials.

Stated security goals:
- Distinguish agent operations from workforce/customer/workload identities.
- Give agents **right-sized** access.
- **Prevent agents from gaining the most critical security roles and systems.**
- Scale identity management to large, churning fleets of agents.

## Identity model

- Each agent gets its **own Entra ID agent account** (an object identity).
- It does **not** use human auth (passwords/MFA/passkeys) and has no mailbox/teams.
- For systems that require a human-shaped principal, an agent identity can be
  paired 1:1 with an **"agent's user account"** (a special Entra user account).
- The human who creates an agent is recorded as its **sponsor**.

## Access models

- **Autonomous access** — rights granted **directly to the agent identity**:
  Microsoft Graph permissions, **Azure RBAC roles**, **Microsoft Entra directory
  roles**, **Microsoft Entra app roles**, and group memberships.
- **Delegated access** — the agent acts **on behalf of a user**, using the
  user's rights (the user controls what is delegated).
- **Inbound auth** — agents accept requests secured by Entra-issued tokens,
  "allowing the agent to **reliably identify the caller** and make authorization
  decisions."

## Audit

Every action is logged **as an AI agent** in Entra sign-in logs and is viewable
in the **Agent identities** tab of the Entra admin center. Activity is subject to
any security policies enforced for agent identities (Conditional Access, etc.).

## Status & licensing

- **Microsoft Entra Agent ID** (the platform for creating/managing agent
  identities + blueprints): **available for all Microsoft Entra customers**.
- **Microsoft Agent 365** license per agent to operate across M365.
- Extending Entra security to agents: **M365 E7** (Agent 365 + Entra Suite) or
  **M365 E5 + Agent 365**. Standalone with Agent 365: Conditional Access (Entra
  ID P1), ID Protection (P2), ID Governance (P1), Network (Entra Internet Access).

## Mapping to Open Engine

- `provider = "azure"`; `subject_id` = the **agent identity's object id** (OID).
- The OE-3 `groups` plumbing carries the **agent's** Entra group OIDs → the rego
  `is_admin` admin carve-out keys on the agent. Provision OE agents into
  **non-admin** groups by default.
- Use the **autonomous** model (the agent's own rights), not delegated, so
  governance attaches to the agent, not the operator.
- "Inbound auth / reliably identify the caller" ≙ the Sentry
  `MCP-Session-Id`↔subject binding.
