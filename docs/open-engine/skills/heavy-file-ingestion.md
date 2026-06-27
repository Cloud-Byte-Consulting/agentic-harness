# Heavy File Ingestion

Converts heavy, agent-hostile files — large PDFs, slide decks, spreadsheets, CSVs, long Word docs — into lightweight Markdown and CSV artifacts plus an index file, before any analysis begins. The skill enforces a discipline: never analyze a heavy file directly in context; convert it to lean text artifacts first, then analyze those. It includes the conversion recipes (which tools to use per file type) and the index format so a folder of converted material stays navigable.

## Why Build It
Heavy files silently destroy agent sessions. A 200-page PDF or a 40-tab spreadsheet read directly into context burns the context window, degrades reasoning, and leaves nothing reusable behind. Ingest-first means you pay the conversion cost once and every future session works from clean, greppable text. This is the foundational skill of the research runbook.

## What You Need


## Prompt / Setup
```xml
<prompt>
  <task>
    Create a new skill for my AI coding agent called "heavy-file-ingestion", stored
wherever my harness loads skills from.

The skill's job: when I hand you a heavy file (big PDF, slide deck, spreadsheet, CSV
dump, long doc), convert it into lightweight Markdown/CSV artifacts plus an index
BEFORE doing any analysis — never analyze the heavy file directly.

Before writing it, interview me for: where converted artifacts should live (suggest a
convention like an `_ingested/` folder next to the source), and which file types I
handle most often.

The skill must include: (1) trigger conditions — any heavy or binary document I share,
or any analysis request that touches one; (2) per-file-type conversion recipes using
tools available on my machine, installing what's missing; (3) a standard index file
listing each artifact with a one-line summary; (4) the rule that analysis always reads
the converted artifacts, never the original; (5) chunking guidance for very large
sources so each artifact stays comfortably readable.

After writing it, test it on one real PDF or deck I give you and show me the artifact
folder and index.
  </task>
</prompt>
```
