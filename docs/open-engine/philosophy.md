# Open Engine: The Handoff, Not the Model

You finish a client call and the work starts its commute. The transcript goes to Claude to find the argument. Codex changes the file. ChatGPT reads the draft again. A browser agent checks the page actually rendered. Slack has the conversation, Linear has the task, and the calendar decides whether any of it survives the afternoon.

That’s not seven jobs. It’s one job crossing seven systems. And every time it crosses, you carry the state: what was decided, what source mattered, what changed, what the next tool is allowed to touch.

I don’t want one of those tools to swallow the others. Most serious AI users I know don’t either. They have preferences. They know Claude is better at one thing, Codex at another, a local agent at a third, and they don’t want to crown a favorite and pretend the rest of the world disappeared. The trouble isn’t that we own too many good tools. It’s the boring middle between them, the place where one tool’s result becomes the next tool’s task with its source and limits still attached, and for most of us that middle is still a person. Right now, the integration layer is you. It doesn’t have to stay that way.

If you can code, you can build your way around this. You can wire the tools together with APIs, a custom harness, a few cron jobs, and for engineers that’s getting easier every month. But that’s not an answer for everyone else, and it’s barely an answer for a team. The need underneath isn’t exotic. You want one AI’s result to become another AI’s task, with the sources attached, the limits visible, and enough of a trail that the next person or agent doesn’t have to read a giant chat to catch up.

I know a product lead with a newborn and an agency. She runs Claude Code, she has loops and automations, she’s looked hard at OpenClaw. She’s not new to any of this. Her problem is smaller and more maddening than that. A client call collides with a baby appointment. The product scoping still has to move, the team still needs to know what changed, and she’s the one copy-pasting the state of her life between five tools while holding a baby. That’s the load Open Engine takes off her. Not the judgment. Not the taste. The handoffs around them.

I’ve been running a working version to get content out, organize my life, move houses, and coordinate with my team. I’m releasing it now because the next real AI problem isn’t “which model is smartest.” It’s whether the work can move between models at all.

Here’s what’s inside:
* **Open Engine itself.** The build I actually run, packaged as copy-paste templates you hand to the AI you already use, so you have a loop running today.
* **The handoff, not the model.** Why the thing that breaks is never “can the agent do it” but “can the work survive the trip to the next tool.”
* **The smallest useful version.** A shared task list and a seven-part task record that carry the job across tools, so a good answer stops dying in a private chat.
* **The one-loop audit, and the 30-minute build.** Nine questions that turn one annoying handoff into a task an agent can claim, pause, resume, and finish with evidence.
* **The receipt.** The short vocabulary that keeps an agent accountable after the run ends, so “done” stops meaning “now go audit it yourself.”
* **Teams, and the rest of the field.** How one person’s agent hands another person’s agent real work, and where this sits next to OpenClaw, Hermes, and Symphony.

Start where everyone feels it first: the moment the work has to leave one tool and land in the next.

---

## The break is between tools

A lot of agent demos are built around action. The agent opens the browser, edits the file, runs the test, checks the inbox, writes the email, schedules the thing, and goes off while you’re not watching.

I understand why that is exciting. We have spent years with AI that could only explain work back to us. Of course people want software that acts.

But once agents can act, the question changes. It’s bigger than, “Can this agent do the task?” It becomes, “Can the task survive the trip to the next place?”

Can the work leave Claude and arrive in Codex? Can a teammate’s agent pick up a task created by my agent? Can a support loop escalate to the person with authority without losing the original message, customer history, and reason the agent stopped? Can a household scheduling loop draft the next move without pretending it gets to decide for the whole family?

The model still matters, but this is mainly a handoff problem, the one that shows up after the model has already done something useful.

Earlier this week I wrote that the most useful agents are really loop managers: a workflow you can run again, one that notices what changed and pulls you in when a decision actually needs you. That’s true for one agent. But when every loop lives in its own room, the human becomes the hallway between them.

