# Detection Engineering and Rule Writing

## Contents
- Principles
- The detection lifecycle (detection-as-code)
- Picking the right altitude (specificity vs. recall)
- Rule languages and idioms (Sigma, SPL, KQL, Suricata, YARA, EDR)
- Tuning and false-positive reduction
- Pairing detections with triage playbooks
- ROI detections (high-value patterns)
- Anti-patterns

## Principles

- **Detect behavior, not just IOCs.** IOC matches (hashes, IPs, domains) are brittle. Behavioral
  detections mapped to ATT&CK techniques survive infrastructure churn.
- **Every detection has an owner, an ATT&CK tag, and a triage playbook.** A rule with no
  documented response wastes the analyst who catches it.
- **Quantify efficacy from day one.** Capture per-rule outcomes (true positive, benign, noise) so
  you can retire or tune the losers. See `soc-program-and-metrics.md`.
- **Tune toward signal, not toward silence.** The goal is fewer false positives, not fewer
  alerts; a rule tuned until it never fires is worse than no rule because it implies coverage.

## The detection lifecycle (detection-as-code)

Treat detections like software:

1. **Hypothesis / requirement.** State it in plain language, ideally as a user story:
   *"As a SOC, I want to detect a new IAM role granted to a non-admin principal (T1098.003,
   Additional Cloud Roles) so we catch cloud privilege escalation."*
2. **Map to ATT&CK** and identify the data component required (see `mapping-detections.md`).
3. **Author** the rule in the appropriate language, parameterized and commented.
4. **Test** against known-good (no FPs) and known-bad (emulated technique fires it — see
   `adversary-emulation-purple-team.md`).
5. **Review** (peer or senior+junior pairing: senior brings the right logic, junior surfaces
   assumed knowledge).
6. **Deploy** with version control, a unique ID, ATT&CK tags, severity, and a linked playbook.
7. **Measure and tune** continuously; audit rules on a rotating schedule because the environment
   drifts.

Storing rules in Git (Sigma YAML is ideal) gives you diff history, CI validation, and automated
ATT&CK coverage rollups.

## Picking the right altitude

The dominant failure is a rule that is too broad. "Alert on any data transfer over 1 MB" fires
on legitimate uploads, Google Docs, even an emailed image. Make rules specific by:

- Constraining the **source type / index / log channel** first (search less data).
- Adding **behavioral context** (direction, destination category, user/asset role, time-of-day,
  paired events) rather than stacking one-off exclusions.
- Raising thresholds to a level that reflects your acceptable-use policy.

Resist the trap of tuning by appending IP/domain exclusions one at a time until the query is six
lines of `NOT` clauses — that is unmaintainable and burns compute. Instead, re-architect the
logic around the behavior.

## Rule languages and idioms

**Sigma** (vendor-neutral, the lingua franca for shareable detections). Tag every rule with
ATT&CK:

```yaml
title: New Cloud Role Assigned to Non-Privileged Principal
id: 6f1c2a9e-...            # stable UUID
status: experimental
logsource:
  product: aws
  service: cloudtrail
detection:
  selection:
    eventName: 'AttachRolePolicy'
  filter:
    userIdentity.arn|contains: ':role/Admin'
  condition: selection and not filter
falsepositives:
  - Automated IaC pipelines that legitimately attach policies
level: medium
tags:
  - attack.persistence
  - attack.t1098.003        # Additional Cloud Roles
```

**Splunk SPL** — constrain the index/sourcetype, compute, filter, then table for fast triage.
Example for large outbound transfers (note the tightened source typing and human-readable units):

```spl
index=network_data (tag=web OR tag=proxy OR tag=email) sourcetype=meraki_flows
| eval mb_out = bytes_out/1024/1024
| where mb_out >= 50
| table _time, user, src_ip, dest, mb_out
```

Compare with the naive version it replaced — `index=net bytes_out=1000000` — which buried
analysts in benign hits. The improvement is specificity + readable output, not a higher threshold
alone.

**KQL** (Microsoft Sentinel / Defender). Use the analytics rule's native `tactics:` and
`relevantTechniques:` properties so it reports into Sentinel's MITRE coverage blade:

```kql
SigninLogs
| where ResultType !in ("0")                       // failed sign-ins
| summarize failures = count() by IPAddress, bin(TimeGenerated, 1m)
| where failures > 5                               // T1110 Brute Force
```

**Suricata** (network IDS) for protocol/traffic detections; **YARA** for file/memory content
detections. Both pair with ATT&CK via metadata fields. Pick the language to the data: endpoint
process telemetry → Sigma/EDR query language; network flow → Suricata/SPL; files → YARA.

**EDR** vendor query languages (CrowdStrike, SentinelOne, Carbon Black, Defender) excel at
process-lineage and behavioral detections (T1059 Command and Scripting Interpreter, T1055 Process
Injection, T1003 OS Credential Dumping). Prefer them for endpoint behavior over reconstructing
the same logic in the SIEM.

## Tuning and false-positive reduction

- **Triage every closure with a state and a value.** Closed states like *False Positive /
  Benign-Expected / Benign-Unexpected / Suspicious / Malicious / Other*, plus a *Value Added*
  (None / Low / High), let you find rules that are mostly noise (high "None") and target them.
- **Automate the repeatable benign.** If a rule reliably fires on a known-good source (e.g.,
  analytics beacons starting with a known prefix), automate closure via SOAR or a scripted
  check rather than burning analyst minutes — at ~4 min/alert, 20/day is ~90 min/day.
- **Whitelist with care.** Persistent benign-expected hits justify an allow-list entry; review
  the list periodically so it does not become an attacker's blind spot.
- **Audit on rotation.** Re-read deployed rules regularly; the environment changes, and
  yesterday's good rule may now be noisy or obsolete.

## Pairing detections with triage playbooks

A detection without a triage path does not scale. Every rule should ship with a playbook (list,
flowchart, or SOAR runbook) giving the steps to investigate: which logs to pull, which OSINT to
run (VirusTotal, DomainTools, WHOIS), the decision points, and the closure/escalation criteria.
See `soc-program-and-metrics.md` for playbook patterns and IR alignment.

## ROI detections (consistently high value)

These tend to be near-always-actionable, even when the trigger is a misconfiguration:

- **Overly permissive cloud security groups / public buckets** (`0.0.0.0/0` ingress, public ACL).
  Maps to T1562 Impair Defenses (Disable/Modify Cloud Firewall), T1040 Network Sniffing, T1078
  Valid Accounts (Cloud). Cheap to build (GuardDuty / CloudTrail), almost always yields an action.
- **New IAM credentials/roles, device registrations, SSH authorized-keys changes** (T1098 Account
  Manipulation family) — strong cloud-persistence signal.
- **MFA fatigue / repeated push prompts** (T1621 MFA Request Generation) — caught the Uber-style
  intrusion pattern.
- **Cloud resource deletion** (T1485/T1565) — usually a true positive by virtue of being an
  action; quick verify against change tickets.
- **Impossible-travel / anomalous-location auth on remote services** (T1021, T1078) — high value
  when paired with MFA-acceptance location.

## Anti-patterns

- Vague rules with no behavioral context (the "large transfer" trap).
- Tuning by stacking single-value exclusions instead of re-architecting logic.
- Shipping detections without ATT&CK tags, ownership, or a playbook.
- Detecting only IOCs and assuming behavioral coverage.
- Never measuring efficacy, so noisy rules live forever and create false confidence.
- Building many rules at once for matrix completion rather than a few high-signal ones.
