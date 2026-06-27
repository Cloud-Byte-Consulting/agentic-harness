# Reliability and Operations

What separates an agent demo from an agent *system*. The architecture can be
correct — right decomposition, clean knowledge layer, honest evaluation — and the
thing still falls over in production because a rate limit hit one node, a tool
timed out, a retry double-charged a side effect, or nobody can tell which subagent
burned the budget. This file is the operational layer: making individual calls
reliable, routing around failure, seeing inside the agent tree, caching, designing
the tool interface, and locating the file-based resumability pattern relative to
off-the-shelf orchestration engines.

## Contents

1. Retries, backoff, timeouts
2. Idempotency and partial-failure recovery
3. Circuit breakers and graceful degradation
4. Fallback model routing and cost tiering
5. Observability and tracing across an agent tree
6. Caching beyond the prompt prefix
7. Designing the tool / MCP interface
8. Resumability vs off-the-shelf orchestration engines

---

## 1. Retries, backoff, timeouts

Every call out of the agent — model API, tool, subagent — can fail transiently:
rate limits (429), server errors (5xx), network blips, truncated streams. The
non-negotiable baseline is **retry with exponential backoff and jitter, under a
timeout, with a cap on attempts.**

```python
import random, time

def call_with_retry(fn, *, max_attempts=5, base=0.5, cap=30.0, timeout=60):
    for attempt in range(max_attempts):
        try:
            return fn(timeout=timeout)          # per-attempt timeout is mandatory
        except RetryableError as e:             # 429 / 5xx / network / stream cut
            if attempt == max_attempts - 1:
                raise
            sleep = min(cap, base * 2 ** attempt)
            sleep = sleep / 2 + random.uniform(0, sleep / 2)   # full jitter
            time.sleep(sleep)
        # NonRetryableError (400, auth, schema) propagates immediately — do not retry
```

- **Jitter is not optional at fan-out.** Without it, N children that all hit a
  rate limit retry in lockstep and re-collide (the thundering herd). Jittered
  backoff spreads them out.
- **Distinguish retryable from terminal.** Retrying a 400/auth/validation error
  just wastes the budget and delays the real failure. Retry transient classes
  only.
- **Always set a per-attempt timeout.** A hung tool call with no timeout pins a
  worker forever and, at fan-out, can stall the whole tree behind one slow child.
  Bound it, then treat the timeout as a retryable failure.
- **Respect `Retry-After`.** When the provider tells you how long to wait, honor
  it instead of your own backoff curve.

## 2. Idempotency and partial-failure recovery

Retries are only safe if the operation is **idempotent** — running it twice has
the same effect as once. A retried subtask that posts a comment, sends an email,
charges an account, or appends to a file *twice* is a correctness bug created by
your reliability layer.

- **Idempotency keys.** Attach a stable key to each side-effecting operation; the
  downstream system (or your own dedup layer) ignores a repeat of the same key.
- **Check-then-act.** Before performing the effect, check whether it already
  happened (does the file exist, was the row written, is the PR open). Cheap to
  add, removes whole classes of double-apply bugs.
- **Make subtask outputs content-addressed.** Writing each child's result to a path
  keyed by its slice id (not an append) means a re-run overwrites rather than
  duplicates — the shared-file aggregation pattern is idempotent by construction.

**Partial-failure recovery.** In a fan-out, some children succeed and some fail.
Do not throw away the successes. Persist each child's result as it lands, then on
recovery **re-run only the missing/failed slices** against the persisted set. The
unit of retry is the failed subtask, not the whole run — the same property that
makes on-disk manifests (evaluation-and-auditability) support selective replay.
Record per-child status (success / failed / pending) so "what still needs doing?"
is a query, not a guess.

## 3. Circuit breakers and graceful degradation

When a dependency is failing *hard* (not a transient blip), retrying makes it
worse and burns budget. A **circuit breaker** trips after a threshold of failures
and short-circuits further calls for a cool-down window, then probes with a single
trial call before closing again.

