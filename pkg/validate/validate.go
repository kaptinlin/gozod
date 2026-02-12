package validate

import (
	"math"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	json "github.com/go-json-experiment/json"
	"github.com/golang-jwt/jwt/v5"

	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/regex"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// Pre-compiled regex patterns.
var (
	isoTimeRegex  = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9]([.,][0-9]+)?)?$`)
	durationRegex = regexp.MustCompile(`^P(?:\d+W|(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:[.,]\d+)?S)?)?)$`)
)

// matchString extracts a string from value and checks it against pattern.
func matchString(value any, pattern *regexp.Regexp) bool {
	str, ok := reflectx.ExtractString(value)
	if !ok {
		return false
	}
	return pattern.MatchString(str)
}

// Numeric validation

// Lt validates if value is less than limit
func Lt(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) < toFloat64(limit)
}

// Lte validates if value is less than or equal to limit
func Lte(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) <= toFloat64(limit)
}

// Gt validates if value is greater than limit
func Gt(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) > toFloat64(limit)
}

// Gte validates if value is greater than or equal to limit
func Gte(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) >= toFloat64(limit)
}

// Positive validates if numeric value is positive (> 0)
func Positive(value any) bool {
	return Gt(value, 0)
}

// Negative validates if numeric value is negative (< 0)
func Negative(value any) bool {
	return Lt(value, 0)
}

// NonPositive validates if numeric value is non-positive (<= 0)
func NonPositive(value any) bool {
	return Lte(value, 0)
}

// NonNegative validates if numeric value is non-negative (>= 0)
func NonNegative(value any) bool {
	return Gte(value, 0)
}

// MultipleOf validates if value is a multiple of the given divisor
func MultipleOf(value, divisor any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(divisor) {
		return false
	}

	valFloat := toFloat64(value)
	divFloat := toFloat64(divisor)

	if divFloat == 0 {
		return false
	}

	remainder := math.Mod(valFloat, divFloat)
	return math.Abs(remainder) < 1e-10 // Handle floating point precision
}

// Length validation

// MaxLength validates if value's length is at most maximum
func MaxLength(value any, maximum int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l <= maximum
}

// MinLength validates if value's length is at least minimum
func MinLength(value any, minimum int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l >= minimum
}

// Length validates if value's length equals exactly the expected length
func Length(value any, expected int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l == expected
}

// Size validation

// getCollectionSize returns the size of a collection (map, slice, array, string, etc.).
func getCollectionSize(value any) (int, bool) {
	if reflectx.IsMap(value) {
		// Support both map[string]any and map[any]any first
		if m, ok := value.(map[string]any); ok {
			return mapx.Count(m), true
		}
		if m, ok := value.(map[any]any); ok {
			return len(m), true
		}
		// Fall back to reflection for any other map type (e.g., map[T]struct{} for sets)
		if size, ok := reflectx.Size(value); ok {
			return size, true
		}
	}
	if reflectx.HasLength(value) {
		return reflectx.Length(value)
	}
	return 0, false
}

// MaxSize validates if collection's size is at most maximum
func MaxSize(value any, maximum int) bool {
	if size, ok := getCollectionSize(value); ok {
		return size <= maximum
	}
	return false
}

// MinSize validates if collection's size is at least minimum
func MinSize(value any, minimum int) bool {
	if size, ok := getCollectionSize(value); ok {
		return size >= minimum
	}
	return false
}

// Size validates if collection's size equals exactly the expected size
func Size(value any, expected int) bool {
	if size, ok := getCollectionSize(value); ok {
		return size == expected
	}
	return false
}

// String format validation

// Regex validates if string matches the pattern.
func Regex(value any, pattern *regexp.Regexp) bool {
	return matchString(value, pattern)
}

// Lowercase validates if string contains no uppercase letters.
// Matches Zod v4's lowercase check: /^[^A-Z]*$/
func Lowercase(value any) bool {
	return matchString(value, regex.Lowercase)
}

// Uppercase validates if string contains no lowercase letters.
// Matches Zod v4's uppercase check: /^[^a-z]*$/
func Uppercase(value any) bool {
	return matchString(value, regex.Uppercase)
}

// Includes validates if string contains the substring
func Includes(value any, substring string) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return strings.Contains(str, substring)
	}
	return false
}

// StartsWith validates if string starts with the prefix
func StartsWith(value any, prefix string) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return strings.HasPrefix(str, prefix)
	}
	return false
}

// EndsWith validates if string ends with the suffix
func EndsWith(value any, suffix string) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return strings.HasSuffix(str, suffix)
	}
	return false
}

// Format validation

// Email validates if string is a valid email format.
func Email(value any) bool {
	return matchString(value, regex.Email)
}

// URL validates if string is a valid URL format.
func URL(value any) bool {
	return matchString(value, regex.URL)
}

// Hostname validates if string is a valid DNS hostname (RFC 1123, max 253 chars).
func Hostname(value any) bool {
	str, ok := reflectx.ExtractString(value)
	if !ok || len(str) == 0 || len(str) > 253 {
		return false
	}
	return regex.Hostname.MatchString(str)
}

// UUID validates if string is a valid UUID format.
func UUID(value any) bool { return matchString(value, regex.UUID) }

// GUID validates if string is a valid GUID format.
func GUID(value any) bool { return matchString(value, regex.GUID) }

// CUID validates if string is a valid CUID format.
func CUID(value any) bool { return matchString(value, regex.CUID) }

// CUID2 validates if string is a valid CUID2 format.
func CUID2(value any) bool { return matchString(value, regex.CUID2) }

// NanoID validates if string is a valid NanoID format.
func NanoID(value any) bool { return matchString(value, regex.NanoID) }

// ULID validates if string is a valid ULID format.
func ULID(value any) bool { return matchString(value, regex.ULID) }

// XID validates if string is a valid XID format.
func XID(value any) bool { return matchString(value, regex.XID) }

// KSUID validates if string is a valid KSUID format.
func KSUID(value any) bool { return matchString(value, regex.KSUID) }

// IPv4 validates if string is a valid IPv4 address.
func IPv4(value any) bool { return matchString(value, regex.IPv4) }

// IPv6 validates if string is a valid IPv6 address.
func IPv6(value any) bool { return matchString(value, regex.IPv6) }

// CIDR validates CIDR notation with strict format checking.
// version: 0=both, 4=IPv4 only, 6=IPv6 only
func CIDR(value any, version int) bool {
	str, ok := reflectx.ExtractString(value)
	if !ok {
		return false
	}

	// Step 1: Strict check for exactly one slash
	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		return false
	}

	// Step 2: Use standard library for parsing
	_, ipnet, err := net.ParseCIDR(str)
	if err != nil {
		return false
	}

	// Step 3: Optional version check
	if version != 0 {
		isV4 := ipnet.IP.To4() != nil
		if version == 4 && !isV4 {
			return false
		}
		if version == 6 && isV4 {
			return false
		}
	}

	return true
}

// CIDRv4 validates if string is a valid IPv4 CIDR format.
func CIDRv4(value any) bool {
	return CIDR(value, 4)
}

// CIDRv6 validates if string is a valid IPv6 CIDR format
func CIDRv6(value any) bool {
	return CIDR(value, 6)
}

// Base64 validates if string is valid Base64 encoding.
func Base64(value any) bool { return matchString(value, regex.Base64) }

// Base64URL validates if string is valid Base64URL encoding.
func Base64URL(value any) bool { return matchString(value, regex.Base64URL) }

// Hex validates if string is a valid hexadecimal string.
func Hex(value any) bool { return matchString(value, regex.Hex) }

// Hash validation

// MD5Hex validates if string is a valid MD5 hash in hex format (32 chars).
func MD5Hex(value any) bool { return matchString(value, regex.MD5Hex) }

// SHA1Hex validates if string is a valid SHA-1 hash in hex format (40 chars).
func SHA1Hex(value any) bool { return matchString(value, regex.SHA1Hex) }

// SHA256Hex validates if string is a valid SHA-256 hash in hex format (64 chars).
func SHA256Hex(value any) bool { return matchString(value, regex.SHA256Hex) }

// SHA384Hex validates if string is a valid SHA-384 hash in hex format (96 chars).
func SHA384Hex(value any) bool { return matchString(value, regex.SHA384Hex) }

// SHA512Hex validates if string is a valid SHA-512 hash in hex format (128 chars).
func SHA512Hex(value any) bool { return matchString(value, regex.SHA512Hex) }

// E164 validates if string is a valid E.164 phone number format.
func E164(value any) bool { return matchString(value, regex.E164) }

// MAC address validation

// MACOptions defines configuration for MAC address validation
type MACOptions struct {
	// Delimiter specifies the separator between hex pairs (default: ":")
	// Common delimiters: ":", "-", "."
	// Examples:
	//   ":" -> "00:1A:2B:3C:4D:5E"
	//   "-" -> "00-1A-2B-3C-4D-5E"
	//   "." -> "00.1A.2B.3C.4D.5E"
	Delimiter string
}

// MAC validates if string is a valid MAC address using default colon delimiter.
// Accepts both uppercase and lowercase hex digits.
// Examples:
//
//	"00:1A:2B:3C:4D:5E" -> true
//	"00:1a:2b:3c:4d:5e" -> true
//	"00-1A-2B-3C-4D-5E" -> false (wrong delimiter)
//	"00:1A:2B:3C:4D"    -> false (incomplete)
func MAC(value any) bool {
	return MACWithOptions(value, MACOptions{Delimiter: ":"})
}

// MACWithOptions validates if string is a valid MAC address with custom delimiter.
// This implementation matches Zod's TypeScript version using case-sensitive branches.
//
// The validation accepts MAC addresses in the format:
//
//	XX:XX:XX:XX:XX:XX (where X is a hex digit, : can be replaced by delimiter)
//
// Both uppercase and lowercase hex digits are valid, but they should not be mixed
// within a single validation (though the regex allows both).
//
// Examples:
//
//	MACWithOptions("00:1a:2b:3c:4d:5e", MACOptions{Delimiter: ":"})  -> true
//	MACWithOptions("00-1A-2B-3C-4D-5E", MACOptions{Delimiter: "-"})  -> true
//	MACWithOptions("00.1a.2b.3c.4d.5e", MACOptions{Delimiter: "."})  -> true
func MACWithOptions(value any, opts MACOptions) bool {
	str, ok := reflectx.ExtractString(value)
	if !ok {
		return false
	}

	delim := opts.Delimiter
	if delim == "" {
		delim = ":"
	}

	return regex.MAC(delim).MatchString(str)
}

// JWT validation

// JWTOptions defines options for JWT validation
type JWTOptions struct {
	// Algorithm specifies the expected signing algorithm
	// If nil, any algorithm is accepted (not recommended for production)
	Algorithm *string
}

// JWT validates JWT format only (structure and basic claims).
// WARNING: Does NOT verify signatures.
func JWT(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return isValidJWT(str, nil)
	}
	return false
}

// JWTWithOptions validates JWT format with algorithm constraint.
// WARNING: Does NOT verify signatures.
func JWTWithOptions(value any, options JWTOptions) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return isValidJWT(str, options.Algorithm)
	}
	return false
}

// isValidJWT validates JWT structure without signature verification.
func isValidJWT(token string, expectedAlgorithm *string) bool {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())

	parsedToken, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return false
	}

	return validateJWTHeader(parsedToken.Header, expectedAlgorithm)
}

// validateJWTHeader validates JWT header fields.
func validateJWTHeader(header map[string]any, expectedAlgorithm *string) bool {
	// Check if typ claim exists and is "JWT" (if present)
	if typ, exists := header["typ"]; exists {
		if typStr, ok := typ.(string); ok && typStr != "JWT" {
			return false
		}
	}

	// Check if alg claim exists
	alg, exists := header["alg"]
	if !exists {
		return false
	}

	algStr, ok := alg.(string)
	if !ok {
		return false
	}

	// Reject "none" algorithm for security
	if algStr == "none" {
		return false
	}

	// Validate expected algorithm if specified
	if expectedAlgorithm != nil && algStr != *expectedAlgorithm {
		return false
	}

	return true
}

// Emoji validation

// Emoji validates if string is valid emoji.
func Emoji(value any) bool { return matchString(value, regex.Emoji) }

// ISODateTimeOptions defines parameters for ISO datetime validation
type ISODateTimeOptions struct {
	// Precision specifies number of decimal places for seconds
	// If nil, matches any number of decimal places
	// If 0 or negative, no decimal places allowed
	Precision *int

	// Offset if true, allows timezone offsets like +01:00
	Offset bool

	// Local if true, makes the 'Z' timezone marker optional
	Local bool
}

// ISODateTimeWithOptions validates if string is a valid ISO datetime format with options
func ISODateTimeWithOptions(value any, options ISODateTimeOptions) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		datetimeRegex := regex.Datetime(regex.DatetimeOptions{
			Precision: options.Precision,
			Offset:    options.Offset,
			Local:     options.Local,
		})
		return datetimeRegex.MatchString(str)
	}
	return false
}

// ISODateTime validates if string is a valid ISO datetime format
func ISODateTime(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		_, err := time.Parse(time.RFC3339, str)
		return err == nil
	}
	return false
}

// ISODate validates if string is a valid ISO date format
func ISODate(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		_, err := time.Parse("2006-01-02", str)
		return err == nil
	}
	return false
}

// ISOTimeOptions defines parameters for ISO time validation
type ISOTimeOptions struct {
	// Precision specifies number of decimal places for seconds
	// If nil, matches any number of decimal places
	// If 0 or negative, no decimal places allowed
	// Special case: -1 means minute precision only (no seconds)
	Precision *int
}

// ISOTimeWithOptions validates if string is a valid ISO time format with options
func ISOTimeWithOptions(value any, options ISOTimeOptions) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		timeRegex := regex.Time(regex.TimeOptions{
			Precision: options.Precision,
		})
		return timeRegex.MatchString(str)
	}
	return false
}

// ISOTime validates if string is a valid ISO time format.
func ISOTime(value any) bool {
	return matchString(value, isoTimeRegex)
}

// ISODuration validates if string is a valid ISO 8601 duration format.
func ISODuration(value any) bool {
	duration, ok := reflectx.ExtractString(value)
	if !ok {
		return false
	}

	if !durationRegex.MatchString(duration) {
		return false
	}

	// Reject empty patterns like "P" or "PT".
	if duration == "P" || duration == "PT" {
		return false
	}

	// If there's a T, there must be at least one time component after it.
	if _, afterT, found := strings.Cut(duration, "T"); found {
		if !strings.ContainsAny(afterT, "HMS") {
			return false
		}
	}

	// Ensure at least one digit after P.
	for i, r := range duration {
		if i > 0 && r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// JSON validates if string is valid JSON
func JSON(value any) bool {
	str, err := coerce.ToString(value)
	if err != nil {
		return false
	}
	var v any
	return json.Unmarshal([]byte(str), &v) == nil
}

// Property validation

// Property validates if object has a property that passes validation.
func Property(obj any, key string, validator func(any) bool) bool {
	if reflectx.IsMap(obj) {
		if m, ok := obj.(map[string]any); ok {
			if value, exists := mapx.Get(m, key); exists {
				return validator(value)
			}
		}
	}
	return false
}

// MIME type validation

// Mime validates if value has allowed MIME types.
func Mime(value any, allowedTypes []string) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		// Basic MIME type format check (type/subtype)
		parts := strings.Split(str, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return slicex.Contains(allowedTypes, str)
		}
	}
	return false
}

// Helpers

// toFloat64 converts any value to float64 using the coerce package.
func toFloat64(value any) float64 {
	if result, err := coerce.ToFloat64(value); err == nil {
		return result
	}
	return 0
}

// URL validation with options

// URLOptions defines options for URL validation.
type URLOptions struct {
	// Hostname validation pattern
	Hostname *regexp.Regexp
	// Protocol validation pattern
	Protocol *regexp.Regexp
}

// URLWithOptions validates if string is a valid URL format with optional constraints
func URLWithOptions(value any, options URLOptions) bool {
	str, ok := reflectx.ExtractString(value)
	if !ok {
		return false
	}

	// First check basic URL format using regex
	if !regex.URL.MatchString(str) {
		return false
	}

	// If no constraints, basic validation is sufficient
	if options.Hostname == nil && options.Protocol == nil {
		return true
	}

	// Parse URL for constraint validation
	u, err := url.Parse(str)
	if err != nil {
		return false
	}

	// Validate hostname constraint if provided
	if options.Hostname != nil && !options.Hostname.MatchString(u.Hostname()) {
		return false
	}

	// Validate protocol constraint if provided
	if options.Protocol != nil && !options.Protocol.MatchString(u.Scheme) {
		return false
	}

	return true
}
