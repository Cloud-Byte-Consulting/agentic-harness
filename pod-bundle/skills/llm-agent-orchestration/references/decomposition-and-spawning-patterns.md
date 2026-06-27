# Decomposition and Spawning Patterns

How to split a task across child agents and wire the children back together. This is a
cross-cutting synthesis: it pulls the recurring shapes out of two lines of work — recursing
a full agent harness over a long-context workload, and coordinating CLI-agent sessions
through a propose/implement loop for metric-driven discovery — and generalizes them into
decisions you can reuse on any multi-agent or recursive LLM system. It covers the
*independent-work* and *metric-search* branches; for the *collaborate-for-quality* branch
(debate, voting, generator-critic) see collaborative-and-single-agent-patterns.

Read this when you are about to fan work out and are deciding: how to split it, what to spawn
children with, how to keep them from corrupting each other, how to combine their results,
how big each child should be, and how to keep the whole thing from exploding.

## Table of contents

1. A decision tree for decomposition
2. Fan-out/fan-in shapes
3. Tree search as a planning layer (stateful action spaces)
4. Choosing the spawning mechanism by scale
5. Isolation: context isolation vs round isolation
6. Aggregation patterns
7. Right-sizing children with capability tiers
8. Resource control across parallel children
9. Bounding the explosion

---

## 1. A decision tree for decomposition

Before you spawn anything, decide whether to spawn at all, and if so, along which axis. Most
over-engineered agent systems fail here: they fan out work that a single agent could have done
in one context window, paying coordination and token cost for no quality gain. **A single
tool-using agent is the baseline to beat.** Work top-down through these questions.

**Q1 — Does the work fit in one context window with room to reason?**
If the whole input fits and the task is retrieval or light transformation, do not decompose.
A single agent (or even a single model call) is cheaper and has no aggregation error. Decompose
only when the input exceeds the window *or* when the task demands independent reasoning over
many pieces that would interfere with each other if crammed together. (If instead the task is
*one hard question* that would benefit from cross-checking, the right move is a collaborative
pattern, not fan-out — see collaborative-and-single-agent-patterns.)

**Q2 — Is the work a set of independent pieces, or a search for a better solution?**
This is the fork that picks your fan-out shape (Section 2).
- *Independent pieces* over a fixed corpus (one answer per entry, per document, per file) →
  **map-reduce**. Each child owns a disjoint slice; there is one correct decomposition.
- *Searching solution space* for the best artifact against a metric → **iterative
  propose-implement**. There is no fixed slice; you generate candidate directions, build them
  in parallel, score them, and feed the best back into the next round.
- *Searching a large, stateful action space where each action reshapes the landscape* → keep
  the search itself as an explicit, persistent **tree** (Section 3). This is still
  metric-driven search, but the actions are not independent: an applied action changes which
  actions make sense next, a failure is diagnostic signal that should constrain later actions,
  and a success can expose entirely new options. When the *navigation* of the action space is
  the hard part — not the generation of any single candidate — make the tree a first-class
  planning layer rather than a hidden loop.

**Q3 — Are the pieces uniform, or do some pieces hide more work?**
- Uniform pieces (every entry is roughly the same size of subtask) → **flat** fan-out, one level.
- Heterogeneous pieces, where a child may discover its slice is itself a large workload →
  allow **recursive** fan-out: let the child spawn grandchildren rather than choking on its slice.

**Q4 — Does each piece need tools, or just reasoning over text already in hand?**
This decides what the recursive unit *is* (and is the central distinction worth internalizing):
- Reasoning only, no file access / no code execution / no web → the child can be a bare
  **model call**. Cheaper, no harness overhead.
- The child may need to read files, run code, search the web, or itself decompose further →
  the child must be a **full agent harness** (its own context window, filesystem tools, code
  execution, planning loop, and the same spawning capability as its parent).

The model-call-vs-harness choice is the highest-leverage decision in the tree. A harness per
piece buys per-piece tool use and recursive decomposition at the cost of per-piece harness
overhead; a bare model call buys cheap decomposition but caps each child at "reason over the
text I was handed." Pick the smallest unit that can actually finish its slice.

