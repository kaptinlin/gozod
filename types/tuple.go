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

// TupleConstraint defines valid constraint types for tuple schemas.
type TupleConstraint interface {
	[]any | *[]any
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodTupleDef defines the schema definition for tuple validation.
type ZodTupleDef struct {
	core.ZodTypeDef
	Items []core.ZodSchema // Fixed positional item schemas.
	Rest  core.ZodSchema   // Rest element schema (optional, for variadic elements).
}

// ZodTupleInternals contains the internal state for tuple schema.
type ZodTupleInternals struct {
	core.ZodTypeInternals
	Def           *ZodTupleDef
	Items         []core.ZodSchema // Item schemas for runtime validation.
	Rest          core.ZodSchema   // Rest element schema.
	RequiredCount int              // Number of required items (up to first optional from end).
}

// ZodTuple represents a type-safe tuple validation schema.
// T is the base type ([]any), R is the constraint type ([]any or *[]any).
type ZodTuple[T any, R any] struct {
	internals *ZodTupleInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals exposes the internal state for framework usage.
func (z *ZodTuple[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Items returns the tuple item schemas.
func (z *ZodTuple[T, R]) Items() []core.ZodSchema {
	return z.internals.Items
}

// Rest returns the rest element schema.
func (z *ZodTuple[T, R]) Rest() core.ZodSchema {
	return z.internals.Rest
}

// IsOptional reports whether this schema accepts undefined or missing values.
func (z *ZodTuple[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodTuple[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates the input and returns a value matching the constraint type R.
func (z *ZodTuple[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

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
		return zero, err
	}

	// Handle different return types from ParseComplex.
	switch v := result.(type) {
	case []any:
		return convertToTupleConstraintType[R](v), nil
	case *[]any:
		if v == nil {
			return zero, nil
		}
		return convertToTupleConstraintType[R](*v), nil
	case nil:
		return zero, nil
	default:
		if typedResult, ok := result.(R); ok {
			return typedResult, nil
		}
		parseCtx := core.NewParseContext()
		if len(ctx) > 0 && ctx[0] != nil {
			parseCtx = ctx[0]
		}
		return zero, issues.CreateInvalidTypeError(core.ZodTypeTuple, result, parseCtx)
	}
}

// MustParse validates the input value and panics on failure.
func (z *ZodTuple[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring the exact type T.
func (z *ZodTuple[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	var zero R

	// Convert T to R for ParseComplexStrict.
	constraintInput, ok := convertToTupleType[T, R](input)
	if !ok {
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

	return engine.ParseComplexStrict[[]any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeTuple,
		z.extractTupleForEngine,
		z.extractTuplePtrForEngine,
		z.validateTupleForEngine,
		ctx...,
	)
}

// MustStrictParse provides compile-time type safety and panics on failure.
func (z *ZodTuple[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input and returns any type for runtime interface usage.
func (z *ZodTuple[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional marks this tuple as optional.
func (z *ZodTuple[T, R]) Optional() *ZodTuple[T, *[]any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodTuple[T, R]) ExactOptional() *ZodTuple[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable marks this tuple as nilable.
func (z *ZodTuple[T, R]) Nilable() *ZodTuple[T, *[]any] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish marks this tuple as nullish (both optional and nilable).
func (z *ZodTuple[T, R]) Nullish() *ZodTuple[T, *[]any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value for nil input.
func (z *ZodTuple[T, R]) Default(v []any) *ZodTuple[T, R] {
	newInternals := z.internals.Clone()
	newInternals.SetDefaultValue(v)
	return z.withInternals(newInternals)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets the minimum tuple length validation.
func (z *ZodTuple[T, R]) Min(minimum int, params ...any) *ZodTuple[T, R] {
	newInternals := z.internals.Clone()
	newInternals.AddCheck(checks.MinLength(minimum, params...))
	return z.withInternals(newInternals)
}

// Max sets the maximum tuple length validation.
func (z *ZodTuple[T, R]) Max(maximum int, params ...any) *ZodTuple[T, R] {
	newInternals := z.internals.Clone()
	newInternals.AddCheck(checks.MaxLength(maximum, params...))
	return z.withInternals(newInternals)
}

// Length sets the exact tuple length validation.
func (z *ZodTuple[T, R]) Length(length int, params ...any) *ZodTuple[T, R] {
	newInternals := z.internals.Clone()
	newInternals.AddCheck(checks.Length(length, params...))
	return z.withInternals(newInternals)
}

// NonEmpty ensures the tuple has at least one element.
func (z *ZodTuple[T, R]) NonEmpty(params ...any) *ZodTuple[T, R] {
	return z.Min(1, params...)
}

// Refine adds a custom validation function.
func (z *ZodTuple[T, R]) Refine(fn func([]any) bool, params ...any) *ZodTuple[T, R] {
	wrapper := func(v any) bool {
		if arr, ok := v.([]any); ok {
			return fn(arr)
		}
		return false
	}
	newInternals := z.internals.Clone()
	newInternals.AddCheck(checks.NewCustom[any](wrapper, params...))
	return z.withInternals(newInternals)
}

// Check adds a custom validation check.
func (z *ZodTuple[T, R]) Check(check core.ZodCheck) *ZodTuple[T, R] {
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// With is an alias for Check, for Zod v4 compatibility.
func (z *ZodTuple[T, R]) With(check core.ZodCheck) *ZodTuple[T, R] {
	return z.Check(check)
}

// =============================================================================
// METADATA METHODS
// =============================================================================

// Meta stores metadata for this tuple schema.
func (z *ZodTuple[T, R]) Meta(meta core.GlobalMeta) *ZodTuple[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
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

// WithRest sets a rest element schema for additional elements beyond the fixed items.
func (z *ZodTuple[T, R]) WithRest(restSchema core.ZodSchema) *ZodTuple[T, R] {
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

// extractTupleForEngine extracts the tuple type from input for engine.ParseComplex.
func (z *ZodTuple[T, R]) extractTupleForEngine(input any) ([]any, bool) {
	return toSliceAny(input)
}

// extractTuplePtrForEngine extracts a pointer to tuple for engine.ParseComplex.
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

// collectParseIssues extracts ZodRawIssues from a parse error at the given index.
func collectParseIssues(err error, index int, input any) []core.ZodRawIssue {
	var zodErr *issues.ZodError
	if issues.IsZodError(err, &zodErr) {
		result := make([]core.ZodRawIssue, 0, len(zodErr.Issues))
		for _, issue := range zodErr.Issues {
			result = append(result, core.ZodRawIssue{
				Code:       issue.Code,
				Message:    issue.Message,
				Input:      issue.Input,
				Path:       append([]any{index}, issue.Path...),
				Properties: make(map[string]any),
			})
		}
		return result
	}
	rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, input)
	rawIssue.Path = []any{index}
	return []core.ZodRawIssue{rawIssue}
}

// validateTupleForEngine performs tuple validation for engine.ParseComplex.
func (z *ZodTuple[T, R]) validateTupleForEngine(arr []any, chks []core.ZodCheck, ctx *core.ParseContext) ([]any, error) {
	// Check minimum length (required elements).
	if len(arr) < z.internals.RequiredCount {
		return nil, issues.CreateTooSmallError(z.internals.RequiredCount, true, "tuple", arr, ctx)
	}

	// Check maximum length when no rest schema.
	maxLen := len(z.internals.Items)
	if z.internals.Rest == nil && len(arr) > maxLen {
		return nil, issues.CreateTooBigError(maxLen, true, "tuple", arr, ctx)
	}

	result := make([]any, len(arr))
	var collectedIssues []core.ZodRawIssue

	// Validate fixed items.
	for i, schema := range z.internals.Items {
		if i >= len(arr) {
			break
		}

		val, err := schema.ParseAny(arr[i], ctx)
		if err != nil {
			collectedIssues = append(collectedIssues, collectParseIssues(err, i, arr[i])...)
			continue
		}
		result[i] = val
	}

	// Validate rest elements (beyond fixed items).
	if z.internals.Rest != nil && len(arr) > len(z.internals.Items) {
		for i := len(z.internals.Items); i < len(arr); i++ {
			val, err := z.internals.Rest.ParseAny(arr[i], ctx)
			if err != nil {
				collectedIssues = append(collectedIssues, collectParseIssues(err, i, arr[i])...)
				continue
			}
			result[i] = val
		}
	}

	// Return error if any issues collected.
	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}

	// Apply additional checks.
	if len(chks) > 0 {
		checkedResult, err := engine.ApplyChecks[[]any](result, chks, ctx)
		if err != nil {
			return nil, err
		}
		result = checkedResult
	}

	return result, nil
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// toSliceAny converts the input to []any.
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

// withInternals creates a new instance while preserving type parameters.
func (z *ZodTuple[T, R]) withInternals(in *core.ZodTypeInternals) *ZodTuple[T, R] {
	return &ZodTuple[T, R]{internals: &ZodTupleInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
		RequiredCount:    z.internals.RequiredCount,
	}}
}

// withPtrInternals creates a new ZodTuple with pointer constraint *[]any.
func (z *ZodTuple[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodTuple[T, *[]any] {
	return &ZodTuple[T, *[]any]{internals: &ZodTupleInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
		RequiredCount:    z.internals.RequiredCount,
	}}
}

// convertToTupleConstraintType converts []any to the constraint type R.
func convertToTupleConstraintType[R any](arr []any) R {
	var zero R
	zeroType := reflect.TypeOf(zero)

	if zeroType == nil {
		return any(arr).(R)
	}

	if zeroType.Kind() == reflect.Pointer && zeroType.Elem().Kind() == reflect.Slice {
		return any(new(arr)).(R)
	}
	return any(arr).(R)
}

// convertToTupleType converts the input type T to the constraint type R.
func convertToTupleType[T any, R any](input T) (R, bool) {
	var zero R

	// Direct type match.
	if r, ok := any(input).(R); ok {
		return r, true
	}

	// Handle pointer conversion.
	if _, ok := any(zero).(*[]any); ok {
		if arr, ok := any(input).([]any); ok {
			return any(new(arr)).(R), true
		}
	}

	return zero, false
}

// calculateRequiredCount returns the number of required items.
// Required items are counted from the beginning until all remaining items are optional.
func calculateRequiredCount(items []core.ZodSchema) int {
	lastRequired := -1
	for i := len(items) - 1; i >= 0; i-- {
		if !items[i].Internals().IsOptional() {
			lastRequired = i
			break
		}
	}
	return lastRequired + 1
}

// newZodTupleFromDef constructs a new ZodTuple from a definition.
func newZodTupleFromDef[T any, R any](def *ZodTupleDef, requiredCount int) *ZodTuple[T, R] {
	internals := &ZodTupleInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Items:            def.Items,
		Rest:             def.Rest,
		RequiredCount:    requiredCount,
	}

	// Provide constructor for AddCheck functionality.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		tupleDef := &ZodTupleDef{
			ZodTypeDef: *newDef,
			Items:      def.Items,
			Rest:       def.Rest,
		}
		return any(newZodTupleFromDef[T, R](tupleDef, requiredCount)).(core.ZodType[any])
	}

	schema := &ZodTuple[T, R]{internals: internals}

	if def.Error != nil {
		internals.Error = def.Error
	}

	if len(def.Checks) > 0 {
		internals.Checks = append(internals.Checks, def.Checks...)
	}

	// Ensure validator is called when item schemas exist.
	if len(def.Items) > 0 || def.Rest != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		internals.AddCheck(alwaysPassCheck)
	}

	return schema
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Tuple creates a tuple schema with fixed positional items.
func Tuple(items ...core.ZodSchema) *ZodTuple[[]any, []any] {
	return TupleTyped[[]any, []any](items, nil)
}

// TupleWithRest creates a tuple schema with fixed items and a rest element.
func TupleWithRest(items []core.ZodSchema, rest core.ZodSchema, params ...any) *ZodTuple[[]any, []any] {
	return TupleTyped[[]any, []any](items, rest, params...)
}

// TupleTyped creates a tuple schema with explicit type parameters.
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

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodTupleFromDef[T, R](def, requiredCount)
}

// TuplePtr creates an optional tuple schema that returns *[]any.
func TuplePtr(items ...core.ZodSchema) *ZodTuple[[]any, *[]any] {
	schema := Tuple(items...)
	return schema.Optional()
}
