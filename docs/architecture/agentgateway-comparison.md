# Landscape: agentgateway · ActPlane · the AIR / Omnigent platform

**Status:** Analysis (2026-06-27). Sources — **agentgateway:** `agentgateway.dev`,
the `agentgateway/agentgateway` GitHub README, the docs (verify-checked).
**ActPlane:** the paper *"ActPlane: Programmable OS-Level Policy Enforcement for
Agent Harnesses"* (arXiv 2606.25189v1, Jun 2026; UC Santa Cruz / Virginia Tech /
HKUST / eunomia-bpf / Alibaba), open-source at `github.com/eunomia-bpf/ActPlane`.
**Ours:** this folder + the repos. Ambiguous third-party claims are flagged
**(unconfirmed)**.

This doc maps **every dimension to a pillar**, says what each is, places all three
products, and gives the **overlap-vs-supplement** read.

---

## 0. The load-bearing distinction — these are three *layers*, not three competitors

| | What it is | Where it sits |
| :-- | :-- | :-- |
| **agentgateway** | a **connectivity data plane** — a Rust HTTP/gRPC proxy for AI-native protocols (HTTP/gRPC/LLM/MCP/A2A) | **the wire** — between agents, models, MCP servers, backends |
| **ActPlane** | a **kernel enforcement substrate** — eBPF programs that enforce agent policy at the OS/syscall level | **the kernel** — below the tool API, on every execution path |
| **ours (AIR/Omnigent)** | an **end-to-end governed-agent platform** — Open Engine workflow + Omnigent runtime + OPA-as-authority + 3-plane audit | **the platform** — the whole loop |

> agentgateway is **horizontal connectivity**; ActPlane is a **deep enforcement
> primitive**; ours is a **vertical governed runtime + workflow**. They don't
> occupy the same box.

### The key insight for us: our two enforcement planes each have a deeper external option

Our platform enforces in **two planes**, and each maps to one of these projects:

- **MCP plane** — Agentic-Sentry gateways `tools/call` → OPA.
  **↔ agentgateway** is a broader proxy for that exact seam.
- **Native plane** — `opa_delegate` (Omnigent PreToolUse) + `opa-hook` (editor
  PreToolUse) gate host tools (Bash/Edit) → OPA.
  **↔ ActPlane** is a *deeper* mechanism for that exact seam — and it **exposes a
  real weakness in ours**: our native plane is **tool-call interception**, which
  ActPlane's paper empirically shows **misses indirect execution paths** (a
  `git commit` inside a script the agent wrote bypasses a PreToolUse hook;
  tool-regex enforcement ≈ our opa-hook scores **near-zero** on those "script /
  hidden" traces, while ActPlane's kernel IFC catches them).

So the honest framing is: **agentgateway can deepen our MCP plane; ActPlane can
deepen our native plane; neither touches the governance/lifecycle/audit that *is*
the platform.**

---

## 1. Pillar map (every dimension, all three)

