# Stakeholder Management

Finding the people who matter, building the right communication channels, and keeping leadership confident — at both project and program scope.

## Contents
- The two halves of a communication plan
- Communication types (the tiered ladder)
- Communication timing (natural accountability)
- Finding stakeholders
- The stakeholder list
- RACI (roles and responsibilities)
- Anatomy of a status report
- Traffic-light discipline
- Path to green
- Above the fold
- Tech-landscape wrinkles
- Program-level: kickoff, leadership syncs, intervention

## The two halves of a communication plan

Stakeholder management is **managing expectations**. A communication plan has two parts:
1. **Communication types** — defined once per team/company and reused; mostly static templates.
2. **Stakeholder list** — varies per project, kept live.

Define types first (for consistency of format/cadence), then identify stakeholders, then make minor tweaks. Consistency means stakeholders learn your format and stop calling unnecessary meetings.

## Communication types (the tiered ladder)

Each tier serves a different audience and *builds on the tier below it*:

| Type | Goal | Cadence | Owner | Audience |
|---|---|---|---|---|
| Standup | Day-to-day collaboration, unblocking | Daily | SDE lead | Dev team |
| Status update | Milestone-level project status | Weekly | Project TPM | Working stakeholders |
| Leadership review (LR) | Longer-term program health, key insights | Monthly | TPM/PM-T lead | Leadership |
| Senior leadership review (SLR) | End-goal trajectory, big wins/misses | Quarterly | TPM/Director | Senior leadership/sponsors |

Use the **satellite-zoom** analogy to tailor depth: standup = street level (individual buildings); LR = neighborhood; SLR = whole city. Names vary (Amazon calls the LR a Monthly Business Review); some companies use meetings or slides instead of written reports — the *intent* is constant.

