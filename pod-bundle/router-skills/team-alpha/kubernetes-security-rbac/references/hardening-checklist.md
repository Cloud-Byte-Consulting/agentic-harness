# Cluster Hardening Checklist

A practical, CIS-Benchmark-driven checklist plus audit logging, upgrade hygiene, and an incident triage flow. Use
it to assess "is this cluster secure?" — knowing that ~70%+ of Kubernetes incidents trace to misconfiguration, not
exotic exploits.

## Table of contents
- CIS Benchmark & scanners
- Identity & access
- Workload hardening
- Secrets
- Network
- Admission & policy
- Runtime & images
- Audit logging
- API server & upgrade hygiene
- OWASP Kubernetes Top 10
- Incident triage flow

## CIS Benchmark & scanners

The **CIS Kubernetes Benchmark** is the de-facto hardening standard (per-version PDF from cisecurity.org). Don't
read 300 pages by hand — run scanners that check against CIS + the NVD:
- **kube-bench** — runs the CIS node/control-plane checks on a cluster.
- **kubescape** — CIS + NSA/MITRE ATT&CK frameworks, plus an RBAC visualizer for the "who can do what" sprawl.
- **Checkov / Trivy** — scan manifests and images pre-deploy.
- **CIS-CAT** — OS-level benchmark (Ubuntu/RHEL) for the underlying nodes.
Harden the nodes too (CIS-hardened AMIs/images), not just Kubernetes.

## Identity & access

- [ ] Anonymous auth disabled on the API server (`--anonymous-auth=false` where feasible).
- [ ] Human users authenticate via **OIDC** (immutable `sub` as username claim), not static tokens or shared certs.
- [ ] No static token files / no HTTP basic auth.
- [ ] RBAC bindings reference **groups**, not individual users.
- [ ] No unnecessary `cluster-admin` bindings; audit with `kubectl auth can-i '*' '*' --as=…` and kubescape's RBAC view.
- [ ] Each workload uses a **dedicated least-privilege ServiceAccount**; `automountServiceAccountToken: false`
      where the pod never calls the API.
- [ ] The default SA in each namespace has no extra bindings.
- [ ] Wildcard RBAC rules (`*`) avoided except for genuine admin roles.

## Workload hardening

- [ ] **Pod Security Admission** labels on every namespace — `enforce: restricted` (or `baseline` where a workload
      truly needs it); `kube-system` may be `privileged`.
- [ ] Pods set `runAsNonRoot`, non-zero `runAsUser`, `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem`,
      `capabilities.drop: ["ALL"]`, `seccompProfile: RuntimeDefault`.
- [ ] No `privileged: true`, `hostPID/hostIPC/hostNetwork`, or `hostPath` mounts except for vetted node agents.
- [ ] Untrusted/multi-tenant workloads use a sandboxed runtime (gVisor/Kata) via RuntimeClass.

## Secrets

- [ ] **Encryption at rest** enabled for Secrets (ideally a KMS provider, not just a local aescbc key).
- [ ] No secret data in Git (no Sealed Secrets as a Git store); externalize to Vault + ESO or SOPS-with-KMS.
- [ ] Secrets consumed via **volume mounts**, not env vars, where possible.
- [ ] External secret access uses the **Pod's identity** (projected token + TokenReview), not static keys.
- [ ] Runtime policy (KubeArmor) restricts which process can read sensitive files.

## Network

(Mechanics live in **kubernetes-networking**; the strategy is owned here.)
- [ ] Default-deny ingress (and egress) NetworkPolicy per namespace, then allow-list.
- [ ] TLS everywhere: ingress, intra-cluster (cert-manager + internal CA or a mesh's mTLS), API server↔etcd (mTLS).
- [ ] A CNI that supports NetworkPolicy (Calico/Cilium); confirm policies actually enforce.

## Admission & policy

- [ ] Mutating + validating webhook controllers enabled; `ResourceQuota`, `LimitRanger`, `NodeRestriction`,
      `AlwaysPullImages` on.
- [ ] Kyverno or Gatekeeper enforcing org policies (allowed registries, no `:latest`, required labels, image
      signatures) — rolled out `audit`/`warn` first, then `enforce`/`deny`.

## Runtime & images

- [ ] Images scanned in CI and in the registry (Trivy/Grype); pipeline fails on critical CVEs.
- [ ] Images signed (**Cosign**) and signatures **verified at admission**.
- [ ] Images pinned by digest; `:latest` blocked.
- [ ] Runtime detection (Falco) and/or inline enforcement (KubeArmor) deployed; alerts shipped centrally.

## Audit logging

Often off by default — turn it on (an OWASP K8s Top-10 item). The audit API exists but logs nothing until you
supply a policy. Minimal start, then refine:
```yaml
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: Metadata
```
API-server flags (kubeadm: edit `/etc/kubernetes/manifests/kube-apiserver.yaml`; mount policy + log as hostPath):
```
--audit-policy-file=/etc/kubernetes/audit/policy.yaml
--audit-log-path=/var/log/audit.log
--audit-log-maxage=7 --audit-log-maxbackup=10 --audit-log-maxsize=100
```
Ship logs to a central store (OpenSearch/EFK/SIEM) and alert on suspicious patterns. On managed clusters
(EKS/AKS/GKE) enable the provider's audit logging instead. Captures: who did what, when, allowed/denied.

## API server & upgrade hygiene

- [ ] Run a supported, patched Kubernetes version — old API versions accumulate unpatched CVEs. Upgrade order:
      control plane (`kubeadm upgrade plan` → `apply`), then nodes (`kubeadm upgrade node`), then kubelet/kubectl;
      test and back up first.
- [ ] etcd encrypted, access-controlled, and backed up.
- [ ] Control-plane components not exposed to the public internet unnecessarily.

## OWASP Kubernetes Top 10

Sanity-check against it: insecure workload configs, supply-chain vulns, over-permissive RBAC, lack of policy
enforcement, **inadequate logging/monitoring**, broken authentication, missing network segmentation, secrets
management failures, misconfigured cluster components, vulnerable components.

## Incident triage flow

1. **Identify** — what access did the attacker have? Correlate API **audit log** (who/what/when) with **runtime**
   logs (KubeArmor/Falco — what ran inside the container).
2. **Contain** — cordon affected nodes; apply default-deny NetworkPolicy to the namespace; revoke the identity
   (rotate the OIDC session / delete & recreate the SA / rotate certs — note x509 client certs can't be revoked,
   only expired).
3. **Eradicate** — delete compromised pods; rebuild from a known-good, signed image; rotate every secret that the
   workload could touch.
4. **Recover & assume the worst** — a compromised cluster is hard to prove clean; for a confirmed control-plane/
   etcd breach, rebuilding the cluster is often the safest path. Restore from backup, re-key Secrets and any
   external credentials.
5. **Harden** — close the misconfiguration that allowed it (re-run this checklist), add the missing policy/
   detection, and verify with a scanner.
