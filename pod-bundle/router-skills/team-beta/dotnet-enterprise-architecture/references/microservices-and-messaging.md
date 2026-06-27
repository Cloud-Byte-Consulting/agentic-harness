# Microservices, messaging, and event-driven integration

Contents:
- Finding the right service granularity
- Synchronous vs asynchronous integration
- Messaging with RabbitMQ / AMQP
- MassTransit on .NET
- Webhooks for low-coupling notifications
- The outbox pattern
- Eventual consistency and idempotency
- Orchestration vs choreography
- A note on the history of integration

## Finding the right service granularity

The hard problem in microservices is not the technology — it is deciding *where the boundaries go*. DDD bounded contexts are the answer: one context → one service → one data store → one contract. Each major entity with its own life cycle (Book, Author) is a candidate boundary; minor entities live inside an aggregate.

Signs a boundary is wrong:
- Two services must be deployed together to ship a feature → they share a context; merge them.
- A "service" has almost no logic and just proxies another → grain too fine.
- Frequent chatty synchronous calls between two services on the hot path → likely one context split prematurely.

Default to a modular monolith; extract a service when a concrete force (independent scaling, team autonomy, release cadence) demands it.

## Synchronous vs asynchronous integration

- **Synchronous (REST/gRPC)** — request/response, immediate result, simple to reason about. Use for *queries* where the caller needs an answer now. Couples availability: if the callee is down, the caller's request fails (mitigate with resilience — see `cross-cutting-concerns.md`).
- **Asynchronous (messaging/events)** — "something happened" or "please do this eventually". Use to **decouple producers from consumers**: the producer doesn't know or wait for consumers, consumers can be added/removed freely, and a broker absorbs load spikes and outages. The cost is eventual consistency.

Rule of thumb: read across a boundary → sync call behind an interface/config URL; react to a state change → publish an event.

Decouple sync calls via **dependency inversion**: the caller depends on a configured URL/interface, not on the concrete callee. The callee implements the agreed contract (OpenAPI/pivotal format). Neither hard-codes the other.

## Messaging with RabbitMQ / AMQP

RabbitMQ (AMQP 0-9-1) is a robust, approachable broker with first-class .NET support. Core concepts:

- **Producer** publishes to an **exchange** (not directly to a queue).
- The exchange routes to **queues** by binding rules (direct, topic, fanout, headers).
- **Consumers** read from queues; messages are acknowledged (`ack`) after successful processing, or `nack`/requeued/dead-lettered on failure.

```csharp
// Producer (RabbitMQ.Client v7+ is async)
var factory = new ConnectionFactory { HostName = "localhost" };
await using var conn = await factory.CreateConnectionAsync();
await using var channel = await conn.CreateChannelAsync();
await channel.ExchangeDeclareAsync("catalog.events", ExchangeType.Topic, durable: true);

var body = JsonSerializer.SerializeToUtf8Bytes(new BookReadyToPrint(bookId, DateTimeOffset.UtcNow));
var props = new BasicProperties { Persistent = true, MessageId = Guid.NewGuid().ToString() };
await channel.BasicPublishAsync(
    exchange: "catalog.events", routingKey: "book.readytoprint",
    mandatory: false, basicProperties: props, body: body);
```

Operational essentials: declare exchanges/queues **durable** and publish **persistent** messages so they survive broker restarts; configure a **dead-letter exchange** for messages that repeatedly fail; set prefetch to control consumer concurrency.

Kafka is the alternative when you need a high-throughput, replayable event *log* (event sourcing, stream processing, big-data pipelines) rather than per-message work queues. Choose Kafka for log/replay semantics, RabbitMQ for task queues and routing — don't reach for either when a single well-designed service with eventual consistency would do.

## MassTransit on .NET

MassTransit is a higher-level abstraction over RabbitMQ/Azure Service Bus/Amazon SQS that handles serialization, retries, dead-lettering, the outbox, and saga orchestration. Prefer it over raw client code for non-trivial systems.

```csharp
// Program.cs
builder.Services.AddMassTransit(x =>
{
    x.AddConsumer<BookReadyToPrintConsumer>();
    x.UsingRabbitMq((ctx, cfg) =>
    {
        cfg.Host("localhost", "/", h => { h.Username("guest"); h.Password("guest"); });
        cfg.ReceiveEndpoint("printing", e =>
        {
            e.UseMessageRetry(r => r.Exponential(5,
                TimeSpan.FromSeconds(1), TimeSpan.FromSeconds(30), TimeSpan.FromSeconds(2)));
            e.ConfigureConsumer<BookReadyToPrintConsumer>(ctx);
        });
    });
});

public sealed class BookReadyToPrintConsumer(IPrinter printer) : IConsumer<BookReadyToPrint>
{
    public Task Consume(ConsumeContext<BookReadyToPrint> context) =>
        printer.QueueAsync(context.Message.BookId);
}

// Publishing
await publishEndpoint.Publish(new BookReadyToPrint(bookId, DateTimeOffset.UtcNow));
```

