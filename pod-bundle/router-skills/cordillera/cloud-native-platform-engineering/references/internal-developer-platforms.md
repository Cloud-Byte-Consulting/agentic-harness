# Internal Developer Platforms (IDPs)

Capabilities of an IDP, reference frameworks (CNOE, BACK stack), why Kubernetes
is the substrate, extending Kubernetes with CRDs/operators, managing
infrastructure and clusters as Kubernetes resources, multi-cluster patterns, and
composing CI/CD, observability, and data/XaaS capabilities.

> For *implementation* depth, hand off to the hands-on skills:
> **kubernetes-gitops-cicd** (Argo CD, pipelines), **kubernetes-observability**
> (Prometheus/Loki/OTel), **kubernetes-security-rbac** (RBAC, policy, secrets),
> **kubernetes-cluster-operations** (cluster lifecycle, upgrades). This file is the
> architectural map of *what to compose and why*.

## Contents
- What an IDP is and its capabilities
- Reference frameworks: CNOE and the BACK stack
- Kubernetes as the substrate
- Extending Kubernetes: CRDs and operators
- Managing infrastructure as Kubernetes resources (Crossplane)
- Managing clusters as resources (Cluster API) and multi-cluster patterns
- Composing the capabilities: pipelines, observability, data/XaaS

## What an IDP is and its capabilities

An IDP is a composite of interrelated capabilities exposed through user-friendly
**interfaces** (portal/UI, API, CLI) whose single goal is to accelerate and simplify
software delivery and reduce developer cognitive load. Per the CNCF Platforms white paper,
the core capabilities are:

- **Compute / runtime** — orchestration, scaling, self-healing (Kubernetes).
- **Pipelines** — build/test/deploy automation; infra automation; IaC.
- **Observability & reliability** — metrics, events, logs, traces (MELT); SLOs.
- **Data services** — DBaaS, state stores, data integration, backup/restore.
- **Built-in security** — code/dependency/image scanning, SBOMs, secrets & cert
  management, policy/compliance.

Interfaces should let users **observe, provision, and manage** platform resources. A good
IDP is an orchestra, not a toolbox: each capability is designed around the developer.

## Reference frameworks: CNOE and the BACK stack

Two ways to anchor your architecture:

**CNOE (Cloud Native Operational Excellence)** — a broad, open-source framework that maps
platform capabilities and how to *integrate existing tools*, helping enterprises de-risk
their tooling bets. It is "tools, not practices": Kubernetes-centric but not Kubernetes-only,
open-source-first, community-driven. CNOE addresses the "paralysis of choice" in the CNCF
landscape but does *not* operate your toolchain for you.

**BACK stack** — a prescriptive Minimum Viable Platform from four CNCF projects:
- **B**ackstage — developer portal / self-service interface and catalog.
- **A**rgo CD — GitOps continuous delivery (sync Git desired-state to clusters).
- **C**rossplane — manage cloud infrastructure as Kubernetes resources.
- **K**yverno — policy-as-code (validate/mutate/generate) to enforce standards.

Use CNOE to reason about *which* capabilities you need; use the BACK stack as a concrete
starting point.

## Kubernetes as the substrate

>70% of IDPs are built on Kubernetes because it provides:
- **Unified orchestration** across on-prem/cloud/hybrid (portability).
- **Scalability + self-healing** (reschedules failed pods/nodes).
- **Extensible, modular** plug-and-play architecture.
- A huge **ecosystem, community, and standardized skills/certs** (easier to staff).
- A **declarative control-loop API** (observe → diff → reconcile) that everything else
  builds on.

But: **running Kubernetes ≠ having a platform.** The platform is the developer-facing
*abstraction* layered on top. Kubernetes is the "platform of platforms," not the product.

Supporting non-container workloads on the same substrate:
- **Knative** — serverless functions on Kubernetes (scale-to-zero, event-driven).
- **KubeVirt** — run/manage VMs alongside containers (one pod per VM).

## Extending Kubernetes: CRDs and operators

The control-loop (Observe → Analyze → Act) is what gives Kubernetes self-healing. You
extend it with **Custom Resource Definitions (CRDs)** + **custom controllers**:

- A **CRD** defines a new resource type (name/version/schema) that behaves like a built-in
  object and is managed with `kubectl`.
- A **controller** watches that resource and reconciles current → desired state.
- An **Operator** = CRD(s) + controller(s) + encoded operational knowledge, used to manage
  complex stateful apps (databases, brokers), complex installs/upgrades, multi-tenant
  isolation, and auto-remediation (failover/backup).

Tooling: **Kubebuilder** or the **Operator SDK** scaffold the project, generate CRD
manifests, and let you implement reconcile logic, e.g.:

```bash
# Operator SDK
operator-sdk init --domain example.com --repo github.com/example/my-operator
operator-sdk create api --group webapp --version v1 --kind AppService --resource --controller
make generate && make manifests   # generate CRD manifests
make install                       # install CRDs into the cluster
make deploy IMG="myrepo/my-operator:v0.1"
```

The key idea: **encode operational expertise as software** so day-2 operations become
declarative and automated.

