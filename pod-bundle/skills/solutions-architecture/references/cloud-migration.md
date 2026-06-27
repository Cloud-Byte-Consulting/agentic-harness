# Cloud Migration & Cloud-Native Architecture

Cloud models, the migration strategies (6 Rs), the migration process, and hybrid/multi-cloud. Read this for any migration question.

## Contents
- Cloud deployment models
- Cloud service models (IaaS/PaaS/SaaS/FaaS)
- Cloud-native thinking and design
- The 6 Rs (7 Rs) migration strategies
- Choosing a strategy
- Migration risks
- The migration process (discover → optimize)
- Migration execution: data and servers
- Hybrid connectivity, multi-cloud, CloudOps

## Cloud deployment models

- **Public cloud** — provider-owned, multi-tenant, internet-accessible IT resources on pay-as-you-go (AWS, Azure, GCP, Alibaba, OCI). Like an electricity utility: pay for what you use; the provider abstracts the heavy lifting and brings cutting-edge services (analytics, ML, IoT, GenAI).
- **Private cloud** — single-organization, on-premises; replication/extension of the company's data center.
- **Hybrid cloud** — mix of on-prem and public cloud; common when legacy/licensed apps must stay on-prem or compliance requires on-prem data; also dev/test in cloud, prod on-prem (or vice versa).
- **Multi-cloud** — workloads distributed across providers to leverage each one's strengths or match team skills; weigh added complexity.

## Cloud service models

Customer responsibility decreases as you move up the stack:
- **IaaS** — provider manages infrastructure (compute, network, storage); you manage OS up.
- **PaaS** — provider also manages OS, runtime, patching; you focus on business logic and data.
- **SaaS** — provider delivers ready-to-use software; you just use it (Gmail, Salesforce).
- **FaaS** — serverless functions (e.g. AWS Lambda); pay per execution; basis of serverless architecture.

## Cloud-native thinking and design

Cloud-native means *leveraging* cloud capabilities, not merely hosting on the cloud. Optimize for on-demand scale, distributed design, and replacing (not fixing) failed components, using CI/CD and IaC for automated recovery, scaling, self-healing, and HA. Develop "cloud thinking": pay-as-you-go (run servers only when needed), elasticity (no excess capacity for peaks), global deployment for HA/DR, cloud-native monitoring (e.g. CloudWatch) and auditing (e.g. CloudTrail) instead of costly third-party licenses, and the **shared responsibility model** (provider secures infrastructure; you secure app + data).

Cloud-native moves include: containerizing a monolith into microservices with a CI/CD pipeline; building serverless (Lambda + DynamoDB + API Gateway); a serverless data lake (S3 + Glue + Athena). Benefits: fast scale-out on demand, quick replication via IaC, easy tear-up/tear-down for experiments. A holistic view across performance, scaling, HA, DR, fault tolerance, security, and automation is required.

## The 6 Rs (7 Rs) migration strategies

