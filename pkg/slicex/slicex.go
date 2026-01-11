package slicex

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Sentinel errors for slicex package.
var (
	ErrInvalidReflectValue   = errors.New("invalid reflect value")
	ErrNotSliceArrayOrString = errors.New("input is not a slice, array, or string")

	ErrCannotConvertSliceElement = errors.New("cannot convert slice element")
	ErrCannotConvertFirstSlice   = errors.New("cannot convert first slice")
	ErrCannotConvertSecondSlice  = errors.New("cannot convert second slice")
	ErrCannotConvertSlice        = errors.New("cannot convert slice")
)

// =============================================================================
// SLICE CONVERSION FUNCTIONS
// =============================================================================

// ToAny converts any slice/array to []any
func ToAny(input any) ([]any, error) {
	if input == nil {
		return nil, nil
	}

	switch v := input.(type) {
	case []any:
		return v, nil
	case []string:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []int:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []int32:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []int64:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []float32:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []float64:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []bool:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case string:
		// Treat entire string as a single element when coercing to []any
		return []any{v}, nil
	default:
		// Try reflection-based conversion
		return FromReflect(reflect.ValueOf(input))
	}
}

// ToTyped converts any slice to []T using generics
func ToTyped[T any](input any) ([]T, error) {
	if input == nil {
		return nil, nil
	}

	// Convert to []any first
	anySlice, err := ToAny(input)
	if err != nil {
		return nil, err
	}

	result := make([]T, len(anySlice))
	for i, item := range anySlice {
		if typedItem, ok := item.(T); ok {
			result[i] = typedItem
		} else {
			// Try type conversion
			itemValue := reflect.ValueOf(item)
			targetType := reflect.TypeOf((*T)(nil)).Elem()

			if itemValue.Type().ConvertibleTo(targetType) {
				converted := itemValue.Convert(targetType)
				result[i] = converted.Interface().(T)
			} else {
				return nil, fmt.Errorf("%w: cannot convert %v to type %v", ErrCannotConvertSliceElement, item, targetType)
			}
		}
	}

	return result, nil
}

// ToStrings converts any slice to []string
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

// FromReflect converts reflect.Value to []any
func FromReflect(rv reflect.Value) ([]any, error) {
	if !rv.IsValid() {
		return nil, ErrInvalidReflectValue
	}

	//nolint:exhaustive // Only Slice, Array, String are valid; all others return same error
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		length := rv.Len()
		result := make([]any, length)
		for i := range length {
			result[i] = rv.Index(i).Interface()
		}
		return result, nil
	case reflect.String:
		// Convert string to slice of characters
		str := rv.String()
		return StringToChars(str), nil
	default:
		return nil, fmt.Errorf("input is not a slice, array, or string, got %v: %w", rv.Kind(), ErrNotSliceArrayOrString)
	}
}

// StringToChars converts a string to []any of characters
func StringToChars(s string) []any {
	runes := []rune(s)
	result := make([]any, len(runes))
	for i, r := range runes {
		result[i] = string(r)
	}
	return result
}

// =============================================================================
// SLICE EXTRACTION FUNCTIONS
// =============================================================================

// Extract extracts slice from input, returns the slice and whether extraction was successful
func Extract(input any) ([]any, bool) {
	result, err := ToAny(input)
	return result, err == nil
}

// ExtractArray extracts array from input, returns the slice and whether extraction was successful
func ExtractArray(input any) ([]any, bool) {
	if input == nil {
		return nil, false
	}
	rv := reflect.ValueOf(input)
	if rv.Kind() != reflect.Array {
		return nil, false
	}
	result, err := ToAny(input)
	return result, err == nil
}

// ExtractSlice extracts slice from input, returns the slice and whether extraction was successful
func ExtractSlice(input any) ([]any, bool) {
	if input == nil {
		return nil, false
	}
	rv := reflect.ValueOf(input)
	if rv.Kind() != reflect.Slice {
		return nil, false
	}
	result, err := ToAny(input)
	return result, err == nil
}

// =============================================================================
// SLICE MERGE FUNCTIONS
// =============================================================================

// Merge merges two slices of any type
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

	// Convert both to []any first
	sliceA, errA := ToAny(a)
	if errA != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertFirstSlice, errA)
	}

	sliceB, errB := ToAny(b)
	if errB != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSecondSlice, errB)
	}

	merged := append(sliceA, sliceB...) //nolint:gocritic // Creating new slice intentionally
	return convertToOriginalType(merged, a, b)
}

// Append appends elements to a slice
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

	return convertToOriginalType(result, slice, nil)
}

// Prepend prepends elements to a slice
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

	return convertToOriginalType(result, slice, nil)
}

