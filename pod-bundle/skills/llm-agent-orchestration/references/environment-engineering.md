# Environment Engineering for Autonomous Agents

When you put an off-the-shelf agent inside an iterative loop — propose, build,
measure, repeat — a high-leverage lever in this setting is not the agent's prompt
but the *environment* you place it in. This file covers the affordance-based
thesis, the loop structure that operationalizes it, and the four dimensions you
actually tune. The thesis is specific to the optimization-loop setting; it is not
a claim that prompting never matters in general.

## Contents

1. The affordance-based thesis
2. The PREPARE / PROPOSE / IMPLEMENT loop
3. Dimension 1: permissions
4. Dimension 2: artifacts
5. Dimension 3: budgets
6. Dimension 4: human-in-the-loop
7. Why this ages well

## 1. The affordance-based thesis

There are two ways to make an agent better at an open-ended optimization task:

- **Prescribe the workflow** — script the steps, force a fixed plan, constrain
  the agent to a pipeline you designed.
- **Engineer the environment** — leave the agent free to choose its strategy,
  but shape the *affordances* around it so productive moves are easy and harmful
  moves are hard or impossible.

In the optimization-loop setting, the second tends to win, for a structural
reason: a prescribed workflow bakes in *your* assumptions about how to solve the
task, and those assumptions get worse as the base model gets smarter. An
engineered environment amplifies whatever the agent is good at and suppresses
whatever it tends to do wrong, without dictating the path. You are not telling
the agent how to think; you are arranging the room so that good outcomes are the
path of least resistance. (Note the scope: this is about *open-ended search
against a metric*. For one-shot tasks, well-crafted instructions and tool design
still matter a great deal — see collaborative-and-single-agent-patterns.)

Concretely: instead of "first profile, then optimize the hot loop, then
benchmark," you give the agent a sandbox where benchmarking is one trusted call
away, where it cannot touch the grader, where its time budget is visible, and
where a human can glance in. Then you let it decide what to do.

## 2. The PREPARE / PROPOSE / IMPLEMENT loop

The loop has one setup phase and a repeating two-phase round. **PROPOSE is the
fan-in step; IMPLEMENT is the fan-out step** — the inverse of map-reduce, where
the split fans out and the reduce fans in. Getting this orientation right is the
whole point of the pattern, so it is worth stating precisely.

**PREPARE (once).** Before any iteration, validate the world. Confirm the runtime
builds and runs, confirm the evaluator is reachable and returns a score on a
known input, snapshot the starting state. The point is to fail loudly *before*
spending an iteration budget on a broken harness. Treat PREPARE as a gate: if it
does not pass, no rounds run.

**PROPOSE (fan-in).** A *single* proposal session reads the accumulated evidence
— the PREPARE summary, the ranked history of prior solutions, and (optionally)
web search / browsing — and distills all of it into a manifest of **up to P
candidate hypotheses**. This is the fan-in: many sources of evidence collapse
into one ranked set of directions. The diversity you want comes from this one
session deliberately proposing *different* hypotheses, not from running many
proposers in parallel.

**IMPLEMENT (fan-out).** For *each* of the P hypotheses, a **separate
implementation agent session runs in parallel**, each in its own workspace, each
iterating against the hidden evaluator to realize and score its assigned
hypothesis. This is the fan-out. After all implementers finish, the controller
**ranks all valid submissions** and folds the winner into the shared best-state,
which the next PROPOSE round reads.

**Same-round isolation lives in IMPLEMENT, not PROPOSE.** The parallel
implementers in one round may learn from *previous* rounds (the ranked history is
shared) but **cannot inspect or copy peers in the same round**. Isolation is a
property of the fan-out implementers, because they are the ones running
concurrently; the single PROPOSE session has no same-round peers to isolate from.

```text
PREPARE() -> ok or abort
loop until budget exhausted or target met:
    # fan-in: ONE proposal session reads history + research, emits up to P hypotheses
    hypotheses = propose(history, research, max=P)
    # fan-out: P parallel implementers, each isolated from same-round peers
    scored     = parallel_implement(hypotheses, evaluator)   # each in its own workspace
    ranked     = rank(valid(scored))
    best_state = fold_winner(best_state, ranked)
    history.append(ranked)                                   # ranked history, shared next round
```

