---
name: security-operations-mitre-attack
description: >-
  Run threat-informed security operations with the MITRE ATT&CK framework. Use when mapping
  detections, telemetry, alerts, or SIEM/EDR rules to ATT&CK tactics, techniques, and
  sub-techniques (T-codes like T1078, T1059, T1110); doing coverage and gap analysis with the
  ATT&CK Navigator; writing or tuning detection rules (Splunk SPL, Sigma, KQL, Suricata, YARA)
  and cutting false positives; building threat-hunting hypotheses; planning purple teams or
  adversary emulation with Atomic Red Team or Caldera; integrating cyber threat intelligence;
  aligning incident response playbooks and runbooks; or measuring SOC maturity and metrics
  (MTTD, MTTR, alert efficacy). Also triggers on phrases like threat-informed defense,
  detection engineering, coverage map, ATT&CK matrix, technique mapping, SOC metrics,
  emulation plan, blue team, red team, hunt.
---

# Security Operations with MITRE ATT&CK

This skill equips Claude to plan and run threat-informed security operations: mapping
telemetry and detections to MITRE ATT&CK, engineering high-signal detection rules, framing
threat hunts, emulating adversaries through purple teaming, doing coverage and gap analysis,
and measuring SOC effectiveness.

## When to use this skill

- Mapping a detection, log source, alert, or incident to ATT&CK tactics/techniques (T-codes).
- Writing, reviewing, or tuning detection content (Splunk SPL, Sigma, KQL, Suricata, YARA, EDR rules).
- Building a coverage map or gap analysis (often with the ATT&CK Navigator) and prioritizing what to close.
- Forming a threat-hunting hypothesis from a technique, an intel report, or a "what would I be missing?" question.
- Planning a purple team or adversary-emulation exercise (Atomic Red Team, Caldera) and scoring results.
- Turning cyber threat intelligence (a group, a campaign, a report) into prioritized detections and hunts.
- Aligning IR playbooks/runbooks to techniques, or measuring SOC maturity (MTTD/MTTR, alert efficacy, coverage).
- Any time someone says "threat-informed defense," "detection engineering," "ATT&CK matrix,"
  "emulation plan," or asks why an alert is noisy.

## Core concepts

