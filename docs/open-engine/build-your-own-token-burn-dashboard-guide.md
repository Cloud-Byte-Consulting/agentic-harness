# Build your own token-burn dashboard | Unlock AI

Build guide

# Build your own token-burn dashboard.

Your AI spend is the most honest report card you have. This guide turns that spend into a dashboard you can build yourself: one screen for seeing where tokens go across Codex, Claude, and ChatGPT.

Burn rate is not a vanity stat. It is the clearest signal of whether you are getting fluent or just getting expensive.

A weekendBuild time$0Local data + Vercel5 viewsHeatmap, trend, drivers, scale, tableCodexexactClaude CodeexactClaude chatestimatedChatGPTestimateddaily-burn.jsonone row per daydriver labelsbehavior signalToken burn dashboard

## AI usage by day

90d180d1yallTotal18.4MtokensPeak day756Kshipping spike7d average312K-18% from peakDaily burnlog scaleWeekly trendpeak labeledDriversshareshipping41%review27%research20%video12%Codex exactClaude Code exactClaude chat estimatedChatGPT estimated

What you build

## One tidy table, five honest reads.

The dashboard is not complicated. It is five different views over normalized daily totals. The discipline is in the data shape and the labels.

01

### Daily burn heatmap

Every day colored by tokens spent, on a log scale.

02

### Weekly trend line

A cleaner read on whether your usage is climbing or getting leaner.

03

### Burn drivers

The work families that eat the budget: shipping, research, review, video, admin.

04

### Scale equivalents

A huge token number translated into something a human can feel.

05

### Moving-average table

The receipts: per tool, per day, exact and estimated side by side.

Starter kits

## Start from the stack you actually use.

Each download is an agent skill with the dashboard starter files baked in. Your AI reads the skill, copies the starter app, interviews you for missing data, and builds from there.

Codex

### Exact local usage first.

Use this when Codex is where your tokens are already visible. The kit starts with measured Codex totals and leaves estimates out until you ask for them.

 [Download skill zip](/guides/token-burn/starter-kits/token-burn-codex-skill.zip) Claude

### Claude Code plus Claude chat.

Use this when Claude Code logs are the hard numbers and Claude chat needs honest estimation.

 [Download skill zip](/guides/token-burn/starter-kits/token-burn-claude-skill.zip) ChatGPT

### Export-based estimates.

Use this when your main usage lives in ChatGPT and the dashboard needs to reason from exports or conversation artifacts.

 [Download skill zip](/guides/token-burn/starter-kits/token-burn-chatgpt-skill.zip) All Sources

### Codex, Claude, and ChatGPT together.

Use this when you want every source on the same dashboard axes from the beginning.

 [Download skill zip](/guides/token-burn/starter-kits/token-burn-all-sources-skill.zip)  [](/guides/token-burn/starter-kits/manifest.json) `/llms.txt``/agents.txt`

Agents should use the starter-kit manifest for SHA-256 checksums, source links, and deterministic download URLs. The same contract is advertised in  and .

Watch the build

## A screen-led walkthrough before you hand it to your agent.

This is the quick pass: choose the kit, let the agent fetch the guide context, build locally, then decide what data is safe enough to share.

AI-narrated guide voiceover. This is not a cloned Nate voice.

Visual pass

## Make the invisible workflow visible.

The page should feel like the system it is asking you to build: logs moving into rows, rows becoming views, views getting verified before anything ships.

