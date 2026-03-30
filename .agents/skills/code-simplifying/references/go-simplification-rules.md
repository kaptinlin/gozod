# Go Simplification Rules

Language-specific simplification rules for Go 1.26+.

## Go 1.26+ New Features

### Use `range over int` for Simple Loops

**❌ Old Style**
```go
for i := 0; i < 10; i++ {
    process(i)
}
```

**✅ Go 1.22+ Style**
```go
for i := range 10 {
    process(i)
}
```

**Benefits**: More concise, clearer intent

### Use `iter.Seq` for Iterator Patterns

**❌ Old Style**
```go
func GetItems() []Item {
    items := make([]Item, 0)
    // ... populate items
    return items
}

for _, item := range GetItems() {
    process(item)
}
```

**✅ Go 1.23+ Style**
```go
func GetItems() iter.Seq[Item] {
    return func(yield func(Item) bool) {
        // ... yield items
    }
}

for item := range GetItems() {
    process(item)
}
```

**Benefits**: Memory efficient, lazy evaluation, composable

### Use `new(expr)` for Pointer Creation

**❌ Old Style**
```go
tmp := SomeComplexExpression()
ptr := &tmp
```

**✅ Go 1.26+ Style**
```go
ptr := new(SomeComplexExpression())
```

**Benefits**: More concise, clearer intent

### Use `clear()` for Map/Slice Cleanup

**❌ Old Style**
```go
m = make(map[string]int)
s = s[:0]
```

**✅ Go 1.21+ Style**
```go
clear(m)
clear(s)
```

**Benefits**: More explicit, works with both maps and slices

### Use `slices` and `maps` Packages

**❌ Old Style**
```go
func contains(slice []int, val int) bool {
    for _, v := range slice {
        if v == val {
            return true
        }
    }
    return false
}
```

**✅ Go 1.21+ Style**
```go
import "slices"

found := slices.Contains(slice, val)
```

**Benefits**: Standard library, well-tested, optimized

## Go Idioms

### Early Return to Reduce Nesting

**❌ Nested Style**
```go
func Process(data []byte) error {
    if len(data) > 0 {
        if valid := Validate(data); valid {
            if result, err := Transform(data); err == nil {
                return Save(result)
            } else {
                return err
            }
        } else {
            return errors.New("invalid data")
        }
    } else {
        return errors.New("empty data")
    }
}
```

**✅ Early Return Style**
```go
func Process(data []byte) error {
    if len(data) == 0 {
        return errors.New("empty data")
    }

    if !Validate(data) {
        return errors.New("invalid data")
    }

    result, err := Transform(data)
    if err != nil {
        return err
    }

    return Save(result)
}
```

**Benefits**: Flatter structure, easier to read, clearer error paths

### Use `errors.Is/As` Instead of Type Assertions

**❌ Type Assertion**
```go
if err != nil {
    if e, ok := err.(*MyError); ok {
        // handle MyError
    }
}
```

**✅ errors.As**
```go
var myErr *MyError
if errors.As(err, &myErr) {
    // handle MyError
}
```

**Benefits**: Works with wrapped errors, more robust

### Use `context.Context` for Cancellation

**❌ Without Context**
```go
func LongRunning() error {
    for {
        // ... work
    }
}
```

**✅ With Context**
```go
func LongRunning(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // ... work
        }
    }
}
```

**Benefits**: Cancellable, timeout support, request-scoped values

### Use `sync.OnceValue` for Lazy Initialization

**❌ Manual sync.Once**
```go
var (
    instance *Service
    once     sync.Once
)

func GetService() *Service {
    once.Do(func() {
        instance = NewService()
    })
    return instance
}
```

**✅ sync.OnceValue**
```go
var getService = sync.OnceValue(func() *Service {
    return NewService()
})

func GetService() *Service {
    return getService()
}
```

**Benefits**: More concise, type-safe, clearer intent

### Use `slog` for Structured Logging

**❌ fmt.Printf**
```go
fmt.Printf("user %s logged in from %s\n", userID, ip)
```

