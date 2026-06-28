# GCP Agent Identity — benefits & customer scenarios (`google`)

Companion to [google.md](google.md) (the mechanics) and the
[Getting Started hub](../../../GETTING_STARTED.md). This doc is the
**"why a customer picks it"** view for **GCP IAM Agent Identity / Vertex AI
Agent Engine / Workload Identity Federation (WIF)**.

## What it is (one line)

A per-agent, **SPIFFE/X.509-attested** cloud identity: each agent gets its own
cryptographic principal, IAM-authorized like any GCP resource, with **access
tokens bound to the agent's certificate** so a stolen token is useless off the
agent. (Preview.)

## Why a customer picks it

- **No long-lived keys to leak.** No service-account JSON files for an agent to
  hold (or have stolen) — the historic GCP footgun. WIF already does this for
  CI/external workloads; Agent Identity extends it per-agent.
- **Token theft is cryptographically dead, not just audited.** The X.509
  certificate binding is the strongest anti-theft of the three clouds — Entra
  leans on Conditional Access, AWS on a token vault; only GCP binds the token to
  the cert itself.
- **Least-privilege the IAM admins already know.** Agent permissions are plain
  IAM bindings on the agent principal — same allow/deny, conditions, and Cloud
  Audit Logs surface the team runs for humans. No new authz system to learn.
- **Native MCP + A2A.** GCP agents authenticate to **MCP servers** and to other
  agents (Agent2Agent protocol) out of the box — the exact governed-path shape
  Open Engine wants.

## Customer scenarios

1. **GCP-native shop, Vertex AI Agent Engine.** A team runs agents on Agent
   Engine Runtime. Each agent gets its own SPIFFE identity, IAM-scoped to only
   the BigQuery datasets / GCS buckets it needs. A leaked token can't be replayed
   from an attacker's box — it's cert-bound.

2. **Killing service-account-key sprawl.** A customer has dozens of agents
   sharing one over-privileged SA and its downloaded key. Migrating to per-agent
   identity removes the key entirely and gives per-agent Cloud Audit Log lines —
   "which agent touched this table" instead of "the shared SA did."

3. **Multi-agent / A2A workflow.** A planner agent calls specialist agents via
   A2A; each hop is mutually authenticated by SPIFFE identity, and each specialist
   is IAM-scoped independently. No shared god-credential across the fleet.

4. **Hybrid / non-GCP agent reaching GCP via WIF.** An agent running off-GCP
   federates into GCP through Workload Identity Federation (OIDC), getting
   short-lived GCP tokens with no stored key — the bridge for customers not yet
   fully on Agent Engine.

## Fit & limits for Open Engine

- **Fit.** `provider = "google"`; `subject_id` from the agent token, groups via
  the Google `groups` claim (normalized by `subjectFromClaims` in
  `Agentic-Sentry/internal/auth/auth.go`). Use **own-authority**, not OBO, so the
  OE-3 groups → OPA admin carve-out attaches to the *agent* (decision D17). The
  cert-bound token is complementary to — not a replacement for — Sentry's
  `MCP-Session-Id`↔subject binding: GCP stops token *theft*, our binding stops
  session-id *replay* for the three-plane audit join.
- **Limits.** Preview surface, GCP-only, and the X.509 binding protects the token
  toward *GCP* — it does **not** by itself authenticate the agent to our Sentry
  gateway. Today our gateway verifies a Bearer OIDC token; to consume a GCP agent
  token directly, the `google` OIDC provider must be configured in
  `OIDC_PROVIDERS` with the right issuer/audience, and we still need the WIF/Agent
  Engine token to land in the `Authorization` header our middleware reads.

## Follow-ups

- The implementation story below: make the `google` provider actually accept a
  Vertex/WIF agent token end-to-end through Sentry.
- Cross-cloud: keeps parity with [azure.md](azure.md) / [aws.md](aws.md); the
  shared D17 wiring lives in [README.md](README.md).
- Related open gaps: **CLO-27** (Cachy native Gemini attribution), **CLO-28**
  (gateway v2 backend-MCP proxy).
