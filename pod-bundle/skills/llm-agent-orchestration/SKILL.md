---
name: llm-agent-orchestration
description: >-
  Design and operate single-agent, multi-agent, recursive, and autonomous LLM systems. Use
  when building an orchestrator that spawns subagents or workers, decomposing a job into
  parallel subtasks, or reasoning over inputs larger than the context window (long-context
  aggregation, map-reduce). Also for a single tool-using agent (ReAct / plan-act-observe),
  collaborative patterns (debate, voting/self-consistency, generator-critic) including
  heterogeneous agents and agreement-routed conditional debate, tree search as a planning
  layer over large stateful action spaces, agent memory, reliability and ops (retries,
  fallback routing, tracing, caching, tool/MCP design), and input safety (prompt injection,
  PII, permission scoping). Covers code-driven vs tool-call spawning, recursion and cost
  control, context isolation, propose-implement loops, environment engineering, and reward
  hacking. Triggers on orchestrate subtasks, spawn workers, debate routing, planning tree,
  swarm, even when agent absent.
---

# LLM Agent Orchestration

A design playbook for engineers building agentic LLM systems — from a single
tool-using agent up through multi-agent, recursive, and autonomous loops. Its
central thesis: **with the model held fixed, the largest and most attributable
reliability gains come from architecture, not prompt wording** — the harness
that spawns work, the environment it runs in, the knowledge layer it reasons
over, and the operational scaffolding (retries, budgets, observability) around
it. Treat the model as a fixed component and engineer the system around it.

## What this skill covers and what it does not

This skill is an **engineering playbook for the system around the model**:
decomposition, spawning, isolation, environment design, knowledge substrate,
memory, reliability/ops, input safety, evaluation, and failure diagnosis. It is
deliberately opinionated toward **operability** — making systems cheap to run,
audit, and debug.

It is **not** a guide to prompt engineering, fine-tuning, RAG-chunking
heuristics, or model selection on capability grounds (it treats the backbone as
fixed). It does not claim multi-agent always beats single-agent — the opposite
is often true. Several patterns here are distilled from specific research lines
(recursive harnesses, environment-engineered optimization loops, knowledge-graph
substrates); where a claim is source-specific it is flagged as such, and the
generalizable principle is stated separately. Use the principle; treat the
paper-specific number or mechanism as an illustration, not a law.

## When to use this skill

Reach for this skill when you are:

- Deciding **whether you even need multi-agent** — and choosing among a single
  tool-using agent, a collaborative pattern (debate/voting/generator-critic), or
  a fan-out decomposition. Including *how* to run a collaborative pattern
  efficiently: heterogeneous agents and agreement-routed (conditional) debate
  instead of fixed rounds on every input.
- Architecting a parent/orchestrator that spawns subagents or workers, and
  deciding how it spawns, recurses, and aggregates results.
- Decomposing a task across many parallel subtasks: map-reduce over documents,
  per-entry classification, batch reasoning over a large list.
- Reasoning over an input larger than the context window, where you need both
  large-scale navigation and faithful per-slice reasoning.
- Building an autonomous loop that proposes, implements, evaluates, and iterates
  (research agents, self-improving pipelines, optimize-against-a-metric jobs) —
  including searches over **large, stateful action spaces** where the navigation
  is best held as an explicit search tree with gain/cost/risk-scored actions and a
  separate critic guarding stability and measurement integrity.
- Engineering the environment an agent runs in: sandboxing, permissions, hidden
  evaluators, budgets, resumability, and human oversight.
- Managing **agent memory and context**: history compaction, working/episodic/
  semantic memory, eviction, when to persist vs discard.
- Hardening **reliability and operations**: retries/backoff, timeouts,
  idempotency, partial-failure recovery, circuit breakers, fallback model
  routing, cost tiering, tracing across an agent tree, caching, tool/MCP design.
- Defending against **adversarial inputs**: prompt injection via retrieved
  content and tool output, output filtering, PII, tool-permission scoping.
- Organizing the knowledge agents retrieve over: graphs, memory, multi-source
  retrieval, provenance, and stable identifiers.
- Evaluating or verifying an agent system, or diagnosing why a multi-agent
  system underperforms a single agent.

Trigger words: orchestrate subtasks, spawn workers, swarm, scale to thousands,
long-context aggregation, self-improvement, auditable runs, agent debate,
debate routing, heterogeneous agents, tree search / planning over actions,
retries/backoff, fallback model, prompt injection, agent memory — even when the
word "agent" never appears.

