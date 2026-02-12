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

// StringBoolConstraint restricts values to bool or *bool.
type StringBoolConstraint interface {
	bool | *bool
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// StringBoolOptions configures truthy/falsy values and case sensitivity.
type StringBoolOptions struct {
	Truthy []string // Values that evaluate to true.
	Falsy  []string // Values that evaluate to false.
	Case   string   // "sensitive" or "insensitive".
}

// ZodStringBoolDef defines the configuration for a string-boolean schema.
type ZodStringBoolDef struct {
	core.ZodTypeDef
	Truthy        []string
	Falsy         []string
	Case          string
	CustomOptions bool
}

// ZodStringBoolInternals contains string-boolean validator internal state.
type ZodStringBoolInternals struct {
	core.ZodTypeInternals
	Def    *ZodStringBoolDef
	Truthy map[string]struct{}
	Falsy  map[string]struct{}
}

// ZodStringBool represents a type-safe string-to-boolean validation schema.
type ZodStringBool[T StringBoolConstraint] struct {
	internals *ZodStringBoolInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodStringBool[T]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodStringBool[T]) IsOptional() bool { return z.internals.IsOptional() }

// IsNilable reports whether this schema accepts nil values.
func (z *ZodStringBool[T]) IsNilable() bool { return z.internals.IsNilable() }

// Coerce converts input to a recognized truthy/falsy string, implementing the Coercible interface.
func (z *ZodStringBool[T]) Coerce(input any) (any, bool) {
	if str, err := coerce.ToString(input); err == nil {
		if _, ok := z.tryStringToBool(str); ok {
			return str, true
		}
	}
	return input, false
}

// Parse validates input and returns a value matching the generic type T.
func (z *ZodStringBool[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	result, err := engine.ParseComplex[bool](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		z.extractStringBoolForEngine,
		z.extractStringBoolPtrForEngine,
		engine.ApplyChecks[bool],
		ctx...,
	)
	if err != nil {
		var zero T
		return zero, err
	}
	return engine.ConvertToConstraintType[bool, T](result, core.NewParseContext(), z.expectedType())
}

// MustParse validates input and panics on failure.
func (z *ZodStringBool[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any type for runtime interface usage.
func (z *ZodStringBool[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// MustParseAny validates input and panics on failure.
func (z *ZodStringBool[T]) MustParseAny(input any, ctx ...*core.ParseContext) any {
	result, err := z.ParseAny(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type T.
func (z *ZodStringBool[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[bool, T](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		engine.ApplyChecks[bool],
		ctx...,
	)
}

// MustStrictParse provides compile-time type safety and panics on failure.
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

// Optional returns a new schema that accepts nil, with *bool constraint.
func (z *ZodStringBool[T]) Optional() *ZodStringBool[*bool] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
// Unlike Optional, ExactOptional only accepts absent keys in object fields.
func (z *ZodStringBool[T]) ExactOptional() *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a new schema that accepts nil values, with *bool constraint.
func (z *ZodStringBool[T]) Nilable() *ZodStringBool[*bool] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a new schema combining optional and nilable modifiers.
func (z *ZodStringBool[T]) Nullish() *ZodStringBool[*bool] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits validation).
func (z *ZodStringBool[T]) Default(v bool) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits validation).
func (z *ZodStringBool[T]) DefaultFunc(fn func() bool) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
// Accepts string input type per Zod v4 semantics for StringBool.
func (z *ZodStringBool[T]) Prefault(v string) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
// Returns string input type per Zod v4 semantics for StringBool.
func (z *ZodStringBool[T]) PrefaultFunc(fn func() string) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this schema in the global registry.
func (z *ZodStringBool[T]) Meta(meta core.GlobalMeta) *ZodStringBool[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodStringBool[T]) Describe(description string) *ZodStringBool[T] {
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

// Refine applies a custom validation function matching the schema's output type T.
func (z *ZodStringBool[T]) Refine(fn func(T) bool, params ...any) *ZodStringBool[T] {
	wrapper := func(v any) bool {
		var zero T
		switch any(zero).(type) {
		case bool:
			if v == nil {
				return false
			}
			if boolVal, ok := v.(bool); ok {
				return fn(any(boolVal).(T))
			}
			return false
		case *bool:
			if v == nil {
				return fn(any((*bool)(nil)).(T))
			}
			if boolVal, ok := v.(bool); ok {
				bCopy := boolVal
				return fn(any(&bCopy).(T))
			}
			return false
		default:
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	return z.withCheck(check)
}

// RefineAny applies a custom validation function that receives the raw value.
func (z *ZodStringBool[T]) RefineAny(fn func(any) bool, params ...any) *ZodStringBool[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}
	check := checks.NewCustom[any](fn, errorMessage)
	return z.withCheck(check)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed value.
func (z *ZodStringBool[T]) Transform(fn func(bool, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(extractStringBool(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodStringBool[T]) Overwrite(transform func(T) T, params ...any) *ZodStringBool[T] {
	transformAny := func(input any) any {
		converted, ok := convertToStringBoolType[T](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	check := checks.NewZodCheckOverwrite(transformAny, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodStringBool[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	targetFn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractStringBool(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, targetFn)
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodStringBool[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodStringBool[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToStringBoolType converts only bool values to the target type T.
func convertToStringBoolType[T StringBoolConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		switch any(zero).(type) {
		case *bool:
			return zero, true
		default:
			return zero, false
		}
	}

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
		return zero, false
	}

	if !isValid {
		return zero, false
	}

	switch any(zero).(type) {
	case bool:
		return any(boolValue).(T), true
	case *bool:
		return any(&boolValue).(T), true
	default:
		return zero, false
	}
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// expectedType returns the schema's type code, defaulting to ZodTypeStringBool.
func (z *ZodStringBool[T]) expectedType() core.ZodTypeCode {
	if z.internals.Type != "" {
		return z.internals.Type
	}
	return core.ZodTypeStringBool
}

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodStringBool[T]) withCheck(check core.ZodCheck) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withPtrInternals creates a new *bool schema from cloned internals.
func (z *ZodStringBool[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodStringBool[*bool] {
	return &ZodStringBool[*bool]{internals: &ZodStringBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Truthy:           z.internals.Truthy,
		Falsy:            z.internals.Falsy,
	}}
}

// withInternals creates a new schema preserving generic type T.
func (z *ZodStringBool[T]) withInternals(in *core.ZodTypeInternals) *ZodStringBool[T] {
	return &ZodStringBool[T]{internals: &ZodStringBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Truthy:           z.internals.Truthy,
		Falsy:            z.internals.Falsy,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodStringBool[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodStringBool[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractStringBool extracts the underlying bool from a generic constraint type.
func extractStringBool[T StringBoolConstraint](value T) bool {
	if ptr, ok := any(value).(*bool); ok {
		return ptr != nil && *ptr
	}
	return any(value).(bool)
}

// tryStringToBool converts a string to bool using the configured truthy/falsy sets.
func (z *ZodStringBool[T]) tryStringToBool(value string) (bool, bool) {
	normalized := value
	if z.internals.Def.Case == "insensitive" {
		normalized = strings.ToLower(value)
	}

	if _, ok := z.internals.Truthy[normalized]; ok {
		return true, true
	}
	if _, ok := z.internals.Falsy[normalized]; ok {
		return false, true
	}
	return false, false
}

// extractStringBoolForEngine extracts bool from string input for ParseComplex.
func (z *ZodStringBool[T]) extractStringBoolForEngine(input any) (bool, bool) {
	switch v := input.(type) {
	case string:
		return z.tryStringToBool(v)
	case *string:
		if v == nil {
			return false, false
		}
		return z.tryStringToBool(*v)
	}

	// Try coercion if enabled.
	if z.internals.IsCoerce() {
		if coerced, ok := z.Coerce(input); ok {
			return z.extractStringBoolForEngine(coerced)
		}
	}

	return false, false
}

// extractStringBoolPtrForEngine extracts *bool from input for ParseComplex.
func (z *ZodStringBool[T]) extractStringBoolPtrForEngine(input any) (*bool, bool) {
	if ptr, ok := input.(*bool); ok {
		return ptr, true
	}
	return nil, false
}

// newZodStringBoolFromDef constructs a new ZodStringBool from a definition.
func newZodStringBoolFromDef[T StringBoolConstraint](def *ZodStringBoolDef) *ZodStringBool[T] {
	internals := &ZodStringBoolInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:    def,
		Truthy: make(map[string]struct{}, len(def.Truthy)),
		Falsy:  make(map[string]struct{}, len(def.Falsy)),
	}

	for _, v := range def.Truthy {
		key := v
		if def.Case == "insensitive" {
			key = strings.ToLower(v)
		}
		internals.Truthy[key] = struct{}{}
	}
	for _, v := range def.Falsy {
		key := v
		if def.Case == "insensitive" {
			key = strings.ToLower(v)
		}
		internals.Falsy[key] = struct{}{}
	}

	// Provide constructor for AddCheck functionality.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		sbDef := &ZodStringBoolDef{
			ZodTypeDef:    *newDef,
			Truthy:        def.Truthy,
			Falsy:         def.Falsy,
			Case:          def.Case,
			CustomOptions: def.CustomOptions,
		}
		return any(newZodStringBoolFromDef[T](sbDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodStringBool[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// StringBool creates a string-to-bool validation schema.
func StringBool(params ...any) *ZodStringBool[bool] {
	return StringBoolTyped[bool](params...)
}

// StringBoolPtr creates a string-to-*bool validation schema.
func StringBoolPtr(params ...any) *ZodStringBool[*bool] {
	return StringBoolTyped[*bool](params...)
}

// StringBoolTyped is the generic constructor for string-boolean schemas.
func StringBoolTyped[T StringBoolConstraint](params ...any) *ZodStringBool[T] {
	var options *StringBoolOptions
	var schemaParams []any

	if len(params) > 0 {
		switch v := params[0].(type) {
		case nil:
			if len(params) > 1 {
				schemaParams = params[1:]
			}
		case *StringBoolOptions:
			options = v
			if len(params) > 1 {
				schemaParams = params[1:]
			}
		case StringBoolOptions:
			options = &v
			if len(params) > 1 {
				schemaParams = params[1:]
			}
		default:
			schemaParams = params
		}
	}

	truthy := []string{"true", "1", "yes", "on", "y", "enabled"}
	falsy := []string{"false", "0", "no", "off", "n", "disabled"}
	caseMode := "insensitive"
	customOptions := false

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

	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodStringBoolFromDef[T](def)
}

// CoercedStringBool creates a coerced string-to-bool schema.
func CoercedStringBool(params ...any) *ZodStringBool[bool] {
	schema := StringBool(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedStringBoolPtr creates a coerced string-to-*bool schema.
func CoercedStringBoolPtr(params ...any) *ZodStringBool[*bool] {
	schema := StringBoolPtr(params...)
	schema.internals.SetCoerce(true)
	return schema
}
