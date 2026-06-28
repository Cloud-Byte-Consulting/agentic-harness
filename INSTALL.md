# Install guide

How to install and run the governed agentic platform — from a laptop to a cloud cluster.

> Pairs with [GETTING_STARTED.md](GETTING_STARTED.md) (the pillars) and the [Glossary](GLOSSARY.md).
> The consolidation is mid-flight: **what's working today vs tracked-as-work is marked inline** and in [Tracked work](#tracked-work).

## Deployment targets at a glance

| Target | Status | Identity | Notes |
| :-- | :-- | :-- | :-- |
| **Local** | ✅ working | static token or any OIDC | native binaries or `docker compose` |
| **Minikube** | ◐ partial → [CLO-16](#tracked-work) | any OIDC | omnigent + dashboard have k8s bases; policy-plane base + umbrella pending |
| **Kubernetes** (generic) | ◐ partial → [CLO-15](#tracked-work)/[16](#tracked-work) | any OIDC | same bases; needs the policy-plane base + an umbrella chart |
| **Azure AKS** | ☐ → [CLO-17](#tracked-work) | **Entra** | AKS + ACR + Entra Workload Identity |
| **Google GKE** | ☐ → [CLO-18](#tracked-work) | **Google** | GKE + Artifact Registry + Workload Identity |
| **AWS EKS** | ☐ → [CLO-19](#tracked-work) | **AWS Cognito** | EKS + ECR + IRSA + Cognito |

**Identity providers** (all real today, env-driven): **Microsoft Entra**, **Google**, **AWS Cognito** — see [Identity providers](#identity-providers).

---

## 0. Prerequisites

| Tool | For | Min |
| :-- | :-- | :-- |
| Docker + Compose | OPA + gateway stack, images | recent |
| Python + uv | Omnigent engine | Python ≥ 3.12, uv ≥ 0.11 |
| Go | gateway, Cachy, teo, `air` | 1.25+ (gateway ≥ 1.22) |
| Node + npm | token-dashboard, web UI | Node 22 |
| tmux + bubblewrap | Omnigent native terminals (Linux) | — |
| kubectl + kustomize | any k8s target | recent |
| minikube / `az` / `gcloud` / `eksctl` | the matching target | — |
| A model key | LLM turns | `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` |

---

## 1. Local

### 1a. See governance work in 2 minutes (no build)

The proven CLO-9 experiment — our real policy through agentgateway → OPA:

```bash
cd ~/air/workspace/agentic-harness/experiments/cl09-real-policy
docker compose up -d && GW=http://localhost:13000 OPA=http://localhost:18181 ./test.sh
#  get_repo 200/allow · delete_repository 403/deny · get_email 428/require_approval
docker compose down
```

### 1b. Full local install

**Policy plane** — OPA + the shared bundle + the Sentry gateway (containers):

```bash
cd ~/air/workspace/Agentic-Sentry
cp deploy/.env.example .env                       # set GATEWAY_API_TOKEN (+ OIDC vars for real auth)
GATEWAY_API_TOKEN=dev-token-change-me docker compose up --build
#  OPA :8181 (loads mcp_auth.rego; decision /v1/data/mcp/auth/decision) · gateway :8080/mcp
#  the gateway DEFAULT-DENIES every tools/call if OPA is unreachable (fail-closed)
```

> Or run them as **native binaries** (lighter; what this workspace currently does): `go run ./cmd/gateway` + a local `opa run --server`. Either way they bind `127.0.0.1`.

**Engine** — Omnigent:

```bash
cd ~/air/workspace/omnigent
uv sync --extra all --extra dev && source .venv/bin/activate
export ANTHROPIC_API_KEY=...                       # or OPENAI_API_KEY
pip install 'omnigent[semantic]'                   # optional: CLO-7 semantic stop
pip install 'rlms[docker]'                         # optional: CLO-11 rlm_query
omnigent run examples/role_router/                 # or: omnigent server start  (:6767)
```

**Turn governance on:**

```bash
export OMNIGENT_OPA_URL=http://127.0.0.1:8181
export OMNIGENT_OPA_DELEGATE_MODE=shadow           # observe; then 'enforce' (fails closed)
export AIR_SENTRY_GATEWAY_URL=http://127.0.0.1:8080/mcp
export GATEWAY_API_TOKEN=dev-token-change-me
```

⚠ **Wired vs manual:** `enforce` is AND-gated on a session carrying the `openengine.profile=<name>` label, and the OE profile's **MCP-at-Sentry injection isn't consumed yet** ([CLO-20](#tracked-work)). Route MCP through Sentry **per-agent in YAML** for now:
```yaml
tools:
  agentic-sentry: { type: mcp, url: http://127.0.0.1:8080/mcp, token_env: GATEWAY_API_TOKEN }
```

**Optional token plane:** `cd Cachy && go build ./cmd/cachy && ./cachy proxy --listen 127.0.0.1:8787 --target <upstream>` · `cd token-dashboard && npm install && npm start` (:3000).

**Optional `air` content CLI:** `cd agentic-harness && export GOPRIVATE=truenas-scale-1.tail5a208d.ts.net && go build -o air ./air` (needs Gitea module auth for `teo`). `air bootstrap`/`session` are removed — runtime lives in Omnigent.

---

## 2. Kubernetes (minikube + generic)

**What exists:** kustomize bases for **omnigent** (`omnigent/deploy/kubernetes/base/`) and **token-dashboard** (`token-dashboard/deploy/k8s/`); all four core services have **Dockerfiles**.
**What's pending:** a kustomize base for the **policy plane** (Sentry gateway + OPA) — [CLO-15](#tracked-work) — and an **umbrella chart + minikube quickstart** that composes OPA + gateway + omnigent + dashboard with one set of values (image tags, ingress, the OIDC/IdP secrets, the shared bundle) — [CLO-16](#tracked-work).

**Intended flow (per CLO-16):**
```bash
minikube start --addons=ingress
eval $(minikube docker-env)                        # build images into the cluster
#   build: agentic-sentry, omnigent, token-dashboard, cachy
helm install openengine deploy/helm/openengine \   # (chart delivered by CLO-16)
  -f values-minikube.yaml \
  --set oidc.providers=azure --set-file opa.bundle=mcp-policies/policies/mcp_auth.rego
#   smoke: a tools/call through the gateway Service → allow 200 / deny 403 / approval 428
```
Generic k8s is the same chart against any cluster + a real registry + ingress host.

---

## 3. Cloud (AKS · GKE · EKS)

Each cloud target = the umbrella chart on a managed cluster + that cloud's registry + **workload-identity-federated** OIDC so the gateway validates that cloud's tokens. Per **decision D17**, each agent runs under its **own dedicated, non-admin agent identity** (see [agent-identity docs](docs/architecture/agent-identity/README.md)).

| | Cluster | Registry | Workload identity → IdP | Tracked |
| :-- | :-- | :-- | :-- | :-- |
| **Azure** | AKS | ACR | AKS Workload Identity → **Entra** (`azure`) | [CLO-17](#tracked-work) |
| **Google** | GKE | Artifact Registry | GKE Workload Identity → **Google** (`google`) | [CLO-18](#tracked-work) |
| **AWS** | EKS | ECR | IRSA + **Cognito** (`aws`) | [CLO-19](#tracked-work) |

Each issue carries the full provisioning + identity + secrets steps for its cloud. Secrets (the `GATEWAY_API_TOKEN`, IdP client IDs, bundle-store creds) come from the cloud's secret store via the CSI driver — never baked into images.

---

## Identity providers

The gateway resolves OIDC providers from the **`OIDC_PROVIDERS`** env list (`Agentic-Sentry/internal/auth/auth.go`). Each token's `groups` claim feeds the OPA **admin carve-out**. Set `OIDC_PROVIDERS` empty to use static `GATEWAY_API_TOKEN` (dev only).

| Provider | `OIDC_PROVIDERS` value | Required env | Per-cloud identity model |
| :-- | :-- | :-- | :-- |
| **Microsoft Entra** | `azure` | `AZURE_TENANT_ID`, `AZURE_CLIENT_ID` | Entra Agent ID — [azure.md](docs/architecture/agent-identity/azure.md) |
| **Google** | `google` | `GOOGLE_CLIENT_ID` | GCP IAM / Vertex — [google.md](docs/architecture/agent-identity/google.md) |
| **AWS Cognito** | `aws` | `AWS_OIDC_ISSUER`, `AWS_OIDC_CLIENT_ID` | Bedrock AgentCore — [aws.md](docs/architecture/agent-identity/aws.md) |

Multiple providers can run at once: `OIDC_PROVIDERS=azure,google,aws`. The engine's own SSO is `OMNIGENT_OIDC_ISSUER` (optional).

---

## Tracked work

The incomplete deployment targets are on the Linear board — each issue is self-contained (project context, real file paths, env, acceptance):

| Issue | Work |
| :-- | :-- |
| **CLO-15** | Kustomize base for the policy plane (Sentry gateway + OPA sidecar, fail-closed) |
| **CLO-16** | Umbrella chart + **minikube** quickstart (composes all core services) |
| **CLO-17** | **Azure AKS** + ACR + Entra Workload Identity |
| **CLO-18** | **Google GKE** + Artifact Registry + Workload Identity |
| **CLO-19** | **AWS EKS** + ECR + IRSA + Cognito |
| **CLO-20** | OE stack-profile **Lane-B** — auto-inject MCP servers + skills (removes the manual per-agent YAML) |

## Honest status

- **Local** is fully working (native or compose).
- **K8s** has two of four bases; the policy-plane base, the umbrella chart, and the cloud overlays are tracked (CLO-15…19).
- **`opa_delegate`** is opt-in (`off → shadow → enforce`) and AND-gated on the profile label; **OE-profile MCP auto-inject** is CLO-20.
- **ActPlane** (kernel indirect-exec) is staged — [`experiments/cl10-actplane/`](experiments/cl10-actplane/). **SIEM/OTel** export is Phase 5.
