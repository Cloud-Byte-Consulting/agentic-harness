# RBAC: Roles, Bindings, and Debugging

The authorization catalog. Lead with the manifests, then the rules of the model, then debugging.

## Table of contents
- The four objects
- Anatomy of a rule (apiGroups / resources / verbs)
- Role vs ClusterRole vs the binding matrix
- Aggregated ClusterRoles
- Default ClusterRoles
- Mapping enterprise identities (groups, not users)
- Checking access: kubectl auth can-i
- Debugging: 403 messages, audit log, audit2rbac
- RBAC has no deny ("everything except")

## The four objects

Namespace-scoped Role + RoleBinding:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: app
  name: pod-and-log-reader
rules:
  - apiGroups: [""]                       # core API group
    resources: ["pods", "pods/log"]       # sub-resource pods/log is explicit
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pod-and-log-reader
  namespace: app
subjects:
  - kind: ServiceAccount
    name: reader
    namespace: app
  - kind: User
    name: jane@example.com               # asserted by authn; no User object exists
  - kind: Group
    name: devs
roleRef:
  kind: Role
  name: pod-and-log-reader
  apiGroup: rbac.authorization.k8s.io
```

Cluster-wide ClusterRole + ClusterRoleBinding (and used for non-namespaced resources like nodes,
persistentvolumes, storageclasses):
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pv-reader
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: pv-reader
subjects:
  - kind: Group
    name: storage-admins
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: pv-reader
  apiGroup: rbac.authorization.k8s.io
```

> For a `ServiceAccount` subject, always include `namespace`; omit `apiGroup` (SAs are core). For `User`/`Group`
> subjects, set `apiGroup: rbac.authorization.k8s.io`.

## Anatomy of a rule

- **apiGroups** — the API group, *without* the version. Core resources (pods, services, secrets, configmaps,
  serviceaccounts) use `""`. Find a resource's group:
  ```bash
  kubectl api-resources -o wide | grep -i ingress
  # ingresses  ing  networking.k8s.io/v1  true  Ingress  [create delete get list patch update watch]
  ```
  → `apiGroups: ["networking.k8s.io"]`.
- **resources** — the plural resource name; sub-resources are separate strings (`pods/log`, `pods/exec`,
  `pods/status`, `deployments/scale`). Authorizing `pods` does **not** grant `pods/log` or `pods/exec`.
- **resourceNames** — optional list to restrict a rule to specific named objects (e.g. one Secret). Note:
  `list`/`watch`/`create`/`deletecollection` cannot be restricted by name.
- **verbs** — RBAC verbs, *not* HTTP verbs (there is no `GET` verb). Common: `get`, `list`, `watch`, `create`,
  `update`, `patch`, `delete`, `deletecollection`. HTTP→RBAC mapping: GET/HEAD→get (or list/watch on collections),
  POST→create, PUT→update, PATCH→patch, DELETE→delete (or deletecollection). Special verbs exist for
  impersonation (`impersonate` on users/groups/serviceaccounts), `bind`/`escalate` on roles, and `use` on
  podsecuritypolicies (legacy).

Wildcards `["*"]` match all current *and future* members — use sparingly; in particular `apiGroups:["*"],
resources:["*"], verbs:["*"]` is full admin and will silently pick up new CRDs.

## Aggregated ClusterRoles

Instead of maintaining one giant ClusterRole, aggregate smaller ones by label selector. The built-in `admin`,
`edit`, and `view` ClusterRoles work this way, which is why new CRDs aren't automatically grantable to namespace
admins until you add a labeled ClusterRole:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aggregate-widgets-admin
  labels:
    rbac.authorization.k8s.io/aggregate-to-admin: "true"   # fold into the default admin role
rules:
  - apiGroups: ["example.com"]
    resources: ["widgets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```
The controller recomputes the aggregate target's rules to include every ClusterRole matching the selector.

## Default ClusterRoles

- `cluster-admin` — superuser (`*`/`*`); bind with extreme care.
- `admin` — near-full control **within a namespace** (via RoleBinding), including creating Roles/RoleBindings —
  but cannot mutate the Namespace object or cluster-scoped quotas. In multi-tenant clusters this is risky: a
  namespace admin can grant themselves more.
- `edit` — like `admin` but **cannot** manage RBAC (no Role/RoleBinding writes).
- `view` — read-only; deliberately **excludes Secrets** (a Secret may hold a SA token that escalates privilege).

## Mapping enterprise identities (groups, not users)

Bind to **groups** asserted by your IdP, never individual users:
```yaml
subjects:
  - kind: Group
    name: cn=k8s-cluster-admins,ou=Groups,DC=example,DC=com   # the group string your OIDC/LDAP asserts
    apiGroup: rbac.authorization.k8s.io
roleRef: { kind: ClusterRole, name: cluster-admin, apiGroup: rbac.authorization.k8s.io }
```
Per-user bindings don't scale, can't be queried for "all bindings for user X", and break when the IdP URL or a
user's email changes. (With OIDC, a user's `User` subject is often prefixed with the issuer URL, e.g.
`https://idp/...#sub` — another reason to bind groups.)

## Checking access: kubectl auth can-i

```bash
kubectl auth can-i create deployments -n app                       # as yourself
kubectl auth can-i list secrets --as=jane@example.com -n app       # impersonate a user
kubectl auth can-i get pods --as-group=devs --as=jane -n app       # user + group
kubectl auth can-i list pods \
  --as=system:serviceaccount:app:reader -n app                     # a ServiceAccount
kubectl auth can-i --list -n app                                   # everything you can do here
kubectl auth can-i '*' '*'                                          # am I cluster-admin?
```
(`--as*` impersonation itself requires the `impersonate` verb.)

## Debugging Forbidden errors

The 403 message is a complete spec of the missing rule:
```
Error from server (Forbidden): services is forbidden: User
"system:serviceaccount:app:reader" cannot list resource "services" in API group "" in the namespace "app"
```
→ add to the bound Role: `apiGroups: [""]`, `resources: ["services"]`, `verbs: ["list"]`.

For opaque failures (operators, controllers that silently misbehave — e.g. Prometheus not listing Services), don't
guess: enable the **audit log** and run **audit2rbac**.

### Audit policy (passed to the API server, not applied with kubectl)
```yaml
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: Metadata           # Levels: None, Metadata, Request, RequestResponse
    omitStages: ["RequestReceived"]
```
Enable with API-server flags (kubeadm: edit `/etc/kubernetes/manifests/kube-apiserver.yaml`, mount the policy +
log path as hostPath volumes):
```
--audit-policy-file=/etc/kubernetes/audit/policy.yaml
--audit-log-path=/var/log/k8s/audit.log
--audit-log-maxage=7 --audit-log-maxbackup=10 --audit-log-maxsize=100
```
Rules match in order; only the first matching rule applies — order specific rules before broad ones. Start broad
to learn, then tighten (logging everything is expensive).

### audit2rbac
Reverse-engineers a minimal Role/Binding from observed denials:
```bash
audit2rbac --filename=/var/log/k8s/audit.log --user=jane@example.com
```
It emits a working ClusterRole + binding. **Refactor it**: bind to the relevant group rather than the literal
user before committing.

## RBAC has no deny ("everything except")

You cannot express "allow all except this one Secret." RBAC is allow-only and unordered, so a deny would require
rule precedence the engine doesn't have. The deliberate trade-off keeps the engine simple and auditable.
Enumerate the resources you *do* allow, or layer a deny with an admission controller (Gatekeeper/Kyverno). Custom
authorization webhooks or RBAC-generating controllers can fake it but are security anti-patterns.
