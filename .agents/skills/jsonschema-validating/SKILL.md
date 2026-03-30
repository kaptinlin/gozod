---
description: Validate Go data against JSON Schema Draft 2020-12 using github.com/kaptinlin/jsonschema. Use when compiling JSON schemas, validating JSON/structs/maps, unmarshaling with defaults, building schemas programmatically, or generating schemas from struct tags.
name: jsonschema-validating
---


# JSON Schema Validation with kaptinlin/jsonschema

Validate Go data against JSON Schema Draft 2020-12 using `github.com/kaptinlin/jsonschema`.

## Decision Flowchart

```
How do you define the schema?
|
+- JSON schema string/file
|  +- compiler.Compile([]byte) -> *Schema
|
+- Go struct tags (jsonschema:"...")
|  +- jsonschema.FromStruct[T]() -> *Schema
|
+- Programmatic (type-safe constructors)
|  +- jsonschema.Object(Prop(...), ...) -> *Schema
|     (no compilation step needed)
|
What data type are you validating?
|
+- []byte JSON        -> schema.ValidateJSON(data)
+- Go struct          -> schema.ValidateStruct(data)
+- map[string]any     -> schema.ValidateMap(data)
+- Unknown / mixed    -> schema.Validate(data)  (auto-detect)
|
Need to unmarshal with defaults?
|
+- Yes -> schema.Validate(data) first, then schema.Unmarshal(&dst, data)
+- No  -> just use validation result
```

## Quick Start

```go
import "github.com/kaptinlin/jsonschema"

// 1. Compile schema
compiler := jsonschema.NewCompiler()
schema, err := compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "name":  {"type": "string", "minLength": 1},
        "email": {"type": "string", "format": "email"},
        "age":   {"type": "integer", "minimum": 0}
    },
    "required": ["name", "email"]
}`))
if err != nil {
    log.Fatal(err) // compilation error (e.g., invalid regex)
}

// 2. Validate
result := schema.ValidateJSON([]byte(`{"name": "Alice", "email": "alice@example.com", "age": 30}`))
if !result.IsValid() {
    for field, evalErr := range result.Errors {
        fmt.Printf("%s: %s\n", field, evalErr.Message)
    }
}
```

## Compiler Configuration

```go
compiler := jsonschema.NewCompiler()

// Enable format assertion (disabled by default per Draft 2020-12)
compiler.SetAssertFormat(true)

// Preserve extension fields (x-component, x-go-type, etc.)
compiler.SetPreserveExtra(true)

// Set base URI for resolving relative $ref
compiler.SetDefaultBaseURI("https://example.com/schemas/")

// Register custom format validator
compiler.RegisterFormat("custom-id", func(v any) bool {
    s, ok := v.(string)
    return ok && strings.HasPrefix(s, "ID-")
}, "string")

// Register dynamic default function
compiler.RegisterDefaultFunc("uuid", func(args ...any) (any, error) {
    return uuid.New().String(), nil
})
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)

// Compile with explicit URI
schema, err := compiler.Compile(schemaJSON, "https://example.com/schemas/user.json")

// MustCompile panics on error (useful for package-level vars)
schema = compiler.MustCompile(schemaJSON)

// Batch compile interdependent schemas efficiently
schemas, err := compiler.CompileBatch(map[string][]byte{
    "https://example.com/user.json":    userSchemaJSON,
    "https://example.com/address.json": addressSchemaJSON,
})
```

## Schema Construction API -- [details](references/constructor-api.md)

Build schemas programmatically without JSON:

```go
schema := jsonschema.Object(
    jsonschema.Prop("name", jsonschema.String(jsonschema.MinLength(1))),
    jsonschema.Prop("email", jsonschema.Email()),
    jsonschema.Prop("age", jsonschema.Integer(jsonschema.Min(0), jsonschema.Max(150))),
    jsonschema.Prop("tags", jsonschema.Array(jsonschema.Items(jsonschema.String()), jsonschema.UniqueItems(true))),
    jsonschema.Required("name", "email"),
    jsonschema.AdditionalProps(false),
)
// Validate immediately - no compilation step needed
result := schema.Validate(data)
```

## Struct Tag Schema Generation -- [details](references/struct-tags.md)

Generate schemas from Go struct definitions:

```go
type User struct {
    Name  string `jsonschema:"required,minLength=2,maxLength=50"`
    Email string `jsonschema:"required,format=email"`
    Age   int    `jsonschema:"minimum=18,maximum=120"`
    Role  string `jsonschema:"enum=admin|user|guest,default=user"`
}

schema, err := jsonschema.FromStruct[User]()
if err != nil {
    var tagErr *jsonschema.StructTagError
    if errors.As(err, &tagErr) {
        log.Printf("Field %s: %s", tagErr.FieldName, tagErr.Message)
    }
}
result := schema.Validate(userData)
```

Code generation for performance:

```bash
go install github.com/kaptinlin/jsonschema/cmd/schemagen@latest
schemagen                    # generate for current package
//go:generate schemagen      # add to struct files
```

## Validation and Unmarshal Workflow

```go
data := []byte(`{"name": "Alice"}`)

// Step 1: Validate
result := schema.Validate(data)
if !result.IsValid() {
    // Handle errors (see Error Handling below)
    return
}

