package coerce

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"strconv"
	"strings"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// INTEGER CONVERSION
// =============================================================================

// ToInt64 converts any value to int64 with overflow checking
func ToInt64(v any) (int64, error) {
	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return 0, fmt.Errorf("cannot convert nil pointer to int64: %w", ErrNilPointer)
	}

	// Fast path for integer types
	switch x := derefed.(type) {
	case int64:
		return x, nil
	case int:
		return int64(x), nil
	case int8:
		return int64(x), nil
	case int16:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case uint:
		if x > math.MaxInt64 {
			return 0, NewOverflowError(x, "int64")
		}
		return int64(x), nil
	case uint8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case uint64:
		if x > math.MaxInt64 {
			return 0, NewOverflowError(x, "int64")
		}
		return int64(x), nil
	case float32:
		if math.Trunc(float64(x)) != float64(x) {
			return 0, NewNotWholeError(x)
		}
		return int64(x), nil
	case float64:
		if math.Trunc(x) != x {
			return 0, NewNotWholeError(x)
		}
		if x > math.MaxInt64 || x < math.MinInt64 {
			return 0, NewOverflowError(x, "int64")
		}
		return int64(x), nil
	case string:
		return stringToInt64(x)
	case bool:
		if x {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, NewUnsupportedError(fmt.Sprintf("%T", derefed), "int64")
	}
}

// stringToInt64 converts string to int64 with validation
func stringToInt64(s string) (int64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, NewEmptyInputError("int64")
	}

	i, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, NewFormatError(s, "int64")
	}
	return i, nil
}

// =============================================================================
// FLOAT CONVERSION
// =============================================================================

// ToFloat64 converts any value to float64
func ToFloat64(v any) (float64, error) {
	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return 0, fmt.Errorf("cannot convert nil pointer to float64: %w", ErrNilPointer)
	}

	// Fast path for numeric types
	switch x := derefed.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int8:
		return float64(x), nil
	case int16:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case uint:
		return float64(x), nil
	case uint8:
		return float64(x), nil
	case uint16:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case *big.Int:
		if x == nil {
			return 0, fmt.Errorf("cannot convert nil *big.Int to float64: %w", ErrNilPointer)
		}
		// Convert big.Int to float64
		// Note: This may lose precision for very large numbers
		f, _ := x.Float64()
		return f, nil
	case big.Int:
		// Convert big.Int to float64
		// Note: This may lose precision for very large numbers
		f, _ := x.Float64()
		return f, nil
	case complex64:
		// Convert complex64 to magnitude (float64)
		return cmplx.Abs(complex128(x)), nil
	case complex128:
		// Convert complex128 directly to magnitude
		return cmplx.Abs(x), nil
	case string:
		return stringToFloat64(x)
	case bool:
		if x {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, NewUnsupportedError(fmt.Sprintf("%T", derefed), "float64")
	}
}

// stringToFloat64 converts string to float64 with validation
func stringToFloat64(s string) (float64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, NewEmptyInputError("float64")
	}

	f, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, NewFormatError(s, "float64")
	}
	return f, nil
}

// =============================================================================
// BIG INTEGER CONVERSION
// =============================================================================

// ToBigInt converts any value to *big.Int
func ToBigInt(v any) (*big.Int, error) {
	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return nil, fmt.Errorf("cannot convert nil pointer to *big.Int: %w", ErrNilPointer)
	}

	switch x := derefed.(type) {
	case *big.Int:
		if x == nil {
			return nil, fmt.Errorf("cannot convert nil *big.Int: %w", ErrNilPointer)
		}
		// Return a copy to avoid mutation of the original
		return new(big.Int).Set(x), nil
	case int64:
		return big.NewInt(x), nil
	case int:
		return big.NewInt(int64(x)), nil
	case int8:
		return big.NewInt(int64(x)), nil
	case int16:
		return big.NewInt(int64(x)), nil
	case int32:
		return big.NewInt(int64(x)), nil
	case uint64:
		return new(big.Int).SetUint64(x), nil
	case uint:
		return new(big.Int).SetUint64(uint64(x)), nil
	case uint8:
		return big.NewInt(int64(x)), nil
	case uint16:
		return big.NewInt(int64(x)), nil
	case uint32:
		return big.NewInt(int64(x)), nil
	case float32:
		if float32(int64(x)) != x {
			return nil, NewNotWholeError(x)
		}
		return big.NewInt(int64(x)), nil
	case float64:
		if float64(int64(x)) != x {
			return nil, NewNotWholeError(x)
		}
		return big.NewInt(int64(x)), nil
	case string:
		return stringToBigInt(x)
	case bool:
		if x {
			return big.NewInt(1), nil
		}
		return big.NewInt(0), nil
	default:
		return nil, NewUnsupportedError(fmt.Sprintf("%T", derefed), "*big.Int")
	}
}

