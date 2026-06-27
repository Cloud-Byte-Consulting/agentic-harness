# Assumption Checker

Audits a plan, argument, or strategy doc for world-model problems: unstated assumptions, missing evidence, internal contradictions, and gaps between what the document claims and what it actually demonstrates. The output is a structured diagnostic — each assumption listed, rated by how load-bearing it is and how well-supported, with the single most dangerous assumption flagged.

## Why Build It
Agents are excellent at making plans sound coherent, which is precisely the danger. A dedicated adversarial pass — run as its own skill with its own posture, not as an afterthought in the same conversation that produced the plan — reliably catches the "we assumed the API does X" and "this only works if users behave like Y" failures before they cost you a week.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "assumption-checker", stored wherever
my harness loads skills from.

The skill's job: adversarially audit a plan, argument, or strategy document for
unstated assumptions, missing evidence, contradictions, and world-model gaps.

The skill must include: (1) trigger conditions — when I ask you to check, stress-test,
or red-team a plan or document; (2) a posture rule: in this mode you are a skeptic,
not a collaborator — do not soften findings or balance them with praise; (3) an output
format: each assumption stated plainly, rated for how load-bearing it is and how
well-evidenced, with the single most dangerous assumption called out at the top;
(4) a rule to check claims against the actual sources or code when they're available,
not just against internal consistency; (5) a closing section: the three questions that
would most reduce risk if answered.

After writing it, test it on any plan or doc I give you — or on one of your own recent
plans from this session.
  </task>
</prompt>
```
