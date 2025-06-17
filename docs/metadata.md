# Metadata and Registries

It's often useful to associate a schema with some additional *metadata* for documentation, code generation, AI structured outputs, form validation, and other purposes.

> **GoZod Integration** — GoZod provides native JSON Schema conversion that leverages registries to generate idiomatic JSON Schema from GoZod schemas. Refer to the [JSON Schema](./json-schema.md) page for more information.

## Registries

Metadata in GoZod is handled via *registries*. Registries are collections of schemas, each associated with some *strongly-typed* metadata. To create a simple registry:

```go
import "github.com/kaptinlin/gozod"

// Define metadata structure
type MyMetadata struct {
    Description string `json:"description"`
}

// Create a registry with typed metadata
myRegistry := gozod.NewRegistry[MyMetadata]()
```

To register, lookup, and remove schemas from this registry:

```go
mySchema := gozod.String()

// Add schema with metadata
myRegistry.Add(mySchema, MyMetadata{
    Description: "A cool schema!",
})

// Check if schema exists
exists := myRegistry.Has(mySchema) // => true

// Get metadata for schema
metadata, found := myRegistry.Get(mySchema) // => MyMetadata{Description: "A cool schema!"}, true

// Remove schema from registry
myRegistry.Remove(mySchema)
```

Go's type system enforces that the metadata for each schema matches the registry's **metadata type**.

```go
// ✅ Valid - matches MyMetadata structure
myRegistry.Add(mySchema, MyMetadata{
    Description: "A cool schema!",
})

// ❌ Compile-time error - wrong type
myRegistry.Add(mySchema, "just a string")
```

### `.Register()` Method

> **Note** — This method is special in that it does not return a new schema; instead, it returns the original schema. No other GoZod method does this! That includes `.Meta()` and `.Describe()` (documented below) which return a new instance.

Schemas provide a `.Register()` method to more conveniently add them to a registry:

```go
mySchema := gozod.String()

// Register returns the original schema for method chaining
registeredSchema := mySchema.Register(myRegistry, MyMetadata{
    Description: "A cool schema!",
})
// registeredSchema is the same as mySchema
```

This lets you define metadata "inline" in your schemas:

```go
userSchema := gozod.Struct(gozod.ObjectSchema{
    "name": gozod.String().Register(myRegistry, MyMetadata{
        Description: "The user's name",
    }),
    "age": gozod.Int().Register(myRegistry, MyMetadata{
        Description: "The user's age",
    }),
})
```

### Untyped Registries

If a registry is defined without a metadata type, you can use it as a generic "collection", no metadata required:

```go
// Create an untyped registry
genericRegistry := gozod.NewRegistry[interface{}]()

// Add schemas without specific metadata structure
genericRegistry.Add(gozod.String(), "any metadata")
genericRegistry.Add(gozod.Int(), map[string]string{"type": "number"})
```

## Metadata

### Global Registry

For convenience, GoZod provides a global registry (`gozod.GlobalRegistry`) that can be used to store metadata for JSON Schema generation or other purposes. It accepts the following metadata:

```go
// GlobalMetadata structure
type GlobalMetadata struct {
    ID          string        `json:"id,omitempty"`
    Title       string        `json:"title,omitempty"`
    Description string        `json:"description,omitempty"`
    Examples    []interface{} `json:"examples,omitempty"`
    // Additional properties can be added via embedded struct or interface{}
    Extra       map[string]interface{} `json:"-"` // Custom properties
}
```

To register some metadata in `gozod.GlobalRegistry` for a schema:

```go
import "github.com/kaptinlin/gozod"

emailSchema := gozod.Email().Register(gozod.GlobalRegistry, gozod.GlobalMetadata{
    ID:          "email_address", 
    Title:       "Email address",
    Description: "Your email address",
    Examples:    []interface{}{"first.last@example.com"},
})
```

### `.Meta()` Method

For a more convenient approach, use the `.Meta()` method to register a schema in `gozod.GlobalRegistry`:

```go
emailSchema := gozod.Email().Meta(gozod.GlobalMetadata{
    ID:          "email_address",
    Title:       "Email address", 
    Description: "Please enter a valid email address",
})
```

