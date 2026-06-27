# Personal Voice Skill

Encodes how you actually write — across contexts, not as a single tone preset. The skill captures your voice along multiple registers (direct/instructional, warm/relational, analytical, business-formal), with real samples of each, plus the rules of when to use which: when you're blunt, when you soften, what words and constructions you never use, how your emails differ from your posts. The agent then writes drafts that need light edits instead of rewrites.

## Why Build It
Generic AI prose is the most recognizable writing style on the internet right now, and "write this in a friendly tone" doesn't fix it. A voice skill built from your actual writing samples — with explicit anti-patterns ("never open with 'I hope this finds you well'", "never use 'delve'") — is the difference between an agent that drafts for you and one that drafts as you. This is consistently one of the highest-leverage skills for anyone who publishes or sends a lot of words.

## What You Need


## Prompt / Setup
```xml
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
