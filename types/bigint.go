package types

import (
	"math/big"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// BigIntegerConstraint restricts values to *big.Int or **big.Int.
type BigIntegerConstraint interface {
	*big.Int | **big.Int
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodBigIntDef defines the configuration for big.Int validation
type ZodBigIntDef struct {
	core.ZodTypeDef
}

// ZodBigIntInternals contains big.Int validator internal state
type ZodBigIntInternals struct {
	core.ZodTypeInternals
	Def *ZodBigIntDef // Schema definition
}

// ZodBigInt represents a big.Int validation schema with type safety
type ZodBigInt[T BigIntegerConstraint] struct {
	internals *ZodBigIntInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodBigInt[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodBigInt[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodBigInt[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements Coercible interface for big.Int type conversion
func (z *ZodBigInt[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToBigInt(input)
	return result, err == nil
}

// Parse returns a value that matches the generic type T with full type safety.
func (z *ZodBigInt[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	// Handle nil input explicitly before ParsePrimitive
	if input == nil {
		var zero T
		var parseCtx *core.ParseContext
		if len(ctx) > 0 {
			parseCtx = ctx[0]
		} else {
			parseCtx = &core.ParseContext{}
		}

		// Check modifiers for nil handling
		internals := &z.internals.ZodTypeInternals

		// NonOptional flag - nil not allowed
		if internals.NonOptional {
			return zero, issues.CreateNonOptionalError(parseCtx)
		}

		// Default/DefaultFunc - short circuit
		if internals.DefaultValue != nil {
			return engine.ConvertToConstraintType[*big.Int, T](internals.DefaultValue, parseCtx, core.ZodTypeBigInt)
		}
		if internals.DefaultFunc != nil {
			defaultValue := internals.DefaultFunc()
			return engine.ConvertToConstraintType[*big.Int, T](defaultValue, parseCtx, core.ZodTypeBigInt)
		}

		// Prefault/PrefaultFunc - use as new input
		switch {
		case internals.PrefaultValue != nil:
			input = internals.PrefaultValue
		case internals.PrefaultFunc != nil:
			input = internals.PrefaultFunc()
		case internals.Optional || internals.Nilable:
			// Optional/Nilable - allow nil
			return engine.ConvertToConstraintType[*big.Int, T](nil, parseCtx, core.ZodTypeBigInt)
		default:
			// Reject nil input
			return zero, issues.CreateInvalidTypeError(core.ZodTypeBigInt, nil, parseCtx)
		}
	}

	return engine.ParsePrimitive[*big.Int, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeBigInt,
		engine.ApplyChecks[*big.Int],
		engine.ConvertToConstraintType[*big.Int, T],
		ctx...,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodBigInt[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodBigInt[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodBigInt[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Create validator that applies checks to the constraint type T
	validator := func(value *big.Int, checks []core.ZodCheck, c *core.ParseContext) (*big.Int, error) {
		return engine.ApplyChecks[*big.Int](value, checks, c)
	}

	return engine.ParsePrimitiveStrict[*big.Int, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeBigInt,
		validator,
		ctx...,
	)
}

// MustStrictParse provides compile-time type safety and panics on validation failure.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodBigInt[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns **big.Int for nullable semantics
func (z *ZodBigInt[T]) Optional() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns **big.Int
func (z *ZodBigInt[T]) Nilable() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodBigInt[T]) Nullish() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes optional flag and returns *big.Int constraint (single pointer).
// This is useful after Optional()/Nilable() when nil values should be disallowed again.
func (z *ZodBigInt[T]) NonOptional() *ZodBigInt[*big.Int] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodBigInt[*big.Int]{
		internals: &ZodBigIntInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default preserves current generic type T
func (z *ZodBigInt[T]) Default(v *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodBigInt[T]) DefaultFunc(fn func() *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodBigInt[T]) Prefault(v *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps current generic type T.
func (z *ZodBigInt[T]) PrefaultFunc(fn func() *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this bigint schema.
func (z *ZodBigInt[T]) Meta(meta core.GlobalMeta) *ZodBigInt[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min adds minimum value validation
func (z *ZodBigInt[T]) Min(minimum *big.Int, params ...any) *ZodBigInt[T] {
	check := checks.Gte(minimum, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max adds maximum value validation
func (z *ZodBigInt[T]) Max(maximum *big.Int, params ...any) *ZodBigInt[T] {
	check := checks.Lte(maximum, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Gt adds greater than validation (exclusive)
func (z *ZodBigInt[T]) Gt(value *big.Int, params ...any) *ZodBigInt[T] {
	check := checks.Gt(value, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodBigInt[T]) Gte(value *big.Int, params ...any) *ZodBigInt[T] {
	check := checks.Gte(value, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Lt adds less than validation (exclusive)
func (z *ZodBigInt[T]) Lt(value *big.Int, params ...any) *ZodBigInt[T] {
	check := checks.Lt(value, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodBigInt[T]) Lte(value *big.Int, params ...any) *ZodBigInt[T] {
	check := checks.Lte(value, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Positive adds positive number validation (> 0)
func (z *ZodBigInt[T]) Positive(params ...any) *ZodBigInt[T] {
	return z.Gt(big.NewInt(0), params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodBigInt[T]) Negative(params ...any) *ZodBigInt[T] {
	return z.Lt(big.NewInt(0), params...)
}

// NonNegative adds non-negative number validation (>= 0)
func (z *ZodBigInt[T]) NonNegative(params ...any) *ZodBigInt[T] {
	return z.Gte(big.NewInt(0), params...)
}

// NonPositive adds non-positive number validation (<= 0)
func (z *ZodBigInt[T]) NonPositive(params ...any) *ZodBigInt[T] {
	return z.Lte(big.NewInt(0), params...)
}

// MultipleOf adds multiple of validation
func (z *ZodBigInt[T]) MultipleOf(value *big.Int, params ...any) *ZodBigInt[T] {
	check := checks.MultipleOf(value, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates type-safe transformation pipeline
func (z *ZodBigInt[T]) Transform(fn func(*big.Int, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		bigIntValue := extractBigInt(input)
		return fn(bigIntValue, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms big.Int value while keeping the same type
func (z *ZodBigInt[T]) Overwrite(transform func(T) T, params ...any) *ZodBigInt[T] {
	check := checks.NewZodCheckOverwrite(func(input any) any {
		// Convert input to the correct constraint type
		if converted, ok := convertToBigIntType[T](input); ok {
			return transform(converted)
		}
		return input
	}, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates validation pipeline to another schema
func (z *ZodBigInt[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	targetFn := func(input T, ctx *core.ParseContext) (any, error) {
		bigIntValue := extractBigInt(input)
		return target.Parse(bigIntValue, ctx)
	}
	return core.NewZodPipe[T, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation with automatic type conversion
func (z *ZodBigInt[T]) Refine(fn func(T) bool, params ...any) *ZodBigInt[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case *big.Int:
			// Schema output is *big.Int
			if v == nil {
				return false // nil should never reach here for bigint schema
			}
			if bigIntVal, ok := v.(*big.Int); ok {
				return fn(any(bigIntVal).(T))
			}
			return false
		case **big.Int:
			// Schema output is **big.Int – convert incoming value to **big.Int
			if v == nil {
				return fn(any((**big.Int)(nil)).(T))
			}
			if bigIntVal, ok := v.(*big.Int); ok {
				ptr := &bigIntVal
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false // Unsupported type
		}
	}

	// Use checks package for custom validation
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodBigInt[T]) RefineAny(fn func(any) bool, params ...any) *ZodBigInt[T] {
	// MUST use checks package for custom validation
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// convertToBigIntType converts any value to the specified BigInt constraint type with strict type checking
func convertToBigIntType[T BigIntegerConstraint](v any) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case *big.Int:
		// For *big.Int constraint
		if bigIntVal, ok := v.(*big.Int); ok {
			return any(bigIntVal).(T), true
		}
		return zero, false
	case **big.Int:
		// For **big.Int constraint
		if bigIntVal, ok := v.(*big.Int); ok {
			ptr := &bigIntVal
			return any(ptr).(T), true
		}
		if ptrPtr, ok := v.(**big.Int); ok {
			return any(ptrPtr).(T), true
		}
		return zero, false
	default:
		return zero, false
	}
}

// withPtrInternals creates new instance with **big.Int type
func (z *ZodBigInt[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodBigInt[**big.Int] {
	return &ZodBigInt[**big.Int]{internals: &ZodBigIntInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates new instance preserving generic type T
func (z *ZodBigInt[T]) withInternals(in *core.ZodTypeInternals) *ZodBigInt[T] {
	return &ZodBigInt[T]{internals: &ZodBigIntInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodBigInt[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodBigInt[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractBigInt extracts *big.Int value from generic type T
func extractBigInt[T BigIntegerConstraint](value T) *big.Int {
	if ptr, ok := any(value).(**big.Int); ok {
		if ptr != nil {
			return *ptr
		}
		return nil
	}
	return any(value).(*big.Int)
}

// newZodBigIntFromDef constructs new ZodBigInt from definition
func newZodBigIntFromDef[T BigIntegerConstraint](def *ZodBigIntDef) *ZodBigInt[T] {
	internals := &ZodBigIntInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		bigIntDef := &ZodBigIntDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodBigIntFromDef[T](bigIntDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodBigInt[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// BigInt creates *big.Int schema with type-inference support
func BigInt(params ...any) *ZodBigInt[*big.Int] {
	return BigIntTyped[*big.Int](params...)
}

// BigIntPtr creates schema for **big.Int
func BigIntPtr(params ...any) *ZodBigInt[**big.Int] {
	return BigIntTyped[**big.Int](params...)
}

// BigIntTyped is the generic constructor for big.Int schemas
func BigIntTyped[T BigIntegerConstraint](params ...any) *ZodBigInt[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodBigIntDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeBigInt,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply normalized parameters to schema definition
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodBigIntFromDef[T](def)
}

// CoercedBigInt creates coerced *big.Int schema
func CoercedBigInt(params ...any) *ZodBigInt[*big.Int] {
	schema := BigInt(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedBigIntPtr creates coerced **big.Int schema
func CoercedBigIntPtr(params ...any) *ZodBigInt[**big.Int] {
	schema := BigIntPtr(params...)
	schema.internals.SetCoerce(true)
	return schema
}
