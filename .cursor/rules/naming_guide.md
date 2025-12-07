# Go Naming Guide

Naming is one of the most critical factors affecting code readability and maintainability. This guide provides unified, professional, and scalable naming conventions for Go projects, following Go 1.25+ best practices.

---

## Table of Contents

1. [General Principles](#1-general-principles)
2. [Variable Naming](#2-variable-naming)
3. [Function and Method Naming](#3-function-and-method-naming)
4. [Interface Naming](#4-interface-naming)
5. [Struct Naming](#5-struct-naming)
6. [Package Naming](#6-package-naming)
7. [Constant Naming](#7-constant-naming)
8. [Receiver Naming](#8-receiver-naming)
9. [Error Naming](#9-error-naming)
10. [File Naming](#10-file-naming)
11. [Acronyms and Initialisms](#11-acronyms-and-initialisms)
12. [Naming Anti-Patterns](#12-naming-anti-patterns)
13. [Naming Vocabulary](#13-naming-vocabulary)

---

# 1. General Principles

### Go favors **brevity** and **clarity** over verbosity

> Good naming = self-explanatory within context

### Summary of Principles:

* Use **MixedCaps** (exported) or **mixedCaps** (unexported), never underscores
* Shorter names are preferred when scope is limited
* Exported names carry the package name as context
* Names should be obvious from their usage, not their declaration
* Avoid stuttering with package names (`http.HTTPServer` → `http.Server`)
* Difficulty naming often signals a need to split logic

### Go's Exported vs Unexported Convention

```go
// Exported (capital first letter) - visible outside package
type User struct {}
func NewUser() *User {}

// Unexported (lowercase first letter) - package-private
type userCache struct {}
func validateEmail(email string) bool {}
```

---

# 2. Variable Naming

### Format: **mixedCaps** (camelCase)

### 2.1 Scope-Based Length

Go prefers short names in limited scopes:

| Scope | Style | Example |
|-------|-------|---------|
| Loop variable | 1-2 chars | `i`, `j`, `k`, `v`, `ok` |
| Local variable | Short | `user`, `ctx`, `cfg` |
| Parameter | Short, descriptive | `r`, `w`, `id`, `name` |
| Package-level | More descriptive | `defaultTimeout`, `maxRetries` |

### 2.2 Common Short Names

These abbreviated names are idiomatic in Go:

```go
// Standard abbreviations
ctx     // context.Context
cfg     // configuration
err     // error
req     // request
resp    // response
buf     // buffer
ch      // channel
wg      // sync.WaitGroup
mu      // sync.Mutex
rw      // sync.RWMutex
db      // database connection
tx      // transaction
```

### 2.3 Boolean Naming

Booleans should read naturally in `if` statements:

```go
// Good - reads naturally
if enabled { ... }
if valid { ... }
if found { ... }
if ok { ... }

// Acceptable prefixes when clarity needed
isReady      // state check
hasPermission // possession check
canEdit      // capability check
shouldRetry  // conditional action

// Avoid
if isIsReady { ... }  // redundant
if flagEnabled { ... } // "flag" adds nothing
```

### 2.4 Avoid Meaningless Names

| Avoid | Recommended |
|-------|-------------|
| `data` | `users`, `payload`, `body` |
| `info` | `metadata`, `details`, `stats` |
| `temp` | `draft`, `pending`, `buffer` |
| `result` | `user`, `count`, `total` |

---

# 3. Function and Method Naming

### Format: **MixedCaps** (exported) or **mixedCaps** (unexported)

### 3.1 No "Get" Prefix for Getters

Go convention: getters don't use `Get` prefix:

```go
// Good - idiomatic Go
func (u *User) Name() string { return u.name }
func (u *User) Email() string { return u.email }
func (u *User) IsActive() bool { return u.active }

// Avoid - not idiomatic
func (u *User) GetName() string { return u.name }
func (u *User) GetEmail() string { return u.email }
```

### 3.2 Setter Naming

Setters **do** use `Set` prefix:

```go
func (u *User) SetName(name string) { u.name = name }
func (u *User) SetEmail(email string) error { ... }
```

### 3.3 Constructor Naming

Use `New` prefix or return type name:

```go
// Standard constructor
func NewUser(name string) *User { ... }
func NewUserWithOptions(opts Options) *User { ... }

// When package has single main type
// In package "cache"
func New(size int) *Cache { ... }  // cache.New()

// Multiple constructors
func NewReader(r io.Reader) *Reader { ... }
func NewReaderSize(r io.Reader, size int) *Reader { ... }
```

### 3.4 Verb-Object Structure for Actions

```go
// Actions clearly express what they do
ParseConfig()
ValidateEmail()
SendNotification()
GenerateToken()
SerializeJSON()
```

### 3.5 Avoid Abstract Verbs

| Avoid (Vague) | Recommended (Specific) |
|---------------|------------------------|
| `Handle()` | `ProcessRequest()`, `ServeHTTP()` |
| `Do()` | `Execute()`, `Run()`, specific action |
| `Process()` | `Validate()`, `Transform()`, `Parse()` |
| `Manage()` | `Start()`, `Stop()`, `Configure()` |

---

# 4. Interface Naming

### Format: **MixedCaps** (PascalCase)

### 4.1 Single-Method Interfaces: Add `-er` Suffix

```go
// Standard pattern: method name + "er"
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Closer interface { Close() error }
type Stringer interface { String() string }
type Formatter interface { Format(f fmt.State, verb rune) }

// Custom examples
type Validator interface { Validate() error }
type Processor interface { Process(ctx context.Context) error }
type Notifier interface { Notify(event Event) error }
```

### 4.2 Multi-Method Interfaces

Use descriptive noun-based names:

```go
// Combination interfaces
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

// Domain interfaces
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

type AuthService interface {
    Authenticate(ctx context.Context, credentials Credentials) (*Token, error)
    Validate(ctx context.Context, token string) (*Claims, error)
    Revoke(ctx context.Context, token string) error
}
```

### 4.3 Interface Location

Define interfaces where they're **used**, not where they're implemented:

```go
// In consumer package
package handler

type UserStore interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

func NewHandler(store UserStore) *Handler { ... }
```

---

# 5. Struct Naming

### Format: **MixedCaps** (PascalCase for exported)

### 5.1 Clear, Domain-Specific Names

```go
// Good - clear domain terms
type User struct { ... }
type Order struct { ... }
type Repository struct { ... }
type Handler struct { ... }

// With context when needed
type HTTPClient struct { ... }
type UserService struct { ... }
type OrderRepository struct { ... }
```

### 5.2 Options Pattern

```go
// Options struct for configuration
type ServerOptions struct {
    Address     string
    Timeout     time.Duration
    MaxConns    int
    EnableTLS   bool
}

// Functional options pattern
type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}
```

### 5.3 Avoid Redundant Suffixes

| Avoid | Recommended |
|-------|-------------|
| `UserStruct` | `User` |
| `UserModel` | `User` |
| `UserData` | `User` |
| `UserObject` | `User` |

---

# 6. Package Naming

### Format: **lowercase**, single word when possible

### 6.1 Core Rules

```go
// Good - short, lowercase, no underscores
package http
package json
package user
package cache
package config

// Avoid
package httpUtil      // mixed case
package http_util     // underscores
package httputils     // acceptable but verbose
package myPackage     // generic
```

### 6.2 Avoid Stuttering

Package name is part of the qualified name:

```go
// Bad - stutters when used
package http
type HTTPClient struct {}  // http.HTTPClient

// Good - no stutter
package http
type Client struct {}  // http.Client

// Bad - redundant
package user
func NewUser() *User {}  // user.NewUser()

// Good - leverage package name
package user
func New() *User {}  // user.New()
```

### 6.3 Meaningful Package Names

| Avoid | Recommended |
|-------|-------------|
| `utils` | `strings`, `validate`, `convert` |
| `common` | `errors`, `types`, `config` |
| `helpers` | `format`, `parse`, `build` |
| `misc` | Split into specific packages |
| `base` | `core`, `domain`, or specific name |

---

# 7. Constant Naming

### Format: **MixedCaps** (not SCREAMING_SNAKE_CASE)

Go uses MixedCaps for constants, unlike other languages:

```go
// Good - Go convention
const MaxRetryCount = 3
const DefaultTimeout = 30 * time.Second
const APIVersion = "v1"

// Grouped constants
const (
    StatusPending   = "pending"
    StatusActive    = "active"
    StatusCompleted = "completed"
)

// Avoid - not idiomatic Go
const MAX_RETRY_COUNT = 3  // SCREAMING_SNAKE_CASE
const default_timeout = 30 // snake_case
```

### 7.1 Iota for Enumerations

```go
type Status int

const (
    StatusUnknown Status = iota
    StatusPending
    StatusActive
    StatusCompleted
)

type Permission int

const (
    PermRead Permission = 1 << iota
    PermWrite
    PermDelete
    PermAdmin
)
```

---

# 8. Receiver Naming

### Format: **1-2 lowercase letters**, abbreviation of type name

### 8.1 Standard Convention

```go
// Good - short, consistent
func (u *User) Name() string { ... }
func (u *User) SetName(name string) { ... }
func (u *User) Validate() error { ... }

func (s *Server) Start() error { ... }
func (s *Server) Stop() error { ... }

func (c *Client) Do(req *Request) (*Response, error) { ... }

// For longer type names
func (rb *RequestBuilder) Build() *Request { ... }
func (uc *UserController) HandleCreate(w, r) { ... }
```

### 8.2 Consistency Within Type

Always use the same receiver name for a type:

```go
// Good - consistent receiver name
func (u *User) Name() string { ... }
func (u *User) Email() string { ... }
func (u *User) Validate() error { ... }

// Bad - inconsistent
func (u *User) Name() string { ... }
func (user *User) Email() string { ... }  // inconsistent
func (this *User) Validate() error { ... } // not idiomatic
```

### 8.3 Avoid Generic Receivers

| Avoid | Recommended |
|-------|-------------|
| `this` | Type abbreviation |
| `self` | Type abbreviation |
| `me` | Type abbreviation |

---

# 9. Error Naming

### 9.1 Error Variables

```go
// Sentinel errors: Err prefix
var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")
var ErrUnauthorized = errors.New("unauthorized")

// Package-level error
var ErrUserNotFound = errors.New("user: not found")
```

### 9.2 Error Types

```go
// Error types: Error suffix
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
}
```

### 9.3 Error Wrapping

```go
// Use %w for wrappable errors
if err != nil {
    return fmt.Errorf("failed to load user: %w", err)
}

// Check wrapped errors
if errors.Is(err, ErrNotFound) { ... }
if errors.As(err, &validationErr) { ... }
```

---

# 10. File Naming

### Format: **lowercase**, underscores for separation

### 10.1 Standard Pattern

```go
user.go           // Main type/logic
user_test.go      // Tests
user_repository.go // Related component
user_handler.go    // HTTP handler
```

### 10.2 Special Files

```go
doc.go            // Package documentation
*_test.go         // Test files (excluded from build)
*_linux.go        // Platform-specific (build tags)
*_amd64.go        // Architecture-specific
export_test.go    // Export internals for testing
```

### 10.3 Organization

```
package/
  doc.go          # Package documentation
  types.go        # Type definitions
  service.go      # Main logic
  repository.go   # Data access
  handler.go      # HTTP handlers
  service_test.go # Tests
```

---

# 11. Acronyms and Initialisms

### All caps for acronyms

```go
// Good - acronyms in all caps
type HTTPClient struct {}
type URLParser struct {}
type JSONEncoder struct {}
type XMLDecoder struct {}
func ParseURL(raw string) (*URL, error) {}
func SerializeJSON(v any) ([]byte, error) {}

// In middle of name
type userID string      // unexported
type UserID string      // exported
type htmlParser struct {}
type HTMLParser struct {}

// Common acronyms
ID    // identifier
URL   // uniform resource locator
HTTP  // hypertext transfer protocol
JSON  // JavaScript object notation
XML   // extensible markup language
SQL   // structured query language
API   // application programming interface
TLS   // transport layer security
TCP   // transmission control protocol
UDP   // user datagram protocol
```

---

# 12. Naming Anti-Patterns

| Anti-Pattern | Problem |
|--------------|---------|
| `data` / `info` / `stuff` | Meaningless, says nothing about content |
| `tmp` / `temp` | "Temporary" code never stays temporary |
| `util` / `helper` / `common` | Dumping ground packages |
| `GetXxx()` for getters | Not idiomatic Go |
| `this` / `self` receivers | Not idiomatic Go |
| `SCREAMING_CASE` | Not Go convention for constants |
| `snake_case` variables | Go uses mixedCaps |
| Stuttering names | `http.HTTPServer`, `user.NewUser` |
| Type in name | `userSlice`, `countInt` |
| Hungarian notation | `strName`, `iCount` |

---

# 13. Naming Vocabulary

### Data Operations

| Verb | Meaning | Example |
|------|---------|---------|
| `Get` | Retrieve (avoid for simple getters) | `GetByID()` |
| `Find` | Search for item(s) | `FindByEmail()` |
| `List` | Return multiple items | `ListUsers()` |
| `Fetch` | Retrieve from external source | `FetchRemote()` |
| `Load` | Load from storage | `LoadConfig()` |
| `Read` | Read data/stream | `ReadAll()` |

### Create / Modify

| Verb | Meaning | Example |
|------|---------|---------|
| `New` | Constructor | `NewUser()` |
| `Create` | Create and persist | `CreateUser()` |
| `Add` | Add to collection | `AddItem()` |
| `Set` | Set value | `SetName()` |
| `Update` | Modify existing | `UpdateProfile()` |
| `Delete` | Remove permanently | `DeleteUser()` |
| `Remove` | Remove from collection | `RemoveItem()` |

### Processing

| Verb | Meaning | Example |
|------|---------|---------|
| `Parse` | Parse string/data | `ParseConfig()` |
| `Format` | Format for output | `FormatDate()` |
| `Validate` | Check validity | `ValidateEmail()` |
| `Transform` | Transform structure | `TransformResponse()` |
| `Convert` | Type conversion | `ConvertToJSON()` |
| `Encode` | Encode data | `EncodeBase64()` |
| `Decode` | Decode data | `DecodeToken()` |
| `Serialize` | To bytes/string | `SerializeJSON()` |
| `Deserialize` | From bytes/string | `DeserializeJSON()` |

### Lifecycle

| Verb | Meaning | Example |
|------|---------|---------|
| `Init` | Initialize | `InitConfig()` |
| `Start` | Start process | `StartServer()` |
| `Stop` | Stop process | `StopServer()` |
| `Run` | Execute | `RunMigration()` |
| `Close` | Release resources | `Close()` |
| `Shutdown` | Graceful shutdown | `Shutdown()` |
| `Reset` | Reset to initial | `ResetCache()` |

### State Checks (Boolean Returns)

| Pattern | Meaning | Example |
|---------|---------|---------|
| `Is...` | State check | `IsValid()`, `IsEmpty()` |
| `Has...` | Possession check | `HasPermission()` |
| `Can...` | Capability check | `CanEdit()` |
| `Should...` | Conditional | `ShouldRetry()` |

---

# Summary

Go naming is intentionally simple.

### Key Principles

* **Brevity**: Short names for short scopes
* **Clarity**: Names obvious from context
* **Consistency**: Same patterns throughout codebase
* **No stutter**: Package name provides context
* **MixedCaps**: Always, never underscores in names

### The Golden Rule

> A name's length should be proportional to its scope.

### Quick Reference

| Element | Convention | Example |
|---------|------------|---------|
| Package | lowercase | `user`, `http` |
| Exported | MixedCaps | `User`, `NewUser` |
| Unexported | mixedCaps | `userCache`, `validate` |
| Constant | MixedCaps | `MaxRetries`, `DefaultTimeout` |
| Receiver | 1-2 chars | `u`, `s`, `c` |
| Acronym | All caps | `HTTP`, `URL`, `ID` |
| Error var | Err prefix | `ErrNotFound` |
| Error type | Error suffix | `ValidationError` |
| Interface (1 method) | -er suffix | `Reader`, `Writer` |
| File | lowercase_underscore | `user_service.go` |