Per-application strategies (the book's "7 Rs" = retain, retire, relocate, rehost, repurchase, replatform, refactor):

**Lift-and-shift (fastest, least cloud-native):**
- **Rehost** — lift and shift with minimal change; fast, predictable, cheap. For temporary dev/test, packaged software (SAP, SharePoint), or apps without an active roadmap. Optimize *after* migrating.
- **Replatform** — upgrade the platform during migration without changing the core architecture (OS upgrade, 32→64-bit, DB engine change, adopt managed services). Requires re-testing.
- **Relocate** — move containerized or VMware workloads (VMotion/Docker) quickly with minimal change; relocate many apps in days.

**Cloud-native (more upfront work, higher payoff):**
- **Refactor** — rearchitect/rewrite before migrating to become cloud-native (monolith → microservices, traditional DB → cloud DB, containerize/serverless). Most time/resource intensive; for experienced teams. Alternative: migrate first, then optimize.
- **Repurchase** — drop the existing app and buy a cloud-compatible/SaaS replacement (e.g. replace on-prem CRM with Salesforce).

**Keep or remove:**
- **Retain** — leave on-prem for now (technical constraints, legacy coupling); may need hybrid connectivity to migrated apps.
- **Retire** — decommission unused/redundant/incompatible apps (DR servers, duplicates from M&A, third-party tools now native in cloud).

**Workload rationalization** — consolidate duplicate systems (e.g. multiple CRMs into one) to cut redundancy and complexity while deciding what to keep/update/retire/consolidate.

## Choosing a strategy

Driven by business goals and constraints (financial, resource, time, skills). Cost-efficiency goal → mass migration heavy on rehost. Agility/innovation goal → cloud-native refactor/rearchitect. Most projects mix strategies. Effort/time/cost rises from rehost/relocate → replatform/repurchase → refactor; optimization opportunity rises the same direction. **Take a phased approach**: migrate first (realize cost/agility), then optimize in later phases (e.g. migrate app, then DB, then move to a serverless stack), monitoring risk and stability.

## Migration risks

Data loss/leakage (encrypt and manage data), downtime (phase migrations / off-peak windows), cost overruns (understand pricing; budget), performance issues (test/optimize post-migration), skill gaps (train/hire), interoperability/integration challenges, compliance (regulated sectors), and **vendor lock-in** (over-reliance on one provider's proprietary services). Migration discovery often surfaces unexpected wins: workload optimization and tighter security.

## The migration process (discover → optimize)

Set up a cloud **Center of Excellence (CoE)** and a standardized **migration factory**. Steps:
1. **Discover** — inventory the portfolio and on-prem workloads: servers, apps, interdependencies, baseline performance, storage, networking, security/compliance, release frequency, DevOps model, escalation paths, OS patching, licensing, external dependencies. Use agent-based or agentless tools (port/packet scanning); run for at least a couple of weeks.
2. **Analyze** — identify dependencies (network/port/process data), group servers (hostname prefixes/tags), and right-size from utilization (not just specs); prioritize by criticality or ease.
3. **Plan** — choose per-app strategy and priority; define success criteria; create **migration waves** (logical groupings deployed sequentially); set up sprint teams, tools, account/network structure, and hybrid connectivity. Order apps by business/technical dimensions, dependencies (locked/tightly/loosely coupled), and the org's prioritization strategy (risk → criticality weight; ease → rehost-first).
4. **Design** — meet the planning success criteria; close architecture gaps; design network (packet flows, routing, firewall rules, isolation, DDoS protection, governance, prod/non-prod separation, multi-account relationships); leverage global infrastructure, scalable/stateless design, and serverless to cut ops complexity. Produce a per-app design document (user accounts, network config, ACLs, hosting, backup, licensing, monitoring, security, maintenance/patching).
5. **Migrate** — execute per the plan; set up the target foundation/core services first; pre-steps (backup/sync, shutdown, unmount); configure networking/firewall/auth/accounts; verify connectivity and logging/monitoring.
6. **Integrate** — connect to other application/system dependencies.
7. **Validate** — functional checks post-migration.
8. **Operate** — plan to run in the cloud.
9. **Optimize** — right-size and adopt cloud-native services (e.g. Lambda/API Gateway/DynamoDB).

## Migration execution: data and servers

**Data migration** — either a single lift-and-shift move, or a hybrid model weighted toward the cloud with legacy on-prem data shifting over time. Estimate data volume to estimate transfer time given bandwidth/connectivity; for large volumes use offline transfer (e.g. Snowball/Snowmobile). Depending on strategy, migrate the whole server (app + infra) or just the data.

**On-prem → cloud example**: rehost web servers and add auto-scaling + load balancers; refactor app servers; replatform the database to a managed cloud DB (e.g. RDS) with a multi-AZ standby; distribute across Availability Zones.

## Hybrid connectivity, multi-cloud, CloudOps

Hybrid connectivity (VPN, dedicated links like Direct Connect) keeps retained on-prem systems talking to migrated cloud apps. Multi-cloud spreads workloads across providers. **CloudOps** operationalizes the cloud estate (governance, monitoring, cost, security) after migration. Migration is also a good moment to adopt or deepen a DevOps culture.
