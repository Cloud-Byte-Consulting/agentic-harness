# Glossary

Every acronym and platform-specific term used across the docs (the [getting-started
hub](GETTING_STARTED.md), the [enterprise pitch](docs/enterprise-pitch.md), and the
[architecture docs](docs/architecture/README.md)). One line each.

## Platform & repos

| Term | Meaning |
| :-- | :-- |
| **Open Engine** | The name of the consolidation: AIR + Omnigent unified into one governed agentic platform. |
| **AIR** | The original harness/CLI (the `air` command); now **content-only** (agents, skills, personas, docs) — runtime moved to Omnigent. |
| **Omnigent** | The agent **engine** — the meta-harness that runs governed sessions across many coding agents. |
| **Agentic-Sentry** | The **MCP policy gateway** — gates every tool call through auth + policy before execution. |
| **Cachy** | LLM context-**optimization** proxy (the token-plane proxy that reduces spend). |
| **teo / TEO** | **Token-Efficient Output** — a dense, line-oriented output format (Go library + CLI); the `air` CLI emits it by default. |
| **token-dashboard** | Cost-**visibility** dashboard tracking token burn in exact / activity / estimate lanes. |
| **ARD** | **Agentic Resource Discovery** — a federated catalog of MCP tools, agents, and skills. |
| **agentic-harness** | This repo — the docs / architecture **hub** + the content `air` CLI. |

## Governance & policy

| Term | Meaning |
| :-- | :-- |
| **OPA** | **Open Policy Agent** — the deterministic policy engine that makes every allow/deny decision. |
| **Rego** | OPA's policy language; our policy lives in `mcp_auth.rego` (versioned, `opa test`-gated). |
| **MCP** | **Model Context Protocol** — the standard by which agents call external tools (`tools/call`). |
| **MCP plane** | One of two enforcement planes — MCP tool calls gated by Agentic-Sentry / agentgateway. |
| **native plane** | The other enforcement plane — native tool calls (Bash/Write/Edit) gated by `opa_delegate`, kernel-backed by ActPlane. |
| **agentgateway** | A Rust MCP data-plane proxy; its `extAuthz` delegates allow/deny to OPA. |
| **extAuthz** | Envoy **external authorization** — the gRPC protocol agentgateway speaks to OPA (no shim). |
| **opa_delegate / opa-hook** | The native-plane PreToolUse gate that POSTs a tool call to the **same** OPA bundle the MCP plane uses. |
| **ActPlane** | eBPF / BPF-LSM enforcement *beneath* the native plane — catches **indirect exec** (a `git`/`curl`/`rm` reached via a model-written script) that tool-call hooks miss. (CLO-10) |
| **eBPF** | extended Berkeley Packet Filter — safe programmable hooks in the Linux kernel. |
| **BPF-LSM** | BPF **Linux Security Module** — the kernel hook ActPlane uses to enforce; requires a BPF-LSM-capable host. |
| **tri-state** | The verdict shape: **allow / deny / require_approval** (vs a plain allow/deny boolean). |
| **require_approval** | The "ask a human first" verdict; surfaces as HTTP **428** at the gateway and an **ASK** in the engine. |
| **ASK** | Human-in-the-loop approval (elicitation) — the engine pauses for a person to approve/deny. |
| **fail closed** | On any error/unreachable policy/missing config → **deny or bound**, never silently allow. |
| **φ (phi)** | **Execution-admission**: what the engine may *act on* after the agent reasons — an OPA decision (φ), not an LLM. A first-class artifact. (CLO-8) |
| **TDF / FRAME** | The Agentic-BPM formalism (Task-Decomposition / the aggregate policy set) we adopt the φ vocabulary from. (CLO-8) |
| **nessie** | The Omnigent policy module (`inner/nessie/policies.py`) holding the governance policies. |
| **hillclimb_budget** | The nessie policy bounding the refine/Judge rework loop (rounds, plateau, convergence). |
| **spawn_bounds / cost_plan** | nessie policies that cap sub-agent/RLM fan-out and token-cost exposure. |
| **STOP_CONVERGED** | The semantic early-stopping terminal — stop refining when outputs stop changing in meaning. (CLO-7) |
| **Judge** | The deterministic two-tier verdict gate (heuristic floor + independent cross-vendor LLM review). |

## Identity & audit

| Term | Meaning |
| :-- | :-- |
| **OIDC** | **OpenID Connect** — the identity/auth protocol the gateway uses. |
| **Entra** | **Microsoft Entra ID** — the identity provider; security **groups** bind tool access (groups only *relax* policy). |
| **SIEM** | **Security Information and Event Management** — where enterprise audit logs are forwarded/retained. |
| **OTel / OTLP** | **OpenTelemetry** (Protocol) — the telemetry standard for exporting session/decision events. |
| **three planes** | The audit fabric: **authz** (Sentry+OPA) · **runtime** (Omnigent session) · **artifact** (teo + cost), correlated by `request_id`/`session_id`/`subject_id`. |

## Economics

| Term | Meaning |
| :-- | :-- |
| **CapEx / OpEx** | Capital vs operating expenditure. The thesis: the governed harness is the **CapEx** that bends agentic **OpEx** (rework, incidents, audit, spend) down. |
| **FinOps** | Cloud **financial operations** — the cost-accountability discipline / buyer persona. |
| **SDLC** | **Software Development Life Cycle** — Google's "New SDLC with Vibe Coding" framing. |
| **Agent = Model + Harness** | Google's New-SDLC equation: the model is ~10% of a working agent; the harness is the other 90%. |
| **vibe coding** | Prompt-and-accept — low CapEx, high/rising OpEx. Contrast: **agentic engineering** (AI inside enforced constraints). |
| **ROI / CFO** | Return on investment / Chief Financial Officer — the economics buyer. |

## The governance spikes (CLO-7…11)

| Term | Meaning |
| :-- | :-- |
| **CLO-N** | Issue IDs in our Linear tracker (the work surface). **CLO-7…11** are the five governance spikes. |
| **CLO-7** | Semantic early-stopping (`STOP_CONVERGED`). |
| **CLO-8** | Agentic-BPM / TDF / FRAME formalism (φ as first-class). |
| **CLO-9** | agentgateway MCP plane — our real `mcp_auth.rego` enforced live (200/403/428). |
| **CLO-10** | ActPlane eBPF native-plane enforcement. |
| **CLO-11** | **RLM** long-context — a governed, opt-in tool. |
| **RLM** | **Recursive Language Models** — load a large corpus as a REPL variable and recurse over it for long-context tasks. (CLO-11) |

## Infrastructure

| Term | Meaning |
| :-- | :-- |
| **K8s** | Kubernetes — the deployment target profile. |
| **CI** | Continuous Integration — runs on **Gitea** only (GitHub Actions disabled). |
| **GitHub / Gitea** | GitHub = source of truth (`Cloud-Byte-Consulting`); Gitea = the downstream mirror running CI, packages, workflows. |

## Content concepts

| Term | Meaning |
| :-- | :-- |
| **pod** | A markdown-driven agent persona/role bundle (5-file anatomy). |
| **persona** | A non-SWE role pack (TPM, PM, EM, SRE, …) built on the pod methodology. |
| **skill** | A reusable, markdown-defined capability shared across tools via `air skills link`/`sync`. |
