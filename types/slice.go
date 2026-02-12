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

var (
	ErrParseComplexUnexpectedType = errors.New("internal error: ParseComplex returned unexpected type")
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodSliceDef defines the configuration for a slice schema.
type ZodSliceDef struct {
	core.ZodTypeDef
	Element any
}

// ZodSliceInternals contains slice validator internal state.
type ZodSliceInternals[T any] struct {
	core.ZodTypeInternals
	Def     *ZodSliceDef
	Element any
}

// ZodSlice represents a type-safe slice validation schema.
// T is the element type, R is the constraint type (value or pointer).
type ZodSlice[T any, R any] struct {
	internals *ZodSliceInternals[T]
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema.
func (z *ZodSlice[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodSlice[T, R]) IsOptional() bool { return z.internals.IsOptional() }

// IsNilable reports whether this schema accepts nil values.
func (z *ZodSlice[T, R]) IsNilable() bool { return z.internals.IsNilable() }

// Parse validates input using slice-specific parsing logic.
func (z *ZodSlice[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

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
		return zero, err
	}

	switch v := result.(type) {
	case []T:
		return convertToSliceConstraintType[T, R](v), nil
	case *[]T:
		return convertToSliceConstraintType[T, R](v), nil
	case **[]T:
		if v != nil {
			return convertToSliceConstraintType[T, R](*v), nil
		}
		return convertToSliceConstraintType[T, R](nil), nil
	case nil:
		return convertToSliceConstraintType[T, R](nil), nil
	default:
		if typedResult, ok := result.(R); ok {
			return typedResult, nil
		}
		return zero, fmt.Errorf("%w %T", ErrParseComplexUnexpectedType, result)
	}
}

// MustParse validates input and panics on error.
func (z *ZodSlice[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodSlice[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParseComplexStrict[[]T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSlice,
		z.extractSlice,
		z.extractSlicePtr,
		z.validateSlice,
		ctx...,
	)
}

// MustStrictParse validates input with type safety and panics on error.
func (z *ZodSlice[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodSlice[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates an optional slice schema that returns a pointer constraint.
func (z *ZodSlice[T, R]) Optional() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns a pointer constraint.
func (z *ZodSlice[T, R]) Nilable() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodSlice[T, R]) Nullish() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the Optional flag and enforces a non-nil slice constraint.
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

// Default sets a default value when input is nil (short-circuit, bypasses validation).
func (z *ZodSlice[T, R]) Default(v []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a dynamic default value using a function.
func (z *ZodSlice[T, R]) DefaultFunc(fn func() []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides a fallback value that goes through full validation.
func (z *ZodSlice[T, R]) Prefault(v []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides a dynamic fallback value using a function.
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

// Min sets the minimum number of elements.
func (z *ZodSlice[T, R]) Min(minLen int, params ...any) *ZodSlice[T, R] {
	return z.withCheck(checks.MinSize(minLen, params...))
}

// Max sets the maximum number of elements.
func (z *ZodSlice[T, R]) Max(maxLen int, params ...any) *ZodSlice[T, R] {
	return z.withCheck(checks.MaxSize(maxLen, params...))
}

// Length sets the exact number of elements.
func (z *ZodSlice[T, R]) Length(exactLen int, params ...any) *ZodSlice[T, R] {
	return z.withCheck(checks.Size(exactLen, params...))
}

// NonEmpty ensures the slice has at least one element.
func (z *ZodSlice[T, R]) NonEmpty(params ...any) *ZodSlice[T, R] {
	return z.withCheck(checks.MinSize(1, params...))
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Element returns the element schema.
func (z *ZodSlice[T, R]) Element() any { return z.internals.Element }

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed slice value.
func (z *ZodSlice[T, R]) Transform(fn func(R, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(input, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the slice value while keeping the same type.
func (z *ZodSlice[T, R]) Overwrite(transform func(R) R, params ...any) *ZodSlice[T, R] {
	transformAny := func(input any) any {
		if converted, ok := convertToSliceType[T, R](input); ok {
			return transform(converted)
		}
		return input
	}
	return z.withCheck(checks.NewZodCheckOverwrite(transformAny, params...))
}

// Pipe creates a validation pipeline with a target schema.
func (z *ZodSlice[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(input, ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a type-safe custom validation function.
func (z *ZodSlice[T, R]) Refine(fn func(R) bool, params ...any) *ZodSlice[T, R] {
	wrapper := func(value any) bool {
		return fn(convertToSliceConstraintType[T, R](value))
	}
	param := utils.FirstParam(params...)
	return z.withCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(param)))
}

// RefineAny adds a custom validation function without type conversion.
func (z *ZodSlice[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodSlice[T, R] {
	param := utils.FirstParam(params...)
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(param)))
}

// =============================================================================
// COMPOSITION METHODS
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodSlice[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodSlice[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodSlice[T, R]) withCheck(check core.ZodCheck) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withInternals creates a new instance preserving the current constraint type.
func (z *ZodSlice[T, R]) withInternals(in *core.ZodTypeInternals) *ZodSlice[T, R] {
	return &ZodSlice[T, R]{internals: &ZodSliceInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Element:          z.internals.Element,
	}}
}

// withPtrInternals creates a new instance with pointer constraint type.
func (z *ZodSlice[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodSlice[T, *[]T] {
	return &ZodSlice[T, *[]T]{internals: &ZodSliceInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Element:          z.internals.Element,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodSlice[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodSlice[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractSlice extracts []T from the input value.
func (z *ZodSlice[T, R]) extractSlice(value any) ([]T, bool) {
	if sliceVal, ok := value.([]T); ok {
		return sliceVal, true
	}

	if anySlice, ok := value.([]any); ok {
		converted := make([]T, len(anySlice))
		for i, elem := range anySlice {
			typedElem, ok := elem.(T)
			if !ok {
				return nil, false
			}
			converted[i] = typedElem
		}
		return converted, true
	}

	if value == nil {
		return nil, false
	}

	// Fall back to slicex for non-standard slice types.
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Slice {
		if converted, err := slicex.ToAny(value); err == nil && converted != nil {
			typedSlice := make([]T, len(converted))
			for i, elem := range converted {
				typedElem, ok := elem.(T)
				if !ok {
					return nil, false
				}
				typedSlice[i] = typedElem
			}
			return typedSlice, true
		}
	}

	return nil, false
}

// extractSlicePtr extracts *[]T from the input value.
func (z *ZodSlice[T, R]) extractSlicePtr(value any) (*[]T, bool) {
	// Preserve original pointer identity when possible.
	if sliceVal, ok := value.(*[]T); ok {
		return sliceVal, true
	}

	if sliceVal, ok := value.([]T); ok {
		return &sliceVal, true
	}

	return nil, false
}

// validateSlice validates slice elements using the element schema.
func (z *ZodSlice[T, R]) validateSlice(value []T, checks []core.ZodCheck, ctx *core.ParseContext) ([]T, error) {
	var collectedIssues []core.ZodRawIssue

	payload := core.NewParsePayload(value)
	result := engine.RunChecksOnValue(value, checks, payload, ctx)

	if result.HasIssues() {
		collectedIssues = append(collectedIssues, result.GetIssues()...)
	}

	// Use the potentially transformed value for element validation.
	validatedValue := value
	if result.GetValue() != nil {
		if converted, ok := result.GetValue().([]T); ok {
			validatedValue = converted
		}
	}

	if z.internals.Element != nil {
		for i, element := range validatedValue {
			if err := z.validateElement(element, z.internals.Element, ctx); err != nil {
				var zodErr *issues.ZodError
				if errors.As(err, &zodErr) {
					for _, elementIssue := range zodErr.Issues {
						collectedIssues = append(collectedIssues, issues.ConvertZodIssueToRawWithProperties(elementIssue, []any{i}))
					}
				} else {
					rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, element)
					rawIssue.Path = []any{i}
					collectedIssues = append(collectedIssues, rawIssue)
				}
			}
		}
	}

	if len(collectedIssues) > 0 {
		var zero []T
		return zero, issues.CreateArrayValidationIssues(collectedIssues)
	}

	return validatedValue, nil
}

// validateElement validates a single element using the provided schema.
func (z *ZodSlice[T, R]) validateElement(element T, schema any, ctx *core.ParseContext) error {
	if schema == nil {
		return nil
	}

	// Use reflection to call Parse method; handles all schema types including Lazy.
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

	args := []reflect.Value{reflect.ValueOf(element)}
	if methodType.NumIn() > 1 && methodType.In(1).String() == "*core.ParseContext" {
		args = append(args, reflect.ValueOf(ctx))
	}

	results := parseMethod.Call(args)
	if len(results) >= 2 {
		if errInterface := results[1].Interface(); errInterface != nil {
			if err, ok := errInterface.(error); ok {
				return err
			}
		}
	}

	return nil
}

// convertToSliceType converts any value to the specified slice constraint type.
func convertToSliceType[T any, R any](v any) (R, bool) {
	var zero R

	if slice, ok := v.([]T); ok {
		return convertToSliceConstraintType[T, R](slice), true
	}

	if ptrSlice, ok := v.(*[]T); ok {
		return convertToSliceConstraintType[T, R](ptrSlice), true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
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

// convertToSliceConstraintType converts a value to the constraint type R.
func convertToSliceConstraintType[T any, R any](value any) R {
	var zero R
	zeroType := reflect.TypeOf(zero)

	if value == nil {
		return zero
	}

	if zeroType != nil && zeroType.Kind() == reflect.Pointer {
		switch v := value.(type) {
		case []T:
			ptr := &v
			return any(ptr).(R)
		case *[]T:
			return any(v).(R)
		}
	} else {
		switch v := value.(type) {
		case []T:
			return any(v).(R)
		case *[]T:
			if v != nil {
				return any(*v).(R)
			}
		}
	}

	if converted, ok := value.(R); ok {
		return converted
	}

	return zero
}

// newZodSliceFromDef constructs a new ZodSlice from a definition.
func newZodSliceFromDef[T any, R any](def *ZodSliceDef) *ZodSlice[T, R] {
	internals := &ZodSliceInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Element:          def.Element,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		sliceDef := &ZodSliceDef{
			ZodTypeDef: *newDef,
			Element:    def.Element,
		}
		return any(newZodSliceFromDef[T, R](sliceDef)).(core.ZodType[any])
	}

	schema := &ZodSlice[T, R]{internals: internals}

	if def.Error != nil {
		internals.Error = def.Error
	}

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

// Slice creates a slice schema with element validation.
func Slice[T any](elementSchema any, paramArgs ...any) *ZodSlice[T, []T] {
	return SliceTyped[T, []T](elementSchema, paramArgs...)
}

// SlicePtr creates a slice schema with pointer constraint.
func SlicePtr[T any](elementSchema any, paramArgs ...any) *ZodSlice[T, *[]T] {
	return SliceTyped[T, *[]T](elementSchema, paramArgs...)
}

// SliceTyped creates a slice schema with an explicit constraint type.
func SliceTyped[T any, R any](elementSchema any, paramArgs ...any) *ZodSlice[T, R] {
	param := utils.FirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodSliceDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeSlice,
			Checks: []core.ZodCheck{},
		},
		Element: elementSchema,
	}

	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	sliceSchema := newZodSliceFromDef[T, R](def)

	// Add a minimal check to trigger element validation when element schema exists.
	if elementSchema != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		sliceSchema.internals.AddCheck(alwaysPassCheck)
	}

	return sliceSchema
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodSlice[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSlice[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Handle pointer/value mismatch via reflection.
		var zero R
		zeroTyp := reflect.TypeOf(zero)
		if zeroTyp != nil && zeroTyp.Kind() == reflect.Pointer {
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
	return z.withCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
}

// With is an alias for Check.
func (z *ZodSlice[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSlice[T, R] {
	return z.Check(fn, params...)
}
