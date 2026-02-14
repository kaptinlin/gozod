// Package checks provides validation check factories and utilities for creating
// reusable validation rules that can be attached to schemas.
package checks

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/regex"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// newFormatCheck creates a format validation check with standard JSON Schema
// annotations. Most format checks share the same structure: validate with a
// function, report an invalid_format issue on failure, and attach format +
// pattern metadata on schema attachment.
// This is an unexported helper function for internal use.
func newFormatCheck(
	checkID string,
	validateFn func(any) bool,
	format string,
	pattern *regexp.Regexp,
	params ...any,
) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: checkID}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validateFn(payload.Value()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue(checkID, payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", format)
				if pattern != nil {
					addPatternToSchema(schema, pattern.String())
				}
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Email creates an email format validation check.
func Email(params ...any) core.ZodCheck {
	return newFormatCheck("email", validate.Email, "email", regex.Email, params...)
}

// EmailWithPattern creates an email validation check with a custom regex pattern.
func EmailWithPattern(pattern *regexp.Regexp, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "email"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			str, ok := payload.Value().(string)
			if !ok {
				payload.AddIssue(issues.CreateInvalidTypeIssue(core.ZodTypeString, payload.Value()))
				return
			}
			if !pattern.MatchString(str) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("email", payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "email")
				addPatternToSchema(schema, pattern.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// HTML5Email creates an HTML5 email validation check.
func HTML5Email(params ...any) core.ZodCheck {
	return EmailWithPattern(regex.HTML5Email, params...)
}

// RFC5322Email creates an RFC 5322 email validation check.
func RFC5322Email(params ...any) core.ZodCheck {
	return EmailWithPattern(regex.RFC5322Email, params...)
}

// UnicodeEmail creates a Unicode email validation check.
func UnicodeEmail(params ...any) core.ZodCheck {
	return EmailWithPattern(regex.UnicodeEmail, params...)
}

// BrowserEmail creates a browser-compatible email validation check.
func BrowserEmail(params ...any) core.ZodCheck {
	return EmailWithPattern(regex.BrowserEmail, params...)
}

// URL creates a URL format validation check.
func URL(params ...any) core.ZodCheck {
	return newFormatCheck("url", validate.URL, "uri", regex.URL, params...)
}

// URLWithOptions creates a URL validation check with optional constraints.
func URLWithOptions(options validate.URLOptions, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "url"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.URLWithOptions(payload.Value(), options) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("url", payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "uri")
				addPatternToSchema(schema, regex.URL.String())
				SetBagProperty(schema, "type", "string")
				if options.Hostname != nil {
					SetBagProperty(schema, "hostnamePattern", options.Hostname.String())
				}
				if options.Protocol != nil {
					SetBagProperty(schema, "protocolPattern", options.Protocol.String())
				}
			},
		},
	}
}

// IPv4 creates an IPv4 address format validation check.
func IPv4(params ...any) core.ZodCheck {
	return newFormatCheck("ipv4", validate.IPv4, "ipv4", regex.IPv4, params...)
}

// IPv6 creates an IPv6 address format validation check.
func IPv6(params ...any) core.ZodCheck {
	return newFormatCheck("ipv6", validate.IPv6, "ipv6", regex.IPv6, params...)
}

// Hostname creates a DNS hostname validation check.
func Hostname(params ...any) core.ZodCheck {
	return newFormatCheck("hostname", validate.Hostname, "hostname", regex.Hostname, params...)
}

// MAC creates a MAC address validation check with default colon delimiter.
func MAC(params ...any) core.ZodCheck {
	return MACWithOptions(validate.MACOptions{Delimiter: ":"}, params...)
}

// MACWithDelimiter creates a MAC address validation check with a custom delimiter.
func MACWithDelimiter(delimiter string, params ...any) core.ZodCheck {
	return MACWithOptions(validate.MACOptions{Delimiter: delimiter}, params...)
}

// MACWithOptions creates a MAC address validation check with full configuration.
func MACWithOptions(options validate.MACOptions, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "mac"}
	ApplyCheckParams(def, cp)

	delim := options.Delimiter
	if delim == "" {
		delim = ":"
	}

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MACWithOptions(payload.Value(), options) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("mac", payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "mac")
				addPatternToSchema(schema, regex.MAC(delim).String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// CIDRv4 creates an IPv4 CIDR notation validation check.
func CIDRv4(params ...any) core.ZodCheck {
	return newFormatCheck("cidrv4", validate.CIDRv4, "cidrv4", regex.CIDRv4, params...)
}

// CIDRv6 creates an IPv6 CIDR notation validation check.
func CIDRv6(params ...any) core.ZodCheck {
	return newFormatCheck("cidrv6", validate.CIDRv6, "cidrv6", regex.CIDRv6, params...)
}

// Base64 creates a Base64 encoding validation check.
func Base64(params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "base64"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Base64(payload.Value()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("base64", payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "base64")
				SetBagProperty(schema, "contentEncoding", "base64")
				addPatternToSchema(schema, regex.Base64.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Base64URL creates a Base64URL encoding validation check.
func Base64URL(params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "base64url"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Base64URL(payload.Value()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("base64url", payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "base64url")
				SetBagProperty(schema, "contentEncoding", "base64url")
				addPatternToSchema(schema, regex.Base64URL.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// JWT creates a JWT token validation check.
func JWT(params ...any) core.ZodCheck {
	return JWTWithOptions(validate.JWTOptions{}, params...)
}

// JWTWithAlgorithm creates a JWT token validation check with algorithm constraint.
func JWTWithAlgorithm(algorithm string, params ...any) core.ZodCheck {
	return JWTWithOptions(validate.JWTOptions{Algorithm: &algorithm}, params...)
}

// JWTWithOptions creates a JWT token validation check with full configuration.
func JWTWithOptions(options validate.JWTOptions, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "jwt"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.JWTWithOptions(payload.Value(), options) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("jwt", payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "jwt")
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// E164 creates an E.164 phone number validation check.
func E164(params ...any) core.ZodCheck {
	return newFormatCheck("e164", validate.E164, "e164", regex.E164, params...)
}

// ISODateTimeWithOptions creates an ISO 8601 datetime validation check with options.
func ISODateTimeWithOptions(options validate.ISODateTimeOptions, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_datetime"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISODateTimeWithOptions(payload.Value(), options) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("iso_datetime", payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "iso_datetime")
				addPatternToSchema(schema, regex.DefaultDatetime.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// ISODateTime creates an ISO 8601 datetime format validation check.
func ISODateTime(params ...any) core.ZodCheck {
	return newFormatCheck("iso_datetime", validate.ISODateTime, "iso_datetime", regex.DefaultDatetime, params...)
}

// ISODate creates an ISO 8601 date format validation check.
func ISODate(params ...any) core.ZodCheck {
	return newFormatCheck("iso_date", validate.ISODate, "iso_date", regex.Date, params...)
}

// ISODateMin creates a minimum date validation check.
// Validates that the ISO date is on or after the specified minimum date.
func ISODateMin(minDate string, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "min_date"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if dateStr, ok := payload.Value().(string); ok && dateStr < minDate {
				issue := issues.CreateTooSmallIssue(minDate, true, "date", payload.Value())
				issue.Message = "Date must be on or after " + minDate
				payload.AddIssue(issue)
			}
		},
		OnAttach: []func(any){
			func(schema any) { SetBagProperty(schema, "minimum", minDate) },
		},
	}
}

// ISODateMax creates a maximum date validation check.
// Validates that the ISO date is on or before the specified maximum date.
func ISODateMax(maxDate string, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "max_date"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if dateStr, ok := payload.Value().(string); ok && dateStr > maxDate {
				issue := issues.CreateTooBigIssue(maxDate, true, "date", payload.Value())
				issue.Message = "Date must be on or before " + maxDate
				payload.AddIssue(issue)
			}
		},
		OnAttach: []func(any){
			func(schema any) { SetBagProperty(schema, "maximum", maxDate) },
		},
	}
}

// ISOTimeWithOptions creates an ISO 8601 time validation check with configuration.
// Uses self-reference to attach Inst on the created issue.
func ISOTimeWithOptions(options validate.ISOTimeOptions, params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_time"}
	ApplyCheckParams(def, cp)

	var check *core.ZodCheckInternals
	check = &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISOTimeWithOptions(payload.Value(), options) {
				iss := issues.CreateInvalidFormatIssue("iso_time", payload.Value(), nil)
				iss.Inst = check
				payload.AddIssue(iss)
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "iso_time")
				addPatternToSchema(schema, regex.DefaultTime.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
	return check
}

// ISOTime creates an ISO 8601 time validation check.
func ISOTime(params ...any) core.ZodCheck {
	return newFormatCheck("iso_time", validate.ISOTime, "iso_time", regex.DefaultTime, params...)
}

// ISODuration creates an ISO 8601 duration validation check.
func ISODuration(params ...any) core.ZodCheck {
	return newFormatCheck("iso_duration", validate.ISODuration, "iso_duration", regex.Duration, params...)
}

// CUID creates a CUID format validation check.
func CUID(params ...any) core.ZodCheck {
	return newFormatCheck("cuid", validate.CUID, "cuid", regex.CUID, params...)
}

// CUID2 creates a CUID2 format validation check.
func CUID2(params ...any) core.ZodCheck {
	return newFormatCheck("cuid2", validate.CUID2, "cuid2", regex.CUID2, params...)
}

// ULID creates a ULID format validation check.
func ULID(params ...any) core.ZodCheck {
	return newFormatCheck("ulid", validate.ULID, "ulid", regex.ULID, params...)
}

// XID creates an XID format validation check.
func XID(params ...any) core.ZodCheck {
	return newFormatCheck("xid", validate.XID, "xid", regex.XID, params...)
}

// KSUID creates a KSUID format validation check.
func KSUID(params ...any) core.ZodCheck {
	return newFormatCheck("ksuid", validate.KSUID, "ksuid", regex.KSUID, params...)
}

// NanoID creates a NanoID format validation check.
func NanoID(params ...any) core.ZodCheck {
	return newFormatCheck("nanoid", validate.NanoID, "nanoid", regex.NanoID, params...)
}

// JSON creates a JSON format validation check.
func JSON(params ...any) core.ZodCheck {
	cp := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "json"}
	ApplyCheckParams(def, cp)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.JSON(payload.Value()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("json", payload.Value(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "contentMediaType", "application/json")
				addPatternToSchema(schema, regex.JSONString.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Emoji creates an emoji validation check.
func Emoji(params ...any) core.ZodCheck {
	return newFormatCheck("emoji", func(v any) bool {
		return validate.Regex(v, regex.Emoji)
	}, "emoji", regex.Emoji, params...)
}

// UUID creates a UUID format validation check.
func UUID(params ...any) core.ZodCheck { return buildUUIDCheck("uuid", regex.UUID, params...) }

// GUID creates a GUID format validation check.
func GUID(params ...any) core.ZodCheck { return buildUUIDCheck("guid", regex.GUID, params...) }

// UUIDv4 creates a UUID version 4 format validation check.
func UUIDv4(params ...any) core.ZodCheck { return buildUUIDCheck("uuidv4", regex.UUID4, params...) }

// UUID6 creates a UUID v6 format validation check.
func UUID6(params ...any) core.ZodCheck { return buildUUIDCheck("uuid6", regex.UUID6, params...) }

// UUID7 creates a UUID v7 format validation check.
func UUID7(params ...any) core.ZodCheck { return buildUUIDCheck("uuid7", regex.UUID7, params...) }

// buildUUIDCheck constructs UUID-related checks. All UUID variants use "uuid"
// as the JSON Schema format annotation.
func buildUUIDCheck(checkID string, pattern *regexp.Regexp, params ...any) core.ZodCheck {
	return newFormatCheck(checkID, func(v any) bool {
		return validate.Regex(v, pattern)
	}, "uuid", pattern, params...)
}

// Hex creates a hexadecimal string validation check.
func Hex(params ...any) core.ZodCheck {
	return newFormatCheck("hex", validate.Hex, "hex", regex.Hex, params...)
}

// newHashCheck creates a hash validation check for a specific algorithm.
func newHashCheck(checkID string, validateFn func(any) bool, pattern *regexp.Regexp, params ...any) core.ZodCheck {
	return newFormatCheck(checkID, validateFn, checkID, pattern, params...)
}

// MD5 creates an MD5 hash validation check (32 hex chars).
func MD5(params ...any) core.ZodCheck {
	return newHashCheck("md5", validate.MD5Hex, regex.MD5Hex, params...)
}

// SHA1 creates a SHA-1 hash validation check (40 hex chars).
func SHA1(params ...any) core.ZodCheck {
	return newHashCheck("sha1", validate.SHA1Hex, regex.SHA1Hex, params...)
}

// SHA256 creates a SHA-256 hash validation check (64 hex chars).
func SHA256(params ...any) core.ZodCheck {
	return newHashCheck("sha256", validate.SHA256Hex, regex.SHA256Hex, params...)
}

// SHA384 creates a SHA-384 hash validation check (96 hex chars).
func SHA384(params ...any) core.ZodCheck {
	return newHashCheck("sha384", validate.SHA384Hex, regex.SHA384Hex, params...)
}

// SHA512 creates a SHA-512 hash validation check (128 hex chars).
func SHA512(params ...any) core.ZodCheck {
	return newHashCheck("sha512", validate.SHA512Hex, regex.SHA512Hex, params...)
}
