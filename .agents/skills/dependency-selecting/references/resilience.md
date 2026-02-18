# Resilience & Fault Tolerance

## KISS Rule

- Use one resilience library by default: `github.com/failsafe-go/failsafe-go`.
- Do not mix retry libraries in the same project unless there is a measured reason.
- For simple reconnect/poll loops inside one package, prefer a project-local backoff helper (`internal/backoff`) instead of adding a new dependency.

## `github.com/failsafe-go/failsafe-go` — Default Resilience Library

Composable resilience policies for production services.

**Policies:**
- **Failure handling:** Retry, Fallback
- **Load limiting:** Circuit Breaker, Adaptive Limiter, Adaptive Throttler, Bulkhead, Rate Limiter, Cache
- **Time limiting:** Timeout, Hedge

**Key features:**
- Policy composition (layer multiple policies; order matters)
- Async execution with `ExecutionResult`
- Type-safe generics
- HTTP and gRPC integration

**When to use:** Retry with backoff, or any composition need (retry + timeout/circuit breaker/rate limiter), unreliable external services, gRPC/HTTP resilience.

```go
retryPolicy := retrypolicy.Builder[*http.Response]().
    HandleIf(func(resp *http.Response, err error) bool {
        return resp != nil && resp.StatusCode == 429
    }).
    WithBackoff(time.Second, 30*time.Second).
    Build()

cb := circuitbreaker.Builder[*http.Response]().
    WithFailureThreshold(5).
    WithDelay(30 * time.Second).
    Build()

resp, err := failsafe.Get(func() (*http.Response, error) {
    return http.Get(url)
}, retryPolicy, cb)
```

## Decision Tree

```
Need resilience?
├── Package-local reconnect/poll loop?
│   └── internal/backoff helper (project-local)
├── Retry for service calls or task execution?
│   └── failsafe-go
├── Circuit breaker, rate limiter, bulkhead, hedge, or composition?
│   └── failsafe-go
└── Just a timeout?
    └── context.WithTimeout (stdlib)
```

## Do NOT Use

| Library | Reason |
|---------|--------|
| `sethvargo/go-retry` | Avoid mixed retry abstractions unless benchmark-driven |
| `cenkalti/backoff` | Avoid mixed retry abstractions unless benchmark-driven |
| `sony/gobreaker` | failsafe-go covers circuit breaker + more |
| `afex/hystrix-go` | Abandoned Netflix port |
| Manual retry loops | Use a library for jitter, backoff, context |
