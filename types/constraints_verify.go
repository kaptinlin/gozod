package types

// Compile-time interface constraint verification.
//
// These assertions use Go 1.26 self-referential generic constraints
// (core.Describable, core.Refineable) to guarantee at compile time
// that all schema types implement the expected chaining API.
//
// If a schema type is missing a required method (Describe, Meta,
// RefineAny, or Internals), compilation will fail with a clear error.

import (
	"math/big"
	"time"

	"github.com/kaptinlin/gozod/core"
)

// --- Describable constraint verification ---
// Ensures each schema type has: Describe(string) Self, Meta(GlobalMeta) Self, Internals()

var (
	_ core.Describable[*ZodAny[any, any]]                             = (*ZodAny[any, any])(nil)
	_ core.Describable[*ZodArray[any, any]]                           = (*ZodArray[any, any])(nil)
	_ core.Describable[*ZodBigInt[*big.Int]]                          = (*ZodBigInt[*big.Int])(nil)
	_ core.Describable[*ZodBool[bool]]                                = (*ZodBool[bool])(nil)
	_ core.Describable[*ZodComplex[complex128]]                       = (*ZodComplex[complex128])(nil)
	_ core.Describable[*ZodDiscriminatedUnion[any, any]]              = (*ZodDiscriminatedUnion[any, any])(nil)
	_ core.Describable[*ZodEnum[string, string]]                      = (*ZodEnum[string, string])(nil)
	_ core.Describable[*ZodFile[string, string]]                      = (*ZodFile[string, string])(nil)
	_ core.Describable[*ZodFloatTyped[float64, float64]]              = (*ZodFloatTyped[float64, float64])(nil)
	_ core.Describable[*ZodFunction[any]]                             = (*ZodFunction[any])(nil)
	_ core.Describable[*ZodIntegerTyped[int, int]]                    = (*ZodIntegerTyped[int, int])(nil)
	_ core.Describable[*ZodIntersection[any, any]]                    = (*ZodIntersection[any, any])(nil)
	_ core.Describable[*ZodLazy[any]]                                 = (*ZodLazy[any])(nil)
	_ core.Describable[*ZodLiteral[string, string]]                   = (*ZodLiteral[string, string])(nil)
	_ core.Describable[*ZodMap[map[any]any, map[any]any]]             = (*ZodMap[map[any]any, map[any]any])(nil)
	_ core.Describable[*ZodNever[any, any]]                           = (*ZodNever[any, any])(nil)
	_ core.Describable[*ZodNil[any, any]]                             = (*ZodNil[any, any])(nil)
	_ core.Describable[*ZodObject[map[string]any, map[string]any]]    = (*ZodObject[map[string]any, map[string]any])(nil)
	_ core.Describable[*ZodRecord[map[string]any, map[string]any]]    = (*ZodRecord[map[string]any, map[string]any])(nil)
	_ core.Describable[*ZodSet[string, map[string]struct{}]]          = (*ZodSet[string, map[string]struct{}])(nil)
	_ core.Describable[*ZodSlice[any, any]]                           = (*ZodSlice[any, any])(nil)
	_ core.Describable[*ZodString[string]]                            = (*ZodString[string])(nil)
	_ core.Describable[*ZodStringBool[bool]]                          = (*ZodStringBool[bool])(nil)
	_ core.Describable[*ZodStruct[any, any]]                          = (*ZodStruct[any, any])(nil)
	_ core.Describable[*ZodTime[time.Time]]                           = (*ZodTime[time.Time])(nil)
	_ core.Describable[*ZodTuple[[]any, []any]]                       = (*ZodTuple[[]any, []any])(nil)
	_ core.Describable[*ZodUnion[any, any]]                           = (*ZodUnion[any, any])(nil)
	_ core.Describable[*ZodUnknown[any, any]]                         = (*ZodUnknown[any, any])(nil)
	_ core.Describable[*ZodXor[any, any]]                             = (*ZodXor[any, any])(nil)
)

// --- Refineable constraint verification ---
// Ensures each schema type has: RefineAny(func(any) bool, ...any) Self, Internals()

var (
	_ core.Refineable[*ZodAny[any, any]]                             = (*ZodAny[any, any])(nil)
	_ core.Refineable[*ZodArray[any, any]]                           = (*ZodArray[any, any])(nil)
	_ core.Refineable[*ZodBigInt[*big.Int]]                          = (*ZodBigInt[*big.Int])(nil)
	_ core.Refineable[*ZodBool[bool]]                                = (*ZodBool[bool])(nil)
	_ core.Refineable[*ZodComplex[complex128]]                       = (*ZodComplex[complex128])(nil)
	_ core.Refineable[*ZodDiscriminatedUnion[any, any]]              = (*ZodDiscriminatedUnion[any, any])(nil)
	_ core.Refineable[*ZodEnum[string, string]]                      = (*ZodEnum[string, string])(nil)
	_ core.Refineable[*ZodFile[string, string]]                      = (*ZodFile[string, string])(nil)
	_ core.Refineable[*ZodFloatTyped[float64, float64]]              = (*ZodFloatTyped[float64, float64])(nil)
	_ core.Refineable[*ZodFunction[any]]                             = (*ZodFunction[any])(nil)
	_ core.Refineable[*ZodIntegerTyped[int, int]]                    = (*ZodIntegerTyped[int, int])(nil)
	_ core.Refineable[*ZodIntersection[any, any]]                    = (*ZodIntersection[any, any])(nil)
	_ core.Refineable[*ZodLazy[any]]                                 = (*ZodLazy[any])(nil)
	_ core.Refineable[*ZodLiteral[string, string]]                   = (*ZodLiteral[string, string])(nil)
	_ core.Refineable[*ZodMap[map[any]any, map[any]any]]             = (*ZodMap[map[any]any, map[any]any])(nil)
	_ core.Refineable[*ZodNever[any, any]]                           = (*ZodNever[any, any])(nil)
	_ core.Refineable[*ZodNil[any, any]]                             = (*ZodNil[any, any])(nil)
	_ core.Refineable[*ZodObject[map[string]any, map[string]any]]    = (*ZodObject[map[string]any, map[string]any])(nil)
	_ core.Refineable[*ZodRecord[map[string]any, map[string]any]]    = (*ZodRecord[map[string]any, map[string]any])(nil)
	_ core.Refineable[*ZodSet[string, map[string]struct{}]]          = (*ZodSet[string, map[string]struct{}])(nil)
	_ core.Refineable[*ZodSlice[any, any]]                           = (*ZodSlice[any, any])(nil)
	_ core.Refineable[*ZodString[string]]                            = (*ZodString[string])(nil)
	_ core.Refineable[*ZodStringBool[bool]]                          = (*ZodStringBool[bool])(nil)
	_ core.Refineable[*ZodStruct[any, any]]                          = (*ZodStruct[any, any])(nil)
	_ core.Refineable[*ZodTime[time.Time]]                           = (*ZodTime[time.Time])(nil)
	_ core.Refineable[*ZodUnion[any, any]]                           = (*ZodUnion[any, any])(nil)
	_ core.Refineable[*ZodUnknown[any, any]]                         = (*ZodUnknown[any, any])(nil)
	_ core.Refineable[*ZodXor[any, any]]                             = (*ZodXor[any, any])(nil)
)
