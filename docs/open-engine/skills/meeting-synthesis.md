# Meeting Synthesis

Turns meeting recordings or transcripts into a structured synthesis: key takeaways, decisions made (with who made them), action items (with owners and deadlines where stated), open questions, and reusable context worth keeping beyond the meeting. The skill enforces a separation between what was actually said versus what the agent inferred, so the synthesis stays trustworthy.

## Why Build It
Meeting notes done by hand are either too thin to be useful or too long to be read. An agent with a fixed synthesis format produces the same reliable artifact from every meeting, and the decisions/actions/questions split means the output plugs directly into your task system instead of becoming another unread document.

## What You Need


## Prompt / Setup
```xml
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
