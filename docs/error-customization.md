# Error Customization

GoZod provides flexible error customization to create user-friendly validation messages that match TypeScript Zod v4 patterns. This document explains how to work with validation errors and customize error messages following Go language conventions.

## Understanding ZodError

When validation fails, GoZod returns a `*ZodError` containing detailed information about each validation issue:

```go
package main

import (
    "errors"
    "fmt"
    "github.com/kaptinlin/gozod"
)

func main() {
    schema := gozod.String()
    _, err := schema.Parse(42)
    
    if err != nil {
        var zodErr *gozod.ZodError
        if errors.As(err, &zodErr) {
            issue := zodErr.Issues[0]
            fmt.Printf("Code: %s\n", issue.Code)       // "invalid_type"
            fmt.Printf("Message: %s\n", issue.Message) // "Invalid input: expected string, received number"
            fmt.Printf("Path: %v\n", issue.Path)       // []
            fmt.Printf("Input: %v\n", issue.Input)     // 42
        }
    }
}
```

Each validation issue contains:
- **Code** - Error type identifier (e.g., "invalid_type", "too_small")
- **Message** - Human-readable error description
- **Path** - Field path where validation failed
- **Input** - The invalid input value
- **Properties** - Additional context-specific metadata

## Schema-Level Custom Errors

Every GoZod schema accepts custom error messages through the error parameter:

```go
// String parameter shorthand
schema := gozod.String("Please enter a valid string")

// SchemaParams struct
schema := gozod.String(gozod.SchemaParams{
    Error: "Please enter a valid string",
})

// Validation method errors
emailSchema := gozod.String().Email("Please enter a valid email address")
minLengthSchema := gozod.String().Min(5, "Password must be at least 5 characters")

// Collection validations
arraySchema := gozod.Slice(gozod.String(), "Must be an array of strings")
```

## Dynamic Error Functions

Use error functions for context-aware messages:

```go
schema := gozod.String().Min(8, func(issue gozod.ZodRawIssue) string {
    if issue.Input == nil {
        return "Password is required"
    }
    current := len(fmt.Sprintf("%v", issue.Input))
    minimum := issue.Minimum()
    return fmt.Sprintf("Password too short (%d/%v characters)", current, minimum)
})
```

## Per-Parse Error Customization

Customize errors on a per-parse basis by passing an error function to the parse method:

```go
schema := gozod.String()

result, err := schema.ParseWithContext(12, &gozod.ParseContext{
    Error: func(issue gozod.ZodRawIssue) string {
        return "Custom per-parse error message"
    },
})
```

This has lower precedence than schema-level custom messages.

## Global Error Configuration

Set global error customization using `gozod.Config()`:

```go
// Set global custom error map
gozod.Config(&gozod.ZodConfig{
    CustomError: func(issue gozod.ZodRawIssue) string {
        switch issue.Code {
        case "invalid_type":
            return fmt.Sprintf("Expected %s, received %s", 
                              issue.Expected(), issue.Received())
        case "too_small":
            return fmt.Sprintf("Value too small (minimum: %v)", issue.Minimum())
        case "too_big":
            return fmt.Sprintf("Value too large (maximum: %v)", issue.Maximum())
        default:
            return fmt.Sprintf("Validation failed: %s", issue.Code)
        }
    },
})
```

## Internationalization

GoZod provides built-in locale support for error messages:

```go
import "github.com/kaptinlin/gozod/locales"

// Set global locale to Chinese
gozod.Config(locales.ZhCN())

// All subsequent validation errors will be in Chinese
_, err := gozod.String().Email().Parse("invalid-email")
// Error message will be: "格式无效：请输入有效的电子邮件地址"
```

Available locales:
- `locales.EN()` - English (default)
- `locales.ZhCN()` - Simplified Chinese

## Error Message Context

Error functions receive detailed validation context through the `ZodRawIssue`:

```go
userSchema := gozod.Object(gozod.ObjectSchema{
    "username": gozod.String().Min(3, func(issue gozod.ZodRawIssue) string {
        switch issue.Code {
        case "invalid_type":
            return "Username must be text"
        case "too_small":
            return fmt.Sprintf("Username must be at least %v characters", issue.Minimum())
        default:
            return "Invalid username"
        }
    }),
    "age": gozod.Int().Min(18, func(issue gozod.ZodRawIssue) string {
        if issue.Code == "too_small" {
            return "Must be at least 18 years old"
        }
        return "Invalid age"
    }),
})
```

### Available Context Properties

The `ZodRawIssue` provides validation context through getter methods:

```go
// Type validation
if expected := issue.Expected(); expected != "" {
    fmt.Printf("Expected: %s\n", expected)
}
if received := issue.Received(); received != "" {
    fmt.Printf("Received: %s\n", received)
}

// Size constraints
if minimum := issue.Minimum(); minimum != nil {
    fmt.Printf("Minimum: %v\n", minimum)
}
if maximum := issue.Maximum(); maximum != nil {
    fmt.Printf("Maximum: %v\n", maximum)
}

// Format validation
if format := issue.Format(); format != "" {
    fmt.Printf("Format: %s\n", format) // "email", "url", etc.
}

// Pattern validation
if pattern := issue.Pattern(); pattern != "" {
    fmt.Printf("Pattern: %s\n", pattern)
}

// Object validation
if keys := issue.Keys(); len(keys) > 0 {
    fmt.Printf("Unrecognized keys: %v\n", keys)
}
```

## Error Codes

GoZod uses standardized error codes compatible with TypeScript Zod v4:

