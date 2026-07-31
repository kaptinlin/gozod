# JSON Schema

GoZod provides built-in support for converting schemas to [JSON Schema](https://json-schema.org/) for validation libraries, API contracts, and structured data exchange.

GoZod emits Draft 2020-12-compatible schemas for the GoZod constructs it supports. JSON Schema import is intentionally fail-closed: `FromJSONSchema` rejects unsupported semantics, while `FromJSONSchemaLossy` returns every intentionally omitted keyword with its JSON Pointer and cause.

To convert a GoZod schema to JSON Schema, use the `gozod.ToJSONSchema()` function:

```go
import (
    "github.com/kaptinlin/gozod"
    "github.com/kaptinlin/jsonschema"
)

schema := gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(),
    "age":  gozod.Int(),
})

jsonSchema, _ := gozod.ToJSONSchema(schema)
// Returns a *jsonschema.Schema instance ready for validation.
```

Supported GoZod schemas and validation methods are converted to the closest JSON Schema equivalent. Some Go types and runtime behaviors have no honest JSON Schema analog and cannot be reasonably represented. See the [`unrepresentable`](#unrepresentable-types) section below for those cases.

## Direct Integration with kaptinlin/jsonschema

GoZod's `ToJSONSchema()` function returns a `*jsonschema.Schema` instance from the [`kaptinlin/jsonschema`](https://github.com/kaptinlin/jsonschema) library, which provides robust JSON Schema validation compliant with Draft 2020-12. No additional compilation step is needed:

```go
import (
    "encoding/json"
    "fmt"
    
    "github.com/kaptinlin/gozod"
    "github.com/kaptinlin/jsonschema"
)

func main() {
    // Define GoZod schema
    userSchema := gozod.Object(gozod.ObjectSchema{
        "username": gozod.String().Min(3),
        "email":    gozod.Email(),
        "age":      gozod.Int().Min(18),
    })
    
    // Convert to JSON Schema - returns *jsonschema.Schema directly
    validator, err := gozod.ToJSONSchema(userSchema)
    if err != nil {
        panic(err)
    }
    
    // Validate data directly
    userData := map[string]any{
        "username": "john",
        "email":    "john@example.com",
        "age":      25,
    }
    
    result := validator.Validate(userData)
    if !result.IsValid() {
        details, _ := json.MarshalIndent(result.ToList(), "", "  ")
        fmt.Println("Validation failed:", string(details))
    } else {
        fmt.Println("Validation passed!")
    }
}
```

## String Formats

GoZod exports regex-backed string checks with their authoritative `pattern`.
It does not also emit `format`, because a format implementation may accept a
different set and would silently narrow the GoZod rule:

```go
gozod.Email()         // => {"type": "string", "pattern": "..."}
gozod.UUID()          // => {"type": "string", "pattern": "..."}
gozod.URL()           // => {"type": "string", "pattern": "..."}
gozod.IsoDateTime()   // => {"type": "string", "pattern": "..."}
gozod.IsoDate()       // => {"type": "string", "pattern": "..."}
gozod.IsoTime()       // => {"type": "string", "pattern": "..."}
gozod.IsoDuration()   // => {"type": "string", "pattern": "..."}
gozod.IPv4()          // => {"type": "string", "pattern": "..."}
gozod.IPv6()          // => {"type": "string", "pattern": "..."}
```

Checks without an authoritative regex can use `format`:

```go
gozod.JWT()           // => {"type": "string", "format": "jwt"}
gozod.Time()          // => {"type": "string", "format": "time"}
```

Encoded strings combine `contentEncoding` with their pattern:

```go
gozod.Base64()        // => {"type": "string", "contentEncoding": "base64", "pattern": "..."}
gozod.Base64URL()     // => {"type": "string", "contentEncoding": "base64url", "pattern": "..."}
```

Other regex-backed checks follow the same rule:

```go
gozod.String().Regex(regexp.MustCompile("^[a-z]+$"))
// => {"type": "string", "pattern": "^[a-z]+$"}

gozod.UUIDv4()        // => {"type": "string", "pattern": "..."}
gozod.UUIDv6()        // => {"type": "string", "pattern": "..."}
gozod.UUIDv7()        // => {"type": "string", "pattern": "..."}
gozod.CIDRv4()        // => {"type": "string", "pattern": "..."}
gozod.CIDRv6()        // => {"type": "string", "pattern": "..."}
gozod.CUID()          // => {"type": "string", "pattern": "..."}
gozod.CUID2()         // => {"type": "string", "pattern": "..."}
gozod.ULID()          // => {"type": "string", "pattern": "..."}
gozod.KSUID()         // => {"type": "string", "pattern": "..."}
gozod.NanoID()        // => {"type": "string", "pattern": "..."}
```

## File Types

GoZod supports file validation that translates to JSON Schema's `binary` format and `contentMediaType` for MIME types:

```go
// A generic file
gozod.File()
// => {"type": "string", "format": "binary", "contentEncoding": "binary"}

// A file with MIME type and size constraints
gozod.File().Mime([]string{"image/png"}).Min(1000).Max(10000)
// => {"type":"string","format":"binary","contentEncoding":"binary","contentMediaType":"image/png","minLength":1000,"maxLength":10000}

// A file with multiple possible MIME types
gozod.File().Mime([]string{"image/png", "image/jpeg"})
// => {"anyOf": [{"contentMediaType":"image/png", ...}, {"contentMediaType":"image/jpeg", ...}]}
```

## Numeric Types

GoZod converts the following numeric types to JSON Schema:

```go
// number
gozod.Float64()  // => {"type": "number"}
gozod.Number()   // => {"type": "number"}
gozod.Float32()  // => {"type": "number", "minimum": ..., "maximum": ...}

// integer
gozod.Int()      // => integer with platform int minimum/maximum
gozod.Int32()    // => {"type": "integer", "minimum": ..., "maximum": ...}
gozod.Int64()    // => {"type": "integer", "minimum": ..., "maximum": ...}
```

## Tuple Types

GoZod tuples convert to JSON Schema arrays with `prefixItems`:

```go
schema := gozod.Tuple(gozod.String(), gozod.Int(), gozod.Bool())
jsonSchema, _ := gozod.ToJSONSchema(schema)
```

Result:
```json
{
  "type": "array",
  "prefixItems": [
    {"type": "string"},
    {"type": "integer"},
    {"type": "boolean"}
  ],
  "minItems": 3,
  "maxItems": 3
}
```

### Tuple with Rest

Tuples with rest elements use `items` for the rest schema:

```go
schema := gozod.TupleWithRest(
    []core.ZodSchema{gozod.String(), gozod.Int()},
    gozod.Bool(),
)
jsonSchema, _ := gozod.ToJSONSchema(schema)
```

Result:
```json
{
  "type": "array",
  "prefixItems": [
    {"type": "string"},
    {"type": "integer"}
  ],
  "items": {"type": "boolean"}
}
```

## Xor (Exclusive Union)

Xor schemas convert to JSON Schema `oneOf` (exactly one must match):

```go
schema := gozod.Xor([]any{
    gozod.String().Email(),
    gozod.String().URL(),
})
jsonSchema, _ := gozod.ToJSONSchema(schema)
```

Result:
```json
{
  "oneOf": [
    {"type": "string", "format": "email"},
    {"type": "string", "format": "uri"}
  ]
}
```

> **Note:** Both Xor and DiscriminatedUnion generate `oneOf` in JSON Schema.

## Nullability

GoZod distinguishes between optional and nullable fields, which affects how they are represented in JSON Schema.

- `Optional()`: Marks a field as not required in an object. When used on a standalone schema, it has no effect on the output.
- `Nilable()`: Allows a value to be `null`. This is represented using `anyOf` in JSON Schema.

```go
// Optional fields are handled by their absence from the `required` array in an object schema.
gozod.String().Optional()
// As a standalone schema => {"type": "string"}

// Nilable schemas can be their base type or null.
gozod.String().Nilable()
// => {"anyOf": [{"type": "string"}, {"type": "null"}]}

gozod.Any()       // => {} (represents any value)
gozod.Nil()       // => {"type": "null"}
```

## Configuration

A second argument can be used to customize the conversion logic:

```go
gozod.ToJSONSchema(schema, gozod.JSONSchemaOptions{
    // Configuration options
})
```

Below is a quick reference for each supported parameter:

```go
type JSONSchemaOptions struct {
    // Whole-record metadata overrides for schemas present in this registry.
    // Missing entries fall back to schema-owned metadata.
    Metadata *gozod.Registry[gozod.GlobalMeta]
    
    // How to handle unrepresentable types.
    // JSONSchemaUnrepresentableThrow is the default.
    Unrepresentable gozod.JSONSchemaUnrepresentableMode
    
    // How to handle cycles.
    // JSONSchemaCyclesRef is the default.
    Cycles gozod.JSONSchemaCyclesMode
    
    // How to handle reused schemas.
    // JSONSchemaReusedInline is the default.
    Reused gozod.JSONSchemaReusedMode
    
    // A function used to convert ID values to URIs for external $refs
    URI func(id string) string

    // IO specifies whether to convert the "input" or "output" schema.
    // JSONSchemaIOOutput is the default.
    IO gozod.JSONSchemaIOMode

    // Override is a custom logic to modify the schema after generation.
    Override func(ctx gozod.OverrideContext)
}
```

### Metadata

GoZod includes schema-owned metadata in generated JSON Schema by default:

```go
// Add metadata to a schema
emailSchema := gozod.String().Meta(gozod.GlobalMeta{
    Title:       "Email Address",
    Description: "User's email address",
    Examples:    []any{"user@example.com"},
})

jsonSchema, _ := gozod.ToJSONSchema(emailSchema)
// The returned JSON schema string will include all metadata
```

An explicit registry entry replaces the complete schema-owned metadata record
for that conversion. A missing entry falls back to the schema value; fields are
not implicitly merged across the two records.

### Unrepresentable Types

Some GoZod features and validation patterns cannot be directly represented in JSON Schema. By default, GoZod will throw an error if these are encountered:

```go
// ❌ Transform functions cannot be represented in JSON Schema
gozod.String().Transform(func(s string) string {
    return strings.ToUpper(s)
})

// ❌ Custom refinements cannot be represented in JSON Schema  
gozod.String().Refine(func(val any) bool {
    if s, ok := val.(string); ok {
        return len(s) > 5
    }
    return false
}, gozod.CustomParams{
    Error: "String must be longer than 5 characters",
})

// ❌ Function, BigInt, and Complex schemas cannot be represented in JSON Schema
gozod.Function()
gozod.BigInt()
gozod.Complex()
```

> **Note on Discriminated Unions**: While GoZod supports discriminated unions, the `discriminator` keyword is not added to the generated JSON Schema. This is because the standard `jsonschema.Schema` struct does not include this field. Validation still works correctly using the `oneOf` keyword.

By default, GoZod will return an error if any of these are encountered:

```go
schema := gozod.String().Transform(func(s string) string {
    return strings.ToUpper(s)
})

jsonSchema, _ := gozod.ToJSONSchema(schema, gozod.JSONSchemaOptions{
    Unrepresentable: gozod.JSONSchemaUnrepresentableAny,
})
// => returns a *jsonschema.Schema representing {} (accepts any value)
```

### Cycles

How to handle cycles. If a cycle is encountered as `gozod.ToJSONSchema()` traverses the schema, it will be represented using `$ref`:

```go
// Define a recursive user schema
var UserSchema gozod.ZodSchema
UserSchema = gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(),
    "friend": gozod.Lazy(func() gozod.ZodSchema {
        return UserSchema
    }),
})

jsonSchema, _ := gozod.ToJSONSchema(UserSchema)
// Returns *jsonschema.Schema with proper $ref handling for cycles
```

If instead you want to throw an error, set the `Cycles` option to
`gozod.JSONSchemaCyclesThrow`:

```go
_, err := gozod.ToJSONSchema(UserSchema, gozod.JSONSchemaOptions{
    Cycles: gozod.JSONSchemaCyclesThrow,
})
// => returns error if cycles are detected
```

### Reused Schemas

How to handle schemas that occur multiple times in the same schema. By default, GoZod will inline these schemas:

```go
nameSchema := gozod.String()
userSchema := gozod.Object(gozod.ObjectSchema{
    "firstName": nameSchema,
    "lastName":  nameSchema,
})

jsonSchema, _ := gozod.ToJSONSchema(userSchema)
// Both firstName and lastName will have inlined string schemas
```

Instead you can set the `Reused` option to `gozod.JSONSchemaReusedRef` to
extract these schemas into `$defs`:

```go
jsonSchema, _ := gozod.ToJSONSchema(userSchema, gozod.JSONSchemaOptions{
    Reused: gozod.JSONSchemaReusedRef,
})
// Common schemas will be extracted to $defs and referenced
```

### Input/Output Schemas (`IO`)

The `IO` option controls how schemas with default values are handled. In `"input"` mode, fields with defaults are optional. In `"output"` mode (the default), they are required.

```go
schema := gozod.Object(gozod.ObjectSchema{
    "a": gozod.String(),
    "b": gozod.String().Optional(),
    "c": gozod.String().Default("hello"),
})

// In "input" mode, 'c' is optional because it has a default.
// "required" will be ["a"]
gozod.ToJSONSchema(schema, gozod.JSONSchemaOptions{IO: gozod.JSONSchemaIOInput})

// In "output" mode, 'c' is required.
// "required" will be ["a", "c"]
gozod.ToJSONSchema(schema, gozod.JSONSchemaOptions{IO: gozod.JSONSchemaIOOutput})
```

Export always emits the supported Draft 2020-12 dialect. There is no target
option until the converter implements another dialect faithfully.

### Override

The `Override` option allows you to programmatically modify the generated JSON schema. This is useful for adding custom keywords or making adjustments that GoZod doesn't support natively.

```go
schema := gozod.String()
opts := gozod.JSONSchemaOptions{
    Override: func(ctx gozod.OverrideContext) {
        // Add a title to all string schemas
        if _, ok := ctx.ZodSchema.(*types.ZodString[string]); ok {
            title := "Overridden Title"
            ctx.JSONSchema.Title = &title
        }
    },
}
jsonSchema, _ := gozod.ToJSONSchema(schema, opts)
// The resulting schema will contain: "title": "Overridden Title"
```

## Working with Registries

For complex applications with multiple schemas, use `ToJSONSchemaRegistry` to generate a single root schema containing every registered schema in `$defs`.

```go
// Create a registry for related schemas
registry := gozod.NewRegistry[gozod.GlobalMeta]()

// Define schemas, then register the metadata that names their `$defs` entries.
var userSchema, postSchema gozod.ZodSchema

userSchema = gozod.Object(gozod.ObjectSchema{
    "id":   gozod.UUID(),
    "name": gozod.String(),
    "posts": gozod.Lazy(func() gozod.ZodSchema {
        return gozod.Slice[any](postSchema)
    }),
})

postSchema = gozod.Object(gozod.ObjectSchema{
    "id":      gozod.UUID(),
    "title":   gozod.String(),
    "author":  gozod.Lazy(func() gozod.ZodSchema { return userSchema }),
})

registry.Add(userSchema, gozod.GlobalMeta{ID: "User"})
registry.Add(postSchema, gozod.GlobalMeta{ID: "Post"})

// Convert the entire registry to a single root JSON Schema.
// Schemas with IDs will be defined in `$defs` and can be referenced.
rootSchema, _ := gozod.ToJSONSchemaRegistry(registry)
```

## Converting JSON Schema to GoZod

GoZod also supports converting JSON Schema to GoZod schemas using `FromJSONSchema()`:

```go
import (
    "github.com/kaptinlin/gozod"
    lib "github.com/kaptinlin/jsonschema"
)

// Create a JSON Schema
jsonSchema := &lib.Schema{}
jsonSchema.Type = []string{"object"}
props := lib.SchemaMap{
    "name": &lib.Schema{Type: []string{"string"}},
    "age":  &lib.Schema{Type: []string{"integer"}},
}
jsonSchema.Properties = &props
jsonSchema.Required = []string{"name"}

// Convert to GoZod schema
zodSchema, err := gozod.FromJSONSchema(jsonSchema)
if err != nil {
    panic(err)
}

// Use the schema for validation
result, err := zodSchema.ParseAny(map[string]any{
    "name": "John",
    "age":  30,
})
```

### Tuple Conversion (prefixItems)

JSON Schema Draft 2020-12 `prefixItems` are automatically converted to GoZod Tuple schemas:

```go
// JSON Schema with prefixItems
schema := &lib.Schema{}
schema.Type = []string{"array"}
schema.PrefixItems = []*lib.Schema{
    {Type: []string{"string"}},
    {Type: []string{"integer"}},
}

zodSchema, _ := gozod.FromJSONSchema(schema)
// Equivalent to: gozod.Tuple(gozod.String(), gozod.Int())

result, _ := zodSchema.ParseAny([]any{"hello", 42})
```

With rest elements:

```go
// JSON Schema with prefixItems and items (rest)
schema := &lib.Schema{}
schema.Type = []string{"array"}
schema.PrefixItems = []*lib.Schema{
    {Type: []string{"string"}},
}
schema.Items = &lib.Schema{Type: []string{"boolean"}}

zodSchema, _ := gozod.FromJSONSchema(schema)
// Equivalent to: gozod.TupleWithRest([]core.ZodSchema{gozod.String()}, gozod.Bool())

result, _ := zodSchema.ParseAny([]any{"hello", true, false})
```

### Unsupported Keywords

`FromJSONSchema` fails by default when a JSON Schema keyword cannot be represented
without losing validation semantics:

```go
zodSchema, err := gozod.FromJSONSchema(schema)
if err != nil {
    // Handle unsupported features
}
```

Use `FromJSONSchemaLossy` only when partial import is intentional. It returns a
caller-owned, location-aware loss snapshot alongside the imported schema:

```go
zodSchema, losses, err := gozod.FromJSONSchemaLossy(schema)
if err != nil {
    // Fatal conversion errors still stop the import.
}
for _, loss := range losses {
    fmt.Printf("%s at %s: %v\n", loss.Keyword, loss.Pointer, loss.Err)
}
```

Imported metadata (`$id`, `title`, `description`, and `examples`) belongs to the
returned schema by default. Pass a registry to make it the sole metadata
destination; import never writes `gozod.GlobalRegistry` implicitly:

```go
metadata := gozod.NewRegistry[gozod.GlobalMeta]()
zodSchema, err := gozod.FromJSONSchema(schema, gozod.FromJSONSchemaOptions{
    Metadata: metadata,
})
```

Both destinations receive a snapshot. Nested examples no longer alias the
input JSON Schema document.

Import does not claim the whole JSON Schema language. The current fail-closed
keywords are:

Type-specific validation keywords such as `minLength`, `minimum`, `items`, and
`properties` also fail closed when no `type` or resolved reference anchors
their instance family. Inferring a narrow type would reject values that JSON
Schema intentionally leaves unaffected, while importing `{}` would drop the
validation.

| Keyword | Reason |
|---------|--------|
| `if` / `then` / `else` | Conditional validation has no direct GoZod schema shape. |
| `patternProperties` | Only the pure one-pattern shape exported by a regex `LooseRecord` is imported; combined or multi-pattern shapes fail closed. |
| `$dynamicRef` | Dynamic reference resolution is outside GoZod's schema graph. |
| `unevaluatedProperties` | Evaluation bookkeeping is a JSON Schema runtime concept. |
| `unevaluatedItems` | Evaluation bookkeeping is a JSON Schema runtime concept. |
| `not` | Negated schemas are not imported as a general validation form. |
| `uniqueItems: true` | Slice uniqueness is not a first-class GoZod array contract. |
| `dependentSchemas` | Cross-field schema dependency evaluation is not imported. |
| `dependentRequired` | Conditional cross-field requirements are not imported without a path-preserving typed check. |
| `contentEncoding` | Encoded-content validation is not imported because existing string checks do not match the dependency decoder exactly. |
| `contentMediaType` | Media handlers other than raw `application/json` have no equivalent GoZod string check. |
| `contentSchema` | Validation after content decoding and unmarshaling has no GoZod schema boundary. |
| `propertyNames` | Only pure non-exhaustive string-key records with `additionalProperties` are imported; discrete or combined object shapes fail closed. |
| `contains` / `minContains` / `maxContains` | Positional search constraints are not imported. |

Round-trip expectations apply only to the overlap GoZod owns: primitive types,
known string formats, numeric and length constraints, arrays, tuples, objects,
records, enums, literals, and `allOf` / `anyOf` / `oneOf` composition.

Imported `integer` schemas use Go's platform-sized `int` domain. Rational
bounds are rounded to the equivalent inclusive integer bound. If a bound,
exclusive-bound adjustment, or `multipleOf` divisor cannot be represented in
that domain, both strict and lossy import return a fatal
`ErrInvalidJSONSchema` with the keyword's JSON Pointer.

### Supported Conversions

| JSON Schema | GoZod Schema |
|-------------|--------------|
| `type: "string"` | `gozod.String()` |
| `type: "number"` | `gozod.Number()` |
| `type: "integer"` | `gozod.Int()` |
| `type: "boolean"` | `gozod.Bool()` |
| `type: "null"` | `gozod.Nil()` |
| `type: "array"` | `gozod.Slice()` |
| `type: "object"` | `gozod.Object()` |
| `propertyNames` + `additionalProperties` pure record | `gozod.Record()` |
| one `patternProperties` pure record | `gozod.LooseRecord()` |
| `prefixItems` | `gozod.Tuple()` |
| `anyOf` | `gozod.Union()` |
| `oneOf` | `gozod.Xor()` |
| `allOf` | `gozod.Intersection()` |
| `const` | `gozod.Literal()` |
| `enum` | `gozod.Enum()` |
| `format: "email"` | `gozod.Email()` |
| `format: "uuid"` | `gozod.UUID()` |
| `format: "uri"` | `gozod.URL()` |
| `format: "date-time"` | `gozod.IsoDateTime()` |
| `format: "date"` | `gozod.IsoDate()` |
| `format: "time"` | `gozod.IsoTime()` |
| `format: "ipv4"` | `gozod.IPv4()` |
| `format: "ipv6"` | `gozod.IPv6()` |
| `type: "string"`, `contentMediaType: "application/json"` | `gozod.String().JSON()` |
