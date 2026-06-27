# Flux (Flux CD v2)

Flux is a pull-based GitOps agent built from a set of GitOps Toolkit controllers. It continuously syncs sources
(Git/OCI/Helm repos) to the cluster. Read this for bootstrap, the source/kustomize/helm/image-automation
controllers, SOPS, and the Argo CD vs Flux decision.

## Controllers (the moving parts)

- **source-controller** — fetches and caches artifacts from `GitRepository`, `OCIRepository`, `HelmRepository`,
  `Bucket`; detects new revisions.
- **kustomize-controller** — applies `Kustomization` resources (builds Kustomize/plain YAML and reconciles).
- **helm-controller** — manages `HelmRelease` resources (declarative Helm install/upgrade/rollback).
- **notification-controller** — inbound webhooks (trigger reconcile) and outbound alerts (Slack, etc.).
- **image-reflector-controller** + **image-automation-controller** — scan registries for new tags and write the
  new image reference back to Git automatically.

All install into the `flux-system` namespace.

## Bootstrap

`flux bootstrap` installs the controllers, commits their manifests to your repo, and configures Flux to manage
itself from Git (subsequent upgrades happen via `git push`):

```bash
flux check --pre                          # verify cluster prerequisites
export GITHUB_TOKEN=ghp_xxx               # PAT with repo admin scope
flux bootstrap github \
  --token-auth \
  --owner=my-org \
  --repository=fleet-config \
  --branch=main \
  --path=clusters/prod \                  # per-cluster path in the repo
  --personal
```

This creates `clusters/prod/flux-system/` with the toolkit manifests and a self-referencing `GitRepository` +
`Kustomization` named `flux-system`. Verify: `kubectl get all -n flux-system`, `flux get kustomizations`.

Bootstrap supports GitHub, GitLab, Bitbucket, generic Git (`flux bootstrap git`) — see the docs for providers.

## Sources

```yaml
# GitRepository — track a branch/tag of a Git repo
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: app-config
  namespace: flux-system
spec:
  interval: 1m
  url: https://github.com/my-org/app-config
  ref:
    branch: main
  secretRef:
    name: app-config-auth        # for private repos (PAT or SSH); base64 username/password Secret
```

```yaml
# OCIRepository — track an OCI artifact (manifests/charts pushed to a registry); increasingly preferred
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: OCIRepository
metadata:
  name: app-oci
  namespace: flux-system
spec:
  interval: 5m
  url: oci://registry.example.com/my-org/app-config
  ref:
    tag: 1.4.0
```

## Kustomization — apply manifests from a source

The Flux `Kustomization` (note: *not* the Kustomize `kustomization.yaml`; this is a Flux CR) selects a source +
path and reconciles it:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: app
  namespace: flux-system
spec:
  interval: 10m
  sourceRef:
    kind: GitRepository
    name: app-config
  path: ./overlays/prod          # runs kustomize build on this path
  prune: true                    # delete resources removed from Git (self-heal of deletions)
  wait: true                     # wait for applied resources to become ready
  timeout: 5m
  targetNamespace: my-app
```

```bash
flux get kustomizations          # READY + applied revision (e.g. main@sha1:...)
flux reconcile kustomization app --with-source   # force an immediate sync
```

## HelmRelease — declarative Helm

```yaml
# Define the chart source once
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: ingress-nginx
  namespace: flux-system
spec:
  interval: 1h
  url: https://kubernetes.github.io/ingress-nginx
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: ingress-nginx
  namespace: flux-system
spec:
  interval: 10m
  chart:
    spec:
      chart: ingress-nginx
      version: "4.x"
      sourceRef:
        kind: HelmRepository
        name: ingress-nginx
  values:
    controller:
      replicaCount: 2
  install:
    createNamespace: true
  upgrade:
    remediation:
      retries: 3                 # auto-rollback behavior on failed upgrade
```
`HelmRepository` can also be an OCI registry (`type: oci`, `url: oci://…`). Check with `flux get helmreleases`.

## Image update automation

Flux can watch a registry and bump the image tag in Git for you (the GitOps-native alternative to a CI step
committing the tag):

```yaml
apiVersion: image.toolkit.fluxcd.io/v1beta2
kind: ImageRepository
metadata: { name: app, namespace: flux-system }
spec:
  image: registry.example.com/my-org/app
  interval: 5m
---
apiVersion: image.toolkit.fluxcd.io/v1beta2
kind: ImagePolicy
metadata: { name: app, namespace: flux-system }
spec:
  imageRepositoryRef: { name: app }
  policy:
    semver: { range: ">=1.0.0" }   # or filter by SHA/timestamp pattern
---
apiVersion: image.toolkit.fluxcd.io/v1beta1
kind: ImageUpdateAutomation
metadata: { name: app, namespace: flux-system }
spec:
  interval: 5m
  sourceRef: { kind: GitRepository, name: flux-system }
  git:
    commit:
      author: { name: fluxbot, email: flux@example.com }
    push: { branch: main }
  update:
    path: ./overlays/prod
    strategy: Setters
```
Mark the field to update in the manifest with a setter comment:
`image: registry.example.com/my-org/app:1.0.0 # {"$imagepolicy": "flux-system:app"}`.

## Secrets with SOPS (Flux-native)

Flux decrypts SOPS-encrypted manifests at apply time — point a `Kustomization` at a decryption key:
```yaml
spec:
  decryption:
    provider: sops
    secretRef:
      name: sops-age            # Secret holding the age private key (age.agekey)
```
Encrypt with `sops --encrypt --age <pubkey> secret.yaml > secret.enc.yaml` and commit the encrypted file. See
`secrets-in-gitops.md`.

## Multi-environment with Flux

Mirror the cluster layout in the repo (`clusters/dev`, `clusters/staging`, `clusters/prod`), bootstrap each
cluster at its `--path`, and use per-environment Kustomize overlays or separate `Kustomization` CRs. Promotion =
commit the tested image/manifest into the next environment's path. Flux's `Kustomization` `dependsOn` orders
reconciliation (e.g. infra before apps).

## Argo CD vs Flux

| | Argo CD | Flux |
|---|---|---|
| Interface | Web UI + CLI | CLI-first (optional web UI) |
| Install | CRDs + dedicated namespace | `flux bootstrap` (self-managing from Git) |
| Sync | Pull, default ~3 min | Pull, configurable down to seconds |
| Multi-cluster | First-class (Cluster generator, ApplicationSets, central control plane) | Per-instance / per-cluster; no built-in central multi-cluster console |
| Secrets | Integrates with external secret tools | Built-in SOPS decryption |
| Config tools | Helm, Kustomize, Jsonnet, plain YAML | Helm (helm-controller), Kustomize, plain YAML |
| Image automation | Argo Image Updater (separate) | Built-in image-automation controllers |
| Best fit | UI-driven, multi-cluster, fine-grained RBAC/Projects | GitOps-native CLI flow, image automation, smaller autonomous teams |

Flux demands stronger up-front Helm/Kustomize/Kubernetes fluency; Argo CD can run plain manifests with a gentler
on-ramp and a visual diff. Both are CNCF-graduated and pull-based — choose on workflow, not capability gaps.
Cloud defaults differ (e.g. AKS ships a Flux extension; OpenShift ships Argo CD as GitOps).
