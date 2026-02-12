package issues

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// RawIssueExpected returns the expected value from a raw issue.
func RawIssueExpected(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "expected", "")
}

// RawIssueReceived returns the received value from a raw issue.
func RawIssueReceived(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "received", "")
}

// RawIssueOrigin returns the origin value from a raw issue.
func RawIssueOrigin(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "origin", "")
}

// RawIssueFormat returns the format value from a raw issue.
func RawIssueFormat(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "format", "")
}

// RawIssuePattern returns the pattern value from a raw issue.
func RawIssuePattern(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "pattern", "")
}

// RawIssuePrefix returns the prefix value from a raw issue.
func RawIssuePrefix(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "prefix", "")
}

// RawIssueSuffix returns the suffix value from a raw issue.
func RawIssueSuffix(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "suffix", "")
}

// RawIssueIncludes returns the includes value from a raw issue.
func RawIssueIncludes(r core.ZodRawIssue) string {
	return mapx.StringOr(r.Properties, "includes", "")
}

// RawIssueMinimum returns the minimum value from a raw issue.
func RawIssueMinimum(r core.ZodRawIssue) any {
	return mapx.AnyOr(r.Properties, "minimum", nil)
}

// RawIssueMaximum returns the maximum value from a raw issue.
func RawIssueMaximum(r core.ZodRawIssue) any {
	return mapx.AnyOr(r.Properties, "maximum", nil)
}

// RawIssueInclusive returns the inclusive value from a raw issue.
func RawIssueInclusive(r core.ZodRawIssue) bool {
	return mapx.BoolOr(r.Properties, "inclusive", false)
}

// RawIssueDivisor returns the divisor value from a raw issue.
func RawIssueDivisor(r core.ZodRawIssue) any {
	return mapx.AnyOr(r.Properties, "divisor", nil)
}

// RawIssueKeys returns the keys value from a raw issue.
func RawIssueKeys(r core.ZodRawIssue) []string {
	return mapx.StringsOr(r.Properties, "keys", nil)
}

// RawIssueValues returns the values from a raw issue's properties map.
func RawIssueValues(r core.ZodRawIssue) []any {
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

// RawIssueProperty returns any property value by key.
func RawIssueProperty(r core.ZodRawIssue, key string) any {
	return mapx.AnyOr(r.Properties, key, nil)
}

// RawIssueStringProperty returns a string property by key.
func RawIssueStringProperty(r core.ZodRawIssue, key string) string {
	return mapx.StringOr(r.Properties, key, "")
}

// RawIssueBoolProperty returns a bool property by key.
func RawIssueBoolProperty(r core.ZodRawIssue, key string) bool {
	return mapx.BoolOr(r.Properties, key, false)
}

// HasRawIssueProperty checks if a property exists.
func HasRawIssueProperty(r core.ZodRawIssue, key string) bool {
	return mapx.Has(r.Properties, key)
}
