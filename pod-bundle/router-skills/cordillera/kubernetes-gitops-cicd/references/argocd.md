# Argo CD

Argo CD is a declarative, pull-based GitOps CD agent for Kubernetes with a web UI. It continuously compares the
desired state in Git against the live cluster and reconciles. Read this for `Application`, `ApplicationSet` +
generators, app-of-apps, sync waves/hooks/options, `AppProject`/RBAC, multi-cluster, and hardening.

## Architecture (what each piece does)

- **API server** — serves the API/UI/CLI; handles auth and app state.
- **Repo server** — fetches Git contents and renders manifests (understands Helm, Kustomize, Jsonnet, plain YAML).
- **Application controller** — the reconciliation engine: compares live vs desired and applies create/update/delete.
- **Dex** (optional) — OIDC/SSO bridge to external identity providers.

Sync is **pull-based**, default interval ~3 minutes (webhooks make it near-instant). All custom resources
(`Application`, `ApplicationSet`, `AppProject`) live in the `argocd` namespace by default.

## Application — the unit of deployment

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io   # cascade-delete managed resources when the App is deleted
spec:
  project: default
  source:
    repoURL: https://github.com/example/config-repo.git
    targetRevision: main          # branch, tag, or commit
    path: overlays/prod           # path within the repo (Kustomize/Helm/plain YAML auto-detected)
  destination:
    server: https://kubernetes.default.svc   # in-cluster; or a registered remote cluster URL/name
    namespace: my-app
  syncPolicy:
    automated:
      prune: true                 # delete resources removed from Git
      selfHeal: true              # revert manual drift back to Git
    syncOptions:
      - CreateNamespace=true       # create the destination namespace if missing
      - ApplyOutOfSyncOnly=true
```

**Sync status** = `Synced` / `OutOfSync` (live vs Git). **Health** = `Healthy` / `Progressing` / `Degraded` /
`Missing`. A freshly created App is `OutOfSync` until you sync (UI **Sync** button or `argocd app sync my-app`).

Source variants:
- **Helm:** add a `helm:` block under `source` with `releaseName` and `valueFiles`/`values`.
- **Kustomize:** point `path` at an overlay dir; Argo runs `kustomize build`.
- **Multi-source** (`sources:` array) lets one App combine e.g. a chart from one repo with values from another
  (via `ref:` and `$refName`).

```bash
kubectl apply -f my-app.yaml          # register the Application
argocd app get my-app                  # status, sync, health, diff
argocd app sync my-app                 # manual sync
argocd app rollback my-app <history-id>
```

## ApplicationSet — templated Applications at scale

An `ApplicationSet` (handled by the ApplicationSet controller, installed alongside Argo CD) generates *many*
Applications from one definition using **generators**. Unlike a single Application (one repo → one cluster/ns),
an ApplicationSet spans many clusters/namespaces/repos.

Generators: **List** (static params), **Cluster** (iterate registered clusters by label), **Git**
(files/directories in a repo), **Pull Request** (per open PR — great for ephemeral preview envs), **Matrix**
(combine two generators), **SCM Provider**, etc.

**Cluster generator** — deploy an app to every cluster matching a label (label clusters in
Settings → Clusters or via `argocd cluster add … --label env=prod`):

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: ingress-nginx
  namespace: argocd
spec:
  generators:
    - clusters:
        selector:
          matchLabels:
            env: prod
            core-basic: enabled       # multiple labels = AND
        values:
          branch: main
  template:
    metadata:
      name: "{{name}}-ingress-nginx"   # {{name}} = cluster name -> in-cluster-ingress-nginx
    spec:
      project: default
      sources:
        - repoURL: git@github.com:example/config-repo.git
          targetRevision: "{{values.branch}}"
          path: networking/ingress-nginx
          helm:
            releaseName: ingress-nginx
            valueFiles:
              - values.yaml
              # per-cluster overrides via a second source (ref) and $valuesRepo/...
      destination:
        name: "{{name}}"
        namespace: ingress-nginx
      syncPolicy:
        syncOptions: [ CreateNamespace=true ]
```

```bash
kubectl get applicationsets -n argocd
kubectl get applications -n argocd      # the generated Apps appear here (the AppSet itself isn't in the UI)
```

Use **labels on clusters + Cluster generator** to selectively deploy stacks (e.g. only clusters labelled
`security-basic: enabled` get cert-manager). This is the foundation of a scalable service catalog — see
`repo-structure-and-environments.md`.

## App-of-apps

A root Application whose `source.path` points at a **folder of child Application manifests** (often with
`directory.recurse: true`), so one root App declares and syncs many child Apps. Good for **cluster bootstrap**
(deploy a standard stack to every new cluster) and managing apps purely via Git (no UI/CLI). For most scaling
needs **ApplicationSets are preferred** (more flexible, multi-cluster, generator-driven); app-of-apps suits a
related bundle in one repo.

