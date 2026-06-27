---
name: kubernetes-security-rbac
description: >-
  Secure Kubernetes clusters and workloads. Use for authentication (x509 certs, OIDC,
  ServiceAccount and projected/bound tokens), authorization with RBAC (Roles, ClusterRoles,
  RoleBindings, least privilege, kubectl auth can-i), Secrets management (etcd encryption at
  rest, External Secrets Operator, Vault, Sealed Secrets, SOPS), Pod Security Admission
  (privileged/baseline/restricted; PodSecurityPolicy was removed in 1.25) and securityContext
  hardening, admission control and policy engines (OPA/Gatekeeper, Kyverno), runtime security
  (Falco, KubeArmor), image scanning and signature verification (Cosign), multi-tenancy with
  namespaces/ResourceQuota/LimitRange/vClusters, audit logging, and CIS hardening. Trigger on
  Forbidden or RBAC errors, 'who can access', securityContext, runAsNonRoot, admission
  webhooks, quotas, or any request to lock down a cluster - even without saying Kubernetes.
  NetworkPolicy mechanics live in kubernetes-networking; image build hardening in
  container-fundamentals.
---

# Kubernetes Security & RBAC

This skill equips Claude to design and debug the full Kubernetes security stack: who can authenticate, what they
are authorized to do, how workloads get identity, how secrets and policies are enforced, and how to harden a
cluster against real-world attacks.

## When to use this skill

- Writing or debugging **RBAC** — `Role`, `ClusterRole`, `RoleBinding`, `ClusterRoleBinding`, choosing verbs/
  resources/apiGroups, least-privilege design, `kubectl auth can-i`, aggregation.
- Any **`Forbidden` (403)** or **`Unauthorized` (401)** error, "User X cannot list/get/create resource Y",
  "service account cannot…", impersonation issues.
- Setting up **authentication**: x509 client certs + CSR API, OIDC/SSO integration, ServiceAccount tokens,
  projected/bound tokens, `imagePullSecrets` for private registries.
- **Secrets**: encryption at rest in etcd, why Secrets aren't encrypted by default, External Secrets Operator,
  HashiCorp Vault, Sealed Secrets, SOPS, env-var vs file mounting risks.
- **Pod hardening**: Pod Security Admission (Privileged/Baseline/Restricted), `securityContext`
  (`runAsNonRoot`, `readOnlyRootFilesystem`, capability drops, seccomp). Note PodSecurityPolicy was **removed in 1.25**.
- **Admission control & policy**: validating/mutating webhooks, OPA/Gatekeeper, Kyverno — and choosing between them.
- **Runtime & image security**: KubeArmor, Falco, image scanning, signing/verifying with Cosign.
- **Multi-tenancy**: namespaces as the security boundary, `ResourceQuota`, `LimitRange`, soft vs hard isolation, vClusters.
- **Auditing & hardening**: audit policy/log, CIS Benchmark checklist.

Boundary: NetworkPolicy *mechanics* live in the **kubernetes-networking** skill — own the network *security
strategy* here and cross-link. Image *build* hardening (Dockerfile, minimal base images) → **container-fundamentals**.

## Core concepts

### The request pipeline: authn → authz → admission
Every call to the API server passes through three gates, in order:

1. **Authentication** — "who are you?" Modules run in a chain until one succeeds. Result: a username + groups
   (+ UID). Failure → `401 Unauthorized` (or treated as `system:anonymous` if anonymous is allowed).
2. **Authorization** — "are you allowed?" RBAC (and/or Node, Webhook) evaluates user/groups vs request verb +
   resource. RBAC is **purely additive/allow-only — there are no deny rules**; everything is denied by default.
   Failure → `403 Forbidden`.
3. **Admission control** — mutating webhooks run first (can modify the object), then validating webhooks +
   built-in controllers (can reject). Admission cannot block read verbs (get/list/watch).

**401 vs 403 is the single most useful debugging signal**: 401 = the API server doesn't know who you are
(bad/expired cert or token); 403 = it knows you but RBAC denies the action.

