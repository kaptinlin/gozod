# Design Philosophy: KISS, DRY, YAGNI

This document provides real-world examples and anti-patterns for applying KISS, DRY, and YAGNI principles in Golang library design.

## KISS - Keep It Simple, Stupid

### Principle
Prefer simple solutions over clever abstractions. Clear code beats micro-optimizations. Three similar lines beat a single-use helper.

### Real-World Examples

#### Good: Simple Error Aggregation (Go 1.20+)
```go
// Simple and clear
var errs []error
for _, item := range items {
    if err := validate(item); err != nil {
        errs = append(errs, err)
    }
}
return errors.Join(errs...)
```

#### Bad: Custom Error Aggregation
```go
// Over-engineered
type ErrorCollector struct {
    errors []error
    mu     sync.Mutex
}

func (ec *ErrorCollector) Add(err error) { ... }
func (ec *ErrorCollector) HasErrors() bool { ... }
func (ec *ErrorCollector) Join() error { ... }
```

**Why bad:** Standard library already provides `errors.Join()`. No need for custom abstraction.

#### Good: Simple Iteration (Go 1.22+)
```go
// Clean and obvious
for i := range 100 {
    process(i)
}
```

#### Bad: Custom Iterator
```go
// Unnecessary complexity
func ProcessSeq(items []Item) iter.Seq[Result] {
    return func(yield func(Result) bool) {
        for _, item := range items {
            if !yield(process(item)) {
                return
            }
        }
    }
}

// Usage is more complex than a simple for loop
for result := range ProcessSeq(items) {
    // ...
}
```

**Why bad:** Users can write their own for loops. Custom iterators add complexity without clear benefit.

### From go-fsm

**KISS Statement:**
> Each concept has exactly one representation. The core of a state machine is states + events + transitions. No need for five kinds of trigger behaviour, no need to distinguish transitioning / reentry / internal / ignored / dynamic. Keep the implementation lean, achieving nanosecond-level zero-allocation performance.

**Example:**
```go
// KISS: Single transition representation
type Rule[S, E comparable] struct {
    From   S
    On     E
    To     S
    Guard  GuardFunc[S, E]   // nil = unconditional
    Action ActionFunc[S, E]  // nil = no callback
}

// NOT: Multiple trigger types
// type InternalTransition, type ReentryTransition, type DynamicTransition, etc.
```

### From jsonmerge

**KISS Statement:**
> Keep hot path code inline - function call overhead matters in performance-critical paths.

**Example:**
```go
// KISS: Inline hot path
func applyPatch(target, patch any) any {
    patchMap, ok := patch.(map[string]any)
    if !ok {
        return patch  // Inline, no function call
    }
    // ... rest of logic inline
}

// NOT: Extracted helpers (53% slower)
func applyPatch(target, patch any) any {
    if !isObject(patch) {  // Function call overhead
        return patch
    }
    return mergeObjects(target, patch)  // Function call overhead
}
```

### From go-secrets

**KISS Statement:**
> All Store implementations use `storekey.CheckClosed()` directly for closed checks (KISS — no wrapper indirection).

**Example:**
```go
// KISS: Direct check
func (s *Store) Get(ctx context.Context, scope, name string) ([]byte, error) {
    if err := storekey.CheckClosed(&s.closed); err != nil {
        return nil, err
    }
    // ...
}

// NOT: Wrapper indirection
func (s *Store) checkClosed() error {
    return storekey.CheckClosed(&s.closed)
}

func (s *Store) Get(ctx context.Context, scope, name string) ([]byte, error) {
    if err := s.checkClosed(); err != nil {  // Unnecessary indirection
        return nil, err
    }
    // ...
}
```

---

## DRY - Don't Repeat Yourself

### Principle
Reuse stdlib packages, shared patterns across packages, eliminate duplicate validation/error handling/test setup.

### Real-World Examples

#### Good: Reuse Stdlib (Go 1.21+)
```go
// DRY: Use slices package
import "slices"

func deduplicate(items []string) []string {
    result := make([]string, 0, len(items))
    for _, item := range items {
        if !slices.Contains(result, item) {
            result = append(result, item)
        }
    }
    return result
}
```

