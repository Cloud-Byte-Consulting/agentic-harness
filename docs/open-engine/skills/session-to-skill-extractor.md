# Session-to-Skill Extractor

A continuous-learning loop for your skill library: at the end of substantial work sessions, this skill reviews what happened and asks whether any pattern is worth preserving — a workflow you'd repeat, a hard-won API discovery, a debugging path, a decision procedure. When a pattern clears the bar (recurring, non-obvious, codifiable), it drafts the new skill or updates an existing one, then files it for your review. Your library grows out of your actual work instead of requiring dedicated authoring time.

## Why Build It
Every skill in this library started as a session where someone solved a problem and refused to let the solution evaporate. This skill automates that refusal. It has a high bar built in — most sessions yield nothing, and that's correct — but compounding is the point: six months of extraction produces a library shaped precisely like your work. This is the skill that makes all the other skills self-multiplying.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "session-to-skill-extractor", stored
wherever my harness loads skills from.

The skill's job: at the end of substantial work sessions, evaluate whether anything we
did is worth preserving as a new skill or an update to an existing one — and if so,
draft it.

The skill must include: (1) trigger conditions — when I say "wrap up," "anything worth
keeping?", or at the natural end of a session where we solved something non-trivially;
(2) a high extraction bar, stated explicitly: the pattern must be RECURRING (I'll
plausibly need it again), NON-OBVIOUS (a fresh session wouldn't just derive it), and
CODIFIABLE (it can be written as a procedure) — most sessions yield nothing, and
"nothing worth extracting" is a good answer; (3) a check against my existing skill
library first: if an existing skill covers 80% of the pattern, propose an update, not
a new skill; (4) drafts follow my skills' standard format with trigger conditions, and
land somewhere for my review — never silently into the live library; (5) a sanitize
rule: extracted skills generalize the pattern and strip project/client specifics,
which stay in repo-local runbooks.

After writing it, test it on THIS session: evaluate whether our setup work contains an
extractable pattern, and show me your reasoning either way.
  </task>
</prompt>
```