```text
CLOSED  -> calls flow; count failures
          on failure-rate > threshold -> OPEN
OPEN    -> calls fail fast (no call made) for a cooldown window
          after cooldown -> HALF-OPEN
HALF-OPEN -> allow one trial; success -> CLOSED, failure -> OPEN
```

Pair the breaker with **graceful degradation**: when a path is open/unavailable,
fall back to a degraded-but-working mode rather than failing the run —

- serve a cached or stale result,
- route to a fallback model/provider (next section),
- skip an optional enrichment step and flag the output as partial,
- reduce fan-out width to stay under a struggling dependency's capacity.

The principle: a single sick dependency should degrade the system, not kill it.

## 4. Fallback model routing and cost tiering

Treat "which model handles this call" as a routing decision, not a constant.

**Fallback routing (reliability).** If the primary model errors, is rate-limited,
or times out past its retry budget, route the same request to a fallback (an
alternate provider, region, or model). Keep the prompt provider-agnostic enough
that the fallback can serve it. This is what keeps a run alive through a provider
incident.

**Cost tiering (economics).** Heterogeneous models have very different price/
capability points. Spend the expensive model only where it earns its keep:

- **Cheap-first, escalate on failure.** Try a small/fast model; if it fails a
  verifier, returns low confidence, or the task is flagged hard, escalate to a
  stronger model. Most easy instances never touch the expensive tier.
- **Router model.** A tiny classifier (or a cheap LLM call) inspects each task and
  picks the tier up front — cheap model for simple extraction/classification,
  strong model for open-ended reasoning.
- **Role-based tiering.** In a collaborative loop, a cheap model *generates* and a
  strong model *verifies* or *judges* (or vice versa) — see
  collaborative-and-single-agent-patterns.

```text
route(task):
    tier = router.classify(task)            # cheap up-front decision
    try:
        return models[tier].run(task)
    except (RateLimited, ServerError, Timeout):
        return models[fallback_of(tier)].run(task)   # reliability fallback
    # quality fallback: if a verifier rejects the cheap result, escalate a tier
```

Log which tier and which model actually served each call (it feeds the cost
dashboard and the evaluation manifest — the backbone must be known to attribute
results).

## 5. Observability and tracing across an agent tree

A multi-level agent run is a *tree* of calls. Without tracing, "it was slow /
expensive / wrong" has no per-node answer. Instrument it like a distributed
system.

- **A span per unit of work.** Emit a span for every agent invocation and every
  tool/model call, each carrying: a stable id, parent id (so the tree
  reconstructs), role/tier, model id, token counts (in/out), latency, status,
  and a link to its output artifact. Spans turn an opaque run into a flame graph.
- **Token and latency dashboards by node.** Aggregate spend and latency *per
  node type and per depth* so you can see that, say, the depth-2 region-children
  dominate cost, or one tool is the latency tail. This is what makes the
  cost-blowup and contention failure modes detectable rather than mysterious.
- **Propagate a correlation/run id** through the whole tree so logs from
  independent workers stitch into one run.
- **Eval-in-production.** Sample a fraction of live runs and score them with your
  offline evaluator (or an LLM judge) to catch quality regressions that unit tests
  miss. Watch for drift: a rising fallback rate, growing latency tails, or a
  climbing share of contract failures are early warnings.
- **Capture inputs/outputs at boundaries** (subject to PII handling — see
  input-safety-and-guardrails) so a wrong result can be reproduced from its
  recorded inputs.

This is the machinery behind the "low-friction oversight" the design-philosophy
guidance calls for: oversight you can afford is oversight that happens.

## 6. Caching beyond the prompt prefix

Prompt-prefix caching (recurring across this skill) cuts the cost of re-reading a
shared context. There is more cache to harvest above that layer:

- **Result caching (memoization).** Cache the *output* of a deterministic step
  keyed by its inputs. Identical tool calls (the same web fetch, the same DB
  query, the same file read) across siblings or rounds return the cached result
  instead of re-executing. Define the key carefully and set a TTL for anything
  that can go stale.
- **Semantic caching.** Cache by *embedding similarity* of the request, not exact
  match, so near-duplicate queries ("summarize this paragraph" vs the same
  paragraph reworded) hit the cache. Powerful but risky: a too-loose similarity
  threshold returns a confidently wrong cached answer. Tune the threshold and
  prefer it for read-only/idempotent operations.
