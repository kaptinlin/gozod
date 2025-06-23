// Package checks provides format validation checks
package checks

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/validate"
)

// =============================================================================
// EMAIL AND URL VALIDATION
// =============================================================================

// Email creates an email format validation check with JSON Schema support
// Supports: Email("invalid email") or Email(CheckParams{Error: "invalid email format"})
func Email(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "email"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "email")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// URL creates a URL format validation check with JSON Schema support
// Supports: URL("invalid URL") or URL(CheckParams{Error: "invalid URL format"})
func URL(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "url"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "uri")
				setBagProperty(schema, "type", "string")
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
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "uuid"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.UUID(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("uuid", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set uuid format for JSON Schema
				setBagProperty(schema, "format", "uuid")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// GUID creates a GUID format validation check with JSON Schema support
// Supports: GUID("invalid GUID") or GUID(CheckParams{Error: "invalid GUID format"})
func GUID(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "guid"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.GUID(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("guid", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set uuid format for JSON Schema (GUID is UUID variant)
				setBagProperty(schema, "format", "uuid")
				setBagProperty(schema, "type", "string")
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
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ipv4"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "ipv4")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// IPv6 creates an IPv6 address format validation check with JSON Schema support
// Supports: IPv6("invalid IPv6") or IPv6(CheckParams{Error: "invalid IPv6 address"})
func IPv6(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ipv6"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "ipv6")
				setBagProperty(schema, "type", "string")
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
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "cidrv4"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "cidrv4")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// CIDRv6 creates an IPv6 CIDR notation validation check with JSON Schema support
// Supports: CIDRv6("invalid CIDR") or CIDRv6(CheckParams{Error: "invalid IPv6 CIDR"})
func CIDRv6(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "cidrv6"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "cidrv6")
				setBagProperty(schema, "type", "string")
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
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "base64"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "base64")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// Base64URL creates a Base64URL encoding validation check with JSON Schema support
// Supports: Base64URL("invalid base64url") or Base64URL(CheckParams{Error: "invalid base64url encoding"})
func Base64URL(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "base64url"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "base64url")
				setBagProperty(schema, "type", "string")
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
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "jwt"}
	applyCheckParams(def, checkParams)

	return &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.JWT(payload.GetValue()) {
				payload.AddIssue(issues.CreateInvalidFormatIssue("jwt", payload.GetValue(), nil))
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set custom format for JSON Schema
				setBagProperty(schema, "format", "jwt")
				setBagProperty(schema, "type", "string")
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
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "e164"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "e164")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// =============================================================================
// DATE AND TIME VALIDATION
// =============================================================================

// ISODateTime creates an ISO 8601 datetime validation check with JSON Schema support
// Supports: ISODateTime("invalid datetime") or ISODateTime(CheckParams{Error: "invalid ISO datetime"})
func ISODateTime(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_datetime"}
	applyCheckParams(def, checkParams)

	var check *core.ZodCheckInternals
	check = &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISODateTime(payload.GetValue()) {
				iss := issues.CreateInvalidFormatIssue("iso_datetime", payload.GetValue(), nil)
				iss.Inst = check
				payload.AddIssue(iss)
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set datetime format for JSON Schema
				setBagProperty(schema, "format", "date-time")
				setBagProperty(schema, "type", "string")
			},
		},
	}
	return check
}

// ISODate creates an ISO 8601 date validation check with JSON Schema support
// Supports: ISODate("invalid date") or ISODate(CheckParams{Error: "invalid ISO date"})
func ISODate(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_date"}
	applyCheckParams(def, checkParams)

	var check *core.ZodCheckInternals
	check = &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISODate(payload.GetValue()) {
				iss := issues.CreateInvalidFormatIssue("iso_date", payload.GetValue(), nil)
				iss.Inst = check
				payload.AddIssue(iss)
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set date format for JSON Schema
				setBagProperty(schema, "format", "date")
				setBagProperty(schema, "type", "string")
			},
		},
	}
	return check
}

// ISOTime creates an ISO 8601 time validation check with JSON Schema support
// Supports: ISOTime("invalid time") or ISOTime(CheckParams{Error: "invalid ISO time"})
func ISOTime(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_time"}
	applyCheckParams(def, checkParams)

	var check *core.ZodCheckInternals
	check = &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISOTime(payload.GetValue()) {
				iss := issues.CreateInvalidFormatIssue("iso_time", payload.GetValue(), nil)
				iss.Inst = check
				payload.AddIssue(iss)
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set time format for JSON Schema
				setBagProperty(schema, "format", "time")
				setBagProperty(schema, "type", "string")
			},
		},
	}
	return check
}

// ISODuration creates an ISO 8601 duration validation check with JSON Schema support
// Supports: ISODuration("invalid duration") or ISODuration(CheckParams{Error: "invalid ISO duration"})
func ISODuration(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "iso_duration"}
	applyCheckParams(def, checkParams)

	var check *core.ZodCheckInternals
	check = &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			if !validate.ISODuration(payload.GetValue()) {
				iss := issues.CreateInvalidFormatIssue("iso_duration", payload.GetValue(), nil)
				iss.Inst = check
				payload.AddIssue(iss)
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set duration format for JSON Schema
				setBagProperty(schema, "format", "duration")
				setBagProperty(schema, "type", "string")
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
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "cuid"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "cuid")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// CUID2 creates a CUID2 format validation check with JSON Schema support
// Supports: CUID2("invalid CUID2") or CUID2(CheckParams{Error: "invalid CUID2 format"})
func CUID2(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "cuid2"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "cuid2")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// ULID creates a ULID format validation check with JSON Schema support
// Supports: ULID("invalid ULID") or ULID(CheckParams{Error: "invalid ULID format"})
func ULID(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ulid"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "ulid")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// XID creates an XID format validation check with JSON Schema support
// Supports: XID("invalid XID") or XID(CheckParams{Error: "invalid XID format"})
func XID(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "xid"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "xid")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// KSUID creates a KSUID format validation check with JSON Schema support
// Supports: KSUID("invalid KSUID") or KSUID(CheckParams{Error: "invalid KSUID format"})
func KSUID(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "ksuid"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "ksuid")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}

// NanoID creates a NanoID format validation check with JSON Schema support
// Supports: NanoID("invalid NanoID") or NanoID(CheckParams{Error: "invalid NanoID format"})
func NanoID(params ...any) core.ZodCheck {
	checkParams := normalizeCheckParams(params...)
	def := &core.ZodCheckDef{Check: "nanoid"}
	applyCheckParams(def, checkParams)

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
				setBagProperty(schema, "format", "nanoid")
				setBagProperty(schema, "type", "string")
			},
		},
	}
}
