# Technical Depth for TPMs

The technical toolset is the pillar that separates a TPM from a generalist PM. This is system-design literacy, enough code fluency to reason about work, and architecture-landscape awareness — not coding ability.

## Contents
- Why a technical background
- Code literacy (how much)
- The functional spec
- System design vs architecture landscape ("the forest and the trees")
- Common architectural patterns
- Design considerations (the levers)
- Latency, availability, scalability
- Defensible choices
- Areas of concern by level
- System-design interview prep

## Why a technical background

A program manager who understands the domain eliminates trial-and-error: they notice risks faster, mitigate sooner, and estimate better. The closer the TPM's technical background to the team's, the better they unblock and set the team up. The technical toolset *amplifies* the PM/program-management pillars — it doesn't replace them. You will rarely write code, but you must read it, reason about it, and challenge it.

## Code literacy (how much)

Enough to read and reason, not to ship. Specifically:
- Follow a method/API signature; understand what a class or function does.
- Challenge and refine estimates; confirm timelines; contribute to designs.
- Knowing *one* language matters more than *which* — fundamentals (capabilities, limits, problem-solving) transfer. Caveat: paradigms don't fully transfer — object-oriented (Java, C#, Objective-C) vs functional (Scala, Haskell) behave fundamentally differently; Python spans both. Knowing one OO language doesn't grant instant fluency in a functional one (and vice versa), but it's far easier than starting from zero.

## The functional spec

A **functional specification** maps business requirements to functionality (APIs, user actions, features). It's the intermediary artifact from requirements → plan, and the document the dev team builds from. Include: the requirement ID each item traces to, a team-specific task name, the **systems impacted** (services/packages/classes), and an estimate. A requirement may map to multiple teams (frontend + backend). Best practice: draft it yourself, then collaborate with the dev teams to refine (people are more productive critiquing a draft than starting blank — and it minimizes the time you take from them). Add a change log and project summary for context. Writing a credible functional spec *requires* understanding the system it references.

## System design vs architecture landscape ("the forest and the trees")

