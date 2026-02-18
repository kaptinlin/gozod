# HTTP & API

## `github.com/kaptinlin/requests` — HTTP Client

- Simplified HTTP interface, reduced boilerplate
- Easy request building and response handling

**When to use:** Making HTTP calls where `net/http` boilerplate is excessive. API integrations, webhooks, external service calls.

**When NOT to use:** If `net/http` is sufficient for your use case (simple GET/POST).

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
├── Generate client from OpenAPI spec → agentable/openapi-generator
├── Type-safe OpenAPI client calls → agentable/openapi-request
├── Extract content from web pages → kaptinlin/defuddle-go
└── HTTP server routing → net/http with Go 1.22+ enhanced routing (stdlib)
```
