# Pod Security & securityContext

Hardening the workload itself: the three Pod Security Standards, how to enforce them with Pod Security Admission,
the full hardened `securityContext`, capabilities/seccomp, sandboxed runtimes, and the PSP→PSA migration.

## Table of contents
- PodSecurityPolicy is gone (1.25)
- The three Pod Security Standards
- Pod Security Admission (namespace labels)
- A hardened securityContext (copy-paste baseline)
- Pod-level vs container-level
- Linux capabilities
- privileged: true is a breakout
- seccomp, AppArmor, SELinux
- Sandboxed runtimes (gVisor, Kata) via RuntimeClass
- PSP → PSA → policy-engine migration

## PodSecurityPolicy is gone (1.25)

**PodSecurityPolicy (PSP) was deprecated in 1.21 and removed in 1.25.** Do not author PSPs. Replacements:
**Pod Security Admission** (built-in, the three standard profiles) for the common case, or **Kyverno/Gatekeeper**
for custom/fine-grained rules. PSA has **no mutation** (it only validates), so unlike PSP it won't auto-fill a
`runAsUser` for you — set securityContext explicitly or use a mutating engine.

## The three Pod Security Standards

- **Privileged** — unrestricted; for trusted, system-level workloads (CNI, node agents) only.
- **Baseline** — blocks known privilege escalations: no `privileged`, no host namespaces, no hostPath, limited
  capabilities. A minimal "don't be obviously dangerous" bar.
- **Restricted** — heavily hardened, current best practice for normal apps: must `runAsNonRoot`,
  `allowPrivilegeEscalation: false`, drop ALL capabilities (NET_BIND_SERVICE may be re-added), seccomp
  `RuntimeDefault`, no host namespaces/ports/paths.

## Pod Security Admission (namespace labels)

PSA is a built-in admission controller configured per namespace via labels. Three modes — **enforce** (reject),
**audit** (annotate audit log), **warn** (return a warning to the user) — each pinned to a profile and optionally
a version:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: app
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: latest
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```
Roll out safely by setting `warn`/`audit` to `restricted` first, fixing violations, then flipping `enforce`. PSA
can't make exemptions per-workload — for that granularity use Kyverno/Gatekeeper.

## A hardened securityContext (copy-paste baseline)

This pod passes the **Restricted** profile:
```yaml
apiVersion: v1
kind: Pod
metadata: { name: hardened, namespace: app }
spec:
  securityContext:                 # pod-level: applies to all containers
    runAsNonRoot: true
    runAsUser: 1000
    runAsGroup: 3000
    fsGroup: 2000
    seccompProfile:
      type: RuntimeDefault
  containers:
    - name: app
      image: app:1.4.2
      securityContext:             # container-level: overrides/refines pod-level
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
        runAsNonRoot: true
        capabilities:
          drop: ["ALL"]
          # add: ["NET_BIND_SERVICE"]   # only if you must bind a port < 1024
      volumeMounts:
        - name: tmp
          mountPath: /tmp           # writable scratch since root FS is read-only
  volumes:
    - name: tmp
      emptyDir: {}
```
Key fields: `runAsNonRoot`/`runAsUser` (no root in container), `allowPrivilegeEscalation: false` (block setuid
escalation), `readOnlyRootFilesystem: true` (immutable base — mount `emptyDir` for needed write paths),
`capabilities.drop: ["ALL"]`, and `seccompProfile.type: RuntimeDefault`.

## Pod-level vs container-level

Pod-level `securityContext` sets defaults inherited by all containers (good for uniform settings like
`runAsNonRoot`, `fsGroup`). Container-level overrides for that container only. Some fields exist *only* at one
level: `fsGroup`/`runAsGroup` and `seccompProfile` (pod or container), `capabilities`/`readOnlyRootFilesystem`/
`allowPrivilegeEscalation`/`privileged` are container-level. The container value wins where both are set.

## Linux capabilities

Capabilities split root's power into units. Start from `drop: ["ALL"]` and add back only what's needed:
```yaml
securityContext:
  capabilities:
    drop: ["ALL"]
    add: ["NET_BIND_SERVICE"]   # bind ports < 1024 without full root
```
Avoid `NET_ADMIN`, `SYS_ADMIN`, `SYS_PTRACE` unless essential — they approach full root.

## privileged: true is a breakout

```yaml
securityContext: { privileged: true }   # DANGER
```
A privileged container has nearly all capabilities and host device access; combined with `hostPID: true` it lets
an attacker `nsenter` into PID 1's mount namespace and write the host filesystem — a full node compromise. The
classic exploit: a privileged + hostPID pod running `nsenter --mount=/proc/1/ns/mnt -- /bin/bash`. Only genuine
node agents (CNI, storage, monitoring) should ever be privileged, and they should be gated by RBAC + admission
policy. This is also why **multi-tenancy demands** a Baseline/Restricted enforcement.

## seccomp, AppArmor, SELinux

- **seccomp** — filters syscalls. `RuntimeDefault` is the recommended baseline; custom profiles via
  `seccompProfile: { type: Localhost, localhostProfile: my-profile.json }`.
- **AppArmor** (Debian/Ubuntu/SUSE) and **SELinux** (RHEL) — Linux Security Modules enforcing mandatory access
  control. Set AppArmor via `securityContext.appArmorProfile` (1.30+) or the legacy annotation; SELinux via
  `seLinuxOptions`. KubeArmor (see runtime reference) generates LSM policies for you across distros.

## Sandboxed runtimes (gVisor, Kata) via RuntimeClass

For stronger isolation than namespaces/cgroups give (e.g. untrusted multi-tenant workloads):
- **gVisor** — a user-space kernel intercepting syscalls; strong isolation, some overhead.
- **Kata Containers** — runs each pod in a lightweight VM; near-VM isolation with container ergonomics.

Select per pod with a RuntimeClass (the class must be installed on the nodes):
```yaml
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata: { name: gvisor }
handler: runsc
---
apiVersion: v1
kind: Pod
metadata: { name: untrusted }
spec:
  runtimeClassName: gvisor
  containers:
    - name: app
      image: app:1.0
```

## PSP → PSA → policy-engine migration

If migrating off PSP:
- A PSP that mandated `runAsNonRoot`, dropped caps, and blocked privileged ≈ the **Restricted** PSA profile —
  label the namespace and you're most of the way there.
- PSP's built-in mutations (e.g. defaulting `runAsUser` to 1 when root was disallowed) have **no PSA equivalent** —
  PSA only validates. To restore defaulting, use **Gatekeeper mutations** or **Kyverno mutate** rules.
- OpenShift uses its own **Security Context Constraints (SCCs)** (pre-dating PSP, not RBAC-bound) rather than
  PSA/Gatekeeper — relevant if targeting OpenShift.
- For per-workload exceptions PSA can't express, move enforcement to Kyverno/Gatekeeper (see
  `references/policy-engines-opa-kyverno.md`).
