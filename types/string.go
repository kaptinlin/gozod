package types

import (
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// StringConstraint restricts values to string or *string.
type StringConstraint interface {
	string | *string
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodStringDef defines the configuration for string validation
type ZodStringDef struct {
	core.ZodTypeDef
}

// ZodStringInternals contains string validator internal state
type ZodStringInternals struct {
	core.ZodTypeInternals
	Def *ZodStringDef // Schema definition
}

// ZodString represents a string validation schema with type safety
type ZodString[T StringConstraint] struct {
	internals *ZodStringInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodString[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodString[T]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodString[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Coerce implements Coercible interface for string type conversion
func (z *ZodString[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse returns a value that matches the generic type T with full type safety.
func (z *ZodString[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	// Use the internally recorded type code by default, fall back to string if not set
	expectedType := z.internals.Type
	if expectedType == "" {
		expectedType = core.ZodTypeString
	}

	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		expectedType,
		engine.ApplyChecks[string],
		engine.ConvertToConstraintType[string, T],
		ctx...,
	)
}

// MustParse is the type-safe variant that panics on error.
func (z *ZodString[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodString[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Use the internally recorded type code by default, fall back to string if not set
	expectedType := z.internals.Type
	if expectedType == "" {
		expectedType = core.ZodTypeString
	}

	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		expectedType,
		engine.ApplyChecks[string],
		ctx...,
	)
}

// MustStrictParse is the strict mode variant that panics on error.
// Provides compile-time type safety with maximum performance.
func (z *ZodString[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// MustParseAny validates the input value and panics on failure
func (z *ZodString[T]) MustParseAny(input any, ctx ...*core.ParseContext) any {
	result, err := z.ParseAny(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodString[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *string because the optional value may be nil.
func (z *ZodString[T]) Optional() *ZodString[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable always returns *string because the value may be nil.
func (z *ZodString[T]) Nilable() *ZodString[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility
func (z *ZodString[T]) Nullish() *ZodString[*string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default keeps the current generic type T.
func (z *ZodString[T]) Default(v string) *ZodString[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic type T.
func (z *ZodString[T]) DefaultFunc(fn func() string) *ZodString[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault keeps the current generic type T.
func (z *ZodString[T]) Prefault(v string) *ZodString[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current generic type T.
func (z *ZodString[T]) PrefaultFunc(fn func() string) *ZodString[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this schema in the global registry.
// This does not clone internals because metadata does not affect validation.
func (z *ZodString[T]) Meta(meta core.GlobalMeta) *ZodString[T] {
	// Create a shallow clone so that metadata can differ per usage (parity with TS Zod .describe()).
	clone := z.withInternals(&z.internals.ZodTypeInternals)

	// Propagate existing metadata from the source (if any) so we don't lose previously set fields.
	if m, ok := core.GlobalRegistry.Get(z); ok {
		combined := m
		if meta.ID != "" {
			combined.ID = meta.ID
		}
		if meta.Title != "" {
			combined.Title = meta.Title
		}
		if meta.Description != "" {
			combined.Description = meta.Description
		}
		if len(meta.Examples) > 0 {
			combined.Examples = meta.Examples
		}
		core.GlobalRegistry.Add(clone, combined)
	} else {
		core.GlobalRegistry.Add(clone, meta)
	}

	return clone
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min adds minimum length validation
func (z *ZodString[T]) Min(minLen int, params ...any) *ZodString[T] {
	check := checks.MinLength(minLen, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max adds maximum length validation
func (z *ZodString[T]) Max(maxLen int, params ...any) *ZodString[T] {
	check := checks.MaxLength(maxLen, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Length adds exact length validation
func (z *ZodString[T]) Length(length int, params ...any) *ZodString[T] {
	check := checks.Length(length, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Regex adds custom regex validation
func (z *ZodString[T]) Regex(pattern *regexp.Regexp, params ...any) *ZodString[T] {
	check := checks.Regex(pattern, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RegexString adds custom regex validation using string pattern (convenience method)
func (z *ZodString[T]) RegexString(pattern string, params ...any) *ZodString[T] {
	compiled := regexp.MustCompile(pattern)
	return z.Regex(compiled, params...)
}

// StartsWith adds prefix validation
func (z *ZodString[T]) StartsWith(prefix string, params ...any) *ZodString[T] {
	check := checks.StartsWith(prefix, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// EndsWith adds suffix validation
func (z *ZodString[T]) EndsWith(suffix string, params ...any) *ZodString[T] {
	check := checks.EndsWith(suffix, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Includes adds substring validation
func (z *ZodString[T]) Includes(substring string, params ...any) *ZodString[T] {
	check := checks.Includes(substring, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Trim adds string trimming transformation
func (z *ZodString[T]) Trim(params ...any) *ZodString[T] {
	// Use Overwrite to transform the value by trimming surrounding whitespace.
	transform := func(val T) T {
		switch v := any(val).(type) {
		case string:
			return any(strings.TrimSpace(v)).(T)
		case *string:
			if v == nil {
				return val // keep nil pointer untouched
			}
			trimmed := strings.TrimSpace(*v)
			ptr := &trimmed
			return any(ptr).(T)
		default:
			return val
		}
	}
	return z.Overwrite(transform, params...)
}

// JSON adds JSON format validation
func (z *ZodString[T]) JSON(params ...any) *ZodString[T] {
	check := checks.JSON(params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Email adds email format validation
func (z *ZodString[T]) Email(params ...any) *ZodString[T] {
	check := checks.Email(params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TRANSFORMATION METHODS
// =============================================================================

// Transform applies a transformation function to the validated string
func (z *ZodString[T]) Transform(fn func(string, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	return core.NewZodTransform(z, func(input T, ctx *core.RefinementContext) (any, error) {
		str := extractString(input)
		return fn(str, ctx)
	})
}

// Overwrite applies a transformation function that must return the same type T
func (z *ZodString[T]) Overwrite(transform func(T) T, params ...any) *ZodString[T] {
	// Create a transformation function that works with the exact type T
	transformAny := func(input any) any {
		// Try to convert input to type T
		converted, ok := convertToStringType[T](input)
		if !ok {
			// If conversion fails, return original value unchanged
			return input
		}
		// Apply transformation directly on type T
		return transform(converted)
	}

	check := checks.NewZodCheckOverwrite(transformAny, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline with another schema
func (z *ZodString[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	return core.NewZodPipe(z, target, func(input T, ctx *core.ParseContext) (any, error) {
		str := extractString(input)
		return target.Parse(str, ctx)
	})
}

// Check adds a custom validation function that can push multiple issues via ParsePayload.
func (z *ZodString[T]) Check(fn func(value T, payload *core.ParsePayload), params ...any) *ZodString[T] {
	// Wrap the user callback to support both value and pointer generic forms transparently.
	wrapped := func(payload *core.ParsePayload) {
		// First attempt: direct type assertion to the generic type T.
		if val, ok := payload.GetValue().(T); ok {
			fn(val, payload)
			return
		}

		// Special handling when T is a pointer type (*string) but the underlying value is its base type (string).
		var zero T
		// Use type switch on the zero value's dynamic type to detect pointer scenarios without reflection overhead.
		switch any(zero).(type) {
		case *string:
			if strVal, ok := payload.GetValue().(string); ok {
				strCopy := strVal // Create a new copy to take address safely
				ptr := &strCopy
				fn(any(ptr).(T), payload)
			}
		// Additional pointer specialisations can be added here if required in the future.
		default:
			// No convertible path found – do nothing.
		}
	}

	check := checks.NewCustom[T](wrapped, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies a custom validation function that matches the schema's output type T.
func (z *ZodString[T]) Refine(fn func(T) bool, params ...any) *ZodString[T] {
	// Wrapper converts the raw value (always string or nil) into T before calling fn.
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case string:
			// Schema output is string
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			// Schema output is *string – convert incoming value (string or nil) to *string
			if v == nil {
				return true // Allow nil to pass refinement for nilable types
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				ptr := &sCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodString[T]) RefineAny(fn func(any) bool, params ...any) *ZodString[T] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// INTERNAL HELPER METHODS
// =============================================================================

// withPtrInternals creates a new ZodString instance of type *string.
func (z *ZodString[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodString[*string] {
	clone := &ZodString[*string]{
		internals: &ZodStringInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
	if meta, ok := core.GlobalRegistry.Get(z); ok {
		core.GlobalRegistry.Add(clone, meta)
	}
	return clone
}

// withInternals creates a new ZodString instance that keeps the original generic type T.
func (z *ZodString[T]) withInternals(in *core.ZodTypeInternals) *ZodString[T] {
	clone := &ZodString[T]{
		internals: &ZodStringInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
	if meta, ok := core.GlobalRegistry.Get(z); ok {
		core.GlobalRegistry.Add(clone, meta)
	}
	return clone
}

// CloneFrom copies the internal state from another schema
func (z *ZodString[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodString[T]); ok && src != nil {
		z.internals = &ZodStringInternals{
			ZodTypeInternals: *src.internals.ZodTypeInternals.Clone(),
			Def:              src.internals.Def,
		}
	}
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// convertToStringType converts any value to the constrained string type T
func convertToStringType[T StringConstraint](v any) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case string:
		if str, ok := v.(string); ok {
			return any(str).(T), true
		}
	case *string:
		if v == nil {
			return any((*string)(nil)).(T), true
		}
		if str, ok := v.(string); ok {
			sCopy := str
			return any(&sCopy).(T), true
		}
		if strPtr, ok := v.(*string); ok {
			return any(strPtr).(T), true
		}
	}

	return zero, false
}

// extractString extracts the string value from a StringConstraint type
func extractString[T StringConstraint](value T) string {
	switch v := any(value).(type) {
	case string:
		return v
	case *string:
		if v == nil {
			return ""
		}
		return *v
	default:
		return ""
	}
}

// newZodStringFromDef constructs a new ZodString from the given definition.
func newZodStringFromDef[T StringConstraint](def *ZodStringDef) *ZodString[T] {
	internals := &ZodStringInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   core.ZodTypeString,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Provide a constructor so that AddCheck can create new schema instances.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		stringDef := &ZodStringDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodStringFromDef[T](stringDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodString[T]{
		internals: internals,
	}
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// String creates a new string schema
func String(params ...any) *ZodString[string] {
	return StringTyped[string](params...)
}

// StringPtr creates a new string schema with pointer type
func StringPtr(params ...any) *ZodString[*string] {
	return StringTyped[*string](params...)
}

// StringTyped creates a new string schema with specific type
func StringTyped[T StringConstraint](params ...any) *ZodString[T] {
	schemaParams := utils.NormalizeParams(params...)
	def := &ZodStringDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:     core.ZodTypeString,
			Required: true,
			Checks:   []core.ZodCheck{},
		},
	}

	// Parse parameters for custom configuration
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodStringFromDef[T](def)
}

// CoercedString creates a new string schema with coercion enabled
func CoercedString(params ...any) *ZodString[string] {
	schema := StringTyped[string](params...)
	schema.internals.Coerce = true
	return schema
}

// CoercedStringPtr creates a new string schema with pointer type and coercion enabled
func CoercedStringPtr(params ...any) *ZodString[*string] {
	schema := StringTyped[*string](params...)
	schema.internals.Coerce = true
	return schema
}

// ToLowerCase transforms the string to lower case
func (z *ZodString[T]) ToLowerCase(params ...any) *ZodString[T] {
	transform := func(val T) T {
		switch v := any(val).(type) {
		case string:
			return any(strings.ToLower(v)).(T)
		case *string:
			if v == nil {
				return val
			}
			lower := strings.ToLower(*v)
			ptr := &lower
			return any(ptr).(T)
		default:
			return val
		}
	}
	return z.Overwrite(transform, params...)
}

// ToUpperCase transforms the string to upper case
func (z *ZodString[T]) ToUpperCase(params ...any) *ZodString[T] {
	transform := func(val T) T {
		switch v := any(val).(type) {
		case string:
			return any(strings.ToUpper(v)).(T)
		case *string:
			if v == nil {
				return val
			}
			upper := strings.ToUpper(*v)
			ptr := &upper
			return any(ptr).(T)
		default:
			return val
		}
	}
	return z.Overwrite(transform, params...)
}

// NonOptional removes the optional flag and returns a new schema with string value type
func (z *ZodString[T]) NonOptional() *ZodString[string] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodString[string]{
		internals: &ZodStringInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}
