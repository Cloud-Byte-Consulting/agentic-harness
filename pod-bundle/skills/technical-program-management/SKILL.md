---
name: technical-program-management
description: >-
  Do the technical program manager (TPM) job well - drive cross-functional, cross-team
  technical programs end to end. Use when planning or running a program or project, building a
  roadmap, work breakdown, or Gantt; managing stakeholders and influencing without authority;
  tracking risks, dependencies, and the critical path; writing status reports, escalations,
  leadership/executive updates, RACI charts, or a program charter; defining metrics, OKRs,
  KPIs, or program health (red/yellow/green); coordinating launches, on-call, or incidents;
  running standups, sprints, or scaled/agile delivery; or deciding project vs program scope.
  Also triggers on phrases like "drive this program," "keep this on track," "who owns this,"
  "what's the status," "is this red or green," "path to green," "manage these dependencies,"
  "align stakeholders," "estimate this work," "TPM vs PM vs engineering manager," or TPM
  career growth and leveling. Practical and people-centric, not code-heavy.
---

# Technical Program Management

This skill equips Claude to perform the technical program manager (TPM) role competently: driving complex, cross-functional technical programs from ambiguous goal to delivered launch by combining project/program management discipline with enough technical literacy to be a credible bridge between engineering and the business.

## When to use this skill

- Planning or running a program or project: requirements, work breakdown, estimates, Gantt, dependencies, resourcing.
- Building or revising a roadmap; deciding whether something is a project or a program.
- Stakeholder management: communication plans, status reports, leadership/executive updates, RACI, kickoff.
- Risk and dependency management: risk register, scoring, mitigation strategies, critical-path and cross-project risk.
- Reporting program health (red/yellow/green), defining "path to green," escalating, or wording a hard message.
- Metrics, OKRs, KPIs, and measuring whether a program is succeeding.
- Launch coordination, on-call/overhead planning, incident or "all-hands" disruption response.
- Agile/scaled delivery: standups, sprints, scrum-of-scrums, leadership syncs.
- TPM career, leveling, and how the role differs from project manager, product manager, and engineering manager.

## What a TPM is (and is not)

A **program** is a set of related projects driving toward one shared goal; it is temporary and ends when the goal is met. A **project** is a single bounded deliverable. A **product** is a long-lived vertical (business + dev + the thing users use). The mental model: **products are vertical, programs are horizontal** - programs cut *across* products/teams to deliver something no single team owns. TPMs are most valuable on horizontal, cross-team initiatives.

The TPM stands on three pillars: **project management**, **program management** (project management at larger scope), and a **technical toolset** (system-design literacy, enough code fluency to read it, architecture-landscape awareness). The technical pillar is what separates a TPM from a generalist PM - it lets you challenge estimates, spot cross-system risk, and translate both directions between engineers and executives.

Distinguish the adjacent roles - do not duplicate them, fill the gap when one is missing:
- **TPM**: owns *delivery* of the program/project. Driven by "is it moving forward?"
- **Software/Engineering Manager (SDM/SEM)**: owns *people* and a *service/domain*. Can run domain projects but not cross-team complexity.
- **Product Manager (PM-T)**: owns *product* roadmap, strategy, prioritization.

A TPM's defining behavior is **driving toward clarity** and being a **force multiplier**: relentlessly removing ambiguity (about the problem, the solution, the owner, the next step) and unblocking others so the program keeps moving. Influence comes from clarity and trust, not authority.

## How to approach TPM work

