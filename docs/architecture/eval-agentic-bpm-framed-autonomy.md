# Evaluation plan — Agentic BPM / "framed autonomy" (TDF · FRAME · CUGA FLO)

**Type:** Spike (~1–2 weeks). **Owner:** TBD. **Status:** Proposed.
**Paper:** *A Process Harness for Uplifting Legacy Workflows to Agentic BPM:
Design and Realization in CUGA FLO* (Fournier & Limonad, IBM; arXiv 2606.27188,
Jun 2026). **Informs:** [research-triage.md](research-triage.md) #4;
[governed-platform-architecture.md](governed-platform-architecture.md);
[enterprise-pitch.md](../enterprise-pitch.md).

## 1. The decision this spike informs

Two linked questions: (a) should we **adopt the paper's Task–Decision–Flow (TDF) /
FRAME formalism** as the rigorous description of our governed-agent architecture,
and (b) is **agentic BPM a vertical** worth pursuing — i.e. does **Open Engine
generalize from tracker tasks to deterministic workflow (BPMN) control points**?

Outcomes: **ADOPT-FORMALISM + PURSUE-VERTICAL** · **ADOPT-FORMALISM only** ·
**NEITHER** (note the parallel, move on).

## 2. Why this one matters

The paper is a peer-reviewed formalization of *our exact thesis*: a **process
harness** places a *policy-governed agentic layer around a deterministic workflow
engine, intercepting designated control points, while the engine retains
structural authority.* That is what we built — almost 1:1:

| Paper | Ours |
| :-- | :-- |
| **FRAME** — the aggregate policy set governing all LLM calls | **OPA/Rego bundle** as the declared sole authority |
| **TaskAgent** (knowledge-intensive task execution) | the worker/agent in a governed Omnigent session |
| **DecisionAgent** (per-case gateway routing) | a policy-gated routing/decision point |
| **FlowAgent** (runtime flow adaptation via hooks) | our **PreToolUse hooks** / runtime guardrails |
| **"framed autonomy"** | our "governed agent" — autonomy bounded by policy |
| deterministic engine keeps **structural authority** | the tracker/workflow keeps the lifecycle; we gate it |

It also opens a **vertical we are well-positioned for**: agentic BPM — uplifting
legacy BPMN workflows (Camunda/Flowable/IBM BAW) with a governed agentic layer
*without replacing the engine*, demonstrated on a **loan-approval workflow with
hook-driven regulatory override** — exactly our governed / audited / human-ASK
sweet spot.

## 3. Hypotheses

- **H1 (formalism fits):** TDF + FRAME describes our architecture faithfully; gaps
  (if any) point at things we should add (e.g. an explicit DecisionAgent role).
- **H2 (Open Engine generalizes):** our tracker→session loop generalizes to a BPMN
  **control point** — a DecisionAgent at a gateway and a FlowAgent at a hook —
  with OPA as the FRAME, without replacing the workflow engine.
- **H3 (vertical is real):** governed agentic BPM (deterministic engine + our
  policy-framed agent layer + human-ASK + three-plane audit) is a credible
  enterprise offer (loan approval / regulatory override / claims).

## 4. Make-or-break questions

| # | Question | Gate |
| :-- | :-- | :-- |
| Q1 | Does TDF/FRAME describe us completely, or expose missing roles/structure? | Whether to adopt the formalism. |
| Q2 | Can we attach our governance to a real workflow engine's **control points** without owning the lifecycle? | Whether the vertical is feasible. |
| Q3 | Does a DecisionAgent's per-case routing map to an **OPA decision** (deterministic, auditable)? | Keeps "no LLM for security" intact. |
| Q4 | Does the FlowAgent **hook** model add anything our PreToolUse hooks don't? | What to borrow. |
| Q5 | Is the agentic-BPM use-case differentiated for our pitch (vs theirs / vs vanilla BPM)? | GTM value. |

## 5. Experiments

