# CLO-10 — ActPlane: our boundaries → DSL, + the indirect-exec demo (staged)

**Status:** staged for a live run. This machine **is** ActPlane-capable — kernel
7.0.12 with **BPF-LSM active**, BTF present, clang/bpftool, and the prebuilt CO-RE
object — but the live run needs **root (eBPF load)** and a **Rust build** (`cargo` is
absent here), so it's a one-command `sudo` step for the operator, not an autonomous run.

## The point

Our native plane (`opa_delegate` / `opa-hook`) is **tool-call interception**: it sees
`Bash(command=…)`, not what a script *inside* it does. ActPlane enforces at the
**syscall layer over the agent's exec lineage**, so a denied `git`/`rm`/`curl` reached
via a script is caught. This stages (a) our OE boundaries in ActPlane DSL and (b) a
one-command demo of the indirect-`git` catch.

## Run the demo (root)

```bash
sudo bash ~/air/workspace/scripts/cl10-actplane.sh
```

Expect **BOTH** `git --version` (direct) and `bash -c 'git --version'` (indirect) to
report **BLOCKED** — the indirect one is what tool-call hooks miss.

## E3 — our boundaries → ActPlane DSL (coverage matrix)

`our-boundaries.actplane.yaml` translates our boundaries (adapted from the repo's
`test/policies/*.yaml`). Coverage:

| Our boundary | ActPlane DSL | Coverage |
| :-- | :-- | :-- |
| **tests-before-commit** (temporal) | `kill exec "git" "commit" … unless after exec "**/pytest" exits 0 since write "**/*.py"` | ✅ — and the temporal gate is a class OPA tool-call hooks **can't** express |
| **workspace confinement** | `block unlink/​write file "/**" … unless target "<ws>/**"` | ✅ pre-op block |
| **.env → network** | `source SECRET = file "**/.env"; block connect endpoint "*" if SECRET` | ✅ |
| git-*subcommand* denies | `kill exec "git" "commit"` | ⚠ **kill** (post-exec argv), not pre-op `block` |
| network to a **hostname/domain** | — | ❌ kernel matches **numeric IPv4 only** → stays at OPA / an egress proxy |
| **group-RBAC** carve-out (Entra) | — | ❌ binds to process *subtrees*, not identity → stays at OPA |
| **require_approval** (tri-state) | — | ❌ effects are notify/block/kill, no human ASK → stays at **Omnigent** |
| destructive **MCP tool** (`delete_repository`) | — | ❌ HTTPS call, no syscall analog → stays at MCP / OPA |

→ the deterministic **hard-deny + tests-before-commit + confinement** core translates
cleanly (the GO center of gravity); hostname-network, RBAC, tri-state, and MCP-tool
denies stay at OPA — the PARTIAL factors from the spike, confirmed.

## E4 — higher-authority (for the live session)

Load these rules **`mode: locked` before the agent starts** → kernel-immutable; an
agent-authored child rule may **narrow** but not weaken/disable. Maps to: our
human-authored **OPA bundle = the locked higher-authority**; agent rules = relax-only,
verified-source — the "no LLM for security" reconciliation, enforced by the kernel.

## Files
`our-boundaries.actplane.yaml` (the DSL) · run harness: `~/air/workspace/scripts/cl10-actplane.sh`.
