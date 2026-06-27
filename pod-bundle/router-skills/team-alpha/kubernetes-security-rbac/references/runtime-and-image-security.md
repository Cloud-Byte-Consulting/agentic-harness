# Runtime & Image Security

What RBAC, admission, and Pod Security can't see: what a process *does once running inside a container*. Covers
KubeArmor, Falco, image scanning, and supply-chain basics. Pairs with image *signing* in
`references/policy-engines-opa-kyverno.md`.

## Table of contents
- The runtime security gap
- KubeArmor (eBPF + LSM)
- KubeArmorPolicy examples
- Default posture (block vs audit)
- karmor CLI
- Inline vs post-attack mitigation
- Falco
- Image scanning
- Supply-chain notes

## The runtime security gap

Kubernetes audits API calls, but actions *inside* a container — `kubectl exec` then reading a file, dropping a
binary, launching an SSH daemon or a crypto miner — are largely invisible to the API server and unstoppable by
RBAC. Crucially, **a secrets manager doesn't protect a running secret**: anyone who can `exec` into the pod can
`cat` the mounted secret file or dump env vars. Runtime security closes this gap by mediating syscalls/file/
network/process events against policy, in real time.

## KubeArmor (eBPF + LSM)

KubeArmor (CNCF) sits between the pod runtime and the host kernel, using **eBPF** (in-kernel observation) and
**LSMs** (AppArmor/SELinux — Mandatory Access Control enforcement) to *allow, audit, or block* process, file,
network, and capability events. It abstracts away LSM syntax: you write one `KubeArmorPolicy` and KubeArmor
generates the right AppArmor/SELinux profile per node/distro. It runs as a DaemonSet (one `kubearmor` pod per
node) plus a controller (admission for policy management) and a relay (log aggregation).

Deploy with the `karmor` CLI or Helm:
```bash
karmor install            # deploys operator, DaemonSet, controller, relay
karmor probe              # lists supported LSM + default posture per namespace
```

## KubeArmorPolicy examples

Block writes to `/bin` in a namespace (any process):
```yaml
apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata: { name: block-write-bin, namespace: app }
spec:
  action: Block               # Block | Allow | Audit (Block is default)
  file:
    matchDirectories:
      - dir: /bin/
        readOnly: true         # reads allowed, writes denied
        recursive: true
  message: "Write attempt to /bin denied"
  severity: 5                  # your own 1–10 rating
```

Allow-list so **only the nginx binary** can read a mounted secret — even root in the container is blocked
otherwise (this is how you defend a secret that Vault delivered to a file):
```yaml
apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata: { name: protect-secret, namespace: app }
spec:
  selector:
    matchLabels: { app: nginx-web }
  file:
    matchDirectories:
      - dir: /etc/secrets/
        recursive: true
        fromSource:
          - path: /usr/sbin/nginx     # only this process may read
      - dir: /etc/secrets/
        recursive: true
        action: Block                 # everything else: blocked
  process:
    matchPaths:
      - path: /usr/sbin/nginx
  action: Allow
```
Result: `cat /etc/secrets/myenv` → `Permission denied`, but nginx still serves it. Other useful targets:
`process.matchPaths` (e.g. block `/usr/sbin/sshd` so no container can run an SSH daemon), `network`, `capabilities`.

## Default posture (block vs audit)

For an **Allow** policy, anything not explicitly allowed is, by default, only **audited** (logged), not blocked.
Flip the default posture to `block` globally (edit the `kubearmor-config` ConfigMap):
```yaml
defaultFilePosture: block
defaultNetworkPosture: block
defaultCapabilitiesPosture: block
```
Or per namespace via annotation:
```bash
kubectl annotate ns app kubearmor-file-posture=block --overwrite
```
Use the `Audit` action (or audit posture) to test a policy's impact before enforcing — avoids breaking a workload
by denying a file it actually needs.

## karmor CLI

- `karmor logs [--json] [--namespace|--pod|--labels …] [--logPath FILE]` — stream Alert/Log/Message events. Enable
  stdout logging on the relay via env `ENABLE_STDOUT_LOGS=true` / `ENABLE_STDOUT_ALERTS=true`.
- `karmor profile [-n ns] [-p pod]` — interactive view of observed process/file/network/syscall activity.
- `karmor recommend [-n ns | -i image] [-t CIS,NIST,MITRE,PCI-DSS,STIG]` — auto-generate hardening policies mapped
  to compliance frameworks; can run against an image not yet deployed (`-i bitnami/nginx`).
- `karmor probe` / `summary` / `sysdump` / `vm` (KubeVirt VMs) / `uninstall [--force]` (force removes leftover LSM
  profiles too).

A blocked event logs everything for forensics — namespace, host, pod, process, violated policy, operation,
resource, enforcer, and `Action: Block / Result: Permission denied`. Correlate with the API audit log (which
captures the `kubectl exec`) to reconstruct who did what.

## Inline vs post-attack mitigation

- **Post-attack** (detect-then-react, e.g. Falco alone): the action is *allowed*, then an alert fires and a
  handler reacts (delete pod, add a deny NetworkPolicy). By the time you react, data may already be
  exfiltrated/encrypted.
- **Inline** (KubeArmor): the action is *blocked before it executes* — the lock is on the door, not just a camera.
  Preferred for fast-moving threats. KubeArmor still logs the attempt for evidence.

## Falco

Falco (CNCF) is the leading **detection** engine: it watches syscalls (via eBPF or a kernel module) and fires
alerts on rule matches (e.g. "shell spawned in a container", "write to a sensitive file", "unexpected outbound
connection"). It is primarily *post-attack* — excellent for real-time alerting, threat hunting, and feeding a
response system, but doesn't block inline by itself. Common pattern: Falco detects → Falcosidekick routes the
alert → an automation enforces a response. Pair Falco (broad detection) with KubeArmor (inline enforcement) for
defense in depth.

## Image scanning

Scan images against the NVD and CIS benchmarks to find known CVEs **before** deploy (in CI) and continuously in
the registry. Tools: **Trivy**, **Grype**, Snyk (`docker scan` uses Snyk), Clair. Gate the pipeline on severity
and fail the build on critical/high CVEs. Scanning finds *vulnerabilities*; signing (Cosign) proves *provenance*
— do both, and verify signatures at admission (see the policy-engines reference).

## Supply-chain notes

- Pin images by **digest** (`@sha256:…`), not floating tags; block `:latest` via Kyverno/Gatekeeper.
- Generate and store an **SBOM** (e.g. Syft) so you can answer "are we affected by CVE-X?" quickly.
- Use `AlwaysPullImages` admission so a cached image on a node can't bypass registry authorization.
- Build minimal/distroless base images and run as non-root — that hardening lives in **container-fundamentals**;
  here, enforce it at admission and runtime.
