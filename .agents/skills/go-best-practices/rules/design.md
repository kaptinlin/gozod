# Design Patterns

API design decisions around interfaces, option patterns, global state, and type choices determine long-term maintainability and extensibility.

## Contents
- Interfaces Belong to Consumers
- Option Structs
- Variadic Options
- Avoid Global State
- Context Conventions
- Pass Values, Not Unnecessary Pointers
- Choosing Receiver Types
- Use Generics Judiciously

---

## Interfaces Belong to Consumers

Define interfaces in the package that **uses** them, not the package that implements them. Implementation packages return concrete types.

**Incorrect (interface defined in producer):**

```go
package producer

type Thinger interface { Thing() bool }
type defaultThinger struct{ ... }
func NewThinger() Thinger { return defaultThinger{ ... } }
```

**Correct (interface in consumer, concrete in producer):**

```go
package consumer
type Thinger interface { Thing() bool }
func Foo(t Thinger) string { ... }

package producer
type Thinger struct{ ... }
func (t Thinger) Thing() bool { ... }
func NewThinger() Thinger { return Thinger{ ... } }
```

Don't define interfaces before you have a real use case. Don't export interfaces users don't need. Don't use interface parameters when only one type will ever be passed.

---

## Option Structs

When functions have many parameters, use an option struct as the last argument. Self-documenting, extensible, and allows omitting default fields.

**Incorrect:**

```go
func EnableReplication(ctx context.Context, config *replicator.Config, primaryRegions, readonlyRegions []string, replicateExisting, overwritePolicies bool, replicationInterval time.Duration, copyWorkers int, healthWatcher health.Watcher) {
}
```

**Correct:**

```go
type ReplicationOptions struct {
    Config              *replicator.Config
    PrimaryRegions      []string
    ReadonlyRegions     []string
    ReplicateExisting   bool
    OverwritePolicies   bool
    ReplicationInterval time.Duration
    CopyWorkers         int
    HealthWatcher       health.Watcher
}

func EnableReplication(ctx context.Context, opts ReplicationOptions) {
}
```

Prefer option structs when: all callers need at least one option, many callers need many options, or options are shared across multiple functions. Context should never be in the option struct.

---

## Variadic Options

When most callers don't need configuration, use the variadic option pattern (functional options).

```go
type replicationOptions struct {
    readonlyCells     []string
    replicateExisting bool
    copyWorkers       int
}

type ReplicationOption func(*replicationOptions)

func ReadonlyCells(cells ...string) ReplicationOption {
    return func(opts *replicationOptions) {
        opts.readonlyCells = append(opts.readonlyCells, cells...)
    }
}

func EnableReplication(ctx context.Context, config *placer.Config, primaryCells []string, opts ...ReplicationOption) {
    var options replicationOptions
    for _, opt := range opts {
        opt(&options)
    }
}
```

Binary settings should accept `bool` (`FailFast(enable bool)` over `EnableFailFast()`). Enum settings should accept enum constants (`log.Format(log.Capacitor)` over `log.CapacitorFormat()`).

---

## Avoid Global State

Libraries should not expose APIs depending on global state. Allow clients to create and use isolated instances.

**Incorrect:**

```go
package useradmin

var client pb.UserAdminServiceClientInterface

func Client() *pb.UserAdminServiceClient {
    if client == nil {
        client = ... // setup
    }
    return client
}
```

**Correct:**

```go
package sidecar

type Registry struct { plugins map[string]*Plugin }
func New() *Registry { return &Registry{plugins: make(map[string]*Plugin)} }
func (r *Registry) Register(name string, p *Plugin) error { ... }
```

Problematic patterns: top-level variables, global registries, global callback registration, thick-client singletons. Global state makes testing, parallelism, and refactoring difficult.

---

## Context Conventions

`context.Context` is always the **first parameter**. Never put it in structs. Never create custom context types.

**Incorrect:**

```go
type Server struct {
    ctx context.Context
    db  *sql.DB
}
```

**Correct:**

```go
func (s *Server) HandleRequest(ctx context.Context, req *Request) error {
    // ...
}
```

Exceptions:
- HTTP handlers: use `req.Context()`
- Streaming RPC: context comes from the stream
- Entry points (`main`, `init`, `TestXXX`): use `context.Background()`

Context is immutable — the same context can be passed to multiple calls sharing the same deadline, cancellation, and credentials.

---

## Pass Values, Not Unnecessary Pointers

Don't pass pointers just to save bytes. If a function only reads `*x`, pass `x` directly. Applies to `string`, `io.Reader`, and other fixed-size values.

**Incorrect:**

```go
func process(s *string) { fmt.Println(*s) }
func handle(r *io.Reader) { io.Copy(os.Stdout, *r) }
```

**Correct:**

```go
func process(s string) { fmt.Println(s) }
func handle(r io.Reader) { io.Copy(os.Stdout, r) }
```

Does not apply to large structs or structs that may grow. Protocol buffer messages should generally be passed by pointer.

---

## Choosing Receiver Types

Choose receiver type based on correctness, then consistency. Prefer pointer receivers when in doubt.

Use **pointer receiver** when: method mutates the receiver, receiver contains non-copyable fields (`sync.Mutex`), or receiver is a large struct.

Use **value receiver** when: receiver is a small immutable struct, a builtin type, a map/function/channel, or a slice where the method doesn't reslice.

**Incorrect (mixed receiver types):**

```go
func (c Config) Get() string { return c.name }
func (c *Config) Set(name string) { c.name = name }
```

**Correct (consistent pointer receivers):**

```go
func (c *Config) Get() string { return c.name }
func (c *Config) Set(name string) { c.name = name }
```

---

## Use Generics Judiciously

Don't use generics just because an algorithm is type-agnostic. If only one type is instantiated, write concrete code first.

**Incorrect (premature generics — only ever called with `[]User`):**

```go
func Filter[T any](items []T, fn func(T) bool) []T {
    var result []T
    for _, item := range items {
        if fn(item) { result = append(result, item) }
    }
    return result
}
```

**Correct (concrete until proven generic):**

```go
func FilterUsers(users []User, fn func(User) bool) []User {
    var result []User
    for _, u := range users {
        if fn(u) { result = append(result, u) }
    }
    return result
}
```

Don't use generics to build DSLs. If types share a useful interface, consider modeling with that interface instead.
