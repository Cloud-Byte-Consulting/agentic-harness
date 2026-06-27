# Agentic Harness Primer

A starter primer for building, running, and evaluating agentic harnesses.

This repository is a practical reference for agent workflows: how agents receive tasks, use tools, manage state, evaluate outcomes, and expose reliable execution patterns for developers.

## Deprecation notices (Phase 0 — consolidation-plan.md)

| Path / concern | Status | Moving to |
| :--- | :--- | :--- |
| `air bootstrap` / `air session` (runtime paths) | **Deprecated (Phase 1)** | Omnigent stack profile (`omnigent/profiles/openengine_stack.yaml`) |
| `copilot-role-router` orchestration | **Deprecated (Phase 2)** | Omnigent supervisor YAML |
| `aos` repo | **Pending archive (Phase 4)** | Docs merged → `docs/architecture/` (see `aos-vs-air.md`, `air-vs-omnigent.md`) |

`air` shrinks to **content + docs only** (pod-bundle, personas, skills, stack docs, `air` CLI for content operations). All runtime concerns (session lifecycle, MCP inject, bootstrap, policy enforcement) move to Omnigent. See [consolidation-plan.md](docs/architecture/consolidation-plan.md).

---

## Key Resources

- **[Architecture — consolidation](docs/architecture/README.md)**: Single-owner stack (Omnigent runtime + Sentry/OPA + satellites), phased plan, policy/audit model, enterprise pitch.
- **[Open Engine](docs/open-engine/open-engine.md)**: A shared operating surface for AI agents (Linear queue integration, status ledger, runner loop, templates, and standing skills).
- **[Agentic Workflows Primer](docs/agentic_workflows_primer.md)**: A comprehensive guide covering assistant vs. agent model, harness architectures, tool sandbox policies, and instruction fanning.
- **[Agent Ownership Playbook](docs/agent_ownership_playbook.md)**: An operational playbook on how to document, own, and care for team agents, featuring templates and self-documenting prompts.
- **[The Agent Harness Stack](docs/the_agent_harness_stack.md)**: A map of the workspace projects as one stack — Discovery → Orchestration → Optimization → Observability → Output — and how they fit together.
- **[Pods & Skill Routing](docs/pods_and_skill_routing.md)**: The markdown-driven Pod methodology (5-file anatomy), the skill-routing contract, two-tier activation, and mutation-gate tiers. Toolkit vendored under [`pod-bundle/`](pod-bundle/).
- **[Cross-Tool Skills](docs/cross_tool_skills.md)**: One canonical `pod-bundle/skills/` folder, shared across Claude / Cursor / Copilot / Gemini via `air skills link` (symlinks) or `air skills sync` (copies) — edit a skill once; plus the skill-routing contract v2 update (personas + planned skills).
- **[Token-Efficient Output (TEO)](docs/teo_output.md)**: `air` emits a dense, parseable format **by default** (`--human` to opt out); backed by an emitter/parser/validator and round-trip + per-command conformance tests.
- **[Packaging & Persona Packs](docs/packaging_and_personas.md)**: How to package the polyglot, independently-versioned projects as one harness (a BOM manifest + `air` installer), and how to extend Pods into persona packs for non-SWE roles (TPM, Lead PM, EM, SRE, DevOps, Business Analyst, Data Engineer, Cybersecurity).
- **[Team Structures & Operating Models](docs/team_structures_and_operating_models.md)**: The init-time configuration for how a team is shaped in an agentic world — economics (use/compose/build), pod & hourglass shapes, operating models A/B/C, and governance-as-infrastructure. Cordillera/TRAPI become configured team profiles, not built-ins.
- **[Orchestration Patterns Design Reference](docs/patterns/orchestration_patterns.md)**: A comparative guide of core orchestration patterns.
- **[Hill-Climb Refinement Loop](docs/patterns/hill_climb.md)**: Sequential self-refinement with limits and budgets.
- **[Fan-Out / Best-of-N](docs/patterns/fan_out.md)**: Concurrent generation and evaluation patterns.
- **[Adversary / Red-Team](docs/patterns/adversary.md)**: Multi-agent debate loops for safety.
- **[Tree of Thoughts (ToT)](docs/patterns/tree_of_thoughts.md)**: Branching search with backtracking and evaluations.
- **[Self-Hosted Docker MCP Gateway](docs/docker_mcp_gateway.md)**: Run MCP servers as Docker containers behind one gateway on native Linux — install, profiles, a persistent systemd service, and per-editor JSON config.
- **[Tracing & Evaluation](docs/tracing_and_evaluation.md)**: A guide covering execution traces, spans, and standard LLM-as-judge / unit tests.
- **[Evaluation & Auditability](docs/evaluation_and_auditability.md)**: A deep dive into benchmarking rigor, uncertainty estimation, normalization, and evaluation security.
- **[Securing Agents with OPA](docs/agent_security_opa.md)**: A guide covering how to implement Open Policy Agent (OPA) to enforce deterministic, read-only boundaries on MCP tools.
- **[ADR-0001: ARD Private Hosting](docs/adr-0001-ard-private-hosting.md)**: Architecture Decision Record detailing our serverless, scale-to-zero private cloud hosting architecture for Agentic Resource Discovery.

