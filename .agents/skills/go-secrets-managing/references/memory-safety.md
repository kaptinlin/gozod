# Memory Safety

go-secrets provides memory safety for secret values through the `Value` type, which wraps plaintext with controlled access and zeroing semantics.

## Value Type

```go
type Value struct {
    buf []byte
}
```

| Method | Returns | Purpose |
|--------|---------|---------|
| `Expose()` | `[]byte` | Raw plaintext bytes (shares underlying buffer) |
| `ExposeString()` | `string` | Plaintext as Go string |
| `String()` | `"[REDACTED]"` | Safe for logging — never exposes plaintext |
| `Close()` | — | Zeros the buffer using `clear()` (Go 1.21+) |

## The Close Pattern

Every `Get` call returns a `Value` that MUST be closed to zero plaintext from memory:

```go
val, err := s.Get(ctx, "prod", "db_password")
if err != nil {
    return err
}
defer val.Close()  // zeros plaintext buffer

// Use the value
dsn := fmt.Sprintf("postgres://app:%s@db:5432/mydb", val.ExposeString())
```

### What Close Does

```go
func (v *Value) Close() {
    clear(v.buf)  // Go 1.21+ built-in: sets all bytes to zero
}
```

The `clear()` built-in zeroes the entire byte slice, preventing plaintext from lingering in memory after use.

### Why It Matters

Without `Close()`:
- Plaintext remains in the process's memory until garbage collected
- GC timing is non-deterministic — plaintext may persist for extended periods
- Memory dumps, core dumps, or swap files could expose secrets
- In long-running services, accumulated unclosed values increase exposure window

## Safe Logging

`Value.String()` implements `fmt.Stringer` to prevent accidental plaintext exposure in logs:

```go
val, _ := s.Get(ctx, "prod", "api_key")
defer val.Close()

// Safe: logs [REDACTED]
logger.Info("loaded secret", slog.String("value", val.String()))
fmt.Printf("secret: %s\n", val)  // prints: secret: [REDACTED]

// Dangerous: logs actual plaintext — only use when needed
dsn := val.ExposeString()
```

### Logging Rules

| Method | Log Safe | When to Use |
|--------|----------|-------------|
| `val.String()` | Yes | Logging, debugging, error messages |
| `val.ExposeString()` | **No** | Passing to database drivers, API calls, crypto operations |
| `val.Expose()` | **No** | Byte-level operations, hashing, key derivation |
| `fmt.Sprintf("%s", val)` | Yes | Uses `String()` → `[REDACTED]` |
| `fmt.Sprintf("%v", val)` | Yes | Uses `String()` → `[REDACTED]` |

## Internal Memory Management

### Value Creation (inside Secrets.Get)

```go
// Simplified flow inside Secrets.Get:
raw, meta, err := s.store.Get(ctx, scope, name)   // encrypted bytes from store
plaintext, err := s.cipher.Decrypt(raw, aad)       // decrypt
val := secrets.NewValue(bytes.Clone(plaintext))     // clone into new Value
clear(plaintext)                                     // zero decrypted buffer
return val, nil
```

Key properties:
- `NewValue()` clones the input bytes — caller's buffer is independent
- Decrypted plaintext is zeroed immediately after cloning into Value
- Each `Get` returns a fresh, independent Value

### Collect (bulk retrieval)

```go
all, err := s.Collect(ctx, "prod")
// Returns map[string]string — each value is a plaintext string
// Values are closed internally after string extraction
```

`Collect` handles `Close()` automatically for each value, but the returned strings are standard Go strings (not zeroed on discard). Use `Get` + `Close` for maximum memory safety.

## Patterns

### Short-Lived Use

```go
val, err := s.Get(ctx, "prod", "db_password")
if err != nil {
    return err
}
defer val.Close()

db, err := sql.Open("postgres", fmt.Sprintf("...%s...", val.ExposeString()))
```

### Pass to Function, Then Close

```go
val, err := s.Get(ctx, "prod", "hmac_key")
if err != nil {
    return err
}
defer val.Close()

mac := hmac.New(sha256.New, val.Expose())
mac.Write(message)
```

### Context-Based Retrieval

```go
ctx = secrets.NewContext(ctx, s)
ctx = secrets.WithScope(ctx, "prod")

val, err := secrets.FromContext(ctx, "api_key")
if err != nil {
    return err
}
defer val.Close()
```

### Testing Without Encryption

```go
func newTestSecrets(t *testing.T) *secrets.Secrets {
    t.Helper()
    s, err := secrets.New(
        secrets.WithStore(memstore.New()),
        secrets.WithNoEncryption(),
    )
    require.NoError(t, err)
    t.Cleanup(func() { s.Close() })
    return s
}
```

## Anti-Patterns

### Forgetting Close

```go
// BAD: plaintext lingers in memory
val, _ := s.Get(ctx, "prod", "key")
password := val.ExposeString()
// val never closed — plaintext stays until GC
```

### Logging ExposeString

```go
// BAD: plaintext in log files
val, _ := s.Get(ctx, "prod", "key")
defer val.Close()
logger.Info("got secret", slog.String("value", val.ExposeString()))

// GOOD: redacted
logger.Info("got secret", slog.String("value", val.String()))
```

### Storing Value Beyond Scope

```go
// BAD: Value outlives defer scope
var saved *secrets.Value
func loadSecret(s *secrets.Secrets) {
    val, _ := s.Get(ctx, "prod", "key")
    saved = val  // escaped — who calls Close()?
}

// GOOD: extract what you need, close immediately
func loadDSN(s *secrets.Secrets) (string, error) {
    val, err := s.Get(ctx, "prod", "db_password")
    if err != nil {
        return "", err
    }
    defer val.Close()
    return fmt.Sprintf("postgres://app:%s@db:5432/mydb", val.ExposeString()), nil
}
```

## Limitations

- Go strings are immutable and cannot be zeroed — `ExposeString()` creates an uncontrolled copy
- `Collect()` returns `map[string]string` where strings cannot be zeroed
- GC may move Value's buffer before `Close()` is called, leaving copies in memory
- These are inherent Go runtime limitations; `Value.Close()` provides best-effort protection
- For maximum security, keep the window between `Get` and `Close` as short as possible
