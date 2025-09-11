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

// Static error variables
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

// Parse validates input using array-specific parsing logic
// Parse validates input using array-specific parsing logic with engine.ParseComplex
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

	// Handle different return types from ParseComplex
	switch v := result.(type) {
	case []any:
		// Direct array type - convert to constraint type R
		return convertArrayFromGeneric[T, R](v), nil
	case *[]any:
		// Pointer to array - convert to constraint type R
		if v == nil {
			var zero R
			return zero, nil
		}
		return convertArrayFromGeneric[T, R](*v), nil
	case nil:
		// Nil result (for optional/nilable cases)
		var zero R
		return zero, nil
	default:
		// Try direct type assertion as fallback
		if typedResult, ok := result.(R); ok {
			return typedResult, nil
		}
		var zero R
		return zero, issues.CreateInvalidTypeError(core.ZodTypeArray, result, ctx[0])
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
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodArray[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	// Convert T to R for ParseComplexStrict
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

	result, err := engine.ParseComplexStrict[[]any, R](
		constraintInput,
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

	return result, nil
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
// This method provides zero-overhead abstraction with strict type constraints.
func (z *ZodArray[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
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

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum number of elements
func (z *ZodArray[T, R]) Min(minLen int, args ...any) *ZodArray[T, R] {
	check := checks.MinLength(minLen, utils.GetFirstParam(args...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of elements
func (z *ZodArray[T, R]) Max(maxLen int, args ...any) *ZodArray[T, R] {
	check := checks.MaxLength(maxLen, utils.GetFirstParam(args...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Length sets exact number of elements
func (z *ZodArray[T, R]) Length(exactLen int, args ...any) *ZodArray[T, R] {
	check := checks.Length(exactLen, utils.GetFirstParam(args...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// NonEmpty requires at least one element
func (z *ZodArray[T, R]) NonEmpty(args ...any) *ZodArray[T, R] {
	return z.Min(1, utils.GetFirstParam(args...))
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

// Transform applies a transformation function using the WrapFn pattern.
// Array types implement direct extraction of T values for transformation.
func (z *ZodArray[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	// WrapFn Pattern: Create wrapper function for type-safe extraction
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		arrayValue := extractArrayValue[T, R](input) // Use existing extraction logic
		return fn(arrayValue, ctx)
	}

	// Use the new factory function for ZodTransform
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodArray[T, R]) Overwrite(transform func(R) R, params ...any) *ZodArray[T, R] {
	// Create a transformation function that works with the constraint type R
	transformAny := func(input any) any {
		// Try to convert input to constraint type R
		converted, ok := convertToArrayType[T, R](input)
		if !ok {
			// If conversion fails, return original value
			return input
		}

		// Apply transformation directly on constraint type R
		return transform(converted)
	}

	check := checks.NewZodCheckOverwrite(transformAny, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using the WrapFn pattern.
// Instead of using adapter structures, this creates a target function that handles type conversion.
//
// WrapFn Implementation:
//  1. Create a target function that extracts T from constraint type R
//  2. Apply the target schema to the extracted T
//  3. Return a ZodPipe with the target function
//
// Zero Redundancy:
//   - No arrayTypeConverter structure needed
//   - Direct function composition eliminates overhead
//   - Type-safe extraction from constraint type R to T
//
// Example:
//
//	arrayToString := Array([]any{String()}).Pipe(String())  // []any -> string conversion
func (z *ZodArray[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	// WrapFn Pattern: Create target function for type conversion and validation
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		// Extract T value from constraint type R
		arrayValue := extractArrayValue[T, R](input)
		// Apply target schema to the extracted T
		return target.Parse(arrayValue, ctx)
	}

	// Use the new factory function for ZodPipe
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation that matches the schema's output type R.
func (z *ZodArray[T, R]) Refine(fn func(R) bool, params ...any) *ZodArray[T, R] {
	// Wrapper converts the raw value (always T or nil) into R before calling fn.
	wrapper := func(v any) bool {
		var zero R

		switch any(zero).(type) {
		case *T:
			// Schema output is *T – convert incoming value (T or nil) to *T
			if v == nil {
				return fn(any((*T)(nil)).(R))
			}
			if arrayVal, ok := v.(T); ok {
				arrayValCopy := arrayVal
				ptr := &arrayValCopy
				return fn(any(ptr).(R))
			}
			return false
		default:
			// Schema output is T
			if v == nil {
				// nil should never reach here for T schema; treat as failure.
				return false
			}
			if arrayVal, ok := v.(T); ok {
				return fn(any(arrayVal).(R))
			}
			return false
		}
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

// RefineAny provides flexible validation without type conversion
func (z *ZodArray[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodArray[T, R] {
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

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodArray instance of constraint type *T.
// Used by modifiers such as Optional, Nilable, and Nullish that must return a pointer constraint.
func (z *ZodArray[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodArray[T, *T] {
	return &ZodArray[T, *T]{internals: &ZodArrayInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
	}}
}

// withInternals creates a new ZodArray instance that keeps the original constraint type R.
// Used by modifiers that retain the original constraint, such as Default, Prefault, and Transform.
func (z *ZodArray[T, R]) withInternals(in *core.ZodTypeInternals) *ZodArray[T, R] {
	return &ZodArray[T, R]{internals: &ZodArrayInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Items:            z.internals.Items,
		Rest:             z.internals.Rest,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodArray[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodArray[T, R]); ok {
		// Preserve original checks to avoid overwriting them
		originalChecks := z.internals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.Checks = originalChecks
	}
}

// convertArrayFromGeneric converts from generic []any to constraint type R
func convertArrayFromGeneric[T any, R any](arrayValue []any) R {
	// Handle direct array assignment
	if directValue, ok := any(arrayValue).(R); ok {
		return directValue
	}

	// Try type conversion for pointer types
	var zero R
	zeroType := reflect.TypeOf((*R)(nil)).Elem()

	// Check if R is a pointer type (like *[]any)
	if zeroType.Kind() == reflect.Ptr {
		// Create pointer to the array
		ptrVal := any(&arrayValue)
		if converted, ok := ptrVal.(R); ok {
			return converted
		}
		// Return nil pointer if conversion fails
		return zero
	}

	// For non-pointer types, try direct conversion
	if converted, ok := any(arrayValue).(R); ok {
		return converted
	}

	// Fallback to zero value if all conversions fail
	return zero
}

// convertToArrayType converts any value to the array constraint type R with strict type checking
func convertToArrayType[T any, R any](
	value any,
) (R, bool) {
	var zero R

	if value == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeOf((*R)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Extract array value from input
	var arrayValue []any
	var isValid bool

	switch val := value.(type) {
	case []any:
		arrayValue, isValid = val, true
	case *[]any:
		if val != nil {
			arrayValue, isValid = *val, true
		}
	default:
		// Try to extract using reflection for other slice types
		if rv := reflect.ValueOf(value); rv.Kind() == reflect.Slice {
			arrayValue = make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				arrayValue[i] = rv.Index(i).Interface()
			}
			isValid = true
		} else {
			return zero, false // Reject all non-array types
		}
	}

	if !isValid {
		return zero, false
	}

	// Convert to target constraint type R
	zeroType := reflect.TypeOf((*R)(nil)).Elem()
	//nolint:exhaustive
	switch zeroType.Kind() {
	case reflect.Slice:
		if reflect.TypeOf(value).AssignableTo(reflect.TypeOf((*R)(nil)).Elem()) {
			return value.(R), true
		}
	case reflect.Ptr:
		// R is *[]any
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

// extractArray converts input to []any array
func (z *ZodArray[T, R]) extractArray(value any) ([]any, error) {
	switch v := value.(type) {
	case []any:
		return v, nil
	case *[]any:
		if v != nil {
			return *v, nil
		}
		return nil, fmt.Errorf("%w", ErrNilPointerToArray)
	default:
		// Try to convert using reflection
		rv := reflect.ValueOf(value)

		// Handle pointer to slice/array
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil, fmt.Errorf("%w", ErrNilPointer)
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil, fmt.Errorf("%w, got %T", ErrExpectedArrayOrSlice, value)
		}

		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result, nil
	}
}

// validateArrayWithIssues validates the array content using raw issues (new approach)
func (z *ZodArray[T, R]) validateArrayWithIssues(value []any, checks []core.ZodCheck, ctx *core.ParseContext) ([]any, error) {
	// Defensive check for nil internals
	if z.internals == nil {
		return nil, issues.CreateInvalidTypeError(core.ZodTypeArray, value, ctx)
	}

	// First apply checks (including Overwrite transformations) to get the transformed value
	transformedValue, err := engine.ApplyChecks[[]any](value, checks, ctx)
	if err != nil {
		return nil, err
	}

	// Use the transformed value for subsequent validation
	value = transformedValue

	fixedLen := len(z.internals.Items)
	actualLen := len(value)
	hasRest := z.internals.Rest != nil

	var collectedIssues []core.ZodRawIssue

	// Validate length first - fail fast on length errors (TypeScript Zod v4 behavior)
	if hasRest {
		// With rest parameter: must have at least fixed length
		if actualLen < fixedLen {
			issue := issues.CreateTooSmallIssue(fixedLen, true, "array", value)
			issue.Properties["is_rest_param"] = true
			collectedIssues = append(collectedIssues, issue)
			// Fail fast - return immediately on length error
			return nil, issues.CreateArrayValidationIssues(collectedIssues)
		}
	} else {
		// Without rest parameter: must match exactly - fail fast on length error
		if actualLen != fixedLen {
			if actualLen < fixedLen {
				issue := issues.CreateFixedLengthArrayIssue(fixedLen, actualLen, value, true)
				collectedIssues = append(collectedIssues, issue)
			} else {
				issue := issues.CreateFixedLengthArrayIssue(fixedLen, actualLen, value, false)
				collectedIssues = append(collectedIssues, issue)
			}
			// Fail fast - return immediately on length error (TypeScript Zod v4 behavior)
			return nil, issues.CreateArrayValidationIssues(collectedIssues)
		}
	}

	// Length is correct, now validate all elements and collect multiple errors
	// Validate fixed part - collect all element errors instead of stopping at first
	for i := 0; i < fixedLen && i < actualLen; i++ {
		if err := z.validateElement(value[i], z.internals.Items[i], ctx, i); err != nil {
			issue := issues.CreateElementValidationIssue(i, "array", value[i], err)
			collectedIssues = append(collectedIssues, issue)
		}
	}

	// Validate rest part - collect all rest element errors
	if hasRest && actualLen > fixedLen {
		for i := fixedLen; i < actualLen; i++ {
			if err := z.validateElement(value[i], z.internals.Rest, ctx, i); err != nil {
				issue := issues.CreateElementValidationIssue(i, "array rest", value[i], err)
				collectedIssues = append(collectedIssues, issue)
			}
		}
	}

	// If we collected any issues, return them as a combined error
	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}

	return value, nil
}

// validateElement validates a single array element against its schema
func (z *ZodArray[T, R]) validateElement(value any, schema any, ctx *core.ParseContext, index int) error {
	if schema == nil {
		return nil
	}

	// Defensive check for context
	if ctx == nil {
		return issues.CreateCustomError("validation context is nil", nil, value, &core.ParseContext{})
	}

	// Try using reflection to call Parse method - this handles all schema types
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

	// Build arguments for Parse call
	args := []reflect.Value{reflect.ValueOf(value)}
	if methodType.NumIn() > 1 && methodType.In(1).String() == "*core.ParseContext" {
		// Add context parameter if expected
		args = append(args, reflect.ValueOf(ctx))
	}

	// Call Parse method
	results := parseMethod.Call(args)
	if len(results) >= 2 {
		// Check if there's an error (second return value)
		if errInterface := results[1].Interface(); errInterface != nil {
			if err, ok := errInterface.(error); ok {
				return err
			}
		}
	}

	return nil
}

// newZodArrayFromDef constructs new ZodArray from definition
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

	// Provide constructor for AddCheck functionality
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

// Array creates tuple schema with fixed elements
// Supports the following patterns:
// Array() - empty tuple
// Array([]any{String(), Int()}) - fixed length tuple
// Array([]any{String(), Int()}, Bool()) - tuple with rest parameter
//
// Graceful handling: Non-[]any arguments are converted to single-element array for convenience
func Array(args ...any) *ZodArray[[]any, []any] {
	// No arguments - create empty array
	if len(args) == 0 {
		return ArrayTyped[[]any, []any]([]any{})
	}

	// First argument must be []any - explicit API format
	if items, ok := args[0].([]any); ok {
		return ArrayTyped[[]any, []any](items, args[1:]...)
	}

	// Graceful handling: treat single non-[]any argument as single-element array
	// This provides better user experience while maintaining API consistency
	return ArrayTyped[[]any, []any]([]any{args[0]}, args[1:]...)
}

// ArrayPtr creates pointer-capable tuple schema
//
// Graceful handling: Non-[]any arguments are converted to single-element array for convenience
func ArrayPtr(args ...any) *ZodArray[[]any, *[]any] {
	// No arguments - create empty array
	if len(args) == 0 {
		return ArrayTyped[[]any, *[]any]([]any{})
	}

	// First argument must be []any - explicit API format
	if items, ok := args[0].([]any); ok {
		return ArrayTyped[[]any, *[]any](items, args[1:]...)
	}

	// Graceful handling: treat single non-[]any argument as single-element array
	// This provides better user experience while maintaining API consistency
	return ArrayTyped[[]any, *[]any]([]any{args[0]}, args[1:]...)
}

// ArrayTyped is the generic constructor for tuple schemas
// Supports explicit syntax only:
// ArrayTyped([]any{schemas...}) - fixed length tuple
// ArrayTyped([]any{schemas...}, params) - fixed length tuple with custom params
// ArrayTyped([]any{schemas...}, Rest) - tuple with rest parameter
// ArrayTyped([]any{schemas...}, Rest, params) - with custom params
func ArrayTyped[T any, R any](items []any, args ...any) *ZodArray[T, R] {
	var Rest any
	var param any

	// Parse remaining arguments
	for _, arg := range args {
		switch v := arg.(type) {
		case core.SchemaParams:
			param = v
		default:
			// First non-params argument is rest schema
			if Rest == nil {
				Rest = v
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
		Rest:  Rest,
	}

	// Apply normalized parameters to schema definition
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodArrayFromDef[T, R](def)
}

// Check adds a custom validation function that can report multiple issues for array schema.
func (z *ZodArray[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodArray[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// If R is a value type and the value in payload is the corresponding value type, perform automatic value adaptation
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// If R is a pointer type and the value in payload is the corresponding value type, perform automatic pointer adaptation
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

// extractArrayForEngine extracts []any from input for engine.ParseComplex
func (z *ZodArray[T, R]) extractArrayForEngine(input any) ([]any, bool) {
	result, err := z.extractArray(input)
	if err != nil {
		return nil, false
	}
	return result, true
}

// extractArrayPtrForEngine extracts pointer to []any from input for engine.ParseComplex
func (z *ZodArray[T, R]) extractArrayPtrForEngine(input any) (*[]any, bool) {
	// Try direct pointer extraction
	if ptr, ok := input.(*[]any); ok {
		return ptr, true
	}

	// Try extracting array and return pointer to it
	result, err := z.extractArray(input)
	if err != nil {
		return nil, false
	}
	return &result, true
}

// validateArrayForEngine validates []any for engine.ParseComplex
func (z *ZodArray[T, R]) validateArrayForEngine(value []any, checks []core.ZodCheck, ctx *core.ParseContext) ([]any, error) {
	return z.validateArrayWithIssues(value, checks, ctx)
}
