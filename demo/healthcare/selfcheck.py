#!/usr/bin/env python3
"""Offline self-check for the healthcare tri-state policy.

Configuration only: the decision logic, action map, verdict table, per-call gateway simulator,
and audit writer all live in ../common/policy_engine.py (shared with the other demo slices, so
they cannot drift). This file supplies the healthcare token lists, reasons, and expected cases —
the stdlib mirror of healthcare_auth.rego. If you change the token lists or precedence here,
change healthcare_auth.rego in lockstep; test.sh cross-checks the two when the `opa` binary is present.

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
    name="healthcare",
    server_name="clinical-assistant",
    # Tokens that make a READ touch protected health information (PHI) -> human gate.
    sensitive=("phi", "patient", "record", "chart", "mrn", "ssn", "diagnosis",
               "lab", "medication", "prescription", "history"),
    # Tokens that disclose PHI OUTSIDE the covered entity (egress) -> human gate.
    egress=("send", "fax", "share", "export", "message", "transmit", "email",
            "external", "disclose", "post"),
    # Consequential / destructive clinical or data actions -> hard deny.
    deny=("delete", "discharge", "administer", "prescribe", "order", "cancel",
          "transfer", "write", "update", "create", "merge"),
    reasons={
        "deny": "consequential clinical action",
        "egress": "PHI disclosure outside the covered entity",
        "sensitive_read": "read of protected health information",
        "read": "read-prefix, no PHI",
        "default": "secure default - no rule matched",
    },
    cases=(
        ("read_care_guidelines", "allow"),
        ("get_patient_record", "require_approval"),
        ("send_referral_fax", "require_approval"),
        ("order_medication", "deny"),
    ),
    # Decisions land next to this script (stable regardless of CWD, and covered by .gitignore).
    # The durable, tamper-evident trail is OPA's decision log shipped to a sink in production;
    # this local file is a visibility aid only — it is not itself append-only.
    audit_log=os.environ.get("OE_AUDIT_LOG", os.path.join(_HERE, "decisions.jsonl")),
)


if __name__ == "__main__":
    run_cli(POLICY, sys.argv)
