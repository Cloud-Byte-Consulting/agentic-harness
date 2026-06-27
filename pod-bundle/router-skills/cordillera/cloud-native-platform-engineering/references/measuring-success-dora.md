# Measuring Platform Engineering Success

DORA, SPACE, and DevEx metrics; flow; the cloud-native maturity model; the
outputs→outcomes→impact progression; leading vs lagging indicators; and advanced
platform KPIs including adoption.

## Contents
- Outputs vs outcomes vs impact
- Productivity myths to debunk
- DORA metrics
- SPACE framework
- DevEx metrics
- Flow
- Leading vs lagging indicators (and DORA's limits)
- Cloud-native maturity model
- Advanced KPIs and adoption

## Outputs vs outcomes vs impact

Measure the *progression*, not just activity:
- **Outputs** — direct deliverables (features built, bugs fixed, services migrated,
  platform tools adopted). Necessary for tracking, but don't prove value.
- **Outcomes** — the change/benefit the outputs produce (faster time-to-market, higher
  deployment frequency, better reliability, more engagement). A better measure of
  effectiveness.
- **Impact** — long-term business/community effect (revenue, customer satisfaction, market
  share, business agility).

As a platform matures, shift the focus of measurement **outputs → outcomes → impact**.
Early on, count services migrated / adoption (outputs); later, measure time-to-market and
reliability (outcomes); ultimately, business growth (impact).

## Productivity myths to debunk

- Productivity is *not* just activity (lines of code, tasks done ≠ value).
- Productivity is *not* purely individual — it's a collective outcome of team dynamics,
  tools, and process. Overweighting individual metrics breeds competition over collaboration.
- *No single metric* tells the whole story — combine metrics across dimensions.
- Productivity measures help *developers*, not just managers (self-improvement).
- Productivity depends on culture/communication/alignment, not only tools.

## DORA metrics

Four standardized software-delivery metrics:
- **Deployment frequency** — how often you deploy to production.
- **Lead time for changes** — commit → production.
- **Change failure rate** — % of deployments causing failures/rollbacks/hotfixes.
- **Mean Time to Recovery (MTTR) / time to restore** — how fast you recover from an
  incident.

Excellent for assessing delivery efficiency/reliability — but mostly **lagging** indicators
(retrospective; no predictive insight). Complement with leading indicators and other
frameworks.

## SPACE framework

A holistic productivity model combining leading + lagging indicators across five dimensions:
**S**atisfaction & well-being, **P**erformance, **A**ctivity, **C**ommunication &
collaboration, **E**fficiency & flow. Use it so you don't over-rotate on one dimension
(e.g. activity) and to balance throughput with developer well-being.

## DevEx metrics

Developer Experience metrics (from the SPACE creators) capture day-to-day friction:
- **Usability & tooling** — tool usability (surveys, usability testing, adoption rates),
  integration quality, onboarding time.
- **Workflow efficiency** — cycle time, lead time, context-switching frequency/impact.
- **Satisfaction & well-being** — NPS, feedback scores, burnout indicators.
- **Collaboration & communication** — collaboration-tool usage, communication patterns,
  cross-team interactions.

Improve via baselines → regular feedback loops → iterative changes → transparent
communication → benchmarking.

## Flow

Flow is the smooth, uninterrupted progression of work from idea to deployment. Metrics:
**cycle time**, **lead time**, **throughput**, **Work In Progress (WIP)**. Improve flow with
value-stream mapping, WIP limits, CI/CD, Kanban boards, and feedback loops. Benefits:
efficiency, quality, developer experience, predictability.

## Leading vs lagging indicators (and DORA's limits)

- **Leading** (predictive): e.g. cycle time — early signal of an improving lifecycle.
- **Lagging** (reflective): e.g. MTTR, change failure rate — results of past actions.

DORA is primarily lagging; pair it with leading indicators (and SPACE/DevEx) for a complete,
predictive picture.

## Cloud-native maturity model

Five levels; IDPs typically become effective at the **Scale** level:
1. **Build** — baseline: containerization, basic orchestration, basic CI/CD; learning/
   experimentation; goals: solid foundation, minimize tech debt.
2. **Operate** — production-ready infra, advanced orchestration (autoscaling, self-healing),
   security in pipelines (RBAC, network policy, secrets), operational visibility; goals:
   reliability, security.
3. **Scale** — standardization, cost optimization, enhanced automation, resilience/
   scalability; **platform engineering makes a big difference here** and drives down both
   legacy and cloud cost.
4. **Improve** — advanced security, policy enforcement (policy-as-code/OPA), governance
   frameworks, continuous compliance; balance security with agility.
5. **Adapt** — continuous monitoring/optimization, feedback loops, innovation/
   experimentation, agile/adaptive; revisit earlier decisions and optimize.

## Advanced KPIs and adoption

Go beyond DORA/SPACE/DevEx with platform-specific KPIs:
- **Cost efficiency** — cost per deployment, infrastructure utilization.
- **User adoption & engagement** — **adoption rate** (% of developers using the platform),
  **active users**. Adoption is the platform-as-a-product north star.
- **Incident management** — incident frequency, incident resolution time.
- **Innovation & experimentation** — experimentation rate, feature adoption rate.

Implement by defining clear objectives, setting baselines/targets, collecting/analyzing
data, and reporting regularly. Surface these in a **platform dashboard** (e.g. Grafana) and
in **Backstage** via plugins so developers and leadership see real-time platform performance
and DORA/SPACE metrics. (For dashboard/metrics implementation, see kubernetes-observability.)
