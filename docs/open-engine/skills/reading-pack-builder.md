# Reading Pack Builder

Takes a pile of local documents — docs, SOPs, change requests, research notes, review materials — and builds a controlled reading surface: a local HTML reading pack that presents one document at a time, in a deliberate order, with an index and progress tracking. Instead of "here are 14 files, good luck," you get a guided review experience your agent assembled.

## Why Build It
Review is a workflow, not a folder. When you (or a collaborator) actually need to read and sign off on a set of materials, structure matters: order, one-at-a-time focus, and a record of what's been covered. This skill is also a nice demonstration that agent output doesn't have to be text in a chat — it can be a purpose-built interface, generated in seconds.

## What You Need


## Prompt / Setup
```xml
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
