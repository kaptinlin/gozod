package issues

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
)

// =============================================================================
// ISSUE FINALIZATION AND PROCESSING
// =============================================================================

// FinalizeIssue creates a finalized ZodIssue from a ZodRawIssue
// Handles error message resolution chain and property mapping
func FinalizeIssue(iss core.ZodRawIssue, ctx *core.ParseContext, config *core.ZodConfig) core.ZodIssue {
	// Ensure path is not nil using slicex
	path := iss.Path
	if path == nil {
		path = []any{}
	}

	// Generate message using error resolution chain
	message := iss.Message
	if message == "" {
		// Try to get message from various sources in priority order:
		// 1. inst?.error (schema-level error)
		// 2. ctx?.error (context-level error)
		// 3. config.customError (global custom error)
		// 4. config.localeError (locale error)
		// 5. default message

		// Schema-level error handling
		if iss.Inst != nil {
			if instMsg := ExtractSchemaLevelError(iss); instMsg != "" {
				message = instMsg
			}
		}

		// Check context-level error
		if message == "" && ctx != nil && ctx.Error != nil {
			if ctxMsg := ctx.Error(iss); ctxMsg != "" {
				message = ctxMsg
			}
		}

		// Config-level error handling
		if message == "" && config != nil {
			if configMsg := ExtractConfigLevelError(iss, config); configMsg != "" {
				message = configMsg
			}
		}

		// If no custom message found, generate default message
		if message == "" {
			message = GenerateDefaultMessage(iss)
		}
	}

	// Create finalized issue with base information
	issue := core.ZodIssue{
		ZodIssueBase: core.ZodIssueBase{
			Code:    iss.Code,
			Path:    path,
			Message: message,
		},
	}

	// Handle input field based on context ReportInput setting
	if ctx == nil || ctx.ReportInput {
		issue.Input = iss.Input
	}

	// Map properties from raw issue to typed fields using mapx
	if iss.Properties != nil {
		MapPropertiesToIssue(&issue, iss.Properties)
	}

	return issue
}

// ExtractSchemaLevelError extracts error message from schema instance
func ExtractSchemaLevelError(iss core.ZodRawIssue) string {
	if iss.Inst == nil {
		return ""
	}

	// Try to extract error mapping from different types of schema internals
	switch inst := iss.Inst.(type) {
	case *core.ZodTypeInternals:
		// Handle direct ZodTypeInternals (most common case from ParsePrimitive)
		if inst.Error != nil {
			return (*inst.Error)(iss)
		}

	case *core.ZodCheckInternals:
		// Handle check internals directly
		if inst.Def != nil && inst.Def.Error != nil {
			return (*inst.Def.Error)(iss)
		}

	case interface{ GetError() *core.ZodErrorMap }:
		// Handle schemas with direct error mapping access
		if errorMap := inst.GetError(); errorMap != nil {
			return (*errorMap)(iss)
		}

	case interface{ GetInternals() *core.ZodTypeInternals }:
		// Handle schemas with internals that contain error mapping
		if internals := inst.GetInternals(); internals != nil && internals.Error != nil {
			return (*internals.Error)(iss)
		}

	case interface {
		GetZod() *core.ZodCheckInternals
	}:
		// Handle check types with check internals
		if checkInternals := inst.GetZod(); checkInternals != nil &&
			checkInternals.Def != nil && checkInternals.Def.Error != nil {
			return (*checkInternals.Def.Error)(iss)
		}
	}

	// Try extraction using helper
	if checkInternals := tryExtractCheckInternalsFromAny(iss.Inst); checkInternals != nil {
		if checkInternals.Def != nil && checkInternals.Def.Error != nil {
			return (*checkInternals.Def.Error)(iss)
		}
	}

	// Handle ZodCheckCustomInternals directly (for refine operations)
	// Try direct field access for custom check internals
	if customInternals, ok := iss.Inst.(interface {
		GetZod() *core.ZodCheckInternals
	}); ok {
		if zodInternals := customInternals.GetZod(); zodInternals != nil {
			if zodInternals.Def != nil && zodInternals.Def.Error != nil {
				return (*zodInternals.Def.Error)(iss)
			}
		}
	}

	return ""
}

// tryExtractCheckInternalsFromAny extracts check internals from any type
func tryExtractCheckInternalsFromAny(inst any) *core.ZodCheckInternals {
	// This handles the specific case of ZodCheckCustomInternals from checks package
	// We use the fact that it embeds core.ZodCheckInternals
	if v, ok := inst.(interface {
		ZodCheckInternals() *core.ZodCheckInternals
	}); ok {
		return v.ZodCheckInternals()
	}

	// Try to access via embedded field using type assertion
	// Since ZodCheckCustomInternals embeds core.ZodCheckInternals, we can try to access it
	if embedder, ok := inst.(interface {
		GetCheckInternals() *core.ZodCheckInternals
	}); ok {
		return embedder.GetCheckInternals()
	}

	return nil
}

