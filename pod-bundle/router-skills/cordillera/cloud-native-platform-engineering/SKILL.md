---
name: cloud-native-platform-engineering
description: >-
  Strategy and architecture for cloud-native platforms and platform engineering. Use for
  platform-as-a-product thinking, internal developer platforms (IDPs) and their reference
  architecture on Kubernetes, golden paths and self-service, developer portals (Backstage),
  Team Topologies and how to staff and lead a platform team, platform capabilities
  (provisioning, CI/CD, observability, data, environments) and how they compose, FinOps and
  cost control, governance and guardrails-versus-gates, security and compliance as an enabler,
  measuring success with DORA metrics and developer experience, the cloud-native anti-patterns
  catalog and the right pattern for each, and migrating from legacy to cloud-native. Trigger
  whenever the user plans an IDP or platform team, designs golden paths or self-service, sets
  platform strategy or governance, tackles FinOps, or measures DevOps and platform success.
  For hands-on implementation see kubernetes-gitops-cicd, kubernetes-observability, and
  kubernetes-security-rbac.
---

# Cloud-Native Platform Engineering

This skill equips Claude to advise on the STRATEGY and ARCHITECTURE layer of
cloud-native platforms: deciding *whether* and *how* to build an Internal Developer
Platform, how to staff and lead the team, which capabilities to compose, how to drive
adoption, measure success, control cost, and avoid the classic anti-patterns. It is
deliberately more advisory than YAML-heavy — it tells you what to build and why, and
hands off the *how* to the sibling hands-on skills.

## When to use this skill

- Someone asks "do we need a platform team?", "what's an IDP?", or "platform engineering
  vs DevOps vs SRE".
- Designing an IDP reference architecture or choosing capabilities (compute, CI/CD,
  observability, data services, environment management) and how they compose.
- Defining **golden paths / paved roads**, self-service workflows, platform APIs, or a
  developer portal (Backstage).
- Org design: applying **Team Topologies** (platform / stream-aligned / enabling /
  complicated-subsystem teams), Conway's Law, cognitive load, staffing a platform team.
- **FinOps**, cost guardrails, tagging taxonomy, showback/chargeback, "bill shock".
- **Governance**: guardrails-vs-gates, RACI, policy-as-code, centralized-vs-decentralized,
  Cloud Center of Excellence.
- Treating **security/compliance as an enabler** (shift-left, guardrails, DevSecOps) rather
  than a release-blocking gate.
- Measuring success: **DORA / SPACE / DevEx**, adoption, cloud-native maturity, outputs vs
  outcomes vs impact.
- Diagnosing or avoiding **cloud-native anti-patterns** (lift-and-shift, distributed
  monolith, over-engineering, ignoring cost, security-as-afterthought, calcified governance).
- Planning a **legacy → cloud-native migration** (7 Rs, lighthouse app, transition
  architecture).

## Core concepts

**Platform engineering is the discipline of building and operating an Internal Developer
Platform (IDP) as a product** that other teams self-serve from. It is *not* a rebrand of
DevOps or SRE:

- **DevOps** is the goal/culture — fast, reliable delivery via shared ownership and
  automation, "you build it, you run it." It applies across the whole SDLC.
- **SRE** focuses on the reliability and uptime of *specific production services*
  (SLOs, error budgets, incident response, postmortems).
- **Platform engineering** is the *means*: it builds the shared, reusable internal platform
  and tooling so stream-aligned teams can do DevOps without each reinventing it. Think
  "DevOps scaled across the org" the way SAFe scales Agile. The platform team's customers
  are the other engineering teams — SRE is often a customer too.

**Platform-as-a-product mindset** is the single most important idea. The platform has
users (developers), a backlog, adoption metrics, a roadmap, and must earn its keep through
developer experience — not be imposed top-down. If you forget this, you slide back into
ad-hoc DevOps ticket-ops.

**Where the platform ends and products begin.** Layer it: Layer 0 cloud infrastructure
(owned by infra/cloud team), Layer 1 the IDP — runtime + pipelines + observability + data
+ security (owned by the platform team), and the product/business logic on top (owned by
stream-aligned teams). Keep boundaries clean and **API-driven** so the platform can evolve
without breaking products, and products can run even if a platform feature is unavailable.

