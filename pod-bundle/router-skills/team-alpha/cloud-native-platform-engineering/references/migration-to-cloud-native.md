# Migrating from Legacy to Cloud-Native

How to plan a migration that adds value (not just moves workloads): assess and
classify, the 7 Rs treatment plan, the lighthouse app and T-shirt sizing, cloud
model choice, transition architectures, stakeholder engagement, skill gaps, and
good habits for continuous improvement.

## Contents
- Plan first
- Assess and classify the estate
- The 7 Rs (treatment plan)
- Prioritization: lighthouse app and T-shirt sizing
- Choosing the cloud model
- Choosing the cloud platform
- Phased rollout and timelines
- Transition architectures (strangler-fig)
- Stakeholder engagement
- Closing skill gaps
- Good habits / continuous improvement

## Plan first

The biggest cause of migration failure is jumping in without a strategy. Rushed migrations
produce un-optimized apps, misaligned workloads, surprise costs, unrealized ROI, and missed
compliance/security. A migration must align with **business goals** from the start — you're
adding value, not just relocating workloads.

## Assess and classify the estate

Inventory critical applications, dependencies, integrations, and network/latency needs
using migration-assessment tooling (AWS Migration Evaluator, Azure Migrate, Google Migrate
for Compute Engine). Output: a list of assets with their resource requirements (CPU/RAM/OS),
so you can decide what to retire, rework, or keep, and plan instance sizing.

## The 7 Rs (treatment plan)

A treatment plan categorizes each workload (AWS 7 Rs / Azure CAF / GCP 6 Rs share the same
principles):
- **Rehost (lift-and-shift)** — move with minimal change. Fast, simple; but leverages no
  cloud-native benefits (see the anti-pattern). Use when time-boxed (e.g. data-center exit),
  then plan to refactor.
- **Replatform** — small tweaks to optimize (e.g. managed app platform/DB) without a full
  overhaul.
- **Repurchase** — swap on-prem software for SaaS.
- **Refactor** — redesign to be truly cloud-native (serverless/FaaS, microservices,
  Kubernetes). Highest value, highest effort.
- **Retire** — decommission what's no longer needed.
- **Retain** — keep on-prem (not cloud-ready or compliance-bound), often via hybrid.
- **Relocate** — move workloads between clouds/environments for cost/performance/compliance.

## Prioritization: lighthouse app and T-shirt sizing

Don't migrate everything at once. Pick a **lighthouse**: a smaller, low-risk but visible app
that clearly benefits from cloud-native features (auto-scaling, serverless) to prove value
and build momentum. Size the rest with **T-shirt sizing**:
- **Small** — simple, low-impact; easy lift-and-shift; do early for quick wins.
- **Medium** — needs some replatforming/tweaking; specific latency/integration needs.
- **Large** — mission-critical; needs significant re-architecture; phased with detailed
  planning.

Use the cloud frameworks (AWS CAF, Azure CAF, Google CAF) as reference structures.

## Choosing the cloud model

Deliberately choose per workload:
- **Single cloud** — simplest to operate (one set of tools/APIs); fine for less-complex/
  internal apps; HA still achievable via multiple AZs/regions. Trade-off: vendor lock-in and
  provider-outage exposure.
- **Multi-cloud** — resilience/redundancy for mission-critical apps and some compliance
  needs; but high complexity (multiple toolchains/skills) and **expensive cross-cloud data
  transfer** (especially continuous DR replication).
- **Hybrid** — mix on-prem + cloud (legacy, data-residency, regulatory); tools like AWS
  Outposts, Azure Arc/Stack, Google Anthos bridge the gap. Often a **transitory** phase of a
  longer migration.

Beware the multi-cloud "lowest common denominator" trap (everything on VMs/containers to stay
portable), which forfeits managed-service benefits and raises operational complexity.

## Choosing the cloud platform

