---
name: kubernetes-networking
description: >-
  Design and debug Kubernetes networking. Use for Service types (ClusterIP, NodePort,
  LoadBalancer, ExternalName, headless), kube-proxy and EndpointSlices, CoreDNS service
  discovery and the ndots gotcha, Ingress and ingress controllers with TLS, the Gateway API
  (GatewayClass, Gateway, HTTPRoute) that supersedes Ingress, NetworkPolicies (default-deny,
  ingress/egress, pod and namespace selectors), bare-metal load balancing with MetalLB,
  external-dns, multi-cluster traffic, and Istio service mesh (sidecars,
  VirtualService/DestinationRule, mTLS, traffic splitting). Trigger whenever the user exposes
  a Pod or app, cannot reach a Service, debugs DNS or connection-refused/timeout between pods,
  sets up Ingress or Gateway routes, writes NetworkPolicies, configures a LoadBalancer IP on
  bare metal, or works with a service mesh - even if they do not say Kubernetes. For
  NetworkPolicy security strategy see kubernetes-security-rbac; for CNI install see
  kubernetes-cluster-operations.
---

# Kubernetes Networking

Equips Claude to design, author, and debug the full stack of Kubernetes traffic routing:
the pod network model, Services, DNS, Ingress, Gateway API, NetworkPolicies, bare-metal
load balancing, multi-cluster DNS failover, and the Istio service mesh.

## When to use this skill

- Writing or reviewing a Service, Ingress, `HTTPRoute`/`Gateway`, or `NetworkPolicy` manifest.
- Choosing a Service type or deciding Ingress vs. Gateway API vs. LoadBalancer.
- Debugging connectivity: "service has no endpoints", `EXTERNAL-IP` stuck `<pending>`,
  intermittent 503s, DNS lookups failing or slow, a NetworkPolicy that blocks too much.
- Exposing apps externally: nginx ingress + TLS, MetalLB on bare metal, external-dns
  for automatic DNS records, multi-cluster GSLB/failover.
- Service mesh work: injecting sidecars, splitting traffic across versions, enforcing mTLS,
  routing with `VirtualService`/`DestinationRule`.

Boundaries — cross-link, don't duplicate:
- **NetworkPolicy *strategy*** (zero-trust posture, tenant isolation design) → `kubernetes-security-rbac`.
  This skill owns the *mechanics* (selectors, rule shapes, default-deny).
- **Installing/troubleshooting the CNI plugin itself** (Calico/Cilium pod-to-pod breakage,
  IPAM, kernel issues) → `kubernetes-cluster-operations`.
- **HorizontalPodAutoscaler / scaling** → the autoscaling skill.

## Core concepts (the mental model)

**The flat pod network.** Every pod gets its own routable IP from the CNI plugin. The
Kubernetes model mandates: every pod can reach every other pod **without NAT**, and a
pod sees its own IP the same way others see it. There are four communication paths:
container↔container (via `localhost` inside a pod), pod↔pod (flat network), pod↔Service
(stable virtual IP), and external↔Service (NodePort/LB/Ingress).

**Pod IPs are ephemeral — never hardcode them.** A deleted-and-recreated pod gets a new
IP. The whole reason Services exist is to provide a stable name/VIP in front of a churning
set of pods. The same applies to picking endpoints: Services select pods by **label**, not
by IP.

**Services select pods by label selector → produce EndpointSlices.** A Service with a
`selector` continuously watches for pods whose labels match, and the matching pod IPs become
its endpoints (stored in `EndpointSlice` objects, the scalable successor to the monolithic
`Endpoints` object). `kubectl get endpointslices -l kubernetes.io/service-name=<svc>` (or
the legacy `kubectl get endpoints <svc>`) is the single most useful debugging command — **no
endpoints means no ready pods match the selector.**

**kube-proxy implements the Service VIP.** A Service is a logical object, not a process.
On every node, kube-proxy programs forwarding rules so traffic to the ClusterIP is
load-balanced to a backing pod. Two common modes:
- `iptables` (default in many distros): random/hash-based distribution.
- `ipvs`: true round-robin plus more algorithms (least-conn, source-hash) and better
  performance at scale.
Round-robin is *not* guaranteed in iptables mode — don't promise even distribution.

**Readiness gates Service membership.** A pod only receives Service traffic once its
`readinessProbe` passes. This is the correct fix for "traffic hit my pod before it was
ready" — see `references/services.md`.

**Service DNS.** CoreDNS gives every Service an A record at
`<service>.<namespace>.svc.cluster.local`. Same namespace → short name `my-svc`; cross
namespace → `my-svc.other-ns` or the FQDN. The **ndots:5 gotcha** (slow/failing external
lookups) lives in `references/dns-and-discovery.md`.

**Service types — pick by reachability need:**

