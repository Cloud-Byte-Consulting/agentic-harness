# opa-hook — standalone OPA gate for editors (Phase 3b)

`tools/opa_hook.py` is the Open Engine enforcement point for editors that use a
native `PreToolUse` hook but run **without** an Omnigent session (Claude Code,
Codex, GitHub Copilot, Antigravity used directly). It queries the **same** shared
OPA bundle as the Sentry gateway and the Omnigent native hook — the native-plane
`data.mcp.auth.oe_decision` rule — so the OE boundaries are enforced uniformly.
**One policy, multiple enforcement points** (Sentry = MCP plane; native hook =
Omnigent-wrapped sessions; opa-hook = bare editors).

stdlib only, no Omnigent/Sentry import — it ships standalone with any editor.

## Prerequisites

- **OPA running with the bundle** so `oe_decision` is queryable, e.g.:
  `opa run --server Agentic-Sentry/mcp-policies/` (default `:8181`).
- Env:
  - `OMNIGENT_OPA_URL` — OPA base URL (default `http://127.0.0.1:8181`).
  - `OMNIGENT_OPA_DELEGATE_MODE` — `off` (default) / `shadow` / `enforce` (shared
    with the Omnigent native hook).

## Verdict → editor permission

| OPA `oe_decision` verdict | Editor `permissionDecision` |
| :--- | :--- |
| `deny` (delete / credentials / billing) | **`deny`** (+ reason) |
| `require_approval` (publish / email / deploy) | **`ask`** — the editor prompts the user (the harness's own prompt is the human channel; no Omnigent ASK gate needed for a bare editor) |
| `allow` / mode `off` | no output ("no opinion" — the editor's own permission system still runs) |

**Fail-closed:** in `enforce`, an unreachable OPA, an unparseable payload, or an
unknown verdict → `deny`. In `off`/`shadow`, it never blocks.

## Claude Code / Codex setup

Add to the editor's `settings.json` (Claude Code shown; Codex uses the same
`PreToolUse` payload contract):

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

## Other editors

- **GitHub Copilot** — same `PreToolUse` command-hook contract.
- **Antigravity** — `PreToolUse` in `hooks.json`. **Re-verify on agy 2.0** before
  relying on it: the Omnigent audit found PreToolUse did **not** fire on agy 1.0.8
  (see `policy-opa-hooks-audit.md`). Until verified, treat Antigravity as
  MCP-only (Sentry gateway).

## Test

- `python3 tools/opa_hook.py --self-check` — offline assertions (OPA stubbed).
- Live shadow: start OPA, `OMNIGENT_OPA_DELEGATE_MODE=shadow`, run a tool in the
  editor, watch stderr for the logged verdict. Then flip to `enforce`.
