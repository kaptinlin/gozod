package engine

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// processModifiers handles modifier processing for nil input.
// Returns early for non-nil input.
func processModifiers[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	_ func(any) (any, error),
	ctx *core.ParseContext,
) (result any, handled bool, err error) {
	return processModifiersCore[T](input, internals, expectedType, ctx)
}

// processModifiersStrict is the strict variant where default values
// skip parseCore.
func processModifiersStrict[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	_ func(any) (any, error),
	ctx *core.ParseContext,
) (result any, handled bool, err error) {
	return processModifiersCore[T](input, internals, expectedType, ctx)
}

// processModifiersCore contains the shared modifier logic.
// Priority order: Default > Prefault > NonOptional > Optional/Nilable > Pointer > Unknown.
func processModifiersCore[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (result any, handled bool, err error) {
	if !isNilInput(input) {
		return nil, false, nil
	}

	// 1. Default/DefaultFunc — short-circuit, highest priority
	if internals.DefaultValue != nil {
		clonedValue := cloneDefaultValue(internals.DefaultValue)
		if hasOverwriteCheck(internals.Checks) {
			result, err := ApplyChecks(clonedValue, internals.Checks, ctx)
			return result, true, err
		}
		return clonedValue, true, nil
	}
	if internals.DefaultFunc != nil {
		defaultValue := internals.DefaultFunc()
		if hasOverwriteCheck(internals.Checks) {
			result, err := ApplyChecks(defaultValue, internals.Checks, ctx)
			return result, true, err
		}
		return defaultValue, true, nil
	}

	// 2. Prefault/PrefaultFunc — preprocessing, continues normal parsing
	if internals.PrefaultValue != nil {
		return internals.PrefaultValue, false, nil
	}
	if internals.PrefaultFunc != nil {
		return internals.PrefaultFunc(), false, nil
	}

	// 3. NonOptional — higher priority than Optional/Nilable
	tType := reflect.TypeOf((*T)(nil)).Elem()
	isPointerType := tType.Kind() == reflect.Ptr

	if internals.NonOptional && !isPointerType {
		return nil, true, issues.CreateNonOptionalError(ctx)
	}

	// 4. Optional/Nilable — allows nil values
	if internals.Optional || internals.Nilable {
		if nilChecks := filterNilApplicableChecks(internals.Checks); len(nilChecks) > 0 {
			result, err := ApplyChecks[any](nil, nilChecks, ctx)
			return result, true, err
		}
		return nil, true, nil
	}

	// 5. Pointer types naturally allow nil
	if isPointerType {
		if nilChecks := filterNilApplicableChecks(internals.Checks); len(nilChecks) > 0 {
			result, err := ApplyChecks[any](nil, nilChecks, ctx)
			return result, true, err
		}
		return nil, true, nil
	}

	// 6. Unknown type allows nil
	if expectedType == core.ZodTypeUnknown {
		return nil, true, nil
	}

	return nil, true, issues.CreateInvalidTypeError(expectedType, input, ctx)
}

// applyTransformIfPresent applies the transform function if configured.
func applyTransformIfPresent(result any, internals *core.ZodTypeInternals, ctx *core.ParseContext) (any, error) {
	if internals.Transform == nil {
		return result, nil
	}
	refCtx := &core.RefinementContext{ParseContext: ctx}
	return internals.Transform(result, refCtx)
}

// filterNilApplicableChecks returns only checks applicable to nil values
// (overwrite, refine, and custom checks).
func filterNilApplicableChecks(checks []core.ZodCheck) []core.ZodCheck {
	var result []core.ZodCheck
	for _, check := range checks {
		if check == nil {
			continue
		}
		ci := check.GetZod()
		if ci == nil || ci.Def == nil {
			continue
		}
		switch ci.Def.Check {
		case "overwrite", "refine", "custom":
			result = append(result, check)
		}
	}
	return result
}

// cloneDefaultValue creates a shallow copy of map/slice default values.
func cloneDefaultValue(v any) any {
	if v == nil {
		return nil
	}

	if reflectx.IsSlice(v) {
		rv := reflect.ValueOf(v)
		newSlice := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Cap())
		reflect.Copy(newSlice, rv)
		return newSlice.Interface()
	}

	if reflectx.IsMap(v) {
		if m, ok := v.(map[string]any); ok {
			return mapx.Copy(m)
		}
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Map {
			newMap := reflect.MakeMap(rv.Type())
			iter := rv.MapRange()
			for iter.Next() {
				newMap.SetMapIndex(iter.Key(), iter.Value())
			}
			return newMap.Interface()
		}
	}

	return v
}
