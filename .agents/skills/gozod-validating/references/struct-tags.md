# Struct Tag Validation Reference

## Contents
- Tag syntax rules
- Complete tag reference by type
- Nested struct and circular reference handling
- Cross-field validation with Refine
- Code generation for performance
- JSON field name mapping

## Tag Syntax

Format: `gozod:"rule1,rule2=value,rule3"`

Core rules:
- Fields are **optional by default** unless `required` is specified
- Use `gozod:"-"` to skip validation for a field entirely
- Comma-separated rules, parameters with `=`
- JSON tag names are used for validation error paths automatically

```go
type Example struct {
    Name    string   `json:"name"    gozod:"required,min=2,max=50"`
    Email   string   `json:"email"   gozod:"required,email"`
    Bio     string   `json:"bio"     gozod:"max=500"`       // optional
    Ignored string   `gozod:"-"`                             // skipped
    Tags    []string `json:"tags"    gozod:"min=1,max=10"`
}

schema := gozod.FromStruct[Example]()
```

## Complete Tag Reference

### String Fields

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Non-empty string required | `gozod:"required"` |
| `min=N` | Minimum string length | `gozod:"min=3"` |
| `max=N` | Maximum string length | `gozod:"max=100"` |
| `length=N` | Exact string length | `gozod:"length=2"` |
| `email` | Email format | `gozod:"email"` |
| `url` | URL format | `gozod:"url"` |
| `uuid` | UUID format | `gozod:"uuid"` |
| `regex=pattern` | Custom regex | `gozod:"regex=^[A-Z][a-z]+$"` |

### Numeric Fields (int, int8-64, uint, uint8-64, float32, float64)

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Non-zero value required | `gozod:"required"` |
| `min=N` | Minimum value | `gozod:"min=0"` |
| `max=N` | Maximum value | `gozod:"max=120"` |
| `positive` | Must be > 0 | `gozod:"positive"` |
| `negative` | Must be < 0 | `gozod:"negative"` |
| `nonnegative` | Must be >= 0 | `gozod:"nonnegative"` |
| `nonpositive` | Must be <= 0 | `gozod:"nonpositive"` |

### Array/Slice Fields

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Must be present | `gozod:"required"` |
| `min=N` | Minimum element count | `gozod:"min=1"` |
| `max=N` | Maximum element count | `gozod:"max=10"` |
| `length=N` | Exact element count | `gozod:"length=5"` |
| `nonempty` | At least one element | `gozod:"nonempty"` |

### Nested Struct Fields

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Nested struct must be present | `gozod:"required"` |

Nested struct fields are validated recursively. Their own `gozod` tags apply.

## Nested Structs

```go
type Address struct {
    Street  string `gozod:"required,min=5"`
    City    string `gozod:"required"`
    ZipCode string `gozod:"required,regex=^\\d{5}$"`
}

type User struct {
    Name    string  `gozod:"required,min=2"`
    Address Address `gozod:"required"` // validated recursively
}

schema := gozod.FromStruct[User]()
// Errors will have paths like ["address", "street"]
```

## Circular References (Automatic)

GoZod auto-detects circular references in struct types and uses lazy evaluation:

```go
type TreeNode struct {
    Value    string      `gozod:"required"`
    Children []*TreeNode `gozod:"max=10"` // circular reference -- handled automatically
}

schema := gozod.FromStruct[TreeNode]() // no stack overflow
```

## Cross-Field Validation

Struct tags cannot express cross-field rules. Use `.Refine()` after `FromStruct`:

```go
type DateRange struct {
    Start string `gozod:"required"`
    End   string `gozod:"required"`
}

schema := gozod.FromStruct[DateRange]().Refine(func(r DateRange) bool {
    return r.Start < r.End
}, "Start must be before End")
```

## Custom Error Messages via Error Tag

```go
type User struct {
    Name  string `gozod:"required,min=2" error:"Name must be at least 2 characters"`
    Email string `gozod:"required,email" error:"Please provide a valid email"`
}
```

## Code Generation

For high-throughput validation, generate zero-reflection code:

```go
//go:generate gozodgen

type User struct {
    Name  string `gozod:"required,min=2"`
    Email string `gozod:"required,email"`
    Age   int    `gozod:"required,min=18"`
}
```

Run `go generate ./...` to produce `*_gen.go` files. `FromStruct[T]()` automatically uses generated code when available (5-10x faster).

## Performance Tips

1. **Build schemas at init time**, not per-request:
   ```go
   var userSchema = gozod.FromStruct[User]() // package-level

   func handler(w http.ResponseWriter, r *http.Request) {
       result, err := userSchema.Parse(user)
   }
   ```

2. **Use `StrictParse`** when the input type is already correct:
   ```go
   result, err := userSchema.StrictParse(user) // skips runtime type check
   ```

3. **Use `gozodgen`** for zero-reflection validation in hot paths.
