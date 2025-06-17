# Error Customization

GoZod provides flexible error customization to create user-friendly validation messages. This document explains how to work with validation errors and customize error messages.

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
- **Context-specific properties** - Additional details accessible via getter methods

## Custom Error Messages

Every GoZod schema and validation method accepts custom error messages through `SchemaParams`:

```go
// Basic custom error message
schema := gozod.String(gozod.SchemaParams{
    Error: "Please enter a valid string",
})

_, err := schema.Parse(42)
if err != nil {
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        fmt.Println(zodErr.Issues[0].Message) // "Please enter a valid string"
    }
}
```

## Where to Apply Custom Errors

Custom error messages can be applied to any schema or validation method:

```go
// Schema constructors
stringSchema := gozod.String(gozod.SchemaParams{Error: "Must be text"})
numberSchema := gozod.Int(gozod.SchemaParams{Error: "Must be a number"})

// Validation methods
emailSchema := gozod.String().Email(gozod.SchemaParams{Error: "Invalid email address"})
minLengthSchema := gozod.String().Min(5, gozod.SchemaParams{Error: "Too short"})

// Collection validations
arraySchema := gozod.Slice(gozod.String(), gozod.SchemaParams{Error: "Must be an array of strings"})

// Object validations
userSchema := gozod.Object(gozod.ObjectSchema{
    "name": gozod.String(gozod.SchemaParams{Error: "Name is required"}),
    "age":  gozod.Int().Min(18, gozod.SchemaParams{Error: "Must be at least 18"}),
})
```

## Dynamic Error Messages

Use functions to create context-aware error messages:

```go
schema := gozod.String().Min(8, gozod.SchemaParams{
    Error: func(issue gozod.ZodRawIssue) string {
        if issue.Input == nil {
            return "Password is required"
        }
        current := len(fmt.Sprintf("%v", issue.Input))
        required := issue.GetMinimum()
        return fmt.Sprintf("Password too short (%d/%d characters)", current, required)
    },
})
```

## Context-Aware Error Functions

Error functions receive detailed validation context:

```go
userSchema := gozod.Object(gozod.ObjectSchema{
    "username": gozod.String().Min(3, gozod.SchemaParams{
        Error: func(issue gozod.ZodRawIssue) string {
            switch issue.Code {
            case "invalid_type":
                return "Username must be text"
            case "too_small":
                return fmt.Sprintf("Username must be at least %v characters", issue.GetMinimum())
            default:
                return "Invalid username"
            }
        },
    }),
    "age": gozod.Int().Min(18, gozod.SchemaParams{
        Error: func(issue gozod.ZodRawIssue) string {
            if issue.Code == "too_small" {
                return "Must be at least 18 years old"
            }
            return "Invalid age"
        },
    }),
})
```

## Available Context Properties

The `ZodRawIssue` provides validation context through getter methods:

```go
// Type validation
if expected := issue.GetExpected(); expected != "" {
    fmt.Printf("Expected: %s\n", expected)
}
if received := issue.GetReceived(); received != "" {
    fmt.Printf("Received: %s\n", received)
}

// Size constraints
if minimum := issue.GetMinimum(); minimum != nil {
    fmt.Printf("Minimum: %v\n", minimum)
}
if maximum := issue.GetMaximum(); maximum != nil {
    fmt.Printf("Maximum: %v\n", maximum)
}

// Format validation
if format := issue.GetFormat(); format != "" {
    fmt.Printf("Format: %s\n", format) // "email", "url", etc.
}

// Pattern validation
if pattern := issue.GetPattern(); pattern != "" {
    fmt.Printf("Pattern: %s\n", pattern)
}

// Object validation
if keys := issue.GetKeys(); len(keys) > 0 {
    fmt.Printf("Unrecognized keys: %v\n", keys)
}
```

## Error Codes

GoZod uses standardized error codes compatible with TypeScript Zod:

