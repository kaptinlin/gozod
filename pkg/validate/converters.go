package validate

import (
	"regexp"
	"strings"
)

// Pre-compiled regular expressions for Slugify function
// This avoids recompiling on every call for better performance
var (
	slugifyNonWordSpaceHyphen    = regexp.MustCompile(`[^\w\s-]`)
	slugifySpaceUnderscoreHyphen = regexp.MustCompile(`[\s_-]+`)
	slugifyLeadingTrailingHyphen = regexp.MustCompile(`^-+|-+$`)
)

// Slugify converts a string to a URL-friendly slug.
// Matches Zod's implementation:
// 1. Lowercase
// 2. Trim whitespace
// 3. Remove non-word/non-space/non-hyphen characters
// 4. Replace spaces and underscores with hyphens
// 5. Trim leading/trailing hyphens
func Slugify(input string) string {
	slug := strings.ToLower(input)
	slug = strings.TrimSpace(slug)
	slug = slugifyNonWordSpaceHyphen.ReplaceAllString(slug, "")
	slug = slugifySpaceUnderscoreHyphen.ReplaceAllString(slug, "-")
	slug = slugifyLeadingTrailingHyphen.ReplaceAllString(slug, "")
	return slug
}
