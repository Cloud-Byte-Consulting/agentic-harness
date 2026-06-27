# Codex threading & child threads | Unlock AI

Generated 2.39:1 Codex threading masthead showing parent and child work lanes branching on the right side.

# Threading & child threads.

Threading is how Codex scales past one conversation. A parent can hold the plan while child threads execute focused work, report status, and hand results back.

Use this guide to decide when to steer, when to queue, how to split work into child threads, and how to keep handoffs concrete enough that momentum survives the move.

One thread, one outcome. Everything on this page is that rule, applied.

**The modelA thread is a unit of work, not a chat log.Once you treat threads as disposable, single-purpose work orders — scoped to one project, aimed at one outcome — everything else about running Codex hard starts to make sense.**

These prompts run in the Codex app. Several of them are meta — they ask Codex to manage threads, goals, and child threads — so expect the app to create and reference other threads as they execute.

### Scope every thread to one sentence

A thread holds a prompt, the model's work, and every tool call along the way — and it's scoped to a project at creation. The practical disciplines: pick the project folder before you start the thread, name the thread for its outcome, and when you can't state the goal in one sentence, that's two threads.

Long-lived kitchen-sink threads degrade in quiet ways: stale context crowds out current intent, and the agent starts serving the conversation's history instead of your goal.

You do

Choose the project first, then open the thread. Name it for the outcome. Kill-and-replace threads that have drifted.

The AI does

Tells you when the current thread has accumulated enough conflicting context that a fresh start would serve you better.

Copy prompt
```
<prompt>
  <task>
    Act as this thread's health monitor from here on.

1. Restate this thread's goal in one sentence. If you can't, tell me — that's a finding, and we should split the work.
2. Flag context drift when you see it: instructions early in the thread that now conflict with what I'm asking, or accumulated state that's steering you away from the current goal.
3. When drift is real, say plainly: "This thread should hand off" — and offer to write the handoff summary for a fresh thread.

Right now: assess this thread's current health on those terms.
  </task>
</prompt>
```

**In flightSteering and queueing: talking to a run in progress.While Codex works, your messages either steer — redirect the run now — or queue, waiting for the next safe pause. Which one Enter does is a setting. Knowing the difference is the skill.**

### Choose interruption deliberately

Steering is for course corrections that can't wait: wrong file, wrong approach, stop. Queueing is for everything that's merely next: also do this, then check that. The cost of steering is churn — the agent re-plans mid-stride; the cost of queueing is latency — your input waits for a pause.

The practical default: queue by reflex, steer on purpose. And when a run has gone wrong at the root, neither — stop it and start clean.

