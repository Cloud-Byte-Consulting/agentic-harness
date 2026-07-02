# opa-hook backend — Docker & minikube deploy

Stands up OPA loaded with the real `mcp_auth.rego` (the `oe_decision` rule that
`tools/opa_hook.py` queries) as a container, instead of a bare `opa run` host
process. Scoped to exactly two targets on purpose: **Docker** (single
container, fastest local loop) and **minikube** (local Kubernetes, closer to a
real cluster deploy). No other target (bare cloud VM, EKS/GKE/AKS, a managed
serverless container service) is in scope here — see
[docs/architecture/deploy/](../../docs/architecture/deploy/) for those.

Guided, step-by-step walkthrough (detects what you have, offers the Docker/
minikube choice, wires the CLI hooks after):
[opa-hook-setup-prompt.md](../../docs/architecture/opa-hook-setup-prompt.md).

Both scripts need [Agentic-Sentry](../../../Agentic-Sentry) cloned as a sibling
of this repo (`../Agentic-Sentry` relative to `agentic-harness/`) — override
with `AGENTIC_SENTRY_POLICIES` if it lives elsewhere.

## Docker

```bash
./docker/run.sh
```

One container, bind-mounts the live `Agentic-Sentry/mcp-policies/` tree
read-only — no build or copy step, so a rego edit just needs a container
restart to take effect. Stop with `docker rm -f opa-hook-backend`.

## minikube

```bash
minikube start   # if not already running
kubectl config use-context minikube
./minikube/deploy.sh
```

Deploys a 1-replica `Deployment` + `Service`. `mcp_auth.rego` is loaded via a
`ConfigMap` generated at deploy time from the live file (not checked into
`opa.yaml`, so it can't drift). **`deploy.sh` refuses to run unless the
current `kubectl` context is literally `minikube`** — this is a local-dev tool,
not a path to any other cluster, by design.

Reach the deployed OPA with:

```bash
kubectl port-forward svc/opa-hook-backend 8181:8181
```

**Gotcha already fixed here, worth knowing if you edit `opa.yaml`:** point OPA
at the specific file (`/policies/mcp_auth.rego`), not the `/policies` mount
directory. A ConfigMap volume is a symlink tree (`..data/`, a timestamped
snapshot dir, and the file itself all resolve to the same content); OPA's
directory walk picks up all of those as separate copies of the same package
and fails to start with `multiple default rules found`. Loading the file
directly sidesteps the walk. (Verified live against a real minikube cluster —
this isn't theoretical.)

## Either way, then point opa_hook.py at it

```bash
export OMNIGENT_OPA_URL=http://127.0.0.1:8181   # only needed if not this default
export OMNIGENT_OPA_DELEGATE_MODE=shadow          # observe first, then enforce
python3 ../../tools/opa_hook.py --self-check       # offline sanity check, either way
```

## What this does NOT do

Deploys OPA only — the plain `/v1/data/mcp/auth/oe_decision` query
`tools/opa_hook.py` needs. It does not stand up the Sentry gateway or
agentgateway MCP-plane (`experiments/cl09-real-policy/` already covers that
separately, for the MCP `tools/call` enforcement point — a different concern
from the native-tool hook this backs).
