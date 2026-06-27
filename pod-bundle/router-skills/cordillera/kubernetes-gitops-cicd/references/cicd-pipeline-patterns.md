# CI/CD pipeline patterns & anti-patterns

How to build the CI side that feeds GitOps: stages, push/pull, deployment strategies, design patterns,
anti-patterns to avoid, and concrete pipeline-as-YAML for GitHub Actions, Azure Pipelines, and GitLab CI. The
golden rule under GitOps: **CI builds/tests/scans and writes the artifact reference to Git; the in-cluster agent
deploys.** CI should not `kubectl apply` to prod.

## Pipeline stages

A CI/CD pipeline is an ordered series of stages, each with one purpose:

1. **Source / checkout** — triggered by a VCS event (push, PR, tag).
2. **Build** — compile, produce artifacts; for containers, build the image.
3. **Test** — unit → integration → (selective) e2e. See the test pyramid/diamond below.
4. **Scan** — SAST, dependency/image vulnerability scan (Trivy/Anchore), policy checks. Shift-left.
5. **Package / push** — push the image to a registry, tagged with the **immutable commit SHA** (never `latest`).
6. **Deliver to Git** — write the new image tag/manifest into the config repo (commit/PR). *This is the handoff to CD.*
7. **(CD, by the agent)** — Argo CD/Flux reconciles the change into the cluster.
8. **Monitor / feedback** — observe and feed results back.

The CI↔CD boundary: CI ends at "artifact published + referenced in Git"; CD begins when the agent picks it up.

## Push vs pull (recap, with CI focus)

- **Push**: a code commit triggers the pipeline, which deploys directly (holds cluster creds). Immediate but
  larger attack surface, no drift detection.
- **Pull**: the pipeline only updates Git; an in-cluster controller pulls and applies. GitOps-native; CI needs no
  cluster credentials.

## Deployment strategies (chosen in the CD/agent layer, see progressive-delivery.md)

- **Rolling** — replace instances incrementally; cheap, default, in-place.
- **Blue-green (red-black)** — two identical prod envs; switch traffic at once; instant rollback; 2× cost. (In
  blue-green both may briefly receive traffic; in strict red-black only one does.)
- **Canary** — release to a small user subset, monitor, then ramp; best blast-radius control; needs metrics +
  traffic routing.
- **Feature toggles** — enable/disable features at runtime; decouples deploy from release (not a deploy strategy
  per se but key for risk control).

## Design patterns worth applying

- **Pipeline as code** — pipeline defined in version-controlled YAML (Jenkinsfile, `.gitlab-ci.yml`,
  `azure-pipelines.yml`, GH Actions workflow). Reviewable, reproducible, branchable.
- **Immutable infrastructure** — replace, don't mutate; rebuild from IaC. Removes config drift.
- **IaC** — Terraform/OpenTofu/CloudFormation; plan as a stored artifact, PR-gated, drift detection.
- **Automated testing** in the pipeline (unit/integration/e2e), quick feedback.
- **GitOps** — Git as source of truth for deployment; auditability + rollback.
- **Microservices** — independent deploy/scale (raises the bar on integration testing — see test diamond).
- Classic GoF patterns map onto pipelines: **Factory Method** to create per-environment pipelines, **Strategy**
  to select a deployment algorithm (blue-green/canary/rolling) at runtime, **Observer** to notify stakeholders of
  pipeline status, **Adapter** to plug heterogeneous tools into one pipeline.

## Testing: pyramid vs diamond

- **Test pyramid** — many fast unit tests at the base, fewer integration, fewest e2e at the top. Best for
  monoliths. Don't run expensive tests until cheap ones pass.
- **Test diamond** — emphasizes the *integration* middle layer; suits microservices where the cross-service
  communication is the risk. Same test types, different weighting driven by architecture.

## Anti-patterns (avoid these)

- **CI deploys to prod directly** (`kubectl apply` from the pipeline) — push CD masquerading as GitOps. Write to
  Git, let the agent apply.
- **Lack of proper pipeline modeling** — undefined stage flow, steps skipped, deploy-to-prod after every step.
  Map stages explicitly; never let untested code reach prod.
- **`image: latest` / untagged images** — mutable, non-reproducible, breaks rollback. Tag with commit SHA/digest.
- **Ignoring pipeline failures** — silently continuing past failures. Use `continueOnError`/`catchError` *only*
  for genuinely non-critical steps, with visibility; otherwise fail fast.
- **No rollback / error handling** in the deploy path — if prod is bad, you need an automated way back (revert
  commit, blue-green flip, canary auto-abort).
- **Poor/incomplete test automation** — flaky tests, stale suites, over-reliance on slow e2e instead of a balanced
  pyramid.
