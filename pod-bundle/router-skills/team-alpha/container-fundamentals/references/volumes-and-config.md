# Volumes & Configuration

Persisting data and injecting configuration/secrets into containers.

## Contents
- [Why containers need volumes](#why-containers-need-volumes)
- [Named volumes](#named-volumes)
- [Bind mounts](#bind-mounts)
- [tmpfs mounts](#tmpfs-mounts)
- [Read-only mounts and sharing data](#read-only-mounts-and-sharing-data)
- [The VOLUME instruction](#the-volume-instruction)
- [Volume drivers](#volume-drivers)
- [Inspecting and cleaning up volumes](#inspecting-and-cleaning-up-volumes)
- [Configuration: environment variables](#configuration-environment-variables)
- [ARG vs ENV vs --build-arg vs --env](#arg-vs-env-vs---build-arg-vs---env)
- [Secrets — what NOT to do, and what to do](#secrets)
- [Stateful container patterns](#stateful-container-patterns)

## Why containers need volumes

A running container has one thin writable layer on top of the image's read-only layers. **Removing the
container deletes that layer and all its writes.** `docker container diff <c>` shows the changes (A/C/D). For
any data that must outlive a container — databases, uploads, logs you want to keep — mount external storage.
The mounted path is **excluded from the union FS**, so writes there go to the backing store, not the doomed
container layer.

## Named volumes

Docker-managed storage, the **preferred** mechanism for persistence. Portable, easy to back up, survives
container removal.

```bash
docker volume create sample
docker volume inspect sample          # shows Mountpoint, e.g. /var/lib/docker/volumes/sample/_data (Linux)
docker container run -it -v sample:/data alpine /bin/sh      # mount into a container
docker volume ls
```
On Docker Desktop (Mac/Windows) the data lives inside the Linux VM, so the Linux `Mountpoint` path isn't
directly browsable from the host — that's expected.

## Bind mounts

Mount a host file/directory into the container. Great for **development live-reload** (edit on host → changes
appear in the container, and vice versa). Less portable (depends on host paths/OS) and a security
consideration (container can touch host files).

```bash
docker container run --rm -it -v "$(pwd)":/usr/share/nginx/html -p 8080:80 my-website:1.0
docker container run --rm -it -v "$(pwd)/src":/app/src alpine /bin/sh
```
Always use **absolute paths** for the host side (`$(pwd)` on Unix; on Windows you may be prompted to share the
drive). The bind mount **replaces** whatever was at the container target path with the host content.

## tmpfs mounts

In-memory, never written to disk — for transient/sensitive scratch data:
```bash
docker run --tmpfs /tmp:rw,size=64m my-app
```
Also useful to provide writable scratch when running with `--read-only` (see container-security.md).

## Read-only mounts and sharing data

Multiple containers can share one volume. To avoid concurrent-write/race-condition problems, let one container
write and others mount **read-only** with `:ro`:

```bash
docker run -it --name writer -v shared-data:/data alpine /bin/sh        # read/write
docker run -it --name reader -v shared-data:/app/data:ro ubuntu /bin/bash  # read-only
# a write in 'reader' fails: "Read-only file system"
```

## The VOLUME instruction

Image authors declare mount points so data is persisted even if the user forgets `-v` (e.g. database images):
```dockerfile
VOLUME /app/data
VOLUME ["/data/db", "/data/configdb"]
```
At run time Docker auto-creates an anonymous volume for each declared path and mounts it (excluded from the
union FS). Operators are then responsible for backing up the backing store. Inspect what an image declares:
`docker image inspect --format='{{json .Config.Volumes}}' mongo:8.0.8 | jq .`

## Volume drivers

The default `local` driver stores on the host filesystem (single-host/dev). Plugin drivers back volumes with
networked/cloud storage for distributed setups — select with `--driver`:

| Driver | Backing | Use |
|---|---|---|
| `local` | Host FS | Default; single-host, dev |
| `nfs` / `cifs` | NFS / SMB shares | Shared storage across hosts |
| `rexray/ebs`, `gce-pd`, `azurefile` | Cloud block/file | Cloud persistence |
| `portworx`, `glusterfs`, `rclone` | SDS / cloud object | Enterprise: replication, snapshots, encryption |

For multi-host persistence in production, the orchestrator's storage abstraction (PVs/PVCs, StatefulSets) is
usually the better answer — see the **kubernetes-storage** skill.

## Inspecting and cleaning up volumes

```bash
docker volume inspect my_volume        # Mountpoint, driver, labels
docker volume rm sample                # fails if a container still uses it
docker volume prune                    # remove all unused volumes (destructive)
docker container rm -v <c>             # also remove the container's anonymous volumes
docker run --rm ...                    # --rm doesn't remove named volumes, only the container
```
Removing a volume **irreversibly destroys its data** — back up first.

## Configuration: environment variables

Each container is a sandbox with its own env — what you see on the host is *not* what the container sees.

```bash
docker run --rm -it --env LOG_DIR=/var/log/my-log alpine /bin/sh   # single var
docker run --rm -it -e A=1 -e B=2 -e C=3 alpine /bin/sh            # multiple
docker run --rm -it --env-file ./dev.env alpine /bin/sh           # many vars from a file
```
`dev.env` is one `KEY=value` per line:
```
LOG_DIR=/var/log/my-log
MAX_LOG_FILES=5
MAX_LOG_SIZE=1G
```
`--env-file` is the clean way to handle "my app needs 30 config vars."

## ARG vs ENV vs --build-arg vs --env

| Mechanism | When it applies | Set by | Notes |
|---|---|---|---|
| `ARG name[=default]` | **Build time only** | `--build-arg name=val` | Parameterize the build (e.g. base version). Not present at runtime. |
| `ENV KEY=val` | Build **and** run time | Dockerfile | Image default; visible at runtime. |
| `--env` / `-e` | **Run time** | `docker run` | Overrides image `ENV`. |
| `--env-file` | **Run time** | `docker run` | Bulk runtime vars. |

```dockerfile
ARG BASE_IMAGE_VERSION=20-bookworm
FROM node:${BASE_IMAGE_VERSION}
ENV LOG_DIR=/var/log/my-log
```
```bash
docker build --build-arg BASE_IMAGE_VERSION=20-alpine -t app .   # build-time override
docker run --rm -it --env LOG_DIR=/tmp/logs app                  # runtime override of the ENV default
```
Rule of thumb: `ARG`/`--build-arg` parameterize the **image build**; `--env`/`--env-file` configure the
**running app**. Image-baked `ENV` provides defaults you can override at run.

## Secrets

**Never** bake secrets into images or `ENV`: they persist in layers, leak via `docker history`/`docker
inspect`, travel everywhere the image goes, and force a rebuild to rotate. (A 2023 study found ~8.5% of public
images leaked secrets.)

- **Build-time secrets** (private repo creds during build): BuildKit `RUN --mount=type=secret` — exposed to one
  `RUN` as a file, never persisted. See `dockerfile-authoring.md`.
- **Runtime secrets, file-based** (Docker/Compose secrets) — mounted at `/run/secrets/<name>`; the app reads
  the **file**, not an env var:
  ```yaml
  services:
    web:
      image: hello-ruby:latest
      secrets: [db_password]
      environment:
        DB_PASSWORD_FILE: /run/secrets/db_password
  secrets:
    db_password:
      file: ./db_password.txt
  ```
  ```ruby
  password = File.read("/run/secrets/db_password").strip
  ```
  Compose secrets are file-mounted but lack encryption-at-rest/rotation. Swarm secrets add encryption and only
  work with `docker service` (not plain `docker run`).
- **External secret managers** (HashiCorp Vault, AWS/Azure/GCP secret stores) for production — fetched at
  startup via IAM/OIDC, often through a sidecar; centralized rotation, RBAC, and audit. Full treatment in
  `container-security.md`.

## Stateful container patterns

- **Named volumes** — manage storage independently of any container; the modern default for persistence and
  sharing.
- **Data volume containers** — older pattern (a container that exists only to hold a volume); superseded by
  named volumes.
- **Bind mounts** — dev-time code mounting; not for production persistence.
- **Volume plugins** — remote/cloud backends for portability across environments.
- In orchestration, **StatefulSets** + per-Pod **PersistentVolumes** provide stable identity and storage —
  out of scope here; see the **kubernetes-storage** / **kubernetes-workloads** skills.

Best practices: prefer Docker-managed volumes over bind mounts for persistence; back up volumes regularly;
monitor disk usage (a full volume crashes the app); secure/encrypt sensitive data; use volume drivers to reach
external storage.