// stringToBigInt converts string to *big.Int with multiple base support
func stringToBigInt(s string) (*big.Int, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil, NewEmptyInputError("*big.Int")
	}

	bigInt := new(big.Int)

	// Try decimal first
	if _, ok := bigInt.SetString(trimmed, 10); ok {
		return bigInt, nil
	}

	// Try hexadecimal for 0x prefixed strings
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		if _, ok := bigInt.SetString(trimmed[2:], 16); ok {
			return bigInt, nil
		}
	}

	return nil, NewFormatError(s, "big integer")
}

// =============================================================================
// GENERIC CONVERSION WITH BOUNDS CHECKING
// =============================================================================

// ToInteger converts any value to a specific integer type with bounds checking
func ToInteger[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](v any) (T, error) {
	var zero T

	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return zero, fmt.Errorf("cannot convert nil pointer to %T: %w", zero, ErrNilPointer)
	}

	// Fast path for exact type match
	if result, ok := derefed.(T); ok {
		return result, nil
	}

	// Convert to int64 first, then check bounds
	var int64Val int64
	var err error

	switch x := derefed.(type) {
	case string:
		int64Val, err = stringToInt64(x)
		if err != nil {
			return zero, err
		}
	case int:
		int64Val = int64(x)
	case int8:
		int64Val = int64(x)
	case int16:
		int64Val = int64(x)
	case int32:
		int64Val = int64(x)
	case int64:
		int64Val = x
	case uint:
		if x > math.MaxInt64 {
			return zero, NewOverflowError(x, "int64")
		}
		int64Val = int64(x)
	case uint8:
		int64Val = int64(x)
	case uint16:
		int64Val = int64(x)
	case uint32:
		int64Val = int64(x)
	case uint64:
		if x > math.MaxInt64 {
			return zero, NewOverflowError(x, "int64")
		}
		int64Val = int64(x)
	case float32:
		if math.Trunc(float64(x)) != float64(x) {
			return zero, NewNotWholeError(x)
		}
		int64Val = int64(x)
	case float64:
		if math.Trunc(x) != x {
			return zero, NewNotWholeError(x)
		}
		if x > math.MaxInt64 || x < math.MinInt64 {
			return zero, NewOverflowError(x, "int64")
		}
		int64Val = int64(x)
	case bool:
		// Booleans should not be coerced to integers to avoid unintended conversions
		return zero, NewUnsupportedError("bool", fmt.Sprintf("%T", zero))
	default:
		return zero, NewUnsupportedError(fmt.Sprintf("%T", derefed), fmt.Sprintf("%T", zero))
	}

	// Check bounds for target type T
	if err := checkIntegerTypeBounds(int64Val, zero); err != nil {
		return zero, err
	}

	return T(int64Val), nil
}

// ToFloat converts any value to a specific float type
func ToFloat[T ~float32 | ~float64](v any) (T, error) {
	var zero T

	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return zero, fmt.Errorf("cannot convert nil pointer to %T: %w", zero, ErrNilPointer)
	}

	// Fast path for exact type match
	if result, ok := derefed.(T); ok {
		return result, nil
	}

	// Convert to float64 first
	float64Val, err := ToFloat64(derefed)
	if err != nil {
		return zero, err
	}

	// Check bounds for float32 explicitly
	if _, ok := any(zero).(float32); ok {
		if math.Abs(float64Val) > math.MaxFloat32 {
			return zero, NewOverflowError(float64Val, "float32")
		}
	}

	return T(float64Val), nil
}

// checkIntegerTypeBounds verifies that an int64 value fits in the target integer type
func checkIntegerTypeBounds[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](value int64, target T) error {
	switch any(target).(type) {
	case int8:
		if value < -128 || value > 127 {
			return NewOverflowError(value, "int8")
		}
	case int16:
		if value < -32768 || value > 32767 {
			return NewOverflowError(value, "int16")
		}
	case int32:
		if value < -2147483648 || value > 2147483647 {
			return NewOverflowError(value, "int32")
		}
	case int:
		// Platform-dependent check for int
		if value < -2147483648 || value > 2147483647 {
			return NewOverflowError(value, "int (32-bit platforms)")
		}
	case uint8:
		if value < 0 || value > 255 {
			if value < 0 {
				return NewNegativeUintError(value, "uint8")
			}
			return NewOverflowError(value, "uint8")
		}
	case uint16:
		if value < 0 || value > 65535 {
			if value < 0 {
				return NewNegativeUintError(value, "uint16")
			}
			return NewOverflowError(value, "uint16")
		}
	case uint32:
		if value < 0 || value > 4294967295 {
			if value < 0 {
				return NewNegativeUintError(value, "uint32")
			}
			return NewOverflowError(value, "uint32")
		}
	case uint:
		if value < 0 || value > 4294967295 {
			if value < 0 {
				return NewNegativeUintError(value, "uint")
			}
			return NewOverflowError(value, "uint (32-bit platforms)")
		}
	case uint64:
		if value < 0 {
			return NewNegativeUintError(value, "uint64")
		}
	}
	return nil
}

// =============================================================================
// COMPLEX NUMBER CONVERSION
// =============================================================================