| Code | Description | Available Properties |
|------|-------------|---------------------|
| `invalid_type` | Wrong data type | `GetExpected()`, `GetReceived()` |
| `too_small` | Below minimum constraint | `GetMinimum()`, `GetInclusive()` |
| `too_big` | Above maximum constraint | `GetMaximum()`, `GetInclusive()` |
| `invalid_format` | Format validation failed | `GetFormat()` |
| `not_multiple_of` | Not a multiple of required value | `GetDivisor()` |
| `unrecognized_keys` | Object contains unknown keys | `GetKeys()` |
| `invalid_union` | No union member matched | - |
| `invalid_value` | Value not in allowed set | `GetValues()` |
| `custom` | Custom validation failed | - |

## Form Validation Example

```go
package main

import (
    "errors"
    "fmt"
    "github.com/kaptinlin/gozod"
)

func validateUserForm(data map[string]interface{}) error {
    schema := gozod.Object(gozod.ObjectSchema{
        "username": gozod.String().Min(3, gozod.SchemaParams{
            Error: "Username must be at least 3 characters",
        }),
        "email": gozod.String().Email(gozod.SchemaParams{
            Error: "Please enter a valid email address",
        }),
        "password": gozod.String().Min(8, gozod.SchemaParams{
            Error: func(issue gozod.ZodRawIssue) string {
                return fmt.Sprintf("Password must be at least %v characters", issue.GetMinimum())
            },
        }),
        "age": gozod.Int().Min(18, gozod.SchemaParams{
            Error: "You must be at least 18 years old",
        }),
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

func main() {
    invalidData := map[string]interface{}{
        "username": "ab",
        "email":    "not-an-email",
        "password": "short",
        "age":      16,
    }
    
    validateUserForm(invalidData)
    // Output:
    // Validation errors:
    // - username: Username must be at least 3 characters
    // - email: Please enter a valid email address
    // - password: Password must be at least 8 characters
    // - age: You must be at least 18 years old
}
```

## Global Error Configuration

Configure error handling globally:

```go
// Set global custom error handler
gozod.Config(&gozod.ZodConfig{
    CustomError: func(issue gozod.ZodRawIssue) string {
        switch issue.Code {
        case "invalid_type":
            return fmt.Sprintf("Expected %s, received %s", 
                              issue.GetExpected(), issue.GetReceived())
        case "too_small":
            return fmt.Sprintf("Value too small (minimum: %v)", issue.GetMinimum())
        case "too_big":
            return fmt.Sprintf("Value too large (maximum: %v)", issue.GetMaximum())
        default:
            return fmt.Sprintf("Validation failed: %s", issue.Code)
        }
    },
})
```

## Error Priority Chain

GoZod resolves error messages using this priority order:

1. **Schema-level error** (highest priority)
2. **ParseContext.Error** 
3. **Global CustomError**
4. **Global LocaleError**
5. **Default error message** (lowest priority)

```go
// Schema-level error (highest priority)
schema := gozod.String(gozod.SchemaParams{
    Error: "Schema-level error",
})

// ParseContext error (second priority)
ctx := &gozod.ParseContext{
    Error: func(issue gozod.ZodRawIssue) string {
        return "Context-level error"
    },
}

_, err := schema.Parse(42, ctx)
// Uses "Schema-level error" (highest priority)
```

## Internationalization

Create localized error messages:

```go
// Error message mapping
var errorMessages = map[string]map[string]string{
    "en": {
        "username_required": "Username is required",
        "email_invalid":     "Invalid email address",
        "password_short":    "Password too short",
    },
    "zh": {
        "username_required": "用户名为必填项",
        "email_invalid":     "邮箱地址无效",
        "password_short":    "密码过短",
    },
}

func localizedError(locale string) func(gozod.ZodRawIssue) string {
    messages := errorMessages[locale]
    if messages == nil {
        messages = errorMessages["en"] // fallback
    }
    
    return func(issue gozod.ZodRawIssue) string {
        switch issue.Code {
        case "invalid_type":
            if issue.Input == nil {
                return messages["username_required"]
            }
            return messages["email_invalid"]
        case "too_small":
            return messages["password_short"]
        default:
            return "Validation error"
        }
    }
}

// Use with schema
schema := gozod.String(gozod.SchemaParams{
    Error: localizedError("zh"),
})
```