This is the same fan-out/fan-in machinery as a recursive harness, applied to
*time* (rounds) instead of *space* (input slices) — but with the orientation
flipped: in map-reduce the split fans out and the reduce fans in; here the
PROPOSE distillation fans in and the parallel IMPLEMENT fans out.

## 3. Dimension 1: permissions

Permissions decide what the agent can read, write, and run. Set them so the agent
cannot reach anything that would let it cheat or destabilize the loop:

- **Hidden evaluator service.** The grader runs as a service the agent can *call*
  but cannot *read or modify*. If the agent can see the test cases or the scoring
  code, it will (eventually) optimize the grader instead of the task.
- **Controller-owned files.** The authoritative best-state and the ranked history
  are owned by the controller, not the agent. The agent proposes changes; the
  controller commits them. The agent cannot rewrite the record of its own
  progress.
- **Same-round isolation (among parallel implementers).** The parallel IMPLEMENT
  sessions in one round cannot see each other's in-flight work. (See the
  failure-modes reference on premature convergence.)
- **Default-deny on hardware and external effects.** Network, devices, and
  anything with side effects are denied unless explicitly granted. Grant the
  minimum the task needs.

## 4. Dimension 2: artifacts

Artifacts are the durable objects the loop reads and writes. Design them so state
is explicit and resumable:

- A **ranked-solution history** — every accepted state with its score, ordered.
  This is both the memory of the run and its audit trail, and it is exactly the
  evidence the PROPOSE fan-in reads each round.
- A **current best-state** the next round builds on.
- **Per-round records** of what each implementer built and what it scored, so a
  human (or a later analysis) can see why a direction was taken or dropped.

Because the best-state and history are files, the run is **resumable**: kill it,
restart, and it continues from the last accepted state rather than from scratch.

## 5. Dimension 3: budgets

Budgets must be first-class, not an afterthought, and the agent needs *time
awareness* — but the cost signal is handled differently from the time signal.

- **Active time awareness.** The agent can call a helper to ask how much *time*
  budget remains and adapt (e.g. stop exploring, consolidate gains) as it runs
  low.
- **Passive time awareness.** The harness enforces the budget regardless — a
  round that overruns is cut off, not trusted to stop itself — and injects a
  deadline warning so the agent knows to stop exploring and write its
  deliverables.
- **Expose time; track cost externally.** What the loop actually does is expose
  the *time* budget to the agent (the active checker plus the passive deadline
  warning) while **tracking accumulated API/token cost without exposing it to the
  agent**. The general best-practice reasoning behind keeping cost out of the
  agent's view is that a raw cost signal fed into the objective invites gaming —
  an agent rewarded on cost can learn to emit trivial cheap changes rather than
  solve the task. Time is something the agent can reason about usefully (prioritize
  deliverables under a deadline); cost is enforced as a hard external stop. Expose
  cost only if cost genuinely *is* the optimization target.
- **Resumability** (above) is itself a budget feature: a long run can be paused
  and resumed without losing spend.

## 6. Dimension 4: human-in-the-loop

Autonomy does not mean unattended. Provide **low-friction oversight**:

- A live view of the ranked history and the current best-state, so a human can
  glance at progress without interrupting the loop.
- A cheap **pause/inspect/resume** path (enabled by resumable artifacts).
- The ability to inject a constraint or veto a direction between rounds.

The goal is that supervising the agent costs the human seconds, not a context
switch — oversight you can actually afford to do is oversight that happens.

## 7. Why this ages well

In this setting, every prescribed step is a bet that your decomposition is better
than the model's. Environment engineering makes the opposite bet: that a capable
agent, given clean affordances and hard guardrails, will find a better path than
you would have scripted — and that as models improve, the same environment keeps
paying off without a rewrite. Prescribed *optimization* workflows are technical
debt against the next model; engineered environments are not. (This is a claim
about open-ended search loops specifically, not a blanket dismissal of structured
prompting or fixed pipelines, which remain the right tool for well-understood,
non-search tasks.)
