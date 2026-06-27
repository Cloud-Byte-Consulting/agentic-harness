# Personal Site Publisher

Publishes a finished page to your personal or company website as a real, share-ready URL — handling everything that separates "an HTML file" from "a published page": the design language, a clean slug and URL route, a page-specific Open Graph preview image (1200×630) so links unfurl properly, share title and description, indexing controls (public vs. unlisted), local verification before deploy, and the deploy itself. The skill encodes your site's stack and conventions so publishing is a procedure, not a project.

## Why Build It
The gap between "the agent made a great page" and "that page is live at a clean URL with a proper link preview" is where most agent web output stalls. Encoding the full publish path once — including the unglamorous parts like OG images and noindex flags — means anything your agent produces is one sentence away from shippable. This skill is the final step of half the runbooks in this library.

## What You Need


## Prompt / Setup
```xml
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
