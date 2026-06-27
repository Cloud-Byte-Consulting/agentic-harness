# Services: types, ports, DNS, probes, EndpointSlices

Contents:
1. The three ports (port / targetPort / nodePort)
2. ClusterIP
3. NodePort
4. LoadBalancer
5. ExternalName
6. Headless Services (and StatefulSet DNS)
7. sessionAffinity
8. Multi-port and multi-protocol Services
9. EndpointSlices & how routing actually works
10. Readiness / liveness / startup probes (gating Service traffic)
11. Quick debugging recipes

---

## 1. The three ports

A Service deals with up to three distinct port numbers — keep them straight:

- **`port`** — the port the *Service itself* listens on (what clients hit: `my-svc:port`).
- **`targetPort`** — the port on the *backing pods/containers* that traffic is forwarded to.
  May be a number or a **named** container port.
- **`nodePort`** — (NodePort/LoadBalancer only) the port opened on *every node*, in the
  `30000–32767` range. Auto-assigned if omitted.

`port` and `targetPort` are independent; people often set them equal for clarity, but they
need not be. No worker-node port is involved for ClusterIP.

---

## 2. ClusterIP (default) — internal-only

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-clusterip
spec:
  type: ClusterIP            # default; can be omitted
  selector:
    app: nginx              # MUST match the pods' labels
  ports:
    - name: http
      port: 80              # clients call nginx-clusterip:80
      targetPort: 80        # forwarded to container port 80
      protocol: TCP
```

Reachable only from within the cluster, via its stable ClusterIP or DNS name
`nginx-clusterip.<namespace>.svc.cluster.local`. The most common type — use it for
inter-pod communication and internal backends (databases, APIs). Not reachable from outside;
test with `kubectl port-forward` or a debug pod.

---

## 3. NodePort — `<nodeIP>:<nodePort>` external access

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nodeport-whoami
spec:
  type: NodePort
  selector:
    app: whoami
  ports:
    - port: 80              # the Service's own port (also reachable as a ClusterIP internally)
      targetPort: 80        # container port
      nodePort: 30001       # optional; 30000-32767; omit to auto-assign
      protocol: TCP
```

A NodePort is also a ClusterIP internally — it just additionally opens `nodePort` on every
node. Reach it at `<ANY_NODE_IP>:30001`. Caveats: the port range is restrictive (not 80/443),
the assigned port changes if you delete+recreate the Service, and targeting an unhealthy node
fails. Best uses: dev/test, sitting *behind* an external load balancer, or temporarily
swapping a ClusterIP to NodePort to **bypass the ingress controller** during triage.

---

## 4. LoadBalancer — external L4

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-lb
spec:
  type: LoadBalancer
  selector:
    app: nginx
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
```

Provisions an external L4 load balancer. On cloud (AWS/GCP/Azure/OpenStack) the
cloud-controller-manager creates a real LB asynchronously and the IP appears in
`status.loadBalancer` / the `EXTERNAL-IP` column. **On bare metal there is no provider**, so
`EXTERNAL-IP` stays `<pending>` until you install MetalLB or similar
(see `loadbalancing-and-external-dns.md`).

- It is L4: no host/path routing, no HTTP TLS termination. Use it for non-HTTP (TCP/UDP)
  workloads, or as the single entry point in front of an ingress controller.
- Static IP: `spec.loadBalancerIP` is **deprecated since 1.24**. Use the LB implementation's
  mechanism instead (MetalLB: `metallb.universe.tf/loadBalancerIPs: 172.18.200.210` annotation).
- `spec.loadBalancerClass` selects a non-default LB implementation; immutable once set.

---

## 5. ExternalName — DNS CNAME to an external host

```yaml
apiVersion: v1
kind: Service
metadata:
  name: mysql-db
  namespace: prod
spec:
  type: ExternalName
  externalName: app-db.database.example.com   # no selector, no ports
```

No proxying and no pods — the cluster DNS returns a CNAME to `externalName`. In-cluster
clients use the stable name `mysql-db`; redirection happens purely at DNS. Great for aliasing
an external database/SaaS, and for migration: later swap the ExternalName for a real ClusterIP
Service of the same name (pointing at in-cluster pods) with **zero client changes**.

Watch out for **TLS name mismatches**: the client connects using the Service name, but the
remote cert may not list it — fix with a SAN on the cert.

---

## 6. Headless Services (and StatefulSet DNS)

Set `clusterIP: None` to get a headless Service — no VIP, no kube-proxy load balancing.
Instead, DNS returns the **A records of all backing pods** directly; the client picks one.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-headless
spec:
  clusterIP: None
  selector:
    app: nginx
  ports:
    - port: 80
      targetPort: 80
```

Use for clustered stateful systems (LDAP, databases, peer discovery) where clients need each
pod individually. With a **StatefulSet**, a headless Service gives each pod a stable DNS name:

```
<pod-name>.<service-name>.<namespace>.svc.cluster.local
# e.g.  mysql-0.mysql.default.svc.cluster.local
```

