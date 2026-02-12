package transform

import (
	"strings"
	"unicode"
)

// Slugify converts a string to a URL-friendly slug matching Zod v4's
// slugify behavior:
//
//  1. Lowercase the input
//  2. Trim whitespace
//  3. Remove non-word, non-space, non-hyphen characters
//  4. Collapse consecutive whitespace, underscores, and hyphens
//     into a single hyphen
//  5. Trim leading and trailing hyphens
func Slugify(s string) string {
	if s == "" {
		return ""
	}

	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))

	prev := false // whether the last written byte was a hyphen
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			prev = false
			b.WriteRune(r)
		case isSeparator(r):
			if !prev && b.Len() > 0 {
				b.WriteByte('-')
				prev = true
			}
		}
	}

	return strings.TrimRight(b.String(), "-")
}

// isSeparator reports whether r is a whitespace or delimiter character
// that should be collapsed into a single hyphen.
func isSeparator(r rune) bool {
	return r == ' ' || r == '_' || r == '-' || r == '\t' || r == '\n' || r == '\r'
}
