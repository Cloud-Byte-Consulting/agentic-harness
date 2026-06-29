#!/usr/bin/env python3
"""Offline self-check for the healthcare tri-state policy.

Mirrors healthcare_auth.rego exactly, in the standard library only, so the demo's verdicts
can be proven with zero dependencies (no OPA, no Docker) — the same pattern as
tools/oe_correlate.py --self-check. If you change the .rego, change this in lockstep.

Run:  python3 selfcheck.py        # prints the verdict table; exits non-zero on any mismatch
"""
from __future__ import annotations
import json, os, uuid
from datetime import datetime, timezone

# Where decisions are recorded. In production this is OPA's decision log shipped to a sink;
# offline the demo appends one JSON record per call to this append-only file (override with env).
AUDIT_LOG = os.environ.get("OE_AUDIT_LOG", "decisions.jsonl")
POLICY_PATH = "data.healthcare.auth.verdict"
SERVER_NAME = "clinical-assistant"

READ_PREFIXES = ("get_", "list_", "search_", "read_", "view_")
SENSITIVE = ("phi", "patient", "record", "chart", "mrn", "ssn", "diagnosis",
             "lab", "medication", "prescription", "history")
EGRESS = ("send", "fax", "share", "export", "message", "transmit", "email",
          "external", "disclose", "post")
DENY = ("delete", "discharge", "administer", "prescribe", "order", "cancel",
        "transfer", "write", "update", "create", "merge")


# verdict -> (gateway action, HTTP status) the orchestrator maps it to.
ACTION = {"allow": ("forward", 200), "require_approval": ("HUMAN APPROVAL (ASK)", 428), "deny": ("BLOCK", 403)}


def decide(tool_name: str) -> tuple[str, str]:
    """Deterministic tri-state decision + reason. Precedence: deny > require_approval > allow > deny."""
    n = tool_name.lower()
    has = lambda toks: any(t in n for t in toks)
    is_read = n.startswith(READ_PREFIXES)

    if has(DENY):
        return "deny", "consequential clinical action"
    if has(EGRESS):
        return "require_approval", "PHI disclosure outside the covered entity"
    if is_read and has(SENSITIVE):
        return "require_approval", "read of protected health information"
    if is_read:
        return "allow", "read-prefix, no PHI"
    return "deny", "secure default - no rule matched"


def verdict(tool_name: str) -> str:
    return decide(tool_name)[0]


# (tool, expected_verdict, control rationale shown in the table)
CASES = [
    ("read_care_guidelines", "allow",            "read-prefix, no PHI"),
    ("get_patient_record",   "require_approval", "read of protected health information"),
    ("send_referral_fax",    "require_approval", "PHI disclosure outside the covered entity"),
    ("order_medication",     "deny",             "consequential clinical action"),
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
    print("PASS: all healthcare-policy verdicts correct.")
    return 0


def audit(tool_name: str, verdict_: str, action: str, code: int, reason: str) -> dict:
    """Append one decision record (OPA decision-log shape) to the append-only audit log."""
    rec = {
        "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "decision_id": str(uuid.uuid4()),
        "path": POLICY_PATH,
        "input": {"server_name": SERVER_NAME, "tool_name": tool_name, "groups": []},
        "result": verdict_,
        "action": action,
        "http_status": code,
        "reason": reason,
    }
    with open(AUDIT_LOG, "a") as f:
        f.write(json.dumps(rec) + "\n")
    return rec


def call(tool_name: str) -> int:
    """Answer one MCP tool call the way the gateway would, and write an audit record."""
    v, why = decide(tool_name)
    action, code = ACTION[v]
    rec = audit(tool_name, v, action, code, why)
    print(f"call {tool_name}\n  -> {v}   HTTP {code} {action}   ({why})")
    print(f"  audit: {rec['decision_id']} -> {AUDIT_LOG}\n")
    return 0 if v == "allow" else code


if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1:                       # `selfcheck.py <tool> ...` -> per-call gateway verdict
        for t in sys.argv[1:]:
            call(t)
        raise SystemExit(0)                      # demo path: non-zero codes are verdicts, not failures
    raise SystemExit(main())                     # no args -> verdict table (used by test.sh / CI)
