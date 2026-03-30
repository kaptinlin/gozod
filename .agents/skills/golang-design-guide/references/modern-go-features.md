# Modern Go Features (Go 1.20-1.26)

This document provides guidance on using modern Go features in library design with real-world examples.

## Table of Contents

1. [Go 1.20 Features](#go-120-features)
2. [Go 1.21 Features](#go-121-features)
3. [Go 1.22 Features](#go-122-features)
4. [Go 1.23 Features](#go-123-features)
5. [Go 1.24 Features](#go-124-features)
6. [Go 1.26 Features](#go-126-features)
7. [Features to Avoid](#features-to-avoid)

---

## Go 1.20 Features

### errors.Join() - Multiple Error Aggregation

**When to use:** Collecting multiple errors from batch operations, validation

```go
// Before Go 1.20: Custom error aggregation
type multiError struct {
    errors []error
}

func (m *multiError) Error() string {
    var msgs []string
    for _, err := range m.errors {
        msgs = append(msgs, err.Error())
    }
    return strings.Join(msgs, "; ")
}

// Go 1.20+: Use errors.Join
func ValidateBatch(items []Item) error {
    var errs []error

    for i, item := range items {
        if err := Validate(item); err != nil {
            errs = append(errs, fmt.Errorf("item %d: %w", i, err))
        }
    }

    if len(errs) > 0 {
        return errors.Join(errs...)
    }

    return nil
}
```

**Real-World Example: go-fsm**

```go
func (b *Builder[S, E]) Build() (*Machine[S, E], error) {
    var errs []error

    if len(b.rules) == 0 {
        errs = append(errs, ErrNoRules)
    }

    for _, r := range b.rules {
        if b.finals[r.from] {
            errs = append(errs, fmt.Errorf("%w: state %v", ErrFinalOutgoing, r.from))
        }
    }

    if len(errs) > 0 {
        return nil, errors.Join(errs...)
    }

    return &Machine[S, E]{...}, nil
}
```

### strings.Cut/CutPrefix/CutSuffix - String Manipulation

**When to use:** Parsing strings, splitting on delimiters

```go
// Before Go 1.20: Manual string manipulation
func parseKeyValue(s string) (key, value string, ok bool) {
    idx := strings.Index(s, "=")
    if idx == -1 {
        return "", "", false
    }
    return s[:idx], s[idx+1:], true
}

// Go 1.20+: Use strings.Cut
func parseKeyValue(s string) (key, value string, ok bool) {
    return strings.Cut(s, "=")
}

// CutPrefix example
func stripScheme(url string) string {
    if after, found := strings.CutPrefix(url, "https://"); found {
        return after
    }
    if after, found := strings.CutPrefix(url, "http://"); found {
        return after
    }
    return url
}

// CutSuffix example
func stripExtension(filename string) string {
    name, _, _ := strings.Cut(filename, ".")
    return name
}
```

**Real-World Example: go-config**

```go
func parseInterpolation(s string) (kind, key string, ok bool) {
    // ${env:VAR} → kind="env", key="VAR"
    s, ok = strings.CutPrefix(s, "${")
    if !ok {
        return "", "", false
    }

    s, ok = strings.CutSuffix(s, "}")
    if !ok {
        return "", "", false
    }

    kind, key, ok = strings.Cut(s, ":")
    return kind, key, ok
}
```

---

## Go 1.21 Features

### Built-in Functions: min(), max(), clear()

**When to use:** Comparing values, clearing maps/slices

```go
// min/max for ordered types
func clamp(value, minVal, maxVal int) int {
    return min(max(value, minVal), maxVal)
}

// clear() for maps
func (c *Cache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    clear(c.data)  // More efficient than c.data = make(map[string][]byte)
}

// clear() for slices (sets length to 0, keeps capacity)
func (b *Builder) Reset() {
    clear(b.rules)  // Keeps underlying array
}
```

**Real-World Example: go-cache**

```go
func (c *cache[T]) Clear(ctx context.Context) error {
    if err := c.store.Clear(ctx); err != nil {
        return err
    }

    c.mu.Lock()
    clear(c.localCache)  // Clear local cache map
    c.mu.Unlock()

    return nil
}
```

### slices Package - Slice Operations

**When to use:** Sorting, searching, comparing slices

```go
import "slices"

// Sorting
func sortUsers(users []User) {
    slices.SortFunc(users, func(a, b User) int {
        return cmp.Compare(a.Name, b.Name)
    })
}

// Searching
func contains(items []string, target string) bool {
    return slices.Contains(items, target)
}

// Comparing
func equal(a, b []int) bool {
    return slices.Equal(a, b)
}

// Growing capacity
func ensureCapacity(slice []int, n int) []int {
    return slices.Grow(slice, n)
}

// Cloning
func clone(slice []string) []string {
    return slices.Clone(slice)
}
```

**Real-World Example: go-fsm**

```go
func (m *Machine[S, E]) Permitted(ctx context.Context) ([]E, error) {
    current, err := m.store.Get(ctx)
    if err != nil {
        return nil, err
    }

    si := m.stateIndex[current]
    row := m.table[si]

    var events []E
    for ei, cell := range row {
        if cell.to != m.zero {
            events = append(events, m.indexToEvent[ei])
        }
    }

    // Sort for deterministic output
    slices.SortFunc(events, func(a, b E) int {
        return cmp.Compare(a, b)
    })

    return events, nil
}
```

### maps Package - Map Operations

**When to use:** Cloning, comparing, copying maps

```go
import "maps"

// Cloning
func cloneConfig(cfg map[string]any) map[string]any {
    return maps.Clone(cfg)
}

// Comparing
func equal(a, b map[string]int) bool {
    return maps.Equal(a, b)
}

// Copying (merge)
func merge(dst, src map[string]any) {
    maps.Copy(dst, src)
}
```

**Real-World Example: go-config**

```go
func (c *Config[T]) Load(ctx context.Context, sources ...Source) error {
    merged := make(map[string]any)

    for _, source := range sources {
        data, err := source.Load(ctx)
        if err != nil {
            return err
        }

        // Merge using maps.Copy
        maps.Copy(merged, data)
    }

    result, err := decode[T](merged)
    if err != nil {
        return err
    }

    c.snapshot.Store(&result)
    return nil
}
```

### cmp Package - Comparison

**When to use:** Comparing ordered types, custom comparisons

```go
import "cmp"

// Comparing ordered types
func compareInts(a, b int) int {
    return cmp.Compare(a, b)
}

// Custom comparison with Or
func compareUsers(a, b User) int {
    return cmp.Or(
        cmp.Compare(a.LastName, b.LastName),
        cmp.Compare(a.FirstName, b.FirstName),
        cmp.Compare(a.ID, b.ID),
    )
}
```

---

## Go 1.22 Features

### for range N - Integer Range Loops

**When to use:** Iterating N times without index variable

```go
// Before Go 1.22: Unused index variable
for i := 0; i < 10; i++ {
    process()
}

// Go 1.22+: Direct integer range
for range 10 {
    process()
}

// With index
for i := range 10 {
    fmt.Println(i)  // 0, 1, 2, ..., 9
}
```

**Real-World Example: Testing**

```go
func BenchmarkOperation(b *testing.B) {
    // Warmup
    for range 100 {
        _ = Operation()
    }

    b.ResetTimer()
    for b.Loop() {
        _ = Operation()
    }
}
```

### Loop Variable Scoping - Automatic Fix

**What changed:** Loop variables are now per-iteration, not per-loop

```go
// Before Go 1.22: Bug - all goroutines see final value
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func() {
        defer wg.Done()
        process(item)  // BUG: All goroutines see last item
    }()
}

// Before Go 1.22: Workaround
for _, item := range items {
    item := item  // Shadow variable
    wg.Add(1)
    go func() {
        defer wg.Done()
        process(item)  // OK: Each goroutine sees correct item
    }()
}

// Go 1.22+: Automatic fix
for _, item := range items {
    wg.Add(1)
    go func() {
        defer wg.Done()
        process(item)  // OK: Automatic per-iteration scoping
    }()
}
```

---

## Go 1.23 Features

### iter.Seq/Seq2 - Custom Iterators

**When to use:** Lazy evaluation, streaming data, custom iteration patterns

**⚠️ Use sparingly:** Most cases don't need custom iterators. Users can write their own for loops.

```go
import "iter"

// Basic iterator
func Range(start, end int) iter.Seq[int] {
    return func(yield func(int) bool) {
        for i := start; i < end; i++ {
            if !yield(i) {
                return
            }
        }
    }
}

// Usage
for i := range Range(0, 10) {
    fmt.Println(i)
}

// Seq2 for key-value pairs
func Items[K comparable, V any](m map[K]V) iter.Seq2[K, V] {
    return func(yield func(K, V) bool) {
        for k, v := range m {
            if !yield(k, v) {
                return
            }
        }
    }
}

// Usage
for k, v := range Items(myMap) {
    fmt.Println(k, v)
}
```

**Real-World Example: go-crawler (Hypothetical)**

```go
// Lazy page crawling
func (c *Crawler) Pages(ctx context.Context, seed string) iter.Seq[Page] {
    return func(yield func(Page) bool) {
        queue := []string{seed}
        visited := make(map[string]bool)

        for len(queue) > 0 {
            url := queue[0]
            queue = queue[1:]

            if visited[url] {
                continue
            }
            visited[url] = true

            page, err := c.Fetch(ctx, url)
            if err != nil {
                continue
            }

            if !yield(page) {
                return  // Consumer stopped iteration
            }

            // Add links to queue
            queue = append(queue, page.Links...)
        }
    }
}

// Usage: Process pages lazily
for page := range crawler.Pages(ctx, "https://example.com") {
    if page.StatusCode != 200 {
        break  // Stop crawling on error
    }
    process(page)
}
```

**When NOT to use:**

```go
// DON'T: Simple slice iteration
func Items[T any](slice []T) iter.Seq[T] {
    return func(yield func(T) bool) {
        for _, item := range slice {
            if !yield(item) {
                return
            }
        }
    }
}

// Users can just write: for _, item := range slice { ... }
```

---

## Go 1.24 Features

### testing.B.Loop() - Benchmark Loop

**When to use:** All benchmarks (replaces `for i := 0; i < b.N; i++`)

```go
// Before Go 1.24
func BenchmarkOperation(b *testing.B) {
    data := prepareData()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _ = Operation(data)
    }
}

// Go 1.24+
func BenchmarkOperation(b *testing.B) {
    data := prepareData()
    b.ResetTimer()

    for b.Loop() {
        _ = Operation(data)
    }
}
```

**Benefits:**
- Cleaner syntax
- Better integration with benchmark infrastructure
- Potential for future optimizations

### Swiss Tables - Automatic Map Performance

**What changed:** Maps are 10-35% faster automatically

```go
// No code changes needed - automatic performance boost
m := make(map[string]int)
m["key"] = 42

// Benchmarks show improvement:
// Go 1.23: BenchmarkMapAccess-8    50000000    25.3 ns/op
// Go 1.24: BenchmarkMapAccess-8    75000000    16.8 ns/op
```

**Real-World Impact: go-fsm**

```go
// State/event index lookups are faster
func (m *Machine[S, E]) Fire(ctx context.Context, event E) error {
    current, _ := m.store.Get(ctx)
    si := m.stateIndex[current]  // 10-35% faster
    ei := m.eventIndex[event]    // 10-35% faster
    cell := m.table[si][ei]
    // ...
}
```

### crypto/aes.NewGCMWithRandomNonce - Automatic Nonce Management

**When to use:** AES-GCM encryption without manual nonce management

```go
// Before Go 1.24: Manual nonce management
func encrypt(key, plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}

// Go 1.24+: Automatic nonce management
func encrypt(key, plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := aes.NewGCMWithRandomNonce(block)
    if err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nil, plaintext, nil)
    return ciphertext, nil
}
```

---

## Go 1.26 Features

### testing/synctest - Testing Concurrent Code

**When to use:** Testing time-dependent code, concurrent operations

```go
import "testing/synctest"

func TestCache_Expiration(t *testing.T) {
    synctest.Run(func() {
        c := NewCache()
        c.Set("key", []byte("value"), 100*time.Millisecond)

        // Advance time deterministically
        time.Sleep(150 * time.Millisecond)

        _, err := c.Get("key")
        require.ErrorIs(t, err, ErrExpired)
    })
}

func TestConcurrentAccess(t *testing.T) {
    synctest.Run(func() {
        c := NewCache()

        // Spawn concurrent goroutines
        go c.Set("key1", []byte("value1"))
        go c.Set("key2", []byte("value2"))
        go c.Get("key1")
        go c.Get("key2")

        // synctest ensures deterministic execution
    })
}
```

**Benefits:**
- Deterministic concurrent test execution
- No flaky tests due to timing
- Easier to reproduce race conditions

### sync.WaitGroup.Go() - Goroutine Management

**When to use:** Spawning goroutines with automatic WaitGroup management

```go
// Before Go 1.26: Manual WaitGroup
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(it Item) {
        defer wg.Done()
        process(it)
    }(item)
}
wg.Wait()

// Go 1.26+: Automatic WaitGroup
var wg sync.WaitGroup
for _, item := range items {
    wg.Go(func() {
        process(item)
    })
}
wg.Wait()
```

**Real-World Example: Parallel Processing**

```go
func ProcessBatch(items []Item) error {
    var (
        wg   sync.WaitGroup
        mu   sync.Mutex
        errs []error
    )

    for _, item := range items {
        wg.Go(func() {
            if err := Process(item); err != nil {
                mu.Lock()
                errs = append(errs, err)
                mu.Unlock()
            }
        })
    }

    wg.Wait()

    if len(errs) > 0 {
        return errors.Join(errs...)
    }

    return nil
}
```

---

## Features to Avoid

### ❌ sync.Pool - Only Use With Profiling Evidence

**Why avoid:** Premature optimization, adds complexity

```go
// DON'T: Speculative sync.Pool
var userPool = sync.Pool{
    New: func() any {
        return &User{}
    },
}

func GetUser(id string) (*User, error) {
    user := userPool.Get().(*User)
    defer userPool.Put(user)
    // ...
}

// DO: Only add after profiling shows allocation bottleneck
// Example: messageformat-go v1 (proven hot path)
var pool = sync.Pool{
    New: func() any {
        return &formatter{}
    },
}
```

**When to use:**
- Profiling shows allocation bottleneck
- Hot path with proven performance impact
- High-throughput scenarios (>10k ops/sec)

### ❌ Custom Iterators (iter package) - Rarely Needed

**Why avoid:** Users can write their own for loops

```go
// DON'T: Unnecessary iterator
func Items[T any](slice []T) iter.Seq[T] {
    return func(yield func(T) bool) {
        for _, item := range slice {
            if !yield(item) {
                return
            }
        }
    }
}

// Users can just write:
for _, item := range slice {
    // ...
}
```

**When to use:**
- Lazy evaluation (infinite sequences, streaming)
- Complex iteration logic (tree traversal, graph search)
- Resource management (database cursors, file scanning)

### ❌ log/slog in Libraries - Let Callers Control Logging

**Why avoid:** Libraries shouldn't dictate logging strategy

```go
// DON'T: Add log/slog dependency
import "log/slog"

func (c *Cache) Get(key string) ([]byte, error) {
    slog.Info("cache get", "key", key)
    // ...
}

// DO: Return errors, let application code handle logging
func (c *Cache) Get(key string) ([]byte, error) {
    value, err := c.store.Get(key)
    if err != nil {
        return nil, fmt.Errorf("get from store: %w", err)
    }
    return value, nil
}

// Application code decides logging
value, err := cache.Get(key)
if err != nil {
    slog.Error("cache get failed", "key", key, "error", err)
}
```

**Exception:** CLI applications can use log/slog

---

## Feature Adoption Strategy

### Recommended (Use Liberally)

| Feature | Version | Use Case |
|---------|---------|----------|
| `errors.Join()` | 1.20 | Error aggregation |
| `strings.Cut*` | 1.20 | String parsing |
| `min/max/clear` | 1.21 | Value comparison, map/slice clearing |
| `slices` package | 1.21 | Slice operations |
| `maps` package | 1.21 | Map operations |
| `cmp` package | 1.21 | Comparisons |
| `for range N` | 1.22 | Integer loops |
| Loop variable scoping | 1.22 | Automatic (no action needed) |
| `testing.B.Loop()` | 1.24 | Benchmarks |
| Swiss Tables | 1.24 | Automatic (no action needed) |
| `sync.WaitGroup.Go()` | 1.26 | Goroutine management |
| `testing/synctest` | 1.26 | Concurrent testing |

### Use With Caution

| Feature | Version | When to Use |
|---------|---------|-------------|
| `iter.Seq/Seq2` | 1.23 | Lazy evaluation, streaming, complex iteration |
| `sync.Pool` | All | Only with profiling evidence |
| `log/slog` | 1.21 | CLI apps only, not libraries |

---

## Migration Guide

### Migrating to Go 1.21+

```go
// Replace custom helpers with stdlib
- func contains(slice []string, item string) bool { ... }
+ import "slices"
+ slices.Contains(slice, item)

// Replace custom min/max
- func min(a, b int) int { if a < b { return a }; return b }
+ min(a, b)

// Replace custom map clone
- func clone(m map[string]int) map[string]int { ... }
+ import "maps"
+ maps.Clone(m)
```

### Migrating to Go 1.22+

```go
// Replace integer loops
- for i := 0; i < 10; i++ { process() }
+ for range 10 { process() }

// Remove loop variable shadowing
- for _, item := range items {
-     item := item
-     go func() { process(item) }()
- }
+ for _, item := range items {
+     go func() { process(item) }()
+ }
```

### Migrating to Go 1.24+

```go
// Replace benchmark loops
- for i := 0; i < b.N; i++ { ... }
+ for b.Loop() { ... }
```

### Migrating to Go 1.26+

```go
// Replace manual WaitGroup
- var wg sync.WaitGroup
- wg.Add(1)
- go func() {
-     defer wg.Done()
-     process()
- }()
+ var wg sync.WaitGroup
+ wg.Go(func() { process() })

// Use synctest for time-dependent tests
- time.Sleep(100 * time.Millisecond)
+ import "testing/synctest"
+ synctest.Run(func() {
+     time.Sleep(100 * time.Millisecond)
+ })
```

---

## Checklist for Modern Go Features

- [ ] Use `errors.Join()` for error aggregation
- [ ] Use `strings.Cut*` for string parsing
- [ ] Use `slices` package instead of custom helpers
- [ ] Use `maps` package for map operations
- [ ] Use `for range N` for integer loops
- [ ] Use `testing.B.Loop()` in benchmarks
- [ ] Use `sync.WaitGroup.Go()` for goroutine management
- [ ] Use `testing/synctest` for concurrent tests
- [ ] Avoid `sync.Pool` without profiling evidence
- [ ] Avoid custom iterators unless truly needed
- [ ] Avoid `log/slog` in library packages
