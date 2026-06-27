# Ingress (nginx) and the Gateway API

Contents:
1. When to use which
2. Ingress: anatomy & name/path routing
3. Ingress TLS termination
4. nginx annotation catalog
5. IngressClass & multiple controllers
6. Installing the nginx ingress controller
7. Gateway API: GatewayClass / Gateway / HTTPRoute
8. Gateway API TLS
9. Gateway API traffic splitting & cross-namespace
10. Why Gateway API supersedes Ingress + migration

---

## 1. When to use which

Both are **L7** (HTTP/HTTPS) entry points: host- and path-based routing, TLS termination,
one IP for many domains — things an L4 `LoadBalancer` Service cannot do. An ingress
controller (or Gateway) is itself exposed via one `LoadBalancer`/NodePort, and many routes
share it.

- **Ingress** (`networking.k8s.io/v1`): ubiquitous, simple, but **feature-frozen** — no new
  capabilities will land. Fine for existing setups and simple host/path routing.
- **Gateway API** (`gateway.networking.k8s.io/v1`): the successor. More expressive
  (cross-namespace routing, role separation, richer matching, header/weight rules), and it's
  where the ecosystem is heading. **Prefer it for new designs.**

The L7 controller is just a Deployment/DaemonSet of proxy pods (nginx, Envoy, Traefik,
HAProxy) that watch Ingress/Gateway objects and reconfigure themselves. Without a controller
installed, the objects do nothing.

---

## 2. Ingress: anatomy & name/path routing

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: portal-ingress
  namespace: web
spec:
  ingressClassName: nginx          # modern way to bind to a controller
  rules:
    - host: shop.example.com       # name-based virtual host; omit to match all hosts
      http:
        paths:
          - path: /video
            pathType: Prefix        # Prefix | Exact | ImplementationSpecific
            backend:
              service:
                name: video-service
                port:
                  number: 8080
          - path: /shopping
            pathType: Prefix
            backend:
              service:
                name: shopping-service
                port:
                  number: 8080
          - path: /                 # catch-all → default app
            pathType: Prefix
            backend:
              service:
                name: blog-service
                port:
                  number: 8080
```

- `pathType` is required: `Prefix` (path tree), `Exact` (exact match), or
  `ImplementationSpecific` (controller-defined). Always set it explicitly.
- Multiple `host` entries enable name-based virtual hosting on one IP.
- The controller learns the requested host from the HTTP `Host` header (HTTP) or **SNI** in
  the TLS handshake (HTTPS) — that's how it routes before/without decrypting.

How a request flows: client → DNS resolves the host to the LB/ingress IP → LB forwards to an
ingress controller pod → controller matches host+path against rules → forwards to the chosen
Service's endpoints. The backend Service is normally a plain `ClusterIP`.

---

## 3. Ingress TLS termination

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: secure-ingress
  namespace: web
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - shop.example.com
      secretName: shop-tls          # a kubernetes.io/tls Secret in THIS namespace
  rules:
    - host: shop.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: shop-service
                port:
                  number: 80
```

