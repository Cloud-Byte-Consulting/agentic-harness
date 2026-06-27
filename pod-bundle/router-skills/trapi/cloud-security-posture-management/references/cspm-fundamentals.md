# CSPM Fundamentals

Conceptual model, the assessment engine, agent vs agentless, risk prioritization, and how to select tooling.

## Contents
- Shared responsibility — the reason posture matters
- What a CSPM tool actually does
- Why misconfigurations happen
- Agent-based vs agentless
- Risk prioritization and attack paths
- Native vs third-party vs open source
- Running a PoC
- Tool landscape (quick reference)

## Shared responsibility — the reason posture matters

Security in the cloud is split between provider and customer. The provider is responsible for **security *of* the cloud** — the physical data centers, hardware, hypervisor, and the internals of managed services. The customer is responsible for **security *in* the cloud** — and exactly what that covers depends on the service model:

| Layer | On-prem | IaaS (EC2/VM/GCE) | PaaS (Lambda/App Service/App Engine) | SaaS (M365/Workspace) |
|---|---|---|---|---|
| Data, identities, access | Customer | Customer | Customer | Customer |
| Application | Customer | Customer | Customer | Provider |
| OS, runtime, patching | Customer | **Customer** | Provider | Provider |
| Network controls | Customer | Customer | Shared | Provider |
| Hypervisor, hardware, facility | Customer | Provider | Provider | Provider |

The rule that never changes: **the customer always owns accounts/identities, devices, and data.** Misconfiguration of the customer's half — open network rules, over-permissive IAM, unencrypted data, disabled logging — is the dominant root cause of cloud breaches (Gartner: the large majority of cloud security failures are customer-caused, and nearly all successful attacks trace to customer-side misconfiguration/mismanagement). CSPM exists to continuously check that half.

## What a CSPM tool actually does

