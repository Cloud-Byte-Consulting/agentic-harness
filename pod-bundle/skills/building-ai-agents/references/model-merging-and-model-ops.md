# Model Merging & Model-Ops for Agents

How you get the *capabilities* an agent needs into the model(s) behind it — not just which
off-the-shelf model to call, but whether to merge several task-specialized models into one,
route between them at runtime, or fine-tune from scratch. This is the model-supply layer
under the "pick the model per agent" step in `SKILL.md` and the "choosing the model" section
of `agent-anatomy.md`.

## Table of contents
- When to read this (the decision)
- Decision rule: merge vs. route vs. fine-tune
- Task vectors and task arithmetic (the primitive)
- Static vs. dynamic merging
- Auto-FlexSwitch: learnable task-vector compression for cheap dynamic merging
- Practical guidance for agent builders
- Pitfalls

## When to read this (the decision)

Read this when an agent (or a fleet of agents) needs **several distinct skills** — e.g. a
support agent that does SQL, summarization, and policy lookup; or a multi-tenant assistant
where each customer has a fine-tuned variant — and you're deciding how to package those skills
into the model layer. The base SDK advice ("cheap model for triage, strong model for
reasoning") covers *which API model to call*. This file covers the case where the skills come
from **your own fine-tunes** and you must decide how to host and combine them.

## Decision rule: merge vs. route vs. fine-tune

Lead with this. Three options, in rough order of operational cost:

| Approach | What it is | Choose when |
|---|---|---|
| **Single merged model** | Fold N task-specific fine-tunes into one set of weights; serve one model | Tasks are related, you want one cheap endpoint, and you can tolerate some per-task accuracy give-up from interference |
| **Route between specialists** | Keep N fine-tuned models (or N task vectors), pick/compose per request | Tasks conflict, each must stay near its individual-fine-tune accuracy, and you can pay the storage/serving overhead — this is the *dynamic merging* regime |
| **Fine-tune one new model** | Train a fresh multi-task model on combined data | You have the training data, ground truth, and budget; you want a single artifact and accept the train cost and the need to retrain to add a skill |

Rules of thumb:
- **Don't fine-tune if a merge gets you there.** Merging is *data-light and training-light* — it operates on already-trained checkpoints. Auto-FlexSwitch trains its compression on only **N=100 exemplars per task** (no labels needed for the routing query set) for **500 optimization steps**, vs. the thousands of labeled steps a from-scratch multi-task model (MTL) needs. In the paper, building the merged system took ~1,400s on ViT-B/32 vs. ~17,600s to train MTL — roughly **13× cheaper** — while *beating* MTL accuracy (91.0% vs. 88.8%).
- **Merge (static) when tasks are friendly; route (dynamic) when they fight.** The failure mode of a single merged model is *task-vector interference* — gradients for different tasks point in conflicting directions and cancel. On the paper's 7-domain object-detection benchmark, static merging collapsed to **<4% mAP** because the domains were too far apart; dynamic merging recovered to **42% mAP**. If your tasks are that heterogeneous, route.
- **Fine-tune from scratch only when you must.** It's the most expensive path to maintain (retrain to add a skill) and brings the full build-vs-not-build checklist (`agent-anatomy.md`): data, ground truth, integration, cost. Treat it as the fallback, not the default.

For agents specifically: a **triage/router agent that selects among specialist models** is the
agentic analog of dynamic merging. Sometimes the right design is *not* one clever merged model
but a router agent over a few specialists (see `multi-agent-and-handoffs.md`). Merging is the
move when you want one served endpoint; routing is the move when you want isolation and
oversight. They compose: a merged model can be one specialist behind a router.

## Task vectors and task arithmetic (the primitive)

A **task vector** is the weight delta from fine-tuning: `τ_k = Θ_k − Θ` (fine-tuned weights
minus the pre-trained base), one per task `k`. This is the unit everything else is built on.

**Task arithmetic** is the observation that these vectors compose by simple linear algebra:
- *Add* task vectors to a base to install multiple skills at once: `Θ + Σ λ_k τ_k`.
- *Negate* a task vector to remove a behavior (un-learn a capability).
- *Scale* (`λ_k`) to dial a skill's strength up or down.

Static merging methods are mostly recipes for combining task vectors well: **Weight-Averaging**
/ Model Soups (mean of the fine-tunes), **Task-Arithmetic** (scaled sum), **TIES-Merging** and
**DARE** (sparsify the vectors first to reduce conflict), **Fisher / RegMean / AdaMerging**
(learn or compute per-parameter combination weights). They differ in how they fight
interference, but all produce one fixed weight set.

Two empirical properties of task vectors make compression and dynamic merging cheap — both
established in the paper on CLIP-ViT-B/32 across 8 vision tasks:
1. **Impulse-like activation (sparsifiable).** Only the high-magnitude entries carry the
   task-specific knowledge. Pruning the smallest entries doesn't just preserve accuracy — it
   often *improves* it, with gains rising until the pruning ratio passes **~70%**. So most of a
   task vector is removable.
2. **Robustness to low-bit (quantizable).** Replacing surviving entries with just their **sign**
   (binary, ±1) and restoring scale via an L2-norm factor keeps accuracy. The gap between
   full-precision and binarized *shrinks* as sparsity rises. Worst case observed was a 2.55%
   drop (Cars at 10% pruning); at high sparsity the binarized version sometimes wins.

Together these mean a task vector ≈ a **binary sparse mask + a sign vector + one scalar**,
yielding ~**16×** storage reduction with negligible accuracy loss. This is the lever that makes
keeping many specialists (dynamic merging) affordable.

## Static vs. dynamic merging

- **Static merging** bakes one fixed weight set at build time. One model to serve, lowest
  inference cost, no routing. But a single point in weight-space can't satisfy conflicting
  tasks — accuracy degrades, badly when tasks are far apart.
- **Dynamic merging** keeps task-specific components and composes them *per input* at inference
  time. It sidesteps interference (each request emphasizes the relevant task vector) and on the
  paper's benchmarks matches or beats individual fine-tunes. The cost: you must **store a
  component per task** and pay a routing step. Naively this is brutal — the paper's dynamic
  baselines need **540 MB–5.2 GB** of extra storage on ViT (Twin-Merging, MoW-Merging,
  EMR-Merging). That storage blow-up is the problem the next section solves.

## Auto-FlexSwitch: learnable task-vector compression for cheap dynamic merging

`Auto-FlexSwitch` (Gao et al., "Auto-FlexSwitch: Efficient Dynamic Model Merging via Learnable
Task Vector Compression", arXiv:2604.28109) makes dynamic merging cheap enough to be the
default by **learning** how to compress each task vector. The pipeline has four parts:

- **T-Switch (the static recipe it improves on).** Decompose each task vector into three compact
  "switches": an **Activation Switch** (binary sparse mask = which params fire), a **Polarity
  Switch** (sign vector = direction), and a **Switch Knob** (one scalar = magnitude). ~16×
  smaller, performance preserved. Fixed sparsity rate, hard binarization, static L2 scaling.
- **FlexSwitch (the learnable upgrade).** Replaces T-Switch's fixed rules with end-to-end
  optimization per model module:
  - **LGS (Learnable Gating Sparsification)** — a learnable magnitude threshold + temperature-
    controlled sigmoid gives differentiable sparsification, plus a learnable scale. Reaches
    **~97% average sparsity** while matching or beating full fine-tuning, where the fixed-rule
    baseline (P-Spar) collapses above 90% sparsity (>10-point DTD drop at 97%).
  - **BAS (Bit-width Adaptive Selection)** — learns the quantization bit-width per module from
    `{1,2,4,8}` instead of forcing 1-bit everywhere. Sensitivity is heterogeneous: attention
    in/out-proj and MLP fc/proj weights, and deeper layers, tolerate less compression — so
    uniform bit-width wastes budget. BAS also *relaxes* LGS's sparsity ceiling, so the two
    together land at lower storage than either alone.
  - **SASS (Sparsity-Aware Storage Strategy)** — grouped COO encoding that adaptively picks the
    group size from the actual sparsity ratio. Beats naive per-element storage once sparsity
    >0.5; at 0.98 sparsity it uses **<1/8** the bits, and vs. the original full-precision weights
    hits **>40×** compression past 0.9 sparsity, **>200×** at 0.99 (1.52 MB on ViT).
- **Auto-Switch → Auto-FlexSwitch (the routing).** Training-free dynamic merging via
  inference-time retrieval: build a small query set of feature vectors per task (input examples
  only, **no labels**), KNN-match each input's features against it, and weight the task switches
  by neighbor counts (Eq. 5). Auto-FlexSwitch adds a **learned low-rank metric** (rank r≈32) so
  the KNN runs in a discriminative low-dim space — more accurate routing *and* lower, flatter
  inference cost (~2.1× a single forward pass, stable as the query set grows, vs. Auto-Switch's
  2.5–3.4×).

**Headline results (faithful to the paper):**
- **ViT-B/32, 8 tasks:** Auto-FlexSwitch hits **91.16% avg** (beats individual fine-tunes at
  91.01% and MTL at 88.78%) using **~41–51 MB** of task-vector storage — vs. 540 MB–2.2 GB for
  the dynamic baselines (a **~10–50×** storage cut).
- **ViT-L/14:** up to **6.9×** smaller than its own uncompressed Auto-Switch variant.
- **Object detection (DETR, 7 domains):** **42.40% mAP** at 46.97 MB vs. Twin-Merging's 38.05%
  at 313.98 MB — better accuracy at **<15%** the storage; static merging was <4%.
- **NLU (GLUE on RoBERTa-base / Mamba-130M):** best avg accuracy at **3–17%** of the storage of
  other dynamic methods.
- **As a plain fine-tune-compression tool (no merging):** FlexSwitch stored Llama-3.2-3B-Instruct
  fine-tuned deltas in **5.21 MB (0.04%** of original) and Gemma-2-9b-it deltas in **731.57 MB
  (2.08%)**, each *slightly beating* the uncompressed fine-tune on reasoning/code benchmarks.

The transferable lesson: **fine-tuned deltas are extremely compressible**, so "keep a cheap
specialist per task and switch among them at runtime" is now a viable serving strategy, not a
storage nightmare.

## Practical guidance for agent builders

- **Start by asking whether you even need custom weights.** Prompting + tools + RAG covers most
  agent skills without touching model weights. Merging/fine-tuning is for when behavior must be
  *in the model* (latency-critical skills, offline/edge deploys, capabilities that resist
  prompting). Don't reach here first.
- **If you have multiple fine-tunes and need one endpoint:** try **static merging first** (cheap,
  one artifact). Sparsify the task vectors before merging (TIES/DARE-style) to cut interference.
  Validate per-task accuracy — if any task drops below its bar, escalate to dynamic.
- **If static merging loses too much per-task accuracy:** go **dynamic** (route/compose per
  request). Compress the per-task components so the storage stays sane — the Auto-FlexSwitch
  result shows you can keep dozens of specialists in tens of MB. For an *agent*, the simplest
  realization is a **router agent over specialist models/adapters**, which also gives you the
  oversight and isolation a single merged blob can't (`multi-agent-and-handoffs.md`).
- **Adding a new skill?** Merging/dynamic switching lets you *append a task vector* without
  retraining the others — much cheaper than re-running a from-scratch multi-task fine-tune.
  Negation lets you *remove* a skill or behavior. This composability is the main operational win
  over monolithic fine-tuning.
- **Compress fine-tuned deltas even when you're not merging.** If you ship many fine-tuned
  variants (per-customer, per-locale), storing sign+mask+scale deltas instead of full weights is
  a large, near-free storage win — the LLM-delta results above are the existence proof.
- **Budget the routing cost.** Dynamic merging adds an inference-time retrieval step (~2× a
  forward pass in the paper). For a latency-critical triage agent, that may argue for a single
  static model or a cheap classifier-based router instead of feature-KNN.
- **Match the model to the task per agent regardless.** None of this overrides `agent-anatomy.md`:
  triage/guardrail agents still want a cheap fast model; reasoning steps still want a strong one.
  Merging is about how you *build* a specialist, not a license to point every agent at one model.

## Pitfalls

- **Treating merging as free accuracy.** A single merged model trades per-task peak for
  one-endpoint convenience. On heterogeneous tasks the trade is severe (the <4% mAP collapse).
  Always measure per-task, not just the average.
- **Uniform compression across modules.** Sensitivity is heterogeneous — attention/MLP weights
  and deep layers tolerate far less sparsification/quantization than shallow layers. A global
  sparsity rate or bit-width either wastes storage or tanks sensitive modules; adaptive
  per-module allocation (the BAS/LGS idea) is why the learnable approach beats fixed rules.
- **Forgetting dynamic merging's storage tail.** The accuracy is great; naive implementations
  cost hundreds of MB to GBs per model. If you go dynamic, you *must* compress the components or
  the serving footprint kills the design.
- **Over-pruning the task vector.** Sparsity helps up to a point (gains until ~70% in the paper),
  then accuracy falls. There's a task-dependent sweet spot; don't crank sparsity blindly.
- **Fine-tuning when a merge would do.** The most expensive option by lifecycle cost (retrain to
  add skills, full data/ground-truth burden). Reserve it for when you genuinely need a new
  capability that merging existing checkpoints can't supply.
