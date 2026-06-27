# CNAPP, CIEM & the Cloud Security Ecosystem

How CSPM relates to the neighboring acronyms, why CSPM alone is not enough, what a CNAPP converges, and the least-privilege (CIEM) workflow.

## Contents
- Why CSPM is not enough
- The acronym map (one line each)
- CNAPP — the converged platform
- CWPP — runtime workload protection
- CASB — SaaS/cloud-traffic broker
- DSPM — data-first posture
- CIEM — entitlements & least privilege (with workflow)
- How they relate to CSPM

## Why CSPM is not enough

CSPM is excellent at one thing: **configuration and compliance posture** of cloud infrastructure. It does **not** cover:
- data encryption/privacy at the data layer (that's DSPM),
- runtime threats — malware, in-memory exploits, container runtime (CWPP),
- identity/access management depth and entitlement right-sizing (CIEM),
- application-layer security (WAF, SAST/DAST),
- user-to-SaaS traffic control and shadow IT (CASB),
- active threat detection / incident response (GuardDuty/Defender XDR/Event Threat Detection + SIEM/SOAR),
- emerging/zero-day threats needing threat intel + behavioral analysis.

So CSPM is **one layer of defense in depth**, not the whole strategy. The pieces combine into a **CNAPP**.

## The acronym map (one line each)

- **CSPM** — is the *configuration* secure and compliant? (control plane, pre-runtime)
- **CWPP** — is the *running workload* protected? (runtime: VMs, containers, serverless)
- **CIEM** — does each *identity* have only the permissions it actually uses? (entitlements)
- **DSPM** — where is *sensitive data*, who can reach it, is it protected? (data-first)
- **CASB** — is *user access to SaaS/cloud apps* visible and controlled? (traffic broker)
- **CNAPP** — a platform that *unifies* CSPM + CWPP + CIEM + IaC scanning + (often) DSPM.

## CNAPP — the converged platform

Gartner's term for a **unified, tightly integrated** set of security and compliance capabilities securing cloud-native apps **from development to runtime**. A CNAPP consolidates previously siloed tools: container scanning, **CSPM**, **IaC scanning**, **CIEM**, runtime **CWPP**, and vulnerability/config scanning — under one data model so it can correlate signals into **attack paths**.

Two guiding principles:
- **Shift left** — embed security, vuln scanning, IaC checks, and compliance from the earliest dev stages.
- **Shield right** — real-time detection and response during runtime.

CNAPPs use **both** instrumentation paradigms: **agent-based** (deep runtime/system-level context) and **agentless** (API/snapshot reads for inventory, known CVEs, audit-log anomalies). The best implementations use both. **CSPM is effectively a subset of a CNAPP.** Examples: Prisma Cloud, Wiz, Orca, Lacework, CrowdStrike Falcon Cloud Security, Microsoft Defender for Cloud (Defender CSPM + plans).

## CWPP — runtime workload protection

Secures **workloads** (VMs, containers, serverless, storage, networking) while they run. Gartner's four essential capabilities: hypervisor-based security, cloud-native app protection (microservices/containers/serverless), DevOps integration, and **API-based controls**. The eight control layers: hardening, configuration, vulnerability management, network firewalling, visibility & micro-segmentation, system-integrity assurance, application control/allowlisting, exploitation-prevention & memory protection. Use cases: malware prevention, IDS/IPS, DLP, vuln management, behavioral analytics, micro-segmentation. Adopts **zero-trust/default-deny** at runtime instead of AV-centric strategies.

**CWPP vs CSPM**: CWPP protects *individual workloads at runtime*; CSPM secures the *overall config/posture of the whole environment* pre-runtime. Most CSPM vendors also sell CWPP (extra license) — use both. (EDR is the endpoint analog of CWPP for laptops/servers/devices.)

## CASB — SaaS/cloud-traffic broker

Sits between users and cloud apps. **Four pillars**: **Visibility** (incl. shadow-IT discovery, impossible-travel detection), **Data Security** (DLP, encryption, access control), **Threat Protection** (UEBA, anomaly detection), **Compliance**. Deployment modes: **API scanning** (data at rest, no real-time block), **forward proxy** (real-time DLP on managed devices), **reverse proxy** (covers unmanaged devices). **CASB vs CSPM**: CASB governs *user interactions with cloud/SaaS apps*; CSPM governs *infrastructure config*.

## DSPM — data-first posture

Provides visibility into **where sensitive data is, who can access it, how it's used, and its security posture**. Capabilities: data discovery, classification (PII/financial/IP), data-flow mapping, encryption/tokenization, access control, DLP, monitoring. **DSPM vs CSPM**: DSPM protects *data wherever it lives*; CSPM secures *the infrastructure*. They complement — DSPM tells CSPM *which* misconfig matters most (the bucket holding crown-jewel data), and CSPM ensures the surrounding infra is hardened. Many CSPM/CNAPP vendors now bundle DSPM.

## CIEM — entitlements & least privilege

CIEM manages and secures **permissions and entitlements** in the cloud to enforce **least privilege (PoLP)**. Capabilities:
- **Entitlement monitoring** — inventory every identity's roles/permissions.
- **Permission-gap detection** — the gap between **granted** and **actually used** permissions (the core CIEM signal).
- **Relationship visualization** — map identity → role → resource (find toxic combinations).
- **Policy modification** — right-size to remove excess.
- **Alerts** — privilege escalation, credential abuse, suspicious access.

**CIEM vs CSPM**: CIEM = *access/entitlement* risk; CSPM = *configuration* risk. They overlap on access control and risk mitigation and are best combined.

### Least-privilege workflow (native CIEM)
1. **Inventory** all identities (users, roles, service accounts) and their grants.
2. **Measure usage**: AWS **IAM Access Advisor** / **Access Analyzer** (generates policies from CloudTrail), Azure **Entra Permissions Management** (Permissions Creep Index), GCP **IAM Recommender** (90-day usage).
3. **Find the gap**: flag wildcards (`*`), broad managed roles (`AdministratorAccess`/`Owner`/`Contributor`/`roles/editor`), unused keys, stale principals, shared credentials.
4. **Right-size**: replace with scoped predefined/custom roles; remove unused permissions; require MFA; eliminate long-lived keys (federate via SSO / Workload Identity / OIDC).
5. **Add JIT**: Azure **PIM**, AWS IAM Identity Center permission sets with approval, GCP privileged access — elevate on demand, time-boxed (JIT + JEA).
6. **Monitor & alert** on entitlement changes and privilege escalation; re-run periodically (entitlements drift).

Vendors offering CIEM: Microsoft (Entra Permissions Management), Tenable, CyberArk, BeyondTrust, plus CNAPPs (Wiz, Prisma Cloud, Orca, Lacework).

## How they relate to CSPM

```
                 ┌─────────────────────────── CNAPP ───────────────────────────┐
   shift left →  │  IaC scanning → CSPM (config) → CIEM (identity) ─┐           │
                 │                                                   ├→ attack  │ → shield right
   data layer →  │  DSPM (sensitive data)                           │   paths  │   CWPP (runtime)
                 └───────────────────────────────────────────────────┘         │
   user/SaaS edge:  CASB (separate but complementary)                          │
```
Use **CSPM** to fix the configuration, **CIEM** to fix who can reach it, **DSPM** to know what's worth protecting, **CWPP** to defend it at runtime, **CASB** to govern SaaS access — and a **CNAPP** to correlate them into prioritized attack paths instead of disconnected alerts.