- **Poor monitoring/observability** — no build/test/deploy metrics; you fly blind on regressions and slow stages.
- **Single point of failure (SPOF)** in CI/CD infra — one runner/server/expert; add redundancy.
- **Bad security integration** — security bolted on at the end. Shift left: threat modeling at design, secure
  coding, automated SAST/DAST, strict access controls, secret scanning.
- **Static one-size-fits-all pipeline** that can't adapt — favor parameterized/templated pipelines.
- **Branch-per-environment** as the env model, **big-bang deployments**, **manual approval gates everywhere**,
  **configuration drift** between envs.

## GitHub Actions — build/scan/push, then write tag to config repo (pull CD)

```yaml
name: ci
on:
  push:
    branches: [ main ]
    paths-ignore: [ 'deployment/**' ]   # don't re-trigger on config-repo path edits
jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - name: Log in to registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Build & push (tag = commit SHA)
        uses: docker/build-push-action@v6
        with:
          push: true
          tags: ghcr.io/${{ github.repository }}:${{ github.sha }}
      - name: Scan image
        uses: aquasecurity/trivy-action@0.24.0
        with:
          image-ref: ghcr.io/${{ github.repository }}:${{ github.sha }}
          severity: HIGH,CRITICAL
          exit-code: '1'
      - name: Bump image tag in config repo   # the GitOps handoff
        run: |
          git clone https://x-access-token:${{ secrets.CONFIG_REPO_TOKEN }}@github.com/org/config-repo.git
          cd config-repo
          kustomize edit set image app=ghcr.io/${{ github.repository }}:${{ github.sha }} \
            --kustomization overlays/prod
          git commit -am "deploy ${{ github.sha }}" && git push    # Argo CD/Flux applies it
```

If using Terraform + GitHub Actions for infra, run `terraform init` with backend config from secrets, save the
plan as an artifact, and apply in a gated job; build the image with QEMU/Buildx tagged by SHA; configure kubectl
via `az aks get-credentials` only if you must push (prefer the pull handoff above). Use **GitHub Secrets** for all
creds; make secret-creating steps idempotent with `kubectl ... --dry-run=client -o yaml | kubectl apply -f -`.

## Azure Pipelines — stages, templates, continueOnError

```yaml
trigger:
  branches: { include: [ main ] }
pool:
  vmImage: ubuntu-latest
stages:
  - stage: Build
    jobs:
      - job: build
        steps:
          - script: echo "build & test"
            displayName: Build
          - script: exit 1
            displayName: Non-critical lint
            continueOnError: true     # pipeline proceeds; failure stays visible
  - stage: Release            # only runs if Build succeeded; rerun a single stage if it fails
    dependsOn: Build
    jobs:
      - job: publish
        steps:
          - script: echo "push image + write tag to config repo"
```

**Template reuse** (DRY across pipelines): extract steps into `dotnet-build-steps.yml` with `parameters:` and
reference it (optionally from another repo) so many apps share one definition:
```yaml
# azure-pipelines.yml
steps:
  - template: dotnet-build-steps.yml
    parameters:
      buildConfiguration: Release
```
**Template expressions** (`${{ if }}`, `${{ each }}`, `eq/contains/coalesce/format`) add runtime logic/conditional
steps. Caveats: YAML pipelines are Azure-specific (not portable), have a ~4 MB expanded-YAML limit, and need
YAML/Azure-Pipelines fluency. Benefit: stored as code, versioned, reviewable; rerun a single failed stage instead
of the whole pipeline.

## GitLab CI — concise build/test/deploy-to-Git

```yaml
stages: [ build, test, scan, deliver ]
variables:
  IMAGE: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
build:
  stage: build
  script:
    - docker build -t $IMAGE .
    - docker push $IMAGE
test:
  stage: test
  script: [ "run unit + integration tests" ]
scan:
  stage: scan
  script: [ "trivy image --severity HIGH,CRITICAL --exit-code 1 $IMAGE" ]
deliver:
  stage: deliver            # write the new tag to the config repo; Flux/Argo CD deploys
  script:
    - git clone https://oauth2:$CONFIG_TOKEN@gitlab.com/org/config-repo.git
    - cd config-repo && kustomize edit set image app=$IMAGE --kustomization overlays/prod
    - git commit -am "deploy $CI_COMMIT_SHA" && git push
```

Common orchestrators: Jenkins, GitLab CI/CD, GitHub Actions, Azure DevOps, Travis CI, plus Kubernetes-native
**Tekton Pipelines** and **Argo Workflows**. Argo CD/Flux are the CD/orchestration layer the CI feeds. Whatever
the tool, keep the same shape: build → test → scan → push (SHA tag) → write to Git → let the agent reconcile.
