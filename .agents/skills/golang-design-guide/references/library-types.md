# Library Type Patterns

This document provides detailed patterns and real-world examples for each library type classification.

## 1. Core Types & State Machines

**Examples:** go-fsm, go-command

### Characteristics
- Generic-based with `comparable` constraints
- Zero-allocation hot paths (< 100 ns/op, 0 allocs)
- Immutable runtime after build phase
- Compile-time type safety
- Two-phase separation (build vs runtime)

### Key Design Patterns

#### Two-Phase Separation
```go
// Build phase - mutable, validation
type Builder[S, E comparable] struct {
    initial S
    rules   []rule[S, E]
    finals  map[S]bool
}

func (b *Builder[S, E]) Build() (*Machine[S, E], error) {
    // Validation: initial state exists, finals have no outgoing edges
    // Compile to immutable runtime structure
}

// Runtime - immutable, concurrency-safe, zero-allocation
type Machine[S, E comparable] struct {
    table      [][]cell[S, E]        // Pre-allocated 2D array
    stateIndex map[S]int             // State value → index
    eventIndex map[E]int             // Event value → index
}
```

**Benefits:**
- Build-time validation catches errors early
- Runtime is inherently concurrency-safe (no locks needed)
- Pre-allocated lookup tables enable zero-allocation hot paths

#### Internal Array Lookup
```go
// Build phase: Assign contiguous indices
stateIndex := make(map[S]int)
for i, state := range states {
    stateIndex[state] = i
}

// Runtime: Two map lookups + direct array access
func (m *Machine[S, E]) Fire(ctx context.Context, event E) error {
    si := m.stateIndex[current]  // O(1) map lookup
    ei := m.eventIndex[event]    // O(1) map lookup
    cell := m.table[si][ei]      // O(1) array access
    // ...
}
```

**Performance:** For `iota` integer types, Go's map uses `mapaccess1_fast64` fast path.

#### Dual Declaration Styles
```go
// Style A: Builder method chain (best for hooks, complex guards)
m := New[State, Event](Initial).
    From(Unpaid).
        OnEntry(logEntry).
        On(PaySuccess).To(Paid).Action(publishPaid).
    Final(Completed).
    Build()

// Style B: Transition table (best for pure transitions, compact)
m := New[State, Event](Initial,
    WithRules(
        Rule{Unpaid, PaySuccess, Paid, nil, publishPaid},
        Rule{Paid, Ship, Shipped, nil, notifyWarehouse},
    ),
    WithFinal(Completed),
)
```

**Both styles share the same internal builder.**

### Real-World Example: go-fsm

**Design Philosophy:**
- KISS: Each concept has exactly one representation (no five kinds of trigger behavior)
- DRY: Transition table serves Fire(), Build(), and Mermaid() simultaneously
- YAGNI: No sub-states, parallel states, Actor model, JSON config

**Performance Targets:**
| Scenario | Target |
|----------|--------|
| Pure transition (no guards, no callbacks) | < 60 ns/op, 0 allocs |
| With 1 guard | < 100 ns/op, 0 allocs |
| With Entry callback | < 120 ns/op, 0 allocs |

**API Surface:**
```go
// 4 builder types, each with single responsibility
type Builder[S, E]    // From, Final, OnTransition, Store, Build
type StateGroup[S, E] // On, OnEntry, OnExit
type OnClause[S, E]   // To
type Permit[S, E]     // Guard, Action

// Runtime (immutable)
type Machine[S, E]    // Fire, Can, Current, IsFinal, Permitted, Mermaid, ASCII
```

---

## 2. Data Processing & Transformation

**Examples:** jsondiff, jsonmerge, jsonpatch, jsonschema

### Characteristics
- RFC compliance first (RFC 7386, RFC 6902, etc.)
- Flat output schemas (not nested delta trees)
- Zero-copy optimization where possible
- Type-safe generics with Document constraint
- Benchmark-driven performance

### Key Design Patterns

#### Fast/Slow Path Separation
```go
func Merge[T Document](target, patch T, opts ...Option) (*Result[T], error) {
    // Fast path: Type assertions for common JSON types (zero reflection)
    switch t := any(target).(type) {
    case map[string]any:
        // Direct map operations, zero conversion overhead
        return mergeMap(t, any(patch).(map[string]any))
    case []any:
        // Direct slice operations
        return mergeSlice(t, any(patch).([]any))
    }

    // Slow path: Reflection for typed slices, structs
    return mergeSlow(target, patch)
}
```

