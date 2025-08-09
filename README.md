# GoZod 🔷

**GoZod** is a TypeScript Zod-inspired validation library for Go, providing strongly-typed, zero-dependency data validation with intelligent type inference.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Test Status](https://img.shields.io/badge/Tests-Passing-green.svg)](https://github.com/kaptinlin/gozod)

## ✨ Key Features

- **TypeScript Zod v4 Compatible API** - Familiar syntax with Go-native optimizations
- **Type-Safe Validation** - All methods require exact input types, zero automatic conversions
- **Native Go Struct Support** - First-class struct validation with field-level validation and JSON tag mapping
- **Maximum Performance** - Zero-overhead validation with optimal execution paths
- **Zero Dependencies** - Pure Go implementation, no external libraries
- **Rich Validation Methods** - Comprehensive built-in validators for all Go types
- **Type-Safe Method Chaining** - Fluent API with compile-time type safety

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
    // String validation with chaining
    nameSchema := gozod.String().Min(2).Max(50)
    result, err := nameSchema.Parse("Alice")
    if err == nil {
        fmt.Println("Valid name:", result) // "Alice"
    }

    // Compile-time type safety with StrictParse
    name := "Alice"
    result, err = nameSchema.StrictParse(name) // Input type guaranteed at compile-time
    if err == nil {
        fmt.Println("Validated name:", result)
    }

    // Email validation
    emailSchema := gozod.String().Email()
    result, err = emailSchema.Parse("user@example.com")
    // result: "user@example.com", err: nil

    // StrictParse for known string input
    email := "user@example.com"
    result, err = emailSchema.StrictParse(email)
    // result: "user@example.com", err: nil
}
```

### Struct Schema Validation

```go
// Define Go struct
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

// Basic struct validation
basicSchema := gozod.Struct[User]()
user := User{Name: "Alice", Age: 25, Email: "alice@example.com"}
result, err := basicSchema.Parse(user)
if err == nil {
    fmt.Printf("Valid user: %+v\n", result)
}

// Struct with field validation
userSchema := gozod.Struct[User](gozod.StructSchema{
    "name":  gozod.String().Min(2).Max(50),
    "age":   gozod.Int().Min(0).Max(120),
    "email": gozod.String().Email().Optional(),
})

validUser := User{Name: "Bob", Age: 30, Email: "bob@example.com"}
result, err = userSchema.Parse(validUser)
if err == nil {
    fmt.Printf("Validated struct: %+v\n", result)
}
```

### Object Schema Validation

```go
// For JSON-like data (map[string]any)
userObjectSchema := gozod.Object(gozod.ObjectSchema{
    "name":  gozod.String().Min(2).Max(50),
    "age":   gozod.Int().Min(0).Max(120),
    "email": gozod.String().Email().Optional(),
})

// Validate JSON-like data
userData := map[string]any{
    "name": "Alice",
    "age":  25,
    "email": "alice@example.com",
}

result, err := userObjectSchema.Parse(userData)
if err != nil {
    fmt.Printf("Validation failed: %v\n", err)
    return
}

fmt.Printf("Valid user data: %+v\n", result)
```

### Advanced Struct Usage

```go
// Struct with pointer fields and validation
type Profile struct {
    ID       int     `json:"id"`
    Username string  `json:"username"`
    Email    *string `json:"email,omitempty"` // Optional field
    Active   bool    `json:"active"`
}

// Schema with field-level validation
profileSchema := gozod.Struct[Profile](gozod.StructSchema{
    "id":       gozod.Int().Min(1),
    "username": gozod.String().Min(3).Max(20),
    "email":    gozod.String().Email().Optional(), // Handles pointer fields
    "active":   gozod.Bool(),
})

// Validate struct with optional field
email := "john@example.com"
profile := Profile{
    ID:       123,
    Username: "john_doe",
    Email:    &email, // Pointer to string
    Active:   true,
}

result, err := profileSchema.Parse(profile)
if err == nil {
    fmt.Printf("Valid profile: %+v\n", result)
}

