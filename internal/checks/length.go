// Package checks provides length and size constraint validation
package checks

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// LENGTH CONSTRAINT FACTORY FUNCTIONS
// =============================================================================

// MaxLength creates a maximum length validation check with JSON Schema support
// Supports: MaxLength(5, "too long") or MaxLength(5, CheckParams{Error: "too long"})
func MaxLength(maximum int, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "max_length"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MaxLength(payload.GetValue(), maximum) {
				origin := utils.GetLengthableOrigin(payload.GetValue())
				payload.AddIssue(issues.CreateTooBigIssue(maximum, true, origin, payload.GetValue()))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set maxLength for JSON Schema
				mergeMaximumLengthConstraint(schema, maximum)
			},
		},
	}
}

// MinLength creates a minimum length validation check with JSON Schema support
// Supports: MinLength(3, "too short") or MinLength(3, CheckParams{Error: "too short"})
func MinLength(minimum int, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "min_length"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MinLength(payload.GetValue(), minimum) {
				origin := utils.GetLengthableOrigin(payload.GetValue())
				payload.AddIssue(issues.CreateTooSmallIssue(minimum, true, origin, payload.GetValue()))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set minLength for JSON Schema
				mergeMinimumLengthConstraint(schema, minimum)
			},
		},
	}
}

// Length creates an exact length validation check with JSON Schema support
// Supports: Length(10, "must be exactly 10 chars") or Length(10, CheckParams{Error: "exact length required"})
func Length(exact int, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "length_equals"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Length(payload.GetValue(), exact) {
				// Determine if too big or too small based on actual length
				actualLength := getActualLength(payload.GetValue())
				origin := utils.GetLengthableOrigin(payload.GetValue())
				if actualLength > exact {
					payload.AddIssue(issues.CreateTooBigIssue(exact, true, origin, payload.GetValue()))
				} else {
					payload.AddIssue(issues.CreateTooSmallIssue(exact, true, origin, payload.GetValue()))
				}
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Exact length sets both min and max
				SetBagProperty(schema, "minLength", exact)
				SetBagProperty(schema, "maxLength", exact)
			},
		},
	}
}

// =============================================================================
// SIZE CONSTRAINT FACTORY FUNCTIONS
// =============================================================================

// MaxSize creates a maximum size validation check with JSON Schema support
// Supports: MaxSize(100, "too many items") or MaxSize(100, CheckParams{Error: "size limit exceeded"})
func MaxSize(maximum int, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "max_size"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MaxSize(payload.GetValue(), maximum) {
				origin := utils.GetSizableOrigin(payload.GetValue())
				payload.AddIssue(issues.CreateTooBigIssue(maximum, true, origin, payload.GetValue()))
			}
		},
		When: func(payload *core.ParsePayload) bool {
			// Only execute when value has size property
			return reflectx.HasSize(payload.GetValue()) || reflectx.HasLength(payload.GetValue())
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set maxItems/maxProperties based on type
				setMaxSizeProperty(schema, maximum)
			},
		},
	}
}

// MinSize creates a minimum size validation check with JSON Schema support
// Supports: MinSize(1, "cannot be empty") or MinSize(1, CheckParams{Error: "minimum size required"})
func MinSize(minimum int, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "min_size"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MinSize(payload.GetValue(), minimum) {
				origin := utils.GetSizableOrigin(payload.GetValue())
				payload.AddIssue(issues.CreateTooSmallIssue(minimum, true, origin, payload.GetValue()))
			}
		},
		When: func(payload *core.ParsePayload) bool {
			// Only execute when value has size property
			return reflectx.HasSize(payload.GetValue()) || reflectx.HasLength(payload.GetValue())
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set minItems/minProperties based on type
				setMinSizeProperty(schema, minimum)
			},
		},
	}
}

// Size creates an exact size validation check with JSON Schema support
// Supports: Size(5, "must have exactly 5 items") or Size(5, CheckParams{Error: "exact size required"})
func Size(exact int, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "size_equals"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Size(payload.GetValue(), exact) {
				// Determine if too big or too small based on actual size
				if actualSize, ok := reflectx.Size(payload.GetValue()); ok {
					origin := utils.GetSizableOrigin(payload.GetValue())
					if actualSize > exact {
						payload.AddIssue(issues.CreateTooBigIssue(exact, true, origin, payload.GetValue()))
					} else {
						payload.AddIssue(issues.CreateTooSmallIssue(exact, true, origin, payload.GetValue()))
					}
				}
			}
		},
		When: func(payload *core.ParsePayload) bool {
			// Only execute when value has size property
			return reflectx.HasSize(payload.GetValue()) || reflectx.HasLength(payload.GetValue())
		},
		OnAttach: []func(any){
			func(schema any) {
				// Exact size sets both min and max
				setMinSizeProperty(schema, exact)
				setMaxSizeProperty(schema, exact)
			},
		},
	}
}

