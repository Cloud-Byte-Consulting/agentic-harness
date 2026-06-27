# Container Security & Hardening

Securing containers across the lifecycle: build → ship → run. Plus rootless/Podman, scanning, SBOMs, signing,
secrets, and runtime monitoring.

## Contents
- [Threat model in one paragraph](#threat-model-in-one-paragraph)
- [Secure the build](#secure-the-build)
- [Run as non-root](#run-as-non-root)
- [Drop Linux capabilities](#drop-linux-capabilities)
- [no-new-privileges](#no-new-privileges)
- [Read-only rootfs and minimal mounts](#read-only-rootfs-and-minimal-mounts)
- [Minimal images](#minimal-images)
- [Kernel-level controls: seccomp, AppArmor, SELinux, user namespaces](#kernel-level-controls)
- [Resource limits](#resource-limits)
- [Rootless Docker and Podman](#rootless-docker-and-podman)
- [Scanning, SBOM, signing (supply chain)](#scanning-sbom-signing-supply-chain)
- [Secrets management](#secrets-management)
- [Runtime security monitoring (Falco)](#runtime-security-monitoring-falco)
- [A hardened run command](#a-hardened-run-command)
- [Hardening trade-offs](#hardening-trade-offs)

## Threat model in one paragraph

Containers share the host **kernel**, so a container escape or a compromised process attacks the host's kernel
surface — isolation is real but thinner than a VM. The biggest risks are: vulnerable/outdated base images,
secrets baked into images, over-privileged containers (root, extra capabilities, `--privileged`), and writable
root filesystems. Defense-in-depth across build, ship, and run shrinks both the attack surface and the blast
radius if something does get in.

## Secure the build

- Pin a **minimal, verified base** (`python:3.12-slim`, official/verified-publisher; avoid `:latest`).
- **Multi-stage** so compilers/SDKs/test tools never ship in the runtime image.
- Don't install unnecessary packages (`--no-install-recommends`; clean caches).
- Add a **non-root `USER`**. Never `COPY` secrets or use `ENV SECRET=…`.

```dockerfile
FROM python:3.12-slim
RUN adduser --disabled-password appuser
WORKDIR /app
COPY . .
USER appuser
CMD ["python", "main.py"]
```

## Run as non-root

Most base images default to **root**; an escape then has root-level reach. Create a system user in the
Dockerfile and switch to it, or override at run time with `--user`:
```dockerfile
RUN groupadd -r appuser && useradd -r -g appuser appuser
RUN chown -R appuser:appuser /app
USER appuser
```
```bash
docker run --user 1000:1000 my-app
```
The app must only write to paths the non-root user owns.

## Drop Linux capabilities

Even non-root containers get a default set of capabilities. Drop all, add back only what's needed:
```bash
docker run --cap-drop ALL --cap-add NET_BIND_SERVICE my-app    # e.g. bind to a low port
docker run --cap-drop ALL --cap-add CHOWN my-app
```
Never run `--privileged` unless truly unavoidable (it grants nearly all capabilities + device access).

## no-new-privileges

Prevents privilege escalation (e.g. via setuid binaries or kernel exploits) — a process can't gain privileges
beyond what it starts with:
```bash
docker run --security-opt no-new-privileges my-app
# (Linux flag form: --no-new-privileges; on Docker Desktop/macOS use --security-opt no-new-privileges:true)
```

## Read-only rootfs and minimal mounts

Make the root filesystem read-only and grant writable space only where needed (volumes or tmpfs):
```bash
docker run --read-only --tmpfs /tmp:rw,size=64m my-app
docker run --read-only -v applogs:/var/log/app my-app
```
This blocks attackers from dropping backdoors or tampering with binaries in the container.

## Minimal images

Fewer binaries = smaller attack surface and fewer CVEs:
- Prefer `-slim`/`-alpine` over full distros; **distroless** (no shell/package manager) or **scratch** where
  feasible.
- Multi-stage so build-time deps don't reach production.
- Trade-off: distroless/scratch are harder to debug (no shell) — use ephemeral/sidecar debug containers
  (see `debugging-containers.md`).

## Kernel-level controls

- **seccomp** — restricts syscalls. Docker ships a sensible default profile; you can supply a custom one.
- **AppArmor / SELinux** — MAC policies restricting filesystem/resource access per container.
- **User namespaces** — remap container UIDs to unprivileged host UIDs so container-root ≠ host-root.
- **`--icc=false`** — disable inter-container communication on the default bridge by default.

```bash
docker run --security-opt seccomp=profile.json --security-opt apparmor=my-profile my-app
```

## Resource limits

Stop a container from starving the host (also mitigates DoS):
```bash
docker run --memory 256m --cpus 0.5 --pids-limit 200 --ulimit nofile=1024:2048 my-app
```
Measure typical usage first, then set limits with headroom; exceeding `--memory` triggers an OOM kill
(exit 137). Scale horizontally rather than over-restricting.

## Rootless Docker and Podman

- **Rootless Docker** runs the daemon and containers as a non-root user (now stable/production-ready),
  dramatically reducing the impact of a daemon compromise — preferred in compliance-heavy environments.
- **Podman** is a daemonless, rootless-by-default, OCI-compatible engine. Its CLI is largely a drop-in for
  Docker (`alias docker=podman` works for most commands); it builds OCI images that run anywhere Docker images
  do, and groups containers into "pods" (Kubernetes-like). Choose it when you want no long-running root daemon.

Both produce/consume the same OCI images and run on the same `runc`/`crun` + containerd-class plumbing as
Docker, so nothing in the rest of this skill changes.

## Scanning, SBOM, signing (supply chain)

Secure the path from source to running image (full commands in `images-and-registries.md`):

- **Scan** every image for CVEs and gate CI: `trivy image --severity CRITICAL,HIGH --exit-code 1 <img>`
  (also Docker Scout, Grype). Re-scan regularly — new CVEs hit images you already shipped.
- **SBOM** with Syft: `syft <img> -o cyclonedx-json > sbom.json` — inventory for "am I affected by CVE-X?"
  and compliance (NIST SP 800-218, EO 14028).
- **Sign** with Cosign and verify before deploy:
  ```bash
  cosign generate-key-pair
  cosign sign --key cosign.key ghcr.io/acme/app:v1.0.0
  cosign verify --key cosign.pub ghcr.io/acme/app:v1.0.0
  ```
- **Enforce** at runtime: registries (Harbor) can require signed images; Kubernetes admission controllers
  (Kyverno, Connaisseur, Gatekeeper) reject unsigned/unverified images before Pods start.
- Pipeline hygiene: pin base-image versions, use private registries with RBAC, multi-stage to avoid leaking
  build secrets, isolated/patched build runners, rebuild when new CVEs drop.

## Secrets management

**Never** bake secrets into images or `ENV` (they persist in layers, leak via `docker history`/`inspect`,
travel with the image, and need a rebuild to rotate).

- **Build-time** (private deps during build): BuildKit `RUN --mount=type=secret,id=…` — present only for that
  `RUN`, never persisted.
- **Runtime, file-based**: Docker/Compose/Swarm secrets mounted at `/run/secrets/<name>`; the app reads the
  **file**, not an env var.
  - Swarm: `docker secret create db_password db_password.txt` → `docker service create --secret db_password
    -e DB_PASSWORD_FILE=/run/secrets/db_password …`. Encrypted at rest/in transit; **services only** (not
    `docker run`). Rotate by creating a new secret and updating the service (no in-place edit).
  - Compose: define under top-level `secrets:` and reference per service (file-mounted; no encryption-at-rest
    or rotation).
- **External managers** (HashiCorp Vault, AWS/Azure/GCP secret stores) for production: stored with RBAC/audit/
  versioning, fetched at startup via IAM/OIDC (often through a sidecar injector), rotated centrally. Decouples
  secrets from code and images entirely.
- Always rotate frequently and avoid long-lived credentials. Secret injection reduces exposure but doesn't make
  a breached container's in-memory secrets safe.

## Runtime security monitoring (Falco)

Scanning/hardening guard against *known* issues at build/deploy; **runtime** tools catch *live* anomalies
(zero-days, drift, suspicious behavior). **Falco** watches kernel events (via eBPF) against YAML rules and
alerts when, e.g., a shell spawns in a web-server container, `/etc/passwd` is written, or unexpected network
activity occurs. Deploy as a daemon (DaemonSet in Kubernetes); enrich alerts with container/image metadata;
route them to webhooks/SIEM/Slack (Falcosidekick) for response. Falco **detects**, it doesn't block by itself —
pair with automated responses or admission policies. Trade-offs: tune rules to avoid noise, benchmark overhead,
prefer eBPF over kernel modules, and lock down Falco's own privileges.

## A hardened run command

```bash
docker run --rm \
  --read-only \
  --cap-drop ALL --cap-add NET_BIND_SERVICE \
  --security-opt no-new-privileges \
  --user 1000:1000 \
  --memory 256m --cpus 0.5 --pids-limit 200 \
  --tmpfs /tmp:rw,size=32m \
  -p 127.0.0.1:8080:8080 \
  my-app:1.0
```
Read-only rootfs, no extra capabilities, no privilege escalation, non-root, resource-limited, loopback-only
exposure, writable scratch via tmpfs.

## Hardening trade-offs

| Measure | Benefit | Watch out for |
|---|---|---|
| Non-root + drop caps | Less to abuse on escape | Low-port bind / chown may need a specific cap added back |
| `--no-new-privileges` | No setuid escalation | Rare apps that legitimately escalate break |
| Read-only rootfs | No tampering/backdoors | Apps writing temp/logs need explicit volumes/tmpfs |
| Distroless/slim | Smaller attack surface | No shell/tools → harder debugging |
| seccomp/AppArmor | Restrict syscalls/FS | Over-strict profiles cause subtle runtime failures; start from defaults |
| Resource limits | Prevents DoS/runaway | Too tight → throttling/OOM; measure first |
| User namespaces | Container-root ≠ host-root | Volume ownership/UID-mapping complexity |

In Kubernetes these map to `securityContext`/Pod Security Standards, RBAC, NetworkPolicies, and admission
control — see the **kubernetes-security** skill. This skill stops at the single-host container boundary.
