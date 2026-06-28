# Agent Integration — Devin (Cognition)

**Spike:** CLO-34 · **Type:** cli (cloud-hosted remote agent) · **Status:** exploratory
Hub: [Getting Started](../../GETTING_STARTED.md) · Baseline: [Claude Code](agent-integration-claude-code.md) · Identity: [agent-identity/](agent-identity/README.md)

## What it is — and why it's different

Devin is **not a local CLI**. It is a fully **cloud-hosted autonomous engineer** run inside
Cognition's own infrastructure: each Devin session executes in a Cognition-managed VM/sandbox,
reached by the customer through the web app, Slack, or the Devin API. The agent loop, the
model calls, and the native shell/editor all live **on Cognition's side of the network** — the
customer never hosts the runtime. That single fact decides which Open Engine planes are even
*reachable*, before any policy question.

The consequence: our two in-process control points — Cachy base-URL override (D1) and the
`opa_hook.py` PreToolUse gate (D2) — **cannot be injected**, because we don't run the process.
The only plane that can reach Devin is the **MCP data plane**, and only if (a) Devin Enterprise
exposes remote-MCP-server config and (b) our Agentic-Sentry gateway is published on a
public, egress-reachable URL with Bearer auth.

## Gap matrix

| Dimension | Status | Mechanism / config | Gap |
| :-- | :--: | :-- | :-- |
| **D1 Model / Cost (Cachy)** | ❌ | Devin uses Cognition-managed frontier models; no customer base-URL override. LLM traffic never transits Cachy `:8787`. | No BYO-LLM proxy hook. Spend/attribution invisible to us. Open question: does the Devin API surface any model-routing or proxy setting at all? |
| **D2 Native-tool hook (opa_delegate)** | ❌ | Native Bash/Edit/Write run inside Cognition's VM. No PreToolUse extension point; we don't own the process. | `opa_hook.py` / `*_native_bridge.py` are structurally unreachable. Native actions are ungoverned by us. |
| **D3 MCP / Sentry** | ◐ | Devin Enterprise advertises remote MCP server support. Point Devin at a **publicly-exposed** Sentry `:8080/mcp` with `GATEWAY_API_TOKEN` (or OIDC Bearer). Gateway fails closed. | Unverified that Devin honours per-call Bearer + streamable-HTTP MCP. Still blocked by CLO-28 (v2 backend-MCP proxy). Requires exposing the gateway to the public internet. |
| **D4 Omnigent engine bridge** | ❌ | Omnigent runs agents *we* host; Devin's loop is Cognition's. Can't wrap it. | Out of scope unless Devin is treated purely as an MCP *client* of our engine. |
| **D5 Identity (D17)** | ◐ | Devin authenticates as a Cognition machine identity / API key. Maps to one shared non-human identity at our gateway via its Bearer token. | No per-cloud OIDC identity (azure/google/aws). Cannot satisfy D17 "agent under its own non-admin identity" without a dedicated issued token per Devin workspace. |
| **D6 Network / Egress** | ◐ | Reachability flips from "localhost" to "public ingress". Sentry + OPA must terminate TLS on an internet-facing endpoint; allowlist Cognition egress ranges. | New attack surface. Need IP allowlist + mTLS/Bearer; depends on Cognition publishing stable egress CIDRs. |

Legend: ✅ wired · ◐ partial/conditional · ❌ not reachable.

## Recommended governed path

Treat Devin as an **external MCP client of the policy gateway**, not as a harness we instrument:

1. Publish Agentic-Sentry `:8080/mcp` behind public TLS ingress (own subdomain), fail-closed.
2. Issue a **dedicated Bearer/OIDC token per Devin workspace** (not the admin token) so the
   gateway can attribute and scope its `tools/call` traffic — partial D17.
3. Restrict ingress to Cognition's published egress ranges; deny all else at the edge.
4. Register the Sentry MCP endpoint in Devin Enterprise's MCP settings with that Bearer.
5. Accept that D1/D2/D4 stay dark: model spend and native shell actions on Devin's VM are
   **outside our governance**. Govern only what crosses our MCP boundary.

This means Devin is governed **only at the MCP edge** — a strictly weaker posture than a
locally-hosted harness (Claude Code, Codex, Cursor) where native actions are also gated.

## Open questions (exploratory)

- Does Devin Enterprise actually support **per-server Bearer auth + streamable-HTTP MCP**, or
  only stdio/SSE? (Determines whether D3 is real or vapor.)
- Are Cognition's **egress CIDRs** stable and published for ingress allowlisting?
- Any Devin API hook for **model routing / cost export** that could give D1 a foothold?
- Contractual: is exposing an internal policy gateway to a third-party cloud acceptable to the
  customer's threat model, vs. running an MCP relay in the customer VPC?

## Follow-up implementation

- **CLO-34a** — Validate Devin Enterprise remote-MCP against a public Sentry instance: confirm
  Bearer pass-through and the 200/403/428 tri-state round-trips (see code story below).
- Depends on **CLO-28** (Sentry v2 backend-MCP proxy) and **D17** (per-agent identity).
- Related: [CLO-27](agent-integration-claude-code.md) attribution gap is *worse* here (no Cachy at all).
