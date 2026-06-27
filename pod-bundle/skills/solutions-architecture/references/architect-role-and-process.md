# The Solutions Architect Role & Process

How the role works, how to gather and translate requirements, and how to run the design process. Read this when scoping a role, defining a delivery process, or running a requirements exercise.

## Contents
- The role in one sentence
- Generalist vs specialist architect variants
- The solution delivery lifecycle
- Core responsibilities
- Functional vs non-functional requirements (with the NFR question set)
- Architecture constraints (with the constraint question set)
- Technology selection, POCs, and prototypes
- MVP and MoSCoW
- The architect in an Agile organization
- IT procurement (RFx) — see also documentation-and-communication.md

## The role in one sentence

A solutions architect converts an organization's business vision into a technical solution and acts as the liaison between business and technical stakeholders, owning the design through delivery and operation. The value they add is *translation* (business ↔ technical) and *trade-off analysis* — not just choosing technology.

Strategic vs tactical: strategically the architect builds a long-term, extensible vision; tactically they design something that handles today's workload. Both matter. Solution architecture spans more than software — it covers infrastructure, networking, security, compliance, operations, cost, and reliability.

## Generalist vs specialist architect variants

Titles overlap by organization; responsibilities blur. Know the landscape so you can pitch at the right altitude.

**Generalists** (broad across domains):
- **Enterprise architect** — organization-wide IT strategy and standards; aligns IT with business goals across departments. Broader and more strategic than a (project-centric) solutions architect.
- **Application / software architect** — software design for a project: API design, scalability, integration, best practices, mentoring developers.
- **Cloud architect** — cloud strategy, migration, hybrid design, cloud-native architectures.
- **Architect evangelist / technology evangelist** — drives platform adoption via content, talks, and trusted-advisor relationships.

**Specialists (SSAs)** (deep in one area): infrastructure, network, data/analytics/big-data, ML, GenAI, security, DevOps, and industry architects (finance, healthcare, retail, manufacturing), plus product/platform specialists (Salesforce, SAP, Snowflake, etc.). A generalist typically coordinates several specialists on a complex project.

## The solution delivery lifecycle

The architect is involved across all phases; it is iterative — production feedback feeds the next iteration.

1. **Business requirement & vision** — work with business stakeholders to understand the vision.
2. **Requirement analysis & technical vision** — analyze requirements; define a technical vision to execute the strategy.
3. **Prototyping & recommendation** — build POCs, showcase prototypes, make technology selections.
4. **Solution design** — design to org standards; this is the architect's primary ownership.
5. **Development** — bridge business and dev; unblock the team.
6. **Integration & testing** — verify FRs and NFRs are met.
7. **Implementation** — guide smooth deployment.
8. **Operation & maintenance** — logging/monitoring, scaling, DR.

During **solution design** the architect specifically: documents solution standards; defines the high-level design, cross-system integration, solution phases, implementation approach, and monitoring/alert approach; documents the pros/cons of design choices and audit/compliance requirements.

## Core responsibilities

- Analyze functional requirements and translate them into a design.
- Define and champion NFRs (the most-dropped part of any design).
- Understand and engage stakeholders (technical and non-technical, internal and external).
- Understand and balance architecture constraints; manage scope creep.
- Make technology selections against explicit criteria.
- Build POCs and prototypes to de-risk choices.
- Own solution design and delivery; mentor the dev team.
- Ensure post-launch operability, scaling, and DR.
- Act as a technology evangelist where applicable.
- Help PMs with resourcing, cost estimation, timeline, and release/support planning.

## Functional vs non-functional requirements

**FRs** specify *what* the system does — behaviors, functions, features tied to user interactions ("place an order", "reset a password"). **NFRs** specify *how well* — the quality attributes and operating conditions. NFRs are invisible to users until they are missing, and they slip when teams over-focus on features. The architect's signature move is asking the NFR questions nobody asked, and **quantifying** each one.

