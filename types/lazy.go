package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// LazyConstraint restricts values to any or *any for lazy schema types
type LazyConstraint interface {
	any | *any
}

// ZodSchemaType represents any Zod schema type that implements the basic interface
type ZodSchemaType interface {
	GetInternals() *core.ZodTypeInternals
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodLazyDef defines the configuration for lazy validation
type ZodLazyDef struct {
	core.ZodTypeDef
	Getter func() any // Schema getter function that returns any schema type
}

// ZodLazyInternals contains lazy validator internal state
type ZodLazyInternals struct {
	core.ZodTypeInternals
	Def       *ZodLazyDef       // Schema definition
	Getter    func() any        // Schema getter function that returns any schema type
	InnerType core.ZodType[any] // Cached inner schema (converted to interface)
	Cached    bool              // Whether inner schema is cached
}

// ZodLazy represents a lazy validation schema for recursive type definitions
type ZodLazy[T LazyConstraint] struct {
	internals *ZodLazyInternals
}

// ZodLazyTyped is a type-safe wrapper that preserves the inner schema type
type ZodLazyTyped[S ZodSchemaType] struct {
	*ZodLazy[any]
	getter func() S
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodLazy[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodLazy[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodLazy[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements Coercible interface - lazy schemas delegate to inner schema
func (z *ZodLazy[T]) Coerce(input any) (any, bool) {
	innerSchema := z.getInnerType()
	if innerSchema == nil {
		return input, false
	}

	// Check if inner schema supports coercion using type assertion
	type Coercible interface {
		Coerce(any) (any, bool)
	}

	if coercible, ok := innerSchema.(Coercible); ok {
		result, success := coercible.Coerce(input)
		if success {
			return result, true
		}
		// If coercion failed, return original input
		return input, false
	}

	return input, false
}

// Parse validates and returns the parsed value using lazy evaluation
func (z *ZodLazy[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Handle modifiers first
	internals := &z.internals.ZodTypeInternals

	// Handle nil input explicitly
	if input == nil {
		// Check for modifiers first
		if internals.NonOptional {
			var zero T
			return zero, issues.CreateNonOptionalError(parseCtx)
		}

		// Default/DefaultFunc - short circuit
		if internals.DefaultValue != nil {
			return any(internals.DefaultValue).(T), nil //nolint:unconvert // Required for generic type constraint conversion
		}
		if internals.DefaultFunc != nil {
			defaultValue := internals.DefaultFunc()
			return any(defaultValue).(T), nil //nolint:unconvert // Required for generic type constraint conversion
		}

		// Prefault/PrefaultFunc - use as new input
		switch {
		case internals.PrefaultValue != nil:
			input = internals.PrefaultValue
		case internals.PrefaultFunc != nil:
			input = internals.PrefaultFunc()
		case internals.Optional || internals.Nilable:
			// Optional/Nilable - allow nil
			var zero T
			return zero, nil
		default:
			// Reject nil input
			var zero T
			return zero, issues.CreateInvalidTypeError(core.ZodTypeLazy, nil, parseCtx)
		}
	}

	// Core lazy evaluation logic - validate with inner schema
	result, err := z.validateLazy(input, internals.Checks, parseCtx)
	if err != nil {
		var zero T
		return zero, err
	}

	if result == nil {
		var zero T
		return zero, nil
	}

	// Handle type conversion for pointer types
	var zero T
	switch any(zero).(type) {
	case *any:
		// If T is *any, we need to convert result to *any
		if result == nil {
			return any((*any)(nil)).(T), nil
		}
		// Create a pointer to the result
		resultPtr := &result
		return any(resultPtr).(T), nil
	default:
		// For any type, direct conversion
		return result.(T), nil
	}
}

// MustParse is the type-safe variant that panics on error
func (z *ZodLazy[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodLazy[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse validates the input using strict parsing rules
func (z *ZodLazy[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	result, err := engine.ParseComplexStrict[any](
		any(input),
		&z.internals.ZodTypeInternals,
		core.ZodTypeLazy,
		z.extractLazy,
		z.extractLazyPtr,
		z.validateLazy,
		ctx...,
	)
	if err != nil {
		var zero T
		return zero, err
	}

	// Type assert the result to T
	if result == nil {
		var zero T
		return zero, nil
	}

	if typedResult, ok := result.(T); ok {
		return typedResult, nil
	}

	// For any type, return result directly
	return any(result).(T), nil //nolint:unconvert
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodLazy[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional allows the lazy schema to be nil, returns pointer type
func (z *ZodLazy[T]) Optional() *ZodLazy[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows the lazy schema to be nil, returns pointer type
func (z *ZodLazy[T]) Nilable() *ZodLazy[*any] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers, returns pointer type
func (z *ZodLazy[T]) Nullish() *ZodLazy[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil value (any).
func (z *ZodLazy[T]) NonOptional() *ZodLazy[any] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodLazy[any]{
		internals: &ZodLazyInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Getter:           z.internals.Getter,
		},
	}
}

// Default preserves the current generic type T
func (z *ZodLazy[T]) Default(v any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves the current generic type T
func (z *ZodLazy[T]) DefaultFunc(fn func() any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(fn)
	return z.withInternals(in)
}

// Prefault preserves the current generic type T
func (z *ZodLazy[T]) Prefault(v any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc preserves the current generic type T
func (z *ZodLazy[T]) PrefaultFunc(fn func() any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(fn)
	return z.withInternals(in)
}

// Meta stores metadata for this lazy schema.
func (z *ZodLazy[T]) Meta(meta core.GlobalMeta) *ZodLazy[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodLazy[T]) Describe(description string) *ZodLazy[T] {
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
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodLazy[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(any(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodLazy[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(any(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// TYPE CONVERSION - NO LONGER NEEDED (USING WRAPFN PATTERN)
// =============================================================================

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies custom validation function
func (z *ZodLazy[T]) Refine(fn func(T) bool, params ...any) *ZodLazy[T] {
	wrapper := func(v any) bool {
		if typedVal, ok := v.(T); ok {
			return fn(typedVal)
		}
		return false
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodLazy[T]) RefineAny(fn func(any) bool, params ...any) *ZodLazy[T] {
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
// TYPE-SPECIFIC METHODS
// =============================================================================

// Unwrap resolves and returns the inner schema
func (z *ZodLazy[T]) Unwrap() core.ZodType[any] {
	return z.getInnerType()
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// getInnerType implements lazy evaluation with caching
func (z *ZodLazy[T]) getInnerType() core.ZodType[any] {
	if !z.internals.Cached {
		rawSchema := z.internals.Getter()
		z.internals.InnerType = convertToAnyInterface(rawSchema)
		z.internals.Cached = true
	}
	return z.internals.InnerType
}

// convertToAnyInterface converts any schema type to ZodType[any]
func convertToAnyInterface(schema any) core.ZodType[any] {
	if schema == nil {
		return nil
	}

	// Check if it already implements ZodType[any]
	if anySchema, ok := schema.(core.ZodType[any]); ok {
		return anySchema
	}

	// Create wrapper for different schema types
	return &schemaWrapper{inner: schema}
}

// schemaWrapper implements ZodType[any] by wrapping any schema type
type schemaWrapper struct {
	inner any
}

func (w *schemaWrapper) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// Use type assertion to call the correct Parse method
	switch s := w.inner.(type) {
	case interface {
		Parse(any, ...*core.ParseContext) (any, error)
	}:
		return s.Parse(input, ctx...)
	case interface {
		Parse(any, ...*core.ParseContext) (string, error)
	}:
		return s.Parse(input, ctx...)
	case interface {
		Parse(any, ...*core.ParseContext) (bool, error)
	}:
		return s.Parse(input, ctx...)
	case interface {
		Parse(any, ...*core.ParseContext) (int, error)
	}:
		return s.Parse(input, ctx...)
	case interface {
		Parse(any, ...*core.ParseContext) (*string, error)
	}:
		return s.Parse(input, ctx...)
	case interface {
		Parse(any, ...*core.ParseContext) (*bool, error)
	}:
		return s.Parse(input, ctx...)
	default:
		rawIssue := issues.NewRawIssue(core.InvalidType, input, issues.WithExpected(string(core.ZodTypeLazy)))
		return nil, issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(rawIssue, nil, nil)})
	}
}

func (w *schemaWrapper) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := w.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

func (w *schemaWrapper) GetInternals() *core.ZodTypeInternals {
	// Use type assertion to call the correct GetInternals method
	if s, ok := w.inner.(interface{ GetInternals() *core.ZodTypeInternals }); ok {
		return s.GetInternals()
	}
	return &core.ZodTypeInternals{}
}

func (w *schemaWrapper) Coerce(input any) (any, bool) {
	// Use type assertion to call the correct Coerce method
	if s, ok := w.inner.(interface{ Coerce(any) (any, bool) }); ok {
		return s.Coerce(input)
	}
	return input, false
}

// IsOptional returns true if this schema accepts undefined/missing values
func (w *schemaWrapper) IsOptional() bool {
	if s, ok := w.inner.(interface{ IsOptional() bool }); ok {
		return s.IsOptional()
	}
	return false
}

// IsNilable returns true if this schema accepts nil values
func (w *schemaWrapper) IsNilable() bool {
	if s, ok := w.inner.(interface{ IsNilable() bool }); ok {
		return s.IsNilable()
	}
	return false
}

// GetInner returns the wrapped inner schema for JSON Schema conversion
func (w *schemaWrapper) GetInner() any {
	return w.inner
}

// extractLazy extracts a value from input by delegating to inner schema
func (z *ZodLazy[T]) extractLazy(value any) (any, bool) {
	// For Lazy types, we accept any value and let validation handle the logic
	// This is different from other extractors because Lazy delegates to inner schema
	return value, true
}

// extractLazyPtr extracts a pointer value from input by delegating to inner schema
func (z *ZodLazy[T]) extractLazyPtr(value any) (*any, bool) {
	// Check if T is actually a pointer type first
	var zero T
	if _, isPtr := any(zero).(*any); !isPtr {
		// T is not a pointer type, so this extractor should not handle it
		return nil, false
	}

	if value == nil {
		return nil, true
	}

	// If it's already a pointer
	if ptr, ok := value.(*any); ok {
		return ptr, true
	}

	// Otherwise, delegate to inner schema and create pointer
	innerSchema := z.getInnerType()
	if innerSchema == nil {
		return nil, false
	}

	result, err := innerSchema.Parse(value)
	if err != nil {
		return nil, false
	}

	return &result, true
}

// validateLazy validates that the lazy schema can be resolved and the value is valid
func (z *ZodLazy[T]) validateLazy(value any, checks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	if value == nil {
		internals := z.GetInternals()
		if internals.Optional || internals.Nilable {
			return value, nil
		}
		rawIssue := issues.NewRawIssue(core.InvalidType, value, issues.WithExpected(string(core.ZodTypeLazy)))
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Get inner schema
	innerSchema := z.getInnerType()
	if innerSchema == nil {
		rawIssue := issues.NewRawIssue(core.InvalidType, value, issues.WithExpected(string(core.ZodTypeLazy)))
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Delegate validation to inner schema - this is the core lazy validation logic
	// that must always run regardless of whether checks are present
	result, err := innerSchema.Parse(value, ctx)
	if err != nil {
		// Recursive reference case – treat current value as already validated
		if isExpectedLazyError(err) {
			return engine.ApplyChecks[any](value, checks, ctx)
		} else {
			return nil, err
		}
	}

	// Run standard checks validation on the result
	return engine.ApplyChecks[any](result, checks, ctx)
}

// withPtrInternals creates a new ZodLazy instance with pointer type
func (z *ZodLazy[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodLazy[*any] {
	return &ZodLazy[*any]{internals: &ZodLazyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Getter:           z.internals.Getter,
		InnerType:        z.internals.InnerType,
		Cached:           z.internals.Cached,
	}}
}

// withInternals creates a new ZodLazy instance preserving generic type T
func (z *ZodLazy[T]) withInternals(in *core.ZodTypeInternals) *ZodLazy[T] {
	return &ZodLazy[T]{internals: &ZodLazyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Getter:           z.internals.Getter,
		InnerType:        z.internals.InnerType,
		Cached:           z.internals.Cached,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodLazy[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodLazy[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// newZodLazyFromDef constructs a new ZodLazy from the given definition
func newZodLazyFromDef[T LazyConstraint](def *ZodLazyDef) *ZodLazy[T] {
	internals := &ZodLazyInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:    def,
		Getter: def.Getter,
		Cached: false,
	}

	// Provide a constructor so that AddCheck can create new schema instances
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		lazyDef := &ZodLazyDef{
			ZodTypeDef: *newDef,
			Getter:     def.Getter,
		}
		return any(newZodLazyFromDef[T](lazyDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodLazy[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Lazy creates a type-safe lazy schema with compile-time type checking.
// This allows for syntax like: Lazy[*ZodString[string]](func() *ZodString[string] { return String().Min(3) })
func Lazy[S ZodSchemaType](getter func() S, params ...any) *ZodLazyTyped[S] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodLazyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeLazy,
			Checks: []core.ZodCheck{},
		},
		Getter: func() any { return getter() },
	}

	// Apply the normalized parameters to the schema definition
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodLazyTypedFromDef[S](def, getter)
}

// LazyAny creates a lazy schema that defers evaluation until needed, enabling recursive type definitions
func LazyAny(getter func() any, params ...any) *ZodLazy[any] {
	return LazyTyped[any](getter, params...)
}

// LazyPtr creates a schema for *any lazy type
func LazyPtr(getter func() any, params ...any) *ZodLazy[*any] {
	return LazyTyped[*any](getter, params...)
}

// LazyTyped is the underlying generic function for creating lazy schemas
func LazyTyped[T LazyConstraint](getter func() any, params ...any) *ZodLazy[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodLazyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeLazy,
			Checks: []core.ZodCheck{},
		},
		Getter: getter,
	}

	// Apply the normalized parameters to the schema definition
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodLazyFromDef[T](def)
}

// =============================================================================
// TYPE-SAFE LAZY SCHEMA IMPLEMENTATION
// =============================================================================

// Parse validates and returns the parsed value with type safety
func (z *ZodLazyTyped[S]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return z.ZodLazy.Parse(input, ctx...)
}

// MustParse is the type-safe variant that panics on error
func (z *ZodLazyTyped[S]) MustParse(input any, ctx ...*core.ParseContext) any {
	return z.ZodLazy.MustParse(input, ctx...)
}

// GetInnerSchema returns the inner schema with its original type
func (z *ZodLazyTyped[S]) GetInnerSchema() S {
	return z.getter()
}

// Optional creates an optional version
func (z *ZodLazyTyped[S]) Optional() *ZodLazy[*any] {
	return z.ZodLazy.Optional()
}

// Nilable creates a nilable version
func (z *ZodLazyTyped[S]) Nilable() *ZodLazy[*any] {
	return z.ZodLazy.Nilable()
}

// NonOptional removes Optional flag and enforces non-nil value.
func (z *ZodLazyTyped[S]) NonOptional() *ZodLazy[any] {
	return z.ZodLazy.NonOptional()
}

// Default provides a default value for the lazy schema
func (z *ZodLazyTyped[S]) Default(v any) *ZodLazyTyped[S] {
	newLazy := z.ZodLazy.Default(v)
	return &ZodLazyTyped[S]{ZodLazy: newLazy, getter: z.getter}
}

// Refine applies custom validation with type safety
func (z *ZodLazyTyped[S]) Refine(fn func(any) bool, params ...any) *ZodLazyTyped[S] {
	newLazy := z.RefineAny(fn, params...)
	return &ZodLazyTyped[S]{ZodLazy: newLazy, getter: z.getter}
}

// newZodLazyTypedFromDef creates a new type-safe lazy schema
func newZodLazyTypedFromDef[S ZodSchemaType](def *ZodLazyDef, getter func() S) *ZodLazyTyped[S] {
	lazySchema := newZodLazyFromDef[any](def)
	return &ZodLazyTyped[S]{
		ZodLazy: lazySchema,
		getter:  getter,
	}
}

// isExpectedLazyError returns true if the provided error represents an
// `invalid_type` issue whose expected type is `lazy`. This is used to detect
// recursive lazy-schema validation without relying on fragile string matching.
func isExpectedLazyError(err error) bool {
	var zErr *issues.ZodError
	if !issues.IsZodError(err, &zErr) {
		return false
	}
	for _, iss := range zErr.Issues {
		if iss.Code == core.InvalidType && iss.Expected == core.ZodTypeLazy {
			return true
		}
	}
	return false
}
