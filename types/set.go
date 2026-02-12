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

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodSetDef defines the configuration for a set schema.
type ZodSetDef struct {
	core.ZodTypeDef
	ValueType any
}

// ZodSetInternals contains set validator internal state.
type ZodSetInternals[T comparable] struct {
	core.ZodTypeInternals
	Def       *ZodSetDef
	ValueType any
}

// ZodSet represents a type-safe set validation schema.
// T is the element type (must be comparable), R is the constraint type (value or pointer).
type ZodSet[T comparable, R any] struct {
	internals *ZodSetInternals[T]
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema.
func (z *ZodSet[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodSet[T, R]) IsOptional() bool { return z.internals.IsOptional() }

// IsNilable reports whether this schema accepts nil values.
func (z *ZodSet[T, R]) IsNilable() bool { return z.internals.IsNilable() }

// Parse validates input using set-specific parsing logic.
func (z *ZodSet[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

	result, err := engine.ParseComplex[map[T]struct{}](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSet,
		z.extractSet,
		z.extractSetPtr,
		z.validateSet,
		ctx...,
	)
	if err != nil {
		return zero, err
	}

	switch v := result.(type) {
	case map[T]struct{}:
		return convertToSetConstraintType[T, R](v), nil
	case *map[T]struct{}:
		return convertToSetConstraintType[T, R](v), nil
	case **map[T]struct{}:
		if v != nil {
			return convertToSetConstraintType[T, R](*v), nil
		}
		return convertToSetConstraintType[T, R](nil), nil
	case nil:
		return convertToSetConstraintType[T, R](nil), nil
	default:
		if typedResult, ok := result.(R); ok {
			return typedResult, nil
		}
		return zero, fmt.Errorf("%w %T", ErrParseComplexUnexpectedType, result)
	}
}

// MustParse validates input and panics on error.
func (z *ZodSet[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodSet[T, R]) StrictParse(input map[T]struct{}, ctx ...*core.ParseContext) (R, error) {
	constraintInput := convertToSetConstraintType[T, R](input)

	return engine.ParseComplexStrict[map[T]struct{}, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSet,
		z.extractSet,
		z.extractSetPtr,
		z.validateSet,
		ctx...,
	)
}

// MustStrictParse validates input with type safety and panics on error.
func (z *ZodSet[T, R]) MustStrictParse(input map[T]struct{}, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodSet[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates an optional set schema that returns a pointer constraint.
func (z *ZodSet[T, R]) Optional() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns a pointer constraint.
func (z *ZodSet[T, R]) Nilable() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodSet[T, R]) Nullish() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the Optional flag and enforces a non-nil set constraint.
func (z *ZodSet[T, R]) NonOptional() *ZodSet[T, map[T]struct{}] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodSet[T, map[T]struct{}]{
		internals: &ZodSetInternals[T]{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			ValueType:        z.internals.ValueType,
		},
	}
}

// Default sets a default value when input is nil (short-circuit, bypasses validation).
func (z *ZodSet[T, R]) Default(v map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a dynamic default value using a function.
func (z *ZodSet[T, R]) DefaultFunc(fn func() map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides a fallback value that goes through full validation.
func (z *ZodSet[T, R]) Prefault(v map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides a dynamic fallback value using a function.
func (z *ZodSet[T, R]) PrefaultFunc(fn func() map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this set schema in the global registry.
func (z *ZodSet[T, R]) Meta(meta core.GlobalMeta) *ZodSet[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodSet[T, R]) Describe(description string) *ZodSet[T, R] {
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
func (z *ZodSet[T, R]) Min(minLen int, params ...any) *ZodSet[T, R] {
	return z.withCheck(checks.MinSize(minLen, params...))
}

// Max sets the maximum number of elements.
func (z *ZodSet[T, R]) Max(maxLen int, params ...any) *ZodSet[T, R] {
	return z.withCheck(checks.MaxSize(maxLen, params...))
}

// Size sets the exact number of elements.
func (z *ZodSet[T, R]) Size(exactLen int, params ...any) *ZodSet[T, R] {
	return z.withCheck(checks.Size(exactLen, params...))
}

// NonEmpty ensures the set has at least one element.
func (z *ZodSet[T, R]) NonEmpty(params ...any) *ZodSet[T, R] {
	return z.withCheck(checks.MinSize(1, params...))
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// ValueType returns the value schema for this set.
func (z *ZodSet[T, R]) ValueType() any { return z.internals.ValueType }

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed set value.
func (z *ZodSet[T, R]) Transform(fn func(R, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(input, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the set value while keeping the same type.
func (z *ZodSet[T, R]) Overwrite(transform func(R) R, params ...any) *ZodSet[T, R] {
	transformAny := func(input any) any {
		if converted, ok := convertToSetType[T, R](input); ok {
			return transform(converted)
		}
		return input
	}
	return z.withCheck(checks.NewZodCheckOverwrite(transformAny, params...))
}

// Pipe creates a validation pipeline with a target schema.
func (z *ZodSet[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(input, ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe custom validation with constraint type matching.
func (z *ZodSet[T, R]) Refine(fn func(R) bool, params ...any) *ZodSet[T, R] {
	wrapper := func(value any) bool {
		return fn(convertToSetConstraintType[T, R](value))
	}
	param := utils.FirstParam(params...)
	return z.withCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(param)))
}

// RefineAny provides flexible validation without type conversion.
func (z *ZodSet[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodSet[T, R] {
	param := utils.FirstParam(params...)
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(param)))
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodSet[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodSet[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodSet[T, R]) withCheck(check core.ZodCheck) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withInternals creates a new instance preserving the current constraint type.
func (z *ZodSet[T, R]) withInternals(in *core.ZodTypeInternals) *ZodSet[T, R] {
	return &ZodSet[T, R]{internals: &ZodSetInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

// withPtrInternals creates a new instance with pointer constraint type.
func (z *ZodSet[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodSet[T, *map[T]struct{}] {
	return &ZodSet[T, *map[T]struct{}]{internals: &ZodSetInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodSet[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodSet[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractSet extracts map[T]struct{} from value with proper conversion.
func (z *ZodSet[T, R]) extractSet(value any) (map[T]struct{}, bool) {
	if setVal, ok := value.(map[T]struct{}); ok {
		return setVal, true
	}

	if slice, ok := value.([]T); ok {
		set := make(map[T]struct{}, len(slice))
		for _, elem := range slice {
			set[elem] = struct{}{}
		}
		return set, true
	}

	if anySlice, ok := value.([]any); ok {
		set := make(map[T]struct{}, len(anySlice))
		for _, elem := range anySlice {
			typedElem, ok := elem.(T)
			if !ok {
				return nil, false
			}
			set[typedElem] = struct{}{}
		}
		return set, true
	}

	if value == nil {
		return nil, false
	}

	// Fall back to reflection for non-standard map/slice types.
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Map && rv.Type().Elem() == reflect.TypeOf(struct{}{}) {
		set := make(map[T]struct{}, rv.Len())
		for _, key := range rv.MapKeys() {
			typedKey, ok := key.Interface().(T)
			if !ok {
				return nil, false
			}
			set[typedKey] = struct{}{}
		}
		return set, true
	}
	if rv.Kind() == reflect.Slice {
		set := make(map[T]struct{}, rv.Len())
		for i := range rv.Len() {
			typedElem, ok := rv.Index(i).Interface().(T)
			if !ok {
				return nil, false
			}
			set[typedElem] = struct{}{}
		}
		return set, true
	}

	return nil, false
}

// extractSetPtr extracts *map[T]struct{} from value.
func (z *ZodSet[T, R]) extractSetPtr(value any) (*map[T]struct{}, bool) {
	if setVal, ok := value.(*map[T]struct{}); ok {
		return setVal, true
	}

	if setVal, ok := value.(map[T]struct{}); ok {
		return &setVal, true
	}

	if setVal, ok := z.extractSet(value); ok {
		return &setVal, true
	}

	return nil, false
}

// validateSet validates set elements using the value schema.
func (z *ZodSet[T, R]) validateSet(value map[T]struct{}, checksToRun []core.ZodCheck, ctx *core.ParseContext) (map[T]struct{}, error) {
	var collectedIssues []core.ZodRawIssue

	payload := core.NewParsePayload(value)
	result := engine.RunChecksOnValue(value, checksToRun, payload, ctx)

	if result.HasIssues() {
		collectedIssues = append(collectedIssues, result.GetIssues()...)
	}

	// Use the potentially transformed value for element validation.
	validatedValue := value
	if result.GetValue() != nil {
		if converted, ok := result.GetValue().(map[T]struct{}); ok {
			validatedValue = converted
		}
	}

	if z.internals.ValueType != nil {
		for elem := range validatedValue {
			if err := z.validateElement(elem, z.internals.ValueType, ctx); err != nil {
				var zodErr *issues.ZodError
				if errors.As(err, &zodErr) {
					for _, elementIssue := range zodErr.Issues {
						collectedIssues = append(collectedIssues, issues.ConvertZodIssueToRawWithProperties(elementIssue, []any{elem}))
					}
				} else {
					rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, elem)
					rawIssue.Path = []any{elem}
					collectedIssues = append(collectedIssues, rawIssue)
				}
			}
		}
	}

	if len(collectedIssues) > 0 {
		var zero map[T]struct{}
		return zero, issues.CreateArrayValidationIssues(collectedIssues)
	}

	return validatedValue, nil
}

// validateElement validates a single element using the provided schema.
func (z *ZodSet[T, R]) validateElement(element T, schema any, ctx *core.ParseContext) error {
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

// convertToSetType converts any value to the specified set constraint type.
func convertToSetType[T comparable, R any](v any) (R, bool) {
	var zero R

	if set, ok := v.(map[T]struct{}); ok {
		return convertToSetConstraintType[T, R](set), true
	}

	if ptrSet, ok := v.(*map[T]struct{}); ok {
		return convertToSetConstraintType[T, R](ptrSet), true
	}

	return zero, false
}

// convertToSetConstraintType converts values to the constraint type R.
func convertToSetConstraintType[T comparable, R any](value any) R {
	var zero R
	zeroType := reflect.TypeOf(zero)

	if value == nil {
		return zero
	}

	if zeroType != nil && zeroType.Kind() == reflect.Pointer {
		switch v := value.(type) {
		case map[T]struct{}:
			ptr := &v
			return any(ptr).(R)
		case *map[T]struct{}:
			return any(v).(R)
		}
	} else {
		switch v := value.(type) {
		case map[T]struct{}:
			return any(v).(R)
		case *map[T]struct{}:
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

// newZodSetFromDef constructs a new ZodSet from a definition.
func newZodSetFromDef[T comparable, R any](def *ZodSetDef) *ZodSet[T, R] {
	internals := &ZodSetInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		ValueType:        def.ValueType,
	}

	// Provide constructor for AddCheck functionality.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		setDef := &ZodSetDef{
			ZodTypeDef: *newDef,
			ValueType:  def.ValueType,
		}
		return any(newZodSetFromDef[T, R](setDef)).(core.ZodType[any])
	}

	schema := &ZodSet[T, R]{internals: internals}

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

// Set creates a set schema with element validation.
func Set[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, map[T]struct{}] {
	return SetTyped[T, map[T]struct{}](valueSchema, paramArgs...)
}

// SetPtr creates a set schema returning a pointer constraint.
func SetPtr[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, *map[T]struct{}] {
	return SetTyped[T, *map[T]struct{}](valueSchema, paramArgs...)
}

// SetTyped creates a typed set schema with generic constraints.
func SetTyped[T comparable, R any](valueSchema any, paramArgs ...any) *ZodSet[T, R] {
	param := utils.FirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodSetDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeSet,
			Checks: []core.ZodCheck{},
		},
		ValueType: valueSchema,
	}

	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	setSchema := newZodSetFromDef[T, R](def)

	// Ensure validator is called when value schema exists.
	if valueSchema != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		setSchema.internals.AddCheck(alwaysPassCheck)
	}

	return setSchema
}

// Check adds a custom validation function that can report multiple issues.
func (z *ZodSet[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSet[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

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
func (z *ZodSet[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSet[T, R] {
	return z.Check(fn, params...)
}
