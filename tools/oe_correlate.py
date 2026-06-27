#!/usr/bin/env python3
"""Open Engine three-plane audit correlation query (consolidation Phase 5 / CP-4).

Resolves one task_ref -> one session_id, then fans out to the three audit planes
(authz / runtime / cost) keyed by session_id and emits a single correlated trace
per docs/architecture/audit-correlation-design.md §2-§4.

Stdlib only. Read-side query: it adds no source dependency on any satellite
(CP-4 criterion 5 is structural — this file imports nothing from Sentry/omnigent/
token-dashboard/teo, it only reads their *outputs*: a JSONL decision log, a saved
`GET /sessions/{id}` snapshot, and the tracker-comment receipt markers).

Contracts (pinned, do not invent variants):
  A) authz decision log  — one JSON object per line, schema:
       {"ts","plane","session_id","subject_id","server_name","tool_name","verdict","reason"}
  B) receipt marker      — HTML comment in CLAIMED/DONE tracker comments:
       <!-- openengine session=<id> task_ref=<provider>:<id> phase=claimed|done -->
  C) correlated trace    — {task_ref, session_id, subject_id, subject_email,
                            authz[], runtime[], cost{}}

Run the bundled CP-4 self-check (no network, fixtures under testdata/cp4/):
    python3 oe_correlate.py --self-check
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
import urllib.request

# Contract B: parse the receipt marker, then split key=value tokens.
_MARKER_RE = re.compile(r"<!--\s*openengine\s+(.*?)\s*-->")


class CorrelationError(Exception):
    """A CP-4 invariant failed. Caller exits non-zero."""


def _marker_tokens(blob: str) -> list[dict[str, str]]:
    """Every openengine marker in `blob` as a dict of its key=value tokens."""
    out = []
    for body in _MARKER_RE.findall(blob):
        kv = {}
        for tok in body.split():
            if "=" in tok:
                k, v = tok.split("=", 1)
                kv[k] = v
        out.append(kv)
    return out


def resolve_session_id(task_ref: str, *, session_id: str | None,
                       comments: str | None) -> tuple[str, list[dict[str, str]]]:
    """task_ref -> (session_id, receipt-markers-for-this-task).

    Priority: explicit --session-id, else the receipt marker (contract B).
    The Omnigent label query (openengine.issue=<task_ref>) is the documented
    fallback; it is not wired here because CP-4 runs offline over fixtures.

    CP-4 #1 (exactly one) + #4 (no cross-contamination at resolution): if the
    comments carry markers for >1 distinct session under this task_ref and no
    explicit session was pinned, that is ambiguous -> hard error, never a guess.
    """
    markers = [m for m in _marker_tokens(comments or "") if m.get("task_ref") == task_ref]
    if session_id:
        return session_id, markers
    candidates = sorted({m["session"] for m in markers if "session" in m})
    if not candidates:
        raise CorrelationError(
            f"no session_id resolves for task_ref={task_ref!r} "
            f"(no --session-id and no receipt marker in --comments-file)")
    if len(candidates) > 1:
        raise CorrelationError(
            f"task_ref={task_ref!r} resolves to {len(candidates)} sessions "
            f"{candidates} — ambiguous (CP-4 #4: one issue -> one trace)")
    return candidates[0], markers


def authz_rows(session_id: str, log_text: str) -> list[dict]:
    """Contract-A JSONL filtered to this session. Each kept row -> trace row.

    Enforces CP-4 #3 (single subject across all authz rows) and #4 (a decoy
    row for another session is dropped, never leaked).
    """
    rows, subjects = [], set()
    for line in log_text.splitlines():
        line = line.strip()
        if not line:
            continue
        rec = json.loads(line)
        if rec.get("session_id") != session_id:  # #4: drop other sessions
            continue
        subj = rec.get("subject_id", "")
        if subj:  # empty subject = native-plane "unknown" (pre-OE-3), not an actor
            subjects.add(subj)
        rows.append({
            "tool": rec.get("tool_name"),
            "verdict": rec.get("verdict"),
            "subject_id": rec.get("subject_id", ""),
            "subject_email": rec.get("subject_email", ""),  # contract-A optional
            "ts": rec.get("ts"),
        })
    if len(subjects) > 1:
        raise CorrelationError(
            f"authz rows for session {session_id} carry {len(subjects)} distinct "
            f"non-empty subject_ids {sorted(subjects)} — one actor per session violated "
            f"(CP-4 #3). Empty subject_ids (native-plane rows pre-OE-3) are unknown, "
            f"not a second actor.")
    return rows


def load_snapshot(session_id: str, *, snapshot_file: str | None,
                  omnigent_url: str | None, token: str | None) -> dict:
    """A `GET /sessions/{id}` snapshot, from a fixture file or live fetch.

    CP-4 #4: the snapshot's own id must equal the resolved session — a snapshot
    for a different session must not be joined in.
    """
    if snapshot_file:
        with open(snapshot_file, encoding="utf-8") as fh:
            snap = json.load(fh)
    elif omnigent_url:
        req = urllib.request.Request(
            f"{omnigent_url.rstrip('/')}/v1/sessions/{session_id}")
        if token:
            req.add_header("Authorization", f"Bearer {token}")
        with urllib.request.urlopen(req, timeout=15) as resp:  # noqa: S310 (trusted url)
            snap = json.load(resp)
    else:
        raise CorrelationError("need --snapshot or --omnigent-url for the runtime/cost planes")
    if snap.get("id") != session_id:
        raise CorrelationError(
            f"snapshot id={snap.get('id')!r} != resolved session {session_id!r} "
            f"(CP-4 #4: cross-session snapshot)")
    return snap


def correlate(task_ref: str, *, session_id: str | None = None,
              comments: str | None = None, authz_log: str = "",
              snapshot_file: str | None = None, omnigent_url: str | None = None,
              token: str | None = None) -> dict:
    """Resolve + fan out + assert the CP-4 invariants. Returns the contract-C trace."""
    sid, markers = resolve_session_id(task_ref, session_id=session_id, comments=comments)

    # Plane A — authz.
    authz = authz_rows(sid, authz_log)

    # Plane B — runtime: receipt lifecycle markers for THIS session (phase
    # claimed/done) + the snapshot status. Markers are session-keyed (they carry
    # session=<id>), so a decoy session's markers cannot leak in.
    snap = load_snapshot(sid, snapshot_file=snapshot_file,
                         omnigent_url=omnigent_url, token=token)
    runtime = [
        {"event": f"AGENT_{m.get('phase', '').upper()}", "phase": m.get("phase"),
         "session_id": m.get("session")}
        for m in markers if m.get("session") == sid and m.get("phase")
    ]
    runtime.append({"event": "session.status", "status": snap.get("status"),
                    "items": len(snap.get("items", []))})

    # Plane C — cost: Omnigent per-session usage (real join). Dashboard/teo are
    # link-attached only (no session_id column) — design §2 Plane C.
    cost = {
        "session_usage": {
            "last_total_tokens": snap.get("last_total_tokens"),
            "total_cost_usd": snap.get("total_cost_usd"),
            "usage_by_model": snap.get("usage_by_model"),
        },
    }

    # subject — the single NON-EMPTY id across authz (already enforced); attach it
    # + best-effort email. Native-plane rows with an empty subject are tolerated.
    subjects = {r["subject_id"] for r in authz if r["subject_id"]}
    subject_id = next(iter(subjects)) if subjects else ""
    subject_email = next((r["subject_email"] for r in authz if r["subject_email"]), "")

    trace = {
        "task_ref": task_ref,
        "session_id": sid,
        "subject_id": subject_id,
        "subject_email": subject_email,
        "authz": authz,
        "runtime": runtime,
        "cost": cost,
    }
    _assert_cp4(trace)
    return trace


def _assert_cp4(trace: dict) -> None:
    """The CP-4 read-side invariants, in code. Raises CorrelationError on any miss."""
    sid = trace["session_id"]
    # #1: a session_id resolved (truthy) — resolution already guaranteed exactly one.
    if not sid:
        raise CorrelationError("CP-4 #1: no session_id on trace")
    # #2: all three planes non-empty for this session.
    if len(trace["authz"]) < 1:
        raise CorrelationError("CP-4 #2: authz plane empty (need >=1 decision)")
    phases = {r.get("phase") for r in trace["runtime"]}
    if not {"claimed", "done"} <= phases:
        raise CorrelationError(
            f"CP-4 #2: runtime plane missing claimed+done (saw phases {sorted(p for p in phases if p)})")
    if trace["cost"]["session_usage"]["last_total_tokens"] is None:
        raise CorrelationError("CP-4 #2: cost plane has no per-session usage")
    # #3: exactly one NON-EMPTY subject across authz (native-plane empties are
    # "unknown", not a second actor), and it is the one on the trace.
    subjects = {r["subject_id"] for r in trace["authz"] if r["subject_id"]}
    if len(subjects) != 1 or trace["subject_id"] not in subjects:
        raise CorrelationError(f"CP-4 #3: subject_id not single/consistent (non-empty {sorted(subjects)})")
    # #4: every authz + runtime row is for THIS session, none for another.
    if any(r["session_id"] != sid for r in trace["runtime"] if "session_id" in r):
        raise CorrelationError("CP-4 #4: a runtime row references a different session")


# ─────────────────────────────────── CLI / self-check ───────────────────────────────────

_FIX = os.path.join(os.path.dirname(os.path.abspath(__file__)), "testdata", "cp4")
_TASK = "github:Cloud-Byte-Consulting/agentic-harness#42"


def _read(name: str) -> str:
    with open(os.path.join(_FIX, name), encoding="utf-8") as fh:
        return fh.read()


def self_check() -> int:
    """Run the full CP-4 query over the bundled fixtures incl. negative cases."""
    # Positive: one issue -> one correlated trace satisfying all 5 criteria.
    trace = correlate(_TASK, comments=_read("comments.txt"),
                      authz_log=_read("authz.jsonl"),
                      snapshot_file=os.path.join(_FIX, "snapshot.json"))
    assert trace["session_id"] == "conv_abc123", trace["session_id"]            # #1
    assert len(trace["authz"]) == 2, trace["authz"]                              # #2 authz
    assert {"claimed", "done"} <= {r.get("phase") for r in trace["runtime"]}    # #2 runtime
    assert trace["cost"]["session_usage"]["last_total_tokens"] == 41200         # #2 cost
    assert trace["subject_id"] == "e1f2a3b4-c5d6-7890-entra-oid"                # #3
    # native-plane rows carry subject_id="" (unknown pre-OE-3) — tolerated, not a 2nd actor.
    assert all(r["subject_id"] in ("", trace["subject_id"]) for r in trace["authz"])  # #3
    # #4: the decoy session conv_decoy999 (present in authz.jsonl) never leaks.
    assert all(r["ts"] for r in trace["authz"])
    assert "conv_decoy999" not in json.dumps(trace), "decoy session leaked into trace"
    print("PASS positive: trace =", json.dumps(trace, indent=2))

    # Negative A: two authz rows for the same session with different subject_ids
    # must be rejected (CP-4 #3), not silently picking one.
    try:
        correlate(_TASK, comments=_read("comments.txt"),
                  authz_log=_read("authz_mismatch.jsonl"),
                  snapshot_file=os.path.join(_FIX, "snapshot.json"))
        print("FAIL: mismatched subject_id was accepted")
        return 1
    except CorrelationError as e:
        assert "subject" in str(e).lower(), e
        print("PASS negative (subject mismatch rejected):", e)

    # Negative B: two distinct sessions stamped under the same task_ref must be
    # ambiguous (CP-4 #4), not a guess.
    try:
        correlate(_TASK, comments=_read("comments_ambiguous.txt"),
                  authz_log=_read("authz.jsonl"),
                  snapshot_file=os.path.join(_FIX, "snapshot.json"))
        print("FAIL: cross-contaminated task_ref resolved silently")
        return 1
    except CorrelationError as e:
        assert "ambiguous" in str(e).lower(), e
        print("PASS negative (cross-contamination rejected):", e)

    print("\nCP-4 self-check: ALL PASS")
    return 0


def main(argv: list[str] | None = None) -> int:
    ap = argparse.ArgumentParser(description="Open Engine three-plane audit correlation (CP-4).")
    ap.add_argument("--self-check", action="store_true", help="run the CP-4 fixture self-check and exit")
    ap.add_argument("--task-ref", help="<provider>:<id>, e.g. github:owner/repo#42 or linear:ENG-123")
    ap.add_argument("--session-id", help="pin the session_id (skips receipt-marker resolution)")
    ap.add_argument("--comments-file", help="tracker comments (a fixture or `gh issue view <ref> --comments`)")
    ap.add_argument("--authz-log", help="contract-A JSONL decision log (the OE_AUDIT_LOG file)")
    ap.add_argument("--snapshot", help="saved `GET /sessions/{id}` JSON")
    ap.add_argument("--omnigent-url", help="fetch the snapshot live instead of --snapshot")
    ap.add_argument("--token", help="bearer token for --omnigent-url")
    args = ap.parse_args(argv)

    if args.self_check:
        return self_check()
    if not args.task_ref:
        ap.error("--task-ref is required (or use --self-check)")

    comments = open(args.comments_file, encoding="utf-8").read() if args.comments_file else None
    authz_log = open(args.authz_log, encoding="utf-8").read() if args.authz_log else ""
    try:
        trace = correlate(args.task_ref, session_id=args.session_id, comments=comments,
                          authz_log=authz_log, snapshot_file=args.snapshot,
                          omnigent_url=args.omnigent_url, token=args.token)
    except CorrelationError as e:
        print(f"correlation failed: {e}", file=sys.stderr)
        return 2
    json.dump(trace, sys.stdout, indent=2)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
