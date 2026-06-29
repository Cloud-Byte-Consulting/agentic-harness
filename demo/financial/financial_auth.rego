# Financial-services tri-state policy — DEMO variant of the production mcp_auth.rego.
#
# Same contract as the real policy: query `data.financial.auth.verdict`, get one of
# "allow" | "deny" | "require_approval", computed deterministically from the tool name
# and (in production) the caller's verified groups. The orchestrator maps the verdict to
# an action: allow -> forward, deny -> 403, require_approval -> 428 (human ASK flow).
#
# This file reframes the production GitHub-tool demo in financial-institution terms so the
# live verdicts read in domain language. The decision *shape* is identical to mcp_auth.
package financial.auth

import rego.v1

# Secure default: nothing runs unless a rule explicitly permits it.
default verdict := "deny"

# Read-only prefixes (low-risk, non-mutating) — mirrors the production read-prefix set.
read_prefixes := ["get_", "list_", "search_", "read_", "view_"]

# Tokens that make a READ touch confidential member data -> human gate (declassification).
sensitive_tokens := ["pii", "ssn", "account", "balance", "member_data", "statement", "card"]

# Tokens that move data OUT of the institution (egress) -> human gate.
egress_tokens := ["send", "email", "share", "export", "message", "post", "external"]

# Consequential / destructive tokens -> hard deny (segregation-of-duties boundary).
deny_tokens := ["delete", "transfer", "move_funds", "wire", "pay", "close", "write", "update", "create", "merge"]

_name := lower(input.tool_name)

_has(tokens) if {
	some t in tokens
	contains(_name, t)
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
