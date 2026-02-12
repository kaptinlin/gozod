package types

import (
	"sync"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// LazyConstraint restricts values to any or *any for lazy schema types.
type LazyConstraint interface {
	any | *any
}

// ZodSchemaType represents any Zod schema type that implements the basic interface.
type ZodSchemaType interface {
	Internals() *core.ZodTypeInternals
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodLazyDef defines the configuration for lazy validation.
type ZodLazyDef struct {
	core.ZodTypeDef
	Getter func() any
}

// ZodLazyInternals contains lazy validator internal state.
type ZodLazyInternals struct {
	core.ZodTypeInternals
	Def       *ZodLazyDef
	Getter    func() any
	innerType core.ZodType[any]
	once      sync.Once
}

// ZodLazy represents a lazy validation schema for recursive type definitions.
type ZodLazy[T LazyConstraint] struct {
	internals *ZodLazyInternals
}

// ZodLazyTyped is a type-safe wrapper that preserves the inner schema type.
type ZodLazyTyped[S ZodSchemaType] struct {
	*ZodLazy[any]
	getter func() S
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodLazy[T]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodLazy[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodLazy[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodLazy[T]) withCheck(check core.ZodCheck) *ZodLazy[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// newLazyTypeError creates a ZodError for invalid type inputs.
func newLazyTypeError(value any, ctx *core.ParseContext) error {
	rawIssue := issues.NewRawIssue(core.InvalidType, value, issues.WithExpected(string(core.ZodTypeLazy)))
	return issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(rawIssue, ctx, nil)})
}

// Coerce delegates coercion to the inner schema.
func (z *ZodLazy[T]) Coerce(input any) (any, bool) {
	innerSchema := z.getInnerType()
	if innerSchema == nil {
		return input, false
	}
	if coercible, ok := innerSchema.(interface{ Coerce(any) (any, bool) }); ok {
		if result, success := coercible.Coerce(input); success {
			return result, true
		}
	}
	return input, false
}

// convertLazyResult converts the engine result to the constraint type T.
func (z *ZodLazy[T]) convertLazyResult(result any) T {
	if result == nil {
		var zero T
		return zero
	}
	var zero T
	switch any(zero).(type) {
	case *any:
		resultPtr := &result
		return any(resultPtr).(T)
	default:
		return any(result).(T) //nolint:unconvert
	}
}

// Parse validates and returns the parsed value using lazy evaluation.
func (z *ZodLazy[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	internals := &z.internals.ZodTypeInternals

	// Handle nil input with modifier precedence.
	if input == nil {
		if internals.NonOptional {
			var zero T
			return zero, issues.CreateNonOptionalError(parseCtx)
		}
		if internals.DefaultValue != nil {
			return any(internals.DefaultValue).(T), nil //nolint:unconvert
		}
		if internals.DefaultFunc != nil {
			return any(internals.DefaultFunc()).(T), nil //nolint:unconvert
		}
		switch {
		case internals.PrefaultValue != nil:
			input = internals.PrefaultValue
		case internals.PrefaultFunc != nil:
			input = internals.PrefaultFunc()
		case internals.Optional || internals.Nilable:
			var zero T
			return zero, nil
		default:
			var zero T
			return zero, issues.CreateInvalidTypeError(core.ZodTypeLazy, nil, parseCtx)
		}
	}

	result, err := z.validateLazy(input, internals.Checks, parseCtx)
	if err != nil {
		var zero T
		return zero, err
	}
	return z.convertLazyResult(result), nil
}

// MustParse validates input and panics on error.
func (z *ZodLazy[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (z *ZodLazy[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse provides compile-time type safety by requiring exact type matching.
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
	return z.convertLazyResult(result), nil
}

// MustStrictParse validates input with strict type matching and panics on error.
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

// Optional returns a schema that accepts nil, with constraint type *any.
func (z *ZodLazy[T]) Optional() *ZodLazy[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil, with constraint type *any.
func (z *ZodLazy[T]) Nilable() *ZodLazy[*any] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility.
func (z *ZodLazy[T]) Nullish() *ZodLazy[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil value.
func (z *ZodLazy[T]) NonOptional() *ZodLazy[any] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodLazy[any]{internals: &ZodLazyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Getter:           z.internals.Getter,
	}}
}

// Default sets a default value returned when input is nil, bypassing validation.
func (z *ZodLazy[T]) Default(v any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory function that provides the default value when input is nil.
func (z *ZodLazy[T]) DefaultFunc(fn func() any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(fn)
	return z.withInternals(in)
}

// Prefault sets a prefault value that goes through the full parsing pipeline when input is nil.
func (z *ZodLazy[T]) Prefault(v any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a factory function that provides the prefault value through the full parsing pipeline.
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

// Transform creates a type-safe transformation pipeline.
func (z *ZodLazy[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(any(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Pipe creates a validation pipeline to another schema.
func (z *ZodLazy[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(any(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a typed custom validation function.
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
		errorMessage = schemaParams.Error
	}
	return z.withCheck(checks.NewCustom[any](wrapper, errorMessage))
}

// RefineAny adds a flexible custom validation function accepting any input.
func (z *ZodLazy[T]) RefineAny(fn func(any) bool, params ...any) *ZodLazy[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}
	return z.withCheck(checks.NewCustom[any](fn, errorMessage))
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Unwrap resolves and returns the inner schema.
func (z *ZodLazy[T]) Unwrap() core.ZodType[any] {
	return z.getInnerType()
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// getInnerType implements lazy evaluation with thread-safe caching using sync.Once.
func (z *ZodLazy[T]) getInnerType() core.ZodType[any] {
	z.internals.once.Do(func() {
		rawSchema := z.internals.Getter()
		z.internals.innerType = convertToAnyInterface(rawSchema)
	})
	return z.internals.innerType
}

// convertToAnyInterface converts any schema type to ZodType[any].
func convertToAnyInterface(schema any) core.ZodType[any] {
	if schema == nil {
		return nil
	}
	if anySchema, ok := schema.(core.ZodType[any]); ok {
		return anySchema
	}
	return &schemaWrapper{inner: schema}
}

// schemaWrapper implements ZodType[any] by wrapping any schema type.
type schemaWrapper struct {
	inner any
}

func (w *schemaWrapper) Parse(input any, ctx ...*core.ParseContext) (any, error) {
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
		Parse(any, ...*core.ParseContext) (float64, error)
	}:
		return s.Parse(input, ctx...)
	case interface {
		Parse(any, ...*core.ParseContext) (int64, error)
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
		return nil, newLazyTypeError(input, nil)
	}
}

func (w *schemaWrapper) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := w.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

func (w *schemaWrapper) Internals() *core.ZodTypeInternals {
	if s, ok := w.inner.(interface{ Internals() *core.ZodTypeInternals }); ok {
		return s.Internals()
	}
	return &core.ZodTypeInternals{}
}

func (w *schemaWrapper) Coerce(input any) (any, bool) {
	if s, ok := w.inner.(interface{ Coerce(any) (any, bool) }); ok {
		return s.Coerce(input)
	}
	return input, false
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (w *schemaWrapper) IsOptional() bool {
	if s, ok := w.inner.(interface{ IsOptional() bool }); ok {
		return s.IsOptional()
	}
	return false
}

// IsNilable reports whether this schema accepts nil values.
func (w *schemaWrapper) IsNilable() bool {
	if s, ok := w.inner.(interface{ IsNilable() bool }); ok {
		return s.IsNilable()
	}
	return false
}

// Inner returns the wrapped inner schema for JSON Schema conversion.
func (w *schemaWrapper) Inner() any {
	return w.inner
}

// extractLazy extracts a value from input by accepting any value.
func (z *ZodLazy[T]) extractLazy(value any) (any, bool) {
	return value, true
}

// extractLazyPtr extracts a pointer value from input.
func (z *ZodLazy[T]) extractLazyPtr(value any) (*any, bool) {
	var zero T
	if _, isPtr := any(zero).(*any); !isPtr {
		return nil, false
	}
	if value == nil {
		return nil, true
	}
	if ptr, ok := value.(*any); ok {
		return ptr, true
	}
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

// validateLazy validates that the lazy schema can be resolved and the value is valid.
func (z *ZodLazy[T]) validateLazy(value any, chks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	if value == nil {
		internals := z.Internals()
		if internals.Optional || internals.Nilable {
			return value, nil
		}
		return nil, newLazyTypeError(value, ctx)
	}
	innerSchema := z.getInnerType()
	if innerSchema == nil {
		return nil, newLazyTypeError(value, ctx)
	}
	result, err := innerSchema.Parse(value, ctx)
	if err != nil {
		if isExpectedLazyError(err) {
			return engine.ApplyChecks[any](value, chks, ctx)
		}
		return nil, err
	}
	return engine.ApplyChecks[any](result, chks, ctx)
}

// withPtrInternals creates a new ZodLazy instance with pointer type.
func (z *ZodLazy[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodLazy[*any] {
	return &ZodLazy[*any]{internals: z.cloneInternals(in)}
}

// withInternals creates a new ZodLazy instance preserving generic type T.
func (z *ZodLazy[T]) withInternals(in *core.ZodTypeInternals) *ZodLazy[T] {
	return &ZodLazy[T]{internals: z.cloneInternals(in)}
}

// cloneInternals creates a new ZodLazyInternals with the given type internals,
// preserving the cached inner type from the source schema.
func (z *ZodLazy[T]) cloneInternals(in *core.ZodTypeInternals) *ZodLazyInternals {
	cloned := &ZodLazyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Getter:           z.internals.Getter,
		innerType:        z.internals.innerType,
	}
	// Pre-exhaust sync.Once if inner type is already resolved.
	if z.internals.innerType != nil {
		cloned.once.Do(func() {})
	}
	return cloned
}

// CloneFrom copies configuration from another schema.
func (z *ZodLazy[T]) CloneFrom(source any) {
	src, ok := source.(*ZodLazy[T])
	if !ok {
		return
	}
	originalChecks := z.internals.Checks
	z.internals.ZodTypeInternals = src.internals.ZodTypeInternals
	z.internals.Def = src.internals.Def
	z.internals.Getter = src.internals.Getter
	z.internals.innerType = src.internals.innerType
	// Note: once is intentionally not copied — new instance gets fresh sync.Once.
	// Pre-exhaust if source had resolved inner type.
	if src.internals.innerType != nil {
		z.internals.once.Do(func() {})
	}
	z.internals.Checks = originalChecks
}

// newZodLazyFromDef constructs a new ZodLazy from the given definition.
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
	}
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
func Lazy[S ZodSchemaType](getter func() S, params ...any) *ZodLazyTyped[S] {
	schemaParams := utils.NormalizeParams(params...)
	def := &ZodLazyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeLazy,
			Checks: []core.ZodCheck{},
		},
		Getter: func() any { return getter() },
	}
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}
	return newZodLazyTypedFromDef[S](def, getter)
}

// LazyAny creates a lazy schema that defers evaluation until needed.
func LazyAny(getter func() any, params ...any) *ZodLazy[any] {
	return LazyTyped[any](getter, params...)
}

// LazyPtr creates a lazy schema for *any type.
func LazyPtr(getter func() any, params ...any) *ZodLazy[*any] {
	return LazyTyped[*any](getter, params...)
}

// LazyTyped is the underlying generic function for creating lazy schemas.
func LazyTyped[T LazyConstraint](getter func() any, params ...any) *ZodLazy[T] {
	schemaParams := utils.NormalizeParams(params...)
	def := &ZodLazyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeLazy,
			Checks: []core.ZodCheck{},
		},
		Getter: getter,
	}
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}
	return newZodLazyFromDef[T](def)
}

// =============================================================================
// TYPE-SAFE LAZY SCHEMA IMPLEMENTATION
// =============================================================================

// Parse validates and returns the parsed value with type safety.
func (z *ZodLazyTyped[S]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return z.ZodLazy.Parse(input, ctx...)
}

// MustParse validates input and panics on error.
func (z *ZodLazyTyped[S]) MustParse(input any, ctx ...*core.ParseContext) any {
	return z.ZodLazy.MustParse(input, ctx...)
}

// InnerSchema returns the inner schema with its original type.
func (z *ZodLazyTyped[S]) GetInnerSchema() S {
	return z.getter()
}

// Optional returns a schema that accepts nil.
func (z *ZodLazyTyped[S]) Optional() *ZodLazy[*any] {
	return z.ZodLazy.Optional()
}

// Nilable returns a schema that accepts nil.
func (z *ZodLazyTyped[S]) Nilable() *ZodLazy[*any] {
	return z.ZodLazy.Nilable()
}

// NonOptional removes Optional flag and enforces non-nil value.
func (z *ZodLazyTyped[S]) NonOptional() *ZodLazy[any] {
	return z.ZodLazy.NonOptional()
}

// Default sets a default value returned when input is nil, bypassing validation.
func (z *ZodLazyTyped[S]) Default(v any) *ZodLazyTyped[S] {
	newLazy := z.ZodLazy.Default(v)
	return &ZodLazyTyped[S]{ZodLazy: newLazy, getter: z.getter}
}

// Refine adds a custom validation function.
func (z *ZodLazyTyped[S]) Refine(fn func(any) bool, params ...any) *ZodLazyTyped[S] {
	newLazy := z.RefineAny(fn, params...)
	return &ZodLazyTyped[S]{ZodLazy: newLazy, getter: z.getter}
}

// newZodLazyTypedFromDef creates a new type-safe lazy schema.
func newZodLazyTypedFromDef[S ZodSchemaType](def *ZodLazyDef, getter func() S) *ZodLazyTyped[S] {
	return &ZodLazyTyped[S]{
		ZodLazy: newZodLazyFromDef[any](def),
		getter:  getter,
	}
}

// isExpectedLazyError reports whether the error represents an `invalid_type`
// issue whose expected type is `lazy`. This detects recursive lazy-schema
// validation without relying on fragile string matching.
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