// Step 2: Unmarshal with defaults applied
var user User
if err := schema.Unmarshal(&user, data); err != nil {
    log.Fatal(err)
}
// user.Role == "user" (default applied)
```

Unmarshal accepts `[]byte`, `map[string]any`, or structs as source. Destination must be a non-nil pointer.

## Error Handling -- [details](references/error-handling.md)

### Compilation Errors

```go
schema, err := compiler.Compile(schemaJSON)
if err != nil {
    if errors.Is(err, jsonschema.ErrRegexValidation) {
        // Invalid regex pattern in schema
        var regexErr *jsonschema.RegexPatternError
        if errors.As(err, &regexErr) {
            log.Printf("Bad pattern at %s: %s", regexErr.Location, regexErr.Pattern)
        }
    }
    log.Fatal(err)
}
```

### Validation Errors

```go
result := schema.Validate(data)
if !result.IsValid() {
    // Top-level errors (field -> EvaluationError)
    for field, evalErr := range result.Errors {
        fmt.Printf("%s: %s (keyword=%s)\n", field, evalErr.Message, evalErr.Keyword)
    }

    // Flat list with schema locations
    list := result.ToList()
    for _, detail := range list.Details {
        for keyword, msg := range detail.Errors {
            fmt.Printf("[%s] %s: %s\n", detail.InstanceLocation, keyword, msg)
        }
    }

    // Detailed errors from nested hierarchy
    detailedErrs := result.DetailedErrors()
    for path, msg := range detailedErrs {
        fmt.Printf("%s: %s\n", path, msg)
    }
}
```

### EvaluationResult Structure

| Field | Type | Description |
|-------|------|-------------|
| `Valid` | `bool` | Whether validation passed |
| `Errors` | `map[string]*EvaluationError` | Keyword-keyed validation errors |
| `Details` | `[]*EvaluationResult` | Nested sub-schema results |
| `EvaluationPath` | `string` | Path through schema keywords |
| `SchemaLocation` | `string` | Location in the schema document |
| `InstanceLocation` | `string` | Location in the validated data |

### Output Formats

| Method | Returns | Use Case |
|--------|---------|----------|
| `result.IsValid()` | `bool` | Quick pass/fail check |
| `result.ToFlag()` | `*Flag` | Simple `{valid: bool}` |
| `result.ToList()` | `*List` | Flat error list with locations |
| `result.ToList(false)` | `*List` | Flat list without hierarchy |
| `result.DetailedErrors()` | `map[string]string` | Field path to message map |
| `result.ToLocalizeList(localizer)` | `*List` | Localized error list |
| `result.DetailedErrors(localizer)` | `map[string]string` | Localized field errors |

## Internationalization

```go
bundle, _ := jsonschema.I18n()
// Supported: en, zh-Hans, zh-Hant, de-DE, es-ES, fr-FR, ja-JP, ko-KR, pt-BR
localizer := bundle.NewLocalizer("zh-Hans")

result := schema.Validate(data)
if !result.IsValid() {
    localizedList := result.ToLocalizeList(localizer)
    for field, msg := range localizedList.Errors {
        fmt.Printf("%s: %s\n", field, msg)
    }
}
```

## Schema References ($ref)

```go
// Internal references
schemaJSON := `{
    "type": "object",
    "properties": {
        "billing":  {"$ref": "#/$defs/Address"},
        "shipping": {"$ref": "#/$defs/Address"}
    },
    "$defs": {
        "Address": {
            "type": "object",
            "properties": {
                "street": {"type": "string"},
                "city":   {"type": "string"}
            },
            "required": ["street", "city"]
        }
    }
}`

// Cross-schema references: compile schemas with URIs
compiler.Compile(addressJSON, "https://example.com/address.json")
compiler.Compile([]byte(`{
    "type": "object",
    "properties": {
        "home": {"$ref": "https://example.com/address.json"}
    }
}`))
```

Supported reference types: `$ref`, `$dynamicRef`, `$anchor`, `$dynamicAnchor`.

## Extension Fields

```go
compiler := jsonschema.NewCompiler()
compiler.SetPreserveExtra(true)

schema, _ := compiler.Compile([]byte(`{
    "type": "string",
    "x-component": "DatePicker",
    "x-component-props": {"format": "YYYY-MM-DD"}
}`))

component := schema.Extra["x-component"] // "DatePicker"
```

## Performance Considerations

| Technique | Impact |
|-----------|--------|
| Pre-compile schemas (reuse `*Schema`) | Avoid re-parsing on every validation |
| Use type-specific methods (`ValidateJSON`, `ValidateStruct`, `ValidateMap`) | 5-10x faster than reflection path |
| `CompileBatch` for interdependent schemas | Single-pass reference resolution |
| `FromStruct` caches by default (`CacheEnabled: true`) | Avoids repeated reflection |
| `schemagen` code generation | Zero runtime reflection overhead |

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Recompiling schema on every request | Wastes CPU parsing JSON | Compile once, reuse `*Schema` |
| Using `Validate()` for known types | Auto-detection has overhead | Use `ValidateJSON`, `ValidateStruct`, or `ValidateMap` |
| Ignoring compilation errors | Invalid regex silently passes | Always check `err` from `Compile` |
| Unmarshaling without validating first | Defaults applied to invalid data | Always `Validate()` then `Unmarshal()` |
| Assuming format is validated by default | Draft 2020-12 treats format as annotation | Call `compiler.SetAssertFormat(true)` |
| Mutating source data passed to `Unmarshal` | `Unmarshal` deep-copies, but source maps shared elsewhere may not | Pass `[]byte` or fresh maps |
| Using `MustCompile` with user input | Panics on bad schema | Use `Compile` and handle `error` |
