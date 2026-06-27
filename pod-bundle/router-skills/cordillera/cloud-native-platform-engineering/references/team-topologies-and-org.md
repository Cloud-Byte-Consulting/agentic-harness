# Team Topologies & Platform Org Design

Conway's Law, the four team types and three interaction modes from Team
Topologies, the Fit-for-Purpose model, defining the platform team API,
innersourcing, onboarding/enablement, and the platform engineer role.

## Contents
- Conway's Law (use it deliberately)
- Domain-Driven Design + Conway's Law
- The four team types and three interaction modes
- Where a platform engineering team fits
- Fit-for-Purpose (WHY/WHAT/HOW/WHO)
- Platform team placement
- The platform team API
- Innersourcing (other teams contributing)
- Onboarding, support, docs, outreach, enablement
- The platform engineer role
- Squad trifecta, T-shaped engineers, champions

## Conway's Law (use it deliberately)

> "Any organization that designs a system will produce a design whose structure is a copy
> of the organization's communication structure." — Melvin Conway, 1968.

Implication: team structure determines system structure. Rather than fight this, **reverse
it** — design the architecture you want, then derive the team structure that will naturally
produce it (the "inverse Conway maneuver").

Two contrasting setups illustrate the point:
- **Embedded DevOps engineer per product** → many bespoke, duplicated, fragmented
  automation solutions (each team solves the same problems differently).
- **Dedicated platform team** → a coherent, unified platform product reflecting the
  platform team's structure and communication. Risk: the platform team can become a
  bottleneck if it requires manual/ad-hoc intervention. Mitigate with self-service +
  innersourcing + treating the platform as a product.

## Domain-Driven Design + Conway's Law

DDD says align software boundaries with **business domains / bounded contexts**. Combined
with Conway's Law: structure **small, cross-functional, autonomous stream-aligned teams**
each owning a bounded context end-to-end (design, build, test, deploy, operate). This yields
loose coupling + high cohesion and "you build it, you run it." Note a bounded context may
map to one *or several* microservices — don't equate "bounded context" with "one
microservice."

## The four team types and three interaction modes

From *Team Topologies* (Skelton & Pais):

1. **Stream-aligned team** — aligned to a flow of work from a business domain segment; the
   primary team type; cross-functional; delivers value to end users. Org by *domain*, not by
   architecture layer.
2. **Platform team** — provides a **self-service API/product** to reduce stream-aligned
   teams' cognitive load. *The platform engineering team is this type.*
3. **Enabling team** — helps other teams build capability and overcome obstacles; a force
   multiplier; mentors rather than does the work for them.
4. **Complicated-subsystem team** — owns a part requiring deep specialist knowledge (e.g.
   GPU/AI-training infra, HPC, a mainframe), abstracting it behind an interface. Rare.

Three interaction modes:
- **Collaboration** — two teams work closely for a time (e.g. platform + stream team
  co-developing tooling).
- **X-as-a-Service** — one team consumes another's capability as a service (the platform's
  default steady state).
- **Facilitating** — one team helps another reach a goal and become self-sufficient
  (enabling-team mode).

Anti-patterns: assuming too many things are "complicated subsystems" (a complex *business*
domain is not a complicated subsystem — that drags you back to architecture-aligned teams);
enabling teams *doing the work* instead of mentoring; using X-as-a-Service for things that
are everyone's responsibility (notably **security** — you cannot outsource security to a
service; you can only provide enabling tooling and defaults).

## Where a platform engineering team fits

Primarily a **platform team** (self-service API), but it flexibly also acts as an
**enabling team** (training/evangelism/help using new features) and sometimes a
**complicated-subsystem team** (specialized infra). Effective platform teams blend all
three, shifting focus as needed.

## Fit-for-Purpose (WHY/WHAT/HOW/WHO)

