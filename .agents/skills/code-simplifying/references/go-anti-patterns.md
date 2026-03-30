# Go Anti-Patterns

Common anti-patterns to avoid in Go code.

## Error Handling Anti-Patterns

### Ignoring Errors

**❌ Anti-Pattern**
```go
file.Close()
json.Unmarshal(data, &result)
```

**✅ Correct**
```go
if err := file.Close(); err != nil {
    log.Printf("failed to close file: %v", err)
}

if err := json.Unmarshal(data, &result); err != nil {
    return fmt.Errorf("unmarshal: %w", err)
}
```

### Using Panic for Normal Errors

**❌ Anti-Pattern**
```go
func MustConnect() *sql.DB {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        panic(err)
    }
    return db
}
```

**✅ Correct**
```go
func Connect() (*sql.DB, error) {
    return sql.Open("postgres", dsn)
}
```

### Swallowing Errors

**❌ Anti-Pattern**
```go
if err := doSomething(); err != nil {
    log.Println(err)
    // continue as if nothing happened
}
```

**✅ Correct**
```go
if err := doSomething(); err != nil {
    return fmt.Errorf("do something: %w", err)
}
```

## Concurrency Anti-Patterns

### Goroutine Leaks

**❌ Anti-Pattern**
```go
func leak() {
    ch := make(chan int)
    go func() {
        val := <-ch  // blocks forever if nothing sends
        process(val)
    }()
    // forgot to send to ch or close it
}
```

**✅ Correct**
```go
func noLeak(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case val := <-ch:
            process(val)
        case <-ctx.Done():
            return
        }
    }()
    defer close(ch)
    // ...
}
```

### Race Conditions

**❌ Anti-Pattern**
```go
var counter int

func increment() {
    counter++ // race condition
}

go increment()
go increment()
```

**✅ Correct**
```go
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}
```

### Not Closing Channels

**❌ Anti-Pattern**
```go
func produce() <-chan int {
    ch := make(chan int)
    go func() {
        for i := 0; i < 10; i++ {
            ch <- i
        }
        // forgot to close(ch)
    }()
    return ch
}

for val := range produce() { // hangs after 10 values
    process(val)
}
```

**✅ Correct**
```go
func produce() <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 0; i < 10; i++ {
            ch <- i
        }
    }()
    return ch
}
```

### Capturing Loop Variables

**❌ Anti-Pattern**
```go
for _, item := range items {
    go func() {
        process(item) // all goroutines see last item
    }()
}
```

**✅ Correct**
```go
for _, item := range items {
    item := item // capture loop variable
    go func() {
        process(item)
    }()
}

// Or (Go 1.22+)
for _, item := range items {
    go func() {
        process(item) // loop variable is per-iteration
    }()
}
```

## Performance Anti-Patterns

### Unnecessary Allocations

**❌ Anti-Pattern**
```go
func concat(strs []string) string {
    result := ""
    for _, s := range strs {
        result += s // allocates new string each time
    }
    return result
}
```

**✅ Correct**
```go
func concat(strs []string) string {
    var b strings.Builder
    for _, s := range strs {
        b.WriteString(s)
    }
    return b.String()
}
```

### Growing Slices Inefficiently

**❌ Anti-Pattern**
```go
var results []Result
for _, item := range items {
    results = append(results, process(item)) // may reallocate many times
}
```

**✅ Correct**
```go
results := make([]Result, 0, len(items)) // pre-allocate capacity
for _, item := range items {
    results = append(results, process(item))
}
```

### Using defer in Loops

**❌ Anti-Pattern**
```go
for _, filename := range filenames {
    file, err := os.Open(filename)
    if err != nil {
        continue
    }
    defer file.Close() // defers accumulate, files stay open
    process(file)
}
```

**✅ Correct**
```go
for _, filename := range filenames {
    func() {
        file, err := os.Open(filename)
        if err != nil {
            return
        }
        defer file.Close() // closes at end of iteration
        process(file)
    }()
}
```

### Copying Mutexes

