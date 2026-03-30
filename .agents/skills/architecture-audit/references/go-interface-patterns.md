# Go Interface Patterns

Best practices for interface design in Go 1.26+ following ISP (Interface Segregation Principle).

## Fat Interface Detection

### Automated Detection with staticcheck

```bash
# Install staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# Check for interface issues
staticcheck -checks=ST1016 ./...
```

### Manual Detection

```bash
# Find interfaces with method counts
grep -A 50 "type.*interface" **/*.go | \
  awk '/type.*interface/{name=$2; count=0} /^\s*[A-Z]/{count++} /^}$/{if(count>0) print name, count}'

# Find interfaces with >3 methods
find . -name "*.go" -exec awk '
  /type [A-Z][a-zA-Z0-9_]* interface/ {
    iface=$2; methods=0
  }
  /^\s+[A-Z][a-zA-Z0-9_]*\(/ {
    methods++
  }
  /^}/ {
    if(methods > 3 && iface != "") print FILENAME ":" iface " has " methods " methods"
    iface=""
  }
' {} \;
```

## Single Implementation Detection

### Using gopls

```bash
# Find implementations of an interface
gopls implementation <file>:<line>:<column>
```

### Manual Check

```bash
# Search for types implementing interface methods
# Example: Find implementations of io.Reader
grep -r "func.*Read(.*\[\]byte)" --include="*.go"
```

## ISP Best Practices

### Small Interfaces (1-3 methods)

```go
// ✅ GOOD: Small, focused interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Compose when needed
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

### Fat Interface Anti-Pattern

```go
// ❌ BAD: Fat interface violates ISP
type Repository interface {
    Create(ctx context.Context, item Item) error
    Read(ctx context.Context, id string) (Item, error)
    Update(ctx context.Context, item Item) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]Item, error)
    Search(ctx context.Context, query string) ([]Item, error)
    Count(ctx context.Context) (int, error)
    Exists(ctx context.Context, id string) (bool, error)
}

// ✅ GOOD: Split into focused interfaces
type Creator interface {
    Create(ctx context.Context, item Item) error
}

type Reader interface {
    Read(ctx context.Context, id string) (Item, error)
}

type Updater interface {
    Update(ctx context.Context, item Item) error
}

type Deleter interface {
    Delete(ctx context.Context, id string) error
}

type Lister interface {
    List(ctx context.Context) ([]Item, error)
}

type Searcher interface {
    Search(ctx context.Context, query string) ([]Item, error)
}

// Compose for specific use cases
type ReadWriter interface {
    Reader
    Creator
    Updater
}
```

## Acceptable Exceptions

### Domain Service Boundaries (4-6 methods)

```go
// ⚠️ ACCEPTABLE: Cohesive domain operations
type UserService interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id string) error
    ListUsers(ctx context.Context) ([]*User, error)
}
```

**When acceptable:**
- Methods are cohesive (all operate on same domain entity)
- Represents standard CRUD operations
- Used as adapter/service layer boundary
- Consumers typically need all operations

**When to split:**
- Methods operate on different entities
- Consumers only use subset of methods
- Interface grows beyond 6 methods

## Go 1.26 Generic Interfaces

### Type-Safe Generic Interfaces

```go
// Generic repository interface
type Repository[T any] interface {
    Get(ctx context.Context, id string) (T, error)
    Save(ctx context.Context, item T) error
    Delete(ctx context.Context, id string) error
}

// Concrete implementation
type UserRepository struct{}

func (r *UserRepository) Get(ctx context.Context, id string) (*User, error) {
    // implementation
}

func (r *UserRepository) Save(ctx context.Context, user *User) error {
    // implementation
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
    // implementation
}

// Type-safe usage
var repo Repository[*User] = &UserRepository{}
```

### Constraint Interfaces

```go
// Constraint for comparable types
type Comparable interface {
    comparable
}

// Constraint for ordered types
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 | ~string
}

