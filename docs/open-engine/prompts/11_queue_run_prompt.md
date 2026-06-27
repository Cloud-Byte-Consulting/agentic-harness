<task>
  Run one Open Engine queue check for this operator.
</task>

<order>
  <step>Identify this runtime's agent code.</step>
  <step>Open the status ledger issue.</step>
  <step>Find this agent's top-level AGENT STATUS comment.</step>
  <step>Update that comment in place with Last queue result: checking and the current timestamp.</step>
  <step>Run mandatory standing context preflight before new task work.</step>
  <step>Check subscribed optional standing skills before new task work. Apply same-scope updates automatically and leave AGENT SKILL UPDATED only after a real local update. Do not browse or install unapproved optional skills during this routine run.</step>
  <step>Check AGENT HUMAN HOLD issues before new task work. If a held issue now shows AGENT HUMAN ANSWERED, move it back to Agent Working, leave AGENT RESUMED, finish it, and stop after this one issue.</step>
  <step>Check AGENT BLOCKED issues before new task work. If a blocked issue now has its answer on the same issue, move it back to Agent Working, leave AGENT UNBLOCKED then AGENT RESUMED, finish it, and stop after this one issue.</step>
  <step>Check delegated issues this agent routed to someone else.</step>
  <step>If no hold or blocked issue is ready to resume, find the oldest eligible Agent Todo task assigned to this operator.</step>
  <step>Eligible tasks must have the agent-instructions label, [agent instructions] in the title, and this runtime's agent code as the second title bracket, e.g. [agent instructions][alex-codex][task].</step>
  <step>If no eligible issue exists, update the ledger with Last queue result: none and stop.</step>
  <step>If an issue exists, move it to Agent Working and leave AGENT CLAIMED.</step>
  <step>Re-read the issue after claiming.</step>
  <step>Do only the scoped work.</step>
  <step>If complete with no human judgment needed, leave AGENT DONE and move to Agent Done.</step>
  <step>If complete but review, QA, approval, inspection, or publishing is needed, leave AGENT DONE and move to Agent Review.</step>
  <step>If the missing answer belongs on Linear, ask one specific question, leave AGENT BLOCKED, move to Agent Needs Input, update the ledger with blocked ISSUE-ID, and stop.</step>
  <step>If the answer belongs in the human's own agent thread/app, ask there, leave AGENT HUMAN HOLD, move to Agent Needs Input, update the ledger with holding ISSUE-ID, and stop.</step>
  <step>If execution fails unexpectedly, leave AGENT FAILED with the last safe step and retry count.</step>
  <step>Update the status ledger with completed, blocked, holding, resumed, failed, observed, or claimed for the issue id.</step>
  <step>Stop after exactly one task issue.</step>
</order>

<receipts>
  Use AGENT CLAIMED before work, AGENT DONE when complete, AGENT BLOCKED for Linear-answerable blockers, AGENT HUMAN HOLD for owner-thread permissions, AGENT UNBLOCKED then AGENT RESUMED when a blocked issue's answer arrives, AGENT HUMAN ANSWERED then AGENT RESUMED when a hold is answered, AGENT SKILL SUBSCRIBED / INSTALLED / UPDATED / DECLINED for optional standing skills, and AGENT FAILED only for unrecoverable failure.
</receipts>

<boundaries>
  Never publish, email, Slack-post, deploy, delete, change billing, change credentials, or make outward-facing changes unless the issue explicitly grants that approval.
</boundaries>