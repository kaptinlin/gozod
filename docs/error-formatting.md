# Error Formatting

GoZod provides TypeScript Zod v4 compatible utilities to format validation errors for different use cases. This document covers error formatting functions that transform ZodError into structured, readable, or flat representations following Go language conventions.

## TypeScript Zod v4 Compatible Functions

GoZod offers three main error formatting utilities that match TypeScript Zod v4 patterns:

1. **`TreeifyError()`** - Creates nested error structure similar to `z.treeifyError()`
2. **`PrettifyError()`** - Generates human-readable string similar to `z.prettifyError()`
3. **`FlattenError()`** - Produces flat structure similar to `z.flattenError()`

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
        "username": "ab",                      // too short
        "numbers":  []any{1, "text", -5},      // mixed types and negative number
        "extra":    "not allowed",             // unexpected key
    }

    _, err := schema.Parse(invalidData)
    
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // 1. PrettifyError - Human-readable output
        pretty := gozod.PrettifyError(zodErr.Issues)
        fmt.Println("=== PrettifyError ===")
        fmt.Println(pretty)
        
        // 2. TreeifyError - Nested structure
        tree := gozod.TreeifyError(zodErr.Issues)
        fmt.Println("\n=== TreeifyError ===")
        fmt.Printf("Username errors: %v\n", tree.Properties["username"].Errors)
        
        // 3. FlattenError - Flat structure
        flattened := gozod.FlattenError(zodErr.Issues)
        fmt.Println("\n=== FlattenError ===")
        fmt.Printf("Field errors: %v\n", flattened.FieldErrors["username"])
    }
}
```

## TreeifyError

Converts ZodError into a nested structure that mirrors your schema's shape, similar to TypeScript Zod v4's `treeifyError()`:

```go
func TreeifyError(issues []core.ZodRawIssue, mapper ...func(core.ZodRawIssue) string) *TreeNode
```

The `TreeNode` structure:

```go
type TreeNode struct {
    Errors     []string               `json:"errors"`
    Properties map[string]*TreeNode   `json:"properties,omitempty"`
    Items      []*TreeNode            `json:"items,omitempty"`
}
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
        tree := gozod.TreeifyError(zodErr.Issues)

        // Access nested errors
        userErrors := tree.Properties["user"]
        nameErrors := userErrors.Properties["name"].Errors
        ageErrors := userErrors.Properties["age"].Errors
        
        settingsErrors := tree.Properties["settings"]
        themeErrors := settingsErrors.Properties["theme"].Errors
        
        fmt.Printf("Name errors: %v\n", nameErrors)
        fmt.Printf("Age errors: %v\n", ageErrors)
        fmt.Printf("Theme errors: %v\n", themeErrors)
    }
}
```

## PrettifyError

Converts ZodError into a human-readable string representation, similar to TypeScript Zod v4's `prettifyError()`:

```go
func PrettifyError(issues []core.ZodRawIssue, mapper ...func(core.ZodRawIssue) string) string
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
        fmt.Println(gozod.PrettifyError(zodErr.Issues))
        
        // Custom formatting with mapper
        customPretty := gozod.PrettifyError(zodErr.Issues, func(issue core.ZodRawIssue) string {
            return fmt.Sprintf("[%s] %s", issue.Code, issue.Message)
        })
        fmt.Println("\nCustom format:")
        fmt.Println(customPretty)
    }
}
```

## FlattenError

Converts ZodError into a flat structure with form-level and field-level errors, similar to TypeScript Zod v4's `flattenError()`:

```go
func FlattenError(issues []core.ZodRawIssue, mapper ...func(core.ZodRawIssue) string) *FlatError
```

The `FlatError` structure:

```go
type FlatError struct {
    FormErrors  []string            `json:"formErrors"`
    FieldErrors map[string][]string `json:"fieldErrors"`
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
        flattened := gozod.FlattenError(zodErr.Issues)
        
        // Form-level errors (unexpected fields, etc.)
        if len(flattened.FormErrors) > 0 {
            fmt.Printf("Form errors: %v\n", flattened.FormErrors)
        }
        
        // Field-specific errors
        for field, errs := range flattened.FieldErrors {
            fmt.Printf("Field '%s': %v\n", field, errs)
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
        customMapper := func(issue core.ZodRawIssue) string {
            switch issue.Code {
            case "too_small":
                minimum := issue.Minimum()
                return fmt.Sprintf("Must be at least %v characters", minimum)
            case "invalid_format":
                return "Please enter a valid email address"
            default:
                return "Invalid input"
            }
        }
        
        // Apply mapper to different formats
        customTree := gozod.TreeifyError(zodErr.Issues, customMapper)
        customFlat := gozod.FlattenError(zodErr.Issues, customMapper)
        customPretty := gozod.PrettifyError(zodErr.Issues, customMapper)
        
        fmt.Printf("Custom tree: %v\n", customTree)
        fmt.Printf("Custom flat: %v\n", customFlat.FieldErrors)
        fmt.Printf("Custom pretty:\n%s\n", customPretty)
    }
}
```

## Internationalization

All formatting functions work with localized error messages:

```go
import "github.com/kaptinlin/gozod/locales"

