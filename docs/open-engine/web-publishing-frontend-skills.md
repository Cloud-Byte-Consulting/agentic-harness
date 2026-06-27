# Web Publishing & Frontend - Open Skills | Unlock AI

Generated 2.39:1 Open Skills category masthead showing categorized skill drawers on the right side. [←Skills directory](/open-skills/skills) 

# Web Publishing & Frontend

Skills that take agent output public — with taste, verification, and a repeatable shipping procedure.

Use this category page to compare the primitives, check their requirements, and copy the setup prompt for the smallest skill that makes the next workflow repeatable.

Pick one primitive, install it, test it, then come back when your workflow teaches it a better default.

4 skills

### Frontend Taste System

Replaces your agent's default frontend instincts with a much stronger taste system: deliberate layout variance instead of the same hero-and-three-cards page, stricter component decisions, real typography, restrained color, and mandatory visual verification (screenshot, inspect, fix, repeat) before any frontend work is called done. Structurally, it's a bundle — a core skill with nested sub-skills for specific directions (minimalist editorial UI, data-dense dashboard UI, premium landing pages, mobile app concepts, redesigning existing projects) the agent loads as relevant.

Why build it

Agent-generated frontend has a recognizable look — and it isn't a compliment. The fix isn't "make it prettier" in the moment; it's a standing taste system the agent applies to every frontend task. The nested-bundle structure also teaches a key skill-architecture pattern: a core philosophy skill that routes to specialized sub-skills, instead of one unloadable mega-document.

What you need

Nothing required; a screenshot-capable browser tool (most harnesses have one) for visual verification

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill bundle for my AI coding agent called "frontend-taste", stored
wherever my harness loads skills from.

The job: replace your default frontend design instincts with a stronger taste system
that applies to all websites, apps, landing pages, and UI work.

Structure it as a core skill plus nested sub-skills. The core skill must include:
(1) trigger conditions — all frontend design and implementation work; (2) layout
rules: deliberate variance, no default hero-plus-three-cards pattern, real grids,
generous whitespace used intentionally; (3) typography rules: a real type scale,
restrained pairings, no default-stack sloppiness; (4) color rules: restrained
palettes, one accent doing real work, no purple-gradient-on-white clichés;
(5) mandatory visual verification: screenshot the result, inspect it critically, fix
what's weak, repeat before calling it done.

Create nested sub-skills for: minimalist/editorial UI, data-dense dashboard UI,
premium marketing/landing pages, and redesigning existing projects without breaking
them. The core skill routes to these based on the task.

Interview me first for my taste references — 2–3 sites or apps whose design I admire
and why. After writing the bundle, test it by building one landing page section and
running your own visual verification loop on it.
  </task>
</prompt>
```

### Personal Site Publisher

Publishes a finished page to your personal or company website as a real, share-ready URL — handling everything that separates "an HTML file" from "a published page": the design language, a clean slug and URL route, a page-specific Open Graph preview image (1200×630) so links unfurl properly, share title and description, indexing controls (public vs. unlisted), local verification before deploy, and the deploy itself. The skill encodes your site's stack and conventions so publishing is a procedure, not a project.

Why build it

The gap between "the agent made a great page" and "that page is live at a clean URL with a proper link preview" is where most agent web output stalls. Encoding the full publish path once — including the unglamorous parts like OG images and noindex flags — means anything your agent produces is one sentence away from shippable. This skill is the final step of half the runbooks in this library.

What you need

A website you control with a deploy path your agent can run (static site, framework site, or hosting CLI)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "site-publisher", stored wherever my
harness loads skills from.

The skill's job: take a finished page or artifact and publish it to my website as a
real shareable URL, end to end.

Before writing it, explore my website repo and interview me for: the repo path and
stack, how routes/pages are added, my deploy command and any verification steps, my
design language (or which existing pages to match), and my default indexing preference
for one-off share pages (public vs. unlisted).

The skill must include: (1) trigger conditions — ONLY when I explicitly ask to
publish/ship/put something on the site, never auto-triggered; (2) the full procedure:
clean slug, page creation matching site conventions, a page-specific 1200x630 Open
Graph image (route generation through my image-gateway skill if I have one), share
title and description, indexing controls; (3) local verification before deploy — build
and view the page; (4) the deploy procedure; (5) post-publish checks: live URL loads,
OG preview renders correctly.

After writing it, test it by publishing one unlisted test page end to end, then walk
me through cleaning it up or keeping it.
  </task>
</prompt>
```

