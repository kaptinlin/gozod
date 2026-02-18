# Test Patterns

## Table-Driven Tests

The standard pattern for testing multiple inputs/outputs through the same logic.

### Basic pattern

```go
func TestParseSize(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    int64
        wantErr bool
    }{
        {name: "bytes", input: "100B", want: 100},
        {name: "kilobytes", input: "2KB", want: 2048},
        {name: "megabytes", input: "1MB", want: 1_048_576},
        {name: "empty", input: "", wantErr: true},
        {name: "invalid", input: "xyz", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            got, err := ParseSize(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**Rules:**
- Name the slice `tests`, each case `tt`
- Every subtest calls `t.Parallel()`
- Use `name` field for clear test output
- `wantErr bool` for error cases, or `wantErr error` for specific errors

### With specific error matching

```go
tests := []struct {
    name    string
    input   string
    want    *User
    wantErr error  // nil means no error expected
}{
    {name: "valid", input: "alice", want: &User{Name: "alice"}},
    {name: "not found", input: "unknown", wantErr: ErrNotFound},
    {name: "empty", input: "", wantErr: ErrInvalidInput},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        got, err := FindUser(tt.input)
        if tt.wantErr != nil {
            assert.ErrorIs(t, err, tt.wantErr)
            return
        }
        require.NoError(t, err)
        assert.Equal(t, tt.want, got)
    })
}
```

### With setup/teardown per case

```go
tests := []struct {
    name  string
    setup func(t *testing.T) *Store  // per-case setup
    key   string
    want  string
}{
    {
        name: "existing key",
        setup: func(t *testing.T) *Store {
            t.Helper()
            s := NewStore()
            require.NoError(t, s.Set("k", "v"))
            return s
        },
        key:  "k",
        want: "v",
    },
    {
        name:  "empty store",
        setup: func(t *testing.T) *Store { t.Helper(); return NewStore() },
        key:   "k",
        want:  "",
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        store := tt.setup(t)
        got := store.Get(tt.key)
        assert.Equal(t, tt.want, got)
    })
}
```

---

## Subtests

Use `t.Run()` to group related test cases under one test function.

### Grouping by behavior

```go
func TestOrderLifecycle(t *testing.T) {
    t.Parallel()

    t.Run("creation", func(t *testing.T) {
        t.Parallel()
        order := NewOrder("item-1", 2)
        assert.Equal(t, StatusPending, order.Status)
        assert.Equal(t, 2, order.Quantity)
    })

    t.Run("payment", func(t *testing.T) {
        t.Parallel()
        order := NewOrder("item-1", 2)
        err := order.Pay(1000)
        require.NoError(t, err)
        assert.Equal(t, StatusPaid, order.Status)
    })

    t.Run("payment on cancelled order", func(t *testing.T) {
        t.Parallel()
        order := NewOrder("item-1", 2)
        order.Cancel()
        err := order.Pay(1000)
        assert.ErrorIs(t, err, ErrDenied)
    })
}
```

---

## Parallel Testing

### Rules

1. Call `t.Parallel()` on both parent test and subtests
2. Never capture loop variables directly in Go < 1.22 (auto-scoped in 1.22+)
3. Never use `t.Chdir()` in parallel tests (affects whole process)
4. Each parallel subtest must use its own test data — no shared mutable state

```go
func TestValidation(t *testing.T) {
    t.Parallel()

    cases := []struct {
        name  string
        input string
        valid bool
    }{
        {"valid email", "a@b.com", true},
        {"missing @", "abc.com", false},
        {"empty", "", false},
    }

    for _, tt := range cases {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // tt is safe to use directly in Go 1.22+
            got := IsValidEmail(tt.input)
            assert.Equal(t, tt.valid, got)
        })
    }
}
```

---

## Test Helpers

### Pattern

```go
// newTestDB creates a test database and returns cleanup.
func newTestDB(t *testing.T) *DB {
    t.Helper()
    db, err := OpenDB(":memory:")
    require.NoError(t, err)
    t.Cleanup(func() {
        require.NoError(t, db.Close())
    })
    return db
}

// seedTestData populates the database with test fixtures.
func seedTestData(t *testing.T, db *DB) {
    t.Helper()
    for _, u := range testUsers {
        require.NoError(t, db.CreateUser(u))
    }
}
```

**Rules:**
- Always call `t.Helper()` first — makes failure messages point to the calling test, not the helper
- Use `require` (not `assert`) in helpers — if setup fails, the test is meaningless
- Use `t.Cleanup()` for teardown — runs in LIFO order after test completes
- Pass `*testing.T` as first argument

### Shared fixtures with TestMain

```go
var testDB *DB

