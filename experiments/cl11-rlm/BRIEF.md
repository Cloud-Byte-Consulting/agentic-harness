# CLO-11 — RLM governed long-context: complete implementation brief

Self-contained context to finish CLO-11. Pairs with the spike
[eval-rlm-long-context.md](../../docs/architecture/eval-rlm-long-context.md) (the full
grounded deep-dive) and [target-architecture-all-spikes.md](../../docs/architecture/target-architecture-all-spikes.md).

## 0. What to build (TL;DR)

A **governed, opt-in `rlm_query` tool** the agent invokes **on-demand** for long-context
jobs — wrapping the RLM library with our caps (sub-call fan-out, cost, sandbox) so it
can't blow up cost or escape governance. Not a default; not a session mode.

## 1. Decision history (settled)

- **RLM** = recursive long-context: load the prompt as a Python-REPL *variable* and let
  the model write code that recursively calls itself over slices. Beats compaction
  ~26% (paper; unmeasured on our stack).
- **Decision (2026-06-27):** KEEP, **opt-in now → target on-demand, agent-invoked**
  (the E5(a) `rlm_query` tool, not a session mode).
- **Blocked by CLO-10 (ActPlane)** for *full* in-REPL exec governance. Without it, the
  **Omnigent sandbox is the containment floor** — so a first cut can ship opt-in with
  the sandbox; full coverage waits on CLO-10.

## 2. The RLM library (confirmed against the clone)

- Repo: `research-repos/rlm`. Install: **`pip install 'rlms[docker]'`** (extras:
  `[modal]`/`[e2b]`/`[daytona]`/`[prime]`).
- Entry: `RLM(backend="openai", backend_kwargs={"model_name": …}, environment="local"|"docker",
  max_depth=1, …).completion(prompt, root_prompt=question).response`
  — `prompt` = the haystack (str | list[str]); `root_prompt` = the small question the
  root LM sees; read `.response`. Constructor: `rlm/core/rlm.py:41`.
- **Governance hooks (constructor args):**
  - `on_subcall_start: Callable[[depth, model, preview], None]` — fires on every
    sub-call → our per-turn fan-out cap hooks here.
  - `max_concurrent_subcalls` (default 4) — caps *concurrency, NOT total* dispatches.
  - `max_budget` (USD) — **no-op unless the backend reports cost** (e.g. OpenRouter);
    plain OpenAI does not → derive cost from tokens.
  - `custom_tools` (root REPL only) / `custom_sub_tools` (sub-agents; **set explicitly
    or sub-LLMs run ungoverned**).
  - `max_depth` (1 = the cost/quality sweet spot; >1 opt-in).
- **Fan-out chokepoints:** `rlm/core/rlm.py::_subcall` + `rlm/core/lm_handler.py::LMHandler`
  (the socket broker — single host-side chokepoint for every sub-LM request).
- **Indirect-exec surface:** `rlm/core/local_repl.py::execute_code` runs `exec()` with
  `__import__`/`open` allowed → model-written code can shell out. Tool-call gating
  (`opa_delegate`/`opa-hook`) sees only `rlm.completion()`. README: local REPL is "not
  for production."
- **Sandbox:** `environment="docker"` (`make docker-repl`) — but stock Docker has
  **open egress + host bind-mount + no resource caps**; sub-RLMs spawn **host-side**
  (caps sit host-side regardless of env).

## 3. Integration design

- New **omnigent tool/skill `rlm_query`** — sibling to `examples/role_router/skills/judge/`.
  `rlm_query(haystack, question) -> answer`.
- **Agent auto-select trigger (the on-demand part):** the agent reaches for `rlm_query`
  when a job needs dense access over a large input — signal on a context-won't-fit /
  over-token-budget estimate, or an explicit job label. Surface as a skill the
  supervisor (role_router) can dispatch.

## 4. Governance wiring (the heart of the work)

