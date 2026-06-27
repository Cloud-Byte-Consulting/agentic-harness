# Branded Image Prompting Guide

A complete prompting guide for generating images in your visual brand — your colors, typography direction, composition style, and recurring formats (thumbnails, infographics, diagrams, photoreal scenes, UI mockups). It includes brand guidelines the agent applies automatically, techniques for both natural-language and JSON-structured prompting on current image models, a library of proven prompt templates for your common formats, and corrective prompting recipes for when models drift off-brand.

## Why Build It
Image models can hold a brand — but only if the brand is written down in prompt-shaped form. Without this skill, every image is a fresh negotiation and your visual output looks like ten different people made it. With it, "make me a thumbnail about X" returns something on-brand on the first try, and the prompt library compounds: every prompt that works gets added.

## What You Need


## Prompt / Setup
```xml
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
