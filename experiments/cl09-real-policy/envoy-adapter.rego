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

# ext_authz is binary: allow ONLY on "allow". `deny` and `require_approval` both block
# at the gateway; require_approval is then resolved by Omnigent's ASK (per the spike).
default allow := false

allow if verdict == "allow"
