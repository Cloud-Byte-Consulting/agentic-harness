# Dockerfile Authoring

Practical reference for writing correct, fast-building, small Dockerfiles.

## Contents
- [Instruction reference](#instruction-reference)
- [ENTRYPOINT vs CMD](#entrypoint-vs-cmd)
- [Exec form vs shell form (and PID 1)](#exec-form-vs-shell-form-and-pid-1)
- [HEALTHCHECK](#healthcheck)
- [Layer caching](#layer-caching)
- [.dockerignore](#dockerignore)
- [Multi-stage builds](#multi-stage-builds)
- [Minimal final images: slim, alpine, distroless, scratch](#minimal-final-images)
- [BuildKit cache and secret mounts](#buildkit-cache-and-secret-mounts)
- [Worked examples](#worked-examples)
- [Containerizing a legacy app (lift and shift)](#containerizing-a-legacy-app-lift-and-shift)

## Instruction reference

| Instruction | Purpose | Notes |
|---|---|---|
| `FROM image:tag [AS name]` | Base image; starts a build stage | Every Dockerfile starts here. `FROM scratch` = empty base (no layer). Pin the tag. |
| `RUN <cmd>` | Execute a command at **build** time → new layer | Chain with `&&`, clean caches in the same `RUN`. |
| `COPY src dst` | Copy files from build context into image | Preferred over `ADD`. Supports `--chown=uid:gid`, `--from=<stage>`. |
| `ADD src dst` | Like COPY but also untars local tarballs and fetches URLs | Use `COPY` unless you specifically need ADD's extras. |
| `WORKDIR /path` | Set working dir for later instructions and runtime | Persists across layers; `RUN cd /x` does **not**. Create-if-absent. |
| `ENV KEY=val` | Default env var in image and at runtime | Overridable with `--env` at run. |
| `ARG KEY[=default]` | Build-time variable | Set with `--build-arg`. An `ARG` before `FROM` can parameterize the base. |
| `EXPOSE port[/proto]` | Document a listening port | **Does not publish.** Use `-p` at run time. |
| `VOLUME ["/path"]` | Declare a mount point for a managed volume at runtime | Path is excluded from the union FS. |
| `USER user[:group]` | Run subsequent steps and the container as this user | Create the user first (`adduser`/`useradd`). |
| `ENTRYPOINT [...]` | The fixed executable | See below. |
| `CMD [...]` | Default args (or default command) | See below. |
| `HEALTHCHECK ...` | Define a liveness probe | See below. |
| `LABEL k=v` | Image metadata | e.g. `org.opencontainers.image.source`. |
| `STOPSIGNAL sig` | Signal sent to stop the container | Default SIGTERM. |

## ENTRYPOINT vs CMD

Think of a command line as `<command> <args>`. **`ENTRYPOINT` is the command; `CMD` is the default args.**

```dockerfile
FROM alpine:3.21
ENTRYPOINT ["ping"]
CMD ["-c", "3", "8.8.8.8"]
```

- `docker run img` → runs `ping -c 3 8.8.8.8`.
- `docker run img -w 5 127.0.0.1` → args after the image **override CMD** → `ping -w 5 127.0.0.1`.
- Override ENTRYPOINT itself with `docker run --entrypoint sh img`.

If `ENTRYPOINT` is omitted, it defaults to `/bin/sh -c` and `CMD`'s value is passed as the command string.
So `CMD wget -O - http://x` actually runs `/bin/sh -c "wget -O - http://x"` (a shell child process). For a
simple single-command image, putting the whole command in `CMD` (exec form) is fine; the ENTRYPOINT+CMD split
is the idiomatic choice when you want a fixed program with overridable arguments.

## Exec form vs shell form (and PID 1)

- **Exec form** (JSON array): `CMD ["node", "app.js"]`. The process is `node` directly → it is PID 1 and
  receives signals (SIGTERM on `docker stop`). **Prefer this.**
- **Shell form**: `CMD node app.js`. Runs as `/bin/sh -c "node app.js"`. The shell is PID 1 and frequently
  does **not** forward SIGTERM to its child, so `docker stop` waits the full 10s grace period and then SIGKILLs.

PID 1 also doesn't reap zombie children by default. If your process spawns children or you see the slow-stop
behavior, add an init: `docker run --init …`, or bake in `tini` and use it as the ENTRYPOINT.

## HEALTHCHECK

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
```

Docker marks the container `healthy`/`unhealthy`; orchestrators and `depends_on: condition: service_healthy`
in Compose use this. Keep the probe cheap and dependency-free. `--start-period` gives slow-starting apps grace
before failures count.

## Layer caching

Each instruction that changes the filesystem is a cached layer keyed by its content and the previous layer.
**If a layer changes, every layer after it rebuilds.** Order from least- to most-frequently-changing.

The classic dependency-caching pattern — copy the dependency manifest and install **before** copying source:

```dockerfile
FROM node:23-bookworm
WORKDIR /app
COPY package.json /app/      # changes rarely
RUN npm install              # expensive — cached unless package.json changes
COPY . /app                  # source changes often, AFTER the install
CMD ["npm", "start"]
```

Reversing lines 3–5 (`COPY . /app` before `npm install`) re-runs `npm install` on *every* source edit.

Minimize layer count by chaining and cleaning in one `RUN`:

```dockerfile
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates curl \
 && rm -rf /var/lib/apt/lists/*
```

## .dockerignore

Keep the build context (and thus the image) lean and builds fast. Same syntax as `.gitignore`:

```
.git
node_modules
*.log
**/__pycache__
.env
Dockerfile
```

## Multi-stage builds

Compile in a fat builder stage; copy only the artifact into a minimal final stage. Dramatically smaller images
and a far smaller attack surface (no compilers/SDK in production).

C → `scratch` (260 MB toolchain image shrinks to ~136 KB):
```dockerfile
FROM alpine:3.21 AS build
RUN apk add --update alpine-sdk
WORKDIR /app
COPY . /app
RUN gcc -static -O2 hello.c -o /app/hello   # static so runtime needs no libs

FROM scratch
COPY --from=build /app/hello /app/hello
ENTRYPOINT ["/app/hello"]
```

Go → `scratch`:
```dockerfile
FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /hello

FROM scratch
COPY --from=builder /hello /hello
ENTRYPOINT ["/hello"]
```

`COPY --from=<stage>` (or `--from=<external-image>`) pulls files from any earlier stage or image. You can have
many stages and even target one with `docker build --target build`.

## Minimal final images

Trade-off ladder, smallest/hardest-to-debug last:

1. **`-slim`** (e.g. `python:3.12-slim`, Debian-based, no extras) — easy, still has a shell and package manager.
2. **`-alpine`** (musl libc, ~5–8 MB base) — very small; watch for musl-vs-glibc incompatibilities with some
   binaries/wheels.
3. **distroless** (`gcr.io/distroless/*`) — runtime + your app, **no shell, no package manager**. Small attack
   surface; debugging needs an ephemeral/sidecar container.
4. **`scratch`** — completely empty; only for statically linked single binaries.

Fewer packages = smaller download/disk/memory and fewer CVEs. Use `.dockerignore`, avoid `--install-recommends`,
and don't install debug tools in the production stage.

## BuildKit cache and secret mounts

BuildKit is the default builder (`docker build`/`docker buildx`). Two high-value features:

**Cache mount** — persist a package cache across builds without baking it into a layer:
```dockerfile
# syntax=docker/dockerfile:1
RUN --mount=type=cache,target=/root/.cache/pip pip install -r requirements.txt
```

**Build secret** — expose a credential to one `RUN` only; it never lands in a layer or `docker history`:
```dockerfile
# syntax=docker/dockerfile:1
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc npm ci
```
```bash
docker build --secret id=npmrc,src=$HOME/.npmrc -t app .
```
This is the correct way to pull private dependencies at build time — far safer than `ARG`/`ENV` (which persist).

## Worked examples

Node.js app (cache-friendly, exec form):
```dockerfile
FROM node:23-bookworm
WORKDIR /app
COPY package.json /app/
RUN npm install
COPY . /app
ENTRYPOINT ["npm"]
CMD ["start"]
```

Static nginx site:
```dockerfile
FROM nginx:alpine
COPY . /usr/share/nginx/html
```

Ubuntu with a tool, exec-form entrypoint:
```dockerfile
FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends iputils-ping \
 && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["ping"]
CMD ["127.0.0.1"]
```

## Containerizing a legacy app (lift and shift)

A repeatable process for wrapping an existing (Java/.NET/Python/etc.) app without a rewrite:

1. **Inventory external dependencies** — databases & connection strings, external APIs/keys, message buses.
2. **Gather source + build instructions** into one project root (this becomes the build context). Record the
   exact build command (Maven/MSBuild/make).
3. **Classify configuration**: build-time (needed to build the image), environment (varies dev/stage/prod,
   injected at container start), runtime (e.g. secrets fetched while running).
4. **Handle secrets externally** — never hardcode or default them in `ENV`; pull from a store or runtime secret.
5. **Author the Dockerfile**: choose a base matching the runtime; `WORKDIR /app` + `COPY . .`; `RUN <build>`;
   `ENV` defaults (no secret defaults); `EXPOSE` the listening ports; define the start command (`ENTRYPOINT`/
   `CMD`, or a `docker-entrypoint.sh` for pre-run setup — make it executable with `chmod +x`).
6. Iterate until the image builds and behaves identically to the legacy deployment, then refactor incrementally.

Reported payoff from real migrations: ~50% lower maintenance cost and up to ~90% shorter release cycles, with
no application-logic rewrite.

Once the image is built, deploying it to a cluster (Pods, Deployments) is covered by the **kubernetes-workloads** skill.
