package validate

import (
	"math"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/golang-jwt/jwt/v5"

	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/regex"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// Pre-compiled regex patterns for ISO time and duration validation.
var (
	isoTimeRegex  = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9]([.,][0-9]+)?)?$`)
	durationRegex = regexp.MustCompile(`^P(?:\d+W|(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:[.,]\d+)?S)?)?)$`)
)

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

// Lt reports whether value is less than limit.
func Lt(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) < toFloat64(limit)
}

// Lte reports whether value is less than or equal to limit.
func Lte(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) <= toFloat64(limit)
}

// Gt reports whether value is greater than limit.
func Gt(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) > toFloat64(limit)
}

// Gte reports whether value is greater than or equal to limit.
func Gte(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) >= toFloat64(limit)
}

// Positive reports whether the numeric value is positive (> 0).
func Positive(value any) bool { return Gt(value, 0) }

// Negative reports whether the numeric value is negative (< 0).
func Negative(value any) bool { return Lt(value, 0) }

// NonPositive reports whether the numeric value is non-positive (<= 0).
func NonPositive(value any) bool { return Lte(value, 0) }

// NonNegative reports whether the numeric value is non-negative (>= 0).
func NonNegative(value any) bool { return Gte(value, 0) }

// MultipleOf reports whether value is a multiple of the given divisor.
func MultipleOf(value, divisor any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(divisor) {
		return false
	}
	val := toFloat64(value)
	div := toFloat64(divisor)
	if div == 0 {
		return false
	}
	// Handle floating point precision.
	return math.Abs(math.Mod(val, div)) < 1e-10
}

// MaxLength reports whether value's length is at most maximum.
func MaxLength(value any, maximum int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l <= maximum
}

// MinLength reports whether value's length is at least minimum.
func MinLength(value any, minimum int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l >= minimum
}

// Length reports whether value's length equals exactly the expected length.
func Length(value any, expected int) bool {
	l, ok := reflectx.Length(value)
	if !ok {
		return false
	}
	return l == expected
}

// MaxSize reports whether the collection's size is at most maximum.
func MaxSize(value any, maximum int) bool {
	size, ok := collectionSize(value)
	if !ok {
		return false
	}
	return size <= maximum
}

// MinSize reports whether the collection's size is at least minimum.
func MinSize(value any, minimum int) bool {
	size, ok := collectionSize(value)
	if !ok {
		return false
	}
	return size >= minimum
}

// Size reports whether the collection's size equals exactly the expected size.
func Size(value any, expected int) bool {
	size, ok := collectionSize(value)
	if !ok {
		return false
	}
	return size == expected
}

// Regex reports whether the string matches the pattern.
func Regex(value any, pattern *regexp.Regexp) bool {
	return matchString(value, pattern)
}

// Lowercase reports whether the string contains no uppercase letters.
// Matches Zod v4's lowercase check: /^[^A-Z]*$/
func Lowercase(value any) bool { return matchString(value, regex.Lowercase) }

// Uppercase reports whether the string contains no lowercase letters.
// Matches Zod v4's uppercase check: /^[^a-z]*$/
func Uppercase(value any) bool { return matchString(value, regex.Uppercase) }

// Includes reports whether the string contains the substring.
func Includes(value any, substring string) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return strings.Contains(str, substring)
}

// StartsWith reports whether the string starts with the prefix.
func StartsWith(value any, prefix string) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return strings.HasPrefix(str, prefix)
}

// EndsWith reports whether the string ends with the suffix.
func EndsWith(value any, suffix string) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return strings.HasSuffix(str, suffix)
}

// Email reports whether the string is a valid email format.
func Email(value any) bool { return matchString(value, regex.Email) }

// URL reports whether the string is a valid URL format.
func URL(value any) bool { return matchString(value, regex.URL) }

// Hostname reports whether the string is a valid DNS hostname (RFC 1123, max 253 chars).
func Hostname(value any) bool {
	str, ok := reflectx.StringVal(value)
	if !ok || len(str) == 0 || len(str) > 253 {
		return false
	}
	return regex.Hostname.MatchString(str)
}

// URLWithOptions reports whether the string is a valid URL with the given constraints.
func URLWithOptions(value any, options URLOptions) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	if !regex.URL.MatchString(str) {
		return false
	}
	// If no constraints, basic validation is sufficient.
	if options.Hostname == nil && options.Protocol == nil {
		return true
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if options.Hostname != nil && !options.Hostname.MatchString(u.Hostname()) {
		return false
	}
	if options.Protocol != nil && !options.Protocol.MatchString(u.Scheme) {
		return false
	}
	return true
}

// UUID reports whether the string is a valid UUID format.
func UUID(value any) bool { return matchString(value, regex.UUID) }

// GUID reports whether the string is a valid GUID format.
func GUID(value any) bool { return matchString(value, regex.GUID) }

// CUID reports whether the string is a valid CUID format.
func CUID(value any) bool { return matchString(value, regex.CUID) }

// CUID2 reports whether the string is a valid CUID2 format.
func CUID2(value any) bool { return matchString(value, regex.CUID2) }

// NanoID reports whether the string is a valid NanoID format.
func NanoID(value any) bool { return matchString(value, regex.NanoID) }

// ULID reports whether the string is a valid ULID format.
func ULID(value any) bool { return matchString(value, regex.ULID) }

// XID reports whether the string is a valid XID format.
func XID(value any) bool { return matchString(value, regex.XID) }

// KSUID reports whether the string is a valid KSUID format.
func KSUID(value any) bool { return matchString(value, regex.KSUID) }

// IPv4 reports whether the string is a valid IPv4 address.
func IPv4(value any) bool { return matchString(value, regex.IPv4) }

// IPv6 reports whether the string is a valid IPv6 address.
func IPv6(value any) bool { return matchString(value, regex.IPv6) }

// CIDR reports whether the string is valid CIDR notation.
// The version parameter filters by IP version: 0 for any, 4 for IPv4, 6 for IPv6.
func CIDR(value any, version int) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	if parts := strings.Split(str, "/"); len(parts) != 2 {
		return false
	}
	_, ipnet, err := net.ParseCIDR(str)
	if err != nil {
		return false
	}
	if version == 0 {
		return true
	}
	isV4 := ipnet.IP.To4() != nil
	if version == 4 && !isV4 {
		return false
	}
	if version == 6 && isV4 {
		return false
	}
	return true
}

// CIDRv4 reports whether the string is a valid IPv4 CIDR notation.
func CIDRv4(value any) bool { return CIDR(value, 4) }

// CIDRv6 reports whether the string is a valid IPv6 CIDR notation.
func CIDRv6(value any) bool { return CIDR(value, 6) }

// Base64 reports whether the string is valid Base64 encoding.
func Base64(value any) bool { return matchString(value, regex.Base64) }

// Base64URL reports whether the string is valid Base64URL encoding.
func Base64URL(value any) bool { return matchString(value, regex.Base64URL) }

// Hex reports whether the string is a valid hexadecimal string.
func Hex(value any) bool { return matchString(value, regex.Hex) }

// MD5Hex reports whether the string is a valid MD5 hash in hex format (32 chars).
func MD5Hex(value any) bool { return matchString(value, regex.MD5Hex) }

// SHA1Hex reports whether the string is a valid SHA-1 hash in hex format (40 chars).
func SHA1Hex(value any) bool { return matchString(value, regex.SHA1Hex) }

// SHA256Hex reports whether the string is a valid SHA-256 hash in hex format (64 chars).
func SHA256Hex(value any) bool { return matchString(value, regex.SHA256Hex) }

// SHA384Hex reports whether the string is a valid SHA-384 hash in hex format (96 chars).
func SHA384Hex(value any) bool { return matchString(value, regex.SHA384Hex) }

// SHA512Hex reports whether the string is a valid SHA-512 hash in hex format (128 chars).
func SHA512Hex(value any) bool { return matchString(value, regex.SHA512Hex) }

// E164 reports whether the string is a valid E.164 phone number format.
func E164(value any) bool { return matchString(value, regex.E164) }

// Emoji reports whether the string is a valid emoji.
func Emoji(value any) bool { return matchString(value, regex.Emoji) }

// MAC reports whether the string is a valid MAC address using the default colon delimiter.
// Accepts both uppercase and lowercase hex digits.
func MAC(value any) bool {
	return MACWithOptions(value, MACOptions{Delimiter: ":"})
}

// MACWithOptions reports whether the string is a valid MAC address with a custom delimiter.
func MACWithOptions(value any, opts MACOptions) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	delim := opts.Delimiter
	if delim == "" {
		delim = ":"
	}
	return regex.MAC(delim).MatchString(str)
}

// JWT reports whether the string is a valid JWT format (structure and basic claims).
// It does NOT verify signatures.
func JWT(value any) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return isValidJWT(str, nil)
}

// JWTWithOptions reports whether the string is a valid JWT format with an algorithm constraint.
// It does NOT verify signatures.
func JWTWithOptions(value any, options JWTOptions) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return isValidJWT(str, options.Algorithm)
}

// ISODateTimeWithOptions reports whether the string is a valid ISO datetime with the given options.
func ISODateTimeWithOptions(value any, options ISODateTimeOptions) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	datetimeRegex := regex.Datetime(regex.DatetimeOptions{
		Precision: options.Precision,
		Offset:    options.Offset,
		Local:     options.Local,
	})
	return datetimeRegex.MatchString(str)
}

// ISODateTime reports whether the string is a valid ISO datetime format.
func ISODateTime(value any) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	_, err := time.Parse(time.RFC3339, str)
	return err == nil
}

// ISODate reports whether the string is a valid ISO date format.
func ISODate(value any) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	_, err := time.Parse("2006-01-02", str)
	return err == nil
}

// ISOTimeWithOptions reports whether the string is a valid ISO time with the given options.
func ISOTimeWithOptions(value any, options ISOTimeOptions) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	timeRegex := regex.Time(regex.TimeOptions{
		Precision: options.Precision,
	})
	return timeRegex.MatchString(str)
}

// ISOTime reports whether the string is a valid ISO time format.
func ISOTime(value any) bool { return matchString(value, isoTimeRegex) }

// ISODuration reports whether the string is a valid ISO 8601 duration format.
func ISODuration(value any) bool {
	duration, ok := reflectx.StringVal(value)
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

// JSON reports whether the string is valid JSON.
func JSON(value any) bool {
	str, err := coerce.ToString(value)
	if err != nil {
		return false
	}
	var v any
	return json.Unmarshal([]byte(str), &v) == nil
}

// Property reports whether the object has a property at key that passes the validator.
func Property(obj any, key string, validator func(any) bool) bool {
	if !reflectx.IsMap(obj) {
		return false
	}
	m, ok := obj.(map[string]any)
	if !ok {
		return false
	}
	value, exists := mapx.Get(m, key)
	if !exists {
		return false
	}
	return validator(value)
}

// Mime reports whether the value is one of the allowed MIME types.
func Mime(value any, allowedTypes []string) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	parts := strings.Split(str, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	return slicex.Contains(allowedTypes, str)
}

// matchString extracts a string from value and checks it against pattern.
func matchString(value any, pattern *regexp.Regexp) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return pattern.MatchString(str)
}

// toFloat64 converts any value to float64 using the coerce package.
func toFloat64(value any) float64 {
	result, err := coerce.ToFloat64(value)
	if err != nil {
		return 0
	}
	return result
}

// collectionSize returns the size of a collection (map, slice, array, string).
func collectionSize(value any) (int, bool) {
	if reflectx.IsMap(value) {
		if m, ok := value.(map[string]any); ok {
			return mapx.Count(m), true
		}
		if m, ok := value.(map[any]any); ok {
			return len(m), true
		}
		// Fall back to reflection for other map types (e.g., map[T]struct{} for sets).
		if size, ok := reflectx.Size(value); ok {
			return size, true
		}
	}
	if reflectx.HasLength(value) {
		return reflectx.Length(value)
	}
	return 0, false
}

// isValidJWT reports whether the token has valid JWT structure without verifying signatures.
func isValidJWT(token string, expectedAlgorithm *string) bool {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	parsedToken, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return false
	}
	return validateJWTHeader(parsedToken.Header, expectedAlgorithm)
}

// validateJWTHeader reports whether the JWT header fields are valid.
func validateJWTHeader(header map[string]any, expectedAlgorithm *string) bool {
	if typ, exists := header["typ"]; exists {
		if typStr, ok := typ.(string); ok && typStr != "JWT" {
			return false
		}
	}
	alg, exists := header["alg"]
	if !exists {
		return false
	}
	algStr, ok := alg.(string)
	if !ok || algStr == "none" {
		return false
	}
	if expectedAlgorithm != nil && algStr != *expectedAlgorithm {
		return false
	}
	return true
}
