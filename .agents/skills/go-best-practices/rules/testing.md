# Testing

Idiomatic Go testing patterns including table-driven tests, proper failure messages, and test helpers improve test reliability and diagnostics.

## Contents
- Table-Driven Tests
- Assertion Style and `testify`
- Got Before Want in Messages
- Test Helper Conventions
- Scope Setup to Specific Tests
- Modern Test APIs
- Test Error Semantics
- Don't Call t.Fatal from Goroutines

---

## Table-Driven Tests

When many test cases share similar logic, use table-driven tests. Use field names in struct literals. Add test descriptions — never identify by index alone.

**Incorrect:**

```go
for i, d := range tests {
    if strings.ToUpper(d.input) != d.want {
        t.Errorf("Failed on case #%d", i)
    }
}
```

**Correct:**

```go
func TestCompare(t *testing.T) {
    tests := []struct {
        a, b string
        want int
    }{
        {a: "", b: "", want: 0},
        {a: "a", b: "", want: 1},
        {a: "", b: "a", want: -1},
        {a: "abc", b: "abc", want: 0},
    }

    for _, test := range tests {
        got := Compare(test.a, test.b)
        if got != test.want {
            t.Errorf("Compare(%q, %q) = %v, want %v", test.a, test.b, got, test.want)
        }
    }
}
```

Use `t.Run` for subtests when individual cases need setup/teardown. Subtest names should be readable and usable with `-run` flag filtering.

---

## Assertion Style and `testify`

`testing` is still the core test framework. Prefer plain `testing` plus `cmp.Diff` for most new tests. `testify/assert` and `testify/require` are acceptable when the repository already standardizes on them or when they make the test materially clearer.

Rules:
- Use one assertion style per package or file. Don't mix ad hoc styles.
- Use `require` for preconditions and setup gates after which the rest of the test is meaningless.
- Use `assert` for additional checks when the test can continue.
- Use `cmp.Diff` or `protocmp.Transform()` when a structural diff is more useful than a boolean assertion.
- Avoid `testify/suite` by default. Subtests, helpers, and explicit setup are usually clearer in Go.

**Plain `testing` is still good default:**

```go
want := BlogPost{Comments: 2, Body: "Hello, world!"}
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("Blog post mismatch (-want +got):\n%s", diff)
}
```

**`testify` is acceptable when it improves flow:**

```go
func TestLoadUser(t *testing.T) {
    user, err := LoadUser("alice")
    require.NoError(t, err)
    require.NotNil(t, user)

    assert.Equal(t, "alice", user.Name)
    assert.True(t, user.Active)
}
```

For proto comparison, use `protocmp.Transform()`:

```go
if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
    t.Errorf("unexpected diff (-want +got):\n%s", diff)
}
```

---

## Got Before Want in Messages

Use the format `YourFunc(%v) = %v, want %v`. Always identify the function name and inputs. Put actual result (got) before expected (want). Use "got"/"want" not "actual"/"expected".

**Incorrect:**

```go
t.Errorf("got %v, want %v", result, expected)
```

**Correct:**

```go
t.Errorf("Compare(%q, %q) = %v, want %v", test.a, test.b, got, test.want)
```

For complex diffs, always label the direction:

```go
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("Foo() returned unexpected difference (-want +got):\n%s", diff)
}
```

Prefer `t.Error` over `t.Fatal` to continue testing and report all failures. Use `t.Fatal` only when subsequent tests would be meaningless.

---

## Test Helper Conventions

Test helpers that call `t.Fatal` should call `t.Helper()` first. Prefix must-succeed helpers with `must`. Prefer `t.Fatal` over returning errors in setup helpers.

**Incorrect (returns error from setup helper):**

```go
func addGameAssets(t *testing.T, dir string) error {
    t.Helper()
    if err := os.WriteFile(path.Join(dir, "pak0.pak"), pak0, 0644); err != nil {
        return err
    }
    return nil
}
```

