# JSON Schema

GoZod provides built-in support for converting schemas to [JSON Schema](https://json-schema.org/), making it easy to integrate with validation libraries and define structured data for APIs and AI applications.

> **New Feature** — GoZod introduces native [JSON Schema](https://json-schema.org/) conversion that integrates seamlessly with the [`kaptinlin/jsonschema`](https://github.com/kaptinlin/jsonschema) validator library. JSON Schema is widely used in [OpenAPI](https://www.openapis.org/) definitions and for defining [structured outputs](https://platform.openai.com/docs/guides/structured-outputs) for AI.

> **Latest JSON Schema Support** — GoZod is compliant with JSON Schema Draft 2020-12. This library does not support earlier versions of JSON Schema, ensuring you always work with the most modern and feature-complete specification.

To convert a GoZod schema to JSON Schema, use the `gozod.ToJSONSchema()` function:

```go
import (
    "github.com/kaptinlin/gozod"
    "github.com/kaptinlin/jsonschema"
)

schema := gozod.Struct(gozod.ObjectSchema{
    "name": gozod.String(),
    "age":  gozod.Int(),
})

jsonSchema, _ := gozod.ToJSONSchema(schema)
// Returns a *jsonschema.Schema instance ready for validation.
```

All GoZod schemas and validation methods are converted to their closest JSON Schema equivalent. Some Go types have no analog and cannot be reasonably represented. See the [`unrepresentable`](#unrepresentable-types) section below for more information on handling these cases.

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
    userSchema := gozod.Struct(gozod.ObjectSchema{
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

GoZod converts the following schema types to the equivalent JSON Schema `format`:

```go
// Supported via `format`
gozod.Email()         // => {"type": "string", "format": "email"}
gozod.Uuid()          // => {"type": "string", "format": "uuid"}
gozod.URL()           // => {"type": "string", "format": "uri"}
gozod.JWT()           // => {"type": "string", "format": "jwt"}
gozod.IsoDateTime()   // => {"type": "string", "format": "date-time"}
gozod.IsoDate()       // => {"type": "string", "format": "date"}
gozod.IsoTime()       // => {"type": "string", "format": "time"}
gozod.IsoDuration()   // => {"type": "string", "format": "duration"}
gozod.IPv4()          // => {"type": "string", "format": "ipv4"}
gozod.IPv6()          // => {"type": "string", "format": "ipv6"}
```

These schemas are supported via `contentEncoding`:

```go
gozod.Base64()        // => {"type": "string", "contentEncoding": "base64"}
gozod.Base64URL()     // => {"type": "string", "contentEncoding": "base64url", "format": "base64url"}
```

String patterns and custom formats are supported via `pattern`:

```go
gozod.String().Regex(regexp.MustCompile("^[a-z]+$"))
// => {"type": "string", "pattern": "^[a-z]+$"}

gozod.Uuidv4()        // => {"type": "string", "format": "uuid", "pattern": "..."}
gozod.Uuidv6()        // => {"type": "string", "format": "uuid", "pattern": "..."}
gozod.Uuidv7()        // => {"type": "string", "format": "uuid", "pattern": "..."}
gozod.CIDRv4()        // => {"type": "string", "format": "cidrv4", "pattern": "..."}
gozod.CIDRv6()        // => {"type": "string", "format": "cidrv6", "pattern": "..."}
gozod.Cuid()          // => {"type": "string", "format": "cuid", "pattern": "..."}
gozod.Cuid2()         // => {"type": "string", "format": "cuid2", "pattern": "..."}
gozod.Ulid()          // => {"type": "string", "format": "ulid", "pattern": "..."}
gozod.Ksuid()         // => {"type": "string", "format": "ksuid", "pattern": "..."}
gozod.Nanoid()        // => {"type": "string", "format": "nanoid", "pattern": "..."}
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
gozod.Int()      // => {"type": "integer"}
gozod.Int32()    // => {"type": "integer", "minimum": ..., "maximum": ...}
gozod.Int64()    // => {"type": "integer", "minimum": ..., "maximum": ...}
```

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
    // A registry used to look up metadata for each schema
    // Any schema with an ID property will be extracted as a $def
    Metadata *gozod.Registry[gozod.GlobalMeta]
    
    // How to handle unrepresentable types
    // "throw" (default) - Unrepresentable types throw an error
    // "any" - Unrepresentable types become {}
    Unrepresentable string
    
    // How to handle cycles
    // "ref" (default) - Cycles will be broken using $defs
    // "throw" - Cycles will throw an error if encountered
    Cycles string
    
    // How to handle reused schemas
    // "inline" (default) - Reused schemas will be inlined
    // "ref" - Reused schemas will be extracted as $defs
    Reused string
    
    // A function used to convert ID values to URIs for external $refs
    URI func(id string) string

    // IO specifies whether to convert the "input" or "output" schema.
    // "output" (default) or "input". Affects handling of default values.
    IO string

    // Override is a custom logic to modify the schema after generation.
    Override func(ctx gozod.OverrideContext)
}
```

### Metadata

GoZod supports metadata storage that will be included in the generated JSON Schema:

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
    Unrepresentable: "any",
})
// => returns a *jsonschema.Schema representing {} (accepts any value)
```

### Cycles

How to handle cycles. If a cycle is encountered as `gozod.ToJSONSchema()` traverses the schema, it will be represented using `$ref`:

```go
// Define a recursive user schema
var UserSchema gozod.ZodSchema
UserSchema = gozod.Struct(gozod.ObjectSchema{
    "name": gozod.String(),
    "friend": gozod.Lazy(func() gozod.ZodSchema {
        return UserSchema
    }),
})

jsonSchema, _ := gozod.ToJSONSchema(UserSchema)
// Returns *jsonschema.Schema with proper $ref handling for cycles
```

If instead you want to throw an error, set the `Cycles` option to `"throw"`:

```go
_, err := gozod.ToJSONSchema(UserSchema, gozod.JSONSchemaOptions{
    Cycles: "throw",
})
// => returns error if cycles are detected
```

### Reused Schemas

How to handle schemas that occur multiple times in the same schema. By default, GoZod will inline these schemas:

```go
nameSchema := gozod.String()
userSchema := gozod.Struct(gozod.ObjectSchema{
    "firstName": nameSchema,
    "lastName":  nameSchema,
})

jsonSchema, _ := gozod.ToJSONSchema(userSchema)
// Both firstName and lastName will have inlined string schemas
```

Instead you can set the `Reused` option to `"ref"` to extract these schemas into `$defs`:

```go
jsonSchema, _ := gozod.ToJSONSchema(userSchema, gozod.JSONSchemaOptions{
    Reused: "ref",
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
gozod.ToJSONSchema(schema, gozod.JSONSchemaOptions{IO: "input"})

// In "output" mode, 'c' is required.
// "required" will be ["a", "c"]
gozod.ToJSONSchema(schema, gozod.JSONSchemaOptions{IO: "output"})
```

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

For complex applications with multiple schemas, you can use a registry to manage schema relationships and generate all schemas at once. `ToJSONSchema` can directly accept a registry and will return a single root schema containing all registered schemas in its `$defs`.

```go
// Create a registry for related schemas
registry := gozod.NewRegistry[gozod.GlobalMeta]()

// Define schemas with IDs
var userSchema, postSchema gozod.ZodSchema

userSchema = gozod.Struct(gozod.ObjectSchema{
    "id":   gozod.String().Uuid(),
    "name": gozod.String(),
    "posts": gozod.Lazy(func() gozod.ZodSchema {
        return gozod.Slice(postSchema)
    }),
}).Meta(gozod.GlobalMeta{ID: "User"})

postSchema = gozod.Struct(gozod.ObjectSchema{
    "id":      gozod.String().Uuid(),
    "title":   gozod.String(),
    "author":  gozod.Lazy(func() gozod.ZodSchema { return userSchema }),
}).Meta(gozod.GlobalMeta{ID: "Post"})

registry.Add(userSchema)
registry.Add(postSchema)

// Convert the entire registry to a single root JSON Schema.
// Schemas with IDs will be defined in `$defs` and can be referenced.
rootSchema, _ := gozod.ToJSONSchema(registry)
```
