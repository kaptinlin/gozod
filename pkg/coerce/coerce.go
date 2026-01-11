package coerce

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// COERCION ERRORS
// =============================================================================

var (
	// Type compatibility errors
	ErrUnsupported = errors.New("conversion not supported")
	ErrNilPointer  = errors.New("nil pointer")

	// Format and parsing errors
	ErrInvalidFormat = errors.New("invalid format")
	ErrEmptyInput    = errors.New("empty input")

	// Numeric conversion errors
	ErrOverflow     = errors.New("value overflow")
	ErrNegativeUint = errors.New("negative to unsigned")
	ErrNotWhole     = errors.New("not whole number")
)

// NewUnsupportedError creates a detailed unsupported conversion error
func NewUnsupportedError(from, to string) error {
	return fmt.Errorf("cannot convert %s to %s: %w", from, to, ErrUnsupported)
}

// NewFormatError creates a detailed format error with the problematic value
func NewFormatError(value, targetType string) error {
	return fmt.Errorf("cannot parse %q as %s: %w", value, targetType, ErrInvalidFormat)
}

// NewOverflowError creates a detailed overflow error with the problematic value
func NewOverflowError(value any, targetType string) error {
	return fmt.Errorf("value %v overflows %s: %w", value, targetType, ErrOverflow)
}

// NewEmptyInputError creates a detailed empty input error for specific type
func NewEmptyInputError(targetType string) error {
	return fmt.Errorf("empty input cannot convert to %s: %w", targetType, ErrEmptyInput)
}

// NewNegativeUintError creates a detailed negative to unsigned conversion error
func NewNegativeUintError(value any, targetType string) error {
	return fmt.Errorf("negative value %v cannot convert to %s: %w", value, targetType, ErrNegativeUint)
}

// NewNotWholeError creates a detailed non-whole number error
func NewNotWholeError(value any) error {
	return fmt.Errorf("value %v is not a whole number: %w", value, ErrNotWhole)
}

// =============================================================================
// BOOLEAN CONVERSION
// =============================================================================

// ToBool converts any value to boolean with fast-path optimizations
func ToBool(v any) (bool, error) {
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return false, fmt.Errorf("cannot convert nil pointer to bool: %w", ErrNilPointer)
	}

	switch x := derefed.(type) {
	case bool:
		return x, nil
	case string:
		return stringToBool(x)
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(x).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(x).Uint() != 0, nil
	case float32:
		return x != 0, nil
	case float64:
		return x != 0, nil
	default:
		return false, NewUnsupportedError(fmt.Sprintf("%T", derefed), "bool")
	}
}

func stringToBool(s string) (bool, error) {
	s = strings.TrimSpace(s)
	switch strings.ToLower(s) {
	case "true", "1", "yes", "on", "y":
		return true, nil
	case "false", "0", "no", "off", "n", "":
		return false, nil
	default:
		return false, NewFormatError(s, "bool")
	}
}

// =============================================================================
// STRING & TIME CONVERSION
// =============================================================================

// ToString converts any value to string with fast-path optimizations
func ToString(v any) (string, error) {
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return "", fmt.Errorf("cannot convert nil pointer to string: %w", ErrNilPointer)
	}

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
	case *big.Int:
		if x == nil {
			return "0", nil
		}
		return x.String(), nil
	case big.Int:
		return x.String(), nil
	case time.Time:
		return x.Format(time.RFC3339), nil
	default:
		return "", NewUnsupportedError(fmt.Sprintf("%T", derefed), "string")
	}
}

