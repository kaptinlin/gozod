# Error Formatting

GoZod provides utilities to format validation errors for different use cases. This document covers three formatting functions that transform ZodError into structured, readable, or flat representations.

## Formatting Functions

GoZod offers three error formatting utilities:

1. **`FormatError()`** - Creates nested error structure mirroring schema hierarchy
2. **`PrettifyError()`** - Generates human-readable string representation  
3. **`FlattenError()`** - Produces flat structure with form-level and field-level errors

## Basic Example

```go
package main

import (
    "errors"
    "fmt"
    "github.com/kaptinlin/gozod"
)

func main() {
    // Define validation schema
    schema := gozod.Object(gozod.ObjectSchema{
        "username": gozod.String().Min(3),
        "numbers":  gozod.Slice(gozod.Int().Positive()),
    })

    // Invalid data with multiple issues
    invalidData := map[string]any{
        "username": "ab",                        // too short
        "numbers":  []any{1, "text", -5}, // mixed types and negative number
        "extra":    "not allowed",               // unexpected key
    }

    _, err := schema.Parse(invalidData)
    
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // 1. PrettifyError - Human-readable output
        pretty := gozod.PrettifyError(zodErr)
        fmt.Println("=== PrettifyError ===")
        fmt.Println(pretty)
        
        // 2. FormatError - Nested structure
        formatted := gozod.FormatError(zodErr)
        fmt.Println("\n=== FormatError ===")
        fmt.Printf("Username errors: %v\n", formatted["username"])
        
        // 3. FlattenError - Flat structure
        flattened := gozod.FlattenError(zodErr)
        fmt.Println("\n=== FlattenError ===")
        fmt.Printf("Field errors: %v\n", flattened.FieldErrors["username"])
    }
}
```

## FormatError

Converts ZodError into a nested structure that mirrors your schema's shape. Useful for accessing errors at specific field paths.

```go
func FormatError(err *ZodError) ZodFormattedError
func FormatErrorWithMapper(err *ZodError, mapper func(ZodIssue) string) ZodFormattedError
```

The `ZodFormattedError` type is a map structure where each field contains an `_errors` array:

```go
type ZodFormattedError map[string]any
// Structure: { "field": { "_errors": []string, "nested": { "_errors": []string } } }
```

### Nested Structure Example

```go
package main

import (
    "errors"
    "fmt"
    "github.com/kaptinlin/gozod"
)

func main() {
    schema := gozod.Object(gozod.ObjectSchema{
        "user": gozod.Object(gozod.ObjectSchema{
            "name": gozod.String().Min(2),
            "age":  gozod.Int().Min(18),
        }),
        "settings": gozod.Object(gozod.ObjectSchema{
            "theme": gozod.Enum("light", "dark"),
        }),
    })

    invalidData := map[string]any{
        "user": map[string]any{
            "name": "A",
            "age":  16,
        },
        "settings": map[string]any{
            "theme": "blue",
        },
    }

    _, err := schema.Parse(invalidData)
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        formatted := gozod.FormatError(zodErr)

        // Access nested errors
        userErrors := formatted["user"].(gozod.ZodFormattedError)
        nameErrors := userErrors["name"].(gozod.ZodFormattedError)["_errors"].([]string)
        ageErrors := userErrors["age"].(gozod.ZodFormattedError)["_errors"].([]string)
        
        settingsErrors := formatted["settings"].(gozod.ZodFormattedError)
        themeErrors := settingsErrors["theme"].(gozod.ZodFormattedError)["_errors"].([]string)
        
        fmt.Printf("Name errors: %v\n", nameErrors)
        fmt.Printf("Age errors: %v\n", ageErrors)
        fmt.Printf("Theme errors: %v\n", themeErrors)
    }
}
```

## PrettifyError

Converts ZodError into a human-readable string representation. Ideal for console output, logging, and debugging.

```go
func PrettifyError(err *ZodError) string
func PrettifyErrorWithMapper(err *ZodError, mapper func(ZodIssue) string) string
```

### Output Format

The prettified output uses a consistent format:
- `✖` symbol for each error
- `→ at path` to show field location
- Errors sorted by path length (root-level first)

