# Error Handling Modernization

Modern Go provides richer error APIs: multi-error joining, generic type extraction, context cancellation causes, and standardized sentinel errors.

## Contents
- errors.Join (1.20+)
- fmt.Errorf with multiple %w (1.20+)
- errors.AsType[T] (1.26+)
- context.WithCancelCause (1.20+)
- context.AfterFunc (1.21+)
- errors.ErrUnsupported (1.20+)

---

## errors.Join (Go 1.20+)

Wraps multiple errors into a single error that implements `Unwrap() []error`. The result works with `errors.Is` and `errors.As` — both check against all wrapped errors.

### When to use
- Accumulating errors from a loop (validation, batch processing)
- Combining errors from goroutines
- Combining cleanup errors in `defer` with the main error via named returns
- Replacing `go.uber.org/multierr` or `github.com/hashicorp/go-multierror`

### When NOT to use
- For a single error with context — use `fmt.Errorf("context: %w", err)` instead
- When error ordering/hierarchy matters — `errors.Join` is a flat list
- When you need a formatted message wrapping errors — use `fmt.Errorf` with multiple `%w`

```go
// Accumulate validation errors
var errs []error
for _, item := range items {
    if err := validate(item); err != nil {
        errs = append(errs, err)
    }
}
return errors.Join(errs...)

// Combine cleanup error with main error (defer pattern)
func processFile(path string) (err error) {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer func() { err = errors.Join(err, f.Close()) }()

    // ... process f ...
    return nil
}

// Combine goroutine errors
var (
    mu   sync.Mutex
    errs []error
)
var wg sync.WaitGroup
for _, task := range tasks {
    wg.Go(func() {  // Go 1.25+
        if err := task.Run(); err != nil {
            mu.Lock()
            errs = append(errs, err)
            mu.Unlock()
        }
    })
}
wg.Wait()
return errors.Join(errs...)
```

**Key details**:
- `errors.Join` skips nil errors — returns nil if all inputs are nil
- Output is multi-line (errors separated by `\n`) — use `fmt.Errorf` with multiple `%w` for a single-line message
- **Must use named returns** in the defer pattern — unnamed returns won't be modified by defer

---

## fmt.Errorf with multiple %w (Go 1.20+)

### When to use
- When you want a **formatted message** that wraps multiple errors
- When errors have a logical relationship (e.g., "failed X and also failed Y")

### When NOT to use
- For simple error accumulation (list of independent errors) — use `errors.Join`
- For a single error — one `%w` is sufficient

```go
// Old — lost the second error
if err1 != nil && err2 != nil {
    return fmt.Errorf("write failed: %w (also close failed: %v)", err1, err2)
}

// New (Go 1.20+) — both errors are unwrappable
return fmt.Errorf("operation failed: %w; cleanup also failed: %w", writeErr, closeErr)
```

### %w vs %v — API design consideration

`%w` makes the inner error part of your **public API** — callers can `errors.Is(err, sql.ErrNoRows)`. If you switch databases, this is a breaking change. Use `%v` when you don't want to expose the inner error type:

```go
// Public API: inner error is exposed
return fmt.Errorf("query user: %w", err)

// Internal: inner error is hidden from callers
return fmt.Errorf("query user: %v", err)
```

---

## errors.AsType[T] (Go 1.26+)

Generic, type-safe replacement for `errors.As`. No reflection, no runtime panics for wrong target types, better performance (~30ns vs ~95ns).

### When to use
- **Always prefer** over `errors.As` in Go 1.26+ code
- Checking for specific error types in error chains
- Extracting typed error information

### When NOT to use
- When targeting Go < 1.26 — use `errors.As` for backwards compatibility
- `go fix ./...` (Go 1.26+) can auto-migrate `errors.As` calls to `errors.AsType`