**❌ Anti-Pattern**
```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c Counter) Increment() { // copies mutex
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

**✅ Correct**
```go
func (c *Counter) Increment() { // pointer receiver
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

## Design Anti-Patterns

### God Objects

**❌ Anti-Pattern**
```go
type Service struct {
    db     *sql.DB
    cache  *redis.Client
    logger *log.Logger
    config *Config
    // ... 20 more fields
}

func (s *Service) DoEverything() {
    // 500 lines of code
}
```

**✅ Correct**
```go
type UserService struct {
    repo   UserRepository
    logger *log.Logger
}

type OrderService struct {
    repo   OrderRepository
    logger *log.Logger
}
```

### Premature Abstraction

**❌ Anti-Pattern**
```go
// Only one implementation exists
type UserGetter interface {
    GetUser(id int) (*User, error)
}

type UserRepository struct{}

func (r *UserRepository) GetUser(id int) (*User, error) {
    // ...
}
```

**✅ Correct**
```go
// Start with concrete type
type UserRepository struct{}

func (r *UserRepository) GetUser(id int) (*User, error) {
    // ...
}

// Add interface when second implementation appears
```

### Using `interface{}` Everywhere

**❌ Anti-Pattern**
```go
func Process(data interface{}) interface{} {
    // type assertions everywhere
    if str, ok := data.(string); ok {
        return len(str)
    }
    if num, ok := data.(int); ok {
        return num * 2
    }
    return nil
}
```

**✅ Correct**
```go
func ProcessString(s string) int {
    return len(s)
}

func ProcessInt(n int) int {
    return n * 2
}

// Or use generics
func Process[T any](data T, fn func(T) T) T {
    return fn(data)
}
```

### Getters and Setters Everywhere

**❌ Anti-Pattern**
```go
type User struct {
    name string
    age  int
}

func (u *User) GetName() string { return u.name }
func (u *User) SetName(name string) { u.name = name }
func (u *User) GetAge() int { return u.age }
func (u *User) SetAge(age int) { u.age = age }
```

**✅ Correct**
```go
type User struct {
    Name string
    Age  int
}

// Only add getters/setters if you need validation or side effects
func (u *User) SetAge(age int) error {
    if age < 0 {
        return errors.New("age cannot be negative")
    }
    u.Age = age
    return nil
}
```

## Testing Anti-Patterns

### Testing Implementation Details

**❌ Anti-Pattern**
```go
func TestUserService_internal(t *testing.T) {
    s := &UserService{}
    // testing private methods or internal state
    if s.cache == nil {
        t.Error("cache not initialized")
    }
}
```

**✅ Correct**
```go
func TestUserService_GetUser(t *testing.T) {
    s := NewUserService()
    user, err := s.GetUser(123)
    assert.NoError(t, err)
    assert.Equal(t, "John", user.Name)
}
```

### Not Using Table-Driven Tests

**❌ Anti-Pattern**
```go
func TestAdd(t *testing.T) {
    if Add(1, 2) != 3 {
        t.Error("1+2 should be 3")
    }
    if Add(-1, -2) != -3 {
        t.Error("-1+-2 should be -3")
    }
    // ... many more if statements
}
```

**✅ Correct**
```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Package Organization Anti-Patterns

### Circular Dependencies

**❌ Anti-Pattern**
```
package user imports package order
package order imports package user
```

**✅ Correct**
```
package user imports package domain
package order imports package domain
package domain has shared types
```

### Utils/Helpers Packages

**❌ Anti-Pattern**
```
utils/
├── string_utils.go
├── time_utils.go
├── http_utils.go
└── db_utils.go
```

**✅ Correct**
```
stringutil/
├── format.go
└── validate.go

timeutil/
└── parse.go

httputil/
└── client.go
```

### Too Many Small Packages

**❌ Anti-Pattern**
```
user/
├── user.go
userservice/
├── service.go
userrepository/
├── repository.go
```

**✅ Correct**
```
user/
├── user.go
├── service.go
└── repository.go
```

## References

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://go.dev/doc/effective_go)
- [Common Mistakes](https://github.com/golang/go/wiki/CommonMistakes)
