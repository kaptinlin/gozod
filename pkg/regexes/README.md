# GoZod Regular Expressions

A collection of common regular expressions for Go, ported from TypeScript Zod v4 regexes.

## Usage Example

```go
package main

import (
	"fmt"
	"regexp"
	"github.com/kaptinlin/gozod/pkg/regexes"
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

## Quick Reference

```go
import (
	"regexp"
	"github.com/kaptinlin/gozod/pkg/regexes"
)

// Identifiers
regexes.UUID4.MatchString(id)          // Validate UUID v4
regexes.CUID.MatchString(id)           // Validate CUID
regexes.NanoID.MatchString(id)         // Validate NanoID

// Networking
regexes.IPv4.MatchString("192.168.0.1")      // true
regexes.CIDRv4.MatchString("10.0.0.0/24")    // true

// Date & time (ISO-8601)
regexes.DateTime.MatchString("2024-06-12T08:04:00Z") // true
regexes.Date.MatchString("2024-03-15")              // true

// Custom length-restricted string
short := regexes.StringRegex(1, 5)
short.MatchString("four") // true

// Use pattern constants with regexp
hex32 := regexp.MustCompile(`^[0-9a-f]{32}$`)
hex32.MatchString("aabbccddeeff00112233445566778899")
```

## API Cheat-Sheet

| Category      | Patterns & Functions                | Notes                       |
|---------------|-------------------------------------|-----------------------------|
| Identifiers   | `UUID4` `CUID` `NanoID` `ULID` ...  | Unique IDs                  |
| Email         | `Email`                             | Practical email validation  |
| Networking    | `IPv4` `IPv6` `CIDRv4` `CIDRv6`     | IP & CIDR                   |
| Date & Time   | `Date` `DateTime` `Time`            | ISO-8601 formats            |
| Encoding      | `Base64` `Base64URL`                | Base64, URL-safe            |
| Phone         | `E164`                              | E.164 phone numbers         |
| Text          | `Emoji` `JWT`                       | Emoji, JWT                  |
| URL           | `URL`                               | URL validation              |
| Utility       | `StringRegex(min, max)`             | Custom string length        |