### There are no User objects
Kubernetes has **no `User` or `Group` object**. Users and groups are *asserted* at authentication time (by a
cert's CN/O fields, an OIDC token's claims, or a webhook). Only **ServiceAccounts** are real objects. A SA
authenticates as `system:serviceaccount:<namespace>:<name>` and is a member of system groups like
`system:serviceaccounts` and `system:serviceaccounts:<namespace>` — you cannot put a SA in an arbitrary group.

### RBAC in one breath
A **Role/ClusterRole** is a list of rules; each rule = `apiGroups` + `resources` (+ optional `resourceNames`/
sub-resources) + `verbs`. A **RoleBinding/ClusterRoleBinding** ties subjects (User/Group/ServiceAccount) to one
role. The combinations:

| Role type   | Binding type        | Scope                                              |
|-------------|---------------------|----------------------------------------------------|
| Role        | RoleBinding         | permissions in **one namespace**                   |
| ClusterRole | ClusterRoleBinding  | permissions **cluster-wide** + on non-namespaced resources |
| ClusterRole | RoleBinding         | the ClusterRole's permissions, **scoped to that one namespace** (great for reuse) |

The core API group is the empty string `""` (pods, services, secrets, configmaps). Sub-resources are explicit:
authorizing `pods` does **not** authorize `pods/log` or `pods/exec`.

### Namespaces are the security boundary
RBAC, ResourceQuota, LimitRange, and most policy all key off the namespace. RBAC can limit who *reads* a Secret
via the API, but **cannot stop a workload in the same namespace from mounting it**. If you need to isolate
secrets, create a new namespace — or a vCluster for hard isolation.

## Workflow / how to approach security tasks

### Designing least-privilege RBAC
1. Identify the **subject**: prefer **groups** (from OIDC/AD) over individual users; for workloads use a
   dedicated **ServiceAccount** per app — never the namespace `default` SA with broad rights.
2. Enumerate exactly the **resources + verbs** needed. Find apiGroups with `kubectl api-resources -o wide`.
   Don't reach for wildcards `["*"]` unless you truly mean "everything, forever, including future CRDs."
3. Choose scope: namespace Role for app-local perms; ClusterRole + per-namespace RoleBinding to reuse one
   definition across many namespaces (e.g. a log shipper that reads pods in several namespaces).
4. Bind, then **verify** with `kubectl auth can-i`:
   ```bash
   kubectl auth can-i list pods --as=system:serviceaccount:app:reader -n app
   kubectl auth can-i '*' '*' --as=jane@example.com          # is this user cluster-admin?
   kubectl auth can-i create deployments --as-group=devs -n app
   ```
5. Confirm identity end-to-end with `kubectl auth whoami` (shows the username + groups the API server sees).

### Debugging "Forbidden" errors
Read the message literally — it names the user, verb, resource, apiGroup, and scope:
`User "system:serviceaccount:app:reader" cannot list resource "services" in API group "" in the namespace "app"`.
Reverse-engineer the missing rule (here: `apiGroups: [""]`, `resources: ["services"]`, `verbs: ["list"]`) and add
it to the bound role. For non-obvious failures (controllers, operators), enable the **audit log** and run
**audit2rbac** against it to auto-generate a minimal working policy — then convert it to bind a **group**, not a
user. See `references/rbac.md`.

### Securing a Pod
Default to the **Restricted** Pod Security Standard and a hardened `securityContext`: `runAsNonRoot: true`, a
non-zero `runAsUser`, `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true`, `capabilities.drop:
["ALL"]`, and a `seccompProfile` of `RuntimeDefault`. Enforce it cluster-wide with **Pod Security Admission**
namespace labels (or Kyverno/Gatekeeper for finer control). See `references/pod-security-and-securitycontext.md`.

### Managing secrets
Decide based on a threat model, not reflex:
- Plain Secrets are *usually fine* — but base64 is **encoding, not encryption**, and Secrets are stored
  unencrypted in etcd by default. Enable **encryption at rest** (and ideally a KMS provider).
- For GitOps/compliance, externalize the source of truth into **Vault** + **External Secrets Operator** (syncs
  into native Secrets) using the Pod's own identity (projected SA token) — never a static key.