- **Standup**: the TPM should generally *not* run it long-term (it's not where your value lies); a scrum master should. Bootstrap it if needed, then hand off. Attend to know task progress and harvest material for the status report.
- **Status update**: TPM-owned, more technical, for people actively working the project (incl. SDMs not at standup, and vendors).
- **LR**: longer-term health, risks/issues/milestones with bigger impact — *not* the place to first surface a blocker (unblock immediately; don't wait for the review). Format is often company- and even leader-specific (be ready to rewrite for a new leader's style — that adaptability is the job).
- **SLR**: combines the quarter's LRs, emphasizing end goals and major wins/misses; strip noise irrelevant to senior leadership.

## Communication timing (natural accountability)

Sequence comms so each *feeds* the next, minimizing data-wrangling. Example: weekly status on Tuesday, LR on the third Wednesday of the month, every third LR is an SLR on the same day. Avoid Friday (glossed over before the weekend) and Monday (holidays push it to Tuesday, feeling inconsistent) sends — though honor any team-specific required day. At program scope, ensure each project's status/LR lands *before* the program LR/SLR; if something changes between, update affected stakeholders out-of-band so there are no surprises.

## Finding stakeholders

Not a one-time step — it recurs as the program evolves. Anyone impacted, however tangential, is a stakeholder. Find them through:
- **Requirements refinement** — reveals who's impacted (service owners).
- **Talking to known stakeholders** — they know clients of their services you don't.
- **Execution / low-level design** — surfaces stakeholders you missed; team members come and go.

**First 90 days on a project**: finding stakeholders is among your most important tasks. Ask your manager for the need-to-know list, meet each one, and ask each who *else* matters — build the network beyond the current project.

## The stakeholder list

Track per stakeholder: name, alias, department, project(s), role, and **which communication type** they receive. Keep it live — it's one of the most-referenced artifacts (used daily). Communication tiers are *inclusive downward*: the dev team that gets weekly status also sees how you represent the project upward. Note that one stakeholder (e.g., a division VP) may span several projects, so a project TPM knows that VP cares about cross-OS-impacting risks. Use the list to keep email distribution lists / directory groups current.

## RACI (roles and responsibilities)

A **RACI** chart assigns each step a role: **Responsible** (does the work / delivers it), **Accountable** (bears consequences if it isn't done), **Consulted** (input but not on the hook — e.g., developers during sprint planning), **Informed** (only needs the outcome — e.g., business teams on designs). Like the comms plan it's mostly static (it maps to *roles*, not people). Notes:
- One person can be both R and A (TPM for the project plan and status report).
- Responsibility shifts across the TPM/SDM/PM-T overlap per project — *verify roles each project*, don't assume.
- Distinguish R from A: the TPM is *accountable* for delivering a milestone on time; the developers are *responsible* for the development.
- Add **optional** (or parenthesize I/C) rather than leaving a cell blank (blank = open to interpretation).

## Anatomy of a status report

Whether a doc, email, or deck, include:
- **Executive summary** — high-level, leadership-focused; includes the traffic-light status and what's driving its color.
- **Project milestones** — the fixed phases (good cross-project metrics).
- **Feature deliverables** — the project-specific epics (when to expect each feature).
- **Open and recently resolved issues** — track open work separately from resolved (move resolved to a separate table to keep focus).
- **Risk log** — high-risk items inline, with a link to the full log.
- **Contact information**, **project description** (people forget what the project is), and a **glossary** (acronyms abound; same term can mean different things to different teams — saves churn and earns goodwill).

What turns okay into great: every action has an **owner and date**; a **defined traffic light**; a **path to green** for any non-green status; **important details above the fold**; and **clean format and grammar** (the less time spent decoding, the more it gets read). Spend time up front designing a format that refreshes easily (milestone tables, issue/risk tables), and run it by stakeholders. **Pre-communicate** any issue to the affected stakeholder *before* it appears in the report.

## Traffic-light discipline

Define your colors and reiterate them (preconceptions vary by culture and prior company):
- **Green** — meeting milestones, on track to deliver at current scope/resourcing (some add "on budget").
- **Yellow/amber** — active issues *may* impact the date; not certain to miss, but needs change to get back on track. Issues here should rarely be surprises — they should already be in the risk log. Keep yellow short-lived.
- **Red** — no current way to hit the date with current resourcing and scope. Not "impossible," but "something *must* change."

Bias to **transparency** — never ship a **watermelon** (green outside, red inside). When an event's color impact is genuinely unclear, use the definitions as your guide; a mitigated risk that didn't affect delivery stays green. Color-change negotiations can get political — clear definitions defuse them. And don't let an "always green" ideal demoralize you: fast, hot projects run yellow; turning green at the last possible moment to launch is fine.

## Path to green

Any yellow or red carries a **path to green (PTG)** — the actions to get back on track, each with owner and date.
- **Yellow PTG** is usually well-understood (that's why it's yellow). If an issue's impact needs investigation, *yellow* is often more appropriate than red (red asserts more certainty than you have); use a **date for a date** and define a **point of no return** — if investigation runs so long the date is unhittable regardless, flip to red.
- **Red PTG** is often exploratory: understand impact, then add resources (crash), de-scope/reprioritize to later phases, or — worst case — move the date (may be blocked by customer/regulatory commitments).

## Above the fold

Borrowed from newspapers: only the top half (no-scroll content) is guaranteed to reach the reader, and execs often stop there. Put **status color + target date, executive summary, and path to green** first. Below the fold: granular detail (sprint schedule, burndown, full risk/issue logs) for team leads and SDMs. At the bottom: near-static info (project description, contacts, comms plan). For critical projects, repeat the next-status date above the fold too.

## Tech-landscape wrinkles

- **Communication systems** — a sprawl of tools (Slack, Teams, Chime, Zoom; Quip, OneNote) and hybrid work mean you may need to track *which tool* each stakeholder uses, and standardize tools per project to cut churn. For mass-communication campaigns (thousands of clients/tickets), route everyone to a single monitored channel rather than per-ticket comments.
- **Tooling** — dashboards give always-on status, but controlling *how* an update is consumed (your narrative) helps manage expectations and prevent panic; aggregate with a portfolio tool, then craft the message yourself. Verify **permissions** (everyone who needs access has it; non-stakeholders are kept out).
- **Technical vs non-technical stakeholders** — your superpower is bridging both. Know your audience; lightly assess each stakeholder's technical and business fluency (legal teams may be neither). Re-articulating a problem for a different audience often clarifies it for you too.

## Program-level: kickoff, leadership syncs, intervention

At program scope it's about **scale** — more stakeholders, more concurrent comms (a standup, status, and LR *per project*). You're not *responsible* for every one but you're *accountable* for them.
- **Kickoff** — all stakeholders align on the program goal (not just their project goal), the comms plan, and each role; project leads see cross-project impacts; new stakeholders surface.
- **Leadership syncs** — a scrum-of-scrums: round-robin each project TPM for quick status/blockers, then "parking lot" any big issue so the unaffected can leave. Keeps you informed *and* keeps project TPMs aware of each other.
- **The art of intervention** — the program TPM is a force multiplier, not a controller. Delegate project issues to the project TPM; step in only for high-score or realized risks that impact the program, to speed cross-project coordination and communication. Keep project syncs on your calendar so you can drop in without invite churn, and give a project TPM a heads-up before their standup so an issue gets addressed fast. Intervention is a lending hand, rarely a takeover.
