---
name: cloud-security-posture-management
description: >-
  Assess and harden cloud security posture across AWS, Azure, and GCP. Use when finding or
  fixing cloud misconfigurations, enabling or interpreting AWS Security Hub / AWS Config,
  Microsoft Defender for Cloud, or GCP Security Command Center, scanning Terraform /
  CloudFormation / ARM / Bicep IaC for security issues (Checkov, tfsec, Trivy, KICS, cfn-nag),
  mapping to CIS benchmarks or PCI DSS / NIST / ISO 27001 / SOC 2 / HIPAA / FedRAMP,
  prioritizing cloud risk and attack paths, managing cloud entitlements and over-permissive
  IAM (CIEM), choosing native vs third-party CSPM or CNAPP tools (Prisma Cloud, Wiz, Orca,
  Lacework, Defender CSPM), shifting security left in CI/CD, or triaging alerts about public
  S3 buckets, open security groups, 0.0.0.0/0 rules, unencrypted storage, exposed databases,
  public IPs, or excessive permissions. Covers posture, not runtime malware (that is CWPP).
---

# Cloud Security Posture Management (CSPM)

This skill equips Claude to assess, prioritize, and remediate misconfigurations and identity risk across AWS, Azure, and GCP using native and third-party posture tools, and to shift posture checks left into IaC and CI/CD.

## When to use this skill

- Finding why a cloud account is failing a security audit, or hardening it proactively.
- A user mentions a **misconfiguration** symptom: public S3/blob/GCS bucket, security group open to `0.0.0.0/0`, unencrypted EBS/disk/database, public IP on a sensitive VM, over-permissive IAM role, disabled logging, no MFA on root/admin.
- Enabling or interpreting **AWS Security Hub / AWS Config**, **Microsoft Defender for Cloud**, or **GCP Security Command Center**.
- Mapping a cloud environment to **CIS Benchmarks, PCI DSS, NIST CSF / 800-53, ISO 27001, SOC 2, HIPAA, FedRAMP**.
- Scanning **Terraform, CloudFormation, ARM/Bicep, Kubernetes** manifests for security issues before deploy (shift-left).
- **Risk prioritization** / attack-path questions ("which of these 4,000 findings matter?").
- **CIEM**: right-sizing entitlements, detecting unused permissions, enforcing least privilege.
- Choosing **native vs third-party** posture tooling, or evaluating **CNAPP** (Prisma Cloud, Wiz, Orca, Lacework, Defender CSPM).

Boundary: CSPM is about *configuration and identity posture*. Runtime workload defense (malware, in-memory exploits, container runtime) is **CWPP**; data-level discovery/classification is **DSPM**; user-to-SaaS traffic control is **CASB**. A **CNAPP** unifies these. See `references/cnapp-and-ciem.md`.

## Core concepts

**Why misconfiguration is the top cloud risk.** The dominant cause of cloud breaches is customer-side misconfiguration, not provider failure — Gartner has long held that the large majority of cloud security failures are the customer's fault. The **shared responsibility model** is why: the provider secures *the* cloud (hardware, hypervisor, managed-service internals); the customer secures *in* the cloud (IAM, network rules, encryption choices, data, OS patching for IaaS). Posture management is the discipline of continuously checking the customer's half.

**What CSPM does.** A CSPM tool continuously: (1) **discovers** an asset inventory across accounts/subscriptions/projects, (2) **assesses** each resource's configuration against benchmarks and policies, (3) **detects** misconfigurations, compliance drift, and risky identities, (4) **prioritizes** findings by real risk, and (5) **remediates** — guided or automated. It is primarily **agentless** (reads cloud control-plane APIs), which makes it fast to onboard and broad in coverage; some tools add agents or snapshot scanning for deeper workload visibility.

**Posture vs runtime.** CSPM operates at the **pre-runtime / control-plane** layer (is this resource configured safely?). It does not inspect packets, block malware, or do EDR. Pair it with CWPP/CNAPP for defense in depth.

