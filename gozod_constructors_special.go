package gozod

import "github.com/kaptinlin/gozod/types"

var (
	Any           = types.Any
	AnyPtr        = types.AnyPtr
	Unknown       = types.Unknown
	UnknownPtr    = types.UnknownPtr
	Never         = types.Never
	NeverPtr      = types.NeverPtr
	Nil           = types.Nil
	NilPtr        = types.NilPtr
	File          = types.File
	FilePtr       = types.FilePtr
	Function      = types.Function
	FunctionPtr   = types.FunctionPtr
	StringBool    = types.StringBool
	StringBoolPtr = types.StringBoolPtr
	LazyAny       = types.LazyAny
	LazyPtr       = types.LazyPtr
)

func Literal[T comparable](value T, params ...any) *ZodLiteral[T, T] {
	return types.Literal(value, params...)
}

func LiteralPtr[T comparable](value T, params ...any) *ZodLiteral[T, *T] {
	return types.LiteralPtr(value, params...)
}

func LiteralOf[T comparable](values []T, params ...any) *ZodLiteral[T, T] {
	return types.LiteralOf(values, params...)
}

func LiteralPtrOf[T comparable](values []T, params ...any) *ZodLiteral[T, *T] {
	return types.LiteralPtrOf(values, params...)
}

func Enum[T comparable](values ...T) *ZodEnum[T, T] {
	return types.Enum(values...)
}

func EnumSlice[T comparable](values []T) *ZodEnum[T, T] {
	return types.EnumSlice(values)
}

func EnumMap[T comparable](entries map[string]T, params ...any) *ZodEnum[T, T] {
	return types.EnumMap(entries, params...)
}

func EnumPtr[T comparable](values ...T) *ZodEnum[T, *T] {
	return types.EnumPtr(values...)
}

func EnumSlicePtr[T comparable](values []T) *ZodEnum[T, *T] {
	return types.EnumSlicePtr(values)
}

func EnumMapPtr[T comparable](entries map[string]T, params ...any) *ZodEnum[T, *T] {
	return types.EnumMapPtr(entries, params...)
}

func Lazy[S types.ZodSchemaType](getter func() S, params ...any) *types.ZodLazyTyped[S] {
	return types.Lazy(getter, params...)
}
