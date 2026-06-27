# Developer Portals, Onboarding, and Monetization

Table of contents:
1. APIs as products
2. Developer experience (DX) and the portal
3. What an API page must contain
4. Onboarding and channels
5. Monetization: what it really means
6. Plans and pricing models
7. Billing
8. Putting it together

---

## 1. APIs as products

API products are intangible goods that satisfy a real consumer need — quick access to functionality, fresh data, or expensive computation the consumer would rather not build (e.g. embedding Google Maps instead of building mapping; calling a forex API instead of scraping rates). An API product delivers a *clearly understood customer benefit*, internal or external. Treating APIs as products means the same design thinking, ownership, ongoing attention, and evolution you'd give any product — and it can turn an integration cost center into a profit center.

Consequences of "API as product": it has an owner, a roadmap, documentation, SLAs, a target audience, plans/pricing, marketing, and support — not just an endpoint.

---

## 2. Developer experience (DX) and the portal

The single most repeated truth: **an API is only as good as its documentation**, and it isn't "done" until the docs and portal page exist. DX is the discipline of making APIs easy to **find, understand, subscribe to, and use**.

A common misconception: "the OpenAPI spec is the documentation." The spec is *part* of it, not enough on its own. Good DX needs narrative docs, examples, an interactive try-it console, and self-service onboarding.

The **developer portal** is the storefront: it lists the org's APIs and lets internal/external developers discover, read, subscribe to, get keys for, and try APIs. Most API management products include a portal (some as a paid add-on). Examples of the broader ecosystem: API **directories** (list + link, e.g. ProgrammableWeb-style), **marketplaces** (list + subscribe + use across orgs, e.g. RapidAPI), and **hubs** (self-service integration of published APIs, e.g. IFTTT-style).

---

## 3. What an API page must contain

A complete API page (per API, per version) includes:

- **Getting started** — onboarding/registration steps, how to get credentials and an app key, a 5-minute "first successful call".
- **Overview** — what the API does and its key features, in plain language.
- **Authentication & authorization** — exactly how to authenticate (keep this section as simple as possible; it's where most developers get stuck). Show the OAuth flow, scopes, and a worked token example.
- **Resource/operation reference** — every resource (or GraphQL operation), its parameters, supported verbs, and **interactive** sample request/response payloads (try-it from the page itself).
- **Errors** — error model, codes, and what each means.
- **Constraints** — rate limits, quotas, payload sizes, plan differences.
- **Versioning / changelog** — current, previous, and upcoming versions, and how to switch.
- **Mock endpoint** — the URL of the mock and an embedded console to experiment.
- **Code scaffolding** — generate client (and sometimes server) SDKs in multiple languages from the spec.
- **Commercial pages** — pricing/plans, payment options, subscription management, terms & conditions.

---

## 4. Onboarding and channels

Make onboarding **self-service**: register → pick a plan → auto-provision an app + key/OAuth client → try in the console → go. Friction here directly suppresses adoption.

Publish to multiple channels as appropriate: the org's own portal (primary), directories (discovery + link back), public marketplaces (cross-org subscribe + use), and hubs (let people compose your API into integration flows). External vs internal audiences may warrant different portals or visibility tiers.

A **content manager / API evangelist** role keeps tutorials, how-tos, and videos fresh and reduces support load — content quality is a first-order driver of adoption, not an afterthought.

---

## 5. Monetization: what it really means

API monetization is broader than "charge per call" — it is **driving revenue through the use of APIs**, directly or indirectly. Direct = the consumer pays to call. Indirect = the API enables revenue elsewhere (drives partner sales, increases platform stickiness, unlocks data network effects, reduces cost). Many of the most valuable API programs are *indirectly* monetized.

Three capabilities are required:
- **API plans** — define monetization schemes (direct, indirect, or custom).
- **API monetization policies** — enforce a plan on an API or group of APIs at the gateway (quotas/rate limits tied to the plan).
- **API billing** — meter usage and bill/invoice per the plan, either via built-in finance features or integration with a billing system (Stripe, Recurly, Zuora, etc.).

---

## 6. Plans and pricing models

| Model | How it works | Fits |
|---|---|---|
| **Free / internal** | No charge; governs access & quota only | Internal reuse, partner enablement |
| **Freemium** | Free tier + paid upgrades | Acquisition funnel, developer adoption |
| **Pay-per-call (metered)** | Charge per request, often tiered | Usage-correlated value (data, compute) |
| **Tiered subscription** | Flat fee per tier with a quota (Gold = 10k/day) | Predictable revenue, simple billing |
| **Per-transaction / value-based** | Charge per business event (payment processed) | When call ≠ value |
| **Revenue share / affiliate** | Provider pays the consumer for value they drive | Marketplaces, distribution partners |
| **Pay-as-you-go + commit** | Metered with a committed minimum | Enterprise |

Tie each plan to **quotas and rate limits** (see `traffic-management.md`): the Gold plan's "10k calls/day" is a quota enforced at the gateway; exceeding it either blocks or bills overage. The more valuable/fresh the data or compute, the higher the price typically scales.

---

## 7. Billing

- **Meter** accurately at the gateway/analytics layer (count per app key/subscription), reconciled with the plan window.
- **Rate & invoice** per plan; either use the platform's finance features or integrate a billing engine.
- Handle **overage**, proration, refunds, taxes, and dunning if you do real money.
- Give consumers a **usage dashboard** in the portal (current consumption vs quota, cost-to-date) — transparency reduces disputes and support load.

---

## 8. Putting it together

A monetizable API product = a well-designed, documented API + a portal page with self-service onboarding + plans wired to quotas/rate limits + metering + billing + analytics + an owner who treats it as a product on a roadmap. Skipping any one of these (most often: docs, onboarding, or analytics) is the usual reason an otherwise good API gets no adoption.