**Q5 — Is there a verifiable score for each result?**
If yes, you can rank children automatically and let the best results drive the next round
(Section 6). If no, you need a typed output contract and deterministic aggregation instead of
ranking. Never let children free-form their results into prose you then have to re-parse with
an LLM judge if a schema would do.

```
                 work to do
                     |
        +------------+-------------+
        | fits one window,         | exceeds window OR
        | retrieval-ish?           | needs independent reasoning per piece?
        |                          |
   single agent          +--------+---------+
   (no fan-out)          | independent      | searching for
                         | pieces over      | a best solution
                         | a corpus         | against a metric
                         |                  |
                    MAP-REDUCE       PROPOSE-IMPLEMENT
                         |             (Section 2c)
            +------------+----------+
            | uniform pieces?       | heterogeneous,
            |                       | pieces hide more work?
         FLAT fan-out         RECURSIVE tree
        (one level)           (children spawn children)
            |
   each piece: tools needed?
       no -> model call    yes -> full harness
```

---

## 2. Fan-out/fan-in shapes

Three fan-out/fan-in shapes cover the large majority of *independent-work and search* systems.
Pick by the answers from Section 1. (A fourth, collaborative family — debate and
generator-critic — is covered in its own reference and is the right choice when the goal is
answer quality on one question rather than throughput. And when the *search* itself is over a
large stateful action space, the planning layer is an explicit search tree — Section 3.)

### 2a. Flat map-reduce

The default for "one corpus, many independent answers." A parent inspects the input, splits it
into N disjoint slices, spawns N children that each produce one structured result, then a reduce
step combines those results into the final answer.

```
parent: inspect input, decide N
        spawn child_1 ... child_N  (parallel, isolated)
        reduce(results) -> final answer
```

When to reach for it: per-entry classification, per-document extraction, per-file analysis —
anything where the decomposition is obvious and there is exactly one right way to slice.
Domain-independent examples: triaging a quarter's worth of support tickets, scanning every file
in a monorepo for a deprecated API, summarizing each of a thousand call transcripts. The parent
should never need to mediate between children; the slices are disjoint by construction.

Keys to getting it right:
- Make slices disjoint and self-contained. A child gets its slice plus the shared context it
  needs, nothing more.
- Children write to a shared output location (one record each); the parent reads only the
  aggregate, not the children's intermediate reasoning (Section 6a).
- The reduce step is deterministic where possible (concatenate, sum, vote), not another LLM
  call you have to trust.

### 2b. Recursive tree

The same as map-reduce, except a child that finds its slice is itself too large does not
struggle — it becomes a parent and spawns its own children. This makes the decomposition
genuinely recursive rather than one fixed level of fan-out. Each node carries the same tools
*and the same spawning capability* as the root, so depth adapts to where the work actually is.

```
root
 +-- child A  (small slice -> answers directly)
 +-- child B  (large slice -> spawns)
 |    +-- grandchild B1
 |    +-- grandchild B2
 +-- child C  (answers directly)
```

When to reach for it: heterogeneous workloads where you cannot predict up front which pieces
hide more work — a codebase where some directories are a single file and some are whole
subsystems, or a document set where some entries are a paragraph and some are book-length. The
win is that you do not have to choose one slice granularity for the whole job; the tree grows
deeper only where needed.

Cost: this is the shape most prone to runaway expansion. It is only safe with a hard depth cap
and width control (Section 9). Without those, one mis-judged "this slice is big" decision at
the root can blow up into thousands of leaves.

### 2c. Iterative propose-implement

For optimization and discovery, not decomposition of a fixed corpus. Instead of slicing input,
you iterate rounds. Each round has a single **propose** step that reads the accumulated evidence
and emits up to P diverse candidate directions, followed by P parallel **implement** steps that
each build and score one candidate. The round's results are ranked and fed back into the next
round's propose step.

```
prepare (once: set up env, validate the scorer)
repeat up to R rounds:
    propose  -> up to P diverse, independently-buildable hypotheses   (fan-in of past evidence)
    implement_1 ... implement_P in parallel, each scored by an evaluator  (fan-out)
    rank all valid results, append to a shared ranked-history file
    stop if budget exhausted or done
```

