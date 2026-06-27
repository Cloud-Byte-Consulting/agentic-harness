# Evaluation plan — Recursive Language Models (RLMs) as a governed long-context capability

**Type:** Time-boxed spike (~1–2 weeks). **Owner:** TBD. **Status:** Proposed.
**Paper:** *Recursive Language Models* (Zhang, Kraska, Khattab, MIT CSAIL;
arXiv 2512.24601v3, May 2026). Code: `github.com/alexzhang13/rlm`.
**Relates to:** the Omnigent runtime, our compaction, the token plane
(Cachy / token-dashboard / TEO), the native enforcement plane, and the supervisor
/ sub-agent guardrails (`spawn_bounds`, `hillclimb_budget`, `cost_plan`).

## 1. The decision this spike informs

Should we add an **RLM-style recursive long-context scaffold** to the Omnigent
runtime — and, distinctively, **govern it with the guardrails the paper leaves
open** (sub-call caps, cost budgets, sandboxed REPL, native-plane authz)?

Outcomes: **GO** (offer RLM as a governed long-context mode) · **PARTIAL** (opt-in
tool for specific task types) · **NO-GO** (compaction stays; document why).

## 2. What RLM is, and why it's relevant to us

An RLM loads an arbitrarily long prompt as a **variable in a Python REPL** (not the
context window) and lets the root LM **write code to peek/decompose/transform it
and recursively call itself (or sub-RLMs) over slices** — "symbolic recursion".
It processes **10M+ tokens**, and on long-context tasks beats vanilla LMs and
common scaffolds **at comparable or cheaper cost**: e.g. BrowseComp-Plus (6–11M
tokens) RLM(GPT-5) 92.0 @ ~$0.99 vs base 0.0 (won't fit) and **>29% over
compaction + retrieval**; OOLONG +28–33% over base; CodeQA (up to 4.2M tokens)
66.0 vs base 24.0 and Claude Code 12.0/62.0. A small model post-trained on 1,000
unrelated samples (RLM-Qwen3-8B) gains +28% and runs ~3× faster.

**Two reasons this matters to our platform:**

1. **It's a stronger long-context strategy than what Omnigent does today.** Our
   runtime handles overflow via **compaction** (`CompactionData` items) — which
   the paper explicitly *out-performs* (RLM beats compaction by ~26% median).
   For OE tasks over large codebases, long issue/log histories, or big documents,
   an RLM scaffold could materially improve agent quality.

2. **It's a new execution surface that needs exactly the guardrails we have — and
   the paper says are missing.** An RLM runs **agent-written code in a REPL** and
   **fans out recursive sub-LM calls**. The paper's own Limitations: *"the best
   mechanisms for implementing guardrails for RLMs remain underexplored"*,
   *"unintentional side-effects like exploding sub-call costs"*, and *"sandboxed
   REPLs … future work."* We already own all three:
   - **Sub-call fan-out** → `spawn_bounds` (per-turn dispatch cap) + `hillclimb_budget`.
   - **Exploding cost** → `cost_plan` session budgets + the ASK-on-threshold pause.
   - **Sandboxed REPL** → the Omnigent OS sandbox (fs isolation + network interception).
   - **The REPL's code-exec itself** → the native enforcement plane
     (`opa_delegate` / `opa-hook`), and — because RLM-written code can shell out —
     the **ActPlane / eBPF** indirect-path concern (see cross-link below).

   So the spike's thesis: **Omnigent could be the *governed* RLM runtime** — the
   open problem in the paper is the capability we already have.

## 3. Hypotheses

- **H1 (quality):** an RLM scaffold beats Omnigent's compaction on a representative
  long-context OE task, at comparable cost.
- **H2 (cost containment):** `spawn_bounds` + `cost_plan` + ASK-on-cost cap the
  "exploding sub-call cost" risk — a deep/runaway recursion hits a budget/spawn
  cap and pauses, rather than blowing up.
- **H3 (governed code-exec):** denied actions emitted *inside* RLM-generated REPL
  code are caught by our enforcement — exposing whether tool-call gating suffices
  or whether the ActPlane/eBPF layer is required (RLM code-exec is *indirect
  execution* by construction).
- **H4 (sandbox):** the REPL runs cleanly inside the Omnigent sandbox.

## 4. Scope

**In:** RLM as a long-context capability *inside* a governed Omnigent session —
quality vs compaction, cost containment, governing the REPL + recursion, the
integration shape. **Out:** native RLM training/post-training (note as future
upside, not spike-scoped); replacing the agent loop wholesale; async sub-call
optimization (the paper's own future work).

## 5. Make-or-break questions

| # | Question | Gate |
| :-- | :-- | :-- |
| Q1 | Does RLM beat our **compaction** on a representative long-context OE task at comparable cost? | The capability bet. |
| Q2 | Do `spawn_bounds` + `cost_plan` + ASK actually **bound** the recursive sub-call cost? | Safe to expose. |
| Q3 | Can the native plane **govern the REPL's code-exec** — and does it need the ActPlane/eBPF layer (shell-outs from RLM code)? | Security of the new surface. |
| Q4 | Does it run in the **Omnigent sandbox** (fs + network isolation)? | Containment. |
| Q5 | What's the **integration shape** — a long-context *tool/skill* the agent invokes, or a session/harness *mode*? | Productization. |

## 6. Experiments

> Prereqs: clone `github.com/alexzhang13/rlm`; a model via our normal credential
> path; the Omnigent sandbox; `spawn_bounds` / `cost_plan` wired (as in the
> `examples/role_router` supervisor); the native plane in `enforce`.

- **E1 — quality vs compaction (Q1).** Pick 2 representative long-context tasks:
  (a) a CodeQA-style query over a large repo (e.g. ask about a 1M-token slice of
  our own `omnigent`/`agentic-harness` trees), (b) a long aggregation (a big issue
  thread / log). Run three ways: **base harness** (native truncation), **Omnigent
  compaction**, **RLM scaffold (depth 0 and 1)**. Record quality + API cost + wall
  time. Output: a quality/cost table vs the paper's compaction baseline.
- **E2 — cost containment (Q2).** Instrument the RLM run to count sub-LM calls +
  cost. Wire `cost_plan`'s session budget + ASK-on-threshold and `spawn_bounds`
  (cap recursive dispatches/turn). Force a runaway (deep recursion / a pathological
  prompt) and confirm it **hits the spawn cap / cost budget and ASK-pauses** rather
  than exploding. Output: the bounded-vs-unbounded cost curves.
- **E3 — governing the REPL code-exec (Q3).** Run the RLM REPL inside a session
  with the native plane in `enforce`. Feed a task whose decomposition would emit a
  **denied action inside generated code** (write outside the workspace, a network
  egress, a destructive op, a `git commit` on a protected boundary). Check: is it
  caught? **Expectation:** tool-call gating (`opa_delegate`/`opa-hook`) sees the
  REPL invocation but **not** what the generated code does internally — i.e. RLM is
  a concrete instance of the **indirect-execution gap**, so this experiment likely
  *motivates* the [ActPlane/eBPF spike](eval-actplane-native-plane.md). Output: a
  matrix of which denied actions are caught at which layer.
- **E4 — sandbox (Q4).** Run the REPL inside the Omnigent OS sandbox; confirm fs
  isolation + network interception apply to the generated code (not just the agent
  tool calls). Output: pass/fail + any sandbox gaps for in-REPL execution.
- **E5 — integration shape (Q5).** Prototype RLM two ways and compare DX/governance:
  (a) an Omnigent **tool/skill** the agent calls for a long-context sub-task
  (`rlm_query(haystack, question)`), vs (b) a session/harness **mode**. Assess
  which composes better with our guardrails + the supervisor. Output: a
  recommendation.

## 7. Success criteria / decision gates

- **GO** if: E1 shows RLM beats compaction on our task profile at comparable cost
  **and** E2 shows our budgets/caps contain the sub-call cost **and** E3/E4 show we
  can govern + sandbox the REPL (even if E3 concludes we *need* the eBPF layer for
  full coverage — that's a known, planned follow-on, not a blocker for an opt-in
  governed mode). → Offer RLM as a **governed long-context mode**, and use it as
  the lead example of "the runtime that closes the RLM guardrails gap."
- **PARTIAL** if RLM wins only on a subset (e.g. info-dense aggregation, huge
  codebases) → ship it as an **opt-in tool** for those task types, keep compaction
  the default.
- **NO-GO** if RLM doesn't beat compaction for our workloads, or the cost/governance
  can't be contained, or the added complexity isn't justified → keep compaction,
  document the result.

## 8. Risks & mitigations

- **Exploding sub-call cost** (the paper's named risk) → the spike's whole point is
  to prove our `spawn_bounds` + `cost_plan` contain it (E2); if they can't, NO-GO.
- **New code-exec attack surface** → RLM-written code is *indirect execution*; tool-
  call gating is insufficient by construction (E3) — pairs with the ActPlane spike;
  meanwhile the sandbox (E4) is the containment floor.
- **Blocking/sequential sub-calls** → latency; the paper notes async sub-calls as
  future work — out of scope, note the latency profile.
- **Research-grade scaffold** → treat the *paradigm* (REPL-as-environment + symbolic
  recursion) as the durable idea; we can re-implement the scaffold against our
  harness rather than ship their code verbatim.

## 9. Deliverables

A quality/cost comparison vs compaction (E1), a bounded-cost demonstration under
our budgets (E2), an enforcement-coverage matrix for the REPL code-exec (E3) +
the explicit "does this need eBPF?" finding, a sandbox result (E4), an
integration-shape recommendation (E5), and a **1-page go/partial/no-go memo** —
including whether to position Omnigent as "the governed RLM runtime."

## 10. Cross-links

- The **REPL code-exec governance** question (E3) is the same indirect-execution
  problem as the [ActPlane native-plane spike](eval-actplane-native-plane.md) —
  run them together if both are greenlit.
- Cost containment leans on the same `cost_plan` / `spawn_bounds` machinery used by
  the `examples/role_router` supervisor (the role-router → Omnigent-supervisor port).
- This is primarily a **token-plane + runtime** capability, distinct from the
  authz/data-plane comparisons in [agentgateway-comparison.md](agentgateway-comparison.md).

## 11. Repo deep-dive — grounded specifics

Clone: `/home/bittahcriminal/air/workspace/research-repos/rlm` (PyPI pkg `rlms`). Install `pip install rlms` (+ `[docker]`/`[modal]`/`[e2b]` extras). Demo: `make quickstart` (self-contained needle-in-haystack). Visualizer `cd visualizer && npm run dev` (localhost:3001) renders the recursion-tree + per-call tokens — good go/no-go memo material.

- **E1 entry point (real API):** `RLM(backend="openai", backend_kwargs={...}, environment="local"|"docker", max_depth=0|1, logger=RLMLogger(...)).completion(prompt, root_prompt=question).response` (`rlm/core/rlm.py:326`). For our E1: pass a repo-slice **string/`list[str]` as `prompt`**, the question as `root_prompt`; depth-0 (REPL only) vs depth-1 (sub-calls) gives the paper's comparison. Context is written to a temp file and `read()` into a `context` REPL var (`local_repl.py:403`) → never enters the root window (the whole point).
- **E2 fan-out chokepoints (where to wire our guardrails):** recursion `rlm/core/rlm.py::_subcall:706`; flat sub-calls funnel through the socket broker `rlm/core/lm_handler.py::LMHandler.get_client()` — **single chokepoint** for token-plane metering + an ASK gate.
- **E2 gaps to close (RLM's native knobs ≠ ours):** `max_concurrent_subcalls` caps **concurrency, not total dispatches/turn** → **`spawn_bounds` is NOT native**; use the `on_subcall_start(depth,model,preview)` constructor callback to count + cap per-turn. `max_budget` (USD) **hard-raises `BudgetExceededError`, no ASK-pause, checked only between root iterations**, and is a **no-op unless the backend reports cost** (plain OpenAI may not) → **derive cost from tokens** (our Cachy/TEO already does); route via OpenRouter/Portkey only if you want its native budget to fire.
- **E3 indirect-exec is provable by static read (no experiment needed to confirm the gap):** `local_repl.py::execute_code` runs `exec(code, …)` with `_SAFE_BUILTINS` that blocks `eval/exec/compile` but **allows `__import__`, `open`, `getattr/setattr`** (`:55`) → model-generated `import os; os.system(...)` runs in-process; tool-call gating sees only `rlm.completion()`. README states the local REPL is "not for production." The experiment just *demonstrates* it.
- **E4 sandbox — gate moves from "does it run" to "add the missing flags".** Easiest isolated env is `environment="docker"` (`make docker-repl`, auto-pulls `python:3.11-slim`). But `docker_repl.py:524` launches with **no `--network none`, no `--memory`/`--cpus`, no `--read-only`, no `--cap-drop`**, and **bind-mounts a host dir** into `/workspace` — open egress (it *needs* net to reach the host LM proxy) + host-fs writes. Our sandbox must add cap-drop/mem/cpu/read-only while preserving the `host.docker.internal` proxy path. Sub-RLMs are spawned **host-side** (code phones home to `LMHandler`), so our caps sit host-side **regardless of env**.
- **E5 integration shape:** `custom_tools={name: callable|value}` injects governed tools into the **root** REPL only — **sub-LLMs need `custom_sub_tools` explicitly** or they run ungoverned. A governed RLM mode injects our enforced tools via both, not by patching their code.
- **Default depth-1** (their cost/quality sweet spot); depth>1 is opt-in (helps info-dense aggregation, but Qwen3-Coder degrades with depth via sub-call syntax errors). Lighter refs if the full repo is heavy: `alexzhang13/rlm-minimal`, `viplismism/rlm-cli`.