```yaml
spec:
  source:
    repoURL: git@github.com:example/config-repo.git
    path: apps/                # folder containing many Application manifests
    targetRevision: main
    directory:
      recurse: true
  syncPolicy:
    automated: { prune: true, selfHeal: true }
```

## Sync waves, hooks, and options

- **Sync waves** order resources within a sync: annotation `argocd.argoproj.io/sync-wave: "-1"` (lower runs
  first). Use to apply CRDs/namespaces before the workloads that need them.
- **Resource hooks**: `argocd.argoproj.io/hook: PreSync|Sync|PostSync|SyncFail` runs a resource (often a Job) at a
  phase — e.g. a `PreSync` DB migration Job before the new app rolls out. `hook-delete-policy` controls cleanup.
- **Sync options** (per-App or per-resource annotation): `CreateNamespace=true`, `Prune=false`,
  `ServerSideApply=true`, `SkipDryRunOnMissingResource=true`, etc.

```yaml
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "1"
    argocd.argoproj.io/hook: PreSync
    argocd.argoproj.io/hook-delete-policy: HookSucceeded
```

## AppProject & RBAC — multitenancy guardrails

An `AppProject` constrains *what a set of Applications may do*: which source repos, which destination
clusters/namespaces, and which resource kinds can be created. This is the primary multitenancy boundary.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: devteam-a
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  description: DevTeam-A deployment guardrails
  sourceRepos:
    - "https://github.com/example/devteam-a-*"
  destinations:
    - namespace: devteam-a
      server: https://kubernetes.default.svc
  clusterResourceBlacklist:        # forbid cluster-scoped resource creation
    - group: ""
      kind: Namespace
  namespaceResourceBlacklist:      # forbid creating these even in their own namespace
    - { group: argoproj.io, kind: AppProject }
    - { group: argoproj.io, kind: Application }
    - { group: "", kind: ResourceQuota }
    - { group: networking.k8s.io, kind: NetworkPolicy }
```

**Destination wildcard trap:** allowing `namespace: "*"` to deploy "anywhere except kube-system" still lets a team
into *other teams'* namespaces, because `!kube-system` only excludes one. Be explicit: list the exact allowed
namespaces. AppProject restricts *creating* resources, not *manipulating* existing ones — pair with cluster RBAC
and admission policies (Kyverno) for full isolation.

RBAC lives in the `argocd-rbac-cm` ConfigMap (`policy.csv`), mapping SSO groups to roles:
```yaml
data:
  policy.csv: |
    p, role:org-admin, applications, *, */*, allow
    p, role:org-admin, clusters, get, *, allow
    g, "GROUP_ID_FROM_IDP", role:org-admin
  policy.default: role:readonly
```

## Multi-cluster topologies (pick per scale)

- **Centralized / one instance, many clusters** — simplest, one UI; but single point of failure and centralized
  admin creds (large blast radius). Good for small teams, dev/staging/prod.
- **Instance per cluster** — best isolation/reliability, ideal for edge/air-gapped; high maintenance overhead at scale.
- **Instance per logical group** (team/region/env) — middle ground; load distribution, group-scoped creds.
- **Cockpit & fleet** — central "cockpit" instance provides platform context to all clusters, plus a dedicated
  "fleet" instance per cluster for team autonomy. Hosted variants (e.g. Akuity) invert it with an outbound agent
  per fleet cluster, removing central cluster creds.

See `repo-structure-and-environments.md` for the full tradeoff tables and when to choose each.

## Hardening (from the CNCF/ControlPlane Argo CD threat model)

- **Disable the local admin account** after switching to SSO; admin creds never expire and are username/password only.
  Set `admin.enabled: "false"` in `argocd-cm`, map an IdP group to `role:org-admin` in `argocd-rbac-cm`, use Dex+OIDC.
- **Rotate/delete the initial admin secret.** `argocd-initial-admin-secret` only stores the bootstrap password in
  clear text and can be deleted after you change the password (`argocd account update-password`).
- **Rotate tenant cluster credentials.** The `argocd-manager` token on managed clusters can be set to never
  expire — prefer workload/managed identities (AKS/EKS) or an external KMS; deleting the cluster Secret triggers a
  new token.
- **Restrict access** to the cockpit/control-plane cluster by firewall/IP allow-list + MFA; manifest generation
  (Helm/Kustomize/Jsonnet) can execute arbitrary code, so limit who can change config-management plugins.
- Use the **HA install** (multiple repo-server/controller replicas) for scale; one instance handles roughly up to
  ~1,500 apps / ~50 clusters / ~200 users before tuning.