## Managing infrastructure as Kubernetes resources (Crossplane)

Extend the operator pattern to *external* cloud resources: define an RDS instance, S3
bucket, etc. as a Kubernetes object and let a controller provision/manage it — same API,
tooling, GitOps, and review workflow developers already know.

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm install crossplane --namespace crossplane-system --create-namespace crossplane-stable/crossplane
```

After installing a provider (e.g. AWS) and a `ProviderConfig` (credentials via a Secret),
declare resources, for example a managed database (provider CRDs evolve — check current
API versions for your provider build):

```yaml
apiVersion: rds.aws.crossplane.io/v1alpha1
kind: DBInstance
metadata:
  name: example-rds-instance
spec:
  forProvider:
    dbInstanceClass: db.t3.micro
    masterUsername: masteruser
    allocatedStorage: 20
    engine: mysql
    region: us-west-2
  providerConfigRef:
    name: default
```

Benefits: a **unified, declarative management interface**, improved developer
productivity (provision without deep cloud expertise), automation, and seamless integration
with the K8s ecosystem. Use Crossplane **Composite Resources (XRs)** to bundle groups of
resources that must be provisioned together — this is how you build higher-level
self-service abstractions (e.g. "give me a production-ready database").

Alternatives in the same Kubernetes-native-IaC space: **AWS ACK / AWS Service Operator**,
**Google Config Connector**, vendor operators (e.g. **MongoDB Atlas Operator**). All enable
the same GitOps-based experience. (For Terraform/OpenTofu vs Kubernetes-native IaC trade-offs,
see kubernetes-cluster-operations and kubernetes-gitops-cicd.)

## Managing clusters as resources (Cluster API) and multi-cluster patterns

**Cluster API (CAPI)** brings declarative, Kubernetes-style APIs to *cluster* lifecycle —
create/scale/upgrade/delete clusters across clouds from a **management (hub) cluster**.
You define `Cluster`, infra-provider objects (e.g. `AWSCluster`/`AWSMachineTemplate`), and
`MachineDeployment`; CAPI uses kubeadm under the hood. Use `clusterctl` to generate the
workload-cluster manifests, then `kubectl apply` them from the management cluster.

How many clusters? The early "one big cluster" approach hits scalability limits, risk
concentration (single point of failure), security/compliance complexity, and resource
contention. The trend is **multiple smaller clusters**. Decide cluster-per-app vs
cluster-per-environment by weighing **isolation** (per-app = max isolation), **cost
efficiency** (more clusters = more overhead), **resource management**, and **security/
compliance** (different apps, different requirements).

Multi-cluster topologies:
- **Cluster federation** — manage many clusters as one (sync config/policy/deploys);
  good for HA, DR, global load balancing.
- **Hub-spoke** — a central hub cluster governs spoke clusters (global policy, config,
  monitoring) while spokes operate independently. Centralized control, decentralized
  operation. This is the common IDP pattern.
- **Cluster replication** — replicate a primary cluster to replicas for HA/DR/geo.

Connectivity: a **service mesh** (Istio, Linkerd) handles intra-cluster service-to-service
concerns (discovery, mTLS, observability, traffic control). **Skupper** is a lightweight L7
interconnect for secure cross-cluster communication without VPNs/firewall changes, deployable
per-namespace. (For mesh depth, see kubernetes-observability / kubernetes-security-rbac.)

## Composing the capabilities

These are the capabilities to assemble; deep how-to lives in the hands-on skills.

**Pipelines / CI-CD.** A common golden pattern is **CI in GitHub Actions** (build/test/scan
on push/PR) handing off to **CD via Argo CD GitOps** (Git is the source of truth; Argo CD
reconciles cluster state to the repo). An Argo CD `Application` points at a repo path; an
`ApplicationSet` templates many Applications across clusters/environments — the backbone of
multi-cluster, multi-env delivery. → **kubernetes-gitops-cicd** for the full treatment.

**Observability.** Prometheus for metrics, **Thanos** for HA + long-term storage + a global
multi-cluster query view, **Loki/Promtail** for logs, **OpenTelemetry Collector** as a
vendor-agnostic central ingest for metrics/logs/traces. Standardize on **SLOs** to avoid the
"watermelon problem" (green on the surface, red inside) caused by teams reporting health
differently. → **kubernetes-observability** for full configs and MELT pipelines.

**Data services / XaaS.** Expose **Database-as-a-Service** through an operator + GitOps so
teams request a DB by committing YAML. Example: the **MongoDB Atlas Operator** watches
`AtlasProject`/`AtlasDeployment` CRs; Argo CD syncs them from Git and the operator
provisions the real Atlas resources. The same pattern generalizes to **AWS ACK**,
**Crossplane**, and **Google Config Connector** — "everything-as-a-service" via declarative,
GitOps-managed Kubernetes resources, shifting provisioning left to product teams.

**Self-service interfaces.** A Backstage portal for discovery/scorecards/docs; GitOps for
the developer "deploy" workflow; built-in security woven through (scanning, SBOM, secrets/
cert management, policy-as-code with Kyverno/OPA). See
`golden-paths-and-self-service.md` and `finops-and-governance.md`.
