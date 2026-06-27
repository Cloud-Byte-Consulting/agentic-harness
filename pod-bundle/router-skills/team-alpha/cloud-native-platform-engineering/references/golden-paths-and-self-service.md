# Golden Paths & Self-Service

Golden paths / paved roads, the golden-path maturity model, how opinionated to
make the platform, the developer portal (Backstage), GitOps as the platform
interface, and repository strategies.

## Contents
- Golden paths / paved roads
- The Golden Path Maturity Model
- Implementing golden paths by maturity level
- Worked examples (pipelines, observability, new apps)
- How opinionated should the platform be?
- The developer portal: Backstage
- GitOps as the IDP interface
- Repository strategies

## Golden paths / paved roads

A **golden path** (Red Hat/VMware term) or **paved road** (Spotify term) is the most
efficient, secure, supported way to accomplish a common task. On the paved road, the
platform supports you and "everything just works." Go off-road and you get less support.

Why they matter:
- **Consistency & standardization** — solutions share design, architecture, and patterns,
  simplifying maintenance and troubleshooting.
- **Accelerated development** — proven solutions to common problems; no reinventing the
  wheel.
- **Knowledge transfer** — new hires learn the org's best practices by following the path
  (the how *and* the why).

The risk to manage: don't let standardization stifle innovation. Provide escape hatches.

## The Golden Path Maturity Model

Adoption is a journey through three levels:

1. **No golden path** — each team has its own DevOps practices/tools. Pro: maximum
   flexibility. Con: duplicated effort, collaboration friction, more errors.
2. **Clone and forget** — central team provides templates/small tools to copy and adapt.
   Pro: some consistency, fewer tools, emerging best practices. Con: used inconsistently;
   not yet a cohesive platform.
3. **Golden paths as products** — paths are first-class entities baked into the IDP with
   self-service onboarding. Pro: maximum efficiency, consistency, scalability — best
   practices are built in, not just suggested. Con: significant upfront investment to build
   and maintain.

Golden-path maturity should progress in parallel with overall DevOps maturity.

## Implementing golden paths by maturity level

- **At "No golden path":** find glaring inconsistencies/redundancies; define a baseline
  path that describes the **why/outcome** and lets teams build their own tooling toward it.
- **At "Clone and forget":** expand to more workflows; provide centrally-maintained
  templates teams clone/adapt; add basic integration.
- **At "Golden paths as products":** design holistic solutions unifying tools under the IDP;
  provide out-of-the-box, self-service onboarding covering the key capabilities.

## Worked examples

**Pipelines (CI/CD).** No path → teams have different/missing pipelines. Clone-and-forget →
a central CI/CD template to copy and customize. As-a-product → a fully featured,
standardized-yet-customizable pipeline (parallel builds, canary, rollback) — e.g. a reusable
"golden pipeline" with Helm + Argo CD.

**Observability.** No path → divergent tools, missed metrics. Clone-and-forget → guidelines
on what to monitor/log/alert. As-a-product → a unified observability platform prescribing
dashboards, KPIs, and SLO-based alerting with consistent cross-team data. (See
kubernetes-observability for implementation.)

**Deploying new apps.** No path → tedious manual config, errors. Clone-and-forget →
central templates/wiki for containerization, env setup, DB init. As-a-product → seamless
scaffolding: template repos, pre-configured templates for web apps/microservices/batch jobs,
tool integration — best practices applied consistently.

## How opinionated should the platform be?

Opinionation is a **dial**, not a switch:

- **Highly opinionated (defined pathways):** uniformity, predictability, fast onboarding,
  consistent best practices — but rigidity can stifle innovation; one-size-fits-none;
  resistance to adoption; dependence on platform updates.
- **Highly flexible (ad-hoc pathways):** fosters innovation and rapid adoption of new
  tech — but causes **tech fragmentation**, operational overhead, inconsistent practices,
  and **analysis paralysis** (the paradox of choice).
- **Semi-opinionated (the sweet spot):** sensible defaults with overrides/escape hatches.

A common, costly misconception: *more flexibility eases adoption*. In practice excessive
flexibility drags the platform team back into fragmented, ad-hoc DevOps. **Optimize the
paved road for the majority**, cater to niche needs via escape hatches, and reassess with
feedback loops.

Evaluate before deciding: diversity of projects, existing tech stack/legacy, developer
preferences/expertise, operational bottlenecks, and future goals/roadmap alignment.

Measure DevX and adoption to tune the dial: surveys/NPS, feedback sessions, platform usage
metrics, and incident/support-request analysis.

## The developer portal: Backstage

**Backstage** (open-sourced by Spotify) unifies tooling, services, and docs into one UI so
developers stop navigating a maze of tools. Core features:

- **Service catalog** — central view of all services, ownership, status, APIs, docs;
  clarifies accountability.
- **Extensible plugin architecture** — teams build/deploy plugins for their needs
  (Kubernetes, CI/CD, monitoring, cost, etc.).
- **Standardized developer experience** — one consistent interface; lower learning curve.
- **Software templates** — scaffold a new microservice/library following best practices.
- **Built-in docs (TechDocs)** — docs live alongside code.

Register a service with a `catalog-info.yaml`:

```yaml
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: your-service
  description: "A description of your service"
  annotations:
    backstage.io/kubernetes-namespace: your-service-namespace
spec:
  type: service
  owner: team-name
  lifecycle: production
```

Different users want different interfaces: a developer prefers a **GitOps workflow** to
deploy; a manager prefers a **portal scorecard** to see service health, ownership, cost
insights, and DORA metrics. Backstage serves the latter and ties them together. Commercial
IDP/orchestrator alternatives include Humanitec, Port, Cortex, OpsLevel, Qovery; open-source
includes Backstage, Crossplane, Krateo. There is **no off-the-shelf one-size-fits-all IDP** —
org diversity, varied tech stacks/legacy, the need for customization, fast-evolving
practices, and unique security/compliance needs all defeat a fully packaged product.

## GitOps as the IDP interface

Treat the developer's primary "deploy" interface as **GitOps**: commit YAML to a repo,
Argo CD (or Flux) reconciles it into clusters. This is auditable, reviewable, reproducible,
and self-service. (Implementation in kubernetes-gitops-cicd.)

## Repository strategies

- **Centralized infra repo** — all IaC in one repo. Pro: consistency, easy oversight/
  approval, shared learning. Con: approval bottlenecks, repo becomes unwieldy at scale.
- **Decentralized team repos** — each team owns app + IaC. Pro: speed/autonomy, focused
  context, resilience. Con: inconsistency risk, coordination overhead, duplicated effort.
- **Hybrid (common end-state)** — a central repo sets global standards and core infra
  (e.g. requesting a new cluster); decentralized repos let teams move fast within guardrails.

Choose based on org size, culture, and operational needs.
