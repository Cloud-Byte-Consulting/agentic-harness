# Evaluation plan — ActPlane as a kernel enforcement layer beneath the native plane

**Type:** Time-boxed spike (~1–2 weeks). **Owner:** TBD. **Status:** Proposed.
**Informs:** [agentgateway-comparison.md](agentgateway-comparison.md) §5.3 (native-plane spike).

## 1. The decision this spike informs

Should we add an **eBPF / ActPlane-style OS-level enforcement layer beneath our
native plane** (`opa_delegate` + `opa-hook`) to close the **indirect-execution
gap**, with our **OPA/Rego bundle as the immutable higher-authority policy**?

Outcomes: **GO** (build/adopt an eBPF enforcement layer) · **PARTIAL** (adopt only
for Linux server-side sessions) · **NO-GO** (document the residual gap + cheaper
mitigations).

## 2. Background — the gap we already have

Our native plane gates **tool calls**: `opa_delegate` (Omnigent PreToolUse,
`omnigent/omnigent/policies/builtins/opa.py`) and `opa-hook`
(`agentic-harness/tools/opa_hook.py`) intercept `Bash`/`Edit`/`run_command`
*invocations* and query OPA `data.mcp.auth.oe_decision` (the `oe_boundary_deny` /
`oe_boundary_approval` named-tool boundaries).

ActPlane's thesis (arXiv 2606.25189): tool-call interception **misses indirect
execution paths** — a denied action reached via a script the agent wrote, a
subprocess, or a compiled binary never hits the hook. Their measurement: tool-
regex enforcement (≈ our `opa-hook`) scores **~0%** on "script/hidden" traces;
ActPlane's kernel IFC catches **77% → 95%** (after one feedback revision).

We must confirm this gap **on our own stack** before deciding to fix it.

## 3. Hypotheses

- **H1 (gap is real here):** an OE-denied action (`git commit` on a `tests-before-
  commit` boundary, `rm -rf` of a protected path, `curl` to a blocked endpoint,
  `delete_repository` semantics) executed via a script/subprocess **bypasses
  `opa_delegate`/`opa-hook`** while a direct tool call is caught.
- **H2 (ActPlane closes it):** the same indirect actions are **caught at the
  syscall level** by ActPlane.
- **H3 (our authority maps cleanly):** our OE boundaries can be expressed as
  ActPlane higher-authority (immutable) DSL rules, and the "agent can only narrow,
  never weaken inherited" model is compatible with our "groups relax-only,
  verified-source" rule and our "no LLM in the security decision loop" rule.
- **H4 (overhead acceptable):** end-to-end overhead on a real Omnigent session is
  within the paper's 1.9–8.4% band.

## 4. Scope

