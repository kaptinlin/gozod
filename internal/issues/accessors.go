package issues

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// Expected returns the expected value from a raw issue.
func Expected(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "expected", "")
}

// Received returns the received value from a raw issue.
func Received(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "received", "")
}

// Origin returns the origin value from a raw issue.
func Origin(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "origin", "")
}

// Format returns the format value from a raw issue.
func Format(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "format", "")
}

// Pattern returns the pattern value from a raw issue.
func Pattern(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "pattern", "")
}

// Prefix returns the prefix value from a raw issue.
func Prefix(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "prefix", "")
}

// Suffix returns the suffix value from a raw issue.
func Suffix(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "suffix", "")
}

// Includes returns the includes value from a raw issue.
func Includes(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "includes", "")
}

// Minimum returns the minimum value from a raw issue.
func Minimum(r core.ZodRawIssue) any {
	return mapx.AnyOr(r.Properties, "minimum", nil)
}

// Maximum returns the maximum value from a raw issue.
func Maximum(r core.ZodRawIssue) any {
	return mapx.AnyOr(r.Properties, "maximum", nil)
}

// Inclusive returns the inclusive value from a raw issue.
func Inclusive(r core.ZodRawIssue) bool {
	return mapx.BoolOr(r.Properties, "inclusive", false)
}

// Divisor returns the divisor value from a raw issue.
func Divisor(r core.ZodRawIssue) any {
	return mapx.AnyOr(r.Properties, "divisor", nil)
}

// Keys returns the keys value from a raw issue.
func Keys(r core.ZodRawIssue) []string {
	return mapx.StringsOr(r.Properties, "keys", nil)
}

// Values returns the values from a raw issue's properties map.
func Values(r core.ZodRawIssue) []any {
	values := mapx.AnySliceOr(r.Properties, "values", nil)
	if values != nil {
		return values
	}

	if val, ok := mapx.Get(r.Properties, "values"); ok {
		if converted, err := slicex.ToAny(val); err == nil {
			return converted
		}
	}

	return nil
}

// IssueExpected returns the expected type for invalid_type issues.
func IssueExpected(i core.ZodIssue) (core.ZodTypeCode, bool) {
	if i.Code != core.InvalidType {
		return "", false
	}
	return i.Expected, i.Expected != ""
}

// IssueReceived returns the received type for invalid_type issues.
func IssueReceived(i core.ZodIssue) (core.ZodTypeCode, bool) {
	if i.Code != core.InvalidType {
		return "", false
	}
	return i.Received, i.Received != ""
}

// IssueMinimum returns the minimum value for too_small issues.
func IssueMinimum(i core.ZodIssue) (any, bool) {
	if i.Code != core.TooSmall {
		return nil, false
	}
	return i.Minimum, i.Minimum != nil
}

// IssueMaximum returns the maximum value for too_big issues.
func IssueMaximum(i core.ZodIssue) (any, bool) {
	if i.Code != core.TooBig {
		return nil, false
	}
	return i.Maximum, i.Maximum != nil
}

// IssueFormat returns the format for invalid_format issues.
func IssueFormat(i core.ZodIssue) (string, bool) {
	if i.Code != core.InvalidFormat {
		return "", false
	}
	return i.Format, i.Format != ""
}

// IssueDivisor returns the divisor for not_multiple_of issues.
func IssueDivisor(i core.ZodIssue) (any, bool) {
	if i.Code != core.NotMultipleOf {
		return nil, false
	}
	return i.Divisor, i.Divisor != nil
}

// Property returns any property value by key.
func Property(r core.ZodRawIssue, key string) any {
	return mapx.AnyOr(r.Properties, key, nil)
}

// StringProperty returns a string property by key.
func StringProperty(r core.ZodRawIssue, key string) string {
	return mapx.StringOr(r.Properties, key, "")
}

// BoolProperty returns a bool property by key.
func BoolProperty(r core.ZodRawIssue, key string) bool {
	return mapx.BoolOr(r.Properties, key, false)
}

// HasProperty checks if a property exists.
func HasProperty(r core.ZodRawIssue, key string) bool {
	return mapx.Has(r.Properties, key)
}
