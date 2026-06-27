# Single-Agent and Collaborative Patterns

The other half of the design space. Much of this skill is about *isolating*
independent work and aggregating it deterministically — the right move for
high-fan-out throughput. This file covers the opposite end: when **one
tool-using agent** is the correct answer, and when having agents **collaborate**
(talk, debate, vote, critique) genuinely improves the *quality* of the answer
rather than just the throughput. These are first-class, well-established patterns,
not afterthoughts.

The governing question: **are you producing many independent answers (throughput)
or one better answer to a hard question (quality)?** Throughput wants isolation;
quality often wants collaboration.

## Contents

1. The single tool-using agent (the baseline)
2. When one agent is the right call
3. Why collaboration can improve reasoning
4. Debate / multi-agent discussion
5. Adaptive debate: route, stop, and aggregate (debate as conditional computation)
6. Voting, ensembling, self-consistency
7. Planner-critic and generator-verifier
8. Blackboard, message bus, and shared context
9. When collaboration wins vs when isolation wins
10. Cost discipline for collaborative patterns

---

## 1. The single tool-using agent (the baseline)

Before any multi-agent design, the default is **one agent running a ReAct-style
loop**: reason, act (call a tool), observe the result, repeat until done.

```text
loop until done or budget hit:
    THOUGHT   reflect on the goal and what is known so far
    ACTION    call one tool (search, read, run code, write a file, ...)
    OBSERVE   read the tool result back into context
finalize: produce the answer / artifact
```

This single loop — sometimes split into an explicit **plan-then-act** variant
where the agent drafts a plan up front and then executes steps against it — is
astonishingly capable on its own. A large fraction of "we need a swarm" problems
are really "we need one agent with good tools, a clean context, and a tight
stopping rule." It has no coordination overhead, no aggregation error, one trace
to debug, and one budget to bound. **Make every more-elaborate architecture beat
this baseline on quality and cost before you adopt it.**

Variants worth knowing:

- **Plan-act-observe.** Add an explicit planning step that produces a checklist
  the agent works through; good when the task has clear sub-steps and you want
  the plan to be inspectable. Re-plan when an observation invalidates the plan.
- **Reflection / self-refine (still one agent).** After producing a draft, the
  *same* agent critiques its own output against the goal and revises. This is the
  cheapest taste of "verification" — one extra pass, no second agent — and often
  catches a meaningful share of errors before you reach for a separate critic.

## 2. When one agent is the right call

Prefer a single agent when:

- The whole input fits in the context window with room to reason.
- The task is sequential and stateful — each step depends on the last — so there
  is nothing to parallelize.
- The task is retrieval or light transformation, where decomposition only adds
  aggregation risk.
- Latency and simplicity matter more than squeezing the last few points of
  quality.
- You have not yet established that anything more complex helps. (You almost never
  have, at the start.)

Reach beyond one agent only when one of the three front-door branches applies:
throughput/context limits (decompose), open-ended metric search (propose-implement),
or quality-through-perspectives (collaborate — the rest of this file).

## 3. Why collaboration can improve reasoning

Isolation is correct when pieces are independent. But many hard problems are *not*
independent pieces — they are one question where a single forward pass is
error-prone, and a second viewpoint reliably catches what the first missed. The
mechanisms by which collaboration helps are real and studied:

- **Error correction by a fresh viewpoint.** An independent agent (or the same
  agent in a critic role) catches arithmetic slips, missed constraints,
  hallucinated facts, and unstated assumptions that the generator is blind to.
- **Variance reduction by aggregation.** Sampling several independent attempts and
  taking the consensus cancels idiosyncratic mistakes — the same statistical
  reason an ensemble beats a single model.
- **Exploration of distinct strategies.** Different agents (or different prompts)
  attack the problem from different angles; the best of several genuinely
  different attempts beats the single most-likely one.
- **Adversarial pressure.** Forcing a model to *defend* an answer against a
  challenger surfaces weak reasoning that a cooperative single pass would gloss.

These are quality wins, not throughput wins, and they are exactly where the
"isolate everything" advice does *not* apply. The cost is real (more calls, more
latency, a resolution mechanism) — so treat collaboration as a deliberate
quality-for-cost trade, measured against the single-agent baseline.

