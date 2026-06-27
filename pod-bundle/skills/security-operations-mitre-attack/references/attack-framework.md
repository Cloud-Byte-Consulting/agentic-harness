# The ATT&CK Framework: Structure, Matrices, and Context

## Contents
- The object model: tactics, techniques, sub-techniques, procedures
- Identifiers and how to read a technique page
- The matrices and platform views
- Versioning and change management
- ATT&CK vs. the Cyber Kill Chain
- Where ATT&CK fits among threat models (PASTA, STRIDE, VAST, Trike, attack trees)

## The object model

ATT&CK is a curated knowledge base of adversary behavior observed in the real world. Four
object types matter day to day:

- **Tactic** — the adversary's tactical *goal*, the "why" of an action. Example: Credential
  Access (TA0006) — the adversary wants account credentials. There are 14 Enterprise tactics.
- **Technique** — a general *means* of achieving a tactic, the "how." Example: OS Credential
  Dumping (T1003). A technique can belong to more than one tactic.
- **Sub-technique** — a more specific means under a technique. Example: LSASS Memory
  (T1003.001). Not every technique has sub-techniques.
- **Procedure** — a specific, observed in-the-wild implementation, usually attributed to a
  **group** (e.g., APT29) or a piece of **software**/malware (e.g., a named RAT). Procedures are
  the concrete examples on a technique page and are gold for writing precise detections.

Two more objects connect the behavior catalog to defense and to intel:

