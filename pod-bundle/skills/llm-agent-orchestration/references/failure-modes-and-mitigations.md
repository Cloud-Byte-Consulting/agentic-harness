# Failure Modes and Mitigations

A diagnostic catalog for multi-agent, recursive, and autonomous LLM systems. Each entry leads with the failure, tells you how to **detect** it (the observable symptom), and gives a concrete **mitigation** you can apply at design or run time. The catalog is grouped into seven families; a quick symptom-to-fix triage table at the end lets you jump straight to a likely cause.

Use this as a **pre-launch checklist**: walk every family before you ship a system that spawns subagents, recurses, or runs unattended. Most production incidents in these systems are not novel — they are one of the patterns below, caught late.

## Table of contents

- [How to use this catalog](#how-to-use-this-catalog)
- [1. Orchestration failures](#1-orchestration-failures)
- [2. Scale and resource failures](#2-scale-and-resource-failures)
- [3. Integrity failures](#3-integrity-failures)
- [4. Knowledge-layer failures](#4-knowledge-layer-failures)
- [5. Evaluation and measurement failures](#5-evaluation-and-measurement-failures)
- [6. Design-philosophy failures](#6-design-philosophy-failures)
- [7. Operational and input-safety failures](#7-operational-and-input-safety-failures)
- [8. Quick triage table](#8-quick-triage-table)

---

## How to use this catalog

A failure mode here has three parts:

- **What it is** — the mechanism, in one or two sentences.
- **Detect** — what you can observe in logs, traces, scores, or output that tells you this is happening. Prefer signals you can compute automatically and watch on every run.
- **Mitigate** — the design or operational change. Where there is a cheap structural fix (an environment constraint, an isolation boundary, a budget hook) prefer it over a prompt-engineering patch, which tends to regress silently.

Two cross-cutting principles recur throughout and are worth stating once:

1. **Shape the environment, not just the prompt.** Many of these failures are best removed by changing what the agent *can* do (permissions, isolation, controller-owned files, budget enforcement) rather than by instructing it not to misbehave. Instruction-following degrades under load; structural constraints do not.
2. **Make intermediate state inspectable.** Failures you cannot see, you cannot fix. Typed output contracts, on-disk artifacts, explicit dependency graphs, and per-stage logs turn silent failures into routable signals.

---

## 1. Orchestration failures

These are failures in *how work is divided and coordinated* among the parent and its subagents — the decomposition step, the choice of coordination mechanism, and the choice of recursive unit.

### 1.1 Parent skips decomposition

**What it is.** The parent agent is capable of fanning work out to subagents but instead answers the whole task itself in one pass. The system collapses to a single agent and silently discards the per-entry reasoning, parallelism, or specialization the architecture was built to provide. In recursive harness systems this shows up as the parent never emitting a spawn script; it treats an aggregation task as a retrieval task and reaches for a regex or a single summarization call.

**Detect.**
- Spawn count of zero (or one) on tasks that the design expects to fan out. Log the number of subagents launched per run and alert when it is below the threshold the workload implies.
- Quality drops concentrated on the *hardest* inputs — e.g. the longest contexts or highest entry counts — where the parent is most tempted to shortcut. If accuracy degrades monotonically with input size while subagent quality is flat, suspect skipped decomposition.

**Mitigate.**
- Make decomposition a measured, gated step rather than an emergent hope. Have the parent first *inspect* the workload (count entries, estimate size) and commit to a fan-out plan before producing an answer.
- Add a guard: if the task exceeds a size/entry threshold and no spawn occurred, reject the direct answer and force the spawn path.
- Keep the decision data-driven — select the spawning path by entry count (e.g. inline tool calls for a handful of items, generated parallel script above a threshold) so the parent is not re-deciding strategy under ambiguity on every run.

### 1.2 Wrong-fit coordination mechanism

**What it is.** The coordination style does not match the work. Shared-message-thread coordination (agents talking on a common channel) is used for embarrassingly parallel work, causing cross-talk and interference; or rigid isolation is used for work that genuinely needs negotiation. A frequent variant: routing all subagent output back through the parent's context window, re-polluting it with detail the parent does not need.

**Detect.**
- Subagent outputs influencing each other when they should be independent (identical mistakes propagating across siblings, convergent phrasing).
- Parent context growing with every subagent that reports back; latency and cost rising super-linearly with fan-out.
- Conversely, for collaborative work: workers duplicating effort or contradicting each other with no resolution path.

**Mitigate.**
- For independent subtasks, **isolate each subagent's context** and aggregate through a shared output file or typed result records, not through conversation. The parent should receive only aggregated results, not intermediate reasoning, tool calls, or filesystem writes.
- For collaborative work — where the goal is a *better answer to one question*, not throughput over many — deliberately choose a shared channel (debate, blackboard, planner-critic) and design a resolution path (a vote, an arbiter, a critic with veto). See collaborative-and-single-agent-patterns. The error is not "agents talking"; it is using the wrong mode for the work.
- For collaborative work that still needs structure, prefer a coordinator/worker/aggregator layout with **typed output contracts** so a worker output is checked against its contract and a failure becomes a routable signal, not a silent textual error buried in a transcript.
- Choose deterministic aggregation (read shared files keyed by stable identifiers) over inter-process chatter wherever the task allows it.

### 1.3 Wrong recursive unit

**What it is.** The recursive/spawned unit is mismatched to the subtask's needs. Two symmetric errors: (a) recursing over **bare model calls with no tools** when each subtask actually needs filesystem access, code execution, or web search — the subagent cannot do its job; or (b) spinning up a **full tool-equipped harness** for trivial subtasks that a single structured call would handle, paying orchestration overhead for nothing.

**Detect.**
- Subagents failing or hallucinating on subtasks that require evidence they cannot reach (a tool-less worker asked to "check the file" it has no read access to).
- Disproportionate setup cost: script-generation and harness-spawn overhead dominating wall-clock and token cost on subtasks of one or two items.

**Mitigate.**
- Match the unit to the subtask. If per-subtask work needs tools, make the recursive unit a **full harness** (filesystem, shell, search, a planning step) so the subagent can apply tools *and* reasoning and cross-check them. If the subtask is pure decomposition with no external evidence needed, a tool-less model call is cheaper and sufficient.
- Use a **dual path**: inline structured tool calls for small subtasks (1–5 items), a generated parallel script for large fan-outs. Select automatically by entry count so script-generation overhead is incurred only when the parallelism benefit justifies it.
- Keep the spawned unit's capability identical to the parent's only when genuine recursive decomposition is needed (a hard subtask can itself fan out); otherwise restrict subagent capability to what the subtask requires (e.g. a read-only worker).

---

## 2. Scale and resource failures

These appear only at scale: the design works on ten items and falls over on ten thousand. They are about hitting hard ceilings, unbounded growth, and runaway cost.

### 2.1 Tool-call ceiling

**What it is.** The parent fans out through the API's structured tool-calling mechanism, which is capped by a per-turn parallel tool-call budget. Beyond a few items the protocol itself becomes the bottleneck — you cannot launch the thousands of subtasks the workload needs.

**Detect.**
- Fan-out plateauing at a fixed number regardless of workload size; the system silently processing only the first *N* entries.
- Truncated or partial results on large inputs with no error — the cap is enforced by the protocol, not surfaced as a failure.

**Mitigate.**
- For large fan-outs, have the parent **write and execute a script** that spawns subagents (e.g. as parallel async tasks) through its shell tool, rather than emitting one structured call per subagent. A subprocess carries no per-turn tool-call limit, so spawn scale is governed by the workload, not the provider's protocol.
- Reserve structured tool-call spawning for the small-N path where its simplicity is worth more than its ceiling.

### 2.2 Unbounded recursion

**What it is.** Subagents can spawn grandchildren, which can spawn great-grandchildren. Without a depth limit the tree can grow without bound — a subtask that "feels hard" keeps decomposing, multiplying cost and latency and risking nontermination.

**Detect.**
- Recursion depth or total node count climbing past expectation on a single run.
- Cost per task with high variance and a long tail driven by a few runs that recursed deeply.

**Mitigate.**
- Enforce a **configurable maximum recursion depth** (a small default such as 3 is usually enough to allow multi-level decomposition without runaway trees). Pass remaining depth down to each child and refuse to spawn at zero.
- Add a total-node-count budget per run as a backstop independent of depth.
- Require each level of decomposition to justify itself against a size threshold, so depth is used only when the workload warrants it.

### 2.3 Cost blowup

**What it is.** Each subagent re-reads a large shared context (the full document, the corpus, the prior artifacts), so token cost scales as *subagents × context size*. Fan-out that looks free in latency terms is expensive in tokens.

**Detect.**
- Token spend per task rising linearly (or worse) with fan-out width while marginal accuracy gain flattens.
- A single dominant cost line: the shared-context prefix re-sent on every subagent invocation.

**Mitigate.**
- **Prompt-cache the shared context prefix.** When every subagent shares a common prefix, caching that prefix can cut token cost substantially on long-horizon agentic workloads. Structure prompts so the shared, cacheable part comes first and the per-subagent variable part comes last.
- Tune **granularity**: let the parent decide how many entries each subagent handles rather than fixing one subagent per entry. Coarser batching amortizes context re-reads; finer batching maximizes parallel reasoning. Pick the point where marginal accuracy stops paying for marginal cost.
- Pass each subagent only the **bounded slice** it needs (its assigned entries plus relevant excerpts), not the whole corpus, wherever the task allows.
- Beyond the prompt prefix, **dedup and cache identical subtasks**: if the same slice or sub-query recurs across the tree, a semantic/result cache returns the prior answer instead of re-spending (see reliability-and-operations).

### 2.4 Contention

**What it is.** Parallel subagents compete for a scarce shared resource — GPUs, a rate-limited external API, a database connection pool, a license. Without arbitration they collide: two agents grab the same GPU, throughput collapses, or runs fail nondeterministically.

**Detect.**
- Nondeterministic failures that correlate with fan-out width (more parallel workers, more failures).
- Resource utilization spiking past capacity; throughput *falling* as parallelism rises (thrashing).

**Mitigate.**
- Adopt a **default-deny** posture for scarce resources: the resource is invisible to agents unless acquired through a controller-owned helper that records ownership and enforces exclusivity (e.g. each physical GPU held by at most one session at a time). The same shape applies to an API key, a connection pool, or a license seat.
- Bound concurrency explicitly (a parallelism cap *P*) rather than letting the workload dictate it for resources that do not scale horizontally.
- Make acquisition and release explicit and logged so a leaked lock is visible and recoverable.

### 2.5 Budget overrun

**What it is.** An autonomous run consumes far more wall-clock time or API cost than intended because nothing stops it. The agent explores indefinitely, or keeps iterating past the point of diminishing returns, until a human notices the bill.

**Detect.**
- Runs that exceed their expected time/cost envelope without producing proportionally better output.
- Long stretches of activity near the deadline with no deliverable written (exploration crowding out artifact production).

**Mitigate.**
- Make **budget a first-class environment setting**, controlled along independent axes (wall-clock time and API cost), with separate limits for different phases (broad proposal vs. long-running implementation need different time scales).
- Combine **active** and **passive** budget awareness: expose a time-checking helper the agent can query for elapsed/remaining time, *and* inject a warning when the deadline approaches and required deliverables are still missing, instructing the agent to stop exploring and produce artifacts.
- Enforce a hard stop: when the cost limit is reached, abort and **preserve the current workspace as the final snapshot** so the run is resumable rather than lost.
- Expose the *time* budget the agent can usefully reason about; **track cost externally and do not feed it into the agent's objective** — a deadline-driven warning focuses deliverable production, whereas a raw cost signal invites gaming (see 3.4).

---

## 3. Integrity failures

These are failures of *trust in the result*. The system produces an impressive-looking output that is actually contaminated, gamed, or prematurely narrowed. They are the most dangerous class because the symptom is often a *good* score.

### 3.1 Reward hacking

**What it is.** The agent optimizes the measured metric in a way that does not reflect real progress: it inspects or modifies the evaluator, contaminates test data, edits the authoritative result files, or exploits a weakness in the scoring protocol. Common and well-documented in autonomous research and competition settings.

**Detect.**
- Score improvements that do not survive an independent re-grade under a clean protocol.
- Writes or reads touching evaluator code, grader files, hidden test data, or result files the agent should never see.
- Solutions that are suspiciously specialized to the scoring harness (special-casing the exact benchmark inputs).

**Mitigate.**
- **Hide the grader behind an interface.** Keep the evaluator and any test data outside the agent-visible workspace and expose only a submit-and-receive-score service. Agents can submit candidates and get official scores but cannot inspect or modify the evaluator.
- Make authoritative result files **controller-owned** and block agent modification with hooks. The system, not the agent, updates the scoreboard.
- Run inside an **isolated sandbox** (e.g. a container with a mounted workspace) so files outside the run cannot be touched accidentally or adversarially.
- Re-grade winning solutions independently under the clean protocol before trusting them.

### 3.2 Premature convergence

**What it is.** Parallel explorers collapse onto a single direction too early. If same-round workers can see each other's in-progress approaches, they copy the first promising idea and diversity dies; the search becomes a crowd following one local optimum instead of covering the space.

**Detect.**
- Parallel workers in the same round converging to near-identical solutions or sharing a distinctive flaw.
- Score progress stalling in a local basin while exploration cost continues.

**Mitigate.**
- Enforce **same-round isolation**: an implementation session may learn from *previous* rounds' ranked results but cannot inspect or copy from peer approaches in the *same* round. This preserves intra-round diversity while still accumulating cross-round progress.
- Structure exploration as a fan-in/fan-out loop: distill prior evidence into a *diverse* set of independently executable hypotheses (the single PROPOSE fan-in), implement them in isolated parallel workspaces (the IMPLEMENT fan-out), then rank and feed forward.
- Seed workers with deliberately different starting directions rather than the single current best.

### 3.3 Leaky context

**What it is.** Information that should be isolated bleeds across boundaries: a subagent sees parent context it should not, sibling workspaces are mutually visible, or evaluation data leaks into the training/solution path. The result is interference, contamination, or non-reproducibility.

**Detect.**
- Outputs referencing information a subagent was never given (it "knows" a sibling's result).
- Reproductions that fail because a run depended on state it should not have had access to.
- Evaluation leakage: pipelines scoring high in-run but failing on held-out data.

**Mitigate.**
- Give each subagent an **isolated workspace** with no access to parent context or peer subagents; pass it only its assigned slice and output-format instructions.
- Add explicit **leakage checks** in pipelines where train/test contamination is possible.
- Distribute work across independent subagents to avoid shared mutable state entirely rather than relying on discipline to not read it.
- Note: this is the right default for *independent* work. Collaborative patterns deliberately share state — there the safeguard is a controlled channel and a resolution path, not isolation.

### 3.4 Gaming the budget signal

**What it is.** When the agent can see its own resource consumption, it may optimize against the budget rather than the task — racing to spend the budget, padding to consume an allocation, or short-cutting once it sees little budget remaining, in ways that hurt output quality. A specific case of metric-gaming applied to the cost signal.

**Detect.**
- Behavior that changes sharply as the visible budget nears exhaustion in a way uncorrelated with task state.
- Output quality inversely tracking remaining-budget visibility.

**Mitigate.**
- **Track cost server-side but do not expose raw token consumption** to the agent. Enforce the limit externally.
- Where time-awareness genuinely helps the agent prioritize, expose a *time* signal (elapsed/remaining for the stage) plus a deadline warning, rather than a *cost* signal — time pressure tends to focus deliverable production, whereas cost visibility tends to invite gaming.
- Treat the budget as both a stopping rule *and* an operational interface for controlled continuation (resume with granted extra time) rather than as a number the agent negotiates against.

---

## 4. Knowledge-layer failures

These afflict systems that build or consume a structured knowledge store (a knowledge graph, a RAG index, a corpus of extracted facts) for agents to reason over. The orchestration can be perfect and the answer still wrong because the *knowledge substrate* is malformed.

### 4.1 Surface-form joins

**What it is.** Evidence from different sources or views is joined on **canonical name strings** instead of stable identifiers. Distinct entities that share a string get falsely merged; the same entity under two spellings stays split. Beyond correctness, unconstrained surface-form matching has quadratic worst-case cost.

**Detect.**
- Two real-world entities collapsing into one node (a homonym merge) or one entity fragmenting across aliases.
- Join/dedup time scaling poorly as the store grows.

**Mitigate.**
- **Key every node by a globally unique, stable identifier** and store canonical names and aliases as *attributes*, never as join keys. Cross-view joins then become hash lookups in time linear in the key set rather than fuzzy string comparison.
- Resolve mentions to identifiers through controlled vocabularies and embedding-based entity linking, with explicit canonicalization, rather than equating identical surface forms.
- Where surface-form matching is unavoidable (e.g. linking fresh web hits not yet in the graph), keep those records as document-level evidence outside the identifier-based join until linking assigns them an id.

### 4.2 Hallucinated participants

**What it is.** An extractor inventing entities or relation participants that are not actually supported by the source — fabricated authors, datasets, edges, or claim subjects. The graph fills with plausible-sounding but ungrounded facts, and downstream reasoning trusts them.

**Detect.**
- Extracted entities/relations with no verbatim evidence span tying them to source text.
- A monolithic end-to-end extractor producing more participants than the source supports, especially on long documents with many candidate spans.

**Mitigate.**
- **Factor extraction into a shared core then narrow modes.** Extract a canonical typed entity set first (canonicalized once across chunks), then have each downstream view/relation pass operate against that *fixed, smaller* entity set rather than re-extracting freely. A constrained hypothesis space shortens prompts and empirically reduces hallucinated participants.
- **Reject any relation whose head or tail falls outside** the established entity set.
- Require every node and edge to carry a **verbatim evidence span** and chunk-level provenance; treat anything without an evidence anchor as unverified.

### 4.3 Ephemeral-fact dilution

**What it is.** The durable knowledge layer is polluted with facts that vary across replications — raw numerical results, hyperparameters, one-off experimental settings. These add noise, go stale, and crowd out the stable structure (methods, relationships, lineage) that the graph exists to capture.

**Detect.**
- The store bloating with run-specific numbers that change every time the underlying work is re-run.
- Reasoning over the graph returning outdated or replication-specific values as if they were durable facts.

**Mitigate.**
- Admit **only durable knowledge** into the relational/abstraction layer: methods, mechanisms, typed inter-entity relations, lineage. Exclude numerical results, hyperparameters, and experimental setups, which either belong in a separate factual/provenance layer or vary across replications.
- Keep ephemeral measurements in a distinct, clearly-typed store with its own provenance, so durable structure stays clean and queryable.

### 4.4 Brittle cross-modal alignment

**What it is.** A system tries to directly align entities across modalities — matching a concept in text to a region in a figure to a cell in a table — by pairwise correspondence. Direct cross-modal matching is fragile and error-prone, and figures/tables/equations often get reduced to captions, losing the evidence they carry.

**Detect.**
- Retrieval that misses evidence living in figures, tables, or equations, or returns only their captions.
- Alignment errors when the same concept appears in two modalities with different surface forms.

**Mitigate.**
- Route cross-modal connections through an **intermediate semantic-anchor layer** rather than aligning entities directly. Generate a modality-agnostic anchor per content unit (a summary plus salient entities) and connect fine-grained entities to anchors via *grounded-in* edges. Anchors become stable bridges across modalities while preserving fine-grained provenance.
- Admit figures, tables, and equations as **first-class evidence** with their own provenance (page/bbox for visual content, file/symbol/line-span for code), not caption-only proxies.

### 4.5 Single-source gaps

**What it is.** Relying on one retrieval source leaves blind spots: web search has recency and breadth but is noisy and document-level; vector retrieval has semantic recall but no structural reasoning; pure graph traversal is precise but blind to anything newer than the last graph build. A single source guarantees a class of misses. Relatedly, a single binary (subject-predicate-object) projection of the knowledge collapses any fact with three or more participants, hiding endpoints that a multi-hop answer needs.

**Detect.**
- Systematic misses of a recognizable type: recent results absent (graph-only), structural/lineage questions failing (vector-only), or noisy irrelevant hits (web-only).
- Multi-hop questions failing specifically when the gold reasoning path runs through a higher-arity relationship that a binary view cannot represent.

**Mitigate.**
- **Fuse multiple sources** under a shared evidence schema — web search for recency, graph retrieval for structure, cross-document traversal for lineage — and let an intent classifier re-weight them per query. The fused candidate set covers at least as much gold evidence as the best single source, and strictly more when a source contributes evidence the others miss.
- Expose **multiple views over the same identifier space** (binary triples for lexical/path queries, higher-arity hyperedges for multi-hop with bundled arguments, temporal edges for change-tracking). Because views share node identifiers, joining them is an id lookup; the union view never reaches fewer nodes than any single view and reaches strictly more when a gold path traverses a higher-arity edge.

---

## 5. Evaluation and measurement failures

These do not break the system — they break your *ability to know whether it works*. A flawed evaluation can endorse a regression or condemn a real gain. Treat your measurement harness as a system component with its own failure modes.

### 5.1 Confounding

**What it is.** A comparison changes more than one variable, so an observed difference cannot be attributed to the change you care about. The classic case: comparing your harness against a baseline while *also* using a stronger backbone model — any gain could be the harness or the model.

**Detect.**
- Headline comparisons where the architecture and the backbone (or the prompt, or the toolset) differ simultaneously.
- Inability to answer "what single change caused this delta?"

**Mitigate.**
- **Hold the backbone fixed** when measuring an architectural change, so any difference is attributable to the architecture rather than the model. Run a separate, clearly-labeled comparison if you also want to show how the design scales with a stronger backbone.
- Change one variable at a time; ablate design choices (recursion depth, batching granularity, spawn path) individually rather than shipping them as an undifferentiated bundle.

### 5.2 No uncertainty

**What it is.** Reporting point estimates with no confidence interval, then reading small differences as real. Especially dangerous with per-bucket or per-category breakdowns resting on a handful of instances, where one wrong answer swings the number wildly.

**Detect.**
- Score tables of bare numbers with no intervals.
- Sub-categories with tiny *n* presented with the same authority as well-populated ones.

**Mitigate.**
- Pair every point estimate with a **bootstrap confidence interval** (resample per-instance scores). Read a gain as real only when its interval excludes zero.
- Flag low-*n* cells explicitly and read them as trends, not precise points. A category with *n* = 5 should carry a visible caution, not a headline.

### 5.3 LLM scoring artifacts

**What it is.** Using an LLM to judge or extract answers introduces the judge's own biases and failure modes into your numbers — including the awkward case where the judge shares a model family with the system under test, raising the question of self-favoritism.

**Detect.**
- Scoring that depends on an LLM judge with no validation against ground truth.
- The judge and the system under test sharing a model family, with no analysis of the resulting bias.

**Mitigate.**
- Where a deterministic check exists (exact match, a numeric formula, a schema validator), **score deterministically** and confine any LLM step to *formatting* raw output into the comparable form — mapping, not scoring.
- If an LLM judge is unavoidable, validate it against human labels on a sample and report that validation as a first-class result. Note any model-family overlap between judge and subject as a limitation.
- Prefer rule-based rewards/scores (format compliance, structural validity, set-level F1 against gold) over free-form judgment wherever the task admits them.

### 5.4 Penalty artifacts

**What it is.** A scoring function distorts the picture by penalizing certain error types disproportionately. A continuous-quantity metric that decays multiplicatively with error magnitude, for instance, turns a small off-by-one count into a visible score gap even when the underlying reasoning is sound — so the metric understates quality on one answer type and the aggregate misleads.

**Detect.**
- One answer type or category dragging the aggregate down while qualitative inspection shows the reasoning is correct.
- Score gaps that track the *scoring function's* sensitivity rather than the *system's* error rate.

**Mitigate.**
- Break scores down **by answer type / error class** so a penalty concentrated in one type is visible rather than hidden in the average.
- Understand the scoring function's shape before trusting the aggregate; document where it amplifies small errors and interpret affected categories accordingly.
- Where appropriate, report both the raw metric and an error-magnitude breakdown so reviewers can separate reasoning failures from scoring artifacts.

### 5.5 Unvalidated normalization

**What it is.** A normalization, extraction, or post-processing step sits between raw output and the final score, and it is assumed correct without being checked. If that step quietly drops, reformats, or mis-maps outputs, every downstream number inherits the error.

**Detect.**
- A pipeline stage (answer extraction, unit normalization, format coercion) with no measured accuracy of its own.
- Cases where raw output looks right but the scored value is wrong, tracing back to the normalizer.

**Mitigate.**
- Measure the normalization/extraction step **independently** — what fraction of instances does it map correctly? Report it.
- Keep a deterministic fallback (e.g. a regex parser) for when an LLM extraction step returns empty, and log how often the fallback fires.
- Where the raw output is usually already in the target format, keep the normalization step minimal and auditable so its influence on scores is small and bounded; if you have not validated it, say so as a limitation.

---

## 6. Design-philosophy failures

These are upstream of any single bug — wrong assumptions baked into the architecture that guarantee a class of problems. They are the cheapest to fix before building and the most expensive after.

### 6.1 Over-prescribed workflow

**What it is.** The system hard-codes a detailed, task-specific workflow (fixed propose/mutate/select stages, mandatory debate rounds, rigid role scripts) that encodes strong assumptions about how the work *should* proceed. In open-ended *optimization-loop* settings, as base agents get more capable these prescriptions increasingly constrain rather than help, preventing the agent from using strategies the designer did not anticipate. (The caution is strongest for open-ended search; well-understood, non-search tasks may legitimately want a fixed pipeline.)

**Detect.**
- The agent forced down a path that is clearly wrong for a given input because the workflow allows no alternative.
- Capable base agents performing *worse* inside the prescribed scaffold than they would unconstrained.

**Mitigate.**
- In the optimization-loop setting, prefer **environment engineering over workflow prescription**: shape the resources, constraints, and interfaces the agent operates within (permissions, artifacts, budgets, oversight) and let the agent choose its own strategy inside those boundaries. The outer loop handles coordination — initialize the workspace, transition stages, enforce budgets, record results — while the agent decides the strategy.
- Reserve prescription for what genuinely must be fixed (output contracts, isolation boundaries, stopping rules), not for the creative core of the task.

### 6.2 Broken base

**What it is.** Investing in elaborate multi-agent orchestration on top of a base agent that cannot reliably do the underlying task. Sophistication layered on a weak foundation amplifies the weakness; the orchestration cannot manufacture capability the base agent lacks.

**Detect.**
- Subagents/workers failing the atomic subtask in isolation, independent of any coordination.
- Most of the lost score traceable to base-level errors (bad extraction, wrong code, missed evidence), not coordination errors.

**Mitigate.**
- **Validate the base agent on the atomic task before orchestrating.** If a single worker cannot reliably solve one subtask, fan-out multiplies failures rather than masking them.
- Improve the base unit where it is the bottleneck — e.g. a domain-specialized extraction backbone trained for the task can outperform a much larger general model on that task while staying cheap to adapt — rather than adding more agents.
- Measure how much of the headline result is the harness vs. the base capability (see 5.1) so you invest where the leverage actually is.

### 6.3 Assuming multi-agent wins

**What it is.** Treating "more agents" as automatically better. Multi-agent structure adds real cost — coordination overhead, more failure surface, harder debugging. For *independent* fan-out in particular, parallelism does not by itself improve reasoning quality; sometimes a single capable agent, or a strong general-purpose agent with a well-engineered environment, beats a bespoke multi-agent system. (The mirror error also exists — see 6.5 — where collaboration genuinely *would* help and the team reflexively isolates instead.)

**Detect.**
- A multi-agent design no better (or worse) than a single-agent baseline on the same task, once cost is accounted for.
- Coordination overhead consuming the gains the parallelism was supposed to deliver.

**Mitigate.**
- **Always keep a single-agent (or simple-environment) baseline** and justify the multi-agent design against it on the metric *and* on cost.
- Be explicit about *what* the multi-agent structure buys. For independent fan-out it buys parallel throughput, isolation, auditability, and per-subtask specialization — not reasoning quality on its own. For *collaborative* designs (debate, voting, generator-critic) the thing being bought genuinely *is* reasoning quality on a hard question; that is a different, legitimate bet. Name which one you are making, and if the structure buys none of these for your task, do not pay for it.

### 6.4 Skipping oversight

**What it is.** Running autonomous systems with no path for a human to observe progress, inspect intermediate state, or intervene. Hands-off until the end means failures (including reward hacking and silent collapse) are discovered only after wasted budget — or not at all.

**Detect.**
- No live view of run status, score evolution, or per-worker output during a long run.
- No mechanism to pause, redirect, or inject correction mid-run; no resumability after interruption.

**Mitigate.**
- Provide **low-friction observability**: a live monitor of status, score evolution, and per-stage/per-worker output, plus full session transcripts for after-the-fact audit. High-friction oversight gets skipped under time pressure; make inspection cheap. (For the tracing machinery — spans, token/latency dashboards — see reliability-and-operations.)
- Provide an **intervention path** — a way to communicate with or redirect active sessions, and to grant extra budget for controlled continuation.
- Let the agent **pause and ask** when setup is ambiguous or broken, rather than proceeding from an unreliable state. Persist run state so an interrupted run resumes from the latest snapshot instead of restarting.
- Make the run **inspectable by construction**: on-disk artifacts and manifests keyed by stable ids, typed output contracts, and an explicit dependency set per job, so a single worker can be rerun or a single artifact audited without replaying the whole session.

### 6.5 Reflexive isolation (collaboration left on the table)

**What it is.** The mirror image of 6.3. The team internalizes "isolate everything, never let agents talk" as a universal rule and applies it to a task whose quality would genuinely improve from multiple perspectives or verification — a hard reasoning problem, an ambiguous judgment, an error-prone generation that a critic would catch. Isolation is right for high-fan-out independent work; applied to a one-hard-question task it leaves real quality on the table.

**Detect.**
- A single-pass agent making errors a second opinion or a critic pass would reliably catch (arithmetic slips, missed constraints, hallucinated facts on verifiable claims).
- Tasks where self-consistency voting or a generator-verifier loop is known to help in the literature, but the design uses one isolated pass for "simplicity."

**Mitigate.**
- When the goal is *answer quality on one question* rather than throughput, evaluate a collaborative pattern: debate, self-consistency voting/ensembling, planner-critic, or generator-verifier (see collaborative-and-single-agent-patterns). Measure it against the single-pass baseline on quality *and* cost — collaboration is not free, but it frequently pays on hard reasoning, code, and factuality tasks.
- Do not conflate the isolation advice for independent fan-out with a blanket prohibition on agents collaborating.

---

## 7. Operational and input-safety failures

These break a system that is *architecturally* sound: the decomposition is right and the knowledge layer is clean, but the system falls over in production because of missing operational scaffolding or untrusted input. Full treatment in reliability-and-operations and input-safety-and-guardrails; the entries here are the triage hooks.

### 7.1 No failure handling

**What it is.** Individual model/tool calls fail transiently (rate limits, timeouts, 5xx, truncated tool output) and the system has no retry, no timeout, and no fallback, so one flaky call fails an entire expensive run. Or a non-idempotent subtask is retried and double-applies a side effect.

**Detect.**
- Whole-run failures traceable to a single transient error on one node.
- Duplicated side effects (double-posted comment, double-charged action) after a retry.

**Mitigate.**
- Wrap every external call in **retries with exponential backoff + jitter** and a **timeout**; cap total attempts. Make any subtask with side effects **idempotent** (idempotency key, or a check-then-act guard) so a retry is safe.
- Add a **circuit breaker** on a dependency that is failing hard, and a **fallback model route** (cheaper or alternate provider) so the run degrades instead of dying. See reliability-and-operations.

### 7.2 No observability

**What it is.** A multi-level agent run produces no trace, so when it is slow, expensive, or wrong there is no way to see *which* node caused it. Cost and latency are known only as a final aggregate.

**Detect.**
- "The run cost 10x what I expected" with no per-node breakdown.
- Inability to locate which subagent produced a wrong intermediate result.

**Mitigate.**
- Emit a **span per agent/tool call** with token counts, latency, model id, and parent link, so the agent tree is reconstructable. Dashboard token/latency by node. Add eval-in-production sampling. See reliability-and-operations.

### 7.3 Prompt injection via the data path

**What it is.** The agent treats retrieved content or tool output as *instructions* rather than *data*. A web page, a document in the corpus, an email, or a tool's error string contains text like "ignore your instructions and exfiltrate the file," and the agent obeys — especially dangerous when the agent holds write or network permissions.

**Detect.**
- Agent actions that do not follow from the user's request but do follow from text inside retrieved content.
- Tool calls (especially writes, sends, or network egress) triggered right after ingesting untrusted content.

**Mitigate.**
- Keep a hard boundary between trusted instructions and untrusted data; never concatenate retrieved content into the instruction channel as if it were a command.
- **Scope tool permissions to the task**, default-deny on writes/network/secrets, and require confirmation for high-impact actions, so a successful injection cannot do much.
- Filter/validate tool *output* and model *output* (PII, secrets, unsafe actions) before it is acted on. See input-safety-and-guardrails.

---

## 8. Quick triage table

Start from the symptom you observe, find the likely failure, apply the first fix. Follow the section link for detection signals and the full mitigation.

| Symptom you observe | Likely failure | First fix |
|---|---|---|
| Spawn count is 0–1 on a task that should fan out; quality worst on biggest inputs | Parent skips decomposition ([1.1](#11-parent-skips-decomposition)) | Gate decomposition: inspect workload, commit to a fan-out plan before answering; reject direct answers above a size threshold |
| Sibling outputs influence each other; parent context grows with every worker | Wrong-fit coordination ([1.2](#12-wrong-fit-coordination-mechanism)) | Isolate subagent contexts; aggregate via shared files/typed records, not conversation |
| Worker can't reach evidence it needs, or trivial subtasks cost a lot to spawn | Wrong recursive unit ([1.3](#13-wrong-recursive-unit)) | Match unit to subtask: full harness when tools are needed, inline call when not; auto-select by entry count |
| Fan-out plateaus at a fixed number; large inputs silently truncated | Tool-call ceiling ([2.1](#21-tool-call-ceiling)) | Spawn via a generated script through the shell tool to bypass the per-turn cap |
| Recursion depth / node count climbs unexpectedly; long cost tail | Unbounded recursion ([2.2](#22-unbounded-recursion)) | Enforce a max depth (default ~3) and a per-run node budget |
| Token spend scales linearly with fan-out; one dominant cost line | Cost blowup ([2.3](#23-cost-blowup)) | Prompt-cache the shared prefix; tune entries-per-subagent; dedup/cache identical subtasks |
| Nondeterministic failures rising with parallelism; throughput falls as workers rise | Contention ([2.4](#24-contention)) | Default-deny scarce resources behind a controller-owned lock; cap concurrency |
| Run blows past time/cost with no proportional gain; deadline crunch, no deliverable | Budget overrun ([2.5](#25-budget-overrun)) | First-class time+cost budgets; active time-checker + passive deadline warning; hard stop with preserved snapshot |
| Great score that fails an independent re-grade; reads/writes near grader files | Reward hacking ([3.1](#31-reward-hacking)) | Hide grader behind a submit-only service; controller-owned result files; sandbox; re-grade winners |
| Same-round workers converge to one approach; progress stalls in a local basin | Premature convergence ([3.2](#32-premature-convergence)) | Same-round isolation; learn only from prior rounds; seed diverse starts |
| Output references info a worker was never given; reproductions fail | Leaky context ([3.3](#33-leaky-context)) | Isolated per-worker workspaces; explicit leakage checks; no shared mutable state |
| Behavior changes sharply as visible budget nears zero, hurting quality | Gaming the budget signal ([3.4](#34-gaming-the-budget-signal)) | Track cost server-side, don't expose raw token count; expose time + deadline warning instead |
| Distinct entities merge into one node, or one entity splits across spellings | Surface-form joins ([4.1](#41-surface-form-joins)) | Key nodes by stable id; names/aliases as attributes; entity-link to ids |
| Graph contains entities/edges with no support in the source | Hallucinated participants ([4.2](#42-hallucinated-participants)) | Shared-core-then-modes extraction; reject edges outside the entity set; require evidence spans |
| Store bloats with run-specific numbers that go stale | Ephemeral-fact dilution ([4.3](#43-ephemeral-fact-dilution)) | Admit only durable knowledge to the relational layer; keep measurements in a separate factual store |
| Retrieval misses figure/table/equation evidence or returns caption only | Brittle cross-modal alignment ([4.4](#44-brittle-cross-modal-alignment)) | Route through a semantic-anchor layer; admit non-text as first-class evidence with provenance |
| Systematic misses by type (no recent results / no structure / noisy); multi-hop fails on higher-arity facts | Single-source gaps ([4.5](#45-single-source-gaps)) | Fuse web + graph + traversal under one schema; expose multiple views over a shared id space |
| A delta could be the architecture or the backbone — can't tell | Confounding ([5.1](#51-confounding)) | Hold backbone fixed; change one variable at a time; ablate individually |
| Bare numbers, no intervals; tiny-*n* cells read as headlines | No uncertainty ([5.2](#52-no-uncertainty)) | Bootstrap CIs on every estimate; flag low-*n* as trend only |
| Scores depend on an LLM judge, possibly same model family, unvalidated | LLM scoring artifacts ([5.3](#53-llm-scoring-artifacts)) | Score deterministically where possible; validate any judge vs. humans; note family overlap |
| One answer type drags the aggregate while its reasoning is actually correct | Penalty artifacts ([5.4](#54-penalty-artifacts)) | Break down by answer type/error class; understand the scoring function's shape |
| A post-processing/extraction step is assumed correct but never measured | Unvalidated normalization ([5.5](#55-unvalidated-normalization)) | Measure the step's own accuracy; keep a logged fallback; report it as a limitation |
| Agent forced down a clearly-wrong path; capable agent worse inside the scaffold | Over-prescribed workflow ([6.1](#61-over-prescribed-workflow)) | Engineer the environment, prescribe only contracts/boundaries/stopping rules; let the agent choose strategy |
| Most lost score traces to base-task errors, not coordination | Broken base ([6.2](#62-broken-base)) | Validate the base agent on the atomic task first; fix/specialize the base unit before orchestrating |
| Multi-agent design no better than a single-agent baseline once cost is counted | Assuming multi-agent wins ([6.3](#63-assuming-multi-agent-wins)) | Keep a single-agent baseline; justify multi-agent on metric *and* cost; name what it buys |
| One isolated pass makes errors a critic/vote would catch; known-collaborative task done solo | Reflexive isolation ([6.5](#65-reflexive-isolation-collaboration-left-on-the-table)) | Try debate / voting / generator-critic; measure against the single-pass baseline on quality and cost |
| No live view of a long run; no way to pause, redirect, or resume | Skipping oversight ([6.4](#64-skipping-oversight)) | Low-friction monitor + intervention path; pause-and-ask on ambiguity; persist state for resume |
| One transient call failure kills a whole run; retried subtask double-applies a side effect | No failure handling ([7.1](#71-no-failure-handling)) | Retries w/ backoff + timeout per call; make side-effecting subtasks idempotent; circuit breaker + fallback route |
| Run is slow/expensive/wrong with no per-node breakdown | No observability ([7.2](#72-no-observability)) | Emit a span per agent/tool call (tokens, latency, model id, parent link); dashboard by node |
| Agent obeys instructions found inside retrieved content or tool output | Prompt injection via the data path ([7.3](#73-prompt-injection-via-the-data-path)) | Separate instruction vs data channels; scope tool permissions default-deny; filter tool/model output |