```go
// Old
var pathErr *fs.PathError
if errors.As(err, &pathErr) {
    fmt.Println("failed at path:", pathErr.Path)
}

// New (Go 1.26+)
if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
    fmt.Println("failed at path:", pathErr.Path)
}

// Check multiple error types
if connErr, ok := errors.AsType[*net.OpError](err); ok {
    handleConnError(connErr)
} else if dnsErr, ok := errors.AsType[*net.DNSError](err); ok {
    handleDNSError(dnsErr)
}
```

---

## context.WithCancelCause (Go 1.20+)

Cancels a context with a **reason** (cause error) retrievable via `context.Cause(ctx)`.

### When to use
- When downstream code needs to know **why** a context was canceled
- Graceful shutdown with reason propagation
- Distinguishing timeout vs manual cancellation vs specific failure
- Replacing custom "cancellation reason" channels

### When NOT to use
- Simple cancellation where the reason is obvious — plain `context.WithCancel` is fine
- If nobody reads `context.Cause` — don't add complexity for nothing
- Third-party libraries that only check `ctx.Err()` won't see your custom cause

### Gotchas
- `cancel(nil)` defaults cause to `context.Canceled` — always pass a non-nil error
- Use `context.Cause(ctx)` instead of `ctx.Err()` to retrieve the actual cause

```go
// Old — no way to know why context was canceled
ctx, cancel := context.WithCancel(parent)
cancel()
// ctx.Err() == context.Canceled, but why?

// New (Go 1.20+) — cancel with a reason
ctx, cancel := context.WithCancelCause(parent)
cancel(fmt.Errorf("shutting down: %w", reason))
// Later:
cause := context.Cause(ctx) // "shutting down: ..."
```

### Related: context.WithDeadlineCause / WithTimeoutCause (Go 1.21+)

```go
ctx, cancel := context.WithTimeoutCause(parent, 5*time.Second,
    fmt.Errorf("database query exceeded 5s limit"))
defer cancel()
```

---

## context.AfterFunc (Go 1.21+)

Registers a callback that runs (in its own goroutine) after a context is done.

### When to use
- Cleanup actions triggered by context cancellation
- Stopping background workers when parent context ends
- Replacing `go func() { <-ctx.Done(); cleanup() }()` pattern

### When NOT to use
- For simple cleanup — `defer` is clearer
- The callback runs in a new goroutine — don't assume synchronous execution

```go
// Old
go func() {
    <-ctx.Done()
    conn.Close()
}()

// New (Go 1.21+)
stop := context.AfterFunc(ctx, func() {
    conn.Close()
})
// stop() returns true if the callback was prevented from running
defer stop()
```

---

## context.WithoutCancel (Go 1.21+)

Creates a derived context that is **not** canceled when the parent is canceled. Retains parent's values.

### When to use
- Background cleanup that must complete even after request context cancellation
- Logging/metrics that should survive the parent context

### When NOT to use
- Most cases — cancellation propagation is usually desired
- Don't use to "fix" deadline issues — fix the deadline instead

```go
// Audit log must complete even if request is canceled
func handler(w http.ResponseWriter, r *http.Request) {
    bgCtx := context.WithoutCancel(r.Context())
    go auditLog(bgCtx, r)
}
```

---

## errors.ErrUnsupported (Go 1.20+)

Standardized sentinel error for "not supported" operations. Syscall errors like `ENOSYS`, `ENOTSUP` match via `errors.Is` (Go 1.21+).

### When to use
- Return from interface methods when an operation is not implemented
- Checking if an operation is unsupported across platforms

```go
func (fs readOnlyFS) Remove(name string) error {
    return fmt.Errorf("remove %s: %w", name, errors.ErrUnsupported)
}

// Caller
if errors.Is(err, errors.ErrUnsupported) {
    // fall back to alternative approach
}
```

---

## Migration strategy

1. Run `go fix ./...` (Go 1.26+) to auto-migrate `errors.As` → `errors.AsType`
2. Replace manual error list accumulation with `errors.Join`
3. Add `context.WithCancelCause` to shutdown paths where cause matters
4. Replace `go func() { <-ctx.Done(); ... }()` with `context.AfterFunc`
