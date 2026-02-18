# Concurrency Patterns

Modern Go simplifies common concurrency patterns: lazy initialization, goroutine management, timer semantics, and testing concurrent code.

## Contents
- sync.OnceFunc / OnceValue / OnceValues (1.21+)
- sync.WaitGroup.Go (1.25+)
- Timer/Ticker channel changes (1.23+)
- sync/atomic And/Or (1.23+)
- testing/synctest (1.25+)
- unique package for interning (1.23+)

---

## sync.OnceFunc / OnceValue / OnceValues (Go 1.21+)

Simplifies the `sync.Once` + package-level variable pattern for lazy initialization.

### When to use
- Lazy singleton initialization (DB clients, config, HTTP clients)
- Expensive computation that should run at most once
- Replacing `sync.Once` + separate variable pairs

### When NOT to use
- When initialization can fail and you want to retry — `OnceValue` caches the result (including errors) permanently
- When you need to reset/reinitialize — `sync.Once` and `OnceValue` cannot be reset
- In tests where you need different values per test — the cached value persists

```go
// Old
var (
    clientOnce sync.Once
    client     *http.Client
)
func getClient() *http.Client {
    clientOnce.Do(func() {
        client = &http.Client{Timeout: 10 * time.Second}
    })
    return client
}

// New (Go 1.21+)
var getClient = sync.OnceValue(func() *http.Client {
    return &http.Client{Timeout: 10 * time.Second}
})
// Usage: c := getClient()

// OnceValues — returns two values (e.g., value + error)
var loadConfig = sync.OnceValues(func() (*Config, error) {
    return parseConfig("config.yaml")
})
// Usage: cfg, err := loadConfig()
// WARNING: if parseConfig fails, err is cached forever

// OnceFunc — no return value
var initLogger = sync.OnceFunc(func() {
    slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
})
```

**Key detail**: If the function passed to `OnceValue`/`OnceValues` panics, the panic is re-raised on every subsequent call. It does **not** retry.

### Retriable initialization pattern

Neither `sync.Once` nor `OnceValues` supports retry on failure. If you need retriable init, use a custom pattern:

```go
var (
    mu     sync.Mutex
    client *http.Client
)

func getClient() (*http.Client, error) {
    mu.Lock()
    defer mu.Unlock()
    if client != nil {
        return client, nil
    }
    c, err := connectDB()
    if err != nil {
        return nil, err // next call will retry
    }
    client = c
    return client, nil
}
```

---

## sync.WaitGroup.Go (Go 1.25+)

Combines `wg.Add(1)`, `go func()`, and `defer wg.Done()` into a single call.

### When to use
- All cases where you'd use `wg.Add(1); go func() { defer wg.Done(); ... }()`
- Eliminates the common bug of forgetting `Add` or `Done`

### When NOT to use
- When you need per-goroutine error collection — use `errgroup.Group` from `golang.org/x/sync/errgroup` (it has `func(func() error)` signature)
- When you need to pass arguments to the goroutine — `wg.Go` takes `func()` with no args, use closures

```go
// Old
var wg sync.WaitGroup
for _, url := range urls {
    wg.Add(1)
    go func() {
        defer wg.Done()
        fetch(url)
    }()
}
wg.Wait()

// New (Go 1.25+)
var wg sync.WaitGroup
for _, url := range urls {
    wg.Go(func() { fetch(url) })
}
wg.Wait()
```

---

## Timer/Ticker channel changes (Go 1.23+)

`time.Timer` and `time.Ticker` channels are now **unbuffered** (capacity 0). `Stop`/`Reset` guarantee no stale values in the channel. Unstopped timers are GC'd immediately.

### When to use
- Simplify timer/ticker code — remove manual drain patterns
- Remove `defer t.Stop()` for GC purposes (no longer needed for GC, but still good for stopping the timer)

### When NOT to use
- If `go.mod` says `go 1.22` or earlier, old behavior applies (buffered channels)
- Code checking `len(t.C)` or `cap(t.C)` will get 0 — update such checks
- Revert with `GODEBUG=asynctimerchan=1` if needed

