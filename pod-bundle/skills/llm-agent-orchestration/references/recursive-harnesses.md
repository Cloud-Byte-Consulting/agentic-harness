# Recursive Agent Harnesses

How a parent agent decomposes work across many children, and the mechanics that
make that scale: what you recurse over, how you spawn, how you keep children
isolated, and how you control depth, latency, and cost. This file assumes you
have already decided to fan out *independent* work (SKILL front door, branches
1–2); for collaborative work where children should talk to each other, see
collaborative-and-single-agent-patterns instead.

## Contents

1. The recursive unit
2. Spawning: code execution vs JSON tool calls
3. Automatic path selection
4. Context isolation and shared-file aggregation
5. The child loop: plan, act, reflect
6. Capability tiers
7. Bounded recursion depth
8. Normalize-then-score
9. Cost and latency control

## 1. The recursive unit

The first design decision is *what* you recurse over. Two options:

- **Recurse over a bare model call.** The child is a single prompt/response.
  Cheap, but the child cannot use tools, run code, or plan — so it is limited to
  what one forward pass can do over its slice.
- **Recurse over a full harness.** The child is itself a complete agent: it has
  tools, can execute code, can plan and reflect, and — critically — can be given
  the *same spawn primitive the parent has*.

For tool-needing work, recursing over the harness dominates. The moment children
can spawn their own children, a single flat fan-out becomes a depth-bounded tree,
and the system can attack an input far larger than any one context window: the
root splits the job into regions, each region-child splits again, and leaf
children do the actual per-slice reasoning. The harness is the recursive unit;
the model is just the component inside it. (When the per-slice work is *pure
reasoning over text already in hand*, the bare model call is the cheaper, correct
choice — match the unit to the slice, do not reflexively reach for the full
harness.)

## 2. Spawning: code execution vs JSON tool calls

There are two ways for a parent to launch children:

- **JSON tool calls.** The parent emits parallel tool-call requests in one turn.
  Simple and natural, but bounded by the provider's cap on parallel tool calls
  per turn. Past that cap you either serialize (slow) or stall.
- **Code-driven spawning.** The parent writes a short script that launches the
  children concurrently — e.g. a loop that builds N task descriptors and awaits
  them with a concurrency primitive. Because the fan-out happens *inside code
  the parent runs*, it is not subject to the per-turn tool-call ceiling. Fan-out
  width now scales with the workload, not the API.

Framework-agnostic sketch of code-driven spawning:

```python
def parent(input, spawn, depth):
    slices = split(input)                      # decomposition
    tasks  = [make_task(s, output_spec) for s in slices]
    # concurrency is bounded by a semaphore, NOT by the tool-call cap
    results = run_concurrently(spawn, tasks, max_in_flight=K)
    return aggregate(results)                   # deterministic join
```

The key property: `spawn` is a harness primitive the child also receives, so
`parent` and `child` can be the same function at different depths.

## 3. Automatic path selection

Neither mechanism is universally best. Code-driven spawning has fixed overhead
(writing and running the script); for a handful of children, a direct tool call
is faster and simpler. So *auto-select by estimated width*:

```text
width = estimate_subtask_count(input)
if width <= SMALL_THRESHOLD:
    use direct parallel tool calls
else:
    use code-driven spawning
```

Pick `SMALL_THRESHOLD` near the provider's comfortable parallel-tool-call count.
This gives you the simplicity of tool calls for small jobs and the scaling of
code for large ones, without the operator choosing by hand.

## 4. Context isolation and shared-file aggregation

For independent work, children must not see each other or the parent's full
context. Isolation buys three things: a clean per-child context budget, no
cross-contamination between siblings, and a deterministic, reproducible
aggregation step. (This is the opposite of what a debate or critic loop wants —
there, cross-visibility is the point. Apply isolation when the children's work is
genuinely independent.)

Each child receives exactly:

- its **slice** of the input (one region, one batch of entries);
- the **relevant excerpts** it needs (shared reference material, trimmed);
- an explicit **output contract** (schema, field names, format).

It does *not* receive the parent's reasoning, sibling outputs, or the global
input. Each child writes a structured record to a **shared path** — one file per
child, keyed by child ID. After all children resolve, the parent reads the
directory and joins the records in code. The shared filesystem is the
aggregation channel; messages are not.

```text
/run/<id>/children/000.json
/run/<id>/children/001.json
...
parent: results = [load(p) for p in sorted(glob('/run/<id>/children/*.json'))]
```

A concrete non-research example: classifying 50,000 support tickets by urgency.
Each child gets a batch of tickets plus the shared rubric (cached prefix) and
writes one record per ticket to its own file; the parent reads the directory and
tallies. No child needs to see another's tickets or verdicts.

## 5. The child loop: plan, act, reflect

A harness child should run a small internal loop rather than a single shot:

1. **Plan** — restate its slice and the output contract, decide an approach.
2. **Act** — call tools / run code / (if depth permits) spawn grandchildren.
3. **Reflect** — check its result against the contract before writing it.

The reflect step is what makes children safe to aggregate blindly: a child that
self-validates against the schema rarely emits a record the parent's
deterministic join cannot parse.

## 6. Capability tiers

Not every child needs the full harness. Offer **tiers** and let the parent pick
per child:

- **Lite** — model call only, no tools. For trivial per-entry classification.
- **Standard** — tools + code, no further spawning. For bounded slice work.
- **Full** — tools + code + spawn. For region-children that must subdivide.

Tiering keeps cost proportional to the difficulty of each slice and prevents a
leaf task from accidentally launching a subtree.

## 7. Bounded recursion depth

Recursion without a hard depth cap is the fastest way to a runaway tree and a
runaway bill. Pass an explicit `depth` (or `depth_remaining`) to every child;
when it reaches zero the child must solve its slice without spawning, falling
back to the Lite/Standard tier. Choose the cap from the input size and the
branching factor: `depth >= log_branch(total_entries / leaf_size)` is enough to
reach the leaves, and there is rarely a reason to allow more.

## 8. Normalize-then-score

When children must emit a strict format (e.g. a label from a fixed set, JSON
matching a schema), do not ask the LLM to both decide *and* perfectly format in
one step, and do not score the raw text. Instead:

1. The child emits its best answer.
2. A deterministic **normalizer** maps it onto the canonical form (trim, lowercase,
   alias-map, coerce to schema).
3. A separate **scorer** compares normalized output to ground truth.

Separating normalization from scoring keeps near-miss formatting from being
counted as a wrong answer and keeps the scorer simple and model-independent.

## 9. Cost and latency control

The cost/latency profile of a recursive harness is asymmetric and worth
internalizing:

- **Latency is bounded by the slowest child**, not the sum — children run
  concurrently, so the critical path is one deep branch, not the whole tree.
- **Cost is dominated by re-reading shared context.** Every child re-ingests the
  shared reference material; with wide fan-out that prefix is paid N times.

The single highest-leverage optimization is therefore **prompt-caching the
shared prefix**: put the stable reference material at the front of every child's
context so the provider serves it from cache instead of re-billing it. Combine
with: trimming excerpts to what each slice needs, choosing the cheapest adequate
capability tier, and keeping the branching factor high (fewer, fatter levels
beat many thin ones for the same leaf count). For caching beyond the prompt
prefix — semantic/result caching and deduping identical subtasks — see
reliability-and-operations.
