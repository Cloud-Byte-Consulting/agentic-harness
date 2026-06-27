# Weekly Signal Diff

On a recurring basis (weekly is the natural cadence), reviews a defined set of inputs — your notes, a folder, feeds, project state, saved searches — and reports only what meaningfully changed since the last run: new signals, shifted assumptions, dead threads, emerging patterns. The skill keeps a small state file recording what it saw last time, which is what makes a true diff possible instead of a weekly summary that repeats itself.

## Why Build It
The hard part of staying current isn't gathering information, it's noticing change. A diff against last week's state surfaces exactly the delta — what's new, what moved, what quietly died — and ignores the stable background. This is also a gentle introduction to stateful skills: the state file pattern (skill remembers its last run) unlocks a whole class of recurring workflows.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "weekly-signal-diff", stored wherever
my harness loads skills from.

The skill's job: when I run it, compare a defined set of inputs against the state from
its last run and report only meaningful changes — new signals, shifted assumptions,
threads that died, patterns emerging.

Before writing it, interview me for: which inputs to watch (folders, notes files,
topics to search, project states), what counts as "meaningful" in my work, and where
the report should go.

The skill must include: (1) a state file the skill maintains, recording what it
observed each run, so diffs are real rather than re-summaries; (2) the input list and
how to check each one; (3) an output format ordered by importance of change, not by
source; (4) a rule that no-change is a valid and short answer — never pad a quiet
week; (5) a closing section suggesting at most three follow-ups based on the diff.

After writing it, do an initial baseline run to populate the state file, and tell me
what you recorded.
  </task>
</prompt>
```
