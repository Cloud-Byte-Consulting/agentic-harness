# Open Engine | Unlock AI

OE

# OpenEngine

A shared operating surface for your agents.

June 26, 2026Personal or team

01 / Overview

## What you will have when this works.

Open Engine turns Linear into a shared operating surface for agents. A working engine has a queue, private setup context, a status ledger, standing updates, a repeatable runner, resumable blockers, human-thread holds, delegated follow-up, and one smoke-tested task.

### A shared queue

Linear is the v1 queue in this guide. Agents only touch work marked with the right title pattern, label, status, and assignee.

### A private setup issue

Your private skills, optional standing skill directory, brand voice, org chart, customer context, and account details live in your own private issue or runtime context.

### A status ledger

Every agent updates one status comment in place, so humans can see who is installed, automated, blocked, or stale.

### A recurring loop

The queue runner checks standing updates, resumes paused work, follows delegated tasks, then processes one assigned task per run.

### Have your AI help set this up

Copy the full guide context into your AI assistant and have it walk you through the setup one decision at a time.

Copy full setup prompt

02 / Before You Start

## Make the naming decisions first.

Do this before asking an agent to process work. Most failed first runs come from mismatched team names, labels, statuses, issue titles, or agent codes.

1. 01

Pick the Linear team that will own agent work, or create a small team named Agent Engine.

2. 02

Create one project: Personal Agent Engine for one operator, or Team Agent Engine for a team.

3. 03

Create the exact label agent-instructions. The runner filters on this spelling.

4. 04

Choose stable agent codes for every runtime, such as alex-codex, alex-claude, or sam-codex.

5. 05

Name the private setup issue and status ledger before automation exists, so every prompt points at the same source of truth.

### Use AI to make these decisions

Use this before creating anything in Linear. It turns the naming step into a focused planning conversation.

Copy naming prompt

03 / Step 1

## Connect your agent to Linear.

Open Engine only works if your agent can read and update Linear. Connect your runtime to Linear's official MCP server before you build the queue or schedule a runner.

1. 01

Create or open your Linear workspace at linear.app.

2. 02

Choose the agent runtime that will run the queue: Codex, Claude Code, Claude Desktop, Cursor, or another MCP-capable client.

3. 03

Connect that runtime to Linear's official MCP server. The setup command depends on the runtime.

4. 04

Complete the browser authentication flow with the Linear account that should read and update agent issues.

5. 05

