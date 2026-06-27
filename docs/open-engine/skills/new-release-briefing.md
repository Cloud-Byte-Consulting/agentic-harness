# New Release Briefing

When something significant ships in your field — a new AI model, a major tool release, a platform change — this skill turns gathered release data into a publish-ready briefing package: a structured summary of what actually changed, an analysis post in your voice, a standardized title/subtitle, and image prompts for a matching thumbnail. It assumes the research happened upstream (via Current-Information Search) and focuses on transforming raw release material into a publishable artifact with a consistent format readers learn to expect.

## Why Build It
Release-day content is a race where accuracy usually loses. A briefing skill encodes your quality bar — primary sources, dated claims, a fixed structure — so speed stops costing correctness. The consistent package format is the compounding part: your tenth briefing looks like your first, and your audience knows exactly what they're getting.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "release-briefing", stored wherever
my harness loads skills from.

The skill's job: turn gathered release data about a new model, tool, or platform
change into a publish-ready briefing package.

Before writing it, interview me for: where I publish (newsletter, blog, internal
doc), my audience's sophistication level, and my title/format conventions if I have
them.

The skill must include: (1) trigger conditions — when I say "brief me up on <release>"
or hand you release research to package; (2) a fixed package structure: what actually
changed (facts with dates and sources), why it matters for my audience, what to do
about it, a standardized title and subtitle, and 2–3 thumbnail image prompts matched
to the subject's brand colors; (3) a rule that every factual claim carries a date and
source, and unverified claims are labeled as such; (4) if I have a voice skill, write
the post through it; (5) a rule that this skill packages — if the research is missing
or stale, stop and run current-info search first.

After writing it, test it on the most recent significant release in my field.
  </task>
</prompt>
```
