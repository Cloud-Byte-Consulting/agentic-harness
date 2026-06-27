# Research triage — spike-worthy vs concepts-only

**Date:** 2026-06-27. **Purpose:** for a batch of papers/docs, decide whether each
warrants a **separate spike** or only contributes **concepts** to the governed
agentic platform (or is out of scope). Companion to the spikes in this folder
([agentgateway](eval-agentgateway-mcp-plane.md) · [ActPlane](eval-actplane-native-plane.md)
· [RLM](eval-rlm-long-context.md) · [Agentic-BPM](eval-agentic-bpm-framed-autonomy.md)
· [semantic-stopping](eval-semantic-early-stopping.md)). Each spike now carries a
**"Repo deep-dive — grounded specifics"** section from the cloned repos (below).

## Summary

| # | Document | What it is | Verdict | Priority |
| :-- | :-- | :-- | :-- | :-- |
| 1 | The New SDLC with Vibe Coding (Osmani/Saboo/Kartakis, Google, May 2026) | Strategy whitepaper: agentic engineering, harness engineering, context engineering, factory model, economics | **Concepts** (strategic) | Med |
| 2 | TOPS: Visual Token Pruning (arXiv 2606.27161) | Visual-token pruning for **MLLM** inference efficiency | **Skip** | — |
| 3 | AURORA-AI (arXiv 2606.27005) | HJB/Lyapunov closed-loop + fairness-aware compute orchestration across a model fleet | **Concepts** (low) | Low |
| 4 | Process Harness for Agentic BPM / CUGA FLO (IBM, arXiv 2606.27188) | Policy-governed agentic layer over a deterministic workflow engine; TDF model + FRAME; "framed autonomy" | **SPIKE** (strong) | **High** |
| 5 | Prompt Injection in Résumé Screening (arXiv 2606.27287) | Empirical: prompt injection distorts LLM-judged hiring | **Concepts** (reinforcement) | Low |
| 6 | Semantic Early-Stopping for Agent Loops (arXiv 2606.27009) | Embedding-convergence halting for Writer→Critic loops; −38% tokens; "which round is best" | **SPIKE** (small) | **High** |

## Per-document evaluation

### 4 — Process Harness for Agentic BPM (IBM CUGA FLO) → **SPIKE, highest value**
A peer-reviewed formalization of *our exact thesis*: a **policy-governed agentic
layer around a deterministic engine, intercepting control points, while the engine
keeps structural authority.** Near-1:1 mapping to us:
- their **FRAME** (aggregate policy set governing all LLM calls) ≈ our **OPA bundle
  as sole authority**;
- **TaskAgent / DecisionAgent / FlowAgent** ≈ our supervisor agent roles;
- the **hook mechanism** (FlowAgent governs runtime flow adaptation) ≈ our
  PreToolUse hooks;
- **"framed autonomy"** ≈ our "governed agent."

