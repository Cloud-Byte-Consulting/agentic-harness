# ATT&CK Navigator, Coverage Maps, and Gap Analysis

## Contents
- What the Navigator is
- Layer file structure
- Building a coverage layer
- Building an intel / threat layer
- Combining layers with layer math
- Scoring schemes
- Reading the map: three kinds of gap
- Prioritizing what to close
- Re-baselining across ATT&CK versions

## What the Navigator is

The **ATT&CK Navigator** is the standard web tool for visualizing the matrix and annotating it
with **layers**. A layer colors, scores, and comments on techniques so you can see at a glance
what you detect, what an adversary uses, where they overlap, and where you are blind. It is the
primary instrument for coverage and gap analysis and for communicating posture to stakeholders.

You can run the public hosted Navigator or self-host it; layers are portable JSON, so they live
naturally in version control alongside detection content.

## Layer file structure

A layer is JSON. The fields you touch most:

```json
{
  "name": "Detection Coverage - 2026 Q2",
  "versions": { "attack": "16", "navigator": "5.x", "layer": "4.5" },
  "domain": "enterprise-attack",
  "description": "Current SIEM/EDR detection coverage",
  "techniques": [
    { "techniqueID": "T1059.001", "score": 75, "color": "",
      "comment": "Sigma rule sig-ps-encoded; alerts, some FPs",
      "metadata": [{ "name": "rule_id", "value": "6f1c2a9e" }] },
    { "techniqueID": "T1003.001", "score": 0,
      "comment": "No LSASS-access detection; telemetry exists" }
  ],
  "gradient": { "colors": ["#ff6666", "#ffe766", "#8ec843"], "minValue": 0, "maxValue": 100 }
}
```

Always set `versions.attack` — a layer is meaningless without knowing which ATT&CK version its
technique IDs belong to.

## Building a coverage layer

1. Inventory deployed detections and tag each with its technique(s) — ideally pull these straight
   from your detection repo's ATT&CK tags (Sigma `tags`, Sentinel rule techniques, etc.) so the
   layer regenerates automatically.
2. Assign each technique a **score** reflecting *effective* coverage, not mere existence (see
   scoring schemes).
3. Color by gradient (red = uncovered, green = strong). Add comments linking to rule IDs.
4. Generate programmatically where possible (`mitreattack-python`) so the map stays current.

## Building an intel / threat layer

For a threat-informed view, build a second layer of the techniques a relevant adversary uses.
Pull a group's technique list from its ATT&CK Groups page, or map an intel report's TTPs (see
`mapping-detections.md`), and score those techniques (e.g., all = 1) to highlight them.

## Combining layers with layer math

The Navigator can compute a new layer from existing ones via per-technique expressions
(`a` and `b` reference the input layers). The high-value combination:

- Layer **a** = threat/intel coverage (techniques the adversary uses, scored 1).
- Layer **b** = your detection coverage (scored 0–100).
- New layer score = `a - (b/100)` or similar → highlights techniques the adversary uses that you
  do **not** cover well. That delta is your prioritized work list.

You can layer in a third dimension — **emulation results** (which techniques you actually
validated via Atomic/Caldera) — to distinguish "we think we cover it" from "we proved we cover
it." Techniques that are intel-relevant, claimed-covered, but emulation-failed are the most urgent.

## Scoring schemes

Pick one and apply it consistently:

- **Binary** — 0 (no coverage) / 100 (covered). Simple, but hides quality.
- **Maturity tiers** — e.g., 0 none, 25 telemetry only, 50 detection exists, 75 detection +
  validated, 100 detection + validated + automated response. Recommended: it encodes the three
  gap types directly.
- **Confidence-weighted** — score by detection efficacy (true-positive rate from your metrics).

Avoid letting the map reward quantity over quality — a technique "covered" by a 90%-false-positive
rule should not score green.

## Reading the map: three kinds of gap

A blank or red cell means one of three very different things — label which:

1. **Telemetry gap** — you don't collect the data source the technique needs. *Fix: onboard the
   log source.* (Score ~0–25.)
2. **Detection gap** — you have the telemetry but no rule. *Fix: engineer a detection.* (~50.)
3. **Efficacy gap** — you have a rule but it's noisy or doesn't fire on real procedures. *Fix:
   tune, and re-validate by emulation.* (~50–75 capped until proven.)

Treating all three as "just write a rule" is the classic coverage-map mistake — the first one
needs data engineering, not detection engineering.

## Prioritizing what to close

Rank the gap list by a blend of:

- **Risk** — impact × likelihood, tied to your risk registry.
- **Intel relevance** — techniques used by groups targeting your sector weigh more.
- **Effort** — use an impact-vs-effort quadrant; sequence some quick wins alongside the big
  high-impact / high-effort items so the team banks momentum.
- **Synergy** — fixes that close several gaps at once (one log source lighting up many
  techniques) go first.

Document gaps you choose *not* to close in the risk registry with an owner and rationale — an
explicit accepted risk, not a silent hole.

## Re-baselining across ATT&CK versions

When ATT&CK updates, technique IDs can be added, renamed, deprecated, split, or merged. After an
upgrade: diff the release notes, remap affected technique IDs in every layer, re-tag detection
content, and bump `versions.attack`. A layer pinned to an old version will silently mis-render
against the current matrix and overstate or understate coverage.
