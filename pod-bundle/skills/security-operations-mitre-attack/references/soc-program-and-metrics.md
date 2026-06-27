# SOC Program, IR Alignment, and Metrics

## Contents
- SOC structure and coverage strategy
- Risk registry
- Incident response and playbook alignment to ATT&CK
- Triage feedback fields
- Core and supporting metrics
- Maturity model
- Compliance crosswalks
- Automation and the AI trajectory

## SOC structure and coverage strategy

There is no one-size-fits-all SOC; structure follows the organization's risk, scale, and growth.
Common roles: SOC analysts, incident responders, detection engineers, threat-intel analysts,
security engineers, red-team/purple-team operators, NOC analysts, and managers. Rough scaling
heuristics (tune to your context): ~1 IR per 200–300 employees; ~1 analyst + 1 security engineer
per 500 employees; ~1 intel analyst per 1,000; ~1 NOC analyst per 5,000 network assets; ~1
manager per 5–10 staff.

**24/7 coverage** options, cheapest to most robust: on-call roster (start here while processes
mature) → shift patterns (e.g., 4-shift 12-hour rotation) → MSSP (smaller in-house team, high
cost) → follow-the-sun (analysts across time zones working local 9–5; also a cost play for
US-based orgs).

**Coverage of telemetry** is the foundation — no logs, nothing to detect. Baseline what's
ingested before any roadmap decisions, find the gaps, prioritize, and fill. When SIEM ingest is
cost-prohibitive, tier logs or route to a **data lake** (e.g., run analytics over cheap storage,
ingest only the detection-relevant subset). The NOC is an extension of the monitoring function
and should get the same trust-but-verify treatment and network-focused purple teaming.

## Risk registry

The registry is where gaps and risks live and get prioritized. Keep it balanced — technical
enough to act on, accessible enough for all stakeholders. Practical columns: Risk ID (optionally
tied to a control like NIST 800-53 CM-6), business area, title, description, category, root cause,
**Impact (1–5)**, **Likelihood (1–5)**, **Risk Score (Impact × Likelihood, 1–25)**, treatment
strategy (Mitigated/Accepted/Deferred/Denied), compensating controls, risk owner, status, notes.

- Auto-escalate Impact when PII / PCI / PHI is in scope.
- Every risk needs an **owner** who accepts it.
- Review quarterly with stakeholders.
- **Tie risks to ATT&CK techniques.** Mapping each risk to its techniques lets you use ATT&CK's
  mitigation/detection guidance to assess impact and likelihood and to drive the detections that
  lower the score. Gaps you won't fix get logged here as accepted risk with a rationale.

## Incident response and playbook alignment to ATT&CK

Every detection should ship with a triage path, and every IR plan benefits from ATT&CK framing.

- **Playbooks** come in three forms, increasing in automation: numbered/bulleted lists →
  flowcharts → SOAR runbooks. Build the list first (pair a senior + junior analyst so no
  knowledge is assumed), turn it into a flowchart to expose decision branches, then automate the
  repeatable parts.
- **Map playbooks to techniques** so analysts know the behavior they're chasing (a phishing
  playbook → T1566; ransomware → Impact techniques like T1486 Data Encrypted for Impact).
- **Automate the mechanical steps.** SOAR can enrich (WHOIS, VirusTotal, DomainTools), open the
  case-management ticket, quarantine a host, push a block, and close repeatable benign alerts —
  reserving human judgment for the ambiguous. Even automating just enrichment and ticketing saves
  hours per week.
- Keep a **roles/contacts roster** (DFIR retainer, legal, comms) as an IR-plan appendix; use
  roles not names so it survives staff churn. Review it routinely.

During an incident, ATT&CK gives a shared language to track what the adversary did (which
techniques, in which order), which feeds both the after-action review and a fresh hunt for the
same TTPs elsewhere.

## Triage feedback fields

Capture structured outcomes at alert closure so detections can be measured and tuned:

- **Closed State**: False Positive / Benign-Expected / Benign-Unexpected / Suspicious / Malicious
  / Other.
- **Value Added**: None / Low / High.
- **Labels**: detection name, tool, ATT&CK technique.

Feed these into dashboards: rules with high "None"/"False Positive" rates get reviewed and tuned;
tools generating most tickets reveal where to focus; per-analyst load informs work balancing.
Without this loop you cannot tell good detections from noise and you operate on a false sense of
security.

## Core and supporting metrics

**Core (baseline everywhere):**
- **MTTD** — mean time to detect.
- **MTTR / MTTC** — mean time to respond / contain / mitigate.
- **Alert efficacy** — true-positive / value-added rate per detection.

Present against SLA targets (e.g., red/yellow/green vs. a detect-within target). Automate
collection via SIEM/case-management APIs rather than manual spreadsheets.

**Supporting / contextual:**
- Coverage % over the relevant matrices (from the Navigator) — but weight by quality, not count.
- Alert volume and noise trends (watch that this doesn't incentivize FP-heavy "coverage").
- Purple-team score distribution over time (stopped/alerted/logged-only/silent).
- Hunts run, detections born from hunts, telemetry gaps found.
- Project vs. operational split (e.g., 60/40), points per sprint for capacity planning.

Metrics justify resourcing: if avg triage is 20 min and you get 200 alerts/day, that's ~67
analyst-hours/day — concrete grounds to fund SOAR or headcount. **Even the absence of a metric
tells a story:** zero detections in an area could mean "locked down" or "no telemetry" — name
which.

## Maturity model

Rough progression: ad hoc reactive alerting → defined processes + risk registry → ATT&CK-mapped
coverage with measured efficacy → regular purple teaming and hunting → automated triage/response
and continuous validation. Markers of maturity: documented policies/standards, quarterly (or
continuous) purple teams and tabletop exercises, a maintained risk registry, detection-as-code,
and metrics-driven prioritization.

## Compliance crosswalks

ATT&CK is **not** a compliance framework — don't audit against it for accreditation. But you can
crosswalk: run a real audit (PCI DSS, HIPAA, NIST 800-53, ISO 27001, SOC 2, DISA STIGs) and map
non-compliant findings to ATT&CK techniques to understand the adversary risk behind each. Example:
MFA maps to NIST 800-53 IA-2, PCI DSS Req. 8, HIPAA 164.312 — and to ATT&CK T1556 (Modify
Authentication Process). Mappings are largely manual; build a reusable crosswalk table once.

Use ATT&CK to drive **policies and standards** where a technique can't be technically mitigated:
a *policy* states intent (e.g., approved-browser-extension policy for T1176), a *standard* sets
enforceable rules (password complexity, golden-image config). Review yearly; audit adherence.

## Automation and the AI trajectory

The clear direction of travel: AI/ML and automation absorb routine triage and rule generation,
shifting analysts toward engineering, hunting, and ML-training roles, with detections increasingly
auto-mapped to frameworks like ATT&CK. Practical near-term uses: LLM assistance for drafting and
explaining detection logic (always review and test the output — never deploy unverified
generated rules), SOAR for response automation, and automated adversary emulation for continuous
validation (see `adversary-emulation-purple-team.md`). Detect-and-respond will keep needing
humans for risk decisions (breach-notification thresholds, liability, root cause) for the
foreseeable future — automation reduces toil, it doesn't remove judgment.
