# Liquid Standard Compatibility

Detailed compatibility between `kaptinlin/template` and [Shopify Liquid](https://shopify.github.io/liquid/) v5.x.

## Overview

| Metric | Count |
|--------|-------|
| Liquid standard filters | 46 |
| Fully compliant | 41 |
| Behavioral differences | 2 |
| Missing | 3 |
| Extension filters | 20 |
| Convenience aliases | 9 |

## Syntax Differences

| Feature | Liquid | This Engine |
|---------|--------|-------------|
| Variable output | `{{ variable }}` | `{{ variable }}` |
| Filters | `{{ value \| filter: arg1, arg2 }}` | `{{ value \| filter:arg1,arg2 }}` |
| Comments | `{% comment %}...{% endcomment %}` | `{# ... #}` |
| If/else | `{% if %}{% elsif %}{% endif %}` | `{% if %}{% elif %}{% endif %}` |
| For loop variable | `forloop.index` (1-based) | `loop.Index` (0-based) |

## Behavioral Differences

### `default` filter

| | Liquid | This Engine |
|---|--------|-------------|
| Signature | `default: fallback, allow_false: true` | `default:'fallback'` |
| `allow_false` param | When `true`, only `nil`/empty trigger fallback; `false` passes through | Not supported (no named parameter syntax). Uses Go/Django truthiness: empty string, `nil`, `0`, `false` all trigger fallback |

### `date` filter

| | Liquid | This Engine |
|---|--------|-------------|
| Format syntax | strftime (`%Y-%m-%d %H:%M`) | PHP-style (`Y-m-d H:i`) |
| `"now"` / `"today"` input | Supported | Not supported |

**Format specifier mapping:**

| Liquid (strftime) | This Engine (PHP) | Description |
|---|---|---|
| `%Y` | `Y` | 4-digit year |
| `%y` | `y` | 2-digit year |
| `%m` | `m` | 2-digit month (01-12) |
| `%d` | `d` | 2-digit day (01-31) |
| `%H` | `H` | 24-hour (00-23) |
| `%I` | `h` | 12-hour (01-12) |
| `%M` | `i` | Minutes (00-59) |
| `%S` | `s` | Seconds (00-59) |
| `%p` | `A` | AM/PM |
| `%A` | `l` | Full weekday name |
| `%a` | `D` | Abbreviated weekday |
| `%B` | `F` | Full month name |
| `%b` | `M` | Abbreviated month |

## Missing Filters

| Filter | Description | Workaround |
|--------|-------------|------------|
| `newline_to_br` | Replace `\n` with `<br />\n` | Use `replace` filter: `{{ text\|replace:'\n','<br />' }}` (partial) |
| `base64_url_safe_encode` | URL-safe Base64 (uses `-` and `_`) | Register custom filter using `encoding/base64.URLEncoding` |
| `base64_url_safe_decode` | Decode URL-safe Base64 | Register custom filter using `encoding/base64.URLEncoding` |

## Full Filter Compliance Checklist

### String Filters (29 in Liquid)

| # | Filter | Status |
|---|--------|--------|
| 1 | `append` | Compliant |
| 2 | `base64_decode` | Compliant |
| 3 | `base64_encode` | Compliant |
| 4 | `base64_url_safe_decode` | **Missing** |
| 5 | `base64_url_safe_encode` | **Missing** |
| 6 | `capitalize` | Compliant |
| 7 | `downcase` | Compliant |
| 8 | `escape` / `h` | Compliant |
| 9 | `escape_once` | Compliant |
| 10 | `lstrip` | Compliant |
| 11 | `newline_to_br` | **Missing** |
| 12 | `prepend` | Compliant |
| 13 | `remove` | Compliant |
| 14 | `remove_first` | Compliant |
| 15 | `remove_last` | Compliant |
| 16 | `replace` | Compliant |
| 17 | `replace_first` | Compliant |
| 18 | `replace_last` | Compliant |
| 19 | `rstrip` | Compliant |
| 20 | `slice` | Compliant |
| 21 | `split` | Compliant |
| 22 | `strip` | Compliant |
| 23 | `strip_html` | Compliant |
| 24 | `strip_newlines` | Compliant |
| 25 | `truncate` | Compliant (defaults to 50) |
| 26 | `truncatewords` | Compliant (defaults to 15) |
| 27 | `upcase` | Compliant |
| 28 | `url_decode` | Compliant |
| 29 | `url_encode` | Compliant |

### Math Filters (11 in Liquid)

| # | Filter | Status |
|---|--------|--------|
| 1 | `abs` | Compliant |
| 2 | `at_least` | Compliant |
| 3 | `at_most` | Compliant |
| 4 | `ceil` | Compliant |
| 5 | `divided_by` | Compliant |
| 6 | `floor` | Compliant |
| 7 | `minus` | Compliant |
| 8 | `modulo` | Compliant |
| 9 | `plus` | Compliant |
| 10 | `round` | Compliant (defaults to 0) |
| 11 | `times` | Compliant |

### Array Filters (18 in Liquid)

| # | Filter | Status |
|---|--------|--------|
| 1 | `compact` | Compliant |
| 2 | `concat` | Compliant |
| 3 | `find` | Compliant |
| 4 | `find_index` | Compliant |
| 5 | `first` | Compliant |
| 6 | `has` | Compliant |
| 7 | `join` | Compliant (defaults to `" "`) |
| 8 | `last` | Compliant |
| 9 | `map` | Compliant |
| 10 | `reject` | Compliant |
| 11 | `reverse` | Compliant |
| 12 | `size` | Compliant (works on strings, arrays, and maps) |
| 13 | `sort` | Compliant |
| 14 | `sort_natural` | Compliant |
| 15 | `sum` | Compliant (supports `sum:'property'`) |
| 16 | `uniq` | Compliant (supports `uniq:'property'`) |
| 17 | `where` | Compliant |
| 18 | `slice` | Compliant (shared with string) |

### Other Filters (2 in Liquid)

| # | Filter | Status |
|---|--------|--------|
| 1 | `date` | Behavioral difference (PHP-style format) |
| 2 | `default` | Behavioral difference (no `allow_false:` param) |

## Extension Filters (Not in Liquid Standard)

### String

`titleize`, `camelize`, `pascalize`, `dasherize`, `slugify`, `pluralize`, `ordinalize`, `length`

### Array

`random`, `shuffle`, `max`, `min`, `average`

### Date

`day`, `month`, `month_full`, `year`, `week`, `weekday`, `time_ago`

### Number

`number`, `bytes`

### Other

`json`, `extract`

## Convenience Aliases

| Alias | Liquid Name | Note |
|-------|-------------|------|
| `trim` | `strip` | |
| `trim_left` | `lstrip` | |
| `trim_right` | `rstrip` | |
| `upper` | `upcase` | |
| `lower` | `downcase` | |
| `h` | `escape` | Also standard Liquid alias |
| `truncate_words` | `truncatewords` | |
| `unique` | `uniq` | |
| `divide` | `divided_by` | |
| `timeago` | `time_ago` | Extension alias |
