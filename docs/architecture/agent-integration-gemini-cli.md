# Agent Integration — Gemini CLI

How Google's **Gemini CLI** (`google-gemini/gemini-cli`, the open-source terminal
agent) runs as a governed agent on the Open Engine platform. Companion to the
gold-standard [Claude Code reference](./agent-integration-claude-code.md).
Unlike Claude Code / Codex / Copilot CLI, Gemini CLI's hook event name and
output contract are **not** wire-compatible with `PreToolUse` — `tools/opa_hook.py`
needs its `--format gemini` mode. Hub: [Getting Started](../../GETTING_STARTED.md).

## D2 — Native-tool hook (OPA), verified against current upstream docs

Gemini CLI shipped a native hooks system in **v0.26.0** (2026-01-28), enabled by
default. The pre-tool event is named **`BeforeTool`** (not `PreToolUse` — early
proposals used Claude Code's name; the shipped implementation renamed it). The
stdin payload is close to Claude Code's (`tool_name`, `tool_input`, `session_id`),
but the **output** contract differs in a way that matters for policy enforcement:

- Claude/Codex/Copilot: `hookSpecificOutput.permissionDecision` ∈
  `allow | deny | ask`.
- Gemini: top-level `{"decision": "allow" | "deny", "reason": ...}` (or exit
  code 2 = deny). **There is no `ask` / escalate verdict.**

Because OPA's tri-state policy has a `require_approval` verdict with no direct
Gemini equivalent, `tools/opa_hook.py --format gemini` maps `require_approval`
to **`deny`** (fail-safe) rather than silently downgrading it to `allow` — see
`evaluate()` in `tools/opa_hook.py`. The denial reason explains that the tool
needs approval and names a hook-capable CLI (Claude Code, Copilot CLI) as the
path to get the ASK flow. This is a deliberate, documented policy choice: no
alternative in Gemini's current hook contract preserves the tri-state without a
human-in-the-loop channel the hook itself can drive.

Config file (project `.gemini/settings.json` or user `~/.gemini/settings.json`,
merged across project → user → system `/etc/gemini-cli/settings.json`):

```jsonc
{
  "hooks": {
    "BeforeTool": [
      { "matcher": ".*", "hooks": [
        { "name": "opa-gate", "type": "command",
          "command": "python3 /ABSOLUTE/path/to/agentic-harness/tools/opa_hook.py --format gemini",
          "timeout": 10000 }
      ]}
    ]
  }
}
```

(`OPA_HOOK_FORMAT=gemini` as an environment variable works the same as the
`--format gemini` flag, if you'd rather not hardcode it in the command string.)

Set `OMNIGENT_OPA_DELEGATE_MODE=enforce` (and `OMNIGENT_OPA_URL` if not local) in
the environment Gemini CLI runs in.

**MCP tools are covered too** — `BeforeTool` fires for MCP tool calls (matcher
`.*` catches them), which `tools/opa_hook.py`'s `_parse_tool()` splits into
`(server, tool)` the same way as for the other CLIs.

### Version floor and caveats

- **Minimum v0.26.0** — hooks did not exist before this release.
- **Hook timeout is 10000ms in the example above**; keep it above
  `opa_hook.py`'s internal 5s OPA request timeout so the hook's own fail-closed
  path always resolves first. Confirm current Gemini CLI timeout-failure
  semantics (open vs closed) before relying on this in `enforce` mode — verify
  against the installed version's docs rather than assuming Claude Code's
  fail-open-on-timeout behavior carries over.
- **No `ask` verdict**: any policy that leans on `require_approval` as a
  distinct UX (not just "blocked") will present as a flat deny in Gemini CLI.
  If that's unacceptable for a given deployment, don't run Gemini CLI in
  `enforce` mode for tools whose policy commonly returns `require_approval` —
  use `shadow` mode there and route those tools through a hook-capable CLI.
- Complementary non-hook controls exist if you want defense in depth alongside
  the OPA hook: `tools.core`/`tools.exclude` (allow/blocklist by tool name),
  `mcpServers.<name>.includeTools`/`excludeTools`/`trust`, and
  `general.defaultApprovalMode` (`default`/`auto_edit`/`plan`) in
  `settings.json`. These are Gemini-native and don't replace the shared OPA
  bundle as the authorization source of truth — see
  [policy-opa-hooks-audit.md](./policy-opa-hooks-audit.md)'s ban on
  non-OPA security gates.

## Gap matrix (6 dimensions)

| Dimension | Status | Mechanism / config | Gap |
| :-- | :--: | :-- | :-- |
| **D1 — Model/cost** | ◐ | No documented Cachy integration for Gemini CLI yet (unlike `cachy integrations claude/codex install`). | No managed base-URL install path. |
| **D2 — Native-tool gate (OPA hook)** | ✅ | `tools/opa_hook.py --format gemini` as the `BeforeTool` command in `settings.json`. `require_approval` fails closed to `deny` (no `ask` in Gemini's contract). `OMNIGENT_OPA_DELEGATE_MODE=enforce` fails closed on OPA-unreachable too. | `require_approval` UX is flattened to deny; see caveat above. |
| **D3 — MCP data plane (Sentry)** | ◐ | Gemini CLI's `mcpServers` config can point at the Sentry gateway as an HTTP/SSE MCP server. | Same CLO-28 gap as every other agent. |
| **D4 — Omnigent engine bridge** | ❌ | No governed-session bridge exists for Gemini CLI yet (compare `claude_native_bridge`, `codex_native_bridge`). | Follow-up if full governed-session state is needed. |
| **D5 — Agent identity (OIDC)** | ◐ | Bearer token propagation to the gateway, same pattern as other agents. | No dedicated per-agent non-admin identity yet — **D17**. |
| **D6 — Network / egress** | ◐ | OPA rules + `tools.sandbox`/`tools.sandboxNetworkAccess` (Gemini-native sandbox controls) + container egress. | No Gemini-CLI-specific egress policy beyond its own sandbox settings. |

Legend: ✅ wired by default · ◐ partial / residual gap · ❌ absent.

## Recommended governed path

1. **Native gate** — add the `BeforeTool` hook config above to
   `~/.gemini/settings.json` (or project-level `.gemini/settings.json`).
   `OMNIGENT_OPA_DELEGATE_MODE=shadow` first to observe (watch stderr for
   logged verdicts, since `--format gemini` shadow mode never emits stdout
   output — matching the other formats), then flip to `enforce`.
2. Decide, per deployment, whether tools that commonly return
   `require_approval` should run under Gemini CLI in `enforce` mode at all
   (see the no-`ask` caveat) — or be restricted to a hook-capable CLI.
3. **MCP** — point `mcpServers` at the Sentry gateway, Bearer-authed, same
   pattern as other agents pending CLO-28.
4. Smoke-test: an allowed tool proceeds silently, a denied tool is blocked
   with the OPA reason surfaced as `systemMessage`/`reason`, and a
   `require_approval` tool is also blocked (not silently allowed) — confirming
   the fail-safe mapping is actually active.

## Test

- `python3 tools/opa_hook.py --self-check` — offline assertions (OPA stubbed),
  covers both the default (`claude`) and `gemini` output formats, including the
  `require_approval → deny` fail-safe mapping.
- Live shadow: start OPA, `OMNIGENT_OPA_DELEGATE_MODE=shadow`, run a tool in
  Gemini CLI, watch stderr for the logged verdict. Then flip to `enforce`.

> Related: [opa-hook.md](./opa-hook.md) ·
> [agent-integration-copilot-cli.md](./agent-integration-copilot-cli.md) ·
> [policy-opa-hooks-audit.md](./policy-opa-hooks-audit.md) — the existing
> editor-hooks matrix already anticipated `BeforeTool` for Antigravity/Gemini CLI
> and flagged Antigravity's own PreToolUse as unverified on agy 1.0.8; this doc
> covers Gemini CLI used directly (not the Antigravity wrapper).
