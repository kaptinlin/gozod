package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// Static error variables
var (
	ErrNilPointerCannotConvertToObject = errors.New("nil pointer cannot be converted to object")

	// Pick/Omit/Extend errors - Zod v4 compatible error messages
	// See: .reference/zod/packages/zod/src/v4/core/util.ts:599, 628, 664
	ErrPickRefinements   = errors.New(".pick() cannot be used on object schemas containing refinements")
	ErrOmitRefinements   = errors.New(".omit() cannot be used on object schemas containing refinements")
	ErrExtendRefinements = errors.New(".extend() cannot overwrite keys on object schemas containing refinements. Use .SafeExtend() instead")
	ErrUnrecognizedKey   = errors.New("unrecognized key")
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ObjectMode defines how to handle unknown keys in object validation
type ObjectMode string

const (
	ObjectModeStrict      ObjectMode = "strict"      // Reject unknown keys
	ObjectModeStrip       ObjectMode = "strip"       // Remove unknown keys
	ObjectModePassthrough ObjectMode = "passthrough" // Allow unknown keys
)

// ZodObjectDef defines the schema definition for object validation
type ZodObjectDef struct {
	core.ZodTypeDef
	Shape       core.ObjectSchema // Field schemas
	Catchall    core.ZodSchema    // Schema for unrecognized keys
	UnknownKeys ObjectMode        // How to handle unknown keys
}

// ZodObjectInternals contains the internal state for object schema
type ZodObjectInternals struct {
	core.ZodTypeInternals
	Def                *ZodObjectDef     // Schema definition reference
	Shape              core.ObjectSchema // Field schemas for runtime validation
	Catchall           core.ZodSchema    // Catchall schema for unknown fields
	UnknownKeys        ObjectMode        // Validation mode
	IsPartial          bool              // Whether this is a partial object (all fields optional)
	PartialExceptions  map[string]bool   // Fields that should remain required in partial mode
	HasUserRefinements bool              // Whether user has added refinements via Refine/RefineAny/SuperRefine
}

// ZodObject represents a type-safe object validation schema
type ZodObject[T any, R any] struct {
	internals *ZodObjectInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodObject[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetUnknownKeys returns the unknown keys handling mode for JSON Schema conversion
func (z *ZodObject[T, R]) GetUnknownKeys() ObjectMode {
	return z.internals.UnknownKeys
}

// GetCatchall returns the catchall schema for JSON Schema conversion
func (z *ZodObject[T, R]) GetCatchall() core.ZodSchema {
	return z.internals.Catchall
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodObject[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodObject[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input using object-specific parsing logic with engine.ParseComplex
func (z *ZodObject[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
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
		var zero R
		return zero, err
	}

	// Convert result to constraint type R
	if objectMap, ok := result.(map[string]any); ok {
		return convertToObjectConstraintType[T, R](any(objectMap).(T)), nil
	}

	// Handle pointer to map[string]any
	if objectMapPtr, ok := result.(*map[string]any); ok {
		if objectMapPtr == nil {
			var zero R
			return zero, nil
		}
		return convertToObjectConstraintType[T, R](any(*objectMapPtr).(T)), nil
	}

	// Handle nil result for optional/nilable schemas
	if result == nil {
		var zero R
		return zero, nil
	}

	// This should not happen in well-formed schemas
	var zero R
	return zero, issues.CreateTypeConversionError(
		fmt.Sprintf("%T", result),
		fmt.Sprintf("%T", zero),
		result,
		&core.ParseContext{},
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodObject[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety and enhanced performance.
// This method provides zero-overhead abstraction with strict type constraints.
func (z *ZodObject[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	// Convert T to R for ParseComplexStrict
	constraintInput := convertToObjectConstraintType[T, R](input)

	result, err := engine.ParseComplexStrict[map[string]any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeObject,
		z.extractObjectForEngine,
		z.extractObjectPtrForEngine,
		z.validateObjectForEngine,
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
func (z *ZodObject[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodObject[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates optional object schema
func (z *ZodObject[T, R]) Optional() *ZodObject[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values
func (z *ZodObject[T, R]) Nilable() *ZodObject[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodObject[T, R]) Nullish() *ZodObject[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag and enforces non-nil value type.
// It returns a schema whose constraint type is the base value (T), mirroring
// the behaviour of .Optional().NonOptional() chain in TypeScript Zod.
func (z *ZodObject[T, R]) NonOptional() *ZodObject[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodObject[T, T]{
		internals: &ZodObjectInternals{
			ZodTypeInternals:   *in,
			Def:                z.internals.Def,
			Shape:              z.internals.Shape,
			Catchall:           z.internals.Catchall,
			UnknownKeys:        z.internals.UnknownKeys,
			IsPartial:          z.internals.IsPartial,
			PartialExceptions:  z.internals.PartialExceptions,
			HasUserRefinements: z.internals.HasUserRefinements,
		},
	}
}

// Default preserves current type
func (z *ZodObject[T, R]) Default(v T) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current type
func (z *ZodObject[T, R]) DefaultFunc(fn func() T) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodObject[T, R]) Prefault(v T) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodObject[T, R]) PrefaultFunc(fn func() T) *ZodObject[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta attaches GlobalMeta to this object schema via the global registry.
func (z *ZodObject[T, R]) Meta(meta core.GlobalMeta) *ZodObject[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
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

// Min sets minimum number of fields
func (z *ZodObject[T, R]) Min(minLen int, params ...any) *ZodObject[T, R] {
	check := checks.MinSize(minLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of fields
func (z *ZodObject[T, R]) Max(maxLen int, params ...any) *ZodObject[T, R] {
	check := checks.MaxSize(maxLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Size sets exact number of fields
func (z *ZodObject[T, R]) Size(exactLen int, params ...any) *ZodObject[T, R] {
	check := checks.Size(exactLen, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Property validates a specific property using the provided schema
func (z *ZodObject[T, R]) Property(key string, schema core.ZodSchema, params ...any) *ZodObject[T, R] {
	check := checks.NewProperty(key, schema, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check.GetZod())
	return z.withInternals(newInternals)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Shape returns the object shape (field schemas)
func (z *ZodObject[T, R]) Shape() core.ObjectSchema {
	result := make(core.ObjectSchema)
	for k, v := range z.internals.Shape {
		result[k] = v
	}
	return result
}

// Properties is an alias for Shape(), returning the object's field schemas.
// This is included for alignment with documentation and developer convenience.
func (z *ZodObject[T, R]) Properties() core.ObjectSchema {
	return z.Shape()
}

// Pick creates a new object with only specified keys.
// Returns error if any key does not exist or schema contains refinements.
// Zod v4 compatible: throws on refinements or unrecognized keys.
func (z *ZodObject[T, R]) Pick(keys []string, params ...any) (*ZodObject[T, R], error) {
	if z.hasRefinements() {
		return nil, ErrPickRefinements
	}

	newShape := make(core.ObjectSchema)
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

// Omit creates a new object excluding specified keys.
// Returns error if any key does not exist or schema contains refinements.
// Zod v4 compatible: throws on refinements or unrecognized keys.
func (z *ZodObject[T, R]) Omit(keys []string, params ...any) (*ZodObject[T, R], error) {
	if z.hasRefinements() {
		return nil, ErrOmitRefinements
	}

	// Validate all keys exist first
	for _, key := range keys {
		if _, exists := z.internals.Shape[key]; !exists {
			return nil, fmt.Errorf("%w: %q", ErrUnrecognizedKey, key)
		}
	}

	excludeSet := make(map[string]bool)
	for _, key := range keys {
		excludeSet[key] = true
	}

	newShape := make(core.ObjectSchema)
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

// hasRefinements checks if the schema has user-defined refinements applied.
// This distinguishes between internal checks (like alwaysPassCheck) and user-added refinements.
func (z *ZodObject[T, R]) hasRefinements() bool {
	return z.internals.HasUserRefinements
}

// Extend creates a new object by extending with additional fields.
// Returns error if the schema has refinements AND augmentation overlaps existing keys.
// Zod v4 compatible: .extend() throws when overwriting keys on refined schemas.
// See: .reference/zod/packages/zod/src/v4/core/util.ts:651-677
func (z *ZodObject[T, R]) Extend(augmentation core.ObjectSchema, params ...any) (*ZodObject[T, R], error) {
	// Check for refinement + overlapping key conflict (Zod v4 behavior)
	if z.internals.HasUserRefinements {
		for k := range augmentation {
			if _, exists := z.internals.Shape[k]; exists {
				return nil, ErrExtendRefinements
			}
		}
	}

	// Create new shape combining existing + extension fields
	newShape := make(core.ObjectSchema)

	// Copy existing shape
	for k, v := range z.internals.Shape {
		newShape[k] = v
	}

	// Add augmentation fields
	for k, schema := range augmentation {
		newShape[k] = schema
	}

	return ObjectTyped[T, R](newShape, params...), nil
}

// SafeExtend creates a new object by extending with additional fields without checking refinements.
// Use this when you explicitly want to overwrite fields on a refined schema.
// Zod v4 compatible: .safeExtend() bypasses the refinement check.
// See: .reference/zod/packages/zod/src/v4/core/util.ts:679-691
func (z *ZodObject[T, R]) SafeExtend(augmentation core.ObjectSchema, params ...any) *ZodObject[T, R] {
	// Create new shape combining existing + extension fields
	newShape := make(core.ObjectSchema)

	// Copy existing shape
	for k, v := range z.internals.Shape {
		newShape[k] = v
	}

	// Add augmentation fields
	for k, schema := range augmentation {
		newShape[k] = schema
	}

	return ObjectTyped[T, R](newShape, params...)
}

// Merge combines two object schemas.
// Zod v4 compatible: .merge() clears all checks/refinements.
// See: .reference/zod/packages/zod/src/v4/core/util.ts:693-707
func (z *ZodObject[T, R]) Merge(other *ZodObject[T, R], params ...any) *ZodObject[T, R] {
	// Create new shape combining existing + other fields
	newShape := make(core.ObjectSchema)

	// Copy existing shape
	for k, v := range z.internals.Shape {
		newShape[k] = v
	}

	// Add other schema's fields
	for k, schema := range other.internals.Shape {
		newShape[k] = schema
	}

	// Zod v4: merge() clears checks, so we create a fresh schema
	// HasUserRefinements will be false on the new schema
	return ObjectTyped[T, R](newShape, params...)
}

// Partial makes all fields optional
func (z *ZodObject[T, R]) Partial(keys ...[]string) *ZodObject[T, R] {
	newInternals := z.internals.Clone()

	var partialExceptions map[string]bool
	if len(keys) > 0 && len(keys[0]) > 0 {
		// Specific keys provided - these are the ones to make optional
		// All other fields remain required
		partialExceptions = make(map[string]bool)
		for fieldName := range z.internals.Shape {
			partialExceptions[fieldName] = true // Mark all as exceptions initially
		}
		// Remove the keys that should be made optional from exceptions
		for _, key := range keys[0] {
			delete(partialExceptions, key)
		}
	}

	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:   *newInternals,
		Def:                z.internals.Def,
		Shape:              z.internals.Shape,
		Catchall:           z.internals.Catchall,
		UnknownKeys:        z.internals.UnknownKeys,
		IsPartial:          true,
		PartialExceptions:  partialExceptions,
		HasUserRefinements: z.internals.HasUserRefinements,
	}}
}

// Required makes all fields required (opposite of Partial)
func (z *ZodObject[T, R]) Required(fields ...[]string) *ZodObject[T, R] {
	newInternals := z.internals.Clone()

	var partialExceptions map[string]bool
	if len(fields) > 0 && len(fields[0]) > 0 {
		// Specific fields provided - these become required
		partialExceptions = make(map[string]bool)
		for _, fieldName := range fields[0] {
			partialExceptions[fieldName] = true
		}
	}

	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:   *newInternals,
		Def:                z.internals.Def,
		Shape:              z.internals.Shape,
		Catchall:           z.internals.Catchall,
		UnknownKeys:        z.internals.UnknownKeys,
		IsPartial:          true,              // Keep as partial, but with specific required fields
		PartialExceptions:  partialExceptions, // Fields in this map are required
		HasUserRefinements: z.internals.HasUserRefinements,
	}}
}

// Strict sets strict mode (no unknown keys allowed)
func (z *ZodObject[T, R]) Strict() *ZodObject[T, R] {
	return z.withUnknownKeys(ObjectModeStrict)
}

// Strip sets strip mode (unknown keys are removed)
func (z *ZodObject[T, R]) Strip() *ZodObject[T, R] {
	return z.withUnknownKeys(ObjectModeStrip)
}

// Passthrough sets passthrough mode (unknown keys are allowed)
func (z *ZodObject[T, R]) Passthrough() *ZodObject[T, R] {
	return z.withUnknownKeys(ObjectModePassthrough)
}

// Catchall sets a schema for unknown keys
func (z *ZodObject[T, R]) Catchall(catchallSchema core.ZodSchema) *ZodObject[T, R] {
	newInternals := z.internals.Clone()
	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:   *newInternals,
		Def:                z.internals.Def,
		Shape:              z.internals.Shape,
		Catchall:           catchallSchema,
		UnknownKeys:        z.internals.UnknownKeys,
		IsPartial:          z.internals.IsPartial,
		PartialExceptions:  z.internals.PartialExceptions,
		HasUserRefinements: z.internals.HasUserRefinements,
	}}
}

// Keyof returns a string literal schema of all keys
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

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodObject[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractObjectValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodObject[T, R]) Overwrite(transform func(R) R, params ...any) *ZodObject[T, R] {
	// Create a transformation function that works with the constraint type R
	transformAny := func(input any) any {
		// Try to convert input to constraint type R
		converted, ok := convertToObjectType[T, R](input)
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

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodObject[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractObjectValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation using constraint type.
// Schemas with refinements cannot use Pick/Omit (Zod v4 compatible).
func (z *ZodObject[T, R]) Refine(fn func(R) bool, params ...any) *ZodObject[T, R] {
	wrapper := func(v any) bool {
		if v == nil {
			return true // Skip refinement for nil values
		}
		// Convert any to constraint type R for type-safe refinement
		if constraintVal, ok := convertToObjectType[T, R](v); ok {
			return fn(constraintVal)
		}
		return false
	}

	// Use unified parameter handling
	param := utils.GetFirstParam(params...)
	customParams := utils.NormalizeCustomParams(param)
	check := checks.NewCustom[any](wrapper, customParams)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	result := z.withInternals(newInternals)
	result.internals.HasUserRefinements = true // Mark as having user refinements
	return result
}

// RefineAny provides flexible validation without type conversion.
// Schemas with refinements cannot use Pick/Omit (Zod v4 compatible).
func (z *ZodObject[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodObject[T, R] {
	// Use unified parameter handling
	param := utils.GetFirstParam(params...)
	customParams := utils.NormalizeCustomParams(param)
	check := checks.NewCustom[any](fn, customParams)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	result := z.withInternals(newInternals)
	result.internals.HasUserRefinements = true // Mark as having user refinements
	return result
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates new instance with pointer constraint type
func (z *ZodObject[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodObject[T, *T] {
	return &ZodObject[T, *T]{internals: &ZodObjectInternals{
		ZodTypeInternals:   *in,
		Def:                z.internals.Def,
		Shape:              z.internals.Shape,
		Catchall:           z.internals.Catchall,
		UnknownKeys:        z.internals.UnknownKeys,
		IsPartial:          z.internals.IsPartial,
		PartialExceptions:  z.internals.PartialExceptions,
		HasUserRefinements: z.internals.HasUserRefinements,
	}}
}

// withInternals creates new instance preserving type
func (z *ZodObject[T, R]) withInternals(in *core.ZodTypeInternals) *ZodObject[T, R] {
	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:   *in,
		Def:                z.internals.Def,
		Shape:              z.internals.Shape,
		Catchall:           z.internals.Catchall,
		UnknownKeys:        z.internals.UnknownKeys,
		IsPartial:          z.internals.IsPartial,
		PartialExceptions:  z.internals.PartialExceptions,
		HasUserRefinements: z.internals.HasUserRefinements,
	}}
}

// withUnknownKeys creates new instance with specified unknown keys handling
func (z *ZodObject[T, R]) withUnknownKeys(mode ObjectMode) *ZodObject[T, R] {
	newInternals := z.internals.Clone()
	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:   *newInternals,
		Def:                z.internals.Def,
		Shape:              z.internals.Shape,
		Catchall:           z.internals.Catchall,
		UnknownKeys:        mode,
		IsPartial:          z.internals.IsPartial,
		PartialExceptions:  z.internals.PartialExceptions,
		HasUserRefinements: z.internals.HasUserRefinements,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodObject[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodObject[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// convertToObjectConstraintType converts base type T to constraint type R
func convertToObjectConstraintType[T any, R any](value T) R {
	// For value types, R should be T
	// For pointer constraints, R should be *T
	var result R
	resultVal := reflect.ValueOf(&result).Elem()

	// Check if R is a pointer to T
	if resultVal.Kind() == reflect.Pointer {
		// Create a new instance of T and set it
		if reflect.TypeOf(value).AssignableTo(resultVal.Type().Elem()) {
			newVal := reflect.New(resultVal.Type().Elem())
			newVal.Elem().Set(reflect.ValueOf(value))
			resultVal.Set(newVal)
		}
	} else {
		// Direct assignment for value types
		if reflect.TypeOf(value).AssignableTo(resultVal.Type()) {
			resultVal.Set(reflect.ValueOf(value))
		}
	}

	return result
}

// extractObjectValue extracts the base value from constraint type R
func extractObjectValue[T any, R any](value R) T {
	// Handle direct assignment (when T == R)
	if directValue, ok := any(value).(T); ok {
		return directValue
	}

	// Handle pointer dereferencing
	if ptrValue := reflect.ValueOf(value); ptrValue.Kind() == reflect.Ptr && !ptrValue.IsNil() {
		if derefValue, ok := ptrValue.Elem().Interface().(T); ok {
			return derefValue
		}
	}

	// Fallback to zero value
	var zero T
	return zero
}

// convertToObjectType converts any value to the object constraint type R with strict type checking
func convertToObjectType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeOf((*R)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Extract object value from input
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
		// Convert map[any]any to map[string]any
		objectValue = make(map[string]any)
		for k, v := range val {
			if strKey, ok := k.(string); ok {
				objectValue[strKey] = v
			} else {
				return zero, false // Non-string key found
			}
		}
		isValid = true
	default:
		return zero, false // Reject all non-map types
	}

	if !isValid {
		return zero, false
	}

	// Convert to target constraint type R
	zeroType := reflect.TypeOf((*R)(nil)).Elem()
	if zeroType.Kind() == reflect.Ptr {
		// R is *map[string]any
		if converted, ok := any(&objectValue).(R); ok {
			return converted, true
		}
	} else {
		// R is map[string]any
		if converted, ok := any(objectValue).(R); ok {
			return converted, true
		}
	}

	return zero, false
}

// extractObject extracts map[string]any from value with proper conversion
func (z *ZodObject[T, R]) extractObject(value any) (map[string]any, error) {
	// Handle direct map[string]any
	if objectVal, ok := value.(map[string]any); ok {
		return objectVal, nil
	}

	// Handle pointer to map[string]any
	if objectPtr, ok := value.(*map[string]any); ok && objectPtr != nil {
		return *objectPtr, nil
	}

	// Handle struct conversion using reflection
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("%w", ErrNilPointerCannotConvertToObject)
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		result := make(map[string]any)
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.IsExported() {
				// Use json tag if present, otherwise use field name
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

	ctx := core.NewParseContext()
	return nil, issues.CreateTypeConversionError(
		fmt.Sprintf("%T", value),
		"map[string]any",
		value,
		ctx,
	)
}

// validateObject validates object fields using field schemas with multiple error collection (TypeScript Zod v4 object behavior)
func (z *ZodObject[T, R]) validateObject(value map[string]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[string]any, error) {
	var collectedIssues []core.ZodRawIssue
	resultObject := make(map[string]any)

	// Validate known fields and collect all errors (TypeScript Zod v4 behavior)
	for fieldName, fieldSchema := range z.internals.Shape {
		fieldValue, exists := value[fieldName]

		if !exists {
			if !z.isFieldOptional(fieldSchema, fieldName) {
				// Create missing required field issue
				rawIssue := issues.CreateIssue(core.InvalidType, fmt.Sprintf("Missing required field: %s", fieldName), map[string]any{
					"expected": "nonoptional",
					"received": "undefined",
				}, nil)
				rawIssue.Path = []any{fieldName}
				collectedIssues = append(collectedIssues, rawIssue)
			}
			continue
		}

		// Check for explicit nil on ExactOptional fields (Zod v4: exactOptional rejects undefined/nil)
		if fieldValue == nil && z.isFieldExactOptional(fieldSchema) {
			rawIssue := issues.CreateIssue(core.InvalidType, fmt.Sprintf("Field %s cannot be explicitly nil (use absent key instead)", fieldName), map[string]any{
				"expected": "string", // or the actual expected type
				"received": "nil",
			}, nil)
			rawIssue.Path = []any{fieldName}
			collectedIssues = append(collectedIssues, rawIssue)
			continue
		}

		// Validate field directly to preserve original error codes
		if err := z.validateField(fieldValue, fieldSchema, ctx); err != nil {
			// Collect field validation errors with path prefix (TypeScript Zod v4 behavior)
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				// Propagate all issues from field validation with path prefix
				for _, fieldIssue := range zodErr.Issues {
					// Create raw issue preserving original code and essential properties
					rawIssue := core.ZodRawIssue{
						Code:       fieldIssue.Code,
						Message:    fieldIssue.Message,
						Input:      fieldIssue.Input,
						Path:       append([]any{fieldName}, fieldIssue.Path...), // Prepend field name to path
						Properties: make(map[string]any),
					}
					// Copy essential properties from ZodIssue to ZodRawIssue
					if fieldIssue.Minimum != nil {
						rawIssue.Properties["minimum"] = fieldIssue.Minimum
					}
					if fieldIssue.Maximum != nil {
						rawIssue.Properties["maximum"] = fieldIssue.Maximum
					}
					if fieldIssue.Expected != "" {
						rawIssue.Properties["expected"] = fieldIssue.Expected
					}
					if fieldIssue.Received != "" {
						rawIssue.Properties["received"] = fieldIssue.Received
					}
					rawIssue.Properties["inclusive"] = fieldIssue.Inclusive
					collectedIssues = append(collectedIssues, rawIssue)
				}
			} else {
				// Handle non-ZodError by creating a raw issue with field path
				rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, fieldValue)
				rawIssue.Path = []any{fieldName}
				collectedIssues = append(collectedIssues, rawIssue)
			}
		} else {
			// Field validation succeeded, add to result
			resultObject[fieldName] = fieldValue
		}
	}

	// Handle unknown fields and collect unrecognized_keys errors (TypeScript Zod v4 behavior)
	var unrecognizedKeys []string
	for fieldName, fieldValue := range value {
		if _, isKnown := z.internals.Shape[fieldName]; !isKnown {
			switch z.internals.UnknownKeys {
			case ObjectModeStrict:
				// Collect unrecognized keys for strict mode
				unrecognizedKeys = append(unrecognizedKeys, fieldName)
			case ObjectModeStrip:
				// Strip mode - don't add unknown fields to result
				continue
			case ObjectModePassthrough:
				if z.internals.Catchall != nil {
					// Validate unknown fields with catchall schema
					if err := z.validateField(fieldValue, z.internals.Catchall, ctx); err != nil {
						// Collect catchall validation errors
						var zodErr *issues.ZodError
						if errors.As(err, &zodErr) {
							for _, catchallIssue := range zodErr.Issues {
								rawIssue := core.ZodRawIssue{
									Code:       catchallIssue.Code,
									Message:    catchallIssue.Message,
									Input:      catchallIssue.Input,
									Path:       append([]any{fieldName}, catchallIssue.Path...),
									Properties: make(map[string]any),
								}
								// Copy properties
								if catchallIssue.Minimum != nil {
									rawIssue.Properties["minimum"] = catchallIssue.Minimum
								}
								if catchallIssue.Maximum != nil {
									rawIssue.Properties["maximum"] = catchallIssue.Maximum
								}
								if catchallIssue.Expected != "" {
									rawIssue.Properties["expected"] = catchallIssue.Expected
								}
								if catchallIssue.Received != "" {
									rawIssue.Properties["received"] = catchallIssue.Received
								}
								rawIssue.Properties["inclusive"] = catchallIssue.Inclusive
								collectedIssues = append(collectedIssues, rawIssue)
							}
						} else {
							rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, fieldValue)
							rawIssue.Path = []any{fieldName}
							collectedIssues = append(collectedIssues, rawIssue)
						}
					} else {
						// Catchall validation succeeded, add to result
						resultObject[fieldName] = fieldValue
					}
				} else {
					// No catchall - passthrough unknown fields
					resultObject[fieldName] = fieldValue
				}
			}
		}
	}

	// Add unrecognized_keys error for strict mode (TypeScript Zod v4 behavior)
	if len(unrecognizedKeys) > 0 {
		rawIssue := issues.CreateIssue(core.UnrecognizedKeys, "", map[string]any{
			"keys": unrecognizedKeys,
		}, value)
		collectedIssues = append(collectedIssues, rawIssue)
	}

	// Apply object-level checks and collect any issues
	if len(checks) > 0 {
		payload := core.NewParsePayload(resultObject)
		result := engine.RunChecksOnValue(resultObject, checks, payload, ctx)

		// Collect any object-level issues
		if result.HasIssues() {
			objectIssues := result.GetIssues()
			collectedIssues = append(collectedIssues, objectIssues...)
		}

		// Get the potentially transformed value
		if result.GetValue() != nil {
			if transformed, ok := result.GetValue().(map[string]any); ok {
				resultObject = transformed
			}
		}
	}

	// If we collected any issues (field-level, unknown keys, or object-level), return them as a combined error
	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}

	// Return the validated and potentially transformed object
	return resultObject, nil
}

// validateField validates a single field using reflection (without wrapping)
func (z *ZodObject[T, R]) validateField(element any, schema any, ctx *core.ParseContext) error {
	if schema == nil {
		return nil
	}

	// Try using reflection to call Parse method - this handles all schema types
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

	// Build arguments for Parse call
	var firstArg reflect.Value
	if element == nil {
		// If element is nil, use zero value of the method's first argument type
		firstArg = reflect.Zero(methodType.In(0))
	} else {
		firstArg = reflect.ValueOf(element)
	}
	args := []reflect.Value{firstArg}
	if methodType.NumIn() > 1 && methodType.In(1).String() == "*core.ParseContext" {
		// Add context parameter if expected
		if ctx == nil {
			args = append(args, reflect.Zero(methodType.In(1)))
		} else {
			args = append(args, reflect.ValueOf(ctx))
		}
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

// isFieldOptional checks if a field schema is optional using reflection or partial state
func (z *ZodObject[T, R]) isFieldOptional(schema any, fieldName string) bool {
	if schema == nil {
		return true
	}

	// Check if this object is in partial mode and this field should be optional in partial mode
	if z.internals.IsPartial {
		// If there are no exceptions, all fields are optional
		if z.internals.PartialExceptions == nil {
			return true
		}
		// If this field is not in the exceptions list, it's optional
		if !z.internals.PartialExceptions[fieldName] {
			return true
		}
	}

	schemaValue := reflect.ValueOf(schema)
	if !schemaValue.IsValid() || schemaValue.IsNil() {
		return true
	}

	// Try to get internals to check if optional
	if internalsMethod := schemaValue.MethodByName("GetInternals"); internalsMethod.IsValid() {
		results := internalsMethod.Call(nil)
		if len(results) > 0 {
			if internals, ok := results[0].Interface().(*core.ZodTypeInternals); ok {
				return internals.Optional
			}
		}
	}

	return false
}

// isFieldExactOptional checks if a field schema has ExactOptional set
// ExactOptional accepts absent keys but rejects explicit nil values
func (z *ZodObject[T, R]) isFieldExactOptional(schema any) bool {
	if schema == nil {
		return false
	}

	schemaValue := reflect.ValueOf(schema)
	if !schemaValue.IsValid() || schemaValue.IsNil() {
		return false
	}

	// Try to get internals to check if ExactOptional
	if internalsMethod := schemaValue.MethodByName("GetInternals"); internalsMethod.IsValid() {
		results := internalsMethod.Call(nil)
		if len(results) > 0 {
			if internals, ok := results[0].Interface().(*core.ZodTypeInternals); ok {
				return internals.ExactOptional
			}
		}
	}

	return false
}

// newZodObjectFromDef constructs new ZodObject from definition
func newZodObjectFromDef[T any, R any](def *ZodObjectDef) *ZodObject[T, R] {
	internals := &ZodObjectInternals{
		ZodTypeInternals:  engine.NewBaseZodTypeInternals(def.Type),
		Def:               def,
		Shape:             def.Shape,
		Catchall:          def.Catchall,
		UnknownKeys:       def.UnknownKeys,
		IsPartial:         false,
		PartialExceptions: nil,
	}

	// Provide constructor for AddCheck functionality
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

// ObjectTyped creates a typed object schema with explicit type parameters
func ObjectTyped[T any, R any](shape core.ObjectSchema, params ...any) *ZodObject[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodObjectDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeObject,
			Checks: []core.ZodCheck{},
		},
		Shape:       make(core.ObjectSchema),
		Catchall:    nil,
		UnknownKeys: ObjectModeStrip, // Default mode
	}

	// Copy the shape directly since core.ObjectSchema contains ZodSchema types
	for key, schema := range shape {
		def.Shape[key] = schema
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	objectSchema := newZodObjectFromDef[T, R](def)

	// Add a minimal check to trigger field validation
	alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
	objectSchema.internals.AddCheck(alwaysPassCheck)

	return objectSchema
}

// Object creates object schema with default types (map[string]any, map[string]any)
func Object(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return ObjectTyped[map[string]any, map[string]any](shape, params...)
}

// ObjectPtr creates object schema with pointer constraint
func ObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...)
}

// StrictObject creates strict object schema with default types
func StrictObject(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return Object(shape, params...).Strict()
}

// LooseObject creates loose object schema with default types
func LooseObject(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return Object(shape, params...).Passthrough()
}

// StrictObjectPtr creates strict object schema with pointer constraint
func StrictObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...).Strict()
}

// LooseObjectPtr creates loose object schema with pointer constraint
func LooseObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...).Passthrough()
}

// Check adds a custom validation function that can report multiple issues for object schema.
func (z *ZodObject[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodObject[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// Direct assertion attempt.
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Pointer/value mismatch handling.
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
func (z *ZodObject[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodObject[T, R] {
	return z.Check(fn, params...)
}

// extractObjectForEngine extracts map[string]any from input for engine.ParseComplex
func (z *ZodObject[T, R]) extractObjectForEngine(input any) (map[string]any, bool) {
	result, err := z.extractObject(input)
	if err != nil {
		return nil, false
	}
	return result, true
}

// extractObjectPtrForEngine extracts pointer to map[string]any from input for engine.ParseComplex
func (z *ZodObject[T, R]) extractObjectPtrForEngine(input any) (*map[string]any, bool) {
	// Try direct pointer extraction
	if ptr, ok := input.(*map[string]any); ok {
		return ptr, true
	}

	// Try extracting object and return pointer to it
	result, err := z.extractObject(input)
	if err != nil {
		return nil, false
	}
	return &result, true
}

// validateObjectForEngine validates map[string]any for engine.ParseComplex
func (z *ZodObject[T, R]) validateObjectForEngine(value map[string]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[string]any, error) {
	return z.validateObject(value, checks, ctx)
}
