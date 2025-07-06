# COERCE

Type-safe, panic-free conversion helpers for modern Go.

## Usage Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/kaptinlin/gozod/pkg/coerce"
)

func main() {
    raw := map[string]any{"port": "8080", "debug": "on", "timeout": 2.5}

    // Integer conversion
    port, err := coerce.ToInteger[int](raw["port"])
    if err != nil { log.Fatal(err) }

    // Boolean conversion
    debug, _ := coerce.ToBool(raw["debug"])
    // Float conversion
    timeout, _ := coerce.ToFloat64(raw["timeout"])

    fmt.Printf("port=%d debug=%t timeout=%.1fs\n", port, debug, timeout)
}
```

## Quick Reference

```go
import "github.com/kaptinlin/gozod/pkg/coerce"

// Generic conversion
coerce.To[T](val)              // convert to type T
coerce.ToLiteral(val)           // literal value conversion

// String & time
coerce.ToString(val)            // string conversion
coerce.ToTime(val)              // time.Time conversion

// Boolean
coerce.ToBool(val)              // bool conversion

// Numeric (scalar)
coerce.ToInt64(val)             // int64 conversion
coerce.ToFloat64(val)           // float64 conversion
coerce.ToBigInt(val)            // *big.Int conversion

// Numeric (generic)
coerce.ToInteger[I](val)        // integer type I
coerce.ToFloat[F](val)          // float type F

// Complex
coerce.ToComplex64(val)         // complex64
coerce.ToComplex128(val)        // complex128
coerce.ToComplexFromString(str) // parse complex from string
```

