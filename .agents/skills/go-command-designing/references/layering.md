# Go CLI Application Layering

Clean architecture for Go CLI applications using go-command framework.

## Layer Overview

```
cmd/app/
├── main.go           # Entry point
├── commands/         # Command definitions (thin)
└── services/         # Service layer (coordination)

pkg/                  # Domain layer (business logic)
├── domain/           # Domain models
├── repository/       # Data access interfaces
└── usecase/          # Business use cases

internal/             # Internal implementations
├── adapter/          # External service adapters
└── infrastructure/   # Infrastructure code
```

## Layer Responsibilities

### Commands Layer (`cmd/app/commands/`)

**Purpose:** Define CLI interface, parse arguments, delegate to services.

**Responsibilities:**
- Command structure and options
- Argument validation
- Call services
- Format output using `command.Respond()`

**Rules:**
- ✅ Thin handlers (<20 lines)
- ✅ No business logic
- ✅ Create services in handlers
- ✅ Use typed config structs with `ctx.Bind()`
- ✅ Validate after bind using `gozod.FromStruct[T]()`
- ❌ No domain types
- ❌ No database access
- ❌ No external API calls
- ❌ No `map[string]any` for config

**Example:**
```go
type CreateConfig struct {
    Username string `flag:"username" validate:"required,min=3"`
    Email    string `flag:"email" validate:"required,email"`
    Admin    bool   `flag:"admin"`
}

func Command() *command.Command {
    return &command.Command{
        Name: "user",
        Commands: []*command.Command{
            {
                Name: "create",
                Options: []command.Option{
                    {Name: "username", Required: true},
                    {Name: "email", Required: true},
                    {Name: "admin", IsBool: true},
                },
                Handler: func(ctx *command.Context) error {
                    var cfg CreateConfig

                    // Bind flags to struct
                    if err := ctx.Bind(&cfg); err != nil {
                        return err
                    }

                    // Validate using gozod
                    schema := gozod.FromStruct[CreateConfig](gozod.WithTagName("validate"))
                    if _, err := schema.Parse(cfg); err != nil {
                        return err
                    }

                    // Delegate to service
                    svc := services.NewUserService()
                    user, err := svc.Create(ctx.Context(), cfg.Username, cfg.Email, cfg.Admin)
                    if err != nil {
                        return err
                    }
                    return command.Respond(ctx, user)
                },
            },
        },
    }
}
```

### Services Layer (`cmd/app/services/`)

**Purpose:** Coordinate between commands and domain logic, provide CLI-specific DTOs when needed.

**Responsibilities:**
- Orchestrate domain operations
- Convert domain types to CLI DTOs (only when necessary)
- Handle CLI-specific concerns
- Nil receiver checks

**Rules:**
- ✅ Zero or single dependency constructors
- ✅ All methods check `if s == nil`
- ✅ Context-first signatures: `(ctx context.Context, ...params) (result, error)`
- ✅ Return domain types directly when CLI can use them
- ✅ Return typed DTOs only when combining/transforming data
- ❌ No business logic (delegate to pkg/)
- ❌ No 1:1 DTO mappings (use domain types directly)

**DTO Decision Tree:**

When should you create a DTO in services layer?

✅ **Create DTO when:**
1. Combining multiple domain types
2. Avoiding `map[string]any`
3. Customizing display format for CLI

❌ **Don't create DTO when:**
1. 1:1 field mapping with domain type
2. CLI can directly use domain type
3. Service is just a simple passthrough

**Example:**
```go
type UserService struct {
    repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

// ✅ Good: Combining multiple fields for CLI display
type CacheStatusDTO struct {
    FileKey      string    // Extracted from filename
    DBPath       string    // Composed path
    LastSyncedAt time.Time // From file info
    SizeBytes    int64     // From file info
}

// ✅ Good: Avoiding map[string]any
type ColorEntry struct {
    Value string `json:"value"`
    Count int    `json:"count"`
}

// ✅ Good: Customizing display format
type UserListEntry struct {
    ID       string  // Only show ID
    Username string  // Only show username
    Status   string  // Only show status
}
// Instead of returning full user.User with 20+ fields

// ❌ Bad: 1:1 mapping - just return domain type directly
// type DiagnosticDTO struct {
//     File     string
//     Line     int
//     Severity string
//     Message  string
// }
// Just return lint.Diagnostic directly!

// ✅ Good: Return domain type when CLI can use it
func (s *UserService) Validate(ctx context.Context, path string) ([]lint.Diagnostic, error) {
    if s == nil {
        return nil, fmt.Errorf("%w: UserService", ErrNilReceiver)
    }
    return s.validator.Validate(ctx, path) // Direct passthrough
}

// ✅ Good: Return DTO when combining data
func (s *UserService) GetCacheStatus(ctx context.Context, fileKey string) (*CacheStatusDTO, error) {
    if s == nil {
        return nil, fmt.Errorf("%w: UserService", ErrNilReceiver)
    }

    dbPath := filepath.Join(s.cacheDir, fileKey+".db")
    info, err := os.Stat(dbPath)
    if err != nil {
        return nil, err
    }

    // Combining multiple sources into DTO
    return &CacheStatusDTO{
        FileKey:      fileKey,
        DBPath:       dbPath,
        LastSyncedAt: info.ModTime(),
        SizeBytes:    info.Size(),
    }, nil
}
```

### Domain Layer (`pkg/`)

**Purpose:** Business logic, domain models, use cases.

**Responsibilities:**
- Define domain entities
- Business rules and validation
- Use case implementations
- Repository interfaces

**Rules:**
- ✅ Pure business logic
- ✅ Framework-agnostic
- ✅ Testable without CLI
- ❌ No CLI dependencies
- ❌ No `cmd/` imports