It also opens a **new vertical** — agentic BPM (uplifting legacy BPMN workflows,
demo'd on loan approval with regulatory override), which is precisely our
governed/audited/human-in-the-loop sweet spot.
**Spike scope:** (a) adopt the **Task–Decision–Flow (TDF)** formalism + the FRAME
vocabulary as a rigorous description of our architecture; (b) assess whether
**Open Engine generalizes from trackers to BPMN control points** (a DecisionAgent
at a gateway, a FlowAgent at a hook); (c) the BPM go-to-market angle for
[enterprise-pitch.md](../enterprise-pitch.md). *Read the full paper before scoping.*

### 6 — Semantic Early-Stopping → **SPIKE (small), fold into `hillclimb_budget`**
Direct improvement to the sticky rework budget we shipped. Ours halts on
`max_rounds` + a confidence-delta / gap-count plateau; this paper halts on
**embedding cosine-convergence** (consecutive drafts stop changing in *meaning*,
patience window) — a more principled, **judge-free** plateau signal (no extra LLM
call → aligns with our cost goals), saving **38% tokens at parity**. The
**"it's not *when to stop*, it's *which round is best*"** oracle insight maps to
our **Judge skill** (keep-the-best-of-N, not just halt).
**Spike scope:** add a semantic-convergence signal to `hillclimb_budget`'s plateau
test; evaluate "keep best round" in the Judge skill; low-risk, small.

### 1 — The New SDLC with Vibe Coding (Google) → **Concepts (strategic), no spike**
Literally a chapter on *"Harness Engineering: what surrounds the model"* — the
industry framing of what we build. No mechanism to spike, but the vocabulary +
positioning are valuable: **"agentic engineering"** (AI under human-designed
constraints/tests = our governance), the **conductor vs orchestrator** split
(= our supervisor / sub-agent), and the economics (vibe = low-CapEx/high-OpEx vs
agentic = high-CapEx/low-OpEx = our governance-ROI argument).
**Action:** harvest the framing into `enterprise-pitch.md`; align our docs with
"agentic engineering" / "harness engineering" vocabulary.

### 5 — Prompt Injection in Résumé Screening → **Concepts (reinforcement)**
Domain-specific (hiring ATS), but the general finding — **LLM-judged selection is
manipulable by prompt injection** (effective when rare, collapses when common) —
is direct evidence for our **"no LLM for security/selection gates"** rule (we use
deterministic OPA; redaction is post-result, never authorization).
**Action:** cite as supporting evidence in the policy/authz docs; no spike.

### 3 — AURORA-AI → **Concepts (low priority)**
Closed-loop control ("react before degradation") + stability is an interesting
frame for *adaptive model routing under cost/SLA*, but the HJB/Lyapunov machinery
and the **fairness / demographic-parity** objective are mismatched with our
OPA-authz governance, and it targets a model *fleet*, not single-agent governance.
**Action:** park; revisit only if we build adaptive multi-model cost orchestration.

### 2 — TOPS (Visual Token Pruning) → **Skip**
Visual-token pruning for MLLM efficiency. We are text/code agents — out of domain.

## Recommended actions
1. **Write a spike** for **#4 (Agentic BPM / framed autonomy)** — architectural
   formalization + a new vertical.
2. **Write a small spike** for **#6 (semantic early-stopping)** — fold into
   `hillclimb_budget` + the Judge skill.
3. **Harvest concepts** from **#1** (positioning → enterprise-pitch) and **#5**
   (reinforces no-LLM-for-security).
4. **Park #3**, **skip #2**.

> **Unharvested follow-up from #1 (deep-read):** position **ARD + teo as the
> "context plane"** and **model-routing as a governed OpEx lever** — the
> whitepaper's "the model is ~10%, context + routing are the levers." Not yet in
> the pitch beyond the one-line hook.

## Cloned reference repos

Cloned `--depth 1` into `/home/bittahcriminal/air/workspace/research-repos/` (siblings of our repos, **not tracked here**) for hands-on analysis / the spikes' experiments:

| `research-repos/<dir>` | Spike / use | Stand-up demo |
| :-- | :-- | :-- |
| `agentgateway` | [agentgateway](eval-agentgateway-mcp-plane.md) | `cargo run -- -f examples/authorization/config.yaml` → UI `:15000/ui` |
| `ActPlane` | [ActPlane](eval-actplane-native-plane.md) | `make && sudo bash script/e2e_examples.sh` (live eBPF; Linux **5.8+ + BTF**, BPF-LSM for `block`) |
| `rlm` | [RLM](eval-rlm-long-context.md) | `pip install rlms && make quickstart` (needs `OPENAI_API_KEY`); visualizer `:3001` |
| `cuga-agent` **(branch `cugaflo`)** | [Agentic-BPM](eval-agentic-bpm-framed-autonomy.md) | `python docs/examples/flow_agent_app_inline/run.py loan_approval` (`applicant_id="4321"` → override) |
| `semantic-halting-problem` | [semantic-stopping](eval-semantic-early-stopping.md) | figures pre-committed; `cd backend && python experiments/run_experiment.py --split test --provider nvidia --max-rounds 6 --ablations` |
| `Prompt_Injection_ACL26` | concept (#5) | offline, no key: unzip `RESULTS.zip` → `python FIGURES_ALL.py` |

**No public repo (not cloned):** AURORA-AI (#3), TOPS (#2), Vibe Coding whitepaper (#1).
