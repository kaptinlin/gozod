# Testing

Idiomatic Go testing patterns including table-driven tests, proper failure messages, and test helpers improve test reliability and diagnostics.

## Contents
- Table-Driven Tests
- No Assertion Libraries
- Got Before Want in Messages
- Test Helper Conventions
- Scope Setup to Specific Tests
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

## No Assertion Libraries

Go's `testing` package is the only allowed test framework. Don't use third-party assertion libraries. Use `cmp.Diff` for complex comparisons.

**Incorrect:**

```go
assert.IsNotNil(t, "obj", obj)
assert.StringEq(t, "obj.Type", obj.Type, "blogPost")
assert.IntEq(t, "obj.Comments", obj.Comments, 2)
```

**Correct:**

```go
want := BlogPost{Comments: 2, Body: "Hello, world!"}

if !cmp.Equal(got, want) {
    t.Errorf("Blog post = %v, want = %v", got, want)
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
wg.Add(num)
for i := 0; i < num; i++ {
    go func() {
        defer wg.Done()
        if err := engine.Vroom(); err != nil {
            t.Errorf("No vroom left on engine: %v", err)
            return
        }
    }()
}
wg.Wait()
```
