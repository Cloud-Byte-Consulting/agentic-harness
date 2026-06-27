# Cloud-Native Anti-Patterns Catalog

Each entry: the anti-pattern, why it bites, and the correct pattern. Use this when
reviewing a design for mistakes or when a user describes a symptom.

## Contents
- Architecture & adoption anti-patterns
- Delivery & release anti-patterns
- Security anti-patterns
- Cost / FinOps anti-patterns
- Governance anti-patterns
- Culture & organizational anti-patterns
- Migration anti-patterns

## Architecture & adoption anti-patterns

**Lift-and-shift = cloud-native.** Moving a VM/DB to the cloud unchanged keeps all the
operational toil (patching, vuln scanning, hand-built security/HA) and can cost *more* than
on-prem; it leverages no cloud-native benefits. → **Correct:** refactor toward managed/
cloud-native services (managed DB, FaaS, object storage) to cut operational complexity and
inherit scalability/resilience/observability. A data-center-exit deadline may justify a
temporary rehost, but plan the subsequent refactor.

**"Containers solve everything" / "cloud-native requires containers/Kubernetes."**
Containerizing a monolith doesn't make it cloud-native; not every workload needs containers,
and Kubernetes has a steep learning curve that's often underestimated. → **Correct:** choose
the right abstraction — FaaS or managed services where they fit; use Kubernetes for genuine
microservice/orchestration needs and only if the skills exist in your market.

**Cloud-native = microservices (conflation).** Microservices is an *architecture* guidance;
cloud-native is a broader approach (scalability, resilience, CD, leveraging cloud infra,
DevSecOps culture, CI/CD). → **Correct:** treat them as distinct; microservices in the cloud
*can* be part of a cloud-native strategy but aren't required.

**Distributed monolith.** Microservices that must be deployed together or share a database —
all the operational complexity of microservices with none of the independence. → **Correct:**
use DDD bounded contexts and Conway's Law deliberately; ensure services deploy and roll back
independently and own their data.

**Over-engineering / under-engineering the platform.** Building a fully opinionated rigid
platform (resistance, one-size-fits-none) *or* a fully flexible one (fragmentation, analysis
paralysis, regression to ad-hoc DevOps). → **Correct:** semi-opinionated — sensible defaults
with escape hatches; optimize golden paths for the majority.

**"Cloud-native automatically saves money."** Only true if architected right (see
lift-and-shift, containers). → **Correct:** architect for cost; adopt FinOps.

**Premature platform team.** Building a platform team before the tipping point adds overhead.
→ **Correct:** confirm duplicated DevOps effort / crushing cognitive load first; below that,
use shared tooling or embedded SRE.

## Delivery & release anti-patterns

**Siloed release process.** Many hand-offs, no end-to-end ownership; large infrequent
bundles → poor deployment frequency, long lead time, high change-failure rate, slow MTTR,
unclear ownership. → **Correct:** empowered, cross-functional teams that own delivery
("you build it, you run it"), supported by enabling teams that provide platform tools,
automated pipelines, and an observability platform (DevSecOps).

**Velocity-as-deadlines.** Treating CI/CD efficiency gains as an excuse for tighter
deadlines undermines the gains. → **Correct:** decouple feature work from the release
schedule; release when ready for a *sustainable* velocity increase.

**Environment-branch Gitflow drift.** A branch per environment causes config drift; hotfixes
mean the first real test of a feature + hotfix together is in production. → **Correct:**
trunk-based development; promote a **single immutable artifact** across environments; use
**feature flags** (prefer enums like Baseline/Configurable/Off over booleans) to decouple
merge from release.

**Irreversible changes.** Mutable artifacts, destructive changes (dropping a table), or no
reverse migration → can't roll back, high-pressure hotfixes. → **Correct:** immutable
versioned artifacts; deprecation schedules for destructive changes; forward *and tested*
reverse migrations.

**QA/security at the end.** Invoking QA only when a feature is "done"; long-lived feature
branches; ignored/intermittent tests; giant PRs nobody reviews. → **Correct:** shift left —
ephemeral cloud environments for early QA, TDD/BDD (red-green-refactor), enforce coding
standards on small atomic PRs, trust the test suite.

## Security anti-patterns

**Security as an afterthought / bolt-on.** Retrofitting security widens the attack surface,
risks breaches, expired certs, and non-compliance. "Security is job zero." → **Correct:**
shift left — IDE checks → pipeline scans (SAST, SCA, secret scanning, image/vuln scanning) →
admission policy → runtime monitoring; security-by-design from inception (threat modeling).

