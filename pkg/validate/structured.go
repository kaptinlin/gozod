package validate

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/golang-jwt/jwt/v5"

	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/regex"
)

// Pre-compiled regex patterns for ISO time and duration validation.
var (
	isoTimeRegex  = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9]([.,][0-9]+)?)?$`)
	durationRegex = regexp.MustCompile(`^P(?:\d+W|(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:[.,]\d+)?S)?)?)$`)
)

// JWT reports whether the string is a valid JWT format.
func JWT(value any) bool {
	str, ok := reflectx.StringVal(value)
	if !ok {
		return false
	}
	return isValidJWT(str, nil)
}

// JWTWithOptions reports whether the string is a valid JWT format with an algorithm constraint.
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
	if duration == "P" || duration == "PT" {
		return false
	}
	if _, afterT, found := strings.Cut(duration, "T"); found {
		if !strings.ContainsAny(afterT, "HMS") {
			return false
		}
	}
	return strings.ContainsAny(duration[1:], "0123456789")
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
	if validator == nil {
		return false
	}
	m, ok := obj.(map[string]any)
	if !ok {
		return false
	}
	value, exists := m[key]
	if !exists {
		return false
	}
	return validator(value)
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
	if typStr, ok := header["typ"].(string); ok && typStr != "JWT" {
		return false
	}
	algStr, ok := header["alg"].(string)
	if !ok || algStr == "none" {
		return false
	}
	return expectedAlgorithm == nil || algStr == *expectedAlgorithm
}
