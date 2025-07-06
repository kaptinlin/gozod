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

// NeverConstraint defines the types that can be used with ZodNever
type NeverConstraint interface {
	any
}

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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodNever[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse validates the input value with never logic (always fails except special cases)
func (z *ZodNever[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}
	if parseCtx == nil {
		parseCtx = core.NewParseContext()
	}

	// Handle nil input first
	if input == nil {
		// Try default value first
		if z.internals.ZodTypeInternals.DefaultValue != nil {
			return convertToNeverConstraintType[T, R](any(z.internals.ZodTypeInternals.DefaultValue).(T)), nil //nolint:unconvert
		}

		// Try default function
		if z.internals.ZodTypeInternals.DefaultFunc != nil {
			defaultValue := z.internals.ZodTypeInternals.DefaultFunc()
			return convertToNeverConstraintType[T, R](any(defaultValue).(T)), nil //nolint:unconvert
		}

		// Check if nil is allowed (optional/nilable) - after default handling
		if z.internals.ZodTypeInternals.Optional || z.internals.ZodTypeInternals.Nilable {
			var zero R
			return zero, nil
		}

		// For Never type, nil is rejected like any other value, so fall through to prefault handling
	}

	// Never type core logic: always fail for non-nil inputs
	// Try prefault fallback mechanism after primary validation fails
	return z.tryPrefaultFallback(input, parseCtx)
}

// tryPrefaultFallback attempts to use prefault values when validation fails
func (z *ZodNever[T, R]) tryPrefaultFallback(originalInput any, ctx *core.ParseContext) (R, error) {
	internals := &z.internals.ZodTypeInternals

	// Try prefault value first
	if internals.PrefaultValue != nil {
		prefaultValue := internals.PrefaultValue

		if len(internals.Checks) > 0 {
			validated, err := engine.ApplyChecks[any](prefaultValue, internals.Checks, ctx)
			if err != nil {
				var zero R
				return zero, err
			}
			prefaultValue = validated
		}

		if converted, ok := convertToNeverConstraintValue[T, R](prefaultValue); ok {
			return converted, nil
		}

		var zero R
		rawIssue := issues.CreateInvalidTypeIssue(core.ZodTypeNever, prefaultValue)
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, core.GetConfig())
		return zero, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	if internals.PrefaultFunc != nil {
		prefaultValue := internals.PrefaultFunc()

		if len(internals.Checks) > 0 {
			validated, err := engine.ApplyChecks[any](prefaultValue, internals.Checks, ctx)
			if err != nil {
				var zero R
				return zero, err
			}
			prefaultValue = validated
		}

		if converted, ok := convertToNeverConstraintValue[T, R](prefaultValue); ok {
			return converted, nil
		}

		var zero R
		rawIssue := issues.CreateInvalidTypeIssue(core.ZodTypeNever, prefaultValue)
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, core.GetConfig())
		return zero, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// No prefault available, create and return error using unified error handling
	var zero R
	rawIssue := issues.CreateInvalidTypeIssue(core.ZodTypeNever, originalInput)
	rawIssue.Message = "Never type should not accept any value"
	rawIssue.Inst = z // Pass schema instance for custom error message extraction
	finalIssue := issues.FinalizeIssue(rawIssue, ctx, core.GetConfig())
	return zero, issues.NewZodError([]core.ZodIssue{finalIssue})
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

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the never optional (allows nil) - returns pointer constraint
func (z *ZodNever[T, R]) Optional() *ZodNever[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable makes the never nilable (allows nil) - returns pointer constraint
func (z *ZodNever[T, R]) Nilable() *ZodNever[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish makes the never both optional and nilable - returns pointer constraint
func (z *ZodNever[T, R]) Nullish() *ZodNever[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value - preserves current constraint type
func (z *ZodNever[T, R]) Default(value T) *ZodNever[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(value)
	return z.withInternals(in)
}

// DefaultFunc sets a default value function - preserves current constraint type
func (z *ZodNever[T, R]) DefaultFunc(fn func() T) *ZodNever[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a prefault (fallback) value - preserves current constraint type
func (z *ZodNever[T, R]) Prefault(value T) *ZodNever[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(value)
	return z.withInternals(in)
}

// PrefaultFunc sets a prefault function - preserves current constraint type
func (z *ZodNever[T, R]) PrefaultFunc(fn func() T) *ZodNever[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
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
	return core.NewZodPipe[R, any](z, wrapperFn)
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
	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible validation with any type
func (z *ZodNever[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodNever[T, R] {
	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
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

// convertToNeverConstraintType converts a base type T to constraint type R
func convertToNeverConstraintType[T any, R any](value T) R {
	var zero R
	zeroType := reflect.TypeOf(zero)

	// Handle pointer types using reflection
	if zeroType != nil && zeroType.Kind() == reflect.Ptr {
		if any(value) == nil {
			// Return nil pointer
			return reflect.Zero(zeroType).Interface().(R)
		}

		// Create pointer to value
		valuePtr := reflect.New(zeroType.Elem())
		valuePtr.Elem().Set(reflect.ValueOf(value))
		return valuePtr.Interface().(R)
	}

	// For non-pointer constraint types, return value directly
	return any(value).(R)
}

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
		return any(value).(T) //nolint:unconvert
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
			Type:   def.ZodTypeDef.Type,
			Checks: def.ZodTypeDef.Checks,
			Coerce: def.ZodTypeDef.Coerce,
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

	if def.ZodTypeDef.Error != nil {
		internals.Error = def.ZodTypeDef.Error
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
	param := utils.GetFirstParam(paramArgs...)
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
