# Mapping Detections, Telemetry, and Intel to ATT&CK

## Contents
- The mapping mindset
- The Data Sources / data components model
- How to map a log source
- How to map an alert or detection rule
- How to map an intel report or finding
- Worked examples
- One gap touches many techniques (worked enumerations)
- Output formats

## The mapping mindset

Map to **behavior**, not to a tool or a single column. The goal is a faithful statement of which
adversary techniques a given thing (log source, rule, alert, gap, control) can *observe* or
*relates to*. Two rules:

1. **Enumerate broadly, then prune.** A log source or a gap almost always touches several
   techniques across multiple tactics. List every plausible technique first so nothing is
   silently uncovered, then narrow to the ones your telemetry can actually evidence. Leaving one
   out creates a blind spot that a coverage map will report as "covered."
2. **Separate "relates to" from "detects."** A mitigation (e.g., MFA) *relates to* many
   techniques but *detects* none. A log source provides *visibility* into techniques; a rule
   *detects* a subset of those. Keep these distinct or your coverage map will overclaim.

## The Data Sources / data components model

ATT&CK formalizes telemetry as **data sources** and finer-grained **data components**. This is
the connective tissue between a technique and a detection. Examples:

| Data source | Data component (examples) |
|---|---|
| Process | Process Creation; OS API Execution |
| Command | Command Execution |
| Logon Session | Logon Session Creation; Logon Session Metadata |
| File | File Creation; File Modification; File Access |
| Network Traffic | Network Connection Creation; Network Traffic Flow; Network Traffic Content |
| Windows Registry | Windows Registry Key Modification; Key Creation |
| Cloud Service | Cloud Service Modification; Cloud Service Enumeration |
| Module | Module Load |
| Authentication | (logon/auth events, MFA prompts) |

Workflow: open the technique page → read its listed data sources/components → ask "do I collect
that telemetry?" If no, you have a **telemetry gap** (collect first). If yes, ask "do I have a
rule on it?" If no, a **detection gap** (engineer a rule). If yes, "does the rule fire well?" If
not, an **efficacy gap** (tune). These three look identical on a coverage map but need different
fixes — always label which one you mean.

## How to map a log source

1. Identify what the source records (e.g., VPN/zero-trust auth logs, EDR process telemetry,
   CloudTrail, DNS, Windows Security 4688/Sysmon, M365 Unified Audit Log).
2. List the data components it supplies.
3. Enumerate every technique those components can evidence.
4. Note which are *detectable* vs. merely *visible given correlation*.

Example — missing **authentication / VPN / remote-access logs** map to (non-exhaustive):
T1133 External Remote Services, T1021 Remote Services, T1078 Valid Accounts,
T1098 Account Manipulation, T1570 Lateral Tool Transfer, T1046 Network Service Discovery.
The exact set depends on your configuration; enumerate generously, then confirm what you can see.

## How to map an alert or detection rule

State the technique(s) the *logic* detects, not the data source's full visibility. A periodic
beaconing alert detects C2 behavior:
T1071 Application Layer Protocol (Web/DNS/Mail/FTP), T1132 Data Encoding,
T1001 Data Obfuscation. Tag the rule with these in your detection repo (Sigma's `tags:` field
takes `attack.t1071` style labels) so coverage tooling can roll them up automatically.

## How to map an intel report or finding

For a named group, campaign, or report:
1. Extract the behaviors described (TTPs), not just the IOCs.
2. Map each to a technique/sub-technique; pull the group's existing ATT&CK technique list from
   the Groups page as a starting set.
3. Build a Navigator layer of those techniques (see `navigator-and-coverage.md`).
4. Compare against your coverage layer to find what this adversary could do undetected.
5. Prioritize closing those gaps and seed an emulation plan from the same technique set.

This is the core of **threat-informed defense**: intel → techniques → coverage delta → detections
+ hunts + emulation.

## Worked examples

**Risk: unencrypted PII in an S3 bucket (Cloud/IaaS).**
Techniques: T1530 Data from Cloud Storage; T1119 Automated Collection; T1552 Unsecured
Credentials (Private Keys); T1550 Use Alternate Authentication Material (Application Access
Token). Mitigations from ATT&CK: audit bucket access, encrypt at rest, filter network traffic,
require MFA, restrict file/dir permissions, proper account management. Detection: enable
CloudTrail; alert on access patterns / public-ACL changes (e.g., a `0.0.0.0/0` grant).

**Risk: patching does not meet SLA on EOL Windows (Windows matrix).**
Techniques to consider: T1190 Exploit Public-Facing Application, T1210 Exploitation of Remote
Services, T1203 Exploitation for Client Execution, T1068 Exploitation for Privilege Escalation,
T1546 Event Triggered Execution (Application Shimming), T1195 Supply Chain Compromise,
T1574 Hijack Execution Flow (DLL Side-Loading). Narrow to the specific EOL systems; map only
what their exposure justifies.

**Multi-factor authentication** (as a control) relates to a wide spread of techniques — among
them T1621 MFA Request Generation, T1110 Brute Force, T1111 MFA Interception, T1098 Account
Manipulation, T1021 Remote Services, T1078 Valid Accounts, T1556 Modify Authentication Process,
T1539 Steal Web Session Cookie, T1550 Use Alternate Authentication Material. This breadth is why
MFA is high-leverage — but remember it *mitigates*, it does not *detect*.

## One gap touches many techniques (worked enumerations)

- **Overly permissive ACLs / security groups open to the internet:** T1069 Permission Groups
  Discovery, T1046 Network Service Discovery, T1557 Adversary-in-the-Middle, T1563 Remote Service
  Session Hijacking, T1562 Impair Defenses (Disable/Modify Cloud Firewall).
- **No / immature security awareness training:** T1566 Phishing (Spearphishing Link/Attachment),
  T1598 Phishing for Information, T1550 Use Alternate Authentication Material, T1098 Account
  Manipulation. Training has a wide blast radius — map the second-order techniques too.
- **Privileged-access mismanagement:** spans T1548 Abuse Elevation Control Mechanism, T1134
  Access Token Manipulation, T1098 Account Manipulation, T1078 Valid Accounts, T1003 OS
  Credential Dumping, T1558 Steal or Forge Kerberos Tickets, T1550 Use Alternate Authentication
  Material, and more across many tactics.

## Output formats

When you produce a mapping, prefer a structured table the reader can load into Navigator or a
detection repo:

```
| Asset / Source / Gap | Tactic | Technique (ID) | Sub-technique | Data component | Visible? | Detected? | Notes |
```

For detection content, emit ATT&CK tags inline (Sigma `tags`, Splunk savedsearch
`action.notable.param.security_domain` + a custom `mitre_technique` field, Sentinel analytics
rule `tactics`/`techniques` properties) so coverage rolls up without manual bookkeeping.
