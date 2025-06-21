# MapX

Type-safe, modern helpers for working with Go maps and structs.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/gozod/pkg/mapx"
)

func main() {
    cfg := map[string]any{
        "app":     "demo",
        "port":    3000,
        "features": []string{"metrics", "cache"},
        "database": map[string]any{
            "host": "localhost",
            "port": 5432,
        },
    }

    // Get value with type assertion
    name, _ := mapx.GetString(cfg, "app") // "demo"
    port := mapx.GetIntDefault(cfg, "port", 80) // 3000

    // Set and check
    mapx.Set(cfg, "debug", true)
    exists := mapx.Has(cfg, "app") // true

    // Merge and copy
    copy := mapx.Copy(cfg)
    merged := mapx.Merge(cfg, copy)

    // Nested map access
    db := mapx.GetMapDefault(cfg, "database", nil)
    dbHost := mapx.GetStringDefault(db, "host", "127.0.0.1")
    fmt.Printf("%s on :%d, DB %s\n", name, port, dbHost)
}
```

## Quick Reference

```go
import "github.com/kaptinlin/gozod/pkg/mapx"

// Type check
mapx.Is(val)              // is map or struct
mapx.IsStringKey(val)     // is map[string]any

// Basic operations
mapx.Get(m, key)          // get value
mapx.Set(m, key, val)     // set value
mapx.Has(m, key)          // has key
mapx.Count(m)             // number of keys
mapx.Copy(m)              // shallow copy
mapx.Merge(m1, m2)        // merge maps

// Accessors
mapx.GetString(m, key)    // (string, ok)
mapx.GetInt(m, key)       // (int, ok)
mapx.GetBool(m, key)      // (bool, ok)
mapx.GetStrings(m, key)   // ([]string, ok)
mapx.GetMap(m, key)       // (map[string]any, ok)

// Defaults
mapx.GetStringDefault(m, key, def) // string or default
mapx.GetIntDefault(m, key, def)    // int or default
mapx.GetMapDefault(m, key, def)    // map or default

// Object helpers
mapx.Keys(m)              // []string
mapx.Value(m, key)        // (any, ok)

// Conversion
mapx.ToGeneric(val)       // map[any]any
mapx.ToStringKey(val)     // map[string]any
mapx.FromAny(val)         // map[string]any

// Merge
mapx.MergeMaps(m1, m2)    // merge arbitrary maps

// Extract
mapx.Extract(val)         // map[string]any
mapx.ExtractRecord(val)   // map[string]any

// Validation
mapx.ValidateType(val, typeStr) // bool
```

## API Cheat-Sheet

| Category     | Functions                                                      |
|--------------|----------------------------------------------------------------|
| Type check   | `Is` `IsStringKey`                                             |
| Basic ops    | `Get` `Set` `Has` `Count` `Copy` `Merge`                       |
| Accessors    | `GetString` `GetBool` `GetInt` `GetFloat64` `GetStrings` `GetMap` |
| Defaults     | `GetXXXDefault`                                                |
| Object       | `Keys` `Value`                                                 |
| Convert      | `ToGeneric` `ToStringKey` `FromAny`                            |
| Merge        | `MergeMaps`                                                    |
| Extract      | `Extract` `ExtractRecord`                                      |
| Validate     | `ValidateType`                                                 |