- **Dedup identical subtasks.** Before fanning out, collapse duplicate slices so
  the same work is not spawned N times. In a recursive tree the same sub-query can
  surface on multiple branches; a shared cache across the tree pays for itself.
- **Cache derived artifacts.** Anchor summaries, extraction outputs, and
  embeddings are derived data — key them by a content hash of the source so they
  regenerate only when the source (or the model) changes (see
  knowledge-orchestration on treating anchors as derived data).

Caching trades freshness and correctness risk for cost/latency; be explicit about
what may be served stale and never semantic-cache a side-effecting operation.

## 7. Designing the tool / MCP interface

Agents act through tools, and **tool quality bounds agent quality** as much as the
model does. A confusingly described tool or a sloppy schema produces wrong calls
no prompt can fix. Treat the tool surface as a designed API.

- **Description quality.** Each tool's description should state *what it does, when
  to use it, and when not to*, in the agent's own decision terms. Ambiguous or
  overlapping descriptions cause the agent to pick the wrong tool. Disambiguate
  near-duplicates explicitly ("use `search_docs` for internal docs; use
  `web_search` for anything public").
- **Argument schemas.** Use typed, constrained schemas (enums over free strings,
  required vs optional explicit, formats documented). A tight schema lets the
  harness reject a malformed call *before* it executes and makes the valid call
  obvious to the model.
- **Error surfaces are part of the interface.** When a tool fails, return a
  *structured, actionable* error the agent can recover from ("file not found: did
  you mean X?" beats a raw stack trace). The error message is a prompt to the
  agent — write it as one. Distinguish retryable from terminal errors in the
  response so the agent (and your retry layer) routes correctly.
- **Keep tools at the right granularity.** Too fine and the agent burns turns
  orchestrating trivial calls; too coarse and it cannot compose them. Match a tool
  to a unit of intent.
- **Make tools safe by construction.** Scope what each tool *can* do (read-only vs
  write vs network), default-deny on dangerous capability, and validate arguments
  against injection (see input-safety-and-guardrails). The deterministic typed
  operators in knowledge-orchestration are an example of well-designed tools:
  typed in, typed out, provenance attached.
- **Idempotency in the contract.** Mark which tools have side effects and make
  those idempotent (Section 2), so the retry layer can call them safely.

## 8. Resumability vs off-the-shelf orchestration engines

This skill repeatedly recommends a **file-based resumability** pattern: persist
per-stage state (session ids, status, elapsed time, remaining budget, accepted
artifacts) to disk so an interrupted run resumes from the last good state and a
human can inspect it mid-flight. That pattern is a lightweight, dependency-free way
to get durability and observability, and it is often exactly enough.

It sits on a spectrum with heavier off-the-shelf options, and it is worth knowing
where it lands without prescribing one:

- **File-based checkpointing (the pattern here).** State is plain files in a known
  layout. Maximum transparency and control, minimum infrastructure; you write the
  resume logic. Great for single-host runs and research loops.
- **DAG / graph runners** (workflow engines that execute a declared graph of steps
  with retries and caching). Give you retry/cache/visualization for free and a
  declarative structure, at the cost of expressing your agent as a graph and
  adopting the engine.
- **Durable-execution / checkpointing frameworks** (engines that persist execution
  state so a workflow survives process death and resumes mid-function). Strongest
  durability guarantees and the least bespoke recovery code, at the cost of running
  their runtime and fitting their programming model.

How to choose, in one breath: stay with **file-based** while the run is
single-host, the resume logic is simple, and transparency matters most; reach for a
**DAG or durable-execution engine** when you need cross-host orchestration,
guaranteed exactly-once survival across crashes, or you are already operating one.
None is universally correct — the file pattern is the floor that always works, and
the engines are what you graduate to when durability or scale demands outgrow it.
Whatever you pick, the invariants are the same: state is **persisted, keyed by
stable ids, and resumable**, and the unit of retry is the failed subtask, not the
whole run.
