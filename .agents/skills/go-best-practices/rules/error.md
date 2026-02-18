# Error Handling

Go's explicit error handling is a core language feature. Proper error structure, wrapping, and propagation directly affect debuggability and system reliability.

## Contents
- Use Structured Errors
- Add Non-Redundant Context
- Choose %v vs %w for Wrapping
- Place %w at End of Error String
- Return error Interface
- Handle Errors Explicitly
- Indent Error Flow
- Avoid In-Band Error Values
- Error Logging Best Practices

---

## Use Structured Errors

When callers need to distinguish error conditions, use structured errors (sentinel values or typed errors) instead of string matching.

**Incorrect:**

```go
func handlePet(...) {
    err := process(an)
    if regexp.MatchString(`duplicate`, err.Error()) {...}
    if regexp.MatchString(`marsupial`, err.Error()) {...}
}
```

**Correct:**

```go
var (
    ErrDuplicate = errors.New("duplicate")
    ErrMarsupial = errors.New("marsupials are not supported")
)

func handlePet(...) {
    switch err := process(an); {
    case errors.Is(err, ErrDuplicate):
        return fmt.Errorf("feed %q: %v", an, err)
    case errors.Is(err, ErrMarsupial):
        alternate = an.BackupAnimal()
        return handlePet(..., alternate, ...)
    }
}
```

Use `os.PathError`-style types when callers need programmatic access to error fields.

---

## Add Non-Redundant Context

When wrapping errors, add information the caller's context provides that the underlying error doesn't. Avoid restating what the error already says.

**Incorrect (duplicates path already in os error):**

```go
if err := os.Open("settings.txt"); err != nil {
    return fmt.Errorf("could not open settings.txt: %v", err)
}
// Output: could not open settings.txt: open settings.txt: no such file or directory
```

**Correct (adds meaningful context):**

```go
if err := os.Open("settings.txt"); err != nil {
    return fmt.Errorf("launch codes unavailable: %v", err)
}
// Output: launch codes unavailable: open settings.txt: no such file or directory
```

**Incorrect:** `return fmt.Errorf("failed: %v", err)` — just `return err`.

---

## Choose %v vs %w for Wrapping

Use `%v` for simple annotation or at system boundaries. Use `%w` when callers need programmatic error inspection via `errors.Is`/`errors.As`.

**Use `%v` at system boundaries:**

```go
func (*FortuneTeller) SuggestFortune(ctx context.Context, req *pb.SuggestionRequest) (*pb.SuggestionResponse, error) {
    if err != nil {
        return nil, fmt.Errorf("couldn't find fortune database: %v", err)
    }
}
```

**Use `%w` for internal error chains:**

```go
func (s *Server) internalFunction(ctx context.Context) error {
    if err != nil {
        return fmt.Errorf("couldn't find remote file: %w", err)
    }
}
```

At system boundaries (RPC, IPC, storage), prefer converting to canonical error spaces (e.g., gRPC status codes) over wrapping with `%w`.

---

## Place %w at End of Error String

When wrapping errors with `%w`, place it at the end using the `[...]: %w` pattern so the printed chain reads newest-to-oldest.

**Incorrect:**

```go
err2 := fmt.Errorf("%w: err2", err1)
err3 := fmt.Errorf("%w: err3", err2)
fmt.Println(err3) // err1: err2: err3 (oldest-to-newest, confusing)
```

**Correct:**

```go
err2 := fmt.Errorf("err2: %w", err1)
err3 := fmt.Errorf("err3: %w", err2)
fmt.Println(err3) // err3: err2: err1 (newest-to-oldest, natural)
```

---

## Return error Interface

Exported functions should return `error` interface, not concrete error types. A concrete nil pointer wrapped in an interface becomes non-nil.

**Incorrect:**

```go
func Bad() *os.PathError { /*...*/ }
```

**Correct:**

```go
func Good() error { /*...*/ }
```

Error strings should not be capitalized (unless starting with a proper noun) and should not end with punctuation:

```go
// Bad:
err := fmt.Errorf("Something bad happened.")

// Good:
err := fmt.Errorf("something bad happened")
```

---

## Handle Errors Explicitly

Never silently discard errors with `_`. Either handle them, return them, or in rare cases `log.Fatal`. Document why if you intentionally ignore one.

**Incorrect:**

```go
f, _ := os.Open(filename)
```

**Correct:**

```go
f, err := os.Open(filename)
if err != nil {
    return nil, err
}
```

**Acceptable (documented intentional ignore):**

```go
var b *bytes.Buffer
n, _ := b.Write(p) // never returns a non-nil error
```

---

## Indent Error Flow

Handle errors first and return early. Keep the success path at the lowest indentation level.

**Incorrect:**

```go
if err != nil {
    // error handling
} else {
    // normal code that looks abnormal due to indentation
}
```

**Correct:**

```go
if err != nil {
    // error handling
    return // or continue
}
// normal code
```

For variables used across multiple lines, prefer explicit declaration over if-with-initializer:

```go
x, err := f()
if err != nil {
    return
}
// lots of code that uses x
```

---

## Avoid In-Band Error Values

Don't return special values like -1, "", or nil to signal errors. Use Go's multiple return values with `error` or `bool`.

**Incorrect:**

```go
// Lookup returns the value for key or -1 if there is no mapping for key.
func Lookup(key string) int
```

**Correct:**

```go
// Lookup returns the value for key or ok=false if there is no mapping for key.
func Lookup(key string) (value string, ok bool)
```

---

## Error Logging Best Practices

Log errors at the right level and avoid double-logging. If you return an error, let the caller log it.

Key principles:
- If returning an error, don't also log it — let caller decide
- Be cautious with PII in log sinks
- Use `log.Error` sparingly — ERROR level causes flushes with performance impact
- Use verbose levels: `V(1)` for extra info, `V(2)` for tracing, `V(3)` for state dumps

**Incorrect (expensive call even when log is disabled):**

```go
log.V(2).Infof("Handling %v", sql.Explain())
```

**Correct (guard expensive calls):**

```go
if log.V(2) {
    log.Infof("Handling %v", sql.Explain())
}
```

For program initialization errors, propagate to `main` and use `log.Exit` with a human-readable message rather than `log.Fatal` with a stack trace.
