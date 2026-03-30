# Golang Philosophy

## Core Thesis

Clear is better than clever. Simplicity is the foundation of reliability. Go favors explicit communication over shared state, small interfaces over large abstractions, and practical solutions over theoretical perfection.

## Principles

### Clear is better than clever

Write code that is easy to understand, not code that is impressive to write.

Go values clarity and maintainability over cleverness. Code is read far more often than it is written. When you write clear code, you're writing for the next person who will maintain it—including your future self.

**好的做法**:
```go
// Good: Clear intent, easy to understand
func isEligibleForDiscount(user *User) bool {
    isActive := user.Status == "active"
    hasSubscription := user.SubscriptionLevel > 0

    return isActive && hasSubscription
}
```

**不好的做法**:
```go
// Bad: Clever but unclear
func isEligibleForDiscount(u *User) bool {
    return u.Status == "active" && u.SubscriptionLevel > 0 &&
           (u.LastLogin.After(time.Now().Add(-30*24*time.Hour)) ||
            u.TotalPurchases > 1000)
}
```

> **Why**: Clear code reduces cognitive load, makes debugging easier, and enables team collaboration. Clever code creates maintenance burden and knowledge silos.
> **Rejected**: Optimizing for brevity at the expense of clarity — leads to bugs and maintenance nightmares.

### Make the zero value useful

Design types so their zero value is ready to use without explicit initialization.

Go's zero values (`0` for numbers, `""` for strings, `nil` for pointers/slices/maps) should be meaningful. A well-designed type works correctly when declared with `var` without additional setup.

**好的做法**:
```go
// Good: Zero value is immediately useful
type Logger struct {
    prefix string
    writer io.Writer // nil is ok, can default to os.Stderr
}

func (l *Logger) Log(msg string) {
    if l.writer == nil {
        l.writer = os.Stderr
    }
    fmt.Fprintf(l.writer, "%s: %s\n", l.prefix, msg)
}

// Usage: no constructor needed
var log Logger
log.Log("hello") // works immediately
```

**不好的做法**:
```go
// Bad: Zero value is broken, requires constructor
type Logger struct {
    writer io.Writer // nil causes panic
}

func (l *Logger) Log(msg string) {
    fmt.Fprintf(l.writer, "%s\n", msg) // panics if writer is nil
}

// Must use constructor
func NewLogger() *Logger {
    return &Logger{writer: os.Stderr}
}
```

> **Why**: Useful zero values reduce API surface (fewer constructors), make types easier to embed, and prevent initialization bugs.
> **Rejected**: Requiring constructors for all types — increases boilerplate and makes composition harder.

### A little copying is better than a little dependency

Sometimes duplicating a small amount of code is preferable to adding a dependency.

Dependencies add complexity: larger binaries, longer build times, more maintenance burden, and potential security vulnerabilities. For small, self-contained functions, copying is often the pragmatic choice.

**好的做法**:
```go
// Good: Copy small utility function instead of importing large library
// Source: adapted from github.com/example/utils
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

**不好的做法**:
```go
// Bad: Import entire library for one function
import "github.com/example/utils" // 50+ functions, but we only need 1

func processItems(items []string) {
    if utils.Contains(items, "target") {
        // ...
    }
}
```

> **Why**: Minimizing dependencies reduces build times, binary size, and maintenance burden. Small copied functions can be customized for specific needs.
> **Rejected**: Zero duplication at all costs — leads to unnecessary dependencies and tight coupling.
> **Note**: The Go standard library uses this principle. For example, `strconv` implements its own `isPrint` function instead of depending on `unicode`, saving ~150 KB of data tables.

### The bigger the interface, the weaker the abstraction

Small, focused interfaces are more powerful and flexible than large ones.

Go's most powerful interfaces have one or two methods: `io.Reader`, `io.Writer`, `fmt.Stringer`. Small interfaces are easier to implement, test, and compose. They force you to think about essential behavior rather than implementation details.

**好的做法**:
```go
// Good: Small, focused interfaces
type Saver interface {
    Save(data Data) error
}

