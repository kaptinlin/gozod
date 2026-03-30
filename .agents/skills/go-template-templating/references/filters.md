# Filter Reference

Comprehensive filter documentation for `kaptinlin/template`. Filters transform values using the pipe `|` operator and can be chained.

```
{{ variable | filterOne | filterTwo:'arg' }}
```

Arguments use `:` after the filter name and `,` between arguments. String arguments use quotes.

```
{{ 'hello world' | replace:'world','there' }}
```

---

## String Filters

### default

Returns a fallback value if the input is falsy (empty string, nil, 0, false).

```
{{ userName | default:'Guest' }}
→ Guest (when userName is empty)

{{ userName | default:'Guest' }}
→ Alice (when userName is "Alice")
```

### strip / lstrip / rstrip

Remove whitespace. `strip` trims both sides, `lstrip` trims left, `rstrip` trims right.

```
{{ '  hello  ' | strip }}    → hello
{{ '  hello  ' | lstrip }}   → hello··
{{ '  hello  ' | rstrip }}   → ··hello
```

Aliases: `trim`, `trim_left`, `trim_right`

### upcase / downcase

Convert case. Aliases: `upper`, `lower`.

```
{{ 'hello' | upcase }}    → HELLO
{{ 'HELLO' | downcase }}  → hello
```

### capitalize

Capitalize first letter, lowercase the rest.

```
{{ 'hello world' | capitalize }}  → Hello world
```

### truncate

Shorten a string to a specified length including the ellipsis (default `...`). Default length is 50.

```
{{ text | truncate }}           → (first 50 chars)
{{ 'hello world' | truncate:5 }}         → he...
{{ 'hello world' | truncate:5,'--' }}    → hel--
```

### truncatewords

Truncate to a specified number of words. Default is 15 words. Default ellipsis is `...`.

```
{{ text | truncatewords }}                         → (first 15 words)
{{ 'hello beautiful world' | truncatewords:2 }}    → hello beautiful...
```

Alias: `truncate_words`

### replace / replace_first / replace_last

Replace occurrences of a substring.

```
{{ 'hello world' | replace:'world','there' }}    → hello there
{{ 'aabbcc' | replace_first:'b','x' }}           → aaxbcc
{{ 'aabbcc' | replace_last:'b','x' }}            → aabxcc
```

### remove / remove_first / remove_last

Remove occurrences of a substring.

```
{{ 'hello world' | remove:' world' }}      → hello
{{ 'abcabc' | remove_first:'b' }}          → acabc
{{ 'abcabc' | remove_last:'b' }}           → abcac
```

### append / prepend

Add characters to the end or beginning.

```
{{ 'hello' | append:' world' }}     → hello world
{{ 'world' | prepend:'hello ' }}    → hello world
```

### split

Split a string into an array by delimiter.

```
{{ 'one,two,three' | split:',' | size }}  → 3
```

### slice

Extract a substring by offset and optional length. Without length, returns one character.

```
{{ 'hello' | slice:1,3 }}  → ell
{{ 'hello' | slice:1 }}    → e
```

### escape / escape_once

Escape HTML special characters. `escape_once` avoids double-escaping.

```
{{ '<b>bold</b>' | escape }}           → &lt;b&gt;bold&lt;/b&gt;
{{ '&lt;b&gt;' | escape_once }}        → &lt;b&gt; (no double-escape)
```

Alias: `h` (for `escape`)

### strip_html / strip_newlines

Remove HTML tags or newline characters.

```
{{ '<p>Hello <b>World</b></p>' | strip_html }}  → Hello World
{{ 'hello\nworld\n' | strip_newlines }}          → helloworld
```

### url_encode / url_decode

Percent-encode or decode for URLs.

```
{{ 'hello world&foo=bar' | url_encode }}  → hello+world%26foo%3Dbar
{{ 'hello+world%26foo%3Dbar' | url_decode }}  → hello world&foo=bar
```

### base64_encode / base64_decode

Base64 encoding and decoding.

```
{{ 'hello world' | base64_encode }}   → aGVsbG8gd29ybGQ=
{{ 'aGVsbG8gd29ybGQ=' | base64_decode }}  → hello world
```

### titleize

Capitalize the first letter of each word.

```
{{ 'hello world' | titleize }}  → Hello World
```

### camelize / pascalize

Convert to camelCase or PascalCase.

```
{{ 'hello_world' | camelize }}   → helloWorld
{{ 'hello_world' | pascalize }}  → HelloWorld
```

### dasherize / slugify

Convert to dash-separated or URL-friendly slug.

```
{{ 'hello world' | dasherize }}            → hello-world
{{ 'Hello World & Friends' | slugify }}    → hello-world-and-friends
```

### pluralize

Select singular or plural form based on count.

```
{{ 1 | pluralize:'apple','apples' }}  → apple
{{ 2 | pluralize:'apple','apples' }}  → apples
```

### ordinalize

Convert number to ordinal form.

```
{{ 1 | ordinalize }}   → 1st
{{ 2 | ordinalize }}   → 2nd
{{ 3 | ordinalize }}   → 3rd
{{ 11 | ordinalize }}  → 11th
```

### length

Return length of string, array, or map. Works on all types.

