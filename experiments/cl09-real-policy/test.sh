#!/usr/bin/env bash
# Drive 3 MCP tools through agentgateway (HTTP) AND query our policy's verdict directly,
# showing the full tri-state from our REAL mcp_auth.rego.
GW=${GW:-http://localhost:13000}; OPA=${OPA:-http://localhost:18181}
hit() { curl -s -o /dev/null -w "%{http_code}" -X POST "$GW/" -H 'content-type: application/json' \
  -d "{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"$1\"}}"; }
ver() { curl -s "$OPA/v1/data/mcp/auth/verdict" -H 'content-type: application/json' \
  -d "{\"input\":{\"server_name\":\"github\",\"tool_name\":\"$1\",\"groups\":[]}}" \
  | python3 -c 'import sys,json
try: print(json.load(sys.stdin).get("result","?"))
except Exception: print("err")'; }
printf "%-22s %-12s %s\n" tool gateway "our-verdict"
for t in get_repo delete_repository get_email; do
  printf "%-22s %-12s %s\n" "$t" "$(hit "$t")" "$(ver "$t")"
done
echo "expect: get_repo 200/allow · delete_repository 403/deny · get_email 403/require_approval"