// ToTime converts various inputs to time.Time
func ToTime(value any) (time.Time, error) {
	derefed, ok := reflectx.Deref(value)
	if !ok {
		return time.Time{}, fmt.Errorf("cannot convert nil pointer to time: %w", ErrNilPointer)
	}

	switch v := derefed.(type) {
	case time.Time:
		return v, nil
	case string:
		layouts := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"15:04:05",
			"2006/01/02",
			"01/02/2006",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, NewFormatError(v, "time")
	case int64:
		return time.Unix(v, 0), nil
	case int:
		return time.Unix(int64(v), 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	case float32:
		return time.Unix(int64(v), 0), nil
	default:
		return time.Time{}, NewUnsupportedError(fmt.Sprintf("%T", derefed), "time")
	}
}

// =============================================================================
// NUMERIC CONVERSIONS (Int, Float, BigInt, Complex)
// =============================================================================

// --- Int64 helpers ----------------------------------------------------------

func ToInt64(v any) (int64, error) {
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return 0, fmt.Errorf("cannot convert nil pointer to int64: %w", ErrNilPointer)
	}

	switch x := derefed.(type) {
	case int64:
		return x, nil
	case int, int8, int16, int32:
		return reflect.ValueOf(x).Int(), nil
	case uint, uint8, uint16, uint32:
		val := reflect.ValueOf(x).Uint()
		if val > math.MaxInt64 {
			return 0, NewOverflowError(x, "int64")
		}
		return int64(val), nil
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

func stringToInt64(s string) (int64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, nil
	}
	i, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, NewFormatError(s, "int64")
	}
	return i, nil
}

// --- Float helpers ----------------------------------------------------------

func ToFloat64(v any) (float64, error) {
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return 0, fmt.Errorf("cannot convert nil pointer to float64: %w", ErrNilPointer)
	}

	switch x := derefed.(type) {
	case float64:
		if math.IsNaN(x) {
			return 0, NewFormatError("NaN", "float64")
		}
		return x, nil
	case float32:
		if math.IsNaN(float64(x)) {
			return 0, NewFormatError("NaN", "float64")
		}
		return float64(x), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(x).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(x).Uint()), nil
	case *big.Int:
		if x == nil {
			return 0, fmt.Errorf("cannot convert nil *big.Int to float64: %w", ErrNilPointer)
		}
		f, _ := x.Float64()
		return f, nil
	case big.Int:
		f, _ := x.Float64()
		return f, nil
	case complex64:
		return cmplx.Abs(complex128(x)), nil
	case complex128:
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

func stringToFloat64(s string) (float64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, NewFormatError(s, "float64")
	}
	return f, nil
}

// --- BigInt helpers ---------------------------------------------------------

func ToBigInt(v any) (*big.Int, error) {
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return nil, fmt.Errorf("cannot convert nil pointer to *big.Int: %w", ErrNilPointer)
	}

	switch x := derefed.(type) {
	case *big.Int:
		if x == nil {
			return nil, fmt.Errorf("cannot convert nil *big.Int: %w", ErrNilPointer)
		}
		return new(big.Int).Set(x), nil
	case int, int8, int16, int32, int64:
		return big.NewInt(reflect.ValueOf(x).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return new(big.Int).SetUint64(reflect.ValueOf(x).Uint()), nil
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

func stringToBigInt(s string) (*big.Int, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return big.NewInt(0), nil
	}

	bigInt := new(big.Int)
	if _, ok := bigInt.SetString(trimmed, 10); ok {
		return bigInt, nil
	}
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		if _, ok := bigInt.SetString(trimmed[2:], 16); ok {
			return bigInt, nil
		}
	}
	return nil, NewFormatError(s, "big integer")
}

// --- Generic integer/float helpers -----------------------------------------

func ToInteger[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](v any) (T, error) {
	var zero T

	derefed, ok := reflectx.Deref(v)
	if !ok {
		return zero, fmt.Errorf("cannot convert nil pointer to integer type: %w", ErrNilPointer)
	}

	var int64Val int64
	var err error

	switch x := derefed.(type) {
	case int64:
		int64Val = x
	case int, int8, int16, int32:
		int64Val = reflect.ValueOf(x).Int()
	case uint, uint8, uint16, uint32:
		uintVal := reflect.ValueOf(x).Uint()
		if uintVal > math.MaxInt64 {
			var zero T
			return zero, NewOverflowError(x, fmt.Sprintf("%T", zero))
		}
		int64Val = int64(uintVal)
	case uint64:
		if x > math.MaxInt64 {
			var zero T
			return zero, NewOverflowError(x, fmt.Sprintf("%T", zero))
		}
		int64Val = int64(x)
	case float32:
		if float64(x) > math.MaxInt64 || float64(x) < math.MinInt64 {
			var zero T
			return zero, NewOverflowError(x, "int64")
		}
		int64Val = int64(x)
	case float64:
		if math.Trunc(x) != x {
			var zero T
			return zero, NewNotWholeError(x)
		}
		if x > math.MaxInt64 || x < math.MinInt64 {
			var zero T
			return zero, NewOverflowError(x, "int64")
		}
		int64Val = int64(x)
	case string:
		int64Val, err = stringToInt64(x)
		if err != nil {
			return zero, err
		}
	case bool:
		if x {
			return 1, nil
		}
		return 0, nil
	default:
		return zero, NewUnsupportedError(fmt.Sprintf("%T", derefed), fmt.Sprintf("%T", zero))
	}

	if err := checkIntegerTypeBounds(int64Val, zero); err != nil {
		return zero, err
	}

	return T(int64Val), nil
}