// Set global locale to Chinese
gozod.Config(locales.ZhCN())

// Format errors using Chinese formatter function
chineseFormatter := locales.GetLocaleFormatter("zh-CN")
chineseTree := gozod.TreeifyErrorWithMapper(zodErr, chineseFormatter)
chineseFlat := gozod.FlattenErrorWithMapper(zodErr, chineseFormatter) 
chinesePretty := gozod.PrettifyErrorWithMapper(zodErr, chineseFormatter)

fmt.Println(chinesePretty)
// Output will be in Chinese:
// ✖ 格式无效：请输入有效的电子邮件地址
//   → at email
```

## Path Utilities

GoZod provides utility functions for path formatting:

```go

// Convert path to dot notation
path1 := []any{"user", "profile", "email"}
fmt.Println(gozod.ToDotPath(path1)) // "user.profile.email"

path2 := []any{"items", 0, "name"}
fmt.Println(gozod.ToDotPath(path2)) // "items[0].name"

path3 := []any{"user name", "first"}
fmt.Println(gozod.ToDotPath(path3)) // `["user name"].first`

// Format path with different styles
fmt.Println(gozod.FormatErrorPath(path1, "dot"))     // "user.profile.email"
fmt.Println(gozod.FormatErrorPath(path1, "bracket")) // `["user"]["profile"]["email"]`
```

## Use Cases for Each Format

### TreeifyError - Use When:
- Building complex nested forms with deep object structures
- Creating React/Vue components that need field-specific errors
- APIs that return detailed validation feedback with full path information
- Need to programmatically access errors at specific paths

### PrettifyError - Use When:
- Console applications and CLI tools
- Development debugging and logging
- Simple error notifications and alerts
- Human-readable error output

### FlattenError - Use When:
- Simple flat forms with top-level fields
- Legacy APIs that expect flat error structures
- Cases where nested path information isn't needed
- Form libraries that expect field-level error maps

## Practical Examples

### Web API Error Response

```go
package main

import (
    "encoding/json"
    "errors"
    "net/http"
    "github.com/kaptinlin/gozod"
    "github.com/kaptinlin/gozod/internal/issues"
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
            flattened := gozod.FlattenError(zodErr.Issues)
            
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
            if expected := issue.Expected(); expected != "" {
                errorInfo["expected"] = expected
            }
            if received := issue.Received(); received != "" {
                errorInfo["received"] = received
            }
        case "too_small":
            if minimum := issue.Minimum(); minimum != nil {
                errorInfo["minimum"] = minimum
            }
        case "too_big":
            if maximum := issue.Maximum(); maximum != nil {
                errorInfo["maximum"] = maximum
            }
        case "unrecognized_keys":
            if keys := issue.Keys(); len(keys) > 0 {
                errorInfo["unrecognizedKeys"] = keys
            }
        }
        
        result["errors"] = append(result["errors"].([]map[string]any), errorInfo)
    }
    
    return result
}
```

## Integration with Error Customization

Error formatting works seamlessly with custom error messages:

```go
schema := gozod.Object(gozod.ObjectSchema{
    "username": gozod.String().Min(3, "Username must be at least 3 characters"),
    "email": gozod.String().Email("Please provide a valid email address"),
})

_, err := schema.Parse(invalidData)
if err != nil {
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // All formatting functions use your custom error messages
        tree := gozod.TreeifyError(zodErr.Issues)
        pretty := gozod.PrettifyError(zodErr.Issues)
        flat := gozod.FlattenError(zodErr.Issues)
        
        // Or override with mappers
        customFormatted := gozod.TreeifyError(zodErr.Issues, func(issue core.ZodRawIssue) string {
            return fmt.Sprintf("🔥 %s", issue.Message)
        })
    }
}
```

## Best Practices

### 1. Choose the Right Format

- **TreeifyError**: Complex nested forms, detailed API responses
- **PrettifyError**: Console output, logging, debugging
- **FlattenError**: Simple forms, legacy API compatibility

### 2. Handle Errors Gracefully

```go
_, err := schema.Parse(data)
if err != nil {
    var zodErr *gozod.ZodError
    if errors.As(err, &zodErr) {
        // Handle validation errors
        flattened := gozod.FlattenError(zodErr.Issues)
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
func standardErrorMapper(issue core.ZodRawIssue) string {
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
formatted := gozod.TreeifyError(zodErr.Issues, standardErrorMapper)
flattened := gozod.FlattenError(zodErr.Issues, standardErrorMapper)
```

### 4. Consider User Experience

Tailor error formatting to your users:

```go
// For developers (detailed)
devMapper := func(issue core.ZodRawIssue) string {
    return fmt.Sprintf("[%s] %s (path: %s)", 
        issue.Code, issue.Message, gozod.ToDotPath(issue.Path))
}

// For end users (simplified)
userMapper := func(issue core.ZodRawIssue) string {
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

Error formatting in GoZod provides flexible, TypeScript Zod v4 compatible options for presenting validation errors in different contexts. Choose the appropriate formatting function based on your use case, and leverage custom mappers to create consistent, user-friendly error messages that follow Go language conventions.