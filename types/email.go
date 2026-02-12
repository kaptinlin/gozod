package types

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
)

// EmailConstraint restricts values to string or *string.
type EmailConstraint interface {
	string | *string
}

// ZodEmail validates strings as email addresses.
// String modifiers (Min, Max, Regex, etc.) and Pipe/Transform are promoted
// from the embedded *ZodString[T].
type ZodEmail[T EmailConstraint] struct {
	*ZodString[T]
}

func newEmail[T EmailConstraint](s *ZodString[T]) *ZodEmail[T] {
	return &ZodEmail[T]{s}
}

// StrictParse validates input with compile-time type safety.
func (z *ZodEmail[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on error.
func (z *ZodEmail[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Optional returns a new schema that accepts nil values.
func (z *ZodEmail[T]) Optional() *ZodEmail[*string] {
	return newEmail(z.ZodString.Optional())
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodEmail[T]) Nilable() *ZodEmail[*string] {
	return newEmail(z.ZodString.Nilable())
}

// Nullish returns a new schema combining optional and nilable.
func (z *ZodEmail[T]) Nullish() *ZodEmail[*string] {
	return newEmail(z.ZodString.Nullish())
}

// HTML5 switches to the HTML5 email pattern.
func (z *ZodEmail[T]) HTML5(params ...any) *ZodEmail[T] {
	return z.withEmailPattern(checks.HTML5Email(params...))
}

// RFC5322 switches to the RFC 5322 email pattern.
func (z *ZodEmail[T]) RFC5322(params ...any) *ZodEmail[T] {
	return z.withEmailPattern(checks.RFC5322Email(params...))
}

// Unicode switches to the Unicode-aware email pattern.
func (z *ZodEmail[T]) Unicode(params ...any) *ZodEmail[T] {
	return z.withEmailPattern(checks.UnicodeEmail(params...))
}

// Browser switches to the browser-compatible email pattern.
func (z *ZodEmail[T]) Browser(params ...any) *ZodEmail[T] {
	return z.withEmailPattern(checks.BrowserEmail(params...))
}

// Email creates a string email validation schema.
func Email(params ...any) *ZodEmail[string] {
	return EmailTyped[string](params...)
}

// EmailPtr creates a *string email validation schema.
func EmailPtr(params ...any) *ZodEmail[*string] {
	return EmailTyped[*string](params...)
}

// EmailTyped creates an email validation schema with the given type constraint.
func EmailTyped[T EmailConstraint](params ...any) *ZodEmail[T] {
	base := StringTyped[T](params...)

	var check core.ZodCheck
	if len(params) > 0 {
		if pattern, ok := params[0].(*regexp.Regexp); ok {
			check = checks.EmailWithPattern(pattern, params[1:]...)
		}
	}
	if check == nil {
		check = checks.Email(params...)
	}

	internals := base.Internals().Clone()
	internals.AddCheck(check)
	return newEmail(base.withInternals(internals))
}

// removeEmailChecks returns cs with all "email" checks filtered out.
func removeEmailChecks(cs []core.ZodCheck) []core.ZodCheck {
	out := make([]core.ZodCheck, 0, len(cs))
	for _, c := range cs {
		if inst := c.Zod(); inst != nil && inst.Def != nil && inst.Def.Check == "email" {
			continue
		}
		out = append(out, c)
	}
	return out
}

// withEmailPattern replaces existing email checks with check.
func (z *ZodEmail[T]) withEmailPattern(check core.ZodCheck) *ZodEmail[T] {
	internals := z.Internals().Clone()
	internals.Checks = removeEmailChecks(internals.Checks)
	internals.AddCheck(check)
	return newEmail(z.withInternals(internals))
}
