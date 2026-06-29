#!/usr/bin/env python3
"""Offline self-check for the financial tri-state policy.

Mirrors financial_auth.rego exactly, in the standard library only, so the demo's verdicts
can be proven with zero dependencies (no OPA, no Docker) — the same pattern as
tools/oe_correlate.py --self-check. If you change the .rego, change this in lockstep.

Run:  python3 selfcheck.py        # prints the verdict table; exits non-zero on any mismatch
"""
from __future__ import annotations
import sys

READ_PREFIXES = ("get_", "list_", "search_", "read_", "view_")
SENSITIVE = ("pii", "ssn", "account", "balance", "member_data", "statement", "card")
EGRESS = ("send", "email", "share", "export", "message", "post", "external")
DENY = ("delete", "transfer", "move_funds", "wire", "pay", "close",
        "write", "update", "create", "merge")


def verdict(tool_name: str) -> str:
    """Deterministic tri-state decision. Precedence: deny > require_approval > allow > deny."""
    n = tool_name.lower()
    has = lambda toks: any(t in n for t in toks)
    is_read = n.startswith(READ_PREFIXES)

    if has(DENY):
        return "deny"
    if has(EGRESS):
        return "require_approval"
    if is_read and has(SENSITIVE):
        return "require_approval"
    if is_read and not has(SENSITIVE):
        return "allow"
    return "deny"  # secure default


# (tool, expected_verdict, control rationale shown in the table)
CASES = [
    ("read_education_content", "allow",            "read-prefix, no confidential data"),
    ("get_member_account",     "require_approval", "read of confidential member data"),
    ("send_external_message",  "require_approval", "possible data egress"),
    ("transfer_funds",         "deny",             "consequential / destructive boundary"),
]


def main() -> int:
    print(f"{'tool':<24}{'verdict':<18}{'why'}")
    print("-" * 72)
    failures = 0
    for tool, expected, why in CASES:
        got = verdict(tool)
        ok = got == expected
        failures += not ok
        flag = "" if ok else f"  <-- MISMATCH (expected {expected})"
        print(f"{tool:<24}{got:<18}{why}{flag}")
    print("-" * 72)
    if failures:
        print(f"FAIL: {failures} verdict mismatch(es).")
        return 1
    print("PASS: all financial-policy verdicts correct.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
