# Agent Integration — GitHub Copilot CLI

How **GitHub Copilot CLI** (the terminal agent shipped GA 2026-02-25 — not the
older `gh copilot suggest`/`explain` extension) runs as a governed agent on the
Open Engine platform. Companion to the gold-standard
[Claude Code reference](./agent-integration-claude-code.md); shares the same
`PreToolUse` hook contract, so it reuses `tools/opa_hook.py` unmodified. Hub:
[Getting Started](../../GETTING_STARTED.md).

## D2 — Native-tool hook (OPA), verified against current upstream docs

Copilot CLI's native hook events use camelCase names (`preToolUse`, `postToolUse`,
`permissionRequest`, …), but since **v1.0.6** the CLI also accepts Claude Code's
hook config shape verbatim: PascalCase event names (`PreToolUse`), the nested
`matcher`/`hooks` structure, snake_case stdin fields (`hook_event_name`,
`tool_name`, `tool_input`, `session_id`), and the `hookSpecificOutput` output
wrapper — exactly what `tools/opa_hook.py` already emits. No format flag or
adapter is needed; it is wired the same way as Claude Code.

Config file (repo-level `.github/hooks/opa.json`, or personal
`~/.copilot/hooks/opa.json` / `$COPILOT_HOME/hooks/opa.json`):

```jsonc
{
  "version": 1,
  "hooks": {
    "PreToolUse": [
      { "matcher": "*", "hooks": [
        { "type": "command", "command": "python3 /ABSOLUTE/path/to/agentic-harness/tools/opa_hook.py",
          "timeoutSec": 10 }
      ]}
    ]
  }
}
```

Set `OMNIGENT_OPA_DELEGATE_MODE=enforce` (and `OMNIGENT_OPA_URL` if not local) in
the environment the CLI runs in.

**MCP tools are covered too** — `preToolUse`/`PreToolUse` fires for
`mcp__<server>__<tool>` calls (matcher `*` catches them; a narrower matcher like
`mcp__.*` also works), which `tools/opa_hook.py`'s `_parse_tool()` already splits
into `(server, tool)` for the OPA query.

### Version floor and caveats

- **Minimum ~1.0.6** for the Claude-compatible config/payload contract at all;
  **1.0.24+** if you also want `updatedInput` (arg-rewriting hooks) respected —
  not used by `opa_hook.py`, but relevant if you extend it.
- **1.0.62+ recommended**: before this, a hook crash or unexpected non-zero exit
  did not reliably fail closed; 1.0.62 hardened that. `opa_hook.py` itself never
  raises past its own `try/except` (falls back to "no opinion" or fail-closed
  deny per mode), but a broken Python environment (missing interpreter, syntax
  error from a bad edit) is the failure mode this version floor protects against.
- **Hook timeouts fail OPEN** (default 30s), unlike a hook crash. Keep
  `timeoutSec` comfortably above `opa_hook.py`'s internal 5s OPA request timeout
  so the fail-closed path inside the hook always resolves before the CLI's own
  timeout would silently allow the tool through.
- Windows: hook execution needs PowerShell 7+ if using the `powershell` key;
  the `command` key (used above) is the cross-platform alias and works
  everywhere `python3` is on PATH.

## Gap matrix (6 dimensions)

| Dimension | Status | Mechanism / config | Gap |
| :-- | :--: | :-- | :-- |
| **D1 — Model/cost** | ◐ | Not yet a documented Cachy integration for Copilot CLI (unlike `cachy integrations claude/codex install`). | No managed base-URL install path; track alongside D1 for other agents. |
| **D2 — Native-tool gate (OPA hook)** | ✅ | `tools/opa_hook.py` as the `PreToolUse` command in `.github/hooks/` or `~/.copilot/hooks/`; Claude-compatible wire format since 1.0.6. `OMNIGENT_OPA_DELEGATE_MODE=enforce` fails closed. | Pin CLI version ≥1.0.62 for hardened fail-closed-on-crash behavior. |
| **D3 — MCP data plane (Sentry)** | ◐ | Copilot CLI's own MCP config can point at the Sentry gateway as an HTTP MCP server with a Bearer header. | Same CLO-28 gap as every other agent — Sentry v2 backend-MCP proxy not yet shipped. |
| **D4 — Omnigent engine bridge** | ❌ | No `copilot_native_bridge` equivalent exists yet (compare `claude_native_bridge`, `codex_native_bridge`). | Follow-up: build one if full governed-session state is needed for Copilot CLI. |
| **D5 — Agent identity (OIDC)** | ◐ | Bearer token propagation to the gateway, same pattern as other agents. | No dedicated per-agent non-admin identity yet — **D17**. |
| **D6 — Network / egress** | ◐ | OPA rules + container/network egress, same as other agents. | No Copilot-CLI-specific egress policy. |

Legend: ✅ wired by default · ◐ partial / residual gap · ❌ absent.

## Recommended governed path

1. **Native gate** — drop the hook config above in `.github/hooks/opa.json` (repo,
   loads after folder trust) or `~/.copilot/hooks/opa.json` (personal, always
   loads). `OMNIGENT_OPA_DELEGATE_MODE=shadow` to observe, then `enforce`.
2. **MCP** — point Copilot CLI's MCP config at the Sentry gateway, Bearer-authed,
   same as the Codex stdio-shim pattern until CLO-28 lands.
3. Smoke-test: an allowed tool proceeds silently, a denied tool is blocked with
   the OPA reason surfaced, a `require_approval` tool prompts the user (`ask`).
   If a denied tool runs anyway under `enforce`, the hook is not in the path —
   check the CLI version and the hook config's load location.

## Test

- `python3 tools/opa_hook.py --self-check` — offline assertions (OPA stubbed),
  covers both the default (`claude`) and `gemini` output formats.
- Live shadow: start OPA, `OMNIGENT_OPA_DELEGATE_MODE=shadow`, run a tool in
  Copilot CLI, watch stderr for the logged verdict. Then flip to `enforce`.

> Related: [opa-hook.md](./opa-hook.md) ·
> [agent-integration-gemini-cli.md](./agent-integration-gemini-cli.md) ·
> [policy-opa-hooks-audit.md](./policy-opa-hooks-audit.md)
