---
name: container-fundamentals
description: >-
  Build, ship, secure, and debug OCI containers and images on a single host - the layer
  beneath Kubernetes. Use for Dockerfiles and image builds (multi-stage, layer caching,
  distroless/scratch, ENTRYPOINT vs CMD, BuildKit/buildx, multi-arch), tags vs digests,
  registries and push/pull, image scanning and signing (Trivy, Cosign, SBOM), data volumes and
  bind mounts, container networking and port publishing, Docker Compose for local
  multi-service dev, container logging, and debugging running containers (exec, logs, inspect,
  ephemeral/netshoot). Also covers container security: rootless, non-root users, dropped
  capabilities, read-only rootfs, seccomp/AppArmor, Podman. Trigger whenever the user writes
  or fixes a Dockerfile, builds or optimizes an image, hits a CrashLoop/exit-code/OOM in a
  container, sets up Compose, or asks about container hardening - even if they never say
  Docker or Kubernetes. For deploying images to Kubernetes, see kubernetes-workloads.
---

# Container Fundamentals

This skill equips Claude to build correct, small, secure container images and run them competently on a single
host — authoring Dockerfiles and Compose files, managing images/registries, wiring volumes/config/networking,
and debugging/hardening running containers. It is the **container layer**: it stops at the point where images
are handed to an orchestrator.

## When to use this skill

- Writing or reviewing a **Dockerfile** (instructions, layer caching, multi-stage, distroless/scratch, ENTRYPOINT vs CMD, HEALTHCHECK).
- **Building & shipping images**: `docker build`/`buildx`, tagging, fully-qualified names, `push`/`pull`, registries (Docker Hub, GHCR, ECR/ACR/GCR), `save`/`load`.
- Shrinking image size or fixing slow/broken **build caching**.
- Wiring **data**: named volumes, bind mounts, `tmpfs`, `VOLUME`, sharing data between containers, read-only mounts.
- Injecting **configuration/secrets**: `--env`/`--env-file`, `ENV`/`ARG`/`--build-arg`, build secrets, Docker/Compose secrets.
- **Single-host networking**: bridge/host/none, publishing ports (`-p`/`-P`), container DNS, isolating containers on separate networks, reverse proxy.
- Authoring **Docker Compose** for local multi-service dev (services, `depends_on` + healthcheck, scaling, overrides, `include`).
- **Logging & monitoring**: log drivers + rotation, shipping logs (Filebeat/Elastic), exposing/scraping metrics (Prometheus + Grafana).
- **Debugging** a container that won't start, crashes, or misbehaves (`logs`, `exec`, `inspect`, ephemeral debug containers, netshoot).
- **Securing** containers: non-root user, dropped capabilities, read-only rootfs, vulnerability scanning, SBOM, image signing/provenance.
- Using **Podman** as a daemonless, rootless drop-in for Docker.

For Kubernetes **Pods/Deployments/Services/Ingress**, autoscaling, or cluster networking, say so and defer to the
sibling `kubernetes-workloads` / `kubernetes-networking` / `kubernetes-autoscaling` skills. This skill ends at
"I have a built, scanned, signed image in a registry."

## Core concepts (the mental model)

**A container is a process, not a tiny VM.** It is one or more Linux processes isolated by **namespaces**
(pid, net, mnt, uts, ipc, user — each container gets its own view of process tree, network stack, filesystem,
hostname) and constrained by **cgroups** (CPU, memory, PIDs, I/O limits — prevents the "noisy neighbor"
problem). All containers on a host share the **host kernel**; this is why startup is milliseconds, not seconds,
and why a Linux container needs a Linux kernel (on macOS/Windows, Docker Desktop runs a lightweight Linux VM).
Inside its pid namespace the main process is **PID 1** — which has signal-handling consequences (see pitfalls).

**Images are immutable, layered, content-addressed tarballs.** An image is a stack of read-only layers (a
union/overlay filesystem, today `overlay2`). Each Dockerfile instruction that changes the filesystem adds a
layer holding only the *delta*. Layers are shared across images and cached by content hash; a running container
adds one thin **writable layer** on top (copy-on-write). **Everything written to that layer is lost when the
container is removed** — this is why containers are *ephemeral* and why persistent data needs volumes. Images
follow the **OCI** spec, so any conforming runtime (containerd, CRI-O, Podman) can run a Docker-built image.

