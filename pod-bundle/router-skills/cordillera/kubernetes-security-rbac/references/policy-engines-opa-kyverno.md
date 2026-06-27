# Admission Control & Policy Engines (OPA/Gatekeeper, Kyverno, Cosign)

Admission control is the third request gate — it enforces and mutates policy *before* an object is persisted.
This covers the webhook mechanics, the built-in controllers, OPA/Gatekeeper vs Kyverno, and image
signing/verification with Cosign.

## Table of contents
- The two-phase admission process
- Built-in admission controllers
- Dynamic webhooks (validating & mutating)
- Choosing: PSA vs Kyverno vs OPA/Gatekeeper
- OPA/Gatekeeper (Rego, ConstraintTemplate, Constraint, mutations, audit)
- Kyverno (validate / mutate / generate, verifyImages)
- Image signing & verification with Cosign

## The two-phase admission process

After authn + authz, admission runs in two phases:
1. **Mutation** — mutating webhooks/controllers may rewrite the request (add sidecars, default labels, fill
   securityContext, inject Vault agent).
2. **Validation** — validating webhooks/controllers may reject it (policy violations, quota exceeded).

Mutation always runs before validation, so a validating policy sees the final object. Admission **cannot** block
read verbs (get/list/watch). A policy violation returns a clear error to the user.

## Built-in admission controllers

Enabled via the API server `--enable-admission-plugins` flag (kubeadm default set includes):
```
NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,
NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota
```
Notable ones: `ResourceQuota` + `LimitRanger` (enforce namespace quotas/defaults), `NodeRestriction` (limits what
a kubelet can modify), `ServiceAccount` (auto-mounts SA tokens), `AlwaysPullImages` (forces image re-pull so
cached images can't bypass registry auth), and the two **webhook** controllers that dispatch to your dynamic
policies. Inspect with:
```bash
kubectl -n kube-system get pod kube-apiserver-<node> -o yaml | grep enable-admission-plugins
```

## Dynamic webhooks (validating & mutating)

External policy is wired in via `ValidatingWebhookConfiguration` / `MutatingWebhookConfiguration` (API group
`admissionregistration.k8s.io/v1`). The webhook is an anonymous call (no per-request credential), so keep it for
*policy checks*, not for performing cluster mutations like creating bindings — separate policy from workflow.
Policy engines (Gatekeeper, Kyverno) install these configurations for you; you rarely write raw webhooks. The
user's `id_token` is forwarded to the webhook, so policies can decide based on the requester's identity/groups.

There is also a built-in, in-process alternative — **ValidatingAdmissionPolicy** (CEL expressions, GA in 1.30) —
useful for simple checks without running a webhook server.

## Choosing: PSA vs Kyverno vs OPA/Gatekeeper

| Need | Use |
|------|-----|
| Standard pod hardening (Privileged/Baseline/Restricted) | **Pod Security Admission** — built-in, no deps |
| Custom validate/mutate/generate, YAML policies, K8s-only | **Kyverno** — gentle learning curve |
| Powerful, reusable policy (also outside K8s), tested library, audit existing objects | **OPA/Gatekeeper** (Rego) |

Many clusters run PSA + one of Kyverno/Gatekeeper. Kyverno wins on approachability (no new language); Gatekeeper/
OPA wins when you want one policy language across CI, app authz, and the cluster.

## OPA / Gatekeeper

OPA evaluates policies written in **Rego**. Gatekeeper is the Kubernetes-native front end: it installs the
validating webhook, lets you ship policy as CRDs, and also **audits existing objects** (not just new ones).

Rego is allow-by-default with *implicit* control flow: a line that evaluates false halts the rule. A policy
typically produces `violation` messages; any violation = reject.

**ConstraintTemplate** holds the Rego and defines a new constraint CRD:
```yaml
apiVersion: templates.gatekeeper.sh/v1
kind: ConstraintTemplate
metadata: { name: k8sallowedrepos }
spec:
  crd:
    spec:
      names: { kind: K8sAllowedRepos }
      validation:
        openAPIV3Schema:
          type: object
          properties:
            repos: { type: array, items: { type: string } }
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package k8sallowedrepos
        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          not startswith_any(container.image, input.parameters.repos)
          msg := sprintf("image %v is from a disallowed registry", [container.image])
        }
        startswith_any(s, prefixes) { startswith(s, prefixes[_]) }
```
**Constraint** instantiates the template with parameters and a scope:
```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sAllowedRepos
metadata: { name: only-our-registry }
spec:
  enforcementAction: deny          # or `dryrun` / `warn` to roll out safely
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Pod"]
    namespaces: ["app"]
  parameters:
    repos: ["registry.example.com/"]
```
Gatekeeper also supports **mutations** (`Assign`, `AssignMetadata`) to set defaults — e.g. re-create the
PSP-style behavior of defaulting securityContext fields (`k8spsprunasnonroot`, capability defaults). Check
existing-object compliance in each Constraint's `status` (the audit results). Use the
[gatekeeper-library](https://github.com/open-policy-agent/gatekeeper-library) of tested policies before writing
your own. Debug Rego with `trace()` + `sprintf()` and OPA's verbose test output (`opa test`).

## Kyverno

Policies are plain Kubernetes YAML — no Rego. One `ClusterPolicy` with `validate`, `mutate`, or `generate` rules.

Validate (block `:latest` images):
```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata: { name: disallow-latest-tag }
spec:
  validationFailureAction: Enforce   # or Audit
  rules:
    - name: require-image-tag
      match:
        any:
          - resources: { kinds: ["Pod"] }
      validate:
        message: "Using the latest tag (or no tag) is not allowed."
        pattern:
          spec:
            containers:
              - image: "!*:latest"
```
Mutate (inject default securityContext), Generate (create a default NetworkPolicy/Secret per new namespace), and
`verifyImages` for signature checks (see Cosign below) are all expressed similarly. Kyverno runs as both a
validating and mutating webhook and can report policy results via the PolicyReport CRD.

## Image signing & verification with Cosign

Supply-chain defense: sign images at build time, then **verify the signature at admission** so only attested
images run.

Sign (Sigstore Cosign; keyless uses an OIDC identity + transparency log, or use a key pair):
```bash
cosign generate-key-pair                       # cosign.key / cosign.pub
cosign sign --key cosign.key registry.example.com/app:1.4.2
cosign verify --key cosign.pub registry.example.com/app:1.4.2
```

Enforce at admission with Kyverno `verifyImages`:
```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata: { name: verify-signatures }
spec:
  validationFailureAction: Enforce
  rules:
    - name: require-signed-images
      match:
        any:
          - resources: { kinds: ["Pod"] }
      verifyImages:
        - imageReferences: ["registry.example.com/*"]
          attestors:
            - entries:
                - keys:
                    publicKeys: |-
                      -----BEGIN PUBLIC KEY-----
                      <cosign.pub contents>
                      -----END PUBLIC KEY-----
```
Gatekeeper can do the same via its external-data / image-verification providers, and Sigstore's **policy-
controller** is a dedicated admission controller for Cosign verification. Pair signing with **image scanning**
(see `references/runtime-and-image-security.md`) — scanning finds CVEs, signing proves provenance; you want both.