// Struct modifiers
optionalSchema := gozod.Struct[Profile]().Optional()    // Returns *Profile
defaultProfile := Profile{ID: 1, Username: "guest", Active: false}
withDefault := gozod.Struct[Profile]().Default(defaultProfile)

// Partial validation (skip zero values)
partialSchema := profileSchema.Partial() // Makes all fields optional
partialProfile := Profile{Username: "partial_user"} // Only username set
result, err = partialSchema.Parse(partialProfile) // Valid - other fields skipped
```

### Type Coercion

```go
import "github.com/kaptinlin/gozod/coerce"

// Automatic type conversion
stringSchema := coerce.String()
result, _ := stringSchema.Parse(123)    // "123"

numberSchema := coerce.Number()
result, _ = numberSchema.Parse("42")    // 42.0

// Use Coerce constructors
coerceSchema := coerce.Int()
result, _ = coerceSchema.Parse("25")    // 25
```


### Error Handling

```go
schema := gozod.String().Min(5).Email()
_, err := schema.Parse("hi")

if err != nil {
    var zodErr *gozod.ZodError
    if gozod.IsZodError(err, &zodErr) {
        for _, issue := range zodErr.Issues {
            fmt.Printf("Error: %s at %v\n", issue.Message, issue.Path)
        }
    }
}
```

## 🔧 Core Concepts

### Validation Methods
```go
// String validation
gozod.String().Min(3).Max(100).Email().StartsWith("user")

// Number validation  
gozod.Int().Min(0).Max(120).Positive()

// Struct validation
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
gozod.Struct[Person](gozod.StructSchema{
    "name": gozod.String().Min(2),
    "age":  gozod.Int().Min(0).Max(120),
})

// Array validation
gozod.Slice[string](gozod.String()).Min(1).Max(10).NonEmpty()
```

### Type Safety and Modifiers
```go
// Value schemas: strict value input only
gozod.String().Email()       // Requires string input, returns string
gozod.Int().Min(0)          // Requires int input, returns int

// Pointer schemas: strict pointer input only  
gozod.StringPtr().Email()   // Requires *string input, returns *string
gozod.IntPtr().Min(0)       // Requires *int input, returns *int

// Optional/Nilable: flexible input, pointer output
gozod.String().Optional()   // Flexible input (string/*string), returns *string
gozod.String().Nilable()    // Flexible input, returns *string or typed nil

// Default values and prefaults
gozod.String().Default("anonymous")  // Short-circuits on nil input
gozod.String().Transform(strings.ToUpper).Prefault("fallback")  // Pre-parse default for nil
```

### Type Safety Requirements
```go
// All schemas require exact input types - no automatic conversions
stringSchema := gozod.String()     // Requires string, returns string  
stringPtrSchema := gozod.StringPtr() // Requires *string, returns *string

value := "hello"

// Value schema: only accepts values
result1, _ := stringSchema.Parse("hello")    // ✅ string → string
// result1, _ := stringSchema.Parse(&value)  // ❌ Error: requires string, got *string

// Pointer schema: only accepts pointers
result2, _ := stringPtrSchema.Parse(&value)  // ✅ *string → *string (preserves identity)
// result2, _ := stringPtrSchema.Parse("hello") // ❌ Error: requires *string, got string

// For flexible input handling, use Optional/Nilable modifiers
optionalSchema := gozod.String().Optional()  // Returns *string, flexible input
result3, _ := optionalSchema.Parse("hello")  // ✅ string → *string (new pointer)
result4, _ := optionalSchema.Parse(&value)   // ✅ *string → *string (preserves identity)

// Helper functions for strict mode convenience
func Ptr[T any](v T) *T { return &v }
func Deref[T any](ptr *T) T { return *ptr }

// Usage with helpers
result5, _ := stringPtrSchema.Parse(Ptr("hello"))   // Create pointer
result6, _ := stringSchema.Parse(Deref(&value))     // Dereference pointer
```

### Advanced Types
```go
// Struct with nested validation
type Address struct {
    Street string `json:"street"`
    City   string `json:"city"`
}

