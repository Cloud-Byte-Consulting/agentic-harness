#!/usr/bin/env bash
# Send an allowed (echo) and a denied (add) MCP tools/call body through agentgateway.
# Expect: echo -> 200 (OPA allow -> forwarded), add -> 403 (OPA deny, pre-backend).
set -u
GW=${GW:-http://localhost:3000}
hit() { curl -s -o /dev/null -w "%{http_code}" -X POST "$GW/" \
  -H 'Content-Type: application/json' \
  -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"$1\"}}"; }
echo "echo (expect 200): $(hit echo)"
echo "add  (expect 403): $(hit add)"
