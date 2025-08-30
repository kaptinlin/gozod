package validate

import (
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	json "github.com/go-json-experiment/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/regexes"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// NUMERIC VALIDATION FUNCTIONS
// =============================================================================

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

// =============================================================================
// LENGTH VALIDATION FUNCTIONS
// =============================================================================

// MaxLength validates if value's length is at most maximum
func MaxLength(value any, maximum int) bool {
	if !reflectx.HasLength(value) {
		return false
	}
	l, _ := reflectx.GetLength(value)
	return l <= maximum
}

// MinLength validates if value's length is at least minimum
func MinLength(value any, minimum int) bool {
	if !reflectx.HasLength(value) {
		return false
	}
	if l, _ := reflectx.GetLength(value); l >= minimum {
		return true
	}
	return false
}

// Length validates if value's length equals exactly the expected length
func Length(value any, expected int) bool {
	if !reflectx.HasLength(value) {
		return false
	}
	if l, _ := reflectx.GetLength(value); l == expected {
		return true
	}
	return false
}

// =============================================================================
// SIZE VALIDATION FUNCTIONS
// =============================================================================

// MaxSize validates if collection's size is at most maximum
func MaxSize(value any, maximum int) bool {
	if reflectx.IsMap(value) {
		// Support both map[string]any and map[any]any
		if m, ok := value.(map[string]any); ok {
			return mapx.Count(m) <= maximum
		}
		if m, ok := value.(map[any]any); ok {
			return len(m) <= maximum
		}
	}
	if reflectx.HasLength(value) {
		if l, ok := reflectx.GetLength(value); ok {
			return l <= maximum
		}
	}
	return false
}

// MinSize validates if collection's size is at least minimum
func MinSize(value any, minimum int) bool {
	if reflectx.IsMap(value) {
		// Support both map[string]any and map[any]any
		if m, ok := value.(map[string]any); ok {
			return mapx.Count(m) >= minimum
		}
		if m, ok := value.(map[any]any); ok {
			return len(m) >= minimum
		}
	}
	if reflectx.HasLength(value) {
		if l, ok := reflectx.GetLength(value); ok {
			return l >= minimum
		}
	}
	return false
}

// Size validates if collection's size equals exactly the expected size
func Size(value any, expected int) bool {
	if reflectx.IsMap(value) {
		// Support both map[string]any and map[any]any
		if m, ok := value.(map[string]any); ok {
			return mapx.Count(m) == expected
		}
		if m, ok := value.(map[any]any); ok {
			return len(m) == expected
		}
	}
	if reflectx.HasLength(value) {
		if l, ok := reflectx.GetLength(value); ok {
			return l == expected
		}
	}
	return false
}

// =============================================================================
// STRING FORMAT VALIDATION FUNCTIONS
// =============================================================================

// Regex validates if string matches the pattern
func Regex(value any, pattern *regexp.Regexp) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return pattern.MatchString(str)
	}
	return false
}

// Lowercase validates if string is all lowercase
func Lowercase(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return str == strings.ToLower(str) && str != ""
	}
	return false
}

// Uppercase validates if string is all uppercase
func Uppercase(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return str == strings.ToUpper(str) && str != ""
	}
	return false
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

// =============================================================================
// FORMAT VALIDATION FUNCTIONS
// =============================================================================

// Email validates if string is a valid email format
func Email(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.Email.MatchString(str)
	}
	return false
}

// URL validates if string is a valid URL format
func URL(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.URL.MatchString(str)
	}
	return false
}

// UUID validates if string is a valid UUID format
func UUID(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.UUID.MatchString(str)
	}
	return false
}

// GUID validates if string is a valid GUID format
func GUID(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.GUID.MatchString(str)
	}
	return false
}

// CUID validates if string is a valid CUID format
func CUID(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.CUID.MatchString(str)
	}
	return false
}

// CUID2 validates if string is a valid CUID2 format
func CUID2(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.CUID2.MatchString(str)
	}
	return false
}

// NanoID validates if string is a valid NanoID format
func NanoID(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.NanoID.MatchString(str)
	}
	return false
}

// ULID validates if string is a valid ULID format
func ULID(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.ULID.MatchString(str)
	}
	return false
}

// XID validates if string is a valid XID format
func XID(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.XID.MatchString(str)
	}
	return false
}

// KSUID validates if string is a valid KSUID format
func KSUID(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.KSUID.MatchString(str)
	}
	return false
}

// IPv4 validates if string is a valid IPv4 address
func IPv4(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.IPv4.MatchString(str)
	}
	return false
}

// IPv6 validates if string is a valid IPv6 address
func IPv6(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.IPv6.MatchString(str)
	}
	return false
}

// CIDRv4 validates if string is a valid IPv4 CIDR format
func CIDRv4(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.CIDRv4.MatchString(str)
	}
	return false
}

// CIDRv6 validates if string is a valid IPv6 CIDR format
func CIDRv6(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.CIDRv6.MatchString(str)
	}
	return false
}

// Base64 validates if string is valid Base64 encoding
func Base64(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.Base64.MatchString(str)
	}
	return false
}

// Base64URL validates if string is valid Base64URL encoding
func Base64URL(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.Base64URL.MatchString(str)
	}
	return false
}

// E164 validates if string is a valid E.164 phone number format
func E164(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.E164.MatchString(str)
	}
	return false
}

