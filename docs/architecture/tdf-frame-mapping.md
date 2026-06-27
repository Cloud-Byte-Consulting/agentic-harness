# Our architecture in TDF / FRAME terms (framed autonomy)

**Status:** Adopted formalism (CLO-8, 2026-06-27). Adopts the **Task–Decision–Flow
(TDF)** / **FRAME** vocabulary from IBM's *A Process Harness for Agentic BPM*
(arXiv 2606.27188) as the canonical, peer-reviewed description of our governed-agent
architecture. Decision to adopt: [eval-agentic-bpm-framed-autonomy.md](eval-agentic-bpm-framed-autonomy.md);
fits within [target-architecture-all-spikes.md](target-architecture-all-spikes.md).

## Why this vocabulary

Their **"framed autonomy"** *is* our "governed agent." A **process harness** = a
policy-governed agentic layer around a deterministic engine, intercepting designated
control points, with the engine keeping **structural authority** — which is exactly
our thesis. Adopting their formalism gives us rigorous names for what we built. We
adopt the **vocabulary + the φ concept** only; we do **not** adopt their
LLM-interpreted-markdown FRAME — OPA/Rego is our *machine-checkable* FRAME (their
§7.2 names automated FRAME-consistency an open problem; that's our wedge).

## TDF agent types ↔ our components

| TDF (their tuple) | What it is | Ours |
| :-- | :-- | :-- |
| **TaskAgent** `a = (t, C_T, P, Ts, λ)` | knowledge-intensive task execution under policy `C_T` | the worker/sub-agent in a governed Omnigent session (dispatched with an explicit `args.purpose`: implement / explore / review) |
| **DecisionAgent** `d = (g, C_D, P, Ts∪{cond_eval}, λ)` | per-case gateway routing, **deterministic-first** | a policy-gated routing point: **Step-1 deterministic `${var}` compare → an OPA decision; Step-2 LLM *proposes within* `C_D`, OPA validates the chosen flow id** (keeps "no LLM for security") |
| **FlowAgent** `f = (H, Is, φ, C_F, P, Ts, μ, λ)` | runtime flow adaptation via hooks | our **PreToolUse hooks** (`opa_delegate`/`opa-hook`) + the supervisor's runtime guardrails (`spawn_bounds`, `hillclimb_budget`) |

## FRAME ↔ our OPA bundle

**FRAME** `F = {C_T(t_i)} ∪ {C_D(d_j)} ∪ {C_F(h_k)}` — the aggregate policy set
governing **all** LLM calls, declared per element. Ours:

- the **OPA/Rego bundle is the FRAME** (the single policy *source*);
- Rego packages are the **per-element policies** `C_T`/`C_D`/`C_F`;
- `input.*` is the **ControlPointContext** `P = (Pm, Ps, Ph)`.

Invariant preserved: OPA is the sole policy **source**; every per-element policy and
every downstream enforcement artifact compiles from it (never hand-authored).

## φ — the second governance plane (the gap we adopt)

TDF separates two governance **moments**, which we previously had only implicitly:

- **FRAME (`C_*`)** governs what the LLM may **reason**.
- **φ: `Is → {permitted, prohibited}`** governs what the engine may **act on**, applied
  **after** the agent reasons (post-hook admission), declared as `action_permissions`.

This is **"LLM proposes, OPA decides"** as a structural element: the agent proposes
an action/intervention; **φ (an OPA decision, not an LLM)** admits it or substitutes a
safe default. **We adopt φ as a first-class artifact.** Today our admission is
implicit in the gateways/hooks; making it an explicit, named **φ check per control
point** (backed by OPA) is the concrete output of this adoption — and closes the BPM
spike's E1 gap.

## Interventions ↔ ours

FlowAgent's **7 structural interventions** — `CONTINUE / SKIP_NODE / SKIP_TO /
SWAP_NODES / TERMINATE / REMOVE_NODE / ADD_NODE` (topology edits limited to unexecuted
nodes) — are richer than our binary PreToolUse allow/deny. **Borrow:** a
flow-adaptation result type richer than deny, admitted through φ. (Not adopted yet —
listed as a gap below.)

## How it sits over our planes

```
 tracker / workflow (Open Engine)   ── keeps STRUCTURAL authority (the lifecycle)
        │  control points (task · gateway · hook)
 agentic layer (Omnigent + agents)  ── reasons under FRAME (= the OPA bundle)
        │  proposes actions / interventions
 φ  (execution-admission)           ── OPA decision: permit / substitute-safe-default
        │
 enforcement planes                 ── MCP (Sentry) · native (opa-hook → ActPlane)
 audit                              ── oe_correlate joins session_id / subject_id / task_ref
```

The deterministic backbone keeps the lifecycle; the agent reasons under the FRAME at
each control point; φ admits what it may act on; the enforcement planes enforce below;
the three-plane audit threads it.

## Gaps / next (the E1 output)

1. **Make φ first-class** — a named `action_permissions`/φ admission check per control
   point, backed by an OPA decision (today it's implicit in the gateways/hooks).
2. **Make the DecisionAgent split explicit** in our routing — Step-1 = deterministic
   OPA decision; Step-2 = LLM-proposes-within-`C_D`, OPA-validates.
3. **Richer FlowAgent result type** — consider the 7-intervention vocabulary in place
   of binary deny, gated by φ.

Items 1–3 are the concrete follow-on tasks; this doc adopts the **vocabulary** and the
**φ plane** so the rest of our architecture can be described and extended in these terms.
