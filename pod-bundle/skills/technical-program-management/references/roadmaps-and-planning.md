# Roadmaps and Planning

Turning ambiguous requirements into a defensible plan, and keeping it alive as reality intervenes.

## Contents
- The planning sequence
- Requirements refinement
- Use cases
- Task breakdown and estimation
- Buffers (and the PERT vs critical-chain debate)
- Assembling the plan: predecessors, crashing, lags
- Milestones vs feature lists
- The triple constraint
- Tooling reality
- Capacity-constrained prioritization
- Team overhead
- Program vs project planning
- Speeding planning up

## The planning sequence

Drive clarity in four steps. Each artifact traces back to the previous one (use outline numbering so traceability is mechanical):

1. **Requirements gathering & refinement** — the foundation; everything builds on it.
2. **Use cases** — story-form behaviors that span requirements and double as test items.
3. **Task breakdown with estimates** — from SMEs, then buffered.
4. **Assemble the plan** — task, requirement ID, duration, predecessors, crash count, start/end → Gantt.

Experienced TPMs collapse steps 1–2 and 3–4, but know all four explicitly. Don't plan in a closed room — bring stakeholders in; someone with prior experience in the domain (e.g., a past P2P or messaging system) will see gaps you can't.

## Requirements refinement

The single highest-leverage clarity work. Turn vague high-level requirements into specific, testable ones:

- **Kill vagueness.** "Standard UX elements should be available" → enumerate them (address book with add/remove/import/export, user profile with image/alias/bio, presence indicator, access control with accept/block).
- **Split compound requirements.** "All messages accessible until the user deletes a message" bundles a visibility requirement and a deletion exception — split for traceability.
- **Name implicit constraints.** "Create a P2P system" → "a P2P system *with no central servers*" if the goal is to remove centralized overhead.
- **Specify data semantics.** "Send text messages" → support all Unicode incl. emoji; support rich-text formatting (bold/italic/underline/font/size).

The point isn't perfection — it's removing room for divergent assumptions before design and estimation amplify the error.

## Use cases

Write each as **"As a [user/admin/seller/customer], I can [action]."** A use case may cover several requirements and tells a story about system behavior; it also *is* a test-plan item. Trace each back to the requirement IDs it satisfies. Order by story coherence (group under themes like "user profile" or "messaging"), not by requirement number.

## Task breakdown and estimation

Break use cases into implementation/design tasks with effort estimates (in hours/days/weeks of *effort*, not calendar time). Estimates come from the SMEs doing the work; a TPM with the right background can do a first pass and then review with the team so everyone agrees what they're committing to.

Two facts shape estimation:
- Estimates are made as if the task were the *only* thing the engineer is doing — a measure of complexity, not of the working environment. Accounting for the environment (interruptions, overhead) is the TPM's job, via buffers.
- "Effort" excludes overhead like waiting on code review or approvals — treat those as separate tasks or time buffers based on average turnaround.

## Buffers (and the PERT vs critical-chain debate)

No estimate matches reality, so add buffers. A practical approach is a **buffer matrix** keyed on three inputs:

| Task ambiguity | Estimate confidence | Team overhead | Buffer |
|---|---|---|---|
| High | Low | 10% | 40% |
| Medium | Low | 10% | 35% |
| Medium | Medium | 10% | 30% |
| Low | Medium | 10% | 25% |
| Medium | High | 10% | 25% |
| Low | High | 10% | 20% |

Mechanism: start from team overhead, add ~+15% for high ambiguity / +5% for low, and the reverse for confidence (high confidence +5%, low +15%). No buffer is ever 0% — estimates and reality never agree. Revisit your percentages whenever the team or project changes.

Two textbook alternatives worth knowing:
- **PERT**: weighted average of optimistic, most-likely, and pessimistic estimates: (O + 4M + P) / 6. Accurate but asking three estimates per task is heavy — the matrix lets you buffer a single estimate.
- **Critical chain** (Goldratt): buffer only items *on the critical path*, and beware that padding every task can invite procrastination (work expands to fill the estimate). Whether per-task buffers help depends on team culture.

## Assembling the plan: predecessors, crashing, lags

Capture per task: ID (outline-numbered so tasks group into stories/features), requirement ID, duration, **predecessors**, **crash count**, resourcing, start/end dates.

- **Predecessors**: finish-start (FS) is the default and most common; also start-start (SS), finish-finish (FF), start-finish (SF). Use **lags** to fast-track — e.g., implementation can start once the API definition is ready, before the whole design is done (model that as an FF on the API plus an FS on the rest).
- **Crashing** = putting multiple people on one task to compress calendar time (also called *swarming* in agile). The crash count is the *max* parallelizable resources, not the planned count. Capture it up front so you can react fast when a task is at risk — only possible when you understand the code structure (which packages can be worked in parallel). A 4-week task crashed to 2 resources finishes in ~2 calendar weeks.