```go
// Old — manual drain required after Stop
t := time.NewTimer(d)
if !t.Stop() {
    <-t.C // drain to avoid stale value
}
t.Reset(newD)

// New (Go 1.23+) — Stop/Reset guarantee no stale values
t := time.NewTimer(d)
t.Stop()           // guaranteed clean
t.Reset(newD)      // safe, no drain needed

// Old — must Stop to avoid GC leak
t := time.NewTicker(d)
defer t.Stop() // required to avoid leak

// New (Go 1.23+) — GC collects unstopped tickers
// t.Stop() still good practice but not required for GC

// Old — polling channel length
if len(t.C) == 1 { <-t.C }

// New (Go 1.23+) — non-blocking select instead
select {
case <-t.C:
default:
}
```

---

## sync/atomic And/Or (Go 1.23+)

Atomic bitwise AND and OR operations on integer types.

### When to use
- Setting/clearing specific bits atomically (flags, bitmasks)
- Lock-free flag management

### When NOT to use
- For simple load/store/add — existing atomic operations suffice
- If you need complex multi-field atomic updates — use a mutex

```go
var flags atomic.Uint32

// Set bit 3
flags.Or(1 << 3)

// Clear bit 3
flags.And(^uint32(1 << 3))

// Check if bit 3 is set
if flags.Load()&(1<<3) != 0 { ... }
```

---

## testing/synctest (Go 1.25+)

Test concurrent code with **virtualized time** and **isolated goroutines**. Graduated from experimental in Go 1.25.

### When to use
- Testing timeout behavior without real delays
- Testing concurrent code (goroutine coordination, channels, timers)
- Deterministic tests for inherently non-deterministic concurrency
- Replacing `time.Sleep` in tests with virtualized time

### When NOT to use
- For simple synchronous tests — unnecessary overhead
- When testing actual wall-clock timing behavior (e.g., benchmarks)
- Code blocked on external I/O (network, file system) is **not** "durably blocked" — `synctest.Wait()` will hang. Use `net.Pipe()` instead of real connections.

### How it works

`synctest.Wait()` returns when all goroutines in the "bubble" are **durably blocked** — blocked on channels, `time.Sleep`, `sync.Cond.Wait`, or `sync.WaitGroup.Wait`. Mutex acquisition and real I/O are **not** durably blocking.

```go
func TestDebounce(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        var count atomic.Int32
        debounced := Debounce(func() {
            count.Add(1)
        }, 100*time.Millisecond)

        debounced()
        debounced()
        debounced()

        // Advance virtualized time — executes instantly
        time.Sleep(150 * time.Millisecond)
        synctest.Wait() // wait for all goroutines to settle

        if count.Load() != 1 {
            t.Errorf("expected 1 call, got %d", count.Load())
        }
    })
}

// Use net.Pipe for network testing inside synctest
func TestHTTPClient(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        srvConn, cliConn := net.Pipe() // stays within bubble
        defer srvConn.Close()
        defer cliConn.Close()
        // ... test with fake connections ...
    })
}
```

---

## unique package (Go 1.23+)

Canonical interning of comparable values — saves memory when many identical values exist.

### When to use
- Interning strings (e.g., HTTP headers, repeated keys in large datasets)
- Deduplicating immutable values in memory-constrained scenarios
- Replacing manual intern maps with `sync.Map`

### When NOT to use
- Small numbers of distinct values — overhead isn't worth it
- Mutable values — only use with immutable data
- Short-lived values — interning adds overhead for values quickly discarded

```go
// Intern a string
handle := unique.Make("Content-Type")
s := handle.Value() // "Content-Type"

// Two handles to the same value are equal (pointer comparison)
h1 := unique.Make("foo")
h2 := unique.Make("foo")
fmt.Println(h1 == h2) // true
```

---

## Migration strategy

1. Replace `sync.Once` + variable pairs with `sync.OnceValue` / `sync.OnceFunc`
2. Replace `wg.Add(1); go func() { defer wg.Done(); ... }()` with `wg.Go(func() { ... })` (Go 1.25+)
3. Remove manual timer/ticker drain patterns after upgrading to Go 1.23+
4. Use `synctest.Test` for concurrent test code instead of `time.Sleep`-based synchronization
