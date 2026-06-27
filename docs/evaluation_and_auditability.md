# Evaluation and Auditability for Agent Orchestration 🧪🛡️

Measuring orchestration changes credibly, and making runs inspectable enough that a result can be trusted, reproduced, and debugged. This deep dive is about the **evaluation harness around** a multi-agent system, not the agents themselves.

---

## 1. Attribute Gains to the Architecture, Not the Model

**Pattern: hold the backbone fixed.** When you change orchestration — adding subagent fan-out, swapping a single loop for a recursive harness, introducing a propose/implement split — run the new design and the baseline on the *same model* at the *same decoding settings* (temperature, top-p, max tokens, seed). Then any score difference is attributable to the architecture rather than to a stronger model.

*   **Compounding vs. Substituting:** Run the same orchestration on at least two backbones of different strength. If the absolute score rises with the stronger model while the architecture's relative contribution holds or grows, the design compounds — that is the result worth shipping, because it does not become obsolete with the next model release.
*   **Family Coupling:** If your scorer, extractor, or judge shares a model *family* with the system under test, you have a subtle confound: the same family may agree with itself in ways that inflate the score. Cross-check with a judge from a different family where possible.

---

## 2. Report Uncertainty, Not Point Estimates

**Pattern: pair every headline number with a confidence interval.** A single percentage is not a result; it is a point estimate of a random quantity. Report a 95% interval alongside it. The cheapest defensible method is the **bootstrap**: resample your per-instance scores with replacement many times, recompute the aggregate metric on each resample, and take the 2.5th and 97.5th percentiles of the resulting distribution.

```python
import numpy as np

def bootstrap_ci(per_instance_scores, n_boot=10_000, alpha=0.05, seed=0):
    rng = np.random.default_rng(seed)
    scores = np.asarray(per_instance_scores, dtype=float)
    n = len(scores)
    means = np.empty(n_boot)
    for b in range(n_boot):
        idx = rng.integers(0, n, size=n)      # resample with replacement
        means[b] = scores[idx].mean()
    lo, hi = np.percentile(means, [100 * alpha / 2, 100 * (1 - alpha / 2)])
    return scores.mean(), (lo, hi)
```

*   **Small-n Buckets are Trends, Not Points:** When a category bucket holds only a handful of instances, its score has a huge interval. Read these as trends rather than absolute performance caps.
*   **Watch for Metric Artifacts:** Some scoring functions distort the picture in ways that have nothing to do with agent quality (e.g., compounding penalties on continuous outputs, or micro vs. macro averaging skewing representation).

---

## 3. Normalize, Then Score

**Pattern: split scoring into a normalization stage and a deterministic scoring stage.** Raw agent output is rarely in the exact comparison format. Insert a **normalizer** whose only job is to map raw output to the canonical answer format, then score the normalized value deterministically.

```
raw agent output
      │
      ▼
┌─────────────────────────┐   LLM normalizer extracts the answer in the exact
│ normalize (LLM)         │   required format ("spam", a number, an id).
│  └─ regex fallback      │   Deterministic regex fallback if the call returns empty.
└─────────────────────────┘
      │  normalized value
      ▼
┌─────────────────────────┐   exact match / numeric formula / set-F1.
│ score (deterministic)   │   NO model judgment here — pure computation.
└─────────────────────────┘
```

*   **Why two stages:** Normalization is a formatting problem (LLMs are good at this, and variance is bounded). Scoring is a correctness problem and must be deterministic and inspectable. Mixing them opens the door to inflated, irreproducible numbers.

---

## 4. Separate Deterministic Computation from LLM Judgment

**Pattern: push everything that *can* be exact into deterministic code, and let the LLM handle only the genuinely open-ended part.**
*   **Deterministic primitives:** Graph traversals, set operations, exact-match scoring, numeric formulas, schema validation, identifier joins.
*   **LLM judgment:** Synonym handling, multi-step planning, open-ended reasoning, fuzzy semantic matching, novelty assessment.

The agent supplies the open-ended glue; the primitives supply the verifiable substrate. Crucially, the verifiable parts stay verifiable no matter how creative the agent gets.

---

## 5. Secure Evaluators and Anti-Reward-Hacking

An agent optimizing against a metric will exploit any weakness in how that metric is computed (tampering, hidden test sets inspection, grader gaming). Treat your evaluator as an adversarial boundary.

*   **Hide the Grader Behind an Interface:** Keep the evaluation script and test data *outside* the agent-visible workspace. Expose only a submit-and-receive-score service.
*   **Controller-Owned Score Files:** The system—not the agent—writes the official score and ranking files. Block agent writes to those paths.
*   **Sandbox Isolation:** Run each agent session inside a container with only its workspace mounted.
*   **Sibling Isolation:** Prevent parallel fanned-out runs in the same round from inspecting or copying each other.

---

## 6. Auditability: Typed Contracts & On-Disk Manifests

**1. Typed Output Contracts:** Coordinator dispatches work using a typed schema. The aggregator validates worker output against its contract before accepting it. Failed contracts become a routable signal (retry/replan) rather than a silent text error.

**2. On-Disk Manifests:** The output is a manifest with on-disk artifacts, recording: stable job IDs, roles, statuses (success/failure), output paths, evidence IDs, and verbatim error logs. This enables:
*   **Selective Replay:** Re-run a single failed worker without replaying the entire session.
*   **Computable Failure Isolation:** Read the affected sub-tree off the dependency graph rather than parsing chat logs.
*   **Provenance to Evidence:** Trace any claim back to its source reference.

---

## 7. Pre-Launch Evaluation Checklist

Before publishing a number or shipping an orchestration change, verify the following:

### Attribution
*   [ ] Baseline re-run on the *same backbone, version, and decoding settings* as the new design.
*   [ ] Same extraction/normalization/scoring path used for both arms.
*   [ ] Model ID, version string, temperature, and sampling params logged.
*   [ ] If claiming the design compounds with model quality, evaluated on $\ge$ 2 backbones.

### Uncertainty
*   [ ] Every headline number paired with a 95% bootstrap CI.
*   [ ] The *difference* over baseline bootstrapped; difference interval checked for excluding zero.
*   [ ] Small-n buckets labeled as trends, with wide intervals shown.

### Scoring Integrity
*   [ ] Normalization and scoring are separate stages; scoring is deterministic.
*   [ ] Normalizer has a regex fallback; path fired is logged.
*   [ ] Deterministic computation separated from LLM judgment.

### Evaluator Security
*   [ ] Grader and test data live outside the agent-visible workspace (submit-and-score API).
*   [ ] Result files are controller-owned and write-blocked to agents.
*   [ ] Runs are sandboxed; parallel sibling runs are isolated.

### Auditability
*   [ ] Inter-agent outputs validated against typed contracts.
*   [ ] Run produces on-disk manifest with stable job IDs and verbatim failure logs.
*   [ ] Dependency sets recorded for failure isolation.
*   [ ] State persisted to support resume-on-interruption.
