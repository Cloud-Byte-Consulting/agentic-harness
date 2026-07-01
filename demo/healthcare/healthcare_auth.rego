# Healthcare tri-state policy — DEMO variant of the production mcp_auth.rego.
#
# Same contract as the real policy: query `data.healthcare.auth.verdict`, get one of
# "allow" | "deny" | "require_approval", computed deterministically from the tool name
# and (in production) the caller's verified groups. The orchestrator maps the verdict to
# an action: allow -> forward, deny -> 403, require_approval -> 428 (human ASK flow).
#
# This file reframes the production GitHub-tool demo in hospital terms so the live verdicts
# read in domain language. The decision *shape* is identical to mcp_auth.
package healthcare.auth

import rego.v1

# Secure default: nothing runs unless a rule explicitly permits it.
default verdict := "deny"

# Read-only prefixes (low-risk, non-mutating) — mirrors the production read-prefix set.
read_prefixes := ["get_", "list_", "search_", "read_", "view_"]

# Tokens that make a READ touch protected health information (PHI) -> human gate.
sensitive_tokens := ["phi", "patient", "record", "chart", "mrn", "ssn", "diagnosis", "lab", "medication", "prescription", "history"]

# Tokens that disclose PHI OUTSIDE the covered entity (egress) -> human gate.
egress_tokens := ["send", "fax", "share", "export", "message", "transmit", "email", "external", "disclose", "post"]

# Consequential / destructive clinical or data actions -> hard deny.
deny_tokens := ["delete", "discharge", "administer", "prescribe", "order", "cancel", "transfer", "write", "update", "create", "merge"]

_name := lower(input.tool_name)

# Tokens match whole underscore-delimited segments of the tool name, not raw substrings, so
# "order" does not fire on "search_disorder_guidelines" and "lab" does not fire on
# "list_available_appointments". selfcheck.py uses the identical segment match.
_segments := split(_name, "_")

_has(tokens) if {
	some t in tokens
	t in _segments
}

_is_read if {
	some p in read_prefixes
	startswith(_name, p)
}

# Precedence: deny > require_approval > allow > (default) deny.
verdict := "deny" if _has(deny_tokens)

verdict := "require_approval" if {
	not _has(deny_tokens)
	_has(egress_tokens)
}

verdict := "require_approval" if {
	not _has(deny_tokens)
	not _has(egress_tokens)
	_is_read
	_has(sensitive_tokens)
}

verdict := "allow" if {
	not _has(deny_tokens)
	not _has(egress_tokens)
	_is_read
	not _has(sensitive_tokens)
}
