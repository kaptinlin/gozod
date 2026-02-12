package types

import (
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// StringBoolConstraint restricts values to bool or *bool.
type StringBoolConstraint interface {
	bool | *bool
}

// StringBoolOptions configures truthy/falsy values and case sensitivity.
type StringBoolOptions struct {
	Truthy []string // values that evaluate to true
	Falsy  []string // values that evaluate to false
	Case   string   // "sensitive" or "insensitive"
}

// ZodStringBoolDef holds the configuration for string-boolean validation.
type ZodStringBoolDef struct {
	core.ZodTypeDef
	Truthy []string
	Falsy  []string
	Case   string
	Custom bool
}

// ZodStringBoolInternals contains the internal state of a string-boolean validator.
type ZodStringBoolInternals struct {
	core.ZodTypeInternals
	Def    *ZodStringBoolDef
	Truthy map[string]struct{}
	Falsy  map[string]struct{}
}

// ZodStringBool is a type-safe string-to-boolean validation schema.
type ZodStringBool[T StringBoolConstraint] struct {
	internals *ZodStringBoolInternals
}

// Internals returns the internal state of the schema.
func (z *ZodStringBool[T]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodStringBool[T]) IsOptional() bool { return z.internals.IsOptional() }

// IsNilable reports whether this schema accepts nil values.
func (z *ZodStringBool[T]) IsNilable() bool { return z.internals.IsNilable() }

// Coerce converts input to a recognized truthy/falsy string.
func (z *ZodStringBool[T]) Coerce(input any) (any, bool) {
	if s, err := coerce.ToString(input); err == nil {
		if _, ok := z.toBool(s); ok {
			return s, true
		}
	}
	return input, false
}

// Parse validates input and returns a value matching the generic type T.
func (z *ZodStringBool[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	r, err := engine.ParseComplex[bool](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		z.extract,
		z.extractPtr,
		engine.ApplyChecks[bool],
		ctx...,
	)
	if err != nil {
		var zero T
		return zero, err
	}
	return engine.ConvertToConstraintType[bool, T](r, core.NewParseContext(), z.expectedType())
}

// MustParse panics on validation failure.
func (z *ZodStringBool[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	r, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseAny validates input and returns an untyped result.
func (z *ZodStringBool[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse validates input with compile-time type safety.
func (z *ZodStringBool[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[bool, T](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		engine.ApplyChecks[bool],
		ctx...,
	)
}

// MustStrictParse panics on validation failure with compile-time type safety.
func (z *ZodStringBool[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodStringBool[T]) Optional() *ZodStringBool[*bool] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodStringBool[T]) ExactOptional() *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodStringBool[T]) Nilable() *ZodStringBool[*bool] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodStringBool[T]) Nullish() *ZodStringBool[*bool] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits).
func (z *ZodStringBool[T]) Default(v bool) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits).
func (z *ZodStringBool[T]) DefaultFunc(fn func() bool) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
// Accepts string input per Zod v4 semantics for StringBool.
func (z *ZodStringBool[T]) Prefault(v string) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
// Returns string per Zod v4 semantics for StringBool.
func (z *ZodStringBool[T]) PrefaultFunc(fn func() string) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata in the global registry.
func (z *ZodStringBool[T]) Meta(meta core.GlobalMeta) *ZodStringBool[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodStringBool[T]) Describe(desc string) *ZodStringBool[T] {
	in := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = desc
	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// Refine applies a custom validation function matching the schema's output type T.
func (z *ZodStringBool[T]) Refine(fn func(T) bool, params ...any) *ZodStringBool[T] {
	wrapper := func(v any) bool {
		var zero T
		switch any(zero).(type) {
		case bool:
			if v == nil {
				return false
			}
			if b, ok := v.(bool); ok {
				return fn(any(b).(T))
			}
			return false
		case *bool:
			if v == nil {
				return fn(any((*bool)(nil)).(T))
			}
			if b, ok := v.(bool); ok {
				cp := b
				return fn(any(&cp).(T))
			}
			return false
		default:
			return false
		}
	}

	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}

	check := checks.NewCustom[any](wrapper, msg)
	return z.withCheck(check)
}

// RefineAny applies a custom validation function that receives the raw value.
func (z *ZodStringBool[T]) RefineAny(fn func(any) bool, params ...any) *ZodStringBool[T] {
	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}
	check := checks.NewCustom[any](fn, msg)
	return z.withCheck(check)
}

// Transform applies a transformation function to the parsed value.
func (z *ZodStringBool[T]) Transform(fn func(bool, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapper := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(extractStringBool(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapper)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodStringBool[T]) Overwrite(transform func(T) T, params ...any) *ZodStringBool[T] {
	fn := func(input any) any {
		v, ok := convertToStringBoolType[T](input)
		if !ok {
			return input
		}
		return transform(v)
	}
	check := checks.NewZodCheckOverwrite(fn, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodStringBool[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	fn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractStringBool(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, fn)
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodStringBool[T]) Check(fn func(value T, payload *core.ParsePayload), params ...any) *ZodStringBool[T] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(T); ok {
			fn(val, payload)
			return
		}

		var zero T
		if _, ok := any(zero).(*bool); ok {
			if b, ok := payload.Value().(bool); ok {
				cp := b
				fn(any(&cp).(T), payload)
			}
		}
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// With is an alias for Check (Zod v4 API compatibility).
func (z *ZodStringBool[T]) With(fn func(value T, payload *core.ParsePayload), params ...any) *ZodStringBool[T] {
	return z.Check(fn, params...)
}

// And creates an intersection with another schema.
func (z *ZodStringBool[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodStringBool[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// NonOptional removes the optional flag, returning a bool constraint.
func (z *ZodStringBool[T]) NonOptional() *ZodStringBool[bool] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodStringBool[bool]{
		internals: &ZodStringBoolInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Truthy:           z.internals.Truthy,
			Falsy:            z.internals.Falsy,
		},
	}
}

// convertToStringBoolType converts only bool values to the target type T.
func convertToStringBoolType[T StringBoolConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		switch any(zero).(type) {
		case *bool:
			return zero, true
		default:
			return zero, false
		}
	}

	var b bool
	var ok bool

	switch val := v.(type) {
	case bool:
		b, ok = val, true
	case *bool:
		if val != nil {
			b, ok = *val, true
		}
	default:
		return zero, false
	}

	if !ok {
		return zero, false
	}

	switch any(zero).(type) {
	case bool:
		return any(b).(T), true
	case *bool:
		return any(&b).(T), true
	default:
		return zero, false
	}
}

// expectedType returns the schema's type code, defaulting to ZodTypeStringBool.
func (z *ZodStringBool[T]) expectedType() core.ZodTypeCode {
	if z.internals.Type != "" {
		return z.internals.Type
	}
	return core.ZodTypeStringBool
}

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodStringBool[T]) withCheck(check core.ZodCheck) *ZodStringBool[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withPtrInternals creates a new *bool schema from cloned internals.
func (z *ZodStringBool[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodStringBool[*bool] {
	return &ZodStringBool[*bool]{internals: &ZodStringBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Truthy:           z.internals.Truthy,
		Falsy:            z.internals.Falsy,
	}}
}

// withInternals creates a new schema preserving generic type T.
func (z *ZodStringBool[T]) withInternals(in *core.ZodTypeInternals) *ZodStringBool[T] {
	return &ZodStringBool[T]{internals: &ZodStringBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Truthy:           z.internals.Truthy,
		Falsy:            z.internals.Falsy,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodStringBool[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodStringBool[T]); ok {
		orig := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = orig
	}
}

// extractStringBool extracts the underlying bool from a generic constraint type.
func extractStringBool[T StringBoolConstraint](value T) bool {
	if ptr, ok := any(value).(*bool); ok {
		return ptr != nil && *ptr
	}
	return any(value).(bool)
}

// toBool converts a string to bool using the configured truthy/falsy sets.
func (z *ZodStringBool[T]) toBool(value string) (bool, bool) {
	normalized := value
	if z.internals.Def.Case == "insensitive" {
		normalized = strings.ToLower(value)
	}

	if _, ok := z.internals.Truthy[normalized]; ok {
		return true, true
	}
	if _, ok := z.internals.Falsy[normalized]; ok {
		return false, true
	}
	return false, false
}

// extract extracts bool from string input for ParseComplex.
func (z *ZodStringBool[T]) extract(input any) (bool, bool) {
	switch v := input.(type) {
	case string:
		return z.toBool(v)
	case *string:
		if v == nil {
			return false, false
		}
		return z.toBool(*v)
	}

	if z.internals.IsCoerce() {
		if coerced, ok := z.Coerce(input); ok {
			return z.extract(coerced)
		}
	}

	return false, false
}

// extractPtr extracts *bool from input for ParseComplex.
func (z *ZodStringBool[T]) extractPtr(input any) (*bool, bool) {
	if ptr, ok := input.(*bool); ok {
		return ptr, true
	}
	return nil, false
}

// newZodStringBoolFromDef constructs a new ZodStringBool from a definition.
func newZodStringBoolFromDef[T StringBoolConstraint](def *ZodStringBoolDef) *ZodStringBool[T] {
	in := &ZodStringBoolInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:    def,
		Truthy: make(map[string]struct{}, len(def.Truthy)),
		Falsy:  make(map[string]struct{}, len(def.Falsy)),
	}

	for _, v := range def.Truthy {
		key := v
		if def.Case == "insensitive" {
			key = strings.ToLower(v)
		}
		in.Truthy[key] = struct{}{}
	}
	for _, v := range def.Falsy {
		key := v
		if def.Case == "insensitive" {
			key = strings.ToLower(v)
		}
		in.Falsy[key] = struct{}{}
	}

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		sd := &ZodStringBoolDef{
			ZodTypeDef: *d,
			Truthy:     def.Truthy,
			Falsy:      def.Falsy,
			Case:       def.Case,
			Custom:     def.Custom,
		}
		return any(newZodStringBoolFromDef[T](sd)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodStringBool[T]{internals: in}
}

// StringBool creates a string-to-bool validation schema.
func StringBool(params ...any) *ZodStringBool[bool] {
	return StringBoolTyped[bool](params...)
}

// StringBoolPtr creates a string-to-*bool validation schema.
func StringBoolPtr(params ...any) *ZodStringBool[*bool] {
	return StringBoolTyped[*bool](params...)
}

// StringBoolTyped is the generic constructor for string-boolean schemas.
func StringBoolTyped[T StringBoolConstraint](params ...any) *ZodStringBool[T] {
	var opts *StringBoolOptions
	var rest []any

	if len(params) > 0 {
		switch v := params[0].(type) {
		case nil:
			rest = params[1:]
		case *StringBoolOptions:
			opts = v
			rest = params[1:]
		case StringBoolOptions:
			opts = &v
			rest = params[1:]
		default:
			rest = params
		}
	}

	truthy := []string{"true", "1", "yes", "on", "y", "enabled"}
	falsy := []string{"false", "0", "no", "off", "n", "disabled"}
	caseSens := "insensitive"
	custom := false

	if opts != nil {
		custom = true
		if len(opts.Truthy) > 0 {
			truthy = opts.Truthy
		}
		if len(opts.Falsy) > 0 {
			falsy = opts.Falsy
		}
		if opts.Case != "" {
			caseSens = opts.Case
		}
	}

	sp := utils.NormalizeParams(rest...)

	def := &ZodStringBoolDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStringBool,
			Checks: []core.ZodCheck{},
		},
		Truthy: truthy,
		Falsy:  falsy,
		Case:   caseSens,
		Custom: custom,
	}

	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}

	return newZodStringBoolFromDef[T](def)
}

// CoercedStringBool creates a coerced string-to-bool schema.
func CoercedStringBool(params ...any) *ZodStringBool[bool] {
	s := StringBool(params...)
	s.internals.SetCoerce(true)
	return s
}

// CoercedStringBoolPtr creates a coerced string-to-*bool schema.
func CoercedStringBoolPtr(params ...any) *ZodStringBool[*bool] {
	s := StringBoolPtr(params...)
	s.internals.SetCoerce(true)
	return s
}
