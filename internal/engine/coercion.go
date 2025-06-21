package engine

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/mapx"
)

// =============================================================================
// COERCION UTILITIES
// =============================================================================

// ShouldCoerce checks if coercion is enabled in the schema bag
// This reduces repetitive coercion checking code across type files
func ShouldCoerce(bag map[string]any) bool {
	return mapx.GetBoolDefault(bag, "coerce", false)
}

// EnableCoercionForType sets coercion flag for given schema
// This utility function enables coercion on any schema type
func EnableCoercionForType(schema core.ZodType[any, any]) {
	internals := schema.GetInternals()
	if internals == nil {
		return
	}

	if ShouldCoerce(internals.Bag) {
		return // Already enabled
	}

	if internals.Bag == nil {
		internals.Bag = make(map[string]any)
	}
	mapx.Set(internals.Bag, "coerce", true)
}

// TryApplyCoercion attempts to apply coercion to input if enabled
// Elegant coercion interface design: each type handles its own coerce logic
// Avoids large switch statements in core parsing layer, follows open-closed principle
func TryApplyCoercion(schema core.ZodType[any, any], input any) (any, error) {
	internals := schema.GetInternals()
	if internals == nil {
		return input, nil
	}

	// Check if coerce is enabled - use unified ShouldCoerce function
	if !ShouldCoerce(internals.Bag) {
		return input, nil
	}

	// Elegant interface call - types handle their own coerce logic
	if coercible, ok := schema.(core.Coercible); ok {
		if coercedValue, success := coercible.Coerce(input); success {
			return coercedValue, nil
		}
	}

	// Coercion failed or not implemented, return original value for subsequent validation
	return input, nil
}

// TryApplyCoercionGeneric attempts to apply coercion with generic type safety
// Enhanced version that uses coerce package for better type conversion
func TryApplyCoercionGeneric[T any](schema core.ZodType[any, any], input any) (T, error) {
	var zero T
	internals := schema.GetInternals()
	if internals == nil {
		return zero, nil
	}

	// Use mapx to check coerce flag
	if !mapx.GetBoolDefault(internals.Bag, "coerce", false) {
		return zero, nil
	}

	// Use coerce package for type conversion
	if result, err := coerce.To[T](input); err == nil {
		return result, nil
	}

	// If coerce fails, try using schema's custom coerce method
	if coercible, ok := schema.(core.Coercible); ok {
		if coercedValue, success := coercible.Coerce(input); success {
			if typed, err := coerce.To[T](coercedValue); err == nil {
				return typed, nil
			}
		}
	}

	return zero, ErrCannotConvertType
}
