package validate

import "regexp"

// MACOptions configures MAC address validation.
type MACOptions struct {
	// Delimiter specifies the separator between hex pairs.
	// Common delimiters: ":", "-", ".". Default is ":".
	Delimiter string
}

// JWTOptions configures JWT validation.
type JWTOptions struct {
	// Algorithm specifies the expected signing algorithm.
	// If nil, any algorithm is accepted (not recommended for production).
	Algorithm *string
}

// ISODateTimeOptions configures ISO datetime validation.
type ISODateTimeOptions struct {
	// Precision specifies number of decimal places for seconds.
	// If nil, matches any number of decimal places.
	// If 0 or negative, no decimal places allowed.
	Precision *int

	// Offset allows timezone offsets like +01:00 when true.
	Offset bool

	// Local makes the 'Z' timezone marker optional when true.
	Local bool
}

// ISOTimeOptions configures ISO time validation.
type ISOTimeOptions struct {
	// Precision specifies number of decimal places for seconds.
	// If nil, matches any number of decimal places.
	// If 0 or negative, no decimal places allowed.
	// Special case: -1 means minute precision only (no seconds).
	Precision *int
}

// URLOptions configures URL validation with hostname and protocol constraints.
type URLOptions struct {
	// Hostname is a pattern the URL's hostname must match.
	Hostname *regexp.Regexp
	// Protocol is a pattern the URL's scheme must match.
	Protocol *regexp.Regexp
}