> Prereq: the cloned repo at `/home/bittahcriminal/air/workspace/research-repos/cuga-agent`
> **on branch `cugaflo`** (FLO is NOT on `main`). The runnable loan demo is
> `docs/examples/flow_agent_app_inline/loan_approval/`. **Don't stand up Camunda** —
> reuse FLO's engine-agnostic MCP contract (see §9).

- **E1 — formalism mapping (Q1).** Write our architecture in TDF/FRAME terms:
  identify our TaskAgent/DecisionAgent/FlowAgent equivalents and our FRAME (the OPA
  bundle + session policies). Output: a mapping table + a list of **gaps**
  (e.g. do we have a first-class DecisionAgent, or is routing implicit?).
- **E2 — control-point prototype (Q2/Q3).** Stand up a trivial BPMN process on one
  engine with one **gateway** and one **service task**. Attach our governance: at
  the gateway, a **DecisionAgent** whose routing is an **OPA decision** (not an LLM
  judgment); at the task, a governed agent call gated by the native plane. Confirm
  the engine keeps structural authority and we only intercept the control points.
  Output: a working "process harness over an engine" PoC (or a documented blocker).
- **E3 — hook comparison (Q4).** Compare the FlowAgent **hook** semantics to our
  PreToolUse hooks + the Omnigent ASK gate. Output: what (if anything) to borrow —
  e.g. a flow-adaptation hook richer than a pre-tool deny.
- **E4 — human-in-the-loop / regulatory override (Q3/Q5).** Reproduce the loan-
  approval "regulatory override" as our `require_approval` → ASK → human hold →
  resume, with the three-plane audit capturing it. Output: an end-to-end governed-
  BPM trace.
- **E5 — GTM framing (Q5).** Draft the agentic-BPM positioning for
  [enterprise-pitch.md](../enterprise-pitch.md): "framed autonomy over your existing
  workflow engine — no rip-and-replace."

## 6. Success criteria / decision gates

- **ADOPT-FORMALISM + PURSUE-VERTICAL** if E1 shows a faithful mapping (with at
  most additive gaps), E2 proves we can govern an engine's control points with OPA
  as the FRAME, and E4 shows the human-ASK + audit story end-to-end.
- **ADOPT-FORMALISM only** if the mapping is strong but the BPM-engine integration
  isn't worth the lift for our roadmap — still adopt TDF/FRAME vocabulary in our
  architecture docs.
- **NEITHER** if the formalism doesn't fit or the vertical is a distraction —
  record the parallel and move on.

## 7. Risks

- **Scope creep into BPM** — keep the prototype to one gateway + one task; the goal
  is a *decision*, not a product.
- **DecisionAgent-as-LLM temptation** — per-case routing must stay an **OPA
  decision** to preserve "no LLM for security"; an LLM may *propose*, OPA *decides*.
- **Vocabulary churn** — adopting TDF/FRAME terms means a docs pass; only do it if
  the mapping is genuinely cleaner than our current "MCP/native plane" language.

## 8. Deliverables

The TDF/FRAME mapping + gap list (E1), a control-point governance PoC (E2), a hook
borrow/skip note (E3), an end-to-end governed-BPM trace (E4), a GTM paragraph
(E5), and a **1-page adopt/pursue/neither memo**.

## 9. Repo deep-dive — grounded specifics

Clone: `/home/bittahcriminal/air/workspace/research-repos/cuga-agent` — **checkout branch `cugaflo`** (`main` is generic CUGA; FLO/TDF is realized, not design-stage, on `cugaflo`). FLO dir: `src/cuga/backend/cuga_graph/nodes/cuga_flow/` (`flow_agent.py`, `decision_agent.py`, `task_agent.py`, `hook_manager.py`, `langgraph_engine.py`, `workflow_engine.py` ABC, `bpmn_parser.py`); MCP bridge: `src/cuga/backend/server/cuga_flo_mcp/bridge.py` (`MCPFlowBridge` exposes `execute_task`/`route_gateway`/`evaluate_hook`/`get_static_config`/`run_process`). **Run the loan demo:** `python docs/examples/flow_agent_app_inline/run.py loan_approval` (default applicant `A12345`; **`applicant_id="4321"` triggers the regulatory-override hook**).