func TestMain(m *testing.M) {
    var err error
    testDB, err = OpenDB("test.db")
    if err != nil {
        log.Fatal(err)
    }
    code := m.Run()
    testDB.Close()
    os.Exit(code)
}
```

Use `TestMain` only for expensive one-time setup (database, external service). Prefer per-test setup with helpers for isolation.

---

## Error Testing

### Sentinel errors

```go
func TestDelete_NotFound(t *testing.T) {
    t.Parallel()
    store := newTestStore(t)

    err := store.Delete(t.Context(), "nonexistent")
    assert.ErrorIs(t, err, ErrNotFound)
}
```

### Error type matching

```go
func TestParse_InvalidSyntax(t *testing.T) {
    t.Parallel()

    _, err := Parse("{{invalid")

    var parseErr *ParseError
    require.ErrorAs(t, err, &parseErr)
    assert.Equal(t, 1, parseErr.Line)
    assert.Equal(t, 3, parseErr.Column)
}
```

### Error message content (last resort)

```go
// Only when no sentinel error or error type is available
assert.ErrorContains(t, err, "connection refused")
```

Prefer `ErrorIs` > `ErrorAs` > `ErrorContains`. Only check message strings when the error type is opaque.

---

## File Testing

### Temporary directories

```go
func TestWriteConfig(t *testing.T) {
    t.Parallel()
    dir := t.TempDir() // Auto-removed after test

    path := filepath.Join(dir, "config.yaml")
    err := WriteConfig(path, cfg)
    require.NoError(t, err)

    data, err := os.ReadFile(path)
    require.NoError(t, err)
    assert.Contains(t, string(data), "port: 8080")
}
```

### Persistent artifacts (Go 1.26+)

```go
func TestGenerateReport(t *testing.T) {
    report := Generate(data)
    path := filepath.Join(t.ArtifactDir(), "report.html")
    require.NoError(t, os.WriteFile(path, report, 0o644))
    // Persists after test completes for inspection
}
```

### Working directory (non-parallel only)

```go
func TestRelativePaths(t *testing.T) {
    // Do NOT call t.Parallel() — Chdir affects the whole process
    t.Chdir("testdata/project")
    cfg, err := LoadConfig("config.yaml")
    require.NoError(t, err)
    assert.Equal(t, "test-project", cfg.Name)
}
```

### Golden files

```go
func TestRender(t *testing.T) {
    t.Parallel()
    got := Render(input)

    golden := filepath.Join("testdata", t.Name()+".golden")
    if *update {
        require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
    }
    want, err := os.ReadFile(golden)
    require.NoError(t, err)
    assert.Equal(t, string(want), got)
}

var update = flag.Bool("update", false, "update golden files")
```

---

## HTTP Testing

### Handler testing with httptest

```go
func TestGetUser(t *testing.T) {
    t.Parallel()

    // Arrange
    store := newTestStore(t)
    seedTestData(t, store)
    handler := NewUserHandler(store)

    req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
    rec := httptest.NewRecorder()

    // Act
    handler.ServeHTTP(rec, req)

    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

    var user User
    require.NoError(t, json.NewDecoder(rec.Body).Decode(&user))
    assert.Equal(t, "alice", user.Name)
}
```

### Middleware testing

```go
func TestAuthMiddleware(t *testing.T) {
    t.Parallel()

    t.Run("valid token", func(t *testing.T) {
        t.Parallel()
        inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            uid := r.Context().Value(userIDKey).(string)
            assert.Equal(t, "user-123", uid)
            w.WriteHeader(http.StatusOK)
        })

        handler := AuthMiddleware(inner)
        req := httptest.NewRequest(http.MethodGet, "/", nil)
        req.Header.Set("Authorization", "Bearer "+validToken)
        rec := httptest.NewRecorder()

        handler.ServeHTTP(rec, req)
        assert.Equal(t, http.StatusOK, rec.Code)
    })

    t.Run("missing token", func(t *testing.T) {
        t.Parallel()
        inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            t.Fatal("handler should not be called")
        })

        handler := AuthMiddleware(inner)
        req := httptest.NewRequest(http.MethodGet, "/", nil)
        rec := httptest.NewRecorder()

        handler.ServeHTTP(rec, req)
        assert.Equal(t, http.StatusUnauthorized, rec.Code)
    })
}
```

### Integration test with httptest.Server

```go
func TestAPIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    t.Parallel()

    srv := httptest.NewServer(NewRouter(testDeps(t)))
    t.Cleanup(srv.Close)

    resp, err := http.Get(srv.URL + "/health")
    require.NoError(t, err)
    defer resp.Body.Close()
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

---

## Structured Test Output (Go 1.25+)

```go
func TestCriticalPath(t *testing.T) {
    t.Attr("component", "payments")
    t.Attr("priority", "critical")

    // Redirect slog to test output for debugging
    logger := slog.New(slog.NewTextHandler(t.Output(), nil))
    svc := NewPaymentService(WithLogger(logger))

    err := svc.Process(t.Context(), payment)
    require.NoError(t, err)
}
```
