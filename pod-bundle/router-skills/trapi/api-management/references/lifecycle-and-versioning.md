# API Lifecycle and Versioning

Table of contents:
1. The core API lifecycle
2. Design-first loop
3. Service lifecycle & contract testing
4. Consumer lifecycle
5. Versioning strategies
6. Backward compatibility rules
7. Deprecation and retirement

---

## 1. The core API lifecycle

An API is a product with a never-ending (while successful) lifecycle. Always precede it with an API *strategy* that defines why the API exists and what business goal it serves — otherwise the steps lack relevance. The core loop:

1. **Ideation & planning** — identify the business capability and consumers; justify the API against goals.
2. **Design** — model the interface, write the spec (OpenAPI/SDL/.proto/AsyncAPI).
3. **Mock & try** — stand up a mock from the spec so consumers can try it before code exists.
4. **Create / configure** — implement the service and configure the gateway (policies, routing).
5. **Deploy** — promote across environments via CI/CD.
6. **Promote** — publish to the portal, make discoverable (beta → GA).
7. **Deprecate** — announce, set a sunset timeline, run old + new in parallel.
8. **Retire** — decommission once consumers have migrated.
9. **Observe** — throughout: monitor usage, errors, latency; feed insights back to design.

Each step has owners and tools; the loop repeats as the product evolves.

---

## 2. Design-first loop

**Design-first** (a.k.a. spec-first / contract-first) means agreeing the contract *before* implementation:

```
draft spec → mock → share with consumers → collect feedback → revise spec → (repeat) → implement
```

Why: reworking a published, implemented API is expensive; reworking a spec+mock is cheap. Early consumer feedback catches modeling mistakes while they're free to fix. This demands real, fluent communication channels (shared spec repo, chat, an API product owner actively soliciting feedback) — without them the feedback is vague and the loop's value collapses.

Tooling: write OpenAPI/AsyncAPI in an editor (Stoplight, Apicurio, VS Code + linters), mock with **Prism**/Microcks/WireMock, generate docs and SDKs. Contrast with **code-first** (generate the spec from annotations after the fact) — acceptable for internal/low-stakes APIs, risky for public ones because the contract becomes an afterthought.

---

## 3. Service lifecycle & contract testing

The service that implements the API has its own loop: **scaffold/refactor → build → unit test → contract test**. Generate server stubs from the spec to start.

**Contract (interface) testing** verifies the running service matches its published spec. Because REST decouples the IDL (OpenAPI/Blueprint/RAML) from the implementation, the service *can* drift — contract tests prevent shipping a service that doesn't honor its contract.

- Spec-driven: **Dredd** and **Schemathesis** introspect an OpenAPI spec and auto-generate tests; Postman/Newman and ReadyAPI can validate too.
- Consumer-driven: **Pact** captures consumer expectations and verifies providers against them — strongest for microservice fan-out where many consumers depend on one provider.
- Make contract tests a gate in CI: code isn't "done" until they pass.
- For GraphQL and gRPC the IDL is tightly coupled to the implementation, so standalone contract testing matters less — schema/`.proto` compatibility checks (e.g. Buf breaking-change detection, GraphQL schema diff) cover it instead.

---

## 4. Consumer lifecycle

Mirror the producer side with the consumer journey: **discover (portal) → onboard/register → obtain credentials → implement & integrate → use → give feedback**. The producer's job is to make each step self-service and low-friction (see `developer-portals-and-monetization.md`). Feedback from real consumption feeds the next design iteration.

---

## 5. Versioning strategies

REST has no single agreed answer (a genuine industry debate); pick one and apply it consistently across the platform:

| Strategy | Example | Notes |
|---|---|---|
| **URI path** | `/v1/orders` | Most explicit, cache- and gateway-friendly, easy to route. Common default for public APIs. Version the *major* only. |
| **Header / media type** | `Accept: application/vnd.example.v2+json` | Keeps URLs stable, "purer", but harder to test/cache and less discoverable. |
| **Query param** | `/orders?version=2` | Easy but pollutes every request; weaker caching. |
| **No version** | evolve additively | Works if you commit to never breaking — the GraphQL philosophy. |

**GraphQL**: don't version. Add fields; mark removals with `@deprecated(reason: ...)`; monitor field usage before removing. **gRPC/protobuf**: backward-compatible by field-number discipline; if you must, keep package-versioned `.proto` (`orders.v1`, `orders.v2`) but avoid a proliferation of live versions.

Version the **major** (breaking) only. Patch/minor changes that are additive don't need a new version.

---

## 6. Backward compatibility rules

Within a published major version, only **additive, non-breaking** changes are allowed:

**Safe (non-breaking):** add a new endpoint/operation; add an optional request field; add a field to a response; add a new optional query param; add a new enum value *only if clients are told to tolerate unknowns*; relax a validation constraint.

**Breaking (require a new major):** remove or rename a field/endpoint; change a field's type or format; make an optional field required; tighten validation; change default behavior, error codes, or response structure; remove an enum value.

When in doubt, assume a change is breaking. Use contract tests + schema-diff tooling in CI to *detect* breaking changes automatically before release.

---

## 7. Deprecation and retirement

Run multiple versions concurrently so consumers migrate on their own schedule, then sunset:

1. **Announce** deprecation with a clear timeline and migration guide (portal changelog, email, `Deprecation` / `Sunset` HTTP headers per RFC 8594).
2. **Run old + new in parallel**; route via the gateway. Track who still calls the old version (analytics by app key).
3. **Nudge** remaining consumers; consider increasing friction (warnings, then brownouts) as the date nears.
4. **Retire** once usage hits zero (or the announced date passes). Decommission the version and its backend.
5. Use **observability** to *find* retirement candidates — low/declining usage, high error rates, redundancy with a newer API. Keeping zombie versions alive is an "improper inventory management" security risk (OWASP API9), not just tech debt.
