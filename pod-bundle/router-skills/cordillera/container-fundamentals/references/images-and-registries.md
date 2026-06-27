# Images & Registries

Building, tagging, naming, distributing, scanning, and signing OCI images.

## Contents
- [What an image is](#what-an-image-is)
- [Three ways to create an image](#three-ways-to-create-an-image)
- [Image naming and namespaces](#image-naming-and-namespaces)
- [Tags vs digests](#tags-vs-digests)
- [Building (build, buildx, BuildKit)](#building)
- [Multi-architecture images](#multi-architecture-images)
- [Tagging, pushing, pulling](#tagging-pushing-pulling)
- [Registries](#registries)
- [save / load (offline transfer)](#save--load-offline-transfer)
- [Inspecting images](#inspecting-images)
- [Vulnerability scanning](#vulnerability-scanning)
- [SBOMs](#sboms)
- [Signing and provenance](#signing-and-provenance)
- [Digest-based promotion in CI/CD](#digest-based-promotion-in-cicd)

## What an image is

An image is a stack of immutable, content-addressed **layers** plus a JSON config (entrypoint, env, exposed
ports, etc.). The storage driver (`overlay2`) merges the layers into one root filesystem via a union FS; a
running container adds a thin writable layer on top (copy-on-write). Three essential properties: **immutable**,
**one-to-many layers**, **self-contained** (everything the app needs to run). Layers are shared between images
and cached by hash, so a base pulled once is reused.

## Three ways to create an image

1. **Dockerfile build** (the right way — declarative, repeatable): see `dockerfile-authoring.md`.
2. **Interactive commit** (exploration/prototyping only): run a container, make changes, `docker container
   commit <c> my-image`. Inspect what changed with `docker container diff <c>` (A=added, C=changed, D=deleted).
   Not repeatable — don't ship this way.
3. **Load from tarball**: `docker image load -i image.tar` (see save/load below).

## Image naming and namespaces

Fully-qualified form:

```
[registry-host[:port]/][user-or-org/]name[:tag]
```

| Example | Meaning |
|---|---|
| `alpine` | Official `alpine:latest` from Docker Hub |
| `ubuntu:24.04` | Official ubuntu, tag 24.04, from Docker Hub |
| `hashicorp/vault` | Org `hashicorp`, image `vault`, `:latest`, Docker Hub |
| `ghcr.io/acme/web-api:1.0` | GitHub Container Registry, org acme, web-api v1.0 |
| `gcr.io/jdoe/sample-app:1.1` | Google registry, user jdoe |
| `registry.acme.com/engineering/web-app:1.0` | Private registry |

Rules: omit the registry → `docker.io` (Docker Hub); omit the tag → `:latest`; official images need no
user/org. **Official images** are curated and CVE-scanned by Docker — prefer them or verified-publisher images
as bases.

## Tags vs digests

- A **tag** is a mutable, human label (`:1.0`, `:latest`). It can be re-pointed to a different image later.
- A **digest** (`name@sha256:…`) is the immutable content hash of one exact image.
- `:latest` is the default and a production trap — it's a moving target. **Pin tags**, and for fully
  reproducible/auditable deployments reference images **by digest**.

## Building

```bash
docker build -t my-ubuntu .                 # Dockerfile in current dir, context = "."
docker build -t my-ubuntu -f Dockerfile.dev .   # alternate Dockerfile name
```
The trailing `.` is the build context (the dir sent to the builder — keep it small with `.dockerignore`).
BuildKit is the default engine: faster, parallel, cache/secret mounts, multi-arch, SBOM/provenance metadata.

`docker buildx` is the extended BuildKit front-end (multi-platform, multiple builders, remote cache):

```bash
docker buildx build --platform linux/amd64,linux/arm64 --push -t ghcr.io/acme/web:1.0 .
```

## Multi-architecture images

An image built for `amd64` won't run on `arm64` (and vice versa) — a real issue with Apple Silicon and
Raspberry Pi. Options:
- Build a **multi-platform image** with `buildx --platform …`; the manifest holds one variant per arch and the
  runtime auto-selects the right one. Multi-arch images must be **pushed** to a registry (local daemon can't
  hold a multi-arch manifest the same way).
- Build/pull a specific arch explicitly: `docker pull --platform=linux/amd64 ubuntu:24.04`.
- Note: a **Linux** container can't run on a Linux host if it was built for Windows, and you can't run a
  Windows container on Linux. Docker Desktop on Mac/Windows runs Linux containers in a Linux VM.

## Tagging, pushing, pulling

```bash
docker image tag alpine:latest ghcr.io/acme/alpine:1.0   # new reference, same image (no copy)
docker login ghcr.io                                     # authenticate
docker image push ghcr.io/acme/alpine:1.0
docker image pull alpine:3.21
```
Pushing creates the repository if absent (public or private). A private repo requires login + permissions to pull.

## Registries

- **Docker Hub** (`docker.io`) — default; public free, private paid; hosts official images.
- **GitHub Container Registry** (`ghcr.io`) — common in GitHub Actions pipelines.
- **AWS ECR**, **Azure ACR**, **Google GCR/Artifact Registry** — cloud-native; authenticate with the cloud
  CLI/IAM rather than static passwords.
- **JFrog Artifactory**, **Harbor** (self-hosted; Harbor integrates Cosign-based content trust).

## save / load (offline transfer)

An image is just a tarball; move it between hosts without a registry:
```bash
docker image save -o ./backup/my-alpine.tar my-alpine     # export
docker image load -i ./backup/my-alpine.tar               # import → "Loaded image: my-alpine:latest"
```
(Contrast `export`/`import`, which operate on a *container's* flattened filesystem and lose layer/metadata.)

## Inspecting images

```bash
docker image ls                       # list images + sizes
docker image history my-image         # layers and how they were built
docker image inspect --format='{{json .Config.Volumes}}' mongo:8.0.8 | jq .   # e.g. declared volumes
```

## Vulnerability scanning

Scan images for known CVEs in OS packages and language deps. Popular tools: **Trivy** (Aqua), **Docker Scout**,
**Grype**, Clair, Snyk.

```bash
trivy image my-app:latest
# CI gate: fail the build on serious findings
trivy image --severity CRITICAL,HIGH --exit-code 1 my-app:latest
trivy image my-app:latest -f json -o trivy-result.json    # machine-readable for dashboards
```
Re-scan **regularly** — new CVEs are disclosed against images you already shipped. You can allow-list specific
CVEs (false positives / mitigated), but only with documented justification and an expiry.

## SBOMs

A Software Bill of Materials inventories every package/library/version in an image — needed to answer "am I
affected by CVE-X?" and increasingly a compliance requirement (NIST SP 800-218, EO 14028). Generate with
**Syft**:
```bash
syft my-app:latest                                   # human-readable
syft my-app:latest -o cyclonedx-json > sbom.json     # store alongside the image / as CI artifact (SPDX also supported)
```

## Signing and provenance

Scanning proves *what's inside*; signing proves *who built it and that it wasn't tampered with*. Use **Cosign**
(Sigstore). Image must be in a registry and ideally not a floating `:latest`.

```bash
cosign generate-key-pair                                   # cosign.key (private), cosign.pub (public)
cosign sign --key cosign.key ghcr.io/acme/app:v1.0.0
cosign verify --key cosign.pub ghcr.io/acme/app:v1.0.0     # "Verified OK"
cosign sign -a git-commit=abc123 --key cosign.key ghcr.io/acme/app:v1.0.0   # add provenance annotations
```
Cosign also supports keyless (OIDC) and KMS-backed flows and writes to transparency logs. Enforce at
deploy time (Kubernetes admission controllers such as Kyverno/Connaisseur, or Harbor pull policies) so only
signed images run.

## Digest-based promotion in CI/CD

Build once, then promote the **exact same artifact** by digest through stages — never rebuild per environment:
```bash
IMG=ghcr.io/acme/app:${GIT_SHA}
docker push "$IMG"
DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' "$IMG")   # e.g. ghcr.io/acme/app@sha256:…
# deploy $DIGEST to staging, then the identical $DIGEST to production
```
A tag can be re-pointed; a digest can't — digests give content immutability and auditable releases.

Deploying these images to a Kubernetes cluster, image pull secrets, and admission-time signature enforcement
are covered by the **kubernetes-workloads** / **kubernetes-security** skills.
