# Metrics and Program Health

How to measure whether a program is succeeding, and where each measurement belongs.

## Contents
- Metrics vs KPIs vs telemetry vs OKRs
- Where metrics belong (by audience)
- Page-0 metrics
- Traffic-light status as a health signal
- The risk scorecard as a health signal
- Choosing the right metric
- Program vs project health

## Metrics vs KPIs vs telemetry vs OKRs

- **Metric** — any measurement of something (e.g., API transactions per second).
- **KPI (Key Performance Indicator)** — the metrics that measure success against a goal, plus health of existing features (an API's TPS vs a target TPS; a product's page hits, click-through rate, installs).
- **Product telemetry** — in-product usage data (crash reports, feature-usage stats) when enabled.
- **OKRs** — objectives (the goal) with key results (the measurable outcomes). The program goal *is* the objective; KPIs are how you know you hit the key results.

Every program is measured by metrics/KPIs that map to its goal. The program charter's problem statement and business case usually *suggest* the right metrics — derive them there.

## Where metrics belong (by audience)

Match the measurement to the communication tier:
- **Standup / weekly status** — generally **no program metrics**. Program-level metrics don't move enough week-to-week to be worth tracking there, and the audience cares about day-to-day work, not overall health. If a metric *does* change that fast, it's probably the wrong measurement and needs rework. (Status reports do use *progress* signals like a burndown.)
- **Leadership review / senior leadership review** — this is where metrics live. Leadership wants longer-term health and goal trajectory, expressed in KPIs.

## Page-0 metrics

At Amazon, the handful of metrics that measure the program's success go on **"page 0"** — *before* any other content, including the status itself. This signals how central measurement is: leadership sees "are we hitting the numbers that justify this program?" first. Put your most important KPIs ahead of everything in exec reviews.

## Traffic-light status as a health signal

The red/yellow/green status is the most-consumed health signal. Define it and stick to it (full discipline in `stakeholder-management.md`):
- **Green** — on track to deliver at current scope/resourcing.
- **Yellow** — active issues *may* impact the date; recoverable, needs change.
- **Red** — cannot hit the date without a change.

Any non-green status carries a **path to green** (actions, each with owner and date). Bias to transparency — never a "watermelon" (green outside, red inside). The color is your fastest, most legible program-health communication.

## The risk scorecard as a health signal

Risk score = probability + impact on a 3-tier scale (see `risk-and-dependency-management.md`). The aggregate risk picture is a leading health indicator: a rising high-score risk warns of a coming yellow/red before it lands. Track leading indicators per risk (e.g., pipeline deploy counts/failure rates) so health degradation is visible early enough to act.

## Choosing the right metric

- A metric that swings weekly is suspect — it may be measuring noise, not the thing you care about.
- Prefer metrics tied directly to the goal and to the business case (they're defensible and meaningful to sponsors).
- Express program value in terms leadership uses: not just a raw number but its meaning (a 10× transaction-capacity goal isn't just "10×" — it's the systems changed, the clients enabled, the revenue/adoption/usage unlocked). Dollar figures, adoption rates, and usage rates all indirectly measure the *breadth of impact* of the program.

## Program vs project health

Health rolls up the same way comms do. Project-level reports surface project milestones, feature deliverables, and project risks; program-level reports surface only the items that affect the program's critical path and end goal, plus the page-0 KPIs. A program's critical path is the union of its projects' critical paths plus any program-level work. Keep weekly project status feeding the monthly program review so the rolled-up health is current.
