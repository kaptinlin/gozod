# HTTP and Framework Integration

Patterns for integrating `kaptinlin/go-i18n` into HTTP services and frameworks.

## Contents
- Bundle initialization at startup
- Context-based middleware
- Extracting localizer from context
- net/http example
- Chi router example
- Gin example
- Error message localization
- JSON API responses

## Bundle Initialization

Create the bundle once at application startup:

```go
package main

import (
    "embed"
    "log"
    "net/http"

    "github.com/kaptinlin/go-i18n"
)

//go:embed locales/*.json
var localesFS embed.FS

var bundle *i18n.I18n

func init() {
    bundle = i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans", "ja"),
        i18n.WithFallback(map[string][]string{
            "zh-Hans": {"zh", "zh-Hant"},
        }),
    )
    if err := bundle.LoadFS(localesFS, "locales/*.json"); err != nil {
        log.Fatalf("load translations: %v", err)
    }
}
```

## Context Key Pattern

```go
type contextKey string

const localizerKey contextKey = "localizer"

func LocalizerFromContext(ctx context.Context) *i18n.Localizer {
    loc, ok := ctx.Value(localizerKey).(*i18n.Localizer)
    if !ok {
        return bundle.NewLocalizer(bundle.SupportedLanguages()[0].String())
    }
    return loc
}
```

## net/http Middleware

```go
func I18nMiddleware(bundle *i18n.I18n) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Priority: query param > cookie > Accept-Language header
            locale := r.URL.Query().Get("lang")
            if locale == "" {
                if c, err := r.Cookie("lang"); err == nil {
                    locale = c.Value
                }
            }
            if locale == "" {
                locale = bundle.MatchAvailableLocale(r.Header.Get("Accept-Language"))
            } else {
                locale = bundle.MatchAvailableLocale(locale)
            }

            loc := bundle.NewLocalizer(locale)
            ctx := context.WithValue(r.Context(), localizerKey, loc)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

Usage:

```go
mux := http.NewServeMux()
mux.HandleFunc("/api/hello", helloHandler)
http.ListenAndServe(":8080", I18nMiddleware(bundle)(mux))
```

## Handler Example

```go
func helloHandler(w http.ResponseWriter, r *http.Request) {
    loc := LocalizerFromContext(r.Context())

    msg := loc.Get("welcome", i18n.Vars{"name": "User"})

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": msg,
        "locale":  loc.Locale(),
    })
}
```

## Chi Router

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Use(I18nMiddleware(bundle))
r.Get("/api/hello", helloHandler)
```

## Gin

```go
func GinI18nMiddleware(bundle *i18n.I18n) gin.HandlerFunc {
    return func(c *gin.Context) {
        accept := c.GetHeader("Accept-Language")
        locale := bundle.MatchAvailableLocale(accept)
        loc := bundle.NewLocalizer(locale)
        c.Set("localizer", loc)
        c.Next()
    }
}

// In handler:
func helloHandler(c *gin.Context) {
    loc := c.MustGet("localizer").(*i18n.Localizer)
    c.JSON(200, gin.H{"message": loc.Get("hello")})
}
```

## Error Message Localization

Pattern for returning localized error messages from API endpoints:

```go
type AppError struct {
    Key    string         // translation key
    Vars   i18n.Vars      // template variables
    Status int            // HTTP status code
}

func (e *AppError) LocalizedMessage(loc *i18n.Localizer) string {
    return loc.Get(e.Key, e.Vars)
}

// Usage in handler:
func createHandler(w http.ResponseWriter, r *http.Request) {
    loc := LocalizerFromContext(r.Context())

    if err := validate(input); err != nil {
        appErr := &AppError{
            Key:    "validation_error",
            Vars:   i18n.Vars{"field": err.Field},
            Status: http.StatusBadRequest,
        }
        http.Error(w, appErr.LocalizedMessage(loc), appErr.Status)
        return
    }
}
```

Translation file:
```json
{
  "validation_error": "The field {field} is invalid",
  "not_found": "Resource not found",
  "unauthorized": "Please sign in to continue"
}
```

## JSON API Response Pattern

For APIs that return localized strings in structured JSON:

```go
type APIResponse struct {
    Data    any    `json:"data,omitempty"`
    Message string `json:"message"`
    Locale  string `json:"locale"`
}

func respondLocalized(w http.ResponseWriter, loc *i18n.Localizer, key string, vars ...i18n.Vars) {
    msg := loc.Get(key, vars...)
    json.NewEncoder(w).Encode(APIResponse{
        Message: msg,
        Locale:  loc.Locale(),
    })
}
```

## Locale Detection Priority

Recommended priority order for locale detection in web applications:

1. **URL query parameter** (`?lang=zh-Hans`) -- explicit user choice for current request
2. **Cookie** (`lang=zh-Hans`) -- persisted user preference
3. **User profile setting** -- stored in database
4. **Accept-Language header** -- browser/client preference
5. **Default locale** -- configured on the bundle

All candidate values should pass through `bundle.MatchAvailableLocale()` to normalize and validate against supported locales.
