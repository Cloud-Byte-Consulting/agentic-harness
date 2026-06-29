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

# Make a single call and get the gateway verdict (200 / 428 / 403):
python3 selfcheck.py order_medication
```

## Audit
Every call appends a decision record to an append-only log — `decisions.jsonl` by default
(override with `OE_AUDIT_LOG`). The record mirrors OPA's decision-log shape, so it's the
exact artifact a risk/compliance reviewer reads:

```json
{"timestamp":"…Z","decision_id":"…","path":"data.healthcare.auth.verdict",
 "input":{"server_name":"clinical-assistant","tool_name":"order_medication","groups":[]},
 "result":"deny","action":"BLOCK","http_status":403,"reason":"consequential clinical action"}
```

In production these records are OPA decision logs shipped to a sink and correlated with the
runtime + cost planes via `tools/oe_correlate.py`; here they land in the local file so the
audit trail is visible offline. `tail -1 decisions.jsonl | python3 -m json.tool` pretty-prints one.

## Files
- `healthcare_auth.rego` — the tri-state policy (`package healthcare.auth`).
- `inputs/*.json` — one MCP-call input per tool (production input shape).
- `selfcheck.py` — stdlib mirror of the Rego; prints the verdict table, exits non-zero on mismatch.
- `test.sh` — evaluates via `opa eval` when available, else runs `selfcheck.py`.
- `CONTROL_MAPPING.md` — each verdict → the control objective it satisfies.
- `decisions.jsonl` — append-only audit log written at runtime (gitignored).

## How it slots into the live demo
Mount `healthcare_auth.rego` in the `cl09-real-policy` OPA container (alongside or instead of
`mcp_auth.rego`) and point the Envoy adapter's `_policy_input.server_name` at
`clinical-assistant`. The gateway then returns 200 / 403 / 428 for these clinical tools the
same way it does for the GitHub tools today.

> Demo artifact. The hospital framing is a presentation layer over the real tri-state shape;
> production groups come from a verified JWT, not the empty set used here.
