package transform

import (
	"strings"
	"unicode"
)

// Slugify converts a string to a URL-friendly slug.
//
// The transformation matches Zod v4's slugify behavior:
//  1. Lowercase the input
//  2. Trim whitespace
//  3. Remove non-word, non-space, non-hyphen characters
//  4. Collapse consecutive spaces, underscores, and hyphens into a single hyphen
//  5. Trim leading and trailing hyphens
func Slugify(input string) string {
	if input == "" {
		return ""
	}

	lowered := strings.ToLower(strings.TrimSpace(input))
	if lowered == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(lowered))

	prevHyphen := false
	for _, r := range lowered {
		switch {
		case isWordChar(r):
			prevHyphen = false
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-' || r == '\t' || r == '\n' || r == '\r':
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}

	return strings.TrimRight(b.String(), "-")
}

// isWordChar reports whether r matches \w (letters, digits, underscore)
// excluding underscore, since underscores are treated as separators.
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
