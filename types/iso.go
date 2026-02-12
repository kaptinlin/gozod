package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// INTERNAL HELPERS
// =============================================================================

// intPtr returns a pointer to the given int value.
func intPtr(v int) *int { return &v }

// =============================================================================
// PRECISION CONSTANTS
// =============================================================================

// Precision constants control fractional-second digits in ISO 8601 strings.
// Use with IsoDatetimeOptions.Precision or IsoTimeOptions.Precision.
var (
	PrecisionMinute      = intPtr(-1) // no seconds component
	PrecisionSecond      = intPtr(0)  // seconds, no fraction
	PrecisionDecisecond  = intPtr(1)  // 1 fractional digit
	PrecisionCentisecond = intPtr(2)  // 2 fractional digits
	PrecisionMillisecond = intPtr(3)  // 3 fractional digits
	PrecisionMicrosecond = intPtr(6)  // 6 fractional digits
	PrecisionNanosecond  = intPtr(9)  // 9 fractional digits
)

// =============================================================================
// ISO SCHEMA TYPE
// =============================================================================

// IsoConstraint restricts generic type parameters to string or *string.
type IsoConstraint interface {
	string | *string
}

// ZodIso validates ISO 8601 formatted strings (datetime, date, time, duration).
type ZodIso[T IsoConstraint] struct{ *ZodString[T] }

// IsoDatetimeOptions configures ISO 8601 datetime validation.
type IsoDatetimeOptions struct {
	Precision *int // fractional-second digits; nil = any (see Precision* constants)
	Offset    bool // allow timezone offsets like +08:00
	Local     bool // make trailing "Z" optional
}

// IsoTimeOptions configures ISO 8601 time validation.
type IsoTimeOptions struct {
	Precision *int // fractional-second digits; nil = any, -1 = minute precision
}

func newIso[T IsoConstraint](s *ZodString[T]) *ZodIso[T] { return &ZodIso[T]{s} }

func (z *ZodIso[T]) cloneWithCheck(check core.ZodCheck) *ZodIso[T] {
	in := z.GetInternals().Clone()
	in.AddCheck(check)
	return newIso(z.withInternals(in))
}

// =============================================================================
// MODIFIERS
// =============================================================================

// Optional returns a new schema that accepts nil values.
func (z *ZodIso[T]) Optional() *ZodIso[*string] { return newIso(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodIso[T]) Nilable() *ZodIso[*string] { return newIso(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodIso[T]) Nullish() *ZodIso[*string] { return newIso(z.ZodString.Nullish()) }

// Default returns a new schema that uses v when input is nil (bypasses validation).
func (z *ZodIso[T]) Default(v string) *ZodIso[T] { return newIso(z.ZodString.Default(v)) }

// DefaultFunc returns a new schema that calls fn when input is nil (bypasses validation).
func (z *ZodIso[T]) DefaultFunc(fn func() string) *ZodIso[T] {
	return newIso(z.ZodString.DefaultFunc(fn))
}

// Prefault returns a new schema that uses v when input is nil (runs through full validation).
func (z *ZodIso[T]) Prefault(v string) *ZodIso[T] { return newIso(z.ZodString.Prefault(v)) }

// PrefaultFunc returns a new schema that calls fn when input is nil (runs through full validation).
func (z *ZodIso[T]) PrefaultFunc(fn func() string) *ZodIso[T] {
	return newIso(z.ZodString.PrefaultFunc(fn))
}

// Min validates that the ISO string is >= minVal using lexicographic comparison.
func (z *ZodIso[T]) Min(minVal string, params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.StringGte(minVal, params...))
}

// Max validates that the ISO string is <= maxVal using lexicographic comparison.
func (z *ZodIso[T]) Max(maxVal string, params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.StringLte(maxVal, params...))
}

// =============================================================================
// ISO-SPECIFIC VALIDATION
// =============================================================================

// DateTime adds ISO 8601 datetime validation to the schema.
func (z *ZodIso[T]) DateTime(opts ...IsoDatetimeOptions) *ZodIso[T] {
	if len(opts) > 0 {
		o := opts[0]
		return z.cloneWithCheck(checks.ISODateTimeWithOptions(validate.ISODateTimeOptions{
			Precision: o.Precision, Offset: o.Offset, Local: o.Local,
		}))
	}
	return z.cloneWithCheck(checks.ISODateTime())
}

