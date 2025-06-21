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

jsonSchema := gozod.ToJSONSchema(schema)
// Returns a *jsonschema.Schema instance ready for validation
```

All GoZod schemas and validation methods are converted to their closest JSON Schema equivalent. Some Go types have no analog and cannot be reasonably represented. See the [`unrepresentable`](#unrepresentable-types) section below for more information on handling these cases.

## Direct Integration with kaptinlin/jsonschema

GoZod's `ToJSONSchema()` function returns a `*jsonschema.Schema` instance from the [`kaptinlin/jsonschema`](https://github.com/kaptinlin/jsonschema) library, which provides robust JSON Schema validation compliant with Draft 2020-12. No additional compilation step is needed:

```go
import (
    "encoding/json"
    "fmt"
    "log"
    
    "github.com/kaptinlin/gozod"
)

func main() {
    // Define GoZod schema
    userSchema := gozod.Struct(gozod.ObjectSchema{
        "username": gozod.String().Min(3),
        "email":    gozod.Email(),
        "age":      gozod.Int().Min(18),
    })
    
    // Convert to JSON Schema - returns *jsonschema.Schema directly
    validator := gozod.ToJSONSchema(userSchema)
    
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
gozod.Email()          // => {"type": "string", "format": "email"}
gozod.ISO.DateTime()   // => {"type": "string", "format": "date-time"}
gozod.ISO.Date()       // => {"type": "string", "format": "date"}
gozod.ISO.Time()       // => {"type": "string", "format": "time"}
gozod.ISO.Duration()   // => {"type": "string", "format": "duration"}
gozod.IPv4()           // => {"type": "string", "format": "ipv4"}
gozod.IPv6()           // => {"type": "string", "format": "ipv6"}
gozod.UUID()           // => {"type": "string", "format": "uuid"}
gozod.UUIDv4()         // => {"type": "string", "format": "uuid"}
gozod.URL()            // => {"type": "string", "format": "uri"}
```

These schemas are supported via `contentEncoding`:

```go
gozod.Base64()         // => {"type": "string", "contentEncoding": "base64"}
```

String patterns and custom formats are supported via `pattern`:

```go
gozod.String().Regex(regexp.MustCompile("^[a-z]+$"))
// => {"type": "string", "pattern": "^[a-z]+$"}

gozod.CIDRv4()         // => {"type": "string", "pattern": "..."}
gozod.CIDRv6()         // => {"type": "string", "pattern": "..."}
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

GoZod converts both `nil` values to `{"type": "null"}` in JSON Schema:

```go
gozod.Nil()       // => {"type": "null"}
gozod.Any()       // => {} (represents any value)
```

Similarly, `Optional` and `Nilable` are made nullable via `oneOf`:

```go
gozod.String().Optional()
// => {"oneOf": [{"type": "string"}, {"type": "null"}]}

gozod.String().Nilable()
// => {"oneOf": [{"type": "string"}, {"type": "null"}]}
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
    Metadata *gozod.Registry
    
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
}
```

### Metadata

GoZod supports metadata storage that will be included in the generated JSON Schema:

```go
// Add metadata to a schema
emailSchema := gozod.String().Meta(gozod.GlobalMetadata{
    Title:       "Email Address",
    Description: "User's email address",
    Examples:    []any{"user@example.com"},
})

jsonSchema := gozod.ToJSONSchema(emailSchema)
// The returned *jsonschema.Schema will include all metadata
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
}, gozod.SchemaParams{
    Error: "String must be longer than 5 characters",
})

// ❌ Function schemas cannot be represented in JSON Schema
gozod.Function(&gozod.FunctionOptions{
    Input:  gozod.String(),
    Output: gozod.String(),
})
```

By default, GoZod will throw an error if any of these are encountered:

```go
schema := gozod.String().Transform(func(s string) string {
    return strings.ToUpper(s)
})

_, err := gozod.ToJSONSchema(schema)
// => returns error: "transform functions cannot be represented in JSON Schema"
```

You can change this behavior by setting the `Unrepresentable` option to `"any"`. This will convert any unrepresentable types to `{}` (the equivalent of `any` in JSON Schema):

```go
schema := gozod.String().Transform(func(s string) string {
    return strings.ToUpper(s)
})

jsonSchema := gozod.ToJSONSchema(schema, gozod.JSONSchemaOptions{
    Unrepresentable: "any",
})
// => returns *jsonschema.Schema representing {} (accepts any value)
```

### Cycles

How to handle cycles. If a cycle is encountered as `gozod.ToJSONSchema()` traverses the schema, it will be represented using `$ref`:

```go
// Define a recursive user schema
var UserSchema gozod.ZodType[any, any]
UserSchema = gozod.Struct(gozod.ObjectSchema{
    "name": gozod.String(),
    "friend": gozod.Lazy(func() gozod.ZodType[any, any] {
        return UserSchema
    }),
})

jsonSchema := gozod.ToJSONSchema(UserSchema)
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

jsonSchema := gozod.ToJSONSchema(userSchema)
// Both firstName and lastName will have inlined string schemas
```

Instead you can set the `Reused` option to `"ref"` to extract these schemas into `$defs`:

```go
jsonSchema := gozod.ToJSONSchema(userSchema, gozod.JSONSchemaOptions{
    Reused: "ref",
})
// Common schemas will be extracted to $defs and referenced
```

## Advanced Integration Example

Here's a comprehensive example showing how to use GoZod with the `kaptinlin/jsonschema` library features:

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    
    "github.com/kaptinlin/gozod"
    "github.com/kaptinlin/jsonschema"
)

