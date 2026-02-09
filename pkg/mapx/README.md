# MapX

Lightweight helpers for working with Go maps (and structs) in a type-safe way.

## Features

* **Generic accessors** – `ValueOf[T]`, `ValueOrDefault[T]` for any type
* **Basic operations** – `Get`, `Set`, `Has`, `Count`, `Copy`, `Merge`
* **Typed accessors** – `GetString`, `GetBool`, `GetInt`, `GetFloat64`, `GetStrings`, `GetAnySlice`, `GetMap`
* **Default helpers** – `GetXxxDefault` variants for all typed accessors
* **Object helpers** – `Keys` extracts field names from maps **or structs**
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

    // Generic accessors (preferred)
    app, _ := mapx.ValueOf[string](cfg, "app")   // "demo"
    port := mapx.ValueOrDefault(cfg, "port", 80)  // 3000

    // Typed convenience wrappers
    debug := mapx.GetBoolDefault(cfg, "debug", false)

    // Copy & merge
    clone := mapx.Copy(cfg)
    merged := mapx.Merge(cfg, map[string]any{"port": 8080})

    fmt.Println(app, port, debug, clone, merged)
}
```

## API Cheat-Sheet

| Category       | Functions |
|----------------|-----------|
| Generic        | `ValueOf[T]`, `ValueOrDefault[T]` |
| Basic ops      | `Get`, `Set`, `Has`, `Count`, `Copy`, `Merge` |
| Accessors      | `GetString`, `GetBool`, `GetInt`, `GetFloat64`, `GetStrings`, `GetAnySlice`, `GetMap` |
| Defaults       | `GetStringDefault`, `GetBoolDefault`, `GetIntDefault`, `GetFloat64Default`, `GetStringsDefault`, `GetAnySliceDefault`, `GetMapDefault`, `GetAnyDefault` |
| Object helpers | `Keys` |
| Convert        | `ToGeneric` |
