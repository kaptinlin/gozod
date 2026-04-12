# MapX

Lightweight helpers for working with Go maps (and structs) in a type-safe way.

## Features

* **Generic accessors** – `ValueOf[T]`, `ValueOr[T]` for any type
* **Basic operations** – `Get`, `Set`, `Has`, `Count`, `Copy`, `Merge`
* **Typed accessors** – `String`, `Bool`, `IntCoerce`, `Float64Coerce`, `Strings`, `AnySlice`, `Map`
* **Default helpers** – `StringOr`, `BoolOr`, `IntCoerceOr`, `Float64CoerceOr`, `StringsOr`, `AnySliceOr`, `MapOr`, `AnyOr`
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
    port := mapx.ValueOr(cfg, "port", 80)  // 3000

    // Typed convenience wrappers
    debug := mapx.BoolOr(cfg, "debug", false)

    // Copy & merge
    clone := mapx.Copy(cfg)
    merged := mapx.Merge(cfg, map[string]any{"port": 8080})

    fmt.Println(app, port, debug, clone, merged)
}
```

## API Cheat-Sheet

| Category       | Functions |
|----------------|-----------|
| Generic        | `ValueOf[T]`, `ValueOr[T]` |
| Basic ops      | `Get`, `Set`, `Has`, `Count`, `Copy`, `Merge` |
| Accessors      | `String`, `Bool`, `IntCoerce`, `Float64Coerce`, `Strings`, `AnySlice`, `Map` |
| Defaults       | `StringOr`, `BoolOr`, `IntCoerceOr`, `Float64CoerceOr`, `StringsOr`, `AnySliceOr`, `MapOr`, `AnyOr` |
| Object helpers | `Keys` |
| Convert        | `ToGeneric` |
