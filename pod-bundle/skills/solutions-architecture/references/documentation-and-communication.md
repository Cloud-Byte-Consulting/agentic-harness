# Architecture Documentation & Communication

The Solution Architecture Document (SAD), architecture decision records (ADRs), diagramming, IT procurement docs, and communicating to mixed audiences. Read this when producing or reviewing architecture documentation.

## Contents
- Why documentation matters
- The Solution Architecture Document (SAD)
- SAD views (the seven views)
- SAD structure (sections)
- SAD lifecycle
- SAD best practices and pitfalls
- Architecture decision records (ADRs)
- IT procurement: RFP / RFI / RFQ
- Communicating to mixed audiences

## Why documentation matters

The architect must communicate the design to all technical and non-technical stakeholders to get a shared understanding and formal agreement. Good documentation retains knowledge through attrition, makes the design people-independent, traces the solution back to business requirements, and surfaces constraints, assumptions, and risks the implementation team would otherwise miss.

## The Solution Architecture Document (SAD)

The SAD gives an end-to-end view of the application so everyone is on the same page. It serves project managers (coordination), business analysts (alignment to requirements), technical teams (implementation/maintenance), senior management (strategic decisions), and clients/end users (outcome meets needs). Goals: communicate the end-to-end solution; give a high-level overview and multiple views addressing service-quality requirements (reliability, security, performance, scalability); trace the solution to business requirements (FRs and NFRs); provide views for design/build/test/implement; define impacts for estimation/planning/delivery; define business continuity and operations. For modernization, it shows current + target architecture and a transition plan, and records assessments/POCs.

## SAD views (the seven views)

Write for both business and technical readers; put yourself in each stakeholder's shoes. Use standard diagrams (UML or block diagrams). Include, where applicable:
- **Business view** — value proposition; high-level scenarios (use-case diagram); stakeholders and resources.
- **Logical view** — system packages/modules and how users interact with them (each package may be a microservice).
- **Process view** — how critical processes work together (state/sequence diagram).
- **Deployment view** — how it runs in production (network, firewall, load balancer, app servers, DB); a block diagram for business users, a UML deployment diagram for technical users; the physical layout.
- **Implementation view** — the core: architecture and technology choices (3-tier/n-tier/event-driven) with reasoning, pros/cons (e.g. Java vs Node.js), and the resources/skills to execute.
- **Data view** — data flow and storage; data security/integrity; entity-relationship diagram; reports/analytics.
- **Operational view** — post-launch maintenance: SLAs, monitoring/alerting, DR plan, support plan, patching, backup/recovery, security-incident handling.
Add physical, network, or security-controls views as stakeholders require.

## SAD structure (sections)

Adapt sections per project (new build, modernization, or cloud move):
- **Solution overview** — a couple of paragraphs + a high-level block diagram; then subsections: solution **purpose**, **scope** (with explicit out-of-scope), **assumptions**, **constraints** (technical/business/resource, plus compliance and risk/mitigation), **dependencies** (upstream/downstream), and **key architecture decisions** (problem, options with pros/cons, decision, rationale).
- **Business context** — business capabilities, key business requirements (link to a detailed requirements doc), key business processes (process diagram), business stakeholders, and **NFRs** with numbers (scalability, availability/reliability, performance, portability, capacity).
- **Conceptual solution overview** — an abstract big-picture diagram bridging business and technical readers.
- **Solution architecture** — deep dive with subsections: **information architecture** (navigation flow), **application architecture** (technology building blocks), **data architecture** (ERD), **integration architecture** (upstream/downstream systems and data flows), **infrastructure architecture** (deployment diagram, server/network config), **security architecture** (IAM, infra security, app security/WAF/DDoS, data security, threat model).
- **Solution implementation** — development (tools, language, repo, versioning), deployment (approach, tools, checklist), data migration, application decommissioning/exit strategy.
- **Solution management** — production support and ongoing maintenance: operational management/patching, upgrade/release tools, infra management, monitoring/alerts/dashboards, support/SLA/incident management, DR and business-process continuation.
- **Appendix** — supporting research: POC outcomes, tool comparisons, vendor/partner data, open issues.

The SAD does *not* include application-level detail (class diagrams, pseudocode) — that belongs in a software design document owned by the software architect/senior developer.

## SAD lifecycle

A running document aligned to the project lifecycle: **Initiation** (recognize the need, set objectives/scope) → **Gathering requirements** (from stakeholders) → **Drafting** (initial blueprint) → **Review and feedback** (stakeholders) → **Finalization** (incorporate feedback) → **Implementation** (guides delivery) → **Maintenance** (update for tech/business/regulatory changes). Iterate continuously.

## SAD best practices and pitfalls

**Best practices**: keep it clear and concise (avoid jargon); involve stakeholders regularly; keep it up to date; align the architecture with broader business goals. **Pitfalls**: overcomplicating the architecture; insufficient flexibility to adapt to scope changes; insufficient stakeholder engagement (misalignment with real needs); poor documentation practices (misunderstandings, implementation errors).

## Architecture decision records (ADRs)

For a single significant decision, capture it as an ADR: the context/problem, the options considered, the decision, the rationale, and the consequences/trade-offs (what you gave up). ADRs make decisions reviewable, defensible, and revisitable — this rationale *is* the architect's value. They also feed the "key architecture decisions" section of the SAD.

## IT procurement: RFP / RFI / RFQ

Architects respond to or lead procurement (outsourcing, contracting, buying software/SaaS). The RFx documents:
- **RFI (request for information)** — early; collect vendor capabilities to compare and shortlist (e.g. surveying LMS platforms).
- **RFP (request for proposal)** — open-ended; shortlisted vendors propose solutions (approaches, timelines, costs), often with options and pros/cons (e.g. a government IT upgrade). Most common; buyers often jump straight to RFP, so structure it for easy comparison of capabilities/approach/cost.
- **RFQ (request for quotation)** — tightly specified requirements; vendors quote price (e.g. specific cloud compute/memory/storage/bandwidth with discounts).

The architect's role in an RFP: understand requirements; design a preliminary solution (technologies, infra layout, integration); collaborate cross-functionally; estimate resources; assess risks and mitigations; produce technical documentation (system/data-flow diagrams); shape pricing strategy with finance; and present/defend the technical strategy. Include a migration strategy, TCO estimate, and compliance/data-management considerations.

## Communicating to mixed audiences

Write so a non-technical stakeholder follows the *why* and a developer follows the *what*. Translate technical concepts into business language and vice versa; use the right diagram for the audience (block diagram for business, UML for technical); lead with the value proposition and decisions, then the depth. The architect is the liaison that fills the communication gap whose absence is a common cause of project failure.
