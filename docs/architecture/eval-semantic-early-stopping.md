# Evaluation plan — semantic early-stopping for the hill-climb / Judge loop

**Type:** Small spike (~3–5 days). **Owner:** TBD. **Status:** Proposed.
**Paper:** *Semantic Early-Stopping for Iterative LLM Agent Loops: A Judge-Efficient
Study of When to Halt* (Shrivastava; arXiv 2606.27009, Jun 2026).
**Informs:** [research-triage.md](research-triage.md) #6; the `hillclimb_budget`
nessie policy + the Judge skill (from the role-router → Omnigent-supervisor port).

## 1. The decision this spike informs

Should we (a) add an **embedding-convergence plateau signal** to `hillclimb_budget`,
and (b) adopt the **"keep the best round"** pattern in the Judge skill?

Outcomes: **ADOPT-BOTH** · **ADOPT-PLATEAU only** · **NO-GO** (keep the current
counter + confidence-delta plateau).

## 2. Background — direct overlap with what we shipped

We already cap the refine/verify loop with **`hillclimb_budget`**: a sticky
per-intent budget that stops on `max_rounds` **or** a *plateau* (no improvement for
`max_flat_rounds`, where improvement = fewer gaps **or** a confidence gain
> `min_confidence_delta`).

This paper studies the same problem (when to halt a Writer→Critic loop) and finds:
- replacing `max_iterations` with a **semantic stopper** — halt when consecutive
  draft **embeddings stop changing in meaning** (cosine distance < ε for *k*
  rounds) **and** a quality signal stops improving — cuts **operational tokens 38%
  at parity quality**, and the **judge-free** variant is best (a per-round LLM judge
  *dominates cost* — a warning for us);
- the sharpest finding: **"when to stop" is easy; "which round is best" is the open
  problem** — an oracle that keeps the best round beats every practical stopping
  policy by a wide margin (+0.115 Information Score). I.e. *halting early is cheap;
  not losing the best intermediate answer is where the value is.*
- an honesty note: they **retracted** a Banach-contraction convergence claim —
  semantic non-expansiveness is a *measured conjecture*, not a theorem. Treat
  convergence as empirical, don't over-claim.

**Mapping to us:**
- `hillclimb_budget` plateau (`confidence-delta` / `gap-count`) **← add** an
  embedding cosine-convergence signal (more principled, content-aware, judge-free).
- the Judge skill (currently: verify + loop until clean) **← add** "keep the best
  round seen," not just "stop when it stops improving."

## 3. Hypotheses

- **H1:** an embedding-convergence plateau halts our loop at least as early as the
  confidence-delta plateau, with fewer wasted rounds, and **without an extra LLM
  call** (a small/local embedding model or cached embeddings).
- **H2:** keeping the **best** round's output (by the Judge's quality signal)
  beats returning the *last* round on our refine tasks.

## 4. Make-or-break questions

| # | Question | Gate |
| :-- | :-- | :-- |
| Q1 | Does an embedding cosine-distance + patience signal catch plateaus our `confidence-delta` test misses (or sooner)? | Whether to add it. |
| Q2 | Is it **judge-free** (no extra LLM call) — small embedding model / cache? | Cost must not regress. |
| Q3 | Does "keep best round" beat "return last round" on our tasks? | Whether to change the Judge. |
| Q4 | Does it stay **behavior-preserving** for the existing budget (sticky, cross-turn, fail-safe)? | Don't regress what we shipped. |

## 5. Experiments

> Prereq: the `hillclimb_budget` policy + a representative refine/verify loop (the
> supervisor's cross-review/Judge loop). An embedding source (a small local model or
> a cached embedding endpoint).

- **E1 — semantic plateau (Q1/Q2).** Add a cosine-distance-between-consecutive-
  verify-outputs signal with a patience window *k* to `hillclimb_budget`'s plateau
  test (as an **additional** STOP trigger, not a replacement). Replay a set of
  refine loops; compare rounds + tokens vs the current confidence-delta plateau.
  Confirm the embedding step adds no LLM call (measure cost). Output: rounds/token
  table, current vs semantic vs combined.
- **E2 — keep-best-round (Q3).** Modify the Judge/supervisor to retain the best
  round's output (by the Judge verdict / Information-Score-style signal) and return
  *that* on halt, not the last round. Measure quality delta. Output: best-vs-last
  quality comparison.
- **E3 — regression (Q4).** Confirm the existing `hillclimb_budget` semantics
  (sticky terminal, cross-turn closure state, fail-safe, the 87-test suite) still
  pass with the added signal. Output: green test run.
