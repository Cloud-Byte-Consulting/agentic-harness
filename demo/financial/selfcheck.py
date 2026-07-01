#!/usr/bin/env python3
"""Offline self-check for the financial tri-state policy.

Configuration only: the decision logic, action map, verdict table, per-call gateway simulator,
and audit writer all live in ../common/policy_engine.py (shared with the other demo slices, so
they cannot drift). This file supplies the financial token lists, reasons, and expected cases —
the stdlib mirror of financial_auth.rego. If you change the token lists or precedence here,
change financial_auth.rego in lockstep; test.sh cross-checks the two when the `opa` binary is present.

Run:  python3 selfcheck.py        # prints the verdict table; exits non-zero on any mismatch
      python3 selfcheck.py <tool>  # per-call gateway verdict + audit record
"""
from __future__ import annotations

import os
import sys

# Import the shared engine from the sibling demo/common/ directory (CWD-independent).
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "common"))
from policy_engine import Policy, run_cli  # noqa: E402

_HERE = os.path.dirname(os.path.abspath(__file__))

POLICY = Policy(
    name="financial",
    server_name="member-assistant",
    # Tokens that make a READ touch confidential member data -> human gate (declassification).
    sensitive=("pii", "ssn", "account", "balance", "member_data", "statement", "card"),
    # Tokens that move data OUT of the institution (egress) -> human gate.
    egress=("send", "email", "share", "export", "message", "post", "external"),
    # Consequential / destructive tokens -> hard deny (segregation-of-duties boundary).
    deny=("delete", "transfer", "move_funds", "wire", "pay", "close",
          "write", "update", "create", "merge"),
    reasons={
        "deny": "consequential / destructive boundary",
        "egress": "possible data egress",
        "sensitive_read": "read of confidential member data",
        "read": "read-prefix, no confidential data",
        "default": "secure default - no rule matched",
    },
    cases=(
        ("read_education_content", "allow"),
        ("get_member_account", "require_approval"),
        ("send_external_message", "require_approval"),
        ("transfer_funds", "deny"),
    ),
    # Decisions land next to this script (stable regardless of CWD, and covered by .gitignore).
    # The durable, tamper-evident trail is OPA's decision log shipped to a sink in production;
    # this local file is a visibility aid only — it is not itself append-only.
    audit_log=os.environ.get("OE_AUDIT_LOG", os.path.join(_HERE, "decisions.jsonl")),
)


if __name__ == "__main__":
    run_cli(POLICY, sys.argv)