// ToComplex64 converts any value to complex64
func ToComplex64(v any) (complex64, error) {
	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return 0, fmt.Errorf("cannot convert nil pointer to complex64: %w", ErrNilPointer)
	}

	// Fast path for exact type match
	if result, ok := derefed.(complex64); ok {
		return result, nil
	}

	// Handle other complex types
	if c128, ok := derefed.(complex128); ok {
		return complex64(c128), nil
	}

	// Handle numeric types
	switch val := derefed.(type) {
	case int:
		return complex(float32(val), 0), nil
	case int8:
		return complex(float32(val), 0), nil
	case int16:
		return complex(float32(val), 0), nil
	case int32:
		return complex(float32(val), 0), nil
	case int64:
		return complex(float32(val), 0), nil
	case uint:
		return complex(float32(val), 0), nil
	case uint8:
		return complex(float32(val), 0), nil
	case uint16:
		return complex(float32(val), 0), nil
	case uint32:
		return complex(float32(val), 0), nil
	case uint64:
		return complex(float32(val), 0), nil
	case float32:
		return complex(val, 0), nil
	case float64:
		return complex(float32(val), 0), nil
	case string:
		result, err := ToComplexFromString(val)
		if err != nil {
			return 0, err
		}
		return complex64(result), nil
	default:
		return 0, NewUnsupportedError(fmt.Sprintf("%T", derefed), "complex64")
	}
}

// ToComplex128 converts any value to complex128
func ToComplex128(v any) (complex128, error) {
	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return 0, fmt.Errorf("cannot convert nil pointer to complex128: %w", ErrNilPointer)
	}

	// Fast path for exact type match
	if result, ok := derefed.(complex128); ok {
		return result, nil
	}

	// Handle other complex types
	if c64, ok := derefed.(complex64); ok {
		return complex128(c64), nil
	}

	// Handle numeric types
	switch val := derefed.(type) {
	case int:
		return complex(float64(val), 0), nil
	case int8:
		return complex(float64(val), 0), nil
	case int16:
		return complex(float64(val), 0), nil
	case int32:
		return complex(float64(val), 0), nil
	case int64:
		return complex(float64(val), 0), nil
	case uint:
		return complex(float64(val), 0), nil
	case uint8:
		return complex(float64(val), 0), nil
	case uint16:
		return complex(float64(val), 0), nil
	case uint32:
		return complex(float64(val), 0), nil
	case uint64:
		return complex(float64(val), 0), nil
	case float32:
		return complex(float64(val), 0), nil
	case float64:
		return complex(val, 0), nil
	case string:
		return ToComplexFromString(val)
	default:
		return 0, NewUnsupportedError(fmt.Sprintf("%T", derefed), "complex128")
	}
}

// ToComplexFromString parses a string as a complex number
func ToComplexFromString(s string) (complex128, error) {
	// Remove spaces for easier parsing
	s = strings.ReplaceAll(s, " ", "")

	// Handle special cases
	if s == "" {
		return 0, NewEmptyInputError("complex")
	}

	// Try to parse as real number first
	if !strings.ContainsAny(s, "ij+") {
		if real, err := strconv.ParseFloat(s, 64); err == nil {
			return complex(real, 0), nil
		}
	}

	// Handle pure imaginary numbers (e.g., "3i", "i", "-i")
	if strings.HasSuffix(s, "i") || strings.HasSuffix(s, "j") {
		imagStr := s[:len(s)-1]
		if imagStr == "" || imagStr == "+" {
			return complex(0, 1), nil
		}
		if imagStr == "-" {
			return complex(0, -1), nil
		}
		if imag, err := strconv.ParseFloat(imagStr, 64); err == nil {
			return complex(0, imag), nil
		}
	}

	// Handle complex numbers in the form "a+bi" or "a-bi"
	var real, imag float64
	var err error

	// Find the position of the last + or - that's not at the beginning
	var splitPos int = -1
	for i := 1; i < len(s); i++ {
		if s[i] == '+' || s[i] == '-' {
			splitPos = i
		}
	}

	if splitPos > 0 {
		realStr := s[:splitPos]
		imagStr := s[splitPos:]

		// Parse real part
		if real, err = strconv.ParseFloat(realStr, 64); err != nil {
			return 0, NewFormatError(realStr, "real part")
		}

		// Parse imaginary part
		if strings.HasSuffix(imagStr, "i") || strings.HasSuffix(imagStr, "j") {
			imagStr = imagStr[:len(imagStr)-1]
			if imagStr == "+" {
				imag = 1
			} else if imagStr == "-" {
				imag = -1
			} else {
				if imag, err = strconv.ParseFloat(imagStr, 64); err != nil {
					return 0, NewFormatError(imagStr, "imaginary part")
				}
			}
		} else {
			return 0, NewFormatError(imagStr, "imaginary part (must end with 'i' or 'j')")
		}

		return complex(real, imag), nil
	}

	return 0, NewFormatError(s, "complex number")
}