**The finding lifecycle.** detection → triage (is it a true positive? what's the blast radius?) → prioritize (severity × exposure × asset criticality × exploitability) → remediate (fix config / IaC) → verify (re-scan) → prevent recurrence (shift the check left into IaC/CI).

## Workflow / how to approach CSPM tasks

### 1. Establish the asset inventory and scope
You cannot secure what you cannot see. Confirm which **accounts (AWS), subscriptions/management groups (Azure), and projects/folders/organization (GCP)** are in scope, and onboard them centrally:
- **AWS**: aggregate via an Organization, a delegated **Security Hub** administrator, and **Config aggregators**; onboard CSPM via a cross-account IAM role with a read-only/security-audit policy.
- **Azure**: enable **Defender for Cloud** at the **Management Group** root so new subscriptions inherit it.
- **GCP**: enable **Security Command Center** at the **organization** node so all projects/folders are covered.

Tag/label assets (owner, environment, data-classification, criticality) — these attributes drive prioritization later.

### 2. Run a posture assessment against benchmarks
Pick the **CIS Benchmark** for each cloud as the baseline (the providers ship CIS-aligned standards natively), then layer the regulatory frameworks the org must meet (PCI DSS, HIPAA, NIST, ISO 27001, SOC 2). Each native tool maps findings to controls:
- AWS Security Hub → CIS AWS Foundations, AWS FSBP (Foundational Security Best Practices), PCI DSS, NIST 800-53.
- Defender for Cloud → **Microsoft Cloud Security Benchmark (MCSB)** by default, plus CIS, PCI DSS, NIST, ISO.
- GCP SCC → CIS GCP, PCI DSS, NIST, ISO via the built-in compliance dashboard.

See per-cloud detail in `references/aws-posture.md`, `references/azure-posture.md`, `references/gcp-posture.md`; framework detail in `references/compliance-frameworks.md`.

### 3. Triage and prioritize — do not chase raw severity
A flat list of "1,200 HIGH findings" is noise. Prioritize by combining signals:
- **Exposure**: is the resource reachable from the internet (`0.0.0.0/0`, public IP, public bucket)?
- **Asset criticality**: production vs staging; does it hold regulated/sensitive data?
- **Exploitability / active exploitation**: is there a known CVE being exploited in the wild? Feed this from CTI.
- **Identity blast radius**: can a compromise here reach an over-privileged role and move laterally?

A `CVSS 9.8` on an isolated dev box matters less than a `CVSS 7` on an internet-facing prod server holding PII. Modern tools express this as **attack paths** ("public VM → permissive role → admin → crown-jewel data"); fix the *path*, not 50 disconnected findings. See `references/cspm-fundamentals.md` for the risk-prioritization model.

### 4. Remediate the misconfiguration
Identify the misconfiguration class (see `references/remediation-and-iac-scanning.md` for the full catalog with fixes):
- **Network**: restrict ingress from `0.0.0.0/0` to specific CIDRs; close unused ports; default-deny; segment tiers.
- **Identity (IAM)**: enforce least privilege; remove unused/stale principals and keys; require MFA; eliminate wildcard `Action`/`Resource`; avoid shared credentials.
- **Data**: enable encryption at rest and in transit; block public access on object storage; secure backups; never hardcode secrets.
- **Logging/monitoring**: enable CloudTrail / Activity Logs / Cloud Audit Logs; centralize; alert on critical changes.

Prefer **fixing the source IaC** over a console click — a console fix gets overwritten on the next deploy (config drift). For one-off live fixes use the CLI (`aws`, `az`, `gcloud`), then backport into the template.

### 5. Shift left: scan IaC and enforce in CI/CD
The cheapest place to fix a misconfiguration is before it deploys. Scan templates in the pipeline and fail the build on policy violations:
- **Terraform / multi-IaC**: `checkov`, `tfsec` (now folded into **Trivy**), `terrascan`, KICS.
- **CloudFormation**: `cfn-nag`, `cfn-guard`, Checkov, KICS.
- **ARM/Bicep**: Checkov, `templateAnalyzer`, KICS.
- **Kubernetes**: Checkov, KICS, `kube-score`, Trivy.

Codify org rules as **Policy as Code** (OPA/Rego, Checkov custom policies, cfn-guard). Adopt **immutable infrastructure** and DRY/modular templates so a fix in one module propagates. See `references/remediation-and-iac-scanning.md`.

### 6. Manage identities and entitlements (CIEM)
Cloud IAM is the new perimeter and the hardest to right-size. Use CIEM capabilities (native: AWS IAM Access Analyzer + Access Advisor, Azure Entra Permissions Management, GCP IAM Recommender; or CNAPP) to find the **gap between granted and used permissions**, visualize identity→resource relationships, and strip excess. See `references/cnapp-and-ciem.md`.

### 7. Choose the right tooling
- **Native** (Security Hub/Config, Defender for Cloud, SCC) when single-cloud, cost-sensitive, deep provider integration matters. Pros: seamless, cheap, no extra onboarding. Cons: weak cross-cloud, possible blind spots about the provider's own services.
- **Third-party / CNAPP** (Prisma Cloud, Wiz, Orca, Lacework, CrowdStrike, Trend Conformity) when multi-cloud, advanced attack-path analysis, unified CSPM+CWPP+CIEM+DSPM, or specific compliance depth is needed.
- **Open source** (Prowler, ScoutSuite, Cloud Custodian, Steampipe, OpenSCAP) for cost-free single-account audits and automation.
Run a scoped **PoC** with success metrics (asset-discovery accuracy, false-positive rate, time-to-detect/remediate, compliance coverage) before buying. See `references/cspm-fundamentals.md`.

## Common pitfalls & anti-patterns

- **Assuming the provider secures everything.** The provider secures *of* the cloud; misconfigured IAM, network, and encryption are always yours. State the shared-responsibility line explicitly.
- **Treating every HIGH as equal.** Prioritize by exposure + criticality + exploitability + identity blast radius, not raw CVSS/severity. Un-prioritized findings cause alert fatigue and real risks get buried.
- **Fixing in the console, not the code.** Live console fixes get reverted by the next IaC apply (config drift). Fix the template; the live fix is a stopgap.
- **Wildcards in IAM.** `"Action": "*"` / `"Resource": "*"` and broad `roles/owner`/`Contributor` grants violate least privilege. Scope tightly; remove unused permissions revealed by CIEM/Access Advisor.
- **`0.0.0.0/0` ingress** to SSH/RDP/databases/admin ports. Restrict to known CIDRs or use bastion/SSM/IAP; default-deny.
- **Public object storage by accident.** Turn on account-level public-access blocks (S3 Block Public Access, Azure storage "disallow public", GCP public-access prevention) rather than per-bucket toggles.
- **Onboarding at the wrong scope.** Enable posture tools at org/management-group root so new accounts inherit coverage; per-account enablement leaves gaps.
- **CSPM as the whole strategy.** CSPM finds configuration risk only — it is not malware detection, DLP, or IR. Layer CWPP/CNAPP, logging/SIEM, and CASB/DSPM as needed.
- **Disabling/ignoring logging.** Without CloudTrail / Activity Logs / Cloud Audit Logs you cannot detect or investigate; these are foundational CIS controls.
- **Auto-remediation without guardrails.** Blind auto-fixes can break prod. Gate them with testing, scoping, and rollback; start in alert-only mode.
- **Confusing the acronyms.** CSPM (config posture) ≠ CWPP (runtime workloads) ≠ CIEM (entitlements) ≠ DSPM (data) ≠ CASB (SaaS traffic). CNAPP = the converged platform.

## Reference files

- **`references/cspm-fundamentals.md`** — Open when you need the conceptual model: shared responsibility, what CSPM does, agent vs agentless, risk prioritization & attack paths, native vs third-party vs open-source selection and PoC. Includes the tool landscape.
- **`references/aws-posture.md`** — Open for AWS specifics: Security Hub, AWS Config + conformance packs, IAM Access Analyzer, GuardDuty pairing, common AWS misconfigs and CLI fixes.
- **`references/azure-posture.md`** — Open for Azure specifics: Microsoft Defender for Cloud (CSPM plans), Secure Score, MCSB, Entra Permissions Management, Azure Policy, common Azure misconfigs and `az` fixes.
- **`references/gcp-posture.md`** — Open for GCP specifics: Security Command Center tiers, Security Health Analytics, IAM Recommender, Organization Policy, common GCP misconfigs and `gcloud` fixes.
- **`references/compliance-frameworks.md`** — Open when mapping to a standard: CIS Benchmarks, PCI DSS, NIST CSF/800-53, ISO 27001, SOC 2, HIPAA, FedRAMP, CSA CCM, plus governance vs compliance and how native tools map controls.
- **`references/remediation-and-iac-scanning.md`** — Open for the misconfiguration catalog (network/IAM/data/logging) with fixes, IaC scanners (Checkov, tfsec/Trivy, KICS, cfn-nag/cfn-guard), Policy as Code, and shift-left CI/CD integration.
- **`references/cnapp-and-ciem.md`** — Open for the ecosystem and convergence: CNAPP, CWPP, CASB, DSPM, CIEM — what each does, how they relate to CSPM, and least-privilege entitlement workflows.
