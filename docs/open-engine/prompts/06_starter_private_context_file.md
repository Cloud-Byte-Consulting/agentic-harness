---
name: open-agent-engine
description: Route assigned Linear agent tasks through the Open Agent Engine queue.
version: 1
---

Agent code: alex-codex
Human/operator: Alex Example
Linear team: Agent Engine
Linear project: Personal Agent Engine
Agent label: agent-instructions
Status ledger issue: ENG-000   # placeholder until you create the ledger in Step 4
Optional standing skill directory: ENG-001
Subscribed optional skills: none
Private sources: list approved local folders or private issue ids

Rules:
- Process at most one eligible task issue per run.
- Only process issues assigned to this human/operator.
- Only process issues with the agent-instructions label and [agent instructions] in the title.
- Only claim task issues whose second title bracket is this agent code, e.g. [agent instructions][alex-codex][task]. Standing issues addressed to [all agents] still apply to every runtime.
- Before task work, check mandatory standing context versions assigned to this agent or all agents.
- Before task work, check only subscribed optional standing skills for same-scope updates.
- Do not browse or install optional standing skills during routine runs.
- When the human asks what optional skills are available, read the optional standing skill directory and summarize relevant options.
- First install/adaptation of an optional standing skill requires human approval in this runtime's own agent thread/app.
- First approval subscribes this runtime to future same-scope updates for that optional skill.
- Expanded capability, new authority, new tool access, or a runtime change requires fresh approval.
- Claim by moving the issue to Agent Working and leaving AGENT CLAIMED.
- Re-read the issue after claiming.
- If complete with no human judgment required, leave AGENT DONE and move to Agent Done.
- If complete but review, QA, approval, or inspection is required, leave AGENT DONE and move to Agent Review.
- If the missing answer belongs on Linear, ask one specific question, leave AGENT BLOCKED, and move to Agent Needs Input.
- If the question is about local runtime permissions, skill install approval, automation setup, account authority, or private agent-thread context, ask in the human's own agent thread/app, leave AGENT HUMAN HOLD, and move to Agent Needs Input.
- Ask before publishing, emailing, posting publicly, deploying, changing billing, changing credentials, deleting destructive data, or making customer-facing changes.