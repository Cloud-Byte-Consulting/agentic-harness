#!/usr/bin/env bash
# Run the opa-hook backend (OPA loaded with the real mcp_auth.rego, whose
# oe_decision rule tools/opa_hook.py queries) as a single Docker container.
# Bind-mounts the live Agentic-Sentry policy tree read-only — no build step,
# no copy, so it can never drift from the source of truth. Restart the
# container to pick up a rego edit.
set -euo pipefail
cd "$(dirname "$0")"

POLICY_DIR="${AGENTIC_SENTRY_POLICIES:-}"
if [[ -z "$POLICY_DIR" ]]; then
  if [[ -d ../../../../Agentic-Sentry/mcp-policies ]]; then
    POLICY_DIR="$(cd ../../../../Agentic-Sentry/mcp-policies && pwd)"
  fi
fi
if [[ -z "$POLICY_DIR" || ! -d "$POLICY_DIR" ]]; then
  echo "Agentic-Sentry/mcp-policies not found as a sibling of agentic-harness." >&2
  echo "Clone it there, or set AGENTIC_SENTRY_POLICIES=/path/to/Agentic-Sentry/mcp-policies." >&2
  exit 1
fi

docker rm -f opa-hook-backend >/dev/null 2>&1 || true
docker run -d --name opa-hook-backend \
  -p 8181:8181 \
  -v "$POLICY_DIR:/policies:ro" \
  openpolicyagent/opa:latest \
  run --server --addr=0.0.0.0:8181 --set=decision_logs.console=true /policies

echo "Started. Verify oe_decision is queryable:"
echo "  curl -s http://127.0.0.1:8181/v1/data/mcp/auth/oe_decision \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"input\":{\"server_name\":\"native\",\"tool_name\":\"Bash\",\"arguments\":{},\"groups\":[]}}'"
echo
echo "Stop with: docker rm -f opa-hook-backend"