Note the inversion from map-reduce: here the **propose** step is the fan-in (it distills all
prior rounds plus any external research into the next set of directions) and the **implement**
step is the fan-out. The loop deliberately mixes broad parallel exploration within a round with
sequential accumulation across rounds. (Full treatment in environment-engineering.)

When to reach for it: there is an optimizable metric and a hidden/secure evaluator, the search
space is open, and you want to climb toward a best artifact (a kernel, a model, a packing, a
proof construction) rather than answer a fixed set of questions. A separate one-time **prepare**
stage that validates the runtime and the evaluator before any optimization starts pays for
itself — it stops every later round from building on a broken setup.

---

## 3. Tree search as a planning layer (stateful action spaces)

Propose-implement (Section 2c) assumes each round's candidates are roughly independent attempts
at the same target, scored statelessly. That assumption breaks once the action space is **large
and stateful**: applying an action *changes* which actions make sense next, a failure carries
diagnostic information that should constrain later choices, and a success can shift the
bottleneck so that previously irrelevant actions become the most valuable ones. In that regime
the hard part is no longer *generating* a candidate — it is *selecting* which candidate to try
next, given everything tried so far. The answer is to make the search itself an explicit,
persistent **tree** that serves as the system's shared working memory, and to treat that tree as
a cognition/planning layer rather than a transient loop.

### 3a. The shape

```
root = profiled / measured baseline
loop until scores fall below a threshold or budget is exhausted:
    profile/observe   measure the current state to find the live bottleneck(s)
    score             rank candidate actions by a heuristic (expected gain vs cost/risk + explore bonus)
    select            pop the highest-scored action (depth-first)
    execute           dispatch to a worker that implements it; gate the result (correctness + end-to-end)
    update tree       kept  -> action becomes the new baseline for the next profiling pass
                      reverted -> record a diagnostic annotation on the node
                      crashed  -> trigger root-cause analysis, convert the failure into a constraint
    re-score          the bottleneck distribution may have shifted; re-rank remaining candidates
    backtrack         on revert/crash, return to the last verified state before the next action
```

Two properties distinguish this from a fixed fan-out:

- **The tree expands dynamically.** Branches that did not exist at initialization appear when a
  kept action re-profiles and exposes a new bottleneck. You are not enumerating a fixed candidate
  set; the action space *grows* as the search makes progress and shrinks (via constraints) as it
  hits dead ends.
- **Failures are signal, not just waste.** A revert records *why* the idea did not pay off; a
  crash triggers diagnosis that becomes a constraint pruning whole branches before they are
  tried. A persistent knowledge base of past outcomes can seed the scoring priors of future
  campaigns (warm start), so the search starts from a strictly stronger prior each time.

### 3b. The scoring heuristic

Score each candidate action by **expected gain weighted against cost and risk, plus an
exploration bonus** for under-sampled action categories. A workable decomposition:

- *expected gain* — how much the action is likely to move the metric (e.g. the share of the
  bottleneck it targets times a category-specific prior);
- *cost* — wall-clock to implement, deploy, and re-measure it;
- *risk* — empirical failure/crash rate for that category, refined as outcomes accumulate;
- *urgency* — a multiplier that rises as the gap to target grows;
- *exploration bonus* — a UCB-style term that favors action categories tried few times (with the
  right normalization this recovers UCT when costs and risks are uniform).

This is deliberately MCTS-adjacent: each node carries a value estimate, re-scoring after every
measurement plays the role of value backup, and the explore bonus is the UCB term. The pragmatic
differences from textbook MCTS are worth internalizing — measurements are minutes-to-hours long
so there are **no cheap rollouts** (you cannot simulate; every evaluation is real and gated),
priors come from a persistent knowledge base rather than a learned value network, and the action
set is generated by an LLM rather than fixed offline. A learned value model over accumulated
outcomes is the natural extension once enough history exists. Because the heuristic penalizes
cost and risk, the search naturally exhausts cheap, safe, upper-layer actions before committing
to deep, expensive ones — an ordering that *emerges* from scoring rather than being hand-scripted,
which is exactly the kind of baked-in assumption you want the search to discover rather than
prescribe.

