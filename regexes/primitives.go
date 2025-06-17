package regexes

import (
	"fmt"
	"regexp"
)

// StringRegex returns a regex matching strings with optional min/max length limits
func StringRegex(minimum, maximum int) *regexp.Regexp {
	var pattern string
	if maximum > 0 {
		pattern = fmt.Sprintf(`^[\s\S]{%d,%d}$`, minimum, maximum)
	} else {
		pattern = fmt.Sprintf(`^[\s\S]{%d,}$`, minimum)
	}
	return regexp.MustCompile(pattern)
}

// String matches any string with no length restrictions
// TypeScript original code:
//
//	export const string = (params?: { minimum?: number | undefined; maximum?: number | undefined }): RegExp => {
//	  const regex = params ? `[\\s\\S]{${params?.minimum ?? 0},${params?.maximum ?? ""}}` : `[\\s\\S]*`;
//	  return new RegExp(`^${regex}$`);
//	};
var String = regexp.MustCompile(`^[\s\S]*$`)

// Bigint matches big integers
// TypeScript original code:
// export const bigint: RegExp = /^\d+n?$/;
var Bigint = regexp.MustCompile(`^\d+n?$`)

// Integer matches integers
// TypeScript original code:
// export const integer: RegExp = /^\d+$/;
var Integer = regexp.MustCompile(`^\d+$`)

// Number matches numbers including decimals and negative numbers
// TypeScript original code:
// export const number: RegExp = /^-?\d+(?:\.\d+)?/i;
var Number = regexp.MustCompile(`^-?\d+(?:\.\d+)?`)

// Boolean matches boolean values (true/false)
// TypeScript original code:
// export const boolean: RegExp = /true|false/i;
var Boolean = regexp.MustCompile(`(?i)^(true|false)$`)

// Null matches null values
// TypeScript original code:
// const _null: RegExp = /null/i;
// export { _null as null };
var Null = regexp.MustCompile(`(?i)^null$`)

// Undefined matches undefined values
// TypeScript original code:
// const _undefined: RegExp = /undefined/i;
// export { _undefined as undefined };
var Undefined = regexp.MustCompile(`(?i)^undefined$`)

// Lowercase matches strings with no uppercase letters
// TypeScript original code:
// export const lowercase: RegExp = /^[^A-Z]*$/;
var Lowercase = regexp.MustCompile(`^[^A-Z]*$`)

// Uppercase matches strings with no lowercase letters
// TypeScript original code:
// export const uppercase: RegExp = /^[^a-z]*$/;
var Uppercase = regexp.MustCompile(`^[^a-z]*$`)

// JSONString matches any valid JSON string format
// Note: This is a simplistic pattern; actual validation should be done at runtime
var JSONString = regexp.MustCompile(`^[\s\S]*$`)
