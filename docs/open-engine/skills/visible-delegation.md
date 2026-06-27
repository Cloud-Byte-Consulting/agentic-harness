# Visible Delegation

Lets one agent orchestrate another while keeping the delegated work visible — running the delegate in a shared terminal session (tmux is the proven mechanism) instead of a hidden background process, so you can watch, interrupt, and course-correct in real time. The skill covers launching the delegate with a packaged goal prompt, monitoring its progress, when the orchestrator should intervene versus wait, and how results get verified on the way back.

## Why Build It
Multi-agent setups usually fail on supervision: hidden background agents drift for twenty minutes before anyone notices. Keeping delegation observable preserves the leverage of parallel agents without surrendering oversight — and watching how a delegated session goes wrong is also how you learn to write better goal prompts. Pairs directly with Goal Prompt Generator: one packages the work, this runs and supervises it.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "visible-delegation", stored wherever
my harness loads skills from.

The skill's job: delegate work to another agent session while keeping it visible and
supervisable — shared terminal sessions, never hidden background runs.

Before writing it, check that tmux is installed (install it if not) and confirm which
agent CLI(s) I use for delegate sessions.

The skill must include: (1) trigger conditions — when I ask you to delegate, run
something in parallel, or hand work to another agent; (2) the launch procedure: create
a named tmux session, start the delegate agent in it, and pass a goal prompt (built
with my goal-prompt-generator skill if available) — then tell me how to attach and
watch; (3) monitoring rules: check the session at sensible intervals, and define what
warrants intervention (stuck loops, scope drift, destructive commands) versus patience;
(4) a results protocol: when the delegate claims completion, run the verification
gates from the goal prompt yourself before reporting success to me; (5) cleanup —
sessions get closed, not abandoned.

After writing it, test it by delegating one small real task end to end while I watch.
  </task>
</prompt>
```
