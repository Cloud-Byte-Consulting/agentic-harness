# Evaluation and Auditability for Agent Orchestration

Measuring orchestration changes credibly, and making runs inspectable enough that
a result can be trusted, reproduced, and debugged. This deep dive is about the
*evaluation harness around* a multi-agent system, not the agents themselves.

## Contents

- [Attribute gains to the architecture, not the model](#attribute-gains-to-the-architecture-not-the-model)
- [Report uncertainty, not point estimates](#report-uncertainty-not-point-estimates)
- [Normalize, then score](#normalize-then-score)
- [Separate deterministic computation from LLM judgment](#separate-deterministic-computation-from-llm-judgment)
- [Secure evaluators and anti-reward-hacking](#secure-evaluators-and-anti-reward-hacking)
- [Auditability: typed contracts and on-disk manifests](#auditability-typed-contracts-and-on-disk-manifests)
- [LLM-as-judge when there is no clean metric](#llm-as-judge-when-there-is-no-clean-metric)
- [A pre-launch evaluation checklist](#a-pre-launch-evaluation-checklist)

---

## Attribute gains to the architecture, not the model

**Pattern: hold the backbone fixed.** When you change orchestration — adding
subagent fan-out, swapping a single loop for a recursive harness, introducing a
propose/implement split — run the new design and the baseline on the *same model*
at the *same decoding settings* (temperature, top-p, max tokens, seed where
available). Then any score difference is attributable to the architecture rather
than to a stronger model sneaking in.

This sounds obvious and is constantly violated. A new orchestration layer is
usually built and tuned against the latest model, while the baseline number you
are comparing to was measured months ago on something weaker. The "improvement"
then conflates two changes. The fix is mechanical: re-run the baseline yourself,
on your current backbone, under your current settings, before you claim a delta.

**Why it matters: compounding vs substituting.** Once the backbone is fixed you
can ask a sharper question — does the architecture *substitute* for model quality
(it only helps weak models and washes out on strong ones) or does it *compound*
with model quality (the same design keeps adding gains as the backbone improves)?
Run the same orchestration on at least two backbones of different strength. If the
absolute score rises with the stronger model while the architecture's relative
contribution holds or grows, the design compounds — that is the result worth
shipping, because it does not become obsolete with the next model release. A
recursive-harness study, for example, showed the same spawning design improving a
coding-agent baseline with the backbone held fixed, *and* reaching a higher score
when the backbone was upgraded — evidence the harness compounds rather than
substitutes.

**Practical consequences:**

- Treat published baselines as fixed reference points only if they were measured
  on the same data split and protocol you are using. If you do not have their
  per-instance scores, you can still bootstrap your own scores against their point
  estimate (see below), but say so explicitly.
- Use the same answer-extraction and scoring path for both arms. If your system
  routes raw output through an extraction step, the baseline must too, or the
  extraction becomes a hidden confound.
- Log the model id, the exact version string (e.g. a dated snapshot), temperature,
  and any sampling parameters into the run manifest. "A capable frontier model" is
  not a reproducible setting.

**Caveat — family coupling.** If your scorer, extractor, or judge shares a model
*family* with the system under test, you have a subtle confound: the same family
may agree with itself in ways that inflate the score. This is tolerable when the
scoring step is mechanically trivial (it only reformats already-correct output)
but becomes a real threat when the same family is also doing judgment. Note it as a
limitation and, where you can, cross-check with a judge from a different family.

---

## Report uncertainty, not point estimates

**Pattern: pair every headline number with a confidence interval.** A single
percentage is not a result; it is a point estimate of a random quantity. Report a
95% interval alongside it. The cheapest defensible method is the **bootstrap**:
resample your per-instance scores with replacement many times (10,000 resamples is
common and cheap), recompute the aggregate metric on each resample, and take the
2.5th and 97.5th percentiles of the resulting distribution.

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

**Does the interval exclude zero?** When you claim a gain over a baseline, the
honest test is whether the *difference* is distinguishable from noise. Bootstrap
the difference (per-instance paired delta if you have both arms' per-instance
scores; otherwise resample your own scores against the baseline's point estimate)
and check whether the 95% interval excludes zero. A "+9 points" headline whose
interval is [+4, +15] is a real effect; the same headline with an interval of
[-2, +20] is not yet a result. State the interval, not just the midpoint.

**Small-n buckets are trends, not points.** Evaluation suites are often stratified
into buckets (context-length bands, answer types, difficulty tiers). When a bucket
holds only a handful of instances, its score has a huge interval — a single wrong
answer can swing a 5-instance bucket by 20 points. Read these as *trends*: report
the per-bucket interval, and resist drawing conclusions like "the system degrades
at bucket X" when bucket X has n=5 and an interval spanning [20, 100]. The
aggregate over all buckets is the trustworthy number; the per-bucket breakdown is
diagnostic color.

**Watch for metric artifacts.** Some scoring functions distort the picture in ways
that have nothing to do with agent quality. Two recurring traps:

- *Compounding penalties on continuous outputs.* A scorer that decays
  multiplicatively with error magnitude penalizes small numeric errors far more
  than they deserve. *(Illustration from one source benchmark: a metric of the form
  `0.75^|y - ŷ|` for numeric answers penalizes an off-by-one down to 0.75 and
  off-by-two down to ~0.56.)* A counting subtask that is *reasoning-correct* but off
  by a small margin then bleeds score in a way that looks like a capability gap but
  is a scoring artifact. When a category underperforms, inspect whether the metric —
  not the agent — is responsible before you redesign anything.
- *Micro vs macro averaging.* A micro-average over all instances is dominated by
  the largest buckets; a macro-average over categories weights every category
  equally. Decide which question you are answering and label the aggregate
  accordingly. Reporting one while implying the other is a common sleight of hand.

---

## Normalize, then score

**Pattern: split scoring into a normalization stage and a deterministic scoring
stage.** Raw agent output is rarely in the exact comparison format — it may say
`The label is: spam.` when the gold answer is `spam`. Rather than make your
scorer tolerant (and therefore fuzzy and gameable), insert a **normalizer** whose
only job is to map raw output to the canonical answer format, then score the
normalized value deterministically.

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
      │
      ▼  score + which path was taken
```

**Why two stages.** Normalization is a *formatting* problem (LLMs are good at it,
and the variance is bounded because the underlying answer is usually already
present). Scoring is a *correctness* problem and must be deterministic and
inspectable. Mixing them — using an LLM to decide "close enough" — quietly turns
your scorer into a judge and opens the door to inflated, irreproducible numbers.

**Normalizer design:**

- Give it a fixed prompt that asks only for the final answer value in the exact
  format the task specifies — no explanation, no punctuation, no extra words.
- Add a **regex fallback** for when the call returns empty or malformed output, so
  the pipeline never silently drops an instance. Log which path fired; if the
  fallback fires often, your normalizer prompt needs work.
- Keep it dumb. The normalizer should not reason about correctness, only about
  shape. If raw outputs are *typically already* in the target format, the
  normalizer's influence on the final score is small — which is exactly what you
  want for a step you did not separately validate.

**Caveat — validate the normalizer or flag it.** If you never human-check the
normalization step, say so as a limitation. A normalizer that occasionally
"corrects" a wrong answer into a right-looking one is a silent scorer leak. The
risk is lowest when (a) raw outputs are already near-format and (b) the normalizer
shares no incentive with the system under test. The family-coupling caution
applies here too: a normalizer from the same family as the backbone may map
ambiguous output more charitably than an independent one would.

---

## Separate deterministic computation from LLM judgment

**Pattern: push everything that *can* be exact into deterministic code, and let
the LLM handle only the genuinely open-ended part.** This is the single most
important architectural choice for an inspectable system. Draw a hard line:

- **Deterministic primitives** — graph traversals, set operations, exact-match
  scoring, numeric formulas, schema validation, identifier joins. These are fast,
  reproducible, and verifiable. Anyone can re-run them and get the same answer.
- **LLM judgment** — synonym handling, multi-step planning, open-ended reasoning,
  fuzzy semantic matching, novelty assessment. These are where models earn their
  keep and where exact methods fall short.

When you keep the two separate, capabilities *compose*: a deterministic
seed-resolution primitive becomes semantic retrieval when an agent expands the
query into synonyms first; single-hop deterministic graph queries become multi-hop
workflows when the agent chains them with reasoning in between. The agent supplies
the open-ended glue; the primitives supply the verifiable substrate. Crucially,
the verifiable parts stay verifiable no matter how creative the agent gets.

**Apply the same split inside the evaluator.** A scoring example: deciding whether
two relation triples match is two distinct jobs. Normalizing relation names to a
canonical case and treating arguments as a case-insensitive set is *deterministic*.
Deciding whether "ConvNet" and "Convolutional Neural Network" are the same entity
is *judgment*. Implement the first in code; reserve the LLM for the second. The
deterministic core (e.g. set-level F1 over normalized tuples, with both-empty sets
treated as a perfect match) gives you a number you can defend line by line.

**Reward-shaping corollary.** When you compute a scalar reward to drive an agent
or a training loop, decompose it into a dense, cheap-to-compute deterministic
component (format compliance, valid JSON, schema conformance) plus the semantic
component (task F1, judged correctness). The deterministic terms are trivially
auditable and catch malformed output early; the semantic term carries the actual
signal. Keeping them as separate, named summands means you can inspect *which*
part of a reward a run earned, instead of staring at one opaque float.

---

## Secure evaluators and anti-reward-hacking

**Threat model.** A capable agent optimizing against a metric will exploit any
weakness in how that metric is computed. Reported failures include agents
contaminating evaluation data, modifying result files, inspecting hidden test
sets, and exploiting weak grading protocols to score points without solving the
task. Treat your evaluator as an adversarial boundary, not a trusted library call.
(This is the integrity/grader-gaming threat. It is distinct from *input*-side
attacks — prompt injection, malicious tool output — which are covered in
input-safety-and-guardrails.)

**Patterns that harden an evaluator:**

- **Hide the grader behind an interface.** Keep the evaluation script and any test
  data *outside* the agent-visible workspace. Expose only a submit-and-receive-score
  service: the agent can submit a candidate and read back an official score, but
  cannot read or modify the evaluator itself. The grader's logic and its data are
  not affordances the agent has.
- **Make authoritative result files controller-owned.** The system — not the agent
  — writes the official score and ranking files. Add hooks that *block* agent
  writes to those paths. If the agent could edit the file that records its own
  score, the score is meaningless.
- **Isolate the run.** Execute each agent run inside a sandbox (e.g. a container)
  with only its workspace mounted. Files outside the run are protected from
  accidental or adversarial modification, so a misbehaving agent cannot reach the
  grader, peer runs, or the host.
- **Default-deny scarce resources.** For contended resources (GPUs, external API
  quotas), make them invisible by default and hand them out through a helper that
  records ownership and enforces exclusivity. This prevents one agent from starving
  others and keeps resource usage attributable.
- **Enforce isolation between parallel siblings.** When you fan out multiple
  candidate sessions in the same round, prevent them from inspecting or copying
  each other. Same-round isolation stops premature collapse toward one local
  optimum and keeps each candidate an independent data point — which is also what
  makes your per-candidate statistics honest.
- **Constrain what the agent may submit.** A submission that "passes" by reproducing
  a leaked answer, hard-coding expected outputs, or special-casing the test
  harness is a reward hack. Where feasible, randomize benchmark order across
  measured rounds, run warmup rounds you discard, and re-grade external baselines
  under the *same* local protocol so comparisons are apples-to-apples.

**The framing that helps.** Think of the environment as defining *affordances*:
what actions are even possible shapes behavior more than instructions do. A
well-engineered evaluation environment removes high-risk affordances (evaluator
leakage, score tampering, resource contention, sibling copying) while keeping
productive ones (free exploration, tool use, access to prior results). You cannot
reliably suppress reward hacking with prompt instructions alone — the boundary has
to be structural.

---

## Auditability: typed contracts and on-disk manifests

A run you cannot inspect after the fact is a run you cannot trust or debug. Two
patterns make multi-agent runs auditable.

**1. Typed output contracts between agents.** When a coordinator dispatches work to
a worker, the worker's expected output should be a *typed schema*, not a free-form
prompt response. The aggregator validates each worker output against its contract
before accepting it. The payoff is that **a failed contract becomes a routable
signal rather than a silent textual error** — the system knows worker 7 produced
malformed output and can retry or replan, instead of passing garbage downstream
where it corrupts the final answer invisibly.

```
plan = coordinator(graph, task, mode)
# each job carries: role, payload, output_contract (a typed schema), deps
for job in plan:
    out = run_worker(job.role, job.payload)
    if not validates(out, job.output_contract):
        mark_failed(job)            # routable: retry with more context, or replan
    else:
        accept(out)
```

**2. On-disk manifests with dependency sets.** The result of a run should be a
*manifest with on-disk artifacts*, not a conversation history. A good manifest
records, per job: a stable job id, the role, status (success/failure), the output
artifact's path, the ids of the evidence the worker used, and any error messages —
**failures recorded verbatim, not hidden inside a summary.** This buys three
concrete capabilities:

- **Selective replay.** Re-run a single failed worker, or replace one evidence
  bundle, without replaying the entire session. Recovery cost scales with the
  failed sub-task, not the whole run.
- **Computable failure isolation.** Record each job's **dependency set**. When a
  job fails, the affected sub-tree is read directly off the dependency graph rather
  than reconstructed from a chat transcript. You know exactly what is now suspect.
- **Provenance to evidence.** Because the manifest stores which evidence id (which
  figure, paragraph, citation path, or source span) supported each output, a
  reviewer can trace any claim back to its source instead of taking the agent's
  word for it.

**Stable identifiers are the join key.** Across artifacts, views, and sessions,
key everything by stable, globally unique identifiers rather than by surface
strings. Joining two artifacts by id is a hash lookup; joining by name is a fuzzy
match that can falsely merge distinct entities that happen to share a string.
Identifier-based joins are both cheaper and safer, and they are what let you stitch
a coherent audit trail out of independently produced pieces.

**Persist enough to resume.** For long-running orchestrations, persist per-stage
session ids, status, elapsed time, and remaining budget to the filesystem. An
interrupted run should resume from the latest persisted state under the remaining
budget rather than restarting from scratch — and the same persisted state is what
makes the run inspectable mid-flight. (For the relationship between this
file-based pattern and off-the-shelf durable-execution / DAG engines, see
reliability-and-operations.)

---

## LLM-as-judge when there is no clean metric

Some qualities have no exact metric: semantic correctness of an extraction, novelty
of an idea, validity of a reasoning chain, whether a summary captures the central
contribution. Here an **LLM-as-judge** is the right tool — but it must be
disciplined, or it becomes a random number generator with a confident tone.

**Pattern: per-category judging policies.** Do not judge everything with one rubric.
Match the judicial standard to the nature of the field being evaluated:

- **Strict verification** for objective facts (metadata, identifiers, citations
  that must exist): zero tolerance for hallucination, validate each item
  individually, penalize fabricated references and factual errors. Accept only
  valid logical inferences (e.g. deriving a venue from an id pattern).
- **Semantic tolerance** for named entities: lenient on format, strict on facts.
  Accept synonyms and abbreviations ("ConvNet" for "Convolutional Neural Network");
  penalize omission of *primary* items but forgive trivial details.
- **Abstractive equivalence** for high-level claims (contributions, findings,
  limitations): reward semantic essence over wording. Accept a correct high-level
  summary of a specific numeric result; penalize only when the *central* claim is
  absent.

Encoding these as separate, explicitly stated policies makes the judge's behavior
predictable and lets a human disagree with a specific policy rather than with an
opaque global verdict.

**Pattern: rubric scores with named dimensions.** When judging a fuzzy quality,
have the judge score *named dimensions* on a structured rubric rather than emit one
gestalt number. For novelty, score problem formulation, algorithmic mechanism,
training strategy, and target domain separately. For a reasoning answer, score the
*reasoning trace* and the *final answer* as distinct components — this evaluates
traceability and rigor, not just whether the final token happened to match. The
per-dimension breakdown is itself an audit artifact: you can see *why* the judge
scored as it did.

**Ground the judge in evidence.** A novelty judgment over the idea text alone
conflates surface form with substance. Retrieve the most related prior work first,
then ask the judge to score overlap *against that retrieved evidence*. Grounding
turns a vibe check into a defensible comparison and makes the result auditable
against the same sources a human would consult.

**Caveats for any LLM-judge protocol:**

- Treat the judge as a measurement instrument: fix its model and version, use a
  low temperature, and record both in the manifest.
- Beware family coupling between the judge and the system under test (see the first
  section). Where stakes are high, validate against human labels on a sample, or
  use a judge from a different family.
- A binary 1/0 judge over reasoning-plus-answer is a reasonable default for QA-style
  tasks; reserve graded rubrics for tasks where partial credit is meaningful.
- Report judge-based metrics with the same uncertainty treatment as any other —
  bootstrap CIs, small-n caution.
- An LLM judge that *votes* across multiple samples or *deliberates* with a second
  judge is itself a collaborative pattern (self-consistency / debate) — useful for
  reliability, but every added call is cost; see collaborative-and-single-agent-patterns.

---

## A pre-launch evaluation checklist

Before you publish a number or ship an orchestration change, walk this list.

**Attribution**

- [ ] Baseline re-run on the *same backbone, version, and decoding settings* as the
      new design — not quoted from a stale measurement.
- [ ] Same extraction/normalization/scoring path used for both arms.
- [ ] Model id, version string, temperature, and sampling params logged to the
      manifest.
- [ ] If claiming the design compounds with model quality, evaluated on ≥2 backbones
      of different strength.

**Uncertainty**

- [ ] Every headline number paired with a 95% bootstrap CI.
- [ ] The *difference* over baseline bootstrapped; interval reported and checked
      for excluding zero.
- [ ] Small-n buckets labeled as trends, with their (wide) intervals shown.
- [ ] Aggregate averaging method (micro vs macro) stated and matched to the claim.
- [ ] Known metric artifacts (compounding penalties, etc.) identified before
      interpreting category dips.

**Scoring integrity**

- [ ] Normalization and scoring are separate stages; scoring is deterministic.
- [ ] Normalizer has a regex fallback; which path fired is logged.
- [ ] Everything that can be exact is computed deterministically; the LLM is used
      only for genuinely open-ended judgment.
- [ ] Reward (if any) decomposed into named deterministic + semantic components.

**Evaluator security**

- [ ] Grader and test data live outside the agent-visible workspace, exposed only
      through a submit-and-score interface.
- [ ] Authoritative result/score files are controller-owned and write-blocked to
      agents.
- [ ] Runs are sandboxed/isolated; scarce resources are default-deny with tracked
      ownership.
- [ ] Parallel sibling runs cannot inspect or copy each other.
- [ ] External baselines re-graded under the *same* local protocol.

**Auditability**

- [ ] Inter-agent outputs validated against typed contracts; contract failures are
      routable, not silent.
- [ ] Run produces an on-disk manifest (job ids, status, artifact paths, evidence
      ids, verbatim failures) — not just a chat transcript.
- [ ] Dependency sets recorded so failure isolation is computable.
- [ ] Artifacts joined by stable identifiers, not surface strings.
- [ ] Long runs persist enough state to resume and to inspect mid-flight.

**LLM-as-judge (if used)**

- [ ] Per-category policies defined and stated.
- [ ] Fuzzy qualities scored on named-dimension rubrics, not one gestalt number.
- [ ] Judge grounded in retrieved evidence where applicable.
- [ ] Judge model/version/temperature fixed and logged; family coupling noted;
      validated against human labels on a sample where stakes are high.

**Honesty**

- [ ] Limitations stated explicitly (unvalidated normalizer, single-suite
      evaluation, family coupling, uninstrumented cost/latency).
- [ ] No invented numbers; every reported figure traces to a logged run.