## Best Practices

### 1. User-Friendly Messages

Write clear, actionable error messages:

```go
// ❌ Technical message
gozod.String().Email(gozod.SchemaParams{
    Error: "Email format validation failed",
})

// ✅ User-friendly message
gozod.String().Email(gozod.SchemaParams{
    Error: "Please enter a valid email address",
})
```

### 2. Provide Context

Include helpful context in error messages:

```go
gozod.String().Min(8, gozod.SchemaParams{
    Error: func(issue gozod.ZodRawIssue) string {
        current := len(fmt.Sprintf("%v", issue.Input))
        required := issue.GetMinimum()
        return fmt.Sprintf("Password must be at least %v characters (got %d)", 
                          required, current)
    },
})
```

### 3. Consistent Formatting

Maintain consistent error message style:

```go
func formatFieldError(fieldName string, requirement string) string {
    return fmt.Sprintf("%s %s", fieldName, requirement)
}

usernameSchema := gozod.String().Min(3, gozod.SchemaParams{
    Error: formatFieldError("Username", "must be at least 3 characters"),
})

passwordSchema := gozod.String().Min(8, gozod.SchemaParams{
    Error: formatFieldError("Password", "must be at least 8 characters"),
})
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
            return fmt.Sprintf("%s must be %s", fieldName, issue.GetExpected())
        case "too_small":
            return fmt.Sprintf("%s must be at least %v characters", fieldName, issue.GetMinimum())
        case "invalid_format":
            return fmt.Sprintf("%s has invalid format", fieldName)
        default:
            return fmt.Sprintf("%s is invalid", fieldName)
        }
    }
}

userSchema := gozod.Object(gozod.ObjectSchema{
    "username": gozod.String().Min(3, gozod.SchemaParams{Error: createFieldError("Username")}),
    "email":    gozod.String().Email(gozod.SchemaParams{Error: createFieldError("Email")}),
})
```

## Advanced Usage

### Nested Object Validation

```go
profileSchema := gozod.Object(gozod.ObjectSchema{
    "personal": gozod.Object(gozod.ObjectSchema{
        "firstName": gozod.String(gozod.SchemaParams{
            Error: "First name is required",
        }),
        "lastName": gozod.String(gozod.SchemaParams{
            Error: "Last name is required",
        }),
    }),
    "contact": gozod.Object(gozod.ObjectSchema{
        "email": gozod.String().Email(gozod.SchemaParams{
            Error: "Valid email address required",
        }),
        "phone": gozod.String().Min(10, gozod.SchemaParams{
            Error: "Phone number must be at least 10 digits",
        }),
    }),
})
```

### Conditional Error Messages

```go
ageSchema := gozod.Int(gozod.SchemaParams{
    Error: func(issue gozod.ZodRawIssue) string {
        if issue.Code == "invalid_type" {
            if issue.Input == nil {
                return "Age is required"
            }
            return "Age must be a number"
        }
        if issue.Code == "too_small" {
            minimum := issue.GetMinimum()
            if minimum != nil && minimum.(int) == 18 {
                return "You must be at least 18 years old to register"
            }
            return fmt.Sprintf("Age must be at least %v", minimum)
        }
        return "Invalid age"
    },
}).Min(18)
```

Error customization in GoZod provides the flexibility to create user-friendly validation experiences. By leveraging custom error messages, context-aware functions, and proper error handling patterns, you can build applications that provide clear guidance when validation fails.