// ExtractConfigLevelError extracts error message from config
func ExtractConfigLevelError(iss core.ZodRawIssue, config *core.ZodConfig) string {
	if config == nil {
		return ""
	}

	// Try config custom error first
	if customError := GetCustomError(config); customError != nil {
		if customMsg := customError(iss); customMsg != "" {
			return customMsg
		}
	}

	// Try locale error
	if localeError := GetLocaleError(config); localeError != nil {
		if localeMsg := localeError(iss); localeMsg != "" {
			return localeMsg
		}
	}

	return ""
}

// GetCustomError safely extracts custom error from config
func GetCustomError(config *core.ZodConfig) core.ZodErrorMap {
	if config == nil {
		return nil
	}
	return config.CustomError
}

// GetLocaleError safely extracts locale error from config
func GetLocaleError(config *core.ZodConfig) core.ZodErrorMap {
	if config == nil {
		return nil
	}
	return config.LocaleError
}

// MapPropertiesToIssue maps properties to ZodIssue fields using mapx
func MapPropertiesToIssue(issue *core.ZodIssue, properties map[string]any) {
	// Use mapx for safer property access with zero value defaults
	// This ensures that missing or wrong-type properties result in zero values

	// Handle ZodTypeCode fields with proper type conversion
	// Always set these fields, even if empty (to overwrite existing values)
	expectedStr := mapx.GetStringDefault(properties, "expected", "")
	issue.Expected = core.ZodTypeCode(expectedStr)
	receivedStr := mapx.GetStringDefault(properties, "received", "")
	issue.Received = core.ZodTypeCode(receivedStr)

	issue.Origin = mapx.GetStringDefault(properties, "origin", "")
	issue.Format = mapx.GetStringDefault(properties, "format", "")
	issue.Pattern = mapx.GetStringDefault(properties, "pattern", "")
	issue.Prefix = mapx.GetStringDefault(properties, "prefix", "")
	issue.Suffix = mapx.GetStringDefault(properties, "suffix", "")
	issue.Includes = mapx.GetStringDefault(properties, "includes", "")
	issue.Algorithm = mapx.GetStringDefault(properties, "algorithm", "")

	// Handle numeric and any values with nil defaults
	// Always set these fields to ensure existing values are overwritten
	issue.Minimum = mapx.GetAnyDefault(properties, "minimum", nil)
	issue.Maximum = mapx.GetAnyDefault(properties, "maximum", nil)
	issue.Divisor = mapx.GetAnyDefault(properties, "divisor", nil)
	issue.Key = mapx.GetAnyDefault(properties, "key", nil)

	// Handle boolean values with false default
	issue.Inclusive = mapx.GetBoolDefault(properties, "inclusive", false)

	// Handle slices with nil defaults
	issue.Keys = mapx.GetStringsDefault(properties, "keys", nil)
	issue.Values = mapx.GetAnySliceDefault(properties, "values", nil)

	// Handle params map with nil default
	issue.Params = mapx.GetMapDefault(properties, "params", nil)

	// Handle element_error for InvalidElement issues
	if elementError, exists := properties["element_error"]; exists && elementError != nil {
		if rawIssue, ok := elementError.(core.ZodRawIssue); ok {
			// Convert the raw issue to a finalized issue
			finalizedElementIssue := FinalizeIssue(rawIssue, nil, nil)
			issue.Issues = []core.ZodIssue{finalizedElementIssue}
		}
	}
}

// =============================================================================
// CONVENIENCE FUNCTIONS USING MAPX
// =============================================================================

// CopyRawIssueProperties creates a copy of raw issue properties using mapx
func CopyRawIssueProperties(rawIssue core.ZodRawIssue) map[string]any {
	return mapx.Copy(rawIssue.Properties)
}

// MergeRawIssueProperties merges new properties into raw issue using mapx
func MergeRawIssueProperties(rawIssue *core.ZodRawIssue, newProperties map[string]any) {
	if rawIssue.Properties == nil {
		rawIssue.Properties = make(map[string]any)
	}
	rawIssue.Properties = mapx.Merge(rawIssue.Properties, newProperties)
}

// =============================================================================
// CONVERSION FUNCTIONS
// =============================================================================

// ConvertRawIssuesToIssues converts a slice of raw issues to a slice of finalized issues
func ConvertRawIssuesToIssues(rawIssues []core.ZodRawIssue, ctx *core.ParseContext) []core.ZodIssue {
	// Get the global config to be passed down
	config := core.GetConfig()

	// Manually iterate to ensure type correctness
	issues := make([]core.ZodIssue, len(rawIssues))
	for i, rawIssue := range rawIssues {
		issues[i] = FinalizeIssue(rawIssue, ctx, config)
	}
	return issues
}
