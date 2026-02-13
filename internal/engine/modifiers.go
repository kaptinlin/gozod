package engine

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// ----------------------------------------------------------------------------
// Modifier processing
// ----------------------------------------------------------------------------

// processModifiers handles modifier processing for nil input.
// It returns early for non-nil input. The parseCore parameter is accepted
// for API compatibility but unused -- all modifier logic is in processModifiersCore.
func processModifiers[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	_ func(any) (any, error),
	ctx *core.ParseContext,
) (any, bool, error) {
	return processModifiersCore[T](input, internals, expectedType, ctx)
}

// processModifiersStrict delegates to the same core logic as processModifiers.
// It exists as a separate function for call-site clarity in strict parse paths.
func processModifiersStrict[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	_ func(any) (any, error),
	ctx *core.ParseContext,
) (any, bool, error) {
	return processModifiersCore[T](input, internals, expectedType, ctx)
}

// processModifiersCore contains the shared modifier logic.
// Priority: Default > Prefault > NonOptional > Optional/Nilable > Pointer > Unknown.
func processModifiersCore[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (any, bool, error) {
	if !isNilInput(input) {
		return nil, false, nil
	}

	// Default/DefaultFunc — short-circuit, highest priority.
	if internals.DefaultValue != nil {
		v := cloneDefaultValue(internals.DefaultValue)
		if hasOverwriteCheck(internals.Checks) {
			r, err := ApplyChecks(v, internals.Checks, ctx)
			return r, true, err
		}
		return v, true, nil
	}
	if internals.DefaultFunc != nil {
		v := internals.DefaultFunc()
		if hasOverwriteCheck(internals.Checks) {
			r, err := ApplyChecks(v, internals.Checks, ctx)
			return r, true, err
		}
		return v, true, nil
	}

	// Prefault/PrefaultFunc — preprocessing, continues normal parsing.
	if internals.PrefaultValue != nil {
		return internals.PrefaultValue, false, nil
	}
	if internals.PrefaultFunc != nil {
		return internals.PrefaultFunc(), false, nil
	}

	// NonOptional — higher priority than Optional/Nilable.
	isPtr := reflect.TypeFor[T]().Kind() == reflect.Ptr

	if internals.NonOptional && !isPtr {
		return nil, true, issues.CreateNonOptionalError(ctx)
	}

	// Optional/Nilable and pointer types naturally allow nil values.
	if internals.Optional || internals.Nilable || isPtr {
		if nc := filterNilChecks(internals.Checks); len(nc) > 0 {
			r, err := ApplyChecks[any](nil, nc, ctx)
			return r, true, err
		}
		return nil, true, nil
	}

	// Unknown type allows nil.
	if expectedType == core.ZodTypeUnknown {
		return nil, true, nil
	}

	return nil, true, issues.CreateInvalidTypeError(expectedType, input, ctx)
}

// ----------------------------------------------------------------------------
// Transform and utility helpers
// ----------------------------------------------------------------------------

// applyTransformIfPresent applies the transform function if set.
func applyTransformIfPresent(
	result any,
	internals *core.ZodTypeInternals,
	ctx *core.ParseContext,
) (any, error) {
	if internals.Transform == nil {
		return result, nil
	}
	return internals.Transform(result, &core.RefinementContext{ParseContext: ctx})
}

// filterNilChecks returns only checks applicable to nil values
// (overwrite, refine, and custom checks).
func filterNilChecks(checks []core.ZodCheck) []core.ZodCheck {
	var out []core.ZodCheck
	for _, c := range checks {
		if c == nil {
			continue
		}
		ci := c.Zod()
		if ci == nil || ci.Def == nil {
			continue
		}
		switch ci.Def.Check {
		case "overwrite", "refine", "custom":
			out = append(out, c)
		}
	}
	return out
}

// cloneDefaultValue creates a shallow copy of map/slice default values.
func cloneDefaultValue(v any) any {
	if v == nil {
		return nil
	}

	if reflectx.IsSlice(v) {
		rv := reflect.ValueOf(v)
		s := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Cap())
		reflect.Copy(s, rv)
		return s.Interface()
	}

	if reflectx.IsMap(v) {
		if m, ok := v.(map[string]any); ok {
			return mapx.Copy(m)
		}
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Map {
			m := reflect.MakeMap(rv.Type())
			iter := rv.MapRange()
			for iter.Next() {
				m.SetMapIndex(iter.Key(), iter.Value())
			}
			return m.Interface()
		}
	}

	return v
}
