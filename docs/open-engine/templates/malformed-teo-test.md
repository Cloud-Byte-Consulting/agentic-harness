# Malformed-TEO test (CP-2)

Input-validation test for the TEO I/O path: a corrupted TEO artifact must be
**rejected** by the validator, not silently accepted or guessed.

**Precondition:** TEO I/O is on (the agent-profile stanza in
`/home/bittahcriminal/air/workspace/agentic-harness/docs/open-engine/templates/starter-private-context-file.md`).
To get a reliably-malformed artifact without inventing TEO syntax, take a VALID
TEO artifact (e.g. the `AGENT DONE` evidence from the CP-2 round-trip test, test
2.2) and **corrupt it** — delete a delimiter, truncate it mid-record, or remove a
block-header line — then paste the corrupted blob in `## Sources` below.

Title to use: `[agent instructions][<your-agent-code>][task] malformed-TEO validation probe`

```markdown
## Requester
<your name>.

## Desired outcome
Validate the TEO artifact in Sources as if it were an upstream agent's handoff,
and report whether it is well-formed. It is intentionally corrupted; you must
DETECT that, not paper over it.

## Sources
A corrupted TEO artifact (paste your corrupted blob between the fences):
~~~
<corrupted TEO here — a valid artifact with a delimiter deleted / truncated>
~~~

## Do
Run the TEO validator/parser (teo.Validate / teo.Parse) on the artifact above.
Do not hand-parse around it or reconstruct what you think it "meant".

## Acceptance criteria
- The agent runs the TEO validator/parser on the pasted artifact.
- Validation FAILS (non-nil error) and the agent reports the failure explicitly.
- The agent leaves AGENT FAILED (or re-prompts for a clean artifact); it does NOT
  silently accept, guess, or fabricate a parse.
- No downstream action is taken on the unparseable content.

## Boundaries
This issue intentionally supplies broken input to verify validation. A clean
rejection is the correct outcome.
```
