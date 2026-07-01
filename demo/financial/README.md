# demo/financial — financial-services tri-state policy slice

A drop-in demo that reframes the production `mcp_auth` tri-state policy in
financial-institution terms, so live verdicts read in domain language
(`get_member_account`) instead of repo tooling (`get_repo`).

**Same contract as production:** query `data.financial.auth.verdict` → one of
`allow` | `deny` | `require_approval`. The orchestrator maps that to forward / 403 / 428-ASK,
exactly like `experiments/cl09-real-policy`.

## Verdict table

| Tool                     | Verdict          | Rationale                              |
| ------------------------ | ---------------- | -------------------------------------- |
| `read_education_content` | allow            | read-prefix, no confidential data      |
| `get_member_account`     | require_approval | read of confidential member data       |
| `send_external_message`  | require_approval | possible data egress                   |
| `transfer_funds`         | deny             | consequential / destructive boundary   |

## Run

```bash
# Offline, zero deps (guaranteed demo path — mirrors the Rego exactly):
python3 selfcheck.py

# Or evaluate the real Rego if the OPA binary is installed:
./test.sh

# Make a single call and get the gateway verdict (200 / 428 / 403):
python3 selfcheck.py transfer_funds
```

## Audit
Every call appends a decision record to a local log — `decisions.jsonl` next to this script by
default (override with `OE_AUDIT_LOG`). The record mirrors OPA's decision-log shape, so it's the
same *shape* of artifact a risk/compliance reviewer reads. The tamper-evident, append-only
guarantee is a property of the production decision-log sink, not of this local file — here the
file is a visibility aid you can read offline, and any process can rewrite it. In production these
records are OPA decision logs shipped to a sink and correlated with the runtime + cost planes via
`tools/oe_correlate.py`. `tail -1 decisions.jsonl | python3 -m json.tool` pretty-prints one.

## Files
- `financial_auth.rego` — the tri-state policy (`package financial.auth`).
- `inputs/*.json` — one MCP-call input per tool. These carry the subset the policy actually
  reads (`server_name`, `tool_name`, `groups`); the production Envoy adapter also sends
  `arguments` and `session` (see `experiments/cl09-real-policy/envoy-adapter.rego`).
- `selfcheck.py` — thin config over the shared `../common/policy_engine.py` (stdlib mirror of
  the Rego); prints the verdict table, exits non-zero on mismatch.
- `test.sh` — evaluates via `opa eval` when available (asserting against `selfcheck.py --expect`),
  else runs `selfcheck.py`.
- `CONTROL_MAPPING.md` — each verdict → the control objective it satisfies.
- `decisions.jsonl` — local audit log written at runtime (gitignored; see the Audit note above).

## How it slots into the live demo
This slice uses `package financial.auth`, but the `cl09-real-policy` Envoy adapter imports
`data.mcp.auth` and queries `auth.verdict` (see `experiments/cl09-real-policy/envoy-adapter.rego`),
and its `docker-compose.yml` loads `mcp_auth.rego` by explicit path. So dropping this file into
that container unchanged does nothing — the adapter never queries `financial.auth`, and
`_policy_input.server_name` only feeds allow-list lookups inside the policy; it does not select
which package is evaluated. To run these financial verdicts through that container, do **one** of:

- rename this file's package to `mcp.auth` and mount it in place of `mcp_auth.rego` — replace the
  `/policies/mcp_auth.rego` volume and its `opa run` argument in `docker-compose.yml`; or
- repoint the adapter's `import data.mcp.auth` / `auth.verdict` at `data.financial.auth` and add
  this file to the adapter container's `opa run` argument list.

The verdict *shape* is identical to `mcp_auth`, which is what makes either swap mechanical; the
gateway then returns 200 / 403 / 428 for these financial tools the same way it does for the GitHub
tools today.

> Demo artifact. The financial framing is a presentation layer over the real tri-state shape;
> production groups come from a verified JWT, not the empty set used here.
