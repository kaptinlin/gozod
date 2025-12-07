# Go Module Organization Guide

Module organization is one of the most critical factors affecting codebase maintainability and developer productivity. This guide provides unified, professional, and scalable module organization conventions for Go projects, following Go 1.25+ best practices.

---

## Table of Contents

1. [Core Principles](#1-core-principles)
2. [Project Structure](#2-project-structure)
3. [Package Design](#3-package-design)
4. [Internal Packages](#4-internal-packages)
5. [Package Boundaries](#5-package-boundaries)
6. [Interface Placement](#6-interface-placement)
7. [File Organization](#7-file-organization)
8. [Import Management](#8-import-management)
9. [Dependency Injection](#9-dependency-injection)
10. [Testing Organization](#10-testing-organization)
11. [Anti-Patterns to Avoid](#11-anti-patterns-to-avoid)
12. [Summary](#12-summary)

---

## Quick Reference

### Package Decision Tree

```
Is this code meant to be imported by external projects?
├── Yes → Place in root or dedicated package
└── No  → Is it only used within this module?
          ├── Yes → Place in internal/
          └── No  → Reconsider if it's needed
```

### Standard Project Layout

```
project/
  cmd/           # Entry points (main packages)
  internal/      # Private packages
  pkg/           # (Optional) Public library code
  api/           # API definitions (proto, OpenAPI)
  config/        # Configuration
  scripts/       # Build/automation scripts
  docs/          # Documentation
```

---

# 1. Core Principles

### Every package should have a single, clear purpose

> Good organization = findable in under 1 minute

### Summary of Principles:

* **Single Responsibility**: One package has one primary reason to change
* **Minimal Dependencies**: Packages should have few, explicit dependencies
* **Clear API Surface**: Small, focused public API per package
* **Avoid Circular Dependencies**: Design flows in one direction
* If you struggle to name a package, it probably has too many responsibilities

### Go Module Philosophy

```
A Go module is a collection of packages that are released, versioned,
and distributed together. A well-organized module is one where:
- Each package has a clear purpose
- Dependencies flow in one direction
- Internal details are hidden from consumers
```

---

# 2. Project Structure

### 2.1 Application Project

For services, CLI tools, and applications:

```
myapp/
├── cmd/
│   ├── myapp/              # Main application
│   │   └── main.go
│   └── mytool/             # Additional CLI tool
│       └── main.go
├── internal/
│   ├── config/             # Configuration loading
│   ├── domain/             # Domain models and business logic
│   │   ├── user/
│   │   └── order/
│   ├── repository/         # Data access layer
│   ├── service/            # Business services
│   ├── handler/            # HTTP/gRPC handlers
│   └── middleware/         # HTTP middleware
├── api/                    # API definitions
│   ├── proto/              # Protocol Buffers
│   └── openapi/            # OpenAPI specs
├── migrations/             # Database migrations
├── scripts/                # Build and deployment scripts
├── docs/                   # Documentation
├── go.mod
├── go.sum
└── README.md
```

### 2.2 Library Project

For reusable packages:

```
mylib/
├── reader.go               # Main types and functions
├── reader_test.go          # Tests
├── writer.go               # Additional functionality
├── writer_test.go
├── options.go              # Configuration options
├── errors.go               # Error definitions
├── internal/               # Internal implementation
│   └── buffer/
├── examples/               # Example usage
│   └── basic/
│       └── main.go
├── go.mod
└── README.md
```

### 2.3 Monorepo Project

For multiple related services:

```
monorepo/
├── services/
│   ├── user-service/
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── go.mod
│   └── order-service/
│       ├── cmd/
│       ├── internal/
│       └── go.mod
├── libs/
│   ├── common/             # Shared utilities
│   │   └── go.mod
│   └── auth/               # Shared auth package
│       └── go.mod
├── tools/                  # Development tools
├── scripts/                # Monorepo scripts
└── go.work                 # Go workspace file
```

---

# 3. Package Design

### 3.1 Single Responsibility

Each package should do one thing well:

```go
// Good - focused packages
package user       // User domain logic
package auth       // Authentication
package cache      // Caching utilities
package validate   // Validation helpers

// Bad - unfocused packages
package utils      // What utilities?
package common     // Common to what?
package helpers    // Helps with what?
```

### 3.2 Package Size Guidelines

| Size | Guidance |
|------|----------|
| 1-5 files | Ideal for focused packages |
| 6-15 files | Acceptable for domains |
| 15+ files | Consider splitting |

### 3.3 Package by Domain, Not Layer

Group by business capability, not technical layer:

```
# Good - domain-based
internal/
├── user/
│   ├── service.go      # Business logic
│   ├── repository.go   # Data access
│   ├── handler.go      # HTTP handlers
│   └── types.go        # Domain types
├── order/
│   ├── service.go
│   ├── repository.go
│   └── handler.go
└── notification/
    └── ...

# Avoid - layer-based
internal/
├── services/
│   ├── user.go
│   └── order.go
├── repositories/
│   ├── user.go
│   └── order.go
└── handlers/
    ├── user.go
    └── order.go
```

### 3.4 When to Split Packages

Split when:
- File count exceeds 15 files
- Package has multiple unrelated responsibilities
- You need to mock/test subsets independently
- Names get long to avoid conflicts

```go
// Before: large user package
package user
// user.go, service.go, repository.go, handler.go,
// validation.go, notification.go, export.go...

// After: split by responsibility
package user       // Core domain
package userimport // Import functionality
package userexport // Export functionality
```

---

# 4. Internal Packages

### 4.1 Purpose

The `internal/` directory creates a compiler-enforced boundary:

```
mymodule/
├── internal/
│   └── secret/        # Only importable within mymodule
│       └── secret.go
└── public/
    └── api.go         # Can import internal/secret
```

### 4.2 What Goes in Internal

```
internal/
├── config/            # App configuration
├── domain/            # Business models
├── service/           # Business logic
├── repository/        # Data access
├── handler/           # HTTP/gRPC handlers
├── middleware/        # Request middleware
├── platform/          # Platform abstractions
│   ├── database/
│   ├── cache/
│   └── queue/
└── pkg/               # Internal shared utilities
```

### 4.3 Internal vs Root Packages

| Location | Visibility | Use For |
|----------|------------|---------|
| Root (`/`) | Public | Library APIs |
| `internal/` | Module-private | Implementation details |
| `cmd/` | Executables | Entry points |

---

# 5. Package Boundaries

### 5.1 Dependency Direction

Dependencies should flow in one direction:

```
cmd/          → internal/handler
                    ↓
              → internal/service
                    ↓
              → internal/repository
                    ↓
              → internal/domain (no deps)
```

### 5.2 Avoid Circular Dependencies

```go
// Bad - circular dependency
package user
import "myapp/internal/order"  // user imports order

package order
import "myapp/internal/user"   // order imports user (circular!)

// Good - extract shared types
package domain
type User struct { ... }
type Order struct { ... }

package user
import "myapp/internal/domain"

package order
import "myapp/internal/domain"
```

### 5.3 Package Independence

Each package should be testable in isolation:

```go
// Good - explicit dependencies via interface
type UserService struct {
    repo     UserRepository  // interface
    cache    Cache           // interface
    notifier Notifier        // interface
}

// Bad - hidden dependencies
type UserService struct {
    // Uses global database connection
    // Uses global cache
    // Uses global notification system
}
```

---

# 6. Interface Placement

### 6.1 Define Interfaces at Point of Use

Go best practice: define interfaces where they're **consumed**, not where they're implemented:

```go
// In handler package (consumer)
package handler

type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, input CreateUserInput) (*User, error)
}

type Handler struct {
    users UserService  // depends on interface
}

// In service package (implementation)
package service

type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    return s.repo.FindByID(ctx, id)
}
```

### 6.2 Interface Segregation

Keep interfaces small and focused:

```go
// Good - small, focused interfaces
type Reader interface {
    Read(ctx context.Context, id string) (*Entity, error)
}

type Writer interface {
    Write(ctx context.Context, entity *Entity) error
}

type ReadWriter interface {
    Reader
    Writer
}

// Bad - large interface with many methods
type Repository interface {
    Read(...)
    Write(...)
    Delete(...)
    List(...)
    Search(...)
    Count(...)
    Aggregate(...)
    // ... 20 more methods
}
```

### 6.3 Standard Library Interfaces

Prefer standard library interfaces when possible:

```go
// Use standard interfaces
io.Reader
io.Writer
io.Closer
fmt.Stringer
sort.Interface
encoding.BinaryMarshaler
json.Marshaler
```

---

# 7. File Organization

### 7.1 File Naming

```
package/
├── doc.go              # Package documentation
├── types.go            # Type definitions
├── errors.go           # Error definitions
├── service.go          # Main logic
├── service_test.go     # Tests
├── repository.go       # Data access
├── options.go          # Configuration options
├── helpers.go          # Internal helpers (unexported)
└── export_test.go      # Export internals for testing
```

### 7.2 Single Type Per File (for major types)

```go
// user.go
type User struct { ... }
func NewUser(...) *User { ... }
func (u *User) Validate() error { ... }

// order.go
type Order struct { ... }
func NewOrder(...) *Order { ... }
func (o *Order) Calculate() error { ... }
```

### 7.3 Group Related Small Types

```go
// types.go - group related small types
type Status string

const (
    StatusPending Status = "pending"
    StatusActive  Status = "active"
)

type Config struct {
    Timeout time.Duration
    MaxSize int
}

type Metadata struct {
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 7.4 File Size Guidelines

| Size | Guidance |
|------|----------|
| < 200 lines | Good |
| 200-500 lines | Acceptable |
| 500+ lines | Consider splitting |

---

# 8. Import Management

### 8.1 Import Grouping

Standard Go import organization:

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // Third-party packages
    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"

    // Internal packages
    "myapp/internal/domain"
    "myapp/internal/service"
)
```

### 8.2 Import Aliases

Use aliases to resolve conflicts or improve clarity:

```go
import (
    "context"

    userv1 "myapp/api/user/v1"
    userv2 "myapp/api/user/v2"

    domainUser "myapp/internal/domain/user"
    repoUser "myapp/internal/repository/user"
)
```

### 8.3 Avoid Dot Imports

```go
// Bad - pollutes namespace
import . "myapp/internal/testutil"

// Good - explicit package reference
import "myapp/internal/testutil"
```

### 8.4 Blank Imports

Only for side effects (init functions):

```go
import (
    _ "github.com/lib/pq"           // PostgreSQL driver
    _ "myapp/internal/migrations"   // Run migrations
)
```

---

# 9. Dependency Injection

### 9.1 Constructor Injection

```go
type Service struct {
    repo   Repository
    cache  Cache
    logger *slog.Logger
}

func NewService(repo Repository, cache Cache, logger *slog.Logger) *Service {
    return &Service{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}
```

### 9.2 Functional Options

```go
type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.timeout = d
    }
}

func WithLogger(l *slog.Logger) Option {
    return func(s *Server) {
        s.logger = l
    }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:    addr,
        timeout: 30 * time.Second,  // default
        logger:  slog.Default(),     // default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage
server := NewServer(":8080",
    WithTimeout(60*time.Second),
    WithLogger(customLogger),
)
```

---

# 10. Testing Organization

### 10.1 Test File Placement

Tests live next to the code they test:

```
user/
├── service.go
├── service_test.go      # Unit tests
├── repository.go
├── repository_test.go
└── integration_test.go  # Integration tests (build tag)
```

### 10.2 Testify Usage

Use `github.com/stretchr/testify` for assertions and mocking:

```go
import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
)

// assert vs require:
// - assert: continues test on failure (for multiple checks)
// - require: stops test immediately on failure (for setup/prerequisites)

func TestUserService_GetUser(t *testing.T) {
    // Arrange
    svc := user.NewService(mockRepo)

    // Act
    user, err := svc.GetUser(ctx, "123")

    // Assert
    require.NoError(t, err)           // Stop if error
    assert.Equal(t, "123", user.ID)   // Continue checking
    assert.NotEmpty(t, user.Name)
    assert.True(t, user.IsActive)
}

func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"missing @", "testexample.com", true},
        {"empty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 10.3 Test Suites

Use suites for tests with shared setup/teardown:

```go
package user_test

import (
    "testing"

    "github.com/stretchr/testify/suite"
)

type UserServiceSuite struct {
    suite.Suite
    svc  *user.Service
    repo *MockRepository
}

func (s *UserServiceSuite) SetupTest() {
    // Runs before each test
    s.repo = NewMockRepository()
    s.svc = user.NewService(s.repo)
}

func (s *UserServiceSuite) TearDownTest() {
    // Runs after each test
    s.repo.Reset()
}

func (s *UserServiceSuite) TestGetUser() {
    s.repo.On("FindByID", mock.Anything, "123").Return(&user.User{ID: "123"}, nil)

    user, err := s.svc.GetUser(context.Background(), "123")

    s.Require().NoError(err)
    s.Assert().Equal("123", user.ID)
    s.repo.AssertExpectations(s.T())
}

func (s *UserServiceSuite) TestCreateUser() {
    // Another test with same setup
}

// Run the suite
func TestUserServiceSuite(t *testing.T) {
    suite.Run(t, new(UserServiceSuite))
}
```

### 10.4 Mocking with Testify

```go
package user_test

import (
    "context"

    "github.com/stretchr/testify/mock"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Save(ctx context.Context, user *User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

// Usage in tests
func TestGetUser_NotFound(t *testing.T) {
    repo := new(MockRepository)
    repo.On("FindByID", mock.Anything, "999").Return(nil, ErrNotFound)

    svc := NewService(repo)
    _, err := svc.GetUser(context.Background(), "999")

    assert.ErrorIs(t, err, ErrNotFound)
    repo.AssertExpectations(t)
}
```

### 10.5 Test Packages

```go
// Same package - access unexported
package user

func TestValidateEmail(t *testing.T) {
    // Can test unexported functions
    assert.True(t, validateEmail("test@example.com"))
}

// External package - test public API only
package user_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "myapp/internal/user"
)

func TestUserService(t *testing.T) {
    // Only tests exported API
    svc := user.NewService(mockRepo)
    assert.NotNil(t, svc)
}
```

### 10.6 Test Helpers

```
internal/
├── testutil/           # Shared test utilities
│   ├── fixtures.go     # Test data
│   ├── mocks.go        # Mock implementations
│   └── helpers.go      # Test helpers
└── user/
    ├── service.go
    ├── service_test.go
    └── export_test.go  # Export internals for testing
```

### 10.7 Export Test Pattern

```go
// export_test.go
package user

// Export unexported for testing
var ValidateEmail = validateEmail

// In test file
package user_test

func TestValidateEmail(t *testing.T) {
    result := user.ValidateEmail("test@example.com")
    assert.True(t, result)
}
```

### 10.8 Build Tags for Integration Tests

```go
//go:build integration

package user_test

func TestUserRepository_Integration(t *testing.T) {
    // Requires real database
}
```

```bash
# Run unit tests only
go test ./...

# Run integration tests
go test -tags=integration ./...
```

---

# 11. Anti-Patterns to Avoid

| Anti-Pattern | Problem | Solution |
|--------------|---------|----------|
| `utils` package | Becomes dumping ground | Split by purpose |
| `common` package | Everything is "common" | Extract to domain packages |
| Giant packages | Hard to navigate/test | Split by responsibility |
| Circular imports | Architectural breakdown | Extract shared types |
| Deep nesting (>4 levels) | Hard to navigate | Flatten structure |
| Layer-based organization | Changes touch many packages | Use domain-based |
| Global state | Hard to test | Dependency injection |
| Hidden dependencies | Hard to mock | Explicit interfaces |
| Interface in implementation package | Tight coupling | Define at consumer |
| Over-abstraction | Unnecessary complexity | Abstract when needed |

### Specific Examples

```go
// Anti-pattern: utils dumping ground
package utils
func FormatDate() { ... }
func ValidateEmail() { ... }
func CalculatePrice() { ... }
func SendNotification() { ... }

// Better: purpose-specific packages
package timeutil
func FormatDate() { ... }

package validate
func Email() { ... }

package pricing
func Calculate() { ... }

package notify
func Send() { ... }
```

```go
// Anti-pattern: global database connection
package db
var DB *sql.DB

func init() {
    DB, _ = sql.Open(...)
}

// Better: explicit dependency
func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}
```

---

# 12. Summary

Module organization is architecture made visible.

### The Code Organization Triad

| Element | Purpose |
|---------|---------|
| Package | Define boundaries and responsibilities |
| Interface | Define contracts between packages |
| Directory | Physical organization for navigation |

### Benefits of Good Organization

* **Faster onboarding**: New team members find code quickly
* **Lower refactoring cost**: Changes are localized
* **Easier debugging**: Bug location is predictable
* **Better testability**: Isolated packages are easier to test
* **Higher reusability**: Clear boundaries enable reuse

### Quick Checklist

- [ ] Each package has a single, clear purpose
- [ ] No circular dependencies
- [ ] Internal details in `internal/`
- [ ] Interfaces defined at point of use
- [ ] Dependencies flow in one direction
- [ ] Test files next to source files
- [ ] No `utils` or `common` packages

### The Golden Rule

> Organize so that understanding one package doesn't require understanding the entire codebase.

### Key Commands

```bash
# Check for dependency issues
go mod graph | grep cycle

# Tidy dependencies
go mod tidy

# Verify module
go mod verify

# List all packages
go list ./...

# Test all packages
go test ./...

# Test with race detector
go test -race ./...

# Test with coverage
go test -cover ./...
```
