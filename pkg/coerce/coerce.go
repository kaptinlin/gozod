package coerce

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

var (
	ErrUnsupported = errors.New("conversion not supported")
	ErrNilPointer  = errors.New("nil pointer")
	ErrInvalidFmt  = errors.New("invalid format")
	ErrEmptyInput  = errors.New("empty input")
	ErrOverflow    = errors.New("value overflow")
	ErrNegative    = errors.New("negative to unsigned")
	ErrNotWhole    = errors.New("not whole number")
)

func NewUnsupportedError(from, to string) error {
	return fmt.Errorf("cannot convert %s to %s: %w", from, to, ErrUnsupported)
}

func NewFormatError(value, target string) error {
	return fmt.Errorf("cannot parse %q as %s: %w", value, target, ErrInvalidFmt)
}

func NewOverflowError(value any, target string) error {
	return fmt.Errorf("value %v overflows %s: %w", value, target, ErrOverflow)
}

func NewEmptyInputError(target string) error {
	return fmt.Errorf("empty input cannot convert to %s: %w", target, ErrEmptyInput)
}

func NewNegativeError(value any, target string) error {
	return fmt.Errorf("negative value %v cannot convert to %s: %w", value, target, ErrNegative)
}

func NewNotWholeError(value any) error {
	return fmt.Errorf("value %v is not a whole number: %w", value, ErrNotWhole)
}

func NewNilPointerError(target string) error {
	return fmt.Errorf("cannot convert nil pointer to %s: %w", target, ErrNilPointer)
}

// ToBool converts any value to boolean.
func ToBool(v any) (bool, error) {
	d, ok := reflectx.Deref(v)
	if !ok {
		return false, NewNilPointerError("bool")
	}

	switch x := d.(type) {
	case bool:
		return x, nil
	case string:
		return stringToBool(x)
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(x).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(x).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(x).Float() != 0, nil
	default:
		return false, NewUnsupportedError(fmt.Sprintf("%T", d), "bool")
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

// ToString converts any value to its string representation.
func ToString(v any) (string, error) {
	d, ok := reflectx.Deref(v)
	if !ok {
		return "", NewNilPointerError("string")
	}

	switch x := d.(type) {
	case string:
		return x, nil
	case []byte:
		return string(x), nil
	case bool:
		return strconv.FormatBool(x), nil
	case int:
		return strconv.Itoa(x), nil
	case int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(x).Int(), 10), nil
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(x).Uint(), 10), nil
	case float32:
		return strconv.FormatFloat(float64(x), 'g', -1, 32), nil
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64), nil
	case complex64, complex128:
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
		return "", NewUnsupportedError(fmt.Sprintf("%T", d), "string")
	}
}

// timeLayouts defines the supported time string formats for parsing, tried in order.
var timeLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"15:04:05",
	"2006/01/02",
	"01/02/2006",
}

// ToTime converts various inputs to time.Time.
func ToTime(v any) (time.Time, error) {
	d, ok := reflectx.Deref(v)
	if !ok {
		return time.Time{}, NewNilPointerError("time")
	}

	switch x := d.(type) {
	case time.Time:
		return x, nil
	case string:
		for _, layout := range timeLayouts {
			if t, err := time.Parse(layout, x); err == nil {
				return t, nil
			}
		}
		return time.Time{}, NewFormatError(x, "time")
	case int64:
		return time.Unix(x, 0), nil
	case int:
		return time.Unix(int64(x), 0), nil
	case float64:
		return time.Unix(int64(x), 0), nil
	case float32:
		return time.Unix(int64(x), 0), nil
	default:
		return time.Time{}, NewUnsupportedError(fmt.Sprintf("%T", d), "time")
	}
}

