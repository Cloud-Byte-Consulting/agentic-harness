# Agent integration: Cursor

**Spike**: CLO-31 · **Type**: `cli` (IDE agent) · See the [getting-started hub](../../GETTING_STARTED.md) and the [Claude Code baseline](agent-integration-claude-code.md) (gold-standard reference).

Cursor is an MCP-capable IDE agent with **native** filesystem/terminal write access plus its own cloud features (Tab, background agent). `air` already treats it as a first-class harness target — rules project to `.cursor/rules/agents.mdc` and skills project to `.cursor/rules/` (`air/internal/agents/agents.go`, `air/internal/targets/targets.go`). The governance question: how much of Cursor's tool reach actually routes through the three Open Engine planes (Cachy, Sentry/OPA, omnigent).

## Gap matrix (6 dimensions)

| Dimension | Status | Mechanism / config | Gap |
| :--- | :---: | :--- | :--- |
| **D1 Model / cost (Cachy)** | ◐ | Settings → Models → *Custom OpenAI base URL* pointed at Cachy `:8787/v1` + key | **No `cachy integrations cursor install`** — base URL set by hand; Cursor's own cloud features (Tab, background agent, indexing) call Cursor's backend directly and **bypass Cachy** entirely, so attribution is partial |
| **D2 Native-tool hook (OPA)** | ◐ | Cursor hooks (`.cursor/hooks.json`): `beforeShellExecution`, `beforeMCPExecution`, `beforeReadFile`, `afterFileEdit` | `opa_hook.py` is shaped for Claude Code's `PreToolUse` JSON — **needs a Cursor-hook adapter** + verification that shell **and** file-edit events are actually gated (native edits may slip the gate) |
| **D3 MCP / Sentry** | ✅ | `.cursor/mcp.json` supports `"transport":"http"` + `headers` → Sentry `:8080/mcp` with `Authorization: Bearer …` | Same backend-MCP proxy limitation as everyone: **CLO-28** (gateway v2 → backend MCP) |
| **D4 Omnigent bridge** | ✅ | omnigent runs Cursor as a governed runtime (cursor relay) → native-plane `opa_delegate` covers what IDE hooks miss | — |
| **D5 Identity (D17)** | ◐ | Bearer carried in `.cursor/mcp.json` `headers` | Static **plaintext secret in a repo file**; needs per-agent non-admin OIDC identity ([agent-identity](agent-identity/README.md), D17) and a secret-ref instead of an inline token |
| **D6 Network / egress** | ◐ | OPA decision + container networking | Cursor's mandatory cloud calls (indexing, Tab) make full egress containment harder than for a pure-CLI agent |

## Recommended governed path

1. **MCP today** — add Sentry to `.cursor/mcp.json` (`http` + `Authorization: Bearer`). Fails closed; this is the strongest lever available now.
2. **Wrap native reach** — run the Cursor session under the **omnigent cursor relay** so the native-plane `opa_delegate` governs shell/edit tools the IDE hooks don't cover (D4 closes the D2 gap until the hook adapter ships).
3. **Cost** — manually point Cursor's custom model base URL at Cachy `:8787/v1`. Accept that Cursor-native cloud features stay unattributed until D1 auto-integration lands.
4. **Identity** — issue a dedicated non-admin OIDC identity per agent (D17) and reference the token by secret, not inline.

## Follow-up implementation

- **`cachy integrations cursor install`** — auto-wire `.cursor/mcp.json` (Sentry http+header) **and** the model base URL, with a secret-ref Bearer (closes D1 + D5 plaintext gap). *See implementation story below.*
- **Cursor-hook adapter for `opa_hook.py`** — translate `.cursor/hooks.json` `beforeShellExecution`/`beforeMCPExecution`/`afterFileEdit` events to the OPA decision call; verify native-edit coverage (D2).
- **Native-vs-MCP reach audit** — enumerate which Cursor tools route through Sentry vs. stay native, to size what omnigent must wrap (D6).
- Cross-refs: CLO-28 (gateway v2 proxy), D17 (agent identity), and the [eval-actplane-native-plane](eval-actplane-native-plane.md) split.