## Repository Layout

```text
.
├── docs/                 # Primers, playbooks, and pattern references
│   └── patterns/         # Orchestration pattern notes
├── air/                  # `air` CLI (Go + Cobra/Viper): one tool for agents, skills, personas, package, install
├── pod-bundle/           # Vendored Pod + skill-routing toolkit (contracts, templates)
│   ├── skills/           # Canonical skills (single source); tool views via `air skills link|sync` (git-ignored)
│   └── personas/         # 9 persona packs (pod files + persona.yaml), generated by `air personas scaffold`
├── profiles/             # Saved team init profiles (cordillera, trapi)
├── harness.manifest.yaml # Bill of materials: components, versions, install kinds
├── harness.init.yaml     # Default team init profile (economics/talent/structure/governance)
├── harness.targets.yaml  # Named harness sets for `air … --harness <name>` (e.g. frontend = claude+copilot)
├── examples/             # Runnable example implementations
├── templates/            # Instruction seeds + work-product templates (ADR, PRD, email, commit, PR, issues) — see templates/README.md
├── tests/                # Test coverage for examples and utilities
├── .gitea/workflows/     # CI: builds + tests the air CLI, packages the release bundle
├── README.md
└── SKILL.md
```

## Purpose

Use this project as a template for documenting and iterating on an agentic harness. The initial goal is to make the core concepts easy to understand before implementation details settle.

## What Is an Agentic Harness?

An agentic harness is the runtime and support structure around an AI agent. It usually coordinates:

- Task intake and goal tracking
- Prompt and instruction assembly
- Tool access and permission boundaries
- Workspace and file-system context
- Memory or state management
- Execution loops and stopping criteria
- Evaluation, logging, and replay
- Human approval or intervention points

## Repository Goals

- Provide a concise primer for agentic harness concepts
- Capture architectural decisions as the project evolves
- Establish a consistent vocabulary for agents, tools, state, and evaluation
- Offer runnable examples once implementation begins
- Keep safety, observability, and reproducibility visible from the start

## Quick Start

Implementation details are still pending. For now, use this section as the future landing spot for setup commands.

```bash
# Example placeholder
git clone <repo-url>
cd agentic-harness
air targets list                          # harnesses + named sets (all, popular, frontend, …)
air agents link --harness claude,copilot  # target several harnesses at once
air skills sync  --harness claude,copilot # same vocabulary for skills
air bootstrap    --harness frontend       # wire up agents + skills for a saved set, in one step
air status                                # Token-Efficient Output by default (agent-friendly); add --human for text
```

## Core Concepts

### Agent

The model-driven worker that receives instructions, reasons over context, invokes tools, and produces outputs.

### Harness

The orchestration layer that gives the agent a controlled environment, manages lifecycle events, and records what happened.

### Tools

External capabilities exposed to the agent, such as shell commands, file editing, search, browser automation, APIs, or domain-specific services.

### State

The durable or ephemeral information used across steps, including task history, workspace changes, tool results, memory, and evaluation traces.

### Policy

The rules that define what the agent may do, when approval is required, and how risky actions are constrained.

## Suggested Architecture

```text
User Request
    |
    v
Task Controller
    |
    v
Agent Runtime <----> Context Builder
    |
    v
Tool Router ----> Tools / APIs / Workspace
    |
    v
Observation Log
    |
    v
Evaluator / Reviewer
```

## Evaluation Areas

Use these categories to judge harness quality:

- Correctness: did the agent solve the requested task?
- Reliability: does the behavior reproduce across runs?
- Safety: are permissions, approvals, and destructive actions controlled?
- Observability: can a developer inspect decisions, tool calls, and outputs?
- Recoverability: can interrupted or failed runs resume cleanly?
- Cost and latency: are execution time and model/tool usage measurable?

## Documentation Roadmap

- [ ] Define the first runnable harness flow
- [ ] Document tool registration and permissions
- [ ] Add examples for common agent tasks
- [ ] Add evaluation fixtures and scoring guidance
- [ ] Capture architecture decisions in ADRs
- [ ] Document operational concerns such as logging, replay, and secrets

## Contributing

Keep documentation concrete and implementation-aware. Prefer examples, diagrams, and repeatable commands over abstract descriptions.

When adding code, include enough context for a new contributor to understand:

- What problem the component solves
- How it fits into the harness lifecycle
- How to run or test it locally
- What assumptions or limitations are known

## License

License information has not been added yet.