Evaluate per business/technical needs: core services/features, cost models (including
egress/data-transfer in multi-cloud), compliance/security certifications and data-residency,
multi-cloud/hybrid interoperability (cloud-agnostic tools like Terraform/Kubernetes reduce
lock-in), and latency/regional presence. Back every choice with an evaluation of services,
cost, compliance, and scalability.

## Phased rollout and timelines

Phase the work with clear milestones:
1. **Assessment** (discovery/planning) — run discovery tools; technical audits of network/
   DB/security/storage; review with stakeholders.
2. **Proof of Concept / pilot** — migrate a non-critical app or two; validate tools and
   process at small scale before scaling.
3. **Full-scale migration** — configure IAM/network/security first; migrate in **batches**
   (e.g. 20% over 4–6 weeks), reassess, fine-tune.
4. **Optimization & fine-tuning** — performance review, rightsizing, auto-scaling, monitoring;
   post-migration review.
With **regular stakeholder check-ins** throughout (weekly/bi-weekly), runbooks, and
architecture diagrams.

## Transition architectures (strangler-fig)

If you can't go current-state → target-state in one release, design **transition
architectures** — intermediary states that keep things working during migration:
- **Interim-state designs** / blueprints per phase.
- **Temporary services / hybrid architectures** for continuity.
- **Integration points** between legacy and cloud-native (API gateways, data sync), e.g. the
  **strangler-fig / facade pattern** (Martin Fowler): route traffic through a facade and
  incrementally replace legacy components behind it.
Use **iterative rollouts** in controlled batches; phase out transition architectures as you
reach the cloud-native target.

## Stakeholder engagement

**Minimum stakeholder commitment** is a top failure cause. Cloud migration is a business
transformation, not an IT-only project. Engage executive leadership (sponsorship/ROI),
technical teams (architects/devs/SREs), operations & security (DevSecOps alignment), and
business stakeholders (goals/metrics, change management, HR for cultural change) from the
start. Mechanisms: steering committees, regular check-ins, defined business goals + KPIs,
budget oversight and workload prioritization, quarterly strategy reviews, and post-migration
optimization. Use collaborative workshops (OKR alignment, cross-functional planning, release
strategy, risk management, cultural integration, technical-business alignment) to resolve
conflicting priorities (e.g. dev wants speed, ops wants stability).

## Closing skill gaps

Underestimating skill gaps causes misconfigurations, missed optimizations, and security
holes. Traditional ops skills don't directly transfer to IaC, serverless, container
orchestration, elasticity, and cloud security models. Close gaps via: targeted cloud
training and **certifications**, cross-functional collaboration (cross-training, collaborative
projects, mentorship), **cloud fluency across the whole org** (stakeholder workshops,
cloud-specific KPIs), and clear ownership (define roles, appoint **cloud champions**).

## Good habits / continuous improvement

Migration doesn't end at cutover. Embed continuous improvement:
- **Stakeholder alignment** — clear communication, RACI, the right roles (cloud architect,
  platform engineer, developer/SRE, security engineer, product owner), product-centric (not
  siloed) teams.
- **Roadmap** — migration plan + transition architectures + delivery initiatives aligned to
  migration phases; modular planning; automation/orchestration; measure success
  (deployment frequency, time-to-recovery, application performance) and adjust.
- **Continuous improvement** — evolve building blocks (enhanced CI/CD with Argo CD/Flux,
  observability with Prometheus/Grafana/OTel, automated policy enforcement with Kyverno/OPA
  Gatekeeper), adaptive governance (dynamic guardrails, decentralized decisions,
  feedback-driven adjustments), collaborative culture (cross-functional squads,
  retrospectives, knowledge sharing), team autonomy + accountability (decentralized
  decisions, clear ownership, DORA success metrics), dependency management, third-party
  integration management (SLAs, monitoring, fallbacks), and feedback loops (real-time
  monitoring, centralized logging — e.g. the LGTM stack, distributed tracing). Leverage
  **DORA metrics** to set baselines, find optimization areas, and celebrate wins.
