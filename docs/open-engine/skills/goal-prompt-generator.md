# Goal Prompt Generator

Transforms a fuzzy implementation plan into a bounded, autonomous objective for an agent: a goal prompt with an explicit definition of done, repo constraints (what may be touched, what must not be), verification gates the agent must pass before claiming completion, and stop conditions for when to halt and ask rather than improvise. The output is a prompt you hand to a fresh agent session — or a different agent entirely — that can be pursued without supervision and checked cleanly afterward.

## Why Build It
The difference between an agent that works autonomously and one that wanders is almost entirely in how the objective is specified. "Definition of done + constraints + verification gates" is the specification pattern that works, and this skill makes producing it a procedure instead of an art. It's also the natural bridge between agents: one agent plans, this skill packages, another agent executes.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "goal-prompt-generator", stored
wherever my harness loads skills from.

The skill's job: turn an implementation plan or task description into a bounded goal
prompt another agent session can pursue autonomously and be checked against.

The skill must include: (1) trigger conditions — when I ask you to package work for
another session, write a goal prompt, or prepare a task for autonomous execution;
(2) a required structure for every goal prompt: the objective in one paragraph; an
explicit DEFINITION OF DONE as a checklist of verifiable statements; repo constraints
(files/areas that may be modified, files/areas that must NOT be touched); verification
gates — the exact commands to run and expected results before claiming completion; and
stop conditions — situations where the agent must halt and ask instead of improvising;
(3) a self-containment rule: the receiving session has none of our conversation
context, so the prompt must include exact paths and all needed background; (4) a
quality check before delivering: "could a competent agent with zero context execute
this and could I verify the result without re-deriving the plan?"

After writing it, test it by packaging the next real task I describe into a goal
prompt.
  </task>
</prompt>
```
