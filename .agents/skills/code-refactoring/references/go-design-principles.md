# Go Design Principles for Refactoring

SOLID principles and Go design philosophy applied to refactoring decisions. Go 1.26+.

## SOLID in Go

### SRP: Small Packages, Single Purpose

Each package should have one reason to change. If the package doc needs "and", split it.

**Before**
```go
// Package user handles user CRUD, authentication, and email notifications.
package user

type Service struct {
    db     *sql.DB
    mailer *smtp.Client
    hasher PasswordHasher
}

func (s *Service) Create(u User) error          { /* validate + insert + hash password + send welcome email */ }
func (s *Service) Authenticate(email, pw string) (*User, error) { /* lookup + verify hash */ }
func (s *Service) SendReset(email string) error  { /* generate token + send email */ }
func (s *Service) List(filter Filter) ([]User, error) { /* query + paginate */ }
```

**After**
```go
// Package user handles user persistence.
package user

type Repository struct{ db *sql.DB }

func (r *Repository) Create(u User) error                  { /* insert */ }
func (r *Repository) FindByEmail(email string) (*User, error) { /* query */ }
func (r *Repository) List(filter Filter) ([]User, error)   { /* query + paginate */ }

// Package auth handles authentication.
package auth

type Service struct {
    users  *user.Repository
    hasher PasswordHasher
}

func (s *Service) Authenticate(email, pw string) (*User, error) { /* lookup + verify */ }

// Package notify handles notifications.
package notify

type Mailer struct{ client *smtp.Client }

func (m *Mailer) SendWelcome(u user.User) error      { /* send */ }
func (m *Mailer) SendPasswordReset(email, token string) error { /* send */ }
```

**Benefits**: Each package has one reason to change. Testing auth does not require an SMTP client.

### OCP: Interfaces + Strategy

Extend behavior by adding new types, not modifying existing code.

**Before**
```go
func Export(format string, data []Record) ([]byte, error) {
    switch format {
    case "csv":
        return exportCSV(data)
    case "json":
        return exportJSON(data)
    case "xml":
        return exportXML(data)
    default:
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
}
```

**After**
```go
type Exporter interface {
    Export(data []Record) ([]byte, error)
}

var exporters = map[string]Exporter{}

func RegisterExporter(format string, e Exporter) {
    exporters[format] = e
}

func Export(format string, data []Record) ([]byte, error) {
    e, ok := exporters[format]
    if !ok {
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
    return e.Export(data)
}
```

**Benefits**: Adding a new format requires zero changes to existing code -- just register a new `Exporter`.

### LSP: Interface Contracts

Any implementation of an interface must be substitutable without changing program correctness.

**Before**
```go
type Cache interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte, ttl time.Duration) error
    Delete(key string) error
}

// NullCache "implements" Cache but panics on Set.
type NullCache struct{}

func (c NullCache) Get(key string) ([]byte, error)                       { return nil, ErrNotFound }
func (c NullCache) Set(key string, value []byte, ttl time.Duration) error { panic("not supported") }
func (c NullCache) Delete(key string) error                              { return nil }
```

**After**
```go
type CacheReader interface {
    Get(key string) ([]byte, error)
}

type CacheWriter interface {
    Set(key string, value []byte, ttl time.Duration) error
    Delete(key string) error
}

type Cache interface {
    CacheReader
    CacheWriter
}

// NullCache is a valid CacheReader -- no need to implement CacheWriter.
type NullCache struct{}

func (c NullCache) Get(key string) ([]byte, error) { return nil, ErrNotFound }
```

**Benefits**: NullCache satisfies exactly the contract it can fulfill. No panics, no surprises.

### ISP: Small Interfaces (1-3 Methods)

Consumers define the interface they need. Large interfaces force unnecessary coupling.

**Before**
```go
// Defined in the provider package.
type UserStore interface {
    Create(User) error
    Update(User) error
    Delete(id string) error
    Get(id string) (*User, error)
    List(filter Filter) ([]User, error)
    Count(filter Filter) (int, error)
    Search(query string) ([]User, error)
    Export(format string) ([]byte, error)
}

// Consumer only needs Get + List but must accept all 8 methods.
func NewDashboard(store UserStore) *Dashboard { /* uses Get and List only */ }
```

**After**
```go
// Defined in the consumer package -- only what it needs.
type UserGetter interface {
    Get(id string) (*User, error)
}

type UserLister interface {
    List(filter Filter) ([]User, error)
}

func NewDashboard(getter UserGetter, lister UserLister) *Dashboard { /* ... */ }
```

**Benefits**: Dashboard is decoupled from Create/Delete/Export. Any store that implements Get + List works.

### DIP: Accept Interfaces, Return Structs

Depend on abstractions (interfaces) for inputs. Return concrete types so callers can use full API.

**Before**
```go
func NewOrderService(db *sql.DB, cache *redis.Client) *OrderService {
    return &OrderService{db: db, cache: cache}
}

func (s *OrderService) Process(orderID int) error {
    row := s.db.QueryRow("SELECT ...", orderID) // concrete dependency
    // ...
}
```

**After**
```go
type OrderRepository interface {
    Get(id int) (*Order, error)
    Save(order *Order) error
}

type OrderCache interface {
    Get(id int) (*Order, bool)
    Set(order *Order)
}

func NewOrderService(repo OrderRepository, cache OrderCache) *OrderService {
    return &OrderService{repo: repo, cache: cache}
}

func (s *OrderService) Process(orderID int) error {
    if cached, ok := s.cache.Get(orderID); ok {
        return s.fulfill(cached)
    }
    order, err := s.repo.Get(orderID)
    if err != nil {
        return err
    }
    s.cache.Set(order)
    return s.fulfill(order)
}
```

