# Control Mapping — Healthcare Tri-State Policy

Each policy verdict maps to a recognized control objective. The point for a hospital or
covered entity: these are **deterministic** controls enforced *before* a tool runs,
independent of the model — so for a given tool, a mistaken or prompt-injected model cannot
change the verdict. Note the boundary this demo enforces is over the **tool name**, not the
call's arguments: it holds only for a vetted tool catalog whose names honestly reflect
behavior. A tool named `read_notes` that returns PHI would be classed as a low-risk read, so
name hygiene (and, in production, argument-aware policy) is part of the control, not an
afterthought.

| Clinical-assistant tool   | Verdict            | Control objective                            | Maps to (illustrative) |
| ------------------------- | ------------------ | -------------------------------------------- | ---------------------- |
| `read_care_guidelines`    | **allow**          | Least privilege — low-risk read forwarded    | HIPAA §164.312(a); SOC 2 CC6.1 |
| `get_patient_record`      | **require_approval** | Minimum-necessary access to PHI; needs a human | HIPAA Privacy §164.502(b); §164.312(a)(1) |
| `send_referral_fax`       | **require_approval** | PHI disclosure outside the covered entity    | HIPAA §164.508; HITECH breach rules |
| `order_medication`        | **deny**           | Patient safety; no autonomous consequential clinical action | HIPAA integrity §164.312(c)(1); clinical governance |
| *(anything unmatched)*    | **deny** (default) | Secure-by-default — deny unless permitted    | NIST AC-3 |

## Why this satisfies a risk reviewer
- **Deterministic, not probabilistic.** The verdict comes from policy (`data.healthcare.auth.verdict`),
  not the LLM. The "anything an agent can do, a prompt injection can also do" risk is closed at
  the policy layer, not mitigated by hoping the model behaves.
- **Tri-state, not blanket block.** `require_approval` preserves autonomy for safe work while
  routing PHI access, disclosures, and consequential clinical actions to a human (the 428 / ASK flow).
- **Fail-closed.** Unknown tool, unreachable policy, or unmatched input → deny.
- **One policy, many enforcement points.** In production the *same* bundle is enforced at the
  MCP gateway (agentgateway/Sentry) and as a PreToolUse gate inside editors (Claude Code,
  Copilot) via `tools/opa_hook.py`. That hook queries `data.mcp.auth.oe_decision` — an object
  carrying `verdict` — so to be consumed by it directly, this demo policy's `healthcare.auth`
  package and bare-string `verdict` rule would need to match that `mcp.auth.oe_decision` shape.
  This slice illustrates the verdict contract for readability; it is not loadable by
  `opa_hook.py` unmodified.
- **Auditable.** Every verdict is logged and correlated (authz + runtime + cost) into one trace
  per task via `tools/oe_correlate.py`.

> Framework references are illustrative, to anchor the conversation with a clinical risk /
> compliance audience — not a certified control attestation.
