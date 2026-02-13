package types

import (
	"sync"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// LazyConstraint restricts values to any or *any for lazy schema types.
type LazyConstraint interface {
	any | *any
}

// ZodSchemaType represents any Zod schema type that implements the basic interface.
type ZodSchemaType interface {
	Internals() *core.ZodTypeInternals
}

// =============================================================================
// Type Definitions
// =============================================================================

// ZodLazyDef is the configuration for lazy validation.
type ZodLazyDef struct {
	core.ZodTypeDef
	Getter func() any
}

// ZodLazyInternals holds the internal state of a lazy validator.
type ZodLazyInternals struct {
	core.ZodTypeInternals
	Def       *ZodLazyDef
	Getter    func() any
	innerType core.ZodType[any]
	once      sync.Once
}

// ZodLazy is a lazy validation schema for recursive type definitions.
type ZodLazy[T LazyConstraint] struct {
	internals *ZodLazyInternals
}

// ZodLazyTyped is a type-safe wrapper that preserves the inner schema type.
type ZodLazyTyped[S ZodSchemaType] struct {
	*ZodLazy[any]
	getter func() S
}

// =============================================================================
// Core Interface Methods
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

// Coerce delegates coercion to the inner schema.
func (z *ZodLazy[T]) Coerce(input any) (any, bool) {
	inner := z.resolveInner()
	if inner == nil {
		return input, false
	}
	if c, ok := inner.(interface{ Coerce(any) (any, bool) }); ok {
		if result, ok := c.Coerce(input); ok {
			return result, true
		}
	}
	return input, false
}

// =============================================================================
// Parsing Methods
// =============================================================================

// Parse validates and returns the parsed value using lazy evaluation.
func (z *ZodLazy[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	pc := &core.ParseContext{}
	if len(ctx) > 0 && ctx[0] != nil {
		pc = ctx[0]
	}

	in := &z.internals.ZodTypeInternals

	if input == nil {
		if in.NonOptional {
			var zero T
			return zero, issues.CreateNonOptionalError(pc)
		}
		if in.DefaultValue != nil {
			return any(in.DefaultValue).(T), nil //nolint:unconvert
		}
		if in.DefaultFunc != nil {
			return any(in.DefaultFunc()).(T), nil //nolint:unconvert
		}
		switch {
		case in.PrefaultValue != nil:
			input = in.PrefaultValue
		case in.PrefaultFunc != nil:
			input = in.PrefaultFunc()
		case in.Optional || in.Nilable:
			var zero T
			return zero, nil
		default:
			var zero T
			return zero, issues.CreateInvalidTypeError(core.ZodTypeLazy, nil, pc)
		}
	}

	result, err := z.validateLazy(input, in.Checks, pc)
	if err != nil {
		var zero T
		return zero, err
	}
	return z.convertResult(result), nil
}

// MustParse validates input and panics on error.
func (z *ZodLazy[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodLazy[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse requires exact type matching for compile-time safety.
func (z *ZodLazy[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	result, err := engine.ParseComplexStrict[any](
		any(input),
		&z.internals.ZodTypeInternals,
		core.ZodTypeLazy,
		z.extractType,
		z.extractPtr,
		z.validateLazy,
		ctx...,
	)
	if err != nil {
		var zero T
		return zero, err
	}
	return z.convertResult(result), nil
}

// MustStrictParse validates with strict type matching and panics on error.
func (z *ZodLazy[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// Modifier Methods
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

// Nullish combines optional and nilable modifiers.
func (z *ZodLazy[T]) Nullish() *ZodLazy[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the Optional flag and enforces non-nil value.
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

// Default sets a value returned when input is nil, bypassing validation.
func (z *ZodLazy[T]) Default(v any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory for the default value when input is nil.
func (z *ZodLazy[T]) DefaultFunc(fn func() any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(fn)
	return z.withInternals(in)
}

// Prefault sets a value that goes through full validation when input is nil.
func (z *ZodLazy[T]) Prefault(v any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a factory for the prefault value through full validation.
func (z *ZodLazy[T]) PrefaultFunc(fn func() any) *ZodLazy[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(fn)
	return z.withInternals(in)
}

// =============================================================================
// Metadata Methods
// =============================================================================

// Meta stores metadata for this lazy schema.
func (z *ZodLazy[T]) Meta(meta core.GlobalMeta) *ZodLazy[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodLazy[T]) Describe(description string) *ZodLazy[T] {
	in := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// Transformation and Pipeline Methods
// =============================================================================

// Transform creates a type-safe transformation pipeline.
func (z *ZodLazy[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrap := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(any(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrap)
}

// Pipe creates a validation pipeline to another schema.
func (z *ZodLazy[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrap := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(any(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrap)
}

// =============================================================================
// Refinement Methods
// =============================================================================

// Refine adds a typed custom validation function.
func (z *ZodLazy[T]) Refine(fn func(T) bool, params ...any) *ZodLazy[T] {
	wrap := func(v any) bool {
		if typed, ok := v.(T); ok {
			return fn(typed)
		}
		return false
	}
	return z.withCheck(checks.NewCustom[any](wrap, utils.NormalizeCustomParams(params...)))
}

// RefineAny adds a custom validation function accepting any input.
func (z *ZodLazy[T]) RefineAny(fn func(any) bool, params ...any) *ZodLazy[T] {
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
}

// Unwrap resolves and returns the inner schema.
func (z *ZodLazy[T]) Unwrap() core.ZodType[any] {
	return z.resolveInner()
}

// =============================================================================
// Internal Helper Methods
// =============================================================================

func (z *ZodLazy[T]) withCheck(c core.ZodCheck) *ZodLazy[T] {
	in := z.internals.Clone()
	in.AddCheck(c)
	return z.withInternals(in)
}

func (z *ZodLazy[T]) convertResult(result any) T {
	var zero T
	if result == nil {
		return zero
	}
	switch any(zero).(type) {
	case *any:
		return any(&result).(T)
	default:
		return any(result).(T) //nolint:unconvert
	}
}

func (z *ZodLazy[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodLazy[*any] {
	return &ZodLazy[*any]{internals: z.cloneState(in)}
}

func (z *ZodLazy[T]) withInternals(in *core.ZodTypeInternals) *ZodLazy[T] {
	return &ZodLazy[T]{internals: z.cloneState(in)}
}

// cloneState creates a new ZodLazyInternals preserving the cached inner type.
func (z *ZodLazy[T]) cloneState(in *core.ZodTypeInternals) *ZodLazyInternals {
	cloned := &ZodLazyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Getter:           z.internals.Getter,
		innerType:        z.internals.innerType,
	}
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
	origChecks := z.internals.Checks
	z.internals.ZodTypeInternals = src.internals.ZodTypeInternals
	z.internals.Def = src.internals.Def
	z.internals.Getter = src.internals.Getter
	z.internals.innerType = src.internals.innerType
	if src.internals.innerType != nil {
		z.internals.once.Do(func() {})
	}
	z.internals.Checks = origChecks
}

// resolveInner implements lazy evaluation with thread-safe caching.
func (z *ZodLazy[T]) resolveInner() core.ZodType[any] {
	z.internals.once.Do(func() {
		raw := z.internals.Getter()
		z.internals.innerType = convertToAnyInterface(raw)
	})
	return z.internals.innerType
}

// extractType accepts any value from input.
func (z *ZodLazy[T]) extractType(value any) (any, bool) {
	return value, true
}

// extractPtr extracts a pointer value from input.
func (z *ZodLazy[T]) extractPtr(value any) (*any, bool) {
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
	inner := z.resolveInner()
	if inner == nil {
		return nil, false
	}
	result, err := inner.Parse(value)
	if err != nil {
		return nil, false
	}
	return &result, true
}

// validateLazy validates that the lazy schema can be resolved and the value is valid.
func (z *ZodLazy[T]) validateLazy(value any, chks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	if value == nil {
		in := z.Internals()
		if in.Optional || in.Nilable {
			return value, nil
		}
		return nil, newLazyTypeError(value, ctx)
	}
	inner := z.resolveInner()
	if inner == nil {
		return nil, newLazyTypeError(value, ctx)
	}
	result, err := inner.Parse(value, ctx)
	if err != nil {
		if isExpectedLazyError(err) {
			return engine.ApplyChecks[any](value, chks, ctx)
		}
		return nil, err
	}
	return engine.ApplyChecks[any](result, chks, ctx)
}

func newLazyTypeError(value any, ctx *core.ParseContext) error {
	raw := issues.NewRawIssue(core.InvalidType, value, issues.WithExpected(string(core.ZodTypeLazy)))
	return issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(raw, ctx, nil)})
}

// =============================================================================
// Schema Wrapper Implementation
// =============================================================================

func convertToAnyInterface(schema any) core.ZodType[any] {
	if schema == nil {
		return nil
	}
	if s, ok := schema.(core.ZodType[any]); ok {
		return s
	}
	return &schemaWrapper{inner: schema}
}

// schemaWrapper implements ZodType[any] by wrapping any schema type.
type schemaWrapper struct{ inner any }

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
func (w *schemaWrapper) Inner() any { return w.inner }

// =============================================================================
// Constructor Functions
// =============================================================================

func newZodLazyFromDef[T LazyConstraint](def *ZodLazyDef) *ZodLazy[T] {
	in := &ZodLazyInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:    def,
		Getter: def.Getter,
	}
	in.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		ld := &ZodLazyDef{
			ZodTypeDef: *newDef,
			Getter:     def.Getter,
		}
		return any(newZodLazyFromDef[T](ld)).(core.ZodType[any])
	}
	if def.Error != nil {
		in.Error = def.Error
	}
	return &ZodLazy[T]{internals: in}
}

// Lazy creates a type-safe lazy schema with compile-time type checking.
func Lazy[S ZodSchemaType](getter func() S, params ...any) *ZodLazyTyped[S] {
	sp := utils.NormalizeParams(params...)
	def := &ZodLazyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeLazy,
			Checks: []core.ZodCheck{},
		},
		Getter: func() any { return getter() },
	}
	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}
	return newZodLazyTypedFromDef(def, getter)
}

// LazyAny creates a lazy schema that defers evaluation until needed.
func LazyAny(getter func() any, params ...any) *ZodLazy[any] {
	return LazyTyped[any](getter, params...)
}

// LazyPtr creates a lazy schema for *any type.
func LazyPtr(getter func() any, params ...any) *ZodLazy[*any] {
	return LazyTyped[*any](getter, params...)
}

// LazyTyped is the underlying generic constructor for lazy schemas.
func LazyTyped[T LazyConstraint](getter func() any, params ...any) *ZodLazy[T] {
	sp := utils.NormalizeParams(params...)
	def := &ZodLazyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeLazy,
			Checks: []core.ZodCheck{},
		},
		Getter: getter,
	}
	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}
	return newZodLazyFromDef[T](def)
}

// =============================================================================
// ZodLazyTyped Methods
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
func (z *ZodLazyTyped[S]) InnerSchema() S {
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

// NonOptional removes the Optional flag and enforces non-nil value.
func (z *ZodLazyTyped[S]) NonOptional() *ZodLazy[any] {
	return z.ZodLazy.NonOptional()
}

// Default sets a value returned when input is nil, bypassing validation.
func (z *ZodLazyTyped[S]) Default(v any) *ZodLazyTyped[S] {
	newLazy := z.ZodLazy.Default(v)
	return &ZodLazyTyped[S]{ZodLazy: newLazy, getter: z.getter}
}

// Refine adds a custom validation function.
func (z *ZodLazyTyped[S]) Refine(fn func(any) bool, params ...any) *ZodLazyTyped[S] {
	newLazy := z.RefineAny(fn, params...)
	return &ZodLazyTyped[S]{ZodLazy: newLazy, getter: z.getter}
}

func newZodLazyTypedFromDef[S ZodSchemaType](def *ZodLazyDef, getter func() S) *ZodLazyTyped[S] {
	return &ZodLazyTyped[S]{
		ZodLazy: newZodLazyFromDef[any](def),
		getter:  getter,
	}
}

// =============================================================================
// Utility Functions
// =============================================================================

// isExpectedLazyError reports whether the error is an invalid_type issue
// with expected type "lazy".
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
