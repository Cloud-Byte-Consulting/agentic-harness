# Agent Integration — OpenAI Codex CLI (CLO-30)

How OpenAI's **Codex CLI** runs as a first-class governed agent on the Open Engine
platform (LLM proposes, OPA decides). Companion to the gold-standard
[Claude Code reference](./agent-integration-claude-code.md); shares the same six
governance dimensions. Hub: [Getting Started](../../GETTING_STARTED.md).

Codex is wired today through three real surfaces:

- **Model/cost** — `cachy integrations codex install` writes a managed
  `[model_providers.cachy]` block into `~/.codex/config.toml`
  (`Cachy/internal/install/codex.go`): `base_url = http://127.0.0.1:8787/v1`,
  `wire_api = "responses"`, `model_provider = "cachy"`. Idempotent managed block,
  preserves any existing top-level `model_provider`.
- **Native-tool hook** — Codex `PreToolUse`/`PostToolUse` command hooks call
  `omnigent/codex_native_hook.py` (or the standalone `tools/opa_hook.py`), which
  shares the policy schema with the Claude-native hook via `native_policy_hook`.
  Fails closed.
- **Engine bridge** — `omnigent/codex_native_bridge.py` runs the Codex TUI as a
  governed Omnigent session (per-session `CODEX_HOME`, bridge token, thread state).

## Gap matrix (6 dimensions)

| Dimension | Status | Mechanism / config | Gap |
| :-- | :--: | :-- | :-- |
| **D1 Model/cost (Cachy)** | ✅ | `cachy integrations codex install` → managed `[model_providers.cachy]` in `~/.codex/config.toml` (`:8787/v1`, `wire_api="responses"`) | None. Token attribution rides the Responses API. |
| **D2 Native-tool hook (OPA)** | ✅ | Codex `PreToolUse`/`PostToolUse` command hook → `codex_native_hook.py` / `tools/opa_hook.py`, fail-closed | None. |
| **D3 MCP plane (Sentry)** | ◐ | `[mcp_servers.*]` in `config.toml`; remote URL behind `[features] rmcp_client = true` | **HTTP-MCP ergonomics**: config.toml MCP is stdio-first (`command`/`args`); the remote-URL path can't inject the Sentry **Bearer** header, so reaching `:8080/mcp` needs an `mcp-remote` stdio shim. Also blocked downstream by Sentry v2 backend-MCP proxy (CLO-28). |
| **D4 Omnigent bridge** | ✅ | `codex_native_bridge.py` — governed TUI session, per-session `CODEX_HOME` | None. |
| **D5 Identity** | ◐ | Bridge propagates Bearer; needs dedicated per-agent OIDC identity | Per-cloud agent identity (D17) — see [agent-identity](./agent-identity/azure.md). |
| **D6 Network/egress** | ◐ | OPA rules + container egress | Depends on deployment network policy, same as Claude Code. |

## Recommended governed path

1. `cachy integrations codex install` — point Codex at the Cachy proxy.
2. Register `codex_native_hook.py` as `PreToolUse`/`PostToolUse` in `CODEX_HOME`
   (the bridge does this automatically; standalone users wire `tools/opa_hook.py`).
3. **Sentry MCP via stdio shim** — until D3 is closed, add the gateway as a stdio
   MCP server that injects the Bearer header:
   ```toml
   [mcp_servers.sentry]
   command = "npx"
   args = ["-y", "mcp-remote", "http://127.0.0.1:8080/mcp",
           "--header", "Authorization: Bearer ${GATEWAY_API_TOKEN}"]
   ```
   (Native remote-URL `rmcp_client` works only once Codex supports per-server auth
   headers — track upstream.)
4. Optional: run under `codex_native_bridge` for full governed-session state.
5. Smoke-test the policy plane: confirm 200 / 403 / 428 from a `tools/call`.

## Follow-up implementation

- **CLO-30-impl** — Build `cachy integrations codex-mcp` to auto-write the
  `[mcp_servers.sentry]` mcp-remote shim block (mirror the `codex.go` managed-block
  installer), so HTTP-MCP-via-Bearer is a one-command setup, not hand-edited TOML.
- **CLO-28** — Sentry v2 backend-MCP proxy (unblocks D3 native remote MCP).
- **CLO-27** — Cachy native attribution (Codex rides Responses API, already covered;
  cross-link for parity).
- **D17** — per-agent OIDC identity ([azure](./agent-identity/azure.md) ·
  [google](./agent-identity/google.md) · [aws](./agent-identity/aws.md)).