The handoff is the moment when one useful output has to become the next useful action. Who gets this next? What do they need? What are they allowed to do? What source are they supposed to trust? What should they show when they’re done? If they’re blocked, where does that blocker live?

This is where babysitting happens. The human is doing more than checking quality. The human is reconstructing state. You are trying to remember what was asked, where the source came from, whether the agent stayed inside the boundary, and whether the next person or tool has enough context to continue.

A beautiful brief in a private chat is output. It becomes work only when someone can review it, accept it, route it, or build on it. If nobody knows where it came from, what standard it used, who reviewed it, or what the next step is, it’s still just a draft in a room by itself.

Open Engine is about getting from output to work without making human beings the copy-paste path.

---

## What Open Engine actually is

Open Engine isn’t another chatbot, a new model, or a claim that one agent can run your company, your household, or your life.

I could have built the impressive version first.

Honestly, the impressive version is tempting. I’ve been living inside these tools long enough to want all of it: the dashboard, the background runner, the agent watching the other agents, the whole command center. I understand why engineering teams go there. If you’re running ten coding agents from an issue tracker, some of that machinery is real.

But every time I made the idea more impressive, it got farther from the thing I actually needed on a hard day.

The questions I kept coming back to were smaller and more annoying. Can this work leave one AI and arrive at the next without me becoming the messenger? Can the next agent know what happened without me writing the recap? Can it stop before it makes a decision I should make? Can I look at the result and know what changed?

That’s the version of the problem I care about most, because it’s the version I keep living. I don’t need a more theatrical agent system. I need fewer useful outputs dying in private chats.

The simplest version is almost embarrassingly plain: a shared task list.

That’s it. The work needs somewhere to live after it leaves a chat window. I use Linear in my current build because it already has tasks, owners, comments, links, history, and a way to mark what stage the work is in. But the point isn’t Linear. The work is no longer trapped in the chat where one AI produced it.

Then each task needs to answer the questions a person would ask before picking up a job.

What do you want done? What should I read? Who gets to make the final call? What am I allowed to do on my own? What should I never do without asking? Where should I put the result?

That’s the difference between prompting an AI and giving an agent a job.

In plain English, Open Engine needs a place for the work to live, clear instructions for the agent, a way for the agent to say “I picked this up,” a way for it to stop and ask when it needs a real decision, and a note at the end that says what it did, where the result is, what it checked, and what still needs a human.

That’s the whole shape. A request becomes a task, the task says what should happen, and the right agent picks it up. The work happens in whatever tool that agent is best at. If the agent gets stuck, it asks in the same place instead of guessing. If it finishes, it leaves enough evidence that the next person or agent can continue.

Once you have that, the agents don’t have to collapse into one perfect assistant. Claude can think through the argument. Codex can edit the files. OpenClaw can run a local loop. A browser agent can inspect the page. A teammate’s agent can review the result. They don’t need to belong to the same company or subscription. They just need a shared place to pick up the next step.

That’s why Open Engine can be small. You’re not waiting for Claude, Codex, ChatGPT, OpenClaw, Slack, and your calendar to integrate with one another. You’re giving them one simple path for work to leave one tool and arrive at the next with the useful context still attached.

And this is why the guide is meant to be something you can hand to the AI you already use, not a white paper you admire and never implement. If you already have an agent that can read and update a task list, the first test is tiny: create one simple task, assign it, let the agent say it has started, make it leave a note when it’s done, and stop. The test isn’t meant to be impressive. Once a tiny task can move through the loop cleanly, you’ve proved the pattern.

Then you can choose one real handoff from your life.

---

## Start with one loop, not your whole life

The wrong way to approach this is to ask, “How do I integrate every AI system I use?”

That’s too large. It turns the setup into an architecture fantasy, and architecture fantasies are how useful tools become weekend projects you never trust.

