// Package checks provides format validation checks
package checks

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/regex"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// =============================================================================
// EMAIL AND URL VALIDATION
// =============================================================================

// Email creates an email format validation check with JSON Schema support
// Supports: Email("invalid email") or Email(CheckParams{Error: "invalid email format"})
func Email(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "email"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Email(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("email", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set email format for JSON Schema
				SetBagProperty(schema, "format", "email")
				addPatternToSchema(schema, regex.Email.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// EmailWithPattern creates an email validation check with custom regex pattern
func EmailWithPattern(pattern *regexp.Regexp, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "email"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			str, ok := payload.GetValue().(string)
			if !ok {
				payload.AddIssue(issues.CreateInvalidTypeIssue(core.ZodTypeString, payload.GetValue()))
				return
			}
			if !pattern.MatchString(str) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("email", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set email format for JSON Schema
				SetBagProperty(schema, "format", "email")
				addPatternToSchema(schema, pattern.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Html5Email creates an HTML5 email validation check
func Html5Email(params ...any) core.ZodCheck {
	return EmailWithPattern(regex.Html5Email, params...)
}

// Rfc5322Email creates an RFC5322 email validation check
func Rfc5322Email(params ...any) core.ZodCheck {
	return EmailWithPattern(regex.Rfc5322Email, params...)
}

// UnicodeEmail creates a Unicode email validation check
func UnicodeEmail(params ...any) core.ZodCheck {
	return EmailWithPattern(regex.UnicodeEmail, params...)
}

// BrowserEmail creates a browser-compatible email validation check
func BrowserEmail(params ...any) core.ZodCheck {
	return EmailWithPattern(regex.BrowserEmail, params...)
}

// URL creates a URL format validation check with JSON Schema support
// Supports: URL("invalid URL") or URL(CheckParams{Error: "invalid URL format"})
func URL(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "url"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.URL(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("url", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set uri format for JSON Schema
				SetBagProperty(schema, "format", "uri")
				addPatternToSchema(schema, regex.URL.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// URLWithOptions creates a URL validation check with optional constraints
// Supports: URLWithOptions(validate.URLOptions{Hostname: hostnameRegex}, "invalid URL")
func URLWithOptions(options validate.URLOptions, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "url"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.URLWithOptions(payload.GetValue(), options) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("url", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set uri format for JSON Schema
				SetBagProperty(schema, "format", "uri")
				addPatternToSchema(schema, regex.URL.String())
				SetBagProperty(schema, "type", "string")

				// Add constraint information to schema
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

// =============================================================================
// IP ADDRESS VALIDATION
// =============================================================================

// IPv4 creates an IPv4 address format validation check with JSON Schema support
// Supports: IPv4("invalid IPv4") or IPv4(CheckParams{Error: "invalid IPv4 address"})
func IPv4(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ipv4"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.IPv4(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("ipv4", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ipv4 format for JSON Schema
				SetBagProperty(schema, "format", "ipv4")
				addPatternToSchema(schema, regex.IPv4.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// IPv6 creates an IPv6 address format validation check with JSON Schema support
// Supports: IPv6("invalid IPv6") or IPv6(CheckParams{Error: "invalid IPv6 address"})
func IPv6(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ipv6"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.IPv6(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("ipv6", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ipv6 format for JSON Schema
				SetBagProperty(schema, "format", "ipv6")
				addPatternToSchema(schema, regex.IPv6.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// HOSTNAME VALIDATION
// =============================================================================

// Hostname creates a DNS hostname validation check with JSON Schema support.
// Supports: Hostname("invalid hostname") or Hostname(CheckParams{Error: "invalid hostname"})
func Hostname(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "hostname"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Hostname(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("hostname", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "hostname")
				addPatternToSchema(schema, regex.Hostname.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// MAC ADDRESS VALIDATION
// =============================================================================

// MAC creates a MAC address validation check with JSON Schema support.
// Uses default colon (":") delimiter.
// Supports: MAC("invalid MAC") or MAC(CheckParams{Error: "invalid MAC address"})
func MAC(params ...any) core.ZodCheck {
	return MACWithOptions(validate.MACOptions{Delimiter: ":"}, params...)
}

// MACWithDelimiter creates a MAC address validation check with custom delimiter.
// Common delimiters: ":", "-", "."
// Supports: MACWithDelimiter("-", "invalid MAC") or MACWithDelimiter(".", CheckParams{Error: "invalid MAC"})
func MACWithDelimiter(delimiter string, params ...any) core.ZodCheck {
	return MACWithOptions(validate.MACOptions{Delimiter: delimiter}, params...)
}

// MACWithOptions creates a MAC address validation check with full configuration options.
// This implementation aligns with Zod's TypeScript version, using case-sensitive matching.
// Supports: MACWithOptions(validate.MACOptions{Delimiter: "-"}, "invalid MAC")
//
// Example validations:
//
//	"00:1A:2B:3C:4D:5E" with delimiter ":" -> valid
//	"00-1a-2b-3c-4d-5e" with delimiter "-" -> valid
//	"00.1A.2B.3C.4D.5E" with delimiter "." -> valid
func MACWithOptions(options validate.MACOptions, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "mac"}
	ApplyCheckParams(def, checkParams)

	// Determine delimiter for schema metadata
	delim := options.Delimiter
	if delim == "" {
		delim = ":"
	}

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MACWithOptions(payload.GetValue(), options) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("mac", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set custom format for JSON Schema
				SetBagProperty(schema, "format", "mac")
				// Use the specific delimiter's regex pattern
				addPatternToSchema(schema, regex.MAC(delim).String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// CIDR NOTATION VALIDATION
// =============================================================================

// CIDRv4 creates an IPv4 CIDR notation validation check with JSON Schema support
// Supports: CIDRv4("invalid CIDR") or CIDRv4(CheckParams{Error: "invalid IPv4 CIDR"})
func CIDRv4(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "cidrv4"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.CIDRv4(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("cidrv4", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set custom format for JSON Schema
				SetBagProperty(schema, "format", "cidrv4")
				addPatternToSchema(schema, regex.CIDRv4.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// CIDRv6 creates an IPv6 CIDR notation validation check with JSON Schema support
// Supports: CIDRv6("invalid CIDR") or CIDRv6(CheckParams{Error: "invalid IPv6 CIDR"})
func CIDRv6(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "cidrv6"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.CIDRv6(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("cidrv6", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set custom format for JSON Schema
				SetBagProperty(schema, "format", "cidrv6")
				addPatternToSchema(schema, regex.CIDRv6.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// ENCODING VALIDATION
// =============================================================================

// Base64 creates a Base64 encoding validation check with JSON Schema support
// Supports: Base64("invalid base64") or Base64(CheckParams{Error: "invalid base64 encoding"})
func Base64(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "base64"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Base64(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("base64", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set custom format for JSON Schema
				SetBagProperty(schema, "format", "base64")
				SetBagProperty(schema, "contentEncoding", "base64")
				addPatternToSchema(schema, regex.Base64.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// Base64URL creates a Base64URL encoding validation check with JSON Schema support
// Supports: Base64URL("invalid base64url") or Base64URL(CheckParams{Error: "invalid base64url encoding"})
func Base64URL(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "base64url"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Base64URL(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("base64url", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set custom format for JSON Schema
				SetBagProperty(schema, "format", "base64url")
				SetBagProperty(schema, "contentEncoding", "base64url")
				addPatternToSchema(schema, regex.Base64URL.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// TOKEN AND AUTHENTICATION VALIDATION
// =============================================================================

// JWT creates a JWT token validation check with JSON Schema support
// Supports: JWT("invalid JWT") or JWT(CheckParams{Error: "invalid JWT token"})
func JWT(params ...any) core.ZodCheck {
	return JWTWithOptions(validate.JWTOptions{}, params...)
}

// JWTWithAlgorithm creates a JWT token validation check with algorithm constraint
// Supports: JWTWithAlgorithm("HS256", "invalid JWT") or JWTWithAlgorithm("RS256", CheckParams{Error: "invalid JWT token"})
func JWTWithAlgorithm(algorithm string, params ...any) core.ZodCheck {
	return JWTWithOptions(validate.JWTOptions{Algorithm: &algorithm}, params...)
}

// JWTWithOptions creates a JWT token validation check with full configuration options
func JWTWithOptions(options validate.JWTOptions, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "jwt"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.JWTWithOptions(payload.GetValue(), options) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("jwt", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set custom format for JSON Schema
				SetBagProperty(schema, "format", "jwt")
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// PHONE NUMBER VALIDATION
// =============================================================================

// E164 creates an E.164 phone number validation check with JSON Schema support
// Supports: E164("invalid phone") or E164(CheckParams{Error: "invalid E.164 format"})
func E164(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "e164"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.E164(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("e164", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set custom format for JSON Schema
				SetBagProperty(schema, "format", "e164")
				addPatternToSchema(schema, regex.E164.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// DATE AND TIME VALIDATION
// =============================================================================

// ISODateTimeWithOptions creates an ISO 8601 datetime validation check with options
// Supports: ISODateTimeWithOptions(options, "invalid datetime") or ISODateTimeWithOptions(options, CheckParams{Error: "invalid datetime"})
func ISODateTimeWithOptions(options validate.ISODateTimeOptions, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_datetime"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISODateTimeWithOptions(payload.GetValue(), options) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("iso_datetime", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ISO datetime format for JSON Schema (custom id)
				SetBagProperty(schema, "format", "iso_datetime")
				addPatternToSchema(schema, regex.DefaultDatetime.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// ISODateTime creates an ISO 8601 datetime format validation check with JSON Schema support
// Supports: ISODateTime("invalid datetime") or ISODateTime(CheckParams{Error: "invalid datetime format"})
func ISODateTime(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_datetime"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISODateTime(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("iso_datetime", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ISO datetime format for JSON Schema (custom id)
				SetBagProperty(schema, "format", "iso_datetime")
				addPatternToSchema(schema, regex.DefaultDatetime.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// ISODate creates an ISO 8601 date format validation check with JSON Schema support
// Supports: ISODate("invalid date") or ISODate(CheckParams{Error: "invalid date format"})
func ISODate(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_date"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISODate(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("iso_date", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ISO date format for JSON Schema
				SetBagProperty(schema, "format", "iso_date")
				addPatternToSchema(schema, regex.Date.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// ISODateMin creates a minimum date validation check
// Validates that the ISO date is on or after the specified minimum date
func ISODateMin(minDate string, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "min_date"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			value := payload.GetValue()
			if dateStr, ok := value.(string); ok {
				if dateStr < minDate {
					issue := issues.CreateTooSmallIssue(minDate, true, "date", payload.GetValue())
					issue.Message = "Date must be on or after " + minDate
					payload.AddIssue(issue)
				}
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "minimum", minDate)
			},
		},
	}
}

// ISODateMax creates a maximum date validation check
// Validates that the ISO date is on or before the specified maximum date
func ISODateMax(maxDate string, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "max_date"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			value := payload.GetValue()
			if dateStr, ok := value.(string); ok {
				if dateStr > maxDate {
					issue := issues.CreateTooBigIssue(maxDate, true, "date", payload.GetValue())
					issue.Message = "Date must be on or before " + maxDate
					payload.AddIssue(issue)
				}
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "maximum", maxDate)
			},
		},
	}
}

// ISOTimeWithOptions creates an ISO 8601 time validation check with configuration options
// Supports: ISOTimeWithOptions(options, "invalid time") or ISOTimeWithOptions(options, CheckParams{Error: "invalid ISO time"})
func ISOTimeWithOptions(options validate.ISOTimeOptions, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_time"}
	ApplyCheckParams(def, checkParams)

	var check *core.ZodCheckInternals
	check = &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISOTimeWithOptions(payload.GetValue(), options) {
				iss := issues.CreateInvalidFormatIssue("iso_time", payload.GetValue(), nil)
				iss.Inst = check
				payload.AddIssue(iss)
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ISO time format for JSON Schema
				SetBagProperty(schema, "format", "iso_time")
				addPatternToSchema(schema, regex.DefaultTime.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
	return check
}

// ISOTime creates an ISO 8601 time validation check with JSON Schema support
// Supports: ISOTime("invalid time") or ISOTime(CheckParams{Error: "invalid ISO time"})
func ISOTime(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_time"}
	ApplyCheckParams(def, checkParams)

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISOTime(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("iso_time", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ISO time format for JSON Schema
				SetBagProperty(schema, "format", "iso_time")
				addPatternToSchema(schema, regex.DefaultTime.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
	return check
}

// ISODuration creates an ISO 8601 duration validation check with JSON Schema support
// Supports: ISODuration("invalid duration") or ISODuration(CheckParams{Error: "invalid ISO duration"})
func ISODuration(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_duration"}
	ApplyCheckParams(def, checkParams)

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISODuration(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("iso_duration", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ISO duration format for JSON Schema
				SetBagProperty(schema, "format", "iso_duration")
				addPatternToSchema(schema, regex.Duration.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
	return check
}

// =============================================================================
// UNIQUE IDENTIFIER VALIDATION
// =============================================================================

// CUID creates a CUID format validation check with JSON Schema support
// Supports: CUID("invalid CUID") or CUID(CheckParams{Error: "invalid CUID format"})
func CUID(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "cuid"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.CUID(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("cuid", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set cuid format for JSON Schema
				SetBagProperty(schema, "format", "cuid")
				addPatternToSchema(schema, regex.CUID.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// CUID2 creates a CUID2 format validation check with JSON Schema support
// Supports: CUID2("invalid CUID2") or CUID2(CheckParams{Error: "invalid CUID2 format"})
func CUID2(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "cuid2"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.CUID2(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("cuid2", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set cuid2 format for JSON Schema
				SetBagProperty(schema, "format", "cuid2")
				addPatternToSchema(schema, regex.CUID2.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// ULID creates a ULID format validation check with JSON Schema support
// Supports: ULID("invalid ULID") or ULID(CheckParams{Error: "invalid ULID format"})
func ULID(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ulid"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ULID(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("ulid", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ulid format for JSON Schema
				SetBagProperty(schema, "format", "ulid")
				addPatternToSchema(schema, regex.ULID.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// XID creates an XID format validation check with JSON Schema support
// Supports: XID("invalid XID") or XID(CheckParams{Error: "invalid XID format"})
func XID(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "xid"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.XID(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("xid", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set xid format for JSON Schema
				SetBagProperty(schema, "format", "xid")
				addPatternToSchema(schema, regex.XID.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// KSUID creates a KSUID format validation check with JSON Schema support
// Supports: KSUID("invalid KSUID") or KSUID(CheckParams{Error: "invalid KSUID format"})
func KSUID(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ksuid"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.KSUID(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("ksuid", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set ksuid format for JSON Schema
				SetBagProperty(schema, "format", "ksuid")
				addPatternToSchema(schema, regex.KSUID.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// NanoID creates a NanoID format validation check with JSON Schema support
// Supports: NanoID("invalid NanoID") or NanoID(CheckParams{Error: "invalid NanoID format"})
func NanoID(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "nanoid"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.NanoID(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("nanoid", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set nanoid format for JSON Schema
				SetBagProperty(schema, "format", "nanoid")
				addPatternToSchema(schema, regex.NanoID.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// JSON VALIDATION
// =============================================================================

// JSON creates a JSON format validation check with JSON Schema support
// Supports: JSON("invalid JSON") or JSON(CheckParams{Error: "invalid JSON"})
func JSON(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "json"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.JSON(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("json", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set json format for JSON Schema
				SetBagProperty(schema, "contentMediaType", "application/json")
				addPatternToSchema(schema, regex.JSONString.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// EMOJI VALIDATION
// =============================================================================

// Emoji creates an emoji validation check with JSON Schema support.
// Supports: Emoji("invalid emoji") or Emoji(CheckParams{Error: "invalid emoji"})
func Emoji(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "emoji"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Regex(payload.GetValue(), regex.Emoji) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("emoji", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Add pattern info for JSON Schema generation
				SetBagProperty(schema, "format", "emoji")
				addPatternToSchema(schema, regex.Emoji.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// UUID AND GUID VALIDATION
// =============================================================================

// UUID creates a UUID format validation check with JSON Schema support
// Supports: UUID("invalid UUID") or UUID(CheckParams{Error: "invalid UUID format"})
func UUID(params ...any) core.ZodCheck {
	return buildUUIDCheck("uuid", regex.UUID, params...)
}

// GUID creates a GUID format validation check with JSON Schema support
// Supports: GUID("invalid GUID") or GUID(CheckParams{Error: "invalid GUID format"})
func GUID(params ...any) core.ZodCheck {
	return buildUUIDCheck("guid", regex.GUID, params...)
}

// =============================================================================
// UUID VERSION-SPECIFIC VALIDATION
// =============================================================================

// UUIDv4 creates a UUID version 4 format validation check.
// Supports: UUIDv4("invalid UUIDv4") or UUIDv4(CheckParams{Error: "invalid UUIDv4"})
func UUIDv4(params ...any) core.ZodCheck {
	return buildUUIDCheck("uuidv4", regex.UUID4, params...)
}

// UUID6 creates a UUID v6 format validation check.
// Supports: UUID6("invalid UUIDv6") or UUID6(CheckParams{Error: "invalid UUIDv6"})
func UUID6(params ...any) core.ZodCheck {
	return buildUUIDCheck("uuid6", regex.UUID6, params...)
}

// UUID7 creates a UUID v7 format validation check.
// Supports: UUID7("invalid UUIDv7") or UUID7(CheckParams{Error: "invalid UUIDv7"})
func UUID7(params ...any) core.ZodCheck {
	return buildUUIDCheck("uuid7", regex.UUID7, params...)
}

// buildUUIDCheck constructs UUID-related checks with appropriate format annotation.
func buildUUIDCheck(checkID string, pattern *regexp.Regexp, params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: checkID}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Regex(payload.GetValue(), pattern) {
				payload.AddIssue(issues.CreateInvalidFormatIssue(checkID, payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set format - all UUID variants use "uuid" format
				SetBagProperty(schema, "format", "uuid")
				addPatternToSchema(schema, pattern.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// HEX VALIDATION
// =============================================================================

// Hex creates a hexadecimal string validation check
// Supports: Hex("invalid hex") or Hex(CheckParams{Error: "invalid hex"})
func Hex(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "hex"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.Hex(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("hex", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "hex")
				addPatternToSchema(schema, regex.Hex.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// HASH VALIDATION
// =============================================================================

// MD5 creates an MD5 hash validation check (32 hex chars)
func MD5(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "md5"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.MD5Hex(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("md5", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "md5")
				addPatternToSchema(schema, regex.MD5Hex.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// SHA1 creates a SHA-1 hash validation check (40 hex chars)
func SHA1(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "sha1"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.SHA1Hex(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("sha1", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "sha1")
				addPatternToSchema(schema, regex.SHA1Hex.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// SHA256 creates a SHA-256 hash validation check (64 hex chars)
func SHA256(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "sha256"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.SHA256Hex(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("sha256", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "sha256")
				addPatternToSchema(schema, regex.SHA256Hex.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// SHA384 creates a SHA-384 hash validation check (96 hex chars)
func SHA384(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "sha384"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.SHA384Hex(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("sha384", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "sha384")
				addPatternToSchema(schema, regex.SHA384Hex.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}

// SHA512 creates a SHA-512 hash validation check (128 hex chars)
func SHA512(params ...any) core.ZodCheck {
	checkParams := NormalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "sha512"}
	ApplyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.SHA512Hex(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("sha512", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				SetBagProperty(schema, "format", "sha512")
				addPatternToSchema(schema, regex.SHA512Hex.String())
				SetBagProperty(schema, "type", "string")
			},
		},
	}
}
