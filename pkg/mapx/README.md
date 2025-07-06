# MapX

Lightweight helpers for working with Go maps (and structs) in a type-safe way.

## Features

* **Basic operations** – `Get`, `Set`, `Has`, `Count`, `Copy`, `Merge`
* **Typed accessors** – `GetString`, `GetBool`, `GetInt`, `GetFloat64`, `GetStrings`, `GetAnySlice`, `GetMap`
* **Default helpers** – `GetXXXDefault` (string, bool, int, float64, slice, map, any)
* **Object helpers** – `Keys` extracts field names from maps **or structs** via reflection
* **Conversion** – `ToGeneric` converts any map to `map[any]any`

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/kaptinlin/gozod/pkg/mapx"
)

func main() {
    cfg := map[string]any{
        "app":  "demo",
        "port": 3000,
    }

    // Read values safely
    app, _ := mapx.GetString(cfg, "app")   // "demo"
    port := mapx.GetIntDefault(cfg, "port", 80)

    // Write values
    mapx.Set(cfg, "debug", true)

    // Copy & merge
    clone := mapx.Copy(cfg)
    merged := mapx.Merge(cfg, map[string]any{"port": 8080})

    fmt.Println(app, port, clone, merged)
}
```

## API Cheat-Sheet

| Category       | Functions |
|----------------|-----------|
| Basic ops      | `Get`, `Set`, `Has`, `Count`, `Copy`, `Merge` |
| Accessors      | `GetString`, `GetBool`, `GetInt`, `GetFloat64`, `GetStrings`, `GetAnySlice`, `GetMap` |
| Defaults       | `GetStringDefault`, `GetBoolDefault`, `GetIntDefault`, `GetFloat64Default`, `GetStringsDefault`, `GetAnySliceDefault`, `GetMapDefault`, `GetAnyDefault` |
| Object helpers | `Keys` |
| Convert        | `ToGeneric` |