// ToInt64 converts any value to int64 with overflow detection.
func ToInt64(v any) (int64, error) {
	d, ok := reflectx.Deref(v)
	if !ok {
		return 0, NewNilPointerError("int64")
	}

	switch x := d.(type) {
	case int64:
		return x, nil
	case int, int8, int16, int32:
		return reflect.ValueOf(x).Int(), nil
	case uint8, uint16, uint32:
		return int64(reflect.ValueOf(x).Uint()), nil //nolint:gosec // uint8/16/32 always fit in int64
	case uint:
		if uint64(x) > math.MaxInt64 {
			return 0, NewOverflowError(x, "int64")
		}
		return int64(x), nil //nolint:gosec // overflow checked above
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
		return 0, NewUnsupportedError(fmt.Sprintf("%T", d), "int64")
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

// ToFloat64 converts any value to float64 with NaN detection.
func ToFloat64(v any) (float64, error) {
	d, ok := reflectx.Deref(v)
	if !ok {
		return 0, NewNilPointerError("float64")
	}

	switch x := d.(type) {
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
		f, _ := x.Float64()
		return f, nil
	case big.Int:
		f, _ := x.Float64()
		return f, nil
	case complex64:
		c := complex128(x)
		return math.Sqrt(real(c)*real(c) + imag(c)*imag(c)), nil
	case complex128:
		return math.Sqrt(real(x)*real(x) + imag(x)*imag(x)), nil
	case string:
		return stringToFloat64(x)
	case bool:
		if x {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, NewUnsupportedError(fmt.Sprintf("%T", d), "float64")
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

// ToBigInt converts any value to *big.Int with hex string support.
func ToBigInt(v any) (*big.Int, error) {
	d, ok := reflectx.Deref(v)
	if !ok {
		return nil, NewNilPointerError("*big.Int")
	}

	switch x := d.(type) {
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
		return nil, NewUnsupportedError(fmt.Sprintf("%T", d), "*big.Int")
	}
}

func stringToBigInt(s string) (*big.Int, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return big.NewInt(0), nil
	}

	n := new(big.Int)
	if _, ok := n.SetString(trimmed, 10); ok {
		return n, nil
	}
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		if _, ok := n.SetString(trimmed[2:], 16); ok {
			return n, nil
		}
	}
	return nil, NewFormatError(s, "big integer")
}

// ToInteger converts any value to the specified integer type T with bounds checking.
func ToInteger[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](v any) (T, error) {
	var zero T

	d, ok := reflectx.Deref(v)
	if !ok {
		return zero, NewNilPointerError("integer type")
	}

	var val int64
	var err error

	switch x := d.(type) {
	case int, int8, int16, int32, int64:
		val = reflect.ValueOf(x).Int()
	case uint8, uint16, uint32:
		val = int64(reflect.ValueOf(x).Uint()) //nolint:gosec // uint8/16/32 always fit in int64
	case uint:
		if uint64(x) > math.MaxInt64 {
			return zero, NewOverflowError(x, fmt.Sprintf("%T", zero))
		}
		val = int64(x) //nolint:gosec // overflow checked above
	case uint64:
		if x > math.MaxInt64 {
			return zero, NewOverflowError(x, fmt.Sprintf("%T", zero))
		}
		val = int64(x)
	case float32:
		if float64(x) > math.MaxInt64 || float64(x) < math.MinInt64 {
			return zero, NewOverflowError(x, "int64")
		}
		val = int64(x)
	case float64:
		if math.Trunc(x) != x {
			return zero, NewNotWholeError(x)
		}
		if x > math.MaxInt64 || x < math.MinInt64 {
			return zero, NewOverflowError(x, "int64")
		}
		val = int64(x)
	case string:
		val, err = stringToInt64(x)
		if err != nil {
			return zero, err
		}
	case bool:
		if x {
			return 1, nil
		}
		return 0, nil
	default:
		return zero, NewUnsupportedError(fmt.Sprintf("%T", d), fmt.Sprintf("%T", zero))
	}

	if err := checkIntegerTypeBounds(val, zero); err != nil {
		return zero, err
	}

	return T(val), nil
}

// ToFloat converts any value to the specified float type T with overflow checking.
func ToFloat[T ~float32 | ~float64](v any) (T, error) {
	var zero T

	d, ok := reflectx.Deref(v)
	if !ok {
		return zero, NewNilPointerError(fmt.Sprintf("%T", zero))
	}

	if result, ok := d.(T); ok {
		return result, nil
	}

	fval, err := ToFloat64(d)
	if err != nil {
		return zero, err
	}

	if _, ok := any(zero).(float32); ok {
		if math.Abs(fval) > math.MaxFloat32 {
			return zero, NewOverflowError(fval, "float32")
		}
	}
	return T(fval), nil
}

func checkIntegerTypeBounds[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](v int64, target T) error {
	switch any(target).(type) {
	case int8:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return NewOverflowError(v, "int8")
		}
	case int16:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return NewOverflowError(v, "int16")
		}
	case int32:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return NewOverflowError(v, "int32")
		}
	case int:
		maxInt := int64(^uint(0) >> 1)
		minInt := -maxInt - 1
		if v < minInt || v > maxInt {
			return NewOverflowError(v, fmt.Sprintf("int (%d-bit)", strconv.IntSize))
		}
	case uint8:
		if v < 0 {
			return NewNegativeError(v, "uint8")
		}
		if v > math.MaxUint8 {
			return NewOverflowError(v, "uint8")
		}
	case uint16:
		if v < 0 {
			return NewNegativeError(v, "uint16")
		}
		if v > math.MaxUint16 {
			return NewOverflowError(v, "uint16")
		}
	case uint32:
		if v < 0 {
			return NewNegativeError(v, "uint32")
		}
		if v > math.MaxUint32 {
			return NewOverflowError(v, "uint32")
		}
	case uint:
		if v < 0 {
			return NewNegativeError(v, "uint")
		}
		maxUint := uint64(^uint(0))
		if uint64(v) > maxUint {
			return NewOverflowError(v, fmt.Sprintf("uint (%d-bit)", strconv.IntSize))
		}
	case uint64:
		if v < 0 {
			return NewNegativeError(v, "uint64")
		}
	}
	return nil
}

