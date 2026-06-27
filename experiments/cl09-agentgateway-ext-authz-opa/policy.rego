# OPA-Envoy ext_authz policy — mirrors a governance decision over the MCP tool name.
# The Envoy CheckRequest carries the HTTP body when agentgateway sends includeRequestBody.
# This is the OPA-as-authority point: agentgateway asks, OPA decides (the CLO-9 bet).
package mcp.authz

import rego.v1

default allow := false

# JSON-RPC body may arrive as a decoded string (`body`) or base64 (`rawBody`).
_body := input.attributes.request.http.body

tool := json.unmarshal(_body).params.name

# Governance rule (stand-in for our real Rego): allow `echo`, deny everything else.
allow if tool == "echo"