A team is *fit-for-purpose* when its composition, skills, and objectives align with org
needs. For a platform team (Lagerstedt's model):

- **WHY** — enable the wider org to deliver high-quality products more efficiently and
  securely by abstracting infra complexity and reducing repetitive tasks.
- **WHAT** — build and run an IDP that abstracts infra, automates deployment, enforces
  security/compliance, and provides self-service provisioning.
- **HOW** — cloud-native + DevOps best practices, automation at the core; collaborate with
  other teams, gather feedback, continuously improve.
- **WHO** — primary users are internal development teams; indirectly the wider org,
  stakeholders, and end users.

Design the team around the four pillars: include UX/dev-advocacy skills for *developer
experience*, automation/CI-CD/perf for *productivity*, cybersecurity/compliance for
*security*, and standards-stewardship for *best practices*. Assess fit continuously via
surveys, performance metrics/retros, and skills audits + training.

## Platform team placement

Place the platform team **centrally, between the infrastructure team and the product/
service teams** — it must collaborate closely with both (understand infra constraints +
understand product needs). Position it near decision-making power and resource allocation;
treat it as an **investment**, not a cost center. A centrally placed team gets the holistic
visibility needed to balance diverse, sometimes conflicting, team needs. Its role is to
**enable**, not enforce.

## The platform team API

Formalize *how* the platform team interacts with everyone else:

- **Channels of communication** — Slack/JIRA/meetings; open, support async work.
- **Request & response protocol** — ticketing/service-catalog/self-service portal; set
  expectations on response time and prioritization.
- **Contribution guidelines** — for innersourcing: coding standards, testing, docs, review
  process.
- **Change management** — how updates roll out, how teams are informed, how disruption is
  minimized, how to request changes.
- **Incident management** — how to report, who to contact, response-time expectations.
- **Feedback mechanism** — regular check-ins, satisfaction surveys, suggestion channels.

Review and refine the API over time. Define the team type (cross-functional) and **areas of
ownership**: platform infrastructure, tooling & automation, security & compliance, standards
& best practices, developer experience — and what falls *outside* its purview.

## Innersourcing (other teams contributing)

Run platform tooling like an internal open-source project: the platform team is the
maintainer; any team can contribute. Benefits from diverse perspectives. Requires: clear
contribution guidelines, open communication, mentorship/support, recognition/incentives,
strong docs, and robust CI/CD so contributions integrate safely. Mitigates the
bottleneck risk of a central platform team.

## Onboarding, support, docs, outreach, enablement

Building the platform is the beginning; adoption is the real work.
- **Onboarding** — intro to functionality/purpose/benefits, access instructions, best
  practices, hands-on exercises.
- **Support** — clear help channels, timely responses, proactive monitoring, review of
  support cases for recurring issues.
- **Documentation** — user guides, API references, troubleshooting guides, kept current.
- **Outreach** — demos, newsletters, testimonials/case studies, community forums.
- **Enablement** — training/workshops, self-guided learning, opportunities to contribute,
  a culture of experimentation.

## The platform engineer role

A platform engineer builds a **product** (the IDP) used by all developers — closer to a
*software engineer* than a DevOps engineer. Needs: deep cloud/distributed-systems/
microservices knowledge, programming, automation & CI/CD, IaC, containers, Kubernetes —
plus business context and strong communication. Backgrounds are diverse (dev, ops,
security, networking, DBA); multi-disciplinary teams produce more holistic solutions and
collaborate better. Career path: engineer → senior → platform architect / principal, or
specialize (security, data) or move into leadership. **Soft skills** (communication,
leadership, teamwork, problem-solving) are the mortar holding the technical work together.
Platform engineers must **stay ahead of the technology curve** via community involvement,
hands-on experimentation, and continuous learning, and have a methodical way to **evaluate
and integrate new tools** (define requirements → research → evaluate → pilot → plan →
implement → verify).

## Squad trifecta, T-shaped engineers, champions

- **Squad trifecta leadership** (Spotify atomic squad): engineering manager (technical
  direction), product owner (the customer/value), scrum master (organization). Minimal
  overlap; use **commander's intent** — communicate the *goal*, leave the *how* to
  implementers so they can seize emergent opportunities. Anti-pattern: all developers
  reporting to the engineering manager.
- **T-shaped engineers** — broad cloud-native knowledge (the bar) plus deep expertise in a
  few areas (the stem); a mix gives the team diverse technical opinions.
- **Champions** — for org-wide "job-zero" initiatives (security, accessibility, quality),
  appoint per-team champions, supported by an enabling team (certs, conferences, internal
  knowledge sharing). The company must **invest in people** for these to work.