type Loader interface {
    Load(id string) (Data, error)
}

// Compose when needed
type Repository interface {
    Saver
    Loader
}
```

**不好的做法**:
```go
// Bad: Large interface with many methods
type DataStore interface {
    Save(data Data) error
    Load(id string) (Data, error)
    Delete(id string) error
    List() ([]Data, error)
    Count() (int, error)
    Search(query string) ([]Data, error)
    Backup() error
    Restore(backup []byte) error
}
```

> **Why**: Small interfaces have more implementations, are easier to test (fewer methods to mock), and enable better composition.
> **Rejected**: Comprehensive interfaces that cover all use cases — creates tight coupling and makes testing difficult.

### Errors are values

Errors are just values you can program with, not control flow exceptions.

In Go, errors are returned values, not thrown exceptions. This means you can accumulate them, transform them, wrap them, or make decisions based on them. Don't just check and return—consider what should happen.

**好的做法**:
```go
// Good: Error accumulator pattern
type ErrorWriter struct {
    w   io.Writer
    err error
}

func (ew *ErrorWriter) Write(buf []byte) {
    if ew.err != nil {
        return // Skip if already failed
    }
    _, ew.err = ew.w.Write(buf)
}

// Clean usage
ew := &ErrorWriter{w: w}
ew.Write(header)
ew.Write(body)
ew.Write(footer)
if ew.err != nil {
    return ew.err
}
```

**不好的做法**:
```go
// Bad: Repetitive error checking
if _, err := w.Write(header); err != nil {
    return err
}
if _, err := w.Write(body); err != nil {
    return err
}
if _, err := w.Write(footer); err != nil {
    return err
}
```

> **Why**: Treating errors as values enables creative error-handling strategies and reduces boilerplate.
> **Rejected**: Exception-based error handling — hides control flow and makes error handling implicit.

### Don't communicate by sharing memory; share memory by communicating

Use channels to pass data between goroutines instead of shared memory with locks.

When you send data over a channel, ownership transfers. No simultaneous access means no race conditions. This makes concurrent code safer and easier to reason about.

**好的做法**:
```go
// Good: Communicate via channels
type CacheUpdate struct {
    key   string
    value string
}

updates := make(chan CacheUpdate)

go func() {
    cache := make(map[string]string)
    for update := range updates {
        cache[update.key] = update.value
    }
}()

// Send updates
updates <- CacheUpdate{"user:1", "Alice"}
```

**不好的做法**:
```go
// Bad: Shared memory with mutex
var (
    cache map[string]string
    mu    sync.Mutex
)

func updateCache(key, value string) {
    mu.Lock()
    cache[key] = value
    mu.Unlock()
}
```

> **Why**: Channels make ownership and data flow explicit, preventing race conditions by design. Channels orchestrate; mutexes serialise — use channels for coordination between goroutines, mutexes for protecting simple shared state.
> **Rejected**: Shared memory as the default concurrency primitive — requires careful lock management and is error-prone.

### Concurrency is not parallelism

Concurrency is about structure; parallelism is about execution. Good concurrent design enables parallelism but doesn't require it. Don't assume concurrent code is automatically faster.

> **Why**: Understanding the distinction prevents confusion about performance.
> **Rejected**: Assuming concurrency automatically means faster execution — parallelism depends on hardware and workload.

### Don't just check errors, handle them gracefully

Consider what should happen when an error occurs; don't just return it.

Error handling is a critical part of your program's behavior. Add context, make decisions, use defaults, or retry. Think about what the caller needs to know.

**好的做法**:
```go
// Good: Add context and handle appropriately
func loadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            // Use defaults for missing config
            return defaultConfig(), nil
        }
        return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("invalid config format in %s: %w", path, err)
    }
    return &cfg, nil
}
```

**不好的做法**:
```go
// Bad: Just return without context
func loadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err // What failed? Where?
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err // Was it file read or parsing?
    }
    return &cfg, nil
}
```

> **Why**: Contextual errors help with debugging and enable better error recovery strategies. Use `%w` for error chains (enables `errors.Is`/`errors.As`), `%v` at system boundaries to prevent internal error leaking.
> **Rejected**: Blindly propagating errors without context — makes debugging difficult and error messages unhelpful.

## Design Principles (SOLID in Go)

### Single Responsibility Principle

Each package, type, and function should have one reason to change. Packages focus on a single domain, functions do one thing well.

**好的做法**:
```go
// Good: Focused package — package user only handles user logic
type Service struct{ repo Repository }
func (s *Service) Create(u *User) error { /* only user creation */ }
```

> **Philosophy**: 体现 "Clear is better than clever" 原则。单一职责使代码更清晰、更易维护。

### Open/Closed Principle

Types should be open for extension but closed for modification.

Use interfaces to enable extension without modifying existing code. Go's implicit interface satisfaction makes this natural.

**好的做法**:
```go
// Good: Extend via interface
type Notifier interface {
    Notify(msg string) error
}

