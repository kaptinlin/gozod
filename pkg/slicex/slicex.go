package slicex

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/maphash"
	"math"
	"reflect"
	"slices"
	"sort"
	"strings"
)

// Sentinel errors for input validation.
var (
	ErrInvalidReflectValue = errors.New("invalid reflect value")
	ErrNotCollection       = errors.New("input is not a slice, array, or string")
)

// Sentinel errors for type conversion.
var (
	ErrCannotConvert        = errors.New("cannot convert slice")
	ErrCannotConvertElement = errors.New("cannot convert element")
	ErrCannotConvertFirst   = errors.New("cannot convert first slice")
	ErrCannotConvertSecond  = errors.New("cannot convert second slice")
)

// --- Conversion functions ---

// ToAny converts a slice, array, or string to []any.
// A nil input returns (nil, nil). A bare string returns a single-element []any.
// For unrecognized concrete types, reflection is used as a fallback.
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
// Each element is first tried via type assertion, then via reflect conversion.
// Returns ErrCannotConvertElement if an element cannot be converted to the target type.
func ToTyped[T any](input any) ([]T, error) {
	if input == nil {
		return nil, nil
	}
	items, err := ToAny(input)
	if err != nil {
		return nil, err
	}
	target := reflect.TypeFor[T]()
	result := make([]T, len(items))
	for i, v := range items {
		if typed, ok := v.(T); ok {
			result[i] = typed
			continue
		}
		rv := reflect.ValueOf(v)
		if !rv.Type().ConvertibleTo(target) {
			return nil, fmt.Errorf(
				"element %v (type %T) not convertible to %v: %w",
				v, v, target, ErrCannotConvertElement,
			)
		}
		result[i] = rv.Convert(target).Interface().(T)
	}
	return result, nil
}

// ToStrings converts any slice to []string by formatting each element with fmt.Sprintf.
func ToStrings(input any) ([]string, error) {
	items, err := ToAny(input)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(items))
	for i, v := range items {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result, nil
}

// FromReflect converts a reflect.Value (slice, array, or string) to []any.
// Returns ErrInvalidReflectValue for invalid values and ErrNotCollection for unsupported kinds.
func FromReflect(rv reflect.Value) ([]any, error) {
	if !rv.IsValid() {
		return nil, ErrInvalidReflectValue
	}
	//nolint:exhaustive // Only Slice, Array, String are valid.
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
		return nil, fmt.Errorf("got %v: %w", rv.Kind(), ErrNotCollection)
	}
}

// StringToChars converts a string to []any where each element is a single-character string.
func StringToChars(s string) []any {
	runes := []rune(s)
	result := make([]any, len(runes))
	for i, r := range runes {
		result[i] = string(r)
	}
	return result
}

// --- Extraction functions ---

// Extract converts input to []any, returning whether the conversion succeeded.
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

// --- Mutation functions ---

// Merge concatenates two slices. If both inputs share the same concrete slice type,
// the result preserves that type.
// Returns ErrCannotConvertFirst or ErrCannotConvertSecond if the respective input
// cannot be converted.
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
	first, err := ToAny(a)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertFirst, err)
	}
	second, err := ToAny(b)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvertSecond, err)
	}
	result := make([]any, 0, len(first)+len(second))
	result = append(result, first...)
	result = append(result, second...)
	return restoreType(result, a, b)
}

// Append appends elements to a slice, preserving the original slice's concrete type
// when possible.
func Append(s any, elements ...any) (any, error) {
	if s == nil {
		return elements, nil
	}
	items, err := ToAny(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvert, err)
	}
	result := make([]any, len(items)+len(elements))
	copy(result, items)
	copy(result[len(items):], elements)
	return restoreType(result, s, nil)
}

// Prepend inserts elements before a slice, preserving the original slice's concrete type
// when possible.
func Prepend(s any, elements ...any) (any, error) {
	if s == nil {
		return elements, nil
	}
	items, err := ToAny(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvert, err)
	}
	result := make([]any, len(elements)+len(items))
	copy(result, elements)
	copy(result[len(elements):], items)
	return restoreType(result, s, nil)
}

// Reverse returns a new slice with elements in reverse order.
func Reverse(s any) (any, error) {
	if s == nil {
		return nil, nil
	}
	items, err := ToAny(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvert, err)
	}
	slices.Reverse(items)
	return restoreType(items, s, nil)
}

// Unique removes duplicate elements, preserving first-occurrence order.
// Comparable types use a map for O(1) lookup. Non-comparable types use a
// hash-bucketed approach with structural hashing, falling back to
// reflect.DeepEqual only on hash collisions.
func Unique(s any) (any, error) {
	if s == nil {
		return nil, nil
	}
	items, err := ToAny(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvert, err)
	}
	seen := make(map[any]struct{})
	buckets := make(map[uint64][]any)
	result := make([]any, 0, len(items))
	for _, v := range items {
		if isDuplicate(v, seen, buckets) {
			continue
		}
		if v == nil || reflect.TypeOf(v).Comparable() {
			seen[v] = struct{}{}
		} else {
			h := structuralHash(v)
			buckets[h] = append(buckets[h], v)
		}
		result = append(result, v)
	}
	return restoreType(result, s, nil)
}

// Filter returns elements for which fn returns true.
func Filter(s any, fn func(any) bool) (any, error) {
	if s == nil {
		return nil, nil
	}
	items, err := ToAny(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvert, err)
	}
	result := make([]any, 0, len(items))
	for _, v := range items {
		if fn(v) {
			result = append(result, v)
		}
	}
	return restoreType(result, s, nil)
}

