---
description: Validate data in Go applications using github.com/kaptinlin/gozod, a Zod v4-inspired validation library with strict type semantics, struct tag support, and zero dependencies. Use when defining validation schemas, validating API requests, parsing configuration, or handling validation errors.
name: gozod-validating
---


# Go Data Validation with GoZod

Validate data using `github.com/kaptinlin/gozod` -- a TypeScript Zod v4-inspired library with strict type semantics, method chaining, struct tags, and zero dependencies.

## Decision Flowchart

```
What are you validating?
|
+- Go struct with field rules
|  +- Declarative (struct tags)
|  |  -> gozod.FromStruct[T]()  with  gozod:"required,min=2,email"
|  +- Programmatic (field schemas)
|     -> gozod.Struct[T](gozod.StructSchema{...})
|
+- Dynamic JSON / map[string]any
|  -> gozod.Object(gozod.ObjectSchema{...})
|
+- Single value (string, int, etc.)
|  -> gozod.String().Min(3).Email()
|  -> gozod.Int().Min(0).Max(120)
|
+- Union / discriminated types
|  -> gozod.Union(schemaA, schemaB)
|  -> gozod.DiscriminatedUnion("type", schemas)
|
+- Recursive / self-referencing types
   -> gozod.Lazy(func() S { ... })
   -> gozod.FromStruct[T]()  (auto-detects circular refs)
```

## Architecture Overview

```
gozod (root API)
|
+- types/          Schema implementations (one per file)
|  +- Primitives:  string, bool, integer, float, bigint, complex, time
|  +- Collections: array, slice, tuple, map, record, set
|  +- Objects:     object (map[string]any), struct (typed Go struct)
|  +- Composition: union, xor, intersection, discriminated_union
|  +- Formats:     email, url, uuid, ipv4, ipv6, jwt, hostname, ...
|  +- Special:     any, unknown, never, nil, lazy, literal, enum
|
+- core/           Interfaces (ZodType[T], ZodSchema), checks, config
+- internal/       Engine, issues, checks, utilities
+- locales/        i18n error message bundles (en, zh-CN, ...)
+- jsonschema/     Bidirectional JSON Schema conversion
+- coerce/         Type coercion utilities
+- cmd/gozodgen/   Code generation for zero-reflection validation
```

## Quick Start

```bash
go get github.com/kaptinlin/gozod
```

### Struct Tag Validation (Declarative)

```go
type CreateUserRequest struct {
    Name  string `json:"name"  gozod:"required,min=2,max=50"`
    Email string `json:"email" gozod:"required,email"`
    Age   int    `json:"age"   gozod:"required,min=18,max=120"`
}

schema := gozod.FromStruct[CreateUserRequest]()
user, err := schema.Parse(req) // or schema.StrictParse(req) for compile-time safety
```

### Programmatic Schema

```go
schema := gozod.Object(gozod.ObjectSchema{
    "name":  gozod.String().Min(2).Max(50),
    "email": gozod.String().Email(),
    "age":   gozod.Int().Min(18).Max(120),
    "tags":  gozod.Slice[string](gozod.String()).Max(10),
})
result, err := schema.Parse(data) // data is map[string]any
```

## Parse vs StrictParse

| Method | Input | Safety | Performance | Use When |
|--------|-------|--------|-------------|----------|
| `Parse(any)` | Runtime-checked `any` | Runtime error on type mismatch | Standard | Input from JSON decode, external sources |
| `StrictParse(T)` | Compile-time typed `T` | Compile error on wrong type | Optimal | Input already the correct Go type |
| `MustParse(any)` | Runtime-checked, panics | Panics on error | Standard | Init-time validation, tests |
| `MustStrictParse(T)` | Compile-time typed, panics | Panics on error | Optimal | Init-time validation, tests |

## Schema Type Quick Reference

