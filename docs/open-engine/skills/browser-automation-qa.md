# Browser Automation QA

Professional-grade web testing through browser automation (the Chrome DevTools Protocol, exposed to agents via MCP, is the proven route): performance traces and Core Web Vitals measurement (LCP, INP, CLS), network request monitoring, console error capture, device emulation for responsive testing, accessibility checks, and scripted multi-page workflows — with screenshots and metrics as evidence. The skill encodes which checks to run for which kind of change, and what "passing" means for your projects.

## Why Build It
"It looks fine" is not verification. This skill upgrades your agent from eyeballing pages to measuring them — and because it produces evidence (metrics, screenshots, console output), you can trust a report without re-checking it yourself. Combined with the two skills above, it completes a QA stack: process (Page Testing Memory), institutional memory (Testing Runbook Creator), and instrumentation (this).

## What You Need


## Prompt / Setup
```xml
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
