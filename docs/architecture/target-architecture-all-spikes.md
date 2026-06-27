# Target architecture — if all five spikes land

**Status:** Synthesis (2026-06-27). Produced by a verification workflow (6 per-spike
extractors grounded in the committed `eval-*.md` docs → synthesis → 3 adversarial
critics: correctness / completeness / honesty) and corrected for the critics'
findings. Companion to [research-triage.md](research-triage.md) and the five spikes:
[agentgateway](eval-agentgateway-mcp-plane.md) · [ActPlane](eval-actplane-native-plane.md)
· [RLM](eval-rlm-long-context.md) · [Agentic-BPM](eval-agentic-bpm-framed-autonomy.md)
· [semantic-stopping](eval-semantic-early-stopping.md).

## Bottom line

It stays **one architecture, not four bolt-ons — *if* one discipline holds**: OPA/Rego
stops being "the engine that enforces" and becomes the single policy **source** that
compiles into per-plane enforcement artifacts. A three-part spine holds it, and all
five spikes reinforce it rather than fork it:

1. **OPA/Rego = sole policy source** → projects into the wire, the kernel, the engine.
2. **Omnigent = the one home of the human ASK gate** → it can never be pushed down.
3. **`oe_correlate.py` = the one audit join** on `session_id`/`subject_id`/`task_ref`.

**Honesty caveat (load-bearing):** all five are **unrun spike *plans* (Status:
Proposed)**. Every headline number here — ActPlane's ~75% indirect-exec catch,
RLM's ~26% over compaction, semantic-stop's −38% tokens — is a **paper result on
another benchmark, pending replication on our stack**. "Adopt all five" really means
*a Rego spine + two deeper enforcement options (one PARTIAL) + one opt-in runtime +
one formalism + one loop predicate* — not five equal GOs.

## The original five pillars — what happens to each

- **A · Connectivity / data plane → deepens + gains a full external replacement.**
  Narrow Agentic-Sentry gateway → optionally a Rust **agentgateway** proxy (4 MCP
  transports, federation, OpenAPI→MCP, mTLS, A2A, inference routing). Realistic
  outcome **HYBRID** (agentgateway for transport/federation/A2A; Sentry kept for the
  OPA verdict) unless ext_authz→OPA proves clean. One Go gateway → a **layered data
  plane** (wire proxy + RLM's sub-LM broker + BPM's engine bridge).
- **B · Authz & policy → changes the most; layers while OPA stays the single source.**
  MCP plane ↔ **agentgateway** (Envoy gRPC ext_authz v3 → unchanged Rego; CEL only a
  coarse pre-filter). Native plane ↔ **ActPlane** (eBPF *below* the tool-call hooks →
  block/kill/notify on process-subtree-labelled descendants → closes the
  indirect-execution gap). BPM adds the **φ admission** check + a deterministic-first
  DecisionAgent. "OPA sole authority" survives only reframed as **"OPA sole *source*"**,
  compiling into **three** dialects — CEL, ActPlane DSL (`mode:locked`), φ YAML — each
  a **CI-validated generated projection, never hand-authored**. (We *replace* CUGA's
  markdown FRAME with OPA — the wedge, not a dialect we carry.) Tri-state
  `require_approval` has **no analog** in either deeper plane → ASK stays at Omnigent.
- **C · Governance & lifecycle → deepens; structural-authority model unchanged.**
  semantic-stop → embedding-convergence STOP predicate on `hillclimb_budget`; RLM →
  recursive long-context surface governed by `spawn_bounds`+`cost_plan`; ActPlane →
  kernel backstop (tests-before-commit's temporal `since` gate the hooks can't
  express, workspace confinement); BPM → 7-intervention control-point vocabulary +
  regulatory-override → hold → human-ASK → resume.