| Type | Reachable from | Use for |
|------|----------------|---------|
| `ClusterIP` (default) | inside cluster only | pod-to-pod / internal APIs, DB backends |
| `NodePort` | `<nodeIP>:30000-32767` | dev/test, behind an external LB, quick bypass-the-ingress test |
| `LoadBalancer` | external IP (cloud or MetalLB) | external L4 exposure, non-HTTP (TCP/UDP) workloads |
| `ExternalName` | CNAME to external DNS | aliasing an out-of-cluster host (DB, SaaS) |
| headless (`clusterIP: None`) | per-pod DNS, no VIP | StatefulSets, clients that do their own LB |

**L4 vs L7 — the dividing line for "how do I expose HTTP?"** A `LoadBalancer` Service is
L4: it forwards TCP/UDP and is blind to HTTP. It cannot do host/path routing, TLS
termination, or one IP for many domains. For HTTP(S) you want **L7**: an Ingress controller
or the Gateway API. Use a raw `LoadBalancer` for non-HTTP protocols (databases, DNS, gRPC
streams) or as the single entry point *in front of* an ingress controller.

**Ingress is feature-frozen; Gateway API supersedes it.** The Ingress API receives no new
features. New designs should prefer the **Gateway API** (`gateway.networking.k8s.io/v1`),
which splits responsibilities cleanly: `GatewayClass` (the controller/infra class),
`Gateway` (a listener: ports/protocols/TLS, owned by cluster/infra ops), and `HTTPRoute`
(routing rules, owned by app teams) — enabling cross-namespace and role-oriented routing
that Ingress can't express. Ingress is still everywhere in the wild, so know both.

**NetworkPolicy is default-allow until the first policy selects a pod.** With no policies,
all pods can talk to all pods (and across namespaces). The moment a NetworkPolicy selects a
pod for a direction (ingress/egress), everything *not* explicitly allowed for that direction
is denied. This skill owns the mechanics; see `references/network-policies.md`.

## Workflow / how to approach networking tasks

### Exposing an application — decide the layer first
1. **Internal-only** (other pods call it)? → `ClusterIP`. Done. Reference by DNS name.
2. **External, HTTP/HTTPS, name- or path-based routing, TLS?** → L7. Deploy an **nginx
   Ingress controller** (or another L7 controller) and write an `Ingress`, or prefer the
   **Gateway API** for new work. The controller itself is exposed via one `LoadBalancer`
   Service (cloud) or MetalLB (bare metal) — many app routes share that single entry point.
3. **External, non-HTTP (TCP/UDP), or you need raw L4?** → `LoadBalancer` Service. On bare
   metal this needs **MetalLB** (or similar) or `EXTERNAL-IP` stays `<pending>` forever.
4. **Aliasing an external host** so in-cluster clients use a stable name? → `ExternalName`.

Always pair a freshly exposed Deployment with a `readinessProbe` so the Service never routes
to a not-yet-ready pod.

### Authoring a Service (always declarative)
- Match `selector` to the pod template's labels **exactly** — this is the #1 source of
  "no endpoints".
