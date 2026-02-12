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

GoZod provides multiple methods for parsing data:

### `.Parse()` - Standard Error Handling

Use `.Parse()` for standard Go error handling patterns with flexible input types:

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

### `.StrictParse()` - Compile-Time Type Safety

Use `.StrictParse()` when you have the exact input type and want compile-time guarantees:

```go
// With known string type
nameSchema := gozod.String().Min(2).Max(50)
name := "Alice"
result, err := nameSchema.StrictParse(name)  // Input type guaranteed at compile-time
if err != nil {
    log.Printf("Name validation failed: %v", err)
    return
}
fmt.Printf("Valid name: %s\n", result)

// With struct types
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

userSchema := gozod.Struct[User]()
user := User{Name: "John", Email: "john@example.com"}
validatedUser, err := userSchema.StrictParse(user)  // Type-safe struct validation
if err != nil {
    log.Printf("User validation failed: %v", err)
    return
}
fmt.Printf("Valid user: %+v\n", validatedUser)
```

### `.MustParse()` and `.MustStrictParse()` - Panic on Failure

Use these methods when validation failure should terminate the program:

```go
// Configuration validation during startup
config := ConfigSchema.MustParse(configData)
// Will panic if validation fails - use only when failure is not recoverable

// Type-safe panic version
user := User{Name: "Admin", Email: "admin@example.com"}
validatedUser := userSchema.MustStrictParse(user)  // Panics on validation failure
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
schema := gozod.String().Min(5, "Username must be at least 5 characters")

// Dynamic error message with custom error function
dynamicSchema := gozod.String().Min(5, func(issue gozod.ZodRawIssue) string {
    return fmt.Sprintf("String must be at least %d characters, got %d",
        issue.Minimum(), len(issue.Input.(string)))
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
        
        // Nested tree structure format
        tree := gozod.TreeifyError(zodErr)
        
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

### Type Requirements and Pointer Identity

```go
// GoZod uses type-safe semantics
value := "hello"

// Value schemas: strict value input only
stringSchema := gozod.String()
result1, _ := stringSchema.Parse("hello")      // ✅ string → string
// result1, _ := stringSchema.Parse(&value)    // ❌ Error: requires string, got *string

// Pointer schemas: strict pointer input only  
ptrSchema := gozod.StringPtr()
result2, _ := ptrSchema.Parse(&value)          // ✅ *string → *string (preserves identity)
// result2, _ := ptrSchema.Parse("hello")      // ❌ Error: requires *string, got string

// Optional/Nilable: flexible input with pointer output
optionalSchema := gozod.String().Optional()   // Returns *string, flexible input
result3, _ := optionalSchema.Parse("hello")    // ✅ string → *string (new pointer)
result4, _ := optionalSchema.Parse(&value)     // ✅ *string → *string (preserves identity)
fmt.Println(result4 == &value)                // true - same memory address

// Helper functions for convenience
func Ptr[T any](v T) *T { return &v }
result5, _ := ptrSchema.Parse(Ptr("test"))     // Create pointer inline
```

### Optional vs Nilable Distinction

Understanding the semantic difference:

```go
// Optional: Handles "missing field" semantics
optionalSchema := gozod.String().Optional()
result, _ := optionalSchema.Parse(nil)         // result: nil

// Nilable: Handles "null value" semantics by returning a typed nil pointer
nilableSchema := gozod.String().Nilable()
result, _ = nilableSchema.Parse(nil)          // result: (*string)(nil)
```

### Working with Pointer Schemas

Schemas like `gozod.StringPtr()` and `gozod.IntPtr()` are useful when you need to distinguish between a field that was not provided versus a field that was explicitly set to its zero value (e.g., `""` or `0`). These schemas require strict pointer input and preserve pointer identity, making them ideal for handling partial data updates.

**Important**: All GoZod schemas use type-safe requirements:
- `gozod.String()` only accepts `string` input
- `gozod.StringPtr()` only accepts `*string` input  
- `gozod.Struct[T]()` only accepts `T` input
- `gozod.StructPtr[T]()` only accepts `*T` input

```go
import (
	"encoding/json"
	"fmt"
	"github.com/kaptinlin/gozod"
)

// First, define the Go struct you intend to populate.
// Using pointer types is crucial for optional fields in update operations.
type UpdatePayload struct {
	Nickname *string `json:"nickname"`
	Age      *int    `json:"age"`
}

