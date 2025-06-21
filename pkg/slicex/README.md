# SliceX

Type-safe, modern helpers for slice operations and manipulation in Go.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/gozod/pkg/slicex"
)

func main() {
    // Type check
    vals := []int{1, 2, 3}
    fmt.Println(slicex.Is(vals)) // true

    // Conversion
    anySlice, _ := slicex.ToAny(vals)         // []any{1,2,3}
    intSlice, _ := slicex.ToTyped[int](anySlice) // []int{1,2,3}
    strSlice, _ := slicex.ToStrings(vals)     // []string{"1","2","3"}

    // Extraction
    if s, ok := slicex.Extract(vals); ok {
        fmt.Println(s) // [1 2 3]
    }

    // Operations
    merged, _ := slicex.Merge([]int{1,2}, []int{3,4}) // [1 2 3 4]
    appended, _ := slicex.Append([]int{1,2}, 3, 4)    // [1 2 3 4]
    reversed, _ := slicex.Reverse([]int{1,2,3})       // [3 2 1]
    unique, _ := slicex.Unique([]int{1,2,2,3})        // [1 2 3]

    // Functional
    evens, _ := slicex.Filter([]int{1,2,3,4}, func(v any) bool {
        return v.(int)%2 == 0
    })
    fmt.Println(evens) // [2 4]
}
```

## Quick Reference

```go
import "github.com/kaptinlin/gozod/pkg/slicex"

// Type check
slicex.Is(val)            // is slice
slicex.IsArray(val)       // is array
slicex.IsSliceOrArray(val)// is slice or array

// Conversion
slicex.ToAny(slice)           // []any
slicex.ToTyped[T](slice)      // []T
slicex.ToStrings(slice)       // []string
slicex.StringToChars(str)     // []any (chars)

// Extraction
slicex.Extract(val)           // (slice, ok)
slicex.ExtractArray(val)      // (array, ok)
slicex.ExtractSlice(val)      // (slice, ok)

// Operations
slicex.Merge(a, b)            // merge slices
slicex.Append(slice, vals...) // append
slicex.Prepend(slice, vals...)// prepend

// Utility
slicex.Length(slice)          // length
slicex.IsEmpty(slice)         // bool
slicex.Contains(slice, v)     // bool
slicex.IndexOf(slice, v)      // int
slicex.Reverse(slice)         // reversed
slicex.Unique(slice)          // deduped
slicex.Join(slice, sep)       // string

// Functional
slicex.Filter(slice, fn)      // filter
slicex.Map(slice, fn)         // map/transform
```
