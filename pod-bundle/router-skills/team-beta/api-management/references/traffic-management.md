# Traffic Management

Rate limiting, quotas, throttling, caching, and load balancing — the controls that keep APIs fast, fair, and protected.

Table of contents:
1. Rate limit vs quota vs throttle (don't conflate them)
2. Rate-limiting algorithms
3. Quotas and plans
4. Throttling / spike arrest
5. Response codes & headers
6. Caching
7. Load balancing & resilience

---

## 1. Rate limit vs quota vs throttle

These three are routinely confused; they solve different problems and often run together.

- **Rate limit** — a *hard cap on call count over a short window* (e.g. 100 req/min) per app, user, IP, or route. Purpose: protect against abuse, runaway clients, and denial-of-service. Excess requests get `429`.
- **Quota** — a *business/plan limit over a long window* (e.g. 10,000 calls/day on the Gold plan). Purpose: monetization and fair-use enforcement tied to a subscription. Exceeding it may block for the rest of the period or trigger overage billing.
- **Throttling / spike arrest** — *smooths the rate of flow* so a burst doesn't slam the backend (e.g. allow at most 100 transactions/second, queue or reject the rest). Purpose: protect downstream services and shared infrastructure from being overwhelmed.

A typical stack: spike-arrest (per second, protect backend) + rate-limit (per minute, protect against abuse) + quota (per day/month, enforce the plan).

---

## 2. Rate-limiting algorithms

| Algorithm | How it works | Trade-off |
|---|---|---|
| **Fixed window** | Count resets each interval (e.g. per minute) | Simple, but allows a 2x burst at the window boundary |
| **Sliding window (log)** | Track timestamps over a rolling window | Accurate, more memory |
| **Sliding window (counter)** | Weighted blend of current+previous fixed windows | Good accuracy/cost balance — common default |
| **Token bucket** | Bucket refills at a steady rate; each request takes a token | Allows controlled bursts up to bucket size; great for spiky-but-bounded traffic |
| **Leaky bucket** | Requests drain at a fixed rate (a queue) | Smooths output rate; adds latency under burst |

For distributed gateways, counters must be **shared** (e.g. Redis-backed) or you get N× the intended limit across N nodes. Decide whether you want *local* (fast, approximate) or *cluster-wide* (accurate, needs a shared store) counting — Kong's `rate-limiting-advanced`, Apigee's distributed quota, and APIM's policies all expose this choice.

Choose the limit *key* deliberately: by API key/app (plan enforcement), by authenticated user (fairness), by client IP (anonymous abuse), or by route (protect a hot endpoint).

---

## 3. Quotas and plans

Quotas implement the *business* side of traffic and underpin monetization:
- Attach a quota to a **plan**, not an API directly (Free: 1k/day, Pro: 100k/day, Enterprise: custom).
- Track consumption per **subscription/app key**; expose remaining quota to consumers (`X-RateLimit-Remaining`, or a portal usage dashboard).
- Define overage behavior: hard block, soft warn, or metered overage billing.
- Reset windows (daily/monthly) must align with billing periods.

See `developer-portals-and-monetization.md` for how plans connect to billing.

---

## 4. Throttling / spike arrest

Spike arrest enforces a *maximum instantaneous rate* (e.g. 30 tps) by smoothing — Apigee's `SpikeArrest` literally converts "30 per second" into "one allowed roughly every 33 ms", rejecting bursts even if the per-minute count is under the rate limit. Use it to shield a backend that genuinely can't handle bursts, independent of the longer-window quota/rate limit. Pair with backend **timeouts, retries (with backoff + jitter), and circuit breaking** so a slow backend sheds load instead of cascading failures.

---

## 5. Response codes & headers

- Over rate limit / quota → **`429 Too Many Requests`** with **`Retry-After`** (seconds or HTTP date).
- Backend overloaded / shedding → `503 Service Unavailable` (+ `Retry-After`).
- Expose standard informational headers so clients can self-pace: `RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset` (IETF draft headers; many gateways still emit the `X-RateLimit-*` variants). Document which you use.

---

## 6. Caching

Caching cuts latency and backend load by serving stored responses.

**Where**:
- **Client/browser** and CDN/edge — controlled by HTTP cache headers.
- **Gateway response cache** — the gateway stores a backend response keyed by URL + selected headers/params; a subsequent matching request is served without calling the backend.

**HTTP cache controls** (REST's big advantage — it's HTTP-native): `Cache-Control: max-age=...`, `ETag` + `If-None-Match` (→ `304 Not Modified`), `Last-Modified` + `If-Modified-Since`, and `Vary` to key on headers (e.g. `Accept`, `Authorization`).

**Rules**:
- Cache only **safe, cacheable** responses — generally `GET`s of data that tolerates some staleness. Never cache responses that vary by user unless the cache key includes the user/token (or you'll leak one user's data to another).
- Set TTLs to the data's tolerance for staleness and provide an **invalidation** path (event-driven purge, short TTL, or `ETag` revalidation) so consumers don't get stale data.
- GraphQL and gRPC don't get HTTP caching for free — caching is implementation-side (persisted queries, field-level caches, client normalized caches); plan for it explicitly.

---

## 7. Load balancing & resilience

- **API load balancing** distributes calls across backend instances (round-robin, least-connections, weighted, latency-based). In Kubernetes the ingress/microgateway typically does this; a sidecar topology pushes it per-service.
- **Geo-routing** sends traffic to the nearest regional gateway for latency and data-residency.
- **Health checks** remove unhealthy backends from rotation.
- **Resilience policies** at the gateway/mesh: timeouts (always set one), retries with exponential backoff + jitter (and a retry budget so you don't amplify load), circuit breakers, and bulkheads to isolate failures.
- For async/long-running work, prefer `202 Accepted` + a status resource or a **webhook** callback over making clients poll — polling wastes both client and network resources.
