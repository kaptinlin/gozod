package regex

import "regexp"

// Encoding format patterns.
var (
	// Base64 matches standard Base64-encoded strings (RFC 4648).
	Base64 = regexp.MustCompile(`^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$`)

	// Base64URL matches URL-safe Base64 strings with optional padding.
	Base64URL = regexp.MustCompile(`^[A-Za-z0-9_-]*={0,2}$`)

	// Hex matches hexadecimal strings (any length including empty).
	Hex = regexp.MustCompile(`^[0-9a-fA-F]*$`)
)

// Hash format patterns.
var (
	// MD5Hex matches MD5 hashes in hexadecimal format (32 chars).
	MD5Hex = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)

	// SHA1Hex matches SHA-1 hashes in hexadecimal format (40 chars).
	SHA1Hex = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

	// SHA256Hex matches SHA-256 hashes in hexadecimal format (64 chars).
	SHA256Hex = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

	// SHA384Hex matches SHA-384 hashes in hexadecimal format (96 chars).
	SHA384Hex = regexp.MustCompile(`^[0-9a-fA-F]{96}$`)

	// SHA512Hex matches SHA-512 hashes in hexadecimal format (128 chars).
	SHA512Hex = regexp.MustCompile(`^[0-9a-fA-F]{128}$`)
)