#### Bad: Custom Helper
```go
// Duplicate: Reimplementing stdlib
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func deduplicate(items []string) []string {
    result := make([]string, 0, len(items))
    for _, item := range items {
        if !contains(result, item) {  // Custom helper
            result = append(result, item)
        }
    }
    return result
}
```

#### Good: Shared Functional Options Pattern
```go
// DRY: Consistent pattern across packages
type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) { c.Timeout = d }
}

func New(opts ...Option) *Client {
    cfg := &Config{Timeout: 30 * time.Second}
    for _, opt := range opts {
        opt(cfg)
    }
    return &Client{config: cfg}
}
```

### From go-fsm

**DRY Statement:**
> Rules are declared once; runtime, validation, and visualization share the same data. The transition table serves `Fire()` dispatch, `Build()` validation, and `Mermaid()`/`ASCII()` export simultaneously. No separate graph structure is maintained for visualization.

**Example:**
```go
// DRY: Single source of truth
type builder[S, E comparable] struct {
    rules []rule[S, E]  // Single declaration
}

func (b *builder[S, E]) Build() (*Machine[S, E], error) {
    // Validation uses rules
    for _, r := range b.rules {
        if err := validateRule(r); err != nil {
            return nil, err
        }
    }
    // Runtime uses rules
    table := compileTable(b.rules)
    return &Machine{table: table}, nil
}

func (m *Machine[S, E]) Mermaid() string {
    // Visualization uses same rules (via compiled table)
    return renderMermaid(m.table)
}
```

### From go-config

**DRY Statement:**
> Format decoders are modular packages that self-register on import. File and FS providers resolve format by extension via the global registry.

**Example:**
```go
// DRY: Single registration mechanism
func RegisterFormat(ext string, decoder FormatDecoder) {
    formats[ext] = decoder
}

// format/json package
func init() {
    RegisterFormat(".json", jsonDecoder)
}

// format/yaml package
func init() {
    RegisterFormat(".yaml", yamlDecoder)
    RegisterFormat(".yml", yamlDecoder)
}

// Both file and fs providers use the same registry
func (f *File) Load(ctx context.Context) (map[string]any, error) {
    decoder := LookupFormat(filepath.Ext(f.path))  // Shared lookup
    // ...
}
```

### From go-cache

**DRY Statement:**
> Store interface shared across all backends, codec abstraction eliminates serialization duplication.

**Example:**
```go
// DRY: Single Store interface for all backends
type Store interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
    Close() error
}

// All backends implement the same interface
type MemoryStore struct { ... }
type RedisStore struct { ... }
type SQLiteStore struct { ... }
type PostgresStore struct { ... }

// Cache[T] works with any Store
type cache[T any] struct {
    store Store  // Polymorphic
    codec Codec
}
```

---

## YAGNI - You Aren't Gonna Need It

### Principle
Solve current problems, not hypothetical futures. Evidence-based optimization. No speculative features, unused configurability, or "just in case" abstractions.

### Real-World Examples

#### Good: Evidence-Based sync.Pool
```go
// YAGNI: Only add sync.Pool when profiling shows allocation bottlenecks
// Example from messageformat-go v1 (hot path with proven need)
var pool = sync.Pool{
    New: func() any {
        return &formatter{}
    },
}

func Format(msg string) string {
    f := pool.Get().(*formatter)
    defer pool.Put(f)
    return f.format(msg)
}
```

#### Bad: Speculative sync.Pool
```go
// YAGNI violation: Adding sync.Pool without profiling evidence
var userPool = sync.Pool{
    New: func() any {
        return &User{}
    },
}

func GetUser(id string) (*User, error) {
    user := userPool.Get().(*User)  // Premature optimization
    defer userPool.Put(user)
    // ... fetch user from database
    return user, nil
}
```

**Why bad:** No evidence that User allocation is a bottleneck. Adds complexity without proven benefit.

#### Good: Simple Default
```go
// YAGNI: Simple default, no configurability
func New[T any](store Store) Cache[T] {
    return &cache[T]{
        store:      store,
        defaultTTL: 5 * time.Minute,  // Fixed default
    }
}
```

