# Validate

Tiny, type-safe helpers for runtime validation of numbers, strings, collections, and formats.

## Usage Example

```go
package main

import (
    "fmt"
    "regexp"
    "github.com/kaptinlin/gozod/pkg/validate"
)

type User struct {
    Email string
    Age   int
    Tags  []string
}

func main() {
    // Struct field validation
    u := User{Email: "alice@example.com", Age: 27, Tags: []string{"go", "dev"}}
    profileOK := validate.Email(u.Email) &&
        validate.Positive(u.Age) &&
        validate.MinSize(u.Tags, 1) &&
        validate.MaxSize(u.Tags, 5)
    fmt.Println("profile valid:", profileOK) // true

    // Map property validation (e.g. HTTP headers)
    headers := map[string]any{
        "X-Auth":  "aabbccddeeff00112233445566778899",
        "Version": "v1.2.3",
    }
    hex32 := regexp.MustCompile(`^[0-9a-f]{32}$`)
    if validate.Property(headers, "X-Auth", func(v any) bool {
        return validate.Regex(v, hex32)
    }) {
        fmt.Println("header OK")
    }

    // Numeric rules
    score := 80
    if validate.MultipleOf(score, 10) && validate.Lte(score, 100) {
        fmt.Println("score looks good")
    }
}
```

## Quick Reference

```go
import "github.com/kaptinlin/gozod/pkg/validate"

// Numeric comparisons
validate.Lt(3, 5)            // true
validate.Lte(10, 10)         // true
validate.Gt(a, b)            // a > b
validate.Gte(a, b)           // a >= b
validate.Positive(-1)        // false
validate.Negative(n)         // n < 0
validate.NonPositive(n)      // n <= 0
validate.NonNegative(n)      // n >= 0
validate.MultipleOf(10, 2)   // true

// Length & size
validate.MaxLength("hello", 10)         // true
validate.MinLength("go", 3)             // false
validate.Length("abc", 3)               // true
validate.Size([]int{1, 2, 3}, 3)        // true
validate.MinSize(map[string]int{"a":1}, 1) // true
validate.MaxSize([]string{"a", "b"}, 2)   // true

// String & regex
validate.Lowercase("abc")               // true
validate.Uppercase("ABC")               // true
validate.Includes("gopher", "go")        // true
validate.StartsWith("prefix", "pre")      // true
validate.EndsWith("suffix", "fix")        // true
validate.Regex("alice@example.com", regexes.Email) // true

// Formats & network
validate.Email("bob@example.com")        // true
validate.URL("https://example.com")      // true
validate.UUID("uuid-string")             // true/false
validate.IPv4("192.168.0.1")             // true
validate.CIDRv4("10.0.0.0/24")           // true
validate.ISODateTime("2024-06-12T08:04:00Z") // true
validate.ISODate("2024-06-12")           // true
validate.ISOTime("08:04:00Z")            // true
validate.ISODuration("P1DT2H")           // true

// Encoding & tokens
validate.Base64("aGVsbG8=")              // true
validate.Base64URL("aGVsbG8t")           // true
validate.JWT("eyJhbGciOiJI...")           // true/false

// Object property
headers := map[string]any{"X-Auth": "aabbccddeeff00112233445566778899"}
hex32 := regexp.MustCompile(`^[0-9a-f]{32}$`)
validate.Property(headers, "X-Auth", func(v any) bool {
    return validate.Regex(v, hex32)
}) // true

// MIME
validate.Mime("image/png", []string{"image/png", "image/jpeg"}) // true
```

## API Cheat-Sheet

| Category           | Functions                                                        |
|--------------------|------------------------------------------------------------------|
| Numeric compare    | `Lt` `Lte` `Gt` `Gte`                                            |
| Numeric sign       | `Positive` `Negative` `NonPositive` `NonNegative`                |
| Numeric arithmetic | `MultipleOf`                                                     |
| Length (str/slice) | `MaxLength` `MinLength` `Length`                                 |
| Size (slice/map)   | `MaxSize` `MinSize` `Size`                                       |
| String checks      | `Lowercase` `Uppercase` `Includes` `StartsWith` `EndsWith`       |
| Regex              | `Regex(value, pattern)`                                          |
| Formats            | `Email` `URL` `UUID` `GUID` `CUID` `CUID2` `NanoID` `ULID` ...   |
| Networks           | `IPv4` `IPv6` `CIDRv4` `CIDRv6`                                  |
| Encoding           | `Base64` `Base64URL`                                             |
| Phone & Tokens     | `E164` `JWT`                                                     |
| ISO                | `ISODateTime` `ISODate` `ISOTime` `ISODuration`                  |
| JSON               | `JSON`                                                           |
| Object             | `Property(obj, key, validator)`                                  |
| MIME               | `Mime(value, allowedTypes)`                                      |
