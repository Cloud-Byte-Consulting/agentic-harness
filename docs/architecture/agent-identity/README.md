# Cloud agent-identity models (per OIDC provider)

Reference for how each cloud IdP models **AI-agent identities**, and how Open
Engine uses them. One file per OIDC provider — keyed to the provider names our
gateway resolves (`azure` / `google` / `aws` / `okta` / `generic`, see
`Agentic-Sentry/internal/auth/auth.go` `subjectFromClaims`).

| Provider | Product | File |
| :--- | :--- | :--- |
| `azure` | **Microsoft Entra Agent ID** | [azure.md](azure.md) |
| `aws` | **Amazon Bedrock AgentCore Identity** | [aws.md](aws.md) |
| `google` | **GCP Agent Identity (IAM / Vertex AI Agent Engine)** | [google.md](google.md) |
| `okta` / `generic` | No dedicated agent-identity product surveyed | see note below |

## The common pattern

All three converge on the same shape:

| | Entra Agent ID | AWS AgentCore Identity | GCP Agent Identity |
| :--- | :--- | :--- | :--- |
| Per-agent distinct id | Entra agent account | agent identity directory | SPIFFE / X.509 per agent |
| Own rights vs on-behalf-of | autonomous / delegated | inbound/outbound, OBO | own-authority / OBO |
| Anti-theft | token policies + Conditional Access | token vault (no long-lived creds) | **token bound to X.509** (strongest) |
| Caller auth (inbound) | Entra-token verify | SigV4 / OAuth inbound | mutual SPIFFE / OIDC |
| Audit | "logged as an AI agent" | audit trails | per-agent IAM logs |

**A distinct, least-privilege, lifecycle-bound identity per agent, with its own
claims, plus a verified inbound-auth path so a downstream can trust *who* is
calling.**

## How Open Engine uses this (decision D17)

- **Run each OE agent-code under its own dedicated, NON-admin agent identity** —
  not the human operator's identity. Use the **autonomous** (own-rights) model,
  not delegated/on-behalf-of.
- Then OE-3's group plumbing carries the **agent's** groups: the OPA admin
  carve-out (`is_admin`) keys on the *agent*, and `subject_id` in the three-plane
  audit join is the *agent's* object id. This resolves the standing caveat that
  "an agent running under an admin identity inherits the bypass."
- The Sentry **`MCP-Session-Id`↔subject binding** (`cmd/gateway/main.go`,
  `bindSession`) is exactly the **inbound-auth / "reliably identify the caller"**
  primitive all three platforms describe — the audit-side counterpart to the
  native-plane groups work.

## `okta` / `generic`

No dedicated "agent identity" product was surveyed for these. Treat agents as
standard OIDC principals: provision a per-agent client/identity, ensure its token
carries a `groups`/`roles` claim (consumed by our auth → OPA admin carve-out per
OE-3), and keep it out of admin groups by default. Okta's `groups` and AWS
`cognito:groups` claim shapes are already normalized by `subjectFromClaims`.
