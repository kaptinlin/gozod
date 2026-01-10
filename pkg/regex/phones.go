package regex

import "regexp"

// E164 matches E.164 international phone number standard
// Requires 7-15 digits total, first digit must be 1-9 (valid country codes don't start with 0)
// TypeScript Zod v4 code: /^\+[1-9]\d{6,14}$/
var E164 = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
