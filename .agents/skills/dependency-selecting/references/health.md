# Health Checks

Health check libraries for monitoring application and dependency status.

## Recommended Libraries

### agentable/go-health

**Module:** `github.com/agentable/go-health`

**Use for:**
- Application health monitoring
- Dependency health checks (HTTP, TCP, DNS, Redis, PostgreSQL, gRPC)
- Framework-agnostic health endpoints
- Concurrent health check execution
- Observability via listeners

**Key features:**
- Built-in checks for common dependencies (HTTP, TCP, DNS, Redis, PostgreSQL, gRPC)
- Concurrent execution with configurable intervals and timeouts
- Framework integrations (net/http, Fiber, Kratos)
- Listener pattern for metrics and logging
- Four health states: healthy, unhealthy, degraded, unknown
- Go 1.23+ iterators for result traversal

**Example:**

```go
import (
    "context"
    "time"

    "github.com/agentable/go-health"
    "github.com/agentable/go-health/checks"
    healthhttp "github.com/agentable/go-health/http"
)

// Create health manager
h := health.New()

// Register HTTP endpoint check
h.Register(
    checks.NewHTTP("api", "https://api.example.com/health"),
    health.WithInterval(30*time.Second),
    health.WithTimeout(5*time.Second),
)

// Register Redis check
h.Register(
    checks.NewRedis("cache", redisClient),
    health.WithInterval(10*time.Second),
)

// Register PostgreSQL check
h.Register(
    checks.NewPostgres("database", db),
    health.WithInterval(15*time.Second),
)

// Expose health endpoint (net/http)
http.Handle("/health", healthhttp.Handler(h))

// Check overall health
if h.IsHealthy() {
    // All checks passing
}

// Get individual results
for name, result := range h.Results() {
    fmt.Printf("%s: %s\n", name, result.Status)
}

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
h.Shutdown(ctx)
```

**Custom checks:**

```go
type CustomCheck struct {
    name string
}

func (c *CustomCheck) Name() string {
    return c.name
}

func (c *CustomCheck) Execute(ctx context.Context) health.Result {
    // Perform health check logic
    if err := checkSomething(ctx); err != nil {
        return health.NewUnhealthyResult(err.Error(), err)
    }
    return health.NewHealthyResult("service is operational")
}

// Register custom check
h.Register(&CustomCheck{name: "custom-service"})
```

**Framework integrations:**

```go
// Fiber
import "github.com/agentable/go-health/fiber"
app.Get("/health", fiber.Handler(h))

// Kratos
import "github.com/agentable/go-health/kratos"
kratos.RegisterHealth(srv, h)
```

**Observability:**

```go
type MetricsListener struct{}

func (l *MetricsListener) OnCheckStarted(ctx context.Context, name string) {
    // Record check start time
}

func (l *MetricsListener) OnCheckCompleted(ctx context.Context, name string, result health.Result) {
    // Record metrics (duration, status, etc.)
    metrics.RecordCheckDuration(name, result.Duration)
    metrics.RecordCheckStatus(name, string(result.Status))
}

func (l *MetricsListener) OnCheckRegistered(ctx context.Context, name string) {}
func (l *MetricsListener) OnCheckDeregistered(ctx context.Context, name string) {}

// Register listener
h := health.New(health.WithCheckListener(&MetricsListener{}))
```

**When to use:**
- ✅ Kubernetes liveness/readiness probes
- ✅ Load balancer health checks
- ✅ Monitoring dependency availability
- ✅ Service mesh health reporting
- ✅ Multi-framework applications (framework-agnostic core)

**When NOT to use:**
- ❌ Simple single-dependency check (use direct ping/connection test)
- ❌ Application metrics collection (use Prometheus client instead)
- ❌ Distributed tracing (use OpenTelemetry instead)

## Design Patterns

**Functional options:**
```go
health.WithInterval(30*time.Second)
health.WithTimeout(5*time.Second)
health.WithCheckListener(listener)
health.WithHealthListener(listener)
```

**Interface segregation:**
- `Task` interface (2 methods) for custom checks
- `CheckListener` interface (4 methods) for check observability
- `HealthListener` interface (1 method) for overall health changes
- `RedisClient` interface (1 method) for Redis compatibility
- `PostgresDB` interface (1 method) for PostgreSQL compatibility

**Concurrent execution:**
- Each check runs in its own goroutine
- Configurable intervals and timeouts per check
- Thread-safe result access via mutex
- Graceful shutdown with context cancellation

## Alternatives

None recommended. Use stdlib direct connection testing for simple single-dependency checks.