Calling `.Meta()` without an argument will *retrieve* the metadata for a schema:

```go
metadata, found := emailSchema.Meta()
if found {
    fmt.Printf("Title: %s\n", metadata.Title)
    fmt.Printf("Description: %s\n", metadata.Description)
}
// => Title: Email address
// => Description: Please enter a valid email address
```

Metadata is associated with a *specific schema instance.* This is important to keep in mind, especially since GoZod methods are immutable—they always return a new instance:

```go
schemaA := gozod.String().Meta(gozod.GlobalMetadata{
    Description: "A cool string",
})

metadata, found := schemaA.Meta()
fmt.Println(found) // => true

// Refinement creates a new instance without metadata
schemaB := schemaA.Refine(func(val any) bool { return true })

_, found = schemaB.Meta()
fmt.Println(found) // => false
```

### `.Describe()` Method

> **Compatibility Note** — The `.Describe()` method exists for convenience and compatibility, but `.Meta()` is now the recommended approach for full metadata support.

The `.Describe()` method is a shorthand for registering a schema in `gozod.GlobalRegistry` with just a `Description` field:

```go
emailSchema := gozod.Email()
describedSchema := emailSchema.Describe("An email address")

// Equivalent to:
metaSchema := emailSchema.Meta(gozod.GlobalMetadata{
    Description: "An email address",
})
```

## Custom Registries

You've already seen a simple example of a custom registry. Let's look at some more advanced patterns.

### Advanced Metadata Structures

You can define complex metadata structures to suit your application's needs:

```go
// API documentation metadata
type APIMetadata struct {
    Description string            `json:"description"`
    Examples    []interface{}     `json:"examples"`
    Deprecated  bool              `json:"deprecated,omitempty"`
    Tags        []string          `json:"tags,omitempty"`
    Validation  ValidationRules   `json:"validation,omitempty"`
}

type ValidationRules struct {
    Required bool `json:"required"`
    ReadOnly bool `json:"readOnly"`
}

// Create registry with complex metadata
apiRegistry := gozod.NewRegistry[APIMetadata]()
```

### Form Validation Registry

For form validation use cases:

```go
// Form field metadata
type FormFieldMetadata struct {
    Label       string `json:"label"`
    Placeholder string `json:"placeholder"`
    HelpText    string `json:"helpText,omitempty"`
    Required    bool   `json:"required"`
    Disabled    bool   `json:"disabled,omitempty"`
}

formRegistry := gozod.NewRegistry[FormFieldMetadata]()

// Define form schema with field metadata
contactForm := gozod.Struct(gozod.ObjectSchema{
    "name": gozod.String().Min(1).Register(formRegistry, FormFieldMetadata{
        Label:       "Full Name",
        Placeholder: "Enter your full name",
        Required:    true,
    }),
    "email": gozod.Email().Register(formRegistry, FormFieldMetadata{
        Label:       "Email Address", 
        Placeholder: "you@example.com",
        HelpText:    "We'll never share your email",
        Required:    true,
    }),
    "phone": gozod.String().Optional().Register(formRegistry, FormFieldMetadata{
        Label:       "Phone Number",
        Placeholder: "+1 (555) 123-4567",
        Required:    false,
    }),
})
```

### Schema References and ID Management

For managing schema relationships and references:

```go
// Schema reference metadata
type SchemaRefMetadata struct {
    ID          string   `json:"id"`
    Version     string   `json:"version,omitempty"`
    References  []string `json:"references,omitempty"`
    Namespace   string   `json:"namespace,omitempty"`
}

refRegistry := gozod.NewRegistry[SchemaRefMetadata]()

// Define reusable schemas with IDs
userSchema := gozod.Struct(gozod.ObjectSchema{
    "id":    gozod.String().UUID(),
    "name":  gozod.String(),
    "email": gozod.Email(),
}).Register(refRegistry, SchemaRefMetadata{
    ID:        "User",
    Version:   "1.0",
    Namespace: "api.v1",
})

postSchema := gozod.Struct(gozod.ObjectSchema{
    "id":       gozod.String().UUID(),
    "title":    gozod.String(),
    "content":  gozod.String(),
    "authorId": gozod.String().UUID(),
}).Register(refRegistry, SchemaRefMetadata{
    ID:         "Post",
    Version:    "1.0", 
    Namespace:  "api.v1",
    References: []string{"User"},
})
```

