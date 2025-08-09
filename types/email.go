package types

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
)

// =============================================================================
// Email schema built on top of ZodString
// =============================================================================
// This refactor delegates all generic string behaviour to ZodString via type
// embedding and only keeps e-mail specific helpers.
// =============================================================================

type EmailConstraint interface {
	string | *string
}

// ZodEmail is a thin wrapper around ZodString.
// All generic string modifiers (Min/Max/Regex/…​) plus Pipe/Transform are
// promoted automatically from the embedded *ZodString[T].

type ZodEmail[T EmailConstraint] struct{ *ZodString[T] }

// -----------------------------------------------------------------------------
// internal helpers
// -----------------------------------------------------------------------------

// newFromString wraps an existing *ZodString[T] with ZodEmail.
func newFromString[T EmailConstraint](str *ZodString[T]) *ZodEmail[T] {
	return &ZodEmail[T]{str}
}

// removeEmailChecks filters out existing checks with def.Check == "email".
func removeEmailChecks(checksSlice []core.ZodCheck) []core.ZodCheck {
	filtered := make([]core.ZodCheck, 0, len(checksSlice))
	for _, c := range checksSlice {
		if inst := c.GetZod(); inst != nil && inst.Def != nil && inst.Def.Check == "email" {
			// skip
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

// withEmailPattern replaces any existing email checks with the supplied one.
func (z *ZodEmail[T]) withEmailPattern(patternCheck core.ZodCheck) *ZodEmail[T] {
	cloned := z.GetInternals().Clone()
	cloned.Checks = removeEmailChecks(cloned.Checks)
	cloned.AddCheck(patternCheck)
	return newFromString(z.ZodString.withInternals(cloned))
}

// -----------------------------------------------------------------------------
// modifiers that must return *ZodEmail
// -----------------------------------------------------------------------------

func (z *ZodEmail[T]) Optional() *ZodEmail[*string] { return newFromString(z.ZodString.Optional()) }
func (z *ZodEmail[T]) Nilable() *ZodEmail[*string]  { return newFromString(z.ZodString.Nilable()) }
func (z *ZodEmail[T]) Nullish() *ZodEmail[*string]  { return newFromString(z.ZodString.Nullish()) }

// -----------------------------------------------------------------------------
// email-specific helpers (override default pattern)
// -----------------------------------------------------------------------------

func (z *ZodEmail[T]) Html5(params ...any) *ZodEmail[T] {
	return z.withEmailPattern(checks.Html5Email(params...))
}
func (z *ZodEmail[T]) Rfc5322(params ...any) *ZodEmail[T] {
	return z.withEmailPattern(checks.Rfc5322Email(params...))
}
func (z *ZodEmail[T]) Unicode(params ...any) *ZodEmail[T] {
	return z.withEmailPattern(checks.UnicodeEmail(params...))
}
func (z *ZodEmail[T]) Browser(params ...any) *ZodEmail[T] {
	return z.withEmailPattern(checks.BrowserEmail(params...))
}

// -----------------------------------------------------------------------------
// constructors
// -----------------------------------------------------------------------------

func Email(params ...any) *ZodEmail[string]     { return EmailTyped[string](params...) }
func EmailPtr(params ...any) *ZodEmail[*string] { return EmailTyped[*string](params...) }

// EmailTyped creates an email validation schema with specific type
func EmailTyped[T EmailConstraint](params ...any) *ZodEmail[T] {
	// Leverage StringTyped first – utils.NormalizeParams will already process
	// SchemaParams / error strings.
	base := StringTyped[T](params...)

	var emailCheck core.ZodCheck

	// Support custom pattern as first param: Email(customRegexp, ...)
	if len(params) > 0 {
		if pattern, ok := params[0].(*regexp.Regexp); ok {
			// Remove the pattern from params when forwarding the rest as error
			remaining := params[1:]
			emailCheck = checks.EmailWithPattern(pattern, remaining...)
		}
	}

	// Fallback to default email check if none set above
	if emailCheck == nil {
		emailCheck = checks.Email(params...)
	}

	newInternals := base.GetInternals().Clone()
	newInternals.AddCheck(emailCheck)

	return &ZodEmail[T]{base.withInternals(newInternals)}
}

// StrictParse validates the input using strict parsing rules
func (z *ZodEmail[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return z.ZodString.StrictParse(input, ctx...)
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodEmail[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// GetInternals proxies to the embedded string schema.
func (z *ZodEmail[T]) GetInternals() *core.ZodTypeInternals { return z.ZodString.GetInternals() }
