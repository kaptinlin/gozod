// Package regex provides pre-compiled regular expressions for common data formats.
//
// Key features:
//   - Email patterns (Email, HTML5Email, RFC5322Email, UnicodeEmail)
//   - URL patterns (URL)
//   - Network patterns (IPv4, IPv6, CIDR, MAC)
//   - Date/time patterns (Date, Time, Datetime, Duration)
//   - ID formats (UUID, CUID, CUID2, ULID, NanoID, XID, KSUID)
//   - Encoding patterns (Base64, Base64URL)
//   - Primitives (String, Integer, Number, Boolean)
//
// Usage:
//
//	if regex.Email.MatchString(email) {
//	    // valid email format
//	}
//
//	if regex.UUID.MatchString(uuid) {
//	    // valid UUID format
//	}
package regex
