# GoZod 🔷

**GoZod** is a TypeScript Zod v4-inspired validation library for Go, providing strongly-typed, zero-dependency data validation with intelligent type inference and maximum performance.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Test Status](https://img.shields.io/badge/Tests-Passing-green.svg)](https://github.com/kaptinlin/gozod)

## ✨ Key Features

- **TypeScript Zod v4 Compatible API** - Familiar syntax with Go-native optimizations
- **Complete Strict Type Semantics** - All methods require exact input types, zero automatic conversions
- **🏷️ Declarative Struct Tags** - Define validation rules directly on struct fields with `gozod:"required,min=2,email"`
- **Parse vs StrictParse** - Runtime flexibility or compile-time type safety for optimal performance
- **Native Go Struct Support** - First-class struct validation with field-level validation and JSON tag mapping
- **Automatic Circular Reference Handling** - Lazy evaluation prevents stack overflow in recursive structures
- **Maximum Performance** - Zero-overhead validation with optional code generation (5-10x faster)
- **Zero Dependencies** - Pure Go implementation, no external libraries
- **Rich Validation Methods** - Comprehensive built-in validators for all Go types

## Why GoZod?

- **🎯 Type Safety First** - Compile-time guarantees with runtime flexibility
- **⚡ Maximum Performance** - Zero-overhead abstractions with optional code generation
- **🏷️ Developer Experience** - Familiar API with Go idioms and declarative struct tags
- **🔒 Production Ready** - Battle-tested validation with comprehensive error handling
- **🌟 Zero Dependencies** - Pure Go implementation, no external dependencies

GoZod brings TypeScript Zod's excellent developer experience to Go while embracing Go's type system and performance characteristics. Perfect for API validation, configuration parsing, data transformation, and any scenario where type-safe validation is critical.

## 📦 Quick Start

### Installation

```bash
go get github.com/kaptinlin/gozod
```

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/gozod"
)

func main() {
    // String validation with method chaining
    nameSchema := gozod.String().Min(2).Max(50)
    
    // Parse - Runtime type checking (flexible)
    result, err := nameSchema.Parse("Alice")
    if err == nil {
        fmt.Println("Valid name:", result) // "Alice"
    }

    // StrictParse - Compile-time type safety (optimal performance)
    name := "Alice"
    result, err = nameSchema.StrictParse(name) // Input type guaranteed
    if err == nil {
        fmt.Println("Validated name:", result)
    }

    // Email validation
    emailSchema := gozod.String().Email()
    email := "user@example.com"
    result, err = emailSchema.StrictParse(email)
    if err == nil {
        fmt.Printf("Valid email: %s\n", result)
    }
}
```

### Struct Tag Validation (Declarative)

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/gozod"
)

type User struct {
    Name  string `gozod:"required,min=2,max=50"`
    Email string `gozod:"required,email"`
    Age   int    `gozod:"required,min=18,max=120"`
    Bio   string `gozod:"max=500"`           // Optional by default
}

func main() {
    // Generate schema from struct tags
    schema := gozod.FromStruct[User]()
    
    user := User{
        Name:  "Alice Johnson",
        Email: "alice@example.com",
        Age:   28,
        Bio:   "Software engineer",
    }
    
    // Validate with generated schema
    validatedUser, err := schema.Parse(user)
    if err != nil {
        fmt.Printf("Validation error: %v\n", err)
        return
    }
    
    fmt.Printf("Valid user: %+v\n", validatedUser)
}
```

### Programmatic Struct Validation

```go
package main

import (
    "fmt" 
    "github.com/kaptinlin/gozod"
)

type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

func main() {
    // Basic struct validation
    basicSchema := gozod.Struct[User]()
    user := User{Name: "Alice", Age: 25, Email: "alice@example.com"}
    result, err := basicSchema.Parse(user)
    if err == nil {
        fmt.Printf("Basic validation: %+v\n", result)
    }

    // Struct with field validation
    userSchema := gozod.Struct[User](gozod.StructSchema{
        "name":  gozod.String().Min(2).Max(50),
        "age":   gozod.Int().Min(0).Max(120),
        "email": gozod.String().Email(),
    })

    validUser := User{Name: "Bob", Age: 30, Email: "bob@example.com"}
    result, err = userSchema.Parse(validUser)
    if err == nil {
        fmt.Printf("Field validation: %+v\n", result)
    }
}
```

## 🚀 Advanced Features

### Complete Strict Type Semantics

GoZod uses strict type semantics - no automatic conversions between types:

