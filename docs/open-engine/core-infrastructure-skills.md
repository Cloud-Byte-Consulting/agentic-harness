# Core Infrastructure - Open Skills | Unlock AI

Generated 2.39:1 Open Skills category masthead showing categorized skill drawers on the right side. [←Skills directory](/open-skills/skills) 

# Core Infrastructure

The primitives other skills call. Build these first — several skills and most runbooks in this library depend on them.

Use this category page to compare the primitives, check their requirements, and copy the setup prompt for the smallest skill that makes the next workflow repeatable.

Pick one primitive, install it, test it, then come back when your workflow teaches it a better default.

5 skills

### Image Generation Gateway

Generates or edits images through a single API (OpenRouter is a good choice) with one command and zero per-call setup. The skill stores your saved preferences — default model, output directory, default size — so "generate an image of X" just works. It captures the current request shape of the API: which fields the endpoint expects, which model IDs are live, what each model costs per image, and the gotchas you've already hit. Other skills reference this one instead of writing their own API code.

Why build it

Image APIs change constantly, and every agent session that improvises an API call repeats old mistakes. Centralizing image generation in one skill means you fix an API change once, and every workflow that generates images inherits the fix. This is the clearest example of a skill as a shared primitive: at least three other skills in this library call it rather than reimplementing it.

What you need

An OpenRouter account and API key (or any image API you prefer — the pattern is identical)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new AI coding-agent skill called "image-gateway".
  </task>

  <storage>
    Store the skill wherever this harness loads skills from, such as
    ~/.claude/skills/image-gateway/SKILL.md or ~/.codex/skills/image-gateway/SKILL.md.
  </storage>

  <job>
    The skill generates or edits images through the OpenRouter API with one command.
    It should use saved preferences so routine image requests do not require per-call setup.
  </job>

  <inputs_to_collect>
    <input>Preferred default image model.</input>
    <input>Default output directory.</input>
    <input>Where the OpenRouter API key lives. It must be read from an env file, never written into the skill.</input>
  </inputs_to_collect>

  <requirements>
    <requirement>Define trigger conditions for direct image-generation requests and for other skills that need image generation.</requirement>
    <requirement>Document the current OpenRouter image API request shape.</requirement>
    <requirement>Include a working curl or script example that reads the key from the env file.</requirement>
    <requirement>Store the collected preferences in the skill.</requirement>
    <requirement>Include per-image cost notes for the selected default model.</requirement>
    <requirement>Tell other skills to call this skill instead of writing their own image API code.</requirement>
  </requirements>

  <verification>
    Generate one image with the saved defaults, show me the result, then update the skill with anything learned from the test.
  </verification>
</prompt>
```

### Current-Information Search

Routes the agent's web research through a search API built for discovering new information (Perplexity's API is the canonical choice) instead of the harness's built-in search. The skill defines when to use it — recent releases, pricing, anything that may contradict the model's training data — and carries the exact API call shape, default model choice, and key location. Optionally it can be wired in as a hook so all web searches redirect automatically.

Why build it

Agents are confidently out of date. The single most common failure mode in AI-assisted research is the model "confirming" stale training data instead of discovering what changed last week. A dedicated search skill turns "search for current info" from a hope into a procedure — and it makes every other research-flavored skill in this library more trustworthy.

What you need

A Perplexity API key (or another search API with real-time results)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "current-info-search", stored
wherever my harness loads skills from.

The skill's job: when I ask about anything that changes quickly — AI model releases,
pricing, software versions, news, APIs — call the Perplexity API directly instead of
relying on training data or default web search.

Before writing it, interview me for: where to store my Perplexity API key (env file)
and which Perplexity model to default to (suggest one).

The skill must include: (1) trigger conditions — any question about recent or
fast-moving information, and any time my claim or the agent's knowledge might be
stale; (2) a working curl example for the Perplexity chat completions endpoint that
reads the key from the env file; (3) a rule to cite dates and primary sources in
answers built from search results; (4) a rule that when search results contradict the
model's training data, the search results win.

After writing it, test it by asking yourself one question about something released in
the last month, run the search, and show me the answer with sources.
  </task>
</prompt>
```

### Media Transcription

