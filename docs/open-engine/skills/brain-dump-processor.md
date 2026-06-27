# Brain Dump Processor

Takes messy, multi-topic input — voice memo transcripts, stream-of-consciousness notes, long rambling drafts — and pans for gold: it extracts each distinct idea, separates them cleanly, evaluates which threads are worth pursuing, and files the results. The skill defines the extraction format (idea, context, why it might matter, suggested next step) and a consistent destination so processed ideas accumulate somewhere instead of evaporating.

## Why Build It
Your best ideas arrive mixed with your worst ones, usually while walking. Without a procedure, voice notes get transcribed and never read again. This skill makes the agent the filter: you talk for ten minutes, it hands back five separated ideas with an honest evaluation of each. Paired with the transcription skill, it turns "rambling into your phone" into a legitimate ideation pipeline.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "brain-dump-processor", stored
wherever my harness loads skills from.

The skill's job: process messy multi-topic input — voice memo transcripts, brain
dumps, rambling notes — into cleanly separated, evaluated ideas.

Before writing it, interview me for: where processed ideas should be filed (one inbox
file, a folder of dated notes, or a tool I use), and what I tend to ramble about so
the evaluation criteria fit my work.

The skill must include: (1) trigger conditions — whenever I share a voice transcript,
brain dump, or say "process this"; (2) an extraction format per idea: the idea in one
sentence, surrounding context, an honest assessment of whether it's worth pursuing and
why, and a concrete suggested next step; (3) a rule to separate genuinely distinct
ideas rather than summarizing the whole dump into mush; (4) a rule to flag
contradictions with things I've said before in the same dump; (5) the filing
destination and format.

After writing it, test it on a real note or transcript I give you.
  </task>
</prompt>
```
