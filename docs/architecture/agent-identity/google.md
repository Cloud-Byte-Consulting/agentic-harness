# GCP Agent Identity (`google`)

**Sources:**
- Agent Identity overview (IAM) — <https://docs.cloud.google.com/iam/docs/agent-identity-overview>
- Authenticate using an agent's own authority — <https://docs.cloud.google.com/iam/docs/auth-agent-own-identity>
- Use agent identity with Vertex AI Agent Engine — <https://cloud.google.com/agent-builder/agent-engine/agent-identity>

(Preview.)

## What it is

A **strongly attested, cryptographic identity for each agent, based on the
SPIFFE standard**, with per-agent **X.509 certificates**. The agent can securely
authenticate to **MCP servers**, cloud resources, endpoints, and **other agents**,
acting **on its own behalf or on behalf of an end user**.

## Why it's not a service account

- **Not shared** by multiple workloads by default.
- **Cannot be impersonated**; **no long-lived service-account keys**.
- **Access tokens are cryptographically bound to the agent's X.509 certificate**
  to prevent token theft/replay — the strongest anti-theft of the three clouds.
- **Per-agent**, least-privilege, **tied to the agent's lifecycle** — "a more
  secure principal than service accounts."

## Auth & authorization

- **Own authority** — the agent uses its **primary SPIFFE identity** to request
  Google Cloud access tokens, then accesses first-party tools/APIs/resources
  (authorized via IAM on the agent identity), including other agents on Vertex AI
  Agent Engine via the **Agent2Agent (A2A) protocol**.
- **On behalf of a user** — alternative delegated model.
- Recommended (Preview) for **Vertex AI Agent Engine Runtime**.

## Mapping to Open Engine

- `provider = "google"`; `subject_id` from the agent's token; groups via the
  Google `groups` claim (normalized by `subjectFromClaims`).
- The **X.509-bound token** is a stronger anti-theft mechanism than (and
  complementary to) our Sentry `MCP-Session-Id`↔subject binding — GCP prevents
  token *theft* cryptographically; our binding prevents session-id *replay* for
  the audit join.
- GCP agents already authenticate to **MCP servers** natively — aligns with the
  Open Engine model where agents call MCP tools through the governed path.
- Use **own-authority** (not OBO) so the OE-3 groups → OPA admin carve-out attach
  to the agent.