**Performance Impact:** Fast path is 3-5x faster than slow path.

#### Minimal API Surface
```go
// Only 3 public functions
func Merge[T Document](target, patch T, ...Option) (*Result[T], error)
func Generate[T Document](source, target T) (T, error)
func Valid[T Document](patch T) bool

// Type-safe result wrapper
type Result[T Document] struct {
    Doc T
}
```

**Benefits:** Simple API, easy to learn, hard to misuse.

#### RFC Compliance First
```go
// applyPatch() directly implements RFC 7386 Section 2
func applyPatch(target, patch any) any {
    // RFC 7386: If patch is not an object, replace target
    patchMap, ok := patch.(map[string]any)
    if !ok {
        return patch
    }

    // RFC 7386: If target is not an object, replace with empty object
    targetMap, ok := target.(map[string]any)
    if !ok {
        targetMap = make(map[string]any)
    }

    // RFC 7386: Merge recursively
    for key, patchVal := range patchMap {
        if patchVal == nil {
            delete(targetMap, key)  // RFC 7386: null deletes
        } else {
            targetMap[key] = applyPatch(targetMap[key], patchVal)
        }
    }
    return targetMap
}
```

### Real-World Example: jsonmerge

**Design Philosophy:**
- RFC 7386 First: Every decision prioritizes RFC compliance
- Type Safety Through Generics: Compile-time type safety for structs, maps, JSON bytes
- Immutable by Default: Uses deepclone to prevent side effects
- Minimal API Surface: Only 3 public functions
- Zero-Copy Optimization: map[string]any has zero conversion overhead

**Forbidden Patterns (Documented to Prevent Regression):**
- Using `bytes.Equal` instead of string comparison (34% slower)
- Extracting helpers from hot paths (53% slower)
- Adding `a == b` fast path in deepEqual (panics on uncomparable types)

**Benchmark Results (Apple M3):**
```
BenchmarkMerge-8                  952150     1357 ns/op    1273 B/op   17 allocs/op
BenchmarkMergeWithMutate-8       2400202      466 ns/op     345 B/op    4 allocs/op
```

---

## 3. Infrastructure & Storage

**Examples:** go-cache, go-secrets, go-config

### Characteristics
- Store/Source interface abstraction
- Pluggable backends (memory, Redis, SQLite, Postgres, file, env)
- Explicit configuration (no magic defaults)
- Lock-free reads after load
- Policy-based error handling

### Key Design Patterns

#### Interface Segregation
```go
// Core interface - single method
type Store interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
    Close() error
}

// Optional capabilities - checked via type assertion
type BatchStore interface {
    GetMany(ctx context.Context, keys []string) (map[string][]byte, error)
    SetMany(ctx context.Context, items map[string][]byte, ttl time.Duration) error
}

type Watcher interface {
    Watch(ctx context.Context, onChange func(map[string]any)) error
}
```

**Benefits:** Small interfaces, optional capabilities, no interface pollution.

#### Functional Options with Explicit Configuration
```go
type Config struct {
    Store        Store        // Required, no default
    Codec        Codec        // Required, no default
    DefaultTTL   time.Duration
    SingleFlight bool
}

type Option func(*Config)

func WithTTL(ttl time.Duration) Option {
    return func(c *Config) { c.DefaultTTL = ttl }
}

func New[T any](store Store, opts ...Option) Cache[T] {
    cfg := &Config{
        Store:      store,  // Explicit, no magic
        DefaultTTL: 5 * time.Minute,
    }
    for _, opt := range opts {
        opt(cfg)
    }
    return &cache[T]{config: cfg}
}
```

**No magic defaults for critical components (Store, Codec).**

#### Lock-Free Reads
```go
type Config[T any] struct {
    snapshot atomic.Pointer[T]  // Atomic snapshot
}

func (c *Config[T]) Value() T {
    return *c.snapshot.Load()  // Lock-free, allocation-free
}

func (c *Config[T]) Load(ctx context.Context, sources ...Source) error {
    // Load, merge, decode
    newSnapshot := &result
    c.snapshot.Store(newSnapshot)  // Atomic publish
}
```

