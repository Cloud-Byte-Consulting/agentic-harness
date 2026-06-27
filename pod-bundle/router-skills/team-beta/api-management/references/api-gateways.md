# API Gateways and Microgateways

Table of contents:
1. What the gateway is for (and isn't)
2. Policy ordering
3. Topologies: central, microgateway, sidecar
4. Routing patterns
5. Mediation / composition vs orchestration
6. Config as code
7. How the major products map

---

## 1. What the gateway is for (and isn't)

The gateway is the **policy enforcement and mediation point** between consumers and backend services. It owns the cross-cutting concerns so services don't each reimplement them:

- **Security**: authN (API key, OAuth/JWT, mTLS), authZ (scopes/claims), CORS, threat protection / WAF.
- **Traffic**: rate limiting, quotas, throttling/spike-arrest, response caching, load balancing.
- **Mediation**: routing, header manipulation, payload/format conversion (e.g. XML↔JSON), protocol conversion, redaction, fault handling, composition (split/join).
- **Observability**: request logging, metrics, tracing, analytics.

**What does NOT belong in the gateway:** business logic — `if/then/switch`, `for/while` loops, complex data transformations, and process orchestration. That goes in a service. The guiding principle (from microservices: "smart endpoints and dumb pipes") is to keep the pipe thin. Fat gateways/ESBs that absorbed business logic are exactly the second-generation anti-pattern the third-generation platform model rejects.

The *API* (interface + policies at the gateway) should stay **decoupled** from the *service* (logic + data). In older platforms the API and service were one deployment unit; modern platforms keep them separate and push the gateway close to the service.

---

## 2. Policy ordering

Policies execute as a pipeline on the request (and in reverse on the response). A sane default request order:

1. **Threat protection / WAF** — payload size limits, JSON/XML bomb protection, SQLi/XSS pattern checks, IP allow/deny. Cheapest rejections first.
2. **CORS preflight** handling (for browser clients).
3. **Authentication** — API key validation, then OAuth2/JWT verification (signature via JWKS, `iss`/`aud`/`exp`), and/or mTLS client-cert check.
4. **Authorization** — scope/claim/role checks for the specific route/method.
5. **Rate limiting & quota** — per app/user/route; reject `429` before doing expensive work.
6. **Cache lookup** — return cached response if fresh, skipping the backend.
7. **Mediation** — header injection, payload transform, redaction, composition.
8. **Route** to the backend (with retries/timeouts/circuit-breaking as configured).

On the response: backend → transform/redact → cache store → observability → client. Put rejections as early as cheaply possible; never run mediation or hit a backend for a request that will be rejected by auth or rate limit.

---

## 3. Topologies

**Central gateway** — one (clustered) gateway fronts everything. Simple to govern, but a single choke point and, for Kubernetes backends, often a *second* hop on top of the cluster ingress (outer gateway → ingress LB → service), adding latency, cost, and an extra thing to troubleshoot.

**Microgateway** — a lightweight gateway that *is* the cluster ingress / entry point into an independent runtime (e.g. Kubernetes). Collapses the two-layer "outer gateway + ingress" into one, fits natively into the runtime, and can leverage runtime features (service mesh, mTLS, discovery). Preferred for cloud-native deployments. Trade-off historically: not every vendor shipped a true microgateway (changing fast — Kong, Apigee, Gloo, Tyk, Azure APIM self-hosted gateway all offer lightweight/runtime-deployable gateways now).

**Sidecar gateway** — a gateway container attached to each service pod; the ingress acts only as load balancer/router, and each service's entry point is its own sidecar. Use when:
- services need fundamentally different gateway setups,
- you want blast-radius isolation (one gateway issue doesn't take down all services),
- different teams/services want different gateway vendors.
Trade-off: more gateways = more operational complexity, compute, and (for licensed products) cost.

```
Central:     client → [central GW] → ingress LB → svc
Microgateway: client → [ingress = microGW] → svc        (service mesh under it)
Sidecar:     client → ingress LB → [sidecar GW] → svc   (one per service)
```

This sits alongside a **service mesh** (e.g. Istio, Linkerd, Cilium): the mesh handles east-west service-to-service traffic, mTLS, and discovery; the API gateway handles north-south consumer traffic and API-level policy. They're complementary, not competing.

---

## 4. Routing patterns

- **Resource (URI/path) routing** — map a path to a backend, statically (`/orders → orders-svc`) or by template. The default.
- **Content-based routing** — route on header, body content, app key, or token claim when the URI alone is insufficient (e.g. route reads vs writes to different backends in a CQRS API so the consumer sees one URL).
- **Geo-routing** — route to the nearest regional gateway based on origin (latency, data residency).
- **Canary / weighted routing** — send a % of traffic to a new backend version for progressive rollout.

---

## 5. Mediation, composition vs orchestration

- **Mediation**: header handling, format conversion (XML↔JSON), protocol conversion, redaction (strip PII before it leaves), fault handling (turn a backend 500 into a clean problem+JSON). Stateless, fine at the edge.
- **Composition (aggregation)**: split one consumer request into several backend calls and join the responses into one payload — *with no business logic* (no conditionals, loops, or complex transforms). Cuts client chattiness (great for mobile). Acceptable at the edge or in a thin aggregator service.
- **Orchestration**: a stateful process flow (sequence of activities, branching, compensation). This is business logic — it belongs in a dedicated service/runtime, **not** the gateway. The litmus test: if it needs `if/for/while` or a process state machine, it's orchestration.

---

## 6. Config as code

Treat gateway configuration like application code: declarative, in version control, deployed via CI/CD, promoted across environments.

- **Kong** — declarative YAML applied with `decK` (or Kubernetes CRDs via the Ingress Controller).
- **Azure API Management** — Bicep/ARM/Terraform for the service, named values for secrets, policy XML per API/operation, APIOps pipelines.
- **Apigee** — API proxy bundles (revisions) deployed via `apigeecli`/Maven; sharedflows for reusable policy.
- **AWS API Gateway** — define via OpenAPI with `x-amazon-apigateway-*` extensions, deployed by SAM/CloudFormation/CDK/Terraform.
- **Tyk** — API definitions as JSON, synced via the Tyk Operator (Kubernetes) or Dashboard API.

Never hand-edit production gateway config in a console; round-trip it through the repo.

---

## 7. How the major products map

All implement the same conceptual blocks; names differ.

| Concept | Apigee | Kong | Azure APIM | AWS API Gateway |
|---|---|---|---|---|
| Policy unit | Policy (in proxy/flow) | Plugin | Policy (XML) | Authorizer / integration / usage plan |
| AuthN/OAuth | OAuthV2, VerifyJWT | jwt, oauth2, openid-connect plugins | validate-jwt | Lambda/JWT authorizer, Cognito |
| API key | VerifyAPIKey | key-auth plugin | subscription key | API key + usage plan |
| Rate limit | Quota, SpikeArrest | rate-limiting, rate-limiting-advanced | rate-limit, quota | throttling + usage plan quota |
| Cache | ResponseCache | proxy-cache | cache-lookup/store | API Gateway caching |
| Routing | ProxyEndpoint/TargetEndpoint | Route/Service | API/operation + backend | resource + integration |
| Microgateway | Apigee adapter for Envoy / hybrid | Kong (DB-less) as ingress | self-hosted gateway | (managed; no self-host) |
| Dev portal | Integrated portal | Konnect portal / Kong Dev Portal | Developer portal | (no first-class portal; use third-party) |

When advising, stay vendor-neutral about the *pattern* and note the product-specific name only when the user names their product.
