# Constructor API Reference

Type-safe schema building without JSON strings. All constructors return `*Schema` that can be validated immediately.

## Contents

- Type constructors
- Keyword functions
- Composition and conditionals
- Convenience helpers
- Format constants

## Type Constructors

| Function | JSON Schema equivalent |
|----------|----------------------|
| `String(keywords...)` | `{"type": "string"}` |
| `Integer(keywords...)` | `{"type": "integer"}` |
| `Number(keywords...)` | `{"type": "number"}` |
| `Boolean(keywords...)` | `{"type": "boolean"}` |
| `Null(keywords...)` | `{"type": "null"}` |
| `Array(keywords...)` | `{"type": "array"}` |
| `Object(items...)` | `{"type": "object"}` |
| `Any(keywords...)` | `{}` (no type restriction) |
| `Const(value)` | `{"const": value}` |
| `Enum(values...)` | `{"enum": [...]}` |

## Object Construction

`Object` accepts a mix of `Property` and `Keyword` arguments:

```go
schema := jsonschema.Object(
    // Properties
    jsonschema.Prop("name", jsonschema.String(jsonschema.MinLength(1))),
    jsonschema.Prop("age", jsonschema.Integer(jsonschema.Min(0))),

    // Keywords
    jsonschema.Required("name"),
    jsonschema.AdditionalProps(false),
)
```

## Keyword Functions

### String Keywords

| Function | JSON Schema keyword |
|----------|-------------------|
| `MinLength(n)` | `minLength` |
| `MaxLength(n)` | `maxLength` |
| `Pattern(regex)` | `pattern` |
| `Format(name)` | `format` |

### Number Keywords

| Function | JSON Schema keyword |
|----------|-------------------|
| `Min(n)` | `minimum` |
| `Max(n)` | `maximum` |
| `ExclusiveMin(n)` | `exclusiveMinimum` |
| `ExclusiveMax(n)` | `exclusiveMaximum` |
| `MultipleOf(n)` | `multipleOf` |

### Array Keywords

| Function | JSON Schema keyword |
|----------|-------------------|
| `Items(schema)` | `items` |
| `MinItems(n)` | `minItems` |
| `MaxItems(n)` | `maxItems` |
| `UniqueItems(bool)` | `uniqueItems` |
| `Contains(schema)` | `contains` |
| `MinContains(n)` | `minContains` |
| `MaxContains(n)` | `maxContains` |
| `PrefixItems(schemas...)` | `prefixItems` |
| `UnevaluatedItems(schema)` | `unevaluatedItems` |

### Object Keywords

| Function | JSON Schema keyword |
|----------|-------------------|
| `Required(fields...)` | `required` |
| `AdditionalProps(bool)` | `additionalProperties` (boolean) |
| `AdditionalPropsSchema(schema)` | `additionalProperties` (schema) |
| `MinProps(n)` | `minProperties` |
| `MaxProps(n)` | `maxProperties` |
| `PatternProps(map)` | `patternProperties` |
| `PropertyNames(schema)` | `propertyNames` |
| `UnevaluatedProps(schema)` | `unevaluatedProperties` |
| `DependentRequired(map)` | `dependentRequired` |
| `DependentSchemas(map)` | `dependentSchemas` |

### Annotation Keywords

| Function | JSON Schema keyword |
|----------|-------------------|
| `Title(s)` | `title` |
| `Description(s)` | `description` |
| `Default(v)` | `default` |
| `Examples(values...)` | `examples` |
| `Deprecated(bool)` | `deprecated` |
| `ReadOnly(bool)` | `readOnly` |
| `WriteOnly(bool)` | `writeOnly` |

### Core Keywords

| Function | JSON Schema keyword |
|----------|-------------------|
| `ID(s)` | `$id` |
| `SchemaURI(s)` | `$schema` |
| `Anchor(s)` | `$anchor` |
| `DynamicAnchor(s)` | `$dynamicAnchor` |
| `Defs(map)` | `$defs` |

### Content Keywords

| Function | JSON Schema keyword |
|----------|-------------------|
| `ContentEncoding(s)` | `contentEncoding` |
| `ContentMediaType(s)` | `contentMediaType` |
| `ContentSchema(schema)` | `contentSchema` |

## Composition

```go
// OneOf - exactly one must match
jsonschema.OneOf(jsonschema.String(), jsonschema.Integer())

// AnyOf - at least one must match
jsonschema.AnyOf(jsonschema.String(jsonschema.MinLength(1)), jsonschema.Null())

// AllOf - all must match
jsonschema.AllOf(baseSchema, extensionSchema)

// Not - must not match
jsonschema.Not(jsonschema.String())

// Ref - $ref to another schema
jsonschema.Ref("#/$defs/Address")
```

## Conditional Logic

```go
schema := jsonschema.If(
    jsonschema.Object(
        jsonschema.Prop("type", jsonschema.Const("business")),
        jsonschema.Required("type"),
    ),
).Then(
    jsonschema.Object(
        jsonschema.Prop("taxId", jsonschema.String()),
        jsonschema.Required("taxId"),
    ),
).Else(
    jsonschema.Object(
        jsonschema.Prop("ssn", jsonschema.String()),
    ),
)
```

## Convenience Schema Helpers

Pre-built format schemas:

| Function | Equivalent |
|----------|-----------|
| `Email()` | `String(Format("email"))` |
| `DateTime()` | `String(Format("date-time"))` |
| `Date()` | `String(Format("date"))` |
| `Time()` | `String(Format("time"))` |
| `URI()` | `String(Format("uri"))` |
| `UUID()` | `String(Format("uuid"))` |
| `IPv4()` | `String(Format("ipv4"))` |
| `IPv6()` | `String(Format("ipv6"))` |
| `Hostname()` | `String(Format("hostname"))` |
| `Duration()` | `String(Format("duration"))` |
| `Regex()` | `String(Format("regex"))` |
| `PositiveInt()` | `Integer(Min(1))` |
| `NonNegativeInt()` | `Integer(Min(0))` |

## Custom Compiler for Constructed Schemas

```go
compiler := jsonschema.NewCompiler()
compiler.RegisterDefaultFunc("now", jsonschema.DefaultNowFunc)

schema := jsonschema.Object(
    jsonschema.Prop("createdAt", jsonschema.String(jsonschema.Default("now()"))),
).SetCompiler(compiler)

// Child schemas inherit parent's compiler
```

## Format Constants

Use these constants instead of string literals:

```go
jsonschema.FormatEmail          // "email"
jsonschema.FormatDateTime       // "date-time"
jsonschema.FormatDate           // "date"
jsonschema.FormatTime           // "time"
jsonschema.FormatURI            // "uri"
jsonschema.FormatUUID           // "uuid"
jsonschema.FormatIPv4           // "ipv4"
jsonschema.FormatIPv6           // "ipv6"
jsonschema.FormatHostname       // "hostname"
jsonschema.FormatRegex          // "regex"
jsonschema.FormatDuration       // "duration"
jsonschema.FormatURIRef         // "uri-reference"
jsonschema.FormatURITemplate    // "uri-template"
jsonschema.FormatJSONPointer    // "json-pointer"
jsonschema.FormatIdnEmail       // "idn-email"
jsonschema.FormatIdnHostname    // "idn-hostname"
jsonschema.FormatIRI            // "iri"
jsonschema.FormatIRIRef         // "iri-reference"
jsonschema.FormatRelativeJSONPointer // "relative-json-pointer"
```
