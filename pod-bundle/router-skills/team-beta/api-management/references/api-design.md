# API Design: styles, modeling, and specifications

Table of contents:
1. Choosing a style (decision table)
2. REST design
3. OpenAPI 3.1
4. GraphQL design
5. gRPC / Protocol Buffers
6. AsyncAPI (event-driven)
7. Cross-cutting design rules

---

## 1. Choosing a style

| Criterion | REST | GraphQL | gRPC |
|---|---|---|---|
| Best for | Public/partner CRUD-ish resource APIs | Client-driven data needs, aggregation (mobile, BFF) | Internal service-to-service, high throughput, streaming |
| Transport | HTTP/1.1+ , JSON | HTTP, single endpoint, JSON | HTTP/2, binary (protobuf) |
| Tooling maturity | Highest (mocking, codegen, gateways) | Good and growing | Growing; `protoc` generates client+server |
| Composition (combine many sources) | Weak (resource-bound) | Native (each field can resolve from a different source, in parallel) | Neutral (RPC, no constraint) |
| Caching | Native HTTP caching, gateway support | Implementation-dependent; weaker | Implementation-dependent; weakest |
| AuthN/AuthZ | OAuth2/OIDC out of the box on most gateways | Single URI works, but authZ usually custom in resolvers | OAuth/OIDC possible but custom plumbing |
| Versioning | Debated (URI vs header vs none) | Avoid versions; evolve schema | protobuf is backward-compatible by field number |
| Browser fit | Excellent | Excellent | Poor (needs grpc-web/proxy) |
| Async / streaming | Add webhooks/SSE | Subscriptions | Native bidirectional streaming |

**Rule of thumb:** REST is the safe default for externally exposed APIs. Reach for GraphQL when one client must compose data from many backends and you want the client to choose fields. Use gRPC for internal, high-volume, low-latency, or streaming service-to-service traffic — it complements (not replaces) a REST/GraphQL edge.

---

## 2. REST design

REST (Roy Fielding, 2000) is an architectural *style*, not a standard — a set of constraints, not a wire format. In practice "REST/web API" means: resources identified by URIs, manipulated with HTTP verbs, stateless requests, JSON payloads.

**Resource modeling**
- Resources are **nouns**, collections are **plural**: `/orders`, `/orders/{orderId}`, `/orders/{orderId}/line-items`.
- Verbs come from HTTP, not the path. `GET /orders` (no `/getOrders`).
- Sub-resources express containment; cross-cutting actions that don't fit CRUD can be modeled as a sub-resource (`POST /orders/{id}/cancellation`) rather than an RPC-style verb in the path.

**HTTP methods & idempotency**
- `GET` (safe, cacheable), `POST` (create, not idempotent), `PUT` (full replace, idempotent), `PATCH` (partial update), `DELETE` (idempotent).
- Support an `Idempotency-Key` header on `POST` for safe retries on payment/order creation.

**Status codes (use them correctly)**
- `200` OK, `201` Created (+ `Location`), `202` Accepted (async), `204` No Content.
- `400` bad request, `401` unauthenticated, `403` authenticated-but-forbidden, `404` not found, `409` conflict, `422` validation failed, `429` rate-limited (+ `Retry-After`).
- `500` server error, `502/503/504` upstream/availability. Never return `200` with an error body.

**Pagination** — offset/limit for simple cases, cursor (keyset) for large or frequently-changing sets:
```
GET /orders?limit=100&offset=200          # page 3 of 100
GET /orders?limit=100&cursor=eyJpZCI6...   # cursor/keyset, stable under inserts
GET /orders?limit=100&changed_since=2026-01-01T00:00:00Z
```
Offset pagination is simple but degrades on deep pages and skips/duplicates rows when data shifts; cursors avoid that. Always cap and document the max `limit`.

**Error envelope** — pick one shape and use it everywhere. RFC 9457 *Problem Details* is the current standard:
```json
{
  "type": "https://api.example.com/problems/insufficient-funds",
  "title": "Insufficient funds",
  "status": 422,
  "detail": "Account 12345 has a balance of 30.00, transfer of 50.00 was requested.",
  "instance": "/transfers/abc-123",
  "errors": [{ "field": "amount", "message": "exceeds available balance" }]
}
```

**Filtering / sorting / sparse fields**: `?status=open&sort=-created_at&fields=id,total`. Document them; don't invent ad hoc query syntax per endpoint.

**HATEOAS**: optional. Hypermedia links can decouple clients, but over-use makes REST chatty and most clients ignore them. Don't force it.

---

## 3. OpenAPI 3.1

OpenAPI (formerly Swagger; now an OpenAPI Initiative / Linux Foundation project) is the dominant REST IDL, written in YAML or JSON. **3.1 is current** and is a full superset of JSON Schema 2020-12 (a key fix over 3.0). Use `components` for reuse of schemas, parameters, responses, and `securitySchemes`.

