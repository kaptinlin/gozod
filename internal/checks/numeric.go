// Package checks provides numeric validation checks
package checks

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// NUMERIC CONSTRAINT FACTORY FUNCTIONS
// =============================================================================

// Lt creates a less than validation check with JSON Schema support
// Supports: Lt(10, "too big") or Lt(10, CheckParams{Error: "value must be less than 10"})
func Lt(value any, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "less_than"}
	ApplyCheckParams(def, checkParams)

	internals := &core.ZodCheckInternals{Def: def}

	internals.Check = func(payload *core.ParsePayload) {
		if !validate.Lt(payload.GetValue(), value) {
			origin := utils.GetNumericOrigin(payload.GetValue())
			raw := issues.CreateTooBigIssue(value, false, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) {
			// Set exclusiveMaximum for JSON Schema
			mergeMaximumConstraint(schema, value, false)
		},
	}

	return internals
}

// Lte creates a less than or equal validation check with JSON Schema support
// Supports: Lte(10, "too big") or Lte(10, CheckParams{Error: "value must be at most 10"})
func Lte(value any, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "less_than_or_equal"}
	ApplyCheckParams(def, checkParams)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.Lte(payload.GetValue(), value) {
			origin := utils.GetNumericOrigin(payload.GetValue())
			raw := issues.CreateTooBigIssue(value, true, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) {
			// Set maximum for JSON Schema
			mergeMaximumConstraint(schema, value, true)
		},
	}
	return internals
}

// Gt creates a greater than validation check with JSON Schema support
// Supports: Gt(0, "too small") or Gt(0, CheckParams{Error: "value must be greater than 0"})
func Gt(value any, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "greater_than"}
	ApplyCheckParams(def, checkParams)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.Gt(payload.GetValue(), value) {
			origin := utils.GetNumericOrigin(payload.GetValue())
			raw := issues.CreateTooSmallIssue(value, false, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) {
			// Set exclusiveMinimum for JSON Schema
			mergeMinimumConstraint(schema, value, false)
		},
	}
	return internals
}

// Gte creates a greater than or equal validation check with JSON Schema support
// Supports: Gte(0, "too small") or Gte(0, CheckParams{Error: "value must be at least 0"})
func Gte(value any, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "greater_than_or_equal"}
	ApplyCheckParams(def, checkParams)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.Gte(payload.GetValue(), value) {
			origin := utils.GetNumericOrigin(payload.GetValue())
			raw := issues.CreateTooSmallIssue(value, true, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) {
			// Set minimum for JSON Schema
			mergeMinimumConstraint(schema, value, true)
		},
	}
	return internals
}

// MultipleOf creates a multiple of validation check with JSON Schema support
// Supports: MultipleOf(5, "not divisible") or MultipleOf(5, CheckParams{Error: "must be multiple of 5"})
func MultipleOf(divisor any, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "multiple_of"}
	ApplyCheckParams(def, checkParams)

	internals := &core.ZodCheckInternals{Def: def}
	internals.Check = func(payload *core.ParsePayload) {
		if !validate.MultipleOf(payload.GetValue(), divisor) {
			origin := utils.GetNumericOrigin(payload.GetValue())
			raw := issues.CreateNotMultipleOfIssue(divisor, origin, payload.GetValue())
			raw.Inst = internals
			payload.AddIssue(raw)
		}
	}
	internals.OnAttach = []func(any){
		func(schema any) {
			// Set multipleOf for JSON Schema
			SetBagProperty(schema, "multipleOf", divisor)
		},
	}

	return internals
}

// =============================================================================
// CONVENIENCE FUNCTIONS
// =============================================================================

// Positive creates a positive number validation check (greater than zero)
// Supports: Positive("must be positive") or Positive(CheckParams{Error: "value must be positive"})
func Positive(params ...any) core.ZodCheck {
	return Gt(0, params...)
}

// Negative creates a negative number validation check (less than zero)
// Supports: Negative("must be negative") or Negative(CheckParams{Error: "value must be negative"})
func Negative(params ...any) core.ZodCheck {
	return Lt(0, params...)
}

// NonPositive creates a non-positive number validation check (less than or equal to zero)
// Supports: NonPositive("cannot be positive") or NonPositive(CheckParams{Error: "value cannot be positive"})
func NonPositive(params ...any) core.ZodCheck {
	return Lte(0, params...)
}

// NonNegative creates a non-negative number validation check (greater than or equal to zero)
// Supports: NonNegative("cannot be negative") or NonNegative(CheckParams{Error: "value cannot be negative"})
func NonNegative(params ...any) core.ZodCheck {
	return Gte(0, params...)
}
