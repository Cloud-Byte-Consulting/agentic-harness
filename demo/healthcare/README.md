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
Every call appends a decision record to a local log — `decisions.jsonl` next to this script by
default (override with `OE_AUDIT_LOG`). The record mirrors OPA's decision-log shape, so it's the
same *shape* of artifact a risk/compliance reviewer reads. The tamper-evident, append-only
guarantee is a property of the production decision-log sink, not of this local file — here the
file is a visibility aid you can read offline, and any process can rewrite it:

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
- `inputs/*.json` — one MCP-call input per tool. These carry the subset the policy actually
  reads (`server_name`, `tool_name`, `groups`); the production Envoy adapter also sends
  `arguments` and `session` (see `experiments/cl09-real-policy/envoy-adapter.rego`).
- `selfcheck.py` — stdlib mirror of the Rego; prints the verdict table, exits non-zero on mismatch.
- `test.sh` — evaluates via `opa eval` when available, else runs `selfcheck.py`.
- `CONTROL_MAPPING.md` — each verdict → the control objective it satisfies.
- `decisions.jsonl` — local audit log written at runtime (gitignored; see the Audit note above).

## How it slots into the live demo
This slice uses `package healthcare.auth`, but the `cl09-real-policy` Envoy adapter imports
`data.mcp.auth` and queries `auth.verdict` (see `experiments/cl09-real-policy/envoy-adapter.rego`),
and its `docker-compose.yml` loads `mcp_auth.rego` by explicit path. So dropping this file into
that container unchanged does nothing — the adapter never queries `healthcare.auth`, and
`_policy_input.server_name` only feeds allow-list lookups inside the policy; it does not select
which package is evaluated. To run these clinical verdicts through that container, do **one** of:

- rename this file's package to `mcp.auth` and mount it in place of `mcp_auth.rego` — replace the
  `/policies/mcp_auth.rego` volume and its `opa run` argument in `docker-compose.yml`; or
- repoint the adapter's `import data.mcp.auth` / `auth.verdict` at `data.healthcare.auth` and add
  this file to the adapter container's `opa run` argument list.

The verdict *shape* is identical to `mcp_auth`, which is what makes either swap mechanical; the
gateway then returns 200 / 403 / 428 for these clinical tools the same way it does for the GitHub
tools today.

> Demo artifact. The hospital framing is a presentation layer over the real tri-state shape;
> production groups come from a verified JWT, not the empty set used here.
