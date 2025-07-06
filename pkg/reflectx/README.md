# ReflectX

Type-safe, modern helpers for Go reflection and value extraction.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/gozod/pkg/reflectx"
    "github.com/kaptinlin/gozod/pkg/coerce"
)

func main() {
    var value any = "hello world"

    // Type checking
    if reflectx.IsString(value) {
        fmt.Println("Is string type")
    }

    // Value extraction
    if str, ok := reflectx.ExtractString(value); ok {
        fmt.Printf("String value: %s\n", str)
    }

    // Property checking
    if reflectx.HasLength(value) {
        if length, ok := reflectx.GetLength(value); ok {
            fmt.Printf("Length: %d\n", length)
        }
    }

    // Pointer operations
    ptr := &value
    if deref, ok := reflectx.Deref(ptr); ok {
        fmt.Printf("Dereferenced: %v\n", deref)
    }

    // Generic type conversion
    if result, err := reflectx.ConvertToGeneric[string](123); err == nil {
        fmt.Printf("Generic conversion: %s\n", result)
    }

    // Integration with coerce for conversions
    if reflectx.IsNumeric(123) {
        if str, err := coerce.ToString(123); err == nil {
            fmt.Printf("Number as string: %s\n", str)
        }
    }
}
```

## Quick Reference

```go
import "github.com/kaptinlin/gozod/pkg/reflectx"

// Type checking
reflectx.IsString(val)      // true if string
reflectx.IsNumeric(val)     // true if numeric type
reflectx.IsMap(val)         // true if map
reflectx.IsSlice(val)       // true if slice
reflectx.IsNil(val)         // true if nil
reflectx.IsZero(val)        // true if zero value
reflectx.IsPointer(val)     // true if pointer

// Property
reflectx.HasLength(val)     // true if has length property
length, _ := reflectx.GetLength(val)

// Value extraction
reflectx.ExtractString(val) // (string, ok)
reflectx.ExtractInt(val)    // (int, ok)
reflectx.ExtractSlice(val)  // (slice, ok)
reflectx.ExtractMap(val)    // (map, ok)

// Size helpers
size, _ := reflectx.GetSize(val)

// Pointer operations
reflectx.Deref(ptr)         // (value, ok)
reflectx.DerefAll(ptr)      // (value, ok)
reflectx.ToPointer(val)     // pointer to value

// Conversion
reflectx.ConvertToGeneric[T](val) // (T, error)
