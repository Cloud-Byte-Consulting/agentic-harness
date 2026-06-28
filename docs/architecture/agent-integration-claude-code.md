# Agent Integration — Claude Code (CLI) — the GOLD-STANDARD baseline

> Spike **CLO-29**. Claude Code is the reference agent: all four Open Engine planes
> wire up by default, so its gap matrix is the yardstick every other agent
> (`cursor`, `codex`, `antigravity`, …) is measured against.
> Hub: [Getting Started](../../GETTING_STARTED.md) · Identity: [agent-identity/](./agent-identity/README.md)

Claude Code is Anthropic's terminal coding agent. It has native shell + filesystem
**write** access, so ungoverned it is an unbounded actor. The Open Engine value is that
every one of its consequential actions — model call, native tool, MCP `tools/call` — is
admitted or denied by the shared OPA bundle (`mcp_auth.rego` / native-plane
`data.mcp.auth.oe_decision`), never by the model. The same `claude` binary becomes the
gold standard precisely because each plane has a first-class, by-default wiring.

## Gap matrix (6 dimensions)

| Dimension | Status | Mechanism / config | Gap |
| :-- | :--: | :-- | :-- |
| **D1 — Model / cost (Cachy)** | ✅ | `cachy integrations claude install --anthropic-base-url http://127.0.0.1:8787` sets a Cachy-managed `env.ANTHROPIC_BASE_URL` in `~/.claude/settings.json` (preserves a user-set value). Native Anthropic attribution lands in Cachy/teo. | None. |
| **D2 — Native-tool gate (OPA hook)** | ✅ | `tools/opa_hook.py` wired as the `PreToolUse` command; queries `data.mcp.auth.oe_decision` on OPA `:8181`. `deny→deny`, `require_approval→ask`, else no-opinion. `OMNIGENT_OPA_DELEGATE_MODE=enforce` fails closed. | None. |
| **D3 — MCP data plane (Sentry)** | ◐ | `claude mcp add --transport http <sentry> :8080/mcp` with `Authorization: Bearer <token>`; gateway runs OIDC auth + OPA before `tools/call`, fails closed. | Sentry v2 backend-MCP proxy not yet shipped — **CLO-28**. Direct/custom backend MCP servers aren't yet fronted by the gateway. |
| **D4 — Omnigent engine bridge** | ✅ | Optional `omnigent.claude_native_bridge` wrap: stdio MCP child advertises Omnigent `sys_os_*` / per-turn tools, tmux send-keys delivers web-UI turns into the same pane. Works out of the box. | None. (Channels MCP input path blocked at org policy — bridge uses tmux instead, by design.) |
| **D5 — Agent identity (OIDC)** | ◐ | Propagates a `Bearer` token to the gateway (`OIDC_PROVIDERS` azure/google/aws or `GATEWAY_API_TOKEN`). | No dedicated per-agent non-admin identity yet — **D17**. Today the token is typically a shared/operator credential, so audit can't attribute actions to *this* agent. |
| **D6 — Network / egress** | ◐ | Egress constrained by OPA rules + the container/network the agent runs in. | No Open-Engine-owned egress policy for Claude Code specifically; depends on external container networking. |

Legend: ✅ wired by default · ◐ partial / residual gap · ❌ absent.

## Recommended governed path

Stand the four planes up in this order; each step is independently verifiable:

1. **Cost** — `cachy integrations claude install --anthropic-base-url http://127.0.0.1:8787`
   (use `--dry-run` first; uninstall is reversible and only touches the Cachy-managed block).
2. **Native gate** — set `tools/opa_hook.py` as the `PreToolUse` hook in `settings.json`,
   `OMNIGENT_OPA_DELEGATE_MODE=shadow` to observe, then `enforce`.
3. **MCP** — `claude mcp add --transport http` the Sentry gateway with a `Bearer` token.
4. **Engine (optional)** — wrap in `claude_native_bridge` for full Omnigent session governance.

Smoke-test the live policy plane end-to-end: an allowed tool returns **200**, a denied
tool **403**, a `require_approval` tool **428** (ASK). If any returns 200 under
`enforce` mode when policy says otherwise, the hook/gateway is not in the path.

## Follow-up implementation

- **CLO-28** — Sentry v2 backend-MCP proxy: closes D3 so custom/backend MCP servers are
  fronted by the gateway rather than reached directly.
- **D17** — per-agent OIDC identity: give Claude Code its own non-admin identity
  (see [agent-identity/azure.md](./agent-identity/azure.md) ·
  [google.md](./agent-identity/google.md) · [aws.md](./agent-identity/aws.md)) so D5/D6
  audit attributes actions to this agent, not a shared operator token.
- The concrete code task that *makes this doc true and keeps it true* — an automated
  one-shot wiring + smoke-test command — is filed as the implementation story below.

> Related: [opa-hook.md](./opa-hook.md) · [policy-opa-hooks-audit.md](./policy-opa-hooks-audit.md) ·
> [governed-platform-architecture.md](./governed-platform-architecture.md) ·
> [eval-agentgateway-mcp-plane.md](./eval-agentgateway-mcp-plane.md)
