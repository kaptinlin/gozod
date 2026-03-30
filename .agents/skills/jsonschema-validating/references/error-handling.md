# Error Handling Reference

Comprehensive guide to compilation errors, validation errors, and error formatting in `github.com/kaptinlin/jsonschema`.

## Contents

- Compilation errors
- Validation result structure
- Error iteration patterns
- Localized error messages
- Sentinel errors

## Compilation Errors

Returned by `Compile`, `FromStruct`, and related methods:

```go
schema, err := compiler.Compile(schemaJSON)
if err != nil {
    // Regex validation errors
    if errors.Is(err, jsonschema.ErrRegexValidation) {
        var regexErr *jsonschema.RegexPatternError
        if errors.As(err, &regexErr) {
            fmt.Printf("Keyword: %s\n", regexErr.Keyword)       // "pattern" or "patternProperties"
            fmt.Printf("Location: %s\n", regexErr.Location)     // "#/properties/email/pattern"
            fmt.Printf("Pattern: %s\n", regexErr.Pattern)       // the invalid regex
            fmt.Printf("Error: %s\n", regexErr.Err)             // regexp compile error
        }
    }
}
```

Struct tag compilation errors:

```go
schema, err := jsonschema.FromStruct[User]()
if err != nil {
    var tagErr *jsonschema.StructTagError
    if errors.As(err, &tagErr) {
        fmt.Printf("Struct: %s\n", tagErr.StructType)  // "main.User"
        fmt.Printf("Field: %s\n", tagErr.FieldName)    // "Email"
        fmt.Printf("Rule: %s\n", tagErr.TagRule)        // "pattern=..."
        fmt.Printf("Message: %s\n", tagErr.Message)     // human-readable
        fmt.Printf("Cause: %v\n", tagErr.Err)           // underlying error
    }
}
```

## EvaluationResult

Returned by all `Validate*` methods:

```go
type EvaluationResult struct {
    Valid            bool
    EvaluationPath   string                      // path through schema keywords
    SchemaLocation   string                      // location in schema document
    InstanceLocation string                      // location in validated data
    Annotations      map[string]any              // title, description, default, etc.
    Errors           map[string]*EvaluationError  // keyword -> error
    Details          []*EvaluationResult          // nested sub-results
}
```

## EvaluationError

Each validation failure:

```go
type EvaluationError struct {
    Keyword string         // JSON Schema keyword (e.g., "required", "minLength")
    Code    string         // machine-readable code (e.g., "required_missing")
    Message string         // English message template with {placeholders}
    Params  map[string]any // template parameters
}
```

## Error Iteration Patterns

### Top-level errors only

```go
result := schema.Validate(data)
if !result.IsValid() {
    for keyword, evalErr := range result.Errors {
        fmt.Printf("%s: %s\n", keyword, evalErr.Message)
    }
}
```

### Hierarchical list (default)

```go
list := result.ToList() // or result.ToList(true)
// list.Errors: top-level errors
// list.Details: nested results preserving schema structure
for _, detail := range list.Details {
    if !detail.Valid {
        fmt.Printf("At %s:\n", detail.InstanceLocation)
        for keyword, msg := range detail.Errors {
            fmt.Printf("  %s: %s\n", keyword, msg)
        }
    }
}
```

### Flat list (no hierarchy)

```go
list := result.ToList(false)
// All errors flattened into list.Details without nesting
for _, detail := range list.Details {
    for keyword, msg := range detail.Errors {
        fmt.Printf("[%s] %s: %s\n", detail.InstanceLocation, keyword, msg)
    }
}
```

### Detailed errors (leaf-level, path-keyed)

```go
errs := result.DetailedErrors()
// map[string]string: "fieldPath" -> "error message"
for path, msg := range errs {
    fmt.Printf("%s: %s\n", path, msg)
}
```

## Localized Error Messages

```go
bundle, _ := jsonschema.I18n()
localizer := bundle.NewLocalizer("zh-Hans")

// Localized list
list := result.ToLocalizeList(localizer)
for field, msg := range list.Errors {
    fmt.Printf("%s: %s\n", field, msg) // Chinese messages
}

// Localized detailed errors
errs := result.DetailedErrors(localizer)

// Localize a single error
evalErr := result.Errors["required"]
localizedMsg := evalErr.Localize(localizer)
```

### Supported Locales

| Code | Language |
|------|----------|
| `en` | English (default) |
| `zh-Hans` | Simplified Chinese |
| `zh-Hant` | Traditional Chinese |
| `de-DE` | German |
| `es-ES` | Spanish |
| `fr-FR` | French |
| `ja-JP` | Japanese |
| `ko-KR` | Korean |
| `pt-BR` | Brazilian Portuguese |

## Unmarshal Errors

```go
err := schema.Unmarshal(&dst, src)
if err != nil {
    var unmarshalErr *jsonschema.UnmarshalError
    if errors.As(err, &unmarshalErr) {
        fmt.Printf("Type: %s\n", unmarshalErr.Type)    // "source", "destination", "defaults", etc.
        fmt.Printf("Field: %s\n", unmarshalErr.Field)   // field name if applicable
        fmt.Printf("Reason: %s\n", unmarshalErr.Reason) // human-readable
        fmt.Printf("Cause: %v\n", unmarshalErr.Err)     // underlying error
    }
}
```

## Key Sentinel Errors

| Error | Trigger |
|-------|---------|
| `ErrRegexValidation` | Invalid regex in `pattern` or `patternProperties` |
| `ErrSchemaCompilation` | General schema compilation failure |
| `ErrReferenceResolution` | Unresolvable `$ref` |
| `ErrInvalidSchemaType` | Invalid `type` value in schema |
| `ErrExpectedStructType` | Non-struct passed to `FromStruct` |
| `ErrStructTagParsing` | Malformed struct tag |
| `ErrNilDestination` | `nil` passed as unmarshal destination |
| `ErrNotPointer` | Non-pointer passed as unmarshal destination |
| `ErrJSONDecode` | Invalid JSON in byte input |

All errors support `errors.Is()` and `errors.As()` for matching.