// Next, create a schema from your struct. The keys in the schema
// should match the `json` tags of your struct fields.
var updateSchema = gozod.Struct[UpdatePayload](gozod.StructSchema{
	"nickname": gozod.StringPtr().Min(2).Optional(),
	"age":      gozod.IntPtr().Min(0).Optional(),
})

// Example: Simulate a PATCH request where only the nickname is provided.
jsonData := `{"nickname": "gopher"}`
var payload UpdatePayload
_ = json.Unmarshal([]byte(jsonData), &payload)

// Parse the struct directly. GoZod returns a new, validated struct.
validatedPayload, err := updateSchema.Parse(payload)
if err != nil {
	// ... handle error
}

// You can now safely work with the strongly-typed, validated struct.
if validatedPayload.Nickname != nil {
	fmt.Printf("Nickname was provided: %s\n", *validatedPayload.Nickname)
} else {
	fmt.Println("Nickname was not provided.")
}

if validatedPayload.Age != nil {
	fmt.Printf("Age was provided: %d\n", *validatedPayload.Age)
} else {
	fmt.Println("Age was not provided.")
}
// Output:
// Nickname was provided: gopher
// Age was not provided.

// Strict type requirements in action:
valueSchema := gozod.String().Min(2)
ptrSchema := gozod.StringPtr().Min(2)
name := "Alice"

// Value schema requires value input
result1, err := valueSchema.Parse("Alice")  // ✅ string → string
// result1, err := valueSchema.Parse(&name) // ❌ Error: requires string, got *string

// Pointer schema requires pointer input  
result2, err := ptrSchema.Parse(&name)      // ✅ *string → *string
// result2, err := ptrSchema.Parse("Alice") // ❌ Error: requires *string, got string

if err == nil {
    fmt.Printf("Validated: %s\n", *result2)
    fmt.Printf("Same pointer: %v\n", result2 == &name)  // true - identity preserved
}
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

// Prefault provides pre-parse default for nil input (goes through full parsing)
safeValue := gozod.String().Transform(func(s string) string { return strings.ToUpper(s) }).Prefault("fallback")
result, err := safeValue.Parse(nil)  // "FALLBACK" (nil input uses prefault, then transforms)
```

### Type Coercion

**Note**: In alignment with TypeScript Zod v4, coercion is now limited to primitive types only.

```go
import "github.com/kaptinlin/gozod/coerce"

// Basic primitive type coercion
stringSchema := coerce.String()
result, _ := stringSchema.Parse(123)     // "123"

numberSchema := coerce.Number()
result, _ = numberSchema.Parse("42")     // 42.0

boolSchema := coerce.Bool()
result, _ = boolSchema.Parse("true")     // true

bigintSchema := coerce.BigInt()
result, _ = bigintSchema.Parse("123")    // *big.Int(123)

// Collection types (object, map, record, slice) no longer support coercion
```

### Custom Validation

```go
// Custom validation logic using Refine
passwordSchema := gozod.String().Min(8).Refine(func(s string) bool {
    return strings.ContainsAny(s, "!@#$%^&*()")
}, "Password must contain at least one special character")

// Same idea with the lightweight Check helper
passwordSchema := gozod.String().Check(func(v string, p *gozod.ParsePayload) {
    if len(v) < 8 {
        p.AddIssueWithCode(gozod.IssueTooSmall, "password too short")
    }
    if !strings.ContainsAny(v, "!@#$%^&*()") {
        p.AddIssueWithMessage("missing special character")
    }
})

// Or using Min method with custom error message
passwordSchema2 := gozod.String().Min(8, "Password too short").Refine(func(s string) bool {
    return strings.ContainsAny(s, "!@#$%^&*()")
}, "Password must contain at least one special character")

// .Check gives you full access to the accumulating ParsePayload so you can
// push multiple issues in one pass without defining separate refinements.
```

### Union Types

```go
// Union types (OR logic)
stringOrNumber := gozod.Union(gozod.String(), gozod.Int())

// Discriminated union for better performance
apiResponse := gozod.DiscriminatedUnion("status", 
    gozod.Object(gozod.ObjectSchema{
        "status": gozod.Literal("success"),
        "data":   gozod.String(),
    }),
    gozod.Object(gozod.ObjectSchema{
        "status": gozod.Literal("error"),
        "error":  gozod.String(),
    }),
)
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