### 3c. The multi-agent structure that sustains it

A long stateful search exceeds any single agent's context and competence, so split it by
**cognitive function** — not by territory:

- A **driver/orchestrator** maintains the tree and runs the loop: profile, score, select,
  dispatch, gate, update. It decides *what* to try and *whether* a result passes verification,
  but does not implement actions itself.
- **Specialists**, ideally *constructed on demand* for each action (their brief composed from the
  target, the environment context, and relevant knowledge-base entries), implement one action
  each with local validation, and return a verified result. Building the specialist per-task —
  rather than maintaining a fixed roster — lets the same nominal role adapt its knowledge and
  constraints to the current context.
- A **critic/guardian** that operates over the *whole trajectory*, not individual actions:
  diagnoses crashes into reusable constraints, decides whether a reverted idea still has merit
  under a different implementation (spawning a refined sub-action when so), and guards
  measurement integrity.

The critic is the load-bearing piece, and the reason is a checks-and-balances tension: the
driver pushes aggressively for metric gains while the critic can veto changes that pass the local
metric but are actually unstable or measured wrongly. This lets the system *deliberately trade
short-term metric for long-horizon viability* — a trade no single aggressive agent makes, and the
thing that keeps multi-hour/multi-day campaigns from destroying themselves. Empirically the
separation matters a great deal: a single unstructured agent with the same tools makes fast early
progress but cannot recover from a state-corrupting failure and terminates irrecoverably; removing
specialist depth caps the gains reachable; and removing the critic produces *invalid* progress —
optimizing confidently toward configurations that skipped correctness gating or were measured under
a mismatched benchmark. The headline lesson generalizes: **measurement integrity is a distinct
job from making progress**, and on any metric-driven autonomous loop something must be empowered
to reject a "win" that is really a measurement artifact or a latent instability. (See
collaborative-and-single-agent-patterns for generator-verifier and planner-critic; see
environment-engineering for the validation-gating and budget machinery; this is the
search-structured, role-by-cognitive-function specialization of those ideas.)

### 3d. When to reach for it (and when not)

Use a stateful search tree when **all** of these hold: the action space is large and you cannot
enumerate a fixed candidate set up front; actions are *coupled* (one changes the value of others)
rather than independent; failures are informative and worth converting into constraints; and the
campaign runs long enough (hours to days) that an explicit, persistent, inspectable state is worth
the bookkeeping. Cross-layer / full-system optimization — where a change at one layer can
invalidate work at another and only end-to-end measurement reveals it — is the canonical fit, and
it is why **every** candidate must be gated on the *real* end-to-end metric, not a local proxy: a
large fraction of locally-validated improvements regress the full system once integrated.

Do **not** reach for it when propose-implement (Section 2c) suffices — independent candidates,
stateless scoring, a short horizon — or when the work is really independent pieces (map-reduce).
The tree adds real bookkeeping (state, annotations, re-scoring, backtracking); pay for it only
when navigating the action space, not generating candidates, is the bottleneck. Note the
formulation is **domain-agnostic** even though it is most visible in systems optimization: the
tree operates over observation-derived bottlenecks and scored actions, so it transfers to any
setting with the coupled-stateful-action structure above. A practical caveat: the scoring
constants (the gain/cost weights, the urgency and exploration coefficients) are usually picked
from early experience rather than tuned, and behavior is generally robust to them — but they are
knobs to revisit if the search mis-prioritizes.

### 3e. Relation to the recursive decomposition tree (Section 2b)

Both are "trees," but they are different objects and should not be conflated:

- The **recursive decomposition tree** (Section 2b) is a *static* hierarchy of *work*: a node
  exists because a slice was too big, its children partition that slice, and the tree's job is
  done when every leaf has produced its piece. Children are independent and the tree does not
  change once built.
- The **search/planning tree** here is a *dynamic* record of *attempts*: a node is an action and
  its measured outcome, the path to a node is the cumulative state, branches appear and get
  pruned as measurements come in, and the tree *is* the shared memory the agents reason over. It
  is never "done" until scores plateau or the budget runs out.

