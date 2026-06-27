# Agentic Harness Designer

A design-review skill for building agent-powered products and systems: when the problem is "how should this AI system actually work," it walks the real architecture questions — tool-use design, permission and approval models, workflow state and durability, context and memory strategy, evaluation approach, observability, and operator visibility — and produces a phased implementation plan. It encodes the hard-won principle that most "AI product" problems are agent-system problems: the model matters less than the harness around it.

## Why Build It
If you build anything agent-powered — internal tools, products, even sophisticated personal automations — the failure modes live in the harness: missing approval gates, no durable state, no evaluation plan, no way to see what the agent did. A skill that forces those questions in order, every time, is the difference between a demo and a system. It's the most conceptual skill in the library, and for builders, frequently the most valuable.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "agentic-harness-designer", stored
wherever my harness loads skills from.

The skill's job: when I'm designing or reviewing an agent-powered system or product,
walk the real architecture questions and produce a phased plan — treating the problem
as an agent-SYSTEM problem, not a model-choice problem.

The skill must include: (1) trigger conditions — designing, evaluating, or debugging
any AI-agent-powered product, tool, or serious automation; (2) the design walk, in
order: what tools the agent gets and their exact contracts; the permission model
(what's autonomous, what needs approval, what's forbidden); workflow state and
durability (what survives a crash or restart); context and memory strategy (what the
agent knows, from where, and what it must not accumulate); evaluation (how we'll know
it works — concrete checks, not vibes); observability (what's logged, what the
operator can see mid-run); (3) failure-mode review against the common killers: missing
approval gates, non-durable state, unbounded context growth, no evals, invisible
execution; (4) output: a design doc with decisions and rationale, plus a phased
implementation plan where each phase is independently shippable and testable.

After writing it, test it by reviewing an agent system or automation I describe — or
one we've already built — and showing me the design doc.
  </task>
</prompt>
```