- **E4 — honesty guardrail.** Document the embedding-convergence signal as an
  **empirical** stopper (per the paper's retraction), not a proven contraction; a
  hard `max_rounds` failsafe stays the backstop.

## 6. Success criteria / decision gates

- **ADOPT-BOTH** if E1 shows the semantic plateau halts as-early-or-earlier at no
  extra LLM cost **and** E2 shows keep-best beats keep-last **and** E3 stays green.
- **ADOPT-PLATEAU only** if E1 wins but E2's keep-best is marginal/complex.
- **NO-GO** if the embedding signal adds cost or doesn't beat the current plateau.

## 7. Risks

- **Hidden cost** — an embedding call per round must be cheap/local, or it defeats
  the purpose (the paper's own warning about per-round judging cost). Q2 gates this.
- **Over-claiming convergence** — keep it empirical + the `max_rounds` failsafe.
- **Scope** — this is additive to a shipped policy; keep the diff small.

## 8. Deliverables

A rounds/token comparison (E1), a best-vs-last quality result (E2), a green
regression (E3), and a **short adopt/partial/no-go note** — with the one-paragraph
patch to `hillclimb_budget` + the Judge skill if ADOPT.

## 9. Repo deep-dive — grounded specifics

Clone: `/home/bittahcriminal/air/workspace/research-repos/semantic-halting-problem`. Reproduce the 38% result (from `backend/`): `python scripts/build_dataset.py --n 80 --dev 20` → `python experiments/run_experiment.py --split test --provider nvidia --max-rounds 6 --ablations` → `python experiments/make_figures.py --run-id test_nvidia_mr6`. **Figures are already committed** (`backend/results/test_nvidia_mr6/figures/`) → instant demo, no run needed. Embeddings are **always local** (no key); only the agent loop needs an LLM key. (Frontend has **no committed `package.json`** → not turn-key; lean on committed figures + the FastAPI backend.)

- **E1 — exact signal to fold into the plateau test:** embed each verify-output with **`BAAI/bge-small-en-v1.5`** (local, 384-dim, **0 tokens**), `d_t = 1 − cos(e_t, e_{t-1})`; STOP when `d_t < ε` for `k` consecutive rounds. **Defaults: ε=0.06, patience k=2** (`backend/shp/config.py:79,83,86`; distance `semantic_entropy.py`; cascade `halting.py`). Add as an **extra STOP trigger before the `max_rounds` failsafe**.
- **Cascade order (my draft had it wrong):** real priority is **critic_approved → entropy_convergence → no_information_gain → failsafe** (critic is checked *first*). Our entropy-plateau is the content-aware sibling of the existing `confidence-delta`/`gap-count` plateau.
- **E2 judge-free confirmed (the −38% is real):** `entropy_only` **6 862.9 op-tokens vs `fixed_k6` 11 070.2 → −38.0%** at parity (ΔIS −0.0037, p=0.81; `results/test_nvidia_mr6/summary.json`). `EntropyOnlyPolicy.uses_judge=False` — one local embed pass/draft, no API.
- **DON'T replicate the full SHP cascade — it's the trap our spike warns about.** Turning on the per-round quality (judge) signal = **+129% tokens for no quality gain**. Keep our loop's stop decision judge-free (entropy + failsafe only).
- **E2 keep-best — sharper, conditional.** `OraclePolicy` = `argmax_i quality[i]`, +**0.115 IS** vs every practical policy — **but `uses_judge=True`** (needs *every* round scored). **Cheap for us only if our Judge already emits a per-round verdict we can rank** → frame E2 as "**return the round with the best verdict, not the last**," NOT "add per-round scoring" (that inherits the +129%). Best-round *identification* is the paper's stated open problem — the oracle is an upper bound, not a deployable online stopper.
- **HotpotQA caveat bounds the win:** on short answers iteration barely helps (`fixed_k1` cheapest **and** best). Our refine/verify loops are longer-form → likely **more** keep-best headroom, but E2 must measure on *our* tasks.
- **E4 honesty (grounded):** the **Banach-contraction claim was retracted**; only termination/well-definedness/halt-priority are machine-checked (`shp/theory_checks.py`). Non-expansiveness is a measured conjecture — only **5% of trajectories strictly monotone**, mean slope **−0.0092**. That noise is *why* `k=2` exists (tolerates one fluke); keep `k≥2` + the hard `max_rounds` failsafe as the only proven backstop.
- **Net:** E1 plateau = clean **ADOPT** (judge-free, −38%, additive). E2 keep-best = **conditional** → most likely **ADOPT-PLATEAU-only** unless our Judge's existing verdict ranks rounds for free.
