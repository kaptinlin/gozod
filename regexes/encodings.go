package regexes

import "regexp"

// Base64 matches Base64-encoded strings
// TypeScript original code:
// export const base64: RegExp = /^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$/;
var Base64 = regexp.MustCompile(`^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$`)

// Base64URL matches URL-safe Base64-encoded strings
// TypeScript original code:
// export const base64url: RegExp = /^[A-Za-z0-9_-]*$/;
var Base64URL = regexp.MustCompile(`^[A-Za-z0-9_-]*$`)
