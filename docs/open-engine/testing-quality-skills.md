# Testing & Quality - Open Skills | Unlock AI

Generated 2.39:1 Open Skills category masthead showing categorized skill drawers on the right side. [←Skills directory](/open-skills/skills) 

# Testing & Quality

Skills that make agent-built things trustworthy — and make the next agent session smarter than the last one.

Use this category page to compare the primitives, check their requirements, and copy the setup prompt for the smallest skill that makes the next workflow repeatable.

Pick one primitive, install it, test it, then come back when your workflow teaches it a better default.

3 skills

### Testing Runbook Creator

Enforces one rule: testing discoveries must not die in chat. Whenever your agent tests, QAs, smoke-tests, or verifies anything in a repo, this skill makes it leave behind a repo-local runbook entry — how to test that page or workflow, which actions are safe vs. destructive, setup and seed data requirements, cleanup steps, and the exact verification commands. The runbook lives in the repo, accumulates over time, and every future agent session reads it before re-testing.

Why build it

Without this, every agent session rediscovers your app from scratch — which test account, which route, which actions are safe — and the discoveries evaporate when the session ends. With it, testing knowledge compounds: session twenty inherits everything sessions one through nineteen learned. This is the single highest-leverage habit-skill in the library, and the purest expression of the principle that agent work should leave durable artifacts.

What you need

Nothing. Works in any repo from day one.

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "testing-runbook-creator", stored
wherever my harness loads skills from.

The skill's job: whenever you test, verify, smoke-test, QA, or debug anything in a
repo, capture what you learned as a repo-local runbook entry so future sessions don't
rediscover it.

The skill must include: (1) trigger conditions — ANY testing or verification activity
in a repo, not just when I say "runbook"; (2) a standard runbook location (suggest
docs/testing-runbook.md or similar) and entry format: the page/workflow/feature, how
to test it step by step, safe actions vs. destructive actions, setup/seed
requirements, cleanup steps, and exact verification commands with expected output;
(3) a read-first rule: before testing anything, check whether the runbook already
covers it and follow the existing recipe; (4) an update rule: when reality differs
from the runbook, fix the runbook in the same session; (5) a rule to record
discoveries as you go, not as an end-of-session afterthought.

After writing it, test it by smoke-testing one workflow in a project I point you to
and showing me the runbook entry it produces.
  </task>
</prompt>
```

### Page Testing Memory

The architectural partner to Testing Runbook Creator, encoding a specific split: the global skill teaches the page-QA process (how to approach testing any web page — routes, states, forms, auth, responsive checks), while page-specific facts — selectors, test accounts, magic URLs, cleanup quirks — belong in repo-local runbooks. The skill teaches your agent which knowledge goes where, keeping global skills lean and portable while repo knowledge stays with the repo.

Why build it

The most common failure mode in a growing skill library is global skills bloated with project specifics — selectors from one client's app baked into a skill that loads in every session. This skill is the antidote, and the global-process/local-facts split it teaches generalizes to your whole library. It's a skill about how to structure skills, disguised as a QA skill.

What you need [](/open-skills/testing-quality#testing-runbook-creator) 

Testing Runbook Creator (they're designed as a pair)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Create a new skill for my AI coding agent called "page-testing-memory", stored
wherever my harness loads skills from. It partners with my testing-runbook-creator
skill.

The skill's job: teach the general page-QA process globally, while keeping all
page-specific knowledge in repo-local runbooks — never in this skill.

The skill must include: (1) trigger conditions — QA or verification of any web page or
UI; (2) the general process: identify the page's states (empty, loaded, error,
loading), test forms with valid/invalid/edge input, verify auth boundaries, check
responsive behavior at standard breakpoints, capture screenshots as evidence; (3) the
knowledge split, stated explicitly: process lives here; selectors, routes, test
accounts, seed data, and cleanup quirks live in the repo's testing runbook; (4) a rule
that when you learn a page-specific fact during QA, it goes into the repo runbook
immediately — and if you find yourself wanting to add a project detail to THIS skill,
that's the signal it belongs in the repo instead.

After writing it, test it by QAing one page in a project I choose, and show me both
the QA findings and what got written to the repo runbook.
  </task>
</prompt>
```

### Browser Automation QA

Professional-grade web testing through browser automation (the Chrome DevTools Protocol, exposed to agents via MCP, is the proven route): performance traces and Core Web Vitals measurement (LCP, INP, CLS), network request monitoring, console error capture, device emulation for responsive testing, accessibility checks, and scripted multi-page workflows — with screenshots and metrics as evidence. The skill encodes which checks to run for which kind of change, and what "passing" means for your projects.

Why build it

"It looks fine" is not verification. This skill upgrades your agent from eyeballing pages to measuring them — and because it produces evidence (metrics, screenshots, console output), you can trust a report without re-checking it yourself. Combined with the two skills above, it completes a QA stack: process (Page Testing Memory), institutional memory (Testing Runbook Creator), and instrumentation (this).

What you need

Chrome plus a DevTools MCP server connected to your harness (your agent can set this up — that's step one of the prompt)

Copy prompt

**Show the full setup prompt**

```
<prompt>
  <task>
    Set up browser automation QA for my AI coding agent, then create a skill called
"browser-qa" that uses it, stored wherever my harness loads skills from.

Step 1: Set up a Chrome DevTools MCP server for my harness so you can drive a real
browser — navigate, screenshot, read console and network activity, run performance
traces, and emulate devices. Walk me through any installation steps you can't do
yourself, and verify the connection works before writing the skill.

Step 2: The skill must include: (1) trigger conditions — when I ask to verify a web
change, check performance, audit a page, or test responsive behavior; (2) check
recipes by change type: layout changes get screenshots at desktop/tablet/mobile
breakpoints; performance-relevant changes get a trace with Core Web Vitals (LCP, INP,
CLS) against stated thresholds; new features get console-error and failed-request
checks during a scripted walkthrough; (3) an evidence rule: every finding ships with
its screenshot, metric, or log excerpt — no unevidenced "looks fine"; (4) a standard
short report format: what was checked, what passed with evidence, what failed with
reproduction steps; (5) integration: findings about how to test a page get written to
the repo's testing runbook per my testing-runbook-creator skill.

After writing it, test it by auditing one live page of mine and showing me the report.
  </task>
</prompt>
```
 [](/open-skills/skills)  [](/open-skills/runbooks) 

Back to the Skills directory or continue into runbook compositions.