```go
package main

import (
    "errors"
    "fmt"
    "github.com/kaptinlin/gozod"
)

func main() {
    schema := gozod.Object(gozod.ObjectSchema{
        "username": gozod.String().Min(3),
        "data":     gozod.Slice(gozod.Int()),
    })

    invalidData := map[string]any{
        "username": "ab",
        "data":     []any{1, "invalid", 3},
        "extra":    "field",
    }

    _, err := schema.Parse(invalidData)
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // Standard prettified output
        fmt.Println(gozod.PrettifyError(zodErr))
        
        // Custom formatting with mapper
        customPretty := gozod.PrettifyErrorWithMapper(zodErr, func(issue gozod.ZodIssue) string {
            return fmt.Sprintf("[%s] %s", issue.Code, issue.Message)
        })
        fmt.Println("\nCustom format:")
        fmt.Println(customPretty)
    }
}
```

## FlattenError

Converts ZodError into a flat structure with form-level and field-level errors. Useful for web form validation where errors need to be displayed next to specific form fields.

```go
func FlattenError(err *ZodError) *FlattenedError
func FlattenErrorWithMapper(err *ZodError, mapper func(ZodIssue) string) *FlattenedError
```

The `FlattenedError` type contains:

```go
type FlattenedError struct {
    FormErrors  []string            // Root-level errors (empty path)
    FieldErrors map[string][]string // Field-specific errors (keyed by first path segment)
}
```

### Form Validation Example

```go
package main

import (
    "errors"
    "fmt"
    "github.com/kaptinlin/gozod"
)

func main() {
    formSchema := gozod.Object(gozod.ObjectSchema{
        "username": gozod.String().Min(3).Max(20),
        "email":    gozod.String().Email(),
        "age":      gozod.Int().Min(18),
    })

    formData := map[string]any{
        "username": "ab",
        "email":    "invalid-email",
        "age":      16,
        "extra":    "field",
    }

    _, err := formSchema.Parse(formData)
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        flattened := gozod.FlattenError(zodErr)
        
        // Form-level errors (unexpected fields, etc.)
        if len(flattened.FormErrors) > 0 {
            fmt.Printf("Form errors: %v\n", flattened.FormErrors)
        }
        
        // Field-specific errors
        for field, errors := range flattened.FieldErrors {
            fmt.Printf("Field '%s': %v\n", field, errors)
        }
    }
}
```

## Custom Error Mappers

All formatting functions support custom mappers to transform error messages:

```go
package main

import (
    "errors"
    "fmt"
    "github.com/kaptinlin/gozod"
)

func main() {
    schema := gozod.Object(gozod.ObjectSchema{
        "username": gozod.String().Min(3),
        "email":    gozod.String().Email(),
    })
    
    invalidData := map[string]any{
        "username": "ab",
        "email":    "not-an-email",
    }
    
    _, err := schema.Parse(invalidData)
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // Custom error mapper
        customMapper := func(issue gozod.ZodIssue) string {
            switch issue.Code {
            case "too_small":
                minimum, _ := issue.GetMinimum()
                return fmt.Sprintf("Must be at least %v characters", minimum)
            case "invalid_format":
                return "Please enter a valid email address"
            default:
                return "Invalid input"
            }
        }
        
        // Apply mapper to different formats
        customFormatted := gozod.FormatErrorWithMapper(zodErr, customMapper)
        customFlattened := gozod.FlattenErrorWithMapper(zodErr, customMapper)
        customPrettified := gozod.PrettifyErrorWithMapper(zodErr, customMapper)
        
        fmt.Printf("Custom formatted: %v\n", customFormatted)
        fmt.Printf("Custom flattened: %v\n", customFlattened.FieldErrors)
        fmt.Printf("Custom prettified:\n%s\n", customPrettified)
    }
}
```

## Use Cases for Each Format

### FormatError - Use When:
- Building complex nested forms with deep object structures
- Creating React/Vue components that need field-specific errors
- APIs that return detailed validation feedback with full path information

### PrettifyError - Use When:
- Console applications and CLI tools
- Development debugging and logging
- Simple error notifications and alerts

### FlattenError - Use When:
- Simple flat forms with top-level fields
- Legacy APIs that expect flat error structures
- Cases where nested path information isn't needed

## Practical Examples

### Web API Error Response