```go
// Value schemas require exact value types
stringSchema := gozod.String()
result, _ := stringSchema.Parse("hello")     // ✅ string → string
// result, _ := stringSchema.Parse(&str)     // ❌ Error: requires string, got *string

// Pointer schemas require exact pointer types
stringPtrSchema := gozod.StringPtr()
result, _ = stringPtrSchema.Parse(&str)      // ✅ *string → *string
// result, _ = stringPtrSchema.Parse("hello") // ❌ Error: requires *string, got string

// For flexible input handling, use Optional/Nilable modifiers
optionalSchema := gozod.String().Optional()  // Flexible input, *string output
result, _ = optionalSchema.Parse("hello")    // ✅ string → *string (new pointer)
result, _ = optionalSchema.Parse(&str)       // ✅ *string → *string (preserves identity)
```

### Parse vs StrictParse Methods

Choose the right parsing method for your use case:

```go
schema := gozod.String().Min(3)

// Parse - Runtime type checking (flexible input)
result, err := schema.Parse("hello")        // ✅ Works with any input type
result, err = schema.Parse(42)              // ❌ Runtime error: invalid type

// StrictParse - Compile-time type safety (optimal performance)
str := "hello"
result, err = schema.StrictParse(str)       // ✅ Compile-time guarantee, optimal performance
// result, err := schema.StrictParse(42)    // ❌ Compile-time error
```

### Custom Validation with Refine

Use `.Refine()` for custom validation logic:

```go
package main

import (
    "strings"
    "github.com/kaptinlin/gozod"
)

func main() {
    // Custom validation with Refine
    usernameSchema := gozod.String().
        Min(3).
        Max(20).
        Refine(func(username string) bool {
            // Check against blacklist
            blacklist := map[string]bool{"admin": true, "root": true}
            return !blacklist[strings.ToLower(username)]
        }, "Username is not allowed")
    
    // Valid username
    result, err := usernameSchema.Parse("johndoe")  // ✅ Valid
    
    // Invalid username  
    _, err = usernameSchema.Parse("admin")  // ❌ Validation fails
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
    }
    
    // Struct validation with custom logic
    type User struct {
        Name  string `gozod:"required,min=2"`
        Email string `gozod:"required,email"`
        Age   int    `gozod:"min=18"`
    }
    
    schema := gozod.FromStruct[User]().Refine(func(user User) bool {
        // Cross-field validation
        return user.Age >= 21 || !strings.Contains(user.Email, "company.com")
    }, "Users under 21 cannot have company emails")
}

### Automatic Circular Reference Handling

GoZod automatically detects and handles circular references:

```go
type User struct {
    Name    string  `gozod:"required,min=2"`
    Email   string  `gozod:"required,email"`
    Friends []*User `gozod:"max=10"`       // Circular reference
}

// No stack overflow - automatically uses lazy evaluation
schema := gozod.FromStruct[User]()

alice := &User{Name: "Alice", Email: "alice@example.com"}
bob := &User{Name: "Bob", Email: "bob@example.com", Friends: []*User{alice}}
alice.Friends = []*User{bob} // Circular reference

result, err := schema.Parse(*alice) // ✅ Handles circular reference safely
```

### Union, Xor, and Intersection Types

```go
// Union types - accepts one of multiple schemas (any match succeeds)
unionSchema := gozod.Union(
    gozod.String(),
    gozod.Int(),
)
result, _ := unionSchema.Parse("hello") // ✅ Matches string
result, _ = unionSchema.Parse(42)       // ✅ Matches int
result, _ = unionSchema.Parse(true)     // ❌ No union member matched

// Xor types - exactly one schema must match (exclusive union)
xorSchema := gozod.Xor([]any{
    gozod.Email(),  // Email format validator
    gozod.URL(),    // URL format validator
})
result, _ = xorSchema.Parse("user@example.com")  // ✅ Matches email only
result, _ = xorSchema.Parse("https://site.com")  // ✅ Matches URL only
result, _ = xorSchema.Parse("invalid")           // ❌ Matches neither

// Intersection types - must satisfy all schemas
intersectionSchema := gozod.Intersection(
    gozod.String().Min(3),              // At least 3 chars
    gozod.String().Max(10),             // At most 10 chars
    gozod.String().RegexString(`^[a-z]+$`), // Only lowercase
)
result, _ = intersectionSchema.Parse("hello")  // ✅ Satisfies all constraints
result, _ = intersectionSchema.Parse("HELLO")  // ❌ Not lowercase

