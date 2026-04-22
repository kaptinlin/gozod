package validate

import (
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/regex"
)

// Regex reports whether the string matches the pattern.
func Regex(value any, pattern *regexp.Regexp) bool {
	if pattern == nil {
		return false
	}
	return matchString(value, pattern)
}

// Lowercase reports whether the string contains no uppercase letters.
func Lowercase(value any) bool { return matchString(value, regex.Lowercase) }

// Uppercase reports whether the string contains no lowercase letters.
func Uppercase(value any) bool { return matchString(value, regex.Uppercase) }

// Includes reports whether the string contains the substring.
func Includes(value any, substring string) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return strings.Contains(str, substring)
}

// StartsWith reports whether the string starts with the prefix.
func StartsWith(value any, prefix string) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return strings.HasPrefix(str, prefix)
}

// EndsWith reports whether the string ends with the suffix.
func EndsWith(value any, suffix string) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return strings.HasSuffix(str, suffix)
}

// Email reports whether the string is a valid email format.
func Email(value any) bool { return matchString(value, regex.Email) }

// Base64 reports whether the string is valid Base64 encoding.
func Base64(value any) bool { return matchString(value, regex.Base64) }

// Base64URL reports whether the string is valid Base64URL encoding.
func Base64URL(value any) bool { return matchString(value, regex.Base64URL) }

// Hex reports whether the string is a valid hexadecimal string.
func Hex(value any) bool { return matchString(value, regex.Hex) }

// MD5Hex reports whether the string is a valid MD5 hash in hex format (32 chars).
func MD5Hex(value any) bool { return matchString(value, regex.MD5Hex) }

// SHA1Hex reports whether the string is a valid SHA-1 hash in hex format (40 chars).
func SHA1Hex(value any) bool { return matchString(value, regex.SHA1Hex) }

// SHA256Hex reports whether the string is a valid SHA-256 hash in hex format (64 chars).
func SHA256Hex(value any) bool { return matchString(value, regex.SHA256Hex) }

// SHA384Hex reports whether the string is a valid SHA-384 hash in hex format (96 chars).
func SHA384Hex(value any) bool { return matchString(value, regex.SHA384Hex) }

// SHA512Hex reports whether the string is a valid SHA-512 hash in hex format (128 chars).
func SHA512Hex(value any) bool { return matchString(value, regex.SHA512Hex) }

// E164 reports whether the string is a valid E.164 phone number format.
func E164(value any) bool { return matchString(value, regex.E164) }

// Emoji reports whether the string is a valid emoji.
func Emoji(value any) bool { return matchString(value, regex.Emoji) }
