# CLO-9 — agentgateway `extAuthz → OPA` (the bet): **PROVEN**

Empirically answers the agentgateway spike's central question (Q1/E2): **can
agentgateway defer the authorization verdict to our OPA bundle?** Yes.

## Result (2026-06-27, live)

```
echo (expect 200): 200      # OPA allow → forwarded to the backend
add  (expect 403): 403      # OPA deny  → blocked before the backend
```

OPA's decision log (per call) shows it received the **full Envoy `ext_authz` *v3***
`CheckRequest` **including the JSON-RPC body** — `parsed_body.params.name` = `echo`/`add` —
evaluated `mcp/authz/allow`, and returned `result: true`/`false`. agentgateway
**enforced** that verdict. So:

- `extAuthz` speaks **Envoy ext_authz v3 gRPC** → OPA's `envoy_ext_authz_grpc` plugin
  natively — **no shim, no Rego→CEL port**.
- **Per-MCP-tool** decisions work: `includeRequestBody` ships the JSON-RPC body and OPA
  reads `params.name` itself (extAuthz is HTTP-layer, not MCP-decoded).
- The decision **authority is OPA**, not agentgateway's CEL.

## Run it

```bash
docker compose up -d
GW=http://localhost:13000 ./test.sh        # echo→200, add→403
docker compose logs opa | grep '"msg":"Decision Log"'   # the verdicts + input
docker compose down
```

Images are pulled (`ghcr.io/agentgateway/agentgateway:latest`, `openpolicyagent/opa:latest-envoy`,
`hashicorp/http-echo`). The gateway is on host **:13000** (3000 inside).

## The config gotcha (for the next person)

The standalone `extAuthz` **flattens** the backend — the `Opaque` variant is keyed
`host:` directly under `extAuthz` (NOT `target: { host: … }`, which is the K8s CRD
form). See `agentgateway.yaml`.

## Not yet covered (follow-ups, per the spike)

- Plain-HTTP route to a **stub** backend here, to isolate the authz delegation. Next:
  a real **MCP backend** on an `mcp` route (per-tool end to end).
- `policy.rego` is a 1-rule stand-in. Next: point `extAuthz` at our **actual
  `mcp-policies` bundle** (`data.mcp.auth.decision`) via the plugin path.
- **Tri-state** `require_approval` — ext_authz is allow/deny; keep ASK at Omnigent (or
  a `428`/deny-with-status signal).
- **Entra groups** — forward the JWT (`includeRequestHeaders: ["authorization"]`) so the
  `is_admin` carve-out works; **`bindSession`** anti-spoof; the OTEL **audit-join** fields.

## Files
`docker-compose.yml` · `agentgateway.yaml` (extAuthz config) · `policy.rego` (OPA-Envoy
policy reading the JSON-RPC body) · `test.sh` (allow/deny).
