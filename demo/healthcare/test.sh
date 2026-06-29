#!/usr/bin/env bash
# Evaluate the healthcare tri-state policy for each clinical-assistant tool.
#
# If the `opa` binary is installed, this evaluates the REAL Rego (data.healthcare.auth.verdict)
# the same way the gateway would. Otherwise it falls back to the stdlib self-check, so the
# demo always has a runnable path.
set -euo pipefail
cd "$(dirname "$0")"

if command -v opa >/dev/null 2>&1; then
  printf "%-24s %s\n" "tool" "verdict (via opa eval)"
  printf -- "------------------------------------------------\n"
  for f in inputs/*.json; do
    tool=$(basename "$f" .json)
    v=$(opa eval -d healthcare_auth.rego -i "$f" 'data.healthcare.auth.verdict' --format raw 2>/dev/null || echo "?")
    printf "%-24s %s\n" "$tool" "$v"
  done
  echo "expect: read_care_guidelines/allow · get_patient_record/require_approval · send_referral_fax/require_approval · order_medication/deny"
else
  echo "opa not found — running the stdlib self-check instead:"
  exec python3 selfcheck.py
fi