| Code | Description | Available Properties |
|------|-------------|---------------------|
| `invalid_type` | Wrong data type | `Expected()`, `Received()` |
| `too_small` | Below minimum constraint | `Minimum()`, `Inclusive()` |
| `too_big` | Above maximum constraint | `Maximum()`, `Inclusive()` |
| `invalid_format` | Format validation failed | `Format()` |
| `not_multiple_of` | Not a multiple of required value | `Divisor()` |
| `unrecognized_keys` | Object contains unknown keys | `Keys()` |
| `invalid_union` | No union member matched | - |
| `invalid_value` | Value not in allowed set | `Values()` |
| `custom` | Custom validation failed | - |

## Error Message Precedence

GoZod resolves error messages using this priority order (highest to lowest):

1. **Schema-level error** - Custom messages defined in schemas
2. **Per-parse error** - Custom error functions passed to Parse methods
3. **Global CustomError** - Global error map set via `gozod.Config()`
4. **Global LocaleError** - Locale-specific error messages
5. **Default error message** - Built-in English messages

```go
// Schema-level error (highest priority)
schema := gozod.String("Schema-level error")

// Per-parse error (second priority)
ctx := &gozod.ParseContext{
    Error: func(issue gozod.ZodRawIssue) string {
        return "Context-level error"
    },
}

_, err := schema.ParseWithContext(42, ctx)
// Uses "Schema-level error" (highest priority)
```

## Best Practices

### 1. User-Friendly Messages

Write clear, actionable error messages:

```go
// ❌ Technical message
gozod.String().Email("Email format validation failed")

// ✅ User-friendly message
gozod.String().Email("Please enter a valid email address")
```

### 2. Provide Context

Include helpful context in error messages:

```go
gozod.String().Min(8, func(issue gozod.ZodRawIssue) string {
    current := len(fmt.Sprintf("%v", issue.Input))
    required := issue.Minimum()
    return fmt.Sprintf("Password must be at least %v characters (got %d)", 
                      required, current)
})
```

### 3. Consistent Formatting

Maintain consistent error message style:

```go
func formatFieldError(fieldName string, requirement string) string {
    return fmt.Sprintf("%s %s", fieldName, requirement)
}

usernameSchema := gozod.String().Min(3, 
    formatFieldError("Username", "must be at least 3 characters"))

passwordSchema := gozod.String().Min(8, 
    formatFieldError("Password", "must be at least 8 characters"))
```

### 4. Handle Different Error Types

Create comprehensive error handlers:

```go
func createFieldError(fieldName string) func(gozod.ZodRawIssue) string {
    return func(issue gozod.ZodRawIssue) string {
        switch issue.Code {
        case "invalid_type":
            if issue.Input == nil {
                return fmt.Sprintf("%s is required", fieldName)
            }
            return fmt.Sprintf("%s must be %s", fieldName, issue.Expected())
        case "too_small":
            return fmt.Sprintf("%s must be at least %v characters", fieldName, issue.Minimum())
        case "invalid_format":
            return fmt.Sprintf("%s has invalid format", fieldName)
        default:
            return fmt.Sprintf("%s is invalid", fieldName)
        }
    }
}

userSchema := gozod.Object(gozod.ObjectSchema{
    "username": gozod.String().Min(3, createFieldError("Username")),
    "email":    gozod.String().Email(createFieldError("Email")),
})
```

## Advanced Usage

### Nested Object Validation

```go
profileSchema := gozod.Object(gozod.ObjectSchema{
    "personal": gozod.Object(gozod.ObjectSchema{
        "firstName": gozod.String("First name is required"),
        "lastName": gozod.String("Last name is required"),
    }),
    "contact": gozod.Object(gozod.ObjectSchema{
        "email": gozod.String().Email("Valid email address required"),
        "phone": gozod.String().Min(10, "Phone number must be at least 10 digits"),
    }),
})
```

### Conditional Error Messages

```go
ageSchema := gozod.Int(func(issue gozod.ZodRawIssue) string {
    switch issue.Code {
    case "invalid_type":
        if issue.Input == nil {
            return "Age is required"
        }
        return "Age must be a number"
    case "too_small":
        minimum := issue.Minimum()
        if minimum != nil && minimum.(int) == 18 {
            return "You must be at least 18 years old to register"
        }
        return fmt.Sprintf("Age must be at least %v", minimum)
    default:
        return "Invalid age"
    }
}).Min(18)
```

### Form Validation Example

```go
func validateUserForm(data map[string]any) error {
    schema := gozod.Object(gozod.ObjectSchema{
        "username": gozod.String().Min(3, "Username must be at least 3 characters"),
        "email": gozod.String().Email("Please enter a valid email address"),
        "password": gozod.String().Min(8, func(issue gozod.ZodRawIssue) string {
            return fmt.Sprintf("Password must be at least %v characters", issue.Minimum())
        }),
        "age": gozod.Int().Min(18, "You must be at least 18 years old"),
    })
    
    _, err := schema.Parse(data)
    if err != nil {
        var zodErr *gozod.ZodError
        if errors.As(err, &zodErr) {
            fmt.Println("Validation errors:")
            for _, issue := range zodErr.Issues {
                field := "form"
                if len(issue.Path) > 0 {
                    field = fmt.Sprintf("%v", issue.Path[0])
                }
                fmt.Printf("- %s: %s\n", field, issue.Message)
            }
        }
        return err
    }
    
    return nil
}
```

Error customization in GoZod provides the flexibility to create user-friendly validation experiences that follow Go language conventions while maintaining compatibility with TypeScript Zod v4 patterns. By leveraging custom error messages, context-aware functions, and proper error handling patterns, you can build applications that provide clear guidance when validation fails.