### 1. Establish the charter and goal
Before planning, pin down *why* and *what*. Capture a problem statement, a business case, and a single clear goal (Amazon's PR/FAQ is one form; a one-page charter is another). The goal should bound scope without over-specifying *how*. A goal too vague ("100% user reach") needs probing; a goal too prescriptive locks out better solutions. Decide **project vs program** with three litmus questions: multiple distinct end goals? timeline fixed vs goal-driven? spanning multiple orgs/teams? More "yes" leans toward a program.

### 2. Drive clarity from requirements to a plan
Refine vague requirements into specific, traceable ones, then into **use cases** ("As a user, I can..."), then into **tasks with estimates**. Break compound requirements apart for traceability. Build the plan: task list, durations, predecessors (finish-start is default; also start-start, finish-finish, with lags to fast-track), resourcing, and a Gantt. Note where tasks can be **crashed** (parallelized) ahead of time so you can react fast later. Then **add buffers** - real estimates never match reality. See `references/roadmaps-and-planning.md`.

### 3. Manage the triple constraint
Scope, time, and resources are bound together (the project-management triangle); change one and another must give, or quality suffers. "Faster with less" almost always means hidden overtime or cut corners. Most tech orgs use **capacity-constrained prioritization** (resources re-allocated each cycle), not fixed project funding - so front-load planning (crash counts, buffers, predecessors) to replan quickly when capacity shifts.

### 4. Run continuous risk and dependency management
A risk is something that *might* happen and change the plan; an issue is a realized risk. Continuously: identify (register, stakeholders, plan) → analyze (probability + impact → score) → plan a strategy → track → document. The four strategies: **avoidance, mitigation, transference, acceptance** (a risk can have several, ordered by which you'd try first). At program scope, focus on **cross-project dependencies and critical path**. See `references/risk-and-dependency-management.md`.

### 5. Communicate relentlessly and tailor to audience
Set up a layered communication plan keyed to *audience*, not just cadence: standup (daily, dev), status report (weekly, working stakeholders), leadership review (monthly), senior leadership review (quarterly, execs). Sequence them so each feeds the next. Tailor depth: execs want trajectory and risk; engineers want the day-to-day. **Over-communicate** and be consistent - missed or late status erodes trust. Every action needs an **owner and a date** (use a "date for a date" if unknown). See `references/communication-and-influence.md` and `references/stakeholder-management.md`.

### 6. Report health honestly
Use a defined **traffic light**: green = on track at current scope/resources; yellow = at-risk but recoverable; red = cannot hit the date without a change. Any non-green status needs a **path to green** with owner and date. Bias to transparency - never ship a "watermelon" (green outside, red inside). Keep the most critical info **above the fold**. See `references/communication-and-influence.md`.

### 7. Use the technical toolset as leverage
Read code well enough to follow a method/API signature and challenge estimates; read and critique **system designs** (know the common architectural patterns and the latency/availability/scalability levers); hold a mental **architecture landscape** of how systems connect. This lets you find cross-system risk early, write a credible functional spec, and arbitrate technical disputes. Push depth into the engineers; your job is the connective view. See `references/technical-depth-for-tpms.md`.

### 8. Measure what matters
Programs are judged by metrics/KPIs/OKRs that map to the goal - put them on "page 0" of leadership reviews. Status reports rarely need metrics (they don't move weekly); leadership/senior reviews do. See `references/metrics-and-program-health.md`.

## Common pitfalls & anti-patterns

- **Treating waterfall stages as sequential silos.** Every stage is a mini-cycle that feeds the others; plan, risk, and stakeholder work run continuously in both waterfall and agile.
- **Skipping clarity to "save time."** Ambiguity surfaces later as scope creep, bad estimates, and rework. Clarifying up front is cheaper. Distinguish true scope creep from planned iterative/greenfield exploration.
- **Gold-plating and unmanaged derived requirements.** Features nobody asked for, or dev-added scope, still consume the triangle. Surface them through change management.
- **Reporting a "watermelon" status or inconsistent cadence.** Both destroy stakeholder trust faster than an honest red.
- **Action items with no owner or date.** Untracked work slips silently.
- **A senior TPM running standups long-term.** Help bootstrap, then hand off to a scrum master; your leverage is strategic, not facilitating dailies.
- **Designing/changing a system without checking upstream and downstream.** The classic failure: rebuild a storage/service model from the local team's pain points, ignore who produces and who consumes the data, ship something that misses half the real requirements.
- **Forgetting team overhead.** On-call rotations, training, hackathon weeks, deploy freezes, and meetings can eat 10-40% of capacity. Bake it into buffers.
- **Confusing the TPM with the PM-T or SDM.** Cover gaps short-term, but the value of each role is distinct; long-term role confusion hurts delivery.

GenAI can accelerate the administrative TPM tasks (drafting status reports, summarizing meetings, first-pass risk lists, planning artifacts) - treat it as a productivity tool, verify its output (check cited sources, obfuscate confidential data before pasting), and lean harder on the emotional-intelligence soft skills (judgment, empathy, influence) it cannot replace. See `references/emotional-intelligence-and-ai.md`.

## Reference files

- `references/tpm-role-and-scope.md` - what a TPM is, the three pillars, role boundaries vs PM/PM-T/SDM, project vs program vs product, charters, specializations, and career path/leveling.
- `references/roadmaps-and-planning.md` - requirements → use cases → tasks → Gantt; estimation, buffers (PERT vs critical-chain), crashing, predecessors, the triple constraint, capacity-constrained prioritization, team overhead, and tooling.
- `references/stakeholder-management.md` - finding stakeholders, communication-plan types and cadence, RACI charts, kickoff, leadership syncs, technical vs non-technical audiences, and the first-90-days playbook.
- `references/risk-and-dependency-management.md` - the five-step risk cycle, scoring, the four strategies, the risk register, tech-specific risk categories, and cross-project/critical-path risk.
- `references/communication-and-influence.md` - status report anatomy, traffic-light definitions, path to green, above-the-fold, influence without authority, emotional intelligence, escalation, and hard conversations.
- `references/metrics-and-program-health.md` - KPIs/metrics/OKRs, page-0 metrics, what to report at which level, and measuring program success and health.
- `references/technical-depth-for-tpms.md` - code literacy, design patterns, system-design patterns and levers (latency/availability/scalability), architecture landscapes, functional specs, and system-design interview prep.
- `references/emotional-intelligence-and-ai.md` - the four EQ components and how to grow/assess them, applying EQ to decisions/risk-attitude/stakeholders, and using GenAI safely across the key management areas (with grounding, data-leakage, and prompt-engineering discipline).
