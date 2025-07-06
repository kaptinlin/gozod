package types

import (
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/slicex"
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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodSlice[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
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
		return zero, fmt.Errorf("internal error: ParseComplex returned unexpected type %T", result)
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values (always returns pointer constraint)
func (z *ZodSlice[T, R]) Nilable() *ZodSlice[T, *[]T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers (always returns pointer constraint)
func (z *ZodSlice[T, R]) Nullish() *ZodSlice[T, *[]T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil slice constraint ([]T).
func (z *ZodSlice[T, R]) NonOptional() *ZodSlice[T, []T] {
	in := z.internals.ZodTypeInternals.Clone()
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type
func (z *ZodSlice[T, R]) DefaultFunc(fn func() []T) *ZodSlice[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure (preserves constraint type)
func (z *ZodSlice[T, R]) Prefault(v []T) *ZodSlice[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values (preserves constraint type)
func (z *ZodSlice[T, R]) PrefaultFunc(fn func() []T) *ZodSlice[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum number of elements
func (z *ZodSlice[T, R]) Min(minLen int, params ...any) *ZodSlice[T, R] {
	check := checks.MinSize(minLen, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of elements
func (z *ZodSlice[T, R]) Max(maxLen int, params ...any) *ZodSlice[T, R] {
	check := checks.MaxSize(maxLen, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Length sets exact number of elements
func (z *ZodSlice[T, R]) Length(exactLen int, params ...any) *ZodSlice[T, R] {
	check := checks.Size(exactLen, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// NonEmpty ensures slice has at least one element
func (z *ZodSlice[T, R]) NonEmpty(params ...any) *ZodSlice[T, R] {
	check := checks.MinSize(1, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	return core.NewZodPipe[R, any](z, targetFn)
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

	checkParams := checks.NormalizeCheckParams(params...)
	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodSlice[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodSlice[T, R] {
	checkParams := checks.NormalizeCheckParams(params...)
	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
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
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
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

// validateSlice validates slice elements using element schema
func (z *ZodSlice[T, R]) validateSlice(value []T, checks []core.ZodCheck, ctx *core.ParseContext) ([]T, error) {
	// First validate the slice itself using standard checks (including Overwrite transformations)
	validatedValue, err := engine.ApplyChecks[[]T](value, checks, ctx)
	if err != nil {
		var zero []T
		return zero, err
	}

	// Use the validated (potentially transformed) value for element validation
	// Always validate each element if element schema is provided - this is what makes Slice different from []any
	if z.internals.Element != nil {
		for i, element := range validatedValue {
			if err := z.validateElement(element, z.internals.Element, ctx, i); err != nil {
				var zero []T
				return zero, err
			}
		}
	}

	// Return the validated and potentially transformed value
	return validatedValue, nil
}

// validateElement validates a single element using the provided schema
func (z *ZodSlice[T, R]) validateElement(element T, schema any, ctx *core.ParseContext, index int) error {
	if schema == nil {
		return nil
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
				return fmt.Errorf("element %d validation failed: %w", index, err)
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
		for i := 0; i < rv.Len(); i++ {
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
		sliceSchema.internals.ZodTypeInternals.AddCheck(alwaysPassCheck)
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
	check := checks.NewCustom[R](wrapper, utils.GetFirstParam(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}
