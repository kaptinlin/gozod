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
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// Static error variables
var (
	ErrParseComplexUnexpectedType = errors.New("internal error: ParseComplex returned unexpected type")
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodSliceDef defines the schema definition for slice validation
type ZodSliceDef struct {
	core.ZodTypeDef
	Element any // The element schema (type-erased for flexibility)
}

// ZodSliceInternals contains the internal state for slice schema
type ZodSliceInternals[T any] struct {
	core.ZodTypeInternals
	Def     *ZodSliceDef // Schema definition reference
	Element any          // Element schema for runtime validation
}

// ZodSlice represents a type-safe slice validation schema with dual generic parameters
// T: element type, R: constraint type ([]T or *[]T)
type ZodSlice[T any, R any] struct {
	internals *ZodSliceInternals[T]
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodSlice[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodSlice[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodSlice[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input using slice-specific parsing logic with type safety
func (z *ZodSlice[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplex[[]T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSlice,
		z.extractSlice,
		z.extractSlicePtr,
		z.validateSlice,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	// Handle different return types from ParseType
	switch v := result.(type) {
	case []T:
		// Direct slice type - convert to constraint type R
		return convertToSliceConstraintType[T, R](v), nil
	case *[]T:
		// Pointer to slice - convert to constraint type R
		return convertToSliceConstraintType[T, R](v), nil
	case **[]T:
		// Double pointer (from pointer extraction) - unwrap once
		if v != nil {
			return convertToSliceConstraintType[T, R](*v), nil
		}
		return convertToSliceConstraintType[T, R](nil), nil
	case nil:
		// Nil result (for optional/nilable cases)
		return convertToSliceConstraintType[T, R](nil), nil
	default:
		// Try direct type assertion as fallback
		if typedResult, ok := result.(R); ok {
			return typedResult, nil
		}
		var zero R
		return zero, fmt.Errorf("%w %T", ErrParseComplexUnexpectedType, result)
	}
}

// MustParse validates the input value and panics on failure
func (z *ZodSlice[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodSlice[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	// Convert T to R for ParseComplexStrict
	constraintInput := convertToSliceConstraintType[T, R](input)

	result, err := engine.ParseComplexStrict[[]T, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSlice,
		z.extractSlice,
		z.extractSlicePtr,
		z.validateSlice,
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
func (z *ZodSlice[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodSlice[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates optional slice schema (always returns pointer constraint)
func (z *ZodSlice[T, R]) Optional() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values (always returns pointer constraint)
func (z *ZodSlice[T, R]) Nilable() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers (always returns pointer constraint)
func (z *ZodSlice[T, R]) Nullish() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil slice constraint ([]T).
func (z *ZodSlice[T, R]) NonOptional() *ZodSlice[T, []T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodSlice[T, []T]{
		internals: &ZodSliceInternals[T]{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Element:          z.internals.Element,
		},
	}
}

// Default preserves current constraint type
func (z *ZodSlice[T, R]) Default(v []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type
func (z *ZodSlice[T, R]) DefaultFunc(fn func() []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure (preserves constraint type)
func (z *ZodSlice[T, R]) Prefault(v []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values (preserves constraint type)
func (z *ZodSlice[T, R]) PrefaultFunc(fn func() []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this slice schema in the global registry.
func (z *ZodSlice[T, R]) Meta(meta core.GlobalMeta) *ZodSlice[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodSlice[T, R]) Describe(description string) *ZodSlice[T, R] {
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
func (z *ZodSlice[T, R]) Min(minLen int, params ...any) *ZodSlice[T, R] {
	check := checks.MinSize(minLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of elements
func (z *ZodSlice[T, R]) Max(maxLen int, params ...any) *ZodSlice[T, R] {
	check := checks.MaxSize(maxLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Length sets exact number of elements
func (z *ZodSlice[T, R]) Length(exactLen int, params ...any) *ZodSlice[T, R] {
	check := checks.Size(exactLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// NonEmpty ensures slice has at least one element
func (z *ZodSlice[T, R]) NonEmpty(params ...any) *ZodSlice[T, R] {
	check := checks.MinSize(1, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Element returns the element schema for this slice
func (z *ZodSlice[T, R]) Element() any {
	return z.internals.Element
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function using the WrapFn pattern.
// Slice types pass the constraint type R directly to transformation functions.
func (z *ZodSlice[T, R]) Transform(fn func(R, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	// WrapFn Pattern: Create wrapper function - no extraction needed, pass R directly
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(input, ctx) // Pass constraint type R directly to user function
	}

	// Use the new factory function for ZodTransform
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms slice value while keeping the same type
func (z *ZodSlice[T, R]) Overwrite(transform func(R) R, params ...any) *ZodSlice[T, R] {
	check := checks.NewZodCheckOverwrite(func(input any) any {
		// Convert input to the correct constraint type
		if converted, ok := convertToSliceType[T, R](input); ok {
			return transform(converted)
		}
		return input
	}, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using the WrapFn pattern.
// Instead of using adapter structures, this creates a target function that handles type conversion.
func (z *ZodSlice[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	// WrapFn Pattern: Create target function - no extraction needed, pass R directly
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		// Apply target schema to constraint type R directly
		return target.Parse(input, ctx)
	}

	// Use the new factory function for ZodPipe
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation with automatic constraint type matching
func (z *ZodSlice[T, R]) Refine(fn func(R) bool, params ...any) *ZodSlice[T, R] {
	// Type-safe wrapper that handles both []T and *[]T constraint types
	wrapper := func(value any) bool {
		// Convert input to constraint type R and call user function
		constraintValue := convertToSliceConstraintType[T, R](value)
		return fn(constraintValue)
	}

	param := utils.GetFirstParam(params...)
	customParams := utils.NormalizeCustomParams(param)
	check := checks.NewCustom[any](wrapper, customParams)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodSlice[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodSlice[T, R] {
	param := utils.GetFirstParam(params...)
	customParams := utils.NormalizeCustomParams(param)
	check := checks.NewCustom[any](fn, customParams)
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
//	schema := gozod.Slice[string](gozod.String()).Min(1).And(gozod.Slice[string](gozod.String()).Max(10))
//	result, _ := schema.Parse([]string{"a", "b"}) // Must satisfy both constraints
func (z *ZodSlice[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
// Enables chaining: schema.Or(other).Or(another)
// TypeScript Zod v4 equivalent: schema.or(other)
//
// Example:
//
//	schema := gozod.Slice[string](gozod.String()).Or(gozod.Slice[int](gozod.Int()))
func (z *ZodSlice[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withInternals creates new instance preserving current constraint type
func (z *ZodSlice[T, R]) withInternals(in *core.ZodTypeInternals) *ZodSlice[T, R] {
	return &ZodSlice[T, R]{internals: &ZodSliceInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Element:          z.internals.Element,
	}}
}

// withPtrInternals creates new instance with pointer constraint type
func (z *ZodSlice[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodSlice[T, *[]T] {
	return &ZodSlice[T, *[]T]{internals: &ZodSliceInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Element:          z.internals.Element,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodSlice[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodSlice[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractSlice extracts []T from value with proper conversion
func (z *ZodSlice[T, R]) extractSlice(value any) ([]T, bool) {
	// Handle direct []T
	if sliceVal, ok := value.([]T); ok {
		return sliceVal, true
	}

	// Handle []any to []T conversion
	if anySlice, ok := value.([]any); ok {
		converted := make([]T, len(anySlice))
		for i, elem := range anySlice {
			if typedElem, ok := elem.(T); ok {
				converted[i] = typedElem
			} else {
				// Type conversion failed
				return nil, false
			}
		}
		return converted, true
	}

	// Try to convert using slicex for generic slice types
	if value != nil {
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			if converted, err := slicex.ToAny(value); err == nil && converted != nil {
				// Convert []any to []T
				typedSlice := make([]T, len(converted))
				for i, elem := range converted {
					if typedElem, ok := elem.(T); ok {
						typedSlice[i] = typedElem
					} else {
						return nil, false
					}
				}
				return typedSlice, true
			}
		}
	}

	return nil, false
}

// extractSlicePtr extracts *[]T from value
func (z *ZodSlice[T, R]) extractSlicePtr(value any) (*[]T, bool) {
	// Handle direct *[]T - CRITICAL: preserve original pointer identity
	if sliceVal, ok := value.(*[]T); ok {
		return sliceVal, true
	}

	// Handle []T - we cannot preserve pointer identity here since input is not a pointer
	// This is only used when the input is a slice value, not a pointer to slice
	if sliceVal, ok := value.([]T); ok {
		ptrSlice := &sliceVal
		return ptrSlice, true
	}

	return nil, false
}

// validateSlice validates slice elements using element schema with multiple error collection (TypeScript Zod v4 array behavior)
func (z *ZodSlice[T, R]) validateSlice(value []T, checks []core.ZodCheck, ctx *core.ParseContext) ([]T, error) {
	var collectedIssues []core.ZodRawIssue

	// Apply slice-level checks (length, size, Overwrite, etc.) and collect any issues
	// Unlike tuples, arrays continue validation even with size constraint errors
	payload := core.NewParsePayload(value)
	result := engine.RunChecksOnValue(value, checks, payload, ctx)

	// Collect any slice-level issues (length, size, etc.)
	if result.HasIssues() {
		sliceIssues := result.GetIssues()
		collectedIssues = append(collectedIssues, sliceIssues...)
	}

	// Get the potentially transformed value for element validation
	var validatedValue []T
	if result.GetValue() != nil {
		if converted, ok := result.GetValue().([]T); ok {
			validatedValue = converted
		} else {
			// Type conversion failed, but continue with original value for element validation
			validatedValue = value
		}
	} else {
		validatedValue = value
	}

	// Collect all element validation errors (TypeScript Zod v4 array behavior)
	// Always validate each element if element schema is provided - this is what makes Slice different from []any
	if z.internals.Element != nil {
		for i, element := range validatedValue {
			// Directly validate element schema without wrapping in CreateInvalidElementError
			if err := z.validateElement(element, z.internals.Element, ctx); err != nil {
				// Convert element error to raw issue and add path prefix (TypeScript Zod v4 array behavior)
				var zodErr *issues.ZodError
				if errors.As(err, &zodErr) {
					// Propagate all issues from element validation with path prefix
					for _, elementIssue := range zodErr.Issues {
						// Create raw issue preserving original code and essential properties
						rawIssue := core.ZodRawIssue{
							Code:       elementIssue.Code,
							Message:    elementIssue.Message,
							Input:      elementIssue.Input,
							Path:       []any{i}, // Set path to array index only
							Properties: make(map[string]any),
						}
						// Copy essential properties from ZodIssue to ZodRawIssue
						if elementIssue.Minimum != nil {
							rawIssue.Properties["minimum"] = elementIssue.Minimum
						}
						if elementIssue.Maximum != nil {
							rawIssue.Properties["maximum"] = elementIssue.Maximum
						}
						if elementIssue.Expected != "" {
							rawIssue.Properties["expected"] = elementIssue.Expected
						}
						if elementIssue.Received != "" {
							rawIssue.Properties["received"] = elementIssue.Received
						}
						rawIssue.Properties["inclusive"] = elementIssue.Inclusive
						collectedIssues = append(collectedIssues, rawIssue)
					}
				} else {
					// Handle non-ZodError by creating a raw issue with path
					rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, element)
					rawIssue.Path = []any{i}
					collectedIssues = append(collectedIssues, rawIssue)
				}
			}
		}
	}

	// If we collected any issues (slice-level or element-level), return them as a combined error
	if len(collectedIssues) > 0 {
		var zero []T
		return zero, issues.CreateArrayValidationIssues(collectedIssues)
	}

	// Return the validated and potentially transformed value
	return validatedValue, nil
}

// validateElement validates a single element using the provided schema (without wrapping)
func (z *ZodSlice[T, R]) validateElement(element T, schema any, ctx *core.ParseContext) error {
	if schema == nil {
		return nil
	}

	// Special handling for Lazy schemas - DON'T unwrap, let Parse handle it
	// The Lazy Parse method will properly evaluate the inner schema
	// This ensures validation messages are correct and lazy evaluation works properly

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
	args := []reflect.Value{reflect.ValueOf(element)}
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
				return err // Return the error directly without wrapping
			}
		}
	}

	return nil
}

// convertToSliceType converts any value to the specified slice constraint type with strict type checking
func convertToSliceType[T any, R any](v any) (R, bool) {
	var zero R

	// Try to convert to []T first
	if slice, ok := v.([]T); ok {
		return convertToSliceConstraintType[T, R](slice), true
	}

	// Try to convert to *[]T
	if ptrSlice, ok := v.(*[]T); ok {
		return convertToSliceConstraintType[T, R](ptrSlice), true
	}

	// Use reflection as fallback for complex type conversions
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		// Try to convert each element
		elemType := reflect.TypeOf((*T)(nil)).Elem()
		slice := make([]T, rv.Len())
		for i := range rv.Len() {
			elem := rv.Index(i).Interface()
			if reflect.TypeOf(elem).ConvertibleTo(elemType) {
				slice[i] = reflect.ValueOf(elem).Convert(elemType).Interface().(T)
			} else {
				return zero, false
			}
		}
		return convertToSliceConstraintType[T, R](slice), true
	}

	return zero, false
}

// convertToSliceConstraintType converts values to the constraint type R
func convertToSliceConstraintType[T any, R any](value any) R {
	var zero R
	zeroType := reflect.TypeOf(zero)

	// Handle nil case
	if value == nil {
		return zero
	}

	// Check if R is a pointer type (*[]T)
	if zeroType != nil && zeroType.Kind() == reflect.Ptr {
		// R is *[]T
		switch v := value.(type) {
		case []T:
			// Convert []T to *[]T
			ptr := &v
			return any(ptr).(R)
		case *[]T:
			// Already *[]T
			return any(v).(R)
		}
	} else {
		// R is []T
		switch v := value.(type) {
		case []T:
			// Already []T
			return any(v).(R)
		case *[]T:
			// Convert *[]T to []T
			if v != nil {
				return any(*v).(R)
			}
		}
	}

	// Fallback: attempt direct conversion
	if converted, ok := value.(R); ok {
		return converted
	}

	return zero
}

// newZodSliceFromDef constructs new ZodSlice from definition
func newZodSliceFromDef[T any, R any](def *ZodSliceDef) *ZodSlice[T, R] {
	internals := &ZodSliceInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Element:          def.Element,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		sliceDef := &ZodSliceDef{
			ZodTypeDef: *newDef,
			Element:    def.Element,
		}
		return any(newZodSliceFromDef[T, R](sliceDef)).(core.ZodType[any])
	}

	schema := &ZodSlice[T, R]{internals: internals}

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
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Slice creates slice schema with element validation (returns value constraint)
func Slice[T any](elementSchema any, paramArgs ...any) *ZodSlice[T, []T] {
	return SliceTyped[T, []T](elementSchema, paramArgs...)
}

// SlicePtr creates slice schema with pointer constraint (returns pointer constraint)
func SlicePtr[T any](elementSchema any, paramArgs ...any) *ZodSlice[T, *[]T] {
	return SliceTyped[T, *[]T](elementSchema, paramArgs...)
}

// SliceTyped creates slice schema with explicit constraint type (universal constructor)
func SliceTyped[T any, R any](elementSchema any, paramArgs ...any) *ZodSlice[T, R] {
	param := utils.GetFirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodSliceDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeSlice,
			Checks: []core.ZodCheck{},
		},
		Element: elementSchema,
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	sliceSchema := newZodSliceFromDef[T, R](def)

	// Ensure validator is called when element schema exists
	// Add a minimal check that always passes to trigger validation
	if elementSchema != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		sliceSchema.internals.AddCheck(alwaysPassCheck)
	}

	return sliceSchema
}

// Check adds a custom validation function for slice schema that can push multiple issues.
func (z *ZodSlice[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSlice[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// direct assertion
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// handle pointer/value mismatch
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

// With is an alias for Check - adds a custom validation function.
// TypeScript Zod v4 equivalent: schema.with(...)
func (z *ZodSlice[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSlice[T, R] {
	return z.Check(fn, params...)
}
