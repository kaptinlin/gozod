package types

import (
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/transform"
	"github.com/kaptinlin/gozod/pkg/validate"
	"golang.org/x/text/unicode/norm"
)

// StringConstraint restricts values to string or *string.
type StringConstraint interface {
	~string | ~*string
}

// ZodStringDef holds the configuration for a string schema.
type ZodStringDef struct {
	core.ZodTypeDef
}

// ZodStringInternals holds string validator internal state.
type ZodStringInternals struct {
	core.ZodTypeInternals
	Def *ZodStringDef
}

// ZodString is a type-safe string validation schema.
type ZodString[T StringConstraint] struct {
	internals *ZodStringInternals
}

// Internals returns the internal state of the schema.
func (z *ZodString[T]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodString[T]) IsOptional() bool { return z.internals.IsOptional() }

// IsNilable reports whether this schema accepts nil values.
func (z *ZodString[T]) IsNilable() bool { return z.internals.IsNilable() }

// Coerce converts input to string.
func (z *ZodString[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse validates input and returns a value of type T.
func (z *ZodString[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[string, T](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		engine.ApplyChecks[string],
		engine.ConvertToConstraintType[string, T],
		ctx...,
	)
}

// MustParse validates input and panics on failure.
func (z *ZodString[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse requires exact type T for compile-time type safety.
func (z *ZodString[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[string, T](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		engine.ApplyChecks[string],
		ctx...,
	)
}

// MustStrictParse requires exact type T and panics on failure.
func (z *ZodString[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// MustParseAny validates input and panics on failure.
func (z *ZodString[T]) MustParseAny(input any, ctx ...*core.ParseContext) any {
	result, err := z.ParseAny(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any for runtime interface usage.
func (z *ZodString[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional returns a new schema that accepts nil, with *string output.
func (z *ZodString[T]) Optional() *ZodString[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional returns a new schema that accepts absent keys but rejects explicit nil.
// Unlike Optional, it only accepts absent keys in object fields.
func (z *ZodString[T]) ExactOptional() *ZodString[T] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a new schema that accepts nil values, with *string constraint.
func (z *ZodString[T]) Nilable() *ZodString[*string] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a new schema that combines optional and nilable modifiers.
func (z *ZodString[T]) Nullish() *ZodString[*string] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits validation).
func (z *ZodString[T]) Default(v string) *ZodString[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits validation).
func (z *ZodString[T]) DefaultFunc(fn func() string) *ZodString[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodString[T]) Prefault(v string) *ZodString[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodString[T]) PrefaultFunc(fn func() string) *ZodString[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta attaches metadata to this schema via the global registry.
func (z *ZodString[T]) Meta(meta core.GlobalMeta) *ZodString[T] {
	return z.withMeta(meta)
}

// Min adds minimum length validation.
func (z *ZodString[T]) Min(minLen int, params ...any) *ZodString[T] {
	return z.withCheck(checks.MinLength(minLen, params...))
}

// Max adds maximum length validation.
func (z *ZodString[T]) Max(maxLen int, params ...any) *ZodString[T] {
	return z.withCheck(checks.MaxLength(maxLen, params...))
}

// Length adds exact length validation.
func (z *ZodString[T]) Length(length int, params ...any) *ZodString[T] {
	return z.withCheck(checks.Length(length, params...))
}

// Regex adds custom regex validation.
func (z *ZodString[T]) Regex(pattern *regexp.Regexp, params ...any) *ZodString[T] {
	return z.withCheck(checks.Regex(pattern, params...))
}

// RegexString adds custom regex validation using a string pattern.
func (z *ZodString[T]) RegexString(pattern string, params ...any) *ZodString[T] {
	return z.Regex(regexp.MustCompile(pattern), params...)
}

// StartsWith adds prefix validation.
func (z *ZodString[T]) StartsWith(prefix string, params ...any) *ZodString[T] {
	return z.withCheck(checks.StartsWith(prefix, params...))
}

// EndsWith adds suffix validation.
func (z *ZodString[T]) EndsWith(suffix string, params ...any) *ZodString[T] {
	return z.withCheck(checks.EndsWith(suffix, params...))
}

// Includes adds substring validation.
func (z *ZodString[T]) Includes(substring string, params ...any) *ZodString[T] {
	return z.withCheck(checks.Includes(substring, params...))
}

// Trim adds string trimming transformation.
func (z *ZodString[T]) Trim(params ...any) *ZodString[T] {
	return z.Overwrite(func(val T) T { return applyStringTransform(val, strings.TrimSpace) }, params...)
}

// JSON adds JSON format validation.
func (z *ZodString[T]) JSON(params ...any) *ZodString[T] {
	return z.withCheck(checks.JSON(params...))
}

// Email adds email format validation.
func (z *ZodString[T]) Email(params ...any) *ZodString[T] {
	return z.withCheck(checks.Email(params...))
}

// MAC adds MAC address format validation.
//
// Accepts an optional delimiter string (":", "-", ".") or [validate.MACOptions].
// Without arguments, defaults to colon delimiter.
//
//	MAC()                                     // default ":"
//	MAC("-")                                  // dash delimiter
//	MAC(validate.MACOptions{Delimiter: "."})  // full options
func (z *ZodString[T]) MAC(params ...any) *ZodString[T] {
	if len(params) == 0 {
		return z.withCheck(checks.MAC())
	}

	// Single param: check for delimiter or MACOptions.
	if len(params) == 1 {
		switch v := params[0].(type) {
		case string:
			if isDelimiter(v) {
				return z.withCheck(checks.MACWithDelimiter(v))
			}
			return z.withCheck(checks.MAC(v))
		case validate.MACOptions:
			return z.withCheck(checks.MACWithOptions(v))
		}
		return z.withCheck(checks.MAC(params...))
	}

	// Multiple params: first may be a delimiter followed by error/options.
	if delim, ok := params[0].(string); ok && isDelimiter(delim) {
		return z.withCheck(checks.MACWithDelimiter(delim, params[1:]...))
	}
	return z.withCheck(checks.MAC(params...))
}

// JWT adds JWT token format validation.
//
// Accepts an optional algorithm string ("HS256", "RS256", etc.) or [validate.JWTOptions].
//
//	JWT()                                        // standard format
//	JWT("HS256")                                 // require specific algorithm
//	JWT(validate.JWTOptions{Algorithm: "RS256"}) // full options
func (z *ZodString[T]) JWT(params ...any) *ZodString[T] {
	if len(params) == 0 {
		return z.withCheck(checks.JWT())
	}

	if len(params) == 1 {
		switch v := params[0].(type) {
		case string:
			if isJWTAlgorithm(v) {
				return z.withCheck(checks.JWTWithAlgorithm(v))
			}
			return z.withCheck(checks.JWT(v))
		case validate.JWTOptions:
			return z.withCheck(checks.JWTWithOptions(v))
		}
		return z.withCheck(checks.JWT(params...))
	}

	if alg, ok := params[0].(string); ok && isJWTAlgorithm(alg) {
		return z.withCheck(checks.JWTWithAlgorithm(alg, params[1:]...))
	}
	return z.withCheck(checks.JWT(params...))
}

// isDelimiter reports whether s is a recognized MAC address delimiter.
func isDelimiter(s string) bool {
	return len(s) == 1 || s == ":" || s == "-" || s == "."
}

// isJWTAlgorithm reports whether s looks like a JWA algorithm identifier.
func isJWTAlgorithm(s string) bool {
	return len(s) > 0 && len(s) <= 10 && (strings.HasPrefix(s, "HS") || strings.HasPrefix(s, "RS") ||
		strings.HasPrefix(s, "ES") || strings.HasPrefix(s, "PS") ||
		strings.HasPrefix(s, "Ed") || s == "none")
}

// Describe attaches a description to this schema via the global registry.
func (z *ZodString[T]) Describe(description string) *ZodString[T] {
	return z.withMeta(core.GlobalMeta{Description: description})
}

// Transform applies a transformation function to the validated string.
func (z *ZodString[T]) Transform(fn func(string, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	return core.NewZodTransform(z, func(input T, ctx *core.RefinementContext) (any, error) {
		str := extractString(input)
		return fn(str, ctx)
	})
}

// Overwrite applies a same-type transformation to the validated string.
func (z *ZodString[T]) Overwrite(fn func(T) T, params ...any) *ZodString[T] {
	wrapped := func(input any) any {
		converted, ok := convertToStringType[T](input)
		if !ok {
			return input
		}
		return fn(converted)
	}
	return z.withCheck(checks.NewZodCheckOverwrite(wrapped, params...))
}

// Pipe creates a pipeline with another schema.
func (z *ZodString[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	return core.NewZodPipe(z, target, func(input T, ctx *core.ParseContext) (any, error) {
		str := extractString(input)
		return target.Parse(str, ctx)
	})
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodString[T]) Check(fn func(value T, payload *core.ParsePayload), params ...any) *ZodString[T] {
	wrapped := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(T); ok {
			fn(val, payload)
			return
		}
		// Handle pointer type: T is *string but value is string.
		var zero T
		if _, ok := any(zero).(*string); ok {
			if strVal, ok := payload.Value().(string); ok {
				strCopy := strVal
				fn(any(&strCopy).(T), payload)
			}
		}
	}
	return z.withCheck(checks.NewCustom[T](wrapped, utils.NormalizeCustomParams(params...)))
}

// With is an alias for Check (Zod v4 API compatibility).
func (z *ZodString[T]) With(fn func(value T, payload *core.ParsePayload), params ...any) *ZodString[T] {
	return z.Check(fn, params...)
}

// Refine applies a custom validation function matching the schema's output type T.
func (z *ZodString[T]) Refine(fn func(T) bool, params ...any) *ZodString[T] {
	wrapper := func(v any) bool {
		var zero T
		switch any(zero).(type) {
		case string:
			if v == nil {
				return false
			}
			if strVal, ok := v.(string); ok {
				return fn(any(strVal).(T))
			}
			return false
		case *string:
			if v == nil {
				return true // Allow nil for nilable types
			}
			if strVal, ok := v.(string); ok {
				sCopy := strVal
				return fn(any(&sCopy).(T))
			}
			return false
		default:
			return false
		}
	}
	return z.withCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
}

// RefineAny adds custom validation that receives the raw value as any.
func (z *ZodString[T]) RefineAny(fn func(any) bool, params ...any) *ZodString[T] {
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
}

// And creates an intersection with another schema.
//
// Example:
//
//	schema := gozod.String().Min(3).And(gozod.String().Max(10))
func (z *ZodString[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
//
// Example:
//
//	schema := gozod.String().Or(gozod.Int())
func (z *ZodString[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// expectedType returns the schema's type code, defaulting to ZodTypeString.
func (z *ZodString[T]) expectedType() core.ZodTypeCode {
	if z.internals.Type != "" {
		return z.internals.Type
	}
	return core.ZodTypeString
}

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodString[T]) withCheck(check core.ZodCheck) *ZodString[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withMeta clones internals, merges metadata, and returns a new schema.
func (z *ZodString[T]) withMeta(meta core.GlobalMeta) *ZodString[T] {
	clone := z.withInternals(&z.internals.ZodTypeInternals)
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		core.GlobalRegistry.Add(clone, meta)
		return clone
	}
	if meta.ID != "" {
		existing.ID = meta.ID
	}
	if meta.Title != "" {
		existing.Title = meta.Title
	}
	if meta.Description != "" {
		existing.Description = meta.Description
	}
	if len(meta.Examples) > 0 {
		existing.Examples = meta.Examples
	}
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// withPtrInternals creates a new ZodString instance of type *string.
func (z *ZodString[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodString[*string] {
	clone := &ZodString[*string]{
		internals: &ZodStringInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
	if meta, ok := core.GlobalRegistry.Get(z); ok {
		core.GlobalRegistry.Add(clone, meta)
	}
	return clone
}

// withInternals creates a new ZodString instance keeping the original generic type T.
func (z *ZodString[T]) withInternals(in *core.ZodTypeInternals) *ZodString[T] {
	clone := &ZodString[T]{
		internals: &ZodStringInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
	if meta, ok := core.GlobalRegistry.Get(z); ok {
		core.GlobalRegistry.Add(clone, meta)
	}
	return clone
}

// CloneFrom copies the internal state from another schema.
func (z *ZodString[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodString[T]); ok && src != nil {
		z.internals = &ZodStringInternals{
			ZodTypeInternals: *src.internals.Clone(),
			Def:              src.internals.Def,
		}
	}
}

// applyStringTransform applies fn to the underlying string value of val.
func applyStringTransform[T StringConstraint](val T, fn func(string) string) T {
	switch v := any(val).(type) {
	case string:
		return any(fn(v)).(T)
	case *string:
		if v == nil {
			return val
		}
		result := fn(*v)
		return any(&result).(T)
	default:
		return val
	}
}

// convertToStringType converts v to the constrained string type T.
func convertToStringType[T StringConstraint](v any) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case string:
		if str, ok := v.(string); ok {
			return any(str).(T), true
		}
	case *string:
		if v == nil {
			return any((*string)(nil)).(T), true
		}
		if str, ok := v.(string); ok {
			sCopy := str
			return any(&sCopy).(T), true
		}
		if strPtr, ok := v.(*string); ok {
			return any(strPtr).(T), true
		}
	}

	return zero, false
}

// extractString returns the underlying string from a StringConstraint value.
func extractString[T StringConstraint](value T) string {
	switch v := any(value).(type) {
	case string:
		return v
	case *string:
		if v == nil {
			return ""
		}
		return *v
	default:
		return ""
	}
}

// newZodStringFromDef constructs a new ZodString from the given definition.
func newZodStringFromDef[T StringConstraint](def *ZodStringDef) *ZodString[T] {
	internals := &ZodStringInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   core.ZodTypeString,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Provide a constructor so that AddCheck can create new schema instances.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		stringDef := &ZodStringDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodStringFromDef[T](stringDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodString[T]{
		internals: internals,
	}
}

// String creates a new string schema.
func String(params ...any) *ZodString[string] {
	return StringTyped[string](params...)
}

// StringPtr creates a new string schema with pointer type.
func StringPtr(params ...any) *ZodString[*string] {
	return StringTyped[*string](params...)
}

// StringTyped creates a new string schema with a specific constraint type.
func StringTyped[T StringConstraint](params ...any) *ZodString[T] {
	schemaParams := utils.NormalizeParams(params...)
	def := &ZodStringDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:     core.ZodTypeString,
			Required: true,
			Checks:   []core.ZodCheck{},
		},
	}

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodStringFromDef[T](def)
}

// CoercedString creates a new string schema with coercion enabled.
func CoercedString(params ...any) *ZodString[string] {
	schema := StringTyped[string](params...)
	schema.internals.Coerce = true
	return schema
}

// CoercedStringPtr creates a new string schema with pointer type and coercion.
func CoercedStringPtr(params ...any) *ZodString[*string] {
	schema := StringTyped[*string](params...)
	schema.internals.Coerce = true
	return schema
}

// Lowercase validates that the string contains no uppercase letters.
// Matches Zod v4's regex: /^[^A-Z]*$/
func (z *ZodString[T]) Lowercase(params ...any) *ZodString[T] {
	return z.withCheck(checks.Lowercase(params...))
}

// Uppercase validates that the string contains no lowercase letters.
// Matches Zod v4's regex: /^[^a-z]*$/
func (z *ZodString[T]) Uppercase(params ...any) *ZodString[T] {
	return z.withCheck(checks.Uppercase(params...))
}

// ToLowerCase transforms the string to lower case.
func (z *ZodString[T]) ToLowerCase(params ...any) *ZodString[T] {
	return z.Overwrite(func(val T) T { return applyStringTransform(val, strings.ToLower) }, params...)
}

// ToUpperCase transforms the string to upper case.
func (z *ZodString[T]) ToUpperCase(params ...any) *ZodString[T] {
	return z.Overwrite(func(val T) T { return applyStringTransform(val, strings.ToUpper) }, params...)
}

// Normalize transforms the string using Unicode normalization.
// Supported forms: "NFC" (default), "NFD", "NFKC", "NFKD".
func (z *ZodString[T]) Normalize(form ...string) *ZodString[T] {
	normForm := "NFC"
	if len(form) > 0 && form[0] != "" {
		normForm = form[0]
	}
	return z.Overwrite(func(val T) T {
		return applyStringTransform(val, func(s string) string { return normalizeUnicode(s, normForm) })
	})
}

// normalizeUnicode normalizes a string using the specified Unicode form.
func normalizeUnicode(s string, form string) string {
	switch form {
	case "NFD":
		return norm.NFD.String(s)
	case "NFKC":
		return norm.NFKC.String(s)
	case "NFKD":
		return norm.NFKD.String(s)
	default: // "NFC" is the default
		return norm.NFC.String(s)
	}
}

// Slugify transforms the string to a URL-friendly slug.
func (z *ZodString[T]) Slugify(params ...any) *ZodString[T] {
	return z.Overwrite(func(val T) T { return applyStringTransform(val, transform.Slugify) }, params...)
}

// NonOptional removes the optional flag and returns a required string schema.
func (z *ZodString[T]) NonOptional() *ZodString[string] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodString[string]{
		internals: &ZodStringInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}
