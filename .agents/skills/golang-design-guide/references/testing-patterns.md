# Testing Patterns

This document provides comprehensive testing strategies for Golang libraries with real-world examples.

## Table of Contents

1. [Test Structure](#test-structure)
2. [Table-Driven Tests](#table-driven-tests)
3. [Subtests and Parallel Execution](#subtests-and-parallel-execution)
4. [Benchmark Patterns](#benchmark-patterns)
5. [Test Helpers](#test-helpers)
6. [Mocking and Interfaces](#mocking-and-interfaces)
7. [Coverage Strategy](#coverage-strategy)

---

## Test Structure

**Characteristics:**
- One test file per source file (`foo.go` → `foo_test.go`)
- Test functions start with `Test`
- Use `testing.T` for tests, `testing.B` for benchmarks
- Arrange-Act-Assert pattern

### Basic Pattern

```go
package cache_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/example/cache"
)

func TestCache_Get(t *testing.T) {
    // Arrange
    c := cache.New(cache.NewMemoryStore())
    key := "test-key"
    value := []byte("test-value")
    c.Set(key, value)

    // Act
    result, err := c.Get(key)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, value, result)
}

func TestCache_Get_NotFound(t *testing.T) {
    // Arrange
    c := cache.New(cache.NewMemoryStore())

    // Act
    result, err := c.Get("nonexistent")

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, cache.ErrNotFound)
    assert.Nil(t, result)
}
```

### Real-World Example: go-fsm

```go
func TestMachine_Fire(t *testing.T) {
    // Arrange
    store := fsm.NewMemoryStore(Unpaid)
    m, err := fsm.New[State, Event](Unpaid).
        From(Unpaid).On(PaySuccess).To(Paid).
        From(Paid).On(Ship).To(Shipped).
        Build()
    require.NoError(t, err)

    // Act
    err = m.Fire(context.Background(), PaySuccess, store)

    // Assert
    require.NoError(t, err)
    current, _ := store.Get(context.Background())
    assert.Equal(t, Paid, current)
}
```

---

## Table-Driven Tests

**When to use:** Multiple test cases with similar structure, testing edge cases, parameterized tests

**Characteristics:**
- Slice of test case structs
- Loop over test cases with `t.Run()`
- Clear test case names
- Easy to add new cases

### Basic Pattern

```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   User
        wantErr error
    }{
        {
            name:    "valid user",
            input:   User{Email: "test@example.com", Age: 25},
            wantErr: nil,
        },
        {
            name:    "empty email",
            input:   User{Email: "", Age: 25},
            wantErr: ErrInvalidEmail,
        },
        {
            name:    "negative age",
            input:   User{Email: "test@example.com", Age: -1},
            wantErr: ErrInvalidAge,
        },
        {
            name:    "zero age",
            input:   User{Email: "test@example.com", Age: 0},
            wantErr: ErrInvalidAge,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }

            require.NoError(t, err)
        })
    }
}
```

### Advanced: Table-Driven with Setup/Teardown

```go
func TestCache_Operations(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*Cache)
        op      func(*Cache) error
        verify  func(*testing.T, *Cache)
        wantErr error
    }{
        {
            name: "set and get",
            setup: func(c *Cache) {
                c.Set("key", []byte("value"))
            },
            op: func(c *Cache) error {
                _, err := c.Get("key")
                return err
            },
            verify: func(t *testing.T, c *Cache) {
                val, _ := c.Get("key")
                assert.Equal(t, []byte("value"), val)
            },
            wantErr: nil,
        },
        {
            name:  "get nonexistent",
            setup: func(c *Cache) {},
            op: func(c *Cache) error {
                _, err := c.Get("nonexistent")
                return err
            },
            verify:  func(t *testing.T, c *Cache) {},
            wantErr: ErrNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            c := NewCache()
            tt.setup(c)

            // Act
            err := tt.op(c)

            // Assert
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }

            require.NoError(t, err)
            tt.verify(t, c)
        })
    }
}
```

### Real-World Example: jsonmerge

```go
func TestMerge(t *testing.T) {
    tests := []struct {
        name    string
        target  map[string]any
        patch   map[string]any
        want    map[string]any
        wantErr error
    }{
        {
            name:   "simple merge",
            target: map[string]any{"a": 1, "b": 2},
            patch:  map[string]any{"b": 3, "c": 4},
            want:   map[string]any{"a": 1, "b": 3, "c": 4},
        },
        {
            name:   "null deletes",
            target: map[string]any{"a": 1, "b": 2},
            patch:  map[string]any{"b": nil},
            want:   map[string]any{"a": 1},
        },
        {
            name:   "nested merge",
            target: map[string]any{"a": map[string]any{"x": 1, "y": 2}},
            patch:  map[string]any{"a": map[string]any{"y": 3, "z": 4}},
            want:   map[string]any{"a": map[string]any{"x": 1, "y": 3, "z": 4}},
        },
        {
            name:   "replace non-object",
            target: map[string]any{"a": 1},
            patch:  map[string]any{"a": map[string]any{"x": 1}},
            want:   map[string]any{"a": map[string]any{"x": 1}},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := jsonmerge.Merge(tt.target, tt.patch)

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, result.Doc)
        })
    }
}
```

---

## Subtests and Parallel Execution

**When to use:** Grouping related tests, parallel test execution, test isolation

**Characteristics:**
- Use `t.Run()` for subtests
- Use `t.Parallel()` for parallel execution
- Each subtest has isolated state
- Clear test hierarchy

### Basic Pattern

```go
func TestUser(t *testing.T) {
    t.Run("validation", func(t *testing.T) {
        t.Parallel()

        t.Run("valid email", func(t *testing.T) {
            t.Parallel()
            err := ValidateEmail("test@example.com")
            require.NoError(t, err)
        })

        t.Run("invalid email", func(t *testing.T) {
            t.Parallel()
            err := ValidateEmail("invalid")
            require.Error(t, err)
        })
    })

    t.Run("creation", func(t *testing.T) {
        t.Parallel()

        t.Run("with valid data", func(t *testing.T) {
            t.Parallel()
            user, err := NewUser("test@example.com", 25)
            require.NoError(t, err)
            assert.Equal(t, "test@example.com", user.Email)
        })

        t.Run("with invalid age", func(t *testing.T) {
            t.Parallel()
            _, err := NewUser("test@example.com", -1)
            require.Error(t, err)
        })
    })
}
```

### Real-World Example: go-config

```go
func TestConfig_Load(t *testing.T) {
    t.Run("single source", func(t *testing.T) {
        t.Parallel()

        cfg := config.New[AppConfig](
            file.New("testdata/config.yaml"),
        )

        err := cfg.Load(context.Background())
        require.NoError(t, err)

        val := cfg.Value()
        assert.Equal(t, "production", val.Environment)
    })

    t.Run("multiple sources", func(t *testing.T) {
        t.Parallel()

        cfg := config.New[AppConfig](
            file.New("testdata/base.yaml"),
            file.New("testdata/override.yaml"),
        )

        err := cfg.Load(context.Background())
        require.NoError(t, err)

        val := cfg.Value()
        assert.Equal(t, "override-value", val.Setting)
    })

    t.Run("optional source missing", func(t *testing.T) {
        t.Parallel()

        cfg := config.New[AppConfig](
            file.New("testdata/config.yaml"),
            file.New("testdata/missing.yaml", file.WithPolicy(file.PolicyOptional)),
        )

        err := cfg.Load(context.Background())
        require.NoError(t, err)
    })
}
```

---

## Benchmark Patterns

**When to use:** Performance testing, optimization validation, regression detection

**Characteristics:**
- Benchmark functions start with `Benchmark`
- Use `b.Loop()` (Go 1.24+) or `for i := 0; i < b.N; i++`
- Call `b.ResetTimer()` after setup
- Report allocations with `b.ReportAllocs()`

### Basic Pattern (Go 1.24+)

```go
func BenchmarkOperation(b *testing.B) {
    // Setup
    data := prepareData()
    b.ResetTimer()

    // Benchmark loop
    for b.Loop() {
        _ = Operation(data)
    }
}

func BenchmarkOperationWithAllocs(b *testing.B) {
    data := prepareData()
    b.ResetTimer()
    b.ReportAllocs()

    for b.Loop() {
        _ = Operation(data)
    }
}
```

### Legacy Pattern (Go < 1.24)

```go
func BenchmarkOperation(b *testing.B) {
    data := prepareData()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _ = Operation(data)
    }
}
```

### Advanced: Benchmark with Subtests

```go
func BenchmarkMerge(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            target := generateMap(size)
            patch := generateMap(size / 2)
            b.ResetTimer()
            b.ReportAllocs()

            for b.Loop() {
                _, _ = jsonmerge.Merge(target, patch)
            }
        })
    }
}
```

### Real-World Example: jsonmerge

```go
func BenchmarkMerge(b *testing.B) {
    target := map[string]any{
        "name":  "John",
        "age":   30,
        "email": "john@example.com",
        "address": map[string]any{
            "street": "123 Main St",
            "city":   "New York",
        },
    }

    patch := map[string]any{
        "age": 31,
        "address": map[string]any{
            "city": "San Francisco",
        },
    }

    b.ResetTimer()
    b.ReportAllocs()

    for b.Loop() {
        _, _ = jsonmerge.Merge(target, patch)
    }
}

// Output:
// BenchmarkMerge-8    952150    1357 ns/op    1273 B/op   17 allocs/op
```

### Performance Targets

Document performance targets in benchmarks:

```go
func BenchmarkFSM_Fire(b *testing.B) {
    // Target: < 60 ns/op, 0 allocs for pure transition
    store := fsm.NewMemoryStore(Unpaid)
    m, _ := fsm.New[State, Event](Unpaid).
        From(Unpaid).On(PaySuccess).To(Paid).
        Build()

    ctx := context.Background()
    b.ResetTimer()
    b.ReportAllocs()

    for b.Loop() {
        _ = m.Fire(ctx, PaySuccess, store)
        store.Set(ctx, Unpaid)  // Reset for next iteration
    }
}

// Output:
// BenchmarkFSM_Fire-8    25000000    48.2 ns/op    0 B/op    0 allocs/op
// ✓ Meets target: < 60 ns/op, 0 allocs
```

---

## Test Helpers

**When to use:** Reducing test boilerplate, shared setup/teardown, test utilities

**Characteristics:**
- Helper functions accept `*testing.T`
- Call `t.Helper()` to mark as helper
- Return errors or call `t.Fatal()` directly
- Reusable across test files

### Basic Pattern

```go
func setupCache(t *testing.T) *Cache {
    t.Helper()

    c := NewCache()
    t.Cleanup(func() {
        c.Close()
    })

    return c
}

func assertCacheValue(t *testing.T, c *Cache, key string, want []byte) {
    t.Helper()

    got, err := c.Get(key)
    require.NoError(t, err)
    assert.Equal(t, want, got)
}

// Usage
func TestCache(t *testing.T) {
    c := setupCache(t)
    c.Set("key", []byte("value"))
    assertCacheValue(t, c, "key", []byte("value"))
}
```

### Advanced: Test Fixtures

```go
type fixture struct {
    cache  *Cache
    store  Store
    tmpDir string
}

func newFixture(t *testing.T) *fixture {
    t.Helper()

    tmpDir := t.TempDir()
    store := NewFileStore(tmpDir)
    cache := NewCache(store)

    t.Cleanup(func() {
        cache.Close()
        store.Close()
    })

    return &fixture{
        cache:  cache,
        store:  store,
        tmpDir: tmpDir,
    }
}

func (f *fixture) set(t *testing.T, key string, value []byte) {
    t.Helper()
    err := f.cache.Set(key, value)
    require.NoError(t, err)
}

func (f *fixture) get(t *testing.T, key string) []byte {
    t.Helper()
    value, err := f.cache.Get(key)
    require.NoError(t, err)
    return value
}

// Usage
func TestCacheOperations(t *testing.T) {
    f := newFixture(t)
    f.set(t, "key", []byte("value"))
    got := f.get(t, "key")
    assert.Equal(t, []byte("value"), got)
}
```

### Real-World Example: go-fsm

```go
func newTestMachine(t *testing.T) (*fsm.Machine[State, Event], fsm.StateStore[State]) {
    t.Helper()

    store := fsm.NewMemoryStore(Unpaid)
    m, err := fsm.New[State, Event](Unpaid).
        From(Unpaid).On(PaySuccess).To(Paid).
        From(Paid).On(Ship).To(Shipped).
        From(Shipped).On(Deliver).To(Completed).
        Final(Completed).
        Build()

    require.NoError(t, err)
    return m, store
}

func assertState(t *testing.T, store fsm.StateStore[State], want State) {
    t.Helper()

    got, err := store.Get(context.Background())
    require.NoError(t, err)
    assert.Equal(t, want, got)
}

// Usage
func TestOrderFlow(t *testing.T) {
    m, store := newTestMachine(t)

    m.Fire(context.Background(), PaySuccess, store)
    assertState(t, store, Paid)

    m.Fire(context.Background(), Ship, store)
    assertState(t, store, Shipped)
}
```

---

## Mocking and Interfaces

**When to use:** Testing with external dependencies, isolating units, testing error paths

**Characteristics:**
- Define interfaces for dependencies
- Create mock implementations for tests
- Use table-driven tests with different mocks
- Consider using mockery or gomock for complex mocks

### Basic Pattern

```go
// Interface for dependency
type Store interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte) error
}

// Mock implementation
type mockStore struct {
    data map[string][]byte
    err  error
}

func (m *mockStore) Get(key string) ([]byte, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.data[key], nil
}

func (m *mockStore) Set(key string, value []byte) error {
    if m.err != nil {
        return m.err
    }
    m.data[key] = value
    return nil
}

// Usage in tests
func TestCache_Get_StoreError(t *testing.T) {
    mock := &mockStore{err: errors.New("store error")}
    c := NewCache(mock)

    _, err := c.Get("key")
    require.Error(t, err)
}
```

### Advanced: Mock with Call Tracking

```go
type mockStore struct {
    getCalls []string
    setCalls map[string][]byte
    data     map[string][]byte
    err      error
}

func newMockStore() *mockStore {
    return &mockStore{
        getCalls: make([]string, 0),
        setCalls: make(map[string][]byte),
        data:     make(map[string][]byte),
    }
}

func (m *mockStore) Get(key string) ([]byte, error) {
    m.getCalls = append(m.getCalls, key)
    if m.err != nil {
        return nil, m.err
    }
    return m.data[key], nil
}

func (m *mockStore) Set(key string, value []byte) error {
    m.setCalls[key] = value
    if m.err != nil {
        return m.err
    }
    m.data[key] = value
    return nil
}

// Usage
func TestCache_SingleFlight(t *testing.T) {
    mock := newMockStore()
    c := NewCache(mock, WithSingleFlight())

    // Concurrent gets
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c.Get("key")
        }()
    }
    wg.Wait()

    // Verify only one Get call to store
    assert.Len(t, mock.getCalls, 1)
}
```

### Real-World Example: go-config

```go
type mockSource struct {
    data map[string]any
    err  error
}

func (m *mockSource) Load(ctx context.Context) (map[string]any, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.data, nil
}

func TestConfig_Load_SourceError(t *testing.T) {
    mock := &mockSource{err: errors.New("load error")}
    cfg := config.New[AppConfig](mock)

    err := cfg.Load(context.Background())
    require.Error(t, err)
}

func TestConfig_Load_Merge(t *testing.T) {
    source1 := &mockSource{data: map[string]any{"a": 1, "b": 2}}
    source2 := &mockSource{data: map[string]any{"b": 3, "c": 4}}

    cfg := config.New[map[string]any](source1, source2)
    err := cfg.Load(context.Background())
    require.NoError(t, err)

    val := cfg.Value()
    assert.Equal(t, map[string]any{"a": 1, "b": 3, "c": 4}, val)
}
```

---

## Coverage Strategy

**Characteristics:**
- Set coverage thresholds per package type
- Exclude examples and generated code
- Focus on critical paths
- Use coverage reports to find gaps

### Coverage Thresholds

| Package Type | Target Coverage | Rationale |
|--------------|----------------|-----------|
| Core packages (fsm, command) | 80-90% | Critical business logic |
| Infrastructure (cache, config) | 70-80% | Important but more I/O |
| Data processing (jsonmerge) | 85-95% | Pure functions, easy to test |
| Examples | Excluded | Documentation, not critical |
| Generated code | Excluded | Auto-generated |

### Running Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Check coverage threshold
go test -cover ./... | grep -E 'coverage: [0-9]+\.[0-9]+%'

# Per-package coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Coverage Configuration

```go
// go.mod
module github.com/example/mylib

go 1.26

// Exclude from coverage
//go:build !coverage
// +build !coverage

package examples
```

### Real-World Example: go-fsm

```bash
# Run tests with coverage
task test-coverage

# Output:
# github.com/agentable/go-fsm          coverage: 87.3% of statements
# github.com/agentable/go-fsm/store    coverage: 82.1% of statements
```

---

## Testing Anti-Patterns

### Anti-Pattern 1: Testing Implementation Details

```go
// BAD: Testing internal state
func TestCache_InternalMap(t *testing.T) {
    c := NewCache()
    c.Set("key", []byte("value"))
    assert.Len(t, c.data, 1)  // Testing internal field
}

// GOOD: Testing public API
func TestCache_SetAndGet(t *testing.T) {
    c := NewCache()
    c.Set("key", []byte("value"))
    got, err := c.Get("key")
    require.NoError(t, err)
    assert.Equal(t, []byte("value"), got)
}
```

### Anti-Pattern 2: Flaky Tests

```go
// BAD: Time-dependent test
func TestCache_Expiration(t *testing.T) {
    c := NewCache()
    c.Set("key", []byte("value"), 100*time.Millisecond)
    time.Sleep(150 * time.Millisecond)  // Flaky timing
    _, err := c.Get("key")
    assert.ErrorIs(t, err, ErrExpired)
}

// GOOD: Use fake clock or longer timeouts
func TestCache_Expiration(t *testing.T) {
    clock := clockwork.NewFakeClock()
    c := NewCache(WithClock(clock))
    c.Set("key", []byte("value"), 100*time.Millisecond)
    clock.Advance(150 * time.Millisecond)
    _, err := c.Get("key")
    assert.ErrorIs(t, err, ErrExpired)
}
```

### Anti-Pattern 3: Shared Test State

```go
// BAD: Shared state between tests
var globalCache *Cache

func TestCache_Set(t *testing.T) {
    globalCache.Set("key", []byte("value"))
}

func TestCache_Get(t *testing.T) {
    // Depends on TestCache_Set running first
    val, _ := globalCache.Get("key")
    assert.Equal(t, []byte("value"), val)
}

// GOOD: Isolated test state
func TestCache_Set(t *testing.T) {
    c := NewCache()
    c.Set("key", []byte("value"))
}

func TestCache_Get(t *testing.T) {
    c := NewCache()
    c.Set("key", []byte("value"))
    val, _ := c.Get("key")
    assert.Equal(t, []byte("value"), val)
}
```

---

## Checklist for Testing

- [ ] All public functions have tests
- [ ] Error paths are tested
- [ ] Edge cases are covered (empty input, nil, zero values)
- [ ] Tests use table-driven pattern where appropriate
- [ ] Tests are isolated and can run in parallel
- [ ] Benchmarks exist for performance-critical code
- [ ] Test helpers call `t.Helper()`
- [ ] Coverage meets package threshold
- [ ] No flaky tests (timing, randomness, shared state)
- [ ] Tests document expected behavior clearly
