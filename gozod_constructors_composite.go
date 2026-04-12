package gozod

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
)

var (
	Array                 = types.Array
	ArrayPtr              = types.ArrayPtr
	Map                   = types.Map
	MapPtr                = types.MapPtr
	Tuple                 = types.Tuple
	TupleWithRest         = types.TupleWithRest
	TuplePtr              = types.TuplePtr
	LooseRecord           = types.LooseRecord
	LooseRecordPtr        = types.LooseRecordPtr
	Object                = types.Object
	ObjectPtr             = types.ObjectPtr
	StrictObject          = types.StrictObject
	StrictObjectPtr       = types.StrictObjectPtr
	LooseObject           = types.LooseObject
	LooseObjectPtr        = types.LooseObjectPtr
	Union                 = types.Union
	UnionPtr              = types.UnionPtr
	Xor                   = types.Xor
	XorPtr                = types.XorPtr
	XorOf                 = types.XorOf
	Intersection          = types.Intersection
	IntersectionPtr       = types.IntersectionPtr
	DiscriminatedUnion    = types.DiscriminatedUnion
	DiscriminatedUnionPtr = types.DiscriminatedUnionPtr
)

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