// =============================================================================
// RANGE CONSTRAINT FUNCTIONS
// =============================================================================

// LengthRange creates a length range validation check with JSON Schema support
// Supports: LengthRange(3, 10, "length must be 3-10") or LengthRange(3, 10, CheckParams{Error: "invalid length range"})
func LengthRange(minimum, maximum int, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "length_range"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			actualLength := getActualLength(payload.GetValue())
			origin := utils.GetLengthableOrigin(payload.GetValue())

			if actualLength < minimum {
				payload.AddIssue(issues.CreateTooSmallIssue(minimum, true, origin, payload.GetValue()))
			} else if actualLength > maximum {
				payload.AddIssue(issues.CreateTooBigIssue(maximum, true, origin, payload.GetValue()))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set both min and max length for JSON Schema
				mergeMinimumLengthConstraint(schema, minimum)
				mergeMaximumLengthConstraint(schema, maximum)
			},
		},
	}
}

// SizeRange creates a size range validation check with JSON Schema support
// Supports: SizeRange(1, 100, "size must be 1-100") or SizeRange(1, 100, CheckParams{Error: "invalid size range"})
func SizeRange(minimum, maximum int, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "size_range"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if actualSize, ok := reflectx.Size(payload.GetValue()); ok {
				origin := utils.GetSizableOrigin(payload.GetValue())
				if actualSize > maximum {
					payload.AddIssue(issues.CreateTooBigIssue(maximum, true, origin, payload.GetValue()))
				} else if actualSize < minimum {
					payload.AddIssue(issues.CreateTooSmallIssue(minimum, true, origin, payload.GetValue()))
				}
			}
		},
		When: func(payload *core.ParsePayload) bool {
			// Only execute when value has size property
			return reflectx.HasSize(payload.GetValue()) || reflectx.HasLength(payload.GetValue())
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set both min and max size for JSON Schema
				setMinSizeProperty(schema, minimum)
				setMaxSizeProperty(schema, maximum)
			},
		},
	}
}

// =============================================================================
// CONVENIENCE FUNCTIONS
// =============================================================================

// NonEmpty creates a non-empty validation check with JSON Schema support
// Supports: NonEmpty("cannot be empty") or NonEmpty(CheckParams{Error: "value required"})
func NonEmpty(params ...any) core.ZodCheck {
	return MinLength(1, params...)
}

// Empty creates an empty validation check with JSON Schema support
// Supports: Empty("must be empty") or Empty(CheckParams{Error: "value must be empty"})
func Empty(params ...any) core.ZodCheck {
	return Length(0, params...)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// getActualLength returns the actual length using optimized utilities
func getActualLength(value any) int {
	if l, ok := reflectx.Length(value); ok {
		return l
	}
	return 0
}

// mergeMinimumLengthConstraint merges minLength constraint
func mergeMinimumLengthConstraint(schema any, length int) {
	mergeConstraint(schema, "minLength", length, func(old, new any) any {
		// Choose stricter (larger) minimum
		if oldInt, ok := old.(int); ok {
			if newInt, ok := new.(int); ok && newInt > oldInt {
				return newInt
			}
			return oldInt
		}
		return new
	})
}

// mergeMaximumLengthConstraint merges maxLength constraint
func mergeMaximumLengthConstraint(schema any, length int) {
	mergeConstraint(schema, "maxLength", length, func(old, new any) any {
		// Choose stricter (smaller) maximum
		if oldInt, ok := old.(int); ok {
			if newInt, ok := new.(int); ok && newInt < oldInt {
				return newInt
			}
			return oldInt
		}
		return new
	})
}

// setMinSizeProperty sets appropriate min size property based on type
func setMinSizeProperty(schema any, size int) {
	bag := ensureBag(schema)
	if bag == nil {
		return
	}

	// Determine property name based on schema type
	if typeStr, exists := bag["type"]; exists {
		switch typeStr {
		case "array":
			bag["minItems"] = size
		case "object":
			bag["minProperties"] = size
		default:
			bag["minLength"] = size
		}
	} else {
		// Default to minItems for collections
		bag["minItems"] = size
	}
}

// setMaxSizeProperty sets appropriate max size property based on type
func setMaxSizeProperty(schema any, size int) {
	bag := ensureBag(schema)
	if bag == nil {
		return
	}

	// Determine property name based on schema type
	if typeStr, exists := bag["type"]; exists {
		switch typeStr {
		case "array":
			bag["maxItems"] = size
		case "object":
			bag["maxProperties"] = size
		default:
			bag["maxLength"] = size
		}
	} else {
		// Default to maxItems for collections
		bag["maxItems"] = size
	}
}