### Image Model Comparison Arena

Builds and publishes comparison test pages for image-generation models: each model gets its own review page (same prompts, that model's outputs, cost and behavior notes), and all models share a side-by-side comparison viewer — all generated from a single config file. Adding a new model means adding a config entry and re-running; the skill handles generation, image optimization, page builds, and publishing. It maintains a registry of model costs and content-policy quirks discovered along the way.

Why build it

Beyond its direct use (genuinely useful model comparisons), this skill is the library's best example of composition as architecture: it doesn't generate images (it calls Image Generation Gateway) and it doesn't publish (it calls Personal Site Publisher). It owns exactly one thing — the comparison methodology and page generation — and delegates the rest. When a new image model drops, you can have a published, evidence-based comparison the same afternoon.

What you need [](/open-skills/core-infrastructure#image-generation-gateway)  [](/open-skills/web-publishing-frontend#personal-site-publisher) 

Image Generation Gateway and Personal Site Publisher built first · Budget for generation costs (typically a few dollars per model)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "image-model-arena", stored wherever
my harness loads skills from.

The skill's job: build and publish image-model comparison pages — one review page per
model plus a shared side-by-side viewer — generated from a single config.

This skill COMPOSES two skills I already have: image generation goes through my
image-gateway skill, and publishing goes through my site-publisher skill. It must
never reimplement either.

Before writing it, interview me for: my standard test prompt set (help me design 6–10
prompts covering photorealism, text rendering, diagrams, people, and style range), and
where comparison configs and generated images should live.

The skill must include: (1) trigger conditions — when I want to test a new image
model, compare models, or add a model to an existing comparison; (2) a single config
format defining models, prompts, and page metadata; (3) the pipeline: generate via
image-gateway, optimize images for web, build per-model pages and the shared
comparison viewer, publish via site-publisher; (4) a model registry tracking per-image
cost and content-policy quirks observed; (5) regeneration support — adding one model
must not require redoing the others.

After writing it, test it with two models on a 3-prompt subset before running anything
at full scale.
  </task>
</prompt>
```

### Essay Illustration Gallery

Takes a finished essay or long post and produces a complete illustration package: the agent reads the piece, selects ~15–20 image-worthy moments across the essay's full arc (not just the obvious opener), locks a single illustration style so every frame is visually consistent, generates the images, writes a short "why this moment" caption per frame, and assembles everything into a gallery page — plus a ready-to-paste social note announcing it in the author's voice.

Why build it

The hard part of illustrating an essay isn't generating images — it's editorial judgment (which moments deserve images) and consistency (twenty images that look like one artist made them). This skill encodes both. It's also a great study in multi-skill composition packaged as one skill: analysis, style-locking, generation, captioning, gallery assembly, and publishing in a single repeatable pipeline.

What you need [](/open-skills/core-infrastructure#image-generation-gateway)  [](/open-skills/web-publishing-frontend#personal-site-publisher)  [](/open-skills/writing-voice-content#personal-voice-skill) 

Image Generation Gateway · Personal Site Publisher if you want the gallery published · Personal Voice Skill for the social note

Copy prompt

**Show the full setup prompt**

```
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
 [](/open-skills/skills)  [](/open-skills/runbooks) 

Back to the Skills directory or continue into runbook compositions.