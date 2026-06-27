# DNS & service discovery (CoreDNS)

Contents:
1. The Service FQDN and search domains
2. Short names, cross-namespace, and the `ndots:5` gotcha
3. Pod DNS policy & custom dnsConfig
4. Headless / SRV records & StatefulSet pod names
5. ExternalName resolution
6. CoreDNS as the cluster DNS server (and extending it)
7. Debugging DNS

---

## 1. The Service FQDN and search domains

CoreDNS (the default cluster DNS server) gives every Service an A record:

```
<service>.<namespace>.svc.cluster.local
```

Breakdown: the first label is the **Service name**, the second is its **namespace**, then the
fixed suffix `svc.cluster.local` (the cluster domain — `cluster.local` is the common default).
Example: `mysql-web` in namespace `database` → `mysql-web.database.svc.cluster.local`.

The kube-dns ClusterIP (usually `10.96.0.10`) is injected into every pod's `/etc/resolv.conf`
along with a **search list** so partial names resolve:

```
nameserver 10.96.0.10
search <namespace>.svc.cluster.local svc.cluster.local cluster.local
options ndots:5
```

---

## 2. Short names, cross-namespace, and the `ndots:5` gotcha

Because of the search list:

| From a pod in namespace… | To reach Service `mysql-web` in `database` | Valid names |
|---|---|---|
| `database` (same ns) | | `mysql-web` |
| any other ns | | `mysql-web.database`, `mysql-web.database.svc`, `mysql-web.database.svc.cluster.local` |

Same-namespace clients use the bare Service name; cross-namespace clients must add at least the
namespace.

**The ndots:5 gotcha.** `options ndots:5` means any name with **fewer than 5 dots** is first
tried with each search-domain suffix appended *before* being tried as-is. So a query for
`api.github.com` (2 dots) becomes a cascade of failing lookups —
`api.github.com.<ns>.svc.cluster.local`, `…svc.cluster.local`, `…cluster.local` — before the
real `api.github.com`. Symptoms: **slow or intermittently failing external DNS** from pods,
extra CoreDNS load.

Fixes:
- Use a **fully-qualified name with a trailing dot** for external hosts: `api.github.com.`
  (the dot makes it absolute — no search-domain expansion).
- Lower `ndots` for a specific pod via `dnsConfig` (see §3), e.g. `ndots: 1`.
- For internal Service-to-Service calls, prefer the **full FQDN** so the first lookup succeeds.

---

## 3. Pod DNS policy & custom dnsConfig

`spec.dnsPolicy` (default `ClusterFirst` — cluster DNS first, then upstream). Override search
behavior with `dnsConfig`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: low-ndots
spec:
  dnsPolicy: ClusterFirst
  dnsConfig:
    options:
      - name: ndots
        value: "1"          # cut the search-domain cascade for external names
  containers:
    - name: app
      image: myapp:latest
```

Use `dnsPolicy: None` + a fully custom `dnsConfig` (nameservers + searches) when a pod must use
a non-cluster resolver.

---

## 4. Headless / SRV records & StatefulSet pod names

A **headless** Service (`clusterIP: None`) returns the A records of all backing pods instead of
a single VIP. Combined with a **StatefulSet**, each pod gets a stable per-pod DNS name:

```
<pod-name>.<service-name>.<namespace>.svc.cluster.local
# mysql-0.mysql.default.svc.cluster.local
```

SRV records are published for named ports: `_<port-name>._<protocol>.<service>.<ns>.svc.cluster.local`,
e.g. `_grpc._tcp.mysql.default.svc.cluster.local`. Clustered systems (databases, LDAP, peer
discovery) use these to address individual peers. The StatefulSet's `serviceName` must reference
the headless Service.

---

## 5. ExternalName resolution

An `ExternalName` Service has no A record of its own — CoreDNS returns a **CNAME** to the
configured external host:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: mysql-db
  namespace: prod
spec:
  type: ExternalName
  externalName: app-db.database.example.com
```

A pod querying `mysql-db.prod.svc.cluster.local` is CNAME'd to `app-db.database.example.com`,
which then resolves via upstream DNS. No proxying — the redirection is purely at the DNS layer.
(TLS pitfall: the client connects using the Service name; ensure the remote cert lists it as a
SAN.)

---

## 6. CoreDNS as the cluster DNS server (and extending it)

CoreDNS runs as a Deployment in `kube-system`, fronted by the `kube-dns` Service. Its behavior
is a `Corefile` in a ConfigMap. Beyond in-cluster names, CoreDNS can host **additional zones** —
this is how external-dns publishes LoadBalancer records into the cluster's own DNS. Example
ConfigMap stanza adding an etcd-backed zone:

```
foowidgets.k8s {
    etcd {
        path /skydns
        endpoint http://10.96.149.223:2379
    }
    cache 30
}
```

This is the bridge external-dns uses; details and the enterprise-DNS delegation/forwarding
pattern live in `loadbalancing-and-external-dns.md`.

---

## 7. Debugging DNS

```bash
# Spin up a tools pod
kubectl run tmp --rm -it --image nicolaka/netshoot -- bash

# Inside the pod:
cat /etc/resolv.conf                       # check nameserver, search list, ndots
nslookup mysql-web.database.svc.cluster.local
nslookup mysql-web.database                 # short form
nslookup api.github.com.                    # trailing dot = skip search domains

# Find the cluster DNS Service IP
kubectl get svc kube-dns -n kube-system

# Is CoreDNS healthy?
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl logs -n kube-system -l k8s-app=kube-dns
```

Triage:
- Lookup of a Service FQDN fails → the Service has no endpoints, or wrong namespace in the
  name (`kubectl get endpointslices -l kubernetes.io/service-name=<svc>`).
- External names slow/flaky → ndots:5 cascade (§2): use a trailing dot or lower ndots.
- Nothing resolves at all → CoreDNS pods down, or a default-deny **egress** NetworkPolicy is
  blocking UDP/TCP 53 to kube-dns (see network-policies.md §7).