// Map transforms each element using fn and returns the resulting slice.
func Map(s any, fn func(any) any) (any, error) {
	if s == nil {
		return nil, nil
	}
	items, err := ToAny(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotConvert, err)
	}
	result := make([]any, len(items))
	for i, v := range items {
		result[i] = fn(v)
	}
	return restoreType(result, s, nil)
}

// --- Query functions ---

// Length returns the length of a slice, array, or string.
// Returns ErrNotCollection for other types.
func Length(input any) (int, error) {
	if input == nil {
		return 0, nil
	}
	rv := reflect.ValueOf(input)
	//nolint:exhaustive // Only handling slice, array, and string.
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		return rv.Len(), nil
	default:
		return 0, ErrNotCollection
	}
}

// IsEmpty reports whether input is nil, empty, or not a recognized collection type.
func IsEmpty(input any) bool {
	length, err := Length(input)
	return err != nil || length == 0
}

// Contains reports whether s contains value, compared using reflect.DeepEqual.
func Contains(s any, value any) bool {
	items, err := ToAny(s)
	if err != nil {
		return false
	}
	return slices.ContainsFunc(items, func(v any) bool {
		return reflect.DeepEqual(v, value)
	})
}

// IndexOf returns the index of the first occurrence of value in s, or -1 if not found.
// Comparison uses reflect.DeepEqual.
func IndexOf(s any, value any) int {
	items, err := ToAny(s)
	if err != nil {
		return -1
	}
	return slices.IndexFunc(items, func(v any) bool {
		return reflect.DeepEqual(v, value)
	})
}

// Join formats each element with fmt.Sprintf and joins them with sep.
// Returns "" for empty or invalid input.
func Join(s any, sep string) string {
	items, err := ToAny(s)
	if err != nil || len(items) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range items {
		if i > 0 {
			b.WriteString(sep)
		}
		if v != nil {
			fmt.Fprintf(&b, "%v", v)
		}
	}
	return b.String()
}

// --- Internal helpers ---

// toAnySlice converts a typed slice to []any, eliminating repeated
// make+loop boilerplate across type-switch branches.
func toAnySlice[T any](v []T) []any {
	result := make([]any, len(v))
	for i, item := range v {
		result[i] = item
	}
	return result
}

// hashSeed is a per-process seed for structural hashing of non-comparable types.
var hashSeed = maphash.MakeSeed()

// isDuplicate checks whether v has already been seen.
// Comparable types are looked up in the seen map. Non-comparable types are
// looked up in buckets keyed by structural hash, falling back to DeepEqual
// only on hash collisions.
func isDuplicate(v any, seen map[any]struct{}, buckets map[uint64][]any) bool {
	if v == nil || reflect.TypeOf(v).Comparable() {
		_, ok := seen[v]
		return ok
	}
	h := structuralHash(v)
	for _, existing := range buckets[h] {
		if reflect.DeepEqual(existing, v) {
			return true
		}
	}
	return false
}

// structuralHash computes a hash for a value based on its structure.
// It handles common JSON-like types (nil, bool, float64, int, string,
// []any, map[string]any) directly, and falls back to fmt.Sprintf for
// other types.
func structuralHash(v any) uint64 {
	var h maphash.Hash
	h.SetSeed(hashSeed)
	writeHash(&h, v)
	return h.Sum64()
}

// writeHash writes a structural representation of v into the hash.
func writeHash(h *maphash.Hash, v any) {
	if v == nil {
		h.WriteByte(0)
		return
	}
	switch val := v.(type) {
	case bool:
		h.WriteByte(1)
		if val {
			h.WriteByte(1)
		} else {
			h.WriteByte(0)
		}
	case float64:
		h.WriteByte(2)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(val))
		h.Write(buf[:])
	case int:
		h.WriteByte(3)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(val))
		h.Write(buf[:])
	case string:
		h.WriteByte(4)
		h.WriteString(val)
	case []any:
		h.WriteByte(5)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(len(val)))
		h.Write(buf[:])
		for _, elem := range val {
			writeHash(h, elem)
		}
	case map[string]any:
		h.WriteByte(6)
		// Sort keys for deterministic ordering.
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(len(keys)))
		h.Write(buf[:])
		for _, k := range keys {
			h.WriteString(k)
			writeHash(h, val[k])
		}
	default:
		// Fallback: use fmt representation for unknown types.
		h.WriteByte(7)
		h.WriteString(fmt.Sprintf("%v", v))
	}
}

// restoreType attempts to convert []any back to the concrete slice type
// of the first non-nil original argument.
func restoreType(items []any, orig, fallback any) (any, error) {
	typ := resolveType(orig, fallback)
	if typ == nil || typ == reflect.TypeFor[[]any]() || typ.Kind() != reflect.Slice {
		return items, nil
	}
	elem := typ.Elem()
	result := reflect.MakeSlice(typ, len(items), len(items))
	for i, v := range items {
		rv := reflect.ValueOf(v)
		switch {
		case rv.Type().AssignableTo(elem):
			result.Index(i).Set(rv)
		case rv.Type().ConvertibleTo(elem):
			result.Index(i).Set(rv.Convert(elem))
		default:
			return items, nil
		}
	}
	return result.Interface(), nil
}

// resolveType returns the reflect.Type of the first non-nil argument.
func resolveType(orig, fallback any) reflect.Type {
	if orig != nil {
		return reflect.TypeOf(orig)
	}
	if fallback != nil {
		return reflect.TypeOf(fallback)
	}
	return nil
}
