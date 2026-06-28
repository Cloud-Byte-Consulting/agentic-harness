# Getting Started — the Governed Agentic Platform

This is the entry point for the **Open Engine** consolidation: **AIR + Omnigent unified
into one governed agentic platform**, where coding agents run under enforced policy
rather than on the honor system. The guiding rule is **the LLM proposes, OPA decides** —
every consequential action is admitted or denied by a deterministic policy bundle, never
by the model itself.

This guide maps the **pillars**, shows **how they fit**, and points you at **where to
start**. Each repo's own README links back here. New to the terminology? Keep the
**[Glossary](GLOSSARY.md)** open alongside. Ready to stand it up? → **[INSTALL.md](INSTALL.md)**.

---

## The pillars

| Pillar | Repo | What it is |
| :-- | :-- | :-- |
| **Engine** | [omnigent](https://github.com/Cloud-Byte-Consulting/omnigent) | The meta-harness that actually *runs* governed agent sessions (Claude Code, Codex, Cursor, Pi, YAML agents). Native-plane OPA delegate, `inner/nessie` policies, `role_router` supervisor, `rlm_query`. |
| **Policy Gateway** | [Agentic-Sentry](https://github.com/Cloud-Byte-Consulting/Agentic-Sentry) | The MCP data plane: every `tools/call` passes OIDC auth + OPA policy (`mcp_auth.rego`) before execution. Emits the tri-state **allow / deny / require_approval**. |
| **Token / Cost plane** | [Cachy](https://github.com/Cloud-Byte-Consulting/Cachy) · [teo](https://github.com/Cloud-Byte-Consulting/teo) · [token-dashboard](https://github.com/Cloud-Byte-Consulting/token-dashboard) | Cachy = LLM context-optimization proxy (spend less); teo = token-efficient output format; token-dashboard = cost visibility across tools. |
| **Docs / Architecture hub** | **agentic-harness** (this repo) | The architecture, the consolidation plan, the governance spikes (CLO-7…11), the cross-plane glue tools, and the content-only `air` CLI. |

---

## How they fit together

```
                         ┌──────────────────────────────────────────┐
                         │  Omnigent ENGINE  (runs the agent)        │
   agent / supervisor →  │  role_router · rlm_query · sandboxes      │
                         └───────────────┬───────────────┬──────────┘
                    native tool calls    │               │   MCP tool calls
                    (Bash/Write/Edit)    │               │   (tools/call)
                                         ▼               ▼
                            ┌────────────────┐   ┌─────────────────────┐
                            │ NATIVE plane   │   │  MCP plane          │
                            │ opa_delegate   │   │  Agentic-Sentry /   │
                            │ (+ ActPlane    │   │  agentgateway       │
                            │  eBPF, CLO-10) │   │  extAuthz           │
                            └───────┬────────┘   └──────────┬──────────┘
                                    │  same shared bundle    │
                                    ▼                        ▼
                            ┌──────────────────────────────────────────┐
                            │  OPA  —  data.mcp.auth  (mcp_auth.rego)   │
                            │  → allow · deny · require_approval(428)   │
                            └──────────────────────────────────────────┘

   LLM traffic ─→ Cachy (optimize) ─→ provider        cost ─→ token-dashboard
   output ─→ teo (token-efficient format)
```

**The key idea:** both enforcement planes — native (`opa_delegate`, kernel-backed by
ActPlane) and MCP (Agentic-Sentry / agentgateway) — call **one shared Rego bundle**
(`data.mcp.auth` in `mcp_auth.rego`). One policy, two choke points. Groups (from Entra)
only *relax* enforcement, and only from verified tokens.

---

## The governance model

- **OPA decides, the LLM proposes.** Policy is deterministic Rego, not a model call.
- **Tri-state:** `allow` (proceed) · `deny` (hard stop) · `require_approval` (human ASK;
  surfaces as HTTP `428` at the gateway).
- **Two enforcement planes, one bundle:** MCP plane ↔ agentgateway/Sentry; native plane ↔
  `opa_delegate` + ActPlane (eBPF, catches indirect exec the tool-call hooks miss).
- **Fail closed.** Missing config, missing cost data, a renamed upstream hook → bounded or
  denied, never silently opened.

Deep dives: [`docs/architecture/governed-platform-architecture.md`](docs/architecture/governed-platform-architecture.md)
· [`docs/architecture/consolidation-plan.md`](docs/architecture/consolidation-plan.md)
· [`docs/agent_security_opa.md`](docs/agent_security_opa.md).

---

## The governance spikes (CLO-7…11) — accepted

Five research-grounded spikes, each a runnable experiment, now merged:

| # | Spike | Result | Where |
| :-- | :-- | :-- | :-- |
| CLO-7 | Semantic early-stopping | embedder-driven `STOP_CONVERGED` in the hill-climb/Judge loop | omnigent `inner/nessie/policies.py` · [`eval-semantic-early-stopping.md`](docs/architecture/eval-semantic-early-stopping.md) |
| CLO-8 | Agentic-BPM / TDF / **φ** | formalism adopted; φ ≈ `opa_delegate` admission | [`tdf-frame-mapping.md`](docs/architecture/tdf-frame-mapping.md) |
| CLO-9 | agentgateway MCP plane | **our real `mcp_auth.rego` enforced through extAuthz→OPA**, tri-state proven live (allow/deny/428) | [`experiments/cl09-real-policy/`](experiments/cl09-real-policy/) |
| CLO-10 | ActPlane native plane | eBPF/BPF-LSM enforcement; OE boundaries → DSL; indirect-exec catch staged | [`experiments/cl10-actplane/`](experiments/cl10-actplane/) |
| CLO-11 | RLM long-context | governed, opt-in `rlm_query` tool (fan-out cap, token-budget, sandbox) | omnigent `tools/builtins/rlm_query.py` · [`experiments/cl11-rlm/BRIEF.md`](experiments/cl11-rlm/BRIEF.md) |

Index + decision gates: [`docs/architecture/target-architecture-all-spikes.md`](docs/architecture/target-architecture-all-spikes.md).

---

## Where to start

| You want to… | Start here |
| :-- | :-- |
| **Run a governed agent** | omnigent — `profiles/openengine_stack.yaml` (the OE stack profile) + `examples/role_router/config.yaml` |
| **See policy enforced end-to-end** | [`experiments/cl09-real-policy/`](experiments/cl09-real-policy/) — `docker compose up -d && ./test.sh` |
| **Understand the architecture** | [`docs/architecture/governed-platform-architecture.md`](docs/architecture/governed-platform-architecture.md) |
| **Read/extend the policy** | Agentic-Sentry — `mcp-policies/policies/mcp_auth.rego` (the shared bundle) |
| **Gate native tool calls standalone** | [`tools/opa_hook.py`](tools/opa_hook.py) (PreToolUse OPA gate) |
| **Correlate the three planes' audit** | [`tools/oe_correlate.py`](tools/oe_correlate.py) (authz + session + cost) |
| **Track token spend** | token-dashboard (`npm run build-data && npm run dev`) · Cachy (proxy) |
| **Run work as tracked tasks** | [`docs/open-engine/open-engine.md`](docs/open-engine/open-engine.md) (Linear-as-operating-surface) |

---

## Repo map

| Repo | Pillar | README |
| :-- | :-- | :-- |
| **agentic-harness** | Docs/architecture hub + `air` content CLI | [README](README.md) |
| omnigent | Engine (runtime) | [README](https://github.com/Cloud-Byte-Consulting/omnigent) |
| Agentic-Sentry | MCP policy gateway | [README](https://github.com/Cloud-Byte-Consulting/Agentic-Sentry) |
| Cachy | Token plane — optimization proxy | [README](https://github.com/Cloud-Byte-Consulting/Cachy) |
| teo | Token plane — output format | [README](https://github.com/Cloud-Byte-Consulting/teo) |
| token-dashboard | Token plane — cost visibility | [README](https://github.com/Cloud-Byte-Consulting/token-dashboard) |

> Source of truth is **GitHub** (`Cloud-Byte-Consulting`); a **Gitea** mirror runs CI,
> packages, and workflows. All CI is Gitea-only.
