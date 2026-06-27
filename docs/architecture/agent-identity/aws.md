# Amazon Bedrock AgentCore Identity (`aws`)

**Sources:**
- Overview — <https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/identity-overview.html>
- Security blog — <https://aws.amazon.com/blogs/security/securing-ai-agents-with-amazon-bedrock-agentcore-identity/>
- Launch blog — <https://aws.amazon.com/blogs/machine-learning/introducing-amazon-bedrock-agentcore-identity-securing-agentic-ai-at-scale/>

(Preview era, 2025; standalone service.)

## What it is

An **identity and credential management service for AI agents and automated
workloads**. Provides authentication, authorization, and credential management so
agents and tools can access AWS resources and third-party services **on behalf of
users** while keeping strict security controls and **audit trails**. Runs wherever
the agent runs — Amazon ECS, EKS, AWS Lambda, or on-premises.

## Identity model & auth

**Dual authentication:**
- **Inbound** — validates *who is invoking the agent*: AWS IAM credentials
  (**SigV4**) for AWS auth, and integration with **OAuth 2.0 / OpenID Connect**
  identity providers.
- **Outbound** — the agent accessing AWS and third-party services via **SigV4**,
  standardized **OAuth 2.0** flows, and **API keys**.

**Components:**
- **Agent identity directory** — the registry of agent identities.
- **Agent authorizer** — inbound authorization for callers invoking the agent.
- **Resource credential provider** — brokers downstream credentials.
- **Resource token vault** — stores/issues tokens so agents don't hold long-lived
  secrets.

Supports the agent acting **on behalf of a user** and **as its own identity**.

## Status

Available as a standalone service; was in preview (no additional charge through
Sept 16, 2025).

## Mapping to Open Engine

- `provider = "aws"`; `subject_id` from the agent's token `sub`. Groups follow
  the AWS shape (`cognito:groups`, already normalized by `subjectFromClaims`).
- AgentCore's **inbound auth (agent authorizer)** is the conceptual match for our
  Sentry `MCP-Session-Id`↔subject binding — validating the caller of an agent.
- Use the agent's **own identity** (not OBO) so the OE-3 groups → OPA admin
  carve-out attach to the agent. The **resource token vault** (no long-lived
  agent secrets) complements our anti-spoof binding on the audit side.
