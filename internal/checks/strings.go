package checks

import (
	"regexp"
	"slices"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// Regex creates a regex pattern validation check.
func Regex(pattern *regexp.Regexp, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "regex"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Regex(payload.GetValue(), pattern) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("regex", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				addPatternToSchema(schema, pattern.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Includes creates a substring inclusion check.
func Includes(substring string, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "includes"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Includes(payload.GetValue(), substring) {
				props := map[string]any{"includes": substring}
				payload.AddIssue(issues.CreateInvalidFormatIssue("includes", payload.GetValue(), props))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				addPatternToSchema(schema, regexp.QuoteMeta(substring))
			},
		},
	}
}

// StartsWith creates a prefix validation check.
func StartsWith(prefix string, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "starts_with"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.StartsWith(payload.GetValue(), prefix) {
				props := map[string]any{"prefix": prefix}
				payload.AddIssue(issues.CreateInvalidFormatIssue("starts_with", payload.GetValue(), props))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				addPatternToSchema(schema, "^"+regexp.QuoteMeta(prefix)+".*")
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// EndsWith creates a suffix validation check.
func EndsWith(suffix string, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ends_with"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.EndsWith(payload.GetValue(), suffix) {
				props := map[string]any{"suffix": suffix}
				payload.AddIssue(issues.CreateInvalidFormatIssue("ends_with", payload.GetValue(), props))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				addPatternToSchema(schema, ".*"+regexp.QuoteMeta(suffix)+"$")
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Lowercase creates a lowercase format validation check.
func Lowercase(params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "lowercase"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Lowercase(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("lowercase", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				addPatternToSchema(schema, "^[^A-Z]*$")
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Uppercase creates an uppercase format validation check.
func Uppercase(params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "uppercase"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Uppercase(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("uppercase", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				addPatternToSchema(schema, "^[^a-z]*$")
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// StringGte creates a string greater-than-or-equal check using
// lexicographic comparison, suitable for ISO 8601 date/time strings.
func StringGte(minValue string, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "string_gte"}
	ApplyCheckParams(def, cp)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if s, ok := reflectx.StringVal(payload.GetValue()); ok {
			if s < minValue {
				raw := issues.CreateTooSmallIssue(minValue, true, "string", payload.GetValue())
				raw.Inst = internals
				payload.AddIssue(raw)
			}
		} else {
			raw := issues.CreateInvalidTypeIssue(core.ZodTypeString, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	return internals
}

// StringLte creates a string less-than-or-equal check using
// lexicographic comparison, suitable for ISO 8601 date/time strings.
func StringLte(maxValue string, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "string_lte"}
	ApplyCheckParams(def, cp)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if s, ok := reflectx.StringVal(payload.GetValue()); ok {
			if s > maxValue {
				raw := issues.CreateTooBigIssue(maxValue, true, "string", payload.GetValue())
				raw.Inst = internals
				payload.AddIssue(raw)
			}
		} else {
			raw := issues.CreateInvalidTypeIssue(core.ZodTypeString, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	return internals
}

// addPatternToSchema adds a regex pattern to schema, deduplicating.
func addPatternToSchema(schema any, pattern string) {
	bag := ensureBag(schema)
	if bag == nil {
		return
	}
	var patterns []string
	if existing, ok := bag["patterns"]; ok {
		if p, ok := existing.([]string); ok {
			patterns = p
		}
	}
	if slices.Contains(patterns, pattern) {
		return
	}
	bag["patterns"] = append(patterns, pattern)
}
