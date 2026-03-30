# Struct Tag Schema Generation

Generate JSON Schemas from Go struct definitions using `jsonschema` struct tags.

## Contents

- Basic usage
- Tag syntax and supported rules
- Options and configuration
- Code generation with schemagen
- Circular references

## Basic Usage

```go
type User struct {
    Name  string `jsonschema:"required,minLength=2,maxLength=50"`
    Email string `jsonschema:"required,format=email"`
    Age   int    `jsonschema:"minimum=18,maximum=120"`
}

// Generate schema (cached after first call)
schema, err := jsonschema.FromStruct[User]()
if err != nil {
    log.Fatal(err)
}

// Validate
result := schema.Validate(data)
```

## Tag Syntax

Tags use the `jsonschema` key. Rules are comma-separated. Parameters use `=` delimiter. Multi-value parameters use `|` delimiter.

```go
type Example struct {
    Field string `jsonschema:"required,minLength=1,maxLength=100"`
    Role  string `jsonschema:"enum=admin|user|guest"`
    Score float64 `jsonschema:"minimum=0,maximum=100,multipleOf=0.5"`
}
```

## Supported Tag Rules

### String Rules

| Rule | Example | JSON Schema |
|------|---------|-------------|
| `minLength=N` | `minLength=1` | `{"minLength": 1}` |
| `maxLength=N` | `maxLength=100` | `{"maxLength": 100}` |
| `pattern=REGEX` | `pattern=^[a-z]+$` | `{"pattern": "^[a-z]+$"}` |
| `format=FMT` | `format=email` | `{"format": "email"}` |

### Numeric Rules

| Rule | Example | JSON Schema |
|------|---------|-------------|
| `minimum=N` | `minimum=0` | `{"minimum": 0}` |
| `maximum=N` | `maximum=100` | `{"maximum": 100}` |
| `exclusiveMinimum=N` | `exclusiveMinimum=0` | `{"exclusiveMinimum": 0}` |
| `exclusiveMaximum=N` | `exclusiveMaximum=100` | `{"exclusiveMaximum": 100}` |
| `multipleOf=N` | `multipleOf=0.5` | `{"multipleOf": 0.5}` |

### Array Rules

| Rule | Example | JSON Schema |
|------|---------|-------------|
| `minItems=N` | `minItems=1` | `{"minItems": 1}` |
| `maxItems=N` | `maxItems=10` | `{"maxItems": 10}` |
| `uniqueItems` | `uniqueItems` | `{"uniqueItems": true}` |
| `contains=TYPE` | `contains=string` | `{"contains": {"type": "string"}}` |
| `prefixItems=T1\|T2` | `prefixItems=string\|integer` | `{"prefixItems": [...]}` |

### Object Rules

| Rule | Example | JSON Schema |
|------|---------|-------------|
| `additionalProperties=BOOL` | `additionalProperties=false` | `{"additionalProperties": false}` |
| `minProperties=N` | `minProperties=1` | `{"minProperties": 1}` |
| `maxProperties=N` | `maxProperties=10` | `{"maxProperties": 10}` |
| `patternProperties=PAT\|TYPE` | `patternProperties=^x-\|string` | `{"patternProperties": {...}}` |

### Composition Rules

| Rule | Example | JSON Schema |
|------|---------|-------------|
| `allOf=T1\|T2` | `allOf=string\|integer` | `{"allOf": [...]}` |
| `anyOf=T1\|T2` | `anyOf=string\|null` | `{"anyOf": [...]}` |
| `oneOf=T1\|T2` | `oneOf=string\|integer` | `{"oneOf": [...]}` |
| `not=TYPE` | `not=null` | `{"not": {"type": "null"}}` |

### Metadata Rules

| Rule | Example | JSON Schema |
|------|---------|-------------|
| `required` | `required` | Added to parent `required` array |
| `title=TEXT` | `title=User Name` | `{"title": "User Name"}` |
| `description=TEXT` | `description=The user name` | `{"description": "..."}` |
| `default=VAL` | `default=active` | `{"default": "active"}` |
| `examples=V1\|V2` | `examples=foo\|bar` | `{"examples": ["foo","bar"]}` |
| `deprecated` | `deprecated` | `{"deprecated": true}` |
| `readOnly` | `readOnly` | `{"readOnly": true}` |
| `writeOnly` | `writeOnly` | `{"writeOnly": true}` |
| `enum=V1\|V2` | `enum=a\|b\|c` | `{"enum": ["a","b","c"]}` |
| `const=VAL` | `const=fixed` | `{"const": "fixed"}` |

### Reference Rules

| Rule | Example |
|------|---------|
| `ref=URI` | `ref=#/$defs/Address` |
| `anchor=NAME` | `anchor=addressDef` |

## Options

```go
schema, err := jsonschema.FromStructWithOptions[User](&jsonschema.StructTagOptions{
    TagName:             "jsonschema",    // tag name to parse (default)
    AllowUntaggedFields: false,           // include fields without tags
    DefaultRequired:     false,           // make all fields required by default
    CacheEnabled:        true,            // cache generated schemas
    SchemaVersion:       "https://json-schema.org/draft/2020-12/schema",
    RequiredSort:        jsonschema.RequiredSortAlphabetical, // deterministic ordering

    // Schema-level properties
    SchemaProperties: map[string]any{
        "additionalProperties": false,
        "title":               "User Schema",
    },
})
```

### Cache Management

```go
// Clear all cached schemas
jsonschema.ClearSchemaCache()

// Check cache stats
stats := jsonschema.CacheStats()
fmt.Printf("Cached schemas: %d\n", stats["cached_schemas"])
```

## Pointer Fields (Nullable)

Pointer fields automatically become nullable via `anyOf`:

```go
type Config struct {
    Name    string  `jsonschema:"required"`
    Timeout *int    `jsonschema:"minimum=1"`
}
// Timeout generates: {"anyOf": [{"type": "integer", "minimum": 1}, {"type": "null"}]}
```

## Nested Structs and Circular References

Nested structs are placed in `$defs` and referenced via `$ref`. Circular references are detected automatically:

```go
type TreeNode struct {
    Value    string      `jsonschema:"required"`
    Children []*TreeNode `jsonschema:""`
}
// Generates $defs/TreeNode with $ref for Children items
```

## Error Handling

```go
schema, err := jsonschema.FromStruct[User]()
if err != nil {
    var tagErr *jsonschema.StructTagError
    if errors.As(err, &tagErr) {
        fmt.Printf("Struct: %s, Field: %s, Rule: %s\n",
            tagErr.StructType, tagErr.FieldName, tagErr.TagRule)
    }
}
```

## Code Generation with schemagen

For maximum performance, pre-generate schema methods at build time:

```bash
# Install
go install github.com/kaptinlin/jsonschema/cmd/schemagen@latest

# Generate for current package
schemagen

# Add to source for go generate
//go:generate schemagen
```

This generates `*_schema.go` files with compiled schema methods, eliminating runtime reflection.

## Custom Validators

Register custom tag validators globally:

```go
jsonschema.RegisterCustomValidator("phone", func(t reflect.Type, params []string) []jsonschema.Keyword {
    return []jsonschema.Keyword{
        jsonschema.Pattern(`^\+?[1-9]\d{1,14}$`),
    }
})

type Contact struct {
    Phone string `jsonschema:"phone"`
}
```
