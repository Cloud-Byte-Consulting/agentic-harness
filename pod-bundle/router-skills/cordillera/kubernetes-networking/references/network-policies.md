# NetworkPolicies: the mechanics

This file owns the **mechanics** of `networking.k8s.io/v1` NetworkPolicies — how the selectors,
rule shapes, and defaults actually behave. For *strategy and posture* (zero-trust rollout,
tenant-isolation design, which namespaces to lock down first), defer to the
**kubernetes-security-rbac** skill.

Contents:
1. Prerequisite: a policy-enforcing CNI
2. The default-allow → default-deny switch
3. Anatomy: the four parts
4. The podSelector / namespaceSelector AND-vs-OR trap
5. Default-deny recipes
6. Allow rules: pods, namespaces, IP blocks, ports
7. The must-allow-DNS gotcha
8. Testing a policy

---

## 1. Prerequisite: a policy-enforcing CNI

NetworkPolicy is **optional and CNI-dependent**. Calico and Cilium enforce it; plain flannel
does **not**. If the CNI doesn't support policies, every NetworkPolicy you apply is silently
inert — traffic flows as if it weren't there. Confirm the CNI before relying on policies.
(CNI install/troubleshooting itself → kubernetes-cluster-operations.)

---

## 2. The default-allow → default-deny switch

By default **all pods can talk to all pods**, including across namespaces. NetworkPolicies are
**additive allow-lists** that flip this:

- As soon as *any* policy selects a pod for a **direction** (`Ingress` or `Egress`), all
  traffic in that direction that isn't explicitly allowed is **denied** for that pod.
- A direction you never mention stays wide open. If you only write `Ingress` rules, **all
  egress remains allowed.**
- Policies are namespaced and combine as a union — if any policy allows a flow, it's allowed.

---

## 3. Anatomy: the four parts

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-netpol
  namespace: sales
spec:
  podSelector:              # 1. WHICH pods this policy applies to ({} = all in namespace)
    matchLabels:
      app: backend-db
  policyTypes:              # 2. which directions this policy governs
    - Ingress
    - Egress
  ingress:                  # 3. allowed inbound (optional)
    - from:
        - podSelector:
            matchLabels:
              app: frontend
      ports:
        - protocol: TCP
          port: 5432
  egress:                   # 4. allowed outbound (optional)
    - to:
        - podSelector:
            matchLabels:
              app: backend-db
```

- `podSelector` scopes the policy. `{}` selects **every pod in the namespace**.
- `policyTypes` declares the directions. Omitting `egress` from the list (and the spec) leaves
  egress unrestricted even if ingress is locked down.
- An empty `ingress`/`egress` rule list (`ingress: []` or omitted while listed in
  `policyTypes`) means "allow nothing in that direction" → default-deny for it.

---

## 4. The podSelector / namespaceSelector AND-vs-OR trap

This is the single most common NetworkPolicy mistake. Within a single `from`/`to` **list
element**, `podSelector` + `namespaceSelector` are **ANDed**. Across **separate list
elements**, they are **ORed**.

**AND** — "pods labeled `app=database`, *and only those* in namespaces labeled `team=backend`":

```yaml
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              team: backend
          podSelector:
            matchLabels:
              app: database
```

**OR** — "anything from namespaces labeled `team=backend`, OR any pod labeled `app=database`
(in this namespace)":

```yaml
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              team: backend
        - podSelector:
            matchLabels:
              app: database
```

Note the indentation: same `-` element = AND; separate `-` = OR. Getting this wrong produces
rules that are either far too broad or far too narrow.

---

## 5. Default-deny recipes

Deny **all** ingress and egress for a whole namespace (the foundation of a zero-trust posture —
then layer narrow allows on top):

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: sales
spec:
  podSelector: {}            # all pods
  policyTypes:
    - Ingress
    - Egress
  # no ingress/egress rules → nothing allowed
```

Default-deny **ingress only** (egress stays open):

```yaml
spec:
  podSelector: {}
  policyTypes:
    - Ingress
```

Allow-all egress explicitly (useful alongside a default-deny ingress when you don't want to
restrict outbound):

```yaml
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    - {}                     # {} = allow all egress
```

---

## 6. Allow rules: pods, namespaces, IP blocks, ports

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: db-allow
  namespace: sales
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: postgresql
  policyTypes:
    - Ingress
  ingress:
    - from:
        # OR-list of allowed sources:
        - ipBlock:
            cidr: 192.168.0.0/16
            except:
              - 192.168.5.0/24      # carve-outs
        - namespaceSelector:
            matchLabels:
              app: backend
        - podSelector:
            matchLabels:
              app: frontend
      ports:
        - protocol: TCP
          port: 5432
```

- `ipBlock` matches external/CIDR sources (e.g. for traffic entering via a LoadBalancer or
  on-prem clients); `except` removes sub-ranges.
- A rule with `ports` allows **only those ports** from the matched peers — any other port from
  the same peer is denied. Omitting `ports` allows all ports from those peers.
- `to` (egress) mirrors `from` exactly: same `podSelector`/`namespaceSelector`/`ipBlock`/`ports`.

---

## 7. The must-allow-DNS gotcha

If you apply a **default-deny egress**, pods can no longer resolve DNS — every `nslookup`/Service
lookup fails, breaking nearly everything. Always pair default-deny egress with an explicit
allow for DNS to CoreDNS (UDP and TCP 53):

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns-egress
  namespace: sales
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
          podSelector:
            matchLabels:
              k8s-app: kube-dns
      ports:
        - protocol: UDP
          port: 53
        - protocol: TCP
          port: 53
```

(The built-in `kubernetes.io/metadata.name` namespace label is reliable for selecting
`kube-system`.)

---

## 8. Testing a policy

```bash
# Find the target pod's IP
kubectl get pods -n sales -o wide

# Connection that SHOULD be allowed (matching label) — succeeds
kubectl run probe --rm -it --image nicolaka/netshoot --labels app=frontend -n sales \
  -- nc -vz <db-pod-ip> 5432

# Connection that should be BLOCKED (wrong label) — times out
kubectl run probe --rm -it --image nicolaka/netshoot --labels app=wronglabel -n sales \
  -- nc -vz <db-pod-ip> 5432
```

A timeout (not a refusal) is the typical signature of a NetworkPolicy drop. If an allowed
connection also fails, recheck the AND/OR selector semantics (§4) and whether you blocked DNS
or the return path with an over-broad egress default-deny.
