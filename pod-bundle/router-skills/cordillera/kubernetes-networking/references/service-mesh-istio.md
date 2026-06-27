# Service mesh fundamentals (Istio)

A practical introduction to Istio for traffic management, mTLS, and observability. Keep it
practical — Istio is huge; this covers the resources you'll actually author.

> **API versions:** Istio has stabilized its CRD groups. Use **`networking.istio.io/v1`**
> (VirtualService, DestinationRule, Gateway, ServiceEntry, Sidecar) and
> **`security.istio.io/v1`** (PeerAuthentication, RequestAuthentication, AuthorizationPolicy)
> on current Istio. Older `v1alpha3` / `v1beta1` examples you'll find in docs/books still apply
> structurally; just bump the version. Gateway/HTTPRoute here are the **Istio** mesh-routing
> objects — distinct from the upstream Gateway API in `ingress-and-gateway-api.md` (Istio can
> also implement that, but the objects below are the classic Istio mesh API).

Contents:
1. Control plane vs data plane
2. What the mesh buys you
3. Installing & enabling injection
4. Gateway + VirtualService (getting traffic in)
5. DestinationRule & traffic splitting (canary / blue-green)
6. mTLS with PeerAuthentication
7. RequestAuthentication + AuthorizationPolicy
8. ServiceEntry & Sidecar (egress scope)
9. Observability (Kiali / Prometheus / Jaeger)
10. Ambient mesh (the future)
11. Pitfalls

---

## 1. Control plane vs data plane

- **Control plane** = `istiod` (a single daemon, since Istio 1.5 — it bundles the old Pilot,
  Galley, Citadel). It does service discovery, config distribution to proxies, certificate
  issuance/rotation for mTLS, and sidecar injection.
- **Data plane** = an **Envoy proxy sidecar** injected next to each app container. It
  transparently intercepts all inbound/outbound pod traffic and applies routing, mTLS, retries,
  and telemetry — **no application code changes**.
- **istio-ingressgateway** = a dedicated Envoy at the mesh edge (the entry point for external
  traffic). **istio-egressgateway** = optional controlled exit point for outbound traffic.

---

## 2. What the mesh buys you (without writing code)

- **mTLS** between services: encryption + workload identity, certs managed by istiod.
- **Traffic management**: weighted splitting (canary/blue-green), retries, timeouts, fault
  injection (simulate delays/HTTP errors before prod).
- **Observability**: per-service metrics, distributed traces, a live traffic graph (Kiali).
- **Coarse authorization & request auth**: enforce JWT presence and group/claim checks at the
  proxy.

Decision: if you only need basic L7 HTTP routing, an ingress controller or Gateway API is far
simpler. Reach for a mesh when you need mTLS everywhere, fine traffic control, or deep
observability across many services.

---

## 3. Installing & enabling injection

```bash
# Install (demo profile = istiod + ingress + egress gateways; use 'default' for prod)
istioctl install --set profile=demo -y
istioctl verify-install

# Turn on automatic sidecar injection for a namespace
kubectl label namespace my-app istio-injection=enabled
# (re-deploy pods so they get the Envoy sidecar)
```

Profiles: `default` (istiod + ingressgateway), `demo` (adds egressgateway), `minimal` (istiod
only), `ambient` (istiod + CNI + ztunnel). After injection, every pod in the namespace runs an
Envoy sidecar.

---

## 4. Gateway + VirtualService (getting traffic in)

These two work together: the **Gateway** configures the ingressgateway Envoy to *accept*
traffic for a host/port/TLS; the **VirtualService** says *where* that traffic goes.

```yaml
apiVersion: networking.istio.io/v1
kind: Gateway
metadata:
  name: sales-gateway
  namespace: sales
spec:
  selector:
    istio: ingressgateway          # binds to the default istio-ingressgateway pod
  servers:
    - port:
        number: 443
        name: https
        protocol: HTTPS
      hosts:
        - sales.example.com
      tls:
        mode: SIMPLE               # SIMPLE=terminate | PASSTHROUGH=route by SNI | ISTIO_MUTUAL
        credentialName: sales-tls  # TLS Secret — must live in the ingressgateway's namespace (istio-system)
---
apiVersion: networking.istio.io/v1
kind: VirtualService
metadata:
  name: sales-vs
  namespace: sales
spec:
  hosts:
    - sales.example.com
  gateways:
    - sales-gateway
  http:
    - route:
        - destination:
            host: sales-frontend   # the target Kubernetes Service name
            port:
              number: 80
```

