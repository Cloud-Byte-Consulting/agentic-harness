# The Agent Maintenance Loop | Unlock AI

Generated 2.39:1 agent maintenance masthead showing an inspection bench for auditing an agent harness on the right side.

# The Agent Maintenance Loop

The Agent Maintenance Loop is a repeatable inspection pass for agents that have moved from experiment into real work. It checks the job, diet, memory, tools, reach, proof, and value of the harness.

Use this guide to audit recent runs, identify the surface that is drifting, and decide whether to keep, change, pause, or retire the agent before the next production cycle.

Use this guide when an agent has become part of real work. Inspect the recent runs and the seven surfaces (job, diet, memory, tools, reach, proof, and value), then decide whether to keep, change, pause, or retire the harness before the next run.

June 16, 2026Last updated6 stepsMaintenance loop1 promptFull-loop audit

**01 / HarnessWhat you are maintaining.You are not maintaining a prompt. You are maintaining the whole harness around delegated work.**

Before you change anything, name the parts of the system you are actually maintaining. This section is the map; the rest of the guide walks it.

### Maintain the harness, not just the prompt.

When an agent misbehaves, the reflex is to edit the prompt. But the prompt is one part of a larger system. The harness is everything that turns a model into a worker: the instructions, the sources and examples it reads, the memory it carries between runs, the tools it can call, the permissions it has, the model and its settings, the human review before its output is used, and any evals that check it.

A drifting agent usually still sounds fluent, so the useful question is not "is this output well-written" but "is this fluent output still doing the current job." You can only answer that by looking at the whole harness. That is what this loop does.

You do

Write down the concrete parts of this agent's harness as they exist now: which instructions, which sources, which tools, who reviews it, and what it is allowed to do.

The AI does

Have your AI inventory the harness from its config and docs, and flag any part you cannot point to: an unnamed source, an unclear review step, a tool nobody remembers granting.

### The seven surfaces you will inspect.

The loop inspects the harness through seven surfaces. Each is a place an agent can quietly drift, and together they cover the whole system. Here is what each one is, using a refund-reply agent as the running example.

Job. The one sentence of work it owns, like "draft refund replies for billing tickets."

Diet. What you feed it each run: the refund policy, past approved replies, and the ticket it is answering.

Memory. What it carries between runs, like the saved fact "this account is on the legacy plan."

Tools. The actions it can take: search the policy, draft a reply, tag the ticket.

Reach. What it can do without a human. Here it can draft only; it cannot send the reply or issue the refund.

Proof. The evidence it shows, like the policy clause it cited, so a human can check the work instead of trusting it.

Value. Whether the drafts actually get sent, or get rewritten from scratch every time.

You will inspect each of these in Step 3. For now, just confirm you can fill in all seven for your own agent.

If you cannot name one of the seven for your agent, that gap is already a finding. Write it down and keep going.

**02 / TriggerWhen to run the loop.Run this loop when something changes, not on a schedule and not only when something breaks.**

You do not audit on a calendar. You run the loop when a trigger fires. Find the trigger that brought you here; it is the thread you will pull through the steps that follow.

### Run the loop when a trigger fires.

Most things that should start maintenance fall into four families. Any single one is reason enough.

Upstream change. Something the agent depends on moved: a new model version, a changed tool or connector, or an updated source of truth like a revised policy.

Scope creep. The agent is being used beyond its original job, or it keeps asking for more access to keep up.

Rising human cost. People keep fixing the same thing, review takes longer than the work it saves, or cost and latency have climbed.

Quiet failure. A near miss that almost shipped, or output that nobody uses anymore.

### Start with the smallest loop that catches the drift.

You do not need to inspect everything every time. Take the single trigger that brought you here and run the loop against that one agent and that one signal. The goal is to find what changed while the failure is still small enough to fix in one pass, not to launch a governance review.

If the trigger turns out to touch several agents or several surfaces, finish this pass first, then repeat the loop for the next one.

You do

