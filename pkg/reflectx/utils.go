package reflectx

import (
	"errors"
	"fmt"
	"reflect"
)

// Sentinel errors returned by ConvertToGeneric.
var (
	ErrNilValue              = errors.New("cannot convert nil")
	ErrUnsupportedConversion = errors.New("unsupported conversion")
)

// ConvertToGeneric converts arbitrary value v into type T using reflection only
// when necessary. It returns zero value of T plus an error when conversion is
// impossible. Fast paths are kept extremely short for performance.
func ConvertToGeneric[T any](v any) (T, error) {
	var zero T
	target := reflect.TypeOf(zero)

	if v == nil {
		return zero, fmt.Errorf("cannot convert nil to %v: %w", target, ErrNilValue)
	}

	srcVal := reflect.ValueOf(v)
	srcType := srcVal.Type()

	// 1️⃣  direct assignment (no allocation)
	if srcType.AssignableTo(target) {
		return v.(T), nil
	}

	// 2️⃣  reflect.Convert handles safe numeric / string casts
	if srcType.ConvertibleTo(target) {
		return srcVal.Convert(target).Interface().(T), nil
	}

	// 3️⃣  unsupported – bail out quickly
	return zero, fmt.Errorf("unsupported conversion from %T to %v: %w", v, target, ErrUnsupportedConversion)
}