Start smaller. Pick one handoff that already annoys you. It should be frequent enough that the pain is real, but small enough that failure isn’t catastrophic. Don’t start with payroll, legal filings, customer refunds, production deploys, or anything where a bad first run creates a mess. Start with the handoff you already do manually.

Good first loops look like this:
* A transcript becomes owners, next actions, and a draft update.
* A script draft gets reviewed by another agent and comes back with specific notes.
* A metrics pull gets routed to the person whose agent has the right access.
* A support thread gets classified, summarized, and prepared for escalation.
* A moving plan drafts messages to the utility companies, landlord, and movers, then pauses for scheduling.

The test is simple: where did AI help you, but still leave you carrying the handoff?

Look for the moment where you usually say, “Okay, now I need to tell the next thing what just happened.”

That is the loop.

---

## The one-loop audit

Before you build anything, write down the loop in plain English. Not a spec. Not an automation diagram. Just the annoying sequence you’re tired of carrying.

Use these questions:
1. What starts the loop?
2. What source material does the agent need?
3. What is the exact outcome?
4. What should the agent never do without asking?
5. What counts as finished?
6. What counts as reviewable but not finished?
7. What question should the agent ask if it gets stuck?
8. Where should the answer or artifact be left?
9. Who or what might need to continue from there?

If you can’t answer those questions, the agent can’t be expected to behave well. It may still produce something impressive, but you’ll pay for it afterward in review time.

This is the first useful shift: stop describing the task as a prompt and start describing it as a handoff.

A prompt asks for output. A handoff explains the job. That difference matters because most agent failures aren’t failures of language. They are failures of ownership, source, boundary, status, and proof.

---

## The task record

Open Engine is built around a boring object: a task record.

In the current version, that record is a Linear issue. It could be something else. Linear is the recommended v1 because agents can read it, update it, leave comments, and move status in a way people can inspect. The important part isn’t the brand of task manager. The important part is that the record carries the job across tools.

A good agent task has seven parts:
1. **The requester:** Who is asking, and who owns the decision?
2. **The desired outcome:** What should exist when this is done?
3. **The sources:** What should the agent read, and what should it not invent?
4. **The acceptance criteria:** What makes the result good enough to stop?
5. **The boundaries:** What must the agent ask before doing?
6. **The blocker rule:** If the agent can’t continue, what exact kind of question should it ask, and where should it ask it?
7. **The receipt:** When the agent claims, blocks, resumes, or finishes, what should it leave behind so the next person can trust the state?

That’s the primitive: not a smarter chatbot, not an agent personality, and not a magic planning layer. A task record with enough context and enough proof that the human doesn’t have to hold the whole run in memory.

### The Plain Template

```text
Title:
agent-instructions | agent-code | plain task name
(agent-code is the short name of the agent you want, e.g. codex or claude)

Requester:
person who owns the outcome

Desired outcome:
what should exist when the task is finished

Sources:
- source 1
- source 2
- private source rules if needed

Acceptance criteria:
- specific check
- specific check
- where the output should be left

Boundaries:
- Do not publish, email, post, deploy, delete, change billing, change credentials, or make customer-facing changes without human approval.
- Do not act outside the named source, repo, folder, project, or person without asking.
- If the task requires judgment that is not stated here, ask before acting.

If blocked:
Ask one specific question on this same task.
Leave AGENT BLOCKED, or AGENT HUMAN HOLD if the answer needs the owner's approval.
Move the task to Agent Needs Input.
Do not create a second task just to ask the question.

Receipts:
- AGENT CLAIMED when work starts.
- AGENT BLOCKED when a specific answer belongs on this task, like a date range or source.
- AGENT HUMAN HOLD when the answer needs your approval or your own agent, like a permission or an install.
- AGENT RESUMED when the answer arrives and work continues.
- AGENT DONE when the scoped work is finished.
- AGENT FAILED if the run fails unexpectedly, with the last safe step.
```

