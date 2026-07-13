package gozod

import "github.com/kaptinlin/gozod/types"

func Any(params ...any) *ZodAny[any, any] {
	return types.Any(params...)
}

func AnyPtr(params ...any) *ZodAny[any, *any] {
	return types.AnyPtr(params...)
}

func Unknown(params ...any) *ZodUnknown[any, any] {
	return types.Unknown(params...)
}

func UnknownPtr(params ...any) *ZodUnknown[any, *any] {
	return types.UnknownPtr(params...)
}

func Never(params ...any) *ZodNever[any, any] {
	return types.Never(params...)
}

func NeverPtr(params ...any) *ZodNever[any, *any] {
	return types.NeverPtr(params...)
}

func Nil(params ...any) *ZodNil[any, any] {
	return types.Nil(params...)
}

func NilPtr(params ...any) *ZodNil[any, *any] {
	return types.NilPtr(params...)
}

func File(params ...any) *ZodFile[any, any] {
	return types.File(params...)
}

func FilePtr(params ...any) *ZodFile[any, *any] {
	return types.FilePtr(params...)
}

func Function(params ...any) *ZodFunction[any] {
	return types.Function(params...)
}

func FunctionPtr(params ...any) *ZodFunction[*any] {
	return types.FunctionPtr(params...)
}

func StringBool(params ...any) *ZodStringBool[bool] {
	return types.StringBool(params...)
}

func StringBoolPtr(params ...any) *ZodStringBool[*bool] {
	return types.StringBoolPtr(params...)
}

func LazyAny(getter func() any, params ...any) *ZodLazy[any] {
	return types.LazyAny(getter, params...)
}

func LazyPtr(getter func() any, params ...any) *ZodLazy[*any] {
	return types.LazyPtr(getter, params...)
}

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

func LazyTyped[T types.LazyConstraint](getter func() any, params ...any) *types.ZodLazyOutput[T] {
	return types.LazyTyped[T](getter, params...)
}
