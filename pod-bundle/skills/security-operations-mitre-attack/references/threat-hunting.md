# Threat Hunting with ATT&CK

## Contents
- When to hunt vs. when to detect
- Sources of hunting hypotheses
- The hunt loop
- Baselining "normal"
- Worked hunt examples
- Promoting a hunt to a detection
- Documenting and measuring hunts

## When to hunt vs. when to detect

Write a **detection** when the behavior is crisp enough to express as a reliable rule. **Hunt**
when the behavior is too ambiguous, too context-dependent, or too novel to alert on without
drowning in false positives — or when you suspect an adversary is already inside and want to look
proactively. Hunting is the human-driven, hypothesis-first search through telemetry; successful
hunts often graduate into detections.

## Sources of hunting hypotheses

A good hypothesis is specific and falsifiable, and it names the ATT&CK technique(s) in scope.
Generate them from:

1. **A technique** — "Adversaries use DCSync (T1003.006). If it's happening, I'd see replication
   requests from non-domain-controller hosts." Walk the matrix for high-risk, low-coverage
   techniques (use your Navigator gap layer — see `navigator-and-coverage.md`).
2. **Threat intelligence** — a group active in your sector uses specific TTPs; hunt for those
   first. Map the report to techniques, then hunt the ones you can't reliably detect.
3. **Crown jewels** — start from the assets that matter and work backward through the techniques
   an adversary would need (Collection, Exfiltration, Impact).
4. **Anomaly / "what's weird"** — driven by a baseline of normal (below).
5. **Post-incident** — after an incident, hunt for the same TTPs elsewhere in the estate.

## The hunt loop

1. **Hypothesis** — state the behavior, the technique ID, and the expected observable.
2. **Scope the data** — which log sources / data components reveal it; confirm you actually
   collect them (a hunt that needs telemetry you don't have is really a telemetry-gap finding).
3. **Query and pivot** — search broadly, then pivot on suspicious results (process lineage, auth
   chains, network peers). Most of a hunt is iterative narrowing.
4. **Triage findings** — separate benign-unusual from malicious. Document both; "proving the
   negative" (showing an attack path is invisible) is itself a valuable result that justifies new
   telemetry or detections.
5. **Act** — escalate true positives to IR; convert durable patterns into detections; file
   telemetry/coverage gaps into the backlog and risk registry.

## Baselining "normal"

You cannot find anomalies without knowing normal. A recurring SOC shortfall is having no baseline
of ordinary network and host behavior to compare against. Build baselines for:

- **Network** — typical talkers, protocols, volumes, beacon-like periodicity (and which periodic
  traffic is *legitimate*, e.g., analytics/telemetry SDKs, software update checks).
- **Hosts** — golden-image processes, scheduled tasks, services, autoruns; deviations are leads.
- **Identity** — normal logon locations/times, service-account behavior, admin activity windows.

Golden images and configuration baselines make "what changed?" answerable, which is the heart of
most hunts and of NOC/SOC trust-but-verify work.

## Worked hunt examples

- **Living-off-the-land execution (T1059):** hunt for `powershell.exe`/`cmd.exe`/`wscript.exe`
  spawned by Office apps, or encoded/`-enc` PowerShell, or LOLBins (rundll32, mshta, regsvr32)
  reaching out to the network. Pivot on parent-child process lineage and outbound connections.
- **Credential dumping (T1003):** look for handle access to LSASS by non-system processes, NTDS
  access off a DC, or replication (DRSUAPI) traffic from non-DC hosts (DCSync).
- **Cloud persistence (T1098):** new long-lived access keys, added IAM roles, registered MFA
  devices, or SSH authorized-keys edits on instances — correlate against change tickets.
- **C2 beaconing (T1071 / T1573):** periodic, low-jitter outbound to rare destinations; baseline
  out the legitimate periodic traffic first or you'll chase analytics SDKs all day.
- **Discovery bursts (T1046/T1018/T1087):** a single host enumerating many services/hosts/accounts
  in a short window is a classic pre-lateral-movement tell.

## Promoting a hunt to a detection

When a hunt repeatedly surfaces the same actionable pattern with manageable noise, codify it:
write the rule (`detection-engineering.md`), tag it with the technique, attach a triage playbook,
and add the technique to your covered Navigator layer. This is how hunting steadily expands
automated coverage instead of staying a one-off effort.

## Documenting and measuring hunts

- Record each hunt: hypothesis, technique(s), data sources, queries, findings (incl. negative
  results), and follow-up actions. The negatives feed your coverage story and risk registry.
- Track simple metrics: hunts run, new detections born, telemetry gaps found, true positives
  surfaced. Over time these show the program maturing from purely reactive alerting toward
  proactive defense — the direction the field is heading as automation absorbs routine triage.
