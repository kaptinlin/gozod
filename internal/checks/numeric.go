package checks

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// Lt creates a less-than validation check.
func Lt(value any, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "less_than"}
	ApplyCheckParams(def, cp)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.Lt(payload.GetValue(), value) {
			origin := utils.NumericOrigin(payload.GetValue())
			raw := issues.CreateTooBigIssue(value, false, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) { mergeMaximumConstraint(schema, value, false) },
	}
	return internals
}

// Lte creates a less-than-or-equal validation check.
func Lte(value any, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "less_than_or_equal"}
	ApplyCheckParams(def, cp)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.Lte(payload.GetValue(), value) {
			origin := utils.NumericOrigin(payload.GetValue())
			raw := issues.CreateTooBigIssue(value, true, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) { mergeMaximumConstraint(schema, value, true) },
	}
	return internals
}

// Gt creates a greater-than validation check.
func Gt(value any, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "greater_than"}
	ApplyCheckParams(def, cp)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.Gt(payload.GetValue(), value) {
			origin := utils.NumericOrigin(payload.GetValue())
			raw := issues.CreateTooSmallIssue(value, false, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) { mergeMinimumConstraint(schema, value, false) },
	}
	return internals
}

// Gte creates a greater-than-or-equal validation check.
func Gte(value any, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "greater_than_or_equal"}
	ApplyCheckParams(def, cp)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.Gte(payload.GetValue(), value) {
			origin := utils.NumericOrigin(payload.GetValue())
			raw := issues.CreateTooSmallIssue(value, true, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) { mergeMinimumConstraint(schema, value, true) },
	}
	return internals
}

// MultipleOf creates a multiple-of validation check.
func MultipleOf(divisor any, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "multiple_of"}
	ApplyCheckParams(def, cp)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.MultipleOf(payload.GetValue(), divisor) {
			origin := utils.NumericOrigin(payload.GetValue())
			raw := issues.CreateNotMultipleOfIssue(divisor, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) { SetBagProperty(schema, "multipleOf", divisor) },
	}

	return internals
}

// Positive creates a check that the value is greater than zero.
func Positive(params ...any) core.ZodCheck { return Gt(0, params...) }

// Negative creates a check that the value is less than zero.
func Negative(params ...any) core.ZodCheck { return Lt(0, params...) }

// NonPositive creates a check that the value is less than or equal to zero.
func NonPositive(params ...any) core.ZodCheck { return Lte(0, params...) }

// NonNegative creates a check that the value is greater than or equal to zero.
func NonNegative(params ...any) core.ZodCheck { return Gte(0, params...) }
