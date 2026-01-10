package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/pkg/regex"
)

// =============================================================================
// Emoji
// =============================================================================

type ZodEmoji[T StringConstraint] struct{ *ZodString[T] }

// StrictParse validates the input using strict parsing rules
func (z *ZodEmoji[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodEmoji[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

func Emoji(params ...any) *ZodEmoji[string] {
	base := StringTyped[string](params...).Regex(regex.Emoji)
	return &ZodEmoji[string]{base}
}

func EmojiPtr(params ...any) *ZodEmoji[*string] {
	base := StringPtr(params...).Regex(regex.Emoji)
	return &ZodEmoji[*string]{base}
}

// =============================================================================
// JWT
// =============================================================================

// JWTOptions defines options for JWT validation
type JWTOptions struct {
	// Algorithm specifies the expected signing algorithm
	// If empty, any algorithm is accepted (basic structure validation only)
	Algorithm string
}

type ZodJWT[T StringConstraint] struct{ *ZodString[T] }

// StrictParse validates the input using strict parsing rules
func (z *ZodJWT[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodJWT[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// JWT creates a JWT token validation schema
// Supports various parameter combinations:
//
//	JWT() - basic JWT structure validation
//	JWT("error message") - basic validation with custom error message
//	JWT(JWTOptions{Algorithm: "HS256"}) - with algorithm constraint
//	JWT(options, "error message") - options with error message
//	JWT(options, core.SchemaParams{Description: "JWT"}) - options with schema params
func JWT(params ...any) *ZodJWT[string] {
	return JWTTyped[string](params...)
}

// JWTPtr creates a JWT token validation schema for pointer type
func JWTPtr(params ...any) *ZodJWT[*string] {
	return JWTTyped[*string](params...)
}

// JWTTyped creates a JWT token validation schema with specific type
func JWTTyped[T StringConstraint](params ...any) *ZodJWT[T] {
	// Extract JWTOptions (if any) and keep the remaining parameters as-is.
	var options *JWTOptions
	var forwarded []any

	for _, p := range params {
		if opt, ok := p.(JWTOptions); ok {
			options = &opt
			// do NOT include in forwarded slice
		} else {
			forwarded = append(forwarded, p)
		}
	}

	// Build base string schema – StringTyped already handles SchemaParams / error messages.
	base := StringTyped[T](forwarded...)

	// Select appropriate JWT check.
	var jwtCheck core.ZodCheck
	if options != nil && options.Algorithm != "" {
		jwtCheck = checks.JWTWithAlgorithm(options.Algorithm, forwarded...)
	} else {
		jwtCheck = checks.JWT(forwarded...)
	}

	newInternals := base.internals.Clone()
	newInternals.AddCheck(jwtCheck)

	return &ZodJWT[T]{base.withInternals(newInternals)}
}

// =============================================================================
// Base64
// =============================================================================

// ZodBase64 defines a schema for Base64 encoded strings.
type ZodBase64[T StringConstraint] struct{ *ZodString[T] }

// StrictParse validates the input using strict parsing rules
func (z *ZodBase64[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodBase64[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Base64 creates a Base64 encoded string validation schema.
// Supports custom error messages and schema parameters.
//
//	Base64()
//	Base64("custom error message")
//	Base64(core.SchemaParams{Description: "Base64 string"})
func Base64(params ...any) *ZodBase64[string] {
	return Base64Typed[string](params...)
}

// Base64Ptr creates a Base64 encoded string validation schema for a pointer type.
func Base64Ptr(params ...any) *ZodBase64[*string] {
	return Base64Typed[*string](params...)
}

// Base64Typed creates a Base64 encoded string validation schema for a specific type.
func Base64Typed[T StringConstraint](params ...any) *ZodBase64[T] {
	// Leverage StringTyped and Base64 check directly – utils.NormalizeParams is
	// already used inside StringTyped to handle SchemaParams / error strings.
	base := StringTyped[T](params...)

	// Attach Base64 format check (it will process params for custom error).
	base64Check := checks.Base64(params...)

	newInternals := base.internals.Clone()
	newInternals.AddCheck(base64Check)

	return &ZodBase64[T]{base.withInternals(newInternals)}
}

// =============================================================================
// Base64URL
// =============================================================================

// ZodBase64URL defines a schema for Base64URL encoded strings.
type ZodBase64URL[T StringConstraint] struct{ *ZodString[T] }

// StrictParse validates the input using strict parsing rules
func (z *ZodBase64URL[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodBase64URL[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Base64URL creates a Base64URL encoded string validation schema.
// Supports custom error messages and schema parameters.
//
//	Base64URL()
//	Base64URL("custom error message")
//	Base64URL(core.SchemaParams{Description: "Base64URL string"})
func Base64URL(params ...any) *ZodBase64URL[string] {
	return Base64URLTyped[string](params...)
}

// Base64URLPtr creates a Base64URL encoded string validation schema for a pointer type.
func Base64URLPtr(params ...any) *ZodBase64URL[*string] {
	return Base64URLTyped[*string](params...)
}

// Base64URLTyped creates a Base64URL encoded string validation schema for a specific type.
func Base64URLTyped[T StringConstraint](params ...any) *ZodBase64URL[T] {
	base := StringTyped[T](params...)

	// Attach Base64URL format check (handles custom error via params).
	base64URLCheck := checks.Base64URL(params...)

	newInternals := base.internals.Clone()
	newInternals.AddCheck(base64URLCheck)

	return &ZodBase64URL[T]{base.withInternals(newInternals)}
}

// =============================================================================
// Hex
// =============================================================================

// ZodHex represents a hexadecimal string validation schema
type ZodHex[T StringConstraint] struct{ *ZodString[T] }

// StrictParse validates the input using strict parsing rules
func (z *ZodHex[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodHex[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Hex creates a hexadecimal string validation schema
// TypeScript Zod v4 equivalent: z.hex()
//
// Examples:
//
//	schema := Hex()
//	schema.Parse("")           // valid (empty string is valid hex)
//	schema.Parse("123abc")     // valid
//	schema.Parse("DEADBEEF")   // valid
//	schema.Parse("xyz")        // invalid
func Hex(params ...any) *ZodHex[string] {
	return HexTyped[string](params...)
}

// HexPtr creates a hexadecimal string validation schema for pointer types
func HexPtr(params ...any) *ZodHex[*string] {
	return HexTyped[*string](params...)
}

// HexTyped creates a hexadecimal string validation schema for a specific type
func HexTyped[T StringConstraint](params ...any) *ZodHex[T] {
	base := StringTyped[T](params...)

	// Attach Hex format check
	hexCheck := checks.Hex(params...)

	newInternals := base.internals.Clone()
	newInternals.AddCheck(hexCheck)

	return &ZodHex[T]{base.withInternals(newInternals)}
}
