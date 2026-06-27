<role>
You are helping me set up "Open Engine," a Linear-based operating surface for my AI agents. By the end I should have a working queue where each of my agents can find assigned work, claim exactly one task, do the scoped work, leave receipts, pause cleanly when it needs input, resume correctly, and keep a shared status ledger current. Assume I have built none of this yet. Walk me through the whole thing from start to finish, and do not assume I know Linear or MCP.
</role>

<how_to_work_with_me>
- Go in order through the steps below. Do one step at a time and do not skip ahead.
- Ask one focused question at a time. Do not create anything until I confirm the names.
- At each step tell me exactly what to create, what to name it, what to paste, and how to verify it before moving on.
- Prefer the smallest version that works for one operator; we can expand to a team later.
- Use the same names consistently in every issue, prompt, file, and receipt.
</how_to_work_with_me>

<step_0_prerequisites>
1. A Linear workspace (linear.app). Custom workflow statuses can require a paid plan — check mine before we rely on them.
2. One or more AI runtimes that run on my machine (for example Codex, Claude Code, Claude Desktop, or Cursor). Each runtime becomes one "agent."
3. Each runtime must be connected to Linear through Linear's official MCP server so it can read issues, comment, and change status. Help me connect at least one runtime now:
   - Codex: run "codex mcp add linear --url https://mcp.linear.app/mcp", then "codex mcp login linear". If this is my first remote MCP in Codex, first set rmcp_client = true under [features] in ~/.codex/config.toml.
   - Claude Code: run "claude mcp add --transport sse linear-server https://mcp.linear.app/sse", then open Claude Code and run "/mcp" to finish the Linear auth.
4. Verify the connection before continuing: confirm the agent can list a Linear team/project, name the connected account, comment on one throwaway test issue, and change that test issue's status. Do not touch real issues during this check.
</step_0_prerequisites>

<step_1_build_the_queue>
Build the queue inside one team (statuses belong to a team, so pick the team first):
- Create or choose the team that will own agent work.
- Create six workflow statuses, in this order. Agent Done must be in the Completed category; Agent Todo is a Todo/unstarted status; keep Standing, Agent Working, Agent Needs Input, and Agent Review as active (started) states:
  - Standing — durable context that never closes (setup, ledger, routing maps, SOPs).
  - Agent Todo — finite tasks waiting for an agent.
  - Agent Working — the claim lock.
  - Agent Needs Input — paused, waiting for an answer.
  - Agent Review — done but needs human judgment, QA, or approval.
  - Agent Done — done with a receipt, no review needed.
- Create a label named exactly agent-instructions (the runner filters on this exact spelling).
- Create a project: Personal Agent Engine for just me, or Team Agent Engine for a team.
</step_1_build_the_queue>

<step_2_naming>
Help me decide and write down:
- A stable lowercase agent code per runtime, e.g. alex-codex, alex-claude, sam-codex. One runtime maps to exactly one code.
- The title patterns we will use everywhere:
  - Task issue: [agent instructions][<agent-code>][task] <short outcome>
  - Standing setup issue: [agent instructions][all agents][standing_skill] Install Open Engine core context v1
  - Status ledger issue: [agent instructions][all agents][standing_status] Open Engine status ledger
  - Optional standing skill directory: [agent instructions][all agents][optional_standing_skill_directory] Open Engine optional skill directory
- Where each runtime's private local context file lives (for Codex, a good default is ~/.codex/skills/open-agent-engine/SKILL.md).
</step_2_naming>

<step_3_private_context>
For each runtime, create a local private context file holding: engine version, agent code, Linear team/project, the agent-instructions label, allowed local sources, the status ledger issue ID (use a placeholder until we create the ledger in step 4), the optional standing skill directory issue ID, and the safety boundaries below. Then create the private Standing setup issue ([agent instructions][all agents][standing_skill] ...) listing exactly what to install or adapt locally, the local paths, the ledger issue ID, the optional standing skill directory ID, the smoke-test expectations, and the receipt meanings.

Create the optional standing skill directory as a Standing issue. This directory is part of standard setup, but optional skills inside it are not installed during setup. A human must ask the agent to inspect or install an optional skill. First install approval also subscribes that runtime to future same-scope updates for that optional skill. Any expanded capability or new authority needs fresh approval.

Keep org charts, brand voice, customer context, secrets, and account details inside this private issue or the local file, never anywhere public.
</step_3_private_context>

<step_4_status_ledger>
Create the status ledger as a Standing issue with the agent-instructions label. Each agent owns exactly ONE top-level comment that it updates in place every run (never a fresh comment each time). Use this format:

AGENT STATUS
Agent: <agent-code>
Human/operator: <name>
Runtime: <Codex | Claude | other>
Automation: <automation name or manual>
Automation state: <installed | manual-required | blocked | paused>
Last heartbeat: <ISO8601 timestamp>
Last queue result: <checking | none | claimed ISSUE-ID | completed ISSUE-ID | blocked ISSUE-ID | holding ISSUE-ID | resumed ISSUE-ID | failed ISSUE-ID>
Last successful run: <ISO8601 timestamp or unknown>
Local context: <engine version>; <routing map version>
Optional skills: <none or skill-id@version subscribed>
Notes: <none or short blocker>
</step_4_status_ledger>

