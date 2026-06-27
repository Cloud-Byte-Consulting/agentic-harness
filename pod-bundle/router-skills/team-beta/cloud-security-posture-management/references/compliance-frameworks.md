# Compliance Frameworks & Governance

Which standard means what, how CSPM tools map to them, and the governance context. Lead with the practical mapping, then the per-framework detail.

## Contents
- Governance vs compliance (and why both)
- How native tools map controls (quick table)
- CIS Benchmarks
- PCI DSS
- NIST CSF and NIST 800-53
- ISO/IEC 27001
- SOC 2
- HIPAA
- FedRAMP / FISMA
- GDPR / CCPA / CPRA / PDPA (privacy)
- CSA Cloud Controls Matrix
- Cloud-provider frameworks (AWS WAF, MCSB)
- Global vs regional considerations

## Governance vs compliance (and why both)

- **Compliance** = meeting *external* rules (laws, regulations, industry standards). Reactive/operational — "are we adhering to the rules?"
- **Governance** = the *internal* framework of policies, decision-making, resource management, and risk tolerance that guides cloud use. Proactive/strategic — it *sets the rules* and includes compliance as a subset.
Governance sets the tone; compliance operationalizes it. CSPM is the engine that *continuously demonstrates* both: it assesses live config against codified policy, produces audit-ready reports, and drives remediation.

## How native tools map controls (quick table)

| Framework | AWS | Azure (Defender for Cloud) | GCP (SCC Premium) |
|---|---|---|---|
| CIS Benchmark | Security Hub: CIS AWS Foundations | Built-in CIS Azure | Built-in CIS GCP |
| PCI DSS | Security Hub PCI standard; Config pack | Regulatory compliance dashboard | Compliance dashboard |
| NIST 800-53 | Security Hub NIST standard; Config pack | Built-in NIST initiative | Compliance dashboard |
| ISO 27001 | Config conformance pack | Built-in ISO initiative | Compliance dashboard |
| SOC 2 | Config conformance pack / Audit Manager | Built-in SOC 2 initiative | (via custom / partner) |
| HIPAA | Config pack / Audit Manager | Built-in HIPAA HITRUST initiative | (via custom / partner) |
| Provider baseline | FSBP / AWS WAF | **MCSB** (default) | Google security foundations |

CSPM doesn't *make* you compliant — it gives continuous evidence, finds the gaps, and maps each finding to the control(s) it affects, replacing point-in-time manual audits with continuous assurance.

## CIS Benchmarks

The practical starting point. Vendor-neutral, community-developed configuration guidelines maintained by the **Center for Internet Security**, with per-platform benchmarks (CIS AWS Foundations, CIS Azure, CIS GCP, CIS Kubernetes, etc.). They include a **scoring mechanism** so you can quantify posture and track improvement, and they **align with NIST/ISO**, making them a bridge to broader frameworks. Use the CIS benchmark as the baseline, then customize for your risk profile (not every recommendation applies everywhere). All three clouds ship CIS-aligned standards natively.

## PCI DSS

For any entity that **stores, processes, or transmits payment card data**. Maintained by the PCI SSC; **12 high-level requirements** covering network security, access control (least privilege, unique IDs, strong auth), **encryption** of cardholder data in transit and at rest, **vulnerability management**, **network segmentation** (to shrink scope), logging/monitoring, regular testing, and an incident-response plan. Validated by a **QSA** or **SAQ**. CSPM directly supports the config, encryption, segmentation, logging, and vuln-management requirements.

## NIST CSF and NIST 800-53

- **NIST Cybersecurity Framework (CSF)** — flexible, risk-based, organized into core functions: **Identify, Protect, Detect, Respond, Recover** (CSF 2.0 adds **Govern**). Not prescriptive; a common language for risk. Widely adopted beyond critical infrastructure.
- **NIST SP 800-53** — the detailed control catalog (the basis for FedRAMP). CSPM standards in all three clouds map to 800-53 control IDs.

## ISO/IEC 27001

International standard for an **Information Security Management System (ISMS)**. Risk-based, built on the **Plan-Do-Check-Act** cycle, with **Annex A** controls (cryptography, access control, physical security, IR, etc.). Certification by an accredited body demonstrates a mature, continually-improving security program. Applies to any org handling sensitive information.

## SOC 2

AICPA framework reporting on controls against the **Trust Services Criteria**: **Security, Availability, Processing Integrity, Confidentiality, Privacy**. **Type 1** = control design at a point in time; **Type 2** = operating effectiveness over a period (typically 6+ months). Conducted by independent auditors; the resulting report is shared with customers/partners. Common requirement for SaaS and cloud service providers. CSPM provides the continuous control evidence auditors want.

## HIPAA

US healthcare law protecting **PHI/ePHI**. Key rules: **Privacy Rule**, **Security Rule** (safeguards for ePHI — encryption, access controls, risk assessments), and **Breach Notification Rule**. Applies to covered entities and business associates. CSPM enforces the technical safeguards (encryption, access control, audit logging) the Security Rule requires.

## FedRAMP / FISMA

- **FedRAMP** — standardized security assessment/authorization for cloud services used by US **federal agencies**, built on **NIST 800-53**, with **Low/Moderate/High** impact tiers, JAB or agency authorization, continuous monitoring, and the FedRAMP Marketplace of authorized offerings.
- **FISMA** — the broader federal law mandating security programs for government information systems (confidentiality, integrity, availability).

## GDPR / CCPA / CPRA / PDPA (privacy)

Privacy regulations CSPM supports indirectly (by enforcing encryption, access control, data residency, and logging):
- **GDPR** (EU, extraterritorial) — data-subject rights, consent, DPIAs, DPOs, **72-hour breach notification**, large fines (up to €20M or 4% of global revenue), privacy-by-design, data-residency/transfer rules.
- **CCPA / CPRA** (California) — consumer rights to know/delete/opt-out; CPRA adds the CPPA regulator and "sensitive personal information."
- **PDPA** (Singapore), and many other regional laws. Data **residency** and **sovereignty** are the cloud-specific levers — pin regulated data to compliant regions and verify via posture checks.

## CSA Cloud Controls Matrix (CCM)

Cloud Security Alliance framework: control objectives organized into domains (~17), tailored to cloud, harmonized with other standards, with the **CAIQ** questionnaire for assessing providers. Useful for vendor/provider due diligence and as a cloud-specific control baseline.

## Cloud-provider frameworks

- **AWS Well-Architected Framework (WAF)** — five pillars: **Operational Excellence, Security, Reliability, Performance Efficiency, Cost Optimization** (plus Sustainability). The Security pillar guides access control, encryption, and detection. Reviewed with the Well-Architected Tool. (Distinct from AWS *Web Application Firewall*, also "WAF".)
- **Microsoft Cloud Security Benchmark (MCSB)** — Defender for Cloud's default benchmark; consolidates **CIS, NIST, PCI DSS** into Azure-specific controls and per-service baselines. Evolution of the Azure Security Benchmark.

## Global vs regional considerations

Multi-region orgs must satisfy both **global** standards (GDPR's extraterritorial reach, ISO 27001) and **regional** ones (CCPA for California residents, PDPA for Singapore). Adapt governance to the org: understand the cloud landscape, identify applicable regulations, assess current state, customize policies/controls, classify and protect data, enforce least privilege + encryption, audit continuously, vet vendors, train staff, and iterate. Map each obligation to concrete CSPM checks and a compliance dashboard so adherence is continuous and demonstrable.
