# Error Handling Patterns

This document provides comprehensive error handling strategies for Golang libraries with real-world examples.

## Table of Contents

1. [Sentinel Errors](#sentinel-errors)
2. [Structured Errors](#structured-errors)
3. [Error Wrapping](#error-wrapping)
4. [Policy-Based Error Handling](#policy-based-error-handling)
5. [Error Aggregation](#error-aggregation)
6. [Context-Aware Errors](#context-aware-errors)

---

## Sentinel Errors

**When to use:** Simple error conditions, error identity checks, public API errors

**Characteristics:**
- Package-level variables
- Use `errors.New()` for simple messages
- Check with `errors.Is()`
- Immutable and comparable

### Basic Pattern

```go
package cache

import "errors"

// Define sentinel errors at package level
var (
    ErrNotFound   = errors.New("key not found")
    ErrClosed     = errors.New("cache closed")
    ErrInvalidKey = errors.New("invalid key")
    ErrExpired    = errors.New("key expired")
)

// Usage
func (c *Cache) Get(key string) ([]byte, error) {
    if c.closed {
        return nil, ErrClosed
    }

    if key == "" {
        return nil, ErrInvalidKey
    }

    value, exists := c.data[key]
    if !exists {
        return nil, ErrNotFound
    }

    return value, nil
}

// Caller checks with errors.Is
if errors.Is(err, cache.ErrNotFound) {
    // Handle not found
}
```

### Real-World Example: go-fsm

```go
package fsm

var (
    ErrNoRules        = errors.New("no rules defined")
    ErrInvalidInitial = errors.New("initial state not in rules")
    ErrFinalOutgoing  = errors.New("final state has outgoing edge")
    ErrNoTransition   = errors.New("no transition for event")
    ErrGuardFailed    = errors.New("guard condition failed")
)

func (b *Builder[S, E]) Build() (*Machine[S, E], error) {
    if len(b.rules) == 0 {
        return nil, ErrNoRules
    }

    for _, r := range b.rules {
        if b.finals[r.from] {
            return nil, fmt.Errorf("%w: state %v", ErrFinalOutgoing, r.from)
        }
    }

    return &Machine[S, E]{...}, nil
}
```

---

## Structured Errors

**When to use:** Errors with additional context, programmatic error inspection, error codes

**Characteristics:**
- Custom error types with fields
- Implement `Error() string` method
- Implement `Unwrap() error` for error chains
- Check with `errors.As()`

### Basic Pattern

```go
// Structured error type
type ValidationError struct {
    Field   string
    Value   any
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Usage
func Validate(user User) error {
    if user.Email == "" {
        return &ValidationError{
            Field:   "email",
            Value:   user.Email,
            Message: "email is required",
        }
    }
    return nil
}

// Caller checks with errors.As
var valErr *ValidationError
if errors.As(err, &valErr) {
    log.Printf("Field: %s, Message: %s", valErr.Field, valErr.Message)
}
```

### Advanced: Error with Unwrap

```go
type HTTPError struct {
    StatusCode int
    URL        string
    Err        error  // Wrapped error
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s: %v", e.StatusCode, e.URL, e.Err)
}

func (e *HTTPError) Unwrap() error {
    return e.Err
}

// Usage
func Fetch(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return &HTTPError{
            StatusCode: 0,
            URL:        url,
            Err:        err,
        }
    }

    if resp.StatusCode != 200 {
        return &HTTPError{
            StatusCode: resp.StatusCode,
            URL:        url,
            Err:        fmt.Errorf("unexpected status"),
        }
    }

    return nil
}

// Caller can check both HTTPError and wrapped error
var httpErr *HTTPError
if errors.As(err, &httpErr) {
    log.Printf("Status: %d, URL: %s", httpErr.StatusCode, httpErr.URL)
}

if errors.Is(err, context.DeadlineExceeded) {
    log.Println("Request timed out")
}
```

### Real-World Example: go-secrets

```go
type Error struct {
    Code    ErrorCode
    Message string
    Scope   string
    Name    string
    Err     error
}

type ErrorCode int

const (
    ErrCodeNotFound ErrorCode = iota
    ErrCodeInvalidKey
    ErrCodeClosed
    ErrCodePermission
)

func (e *Error) Error() string {
    if e.Scope != "" && e.Name != "" {
        return fmt.Sprintf("%s: %s/%s: %v", e.Code, e.Scope, e.Name, e.Message)
    }
    return fmt.Sprintf("%s: %v", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
    return e.Err
}

func (e ErrorCode) String() string {
    switch e {
    case ErrCodeNotFound:
        return "not_found"
    case ErrCodeInvalidKey:
        return "invalid_key"
    case ErrCodeClosed:
        return "closed"
    case ErrCodePermission:
        return "permission_denied"
    default:
        return "unknown"
    }
}
```

---

## Error Wrapping

**When to use:** Adding context to errors, preserving error chains, debugging

**Characteristics:**
- Use `fmt.Errorf()` with `%w` verb
- Preserves error identity for `errors.Is()`
- Preserves error type for `errors.As()`
- Creates error chains

### Basic Pattern

```go
func ProcessFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("read file %s: %w", path, err)
    }

    result, err := Parse(data)
    if err != nil {
        return fmt.Errorf("parse file %s: %w", path, err)
    }

    if err := Validate(result); err != nil {
        return fmt.Errorf("validate file %s: %w", path, err)
    }

    return nil
}

// Error chain example:
// validate file config.json: validation failed for email: email is required
```

### Multiple Error Wrapping (Go 1.20+)

```go
import "errors"

// Join multiple errors
func ValidateAll(users []User) error {
    var errs []error

    for i, user := range users {
        if err := Validate(user); err != nil {
            errs = append(errs, fmt.Errorf("user %d: %w", i, err))
        }
    }

    if len(errs) > 0 {
        return errors.Join(errs...)
    }

    return nil
}

// Error output:
// user 0: email is required
// user 2: invalid phone number
// user 5: age must be positive
```

### Real-World Example: go-config

```go
func (c *Config[T]) Load(ctx context.Context, sources ...Source) error {
    merged := make(map[string]any)

    for i, source := range sources {
        data, err := source.Load(ctx)
        if err != nil {
            // Check if source has optional policy
            if sr, ok := source.(SkipReporter); ok {
                if sr.SkipError() != nil {
                    continue  // Skip optional source
                }
            }
            return fmt.Errorf("load source %d: %w", i, err)
        }

        if err := merge(merged, data); err != nil {
            return fmt.Errorf("merge source %d: %w", i, err)
        }
    }

    result, err := decode[T](merged)
    if err != nil {
        return fmt.Errorf("decode config: %w", err)
    }

    c.snapshot.Store(&result)
    return nil
}
```

---

## Policy-Based Error Handling

**When to use:** Optional vs required operations, graceful degradation, configurable error behavior

**Characteristics:**
- Policy enum (Required, Optional, etc.)
- Error absorption with reporting
- Configurable at construction time

### Basic Pattern

```go
type Policy int

const (
    PolicyRequired Policy = iota  // Error is fatal
    PolicyOptional                // Error is recorded, operation skipped
    PolicyBestEffort              // Error is logged, continue with partial results
)

type Config struct {
    Policy Policy
    skipErr error  // Absorbed error for provenance
}

type SkipReporter interface {
    SkipError() error
}

func (c *Config) SkipError() error {
    return c.skipErr
}

// Usage
func Load(sources []Source) (map[string]any, error) {
    result := make(map[string]any)

    for _, source := range sources {
        data, err := source.Load()
        if err != nil {
            // Check if source is optional
            if sr, ok := source.(SkipReporter); ok {
                if sr.SkipError() != nil {
                    continue  // Skip optional source
                }
            }
            return nil, err  // Required source failed
        }

        merge(result, data)
    }

    return result, nil
}
```

### Real-World Example: go-config

```go
package file

type Policy int

const (
    PolicyRequired Policy = iota
    PolicyOptional
)

type File struct {
    path    string
    policy  Policy
    skipErr error
}

type Option func(*File)

func WithPolicy(p Policy) Option {
    return func(f *File) {
        f.policy = p
    }
}

func New(path string, opts ...Option) *File {
    f := &File{
        path:   path,
        policy: PolicyRequired,
    }

    for _, opt := range opts {
        opt(f)
    }

    return f
}

func (f *File) Load(ctx context.Context) (map[string]any, error) {
    data, err := os.ReadFile(f.path)
    if err != nil {
        if f.policy == PolicyOptional {
            f.skipErr = err  // Record error for provenance
            return nil, err  // Return error but caller will skip
        }
        return nil, fmt.Errorf("read file: %w", err)
    }

    // Decode and return
    return decode(data)
}

func (f *File) SkipError() error {
    if f.policy == PolicyOptional {
        return f.skipErr
    }
    return nil
}

// Usage
cfg := config.New[AppConfig](
    file.New("config.yaml"),                                    // Required
    file.New("config.local.yaml", file.WithPolicy(file.PolicyOptional)),  // Optional
    env.New("APP"),                                             // Required
)
```

---

## Error Aggregation

**When to use:** Batch operations, validation of multiple items, parallel processing

**Characteristics:**
- Collect errors during processing
- Use `errors.Join()` (Go 1.20+)
- Return all errors, not just first

### Basic Pattern (Go 1.20+)

```go
func ValidateBatch(items []Item) error {
    var errs []error

    for i, item := range items {
        if err := Validate(item); err != nil {
            errs = append(errs, fmt.Errorf("item %d: %w", i, err))
        }
    }

    if len(errs) > 0 {
        return errors.Join(errs...)
    }

    return nil
}

// Caller can inspect all errors
if err != nil {
    for _, e := range strings.Split(err.Error(), "\n") {
        log.Println(e)
    }
}
```

### Advanced: Parallel Error Collection

```go
func ProcessParallel(items []Item) error {
    var (
        wg   sync.WaitGroup
        mu   sync.Mutex
        errs []error
    )

    for i, item := range items {
        wg.Add(1)
        go func(idx int, it Item) {
            defer wg.Done()

            if err := Process(it); err != nil {
                mu.Lock()
                errs = append(errs, fmt.Errorf("item %d: %w", idx, err))
                mu.Unlock()
            }
        }(i, item)
    }

    wg.Wait()

    if len(errs) > 0 {
        return errors.Join(errs...)
    }

    return nil
}
```

### Real-World Example: go-fsm

```go
func (b *Builder[S, E]) Build() (*Machine[S, E], error) {
    var errs []error

    // Validate: at least one rule
    if len(b.rules) == 0 {
        errs = append(errs, ErrNoRules)
    }

    // Validate: initial state exists
    hasInitial := false
    for _, r := range b.rules {
        if r.from == b.initial {
            hasInitial = true
            break
        }
    }
    if !hasInitial {
        errs = append(errs, ErrInvalidInitial)
    }

    // Validate: final states have no outgoing edges
    for _, r := range b.rules {
        if b.finals[r.from] {
            errs = append(errs, fmt.Errorf("%w: state %v", ErrFinalOutgoing, r.from))
        }
    }

    if len(errs) > 0 {
        return nil, errors.Join(errs...)
    }

    // Compile to runtime
    return &Machine[S, E]{...}, nil
}
```

---

## Context-Aware Errors

**When to use:** Request tracing, debugging, error reporting with context

**Characteristics:**
- Include request ID, user ID, trace ID
- Add context without changing error identity
- Use structured logging for context

### Basic Pattern

```go
type contextKey int

const (
    requestIDKey contextKey = iota
    userIDKey
)

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}

// Add context to errors
func Process(ctx context.Context, item Item) error {
    if err := Validate(item); err != nil {
        requestID := GetRequestID(ctx)
        return fmt.Errorf("request %s: validate item: %w", requestID, err)
    }
    return nil
}
```

### Real-World Example: go-crawler

```go
func (p *Pipeline) Acquire(ctx context.Context, url string) (Document, error) {
    // Extract trace ID from context
    traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

    page, err := p.fetcher.Fetch(ctx, url)
    if err != nil {
        return Document{}, fmt.Errorf("trace %s: fetch %s: %w", traceID, url, err)
    }

    content, err := p.extractor.Extract(ctx, page)
    if err != nil {
        // Fallback: full-page conversion
        return p.converter.ConvertRaw(ctx, page)
    }

    doc, err := p.converter.Convert(ctx, content)
    if err != nil {
        return Document{}, fmt.Errorf("trace %s: convert %s: %w", traceID, url, err)
    }

    return doc, nil
}
```

---

## Error Handling Anti-Patterns

### Anti-Pattern 1: Ignoring Errors

```go
// BAD: Silent failure
data, _ := os.ReadFile(path)

// GOOD: Handle or propagate
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("read file: %w", err)
}
```

### Anti-Pattern 2: Generic Error Messages

```go
// BAD: No context
return errors.New("failed")

// GOOD: Specific context
return fmt.Errorf("parse config file %s: %w", path, err)
```

### Anti-Pattern 3: Panic in Libraries

```go
// BAD: Panic in library code
func Get(key string) []byte {
    if key == "" {
        panic("empty key")
    }
    return data[key]
}

// GOOD: Return error
func Get(key string) ([]byte, error) {
    if key == "" {
        return nil, ErrInvalidKey
    }
    return data[key], nil
}
```

### Anti-Pattern 4: Losing Error Context

```go
// BAD: Losing original error
if err != nil {
    return errors.New("operation failed")
}

// GOOD: Preserve error chain
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

---

## Pattern Selection Guide

| Use Case | Recommended Pattern | Example |
|----------|-------------------|---------|
| Simple error conditions | Sentinel Errors | `ErrNotFound`, `ErrClosed` |
| Errors with context | Structured Errors | `ValidationError`, `HTTPError` |
| Adding context to errors | Error Wrapping | `fmt.Errorf("parse: %w", err)` |
| Optional operations | Policy-Based | `PolicyRequired`, `PolicyOptional` |
| Batch operations | Error Aggregation | `errors.Join(errs...)` |
| Request tracing | Context-Aware | `fmt.Errorf("request %s: %w", id, err)` |

**Mix and match patterns as needed.** Most libraries use multiple patterns (e.g., go-config uses Sentinel + Wrapping + Policy-Based).

---

## Checklist for Error Handling

- [ ] All errors have clear, actionable messages
- [ ] Error chains preserve original error with `%w`
- [ ] Public errors are documented in package godoc
- [ ] Sentinel errors are package-level variables
- [ ] Structured errors implement `Error()` and `Unwrap()`
- [ ] Library code returns errors instead of panicking
- [ ] Error messages include relevant context (file path, key, URL)
- [ ] Batch operations collect all errors, not just first
- [ ] Optional operations use policy-based error handling
- [ ] Error handling is consistent across the package
