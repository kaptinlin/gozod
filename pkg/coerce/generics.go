package coerce

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// GENERIC CONVERSION ENTRY POINT
// =============================================================================

// To is a generic conversion function that attempts to convert any value
// to the specified target type T using type parameters for type-safe conversions
func To[T any](v any) (T, error) {
	var zero T

	// Get the target type information
	targetType := reflect.TypeOf(zero)

	// Handle nil target type (should not happen with valid generic usage)
	if targetType == nil {
		return zero, NewUnsupportedError("unknown", "unknown")
	}

	// Dispatch to appropriate conversion function based on target type
	switch any(zero).(type) {
	case string:
		result, err := ToString(v)
		if err != nil {
			return zero, err
		}
		return any(result).(T), nil

	case bool:
		result, err := ToBool(v)
		if err != nil {
			return zero, err
		}
		return any(result).(T), nil

	case int64:
		result, err := ToInt64(v)
		if err != nil {
			return zero, err
		}
		return any(result).(T), nil

	case float64:
		result, err := ToFloat64(v)
		if err != nil {
			return zero, err
		}
		return any(result).(T), nil

	case *big.Int:
		result, err := ToBigInt(v)
		if err != nil {
			return zero, err
		}
		return any(result).(T), nil

	case map[any]any:
		result, err := ToMap(v)
		if err != nil {
			return zero, err
		}
		return any(result).(T), nil

	case map[string]any:
		result, err := ToObject(v)
		if err != nil {
			return zero, err
		}
		return any(result).(T), nil

	case []any:
		result, err := ToSlice(v)
		if err != nil {
			return zero, err
		}
		return any(result).(T), nil

	default:
		// For other types, delegate to reflectx
		return reflectx.ConvertToGeneric[T](v)
	}
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// ToValueSet converts a slice to a set (map[any]bool) for enum validation
func ToValueSet(slice any) (map[any]bool, error) {
	if slice == nil {
		return nil, NewUnsupportedError("nil", "value set")
	}

	sliceAny, err := ToSlice(slice)
	if err != nil {
		return nil, fmt.Errorf("cannot convert to slice for value set: %w", err)
	}

	result := make(map[any]bool)
	for _, value := range sliceAny {
		result[value] = true
	}
	return result, nil
}

// ToLiteral attempts to coerce common textual representations into their primitive literal values.
// For example "42" -> 42, "3.14" -> 3.14, "true" -> true.
// If no meaningful coercion is possible the original value is returned.
func ToLiteral(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return value, nil
		}

		// Try boolean first
		if b, err := strconv.ParseBool(trimmed); err == nil {
			return b, nil
		}

		// Try integer
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			// Special-case 1/0 to boolean for literal coercion convenience
			if i == 1 {
				return true, nil
			}
			if i == 0 {
				return false, nil
			}

			// Fits into native int range?
			maxInt := int64(^uint(0) >> 1)
			minInt := -maxInt - 1
			if i >= minInt && i <= maxInt {
				return int(i), nil
			}
			return i, nil
		}

		// Try float
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			// If the float is an integer value (e.g., 42.0) convert to int to match typical literal usage
			if math.Mod(f, 1) == 0 {
				return int(f), nil
			}
			return f, nil
		}
		return v, nil
	default:
		// Handle float to int conversion for integer values
		switch n := value.(type) {
		case float32:
			// If the float is an integer value, convert to int
			if math.Mod(float64(n), 1) == 0 {
				if n == 1.0 {
					return true, nil
				}
				if n == 0.0 {
					return false, nil
				}
				return int(n), nil
			}
		case float64:
			// If the float is an integer value, convert to int
			if math.Mod(n, 1) == 0 {
				if n == 1.0 {
					return true, nil
				}
				if n == 0.0 {
					return false, nil
				}
				return int(n), nil
			}
		}

		// Handle numeric to boolean conversion (1 -> true, 0 -> false)
		switch n := value.(type) {
		case int:
			if n == 1 {
				return true, nil
			}
			if n == 0 {
				return false, nil
			}
		case int8, int16, int32, int64:
			if reflect.DeepEqual(n, int64(1)) || reflect.DeepEqual(n, int8(1)) || reflect.DeepEqual(n, int16(1)) || reflect.DeepEqual(n, int32(1)) {
				return true, nil
			}
			if reflect.DeepEqual(n, int64(0)) || reflect.DeepEqual(n, int8(0)) || reflect.DeepEqual(n, int16(0)) || reflect.DeepEqual(n, int32(0)) {
				return false, nil
			}
		case uint, uint8, uint16, uint32, uint64:
			rv := reflect.ValueOf(n).Convert(reflect.TypeOf(uint64(0))).Uint()
			if rv == 1 {
				return true, nil
			}
			if rv == 0 {
				return false, nil
			}
		}
		return value, nil
	}
}