- Distinguish `port` (the Service's own port), `targetPort` (the container port), and for
  NodePort, `nodePort` (30000–32767). They're independent; conventionally `port==targetPort`.
- Avoid `kubectl expose --port=...`: it hides options and always defaults to ClusterIP. Use
  YAML. Use `kubectl port-forward` only for *testing*, never as a production exposure path.
- Multi-protocol (TCP+UDP same Service, e.g. DNS): supported on `LoadBalancer` since the
  `MixedProtocolLBService` gate went GA in 1.26 — just give each port a unique `name`.

### Authoring an Ingress (nginx)
- `apiVersion: networking.k8s.io/v1`. Use `ingressClassName` (the modern way) instead of the
  deprecated `kubernetes.io/ingress.class` annotation.
- Set `pathType` explicitly (`Prefix` or `Exact`). Use `host:` for name-based virtual hosting.
- TLS termination: reference a `kubernetes.io/tls` Secret in `spec.tls`. For automated certs,
  use cert-manager. nginx serves the right cert for HTTPS via SNI.
- Controller-specific behavior lives in annotations (`nginx.ingress.kubernetes.io/...`):
  rewrite-target, ssl-redirect, sticky sessions, TLS passthrough. Full catalog in
  `references/ingress-and-gateway-api.md`.

### Migrating to / authoring Gateway API
- Three objects: `GatewayClass` → `Gateway` → `HTTPRoute`. App teams write only `HTTPRoute`s
  and attach them to a shared `Gateway` via `parentRefs`. Full manifests in
  `references/ingress-and-gateway-api.md`.
- Gateway API has no Ingress backward-compat resource — migration is a one-time conversion
  (the `ingress2gateway` tool helps).

### Writing a NetworkPolicy
1. Confirm the **CNI enforces policies** (Calico/Cilium do; flannel alone does not) — else the
   policy is silently inert.
2. Start from a namespace **default-deny** (`podSelector: {}`, both `policyTypes`), then add
   narrow allow rules. Remember each rule's `from`/`to` peer combines `podSelector` +
   `namespaceSelector` as **AND** when in the same list element, but separate list elements
   are **OR** — a frequent source of overly-broad or overly-narrow rules. See
   `references/network-policies.md`.
3. Omitting a `policyType` leaves that direction wide open. If you only define `Ingress`,
   all egress is still allowed.

### Debugging connectivity (ordered triage)
1. `kubectl get endpointslices -l kubernetes.io/service-name=<svc>` — empty? selector
   mismatch or no ready pods. Fix labels or the readinessProbe.
2. `kubectl run tmp --rm -it --image nicolaka/netshoot -- bash` then `nslookup <svc>`,
   `curl <svc>.<ns>` from *inside* the cluster (ClusterIP/cluster DNS aren't reachable from
   your laptop).
3. `EXTERNAL-IP <pending>` on a LoadBalancer → no LB provider (bare metal needs MetalLB).
4. Ingress 404 → host/path/`ingressClassName` mismatch, or the backend Service has no
   endpoints. 503 → backend pods unready/crashing.
5. Connection works without a policy but fails after one → NetworkPolicy too restrictive;
   check label selectors and that you allowed the *return* path / DNS egress (UDP 53 to
   kube-dns) when using default-deny egress.
6. Slow/failing external DNS from pods → the `ndots:5` problem (`references/dns-and-discovery.md`).

### Service mesh (Istio) tasks
- Enable injection per namespace: `kubectl label ns <ns> istio-injection=enabled`; each pod
  then gets an Envoy sidecar — no app code changes.
- Route external traffic with `Gateway` + `VirtualService` (mesh entry), split traffic
  across versions with `DestinationRule` subsets + `VirtualService` weights, enforce mTLS
  with `PeerAuthentication` (`STRICT`/`PERMISSIVE`). Full manifests and patterns in
  `references/service-mesh-istio.md`.
- Keep mesh-layer authorization **coarse-grained** (is-this-user-allowed-to-call-the-service);
  push business/entitlement decisions into the service. Detail in the Istio reference.

## Common pitfalls & anti-patterns

- **Hardcoding pod IPs** in config/code — they change on every reschedule. Use a Service DNS name.
- **Selector ≠ pod labels** — the Service exists but has zero endpoints and silently drops
  traffic. Verify with `kubectl get endpointslices`.
- **Expecting a ClusterIP/NodePort to be reachable from your laptop** — ClusterIP is
  in-cluster only; test from a pod or via `port-forward`.
- **Using `LoadBalancer` per microservice** — each provisions (and bills for) a separate cloud
  LB. Prefer one ingress/Gateway entry point fronting many routes.
- **`EXTERNAL-IP <pending>` on bare metal** — there's no built-in LB; install MetalLB.
- **Trying host/path routing or TLS termination on an L4 LoadBalancer** — impossible; that's
  an L7 (Ingress/Gateway) job.
- **`spec.loadBalancerIP`** — deprecated since 1.24; use the LB implementation's mechanism
  (e.g. MetalLB's `metallb.universe.tf/loadBalancerIPs` annotation) for a static IP.
- **NetworkPolicy with the wrong CNI** — flannel won't enforce it; the policy is a no-op.
- **Default-deny egress without allowing DNS** — pods can't resolve anything; always allow
  egress UDP/TCP 53 to kube-dns.
- **New Ingress designs** — prefer Gateway API; Ingress is feature-frozen.
- **`kubectl expose` / `port-forward` in production** — port-forward is a dev tool tied to
  your terminal; `expose` hides Service options. Use declarative YAML.

## Reference files

- **`references/services.md`** — every Service type with complete YAML, headless/StatefulSet
  DNS, sessionAffinity, multi-port, multi-protocol, readiness/liveness/startup probes, the
  port/targetPort/nodePort distinction, EndpointSlices internals.
- **`references/ingress-and-gateway-api.md`** — nginx Ingress manifests, TLS termination,
  the key annotation catalog, IngressClass & multiple controllers; full Gateway API
  (`GatewayClass`/`Gateway`/`HTTPRoute`) with TLS, traffic splitting, cross-namespace, and
  the Ingress→Gateway migration story.
- **`references/network-policies.md`** — default-deny recipes, ingress/egress rule shapes,
  the podSelector/namespaceSelector AND-vs-OR semantics, ipBlock/CIDR, allow-DNS patterns.
  (Strategy/posture → `kubernetes-security-rbac`.)
- **`references/dns-and-discovery.md`** — CoreDNS, FQDN/search-domain rules, the ndots:5
  gotcha and fixes, headless/SRV records, ExternalName resolution, debugging DNS.
- **`references/loadbalancing-and-external-dns.md`** — MetalLB (L2 + BGP, IPAddressPool,
  L2Advertisement, pools/priority/scoping/static IPs), external-dns wiring, multi-cluster
  GSLB/failover (K8GB), L4-vs-L7 deep dive.
- **`references/service-mesh-istio.md`** — sidecar/control-plane model, Gateway+VirtualService,
  DestinationRule subsets & load-balancing, canary/blue-green traffic splitting, mTLS via
  PeerAuthentication, RequestAuthentication/AuthorizationPolicy, ServiceEntry/Sidecar,
  observability (Kiali/Jaeger), ambient mesh.
