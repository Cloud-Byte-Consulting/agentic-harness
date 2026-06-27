# Agent Memory and Context Management

How a long-running or multi-turn agent remembers things across a session and
across sessions, without drowning its context window in stale detail. This is the
*agent's own working state* — distinct from the durable knowledge substrate it
reasons over (knowledge-orchestration), though the two share a key discipline:
keep durable facts separate from ephemeral scratch.

## Contents

1. Three kinds of memory
2. The context window is not memory
3. Conversation-history compaction
4. Context-window management and eviction
5. When to persist vs discard
6. How this relates to (and differs from) the knowledge graph
7. A practical default policy

---

## 1. Three kinds of memory

Borrowing loosely from cognitive terms, an agent has three distinct memory needs,
and conflating them is the root of most memory bugs:

- **Working memory** — what the agent is holding *right now* to do the current
  step: the live context window. Small, volatile, expensive per token, and the
  binding constraint on every turn. Everything in working memory is competing for
  the same scarce space.
- **Episodic memory** — the record of *what happened*: past turns, prior tool
  calls and their results, earlier decisions and their outcomes, previous sessions
  with this user/task. Time-ordered and specific ("on turn 12 the build failed
  with error X"; "last session the user said they prefer metric units").
- **Semantic memory** — distilled, durable *facts and preferences* abstracted away
  from when they were learned ("this user's project uses Postgres"; "the API base
  URL is Y"). Stable, reusable across sessions, not tied to a particular episode.

The design move is to keep **working memory tiny and the other two out of the
context window until needed**, retrieving the relevant episodic/semantic items
back into working memory on demand. An agent that tries to keep all three in the
context window runs out of room and degrades.

## 2. The context window is not memory

A common mistake is treating the context window *as* the agent's memory — letting
the full transcript accumulate turn after turn. This fails three ways:

- **It fills up.** Long agentic runs blow past the window; once full, either the
  oldest content silently falls off (uncontrolled forgetting) or the call errors.
- **It gets expensive and slow.** Every turn re-processes the entire history;
  cost and latency grow with the conversation, not with the task.
- **It degrades reasoning.** A window stuffed with stale, low-relevance detail
  dilutes attention on what matters now (the "lost in the middle" effect). More
  context is not more intelligence past a point.

So memory is an *architecture*, not a side effect of appending to a transcript.
You explicitly decide what stays in working memory, what is summarized, what is
evicted to external storage, and what is retrieved back.

## 3. Conversation-history compaction

The core technique for a long session: **compact the history** so the window
holds a faithful summary plus the recent, high-relevance detail, instead of the
raw transcript.

- **Rolling summarization.** When the transcript approaches a threshold, summarize
  the older portion into a compact "story so far" and replace the raw turns with
  it. Keep the last few turns verbatim (recency matters most) and the summary for
  everything before.
- **Summarize at natural boundaries.** Compact at the end of a sub-task, a tool
  result that closes a thread, or a topic shift — not mid-reasoning — so the
  summary captures completed units rather than half-finished work.
- **Preserve the load-bearing specifics.** A good compaction keeps decisions,
  constraints, open questions, identifiers, and commitments; it drops
  pleasantries, superseded attempts, and verbose tool output already acted on. The
  failure mode is summarizing away a detail the agent later needs — bias toward
  keeping anything that looks like a fact, an id, or a decision.
- **Tool output is the usual bloat.** Large tool results (a file dump, a search
  page, a long API response) dominate token usage. Summarize or extract the
  relevant slice immediately and keep a pointer to the full result rather than the
  whole thing in-context.

Compaction is itself an LLM call and can lose information — treat the summary as
*derived data* and, where the stakes are high, keep the raw transcript in external
storage so you can re-derive a better summary or audit what was dropped.

## 4. Context-window management and eviction

Beyond summarizing prose, manage the window like a cache with an eviction policy:

- **Relevance-ranked retrieval, not chronology.** Rather than "most recent N
  turns," retrieve the items *most relevant to the current step* from episodic
  storage (by embedding similarity, recency, and importance) and load only those.
  The window holds the live task plus a working set fetched on demand.
- **Eviction policy.** When the window is pressured, evict by a blend of *recency*
  (old), *relevance* (off-topic now), and *redundancy* (already summarized or
  superseded). Pin items that must never be evicted (the system prompt, the task
  spec, hard constraints, the output contract).
- **Externalize the scratchpad.** Let the agent write intermediate state to a file
  or a structured store and read it back, rather than carrying it in-context. The
  filesystem becomes extended working memory — the same move as shared-file
  aggregation, applied to one agent's own long task.
- **Budget the window explicitly.** Reserve fractions for system/instructions,
  retrieved memory, recent turns, and the live working set, so no single source
  (usually tool output) can crowd out the rest.

## 5. When to persist vs discard

Not everything an agent produces deserves to be remembered. The litmus test mirrors
the durable-vs-ephemeral rule in knowledge-orchestration:

**Persist** (to episodic or semantic store):
- Decisions and their rationale, commitments made, constraints discovered.
- Stable facts and preferences that will matter next time (semantic memory).
- Outcomes worth learning from — what worked, what failed and why.
- Anything keyed to a user/project that future sessions should recall.

**Discard** (let it fall out of context, do not persist):
- Superseded intermediate attempts and abandoned branches.
- Verbose raw tool output once its relevant content is extracted.
- Transient reasoning scratch that led to a now-recorded conclusion.
- Pleasantries and turn-taking overhead.

**Promote deliberately, never silently.** Moving something from episodic ("the
agent observed X this session") to semantic ("X is a durable fact about this
project") should be a deliberate step with provenance, not an automatic merge — an
unverified observation promoted to a durable "fact" poisons every future session.
This is the same discipline as gating writes into the knowledge graph on a decision
rather than the agent's own judgement.

A note on privacy: persisted memory often contains user data. Decide retention and
deletion up front, and apply the PII handling in input-safety-and-guardrails to
anything written to a durable store.

## 6. How this relates to (and differs from) the knowledge graph

Agent memory and the knowledge substrate (knowledge-orchestration) are easy to
conflate because both store facts and both get retrieved into context. They are
different layers with different owners:

| | Agent memory | Knowledge substrate (graph/index) |
|---|---|---|
| **What it holds** | the agent's own session/working state, decisions, user prefs | the domain facts the agent reasons *over* (documents, entities, relations) |
| **Lifespan** | per-session (working/episodic) or per-user (semantic) | durable, shared across all agents and sessions |
| **Owner** | the agent/session | the offline construction pipeline |
| **Authority** | provisional until promoted | authoritative ground truth |
| **Trust on write** | agent writes freely to its own scratch | agent writes gated on a human/controller decision |

The shared discipline: **keep ephemeral working state out of the authoritative
durable store.** A worker's intermediate hypothesis lives in *agent memory* and is
discarded or *deliberately promoted* — never silently merged into the knowledge
graph (knowledge-orchestration §4c). Conversely, durable domain facts belong in the
substrate, retrieved into the agent's working memory when relevant, not copied
permanently into the transcript. Memory is the agent's notebook; the substrate is
the library.

## 7. A practical default policy

A reasonable starting point for a long-running agent, to tune from:

1. **Pin** the system prompt, task spec, hard constraints, and output contract —
   never evicted.
2. **Keep verbatim** the last few turns (recency window).
3. **Roll up** older turns into a running summary at sub-task boundaries; keep the
   summary, drop the raw turns from context (retain raw externally if stakes are
   high).
4. **Externalize** large tool outputs immediately — extract the relevant slice into
   context, store the full result behind a pointer.
5. **Retrieve on demand** from episodic/semantic storage by relevance to the
   current step, rather than carrying everything.
6. **Persist** decisions, stable facts, and outcomes; **discard** superseded scratch.
7. **Promote** episodic observations to semantic facts only through a deliberate,
   provenance-bearing step.

This keeps working memory small and sharp, makes long runs affordable, and keeps
the durable stores clean — the same throughline as the rest of the skill: separate
durable structure from ephemeral state, and retrieve rather than accumulate.