```
{{ 'hello' | length }}  → 5
```

---

## Math Filters

### plus / minus / times / divided_by / modulo

Basic arithmetic. `divided_by` alias: `divide`.

```
{{ 7 | plus:3 }}         → 10
{{ 10 | minus:2 }}       → 8
{{ 5 | times:2 }}        → 10
{{ 20 | divided_by:4 }}  → 5
{{ 10 | modulo:3 }}      → 1
```

### abs

Absolute value.

```
{{ -42 | abs }}  → 42
```

### round

Round to specified decimal places. Default precision is 0.

```
{{ 3.7 | round }}      → 4
{{ 3.14159 | round:2 }}  → 3.14
```

### floor / ceil

Round down or up to nearest integer.

```
{{ 3.99 | floor }}  → 3
{{ 3.01 | ceil }}   → 4
```

### at_least / at_most

Clamp to minimum or maximum value.

```
{{ 5 | at_least:10 }}  → 10
{{ 15 | at_most:10 }}  → 10
```

---

## Array Filters

### join

Join elements into a string. Default separator is `" "` (space).

```
{{ value | join }}        → one two three
{{ value | join:'-' }}    → one-two-three
```

### first / last

First or last element.

```
{{ value | first }}  → first_element
{{ value | last }}   → last_element
```

### size

Length of array, string, or map.

```
{{ items | size }}     → 3
{{ 'hello' | size }}   → 5
```

### reverse

Reverse element order.

```
{{ value | reverse | join:',' }}  → 3,2,1
```

### sort / sort_natural

Sort ascending. `sort_natural` is case-insensitive. Optional property key for arrays of objects.

```
{{ value | sort | join:',' }}              → 1,2,3
{{ value | sort:'name' | map:'name' }}     → Alice,Bob,Charlie
{{ value | sort_natural | join:',' }}      → apple,Banana,Cherry
```

### uniq

Remove duplicates. Optional property key for arrays of objects.

```
{{ value | uniq | join:',' }}          → 1,2,3
{{ value | uniq:'role' | map:'name' }}  → Alice,Bob  (first of each role)
```

Alias: `unique`

### compact

Remove nil values. Optional property key.

```
{{ value | compact | join:',' }}  → 1,2,3
```

### concat

Combine two arrays.

```
{{ a | concat:b | join:',' }}  → 1,2,3,4,5,6
```

### map

Extract a key from each object.

```
{{ users | map:'name' | join:', ' }}  → John, Jane
```

### where / reject

Filter objects by key/value. `where` keeps matches; `reject` removes them. Without a value argument, filters by truthiness.

```
{{ users | where:'active','true' | map:'name' }}   → Alice,Charlie
{{ users | reject:'active','false' | map:'name' }}  → Alice,Charlie
```

### find / find_index

Find the first object matching key/value. `find` returns the object; `find_index` returns its index.

```
{{ users | find:'name','Bob' | extract:'age' }}  → 25
{{ users | find_index:'name','Bob' }}            → 1
```

### has

Check if any object matches key/value. Returns `true` or `false`.

```
{{ users | has:'name','Alice' }}    → true
{{ users | has:'name','Unknown' }}  → false
```

### sum

Sum numeric values. Optional property key for arrays of objects.

```
{{ scores | sum }}             → 6
{{ items | sum:'price' }}      → 55.0
```

### max / min / average

Aggregation over numeric arrays.

```
{{ scores | max }}      → 3.3
{{ scores | min }}      → 1.1
{{ scores | average }}  → 2
```

### shuffle / random

Randomly rearrange elements or pick one element.

```
{{ items | shuffle }}  → (random order)
{{ items | random }}   → (random element)
```

---

## Date Filters

**Important:** This engine uses PHP-style format specifiers, not Liquid's strftime.

### date

Format a timestamp.

```
{{ timestamp | date:'Y-m-d' }}      → 2024-03-30
{{ timestamp | date:'Y-m-d H:i' }}  → 2024-03-30 14:30
```

Common format specifiers: `Y` (year), `m` (month), `d` (day), `H` (24h hour), `i` (minutes), `s` (seconds), `F` (full month name), `l` (full weekday).

### day / month / year

Extract date components.

```
{{ timestamp | day }}    → 30
{{ timestamp | month }}  → 3
{{ timestamp | year }}   → 2024
```

### month_full / weekday

Full names.

```
{{ timestamp | month_full }}  → March
{{ timestamp | weekday }}     → Saturday
```

### week

ISO week number.

```
{{ timestamp | week }}  → 13
```

### time_ago

Human-readable relative time.

```
{{ pastDate | time_ago }}  → 2 days ago
```

Alias: `timeago`

---

## Number Filters

### number

Format a number with a pattern string.

```
{{ 1234567.89 | number:'#,###.##' }}  → 1,234,567.89
```

### bytes

Human-readable byte size.

```
{{ 2048 | bytes }}  → 2.0 KB
```

---

## Other Filters

### json

Serialize to JSON with deterministic key ordering.

```
{{ data | json }}  → {"key":"value"}
```

### extract

Access nested values using dot-separated key path.

```
{{ data | extract:'user.profile.age' }}  → 30
```
