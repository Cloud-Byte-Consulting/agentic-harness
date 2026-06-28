# Why a customer picks Microsoft Entra (Entra Agent ID) — benefits & scenarios

Spike **CLO-21**. The *what/how* of Entra Agent ID lives in [azure.md](azure.md);
this doc is the **benefits + buying-scenario** layer on top of it. See also the
[agent-identity hub](README.md) (decision D17) and the
[getting-started hub](../../../GETTING_STARTED.md).

## Why a customer picks Entra

The decision is rarely "best agent-identity model in the abstract" — it's
**"govern agents inside the estate we already run."** Entra wins when the
customer is already a Microsoft shop:

- **Conditional Access reuse.** The same CA engine that gates human sign-in
  (device compliance, risk, named locations, session controls) applies to agent
  identities with **no new policy plane**. One blast radius, one audit story.
- **The M365 estate is the data.** Agents that touch SharePoint/Exchange/Teams/
  Graph get Entra-issued tokens and **per-agent Graph permissions** — the access
  surface and the identity are the same system. Agent 365 makes the agent a
  first-class, governable object across that estate.
- **Token policies + ID Protection** apply to agents: risky-token detection,
  revocation, and lifecycle (create in bulk, retire without orphaned creds) for
  fleets that churn thousands of agents a day.
- **Sponsor + sign-in logs** give the compliance/SecOps team an answer to "which
  human stood this agent up, and what did it do" with **zero new tooling** —
  it's the Entra admin center they already use.

If the customer is *not* on Entra, none of this compounds — that is the limit
below, and the reason the hub keeps three per-cloud files instead of one.

## Customer scenarios

1. **Regulated M365 enterprise (finance/health).** Already enforces device-
   compliant CA for staff. Wants agents under the *same* CA + DLP envelope, not a
   side channel. Entra Agent ID lets the agent inherit named-location/risk policy
   and shows up in the same sign-in logs auditors already pull.
2. **Copilot Studio / low-code fleet.** Business users spin up many short-lived
   agents. Entra's bulk-create + blueprint + retire-without-orphaned-creds
   lifecycle is the only one of the three clouds purpose-built for that churn.
3. **Graph-scoped task agent.** An agent that only reads one SharePoint site +
   sends Teams messages gets a **right-sized** Graph permission set on its own
   account — and Entra's "prevent agents from gaining critical roles" guardrail
   keeps it out of Global Admin by construction.
4. **Mixed human/agent approval flow.** A system that needs a human-shaped
   principal pairs the agent 1:1 with an *agent's user account*, so existing
   user-targeted policy and access reviews keep working unchanged.

## Fit & limits for Open Engine

**Fit.** Per decision **D17**, OE runs each agent under its *own non-admin*
Entra agent account (autonomous model), and the agent's Entra **group OIDs** flow
through OE-3 group plumbing into the rego `is_admin` carve-out — so Entra's
"keep agents out of critical roles" guardrail and OE's OPA admin carve-out are
**the same control expressed twice**. Entra's *inbound auth / "reliably identify
the caller"* maps to the Sentry `MCP-Session-Id`↔subject binding.

**Limits.**
- **CA does not reach our MCP plane.** Conditional Access gates *token issuance*
  to the agent; it does **not** see individual `tools/call`s at Sentry. The
  per-call decision stays in OPA (`mcp_auth.rego`). Entra is the identity/coarse
  gate; OE is the fine-grained gate. Don't sell CA as tool-level governance.
- **Value is Microsoft-conditional.** The benefits above only compound on an
  Entra/M365 estate; for non-Microsoft customers see [aws.md](aws.md) /
  [google.md](google.md).
- **Licensing gates the good parts.** CA-for-agents/ID Protection need P1/P2 +
  Agent 365 (E7 or E5+Agent 365). Flag this in any customer fit assessment.

## Follow-ups

- Verify Entra agent **group OIDs** actually surface in the access token consumed
  by `subjectFromClaims` (autonomous agents may need an explicit group claim
  mapping) — see the implementation story below.
- Cross-check the licensing tiers in [azure.md](azure.md) against current
  Microsoft docs before any customer-facing fit assessment.
