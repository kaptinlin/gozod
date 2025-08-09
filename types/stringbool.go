package types

import (
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

// StringBoolConstraint restricts values to bool or *bool for StringBool output.
type StringBoolConstraint interface {
	bool | *bool
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// StringBoolOptions provides configuration for stringbool schema creation
type StringBoolOptions struct {
	Truthy []string // Values that evaluate to true
	Falsy  []string // Values that evaluate to false
	Case   string   // "sensitive" or "insensitive"
}

// ZodStringBoolDef defines the configuration for string boolean validation
type ZodStringBoolDef struct {
	core.ZodTypeDef
	Truthy        []string // Truthy string values
	Falsy         []string // Falsy string values
	Case          string   // "sensitive" or "insensitive"
	CustomOptions bool     // Whether custom options were provided
}

// ZodStringBoolInternals contains stringbool validator internal state
type ZodStringBoolInternals struct {
	core.ZodTypeInternals
	Def    *ZodStringBoolDef   // Schema definition
	Truthy map[string]struct{} // Truthy values set for fast lookup
	Falsy  map[string]struct{} // Falsy values set for fast lookup
}

// ZodStringBool represents a string-to-boolean validation schema with type safety
type ZodStringBool[T StringBoolConstraint] struct {
	internals *ZodStringBoolInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodStringBool[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodStringBool[T]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodStringBool[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Coerce implements Coercible interface for string-to-bool type conversion
func (z *ZodStringBool[T]) Coerce(input any) (any, bool) {
	// First try to coerce to string
	if str, err := coerce.ToString(input); err == nil {
		// Then try to convert string to bool using our custom logic
		if _, ok := z.tryStringToBool(str); ok {
			return str, true // Return the string, not the bool
		}
	}
	return input, false
}

// validateStringBoolValue is the validator function for StringBool type
func (z *ZodStringBool[T]) validateStringBoolValue(value bool, checks []core.ZodCheck, ctx *core.ParseContext) (bool, error) {
	// Apply checks and return the result
	return engine.ApplyChecks[bool](value, checks, ctx)
}

// Parse returns a value that matches the generic type T with full type safety.
func (z *ZodStringBool[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	result, err := engine.ParseComplex[bool](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeStringBool,
		z.extractStringBoolForEngine,
		z.extractStringBoolPtrForEngine,
		z.validateStringBoolForEngine,
		ctx...,
	)
	if err != nil {
		var zero T
		return zero, err
	}

	return engine.ConvertToConstraintType[bool, T](result, core.NewParseContext(), core.ZodTypeStringBool)
}

// MustParse is the type-safe variant that panics on error.
func (z *ZodStringBool[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodStringBool[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodStringBool[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Custom validator for StringBool that applies checks
	validator := func(value bool, checks []core.ZodCheck, ctx *core.ParseContext) (bool, error) {
		return z.validateStringBoolValue(value, checks, ctx)
	}

	return engine.ParsePrimitiveStrict[bool, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeStringBool,
		validator,
		ctx...,
	)
}

// MustStrictParse provides compile-time type safety and panics on validation failure.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodStringBool[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *bool for nullable semantics
func (z *ZodStringBool[T]) Optional() *ZodStringBool[*bool] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodStringBool[T]) Nilable() *ZodStringBool[*bool] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodStringBool[T]) Nullish() *ZodStringBool[*bool] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodStringBool[T]) Default(v bool) *ZodStringBool[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodStringBool[T]) DefaultFunc(fn func() bool) *ZodStringBool[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
// According to Zod v4 semantics, prefault accepts input type (string) for StringBool
func (z *ZodStringBool[T]) Prefault(v string) *ZodStringBool[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps current generic type T.
// According to Zod v4 semantics, prefault function returns input type (string) for StringBool
func (z *ZodStringBool[T]) PrefaultFunc(fn func() string) *ZodStringBool[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this stringbool schema.
func (z *ZodStringBool[T]) Meta(meta core.GlobalMeta) *ZodStringBool[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Refine adds type-safe custom validation logic to the stringbool schema
func (z *ZodStringBool[T]) Refine(fn func(T) bool, params ...any) *ZodStringBool[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case bool:
			// Schema output is bool
			if v == nil {
				return false // nil should never reach here for bool schema
			}
			if boolVal, ok := v.(bool); ok {
				return fn(any(boolVal).(T))
			}
			return false
		case *bool:
			// Schema output is *bool – convert incoming value to *bool
			if v == nil {
				return fn(any((*bool)(nil)).(T))
			}
			if boolVal, ok := v.(bool); ok {
				bCopy := boolVal
				ptr := &bCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false // Unsupported type
		}
	}

	// MUST use checks package for custom validation
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodStringBool[T]) RefineAny(fn func(any) bool, params ...any) *ZodStringBool[T] {
	// MUST use checks package for custom validation
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodStringBool[T]) Transform(fn func(bool, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		boolValue := extractStringBool(input)
		return fn(boolValue, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodStringBool[T]) Overwrite(transform func(T) T, params ...any) *ZodStringBool[T] {
	// Create a transformation function that works with the exact type T
	transformAny := func(input any) any {
		// Try to convert input to type T
		converted, ok := convertToStringBoolType[T](input)
		if !ok {
			// If conversion fails, return original value
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

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodStringBool[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		boolValue := extractStringBool(input)
		return target.Parse(boolValue, ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates new instance with pointer type
func (z *ZodStringBool[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodStringBool[*bool] {
	return &ZodStringBool[*bool]{internals: &ZodStringBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Truthy:           z.internals.Truthy,
		Falsy:            z.internals.Falsy,
	}}
}

// withInternals creates new instance preserving generic type T
func (z *ZodStringBool[T]) withInternals(in *core.ZodTypeInternals) *ZodStringBool[T] {
	return &ZodStringBool[T]{internals: &ZodStringBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Truthy:           z.internals.Truthy,
		Falsy:            z.internals.Falsy,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodStringBool[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodStringBool[T]); ok {
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
	}
}

// extractStringBool extracts the boolean value from the constraint type T
func extractStringBool[T StringBoolConstraint](value T) bool {
	switch v := any(value).(type) {
	case bool:
		return v
	case *bool:
		if v != nil {
			return *v
		}
		return false
	default:
		return false
	}
}

// convertToStringBoolType converts any value to the stringbool constraint type T with strict type checking
func convertToStringBoolType[T StringBoolConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		// Handle nil values for pointer types
		switch any(zero).(type) {
		case *bool:
			return any((*bool)(nil)).(T), true // return nil pointer for pointer types
		default:
			return zero, false // nil not allowed for value types
		}
	}

	// Extract boolean value from input
	var boolValue bool
	var isValid bool

	switch val := v.(type) {
	case bool:
		boolValue, isValid = val, true
	case *bool:
		if val != nil {
			boolValue, isValid = *val, true
		}
	default:
		return zero, false // Reject all non-bool types
	}

	if !isValid {
		return zero, false
	}

	// Convert to target type T
	switch any(zero).(type) {
	case bool:
		return any(boolValue).(T), true
	case *bool:
		return any(&boolValue).(T), true
	default:
		return zero, false
	}
}

// tryStringToBool tries to convert string to bool, returns (result, success)
func (z *ZodStringBool[T]) tryStringToBool(value string) (bool, bool) {
	normalizedValue := value
	if z.internals.Def.Case == "insensitive" {
		normalizedValue = strings.ToLower(value)
	}

	// Check truthy values
	if _, exists := z.internals.Truthy[normalizedValue]; exists {
		return true, true
	}

	// Check falsy values
	if _, exists := z.internals.Falsy[normalizedValue]; exists {
		return false, true
	}

	// Unrecognized value
	return false, false
}

// extractStringBoolForEngine extracts bool value from input for ParseComplex
func (z *ZodStringBool[T]) extractStringBoolForEngine(input any) (bool, bool) {
	var boolResult bool
	var success bool

	// Handle string input (primary use case)
	if str, ok := input.(string); ok {
		boolResult, success = z.tryStringToBool(str)
	} else if ptr, ok := input.(*string); ok {
		// Handle *string input
		if ptr == nil {
			return false, false
		}
		boolResult, success = z.tryStringToBool(*ptr)
	}

	// Try coercion if enabled and no success yet
	if !success && z.internals.ZodTypeInternals.IsCoerce() {
		if coerced, ok := z.Coerce(input); ok {
			return z.extractStringBoolForEngine(coerced)
		}
	}

	return boolResult, success
}

// extractStringBoolPtrForEngine extracts *bool value from input for ParseComplex
func (z *ZodStringBool[T]) extractStringBoolPtrForEngine(input any) (*bool, bool) {
	if ptr, ok := input.(*bool); ok {
		return ptr, true
	}
	return nil, false
}

// validateStringBoolForEngine validates bool value for ParseComplex
func (z *ZodStringBool[T]) validateStringBoolForEngine(value bool, checks []core.ZodCheck, ctx *core.ParseContext) (bool, error) {
	return z.validateStringBoolValue(value, checks, ctx)
}

// newZodStringBoolFromDef constructs new ZodStringBool from definition
func newZodStringBoolFromDef[T StringBoolConstraint](def *ZodStringBoolDef) *ZodStringBool[T] {
	internals := &ZodStringBoolInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:    def,
		Truthy: make(map[string]struct{}),
		Falsy:  make(map[string]struct{}),
	}

	// Build lookup maps for efficient validation
	for _, value := range def.Truthy {
		normalizedValue := value
		if def.Case == "insensitive" {
			normalizedValue = strings.ToLower(value)
		}
		internals.Truthy[normalizedValue] = struct{}{}
	}

	for _, value := range def.Falsy {
		normalizedValue := value
		if def.Case == "insensitive" {
			normalizedValue = strings.ToLower(value)
		}
		internals.Falsy[normalizedValue] = struct{}{}
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		stringBoolDef := &ZodStringBoolDef{
			ZodTypeDef:    *newDef,
			Truthy:        def.Truthy,
			Falsy:         def.Falsy,
			Case:          def.Case,
			CustomOptions: def.CustomOptions,
		}
		return any(newZodStringBoolFromDef[T](stringBoolDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodStringBool[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// StringBool creates bool schema with type-inference support
func StringBool(params ...any) *ZodStringBool[bool] {
	return StringBoolTyped[bool](params...)
}

// StringBoolPtr creates schema for *bool
func StringBoolPtr(params ...any) *ZodStringBool[*bool] {
	return StringBoolTyped[*bool](params...)
}

// StringBoolTyped is the generic constructor for stringbool schemas
func StringBoolTyped[T StringBoolConstraint](params ...any) *ZodStringBool[T] {
	var options *StringBoolOptions
	var schemaParams []any

	if len(params) > 0 {
		// Special-case nil placeholder for options
		if params[0] == nil {
			// No options provided, remaining args are treated as params
			if len(params) > 1 {
				schemaParams = params[1:]
			}
		} else if opt, ok := params[0].(*StringBoolOptions); ok {
			options = opt
			if len(params) > 1 {
				schemaParams = params[1:]
			}
		} else if opt, ok := params[0].(StringBoolOptions); ok {
			options = &opt
			if len(params) > 1 {
				schemaParams = params[1:]
			}
		} else {
			schemaParams = params
		}
	}

	// Default values
	truthy := []string{"true", "1", "yes", "on", "y", "enabled"}
	falsy := []string{"false", "0", "no", "off", "n", "disabled"}
	caseMode := "insensitive"
	customOptions := false

	// Apply custom options
	if options != nil {
		customOptions = true
		if len(options.Truthy) > 0 {
			truthy = options.Truthy
		}
		if len(options.Falsy) > 0 {
			falsy = options.Falsy
		}
		if options.Case != "" {
			caseMode = options.Case
		}
	}

	normalizedParams := utils.NormalizeParams(schemaParams...)

	def := &ZodStringBoolDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStringBool,
			Checks: []core.ZodCheck{},
		},
		Truthy:        truthy,
		Falsy:         falsy,
		Case:          caseMode,
		CustomOptions: customOptions,
	}

	// Apply normalized parameters to schema definition
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodStringBoolFromDef[T](def)
}

// CoercedStringBool creates a coerced stringbool schema
func CoercedStringBool(params ...any) *ZodStringBool[bool] {
	schema := StringBool(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedStringBoolPtr creates a coerced stringbool schema for *bool
func CoercedStringBoolPtr(params ...any) *ZodStringBool[*bool] {
	schema := StringBoolPtr(params...)
	schema.internals.SetCoerce(true)
	return schema
}
