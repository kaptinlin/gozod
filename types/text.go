package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/pkg/regexes"
)

// =============================================================================
// Emoji
// =============================================================================

type ZodEmoji[T StringConstraint] struct{ *ZodString[T] }

func Emoji(params ...any) *ZodEmoji[string] {
	base := StringTyped[string](params...).Regex(regexes.Emoji)
	return &ZodEmoji[string]{base}
}

func EmojiPtr(params ...any) *ZodEmoji[*string] {
	base := StringPtr(params...).Regex(regexes.Emoji)
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

	newInternals := base.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(jwtCheck)

	return &ZodJWT[T]{base.withInternals(newInternals)}
}

// =============================================================================
// Base64
// =============================================================================

// ZodBase64 defines a schema for Base64 encoded strings.
type ZodBase64[T StringConstraint] struct{ *ZodString[T] }

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

	newInternals := base.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(base64Check)

	return &ZodBase64[T]{base.withInternals(newInternals)}
}

// =============================================================================
// Base64URL
// =============================================================================

// ZodBase64URL defines a schema for Base64URL encoded strings.
type ZodBase64URL[T StringConstraint] struct{ *ZodString[T] }

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

	newInternals := base.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(base64URLCheck)

	return &ZodBase64URL[T]{base.withInternals(newInternals)}
}