## 4. Debate / multi-agent discussion

**Pattern: two or more agents propose answers and then critique and respond to
each other over a few rounds; a judge (or a final round) settles on an answer.**

```text
round 0: each agent answers the question independently
round 1..R: each agent sees the others' answers + critiques, then revises
settle: a judge agent (or majority of the final round) picks/ synthesizes
```

Debate tends to help on tasks with a verifiable or defensible answer — math,
logical reasoning, factual questions, code correctness — because a wrong answer
is hard to defend against a correct challenger. Keys:

- **Diversity is the fuel, and prompt diversity is not enough.** If all debaters
  share a prompt, model, and temperature, they agree on the same wrong answer
  (sycophancy / herding). Vary the role, the prompt framing, or — most
  effectively — the *model family*. Agents drawn from the same backbone share a
  parameter space, so they also share factual blind spots and reasoning biases; on
  knowledge-intensive tasks, *role-prompt* diversity over one backbone behaves much
  like a single agent (a "knowledge-parametric echo chamber") and can *underperform*
  a plain single chain-of-thought pass, because repeated interaction amplifies the
  shared error. Genuine model heterogeneity — different families as the debaters —
  is what actually de-correlates the errors that debate is supposed to cancel.
  Treat heterogeneity (who debates) and control (whether/how long they debate) as
  *separate* design choices; conflating them is a common MAD mistake.
- **Bound the rounds.** Most of the gain is in the first one or two exchange
  rounds; cost grows linearly and returns diminish fast. 2–3 rounds is a common
  sweet spot.
- **Have an explicit settle step.** Do not let debate run to "consensus" (it may
  converge on a confident error). A separate judge or a final-round vote is the
  resolution path.

Anti-pattern: debate on a subjective task with no defensible answer — it produces
expensive, articulate disagreement with no convergence. A second, subtler
anti-pattern — running debate *unconditionally* on every input — is the subject of
the next section.

## 5. Adaptive debate: route, stop, and aggregate (debate as conditional computation)

Fixed-round debate applies the same procedure to every input, and that uniformity
is itself a failure mode: it **wastes computation** on easy inputs where the agents
already agree, and on hard inputs it can **amplify conformity** by pushing agents
toward a confident-but-wrong consensus over extra rounds. The fix is to treat
debate as **conditional computation** — a controlled decision process, not a fixed
number of rounds — governed by three coupled decisions: **WHO** participates,
**WHEN** debate is triggered (and when it stops), and **HOW** the final answer is
aggregated. These map onto three lightweight, *training-free* controls that sit on
top of any heterogeneous debate:

- **Pre-debate agreement routing (WHEN to start).** Collect each agent's
  *independent* first-pass answer, normalize to a comparable form (option label,
  extracted numeric/symbolic answer), and compute the agreement score as the vote
  share of the most common answer. If agreement clears a threshold (e.g. a
  two-of-three majority), **skip debate entirely** and aggregate the first-pass
  answers; only route the low-agreement inputs into multi-round debate. Initial
  agreement among *heterogeneous* agents is a strong signal that an input is easy —
  high-agreement inputs answered without any debate are typically very accurate,
  while the low-agreement subset is where deliberation actually pays. Agreement is
  a better routing signal than a model's *self-reported confidence*, which is poorly
  calibrated; routing on confidence can collapse on the hardest inputs.
- **Early-agreement stopping (WHEN to stop).** For inputs that do enter debate,
  re-check agreement after each round and **terminate as soon as the agents
  converge** (e.g. all normalized answers match) rather than always running to the
  round cap. This caps the conformity risk of extra rounds and recovers cost on
  inputs that settle quickly.
- **Outlier-aware aggregation (HOW to settle).** Instead of plain majority vote,
  down-weight answers that are *semantically isolated* from the rest — score each
  answer by its average similarity to the others and zero out clear outliers before
  selecting the winning cluster. Always keep a **majority-vote fallback** for the
  case where every answer looks like an outlier (high disagreement), so aggregation
  never discards all the evidence. Caveat: outlier suppression assumes the isolated
  answer is the unreliable one, which fails in the **"lone expert"** case where a
  single minority agent is right because it alone has the needed knowledge — so it
  is a modest refinement over voting, not a substitute for genuine verification.

