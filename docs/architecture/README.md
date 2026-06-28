# Architecture — consolidation and stack ownership

Canonical docs for the **AIR + Omnigent consolidation** (independence of tooling, zero overlap of functionality).

> New to the platform? Start at the [**getting-started hub**](../../GETTING_STARTED.md) for the pillars and how they fit, then dive into the docs below.

| Document | Purpose |
| :--- | :--- |
| [governed-platform-architecture.md](governed-platform-architecture.md) | **Start here** — the as-built architecture of Open Engine on the AIR/Omnigent backbone (planes, lifecycle, governance, ownership) |
| [open-engine-integration-plan.md](open-engine-integration-plan.md) | Execution plan — six-lane sequencing, CP-1..CP-4 gates, risks |
| [checkpoint-manual-tests.md](checkpoint-manual-tests.md) | Runnable manual tests per gate (CP-1..CP-4): input / expected output / usage check |
| [consolidation-plan.md](consolidation-plan.md) | Single-owner matrix, repo disposition, phased execution, open decisions |
| [policy-opa-hooks-audit.md](policy-opa-hooks-audit.md) | OPA-only authz, Sentry MCP gap, editor hooks, audit trail planes |
| [audit-correlation-design.md](audit-correlation-design.md) | Three-plane correlation query + Entra subject mapping (design) |
| [teo-llm-io.md](teo-llm-io.md) | TEO LLM-I/O grammar and the native-format carve-out |
| [opa-hook.md](opa-hook.md) | Standalone OPA `PreToolUse` gate for bare editors (Phase 3b) — `tools/opa_hook.py` |
| [agent-identity/](agent-identity/README.md) | Cloud agent-identity models per OIDC provider (Entra / AWS / GCP) + the OE deployment decision |
| [agentgateway-comparison.md](agentgateway-comparison.md) | Landscape: **agentgateway** (wire/proxy) · **ActPlane** (kernel/eBPF enforcement) · our platform — dimensions by pillar; how each deepens our MCP vs native enforcement planes |
| [eval-agentgateway-mcp-plane.md](eval-agentgateway-mcp-plane.md) | Spike plan — adopt agentgateway as the MCP data plane (gate: can its external-authz delegate to our OPA bundle + carry the tri-state?) |
| [eval-actplane-native-plane.md](eval-actplane-native-plane.md) | Spike plan — eBPF/ActPlane enforcement beneath the native plane (close the indirect-execution gap `opa_delegate`/`opa-hook` have) |
| [eval-rlm-long-context.md](eval-rlm-long-context.md) | Spike plan — Recursive Language Models as a **governed** long-context capability (beat compaction; we own the guardrails the paper leaves open) |
| [eval-agentic-bpm-framed-autonomy.md](eval-agentic-bpm-framed-autonomy.md) | Spike plan — Agentic BPM / "framed autonomy" (IBM TDF·FRAME): adopt the formalism + assess Open Engine over BPMN control points |
| [eval-semantic-early-stopping.md](eval-semantic-early-stopping.md) | Spike plan (small) — embedding-convergence halting folded into `hillclimb_budget` + "keep the best round" in the Judge skill |
| [research-triage.md](research-triage.md) | Triage of inbound papers/docs — spike-worthy vs concepts-only vs skip, with reasoning grounded in our stack |
| [target-architecture-all-spikes.md](target-architecture-all-spikes.md) | Synthesis (workflow-verified) — how the 5 pillars change if all 5 spikes land, the new enforcement planes, trade-offs, and the open-items gaps |
| [tdf-frame-mapping.md](tdf-frame-mapping.md) | Adopted formalism (CLO-8) — our architecture in TDF/FRAME terms (framed autonomy); φ as a first-class execution-admission plane |
| [enterprise-pitch.md](../enterprise-pitch.md) | GTM narrative, BECU-style pilot framing, audit for compliance buyers |
| [air-vs-omnigent.md](air-vs-omnigent.md) | Prior overlap analysis (historical; consolidation supersedes runtime split) |
| [aos-vs-air.md](aos-vs-air.md) | AOS pillar mapping vs air ecosystem (reference; `aos` repo archived after merge) |

**Related (pre-consolidation):**

- [integration-plan.md](../integration-plan.md) — original cross-repo wiring plan
- [the_agent_harness_stack.md](../the_agent_harness_stack.md) — stack map (update in Phase 0 to name Omnigent as runtime)

**Cursor execution plan** (todos, full detail): `~/.cursor/plans/AIR Omnigent Consolidation-c7d8b160.plan.md`
