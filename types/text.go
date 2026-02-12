package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/pkg/regex"
)

// =============================================================================
// INTERNAL HELPERS (DRY)
// =============================================================================

// newTextSchema creates a text schema by adding a check to a base string schema.
// This eliminates repeated clone-addCheck-wrap boilerplate across all text types.
func newTextSchema[T StringConstraint](base *ZodString[T], check core.ZodCheck) *ZodString[T] {
	in := base.Internals().Clone()
	in.AddCheck(check)
	return base.withInternals(in)
}

// =============================================================================
// Emoji
// =============================================================================

// ZodEmoji validates strings containing only emoji characters.
type ZodEmoji[T StringConstraint] struct{ *ZodString[T] }

func newEmoji[T StringConstraint](s *ZodString[T]) *ZodEmoji[T] { return &ZodEmoji[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodEmoji[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodEmoji[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodEmoji[T]) Optional() *ZodEmoji[*string] { return newEmoji(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodEmoji[T]) Nilable() *ZodEmoji[*string] { return newEmoji(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodEmoji[T]) Nullish() *ZodEmoji[*string] { return newEmoji(z.ZodString.Nullish()) }

// Emoji creates an emoji validation schema.
func Emoji(params ...any) *ZodEmoji[string] {
	return newEmoji(StringTyped[string](params...).Regex(regex.Emoji))
}

// EmojiPtr creates a pointer emoji validation schema.
func EmojiPtr(params ...any) *ZodEmoji[*string] {
	return newEmoji(StringPtr(params...).Regex(regex.Emoji))
}

// =============================================================================
// JWT
// =============================================================================

// JWTOptions configures JWT validation behavior.
type JWTOptions struct {
	// Algorithm specifies the expected signing algorithm.
	// If empty, any algorithm is accepted (basic structure validation only).
	Algorithm string
}

// ZodJWT validates strings in JWT (JSON Web Token) format.
type ZodJWT[T StringConstraint] struct{ *ZodString[T] }

func newJWT[T StringConstraint](s *ZodString[T]) *ZodJWT[T] { return &ZodJWT[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodJWT[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodJWT[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodJWT[T]) Optional() *ZodJWT[*string] { return newJWT(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodJWT[T]) Nilable() *ZodJWT[*string] { return newJWT(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodJWT[T]) Nullish() *ZodJWT[*string] { return newJWT(z.ZodString.Nullish()) }

// JWT creates a JWT token validation schema.
//
//	JWT()                                                    - basic structure validation
//	JWT("error message")                                     - with custom error message
//	JWT(JWTOptions{Algorithm: "HS256"})                      - with algorithm constraint
//	JWT(JWTOptions{Algorithm: "HS256"}, "error message")     - options with error message
func JWT(params ...any) *ZodJWT[string] {
	return JWTTyped[string](params...)
}

// JWTPtr creates a pointer JWT token validation schema.
func JWTPtr(params ...any) *ZodJWT[*string] {
	return JWTTyped[*string](params...)
}

// JWTTyped creates a JWT token validation schema for a specific type.
func JWTTyped[T StringConstraint](params ...any) *ZodJWT[T] {
	options, forwarded := extractJWTOptions(params)
	base := StringTyped[T](forwarded...)

	var jwtCheck core.ZodCheck
	if options != nil && options.Algorithm != "" {
		jwtCheck = checks.JWTWithAlgorithm(options.Algorithm, forwarded...)
	} else {
		jwtCheck = checks.JWT(forwarded...)
	}

	return newJWT(newTextSchema(base, jwtCheck))
}

// extractJWTOptions separates JWTOptions from other parameters.
func extractJWTOptions(params []any) (*JWTOptions, []any) {
	var options *JWTOptions
	forwarded := make([]any, 0, len(params))

	for _, p := range params {
		if opt, ok := p.(JWTOptions); ok {
			options = &opt
		} else {
			forwarded = append(forwarded, p)
		}
	}

	return options, forwarded
}

// =============================================================================
// Base64
// =============================================================================

// ZodBase64 validates Base64 encoded strings.
type ZodBase64[T StringConstraint] struct{ *ZodString[T] }

func newBase64[T StringConstraint](s *ZodString[T]) *ZodBase64[T] { return &ZodBase64[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodBase64[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodBase64[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodBase64[T]) Optional() *ZodBase64[*string] { return newBase64(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodBase64[T]) Nilable() *ZodBase64[*string] { return newBase64(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodBase64[T]) Nullish() *ZodBase64[*string] { return newBase64(z.ZodString.Nullish()) }

// Base64 creates a Base64 encoded string validation schema.
func Base64(params ...any) *ZodBase64[string] {
	return newBase64(newTextSchema(StringTyped[string](params...), checks.Base64(params...)))
}

// Base64Ptr creates a pointer Base64 encoded string validation schema.
func Base64Ptr(params ...any) *ZodBase64[*string] {
	return newBase64(newTextSchema(StringPtr(params...), checks.Base64(params...)))
}

// =============================================================================
// Base64URL
// =============================================================================

// ZodBase64URL validates Base64URL encoded strings.
type ZodBase64URL[T StringConstraint] struct{ *ZodString[T] }

func newBase64URL[T StringConstraint](s *ZodString[T]) *ZodBase64URL[T] {
	return &ZodBase64URL[T]{s}
}

// StrictParse validates the input with compile-time type safety.
func (z *ZodBase64URL[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodBase64URL[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodBase64URL[T]) Optional() *ZodBase64URL[*string] {
	return newBase64URL(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodBase64URL[T]) Nilable() *ZodBase64URL[*string] {
	return newBase64URL(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodBase64URL[T]) Nullish() *ZodBase64URL[*string] {
	return newBase64URL(z.ZodString.Nullish())
}

// Base64URL creates a Base64URL encoded string validation schema.
func Base64URL(params ...any) *ZodBase64URL[string] {
	return newBase64URL(newTextSchema(StringTyped[string](params...), checks.Base64URL(params...)))
}

// Base64URLPtr creates a pointer Base64URL encoded string validation schema.
func Base64URLPtr(params ...any) *ZodBase64URL[*string] {
	return newBase64URL(newTextSchema(StringPtr(params...), checks.Base64URL(params...)))
}

// =============================================================================
// Hex
// =============================================================================

// ZodHex validates hexadecimal strings.
type ZodHex[T StringConstraint] struct{ *ZodString[T] }

func newHex[T StringConstraint](s *ZodString[T]) *ZodHex[T] { return &ZodHex[T]{s} }

// StrictParse validates the input with compile-time type safety.
func (z *ZodHex[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input with compile-time type safety and panics on error.
func (z *ZodHex[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodHex[T]) Optional() *ZodHex[*string] { return newHex(z.ZodString.Optional()) }

// Nilable returns a new schema that accepts nil values.
func (z *ZodHex[T]) Nilable() *ZodHex[*string] { return newHex(z.ZodString.Nilable()) }

// Nullish returns a new schema combining optional and nilable.
func (z *ZodHex[T]) Nullish() *ZodHex[*string] { return newHex(z.ZodString.Nullish()) }

// Hex creates a hexadecimal string validation schema.
// Zod v4 equivalent: z.hex()
func Hex(params ...any) *ZodHex[string] {
	return newHex(newTextSchema(StringTyped[string](params...), checks.Hex(params...)))
}

// HexPtr creates a pointer hexadecimal string validation schema.
func HexPtr(params ...any) *ZodHex[*string] {
	return newHex(newTextSchema(StringPtr(params...), checks.Hex(params...)))
}
