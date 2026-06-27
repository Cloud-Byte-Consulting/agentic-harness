# Essay Illustration Gallery

Takes a finished essay or long post and produces a complete illustration package: the agent reads the piece, selects ~15–20 image-worthy moments across the essay's full arc (not just the obvious opener), locks a single illustration style so every frame is visually consistent, generates the images, writes a short "why this moment" caption per frame, and assembles everything into a gallery page — plus a ready-to-paste social note announcing it in the author's voice.

## Why Build It
The hard part of illustrating an essay isn't generating images — it's editorial judgment (which moments deserve images) and consistency (twenty images that look like one artist made them). This skill encodes both. It's also a great study in multi-skill composition packaged as one skill: analysis, style-locking, generation, captioning, gallery assembly, and publishing in a single repeatable pipeline.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "essay-illustration-gallery", stored
wherever my harness loads skills from.

The skill's job: turn a finished essay into a consistent illustration gallery —
selecting the moments, locking one style, generating the images, captioning each, and
assembling a gallery page.

This skill composes my image-gateway skill for generation and my site-publisher skill
for publishing (if I ask for the gallery to go live).

Before writing it, interview me for: my preferred illustration style direction (e.g.
hand-drawn editorial cartoon, photoreal, watercolor — help me write a precise style
descriptor we lock per gallery), and how many frames a typical essay should get.

The skill must include: (1) trigger conditions — when I share an essay and ask for
illustrations, images, or a gallery; (2) moment selection: choose frames across the
FULL arc of the piece, each tied to a specific passage, with a one-line rationale;
(3) style lock: one detailed style descriptor prepended to every prompt so all frames
match; (4) per-frame captions explaining why that moment was chosen; (5) gallery
assembly as a single page (use my html-artifacts conventions); (6) a short
ready-to-paste social note announcing the gallery, in my voice if I have a voice
skill.

After writing it, test it on one essay with a reduced frame count (5–6 frames) first.
  </task>
</prompt>
```