type EmailNotifier struct{}
func (e *EmailNotifier) Notify(msg string) error { /* ... */ }

type SMSNotifier struct{}
func (s *SMSNotifier) Notify(msg string) error { /* ... */ }

// Add new notifiers without changing existing code
```

> **Philosophy**: 体现 "The bigger the interface, the weaker the abstraction" 原则。小接口使扩展更容易。

### Liskov Substitution Principle

Interface implementations must be truly interchangeable.

Any implementation of an interface should work correctly wherever that interface is expected.

**好的做法**:
```go
// Good: All implementations honor the contract
type Storage interface {
    Save(key string, value []byte) error
}

// Both implementations work identically from caller's perspective
type MemoryStorage struct{}
type DiskStorage struct{}
```

> **Philosophy**: 体现 "Errors are values" 原则。接口契约包括错误处理行为。

### Interface Segregation Principle

Keep interfaces small and focused.

This is Go's idiomatic approach. Don't force clients to depend on methods they don't use.

**好的做法**:
```go
// Good: Small, focused interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

> **Philosophy**: 直接体现 "The bigger the interface, the weaker the abstraction" 原则。

### Dependency Inversion Principle

Accept interfaces, return concrete types. This makes code testable and flexible.

**好的做法**:
```go
// Good: Depend on interface, not concrete type
type Service struct{ storage Storage } // Storage is an interface
func NewService(s Storage) *Service { return &Service{storage: s} }
```

> **Philosophy**: 体现 "Make the zero value useful" 和接口组合原则。

## Simplicity Principles

### KISS (Keep It Simple, Stupid)

Simple design beats clever design.

Avoid over-engineering. Start with the simplest solution that works. Add complexity only when needed, not in anticipation of future needs.

**好的做法**:
```go
// Good: Simple, direct solution
func calculateTotal(items []Item) float64 {
    var total float64
    for _, item := range items {
        total += item.Price
    }
    return total
}
```

**不好的做法**:
```go
// Bad: Over-engineered for hypothetical future needs
type Calculator interface {
    Calculate(items []Item) float64
}

type TotalCalculator struct {
    strategy CalculationStrategy
    cache    *Cache
    logger   *Logger
}
```

> **Philosophy**: 直接体现 "Clear is better than clever" 原则。
> **Rejected**: Premature abstraction and over-engineering — adds complexity without proven benefit.

### DRY (Don't Repeat Yourself)

Reuse through interfaces and composition, not inheritance.

Go doesn't have inheritance. Use composition and interfaces to share behavior. But remember: "A little copying is better than a little dependency."

**好的做法**:
```go
// Good: Composition
type Logger struct {
    output io.Writer
}

type Service struct {
    logger *Logger
    db     Database
}
```

> **Philosophy**: 平衡 DRY 和 "A little copying is better than a little dependency" 原则。

