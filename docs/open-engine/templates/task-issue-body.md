# Task issue body

```markdown
<task_issue>
  <title>[agent instructions][<agent-code>][task] <outcome></title>
  <label>agent-instructions</label>
  <status>Agent Todo</status>
  <assignee>The human/operator whose local agent should execute the ticket.</assignee>
  <session_id>populated at claim; omnigent session_id for audit trail join</session_id>
</task_issue>

<body>
  <requester>Who is asking and how to follow up.</requester>
  <desired_outcome>The concrete result wanted.</desired_outcome>
  <context>Why this matters and what background is needed.</context>
  <sources>Links, files, issue IDs, docs, or none.</sources>
  <do>Step-by-step work instructions.</do>
  <acceptance_criteria>Observable success conditions.</acceptance_criteria>
  <output_handoff>Where to put the answer, artifact, pull request, comment, or status update.</output_handoff>
  <boundaries>What the agent may do, what needs approval, and what is out of scope.</boundaries>
</body>
```
