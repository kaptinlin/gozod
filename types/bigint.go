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

// GetInternals returns the internal state of the schema.
func (z *ZodBigInt[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodBigInt[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodBigInt[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce converts input to *big.Int, implementing the Coercible interface.
func (z *ZodBigInt[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToBigInt(input)
	return result, err == nil
}

// Parse validates input and returns a value matching the generic type T.
func (z *ZodBigInt[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	if input == nil {
		result, newInput, done, err := z.parseNilInput(ctx...)
		if done {
			return result, err
		}
		input = newInput
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

// MustParse validates input and panics on failure.
func (z *ZodBigInt[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any type for runtime interface usage.
func (z *ZodBigInt[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse provides compile-time type safety by requiring exact type T.
func (z *ZodBigInt[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[*big.Int, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeBigInt,
		engine.ApplyChecks[*big.Int],
		ctx...,
	)
}

// MustStrictParse provides compile-time type safety and panics on failure.
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

// Optional returns a new schema that accepts nil, with **big.Int constraint.
func (z *ZodBigInt[T]) Optional() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a new schema that accepts nil values, with **big.Int constraint.
func (z *ZodBigInt[T]) Nilable() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a new schema combining optional and nilable modifiers.
func (z *ZodBigInt[T]) Nullish() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag, returning a *big.Int constraint.
// Useful after Optional()/Nilable() to disallow nil again.
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

// Default sets a fallback value returned when input is nil (short-circuits validation).
func (z *ZodBigInt[T]) Default(v *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits validation).
func (z *ZodBigInt[T]) DefaultFunc(fn func() *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodBigInt[T]) Prefault(v *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodBigInt[T]) PrefaultFunc(fn func() *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this schema in the global registry.
func (z *ZodBigInt[T]) Meta(meta core.GlobalMeta) *ZodBigInt[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodBigInt[T]) Describe(description string) *ZodBigInt[T] {
	newInternals := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := z.withInternals(newInternals)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min adds minimum value validation (>=).
func (z *ZodBigInt[T]) Min(minimum *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Gte(minimum, params...))
}

// Max adds maximum value validation (<=).
func (z *ZodBigInt[T]) Max(maximum *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Lte(maximum, params...))
}

// Gt adds greater-than validation (exclusive).
func (z *ZodBigInt[T]) Gt(value *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Gt(value, params...))
}

// Gte adds greater-than-or-equal validation (inclusive).
func (z *ZodBigInt[T]) Gte(value *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Gte(value, params...))
}

// Lt adds less-than validation (exclusive).
func (z *ZodBigInt[T]) Lt(value *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Lt(value, params...))
}

// Lte adds less-than-or-equal validation (inclusive).
func (z *ZodBigInt[T]) Lte(value *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Lte(value, params...))
}

// Positive adds positive number validation (> 0).
func (z *ZodBigInt[T]) Positive(params ...any) *ZodBigInt[T] {
	return z.Gt(big.NewInt(0), params...)
}

// Negative adds negative number validation (< 0).
func (z *ZodBigInt[T]) Negative(params ...any) *ZodBigInt[T] {
	return z.Lt(big.NewInt(0), params...)
}

// NonNegative adds non-negative number validation (>= 0).
func (z *ZodBigInt[T]) NonNegative(params ...any) *ZodBigInt[T] {
	return z.Gte(big.NewInt(0), params...)
}

// NonPositive adds non-positive number validation (<= 0).
func (z *ZodBigInt[T]) NonPositive(params ...any) *ZodBigInt[T] {
	return z.Lte(big.NewInt(0), params...)
}

// MultipleOf adds multiple-of validation.
func (z *ZodBigInt[T]) MultipleOf(value *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.MultipleOf(value, params...))
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation pipeline.
func (z *ZodBigInt[T]) Transform(fn func(*big.Int, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(extractBigInt(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the value while preserving the same type.
func (z *ZodBigInt[T]) Overwrite(transform func(T) T, params ...any) *ZodBigInt[T] {
	check := checks.NewZodCheckOverwrite(func(input any) any {
		if converted, ok := convertToBigIntType[T](input); ok {
			return transform(converted)
		}
		return input
	}, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline to another schema.
func (z *ZodBigInt[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	targetFn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractBigInt(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies a type-safe custom validation function.
func (z *ZodBigInt[T]) Refine(fn func(T) bool, params ...any) *ZodBigInt[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case *big.Int:
			if v == nil {
				return false
			}
			if val, ok := v.(*big.Int); ok {
				return fn(any(val).(T))
			}
			return false
		case **big.Int:
			if v == nil {
				return fn(any((**big.Int)(nil)).(T))
			}
			if val, ok := v.(*big.Int); ok {
				ptr := &val
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// RefineAny applies a custom validation function that receives the raw value.
func (z *ZodBigInt[T]) RefineAny(fn func(any) bool, params ...any) *ZodBigInt[T] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodBigInt[T]) withCheck(check core.ZodCheck) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// parseNilInput handles nil input by checking modifiers in priority order.
// Returns (result, _, true, err) when the result is final (default, optional, or error).
// Returns (_, substituteInput, false, nil) when a prefault value should be parsed.
func (z *ZodBigInt[T]) parseNilInput(ctx ...*core.ParseContext) (T, any, bool, error) {
	var zero T
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	in := &z.internals.ZodTypeInternals

	if in.NonOptional {
		return zero, nil, true, issues.CreateNonOptionalError(parseCtx)
	}

	// Default short-circuits: return value without validation.
	if in.DefaultValue != nil {
		v, err := engine.ConvertToConstraintType[*big.Int, T](in.DefaultValue, parseCtx, core.ZodTypeBigInt)
		return v, nil, true, err
	}
	if in.DefaultFunc != nil {
		v, err := engine.ConvertToConstraintType[*big.Int, T](in.DefaultFunc(), parseCtx, core.ZodTypeBigInt)
		return v, nil, true, err
	}

	// Prefault substitutes input and continues through validation.
	switch {
	case in.PrefaultValue != nil:
		return zero, in.PrefaultValue, false, nil
	case in.PrefaultFunc != nil:
		return zero, in.PrefaultFunc(), false, nil
	case in.Optional || in.Nilable:
		v, err := engine.ConvertToConstraintType[*big.Int, T](nil, parseCtx, core.ZodTypeBigInt)
		return v, nil, true, err
	default:
		return zero, nil, true, issues.CreateInvalidTypeError(core.ZodTypeBigInt, nil, parseCtx)
	}
}

// convertToBigIntType converts any value to the specified BigInt constraint type.
func convertToBigIntType[T BigIntegerConstraint](v any) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case *big.Int:
		if val, ok := v.(*big.Int); ok {
			return any(val).(T), true
		}
		return zero, false
	case **big.Int:
		if val, ok := v.(*big.Int); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if pp, ok := v.(**big.Int); ok {
			return any(pp).(T), true
		}
		return zero, false
	default:
		return zero, false
	}
}

// withPtrInternals creates a new **big.Int schema from cloned internals.
func (z *ZodBigInt[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodBigInt[**big.Int] {
	return &ZodBigInt[**big.Int]{internals: &ZodBigIntInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new schema preserving generic type T.
func (z *ZodBigInt[T]) withInternals(in *core.ZodTypeInternals) *ZodBigInt[T] {
	return &ZodBigInt[T]{internals: &ZodBigIntInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodBigInt[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodBigInt[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractBigInt extracts the underlying *big.Int from a generic constraint type.
func extractBigInt[T BigIntegerConstraint](value T) *big.Int {
	if ptr, ok := any(value).(**big.Int); ok {
		if ptr != nil {
			return *ptr
		}
		return nil
	}
	return any(value).(*big.Int)
}

// newZodBigIntFromDef constructs a new ZodBigInt from a definition.
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

// BigInt creates a *big.Int validation schema.
func BigInt(params ...any) *ZodBigInt[*big.Int] {
	return BigIntTyped[*big.Int](params...)
}

// BigIntPtr creates a **big.Int validation schema.
func BigIntPtr(params ...any) *ZodBigInt[**big.Int] {
	return BigIntTyped[**big.Int](params...)
}

// BigIntTyped is the generic constructor for big.Int schemas.
func BigIntTyped[T BigIntegerConstraint](params ...any) *ZodBigInt[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodBigIntDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeBigInt,
			Checks: []core.ZodCheck{},
		},
	}

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodBigIntFromDef[T](def)
}

// CoercedBigInt creates a coerced *big.Int schema that converts input types.
func CoercedBigInt(params ...any) *ZodBigInt[*big.Int] {
	schema := BigInt(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedBigIntPtr creates a coerced **big.Int schema that converts input types.
func CoercedBigIntPtr(params ...any) *ZodBigInt[**big.Int] {
	schema := BigIntPtr(params...)
	schema.internals.SetCoerce(true)
	return schema
}