| Category | Constructors | Key Methods |
|----------|-------------|-------------|
| **String** | `String()`, `StringPtr()` | `.Min()`, `.Max()`, `.Length()`, `.Email()`, `.URL()`, `.UUID()`, `.Regex()`, `.Trim()`, `.Lowercase()`, `.Uppercase()` |
| **Number** | `Int()`, `Int64()`, `Float64()`, `Uint()`, `Number()`, `BigInt()` + all Go int/uint/float variants | `.Min()`, `.Max()`, `.Positive()`, `.Negative()`, `.NonNegative()`, `.Finite()`, `.MultipleOf()` |
| **Bool** | `Bool()`, `BoolPtr()` | Standard modifiers |
| **Time** | `Time()`, `TimePtr()` | `.After()`, `.Before()` |
| **Array/Slice** | `Array(elem)`, `Slice[T](elem)` | `.Min()`, `.Max()`, `.Length()`, `.NonEmpty()` |
| **Object** | `Object(schema)`, `StrictObject(schema)`, `LooseObject(schema)` | `.Extend()`, `.Pick()`, `.Omit()`, `.Partial()`, `.Required()` |
| **Struct** | `Struct[T]()`, `FromStruct[T]()` | `.Refine()`, `.StrictParse()` |
| **Map** | `Map(valueSchema)` | `.Min()`, `.Max()`, `.NonEmpty()` |
| **Record** | `Record(keySchema, valSchema)`, `LooseRecord(...)` | `.Partial()` |
| **Set** | `Set[T](elemSchema)` | `.Min()`, `.Max()` |
| **Tuple** | `Tuple(schemas...)`, `TupleWithRest(items, rest)` | Positional validation |
| **Union** | `Union(schemas...)`, `Xor(schemas)`, `DiscriminatedUnion(key, schemas)` | `.And()`, `.Or()` |
| **Intersection** | `Intersection(schemas...)` | All schemas must match |
| **Enum** | `Enum(values...)`, `EnumSlice(vals)`, `EnumMap(entries)` | Exact value matching |
| **Literal** | `Literal(value)`, `LiteralOf(values)` | Exact value matching |
| **Lazy** | `Lazy(func() S)` | Recursive/circular schema |
| **Format** | `Email()`, `URL()`, `HTTPURL()`, `UUID()`, `IPv4()`, `IPv6()`, `Hostname()`, `MAC()`, `E164()`, `JWT()`, `Hex()`, `Base64()`, `CIDRv4()`, `IsoDateTime()`, `IsoDuration()` | Format-specific methods |
| **Special** | `Any()`, `Unknown()`, `Never()`, `Nil()`, `Function()`, `StringBool()` | Type-specific |

Every constructor has a `*Ptr()` variant (e.g., `StringPtr()`, `IntPtr()`) returning pointer output type.

## Modifiers (Copy-on-Write)

All modifiers clone the schema and return a new instance:

```go
schema := gozod.String().Min(3)
optional := schema.Optional()   // *string output, accepts nil/missing
nilable  := schema.Nilable()    // *string output, accepts nil
required := optional.NonOptional() // undo Optional
```

| Modifier | Output Type | Behavior |
|----------|------------|----------|
| `.Optional()` | `*T` | Accepts missing/nil values |
| `.Nilable()` | `*T` | Accepts nil values |
| `.Nullish()` | `*T` | Optional + Nilable combined |
| `.NonOptional()` | `T` | Reverts Optional |
| `.Default(val)` | `T` | Short-circuit: returns default on nil (skips validation) |
| `.DefaultFunc(fn)` | `T` | Dynamic default via function |
| `.Prefault(val)` | `T` | Pre-parse: feeds default through full validation pipeline |
| `.PrefaultFunc(fn)` | `T` | Dynamic prefault via function |
| `.Coerce()` | `T` | Enable type coercion (string->int, etc.) |
| `.Transform(fn)` | Changes output | Transform validated value |
| `.Pipe(schema)` | Changes output | Chain into another schema |
| `.Refine(fn, msg)` | `T` | Custom validation logic |
| `.Describe(desc)` | `T` | Attach description metadata |
| `.Meta(meta)` | `T` | Attach structured metadata |

## Struct Tags -- [details](references/struct-tags.md)

Tag syntax: `gozod:"rule1,rule2=value,rule3"`

| Tag | Applies To | Description |
|-----|-----------|-------------|
| `required` | All | Field must be present and non-zero |
| `-` | All | Skip validation entirely |
| `min=N` | String (length), Number (value), Array (count) | Minimum constraint |
| `max=N` | String (length), Number (value), Array (count) | Maximum constraint |
| `length=N` | String, Array | Exact length/count |
| `email` | String | Email format |
| `url` | String | URL format |
| `uuid` | String | UUID format |
| `regex=pattern` | String | Custom regex |
| `positive` | Number | > 0 |
| `negative` | Number | < 0 |
| `nonnegative` | Number | >= 0 |
| `nonpositive` | Number | <= 0 |
| `nonempty` | Array/Slice | At least one element |

## Error Handling -- [details](references/error-handling.md)

```go
_, err := schema.Parse(input)
if err != nil {
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // Iterate structured issues
        for _, issue := range zodErr.Issues {
            fmt.Printf("Path: %v, Code: %s, Message: %s\n",
                issue.Path, issue.Code, issue.Message)
        }

        // Format for API responses (flat field errors)
        flat := gozod.FlattenError(zodErr.Issues)
        // flat.FormErrors   []string            -- root-level errors
        // flat.FieldErrors  map[string][]string  -- per-field errors

        // Format for logging (human-readable)
        fmt.Println(gozod.PrettifyError(zodErr.Issues))

        // Format for nested UIs (tree structure)
        tree := gozod.TreeifyError(zodErr.Issues)
    }
}
```

### Custom Error Messages

