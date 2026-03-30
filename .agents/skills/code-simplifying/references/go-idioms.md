# Go Idioms

Common Go idioms and best practices.

## Error Handling Idioms

### Check Errors Immediately

```go
result, err := DoSomething()
if err != nil {
    return fmt.Errorf("do something: %w", err)
}
// use result
```

### Wrap Errors with Context

```go
if err := process(); err != nil {
    return fmt.Errorf("process user %s: %w", userID, err)
}
```

### Use errors.Is for Sentinel Errors

```go
if errors.Is(err, io.EOF) {
    // handle EOF
}
```

### Use errors.As for Error Types

```go
var netErr *net.OpError
if errors.As(err, &netErr) {
    // handle network error
}
```

## Concurrency Idioms

### Use Channels for Communication

```go
results := make(chan Result)
go func() {
    defer close(results)
    for item := range items {
        results <- process(item)
    }
}()

for result := range results {
    // handle result
}
```

### Use sync.WaitGroup for Goroutine Coordination

```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        process(item)
    }(item)
}
wg.Wait()
```

### Use Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

select {
case result := <-doWork(ctx):
    return result
case <-ctx.Done():
    return ctx.Err()
}
```

### Use errgroup for Error Propagation

```go
g, ctx := errgroup.WithContext(ctx)

for _, item := range items {
    item := item // capture loop variable
    g.Go(func() error {
        return process(ctx, item)
    })
}

if err := g.Wait(); err != nil {
    return err
}
```

## Interface Idioms

### Accept Interfaces, Return Structs

```go
// Good
func Process(r io.Reader) (*Result, error) {
    // ...
}

// Avoid
func Process(r io.Reader) io.Reader {
    // ...
}
```

### Keep Interfaces Small

```go
// Good
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Avoid
type FileOperations interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
    Seek(offset int64, whence int) (int64, error)
    // ... many more methods
}
```

### Use Interface Composition

```go
type ReadWriter interface {
    Reader
    Writer
}
```

## Resource Management Idioms

### Use defer for Cleanup

```go
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()

// use file
```

### Defer in Correct Order

```go
// Defers execute in LIFO order
defer cleanup1() // runs last
defer cleanup2() // runs first
```

### Use defer with Named Return Values for Error Handling

```go
func DoSomething() (err error) {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()

    // ... operations
    return nil
}
```

## Initialization Idioms

### Use init() Sparingly

```go
// Avoid init() when possible
// Prefer explicit initialization

func NewService() *Service {
    return &Service{
        // explicit initialization
    }
}
```

### Use sync.Once for Lazy Initialization

```go
var (
    instance *Service
    once     sync.Once
)

func GetService() *Service {
    once.Do(func() {
        instance = &Service{}
    })
    return instance
}
```

### Use sync.OnceValue (Go 1.21+)

```go
var getConfig = sync.OnceValue(func() *Config {
    return loadConfig()
})

func GetConfig() *Config {
    return getConfig()
}
```

## Testing Idioms

### Use Table-Driven Tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

### Use testify for Assertions

```go
func TestSomething(t *testing.T) {
    result := DoSomething()
    assert.NotNil(t, result)
    assert.Equal(t, expected, result.Value)
}
```

### Use t.Helper() for Test Helpers

```go
func assertValid(t *testing.T, v Validator) {
    t.Helper()
    if err := v.Validate(); err != nil {
        t.Fatalf("validation failed: %v", err)
    }
}
```

## Package Organization Idioms

### One Package Per Directory

```
myproject/
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
│   ├── auth/
│   │   └── auth.go
│   └── db/
│       └── db.go
└── pkg/
    └── api/
        └── api.go
```

### Use internal/ for Private Code

```
myproject/
├── internal/        # private to this module
│   └── helper/
└── pkg/            # public API
    └── client/
```

### Avoid Circular Dependencies

```go
// Bad: package A imports B, B imports A

// Good: extract shared code to package C
// A imports C, B imports C
```

## Naming Idioms

### Use Short Variable Names in Small Scopes

```go
// Good
for i, v := range items {
    // ...
}

// Avoid
for index, value := range items {
    // ...
}
```

### Use Descriptive Names in Large Scopes

```go
// Good
var userRepository *Repository

// Avoid
var ur *Repository
```

### Don't Stutter Package Names

```go
// Bad
user.UserService

// Good
user.Service
```

### Use MixedCaps for Multi-Word Names

```go
// Good
type UserService struct{}
func (s *UserService) GetUserByID(id int) {}

// Avoid
type User_Service struct{}
func (s *User_Service) Get_User_By_ID(id int) {}
```

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
