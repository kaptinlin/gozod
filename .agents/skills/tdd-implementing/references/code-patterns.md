---
name: golang-code-patterns
description: Go-specific code patterns for TDD implementation
---

# Go Code Patterns for TDD

## Constructor Pattern

### Functional Options (Preferred)

```go
type Option func(*Store)

func WithTimeout(d time.Duration) Option {
    return func(s *Store) { s.timeout = d }
}

func NewStore(opts ...Option) (*Store, error) {
    s := &Store{timeout: 30 * time.Second}
    for _, opt := range opts {
        opt(s)
    }
    return s, nil
}
```

### Config Struct

```go
type Config struct {
    Timeout    time.Duration
    MaxRetries int
}

func NewStore(cfg Config) (*Store, error) {
    if cfg.Timeout == 0 {
        return nil, ErrInvalidConfig
    }
    return &Store{cfg: cfg}, nil
}
```

## Error Patterns

### Sentinel Errors

```go
var (
    ErrInvalid       = errors.New("invalid input")
    ErrNotFound      = errors.New("not found")
    ErrAlreadyClosed = errors.New("already closed")
)
```

### Custom Error Types

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

### Error Wrapping

```go
if err != nil {
    return fmt.Errorf("parse token: %w", err)
}
```

## Dependency Injection

### Accept Interfaces, Return Structs

```go
// Interface — small, behavior-focused
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}

// Constructor accepts interface
func NewService(repo UserRepository, logger Logger) *Service {
    return &Service{repo: repo, logger: logger}
}
```

### Testify Mock Implementation

```go
type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepo) Save(ctx context.Context, user *User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}
```

**Pattern for nil-safe Get(0):** Always guard `args.Get(0)` with a nil check before type assertion — avoids panic when mock returns `nil`.

## Concurrency Patterns

### Atomic Flag for Closed State

```go
type Store struct {
    closed atomic.Bool
    mu     sync.RWMutex
}

func (s *Store) Close() error {
    if !s.closed.CompareAndSwap(false, true) {
        return ErrAlreadyClosed
    }
    // cleanup
    return nil
}
```

### Mutex for State Protection

```go
type Store struct {
    mu sync.RWMutex
    m  map[string]string
}

func (s *Store) Get(key string) (string, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    v, ok := s.m[key]
    return v, ok
}

func (s *Store) Set(key, value string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.m[key] = value
}
```

### sync.Once for Setup

```go
type Client struct {
    once sync.Once
    conn *Connection
}

func (c *Client) ensureInit() {
    c.once.Do(func() {
        c.conn = dial()
    })
}
```

## Context Pattern

```go
func (s *Store) Get(ctx context.Context, key string) (string, error) {
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
        // do work
    }
}
```

## Interface Implementation

### Compile-Time Check

```go
var _ StoreInterface = (*Store)(nil)
```

### Minimal Interface

```go
type Storer interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string) error
}
```

## Sub-Module Setup

When creating a new sub-module (e.g., `store/awsstore/`):

```go
// go.mod in store/awsstore/
module github.com/user/project/store/awsstore

go 1.22

require (
    github.com/aws/aws-sdk-go-v2 v1.x.x
)

// In parent go.mod
replace github.com/user/project/store/awsstore => ./awsstore
```

## Naming Conventions

- **Package names**: lowercase, single word, no underscores
- **Exported**: PascalCase
- **Unexported**: camelCase
- **Interface names**: -er suffix (Reader, Writer, Storer)
- **Error variables**: Err prefix (ErrNotFound, ErrInvalid)
- **Mock structs**: Mock prefix (MockRepository, MockService)

## Test File Organization

```
package mypackage

// mypackage.go         — production code
// mypackage_test.go    — tests (same package, access unexported)
// export_test.go       — export unexported for external tests
```

## Resource Cleanup

### defer for Cleanup

```go
func (s *Store) Process() error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    return s.doWork(ctx)
}
```

### t.Cleanup for Tests

```go
func newTestServer(t *testing.T) *httptest.Server {
    t.Helper()
    srv := httptest.NewServer(handler)
    t.Cleanup(srv.Close)
    return srv
}
```

## Go Module Commands

```bash
go mod init github.com/user/project
go mod tidy
go get github.com/stretchr/testify@latest
go get -u ./...
go mod verify
```