Pick one agent and the one trigger that prompted this. Write the trigger in a sentence; it is the lead you will follow through Steps 1 to 3.

The AI does

Have your AI restate that trigger as a specific, checkable question, such as "did the June 1 policy update break the refund agent's citations?", so the inspection stays pointed.

**03 / Step 1Name the current job.The job sentence is the anchor every later step is judged against.**

Write the job the agent does today, not the one it launched with. Every later step is checked against this sentence, so get it right before moving on.

### Write the job in one sentence.

State the agent's current job in one sentence that names five things: the work it produces, the sources it uses, the user it serves, the human review in the path, and the consequence of its output. The template below forces all five.

If you cannot complete the sentence, that is your first finding: the sources are vague, or there is no clear review step, or you cannot name the consequence. A job you cannot state in one sentence is a job the agent cannot reliably hold.

Copy prompt

**Show the full prompt**

```
<prompt>
  <context>
    The job sentence anchors every later maintenance step.
  </context>
  <task>
    Complete the sentence naming all five required parts.
  </task>
  <deliverables>
    <deliverable>This agent's job is to [produce this work] from [these sources] for [these users], with [this human review] before [this consequence].</deliverable>
  </deliverables>
</prompt>
```

### Keep the job narrow enough to maintain.

A narrow job is one you can actually check the agent against. A broad one gives drift room to hide. Compare:

Maintainable: "Draft refund replies for billing tickets under $100, from the refund policy, for a support agent to approve before sending."

Not maintainable: "Handle support."

Maintainable: "Prepare first-pass backlog packets for the product team to refine."

Not maintainable: "Help product."

If your sentence reads like the broad versions, narrow it now. You can always run a second loop for a second job, but each loop needs one clear job to test against.

**04 / Step 2Check the last ten runs.Do not judge the agent in the abstract. Read what it actually did, and find what humans keep fixing.**

Pull the agent's recent real runs (ten is a guide, not a rule) and read them against the job sentence from Step 1. You are gathering evidence here, not fixing anything yet.

### Score each run against the same questions.

Go through the recent runs one at a time and answer the same questions for each. You are looking for where the agent and the human diverged.

For each run: Was the output used, or rewritten or dropped? What did the human change, and why? Which source did the agent rely on? Which tool did it call? What did it say it could not verify? Where did reviewing it take longer than expected?

Capture the answers in a simple list. You do not need a dashboard; a column of notes is enough to see the pattern.

### Treat a repeated correction as a harness problem.

A one-off fix is noise. The same correction across three or more runs is signal: the harness is teaching that mistake, and editing individual outputs will never end it.

When you spot a repeated correction, do not fix it yet. Name the pattern in a few words, like "cites the old refund threshold" or "invents a ticket field," and carry each one into Step 3, where you find which surface is producing it.

**05 / Step 3Inspect the seven surfaces.Take each repeated problem to the surface behind it. Walk the seven in a fixed order so you fix the cause, not the symptom.**

For each surface, ask its question, look for its symptom in the runs you just read, and note the likely fix. Going in order keeps you from piling new instructions onto a problem that lives somewhere else.

### Job, diet, and memory.

Job. Ask whether the work has quietly grown past the sentence from Step 1. It is broken when runs include tasks that sentence never mentioned. Fix by re-narrowing the job, or splitting the new work into its own agent.

Diet. Ask whether everything it reads is still current and correct. It is broken when the agent cites an old policy, leans on a stale example, or retrieves the wrong document. Fix by updating or repointing the sources, not by adding a rule that says "use the latest version."

Memory. Ask whether it is carrying a fact that is no longer true. It is broken when an outdated saved assumption shows up in current work. Fix by clearing or correcting the stored memory.

### Tools, reach, proof, and value.

Tools. Ask whether it can reach the right action without tripping over the wrong one. It is broken when the toolset is so broad or overlapping that the agent picks a wrong or unsafe tool. Fix by removing the tools it does not need for this job.

