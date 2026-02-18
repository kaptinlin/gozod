# Concurrency Testing

## Always Run with Race Detector

```bash
go test -race ./...
```

The race detector catches data races at runtime. It has ~2-10x overhead, so it is for testing, not production. **All CI pipelines must run tests with `-race`.**

---

## testing/synctest (Go 1.25+)

Deterministic testing for concurrent code with **virtual time**. Eliminates flaky `time.Sleep`-based synchronization.

### How it works

- `synctest.Test(t, fn)` runs `fn` in an isolated goroutine group with virtual time
- `time.Sleep` inside `synctest.Test` advances virtual time instantly (no real waiting)
- `synctest.Wait()` blocks until all goroutines in the group are blocked (idle)
- Goroutines created inside the group are isolated from outside goroutines

### Testing retry with backoff

```go
func TestRetryWithBackoff(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        attempts := 0
        fn := func() error {
            attempts++
            if attempts < 3 {
                return errors.New("not yet")
            }
            return nil
        }

        var result error
        var wg sync.WaitGroup
        wg.Go(func() {
            result = RetryWithBackoff(fn, 3, 100*time.Millisecond)
        })

        // Advance virtual time past all retry delays
        time.Sleep(300 * time.Millisecond)
        synctest.Wait()

        wg.Wait()
        assert.NoError(t, result)
        assert.Equal(t, 3, attempts)
    })
}
```

### Testing debounce/throttle

```go
func TestDebounce(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        calls := 0
        debounced := Debounce(func() { calls++ }, 200*time.Millisecond)

        // Rapid calls within debounce window
        debounced()
        debounced()
        debounced()

        // Not enough time — should not fire yet
        time.Sleep(100 * time.Millisecond)
        synctest.Wait()
        assert.Equal(t, 0, calls, "should not fire within debounce window")

        // Enough time — should fire exactly once
        time.Sleep(200 * time.Millisecond)
        synctest.Wait()
        assert.Equal(t, 1, calls, "should fire once after debounce window")
    })
}
```

### Testing timeout handling

```go
func TestWorker_Timeout(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
        defer cancel()

        var result error
        var wg sync.WaitGroup
        wg.Go(func() {
            result = Worker(ctx) // Worker should respect context timeout
        })

        // Advance past timeout
        time.Sleep(6 * time.Second)
        synctest.Wait()

        wg.Wait()
        assert.ErrorIs(t, result, context.DeadlineExceeded)
    })
}
```

### Testing ticker-based loops

```go
func TestMetricsCollector(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        collected := 0
        collector := NewCollector(func() { collected++ }, 1*time.Second)

        var wg sync.WaitGroup
        wg.Go(func() { collector.Run(t.Context()) })

        // Advance 3 tick intervals
        time.Sleep(3 * time.Second)
        synctest.Wait()
        assert.Equal(t, 3, collected)

        // Advance 2 more
        time.Sleep(2 * time.Second)
        synctest.Wait()
        assert.Equal(t, 5, collected)
    })
}
```

### When NOT to use synctest

- Simple synchronous tests
- Tests that need actual wall-clock time (real network calls, file I/O)
- Tests for code that doesn't use timers, channels, or goroutines

---

## Channel Testing

### Testing send/receive with timeout

```go
func TestProducer(t *testing.T) {
    t.Parallel()

    ch := make(chan Event, 10)
    go Produce(t.Context(), ch)

    select {
    case event := <-ch:
        assert.Equal(t, "started", event.Type)
    case <-time.After(1 * time.Second):
        t.Fatal("timed out waiting for event")
    }
}
```

### Testing channel closure

```go
func TestWorkerPool_Shutdown(t *testing.T) {
    t.Parallel()

    pool := NewWorkerPool(4)
    results := pool.Start(t.Context())

    pool.Stop()

    // Drain remaining results and verify channel closes
    var count int
    for range results {
        count++
    }
    // Channel is closed — loop exits
    assert.GreaterOrEqual(t, count, 0)
}
```

---

## Goroutine Leak Detection

Verify that functions clean up goroutines properly:

```go
func TestNoGoroutineLeak(t *testing.T) {
    t.Parallel()

    before := runtime.NumGoroutine()

    ctx, cancel := context.WithCancel(t.Context())
    svc := NewService(ctx)
    svc.Start()

    cancel()
    svc.Wait()

    // Allow goroutines time to exit
    time.Sleep(100 * time.Millisecond)
    after := runtime.NumGoroutine()

    assert.InDelta(t, before, after, 2, "goroutine leak detected")
}
```

For more robust goroutine leak detection, use `go.uber.org/goleak`:

```go
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

---

## Concurrent Access Testing

Test that data structures are safe for concurrent use:

```go
func TestCache_ConcurrentAccess(t *testing.T) {
    t.Parallel()

    cache := NewCache()
    var wg sync.WaitGroup

    // Concurrent writes
    for i := range 100 {
        wg.Go(func() {
            cache.Set(fmt.Sprintf("key-%d", i), i)
        })
    }

    // Concurrent reads
    for i := range 100 {
        wg.Go(func() {
            cache.Get(fmt.Sprintf("key-%d", i))
        })
    }

    wg.Wait()
    // If there's a data race, `go test -race` will catch it
}
```

---

## sync.WaitGroup.Go (Go 1.25+)

Replace the `wg.Add(1); go func() { defer wg.Done(); ... }()` boilerplate:

```go
// Old
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()
wg.Wait()

// New (Go 1.25+)
var wg sync.WaitGroup
wg.Go(func() { doWork() })
wg.Wait()
```

Use `wg.Go` in both production code and tests for cleaner goroutine spawning.
