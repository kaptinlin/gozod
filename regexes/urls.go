package regexes

import "regexp"

// URL matches URLs starting with http or https (simplified Go-specific implementation)
// Note: TypeScript Zod doesn't have a simple URL regex export - URL validation is typically
// done through more comprehensive validation logic rather than regex patterns
var URL = regexp.MustCompile(`^https?://`)