// =============================================================================
// JWT TOKEN VALIDATION FUNCTIONS
// =============================================================================

// JWTOptions defines options for JWT validation
type JWTOptions struct {
	// Algorithm specifies the expected signing algorithm
	// If nil, any algorithm is accepted (not recommended for production)
	Algorithm *string
}

// JWT validates if string is a valid JWT token format
// This function only validates the JWT structure and basic claims
// It does NOT verify the signature (use JWTWithSecret for signature verification)
func JWT(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return isValidJWTStructure(str, nil)
	}
	return false
}

// JWTWithOptions validates if string is a valid JWT token format with algorithm constraint
// This function only validates the JWT structure and basic claims
// It does NOT verify the signature (use JWTWithSecret for signature verification)
func JWTWithOptions(value any, options JWTOptions) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return isValidJWTStructure(str, options.Algorithm)
	}
	return false
}

// isValidJWTStructure validates JWT structure without signature verification
// This is based on the Zod TypeScript implementation but uses golang-jwt for parsing
func isValidJWTStructure(token string, expectedAlgorithm *string) bool {
	// Use golang-jwt to parse the token without verification
	// This validates the structure and basic claims
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())

	// Parse without verification to check structure
	parsedToken, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return false
	}

	// Validate header
	if !validateJWTHeaderWithGolangJWT(parsedToken.Header, expectedAlgorithm) {
		return false
	}

	return true
}

// validateJWTHeaderWithGolangJWT validates JWT header using golang-jwt parsed data
func validateJWTHeaderWithGolangJWT(header map[string]interface{}, expectedAlgorithm *string) bool {
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

// =============================================================================
// EMOJI VALIDATION FUNCTIONS
// =============================================================================

// Emoji validates if string is valid emoji
func Emoji(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		return regexes.Emoji.MatchString(str)
	}
	return false
}

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
		datetimeRegex := regexes.Datetime(regexes.DatetimeOptions{
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
		timeRegex := regexes.Time(regexes.TimeOptions{
			Precision: options.Precision,
		})
		return timeRegex.MatchString(str)
	}
	return false
}

// Pre-compiled regex for ISOTime validation
var isoTimeRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9]([.,][0-9]+)?)?$`)

// Pre-compiled regex for ISODuration validation
var durationRegex = regexp.MustCompile(`^P(?:\d+W|(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:[.,]\d+)?S)?)?)$`)

// ISOTime validates if string is a valid ISO time format
func ISOTime(value any) bool {
	if str, ok := reflectx.ExtractString(value); ok {
		// Pattern:
		// HH:MM            (24-hour)
		// HH:MM:SS         (optional seconds)
		// HH:MM:SS.sss...  (optional fractional seconds, dot or comma)
		// This aligns with ISO-8601 and Zod TS implementation.
		return isoTimeRegex.MatchString(str)
	}
	return false
}

// ISODuration validates if string is a valid ISO 8601 duration format
func ISODuration(value any) bool {
	if duration, ok := reflectx.ExtractString(value); ok {
		// Use pre-compiled regex for performance

		// First check basic format
		if !durationRegex.MatchString(duration) {
			return false
		}

		// Reject empty patterns like "P" or "PT"
		if duration == "P" || duration == "PT" {
			return false
		}

		// Check for invalid patterns like "P1Y2MT" or "P1YT"
		// If there's a T, there must be at least one time component after it
		tIndex := strings.Index(duration, "T")
		if tIndex != -1 {
			// There's a T, check if there's at least one time component after it
			afterT := duration[tIndex+1:]
			if afterT == "" {
				return false // "PT" case
			}

			// Check if there's at least one valid time component (H, M, or S)
			hasTimeComponent := false
			for i, r := range afterT {
				if r >= '0' && r <= '9' {
					// Found a digit, check if it's followed by H, M, or S
					for j := i + 1; j < len(afterT); j++ {
						if afterT[j] == 'H' || afterT[j] == 'M' || afterT[j] == 'S' {
							hasTimeComponent = true
							break
						}
						if afterT[j] < '0' || afterT[j] > '9' {
							if afterT[j] != '.' && afterT[j] != ',' {
								break // Invalid character
							}
						}
					}
					if hasTimeComponent {
						break
					}
				}
			}
			if !hasTimeComponent {
				return false
			}
		}

		// Check if there's at least one digit after P or T
		hasDigit := false
		for i, r := range duration {
			if i > 0 && (r >= '0' && r <= '9') { // Skip the initial 'P'
				hasDigit = true
				break
			}
		}

		return hasDigit
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

// =============================================================================
// PROPERTY VALIDATION FUNCTIONS
// =============================================================================

// Property validates if object has a property that passes validation
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

// =============================================================================
// MIME TYPE VALIDATION FUNCTIONS
// =============================================================================

// Mime validates if value has allowed MIME types
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

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// toFloat64 converts any value to float64 using the coerce package
// Returns 0 if conversion fails
func toFloat64(value any) float64 {
	if result, err := coerce.ToFloat64(value); err == nil {
		return result
	}
	return 0
}

// =============================================================================
// URL VALIDATION WITH OPTIONS
// =============================================================================

// URLOptions defines options for URL validation
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
	if !regexes.URL.MatchString(str) {
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
