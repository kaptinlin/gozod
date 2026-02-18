# Testing Modernization

Go 1.24-1.26 brought major testing improvements: better benchmarks, test-scoped contexts, directory manipulation, concurrent testing with virtual time, and artifact management.

## Contents
- testing.B.Loop (1.24+, inlining safe since 1.26)
- testing.T.Context / B.Context (1.24+)
- testing.T.Chdir (1.24+)
- testing/synctest (1.25+)
- T.Attr / T.Output (1.25+)
- T.ArtifactDir (1.26+)

---

## testing.B.Loop (Go 1.24+)

Replaces `for i := 0; i < b.N; i++` in benchmarks. Safer, less error-prone, and since Go 1.26, does not prevent inlining in the loop body.

### When to use
- **All new benchmarks** — cleaner and avoids the `b.N` pitfall
- Migrating existing benchmarks — `go fix ./...` (Go 1.26+) can auto-apply

### When NOT to use
- If targeting Go < 1.24

```go
// Old
func BenchmarkProcess(b *testing.B) {
    for i := 0; i < b.N; i++ {
        process()
    }
}

// New (Go 1.24+)
func BenchmarkProcess(b *testing.B) {
    for b.Loop() {
        process()
    }
}

// b.Loop handles setup/teardown correctly
func BenchmarkWithSetup(b *testing.B) {
    data := setupData()
    for b.Loop() {
        process(data)
    }
}
```

**Key detail**: `b.Loop()` keeps values alive (prevents dead-code elimination) — no need for `runtime.KeepAlive` hacks.

---

## testing.T.Context / B.Context (Go 1.24+)

Returns a context that is automatically canceled **after the test completes but before cleanup functions run**.

### When to use
- Replace manual `context.WithCancel` + `t.Cleanup(cancel)` in tests
- Any test function that needs a context

### When NOT to use
- When you need a context with a specific timeout — use `context.WithTimeout(t.Context(), d)`
- When you need a context that outlives the test (rare)

```go
// Old
func TestFoo(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    t.Cleanup(cancel)
    result := doSomething(ctx)
    // ...
}

// New (Go 1.24+)
func TestFoo(t *testing.T) {
    ctx := t.Context()
    result := doSomething(ctx)
    // ...
}

// With timeout
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
    defer cancel()
    result := doSomething(ctx)
    // ...
}
```

---

## testing.T.Chdir (Go 1.24+)

Changes the working directory for the duration of the test. Automatically restored when the test completes.

### When to use
- Tests that need to run in a specific directory (e.g., testing file-relative paths)
- Replace manual `os.Chdir` + defer restore pattern

### When NOT to use
- Parallel tests using `t.Parallel()` — `Chdir` affects the whole process
- When you can pass absolute paths instead

```go
// Old
func TestInDir(t *testing.T) {
    orig, _ := os.Getwd()
    os.Chdir("testdata")
    t.Cleanup(func() { os.Chdir(orig) })
    // ... test ...
}

// New (Go 1.24+)
func TestInDir(t *testing.T) {
    t.Chdir("testdata")
    // working directory is "testdata" until test ends
}
```

---

## testing/synctest (Go 1.25+)

Test concurrent code with virtualized time and isolated goroutines. Deterministic testing for timers, channels, and goroutine coordination.

### When to use
- Testing code with `time.Sleep`, `time.After`, `time.NewTimer`
- Testing debounce, throttle, timeout, retry logic
- Replacing flaky `time.Sleep`-based test synchronization
- Testing goroutine lifecycle management

### When NOT to use
- Simple synchronous tests
- When you need actual wall-clock behavior
- Tests that interact with external systems (network, files)

```go
func TestRetryWithBackoff(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        attempts := 0
        fn := RetryWithBackoff(func() error {
            attempts++
            if attempts < 3 {
                return errors.New("not yet")
            }
            return nil
        }, 100*time.Millisecond)

        var wg sync.WaitGroup
        wg.Go(func() { fn() })

        // Advance virtual time past retries
        time.Sleep(300 * time.Millisecond)
        synctest.Wait()

        wg.Wait()
        if attempts != 3 {
            t.Errorf("expected 3 attempts, got %d", attempts)
        }
    })
}
```

**Key behavior**:
- `time.Sleep` inside `synctest.Test` advances virtual time, not real time
- `synctest.Wait()` blocks until all goroutines in the test are blocked
- Goroutines created inside `synctest.Test` are isolated from outside goroutines

---

## T.Attr / T.Output (Go 1.25+)

`T.Attr(key, value)` emits structured attributes to the test log. `T.Output()` returns an `io.Writer` for the test output stream.

### When to use
- Structured test metadata for CI/CD systems that parse test output
- Redirecting library logging to test output

### When NOT to use
- Simple tests where `t.Log` is sufficient

```go
func TestFeature(t *testing.T) {
    t.Attr("component", "auth")
    t.Attr("priority", "high")

    // Redirect slog to test output
    logger := slog.New(slog.NewTextHandler(t.Output(), nil))
    result := processWithLogger(logger)
    // ...
}
```

---

## T.ArtifactDir (Go 1.26+)

Returns a directory path unique to the test for storing output artifacts (screenshots, logs, traces). Directory persists after test completion.

### When to use
- Tests that produce output files (screenshots, traces, generated reports)
- Replacing manual temp directory creation with `t.TempDir()` when artifacts need to persist

### When NOT to use
- Tests that don't produce files
- Ephemeral files that should be cleaned up — use `t.TempDir()` instead (auto-removed)

```go
func TestRenderReport(t *testing.T) {
    report := generateReport(data)
    outPath := filepath.Join(t.ArtifactDir(), "report.html")
    os.WriteFile(outPath, report, 0o644)
    // artifact persists at testdata/artifacts/TestRenderReport/report.html
}
```

---

## Migration strategy

1. Replace `for i := 0; i < b.N; i++` with `for b.Loop()` — or run `go fix ./...` (Go 1.26+)
2. Replace `ctx, cancel := context.WithCancel(context.Background()); t.Cleanup(cancel)` with `t.Context()`
3. Replace manual `os.Chdir` + restore with `t.Chdir()`
4. Replace `time.Sleep`-based test synchronization with `synctest.Test`
