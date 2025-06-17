# GoZod 🔷

**GoZod** is a TypeScript Zod-inspired validation library for Go, providing strongly-typed, zero-dependency data validation with intelligent type inference.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Test Status](https://img.shields.io/badge/Tests-Passing-green.svg)](https://github.com/kaptinlin/gozod)

## ✨ Key Features

- **TypeScript Zod v4 Compatible API** - Familiar syntax with Go-native optimizations
- **Intelligent Type Inference** - Input types preserved in output with smart pointer handling
- **Zero Dependencies** - Pure Go implementation, no external libraries
- **Optimized Performance** - Efficient discriminated unions, optimized validation algorithms
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

    // Email validation
    emailSchema := gozod.String().Email()
    result, err = emailSchema.Parse("user@example.com")
    // result: "user@example.com", err: nil
}
```

### Object Schema Validation

```go
// Define schema for structured data
userSchema := gozod.Object(gozod.ObjectSchema{
    "name":  gozod.String().Min(2).Max(50),
    "age":   gozod.Int().Min(0).Max(120),
    "email": gozod.String().Email().Optional(),
})

// Validate JSON-like data
userData := map[string]interface{}{
    "name": "Alice",
    "age":  25,
    "email": "alice@example.com",
}

result, err := userSchema.Parse(userData)
if err != nil {
    fmt.Printf("Validation failed: %v\n", err)
    return
}

fmt.Printf("Valid user: %+v\n", result)
```

### Type Coercion

```go
// Automatic type conversion
stringSchema := gozod.Coerce.String()
result, _ := stringSchema.Parse(123)    // "123"

numberSchema := gozod.Coerce.Number()
result, _ = numberSchema.Parse("42")    // 42.0

// Schema-level coercion
coerceSchema := gozod.Int(gozod.SchemaParams{Coerce: true})
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

// Array validation
gozod.Slice(gozod.String()).Min(1).Max(10).NonEmpty()
```

### Modifiers and Wrappers
```go
// Optional (allows nil/missing)
gozod.String().Email().Optional()

// Nilable (returns typed nil pointer)
gozod.String().Nilable()  // Returns (*string)(nil) for nil input

// Default values
gozod.String().Default("anonymous")

// Fallback on validation failure
gozod.String().Min(5).Prefault("fallback")
```

### Advanced Types
```go
// Union types (OR logic)
gozod.Union([]gozod.ZodType[any, any]{gozod.String(), gozod.Int()})

// Discriminated unions (efficient lookup)
gozod.DiscriminatedUnion("type", schemas)

// Recursive types
var TreeNode gozod.ZodType[any, any]
TreeNode = gozod.Lazy(func() gozod.ZodType[any, any] {
    return gozod.Object(gozod.ObjectSchema{
        "value":    gozod.String(),
        "children": gozod.Slice(TreeNode),
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
- **Modifiers**: `optional`, `nullable`, `default`, `catch` with Go semantics

### Key Enhancements
- **Pointer Identity Preservation**: Input pointer addresses maintained in output
- **Go-Specific Types**: Support for all Go numeric types (`int8`-`int64`, `uint8`-`uint64`, `float32/64`, `complex64/128`)
- **Smart Nil Handling**: Distinction between "missing field" (Optional) and "null value" (Nilable) semantics
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