Decomposition trees move *work* in parallel and aggregate; search trees navigate a *state* mostly
depth-first and backtrack. A complex system can nest them — a search-tree action can itself be
implemented by a specialist that fans work out map-reduce style — but keep the two roles distinct
in your design.

---

## 4. Choosing the spawning mechanism by scale

Once you know the shape, decide *how* the parent instantiates children. The mechanism should be
chosen by the number of children, because the two mechanisms have opposite overhead profiles.

**Structured tool-call spawning (small N, roughly 1–5 children).**
The parent emits one structured function/tool call per child directly. No intermediate script,
no codegen overhead. This is the right path when there are only a handful of subtasks: the cost
of writing and running a spawning script would dominate. The ceiling is the provider's per-turn
parallel tool-call budget — you cannot exceed it, which is exactly why this path is for small N.

**Code-execution spawning (large N, tens to thousands of children).**
The parent writes a self-contained script that instantiates every child (e.g. one task object
per piece) and runs them concurrently, then executes that script through its shell tool and
reads back only the aggregated output. Because spawning is now ordinary program code rather than
a fixed call convention or a schema-defined primitive, the parent can parametrize concurrency,
per-child instructions, and output paths in the same language it uses for everything else — and
the child count is just a variable, bounded by the workload rather than any protocol cap. This
is what lets a system scale to thousands of children in one fan-out.

**Selection rule.** Let the parent choose the path automatically from the child count, with a
threshold (a small handful) below which it uses direct tool calls and above which it generates a
script. The point of the threshold is that script-generation overhead should only be paid when
the parallelism actually justifies it.

One failure mode to guard against: at very large or intimidating inputs, a parent may "give up"
on spawning and just answer directly with a regex/retrieval shortcut, collapsing the whole
system back to a single agent and discarding the per-piece reasoning. If per-piece reasoning is
the reason you built the system, make spawning the expected path and watch for parents that skip
it on the largest inputs.

| Children | Mechanism | Why |
| --- | --- | --- |
| 1–5 | direct structured tool calls | no codegen overhead; under the per-turn cap |
| many (10s–1000s) | generated script run via shell | bypasses per-turn cap; count is a variable |

---

## 5. Isolation: context isolation vs round isolation

