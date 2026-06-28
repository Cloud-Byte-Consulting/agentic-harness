# AWS serverless containers for Open Engine (App Runner / Lambda-container / Fargate — no EKS)

> Spike **CLO-26**. How to run the Open Engine control plane on AWS *container*
> compute without operating a Kubernetes cluster. See the
> **[Getting Started hub](../../GETTING_STARTED.md)** for the plane topology, and
> the AWS **[agent-identity](../agent-identity/aws.md)** doc for who the agents
> authenticate *as* once deployed.

## What it is

Three AWS ways to run a container image with **no node/cluster to manage** — all
are "serverless" in that AWS owns the host, patching, and scaling:

| Compute | Shape | Scales | Networking |
| :-- | :-- | :-- | :-- |
| **App Runner** | one HTTP image, fully managed | auto, **to-zero-ish** (min 1 provisioned, idle billing) | public/VPC, TLS + ALB-free URL built in |
| **Lambda (container)** | one image as a request/response function, ≤15 min, ≤10 GB | **to zero**, per-request | invoke via API GW / Function URL; no listener |
| **Fargate (ECS)** | a **task = N containers** (sidecars, localhost), long-running | task count autoscale (no to-zero) | ALB/NLB, VPC, service discovery |

All three skip EKS: no control-plane fee, no `kubectl`, no k8s RBAC/CRDs to learn.

## Why a customer picks it (benefits)

- **No cluster tax** — no EKS control-plane cost, no node patching, no k8s
  on-call. The whole reason this spike exists: most OE adopters are a few
  services, not a platform team.
- **Scale managed for you** — App Runner/Lambda scale on traffic and idle cheap;
  Fargate right-sizes per task.
- **Per-customer / single-tenant deploys are trivial** — one Terraform/CDK stack
  per tenant, no shared cluster blast radius.
- **Native AWS identity** — task/function execution roles are real IAM principals,
  which is exactly the SigV4 inbound/outbound story in
  [agent-identity/aws.md](../agent-identity/aws.md) (AgentCore Identity).

## Customer scenarios

1. **Single-tenant governed gateway, no platform team.** Customer wants the
   Sentry MCP gateway + OPA in front of their agents but has no k8s. → **Fargate
   task** with `gateway` + `opa` as two containers in one task (the
   `docker-compose.yml` topology maps 1:1 to a task definition; `OPA_URL` stays
   `localhost`). Behind an ALB, TLS via ACM. Fails closed exactly as today.
2. **Cost-optimization proxy for a small team.** Just want **Cachy** (LLM proxy
   `:8787`) in front of provider APIs. → **App Runner** on the Cachy image — one
   public HTTPS service, autoscaled, no infra. Cheapest path to value.
3. **Policy-as-a-function for an existing app.** Customer already has their own
   agent host and only wants the **OPA decision** (`/v1/data/mcp/auth/decision`)
   as a callable. → **OPA in a Lambda container** behind a Function URL — pure
   stateless eval, scales to zero, pay per decision.
4. **Burst eval / batch governance jobs.** Run the experiment harnesses
   (`eval-*`) or a nightly policy regression as short container runs. →
   **Lambda-container** or **Fargate one-shot task**, not a standing service.

## Fit & limits for Open Engine

- **Fargate = the default for the control plane.** The OE planes are
  **long-running networked servers with sidecars** — Sentry `gateway` (`:8080`)
  + OPA (`:8181`) bundle-poller, Cachy (`:8787`), omnigent engine (`:6767`).
  Fargate's multi-container task is the only one of the three that preserves the
  `gateway`↔`opa` **localhost sidecar** pattern in one unit. Pick it over EKS
  whenever the service count is small and you don't need k8s-native data-plane
  injection (agentgateway extProc / Gateway API — see
  [eval-agentgateway-mcp-plane.md](../eval-agentgateway-mcp-plane.md)).
- **App Runner = one image only.** No sidecars: OPA can't ride along, so either
  bundle OPA into the gateway image or run OPA as a *second* App Runner service
  (extra hop, lose localhost). Good for the single-container planes (Cachy,
  token-dashboard), awkward for gateway+OPA.
- **Lambda = stateless decisions only.** The gateway and Cachy are **persistent
  streaming listeners** (MCP SSE, LLM streaming) — Lambda's request/response +
  15-min cap + cold starts make them a **poor fit**. OPA-as-decision-function is
  the clean Lambda use.
- **Swap the dev deps.** The compose ships **garage** (S3-compat) for the OPA
  policy bundle; on AWS that becomes **real S3** (OPA already speaks the S3
  bundle protocol — only env/creds change). `GATEWAY_API_TOKEN` /
  `*_CLIENT_SECRET` move to **Secrets Manager**; task role replaces static keys.
- **Fail-closed is preserved** on all three — the gateway denies when OPA is
  unreachable regardless of compute host. No new trust assumption.
- **EKS still wins when:** many services + service mesh, k8s-native MCP injection
  (agentgateway as a Gateway API extProc), or multi-tenant at cluster scale.

## Follow-ups

- **Implementation:** ship reusable **Fargate/ECS deploy artifacts** for the
  gateway+OPA task with S3-backed bundle and Secrets Manager (story below).
- Confirm AgentCore Identity SigV4 inbound maps onto an ALB/Function-URL auth
  path (ties to [agent-identity/aws.md](../agent-identity/aws.md), D17).
- Relates to **CLO-28** (gateway v2 backend-MCP proxy) — backend MCPs reached
  from a Fargate task need VPC egress / PrivateLink, not yet specified.
</content>
</invoke>
