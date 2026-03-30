# API Design Patterns

This document provides a complete catalog of API design patterns for Golang libraries with real-world examples.

## Table of Contents

1. [Functional Options Pattern](#functional-options-pattern)
2. [Builder Pattern](#builder-pattern)
3. [Interface Segregation](#interface-segregation)
4. [Two-Phase Construction](#two-phase-construction)
5. [Fluent API](#fluent-api)
6. [Config Struct Pattern](#config-struct-pattern)
7. [Optional Capabilities](#optional-capabilities)
8. [Generic Constraints](#generic-constraints)

---

## Functional Options Pattern

**When to use:** Optional configuration, backward compatibility, sensible defaults

**Characteristics:**
- Zero-value struct with sensible defaults
- Each option is a function that modifies config
- Variadic options parameter
- Backward compatible (new options don't break existing code)

### Basic Pattern

```go
type Config struct {
    Timeout    time.Duration
    MaxRetries int
    UserAgent  string
}

type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.Timeout = d
    }
}

func WithMaxRetries(n int) Option {
    return func(c *Config) {
        c.MaxRetries = n
    }
}

func New(opts ...Option) *Client {
    cfg := &Config{
        Timeout:    30 * time.Second,  // Default
        MaxRetries: 3,                 // Default
        UserAgent:  "MyClient/1.0",    // Default
    }

    for _, opt := range opts {
        opt(cfg)
    }

    return &Client{config: cfg}
}
```

**Usage:**
```go
// Use defaults
client := New()

// Override specific options
client := New(
    WithTimeout(10 * time.Second),
    WithMaxRetries(5),
)
```

### Advanced: Options with Validation

```go
type Option func(*Config) error

func WithTimeout(d time.Duration) Option {
    return func(c *Config) error {
        if d <= 0 {
            return fmt.Errorf("timeout must be positive")
        }
        c.Timeout = d
        return nil
    }
}

func New(opts ...Option) (*Client, error) {
    cfg := &Config{Timeout: 30 * time.Second}

    for _, opt := range opts {
        if err := opt(cfg); err != nil {
            return nil, err
        }
    }

    return &Client{config: cfg}, nil
}
```

### Real-World Example: go-cache

```go
type Config struct {
    Store        Store
    Codec        Codec
    DefaultTTL   time.Duration
    SingleFlight bool
    Metrics      bool
}

type Option func(*Config)

func WithTTL(ttl time.Duration) Option {
    return func(c *Config) { c.DefaultTTL = ttl }
}

func WithSingleFlight() Option {
    return func(c *Config) { c.SingleFlight = true }
}

func WithMetrics() Option {
    return func(c *Config) { c.Metrics = true }
}

func New[T any](store Store, opts ...Option) Cache[T] {
    cfg := &Config{
        Store:      store,
        DefaultTTL: 5 * time.Minute,
    }

    for _, opt := range opts {
        opt(cfg)
    }

    return &cache[T]{config: cfg}
}
```

---

## Builder Pattern

**When to use:** Complex object construction, fluent API, method chaining, validation at build time

**Characteristics:**
- Mutable builder phase
- Immutable result after Build()
- Method chaining for fluent API
- Validation in Build() method

### Basic Pattern

```go
type Builder struct {
    name    string
    timeout time.Duration
    retries int
    headers map[string]string
}

func NewBuilder() *Builder {
    return &Builder{
        timeout: 30 * time.Second,
        retries: 3,
        headers: make(map[string]string),
    }
}

func (b *Builder) Name(name string) *Builder {
    b.name = name
    return b
}

func (b *Builder) Timeout(d time.Duration) *Builder {
    b.timeout = d
    return b
}

func (b *Builder) Header(key, value string) *Builder {
    b.headers[key] = value
    return b
}

func (b *Builder) Build() (*Client, error) {
    if b.name == "" {
        return nil, errors.New("name is required")
    }

    return &Client{
        name:    b.name,
        timeout: b.timeout,
        retries: b.retries,
        headers: b.headers,
    }, nil
}
```

**Usage:**
```go
client, err := NewBuilder().
    Name("my-client").
    Timeout(10 * time.Second).
    Header("User-Agent", "MyApp/1.0").
    Build()
```

### Real-World Example: go-fsm

```go
type Builder[S, E comparable] struct {
    initial    S
    rules      []rule[S, E]
    entryHooks map[S][]ActionFunc[S, E]
    exitHooks  map[S][]ActionFunc[S, E]
    finals     map[S]bool
}

func New[S, E comparable](initial S) *Builder[S, E] {
    return &Builder[S, E]{
        initial:    initial,
        entryHooks: make(map[S][]ActionFunc[S, E]),
        exitHooks:  make(map[S][]ActionFunc[S, E]),
        finals:     make(map[S]bool),
    }
}

func (b *Builder[S, E]) From(states ...S) *StateGroup[S, E] {
    return &StateGroup[S, E]{builder: b, states: states}
}

func (b *Builder[S, E]) Final(states ...S) *Builder[S, E] {
    for _, s := range states {
        b.finals[s] = true
    }
    return b
}

func (b *Builder[S, E]) Build() (*Machine[S, E], error) {
    // Validation
    if len(b.rules) == 0 {
        return nil, errors.New("no rules defined")
    }

    // Check final states have no outgoing edges
    for _, r := range b.rules {
        if b.finals[r.from] {
            return nil, fmt.Errorf("final state %v has outgoing edge", r.from)
        }
    }

    // Compile to immutable runtime structure
    return &Machine[S, E]{
        table: compileTable(b.rules),
        // ...
    }, nil
}
```

**Usage:**
```go
m, err := fsm.New[State, Event](Unpaid).
    From(Unpaid).
        On(PaySuccess).To(Paid).Action(publishPaid).
        On(Cancel).To(Cancelled).
    From(Paid).
        On(Ship).To(Shipped).
    Final(Completed, Cancelled).
    Build()
```

---

## Interface Segregation

**When to use:** Pluggable components, single-method interfaces, optional capabilities

**Characteristics:**
- Small, focused interfaces (1-3 methods)
- Optional capabilities via type assertion
- No interface pollution
- Easy to mock and test

### Basic Pattern

```go
// Core interface - single method
type Store interface {
    Get(ctx context.Context, key string) ([]byte, error)
}

// Optional capabilities - checked via type assertion
type Setter interface {
    Set(ctx context.Context, key string, value []byte) error
}

type Deleter interface {
    Delete(ctx context.Context, key string) error
}

type BatchGetter interface {
    GetMany(ctx context.Context, keys []string) (map[string][]byte, error)
}

// Usage
func process(store Store) {
    // Core functionality
    data, _ := store.Get(ctx, "key")

    // Optional capability
    if setter, ok := store.(Setter); ok {
        setter.Set(ctx, "key", data)
    }

    // Optional batch capability
    if batch, ok := store.(BatchGetter); ok {
        results := batch.GetMany(ctx, keys)
    }
}
```

### Real-World Example: go-config

```go
// Core interface - single method
type Source interface {
    Load(ctx context.Context) (map[string]any, error)
}

// Optional capabilities
type Watcher interface {
    Watch(ctx context.Context, onChange func(map[string]any)) error
}

type StatusReporter interface {
    Status(fn func(changed bool, err error))
}

type SkipReporter interface {
    SkipError() error
}

// File provider implements all capabilities
type File struct {
    path   string
    policy Policy
    skipErr error
}

func (f *File) Load(ctx context.Context) (map[string]any, error) { ... }
func (f *File) Watch(ctx context.Context, onChange func(map[string]any)) error { ... }
func (f *File) SkipError() error { return f.skipErr }

// FS provider implements only Load and SkipReporter (no Watch)
type Provider struct {
    fsys    fs.FS
    path    string
    skipErr error
}

func (p *Provider) Load(ctx context.Context) (map[string]any, error) { ... }
func (p *Provider) SkipError() error { return p.skipErr }

// Usage in Config
func (c *Config[T]) Watch(ctx context.Context) error {
    for _, source := range c.sources {
        if w, ok := source.(Watcher); ok {
            go w.Watch(ctx, c.reload)
        }
    }
}
```

---

## Two-Phase Construction

**When to use:** Validation at build time, immutable runtime, zero-allocation hot paths

**Characteristics:**
- Mutable build phase with validation
- Immutable runtime phase
- Pre-allocated data structures
- Concurrency-safe by design

### Pattern

```go
// Build phase - mutable
type builder struct {
    rules      []rule
    validators []validator
}

func (b *builder) AddRule(r rule) {
    b.rules = append(b.rules, r)
}

func (b *builder) Build() (*Runtime, error) {
    // Validation
    if len(b.rules) == 0 {
        return nil, errors.New("no rules")
    }

    for _, r := range b.rules {
        if err := r.Validate(); err != nil {
            return nil, err
        }
    }

    // Compile to immutable structure
    table := compileTable(b.rules)

    return &Runtime{
        table: table,  // Pre-allocated, immutable
    }, nil
}

// Runtime phase - immutable
type Runtime struct {
    table [][]cell  // Pre-allocated, never modified
}

func (r *Runtime) Execute(input Input) Output {
    // Zero-allocation hot path
    cell := r.table[input.row][input.col]
    return cell.process(input)
}
```

### Real-World Example: go-fsm

```go
// Build phase
type builder[S, E comparable] struct {
    initial    S
    rules      []rule[S, E]
    entryHooks map[S][]ActionFunc[S, E]
    exitHooks  map[S][]ActionFunc[S, E]
    finals     map[S]bool
}

func (b *builder[S, E]) Build() (*Machine[S, E], error) {
    // 1. Validate: at least one rule
    if len(b.rules) == 0 {
        return nil, errors.New("no rules")
    }

    // 2. Validate: initial state exists
    hasInitial := false
    for _, r := range b.rules {
        if r.from == b.initial {
            hasInitial = true
            break
        }
    }
    if !hasInitial {
        return nil, errors.New("initial state not in rules")
    }

    // 3. Validate: final states have no outgoing edges
    for _, r := range b.rules {
        if b.finals[r.from] {
            return nil, fmt.Errorf("final state %v has outgoing edge", r.from)
        }
    }

    // 4. Compile to 2D array lookup table
    stateIndex := assignIndices(b.rules)
    eventIndex := assignIndices(b.rules)
    table := make([][]cell[S, E], len(stateIndex))
    for i := range table {
        table[i] = make([]cell[S, E], len(eventIndex))
    }

    for _, r := range b.rules {
        si := stateIndex[r.from]
        ei := eventIndex[r.on]
        table[si][ei] = cell{to: r.to, guard: r.guard, action: r.action}
    }

    return &Machine[S, E]{
        table:      table,       // Immutable
        stateIndex: stateIndex,  // Immutable
        eventIndex: eventIndex,  // Immutable
    }, nil
}

// Runtime phase - immutable, concurrency-safe
type Machine[S, E comparable] struct {
    table      [][]cell[S, E]
    stateIndex map[S]int
    eventIndex map[E]int
}

func (m *Machine[S, E]) Fire(ctx context.Context, event E) error {
    // Zero-allocation hot path
    current, _ := m.store.Get(ctx)
    si := m.stateIndex[current]  // O(1)
    ei := m.eventIndex[event]    // O(1)
    cell := m.table[si][ei]      // O(1)
    // ...
}
```

---

## Fluent API

**When to use:** Request building, complex configuration, method chaining

**Characteristics:**
- All methods return `*Self` for chaining
- Readable, declarative syntax
- Often combined with Builder pattern

### Pattern

```go
type RequestBuilder struct {
    method  string
    url     string
    headers map[string]string
    query   map[string]string
    body    io.Reader
}

func (rb *RequestBuilder) Header(key, value string) *RequestBuilder {
    rb.headers[key] = value
    return rb
}

func (rb *RequestBuilder) Query(key, value string) *RequestBuilder {
    rb.query[key] = value
    return rb
}

func (rb *RequestBuilder) JSONBody(v any) *RequestBuilder {
    data, _ := json.Marshal(v)
    rb.body = bytes.NewReader(data)
    rb.headers["Content-Type"] = "application/json"
    return rb
}

func (rb *RequestBuilder) Send(ctx context.Context) (*Response, error) {
    // Build and send request
}
```

**Usage:**
```go
resp, err := client.Post("/api/users").
    Header("Authorization", "Bearer "+token).
    Query("page", "1").
    Query("limit", "10").
    JSONBody(user).
    Send(ctx)
```

### Real-World Example: requests

```go
type RequestBuilder struct {
    client  *Client
    method  string
    url     string
    headers http.Header
    query   url.Values
    body    io.Reader
}

func (rb *RequestBuilder) Header(key, value string) *RequestBuilder {
    rb.headers.Set(key, value)
    return rb
}

func (rb *RequestBuilder) Query(key, value string) *RequestBuilder {
    rb.query.Set(key, value)
    return rb
}

func (rb *RequestBuilder) PathParam(key, value string) *RequestBuilder {
    rb.url = strings.ReplaceAll(rb.url, "{"+key+"}", value)
    return rb
}

func (rb *RequestBuilder) JSONBody(v any) *RequestBuilder {
    data, _ := rb.client.jsonEncoder.Encode(v)
    rb.body = bytes.NewReader(data)
    rb.headers.Set("Content-Type", "application/json")
    return rb
}

func (rb *RequestBuilder) Send(ctx context.Context) (*Response, error) {
    req, err := rb.build(ctx)
    if err != nil {
        return nil, err
    }
    return rb.client.do(req)
}
```

---

## Config Struct Pattern

**When to use:** Alternative to functional options, explicit configuration, IDE-friendly

**Characteristics:**
- Struct with public fields
- Zero values are defaults
- No builder, no options
- Explicit and simple

### Pattern

```go
type Config struct {
    Name        string
    Version     string
    Timeout     time.Duration
    MaxRetries  int
    ErrorHandler func(error)
}

func New(cfg Config) *Client {
    // Apply defaults for zero values
    if cfg.Timeout == 0 {
        cfg.Timeout = 30 * time.Second
    }
    if cfg.MaxRetries == 0 {
        cfg.MaxRetries = 3
    }

    return &Client{config: cfg}
}
```

**Usage:**
```go
client := New(Config{
    Name:       "my-client",
    Timeout:    10 * time.Second,
    MaxRetries: 5,
})
```

### Real-World Example: go-command

```go
type Config struct {
    Name            string
    Version         string
    Description     string
    VersionFunc     func() string
    DescriptionFunc func() string
    CommandLoader   func(name string) *Command
    ErrorHandler    ErrorHandler
    HelpRenderer    HelpRenderer
    Formatter       Formatter
    Validator       Validator
}

func New(cfg Config) *App {
    return &App{
        Name:         cfg.Name,
        Version:      cfg.Version,
        Description:  cfg.Description,
        errorHandler: cfg.ErrorHandler,
        helpRenderer: cfg.HelpRenderer,
        formatter:    cfg.Formatter,
        validator:    cfg.Validator,
    }
}
```

**Usage:**
```go
app := command.New(command.Config{
    Name:    "myapp",
    Version: "1.0.0",
    ErrorHandler: customErrorHandler,
})
```

---

## Optional Capabilities

**When to use:** Features that not all implementations need, avoiding interface pollution

**Characteristics:**
- Core interface is minimal
- Optional features via separate interfaces
- Runtime capability detection via type assertion

### Pattern

```go
// Core interface
type Store interface {
    Get(key string) ([]byte, error)
}

// Optional capabilities
type Setter interface {
    Set(key string, value []byte) error
}

type Watcher interface {
    Watch(onChange func(key string, value []byte))
}

type Closer interface {
    Close() error
}

// Implementation
type MemoryStore struct {
    data map[string][]byte
}

func (m *MemoryStore) Get(key string) ([]byte, error) { ... }
func (m *MemoryStore) Set(key string, value []byte) error { ... }
func (m *MemoryStore) Watch(onChange func(key string, value []byte)) { ... }
func (m *MemoryStore) Close() error { ... }

// Usage
func process(store Store) {
    data, _ := store.Get("key")

    if setter, ok := store.(Setter); ok {
        setter.Set("key", data)
    }

    if watcher, ok := store.(Watcher); ok {
        watcher.Watch(func(k string, v []byte) {
            log.Printf("Changed: %s", k)
        })
    }

    if closer, ok := store.(Closer); ok {
        defer closer.Close()
    }
}
```

---

## Generic Constraints

**When to use:** Type-safe APIs, compile-time guarantees, zero interface boxing

**Characteristics:**
- Generic type parameters with constraints
- Compile-time type checking
- Zero runtime overhead

### Pattern

```go
// Document constraint
type Document interface {
    ~[]byte | ~string | map[string]any | any
}

// Generic function with constraint
func Merge[T Document](target, patch T, opts ...Option) (*Result[T], error) {
    // Type-safe operations
}

// Type-safe result wrapper
type Result[T Document] struct {
    Doc T
}

// Usage
result, err := Merge(targetMap, patchMap)  // map[string]any
result, err := Merge(targetJSON, patchJSON)  // []byte
result, err := Merge(targetStruct, patchStruct)  // struct
```

### Real-World Example: go-fsm

```go
// Generic state machine with comparable constraint
type Machine[S comparable, E comparable] struct {
    table      [][]cell[S, E]
    stateIndex map[S]int
    eventIndex map[E]int
}

// Type-safe callbacks
type Info[S, E comparable] struct {
    From, To S
    Event    E
}

type ActionFunc[S, E comparable] func(ctx context.Context, info Info[S, E]) error
type GuardFunc[S, E comparable] func(ctx context.Context, info Info[S, E]) bool

// Usage
type State int
type Event int

m := fsm.New[State, Event](Unpaid).
    From(Unpaid).On(PaySuccess).To(Paid).
    Build()

// Compile-time type safety
m.Fire(ctx, PaySuccess)  // OK
m.Fire(ctx, "invalid")   // Compile error
```

---

## Pattern Selection Guide

| Use Case | Recommended Pattern | Alternative |
|----------|-------------------|-------------|
| Optional configuration | Functional Options | Config Struct |
| Complex object construction | Builder | Functional Options |
| Pluggable components | Interface Segregation | - |
| Validation at build time | Two-Phase Construction | Builder |
| Request building | Fluent API | Builder |
| Explicit configuration | Config Struct | Functional Options |
| Optional features | Optional Capabilities | Interface Segregation |
| Type-safe APIs | Generic Constraints | - |

**Mix and match patterns as needed.** Most libraries combine multiple patterns (e.g., go-fsm uses Builder + Two-Phase Construction + Generic Constraints).
