# Performance

String concatenation strategies, pre-allocation hints, and value passing optimizations yield measurable improvements in hot paths.

## Contents
- String Concatenation
- Size Hints for Pre-allocation
- Use %q for String Formatting
- Use crypto/rand for Keys
- Use any Instead of interface{}

---

## String Concatenation

Use `+` for simple joins, `fmt.Sprintf` for formatting, `strings.Builder` for incremental building. Write directly to `io.Writer` when possible.

**Simple concatenation — use `+`:**

```go
key := "projectid: " + p
```

**Formatting — use `fmt.Sprintf`:**

```go
// Good:
str := fmt.Sprintf("%s [%s:%d]-> %s", src, qos, mtu, dst)

// Bad:
bad := src.String() + " [" + qos.String() + ":" + strconv.Itoa(mtu) + "]-> " + dst.String()
```

**Incremental building — use `strings.Builder` (O(n) vs O(n^2)):**

```go
b := new(strings.Builder)
for i, d := range digitsOfPi {
    fmt.Fprintf(b, "the %d digit of pi is: %d\n", i, d)
}
str := b.String()
```

When output target is `io.Writer`, use `fmt.Fprintf` directly instead of building a string first. Use backtick literals for multi-line constant strings.

---

## Size Hints for Pre-allocation

When the final size is known, pre-allocate slices and maps. But don't over-allocate — wasted memory can hurt performance too.

**Correct (justified size hints):**

```go
var (
    // Target 131072 as preferred buffer size for filesystem: st_blksize
    buf = make([]byte, 131072)
    // Typically processes 8-10 elements per run (16 is safe assumption)
    q = make([]Node, 0, 16)
    // Each shard processes shardSize (typically 32000+) elements
    seen = make(map[string]bool, shardSize)
)
```

**Incorrect (magic numbers):**

```go
buf = make([]byte, 47000) // Why 47000?
q = make([]Node, 0, 500)  // Why 500?
```

Most code doesn't need size hints. Only pre-allocate when backed by empirical analysis.

---

## Use %q for String Formatting

Use `%q` to print strings in double quotes. It handles empty strings, control characters, and special characters safely.

**Incorrect:**

```go
fmt.Printf("value \"%s\" looks like English text", someText)
```

**Correct:**

```go
fmt.Printf("value %q looks like English text", someText)
```

`%q` makes empty strings visible as `""` rather than invisible whitespace.

---

## Use crypto/rand for Keys

Never use `math/rand` for key generation — even disposable ones. Without seeding, `math/rand` is completely predictable.

**Incorrect:**

```go
import "math/rand"

func Key() string {
    buf := make([]byte, 16)
    rand.Read(buf) // predictable!
    return fmt.Sprintf("%x", buf)
}
```

**Correct:**

```go
import (
    "crypto/rand"
    "fmt"
)

func Key() string {
    buf := make([]byte, 16)
    if _, err := rand.Read(buf); err != nil {
        log.Fatalf("out of randomness: %v", err)
    }
    return fmt.Sprintf("%x", buf)
}
```

---

## Use any Instead of interface{}

Since Go 1.18, `any` is the alias for `interface{}`. Prefer `any` in new code for readability.

**Incorrect:**

```go
func Process(data interface{}) interface{} { ... }
```

**Correct:**

```go
func Process(data any) any { ... }
```
