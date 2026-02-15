package types

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/pkg/regex"
)

// newIDSchema adds a check to a base string schema and returns a new instance.
// This is an unexported helper function for creating ID validation schemas.
func newIDSchema[T StringConstraint](base *ZodString[T], check core.ZodCheck) *ZodString[T] {
	in := base.Internals().Clone()
	in.AddCheck(check)
	return base.withInternals(in)
}

// ZodGUID validates strings in GUID format (8-4-4-4-12 hex pattern).
type ZodGUID[T StringConstraint] struct{ *ZodString[T] }

// newGUID creates a new ZodGUID wrapper around a ZodString.
// This is an unexported helper function for internal use.
func newGUID[T StringConstraint](s *ZodString[T]) *ZodGUID[T] {
	return &ZodGUID[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodGUID[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodGUID[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodGUID[T]) Optional() *ZodGUID[*string] {
	return newGUID(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodGUID[T]) Nilable() *ZodGUID[*string] {
	return newGUID(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodGUID[T]) Nullish() *ZodGUID[*string] {
	return newGUID(z.ZodString.Nullish())
}

// GUID creates a GUID schema (8-4-4-4-12 hex pattern).
func GUID(params ...any) *ZodGUID[string] {
	return newGUID(newIDSchema(StringTyped[string](params...), checks.GUID(params...)))
}

// GUIDPtr creates a pointer GUID schema.
func GUIDPtr(params ...any) *ZodGUID[*string] {
	return newGUID(newIDSchema(StringPtr(params...), checks.GUID(params...)))
}

// ZodCUID validates strings in CUID format.
type ZodCUID[T StringConstraint] struct{ *ZodString[T] }

func newCUID[T StringConstraint](s *ZodString[T]) *ZodCUID[T] {
	return &ZodCUID[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodCUID[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodCUID[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodCUID[T]) Optional() *ZodCUID[*string] {
	return newCUID(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodCUID[T]) Nilable() *ZodCUID[*string] {
	return newCUID(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodCUID[T]) Nullish() *ZodCUID[*string] {
	return newCUID(z.ZodString.Nullish())
}

// Cuid creates a CUID schema for collision-resistant unique identifiers.
func Cuid(params ...any) *ZodCUID[string] {
	return newCUID(newIDSchema(StringTyped[string](params...), checks.CUID(params...)))
}

// CuidPtr creates a pointer CUID schema.
func CuidPtr(params ...any) *ZodCUID[*string] {
	return newCUID(newIDSchema(StringPtr(params...), checks.CUID(params...)))
}

// ZodCUID2 validates strings in CUID2 format.
type ZodCUID2[T StringConstraint] struct{ *ZodString[T] }

func newCUID2[T StringConstraint](s *ZodString[T]) *ZodCUID2[T] {
	return &ZodCUID2[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodCUID2[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodCUID2[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodCUID2[T]) Optional() *ZodCUID2[*string] {
	return newCUID2(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodCUID2[T]) Nilable() *ZodCUID2[*string] {
	return newCUID2(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodCUID2[T]) Nullish() *ZodCUID2[*string] {
	return newCUID2(z.ZodString.Nullish())
}

// Cuid2 creates a CUID2 schema for next-generation collision-resistant identifiers.
func Cuid2(params ...any) *ZodCUID2[string] {
	return newCUID2(newIDSchema(StringTyped[string](params...), checks.CUID2(params...)))
}

// Cuid2Ptr creates a pointer CUID2 schema.
func Cuid2Ptr(params ...any) *ZodCUID2[*string] {
	return newCUID2(newIDSchema(StringPtr(params...), checks.CUID2(params...)))
}

// ZodULID validates strings in ULID format.
type ZodULID[T StringConstraint] struct{ *ZodString[T] }

func newULID[T StringConstraint](s *ZodString[T]) *ZodULID[T] {
	return &ZodULID[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodULID[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodULID[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodULID[T]) Optional() *ZodULID[*string] {
	return newULID(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodULID[T]) Nilable() *ZodULID[*string] {
	return newULID(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodULID[T]) Nullish() *ZodULID[*string] {
	return newULID(z.ZodString.Nullish())
}

// Ulid creates a ULID schema.
func Ulid(params ...any) *ZodULID[string] {
	return newULID(newIDSchema(StringTyped[string](params...), checks.ULID(params...)))
}

// UlidPtr creates a pointer ULID schema.
func UlidPtr(params ...any) *ZodULID[*string] {
	return newULID(newIDSchema(StringPtr(params...), checks.ULID(params...)))
}

// ZodXID validates strings in XID format.
type ZodXID[T StringConstraint] struct{ *ZodString[T] }

func newXID[T StringConstraint](s *ZodString[T]) *ZodXID[T] {
	return &ZodXID[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodXID[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodXID[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodXID[T]) Optional() *ZodXID[*string] {
	return newXID(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodXID[T]) Nilable() *ZodXID[*string] {
	return newXID(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodXID[T]) Nullish() *ZodXID[*string] {
	return newXID(z.ZodString.Nullish())
}

// Xid creates an XID schema.
func Xid(params ...any) *ZodXID[string] {
	return newXID(newIDSchema(StringTyped[string](params...), checks.XID(params...)))
}

// XidPtr creates a pointer XID schema.
func XidPtr(params ...any) *ZodXID[*string] {
	return newXID(newIDSchema(StringPtr(params...), checks.XID(params...)))
}

// ZodKSUID validates strings in KSUID format.
type ZodKSUID[T StringConstraint] struct{ *ZodString[T] }

func newKSUID[T StringConstraint](s *ZodString[T]) *ZodKSUID[T] {
	return &ZodKSUID[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodKSUID[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodKSUID[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodKSUID[T]) Optional() *ZodKSUID[*string] {
	return newKSUID(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodKSUID[T]) Nilable() *ZodKSUID[*string] {
	return newKSUID(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodKSUID[T]) Nullish() *ZodKSUID[*string] {
	return newKSUID(z.ZodString.Nullish())
}

// Ksuid creates a KSUID schema.
func Ksuid(params ...any) *ZodKSUID[string] {
	return newKSUID(newIDSchema(StringTyped[string](params...), checks.KSUID(params...)))
}

// KsuidPtr creates a pointer KSUID schema.
func KsuidPtr(params ...any) *ZodKSUID[*string] {
	return newKSUID(newIDSchema(StringPtr(params...), checks.KSUID(params...)))
}

// ZodNanoID validates strings in NanoID format.
type ZodNanoID[T StringConstraint] struct{ *ZodString[T] }

func newNanoID[T StringConstraint](s *ZodString[T]) *ZodNanoID[T] {
	return &ZodNanoID[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodNanoID[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodNanoID[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodNanoID[T]) Optional() *ZodNanoID[*string] {
	return newNanoID(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodNanoID[T]) Nilable() *ZodNanoID[*string] {
	return newNanoID(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodNanoID[T]) Nullish() *ZodNanoID[*string] {
	return newNanoID(z.ZodString.Nullish())
}

// Nanoid creates a NanoID schema.
func Nanoid(params ...any) *ZodNanoID[string] {
	return newNanoID(newIDSchema(StringTyped[string](params...), checks.NanoID(params...)))
}

// NanoidPtr creates a pointer NanoID schema.
func NanoidPtr(params ...any) *ZodNanoID[*string] {
	return newNanoID(newIDSchema(StringPtr(params...), checks.NanoID(params...)))
}

// ZodUUID validates strings in UUID format.
type ZodUUID[T StringConstraint] struct{ *ZodString[T] }

func newUUID[T StringConstraint](s *ZodString[T]) *ZodUUID[T] {
	return &ZodUUID[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodUUID[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodUUID[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodUUID[T]) Optional() *ZodUUID[*string] {
	return newUUID(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodUUID[T]) Nilable() *ZodUUID[*string] {
	return newUUID(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodUUID[T]) Nullish() *ZodUUID[*string] {
	return newUUID(z.ZodString.Nullish())
}

// uuidVersionRE maps version strings to their regex patterns.
var uuidVersionRE = map[string]*regexp.Regexp{
	"4": regex.UUID4,
	"6": regex.UUID6,
	"7": regex.UUID7,
}

// newUUIDSchema creates a UUID schema with an optional version-specific regex.
func newUUIDSchema[T StringConstraint](base *ZodString[T], version string, params []any) *ZodUUID[T] {
	s := newIDSchema(base, checks.UUID(params...))
	if re, ok := uuidVersionRE[version]; ok {
		s = s.Regex(re)
	} else {
		s = s.Regex(regex.UUID)
	}
	return newUUID(s)
}

// parseUUIDVersion extracts UUID version from params if present.
func parseUUIDVersion(params []any) (version string, rest []any) {
	if len(params) == 0 {
		return "", params
	}
	if v, ok := params[0].(string); ok {
		switch v {
		case "v4":
			return "4", params[1:]
		case "v6":
			return "6", params[1:]
		case "v7":
			return "7", params[1:]
		}
	}
	return "", params
}

// UUID creates a UUID schema with optional version parameter: "v4", "v6", "v7".
func UUID(params ...any) *ZodUUID[string] {
	ver, rest := parseUUIDVersion(params)
	return newUUIDSchema(StringTyped[string](rest...), ver, rest)
}

// UUIDPtr creates a pointer UUID schema with optional version parameter.
func UUIDPtr(params ...any) *ZodUUID[*string] {
	ver, rest := parseUUIDVersion(params)
	return newUUIDSchema(StringPtr(rest...), ver, rest)
}

// Uuidv4 creates a UUID v4 schema.
func Uuidv4(params ...any) *ZodUUID[string] {
	return newUUID(newIDSchema(StringTyped[string](params...), checks.UUIDv4(params...)))
}

// Uuidv4Ptr creates a pointer UUID v4 schema.
func Uuidv4Ptr(params ...any) *ZodUUID[*string] {
	return newUUID(newIDSchema(StringPtr(params...), checks.UUIDv4(params...)))
}

// Uuidv6 creates a UUID v6 schema.
func Uuidv6(params ...any) *ZodUUID[string] {
	return newUUID(newIDSchema(StringTyped[string](params...), checks.UUID6(params...)))
}

// Uuidv6Ptr creates a pointer UUID v6 schema.
func Uuidv6Ptr(params ...any) *ZodUUID[*string] {
	return newUUID(newIDSchema(StringPtr(params...), checks.UUID6(params...)))
}

// Uuidv7 creates a UUID v7 schema.
func Uuidv7(params ...any) *ZodUUID[string] {
	return newUUID(newIDSchema(StringTyped[string](params...), checks.UUID7(params...)))
}

// Uuidv7Ptr creates a pointer UUID v7 schema.
func Uuidv7Ptr(params ...any) *ZodUUID[*string] {
	return newUUID(newIDSchema(StringPtr(params...), checks.UUID7(params...)))
}
