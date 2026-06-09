package gozod

import "github.com/kaptinlin/gozod/types"

type FromStructOption = types.FromStructOption

func WithTagName(name string) FromStructOption {
	return types.WithTagName(name)
}

// WithFieldNameTag sets the struct tag used for field names (e.g. "yaml",
// "toml"). It defaults to "json".
func WithFieldNameTag(name string) FromStructOption {
	return types.WithFieldNameTag(name)
}

func FromStruct[T any](opts ...FromStructOption) *types.ZodStruct[T, T] {
	return types.FromStruct[T](opts...)
}

func FromStructPtr[T any](opts ...FromStructOption) *types.ZodStruct[T, *T] {
	return types.FromStructPtr[T](opts...)
}
