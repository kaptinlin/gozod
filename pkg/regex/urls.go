package regex

import "regexp"

var (
	// URL matches valid URLs with common protocols (http, https, ftp, ws, file, etc.).
	URL = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^\s/$.?#].[^\s]*$`)

	// HTTPURL matches URLs starting with http:// or https://.
	HTTPURL = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
)