**Benefits:** Zero contention for readers, non-blocking updates.

#### Policy-Based Error Handling
```go
type Policy int

const (
    PolicyRequired Policy = iota  // Error is fatal
    PolicyOptional                // Error is recorded, source skipped
)

type SkipReporter interface {
    SkipError() error  // Expose absorbed errors for provenance
}

// File provider with policy
file.New(path, file.WithPolicy(file.PolicyOptional))
```

### Real-World Example: go-config

**Design Philosophy:**
- Type-safe first: Users consume typed structs, not untyped getters
- Small surface area: Few top-level concepts (Source, Config, Option)
- Zero-dependency core: Standard library only
- Fast read path: Value() is atomic and allocation-free
- Explicit precedence: Source order defines override behavior

**Pipeline (Fixed, Not Extensible):**
1. Load each source to `map[string]any`
2. Merge by source order (later wins)
3. Interpolate (`${env:VAR}`, `${secret:KEY}`, `${path.to.key}`)
4. Decode merged map into `T` via reflect-based converter
5. Atomically publish new snapshot via `atomic.Pointer[T]`

**Progressive Disclosure:**
- Metadata (name + description): Always in context
- Source.Load(): When source triggers
- Watch(): Only for sources implementing Watcher

---

## 4. Framework-Agnostic Libraries

**Examples:** go-crawler, requests

### Characteristics
- Embeddable core with no framework dependencies
- Dual interface (CLI + library share same kernel)
- Two-layer architecture (API + Core)
- Middleware/plugin architecture
- Graceful degradation

### Key Design Patterns

#### Two-Layer Architecture
```
┌─────────────────────────────────────┐
│  API Layer (crawler.go)              │
│  - Acquire(url, ...Option)           │
│  - AcquireBatch(urls, ...Option)     │
│  - Crawl(seeds, ...Option)           │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  Core Layer                          │
│  - types/     (domain models)        │
│  - errors/    (sentinel errors)      │
│  - fetcher/   (HTTP + browser)       │
│  - extractor/ (readability)          │
│  - converter/ (HTML→Markdown)        │
│  - crawler/   (BFS/DFS traversal)    │
│  - pipeline/  (composition only)     │
└─────────────────────────────────────┘
```

**Dependency Rules:**
- API layer depends on Core layer
- Core packages have no cycles
- Foundation (types/, errors/) has zero internal dependencies
- Capability packages depend only on Foundation
- Pipeline composes Capability components via interfaces

#### Capability Interfaces
```go
// Foundation layer - zero dependencies
type Page struct {
    InputURL, FinalURL string
    StatusCode         int
    Body               []byte
    SourceKind         SourceKind
}

type Content struct {
    Title, ContentHTML, ContentText string
    Byline, Lang                    string
    WordCount, ParagraphCount       int
}

// Capability layer - single-method interfaces
type Fetcher interface {
    Fetch(ctx context.Context, url string) (Page, error)
}

type Extractor interface {
    Extract(ctx context.Context, page Page) (Content, error)
}

type Converter interface {
    Convert(ctx context.Context, content Content) (Document, error)
    ConvertRaw(ctx context.Context, page Page) (Document, error)
}
```

**Benefits:** Testable in isolation, pluggable implementations, clear contracts.

#### Graceful Degradation
```go
// Pipeline routing with fallback
func (p *Pipeline) Acquire(ctx context.Context, url string) (Document, error) {
    page, err := p.fetcher.Fetch(ctx, url)
    if err != nil {
        return Document{}, err
    }

    // Direct passthrough for markdown
    if page.SourceKind == SourceKindMarkdown {
        return p.converter.ConvertRaw(ctx, page)
    }

    // Try extraction
    content, err := p.extractor.Extract(ctx, page)
    if err != nil {
        // Fallback: full-page conversion
        return p.converter.ConvertRaw(ctx, page)
    }

    doc, err := p.converter.Convert(ctx, content)
    if err != nil || doc.Markdown == "" {
        // Fallback: full-page conversion
        return p.converter.ConvertRaw(ctx, page)
    }

    return doc, nil
}
```

### Real-World Example: go-crawler

