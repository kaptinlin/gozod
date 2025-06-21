# JSONX

Type-safe, modern helpers for JSON validation and inspection in Go.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/gozod/pkg/jsonx"
)

func main() {
    // Validate JSON string
    fmt.Println(jsonx.IsValid(`{"ok": true}`)) // true
    fmt.Println(jsonx.IsPrimitive(`42`))         // true
    fmt.Println(jsonx.IsNumber(`"42"`))        // false (string, not number)
    fmt.Println(jsonx.IsValid([]byte(`[1,2,3]`)))// true (byte slice)
}
```

## Quick Reference

```go
import "github.com/kaptinlin/gozod/pkg/jsonx"

jsonx.IsValid(val)        // true if valid JSON (string, []byte, etc)
jsonx.IsValidString(str)  // true if valid JSON string
jsonx.IsPrimitive(val)    // true if JSON primitive (string/number/bool/null)
jsonx.IsNumber(val)       // true if JSON number
```