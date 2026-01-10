package types

import (
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// TupleConstraint defines valid constraint types for tuple schemas
type TupleConstraint interface {
	[]any | *[]any
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodTupleDef defines the schema definition for tuple validation
// TypeScript Zod v4 equivalent: $ZodTupleDef
type ZodTupleDef struct {
	core.ZodTypeDef
	Items []core.ZodSchema // Fixed positional item schemas
	Rest  core.ZodSchema   // Rest element schema (optional, for variadic elements)
}

// ZodTupleInternals contains the internal state for tuple schema
type ZodTupleInternals struct {
	core.ZodTypeInternals
	Def           *ZodTupleDef     // Schema definition reference
	Items         []core.ZodSchema // Item schemas for runtime validation
	Rest          core.ZodSchema   // Rest element schema
	RequiredCount int              // Number of required items (up to first optional from end)
}

// ZodTuple represents a type-safe tuple validation schema
// TypeScript Zod v4 equivalent: z.tuple([...])
// T: base type ([]any), R: constraint type ([]any or *[]any)
type ZodTuple[T any, R any] struct {
	internals *ZodTupleInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodTuple[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetItems returns the tuple item schemas
func (z *ZodTuple[T, R]) GetItems() []core.ZodSchema {
	return z.internals.Items
}

// GetRest returns the rest element schema
func (z *ZodTuple[T, R]) GetRest() core.ZodSchema {
	return z.internals.Rest
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodTuple[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodTuple[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input using tuple-specific parsing logic with engine.ParseComplex
func (z *ZodTuple[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplex[[]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeTuple,
		z.extractTupleForEngine,
		z.extractTuplePtrForEngine,
		z.validateTupleForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	// Handle different return types from ParseComplex
	switch v := result.(type) {
	case []any:
		return convertToTupleConstraintType[R](v), nil
	case *[]any:
		if v == nil {
			var zero R
			return zero, nil
		}
		return convertToTupleConstraintType[R](*v), nil
	case nil:
		var zero R
		return zero, nil
	default:
		if typedResult, ok := result.(R); ok {
			return typedResult, nil
		}
		var zero R
		return zero, issues.CreateInvalidTypeError(core.ZodTypeTuple, result, ctx[0])
	}
}

// MustParse validates the input value and panics on failure
func (z *ZodTuple[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
func (z *ZodTuple[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	// Convert T to R for ParseComplexStrict
	constraintInput, ok := convertToTupleType[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"tuple constraint type",
			any(input),
			ctx[0],
		)
	}

	result, err := engine.ParseComplexStrict[[]any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeTuple,
		z.extractTupleForEngine,
		z.extractTuplePtrForEngine,
		z.validateTupleForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	return result, nil
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
// This method provides zero-overhead abstraction with strict type constraints.
func (z *ZodTuple[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodTuple[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional marks this tuple as optional
func (z *ZodTuple[T, R]) Optional() *ZodTuple[T, *[]any] {
	newInternals := z.internals.Clone()
	newInternals.SetOptional(true)
	return &ZodTuple[T, *[]any]{internals: z.withClonedInternals(newInternals)}
}

// Nilable marks this tuple as nilable
func (z *ZodTuple[T, R]) Nilable() *ZodTuple[T, *[]any] {
	newInternals := z.internals.Clone()
	newInternals.SetNilable(true)
	return &ZodTuple[T, *[]any]{internals: z.withClonedInternals(newInternals)}
}

// Nullish marks this tuple as nullish (optional + nilable)
func (z *ZodTuple[T, R]) Nullish() *ZodTuple[T, *[]any] {
	newInternals := z.internals.Clone()
	newInternals.SetOptional(true)
	newInternals.SetNilable(true)
	return &ZodTuple[T, *[]any]{internals: z.withClonedInternals(newInternals)}
}

// Default sets a default value for nil input
func (z *ZodTuple[T, R]) Default(v []any) *ZodTuple[T, R] {
	newInternals := z.internals.Clone()
	newInternals.SetDefaultValue(v)
	return z.withInternals(newInternals)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum tuple length validation
func (z *ZodTuple[T, R]) Min(minimum int, params ...any) *ZodTuple[T, R] {
	check := checks.MinLength(minimum, params...)
	newInternals := z.internals.Clone()
	newInternals.Checks = append(newInternals.Checks, check)
	return z.withInternals(newInternals)
}

// Max sets maximum tuple length validation
func (z *ZodTuple[T, R]) Max(maximum int, params ...any) *ZodTuple[T, R] {
	check := checks.MaxLength(maximum, params...)
	newInternals := z.internals.Clone()
	newInternals.Checks = append(newInternals.Checks, check)
	return z.withInternals(newInternals)
}

// Length sets exact tuple length validation
func (z *ZodTuple[T, R]) Length(length int, params ...any) *ZodTuple[T, R] {
	check := checks.Length(length, params...)
	newInternals := z.internals.Clone()
	newInternals.Checks = append(newInternals.Checks, check)
	return z.withInternals(newInternals)
}

// NonEmpty ensures tuple has at least one element
func (z *ZodTuple[T, R]) NonEmpty(params ...any) *ZodTuple[T, R] {
	return z.Min(1, params...)
}

// Refine adds custom validation logic
func (z *ZodTuple[T, R]) Refine(fn func([]any) bool, params ...any) *ZodTuple[T, R] {
	wrapper := func(v any) bool {
		if arr, ok := v.([]any); ok {
			return fn(arr)
		}
		return false
	}
	check := checks.NewCustom[any](wrapper, params...)
	newInternals := z.internals.Clone()
	newInternals.Checks = append(newInternals.Checks, check)
	return z.withInternals(newInternals)
}

// Check adds a custom validation check with payload access
func (z *ZodTuple[T, R]) Check(check core.ZodCheck) *ZodTuple[T, R] {
	newInternals := z.internals.Clone()
	newInternals.Checks = append(newInternals.Checks, check)
	return z.withInternals(newInternals)
}

// With is an alias for Check (TypeScript Zod v4 compatibility)
func (z *ZodTuple[T, R]) With(check core.ZodCheck) *ZodTuple[T, R] {
	return z.Check(check)
}

// =============================================================================
// METADATA METHODS
// =============================================================================

// Meta stores metadata for this tuple schema
func (z *ZodTuple[T, R]) Meta(meta core.GlobalMeta) *ZodTuple[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodTuple[T, R]) Describe(description string) *ZodTuple[T, R] {
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
// TUPLE-SPECIFIC METHODS
// =============================================================================

// Rest sets a rest element schema for additional elements beyond fixed items
// TypeScript Zod v4 equivalent: tuple.rest(schema)
func (z *ZodTuple[T, R]) Rest(restSchema core.ZodSchema) *ZodTuple[T, R] {
	newDef := &ZodTupleDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Items:      z.internals.Items,
		Rest:       restSchema,
	}
	return newZodTupleFromDef[T, R](newDef, z.internals.RequiredCount)
}

// =============================================================================
// ENGINE INTEGRATION METHODS
// =============================================================================

// extractTupleForEngine extracts tuple type from input for engine.ParseComplex
func (z *ZodTuple[T, R]) extractTupleForEngine(input any) ([]any, bool) {
	return toSliceAny(input)
}

// extractTuplePtrForEngine extracts pointer to tuple for engine.ParseComplex
func (z *ZodTuple[T, R]) extractTuplePtrForEngine(input any) (*[]any, bool) {
	if input == nil {
		return nil, true
	}
	arr, ok := toSliceAny(input)
	if !ok {
		return nil, false
	}
	return &arr, true
}

// validateTupleForEngine performs tuple validation for engine.ParseComplex
func (z *ZodTuple[T, R]) validateTupleForEngine(arr []any, chks []core.ZodCheck, ctx *core.ParseContext) ([]any, error) {
	// Check minimum length (required elements)
	if len(arr) < z.internals.RequiredCount {
		return nil, issues.CreateTooSmallError(z.internals.RequiredCount, true, "tuple", arr, ctx)
	}

	// Check maximum length when no rest schema
	maxLen := len(z.internals.Items)
	if z.internals.Rest == nil && len(arr) > maxLen {
		return nil, issues.CreateTooBigError(maxLen, true, "tuple", arr, ctx)
	}

	result := make([]any, len(arr))
	var collectedIssues []core.ZodRawIssue

	// Validate fixed items
	for i, schema := range z.internals.Items {
		if i >= len(arr) {
			// Element not provided - only valid for optional items
			break
		}

		val, err := schema.ParseAny(arr[i], ctx)
		if err != nil {
			// Collect error with path
			var zodErr *issues.ZodError
			if issues.IsZodError(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := core.ZodRawIssue{
						Code:       issue.Code,
						Message:    issue.Message,
						Input:      issue.Input,
						Path:       append([]any{i}, issue.Path...),
						Properties: make(map[string]any),
					}
					collectedIssues = append(collectedIssues, rawIssue)
				}
			} else {
				rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, arr[i])
				rawIssue.Path = []any{i}
				collectedIssues = append(collectedIssues, rawIssue)
			}
			continue
		}
		result[i] = val
	}

	// Validate rest elements (beyond fixed items)
	if z.internals.Rest != nil && len(arr) > len(z.internals.Items) {
		for i := len(z.internals.Items); i < len(arr); i++ {
			val, err := z.internals.Rest.ParseAny(arr[i], ctx)
			if err != nil {
				var zodErr *issues.ZodError
				if issues.IsZodError(err, &zodErr) {
					for _, issue := range zodErr.Issues {
						rawIssue := core.ZodRawIssue{
							Code:       issue.Code,
							Message:    issue.Message,
							Input:      issue.Input,
							Path:       append([]any{i}, issue.Path...),
							Properties: make(map[string]any),
						}
						collectedIssues = append(collectedIssues, rawIssue)
					}
				} else {
					rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, arr[i])
					rawIssue.Path = []any{i}
					collectedIssues = append(collectedIssues, rawIssue)
				}
				continue
			}
			result[i] = val
		}
	}

	// Return error if any issues collected
	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}

	// Apply additional checks (already handled by engine, but needed for validation callback)
	if len(chks) > 0 {
		checkedResult, err := engine.ApplyChecks[any](result, chks, ctx)
		if err != nil {
			return nil, err
		}
		if arr, ok := checkedResult.([]any); ok {
			result = arr
		}
	}

	return result, nil
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// toSliceAny converts input to []any
func toSliceAny(input any) ([]any, bool) {
	if arr, ok := input.([]any); ok {
		return arr, true
	}

	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, false
	}

	result := make([]any, v.Len())
	for i := range v.Len() {
		result[i] = v.Index(i).Interface()
	}
	return result, true
}

// withInternals creates new instance preserving type
func (z *ZodTuple[T, R]) withInternals(in *core.ZodTypeInternals) *ZodTuple[T, R] {
	return &ZodTuple[T, R]{internals: &ZodTupleInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
		RequiredCount:    z.internals.RequiredCount,
	}}
}

// withClonedInternals creates ZodTupleInternals from cloned ZodTypeInternals
func (z *ZodTuple[T, R]) withClonedInternals(in *core.ZodTypeInternals) *ZodTupleInternals {
	return &ZodTupleInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
		RequiredCount:    z.internals.RequiredCount,
	}
}

// convertToTupleConstraintType converts []any to constraint type R
func convertToTupleConstraintType[R any](arr []any) R {
	var zero R
	zeroType := reflect.TypeOf(zero)

	if zeroType == nil {
		// R is any or interface{}
		return any(arr).(R)
	}

	switch zeroType.Kind() { //nolint:exhaustive // only handling relevant cases
	case reflect.Ptr:
		// R is *[]any
		return any(&arr).(R)
	case reflect.Slice:
		// R is []any
		return any(arr).(R)
	default:
		return any(arr).(R)
	}
}

// convertToTupleType converts input type T to constraint type R
func convertToTupleType[T any, R any](input T) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(input).(R); ok {
		return r, true
	}

	// Handle pointer conversion
	if _, ok := any(zero).(*[]any); ok {
		// Need to convert []any to *[]any
		if arr, ok := any(input).([]any); ok {
			return any(&arr).(R), true
		}
	}

	return zero, false
}

// calculateRequiredCount calculates the number of required items
// Required items are counted from the beginning until all remaining items are optional
func calculateRequiredCount(items []core.ZodSchema) int {
	// Find the last non-optional item
	lastRequired := -1
	for i := len(items) - 1; i >= 0; i-- {
		if !items[i].GetInternals().IsOptional() {
			lastRequired = i
			break
		}
	}
	return lastRequired + 1
}

// newZodTupleFromDef constructs new ZodTuple from definition
func newZodTupleFromDef[T any, R any](def *ZodTupleDef, requiredCount int) *ZodTuple[T, R] {
	internals := &ZodTupleInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Items:            def.Items,
		Rest:             def.Rest,
		RequiredCount:    requiredCount,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		tupleDef := &ZodTupleDef{
			ZodTypeDef: *newDef,
			Items:      def.Items,
			Rest:       def.Rest,
		}
		return any(newZodTupleFromDef[T, R](tupleDef, requiredCount)).(core.ZodType[any])
	}

	schema := &ZodTuple[T, R]{internals: internals}

	// Set error if provided
	if def.Error != nil {
		internals.Error = def.Error
	}

	// Set checks if provided
	if len(def.Checks) > 0 {
		internals.Checks = append(internals.Checks, def.Checks...)
	}

	// Ensure validator is called when item schemas exist
	// Add a minimal check that always passes to trigger validation
	if len(def.Items) > 0 || def.Rest != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		internals.AddCheck(alwaysPassCheck)
	}

	return schema
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Tuple creates a tuple schema with fixed positional items
// TypeScript Zod v4 equivalent: z.tuple([...])
//
// Usage:
//
//	tuple := gozod.Tuple(gozod.String(), gozod.Int())
//	result, err := tuple.Parse([]any{"hello", 42})
func Tuple(items ...core.ZodSchema) *ZodTuple[[]any, []any] {
	return TupleTyped[[]any, []any](items, nil)
}

// TupleWithRest creates a tuple schema with fixed items and a rest element
// TypeScript Zod v4 equivalent: z.tuple([...]).rest(schema)
//
// Usage:
//
//	tuple := gozod.TupleWithRest([]core.ZodSchema{gozod.String(), gozod.Int()}, gozod.Bool())
//	result, err := tuple.Parse([]any{"hello", 42, true, false})
func TupleWithRest(items []core.ZodSchema, rest core.ZodSchema, params ...any) *ZodTuple[[]any, []any] {
	return TupleTyped[[]any, []any](items, rest, params...)
}

// TupleTyped creates a tuple schema with explicit type parameters
func TupleTyped[T any, R any](items []core.ZodSchema, rest core.ZodSchema, params ...any) *ZodTuple[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	requiredCount := calculateRequiredCount(items)

	def := &ZodTupleDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeTuple,
			Checks: []core.ZodCheck{},
		},
		Items: items,
		Rest:  rest,
	}

	// Apply schema params
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodTupleFromDef[T, R](def, requiredCount)
}

// TuplePtr creates an optional tuple schema that returns *[]any
func TuplePtr(items ...core.ZodSchema) *ZodTuple[[]any, *[]any] {
	schema := Tuple(items...)
	return schema.Optional()
}