**Design Philosophy:**
- KISS: Simple API, minimal abstractions
- DRY: Store interface shared across all backends
- YAGNI: No speculative features
- Single Responsibility: Each package has one clear purpose
- Fail-Fast at Startup: Assembly errors panic at construction time

**Rejected Patterns (Explicitly Documented):**
- Middleware/Hooks System: Adds complexity without clear use case
- Compression/Encryption Codecs: Users can implement custom Codec
- Adaptive Cleanup Frequency: Premature optimization
- Reflection in Hot Paths: Use generics for type safety

---

## 5. HTTP & Network Libraries

**Examples:** requests

### Characteristics
- Fluent builder pattern for request configuration
- Middleware chains for cross-cutting concerns
- Streaming support for large responses
- Retry mechanisms with backoff strategies
- Zero-panic policy

### Key Design Patterns

#### Fluent Builder Pattern
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

func (rb *RequestBuilder) JSONBody(v any) *RequestBuilder {
    data, _ := json.Marshal(v)
    rb.body = bytes.NewReader(data)
    rb.headers.Set("Content-Type", "application/json")
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
    JSONBody(user).
    Send(ctx)
```

#### Middleware Chain
```go
type Middleware func(req *http.Request, next RequestFunc) (*http.Response, error)

type RequestFunc func(*http.Request) (*http.Response, error)

func (c *Client) buildChain() RequestFunc {
    handler := c.httpClient.Do

    // Wrap in reverse order (last middleware wraps first)
    for i := len(c.middlewares) - 1; i >= 0; i-- {
        handler = c.middlewares[i](handler)
    }

    return handler
}

// Example middleware
func LoggingMiddleware(req *http.Request, next RequestFunc) (*http.Response, error) {
    start := time.Now()
    resp, err := next(req)
    log.Printf("%s %s - %d (%s)", req.Method, req.URL, resp.StatusCode, time.Since(start))
    return resp, err
}
```

#### Buffer Pooling
```go
import "github.com/valyala/bytebufferpool"

var bufferPool = &bytebufferpool.Pool{}

func (r *Response) readBody() error {
    buf := bufferPool.Get()
    defer bufferPool.Put(buf)

    _, err := io.Copy(buf, r.RawResponse.Body)
    r.BodyBytes = append([]byte(nil), buf.Bytes()...)
    return err
}
```

**Benefits:** Reduces GC pressure in high-throughput scenarios.

### Real-World Example: requests

**Design Philosophy:**
- Fluent Builder Pattern: All request configuration uses method chaining
- Middleware-First Architecture: Extensible request/response processing pipeline
- Zero-Panic Policy: Library code returns errors instead of panicking
- Memory Efficiency: Uses buffer pooling for high-throughput scenarios
- Modern Go Features: Leverages Go 1.26 features (iterators, Swiss Tables)

**Memory Management:**
- Buffer Pooling: Use GetBuffer()/PutBuffer() for temporary buffers
- Pre-allocation: Pre-allocate maps/slices when size is known
- Zero-Copy: Minimize data copying in hot paths

**Performance Optimization:**
- Buffer pooling is critical for high-throughput scenarios
- HTTP/2 enabled via Config.HTTP2 for connection multiplexing
- Streaming callbacks for large responses to avoid memory spikes

---

## Choosing the Right Pattern

| Library Type | Primary Pattern | Secondary Pattern | Example |
|--------------|----------------|-------------------|---------|
| Core Types & State Machines | Two-phase separation | Internal array lookup | go-fsm |
| Data Processing | Fast/slow path | Minimal API | jsonmerge |
| Infrastructure | Interface segregation | Functional options | go-cache |
| Framework-Agnostic | Two-layer architecture | Capability interfaces | go-crawler |
| HTTP & Network | Fluent builder | Middleware chain | requests |

**Decision Tree:**

1. **Does it manage state transitions?** → Core Types & State Machines
2. **Does it transform data according to RFC/spec?** → Data Processing
3. **Does it provide pluggable storage/sources?** → Infrastructure
4. **Does it need both CLI and library interfaces?** → Framework-Agnostic
5. **Does it make HTTP requests?** → HTTP & Network

**Mix and match patterns as needed.** Most libraries combine multiple patterns (e.g., go-config uses both Interface Segregation and Functional Options).
