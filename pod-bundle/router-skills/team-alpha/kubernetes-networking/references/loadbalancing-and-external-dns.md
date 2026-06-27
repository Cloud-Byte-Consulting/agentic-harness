# Bare-metal load balancing, external-dns & multi-cluster traffic

Contents:
1. L4 vs L7 — the decision
2. MetalLB: why and how
3. MetalLB IPAddressPool & L2Advertisement
4. MetalLB advanced pools (static IP, multiple pools, priority, scoping, buggy nets)
5. BGP mode (overview)
6. external-dns: automatic DNS records
7. Wiring external-dns to CoreDNS (on-prem)
8. Delegating a zone from an enterprise DNS server
9. Multi-cluster traffic & global load balancing (K8GB)

---

## 1. L4 vs L7 — the decision

- **L4 (transport)** sees IP + port only, blind to HTTP. Handles **any TCP/UDP** workload
  (databases, DNS, gRPC streams). A `LoadBalancer` Service and MetalLB operate here.
- **L7 (application)** understands HTTP/HTTPS: host/path routing, TLS termination, one IP for
  many domains, sticky cookies. Ingress controllers and the Gateway API operate here.

Rule of thumb: HTTP(S) with routing/TLS needs → L7 (ingress/Gateway, fronted by one L4 LB).
Non-HTTP or raw port exposure → L4 LoadBalancer. You commonly use both: an L4 LB gives the
ingress controller its external IP, and the controller does L7 routing behind it.

---

## 2. MetalLB: why and how

On cloud, `type: LoadBalancer` is fulfilled by the cloud-controller-manager. On **bare metal /
on-prem there is no provider**, so `EXTERNAL-IP` stays `<pending>` forever. **MetalLB** is the
popular open-source L4 load balancer that fills this gap — it watches `LoadBalancer` Services,
assigns each an IP from a configured pool, and announces that IP on the local network.

Two components (installed from the upstream manifest/Helm chart):
- **controller** (Deployment): assigns IPs to Services.
- **speaker** (DaemonSet): announces the assigned IPs to the network. Runs on every node;
  for a given L2 service only one speaker is the active announcer at a time (failover if it dies).

Two modes: **layer 2** (ARP/NDP — simple, good for dev and small clusters) and **BGP** (peers
with routers to advertise routes — for larger/HA setups). Below focuses on L2.

---

## 3. MetalLB IPAddressPool & L2Advertisement

Two custom resources configure L2 mode. First the pool of IPs MetalLB may hand out:

```yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: pool-01
  namespace: metallb-system
spec:
  addresses:
    - 172.18.200.100-172.18.200.125   # range; CIDR (172.18.200.0/24) also works
```

Then advertise the pool over L2. An `L2Advertisement` with no `ipAddressPools` advertises
**all** pools:

```yaml
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: l2-all-pools
  namespace: metallb-system
# spec omitted → advertises every IPAddressPool
```

To advertise only specific pools:

```yaml
spec:
  ipAddressPools:
    - pool-01
    - pool-03
```

A `LoadBalancer` Service now gets an IP from the pool automatically, reachable on the LAN.

---

## 4. MetalLB advanced pools

**Static IP for a Service** (`spec.loadBalancerIP` is deprecated since 1.24 — use the
annotation):

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-web
  annotations:
    metallb.universe.tf/loadBalancerIPs: 172.18.200.210
spec:
  type: LoadBalancer
  selector:
    app: nginx-web
  ports:
    - port: 80
      targetPort: 8080
```

MetalLB rejects the assignment (with a warning event on the Service) if the IP is outside its
pools or already in use — preventing silent conflicts.

**Pick a specific pool** for a Service:

```yaml
metadata:
  annotations:
    metallb.universe.tf/address-pool: pool-02
```

**Priority** (lower number = higher priority; used when no pool is requested) and
**namespace scoping** (restrict which namespaces may draw from a pool):

```yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: ns-scoped-pool
  namespace: metallb-system
spec:
  addresses:
    - 172.18.205.0/24
  serviceAllocation:
    priority: 50
    namespaces:
      - web
      - sales            # only Services in web/sales may use this pool
```

**Buggy networks** — old gear may flag `.0`/`.255` addresses as a Smurf attack. Avoid them:

```yaml
spec:
  addresses:
    - 172.18.205.0/24
  avoidBuggyIPs: true