- **D · Observability & audit → deepens but most-strained.** More heterogeneous
  emitters; agentgateway makes the audit schema **CEL-driven**. Two real holes:
  (1) kernel events carry a **PID/subtree label, not `session_id`** → need a new
  launch-time binding (subtree-root PID ↔ session_id) or they're unjoinable;
  (2) per-request JWT recovers `subject_id` but **`bindSession`'s session-level
  anti-spoof must be re-engineered**.
- **E · Platform & ecosystem → single-org Go+OPA → polyglot + a vertical.**
  agentgateway → Rust + K8s; ActPlane → hard Linux/eBPF dependency (non-Linux editor
  hosts can't run it → keep the gap → PARTIAL); RLM → unhardened sandboxes; BPM → an
  agentic-BPM vertical (only LangGraph adapter complete). Footprint: **2 components →
  ~Go/Rust/Python/eBPF/K8s + a pinned embedding model + sandbox runtimes**.

## New *planes* (not new pillars)

No spike adds a sixth pillar — they add **planes** that fold into B/C/D but sit where
nothing else does:

1. **Kernel enforcement plane (eBPF/BPF-LSM)** — *ActPlane.* Below the native plane,
   at the syscall boundary; subtree-label propagation catches `git`-via-bash /
   `os.system()`. PARTIAL (Linux-only, IPv4-numeric net, git-subcommands are
   post-exec *kill* not pre-op *block*).
2. **φ execution-admission plane** — *BPM.* FRAME governs what the LLM may *reason*; φ
   governs what the engine may *act on* after hook reasoning. Its relationship to the
   OPA source is an **open modeling question**, not a settled projection.
3. **Sandbox / OS-isolation plane** — *added by the completeness critic.* fs +
   network-interception confinement. **RLM's containment floor** and the
   **cross-platform fallback wherever eBPF can't run**. The stack is wrong without it.
4. **The indirect-execution surface** — a named *seam*, not a built layer: invisible
   to all tool-call gating, and what **ties RLM and ActPlane together**. RLM is *fully*
   governed only once the kernel plane lands; until then the **hardened sandbox is the
   floor** and tool-call gating still governs the `rlm.completion()` boundary (not
   "ungoverned" — just not covered for in-process exec).

→ Defense-in-depth becomes **wire → tool-call → sandbox → kernel → engine**, OPA
feeding policy at each layer, Omnigent holding the human gate.

## Trade-offs

| Axis | Gain | Cost |
| :-- | :-- | :-- |
| **Policy-engine fragmentation** | Native-speed policy in-situ per layer | One Rego source → **3** dialects (CEL, ActPlane DSL, φ YAML); "sole authority" → "sole *source* + generated projections." **We own the anti-drift tooling — it doesn't exist yet.** |
| **Operational surface** | Best-of-breed per layer | 2 runtimes → ~5 + embedding model + sandboxes; Sentry only *partially* retires; kernel-version drift is new |
| **Platform binding** | Kernel-deterministic enforcement; K8s scale | ActPlane **Linux-only** → non-Linux hosts keep the gap (PARTIAL); hostname denies stay at OPA/an **egress proxy** (kernel IPv4-only) |
| **Maturity** | Closes 2 real gaps | ActPlane + RLM **research-grade** ("not for production"); the rec is **build-internally** = real eng-months, not an install |
| **Latency / cost** | RLM ~26%; semantic-stop ~38% — *both unmeasured here* | ext_authz round-trip + body-parse per call; RLM fan-out can explode; ActPlane's **1.9–8.4% per-session tax** on the Linux tier |
| **Team / skills** | Specialist per layer | One team must hold Rego + CEL + eBPF/Rust + BPMN + embedding ops + sandbox hardening — broad, deep, thin |
| **Availability** | — | **OPA becomes a correlated SPOF**: `failureMode:deny` on wire+native+engine → one OPA blip hard-stops everything. Needs HA + degraded-mode posture. |
| **Migration** | Every swap gated + reversible | `bindSession` re-engineered; OE_AUDIT_LOG goes CEL-driven; partial Sentry retirement leaves a seam |

## Open items — gaps the "one spine, one join" story must close

Surfaced by the completeness/correctness critics; **none is addressed by the spikes
as written** and each is load-bearing for the target architecture:

1. **A2A governance + join-key propagation.** An agent→agent hop may not carry the
   originating `session_id`/`subject_id` — the most join-key-threatening feature in
   the stack, sold as a free gain. Gate A2A behind the same ext_authz→OPA + join-key
   propagation as a direct `tools/call`.
2. **Kernel→session binding.** Kernel syscall events carry a PID/subtree label, not a
   `session_id`. Need a launch-time record (agent-subtree-root PID ↔ session_id) when
   Omnigent spawns the labelled subtree, or the kernel plane's audit is unjoinable.
3. **Sandbox / OS-isolation plane** must be explicit in the stack (RLM's floor +
   ActPlane's cross-platform fallback) — not just implied.
4. **OPA as a correlated availability SPOF.** `failureMode:deny` across wire + native
   + engine planes means one OPA outage hard-stops everything. Specify HA/replication
   + a documented degraded-mode posture per plane.
5. **Egress proxy as a named component.** Network enforcement is split: kernel
   (IPv4-numeric) + egress-proxy (hostnames) + OPA. The egress proxy belongs in
   pillar A / the stack.
6. **Federation ↔ OPA RBAC enumeration.** Each federated/OpenAPI-bridged tool needs a
   Rego rule or it's **dead (default-deny)** or **ungoverned (default-allow path)**.
   Federation's real cost is policy coverage, not config.
7. **Kernel + wire don't compose for remote exec.** The kernel plane covers only
   host-local in-subtree exec; a tool federated through agentgateway runs in a remote
   backend with **no kernel backstop** unless ActPlane runs on that host too.
8. **The anti-drift projection pipeline is unbuilt.** "OPA sole source + generated
   projections, never hand-authored" is the discipline that keeps no-LLM-for-security
   and OPA-authority alive across all five spikes — but the generator, the extended
   `oe_correlate`, and the version-pinned CEL emit-contract do not exist yet.

## The honest net + sequence

**One coherent architecture, conditionally** — held by the OPA-source + Omnigent-ASK +
one-audit-join spine. But **coherence is *achievable, not extant*: it depends on the
projection/anti-drift tooling, the extended `oe_correlate`, and the kernel→session
binding (open items 8/2), none of which is built.** Drop the discipline and it
fragments into four bolt-ons with conflicting policy languages.

**Sequence (cheapest/safest first):**

1. **semantic-stop** — now; tiny additive diff, no new plane. Skip "keep-best" unless
   the Judge already ranks rounds for free (else +129% cost); keep `k≥2` + a hard
   `max_rounds` failsafe (the convergence claim is a measured conjecture, not a
   theorem).
2. **BPM — formalism only** — adopt TDF/FRAME + make φ first-class; **defer the
   Camunda/Flowable vertical** (only LangGraph adapter exists).
3. **agentgateway** — gate on ext_authz→OPA; expect **HYBRID**; re-engineer
   `bindSession` before any compliance claim.
4. **ActPlane** — adopt the eBPF *pattern*, build/own internally, **Linux-only, accept
   PARTIAL**; highest-value gap-closer *and* the precondition for governing RLM.
5. **RLM** — **keep, opt-in now** (not the default): a governed long-context capability,
   gated on beating *our* compaction + the sandbox floor (+ ActPlane for full in-REPL
   coverage). **Target end-state: on-demand, agent-invoked** — the agent decides a job
   needs long-context and calls the RLM tool itself (the E5(a) `rlm_query` tool/skill
   shape, not a session mode). So sequence it *after* ActPlane lands (for governing the
   REPL), but it stays on the roadmap rather than shelved.