Run a read test from the agent, then a write test on a throwaway issue before trusting automation.

 [Open Linear](https://linear.app)  [Linear MCP docs](https://linear.app/docs/mcp)  [Claude Code MCP docs](https://docs.anthropic.com/en/docs/claude-code/mcp) 

### Codex Linear MCP setup

Copy prompt

**Show template**

```
# Recommended Codex local MCP setup
codex mcp add linear --url https://mcp.linear.app/mcp

# If this is your first remote MCP in Codex, enable remote MCP in ~/.codex/config.toml:
[features]
rmcp_client = true

# Manual Codex config alternative:
[mcp_servers.linear]
url = "https://mcp.linear.app/mcp"

# Then authenticate:
codex mcp login linear
```

### Claude Code Linear MCP setup

Copy prompt

**Show template**

```
# Claude Code Linear MCP setup
claude mcp add --transport sse linear-server https://mcp.linear.app/sse

# Then open Claude Code and run:
/mcp

# Follow the Linear authentication flow and enable the Linear tools for the session.
```

### Linear connection verification prompt

Copy prompt

**Show template**

```
Verify my Linear connection before we continue setting up Open Engine.

Do these checks in order:
1. List the Linear workspaces, teams, or projects you can see.
2. Tell me which Linear account appears to be connected.
3. Find or create one throwaway test issue named [agent instructions][connection-test][task] Verify Linear MCP access.
4. Add a comment to that issue saying AGENT CONNECTION TEST from this runtime.
5. Move or update only that test issue if you have status-write access.
6. Report whether you have read access, comment access, and status-update access.

Do not touch any real work issues during this test.
```

### Verify before moving on

* Your agent can list at least one Linear workspace, team, project, or issue.
* Your agent can identify the connected Linear account.
* Your agent can comment on one throwaway test issue.
* Your agent can move or update only that test issue's status.

04 / Step 2

## Build the Linear queue.

 [](https://linear.app) 

Linear is an issue tracker and project-management tool. In Open Engine, it becomes the shared queue: agents read assigned issues, move statuses, leave receipts, and use the issue history as the audit trail.

### Standing

Durable setup, status, skills, routing maps, SOPs, brand guides, and other versioned context.

### Agent Todo

Finite assigned tasks waiting for the target operator's agent.

### Agent Working

The claim lock. An agent has taken the issue and should leave AGENT CLAIMED.

### Agent Needs Input

A paused issue waiting for a Linear answer or human-thread answer.

### Agent Review

Completed work that still needs human judgment, quality assurance, or approval.

### Agent Done

Completed finite work with a receipt and no further human review required.

1. 01

Create or choose the Linear team. Statuses belong to a team, so this decision comes before the workflow.

2. 02

Create the six workflow statuses in order: Standing, Agent Todo, Agent Working, Agent Needs Input, Agent Review, and Agent Done.

3. 03

Make Agent Done a completed-category status. The other agent statuses are active workflow states.

4. 04

Create the agent-instructions label and verify you can apply it to a test issue.

5. 05

Create the Personal Agent Engine or Team Agent Engine project in the same team.

### Verify before moving on

* A test issue can be created in the chosen team.
* The status dropdown includes all six Open Engine statuses.
* Agent Done is a completed status.
* The agent-instructions label can be applied to a test issue.

05 / Step 3

## Create the private context packet.

The public guide teaches the method. Your private packet teaches the actual engine: local paths, allowed sources, account boundaries, receipt rules, and the context each runtime must load before it acts.

1. 01

Create a local private context file for each runtime. For Codex, a good default is ~/.codex/skills/open-agent-engine/SKILL.md.

2. 02

Put the engine version, agent code, Linear team/project, label, allowed sources, status ledger issue, optional standing skill directory issue, subscribed optional skills, and safety boundaries in that file. Leave the status ledger issue ID as a placeholder until you create the ledger in the next step.

3. 03

Create a private Standing setup issue titled [agent instructions][all agents][standing_skill] Install Open Agent Engine core context v1.

4. 04

Create or record the optional standing skill directory. This directory is part of setup, but the optional skills inside it are not installed until a human asks.

5. 05

In that setup issue, list the exact required skills/context to install, local paths, optional directory location, smoke-test expectations, and receipt meanings.

6. 06

Require AGENT APPLIED only after a runtime has actually installed or adapted the target version locally.

### Starter private context file

Copy prompt

**Show template**

```
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
```

### Private setup issue body

Copy prompt

**Show template**

```
<task>
  Create the private standing setup issue for our Open Engine.
</task>

<issue>
  Title: [agent instructions][all agents][standing_skill] Install <Engine Name> core context v<version>
  Label: agent-instructions
  Status: Standing
</issue>

<include>
  <item>What this engine is for.</item>
  <item>Status ledger issue ID or URL.</item>
  <item>Routing map issue ID or URL.</item>
  <item>Optional standing skill directory issue ID or URL.</item>
  <item>Private context packs to install or adapt.</item>
  <item>Optional skills are discoverable but not installed during setup.</item>
  <item>Codex setup steps plus manual fallback.</item>
  <item>Other runtime setup notes where known.</item>
  <item>Required AGENT AUTOMATION READY receipt after install and smoke test.</item>
</include>

<privacy>
  Keep org charts, brand guides, customer context, secrets, account details, and private skill bodies inside this private issue or local runtime context.
</privacy>
```

06 / Optional Skills

## Add Standing Skills without auto-installing them.

Standing Skills are optional shared capabilities your agents can discover from a directory. Standard setup tells agents where the directory lives. It does not install optional skills or grant new authority until the human approves one for that runtime.

1. 01

Create one optional standing skill directory issue and keep its ID in the private setup issue and local runtime context.

2. 02

Create one canonical Standing issue per optional skill. Put the skill purpose, target runtime support, install/adaptation source, version, update channel, approval rules, and receipt templates there.

3. 03

Do not install optional skills during standard setup. Setup only makes agents aware that the directory exists.

4. 04

When a human asks to browse options, read the directory and summarize useful skills for that person's workflow.

5. 05

When a human approves an optional skill, install or adapt it for that runtime, record the subscription locally, and leave the skill receipts on the canonical skill issue.

6. 06

During routine queue runs, check only subscribed optional skills for same-scope updates. Scope expansion still pauses for fresh approval.

### Directory, not auto-install

Standard setup records where optional skills live. It does not install them, enable tools, or grant new authority.

### Approval creates subscription

When the human approves first install or adaptation, that same approval covers future bug fixes and same-scope updates for that skill in that runtime.

### Scope expansion asks again

A skill update that adds new permissions, new external actions, new tools, or a different runtime boundary needs fresh approval.

### First example

Visible Grok + Claude Code delegation is the first optional standing skill: useful for people who want Codex to coordinate those runtimes with receipts.

### Ask what optional skills are available

Use this after setup when you want your agent to inspect the directory without installing anything.

Copy discovery prompt

### Optional skill directory issue body

Copy prompt

**Show template**

```
## Directory
Optional standing skills for this Open Engine.

## Rules
- This directory is part of standard setup.
- Optional skills listed here are not automatically installed.
- A human must ask their agent to inspect or install an optional skill.
- First install/adaptation approval subscribes that runtime to future same-scope updates.
- Scope expansion, new external authority, new tools, or a different runtime boundary requires fresh approval.
- Routine queue runs check only optional skills already installed or subscribed locally.

## Skills

### visible-grok-claude-delegation
- Canonical issue: <ISSUE-ID or URL>
- Purpose: Let Codex visibly coordinate Grok and Claude Code work where the runtime supports it.
- Runtime support: Codex first; other runtimes require adapter judgment.
- Update channel: same-scope-auto-update.
- Approval: required before first install or adaptation.
- Receipts: AGENT SKILL SUBSCRIBED, AGENT SKILL INSTALLED, AGENT SKILL UPDATED, AGENT SKILL DECLINED.
```

07 / Step 4

## Create the status ledger.

The ledger is how you know which agents are online, automated, blocked, holding, stale, or manual-only. It is a standing issue, not a task to close.

1. 01

Create the status ledger as a Standing issue with the agent-instructions label.

2. 02

Paste the AGENT STATUS format into the issue description so every runtime uses the same fields.

3. 03

Add one manual AGENT STATUS comment before automation exists, then make the runner update that same comment in place.

4. 04

Use Last queue result values like checking, none, completed ISSUE-ID, blocked ISSUE-ID, holding ISSUE-ID, resumed ISSUE-ID, and failed ISSUE-ID.

### Create the standing ledger issue

Title it [agent instructions][all agents][standing_status] Open Agent Engine status ledger, set status to Standing, and apply agent-instructions.

### One status comment per agent

Each agent owns exactly one top-level AGENT STATUS comment. On every run, update that comment in place instead of adding heartbeat clutter.

### Record pauses accurately

Use blocked ISSUE-ID for Linear-answerable blockers, holding ISSUE-ID for human-thread holds, and completed ISSUE-ID only after the task is actually done.

### Status ledger comment format

Copy prompt

**Show template**

```
AGENT STATUS
Agent: <agent-code>
Human/operator: <name or unknown>
Runtime: <Codex | Claude | Grok | other>
Automation: <automation name or manual>
Automation state: <installed | manual-required | blocked | paused>
Last heartbeat: <ISO8601 timestamp>
Last queue result: <checking | none | observed ISSUE-123 | claimed ISSUE-123 | completed ISSUE-123 | blocked ISSUE-123 | holding ISSUE-123 | resumed ISSUE-123 | failed ISSUE-123>
Last successful run: <ISO8601 timestamp or unknown>
Local context: <engine version>; <routing map version>
Optional skills: <none or skill-id@version subscribed>
Notes: <none or short blocker>
```

### Verify before moving on

* The ledger issue is in Standing.
* It has the agent-instructions label.
* The first top-level comment starts with exactly AGENT STATUS.
* The runner knows to update that existing comment, not add a new one.

08 / Step 5

## Set up the queue runner.

The runner is one instruction your agent repeats. A run (its heartbeat) is a single execution of this prompt: you trigger it by hand, or schedule it with your runtime's scheduler or a cron job. Each run checks shared updates, resumes holds and blockers, follows delegated work, then processes exactly one eligible task.

1. 01

Identify the runtime's agent code, open the status ledger, and update this agent's existing AGENT STATUS comment to checking.

2. 02

Run mandatory standing preflight: compare target versions for shared skills, SOPs, routing maps, voice guides, and safety rules before new task work.

3. 03

Run optional standing skill preflight only for skills this runtime has already installed or subscribed to. Apply same-scope updates automatically; do not browse or install new optional skills during routine runs.

4. 04

Check AGENT HUMAN HOLD issues first. Resume only after the human answers in their own agent thread or the agent records AGENT HUMAN ANSWERED.

5. 05

Check AGENT BLOCKED issues next. Resume only after the missing answer appears on the same Linear issue.

6. 06

Check delegated issues this agent routed to someone else, and leave AGENT FOLLOW-UP if anything changed.

7. 07

Only then claim the oldest eligible assigned Agent Todo issue, process exactly one task, leave the right receipt, update the ledger, and stop.

### Queue-run prompt

Copy prompt

**Show full runner prompt**

```
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
```

09 / Receipts

## The receipt vocabulary, in one place.

Receipts are the short status comments agents leave on issues and the ledger. Use these exact tokens so every runtime and human reads the loop the same way.

### AGENT CLAIMED

Posted right after moving an issue to Agent Working. The claim lock that stops another runtime from taking the same task.

### AGENT DONE

The scoped work is finished. Pair it with Agent Done when no review is needed, or Agent Review when a human must judge it.

### AGENT BLOCKED

The missing answer belongs on this Linear issue. Ask one specific question and move the issue to Agent Needs Input.

### AGENT UNBLOCKED

Posted when a blocked issue's answer has arrived on the same issue, immediately before AGENT RESUMED.

### AGENT HUMAN HOLD

The answer belongs in the human's own agent thread or app: permissions, installs, account authority. Move to Agent Needs Input.

### AGENT HUMAN ANSWERED

Posted on the issue once the human answers a hold in their own thread, clearing it so the work can resume.

### AGENT RESUMED

Posted when continuing a paused issue, after AGENT UNBLOCKED or AGENT HUMAN ANSWERED.

### AGENT FAILED

Unrecoverable failure only. Record the last safe step and the retry count, then stop.

### AGENT APPLIED

Posted by a runtime after it actually installs or adapts a standing context version locally.

### AGENT SKILL SUBSCRIBED

Posted when a human approves first install or adaptation of an optional standing skill. That approval also covers future same-scope updates for this runtime.

### AGENT SKILL INSTALLED

Posted after the runtime actually installs or adapts the optional standing skill locally.

### AGENT SKILL UPDATED

Posted after a subscribed optional standing skill receives a same-scope local update.

### AGENT SKILL DECLINED

Posted when the human declines or defers an optional standing skill.

### AGENT FOLLOW-UP

Posted on a delegated issue this agent routed to someone else when that issue's state has changed.

### AGENT STATUS

The single top-level ledger comment each agent owns and updates in place on every run.

### Deny vs. blocked (OPA / policy rule)

A hard OPA/policy deny is NOT the same as AGENT BLOCKED. A deny verdict means approval is impossible under current authorization — it produces AGENT FAILED (unrecoverable, never silent retry). AGENT HUMAN HOLD is appropriate only when the fix is a human authority change that, once granted, allows the work to proceed. If the policy denial is permanent or structural, the work cannot proceed and should surface as AGENT FAILED with the denied boundary and the reason. If the human can grant approval and the system honors it (via Omnigent ASK), then AGENT HUMAN HOLD → ASK → approval → RESUMED is correct.

10 / Step 6

## Create the first task and run the smoke test.

The smoke test should be tiny. You are testing the loop, not the agent's intelligence. Do not trust the engine until claim, done, blocked-resume, and human-hold behavior have all worked.

### Basic hello-world task

Create [agent instructions][<your-agent-code>][task] Say hello from the queue. Verify AGENT CLAIMED, AGENT DONE, Agent Done status, and a completed ledger result.

### Blocked-resume task

Create an intentionally incomplete issue. The first run should leave AGENT BLOCKED and Agent Needs Input. Answer on the same issue, then verify AGENT UNBLOCKED, AGENT RESUMED, and AGENT DONE.

### Human-hold task

Ask the agent to request local runtime permission in the current agent thread. Verify AGENT HUMAN HOLD, holding ISSUE-ID on the ledger, AGENT HUMAN ANSWERED after your reply, and then completion.

### Optional skill directory check

Ask what optional Standing Skills are available. Verify the agent summarizes the directory and does not install or adapt anything until you approve a specific skill.

### Basic smoke-test issue body

Copy prompt

**Show template**

```
## Requester
<your name>.

## Desired outcome
Leave a short comment proving the queue runner claimed and completed this issue.

## Sources
None.

## Acceptance criteria
- Issue has AGENT CLAIMED from <your-agent-code>.
- Issue has AGENT DONE from <your-agent-code>.
- Issue moved to Agent Done.
- Status ledger says completed for this issue id.

## Boundaries
Do not edit files, publish, email, Slack-post, deploy, delete, or change credentials.

## If blocked
If the missing answer belongs on Linear, leave AGENT BLOCKED with one specific question.
If the missing answer belongs in the human's own agent thread/app, leave AGENT HUMAN HOLD and wait there.
```

### Blocked-resume test

Copy prompt

**Show template**

```
## Desired outcome
Draft a one-paragraph update, but only after you know the missing date range.

## Missing information
I have intentionally omitted the date range.

## Acceptance criteria
- First run leaves AGENT BLOCKED with one specific question.
- Issue moves to Agent Needs Input.
- Human answers the question on this same issue.
- Next run leaves AGENT UNBLOCKED and AGENT RESUMED.
- Final run leaves AGENT DONE and moves the issue to Agent Review or Agent Done.
```

### Human-hold test

Copy prompt

**Show template**

```
## Desired outcome
Ask me in this agent thread whether you may install or update your private local context for this test.

## Acceptance criteria
- Agent asks in the current agent thread/app.
- Linear issue receives AGENT HUMAN HOLD, not AGENT BLOCKED.
- Status ledger says holding for this issue id.
- After I answer in this thread, the agent records AGENT HUMAN ANSWERED on Linear and continues or closes the issue.
```

### Verify before moving on

* The basic task leaves AGENT CLAIMED, then AGENT DONE.
* The issue moves to Agent Done when no review is required.
* The status ledger says completed ISSUE-ID.
* The optional skill directory check summarizes available skills without installing them.
* The runner stops after exactly one task issue.

11 / Team Path

## Expand from personal engine to team engine.

A team engine is the same system with one extra rule: route work to the human who owns the target agent. Do not assign another person's agent task to yourself and expect their automation to see it.

1. 01

Create a private routing map with each human, Linear assignee, runtime, agent code, and ownership area.

2. 02

Onboard one teammate at a time: install context, create their AGENT STATUS comment, then run a tiny smoke test assigned to them.

3. 03

Assign cross-agent work to the human who owns the target agent, not to yourself.

4. 04

Write every routed task so the target agent can read it cold: requester, outcome, sources, acceptance criteria, output location, boundaries, and pause rule.

5. 05

Use one standing update issue per shared context family, and let agents compare target version to local version during preflight.

### Routing map shape

Copy prompt

**Show template**

```
## Agent routing map

Alex Example
- Linear assignee: Alex Example
- Agent codes: alex-codex, alex-claude
- Route to Alex for: engineering, local repo work, QA

Sam Example
- Linear assignee: Sam Example
- Agent codes: sam-codex
- Route to Sam for: editorial review, social drafts

Rules:
- If assigning work to Sam's agent, assign the issue to Sam.
- If the target agent is not online in the status ledger, say that before relying on the handoff.
- If an agent is not listed yet, it may propose a unique agent code, ask its human to fill in the routing details, and leave those details as a comment on this routing map issue.
- Human approval is required for publishing and customer-facing changes.
```

12 / Templates

## Copy the pieces, then customize them.

These are starter contracts. Replace names, issue IDs, setup paths, context families, version numbers, and boundaries before using them with real agents.

### Private setup issue

Copy prompt

**Show template**

```
<task>
  Create the private standing setup issue for our Open Engine.
</task>

<issue>
  Title: [agent instructions][all agents][standing_skill] Install <Engine Name> core context v<version>
  Label: agent-instructions
  Status: Standing
</issue>

<include>
  <item>What this engine is for.</item>
  <item>Status ledger issue ID or URL.</item>
  <item>Routing map issue ID or URL.</item>
  <item>Optional standing skill directory issue ID or URL.</item>
  <item>Private context packs to install or adapt.</item>
  <item>Optional skills are discoverable but not installed during setup.</item>
  <item>Codex setup steps plus manual fallback.</item>
  <item>Other runtime setup notes where known.</item>
  <item>Required AGENT AUTOMATION READY receipt after install and smoke test.</item>
</include>

<privacy>
  Keep org charts, brand guides, customer context, secrets, account details, and private skill bodies inside this private issue or local runtime context.
</privacy>
```

### Queue runner prompt

Copy prompt

**Show template**

```
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
```

### Task issue body

Copy prompt

**Show template**

```
<task_issue>
  <title>[agent instructions][<agent-code>][task] <outcome></title>
  <label>agent-instructions</label>
  <status>Agent Todo</status>
  <assignee>The human/operator whose local agent should execute the ticket.</assignee>
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

### Status ledger comment

Copy prompt

**Show template**

```
AGENT STATUS
Agent: <agent-code>
Human/operator: <name or unknown>
Runtime: <Codex | Claude | Grok | other>
Automation: <automation name or manual>
Automation state: <installed | manual-required | blocked | paused>
Last heartbeat: <ISO8601 timestamp>
Last queue result: <checking | none | observed ISSUE-123 | claimed ISSUE-123 | completed ISSUE-123 | blocked ISSUE-123 | holding ISSUE-123 | resumed ISSUE-123 | failed ISSUE-123>
Last successful run: <ISO8601 timestamp or unknown>
Local context: <engine version>; <routing map version>
Optional skills: <none or skill-id@version subscribed>
Notes: <none or short blocker>
```

13 / Troubleshooting

## If the first run fails, check these in order.

Most Open Engine failures are routing failures, stale runner instructions, unclear human-input semantics, or missing status transitions. Fix the contract, then rerun the smoke test.

### Agent says no issue exists.

Check assignee, agent-instructions label, Agent Todo status, the [agent instructions] title marker, and that the second title bracket matches this runtime's agent code.

### Agent claims but another agent also works it.

Move status to Agent Working before task work, leave AGENT CLAIMED, then re-read the issue. That status move is the visible lock. If two of your own runtimes share one operator, also scope pickup by the agent-code bracket so only the intended runtime claims each task.

### Ledger gets many heartbeat comments.

Find the top-level AGENT STATUS comment for the current agent code and update it in place by comment id.

### Blocked issues never get picked up again.

Treat AGENT BLOCKED as a pause. Look for the same-issue answer, then leave AGENT UNBLOCKED and AGENT RESUMED.

### Agent asks runtime-permission questions in Linear.

Use AGENT HUMAN HOLD. Ask in the human's own agent thread or app, keep the Linear issue in Agent Needs Input, and set the ledger to holding ISSUE-ID.

### Every standing update creates a pile of duplicate tickets.

Use one standing issue per context family. Update the version and changelog in place; agents compare versions during preflight.

### Agent installs an optional skill during setup.

Separate the optional standing skill directory from installation. Standard setup records the directory only. First install or adaptation requires an explicit human approval in that runtime's agent thread or app.

### Installed optional skills do not receive fixes.

Check the local subscription marker, canonical optional skill issue, and AGENT SKILL SUBSCRIBED receipt. Subscribed skills should auto-update for same-scope changes during preflight.

### Work assigned to another agent goes nowhere.

Assign the issue to the human who owns the target agent, then check the routing map, target heartbeat, label, title marker, and status.

### Agent tries to publish, email, post, deploy, bill, credential-change, or delete.

Add explicit ask-first boundaries to the private context and task body. External or destructive actions require issue-level approval.