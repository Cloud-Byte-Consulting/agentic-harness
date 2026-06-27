---
name: kubernetes-gitops-cicd
description: >-
  Deliver applications to Kubernetes with GitOps and CI/CD. Use for GitOps principles
  (declarative, Git-versioned, pulled, continuously reconciled, drift detection), Argo CD
  (Application, ApplicationSet, app-of-apps, sync waves and hooks, projects) and Flux
  (sources, Kustomization, HelmRelease, image automation), packaging with Helm (charts,
  values, releases, OCI) and Kustomize (bases, overlays, patches), repo structure and
  environment promotion, progressive delivery with Argo Rollouts and Flagger (canary,
  blue-green, analysis), secrets in Git (Sealed Secrets, SOPS, External Secrets Operator), and
  CI/CD pipeline design (build/test/scan/deploy stages, deployment strategies, anti-patterns)
  with GitHub Actions, GitLab CI, and Azure Pipelines. Trigger whenever the user sets up Argo
  CD or Flux, writes Helm charts or Kustomize overlays, builds a deploy pipeline, manages
  config across environments, or does canary/blue-green releases - even without saying GitOps.
  Cluster bootstrap is in kubernetes-cluster-operations.
---

# Kubernetes GitOps & CI/CD

This skill equips Claude to design and operate GitOps delivery on Kubernetes: write correct Argo CD / Flux
manifests, structure Helm charts and Kustomize overlays, build CI pipelines that feed a pull-based CD agent,
roll out changes safely with progressive delivery, and keep secrets out of plaintext Git — while avoiding the
classic anti-patterns.

## When to use this skill

- Authoring or debugging **Argo CD** resources: `Application`, `ApplicationSet`, `AppProject`, sync policies,
  sync waves/hooks, app-of-apps, RBAC, `OutOfSync`/health issues.
- Authoring or debugging **Flux** resources: `GitRepository`, `OCIRepository`, `Kustomization`, `HelmRelease`,
  `HelmRepository`, image automation, `flux bootstrap`.
- Writing or fixing **Helm** charts (`Chart.yaml`, `values.yaml`, `templates/`, dependencies/umbrella charts,
  OCI registries) or choosing **Helm vs Kustomize**.
- Structuring **Kustomize** bases/overlays, patches, and generators for multi-environment config.
- Designing **repo strategy**: app repo vs config repo, monorepo vs polyrepo, environments via overlays vs
  branches, and promotion flows (dev → staging → prod).
- **Progressive delivery**: Argo Rollouts or Flagger canary / blue-green with metric analysis gates.
- **Secrets in GitOps**: SOPS, Sealed Secrets, External Secrets Operator.
- **CI pipelines** (GitHub Actions, Azure Pipelines, GitLab CI) that build/test/scan/push an image and then
  hand off to a GitOps agent — including pipeline-as-YAML, templates, and image-tag promotion.
- **GitOps at scale & multitenancy**: many clusters, AppProjects, Cluster generators, vCluster, instance topologies.
- Choosing **CI/CD design patterns** and avoiding **anti-patterns**.

