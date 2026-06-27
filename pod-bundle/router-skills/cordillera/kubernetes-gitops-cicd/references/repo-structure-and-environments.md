# Repo structure, environments, promotion & GitOps at scale

How you lay out Git repos largely determines how manageable GitOps becomes. Read this for app-vs-config repo
split, mono vs poly, environment modeling (overlays/folders vs branches vs per-Git), promotion flows, and
multitenancy/instance topologies at scale.

## App repo vs config repo

Separate **application source** (code, Dockerfile, CI) from **deployment config** (manifests / Helm values /
Kustomize overlays):

- CI builds/tests/scans the image in the **app repo**, then writes the new image tag into the **config repo**.
- The GitOps agent watches only the **config repo**.

Why split: the agent reconciling the config repo won't re-trigger the image build (no infinite loop); cluster-
write concerns are isolated to the config repo; access control is cleaner (devs write app repo, platform gates
config repo). A monorepo *can* work, but exclude the config/deployment path from the CI trigger to avoid redundant
runs (e.g. GitHub Actions `paths-ignore: [ 'deployment/**' ]`).

## Monorepo vs polyrepo (config)

| | Monorepo (one repo, many apps) | Polyrepo (one repo per app/team) |
|---|---|---|
| Pros | Simpler dependency/version tracking; consistent workflow; one place to see everything | Isolates changes; smaller blast radius; per-team access control |
| Cons | Can get unwieldy at scale; needs strong CI/build tooling | Cross-repo coordination harder; dependency management spreads out |
| Fits | Smaller orgs, tightly-coupled stacks | Large orgs, autonomous teams, strict separation |

Argo CD ApplicationSets (Git/SCM generators) and Flux handle both; the choice is organizational, not technical.

## Environment modeling — three approaches

### 1. Folders / overlays in one branch (recommended)
Each environment is a directory in one branch, typically Kustomize base + per-env overlays:
```
config-repo/
├── base/                 # shared, changes rarely
│   ├── deployment.yaml
│   └── kustomization.yaml
└── overlays/
    ├── qa/      { kustomization.yaml, patch.yaml }
    ├── staging/ { kustomization.yaml, patch.yaml }
    └── prod/    { kustomization.yaml, patch.yaml }
```
- Clear, side-by-side view of all stages; no branch-switching to compare.
- Promotion = copy/patch a tested value (image tag, replicas) from one overlay to the next, then commit.
- Aligns with Kubernetes' declarative nature; scales well. **This is the GitOps-idiomatic default.**

### 2. Branch-per-environment (anti-pattern for GitOps/K8s)
`dev`/`staging`/`prod` long-lived branches. Familiar to many devs and "promote = merge," but:
- Merges cause conflicts and unintended changes; cherry-picking dependent changes is painful.
- Branches drift into environment-specific code (config drift).
- Managing many env branches becomes unwieldy.
Acceptable for legacy apps; avoid for new GitOps. (Feature branches + PRs for *testing changes* are still fine —
just not as the environment model.)

### 3. Environment-per-Git-repo
Separate repos for prod vs non-prod, mainly for **security isolation** (e.g. keep junior devs out of prod
config). Stronger boundary, more repos to manage. Most folder-approach problems can instead be solved with CODEOWNERS, branch
protection, validation, and PR review — reserve separate repos for genuinely strict isolation needs.

## Promotion flows (folder approach)

Promotion is a Git operation, not a special tool:
```bash
# Promote a tested image tag from QA to staging
cp overlays/qa/image-tag.yaml overlays/staging/image-tag.yaml
git commit -am "promote api 1.4.0 qa->staging" && git push   # agent applies to staging
```
Promote business-wide settings by "graduating" them down the stages and finally into `base`: add the new config
to QA → test → copy to staging → test → copy to prod → once all stages share it, move it into `base` and remove
the per-stage copies. Use `diff overlays/qa overlays/staging` to see exactly what differs. **Promotion should go
through PR + review, never a direct commit to the prod path.** Automate promotion with any workflow tool; Argo
CD's PR generator manifests Git→cluster (not cluster→Git), so writing the tag back to Git is a separate CI/job step.

## GitOps at scale — service catalog with labels

For many clusters, label each cluster by capability and let an Argo CD **Cluster generator** ApplicationSet
deploy the matching stack. Example label scheme:

| Label | Deploys |
|---|---|
| `core-basic: enabled` | argocd, external-dns, ingress-nginx |
| `security-basic: enabled` | cert-manager, kyverno, sealed-secrets, external-secrets, falco, rbac |
| `monitoring-basic: enabled` | grafana, metrics, alerting, exporters |
| `storage-basic: enabled` | minio-operator, nfs-provisioner |

An ApplicationSet selecting `matchLabels: { env: prod, core-basic: enabled }` deploys core add-ons only to
matching clusters; cert-manager's ApplicationSet additionally requires `security-basic: enabled`. This is a
lightweight internal developer platform (a "Kubernetes Service Catalog"): one central Argo CD manages many
clusters, and a maintenance job/Action bumps Helm chart versions fleet-wide via PR when CVEs drop. See
`argocd.md` for ApplicationSet syntax and `helm.md` for umbrella charts used per-service.

## Multi-instance / multitenancy topologies

| Topology | Summary | Use when | Watch out for |
|---|---|---|---|
| **Centralized** (1 instance, N clusters) | One Argo CD + UI manages all clusters | Small teams; dev/staging/prod | Single point of failure; central admin creds = big blast radius; network cost across regions |
| **Instance per cluster** | Co-located agent per cluster | Edge/air-gapped; strict isolation | Maintenance overhead grows with cluster count |
| **Instance per logical group** | One per team/region/env | Departmental/geographic separation | Multiple instances to maintain; group control cluster is a new SPOF |
| **Cockpit & fleet** | Central "cockpit" provides platform context; per-cluster "fleet" instances for team autonomy | Large orgs needing central governance + local autonomy | Dual system to operate; hosted/agent variants reduce central creds |

**Native Argo CD multitenancy** uses `AppProject` (allowed repos/destinations/resource kinds) + RBAC + namespace
quotas/NetworkPolicies. Its limits: teams can't freely use the UI/CLI without Dex+RBAC; CRDs of `Application`/
`AppProject` must live in the argocd namespace; and "restrict creation" ≠ "restrict manipulation" (pair with
Kyverno + admission webhooks). **vCluster** addresses these by giving each tenant a virtual cluster (own API,
own CRDs, own Argo CD) on a shared host cluster — more isolation, more resource overhead. Choose by your security/
compliance requirements and team GitOps maturity; the *approach* matters more than the specific tool.

## Quick checklist for a new GitOps repo

1. Config repo separate from app repo (or path-excluded monorepo).
2. Kustomize base + per-env overlays (no env branches).
3. Image tag = commit SHA, set via the overlay `images:` transformer or a Helm value.
4. Promotion = copy/patch via PR with review + branch protection on prod paths.
5. Secrets via Sealed Secrets / ESO / SOPS — never plaintext.
6. Agent: manual sync first, then `automated: { prune, selfHeal }` once trusted.
7. At scale: cluster labels + ApplicationSet Cluster generator; AppProject per team.
