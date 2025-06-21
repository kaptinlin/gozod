package coerce

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// BASIC String CONVERSION
// =============================================================================

// ToString converts any value to string with fast-path optimizations
func ToString(v any) (string, error) {
	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return "", fmt.Errorf("cannot convert nil pointer to string: %w", ErrNilPointer)
	}

	// Fast path for common types - zero reflection overhead
	switch x := derefed.(type) {
	case string:
		return x, nil
	case int:
		return strconv.Itoa(x), nil
	case int8:
		return strconv.FormatInt(int64(x), 10), nil
	case int16:
		return strconv.FormatInt(int64(x), 10), nil
	case int32:
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case uint:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint64:
		return strconv.FormatUint(x, 10), nil
	case float32:
		return strconv.FormatFloat(float64(x), 'g', -1, 32), nil
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64), nil
	case bool:
		return strconv.FormatBool(x), nil
	case []byte:
		return string(x), nil
	case complex64:
		return fmt.Sprintf("%g", x), nil
	case complex128:
		return fmt.Sprintf("%g", x), nil
	default:
		// for complex/uncommon types, return unsupported error
		return "", NewUnsupportedError(fmt.Sprintf("%T", derefed), "string")
	}
}

// =============================================================================
// TIME/DATE CONVERSION UTILITIES
// =============================================================================

// ToISODate converts various date inputs to ISO date string
func ToISODate(value any) (string, error) {
	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(value)
	if !ok {
		return "", fmt.Errorf("cannot convert nil pointer to ISO date: %w", ErrNilPointer)
	}

	switch v := derefed.(type) {
	case time.Time:
		// Format as ISO date (YYYY-MM-DD)
		return v.Format("2006-01-02"), nil
	case string:
		// Attempt multiple common layouts and then output canonical ISO date
		layouts := []string{
			time.RFC3339,
			"2006-01-02",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006/01/02",
			"01/02/2006",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, v); err == nil {
				return t.Format("2006-01-02"), nil
			}
		}
		return "", NewFormatError(v, "ISO date")
	default:
		return "", NewUnsupportedError(fmt.Sprintf("%T", derefed), "ISO date")
	}
}
