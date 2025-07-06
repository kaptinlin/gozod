package types

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodIntersectionDef defines the schema definition for intersection validation
type ZodIntersectionDef struct {
	core.ZodTypeDef
	Left  core.ZodSchema // Left schema using unified interface
	Right core.ZodSchema // Right schema using unified interface
}

// ZodIntersectionInternals contains intersection validator internal state
type ZodIntersectionInternals struct {
	core.ZodTypeInternals
	Def   *ZodIntersectionDef // Schema definition reference
	Left  core.ZodSchema      // Left schema for runtime validation
	Right core.ZodSchema      // Right schema for runtime validation
}

// ZodIntersection represents an intersection validation schema with dual generic parameters
// T = base type (any), R = constraint type (any or *any)
type ZodIntersection[T any, R any] struct {
	internals *ZodIntersectionInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodIntersection[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodIntersection[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodIntersection[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse validates the input using direct validation approach with intersection logic
func (z *ZodIntersection[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Handle nil values for optional/nilable cases
	if input == nil {
		if z.internals.Nilable || z.internals.Optional {
			var zero R
			return zero, nil
		}
		if z.internals.DefaultValue != nil {
			return z.Parse(z.internals.DefaultValue, parseCtx)
		}
		if z.internals.DefaultFunc != nil {
			return z.Parse(z.internals.DefaultFunc(), parseCtx)
		}
		// No default, create error
		issue := issues.CreateInvalidTypeIssue(core.ZodTypeIntersection, input)
		finalIssue := issues.FinalizeIssue(issue, parseCtx, core.GetConfig())
		return *new(R), issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Parse with left schema using ParseAny
	leftResult, leftErr := z.internals.Left.ParseAny(input, parseCtx)

	// Parse with right schema using ParseAny
	rightResult, rightErr := z.internals.Right.ParseAny(input, parseCtx)

	// Collect all errors
	var allErrors []core.ZodIssue
	if leftErr != nil {
		var zErr *issues.ZodError
		if issues.IsZodError(leftErr, &zErr) {
			allErrors = append(allErrors, zErr.Issues...)
		} else {
			issue := issues.CreateCustomIssue(leftErr.Error(), nil, input)
			allErrors = append(allErrors, issues.FinalizeIssue(issue, parseCtx, core.GetConfig()))
		}
	}
	if rightErr != nil {
		var zErr *issues.ZodError
		if issues.IsZodError(rightErr, &zErr) {
			allErrors = append(allErrors, zErr.Issues...)
		} else {
			issue := issues.CreateCustomIssue(rightErr.Error(), nil, input)
			allErrors = append(allErrors, issues.FinalizeIssue(issue, parseCtx, core.GetConfig()))
		}
	}

	// If either side failed, return all errors
	if len(allErrors) > 0 {
		return *new(R), issues.NewZodError(allErrors)
	}

	// Both sides succeeded, attempt to merge results
	merged, mergeErr := mergeValues(leftResult, rightResult)
	if mergeErr != nil {
		issue := issues.CreateCustomIssue(mergeErr.Error(), map[string]any{"type": "intersection"}, input)
		finalIssue := issues.FinalizeIssue(issue, parseCtx, core.GetConfig())
		return *new(R), issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Run additional checks if any
	if len(z.internals.Checks) > 0 {
		transformedMerged, err := engine.ApplyChecks[any](merged, z.internals.Checks, parseCtx)
		if err != nil {
			return *new(R), err
		}
		merged = transformedMerged
	}

	// Convert result to constraint type R
	return convertToIntersectionConstraintType[T, R](merged), nil
}

// MustParse validates the input value and panics on failure
func (z *ZodIntersection[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodIntersection[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional makes the intersection optional and returns pointer constraint
func (z *ZodIntersection[T, R]) Optional() *ZodIntersection[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable makes the intersection nilable and returns pointer constraint
func (z *ZodIntersection[T, R]) Nilable() *ZodIntersection[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodIntersection[T, R]) Nullish() *ZodIntersection[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil value (T).
func (z *ZodIntersection[T, R]) NonOptional() *ZodIntersection[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodIntersection[T, T]{
		internals: &ZodIntersectionInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Left:             z.internals.Left,
			Right:            z.internals.Right,
		},
	}
}

// Default preserves current constraint type R
func (z *ZodIntersection[T, R]) Default(v T) *ZodIntersection[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodIntersection[T, R]) DefaultFunc(fn func() T) *ZodIntersection[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodIntersection[T, R]) Prefault(v T) *ZodIntersection[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodIntersection[T, R]) PrefaultFunc(fn func() T) *ZodIntersection[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Left returns the left schema of the intersection
func (z *ZodIntersection[T, R]) Left() core.ZodSchema {
	return z.internals.Left
}

// Right returns the right schema of the intersection
func (z *ZodIntersection[T, R]) Right() core.ZodSchema {
	return z.internals.Right
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates type-safe transformation pipeline using WrapFn pattern
func (z *ZodIntersection[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractIntersectionValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates validation pipeline to another schema using WrapFn pattern
func (z *ZodIntersection[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractIntersectionValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, wrapperFn)
}

// =============================================================================
// TYPE CONVERSION - NO LONGER NEEDED (USING WRAPFN PATTERN)
// =============================================================================

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation with constraint type R
func (z *ZodIntersection[T, R]) Refine(fn func(R) bool, params ...any) *ZodIntersection[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToIntersectionConstraintValue[T, R](v); ok {
			return fn(constraintValue)
		}
		return false
	}

	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodIntersection[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodIntersection[T, R] {
	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodIntersection[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodIntersection[T, *T] {
	return &ZodIntersection[T, *T]{internals: &ZodIntersectionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Left:             z.internals.Left,
		Right:            z.internals.Right,
	}}
}

func (z *ZodIntersection[T, R]) withInternals(in *core.ZodTypeInternals) *ZodIntersection[T, R] {
	return &ZodIntersection[T, R]{internals: &ZodIntersectionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Left:             z.internals.Left,
		Right:            z.internals.Right,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodIntersection[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodIntersection[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToIntersectionConstraintType converts a base type T to constraint type R
func convertToIntersectionConstraintType[T any, R any](value any) R {
	var zero R
	switch any(zero).(type) {
	case *any:
		// Need to return *any from any
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R)
		}
		return any((*any)(nil)).(R)
	default:
		// Return value directly as R
		return any(value).(R)
	}
}

// extractIntersectionValue extracts base type T from constraint type R
func extractIntersectionValue[T any, R any](value R) T {
	switch v := any(value).(type) {
	case *any:
		if v != nil {
			return any(*v).(T)
		}
		var zero T
		return zero
	default:
		return any(value).(T)
	}
}

// convertToIntersectionConstraintValue converts any value to constraint type R if possible
func convertToIntersectionConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok {
		return r, true
	}

	// Handle pointer conversion for intersection types
	if _, ok := any(zero).(*any); ok {
		// Need to convert any to *any
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R), true
		}
		return any((*any)(nil)).(R), true
	}

	return zero, false
}

// =============================================================================
// VALUE MERGING HELPERS
// =============================================================================

// mergeValues attempts to merge two validated values
func mergeValues(a, b any) (any, error) {
	// If values are identical, return one of them
	if reflect.DeepEqual(a, b) {
		return a, nil
	}

	// Handle nil cases
	if a == nil && b == nil {
		return nil, nil
	}
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}

	// Get reflection values
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Allow merging of different struct types by converting to map; otherwise types must match
	if aVal.Type() != bVal.Type() {
		if aVal.Kind() != reflect.Struct || bVal.Kind() != reflect.Struct {
			return nil, fmt.Errorf("cannot merge incompatible types: %v and %v", aVal.Type(), bVal.Type())
		}
		// Continue to struct merging logic below
	}

	// Handle different types
	//nolint:exhaustive
	switch aVal.Kind() {
	case reflect.Map:
		return mergeMaps(a, b)
	case reflect.Slice, reflect.Array:
		return mergeSlices(a, b)
	case reflect.Struct:
		// Convert both structs to map[string]any and merge maps
		aMap := structToMap(aVal)
		bMap := structToMap(bVal)
		return mergeMaps(aMap, bMap)
	default:
		// For basic types, values must be identical
		return nil, fmt.Errorf("cannot merge different values of type %v: %v != %v", aVal.Type(), a, b)
	}
}

// mergeMaps merges two map values
func mergeMaps(a, b any) (any, error) {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	if aVal.Kind() != reflect.Map || bVal.Kind() != reflect.Map {
		return nil, fmt.Errorf("both values must be maps")
	}

	// Create new map of the same type
	mapType := aVal.Type()
	result := reflect.MakeMap(mapType)

	// Copy all keys from a
	for _, key := range aVal.MapKeys() {
		result.SetMapIndex(key, aVal.MapIndex(key))
	}

	// Merge keys from b
	for _, key := range bVal.MapKeys() {
		bValue := bVal.MapIndex(key)
		if aValue := aVal.MapIndex(key); aValue.IsValid() {
			// Key exists in both maps - they must have the same value
			if !reflect.DeepEqual(aValue.Interface(), bValue.Interface()) {
				return nil, fmt.Errorf("conflicting values for key %v: %v != %v",
					key.Interface(), aValue.Interface(), bValue.Interface())
			}
		} else {
			// Key only exists in b, add it
			result.SetMapIndex(key, bValue)
		}
	}

	return result.Interface(), nil
}

// mergeSlices merges two slice values
func mergeSlices(a, b any) (any, error) {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	if aVal.Kind() != reflect.Slice && aVal.Kind() != reflect.Array {
		return nil, fmt.Errorf("first value must be slice or array")
	}
	if bVal.Kind() != reflect.Slice && bVal.Kind() != reflect.Array {
		return nil, fmt.Errorf("second value must be slice or array")
	}

	// For arrays/slices, they must be identical for intersection
	if !reflect.DeepEqual(a, b) {
		return nil, fmt.Errorf("slice/array values must be identical for intersection")
	}

	return a, nil
}

// structToMap converts struct value to map[string]any using exported fields and json tags.
func structToMap(v reflect.Value) map[string]any {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	result := make(map[string]any)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		key := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			tagName := strings.Split(tag, ",")[0]
			if tagName != "" && tagName != "-" {
				key = tagName
			}
		}
		result[key] = v.Field(i).Interface()
	}
	return result
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodIntersectionFromDef constructs new ZodIntersection from definition
func newZodIntersectionFromDef[T any, R any](def *ZodIntersectionDef) *ZodIntersection[T, R] {
	internals := &ZodIntersectionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Left:             def.Left,
		Right:            def.Right,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		intersectionDef := &ZodIntersectionDef{
			ZodTypeDef: *newDef,
			Left:       def.Left,
			Right:      def.Right,
		}
		return any(newZodIntersectionFromDef[T, R](intersectionDef)).(core.ZodType[any])
	}

	schema := &ZodIntersection[T, R]{internals: internals}

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

// Intersection creates intersection schema that validates with both schemas - returns value constraint
func Intersection(left, right any, args ...any) *ZodIntersection[any, any] {
	return IntersectionTyped[any, any](left, right, args...)
}

// IntersectionPtr creates intersection schema that validates with both schemas - returns pointer constraint
func IntersectionPtr(left, right any, args ...any) *ZodIntersection[any, *any] {
	return IntersectionTyped[any, *any](left, right, args...)
}

// IntersectionTyped creates typed intersection schema with generic constraints
func IntersectionTyped[T any, R any](left, right any, args ...any) *ZodIntersection[T, R] {
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	// Convert inputs to core.ZodSchema using direct type assertion
	leftWrapped, err := core.ConvertToZodSchema(left)
	if err != nil {
		panic(fmt.Sprintf("Intersection left schema: %v", err))
	}
	rightWrapped, err := core.ConvertToZodSchema(right)
	if err != nil {
		panic(fmt.Sprintf("Intersection right schema: %v", err))
	}

	def := &ZodIntersectionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeIntersection,
			Checks: []core.ZodCheck{},
		},
		Left:  leftWrapped,
		Right: rightWrapped,
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodIntersectionFromDef[T, R](def)
}
