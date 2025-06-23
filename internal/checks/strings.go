// Package checks provides string validation checks
package checks

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// REGEX PATTERN VALIDATION
// =============================================================================

// Regex creates a regex pattern check with JSON Schema support
// Supports: Regex(pattern, "invalid format") or Regex(pattern, CheckParams{Error: "does not match pattern"})
func Regex(pattern *regexp.Regexp, params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "regex"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Regex(payload.GetValue(), pattern) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("regex", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set pattern for JSON Schema
				setBagProperty(schema, "pattern", pattern.String())
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// STRING CONTENT VALIDATION
// =============================================================================

// Includes creates a substring inclusion check with JSON Schema support
// Supports: Includes("test", "must contain test") or Includes("test", CheckParams{Error: "missing required substring"})
func Includes(substring string, params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "includes"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Includes(payload.GetValue(), substring) {
				additionalProps := map[string]any{
					"includes": substring,
				}
				payload.AddIssue(issues.CreateInvalidFormatIssue("includes", payload.GetValue(), additionalProps))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Add custom property for includes validation
				setBagProperty(schema, "contains", substring)
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// StartsWith creates a prefix check with JSON Schema support
// Supports: StartsWith("prefix", "wrong start") or StartsWith("prefix", CheckParams{Error: "must start with prefix"})
func StartsWith(prefix string, params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "starts_with"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.StartsWith(payload.GetValue(), prefix) {
				additionalProps := map[string]any{
					"prefix": prefix,
				}
				payload.AddIssue(issues.CreateInvalidFormatIssue("starts_with", payload.GetValue(), additionalProps))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Use pattern for starts with validation
				pattern := "^" + regexp.QuoteMeta(prefix)
				addPatternToSchema(schema, pattern)
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// EndsWith creates a suffix check with JSON Schema support
// Supports: EndsWith("suffix", "wrong end") or EndsWith("suffix", CheckParams{Error: "must end with suffix"})
func EndsWith(suffix string, params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ends_with"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.EndsWith(payload.GetValue(), suffix) {
				additionalProps := map[string]any{
					"suffix": suffix,
				}
				payload.AddIssue(issues.CreateInvalidFormatIssue("ends_with", payload.GetValue(), additionalProps))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Use pattern for ends with validation
				pattern := regexp.QuoteMeta(suffix) + "$"
				addPatternToSchema(schema, pattern)
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// CASE VALIDATION
// =============================================================================

// Lowercase creates a lowercase format check with JSON Schema support
// Supports: Lowercase("must be lowercase") or Lowercase(CheckParams{Error: "only lowercase letters allowed"})
func Lowercase(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "lowercase"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Lowercase(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("lowercase", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Use pattern for lowercase validation
				addPatternToSchema(schema, "^[^A-Z]*$")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// Uppercase creates an uppercase format check with JSON Schema support
// Supports: Uppercase("must be uppercase") or Uppercase(CheckParams{Error: "only uppercase letters allowed"})
func Uppercase(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "uppercase"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Uppercase(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("uppercase", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Use pattern for uppercase validation
				addPatternToSchema(schema, "^[^a-z]*$")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// addPatternToSchema adds a regex pattern to schema, combining with existing patterns
func addPatternToSchema(schema any, pattern string) {
	if s, ok := schema.(interface{ GetInternals() *core.ZodTypeInternals }); ok {
		internals := s.GetInternals()
		if internals.Bag == nil {
			internals.Bag = make(map[string]any)
		}

		internals.Bag["pattern"] = pattern
	}
}