// And/Or methods - fluent composition on any schema type
schema := gozod.String().Min(3).And(gozod.String().Max(10))  // Intersection via method
schema = gozod.Int().Or(gozod.String())                       // Union via method
```

### Performance Optimization with Code Generation

For maximum performance, use code generation:

```go
//go:generate gozodgen

type User struct {
    Name  string `gozod:"required,min=2"`
    Email string `gozod:"required,email"`
    Age   int    `gozod:"required,min=18"`
}

// Generated Schema() method provides zero-reflection validation
func main() {
    schema := gozod.FromStruct[User]() // Uses generated code automatically
    
    user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
    result, err := schema.Parse(user)  // 5-10x faster than reflection
}
```

## 🎛️ Comprehensive Type Support

### Primitive Types

```go
// Strings with format validation
gozod.String().Min(3).Max(100).Email()
gozod.String().RegexString(`^\d+$`)
gozod.Uuid()     // UUID format validator
gozod.URL()      // URL format validator
gozod.HttpURL()  // HTTP/HTTPS URL only
gozod.Email()    // Email format validator

// Network formats
gozod.IPv4()     // IPv4: "192.168.1.1"
gozod.IPv6()     // IPv6: "2001:db8::8a2e:370:7334"
gozod.Hostname() // DNS hostname: "example.com"
gozod.MAC()      // MAC address: "00:1A:2B:3C:4D:5E"
gozod.E164()     // E.164 phone: "+14155552671"
gozod.CIDRv4()   // IPv4 CIDR: "192.168.1.0/24"

// Text encodings and hashes
gozod.Hex()      // Hexadecimal string
gozod.Base64()   // Base64 encoding
gozod.JWT()      // JWT token format

// Numbers with range validation  
gozod.Int().Min(0).Max(120).Positive()
gozod.Float64().Min(0.0).Finite()

// Booleans
gozod.Bool()

// Time validation
gozod.Time().After(startDate).Before(endDate)
```

### Complex Types

```go
// Arrays and Slices
gozod.Array(gozod.String()).Min(1).Max(10)
gozod.Slice(gozod.Int()).NonEmpty()

// Tuples - Fixed-length arrays with typed elements
tuple := gozod.Tuple(gozod.String(), gozod.Int(), gozod.Bool())
result, _ := tuple.Parse([]any{"hello", 42, true})

// Tuple with rest element for variadic trailing elements
tupleWithRest := gozod.TupleWithRest(
    []core.ZodSchema{gozod.String(), gozod.Int()},
    gozod.Bool(), // additional elements must be booleans
)

// Maps
gozod.Map(gozod.String()).NonEmpty() // map[string]string, at least one entry
gozod.Map(gozod.Struct[User]()) // map[string]User

// Records with key validation
gozod.Record(gozod.String().Regex(`^[a-z]+$`), gozod.Int()) // lowercase keys only

// LooseRecord - passes through non-matching keys unchanged
gozod.LooseRecord(gozod.String().Regex(`^S_`), gozod.String())

// Objects (map[string]any)
gozod.Object(gozod.ObjectSchema{
    "name": gozod.String().Min(2),
    "age":  gozod.Int().Min(0),
})
```

### Advanced Types

```go
// Transform types
stringToInt := gozod.String().Regex(`^\d+$`).Transform(
    func(s string, ctx *core.RefinementContext) (any, error) {
        return strconv.Atoi(s)
    },
)

// Lazy types for recursive structures
var nodeSchema gozod.ZodType[Node]
nodeSchema = gozod.Lazy(func() gozod.ZodType[Node] {
    return gozod.Struct[Node](gozod.StructSchema{
        "value":    gozod.String(),
        "children": gozod.Array(nodeSchema), // Self-reference
    })
})

// Schema metadata
schema := gozod.String().Email().Describe("User's primary email")
schema = gozod.Int().Meta(gozod.GlobalMeta{
    Title:       "Age",
    Description: "User's age in years",
})

// Apply - Compose reusable schema modifiers
func addCommonChecks[T types.StringConstraint](s *gozod.ZodString[T]) *gozod.ZodString[T] {
    return s.Min(1).Max(100).Trim()
}
schema := gozod.Apply(gozod.String(), addCommonChecks)
```

## 🔧 Error Handling

Comprehensive error information with structured details:

```go
schema := gozod.String().Min(5).Email()
_, err := schema.Parse("hi")