// ToComplex64 converts any value to complex64.
func ToComplex64(v any) (complex64, error) {
	c, err := toComplex128(v, "complex64")
	if err != nil {
		return 0, err
	}
	return complex64(c), nil
}

// ToComplex128 converts any value to complex128.
func ToComplex128(v any) (complex128, error) {
	return toComplex128(v, "complex128")
}

// toComplex128 is the shared implementation for complex conversions.
func toComplex128(v any, target string) (complex128, error) {
	d, ok := reflectx.Deref(v)
	if !ok {
		return 0, NewNilPointerError(target)
	}

	switch x := d.(type) {
	case complex128:
		return x, nil
	case complex64:
		return complex128(x), nil
	case int, int8, int16, int32, int64:
		return complex(float64(reflect.ValueOf(x).Int()), 0), nil
	case uint, uint8, uint16, uint32, uint64:
		return complex(float64(reflect.ValueOf(x).Uint()), 0), nil
	case float32:
		return complex(float64(x), 0), nil
	case float64:
		return complex(x, 0), nil
	case string:
		return ToComplexFromString(x)
	default:
		return 0, NewUnsupportedError(fmt.Sprintf("%T", d), target)
	}
}

// ToComplexFromString parses a complex number from its string representation.
func ToComplexFromString(s string) (complex128, error) {
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, NewEmptyInputError("complex")
	}

	// Try pure real number (no imaginary markers).
	if !strings.ContainsAny(s, "ij+") {
		if re, err := strconv.ParseFloat(s, 64); err == nil {
			return complex(re, 0), nil
		}
	}

	// Try pure imaginary number.
	if strings.HasSuffix(s, "i") || strings.HasSuffix(s, "j") {
		raw := s[:len(s)-1]
		switch raw {
		case "", "+":
			return complex(0, 1), nil
		case "-":
			return complex(0, -1), nil
		default:
			if im, err := strconv.ParseFloat(raw, 64); err == nil {
				return complex(0, im), nil
			}
		}
	}

	// Try complex number with real and imaginary parts (e.g. "3+4i").
	if pos := findComplexSplit(s); pos > 0 {
		return parseComplexParts(s, pos)
	}
	return 0, NewFormatError(s, "complex number")
}

// findComplexSplit returns the index of the +/- separator between real and
// imaginary parts, skipping scientific notation exponents. Returns -1 if not found.
func findComplexSplit(s string) int {
	pos := -1
	for i := 1; i < len(s); i++ {
		if (s[i] == '+' || s[i] == '-') && s[i-1] != 'e' && s[i-1] != 'E' {
			pos = i
		}
	}
	return pos
}

// parseComplexParts parses a complex string split at pos into real and imaginary components.
func parseComplexParts(s string, pos int) (complex128, error) {
	reStr := s[:pos]
	imStr := s[pos:]

	re, err := strconv.ParseFloat(reStr, 64)
	if err != nil {
		return 0, NewFormatError(reStr, "real part")
	}

	if !strings.HasSuffix(imStr, "i") && !strings.HasSuffix(imStr, "j") {
		return 0, NewFormatError(imStr, "imaginary part (must end with 'i' or 'j')")
	}

	imStr = imStr[:len(imStr)-1]
	var im float64
	switch imStr {
	case "+":
		im = 1
	case "-":
		im = -1
	default:
		im, err = strconv.ParseFloat(imStr, 64)
		if err != nil {
			return 0, NewFormatError(imStr, "imaginary part")
		}
	}
	return complex(re, im), nil
}

// To converts any value to the specified type T using the appropriate converter.
func To[T any](v any) (T, error) {
	var zero T
	target := reflect.TypeOf(zero)
	if target == nil {
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
		return reflectx.Convert[T](v)
	}
}

// ToLiteral converts a value to its most natural Go literal representation.
// The error return is reserved for future use and is currently always nil.
func ToLiteral(v any) (any, error) {
	if v == nil {
		return nil, nil
	}

	if s, ok := v.(string); ok {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			return v, nil
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
		return s, nil
	}

	switch n := v.(type) {
	case float32:
		if math.Mod(float64(n), 1) == 0 {
			return intLiteralFromInt64(int64(n)), nil
		}
	case float64:
		if math.Mod(n, 1) == 0 {
			return intLiteralFromInt64(int64(n)), nil
		}
	case int, int8, int16, int32, int64:
		return intLiteralFromInt64(reflect.ValueOf(n).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return uintLiteralFromUint64(reflect.ValueOf(n).Uint()), nil
	}
	return v, nil
}

// intLiteralFromInt64 converts an int64 to its most natural literal:
// 0 -> false, 1 -> true, otherwise -> int.
func intLiteralFromInt64(v int64) any {
	switch v {
	case 0:
		return false
	case 1:
		return true
	default:
		return v
	}
}

// uintLiteralFromUint64 converts a uint64 to its most natural literal:
// 0 -> false, 1 -> true, otherwise the original value is returned unchanged.
func uintLiteralFromUint64(v uint64) any {
	switch v {
	case 0:
		return false
	case 1:
		return true
	default:
		return v
	}
}
