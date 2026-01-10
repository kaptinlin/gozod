package regex

import "regexp"

// Base64 matches Base64-encoded strings
// TypeScript original code:
// export const base64: RegExp = /^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$/;
var Base64 = regexp.MustCompile(`^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$`)

// Base64URL pattern allows URL-safe Base64 characters with optional "=" padding (up to 2)
var Base64URL = regexp.MustCompile(`^[A-Za-z0-9_-]*={0,2}$`)

// Hex matches hexadecimal strings (any length including empty)
// TypeScript Zod v4: /^[0-9a-fA-F]*$/
var Hex = regexp.MustCompile(`^[0-9a-fA-F]*$`)

// =============================================================================
// HASH PATTERNS
// =============================================================================

// MD5 matches MD5 hash in hexadecimal format (32 hex chars)
var MD5Hex = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)

// SHA1 matches SHA-1 hash in hexadecimal format (40 hex chars)
var SHA1Hex = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

// SHA256 matches SHA-256 hash in hexadecimal format (64 hex chars)
var SHA256Hex = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

// SHA384 matches SHA-384 hash in hexadecimal format (96 hex chars)
var SHA384Hex = regexp.MustCompile(`^[0-9a-fA-F]{96}$`)

// SHA512 matches SHA-512 hash in hexadecimal format (128 hex chars)
var SHA512Hex = regexp.MustCompile(`^[0-9a-fA-F]{128}$`)