```go
// Inline string
gozod.String().Email("Please enter a valid email")
gozod.Int().Min(18, "Must be at least 18 years old")

// Dynamic function
gozod.String().Min(8, func(issue gozod.ZodRawIssue) string {
    return fmt.Sprintf("Password needs %v+ chars", issue.Minimum())
})

// Schema-level
gozod.String(gozod.SchemaParams{Error: "Invalid input"})
```

### Error Codes

| Code | Meaning | Context Properties |
|------|---------|-------------------|
| `invalid_type` | Wrong Go type | `Expected()`, `Received()` |
| `too_small` | Below minimum | `Minimum()`, `Inclusive()` |
| `too_big` | Above maximum | `Maximum()`, `Inclusive()` |
| `invalid_format` | Format mismatch | `Format()` |
| `invalid_value` | Not in allowed set | `Values()` |
| `invalid_union` | No union member matched | -- |
| `unrecognized_keys` | Unknown object keys | `Keys()` |
| `not_multiple_of` | Not divisible | `Divisor()` |
| `custom` | Custom Refine failed | -- |

### i18n

```go
import "github.com/kaptinlin/gozod/locales"

gozod.Config(locales.ZhCN()) // all errors now in Chinese

// Per-format locale
formatter := locales.GetLocaleFormatter("zh-CN")
pretty := gozod.PrettifyErrorWithMapper(zodErr, formatter)
```

## Custom Validation with Refine

```go
// Single-value custom check
usernameSchema := gozod.String().Min(3).Max(20).Refine(func(s string) bool {
    return !reservedNames[strings.ToLower(s)]
}, "Username is reserved")

// Cross-field validation on structs
type PasswordForm struct {
    Password string `gozod:"required,min=8"`
    Confirm  string `gozod:"required"`
}
schema := gozod.FromStruct[PasswordForm]().Refine(func(f PasswordForm) bool {
    return f.Password == f.Confirm
}, "Passwords must match")
```

## Schema Composition

```go
// Union: any member matches
gozod.Union(gozod.String(), gozod.Int())

// Xor: exactly one must match
gozod.Xor([]any{gozod.Email(), gozod.URL()})

// Intersection: all must match
gozod.Intersection(gozod.String().Min(3), gozod.String().Max(10))

// Fluent composition
gozod.Int().Or(gozod.String())          // union via method
gozod.String().Min(3).And(gozod.String().Max(10)) // intersection via method

// Discriminated union (optimized key-based dispatch)
gozod.DiscriminatedUnion("type", []any{
    gozod.Object(gozod.ObjectSchema{"type": gozod.Literal("email"), "address": gozod.String().Email()}),
    gozod.Object(gozod.ObjectSchema{"type": gozod.Literal("phone"), "number": gozod.String().E164()}),
})
```

## Transform and Pipe

```go
// Transform: validated value -> new value
stringToInt := gozod.String().Regex(`^\d+$`).Transform(
    func(s string, ctx *core.RefinementContext) (any, error) {
        return strconv.Atoi(s)
    },
)

// Pipe: chain into another schema for re-validation
schema := gozod.String().Pipe(gozod.Int()) // parse string, then validate as int

// Apply: reusable schema modifiers
func trimmed[T types.StringConstraint](s *gozod.ZodString[T]) *gozod.ZodString[T] {
    return s.Trim().Min(1)
}
schema := gozod.Apply(gozod.String(), trimmed)
```

## Code Generation (Performance)

```go
//go:generate gozodgen

type User struct {
    Name  string `gozod:"required,min=2"`
    Email string `gozod:"required,email"`
}

// After running gozodgen, FromStruct uses generated code (5-10x faster, zero reflection)
schema := gozod.FromStruct[User]()
```

## JSON Schema Integration

```go
// GoZod -> JSON Schema (Draft 2020-12, OpenAPI 3.1 compatible)
jsonSchema, err := gozod.ToJSONSchema(schema)

// JSON Schema -> GoZod
zodSchema, err := gozod.FromJSONSchema(jsonSchemaInstance)
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| `Parse(any)` when type is known at compile time | Unnecessary runtime type check overhead | Use `StrictParse(T)` for typed inputs |
| Reusing schema variable after modifier | Modifiers are copy-on-write; original is unchanged | Assign modifier result: `s = s.Optional()` |
| Building `FromStruct` schema on every request | Reflection cost on each call | Build schema once at init, reuse |
| Using `Object()` for typed Go structs | Loses type safety, manual field mapping | Use `Struct[T]()` or `FromStruct[T]()` |
| Ignoring `*Ptr()` constructors for optional fields | Type mismatch when field is `*string` | Use `StringPtr()` for pointer types |
| Using `Default()` expecting validation to run | Default short-circuits; skips validation | Use `Prefault()` to run value through pipeline |
| Logging `zodErr.Error()` only | Loses structured issue details | Use `zodErr.Issues` for programmatic access |
| `WithNoEncryption` coercion in production | Silent data mutation | Only use `Coerce()` when explicitly needed |
