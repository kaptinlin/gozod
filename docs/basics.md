# GoZod Basics

This guide covers GoZod fundamentals including schema creation, data validation, error handling, and type inference. For complete type reference and method documentation, see the [API Reference](api.md).

## Table of Contents

- [Defining a Schema](#defining-a-schema)
- [Parsing Data](#parsing-data)
- [Handling Errors](#handling-errors)
- [Type Inference](#type-inference)
- [Common Patterns](#common-patterns)
- [Working with JSON APIs](#working-with-json-apis)
- [Configuration Validation](#configuration-validation)

## Defining a Schema

Before you can validate data, you need to define a schema. Here's a simple example:

```go
import "github.com/kaptinlin/gozod"

// Define a user schema
var UserSchema = gozod.Object(gozod.ObjectSchema{
    "username": gozod.String().Min(3).Max(20),
    "email":    gozod.String().Email(),
    "age":      gozod.Int().Min(0).Max(120),
    "active":   gozod.Bool().Optional(),
})
```

Schemas can be composed and reused:

```go
// Base schemas
var EmailSchema = gozod.String().Email()
var PhoneSchema = gozod.String().Regex(regexp.MustCompile(`^\+?[1-9]\d{1,14}$`))

// Composed schema
var ContactSchema = gozod.Object(gozod.ObjectSchema{
    "email": EmailSchema,
    "phone": PhoneSchema.Optional(),
})
```

## Parsing Data

GoZod provides two methods for parsing data:

### `.Parse()` - Standard Error Handling

Use `.Parse()` for standard Go error handling patterns:

```go
userData := map[string]any{
    "username": "johndoe",
    "email":    "john@example.com",
    "age":      25,
    "active":   true,
}

result, err := UserSchema.Parse(userData)
if err != nil {
    // Handle validation error
    log.Printf("Validation failed: %v", err)
    return
}

// Use validated data
validatedUser := result.(map[string]any)
fmt.Printf("Valid user: %s\n", validatedUser["username"])
```

### `.MustParse()` - Panic on Failure

Use `.MustParse()` when validation failure should terminate the program:

```go
// Configuration validation during startup
config := ConfigSchema.MustParse(configData)
// Will panic if validation fails - use only when failure is not recoverable
```

## Handling Errors

When validation fails, GoZod returns a structured `ZodError` with detailed information:

```go
schema := gozod.String().Min(5)
_, err := schema.Parse("hi")

if err != nil {
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        for _, issue := range zodErr.Issues {
            fmt.Printf("Path: %v\n", issue.Path)       // [] (root level)
            fmt.Printf("Code: %s\n", issue.Code)       // "too_small"
            fmt.Printf("Message: %s\n", issue.Message) // "String must be at least 5 characters"
            fmt.Printf("Input: %v\n", issue.Input)     // "hi"
        }
    }
}
```

### Custom Error Messages

You can provide custom error messages for better user experience:

```go
// Static error message
schema := gozod.String().Min(5, gozod.SchemaParams{
    Error: "Username must be at least 5 characters",
})

// Dynamic error message
dynamicSchema := gozod.String().Min(5, gozod.SchemaParams{
    Error: func(issue gozod.ZodRawIssue) string {
        return fmt.Sprintf("String must be at least %d characters, got %d",
            issue.GetMinimum(), len(issue.Input.(string)))
    },
})
```

### Error Formatting

GoZod provides different error formatting options:

```go
_, err := UserSchema.Parse(invalidData)
if err != nil {
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // Human-readable format
        pretty := gozod.PrettifyError(zodErr)
        fmt.Println(pretty)
        
        // Nested structure format
        formatted := gozod.FormatError(zodErr)
        
        // Flat structure for forms
        flattened := gozod.FlattenError(zodErr)
        for field, errors := range flattened.FieldErrors {
            fmt.Printf("Field %s: %v\n", field, errors)
        }
    }
}
```

## Type Inference

GoZod preserves Go's type system and provides intelligent type inference:

### Pointer Identity Preservation

```go
stringSchema := gozod.String()

// Basic type inference
result1, _ := stringSchema.Parse("hello")      // result1: "hello" (string)

// Pointer identity preservation
input := "world"
result2, _ := stringSchema.Parse(&input)       // result2: &input (same memory address)
fmt.Println(result2 == &input)                // true - same memory address
```

### Optional vs Nilable Distinction

Understanding the semantic difference:

```go
// Optional: Handles "missing field" semantics
optionalSchema := gozod.String().Optional()
result, _ := optionalSchema.Parse(nil)         // result: nil (generic nil)

// Nilable: Handles "null value" semantics  
nilableSchema := gozod.String().Nilable()
result, _ := nilableSchema.Parse(nil)          // result: (*string)(nil) (typed nil pointer)
```

## Common Patterns

### Schema Composition

```go
// Reusable base schemas
baseUser := gozod.Object(gozod.ObjectSchema{
    "id":   gozod.Int().Min(1),
    "name": gozod.String().Min(2),
})

// Extend for different contexts
adminUser := baseUser.Extend(gozod.ObjectSchema{
    "permissions": gozod.Slice(gozod.String()),
    "role":        gozod.Literal("admin"),
})

publicUser := baseUser.Pick([]string{"id", "name"})
```

### Method Chaining

Validations execute in the order they're chained:

```go
// Chain multiple validations
emailSchema := gozod.String().Min(5).Max(50).Email()
result, err := emailSchema.Parse("user@example.com")  // All validations pass

// Early failure - stops at first failing validation
_, err = emailSchema.Parse("a@b")  // Fails at Min(5), doesn't reach Email()
```

### Default Values and Fallbacks

```go
// Static default value
nameWithDefault := gozod.String().Default("Anonymous")
result, err := nameWithDefault.Parse(nil)  // "Anonymous"

// Function-based default
timestampDefault := gozod.String().DefaultFunc(func() string {
    return time.Now().Format(time.RFC3339)
})

// Prefault provides fallback for ANY validation failure
safeValue := gozod.String().Min(5).Prefault("fallback")
result, err := safeValue.Parse("hi")  // "fallback" (too short, uses fallback)
```

### Type Coercion

**Note**: In alignment with TypeScript Zod v4, coercion is now limited to primitive types only.

```go
// Basic primitive type coercion
stringSchema := gozod.Coerce.String()
result, _ := stringSchema.Parse(123)     // "123"

numberSchema := gozod.Coerce.Number()
result, _ = numberSchema.Parse("42")     // 42.0

boolSchema := gozod.Coerce.Bool()
result, _ = boolSchema.Parse("true")     // true

bigintSchema := gozod.Coerce.BigInt()
result, _ = bigintSchema.Parse("123")    // *big.Int(123)

// Schema-level coercion for primitives
coerceSchema := gozod.String(gozod.SchemaParams{Coerce: true})
result, _ := coerceSchema.Parse(123)  // "123"

// Collection types (object, map, record, slice) no longer support coercion
```

### Custom Validation

```go
// Custom validation logic
passwordSchema := gozod.String().Min(8).Refine(func(s string) bool {
    return strings.ContainsAny(s, "!@#$%^&*()")
}, gozod.SchemaParams{
    Error: "Password must contain at least one special character",
})
```

### Union Types

```go
// Union types (OR logic)
stringOrNumber := gozod.Union([]gozod.ZodType[any, any]{
    gozod.String(),
    gozod.Int(),
})

// Discriminated union for better performance
apiResponse := gozod.DiscriminatedUnion("status", []gozod.ZodType[any, any]{
    gozod.Object(gozod.ObjectSchema{
        "status": gozod.Literal("success"),
        "data":   gozod.String(),
    }),
    gozod.Object(gozod.ObjectSchema{
        "status": gozod.Literal("error"),
        "error":  gozod.String(),
    }),
})
```

## Working with JSON APIs

Here's a practical example of using GoZod in an HTTP API:

```go
func handleAPIRequest(w http.ResponseWriter, r *http.Request) {
    var requestData map[string]any
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    result, err := APIRequestSchema.Parse(requestData)
    if err != nil {
        var zodErr *gozod.ZodError
        if errors.As(err, &zodErr) {
            flattened := gozod.FlattenError(zodErr)
            
            errorResponse := map[string]any{
                "success": false,
                "errors":  flattened.FieldErrors,
            }
            
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(errorResponse)
            return
        }
    }
    
    // Process validated data
    validData := result.(map[string]any)
    processAPIRequest(validData)
}
```

## Configuration Validation

A common use case is validating application configuration:

```go
type AppConfig struct {
    Port     int    `json:"port"`
    Database string `json:"database"`
    Debug    bool   `json:"debug"`
}

var ConfigSchema = gozod.Object(gozod.ObjectSchema{
    "port":     gozod.Int().Min(1).Max(65535),
    "database": gozod.String().URL(),
    "debug":    gozod.Bool().Default(false),
})

func loadConfig(filename string) (*AppConfig, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var configData map[string]any
    if err := json.Unmarshal(data, &configData); err != nil {
        return nil, err
    }
    
    validated, err := ConfigSchema.Parse(configData)
    if err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    // Convert to struct
    validConfig := validated.(map[string]any)
    return &AppConfig{
        Port:     validConfig["port"].(int),
        Database: validConfig["database"].(string),
        Debug:    validConfig["debug"].(bool),
    }, nil
}
```

---

This guide covers the fundamental patterns for using GoZod effectively. For complete type reference, method documentation, and detailed examples, see the [API Reference](api.md).
