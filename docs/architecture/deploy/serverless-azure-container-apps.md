# Deploy: Azure Container Apps (serverless, no AKS)

**Surface:** <https://learn.microsoft.com/en-us/azure/container-apps/overview>
Hub: [Getting started](../../../GETTING_STARTED.md) · Identity:
[`azure`](../agent-identity/azure.md)

## What it is

Azure Container Apps (ACA) is a **serverless container platform** built on AKS +
KEDA + Dapr + Envoy, with the cluster fully managed by Azure. You ship
containers; Azure runs them. No node pools, no kubelet, no control plane to
patch. Key primitives: **scale-to-zero** + KEDA event/HTTP autoscaling,
**Dapr** sidecars (pub/sub, state, service invocation), built-in ingress/mTLS,
revisions for blue-green, and **workload profiles** (Consumption = scale-to-zero
serverless; Dedicated = reserved compute for heavier/long-lived workloads).

## Why a customer picks it

- **Lower ops than AKS** — no cluster, node, or Kubernetes-version management;
  the smallest team can run Open Engine's planes as plain containers.
- **Scale-to-zero = pay-per-use** — the OPA, gateway, and Cachy planes cost
  nothing while idle, then scale on HTTP/queue depth (KEDA).
- **Built-in edge** — managed ingress, TLS, and revisions remove the
  ingress-controller/cert-manager toil AKS demands.
- **Dapr** — pub/sub + state store for the task-queue/receipt contract without
  standing up your own broker.

## Customer scenarios

1. **Pilot / PoC** — a customer wants the governed planes (Cachy :8787, Sentry
   gateway :8080 + OPA :8181) up in a day without owning a cluster. Deploy the
   three stateless planes to Consumption ACA; scale-to-zero between demos.
2. **Bursty governance edge** — high-variance MCP traffic; let KEDA scale the
   gateway/OPA on request rate and pay only for active windows.
3. **No-AKS shop** — a team already standardized on ACA for other services and
   refuses to take on AKS. The governance planes drop in alongside.
4. **Cost-sensitive multi-tenant SaaS** — many small customers, each idle most
   of the time; scale-to-zero per environment keeps the floor near $0.

## Fit & limits for Open Engine

| Plane | ACA fit | Note |
| :-- | :-- | :-- |
| **OPA** (:8181) | ✅ stateless, bundle-driven | scales to zero cleanly; mount `mcp_auth.rego` as the bundle |
| **Sentry gateway** (:8080/mcp) | ✅ stateless HTTP, fails closed | watch `MCP-Session-Id` affinity (sticky sessions) |
| **Cachy** (:8787) | ✅ stateless proxy | KEDA on request rate |
| **Omnigent engine** (:6767) | ◐ **constrained** | see below |

**The engine is the hard part.** Omnigent doesn't make in-process SDK calls — it
**spawns persistent vendor-CLI harnesses** (`claude-native`, `codex-native`, …)
and **local sandboxes** as long-lived child processes that hold session state.
That collides with serverless assumptions:

- **Scale-to-zero kills live sessions.** A persistent claiming CLI is the
  opposite of a request that returns and lets the replica idle out. Set
  **min-replicas ≥ 1** and run the engine on a **Dedicated workload profile**,
  not Consumption.
- **No sticky long-lived process placement.** Revision swaps and replica churn
  evict in-flight harness processes; ACA gives no node affinity to pin them.
- **Sandbox isolation needs privilege ACA won't grant.** bubblewrap/firejail and
  nested-container sandboxes want privileged/seccomp/user-namespace controls;
  ACA containers are unprivileged with no Docker-in-Docker. The sandbox boundary
  the engine relies on can't be reproduced inside a Consumption app.
- **Session↔subject binding.** The `MCP-Session-Id`↔subject join (azure
  `provider`, OID as `subject_id`) requires session-affinity ingress so a
  session lands on the same engine replica.

**Recommended governed path:** run the **stateless governance edge** (OPA +
Sentry gateway + Cachy) on **Consumption ACA** (scale-to-zero, KEDA, managed
TLS), and run the **engine + sandboxes** on a **Dedicated workload profile with
min-replicas=1 and sticky sessions** — or keep execution on a VM/AKS and point
it at the ACA edge. ACA is the right home for the governance planes; it is **not**
a drop-in host for the sandboxed execution plane. Agent identity stays per-agent
(D17): each engine workload runs under its own **Entra Agent ID** managed
identity in a non-admin group, so ACA's managed-identity binding maps 1:1 to the
rego `is_admin` carve-out.

## Follow-ups

- Validate sandbox spawn under an ACA **Dedicated** profile (privilege/seccomp
  reality check) before claiming "serverless execution."
- Bind `MCP-Session-Id` to the authenticated subject (open gateway gap) — sticky
  ingress on ACA is moot until that lands.
- Author the deploy manifest (Bicep/`az containerapp`) + KEDA scale rules: see
  the implementation story below.
- Cross-link from [agent-identity/`azure`](../agent-identity/azure.md) once the
  managed-identity → Entra Agent ID wiring is proven on ACA.
