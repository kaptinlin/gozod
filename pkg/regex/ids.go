package regex

import (
	"regexp"
	"strconv"
)

// ID format patterns.
var (
	// CUID matches strings in CUID format (starts with 'c' or 'C', 9+ chars).
	CUID = regexp.MustCompile(`^[cC][^\s-]{8,}$`)

	// CUID2 matches strings in CUID2 format (lowercase alphanumeric).
	CUID2 = regexp.MustCompile(`^[0-9a-z]+$`)

	// ULID matches strings in ULID format (26 Crockford Base32 chars).
	ULID = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{26}$`)

	// XID matches strings in XID format (20 base32 chars).
	XID = regexp.MustCompile(`^[0-9a-vA-V]{20}$`)

	// KSUID matches strings in KSUID format (27 alphanumeric chars).
	KSUID = regexp.MustCompile(`^[A-Za-z0-9]{27}$`)

	// NanoID matches strings in NanoID format (21 URL-safe chars).
	NanoID = regexp.MustCompile(`^[a-zA-Z0-9_-]{21}$`)

	// GUID matches any UUID-like identifier with 8-4-4-4-12 hex pattern.
	GUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	// UUID matches RFC 4122 UUID (versions 1-8) or the nil UUID.
	UUID = regexp.MustCompile(`^(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000)$`)

	// UUID4 matches version 4 UUIDs.
	UUID4 = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

	// UUID6 matches version 6 UUIDs.
	UUID6 = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-6[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

	// UUID7 matches version 7 UUIDs.
	UUID7 = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-7[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
)

// uuidVersionCache holds pre-compiled regexes for UUID versions 1-8.
var uuidVersionCache = func() [9]*regexp.Regexp {
	var cache [9]*regexp.Regexp
	for v := range 9 {
		if v >= 1 {
			cache[v] = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-` + strconv.Itoa(v) + `[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
		}
	}
	return cache
}()

// UUIDForVersion returns a regex matching a specific UUID version (1-8).
// For out-of-range versions, it returns the general UUID regex.
func UUIDForVersion(version int) *regexp.Regexp {
	if version < 1 || version > 8 {
		return UUID
	}
	return uuidVersionCache[version]
}
