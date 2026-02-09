package slicex

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Sentinel errors for the slicex package.
var (
	ErrInvalidReflectValue   = errors.New("invalid reflect value")
	ErrNotSliceArrayOrString = errors.New("input is not a slice, array, or string")

	ErrCannotConvertSliceElement = errors.New("cannot convert slice element")
	ErrCannotConvertFirstSlice   = errors.New("cannot convert first slice")
	ErrCannotConvertSecondSlice  = errors.New("cannot convert second slice")
	ErrCannotConvertSlice        = errors.New("cannot convert slice")
)

// toAnySlice converts a typed slice to []any, eliminating repeated
// make+loop boilerplate across type-switch branches.
func toAnySlice[T any](v []T) []any {
	result := make([]any, len(v))
	for i, item := range v {
		result[i] = item
	}
	return result
}

// ToAny converts a slice, array, or string to []any.
// A nil input returns (nil, nil). A bare string returns a
// single-element []any. For unrecognized concrete types,
// reflection is used as a fallback.
func ToAny(input any) ([]any, error) {
	if input == nil {
		return nil, nil
	}
	switch v := input.(type) {
	case []any:
		return v, nil
	case []string:
		return toAnySlice(v), nil
	case []int:
		return toAnySlice(v), nil
	case []int32:
		return toAnySlice(v), nil
	case []int64:
		return toAnySlice(v), nil
	case []float32:
		return toAnySlice(v), nil
	case []float64:
		return toAnySlice(v), nil
	case []bool:
		return toAnySlice(v), nil
	case string:
		return []any{v}, nil
	default:
		return FromReflect(reflect.ValueOf(input))
	}
}

// ToTyped converts any slice to []T using generics.
// Each element is first tried via type assertion, then via
// reflect conversion.
//
// Returns ErrCannotConvertSliceElement if an element cannot
// be converted to the target type.
func ToTyped[T any](input any) ([]T, error) {
	if input == nil {
		return nil, nil
	}
	anySlice, err := ToAny(input)
	if err != nil {
		return nil, err
	}
	targetType := reflect.TypeFor[T]()
	result := make([]T, len(anySlice))
	for i, item := range anySlice {
		if typed, ok := item.(T); ok {
			result[i] = typed
			continue
		}
		itemValue := reflect.ValueOf(item)
		if !itemValue.Type().ConvertibleTo(targetType) {
			return nil, fmt.Errorf(
				"element %v (type %T) not convertible to %v: %w",
				item, item, targetType,
				ErrCannotConvertSliceElement,
			)
		}
		result[i] = itemValue.Convert(targetType).Interface().(T)
	}
	return result, nil
}

// ToStrings converts any slice to []string by formatting each
// element with fmt.Sprintf.
func ToStrings(input any) ([]string, error) {
	anySlice, err := ToAny(input)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(anySlice))
	for i, item := range anySlice {
		result[i] = fmt.Sprintf("%v", item)
	}
	return result, nil
}

// FromReflect converts a reflect.Value (slice, array, or string)
// to []any.
//
// Returns ErrInvalidReflectValue for invalid values, and
// ErrNotSliceArrayOrString for unsupported kinds.
func FromReflect(rv reflect.Value) ([]any, error) {
	if !rv.IsValid() {
		return nil, ErrInvalidReflectValue
	}
	//nolint:exhaustive // Only Slice, Array, String are valid
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]any, rv.Len())
		for i := range rv.Len() {
			result[i] = rv.Index(i).Interface()
		}
		return result, nil
	case reflect.String:
		return StringToChars(rv.String()), nil
	default:
		return nil, fmt.Errorf(
			"got %v: %w", rv.Kind(), ErrNotSliceArrayOrString,
		)
	}
}

// StringToChars converts a string to []any where each element
// is a single-character string.
func StringToChars(s string) []any {
	runes := []rune(s)
	result := make([]any, len(runes))
	for i, r := range runes {
		result[i] = string(r)
	}
	return result
}

// Extract converts input to []any, returning whether the
// conversion succeeded.
func Extract(input any) ([]any, bool) {
	result, err := ToAny(input)
	return result, err == nil
}

// ExtractArray extracts an array (not slice) from input.
// Returns false for nil, slices, or non-array types.
func ExtractArray(input any) ([]any, bool) {
	if input == nil {
		return nil, false
	}
	if reflect.ValueOf(input).Kind() != reflect.Array {
		return nil, false
	}
	result, err := ToAny(input)
	return result, err == nil
}

// ExtractSlice extracts a slice (not array) from input.
// Returns false for nil, arrays, or non-slice types.
func ExtractSlice(input any) ([]any, bool) {
	if input == nil {
		return nil, false
	}
	if reflect.ValueOf(input).Kind() != reflect.Slice {
		return nil, false
	}
	result, err := ToAny(input)
	return result, err == nil
}

// Merge concatenates two slices. If both inputs share the same
// concrete slice type, the result preserves that type.
//
// Returns ErrCannotConvertFirstSlice or ErrCannotConvertSecondSlice
// if the respective input cannot be converted.
func Merge(a, b any) (any, error) {
	if a == nil && b == nil {
		return nil, nil
	}
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}
	sliceA, errA := ToAny(a)
	if errA != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertFirstSlice, errA)
	}
	sliceB, errB := ToAny(b)
	if errB != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSecondSlice, errB)
	}
	merged := make([]any, 0, len(sliceA)+len(sliceB))
	merged = append(merged, sliceA...)
	merged = append(merged, sliceB...)
	return restoreType(merged, a, b)
}

