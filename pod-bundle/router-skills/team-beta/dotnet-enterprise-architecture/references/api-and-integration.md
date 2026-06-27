# API and integration design

Contents:
- Contract-first REST and OpenAPI
- Aligning APIs to bounded contexts
- Standards and pivotal formats
- Putting history/time in an API
- Versioning and backward compatibility
- Hypermedia, links, and embedded data
- API gateways
- Testing the contract

## Contract-first REST and OpenAPI

Design the **OpenAPI contract before the implementation**, not the other way around. Generating the spec from controller code (code-first) yields a monolithic, un-segregated contract that mirrors implementation accidents rather than the business. Contract-first forces you to think in the consumer's terms and keeps the contract the source of truth.

REST done well reuses HTTP, not a proprietary layer:
- **Verbs** carry intent: `GET` (read, safe), `POST` (create), `PATCH` (partial update), `PUT` (full replace/idempotent), `DELETE`.
- **Resources** are business entities with URLs (`/books`, `/books/{id}`, `/authors/{id}/addresses`).
- **Status codes** mean what they say: `201 Created` with a `Location` header on create, `404` for missing, `409` for conflicts, `403` for forbidden.
- **Representations** in JSON; content negotiation via `Accept`.

## Aligning APIs to bounded contexts

The clean alignment to aim for (one row implies the next):

```
one bounded context
  → one major entity (aggregate root)
    → one API contract (OpenAPI)
      → one code repository
        → one deployable / Docker image
          → one test collection
```

Minor entities are nested under their parent (`/authors/{id}/addresses`), not given top-level contracts. Apply **interface segregation**: split `/books` (catalog data) from `/books/sales` (sales data) so consumers depend only on what they use and each can version independently.

## Standards and pivotal formats

Before inventing a schema, look for an existing standard — there is one for almost everything (ISO 8601 dates, ISO 4217 currencies, ISO 3166 countries, ISO 639 languages, ISBN, OpenID Connect/OAuth 2.0 + JWT for auth, OData/GraphQL for query, JSON Patch RFC 6902, CMIS for documents, BPMN/DMN for process/rules). A standard is *technically written but not technically restrictive* — precise enough to remove ambiguity, neutral enough not to bind you to a platform.

You rarely need to implement a whole standard — use the slice you need (10% of CMIS to send/query documents) and lean on a library/product that already implements it (Keycloak for OIDC, a Mongo driver speaking the wire protocol).

Where no standard fits, define a **pivotal format**: a neutral, business-described schema that plays the role of a standard within your scope. Build it from existing norms for its inner fields (ISBN, ISO dates). Express it technically (JSON Schema, OpenAPI) but keep it free of any implementation detail. A good pivotal format lets two parties evolve independently — neither depends on the other's tech, only on the shared contract. (This is how cross-org integrations stay decoupled: agree the JSON, then each side is free.)

## Putting history/time in an API

If the domain model includes history (see `domain-driven-design.md` and `persistence-patterns.md`), the API must expose it:

- `GET /books` — list.
- `GET /books/{id}` — current (best-so-far) state.
- `GET /books/{id}?valueDate={iso}` — state at a point in time.
- `GET /books/{id}/history` — full change history.
- `POST /books` — create (returns `201` + `Location`).
- `PATCH /books/{id}` — apply a JSON Patch delta (the primary write; records a change).
- `PATCH /books/{id}?valueDate={iso}` — back-dated change where allowed.
- `PUT /books/{id}` — often *disallowed*: replacing the whole state destroys the lock-free, history-preserving design.
- `DELETE /books/{id}` — usually a status change to archived, not physical removal (true deletion only for regulation).

JSON Patch body (RFC 6902):

```json
[ { "op": "replace", "path": "/title", "value": "Performance in .NET, Second Edition" } ]
```

## Versioning and backward compatibility

- **Backward compatibility** = every call valid in v1 produces the same result on the new version. Treat it as **Liskov for services**.
- Prefer **additive** evolution (new optional fields, new endpoints) over breaking changes. Add a new versioned contract (`/v2`) only when you must break.
- **Hyrum's Law**: with enough consumers, *every observable behavior* — even undocumented quirks, field ordering, timing, a fixed bug — *will* be depended on by someone. The more successful the API, the more forward-compatibility constrains you. Plan for it: document intended behavior, avoid relying on incidental output order, and treat performance characteristics as part of the contract for hot APIs.

ASP.NET Core: use `Asp.Versioning` (the maintained successor to `Microsoft.AspNetCore.Mvc.Versioning`) for URL-segment, header, or query versioning.

## Hypermedia, links, and embedded data

Link related resources rather than duplicating them, but embed a little frequently-needed data to save round-trips:

```json
{
  "isbn13": "978-2409002205",
  "title": "…",
  "links": [
    { "rel": "self",   "href": "https://acme.com/library/books/978-2409002205" },
    { "rel": "author", "href": "https://acme.com/authors/202312-007", "title": "JP Gouigoux" }
  ]
}
```

Decide functionally whether an embedded value (the author's name in a book link) should track changes or stay frozen — and remember a GDPR erase of the author must scrub embedded copies too. Don't embed volatile data (a phone number) that would force mass updates on change.

## API gateways

A gateway (reverse proxy) fronts your APIs to centralize cross-cutting operational concerns: routing, authentication, rate limiting, logging, and usage metering, hiding the real endpoints behind one facade. .NET options: **YARP** (Microsoft's reverse proxy, the current default), or Ocelot for a lighter API-gateway library; full API-management platforms (Azure API Management, etc.) add developer portals and monetization. A gateway also enables the open/closed extension pattern — expose an updated contract that mixes original-API data with additional data from a new service.

## Testing the contract

- **Contract / API tests** are your regression and backward-compatibility net at the integration boundary — above unit and integration tests. Maintain a collection of real consumer calls (e.g. a Postman/Bruno/`.http` collection) and run it in CI; gather what partners actually do with the API and fold those calls in.
- Validate responses against the OpenAPI schema so undocumented drift fails the build.
- For internal logic, keep unit tests on handlers/aggregates and integration tests on repositories — but treat the API contract as the unifying regression surface, since it is where every consumer meets your system.
