package types

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/pkg/regex"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// INTERNAL HELPERS (DRY)
// =============================================================================

// newNetworkSchema creates a network schema by adding a check to a base string schema.
// This eliminates repeated clone-addCheck-wrap boilerplate across all network types.
func newNetworkSchema[T StringConstraint](base *ZodString[T], check core.ZodCheck) *ZodString[T] {
	in := base.Internals().Clone()
	in.AddCheck(check)
	return base.withInternals(in)
}

// =============================================================================
// URLOptions (exported via gozod.go)
// =============================================================================

// URLOptions defines validation options for URL schemas.
type URLOptions struct {
	// Hostname validation pattern
	Hostname *regexp.Regexp
	// Protocol validation pattern
	Protocol *regexp.Regexp
}

// =============================================================================
// IPv4
// =============================================================================

// ZodIPv4 validates strings in IPv4 address format.
type ZodIPv4[T StringConstraint] struct{ *ZodString[T] }

func newIPv4[T StringConstraint](s *ZodString[T]) *ZodIPv4[T] { return &ZodIPv4[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodIPv4[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodIPv4[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodIPv4[T]) Optional() *ZodIPv4[*string] { return newIPv4(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodIPv4[T]) Nilable() *ZodIPv4[*string] { return newIPv4(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodIPv4[T]) Nullish() *ZodIPv4[*string] { return newIPv4(z.ZodString.Nullish()) }

// IPv4 creates an IPv4 address validation schema.
func IPv4(params ...any) *ZodIPv4[string] {
	return newIPv4(newNetworkSchema(StringTyped[string](params...), checks.IPv4()))
}

// IPv4Ptr creates a pointer IPv4 address validation schema.
func IPv4Ptr(params ...any) *ZodIPv4[*string] {
	return newIPv4(newNetworkSchema(StringPtr(params...), checks.IPv4()))
}

// =============================================================================
// IPv6
// =============================================================================

// ZodIPv6 validates strings in IPv6 address format.
type ZodIPv6[T StringConstraint] struct{ *ZodString[T] }

func newIPv6[T StringConstraint](s *ZodString[T]) *ZodIPv6[T] { return &ZodIPv6[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodIPv6[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodIPv6[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodIPv6[T]) Optional() *ZodIPv6[*string] { return newIPv6(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodIPv6[T]) Nilable() *ZodIPv6[*string] { return newIPv6(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodIPv6[T]) Nullish() *ZodIPv6[*string] { return newIPv6(z.ZodString.Nullish()) }

// IPv6 creates an IPv6 address validation schema.
func IPv6(params ...any) *ZodIPv6[string] {
	return newIPv6(newNetworkSchema(StringTyped[string](params...), checks.IPv6()))
}

// IPv6Ptr creates a pointer IPv6 address validation schema.
func IPv6Ptr(params ...any) *ZodIPv6[*string] {
	return newIPv6(newNetworkSchema(StringPtr(params...), checks.IPv6()))
}

// =============================================================================
// CIDRv4
// =============================================================================

// ZodCIDRv4 validates strings in CIDRv4 notation format.
type ZodCIDRv4[T StringConstraint] struct{ *ZodString[T] }

func newCIDRv4[T StringConstraint](s *ZodString[T]) *ZodCIDRv4[T] { return &ZodCIDRv4[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodCIDRv4[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodCIDRv4[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodCIDRv4[T]) Optional() *ZodCIDRv4[*string] { return newCIDRv4(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodCIDRv4[T]) Nilable() *ZodCIDRv4[*string] { return newCIDRv4(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodCIDRv4[T]) Nullish() *ZodCIDRv4[*string] { return newCIDRv4(z.ZodString.Nullish()) }

// CIDRv4 creates a CIDRv4 notation validation schema.
func CIDRv4(params ...any) *ZodCIDRv4[string] {
	return newCIDRv4(newNetworkSchema(StringTyped[string](params...), checks.CIDRv4()))
}

// CIDRv4Ptr creates a pointer CIDRv4 notation validation schema.
func CIDRv4Ptr(params ...any) *ZodCIDRv4[*string] {
	return newCIDRv4(newNetworkSchema(StringPtr(params...), checks.CIDRv4()))
}

// =============================================================================
// CIDRv6
// =============================================================================

// ZodCIDRv6 validates strings in CIDRv6 notation format.
type ZodCIDRv6[T StringConstraint] struct{ *ZodString[T] }

func newCIDRv6[T StringConstraint](s *ZodString[T]) *ZodCIDRv6[T] { return &ZodCIDRv6[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodCIDRv6[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodCIDRv6[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodCIDRv6[T]) Optional() *ZodCIDRv6[*string] { return newCIDRv6(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodCIDRv6[T]) Nilable() *ZodCIDRv6[*string] { return newCIDRv6(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodCIDRv6[T]) Nullish() *ZodCIDRv6[*string] { return newCIDRv6(z.ZodString.Nullish()) }

// CIDRv6 creates a CIDRv6 notation validation schema.
func CIDRv6(params ...any) *ZodCIDRv6[string] {
	return newCIDRv6(newNetworkSchema(StringTyped[string](params...), checks.CIDRv6()))
}

// CIDRv6Ptr creates a pointer CIDRv6 notation validation schema.
func CIDRv6Ptr(params ...any) *ZodCIDRv6[*string] {
	return newCIDRv6(newNetworkSchema(StringPtr(params...), checks.CIDRv6()))
}

// =============================================================================
// URL
// =============================================================================

// ZodURL validates strings in URL format.
type ZodURL[T StringConstraint] struct{ *ZodString[T] }

func newURL[T StringConstraint](s *ZodString[T]) *ZodURL[T] { return &ZodURL[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodURL[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodURL[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodURL[T]) Optional() *ZodURL[*string] { return newURL(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodURL[T]) Nilable() *ZodURL[*string] { return newURL(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodURL[T]) Nullish() *ZodURL[*string] { return newURL(z.ZodString.Nullish()) }

// URL creates a URL validation schema.
// Accepts optional URLOptions as first parameter for custom validation.
func URL(params ...any) *ZodURL[string] {
	return URLTyped[string](params...)
}

// URLPtr creates a pointer URL validation schema.
// Accepts optional URLOptions as first parameter for custom validation.
func URLPtr(params ...any) *ZodURL[*string] {
	return URLTyped[*string](params...)
}

// URLTyped creates a URL validation schema with the given type constraint.
// Accepts optional URLOptions as first parameter for custom validation.
func URLTyped[T StringConstraint](params ...any) *ZodURL[T] {
	base := StringTyped[T](params...)

	var check core.ZodCheck
	if len(params) > 0 {
		if opts, ok := params[0].(URLOptions); ok {
			check = checks.URLWithOptions(validate.URLOptions{
				Hostname: opts.Hostname,
				Protocol: opts.Protocol,
			})
		} else if opts, ok := params[0].(*URLOptions); ok && opts != nil {
			check = checks.URLWithOptions(validate.URLOptions{
				Hostname: opts.Hostname,
				Protocol: opts.Protocol,
			})
		}
	}
	if check == nil {
		check = checks.URL()
	}

	return newURL(newNetworkSchema(base, check))
}

// HttpURL creates a URL validation schema that only accepts http:// or https:// protocols.
func HttpURL(params ...any) *ZodURL[string] {
	return HttpURLTyped[string](params...)
}

// HttpURLPtr creates a pointer URL validation schema that only accepts http:// or https:// protocols.
func HttpURLPtr(params ...any) *ZodURL[*string] {
	return HttpURLTyped[*string](params...)
}

// HttpURLTyped creates a URL validation schema with the given type constraint
// that only accepts http:// or https:// protocols.
func HttpURLTyped[T StringConstraint](params ...any) *ZodURL[T] {
	base := StringTyped[T](params...)
	check := checks.URLWithOptions(validate.URLOptions{
		Protocol: regex.HTTPProtocol,
	})
	return newURL(newNetworkSchema(base, check))
}

// =============================================================================
// Hostname
// =============================================================================

// ZodHostname validates strings in hostname format.
type ZodHostname[T StringConstraint] struct{ *ZodString[T] }

func newHostname[T StringConstraint](s *ZodString[T]) *ZodHostname[T] { return &ZodHostname[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodHostname[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodHostname[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodHostname[T]) Optional() *ZodHostname[*string] { return newHostname(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodHostname[T]) Nilable() *ZodHostname[*string] { return newHostname(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodHostname[T]) Nullish() *ZodHostname[*string] { return newHostname(z.ZodString.Nullish()) }

// Hostname creates a hostname validation schema.
func Hostname(params ...any) *ZodHostname[string] {
	return newHostname(newNetworkSchema(StringTyped[string](params...), checks.Hostname()))
}

// HostnamePtr creates a pointer hostname validation schema.
func HostnamePtr(params ...any) *ZodHostname[*string] {
	return newHostname(newNetworkSchema(StringPtr(params...), checks.Hostname()))
}

// =============================================================================
// MAC
// =============================================================================

// ZodMAC validates strings in MAC address format.
type ZodMAC[T StringConstraint] struct{ *ZodString[T] }

func newMAC[T StringConstraint](s *ZodString[T]) *ZodMAC[T] { return &ZodMAC[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodMAC[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodMAC[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodMAC[T]) Optional() *ZodMAC[*string] { return newMAC(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodMAC[T]) Nilable() *ZodMAC[*string] { return newMAC(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodMAC[T]) Nullish() *ZodMAC[*string] { return newMAC(z.ZodString.Nullish()) }

// MAC creates a MAC address validation schema.
// Accepts optional delimiter string as first parameter (default: any standard delimiter).
func MAC(params ...any) *ZodMAC[string] {
	return MACTyped[string](params...)
}

// MACPtr creates a pointer MAC address validation schema.
// Accepts optional delimiter string as first parameter (default: any standard delimiter).
func MACPtr(params ...any) *ZodMAC[*string] {
	return MACTyped[*string](params...)
}

// MACTyped creates a MAC address validation schema with the given type constraint.
// Accepts optional delimiter string as first parameter (default: any standard delimiter).
func MACTyped[T StringConstraint](params ...any) *ZodMAC[T] {
	base := StringTyped[T](params...)

	var check core.ZodCheck
	if len(params) > 0 {
		if delimiter, ok := params[0].(string); ok {
			check = checks.MACWithDelimiter(delimiter)
		}
	}
	if check == nil {
		check = checks.MAC()
	}

	return newMAC(newNetworkSchema(base, check))
}

// =============================================================================
// E164
// =============================================================================

// ZodE164 validates strings in E.164 phone number format.
type ZodE164[T StringConstraint] struct{ *ZodString[T] }

func newE164[T StringConstraint](s *ZodString[T]) *ZodE164[T] { return &ZodE164[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodE164[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodE164[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodE164[T]) Optional() *ZodE164[*string] { return newE164(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodE164[T]) Nilable() *ZodE164[*string] { return newE164(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodE164[T]) Nullish() *ZodE164[*string] { return newE164(z.ZodString.Nullish()) }

// E164 creates an E.164 phone number validation schema.
func E164(params ...any) *ZodE164[string] {
	return newE164(newNetworkSchema(StringTyped[string](params...), checks.E164()))
}

// E164Ptr creates a pointer E.164 phone number validation schema.
func E164Ptr(params ...any) *ZodE164[*string] {
	return newE164(newNetworkSchema(StringPtr(params...), checks.E164()))
}

// =============================================================================
// COERCED VARIANTS
// =============================================================================

// CoercedIPv4 creates a coerced IPv4 schema that attempts string conversion.
func CoercedIPv4(params ...any) *ZodIPv4[string] {
	schema := IPv4(params...)
	schema.Internals().Coerce = true
	return schema
}

// CoercedIPv6 creates a coerced IPv6 schema that attempts string conversion.
func CoercedIPv6(params ...any) *ZodIPv6[string] {
	schema := IPv6(params...)
	schema.Internals().Coerce = true
	return schema
}

// CoercedCIDRv4 creates a coerced CIDRv4 schema that attempts string conversion.
func CoercedCIDRv4(params ...any) *ZodCIDRv4[string] {
	schema := CIDRv4(params...)
	schema.Internals().Coerce = true
	return schema
}

// CoercedCIDRv6 creates a coerced CIDRv6 schema that attempts string conversion.
func CoercedCIDRv6(params ...any) *ZodCIDRv6[string] {
	schema := CIDRv6(params...)
	schema.Internals().Coerce = true
	return schema
}

// CoercedURL creates a coerced URL schema that attempts string conversion.
func CoercedURL(params ...any) *ZodURL[string] {
	schema := URL(params...)
	schema.Internals().Coerce = true
	return schema
}

// CoercedHostname creates a coerced hostname schema that attempts string conversion.
func CoercedHostname(params ...any) *ZodHostname[string] {
	schema := Hostname(params...)
	schema.Internals().Coerce = true
	return schema
}

// CoercedMAC creates a coerced MAC schema that attempts string conversion.
func CoercedMAC(params ...any) *ZodMAC[string] {
	schema := MAC(params...)
	schema.Internals().Coerce = true
	return schema
}

// CoercedE164 creates a coerced E164 schema that attempts string conversion.
func CoercedE164(params ...any) *ZodE164[string] {
	schema := E164(params...)
	schema.Internals().Coerce = true
	return schema
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// extractNetworkString extracts string value from generic network type T.
func extractNetworkString[T StringConstraint](value T) string {
	if ptr, ok := any(value).(*string); ok {
		if ptr != nil {
			return *ptr
		}
		return ""
	}
	return any(value).(string)
}

// MACWithDelimiter creates a MAC address validation schema with a specific delimiter.
func MACWithDelimiter(delimiter string, params ...any) *ZodMAC[string] {
	base := StringTyped[string](params...)
	check := checks.MACWithDelimiter(delimiter)
	return newMAC(newNetworkSchema(base, check))
}