Reach. Ask whether it can touch more than its owner can review. It is broken when its power to send, spend, change, or publish outruns the human's ability to catch a mistake. Fix by narrowing reach until every risky action passes a person.

Proof. Ask whether a human can check the work or can only trust it. It is broken when output looks finished but shows no sources, no reasoning, and no way to verify. Fix by requiring it to cite or show its work.

Value. Ask whether anyone acts on the output. It is broken when the work is plausible but ignored: rewritten, skipped, or filed unread. Fix by changing the job or retiring the agent. More polish will not help.

You now have a list of surface, problem, and likely fix. Do not apply the fixes yet. Build the replay pack in Step 4 first, so you can prove each one works.

**06 / Step 4Build a replay pack.A small, fixed set of cases with known-right answers: your before-and-after test for any change.**

Assemble the pack before you change anything. You run it now to confirm the problems from Step 3, and again in Step 5 to confirm your fixes helped without breaking something else.

### Choose cases where you already know the right answer.

A replay case is an input where you already know how the agent should behave, so any deviation is obviously wrong. Pick 5 to 20 of them: enough to cover the ways this agent matters, few enough to re-run by hand.

Good cases come from real history: support tickets with a known correct routing, old backlog packets where the product decision is settled, code changes with passing tests and files that must not be touched, research questions with a known source trap, drafts that previously came out in the wrong voice, and at least one high-risk case where the only correct move was to stop and escalate.

Include the problems you found in Step 3. If the agent is citing an old policy, one case should be a ticket that exposes exactly that.

### Score the run, not just the answer.

For each case, do not only check whether the final answer is right. Check how it got there, because that is what predicts the next failure. For every case ask: did it use the right source, choose the right tool, stay inside the job, show its proof, and stop when it should have? And would a human have spent less time reviewing it than doing the work themselves?

Run the pack once now to get a baseline score. That is the number your fixes in Step 5 have to beat.

**07 / Step 5Delete before you add.Most harnesses rot because every fix is one more instruction. Try subtraction first.**

For each problem from Step 3, try removing or narrowing before you add anything new. Re-run the replay pack after each change, so you keep what helps and revert what does not.

### Ask what to remove before what to add.

Before writing a single new instruction, run each problem through the deletion questions. Most agent failures are caused by something that is already there, not something that is missing.

Is a stale source feeding it? Is a bad example teaching it? Is a tool too broad? Is the job too vague? Is an old memory being replayed? Is its reach higher than it needs? Is proof missing? Is the model now good enough that an old workaround is getting in the way?

A "yes" to any of these is a fix by deletion: remove or correct that thing, then re-run the replay pack. Reach for a new instruction only after the deletions are exhausted.

### Add only what the replay pack proves you need.

If a deletion or a narrower scope fixes the behavior in the replay pack, you are done. Do not add a standing instruction on top just because it feels safer; every rule you add is something the next maintainer has to understand and the agent has to weigh on every run.

When you genuinely do need to add something, prove it: the replay pack should fail without it and pass with it. If you cannot show that, you do not yet know the change is doing anything.

**08 / Step 6Decide keep, change, pause, or retire.End with one written decision and the evidence behind it, not a vague sense that the agent feels better.**

Make one call, backed by the replay-pack result and the run evidence. Then write it down, so the next pass starts where this one ended.

### Choose one of four outcomes.

Close the loop with exactly one decision.

Keep. The agent still fits its job and the replay pack passes. Nothing changes but the next review date.

Change. You found and fixed specific surfaces. Record what you changed and the replay-pack result that backs it.

Pause. The agent is useful but currently unsafe or stale, and you cannot fix it in this pass. Stop its risky actions until you can.

Retire. The job changed, the value disappeared, or upkeep costs more than the agent saves. Turn it off and reassign the work.

### Write the record the next pass starts from.

Save a short record so the next maintenance pass, yours or someone else's, does not start from zero. Capture the trigger that started this pass, the current job sentence, the run pattern you found, which surfaces you changed, the replay-pack result, the decision, and the condition that should trigger the next review.

