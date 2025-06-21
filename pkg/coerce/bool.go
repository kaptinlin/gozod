package coerce

import (
	"fmt"
	"strings"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// BOOLEAN CONVERSION
// =============================================================================

// ToBool converts any value to boolean with fast-path optimizations
func ToBool(v any) (bool, error) {
	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return false, fmt.Errorf("cannot convert nil pointer to bool: %w", ErrNilPointer)
	}

	// Fast path for common types
	switch x := derefed.(type) {
	case bool:
		return x, nil
	case string:
		return stringToBool(x)
	case int:
		return x != 0, nil
	case int8:
		return x != 0, nil
	case int16:
		return x != 0, nil
	case int32:
		return x != 0, nil
	case int64:
		return x != 0, nil
	case uint:
		return x != 0, nil
	case uint8:
		return x != 0, nil
	case uint16:
		return x != 0, nil
	case uint32:
		return x != 0, nil
	case uint64:
		return x != 0, nil
	case float32:
		return x != 0, nil
	case float64:
		return x != 0, nil
	default:
		return false, NewUnsupportedError(fmt.Sprintf("%T", derefed), "bool")
	}
}

// stringToBool converts string to boolean using optimized standard mappings
func stringToBool(s string) (bool, error) {
	// Trim leading/trailing whitespace efficiently
	s = strings.TrimSpace(s)

	// Normalize case to lower for simpler comparison so that mixed-case variants
	// like "TrUe" or "FaLsE" are handled gracefully.
	lower := strings.ToLower(s)

	switch lower {
	case "true", "1", "yes", "on", "y":
		return true, nil
	case "false", "0", "no", "off", "n", "":
		return false, nil
	default:
		return false, NewFormatError(s, "bool")
	}
}