**Golden paths (paved roads)** are the opinionated, supported, "everything just works" way
to do a common task (spin up a service, get a pipeline, get observability). Off the paved
road you get less support. They mature through three stages: *no golden path* (every team
DIY) → *clone-and-forget* (shared templates, used inconsistently) → *golden paths as
products* (baked into the IDP with self-service onboarding). See
`references/golden-paths-and-self-service.md`.

**Opinionation is a dial, not a switch.** A fully opinionated platform is rigid and
resisted; a fully flexible one fragments and overwhelms (the paradox of choice → analysis
paralysis → ad-hoc DevOps). Aim semi-opinionated: sensible defaults with escape hatches.

**Kubernetes is the substrate, not the platform.** >70% of IDPs are built on Kubernetes
because it gives unified orchestration, self-healing, a declarative control-loop API, and
an extensible CRD/operator model. But "we run Kubernetes" ≠ "we have a platform" and ≠
"we are cloud-native." The platform is the *developer-facing abstraction* on top.

## Workflow / how to approach platform-engineering tasks

**1. Confirm the tipping point before building.** A dedicated platform team is justified
when DevOps effort is duplicated across teams, cognitive load is crushing stream-aligned
teams, and there's a consistent set of needs to factor out. Below that, a shared-tooling
or embedded-SRE model may be cheaper. Tie the decision to business goals (time-to-market,
reliability, cost) and secure executive sponsorship — frame the platform as an *investment*,
not a cost center. See `references/platform-engineering-fundamentals.md`.