Transcribes local audio or video files with a transcription API (AssemblyAI is a strong default) and packages the output into reusable artifacts: a clean readable Markdown transcript, word-level timestamps, semantic chapters, and speaker labels. The skill captures the current API request shape — including newer fields the docs bury — so transcription work never repeats old API mistakes. The output artifacts are deliberately designed to feed other skills: editing workflows, research synthesis, and content generation all start from these files.

Why build it

Transcripts are the universal input format for media work. Once a video or recording is a timestamped transcript, your agent can edit it, summarize it, fact-check it, extract clips from it, and generate graphics for it. Almost every media runbook in this library starts here. Getting the packaging right once — consistent filenames, chapters, timestamps — is what makes the downstream skills composable.

What you need

An AssemblyAI API key (or comparable transcription API) · ffmpeg installed for audio extraction from video

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "media-transcription", stored
wherever my harness loads skills from.

The skill's job: take a local audio or video file path and produce a complete
transcription package using the AssemblyAI API.

Before writing it, interview me for: where my AssemblyAI API key should live (env
file), and where transcription outputs should be saved (suggest a folder convention
next to the source media).

The skill must include: (1) trigger conditions — any time I give you a media file and
ask for a transcript, captions, chapters, or "make this searchable"; (2) the current
AssemblyAI request shape including the speech model field, with a working script that
reads the key from the env file; (3) a standard output package: readable Markdown
transcript, word-level timestamp JSON, semantic chapters, and speaker labels, all with
consistent filenames; (4) an ffmpeg step to extract audio from video first when
needed; (5) a note that these artifacts are inputs for editing and research skills, so
the format must stay consistent.

After writing it, test it on a short audio file and show me the output package.
  </task>
</prompt>
```

### Heavy File Ingestion

Converts heavy, agent-hostile files — large PDFs, slide decks, spreadsheets, CSVs, long Word docs — into lightweight Markdown and CSV artifacts plus an index file, before any analysis begins. The skill enforces a discipline: never analyze a heavy file directly in context; convert it to lean text artifacts first, then analyze those. It includes the conversion recipes (which tools to use per file type) and the index format so a folder of converted material stays navigable.

Why build it

Heavy files silently destroy agent sessions. A 200-page PDF or a 40-tab spreadsheet read directly into context burns the context window, degrades reasoning, and leaves nothing reusable behind. Ingest-first means you pay the conversion cost once and every future session works from clean, greppable text. This is the foundational skill of the research runbook.

What you need

Nothing beyond standard conversion tools your agent can install (e.g. pdf-to-text utilities); no accounts

Copy prompt

**Show the full setup prompt**

```
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

### HTML Artifact Builder

Turns dense agent output — implementation plans, research explainers, code review summaries, comparison tables, walkthroughs, diagrams, interactive reports — into a single self-contained HTML file with consistent, polished styling. The skill carries your visual conventions (fonts, colors, layout patterns, dark/light preference) so every artifact looks like it came from the same shop, and it enforces self-containment: one file, inline CSS/JS, no external dependencies, openable anywhere.

Why build it

Long chat responses are where good analysis goes to die. A complex comparison or plan rendered as a styled, scrollable, sometimes interactive HTML page is dramatically more useful — you can read it properly, share it, and keep it. Once your agent has a house style for artifacts, "make this a page" becomes a one-line request, and the publishing skill (below) can take any artifact public.

What you need

Nothing. This is pure agent capability plus your taste.

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "html-artifacts", stored wherever my
harness loads skills from.

The skill's job: render dense or visual output — plans, reports, research explainers,
review summaries, comparisons, diagrams, walkthroughs — as a single self-contained
HTML file with my house style, instead of a long chat response.

Before writing it, interview me for: my visual preferences (typeface direction, color
palette or a brand color, dark or light default) and where artifact files should be
saved.

The skill must include: (1) trigger conditions — whenever output would be dense,
visual, interactive, or worth keeping/sharing, offer or produce an HTML artifact;
(2) hard rules: one file, inline CSS and JS, no external dependencies, works offline;
(3) my house style tokens (type, spacing, colors) defined once at the top so every
artifact matches; (4) layout patterns for the common cases: report, comparison table,
timeline, diagram, dashboard; (5) a rule to open or screenshot the result and verify
it renders before declaring it done.

After writing it, test it by converting your own setup summary of this skill into an
artifact and showing me.
  </task>
</prompt>
```
 [](/open-skills/skills)  [](/open-skills/runbooks) 

Back to the Skills directory or continue into runbook compositions.