The StatefulSet's `serviceName` must point at the headless Service. Stable names persist
across reschedules, which is exactly what stateful peers need.

---

## 7. sessionAffinity (sticky by client IP)

Default is `None` (each connection independently balanced). For L4 client-IP stickiness:

```yaml
spec:
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800     # default 3h
```

This is L3/L4 stickiness only. For cookie-based stickiness use an Ingress controller (nginx
annotation) or a mesh `DestinationRule` `consistentHash` cookie.

---

## 8. Multi-port and multi-protocol

Multiple ports must each be **named**:

```yaml
spec:
  selector:
    app: web
  ports:
    - name: http
      port: 80
      targetPort: 80
    - name: https
      port: 443
      targetPort: 443
```

Same port number on **TCP and UDP** (e.g. DNS) on a `LoadBalancer` works since the
`MixedProtocolLBService` feature went GA in Kubernetes 1.26 — just give each a unique name:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: coredns-ext
  namespace: kube-system
spec:
  type: LoadBalancer
  selector:
    k8s-app: kube-dns
  ports:
    - name: dns-tcp
      port: 53
      protocol: TCP
      targetPort: 53
    - name: dns-udp
      port: 53
      protocol: UDP
      targetPort: 53
```

---

## 9. EndpointSlices & how routing actually works

A Service with a `selector` does not route by magic. The endpoint controller watches for
**ready** pods matching the selector and records their `IP:port` into `EndpointSlice` objects
(`discovery.k8s.io/v1`). Each slice holds up to ~100 endpoints; large Services span multiple
slices. This replaced the old single `Endpoints` object, which forced a full re-sync to every
node on any change — EndpointSlices update surgically, scaling far better.

On each node, **kube-proxy** reads these and programs forwarding rules to the Service VIP:
- `iptables` mode (default in many distros): random/hash-based selection of a backend.
- `ipvs` mode: real load-balancing algorithms (`rr` round-robin, `lc` least-connection,
  `sh` source-hashing) and better large-scale performance.

So "even round-robin" is only guaranteed in ipvs mode. A Service is a logical construct; there
is no per-Service proxy process — kube-proxy does all routing.

Inspect endpoints (the #1 debug step):

```bash
kubectl get endpointslices -l kubernetes.io/service-name=nginx-frontend
kubectl get endpoints nginx-frontend            # legacy view, still works
kubectl describe svc nginx-frontend             # shows Selector + Endpoints
```

Empty endpoints ⇒ no ready pod matches the selector (label typo, or all pods failing
readiness).

---

## 10. Probes — gating Service traffic

A Service forwards to a pod as soon as its labels match **and** it is *Ready*. Probes (defined
on the **pod/container**, not the Service) control that.

**readinessProbe** — "may I receive traffic yet?" Failing readiness removes the pod from the
Service's endpoints without killing it. The correct fix for slow-starting apps that 500 on
first requests.

```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 80
  initialDelaySeconds: 5
  periodSeconds: 5
```

**livenessProbe** — "is this pod wedged?" Failing liveness restarts the container. Use for
deadlock detection on long-running services.

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 80
  initialDelaySeconds: 5
  periodSeconds: 5
```

Probe types: `httpGet` (success = HTTP status ≥200 and <400), `tcpSocket` (success = connection
opens — good for non-HTTP like LDAP), `exec` (success = command exits 0), and `grpc` (for apps
implementing the gRPC Health Checking Protocol; liveness/readiness only, not via named port).

**startupProbe** — for legacy apps with long, variable startup. It runs *first*; liveness and
readiness are suppressed until it passes, so a slow boot isn't mistaken for a deadlock. Give it
a generous `failureThreshold * periodSeconds` budget:

```yaml
startupProbe:
  httpGet:
    path: /healthz
    port: 80
  failureThreshold: 30
  periodSeconds: 10        # tolerates up to 300s startup
livenessProbe:
  httpGet:
    path: /healthz
    port: 80
  periodSeconds: 5
```

Shared tunables (all probes): `initialDelaySeconds`, `periodSeconds`, `timeoutSeconds`,
`successThreshold`, `failureThreshold`. Use named ports for `httpGet`/`tcpSocket`:

```yaml
ports:
  - name: web
    containerPort: 8080
livenessProbe:
  httpGet:
    path: /healthz
    port: web              # references the named port
```

---

## 11. Quick debugging recipes

```bash
# Does the Service have endpoints?
kubectl get endpointslices -l kubernetes.io/service-name=<svc>

# Resolve + curl from inside the cluster (ClusterIP isn't reachable from your laptop)
kubectl run tmp --rm -it --image nicolaka/netshoot -- bash
#   nslookup <svc>.<ns>.svc.cluster.local
#   curl http://<svc>.<ns>

# Compare Service selector vs pod labels
kubectl describe svc <svc>
kubectl get pods --show-labels

# Temporary external test (dev only — dies when terminal closes)
kubectl port-forward svc/<svc> 8080:80
```