**Correct (fails directly with descriptive message):**

```go
func mustAddGameAssets(t *testing.T, dir string) {
    t.Helper()
    if err := os.WriteFile(path.Join(dir, "pak0.pak"), pak0, 0644); err != nil {
        t.Fatalf("Setup failed: could not write pak0 asset: %v", err)
    }
}
```

---

## Scope Setup to Specific Tests

Call setup functions explicitly in each test that needs them. Don't use package-level `init()` or `var` for test setup.

**Incorrect:**

```go
var dataset []byte

func init() {
    dataset = mustLoadDataset()
}
```

**Correct:**

```go
func TestParseData(t *testing.T) {
    data := mustLoadDataset(t)
    parsed, err := ParseData(data)
    if err != nil {
        t.Fatalf("Unexpected error parsing data: %v", err)
    }
    // ...
}

func TestRegression682831(t *testing.T) {
    // This test doesn't need the dataset
    if got, want := guessOS("zpc79.example.com"), "grhat"; got != want {
        t.Errorf(`guessOS("zpc79.example.com") = %q, want %q`, got, want)
    }
}
```

Use `sync.Once` for expensive setup shared by some (not all) tests. Use `TestMain` only when all tests need shared setup that requires teardown.

---

## Modern Test APIs

When the module Go version allows it, prefer the newer `testing` APIs that remove boilerplate:

- `for b.Loop()` instead of `for i := 0; i < b.N; i++` in benchmarks (Go 1.24+)
- `t.Context()` instead of manual `context.WithCancel` plus `t.Cleanup(cancel)` (Go 1.24+)
- `t.Chdir()` instead of manual `os.Chdir` plus restore logic (Go 1.24+)
- `testing/synctest` for time-based and concurrent tests that would otherwise need flaky `time.Sleep` coordination (Go 1.25+)

**Benchmark modernization:**

```go
func BenchmarkProcess(b *testing.B) {
    for b.Loop() {
        process()
    }
}
```

**Context modernization:**

```go
func TestWorker(t *testing.T) {
    ctx := t.Context()
    if err := RunWorker(ctx); err != nil {
        t.Fatalf("RunWorker() error = %v", err)
    }
}
```

**Concurrent test modernization:**

```go
func TestDebounce(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        var count atomic.Int32
        fn := Debounce(func() { count.Add(1) }, 100*time.Millisecond)

        fn()
        fn()
        time.Sleep(150 * time.Millisecond)
        synctest.Wait()

        if got := count.Load(); got != 1 {
            t.Errorf("count.Load() = %d, want 1", got)
        }
    })
}
```

---

## Test Error Semantics

Don't match error strings — they're change detectors. Use `errors.Is` for sentinel errors, or check only that err is non-nil when the specific type doesn't matter.

**Incorrect:**

```go
if !strings.Contains(err.Error(), "not found") {
    t.Errorf("expected not found error")
}
```

**Correct:**

```go
if err == nil {
    t.Fatal("expected error, got nil")
}

if !errors.Is(err, fs.ErrNotExist) {
    t.Errorf("got %v, want fs.ErrNotExist", err)
}
```

---

## Don't Call t.Fatal from Goroutines

Never call `t.Fatal`, `t.Fatalf`, or `t.FailNow` from a goroutine other than the one running the Test function. Use `t.Error` instead.
The same rule applies to `require.*`, because it eventually calls `FailNow`.

**Incorrect:**

```go
go func() {
    if err := engine.Vroom(); err != nil {
        t.Fatalf("No vroom: %v", err) // WRONG: not test goroutine
    }
}()
```

**Correct:**

```go
var wg sync.WaitGroup
for i := 0; i < num; i++ {
    wg.Go(func() { // Go 1.25+
        if err := engine.Vroom(); err != nil {
            t.Errorf("engine.Vroom() error = %v", err)
        }
    })
}
wg.Wait()
```