Isolation is what makes parallel *independent* children safe to combine. There are two distinct
flavors, and you usually want both for independent work. The unifying rule: **a child may look
backward (at finished prior work) and downward (at its own children), but never sideways (at
peers running now) or upward (at the parent's full context).**

### 4a. Context isolation (the spatial axis: no sideways or upward access)

Each child runs in its own workspace with its own context window. It does not see the parent's
full context, and it does not see peer children's intermediate state, reasoning, or files. It
receives only a bounded brief: its assigned piece, the shared context it needs, and the output
format expected.

Why this matters:
- It prevents cross-contamination — one child's scratch work cannot leak into another's
  reasoning and bias it.
- It enables deterministic aggregation: because children communicate only through their final
  written outputs, the parent combines clean records rather than reconciling tangled shared
  state.
- It is what lets you scale out without a shared-memory or message-bus bottleneck — there is no
  inter-child channel to contend on.

Note the scope. This is the right default **for high-fan-out independent work**, where a shared
message thread becomes both a coordination bottleneck and a contamination vector. It is *not* a
verdict on conversational multi-agent designs in general: shared threads, blackboards, and
message buses are exactly the right tool when agents genuinely need to negotiate, debate, or
build on each other's partial results (see collaborative-and-single-agent-patterns). Choose
isolation when the work is independent; choose a shared channel when the collaboration *is* the
work.

### 4b. Round isolation (the temporal axis: no peeking at same-round peers)

In an iterative propose-implement loop, add a second rule on top of context isolation: a child
in the current round may read and learn from *finished prior rounds* (their artifacts, code,
scores) but may **not** inspect or copy from *peer children in the same round*.

Why this matters: if same-round peers could see each other, they would converge toward whichever
direction looked best early, collapsing the round's diversity into a single local basin. The
whole point of running P parallel implementers is to explore P genuinely different directions;
round isolation is what preserves that diversity. Prior rounds are fair game precisely because
they are settled — learning from them is accumulation, not premature collapse.

Put together: **look backward (prior rounds) and downward (your own children), never sideways
(same-round peers) or upward (parent context).**

---

## 6. Aggregation patterns

How children's results come back together. Three patterns, increasingly structured.

### 5a. Shared-file aggregation

Every child writes a structured record (one JSON object, one row) to a shared output location on
completion. The parent collects results by reading that location after all children resolve —
not by capturing each child's stdout or reasoning trace, and not through any inter-process
channel. The parent sees only the final aggregate.

Use this as the default for map-reduce and recursive-tree shapes. It pairs naturally with
context isolation: the written file *is* the only communication channel, so there is nothing to
contaminate. Make each record self-describing (include the piece id) so the reduce step can
assemble them order-independently.

### 5b. Auto-ranking plus history

For propose-implement loops: after each round's children submit scored candidates, the system
automatically ranks all valid submissions and appends them to a persistent ranked-history file.
That file becomes the shared progress memory the next round's propose step reads. Couple the
artifacts with version-control history (commit each child's solution with a message describing
both the current solution and what changed) so later rounds can inspect not just the best score
but the *code and trajectory* that produced it.

This pattern turns aggregation into long-term memory: the system does not merely combine one
round's outputs, it accumulates the best-of across all rounds, so the search has somewhere to
stand at the start of each new round.

### 5c. Typed-contract acceptance

Whatever the shape, define the output contract up front: each child returns a value in an exact,
specified format. Then map raw child output to that contract and accept it deterministically —
exact match, numeric comparison, schema validation — rather than scoring with an LLM judge. If
raw outputs are usually already in the target format, a lightweight extraction step (with a
regex fallback) is enough to normalize the occasional stray; the acceptance/scoring itself stays
deterministic. The discipline here is: keep the *judgment* of correctness out of the LLM and in
code wherever a contract can express it.

---

## 7. Right-sizing children with capability tiers

Two independent sizing decisions: how much work each child owns, and how much capability each
child is granted.

**Work per child.** You do not need one child per piece. The parent can decide how many pieces a
single child handles, trading parallelism for fewer children. More pieces per child means fewer
spawns (less overhead, less re-reading of shared context) but less parallelism and a larger blast
radius if one child fails. Fewer pieces per child means maximum parallelism but maximum spawn
overhead and maximum shared-context re-reads. Tune this knob to the workload; it is the main lever
on cost (Section 8) and it is a parameter in the parent's spawning code, not a fixed architectural
choice.

**Capability per child.** Do not hand every child the maximal harness. Offer a small set of
capability tiers and let the parent pick the cheapest tier that can finish the slice:

- *Read-only / fast tier* — can inspect files and data but cannot modify anything. Use for
  pure lookup and reasoning subtasks. Cheapest, safest, no write blast radius.
- *General tier* — full filesystem read/write plus code execution. Use when the child must
  produce or transform files.
- *Execution / shell tier* — runs commands and reports output. Use for build/test/run subtasks
  that are mostly about invoking a tool.

Crucially, per-piece *reasoning quality* should be tier-independent and mechanism-independent: a
child spawned by a script and a child spawned by a direct tool call, or a read-only child and a
general child, should reason about their assigned piece equally well. The tier restricts what a
child *can do to the world*, not how well it thinks. That separation is what makes it safe to
default to the least-capable tier — you lose blast radius, not answer quality.

---

## 8. Resource control across parallel children

Parallel children contend for the same finite resources — files outside the workspace, GPUs, API
budget, wall-clock time. Control them at the environment level, not by trusting each child to
behave.

**Default-deny on scarce/dangerous resources.** A resource a child should not freely touch should
be invisible by default and reachable only through a controller-owned interface. The general
pattern: route acquisition of any exclusive or dangerous resource through a helper that *tracks
ownership* rather than letting children grab it directly. Two recurring applications:
- *Sandboxed filesystem.* Run each child in a container with only its workspace mounted, so files
  outside the run cannot be modified accidentally or adversarially. Keep authoritative
  artifacts (the evaluator, ground-truth result files) outside the agent-visible workspace and
  expose them only through a controlled service — children can submit and read official scores
  but cannot inspect or tamper with the scorer. Enforce with hooks that block writes to
  controller-owned files.
- *Default-deny on an exclusive device (illustration: GPU lock-tracking helper).* Make the device
  invisible unless acquired through a provided helper API that records lock ownership and
  guarantees each physical device is held by at most one child at a time. This converts
  uncontrolled contention (multiple children silently fighting over one device) into an explicit,
  serialized acquisition. The same shape applies to a rate-limited API key, a database connection
  pool, or a software license seat — anything where two children grabbing it at once causes
  thrashing or failure.

**Budget as an environment setting, on two axes.** Track wall-clock time and API/token cost as
first-class limits the environment enforces:
- Time — allow different limits for different stages (a short cap for proposal/planning, a long
  cap for implementation). Make children time-aware both *actively* (a helper API they can call
  to check elapsed/remaining time) and *passively* (the environment injects a warning as a
  deadline approaches, telling the child to stop exploring and write its required deliverables).
  The passive nudge matters: an agent left to its own sense of time will often over-explore and
  produce nothing before the cutoff.
- Cost — accumulate token usage across all children and abort the run when the limit is hit,
  preserving the current workspace as the final snapshot. As a rule, do *not* surface raw token
  spend to the agent; expose remaining *time* (which it can reason about usefully) but keep cost
  as a hard external stop, because a cost signal folded into the objective invites gaming.

Two cost levers worth calling out:
- *Prompt caching.* When every child shares a common context prefix (the same corpus, the same
  task framing), caching that prefix is the single biggest cost reduction available — the
  dominant recurring cost in high-fan-out work is each child re-reading the shared context.
- *Latency is bounded by parallelism, not child count.* Because children run concurrently, the
  parent waits only for the slowest child, not the sum. Spawning more children does not linearly
  increase wall-clock time — but it does linearly increase cost, which is why the work-per-child
  knob (Section 7) and budget caps are the real governors.

Budget control doubles as an operational interface: persist each stage's session id, status,
elapsed time, and remaining budget so an interrupted run can resume from the latest filesystem
state instead of restarting, and so a human can grant extra time to a stage that ran out before
producing its deliverable. (Retries, timeouts, and fallback routing for the *individual* child
calls live in reliability-and-operations.)

---

## 9. Bounding the explosion

Recursive and high-fan-out systems can expand without limit if nothing stops them. Three knobs
bound the total work; set all three explicitly rather than discovering their absence in a runaway
run.

**Depth cap (bounds the tree's height).** Cap recursion depth with a small configurable limit (a
default of around 3 is a reasonable starting point). This prevents unbounded tree growth while
still allowing multi-level decomposition where the workload genuinely warrants it. Every node
inherits the spawning capability, so without a depth cap a single chain of "this slice is still
too big" judgments recurses forever. The cap is the backstop that makes it safe to give every
child the power to spawn.

**Width control (bounds the fan-out per level).** Bound how many children any one parent spawns
per step (the P in a propose-implement round; the slice count in map-reduce). Width is the lever
that trades exploration breadth against per-round cost. Set it from the workload and the budget,
and remember it interacts with work-per-child (Section 7): the same total work can be N narrow
children or N/k wider children. Wider is cheaper per unit of coverage but riskier per child.

**Budget bounds (bounds the total, regardless of shape).** The time and cost caps from Section 7
are the third knob and the ultimate backstop. Depth and width bound the *structure*; budget
bounds the *total consumption* no matter how the structure plays out. Even with a sane depth cap
and width, a pathological workload can still burn the budget — the cost cap is what guarantees
termination. Always set it; treat it as the floor of safety beneath the structural knobs.

Set all three together. Depth without width still allows a wide shallow blowup; width without
depth still allows a deep narrow one; both without a budget cap still allow a workload that
respects the structure but runs far longer or more expensively than intended. The three knobs are
complementary, not redundant.
