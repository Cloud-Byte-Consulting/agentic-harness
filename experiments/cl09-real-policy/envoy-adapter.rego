# Envoy ext_authz adapter — bridges agentgateway's CheckRequest to our REAL policy.
# agentgateway (extAuthz, includeRequestBody) → this adapter → data.mcp.auth.verdict.
package mcp.envoy

import rego.v1
import data.mcp.auth

# The MCP tools/call body agentgateway forwards.
_call := json.unmarshal(input.attributes.request.http.body)

# Build our policy's input from the call. In production server_name comes from the
# route/backend and groups from the verified JWT; here groups are empty (no secrets).
_policy_input := {
	"server_name": "github",
	"tool_name": object.get(_call, ["params", "name"], ""),
	"groups": [],
	"arguments": {},
	"session": {},
}

# Our REAL tri-state decision, evaluated against the constructed input.
# (`with` must live in a rule BODY, not the head.)
verdict := v if {
	v := auth.verdict with input as _policy_input
}

# Structured ext_authz response so the gateway returns the full tri-state, not just
# allow/deny: allow -> forward (200), deny -> 403, require_approval -> 428 (which
# Omnigent resolves via its ASK flow; the x-mcp-verdict header carries the reason).
default response := {"allowed": false, "http_status": 403, "headers": {"x-mcp-verdict": "deny"}}

response := {"allowed": true} if verdict == "allow"

response := {
	"allowed": false,
	"http_status": 428,
	"headers": {"x-mcp-verdict": "require_approval"},
	"body": "Tool requires human approval (Open Engine boundary).",
} if verdict == "require_approval"
