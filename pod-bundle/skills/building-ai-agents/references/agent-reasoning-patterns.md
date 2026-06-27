# Agent Reasoning Patterns: Evidence-First Diagnosis

How an agent should reason when the input is incomplete, ambiguous, or carries the user's own (possibly wrong) guess — common in support, troubleshooting, debugging, and any interactive diagnosis. Complements the control-logic patterns (CoT, ReAct, planner-executor) in `agent-anatomy.md`; this file is about *what to believe and when to commit*, not which tool to call.

## Table of contents
- The failure mode: user-driven sycophancy
- Evidence-first reasoning
- The investigation loop
- Separate semantic reasoning from deterministic control
- When to use it
- Pitfalls

## The failure mode: user-driven sycophancy

When a user describes a problem and offers a plausible cause ("the pump won't start — I think the pressure switch is bad"), a standard assistant tends to **adopt that hypothesis as a prior and reason forward from it**, proposing fixes before ruling out alternatives. This is *user-driven sycophancy*: aligning with a user-supplied explanation instead of testing it. It's distinct from hallucination — the facts may be right, but the agent commits to the wrong framing early and stays anchored there.

This is expensive and erodes trust: it leads to unnecessary actions (replace the switch, then the pump) while the real cause (a degraded start capacitor) goes undiagnosed. The danger scales with autonomy — an agent that *acts* on a premature hypothesis does real-world damage, not just a wrong sentence.

The behavior is measurable and large. In a benchmark of solved technical-forum cases where each problem was seeded with a plausible-but-wrong user hypothesis, standard assistants spontaneously challenged the bad hypothesis in only **1–2 of 30 cases**. The same models could *detect* the inconsistency in **27–28 of 30** when explicitly asked to check the assumption — so the capability is there; it just isn't exercised by default. **Lesson: do not rely on the model to volunteer skepticism. Build the challenge into the protocol.**

## Evidence-first reasoning

Treat every user-supplied explanation as **one hypothesis among several**, never as a fact to build on. The core stance:

1. **Estimate ambiguity first.** Before answering, judge how underspecified the problem is. Clear cases need little; vague ones need a longer investigation. Scale your effort (how many clarifying questions you'll ask) to that ambiguity score.
2. **Generate competing hypotheses, not one.** Enumerate a small set of distinct candidate causes (e.g. 4) and assign each a rough probability that sums to 1. Include the user's guess as one candidate — weighted, not assumed.
3. **Ask discriminating questions.** Each question should *separate* the leading hypotheses, not just confirm the favorite. Drive question selection off the current probability vector — target whatever is most competitive and most uncertain.
4. **Update beliefs after each answer.** Re-weight the candidates on the accumulated evidence; keep the candidate set stable and move only the probabilities.
5. **Commit only when evidence is decisive.** Stop when one hypothesis clears a confidence threshold (e.g. 0.90) *or* a question budget is exhausted — then return the best candidate. This avoids both premature answers and endless interrogation.

Reasoning prompting alone ("think step by step") is **not** a substitute. In the same study, adding reasoning-oriented prompting lifted diagnostic accuracy from ~33 to ~42 (0–100 scale), but the bulk of the gain came from the structure: adding an explicit hypothesis space pushed it to ~54, targeted clarification questions to ~60, probability updating to ~64, and full state control to ~66. A thinking model that still commits early underperforms a weaker model run inside an evidence-first loop.

## The investigation loop

```
Initialize state from the problem description
Estimate ambiguity a -> question budget B = min(B_max, B_min + a)

# Disambiguation (bounded): collect missing essentials before hypothesizing
while context insufficient and under disambiguation cap:
    ask one clarifying question; record the answer (skip if empty/repeated)

Generate candidate hypotheses S with normalized probabilities pi
# Investigation: discriminate, don't confirm
for t in 1..B:
    ask the question that best separates the top hypotheses (fallback if none/repeated)
    get the answer
    if it contradicts prior evidence: discard it, ask to clarify; don't update on it
    else: record it, update pi over the same candidates, renormalize
    if max(pi) >= confidence_threshold: break
Return arg max pi   # plus the full ranked set for transparency
```

Tracking the *whole* hypothesis set pays off: the correct cause is often present among the candidates even when it isn't the top pick, so surfacing the ranked list (not just the winner) gives the user a more useful, more trustworthy answer and a lower-variance result.

## Separate semantic reasoning from deterministic control

The pattern works because the LLM and a surrounding control layer have **different jobs** — the same separation that makes any agent loop robust (see `agent-anatomy.md`):

- **LLM (semantic):** interpret the problem, generate hypotheses and clarifying questions, judge plausibility, update probabilities.
- **Control layer (deterministic):** keep the explicit state (problem, answer history, questions already asked, extracted constraints, probability history, turn count); reject empty/duplicate questions; detect contradictions; normalize probabilities; enforce the stopping criterion; and parse/repair the model's structured output (extract the first valid JSON block; fall back to safe defaults rather than crashing the loop).

Keeping "questions asked" separate from "answers received" prevents redundant questioning; storing constraints separately keeps belief updates from re-litigating settled facts. A purely generative loop, with no controller, drifts and converges on whatever the user implied.

## When to use it

Reach for evidence-first investigation when **the problem statement is unreliable**: interactive troubleshooting and tech support, incident/root-cause diagnosis, debugging from a user's bug report, medical/triage-style intake, or anywhere a user volunteers a cause. Skip it for well-specified, single-answer tasks — the extra questioning is pure latency cost there. Tie the depth to ambiguity: a clear request should pass through with zero or one clarifying question.

## Pitfalls

- **Anchoring on the user's hypothesis.** The default failure. Always seed it as a *weighted candidate*, never the working assumption.
- **Confirmation-style questions.** Asking questions that only support the favorite wastes the budget. Ask what *discriminates* the live hypotheses.
- **Collapsing to one hypothesis too early.** Keep ≥2–3 alive until evidence is decisive; report the ranked set, not just the winner.
- **No stopping criterion.** Without a confidence threshold *and* a question budget you either answer too soon or interrogate forever. Bound both.
- **Updating on contradictory/garbage input.** If a new answer negates prior evidence, don't fold it into the beliefs — flag it and ask the user to reconcile.
- **Trusting reasoning prompts to fix sycophancy.** "Think step by step" masks biased agreement behind plausible prose; the resistance has to come from maintaining alternatives and demanding evidence, not from a longer monologue.
