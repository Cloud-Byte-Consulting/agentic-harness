# Status ledger comment format

```markdown
AGENT STATUS
Agent: <agent-code>
Human/operator: <name or unknown>
Runtime: <Codex | Claude | Grok | other>
Automation: <automation name or manual>
Automation state: <installed | manual-required | blocked | paused>
Last heartbeat: <ISO8601 timestamp>
Last queue result: <checking | none | observed ISSUE-123 | claimed ISSUE-123 | completed ISSUE-123 | blocked ISSUE-123 | holding ISSUE-123 | resumed ISSUE-123 | failed ISSUE-123>
Last session ID: <omnigent session_id or none>
Last successful run: <ISO8601 timestamp or unknown>
Local context: <engine version>; <routing map version>
Optional skills: <none or skill-id@version subscribed>
Notes: <none or short blocker>
```

`Last session ID` echoes the `session` value from the receipt marker in `AGENT CLAIMED`. Receipt marker format (in `AGENT CLAIMED` and `AGENT DONE` issue comments): see [CONTRACT.md §Receipt marker](../CONTRACT.md).

```
<!-- openengine session=<session_id> task_ref=<provider>:<id> phase=claimed|done -->
```
