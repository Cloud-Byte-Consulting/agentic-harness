# Writing, Voice & Content - Open Skills | Unlock AI

Generated 2.39:1 Open Skills category masthead showing categorized skill drawers on the right side. [←Skills directory](/open-skills/skills) 

# Writing, Voice & Content

Skills that make agent writing sound like a specific person addressing a specific audience — instead of like an AI.

Use this category page to compare the primitives, check their requirements, and copy the setup prompt for the smallest skill that makes the next workflow repeatable.

Pick one primitive, install it, test it, then come back when your workflow teaches it a better default.

4 skills

### Personal Voice Skill

Encodes how you actually write — across contexts, not as a single tone preset. The skill captures your voice along multiple registers (direct/instructional, warm/relational, analytical, business-formal), with real samples of each, plus the rules of when to use which: when you're blunt, when you soften, what words and constructions you never use, how your emails differ from your posts. The agent then writes drafts that need light edits instead of rewrites.

Why build it

Generic AI prose is the most recognizable writing style on the internet right now, and "write this in a friendly tone" doesn't fix it. A voice skill built from your actual writing samples — with explicit anti-patterns ("never open with 'I hope this finds you well'", "never use 'delve'") — is the difference between an agent that drafts for you and one that drafts as you. This is consistently one of the highest-leverage skills for anyone who publishes or sends a lot of words.

What you need

5–10 samples of your real writing across different contexts (emails, posts, docs, messages)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "my-voice", stored wherever my
harness loads skills from.

The skill's job: write in my authentic voice across contexts — not a single tone
preset, but a model of how I actually write and when I shift registers.

Before writing it, ask me for 5–10 real writing samples across different contexts
(emails, posts, documentation, casual messages). Then analyze them and propose:
(1) my distinct registers (e.g. directive, relational, analytical, business) with what
distinguishes each; (2) sentence-level patterns I actually use; (3) anti-patterns —
words, openers, and constructions I never use, plus common AI-prose tells to
explicitly avoid; (4) rules for when to use which register based on audience and
stakes. Review your analysis with me before finalizing the skill.

The skill must include: trigger conditions (whenever I ask you to write, rewrite, or
review something in my voice), the register model with one short sample of each, the
anti-pattern list, and a rule that for technical content, accuracy beats voice — never
bend facts to sound like me.

After writing it, test it by drafting one short email and one short post on topics I
give you, and let me grade them.
  </task>
</prompt>
```

### New Release Briefing

When something significant ships in your field — a new AI model, a major tool release, a platform change — this skill turns gathered release data into a publish-ready briefing package: a structured summary of what actually changed, an analysis post in your voice, a standardized title/subtitle, and image prompts for a matching thumbnail. It assumes the research happened upstream (via Current-Information Search) and focuses on transforming raw release material into a publishable artifact with a consistent format readers learn to expect.

Why build it

Release-day content is a race where accuracy usually loses. A briefing skill encodes your quality bar — primary sources, dated claims, a fixed structure — so speed stops costing correctness. The consistent package format is the compounding part: your tenth briefing looks like your first, and your audience knows exactly what they're getting.

What you need [](/open-skills/core-infrastructure#current-information-search)  [](/open-skills/writing-voice-content#personal-voice-skill)  [](/open-skills/core-infrastructure#image-generation-gateway) 

Current-Information Search (or equivalent research input) · Personal Voice Skill makes the output dramatically better · Image Generation Gateway for thumbnails

Copy prompt

**Show the full setup prompt**

```
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

### Audience-Calibrated Content System

Generates content for a specific publication targeting a specific audience level — for example, a beginner-focused newsletter. The skill encodes the publication's content formats (e.g. a quick "snack," a concept explainer, a step-by-step tutorial), the audience's assumed knowledge floor and ceiling, banned jargon with required substitutions, and the weekly cadence. Given a theme, it plans and drafts a full content batch in the right voice at the right level.

Why build it

Writing down a sophistication level is much harder than it looks — expertise leaks in as unexplained jargon and skipped steps. Encoding the audience contract once (what they know, what they don't, what formats serve them) means every piece starts calibrated instead of needing a "make this simpler" revision pass. For anyone running a publication with a defined audience, this turns content production from artisanal to systematic.

What you need [](/open-skills/writing-voice-content#personal-voice-skill) 

A defined publication and audience · Personal Voice Skill if the publication has a named author voice

Copy prompt

**Show the full setup prompt**

```
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

### Branded Image Prompting Guide

A complete prompting guide for generating images in your visual brand — your colors, typography direction, composition style, and recurring formats (thumbnails, infographics, diagrams, photoreal scenes, UI mockups). It includes brand guidelines the agent applies automatically, techniques for both natural-language and JSON-structured prompting on current image models, a library of proven prompt templates for your common formats, and corrective prompting recipes for when models drift off-brand.

Why build it

Image models can hold a brand — but only if the brand is written down in prompt-shaped form. Without this skill, every image is a fresh negotiation and your visual output looks like ten different people made it. With it, "make me a thumbnail about X" returns something on-brand on the first try, and the prompt library compounds: every prompt that works gets added.

What you need [](/open-skills/core-infrastructure#image-generation-gateway) 

Image Generation Gateway (or any image model access) · Your brand basics (colors, type direction, visual references)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "branded-image-prompting", stored
wherever my harness loads skills from.

The skill's job: generate on-brand images by encoding my visual identity as prompting
guidance plus a reusable prompt library.

Before writing it, interview me for: my brand colors (hex), typography direction,
overall visual style (with reference images if I have them), and my most common image
formats (thumbnails, diagrams, infographics, social images, mockups).

The skill must include: (1) trigger conditions — any branded or recurring-format image
request; (2) brand guidelines in prompt-ready language the agent applies by default;
(3) both natural-language and JSON-structured prompt patterns for current image
models, with notes on when each works better; (4) a starter library of 10+ prompt
templates covering my common formats; (5) corrective prompting recipes for typical
drift (wrong colors, mangled text, off-style); (6) a rule to route actual generation
through my image-gateway skill and add successful prompts back to the library.

After writing it, test it by generating one thumbnail and one diagram in my brand and
let me judge them.
  </task>
</prompt>
```
 [](/open-skills/skills)  [](/open-skills/runbooks) 

Back to the Skills directory or continue into runbook compositions.