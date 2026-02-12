package regex

import "regexp"

// URL matches valid URLs with common protocols (http, https, ftp, ws, file, etc.).
var URL = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^\s/$.?#].[^\s]*$`)

// HTTPURL matches URLs starting with http:// or https://.
var HTTPURL = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