func ToFloat[T ~float32 | ~float64](v any) (T, error) {
	var zero T

	derefed, ok := reflectx.Deref(v)
	if !ok {
		return zero, fmt.Errorf("cannot convert nil pointer to %T: %w", zero, ErrNilPointer)
	}

	if result, ok := derefed.(T); ok {
		return result, nil
	}

	float64Val, err := ToFloat64(derefed)
	if err != nil {
		return zero, err
	}

	if _, ok := any(zero).(float32); ok {
		if math.Abs(float64Val) > math.MaxFloat32 {
			return zero, NewOverflowError(float64Val, "float32")
		}
	}
	return T(float64Val), nil
}

func checkIntegerTypeBounds[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](value int64, target T) error {
	switch any(target).(type) {
	case int8:
		if value < math.MinInt8 || value > math.MaxInt8 {
			return NewOverflowError(value, "int8")
		}
	case int16:
		if value < math.MinInt16 || value > math.MaxInt16 {
			return NewOverflowError(value, "int16")
		}
	case int32:
		if value < math.MinInt32 || value > math.MaxInt32 {
			return NewOverflowError(value, "int32")
		}
	case int:
		if value < math.MinInt32 || value > math.MaxInt32 {
			return NewOverflowError(value, "int (32-bit platforms)")
		}
	case uint8:
		if value < 0 || value > math.MaxUint8 {
			if value < 0 {
				return NewNegativeUintError(value, "uint8")
			}
			return NewOverflowError(value, "uint8")
		}
	case uint16:
		if value < 0 || value > math.MaxUint16 {
			if value < 0 {
				return NewNegativeUintError(value, "uint16")
			}
			return NewOverflowError(value, "uint16")
		}
	case uint32:
		if value < 0 || value > math.MaxUint32 {
			if value < 0 {
				return NewNegativeUintError(value, "uint32")
			}
			return NewOverflowError(value, "uint32")
		}
	case uint:
		if value < 0 || value > math.MaxUint32 {
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

// --- Complex helpers --------------------------------------------------------

func ToComplex64(v any) (complex64, error) {
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return 0, fmt.Errorf("cannot convert nil pointer to complex64: %w", ErrNilPointer)
	}

	if result, ok := derefed.(complex64); ok {
		return result, nil
	}
	if c128, ok := derefed.(complex128); ok {
		return complex64(c128), nil
	}

	switch val := derefed.(type) {
	case int, int8, int16, int32, int64:
		return complex(float32(reflect.ValueOf(val).Int()), 0), nil
	case uint, uint8, uint16, uint32, uint64:
		return complex(float32(reflect.ValueOf(val).Uint()), 0), nil
	case float32:
		return complex(val, 0), nil
	case float64:
		return complex(float32(val), 0), nil
	case string:
		res, err := ToComplexFromString(val)
		if err != nil {
			return 0, err
		}
		return complex64(res), nil
	default:
		return 0, NewUnsupportedError(fmt.Sprintf("%T", derefed), "complex64")
	}
}

func ToComplex128(v any) (complex128, error) {
	derefed, ok := reflectx.Deref(v)
	if !ok {
		return 0, fmt.Errorf("cannot convert nil pointer to complex128: %w", ErrNilPointer)
	}

	if result, ok := derefed.(complex128); ok {
		return result, nil
	}
	if c64, ok := derefed.(complex64); ok {
		return complex128(c64), nil
	}

	switch val := derefed.(type) {
	case int, int8, int16, int32, int64:
		return complex(float64(reflect.ValueOf(val).Int()), 0), nil
	case uint, uint8, uint16, uint32, uint64:
		return complex(float64(reflect.ValueOf(val).Uint()), 0), nil
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

func ToComplexFromString(s string) (complex128, error) {
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, NewEmptyInputError("complex")
	}

	if !strings.ContainsAny(s, "ij+") {
		if real, err := strconv.ParseFloat(s, 64); err == nil {
			return complex(real, 0), nil
		}
	}

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

	var real, imag float64
	var err error
	splitPos := -1
	for i := 1; i < len(s); i++ {
		if s[i] == '+' || s[i] == '-' {
			splitPos = i
		}
	}
	if splitPos > 0 {
		realStr := s[:splitPos]
		imagStr := s[splitPos:]
		if real, err = strconv.ParseFloat(realStr, 64); err != nil {
			return 0, NewFormatError(realStr, "real part")
		}
		if strings.HasSuffix(imagStr, "i") || strings.HasSuffix(imagStr, "j") {
			imagStr = imagStr[:len(imagStr)-1]
			switch imagStr {
			case "+":
				imag = 1
			case "-":
				imag = -1
			default:
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

// =============================================================================
// GENERIC ENTRY & UTILITIES
// =============================================================================

func To[T any](v any) (T, error) {
	var zero T
	targetType := reflect.TypeOf(zero)
	if targetType == nil {
		return zero, NewUnsupportedError("unknown", "unknown")
	}

	switch any(zero).(type) {
	case string:
		r, err := ToString(v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case bool:
		r, err := ToBool(v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case time.Time:
		r, err := ToTime(v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case int:
		r, err := ToInteger[int](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case int8:
		r, err := ToInteger[int8](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case int16:
		r, err := ToInteger[int16](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case int32:
		r, err := ToInteger[int32](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case int64:
		r, err := ToInt64(v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case uint:
		r, err := ToInteger[uint](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case uint8:
		r, err := ToInteger[uint8](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case uint16:
		r, err := ToInteger[uint16](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case uint32:
		r, err := ToInteger[uint32](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case uint64:
		r, err := ToInteger[uint64](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case float32:
		r, err := ToFloat[float32](v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case float64:
		r, err := ToFloat64(v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case *big.Int:
		r, err := ToBigInt(v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case complex64:
		r, err := ToComplex64(v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	case complex128:
		r, err := ToComplex128(v)
		if err != nil {
			return zero, err
		}
		return any(r).(T), nil
	default:
		return reflectx.ConvertToGeneric[T](v)
	}
}

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
		if b, err := strconv.ParseBool(trimmed); err == nil {
			return b, nil
		}
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			if i == 1 {
				return true, nil
			}
			if i == 0 {
				return false, nil
			}
			maxInt := int64(^uint(0) >> 1)
			minInt := -maxInt - 1
			if i >= minInt && i <= maxInt {
				return int(i), nil
			}
			return i, nil
		}
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			if math.Mod(f, 1) == 0 {
				return int(f), nil
			}
			return f, nil
		}
		return v, nil
	default:
		switch n := value.(type) {
		case float32:
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
			if math.Mod(n, 1) == 0 {
				if n == 1.0 {
					return true, nil
				}
				if n == 0.0 {
					return false, nil
				}
				return int(n), nil
			}
		case int:
			if n == 1 {
				return true, nil
			}
			if n == 0 {
				return false, nil
			}
		case int8, int16, int32, int64:
			if reflect.ValueOf(n).Int() == 1 {
				return true, nil
			}
			if reflect.ValueOf(n).Int() == 0 {
				return false, nil
			}
		case uint, uint8, uint16, uint32, uint64:
			if reflect.ValueOf(n).Uint() == 1 {
				return true, nil
			}
			if reflect.ValueOf(n).Uint() == 0 {
				return false, nil
			}
		}
		return value, nil
	}
}
