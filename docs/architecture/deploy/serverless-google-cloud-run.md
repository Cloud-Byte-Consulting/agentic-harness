# Deploy: Google Cloud Run — serverless containers for Open Engine

**Hub:** [Getting Started](../../../GETTING_STARTED.md) · **Identity:** [GCP Agent Identity](../agent-identity/google.md) · **Native plane:** [ActPlane eval](../eval-actplane-native-plane.md)

## What it is

Cloud Run runs a container image as a stateless, request-driven service — Google
provisions, scales, and patches the host; you ship an image and a port. No
cluster, no node pool, no `kubectl`. It **scales to zero** (pay only while a
request is in flight), scales out per concurrent request, and gives each service
a managed HTTPS endpoint with a Google-issued cert. It is the no-GKE path: the
same OCI images, none of the Kubernetes operational surface.

## Why a customer picks it

- **Scale-to-zero economics** — an idle policy gateway or LLM proxy costs $0.
  Bursty agent traffic (a CI run fires 200 `tools/call`s, then nothing for an
  hour) is exactly the request-based billing sweet spot.
- **No cluster to run** — no control plane, upgrades, node patching, or capacity
  planning. A two-person team can stand up the whole governed plane.
- **Fast, safe rollout** — revision-per-deploy, traffic splitting, instant
  rollback. Concurrency and min/max instances are one flag each.
- **Native GCP identity** — a Cloud Run service runs *as* a GCP identity and
  mints ID tokens for service-to-service auth with no key files — the same
  identity story the [`google` agent-identity doc](../agent-identity/google.md)
  builds on.

## Customer scenarios

1. **Managed policy gateway** — deploy **Agentic-Sentry** (`:8080/mcp`) +
   sidecar/colocated **OPA** (`:8181`) as one Cloud Run service. `min-instances=1`
   so the fail-closed gate never cold-starts on the critical path; the OPA bundle
   ships baked into the image (or pulled at boot). Bearer auth via the existing
   `OIDC_PROVIDERS` / `GATEWAY_API_TOKEN`.
2. **Cachy as a regional LLM proxy** — **Cachy** (`:8787`) scaled per-request in
   front of provider APIs; scale-to-zero between bursts, autoscale during a big
   agent run. Token/cost attribution still flows to token-dashboard.
3. **Per-customer isolated plane** — one Cloud Run service (or project) per
   tenant, each as its own GCP service account, no shared cluster blast radius —
   cheap because idle tenants cost nothing.
4. **PoC / pilot** — stand up the governed path for an eval without committing to
   GKE; graduate to GKE only if/when the native plane needs it (see limits).

## Fit & limits for Open Engine

| Component | Fit on Cloud Run |
| :-- | :-- |
| **Agentic-Sentry + OPA** (MCP plane) | ✅ Strong. Stateless HTTP, request-scoped. Pin `min-instances=1` and treat fail-closed as a startup-probe gate. |
| **Cachy** (`:8787`) | ✅ Strong. Stateless proxy, bursty — ideal scale-to-zero. |
| **omnigent engine** (`:6767`) | ◐ Partial. The HTTP engine fits, but the **ActPlane eBPF native-plane gate (CLO-10) needs kernel/BPF access Cloud Run does not grant** — eBPF enforcement requires GKE (privileged) or a VM. On Cloud Run the native plane degrades to the userspace `opa_delegate` hook only. |
| **Long agent sessions / sandboxes** | ❌ Request-timeout-bound (max 60 min) and no persistent local FS; long sessions and per-session sandboxes belong on GKE/VMs. |

**Rule of thumb:** Cloud Run for the **stateless governed edge** (Sentry+OPA,
Cachy); **GKE/VM** where the **kernel-backed native plane or long-lived sandboxes**
live. The choke point (`mcp_auth.rego`, fail-closed) ports cleanly; the eBPF
enforcement does not.

## Follow-ups

- **Code:** ship a deployable Cloud Run bundle for Sentry+OPA (Dockerfile,
  `service.yaml`, bundle-at-boot, `min-instances=1`) — see implementation story
  below.
- Confirm OPA bundle delivery on Cloud Run (baked image vs. boot-time pull from a
  bucket) and that fail-closed holds during a cold start / revision swap.
- Decide native-plane posture per deployment: document "Cloud Run = userspace
  `opa_delegate` only; eBPF requires GKE" in the install matrix.
- Cross-check with the GKE deploy doc (when written) for the split-plane topology.
