# Metadata & Registries

GoZod ships with a first-class *Registry* API that allows you to attach **strongly-typed metadata** to any `Schema` instance. Registries unlock a wide range of downstream use-cases: documentation, JSON-Schema generation, UI form builders, code generation, AI structured output—and anything else that benefits from extra context.

> The Registry design follows the spirit of TypeScript Zod 4 but is purposely kept **minimal**. You get the same expressive power without over-engineering.

---

## Quick Start

```go
import "github.com/kaptinlin/gozod"

// 1) Define a metadata struct – any comparable type works.
type FieldMeta struct {
    Title       string            `json:"title"`
    Description string            `json:"description,omitempty"`
    Examples    []any             `json:"examples,omitempty"`
    Extra       map[string]string `json:"extra,omitempty"`
}

// 2) Create a registry that stores FieldMeta
fieldReg := gozod.NewRegistry[FieldMeta]()

// 3) Create a schema and register metadata
nameSchema := gozod.String().Min(1)
fieldReg.Add(nameSchema, FieldMeta{
    Title:       "User Name",
    Description: "The user's full name",
    Examples:    []any{"Jane Doe"},
})

// 4) Retrieve metadata anywhere in your code
if meta, ok := fieldReg.Get(nameSchema); ok {
    fmt.Println(meta.Title) // => "User Name"
}
```

### Registry API Cheatsheet

| Method              | Purpose                                        |
|---------------------|------------------------------------------------|
| `Add(schema, meta)` | Attach or replace metadata                      |
| `Get(schema)`       | Retrieve metadata **and** existence flag        |
| `Remove(schema)`    | Delete metadata                                 |
| `Has(schema)`       | Check if a schema is registered                 |

Each method returns the registry itself so you can chain calls if desired.

---

## Instance Methods (Recommended)

All 26 GoZod schema types now include built-in `.Describe()` and `.Meta()` methods:

```go
// Describe - shorthand for description only
schema := gozod.String().Min(3).Describe("Username")

// Meta - full metadata object
schema := gozod.Object(core.ObjectSchema{
    "name":  gozod.String().Describe("Full name"),
    "email": gozod.String().Email().Meta(gozod.GlobalMeta{
        Title:       "Email Address",
        Description: "Primary contact email",
        Examples:    []any{"user@example.com"},
    }),
})

// Retrieve metadata
meta, ok := gozod.GlobalRegistry.Get(schema)
if ok {
    fmt.Println(meta.Description)  // "Primary contact email"
}
```

### Supported Types

All schema types support `.Describe()` and `.Meta()`:

- **Primitives**: String, Int, Float, Bool, BigInt, Complex, Time
- **Collections**: Array, Slice, Tuple, Map, Record, Object, Struct
- **Composites**: Union, Xor, DiscriminatedUnion, Intersection
- **Special**: Any, Unknown, Never, Nil, Lazy, Function, Enum
- **Format**: StringBool, File

---

## Check Factories (Zod v4 Compatible)

For Zod v4 API compatibility, GoZod also provides check factory functions:

```go
import "github.com/kaptinlin/gozod"

// These are equivalent:
schema1 := gozod.String().Describe("Username")
schema2 := gozod.String().Check(gozod.Describe("Username"))

// Check factories allow composition with other checks
schema := gozod.String().Check(
    gozod.Describe("Username"),
    gozod.Meta(gozod.GlobalMeta{Title: "User Name"}),
)
```

---

## Global Registry

For convenience GoZod exposes a *global* registry so you don't have to pass a registry around everywhere.

```go
import "github.com/kaptinlin/gozod"

// Built-in struct that mirrors common JSON-Schema keys.
type GlobalMeta struct {
    ID          string `json:"id,omitempty"`
    Title       string `json:"title,omitempty"`
    Description string `json:"description,omitempty"`
    Examples    []any  `json:"examples,omitempty"`
}

emailSchema := gozod.Email()

// Store metadata in the global registry
gozod.GlobalRegistry.Add(emailSchema, GlobalMeta{
    ID:          "email_address",
    Title:       "Email Address",
    Description: "A valid e-mail address",
})
```

If you prefer local encapsulation you can ignore `GlobalRegistry` entirely and stick with your own registries.

---

## Integration with JSON Schema

The `jsonschema.FromGoZod` converter will automatically consume registries:

* **`ID` extraction** – Schemas that have an `ID` will be hoisted into `$defs` and referenced with `$ref`.
* **Metadata merging** – Keys like `title`, `description` and `examples` will be copied onto the generated JSON-Schema nodes.

---

## Form Builder Example

```go
// Form field metadata for UI rendering
type FormField struct {
    Label       string `json:"label"`
    Placeholder string `json:"placeholder,omitempty"`
    Required    bool   `json:"required"`
}

formReg := gozod.NewRegistry[FormField]()

userSchema := gozod.Object(map[string]gozod.Schema{
    "name":  gozod.String().Min(1),
    "email": gozod.Email(),
})

formReg.Add(userSchema.Shape()["name"], FormField{
    Label:       "Full Name",
    Placeholder: "Jane Doe",
    Required:    true,
})
```

---

## Best Practices

1. **Keep registries focused** – Use different registries for API docs, forms, analytics, etc.
2. **Leverage static typing** – Strongly-typed metadata prevents accidental mis-use.
3. **Validate examples in CI** – Run `schema.Parse(meta.Examples)` to guarantee examples stay in sync.
4. **Version reusable schemas** – Include `ID` + `Version` in `GlobalMeta` for long-lived contracts.
