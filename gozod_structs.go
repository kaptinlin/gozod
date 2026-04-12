package gozod

import "github.com/kaptinlin/gozod/types"

type FromStructOption = types.FromStructOption

func WithTagName(name string) FromStructOption {
	return types.WithTagName(name)
}

func FromStruct[T any](opts ...FromStructOption) *types.ZodStruct[T, T] {
	return types.FromStruct[T](opts...)
}

func FromStructPtr[T any](opts ...FromStructOption) *types.ZodStruct[T, *T] {
	return types.FromStructPtr[T](opts...)
}