| Concern | RLM reality | Wire to |
| :-- | :-- | :-- |
| **Sub-call fan-out cap** (per turn) | not native (`max_concurrent_subcalls` = concurrency only) | a counter in `on_subcall_start` enforcing a per-turn dispatch cap — mirror **`spawn_bounds`** (`omnigent/omnigent/inner/nessie/policies.py:408`) |
| **Cost cap + ASK** | `max_budget` no-op w/o cost-reporting backend | derive cost from **tokens** via the token plane (Cachy / TEO / token-dashboard); **⚠ a `cost_plan` nessie policy does NOT exist — build it** (closure-budget like `spawn_bounds`) or reuse the budgeting pattern; ASK-pause at a threshold via Omnigent |
| **REPL code-exec** | `exec()` allows `__import__`/`open` → indirect exec uncovered | **sandbox floor now**; **ActPlane (CLO-10)** for full kernel coverage |
| **Sandbox hardening** | stock docker env unhardened | add `--network none` (or our interception) + `--cap-drop`/`--memory`/`--cpus`/`--read-only`; keep the `host.docker.internal` proxy for sub-calls |
| **Inject governed tools** | `custom_tools` = root only | also set **`custom_sub_tools`** |

## 5. Experiments / decision gates (from the spike §5–7)

- **E1** quality vs Omnigent compaction on a long-context OE task (e.g. a 1M-token slice
  of our own tree), at comparable cost.
- **E2** force a runaway recursion → the `on_subcall_start` cap + token-budget ASK-pause
  contain it (bounded vs unbounded cost curves).
- **E3** a denied action inside RLM-generated code → caught? (expect tool-call gating
  misses it → confirms the ActPlane dependency).
- **E4** runs in the hardened sandbox (fs+net isolation apply to generated code).
- **E5** the tool shape + the agent's auto-select trigger.
- **Gates:** GO (opt-in governed mode) · PARTIAL (specific task types) · NO-GO
  (compaction stays). GO needs: beats our compaction (E1) AND caps contain cost (E2)
  AND governable+sandboxed (E3/E4 — E3 may *require* ActPlane; that's a planned
  follow-on, not a blocker for an opt-in mode).

## 6. Environment requirements

- Python ≥3.11; `pip install 'rlms[docker]'`.
- **A model key — there is NONE in this workspace.** Provide `OPENAI_API_KEY`
  (or `PORTKEY_API_KEY` / OpenRouter — OpenRouter is needed for `max_budget` to fire;
  otherwise derive cost from tokens).
- Docker (present) for the sandbox env.

## 7. Gotchas / open items

- **`cost_plan` doesn't exist** — build it (or token-derive budgeting). The spike
  assumed it; only `spawn_bounds` + `hillclimb_budget` are in `nessie/policies.py`.
- **No model key here** — the implementer supplies one.
- **ActPlane (CLO-10)** is the full-governance precondition; ship opt-in with the
  sandbox floor meanwhile.
- RLM is **research-grade** ("not for production"); the durable idea is the *paradigm*
  (REPL-as-environment + symbolic recursion) — consider re-implementing a thin scaffold
  against our harness rather than shipping the lib verbatim (lighter refs:
  `alexzhang13/rlm-minimal`, `viplismism/rlm-cli`).

## 8. Key files / references

- Spike (deep-dive): `docs/architecture/eval-rlm-long-context.md`
- RLM lib: `research-repos/rlm/rlm/core/{rlm.py,lm_handler.py,local_repl.py,docker_repl.py}`
- Our hooks: `omnigent/omnigent/inner/nessie/policies.py` — `spawn_bounds:408`, `hillclimb_budget:556`
- Skill pattern: `omnigent/examples/role_router/skills/judge/SKILL.md`
- CLO-10 dependency: `docs/architecture/eval-actplane-native-plane.md`, `research-repos/ActPlane`, `scripts/cl10-actplane.sh`

## 9. The OE process (how to work it)

Claim it the Open Engine way:
`python ~/air/workspace/scripts/oe.py claim 11 <<< "…"` → In Progress + AGENT CLAIMED
receipt → implement → tests → `oe.py done 11` → In Review. CLO-11 is **blocked-by
CLO-10** in Linear; respect that for the *full* mode, proceed for the opt-in cut.
See the [[oe-working-loop]] memory.