if zodErr, ok := err.(*issues.ZodError); ok {
    // Access structured error information
    for _, issue := range zodErr.Issues {
        fmt.Printf("Path: %v, Code: %s, Message: %s\n", 
            issue.Path, issue.Code, issue.Message)
    }
    
    // Pretty print errors
    fmt.Println(zodErr.PrettifyError())
    
    // Get flattened field errors for forms
    fieldErrors := zodErr.FlattenError()
    for field, errors := range fieldErrors.FieldErrors {
        fmt.Printf("%s: %v\n", field, errors)
    }
}
```

## 🏷️ Complete Tag Reference

### String Tags
- `required` - Field must be present
- `min=N` / `max=N` - Length constraints
- `email` / `url` / `uuid` - Format validation
- `regex=pattern` - Custom regex patterns

### Numeric Tags  
- `min=N` / `max=N` - Value constraints
- `positive` / `negative` - Sign validation
- `nonnegative` / `nonpositive` - Zero-inclusive constraints

### Array Tags
- `min=N` / `max=N` - Element count constraints
- `nonempty` - At least one element
- `length=N` - Exact element count

### Custom Validation
Use `.Refine()` for custom validation logic on any schema:

```go
schema := gozod.FromStruct[Product]().Refine(func(p Product) bool {
    return strings.HasPrefix(p.SKU, "PROD-")
}, "SKU must start with PROD-")
```

## 🌟 Real-World Examples

### API Request Validation

```go
type CreateUserRequest struct {
    Name     string   `json:"name" gozod:"required,min=2,max=50"`
    Email    string   `json:"email" gozod:"required,email"`
    Age      int      `json:"age" gozod:"required,min=18,max=120"`
    Tags     []string `json:"tags" gozod:"max=10"`
    Website  string   `json:"website" gozod:"url"`
    IsActive bool     `json:"is_active"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    schema := gozod.FromStruct[CreateUserRequest]()
    validatedReq, err := schema.Parse(req)
    if err != nil {
        writeValidationError(w, err)
        return
    }
    
    user := createUser(validatedReq)
    json.NewEncoder(w).Encode(user)
}
```

### Configuration Validation

```go
type Config struct {
    Environment string `yaml:"environment" gozod:"required,regex=^(dev|staging|prod)$"`
    Port        int    `yaml:"port" gozod:"required,min=1000,max=9999"`
    Database    struct {
        Host     string `yaml:"host" gozod:"required"`
        Port     int    `yaml:"port" gozod:"required,min=1,max=65535"`
        Name     string `yaml:"name" gozod:"required,min=1"`
        Username string `yaml:"username" gozod:"required"`
        Password string `yaml:"password" gozod:"required,min=8"`
    } `yaml:"database" gozod:"required"`
    Debug bool `yaml:"debug"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    schema := gozod.FromStruct[Config]()
    return schema.Parse(config)
}
```

## 📚 Documentation

- **[API Reference](docs/api.md)** - Complete API documentation with all methods and examples
- **[Struct Tags Guide](docs/tags.md)** - Comprehensive tag syntax and custom validator integration
- **[Feature Mapping](docs/feature-mapping.md)** - Complete TypeScript Zod v4 to GoZod mapping reference
- **[Basics Guide](docs/basics.md)** - Getting started with core concepts and patterns
- **[Error Customization](docs/error-customization.md)** - Custom error messages and internationalization
- **[Error Formatting](docs/error-formatting.md)** - Structured error handling and display
- **[JSON Schema](docs/json-schema.md)** - Generate JSON Schema from GoZod schemas

### OpenAPI 3.1 Support  ✅

GoZod schemas generate JSON Schema Draft 2020-12, which is **fully compatible with OpenAPI 3.1**. Features like nullable types (`["string", "null"]`), numeric exclusive bounds, conditional schemas (`if`/`then`/`else`), and tuple validation work out of the box. See [JSON Schema docs](docs/json-schema.md) for details.

> **Note**: OpenAPI 3.0 is not supported (use OpenAPI 3.1 instead).

- **[Metadata](docs/metadata.md)** - Schema metadata and introspection capabilities
- **[Examples](examples/)** - Working examples for common use cases

## 🚀 Performance

GoZod is designed for maximum performance:

- **Zero Dependencies** - Pure Go implementation
- **Strict Type Semantics** - No runtime type conversions
- **StrictParse Method** - Compile-time type safety eliminates runtime checks
- **Code Generation** - Optional zero-reflection validation (5-10x faster)
- **Efficient Validation Pipeline** - Optimized execution paths

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---
