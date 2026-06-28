# Agentic Harness — Governed Platform Hub

The **docs & architecture hub** for the Open Engine consolidation (AIR + Omnigent as one
governed agentic platform), plus the runnable **governance spikes**, the cross-plane glue
tools, and the now **content-only `air` CLI** (pod-bundle, personas, skills, stack docs).

> **New here? Start with [GETTING_STARTED.md](GETTING_STARTED.md)** — the pillars, how they
> fit, the governance model, and where to begin.

## The platform at a glance

| Pillar | Repo | Role |
| :-- | :-- | :-- |
| Engine | [omnigent](https://github.com/Cloud-Byte-Consulting/omnigent) | runs governed agent sessions |
| Policy Gateway | [Agentic-Sentry](https://github.com/Cloud-Byte-Consulting/Agentic-Sentry) | MCP plane: OIDC + OPA on every `tools/call` |
| Token plane | [Cachy](https://github.com/Cloud-Byte-Consulting/Cachy) · [teo](https://github.com/Cloud-Byte-Consulting/teo) · [token-dashboard](https://github.com/Cloud-Byte-Consulting/token-dashboard) | optimize · format · visualize spend |
| **Hub (this repo)** | agentic-harness | architecture, spikes, glue tools, `air` content CLI |

**The rule:** the LLM proposes, **OPA decides**. Two enforcement planes (native ↔
`opa_delegate`/ActPlane, MCP ↔ agentgateway/Sentry) call one shared Rego bundle. Full map
in [GETTING_STARTED.md](GETTING_STARTED.md).

## Consolidation status (Phase 0 — [consolidation-plan.md](docs/architecture/consolidation-plan.md))

| Path / concern | Status | Moving to |
| :--- | :--- | :--- |
| `air bootstrap` / `air session` (runtime paths) | **Deprecated (Phase 1)** | Omnigent stack profile (`omnigent/profiles/openengine_stack.yaml`) |
| `copilot-role-router` orchestration | **Deprecated (Phase 2)** | Omnigent supervisor YAML |
| `aos` repo | **Pending archive (Phase 4)** | Docs merged → `docs/architecture/` (`aos-vs-air.md`, `air-vs-omnigent.md`) |

`air` shrinks to **content + docs only** (pod-bundle, personas, skills, stack docs, the
`air` CLI for content ops). All runtime concerns (session lifecycle, MCP inject, bootstrap,
policy enforcement) live in Omnigent.

## What's in this repo

- **[GETTING_STARTED.md](GETTING_STARTED.md)** — the platform entry point.
- **[`docs/architecture/`](docs/architecture/README.md)** — the governed-platform architecture, the consolidation plan, the single-owner matrix, the CLO-7…11 spike evals, the TDF/FRAME (φ) mapping, the policy/audit model.
- **[`experiments/`](experiments/)** — runnable spikes: [`cl09-real-policy`](experiments/cl09-real-policy/) (our real `mcp_auth.rego` through agentgateway, tri-state live), [`cl10-actplane`](experiments/cl10-actplane/) (eBPF native plane), [`cl11-rlm`](experiments/cl11-rlm/BRIEF.md) (RLM brief).
- **`tools/`** — [`opa_hook.py`](tools/opa_hook.py) (standalone PreToolUse OPA gate) and [`oe_correlate.py`](tools/oe_correlate.py) (three-plane audit correlator: authz + session + cost).
- **[`docs/open-engine/`](docs/open-engine/open-engine.md)** — Open Engine: Linear as the operating surface (queue, status ledger, runner loop, standing skills).
- **`air/`** — the content-only `air` CLI (Go): agents, skills, personas, package, targets, status, version. Emits Token-Efficient Output by default (`--human` to opt out).
- **`pod-bundle/`** — the vendored Pod + skill-routing toolkit (contracts, templates, 9 persona packs).
- **`harness.manifest.yaml`** — bill of materials naming every pillar repo and its boundary.

## Key documents

- [Architecture index](docs/architecture/README.md) · [Governed-platform architecture](docs/architecture/governed-platform-architecture.md) · [Consolidation plan](docs/architecture/consolidation-plan.md) · [All-spikes target architecture](docs/architecture/target-architecture-all-spikes.md)
- [Securing Agents with OPA](docs/agent_security_opa.md) · [Policy/OPA hooks audit](docs/architecture/policy-opa-hooks-audit.md) · [Audit correlation design](docs/architecture/audit-correlation-design.md)
- [The Agent Harness Stack](docs/the_agent_harness_stack.md) · [Open Engine](docs/open-engine/open-engine.md) · [Token-Efficient Output (TEO)](docs/teo_output.md)
- [Pods & Skill Routing](docs/pods_and_skill_routing.md) · [Packaging & Persona Packs](docs/packaging_and_personas.md) · [Orchestration patterns](docs/patterns/orchestration_patterns.md)

## Repository Layout

```text
.
├── GETTING_STARTED.md    # ← platform entry point (pillars, governance, where to start)
├── docs/
│   ├── architecture/     # governed-platform architecture, consolidation plan, CLO-7..11 evals
│   ├── open-engine/      # Linear-as-operating-surface
│   └── patterns/         # orchestration pattern notes
├── experiments/          # runnable spikes: cl09-real-policy, cl10-actplane, cl11-rlm
├── tools/                # opa_hook.py (OPA gate), oe_correlate.py (3-plane audit)
├── air/                  # `air` CLI (Go) — content-only: agents, skills, personas, package, status
├── pod-bundle/           # vendored Pod + skill-routing toolkit (contracts, templates, personas)
├── profiles/             # saved team init profiles
├── templates/            # instruction seeds + work-product templates
├── harness.manifest.yaml # bill of materials (every pillar repo + its boundary)
└── .gitea/workflows/     # CI (Gitea-only): builds/tests the air CLI, packages the bundle
```

## Distribution

Source of truth is **GitHub** (`Cloud-Byte-Consulting`); a **Gitea** mirror runs CI,
packages, and workflows. All CI is **Gitea-only** (GitHub Actions disabled).
