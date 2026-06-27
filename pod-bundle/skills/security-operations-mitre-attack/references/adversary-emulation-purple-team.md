# Adversary Emulation and Purple Teaming

## Contents
- Why emulate
- Purple teaming defined
- Planning an exercise
- Atomic Red Team (atomic tests)
- Caldera (chained operations)
- Adversary-emulation plans
- Scoring rubric
- The test-and-harden loop
- Why heuristic / shallow validation is not enough
- Safety and rules of engagement

## Why emulate

You do not know whether a detection or mitigation works until an adversary behavior actually
exercises it. Adversary emulation safely executes ATT&CK techniques against your environment to
**validate coverage**, **find gaps**, **train analysts**, and **prove the negative** (demonstrate
that an attack succeeds with no visibility — a powerful argument for more telemetry). It is the
verification stage of the threat-informed loop (map → cover → hunt → emulate → measure).

## Purple teaming defined

A **purple team** is a collaborative exercise: a red action (offense) executes pre-agreed
techniques while the blue team (defense) watches in real time, knowing what is being run and
when. The point is not to "win" but to jointly test and improve detections and response. It is
cheaper than a covert red team (the cost is mostly people-time) and produces immediately
actionable findings. Run some form of purple team on a regular cadence (quarterly is a common
baseline; a mature team does it continuously). Network/NOC capabilities deserve the same
treatment via network-focused exercises.

## Planning an exercise

1. **Purpose and scope.** Evaluate a new SIEM/EDR, test response to a specific scenario, train
   staff, or prove gaps. Decide before engaging the red operator.
2. **Select techniques.** Drive the technique list from intel (what groups target you) and from
   your coverage gaps (your Navigator gap layer). Emulating a relevant group beats running random
   tests.
3. **Choose tooling.** Atomic Red Team for discrete technique tests; Caldera for chained,
   autonomous operations; both are ATT&CK-aligned and open source.
4. **Build a test plan.** A table keyed by technique ID, with the exact command/test run, the
   expected observable, and a result score (rubric below). The plan doubles as a record that the
   commands themselves were correct.
5. **Execute collaboratively**, capture evidence (screenshots, timestamps, alert IDs), and score.
6. **Report** with findings, the score distribution, and prioritized recommendations.

## Atomic Red Team

Open-source library (Red Canary) of small, portable **atomic tests**, each mapped to an ATT&CK
technique. Each test is defined in YAML with the technique ID, supported platforms, the
executor, input args, and cleanup commands.

- Browse the catalog by technique (folders are named `T1059.001`, etc.).
- Run via the PowerShell module `Invoke-AtomicRedTeam`:

```powershell
Import-Module Invoke-AtomicRedTeam
Invoke-AtomicTest T1059.001 -ShowDetailsBrief   # list atomics for PowerShell
Invoke-AtomicTest T1059.001 -TestNumbers 1      # run one
Invoke-AtomicTest T1059.001 -Cleanup            # always clean up afterward
```

Atomics are ideal for **detection unit-testing**: run one technique, confirm your rule fires,
record the outcome. Keep tests scoped, use `-Cleanup`, and run only in approved environments.

## Caldera

MITRE's open-source automated adversary-emulation platform. Where atomics are single behaviors,
Caldera **chains** abilities (each ability maps to a technique) into an **adversary profile** and
runs them through an agent on target hosts, making decisions as it goes. Use it to emulate a
multi-stage operation (initial access → discovery → credential access → lateral movement → C2 →
exfil) and observe which steps your stack catches. Drive profiles from real group TTPs for
realism.

## Adversary-emulation plans

For higher fidelity, emulate a *specific* threat actor end to end. Workflow:
1. Pick a group relevant to your sector (from intel; the group's ATT&CK Groups page lists its
   techniques).
2. Build an ordered plan that follows that group's known kill chain, technique by technique.
3. Map each step to an Atomic test or a Caldera ability.
4. Execute and score, then patch coverage where the group would have gone undetected.

MITRE's Center for Threat-Informed Defense publishes full emulation plans for several groups —
use them as templates rather than starting blank.

## Scoring rubric

Score every test on a consistent scale so results are comparable across exercises. A practical
five-point scheme:

1. **Stopped** — attack identified and blocked by existing controls.
2. **Alerted, not stopped** — succeeded but generated an alert.
3. **Logged only** — succeeded, generated events, but no alert fired (detection gap).
4. **Silent** — succeeded with no events and no alert (telemetry gap).
5. **Failed (test error)** — attack failed due to a bad command, not a control. Re-run.

Tally the distribution (e.g., "of 13 tests: 2 stopped, 3 alerted, 4 logged-only, 4 silent, 0
errors") to show posture and pinpoint where to invest. Logged-only → write a detection;
silent → add telemetry first.

## The test-and-harden loop

Emulation is not a one-shot audit. After every detection or mitigation change, **re-run the test**
and confirm the score improved. This closes the loop: emulate → find gap → engineer fix →
re-emulate → confirm. Treat it as continuous regression testing for your defenses.

## Why heuristic / shallow validation is not enough

A control that "looks right" or passes one canned check is not validated. Research on
adversary-emulation rigor shows that defenses which survive shallow, heuristic probing often fall
to systematic, optimization-driven testing — heuristic checks give a false sense of security,
whereas thorough, automated emulation reliably surfaces the weakness. Two implications for SOC
work:

- **Test breadth and depth, not one happy-path command per technique.** Vary the procedure (the
  same technique has many implementations) so you measure the detection's true robustness, not
  its luck against one sample.
- **Automate and repeat at scale.** Manual, occasional testing under-samples the attack space.
  Automated emulation (Atomic in CI, scheduled Caldera operations) catches regressions and covers
  more variants — and the captured results feed directly back into tuning and, where applicable,
  into hardening the controls themselves.

## Safety and rules of engagement

- Run only in approved, scoped environments with written authorization and a defined window.
- Notify the blue team (purple) or the designated white-cell controller.
- Prefer non-destructive tests; for destructive techniques (Impact: ransomware-like encryption,
  disk wipe), use lab/staging only.
- Always clean up artifacts and document exactly what was run, where, and when.
- Keep a roster of roles/contacts and an escalation path in case a test trips a real response.
