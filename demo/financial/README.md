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
```

## Files
- `financial_auth.rego` — the tri-state policy (`package financial.auth`).
- `inputs/*.json` — one MCP-call input per tool (production input shape).
- `selfcheck.py` — stdlib mirror of the Rego; prints the verdict table, exits non-zero on mismatch.
- `test.sh` — evaluates via `opa eval` when available, else runs `selfcheck.py`.
- `CONTROL_MAPPING.md` — each verdict → the control objective it satisfies.

## How it slots into the live demo
Mount `financial_auth.rego` in the `cl09-real-policy` OPA container (alongside or instead of
`mcp_auth.rego`) and point the Envoy adapter's `_policy_input.server_name` at
`member-assistant`. The gateway then returns 200 / 403 / 428 for these financial tools the
same way it does for the GitHub tools today.

> Demo artifact. The financial framing is a presentation layer over the real tri-state shape;
> production groups come from a verified JWT, not the empty set used here.
