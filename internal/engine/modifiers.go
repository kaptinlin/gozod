package engine

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// MODIFIER PROCESSING
// =============================================================================

// processModifiers handles modifier processing for nil input with performance optimization
// Returns early for non-nil input to avoid unnecessary processing overhead
func processModifiers[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	parseCore func(any) (any, error),
	ctx *core.ParseContext,
) (result any, handled bool, err error) {
	return processModifiersInternal[T](input, internals, expectedType, parseCore, ctx, false)
}

// processModifiersStrict is used for strict parsing where default values should not go through parseCore
func processModifiersStrict[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	parseCore func(any) (any, error),
	ctx *core.ParseContext,
) (result any, handled bool, err error) {
	return processModifiersInternal[T](input, internals, expectedType, parseCore, ctx, true)
}

// processModifiersInternal contains the actual implementation
func processModifiersInternal[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	parseCore func(any) (any, error),
	ctx *core.ParseContext,
	isStrict bool,
) (result any, handled bool, err error) {
	// Fast path: non-nil input doesn't need modifier processing
	if !isNilInput(input) {
		return nil, false, nil
	}

	// NonOptional flag moved to lower priority - allow Prefault to work first

	// Original logic: Default/DefaultFunc always has highest priority (short-circuit)
	// 1. Default/DefaultFunc - short-circuit mechanism, highest priority
	if internals.DefaultValue != nil {
		// Clone default value to prevent shared state issues
		clonedValue := cloneDefaultValue(internals.DefaultValue)

		// Default is a short-circuit mechanism, returns value directly without any parsing process (including Transform)
		if hasOverwriteCheck(internals.Checks) {
			result, err := ApplyChecks(clonedValue, internals.Checks, ctx)
			return result, true, err
		}
		return clonedValue, true, nil
	}

	if internals.DefaultFunc != nil {
		defaultValue := internals.DefaultFunc()
		// DefaultFunc is also a short-circuit mechanism, returns value directly without any parsing process (including Transform)
		if hasOverwriteCheck(internals.Checks) {
			result, err := ApplyChecks(defaultValue, internals.Checks, ctx)
			return result, true, err
		}
		return defaultValue, true, nil
	}

	// 2. Prefault/PrefaultFunc - preprocessing mechanism, higher priority than Optional/Nilable
	if internals.PrefaultValue != nil {
		// Prefault is a preprocessing mechanism, replaces input value and continues normal parsing process (including Transform)
		return internals.PrefaultValue, false, nil // handled=false means continue normal parsing
	}

	if internals.PrefaultFunc != nil {
		// PrefaultFunc is also a preprocessing mechanism, replaces input value and continues normal parsing process (including Transform)
		prefaultValue := internals.PrefaultFunc()
		return prefaultValue, false, nil // handled=false means continue normal parsing
	}

	// 3. NonOptional flag - higher priority than Optional/Nilable
	// Check type first - pointer types should naturally allow nil
	tType := reflect.TypeOf((*T)(nil)).Elem()
	isPointerType := tType.Kind() == reflect.Ptr

	if internals.NonOptional && !isPointerType {
		// NonOptional only applies to non-pointer types and takes precedence over Optional/Nilable
		return nil, true, issues.CreateNonOptionalError(ctx)
	}

	// 4. Optional/Nilable - allows nil values, lowest priority
	if internals.Optional || internals.Nilable {
		// Apply only overwrite and refine checks to nil input, not format checks
		if hasNilApplicableCheck(internals.Checks) {
			nilApplicableChecks := filterNilApplicableChecks(internals.Checks)
			result, err := ApplyChecks[any](nil, nilApplicableChecks, ctx)
			return result, true, err
		}
		return nil, true, nil
	}

	// 5. Special handling for pointer types - pointer types naturally allow nil
	if isPointerType {
		// This is a pointer type, nil should be allowed
		if hasNilApplicableCheck(internals.Checks) {
			nilApplicableChecks := filterNilApplicableChecks(internals.Checks)
			result, err := ApplyChecks[any](nil, nilApplicableChecks, ctx)
			return result, true, err
		}
		return nil, true, nil
	}

	// 5. Special handling for Unknown type
	if expectedType == core.ZodTypeUnknown {
		return nil, true, nil
	}

	// No fallback mechanism available
	return nil, true, issues.CreateInvalidTypeError(expectedType, input, ctx)
}

// applyTransformIfPresent applies transformation if configured, with performance optimization
// Uses fast path when no transformation is needed to minimize overhead
func applyTransformIfPresent(result any, internals *core.ZodTypeInternals, ctx *core.ParseContext) (any, error) {
	if internals.Transform == nil {
		return result, nil // Fast path: no transform
	}

	refCtx := &core.RefinementContext{ParseContext: ctx}
	return internals.Transform(result, refCtx)
}

// =============================================================================
// HELPERS
// =============================================================================

// hasNilApplicableCheck checks if there are any checks that should be applied to nil values
// Only overwrite, refine, and custom checks should be applied to nil values, not format checks
func hasNilApplicableCheck(checks []core.ZodCheck) bool {
	for _, check := range checks {
		if check == nil {
			continue
		}
		checkInternals := check.GetZod()
		if checkInternals == nil || checkInternals.Def == nil {
			continue
		}

		// Check if this is an overwrite, refine, or custom check
		checkType := checkInternals.Def.Check
		if checkType == "overwrite" || checkType == "refine" || checkType == "custom" {
			return true
		}
	}
	return false
}

// filterNilApplicableChecks filters checks to only include those that should be applied to nil values
// Only overwrite, refine, and custom checks should be applied to nil values, not format checks
func filterNilApplicableChecks(checks []core.ZodCheck) []core.ZodCheck {
	var nilApplicableChecks []core.ZodCheck
	for _, check := range checks {
		if check == nil {
			continue
		}
		checkInternals := check.GetZod()
		if checkInternals == nil || checkInternals.Def == nil {
			continue
		}

		// Check if this is an overwrite, refine, or custom check
		checkType := checkInternals.Def.Check
		if checkType == "overwrite" || checkType == "refine" || checkType == "custom" {
			nilApplicableChecks = append(nilApplicableChecks, check)
		}
	}
	return nilApplicableChecks
}

// cloneDefaultValue creates a shallow copy of map/slice default values
func cloneDefaultValue(v any) any {
	if v == nil {
		return nil
	}

	// Try slice cloning
	if reflectx.IsSlice(v) {
		rv := reflect.ValueOf(v)
		// Create new slice with same type, length and capacity
		newSlice := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Cap())
		// Copy elements
		reflect.Copy(newSlice, rv)
		return newSlice.Interface()
	}

	// Try map cloning
	if reflectx.IsMap(v) {
		if m, ok := v.(map[string]any); ok {
			return mapx.Copy(m)
		}
		// Basic copy for generic maps using reflection (fallback)
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