You can copy that into any system that an agent can read and update. Open Engine gives you the full Linear-based version, but the value is visible even before you install anything: the handoff gets clearer the moment the source, boundary, blocker, and receipt are written down.

---

## What this looks like in real work

Take a script review.

The bad version is a Slack message: “Can someone have an agent look at this and tell me if it’s stronger?”

That sounds normal because humans are good at filling gaps. An agent is not. The agent needs to know which script, which version, what changed, what kind of review is wanted, what to ignore, what “stronger” means, where to leave notes, and whether it’s allowed to rewrite or only critique.

The Open Engine version is a task:

```text
Desired outcome:
Review the latest Open Engine script for narrative strength. Do not rewrite the script. Leave a concise review with the three strongest narrative improvements and any line-level problem that would hurt retention.

Sources:
- CMP asset ID or link
- Prior draft link
- Current positioning note

Acceptance criteria:
- Review distinguishes narrative structure from technical accuracy.
- Review says whether the opening earns attention.
- Review identifies any place the script sounds like an explainer instead of a story.
- Notes are left on this issue.

Boundaries:
Do not overwrite the script.
Do not create a new CMP asset.
Do not publish or notify anyone outside this task.
```

Call it bureaucracy if you want. It’s respect for the task. Now the agent has a real job. If it needs the wrong version, it can ask. If it finishes, the next person can see exactly what was reviewed. If another agent needs to revise the script later, it can read the same task and know what happened.

The same pattern works for a team metrics pull.

The bad version is: “Ask Leo’s agent to pull the latest activation metrics.”

The better version says which metric, which source, which date range, what format, where to leave the result, and what to do if the date range is missing. If Leo’s agent is offline, the task should say that. If it’s assigned to the wrong human, Leo’s automation will never see it. That’s the kind of tiny operational detail that breaks agent handoffs in the real world.

---

## Why the receipt matters more than it sounds

The receipt is the part people underestimate.

You don’t abandon agent systems only because they make mistakes. You abandon them because the mistakes become mysterious. Something changed, something got filed, something got marked done, and something got skipped. You don’t know why. And you don’t want to spend your night reading a transcript to find out.

Receipts make the system inspectable.

The `AGENT CLAIMED` receipt tells you the task is owned, which stops two agents from doing the same work.

`AGENT BLOCKED` is the receipt that says the agent stopped because it needs a specific answer that belongs on the task, so a vague failure doesn’t turn into a forgotten task.

`AGENT HUMAN HOLD` is the other kind of pause, for when the answer needs the owner instead of the task, like a permission or an install only a person can grant.

When you see `AGENT RESUMED`, the answer arrived and the agent picked the same record back up. That receipt is the only thing standing between a paused task and a dead one.

`AGENT DONE` is the receipt that carries weight: what changed, where the output is, and what still needs a human.

And `AGENT FAILED` is the receipt you actually want when a run breaks, because it leaves the last safe step, which is the whole difference between a retry and a mess.

They look like cute status labels, but in reality, they’re the minimum language of trust.

When a person tells you “done,” you can ask follow-up questions. When an agent tells you “done” and disappears into a chat window, you have to audit the whole thing yourself. A receipt is how the agent stays accountable after the run ends.

---

## Why Open Engine uses Linear first

Open Engine’s first public version uses Linear because Linear already has the pieces this needs: issues, assignees, statuses, labels, comments, links, and history.

The free tier covers how most people will start: unlimited people, two teams, and a 250 active-issue cap that a solo operator running a few loops won’t reach for a long time. A growing team will eventually want a paid plan, but you don’t need one to build your first loops.

The public guide walks through the setup in order. You create a small queue, add statuses like Standing, Agent Todo, Agent Working, Agent Needs Input, Agent Review, and Agent Done, add an agent-instructions label, write down the private rules the agent should always follow, give each agent one place to say whether it’s online or paused, install the prompt that tells the agent how to check the queue, and smoke test the loop with a tiny issue.