**In:** Linux server-side / container agent execution (Bash/run_command host
tools). Reproduce the gap, prototype ActPlane enforcement, assess policy
translation + overhead.
**Out:** non-Linux editor hosts (eBPF is Linux-only — a stated limitation, not a
spike question); MCP-plane authz (that's the agentgateway spike); content/chat
harms (out of ActPlane's scope by design).

## 5. Make-or-break questions

| # | Question | Why it gates the decision |
| :-- | :-- | :-- |
| Q1 | Does our native plane actually bypass on indirect paths? | If not, there's no gap to fix. |
| Q2 | Does ActPlane catch the same actions at the syscall level? | The core value claim. |
| Q3 | Can our OE boundaries (`oe_boundary_deny`/`oe_boundary_approval`, named-tool denies) be expressed in ActPlane's DSL? | Determines whether we keep one source of truth. |
| Q4 | Can the OPA bundle act as the **immutable higher-authority domain**, with agent-authored rules relax-only? | Reconciles with "no LLM for security" + "relax-only" rules. |
| Q5 | How does `require_approval` (our tri-state) map to `notify/block/kill`? | ASK has no kernel analog — must stay at the OPA/Omnigent layer. |
| Q6 | Overhead on a representative session? | Production viability. |
| Q7 | Can it express a **cross-event/IFC** policy we can't today (`.env` → network)? | Bonus capability, not a gate. |

## 6. Experiments

> Prereqs: a Linux box with a BPF-LSM kernel (≥5.7, LSM=bpf enabled); build
> ActPlane (Rust runner ~3.2K LoC + BPF C ~1.8K LoC) from `github.com/eunomia-bpf/
> ActPlane`; our OPA + `mcp-policies` bundle running; a Claude Code or Omnigent
> session driving `Bash` with `OMNIGENT_OPA_DELEGATE_MODE=enforce`.

- **E1 — reproduce the gap (Q1).** For each of ~5 OE-denied actions, run it three
  ways and record whether the native plane denies it:
  1. **direct** tool call (`Bash(command="git commit …")`) — *expect: denied*;
  2. **script** (`Bash(command="bash ./x.sh")` where `x.sh` runs the action) —
     *expect: bypass*;
  3. **subprocess/shell-out** (`Bash(command="python -c 'os.system(\"git commit\")'")`).
  Output: a detection-rate table (native plane × {direct, script, subprocess}).
- **E2 — ActPlane catches them (Q2).** Author ActPlane DSL rules equivalent to the
  OE boundaries (e.g. `rule tests-before-commit: kill exec "git" "commit" …`).
  Re-run E1's matrix under ActPlane. Output: the same table, ActPlane column.
- **E3 — policy translation fidelity (Q3).** Take the `oe_decision` boundaries from
  `mcp-policies/policies/mcp_auth.rego` and hand-translate them to ActPlane DSL.
  Record: which boundaries translate cleanly, which need IFC/temporal gates, which
  can't be expressed (e.g. group-RBAC carve-out — ActPlane binds to process
  subtrees, not Entra identity). Output: a coverage matrix + gaps.
- **E4 — higher-authority domain mapping (Q4).** Load the OE-derived rules as the
  root domain *before* the agent starts; have the agent author a *narrowing* local
  rule and a *weakening* one. Confirm the kernel **accepts the narrowing, rejects
  the weakening** of an inherited rule. Output: pass/fail + the trust-boundary note.
- **E5 — `require_approval` handling (Q5).** Decide + document the mapping: a
  boundary that is `require_approval` in OPA → does ActPlane `block` + semantic
  feedback (agent re-routes) suffice, or do we keep ASK strictly at the
  OPA/Omnigent layer and use ActPlane only for hard `deny`/`kill`? Output: the
  enforcement-split design (likely: ASK stays at Omnigent; ActPlane enforces deny).
- **E6 — overhead (Q6).** Run a representative Omnigent agent session (build +
  test workload) with and without ActPlane (AP-N at a realistic rule count).
  Output: % end-to-end overhead vs the paper's 1.9–8.4%.
- **E7 — cross-event/IFC (Q7, bonus).** Express *"data read from `.env` must not
  reach the network"* as an ActPlane IFC rule; confirm a leak attempt is blocked.
  Output: a worked example of a policy class we can't express today.

## 7. Success criteria / decision gates

- **GO** if: E1 shows a real bypass (>50% of indirect paths slip the native plane)
  **and** E2 shows ActPlane closes it **and** E3/E4 show the safety-critical
  boundaries translate with the OPA bundle as immutable higher authority **and**
  E6 overhead is acceptable **and** E5 keeps ASK at our layer (so "no LLM for
  security" holds — enforcement stays deterministic, human-authored policy stays
  immutable).
- **PARTIAL** if the above holds but only on Linux server-side sessions (editor
  hosts on macOS/Windows stay on tool-call hooks) — adopt for the server tier,
  document the editor-host residual.
- **NO-GO** if: the DSL can't express our safety-critical boundaries, or the
  research-grade maturity is unacceptable for production, or overhead is too high
  — then document the residual native-plane gap and fall back to mitigations
  (tighten the `Bash` allowlist, run agents in a locked-down sandbox/container,
  restrict shell-outs).

## 8. Risks & mitigations

- **Research-grade maturity** (support, stability, kernel-version drift) →
  treat as an *architecture pattern* to adopt even if we don't ship their code;
  an eBPF enforcement layer is the durable idea.
- **Linux/eBPF-only** → scoped out for non-Linux hosts; PARTIAL outcome covers it.
- **Agent-authored-DSL vs "no LLM for security"** → reconciled only if the
  safety-critical policy is human-authored (OPA-derived) + immutable and agent
  authoring is relax-only; E4 must prove the kernel enforces that.
- **Two policy sources (OPA + ActPlane DSL)** drift → prefer generating the
  ActPlane rules *from* the OPA bundle, or keep OPA authoritative and ActPlane as
  a generated enforcement projection.

## 9. Deliverables

A reproduction of the native-plane gap (E1 table), an ActPlane PoC catching it
(E2), the OPA→DSL coverage matrix (E3), the higher-authority mapping result (E4),
the enforcement-split design (E5), an overhead number (E6), and a **1-page
go/partial/no-go memo** with the recommended integration shape.

## 10. Repo deep-dive — grounded specifics

Clone: `/home/bittahcriminal/air/workspace/research-repos/ActPlane` (reference impl, matches the paper). Key files: `docs/rule-language.md` (DSL grammar §2, IFC propagation §1.5, staleness §1.9), `docs/security_model.md` (domains/authority), `docs/support-matrix.md` (kernel reqs, effects, IPv4 limit), `script/e2e_examples.sh` + `test/e2e_cases.yaml` (live repro rig), `test/policies/{01,04,15,16,20}*.yaml`, `bpf/process.bpf.c`.

- **Kernel prereqs (sharper than §1):** Linux **5.8+ with BTF** (`/sys/kernel/btf/vmlinux`); pre-op **`block` needs BPF-LSM active** (`lsm=...,bpf` in cmdline) — without it you get **`notify`+`kill` only**. `run`/`watch` need root (drops priv to `SUDO_UID`). CLI ships a **prebuilt CO-RE `.o`** → product path needs only a Rust toolchain (`cargo install actplane`).
- **E1/E2 reuse the turnkey rig — don't build one:** `make && sudo bash script/e2e_examples.sh` fires E1–E12 live against the real eBPF engine. The **`E9 cross-tool-git`** case *is* our indirect-exec test: `source AGENT = exec "**/codex"` / `kill exec "git" if AGENT` — caught when `git` is a fork/exec descendant (label flows down the subtree), not caught when launched outside the agent subtree. **That label propagation is the mechanism that closes our native-plane gap.**
- **E3 translations verbatim:** tests-before-commit → `kill exec "git" "commit" if AGENT unless after exec "**/pytest" exits 0 since write "src/**" or write "tests/**"` (the `since` staleness is a class our OPA tool-call hook **cannot** express); `.env`→net → `test/policies/01_secret_no_exfil.yaml` (`block connect endpoint "*" if SECRET`); `rm -rf` → `04_workspace_confinement.yaml` (`block unlink file "/**" if AGENT unless target "/work/**"`, pre-op).
- **Two gates move toward PARTIAL (memo):** (a) **network endpoint matching is numeric-IPv4-only in-kernel** — hostname/domain denies (our `curl`-to-blocked-domain) **do not fire**; they stay at OPA / an egress proxy. (b) git-*subcommand* denies are `kill` (post-exec argv), **not** pre-op `block` — don't claim "never committed" for those (only executable-identity `block exec "git"` is pre-op).
- **E4 higher-authority is in-kernel + LLM-free (reconciles "no LLM for security" + relax-only):** load OPA-derived rules as the root domain bound **`mode: locked` before the agent starts**; child deltas may narrow/add but the kernel **rejects** weaken/disable/widen and won't let a child add a gate that satisfies an inherited `unless`. Caveat: `--approved-by` is **plain metadata, not crypto-verified** → rest immutability on *locked + kernel monotonicity*, not the approval fields. (`test/policies/15,16,20*.yaml`.)
- **E5/RBAC confirmed at OPA:** no kernel analog for `require_approval` (effects are `notify`/`block`/`kill`) → ASK stays at Omnigent; group-RBAC carve-out can't be expressed (subtree lineage ≠ Entra identity) → stays at OPA.
- **Anti-drift is directly implementable:** `actplane compile --json --explain` gives per-clause support + warnings (LSM availability, IPv4-only) → "OPA authoritative, ActPlane a **CI-validated generated projection**" (validate every generated rule in CI before load).
- **Overhead + gap confirmed from the paper:** AP-32 1.9%–6.5%, AP-100 3.8%–8.4% (≤128 rules; largest real repo 66). Hidden-trace DCR: baselines **0%** vs ActPlane **74%** (→94.7% with one feedback revision).
- **Net:** GO core (gap E1, catch E2, translate E3, immutable higher-authority E4) is real + runnable; the IPv4-only network limit and require_approval/RBAC-at-OPA are the PARTIAL-leaning factors — neither blocks the deterministic hard-deny + tests-before-commit + workspace-confinement center of gravity.