### YAGNI (You Aren't Gonna Need It)

Only implement what you need now.

Don't build features for hypothetical future requirements. Wait until requirements are clear, then implement.

**好的做法**:
```go
// Good: Implement current requirements only
type User struct {
    ID   int
    Name string
}
```

**不好的做法**:
```go
// Bad: Anticipating未来 needs
type User struct {
    ID              int
    Name            string
    FutureField1    string // "We might need this"
    FutureField2    int    // "Just in case"
    ExtensionPoints map[string]interface{} // "For flexibility"
}
```

> **Philosophy**: 体现 "Clear is better than clever" 和 KISS 原则。
> **Rejected**: Designing for hypothetical future requirements — adds complexity and maintenance burden.

## Package Design Principles

### Organize by feature, not by layer

Group related functionality together rather than separating by technical layer.

**好的做法**:
```
project/
├── user/
│   ├── user.go
│   ├── repository.go
│   └── service.go
├── order/
└── product/
```

**不好的做法**:
```
project/
├── model/
├── repository/
├── service/
└── controller/
```

> **Why**: Feature-based organization has high cohesion, small blast radius for changes, and easier feature location.
> **Rejected**: Layer-based organization — scatters related code across packages, increases coupling.

### High cohesion, low coupling

Package contents should be highly related; packages should depend on each other minimally. Only export what's necessary — smaller API surface means more freedom to refactor internals.

> **Philosophy**: 体现 Single Responsibility Principle。

### Avoid circular dependencies

Dependencies should flow in one direction. Circular dependencies make code hard to understand, test, and refactor.

### Avoid util/common/helper packages

Use meaningful package names: `project/codec`, `project/validator` — not `project/util`, `project/common`.

> **Why**: Meaningful names make code self-documenting and prevent packages from becoming dumping grounds.

## Naming Conventions

### Avoid repetition

Don't repeat context from the package name or receiver type: `yaml.Parse()` not `yaml.ParseYAML()`, `c.JobName()` not `c.GetJobName()`.

### Nouns for getters, verbs for actions

Getters use nouns (`c.JobName()`), actions use verbs (`c.WriteDetail(w)`). Getters don't use "Get" prefix — this is Go convention.

> **Why**: Reduces noise at call sites and follows Go idioms.

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Package** | Go's unit of code organization | Not "module" (module is a collection of packages) |
| **Interface** | A set of method signatures | Not "abstract class" (Go has no classes) |
| **Embedding** | Composition by including a type | Not "inheritance" (Go has no inheritance) |
| **Goroutine** | Lightweight thread managed by Go runtime | Not "thread" (goroutines are multiplexed onto threads) |
| **Channel** | Typed conduit for communication between goroutines | Not "queue" (channels can be unbuffered) |
| **Zero value** | Default value of a type when declared | Not "null" or "undefined" |

## Forbidden

- **Don't use panic for expected errors**: Panic is for programmer errors and impossible conditions → Return errors for expected failures
- **Don't use global state**: Global variables create hidden dependencies → Pass dependencies explicitly or use dependency injection
- **Don't use init for setup with side effects**: `init()` runs automatically and can't return errors → Use explicit initialization functions
- **Don't use interface{} without good reason**: Empty interface loses all type safety → Use specific interfaces or generics (Go 1.18+)
- **Don't use reflection in normal code**: Reflection is slow and loses compile-time safety → Reserve for libraries like `encoding/json`
- **Don't create god packages**: Packages with too many responsibilities → Follow Single Responsibility Principle
- **Don't use getters/setters by default**: Direct field access is idiomatic → Only add methods when logic is needed

## References

- [00-principles.md] — SPEC 核心原则
- [01-philosophy-template.md] — Philosophy 文档模板
- Go Proverbs (Rob Pike) — 格言式范本
- Effective Go — 官方最佳实践
- Google Go Style Guide — Google 编码规范
- Kubernetes client-go — 包设计实践
