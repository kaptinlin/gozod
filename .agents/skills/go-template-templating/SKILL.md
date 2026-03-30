---
description: Render Liquid-style templates in Go using the kaptinlin/template engine with variable interpolation, filters, conditionals, loops, and custom tags. Use when writing Liquid templates, adding custom filters or tags, or building template rendering pipelines.
name: go-template-templating
---


# Go Templating with kaptinlin/template

A Liquid-compatible template engine for Go. Uses `{{ }}` for output, `{% %}` for tags, and `|` for filters.

```bash
go get github.com/kaptinlin/template
```

## Quick Start

```go
// One-step rendering
output, err := template.Render("Hello, {{ name|upcase }}!", map[string]any{
    "name": "alice",
})
// output: "Hello, ALICE!"

// Compile and reuse
tmpl, err := template.Compile("Hello, {{ name }}!")
output, err := tmpl.Render(map[string]any{"name": "World"})
```

## Liquid Compatibility Summary

This engine follows the [Liquid](https://shopify.github.io/liquid/) standard: **41 of 46 filters are fully compliant**. Read [references/liquid-compatibility.md](references/liquid-compatibility.md) for the full comparison.

### Key Differences from Liquid

| Area | Liquid | This Engine |
|------|--------|-------------|
| Comments | `{% comment %}...{% endcomment %}` | `{# ... #}` |
| `date` format | strftime `%Y-%m-%d` | PHP-style `Y-m-d` |
| `default` filter | Supports `allow_false:` named param | No named parameters |
| Filter args spacing | `{{ v \| filter: arg1, arg2 }}` | `{{ v \| filter:arg1,arg2 }}` (no spaces) |

### Missing Filters (3)

`newline_to_br`, `base64_url_safe_encode`, `base64_url_safe_decode`

### Convenience Aliases

Both forms work identically. Liquid names are preferred.

| Alias | Liquid Name |
|-------|-------------|
| `trim` | `strip` |
| `upper` | `upcase` |
| `lower` | `downcase` |
| `divide` | `divided_by` |
| `unique` | `uniq` |
| `length` | `size` (also extension: works on strings) |

## Template Syntax

### Variables

```
{{ user.name }}
{{ user.address.city }}
{{ items.0 }}
```

### Filters

Chain filters with `|`. Arguments use `:` and `,` separators.

```
{{ name|upcase }}
{{ title|truncate:20 }}
{{ name|downcase|capitalize }}
{{ price|plus:10|times:2 }}
{{ items|where:'active','true'|map:'name'|join:', ' }}
```

### Conditionals

```
{% if score > 80 %}
    Excellent
{% elif score > 60 %}
    Pass
{% else %}
    Fail
{% endif %}
```

Operators: `==`, `!=`, `<`, `>`, `<=`, `>=`, `and`/`&&`, `or`/`||`, `not`, `in`, `not in`.

### Loops

```
{% for item in items %}
    {{ loop.Index }}: {{ item }}
{% endfor %}

{% for key, value in dict %}
    {{ key }}: {{ value }}
{% endfor %}
```

Loop variables: `loop.Index`, `loop.Revindex`, `loop.First`, `loop.Last`, `loop.Length`.

Supports `{% break %}` and `{% continue %}`.

## Filter Reference

For detailed filter documentation with examples, see [references/filters.md](references/filters.md).

### String Filters (Liquid Standard)

| Filter | Description | Example |
|--------|-------------|---------|
| `default` | Default value if empty | `{{ name\|default:'Guest' }}` |
| `upcase` | Uppercase | `{{ name\|upcase }}` |
| `downcase` | Lowercase | `{{ name\|downcase }}` |
| `capitalize` | Capitalize first letter | `{{ name\|capitalize }}` |
| `strip` | Trim whitespace | `{{ text\|strip }}` |
| `lstrip` / `rstrip` | Trim left / right | `{{ text\|lstrip }}` |
| `truncate` | Truncate to length (default 50) | `{{ text\|truncate:20 }}` |
| `truncatewords` | Truncate to word count (default 15) | `{{ text\|truncatewords:5 }}` |
| `replace` | Replace all occurrences | `{{ text\|replace:'old','new' }}` |
| `replace_first` / `replace_last` | Replace first / last | `{{ text\|replace_first:'a','b' }}` |
| `remove` | Remove all occurrences | `{{ text\|remove:'bad' }}` |
| `remove_first` / `remove_last` | Remove first / last | `{{ text\|remove_first:'x' }}` |
| `append` / `prepend` | Add to end / start | `{{ name\|append:'!' }}` |
| `split` | Split by delimiter | `{{ csv\|split:',' }}` |
| `slice` | Substring by offset/length | `{{ text\|slice:1,3 }}` |
| `escape` | Escape HTML (`h` alias) | `{{ html\|escape }}` |
| `escape_once` | Escape without double-escaping | `{{ html\|escape_once }}` |
| `strip_html` | Remove HTML tags | `{{ html\|strip_html }}` |
| `strip_newlines` | Remove newlines | `{{ text\|strip_newlines }}` |
| `url_encode` / `url_decode` | URL encoding | `{{ text\|url_encode }}` |
| `base64_encode` / `base64_decode` | Base64 encoding | `{{ text\|base64_encode }}` |

### String Extensions (Non-Liquid)

| Filter | Description | Example |
|--------|-------------|---------|
| `titleize` | Title Case | `{{ text\|titleize }}` |
| `camelize` / `pascalize` | camelCase / PascalCase | `{{ name\|camelize }}` |
| `dasherize` / `slugify` | dash-case / url-slug | `{{ title\|slugify }}` |
| `pluralize` | Singular/plural by count | `{{ n\|pluralize:'item','items' }}` |
| `ordinalize` | Ordinal form | `{{ 1\|ordinalize }}` → `1st` |
| `length` | String/array/map length | `{{ name\|length }}` |

### Math Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `plus` / `minus` / `times` | Arithmetic | `{{ price\|plus:10 }}` |
| `divided_by` / `modulo` | Division / remainder | `{{ total\|divided_by:3 }}` |
| `abs` | Absolute value | `{{ num\|abs }}` |
| `round` | Round (default precision 0) | `{{ pi\|round:2 }}` |
| `floor` / `ceil` | Floor / ceiling | `{{ num\|floor }}` |
| `at_least` / `at_most` | Clamp min / max | `{{ num\|at_least:0 }}` |

### Array Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `join` | Join (default `" "`) | `{{ items\|join:', ' }}` |
| `first` / `last` | First / last element | `{{ items\|first }}` |
| `size` | Length (strings + arrays) | `{{ items\|size }}` |
| `reverse` | Reverse order | `{{ items\|reverse }}` |
| `sort` / `sort_natural` | Sort / case-insensitive sort | `{{ items\|sort }}` |
| `uniq` | Deduplicate (supports property) | `{{ items\|uniq:'role' }}` |
| `compact` | Remove nil values | `{{ items\|compact }}` |
| `concat` | Combine arrays | `{{ a\|concat:b }}` |
| `map` | Extract key from objects | `{{ users\|map:'name' }}` |
| `where` / `reject` | Filter by key/value | `{{ users\|where:'active','true' }}` |
| `find` / `find_index` | Find item / index | `{{ users\|find:'name','Bob' }}` |
| `has` | Check existence | `{{ users\|has:'name','Alice' }}` |
| `sum` | Sum (supports property) | `{{ items\|sum:'price' }}` |
| `shuffle` / `random` | Shuffle / random pick | `{{ items\|shuffle }}` |
| `max` / `min` / `average` | Aggregation | `{{ scores\|max }}` |

### Date Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `date` | Format date (PHP-style) | `{{ ts\|date:'Y-m-d' }}` |
| `day` / `month` / `year` | Extract parts | `{{ ts\|day }}` |
| `month_full` | Full month name | `{{ ts\|month_full }}` |
| `week` / `weekday` | ISO week / day name | `{{ ts\|weekday }}` |
| `time_ago` | Relative time | `{{ ts\|time_ago }}` |

### Other Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `number` | Format number | `{{ price\|number:'#,###.##' }}` |
| `bytes` | Human-readable bytes | `{{ size\|bytes }}` |
| `json` | Serialize to JSON | `{{ data\|json }}` |
| `extract` | Dot-path access | `{{ data\|extract:'user.name' }}` |

## Custom Filters

```go
template.RegisterFilter("repeat", func(value any, args ...any) (any, error) {
    s := fmt.Sprintf("%v", value)
    n := 2
    if len(args) > 0 {
        if parsed, err := strconv.Atoi(fmt.Sprintf("%v", args[0])); err == nil {
            n = parsed
        }
    }
    return strings.Repeat(s, n), nil
})
// {{ "ha"|repeat:3 }} → "hahaha"
```

**FilterFunc signature:** `func(value any, args ...any) (any, error)`

## Custom Tags

Register tags implementing the `Statement` interface:

```go
template.RegisterTag("set", func(doc *template.Parser, start *template.Token, arguments *template.Parser) (template.Statement, error) {
    varToken, err := arguments.ExpectIdentifier()
    if err != nil {
        return nil, arguments.Error("expected variable name after 'set'")
    }
    if arguments.Match(template.TokenSymbol, "=") == nil {
        return nil, arguments.Error("expected '=' after variable name")
    }
    expr, err := arguments.ParseExpression()
    if err != nil {
        return nil, err
    }
    return &SetNode{VarName: varToken.Value, Expression: expr, Line: start.Line, Col: start.Col}, nil
})
```

## Context Building

```go
// Map directly
output, _ := template.Render(source, map[string]any{"name": "Alice", "age": 30})

// ContextBuilder (supports struct expansion)
ctx, _ := template.NewContextBuilder().KeyValue("name", "Alice").Struct(user).Build()
output, _ := tmpl.Render(ctx)
```