## Registry Operations

### Iterating Over Registries

```go
// Iterate over all schemas in a registry
for schema, metadata := range myRegistry.All() {
    fmt.Printf("Schema: %v, Metadata: %v\n", schema, metadata)
}

// Get all schemas (without metadata)
schemas := myRegistry.Schemas()

// Get all metadata (without schemas)
metadataList := myRegistry.Metadata()

// Get registry size
count := myRegistry.Count()
```

### Registry Merging and Cloning

```go
// Clone a registry
clonedRegistry := myRegistry.Clone()

// Merge registries
targetRegistry := gozod.NewRegistry[MyMetadata]()
sourceRegistry := gozod.NewRegistry[MyMetadata]()

// Merge source into target
targetRegistry.Merge(sourceRegistry)

// Create new registry from merge
mergedRegistry := gozod.MergeRegistries(targetRegistry, sourceRegistry)
```

### Registry Filtering

```go
// Filter registry by metadata criteria
filteredRegistry := myRegistry.Filter(func(metadata MyMetadata) bool {
    return strings.Contains(metadata.Description, "important")
})

// Find schemas by metadata
importantSchemas := myRegistry.FindByMetadata(func(metadata MyMetadata) bool {
    return metadata.Description == "important schema"
})
```

## Integration with JSON Schema

Registries integrate seamlessly with GoZod's JSON Schema generation:

```go
// Create schemas with metadata
userSchema := gozod.Struct(gozod.ObjectSchema{
    "name": gozod.String().Meta(gozod.GlobalMetadata{
        Title:       "User Name",
        Description: "The user's full name",
        Examples:    []interface{}{"John Doe"},
    }),
    "age": gozod.Int().Min(18).Meta(gozod.GlobalMetadata{
        Title:       "User Age", 
        Description: "The user's age (must be 18 or older)",
        Examples:    []interface{}{25},
    }),
})

// Convert to JSON Schema with metadata preserved
jsonSchema := gozod.ToJSONSchema(userSchema)

// The resulting JSON Schema will include all metadata:
// {
//   "type": "object",
//   "properties": {
//     "name": {
//       "type": "string",
//       "title": "User Name",
//       "description": "The user's full name", 
//       "examples": ["John Doe"]
//     },
//     "age": {
//       "type": "integer",
//       "minimum": 18,
//       "title": "User Age",
//       "description": "The user's age (must be 18 or older)",
//       "examples": [25]
//     }
//   },
//   "required": ["name", "age"]
// }
```

## Best Practices

1. **Use typed registries**: Define specific metadata structures for different use cases rather than using `interface{}`.

2. **Namespace your IDs**: When using schema IDs, use namespacing to avoid conflicts (`"api.v1.User"` vs `"User"`).

3. **Document your metadata**: Include comments and documentation for your metadata structures.

4. **Separate concerns**: Use different registries for different purposes (API docs, forms, JSON Schema, etc.).

5. **Version your schemas**: Include version information in metadata for schema evolution.

6. **Validate examples**: Ensure that examples in metadata actually validate against their schemas.

```go
// Good: Typed, namespaced, versioned
type APISchemaMetadata struct {
    ID          string `json:"id"`          // "api.v2.User"
    Version     string `json:"version"`     // "2.1.0"
    Description string `json:"description"` // Clear description
    Examples    []User `json:"examples"`    // Type-safe examples
}

// Good: Separate registries for different concerns
var (
    APIRegistry  = gozod.NewRegistry[APISchemaMetadata]()
    FormRegistry = gozod.NewRegistry[FormFieldMetadata]()  
    DocsRegistry = gozod.NewRegistry[DocumentationMetadata]()
)
```

By leveraging GoZod's metadata and registry system, you can create rich, self-documenting schemas that integrate seamlessly with JSON Schema generation, form rendering, API documentation, and other tooling. 