### The NFR question set (ask these, attach numbers)
- **Performance** — App load time for users? How to handle network latency? (e.g. catalog page under 3 s.)
- **Security & compliance** — How to secure from unauthorized access and malicious attacks? Which local laws and audit requirements apply?
- **Recoverability** — How to recover from an outage? Minimize recovery time? Recover lost data? (Pin RTO/RPO.)
- **Maintainability** — Monitoring and alerting? Support model?
- **Reliability** — How to perform consistently? Detect and correct glitches?
- **Availability** — How to ensure high availability and fault tolerance? (e.g. 99.95% ≈ 22 min/month downtime.)
- **Scalability** — How to meet rising demand and sudden spikes?
- **Usability & accessibility** — Simple to use? Seamless UX? Accessible to a diverse user base (localization, screen readers, low bandwidth)?
- **Portability / interoperability** — Run across platforms? Exchange data via JSON/XML with upstream/downstream systems?

Some NFRs are project-specific (e.g. voice clarity for a call-center solution).

## Architecture constraints

Constraints bound every design and trade off against each other. Name them, balance them, and analyze each trade-off (cutting resources to save cost can blow the timeline; squeezing schedule with fewer people lowers quality and raises rework cost).

### The constraint question set
- **Cost** — Funding available? Expected ROI?
- **Quality** — How closely must outcomes match FRs/NFRs? How is quality tracked?
- **Time** — Delivery deadline? Flexibility?
- **Scope** — Exact expectations? How are requirement gaps handled?
- **Technology** — What can be used? Legacy vs new? Build vs buy?
- **Risk** — What can go wrong; how to mitigate? Stakeholders' risk tolerance?
- **Resource** — What's needed; who will build it?
- **Compliance** — Local legal requirements? Audit/certification needs?

**Scope creep** — the gradual expansion of deliverables without matching increases in resources, time, or budget — is the most dangerous because it degrades every other constraint. Manage it actively.

## Technology selection, POCs, and prototypes

Define **selection criteria** before choosing: integration with other frameworks/APIs, performance, security, team skills, scalability headroom. Pick technology that satisfies current requirements *and* leaves room for future scale.

- **POC** — evaluate a technology against a subset of critical functionality; short-lived; reviewed by experts; throwaway. Use it to de-risk an unproven choice.
- **Prototype** — built for demonstration to a customer (e.g. to secure funding); also not production-ready.

## MVP and MoSCoW

Put the customer first and work backward. Deliver a **minimum viable product** — the smallest set of core features that delivers value and validates the idea — then iterate on feedback. This conserves resources under the cost/time/scope/resource constraints.

**MoSCoW** prioritization:
- **Must have** — critical; cannot launch without.
- **Should have** — highly desirable once in use.
- **Could have** — nice to have; absence doesn't break desired functionality.
- **Won't have** — users won't notice its absence (this iteration).

Plan the MVP from must-haves; defer the rest to later iterations.

## The architect in an Agile organization

Agile is iterative and feedback-driven, not the linear waterfall. The architect provides a solid architectural foundation while staying adaptable: inspect and adapt, re-architect iteratively, build prototypes to minimize risk, design loosely coupled and extendable interfaces, and rely on automation, rapid deployment, and monitoring (microservices + CI/CD). The myth that "Agile means no architecture" is false — the architect reduces the *cost of change* and creates a framework to reverse incorrect decisions quickly.

## Career and frameworks

Useful to know by name: **TOGAF** and the **Zachman Framework** for enterprise architecture; cloud certifications (AWS Certified Solutions Architect, Azure Solutions Architect Expert, Google Professional Cloud Architect). Cloud proficiency is now table stakes.

## IT procurement (RFx)

Architects often respond to or lead procurement. **RFI** (request for information) gathers vendor capabilities early; **RFP** (request for proposal) invites open-ended solution proposals from shortlisted vendors; **RFQ** (request for quotation) requests pricing against tightly specified requirements. Detail and structure of an RFP response are covered in `documentation-and-communication.md`.
