---
name: api-management
description: >-
  Design, secure, govern, and operate APIs and the platforms that run them. Use for API design
  (REST resource modeling, GraphQL schemas, gRPC/protobuf), writing or reviewing
  OpenAPI/Swagger and AsyncAPI specs, and choosing an architectural style. Use for API
  gateways and microgateways (Kong, Apigee, Azure API Management, AWS API Gateway, Tyk),
  gateway policies, routing, and the gateway-vs-service boundary. Use for API security:
  OAuth2, OIDC, JWT validation, API keys, mTLS, scopes, CORS, OWASP API Top 10, rate limiting,
  quotas, throttling, and caching. Use for versioning and deprecation, lifecycle and
  design-first workflows, contract testing, developer portals, API products and monetization,
  analytics, and governance at scale. Triggers: "design an API", "review this OpenAPI spec",
  "rate limit", "401/403 from the gateway", "REST or GraphQL", "deprecate an API version",
  "API product", "developer portal".
---

# API Management

Equips Claude to design APIs, run the API gateway/platform, secure and govern APIs across their full lifecycle, and treat APIs as products — vendor-neutrally, with patterns that map onto Apigee, Kong, Azure API Management, AWS API Gateway, and similar.

## When to use this skill

- Designing a new API: resource modeling, choosing REST vs GraphQL vs gRPC, naming, status codes, pagination, error shapes.
- Writing or reviewing an OpenAPI (3.0/3.1), Swagger, or AsyncAPI specification; design-first vs code-first.
- Configuring an API gateway or microgateway: routing, policies, the split between gateway concerns and service/business logic.
- Securing APIs: OAuth2 grants, OIDC, JWT verification, API keys, mTLS, scopes, CORS, OWASP API Security Top 10, threat protection.
- Traffic management: rate limits, quotas, throttling/spike arrest, response caching, load balancing.
- Versioning, deprecation, and lifecycle governance; contract testing in CI.
- Building a developer portal / onboarding; packaging APIs as products with plans, billing, and monetization.
- Observability and analytics for APIs; SLAs and platform KPIs.
- Establishing API governance, standards, and a target operating model (central vs federated vs platform).

## Core concepts

**An API is a managed product, not a raw endpoint.** Treat each API as something with consumers, documentation, a lifecycle, an owner, and (often) a price. "An API is only as good as its documentation" — it is not "done" until docs and a portal page exist.

**Separate the API from the service.** The *API* is the interface plus the cross-cutting policies (auth, rate limiting, mediation) enforced at the gateway. The *service* implements business logic and data. Keep them decoupled: business logic does not belong in the gateway, and gateway concerns should not be reimplemented in every service ("smart endpoints, dumb pipes"). The third-generation platform model pushes the gateway close to the service (microgateway/sidecar) rather than one central monolithic proxy.

**The gateway is a policy enforcement point.** It does routing, authN/authZ, key validation, rate limiting/quotas, caching, transformation, threat protection, and observability. Anything stateful or business-logic-heavy (orchestration, complex transforms, conditionals/loops) belongs in a service, not the gateway. *Composition* (split/join across backends with no logic) can live at the edge; *orchestration* (a process flow) cannot.

**Design-first beats code-first.** Agree the contract (OpenAPI/AsyncAPI) early, mock it, get consumer feedback, iterate, *then* implement. Contract-test the implementation against the spec in CI so the running service can never drift from its published interface.

**Style fits the use case** (see `references/api-design.md` for the full comparison):
- **REST** — resource/HTTP-native, best tooling and gateway/cache/OAuth support, the safe default for public and partner APIs. Weak at composition; can be chatty.
- **GraphQL** — client picks fields; ideal when one client needs data composed from many sources (mobile, BFF). Avoid versioning by evolving the schema. Weaker HTTP caching and gateway/authZ story.
- **gRPC** — binary over HTTP/2, fast, streaming, great for internal service-to-service and high throughput. Poor browser fit; needs custom auth/caching plumbing.

**Security is layered.** Authentication (who are you) precedes authorization (what may you do). OIDC/SAML authenticate users (SSO); OAuth2 authorizes API access via access tokens carrying scopes. API keys identify the *application*, not a user, and are not authentication. mTLS authenticates the *client app* at the transport layer. See `references/api-security.md`.

## Workflow: how to approach API tasks

### 1. Designing an API
1. Start from the consumer and the business capability, not the database. What jobs does the consumer need done?
2. Pick the style (REST default; GraphQL for client-driven composition; gRPC for internal/high-throughput). Don't mix styles in one API without reason.
3. Model resources as nouns, use HTTP verbs and status codes correctly, plural collection names (`/orders`, `/orders/{id}`), pagination via `limit`/`offset` or cursors, a consistent error envelope, and idempotency for writes where applicable.
4. Capture the contract in **OpenAPI 3.1** (REST), an **SDL schema** (GraphQL), or **.proto** (gRPC). For event/async APIs use **AsyncAPI**.
5. Mock the spec, share it, collect feedback, iterate. Only then build.
6. Decide versioning strategy up front (see lifecycle reference). Default: URI-path major version (`/v1/...`) for REST public APIs; never break within a major.

