# air vs Omnigent

> **Note:** Merged from `aos/docs/` (Phase 0). **Consolidation decision:** Omnigent owns runtime; air owns content/contracts at boundaries. See [consolidation-plan.md](consolidation-plan.md).

[Databricks Omnigent announcement](https://www.databricks.com/blog/introducing-omnigent-meta-harness-combine-control-and-share-your-agents)

Omnigent appears stronger and more productized around **live agent session ownership**. The article describes a meta-harness that wraps existing terminal agents, SDK agents, and custom harnesses behind a common API, then exposes those sessions through terminal, app, web API, web/mobile, and native app interfaces. It also emphasizes session collaboration, shared URLs, cloud execution, strong sandboxing, and network request interception.

air appears broader and more federated around **protocols, contracts, and independently shippable layers**. The local repos describe MCP policy gateway enforcement through Agentic-Sentry, ARD discovery catalogs for MCP/A2A/skills/resources, Cachy as a token-plane proxy and optimization foundation, TEO as a machine-readable output contract, token-dashboard as local cost analytics, and agentic-harness as the manifest/BOM and bootstrap hub.

The simplest distinction: **Omnigent owns live agent sessions; air owns contracts and glue between separately shippable layers**. Omnigent validates the meta-harness layer air is circling, but the two projects emphasize different centers of gravity. Omnigent is closer to a cohesive product surface; air is closer to a federated operating-system style stack.

## Mental model

```text
Omnigent
  user/app/API
      |
  live session wrapper
      |
  Claude Code / Codex / Pi / SDK agents / custom harnesses
      |
  cloud runner + sandbox + request controls

air
  agentic-harness manifest + bootstrap
      |
  role-router | Sentry | ARD | Cachy | TEO | token-dashboard
      |
  Cursor / Claude Code / Codex / Copilot / Gemini / local services
```

| System | Primary ownership | Shape |
|---|---|---|
| Omnigent | Live agent sessions, common API, sharing, app surfaces, sandbox execution | Productized meta-harness above existing agents |
| air | Contracts, policy, discovery, token optimization, output format, cost visibility, role protocol | Federated composition layer across independently versioned repos |

## Overlap

| Capability | Omnigent | air | Notes / maturity |
|---|---|---|---|
| Meta-harness above CLIs | The article describes wrapping terminal-based coding agents and SDK agents behind a common layer. | `agentic-harness` targets multiple harnesses and `copilot-role-router` runs across Cursor, Claude Code, Copilot, Gemini, and similar hosts. | Strong conceptual overlap. Omnigent appears more unified at runtime; air is more manifest/plugin driven. |
| Composition | Composes across models, harnesses, and techniques. | Manifest/BOM composes Sentry, ARD, Cachy, token-dashboard, role-router, TEO, pods, and personas. | air composition is explicit and repo-federated; install providers are still partly plan-level. |
| Multi-harness authoring | The article describes YAML authoring and one-line harness changes. | `harness.targets.yaml`, skills, personas, pods, and role-router protocol provide cross-harness authoring concepts. | Omnigent appears stronger on portability of a single session definition; air has broader content portability. |
| Policies | Stateful contextual policies, cost policies, and permissions are article claims. | Agentic-Sentry enforces MCP `tools/call` authorization via OPA/RBAC; role-router adds mutation policy by role. | air has concrete OPA/MCP policy contracts, but less session-level dynamic policy. |
| Cost awareness | Article claims cost policies. | token-dashboard measures local token/activity/estimate lanes; Cachy has token counting and planned/request-path savings telemetry. | air has stronger local analytics detail; Omnigent appears stronger on inline session budgets. |
| Sandbox/security | Article claims strong OS sandboxing and network request interception/transform. | Sentry gates tool calls; Cachy has a bounded WASM plugin sandbox; no full agent compute sandbox is shipped. | Omnigent appears ahead on execution sandbox. air has stronger tool-policy architecture. |
| Agent sharing/API | Shared live sessions via URL, review/comment/steer, common API. | air has contracts and CLIs, but no shared live session URL or session server. | Clear Omnigent advantage. |
| Tool integration | Common API over terminal agents and SDKs; network interception. | MCP gateway, ARD catalog entries, MCP config injection, skills, endpoint recipes. | air has a rich tool/discovery story; Omnigent appears richer in wrapping active sessions. |
| Role/subagent orchestration | Article claims composition across harnesses and techniques, but does not describe a role protocol in the supplied claims. | role-router defines CO, Recon, Medic, Engineer, QA, Scribe, Judge, read-only rules, Judge gate, docs gate, and hill-climb budget. | air has a clearer governance protocol from local docs. |

## What air is missing from Omnigent

| Missing capability | Why it matters | Omnigent has | air current state | Suggested air implementation |
|---|---|---|---|---|
| Unified session API wrapping CLIs/SDK agents | Gives callers one lifecycle model for start, resume, inspect, stop, tool events, token events, and artifacts. | The article describes a common API over terminal coding agents and SDK agents. | air configures and governs external harnesses but does not own a unified session object. | Build a minimal `air session` API around Claude Code and Codex first: `start`, `attach`, `events`, `status`, `stop`, `export`. |
| Live session sharing/collaboration URL | Lets multiple humans review, comment, and steer the same run without screen sharing or transcript copying. | The article describes sharing live sessions via URL and collaborative review/comment/steering. | No shared live session URL. Collaboration happens through host IDE/chat surfaces and docs. | Add read-only local session viewer first, then signed share links once a server exists. |
| Session server with web/mobile/native/app/API access | Makes the harness product usable outside one terminal or editor. | The article claims terminal, app, web APIs, web/mobile, Mac native app access. | air has CLIs, docs, services, and dashboards, but no unified session server or app shell. | Ship a local HTTP session server before building mobile/native UI. |
| Cloud execution/session runner abstraction | Moves sessions off the user's workstation and enables repeatability, scaling, and isolation. | The article mentions cloud execution and sandbox providers such as Modal and Daytona. | token-dashboard and Sentry have deploy paths; air has no session runner abstraction. | Define `air runner` with local process as v0, then provider adapters for Daytona/Modal-like backends. |
| Strong OS sandbox with network interception | Controls process, filesystem, network, and outbound requests at the runtime boundary. | The article claims strong OS sandbox and network request interception/transform. | Sentry authorizes MCP tool calls; Cachy proxies LLM HTTP; neither is a full OS sandbox for agent execution. | Start with local process sandbox policy and network egress logging; defer stronger isolation to one runner backend. |
| Dynamic contextual policies at session level | Policies need access to session state, task type, user, budget, workspace risk, and prior approvals. | The article describes stateful contextual policies. | OPA/RBAC policies are concrete at the MCP call boundary; role-router has static role rules and gates. | Add session context as policy input: role, branch, task intent, budget state, approval state, and resource identity. |
| Cost budgets that pause/ask before continuing | Prevents runaway spend and turns cost into an explicit approval boundary. | The article claims cost policies. | token-dashboard measures usage; Cachy counts tokens; no unified budget hook pauses a session. | Add budget thresholds to `air session`; feed exact logs from token-dashboard and live counters from Cachy where available. |
| Agent YAML/session authoring portability comparable to Omnigent | Lets the same workload move between harnesses with minimal edits. | The article describes YAML authoring and one-line harness changes. | `harness.manifest.yaml`, targets, personas, pods, and skills exist; not a portable session spec. | Add `air.session.yaml` for harness, model, tools, policies, budget, runner, and output contracts. |
| Omnigent Server MCP / cross-session MCP concept | Makes sessions and session state callable/discoverable by other agents and tools. | Roadmap includes Omnigent Server MCP. | ARD can catalog MCP/A2A/skills/resources; Sentry is an MCP gateway; no cross-session MCP server. | Expose session inspection/control as a local MCP server, registered through ARD and gated by Sentry. |
| Productized UX/cohesive app | Reduces setup friction and makes the stack understandable as one product. | The article describes multiple product interfaces including app and native surfaces. | air is a workspace of repos plus CLI, dashboards, docs, and manifests. | Keep the CLI thin, but provide one bootstrap command and one local status/session UI. |

## What air has over Omnigent

| air advantage | Why it matters | Component/repo | Omnigent coverage from article | Maturity |
|---|---|---|---|---|
| ARD federated discovery catalog across MCP/A2A/skills/resources | Lets resources advertise themselves through a domain-anchored catalog instead of one central registry. | `Agentic-Resource-Discovery` | The article claims common APIs and more harnesses, but not a federated discovery standard in the supplied claims. | Partial: draft spec and conformance tooling; no full reference registry in the reviewed docs. |
| Agentic-Sentry as MCP policy gateway with OPA/RBAC | Gives tool calls a deterministic authorization chokepoint. | `Agentic-Sentry` | Omnigent claims permissions and contextual policies, but the article claims do not specify OPA/RBAC/MCP gateway mechanics. | Partial/Built: gateway and OPA sidecar exist; backend proxy status varies by doc, with v2 forwarding still described as planned in README. |
| Cachy token-plane API-key proxy, compression/CCR architecture | Gives air a route to optimize and observe LLM traffic when clients support configurable base URLs. | `Cachy` | The article claims network interception/transform, but not a token-compression/CCR architecture in the supplied claims. | Partial: transparent proxy, token counting, compression foundations, CCR store, integrations; live path is pass-through only. |
| TEO machine-readable output contract | Reduces output bloat and gives agents a strict parse/validate contract. | `teo` | No comparable output grammar is described in the supplied Omnigent claims. | Built: library, converter, CLI, tests, and `air` consumer path. |
| token-dashboard local token/cost analytics | Keeps exact, activity, and estimate lanes honest across many tools. | `token-dashboard` | Omnigent claims cost policies, but not this local multi-source analytics model in the supplied claims. | Built/Partial: local dashboard and data pipeline; sibling ingestion seams are still missing. |
| Manifest/BOM-driven install/composition via agentic-harness | Keeps components independently versioned while providing one stack view. | `agentic-harness` | Omnigent has multi-harness authoring, but the supplied claims do not describe a federated BOM across repos. | Partial: manifest exists; install providers are still partly plan-oriented. |
| Role-router Judge gate/hill-climb protocol | Makes review, rework budget, and acceptance explicit instead of self-certified. | `copilot-role-router` | The supplied Omnigent claims do not mention a comparable Judge gate/hill-climb protocol. | Built/Partial: protocol and CLI are documented; enforcement depends on host harness integration. |
| Plugin/skills/persona/pod portability across harnesses | Lets operating practices travel across Cursor, Claude, Copilot, Gemini, and related tools. | `agentic-harness`, `copilot-role-router` | Omnigent claims harness composition and YAML portability. | Partial: content portability is strong; runtime portability is less unified than Omnigent's described session wrapper. |
| Explicit AOS layer model: kernel, memory, drivers, sandbox, IPC, LUI | Gives air a vocabulary for what belongs in each layer and what is still missing. | `aos` docs plus workspace repos | Omnigent is framed as a meta-harness product, not as an AOS layer model in the supplied claims. | Conceptual: `aos` repo is a placeholder; mapping docs exist. |

## Plane-by-plane comparison

| Plane | Omnigent | air | Notes |
|---|---|---|---|
| Session plane | Built | Missing/Partial | Omnigent's article centers live sessions and common API. air has host sessions but no owned `air session` abstraction. |
| Orchestration/kernel plane | Partial/Built | Partial | Omnigent composes harnesses/models/techniques. air has role-router governance but no native reasoning kernel. |
| Tool plane | Built | Partial/Built | Omnigent wraps agents and requests. air has MCP config injection, Sentry authorization, ARD catalog concepts, and skills. |
| Token plane | Partial | Partial | Omnigent has cost policies in the article claims. air has Cachy/token-dashboard foundations, but live optimization is not fully wired. |
| Policy plane | Built | Partial/Built | Omnigent claims stateful contextual policies. air has concrete OPA/RBAC MCP policy plus role mutation rules, but less session context. |
| Sandbox/compute plane | Built | Missing/Partial | Omnigent claims OS sandbox, network interception, and cloud providers. air has policy boundaries and WASM plugin sandbox, not full agent compute. |
| Discovery plane | Partial/Missing | Partial | Omnigent roadmap mentions more harnesses and server MCP, but no federated discovery claim was supplied. air has ARD spec/conformance. |
| Output/contract plane | Missing/Unknown | Built | TEO is a concrete air output grammar; no equivalent is described for Omnigent in the supplied claims. |
| Observability/cost plane | Partial/Built | Partial/Built | Omnigent claims cost policies. air has local analytics and token counting, but inline budget enforcement is missing. |

## Strategic implications for air

1. Build a minimal `air session` API/adapter around Claude Code/Codex first. Do not start with every harness; prove lifecycle, events, artifacts, and cancellation on two concrete adapters.
2. Unify Cachy base-url setup and Sentry MCP config in one bootstrap flow. Users should not need separate mental models for token-plane and tool-plane wiring.
3. Add session policy and cost budget enforcement hooks using token-dashboard, Sentry, and role-router. Start with simple thresholds that pause and ask before continuing.
4. Add a local session server/API before collaboration UI. A stable event and control API makes terminal, dashboard, and future app surfaces cheaper.
5. Pick one sandbox runner abstraction. Start with a local process runner, then add one cloud/sandbox provider adapter after the contract survives real use.
6. Keep ARD, TEO, Cachy, and Sentry as differentiators. Do not collapse them into a monolith; compose them through manifest, discovery, MCP, and session contracts.
7. Treat Omnigent as validation of the meta-harness layer, not a reason to abandon AOS framing. The AOS model helps air keep kernel, memory, drivers, sandbox, IPC, LUI, and observability distinct.

## Build/buy/borrow view

**Copy conceptually:** the live session object, common API, shareable session URL, session server, harness adapter boundary, and explicit cost-budget pause points.

**Avoid copying blindly:** a big app-first rewrite, a custom cloud runner before local sessions work, or a monolithic control plane that erases the independent repo contracts air already has.

**Differentiate:** ARD as the discovery layer, Sentry as the MCP/OPA policy gateway, Cachy as the token-plane proxy and optimization layer, TEO as the output contract, token-dashboard as honest local analytics, and role-router as the governance protocol.

## Open questions

- What is the smallest useful `air session` contract: lifecycle only, or lifecycle plus events, tool calls, token usage, artifacts, and approvals?
- Should `air session` own subprocess execution, or should it first attach to existing host sessions and logs?
- Which two harness adapters should define v0 portability: Claude Code and Codex, or Claude Code and Cursor?
- What policy state should be passed to Sentry/OPA at session time: role, task intent, workspace, branch, budget, user identity, approval status, or all of these?
- Should Cachy be mandatory for API-key mode bootstrap, or opt-in until live compression is wired into the proxy path?
- Is ARD the right registry for active sessions, or should sessions expose MCP first and register into ARD second?
- What collaboration primitive comes first: read-only session URL, comment stream, steering controls, or approval gates?
- Where should the cohesive UX live: `agentic-harness`, `aos`, token-dashboard, or a new thin session UI?
