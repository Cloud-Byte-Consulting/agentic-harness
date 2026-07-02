# opa-hook — standalone OPA gate for editors (Phase 3b)

`tools/opa_hook.py` is the Open Engine enforcement point for editors that use a
native pre-tool hook but run **without** an Omnigent session (Claude Code,
Codex, GitHub Copilot CLI, Gemini CLI, Antigravity used directly). It queries
the **same** shared OPA bundle as the Sentry gateway and the Omnigent native
hook — the native-plane `data.mcp.auth.oe_decision` rule — so the OE boundaries
are enforced uniformly. **One policy, multiple enforcement points** (Sentry =
MCP plane; native hook = Omnigent-wrapped sessions; opa-hook = bare editors).

stdlib only, no Omnigent/Sentry import — it ships standalone with any editor.
Two output wire formats are supported (`--format claude|gemini`, default
`claude`) since Claude Code, Codex, and Copilot CLI share one contract while
Gemini CLI uses a different event name and a binary (no `ask`) decision shape —
see the per-CLI docs linked below for exact configs and version floors.

## Prerequisites

- **OPA running with the bundle** so `oe_decision` is queryable, e.g.:
  `opa run --server Agentic-Sentry/mcp-policies/` (default `:8181`).
- Env:
  - `OMNIGENT_OPA_URL` — OPA base URL (default `http://127.0.0.1:8181`).
  - `OMNIGENT_OPA_DELEGATE_MODE` — `off` (default) / `shadow` / `enforce` (shared
    with the Omnigent native hook).

## Verdict → editor permission

| OPA `oe_decision` verdict | `--format claude` (Claude Code / Codex / Copilot CLI) | `--format gemini` (Gemini CLI) |
| :--- | :--- | :--- |
| `deny` (delete / credentials / billing) | **`deny`** (+ reason) | **`deny`** (+ reason) |
| `require_approval` (publish / email / deploy) | **`ask`** — the editor prompts the user (the harness's own prompt is the human channel; no Omnigent ASK gate needed for a bare editor) | **`deny`** (+ reason) — Gemini's hook contract has **no `ask`/escalate verdict**, so this fails safe to deny rather than silently downgrading to allow |
| `allow` / mode `off` | no output ("no opinion" — the editor's own permission system still runs) | no output |

**Fail-closed:** in `enforce`, an unreachable OPA, an unparseable payload, or an
unknown verdict → `deny`, in both formats. In `off`/`shadow`, it never blocks.

## Claude Code / Codex / Copilot CLI setup

Add to the editor's `settings.json` / hook config (Claude Code shown; Codex and
GitHub Copilot CLI ≥1.0.6 accept the identical `PreToolUse` payload/output
contract — see [agent-integration-codex.md](./agent-integration-codex.md) and
[agent-integration-copilot-cli.md](./agent-integration-copilot-cli.md) for
exact file locations and version floors):

```jsonc
{
  "hooks": {
    "PreToolUse": [
      { "matcher": "*", "hooks": [
        { "type": "command", "command": "python3 /ABSOLUTE/path/to/agentic-harness/tools/opa_hook.py" }
      ]}
    ]
  }
}
```

Set `OMNIGENT_OPA_DELEGATE_MODE=enforce` (and `OMNIGENT_OPA_URL` if not local) in
the editor's environment to activate.

## Gemini CLI setup

Gemini CLI (≥v0.26.0) uses a different event name (`BeforeTool`) and a binary
decision shape with no `ask` verdict, so it needs `--format gemini`. Add to
`~/.gemini/settings.json` or project `.gemini/settings.json`:

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

Full details, caveats, and the `require_approval → deny` fail-safe rationale:
[agent-integration-gemini-cli.md](./agent-integration-gemini-cli.md).

## Other editors

- **Antigravity** — `PreToolUse` in `hooks.json`. **Re-verify on agy 2.0** before
  relying on it: the Omnigent audit found PreToolUse did **not** fire on agy 1.0.8
  (see `policy-opa-hooks-audit.md`). Until verified, treat Antigravity as
  MCP-only (Sentry gateway).

## Test

- `python3 tools/opa_hook.py --self-check` — offline assertions (OPA stubbed),
  covering both `claude` and `gemini` output formats.
- Live shadow: start OPA, `OMNIGENT_OPA_DELEGATE_MODE=shadow`, run a tool in the
  editor, watch stderr for the logged verdict. Then flip to `enforce`.

> Per-CLI setup detail: [agent-integration-claude-code.md](./agent-integration-claude-code.md) ·
> [agent-integration-codex.md](./agent-integration-codex.md) ·
> [agent-integration-copilot-cli.md](./agent-integration-copilot-cli.md) ·
> [agent-integration-gemini-cli.md](./agent-integration-gemini-cli.md)
