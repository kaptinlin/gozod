---
name: testing
description: Write Go tests following best practices with testify and Go 1.25+ features. Covers unit tests, table-driven tests, subtests, parallel testing, mocking, concurrency testing (synctest), benchmarks (b.Loop), HTTP/gRPC testing, and integration testing. Use when writing, reviewing, or improving Go test code, or when the user asks how to test a specific Go scenario.
---

# Go Testing with testify (Go 1.25+)

Write effective Go tests using `github.com/stretchr/testify` and modern Go testing features.

## Quick Reference

| Scenario | Guide |
|----------|-------|
| Unit tests, table-driven, subtests, parallel, helpers | [guides/patterns.md](guides/patterns.md) |
| Interface mocking, hand-written mocks, testify/mock | [guides/mocking.md](guides/mocking.md) |
| Race detection, synctest, goroutine lifecycle | [guides/concurrency.md](guides/concurrency.md) |
| b.Loop(), benchstat, profiling, allocation tracking | [guides/benchmarks.md](guides/benchmarks.md) |

## Core Rules

### assert vs require

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

- **`require`** — stops test immediately on failure. Use for **preconditions and setup**: creating objects, connecting to services, operations that subsequent assertions depend on.
- **`assert`** — continues test on failure. Use for **verifying behavior**: checking return values, state, side effects.

```go
func TestExample(t *testing.T) {
    svc, err := NewService(cfg)
    require.NoError(t, err)       // Stop here if setup fails
    require.NotNil(t, svc)

    result, err := svc.Process(input)
    assert.NoError(t, err)        // Continue to check more assertions
    assert.Equal(t, "expected", result.Name)
    assert.Len(t, result.Items, 3)
}
```

**Rule of thumb:** If the next line would panic or be meaningless without this assertion passing, use `require`. Otherwise use `assert`.

### Test function structure

Every test follows this order:

1. **Arrange** — set up test data and dependencies
2. **Act** — call the function under test
3. **Assert** — verify the result

```go
func TestTransfer(t *testing.T) {
    t.Parallel()

    // Arrange
    from := NewAccount(1000)
    to := NewAccount(0)

    // Act
    err := Transfer(from, to, 500)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, 500, from.Balance())
    assert.Equal(t, 500, to.Balance())
}
```

### Go 1.25+ testing features

| Feature | Version | Usage |
|---------|---------|-------|
| `t.Context()` | 1.24+ | Replace `ctx, cancel := context.WithCancel(context.Background()); t.Cleanup(cancel)` |
| `b.Loop()` | 1.24+ | Replace `for i := 0; i < b.N; i++`. Does not prevent inlining (1.26+) |
| `t.Chdir()` | 1.24+ | Replace manual `os.Chdir` + defer restore |
| `synctest.Test()` | 1.25+ | Deterministic concurrent testing with virtual time |
| `t.Attr()` | 1.25+ | Structured test metadata for CI systems |
| `t.Output()` | 1.25+ | `io.Writer` for redirecting logs to test output |
| `t.ArtifactDir()` | 1.26+ | Persistent artifact directory for test output files |

### Always use `t.Context()` for context

```go
// Do this
func TestWithContext(t *testing.T) {
    ctx := t.Context()
    result, err := svc.Fetch(ctx, id)
    require.NoError(t, err)
}

// With timeout
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
    defer cancel()
    result, err := svc.Fetch(ctx, id)
    require.NoError(t, err)
}
```

## Scenario Router

Determine which testing pattern to use:

**What are you testing?**

1. **A pure function or method** (no I/O, no side effects)
   → Table-driven test with subtests. See [guides/patterns.md](guides/patterns.md).

2. **A function that depends on interfaces** (database, API client, cache)
   → Interface-based mock. See [guides/mocking.md](guides/mocking.md).

3. **Code with goroutines, timers, or channels**
   → `synctest.Test` + race detector. See [guides/concurrency.md](guides/concurrency.md).

4. **An HTTP handler or middleware**
   → `httptest.NewServer` / `httptest.NewRecorder`. See [guides/patterns.md § HTTP Testing](guides/patterns.md).

5. **Performance-critical code**
   → Benchmark with `b.Loop()`. See [guides/benchmarks.md](guides/benchmarks.md).

6. **Error paths and edge cases**
   → Table-driven test with error assertions. See [guides/patterns.md](guides/patterns.md).

7. **Code that reads/writes files**
   → `t.TempDir()` for ephemeral, `t.ArtifactDir()` for persistent. `t.Chdir()` if needed.

## Essential Assertions

| Assertion | When to Use |
|-----------|------------|
| `assert.Equal(t, want, got)` | Value equality (uses `reflect.DeepEqual`) |
| `assert.ErrorIs(t, err, target)` | Error chain contains target |
| `assert.ErrorAs(t, err, &target)` | Error chain contains type |
| `assert.Contains(t, str, substr)` | String/slice/map contains element |
| `assert.Len(t, collection, n)` | Collection has exact length |
| `assert.True(t, cond)` / `False` | Boolean conditions |
| `assert.Nil(t, val)` / `NotNil` | Nil checks |
| `assert.Panics(t, func(){...})` | Function panics |
| `assert.InDelta(t, want, got, delta)` | Float comparison with tolerance |
| `assert.WithinDuration(t, t1, t2, d)` | Time comparison with tolerance |
| `assert.JSONEq(t, wantJSON, gotJSON)` | JSON semantic equality |
| `assert.ElementsMatch(t, want, got)` | Slice equality ignoring order |

**Always add a descriptive message for non-obvious assertions:**

```go
assert.True(t, order.IsPaid(), "order should be marked as paid after successful payment")
```

## Test File Conventions

- Test file: `{source}_test.go` in the same package
- Test function: `Test{Type}_{Scenario}` or `Test{Function}_{Scenario}`
- Testdata: `testdata/` directory (ignored by Go tooling)
- Test helpers: mark with `t.Helper()`, use `require` for setup failures
- Every test calls `t.Parallel()` unless it mutates shared state or uses `t.Chdir()`

```go
// Test helper pattern
func newTestServer(t *testing.T) *Server {
    t.Helper()
    srv, err := NewServer(testConfig())
    require.NoError(t, err)
    t.Cleanup(func() { srv.Close() })
    return srv
}
```

## Running Tests

```bash
# All tests with race detection
go test -race ./...

# Specific package
go test -race ./pkg/auth/

# Specific test function
go test -race -run TestLogin_InvalidPassword ./pkg/auth/

# Specific subtest
go test -race -run "TestLogin/empty_password" ./pkg/auth/

# Verbose output
go test -race -v -run TestLogin ./pkg/auth/

# With coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Short mode (skip slow tests)
go test -short ./...
```
