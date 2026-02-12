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

// BigIntegerConstraint restricts values to *big.Int or **big.Int.
type BigIntegerConstraint interface {
	*big.Int | **big.Int
}

// ZodBigIntDef holds the configuration for big.Int validation.
type ZodBigIntDef struct {
	core.ZodTypeDef
}

// ZodBigIntInternals contains the internal state of a big.Int validator.
type ZodBigIntInternals struct {
	core.ZodTypeInternals
	Def *ZodBigIntDef
}

// ZodBigInt is a type-safe big.Int validation schema.
type ZodBigInt[T BigIntegerConstraint] struct {
	internals *ZodBigIntInternals
}

// Internals returns the internal state of the schema.
func (z *ZodBigInt[T]) Internals() *core.ZodTypeInternals {
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

// Coerce converts input to *big.Int.
func (z *ZodBigInt[T]) Coerce(input any) (any, bool) {
	r, err := coerce.ToBigInt(input)
	return r, err == nil
}

// Parse validates input and returns a value matching the generic type T.
func (z *ZodBigInt[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	if input == nil {
		r, sub, done, err := z.parseNilInput(ctx...)
		if done {
			return r, err
		}
		input = sub
	}
	return engine.ParsePrimitive[*big.Int, T](input, &z.internals.ZodTypeInternals, core.ZodTypeBigInt, engine.ApplyChecks[*big.Int], engine.ConvertToConstraintType[*big.Int, T], ctx...)
}

// MustParse panics on validation failure.
func (z *ZodBigInt[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	r, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseAny validates input and returns an untyped result.
func (z *ZodBigInt[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse validates input with compile-time type safety.
func (z *ZodBigInt[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[*big.Int, T](input, &z.internals.ZodTypeInternals, core.ZodTypeBigInt, engine.ApplyChecks[*big.Int], ctx...)
}

// MustStrictParse panics on validation failure with compile-time type safety.
func (z *ZodBigInt[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodBigInt[T]) Optional() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodBigInt[T]) Nilable() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodBigInt[T]) Nullish() *ZodBigInt[**big.Int] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag, returning a *big.Int constraint.
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

// Default sets a fallback value returned when input is nil (short-circuits).
func (z *ZodBigInt[T]) Default(v *big.Int) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits).
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
func (z *ZodBigInt[T]) Describe(desc string) *ZodBigInt[T] {
	in := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = desc
	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// Min adds minimum value validation (>=).
func (z *ZodBigInt[T]) Min(n *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Gte(n, params...))
}

// Max adds maximum value validation (<=).
func (z *ZodBigInt[T]) Max(n *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Lte(n, params...))
}

// Gt adds greater-than validation.
func (z *ZodBigInt[T]) Gt(n *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Gt(n, params...))
}

// Gte adds greater-than-or-equal validation.
func (z *ZodBigInt[T]) Gte(n *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Gte(n, params...))
}

// Lt adds less-than validation.
func (z *ZodBigInt[T]) Lt(n *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Lt(n, params...))
}

// Lte adds less-than-or-equal validation.
func (z *ZodBigInt[T]) Lte(n *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.Lte(n, params...))
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
func (z *ZodBigInt[T]) MultipleOf(n *big.Int, params ...any) *ZodBigInt[T] {
	return z.withCheck(checks.MultipleOf(n, params...))
}

// Transform creates a type-safe transformation pipeline.
func (z *ZodBigInt[T]) Transform(fn func(*big.Int, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapper := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(extractBigInt(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapper)
}

// Overwrite transforms the value while preserving the same type.
func (z *ZodBigInt[T]) Overwrite(transform func(T) T, params ...any) *ZodBigInt[T] {
	check := checks.NewZodCheckOverwrite(func(input any) any {
		if v, ok := toBigIntType[T](input); ok {
			return transform(v)
		}
		return input
	}, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline to another schema.
func (z *ZodBigInt[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	fn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractBigInt(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, fn)
}

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

// RefineAny applies a custom validation function on the raw value.
func (z *ZodBigInt[T]) RefineAny(fn func(any) bool, params ...any) *ZodBigInt[T] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// withCheck clones internals, adds a check, and returns a new schema.
func (z *ZodBigInt[T]) withCheck(c core.ZodCheck) *ZodBigInt[T] {
	in := z.internals.Clone()
	in.AddCheck(c)
	return z.withInternals(in)
}

// parseNilInput handles nil input by checking modifiers in priority order.
// Returns (result, _, true, err) when final, or (_, substitute, false, nil) for prefault.
func (z *ZodBigInt[T]) parseNilInput(ctx ...*core.ParseContext) (T, any, bool, error) {
	var zero T
	pctx := &core.ParseContext{}
	if len(ctx) > 0 {
		pctx = ctx[0]
	}

	ti := &z.internals.ZodTypeInternals

	if ti.NonOptional {
		return zero, nil, true, issues.CreateNonOptionalError(pctx)
	}
	if ti.DefaultValue != nil {
		v, err := engine.ConvertToConstraintType[*big.Int, T](ti.DefaultValue, pctx, core.ZodTypeBigInt)
		return v, nil, true, err
	}
	if ti.DefaultFunc != nil {
		v, err := engine.ConvertToConstraintType[*big.Int, T](ti.DefaultFunc(), pctx, core.ZodTypeBigInt)
		return v, nil, true, err
	}

	switch {
	case ti.PrefaultValue != nil:
		return zero, ti.PrefaultValue, false, nil
	case ti.PrefaultFunc != nil:
		return zero, ti.PrefaultFunc(), false, nil
	case ti.Optional || ti.Nilable:
		v, err := engine.ConvertToConstraintType[*big.Int, T](nil, pctx, core.ZodTypeBigInt)
		return v, nil, true, err
	default:
		return zero, nil, true, issues.CreateInvalidTypeError(core.ZodTypeBigInt, nil, pctx)
	}
}

// toBigIntType converts any value to the specified BigInt constraint type.
func toBigIntType[T BigIntegerConstraint](v any) (T, bool) {
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

// withPtrInternals creates a **big.Int schema from cloned internals.
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
func (z *ZodBigInt[T]) CloneFrom(src any) {
	if s, ok := src.(*ZodBigInt[T]); ok {
		prev := z.internals.Checks
		*z.internals = *s.internals
		z.internals.Checks = prev
	}
}

// extractBigInt extracts the underlying *big.Int from a constraint type.
func extractBigInt[T BigIntegerConstraint](v T) *big.Int {
	if ptr, ok := any(v).(**big.Int); ok {
		if ptr != nil {
			return *ptr
		}
		return nil
	}
	return any(v).(*big.Int)
}

// newZodBigIntFromDef constructs a ZodBigInt from a definition.
func newZodBigIntFromDef[T BigIntegerConstraint](def *ZodBigIntDef) *ZodBigInt[T] {
	in := &ZodBigIntInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		return any(newZodBigIntFromDef[T](&ZodBigIntDef{ZodTypeDef: *d})).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodBigInt[T]{internals: in}
}

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
	sp := utils.NormalizeParams(params...)
	def := &ZodBigIntDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeBigInt,
			Checks: []core.ZodCheck{},
		},
	}
	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}
	return newZodBigIntFromDef[T](def)
}

// CoercedBigInt creates a coerced *big.Int schema.
func CoercedBigInt(params ...any) *ZodBigInt[*big.Int] {
	s := BigInt(params...)
	s.internals.SetCoerce(true)
	return s
}

// CoercedBigIntPtr creates a coerced **big.Int schema.
func CoercedBigIntPtr(params ...any) *ZodBigInt[**big.Int] {
	s := BigIntPtr(params...)
	s.internals.SetCoerce(true)
	return s
}