```

A static IP is valuable for things that other systems must point at by IP — e.g. exposing
CoreDNS so an enterprise DNS server can forward to it (§8), or WAF/firewall allow-lists.

---

## 5. BGP mode (overview)

Instead of L2 ARP announcements, MetalLB can establish BGP peering with your routers and
advertise Service IPs as routes — enabling true multi-node traffic distribution and integration
with existing routed networks. Configure via `BGPPeer` and `BGPAdvertisement` resources. Use
for larger/HA clusters; L2 is fine for dev and small deployments. See
https://metallb.universe.tf/concepts/bgp/.

---

## 6. external-dns: automatic DNS records

A `LoadBalancer` gets an IP but **no DNS name** — manually registering each MetalLB IP is a
maintenance burden. **external-dns** (a Kubernetes SIG controller) watches Services (and/or
Ingresses) and creates DNS records automatically. It is **not a DNS server** — it's a controller
that programs an actual DNS provider (Route 53, Azure DNS, Cloudflare, Google Cloud DNS, CoreDNS,
RFC2136, Pi-hole, … 30+ providers).

Request a record by annotating the Service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-ext-dns
  annotations:
    external-dns.alpha.kubernetes.io/hostname: nginx.foowidgets.k8s
spec:
  type: LoadBalancer
  selector:
    app: nginx
  ports:
    - port: 80
      targetPort: 80
```

external-dns sees the annotation and creates an A record `nginx.foowidgets.k8s → <LB IP>`.

---

## 7. Wiring external-dns to CoreDNS (on-prem)

For providers without native dynamic registration, run external-dns with `--provider=coredns`,
backed by an **etcd** store that CoreDNS reads:

```yaml
# external-dns deployment args
args:
  - --source=service
  - --provider=coredns
  - --log-level=info
# env: ETCD_URLS=http://etcd-dns.etcd-dns.svc:2379
```

Add an etcd-backed zone to the cluster CoreDNS Corefile (ConfigMap in `kube-system`):

```
foowidgets.k8s {
    etcd {
        path /skydns
        endpoint http://10.96.149.223:2379
    }
    cache 30
}
```

external-dns writes records into etcd; CoreDNS serves them. Verify from a pod:
`nslookup nginx.foowidgets.k8s` should return the LoadBalancer IP.

---

## 8. Delegating a zone from an enterprise DNS server

The cluster CoreDNS names aren't reachable outside the cluster until your main DNS server
**delegates/forwards** the zone to it. Steps:

1. Expose CoreDNS externally with a `LoadBalancer` Service on UDP/TCP 53 with a **static IP**
   (so the forwarder target is stable):

   ```yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: kube-dns-ext
     namespace: kube-system
     annotations:
       metallb.universe.tf/loadBalancerIPs: 10.2.1.74
   spec:
     type: LoadBalancer
     selector:
       k8s-app: kube-dns
     ports:
       - name: dns
         port: 53
         protocol: UDP
         targetPort: 53
   ```

2. On the enterprise DNS server, create a **conditional forwarder** (or zone delegation) for
   `foowidgets.k8s` → `10.2.1.74`. Now `nslookup nginx.foowidgets.k8s` works from any client
   that uses the corporate DNS server.

---

## 9. Multi-cluster traffic & global load balancing (K8GB)

Running the same app across clusters (e.g. prod + DR) needs **global server load balancing
(GSLB)** — a DNS-level traffic cop that hands clients the IP of a *healthy* cluster and pulls
failed ones out. Commercial options exist (F5, Citrix, Route 53, Traffic Director); the
CNCF open-source option is **K8GB**, which provides GSLB using plain DNS with **no central
management cluster and no single point of failure**, driven by a single CRD per app.

Strategies K8GB supports: **round robin**, **weighted round robin**, **failover** (all traffic
to the primary cluster until its pods are all unhealthy, then to the secondary), and **GeoIP**
(nearest cluster; requires EDNS0 + a GeoIP DB).

A `Gslb` resource embeds an Ingress spec plus a strategy; K8GB creates the Ingress for you and
keeps each cluster's CoreDNS zone in sync, updating the A record based on **native Kubernetes
health checks**:

```yaml
apiVersion: k8gb.absa.oss/v1beta1
kind: Gslb
metadata:
  name: gslb-failover
  namespace: demo
spec:
  ingress:
    ingressClassName: nginx
    rules:
      - host: fe.gb.foowidgets.k8s        # GSLB-enabled FQDN
        http:
          paths:
            - path: /
              pathType: Prefix
              backend:
                service:
                  name: nginx
                  port:
                    number: 80
  strategy:
    type: failover
    primaryGeoTag: us-nyc                 # primary cluster's geo tag
```

Requirements: an ingress controller in each cluster, K8GB deployed in each cluster, and an edge
DNS server that **delegates** the GSLB zone (e.g. `gb.foowidgets.k8s`) to the per-cluster
CoreDNS LoadBalancer IPs (named `gslb-ns-<geotag>-gb.<edge-zone>`). When the primary's pods scale
to 0/become unhealthy, K8GB rewrites the shared A record to the secondary cluster's ingress IP;
when the primary recovers, it flips back automatically.