![Diagram showing AI usage sources flowing into one daily burn data file.](/guides/token-burn/source-map.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
Source map

### Map the sources before you touch the UI.

Exact logs and honest estimates need to land in one normalized file. That is the whole backbone.

![Dark cyan Unlock AI dashboard mockup showing token burn heatmaps, trend data, source fidelity, and spend drivers.](/guides/token-burn/target-surface.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
Target surface

### Keep the finished dashboard visible while you build.

The target is not a fancy chart collection. It is a readable daily operating surface for AI spend.

![Diagram showing the build loop from usage logs to normalized rows, dashboard views, verification, and deployment.](/guides/token-burn/build-loop.svg?dpl=dpl_5FTpDMvvtC6xsZCjUVkH2a1W7smw)
Build loop

### Work in loops until the math gets boring.

Collect, normalize, build, verify, ship. If the totals do not reconcile, the loop is not finished.

Why build it

## A bill tells you what happened. A burn dashboard changes behavior.

Most people pay the bill or hit the wall and assume that is just what AI costs. It is not. The first step to spending less is being able to see what you spend.

01

### You cannot improve what you cannot see.

Burn hides in habits: raw PDFs, runaway conversations, bad model choice, and expensive context. A dashboard turns invisible waste into a number you watch.

02

### It is a fluency meter, not a bill.

Tokens per outcome is the clearest tell of whether you are getting better with the tools or just spending harder.

03

### The lesson gets more expensive.

As models get more capable, sloppy workflows cost more. Measuring early gives you a way to get sharper before the burn scales up.

The anatomy

## The heatmap is the conscience. The table is the truth.

Start with daily totals and build outward. The visual layer should make spikes visible, but the raw rows still need to reconcile.

Daily burnlog scale: less to more

The old mistake is making this pretty before making it true. Normalize the data first, then give every visualization the same exact-vs-estimated honesty. A dashboard that confidently displays bad math is worse than no dashboard.

`driver`

The field is the move that makes it useful. It turns "I spent a lot" into "I spent a lot on shipping, research, review, or video."

Before you start

## Nothing exotic. Local files, one agent, one deploy.

If you already build with a coding agent and can run a dev server, you have the practical skills you need.

### Tools

* Codex App, Claude Code, Cursor, or another coding agent.
* Node 20+ and a terminal.
* A Vercel account if you want to publish it.
* The usage data from whichever tools you want to track.

### Rules

* Keep raw exports out of public repos.
* Commit only the normalized totals you are comfortable sharing.
* Label estimated values everywhere.
* Build local first. Deploy only when the math reconciles.

Step 1

## Get the data, then admit how good it is.

Some tools log real token usage. Others force you to estimate. The honest move is making that fidelity visible in the interface.

| Source      | Fidelity  | How to pull it                                                                                                                            |
| ----------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Codex       | exact     | Codex app and CLI sessions write local logs with real token usage. Have your agent total input and output by local day.                   |
| Claude Code | exact     | Claude Code stores per-session JSONL with input, output, and cache counts. Include API or agent calls if you use them.                    |
| Claude chat | estimated | The web and desktop chat path has no tidy local token export. Estimate from message counts and average lengths, then label it honestly.   |
| ChatGPT     | estimated | Request your data export and tokenize conversation text by date. Treat this as calibrated estimation unless you have exact provider logs. |

Do not let estimates cosplay as measurements. That is the fastest way to make the dashboard feel useful while quietly training you on bad numbers.

Step 2

## Normalize everything into one daily row.

Every view is just a different read of this shape. Get this right and the UI becomes straightforward.

```
// daily-burn.json - one row per day, in your local timezone
[
  {
    "date": "2026-05-24",
    "codex_tokens": 184320,
    "claude_code_tokens": 512880,
    "claude_code_calls": 47,
    "claude_chat_est": 38000,
    "chatgpt_est": 21000,
    "total": 756200,
    "driver": "shipping",
    "evidence": "dashboard build and review"
  }
]
```

Step 3

## Let the agent build the dashboard, but give it the right spec.

The difference between a useful agent build and a random dashboard is precision: file shape, view order, hard rules, and verification.

### Token-burn dashboard build prompt

Paste this into your coding agent and let it fetch the guide context.

Copy prompt
```
# Token-Burn Dashboard - Agent Handoff Prompt

You are my coding agent. Build me a local token-burn dashboard that shows where my AI usage goes and what work the computer should do next.

## First fetch the guide context

1. Read this guide: https://unlock-ai.natebjones.com/guides/build-your-own-token-burn-dashboard
2. Read the AI discovery file: https://unlock-ai.natebjones.com/llms.txt
3. Read the starter-kit manifest: https://unlock-ai.natebjones.com/guides/token-burn/starter-kits/manifest.json

## Choose and install the right starter kit

From the manifest, choose:

- `codex` if my best source is exact local Codex usage.
- `claude` if I primarily use Claude Code or Claude chat.
- `chatgpt` if I primarily use ChatGPT and need export-based estimates.
- `all-sources` if I want Codex, Claude Code, Claude chat, and ChatGPT on the same dashboard axes.

Download the kit, verify the SHA-256 against the manifest, unzip it, read `SKILL.md`, and copy `assets/dashboard-starter/` into a new local project folder.

## Build the dashboard

Use the bundled starter app unless there is a clear reason not to. It should render:

1. Daily burn heatmap on a logarithmic color scale.
2. Weekly trend line using log-normalized totals.
3. Burn drivers grouped by `driver`.
4. Scale equivalents with visible approximation math.
5. Last-30-days moving-average table with every source column.

The normalized row shape is:

```json
{
  "date": "YYYY-MM-DD",
  "codex_tokens": 0,
  "claude_code_tokens": 0,
  "claude_code_calls": 0,
  "claude_chat_est": 0,
  "chatgpt_est": 0,
  "total": 0,
  "driver": "shipping",
  "evidence": "scrubbed work-family note"
}
```

## Interview me before estimating

If a source cannot be measured directly, ask short questions and infer a conservative range. Then write a labeled estimate. Inferred numbers are fine. Pretending they are exact is not.

Ask:

- Which tools should be included?
- What timezone should define a day?
- Where are exact logs or exports located?
- Which days were shipping, research, review, video, planning, writing, support, or admin?
- Which evidence notes must stay private?

## Privacy rules

- Do not include raw logs, exports, prompts, private paths, client names, project names, or secrets in the app.
- Keep private detail in the local working file if useful.
- Deploy or share only scrubbed normalized rows.

## When done
1. Run `npm install`.
2. Run `npm run build`.
3. Run the dev server and verify all five views.
4. Check that exact and estimated labels are visible.
5. List the files created or changed.
6. Tell me how to add the next day of data.
7. Give me the deploy command, but do not deploy until I confirm the public data is scrubbed.
```

Step 4

## Review the five views like a builder, not a spectator.

Each view has a job and a common failure mode. Check those before you polish anything.

01

### Daily burn heatmap

The at-a-glance conscience. It should make a runaway day obvious without flattening quiet days.

Use a log color scale. Linear heatmaps lie when one spike dominates.

02

### Weekly trend line

Smooth the daily noise into a direction. The question is whether usage is compounding or getting leaner.

Use log y-axis again and label the peak week.

03

### What is driving the burn

Turn 'I spent a lot' into 'I spent a lot on shipping, review, or research.'

Driver labels matter. Keep the vocabulary small.

04

### Scale equivalents

Translate huge token totals into rough human-scale comparisons.

Show the math or it reads as a gimmick.

05

### Moving-average table

The receipts drawer. This is where exact and estimated sources sit side by side.

Make the trust level obvious at every row.

Step 5

## Run it locally. Ship it when the math is boring.

The dashboard does not need a backend for v1. Keep the data file local, build the page, and deploy only the normalized totals you are comfortable exposing.

Local

* `npm install`
* `npm run dev`, then open`localhost:3000`
* Confirm all five views render from your data.
->Vercel

* Push the project to GitHub.
* Import it at Vercel or run`vercel`from the folder.
* Set it private if your usage data is personal.

Verify

## A subtly wrong token dashboard is worse than none.

Before you call it done, make the dashboard prove it is telling the truth.

✓

### The totals reconcile.

Pick one day. Add each source column by hand. It should match the day's total.

✓

### Exact and estimated numbers are labeled everywhere.

Readers should never need to guess whether a number is measured or inferred.

✓

### The log scales actually read.

A 10x day and a 1x day should both be visible. If the map looks flat, the scale is wrong.

✓

### The time-range selector changes every view.

Switch 90 days to all-time and confirm the heatmap, trend, drivers, and table all respond.

Troubleshooting

## The problems are usually data problems.

If the dashboard feels wrong, debug the input assumptions before you redesign the output.

fix

### Timezones smear your days

Logs are often UTC. Pick one timezone, convert on ingest, and bucket by local date before totals are calculated.

fix

### The heatmap looks flat

That almost always means a linear color scale. Switch to log and use enough ramp stops to show quiet days.

fix

### Estimates dwarf the real numbers

A bad tokens-per-message constant can swamp the dashboard. Calibrate it against one real conversation.

Make it yours

## V1 shows the spend. V2 changes the spend.

Once the dashboard is accurate, add the features that push behavior instead of merely reporting it.

+

### A burn budget

Set a daily target and mark days that blow past it. A line you can cross is a line you start to notice.

+

### Tokens per outcome

Tag days with what shipped and divide. Falling cost per outcome is the fluency curve worth bragging about.

+

### Auto-ingest

A small nightly script can read each tool's logs and append a new day, so the dashboard stays current without touching JSON.

Editorial bridge

## The point is the next delegation decision.

The dashboard earns its keep when it changes what you hand to the computer tomorrow.

01

### Start where the tokens are already visible.

Codex and Claude Code are the easy on-ramps because they can expose exact usage. Chat tools still work; they just need estimates with labels.

02

### Borrow the design discipline.

The starter kits already carry the chart shape and restraint. Your job is to keep the dashboard readable enough to return to.

03

### Describe the operating surface in plain language.

Ask for the heatmap, log scale, same-day strip, drivers, source split, and top days. Then iterate until it matches the work you actually do.

04

### Infer what you cannot measure directly.

Have the agent interview you and land on a conservative range. The sin is not estimating. The sin is hiding that you estimated.

05

### Scrub before anything leaves the machine.

Private dashboards can name the real work. Public dashboards should keep only normalized totals and generic drivers.

06

### Ship it, or keep it local.

A Vercel deploy is optional. The actual requirement is that the dashboard exists where you can return to it and decide what to hand the computer next.

 [](https://natesnewsletter.substack.com/p/your-claude-sessions-cost-10x-what) 

This guide is adapted from the original Limited Edition Jonathan build guide and the companion thesis that token burn rate is a revealing metric of AI fluency. Read the related essay on Nate's newsletter.