```yaml
openapi: 3.1.0
info:
  title: Orders API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /orders/{orderId}:
    get:
      operationId: getOrder
      security: [{ oauth2: [orders:read] }]
      parameters:
        - name: orderId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        "200":
          description: The order
          content:
            application/json:
              schema: { $ref: "#/components/schemas/Order" }
        "404": { $ref: "#/components/responses/NotFound" }
components:
  securitySchemes:
    oauth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://auth.example.com/authorize
          tokenUrl: https://auth.example.com/token
          scopes: { "orders:read": "Read orders", "orders:write": "Write orders" }
  schemas:
    Order:
      type: object
      required: [id, status, total]
      properties:
        id: { type: string, format: uuid }
        status: { type: string, enum: [open, paid, shipped, cancelled] }
        total: { type: number, minimum: 0 }
```
Lint with **Spectral** in CI (style-guide rules), generate mocks (Prism), docs (Redocly, Stoplight), and client/server stubs (openapi-generator). Most gateways import OpenAPI directly to provision routes and policies.

---

## 4. GraphQL design

GraphQL (Facebook, open-sourced 2015) exposes a single endpoint and a typed **schema (SDL)**; the client asks for exactly the fields it needs, resolved (often in parallel) from many sources. Designed to give consumers control and end over-/under-fetching.

```graphql
type Query {
  order(id: ID!): Order
  orders(first: Int, after: String, status: OrderStatus): OrderConnection!
}
type Mutation {
  cancelOrder(id: ID!): Order
}
type Order {
  id: ID!
  status: OrderStatus!
  total: Float!
  customer: Customer!          # resolved from a different service
  oldField: String @deprecated(reason: "Use total")
}
enum OrderStatus { OPEN PAID SHIPPED CANCELLED }
```
Design rules:
- **Don't version.** Add fields and mark removals with `@deprecated`; clients only break if a field they request disappears.
- Use the **Relay connection** pattern (`edges`/`node`/`pageInfo`, cursors) for pagination.
- Guard against abusive queries: **query depth limiting, complexity/cost analysis, persisted queries, and timeouts** — a single deep query can be a DoS vector.
- AuthZ usually lives in resolvers/field-level rules, not the gateway. Plan for it explicitly.
- Subscriptions (over WebSocket) cover real-time needs.

---

## 5. gRPC / Protocol Buffers

gRPC (Google, open-sourced 2015, CNCF) is a modern RPC framework over HTTP/2: binary protobuf payloads, header compression, multiplexing, bidirectional streaming, flow control, a strongly-typed IDL, and a single compiler (`protoc`) that generates clients and servers in many languages.

```protobuf
syntax = "proto3";
package orders.v1;

service OrderService {
  rpc GetOrder(GetOrderRequest) returns (Order);
  rpc WatchOrders(WatchRequest) returns (stream Order);   // server streaming
}
message GetOrderRequest { string id = 1; }
message Order {
  string id = 1;
  Status status = 2;
  double total = 3;
  enum Status { OPEN = 0; PAID = 1; SHIPPED = 2; CANCELLED = 3; }
}
```
Compatibility: never reuse or renumber a field tag; add new fields with new numbers; only remove fields you've reserved. This gives backward compatibility by design. Poor browser support (use grpc-web + a proxy, or expose REST/GraphQL at the edge and gRPC internally).

---

## 6. AsyncAPI (event-driven)

For message/event APIs (Kafka, AMQP, MQTT, WebSocket, SSE) use **AsyncAPI** — the OpenAPI analogue for async. It describes channels, messages, and operations (publish/subscribe) and supports the same design-first/mock/codegen tooling.

```yaml
asyncapi: 3.0.0
info: { title: Orders Events, version: 1.0.0 }
channels:
  orderCancelled:
    address: orders.cancelled
    messages:
      OrderCancelled:
        payload:
          type: object
          properties:
            orderId: { type: string }
            cancelledAt: { type: string, format: date-time }
operations:
  onOrderCancelled:
    action: receive
    channel: { $ref: "#/channels/orderCancelled" }
```
Pair async events with the synchronous API (e.g. a CQRS write side emitting events; a webhook delivering them) — see the gateways and lifecycle references.

---

## 7. Cross-cutting design rules

- **Consistency over cleverness.** One naming convention (snake_case or camelCase — pick one), one error shape, one pagination style across the whole platform. Enforce with a style guide + linter.
- **Backward compatibility is sacred** within a major version: additive changes only (new optional fields, new endpoints). Removing/renaming/retyping fields or making optional fields required is breaking.
- **Design from the consumer in**, not the database out. The API contract should reflect the consumer's mental model, not your schema.
- **Mock before you build** so consumers give feedback while change is cheap (design-first; see lifecycle reference).
