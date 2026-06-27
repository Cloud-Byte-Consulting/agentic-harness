# Platform Engineering Fundamentals

What platform engineering is, how it differs from DevOps and SRE, the
platform-as-a-product mindset, the four business pillars, platform/product
boundaries, and how to make the business case for building a platform team.

## Contents
- Platform engineering vs DevOps vs SRE
- Platform-as-a-product
- The four pillars (the value model)
- Layers of SaaS delivery and where the platform sits
- Platform / product boundaries
- The tipping point: when to build a platform team
- Choosing the operating model
- Making the business case

## Platform engineering vs DevOps vs SRE

All three aim at fast, reliable, scalable software delivery, but they occupy
different scopes:

| | Scope | Owns | Customer |
|---|---|---|---|
| **DevOps** | Whole SDLC, culture | Shared ownership, automation, CI/CD, "you build it, you run it" | The org/culture |
| **SRE** | Reliability of *specific* production services | SLOs/error budgets, monitoring, incident response, postmortems | A given product |
| **Platform engineering** | Building/operating the shared internal platform | Reusable IDP + tooling that many teams self-serve | *Other engineering teams* |

Useful framings:
- **DevOps is the goal; platform engineering is the means.** A good platform is *how* you
  let every team practice DevOps without each one rebuilding pipelines, observability, and
  guardrails. Analogy: DevOps : platform engineering :: Agile : SAFe (a way to scale the
  practice across the enterprise).
- **SRE is a customer of the platform.** Classic example: the platform team builds a shared
  distributed-tracing / observability capability; the SRE team consumes it to keep
  production reliable and feeds back improvements. The platform team's job is to meet the
  SRE team's needs as a user.
- **Difference from DevOps engineer:** a platform engineer builds a *product of their own*
  (the IDP) used by *all* developers, leaning more toward software engineering than to
  per-product automation. An embedded DevOps engineer per product tends to produce
  bespoke, duplicated, fragmented automation (Conway's Law) — exactly what a platform team
  factors out.

A common hybrid that works well: a **shared IDP built by a platform team + SREs embedded
in (or consulting to) stream-aligned teams.**

## Platform-as-a-product

This is the cultural keystone. Treat the IDP like any other product:

- It has **users** (developers and other engineering teams), not "tickets."
- It has a **backlog, roadmap, and release process**, driven by user needs and feedback.
- Its success is measured by **adoption and developer experience**, not by how much
  infrastructure it owns.
- It must be **discoverable, documented, and self-service** — onboarding friction kills it.
- The platform team does **outreach/evangelism, onboarding, support, and enablement**, not
  just engineering. Building the platform is the start; nurturing adoption is the real work.

If you skip the product mindset, the platform team regresses to reactive, ad-hoc DevOps
ticket-ops and the duplication you were trying to eliminate returns.

## The four pillars (the value model)

Platforms create business value across four pillars — use these to justify investment and
to design the team:

1. **Developer experience (DevX).** Intuitive, self-service, low cognitive load; developers
   focus on code, trust the platform, feel ownership. Leading indicators: Time-To-First-
   Hello-World, IDE integration, automation coverage. Lagging: onboarding completion rate,
   Developer Satisfaction Score, code quality/maintainability.
2. **Productivity & efficiency.** Standardize + automate to cut toil and operational
   overhead; faster time-to-market, better resource utilization. Measured with SPACE, DORA,
   OKRs, resource/cost efficiency, innovation rate.
3. **Security & compliance.** Embed security-by-design and policy-as-code so every app on
   the platform inherits guardrails; compliance is automated, not retrofitted. Security is
   a *foundation*, not a bolt-on.
4. **Best practices & standards.** Encode standards as policy-as-code, templates, and
   blueprints so consistency happens *without toil* and without friction-heavy enforcement.

Align all four to **business objectives** (identify goals/KPIs, collaborate with
stakeholders, communicate value to leadership). A platform disconnected from business goals
loses funding.

## Layers of SaaS delivery and where the platform sits

SaaS delivery stacks up roughly as:

- **Layer 0 — Cloud infrastructure**: compute, storage, networking; IaC (Terraform/
  CloudFormation), config mgmt. Owned by the **cloud-infrastructure team**.
- **Layer 1 — Cloud tools, runtime, CI/CD, observability**: the developer-facing platform
  (this is the IDP). Owned by the **platform / internal-developer-platform team**.
- **Layer 2 — SaaS engineering / onboarding, billing, metering, auth**: shareable
  business-adjacent logic. Owned by a **SaaS-engineering team**.
- **Product/business logic** on top, owned by **stream-aligned product teams**.

The "platform" can span any layers that are *replicable across teams*. Most platform
efforts start at Layer 1.

## Platform / product boundaries

The platform/product relationship is symbiotic (engine vs car body): the platform supplies
reusable capability; products drive new platform requirements.

Keep boundaries clean:
- The platform exposes a **standardized set of services and APIs**; it is **not tightly
  coupled** to any single product.
- Products **minimize dependencies** on the platform and should be able to function (and be
  built/tested) independently if a platform service is unavailable.
- **Separation of concerns:** platform team owns the tools/services; product teams own the
  apps. This lets the platform evolve without disrupting products.
- **API-driven integration with backward compatibility:** product teams integrate via APIs
  without knowing internals; the platform can ship new features without forcing product
  rewrites.

## The tipping point: when to build a platform team

Build a dedicated platform team when:
- DevOps work is **duplicated** across product teams (each solving the same problems
  slightly differently → fragmentation, Conway's Law).
- **Cognitive load** on stream-aligned teams is too high to also own all of infra/CI/obs.
- There is a **consistent, recurring set of needs** worth factoring into a shared product.
- The org is large/complex enough that linear scaling of infra+process is inefficient.

Use **Value Stream Mapping** to find the inefficiencies and decide what existing DevOps
tooling to fold into the platform. Below the tipping point, prefer shared tooling or
embedded SRE — a premature platform team is overhead.

## Choosing the operating model

Factors: organizational structure (centralized IT favors a platform team; highly autonomous
teams favor embedded SRE), size/complexity, budget, and goals (agility → platform;
reliability → embedded SRE). The common sweet spot is **shared IDP + embedded SREs**.

## Making the business case

To get leadership buy-in, frame in their language:
- **Long-term cost savings** (tool consolidation, reduced overhead, faster time-to-market).
- **Strategic advantage / future-readiness** (rapid adaptation, consistency, predictability
  as risk mitigation).
- The platform is an **investment that multiplies the productivity of the whole org**, not
  a cost center.

To get *team* buy-in: show tangible wins — better integrated tooling, less repetitive toil,
exposure to best practices, more collaboration. Expect resistance from teams attached to
legacy tooling; counter with workshops/pilots, training/incentives, and open feedback
channels. Convincing both leadership and teams is essential; either alone fails.