- **E1 — the TDF tuples are the mapping table's left column.** `TaskAgent a=(t,C_T,P,Ts,λ)`, `DecisionAgent d=(g,C_D,P,Ts∪{cond_eval},λ)`, `FlowAgent f=(H,Is,φ,C_F(h_k),…)`; **FRAME `F = {C_T}∪{C_D}∪{C_F}`** = per-element **markdown policy files** (`loan_approval/policies/task-*.md`, `decision-*.md`, `hook-*.md`). Map our OPA bundle → `F`, Rego packages → per-element policies, `input.*` → `ControlPointContext`.
- **E1 gap to mirror — a SECOND governance plane `φ`.** Beyond FRAME (what the LLM may *reason*), `φ: Is→{permitted,prohibited}` bounds what the engine may *act on* (YAML `action_permissions`), enforced **after** hook reasoning. This is "LLM proposes, OPA admits" by construction — add a first-class `φ` to our model if missing.
- **E2 — reproduce the shipped loan demo (1 XOR gateway `Gateway_09ad5fc` + 1 service task `Activity_0oydey5`); don't build a BPMN.** Engine-agnosticism is already proven via the `WorkflowEngine` ABC + `MCPFlowBridge` + `docs/examples/flow_agent_app_inline/external_mcp_engine/server.py` (a separate engine over MCP). **Our PoC implements the same `execute_task`/`route_gateway`/`evaluate_hook` MCP contract** backed by OPA — lowest lift, directly comparable.
- **E2/E3 — DecisionAgent already matches "no LLM for security"; cite as evidence, not risk.** `decision_agent.py` is two-step: **Step-1 deterministic `${var}` substitution + compare, no `eval()`** (TRUE/FALSE/UNKNOWN); **Step-2 LLM only when inconclusive.** Map Step-1 → an OPA decision; Step-2 = "LLM proposes within `C_D`, OPA validates the flow ID." Mitigates Risk §7's DecisionAgent-as-LLM.
- **E3 — their hook is RICHER than our PreToolUse deny; borrow the vocabulary.** FlowAgent hook returns one of **7 interventions**: `CONTINUE / SKIP_NODE / SKIP_TO / SWAP_NODES / TERMINATE / REMOVE_NODE / ADD_NODE` (topology edits limited to unexecuted nodes), per-edge, with full process state + remaining-tasks. Borrow a flow-adaptation result type + the `φ` admission gate over it.
- **E4 — the regulatory override IS our ASK pattern.** Artifact `loan_approval/policies/hook-audit_credit_decision.md` fires on the approval edge `Flow_0ybszcv`; for `applicant_id=="4321"` it `skip_to`s rejection with `policy_override_active=true`; nominal case routes on `credit_score>0.6` and returns `CONTINUE`. Realize E4 by swapping `skip_to` → a `TERMINATE`/hold + our human-ASK; two independently-governed FRAME planes (gateway-on-score, hook-on-id) = our three-plane audit story.
- **E5 wedge (differentiation):** FLO ships a LangGraph demo engine + **LLM-interpreted markdown policies with NO deterministic policy engine** — their §7.2 names "automated FRAME consistency checking" an **open problem**. **OPA/Rego is exactly that missing machine-checkable FRAME** (+ deterministic DecisionAgent Step-1 + three-plane audit). That's our GTM wedge over their own FLO.
- **Repo-vs-paper, plainly:** FLO present + runnable + faithful; **not present** = production engine adapters (only LangGraph complete; Flowable/`ordo` are WIP branches) and any deterministic policy engine. "FLO is real; the BPM-engine breadth and machine-checked governance are the open space we'd fill."
