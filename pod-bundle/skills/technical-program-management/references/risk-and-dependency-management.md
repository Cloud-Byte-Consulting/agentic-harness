# Risk and Dependency Management

Identifying what might derail the plan, deciding what to do about it, and managing the cross-team dependencies that are a program's biggest risk source.

## Contents
- Risk vs issue
- The five-step risk cycle
- Identifying risks
- Analyzing risks (scoring)
- The four strategies
- Tracking and documenting
- The risk register and tribal knowledge
- Tech-specific risk categories (by SDLC stage)
- Cross-step risks
- Cross-project dependencies and program-level risk
- When risk work has to be quick

## Risk vs issue

A **risk** is something that *might* happen and would change the plan (scope, timeline, or resourcing) — including *positive* risks (opportunities, e.g., a vendor bulk discount that also adds delay). An **issue** is a realized risk — it has happened. Don't conflate them: if you treat risks as "issues that already caused a slip," you stop planning ahead. A well-stocked risk register means a realized risk is never a surprise — you already have a strategy.

## The five-step risk cycle

Continuous throughout the project, not a one-time phase:

1. **Identify** → 2. **Analyze** → 3. **Plan strategy / update plan** → 4. **Track** → 5. **Document**

Risk assessment starts as soon as you have enough information — in a familiar domain, a one-paragraph goal is enough; in an unfamiliar or vague one (Brexit-style "we will leave the EU" with no details), you may need a full requirements doc first (and "the requirements are too vague to assess" is itself a risk).

## Identifying risks

- **Risk register** — a searchable repository of risks from past projects, filtered by shared traits (same system/team/client/complexity). Rarely a real database; if you can build one, do.
- **Kickoff with all stakeholders** — multiple perspectives = combined experience (two people with 15 years each give you 30 years of pattern-matching). Standard in PMP and worth it: the earlier a risk is found, the cheaper it is.
- **The plan itself** — overlapping parallel tasks reveal constraint risks. Example: ten interop tests running in the same two weeks assume nothing fails; one bug (Windows↔Linux) can block every other Windows and Linux integration.
- **Stakeholder input** — they bring perspectives and their own registers; a single conversation can surface an overlooked team that's critical to the data flow, saving months.

## Analyzing risks (scoring)

Rate **probability** and **impact**, both somewhat subjective and context-dependent (the same risk has different probability/impact on a new project than an old one). Use a **3-tier scale** (Low/Med/High) — 4–5 tiers invite nitpicking ("High vs Very High"). Compute the **risk score as the *sum* of probability + impact** (summing two independently-honest scores removes the bias of guessing a combined score directly):

| Probability | Impact | Score |
|---|---|---|
| Low (1) | Low (1) | 2 |
| High (3) | Low (1) | 4 |
| Med (2) | Med (2) | 4 |
| High (3) | Med (2) | 5 |
| High (3) | High (3) | 6 |

The score sets how much attention a risk warrants: high score (high impact or near-inevitable) means watch closely and be ready to act. Where metrics exist (deploy counts, failure rates), use them to estimate probability/impact; where they don't (e.g., attrition), apply averages but remember statistics describe long-term trends, not the next event. Analysis can spawn tasks (a proof-of-concept or training to de-risk a new framework).

## The four strategies

Each risk can have several strategies, listed in the order you'd apply them (fall back as earlier ones fail):

- **Avoidance** — plan so the risk can't occur (don't use the known-bad vendor; train on a new framework *up front*). Often handled organically and never logged — but security threat-modeling logs even fully-avoided risks.
- **Mitigation** — let it happen (or it does) and soften the impact (crash/swarm remaining tasks to cut a 4-week delay to 2). Reactive; a good middle ground and why many companies default to it.
- **Transference** — move the risk to another party (contract a team that knows the framework). May still exist at the program level, and creates a new dependency (on the vendor).
- **Acceptance** — take the hit and adjust timelines, when avoid/mitigate/transfer are impossible or cost more than the impact. Sometimes failure (a slip) is the cheapest option.

