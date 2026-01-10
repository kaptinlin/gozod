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

// ZodSetDef defines the schema definition for set validation.
// In Go, sets are idiomatically represented as map[T]struct{}.
type ZodSetDef struct {
	core.ZodTypeDef
	ValueType any // The value schema (type-erased for flexibility)
}

// ZodSetInternals contains the internal state for set schema
type ZodSetInternals[T comparable] struct {
	core.ZodTypeInternals
	Def       *ZodSetDef // Schema definition reference
	ValueType any        // Value schema for runtime validation
}

// ZodSet represents a type-safe set validation schema with dual generic parameters
// T: element type (must be comparable), R: constraint type (map[T]struct{} or *map[T]struct{})
type ZodSet[T comparable, R any] struct {
	internals *ZodSetInternals[T]
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodSet[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodSet[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodSet[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input using set-specific parsing logic with type safety
func (z *ZodSet[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
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
		var zero R
		return zero, err
	}

	// Handle different return types from ParseComplex
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
		var zero R
		return zero, fmt.Errorf("%w %T", ErrParseComplexUnexpectedType, result)
	}
}

// MustParse validates the input value and panics on failure
func (z *ZodSet[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
func (z *ZodSet[T, R]) StrictParse(input map[T]struct{}, ctx ...*core.ParseContext) (R, error) {
	constraintInput := convertToSetConstraintType[T, R](input)

	result, err := engine.ParseComplexStrict[map[T]struct{}, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSet,
		z.extractSet,
		z.extractSetPtr,
		z.validateSet,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	return result, nil
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
func (z *ZodSet[T, R]) MustStrictParse(input map[T]struct{}, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
func (z *ZodSet[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates optional set schema (always returns pointer constraint)
func (z *ZodSet[T, R]) Optional() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values (always returns pointer constraint)
func (z *ZodSet[T, R]) Nilable() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers (always returns pointer constraint)
func (z *ZodSet[T, R]) Nullish() *ZodSet[T, *map[T]struct{}] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil set constraint.
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

// Default preserves current constraint type
func (z *ZodSet[T, R]) Default(v map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type
func (z *ZodSet[T, R]) DefaultFunc(fn func() map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure (preserves constraint type)
func (z *ZodSet[T, R]) Prefault(v map[T]struct{}) *ZodSet[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values (preserves constraint type)
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
// TypeScript Zod v4 equivalent: schema.describe(description)
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

// Min sets minimum number of elements
func (z *ZodSet[T, R]) Min(minLen int, params ...any) *ZodSet[T, R] {
	check := checks.MinSize(minLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of elements
func (z *ZodSet[T, R]) Max(maxLen int, params ...any) *ZodSet[T, R] {
	check := checks.MaxSize(maxLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Size sets exact number of elements
func (z *ZodSet[T, R]) Size(exactLen int, params ...any) *ZodSet[T, R] {
	check := checks.Size(exactLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// NonEmpty ensures set has at least one element
// TypeScript Zod v4 equivalent: z.set(...).nonempty()
func (z *ZodSet[T, R]) NonEmpty(params ...any) *ZodSet[T, R] {
	check := checks.MinSize(1, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// ValueType returns the value schema for this set
func (z *ZodSet[T, R]) ValueType() any {
	return z.internals.ValueType
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function using the WrapFn pattern.
func (z *ZodSet[T, R]) Transform(fn func(R, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(input, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms set value while keeping the same type
func (z *ZodSet[T, R]) Overwrite(transform func(R) R, params ...any) *ZodSet[T, R] {
	check := checks.NewZodCheckOverwrite(func(input any) any {
		if converted, ok := convertToSetType[T, R](input); ok {
			return transform(converted)
		}
		return input
	}, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using the WrapFn pattern.
func (z *ZodSet[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(input, ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds type-safe custom validation with automatic constraint type matching
func (z *ZodSet[T, R]) Refine(fn func(R) bool, params ...any) *ZodSet[T, R] {
	wrapper := func(value any) bool {
		constraintValue := convertToSetConstraintType[T, R](value)
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
func (z *ZodSet[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodSet[T, R] {
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
// TypeScript Zod v4 equivalent: schema.and(other)
func (z *ZodSet[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
// TypeScript Zod v4 equivalent: schema.or(other)
func (z *ZodSet[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withInternals creates new instance preserving current constraint type
func (z *ZodSet[T, R]) withInternals(in *core.ZodTypeInternals) *ZodSet[T, R] {
	return &ZodSet[T, R]{internals: &ZodSetInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

// withPtrInternals creates new instance with pointer constraint type
func (z *ZodSet[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodSet[T, *map[T]struct{}] {
	return &ZodSet[T, *map[T]struct{}]{internals: &ZodSetInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodSet[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodSet[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractSet extracts map[T]struct{} from value with proper conversion
func (z *ZodSet[T, R]) extractSet(value any) (map[T]struct{}, bool) {
	// Handle direct map[T]struct{}
	if setVal, ok := value.(map[T]struct{}); ok {
		return setVal, true
	}

	// Handle []T conversion (convert slice to set)
	if slice, ok := value.([]T); ok {
		set := make(map[T]struct{}, len(slice))
		for _, elem := range slice {
			set[elem] = struct{}{}
		}
		return set, true
	}

	// Handle []any conversion
	if anySlice, ok := value.([]any); ok {
		set := make(map[T]struct{}, len(anySlice))
		for _, elem := range anySlice {
			if typedElem, ok := elem.(T); ok {
				set[typedElem] = struct{}{}
			} else {
				return nil, false
			}
		}
		return set, true
	}

	// Try reflection for map types
	if value != nil {
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Map {
			// Check if it's a set-like map (values are struct{})
			if rv.Type().Elem() == reflect.TypeOf(struct{}{}) {
				set := make(map[T]struct{}, rv.Len())
				for _, key := range rv.MapKeys() {
					if typedKey, ok := key.Interface().(T); ok {
						set[typedKey] = struct{}{}
					} else {
						return nil, false
					}
				}
				return set, true
			}
		}
		// Try slice via reflection
		if rv.Kind() == reflect.Slice {
			set := make(map[T]struct{}, rv.Len())
			for i := range rv.Len() {
				elem := rv.Index(i).Interface()
				if typedElem, ok := elem.(T); ok {
					set[typedElem] = struct{}{}
				} else {
					return nil, false
				}
			}
			return set, true
		}
	}

	return nil, false
}

// extractSetPtr extracts *map[T]struct{} from value
func (z *ZodSet[T, R]) extractSetPtr(value any) (*map[T]struct{}, bool) {
	// Handle direct *map[T]struct{}
	if setVal, ok := value.(*map[T]struct{}); ok {
		return setVal, true
	}

	// Handle map[T]struct{}
	if setVal, ok := value.(map[T]struct{}); ok {
		ptrSet := &setVal
		return ptrSet, true
	}

	// Try extractSet and return pointer
	if setVal, ok := z.extractSet(value); ok {
		return &setVal, true
	}

	return nil, false
}

// validateSet validates set elements using value schema with multiple error collection
func (z *ZodSet[T, R]) validateSet(value map[T]struct{}, checksToRun []core.ZodCheck, ctx *core.ParseContext) (map[T]struct{}, error) {
	var collectedIssues []core.ZodRawIssue

	// Apply set-level checks (size, etc.) and collect any issues
	payload := core.NewParsePayload(value)
	result := engine.RunChecksOnValue(value, checksToRun, payload, ctx)

	// Collect any set-level issues
	if result.HasIssues() {
		setIssues := result.GetIssues()
		collectedIssues = append(collectedIssues, setIssues...)
	}

	// Get the potentially transformed value for element validation
	var validatedValue map[T]struct{}
	if result.GetValue() != nil {
		if converted, ok := result.GetValue().(map[T]struct{}); ok {
			validatedValue = converted
		} else {
			validatedValue = value
		}
	} else {
		validatedValue = value
	}

	// Validate each element if value schema is provided
	if z.internals.ValueType != nil {
		for elem := range validatedValue {
			if err := z.validateElement(elem, z.internals.ValueType, ctx); err != nil {
				var zodErr *issues.ZodError
				if errors.As(err, &zodErr) {
					for _, elementIssue := range zodErr.Issues {
						rawIssue := core.ZodRawIssue{
							Code:       elementIssue.Code,
							Message:    elementIssue.Message,
							Input:      elementIssue.Input,
							Path:       []any{elem}, // Use element as path key
							Properties: make(map[string]any),
						}
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
					rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, elem)
					rawIssue.Path = []any{elem}
					collectedIssues = append(collectedIssues, rawIssue)
				}
			}
		}
	}

	// If we collected any issues, return them as a combined error
	if len(collectedIssues) > 0 {
		var zero map[T]struct{}
		return zero, issues.CreateArrayValidationIssues(collectedIssues)
	}

	return validatedValue, nil
}

// validateElement validates a single element using the provided schema
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

// convertToSetType converts any value to the specified set constraint type
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

// convertToSetConstraintType converts values to the constraint type R
func convertToSetConstraintType[T comparable, R any](value any) R {
	var zero R
	zeroType := reflect.TypeOf(zero)

	if value == nil {
		return zero
	}

	// Check if R is a pointer type (*map[T]struct{})
	if zeroType != nil && zeroType.Kind() == reflect.Ptr {
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

// newZodSetFromDef constructs new ZodSet from definition
func newZodSetFromDef[T comparable, R any](def *ZodSetDef) *ZodSet[T, R] {
	internals := &ZodSetInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		ValueType:        def.ValueType,
	}

	// Provide constructor for AddCheck functionality
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

// Set creates set schema with element validation (returns value constraint)
// In Go, sets are represented as map[T]struct{} where T must be comparable.
// TypeScript Zod v4 equivalent: z.set(schema)
func Set[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, map[T]struct{}] {
	return SetTyped[T, map[T]struct{}](valueSchema, paramArgs...)
}

// SetPtr creates set schema with pointer constraint (returns pointer constraint)
func SetPtr[T comparable](valueSchema any, paramArgs ...any) *ZodSet[T, *map[T]struct{}] {
	return SetTyped[T, *map[T]struct{}](valueSchema, paramArgs...)
}

// SetTyped creates set schema with explicit constraint type (universal constructor)
func SetTyped[T comparable, R any](valueSchema any, paramArgs ...any) *ZodSet[T, R] {
	param := utils.GetFirstParam(paramArgs...)
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

	// Ensure validator is called when value schema exists
	if valueSchema != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		setSchema.internals.AddCheck(alwaysPassCheck)
	}

	return setSchema
}

// Check adds a custom validation function for set schema that can push multiple issues.
func (z *ZodSet[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSet[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

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
func (z *ZodSet[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodSet[T, R] {
	return z.Check(fn, params...)
}
