package issues

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
)

// FinalizeIssue creates a finalized ZodIssue from a ZodRawIssue.
func FinalizeIssue(iss core.ZodRawIssue, ctx *core.ParseContext, config *core.ZodConfig) core.ZodIssue {
	path := iss.Path
	if path == nil {
		path = []any{}
	}

	message := iss.Message
	if message == "" {
		// Resolution chain: schema-level → context-level → config-level → default
		if iss.Inst != nil {
			if instMsg := ExtractSchemaLevelError(iss); instMsg != "" {
				message = instMsg
			}
		}

		if message == "" && ctx != nil && ctx.Error != nil {
			if ctxMsg := ctx.Error(iss); ctxMsg != "" {
				message = ctxMsg
			}
		}

		if message == "" && config != nil {
			if configMsg := ExtractConfigLevelError(iss, config); configMsg != "" {
				message = configMsg
			}
		}

		if message == "" {
			message = GenerateDefaultMessage(iss)
		}
	}

	issue := core.ZodIssue{
		ZodIssueBase: core.ZodIssueBase{
			Code:    iss.Code,
			Path:    path,
			Message: message,
		},
	}

	if ctx == nil || ctx.ReportInput {
		issue.Input = iss.Input
	}

	if iss.Properties != nil {
		MapPropertiesToIssue(&issue, iss.Properties)
	}

	return issue
}

// ExtractSchemaLevelError extracts error message from schema instance.
func ExtractSchemaLevelError(iss core.ZodRawIssue) string {
	if iss.Inst == nil {
		return ""
	}

	switch inst := iss.Inst.(type) {
	case *core.ZodTypeInternals:
		if inst.Error != nil {
			return (*inst.Error)(iss)
		}

	case *core.ZodCheckInternals:
		if inst.Def != nil && inst.Def.Error != nil {
			return (*inst.Def.Error)(iss)
		}

	case interface{ GetError() *core.ZodErrorMap }:
		if errorMap := inst.GetError(); errorMap != nil {
			return (*errorMap)(iss)
		}

	case interface{ GetInternals() *core.ZodTypeInternals }:
		if internals := inst.GetInternals(); internals != nil && internals.Error != nil {
			return (*internals.Error)(iss)
		}

	case interface {
		GetZod() *core.ZodCheckInternals
	}:
		if checkInternals := inst.GetZod(); checkInternals != nil &&
			checkInternals.Def != nil && checkInternals.Def.Error != nil {
			return (*checkInternals.Def.Error)(iss)
		}
	}

	if checkInternals := tryExtractCheckInternalsFromAny(iss.Inst); checkInternals != nil {
		if checkInternals.Def != nil && checkInternals.Def.Error != nil {
			return (*checkInternals.Def.Error)(iss)
		}
	}

	return ""
}

// tryExtractCheckInternalsFromAny extracts check internals from any type.
func tryExtractCheckInternalsFromAny(inst any) *core.ZodCheckInternals {
	if v, ok := inst.(interface {
		ZodCheckInternals() *core.ZodCheckInternals
	}); ok {
		return v.ZodCheckInternals()
	}

	if embedder, ok := inst.(interface {
		GetCheckInternals() *core.ZodCheckInternals
	}); ok {
		return embedder.GetCheckInternals()
	}

	return nil
}

// ExtractConfigLevelError extracts error message from config.
func ExtractConfigLevelError(iss core.ZodRawIssue, config *core.ZodConfig) string {
	if config == nil {
		return ""
	}

	if customError := CustomError(config); customError != nil {
		if customMsg := customError(iss); customMsg != "" {
			return customMsg
		}
	}

	if localeError := LocaleError(config); localeError != nil {
		if localeMsg := localeError(iss); localeMsg != "" {
			return localeMsg
		}
	}

	return ""
}

// CustomError safely extracts custom error from config.
func CustomError(config *core.ZodConfig) core.ZodErrorMap {
	if config == nil {
		return nil
	}
	return config.CustomError
}

// LocaleError safely extracts locale error from config.
func LocaleError(config *core.ZodConfig) core.ZodErrorMap {
	if config == nil {
		return nil
	}
	return config.LocaleError
}

// MapPropertiesToIssue maps properties to ZodIssue fields using mapx.
func MapPropertiesToIssue(issue *core.ZodIssue, properties map[string]any) {
	if len(properties) == 0 {
		return
	}

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

	issue.Minimum = mapx.GetAnyDefault(properties, "minimum", nil)
	issue.Maximum = mapx.GetAnyDefault(properties, "maximum", nil)
	issue.Divisor = mapx.GetAnyDefault(properties, "divisor", nil)
	issue.Key = mapx.GetAnyDefault(properties, "key", nil)

	issue.Inclusive = mapx.GetBoolDefault(properties, "inclusive", false)

	issue.Keys = mapx.GetStringsDefault(properties, "keys", nil)
	issue.Values = mapx.GetAnySliceDefault(properties, "values", nil)

	issue.Params = mapx.GetMapDefault(properties, "params", nil)

	if elementError, exists := properties["element_error"]; exists && elementError != nil {
		if rawIssue, ok := elementError.(core.ZodRawIssue); ok {
			finalizedElementIssue := FinalizeIssue(rawIssue, nil, nil)
			issue.Issues = []core.ZodIssue{finalizedElementIssue}
		}
	}
}

// CopyRawIssueProperties creates a copy of raw issue properties.
func CopyRawIssueProperties(rawIssue core.ZodRawIssue) map[string]any {
	return mapx.Copy(rawIssue.Properties)
}

// MergeRawIssueProperties merges new properties into raw issue.
func MergeRawIssueProperties(rawIssue *core.ZodRawIssue, newProperties map[string]any) {
	if rawIssue.Properties == nil {
		rawIssue.Properties = make(map[string]any)
	}
	rawIssue.Properties = mapx.Merge(rawIssue.Properties, newProperties)
}

// ConvertRawIssuesToIssues converts raw issues to finalized issues.
func ConvertRawIssuesToIssues(rawIssues []core.ZodRawIssue, ctx *core.ParseContext) []core.ZodIssue {
	if len(rawIssues) == 0 {
		return nil
	}

	config := core.GetConfig()

	issues := make([]core.ZodIssue, len(rawIssues))
	for i, rawIssue := range rawIssues {
		issues[i] = FinalizeIssue(rawIssue, ctx, config)
	}
	return issues
}
