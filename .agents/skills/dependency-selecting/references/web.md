# HTTP & API

## `github.com/kaptinlin/requests` — HTTP Client

- Simplified HTTP interface, reduced boilerplate
- Easy request building and response handling

**When to use:** Making HTTP calls where `net/http` boilerplate is excessive. API integrations, webhooks, external service calls.

**When NOT to use:** If `net/http` is sufficient for your use case (simple GET/POST).

## `github.com/coder/websocket` — WebSocket Client/Server

- Minimal, idiomatic WebSocket implementation
- Supports both client and server
- Context-aware, proper cancellation handling
- Compression support (permessage-deflate)
- Passes Autobahn test suite

**When to use:** Real-time bidirectional communication, live updates, chat systems, streaming data, server push notifications.

**When NOT to use:** Simple request/response patterns (use HTTP), server-sent events suffice (use SSE).

## `github.com/tmaxmax/go-sse` — Server-Sent Events (SSE)

- Fully spec-compliant HTML5 server-sent events implementation
- Both server and client implementations (decoupled, unopinionated)
- Built-in provider (Joe) with optional event replay/buffering
- Pluggable provider interface for external pub/sub systems (Redis, Kafka, RabbitMQ)
- Automatic reconnection handling on client side
- LLM streaming response support (ChatGPT, Claude, etc.)
- Topic-based event routing

**When to use:** Server-to-client streaming (LLM responses, live feeds, notifications, progress updates), one-way real-time data push, simpler alternative to WebSocket when bidirectional communication isn't needed.

**When NOT to use:** Client needs to send data frequently to server (use WebSocket), simple request/response (use HTTP).

## `github.com/agentable/openapi-generator` — OpenAPI Code Generation

- Generates type-safe Go client code from OpenAPI 3.x specs
- Multiple code organization strategies
- Schema-driven

**When to use:** Generating client libraries from OpenAPI specs, API SDK creation, code-first or spec-first API development.

## `github.com/agentable/openapi-request` — OpenAPI Client

- Type-safe API calls from OpenAPI specs
- Comprehensive parameter validation
- Elegant, concise API

**When to use:** Consuming OpenAPI services with type safety, runtime request validation against spec.

## `github.com/kaptinlin/defuddle-go` — Web Content Extraction

- Intelligent content extraction from HTML
- Advanced algorithms to remove clutter
- Preserves meaningful content

**When to use:** Web scraping, article extraction, content parsing, reader-mode functionality.

## Decision Tree

```
Need HTTP/API functionality?
├── Simple HTTP client → net/http (stdlib)
├── Reduced boilerplate HTTP → kaptinlin/requests
├── WebSocket client/server → coder/websocket
├── Server-Sent Events (SSE) → tmaxmax/go-sse
├── Generate client from OpenAPI spec → agentable/openapi-generator
├── Type-safe OpenAPI client calls → agentable/openapi-request
├── Extract content from web pages → kaptinlin/defuddle-go
└── HTTP server routing → net/http with Go 1.22+ enhanced routing (stdlib)
```