```go
package main

import (
    "encoding/json"
    "errors"
    "net/http"
    "github.com/kaptinlin/gozod"
)

func handleUserCreation(w http.ResponseWriter, r *http.Request) {
    var userData map[string]any
    if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    userSchema := gozod.Object(gozod.ObjectSchema{
        "username": gozod.String().Min(3).Max(50),
        "email":    gozod.String().Email(),
        "age":      gozod.Int().Min(18).Optional(),
    })

    _, err := userSchema.Parse(userData)
    if err != nil {
        var zodErr *gozod.ZodError
        if errors.As(err, &zodErr) {
            // Use flattened errors for API response
            flattened := gozod.FlattenError(zodErr)
            
            errorResponse := map[string]any{
                "success": false,
                "message": "Validation failed",
                "errors":  flattened.FieldErrors,
            }
            
            if len(flattened.FormErrors) > 0 {
                errorResponse["formErrors"] = flattened.FormErrors
            }
            
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(errorResponse)
            return
        }
    }
    
    // Process valid data...
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]any{
        "success": true,
        "message": "User created successfully",
    })
}
```

### Form Field Mapping

```go
package main

import (
    "errors"
    "fmt"
    "github.com/kaptinlin/gozod"
)

type FormField struct {
    Name     string
    Value    string
    HasError bool
    Errors   []string
}

func processContactForm(formData map[string]any) ([]FormField, error) {
    schema := gozod.Object(gozod.ObjectSchema{
        "name":    gozod.String().Min(1),
        "email":   gozod.String().Email(),
        "message": gozod.String().Min(10),
    })
    
    fields := []FormField{
        {Name: "name", Value: fmt.Sprintf("%v", formData["name"])},
        {Name: "email", Value: fmt.Sprintf("%v", formData["email"])},
        {Name: "message", Value: fmt.Sprintf("%v", formData["message"])},
    }
    
    _, err := schema.Parse(formData)
    if err != nil {
        var zodErr *gozod.ZodError
        if errors.As(err, &zodErr) {
            flattened := gozod.FlattenError(zodErr)
            
            // Map errors to form fields
            for i := range fields {
                if fieldErrors, ok := flattened.FieldErrors[fields[i].Name]; ok {
                    fields[i].Errors = fieldErrors
                    fields[i].HasError = true
                }
            }
            
            return fields, err
        }
    }
    
    return fields, nil
}
```

### Custom Error Formatter

```go
package main

import (
    "github.com/kaptinlin/gozod"
)

func customErrorFormatter(zodErr *gozod.ZodError) map[string]any {
    result := map[string]any{
        "hasErrors":  true,
        "errorCount": len(zodErr.Issues),
        "errors":     make([]map[string]any, 0, len(zodErr.Issues)),
    }

    for _, issue := range zodErr.Issues {
        errorInfo := map[string]any{
            "code":    issue.Code,
            "message": issue.Message,
            "path":    gozod.ToDotPath(issue.Path),
        }
        
        // Add type-specific information
        switch issue.Code {
        case "invalid_type":
            if expected, ok := issue.GetExpected(); ok {
                errorInfo["expected"] = expected
            }
            if received, ok := issue.GetReceived(); ok {
                errorInfo["received"] = received
            }
        case "too_small":
            if minimum, ok := issue.GetMinimum(); ok {
                errorInfo["minimum"] = minimum
            }
        case "too_big":
            if maximum, ok := issue.GetMaximum(); ok {
                errorInfo["maximum"] = maximum
            }
        case "unrecognized_keys":
            if keys, ok := issue.GetKeys(); ok {
                errorInfo["unrecognizedKeys"] = keys
            }
        }
        
        result["errors"] = append(result["errors"].([]map[string]any), errorInfo)
    }
    
    return result
}
```

## Localized Error Messages

Use mapper functions to provide localized error messages:

