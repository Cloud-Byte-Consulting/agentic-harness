# demo/healthcare — hospital tri-state policy slice

A drop-in demo that reframes the production `mcp_auth` tri-state policy in hospital terms,
so live verdicts read in clinical language (`get_patient_record`) instead of repo tooling
(`get_repo`).

**Same contract as production:** query `data.healthcare.auth.verdict` → one of
`allow` | `deny` | `require_approval`. The orchestrator maps that to forward / 403 / 428-ASK,
exactly like `experiments/cl09-real-policy`.

## Verdict table

| Tool                    | Verdict          | Rationale                                  |
| ----------------------- | ---------------- | ------------------------------------------ |
| `read_care_guidelines`  | allow            | read-prefix, no PHI                        |
| `get_patient_record`    | require_approval | read of protected health information       |
| `send_referral_fax`     | require_approval | PHI disclosure outside the covered entity  |
| `order_medication`      | deny             | consequential clinical action              |

## Run

```bash
# Offline, zero deps (guaranteed demo path — mirrors the Rego exactly):
python3 selfcheck.py

# Or evaluate the real Rego if the OPA binary is installed:
./test.sh
```

## Files
- `healthcare_auth.rego` — the tri-state policy (`package healthcare.auth`).
- `inputs/*.json` — one MCP-call input per tool (production input shape).
- `selfcheck.py` — stdlib mirror of the Rego; prints the verdict table, exits non-zero on mismatch.
- `test.sh` — evaluates via `opa eval` when available, else runs `selfcheck.py`.
- `CONTROL_MAPPING.md` — each verdict → the control objective it satisfies.

## How it slots into the live demo
Mount `healthcare_auth.rego` in the `cl09-real-policy` OPA container (alongside or instead of
`mcp_auth.rego`) and point the Envoy adapter's `_policy_input.server_name` at
`clinical-assistant`. The gateway then returns 200 / 403 / 428 for these clinical tools the
same way it does for the GitHub tools today.

> Demo artifact. The hospital framing is a presentation layer over the real tri-state shape;
> production groups come from a verified JWT, not the empty set used here.
