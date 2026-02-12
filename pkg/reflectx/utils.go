package reflectx

import (
	"errors"
	"fmt"
	"reflect"
)

// Sentinel errors returned by [Convert].
var (
	// ErrNil indicates a nil value was passed to [Convert].
	ErrNil = errors.New("cannot convert nil")
	// ErrUnsupported indicates the source type cannot be converted
	// to the target type via assignment or [reflect.Value.Convert].
	ErrUnsupported = errors.New("unsupported conversion")
)

// Convert converts v to type T. It tries direct assignment first,
// then falls back to [reflect.Value.Convert] for safe numeric/string
// casts. Returns the zero value of T and a wrapped sentinel error
// ([ErrNil] or [ErrUnsupported]) when conversion is impossible.
func Convert[T any](v any) (T, error) {
	var zero T
	target := reflect.TypeOf(zero)

	if v == nil {
		return zero, fmt.Errorf(
			"cannot convert nil to %v: %w", target, ErrNil)
	}

	src := reflect.ValueOf(v)

	// Fast path: direct assignment (no allocation).
	if src.Type().AssignableTo(target) {
		return v.(T), nil
	}

	// Slow path: reflect.Convert for safe numeric/string casts.
	if src.Type().ConvertibleTo(target) {
		return src.Convert(target).Interface().(T), nil
	}

	return zero, fmt.Errorf(
		"unsupported conversion from %T to %v: %w",
		v, target, ErrUnsupported)
}
