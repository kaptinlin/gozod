# Concurrency

Correct goroutine lifecycle management, synchronization patterns, and channel usage prevent leaks, races, and deadlocks.

## Contents
- Goroutine Lifetimes
- Prefer Synchronous Functions
- Specify Channel Direction
- Don't Copy Sync Types
- Don't Panic for Normal Errors
- Avoid Variable Shadowing

---

## Goroutine Lifetimes

When spawning goroutines, make it clear when and whether they exit. Use `sync.WaitGroup` or similar to prevent goroutine leaks.

**Incorrect:**

```go
func (w *Worker) Run() {
    for item := range w.q {
        go process(item) // may never complete, no tracking
    }
}
```

**Correct:**

```go
func (w *Worker) Run(ctx context.Context) error {
    var wg sync.WaitGroup
    for item := range w.q {
        wg.Add(1)
        go func() {
            defer wg.Done()
            process(ctx, item)
        }()
    }
    wg.Wait()
    return nil
}
```

Goroutines blocked on channel operations will leak if no other goroutine holds the channel reference. The garbage collector will not terminate blocked goroutines.

---

## Prefer Synchronous Functions

Prefer synchronous functions that return results directly. Let callers add concurrency if needed — removing unnecessary concurrency from callers is much harder.

**Incorrect:**

```go
func Process(ctx context.Context, input string) <-chan Result {
    ch := make(chan Result, 1)
    go func() {
        ch <- result
    }()
    return ch
}
```

**Correct:**

```go
func Process(ctx context.Context, input string) (Result, error) {
    return result, nil
}

// Caller adds concurrency when needed:
go func() {
    result, err := Process(ctx, input)
}()
```

Synchronous functions localize goroutines within the call, making lifetimes easier to reason about. They're also easier to test.

---

## Specify Channel Direction

Specify channel direction in function signatures to prevent misuse. The compiler catches errors like sending on a receive-only channel.

**Incorrect:**

```go
func sum(values chan int) (out int) {
    for v := range values {
        out += v
    }
    close(values) // dangerous: double close panics
}
```

**Correct:**

```go
func sum(values <-chan int) int {
    var out int
    for v := range values {
        out += v
    }
    return out
}
```

Channel direction documents ownership: `<-chan` consumes but cannot close; `chan<-` produces.

---

## Don't Copy Sync Types

Never copy `sync.Mutex`, `bytes.Buffer`, or any type with pointer-receiver methods. Copying creates separate internal state that leads to races and bugs.

**Incorrect:**

```go
type Counter struct {
    mu    sync.Mutex
    count int
}

func process(c Counter) { // copies the mutex!
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

**Correct:**

```go
type Counter struct {
    mu   sync.Mutex
    data map[string]int64
}

func (c *Counter) IncrementBy(name string, n int64) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[name] += n
}
```

General rule: if a type `T` has methods on `*T`, don't copy `T` values. Use `go vet` to detect accidental copies.

---

## Don't Panic for Normal Errors

Use `error` and multiple returns for normal error handling. Reserve `panic` for truly unrecoverable situations and API misuse. Internal panics must never cross package boundaries.

**Incorrect:**

```go
func Version(o *servicepb.Object) *version.Version {
    return version.MustParse(o.GetVersionString())
}
```

**Correct:**

```go
func Version(o *servicepb.Object) (*version.Version, error) {
    v, err := version.Parse(o.GetVersionString())
    if err != nil {
        return nil, fmt.Errorf("parsing version: %w", err)
    }
    return v, nil
}
```

`Must` functions are acceptable only for package-level "constants" during initialization:

```go
var DefaultVersion = MustParse("1.2.3") // package init only
```

---

## Avoid Variable Shadowing

Be careful with `:=` in inner scopes — it creates new variables that shadow the outer ones. This is a common source of subtle bugs.

**Incorrect:**

```go
func (s *Server) innerHandler(ctx context.Context, req *pb.MyRequest) *pb.MyResponse {
    if *shortenDeadlines {
        ctx, cancel := context.WithTimeout(ctx, 3*time.Second) // shadows outer ctx
        defer cancel()
    }
    // BUG: ctx here is still the original caller's context
}
```

**Correct:**

```go
func (s *Server) innerHandler(ctx context.Context, req *pb.MyRequest) *pb.MyResponse {
    if *shortenDeadlines {
        var cancel func()
        ctx, cancel = context.WithTimeout(ctx, 3*time.Second) // note: = not :=
        defer cancel()
    }
    // ctx is correctly the deadline-capped context
}
```

Don't shadow standard package names (`url`, `path`, etc.) except in very small scopes.
