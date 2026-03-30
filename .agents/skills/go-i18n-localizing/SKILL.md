---
description: Localize Go applications using kaptinlin/go-i18n with ICU MessageFormat support. Use when adding internationalization, loading translation files, formatting pluralized or interpolated messages, configuring locale fallback chains, or integrating i18n into HTTP services.
name: go-i18n-localizing
---


# Go Localization with kaptinlin/go-i18n

Localize Go applications using `github.com/kaptinlin/go-i18n` with ICU MessageFormat v1 via `github.com/kaptinlin/messageformat-go`.

## Decision Flowchart

```
How should translations reach your application?
|
+- Static files shipped with binary (deploy-safe, no runtime I/O)
|  +- go:embed + bundle.LoadFS(localesFS, "locales/*.json")
|
+- Files on disk (easy to update without recompile)
|  +- bundle.LoadGlob("./locales/*.json")
|  +- Or bundle.LoadFiles("./locales/en.json", ...)
|
+- Programmatic / database-driven
|  +- bundle.LoadMessages(map[locale]map[key]text)
|
How are translations formatted?
|
+- Simple key-value (no variables)
|  +- localizer.Get("hello_world")
|
+- Variables / plurals (ICU MessageFormat)
|  +- localizer.Get("messages", i18n.Vars{"count": 5})
|
+- sprintf-style formatting
|  +- localizer.Getf("welcome_%s", username)
|
+- Dynamic messages not in files
   +- localizer.Format("{count, plural, one {# item} other {# items}}", i18n.Vars{"count": n})
```

## Architecture Overview

```
+-------------------------------------------------------+
| Application                                            |
|                                                        |
|  I18n (Bundle)                                         |
|  +- manages locales, fallbacks, unmarshaler            |
|  +- pre-compiles MessageFormat templates on load       |
|  +- stores: map[locale]map[key]*parsedTranslation      |
|                                                        |
|  Localizer (per-locale)                                |
|  +- Get(key, vars)    token-based lookup               |
|  +- GetX(key, ctx)    context-disambiguated lookup     |
|  +- Getf(key, args)   sprintf-style                    |
|  +- Format(msg, vars) direct MessageFormat compile     |
|                                                        |
|  Loading        |  Fallback Chain   |  Accept-Language  |
|  LoadFiles      |  locale -> fb1 -> |  MatchAvailable-  |
|  LoadGlob       |  fb2 -> default   |  Locale(header)   |
|  LoadFS         |                   |                   |
|  LoadMessages   |                   |                   |
+-------------------------------------------------------+
| kaptinlin/messageformat-go/v1  (ICU MessageFormat)     |
| golang.org/x/text/language     (locale matching)       |
+-------------------------------------------------------+
```

## Quick Setup

```bash
go get github.com/kaptinlin/go-i18n@latest
```

### Minimal Example

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/go-i18n"
)

func main() {
    bundle := i18n.NewBundle(
        i18n.WithDefaultLocale("en"),
        i18n.WithLocales("en", "zh-Hans"),
    )
    bundle.LoadGlob("./locales/*.json")

    loc := bundle.NewLocalizer("zh-Hans")
    fmt.Println(loc.Get("hello", i18n.Vars{"name": "World"}))
}
```

**locales/en.json**
```json
{
  "hello": "Hello, {name}",
  "messages": "{count, plural, =0 {No messages} one {1 message} other {# messages}}"
}
```

### Embedded Translations (recommended for production)

```go
//go:embed locales/*.json
var localesFS embed.FS

bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithLocales("en", "zh-Hans", "ja"),
)
bundle.LoadFS(localesFS, "locales/*.json")
```

## Bundle Configuration Options

| Option | Purpose |
|--------|---------|
| `WithDefaultLocale(locale)` | Final fallback locale (default: first in WithLocales, or English) |
| `WithLocales(locales...)` | Declare supported locales |
| `WithFallback(map[string][]string)` | Custom fallback chains per locale |
| `WithUnmarshaler(fn)` | Custom file format (YAML, TOML, INI) |
| `WithMessageFormatOptions(opts)` | Configure MessageFormat engine |
| `WithCustomFormatters(map[string]any)` | Add custom format functions |
| `WithStrictMode(true)` | Fail on MessageFormat parse errors |

## Loading Translations

| Method | Input | Use When |
|--------|-------|----------|
| `LoadFiles(paths...)` | Explicit file paths | You know exact files |
| `LoadGlob(patterns...)` | Glob patterns | Load all `*.json` in a directory |
| `LoadFS(fsys, patterns...)` | `fs.FS` + patterns | `go:embed` compiled-in translations |
| `LoadMessages(map)` | `map[locale]map[key]string` | Programmatic / database / tests |

Filenames determine locale: `zh-Hans.json`, `zh_CN.user.json`, `ZH_CN.json` all normalize to `zh-Hans`. Multiple files for the same locale are merged.

## Translation Lookup

### Token-based (standard i18n)

```go
loc.Get("hello_world")                              // simple key
loc.Get("greeting", i18n.Vars{"name": "Alice"})     // with variables
loc.Getf("welcome_%s", username)                     // sprintf-style
```

### Text-based (key = English sentence, acts as fallback)

```go
loc.Get("I'm fine.")           // returns translation or "I'm fine." if missing
loc.Get("Hello, {name}", i18n.Vars{"name": "World"})  // key is also a template
```

### Context Disambiguation

When the same word has different meanings, append ` <context>` suffix in translation files:

```json
{
  "Post <verb>": "Publish",
  "Post <noun>": "Article"
}
```

```go
loc.GetX("Post", "verb")   // "Publish"
loc.GetX("Post", "noun")   // "Article"
```

### Direct MessageFormat (bypass translation files)

```go
result, err := loc.Format(
    "{count, plural, =0 {empty} one {# item} other {# items}}",
    i18n.Vars{"count": 42},
)
```

## ICU MessageFormat -- [details](references/messageformat-guide.md)

Core patterns supported via `kaptinlin/messageformat-go/v1`:

| Pattern | Syntax | Example |
|---------|--------|---------|
| Variable | `{name}` | `Hello, {name}` |
| Plural | `{var, plural, ...}` | `{count, plural, one {# item} other {# items}}` |
| Select | `{var, select, ...}` | `{gender, select, male {He} female {She} other {They}}` |
| Exact match | `=N` in plural | `{count, plural, =0 {none} one {one} other {#}}` |
| `#` placeholder | Inserts the plural value | `{count, plural, other {# items}}` |
| Custom formatter | `{var, formatter}` | `{name, upper}` (with custom formatter registered) |

## Fallback Chains

```go
bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithLocales("en", "zh-Hans", "zh", "zh-Hant", "en-GB", "en-US"),
    i18n.WithFallback(map[string][]string{
        "zh-Hans": {"zh", "zh-Hant"},
        "en-GB":   {"en-US"},
    }),
)
```

Lookup path: `zh-Hans -> zh -> zh-Hant -> en` (default is always final fallback). Recursive fallbacks are supported and cycle-safe.

Missing translation behavior: if a key is not found in any locale, the key itself is returned as the output.

## Accept-Language Parsing

```go
accept := r.Header.Get("Accept-Language")
locale := bundle.MatchAvailableLocale(accept)
loc := bundle.NewLocalizer(locale)
```

`MatchAvailableLocale` uses `golang.org/x/text/language.Matcher` for confidence-based matching against configured locales.

## Custom Unmarshalers

Default is JSON. Switch to YAML, TOML, or INI:

```go
// YAML
import "gopkg.in/yaml.v3"
i18n.WithUnmarshaler(yaml.Unmarshal)

// TOML
import "github.com/pelletier/go-toml/v2"
i18n.WithUnmarshaler(toml.Unmarshal)
```

INI requires a custom adapter function. See [README examples](https://github.com/kaptinlin/go-i18n#ini-unmarshaler).

Note: nested translation keys are not supported. Use flat keys like `"section.button"` quoted in YAML/TOML.

## HTTP Integration -- [details](references/http-integration.md)

Middleware pattern for HTTP services:

```go
func I18nMiddleware(bundle *i18n.I18n) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            accept := r.Header.Get("Accept-Language")
            locale := bundle.MatchAvailableLocale(accept)
            loc := bundle.NewLocalizer(locale)
            ctx := context.WithValue(r.Context(), localizerKey, loc)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Custom Formatters

```go
bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithCustomFormatters(map[string]any{
        "upper": func(value any, locale string, arg *string) any {
            return strings.ToUpper(fmt.Sprintf("%v", value))
        },
    }),
)

loc := bundle.NewLocalizer("en")
result, _ := loc.Format("Hello, {name, upper}!", i18n.Vars{"name": "world"})
// Output: Hello, WORLD!
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Creating a new `Bundle` per request | Re-parses and re-compiles all templates | Create bundle once at startup, share across requests |
| Nested keys in translation files | Library expects flat `map[string]string` | Use flat keys: `"section.key"` |
| Hardcoding locale strings everywhere | Brittle, easy to typo | Use `MatchAvailableLocale` or constants |
| Ignoring `LoadFiles`/`LoadGlob` errors | Missing translations silently | Always check and handle load errors |
| Using `Format` for stored translations | Recompiles MessageFormat each call | Use `Get` for translations loaded via files; `Format` is for dynamic/ad-hoc messages only |
| Mixing unmarshalers with wrong file format | Unmarshal errors at load time | Match `WithUnmarshaler` to your file extension |
| Forgetting `WithLocales` | Only the default locale is matched | Always declare all supported locales |
| Using `Get` return value as error signal | Returns key on miss, not empty string | Check translations at startup or enable strict mode |
