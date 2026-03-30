# Error Handling Reference

## Contents
- ZodError structure and issue iteration
- Formatting functions (Flatten, Prettify, Treeify)
- Custom error messages (inline, function, schema-level, global)
- Error message precedence
- i18n / locale configuration
- API response patterns

## ZodError Structure

Every failed `Parse` / `StrictParse` returns `*gozod.ZodError` (wraps `error`):

```go
type ZodError struct {
    Issues []ZodRawIssue
}

type ZodRawIssue struct {
    Code    IssueCode  // "invalid_type", "too_small", etc.
    Message string     // Human-readable (locale-aware)
    Path    []any      // ["user", "email"] or ["items", 0, "name"]
    Input   any        // The invalid value
    // ... additional context via getter methods
}
```

### Check with errors.As (idiomatic Go)

```go
import "errors"

_, err := schema.Parse(input)
var zodErr *gozod.ZodError
if errors.As(err, &zodErr) {
    // handle validation errors
}
```

### Alternatively: gozod.IsZodError

```go
var zodErr *gozod.ZodError
if gozod.IsZodError(err, &zodErr) {
    // handle
}
```

## Three Formatting Functions

### FlattenError -- flat field-level errors (forms, API responses)

```go
flat := gozod.FlattenError(zodErr.Issues)

// flat.FormErrors   []string             -- root-level (no path) errors
// flat.FieldErrors  map[string][]string   -- keyed by first path segment

// JSON-friendly API response
type APIError struct {
    Success    bool                `json:"success"`
    FormErrors []string            `json:"formErrors,omitempty"`
    Fields     map[string][]string `json:"fieldErrors,omitempty"`
}
resp := APIError{
    Success:    false,
    FormErrors: flat.FormErrors,
    Fields:     flat.FieldErrors,
}
```

### PrettifyError -- human-readable string (logs, CLI)

```go
fmt.Println(gozod.PrettifyError(zodErr.Issues))
// Output:
// ZodError
// ✖ String must contain at least 3 character(s)
//   → at username
// ✖ Invalid email
//   → at email
```

### TreeifyError -- nested tree (complex UIs)

```go
tree := gozod.TreeifyError(zodErr.Issues)
// tree.Errors       []string              -- root errors
// tree.Properties   map[string]*TreeNode  -- nested by field name
// tree.Items        []*TreeNode           -- array element errors

nameErrs := tree.Properties["user"].Properties["name"].Errors
```

### Custom Mapper on Any Formatter

All three accept an optional mapper to transform issue messages:

```go
mapper := func(issue core.ZodRawIssue) string {
    return fmt.Sprintf("[%s] %s", issue.Code, issue.Message)
}
gozod.FlattenError(zodErr.Issues, mapper)
gozod.PrettifyError(zodErr.Issues, mapper)
gozod.TreeifyError(zodErr.Issues, mapper)
```

## Custom Error Messages

### Precedence (highest to lowest)

1. Schema-level error (inline string or function on the method)
2. Per-parse error (`ParseContext.Error`)
3. Global `CustomError` via `gozod.Config()`
4. Global `LocaleError` via locale config
5. Default English messages

### Inline String

```go
gozod.String().Email("Please enter a valid email")
gozod.String().Min(3, "Username too short")
gozod.Int().Min(18, "Must be 18+")
```

### Dynamic Error Function

```go
gozod.String().Min(8, func(issue gozod.ZodRawIssue) string {
    switch issue.Code {
    case "too_small":
        return fmt.Sprintf("Need %v+ characters, got %d",
            issue.Minimum(), len(fmt.Sprintf("%v", issue.Input)))
    case "invalid_type":
        return "Must be a string"
    default:
        return issue.Message
    }
})
```

### Schema-Level Error

```go
gozod.String(gozod.SchemaParams{Error: "Invalid name"})
gozod.Int(gozod.SchemaParams{
    Error: func(issue gozod.ZodRawIssue) string {
        return "Custom: " + issue.Message
    },
})
```

### Global Custom Error

```go
gozod.Config(&gozod.ZodConfig{
    CustomError: func(issue gozod.ZodRawIssue) string {
        switch issue.Code {
        case "invalid_type":
            return fmt.Sprintf("Expected %s, got %s", issue.Expected(), issue.Received())
        default:
            return issue.Message
        }
    },
})
```

## i18n / Locales

```go
import "github.com/kaptinlin/gozod/locales"

// Global locale change (affects all Parse calls)
gozod.Config(locales.ZhCN())

// Per-format locale override
formatter := locales.GetLocaleFormatter("zh-CN")
pretty := gozod.PrettifyErrorWithMapper(zodErr, formatter)
flat := gozod.FlattenErrorWithMapper(zodErr, formatter)
tree := gozod.TreeifyErrorWithMapper(zodErr, formatter)
```

## Issue Context Properties

Access additional context via getter methods on `ZodRawIssue`:

| Getter | Returns | Relevant Codes |
|--------|---------|---------------|
| `Expected()` | Expected type name | `invalid_type` |
| `Received()` | Received type name | `invalid_type` |
| `Minimum()` | Min constraint value | `too_small` |
| `Maximum()` | Max constraint value | `too_big` |
| `Inclusive()` | Whether bound is inclusive | `too_small`, `too_big` |
| `Format()` | Expected format name | `invalid_format` |
| `Pattern()` | Regex pattern string | `invalid_format` |
| `Keys()` | Unrecognized key list | `unrecognized_keys` |
| `Values()` | Allowed value set | `invalid_value` |
| `Divisor()` | Required divisor | `not_multiple_of` |

## Path Utilities

```go
path := []any{"user", "addresses", 0, "city"}
gozod.ToDotPath(path)         // "user.addresses[0].city"
gozod.FormatErrorPath(path, "dot")     // "user.addresses[0].city"
gozod.FormatErrorPath(path, "bracket") // `["user"]["addresses"][0]["city"]`
```

## HTTP API Pattern

```go
func handleCreate(w http.ResponseWriter, r *http.Request) {
    var input map[string]any
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    result, err := schema.Parse(input)
    if err != nil {
        var zodErr *gozod.ZodError
        if errors.As(err, &zodErr) {
            flat := gozod.FlattenError(zodErr.Issues)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnprocessableEntity)
            json.NewEncoder(w).Encode(map[string]any{
                "success": false,
                "errors":  flat.FieldErrors,
            })
            return
        }
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    // use result...
}
```
