# Kustomize

Kustomize customizes Kubernetes manifests **without templates** — you keep plain YAML and layer changes on top via
a `kustomization.yaml`. It's built into kubectl (`kubectl apply -k`, since v1.14) and standalone (`kustomize`).
Read this for the base/overlay model, patch types, generators, and gotchas.

## Mental model: base + overlays

- **Base** — the canonical, environment-agnostic resources (Deployment, Service, …) plus a `kustomization.yaml`
  listing them. Changes rarely; shared by all environments.
- **Overlay** — a per-environment directory that references the base and applies **patches** + transformers
  (namespace, name prefix, labels, replica counts, image tags). One overlay per env (dev/staging/prod, or
  qa-europe/qa-asia, etc.).

```
myapp/
├── base/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── kustomization.yaml
└── overlays/
    ├── dev/
    │   ├── kustomization.yaml
    │   └── patch.yaml
    ├── staging/
    │   ├── kustomization.yaml
    │   └── patch.yaml
    └── prod/
        ├── kustomization.yaml
        └── patch.yaml
```

### Base
```yaml
# base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml
  - service.yaml
```

```yaml
# base/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simple-webapp
spec:
  replicas: 2
  selector:
    matchLabels: { app: simple-webapp }
  template:
    metadata:
      labels: { app: simple-webapp }
    spec:
      containers:
        - name: simple-webapp
          image: ghcr.io/example/simple-webapp:1.0.1   # base/default tag
          ports: [{ containerPort: 8080 }]
```

### Overlay (strategic-merge patch + transformers)
```yaml
# overlays/qa/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: simple-webapp-qa       # set namespace for all resources
namePrefix: qa-                   # prefix all resource names: qa-simple-webapp
commonLabels:                     # add labels to all resources + selectors
  variant: qa
resources:
  - ../../base
patches:
  - path: patch.yaml              # strategic-merge patch (target inferred from the patch's kind+name)
```

```yaml
# overlays/qa/patch.yaml — only the fields you want to change
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simple-webapp            # must match base resource name (before namePrefix)
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: simple-webapp
          image: ghcr.io/example/simple-webapp:1.1.5   # override image tag for QA
```

```bash
kustomize build overlays/qa                      # render to stdout for review/diff
kustomize build overlays/qa | kubectl apply -f - # apply
kubectl apply -k overlays/qa                      # equivalent, built into kubectl
diff <(kustomize build overlays/qa) <(kustomize build overlays/staging)  # compare environments
```

## Patch types

**Strategic-merge patch** (above) — partial YAML merged by kind+name. Intuitive for most field changes; arrays
merge by patch-strategy keys (e.g. containers by `name`).

**JSON 6902 patch** — precise op-based edits, ideal for arrays/precise paths:
```yaml
patches:
  - target:
      kind: Deployment
      name: simple-webapp
    patch: |-
      - op: replace
        path: /spec/replicas
        value: 3
      - op: add
        path: /spec/template/spec/containers/0/env/-
        value: { name: TIER, value: gold }
```

## Common transformers (set directly in kustomization.yaml)

```yaml
namespace: prod
namePrefix: prod-
nameSuffix: -v2
commonLabels:        { env: prod }
commonAnnotations:   { team: platform }
images:
  - name: ghcr.io/example/simple-webapp   # override image without a patch
    newTag: 1.2.0
replicas:
  - name: simple-webapp
    count: 5
```

The `images` transformer is the idiomatic way for CI to **promote an image tag** — a pipeline bumps `newTag` in
the overlay's `kustomization.yaml` and commits; the GitOps agent applies it.

## Generators

ConfigMap/Secret generators create resources from literals/files and append a content hash to the name, which
*forces a rolling update* when the data changes (solving the "ConfigMap edit doesn't restart pods" problem):

```yaml
configMapGenerator:
  - name: app-config
    literals:
      - LOG_LEVEL=info
    files:
      - app.properties
secretGenerator:
  - name: app-secret
    envs:
      - secret.env          # do NOT commit plaintext secrets; see secrets-in-gitops.md
generatorOptions:
  disableNameSuffixHash: false   # keep the hash so changes trigger rollouts
```

## Kustomize + Helm

`helmCharts` lets Kustomize render a Helm chart then patch the output — combining Helm packaging with Kustomize
overlays:
```yaml
helmCharts:
  - name: ingress-nginx
    repo: https://kubernetes.github.io/ingress-nginx
    version: 4.8.0
    releaseName: ingress-nginx
    valuesFile: values.yaml
# then add patches: / images: to tweak the rendered manifests
```
(Requires `--enable-helm`.) Both Argo CD and Flux can also do Helm-then-Kustomize natively.

## Gotchas

- **Patch name must match the base resource name** (the name *before* `namePrefix`/`nameSuffix` are applied).
- **`commonLabels` are immutable on existing Deployments** — they're added to selectors, and selectors can't
  change after creation. Use `commonAnnotations` or `labels` (non-selector) for things you'll change later.
- **No templating, no conditionals** by design. If you need logic/loops/parameterization at install time, that's
  a Helm signal.
- **Generator hash names** mean downstream references must use the generated name (Kustomize rewrites references
  to ConfigMaps/Secrets it generates automatically; hand-written refs to the hashed name will break).
- **`resources` vs `bases`** — modern Kustomize uses `resources:` for both files and directories; `bases:` is deprecated.
