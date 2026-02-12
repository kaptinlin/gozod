package types

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

//////////////////////////
// TYPE DEFINITIONS
//////////////////////////

// ZodNeverDef defines the structure for never validation schemas
type ZodNeverDef struct {
	core.ZodTypeDef
}

// ZodNeverInternals provides internal state for never schemas
type ZodNeverInternals struct {
	core.ZodTypeInternals
	Def *ZodNeverDef // Schema definition
}

// ZodNever represents a never validation schema that always fails validation
type ZodNever[T any, R any] struct {
	internals *ZodNeverInternals
}

//////////////////////////
// CORE METHODS
//////////////////////////

// GetInternals returns the internal configuration
func (z *ZodNever[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodNever[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodNever[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// validateNeverValue is the validator function for Never type
// Never type should not accept any value, but allows prefault fallback
func validateNeverValue[T any](value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	// Apply any additional checks first
	if len(checks) > 0 {
		validatedValue, err := engine.ApplyChecks[T](value, checks, ctx)
		if err != nil {
			var zero T
			return zero, err
		}
		value = validatedValue
	}

	// Never type rejects all values using standard invalid type error
	var zero T
	return zero, issues.CreateInvalidTypeError(core.ZodTypeNever, value, ctx)
}

// validateNeverValueWithPrefault handles Never type validation with prefault support
// For Never type, all inputs should be rejected with the custom message
func (z *ZodNever[T, R]) validateNeverValueWithPrefault(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	// Never type rejects all values, including prefault values
	return validateNeverValue[T](value, checks, ctx)
}

// Parse validates the input value with never logic using unified engine parsing
func (z *ZodNever[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	// Create a validator that can access internals for prefault handling
	validator := func(value T, checks []core.ZodCheck, parseCtx *core.ParseContext) (T, error) {
		return z.validateNeverValueWithPrefault(value, checks, parseCtx)
	}

	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeNever,
		validator,
		engine.ConvertToConstraintType[T, R],
		ctx...,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodNever[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodNever[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type R.
func (z *ZodNever[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeNever,
		validateNeverValue[T],
		ctx...,
	)
}

// MustStrictParse provides compile-time type safety and panics on validation failure.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type R.
func (z *ZodNever[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the never optional (allows nil) - returns pointer constraint
func (z *ZodNever[T, R]) Optional() *ZodNever[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable makes the never nilable (allows nil) - returns pointer constraint
func (z *ZodNever[T, R]) Nilable() *ZodNever[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish makes the never both optional and nilable - returns pointer constraint
func (z *ZodNever[T, R]) Nullish() *ZodNever[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value - preserves current constraint type
func (z *ZodNever[T, R]) Default(value T) *ZodNever[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(value)
	return z.withInternals(in)
}

// DefaultFunc sets a default value function - preserves current constraint type
func (z *ZodNever[T, R]) DefaultFunc(fn func() T) *ZodNever[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a prefault (fallback) value - preserves current constraint type
func (z *ZodNever[T, R]) Prefault(value T) *ZodNever[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(value)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current generic type R.
func (z *ZodNever[T, R]) PrefaultFunc(fn func() T) *ZodNever[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this never schema.
func (z *ZodNever[T, R]) Meta(meta core.GlobalMeta) *ZodNever[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodNever[T, R]) Describe(description string) *ZodNever[T, R] {
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

//////////////////////////
// TRANSFORMATION/PIPELINE
//////////////////////////

// Transform applies a transformation function using WrapFn pattern (note: rarely called due to Never's nature)
func (z *ZodNever[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractNeverValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a validation pipeline using WrapFn pattern (first stage will be Never)
func (z *ZodNever[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractNeverValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

//////////////////////////
// REFINEMENT
//////////////////////////

// Refine adds a validation function with constraint type R
func (z *ZodNever[T, R]) Refine(fn func(R) bool, params ...any) *ZodNever[T, R] {
	wrapper := func(v any) bool {
		// Convert to constraint type R and call function
		if converted, ok := convertToNeverConstraintValue[T, R](v); ok {
			return fn(converted)
		}
		return false
	}

	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	// Convert back to the format expected by checks.NewCustom
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible validation with any type
func (z *ZodNever[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodNever[T, R] {
	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

//////////////////////////
// HELPER METHODS
//////////////////////////

// withPtrInternals creates a new ZodNever instance with pointer constraint
func (z *ZodNever[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodNever[T, *T] {
	return &ZodNever[T, *T]{internals: &ZodNeverInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new ZodNever instance with updated internals
func (z *ZodNever[T, R]) withInternals(in *core.ZodTypeInternals) *ZodNever[T, R] {
	return &ZodNever[T, R]{internals: &ZodNeverInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodNever[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodNever[T, R]); ok {
		z.internals = src.internals
	}
}

// Unwrap returns the schema itself for compatibility
func (z *ZodNever[T, R]) Unwrap() *ZodNever[T, R] {
	return z
}

//////////////////////////
// TYPE CONVERSION HELPERS
//////////////////////////

// extractNeverValue extracts base type T from constraint type R
func extractNeverValue[T any, R any](value R) T {
	switch v := any(value).(type) {
	case *any:
		if v != nil {
			return any(*v).(T) //nolint:unconvert
		}
		var zero T
		return zero
	default:
		return any(value).(T)
	}
}

// convertToNeverConstraintValue converts any value to constraint type R if possible
func convertToNeverConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok { //nolint:unconvert
		return r, true
	}

	// Handle pointer conversion using reflection for any pointer type
	zeroType := reflect.TypeOf(zero)
	if zeroType != nil && zeroType.Kind() == reflect.Ptr {
		// R is a pointer type (*T, *string, *int, etc.)
		elemType := zeroType.Elem()

		// Try to convert value to the element type first
		if value != nil {
			valueType := reflect.TypeOf(value)
			if valueType != nil {
				// If value is already the element type, create pointer to it
				if valueType == elemType {
					ptrValue := reflect.New(elemType)
					ptrValue.Elem().Set(reflect.ValueOf(value))
					return ptrValue.Interface().(R), true
				}

				// If value is convertible to element type, convert then create pointer
				if valueType.ConvertibleTo(elemType) {
					converted := reflect.ValueOf(value).Convert(elemType)
					ptrValue := reflect.New(elemType)
					ptrValue.Elem().Set(converted)
					return ptrValue.Interface().(R), true
				}
			}
		}

		// Return nil pointer if value is nil or conversion failed
		nilPtr := reflect.Zero(zeroType)
		return nilPtr.Interface().(R), true
	}

	// Handle value types - try conversion
	if value != nil {
		valueType := reflect.TypeOf(value)
		zeroType := reflect.TypeOf(zero)
		if valueType != nil && zeroType != nil && valueType.ConvertibleTo(zeroType) {
			converted := reflect.ValueOf(value).Convert(zeroType)
			return converted.Interface().(R), true
		}
	}

	return zero, false
}

//////////////////////////
// FACTORY FUNCTIONS
//////////////////////////

// newZodNeverFromDef constructs new ZodNever from definition
func newZodNeverFromDef[T any, R any](def *ZodNeverDef) *ZodNever[T, R] {
	internals := &ZodNeverInternals{
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
		neverDef := &ZodNeverDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodNeverFromDef[T, R](neverDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodNever[T, R]{internals: internals}
}

// Never creates never schema that always fails validation - returns value constraint
func Never(paramArgs ...any) *ZodNever[any, any] {
	return NeverTyped[any, any](paramArgs...)
}

// NeverPtr creates never schema that always fails validation - returns pointer constraint
func NeverPtr(paramArgs ...any) *ZodNever[any, *any] {
	return NeverTyped[any, *any](paramArgs...)
}

// NeverTyped creates typed never schema with generic constraints
func NeverTyped[T any, R any](paramArgs ...any) *ZodNever[T, R] {
	param := utils.FirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodNeverDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeNever,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply normalized parameters to schema definition
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodNeverFromDef[T, R](def)
}