Cost model: the routing/stopping controls only govern the *extra* debate cost. You
always pay the K independent first-pass calls (one per agent); a fixed-round design
then pays K·(1+R) calls for *every* input, whereas the adaptive design pays the
full debate cost only on the inputs that are routed in and do not stop early.
On high-agreement workloads this is a large saving (most inputs skip debate
outright); on genuinely hard workloads where almost everything is routed in, the
saving shrinks toward zero — the benefit is **input-dependent**, and the win is an
*accuracy-efficiency* improvement over fixed-round debate using the same agents,
not a guaranteed cost cut. Two practical notes: reliable **answer normalization**
matters, because inconsistent formatting (e.g. stray numeric formats) can hide real
agreement and force needless debate; and a voting-only baseline (heterogeneous
first-pass answers, no debate) is the control to beat — keep debate only if
selective deliberation measurably improves on it.

Generalize the principle beyond debate specifically: **before paying for any
expensive collaborative round, check whether cheap independent passes already agree,
and spend deliberation only where they disagree.** The same route-stop-aggregate
shape applies to voting (Section 6) and verifier loops (Section 7) — gate the
costly mechanism on a cheap disagreement signal rather than running it
unconditionally.

## 6. Voting, ensembling, self-consistency

**Pattern: sample the *same* question N times (independently), then aggregate by
majority vote, score, or a judge.** The cheapest and often highest-ROI
collaborative pattern, because the samples need no coordination at all.

- **Self-consistency.** Sample N reasoning chains at non-zero temperature, take the
  majority *final answer* (marginalizing over the differing reasoning paths). Strong
  on arithmetic and multi-step reasoning where there are many correct paths to one
  answer.
- **Best-of-N with a verifier.** Generate N candidates, then a deterministic checker
  or an LLM judge picks the best. Ideal when verifying is easier than generating —
  unit tests for code, a numeric check for a calculation, a rubric for an essay.
- **Ensembling across models/prompts.** Combine answers from different backbones or
  prompt variants; reduces single-model idiosyncratic error.

Aggregation is mostly *deterministic* (count votes, run the checker), which keeps
this pattern auditable — it sits comfortably with the "deterministic computation
vs LLM judgment" split elsewhere in this skill. The lever is N: quality climbs
with N but with diminishing returns, while cost is strictly linear. Tune N to the
value of being right.

## 7. Planner-critic and generator-verifier

**Pattern: split the work into a producing role and a checking role, run as a short
loop.** Unlike debate (symmetric peers) these are *asymmetric* by design.

- **Planner-critic.** A planner proposes a plan or approach; a critic pokes holes
  (missing steps, wrong assumptions, edge cases); the planner revises. Useful
  before expensive execution — catching a flawed plan is cheaper than running it.
- **Generator-verifier.** A generator produces an artifact (code, a proof, a
  structured extraction); a verifier checks it against a spec, tests, or a rubric
  and returns pass/fail-with-reasons; the generator fixes and resubmits.

```text
loop up to K times:
    artifact = generate(task, prior_feedback)
    verdict  = verify(artifact, spec)      # tests / rubric / checker
    if verdict.passes: return artifact
    prior_feedback = verdict.reasons
return best_so_far (and surface that it did not fully pass)
```

Why asymmetry helps: **verification is often easier and more reliable than
generation.** A verifier grounded in something concrete (a test suite, a schema, a
retrieved source) is a far stronger signal than a peer's opinion. Where the
verifier can be *deterministic* (compile + run tests), this becomes one of the most
reliable patterns available — the LLM generates, code judges. Bound K and always
have a "did not converge" exit so the loop cannot run forever.

## 8. Blackboard, message bus, and shared context

When collaboration needs more than a fixed pairwise exchange, you need a
*coordination substrate*. Three common designs:

- **Blackboard.** A shared, structured workspace that all agents read and write.
  Each agent watches for state it can act on, contributes its piece, and the
  solution accretes on the board. Good for problems assembled from heterogeneous
  contributions (a specialist for each part) where the order of contributions is
  not known up front. The board *is* the shared memory; keep entries typed and
  attributed so contributions are auditable.
