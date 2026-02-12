package types

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

var (
	ErrNilPointerToArray    = errors.New("nil pointer to array")
	ErrNilPointer           = errors.New("nil pointer")
	ErrExpectedArrayOrSlice = errors.New("expected array or slice")
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodArrayDef defines the schema definition for fixed-length array validation
type ZodArrayDef struct {
	core.ZodTypeDef
	Items []any // Element schemas for each position (type-erased for flexibility)
	Rest  any   // Rest schema for variadic elements (nil if no rest)
}

// ZodArrayInternals contains the internal state for array schema
type ZodArrayInternals struct {
	core.ZodTypeInternals
	Def   *ZodArrayDef // Schema definition reference
	Items []any        // Element schemas for runtime validation
	Rest  any          // Rest schema for variadic elements
}

// ZodArray represents a type-safe fixed-length array validation schema with unified constraint
// T is the base array type ([]any), R is the constraint type ([]any | *[]any)
type ZodArray[T any, R any] struct {
	internals *ZodArrayInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodArray[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodArray[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodArray[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input using array-specific parsing logic.
func (z *ZodArray[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplex[[]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeArray,
		z.extractArrayForEngine,
		z.extractArrayPtrForEngine,
		z.validateArrayForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	switch v := result.(type) {
	case []any:
		return convertArrayFromGeneric[T, R](v), nil
	case *[]any:
		if v == nil {
			var zero R
			return zero, nil
		}
		return convertArrayFromGeneric[T, R](*v), nil
	case nil:
		var zero R
		return zero, nil
	default:
		if typedResult, ok := result.(R); ok {
			return typedResult, nil
		}
		var zero R
		parseCtx := core.NewParseContext()
		if len(ctx) > 0 && ctx[0] != nil {
			parseCtx = ctx[0]
		}
		return zero, issues.CreateInvalidTypeError(core.ZodTypeArray, result, parseCtx)
	}
}

// MustParse validates the input value and panics on failure
func (z *ZodArray[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
func (z *ZodArray[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	constraintInput, ok := convertToArrayType[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"array constraint type",
			any(input),
			ctx[0],
		)
	}

	return engine.ParseComplexStrict[[]any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeArray,
		z.extractArrayForEngine,
		z.extractArrayPtrForEngine,
		z.validateArrayForEngine,
		ctx...,
	)
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
func (z *ZodArray[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result.
func (z *ZodArray[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *[]any constraint because the optional value may be nil.
func (z *ZodArray[T, R]) Optional() *ZodArray[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
// Unlike Optional(), which accepts both absent keys AND nil values,
// ExactOptional() only accepts absent keys in object fields.
func (z *ZodArray[T, R]) ExactOptional() *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable always returns *[]any constraint because the value may be nil.
func (z *ZodArray[T, R]) Nilable() *ZodArray[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility
func (z *ZodArray[T, R]) Nullish() *ZodArray[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil value constraint (T).
// This mirrors the behaviour of Optional().NonOptional() in TS Zod, and
// produces dedicated "expected = nonoptional" error when input is nil.
func (z *ZodArray[T, R]) NonOptional() *ZodArray[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodArray[T, T]{
		internals: &ZodArrayInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Items:            z.internals.Items,
			Rest:             z.internals.Rest,
		},
	}
}

// Default keeps the current generic constraint type R.
func (z *ZodArray[T, R]) Default(v T) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic constraint type R.
func (z *ZodArray[T, R]) DefaultFunc(fn func() T) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodArray[T, R]) Prefault(v T) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodArray[T, R]) PrefaultFunc(fn func() T) *ZodArray[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this array schema in the global registry.
func (z *ZodArray[T, R]) Meta(meta core.GlobalMeta) *ZodArray[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodArray[T, R]) Describe(description string) *ZodArray[T, R] {
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

// Min sets minimum number of elements
func (z *ZodArray[T, R]) Min(minLen int, args ...any) *ZodArray[T, R] {
	check := checks.MinLength(minLen, utils.FirstParam(args...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of elements
func (z *ZodArray[T, R]) Max(maxLen int, args ...any) *ZodArray[T, R] {
	check := checks.MaxLength(maxLen, utils.FirstParam(args...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Length sets exact number of elements
func (z *ZodArray[T, R]) Length(exactLen int, args ...any) *ZodArray[T, R] {
	check := checks.Length(exactLen, utils.FirstParam(args...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// NonEmpty requires at least one element
func (z *ZodArray[T, R]) NonEmpty(args ...any) *ZodArray[T, R] {
	return z.Min(1, utils.FirstParam(args...))
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Element returns the schema for the element at the given index
func (z *ZodArray[T, R]) Element(index int) any {
	if index >= 0 && index < len(z.internals.Items) {
		return z.internals.Items[index]
	}
	return nil
}

// Items returns all element schemas
func (z *ZodArray[T, R]) Items() []any {
	result := make([]any, len(z.internals.Items))
	copy(result, z.internals.Items)
	return result
}

// Rest returns the rest parameter schema
func (z *ZodArray[T, R]) Rest() any {
	return z.internals.Rest
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed array value.
func (z *ZodArray[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractArrayValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodArray[T, R]) Overwrite(transform func(R) R, params ...any) *ZodArray[T, R] {
	transformAny := func(input any) any {
		converted, ok := convertToArrayType[T, R](input)
		if !ok {
			return input
		}
		return transform(converted)
	}

	check := checks.NewZodCheckOverwrite(transformAny, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline that passes the parsed value to a target schema.
//
// Example:
//
//	arrayToString := Array([]any{String()}).Pipe(String())
func (z *ZodArray[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractArrayValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation that matches the schema's output type R.
func (z *ZodArray[T, R]) Refine(fn func(R) bool, params ...any) *ZodArray[T, R] {
	wrapper := func(v any) bool {
		var zero R

		switch any(zero).(type) {
		case *T:
			if v == nil {
				return fn(any((*T)(nil)).(R))
			}
			if arrayVal, ok := v.(T); ok {
				arrayValCopy := arrayVal
				return fn(any(&arrayValCopy).(R))
			}
			return false
		default:
			if v == nil {
				return false
			}
			if arrayVal, ok := v.(T); ok {
				return fn(any(arrayVal).(R))
			}
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)

	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion.
func (z *ZodArray[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodArray[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

// And creates an intersection with another schema.
// Enables chaining: schema.And(other).And(another)
// TypeScript Zod v4 equivalent: schema.and(other)
//
// Example:
//
//	schema := gozod.Array(gozod.String()).Min(1).And(gozod.Array(gozod.String()).Max(10))
//	result, _ := schema.Parse([]string{"a", "b"}) // Must satisfy both constraints
func (z *ZodArray[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
// Enables chaining: schema.Or(other).Or(another)
// TypeScript Zod v4 equivalent: schema.or(other)
//
// Example:
//
//	schema := gozod.Array(gozod.String()).Or(gozod.Array(gozod.Int()))
//	result, _ := schema.Parse([]string{"a"})  // Accepts string array
//	result, _ = schema.Parse([]int{1, 2})     // Accepts int array
func (z *ZodArray[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodArray with pointer constraint type *T.
func (z *ZodArray[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodArray[T, *T] {
	return &ZodArray[T, *T]{internals: &ZodArrayInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
	}}
}

// withInternals creates a new ZodArray keeping the original constraint type R.
func (z *ZodArray[T, R]) withInternals(in *core.ZodTypeInternals) *ZodArray[T, R] {
	return &ZodArray[T, R]{internals: &ZodArrayInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodArray[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodArray[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// convertArrayFromGeneric converts from generic []any to constraint type R.
func convertArrayFromGeneric[T any, R any](arrayValue []any) R {
	if directValue, ok := any(arrayValue).(R); ok {
		return directValue
	}

	// Pointer types (e.g. *[]any): wrap in pointer.
	var zero R
	if reflect.TypeOf((*R)(nil)).Elem().Kind() == reflect.Ptr {
		if converted, ok := any(&arrayValue).(R); ok {
			return converted
		}
	}

	return zero
}

// convertToArrayType converts any value to the array constraint type R.
func convertToArrayType[T any, R any](value any) (R, bool) {
	var zero R

	if value == nil {
		if reflect.TypeOf((*R)(nil)).Elem().Kind() == reflect.Ptr {
			return zero, true
		}
		return zero, false
	}

	var arrayValue []any

	switch val := value.(type) {
	case []any:
		arrayValue = val
	case *[]any:
		if val == nil {
			return zero, false
		}
		arrayValue = *val
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice {
			return zero, false
		}
		arrayValue = make([]any, rv.Len())
		for i := range rv.Len() {
			arrayValue[i] = rv.Index(i).Interface()
		}
	}

	// Convert to target constraint type R.
	zeroType := reflect.TypeOf((*R)(nil)).Elem()
	//nolint:exhaustive
	switch zeroType.Kind() {
	case reflect.Slice:
		if reflect.TypeOf(value).AssignableTo(zeroType) {
			return value.(R), true
		}
	case reflect.Ptr:
		if converted, ok := any(&arrayValue).(R); ok {
			return converted, true
		}
	}

	return zero, false
}

// extractArrayValue extracts the base array value T from constraint type R
func extractArrayValue[T any, R any](value R) T {
	if ptr, ok := any(value).(*T); ok {
		if ptr != nil {
			return *ptr
		}
		var zero T
		return zero
	}
	return any(value).(T)
}

// extractArray converts input to []any.
func (z *ZodArray[T, R]) extractArray(value any) ([]any, error) {
	switch v := value.(type) {
	case []any:
		return v, nil
	case *[]any:
		if v != nil {
			return *v, nil
		}
		return nil, ErrNilPointerToArray
	default:
		rv := reflect.ValueOf(value)

		// Handle pointer to slice/array.
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil, ErrNilPointer
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil, fmt.Errorf("%w: got %T", ErrExpectedArrayOrSlice, value)
		}

		result := make([]any, rv.Len())
		for i := range rv.Len() {
			result[i] = rv.Index(i).Interface()
		}
		return result, nil
	}
}

// validateArrayWithIssues validates array content and collects all issues.
func (z *ZodArray[T, R]) validateArrayWithIssues(value []any, checks []core.ZodCheck, ctx *core.ParseContext) ([]any, error) {
	if z.internals == nil {
		return nil, issues.CreateInvalidTypeError(core.ZodTypeArray, value, ctx)
	}

	// Apply checks (including Overwrite transformations) first.
	value, err := engine.ApplyChecks[[]any](value, checks, ctx)
	if err != nil {
		return nil, err
	}

	fixedLen := len(z.internals.Items)
	actualLen := len(value)
	hasRest := z.internals.Rest != nil

	// Fail fast on length errors (TypeScript Zod v4 behavior).
	if hasRest {
		if actualLen < fixedLen {
			issue := issues.CreateTooSmallIssue(fixedLen, true, "array", value)
			issue.Properties["is_rest_param"] = true
			return nil, issues.CreateArrayValidationIssues([]core.ZodRawIssue{issue})
		}
	} else if actualLen != fixedLen {
		tooSmall := actualLen < fixedLen
		issue := issues.CreateFixedLengthArrayIssue(fixedLen, actualLen, value, tooSmall)
		return nil, issues.CreateArrayValidationIssues([]core.ZodRawIssue{issue})
	}

	// Length is correct — validate all elements and collect errors.
	var collectedIssues []core.ZodRawIssue

	for i := range min(fixedLen, actualLen) {
		if err := z.validateElement(value[i], z.internals.Items[i], ctx, i); err != nil {
			collectedIssues = append(collectedIssues, issues.CreateElementValidationIssue(i, "array", value[i], err))
		}
	}

	if hasRest && actualLen > fixedLen {
		for i := fixedLen; i < actualLen; i++ {
			if err := z.validateElement(value[i], z.internals.Rest, ctx, i); err != nil {
				collectedIssues = append(collectedIssues, issues.CreateElementValidationIssue(i, "array rest", value[i], err))
			}
		}
	}

	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}

	return value, nil
}

// validateElement validates a single array element against its schema.
func (z *ZodArray[T, R]) validateElement(value any, schema any, _ *core.ParseContext, _ int) error {
	if schema == nil {
		return nil
	}

	schemaValue := reflect.ValueOf(schema)
	if !schemaValue.IsValid() || schemaValue.IsNil() {
		return nil
	}

	parseMethod := schemaValue.MethodByName("Parse")
	if !parseMethod.IsValid() {
		return nil
	}

	methodType := parseMethod.Type()
	if methodType.NumIn() < 1 {
		return nil
	}

	args := []reflect.Value{reflect.ValueOf(value)}

	results := parseMethod.Call(args)
	if len(results) >= 2 {
		if errVal := results[1].Interface(); errVal != nil {
			if err, ok := errVal.(error); ok {
				return err
			}
		}
	}

	return nil
}

// newZodArrayFromDef constructs a new ZodArray from a definition.
func newZodArrayFromDef[T any, R any](def *ZodArrayDef) *ZodArray[T, R] {
	internals := &ZodArrayInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:   def,
		Items: def.Items,
		Rest:  def.Rest,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		arrayDef := &ZodArrayDef{
			ZodTypeDef: *newDef,
			Items:      def.Items,
		}
		return any(newZodArrayFromDef[T, R](arrayDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodArray[T, R]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Array creates a tuple schema with fixed elements.
//
//	Array()                                    - empty tuple
//	Array([]any{String(), Int()})              - fixed length tuple
//	Array([]any{String(), Int()}, Bool())      - tuple with rest parameter
func Array(args ...any) *ZodArray[[]any, []any] {
	if len(args) == 0 {
		return ArrayTyped[[]any, []any]([]any{})
	}

	if items, ok := args[0].([]any); ok {
		return ArrayTyped[[]any, []any](items, args[1:]...)
	}

	// Graceful handling: treat single non-[]any argument as single-element array.
	return ArrayTyped[[]any, []any]([]any{args[0]}, args[1:]...)
}

// ArrayPtr creates a pointer-capable tuple schema.
func ArrayPtr(args ...any) *ZodArray[[]any, *[]any] {
	if len(args) == 0 {
		return ArrayTyped[[]any, *[]any]([]any{})
	}

	if items, ok := args[0].([]any); ok {
		return ArrayTyped[[]any, *[]any](items, args[1:]...)
	}

	return ArrayTyped[[]any, *[]any]([]any{args[0]}, args[1:]...)
}

// ArrayTyped is the generic constructor for tuple schemas.
func ArrayTyped[T any, R any](items []any, args ...any) *ZodArray[T, R] {
	var rest any
	var param any

	for _, arg := range args {
		switch v := arg.(type) {
		case core.SchemaParams:
			param = v
		default:
			if rest == nil {
				rest = v
			}
		}
	}

	normalizedParams := utils.NormalizeParams(param)

	def := &ZodArrayDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeArray,
			Checks: []core.ZodCheck{},
		},
		Items: items,
		Rest:  rest,
	}

	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodArrayFromDef[T, R](def)
}

// Check adds a custom validation function that can report multiple issues.
func (z *ZodArray[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodArray[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Pointer type adaptation: wrap value type in pointer for R = *T.
		var zero R
		zeroTyp := reflect.TypeOf(zero)
		if zeroTyp != nil && zeroTyp.Kind() == reflect.Ptr {
			elemTyp := zeroTyp.Elem()
			valRV := reflect.ValueOf(payload.GetValue())
			if valRV.IsValid() && valRV.Type() == elemTyp {
				ptr := reflect.New(elemTyp)
				ptr.Elem().Set(valRV)
				if casted, ok := ptr.Interface().(R); ok {
					fn(casted, payload)
				}
			}
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// extractArrayForEngine extracts []any from input for engine.ParseComplex.
func (z *ZodArray[T, R]) extractArrayForEngine(input any) ([]any, bool) {
	result, err := z.extractArray(input)
	if err != nil {
		return nil, false
	}
	return result, true
}

// extractArrayPtrForEngine extracts pointer to []any from input.
func (z *ZodArray[T, R]) extractArrayPtrForEngine(input any) (*[]any, bool) {
	if ptr, ok := input.(*[]any); ok {
		return ptr, true
	}

	result, err := z.extractArray(input)
	if err != nil {
		return nil, false
	}
	return &result, true
}

// validateArrayForEngine validates []any for engine.ParseComplex.
func (z *ZodArray[T, R]) validateArrayForEngine(value []any, checks []core.ZodCheck, ctx *core.ParseContext) ([]any, error) {
	return z.validateArrayWithIssues(value, checks, ctx)
}