- **System design** (the trees) — components and data flow within a system or service; limited scope; may dive into API definitions and the data model. The most-exercised technical skill, because TPMs are tied to a service or group of services. To truly understand an API, ask *why* it exists (often a business reason), not just what it vends — and why a new version was created (it reveals the service's limits and evolution).
- **Architecture landscape** (the forest) — a map of application/service interconnectivity at org or enterprise scale; wider scope, shallower depth; APIs and field-level flows matter less than inputs/outputs and dependencies. When mapping, choose a lens (data flow vs system dependencies) — data flow helps you route a needed piece of data; dependencies help you gauge the blast radius of an API change.

A TPM needs **both breadth and depth** — "see the forest *and* the trees." Roadmaps (program/product/team) and published technology directives are the tools for seeing the forest: they reveal each group of systems' intent and where the landscape is heading. A senior/principal TPM is expected to spot cross-project issues and opportunities from these cues *without prompting*.

Bridging matters here too: the storage-team cautionary tale — a junior team rebuilt a storage system around their *own* pain points, never checking the upstream team that *produced* the data or the downstream teams that *consumed* it, and shipped something that missed half the real requirements (and three years later was up for scrapping). Always check upstream and downstream before (re)designing.

## Common architectural patterns

Know these well enough to recognize them in a design and evaluate trade-offs (designs often combine several):
- **MVP (Model-View-Presenter)** / **MVC** — separation of concerns between data (model), UI (view), and the connecting/event layer (presenter/controller). MVP is common in desktop apps; MVC in client-server. Lets display and data change independently.
- **Object-Oriented Architecture (OOA)** — model real-world concepts as objects; heavy in business/form-based apps. Fast to build (esp. in tools that couple model and view, like classic Visual Basic) but coupling makes large changes expensive.
- **Domain-Driven Design (DDD)** — the design is driven *directly* by the domain (healthcare, finance); often drawn as a hexagon with core domain logic in the center and **adapters** translating each client in/out. Reuses the core across many clients, but pushes high integration/maintenance cost to each client's adapter.
- **Event-Driven Architecture (EDA)** — state changes via user/system events; pairs naturally with OOA and DDD.
- **P2P** — hosts talk directly, no central server. **Structured** (strict topology, fast lookup, heavy discovery traffic) vs **unstructured** (resilient to churn, but finding a specific host/file isn't guaranteed). Modern uses: Bitcoin, some CDNs, payment apps.
- **Service-Oriented Architecture (SOA)** — the dominant pattern today and the basis of cloud architectures; logical separation of concerns into services with enforced data contracts; scale only the pieces that need it. Distributed systems often add synced/centralized data stores to hold stateful data that a desktop app would have kept in memory.
- **Client-server** — classic centralized service (web browsing, mobile-app backends). Designs often *hybridize* patterns (e.g., a web front-end fronting a P2P backend via a web-server fleet).

## Design considerations (the levers)

Make designs clear and defensible:
- **Avoid vague traits** — make pointers directional where direction matters (a read-only datastore shows one-way flow; a network topology can be bidirectional).
- **Don't omit key functionality** — usually dropped because it "seemed obvious" or the design was over-generalized.
- **Understand traffic patterns** — average, peak, and cyclical load (an e-commerce system peaks with its users' waking hours, varying by region). Drives whether auto-scaling fits.

At large-company scale, three concerns appear in nearly every design (often handled by off-the-shelf cloud components but worth understanding):

### Latency
Time to process a request/response. Long (or perceived-long) responses drive users away (lost sales/ads/market share). Combat it: co-locate microservices to cut network hops; pick the right datastore (read-optimized for read-heavy systems); hide latency with async processing (akin to swarming a project). Patterns affect latency differently — evaluate current and projected latency needs when choosing one; introducing latency isn't disqualifying, but it must be countered. Flagging long-term latency implications is a key TPM design contribution.

### Availability
Ability to serve high request volumes without an outage. Mitigate with more hosts behind a **load balancer** (which detects bad hosts and routes around them). Increasing hosts forces other adjustments (concurrent DB connections, possibly DB choice; sometimes a load-balancer fleet).

### Scalability
Ability to respond to fluctuating volume; discussed alongside availability since solutions overlap. Behind a load balancer, hosts can be added/removed on demand without breaking connections — saving cost in off-peak periods.

## Defensible choices

There's always more than one way to design a system. The goal — in reviews and interviews — is to understand the **trade-offs** you made and *why*. So long as you can defend the choices and adjust as new data arrives, the outcome is usually favorable; the point is demonstrating critical thinking, not memorizing one "right" answer. As a working TPM, you review the team's designs against these checks *and* against other in-flight projects/programs to ensure cross-system alignment.

## Areas of concern by level

Depth and breadth grow with level (mirrors how an SDE's purview grows):
- **Entry TPM** — the complete architecture of their SDM's services or their focus area (often an *embedded* TPM).
- **Industry TPM** — adds the dependent systems and initiatives around their focus area.
- **Senior TPM** — the architecture of their whole org (high-level data flow and interactions, not full internal detail of every system).

Highly specialized TPMs may go deeper on their services or on aspects like security. Where no TPM is present, an SDM's purview may expand to fill the gap.

## System-design interview prep

System design appears in nearly every big-tech TPM loop (and increasingly elsewhere). Two question types:
- **Build a system/feature from scratch** — they want your product instincts and intentional design choices; ask clarifying questions about the parameters (driving toward clarity *is* a TPM trait).
- **Describe a system you know well** — they gauge the depth of your real-world knowledge and your ability to explain a technical solution to someone with no context (the bridge skill). Know *why* design choices were made (DB type, one service vs many).

General prep: assume the role leans heavily technical (some interviewers ask about scaling by transactions-per-second); know why each choice was made; understand the system's **use cases** (user base, traffic patterns, availability, data); **research the company's** interview practices (Glassdoor, Indeed, "system design prep for [company]"); and **practice** writing out both question types so nerves don't derail you. These skills aren't just for the interview — they're the daily job.
