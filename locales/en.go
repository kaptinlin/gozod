package locales

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// ENGLISH LOCALE FORMATTER
// =============================================================================

// formatEn provides English error messages
func formatEn(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.GetStringDefault(raw.Properties, "expected", "")
		received := issues.ParsedTypeToString(raw.Input)
		return fmt.Sprintf("Invalid input: expected %s, received %s", expected, received)

	case core.InvalidValue:
		values := mapx.GetAnySliceDefault(raw.Properties, "values", nil)
		if len(values) == 0 {
			return "Invalid value"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Invalid input: expected %s", issues.StringifyPrimitive(values[0]))
		}
		// Use issues.JoinValuesWithSeparator for consistent formatting
		return fmt.Sprintf("Invalid option: expected one of %s", issues.JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return formatSizeConstraintEn(raw, false)

	case core.TooSmall:
		return formatSizeConstraintEn(raw, true)

	case core.InvalidFormat:
		format := mapx.GetStringDefault(raw.Properties, "format", "")
		if format == "" {
			return "Invalid format"
		}
		return formatStringValidationEn(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.GetAnyDefault(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Invalid number: must be a multiple of divisor"
		}
		return fmt.Sprintf("Invalid number: must be a multiple of %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.GetStringsDefault(raw.Properties, "keys", nil)
		if len(keys) == 0 {
			return "Unrecognized key(s) in object"
		}
		keyStr := "key"
		if len(keys) > 1 {
			keyStr = "keys"
		}
		// Use slicex for better key handling and issues.JoinValuesWithSeparator for formatting
		if keysAny, err := slicex.ToAny(keys); err == nil && !slicex.IsEmpty(keysAny) {
			keysJoined := issues.JoinValuesWithSeparator(keysAny, ", ")
			return fmt.Sprintf("Unrecognized %s: %s", keyStr, keysJoined)
		}
		return fmt.Sprintf("Unrecognized %s in object", keyStr)

	case core.InvalidKey:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "Invalid key"
		}
		return fmt.Sprintf("Invalid key in %s", origin)

	case core.InvalidUnion:
		return "Invalid input"

	case core.InvalidElement:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "Invalid element"
		}
		return fmt.Sprintf("Invalid value in %s", origin)

	case core.Custom:
		message := mapx.GetStringDefault(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Invalid input"

	default:
		return "Invalid input"
	}
}

// =============================================================================
// SIZE CONSTRAINT FORMATTING
// =============================================================================

// formatSizeConstraintEn formats size constraint messages
func formatSizeConstraintEn(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.GetStringDefault(raw.Properties, "origin", "")
	if origin == "" {
		origin = "value"
	}

	var threshold any
	if isTooSmall {
		threshold = mapx.GetAnyDefault(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.GetAnyDefault(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Too small"
		}
		return "Too big"
	}

	inclusive := mapx.GetBoolDefault(raw.Properties, "inclusive", true)
	adj := issues.GetComparisonOperator(inclusive, isTooSmall)
	sizing := issues.GetSizing(origin)
	thresholdStr := issues.FormatThreshold(threshold)

	if sizing != nil {
		if isTooSmall {
			return fmt.Sprintf("Too small: expected %s to have %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		} else {
			return fmt.Sprintf("Too big: expected %s to have %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
	}

	if isTooSmall {
		return fmt.Sprintf("Too small: expected %s to be %s%s", origin, adj, thresholdStr)
	} else {
		return fmt.Sprintf("Too big: expected %s to be %s%s", origin, adj, thresholdStr)
	}
}

// =============================================================================
// STRING FORMAT VALIDATION
// =============================================================================

// formatStringValidationEn handles string format validation messages
func formatStringValidationEn(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.GetStringDefault(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Invalid string: must start with specified prefix"
		}
		return fmt.Sprintf("Invalid string: must start with %s", issues.StringifyPrimitive(prefix))
	case "ends_with":
		suffix := mapx.GetStringDefault(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Invalid string: must end with specified suffix"
		}
		return fmt.Sprintf("Invalid string: must end with %s", issues.StringifyPrimitive(suffix))
	case "includes":
		includes := mapx.GetStringDefault(raw.Properties, "includes", "")
		if includes == "" {
			return "Invalid string: must include specified substring"
		}
		return fmt.Sprintf("Invalid string: must include %s", issues.StringifyPrimitive(includes))
	case "regex":
		pattern := mapx.GetStringDefault(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Invalid string: must match pattern"
		}
		return fmt.Sprintf("Invalid string: must match pattern %s", pattern)
	default:
		noun := issues.GetFormatNoun(format)
		return fmt.Sprintf("Invalid %s", noun)
	}
}

// EN returns a ZodConfig configured for the default English locale.
func EN() *core.ZodConfig {
	return &core.ZodConfig{
		LocaleError: formatEn,
	}
}

// =============================================================================
// UNIFIED MESSAGE FORMATTING WITH TYPESCRIPT ZOD V4 COMPATIBILITY
// =============================================================================