## The front door: do you even need multi-agent?

**The default baseline is a single tool-using agent** (a ReAct / plan-act-observe
loop). Most tasks that *feel* like they need a swarm do not; a single agent in a
well-engineered environment, with good tools and a tight context, is cheaper,
easier to debug, and frequently wins outright. Make any more elaborate
architecture earn its keep against this baseline on **both** the quality metric
**and** cost. Three reasons to go beyond one agent — they are not mutually
exclusive:

1. **Throughput / context limits → decompose (map-reduce or recursive tree).**
   The input exceeds the window, or there are many independent pieces that would
   interfere if crammed together. You fan out *independent* work and aggregate
   deterministically. (See decomposition-and-spawning-patterns and
   recursive-harnesses.)

2. **Open-ended search against a metric → propose-implement loop, or a stateful
   search tree.** There is no fixed corpus to slice; you generate candidate
   directions, build them in parallel, score against an evaluator, and feed the
   best forward. When the actions are *coupled* — each one reshapes which actions
   make sense next, failures are diagnostic, and successes expose new options —
   promote the search itself to an explicit, persistent tree that acts as the
   shared planning/working memory, paired with a critic that guards measurement
   integrity. (See environment-engineering and
   decomposition-and-spawning-patterns.)

3. **Quality through multiple perspectives or verification → collaborate.**
   Debate, self-consistency voting/ensembling, planner-critic, and
   generator-verifier loops genuinely *improve reasoning quality* on many tasks
   (math, code, factuality, adversarial robustness) by having models check,
   challenge, or vote on each other's work. This is a real win and a deliberate
   counterpoint to the isolation-first patterns elsewhere in this skill: when the
   goal is a *better answer to one hard question* rather than throughput over many
   independent pieces, collaboration often pays. Make it earn its cost twice over:
   prefer *heterogeneous* agents (different model families, not just role prompts,
   which share a backbone's blind spots) and treat debate as *conditional
   computation* — route to it only when cheap independent passes disagree, stop on
   convergence — rather than running fixed rounds on every input. (See
   collaborative-and-single-agent-patterns.)

The isolation-first, no-sideways-communication advice in much of this skill is
calibrated for **branch 1 (high-fan-out independent work)** and **branch 2
(diversity-preserving search)** — settings where shared threads are a bottleneck
and a contamination vector. It is *not* a universal verdict on collaboration:
branch 3 is exactly where agents talking to each other earns its cost.

## Core concepts

1. **The single agent is the baseline; complexity must beat it.** Start with one
   tool-using agent in a clean environment. Add agents only when throughput,
   independent reasoning, metric-driven search, or verification demands it — and
   measure against the baseline on quality *and* cost.

2. **The recursive unit is a choice — and "tree" means two different things.**
   When you do fan out, recursing over a full agent *harness* (tools + code
   execution + planning) dominates recursing over bare model calls for tool-needing
   work. Give children the parent's spawn capability and a flat fan-out becomes a
   depth-bounded *decomposition tree* — a static hierarchy of independent work that
   attacks arbitrarily large inputs. Distinct from this is a *search/planning tree*:
   a dynamic, persistent record of attempted actions and their measured outcomes
   that the agents reason over to navigate a large stateful action space. The first
   moves work in parallel and aggregates; the second navigates a state depth-first
   and backtracks. Do not conflate them.

3. **Spawn with code, not just tool calls, at scale.** A parent-generated script
   that gathers tasks concurrently escapes the provider's per-turn parallel
   tool-call cap, so fan-out scales with the workload rather than the API limit.
   Keep a direct tool-call path for tiny workloads and auto-select by size.

4. **For independent work, isolate context and aggregate through artifacts.**
   Each child sees only its slice plus relevant excerpts and an output spec — no
   parent or sibling visibility. It writes a structured record to a shared path
   the parent reads deterministically after all children resolve. (For
   *collaborative* work, the opposite holds — see the front door, branch 3.)

5. **Bound recursion and cost explicitly.** Cap tree depth; control fan-out width
   via entries-per-child. Latency is bounded by the slowest child, while cost is
   dominated by re-reading shared context — so prompt-cache the prefix.

6. **In the optimization-loop setting, environment engineering beats workflow
   prescription.** When an off-the-shelf agent iterates against a metric, shaping
   permissions, artifacts, budgets, and oversight around it tends to age better
   than scripting every step, because a prescribed workflow bakes in assumptions
   that get worse as the base model improves. (This is a setting-specific thesis,
   not a claim that prompting never matters.)

7. **Separate deterministic computation from LLM judgment.** Scoring, traversal,
   joins, normalization, and validation are *code*. Reserve the LLM for
   open-ended reasoning and composition. This keeps results stable and decouples
   evaluation from the model family.

8. **Knowledge and memory orchestration are the other half.** Stable-ID joins,
   semantic anchors, typed relations, durable-vs-ephemeral separation, and
   intent-routed retrieval make agent reasoning reliable; working/episodic/
   semantic memory and history compaction keep a long-running agent coherent.
   Agent orchestration moves *work*; knowledge orchestration moves *facts*.

9. **Operability is a first-class concern.** Retries with backoff, timeouts,
   idempotency, fallback model routing, tracing across the agent tree, caching,
   and a well-designed tool/MCP interface are what separate a demo from a system.

## Workflow

**Stage 0 — Justify the architecture.** Start from a single tool-using agent.
Pick a branch (decompose / search / collaborate) only if the baseline cannot
meet the bar, and keep the baseline as the thing to beat on quality and cost.

**Stage 1 — Frame the decomposition** (branches 1–2). Confirm the subtasks are
genuinely independent. Pick the recursive unit (harness vs model call). Decide
whether children may themselves spawn, and if so set a depth limit up front.

**Stage 2 — Choose the spawning mechanism.** Estimate fan-out width. Below a
small threshold, use direct parallel tool calls; past it, switch to code-driven
spawning. Expose spawning as a callable harness primitive, and offer capability
tiers (cheap/fast vs full) per child.

**Stage 3 — Isolate and aggregate** (independent work). Bound each child's
context to its slice, relevant excerpts, and an output contract. Aggregate via
shared structured artifacts read only after all children resolve. When a strict
output format is required, normalize first, then score separately. (For
collaborative work, design the channel — debate transcript, vote, shared
blackboard — instead.)

**Stage 4 — Engineer the environment** (autonomous/iterative systems). Run a
one-time PREPARE pass that validates the runtime and the evaluator. Structure
iteration as fan-in PROPOSE / fan-out IMPLEMENT rounds. Set permission
boundaries: hidden evaluator service, controller-owned files, same-round
isolation among parallel implementers, default-deny on hardware. Make budgets
first-class (expose time, track cost externally, resumable). Provide low-friction
human oversight.

**Stage 5 — Organize knowledge and memory.** Key everything by stable
identifiers. Route cross-modal links through anchor nodes. Fuse multiple
retrieval sources with intent routing. Expose deterministic typed operators
beneath an LLM composition layer. Decide your memory tiers (working/episodic/
semantic), compaction policy, and what persists vs is discarded.

**Stage 6 — Build the operational layer.** Add retries/backoff, timeouts,
idempotent subtasks, circuit breakers, and fallback model routing. Add tracing
(spans, token/latency dashboards) across the agent tree. Cache beyond the prompt
prefix (semantic/result caching, dedup identical subtasks). Design the tool/MCP
interface deliberately. Scope tool permissions against malicious input.

**Stage 7 — Evaluate and verify.** Hold the model backbone fixed across variants.
Report point estimates with bootstrap confidence intervals that exclude zero.
Make runs auditable with typed output contracts and on-disk manifests.

## Common pitfalls

- **Reaching for multi-agent reflexively.** Adding agents adds coordination cost,
  failure surface, and debugging pain. Keep a single-agent baseline and justify
  the architecture on metric *and* cost; name what the structure actually buys.
- **Skipped decomposition.** The parent answers directly and collapses the
  multi-agent system to a single agent — worst at long inputs. Detect whether a
  spawn actually ran; force the decomposition path.
- **Hitting the tool-call ceiling.** Maxing out per-turn parallel tool calls
  instead of switching to code-driven spawning.
- **Unbounded recursion.** No depth cap, causing runaway trees and cost.
- **Leaky context (independent work).** Letting children see the parent or
  siblings reintroduces contamination and breaks deterministic aggregation.
- **Forcing isolation on collaborative work.** The mirror error: starving a
  debate/critic loop of the cross-talk that is the whole point.
- **Fixed-round debate on every input.** Running the same debate procedure
  unconditionally wastes compute on inputs the agents already agree on and
  amplifies conformity on hard ones. Route to debate only on low-agreement inputs,
  stop on convergence, and aggregate with outlier-awareness.
- **Prompt-only diversity in debate.** Role prompts over one backbone share its
  blind spots, so debate reinforces correlated errors instead of cancelling them;
  use genuinely different model families when diversity is the goal.
- **Treating a coupled, stateful search as stateless propose-implement.** When
  actions reshape the landscape and failures are diagnostic, scoring candidates
  independently and discarding failures throws away the signal that should
  constrain the next move; hold an explicit search tree instead.
- **No measurement-integrity guardian in an autonomous metric loop.** Without a
  critic empowered to reject a "win," the system optimizes confidently toward
  configurations that skipped correctness gating or were measured wrong.
- **Reward hacking / evaluator tampering.** The agent can read or modify the
  grader, test data, or authoritative result files.
- **Premature convergence.** Same-round implementers copy each other; preserve
  diversity with same-round isolation.
- **LLM-scoring a verifiable metric.** Couples evaluation to the model family;
  score with code instead.
- **Mishandling the budget signal.** Best practice is to expose the *time* budget
  the agent can usefully reason about and track *cost* externally without feeding
  it into the objective; surfacing raw cost as a reward invites gaming.
- **Surface-form joins on names.** Merges homonyms and splits aliases; join on
  stable IDs.
- **Unbounded context growth in a long run.** No compaction/eviction policy, so
  the agent's history fills the window with stale detail.
- **No operational layer.** No retries, no timeouts, no fallback routing, no
  tracing — the system works in the demo and fails in production.
- **Trusting retrieved content and tool output as instructions.** Prompt
  injection through the data path; scope permissions and filter.
- **Over-prescribing the workflow** (optimization-loop setting). Baked-in
  assumptions worsen as base models improve; engineer the environment instead.
- **Assuming a swarm improves reasoning.** For *independent* fan-out, the payoff
  is throughput/auditability/isolation, not reasoning quality. (Collaboration —
  branch 3 — genuinely can improve reasoning; do not conflate the two.)

## Reference files

- **references/collaborative-and-single-agent-patterns.md** — open first when
  deciding the architecture: the single-agent ReAct loop and when one agent is
  right; collaborative patterns (debate, voting/self-consistency, planner-critic,
  generator-verifier, blackboard/message-bus); adaptive debate as *conditional
  computation* (heterogeneous agents, pre-debate agreement routing, early-stopping,
  outlier-aware aggregation) so deliberation is spent only where independent passes
  disagree; and when collaboration improves reasoning vs when isolation wins.
- **references/recursive-harnesses.md** — open when designing the parent/child
  spawning machinery: recursive unit, code vs tool-call spawning, auto-selection,
  context isolation, child loop, depth bounds, prompt-caching.
- **references/decomposition-and-spawning-patterns.md** — open when deciding
  *how to split* a job and wire children together: decision tree, fan-out shapes,
  tree search as a planning layer over large stateful action spaces (dynamic
  search tree, gain/cost/risk scoring, driver/specialist/critic by cognitive
  function), mechanism selection by scale, isolation flavors, right-sizing,
  resource knobs.
- **references/environment-engineering.md** — open when building an autonomous or
  iterative loop and you need the PREPARE / PROPOSE (fan-in) / IMPLEMENT (fan-out)
  structure and the four environment dimensions (permissions, artifacts, budgets,
  human oversight).
- **references/agent-memory.md** — open when a long-running or multi-turn agent
  needs memory: working vs episodic vs semantic memory, history compaction,
  context-window eviction, persist-vs-discard, relation to the knowledge graph.
- **references/reliability-and-operations.md** — open when hardening for
  production: retries/backoff, timeouts, idempotency, partial-failure recovery,
  circuit breakers, fallback model routing and cost tiering, tracing, caching,
  tool/MCP interface design, and where file-based resumability sits relative to
  off-the-shelf orchestration engines.
- **references/input-safety-and-guardrails.md** — open when an agent ingests
  untrusted content or calls tools on the strength of it: prompt injection via
  retrieved content and tool output, output filtering, PII handling, and
  tool-permission scoping against malicious data.
- **references/knowledge-orchestration.md** — open when the bottleneck is the
  knowledge layer: stable-ID joins, semantic anchors, typed relations,
  multi-source retrieval, deterministic operators, auditable swarms.
- **references/evaluation-and-auditability.md** — open when measuring whether an
  orchestration change actually helped, or making runs inspectable: fixed-backbone
  method, bootstrap CIs, secure evaluators, typed contracts, LLM-as-judge.
- **references/failure-modes-and-mitigations.md** — open when a system
  underperforms or before launch: a diagnostic catalog grouped by failure class
  with detection + mitigation, plus a symptom-to-fix triage table.
