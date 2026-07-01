#!/usr/bin/env bash
# Evaluate the healthcare tri-state policy for each clinical-assistant tool.
#
# If the `opa` binary is installed, this evaluates the REAL Rego (data.healthcare.auth.verdict)
# the same way the gateway would. Otherwise it falls back to the stdlib self-check, so the
# demo always has a runnable path.
set -euo pipefail
cd "$(dirname "$0")"

if command -v opa >/dev/null 2>&1; then
  printf "%-24s %-18s %-18s %s\n" "tool" "verdict (opa)" "expected" "result"
  printf -- "----------------------------------------------------------------------\n"
  fail=0
  # Expected verdicts come from selfcheck.py's CASES — the single source of truth — so this
  # path actually ASSERTS the real Rego instead of just printing it. Any eval error or verdict
  # that disagrees with the expectation fails the script (exit 1), catching Rego↔Python drift.
  while read -r tool expected; do
    f="inputs/${tool}.json"
    if ! v=$(opa eval -d healthcare_auth.rego -i "$f" 'data.healthcare.auth.verdict' --format raw); then
      v="ERROR"
    fi
    if [ "$v" = "$expected" ]; then
      mark="ok"
    else
      mark="MISMATCH"
      fail=1
    fi
    printf "%-24s %-18s %-18s %s\n" "$tool" "$v" "$expected" "$mark"
  done < <(python3 selfcheck.py --expect)
  if [ "$fail" -ne 0 ]; then
    echo "FAIL: opa verdicts disagree with expected (Rego and selfcheck.py have drifted)."
    exit 1
  fi
  echo "PASS: all opa verdicts match expected."
else
  echo "opa not found — running the stdlib self-check instead:"
  exec python3 selfcheck.py
fi