Five core capabilities (Gartner's original 2019 definition of the category):

1. **Asset discovery & inventory** — enumerate every resource across multi-cloud accounts; maintain a live inventory. You cannot secure invisible assets, and cloud resources are created/destroyed rapidly.
2. **Configuration assessment** — evaluate each resource against CIS benchmarks, provider best practices, and custom policy.
3. **Misconfiguration & compliance detection** — flag drift from secure baselines and map it to frameworks (PCI DSS, HIPAA, NIST, ISO, SOC 2).
4. **Risk prioritization** — rank findings by real risk, not raw severity.
5. **Remediation** — guided steps or automated fixes; integrate with DevOps/SIEM/SOAR/ticketing.

It is **continuous and automation-driven** — manual review cannot keep up with cloud scale and velocity.

## Why misconfigurations happen

- **Human error / skill gap** — biggest cause; misunderstanding the cloud model, wrong manual settings, typos.
- **Insufficient access controls** — unauthorized changes to settings.
- **Automation issues** — untested scripts/templates making unintended changes at scale.
- **Lack of monitoring/maintenance** — drift goes unnoticed.
- **Improper deployments / legacy systems** — poorly planned migrations.

Mitigation: least privilege, encryption everywhere, change auditing, codified security policy, automated scanning, MFA, and training.

## Agent-based vs agentless

| | Agent-based | Agentless |
|---|---|---|
| How | Lightweight sensor on each workload | Reads cloud APIs / snapshots |
| Strengths | Deep runtime/OS-level visibility, real-time, works offline | Fast/easy deploy, broad coverage, no overhead, scales |
| Weaknesses | Install/update overhead at scale | Less granular runtime detail; API latency; bound by API quality |

Most modern CSPM is **agentless-first** (control-plane API reads). CNAPPs increasingly do **both** — agentless for inventory/config/known-CVE/audit-log anomalies, agent for real-time runtime context. Snapshot/side-scanning gives near-agent depth without an agent on the workload.

## Risk prioritization and attack paths

Raw severity is a trap. Combine:
- **Exposure** — internet-reachable? (`0.0.0.0/0`, public IP, public bucket, exposed API endpoint)
- **Asset criticality** — prod vs dev; regulated/sensitive data? (drive this from tags/labels and a CMDB)
- **Exploitability** — known CVE? actively exploited in the wild? PoC public? (feed from CTI — see vulnerability prioritization below)
- **Identity blast radius** — can a foothold here assume an over-privileged role and move laterally?

**Attack-path analysis** is the mature form: instead of 50 disconnected findings, the tool surfaces the chain — e.g., *public-facing VM → instance role with `s3:*` → bucket holding PII*. Fix the choke point in the path and many findings collapse at once. This is the single biggest lever against alert fatigue.

**Vulnerability prioritization inputs** (when posture overlaps with vuln management): asset criticality, **CVSS base score**, asset context (a CVSS 10 on staging ≠ on prod), patch availability, exploitation likelihood, and compliance mandates. CVE = the identifier; CVSS = the 0.0–10.0 severity. Use CTI to mark what is actively exploited and prioritize those first.

## Native vs third-party vs open source

**Native (AWS Security Hub/Config, Azure Defender for Cloud, GCP SCC)**
- Choose when: single primary cloud, cost-sensitive, deep provider integration, simple environment, limited integration needs.
- Pros: seamless integration, low/no extra cost, vendor trust, fast enablement.
- Cons: weak cross-cloud, fewer advanced/custom features, possible bias/blind spots about the provider's own services, slower feature cadence than specialists.

**Third-party / CNAPP (Prisma Cloud, Wiz, Orca, Lacework, CrowdStrike Falcon Cloud Security, Trend Conformity, Dome9/Check Point CloudGuard)**
- Choose when: multi-cloud/hybrid, need attack-path analysis, unified CSPM+CWPP+CIEM+DSPM, advanced detection (behavioral/anomaly/AI), specific compliance depth, integration with existing SIEM/SOC.
- Pros: consistent multi-cloud, advanced features, independent validation, customization, specialized expertise.
- Cons: licensing cost, integration effort, vendor dependency/lock-in, learning curve, sharing security data externally.

**Open source (Prowler, ScoutSuite, Cloud Custodian, Steampipe + Powerpipe, OpenSCAP, CloudMapper)**
- Choose when: cost is paramount, single-account audits, want full transparency/customization, strong in-house skills.
- Avoid when: strict compliance/regulatory needs, limited in-house expertise, need vendor support/SLAs.
- Cautions: support/docs gaps, more setup effort, possibly narrower feature set; keep them patched.

Many orgs run **native + third-party together** — native for cheap broad coverage and provider depth, third-party for cross-cloud correlation and attack paths.

## Running a PoC

Before buying, run a scoped proof of concept in a test environment that mirrors production:
1. Define objectives, scope, key use cases.
2. Pick measurable **metrics**: asset-discovery accuracy, **false-positive rate**, **time-to-detect**, **time-to-remediate**, compliance coverage, user satisfaction, ROI.
3. Assemble a team (security + cloud ops + compliance).
4. Onboard, run an initial assessment, exercise concrete use cases (simulate an incident, add a resource and confirm discovery, trigger a remediation workflow, test integrations).
5. Evaluate against metrics, compute ROI = (Net Benefits / Cost) × 100%, document, recommend.

Selection factors: meets security needs, supports your clouds, integrates with existing tools (SIEM/ticketing/IdP), ease of use, automation depth, scalability, reporting/customization, vendor reputation/support, pricing/licensing model and TCO (perpetual vs subscription, scope = per-account/per-resource).

## Tool landscape (quick reference)

- **Native**: AWS Security Hub + AWS Config; Microsoft Defender for Cloud; GCP Security Command Center; Oracle Cloud Guard.
- **Third-party / CNAPP**: Palo Alto **Prisma Cloud**, **Wiz**, **Orca Security**, **Lacework**, **CrowdStrike Falcon Cloud Security**, **Trend Cloud One – Conformity**, Check Point **CloudGuard (Dome9)**, Sophos Cloud Optix, Tenable, Microsoft Defender CSPM.
- **Open source**: **Prowler** (AWS/Azure/GCP), **ScoutSuite** (multi-cloud audit), **Cloud Custodian** (policy-as-rules engine, multi-cloud), **Steampipe/Powerpipe** (SQL over cloud + CIS mods), **OpenSCAP**, **CloudMapper** (AWS attack-surface viz).

(Vendor capabilities evolve quickly; verify current features and the latest Gartner Magic Quadrant / Peer Insights before deciding.)