The plan is a **living document**: it starts as the planned path, then accumulates *actuals* (actual start/finish/resourcing) as the project runs.

## Milestones vs feature lists

Two different health measures, often conflated:
- **Milestones** — frequently the *same* across all projects (design complete, implementation complete, UAT complete, launch). Good for comparing project health across a program (e.g., "how often does deployment run long?").
- **Feature list** — project-specific deliverables, usually the top-level (epic) tasks/stories, each independently shippable (e.g., "presence object," "send text message"). This is what most people *mean* when they say "milestone," and it's how you tell stakeholders when to expect a given capability.

Maintain the feature list separately even though it's derivable from the plan — most stakeholders won't read a full Gantt.

## The triple constraint

Scope, time, and resources are bound like a triangle, with **quality** in the middle. Change one side and another must move to keep quality. Implications:
- Removing people → push the timeline out, shrink scope, or both.
- Adding scope (change request or scope creep) → add time, add resources, or both.
- "Quality product with fewer resources and less time" = hidden overtime or cut corners (testing, requirements, etc.).
- Interdependencies cap your options: a task that *can't* be parallelized won't go faster with more people — only the timeline can move.

## Tooling reality

There is **no standard tool** — choice comes down to preference and company constraints (security/cloud-data, budget, mandated ecosystem). The one tool used at *every* big company is **Excel** (a blank canvas: zero learning curve, infinitely customizable via formulas/macros, but lots of manual work). Beyond that:
- **MS Project** — feature-defining for single projects but weak at portfolio/cross-project work.
- **Portfolio tools** (Smartsheet, Clarizen, Asana, ClickUp, Coda, Monday.com) — handle programs/portfolios and cross-project dependencies; some auto-update cross-project dependencies on a slip (a useful forcing function).
- Pure task/scrum-board tools are *not* program tools — they lack risk logs, milestone tracking, and proper resource management.

Most tools assume **projectized, non-crashing** resourcing: adding resources reduces each person's hours rather than shortening duration, and overbooking is often hidden. Expect to drop to Excel/Gantt and move tasks by hand when you need to model crashing or answer "what dates can I hit with X people?"

## Capacity-constrained prioritization

Tech orgs rarely fund a project for its whole life with fixed resources. Instead they run **capacity-constrained prioritization** — re-allocating existing capacity to the most strategic work monthly/quarterly/yearly (a specialized PDCA loop). Your resourcing can go up or down mid-project, forcing replanning. Counter by front-loading: pre-compute crash counts, buffers, and predecessor tightness so you can re-plan in minutes when capacity shifts, and so you can answer leadership's "what could you deliver with x/y/z resources?"

## Team overhead

Subtract non-project time before promising dates. Overhead runs **10–40%** depending on the team:
- **On-call rotations** — usually a week per rotation; the engineer is largely unavailable for project work. Worse than vacation because it hops person-to-person and is easy to forget. For short-staffed teams it can cost a full headcount-week per month.
- **Training / hackathon / improvement weeks** — whole teams can vanish at once when a popular course or hackathon lands.
- **Company/org-wide freezes** — reliability or security windows, or external events (sales events, launches) that block deployments.
- **Team meetings** — standing meetings; the one bucket people *do* remember, so probe for the others.

Model overhead either as a generalized buffer or as explicit tasks in the plan (explicit tasks explain gaps in the Gantt better but cost more upkeep). A 10-minute chat with the SDM walking the overhead list gives a rough number; buffer it until you have data.

## Program vs project planning

Same tools and steps, larger scope. Differences:
- The **program goal** is itself a requirement that must be refined like any other ("100% user reach" → reach of what? measured how? what about devices with no public IP needing a relay?). Probing two words of a goal can surface issues that must be resolved before projects start.
- Planning **iterates** as projects open and close (a multi-year program revisits planning repeatedly), vs a one-time project plan.
- You manage **cross-project dependencies** and look for **cross-project optimizations** — shared code, consistent designs, a carved-out shared subsystem. Evaluating technical avenues (e.g., a cross-platform framework so work done once is reusable) to increase plan efficiency is a signature TPM contribution.
- Each project lead maintains a plan that rolls up into the program plan; you oversee with input from project TPMs/PMs/SDMs.

## Speeding planning up

When time is short, attack the repeatable work:
- **Repeatable high-level estimates** — keep a list of recurring tasks (config, data modeling) with effort *ranges*; use the high end for complex instances. Also a great backlog for automation candidates.
- **Management checklists** — standardize the repeated process steps (review requirements, build use cases, analyze risks, build the comms plan).
- **Plan templates** — pre-populate the always-present tasks (requirements verification, functional spec, deployment, integration/E2E testing, launch plan).
- **Buffers** — apply the matrix instead of re-deriving per task.