```go
// Localized error mapper
func createLocalizedMapper(locale string) func(gozod.ZodIssue) string {
    messages := map[string]map[string]string{
        "en": {
            "invalid_type": "Invalid type: expected %s, received %s",
            "too_small":    "Too small: minimum value is %v",
            "invalid_email": "Invalid email address",
        },
        "zh": {
            "invalid_type": "类型错误：期望 %s，收到 %s", 
            "too_small":    "数值过小：最小值为 %v",
            "invalid_email": "邮箱地址无效",
        },
    }
    
    localeMessages := messages[locale]
    if localeMessages == nil {
        localeMessages = messages["en"] // fallback
    }
    
    return func(issue gozod.ZodIssue) string {
        switch issue.Code {
        case "invalid_type":
            if expected, ok := issue.GetExpected(); ok {
                if received, ok := issue.GetReceived(); ok {
                    return fmt.Sprintf(localeMessages["invalid_type"], expected, received)
                }
            }
        case "too_small":
            if minimum, ok := issue.GetMinimum(); ok {
                return fmt.Sprintf(localeMessages["too_small"], minimum)
            }
        case "invalid_format":
            return localeMessages["invalid_email"]
        }
        return issue.Message
    }
}

// Usage
chineseFormatted := gozod.FormatErrorWithMapper(validationErr, createLocalizedMapper("zh"))
chineseFlattened := gozod.FlattenErrorWithMapper(validationErr, createLocalizedMapper("zh"))
```

## Path Utilities

GoZod provides a utility function to convert error paths to dot notation:

```go
func ToDotPath(path []any) string
```

This function handles:
- String keys: `["user", "name"]` → `"user.name"`
- Array indices: `["items", 0]` → `"items[0]"`
- Mixed paths: `["users", 1, "profile", "email"]` → `"users[1].profile.email"`
- Special characters: `["user name"]` → `["user name"]`

```go
path1 := []any{"user", "profile", "email"}
fmt.Println(gozod.ToDotPath(path1)) // "user.profile.email"

path2 := []any{"items", 0, "name"}
fmt.Println(gozod.ToDotPath(path2)) // "items[0].name"

path3 := []any{"user name", "first"}
fmt.Println(gozod.ToDotPath(path3)) // `["user name"].first`
```

## Integration with Error Customization

Error formatting works with custom error messages:

```go
schema := gozod.Object(gozod.ObjectSchema{
    "username": gozod.String().Min(3, gozod.SchemaParams{
        Error: "Username must be at least 3 characters",
    }),
    "email": gozod.String().Email(gozod.SchemaParams{
        Error: "Please provide a valid email address",
    }),
})

_, err := schema.Parse(invalidData)
if err != nil {
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // All formatting functions use your custom error messages
        formatted := gozod.FormatError(zodErr)
        prettified := gozod.PrettifyError(zodErr)
        flattened := gozod.FlattenError(zodErr)
        
        // Or override with mappers
        customFormatted := gozod.FormatErrorWithMapper(zodErr, func(issue gozod.ZodIssue) string {
            return fmt.Sprintf("🔥 %s", issue.Message)
        })
    }
}
```

## Best Practices

### 1. Choose the Right Format

- **FormatError**: Complex nested forms, detailed API responses
- **PrettifyError**: Console output, logging, debugging
- **FlattenError**: Simple forms, legacy API compatibility

### 2. Handle Errors Gracefully

```go
_, err := schema.Parse(data)
if err != nil {
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // Handle validation errors
        flattened := gozod.FlattenError(zodErr)
        // Process flattened errors...
    } else {
        // Handle other errors
        log.Printf("Non-validation error: %v", err)
    }
}
```

### 3. Use Mappers for Consistency

Create reusable mapper functions for consistent error formatting across your application:

```go
func standardErrorMapper(issue gozod.ZodIssue) string {
    switch issue.Code {
    case "invalid_type":
        return "Please check the data type"
    case "too_small":
        return "Value is too small"
    case "invalid_format":
        return "Invalid format"
    default:
        return issue.Message
    }
}

// Use consistently across the application
formatted := gozod.FormatErrorWithMapper(err, standardErrorMapper)
flattened := gozod.FlattenErrorWithMapper(err, standardErrorMapper)
```

### 4. Consider User Experience

Tailor error formatting to your users:

```go
// For developers (detailed)
devMapper := func(issue gozod.ZodIssue) string {
    return fmt.Sprintf("[%s] %s (path: %s)", issue.Code, issue.Message, gozod.ToDotPath(issue.Path))
}

// For end users (simplified)
userMapper := func(issue gozod.ZodIssue) string {
    switch issue.Code {
    case "invalid_type":
        return "Please enter a valid value"
    case "too_small":
        return "This field is too short"
    default:
        return "Please check this field"
    }
}
```

Error formatting in GoZod provides flexible options for presenting validation errors in different contexts. Choose the appropriate formatting function based on your use case, and leverage custom mappers to create consistent, user-friendly error messages.
