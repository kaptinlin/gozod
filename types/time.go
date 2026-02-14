package types

import (
	"time"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// Type Definitions
// =============================================================================

// TimeConstraint restricts values to time.Time or *time.Time.
type TimeConstraint interface {
	time.Time | *time.Time
}

// ZodTimeDef holds the configuration for a time schema.
type ZodTimeDef struct {
	core.ZodTypeDef
}

// ZodTimeInternals holds time validator internal state.
type ZodTimeInternals struct {
	core.ZodTypeInternals
	Def *ZodTimeDef
}

// ZodTime is a type-safe time validation schema.
type ZodTime[T TimeConstraint] struct {
	internals *ZodTimeInternals
}

// =============================================================================
// Core Methods
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodTime[T]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodTime[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodTime[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce converts input to time.Time.
func (z *ZodTime[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToTime(input)
	return result, err == nil
}

// =============================================================================
// Parsing Methods
// =============================================================================

// Parse validates input and returns a value of type T.
func (z *ZodTime[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[time.Time, T](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		engine.ApplyChecks[time.Time],
		engine.ConvertToConstraintType[time.Time, T],
		ctx...,
	)
}

// MustParse validates input and panics on failure.
func (z *ZodTime[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse requires exact type T for compile-time type safety.
func (z *ZodTime[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[time.Time, T](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		engine.ApplyChecks[time.Time],
		ctx...,
	)
}

// MustStrictParse requires exact type T and panics on failure.
func (z *ZodTime[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any type for runtime interface usage.
func (z *ZodTime[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// Modifier Methods
// =============================================================================

// Optional returns a new schema that accepts nil, with *time.Time constraint.
func (z *ZodTime[T]) Optional() *ZodTime[*time.Time] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
// Unlike Optional, ExactOptional only accepts absent keys in object fields.
func (z *ZodTime[T]) ExactOptional() *ZodTime[T] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a new schema that accepts nil values, with *time.Time constraint.
func (z *ZodTime[T]) Nilable() *ZodTime[*time.Time] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a new schema combining optional and nilable modifiers.
func (z *ZodTime[T]) Nullish() *ZodTime[*time.Time] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag, returning a time.Time constraint.
func (z *ZodTime[T]) NonOptional() *ZodTime[time.Time] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodTime[time.Time]{
		internals: &ZodTimeInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// =============================================================================
// Default and Prefault Methods
// =============================================================================

// Default sets a fallback value returned when input is nil (short-circuits validation).
func (z *ZodTime[T]) Default(v time.Time) *ZodTime[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits validation).
func (z *ZodTime[T]) DefaultFunc(fn func() time.Time) *ZodTime[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodTime[T]) Prefault(v time.Time) *ZodTime[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodTime[T]) PrefaultFunc(fn func() time.Time) *ZodTime[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// Metadata Methods
// =============================================================================

// Meta stores metadata for this time schema in the global registry.
func (z *ZodTime[T]) Meta(meta core.GlobalMeta) *ZodTime[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodTime[T]) Describe(desc string) *ZodTime[T] {
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

// =============================================================================
// Transformation and Validation Methods
// =============================================================================

// Transform applies a transformation function to the parsed value.
func (z *ZodTime[T]) Transform(fn func(time.Time, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(extractTime(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodTime[T]) Overwrite(transform func(T) T, params ...any) *ZodTime[T] {
	transformAny := func(input any) any {
		converted, ok := convertToTimeType[T](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	check := checks.NewZodCheckOverwrite(transformAny, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodTime[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	targetFn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractTime(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, targetFn)
}

// Refine applies a custom validation function matching the schema's output type T.
func (z *ZodTime[T]) Refine(fn func(T) bool, params ...any) *ZodTime[T] {
	wrapper := func(v any) bool {
		var zero T
		switch any(zero).(type) {
		case time.Time:
			if v == nil {
				return false
			}
			if timeVal, ok := v.(time.Time); ok {
				return fn(any(timeVal).(T))
			}
			return false
		case *time.Time:
			if v == nil {
				return fn(any((*time.Time)(nil)).(T))
			}
			if timeVal, ok := v.(time.Time); ok {
				return fn(any(&timeVal).(T))
			}
			return false
		default:
			return false
		}
	}

	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}

	check := checks.NewCustom[any](wrapper, msg)
	return z.withCheck(check)
}

// RefineAny applies a custom validation function that receives the raw value.
func (z *ZodTime[T]) RefineAny(fn func(any) bool, params ...any) *ZodTime[T] {
	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}
	check := checks.NewCustom[any](fn, msg)
	return z.withCheck(check)
}

// Check applies a custom validation function with full payload access (Zod v4 API).
func (z *ZodTime[T]) Check(fn func(value T, payload *core.ParsePayload), params ...any) *ZodTime[T] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(T); ok {
			fn(val, payload)
			return
		}

		var zero T
		if _, ok := any(zero).(*time.Time); ok {
			if t, ok := payload.Value().(time.Time); ok {
				fn(any(&t).(T), payload)
			}
		}
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// With is an alias for Check (Zod v4 API compatibility).
func (z *ZodTime[T]) With(fn func(value T, payload *core.ParsePayload), params ...any) *ZodTime[T] {
	return z.Check(fn, params...)
}

// =============================================================================
// Composition Methods
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodTime[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodTime[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// Helper Functions
// =============================================================================

// convertToTimeType converts only time values to the target time type T.
func convertToTimeType[T TimeConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		switch any(zero).(type) {
		case *time.Time:
			return zero, true
		default:
			return zero, false
		}
	}

	var timeValue time.Time
	var isValid bool

	switch val := v.(type) {
	case time.Time:
		timeValue, isValid = val, true
	case *time.Time:
		if val != nil {
			timeValue, isValid = *val, true
		}
	default:
		return zero, false
	}

	if !isValid {
		return zero, false
	}

	switch any(zero).(type) {
	case time.Time:
		return any(timeValue).(T), true
	case *time.Time:
		return any(new(timeValue)).(T), true
	default:
		return zero, false
	}
}

// =============================================================================
// Internal Helper Methods
// =============================================================================

// expectedType returns the schema's type code, defaulting to ZodTypeTime.
func (z *ZodTime[T]) expectedType() core.ZodTypeCode {
	if z.internals.Type != "" {
		return z.internals.Type
	}
	return core.ZodTypeTime
}

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodTime[T]) withCheck(check core.ZodCheck) *ZodTime[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withPtrInternals creates a new *time.Time schema from cloned internals.
func (z *ZodTime[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodTime[*time.Time] {
	return &ZodTime[*time.Time]{
		internals: &ZodTimeInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// withInternals creates a new schema preserving generic type T.
func (z *ZodTime[T]) withInternals(in *core.ZodTypeInternals) *ZodTime[T] {
	return &ZodTime[T]{
		internals: &ZodTimeInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodTime[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodTime[T]); ok {
		orig := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = orig
	}
}

// extractTime extracts the underlying time.Time from a generic constraint type.
func extractTime[T TimeConstraint](value T) time.Time {
	if ptr, ok := any(value).(*time.Time); ok {
		if ptr != nil {
			return *ptr
		}
		return time.Time{}
	}
	return any(value).(time.Time)
}

// =============================================================================
// Constructor Functions
// =============================================================================

// newZodTimeFromDef constructs a new ZodTime from a definition.
func newZodTimeFromDef[T TimeConstraint](def *ZodTimeDef) *ZodTime[T] {
	internals := &ZodTimeInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Provide a constructor so that AddCheck can create new schema instances.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		timeDef := &ZodTimeDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodTimeFromDef[T](timeDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodTime[T]{internals: internals}
}

// =============================================================================
// Public API Functions
// =============================================================================

// Time creates a time.Time validation schema.
//
//	Time()                    // no parameters
//	Time("custom error")      // string shorthand
//	Time(SchemaParams{...})   // full parameters
func Time(params ...any) *ZodTime[time.Time] {
	return TimeTyped[time.Time](params...)
}

// TimePtr creates a *time.Time validation schema.
func TimePtr(params ...any) *ZodTime[*time.Time] {
	return TimeTyped[*time.Time](params...)
}

// TimeTyped creates a time schema with explicit type parameterization.
func TimeTyped[T TimeConstraint](params ...any) *ZodTime[T] {
	sp := utils.NormalizeParams(params...)

	def := &ZodTimeDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeTime,
			Checks: []core.ZodCheck{},
		},
	}

	utils.ApplySchemaParams(&def.ZodTypeDef, sp)

	return newZodTimeFromDef[T](def)
}

// CoercedTime creates a time.Time schema with coercion enabled.
func CoercedTime(args ...any) *ZodTime[time.Time] {
	schema := Time(args...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedTimePtr creates a *time.Time schema with coercion enabled.
func CoercedTimePtr(args ...any) *ZodTime[*time.Time] {
	schema := TimePtr(args...)
	schema.internals.SetCoerce(true)
	return schema
}