// Generic function using constraint
func Max[T Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

## Interface Composition Patterns

### Embedding Interfaces

```go
// Standard library pattern
type ReadWriter interface {
    io.Reader
    io.Writer
}

type ReadWriteCloser interface {
    io.Reader
    io.Writer
    io.Closer
}

// Custom composition
type Validator interface {
    Validate() error
}

type Persister interface {
    Save(ctx context.Context) error
}

type ValidatedPersister interface {
    Validator
    Persister
}
```

### Optional Capabilities

```go
// Core interface
type Processor interface {
    Process(data []byte) ([]byte, error)
}

// Optional capability
type Validator interface {
    Validate(data []byte) error
}

// Check for optional capability
func ProcessWithValidation(p Processor, data []byte) ([]byte, error) {
    if v, ok := p.(Validator); ok {
        if err := v.Validate(data); err != nil {
            return nil, err
        }
    }
    return p.Process(data)
}
```

## Thresholds and Guidelines

| Method Count | Assessment | Action |
|--------------|------------|--------|
| 1-3 | ✅ Excellent | ISP compliant, keep as is |
| 4-5 | ⚠️ Acceptable | Review cohesion, consider splitting |
| 6-8 | ⚠️ Warning | Likely violates ISP, plan refactor |
| 9+ | ❌ Violation | Refactor required, split interface |

## Common Mistakes

### Mistake 1: Premature Interface Extraction

```go
// ❌ BAD: Interface with single implementation
type UserGetter interface {
    GetUser(id string) (*User, error)
}

type UserService struct{}

func (s *UserService) GetUser(id string) (*User, error) {
    // only implementation
}

// ✅ GOOD: Use concrete type until second implementation exists
type UserService struct{}

func (s *UserService) GetUser(id string) (*User, error) {
    // implementation
}

// Extract interface when second implementation appears
```

### Mistake 2: Interface Pollution

```go
// ❌ BAD: Unnecessary interface for internal use
type userRepository interface {
    getUser(id string) (*User, error)
}

// ✅ GOOD: Use concrete type for internal code
type userRepository struct{}

func (r *userRepository) getUser(id string) (*User, error) {
    // implementation
}
```

### Mistake 3: God Interface

```go
// ❌ BAD: Interface doing everything
type Service interface {
    // User operations
    GetUser(id string) (*User, error)
    CreateUser(user *User) error

    // Product operations
    GetProduct(id string) (*Product, error)
    CreateProduct(product *Product) error

    // Order operations
    GetOrder(id string) (*Order, error)
    CreateOrder(order *Order) error
}

// ✅ GOOD: Separate interfaces per domain
type UserService interface {
    GetUser(id string) (*User, error)
    CreateUser(user *User) error
}

type ProductService interface {
    GetProduct(id string) (*Product, error)
    CreateProduct(product *Product) error
}

type OrderService interface {
    GetOrder(id string) (*Order, error)
    CreateOrder(order *Order) error
}
```

## Refactoring Strategy

When you find a fat interface:

1. **Identify cohesive groups** of methods
2. **Extract focused interfaces** for each group
3. **Update consumers** to depend on specific interfaces
4. **Compose when needed** using interface embedding
5. **Verify** no consumer depends on the original fat interface
6. **Delete** the fat interface

Example refactoring:

```go
// Before: Fat interface
type Storage interface {
    Read(key string) ([]byte, error)
    Write(key string, data []byte) error
    Delete(key string) error
    List(prefix string) ([]string, error)
    Exists(key string) (bool, error)
    Size(key string) (int64, error)
}

// After: Focused interfaces
type Reader interface {
    Read(key string) ([]byte, error)
}

type Writer interface {
    Write(key string, data []byte) error
}

type Deleter interface {
    Delete(key string) error
}

type Lister interface {
    List(prefix string) ([]string, error)
}

type Metadata interface {
    Exists(key string) (bool, error)
    Size(key string) (int64, error)
}

// Compose for specific use cases
type ReadWriter interface {
    Reader
    Writer
}

type FullStorage interface {
    Reader
    Writer
    Deleter
    Lister
    Metadata
}
```
