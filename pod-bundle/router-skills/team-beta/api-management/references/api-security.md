# API Security

Table of contents:
1. AuthN vs AuthZ vs identity
2. API keys
3. OAuth2 grants (and what's deprecated)
4. OIDC
5. JWT validation at the gateway
6. Scopes and fine-grained authorization
7. mTLS
8. CORS
9. The bearer-token and obscure-token patterns
10. OWASP API Security Top 10 + threat protection

---

## 1. AuthN vs AuthZ vs identity

- **Authentication (AuthN)** — verify *who* the caller is (credentials, certificate, token). Must happen before authorization.
- **Authorization (AuthZ)** — verify the (authenticated) caller is *allowed* to do this specific thing on this specific resource.
- **Identity protocols vs authorization protocols**: SAML and OIDC are primarily *authentication* protocols (SSO — prove who a user is). OAuth2 is an *authorization* protocol — it issues access tokens (carrying grants/scopes) that API consumers present to resource servers. A common mistake is using an OIDC `id_token` to call APIs; call APIs with the OAuth2 **access token**.

---

## 2. API keys

An API key identifies the **application/project**, not a user. It's a coarse identifier for analytics, rate-limiting buckets, and plan/quota enforcement.

- A key is **not authentication and not authorization** — it's a shared secret that's easily leaked (logs, client bundles, repos). Never rely on it alone for anything sensitive.
- Send via header (`X-API-Key` or `apikey`), never in the URL/query string (leaks into logs, history, referrers).
- Support rotation and per-key revocation; scope keys to environments and plans.
- Combine with OAuth2 for real user/app auth when access matters.

---

## 3. OAuth2 grants

OAuth2 defines flows for obtaining an access token. Current guidance (OAuth 2.1 consolidates these best practices):

- **Authorization Code + PKCE** — the default for everything user-facing: web apps, SPAs, mobile, native. PKCE (`code_verifier`/`code_challenge`) protects the code exchange and is now required for all clients, public and confidential.
- **Client Credentials** — service-to-service / machine-to-machine, no user present. The app authenticates with its own credentials and gets a token for its own access.
- **Refresh tokens** — to get new access tokens without re-prompting; rotate them and bind to the client.
- **Device Authorization Grant** — for input-constrained devices (TVs, CLIs).

**Deprecated / avoid:**
- **Implicit grant** — tokens in the URL fragment; removed in OAuth 2.1. Use Auth Code + PKCE instead.
- **Resource Owner Password Credentials (ROPC)** — app handles the user's password directly; removed in 2.1. Only ever for trusted first-party migration, never for third parties.

Roles: *resource owner* (user), *client* (the app), *authorization server* (issues tokens), *resource server* (the API/gateway that validates them).

---

## 4. OIDC

OpenID Connect is an identity layer on top of OAuth2. The authorization server additionally returns an **`id_token`** (a JWT with user identity claims) and exposes a `/userinfo` endpoint and discovery document (`/.well-known/openid-configuration`) and a **JWKS** endpoint for signing keys. Use OIDC when you need to authenticate the *user* (login/SSO) in addition to authorizing API access. Keep the roles straight: `id_token` for "who is the user", access token for "call the API".

---

## 5. JWT validation at the gateway

A JWT presented as a bearer token MUST be fully validated before trusting any claim. Decoding is not validating — an unverified token is attacker-controlled input.

Validate, in order:
1. **Signature** — verify against the issuer's public key from its **JWKS** (`/.well-known/jwks.json`), matching the token's `kid`. Cache keys; refresh on rotation.
2. **Algorithm** — pin expected algs (e.g. `RS256`/`ES256`). **Reject `alg: none`** and reject HMAC algs when you expect asymmetric (the classic alg-confusion attack).
3. **`iss`** — exactly matches your trusted authorization server.
4. **`aud`** — contains this API's identifier (a token minted for another API must not work here).
5. **`exp` / `nbf` / `iat`** — not expired, not before, sane issue time (allow small clock skew).
6. **Scopes / roles** — required scope present for this route (see §6).

Gateway-native validators: Apigee `VerifyJWT`, Kong `jwt`/`openid-connect`, Azure APIM `validate-jwt`, AWS JWT authorizer. Prefer these over rolling your own. Keep access tokens short-lived; use introspection (RFC 7662) for opaque tokens or where instant revocation matters.

```
# APIM validate-jwt sketch
<validate-jwt header-name="Authorization" failed-validation-httpcode="401">
  <openid-config url="https://auth.example.com/.well-known/openid-configuration"/>
  <audiences><audience>orders-api</audience></audiences>
  <issuers><issuer>https://auth.example.com/</issuer></issuers>
  <required-claims>
    <claim name="scope" match="any"><value>orders:read</value></claim>
  </required-claims>
</validate-jwt>
```

---

## 6. Scopes and fine-grained authorization

- **Scopes** are coarse, consent-style permissions on the token (`orders:read`, `orders:write`). Check them at the gateway per route/method.
- **Fine-grained authorization** (this user may read *this* order, but not that customer's) is *resource-level* and usually can't be fully decided at the gateway — enforce it in the service, or with a policy engine (OPA/Rego, Cedar, OpenFGA/Zanzibar-style). The OWASP "broken object-level authorization" (BOLA) risk lives here: never assume a valid token implies the right to a specific object — check ownership/tenant on every access.

---

## 7. mTLS

Mutual TLS authenticates the **client application** at the transport layer: both sides present and verify certificates. Use it for partner/B2B and high-assurance internal traffic where you must prove the calling *app* is trusted (independently of any user token). Often combined with bearer tokens: mTLS proves the app, the token authorizes the action. In Kubernetes, a service mesh can provide mTLS for east-west traffic automatically.

---

## 8. CORS

CORS is a **browser** mechanism, not a server-side security control — it does not protect your API, it only relaxes the browser's same-origin policy for trusted web origins. Configure it deliberately at the gateway:
- Set `Access-Control-Allow-Origin` to an explicit allow-list, **never `*` when credentials are involved**.
- Handle `OPTIONS` preflight; declare allowed methods/headers; set a sane `Access-Control-Max-Age`.
- CORS being permissive is not a vulnerability by itself, but a wildcard + credentials combination is a real one.

---

## 9. Bearer-token and obscure-token patterns

**Bearer of token** (OAuth2): the authorization server authenticates the user/app and issues a token (typically a JWT). The client sends `Authorization: Bearer <token>`; the gateway (resource server) verifies the signature against the AS's signing cert and checks scopes, then allows access. Add mTLS so only trusted client apps are processed. This is the standard pattern for any HTTP/HTTP2 API where the gateway supports OAuth2.

**Bearer of obscure token**: for cases where the client must *never* see token contents (untrusted app, strict compliance). The AS issues an opaque random string instead of a JWT; the gateway introspects it with the AS to exchange it for the real access token, then validates that. Layer mTLS on every hop (client↔AS, client↔gateway, gateway↔AS) and use the auth-code grant with no client secret exposed. More secure but considerably more complex — only when the requirement demands it.

---

## 10. OWASP API Security Top 10 + threat protection

Run every public API against the **OWASP API Security Top 10 (2023)**:

1. **Broken Object Level Authorization (BOLA)** — enforce per-object ownership/tenant checks; don't trust IDs in the request.
2. **Broken Authentication** — strong token validation, no weak/missing auth, protect token endpoints.
3. **Broken Object Property Level Authorization** — don't mass-assign or over-expose properties; filter response fields by role.
4. **Unrestricted Resource Consumption** — rate limits, quotas, payload size limits, query cost limits (esp. GraphQL); protects against DoS and cost-blowup.
5. **Broken Function Level Authorization** — check the caller may invoke *this operation* (admin vs user).
6. **Unrestricted Access to Sensitive Business Flows** — protect flows (signup, checkout) from automated abuse.
7. **Server Side Request Forgery (SSRF)** — validate/allow-list any URLs the API fetches.
8. **Security Misconfiguration** — TLS everywhere, no verbose errors/stack traces, security headers, locked-down CORS.
9. **Improper Inventory Management** — know every API and version (esp. old/shadow/staging endpoints); retire deprecated ones.
10. **Unsafe Consumption of APIs** — validate and sanitize data you receive from upstream/third-party APIs too.

**Gateway threat protection** to apply: payload size caps, JSON/XML depth/entity-expansion limits, schema validation against the OpenAPI contract, regex/SQLi/XSS pattern checks, IP reputation/allow-deny, and a WAF as the outermost layer (the "API firewall" pattern — first line of defense before requests even reach the gateway). Always redact secrets/PII from logs.
