package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
)

// =============================================================================
// ISO-related helpers & schema
// =============================================================================

type IsoConstraint interface {
	string | *string
}

type ZodIso[T IsoConstraint] struct{ *ZodString[T] }

// -----------------------------------------------------------------------------
// Precision helper constants
// -----------------------------------------------------------------------------

var (
	PrecisionMinute      = func() *int { i := -1; return &i }()
	PrecisionSecond      = func() *int { i := 0; return &i }()
	PrecisionDecisecond  = func() *int { i := 1; return &i }()
	PrecisionCentisecond = func() *int { i := 2; return &i }()
	PrecisionMillisecond = func() *int { i := 3; return &i }()
	PrecisionMicrosecond = func() *int { i := 6; return &i }()
	PrecisionNanosecond  = func() *int { i := 9; return &i }()
)

// IsoDatetimeOptions mirrors the one from internal/checks but kept public for API compatibility.
// Fields map 1-to-1 so we can convert directly.

type IsoDatetimeOptions struct {
	Precision *int // see Precision* constants
	Offset    bool // allow timezone offsets
	Local     bool // make trailing "Z" optional
}

type IsoTimeOptions struct {
	Precision *int // fraction seconds precision (nil = any, -1 minute precision)
}

// -----------------------------------------------------------------------------
// internal helpers
// -----------------------------------------------------------------------------

func newIsoFromString[T IsoConstraint](str *ZodString[T]) *ZodIso[T] { return &ZodIso[T]{str} }

func (z *ZodIso[T]) cloneWithCheck(check core.ZodCheck) *ZodIso[T] {
	in := z.GetInternals().Clone()
	in.AddCheck(check)
	return newIsoFromString(z.ZodString.withInternals(in))
}

// -----------------------------------------------------------------------------
// generic modifiers returning *ZodIso
// -----------------------------------------------------------------------------

func (z *ZodIso[T]) Optional() *ZodIso[*string] { return newIsoFromString(z.ZodString.Optional()) }
func (z *ZodIso[T]) Nilable() *ZodIso[*string]  { return newIsoFromString(z.ZodString.Nilable()) }
func (z *ZodIso[T]) Nullish() *ZodIso[*string]  { return newIsoFromString(z.ZodString.Nullish()) }

func (z *ZodIso[T]) Default(v string) *ZodIso[T] { return newIsoFromString(z.ZodString.Default(v)) }
func (z *ZodIso[T]) DefaultFunc(fn func() string) *ZodIso[T] {
	return newIsoFromString(z.ZodString.DefaultFunc(fn))
}
func (z *ZodIso[T]) Prefault(v string) *ZodIso[T] { return newIsoFromString(z.ZodString.Prefault(v)) }
func (z *ZodIso[T]) PrefaultFunc(fn func() string) *ZodIso[T] {
	return newIsoFromString(z.ZodString.PrefaultFunc(fn))
}

// Range validation using lexicographic comparison (perfect for ISO 8601 strings).
func (z *ZodIso[T]) Min(minVal string, params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.StringGte(minVal, params...))
}
func (z *ZodIso[T]) Max(maxVal string, params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.StringLte(maxVal, params...))
}

// -----------------------------------------------------------------------------
// ISO-specific validation helpers
// -----------------------------------------------------------------------------

func (z *ZodIso[T]) DateTime(opts ...IsoDatetimeOptions) *ZodIso[T] {
	if len(opts) > 0 {
		o := opts[0]
		c := checks.ISODateTimeOptions{Precision: o.Precision, Offset: o.Offset, Local: o.Local}
		return z.cloneWithCheck(checks.ISODateTimeWithOptions(c))
	}
	return z.cloneWithCheck(checks.ISODateTime())
}

func (z *ZodIso[T]) Date(params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.ISODate(params...))
}

func (z *ZodIso[T]) Time(opts ...IsoTimeOptions) *ZodIso[T] {
	if len(opts) > 0 {
		o := opts[0]
		c := checks.ISOTimeOptions{Precision: o.Precision}
		return z.cloneWithCheck(checks.ISOTimeWithOptions(c))
	}
	return z.cloneWithCheck(checks.ISOTime())
}

func (z *ZodIso[T]) Duration(params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.ISODuration(params...))
}

// StrictParse validates the input using strict parsing rules
func (z *ZodIso[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodIso[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// -----------------------------------------------------------------------------
// constructors
// -----------------------------------------------------------------------------

func Iso(params ...any) *ZodIso[string]     { return IsoTyped[string](params...) }
func IsoPtr(params ...any) *ZodIso[*string] { return IsoTyped[*string](params...) }

func IsoTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	// Simply forward parameters to StringTyped – users may supply SchemaParams
	str := StringTyped[T](params...)
	return newIsoFromString(str)
}

// Convenience helpers

func IsoDateTime(params ...any) *ZodIso[string]     { return IsoDateTimeTyped[string](params...) }
func IsoDateTimePtr(params ...any) *ZodIso[*string] { return IsoDateTimeTyped[*string](params...) }

func IsoDateTimeTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	var opt *IsoDatetimeOptions
	var rest []any
	for _, p := range params {
		switch v := p.(type) {
		case IsoDatetimeOptions:
			opt = &v
		default:
			rest = append(rest, p)
		}
	}
	schema := IsoTyped[T](rest...)
	if opt != nil {
		return schema.DateTime(*opt)
	}
	return schema.DateTime()
}

func IsoDate(params ...any) *ZodIso[string]     { return IsoDateTyped[string](params...) }
func IsoDatePtr(params ...any) *ZodIso[*string] { return IsoDateTyped[*string](params...) }
func IsoDateTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	return IsoTyped[T](params...).Date()
}

func IsoTime(params ...any) *ZodIso[string]     { return IsoTimeTyped[string](params...) }
func IsoTimePtr(params ...any) *ZodIso[*string] { return IsoTimeTyped[*string](params...) }
func IsoTimeTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	var opt *IsoTimeOptions
	var rest []any
	for _, p := range params {
		switch v := p.(type) {
		case IsoTimeOptions:
			opt = &v
		default:
			rest = append(rest, p)
		}
	}
	schema := IsoTyped[T](rest...)
	if opt != nil {
		return schema.Time(*opt)
	}
	return schema.Time()
}

func IsoDuration(params ...any) *ZodIso[string]     { return IsoDurationTyped[string](params...) }
func IsoDurationPtr(params ...any) *ZodIso[*string] { return IsoDurationTyped[*string](params...) }
func IsoDurationTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	return IsoTyped[T](params...).Duration(params...)
}

// -----------------------------------------------------------------------------
// Proxy GetInternals
// -----------------------------------------------------------------------------
func (z *ZodIso[T]) GetInternals() *core.ZodTypeInternals { return z.ZodString.GetInternals() }