Listing *multiple* strategies per risk and knowing how they fall back on each other is a hallmark of a strong TPM. Example layering for "cross-platform tooling issues" (High/High): Avoidance (up-front training) → Mitigation (resource crashing) → Acceptance (shift timelines).

## Tracking and documenting

- **Track** continuously — impact/probability change as the project moves (impact often drops near completion; probability can rise with attrition). Risks get avoided, expire, or newly appear (a published vulnerability forcing immediate patching). Build a **leading indicator** for each risk (pipeline health metrics: deploys/day, merges, time-to-production) so you act *before* realization, and watch other teams' launch dates that touch the same pipeline.
- **Document** into the project risk log, status updates, and the company register at close. Documentation is institutional memory — memory fades, especially when you switch teams. Keep your own register even if the company has none; one extra contributor improves everyone's odds, and the written risk often triggers the memory in the moment.

## The risk register and tribal knowledge

In the absence of a real register, lean on **tribal knowledge** (institutional memory): **explicit** knowledge (documentable) plus **implicit** knowledge (experience and hunches conveyed only through discussion). Review plans with peers to extract it. Even rough **thematic buckets/tags** (networking, cross-platform coordination, per-transaction-type) make a non-indexed register useful as a shortcut.

## Tech-specific risk categories (by SDLC stage)

- **Requirements** — main risk is *not driving clarity*; vague requirements → scope creep downstream.
- **Design** — technical issues surface here: latency and availability targets for distributed/SaaS systems can add requirements and time; API/architecture/release-strategy decisions. Ask explicitly about hot topics (latency) so they're in the estimates.
- **Implementation** — engineer overhead is consumed here; track the buffer and rebase. Two commonly-missed risks: **time to production** (deploy effort is rarely in the estimate — a task isn't "done" until written, reviewed, *and* in production and tested: "done, done, done") and **security certification** (InfoSec review time, longer for new services, partly out of your control — treat like a vendor dependency).
- **Testing** — where unknown-unknowns bite (edge cases). Risk scales with test-framework maturity and system size; full coverage is impossible in vast distributed state spaces, and new features add new unknowns.
- **Evaluation/iteration** — no inherent risk, but vague feedback can seed scope creep if not clarified before becoming requirements.

## Cross-step risks

These hit resource availability from outside the project:
- **On-call shifts** — sickness/time-off can put an engineer on call more than planned.
- **Mandatory software updates** — version end-of-life upgrades that block progress until done.
- **Industry-wide vulnerabilities** — e.g., Log4Shell (Dec 2021): all-hands patching with no negotiation, projects just impacted. Hard to predict; calling it out as a risk is "crying wolf," but a healthy buffer absorbs it and stakeholders generally forgive delays from genuine industry catastrophes.

## Cross-project dependencies and program-level risk

The risk *process* is identical at program scope; the *focus* shifts from cross-task to **cross-project** dependencies, which need far more coordination. Two patterns:
- **Shared component** — carving a shared subsystem out of N projects removes duplication but creates N dependencies on it; if it slips, all N slip. Any change to its API contract can ripple to every consumer (a cyclical risk given the touchpoints).
- **Mutual interop** — every-OS-tests-every-OS creates a web where one project's delay (Windows) can delay all interop. Mitigate by re-sequencing tests or de-scoping the laggard to launch later.

Cross-project dependency risks are **owned by the program TPM**, who ensures strategies are followed inside the impacted projects and works closely with the shared-component lead. A *project*-level risk important enough to threaten the program is elevated to program visibility: the project TPM still owns it, but the program TPM watches closely and helps. Tools rarely support program-level risk well — workarounds include a "program project" to hold them, or transferring (duplicating) risks down to projects so each impacted team knows.

## When risk work has to be quick

- A searchable, well-tagged **register** is the biggest time-saver — copy identification, re-score probability/impact for context, reuse strategies.
- Thematic buckets beat starting from scratch.
- A single **working session** with stakeholders to walk requirements while identifying/analyzing/planning risks delivers a lot of value fast, and getting your head into the problem surfaces new risks not in the register.
