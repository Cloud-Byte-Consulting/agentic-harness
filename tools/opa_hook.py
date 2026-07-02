#!/usr/bin/env python3
"""opa-hook — standalone OPA PreToolUse gate for editors run WITHOUT Omnigent.

Phase 3b. Editors that expose a native pre-tool hook (Claude Code, Codex, GitHub
Copilot CLI, Gemini CLI, Antigravity) but are NOT wrapped in an Omnigent session
still need the Open Engine boundaries enforced. This CLI is that enforcement
point: it reads a pre-tool hook payload on stdin, queries the SAME shared OPA
bundle that Sentry and the Omnigent native hook use (the native-plane
``data.mcp.auth.oe_decision`` rule), and emits the harness's permission decision.
One policy, multiple enforcement points.

Wire it as the editor's pre-tool command (see
``docs/open-engine/templates/opa-hook-setup.md`` and, per-editor,
``docs/architecture/agent-integration-*.md``). Reads stdin, writes the
hook-output JSON to stdout; emits nothing for "no opinion".

Output format (``--format`` flag or ``OPA_HOOK_FORMAT`` env; default ``claude``):
- ``claude`` — Claude Code, Codex, and GitHub Copilot CLI (>=1.0.6) all speak this
  wire format: event name ``PreToolUse``, output
  ``hookSpecificOutput.permissionDecision``. Verdict mapping:
  - OPA ``deny``             -> ``"deny"`` (+ reason)
  - OPA ``require_approval`` -> ``"ask"``  — the editor prompts the user. For a
    standalone editor the harness's own prompt IS the human channel, so unlike
    the Omnigent-session path no server-side ASK gate is needed here.
  - OPA ``allow`` / off      -> no output ("no opinion"; the editor's own
    permission system still runs).
- ``gemini`` — Gemini CLI (>=0.26.0): event name ``BeforeTool``, output
  ``{"decision": "allow"|"deny", "reason": ...}``. Gemini hooks have **no**
  ``ask``/escalate verdict, so ``require_approval`` maps to ``deny`` (fail-safe)
  with a reason explaining the tool needs approval via a hook-capable CLI.

Mode (``OMNIGENT_OPA_DELEGATE_MODE``, shared with the native hook):
- ``off`` (default) -> never queries OPA; no opinion. Zero behaviour change.
- ``shadow``        -> queries + logs to stderr; not enforced.
- ``enforce``       -> deny/ask as above; OPA unreachable / unknown verdict ->
  fail closed (deny).

stdlib only; no Omnigent/Sentry import, so it ships standalone with any editor.
``OMNIGENT_OPA_URL`` overrides the OPA base URL (default http://127.0.0.1:8181).
"""

from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.request

_MODE_ENV = "OMNIGENT_OPA_DELEGATE_MODE"
_URL_ENV = "OMNIGENT_OPA_URL"
_FORMAT_ENV = "OPA_HOOK_FORMAT"
_DEFAULT_OPA = "http://127.0.0.1:8181"
_DECISION_PATH = "/v1/data/mcp/auth/oe_decision"
_VALID_MODES = ("off", "shadow", "enforce")
_VALID_FORMATS = ("claude", "gemini")
_DEFAULT_FORMAT = "claude"
_PRE = "PreToolUse"                # Claude Code / Codex / Copilot CLI event name
_PRE_GEMINI = "BeforeTool"         # Gemini CLI event name


def _mode() -> str:
    m = os.environ.get(_MODE_ENV, "off").strip().lower()
    return m if m in _VALID_MODES else "off"


def _format() -> str:
    """``--format gemini|claude`` (argv) wins over ``OPA_HOOK_FORMAT`` (env)."""
    if "--format" in sys.argv:
        idx = sys.argv.index("--format")
        if idx + 1 < len(sys.argv):
            f = sys.argv[idx + 1].strip().lower()
            if f in _VALID_FORMATS:
                return f
    f = os.environ.get(_FORMAT_ENV, _DEFAULT_FORMAT).strip().lower()
    return f if f in _VALID_FORMATS else _DEFAULT_FORMAT


