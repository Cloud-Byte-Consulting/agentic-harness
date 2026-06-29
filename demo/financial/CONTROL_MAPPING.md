# Control Mapping — Financial-Services Tri-State Policy

Each policy verdict maps to a recognized control objective. The point for a regulated
financial institution: these are **deterministic** controls enforced *before* a tool runs,
independent of the model — so a mistaken or prompt-injected model cannot cross them.

| Member-assistant tool      | Verdict            | Control objective                         | Maps to (illustrative) |
| -------------------------- | ------------------ | ----------------------------------------- | ---------------------- |
| `read_education_content`   | **allow**          | Least privilege — low-risk read forwarded | SOC 2 CC6.1            |
| `get_member_account`       | **require_approval** | Confidentiality of member data; declassification needs a human | GLBA Safeguards; PCI-DSS 7 |
| `send_external_message`    | **require_approval** | Data-egress / exfiltration prevention     | SOC 2 CC6.7; PCI-DSS 4 |
| `transfer_funds`           | **deny**           | Segregation of duties; no autonomous consequential action | SOX; GLBA |
| *(anything unmatched)*     | **deny** (default) | Secure-by-default — deny unless permitted | NIST AC-3 |

## Why this satisfies a risk reviewer
- **Deterministic, not probabilistic.** The verdict comes from policy (`data.financial.auth.verdict`),
  not the LLM. The "anything an agent can do, a prompt injection can also do" risk is closed at
  the policy layer, not mitigated by hoping the model behaves.
- **Tri-state, not blanket block.** `require_approval` preserves autonomy for safe work while
  routing genuinely consequential or confidential actions to a human (the 428 / ASK flow).
- **Fail-closed.** Unknown tool, unreachable policy, or unmatched input → deny.
- **One policy, many enforcement points.** The same bundle is enforced at the MCP gateway
  (agentgateway/Sentry) and as a PreToolUse gate inside editors (Claude Code, Copilot) via
  `tools/opa_hook.py`.
- **Auditable.** Every verdict is logged and correlated (authz + runtime + cost) into one trace
  per task via `tools/oe_correlate.py`.

> Framework references are illustrative, to anchor the conversation with a risk/compliance
> audience — not a certified control attestation.