That record is part of the harness. Store it where the agent's config and docs live, not in a chat log you will lose.

**09 / ExamplesApply the loop to real agent types.The same loop, run fast against agents you will recognize: the drift you would see, the surface behind it, and the fix.**

These are worked shorthand, not new rules. Read each as symptom, then surface and why, then fix, and match them to your own agents.

### Writing, content pipelines, and Codex.

A writing agent that sounds like an old version of you. The drift is voice, and the surfaces are diet and proof: its examples are stale, and nothing flags an off-voice draft. Fix: refresh the voice examples it learns from, and add a check that catches off-voice drafts before they ship.

A content pipeline that summarizes the video instead of building the article. A job and value problem: the job has slipped from "write the piece" to "recap the source," and nobody uses the recap. Fix: re-narrow the job sentence to the real deliverable, and stop scoring runs nobody publishes.

A Codex workflow that follows a stale ritual instead of the repo in front of it. A memory and diet problem: it is replaying old standing instructions and reading the wrong context. Fix: delete the outdated instructions, and make it read the current repo before it acts.

### Support, product, and revenue-risk agents.

A support agent citing old policy. A diet and proof problem: it is reading a stale source and showing no citation to catch it. Fix: repoint it at the current policy, and require a citation on every reply.

A backlog agent overweighting one loud customer. A diet problem of source precedence: every input is treated as equal weight. Fix: rank the sources so one noisy account cannot outvote the rest.

A revenue-risk agent that cannot reconcile Stripe, Linear, and local desk state. A reach and proof problem: it can act across systems it cannot reliably verify. Fix: narrow what it is allowed to act on, and make it stop and escalate when the sources disagree.

**10 / PromptRun the maintenance prompt.Run the whole loop as a single paste-in audit when you would rather have your AI walk it with you.**

Use this after you have named the job and skimmed recent runs yourself, so you can check the AI's reading against your own.

### Audit an agentic harness.

Paste this into a coding agent or AI workspace that can see the harness: its instructions, recent runs, source list, tool list, and review notes. The more of the harness it can actually read, the better the audit.

Copy prompt

**Show the full prompt**

```
<prompt>
  <task>
    You are helping me run an Agent Maintenance Loop on an existing agentic harness.

Goal:
Decide whether this agent is still fit for its current job, then recommend keep, change, pause, or retire, with evidence.

Work through these steps in order.

1. Name the current job
- Write one sentence:
  This agent's job is to [produce this work] from [these sources] for [these users], with [this human review] before [this consequence].
- If the job is vague, say so and propose a tighter version.

2. Check the last ten runs
- For each run: Was the output used, or changed or dropped? What did the human change? Which source did it rely on? Which tool did it call? What could it not verify? Where did review take too long?
- List any correction that repeats across three or more runs.

3. Inspect the seven surfaces
For each surface give a verdict (ok / drifting / broken), the evidence from the runs, and the fix:
- Job: has the work grown past the job sentence?
- Diet: are the sources, examples, or retrieved context stale or wrong?
- Memory: is an outdated saved fact being replayed?
- Tools: are tools too broad, overlapping, or risky?
- Reach: can it act beyond what its owner can review?
- Proof: does the output show evidence a human can check?
- Value: is the output actually used?

4. Build or revise the replay pack
- List 5 to 20 known cases, including ones that expose the problems above.
- For each, score source choice, tool choice, job fit, proof, review burden, and stop/escalate behavior.
- Give a baseline result before any change.

5. Delete before adding
- For each problem, name what to remove or narrow first: a stale source, a bad example, a broad tool, excess reach, a wrong memory, a vague job, or an old model workaround.
- Only propose a new instruction if the replay pack would fail without it.

6. Decide
- Return one decision: Keep, Change, Pause, or Retire.
- Include the evidence, the exact harness changes, the replay cases to re-run, and the condition that should trigger the next review.
  </task>
</prompt>
```