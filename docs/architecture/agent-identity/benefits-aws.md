# Why AWS for agent identity — benefits & customer scenarios

Companion to [aws.md](aws.md) (the mechanics of **Amazon Bedrock AgentCore
Identity** + **Cognito**). This file is the *sell*: why a customer already on
AWS picks it, what it buys them, and where it fits Open Engine.
Hub: [Getting Started](../../../GETTING_STARTED.md).

## What it is (one line)

A managed identity + credential-vault service for AI agents: validates **who
invokes an agent** (inbound, SigV4 / OAuth-OIDC), brokers **what the agent may
reach downstream** (outbound, SigV4 / OAuth / API keys), and keeps agents off
long-lived secrets via the **resource token vault** — with audit trails.

## Why a customer picks it

- **Already-on-AWS gravity.** IAM, SigV4, Cognito user pools, and CloudTrail are
  already the org's control plane. AgentCore Identity reuses them, so agents
  inherit existing SCPs, permission boundaries, and audit pipelines — no second
  IdP to run.
- **No long-lived agent secrets.** The token vault issues short-lived downstream
  credentials on demand. A leaked agent process doesn't leak a standing key —
  the blast radius is one token TTL, and rotation is the platform's job.
- **Dual auth in one place.** Inbound (the agent authorizer validates the caller)
  and outbound (broker to AWS + third-party APIs) are the same service, so
  "who called this agent" and "what did it touch" land in one audit trail.
- **Own-identity vs on-behalf-of as a first-class switch.** An agent can act as
  itself (autonomous) or OBO a user — the choice that decides whether
  least-privilege attaches to the *agent* or the *human*.
- **Runs anywhere the agent runs.** ECS, EKS, Lambda, or on-prem — not locked to
  a single managed runtime.

## Customer scenarios

1. **Regulated AWS shop, audit-first.** A bank runs OE agents in EKS and must
   answer "which identity touched this resource, on whose authority?" CloudTrail
   + AgentCore inbound auth gives the per-agent SigV4 trail; OE's three-plane
   audit join keys on the agent's `subject_id`.
2. **Third-party SaaS reach without secret sprawl.** An agent needs Salesforce +
   GitHub tokens. The resource credential provider brokers OAuth flows and the
   vault holds the tokens — the agent code carries none, and the SecOps team
   rotates centrally.
3. **Cognito-fronted internal tools.** A customer already authenticates staff
   through a Cognito user pool. Agents invoked by those staff validate inbound
   via the same pool; `cognito:groups` flows straight into OE's group plumbing
   (already normalized by `subjectFromClaims`).
4. **Lambda burst agents, no standing creds.** Event-driven agents spin up per
   request; the vault issues a scoped token per invocation and it expires with
   the function — nothing to leak between runs.

## Fit & limits for Open Engine

**Fit.** `provider = "aws"`, `subject_id` from the token `sub`, groups from
`cognito:groups`. AgentCore **inbound auth** is the conceptual match for Sentry's
`MCP-Session-Id`↔subject binding (`cmd/gateway/main.go`) — both verify *who is
calling*. Run each agent under its **own non-admin identity** (decision D17,
autonomous model) so OE-3's `is_admin` carve-out keys on the agent, not the
operator. The vault complements the anti-spoof binding on the audit side.

**Limits.** AgentCore Identity was preview-era (2025) and AWS-centric — the value
collapses outside an AWS-resident estate. It does **not** decide tool-level
policy: it answers *authenticated?*, not *authorized for this MCP tool?* — that
stays with the Sentry/OPA gate (`mcp_auth.rego`, decision `/v1/data/mcp/auth/decision`).
And the gateway only consumes the **OIDC** inbound path today; pure-SigV4 callers
aren't yet verified by Sentry (see follow-up).

## Follow-ups

- **Code:** teach the Sentry gateway to accept AgentCore inbound auth (SigV4 +
  Cognito OIDC) as a Bearer source, with `cognito:groups`→OPA `input.groups`.
  See the implementation story below.
- Related identity surfaces: [azure.md](azure.md) · [google.md](google.md) ·
  [agent-identity README](README.md).
- Open integration gaps: CLO-27 (Cachy native attribution), CLO-28 (gateway v2
  backend-MCP proxy).