**The runtime stack:** the Docker CLI talks to `dockerd` over a REST API; `dockerd` delegates to **containerd**
(image management, lifecycle, the reference OCI runtime) which in turn calls **runc** (the low-level tool that
actually creates the namespaced/cgrouped process). Kubernetes talks to containerd/CRI-O directly via the CRI —
it does **not** need Docker (the "Docker is removed" change), but Docker-built OCI images still run everywhere.

**Tags vs digests.** A tag (`myapp:1.0`, `:latest`) is a mutable pointer; a digest (`myapp@sha256:…`) is
immutable. Omitting a tag means `:latest`, which is a trap in production — pin versions, and for fully
reproducible/auditable deploys reference images **by digest**.

A fully-qualified image name is `[registry/][user-or-org/]name[:tag]` — e.g. `ghcr.io/acme/web-api:1.0`.
Omit the registry → Docker Hub (`docker.io`); omit the org for an official image (`alpine`, `nginx`).

## Workflow / how to approach container tasks

### Authoring a Dockerfile
1. **Pick the smallest correct base.** Prefer `-slim`/`-alpine` variants or a language's official image; use
   **multi-stage** to compile in a fat builder and copy only artifacts into a minimal final stage
   (`distroless` or `scratch` for static binaries). Pin the tag (`python:3.12-slim`, not `python`).
2. **Order instructions cache-friendly.** Put rarely-changing layers first. Copy *dependency manifests*
   (`package.json`, `requirements.txt`, `go.mod`, `pom.xml`) and install deps **before** copying source — a
   source edit then won't bust the expensive dependency layer. Add a `.dockerignore`.
3. **Minimize layers and size.** Chain related `RUN` commands with `&&` and clean caches in the same layer
   (`apt-get update && apt-get install -y --no-install-recommends X && rm -rf /var/lib/apt/lists/*`).
4. **Set the start command correctly.** Use the **exec form** (`CMD ["node","app.js"]`) so signals reach the
   process. `ENTRYPOINT` = the fixed command, `CMD` = default/overridable args. Add a non-root `USER` and a
   `HEALTHCHECK`.
5. **Build & verify**: `docker build -t name:tag .` (or `docker buildx build` for multi-platform/BuildKit
   features), then run and exercise it.

Full instruction reference, multi-stage patterns, distroless/scratch, and BuildKit secrets/cache mounts are in
`references/dockerfile-authoring.md`.

### Building & shipping images
1. Build and tag: `docker build -t acme/web:1.0 .`
2. Tag for a registry: `docker image tag acme/web:1.0 ghcr.io/acme/web:1.0`
3. Authenticate (`docker login ghcr.io`) and `docker push ghcr.io/acme/web:1.0`.
4. For CI/CD and cross-arch: `docker buildx build --platform linux/amd64,linux/arm64 --push -t … .`
5. Scan before publishing; sign after. Promote by **digest**, not tag.

Tagging rules, registries, multi-arch/buildx, `save`/`load`, scanning gates, and signing live in
`references/images-and-registries.md`.

### Data & configuration
- **Named volume** (Docker-managed, preferred for persistence): `-v mydata:/var/lib/postgresql/data`.
- **Bind mount** (host path, for dev live-reload): `-v "$(pwd)":/app` — host changes appear instantly.
- Mount **read-only** for consumers: `-v shared:/data:ro`.
- Config via env: `--env KEY=val`, or `--env-file dev.env` for many vars. `ENV` sets image defaults;
  `--env` overrides at runtime. `ARG`/`--build-arg` are build-time only.
- **Never** bake secrets into images or `ENV`. Use build secrets (`RUN --mount=type=secret`) at build time and
  file-based runtime secrets (Docker/Compose secrets mounted at `/run/secrets/…`) at run time.

Details, volume drivers, the `VOLUME` instruction, and stateful patterns: `references/volumes-and-config.md`.

### Single-host networking & Compose
- Default **bridge** isolates containers; create **user-defined bridge** networks so containers resolve each
  other by name via Docker's embedded DNS, and to firewall groups apart. **host** shares the host stack (Linux
  only, security trade-off); **none** = no networking.
- **Publish** ports to expose a service: `-p 8080:80` (host:container); `EXPOSE` is documentation only.
- For local multi-service dev, write a `compose.yaml` and `docker compose up`. Use `depends_on` with a
  healthcheck `condition: service_healthy`, scale with `--scale`, layer environments with overrides or `include`.

Bridge/host/none deep dive, port publishing patterns, reverse proxy, and the full Compose guide:
`references/networking-and-compose.md`.