- **Never store secret data in Git**, even encrypted (Sealed Secrets is an anti-pattern for this reason).
- Prefer **volume mounts over env vars** (env vars leak in logs/debug dumps and don't update live).

See `references/secrets-management.md`.

### Choosing a policy engine
- **Pod Security Admission** — built-in, zero-dependency, the three standard profiles. Use first for pod hardening.
- **Kyverno** — Kubernetes-native, policies written in YAML; great for validate/mutate/generate without learning a
  new language. Kubernetes-only.
- **OPA/Gatekeeper** — policies in **Rego**; more powerful and reusable *outside* Kubernetes too, with a tested
  policy library and audit of existing objects. Steeper learning curve.

See `references/policy-engines-opa-kyverno.md`.

## Common pitfalls & anti-patterns

- **Binding to users instead of groups.** Per-user bindings don't scale, can't be audited by query, and break on
  identity-provider URL or email changes. Bind to groups asserted by your IdP.
- **Trying to write "allow everything EXCEPT X" in RBAC.** Impossible by design — RBAC has no deny. Enumerate
  allowed resources, or use an admission controller (Gatekeeper/Kyverno) to layer a deny.
- **Using `mail`/email as the OIDC username claim.** Anti-pattern; emails change. Use the immutable `sub` claim.
- **Using ServiceAccount tokens for human/external auth.** SAs are for in-cluster workload identity. External
  users should use OIDC. Static SA tokens (pre-1.24 style) never expire and can't be grouped — avoid.
- **Authorizing `pods` and expecting `pods/log` or `pods/exec` to work.** Sub-resources need explicit rules.
- **Treating base64 as security.** It is not encryption.
- **`securityContext: { privileged: true }` casually.** A privileged container shares the host PID/namespaces and
  is a one-step container breakout (e.g. `nsenter` into the host). Only for genuine node agents.
- **Forgetting ResourceQuota/LimitRange.** A namespace without them lets one tenant starve the whole cluster.
- **Assuming Vault alone secures a secret.** Anyone who can `kubectl exec` into the pod can read the mounted
  secret file or env var. Layer runtime policy (KubeArmor) to restrict which process may read it.
- **Granting namespace `admin` in multi-tenant clusters.** The aggregated `admin` ClusterRole lets a tenant create
  Roles/RoleBindings and potentially escalate. Consider vClusters or admission guards instead.

## Reference files

- `references/authentication.md` — authn methods (x509 + CSR API, OIDC config & flow, static tokens, webhook/
  proxy, bootstrap tokens), ServiceAccount & projected/bound tokens, the authn→authz→admission flow, registry creds.
- `references/rbac.md` — full RBAC catalog: Role/ClusterRole/bindings YAML, verbs↔HTTP mapping, aggregation,
  default ClusterRoles (admin/edit/view/cluster-admin), `kubectl auth can-i`, audit log + audit2rbac debugging.
- `references/secrets-management.md` — Secret objects, encryption at rest/KMS, External Secrets Operator, Vault
  (TokenReview + sidecar injector), Sealed Secrets, SOPS, consumption (volume vs env vs API), CSI driver.
- `references/pod-security-and-securitycontext.md` — Pod Security Admission profiles & labels, full hardened
  securityContext, capabilities, seccomp/AppArmor/SELinux, runtime classes (gVisor/Kata), the PSP→PSA migration.
- `references/policy-engines-opa-kyverno.md` — validating/mutating webhooks, OPA/Gatekeeper (Rego,
  ConstraintTemplate, Constraint, mutations, audit) vs Kyverno, image signing/verification with Cosign.
- `references/runtime-and-image-security.md` — KubeArmor (eBPF/LSM, KubeArmorPolicy, karmor, inline vs post-attack
  mitigation), Falco, image scanning, supply-chain notes.
- `references/multitenancy-and-quotas.md` — namespaces, ResourceQuota, LimitRange, soft vs hard isolation,
  vClusters (architecture, identity, HA), self-service tenancy.
- `references/hardening-checklist.md` — CIS Benchmark-driven checklist, audit logging setup, API/upgrade hygiene,
  OWASP K8s Top 10, a triage flow for security incidents.
