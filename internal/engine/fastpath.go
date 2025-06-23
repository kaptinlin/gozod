package engine

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
)

// =============================================================================
// FAST PATH PARSING FOR BASIC TYPES
// =============================================================================

// ParseStringFast provides fast path parsing for string types
// Uses direct type assertion to avoid reflection overhead
// Returns any to preserve smart type inference (string -> string, *string -> *string)
func ParseStringFast(input any, internals *core.ZodTypeInternals) (any, error) {
	_, isNil, _ := PreprocessInput(input)

	if isNil {
		if !internals.IsNilable() {
			rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeString), nil)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return nil, nil // Return nil for nil when nilable
	}

	// Direct type assertion for string - preserve original input type
	if str, ok := input.(string); ok {
		return str, nil // Return original string
	}

	// Handle string pointers - preserve pointer type
	if strPtr, ok := input.(*string); ok && strPtr != nil {
		return strPtr, nil // Return original pointer
	}

	rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeString), input)
	finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
	return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// ParseBoolFast provides fast path parsing for boolean types
// Uses direct type assertion to avoid reflection overhead
func ParseBoolFast(input any, internals *core.ZodTypeInternals) (bool, error) {
	dereferenced, isNil, _ := PreprocessInput(input)

	if isNil {
		if !internals.IsNilable() {
			rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeBool), nil)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
			return false, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return false, nil // Return false for nil when nilable
	}

	// Direct type assertion for bool
	if b, ok := dereferenced.(bool); ok {
		return b, nil
	}

	// Handle bool pointers
	if bPtr, ok := dereferenced.(*bool); ok && bPtr != nil {
		return *bPtr, nil
	}

	rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeBool), input)
	finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
	return false, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// ParseIntFast provides fast path parsing for integer types
// Uses direct type assertion to avoid reflection overhead
func ParseIntFast(input any, internals *core.ZodTypeInternals) (int, error) {
	dereferenced, isNil, _ := PreprocessInput(input)

	if isNil {
		if !internals.IsNilable() {
			rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeInteger), nil)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
			return 0, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return 0, nil // Return 0 for nil when nilable
	}

	// Direct type assertion for int
	if i, ok := dereferenced.(int); ok {
		return i, nil
	}

	// Handle int pointers
	if iPtr, ok := dereferenced.(*int); ok && iPtr != nil {
		return *iPtr, nil
	}

	rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeInteger), input)
	finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
	return 0, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// ParseFloatFast provides fast path parsing for float64 types
// Uses direct type assertion to avoid reflection overhead
func ParseFloatFast(input any, internals *core.ZodTypeInternals) (float64, error) {
	dereferenced, isNil, _ := PreprocessInput(input)

	if isNil {
		if !internals.IsNilable() {
			rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeFloat64), nil)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
			return 0.0, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return 0.0, nil // Return 0.0 for nil when nilable
	}

	// Direct type assertion for float64
	if f, ok := dereferenced.(float64); ok {
		return f, nil
	}

	// Handle float64 pointers
	if fPtr, ok := dereferenced.(*float64); ok && fPtr != nil {
		return *fPtr, nil
	}

	// Also handle float32
	if f32, ok := dereferenced.(float32); ok {
		return float64(f32), nil
	}

	rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeFloat64), input)
	finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
	return 0.0, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// ParseNumberFast provides fast path parsing for numeric types
// Handles both integers and floats with type assertion
func ParseNumberFast(input any, internals *core.ZodTypeInternals) (float64, error) {
	dereferenced, isNil, _ := PreprocessInput(input)

	if isNil {
		if !internals.IsNilable() {
			rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeNumber), nil)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
			return 0.0, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return 0.0, nil // Return 0.0 for nil when nilable
	}

	// Handle various numeric types
	switch v := dereferenced.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	}

	rawIssue := issues.CreateInvalidTypeIssue(string(core.ZodTypeNumber), input)
	finalIssue := issues.FinalizeIssue(rawIssue, nil, nil)
	return 0.0, issues.NewZodError([]core.ZodIssue{finalIssue})
}
