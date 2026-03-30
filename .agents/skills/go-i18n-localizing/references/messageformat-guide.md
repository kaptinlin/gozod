# ICU MessageFormat Guide

Detailed guide to ICU MessageFormat v1 patterns supported by `kaptinlin/messageformat-go`.

## Contents
- Variable interpolation
- Pluralization
- Select (gender/category)
- Nested patterns
- Number formatting with #
- Custom formatters
- Common patterns

## Variable Interpolation

Simple replacement with `{variableName}`:

```json
{
  "greeting": "Hello, {name}!",
  "welcome": "Welcome to {city}, {name}"
}
```

```go
loc.Get("greeting", i18n.Vars{"name": "Alice"})
// Output: Hello, Alice!

loc.Get("welcome", i18n.Vars{"city": "Tokyo", "name": "Bob"})
// Output: Welcome to Tokyo, Bob
```

Variable names are case-sensitive. `{Name}` and `{name}` are different placeholders.

## Pluralization

ICU plural categories: `zero`, `one`, `two`, `few`, `many`, `other`. Available categories depend on the locale's CLDR plural rules.

### Basic Plural

```json
{
  "items": "{count, plural, one {# item} other {# items}}"
}
```

```go
loc.Get("items", i18n.Vars{"count": 1})   // "1 item"
loc.Get("items", i18n.Vars{"count": 5})   // "5 items"
```

### Exact Matches

Use `=N` for exact numeric values. Exact matches take priority over plural categories:

```json
{
  "notifications": "{count, plural, =0 {No notifications} =1 {One notification} other {# notifications}}"
}
```

```go
loc.Get("notifications", i18n.Vars{"count": 0})   // "No notifications"
loc.Get("notifications", i18n.Vars{"count": 1})   // "One notification"
loc.Get("notifications", i18n.Vars{"count": 99})  // "99 notifications"
```

### The `#` Placeholder

Inside plural branches, `#` is replaced with the numeric value:

```json
{
  "files": "{count, plural, =0 {No files} one {# file selected} other {# files selected}}"
}
```

`#` only works inside `plural` (and `selectordinal`) blocks. Outside of plural, use the variable name directly.

### Locale-Specific Plural Rules

Different languages have different plural categories. Always provide `other` as the catch-all:

**English**: `one`, `other`
**Arabic**: `zero`, `one`, `two`, `few`, `many`, `other`
**Chinese/Japanese/Korean**: `other` only (no grammatical plural)
**Russian/Polish**: `one`, `few`, `many`, `other`

```json
// Russian (ru.json)
{
  "items": "{count, plural, one {# товар} few {# товара} many {# товаров} other {# товаров}}"
}
```

## Select (Gender/Category)

Route output by a string variable:

```json
{
  "invite": "{gender, select, male {He invited you} female {She invited you} other {They invited you}}"
}
```

```go
loc.Get("invite", i18n.Vars{"gender": "female"})
// Output: She invited you

loc.Get("invite", i18n.Vars{"gender": "unknown"})
// Output: They invited you
```

`other` is required and serves as the default branch.

## Nested Patterns

Combine plural and select:

```json
{
  "activity": "{gender, select, male {{count, plural, one {He has # notification} other {He has # notifications}}} female {{count, plural, one {She has # notification} other {She has # notifications}}} other {{count, plural, one {They have # notification} other {They have # notifications}}}}"
}
```

```go
loc.Get("activity", i18n.Vars{"gender": "male", "count": 3})
// Output: He has 3 notifications
```

Keep nesting to two levels maximum for readability. For deeper logic, split into multiple keys.

## Custom Formatters

Register formatters when creating the bundle:

```go
bundle := i18n.NewBundle(
    i18n.WithDefaultLocale("en"),
    i18n.WithCustomFormatters(map[string]any{
        "upper": func(value any, locale string, arg *string) any {
            return strings.ToUpper(fmt.Sprintf("%v", value))
        },
        "currency": func(value any, locale string, arg *string) any {
            if f, ok := value.(float64); ok {
                return fmt.Sprintf("$%.2f", f)
            }
            return value
        },
    }),
)
```

Use in translation strings:

```json
{
  "price": "Total: {amount, currency}",
  "shout": "{text, upper}"
}
```

Formatter signature: `func(value any, locale string, arg *string) any`
- `value`: the variable value
- `locale`: current locale string (e.g., "en")
- `arg`: optional argument after formatter name (e.g., `{val, fmt, argString}`)

## Common Patterns

### Date-relative messages

```json
{
  "last_seen": "{days, plural, =0 {Online now} =1 {Last seen yesterday} other {Last seen # days ago}}"
}
```

### Item counts with context

```json
{
  "cart_summary": "{count, plural, =0 {Your cart is empty} one {# item in cart ({total, currency})} other {# items in cart ({total, currency})}}"
}
```

### Ordinals (selectordinal)

```json
{
  "place": "{position, selectordinal, one {#st place} two {#nd place} few {#rd place} other {#th place}}"
}
```

### Escaping Braces

Use single quotes to escape literal braces in MessageFormat:

```
"Value is '{count}'"   -> Output: Value is {count}
"It''s working"        -> Output: It's working
```

Single quote is the escape character. Use `''` to produce a literal single quote.
