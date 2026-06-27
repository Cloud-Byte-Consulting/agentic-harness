# Audience-Calibrated Content System

Generates content for a specific publication targeting a specific audience level — for example, a beginner-focused newsletter. The skill encodes the publication's content formats (e.g. a quick "snack," a concept explainer, a step-by-step tutorial), the audience's assumed knowledge floor and ceiling, banned jargon with required substitutions, and the weekly cadence. Given a theme, it plans and drafts a full content batch in the right voice at the right level.

## Why Build It
Writing down a sophistication level is much harder than it looks — expertise leaks in as unexplained jargon and skipped steps. Encoding the audience contract once (what they know, what they don't, what formats serve them) means every piece starts calibrated instead of needing a "make this simpler" revision pass. For anyone running a publication with a defined audience, this turns content production from artisanal to systematic.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "audience-content-system", stored
wherever my harness loads skills from.

The skill's job: generate content for my publication, calibrated precisely to my
audience's level, in my established formats.

Before writing it, interview me for: the publication and its audience (who they are,
what they already know, what they definitely don't), my content formats (e.g. short
tip, concept explainer, tutorial) with length and structure for each, my publishing
cadence, and 2–3 examples of pieces that landed well.

The skill must include: (1) trigger conditions — planning or drafting anything for
this publication; (2) the audience contract: knowledge floor, knowledge ceiling,
banned jargon with plain-language substitutions; (3) a template per content format;
(4) a batch-planning mode: given a theme, propose a full week/cycle of pieces across
formats before drafting; (5) a calibration check before delivering any draft: "would
my least technical reader follow every step?"

After writing it, test it by planning one content batch on a theme I give you and
drafting the shortest piece from the plan.
  </task>
</prompt>
```
