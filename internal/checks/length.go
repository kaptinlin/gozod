package checks

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// MaxLength creates a maximum length validation check.
func MaxLength(maximum int, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "max_length"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MaxLength(payload.Value(), maximum) {
				origin := utils.LengthableOrigin(payload.Value())
				payload.AddIssue(issues.CreateTooBigIssue(maximum, true, origin, payload.Value()))
			}
		},
		OnAttach: []func(any){
			func(schema any) { mergeMaximumLengthConstraint(schema, maximum) },
		},
	}
}

// MinLength creates a minimum length validation check.
func MinLength(minimum int, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "min_length"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MinLength(payload.Value(), minimum) {
				origin := utils.LengthableOrigin(payload.Value())
				payload.AddIssue(issues.CreateTooSmallIssue(minimum, true, origin, payload.Value()))
			}
		},
		OnAttach: []func(any){
			func(schema any) { mergeMinimumLengthConstraint(schema, minimum) },
		},
	}
}

// Length creates an exact length validation check.
func Length(exact int, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "length_equals"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Length(payload.Value(), exact) {
				actualLength := getActualLength(payload.Value())
				origin := utils.LengthableOrigin(payload.Value())
				if actualLength > exact {
					payload.AddIssue(issues.CreateTooBigIssue(exact, true, origin, payload.Value()))
				} else {
					payload.AddIssue(issues.CreateTooSmallIssue(exact, true, origin, payload.Value()))
				}
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "minLength", exact)
				SetBagProperty(schema, "maxLength", exact)
			},
		},
	}
}

// MaxSize creates a maximum size validation check.
func MaxSize(maximum int, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "max_size"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MaxSize(payload.Value(), maximum) {
				origin := utils.SizableOrigin(payload.Value())
				payload.AddIssue(issues.CreateTooBigIssue(maximum, true, origin, payload.Value()))
			}
		},
		When: func(payload *core.ParsePayload) bool {
			return reflectx.HasSize(payload.Value()) || reflectx.HasLength(payload.Value())
		},
		OnAttach: []func(any){
			func(schema any) { setMaxSizeProperty(schema, maximum) },
		},
	}
}

// MinSize creates a minimum size validation check.
func MinSize(minimum int, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "min_size"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MinSize(payload.Value(), minimum) {
				origin := utils.SizableOrigin(payload.Value())
				payload.AddIssue(issues.CreateTooSmallIssue(minimum, true, origin, payload.Value()))
			}
		},
		When: func(payload *core.ParsePayload) bool {
			return reflectx.HasSize(payload.Value()) || reflectx.HasLength(payload.Value())
		},
		OnAttach: []func(any){
			func(schema any) { setMinSizeProperty(schema, minimum) },
		},
	}
}

// Size creates an exact size validation check.
func Size(exact int, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "size_equals"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Size(payload.Value(), exact) {
				if actualSize, ok := reflectx.Size(payload.Value()); ok {
					origin := utils.SizableOrigin(payload.Value())
					if actualSize > exact {
						payload.AddIssue(issues.CreateTooBigIssue(exact, true, origin, payload.Value()))
					} else {
						payload.AddIssue(issues.CreateTooSmallIssue(exact, true, origin, payload.Value()))
					}
				}
			}
		},
		When: func(payload *core.ParsePayload) bool {
			return reflectx.HasSize(payload.Value()) || reflectx.HasLength(payload.Value())
		},
		OnAttach: []func(any){
			func(schema any) {
				setMinSizeProperty(schema, exact)
				setMaxSizeProperty(schema, exact)
			},
		},
	}
}

// LengthRange creates a length range validation check.
func LengthRange(minimum, maximum int, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "length_range"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			actualLength := getActualLength(payload.Value())
			origin := utils.LengthableOrigin(payload.Value())
			if actualLength < minimum {
				payload.AddIssue(issues.CreateTooSmallIssue(minimum, true, origin, payload.Value()))
			} else if actualLength > maximum {
				payload.AddIssue(issues.CreateTooBigIssue(maximum, true, origin, payload.Value()))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				mergeMinimumLengthConstraint(schema, minimum)
				mergeMaximumLengthConstraint(schema, maximum)
			},
		},
	}
}

// SizeRange creates a size range validation check.
func SizeRange(minimum, maximum int, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "size_range"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if actualSize, ok := reflectx.Size(payload.Value()); ok {
				origin := utils.SizableOrigin(payload.Value())
				if actualSize > maximum {
					payload.AddIssue(issues.CreateTooBigIssue(maximum, true, origin, payload.Value()))
				} else if actualSize < minimum {
					payload.AddIssue(issues.CreateTooSmallIssue(minimum, true, origin, payload.Value()))
				}
			}
		},
		When: func(payload *core.ParsePayload) bool {
			return reflectx.HasSize(payload.Value()) || reflectx.HasLength(payload.Value())
		},
		OnAttach: []func(any){
			func(schema any) {
				setMinSizeProperty(schema, minimum)
				setMaxSizeProperty(schema, maximum)
			},
		},
	}
}

// NonEmpty creates a non-empty validation check (minimum length 1).
func NonEmpty(params ...any) core.ZodCheck { return MinLength(1, params...) }

// Empty creates an empty validation check (exact length 0).
func Empty(params ...any) core.ZodCheck { return Length(0, params...) }

// getActualLength returns the length of a value, or 0 if not measurable.
func getActualLength(v any) int {
	if l, ok := reflectx.Length(v); ok {
		return l
	}
	return 0
}

// mergeMinimumLengthConstraint merges minLength, choosing the stricter (larger) value.
func mergeMinimumLengthConstraint(schema any, length int) {
	mergeConstraint(schema, "minLength", length, func(old, new any) any {
		if o, ok := old.(int); ok {
			if n, ok := new.(int); ok && n > o {
				return n
			}
			return o
		}
		return new
	})
}

// mergeMaximumLengthConstraint merges maxLength, choosing the stricter (smaller) value.
func mergeMaximumLengthConstraint(schema any, length int) {
	mergeConstraint(schema, "maxLength", length, func(old, new any) any {
		if o, ok := old.(int); ok {
			if n, ok := new.(int); ok && n < o {
				return n
			}
			return o
		}
		return new
	})
}

// setMinSizeProperty sets the appropriate min size property based on schema type.
func setMinSizeProperty(schema any, size int) {
	bag := ensureBag(schema)
	if bag == nil {
		return
	}
	bag[sizePropertyKey(bag, "min")] = size
}

// setMaxSizeProperty sets the appropriate max size property based on schema type.
func setMaxSizeProperty(schema any, size int) {
	bag := ensureBag(schema)
	if bag == nil {
		return
	}
	bag[sizePropertyKey(bag, "max")] = size
}

// sizePropertyKey returns the JSON Schema property name for a size bound.
func sizePropertyKey(bag map[string]any, prefix string) string {
	if t, ok := bag["type"]; ok {
		switch t {
		case "array":
			return prefix + "Items"
		case "object":
			return prefix + "Properties"
		default:
			return prefix + "Length"
		}
	}
	// Default to Items for collections without explicit type.
	return prefix + "Items"
}
