package types

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// XOR (EXCLUSIVE UNION) - Zod v4 Compatible
// =============================================================================

// ZodXorDef defines the schema definition for exclusive union validation
type ZodXorDef struct {
	core.ZodTypeDef
	Options []core.ZodSchema // Union member schemas
}

// ZodXorInternals contains the internal state for exclusive union schema
type ZodXorInternals struct {
	core.ZodTypeInternals
	Def     *ZodXorDef       // Schema definition reference
	Options []core.ZodSchema // Union member schemas for runtime validation
}

// ZodXor represents an exclusive union validation schema (exactly one must match)
// T = base type (any), R = constraint type (any or *any)
type ZodXor[T any, R any] struct {
	internals *ZodXorInternals
}

// GetInternals exposes internal state for framework usage
func (z *ZodXor[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodXor[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodXor[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input ensuring exactly one option matches
func (z *ZodXor[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	result, err := engine.ParseComplex[any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnion, // Use union type for error reporting
		z.extractXorType,
		z.extractXorPtr,
		z.validateXorValue,
		parseCtx,
	)
	if err != nil {
		var zero R
		return zero, err
	}
	return convertToUnionConstraintType[T, R](result), nil
}

// extractXorType extracts the type from input
func (z *ZodXor[T, R]) extractXorType(input any) (any, bool) {
	return input, true
}

// extractXorPtr extracts pointer from input
func (z *ZodXor[T, R]) extractXorPtr(input any) (*any, bool) {
	if input == nil {
		return nil, true
	}
	return &input, true
}

// validateXorValue validates that exactly one option matches (Zod v4 semantics)
func (z *ZodXor[T, R]) validateXorValue(input any, chks []core.ZodCheck, parseCtx *core.ParseContext) (any, error) {
	var (
		successes []any
		allErrors []error
	)

	// Try each option and count successes
	for i, option := range z.internals.Options {
		if option == nil {
			continue
		}

		if result, err := option.ParseAny(input, parseCtx); err == nil {
			// Apply custom checks on the xor schema itself
			if len(chks) > 0 {
				transformedResult, validationErr := engine.ApplyChecks[any](result, chks, parseCtx)
				if validationErr != nil {
					allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, validationErr))
					continue
				}
				result = transformedResult
			}
			successes = append(successes, result)
		} else {
			allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, err))
		}
	}

	// Exactly one success required
	if len(successes) == 1 {
		return successes[0], nil
	}

	// Zero or multiple successes - error
	if len(successes) == 0 {
		// No matches - same as regular union error
		if len(allErrors) == 0 {
			return nil, issues.CreateInvalidSchemaError("no xor options provided", input, parseCtx)
		}
		return nil, issues.CreateInvalidUnionError(allErrors, input, parseCtx)
	}

	// Multiple matches - exclusive union failure
	return nil, issues.CreateInvalidXorError(len(successes), input, parseCtx)
}

// MustParse validates the input value and panics on failure
func (z *ZodXor[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety
func (z *ZodXor[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	constraintInput, ok := convertToUnionConstraintValue[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"xor constraint type",
			any(input),
			ctx[0],
		)
	}

	result, err := engine.ParseComplexStrict[any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnion,
		z.extractXorType,
		z.extractXorPtr,
		z.validateXorValue,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}
	return result, nil
}

// MustStrictParse validates the input value with compile-time type safety and panics on failure
func (z *ZodXor[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any type
func (z *ZodXor[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Options returns the union member schemas for JSON Schema conversion
func (z *ZodXor[T, R]) Options() []core.ZodSchema {
	return z.internals.Options
}

// =============================================================================
// METADATA METHODS
// =============================================================================

// Meta stores metadata for this exclusive union schema
func (z *ZodXor[T, R]) Meta(meta core.GlobalMeta) *ZodXor[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
//
// Example:
//
//	schema := gozod.Xor(gozod.String(), gozod.Int()).Describe("Exactly one of string or integer")
func (z *ZodXor[T, R]) Describe(description string) *ZodXor[T, R] {
	// Follow Enhanced Copy-on-Write pattern
	newInternals := z.internals.Clone()

	// Get existing metadata or create new
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description

	// Create new schema instance with cloned internals
	clone := z.xorWithInternals(newInternals)
	core.GlobalRegistry.Add(clone, existing)

	return clone
}

// xorWithInternals creates a new ZodXor with updated internals
func (z *ZodXor[T, R]) xorWithInternals(in *core.ZodTypeInternals) *ZodXor[T, R] {
	return &ZodXor[T, R]{
		internals: &ZodXorInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Options:          z.internals.Options,
		},
	}
}

// newZodXorFromDef creates a new ZodXor instance from definition
func newZodXorFromDef[T any, R any](def *ZodXorDef) *ZodXor[T, R] {
	internals := &ZodXorInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Options:          def.Options,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		xorDef := &ZodXorDef{
			ZodTypeDef: *newDef,
			Options:    def.Options,
		}
		return any(newZodXorFromDef[T, R](xorDef)).(core.ZodType[any])
	}

	schema := &ZodXor[T, R]{internals: internals}

	// Set error if provided
	if def.Error != nil {
		internals.Error = def.Error
	}

	// Set checks if provided
	if len(def.Checks) > 0 {
		for _, check := range def.Checks {
			internals.AddCheck(check)
		}
	}

	return schema
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// Xor creates exclusive union schema that requires exactly one option to match
// Unlike Union which succeeds when any option matches, Xor fails if zero or multiple match.
func Xor(options []any, args ...any) *ZodXor[any, any] {
	return XorTyped[any, any](options, args...)
}

// XorPtr creates exclusive union schema returning pointer constraint
func XorPtr(options []any, args ...any) *ZodXor[any, *any] {
	return XorTyped[any, *any](options, args...)
}

// XorOf creates exclusive union schema from variadic arguments
func XorOf(schemas ...any) *ZodXor[any, any] {
	return Xor(schemas)
}

// XorOfPtr creates exclusive union schema from variadic arguments returning pointer
func XorOfPtr(schemas ...any) *ZodXor[any, *any] {
	return XorPtr(schemas)
}

// XorTyped creates typed exclusive union schema with generic constraints
func XorTyped[T any, R any](options []any, args ...any) *ZodXor[T, R] {
	param := utils.FirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	// Convert all options to ZodSchema
	wrappedOptions := make([]core.ZodSchema, len(options))
	for i, option := range options {
		if option == nil {
			wrappedOptions[i] = nil
			continue
		}

		zodSchema, err := core.ConvertToZodSchema(option)
		if err != nil {
			panic(fmt.Sprintf("Xor option %d: %v", i, err))
		}
		wrappedOptions[i] = zodSchema
	}

	def := &ZodXorDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeXor, // Use dedicated Xor type for JSON Schema (oneOf)
			Checks: []core.ZodCheck{},
		},
		Options: wrappedOptions,
	}

	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	xorSchema := newZodXorFromDef[T, R](def)

	// Add minimal check to trigger validation
	alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
	xorSchema.internals.AddCheck(alwaysPassCheck)

	return xorSchema
}