// =============================================================================
// SLICE UTILITY FUNCTIONS
// =============================================================================

// Length returns the length of a slice or array
func Length(input any) (int, error) {
	if input == nil {
		return 0, nil
	}

	rv := reflect.ValueOf(input)
	switch rv.Kind() { //nolint:exhaustive // Only handling slice, array, and string types
	case reflect.Slice, reflect.Array, reflect.String:
		return rv.Len(), nil
	default:
		return 0, ErrNotSliceArrayOrString
	}
}

// IsEmpty checks if a slice is empty
func IsEmpty(input any) bool {
	length, err := Length(input)
	return err != nil || length == 0
}

// Contains checks if a slice contains a specific value
func Contains(slice any, value any) bool {
	sliceAny, err := ToAny(slice)
	if err != nil {
		return false
	}

	return slices.ContainsFunc(sliceAny, func(item any) bool {
		return reflect.DeepEqual(item, value)
	})
}

// IndexOf returns the index of the first occurrence of value in slice
func IndexOf(slice any, value any) int {
	sliceAny, err := ToAny(slice)
	if err != nil {
		return -1
	}

	return slices.IndexFunc(sliceAny, func(item any) bool {
		return reflect.DeepEqual(item, value)
	})
}

// Reverse reverses a slice
func Reverse(slice any) (any, error) {
	if slice == nil {
		return nil, nil
	}
	sliceAny, err := ToAny(slice)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSlice, err)
	}
	slices.Reverse(sliceAny)
	return convertToOriginalType(sliceAny, slice, nil)
}

// Unique removes duplicate elements from a slice
func Unique(slice any) (any, error) {
	if slice == nil {
		return nil, nil
	}
	sliceAny, err := ToAny(slice)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSlice, err)
	}

	// Use two-phase deduplication:
	// 1. Comparable types are recorded using a map.
	// 2. Non-comparable types are recorded using sequential scanning and reflect.DeepEqual.

	seenComparable := make(map[any]struct{})
	var seenNonComparable []any

	result := make([]any, 0, len(sliceAny))

	for _, item := range sliceAny {
		if item == nil {
			// nil can be directly compared
			if _, ok := seenComparable[item]; !ok {
				seenComparable[item] = struct{}{}
				result = append(result, item)
			}
			continue
		}

		itemType := reflect.TypeOf(item)
		if itemType.Comparable() {
			if _, ok := seenComparable[item]; !ok {
				seenComparable[item] = struct{}{}
				result = append(result, item)
			}
		} else {
			// Use DeepEqual sequential scanning for non-comparable types
			duplicated := false
			for _, existing := range seenNonComparable {
				if reflect.DeepEqual(existing, item) {
					duplicated = true
					break
				}
			}
			if !duplicated {
				seenNonComparable = append(seenNonComparable, item)
				result = append(result, item)
			}
		}
	}

	return convertToOriginalType(result, slice, nil)
}

// Filter filters a slice based on a predicate function
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

	return convertToOriginalType(result, slice, nil)
}

// Map transforms each element of a slice using a mapper function
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

	return convertToOriginalType(result, slice, nil)
}

// Join joins slice elements with a separator
func Join(slice any, separator string) string {
	sliceAny, err := ToAny(slice)
	if err != nil {
		return ""
	}

	if len(sliceAny) == 0 {
		return ""
	}

	strValues := make([]string, len(sliceAny))
	for i, v := range sliceAny {
		if v != nil {
			strValues[i] = fmt.Sprintf("%v", v)
		}
	}

	return strings.Join(strValues, separator)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// convertToOriginalType attempts to convert merged slice back to original type
func convertToOriginalType(merged []any, original1, original2 any) (any, error) {
	// Get the type of the first non-nil original
	var originalType reflect.Type
	switch {
	case original1 != nil:
		originalType = reflect.TypeOf(original1)
	case original2 != nil:
		originalType = reflect.TypeOf(original2)
	default:
		return merged, nil
	}

	// If original was []any, return as-is
	if originalType == reflect.TypeOf([]any{}) {
		return merged, nil
	}

	// Try to convert to original slice type
	if originalType.Kind() == reflect.Slice {
		elemType := originalType.Elem()
		result := reflect.MakeSlice(originalType, len(merged), len(merged))

		for i, item := range merged {
			itemValue := reflect.ValueOf(item)
			switch {
			case itemValue.Type().ConvertibleTo(elemType):
				converted := itemValue.Convert(elemType)
				result.Index(i).Set(converted)
			case itemValue.Type().AssignableTo(elemType):
				result.Index(i).Set(itemValue)
			default:
				// Cannot convert, return []any
				return merged, nil
			}
		}

		return result.Interface(), nil
	}

	return merged, nil
}
