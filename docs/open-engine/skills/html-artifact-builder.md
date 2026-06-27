# HTML Artifact Builder

Turns dense agent output — implementation plans, research explainers, code review summaries, comparison tables, walkthroughs, diagrams, interactive reports — into a single self-contained HTML file with consistent, polished styling. The skill carries your visual conventions (fonts, colors, layout patterns, dark/light preference) so every artifact looks like it came from the same shop, and it enforces self-containment: one file, inline CSS/JS, no external dependencies, openable anywhere.

## Why Build It
Long chat responses are where good analysis goes to die. A complex comparison or plan rendered as a styled, scrollable, sometimes interactive HTML page is dramatically more useful — you can read it properly, share it, and keep it. Once your agent has a house style for artifacts, "make this a page" becomes a one-line request, and the publishing skill (below) can take any artifact public.

## What You Need


## Prompt / Setup
```xml
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
