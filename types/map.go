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
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodMapDef defines the configuration for a map schema.
type ZodMapDef struct {
	core.ZodTypeDef
	KeyType   any
	ValueType any
}

// ZodMapInternals contains map validator internal state.
type ZodMapInternals struct {
	core.ZodTypeInternals
	Def       *ZodMapDef
	KeyType   any
	ValueType any
}

// ZodMap represents a type-safe map validation schema.
// T is the base type, R is the constraint type (value or pointer).
type ZodMap[T any, R any] struct {
	internals *ZodMapInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema.
func (z *ZodMap[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodMap[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodMap[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodMap[T, R]) withCheck(check core.ZodCheck) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// Parse validates input using map-specific parsing logic with engine.ParseComplex.
func (z *ZodMap[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

	result, err := engine.ParseComplex[map[any]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeMap,
		z.extractMapForEngine,
		z.extractMapPtrForEngine,
		z.validateMapForEngine,
		ctx...,
	)
	if err != nil {
		return zero, err
	}

	switch v := result.(type) {
	case map[any]any:
		return convertMapFromGeneric[T, R](v), nil
	case *map[any]any:
		if v == nil {
			return zero, nil
		}
		return convertMapFromGeneric[T, R](*v), nil
	case nil:
		return zero, nil
	default:
		parseCtx := core.NewParseContext()
		if len(ctx) > 0 && ctx[0] != nil {
			parseCtx = ctx[0]
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", result), "map", input, parseCtx,
		)
	}
}

// MustParse validates input and panics on error.
func (z *ZodMap[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodMap[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	constraintInput, ok := convertToMapConstraintValue[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input), "map constraint type", any(input), ctx[0],
		)
	}

	return engine.ParseComplexStrict[map[any]any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeMap,
		z.extractMapForEngine,
		z.extractMapPtrForEngine,
		z.validateMapForEngine,
		ctx...,
	)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodMap[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (z *ZodMap[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// KeyType returns the key schema for this map.
func (z *ZodMap[T, R]) KeyType() any { return z.internals.KeyType }

// ValueType returns the value schema for this map.
func (z *ZodMap[T, R]) ValueType() any { return z.internals.ValueType }

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a schema that accepts nil, with constraint type *T.
func (z *ZodMap[T, R]) Optional() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil, with constraint type *T.
func (z *ZodMap[T, R]) Nilable() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility.
func (z *ZodMap[T, R]) Nullish() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the Optional flag and returns base constraint type.
func (z *ZodMap[T, R]) NonOptional() *ZodMap[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodMap[T, T]{internals: &ZodMapInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

// Default sets a default value returned when input is nil, bypassing validation.
func (z *ZodMap[T, R]) Default(v T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory function that provides the default value when input is nil.
func (z *ZodMap[T, R]) DefaultFunc(fn func() T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a prefault value that goes through the full parsing pipeline when input is nil.
func (z *ZodMap[T, R]) Prefault(v T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a factory function that provides the prefault value through the full parsing pipeline.
func (z *ZodMap[T, R]) PrefaultFunc(fn func() T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this map schema.
func (z *ZodMap[T, R]) Meta(meta core.GlobalMeta) *ZodMap[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodMap[T, R]) Describe(description string) *ZodMap[T, R] {
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

// Min sets the minimum number of entries.
func (z *ZodMap[T, R]) Min(minLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}
	return z.withCheck(checks.MinSize(minLen, errorMessage))
}

// Max sets the maximum number of entries.
func (z *ZodMap[T, R]) Max(maxLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}
	return z.withCheck(checks.MaxSize(maxLen, errorMessage))
}

// Size sets the exact number of entries required.
func (z *ZodMap[T, R]) Size(exactLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}
	return z.withCheck(checks.Size(exactLen, errorMessage))
}

// NonEmpty ensures the map has at least one entry. Equivalent to Min(1).
func (z *ZodMap[T, R]) NonEmpty(params ...any) *ZodMap[T, R] {
	return z.Min(1, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed map value.
func (z *ZodMap[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractMapValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the value while preserving the original type.
func (z *ZodMap[T, R]) Overwrite(transform func(R) R, params ...any) *ZodMap[T, R] {
	transformAny := func(input any) any {
		converted, ok := convertToMapType[T, R](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	return z.withCheck(checks.NewZodCheckOverwrite(transformAny, params...))
}

// Pipe creates a pipeline that passes the parsed result to a target schema.
func (z *ZodMap[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractMapValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a typed custom validation function.
func (z *ZodMap[T, R]) Refine(fn func(R) bool, params ...any) *ZodMap[T, R] {
	wrapper := func(v any) bool {
		var zero R
		switch any(zero).(type) {
		case *map[any]any:
			if v == nil {
				return true
			}
			if constraintValue, ok := convertToMapConstraintValue[T, R](v); ok {
				return fn(constraintValue)
			}
			return false
		default:
			if constraintValue, ok := convertToMapConstraintValue[T, R](v); ok {
				return fn(constraintValue)
			}
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}
	return z.withCheck(checks.NewCustom[any](wrapper, errorMessage))
}

// RefineAny adds a flexible custom validation function accepting any input.
func (z *ZodMap[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}
	return z.withCheck(checks.NewCustom[any](fn, errorMessage))
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// withPtrInternals creates a new instance with pointer constraint type *T.
func (z *ZodMap[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodMap[T, *T] {
	return &ZodMap[T, *T]{internals: &ZodMapInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

// withInternals creates a new instance preserving the generic constraint type R.
func (z *ZodMap[T, R]) withInternals(in *core.ZodTypeInternals) *ZodMap[T, R] {
	return &ZodMap[T, R]{internals: &ZodMapInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodMap[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodMap[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertMapFromGeneric converts map[any]any to the constraint type R.
func convertMapFromGeneric[T any, R any](mapValue map[any]any) R {
	var baseValue T
	switch any(baseValue).(type) {
	case map[string]int:
		converted := make(map[string]int, len(mapValue))
		for k, v := range mapValue {
			if strKey, ok := k.(string); ok {
				if intVal, ok := v.(int); ok {
					converted[strKey] = intVal
				}
			}
		}
		baseValue = any(converted).(T)
	case map[string]any:
		converted := make(map[string]any, len(mapValue))
		for k, v := range mapValue {
			if strKey, ok := k.(string); ok {
				converted[strKey] = v
			}
		}
		baseValue = any(converted).(T)
	case map[any]any:
		baseValue = any(mapValue).(T)
	default:
		baseValue = any(mapValue).(T)
	}
	return convertToMapConstraintType[T, R](baseValue)
}

// convertToMapConstraintType converts base type T to constraint type R.
func convertToMapConstraintType[T any, R any](value T) R {
	var zero R
	switch any(zero).(type) {
	case *map[any]any:
		if mapVal, ok := any(value).(map[any]any); ok {
			mapCopy := mapVal
			return any(&mapCopy).(R)
		}
		return any((*map[any]any)(nil)).(R)
	default:
		return any(value).(R)
	}
}

// extractMapValue extracts the base type T from constraint type R.
func extractMapValue[T any, R any](value R) T {
	if directValue, ok := any(value).(T); ok {
		return directValue
	}
	if ptrValue := reflect.ValueOf(value); ptrValue.Kind() == reflect.Pointer && !ptrValue.IsNil() {
		if derefValue, ok := ptrValue.Elem().Interface().(T); ok {
			return derefValue
		}
	}
	var zero T
	return zero
}

// convertToMapType converts any value to the map constraint type R.
func convertToMapType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		zeroType := reflect.TypeFor[R]()
		if zeroType.Kind() == reflect.Pointer {
			return zero, true
		}
		return zero, false
	}

	var mapValue map[any]any
	var isValid bool

	switch val := v.(type) {
	case map[any]any:
		mapValue, isValid = val, true
	case *map[any]any:
		if val != nil {
			mapValue, isValid = *val, true
		}
	case map[string]any:
		mapValue = make(map[any]any, len(val))
		for k, v := range val {
			mapValue[k] = v
		}
		isValid = true
	case *map[string]any:
		if val != nil {
			mapValue = make(map[any]any, len(*val))
			for k, v := range *val {
				mapValue[k] = v
			}
			isValid = true
		}
	default:
		return zero, false
	}

	if !isValid {
		return zero, false
	}

	zeroType := reflect.TypeFor[R]()
	if zeroType.Kind() == reflect.Pointer {
		if converted, ok := any(&mapValue).(R); ok {
			return converted, true
		}
	} else {
		if converted, ok := any(mapValue).(R); ok {
			return converted, true
		}
	}

	return zero, false
}

// convertToMapConstraintValue converts any value to constraint type R if possible.
func convertToMapConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	if r, ok := any(value).(R); ok { //nolint:unconvert // Required for generic type constraint conversion
		return r, true
	}

	// Handle pointer conversion: map[any]any -> *map[any]any
	if _, ok := any(zero).(*map[any]any); ok {
		if mapVal, ok := value.(map[any]any); ok {
			mapCopy := mapVal
			return any(&mapCopy).(R), true
		}
	}

	return zero, false
}

// =============================================================================
// EXTRACTION AND VALIDATION
// =============================================================================

// extractMap converts input to map[any]any.
func (z *ZodMap[T, R]) extractMap(value any, ctx *core.ParseContext) (map[any]any, error) {
	if mapVal, ok := value.(map[any]any); ok {
		return mapVal, nil
	}
	if mapPtr, ok := value.(*map[any]any); ok {
		if mapPtr != nil {
			return *mapPtr, nil
		}
		return nil, issues.CreateNonOptionalError(ctx)
	}
	if reflectx.IsMap(value) {
		if converted, err := mapx.ToGeneric(value); err == nil && converted != nil {
			return converted, nil
		}
	}
	return nil, issues.CreateInvalidTypeError(core.ZodTypeMap, value, ctx)
}

// validateMap validates map entries and collects all errors (Zod v4 multi-error behavior).
func (z *ZodMap[T, R]) validateMap(value map[any]any, chks []core.ZodCheck, ctx *core.ParseContext) (map[any]any, error) {
	// Apply checks (including Overwrite transformations) first.
	transformedValue, err := engine.ApplyChecks[any](value, chks, ctx)
	if err != nil {
		return nil, err
	}

	switch v := transformedValue.(type) {
	case map[any]any:
		value = v
	case *map[any]any:
		if v != nil {
			value = *v
		}
	default:
		// Keep original value if transformation returned unexpected type.
	}

	var collectedIssues []core.ZodRawIssue

	for key, val := range value {
		if z.internals.KeyType != nil {
			collectedIssues = z.collectValidationErrors(key, z.internals.KeyType, key, ctx, collectedIssues)
		}
		if z.internals.ValueType != nil {
			collectedIssues = z.collectValidationErrors(val, z.internals.ValueType, key, ctx, collectedIssues)
		}
	}

	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}
	return value, nil
}

// collectValidationErrors validates a value against a schema and appends any issues with the given path key.
func (z *ZodMap[T, R]) collectValidationErrors(value, schema, pathKey any, ctx *core.ParseContext, dst []core.ZodRawIssue) []core.ZodRawIssue {
	err := z.validateValueDirect(value, schema, ctx)
	if err == nil {
		return dst
	}

	var zodErr *issues.ZodError
	if errors.As(err, &zodErr) {
		for _, issue := range zodErr.Issues {
			dst = append(dst, issues.ConvertZodIssueToRawWithPrependedPath(issue, []any{pathKey}))
		}
	} else {
		rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, value)
		rawIssue.Path = []any{pathKey}
		dst = append(dst, rawIssue)
	}
	return dst
}

// validateValueDirect validates a single value against a schema using reflection.
// It preserves original error codes without wrapping.
func (z *ZodMap[T, R]) validateValueDirect(value any, schema any, ctx *core.ParseContext) error {
	if schema == nil {
		return nil
	}
	if ctx == nil {
		ctx = core.NewParseContext()
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

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// newZodMapFromDef constructs a new ZodMap from the given definition.
func newZodMapFromDef[T any, R any](def *ZodMapDef) *ZodMap[T, R] {
	internals := &ZodMapInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:       def,
		KeyType:   def.KeyType,
		ValueType: def.ValueType,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		return any(newZodMapFromDef[T, R](&ZodMapDef{
			ZodTypeDef: *newDef,
			KeyType:    def.KeyType,
			ValueType:  def.ValueType,
		})).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}
	return &ZodMap[T, R]{internals: internals}
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// Map creates a map schema with key and value validation returning a value constraint.
func Map(keySchema, valueSchema any, paramArgs ...any) *ZodMap[map[any]any, map[any]any] {
	return MapTyped[map[any]any, map[any]any](keySchema, valueSchema, paramArgs...)
}

// MapPtr creates a map schema with key and value validation returning a pointer constraint.
func MapPtr(keySchema, valueSchema any, paramArgs ...any) *ZodMap[map[any]any, *map[any]any] {
	return MapTyped[map[any]any, *map[any]any](keySchema, valueSchema, paramArgs...)
}

// MapTyped creates a typed map schema with explicit generic constraints.
func MapTyped[T any, R any](keySchema, valueSchema any, paramArgs ...any) *ZodMap[T, R] {
	param := utils.FirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodMapDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeMap,
			Checks: []core.ZodCheck{},
		},
		KeyType:   keySchema,
		ValueType: valueSchema,
	}
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	mapSchema := newZodMapFromDef[T, R](def)

	// Register a pass-through check so the engine invokes validateMap for key/value validation.
	if keySchema != nil || valueSchema != nil {
		mapSchema.internals.AddCheck(checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{}))
	}
	return mapSchema
}

// Check adds a custom validation function that can report multiple issues.
func (z *ZodMap[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodMap[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Handle pointer/value mismatch when R is *map but value is map.
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
func (z *ZodMap[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodMap[T, R] {
	return z.Check(fn, params...)
}

// =============================================================================
// ENGINE CALLBACKS
// =============================================================================

// extractMapForEngine extracts map[any]any from input for engine.ParseComplex.
func (z *ZodMap[T, R]) extractMapForEngine(input any) (map[any]any, bool) {
	result, err := z.extractMap(input, core.NewParseContext())
	return result, err == nil
}

// extractMapPtrForEngine extracts pointer to map[any]any for engine.ParseComplex.
func (z *ZodMap[T, R]) extractMapPtrForEngine(input any) (*map[any]any, bool) {
	if ptr, ok := input.(*map[any]any); ok {
		return ptr, true
	}
	result, err := z.extractMap(input, core.NewParseContext())
	if err != nil {
		return nil, false
	}
	return &result, true
}

// validateMapForEngine delegates to validateMap for engine.ParseComplex.
func (z *ZodMap[T, R]) validateMapForEngine(value map[any]any, chks []core.ZodCheck, ctx *core.ParseContext) (map[any]any, error) {
	return z.validateMap(value, chks, ctx)
}
