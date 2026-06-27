# Ineligible-issue test (CP-1)

Input-validation test: an issue the queue runner must **skip**. Proves the runner
does not claim work that fails the eligibility rules (agent-instructions label +
`[agent instructions]` in the title + this runtime's agent code as the second
title bracket).

**Precondition:** make sure no OTHER eligible issue is waiting for this agent on
the same run, so the runner reports `none` rather than claiming a different valid
task. Place this issue in Agent Todo (so it would be a candidate if it were
eligible).

Use ONE deliberately ineligible title (pick any):
- Missing the agent-code bracket: `[agent instructions][task] ineligible probe`
- Wrong agent code (not yours): `[agent instructions][nobody-zzz][task] ineligible probe`
- Omit the `agent-instructions` label entirely.

```markdown
## Desired outcome
This issue is INELIGIBLE on purpose and must be IGNORED by the queue runner. It
exists only to prove the runner does not claim work it should not.

## Acceptance criteria
- The runner does NOT claim this issue (no AGENT CLAIMED appears).
- No Omnigent session is created for this issue (no openengine.issue label).
- With this as the only candidate, the status ledger shows
  `Last queue result: none`.
- The issue stays in Agent Todo, untouched.
```