**Benefits**: OrderService is testable with in-memory fakes. No database or Redis required in tests.

## Go Design Philosophy

### Composition Over Inheritance

Go has no inheritance. Use struct embedding and interface composition.

**Before**
```go
type BaseHandler struct {
    logger *slog.Logger
}

type UserHandler struct {
    base BaseHandler // field access: h.base.logger
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.base.logger.Info("handling request")
}
```

**After**
```go
type BaseHandler struct {
    Logger *slog.Logger
}

func (h *BaseHandler) LogRequest(r *http.Request) {
    h.Logger.Info("handling request", "method", r.Method, "path", r.URL.Path)
}

type UserHandler struct {
    BaseHandler // embedded -- methods promoted
    repo UserRepository
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.LogRequest(r) // promoted from BaseHandler
    // handle user logic
}
```

**Benefits**: Promoted methods reduce boilerplate. UserHandler "inherits" LogRequest without wrapper code.

### Make the Zero Value Useful

Design structs so `var x T` works without explicit initialization.

**Before**
```go
type Buffer struct {
    data []byte
}

func NewBuffer() *Buffer {
    return &Buffer{data: make([]byte, 0, 1024)}
}

// Callers MUST call NewBuffer() or get nil panic.
func (b *Buffer) Write(p []byte) {
    b.data = append(b.data, p...)
}
```

**After**
```go
type Buffer struct {
    data []byte
}

// Zero value works: var buf Buffer; buf.Write([]byte("hello"))
func (b *Buffer) Write(p []byte) {
    b.data = append(b.data, p...) // append handles nil slice
}

func (b *Buffer) Bytes() []byte {
    return b.data
}
```

**Benefits**: No constructor required. `var buf Buffer` is immediately usable.

### Errors Are Values

Treat errors as data to inspect, wrap, and route -- not exceptions to catch.

**Before**
```go
func FetchUser(id string) (*User, error) {
    resp, err := http.Get(apiURL + "/users/" + id)
    if err != nil {
        return nil, err // raw error, no context
    }
    if resp.StatusCode == 404 {
        return nil, errors.New("not found") // string comparison needed
    }
    // ...
}

// Caller
user, err := FetchUser(id)
if err != nil {
    if err.Error() == "not found" { // fragile string matching
        // handle not found
    }
}
```

**After**
```go
var ErrUserNotFound = errors.New("user not found")

func FetchUser(id string) (*User, error) {
    resp, err := http.Get(apiURL + "/users/" + id)
    if err != nil {
        return nil, fmt.Errorf("fetch user %s: %w", id, err)
    }
    if resp.StatusCode == 404 {
        return nil, ErrUserNotFound
    }
    // ...
}

// Caller
user, err := FetchUser(id)
if err != nil {
    if errors.Is(err, ErrUserNotFound) { // type-safe check
        // handle not found
    }
}
```

**Benefits**: Sentinel errors enable `errors.Is()` checks. Wrapping with `%w` preserves the error chain.

## Interface Design Patterns

### Consumer-Defined Interfaces

Define interfaces where they are used, not where they are implemented.

**Before**
```go
// provider/storage.go -- provider defines the interface
package provider

type Storage interface {
    Read(key string) ([]byte, error)
    Write(key string, data []byte) error
    Delete(key string) error
    List(prefix string) ([]string, error)
    Watch(prefix string) (<-chan Event, error)
}

// consumer/indexer.go -- consumer imports provider.Storage but only calls Read + List
package consumer

func NewIndexer(s provider.Storage) *Indexer { /* ... */ }
```

**After**
```go
// consumer/indexer.go -- consumer defines its own interface
package consumer

type ReadLister interface {
    Read(key string) ([]byte, error)
    List(prefix string) ([]string, error)
}

func NewIndexer(store ReadLister) *Indexer { /* ... */ }
```

**Benefits**: Consumer is decoupled from the provider package. Any type with Read + List satisfies the interface.

### Compose Interfaces with Embedding

Build larger interfaces from small ones.

**Before**
```go
type ReadWriteCloser interface {
    Read(p []byte) (int, error)
    Write(p []byte) (int, error)
    Close() error
}
```

**After**
```go
type ReadWriteCloser interface {
    io.Reader
    io.Writer
    io.Closer
}
```

**Benefits**: Reuses standard library interfaces. Readers of the code immediately recognize the contract.

### Avoid Returning Interfaces

Return concrete structs. Let callers define the interface they need.

**Before**
```go
type Logger interface {
    Info(msg string, args ...any)
    Error(msg string, args ...any)
}

func NewLogger(w io.Writer) Logger { // returns interface
    return &slogLogger{handler: slog.NewTextHandler(w, nil)}
}
```

**After**
```go
type SlogLogger struct {
    handler slog.Handler
}

func NewLogger(w io.Writer) *SlogLogger { // returns concrete struct
    return &SlogLogger{handler: slog.NewTextHandler(w, nil)}
}

func (l *SlogLogger) Info(msg string, args ...any)  { /* ... */ }
func (l *SlogLogger) Error(msg string, args ...any) { /* ... */ }
func (l *SlogLogger) WithGroup(name string) *SlogLogger { /* ... */ } // callers can use this too
```

**Benefits**: Callers get access to `WithGroup` and other concrete methods. They define their own interface at the call site if they need abstraction.

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
