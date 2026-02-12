package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// Precision constants control fractional-second digits in ISO 8601 strings.
var (
	PrecisionMinute      = intPtr(-1) // no seconds component
	PrecisionSecond      = intPtr(0)  // seconds, no fraction
	PrecisionDecisecond  = intPtr(1)  // 1 fractional digit
	PrecisionCentisecond = intPtr(2)  // 2 fractional digits
	PrecisionMillisecond = intPtr(3)  // 3 fractional digits
	PrecisionMicrosecond = intPtr(6)  // 6 fractional digits
	PrecisionNanosecond  = intPtr(9)  // 9 fractional digits
)

// IsoConstraint restricts generic type parameters to string or *string.
type IsoConstraint interface {
	string | *string
}

// ZodIso validates ISO 8601 formatted strings (datetime, date, time, duration).
type ZodIso[T IsoConstraint] struct{ *ZodString[T] }

// IsoDatetimeOptions configures ISO 8601 datetime validation.
type IsoDatetimeOptions struct {
	Precision *int // fractional-second digits; nil = any
	Offset    bool // allow timezone offsets like +08:00
	Local     bool // make trailing "Z" optional
}

// IsoTimeOptions configures ISO 8601 time validation.
type IsoTimeOptions struct {
	Precision *int // fractional-second digits; nil = any, -1 = minute
}

// intPtr returns a pointer to the given integer value.
func intPtr(v int) *int {
	return &v
}

// newIso creates a new ZodIso instance wrapping the given ZodString.
func newIso[T IsoConstraint](s *ZodString[T]) *ZodIso[T] {
	return &ZodIso[T]{s}
}

// cloneWithCheck creates a new ZodIso instance with an additional validation check.
func (z *ZodIso[T]) cloneWithCheck(c core.ZodCheck) *ZodIso[T] {
	in := z.Internals().Clone()
	in.AddCheck(c)
	return newIso(z.withInternals(in))
}

// Optional returns a new schema that accepts nil values.
func (z *ZodIso[T]) Optional() *ZodIso[*string] {
	return newIso(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodIso[T]) Nilable() *ZodIso[*string] {
	return newIso(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodIso[T]) Nullish() *ZodIso[*string] {
	return newIso(z.ZodString.Nullish())
}

// Default uses v when input is nil, bypassing validation.
func (z *ZodIso[T]) Default(v string) *ZodIso[T] {
	return newIso(z.ZodString.Default(v))
}

// DefaultFunc calls fn when input is nil, bypassing validation.
func (z *ZodIso[T]) DefaultFunc(fn func() string) *ZodIso[T] {
	return newIso(z.ZodString.DefaultFunc(fn))
}

// Prefault uses v when input is nil, running through full validation.
func (z *ZodIso[T]) Prefault(v string) *ZodIso[T] {
	return newIso(z.ZodString.Prefault(v))
}

// PrefaultFunc calls fn when input is nil, running through full validation.
func (z *ZodIso[T]) PrefaultFunc(fn func() string) *ZodIso[T] {
	return newIso(z.ZodString.PrefaultFunc(fn))
}

// Min validates the ISO string is >= v using lexicographic comparison.
func (z *ZodIso[T]) Min(v string, params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.StringGte(v, params...))
}

// Max validates the ISO string is <= v using lexicographic comparison.
func (z *ZodIso[T]) Max(v string, params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.StringLte(v, params...))
}

// DateTime adds ISO 8601 datetime validation.
func (z *ZodIso[T]) DateTime(opts ...IsoDatetimeOptions) *ZodIso[T] {
	if len(opts) > 0 {
		o := opts[0]
		return z.cloneWithCheck(checks.ISODateTimeWithOptions(validate.ISODateTimeOptions{
			Precision: o.Precision,
			Offset:    o.Offset,
			Local:     o.Local,
		}))
	}
	return z.cloneWithCheck(checks.ISODateTime())
}

// Date adds ISO 8601 date validation.
func (z *ZodIso[T]) Date(params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.ISODate(params...))
}