![Recreated Codex interface showing a run in flight with a tool stream: reading interface state, generating a timeline, exporting a composition.](/_next/image?url=%2Fguides%2Fcodex%2Fcodex-tool-stream.png&w=3840&q=75&dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
A run in flight. Your next message either joins the queue or grabs the wheel — know which before you hit Enter. Mockup recreation.You do

Check your Enter default in Settings. Steer only when continuing down the current path costs more than re-planning.

The AI does

Confirms how each mid-run message was handled, so you learn the feel of both modes fast.

Copy prompt
```
<prompt>
  <task>
    Teach me steering versus queueing on a live run.

1. Start a multi-step task you can narrate: organize an inventory of this project's docs into a summary file, working step by step (don't write anything yet — plan first, then proceed stepwise on my go).
2. While you work, I'll send at least one message. For each, tell me explicitly: did it queue or did it steer, what my current default is, and what the other mode would have done with it instead.
3. At the end, debrief: which of my interjections were worth an interrupt, which should have queued, and the one rule of thumb you'd give me for next time.
  </task>
</prompt>
```

**DelegationChild threads: the parent plans, the children execute.A parent thread can hold the goal and spawn child threads for the steps — each child focused and disposable, each result returning to one place. This is the pattern that turns Codex from an assistant into a small team.**

### Keep the parent clean

The orchestration pattern that scales: the parent thread owns the goal, the plan, and the review — and does as little execution as possible. Each child thread gets one step, runs it with full focus and its own clean context, and reports back. Parallel children give you the subagent identicons in the UI so you can tell the workers apart.

The failure mode is letting the parent do everything and the delegation become decorative. If the parent's transcript is full of file edits, the children aren't earning their keep.

![Diagram of a parent thread holding a goal and plan, delegating to three child threads that execute in parallel and report results back.](/guides/codex/thread-tree.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
The parent holds the plan; results come home to one place. Children are focused and disposable.You do

Bring goals with separable steps. Review at the parent level; resist diving into children except to debug one.

The AI does

Decomposes the goal, runs the steps in focused child threads, and keeps the parent as the single place where status lives.

Copy prompt

**Show the full prompt**

```
<prompt>
  <task>
    Run this goal with strict parent/child discipline.

Goal: [STATE YOUR GOAL — one sentence, with a verifiable outcome]

Rules:
1. This parent thread does planning, coordination, and review ONLY. No direct execution here.
2. Decompose into steps; for each step, spin up a child thread with a tight brief: its one job, the files it may touch, what done looks like.
3. Run independent steps in parallel children; dependent steps in sequence.
4. As each child reports, post to this thread: step, outcome, anything that needs my eyes.
5. Keep a live status block at the end of your messages: ✅ done / 🔄 running / ⬜ queued, per step.

Start by showing me the decomposition for approval before any child launches.
  </task>
</prompt>
```

**Long-runningGoal mode: work that outlives the sitting.A goal gives Codex a persistent objective with completion criteria — a target it can keep working toward across steps, sessions, and your absences. Standard in the app since May 2026.**

### Write goals with finish lines

Goal mode is only as good as the goal. "Improve the test coverage" is a wish; "every module in src/lib has tests, the suite passes, coverage report saved to docs/coverage.md" is a goal — the agent can check itself against it, and so can you. Completion criteria are the difference between an agent that finishes and an agent that putters indefinitely in the right general direction.

Pair goals with scheduled wake-ups and you have work that genuinely continues while you're elsewhere — which is exactly why the finish line and the boundaries need to be written down, not implied.

You do

Define done before you start: verifiable criteria, hard boundaries, and what the agent should do when blocked.

The AI does

Pursues the goal across sessions, checks itself against the criteria, and stops at the boundaries instead of improvising past them.

Copy prompt
```
<prompt>
  <task>
    Help me turn an intention into a real goal, then run it.

My intention: [DESCRIBE WHAT YOU WANT — rough is fine]

1. Interview me briefly to pin down: the verifiable definition of done (what artifact, what test, what observable state), the boundaries (files/folders/systems you must not touch), and the blocked-protocol (what you do when you hit something needing my input — default: note it, work around it if safe, never guess on anything destructive).
2. Write the goal back to me as: Objective / Done when / Boundaries / When blocked. Get my sign-off.
3. Create the goal and start. At every pause, report against "Done when" — percent of criteria met, not vibes.

If at any point the goal turns out to be wrong (the criteria don't fit reality), stop and renegotiate rather than satisfying the letter of a broken spec.
  </task>
</prompt>
```

**ContinuityHandoffs: how work survives the thread that started it.Threads end, sprawl, or turn out to live in the wrong project — and threads can't move. The handoff summary is how momentum survives: state, decisions, and next actions, packed for a cold start.**

### Hand off like a shift change

A good handoff reads like a nurse's shift report: here's the patient, here's what happened, here's what's next, here's what to watch. Goal, state, decisions with their reasoning, full paths, next actions, traps. A fresh thread with that summary outperforms a long thread with five hundred messages of archaeology — fresh context beats deep context surprisingly often.

Use it when threads sprawl, when work needs to jump projects, and at the end of any session you intend to resume tomorrow.

You do

End meaningful sessions by asking for the handoff. Paste it as message one of the successor thread — started in the right project.

The AI does

Compresses the thread into a cold-start brief, and as the receiving thread, validates the brief against reality before acting on it.

Copy prompt

**Show the full prompt**

```
<prompt>
  <task>
    (For the RECEIVING thread — paste the handoff below it.)

Above is a handoff from a previous thread. Before doing anything:

1. Verify it against reality: check that the paths exist, the described state matches what's on disk, and nothing has changed since it was written. Report discrepancies.
2. Restate the goal and the next three actions in your own words so I can confirm we read it the same way.
3. Flag anything in the handoff you'd want clarified before acting — ambiguities now are cheaper than wrong work later.
4. On my confirmation, execute next action #1.

--- HANDOFF BELOW ---
[PASTE THE HANDOFF SUMMARY HERE]
  </task>
</prompt>
```

The matching generator prompt lives in the hub's project section ('One thread, one outcome') — ask the dying thread for the handoff, then bring it here.

 [](/guides/codex#projects) 

Back to project setup on the main guide →