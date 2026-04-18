---
name: golang-test-framework
description: Go testing framework conventions using github.com/stretchr/testify
---

# Go Testing Framework

## Test Framework

Go uses `github.com/stretchr/testify` for assertions, mocking, and test suites.

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)
```

## Assert vs Require

**`require`** — stops test immediately on failure. Use for preconditions.
**`assert`** — continues on failure. Use for the checks you're actually testing.

```go
func TestUser_Create(t *testing.T) {
    // Preconditions — require (test can't proceed without these)
    user, err := CreateUser("alice", "alice@example.com")
    require.NoError(t, err)
    require.NotNil(t, user)

    // Actual assertions — assert (report all failures)
    assert.Equal(t, "alice", user.Name)
    assert.Equal(t, "alice@example.com", user.Email)
    assert.False(t, user.CreatedAt.IsZero())
}
```

## Common Assertions

| Function | Use |
|----------|-----|
| `assert.Equal(t, expected, actual)` | Value equality |
| `assert.NotEqual(t, unexpected, actual)` | Value inequality |
| `assert.True(t, cond)` / `assert.False(t, cond)` | Boolean checks |
| `assert.Nil(t, obj)` / `assert.NotNil(t, obj)` | Nil checks |
| `assert.NoError(t, err)` / `assert.Error(t, err)` | Error presence |
| `assert.ErrorIs(t, err, target)` | Sentinel error match |
| `assert.ErrorAs(t, err, &target)` | Typed error match |
| `assert.ErrorContains(t, err, "substring")` | Error message substring |
| `assert.Contains(t, str, "sub")` | String/slice contains |
| `assert.ElementsMatch(t, expected, actual)` | Order-independent slice equality |
| `assert.Empty(t, val)` / `assert.NotEmpty(t, val)` | Zero-value / empty check |
| `assert.Len(t, collection, n)` | Collection length |
| `assert.Greater(t, a, b)` / `assert.Less(t, a, b)` | Numeric comparison |
| `assert.WithinDuration(t, expected, actual, delta)` | Time comparison |
| `require.XYZ(...)` | Same functions — stops on failure |

## Table-Driven Tests

Group similar test cases as one table-driven test with `t.Run` subtests:

```go
func TestParseToken_InvalidFormats(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {"empty string", "", ErrInvalid},
        {"no separator", "userdomain", ErrInvalid},
        {"missing domain", "user@", ErrInvalid},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ParseToken(tt.input)
            require.ErrorIs(t, err, tt.wantErr)
        })
    }
}
```

### Parallel Table-Driven Tests

```go
func TestValidate_Inputs(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  bool
    }{
        {"valid", "ok", true},
        {"empty", "", false},
    }

    for _, tt := range tests {
        tt := tt // capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got := Validate(tt.input)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Test Helpers

Use `t.Helper()` so failure messages report the caller's line:

```go
func newTestStore(t *testing.T) *Store {
    t.Helper()
    store, err := NewStore(WithTimeout(time.Second))
    require.NoError(t, err)
    t.Cleanup(func() { store.Close() })
    return store
}
```

## Test Setup / Teardown

### Per-Test Cleanup

```go
func TestFeature(t *testing.T) {
    tmpDir := t.TempDir() // auto-cleaned after test
    store := newTestStore(t)
    t.Cleanup(func() { store.Close() }) // runs after test
    // test code
}
```

### Package-Level Setup (TestMain)

```go
func TestMain(m *testing.M) {
    // global setup
    code := m.Run()
    // global teardown
    os.Exit(code)
}
```

## Testify Suite

For tests sharing expensive setup (DB connections, server instances):

```go
type StoreTestSuite struct {
    suite.Suite
    store *Store
}

func (s *StoreTestSuite) SetupSuite() {
    // runs once before all tests in suite
}

func (s *StoreTestSuite) SetupTest() {
    // runs before each test
    var err error
    s.store, err = NewStore()
    s.Require().NoError(err)
}

func (s *StoreTestSuite) TearDownTest() {
    s.store.Close()
}

func (s *StoreTestSuite) TestGet_ReturnsValue() {
    err := s.store.Set(context.Background(), "key", "value")
    s.Require().NoError(err)

    val, err := s.store.Get(context.Background(), "key")
    s.NoError(err)
    s.Equal("value", val)
}

func (s *StoreTestSuite) TestGet_NotFound_ReturnsError() {
    _, err := s.store.Get(context.Background(), "missing")
    s.ErrorIs(err, ErrNotFound)
}

// Entry point — runs the suite
func TestStoreSuite(t *testing.T) {
    suite.Run(t, new(StoreTestSuite))
}
```

**When to use suites vs standalone tests:**
- Use suites when multiple tests share expensive setup (DB, server, external service)
- Use standalone `func TestX(t *testing.T)` for everything else — simpler, clearer

## Testify Mock

### Define Mock

```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetUser(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) SaveUser(ctx context.Context, user *User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}
```

### Use Mock in Tests

```go
func TestService_GetUser_Success(t *testing.T) {
    repo := new(MockRepository)
    svc := NewService(repo)

    expected := &User{ID: "123", Name: "Alice"}
    repo.On("GetUser", mock.Anything, "123").Return(expected, nil)

    user, err := svc.GetUser(context.Background(), "123")
    require.NoError(t, err)
    assert.Equal(t, expected, user)

    repo.AssertExpectations(t)
}

func TestService_GetUser_NotFound(t *testing.T) {
    repo := new(MockRepository)
    svc := NewService(repo)

    repo.On("GetUser", mock.Anything, "999").Return(nil, ErrNotFound)

    _, err := svc.GetUser(context.Background(), "999")
    assert.ErrorIs(t, err, ErrNotFound)

    repo.AssertExpectations(t)
}
```

### Argument Matchers

```go
// Match any value
repo.On("GetUser", mock.Anything, mock.Anything).Return(user, nil)

// Match by type
repo.On("SaveUser", mock.Anything, mock.AnythingOfType("*User")).Return(nil)

// Custom matcher
repo.On("SaveUser", mock.Anything, mock.MatchedBy(func(u *User) bool {
    return u.Email != "" && u.Name != ""
})).Return(nil)
```

### Call Count Expectations

```go
// Expect exactly once
repo.On("SaveUser", mock.Anything, mock.Anything).Return(nil).Once()

// Expect N times
repo.On("GetUser", mock.Anything, "123").Return(user, nil).Times(3)

// Sequential returns (different value each call)
repo.On("GetUser", mock.Anything, "123").
    Return(&User{Name: "v1"}, nil).Once()
repo.On("GetUser", mock.Anything, "123").
    Return(&User{Name: "v2"}, nil).Once()

// Verify
repo.AssertExpectations(t)              // all On() expectations met
repo.AssertCalled(t, "GetUser", mock.Anything, "123")
repo.AssertNotCalled(t, "DeleteUser", mock.Anything, mock.Anything)
repo.AssertNumberOfCalls(t, "GetUser", 2)
```

### Side Effects with Run

```go
repo.On("SaveUser", mock.Anything, mock.Anything).
    Run(func(args mock.Arguments) {
        user := args.Get(1).(*User)
        user.ID = "generated-id" // mutate argument
    }).
    Return(nil)
```

## Mocking Guidelines

**Mock at system boundaries only:**
- HTTP clients → `httptest.NewServer`
- Database → interface + mock or in-memory implementation
- File system → `t.TempDir()` + real files
- Time → `clockwork` package or pass `time.Time` as parameter
- External APIs → `httptest.NewServer` or interface mock

**Never mock internal packages** — test through exported interfaces.

**Prefer real implementations** when feasible (in-memory DB, temp files). Mocks verify interactions; real implementations verify behavior.

## HTTP Handler Testing

```go
func TestHandler_CreateUser(t *testing.T) {
    svc := new(MockService)
    handler := NewHandler(svc)

    svc.On("CreateUser", mock.Anything, mock.Anything).Return(&User{
        ID: "1", Name: "Alice",
    }, nil)

    body := `{"name":"Alice"}`
    req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
    assert.Contains(t, w.Body.String(), "Alice")
    svc.AssertExpectations(t)
}
```

## Error Testing

```go
// Sentinel error
require.ErrorIs(t, err, ErrNotFound)

// Typed error
var validErr *ValidationError
require.ErrorAs(t, err, &validErr)
assert.Equal(t, "email", validErr.Field)

// Error message
assert.ErrorContains(t, err, "invalid token")

// Wrapped error
err := fmt.Errorf("operation failed: %w", ErrNotFound)
require.ErrorIs(t, err, ErrNotFound) // unwraps and matches
```

## Parallel Tests

```go
func TestFeature_X(t *testing.T) {
    t.Parallel()
    // test code
}
```

**Never use `t.Parallel()`** for tests that:
- Mutate shared state
- Write to the same files
- Use shared singletons

## Interface Compliance Check

```go
var _ MyInterface = (*MyImpl)(nil)
```

## Running Tests

```bash
go test ./...                           # all tests
go test -run TestFeature_HappyPath ./...  # specific test
go test -run TestFeature ./pkg/...        # pattern match in package
go test -race ./...                     # race detector
go test -v ./...                        # verbose
go test -count=1 ./...                  # disable test cache
go test -cover ./...                    # coverage summary
go test -coverprofile=coverage.out ./...  # coverage file
go tool cover -html=coverage.out        # coverage in browser
go test -short ./...                    # skip long tests
```
