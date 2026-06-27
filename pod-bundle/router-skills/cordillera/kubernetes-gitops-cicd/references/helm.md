# Helm

Helm is the Kubernetes package manager: it packages a set of related resources into a **chart**, parameterizes
them with Go templates + `values`, and tracks installed instances as **releases** you can upgrade and roll back.
Read this for chart authoring, value overrides, release lifecycle, dependencies/umbrella charts, OCI, and the
Helm-vs-Kustomize decision.

## Three concepts

- **Chart** — the package: a directory of templated YAML manifests + metadata. What you install.
- **Repository** — where charts are stored/shared (HTTP repo or, preferred today, an **OCI registry**). Browse
  public charts on Artifact Hub.
- **Release** — a named, installed instance of a chart in a cluster. You can install one chart many times under
  different release names; each is independently upgradeable/uninstallable.

Helm v3 is fully client-side — **no Tiller** (the v2 in-cluster server component that caused RBAC/privilege
problems). Any guide mentioning Tiller is for the obsolete v2.

## Chart anatomy

```
mychart/
├── Chart.yaml          # metadata: name, version (chart ver), appVersion, dependencies
├── values.yaml         # default config values (the template parameters)
├── values.schema.json  # optional JSON schema validating values.yaml
├── charts/             # optional vendored/dependent subcharts
├── crds/               # optional CRDs (installed before templates, not templated, not upgraded)
└── templates/          # the Go-templated manifests
    ├── deployment.yaml
    ├── service.yaml
    ├── _helpers.tpl     # reusable named templates
    └── NOTES.txt        # post-install usage notes printed to the user
```

`Chart.yaml` essentials:
```yaml
apiVersion: v2          # v2 = Helm 3 chart
name: myapp
version: 0.2.0          # the CHART version — bump on every chart change
appVersion: "2.0.0"     # the app version shipped (informational; quote it)
description: My app
```

Scaffold a new chart with `helm create myapp` (generates a working nginx-style chart you edit down).

## Templating & values

Templates use Go templates. Reference values with `.Values`, chart metadata with `.Chart`, release info with `.Release`:

```yaml
# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-{{ .Chart.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  template:
    spec:
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
```

```yaml
# values.yaml — defaults, overridable at install/upgrade
replicaCount: 1
image:
  repository: nginxdemos/hello
  tag: ""          # falls back to .Chart.AppVersion via the default above
service:
  type: ClusterIP
  port: 80
```

Override values three ways (later wins): chart default → `--values custom.yaml` (one or more files) →
`--set key=value` on the CLI.

```bash
helm template ./myapp                          # render manifests locally, apply nothing (great for review/diff)
helm template --set service.port=8080 ./myapp  # render with an override
helm install myapp ./myapp -n myns --values prod-values.yaml
helm install myapp ./myapp --dry-run --debug   # render + validate against the cluster without applying
```

Prefer a **values file per environment** committed to the config repo over piles of `--set` flags — values files
are reviewable and GitOps-friendly.

## Release lifecycle

```bash
helm install myapp ./myapp -n myns              # create release (REVISION 1)
helm list -n myns                               # list releases: NAME / REVISION / STATUS / CHART / APP VERSION
helm upgrade myapp ./myapp -n myns              # apply changes (REVISION 2); bump Chart.yaml version first
helm history myapp -n myns                      # show revisions: superseded / deployed, with chart+app versions
helm rollback myapp 1 -n myns                   # roll back to a previous revision
helm uninstall myapp -n myns                    # remove the release (PV/PVCs are NOT auto-removed)
```

`helm history` is your audit/rollback trail:
```
1   ... superseded   myapp-0.1.0   1.16.0   Install complete
2   ... deployed     myapp-0.2.0   2.0.0    Upgrade complete
```

Helm also solves the **ConfigMap-doesn't-trigger-rollout** problem: a common pattern is a checksum annotation
(`checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}`) on the pod template
so a config change forces a rolling update.

## Dependencies & umbrella charts

A chart can depend on subcharts (declared in `Chart.yaml`), enabled conditionally:

```yaml
# Chart.yaml
dependencies:
  - name: mariadb
    version: 18.x.x
    repository: oci://registry-1.docker.io/bitnamicharts
    condition: mariadb.enabled        # toggled by values
  - name: memcached
    version: 7.x.x
    repository: oci://registry-1.docker.io/bitnamicharts
    condition: memcached.enabled
```

Run `helm dependency update` to vendor them into `charts/`. The Bitnami WordPress chart, for example, pulls in
MariaDB when `mariadb.enabled: true`.

**Umbrella chart** = a parent chart whose only job is to bundle subcharts (e.g. `kube-prometheus-stack` bundling
Grafana). **Critical gotcha:** to override a subchart's values from the umbrella, you must *namespace under the
subchart name*, not set them at the top level:

```yaml
# umbrella values.yaml — overriding the ingress-nginx subchart
ingress-nginx:                  # subchart name as the key
  controller:
    resources:
      limits: { cpu: 200m, memory: 256Mi }
```

Setting `controller.resources` at the top level silently does nothing — this is the #1 "my values aren't being
applied" bug with umbrella charts. When the umbrella and subchart share a name (e.g. an `ingress-nginx` umbrella
wrapping the `ingress-nginx` subchart), be especially careful: use `global` for umbrella-wide values and the
subchart-name key for subchart overrides.

## OCI registries

Modern distribution is via OCI (the old `helm repo add stable …` HTTP repo is deprecated):

```bash
helm package ./myapp                                          # -> myapp-0.2.0.tgz
helm push myapp-0.2.0.tgz oci://registry.example.com/charts   # push chart to OCI registry
helm install myapp oci://registry.example.com/charts/myapp --version 0.2.0 -n myns
```

`helm search hub <term>` searches Artifact Hub from the CLI; the Add/install instructions appear on each chart's page.

## Helm vs Kustomize — when to use which

| Use **Helm** when… | Use **Kustomize** when… |
|---|---|
| Packaging/redistributing an app for others | Customizing manifests you already have |
| You need dependency management (subcharts) | You want template-free, plain-YAML overlays |
| Runtime parameterization via `values` | Small per-environment deltas (image/replicas/env) |
| You want release versioning + `rollback`/`history` | You want `kubectl apply -k` with no extra tool |
| Sharing across a wide user base | Avoiding a templating language entirely |

They compose well: Helm for third-party components, Kustomize overlays to patch the rendered output per
environment. Both Argo CD and Flux support Helm and Kustomize natively (and combined).

## Helm chart security

- **Verify the source.** Only install from trusted/official repos; a malicious chart runs arbitrary manifests.
- **Audit & scan** charts and their image dependencies (Trivy/Anchore) on a schedule; track known CVEs.
- **Review defaults.** Chart defaults are rarely production-appropriate — review/override before prod.
- **RBAC + namespace isolation.** Restrict who can install/upgrade; install charts into dedicated namespaces to
  limit blast radius.
- **Pin chart and image versions** (no floating `latest`); update deliberately via PR.
