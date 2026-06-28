# CLO-9 finish — our REAL `mcp_auth` policy enforced through agentgateway

**Result (live, 2026-06-27):** our actual Agentic-Sentry `mcp_auth.rego` (the production
tri-state policy) enforced through agentgateway's `extAuthz` → OPA.

| MCP tool | gateway | our policy verdict |
| :-- | :-- | :-- |
| `get_repo` | **200** (forwarded) | `allow` (read-prefix) |
| `delete_repository` | **403** | `deny` (OE boundary: `delete` token) |
| `get_email` | **428** | `require_approval` (read-prefix + `email` token) → Omnigent ASK |

So CLO-9's bet is **finished, not just proven**: agentgateway's `extAuthz` delegates to
our **real Rego bundle** (`data.mcp.auth.verdict`), enforcing the **full tri-state** via a
**structured ext_authz response** — `allow`→forward (200), `deny`→403, and
**`require_approval`→428** (with an `x-mcp-verdict` header), which **Omnigent resolves via
its ASK flow** rather than a blanket block.

## What this adds over [cl09-agentgateway-ext-authz-opa](../cl09-agentgateway-ext-authz-opa/)

That sibling proved the *mechanism* (extAuthz→OPA v3, no shim) with a 1-rule stub. This
proves **our policy**: OPA mounts the **real `mcp_auth.rego`** (from Agentic-Sentry) plus
an Envoy adapter (`package mcp.envoy`) that maps the `CheckRequest` → our input →
`data.mcp.auth.verdict`. **No secrets** — empty group config; the read-prefix `allow`
and the OE-boundary `deny`/`require_approval` need no groups.

## Run

```bash
docker compose up -d
GW=http://localhost:13000 OPA=http://localhost:18181 ./test.sh
docker compose logs opa | grep '"msg":"Decision Log"'   # the verdicts
docker compose down
```

## Files
`docker-compose.yml` (OPA mounts the real `mcp_auth.rego`) · `envoy-adapter.rego`
(`CheckRequest` → our verdict; the `with` must live in a rule body) · `config.json`
(empty `group_rbac`) · `agentgateway.yaml` · `test.sh`.

## True follow-ups (not blockers)
- A real **MCP backend** on an `mcp` route (this uses a plain-HTTP route to isolate authz).
- Forward the **JWT** (`includeRequestHeaders: ["authorization"]`) so **groups** flow → the `is_admin` carve-out.
- `bindSession` anti-spoof + the `OE_AUDIT_LOG` join fields.