type Customer struct {
    Name    string  `json:"name"`
    Address Address `json:"address"`
    Tags    []string `json:"tags"`
}

customerSchema := gozod.Struct[Customer](gozod.StructSchema{
    "name": gozod.String().Min(2),
    "address": gozod.Struct[Address](gozod.StructSchema{
        "street": gozod.String().Min(5),
        "city":   gozod.String().Min(2),
    }),
    "tags": gozod.Slice[string](gozod.String()).Optional(),
})

// Enum types
colorEnum := gozod.Enum("red", "green", "blue")
statusMap := gozod.EnumMap(map[string]string{
	"ACTIVE": "active", 
	"INACTIVE": "inactive"
})

// Go native enum support
type Status int
const (
    Active Status = iota
    Inactive
)
statusEnum := gozod.Enum(Active, Inactive)

// Union types (OR logic)
gozod.Union(gozod.String(), gozod.Int())

// Discriminated unions (efficient lookup)
gozod.DiscriminatedUnion("type", schema1, schema2, schema3)

// Recursive types
var TreeNode = gozod.LazyAny(func() any {
    return gozod.Object(gozod.ObjectSchema{
        "value":    gozod.String(),
        "children": gozod.Slice[any](TreeNode),
    })
})

```

## 📚 Documentation

- **[Getting Started Guide](docs/basics.md)** - Fundamental usage patterns and core concepts
- **[Complete API Reference](docs/api.md)** - Comprehensive type interface documentation
- **[Error Handling & Customization](docs/error-customization.md)** - Custom error messages and internationalization
- **[Error Formatting](docs/error-formatting.md)** - Different error output formats for various use cases
- **[JSON Schema Integration](docs/json-schema.md)** - Generate and work with JSON schemas
- **[Metadata System](docs/metadata.md)** - Attach custom metadata to schemas
- **[TypeScript Zod v4 Feature Mapping](docs/feature-mapping.md)** - Complete compatibility matrix and migration guide

## 🔗 TypeScript Zod v4 Compatibility

GoZod provides comprehensive compatibility with TypeScript Zod v4 while adding Go-specific enhancements:

### Core Type Support
- **Basic Types**: `string`, `number`, `boolean`, `bigint` with Go-native type variants
- **Collections**: `array`, `object`, `record`, `map` with smart type inference
- **Advanced Types**: `union`, `intersection`, `discriminated union`, `literal`, `enum`
- **Modifiers**: `optional`, `nilable`, `default` with Go semantics

### Key Enhancements
- **Type-Safe Validation**: All methods require exact input types with zero automatic conversions
- **Native Go Struct Validation**: Type-safe struct validation with field schemas, JSON tag mapping, and partial validation support
- **Maximum Performance**: Optimized validation paths with minimal overhead and memory allocations
- **Go-Specific Types**: Support for all Go numeric types (`int8`-`int64`, `uint8`-`uint64`, `float32/64`, `complex64/128`)
- **Smart Nil Handling**: Distinction between "missing field" (Optional) and "null value" (Nilable) semantics with simplified zero-value returns
- **Pointer Identity Preservation**: Optional/Nilable modifiers maintain input pointer addresses when appropriate
- **Enhanced Error System**: Structured error handling with custom formatting and internationalization

### Compatibility Status
✅ **Fully Compatible**: All major TypeScript Zod v4 features implemented  
✅ **Go-Enhanced**: Additional features leveraging Go's type system  
✅ **Performance Optimized**: Efficient validation algorithms and discriminated unions

For detailed feature mapping, migration guide, and compatibility matrix, see [Feature Mapping Documentation](docs/feature-mapping.md).

## 🤝 Contributing

Contributions welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

This project is a Go port inspired by the excellent [TypeScript Zod](https://github.com/colinhacks/zod) implementation. We have adapted the core API design and added Go-specific optimizations while maintaining full compatibility with TypeScript Zod v4.

Special thanks to the original Zod project for providing a solid foundation and comprehensive test cases, which enabled this high-quality Go implementation.
