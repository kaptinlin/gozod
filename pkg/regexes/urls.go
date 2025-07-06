package regexes

import "regexp"

// URL matches valid URLs with proper format validation
// Supports common protocols: http, https, ftp, ftps, ws, wss, file, etc.
// Based on RFC 3986 with practical adaptations for common use cases
var URL = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^\s/$.?#].[^\s]*$`)

// HTTPUrl matches URLs starting with http or https specifically
var HTTPUrl = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
