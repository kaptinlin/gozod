package reflectx

import (
	"errors"
	"fmt"
	"reflect"
)

// Sentinel errors returned by [ConvertToGeneric].
var (
	// ErrNilValue indicates a nil value was passed to ConvertToGeneric.
	ErrNilValue = errors.New("cannot convert nil")
	// ErrUnsupportedConversion indicates the source type cannot be converted
	// to the target type via assignment or reflect.Convert.
	ErrUnsupportedConversion = errors.New("unsupported conversion")
)

// ConvertToGeneric converts v to type T. It tries direct assignment first,
// then falls back to [reflect.Value.Convert] for safe numeric/string casts.
// Returns the zero value of T and a wrapped sentinel error ([ErrNilValue] or
// [ErrUnsupportedConversion]) when conversion is impossible.
func ConvertToGeneric[T any](v any) (T, error) {
	var zero T
	target := reflect.TypeOf(zero)

	if v == nil {
		return zero, fmt.Errorf("cannot convert nil to %v: %w", target, ErrNilValue)
	}

	srcVal := reflect.ValueOf(v)
	srcType := srcVal.Type()

	// Fast path: direct assignment (no allocation).
	if srcType.AssignableTo(target) {
		return v.(T), nil
	}

	// Slow path: reflect.Convert handles safe numeric/string casts.
	if srcType.ConvertibleTo(target) {
		return srcVal.Convert(target).Interface().(T), nil
	}

	return zero, fmt.Errorf("unsupported conversion from %T to %v: %w", v, target, ErrUnsupportedConversion)
}