- **Mitigations** — preventive configurations/controls mapped to techniques (e.g., "Multi-factor
  Authentication," "Privileged Account Management," "User Training").
- **Data sources / data components** — the telemetry that lets you observe a technique (e.g.,
  *Process: Process Creation*, *Command: Command Execution*, *Logon Session: Logon Session
  Creation*). These are the bridge from technique to detection — see `mapping-detections.md`.
- **Groups** and **Software** — named threat actors and tools, each linked to the techniques they
  use. These power intel-driven prioritization and emulation-plan scoping.

## Identifiers and reading a technique page

- Tactic: `TAxxxx` (TA0001 = Initial Access).
- Technique: `Txxxx` (T1059 = Command and Scripting Interpreter).
- Sub-technique: `Txxxx.xxx` (T1059.001 = PowerShell; T1574.001 = DLL Search Order Hijacking).

A technique page on attack.mitre.org gives you: the description, the tactic(s) it belongs to,
the platforms it applies to, **procedure examples** (groups/software that used it), **mitigations**,
**detection** guidance (which data sources/components reveal it), and references. When you are
asked "how do I detect Txxxx?", start from the data components listed there, then translate to
your telemetry and rule language.

A single technique frequently spans multiple tactics. Valid Accounts (T1078) is the canonical
case: Initial Access (logging in with stolen creds), Persistence (keeping that access),
Privilege Escalation (the account is privileged), and Defense Evasion (legitimate creds blend
in). This is why you always map behavior, never assume one technique = one column.

## The matrices and platform views

ATT&CK has three top-level domains:

- **Enterprise** — IT environments. Includes platform-specific views you select per environment:
  - Windows, macOS, Linux
  - Cloud: IaaS, SaaS, Office 365, **Entra ID** (formerly Azure AD), Google Workspace
  - Network (network devices: routers, switches, firewalls)
  - Containers (orchestration and runtime)
- **Mobile** — Android, iOS.
- **ICS** — Industrial Control Systems / OT.

The matrices are **pick-and-choose**. Map an AWS S3 finding against the IaaS/Cloud matrix, not
Windows; a phishing-into-M365 case against Office 365 + Entra ID. The matrices differ a lot in
size and content:

- **Windows** is the largest single platform view — it has been a target the longest and absorbs
  the bulk of malware. Expect deep coverage of Execution, Defense Evasion, Credential Access
  (e.g., OS Credential Dumping with sub-techniques like LSASS Memory, NTDS, DCSync, LSA Secrets),
  and Lateral Movement (RDP, SMB/Admin Shares, DCOM, WinRM under Remote Services T1021).
- **Cloud** views are smaller and heavily weighted toward identity: Valid Accounts, Account
  Manipulation (additional cloud credentials/roles, SSH authorized keys, device registration),
  Use Alternate Authentication Material (application access tokens, web session cookies),
  MFA Request Generation (MFA fatigue), and Modify Cloud Compute Infrastructure (create/delete/
  revert instances, create snapshots) under IaaS.
- **Network** is vague and small by design — it names perimeter devices and public-facing apps
  only at a high level (Network Boundary Bridging, Network Sniffing, Traffic Signaling / Port
  Knocking, Modify System Image, Weaken Encryption). You will heavily tailor detections to your
  own gear and traffic.
- **macOS** aligns more with Linux than Windows and has distinct sub-techniques (Plist File
  Modification, Resource Forking, Right-to-Left Override under Masquerading, the `Hide500Users`
  plist setting under Hidden Users).

Choosing the wrong matrix is a common early mistake — it produces mappings and detections that
cannot apply to the platform you actually run.

## Versioning and change management

ATT&CK is released roughly twice a year. Each release adds, deprecates, renames, splits, or
merges techniques and updates groups/software. For scale: it launched in 2015 with 9 tactics and
~96 techniques; by v11 (2022) it was 14 tactics / 191 techniques / 386 sub-techniques; by
mid-2026 Enterprise spans 200+ techniques and 450+ sub-techniques. Azure AD became **Entra ID**;
techniques have been renamed and renumbered over time.

Practical rules:
- **Stamp the version** on every mapping, Navigator layer, and coverage report (e.g., "ATT&CK
  v16, Enterprise").
- **Re-baseline on upgrade**: diff the release notes, remap deprecated/renamed techniques, and
  re-validate Navigator layers and detection tags.
- Prefer pulling the machine-readable **STIX/TAXII** bundle or the Python `mitreattack-python`
  library for programmatic mapping rather than hardcoding technique names.

## ATT&CK vs. the Cyber Kill Chain

The Lockheed Martin **Cyber Kill Chain** has seven linear stages: Reconnaissance, Weaponization,
Delivery, Exploitation, Installation, Command & Control, Actions on Objectives. ATT&CK differs
in two ways that matter:

1. **Granularity.** The Kill Chain stops at stages; ATT&CK decomposes each into concrete
   techniques, sub-techniques, procedures, mitigations, and detections.
2. **Non-linearity and platform-awareness.** The Kill Chain is flat and one-size-fits-all.
   ATT&CK's tactics are not a strict sequence — adversaries loop (Discovery → Lateral Movement →
   Discovery) — and it provides platform-specific matrices.

Use the Kill Chain for high-level narrative/briefings; use ATT&CK for engineering, hunting, and
coverage.

## Where ATT&CK fits among threat models

ATT&CK is a *behavioral knowledge base*, complementary to threat-modeling methodologies you use
during design and risk work:

- **PASTA** (Process for Attack Simulation and Threat Analysis) — 7-step, risk-centric: define
  objectives, scope, decompose the app, analyze threats, analyze vulnerabilities, model attacks,
  risk/impact. Thorough; good for scoped strategy and informational campaigns. Its attack-modeling
  step pairs naturally with ATT&CK technique enumeration.
- **STRIDE** (Spoofing, Tampering, Repudiation, Information disclosure, DoS, Elevation of
  privilege) — design-phase, developer-friendly, maps to the CIA triad. Six categories can be
  limiting/less thorough than PASTA.
- **VAST** (Visual, Agile, Simple Threat) — casts a wide net across technical and operational
  flows; tool-driven (e.g., ThreatModeler); can be overwhelming at scale.
- **Trike** — risk/requirements-centric, paired with risk registries and audits; can miss risks
  outside the initial requirement set.
- **Attack trees** — the original model; logical tree of an attack from initial vector to
  outcomes. Pairs with any of the above and maps cleanly onto ATT&CK technique chains.

In practice you combine these: threat-model the design (PASTA/STRIDE), enumerate concrete
adversary behavior with ATT&CK, then verify with emulation. There is no one-size-fits-all model.
