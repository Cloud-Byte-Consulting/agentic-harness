# Evaluation plan — agentgateway as the MCP data plane

**Type:** Time-boxed spike (~1 week). **Owner:** TBD. **Status:** Proposed.
**Informs:** [agentgateway-comparison.md](agentgateway-comparison.md) §5.2 (MCP-plane spike).

## 1. The decision this spike informs

Should we adopt **agentgateway** as our **MCP data plane** — replacing
Agentic-Sentry's proxy / transport / observability code — while **keeping our
OPA/Rego bundle as the authority and our three-plane audit on top**?

Outcomes: **GO (swap)** · **HYBRID** (agentgateway for transport/federation/A2A,
keep Sentry's OPA gate) · **KEEP** (Sentry stays; don't rewrite the bundle in CEL).

## 2. Background — the overlap

agentgateway is a Rust data-plane proxy whose MCP-gateway + gateway-authn +
traffic-observability slab **overlaps Agentic-Sentry** and exceeds it (tool
federation, all four transports incl. SSE/Streamable, OpenAPI→MCP bridging, mTLS
auto-rotation, K8s-native deploy) — plus A2A and LLM/inference routing we don't
have.

The whole bet hinges on **one question (Q1 below):** can agentgateway defer the
authorization decision to **our OPA bundle** and honor our **tri-state**
(`allow` / `deny` / `require_approval`)? agentgateway's native policy language is
**CEL** (the homepage's "OPA-based" claim conflicts and is **unconfirmed**), but
**external authorization is a listed feature** — so this is a question of *scope
and fit*, not of existence.

## 3. Hypotheses

- **H1 (OPA delegation works):** agentgateway's external-authz can call our OPA
  decision (`data.mcp.auth.decision`) per `tools/call` and enforce the verdict.
- **H2 (tri-state carries):** `require_approval` survives the external-authz path
  (or is cleanly handled by keeping ASK at the Omnigent layer) rather than
  silently collapsing to `deny`.
- **H3 (identity + groups flow):** the authenticated subject and its **Entra
  groups** reach the OPA input, so the **admin carve-out** (`is_admin`) still works.
- **H4 (audit fields present):** agentgateway's decision/OTEL output carries
  `session_id`, `subject_id`, `tool`, `verdict` so the **three-plane correlation**
  join survives the swap.
- **H5 (gains materialize):** federation, SSE/Streamable transport, OpenAPI→MCP,
  and mTLS work with our backends.

## 4. Scope

**In:** the MCP-plane control point only — authn, OPA delegation, tri-state,
groups, audit fields, and the feature gains. **Out:** native host-tool gating
(that's the ActPlane spike); replacing Omnigent/Open Engine/the audit (those stay
ours regardless).

## 5. Make-or-break questions

| # | Question | Gate |
| :-- | :-- | :-- |
| Q1 | Can external-authz delegate the decision to our OPA bundle? | **The bet.** No → KEEP. |
| Q2 | Does `require_approval` carry (or map cleanly to our ASK gate)? | Tri-state integrity. |
| Q3 | Do the Entra **groups** reach the OPA input (admin carve-out)? | Authz correctness. |
| Q4 | Is there a session id we can **bind to the subject** (our `bindSession` anti-spoof)? | Audit trust. |
| Q5 | Do the OTEL/decision logs carry our 3-plane join fields? | Correlation survives. |
| Q6 | Do the feature gains (federation/transports/OpenAPI/mTLS) actually work? | Is the swap worth it. |
| Q7 | Is it CEL-only, or is there native OPA? | Migration cost framing. |

## 6. Experiments

> Prereqs: agentgateway (standalone binary or Docker); our OPA + `mcp-policies`
> bundle on `:8181`; a real MCP backend (e.g. the GitHub MCP) behind it; a real
> **Entra access token** (the OIDC path is already live — `OIDC_PROVIDERS=azure`,
> the group-RBAC overlay in `mcp-policies/tenant-data/mcp/auth/config/data.json`,
> `az account get-access-token`).

- **E1 — stand it up + authn (baseline).** Run agentgateway in front of one MCP
  backend; configure JWT/OIDC auth with our Entra app. Confirm: no token → 401;
  valid Entra token → accepted. (Mirror our `smoke-entra.sh` shape.)
- **E2 — external-authz → OPA (Q1, the bet).** Configure agentgateway's external
  authorization to call our OPA decision per `tools/call`. Send an **allow** tool
  (e.g. `ard.search`) and a **deny** tool (e.g. `github.create_issue` under the
  default-deny RBAC). Confirm the gateway's verdict **matches OPA's** — i.e. the
  decision authority is genuinely ours, not re-encoded in CEL. Record the exact
  external-authz contract (gRPC ext_authz? HTTP hook? request/response shape).
- **E3 — tri-state `require_approval` (Q2).** Hit a tool that OPA returns
  `require_approval` for (an `oe_boundary_approval` tool). Observe: does
  agentgateway surface a third state, or collapse to deny/allow? Decide the
  mapping — likely **keep ASK at the Omnigent layer** and let the gateway enforce
  only allow/deny, with `require_approval` resolved upstream. Document it.
- **E4 — groups + admin carve-out (Q3).** With a token carrying an admin group
  GUID, confirm the **groups reach the OPA input** through external-authz and the
  `is_admin` carve-out fires (an admin-only-allowed tool is allowed; a non-member
  is denied). This is the same check our live OPA group-RBAC smoke already does —
  re-run it *through* agentgateway.
- **E5 — session id + binding (Q4).** Determine agentgateway's session model;
  confirm we can obtain a stable per-session id and bind it to the subject (the
  anti-spoof property `bindSession` gives us today). Note any divergence.
- **E6 — audit fields (Q5).** Capture agentgateway's decision/OTEL output for the
  E2 calls; verify it carries (or can be enriched to carry) `session_id`,
  `subject_id`, `server_name`, `tool_name`, `verdict`. Map it to the contract-A
  `OE_AUDIT_LOG` schema and run `tools/oe_correlate.py` over it to confirm the
  three-plane join still resolves.
- **E7 — feature gains (Q6).** Exercise: **federation** (two upstream MCP servers
  behind one gateway), **SSE + Streamable** transport, **OpenAPI→MCP** (expose a
  REST endpoint as an MCP tool), **mTLS** to a backend. Record which work.

## 7. Success criteria / decision gates

- **GO (swap)** if: E2 proves OPA delegation works **and** E3 keeps the tri-state
  intact (carried or cleanly resolved upstream) **and** E4 flows groups **and** E6
  preserves the audit join **and** E7 confirms real gains. → Plan migration:
  agentgateway as the data plane, retire Sentry's proxy/transport/observability
  code, keep the Rego bundle + `bindSession`-equivalent + audit on top.
- **HYBRID** if OPA delegation is partial (e.g. allow/deny only, no clean tri-
  state) but the transport/federation/A2A gains are valuable → run agentgateway
  for transport/federation/A2A and keep Sentry's OPA gate for the verdict.
- **KEEP** if external-authz can't cleanly delegate to OPA (effectively CEL-only)
  or groups/tri-state can't flow → keep Sentry; **do not** rewrite the two-decision
  Rego bundle (`decision` + `oe_decision`, default-deny vs default-allow) in CEL.

## 8. Risks & mitigations

- **CEL-vs-OPA unconfirmed** (the headline risk) → E2 resolves it first; everything
  else is gated behind it.
- **Tri-state collapse** → fall back to ASK-at-Omnigent (E3); not fatal, but
  changes where the human gate lives.
- **Migration surface** (retiring Sentry code, re-pointing clients) → only scoped
  if GO; keep Sentry running in parallel during the pilot.
- **Loss of `bindSession` anti-spoof** → E5 must find an equivalent or the join's
  trust degrades (acceptable for a single-tenant pilot, not for the compliance
  claim).

## 9. Deliverables

A working **agentgateway + our-OPA** PoC (or a documented blocker at E2), the
external-authz contract shape, the tri-state/groups findings, an audit-join
confirmation via `oe_correlate.py`, a feature-gains checklist, and a **1-page
go/hybrid/keep memo** — with a migration sketch if GO.

## 10. Repo deep-dive — grounded specifics

Clone: `/home/bittahcriminal/air/workspace/research-repos/agentgateway` (Apache-2.0, LF). Ground truth = `schema/config.json` (267 KB). Run: `cargo run -- -f examples/authorization/config.yaml`; data plane on `binds[].port` (examples `:3000`); **admin UI/playground at `http://localhost:15000/ui`**. Example JWTs in `manifests/jwt/`.

- **Q1/E2 — the bet is winnable, no shim.** `extAuthz` speaks **Envoy gRPC ext_authz v3** (`schema/config.json` `ExtAuthz` ~L5490); OPA's `envoy_ext_authz_grpc` plugin speaks it natively. E2 = attach `extAuthz` to the MCP route → point at OPA-Envoy → OK/PermissionDenied. Set `failureMode: deny` (schema default = fail-closed).
- **E2 granularity caveat:** ext_authz is **HTTP-layer, not MCP-decoded** — to gate per `tools/call`, set **`includeRequestBody`** (raise `maxRequestBytes` from the 8192 default, `packAsBytes` for gRPC) so **OPA parses `params.name` from the JSON-RPC body itself**. (The only method-aware native hook is `McpGuardrails` `Processor kind: remote` — a custom webhook, needs its own OPA adapter; ext_authz+body is cleaner.)
- **Q3/E4 — groups flow:** `ExtAuthzProtocol.grpc.metadata` defaults to forwarding `envoy.filters.http.jwt_authn` claims (incl. Entra groups); or have OPA decode the bearer directly (`includeRequestHeaders: ["authorization"]`). `is_admin` carve-out survives unchanged.
- **Q2/E3 — no native `require_approval`; plan ASK-at-Omnigent.** Native `mcpAuthorization` is `allow`/`deny`/`require` where **`require` = "this CEL predicate must hold", not human approval**; ext_authz is OK/Denied only. Carry mechanism to test: OPA signals approval via **deny-with-status `428`** (`failureMode.denyWithStatus`) or an injected header on a synthesized allow → Omnigent maps to ASK. Primary remains ASK-at-Omnigent.
- **Q7 resolved — reframe the gate.** Native engine is **CEL-only (no Rego)**, so the homepage "OPA-based" claim is wrong for the *native* path — **but `extAuthz`→OPA is real**. The decision isn't CEL-vs-OPA; it's **"OPA via ext_authz sidecar" (keeps our Rego authoritative) vs "port Rego→CEL" (avoid).** GO = ext_authz→OPA.
- **E1 — use `mcpAuthentication`, not just `jwtAuth`.** `mcpAuthentication` (`examples/mcp-authentication/`) implements the MCP-OAuth spec (`resourceMetadata`, `/.well-known/oauth-protected-resource`); for Entra use the generic `issuer`/`jwks.url` (no built-in `azure` shortcut).
- **Q4/E5 — `Mcp-Session-Id` exists; subject-binding is undocumented** (engineer it, don't assume). Mitigation: ext_authz forwards the JWT **per request**, so `subject_id` is authoritative on every `tools/call` regardless of session binding.
- **Q5/E6 — audit is CEL-driven, not fixed-schema.** Emit `session_id`/`subject_id`/`server_name`/`tool_name`/`verdict` via `frontendPolicies.accessLog` CEL `add:` expressions (`examples/a2a/config.yaml`) + the `telemetry/` OTel example; E6 specifies those CEL exprs and maps to `OE_AUDIT_LOG`.
- **Q6/E7 — gains are few-line configs (checklist):** federation `examples/multiplex/`; OpenAPI→MCP `examples/openapi/`; transports stdio + remote `mcp:{host}` (Streamable `/mcp`, SSE); TLS/mTLS `examples/tls/`; A2A `examples/a2a/` (`a2a: {}`).
- **Net topology for the memo:** authn (`mcpAuthentication`) → forward JWT + JSON-RPC body to **our OPA via `extAuthz`** for the verdict → native CEL as optional coarse pre-filter. Keeps the Rego bundle authoritative; matches GO/HYBRID **without porting `decision`/`oe_decision` to CEL**. (`examples/delegation/` is *route* delegation, not authz — don't cite for Q1.)