def _parse_tool(name: str) -> tuple[str, str]:
    """``mcp__server__tool`` -> (server, tool); a host tool -> ("native", tool)."""
    if name.startswith("mcp__"):
        parts = name.split("__")
        if len(parts) >= 3:
            return parts[1], "__".join(parts[2:])
    return "native", name


def _query_opa(server: str, tool: str, args: object) -> dict | None:
    """POST the tool call to OPA's oe_decision; return the decision or None.

    Returns None on ANY failure (unreachable, non-2xx, bad JSON, no verdict) so
    the caller fails closed in enforce mode. groups=[] here — a standalone editor
    has no bound subject groups, so the strict boundary applies (fail-safe).
    """
    base = (os.environ.get(_URL_ENV) or _DEFAULT_OPA).rstrip("/")
    body = json.dumps(
        {
            "input": {
                "server_name": server,
                "tool_name": tool,
                "arguments": args if isinstance(args, dict) else {},
                "groups": [],
            }
        }
    ).encode()
    req = urllib.request.Request(
        base + _DECISION_PATH, data=body, headers={"Content-Type": "application/json"}, method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            data = json.loads(resp.read())
    except Exception as exc:  # noqa: BLE001 — any failure -> None -> fail closed
        print(f"opa-hook: OPA query failed: {exc}", file=sys.stderr)
        return None
    result = data.get("result") if isinstance(data, dict) else None
    if not isinstance(result, dict) or "verdict" not in result:
        print("opa-hook: OPA returned no decision", file=sys.stderr)
        return None
    return result


def _decision(kind: str, reason: str, fmt: str = _DEFAULT_FORMAT) -> dict:
    """``kind`` is always 'deny' or 'ask' — callers never pass 'ask' when fmt == 'gemini'."""
    if fmt == "gemini":
        return {"decision": kind, "reason": reason}
    return {
        "hookSpecificOutput": {
            "hookEventName": _PRE,
            "permissionDecision": kind,
            "permissionDecisionReason": reason,
        }
    }


def evaluate(payload: dict, fmt: str = _DEFAULT_FORMAT) -> dict | None:
    """Return the hook-output dict, or None for 'no opinion'."""
    mode = _mode()
    if mode == "off":
        return None
    expected_event = _PRE_GEMINI if fmt == "gemini" else _PRE
    if payload.get("hook_event_name") != expected_event:
        return None
    tool_name = payload.get("tool_name") or ""
    if not tool_name:
        return None
    server, tool = _parse_tool(tool_name)
    decision = _query_opa(server, tool, payload.get("tool_input") or {})
    if decision is None:
        if mode == "enforce":
            return _decision("deny", "OPA policy evaluation unavailable; failing closed (opa-hook enforce).", fmt)
        print(f"opa-hook[shadow]: OPA unavailable for {tool_name!r}", file=sys.stderr)
        return None
    verdict = str(decision.get("verdict"))
    if mode == "shadow":
        print(f"opa-hook[shadow]: tool={tool_name!r} verdict={verdict}", file=sys.stderr)
        return None
    reason = decision.get("reason") or f"Open Engine boundary ({verdict}) for {tool_name}"
    if verdict == "deny":
        return _decision("deny", reason, fmt)
    if verdict == "require_approval":
        if fmt == "gemini":
            # No ask/escalate verdict in Gemini's hook contract — fail safe to deny
            # rather than silently downgrading to allow.
            return _decision(
                "deny",
                f"{reason} — requires human approval. Gemini CLI hooks have no ask "
                "verdict, so this fails closed; approve out-of-band or use a "
                "hook-capable CLI with an ASK flow (Claude Code, Copilot CLI).",
                fmt,
            )
        return _decision("ask", reason, fmt)
    if verdict == "allow":
        return None
    return _decision("deny", f"opa-hook: unknown OPA verdict {verdict!r}; failing closed", fmt)


def main() -> None:
    fmt = _format()
    try:
        payload = json.loads(sys.stdin.read() or "{}")
    except json.JSONDecodeError:
        if _mode() == "enforce":
            sys.stdout.write(
                json.dumps(_decision("deny", "opa-hook: unparseable hook payload; failing closed", fmt))
            )
        return
    if not isinstance(payload, dict):
        return
    out = evaluate(payload, fmt)
    if out is not None:
        sys.stdout.write(json.dumps(out))


def _self_check() -> None:
    """Assert-based smoke test (OPA stubbed — no network). Run with --self-check."""
    global _query_opa
    assert _parse_tool("mcp__github__delete_repository") == ("github", "delete_repository")
    assert _parse_tool("Bash") == ("native", "Bash")
    assert _format() == "claude"

    os.environ[_MODE_ENV] = "off"
    assert evaluate({"hook_event_name": _PRE, "tool_name": "Bash", "tool_input": {}}) is None

    orig = _query_opa
    os.environ[_MODE_ENV] = "enforce"

    # --- claude/codex/copilot format (default) ---
    _query_opa = lambda s, t, a: {"verdict": "deny", "reason": "boundary"}
    assert evaluate({"hook_event_name": _PRE, "tool_name": "mcp__github__delete_repository"})["hookSpecificOutput"]["permissionDecision"] == "deny"
    _query_opa = lambda s, t, a: {"verdict": "require_approval", "reason": "ask?"}
    assert evaluate({"hook_event_name": _PRE, "tool_name": "mcp__x__publish_y"})["hookSpecificOutput"]["permissionDecision"] == "ask"
    _query_opa = lambda s, t, a: {"verdict": "allow"}
    assert evaluate({"hook_event_name": _PRE, "tool_name": "Bash"}) is None
    _query_opa = lambda s, t, a: None  # OPA down -> fail closed
    assert evaluate({"hook_event_name": _PRE, "tool_name": "Bash"})["hookSpecificOutput"]["permissionDecision"] == "deny"
    _query_opa = lambda s, t, a: {"verdict": "weird"}  # unknown -> fail closed
    assert evaluate({"hook_event_name": _PRE, "tool_name": "Bash"})["hookSpecificOutput"]["permissionDecision"] == "deny"

    assert evaluate({"hook_event_name": "PostToolUse", "tool_name": "Bash"}) is None
    # wrong event name for this format -> no opinion, even with a matching payload otherwise
    assert evaluate({"hook_event_name": _PRE_GEMINI, "tool_name": "Bash"}) is None

    # --- gemini format ---
    _query_opa = lambda s, t, a: {"verdict": "deny", "reason": "boundary"}
    out = evaluate({"hook_event_name": _PRE_GEMINI, "tool_name": "Bash"}, fmt="gemini")
    assert out == {"decision": "deny", "reason": "boundary"}
    _query_opa = lambda s, t, a: {"verdict": "require_approval", "reason": "ask?"}
    out = evaluate({"hook_event_name": _PRE_GEMINI, "tool_name": "mcp__x__publish_y"}, fmt="gemini")
    assert out["decision"] == "deny"  # no ask verdict in gemini -> fail closed, not allow
    assert "ask?" in out["reason"] and "human approval" in out["reason"]
    _query_opa = lambda s, t, a: {"verdict": "allow"}
    assert evaluate({"hook_event_name": _PRE_GEMINI, "tool_name": "Bash"}, fmt="gemini") is None
    # claude-format event name is rejected when fmt == "gemini"
    assert evaluate({"hook_event_name": _PRE, "tool_name": "Bash"}, fmt="gemini") is None

    _query_opa = orig

    sys.argv[1:] = ["--format", "gemini"]
    assert _format() == "gemini"
    sys.argv[1:] = ["--self-check"]
    assert _format() == "claude"

    print("opa-hook self-check passed")


if __name__ == "__main__":
    if "--self-check" in sys.argv:
        _self_check()
    else:
        main()