### Debugging a container
1. `docker logs --tail 50 -f <c>` — first stop; apps should log to STDOUT/STDERR.
2. `docker ps -a` — check STATUS/exit code (137=OOM/SIGKILL, 0=clean exit, 125=daemon error).
3. `docker inspect <c>` — mounts, env, network, the exact command, restart reason.
4. `docker exec -it <c> sh` — poke inside a *running* container (starts a new process; does not disturb PID 1).
5. For **distroless/scratch** (no shell) or to avoid touching the image, attach an **ephemeral debug
   container** sharing the target's namespaces, or run `nicolaka/netshoot` with `--network container:<c>`.

Exit-code table, `exec` vs `attach`, inspect filters, and live code-debugging are in
`references/debugging-containers.md`.

### Securing & hardening
Work the lifecycle: **build** (minimal pinned base, non-root `USER`, multi-stage so build tools don't ship) →
**ship** (scan with Trivy/Scout/Grype; generate an SBOM with Syft; sign with Cosign; gate CI on CRITICAL/HIGH)
→ **run** (`--read-only`, `--cap-drop ALL` then add only what's needed, `--no-new-privileges`,
`--user`, resource limits, seccomp/AppArmor). Never embed secrets.

Full hardening flags, scanning/SBOM/signing workflows, and trade-offs: `references/container-security.md`.

## Common pitfalls & anti-patterns

- **Expecting writes to survive container removal.** The writable layer is discarded on `docker rm`. Persist
  state in a **volume**; mount the volume's target path (it's excluded from the union FS).
- **`:latest` in production.** Mutable and non-reproducible. Pin tags; deploy by digest for audit trails.
- **Busting the cache on every build.** `COPY . .` *before* `RUN <install deps>` re-installs dependencies on
  every source change. Copy manifests and install first, then copy the rest.
- **Shell-form `CMD`/`ENTRYPOINT` swallowing signals.** Shell form wraps the process in `/bin/sh -c`, which
  often doesn't forward SIGTERM, so the container takes the full 10s timeout to stop and then gets SIGKILLed.
  Use **exec form** (JSON array). For PID-1 signal/zombie reaping, add an init (`docker run --init` or `tini`).
- **Running as root.** Most base images default to root; an escape then has root on the host's kernel
  surface. Add a non-root `USER` and `--cap-drop ALL`.
- **Baking secrets into images or `ENV`.** They persist in layers and leak via `docker history`/`inspect`.
  Use build secrets and runtime file secrets.
- **`--network host` for convenience.** Removes network isolation; on Docker Desktop it attaches to the VM,
  not your real host, so it doesn't even do what people expect. Use a user-defined bridge and publish ports.
- **Confusing `EXPOSE` with publishing.** `EXPOSE` only documents intent; you still need `-p`/`--publish` to
  reach the service.
- **Two containers on the same host port.** Each published host port is exclusive — give each a unique host
  port, or front them with a single reverse proxy.
- **Treating containers like VMs / one giant image with everything.** Keep one concern per container; compose
  multiple services rather than `RUN`-installing a database, web server, and cron into one image.
- **Logging to files inside the container.** They vanish with the container and aren't picked up by `docker
  logs` or log shippers. Log to STDOUT/STDERR.

## Reference files

- **`references/dockerfile-authoring.md`** — Read when writing/reviewing a Dockerfile: every instruction,
  exec vs shell form, multi-stage builds, distroless/scratch, layer caching, `.dockerignore`, HEALTHCHECK,
  BuildKit cache/secret mounts, the lift-and-shift legacy-app process.
- **`references/images-and-registries.md`** — Read for building, tagging, naming, registries (Docker Hub,
  GHCR, ECR/ACR/GCR), push/pull, `buildx`/multi-arch, `save`/`load`, vulnerability scanning, SBOMs, signing,
  and digest-based promotion.
- **`references/volumes-and-config.md`** — Read for volumes vs bind mounts vs tmpfs, the `VOLUME` instruction,
  sharing data, read-only mounts, env/`--env-file`, `ARG`/`ENV`/`--build-arg`, and stateful patterns.
- **`references/networking-and-compose.md`** — Read for the container network model, bridge/host/none,
  port publishing, container DNS, reverse proxy (Traefik), and the complete Docker Compose guide.
- **`references/debugging-containers.md`** — Read for logs/exec/inspect/attach, exit codes, ephemeral debug
  containers, netshoot, log drivers + rotation, and Prometheus/Grafana monitoring.
- **`references/container-security.md`** — Read for hardening flags, rootless/Podman, capabilities, read-only
  rootfs, seccomp/AppArmor, scanning (Trivy/Scout/Grype), SBOM (Syft), signing/provenance (Cosign/Sigstore),
  and secrets management.
