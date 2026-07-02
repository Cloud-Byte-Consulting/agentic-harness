# opa-hook setup — Claude Code / Codex / Copilot CLI / Gemini CLI

Copy-paste reference. For guided, step-by-step setup with detection and
verification, use the prompt: [opa-hook-setup-prompt.md](../../architecture/opa-hook-setup-prompt.md).
Full per-CLI detail: [opa-hook.md](../../architecture/opa-hook.md).

## 0. Start OPA with the bundle

Bare host process:

```bash
opa run --server Agentic-Sentry/mcp-policies/
# default :8181 — confirm oe_decision is queryable:
curl -s http://127.0.0.1:8181/v1/data/mcp/auth/oe_decision \
  -H 'Content-Type: application/json' \
  -d '{"input":{"server_name":"native","tool_name":"Bash","arguments":{},"groups":[]}}'
```

Or containerized — the only two targets this repo automates
([deploy/opa-hook/](../../../deploy/opa-hook/)):

```bash
./deploy/opa-hook/docker/run.sh                                  # Docker
minikube start && kubectl config use-context minikube \
  && ./deploy/opa-hook/minikube/deploy.sh                         # minikube
```

## Claude Code

`~/.claude/settings.json`:

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

## Codex

Register `tools/opa_hook.py` (or `codex_native_hook.py`) as the `PreToolUse`/
`PostToolUse` command hook in `CODEX_HOME` — same command as Claude Code above.
See [agent-integration-codex.md](../../architecture/agent-integration-codex.md).

## GitHub Copilot CLI (>=1.0.6 required, >=1.0.62 recommended)

`.github/hooks/opa.json` (repo) or `~/.copilot/hooks/opa.json` (personal):

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

## Gemini CLI (>=0.26.0)

`~/.gemini/settings.json` — note `--format gemini` and the different event name:

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

## Activate — shadow first, enforce second

```bash
export OMNIGENT_OPA_DELEGATE_MODE=shadow   # observe only; run a tool call, check stderr for the logged verdict
export OMNIGENT_OPA_DELEGATE_MODE=enforce  # once satisfied, flip to enforce (fails closed if OPA is unreachable)
export OMNIGENT_OPA_URL=http://127.0.0.1:8181   # only if not local/default
```

## Verify

```bash
python3 tools/opa_hook.py --self-check
```

Passes offline (OPA stubbed) for both the `claude` and `gemini` output formats.
