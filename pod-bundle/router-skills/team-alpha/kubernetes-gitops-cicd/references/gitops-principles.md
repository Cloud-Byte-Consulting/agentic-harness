# GitOps principles, reconciliation, and IaC

The mental model that everything else in this skill hangs from. Read this when reasoning about *why* a GitOps
design is right, debugging drift, or deciding whether GitOps even fits a problem.

## The four principles (GitOps Working Group)

1. **Declarative.** The entire system — apps, config, infra, even policy — is expressed declaratively (desired
   *what*, not imperative *how*). Kubernetes objects are inherently declarative, which is why GitOps and
   Kubernetes fit so well.
2. **Versioned and immutable.** Desired state lives in Git: versioned, auditable, and immutable. You change the
   system by committing a new revision, never by editing live resources. Rollback = `git revert` to a known-good commit.
3. **Pulled automatically.** Software agents automatically pull the declared state from the source. Nobody pushes
   `kubectl apply` from a laptop or pipeline.
4. **Continuously reconciled.** Agents continuously observe the live cluster and reconcile it toward the declared
   state, correcting any divergence.

## Reconciliation loop, drift, and self-heal

Kubernetes already runs reconciliation loops (controllers drive actual state → desired state). GitOps extends that
loop *outward* to Git:

```
Git (desired) ──pull──► agent compares ──► diff? ──► apply create/update/delete ──► cluster (live)
       ▲                                                                                  │
       └──────────────────────── observe live state ─────────────────────────────────────┘
```

- **Drift** = live state diverged from Git (e.g. someone ran `kubectl scale deploy/api --replicas=5`, or an
  admission webhook mutated something). The agent flags it as `OutOfSync` (Argo CD) or a non-matching revision (Flux).
- **Self-heal** = the agent automatically re-applies Git to erase the drift. In Argo CD this is
  `syncPolicy.automated.selfHeal: true`; the manual scale-up is reverted within a sync cycle.
- **Prune** = the agent deletes live resources that were removed from Git (`syncPolicy.automated.prune: true`).
  Without prune, deleting a manifest leaves an orphan in the cluster.

**Operational consequence:** under GitOps, *Git is the source of truth, not the cluster.* Hotfixing by editing
live resources is an anti-pattern — it will be reverted (with self-heal) or silently diverge (without). Fix it in Git.

## Push CI vs pull CD

| | Push model (traditional) | Pull model (GitOps) |
|---|---|---|
| Who applies | A CI pipeline runs `kubectl/helm apply` | An in-cluster agent pulls from Git |
| Cluster creds | Pipeline holds kubeconfig/admin creds | Agent uses its own service account; pipeline needs none |
| Trigger | Code push triggers deploy immediately | Commit to config repo; agent reconciles on its interval/webhook |
| Drift handling | None — live state can silently diverge | Detected and optionally auto-corrected |
| Audit | Scattered across pipeline logs | Every change is a Git commit (who/what/when) |
| Attack surface | Larger (creds in CI, network path in) | Smaller (outbound pull, no inbound cluster access) |

GitOps relies on a CI step: a change must pass CI (build/test/scan) and land in Git *before* the agent acts, so
only verified changes deploy. The handoff point: **CI ends when the artifact (image tag) is committed to the
config repo; CD begins when the agent picks it up.**

## GitOps + IaC (Terraform / OpenTofu + Flux)

GitOps extends IaC's "infrastructure as version-controlled code" to continuous reconciliation. Terraform/OpenTofu
declare cloud infra (resource groups, VNets, AKS/EKS) in HCL; Flux can reconcile that too via the **Tofu
Controller** (formerly Weave TF-Controller), which applies Terraform/OpenTofu in a GitOps manner.

```yaml
apiVersion: infra.contrib.fluxcd.io/v1alpha2
kind: Terraform
metadata:
  name: gitops-terraform-automation
  namespace: flux-system
spec:
  interval: 1m
  approvePlan: auto            # auto-apply the plan; use "" to require manual approval of each plan
  destroyResourcesOnDeletion: true
  path: ./iac/azure/vnet       # where the .tf files live in the repo
  sourceRef:
    kind: GitRepository
    name: flux-system
  runnerPodTemplate:
    spec:
      env:
        - name: ARM_SUBSCRIPTION_ID
          valueFrom:
            secretKeyRef: { name: azure-creds, key: ARM_SUBSCRIPTION_ID }
        # ARM_CLIENT_ID / ARM_CLIENT_SECRET / ARM_TENANT_ID likewise from the secret
```

Tofu Controller GitOps models: **GitOps Automation** (full lifecycle, plan+apply), **Hybrid** (adopt GitOps for
only some resources of an existing cluster), **State Enforcement** (force live state to match TFSTATE),
**Drift Detection** (detect-only). Status flows through `kubectl get terraforms -n flux-system -w`:
`Reconciliation in progress → Terraform Planning → Plan generated → Applying → Applied successfully` then `Ready: True`.

**Note (licensing):** Terraform moved to BSL; **OpenTofu** is the MPL-licensed open-source fork. Both use HCL and
work with the Tofu Controller. Prefer cross-linking to `kubernetes-cluster-operations` for cluster *bootstrap*;
use this only for the GitOps reconciliation angle.

## Disaster recovery scope — important caveat

GitOps DR restores **configuration**, not **data**. Reapplying Git rebuilds Deployments, Services, config — but a
deleted database's contents are gone. Always pair GitOps with a separate data backup/restore strategy. Git is for
declarative config recovery and reconstructing a clean cluster from the source of truth.

## Benefits (why teams adopt it)

Faster, more frequent deploys; consistency/reproducibility across environments; straightforward rollback (revert a
commit); built-in audit/compliance (every change is a reviewed commit); reduced MTTR; smaller attack surface;
self-service for developers (commit + merge triggers everything); easier multi-cluster management (agents per cluster).

## When GitOps is *not* the right fit (or needs care)

- **Data-plane recovery / stateful data** — see DR caveat above; needs separate tooling.
- **Highly imperative, one-off operations** — GitOps shines for steady-state desired config, not ad-hoc imperative scripts.
- **Secrets** — require extra tooling (Sealed Secrets / ESO / SOPS); plaintext in Git is never acceptable.
- **Steep onboarding** — teams must learn Git workflows, Kubernetes, and the agent; a cultural shift toward
  "change = PR," not "SSH in and fix it."
- **Repo structure debt** — a sloppy repo layout becomes an operational bottleneck at scale; design it up front
  (see `repo-structure-and-environments.md`).