#### Bad: Over-Configurable
```go
// YAGNI violation: Unused configurability
type Config struct {
    DefaultTTL          time.Duration
    MaxTTL              time.Duration
    MinTTL              time.Duration
    TTLJitter           float64
    TTLJitterMode       string  // "fixed", "random", "exponential"
    AdaptiveTTL         bool
    AdaptiveTTLStrategy string  // "usage-based", "time-based", "hybrid"
}
```

**Why bad:** No evidence that users need adaptive TTL or jitter. Adds complexity without proven need.

### From go-fsm

**YAGNI Statement:**
> Only implement what's truly needed. No sub-states (orders don't need them). No parallel states (that's a workflow engine's job). No Actor model. No JSON declarative config. Add things when needed — the cost of adding later is far lower than maintaining unused features.

**Explicitly Rejected Features:**

| Feature | Why Not |
|---------|---------|
| Sub-states / hierarchical states | Order scenarios don't need them, adds recursive Enter/Exit logic |
| Parallel states | Workflow engine's responsibility, not FSM's |
| Async transitions | Two-phase commit adds massive complexity, users handle async in callbacks |
| Dynamic targets | Breaks declarative predictability, use Guard + multiple rules instead |
| FiringMode | Only Immediate, simple and direct |
| Trigger parameters | `args ...any` is type-unsafe, use `context.WithValue` for business data |
| History | Memory overhead, use logs/database for production recording |
| JSON serialization | Rule definitions are code, state values are StateStore's responsibility |
| Graphviz output | Mermaid + ASCII covers all rendering scenarios |
| `SetState()` / `Force()` | Rule-bypassing methods break state machine invariants |

### From go-cache

**YAGNI Statement:**
> The following patterns are explicitly rejected to maintain simplicity.

**Rejected Patterns:**

| Pattern | Why Rejected |
|---------|--------------|
| Middleware/Hooks System | Adds complexity without clear use case. Users can wrap Cache interface if needed. |
| Compression/Encryption Codecs | Users can implement custom Codec interface. No need for built-in wrappers. |
| Adaptive Cleanup Frequency | Premature optimization. Fixed intervals with configuration are sufficient. |
| Batch Size Enforcement | Documentation is sufficient. Hard limits add complexity and may not fit all use cases. |
| Reflection in Hot Paths | Use generics for type safety and performance. |

### From jsonmerge

**YAGNI Statement:**
> Failed optimization patterns (documented to prevent repetition).

**Documented Failed Optimizations:**
- Using `bytes.Equal` instead of string comparison (34% slower)
- Extracting helpers from `convertToInterface` (53% slower)
- Adding `a == b` fast path in `deepEqual` (panics on uncomparable types)
- Using `reflect.DeepEqual` (panics on uncomparable types)

**Why document failures:** Prevents future contributors from repeating the same mistakes.

### From go-secrets