- **Message bus / pub-sub.** Agents communicate through typed messages on channels
  rather than a shared store. Decouples producers from consumers and scales to many
  agents, at the cost of having to reason about message ordering and delivery. Good
  when roles are stable and the interaction is event-driven.
- **Shared-context / group-chat.** All agents share one growing transcript (the
  simplest substrate). Easy to build and natural for debate, but the transcript
  becomes a *cost and contamination* surface as it grows — every agent re-reads
  everything, and one agent's error propagates to all. This is precisely the
  substrate that high-fan-out independent work should avoid; it is fine for a small,
  bounded collaboration.

Whatever the substrate, give collaboration a **resolution path** (a judge, a vote,
an arbiter, a verifier) and a **stopping rule**. Open-ended chatter with no settle
step is the most common way a collaborative design burns budget without converging.

## 9. When collaboration wins vs when isolation wins

A blunt decision table. "Quality" means a better answer to one question; "throughput"
means many independent answers.

| Situation | Use | Why |
|---|---|---|
| Many independent pieces, one answer each | **Isolation** (map-reduce / tree) | No cross-talk needed; threads are pure overhead and a contamination vector |
| Open-ended search against a metric | **Round isolation** (propose-implement) | Same-round isolation preserves the diversity the search depends on |
| One hard reasoning/math/code question | **Voting / self-consistency** or **generator-verifier** | Aggregation cancels idiosyncratic error; verification is easier than generation |
| Question with a defensible answer, single pass unreliable | **Debate** (2–3 rounds + judge) | Adversarial pressure surfaces weak reasoning; wrong answers are hard to defend |
| A *stream* of such questions of mixed difficulty | **Adaptive debate** (heterogeneous agents + agreement routing) | Skip debate when independent first passes already agree; spend rounds only on the disagreeing minority |
| Expensive execution gated by a fragile plan | **Planner-critic** | Catching a bad plan is cheaper than running it |
| Heterogeneous contributions assembled into one solution | **Blackboard** | Lets specialists contribute in any order onto shared structure |
| Anything you have not yet shown needs more than one agent | **Single agent** | The baseline to beat; cheapest and most debuggable |

Two failure modes bracket this table, and both are real:

- **Reflexive multi-agent** (over-collaboration): adding agents/debate to a task a
  single agent handled fine, paying coordination cost for no quality gain.
- **Reflexive isolation** (under-collaboration): forcing isolation on a
  one-hard-question task that voting or a critic would measurably improve, because
  "isolate everything" got cargo-culted from the throughput setting.

The skill's isolation-first guidance is calibrated for throughput and search. For
quality on a hard single question, collaboration is frequently the better bet —
just measure it.

## 10. Cost discipline for collaborative patterns

Collaboration multiplies calls; treat the cost as a budget line, not a free lunch.

- **Bound the rounds / samples.** N for voting, R for debate, K for verifier loops
  — set each from the value of being right, and stop early when a deterministic
  verifier already passes.
- **Gate the expensive mechanism on a cheap signal (conditional computation).**
  Do not run debate/voting/verification *unconditionally*. Take cheap independent
  first passes, measure their agreement, and route only the low-agreement inputs
  into the costly path; stop deliberation the moment agents converge (Section 5).
  This makes the spend scale with input difficulty instead of with a fixed round
  count, and it pairs with heterogeneous agents so the agreement signal is meaningful.
- **Tier the roles.** A cheap fast model can often *generate* candidates while a
  stronger model *verifies* or *judges* (or vice versa). Do not pay top-tier price
  for every role. (See cost-tiering in reliability-and-operations.)
- **Prefer deterministic verifiers.** A test suite or schema check is free of LLM
  cost and more reliable than an LLM judge; reach for an LLM verifier only when no
  deterministic check exists.
- **Cache and dedup.** Identical sub-queries across debaters or samples can be
  served from a result cache (see reliability-and-operations).
- **Always compare against the single-agent baseline on cost too.** A collaborative
  design that adds two points of quality for 5x the cost is the right call for a
  high-stakes answer and the wrong call for a routine one. State the trade.
