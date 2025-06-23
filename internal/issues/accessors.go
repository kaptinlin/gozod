package issues

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// SIMPLE PROPERTY ACCESSORS FOR ZodRawIssue
// =============================================================================

// GetRawIssueExpected returns the expected value
func GetRawIssueExpected(r core.ZodRawIssue) string {
	return mapx.GetStringDefault(r.Properties, "expected", "")
}

// GetRawIssueReceived returns the received value
func GetRawIssueReceived(r core.ZodRawIssue) string {
	return mapx.GetStringDefault(r.Properties, "received", "")
}

// GetRawIssueOrigin returns the origin value
func GetRawIssueOrigin(r core.ZodRawIssue) string {
	return mapx.GetStringDefault(r.Properties, "origin", "")
}

// GetRawIssueFormat returns the format value
func GetRawIssueFormat(r core.ZodRawIssue) string {
	return mapx.GetStringDefault(r.Properties, "format", "")
}

// GetRawIssuePattern returns the pattern value
func GetRawIssuePattern(r core.ZodRawIssue) string {
	return mapx.GetStringDefault(r.Properties, "pattern", "")
}

// GetRawIssuePrefix returns the prefix value
func GetRawIssuePrefix(r core.ZodRawIssue) string {
	return mapx.GetStringDefault(r.Properties, "prefix", "")
}

// GetRawIssueSuffix returns the suffix value
func GetRawIssueSuffix(r core.ZodRawIssue) string {
	return mapx.GetStringDefault(r.Properties, "suffix", "")
}

// GetRawIssueIncludes returns the includes value
func GetRawIssueIncludes(r core.ZodRawIssue) string {
	return mapx.GetStringDefault(r.Properties, "includes", "")
}

// GetRawIssueMinimum returns the minimum value
func GetRawIssueMinimum(r core.ZodRawIssue) any {
	return mapx.GetAnyDefault(r.Properties, "minimum", nil)
}

// GetRawIssueMaximum returns the maximum value
func GetRawIssueMaximum(r core.ZodRawIssue) any {
	return mapx.GetAnyDefault(r.Properties, "maximum", nil)
}

// GetRawIssueInclusive returns the inclusive value
func GetRawIssueInclusive(r core.ZodRawIssue) bool {
	return mapx.GetBoolDefault(r.Properties, "inclusive", false)
}

// GetRawIssueDivisor returns the divisor value
func GetRawIssueDivisor(r core.ZodRawIssue) any {
	return mapx.GetAnyDefault(r.Properties, "divisor", nil)
}

// GetRawIssueKeys returns the keys value
func GetRawIssueKeys(r core.ZodRawIssue) []string {
	return mapx.GetStringsDefault(r.Properties, "keys", nil)
}

// GetRawIssueValues returns the values from properties map
func GetRawIssueValues(r core.ZodRawIssue) []any {
	values := mapx.GetAnySliceDefault(r.Properties, "values", nil)
	if values != nil {
		return values
	}

	// Try to convert using slicex if direct access failed
	if val, ok := mapx.Get(r.Properties, "values"); ok {
		if converted, err := slicex.ToAny(val); err == nil {
			return converted
		}
	}

	return nil
}

// =============================================================================
// ISSUE-SPECIFIC ACCESSORS
// =============================================================================

// GetIssueExpected returns the expected type for invalid_type issues
func GetIssueExpected(i core.ZodIssue) (core.ZodTypeCode, bool) {
	if i.Code != core.InvalidType {
		return "", false
	}
	return i.Expected, i.Expected != ""
}

// GetIssueReceived returns the received type for invalid_type issues
func GetIssueReceived(i core.ZodIssue) (core.ZodTypeCode, bool) {
	if i.Code != core.InvalidType {
		return "", false
	}
	return i.Received, i.Received != ""
}

// GetIssueMinimum returns the minimum value for too_small issues
func GetIssueMinimum(i core.ZodIssue) (any, bool) {
	if i.Code != core.TooSmall {
		return nil, false
	}
	return i.Minimum, i.Minimum != nil
}

// GetIssueMaximum returns the maximum value for too_big issues
func GetIssueMaximum(i core.ZodIssue) (any, bool) {
	if i.Code != core.TooBig {
		return nil, false
	}
	return i.Maximum, i.Maximum != nil
}

// GetIssueFormat returns the format for invalid_format issues
func GetIssueFormat(i core.ZodIssue) (string, bool) {
	if i.Code != core.InvalidFormat {
		return "", false
	}
	return i.Format, i.Format != ""
}

// GetIssueDivisor returns the divisor for not_multiple_of issues
func GetIssueDivisor(i core.ZodIssue) (any, bool) {
	if i.Code != core.NotMultipleOf {
		return nil, false
	}
	return i.Divisor, i.Divisor != nil
}

// =============================================================================
// GENERIC PROPERTY ACCESSORS
// =============================================================================

// GetRawIssueProperty returns any property value
func GetRawIssueProperty(r core.ZodRawIssue, key string) any {
	return mapx.GetAnyDefault(r.Properties, key, nil)
}

// GetRawIssueStringProperty returns a string property
func GetRawIssueStringProperty(r core.ZodRawIssue, key string) string {
	return mapx.GetStringDefault(r.Properties, key, "")
}

// GetRawIssueBoolProperty returns a bool property
func GetRawIssueBoolProperty(r core.ZodRawIssue, key string) bool {
	return mapx.GetBoolDefault(r.Properties, key, false)
}

// HasRawIssueProperty checks if a property exists
func HasRawIssueProperty(r core.ZodRawIssue, key string) bool {
	return mapx.Has(r.Properties, key)
}
