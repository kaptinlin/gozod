# GoZod Regular Expressions

A collection of common regular expressions for Go, ported from TypeScript Zod v4 regexes.

## Features

- **TypeScript Zod v4 Correspondence**: All regex patterns are directly ported from TypeScript Zod v4 with complete source code references
- **Standard identifiers**: UUID, GUID, CUID, ULID, XID, KSUID, NanoID
- **Email validation**: Multiple email validation patterns (practical, HTML5, RFC 5322, Unicode, browser)
- **Network validation**: IPv4, IPv6, CIDR, hostnames
- **ISO 8601 dates and times**: Duration, date, time, datetime with configurable precision and timezone support
- **Encoding validation**: Base64, Base64URL
- **Primitive type checking**: String, number, boolean, null, undefined patterns
- **Text patterns**: Emoji, case checking (uppercase/lowercase)
- **Phone numbers**: E.164 international format

## Design Principles

- **Complete TypeScript Reference**: Every regex includes the original TypeScript source code in comments
- **Go Documentation Standards**: All comments follow Go documentation conventions
- **Type Safety**: Leverages Go's compile-time type checking
- **No Over-Engineering**: Only includes patterns that exist in TypeScript Zod v4

## Usage Example

```go
package main

import (
	"fmt"
	
	"github.com/kaptinlin/gozod/regexes"
)

func main() {
	// Validate a UUID v4
	if regexes.UUID4.MatchString("f47ac10b-58cc-4372-a567-0e02b2c3d479") {
		fmt.Println("Valid UUID4")
	}
	
	// Validate an email using practical email pattern
	if regexes.Email.MatchString("user@example.com") {
		fmt.Println("Valid email")
	}
	
	// Validate IPv4 address
	if regexes.IPv4.MatchString("192.168.1.1") {
		fmt.Println("Valid IPv4 address")
	}
	
	// Create a length-restricted string pattern
	customStr := regexes.StringRegex(3, 10)
	if customStr.MatchString("hello") {
		fmt.Println("String length is between 3 and 10")
	}
	
	// Check ISO 8601 date format
	if regexes.Date.MatchString("2024-03-15") {
		fmt.Println("Valid ISO 8601 date")
	}
	
	// Validate E.164 phone number
	if regexes.E164.MatchString("+1234567890") {
		fmt.Println("Valid E.164 phone number")
	}
}
```

## File Organization

- `ids.go` - Unique identifiers (UUID, CUID, ULID, etc.)
- `emails.go` - Email validation patterns
- `networks.go` - Network-related patterns (IP, CIDR, hostnames)
- `datetimes.go` - ISO 8601 date and time patterns
- `encodings.go` - Base64 and URL-safe Base64 patterns
- `primitives.go` - Basic type patterns (string, number, boolean, etc.)
- `phones.go` - Phone number patterns
- `text.go` - Text-related patterns (emoji, JWT)
- `urls.go` - URL patterns (Go-specific utilities)

## TypeScript Zod v4 Compatibility

All patterns maintain semantic equivalence with their TypeScript counterparts while adapting to Go's regex engine limitations. Where Go's regex engine doesn't support certain Unicode properties (like `\p{Extended_Pictographic}`), simplified approximations are provided with clear documentation.
