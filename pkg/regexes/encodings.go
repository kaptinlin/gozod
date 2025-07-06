package regexes

import "regexp"

// Base64 matches Base64-encoded strings
// TypeScript original code:
// export const base64: RegExp = /^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$/;
var Base64 = regexp.MustCompile(`^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$`)

// Base64URL pattern allows URL-safe Base64 characters with optional "=" padding (up to 2)
var Base64URL = regexp.MustCompile(`^[A-Za-z0-9_-]*={0,2}$`)