| Pillar | Dimension | agentgateway | ActPlane | ours |
| :-- | :-- | :-- | :-- | :-- |
| **A. Connectivity / data plane** | MCP gateway / proxy | ✅ broad (federation, 4 transports, OpenAPI→MCP) | ❌ | ✅ narrower (Sentry) |
| | A2A (agent-to-agent) | ✅ | ❌ | ❌ |
| | LLM / inference routing | ✅ 12+ providers, GPU pools | ❌ | ◐ Cachy/measure |
| **B. Authorization & policy** | Policy engine | CEL *(homepage "OPA" — unconfirmed)* | **DSL** (source→target + IFC), agent-authored | **OPA/Rego**, sole authority |
| | Authn / identity | ✅ JWT/mTLS/ext-authz | ❌ (binds to **process subtree**, not caller id) | ✅ Entra OIDC + group RBAC |
| | Verdict spectrum | allow/deny + guardrails | **notify / block / kill** + semantic feedback | deny / **require_approval** / allow |
| | Native host-tool authz | ❌ (on-wire only) | ✅✅ **kernel-level, all paths** | ✅ tool-call hooks |
| | **Indirect-path coverage** | ◐ wire only | ✅ **syscall-level (covers shell-outs/subprocesses)** | ❌ **tool-call interception misses it** |
| | **Cross-event / info-flow (IFC)** | ❌ | ✅ **labels propagate across events** | ❌ |
| **C. Governance & lifecycle** | Human-in-the-loop / elicitation | ❌ | ❌ (feedback to the *agent*, not a human) | ✅ ASK → human hold → resume |
| | Agent runtime / lifecycle | ❌ | ❌ (harness-agnostic enforcer) | ✅ Omnigent sessions/sandbox |
| | Tracker → session (Open Engine) | ❌ | ❌ | ✅ |
| | Higher-authority / privilege model | external authz (scope unconfirmed) | ✅ **hierarchical domains; agent can only narrow, never weaken inherited** | ✅ Entra-group **admin carve-out** (relax-only) |
| **D. Observability & audit** | Telemetry | ✅ OTEL-native | ◐ violation logs + semantic feedback | ✅ JSONL → collectors |
| | Cross-plane correlation | ◐ proxy-scoped | ❌ (enforces, not audit-graph) | ✅ 3-plane join, spoof-resistant |
| **E. Platform & ecosystem** | Deployment posture | ✅ binary/Docker/**K8s** (Helm, CRDs) | **eBPF on Linux** (BPF-LSM), ~2–8% overhead | ◐ Go gateway + OPA sidecar |
| | Maturity / backing | LF/AAIF, Apache-2.0, enterprise endorsements | **research** (2026 paper), open-source, early | internal, single-org |

✅ has it · ◐ partial · ❌ no.

---

## Pillar A — Connectivity / data plane *(agentgateway's home turf)*

- **MCP gateway / proxy** — an authenticating, policy-enforcing proxy in front of
  MCP `tools/call`. **agentgateway:** broad (tool federation across many upstream
  servers, all four transports stdio/HTTP/SSE/Streamable, OpenAPI→MCP bridging,
  per-tool/per-session RBAC). **ActPlane:** not a proxy. **ours:** Sentry gates
  `tools/call`; narrower (no SSE/federation). → **agentgateway OVERLAPS + exceeds
  our MCP plane.**
- **A2A** — the Agent2Agent protocol (capability discovery, cross-framework
  collaboration). **Only agentgateway. → SUPPLEMENT.**
- **LLM / inference routing** — provider fan-out + cost/latency-aware GPU-pool
  routing. **Only agentgateway** (ours is compression + measurement). →
  **SUPPLEMENT.**

## Pillar B — Authorization & policy *(the three-way crux)*

- **Policy engine** — the decision engine. **agentgateway:** CEL (homepage's
  "OPA-based" conflicts, unconfirmed). **ActPlane:** a small **DSL** of
  source→target constraints (+ IFC), **authored by the agent** from CLAUDE.md /
  AGENTS.md intent, compiled to eBPF. **ours:** **OPA/Rego as declared sole
  authority** (two decisions: default-deny MCP, default-allow+boundaries native),
  hard "no LLM for security" rule. → **All three differ; this is the seam any
  adoption hinges on.**
- **Verdict spectrum** — **ActPlane:** `notify` (guide, non-blocking) / `block`
  (recoverable synchronous syscall deny, no TOCTOU, agent can retry) / `kill`
  (terminate the process) — each carries a `reason` string fed back to the agent.
  **ours:** deny / `require_approval` (→ human) / allow. **agentgateway:**
  allow/deny + content guardrails. → ActPlane's block/kill split ≈ a richer DENY;
  our `require_approval` (human) is unique; ActPlane's `reason` semantic feedback
  to the *agent* is its own twist.
- **Native host-tool authz + indirect-path coverage** — *the headline.* Our
  `opa_delegate`/`opa-hook` intercept the **tool call** (they see `Bash(command="…")`).
  If the agent runs a script or spawns a subprocess, the inner `git commit` /
  `rm` / `curl` **bypasses the hook**. ActPlane enforces on the **exec/open/write/
  connect syscalls in the kernel**, so it catches the effect regardless of how it
  was launched. The paper measures this directly: on "script/hidden" traces,
  tool-regex (≈ opa-hook) and tool-level IFC (FIDES) drop to **~0%**, while
  ActPlane detects **77%** (→ 95% after one feedback-driven rule revision). →
  **ActPlane OVERLAPS our native plane and fixes its core gap.**
- **Cross-event / information-flow (IFC)** — ActPlane attaches labels to OS
  objects and propagates them along data-flow edges (monotonic), enforcing
  policies like *"data read from `.env` must not reach the network"* and
  *"tests must run after the most recent edit, before commit."* **Neither
  agentgateway nor ours has true cross-event IFC.** → **SUPPLEMENT (a capability
  both other systems lack).**
- **Higher-authority / privilege model** — **ActPlane:** policy is bound to a
  **process subtree (domain)**; a child can add/narrow local rules but **cannot
  remove or weaken any inherited rule**, and higher-authority policy is loaded
  *before* the agent starts (immutable to it). **ours:** the Entra-group
  **admin carve-out** is relax-only and comes only from cryptographically
  verified tokens. → **Strong conceptual alignment:** ActPlane's "agent narrows
  its own, never weakens inherited" is the same self-restriction principle as our
  "groups only ever relax, never tighten." Our human-authored **OPA bundle is the
  natural higher-authority layer**; ActPlane's agent-authored rules would be the
  local refinements beneath it.

## Pillar C — Governance & lifecycle *(ours alone)*

Human-in-the-loop elicitation (`require_approval` → human hold → async resume),
the agent runtime/sandbox (Omnigent), and the tracker→session workflow (Open
Engine) are **ours only** — out of category for both a proxy and a kernel
enforcer. *(ActPlane's "feedback" goes to the agent to self-correct, not to a
human; it has no notion of a task or a session lifecycle.)*

## Pillar D — Observability & audit

agentgateway is **OTEL-native** (per-tool/agent/tenant counters; a candidate
forwarding target for our `OE_AUDIT_LOG`). ActPlane emits **violation logs +
semantic feedback** but is an enforcer, not an audit-graph system (the paper
explicitly distinguishes itself from provenance systems like CamFlow). **Ours**
is the only one doing **cross-plane correlation** (authz + runtime + cost joined
on `session_id`/`subject_id`/`task_ref`, spoof-resistant via session↔subject
binding) — it sees runtime holds and task receipts that neither a wire proxy nor
a kernel hook is in the path of.

## Pillar E — Platform & ecosystem

- **agentgateway:** binary/Docker/**K8s** (Helm, control plane, Gateway API CRDs,
  mTLS rotation); LF/AAIF, Apache-2.0, enterprise endorsements (no confirmed
  production case studies); CNCF Tech Radar "Trial".
- **ActPlane:** **eBPF on Linux** (BPF-LSM hooks + tracepoints), ~3.2K LoC Rust +
  ~1.8K LoC BPF C, **1.9–8.4% overhead**; **research-grade** (2026 paper),
  open-source, early. Linux-kernel-bound by nature.
- **ours:** Go gateway + OPA sidecar (digest-pinned together), cron + CLI;
  internal/single-org.

---

## 2. Overlaps (genuinely the same job)

- **agentgateway ⟷ our MCP plane (Agentic-Sentry):** both authenticate + policy-
  gate `tools/call` and emit per-call telemetry; agentgateway is broader.
- **ActPlane ⟷ our native plane (`opa_delegate`/`opa-hook`):** both gate the
  agent's host execution against policy — but ActPlane does it at the **syscall
  level** and so covers the **indirect-execution paths our tool-call hooks miss.**

## 3. Supplements (capabilities we lack)

- From **agentgateway:** A2A, LLM/inference routing, K8s-native ops + mTLS
  rotation, MCP federation/transports, external governance + community.
- From **ActPlane:** **indirect-path (syscall-level) enforcement**, **cross-event
  information-flow control**, and the **self-restricting hierarchical domain**
  model — none of which we (or agentgateway) have.

## 4. What stays ours (out of both categories)

OPA/Rego-as-sole-authority governance, **human-ASK elicitation**, the **agent
runtime + Open Engine lifecycle**, and the **three-plane correlated audit** with a
spoof-resistant join. Neither a wire proxy nor a kernel enforcer competes here.

---

## 5. Recommendation

1. **All three are complements, not replacements — and they're additive at
   different layers.** Our governance/lifecycle/audit is the product; agentgateway
   and ActPlane are candidate *upgrades to our two enforcement planes*.
2. **MCP-plane spike (agentgateway):** can its **external-authz** delegate to our
   **OPA bundle** *and* carry the tri-state `require_approval` (not just
   allow/deny)? External authz is a listed feature, so this is scope/fit, not
   discovery. If yes → pilot it as the MCP data plane, retire Sentry's
   proxy/transport/observability code, keep our Rego + audit on top (gain
   federation, transports, mTLS, A2A, inference routing, K8s). If effectively
   **CEL-only** → keep Sentry; don't rewrite the bundle in CEL.
3. **Native-plane spike (ActPlane) — arguably the higher-value one:** our
   `opa_delegate`/`opa-hook` have the exact indirect-execution gap ActPlane is
   built to close. Evaluate adding an **eBPF/ActPlane-style enforcement layer
   beneath the native plane** so a `git commit`/`rm`/`curl` reached via a script
   or subprocess is still caught. Two integration shapes to assess:
   - **Belt-and-suspenders:** keep OPA/Rego authoring the policy (human authority)
     and add a kernel enforcer for syscall-level coverage of the same boundaries.
   - **Map the models:** treat our **OPA bundle as ActPlane's higher-authority
     (immutable) domain**, and let agent-authored ActPlane rules be relax-only
     local refinements — which fits our "relax-only, verified-source" rule.
   Caveat: ActPlane is **research-grade, Linux/eBPF-only**, and its **agent-
   authored DSL** sits in tension with our "no LLM in the security decision loop"
   — reconcilable only because enforcement stays deterministic (kernel) and the
   safety-critical higher-authority policy is human-authored and immutable.
4. **Harvest regardless of any swap:** agentgateway's **A2A** + **inference
   routing** as additive capabilities and its **OTEL** surface as our audit-log
   forwarding target; ActPlane's **cross-event IFC** as the model for data-flow
   policies (`.env`→network) we can't currently express.

---

## 6. Caveats / open questions

- **agentgateway:** CEL-vs-OPA (conflicting sources); external-authz scope;
  semantic-caching built-in vs external; no confirmed production case studies.
- **ActPlane:** research prototype (maturity/support); Linux/eBPF-bound; the
  agent-authored-policy philosophy must be reconciled with our human-authored-OPA
  authority rule; content-only and chat/semantic harms are explicitly out of its
  scope (it covers system-observable actions).

**Sources:** `https://agentgateway.dev/`,
`https://github.com/agentgateway/agentgateway`, `https://agentgateway.dev/docs/standalone/latest/`;
ActPlane paper (arXiv 2606.25189v1), `https://github.com/eunomia-bpf/ActPlane`.
