package gozod

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
)

func Array(args ...any) *ZodArray[[]any, []any] {
	return types.Array(args...)
}

func ArrayPtr(args ...any) *ZodArray[[]any, *[]any] {
	return types.ArrayPtr(args...)
}

func Map(keySchema, valueSchema any, paramArgs ...any) *ZodMap[map[any]any, map[any]any] {
	return types.Map(keySchema, valueSchema, paramArgs...)
}

func MapPtr(keySchema, valueSchema any, paramArgs ...any) *ZodMap[map[any]any, *map[any]any] {
	return types.MapPtr(keySchema, valueSchema, paramArgs...)
}

func Tuple(items ...core.ZodSchema) *ZodTuple[[]any, []any] {
	return types.Tuple(items...)
}

func TupleWithRest(items []core.ZodSchema, rest core.ZodSchema, params ...any) *ZodTuple[[]any, []any] {
	return types.TupleWithRest(items, rest, params...)
}

func TuplePtr(items ...core.ZodSchema) *ZodTuple[[]any, *[]any] {
	return types.TuplePtr(items...)
}

func LooseRecord(keySchema, valueSchema any, paramArgs ...any) *ZodRecord[map[string]any, map[string]any] {
	return types.LooseRecord(keySchema, valueSchema, paramArgs...)
}

func LooseRecordPtr(keySchema, valueSchema any, paramArgs ...any) *ZodRecord[map[string]any, *map[string]any] {
	return types.LooseRecordPtr(keySchema, valueSchema, paramArgs...)
}

func Object(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return types.Object(shape, params...)
}

func ObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return types.ObjectPtr(shape, params...)
}

func StrictObject(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return types.StrictObject(shape, params...)
}

func StrictObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return types.StrictObjectPtr(shape, params...)
}

func LooseObject(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return types.LooseObject(shape, params...)
}

func LooseObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return types.LooseObjectPtr(shape, params...)
}

func Union(options []any, args ...any) *ZodUnion[any, any] {
	return types.Union(options, args...)
}

func UnionPtr(options []any, args ...any) *ZodUnion[any, *any] {
	return types.UnionPtr(options, args...)
}

func Xor(options []any, args ...any) *types.ZodXor[any, any] {
	return types.Xor(options, args...)
}

func XorPtr(options []any, args ...any) *types.ZodXor[any, *any] {
	return types.XorPtr(options, args...)
}

func XorOf(schemas ...any) *types.ZodXor[any, any] {
	return types.XorOf(schemas...)
}

func Intersection(left, right any, args ...any) *ZodIntersection[any, any] {
	return types.Intersection(left, right, args...)
}

func IntersectionPtr(left, right any, args ...any) *ZodIntersection[any, *any] {
	return types.IntersectionPtr(left, right, args...)
}

func DiscriminatedUnion(disc string, options []core.ZodSchema, args ...any) (*ZodDiscriminatedUnion[any, any], error) {
	return types.DiscriminatedUnion(disc, options, args...)
}

func DiscriminatedUnionPtr(disc string, options []core.ZodSchema, args ...any) (*ZodDiscriminatedUnion[any, *any], error) {
	return types.DiscriminatedUnionPtr(disc, options, args...)
}

func MustDiscriminatedUnion(disc string, options []core.ZodSchema, args ...any) *ZodDiscriminatedUnion[any, any] {
	return types.MustDiscriminatedUnion(disc, options, args...)
}

func MustDiscriminatedUnionPtr(disc string, options []core.ZodSchema, args ...any) *ZodDiscriminatedUnion[any, *any] {
	return types.MustDiscriminatedUnionPtr(disc, options, args...)
}

func Set[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, map[T]struct{}] {
	return types.Set[T](valueSchema, paramArgs...)
}

func SetPtr[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, *map[T]struct{}] {
	return types.SetPtr[T](valueSchema, paramArgs...)
}

func Record[K any, V any](keySchema any, valueSchema core.ZodType[V], paramArgs ...any) *ZodRecord[map[string]V, map[string]V] {
	return types.RecordTyped[map[string]V, map[string]V](keySchema, valueSchema, paramArgs...)
}

func RecordPtr[K any, V any](keySchema any, valueSchema core.ZodType[V], paramArgs ...any) *ZodRecord[map[string]V, *map[string]V] {
	return types.RecordTyped[map[string]V, *map[string]V](keySchema, valueSchema, paramArgs...)
}

func Slice[T any](elementSchema any, paramArgs ...any) *ZodSlice[T, []T] {
	return types.Slice[T](elementSchema, paramArgs...)
}

func SlicePtr[T any](elementSchema any, paramArgs ...any) *ZodSlice[T, *[]T] {
	return types.SlicePtr[T](elementSchema, paramArgs...)
}

func Struct[T any](params ...any) *ZodStruct[T, T] {
	return types.Struct[T](params...)
}

func StructPtr[T any](params ...any) *ZodStruct[T, *T] {
	return types.StructPtr[T](params...)
}
