# Research & Thinking - Open Skills | Unlock AI

Generated 2.39:1 Open Skills category masthead showing categorized skill drawers on the right side. [←Skills directory](/open-skills/skills) 

# Research & Thinking

Skills that turn raw inputs — voice notes, meetings, document piles, weekly noise — into structured, reviewable thinking.

Use this category page to compare the primitives, check their requirements, and copy the setup prompt for the smallest skill that makes the next workflow repeatable.

Pick one primitive, install it, test it, then come back when your workflow teaches it a better default.

5 skills

### Brain Dump Processor

Takes messy, multi-topic input — voice memo transcripts, stream-of-consciousness notes, long rambling drafts — and pans for gold: it extracts each distinct idea, separates them cleanly, evaluates which threads are worth pursuing, and files the results. The skill defines the extraction format (idea, context, why it might matter, suggested next step) and a consistent destination so processed ideas accumulate somewhere instead of evaporating.

Why build it

Your best ideas arrive mixed with your worst ones, usually while walking. Without a procedure, voice notes get transcribed and never read again. This skill makes the agent the filter: you talk for ten minutes, it hands back five separated ideas with an honest evaluation of each. Paired with the transcription skill, it turns "rambling into your phone" into a legitimate ideation pipeline.

What you need [](/open-skills/core-infrastructure#media-transcription) 

Nothing required; pairs naturally with Media Transcription for voice memos

Copy prompt

**Show the full setup prompt**

```
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

### Meeting Synthesis

Turns meeting recordings or transcripts into a structured synthesis: key takeaways, decisions made (with who made them), action items (with owners and deadlines where stated), open questions, and reusable context worth keeping beyond the meeting. The skill enforces a separation between what was actually said versus what the agent inferred, so the synthesis stays trustworthy.

Why build it

Meeting notes done by hand are either too thin to be useful or too long to be read. An agent with a fixed synthesis format produces the same reliable artifact from every meeting, and the decisions/actions/questions split means the output plugs directly into your task system instead of becoming another unread document.

What you need [](/open-skills/core-infrastructure#media-transcription) 

Nothing required; pairs with Media Transcription if you start from recordings

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "meeting-synthesis", stored wherever
my harness loads skills from.

The skill's job: turn a meeting transcript or recording into a structured synthesis I
can act on.

Before writing it, interview me for: where syntheses should be saved, and whether
action items should also go somewhere specific (task tool, file, email draft).

The skill must include: (1) trigger conditions — any meeting transcript, recording, or
"what happened in this meeting" request; (2) a fixed output structure: takeaways,
decisions (with who decided), action items (with owner and deadline where stated),
open questions, and durable context worth keeping; (3) a hard rule separating what was
said from what you inferred — inferences get marked as such; (4) a rule to preserve
exact quotes for anything contentious or commitment-shaped; (5) handling for
multi-topic meetings: synthesize per topic, not chronologically.

After writing it, test it on one real transcript and show me the synthesis.
  </task>
</prompt>
```

### Weekly Signal Diff

On a recurring basis (weekly is the natural cadence), reviews a defined set of inputs — your notes, a folder, feeds, project state, saved searches — and reports only what meaningfully changed since the last run: new signals, shifted assumptions, dead threads, emerging patterns. The skill keeps a small state file recording what it saw last time, which is what makes a true diff possible instead of a weekly summary that repeats itself.

Why build it

The hard part of staying current isn't gathering information, it's noticing change. A diff against last week's state surfaces exactly the delta — what's new, what moved, what quietly died — and ignores the stable background. This is also a gentle introduction to stateful skills: the state file pattern (skill remembers its last run) unlocks a whole class of recurring workflows.

What you need [](/open-skills/core-infrastructure#current-information-search) 

A defined set of inputs to watch; pairs well with Current-Information Search for external signals

Copy prompt

**Show the full setup prompt**

```
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

### Assumption Checker

Audits a plan, argument, or strategy doc for world-model problems: unstated assumptions, missing evidence, internal contradictions, and gaps between what the document claims and what it actually demonstrates. The output is a structured diagnostic — each assumption listed, rated by how load-bearing it is and how well-supported, with the single most dangerous assumption flagged.

Why build it

Agents are excellent at making plans sound coherent, which is precisely the danger. A dedicated adversarial pass — run as its own skill with its own posture, not as an afterthought in the same conversation that produced the plan — reliably catches the "we assumed the API does X" and "this only works if users behave like Y" failures before they cost you a week.

What you need

Nothing

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "assumption-checker", stored wherever
my harness loads skills from.

The skill's job: adversarially audit a plan, argument, or strategy document for
unstated assumptions, missing evidence, contradictions, and world-model gaps.

The skill must include: (1) trigger conditions — when I ask you to check, stress-test,
or red-team a plan or document; (2) a posture rule: in this mode you are a skeptic,
not a collaborator — do not soften findings or balance them with praise; (3) an output
format: each assumption stated plainly, rated for how load-bearing it is and how
well-evidenced, with the single most dangerous assumption called out at the top;
(4) a rule to check claims against the actual sources or code when they're available,
not just against internal consistency; (5) a closing section: the three questions that
would most reduce risk if answered.

After writing it, test it on any plan or doc I give you — or on one of your own recent
plans from this session.
  </task>
</prompt>
```

### Reading Pack Builder

Takes a pile of local documents — docs, SOPs, change requests, research notes, review materials — and builds a controlled reading surface: a local HTML reading pack that presents one document at a time, in a deliberate order, with an index and progress tracking. Instead of "here are 14 files, good luck," you get a guided review experience your agent assembled.

Why build it

Review is a workflow, not a folder. When you (or a collaborator) actually need to read and sign off on a set of materials, structure matters: order, one-at-a-time focus, and a record of what's been covered. This skill is also a nice demonstration that agent output doesn't have to be text in a chat — it can be a purpose-built interface, generated in seconds.

What you need

Nothing; builds on HTML Artifact Builder's conventions if you have it

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "reading-pack-builder", stored
wherever my harness loads skills from.

The skill's job: given a set of local documents to review, build a self-contained
local HTML reading pack that presents them one at a time in a deliberate order, with
an index page and simple progress tracking.

Before writing it, interview me for: where reading packs should be saved, and my
visual preferences if I don't already have an html-artifacts skill to inherit from.

The skill must include: (1) trigger conditions — when I have a pile of documents to
review or ask for a "reading pack"; (2) conversion of each source document to clean
HTML, preserving structure; (3) an index page with one-line summaries and a suggested
reading order with reasoning; (4) one-at-a-time navigation (previous/next) and a
simple read/unread marker stored locally; (5) self-containment — everything works
offline as local files.

After writing it, test it on 3 or more documents I point you to and open the result.
  </task>
</prompt>
```
 [](/open-skills/skills)  [](/open-skills/runbooks) 

Back to the Skills directory or continue into runbook compositions.