MassTransit's **saga** state machines implement long-running orchestrations with persisted state.

## Webhooks for low-coupling notifications

When you control both ends but want minimal coupling (and no shared broker), webhooks are an elegant fit: a consumer **subscribes** by registering a callback URL (+ an optional filter); the producer calls that URL when the event occurs. The producer never depends on the consumer — it just invokes a URL it was given (dependency inversion).

```csharp
// Producer exposes a subscription endpoint
app.MapPost("/books/subscribe", (string callbackUrl, string? filter, ISubscriptions subs) =>
{
    subs.Add(new Subscription(new Uri(callbackUrl), filter));
    return Results.Accepted();
});

// On a state change, the producer notifies each matching subscriber, with retries
foreach (var sub in subs.Matching(evt))
    await callbackClient.PutAsJsonAsync(sub.CallbackUrl, evt, ct); // PUT => idempotent
```

Design notes:
- Use an **idempotent verb** (PUT) so retries are safe.
- Apply a **retry policy** on the outbound HTTP client (Polly — see `cross-cutting-concerns.md`).
- Decide whether to send the data in the callback body or just an id the consumer then GETs (sending an id simplifies authorization — the consumer's own rights are checked on the GET; sending the body saves a round-trip). Pick per case.
- Persist subscriptions (an in-memory list is lost on restart). For critical delivery, back webhooks with a broker rather than fire-and-forget HTTP.

## The outbox pattern

The classic distributed bug: you commit a DB change *and* publish a message, but the process crashes between the two — now the state and the event disagree. The **transactional outbox** fixes this: within the *same database transaction* that saves the state change, write the outgoing message to an `outbox` table. A separate dispatcher polls the outbox and publishes to the broker, marking rows sent. The state change and the intent-to-publish commit atomically; the dispatcher guarantees at-least-once delivery (hence consumers must be idempotent).

MassTransit and EF Core both have built-in outbox support — prefer it over hand-rolling.

## Eventual consistency and idempotency

Asynchronous integration means data across services is **eventually** consistent, not immediately. Embrace it:

- **Idempotent consumers.** At-least-once delivery means duplicates happen. Make handlers idempotent: dedupe on a message id, use upserts, or use idempotency keys. PUT/upsert semantics help.
- **Periodic reconciliation.** Webhooks/messages can be missed during micro-outages. A low-frequency batch sync (e.g. nightly) that re-pulls and reconciles closes the gaps without a full broker for low-criticality streams.
- **Compensation over distributed transactions.** Avoid two-phase commit across services. Model multi-step workflows as sagas with compensating actions for rollback.
- **Tell the user the truth.** A "new" flag or freshly written value may lag; show it as provisional, or read the acting user's own write from the source of truth immediately.

## Orchestration vs choreography

Two ways to run a multi-step business process:

- **Orchestration** — one coordinator tells each participant what to do and in what order, reading results to decide next steps. Easier to see the whole flow and to handle long-running/human steps; the coordinator is a (manageable) central point. Implement with a saga (MassTransit), a workflow engine, or — for simple cases — a lightweight tool like a workflow automation service.
- **Choreography** — no coordinator; each service reacts to events and emits its own, the flow emerging from the chain. Lowest coupling and most scalable, but the end-to-end flow is implicit and harder to trace/debug.

Choose orchestration when you need visibility, ordering, or human-in-the-loop steps; choreography when you want maximum decoupling and the steps are genuinely independent reactions. Many systems mix both.

Important: you do **not** need a heavy BPMN engine to run processes. Most processes are carried by the sequence of GUI screens plus a well-designed status field on the aggregate and authorization rules; reserve a real engine for processes that are complex, change often, or are long-running across days/months. A status attribute + domain events + webhooks goes a very long way.

## A note on the history of integration

Reusability/interoperability evolved through libraries → COM/DCOM/CORBA → SOAP/WSDL and the WS-* stack → ESB/MOM → REST. REST won because it adds nothing new: it reuses HTTP verbs, status codes, URLs, and JSON, so there's no proprietary layer to maintain. The remaining hard problem is *semantic/functional* interoperability — agreeing on what the data *means* — which is why standards and pivotal formats (see `api-and-integration.md`) matter more than the transport.
