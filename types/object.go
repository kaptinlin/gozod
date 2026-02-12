package types

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// Sentinel errors for object schema operations.
var (
	ErrNilPointerCannotConvertToObject = errors.New("nil pointer cannot be converted to object")
	ErrPickRefinements                 = errors.New("Pick cannot be used on object schemas containing refinements")
	ErrOmitRefinements                 = errors.New("Omit cannot be used on object schemas containing refinements")
	ErrExtendRefinements               = errors.New("Extend cannot overwrite keys on object schemas containing refinements, use SafeExtend instead")
	ErrUnrecognizedKey                 = errors.New("unrecognized key")
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ObjectMode defines how to handle unknown keys in object validation.
type ObjectMode string

const (
	ObjectModeStrict      ObjectMode = "strict"
	ObjectModeStrip       ObjectMode = "strip"
	ObjectModePassthrough ObjectMode = "passthrough"
)

// ZodObjectDef defines the configuration for an object schema.
type ZodObjectDef struct {
	core.ZodTypeDef
	Shape       core.ObjectSchema
	Catchall    core.ZodSchema
	UnknownKeys ObjectMode
}

// ZodObjectInternals contains object validator internal state.
type ZodObjectInternals struct {
	core.ZodTypeInternals
	Def                *ZodObjectDef
	Shape              core.ObjectSchema
	Catchall           core.ZodSchema
	UnknownKeys        ObjectMode
	IsPartial          bool
	PartialExceptions  map[string]bool
	HasUserRefinements bool
}

// ZodObject represents a type-safe object validation schema.
// T is the base type, R is the constraint type (value or pointer).
type ZodObject[T any, R any] struct {
	internals *ZodObjectInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema.
func (z *ZodObject[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetUnknownKeys returns the unknown keys handling mode.
func (z *ZodObject[T, R]) GetUnknownKeys() ObjectMode { return z.internals.UnknownKeys }

// GetCatchall returns the catchall schema for unknown fields.
func (z *ZodObject[T, R]) GetCatchall() core.ZodSchema { return z.internals.Catchall }

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodObject[T, R]) IsOptional() bool { return z.internals.IsOptional() }

// IsNilable reports whether this schema accepts nil values.
func (z *ZodObject[T, R]) IsNilable() bool { return z.internals.IsNilable() }

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodObject[T, R]) withCheck(check core.ZodCheck) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// Parse validates input using object-specific parsing logic with engine.ParseComplex.
func (z *ZodObject[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

	result, err := engine.ParseComplex[map[string]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeObject,
		z.extractObjectForEngine,
		z.extractObjectPtrForEngine,
		z.validateObjectForEngine,
		ctx...,
	)
	if err != nil {
		return zero, err
	}

	switch v := result.(type) {
	case map[string]any:
		return convertToObjectConstraintType[T, R](any(v).(T)), nil
	case *map[string]any:
		if v == nil {
			return zero, nil
		}
		return convertToObjectConstraintType[T, R](any(*v).(T)), nil
	case nil:
		return zero, nil
	default:
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", result), fmt.Sprintf("%T", zero), result, &core.ParseContext{},
		)
	}
}

// MustParse validates input and panics on error.
func (z *ZodObject[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodObject[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	constraintInput := convertToObjectConstraintType[T, R](input)

	return engine.ParseComplexStrict[map[string]any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeObject,
		z.extractObjectForEngine,
		z.extractObjectPtrForEngine,
		z.validateObjectForEngine,
		ctx...,
	)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodObject[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (z *ZodObject[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a schema that accepts nil, with constraint type *T.
func (z *ZodObject[T, R]) Optional() *ZodObject[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil, with constraint type *T.
func (z *ZodObject[T, R]) Nilable() *ZodObject[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodObject[T, R]) Nullish() *ZodObject[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag and returns base constraint type.
func (z *ZodObject[T, R]) NonOptional() *ZodObject[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodObject[T, T]{internals: z.newObjectInternals(in)}
}

// Default sets a default value returned when input is nil, bypassing validation.
func (z *ZodObject[T, R]) Default(v T) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory function that provides the default value when input is nil.
func (z *ZodObject[T, R]) DefaultFunc(fn func() T) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a prefault value that goes through the full parsing pipeline when input is nil.
func (z *ZodObject[T, R]) Prefault(v T) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a factory function that provides the prefault value through the full parsing pipeline.
func (z *ZodObject[T, R]) PrefaultFunc(fn func() T) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta attaches GlobalMeta to this object schema via the global registry.
func (z *ZodObject[T, R]) Meta(meta core.GlobalMeta) *ZodObject[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodObject[T, R]) Describe(description string) *ZodObject[T, R] {
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

// Min sets the minimum number of fields.
func (z *ZodObject[T, R]) Min(minLen int, params ...any) *ZodObject[T, R] {
	return z.withCheck(checks.MinSize(minLen, params...))
}

// Max sets the maximum number of fields.
func (z *ZodObject[T, R]) Max(maxLen int, params ...any) *ZodObject[T, R] {
	return z.withCheck(checks.MaxSize(maxLen, params...))
}

// Size sets the exact number of fields required.
func (z *ZodObject[T, R]) Size(exactLen int, params ...any) *ZodObject[T, R] {
	return z.withCheck(checks.Size(exactLen, params...))
}

// Property validates a specific property using the provided schema.
func (z *ZodObject[T, R]) Property(key string, schema core.ZodSchema, params ...any) *ZodObject[T, R] {
	return z.withCheck(checks.NewProperty(key, schema, params...).GetZod())
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Shape returns the object shape (field schemas).
func (z *ZodObject[T, R]) Shape() core.ObjectSchema { return maps.Clone(z.internals.Shape) }

// Properties is an alias for Shape.
func (z *ZodObject[T, R]) Properties() core.ObjectSchema { return z.Shape() }

// Pick creates a new object with only the specified keys.
// Returns an error if any key does not exist or the schema contains refinements.
func (z *ZodObject[T, R]) Pick(keys []string, params ...any) (*ZodObject[T, R], error) {
	if z.hasRefinements() {
		return nil, ErrPickRefinements
	}

	newShape := make(core.ObjectSchema, len(keys))
	for _, key := range keys {
		schema, exists := z.internals.Shape[key]
		if !exists {
			return nil, fmt.Errorf("%w: %q", ErrUnrecognizedKey, key)
		}
		newShape[key] = schema
	}
	return ObjectTyped[T, R](newShape, params...), nil
}

// MustPick is like Pick but panics on error.
func (z *ZodObject[T, R]) MustPick(keys []string, params ...any) *ZodObject[T, R] {
	result, err := z.Pick(keys, params...)
	if err != nil {
		panic(err)
	}
	return result
}

// Omit creates a new object excluding the specified keys.
// Returns an error if any key does not exist or the schema contains refinements.
func (z *ZodObject[T, R]) Omit(keys []string, params ...any) (*ZodObject[T, R], error) {
	if z.hasRefinements() {
		return nil, ErrOmitRefinements
	}

	excludeSet := make(map[string]bool, len(keys))
	for _, key := range keys {
		if _, exists := z.internals.Shape[key]; !exists {
			return nil, fmt.Errorf("%w: %q", ErrUnrecognizedKey, key)
		}
		excludeSet[key] = true
	}

	newShape := make(core.ObjectSchema, len(z.internals.Shape)-len(keys))
	for key, schema := range z.internals.Shape {
		if !excludeSet[key] {
			newShape[key] = schema
		}
	}
	return ObjectTyped[T, R](newShape, params...), nil
}

// MustOmit is like Omit but panics on error.
func (z *ZodObject[T, R]) MustOmit(keys []string, params ...any) *ZodObject[T, R] {
	result, err := z.Omit(keys, params...)
	if err != nil {
		panic(err)
	}
	return result
}

// hasRefinements reports whether the schema has user-defined refinements.
func (z *ZodObject[T, R]) hasRefinements() bool { return z.internals.HasUserRefinements }

// Extend creates a new object by extending with additional fields.
// Returns an error if the schema has refinements and augmentation overlaps existing keys.
func (z *ZodObject[T, R]) Extend(augmentation core.ObjectSchema, params ...any) (*ZodObject[T, R], error) {
	if z.internals.HasUserRefinements {
		for k := range augmentation {
			if _, exists := z.internals.Shape[k]; exists {
				return nil, ErrExtendRefinements
			}
		}
	}

	newShape := maps.Clone(z.internals.Shape)
	maps.Copy(newShape, augmentation)
	return ObjectTyped[T, R](newShape, params...), nil
}

// SafeExtend creates a new object by extending without checking refinements.
func (z *ZodObject[T, R]) SafeExtend(augmentation core.ObjectSchema, params ...any) *ZodObject[T, R] {
	newShape := maps.Clone(z.internals.Shape)
	maps.Copy(newShape, augmentation)
	return ObjectTyped[T, R](newShape, params...)
}

// Merge combines two object schemas, clearing all checks/refinements.
func (z *ZodObject[T, R]) Merge(other *ZodObject[T, R], params ...any) *ZodObject[T, R] {
	newShape := maps.Clone(z.internals.Shape)
	maps.Copy(newShape, other.internals.Shape)
	return ObjectTyped[T, R](newShape, params...)
}

// Partial makes all fields optional, or specific fields if keys are provided.
func (z *ZodObject[T, R]) Partial(keys ...[]string) *ZodObject[T, R] {
	newInternals := z.internals.Clone()

	var partialExceptions map[string]bool
	if len(keys) > 0 && len(keys[0]) > 0 {
		partialExceptions = make(map[string]bool, len(z.internals.Shape))
		for fieldName := range z.internals.Shape {
			partialExceptions[fieldName] = true
		}
		for _, key := range keys[0] {
			delete(partialExceptions, key)
		}
	}

	oi := z.newObjectInternals(newInternals)
	oi.IsPartial = true
	oi.PartialExceptions = partialExceptions
	return &ZodObject[T, R]{internals: oi}
}

// Required makes all fields required, or specific fields if provided.
func (z *ZodObject[T, R]) Required(fields ...[]string) *ZodObject[T, R] {
	newInternals := z.internals.Clone()

	var partialExceptions map[string]bool
	if len(fields) > 0 && len(fields[0]) > 0 {
		partialExceptions = make(map[string]bool, len(fields[0]))
		for _, fieldName := range fields[0] {
			partialExceptions[fieldName] = true
		}
	}

	oi := z.newObjectInternals(newInternals)
	oi.IsPartial = true
	oi.PartialExceptions = partialExceptions
	return &ZodObject[T, R]{internals: oi}
}

// Strict sets strict mode (rejects unknown keys).
func (z *ZodObject[T, R]) Strict() *ZodObject[T, R] { return z.withUnknownKeys(ObjectModeStrict) }

// Strip sets strip mode (removes unknown keys).
func (z *ZodObject[T, R]) Strip() *ZodObject[T, R] { return z.withUnknownKeys(ObjectModeStrip) }

// Passthrough sets passthrough mode (allows unknown keys).
func (z *ZodObject[T, R]) Passthrough() *ZodObject[T, R] {
	return z.withUnknownKeys(ObjectModePassthrough)
}

// Catchall sets a schema for validating unknown keys.
func (z *ZodObject[T, R]) Catchall(catchallSchema core.ZodSchema) *ZodObject[T, R] {
	newInternals := z.internals.Clone()
	oi := z.newObjectInternals(newInternals)
	oi.Catchall = catchallSchema
	return &ZodObject[T, R]{internals: oi}
}

// Keyof returns a string enum schema of all keys.
func (z *ZodObject[T, R]) Keyof() *ZodEnum[string, string] {
	keys := make([]string, 0, len(z.internals.Shape))
	for key := range z.internals.Shape {
		keys = append(keys, key)
	}
	return EnumSlice(keys)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed object value.
func (z *ZodObject[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractObjectValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the value while preserving the original type.
func (z *ZodObject[T, R]) Overwrite(transform func(R) R, params ...any) *ZodObject[T, R] {
	transformAny := func(input any) any {
		converted, ok := convertToObjectType[T, R](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	return z.withCheck(checks.NewZodCheckOverwrite(transformAny, params...))
}

// Pipe creates a pipeline to another schema.
func (z *ZodObject[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractObjectValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation using constraint type.
// Schemas with refinements cannot use Pick/Omit.
func (z *ZodObject[T, R]) Refine(fn func(R) bool, params ...any) *ZodObject[T, R] {
	wrapper := func(v any) bool {
		if v == nil {
			return true
		}
		if constraintVal, ok := convertToObjectType[T, R](v); ok {
			return fn(constraintVal)
		}
		return false
	}

	param := utils.FirstParam(params...)
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(param))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	result := z.withInternals(newInternals)
	result.internals.HasUserRefinements = true
	return result
}

// RefineAny provides flexible validation without type conversion.
// Schemas with refinements cannot use Pick/Omit.
func (z *ZodObject[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodObject[T, R] {
	param := utils.FirstParam(params...)
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(param))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	result := z.withInternals(newInternals)
	result.internals.HasUserRefinements = true
	return result
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodObject[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodObject[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// newObjectInternals creates a new ZodObjectInternals copying object-specific state from z.
func (z *ZodObject[T, R]) newObjectInternals(in *core.ZodTypeInternals) *ZodObjectInternals {
	return &ZodObjectInternals{
		ZodTypeInternals:   *in,
		Def:                z.internals.Def,
		Shape:              z.internals.Shape,
		Catchall:           z.internals.Catchall,
		UnknownKeys:        z.internals.UnknownKeys,
		IsPartial:          z.internals.IsPartial,
		PartialExceptions:  z.internals.PartialExceptions,
		HasUserRefinements: z.internals.HasUserRefinements,
	}
}

// withPtrInternals creates a new instance with pointer constraint type.
func (z *ZodObject[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodObject[T, *T] {
	return &ZodObject[T, *T]{internals: z.newObjectInternals(in)}
}

// withInternals creates a new instance preserving the constraint type.
func (z *ZodObject[T, R]) withInternals(in *core.ZodTypeInternals) *ZodObject[T, R] {
	return &ZodObject[T, R]{internals: z.newObjectInternals(in)}
}

// withUnknownKeys creates a new instance with the specified unknown keys mode.
func (z *ZodObject[T, R]) withUnknownKeys(mode ObjectMode) *ZodObject[T, R] {
	newInternals := z.internals.Clone()
	oi := z.newObjectInternals(newInternals)
	oi.UnknownKeys = mode
	return &ZodObject[T, R]{internals: oi}
}

// CloneFrom copies configuration from another schema.
func (z *ZodObject[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodObject[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// convertToObjectConstraintType converts base type T to constraint type R.
func convertToObjectConstraintType[T any, R any](value T) R {
	var result R
	resultVal := reflect.ValueOf(&result).Elem()
	if resultVal.Kind() == reflect.Pointer {
		if reflect.TypeOf(value).AssignableTo(resultVal.Type().Elem()) {
			newVal := reflect.New(resultVal.Type().Elem())
			newVal.Elem().Set(reflect.ValueOf(value))
			resultVal.Set(newVal)
		}
	} else {
		if reflect.TypeOf(value).AssignableTo(resultVal.Type()) {
			resultVal.Set(reflect.ValueOf(value))
		}
	}

	return result
}

// extractObjectValue extracts the base value from constraint type R.
func extractObjectValue[T any, R any](value R) T {
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

// convertToObjectType converts any value to the object constraint type R.
func convertToObjectType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		if reflect.TypeFor[R]().Kind() == reflect.Pointer {
			return zero, true
		}
		return zero, false
	}

	var objectValue map[string]any
	var isValid bool

	switch val := v.(type) {
	case map[string]any:
		objectValue, isValid = val, true
	case *map[string]any:
		if val != nil {
			objectValue, isValid = *val, true
		}
	case map[any]any:
		objectValue = make(map[string]any, len(val))
		for k, v := range val {
			strKey, ok := k.(string)
			if !ok {
				return zero, false
			}
			objectValue[strKey] = v
		}
		isValid = true
	default:
		return zero, false
	}

	if !isValid {
		return zero, false
	}

	rType := reflect.TypeFor[R]()
	if rType.Kind() == reflect.Pointer {
		if converted, ok := any(&objectValue).(R); ok {
			return converted, true
		}
	} else {
		if converted, ok := any(objectValue).(R); ok {
			return converted, true
		}
	}

	return zero, false
}

// extractObject extracts map[string]any from value with proper conversion.
func (z *ZodObject[T, R]) extractObject(value any) (map[string]any, error) {
	if objectVal, ok := value.(map[string]any); ok {
		return objectVal, nil
	}
	if objectPtr, ok := value.(*map[string]any); ok && objectPtr != nil {
		return *objectPtr, nil
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, ErrNilPointerCannotConvertToObject
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		result := make(map[string]any, rv.NumField())
		rt := rv.Type()
		for i := range rv.NumField() {
			field := rt.Field(i)
			if field.IsExported() {
				name := field.Name
				if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
					if commaIdx := strings.Index(tag, ","); commaIdx > 0 {
						name = tag[:commaIdx]
					} else {
						name = tag
					}
				}
				result[name] = rv.Field(i).Interface()
			}
		}
		return result, nil
	}

	return nil, issues.CreateTypeConversionError(
		fmt.Sprintf("%T", value), "map[string]any", value, core.NewParseContext(),
	)
}

// collectFieldErrors appends validation errors from err to collectedIssues with the given field path.
func collectFieldErrors(err error, fieldName string, collectedIssues *[]core.ZodRawIssue, fieldValue any) {
	var zodErr *issues.ZodError
	if errors.As(err, &zodErr) {
		for _, fieldIssue := range zodErr.Issues {
			*collectedIssues = append(*collectedIssues, issues.ConvertZodIssueToRawWithPrependedPath(fieldIssue, []any{fieldName}))
		}
	} else {
		rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, fieldValue)
		rawIssue.Path = []any{fieldName}
		*collectedIssues = append(*collectedIssues, rawIssue)
	}
}

// validateObject validates object fields and collects all errors.
func (z *ZodObject[T, R]) validateObject(value map[string]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[string]any, error) {
	var collectedIssues []core.ZodRawIssue
	resultObject := make(map[string]any, len(z.internals.Shape))

	for fieldName, fieldSchema := range z.internals.Shape {
		fieldValue, exists := value[fieldName]

		if !exists {
			if !z.isFieldOptional(fieldSchema, fieldName) {
				rawIssue := issues.CreateIssue(core.InvalidType, fmt.Sprintf("missing required field: %s", fieldName), map[string]any{
					"expected": "nonoptional",
					"received": "undefined",
				}, nil)
				rawIssue.Path = []any{fieldName}
				collectedIssues = append(collectedIssues, rawIssue)
			}
			continue
		}

		if fieldValue == nil && z.isFieldExactOptional(fieldSchema) {
			rawIssue := issues.CreateIssue(core.InvalidType, fmt.Sprintf("field %s cannot be explicitly nil (use absent key instead)", fieldName), map[string]any{
				"expected": "string",
				"received": "nil",
			}, nil)
			rawIssue.Path = []any{fieldName}
			collectedIssues = append(collectedIssues, rawIssue)
			continue
		}

		if err := z.validateField(fieldValue, fieldSchema, ctx); err != nil {
			collectFieldErrors(err, fieldName, &collectedIssues, fieldValue)
		} else {
			resultObject[fieldName] = fieldValue
		}
	}

	var unrecognizedKeys []string
	for fieldName, fieldValue := range value {
		if _, isKnown := z.internals.Shape[fieldName]; isKnown {
			continue
		}
		switch z.internals.UnknownKeys {
		case ObjectModeStrict:
			unrecognizedKeys = append(unrecognizedKeys, fieldName)
		case ObjectModeStrip:
			// Strip mode: omit unknown fields from result.
		case ObjectModePassthrough:
			if z.internals.Catchall != nil {
				if err := z.validateField(fieldValue, z.internals.Catchall, ctx); err != nil {
					collectFieldErrors(err, fieldName, &collectedIssues, fieldValue)
				} else {
					resultObject[fieldName] = fieldValue
				}
			} else {
				resultObject[fieldName] = fieldValue
			}
		}
	}

	if len(unrecognizedKeys) > 0 {
		rawIssue := issues.CreateIssue(core.UnrecognizedKeys, "", map[string]any{
			"keys": unrecognizedKeys,
		}, value)
		collectedIssues = append(collectedIssues, rawIssue)
	}

	if len(checks) > 0 {
		payload := core.NewParsePayload(resultObject)
		result := engine.RunChecksOnValue(resultObject, checks, payload, ctx)
		if result.HasIssues() {
			collectedIssues = append(collectedIssues, result.GetIssues()...)
		}
		if v := result.GetValue(); v != nil {
			if transformed, ok := v.(map[string]any); ok {
				resultObject = transformed
			}
		}
	}

	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}
	return resultObject, nil
}

// validateField validates a single field using reflection-based method dispatch.
func (z *ZodObject[T, R]) validateField(element any, schema any, ctx *core.ParseContext) error {
	if schema == nil {
		return nil
	}

	schemaValue := reflect.ValueOf(schema)
	if !schemaValue.IsValid() || schemaValue.IsNil() {
		return nil
	}

	parseMethod := schemaValue.MethodByName("ParseAny")
	if !parseMethod.IsValid() {
		return nil
	}

	methodType := parseMethod.Type()
	if methodType.NumIn() < 1 {
		return nil
	}

	var firstArg reflect.Value
	if element == nil {
		firstArg = reflect.Zero(methodType.In(0))
	} else {
		firstArg = reflect.ValueOf(element)
	}
	args := []reflect.Value{firstArg}
	if methodType.NumIn() > 1 && methodType.In(1).String() == "*core.ParseContext" {
		if ctx == nil {
			args = append(args, reflect.Zero(methodType.In(1)))
		} else {
			args = append(args, reflect.ValueOf(ctx))
		}
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

// internalsProvider is a local interface for type-safe access to schema internals.
type internalsProvider interface {
	GetInternals() *core.ZodTypeInternals
}

// isFieldOptional reports whether a field is optional based on schema or partial state.
func (z *ZodObject[T, R]) isFieldOptional(schema any, fieldName string) bool {
	if schema == nil {
		return true
	}
	if z.internals.IsPartial {
		if z.internals.PartialExceptions == nil {
			return true
		}
		if !z.internals.PartialExceptions[fieldName] {
			return true
		}
	}
	if ip, ok := schema.(internalsProvider); ok {
		return ip.GetInternals().Optional
	}
	return false
}

// isFieldExactOptional reports whether a field has ExactOptional set.
func (z *ZodObject[T, R]) isFieldExactOptional(schema any) bool {
	if schema == nil {
		return false
	}
	if ip, ok := schema.(internalsProvider); ok {
		return ip.GetInternals().ExactOptional
	}
	return false
}

// newZodObjectFromDef constructs a new ZodObject from a definition.
func newZodObjectFromDef[T any, R any](def *ZodObjectDef) *ZodObject[T, R] {
	internals := &ZodObjectInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Shape:            def.Shape,
		Catchall:         def.Catchall,
		UnknownKeys:      def.UnknownKeys,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		objectDef := &ZodObjectDef{
			ZodTypeDef:  *newDef,
			Shape:       def.Shape,
			Catchall:    def.Catchall,
			UnknownKeys: def.UnknownKeys,
		}
		return any(newZodObjectFromDef[T, R](objectDef)).(core.ZodType[any])
	}

	schema := &ZodObject[T, R]{internals: internals}
	if def.Error != nil {
		internals.Error = def.Error
	}
	for _, check := range def.Checks {
		internals.AddCheck(check)
	}
	return schema
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// ObjectTyped creates a typed object schema with explicit type parameters.
func ObjectTyped[T any, R any](shape core.ObjectSchema, params ...any) *ZodObject[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodObjectDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeObject,
			Checks: []core.ZodCheck{},
		},
		Shape:       make(core.ObjectSchema, len(shape)),
		UnknownKeys: ObjectModeStrip,
	}

	maps.Copy(def.Shape, shape)

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	objectSchema := newZodObjectFromDef[T, R](def)
	objectSchema.internals.AddCheck(
		checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{}),
	)
	return objectSchema
}

// Object creates an object schema with default types.
func Object(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return ObjectTyped[map[string]any, map[string]any](shape, params...)
}

// ObjectPtr creates an object schema with pointer constraint.
func ObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...)
}

// StrictObject creates a strict object schema that rejects unknown keys.
func StrictObject(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return Object(shape, params...).Strict()
}

// LooseObject creates a loose object schema that allows unknown keys.
func LooseObject(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return Object(shape, params...).Passthrough()
}

// StrictObjectPtr creates a strict object schema with pointer constraint.
func StrictObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...).Strict()
}

// LooseObjectPtr creates a loose object schema with pointer constraint.
func LooseObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...).Passthrough()
}

// Check adds a custom validation function that can report multiple issues.
func (z *ZodObject[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodObject[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		rType := reflect.TypeFor[R]()
		if rType.Kind() == reflect.Pointer {
			elemTyp := rType.Elem()
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
func (z *ZodObject[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodObject[T, R] {
	return z.Check(fn, params...)
}

// extractObjectForEngine extracts map[string]any from input for engine.ParseComplex.
func (z *ZodObject[T, R]) extractObjectForEngine(input any) (map[string]any, bool) {
	result, err := z.extractObject(input)
	if err != nil {
		return nil, false
	}
	return result, true
}

// extractObjectPtrForEngine extracts pointer to map[string]any from input.
func (z *ZodObject[T, R]) extractObjectPtrForEngine(input any) (*map[string]any, bool) {
	if ptr, ok := input.(*map[string]any); ok {
		return ptr, true
	}
	result, err := z.extractObject(input)
	if err != nil {
		return nil, false
	}
	return &result, true
}

// validateObjectForEngine validates map[string]any for engine.ParseComplex.
func (z *ZodObject[T, R]) validateObjectForEngine(value map[string]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[string]any, error) {
	return z.validateObject(value, checks, ctx)
}