### 2. Standing up the gateway
1. Configure routing first (resource/path routing; content-based routing only when the URI is insufficient).
2. Layer policies in order: threat protection / WAF → authN (key, OAuth/JWT, mTLS) → authZ (scopes) → rate limit/quota → cache → transform/mediate → route to backend.
3. Keep policy config declarative and version-controlled (e.g. Kong decK, APIM Bicep/ARM, Apigee API proxy bundles, AWS OpenAPI extensions). Treat gateway config as code.
4. Choose topology: a single central gateway, a **microgateway** as the cluster ingress (good fit for Kubernetes), or a **sidecar** gateway per service (when services need different gateway setups or vendors). See `references/api-gateways.md`.

### 3. Securing the API
1. Default-deny. Every public route needs an explicit authN policy.
2. Use **OAuth2 + OIDC** with the **authorization code grant + PKCE** for user-facing apps; **client credentials** for service-to-service. Avoid the deprecated implicit and password grants.
3. Validate JWTs at the gateway: signature against the issuer's JWKS, plus `iss`, `aud`, `exp`/`nbf`, and required `scope`/roles. Don't trust unvalidated claims.
4. Add mTLS for partner/B2B where the *application* must be provably trusted.
5. Apply the OWASP API Security Top 10 checklist (object/property/function-level authorization, unrestricted resource consumption, etc.) — see `references/api-security.md`.

### 4. Traffic management
- **Rate limit** to cap calls per app/user/route (protects against abuse and DoS). **Quota** is the longer-window business limit tied to a plan (e.g. 10k/day on the Gold plan). **Throttle/spike-arrest** smooths bursts so backends aren't overwhelmed. **Cache** responses (with correct TTL/invalidation) to cut backend load and latency. Details and gateway-specific knobs: `references/traffic-management.md`.

### 5. Lifecycle & versioning
- Run the loop: ideate → design → mock → build → contract-test → deploy → promote → deprecate → retire → observe → feed back. Announce deprecations with timelines, support old + new in parallel, then retire. See `references/lifecycle-and-versioning.md`.

### 6. Publish as a product
- Create a portal page (getting-started, auth, full resource reference with interactive try-it, errors, versioning/changelog, SLAs, T&Cs). Define plans (free/freemium/pay-per-call/tiered/revenue-share), wire billing, and onboard developers self-service. See `references/developer-portals-and-monetization.md`.

### 7. Observe & govern
- Monitor uptime/SLAs, throughput, success/error rates, latency. Use analytics to spot retirement candidates and DX problems. Establish standards (style guide, linting like Spectral), an operating model, and a platform team. See `references/governance-and-observability.md`.

## Common pitfalls & anti-patterns

- **Business logic in the gateway.** Conditionals, loops, complex transforms, orchestration belong in services. The gateway mediates and enforces policy.
- **Code-first with no contract.** Leads to inconsistent, undocumented APIs and consumer rework. Design-first + contract tests prevent drift.
- **Treating API keys as security.** Keys identify an app and are easily leaked; they are not authentication or authorization. Use OAuth2/OIDC + scopes.
- **Trusting a JWT without verifying it.** Always check signature, issuer, audience, and expiry at the edge. A decoded-but-unverified token is attacker-controlled input.
- **Breaking changes inside a version.** Adding required fields, removing fields, or changing types breaks consumers. Make additive changes; bump the major for breaking ones.
- **Over-using HATEOAS / chatty REST.** Forces clients into many round-trips. Consider composition, expansion params, or GraphQL for client-driven needs.
- **One central CoE as the only path to ship an API.** Becomes a bottleneck (Conway's law). Prefer a platform team offering self-service to federated product teams.
- **Shipping without docs/portal.** Undocumented APIs don't get adopted. Documentation is part of "done."
- **Ignoring rate limits/quotas until an incident.** Set sane defaults from day one; tie quotas to plans.
- **Versioning a GraphQL API like REST.** Evolve the schema (add fields, deprecate with `@deprecated`) instead of running parallel versions.

## Reference files

- `references/api-design.md` — REST resource modeling, status codes, pagination, error shapes; GraphQL schema design; gRPC/protobuf; OpenAPI 3.1 and AsyncAPI; the REST/GraphQL/gRPC decision table.
- `references/api-gateways.md` — gateway responsibilities, policy ordering, topologies (central, microgateway, sidecar), config-as-code, and how the major products map.
- `references/api-security.md` — OAuth2 grants & PKCE, OIDC, JWT validation, API keys, mTLS, scopes, CORS, OWASP API Security Top 10, threat protection.
- `references/traffic-management.md` — rate limiting algorithms, quotas vs rate limits, throttling/spike arrest, caching strategies, load balancing.
- `references/lifecycle-and-versioning.md` — the full API lifecycle, design-first loop, contract testing, versioning strategies, deprecation and retirement.
- `references/developer-portals-and-monetization.md` — portal/DX content, onboarding, API products, plans, billing, monetization models.
- `references/governance-and-observability.md` — analytics/KPIs, SLAs, governance standards & linting, target operating models, platform team roles.