// Append appends elements to a slice, preserving the original
// slice's concrete type when possible.
func Append(slice any, elements ...any) (any, error) {
	if slice == nil {
		return elements, nil
	}
	sliceAny, err := ToAny(slice)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSlice, err)
	}
	result := make([]any, len(sliceAny)+len(elements))
	copy(result, sliceAny)
	copy(result[len(sliceAny):], elements)
	return restoreType(result, slice, nil)
}

// Prepend inserts elements before a slice, preserving the
// original slice's concrete type when possible.
func Prepend(slice any, elements ...any) (any, error) {
	if slice == nil {
		return elements, nil
	}
	sliceAny, err := ToAny(slice)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSlice, err)
	}
	result := make([]any, len(elements)+len(sliceAny))
	copy(result, elements)
	copy(result[len(elements):], sliceAny)
	return restoreType(result, slice, nil)
}

// Length returns the length of a slice, array, or string.
// Returns ErrNotSliceArrayOrString for other types.
func Length(input any) (int, error) {
	if input == nil {
		return 0, nil
	}
	rv := reflect.ValueOf(input)
	//nolint:exhaustive // Only handling slice, array, and string
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		return rv.Len(), nil
	default:
		return 0, ErrNotSliceArrayOrString
	}
}

// IsEmpty reports whether input is nil, empty, or not a
// recognized collection type.
func IsEmpty(input any) bool {
	length, err := Length(input)
	return err != nil || length == 0
}

// Contains reports whether slice contains value, compared
// using reflect.DeepEqual.
func Contains(slice any, value any) bool {
	sliceAny, err := ToAny(slice)
	if err != nil {
		return false
	}
	return slices.ContainsFunc(sliceAny, func(item any) bool {
		return reflect.DeepEqual(item, value)
	})
}

// IndexOf returns the index of the first occurrence of value
// in slice, or -1 if not found. Comparison uses
// reflect.DeepEqual.
func IndexOf(slice any, value any) int {
	sliceAny, err := ToAny(slice)
	if err != nil {
		return -1
	}
	return slices.IndexFunc(sliceAny, func(item any) bool {
		return reflect.DeepEqual(item, value)
	})
}

// Reverse returns a new slice with elements in reverse order.
func Reverse(slice any) (any, error) {
	if slice == nil {
		return nil, nil
	}
	sliceAny, err := ToAny(slice)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSlice, err)
	}
	slices.Reverse(sliceAny)
	return restoreType(sliceAny, slice, nil)
}

// Unique removes duplicate elements, preserving first-occurrence
// order. Comparable types use a map; non-comparable types fall
// back to reflect.DeepEqual.
func Unique(slice any) (any, error) {
	if slice == nil {
		return nil, nil
	}
	sliceAny, err := ToAny(slice)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSlice, err)
	}
	seenComparable := make(map[any]struct{})
	var seenNonComparable []any
	result := make([]any, 0, len(sliceAny))
	for _, item := range sliceAny {
		if isDuplicate(item, seenComparable, seenNonComparable) {
			continue
		}
		if item == nil || reflect.TypeOf(item).Comparable() {
			seenComparable[item] = struct{}{}
		} else {
			seenNonComparable = append(seenNonComparable, item)
		}
		result = append(result, item)
	}
	return restoreType(result, slice, nil)
}

// isDuplicate checks whether item has already been seen.
func isDuplicate(
	item any,
	seenComparable map[any]struct{},
	seenNonComparable []any,
) bool {
	if item == nil || reflect.TypeOf(item).Comparable() {
		_, exists := seenComparable[item]
		return exists
	}
	for _, existing := range seenNonComparable {
		if reflect.DeepEqual(existing, item) {
			return true
		}
	}
	return false
}

// Filter returns elements for which predicate returns true.
func Filter(slice any, predicate func(any) bool) (any, error) {
	if slice == nil {
		return nil, nil
	}
	sliceAny, err := ToAny(slice)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSlice, err)
	}
	result := make([]any, 0, len(sliceAny))
	for _, item := range sliceAny {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return restoreType(result, slice, nil)
}

// Map transforms each element using mapper and returns the
// resulting slice.
func Map(slice any, mapper func(any) any) (any, error) {
	if slice == nil {
		return nil, nil
	}
	sliceAny, err := ToAny(slice)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSlice, err)
	}
	result := make([]any, len(sliceAny))
	for i, item := range sliceAny {
		result[i] = mapper(item)
	}
	return restoreType(result, slice, nil)
}

// Join formats each element with fmt.Sprintf and joins them
// with separator. Returns "" for empty or invalid input.
func Join(slice any, separator string) string {
	sliceAny, err := ToAny(slice)
	if err != nil || len(sliceAny) == 0 {
		return ""
	}
	strs := make([]string, len(sliceAny))
	for i, v := range sliceAny {
		if v != nil {
			strs[i] = fmt.Sprintf("%v", v)
		}
	}
	return strings.Join(strs, separator)
}

// restoreType attempts to convert []any back to the concrete
// slice type of the first non-nil original argument.
func restoreType(
	merged []any, original1, original2 any,
) (any, error) {
	var originalType reflect.Type
	switch {
	case original1 != nil:
		originalType = reflect.TypeOf(original1)
	case original2 != nil:
		originalType = reflect.TypeOf(original2)
	default:
		return merged, nil
	}
	if originalType == reflect.TypeFor[[]any]() {
		return merged, nil
	}
	if originalType.Kind() != reflect.Slice {
		return merged, nil
	}
	elemType := originalType.Elem()
	result := reflect.MakeSlice(originalType, len(merged), len(merged))
	for i, item := range merged {
		iv := reflect.ValueOf(item)
		switch {
		case iv.Type().ConvertibleTo(elemType):
			result.Index(i).Set(iv.Convert(elemType))
		case iv.Type().AssignableTo(elemType):
			result.Index(i).Set(iv)
		default:
			return merged, nil
		}
	}
	return result.Interface(), nil
}
