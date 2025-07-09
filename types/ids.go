package types

import (
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/pkg/regexes"
)

// =============================================================================
// CUID
// =============================================================================

type ZodCUID[T StringConstraint] struct{ *ZodString[T] }

func Cuid(params ...any) *ZodCUID[string] {
	base := StringTyped[string](params...)
	check := checks.CUID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodCUID[string]{base.withInternals(in)}
}

func CuidPtr(params ...any) *ZodCUID[*string] {
	base := StringPtr(params...)
	check := checks.CUID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodCUID[*string]{base.withInternals(in)}
}

// =============================================================================
// CUID2
// =============================================================================

type ZodCUID2[T StringConstraint] struct{ *ZodString[T] }

func Cuid2(params ...any) *ZodCUID2[string] {
	base := StringTyped[string](params...)
	check := checks.CUID2(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodCUID2[string]{base.withInternals(in)}
}

func Cuid2Ptr(params ...any) *ZodCUID2[*string] {
	base := StringPtr(params...)
	check := checks.CUID2(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodCUID2[*string]{base.withInternals(in)}
}

// =============================================================================
// ULID
// =============================================================================

type ZodULID[T StringConstraint] struct{ *ZodString[T] }

func Ulid(params ...any) *ZodULID[string] {
	base := StringTyped[string](params...)
	check := checks.ULID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodULID[string]{base.withInternals(in)}
}

func UlidPtr(params ...any) *ZodULID[*string] {
	base := StringPtr(params...)
	check := checks.ULID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodULID[*string]{base.withInternals(in)}
}

// =============================================================================
// XID
// =============================================================================

type ZodXID[T StringConstraint] struct{ *ZodString[T] }

func Xid(params ...any) *ZodXID[string] {
	base := StringTyped[string](params...)
	check := checks.XID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodXID[string]{base.withInternals(in)}
}

func XidPtr(params ...any) *ZodXID[*string] {
	base := StringPtr(params...)
	check := checks.XID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodXID[*string]{base.withInternals(in)}
}

// =============================================================================
// KSUID
// =============================================================================

type ZodKSUID[T StringConstraint] struct{ *ZodString[T] }

func Ksuid(params ...any) *ZodKSUID[string] {
	base := StringTyped[string](params...)
	check := checks.KSUID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodKSUID[string]{base.withInternals(in)}
}

func KsuidPtr(params ...any) *ZodKSUID[*string] {
	base := StringPtr(params...)
	check := checks.KSUID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodKSUID[*string]{base.withInternals(in)}
}

// =============================================================================
// NanoID
// =============================================================================

type ZodNanoID[T StringConstraint] struct{ *ZodString[T] }

func Nanoid(params ...any) *ZodNanoID[string] {
	base := StringTyped[string](params...)
	check := checks.NanoID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodNanoID[string]{base.withInternals(in)}
}

func NanoidPtr(params ...any) *ZodNanoID[*string] {
	base := StringPtr(params...)
	check := checks.NanoID(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodNanoID[*string]{base.withInternals(in)}
}

// =============================================================================
// UUID
// =============================================================================

type ZodUUID[T StringConstraint] struct{ *ZodString[T] }

// Uuid supports optional version parameter: "v4", "v6", "v7".
// Otherwise behaves like generic UUID validator.
func Uuid(params ...any) *ZodUUID[string] {
	var version string
	var rest []any

	if len(params) > 0 {
		if v, ok := params[0].(string); ok {
			switch v {
			case "v4":
				version = "4"
			case "v6":
				version = "6"
			case "v7":
				version = "7"
			}
			if version != "" {
				rest = params[1:]
			} else {
				rest = params
			}
		} else {
			rest = params
		}
	}

	var base *ZodString[string]
	switch version {
	case "4":
		base = StringTyped[string](rest...)
		in := base.GetInternals().Clone()
		in.AddCheck(checks.UUID(rest...))
		base = base.withInternals(in)
		base = base.Regex(regexes.UUID4)
	case "6":
		base = StringTyped[string](rest...)
		in := base.GetInternals().Clone()
		in.AddCheck(checks.UUID(rest...))
		base = base.withInternals(in)
		base = base.Regex(regexes.UUID6)
	case "7":
		base = StringTyped[string](rest...)
		in := base.GetInternals().Clone()
		in.AddCheck(checks.UUID(rest...))
		base = base.withInternals(in)
		base = base.Regex(regexes.UUID7)
	default:
		base = StringTyped[string](params...)
		in := base.GetInternals().Clone()
		in.AddCheck(checks.UUID(params...))
		base = base.withInternals(in)
		base = base.Regex(regexes.UUID)
	}
	return &ZodUUID[string]{base}
}

func UuidPtr(params ...any) *ZodUUID[*string] {
	var version string
	var rest []any

	if len(params) > 0 {
		if v, ok := params[0].(string); ok {
			switch v {
			case "v4":
				version = "4"
			case "v6":
				version = "6"
			case "v7":
				version = "7"
			}
			if version != "" {
				rest = params[1:]
			} else {
				rest = params
			}
		} else {
			rest = params
		}
	}

	var base *ZodString[*string]
	switch version {
	case "4":
		base = StringPtr(rest...)
		in := base.GetInternals().Clone()
		in.AddCheck(checks.UUID(rest...))
		base = base.withInternals(in)
		base = base.Regex(regexes.UUID4)
	case "6":
		base = StringPtr(rest...)
		in := base.GetInternals().Clone()
		in.AddCheck(checks.UUID(rest...))
		base = base.withInternals(in)
		base = base.Regex(regexes.UUID6)
	case "7":
		base = StringPtr(rest...)
		in := base.GetInternals().Clone()
		in.AddCheck(checks.UUID(rest...))
		base = base.withInternals(in)
		base = base.Regex(regexes.UUID7)
	default:
		base = StringPtr(params...)
		in := base.GetInternals().Clone()
		in.AddCheck(checks.UUID(params...))
		base = base.withInternals(in)
		base = base.Regex(regexes.UUID)
	}
	return &ZodUUID[*string]{base}
}

// =============================================================================
// UUID version convenience constructors
// =============================================================================

func Uuidv4(params ...any) *ZodUUID[string] {
	base := StringTyped[string](params...)
	check := checks.UUIDv4(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodUUID[string]{base.withInternals(in)}
}

func Uuidv4Ptr(params ...any) *ZodUUID[*string] {
	base := StringPtr(params...)
	check := checks.UUIDv4(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodUUID[*string]{base.withInternals(in)}
}

func Uuidv6(params ...any) *ZodUUID[string] {
	base := StringTyped[string](params...)
	check := checks.UUID6(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodUUID[string]{base.withInternals(in)}
}

func Uuidv6Ptr(params ...any) *ZodUUID[*string] {
	base := StringPtr(params...)
	check := checks.UUID6(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodUUID[*string]{base.withInternals(in)}
}

func Uuidv7(params ...any) *ZodUUID[string] {
	base := StringTyped[string](params...)
	check := checks.UUID7(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodUUID[string]{base.withInternals(in)}
}

func Uuidv7Ptr(params ...any) *ZodUUID[*string] {
	base := StringPtr(params...)
	check := checks.UUID7(params...)
	in := base.GetInternals().Clone()
	in.AddCheck(check)
	return &ZodUUID[*string]{base.withInternals(in)}
}