The Secret must be type `kubernetes.io/tls` (keys `tls.crt`, `tls.key`) and live in the **same
namespace as the Ingress** (note: this differs from Gateway API and Istio, where the cert lives
with the gateway/listener). For automated issuance/renewal use **cert-manager** with an ACME
(Let's Encrypt) `ClusterIssuer`. The controller serves the correct cert per host via SNI.

---

## 4. nginx annotation catalog

Controller-specific behavior is configured via `nginx.ingress.kubernetes.io/*` annotations:

| Annotation | Effect |
|---|---|
| `rewrite-target: /` | rewrite matched path before forwarding (common with `/foo(/|$)(.*)` regex paths) |
| `ssl-redirect: "true"` | force HTTP→HTTPS redirect |
| `force-ssl-redirect: "true"` | redirect even without a TLS block |
| `backend-protocol: "HTTPS"` | speak HTTPS to the backend (re-encrypt) |
| `affinity: "cookie"` + `session-cookie-name` | cookie-based sticky sessions |
| `ssl-passthrough: "true"` | TLS passthrough — don't terminate, route by SNI (see below) |
| `proxy-body-size: "50m"` | raise upload size limit |
| `proxy-read-timeout: "120"` | backend read timeout (seconds) |
| `whitelist-source-range: "10.0.0.0/8"` | restrict client CIDRs |

**TLS passthrough** turns the L7 controller into a pseudo-L4 forwarder: instead of decrypting,
it routes purely by the TLS SNI hostname. This lets you expose non-HTTP-but-TLS backends (e.g. a
database) through the ingress — but you lose all L7 features (path routing, sticky cookies,
header rewrites) for that route, and it can be a security footgun. Prefer terminating TLS at the
ingress when you can.

---

## 5. IngressClass & multiple controllers

Run more than one controller (e.g. a public and an internal one) and bind each Ingress with
`ingressClassName`:

```yaml
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: nginx
  annotations:
    ingressclass.kubernetes.io/is-default-class: "true"   # used when Ingress omits the class
spec:
  controller: k8s.io/ingress-nginx
```

`ingressClassName` on the Ingress replaces the deprecated `kubernetes.io/ingress.class`
annotation. If exactly one IngressClass is marked default, Ingresses without a class get it
automatically. Scenarios: `nginx-public` vs `nginx-private`, or separate controllers per
protocol.

---

## 6. Installing the nginx ingress controller

```bash
# Cloud (provisions a LoadBalancer Service for the controller)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.2/deploy/static/provider/cloud/deploy.yaml

# AWS NLB variant
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.2/deploy/static/provider/aws/deploy.yaml

# minikube
minikube addons enable ingress

# verify
kubectl get pods -n ingress-nginx
kubectl get ingressclass
```

(Always check the docs for the current controller version.) Helm is the preferred method for
production. The controller runs in `ingress-nginx` and is exposed as a `LoadBalancer` Service
on cloud. Some dedicated cloud controllers (e.g. Azure AGIC, AWS ALB controller) talk
**directly to pod IPs** via the cloud SDN, removing a load-balancing hop — bind those with the
matching `ingressClassName` (e.g. `azure-application-gateway`).

---

## 7. Gateway API: GatewayClass / Gateway / HTTPRoute

Three roles, three objects:

- **GatewayClass** — the controller/infra type (like StorageClass). Cluster-scoped, created by
  the platform.
- **Gateway** — a concrete listener (ports, protocols, TLS). Owned by cluster/infra ops.
- **HTTPRoute** (also TCPRoute, GRPCRoute…) — the routing rules. Owned by app teams, attached
  to a Gateway via `parentRefs`.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: prod-gateway-class
spec:
  controllerName: example.net/gateway-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: prod-gateway
  namespace: infra
spec:
  gatewayClassName: prod-gateway-class
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All            # let HTTPRoutes in any namespace attach
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: shop-route
  namespace: web
spec:
  parentRefs:
    - name: prod-gateway
      namespace: infra         # attach to a Gateway in another namespace
  hostnames:
    - "shop.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /video
      backendRefs:
        - name: video-service
          port: 8080
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: blog-service
          port: 8080
```

Match types include `PathPrefix`/`Exact`/`RegularExpression`, plus header, query-param, and
method matching — richer than Ingress.

---

## 8. Gateway API TLS

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: prod-gateway
  namespace: infra
spec:
  gatewayClassName: prod-gateway-class
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate                # Terminate | Passthrough
        certificateRefs:
          - kind: Secret
            name: shop-tls              # kubernetes.io/tls Secret in the Gateway's namespace
```

TLS config lives on the **Gateway listener** (infra-owned), not on the route — a cleaner
separation than Ingress, where the cert sits beside each Ingress object.

---

## 9. Gateway API traffic splitting & cross-namespace

Weighted backends are first-class (no controller-specific annotations needed):

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: canary-route
  namespace: web
spec:
  parentRefs:
    - name: prod-gateway
      namespace: infra
  hostnames:
    - "app.example.com"
  rules:
    - backendRefs:
        - name: app-v1
          port: 80
          weight: 90
        - name: app-v2
          port: 80
          weight: 10        # 10% canary to v2
```

Cross-namespace routing (a route in `web` attaching to a Gateway in `infra`) is controlled by
the Gateway's `allowedRoutes` and, when referencing backends across namespaces, a
`ReferenceGrant` in the target namespace — something Ingress simply cannot express.

---

## 10. Why Gateway API supersedes Ingress + migration

- **Role separation**: GatewayClass (provider) / Gateway (infra) / HTTPRoute (app) maps to
  real org boundaries; multiple teams safely share one Gateway.
- **Expressiveness**: header/method/query matching, native weighted splitting, typed routes
  for non-HTTP (TCP/GRPC), portable TLS config.
- **Extensibility/portability**: a standard API across implementations (nginx, Envoy, Istio,
  cloud LBs) instead of per-controller annotation dialects.
- **Ingress is feature-frozen** — only bug fixes, no new features.

Migration is a **one-time conversion**: Gateway API has no backward-compatible Ingress
resource. The `ingress2gateway` CLI (sigs.k8s.io) translates existing Ingress objects into
Gateway/HTTPRoute as a starting point; review the output before applying.
