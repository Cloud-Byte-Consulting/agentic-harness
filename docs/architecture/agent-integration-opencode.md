# Agent Integration — OpenCode (CLI) · CLO-33

**Spike:** CLO-33 · **Type:** cli gap-analysis · **Status:** discovery
**Hub:** [Getting Started](../../GETTING_STARTED.md) · **Sibling baseline:** [Claude Code (gold standard)](./agent-integration-claude-code.md)

## What it is

[OpenCode](https://opencode.ai) is an open-source terminal coding agent. Omnigent already ships a **mature `opencode-native` harness** (merged PR #576, live-verified against `opencode serve` 1.17.7): the runner spawns `opencode serve`, an SSE forwarder mirrors events into the Omnigent session, and a typed HTTP client injects prompts. Unlike claude/codex (which take `HARNESS_*_GATEWAY_*` env vars), **OpenCode reads its provider, MCP, permission, and plugin config from its own `opencode.json`** under a per-session `XDG_CONFIG_HOME` — so every Open Engine seam is a config-synthesis question, not an env-var question.

Real surfaces (`/home/bittahcriminal/air/workspace/omnigent/omnigent/`): `opencode_native_bridge.py`, `opencode_native_provider.py`, `opencode_native_permissions.py`, `opencode_native_forwarder.py`, `opencode_http_transport.py`, `opencode_native_app_server.py` · design: `omnigent/designs/opencode-native-gaps.md`.

## Gap matrix (6 dimensions)

| Dimension | Status | Mechanism / config | Gap |
|---|---|---|---|
| **D1 Model / base-URL** | ◐ | `opencode_native_provider.py` synthesizes a `@ai-sdk/openai-compatible` provider in `opencode.json` (`baseURL`+`apiKey`). `OpenCodeGatewayResolution` is **already generic** (any OpenAI-compatible base URL). | Hardwired to the **Databricks gateway** (`resolve_databricks_gateway`). No resolver points `baseURL` at **Cachy `:8787/v1`**, so model traffic skips token/cost attribution. |
| **D2 Native-tool / policy hook** | ✅ | First-class plugin lifecycle: `write_opencode_policy_plugin` emits `omnigent-policy.js` bridging `chat.message`→`PHASE_REQUEST` and `tool.execute.after`→`PHASE_TOOL_RESULT`; `permission.asked`→`TOOL_CALL` via `opencode_native_permissions.py`. | Routes to **Omnigent `/policies/evaluate`**, not Sentry/OPA `/v1/data/mcp/auth/decision` directly. Tool-name coverage partial (`block_skills`/github/google name-sets); fails **open** on transport error. |
| **D3 MCP / Sentry** | ◐ | Native `opencode.json` `mcp` block — `local` (stdio) / `remote` (url+headers) — plus `permission: "ask"` routes MCP tool calls through the policy engine. | Not pointed at **Agentic-Sentry `:8080/mcp`** (would be a `remote` entry + `Authorization: Bearer`). Same backend-MCP proxy blocker as **CLO-28**. |
| **D4 Omnigent bridge** | ✅ | Full native-server harness present and live-verified (1.17.7): forwarder, typed client, HTTP+SSE transport, provider, permissions, app-server. Compaction/cost/resume/fork/elicitation all built. | None for the bridge itself. |
| **D5 Identity** | ❌ | Provider `apiKey` and the policy-plugin auth token are **launch snapshots** (Databricks profile / `OMNIGENT_*` env on `opencode serve`); written `0600` into per-session XDG. | No per-agent **OIDC identity (D17)**; tokens don't refresh — long sessions degrade to fail-open / re-resolve only on resume. |
| **D6 Network / egress** | ❓ | `opencode serve` is runner-owned on loopback; egress depends on container netns + OPA. | **Unknown** — not assessed for opencode specifically; flag for a netns/egress spike. |

## Recommended governed path

The harness is done; only the **gateway/identity seams** need rewiring — all in `opencode.json` synthesis, no new transport code:

1. **D1 (Cachy):** add a `resolve_cachy_gateway` peer to `resolve_databricks_gateway` that returns an `OpenCodeGatewayResolution` with `base_url = {CACHY}/v1` + per-agent Bearer. Reuse `build_opencode_provider_config` / the existing atomic-`0600` writer verbatim. **This is the one real code gap (the implementation story below).**
2. **D3 (Sentry):** add the Sentry MCP as a `remote` entry in `build_opencode_mcp_block` (`url = {SENTRY}:8080/mcp`, `Authorization: Bearer`); keep `permission: "ask"` so calls still hit the policy plane. Unblocks with CLO-28.
3. **D2 (OPA):** keep the `omnigent-policy.js` plugin as the enforcement seam; align its `/policies/evaluate` backend to the Sentry/OPA decision (`opa_delegate.py`) so OpenCode shares the bundle (`mcp_auth.rego`).
4. **D5 (identity):** issue a per-agent OIDC token (D17) and write it to a **refreshable token file** instead of a launch snapshot.

## Follow-up implementation

- **CLO-33-impl** (below): Cachy resolver for opencode-native (D1).
- **CLO-28**: Sentry v2 backend-MCP proxy — unblocks D3.
- **D17**: per-agent OIDC identity + refreshable token file — D5.
- **Net-egress spike**: assess D6 (opencode `serve` loopback + container netns).
- Cross-refs: [agent-identity (azure/google/aws)](./agent-identity/) · [Claude Code baseline](./agent-integration-claude-code.md) · [OPA hooks](./opa-hook.md) · [opencode-native-gaps design](../../../omnigent/designs/opencode-native-gaps.md).
