# Agent Integration — agy (Google Antigravity, Gemini)

### 1. References & Scope
- **Spike**: CLO-32
- **Item**: `agy` — the Google Antigravity CLI (Gemini-backed TUI), integrated through omnigent's `antigravity_native_*` bridge.
- **Goal**: Gap-analysis for governing `agy` under Open Engine (LLM proposes, OPA decides). Baseline reference is [Claude Code](./agent-integration-claude-code.md); hub is [GETTING_STARTED](../../GETTING_STARTED.md).
- **Grounded surfaces**: MCP config `~/.gemini/antigravity-cli/mcp_config.json`; launch/governance facts in `omnigent/omnigent/antigravity_native_launch.py` (verified empirically in `omnigent/docs/claude/antigravity-*`).

### 2. Gap Matrix (6 Dimensions)

| Dimension | Status | Mechanism / Config | Gap |
|---|---|---|---|
| **D1 Model / Cost (Cachy)** | ❌ | agy is OAuth-only Gemini; ignores `GEMINI_API_KEY` and base-URL/proxy env knobs (`ANTIGRAVITY_SIDECAR_WEB_PORT`, `ANTIGRAVITY_EXECUTABLE_DATA_DIR` are no-ops). | Cannot point agy LLM traffic at Cachy `:8787`; no token/cost attribution for native Gemini (**CLO-27**). |
| **D2 Native-tool hook (OPA)** | ❌ | agy honors exactly one pre-emptive flag, `--dangerously-skip-permissions` (all-or-nothing); its `hooks.json` PreToolUse hook does **not** fire on tool execution (verified, agy 1.0.8). | No per-tool / web-routed gate. `opa_hook.py` cannot attach — native `run_command`/edit tools run ungoverned or fully auto-approved. |
| **D3 MCP / Sentry** | ◐ | Custom MCP servers load from `~/.gemini/antigravity-cli/mcp_config.json` (today only `linear`). Can add Sentry MCP `:8080/mcp` + Bearer. | agy's own terminal/edit tools are **not** MCP calls, so they bypass the gateway; only explicitly-added MCP servers are gated. Backend-MCP proxy still pending (**CLO-28**). |
| **D4 Omnigent bridge** | ✅ | `antigravity_native_bridge.py` drives sessions; auth inherits `~/.gemini`, workspace = process cwd, conversation id discovered from `brain/<uuid>`. | Bridge wraps the session but cannot add a tool-level gate agy itself doesn't expose. |
| **D5 Identity (D17)** | ◐ | Single shared Google OAuth login under `~/.gemini`; per-session UUID is discovered, not assigned (agy ignores `ANTIGRAVITY_CONVERSATION_ID`). | No per-agent non-admin Google identity ([agent-identity/google](./agent-identity/google.md), D17); all sessions share one human's OAuth. |
| **D6 Network / Egress** | ◐ | agy binds its own ephemeral sidecar ports; governance only possible at container/network boundary. | No app-level egress control; depends on external OPA/network policy. |

### 3. Recommended Governed Path
Because agy exposes no native-tool hook (D2) and no LLM base-URL override (D1), the only sound governed path today is **outer containment, not inner hooks**:
1. Run agy through the `antigravity_native_bridge` (D4) inside a sandboxed container under a dedicated agent identity (D5/D17).
2. Add the **Sentry MCP server** to `mcp_config.json` so any MCP tool calls route through `:8080/mcp` + OPA (`/v1/data/mcp/auth/decision`).
3. Gate agy's **native** terminal/edit tools at the **network/filesystem boundary** (egress + read-only mounts), since OPA cannot see them per-call.
4. Do **not** ship `--dangerously-skip-permissions` in any path that lacks the container boundary — it removes the one TUI prompt agy has.

### 4. Follow-up Implementation
- **CLO-27** — Cachy native Gemini attribution (D1 token/cost).
- **CLO-28** — Sentry v2 backend-MCP proxy (D3).
- **D17 / [agent-identity/google](./agent-identity/google.md)** — per-agent Google OAuth identity (D5).
- **This doc's story**: build a Cachy↔agy interception path so D1/D2 stop being ❌ (see story below) — without it, agy is only containment-governable, never per-tool-governable.

### 5. Customer Note
- **Target**: Gemini-standardized shops wanting agy's TUI in a regulated Open Engine deployment.
- **Caveat**: agy is the *least* natively governable CLI evaluated — no LLM proxy hook, no tool hook. It is safe only behind the container boundary, unlike Claude Code which gates inline.