**Example:**
```go
// pkg/user/user.go
type User struct {
    ID       string
    Username string
    Email    string
    Status   Status
}

type Status int

const (
    StatusActive Status = iota
    StatusInactive
)

func (s Status) String() string {
    return [...]string{"active", "inactive"}[s]
}

// pkg/user/repository.go
type Repository interface {
    FindAll(ctx context.Context) ([]User, error)
    FindByID(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user *User) error
}
```

### Internal Layer (`internal/`)

**Purpose:** Application-specific implementations not exposed as public API.

**Responsibilities:**
- Adapters for external services
- Infrastructure code
- Implementation details

**Rules:**
- ✅ Can import `pkg/`
- ✅ Application-specific code
- ❌ Cannot be imported by external packages
- ❌ No `cmd/` imports

**Example:**
```go
// internal/adapter/database.go
type PostgresUserRepository struct {
    db *sql.DB
}

func (r *PostgresUserRepository) FindAll(ctx context.Context) ([]user.User, error) {
    // Database implementation
}
```

## Type Placement Rules

### Domain Types → `pkg/`

Domain entities, value objects, business logic types.

```go
// pkg/user/user.go
type User struct {
    ID       string
    Username string
    Email    string
}

type ScaffoldOptions struct {
    Username string
    Email    string
    Password string
}
```

### CLI DTOs → `services/`

Presentation types for CLI display only.

```go
// cmd/app/services/user.go
type UserListEntry struct {
    ID       string  // For CLI table
    Username string  // For CLI table
    Status   string  // For CLI table
}
```

### Decision Tree

**Where should this type go?**

1. Is it a domain concept (User, Order, Product)?
   → `pkg/domain/`

2. Is it for CLI display only (table row, summary)?
   → `services/`

3. Is it an implementation detail (DB model, API response)?
   → `internal/`

4. Is it a shared utility (Result, Option)?
   → `pkg/common/`

## Anti-Patterns

### ❌ Input Structs in Services

```go
// Bad - domain concept in services layer
type CreateUserInput struct {
    Username string
    Email    string
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput)
```

**Fix:** Use domain types from `pkg/`

```go
// Good - use domain type
func (s *UserService) Create(ctx context.Context, opts user.ScaffoldOptions)
```

### ❌ map[string]any Returns

```go
// Bad - untyped return
func (s *UserService) Get(ctx context.Context, id string) (map[string]any, error)
```

**Fix:** Return typed DTO

```go
// Good - typed DTO
func (s *UserService) Get(ctx context.Context, id string) (*UserDetail, error)
```

### ❌ Business Logic in Services

```go
// Bad - validation in services
func (s *UserService) Create(ctx context.Context, opts user.ScaffoldOptions) error {
    if len(opts.Username) < 3 {
        return errors.New("username too short")
    }
    // ...
}
```

**Fix:** Delegate to domain

```go
// Good - validation in domain
func (s *UserService) Create(ctx context.Context, opts user.ScaffoldOptions) error {
    u, err := user.New(opts) // validation happens here
    if err != nil {
        return err
    }
    return s.repo.Create(ctx, u)
}
```

## Dependency Flow

```
Commands → Services → Domain (pkg/) ← Internal
   ↓          ↓           ↑              ↑
  CLI      Coord      Business        Impl
  thin     layer       logic         details
```

**Rules:**
- Commands can only import `services/`
- Services can import `pkg/` and `internal/`
- Domain (`pkg/`) imports nothing from app
- Internal can import `pkg/`

## Testing Strategy

**Commands:** Integration tests with real services

**Services:** Unit tests with mocked repositories

**Domain:** Pure unit tests, no mocks needed

**Internal:** Integration tests with real dependencies

## Type-Safe Validation Pattern

### Bind + Validate Pattern

Always use typed structs with `ctx.Bind()` followed by `gozod.FromStruct[T]()` validation:

```go
type CreateConfig struct {
    Username string `flag:"username" validate:"required,min=3,max=50"`
    Email    string `flag:"email" validate:"required,email"`
    Password string `flag:"password" validate:"required,min=8"`
    Admin    bool   `flag:"admin"`
}

Handler: func(ctx *command.Context) error {
    var cfg CreateConfig

    // 1. Bind flags to typed struct
    if err := ctx.Bind(&cfg); err != nil {
        return err
    }

    // 2. Validate using gozod schema
    schema := gozod.FromStruct[CreateConfig](gozod.WithTagName("validate"))
    if _, err := schema.Parse(cfg); err != nil {
        return err
    }

    // 3. Pass validated data to service
    svc := services.NewUserService()
    return svc.Create(ctx.Context(), cfg.Username, cfg.Email, cfg.Password, cfg.Admin)
}
```

### Benefits

- **Type safety** - Compile-time checking of config fields
- **Validation** - Declarative rules in struct tags
- **Autocomplete** - IDE support for config fields
- **Refactoring** - Rename fields safely across codebase
- **Documentation** - Struct tags document validation rules

### Anti-Pattern: Untyped Config

```go
// ❌ Bad - no type safety, no validation
Handler: func(ctx *command.Context) error {
    username := ctx.String("username")
    email := ctx.String("email")
    // Manual validation, error-prone
    if username == "" {
        return errors.New("username required")
    }
}
```

### Validation Rules

Common gozod validation tags:

- `required` - Field must be non-zero
- `min=N,max=N` - String length or number range
- `email` - Valid email format
- `url` - Valid URL format
- `enum=a,b,c` - Value must be one of enum values

See [gozod documentation](https://github.com/kaptinlin/gozod) for full tag syntax.

