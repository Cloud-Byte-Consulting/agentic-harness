# Install — start to finish

How to stand up the governed agentic platform. Two paths:

- **A. See it work in 2 minutes** — the policy plane only, via the proven experiment.
- **B. Full install** — OPA + Sentry gateway → Omnigent engine → (optional) token plane + `air` CLI.

> Read alongside [GETTING_STARTED.md](GETTING_STARTED.md) and the [Glossary](GLOSSARY.md).
> Status: the consolidation is mid-flight — **honest "wired vs manual" notes are inline**, don't skip them.

---

## 0. Prerequisites (one-time)

| Tool | For | Min version |
| :-- | :-- | :-- |
| **Docker + Compose** | OPA + the Sentry gateway stack | recent |
| **Python + uv** | Omnigent engine | Python ≥ 3.12, uv ≥ 0.11 |
| **Go** | Sentry, Cachy, teo, `air` | **1.25+** (covers all; Sentry needs ≥1.22) |
| **Node + npm** | token-dashboard, Omnigent web UI | Node 22 |
| **tmux + bubblewrap** | Omnigent native terminal/sandbox (Linux) | — |
| **A model key** | the LLM turns | `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` |
| **Entra/OIDC app** | production auth (optional for local) | — |

---

## A. See it work in 2 minutes (policy plane)

This is the **proven** CLO-9 experiment — our real policy enforced through agentgateway → OPA:

```bash
cd ~/air/workspace/agentic-harness/experiments/cl09-real-policy
docker compose up -d
GW=http://localhost:13000 OPA=http://localhost:18181 ./test.sh
#  get_repo 200/allow · delete_repository 403/deny · get_email 428/require_approval
docker compose down
```

That's the governance model in miniature: one Rego bundle, tri-state verdicts, fail-closed.

---

## B. Full install

### 1 — Policy plane: OPA + the shared bundle + the Sentry gateway

```bash
cd ~/air/workspace/Agentic-Sentry
cp deploy/.env.example .env          # set GATEWAY_API_TOKEN (+ OIDC vars for prod)
GATEWAY_API_TOKEN=dev-token-change-me docker compose up --build
```

- Brings up **OPA on `:8181`** (loads `mcp-policies/.../mcp_auth.rego`; decision at `/v1/data/mcp/auth/decision`) and the **gateway on `:8080/mcp`**.
- Verify: `curl localhost:8080/health`. The gateway **default-denies every `tools/call` if OPA is unreachable** (fail-closed — by design).
- `mcp_auth.rego` is the **one bundle** the native plane (Omnigent) also uses (`data.mcp.auth.oe_decision`).

### 2 — Engine: Omnigent

```bash
cd ~/air/workspace/omnigent
uv sync --extra all --extra dev        # dev install from this checkout
source .venv/bin/activate
export ANTHROPIC_API_KEY=...           # or OPENAI_API_KEY

# optional governance extras:
pip install 'omnigent[semantic]'       # CLO-7 semantic early-stopping embedder
pip install 'rlms[docker]'             # CLO-11 rlm_query long-context tool

# run a governed agent, or the server + web UI (:6767):
omnigent run examples/role_router/
#   — or —
omnigent server start                  # status: omnigent server status · stop: omnigent stop
```

### 3 — Turn governance on (OPA delegation)

```bash
export OMNIGENT_OPA_URL=http://127.0.0.1:8181
export OMNIGENT_OPA_DELEGATE_MODE=shadow   # query + log only; switch to 'enforce' when ready
export AIR_SENTRY_GATEWAY_URL=http://127.0.0.1:8080/mcp
export GATEWAY_API_TOKEN=dev-token-change-me
```

- `opa_delegate` defaults to **`off`** — set `shadow` (observe) then **`enforce`** (DENY/ASK; OPA-unreachable → fail closed).
- **⚠ Wired vs manual (today):** `enforce` is AND-gated on the session carrying the `openengine.profile=<name>` label, and the OE stack profile's **MCP-at-Sentry injection is not yet consumed** (Lane-B pending; it used to live in the now-removed `air bootstrap`). **To route MCP through Sentry now, declare it per-agent in YAML:**
  ```yaml
  tools:
    agentic-sentry: { type: mcp, url: http://127.0.0.1:8080/mcp, token_env: GATEWAY_API_TOKEN }
  ```
- `require_approval` at the **bare** gateway fails closed (deny); the human **ASK** renders only inside an Omnigent-mediated session.

**Minimal governed platform = steps 1–3.** The rest is optional.

### 4 — (Optional) Token / cost plane

```bash
# Cachy — context-optimization proxy (point an agent's API base at it)
cd ~/air/workspace/Cachy && go build ./cmd/cachy
./cachy proxy --listen 127.0.0.1:8787 --target <upstream-llm-url>

# token-dashboard — cost visibility (:3000)
cd ~/air/workspace/token-dashboard && npm install && npm run build && npm start
#   npm run data   # optional: parse your real ~/.claude and ~/.codex logs
```

### 5 — (Optional) `air` content CLI

```bash
cd ~/air/workspace/agentic-harness
# teo is a private Gitea module — configure module auth first:
export GOPRIVATE=truenas-scale-1.tail5a208d.ts.net
git config --global url."http://<GITEA_TOKEN>@truenas-scale-1.tail5a208d.ts.net:30008/".insteadOf "https://truenas-scale-1.tail5a208d.ts.net/"
go build -o air ./air
./air status        # list manifest components · also: skills sync · personas list · agents link
```

> `air bootstrap` / `air session` are **removed** — all runtime/session/MCP-inject concerns now live in Omnigent (the `air` CLI is content-only).

---

## Verify the whole thing

1. **Policy:** run path **A** → tri-state 200/403/428.
2. **Engine + native plane:** with step 3 in `enforce`, run an agent and attempt a denied action (e.g. a `delete_*` tool) → DENY in the OPA decision log on `:8181`.
3. **Both planes, one bundle:** the same `mcp_auth.rego` decision shows up for an MCP call (via Sentry `:8080`) and a native call (via `opa_delegate`).

## Honest status

- **OE stack profile auto-injection** (MCP/skills/harness-map) is **Lane-B pending** → wire MCP-at-Sentry per-agent in YAML for now.
- **`opa_delegate`** is opt-in (`off → shadow → enforce`) and AND-gated on the profile label.
- **ActPlane** (CLO-10, kernel indirect-exec enforcement) is staged, not a default — see [`experiments/cl10-actplane/`](experiments/cl10-actplane/) (`sudo`, BPF-LSM host).
- **SIEM/OTel export** of decision logs is a Phase-5 deliverable.