That sounds like setup because it’s setup.

But notice what the setup is actually doing. It’s teaching a task system to represent agent work in a way both humans and agents can understand.

The status `Agent Working` is the lock. An agent moves the issue there and leaves `AGENT CLAIMED` before it starts, then re-reads the issue so it doesn’t race another agent.

The status `Agent Needs Input` is the pause state. The agent asks one specific question and stops, on the issue or in the owner’s thread, depending on where the answer lives. Once you answer, the next run resumes the original task instead of starting over.

The status `Agent Review` is the honest middle. The agent may have finished the scoped work, but a human still needs to inspect, approve, QA, publish, or decide.

Each agent has one status note. It says whether the agent is online, blocked, manual-only, paused, or recently finished a specific issue.

The smoke test is intentionally tiny. You’re not testing whether the agent is brilliant. You’re testing whether the loop can claim the work, leave a receipt, complete the task, update its status note, and stop after one task.

That last rule matters: process one task per run. Keep failures small. Make the receipt readable. If the agent can’t do one task cleanly, it has no business chewing through a queue.

---

## The 30-minute version

If you don’t want to build the full guide today, do the small version.

Pick one recurring task. Create one issue. Give it the seven fields above. Add three statuses if your tool supports them:
* Todo
* Working
* Needs Input
* Done

Then tell one agent:
1. Read this task.
2. Move it to Working and leave `AGENT CLAIMED`.
3. Do only the scoped work.
4. Do not publish, send, deploy, delete, or make any external or irreversible change without asking me first.
5. If you need one missing fact, or anything is ambiguous, ask one specific question on this same task, leave `AGENT BLOCKED`, and stop.
6. If you finish, leave `AGENT DONE` with what changed, where the output is, what you checked, and what still needs review.
7. Stop after this one task.

That’s enough to feel the difference. Automation comes later. The first win is simpler: the task stops evaporating into chat.

Once you have one loop that can claim, block, resume, and finish with evidence, you can make it more automatic. Add one status note for each agent. Add a private note with the rules the agent should always follow. Add a recurring queue check. Add a routing map for another person’s agent. Add standing updates for shared context. But don’t start there unless you enjoy building systems more than using them.

Start with one loop, make it leave receipts, and then decide whether it deserves to run on a schedule.

---

## What changes for a team

The team version is where Open Engine gets interesting.

Personal agents already create a coordination problem because you may use Claude for one kind of work, Codex for another, and ChatGPT for another. A team multiplies that problem. Now each person may have their own agent, their own permissions, their own files, their own local context, and their own comfort level with automation.

You don’t solve that by forcing everyone into one model.

You solve it by making the handoff explicit.

If my agent needs Adrienne’s agent to review a script, the task should be assigned to Adrienne, because Adrienne owns the agent loop that will see it. The issue should include the source, desired outcome, acceptance criteria, boundaries, and output location. If Adrienne’s agent hasn’t marked itself online, the system should say that before pretending the handoff happened.

That’s the team-level promise: one person’s agent can give another person’s agent a real task without the reason, source, boundary, or proof falling out of the handoff.

This is also why I care about cross-harness coordination.

I don’t want a future where the team has to pick one AI subscription and bend every workflow around it. Codex should do the work Codex is good at. Claude should do the work Claude is good at. Local agents should run where local context matters. Browser agents should inspect what only a browser can inspect.

What they share is the task record, not the model. It’s a smaller architectural claim than “one agent does everything,” and much closer to how real teams work.

---

## Where OpenClaw, Hermes, and Symphony fit

**OpenClaw** matters because it makes the desire obvious: people want local agents that act outside a chat box.

**Hermes-style** systems matter because repeated workflows are where a lot of the value lives. The most useful AI work is often not one amazing answer. It’s the same kind of task, repeated with variation, where the system should learn the pattern and save the human from doing the coordination again.