**The ATT&CK knowledge base.** ATT&CK is a living, curated catalog of real adversary behavior,
organized as **tactics** (the adversary's goal — *why*), **techniques** and **sub-techniques**
(*how* they achieve it), and **procedures** (specific observed implementations, often tied to
named groups and software). It is not a kill chain and not a compliance checklist — it is a
shared vocabulary for behavior. There are three top-level matrices: **Enterprise** (with
platform-specific views: Windows, macOS, Linux, Cloud/IaaS/SaaS/Office 365/Entra ID/Google
Workspace, Network, Containers), **Mobile** (Android, iOS), and **ICS**.

**Identifiers.** Tactics are `TAxxxx` (e.g., TA0001 Initial Access). Techniques are `Txxxx`
(e.g., T1059 Command and Scripting Interpreter); sub-techniques append `.xxx`
(T1059.001 PowerShell). A single technique can sit under multiple tactics — e.g., Valid
Accounts (T1078) appears in Initial Access, Persistence, Privilege Escalation, and Defense
Evasion. That many-to-many relationship is why you map to *behavior*, not to a single column.

**Threat-informed defense.** Drive detection, hunting, and engineering from what real
adversaries do against environments like yours — not from a generic feature checklist. ATT&CK
is the connective tissue: intel describes who and how, ATT&CK encodes the how as techniques,
detections and hunts cover those techniques, and emulation verifies the coverage actually works.

**Do not chase 100% of the matrix.** ATT&CK is a guide, not an accreditation. Many techniques
do not apply to your environment, and some (e.g., predicting novel malware) are not directly
detectable. Trying to "complete" the matrix produces noisy, low-value detections and burns the
team out. Prioritize by your risk, your platforms, and your telemetry.

**Versioning matters.** ATT&CK ships roughly twice a year; techniques get added, deprecated,
renamed, or split. As of mid-2026 the framework is well past v11 (191 techniques) — Enterprise
now spans 200+ techniques and 450+ sub-techniques across 14 tactics. Always note the ATT&CK
version a mapping or Navigator layer was built against, and re-baseline when you upgrade.

## Workflow: how to approach ATT&CK-driven SOC tasks

The recurring loop is **map → cover → hunt → emulate → measure → improve.** Most requests are
one stage of it.

1. **Scope to the environment.** Identify the platforms in play (Windows-heavy enterprise? AWS
   IaaS? M365? containers?) and pick the matching matrices. The Cloud, Network, and macOS
   matrices differ substantially from Windows — do not map an AWS finding against the Windows
   matrix. See `references/attack-framework.md`.

2. **Map behavior to ATT&CK.** Translate the detection, log source, alert, or intel into the
   tactics/techniques it actually observes. Map to *all* plausible techniques (one log source or
   gap usually touches several) so nothing is silently uncovered, then narrow by what your
   telemetry can really see. Capture the data sources each technique needs. See
   `references/mapping-detections.md`.

3. **Assess coverage and find gaps.** Lay current detections over the matrix (Navigator is the
   standard tool) and identify uncovered, high-priority techniques. Distinguish *no telemetry*
   from *telemetry but no detection* from *detection but ineffective* — they look identical on a
   coverage map but demand different fixes. See `references/navigator-and-coverage.md`.

4. **Prioritize.** Rank gaps by impact × effort, by your risk registry, by intel relevance (what
   groups target your sector), and by synergy (fixes that close several gaps at once). Don't lead
   with the highest-impact item if it's also the highest-effort — sequence some quick wins.

5. **Engineer detections or hunts.** For a covered-but-weak or uncovered technique, either write
   a detection (and a triage playbook for it) or, when the behavior is too ambiguous for a rule,
   run a hunt. Every detection needs a tuning plan and a measure of efficacy — false-positive
   noise is the dominant failure mode. See `references/detection-engineering.md` and
   `references/threat-hunting.md`.

6. **Validate by emulation.** Prove the detection/mitigation works by safely executing the
   technique (Atomic Red Team for atomic tests, Caldera for chained operations) in a purple-team
   setting and scoring the outcome (stopped / alerted / logged-only / silent). Feed results back
   into steps 3–5. See `references/adversary-emulation-purple-team.md`.

7. **Measure and report.** Track MTTD, MTTR/MTTC, alert efficacy (value-added), and coverage
   over time. Use the data to justify resourcing and to retire noisy rules. See
   `references/soc-program-and-metrics.md`.

### Quick reference: the 14 Enterprise tactics (execution order, roughly)

Reconnaissance (TA0043) → Resource Development (TA0042) → Initial Access (TA0001) →
Execution (TA0002) → Persistence (TA0003) → Privilege Escalation (TA0004) →
Defense Evasion (TA0005) → Credential Access (TA0006) → Discovery (TA0007) →
Lateral Movement (TA0008) → Collection (TA0009) → Command and Control (TA0011) →
Exfiltration (TA0010) → Impact (TA0040).

An adversary rarely walks these in a straight line; they loop (discover → move → discover) and
revisit. Map observed behavior to the tactic that matches the adversary's *intent at that step*.

## Common pitfalls and anti-patterns

- **Treating ATT&CK as a compliance scorecard.** Chasing matrix completion adds noisy, valueless
  detections. Map to your risk and platforms; leave inapplicable techniques uncovered on purpose
  and document why (the absence of a detection can itself tell a story).
- **Mapping to a single technique.** Behaviors and gaps usually touch several techniques across
  tactics. Enumerate all plausible mappings first, then prune — otherwise you leave silent gaps.
- **Confusing telemetry gaps with detection gaps.** A blank cell on the Navigator can mean "no
  log source," "logs but no rule," or "rule but it doesn't fire." Each needs a different remedy;
  label them distinctly.
- **Shipping detections without a tuning/feedback loop.** The classic failure: a vague rule (e.g.,
  "alert on any large data transfer") that buries analysts in false positives until they ignore
  it. Quantify efficacy, tune iteratively, and automate closure of repeatable benign hits.
- **Boiling the ocean.** Implementing many controls at once (SSL decrypt + new NDR + dozens of
  fresh alerts simultaneously) yields half-finished projects and FP storms. Prioritize, sequence,
  and finish.
- **Heuristic-only validation gives false confidence.** Confirming a control "looks right" or runs
  one canned test is not validation. Adversary emulation research shows defenses that survive
  shallow, heuristic probing often fail under systematic, optimization-driven testing — emulate
  the technique properly and score it, then re-test after every change.
- **Stale mappings.** ATT&CK changes; a Navigator layer or detection map built two versions ago
  may reference renamed or deprecated techniques. Re-baseline on upgrade and stamp the version.
- **Buying tools to fix a process gap.** Underutilized existing tooling is extremely common.
  Tune and fully use what you have, map the gap to ATT&CK, and only then justify new spend.

## Reference files

- `references/attack-framework.md` — ATT&CK structure (tactics/techniques/sub-techniques/
  procedures), the matrices and platform views, identifiers, versioning, ATT&CK vs. the Cyber
  Kill Chain and other threat models (PASTA, STRIDE, VAST, Trike, attack trees). Read when you
  need the mental model or to choose the right matrix.
- `references/mapping-detections.md` — How to map log sources, telemetry, alerts, and intel to
  techniques; the ATT&CK Data Sources / data components model; worked mappings; common
  one-gap-to-many-techniques examples. Read when mapping anything to ATT&CK.
- `references/detection-engineering.md` — Writing and tuning detections (Sigma, SPL, KQL,
  Suricata, YARA), reducing false positives, detection-as-code, ROI detections, pairing rules
  with triage playbooks. Read when authoring or fixing detection content.
- `references/threat-hunting.md` — Hypothesis-driven and intel-driven hunting, baselining
  "normal," the hunt loop, promoting hunts to detections. Read when framing a hunt.
- `references/adversary-emulation-purple-team.md` — Purple teaming, emulation plans, Atomic Red
  Team and Caldera, scoring rubrics, the test-and-harden loop, safety. Read when planning or
  scoring emulation/purple-team work.
- `references/navigator-and-coverage.md` — ATT&CK Navigator layers, building coverage and gap
  maps, scoring schemes, combining layers (detections + intel + emulation), prioritization. Read
  for coverage/gap analysis.
- `references/soc-program-and-metrics.md` — SOC structure and coverage strategy, risk registry,
  IR/playbook alignment, metrics (MTTD/MTTR, efficacy), maturity, compliance crosswalks, and the
  AI/automation trajectory. Read for program-level and metrics questions.