Notes:
- The Gateway's `credentialName` Secret must be in the **ingressgateway pod's namespace**
  (`istio-system` by default) — a common, hard-to-debug mistake (symptom: connection reset;
  istiod logs `failed to fetch key and certificate`).
- `tls.mode: PASSTHROUGH` skips decryption (needed e.g. for the Kubernetes API's SPDY-based
  `exec`/`port-forward`, which Envoy can't terminate) — but you lose L7 features for that route.
- In a `VirtualService`, `destination.host` is the **Service name**, not the external hostname.

---

## 5. DestinationRule & traffic splitting (canary / blue-green)

A `DestinationRule` defines **subsets** (by pod label) and the load-balancing policy; the
`VirtualService` references those subsets with **weights** to split traffic.

```yaml
apiVersion: networking.istio.io/v1
kind: DestinationRule
metadata:
  name: reviews-dr
  namespace: shop
spec:
  host: reviews                    # the Service
  trafficPolicy:
    loadBalancer:
      simple: ROUND_ROBIN          # or LEAST_CONN, RANDOM, etc.
  subsets:
    - name: v1
      labels:
        version: v1
    - name: v2
      labels:
        version: v2
---
apiVersion: networking.istio.io/v1
kind: VirtualService
metadata:
  name: reviews-vs
  namespace: shop
spec:
  hosts:
    - reviews
  http:
    - route:
        - destination:
            host: reviews
            subset: v1
          weight: 90               # canary: 90% stays on v1
        - destination:
            host: reviews
            subset: v2
          weight: 10               # 10% to v2
```

- **Canary**: start the new version at a small weight, watch metrics in Kiali, ramp up.
- **Blue/green**: keep both at full readiness; flip weights 100→0 / 0→100 to cut over.
- Sticky sessions for monoliths: `trafficPolicy.loadBalancer.consistentHash.httpCookie`
  (`ttl: 0s` = session cookie).
- mTLS to the backend from the rule: `trafficPolicy.tls.mode: ISTIO_MUTUAL`.

Istio also lets you **fault-inject** (HTTP `abort`/`delay`) in a VirtualService to test
resilience before shipping.

---

## 6. mTLS with PeerAuthentication

`PeerAuthentication` controls whether pods require mutual TLS. Scope it mesh-wide (create it in
`istio-system`) or per namespace.

```yaml
# Whole mesh: require mTLS
apiVersion: security.istio.io/v1
kind: PeerAuthentication
metadata:
  name: mesh-mtls
  namespace: istio-system
spec:
  mtls:
    mode: STRICT          # STRICT = reject plaintext | PERMISSIVE = accept both
```

```yaml
# One namespace only
apiVersion: security.istio.io/v1
kind: PeerAuthentication
metadata:
  name: sales-mtls
  namespace: sales
spec:
  mtls:
    mode: STRICT
```

- `STRICT`: plaintext is denied — only mutually-authenticated traffic allowed.
- `PERMISSIVE`: encrypted+identified traffic is used when offered, plaintext still accepted —
  useful during migration into the mesh.
- Common pattern: keep the mesh `PERMISSIVE` and set `STRICT` per sensitive namespace.

---

## 7. RequestAuthentication + AuthorizationPolicy

Two layers of *end-user* auth (distinct from pod-to-pod mTLS above).

**RequestAuthentication** — validates JWTs (it does NOT, by itself, require one):

```yaml
apiVersion: security.istio.io/v1
kind: RequestAuthentication
metadata:
  name: jwt-auth
  namespace: shop
spec:
  selector:
    matchLabels:
      app: frontend
  jwtRules:
    - issuer: "https://issuer.example.com/auth"
      jwksUri: "https://issuer.example.com/auth/certs"
      outputPayloadToHeader: User-Info   # passes claims (base64 JSON) to the app
```

Behavior: an **invalid** token → denied; a **valid** token → allowed; **no** token → still
allowed. To actually *require* a token you must add an AuthorizationPolicy.

**AuthorizationPolicy** — the access decision. Evaluation order: CUSTOM-deny → DENY → (if no
ALLOW matches → deny) → ALLOW. Key rule: **once any policy selects a workload, the default flips
to deny** for non-matching requests (implicit enablement).

```yaml
# Require a valid JWT (a requestPrincipal must exist)
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: require-jwt
  namespace: shop
spec:
  selector:
    matchLabels:
      app: frontend
  action: ALLOW
  rules:
    - from:
        - source:
            requestPrincipals: ["*"]   # "*" = any authenticated user (anonymous excluded)
```

```yaml
# Coarse-grained: only allow GET, only from a group claim
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: frontend-az
  namespace: shop
spec:
  selector:
    matchLabels:
      app: frontend
  action: ALLOW
  rules:
    - to:
        - operation:
            methods: ["GET"]
      when:
        - key: request.auth.claims[groups]
          values: ["cn=group2,ou=Groups,DC=domain,DC=com"]
```

Other handy fields: `from.source.ipBlocks` (restrict client CIDRs), `from.source.principals`
(the *calling service* identity, e.g. the ingressgateway). An empty `spec: {}` AuthorizationPolicy
= **deny all** in the namespace; `spec.rules: - {}` = **allow all**.

**Boundary:** keep mesh-layer authz **coarse-grained** ("is this user allowed to call this
service at all?"). Push **fine-grained entitlements** ("can this user write *this* check?")
into the service code or a dedicated engine (OPA/Cedar/OpenFGA) — entitlement data usually lives
in business databases the cluster shouldn't own.

---

## 8. ServiceEntry & Sidecar (egress scope)

**ServiceEntry** — register an external host so in-mesh workloads can reach it (otherwise the
sidecar can't resolve it):

```yaml
apiVersion: networking.istio.io/v1
kind: ServiceEntry
metadata:
  name: external-api
  namespace: sales
spec:
  hosts:
    - api.partner.com
  ports:
    - number: 443
      name: https
      protocol: TLS
  resolution: DNS
  location: MESH_EXTERNAL
```

**Sidecar** — limit which services each Envoy tracks. By default every sidecar knows about every
service in the mesh; in large clusters this wastes memory and can OOM/crashloop the proxy. A
`Sidecar` object scopes egress to just what's needed:

```yaml
apiVersion: networking.istio.io/v1
kind: Sidecar
metadata:
  name: sales-sidecar
  namespace: sales
spec:
  egress:
    - hosts:
        - "./*"            # services in this namespace
        - "istio-system/*" # always include the Istio control-plane namespace
        - "shared/*"       # plus one other namespace this app calls
```

---

## 9. Observability (Kiali / Prometheus / Jaeger)

Standard add-ons that complete the mesh:
- **Prometheus** — scrapes Envoy/istiod metrics.
- **Jaeger** — distributed tracing (the execution path across services).
- **Grafana** — dashboards.
- **Kiali** — the topology/graph UI: live traffic animation, per-edge mTLS lock icons and RPS,
  request volume/latency, container logs, and an Istio-config view (with object-creation
  wizards). The single best tool for "why is this request slow / where did it fail?".

---

## 10. Ambient mesh (the future)

Ambient mode removes the **per-pod sidecar**. Node-level **ztunnel** handles L4 (mTLS, identity)
and optional per-namespace **waypoint** proxies handle L7 (routing, authz). Benefits: much lower
CPU/RAM at scale, simpler ops, no pod restarts to join the mesh. Maturing toward GA — check
current Istio status before adopting; the sidecar mode above is the stable baseline.

---

## 11. Pitfalls

- **Gateway TLS Secret in the wrong namespace** — must be in the ingressgateway's namespace
  (`istio-system`), not the app namespace. Symptom: connection reset; istiod logs a fetch error.
- **RequestAuthentication without an AuthorizationPolicy** — *no token* requests are **allowed**.
  Always pair with a `requestPrincipals: ["*"]` ALLOW policy to actually require a JWT.
- **AuthorizationPolicy implicit-deny surprise** — denying one port/method doesn't only deny
  that; once a policy selects the workload, everything not explicitly allowed is denied. Add the
  ALLOW rules you need.
- **No `Sidecar` scoping in a big mesh** — Envoy tracks all services, ballooning memory → OOM /
  silent crashloops. Scope egress per namespace.
- **Forgetting injection** — without the `istio-injection=enabled` label (and a pod restart),
  pods have no sidecar and none of the mesh features apply.
- **PASSTHROUGH everywhere** — you lose routing, telemetry, and L7 authz; only use it where
  termination is genuinely impossible.