// Time adds ISO 8601 time validation.
func (z *ZodIso[T]) Time(opts ...IsoTimeOptions) *ZodIso[T] {
	if len(opts) > 0 {
		return z.cloneWithCheck(checks.ISOTimeWithOptions(validate.ISOTimeOptions{
			Precision: opts[0].Precision,
		}))
	}
	return z.cloneWithCheck(checks.ISOTime())
}

// Duration adds ISO 8601 duration validation.
func (z *ZodIso[T]) Duration(params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.ISODuration(params...))
}

// StrictParse validates input with compile-time type safety.
func (z *ZodIso[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodIso[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// Internals returns the schema's internal configuration.
func (z *ZodIso[T]) Internals() *core.ZodTypeInternals {
	return z.ZodString.Internals()
}

// Iso creates a base ISO string schema without format-specific validation.
func Iso(params ...any) *ZodIso[string] {
	return IsoTyped[string](params...)
}

// IsoPtr creates a pointer ISO string schema.
func IsoPtr(params ...any) *ZodIso[*string] {
	return IsoTyped[*string](params...)
}

// IsoTyped creates a typed ISO string schema.
func IsoTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	return newIso(StringTyped[T](params...))
}

// IsoDateTime creates a schema validating ISO 8601 datetime strings.
func IsoDateTime(params ...any) *ZodIso[string] {
	return IsoDateTimeTyped[string](params...)
}

// IsoDateTimePtr creates a pointer ISO 8601 datetime schema.
func IsoDateTimePtr(params ...any) *ZodIso[*string] {
	return IsoDateTimeTyped[*string](params...)
}

// IsoDateTimeTyped creates a typed ISO 8601 datetime schema.
func IsoDateTimeTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	var opt *IsoDatetimeOptions
	var rest []any
	for _, p := range params {
		if v, ok := p.(IsoDatetimeOptions); ok {
			opt = &v
		} else {
			rest = append(rest, p)
		}
	}
	s := IsoTyped[T](rest...)
	if opt != nil {
		return s.DateTime(*opt)
	}
	return s.DateTime()
}

// IsoDate creates a schema validating ISO 8601 date strings (YYYY-MM-DD).
func IsoDate(params ...any) *ZodIso[string] {
	return IsoDateTyped[string](params...)
}

// IsoDatePtr creates a pointer ISO 8601 date schema.
func IsoDatePtr(params ...any) *ZodIso[*string] {
	return IsoDateTyped[*string](params...)
}

// IsoDateTyped creates a typed ISO 8601 date schema.
func IsoDateTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	return IsoTyped[T](params...).Date()
}

// IsoTime creates a schema validating ISO 8601 time strings (HH:MM:SS).
func IsoTime(params ...any) *ZodIso[string] {
	return IsoTimeTyped[string](params...)
}

// IsoTimePtr creates a pointer ISO 8601 time schema.
func IsoTimePtr(params ...any) *ZodIso[*string] {
	return IsoTimeTyped[*string](params...)
}

// IsoTimeTyped creates a typed ISO 8601 time schema.
func IsoTimeTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	var opt *IsoTimeOptions
	var rest []any
	for _, p := range params {
		if v, ok := p.(IsoTimeOptions); ok {
			opt = &v
		} else {
			rest = append(rest, p)
		}
	}
	s := IsoTyped[T](rest...)
	if opt != nil {
		return s.Time(*opt)
	}
	return s.Time()
}

// IsoDuration creates a schema validating ISO 8601 duration strings.
func IsoDuration(params ...any) *ZodIso[string] {
	return IsoDurationTyped[string](params...)
}

// IsoDurationPtr creates a pointer ISO 8601 duration schema.
func IsoDurationPtr(params ...any) *ZodIso[*string] {
	return IsoDurationTyped[*string](params...)
}

// IsoDurationTyped creates a typed ISO 8601 duration schema.
func IsoDurationTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	return IsoTyped[T](params...).Duration()
}