// Date adds ISO 8601 date validation to the schema.
func (z *ZodIso[T]) Date(params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.ISODate(params...))
}

// Time adds ISO 8601 time validation to the schema.
func (z *ZodIso[T]) Time(opts ...IsoTimeOptions) *ZodIso[T] {
	if len(opts) > 0 {
		return z.cloneWithCheck(checks.ISOTimeWithOptions(validate.ISOTimeOptions{
			Precision: opts[0].Precision,
		}))
	}
	return z.cloneWithCheck(checks.ISOTime())
}

// Duration adds ISO 8601 duration validation to the schema.
func (z *ZodIso[T]) Duration(params ...any) *ZodIso[T] {
	return z.cloneWithCheck(checks.ISODuration(params...))
}

// StrictParse validates the input with compile-time type safety.
func (z *ZodIso[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodIso[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// CONSTRUCTORS
// =============================================================================

// Iso creates a base ISO string schema without format-specific validation.
func Iso(params ...any) *ZodIso[string] { return IsoTyped[string](params...) }

// IsoPtr creates a pointer ISO string schema.
func IsoPtr(params ...any) *ZodIso[*string] { return IsoTyped[*string](params...) }

// IsoTyped creates a typed ISO string schema.
func IsoTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	return newIso(StringTyped[T](params...))
}

// IsoDateTime creates a schema that validates ISO 8601 datetime strings.
func IsoDateTime(params ...any) *ZodIso[string] { return IsoDateTimeTyped[string](params...) }

// IsoDateTimePtr creates a pointer ISO 8601 datetime schema.
func IsoDateTimePtr(params ...any) *ZodIso[*string] { return IsoDateTimeTyped[*string](params...) }

// IsoDateTimeTyped creates a typed ISO 8601 datetime schema.
// Accepts IsoDatetimeOptions in params to configure precision, offset, and local behavior.
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
	schema := IsoTyped[T](rest...)
	if opt != nil {
		return schema.DateTime(*opt)
	}
	return schema.DateTime()
}

// IsoDate creates a schema that validates ISO 8601 date strings (YYYY-MM-DD).
func IsoDate(params ...any) *ZodIso[string] { return IsoDateTyped[string](params...) }

// IsoDatePtr creates a pointer ISO 8601 date schema.
func IsoDatePtr(params ...any) *ZodIso[*string] { return IsoDateTyped[*string](params...) }

// IsoDateTyped creates a typed ISO 8601 date schema.
func IsoDateTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	return IsoTyped[T](params...).Date()
}

// IsoTime creates a schema that validates ISO 8601 time strings (HH:MM:SS).
func IsoTime(params ...any) *ZodIso[string] { return IsoTimeTyped[string](params...) }

// IsoTimePtr creates a pointer ISO 8601 time schema.
func IsoTimePtr(params ...any) *ZodIso[*string] { return IsoTimeTyped[*string](params...) }

// IsoTimeTyped creates a typed ISO 8601 time schema.
// Accepts IsoTimeOptions in params to configure precision.
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
	schema := IsoTyped[T](rest...)
	if opt != nil {
		return schema.Time(*opt)
	}
	return schema.Time()
}

// IsoDuration creates a schema that validates ISO 8601 duration strings (e.g., P1Y2M3DT4H5M6S).
func IsoDuration(params ...any) *ZodIso[string] { return IsoDurationTyped[string](params...) }

// IsoDurationPtr creates a pointer ISO 8601 duration schema.
func IsoDurationPtr(params ...any) *ZodIso[*string] { return IsoDurationTyped[*string](params...) }

// IsoDurationTyped creates a typed ISO 8601 duration schema.
func IsoDurationTyped[T IsoConstraint](params ...any) *ZodIso[T] {
	return IsoTyped[T](params...).Duration()
}

// =============================================================================
// PROXY METHODS
// =============================================================================

// GetInternals returns the schema's internal configuration.
func (z *ZodIso[T]) GetInternals() *core.ZodTypeInternals { return z.ZodString.GetInternals() }