**Boundaries (cross-link, don't duplicate):** cluster bootstrap/install and node/control-plane ops →
`kubernetes-cluster-operations`. Observability tooling (Prometheus/Grafana/Loki setup, dashboards) →
`kubernetes-observability`. Building/optimizing container images and Dockerfiles → `container-fundamentals`.
This skill consumes those (a pipeline *builds* an image and a Rollout *reads* Prometheus metrics) but does not own them.

## Core concepts

**GitOps in one sentence.** The desired state of the system is declared in Git (declarative + versioned + immutable),
and an in-cluster **agent continuously reconciles** the live cluster to match Git, detecting and (optionally)
auto-correcting drift. Git is the single source of truth; the cluster is a read-replica that heals itself.

The four GitOps principles (GitOps Working Group):
1. **Declarative** — the whole system is described declaratively, not as imperative scripts.
2. **Versioned & immutable** — desired state is stored in Git, versioned, and immutable (change = new commit, not in-place edit).
3. **Pulled automatically** — software agents automatically pull the declared state from Git.
4. **Continuously reconciled** — agents continuously observe actual state and reconcile it toward desired state.

**Push CI vs pull CD — the critical split.** Traditional CI/CD *pushes* from a pipeline that holds cluster
credentials. GitOps *pulls*: CI builds/tests/scans the image and writes a new image tag into the **config repo**;
an in-cluster agent (Argo CD / Flux) notices the commit and applies it. The CI pipeline never needs cluster
credentials, which shrinks the attack surface. **The number one GitOps mistake is making CI `kubectl apply`
directly — that is push CD wearing a GitOps costume.** Let the agent apply.

```
   CI (push)                         CD (pull)
   ┌──────────────┐  image:sha  ┌──────────────┐  commit  ┌─────────────┐   reconcile   ┌─────────┐
   │ build/test/  │ ──────────► │ push to      │ ───────► │ config repo │ ◄──────────── │ Argo CD │
   │ scan image   │             │ registry     │          │  (Git)      │   (in-cluster │ / Flux  │
   └──────────────┘             └──────────────┘          └─────────────┘    agent pulls)└─────────┘
```

**Drift detection & self-heal.** The agent compares Git (desired) to the cluster (live). A manual `kubectl edit`
creates drift; with self-heal enabled the agent reverts it back to Git. This is why "the cluster is the source of
truth" is false under GitOps — Git is.

**Tooling map (what generates manifests vs what applies them):**
- **Helm** — package manager; templates YAML from `values`. Produces manifests.
- **Kustomize** — template-free overlay/patch engine (built into `kubectl -k`). Produces manifests.
- **Argo CD** — pull-based CD agent with a UI; understands Helm, Kustomize, Jsonnet, plain YAML. Applies manifests.
- **Flux** — pull-based CD agent (controllers: source, kustomize, helm, image-automation, notification). Applies manifests.
- **Argo Rollouts / Flagger** — progressive delivery controllers (canary/blue-green) layered on top.

## Workflow / how to approach GitOps & CI/CD tasks

### 1. Establish repo structure first
Separate **application source code** (Dockerfile, app code, CI) from **deployment config** (manifests / Helm
values / Kustomize overlays). Prefer a dedicated **config repo** so the CI pipeline writing an image tag does not
re-trigger itself, and so cluster-write access is isolated. Within the config repo, model environments with
**Kustomize overlays** or **per-environment folders** — *not* long-lived `staging`/`prod` branches (anti-pattern;
see `references/repo-structure-and-environments.md`). Promotion = copying/patching a tested image tag from one
folder to the next via PR.

### 2. Pick the manifest tool
- **Kustomize** when you have one app and need small per-environment deltas (replicas, image tag, env vars). No
  templating language, stays declarative, `kubectl apply -k`.
- **Helm** when you package/redistribute an app, need dependency management, runtime parameterization, or release
  lifecycle (install/upgrade/rollback/history). Charts are versioned, shareable, OCI-pushable.
- **Both together** is common and fine: Helm for third-party components, Kustomize overlays to patch them, or
  Kustomize's `helmCharts` / Argo CD's Helm+Kustomize support to render then patch.
See `references/helm.md` and `references/kustomize.md`.

### 3. Pick & configure the CD agent
- **Argo CD** for a UI, multi-cluster from one control plane, ApplicationSets, fine-grained RBAC/Projects. Default
  sync interval ~3 min. See `references/argocd.md`.
- **Flux** for a CLI/GitOps-native flow, tight Git→controller mapping, built-in image update automation and SOPS.
  See `references/flux.md`.
Both are pull-based and mature. Start with **manual sync** in prod until you trust the pipeline, then enable
`automated: { prune, selfHeal }`.

### 4. Wire CI to feed CD (don't let CI deploy)
CI pipeline stages: **checkout → build → test → scan (SAST/image) → build+push image (tag = commit SHA, never
`latest`) → update image tag in config repo (commit/PR)**. The GitOps agent takes it from there. Tag images with
the immutable commit SHA for traceability and reproducible rollbacks. See `references/cicd-pipeline-patterns.md`
for GitHub Actions / Azure Pipelines / GitLab CI examples and pipeline-as-YAML template reuse.

### 5. Roll out safely
For risky changes use **progressive delivery**: Argo Rollouts or Flagger shift a small % of traffic to the new
version, run **analysis** against metrics (success rate, latency from Prometheus), and auto-promote or auto-rollback.
Choose the strategy by need: **rolling** (default, in-place, cheap), **blue-green** (instant cutover + instant
rollback, 2× resources), **canary** (gradual, metric-gated, best blast-radius control). See
`references/progressive-delivery.md`.

### 6. Handle secrets correctly
Never commit plaintext secrets. Pick one:
- **Sealed Secrets** — `kubeseal` encrypts to a `SealedSecret` that only the in-cluster controller can decrypt;
  safe to commit. Best for on-prem / no external vault.
- **External Secrets Operator** — `ExternalSecret` pulls from AWS/Azure/GCP secret managers or Vault at runtime;
  nothing sensitive in Git. Best when you already have a vault.
- **SOPS** — encrypt values in-file (age/KMS/PGP); Flux decrypts natively, Argo CD via a plugin.
See `references/secrets-in-gitops.md`.

### 7. Verify
- Argo CD: `argocd app get <app>` / UI — check **Sync status** (Synced/OutOfSync) and **Health** (Healthy/Degraded/Progressing).
- Flux: `flux get kustomizations` / `flux get helmreleases` — check `Ready` and the applied revision.
- After enabling auto-sync, deliberately drift a resource (`kubectl scale`) and confirm the agent reverts it.

## Common pitfalls & anti-patterns

- **CI does `kubectl apply` to prod.** That's push CD; you lose drift detection, audit, and you must hand cluster
  creds to CI. Write the change to Git and let the agent pull.
- **Branch-per-environment** (`dev`/`staging`/`prod` branches). Causes merge conflicts, config drift, cherry-pick
  pain, and fights the Kubernetes ecosystem. Use overlays/folders in one branch; promote by copy/patch via PR.
- **`image: latest` (or no tag).** Mutable, non-reproducible, breaks rollback and drift detection. Pin to an
  immutable digest or commit-SHA tag. Enforce with a Kyverno `disallow-latest-tag` policy.
- **Plaintext secrets in Git.** Use Sealed Secrets / ESO / SOPS. Storing a base64 `Secret` in Git is *not*
  encryption — base64 is encoding.
- **Auto-sync + self-heal on day one in prod.** Start manual; a single bad commit auto-syncs everywhere ("blast
  radius"). Add PR review, validation, and progressive delivery before enabling automation broadly.
- **One Argo CD instance, no AppProjects/RBAC, managing everything.** Single point of failure + over-broad creds.
  Scope teams with `AppProject` (allowed repos/destinations/resources) and RBAC; consider instance-per-group at scale.
- **Umbrella-chart value overrides under the wrong key.** With an umbrella chart you must namespace subchart values
  (`ingress-nginx.controller.resources`), not set them at the top level — a very common "my values aren't applying" bug.
- **Ignoring pipeline failures / no rollback path / over-reliance on slow e2e tests / static one-size pipeline.**
  Classic CI/CD anti-patterns. See `references/cicd-pipeline-patterns.md`.
- **Treating GitOps as data backup.** Git restores *config*, not data. A deleted database is not recoverable from
  Git — keep a separate data backup/DR strategy.

## Reference files

- `references/gitops-principles.md` — principles, push-vs-pull, reconciliation/drift/self-heal, IaC + Terraform/Flux, when GitOps doesn't fit.
- `references/helm.md` — chart anatomy, templating, values/overrides, install/upgrade/rollback/history, dependencies & umbrella charts, OCI registries, Helm vs Kustomize, security.
- `references/kustomize.md` — bases/overlays, patches (strategic-merge & JSON6902), generators, common transformers, `helmCharts`, gotchas.
- `references/argocd.md` — `Application`, `ApplicationSet` + generators, app-of-apps, sync waves/hooks/options, `AppProject` & RBAC, multi-cluster, health/sync, hardening.
- `references/flux.md` — bootstrap, `GitRepository`/`OCIRepository`, `Kustomization`, `HelmRelease`/`HelmRepository`, image update automation, SOPS, Argo CD vs Flux.
- `references/progressive-delivery.md` — Argo Rollouts (canary/blue-green, AnalysisTemplate) and Flagger, metric gates, rollback, strategy selection.
- `references/secrets-in-gitops.md` — Sealed Secrets, External Secrets Operator, SOPS — full manifests, tradeoffs, scaling.
- `references/repo-structure-and-environments.md` — app vs config repo, mono vs poly, overlays vs branches vs per-Git, promotion flows, multitenancy & GitOps-at-scale topologies.
- `references/cicd-pipeline-patterns.md` — stages, push/pull, deployment strategies, design patterns & anti-patterns, GitHub Actions / Azure Pipelines / GitLab CI examples, pipeline-as-YAML.

Open the reference file that matches the task; keep this SKILL.md as the index and mental model.
