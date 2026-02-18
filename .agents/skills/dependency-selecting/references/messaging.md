# Messaging & Events

## `github.com/kaptinlin/emitter` — In-Process Event Emitter

- High-performance, thread-safe
- Modern Go 1.26+ with atomic operations
- Optimized data structures

**When to use:** Decoupling components within a single process, observer pattern, plugin/hook systems.

**When NOT to use:** Cross-process or cross-service communication (use watermill).

## `github.com/kaptinlin/queue` — Background Job Queue

- Built on Asynq with Redis storage
- Automatic error logging, custom error handling
- Retry with configurable policies
- Priority queues, distributed job processing

**When to use:** Async tasks (email, reports), scheduled/delayed jobs, distributed workers, job persistence.

## `github.com/ThreeDotsLabs/watermill` — Distributed Messaging

- Simple handler: `func(*Message) ([]*Message, error)`
- 12+ pub/sub backends: Kafka, RabbitMQ, NATS, Redis Streams, AWS SNS/SQS, Google Cloud Pub/Sub, PostgreSQL, MySQL, HTTP
- Middleware for cross-cutting concerns
- Performance: GoChannel ~315K msg/s; Kafka ~41K pub / ~101K sub msg/s
- Stress-tested with race detection

**When to use:** Event-driven microservices, CQRS/event sourcing, cross-service messaging, saga/orchestration. Swap transport layers without code changes.

## Decision Tree

```
Need event/message handling?
├── Within a single process?
│   ├── Event bus / observer pattern → kaptinlin/emitter
│   └── Background jobs with persistence → kaptinlin/queue
└── Across services / distributed?
    └── watermill (choose appropriate pub/sub backend)
```

## Do NOT Use

| Library | Reason |
|---------|--------|
| Raw `chan` for event bus | Not flexible enough for real event systems |
| `hibiken/asynq` directly | Use `kaptinlin/queue` for better ergonomics |