**Security as a release-blocking gate / X-as-a-Service.** You cannot outsource security to a
service or a gatekeeping team; it's everyone's responsibility. → **Correct:** make security
an **enabling** function: sensible-default enabling artifacts (default WAF, secure-by-default
modules), guardrails (preventive + detective), guardrail observability, pentesting
asynchronous to the deployment path.

**Developers need more cloud permissions to shift-left.** The opposite — broad human perms
cause drift and break DR. → **Correct:** developers get constrained read/inspect perms;
escalate privilege in **CI/CD pipelines**, not to humans.

**Replicating on-prem security controls in the cloud.** Perimeter/"trusted once inside"
models, manual firewall/VPN config don't fit dynamic cloud; creates gaps + overhead. →
**Correct:** Zero Trust (verify always, least privilege, continuous monitoring,
micro-segmentation, device security), identity-based security (IAM + MFA + regular audits),
cloud-native security tooling, automated detection/response. (Implementation in
kubernetes-security-rbac.)

**Misunderstanding the shared responsibility model.** Assuming the CSP handles encryption/
access/backups (the famous "can we call AWS for backups?" after an admin deletes the only
environment). → **Correct:** know your responsibilities (data, endpoints, IAM, account mgmt
always yours); enforce least privilege; test recovery.

## Cost / FinOps anti-patterns

**No tagging standards / no enforcement.** → tag taxonomy + policy-as-code enforcement in
CI/CD; deny untagged resources; audit. (See `finops-and-governance.md`.)
**"Billing will sort itself out."** → budgets, alerts, anomaly detection, ownership.
**Rushing into an expensive third-party FinOps tool** whose license scales with spend → use
native CSP FinOps services unless multi-cloud/non-CSP spend genuinely warrants it.
**Ignoring non-obvious costs** (egress, cross-region/AZ transfer, long-term storage, idle
resources, multi-cloud DR replication) → upfront cost calcs from solution + data-flow
diagrams, well-architected reviews, lifecycle policies, guardrails.
**Penny-pinching / cost-over-value** (hand-rolling a managed service; skipping continuous
improvement) → optimize ROI; ownership + chargeback so teams self-optimize.

## Governance anti-patterns

**Over-centralized / calcified governance.** Central approval boards become bottlenecks;
slow decisions, resistance to change, disengaged teams. → **Correct:** decentralize with
guardrails; thin CCoE; streamline processes; empower frontline teams.

**"Our business is too special for guardrails/standards."** Almost always false; you scramble
at audit time. → **Correct:** standardize with guardrails; allow well-defined exceptions; GRC
+ RACI.

**Missing feedback loops.** Governance/process never adapts. → **Correct:** retrospectives,
game days, monitoring/alerts, cross-functional review boards.

## Culture & organizational anti-patterns

**"The culture doesn't need to change."** Overlaying cloud tech on unchanged practices fails;
cloud-native is primarily a cultural shift. → **Correct:** invest in cultural change
(autonomy, collaboration, continuous improvement — cf. Spotify squads/tribes/chapters/guilds),
change management (ADKAR, Kotter), clear communication.

**"Learning will happen miraculously."** Expecting upskilling in personal time → talent
churn, tech debt, dependence on consultants. → **Correct:** structured learning during work
hours (onboarding bootcamps, weekly learning time, certs, tech talks, tech-debt days,
mentorship), backed by leadership.

**Resistance to change / lack of buy-in / poor communication.** → evidence-based case
studies, leadership endorsement, training, connect technical benefits to business goals,
transparent communication.

**Minimum stakeholder commitment.** Decision-makers engage only when problems arise →
delays, misalignment, budget overruns, reactive rushed lift-and-shifts. → **Correct:** active
ongoing engagement — steering committees, regular check-ins, defined business goals/KPIs,
budget oversight, quarterly reviews.

**Measuring activity instead of outcomes.** → track outputs→outcomes→impact; combine DORA +
SPACE + DevEx; never a single metric. (See `measuring-success-dora.md`.)

**Conway's Law ignored.** Team structure silently shapes architecture. → **Correct:** design
the architecture you want, then the teams (inverse Conway maneuver); align teams to bounded
contexts.

## Migration anti-patterns

**Jumping in without a plan.** → assess/classify workloads, prioritize, choose cloud model,
phase the rollout. (See `migration-to-cloud-native.md`.)
**Underestimating skill gaps.** → cloud training/certs, cross-functional collaboration,
mentorship, cloud champions, cloud fluency across the org.
**Big-bang migration.** → start with a low-risk **lighthouse** app for a quick win; phase by
T-shirt size (small/medium/large); use transition (strangler-fig) architectures.