**Symphony** matters because OpenAI is working on this same class of problem from the Codex side: how issue trackers become control planes for always-on coding agents. That’s serious work, and it shares the same basic assumption: if agents are going to do more work, the work needs a place to live, a way to be claimed, and a way to be reviewed.

I’m not arguing against any of those directions.

I’m saying they don’t remove the need for a task record.

Once agents can act, you need a place to see what they were asked, what they touched, why they stopped, and who owns the next step. Once loops repeat, you need a way to tell whether the loop is safe, stuck, stale, or finished. Once different people have different agents, you need a way for one agent to hand work to another without turning the human into the courier.

Open Engine is the layer I needed for that.

The difference is the starting promise.

Symphony is OpenAI’s open-source Codex orchestration spec, built for people who are comfortable turning an issue tracker into an always-on agent system. OpenClaw is a local assistant world. Hermes-style tools point at scheduled and repeated loops. Open Engine starts with the smallest shared handoff I think an ordinary operator or team can actually use: one queue, one task record, visible limits, visible status, and receipts.

It is not trying to be the one harness where every task lives. It is trying to give agents a shared task list that works across harnesses, while keeping the first principles simple enough that you can use them before you become an infrastructure engineer.

---

## What not to automate first

The fastest way to ruin this is to start with the scariest task because it feels impressive.

Do not start with outward-facing messages that send automatically, billing changes, deleting data, production deploys, or customer refunds. Do not start with anything where “oops” becomes a meeting.

Start with work that benefits from preparation and still has a human decision at the end. Draft before sending. Prepare before publishing. Review before rewriting the canonical file. Summarize before deciding. Flag high-blast-radius problems before fixing them.

That’s how you build trust. You give the system a job where a receipt is valuable, a blocker is useful, and a human can inspect the result without having to undo damage.

The best first Open Engine tasks often end in Agent Review, not Agent Done, and that is fine. Reviewable work is real progress if the agent has gathered the source, preserved the boundary, produced the artifact, and left enough evidence for a human to decide quickly.

---

## The mistakes this prevents

Open Engine is useful because the failures are boring and common:
* The agent says no task exists because the issue is assigned to the wrong person, missing the label, in the wrong status, or titled in a way the runner doesn’t recognize.
* The agent starts the same work twice because it didn’t move the issue into a working state before acting.
* The task board fills with duplicate status comments because each run creates a new check-in instead of updating the one comment that belongs to that agent.
* A blocked task never resumes because the agent treats `AGENT BLOCKED` as a terminal state instead of a pause.
* The agent asks a broad question because the issue didn’t include sources, acceptance criteria, or boundaries.
* The agent tries to publish because nobody wrote the ask-first list into the task.

These aren’t exotic failures. These are Tuesday failures. They’re exactly the kind of failures that make a person say, “I might as well just do it myself.”

We don’t have to make the agent smarter. We have to make the loop less ambiguous.

---

## The point

The reason I care about Open Engine isn’t that I want more software to manage.

I want fewer jobs living in my head.

I want the transcript to become a task without me rewriting the whole story. I want the task to carry the source and boundary. I want the agent to claim the work before acting. I want it to ask one clear question when it’s stuck. I want it to resume from the same record after I answer. I want “done” to come with enough evidence that I don’t have to become the detective.

That is the bar. If an AI tool helps me but leaves me coordinating five follow-up steps, it helped and still left me with the job. If a loop can carry those steps to the point where my judgment is actually needed, that’s different.

* **Open Brain** was about memory: stop making every AI start from zero.
* **Open Engine** was about coordination: stop making the human carry the handoff.

Those are related, but they’re not the same problem. Memory helps the AI know more. Coordination helps the work survive the trip between tools, people, and decisions.

That’s why the boring pieces matter: the task, the source, the boundary, the status, the blocker, the receipt. Not because those words are exciting, but because they’re the difference between six useful AIs and one exhausted person trying to keep all of them pointed at the same job.
