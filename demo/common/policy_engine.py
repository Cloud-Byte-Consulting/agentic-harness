#!/usr/bin/env python3
"""Shared tri-state policy engine for the demo slices (financial, healthcare, …).

One implementation of the decision logic, the gateway action map, the verdict-table
self-check, the per-call gateway simulator, and the (optional) audit-log writer. Each slice's
`selfcheck.py` supplies only its token lists, reason strings, expected cases, and labels via a
`Policy` and calls `run_cli(policy)`. This is the stdlib mirror of every slice's `*.rego` — keep
the token lists and precedence here in lockstep with each Rego; `test.sh` cross-checks a slice's
Rego against its `--expect` output when the `opa` binary is present.

Matching is boundary-aware: a token matches a whole underscore-delimited run of the tool name,
never a raw substring — so `order` does not fire on `search_disorder_guidelines` and `lab` does
not fire on `list_available_appointments`. Multi-word tokens (`move_funds`, `member_data`) are
matched as whole runs too. The slice `*.rego` files use the identical padded-underscore match.

stdlib only; no third-party imports, so the demos stay zero-dependency.
"""
from __future__ import annotations

import json
import os
import uuid
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Sequence

# Read-only prefixes (low-risk, non-mutating) — mirrors the production read-prefix set.
READ_PREFIXES = ("get_", "list_", "search_", "read_", "view_")

# verdict -> (gateway action, HTTP status) the orchestrator maps it to. Shared by all slices.
ACTION = {
    "allow": ("forward", 200),
    "require_approval": ("HUMAN APPROVAL (ASK)", 428),
    "deny": ("BLOCK", 403),
}

# reason keys a slice must provide, one per branch of decide().
REASON_KEYS = ("deny", "egress", "sensitive_read", "read", "default")


@dataclass(frozen=True)
class Policy:
    """A slice's configuration. All logic lives in this module; slices are pure data."""

    name: str                       # e.g. "healthcare" — names the policy path data.<name>.auth.verdict
    server_name: str                # MCP server_name recorded in audit inputs
    sensitive: Sequence[str]        # read + one of these -> require_approval
    egress: Sequence[str]           # any of these -> require_approval (data leaves the boundary)
    deny: Sequence[str]             # any of these -> hard deny (top precedence)
    reasons: dict                   # keys == REASON_KEYS
    cases: Sequence[tuple]          # (tool, expected_verdict) pairs asserted by the self-check
    audit_log: str | None = None    # absolute path to append decision records to; None -> no audit

    def __post_init__(self) -> None:
        missing = [k for k in REASON_KEYS if k not in self.reasons]
        if missing:
            raise ValueError(f"Policy {self.name!r} missing reason keys: {missing}")

    @property
    def policy_path(self) -> str:
        return f"data.{self.name}.auth.verdict"


def _matches(padded_name: str, tokens: Sequence[str]) -> bool:
    """True if any token equals a whole underscore-delimited run of the tool name.

    `padded_name` is the lowercased tool name wrapped in underscores, so a token wrapped in
    underscores is a substring iff it aligns to segment boundaries (handles multi-word tokens).
    """
    return any(f"_{t}_" in padded_name for t in tokens)


def decide(policy: Policy, tool_name: str) -> tuple[str, str]:
    """Deterministic tri-state decision + reason. Precedence: deny > require_approval > allow > deny."""
    n = tool_name.lower()
    padded = f"_{n}_"
    is_read = n.startswith(READ_PREFIXES)

    if _matches(padded, policy.deny):
        return "deny", policy.reasons["deny"]
    if _matches(padded, policy.egress):
        return "require_approval", policy.reasons["egress"]
    if is_read and _matches(padded, policy.sensitive):
        return "require_approval", policy.reasons["sensitive_read"]
    if is_read:
        return "allow", policy.reasons["read"]
    return "deny", policy.reasons["default"]


def audit(policy: Policy, tool_name: str, verdict: str, action: str, code: int, reason: str) -> dict:
    """Build one decision record (OPA decision-log shape); append it to the local log if enabled."""
    rec = {
        "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "decision_id": str(uuid.uuid4()),
        "path": policy.policy_path,
        "input": {"server_name": policy.server_name, "tool_name": tool_name, "groups": []},
        "result": verdict,
        "action": action,
        "http_status": code,
        "reason": reason,
    }
    if policy.audit_log:
        with open(policy.audit_log, "a") as f:
            f.write(json.dumps(rec) + "\n")
    return rec


def table(policy: Policy) -> int:
    """Print the verdict table for the slice's CASES; return non-zero on any mismatch."""
    print(f"{'tool':<24}{'verdict':<18}{'why'}")
    print("-" * 72)
    failures = 0
    for tool, expected in policy.cases:
        got, why = decide(policy, tool)
        ok = got == expected
        failures += not ok
        flag = "" if ok else f"  <-- MISMATCH (expected {expected})"
        print(f"{tool:<24}{got:<18}{why}{flag}")
    print("-" * 72)
    if failures:
        print(f"FAIL: {failures} verdict mismatch(es).")
        return 1
    print(f"PASS: all {policy.name}-policy verdicts correct.")
    return 0


def call(policy: Policy, tool_name: str) -> int:
    """Answer one MCP tool call the way the gateway would; write an audit record if enabled."""
    v, why = decide(policy, tool_name)
    action, code = ACTION[v]
    rec = audit(policy, tool_name, v, action, code, why)
    print(f"call {tool_name}\n  -> {v}   HTTP {code} {action}   ({why})")
    if policy.audit_log:
        # Show a CWD-relative path so the demo reads cleanly (e.g. "decisions.jsonl") no matter
        # where the file physically lives.
        print(f"  audit: {rec['decision_id']} -> {os.path.relpath(policy.audit_log)}")
    print()
    return 0 if v == "allow" else code


def run_cli(policy: Policy, argv: Sequence[str]) -> None:
    """Entry point for a slice's `selfcheck.py`. `argv` is sys.argv (script name at [0])."""
    if len(argv) > 1:
        if argv[1] == "--expect":            # `selfcheck.py --expect` -> `tool expected` lines
            for tool, expected in policy.cases:  # single source of truth for test.sh's opa assertions
                print(f"{tool} {expected}")
            raise SystemExit(0)
        for t in argv[1:]:                   # `selfcheck.py <tool> ...` -> per-call gateway verdict
            call(policy, t)
        raise SystemExit(0)                  # demo path: non-zero codes are verdicts, not failures
    raise SystemExit(table(policy))          # no args -> verdict table (used by test.sh / CI)
