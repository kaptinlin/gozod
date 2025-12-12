// Package checks provides string validation checks
package checks

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// REGEX PATTERN VALIDATION
// =============================================================================

// Regex creates a regex pattern check with JSON Schema support
// Supports: Regex(pattern, "invalid format") or Regex(pattern, CheckParams{Error: "does not match pattern"})
func Regex(pattern *regexp.Regexp, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "regex"}
	ApplyCheckParams(def, checkParams)

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
				addPatternToSchema(schema, pattern.String())
				SetBagProperty(schema, "type", "string")
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
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "includes"}
	ApplyCheckParams(def, checkParams)

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
				// Use pattern to simulate contains, escaping the substring for regex
				pattern := regexp.QuoteMeta(substring)
				addPatternToSchema(schema, pattern)
			},
		},
	}
}

// StartsWith creates a prefix check with JSON Schema support
// Supports: StartsWith("prefix", "wrong start") or StartsWith("prefix", CheckParams{Error: "must start with prefix"})
func StartsWith(prefix string, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "starts_with"}
	ApplyCheckParams(def, checkParams)

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
				pattern := "^" + regexp.QuoteMeta(prefix) + ".*"
				addPatternToSchema(schema, pattern)
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// EndsWith creates a suffix check with JSON Schema support
// Supports: EndsWith("suffix", "wrong end") or EndsWith("suffix", CheckParams{Error: "must end with suffix"})
func EndsWith(suffix string, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ends_with"}
	ApplyCheckParams(def, checkParams)

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
				pattern := ".*" + regexp.QuoteMeta(suffix) + "$"
				addPatternToSchema(schema, pattern)
				SetBagProperty(schema, "type", "string")
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
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "lowercase"}
	ApplyCheckParams(def, checkParams)

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
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Uppercase creates an uppercase format check with JSON Schema support
// Supports: Uppercase("must be uppercase") or Uppercase(CheckParams{Error: "only uppercase letters allowed"})
func Uppercase(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "uppercase"}
	ApplyCheckParams(def, checkParams)

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
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// STRING VALUE COMPARISON FUNCTIONS
// =============================================================================

// StringGte creates a string greater than or equal validation check
// Uses lexicographic comparison, perfect for ISO 8601 date/time strings
func StringGte(minValue string, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "string_gte"}
	ApplyCheckParams(def, checkParams)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if str, ok := reflectx.ExtractString(payload.GetValue()); ok {
			if str < minValue {
				raw := issues.CreateTooSmallIssue(minValue, true, "string", payload.GetValue())
				raw.Inst = internals
				payload.AddIssue(raw)
			}
		} else {
			// If not a string, create a type mismatch issue
			raw := issues.CreateInvalidTypeIssue(core.ZodTypeString, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	return internals
}

// StringLte creates a string less than or equal validation check
// Uses lexicographic comparison, perfect for ISO 8601 date/time strings
func StringLte(maxValue string, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "string_lte"}
	ApplyCheckParams(def, checkParams)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if str, ok := reflectx.ExtractString(payload.GetValue()); ok {
			if str > maxValue {
				raw := issues.CreateTooBigIssue(maxValue, true, "string", payload.GetValue())
				raw.Inst = internals
				payload.AddIssue(raw)
			}
		} else {
			// If not a string, create a type mismatch issue
			raw := issues.CreateInvalidTypeIssue(core.ZodTypeString, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	return internals
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// addPatternToSchema adds a regex pattern to schema, combining with existing patterns
func addPatternToSchema(schema any, pattern string) {
	bag := ensureBag(schema)
	if bag == nil {
		return
	}

	var patterns []string
	if existing, ok := bag["patterns"]; ok {
		if existingPatterns, ok := existing.([]string); ok {
			patterns = existingPatterns
		}
	}
	// Avoid duplicate patterns
	for _, existingPattern := range patterns {
		if existingPattern == pattern {
			// Pattern already exists; no need to add again
			return
		}
	}
	patterns = append(patterns, pattern)
	bag["patterns"] = patterns
}