func main() {
    // Define a complex GoZod schema
    productSchema := gozod.Struct(gozod.ObjectSchema{
        "id": gozod.String().UUID().Meta(gozod.GlobalMetadata{
            Description: "Unique product identifier",
        }),
        "name": gozod.String().Min(1).Max(100),
        "price": gozod.Float64().Min(0),
        "tags": gozod.Slice(gozod.String()).Optional(),
        "metadata": gozod.Record(
            gozod.String(),
            gozod.Any(),
        ).Optional(),
        "created_at": gozod.ISO.DateTime(),
    })
    
    // Convert to JSON Schema - returns *jsonschema.Schema directly
    validator := gozod.ToJSONSchema(productSchema, gozod.JSONSchemaOptions{
        Unrepresentable: "any",
    })
    
    // Test data validation
    validProduct := map[string]any{
        "id":         "550e8400-e29b-41d4-a716-446655440000",
        "name":       "Widget Pro",
        "price":      29.99,
        "tags":       []string{"widget", "premium"},
        "created_at": "2023-01-01T12:00:00Z",
    }
    
    invalidProduct := map[string]any{
        "id":    "invalid-uuid",
        "name":  "", // too short
        "price": -10, // negative price
    }
    
    // Validate valid product
    result := validator.Validate(validProduct)
    if result.IsValid() {
        fmt.Println("✅ Valid product passed validation")
    }
    
    // Validate invalid product
    result = validator.Validate(invalidProduct)
    if !result.IsValid() {
        fmt.Println("❌ Invalid product failed validation:")
        
        // Get detailed error information
        details, _ := json.MarshalIndent(result.ToList(), "", "  ")
        fmt.Println(string(details))
        
        // Optional: Use multilingual error messages
        i18n, err := jsonschema.GetI18n()
        if err == nil {
            localizer := i18n.NewLocalizer("en")
            localizedDetails, _ := json.MarshalIndent(result.ToLocalizeList(localizer), "", "  ")
            fmt.Println("Localized errors:", string(localizedDetails))
        }
    }
}
```

## Custom JSON Encoder/Decoder

Since GoZod returns a `*jsonschema.Schema` directly, you can configure the underlying JSON processing if needed:

```go
// Create a custom compiler for high-performance JSON processing
import "github.com/bytedance/sonic"

// If you need to serialize the schema to JSON later
validator := gozod.ToJSONSchema(schema)

// For high-performance applications, you can configure the underlying
// jsonschema library with custom JSON encoders
compiler := jsonschema.NewCompiler()
compiler.WithEncoderJSON(sonic.Marshal)
compiler.WithDecoderJSON(sonic.Unmarshal)

// Use the compiler for other schemas loaded from JSON
externalSchema, err := compiler.GetSchema("https://json-schema.org/draft/2020-12/schema")
```

## Working with Registries

For complex applications with multiple schemas, you can use registries to manage schema relationships:

```go
// Create a registry for related schemas
registry := gozod.NewRegistry()

// Define schemas with IDs
userSchema := gozod.Struct(gozod.ObjectSchema{
    "id":   gozod.String().UUID(),
    "name": gozod.String(),
}).Meta(gozod.GlobalMetadata{ID: "User"})

postSchema := gozod.Struct(gozod.ObjectSchema{
    "id":       gozod.String().UUID(),
    "title":    gozod.String(),
    "content":  gozod.String(),
    "author":   gozod.Ref("User"), // Reference to User schema
}).Meta(gozod.GlobalMetadata{ID: "Post"})

// Register schemas
registry.Add("User", userSchema)
registry.Add("Post", postSchema)

// Convert registry to JSON Schema collection
validators := gozod.ToJSONSchemaCollection(registry, gozod.JSONSchemaOptions{
    URI: func(id string) string {
        return fmt.Sprintf("https://api.example.com/schemas/%s.json", id)
    },
})

// validators is a map[string]*jsonschema.Schema
userValidator := validators["User"]
postValidator := validators["Post"]
```
