# Testing Runbook Creator

Enforces one rule: testing discoveries must not die in chat. Whenever your agent tests, QAs, smoke-tests, or verifies anything in a repo, this skill makes it leave behind a repo-local runbook entry — how to test that page or workflow, which actions are safe vs. destructive, setup and seed data requirements, cleanup steps, and the exact verification commands. The runbook lives in the repo, accumulates over time, and every future agent session reads it before re-testing.

## Why Build It
Without this, every agent session rediscovers your app from scratch — which test account, which route, which actions are safe — and the discoveries evaporate when the session ends. With it, testing knowledge compounds: session twenty inherits everything sessions one through nineteen learned. This is the single highest-leverage habit-skill in the library, and the purest expression of the principle that agent work should leave durable artifacts.

## What You Need


## Prompt / Setup
```xml
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