**2. Design the team using Conway's Law and Team Topologies.** The platform team is a
**platform team** (provides a self-service API to reduce others' cognitive load) that also
flexes into **enabling** (mentoring, evangelism) and **complicated-subsystem** (e.g. GPU/HPC)
modes. Stream-aligned teams own business domains end-to-end. Define a **platform team API**:
how others request support, contribute (innersourcing), get changes, report incidents, give
feedback. Place the platform team centrally between infra and product teams. Details and
the Fit-for-Purpose WHY/WHAT/HOW/WHO model in `references/team-topologies-and-org.md`.

**3. Scope and compose capabilities.** Pick a reference frame — **CNOE** (broad capability
map) or the prescriptive **BACK stack** (Backstage + Argo CD + Crossplane + Kyverno) — then
compose the capabilities your org actually needs:
- **Compute/runtime** (Kubernetes; Knative for serverless, KubeVirt for VMs; one-big-cluster
  vs cluster-per-app/env; hub-spoke multi-cluster via Cluster API).
- **Pipelines / CI-CD** (GitHub Actions + Argo CD GitOps; ApplicationSet for multi-cluster).
- **Infrastructure as Code** (Crossplane / operators to manage cloud + clusters as K8s
  resources; CRDs + controllers to extend the API).
- **Observability** (Prometheus/Thanos, Loki, OpenTelemetry; SLOs; avoid the "watermelon").
- **Data services / XaaS** (DBaaS via operators + GitOps — Atlas/ACK/Crossplane/Config Connector).
- **Self-service interfaces** (Backstage portal for discovery/scorecards; GitOps for the
  "deploy" workflow; centralized vs decentralized vs hybrid repos).
- **Built-in security** (scanning, SBOM, secrets/cert management, policy-as-code).
Full catalog with valid manifests in `references/internal-developer-platforms.md`.
For deep implementation, **cross-link the hands-on skills** (don't re-derive): GitOps/CI-CD,
observability, security/RBAC, cluster-operations.

**4. Treat governance and cost as guardrails, not gates.** Prefer *preventive* guardrails
(policy-as-code that stops bad configs at deploy, e.g. Kyverno/OPA, SCPs) and *detective*
guardrails (flag/auto-remediate after the fact) over manual approval boards. Decentralize
decision-making; keep a thin CCoE for standards. Establish a tagging taxonomy, budgets/alerts,
rightsizing, and a showback/chargeback model so product teams own their spend. See
`references/finops-and-governance.md`.

**5. Drive adoption and measure success.** Find a low-risk **lighthouse** team, deliver quick
wins, evangelize. Measure with **DORA** (delivery), **SPACE** (productivity+wellbeing),
**DevEx** (day-to-day developer friction), plus adoption/active-users and cost efficiency.
Track the progression **outputs → outcomes → impact** — don't celebrate "services migrated"
when what matters is time-to-market and reliability. See `references/measuring-success-dora.md`.

**6. When migrating from legacy, plan first.** Assess and classify workloads, pick the right
of the **7 Rs** (rehost/replatform/repurchase/refactor/retire/retain/relocate), choose
single/multi/hybrid-cloud deliberately, start with a lighthouse, and phase the rollout with
transition architectures (strangler-fig). Don't lift-and-shift and call it done. See
`references/migration-to-cloud-native.md`.

## Common pitfalls & anti-patterns

The full catalog with the correct pattern for each lives in
`references/anti-patterns-catalog.md`. The headline mistakes:

- **Lift-and-shift = cloud-native.** Moving a VM/DB unchanged keeps all the operational
  toil and can cost *more* than on-prem. Refactor toward managed/cloud-native services.
- **"Containers/Kubernetes solve everything."** Containerizing a monolith, or putting K8s
  under everything, isn't cloud-native and may add a steep learning curve for no benefit.
  FaaS or managed services are sometimes the better fit.
- **Distributed monolith.** Microservices that must deploy together / share a DB. Use DDD
  bounded contexts and Conway's Law deliberately; allow independent deploy + rollback.
- **Security as an afterthought.** Bolt-on security widens the attack surface. Shift left:
  IDE checks → pipeline scans (SAST/SCA/secret/vuln) → admission policy → runtime. Make
  security an *enabling* function with sensible-default artifacts, not a release gate.
- **Ignoring cost / no FinOps.** No tagging, no budgets, no chargeback → bill shock,
  no ownership. Tag, alert, rightsize, commit-spend, and give teams their bill.
- **Over-centralized / calcified governance.** A central approval board becomes a
  bottleneck. Decentralize with guardrails; keep a thin CCoE.
- **"Our business is too special for guardrails."** Almost never true; you'll scramble at
  audit time. Standardize, then allow well-defined exceptions.
- **Over-/under-opinionated platform**, **ignoring cultural change** ("the culture doesn't
  need to change"), **expecting learning to happen miraculously** (no training time),
  **minimum stakeholder commitment**, **measuring activity instead of outcomes**, and
  **environment-branch Gitflow drift** (prefer trunk-based + single artifact promoted across
  envs, with feature flags).

## Reference files

- `references/platform-engineering-fundamentals.md` — PE vs DevOps vs SRE, platform-as-a-product,
  the four pillars, platform/product boundaries, the tipping point, business case. Read when
  framing *whether* to invest or *what* platform engineering even is.
- `references/internal-developer-platforms.md` — IDP capabilities, CNOE & BACK stack,
  Kubernetes-as-substrate, CRDs/operators, Crossplane, Cluster API, multi-cluster patterns,
  data/XaaS. Read when designing the IDP architecture or choosing tools.
- `references/golden-paths-and-self-service.md` — golden paths / paved roads, maturity model,
  opinionation dial, Backstage portal, GitOps-as-interface, repo strategies. Read when
  designing developer-facing self-service.
- `references/team-topologies-and-org.md` — Conway's Law, Team Topologies, Fit-for-Purpose,
  platform team API, innersourcing, the platform engineer role, onboarding/enablement. Read
  for org design and staffing.
- `references/finops-and-governance.md` — tagging taxonomy, cost guardrails, FinOps services,
  showback/chargeback, guardrails-vs-gates, RACI, policy-as-code, CCoE. Read for cost &
  governance questions.
- `references/measuring-success-dora.md` — DORA, SPACE, DevEx, flow, cloud-native maturity
  model, outputs→outcomes→impact, advanced KPIs, adoption. Read when defining metrics.
- `references/anti-patterns-catalog.md` — the full anti-pattern → correct-pattern catalog.
  Read when reviewing a design for mistakes or when a symptom is described.
- `references/migration-to-cloud-native.md` — 7 Rs, lighthouse app, single/multi/hybrid,
  transition architectures, skill gaps, stakeholder engagement, good habits. Read for
  legacy-to-cloud-native journeys.