<step_5_queue_runner>
The runner is the instruction an agent repeats. A "run" (its heartbeat) is one execution of this loop — I trigger it by hand or schedule it with the runtime's scheduler or a cron job. Each run does, in order:
1. Identify this runtime's agent code; open the ledger; set its AGENT STATUS comment's Last queue result to checking.
2. Mandatory standing preflight: compare local context versions against the standing issues addressed to this agent or [all agents]; leave AGENT APPLIED only after actually installing or adapting locally.
3. Optional standing skill preflight: check only optional skills this runtime has already installed or subscribed to. Apply same-scope updates automatically and leave AGENT SKILL UPDATED only after a real local update. Do not browse or install unapproved optional skills during routine runs.
4. Check AGENT HUMAN HOLD issues. If one now shows AGENT HUMAN ANSWERED, move it back to Agent Working, leave AGENT RESUMED, finish it, and stop.
5. Check AGENT BLOCKED issues. If one now has its answer on the same issue, move it back to Agent Working, leave AGENT UNBLOCKED then AGENT RESUMED, finish it, and stop.
6. Check issues this agent delegated to others; leave AGENT FOLLOW-UP if anything changed.
7. Otherwise claim the oldest eligible Agent Todo issue. Eligible = the agent-instructions label, [agent instructions] in the title, AND this runtime's agent code as the second title bracket. Move it to Agent Working, leave AGENT CLAIMED, then re-read the issue.
8. Do only the scoped work. If done with no human judgment needed, leave AGENT DONE and move to Agent Done. If done but it needs review, QA, approval, or publishing, leave AGENT DONE and move to Agent Review.
9. If the missing answer belongs on Linear, ask one specific question, leave AGENT BLOCKED, move to Agent Needs Input, set the ledger to blocked ISSUE-ID, and stop. If it belongs in my own agent thread/app, ask me there, leave AGENT HUMAN HOLD, move to Agent Needs Input, set the ledger to holding ISSUE-ID, and stop.
10. If execution fails unexpectedly, leave AGENT FAILED with the last safe step and retry count.
11. Update the ledger and stop after exactly one task issue.
</step_5_queue_runner>

<receipts>
Use these exact tokens so every runtime and human reads the loop the same way:
- AGENT CLAIMED — posted right after moving to Agent Working; the claim lock.
- AGENT DONE — scoped work finished; pair with Agent Done or Agent Review.
- AGENT BLOCKED — answer belongs on the Linear issue; ask one question; move to Agent Needs Input.
- AGENT UNBLOCKED — a blocked issue's answer arrived; posted just before AGENT RESUMED.
- AGENT HUMAN HOLD — answer belongs in my own agent thread/app; move to Agent Needs Input.
- AGENT HUMAN ANSWERED — I answered a hold in my thread; clears the hold.
- AGENT RESUMED — continuing a paused issue after UNBLOCKED or HUMAN ANSWERED.
- AGENT FAILED — unrecoverable failure only; record the last safe step and retry count.
- AGENT APPLIED — a runtime installed or adapted a standing context version locally.
- AGENT SKILL SUBSCRIBED — a human approved first install/adaptation of an optional standing skill and future same-scope updates for that runtime.
- AGENT SKILL INSTALLED — the runtime actually installed or adapted an optional standing skill.
- AGENT SKILL UPDATED — a subscribed optional standing skill received a same-scope local update.
- AGENT SKILL DECLINED — the human declined or deferred an optional standing skill.
- AGENT FOLLOW-UP — a delegated issue's state changed.
- AGENT STATUS — the single ledger comment each agent updates in place.
</receipts>

<step_6_smoke_tests>
Prove the loop before trusting it. Keep tasks tiny.
1. Basic: create [agent instructions][<agent-code>][task] Say hello from the queue. Expect AGENT CLAIMED, AGENT DONE, Agent Done, and completed ISSUE-ID on the ledger; the runner stops after one task.
2. Blocked-resume: create a task missing one needed fact. First run leaves AGENT BLOCKED and Agent Needs Input. Answer on the same issue. Next run leaves AGENT UNBLOCKED, AGENT RESUMED, then AGENT DONE.
3. Human-hold: ask the agent to request a local runtime permission in my own agent thread. Expect AGENT HUMAN HOLD (not AGENT BLOCKED), holding ISSUE-ID on the ledger, AGENT HUMAN ANSWERED after I reply, then completion.
4. Optional directory: ask the agent what optional Standing Skills are available. Expect a summary of the directory and no install/adaptation until I approve one.
</step_6_smoke_tests>

<safety>
Ask me before publishing, emailing, posting to Slack or anywhere public, deploying, deleting data, changing billing, changing credentials, or making any customer-facing change. External or destructive actions need explicit issue-level approval.
</safety>

<start_here>
Start by helping me choose: the Linear team name, the project name (Personal or Team Agent Engine), my agent codes, the status ledger issue title, the private setup issue title, and where my private local context should live. Ask me one question at a time, then move through steps 0 to 6 above and verify each one before continuing.
</start_here>