**YAGNI Statement:**
> No business model awareness in core (no workspace/app/user concepts). No secret versioning/rollback in core (Store implementor's concern). No RBAC/ACL in core (business layer concern). No Watch/real-time push (secrets change infrequently).

**Forbidden:**
- No global mutable state
- No business model awareness in core
- No secret versioning/rollback in core (Store implementor's concern)
- No RBAC/ACL in core (business layer concern)
- No Watch/real-time push (secrets change infrequently)
- No premature abstraction — three similar lines are better than a helper used once

---

## Combining KISS, DRY, YAGNI

### Example: go-config Source Interface

**KISS:** Single-method interface
```go
type Source interface {
    Load(ctx context.Context) (map[string]any, error)
}
```

**DRY:** Optional capabilities via type assertion (not separate interfaces)
```go
type Watcher interface {
    Watch(ctx context.Context, onChange func(map[string]any)) error
}

// Check capability
if w, ok := source.(Watcher); ok {
    w.Watch(ctx, onChange)
}
```

**YAGNI:** No speculative features
```go
// NOT: Unused capabilities
type Source interface {
    Load(ctx context.Context) (map[string]any, error)
    Watch(ctx context.Context, onChange func(map[string]any)) error  // Not all sources need this
    Validate() error  // Not all sources need this
    Reload() error    // Not all sources need this
    Export() ([]byte, error)  // Not all sources need this
}
```

### Example: go-fsm Machine API

**KISS:** Minimal runtime API
```go
type Machine[S, E comparable] struct{}

func (m *Machine[S, E]) Fire(ctx context.Context, event E) error
func (m *Machine[S, E]) Can(ctx context.Context, event E) bool
func (m *Machine[S, E]) Current(ctx context.Context) (S, error)
func (m *Machine[S, E]) IsFinal(ctx context.Context) (bool, error)
func (m *Machine[S, E]) Permitted(ctx context.Context) ([]E, error)
```

**DRY:** Single internal representation for all operations
```go
type Machine[S, E comparable] struct {
    table [][]cell[S, E]  // Used by Fire, Can, Permitted, Mermaid, ASCII
}
```

**YAGNI:** No unused methods
```go
// NOT: Speculative features
func (m *Machine[S, E]) SetState(s S) error  // Breaks invariants
func (m *Machine[S, E]) History() []S        // Memory overhead
func (m *Machine[S, E]) ToJSON() ([]byte, error)  // Unused
```

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Premature Abstraction
```go
// BAD: Abstract before you have 3+ concrete cases
type Processor interface {
    Process(data any) (any, error)
}

type JSONProcessor struct{}
func (p *JSONProcessor) Process(data any) (any, error) { ... }

// GOOD: Wait until you have multiple processors
func ProcessJSON(data []byte) (Result, error) { ... }
// Add interface only when you have ProcessXML, ProcessYAML, etc.
```

### Anti-Pattern 2: Over-Engineering Error Handling
```go
// BAD: Custom error types for everything
type ValidationError struct { Field string; Message string }
type NetworkError struct { URL string; StatusCode int }
type DatabaseError struct { Query string; Cause error }

// GOOD: Simple sentinel errors + wrapping
var (
    ErrValidation = errors.New("validation failed")
    ErrNetwork    = errors.New("network error")
    ErrDatabase   = errors.New("database error")
)

return fmt.Errorf("validation failed: %w", ErrValidation)
```

### Anti-Pattern 3: Speculative Configuration
```go
// BAD: Configuration for hypothetical features
type Config struct {
    EnableFeatureA bool
    EnableFeatureB bool
    EnableFeatureC bool
    FeatureAMode   string
    FeatureBMode   string
    FeatureCMode   string
}

// GOOD: Configuration for actual features
type Config struct {
    Timeout time.Duration
    MaxRetries int
}
```

### Anti-Pattern 4: Unused Interfaces
```go
// BAD: Interface with no implementations
type Serializer interface {
    Serialize(v any) ([]byte, error)
    Deserialize(data []byte, v any) error
}

// Only one implementation exists
type JSONSerializer struct{}

// GOOD: No interface until you have 2+ implementations
func SerializeJSON(v any) ([]byte, error) { ... }
func DeserializeJSON(data []byte, v any) error { ... }
```

---

## Checklist for Design Reviews

### KISS Checklist
- [ ] Can this be solved with stdlib instead of custom code?
- [ ] Is this abstraction used in 3+ places?
- [ ] Can a simple for loop replace this custom iterator?
- [ ] Is this helper function called more than once?
- [ ] Does this code prioritize clarity over cleverness?

### DRY Checklist
- [ ] Is this validation logic duplicated across functions?
- [ ] Are multiple packages reimplementing the same pattern?
- [ ] Can this be extracted to a shared utility?
- [ ] Is this error handling repeated in multiple places?
- [ ] Are test setup/teardown duplicated across test files?

### YAGNI Checklist
- [ ] Is this feature requested by users or speculative?
- [ ] Is this optimization backed by profiling data?
- [ ] Is this configuration option actually used?
- [ ] Is this interface implemented by 2+ types?
- [ ] Can this feature be added later without breaking changes?

---

## Summary

**KISS:** Prefer simple solutions. Standard library first. Clear code beats clever code.

**DRY:** Reuse stdlib. Share patterns. Eliminate duplication.

**YAGNI:** Solve current problems. Evidence-based optimization. No speculative features.

**Balance:** Sometimes KISS conflicts with DRY (duplication is simpler than abstraction). When in doubt, favor KISS and YAGNI over DRY. Three similar lines are better than a premature abstraction.