**✅ slog**
```go
slog.Info("user logged in",
    "user_id", userID,
    "ip", ip)
```

**Benefits**: Structured, filterable, machine-readable

## Simplification Patterns

### Merge Duplicate Error Handling

**❌ Repeated Error Handling**
```go
data1, err := Fetch1()
if err != nil {
    return fmt.Errorf("fetch1 failed: %w", err)
}

data2, err := Fetch2()
if err != nil {
    return fmt.Errorf("fetch2 failed: %w", err)
}

data3, err := Fetch3()
if err != nil {
    return fmt.Errorf("fetch3 failed: %w", err)
}
```

**✅ Extract Helper**
```go
func fetchWithContext(name string, fn func() (Data, error)) (Data, error) {
    data, err := fn()
    if err != nil {
        return Data{}, fmt.Errorf("%s failed: %w", name, err)
    }
    return data, nil
}

data1, err := fetchWithContext("fetch1", Fetch1)
if err != nil {
    return err
}
// ... similar for data2, data3
```

**Benefits**: DRY, consistent error messages

### Extract Repeated Validation Logic

**❌ Repeated Validation**
```go
func CreateUser(name, email string) error {
    if name == "" {
        return errors.New("name required")
    }
    if len(name) > 100 {
        return errors.New("name too long")
    }
    if email == "" {
        return errors.New("email required")
    }
    if !strings.Contains(email, "@") {
        return errors.New("invalid email")
    }
    // ...
}
```

**✅ Extract Validators**
```go
func validateName(name string) error {
    if name == "" {
        return errors.New("name required")
    }
    if len(name) > 100 {
        return errors.New("name too long")
    }
    return nil
}

func validateEmail(email string) error {
    if email == "" {
        return errors.New("email required")
    }
    if !strings.Contains(email, "@") {
        return errors.New("invalid email")
    }
    return nil
}

func CreateUser(name, email string) error {
    if err := validateName(name); err != nil {
        return err
    }
    if err := validateEmail(email); err != nil {
        return err
    }
    // ...
}
```

**Benefits**: Reusable, testable, clearer

### Use Functional Options Pattern

**❌ Many Parameters**
```go
func NewServer(addr string, timeout time.Duration, maxConns int, logger *log.Logger) *Server {
    // ...
}

server := NewServer(":8080", 30*time.Second, 100, logger)
```

**✅ Functional Options**
```go
type ServerOption func(*Server)

func WithTimeout(d time.Duration) ServerOption {
    return func(s *Server) { s.timeout = d }
}

func WithMaxConns(n int) ServerOption {
    return func(s *Server) { s.maxConns = n }
}

func NewServer(addr string, opts ...ServerOption) *Server {
    s := &Server{addr: addr}
    for _, opt := range opts {
        opt(s)
    }
    return s
}

server := NewServer(":8080",
    WithTimeout(30*time.Second),
    WithMaxConns(100))
```

**Benefits**: Extensible, optional parameters, self-documenting

### Avoid Unnecessary Type Conversions

**❌ Unnecessary Conversion**
```go
var count int64 = 10
for i := int64(0); i < count; i++ {
    // ...
}
```

**✅ Direct Use**
```go
var count int64 = 10
for i := range count {
    // ...
}
```

**Benefits**: Cleaner, less noise

## Anti-Patterns to Avoid

### Don't Use `interface{}` When You Can Use Generics

**❌ interface{}**
```go
func Max(a, b interface{}) interface{} {
    // type assertions needed
}
```

**✅ Generics**
```go
func Max[T constraints.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

### Don't Ignore Errors

**❌ Ignored Error**
```go
file.Close()
```

**✅ Handle Error**
```go
if err := file.Close(); err != nil {
    return fmt.Errorf("close file: %w", err)
}
```

### Don't Use Panic for Normal Errors

**❌ Panic**
```go
func MustParse(s string) int {
    n, err := strconv.Atoi(s)
    if err != nil {
        panic(err)
    }
    return n
}
```

**✅ Return Error**
```go
func Parse(s string) (int, error) {
    return strconv.Atoi(s)
}
```

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
