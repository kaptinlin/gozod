package types

import "github.com/kaptinlin/gozod/pkg/regexes"

// =============================================================================
// CUID
// =============================================================================

type ZodCUID[T StringConstraint] struct{ *ZodString[T] }

func Cuid(params ...any) *ZodCUID[string] {
	base := StringTyped[string](params...).Regex(regexes.CUID)
	return &ZodCUID[string]{base}
}

func CuidPtr(params ...any) *ZodCUID[*string] {
	base := StringPtr(params...).Regex(regexes.CUID)
	return &ZodCUID[*string]{base}
}

// =============================================================================
// CUID2
// =============================================================================

type ZodCUID2[T StringConstraint] struct{ *ZodString[T] }

func Cuid2(params ...any) *ZodCUID2[string] {
	base := StringTyped[string](params...).Regex(regexes.CUID2)
	return &ZodCUID2[string]{base}
}

func Cuid2Ptr(params ...any) *ZodCUID2[*string] {
	base := StringPtr(params...).Regex(regexes.CUID2)
	return &ZodCUID2[*string]{base}
}

// =============================================================================
// ULID
// =============================================================================

type ZodULID[T StringConstraint] struct{ *ZodString[T] }

func Ulid(params ...any) *ZodULID[string] {
	base := StringTyped[string](params...).Regex(regexes.ULID)
	return &ZodULID[string]{base}
}

func UlidPtr(params ...any) *ZodULID[*string] {
	base := StringPtr(params...).Regex(regexes.ULID)
	return &ZodULID[*string]{base}
}

// =============================================================================
// XID
// =============================================================================

type ZodXID[T StringConstraint] struct{ *ZodString[T] }

func Xid(params ...any) *ZodXID[string] {
	base := StringTyped[string](params...).Regex(regexes.XID)
	return &ZodXID[string]{base}
}

func XidPtr(params ...any) *ZodXID[*string] {
	base := StringPtr(params...).Regex(regexes.XID)
	return &ZodXID[*string]{base}
}

// =============================================================================
// KSUID
// =============================================================================

type ZodKSUID[T StringConstraint] struct{ *ZodString[T] }

func Ksuid(params ...any) *ZodKSUID[string] {
	base := StringTyped[string](params...).Regex(regexes.KSUID)
	return &ZodKSUID[string]{base}
}

func KsuidPtr(params ...any) *ZodKSUID[*string] {
	base := StringPtr(params...).Regex(regexes.KSUID)
	return &ZodKSUID[*string]{base}
}

// =============================================================================
// NanoID
// =============================================================================

type ZodNanoID[T StringConstraint] struct{ *ZodString[T] }

func Nanoid(params ...any) *ZodNanoID[string] {
	base := StringTyped[string](params...).Regex(regexes.NanoID)
	return &ZodNanoID[string]{base}
}

func NanoidPtr(params ...any) *ZodNanoID[*string] {
	base := StringPtr(params...).Regex(regexes.NanoID)
	return &ZodNanoID[*string]{base}
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
		base = StringTyped[string](rest...).Regex(regexes.UUID4)
	case "6":
		base = StringTyped[string](rest...).Regex(regexes.UUID6)
	case "7":
		base = StringTyped[string](rest...).Regex(regexes.UUID7)
	default:
		base = StringTyped[string](params...).Regex(regexes.UUID)
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
		base = StringPtr(rest...).Regex(regexes.UUID4)
	case "6":
		base = StringPtr(rest...).Regex(regexes.UUID6)
	case "7":
		base = StringPtr(rest...).Regex(regexes.UUID7)
	default:
		base = StringPtr(params...).Regex(regexes.UUID)
	}
	return &ZodUUID[*string]{base}
}

// =============================================================================
// UUID version convenience constructors
// =============================================================================

func Uuidv4(params ...any) *ZodUUID[string] {
	base := StringTyped[string](params...).Regex(regexes.UUID4)
	return &ZodUUID[string]{base}
}

func Uuidv4Ptr(params ...any) *ZodUUID[*string] {
	base := StringPtr(params...).Regex(regexes.UUID4)
	return &ZodUUID[*string]{base}
}

func Uuidv6(params ...any) *ZodUUID[string] {
	base := StringTyped[string](params...).Regex(regexes.UUID6)
	return &ZodUUID[string]{base}
}

func Uuidv6Ptr(params ...any) *ZodUUID[*string] {
	base := StringPtr(params...).Regex(regexes.UUID6)
	return &ZodUUID[*string]{base}
}

func Uuidv7(params ...any) *ZodUUID[string] {
	base := StringTyped[string](params...).Regex(regexes.UUID7)
	return &ZodUUID[string]{base}
}

func Uuidv7Ptr(params ...any) *ZodUUID[*string] {
	base := StringPtr(params...).Regex(regexes.UUID7)
	return &ZodUUID[*string]{base}
}
