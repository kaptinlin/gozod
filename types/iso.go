package types

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

//////////////////////////////////////////
//////////////////////////////////////////
//////////                      //////////
//////////      ISO TYPES       //////////
//////////                      //////////
//////////////////////////////////////////
//////////////////////////////////////////

// ZodISO provides access to ISO format validators
type ZodISO struct{}

// ISO provides access to ISO format validators
var ISO = ZodISO{}

//////////////////////////////////////////
//////////   ISO Date Type       ////////
//////////////////////////////////////////

// ZodISODateDef defines the configuration for ISO date validation
type ZodISODateDef struct {
	core.ZodTypeDef
	Type   string          // "string"
	Checks []core.ZodCheck // ISO date-specific validation checks
}

// ZodISODateInternals contains ISO date internal state
type ZodISODateInternals struct {
	core.ZodTypeInternals
	Def     *ZodISODateDef             // Schema definition
	Checks  []core.ZodCheck            // Validation checks
	Isst    issues.ZodIssueInvalidType // Invalid type issue template
	Pattern *regexp.Regexp             // Regex pattern (if any)
	Values  map[string]struct{}        // Allowed string values set
	Bag     map[string]any             // Additional metadata
}

// ZodISODate represents an ISO date schema with validation support
type ZodISODate struct {
	internals *ZodISODateInternals
}

// GetInternals implements ZodType interface
func (z *ZodISODate) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// CloneFrom implements Cloneable interface for proper state copying in AddCheck
func (z *ZodISODate) CloneFrom(source any) {
	if src, ok := source.(*ZodISODate); ok {
		srcState := src.internals
		tgtState := z.internals

		// Copy Bag state (including coercion flag)
		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]any)
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}

		// Copy state from core.ZodTypeInternals.Bag (used by parseType)
		if len(srcState.ZodTypeInternals.Bag) > 0 {
			if tgtState.ZodTypeInternals.Bag == nil {
				tgtState.ZodTypeInternals.Bag = make(map[string]any)
			}
			for key, value := range srcState.ZodTypeInternals.Bag {
				tgtState.ZodTypeInternals.Bag[key] = value
			}
		}

		// Copy Values state
		if len(srcState.Values) > 0 {
			if tgtState.Values == nil {
				tgtState.Values = make(map[string]struct{})
			}
			for key, value := range srcState.Values {
				tgtState.Values[key] = value
			}
		}

		// Copy Pattern state
		if srcState.Pattern != nil {
			tgtState.Pattern = srcState.Pattern
		}
	}
}

// Parse implements ZodType interface with smart type inference
func (z *ZodISODate) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	} else {
		parseCtx = core.NewParseContext()
	}

	return engine.ParseType(
		input,
		&z.internals.ZodTypeInternals,
		"string",
		reflectx.ExtractString,
		func(v any) (*string, bool) {
			if reflectx.IsPointer(v) {
				if deref, ok := reflectx.Deref(v); ok {
					if str, ok := deref.(string); ok {
						return &str, true
					}
				}
			}
			return nil, false
		},
		func(value string, checks []core.ZodCheck, ctx *core.ParseContext) error {
			// Run all checks including date format validation
			if len(checks) > 0 {
				checkPayload := &core.ParsePayload{
					Value:  value,
					Issues: make([]core.ZodRawIssue, 0),
				}
				engine.RunChecksOnValue(value, checks, checkPayload, ctx)
				if len(checkPayload.Issues) > 0 {
					return &issues.ZodError{Issues: issues.ConvertRawIssuesToIssues(checkPayload.Issues, ctx)}
				}
			}
			return nil
		},
		func(v any) (string, bool) {
			// Use pkg/coerce for type coercion
			if result, err := coerce.ToISODate(v); err == nil {
				return result, true
			}
			return "", false
		},
		parseCtx,
	)
}

// MustParse implements ZodType interface
func (z *ZodISODate) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Coerce implements Coercible interface for ISO date coercion
func (z *ZodISODate) Coerce(input any) (any, bool) {
	// Use pkg/coerce instead of custom coerceToISODate
	if result, err := coerce.ToISODate(input); err == nil {
		return result, true
	}
	return nil, false
}

// Min adds minimum date validation using existing checks system
func (z *ZodISODate) Min(minDate string, params ...any) *ZodISODate {
	// Parse the date to validate it's a valid ISO date
	if _, err := time.Parse("2006-01-02", minDate); err != nil {
		panic("Invalid ISO date format for Min: " + minDate)
	}

	// Extract custom error message if provided in params
	var message string
	if len(params) > 0 {
		switch p := params[0].(type) {
		case string:
			message = p
		case core.SchemaParams:
			if p.Error != nil {
				if msg, ok := p.Error.(string); ok {
					message = msg
				}
			}
		}
	}

	// Custom min-date check that:
	// 1) Runs only when input already passes ISO format.
	// 2) Emits "too_small" code on violation (consistent with Zod TS).
	check := &core.ZodCheckInternals{
		Def: &core.ZodCheckDef{Check: "iso_date_min"},
	}
	check.Check = func(payload *core.ParsePayload) {
		str, ok := reflectx.ExtractString(payload.Value)
		if !ok {
			return // non-string handled by type check earlier
		}

		// Skip when format invalid – format check already appended an issue
		if _, err := time.Parse("2006-01-02", str); err != nil {
			return
		}

		inputT, _ := time.Parse("2006-01-02", str)
		minT, _ := time.Parse("2006-01-02", minDate)
		if inputT.Before(minT) {
			iss := issues.CreateTooSmallIssue(minDate, true, "string", str)
			if message != "" {
				iss.Message = message
			}
			iss.Inst = check
			payload.Issues = append(payload.Issues, iss)
		}
	}

	// Use AddCheck to properly add the check and maintain state
	result := engine.AddCheck(z, check).(*ZodISODate)
	// store min date in bag for getters (both internal bags for consistency)
	result.internals.Bag["minimum"] = minDate
	if result.internals.ZodTypeInternals.Bag == nil {
		result.internals.ZodTypeInternals.Bag = make(map[string]any)
	}
	result.internals.ZodTypeInternals.Bag["minimum"] = minDate
	return result
}

// Max adds maximum date validation using existing checks system
func (z *ZodISODate) Max(maxDate string, params ...any) *ZodISODate {
	// Parse the date to validate it's a valid ISO date
	if _, err := time.Parse("2006-01-02", maxDate); err != nil {
		panic("Invalid ISO date format for Max: " + maxDate)
	}

	// Extract custom error message if provided in params
	var message string
	if len(params) > 0 {
		switch p := params[0].(type) {
		case string:
			message = p
		case core.SchemaParams:
			if p.Error != nil {
				if msg, ok := p.Error.(string); ok {
					message = msg
				}
			}
		}
	}

	check := &core.ZodCheckInternals{
		Def: &core.ZodCheckDef{Check: "iso_date_max"},
	}
	check.Check = func(payload *core.ParsePayload) {
		str, ok := reflectx.ExtractString(payload.Value)
		if !ok {
			return
		}
		if _, err := time.Parse("2006-01-02", str); err != nil {
			return
		}
		inputT, _ := time.Parse("2006-01-02", str)
		maxT, _ := time.Parse("2006-01-02", maxDate)
		if inputT.After(maxT) {
			iss := issues.CreateTooBigIssue(maxDate, true, "string", str)
			if message != "" {
				iss.Message = message
			}
			iss.Inst = check
			payload.Issues = append(payload.Issues, iss)
		}
	}

	// Use AddCheck to properly add the check and maintain state
	result := engine.AddCheck(z, check).(*ZodISODate)
	result.internals.Bag["maximum"] = maxDate
	if result.internals.ZodTypeInternals.Bag == nil {
		result.internals.ZodTypeInternals.Bag = make(map[string]any)
	}
	result.internals.ZodTypeInternals.Bag["maximum"] = maxDate
	return result
}

// MinDate returns the minimum date constraint if set
func (z *ZodISODate) MinDate() string {
	if z.internals.ZodTypeInternals.Bag != nil {
		if minValue, exists := z.internals.ZodTypeInternals.Bag["minimum"]; exists {
			if minStr, ok := minValue.(string); ok {
				return minStr
			}
		}
	}
	return ""
}

// MaxDate returns the maximum date constraint if set
func (z *ZodISODate) MaxDate() string {
	if z.internals.ZodTypeInternals.Bag != nil {
		if maxValue, exists := z.internals.ZodTypeInternals.Bag["maximum"]; exists {
			if maxStr, ok := maxValue.(string); ok {
				return maxStr
			}
		}
	}
	return ""
}

// RefineAny adds custom validation with any input type
func (z *ZodISODate) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	result := engine.AddCheck(z, check)
	return result
}

// TransformAny adds data transformation with any input/output types
func (z *ZodISODate) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Optional makes the schema optional
func (z *ZodISODate) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the schema accept nil values
func (z *ZodISODate) Nilable() core.ZodType[any, any] {
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Pipe implements ZodType interface
func (z *ZodISODate) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////////////////////
//////////   ISO Date Options   //////////
//////////////////////////////////////////

// ISODateOptions defines parameters for ISO date validation
type ISODateOptions struct {
	Error  any  // core.Custom error message
	Coerce bool // Enable coercion
}

//////////////////////////////////////////
//////////   ISO Time Options   //////////
//////////////////////////////////////////

// ISOTimeOptions defines parameters for ISO time validation
type ISOTimeOptions struct {
	// Precision specifies number of decimal places for seconds
	// If nil, matches any number of decimal places
	// If 0 or negative, no decimal places allowed
	Precision *int `json:"precision,omitempty"`
	Error     any  `json:"error,omitempty"`  // core.Custom error message
	Coerce    bool `json:"coerce,omitempty"` // Enable coercion
}

//////////////////////////////////////////
//////////  ISO DateTime Options ////////
//////////////////////////////////////////

// ISODateTimeOptions defines parameters for ISO datetime validation
type ISODateTimeOptions struct {
	// Precision specifies number of decimal places for seconds
	// If nil, matches any number of decimal places
	// If 0 or negative, no decimal places allowed
	Precision *int `json:"precision,omitempty"`

	// Offset if true, allows timezone offsets like +01:00
	Offset bool `json:"offset,omitempty"`

	// Local if true, makes the 'Z' timezone marker optional
	Local bool `json:"local,omitempty"`

	Error  any  `json:"error,omitempty"`  // core.Custom error message
	Coerce bool `json:"coerce,omitempty"` // Enable coercion
}

//////////////////////////////////////////
//////////  ISO Duration Options ////////
//////////////////////////////////////////

// ISODurationOptions defines parameters for ISO duration validation
type ISODurationOptions struct {
	Error  any  // core.Custom error message
	Coerce bool // Enable coercion
}

//////////////////////////////////////////
//////////     ISO Date API      ////////
//////////////////////////////////////////

// Date returns an ISO date format validator (YYYY-MM-DD) with Min/Max support
func (ZodISO) Date(options ...any) *ZodISODate {
	opts := parseISODateOptions(options...)

	// Create definition with regex validation check
	def := &ZodISODateDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "string",
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   "string",
		Checks: make([]core.ZodCheck, 0),
	}

	// Add ISO date format validation check using existing checks system
	dateCheck := checks.ISODate(getErrorMessageFromOptions(opts.Error, "Invalid ISO date format"))
	def.Checks = append(def.Checks, dateCheck)
	def.ZodTypeDef.Checks = append(def.ZodTypeDef.Checks, dateCheck)

	// If a custom error message/function is provided, attach it directly to the check definition
	if opts.Error != nil {
		if em := issues.CreateErrorMap(opts.Error); em != nil {
			dateCheck.GetZod().Def.Error = em
		}
	}

	// Create schema using the unified constructor
	schema := createZodISODateFromDef(def)

	// Apply options
	if opts.Coerce {
		schema.internals.Bag["coerce"] = true
		schema.internals.ZodTypeInternals.Bag["coerce"] = true
	}

	return schema
}

//////////////////////////////////////////
//////////     ISO Time API      ////////
//////////////////////////////////////////

// Time returns an ISO time format validator (HH:MM:SS)
func (ZodISO) Time(options ...any) core.ZodType[any, any] {
	opts := parseISOTimeOptions(options...)

	// Create ISO time regex pattern based on precision
	var pattern string
	if opts.Precision != nil {
		p := *opts.Precision
		if p <= 0 {
			// HH:MM or HH:MM:SS without fractional seconds
			pattern = `^([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9])?$`
		} else {
			// Exactly p fractional second digits
			pattern = fmt.Sprintf(`^([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]\.[0-9]{%d}$`, p)
		}
	} else {
		// seconds optional, fractional seconds optional (any length)
		pattern = `^([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9](\.[0-9]+)?)?$`
	}

	var schema *ZodString
	if opts.Error != nil && opts.Coerce {
		schema = String(core.SchemaParams{Error: opts.Error, Coerce: true}).Regex(regexp.MustCompile(pattern))
	} else if opts.Error != nil {
		schema = String(core.SchemaParams{Error: opts.Error}).Regex(regexp.MustCompile(pattern))
	} else if opts.Coerce {
		schema = String(core.SchemaParams{Coerce: true}).Regex(regexp.MustCompile(pattern))
	} else {
		schema = String().Regex(regexp.MustCompile(pattern))
	}

	return any(schema).(core.ZodType[any, any])
}

//////////////////////////////////////////
//////////   ISO DateTime API    ////////
//////////////////////////////////////////

// DateTime returns an ISO 8601 datetime format validator
func (ZodISO) DateTime(options ...any) core.ZodType[any, any] {
	opts := parseISODateTimeOptions(options...)

	// Create ISO datetime regex pattern based on options
	var pattern string
	if opts.Precision != nil {
		p := *opts.Precision
		if p <= 0 {
			// Hours 00-23, minutes 00-59, optional seconds 00-59
			pattern = `^\d{4}-\d{2}-\d{2}T([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9])?`
		} else {
			// Exact fractional second digits
			pattern = fmt.Sprintf(`^\d{4}-\d{2}-\d{2}T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]\.\d{%d}`, p)
		}
	} else {
		// Hours 00-23, optional seconds and fractional seconds
		pattern = `^\d{4}-\d{2}-\d{2}T([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9](\.\d+)?)?`
	}

	// Add timezone handling – follows Zod behaviour
	if opts.Offset && opts.Local {
		// Timezone optional, but if present must be Z or offset
		pattern += `([+-]\d{2}:\d{2}|Z)?`
	} else if opts.Offset {
		// Require Z or offset
		pattern += `([+-]\d{2}:\d{2}|Z)`
	} else if opts.Local {
		// Make trailing Z optional
		pattern += `(Z)?`
	} else {
		// Require literal Z
		pattern += `Z`
	}
	pattern += `$`

	var schema *ZodString
	if opts.Error != nil && opts.Coerce {
		schema = String(core.SchemaParams{Error: opts.Error, Coerce: true}).Regex(regexp.MustCompile(pattern))
	} else if opts.Error != nil {
		schema = String(core.SchemaParams{Error: opts.Error}).Regex(regexp.MustCompile(pattern))
	} else if opts.Coerce {
		schema = String(core.SchemaParams{Coerce: true}).Regex(regexp.MustCompile(pattern))
	} else {
		schema = String().Regex(regexp.MustCompile(pattern))
	}

	return any(schema).(core.ZodType[any, any])
}

//////////////////////////////////////////
//////////   ISO Duration API    ////////
//////////////////////////////////////////

// Duration returns an ISO 8601 duration format validator
func (ZodISO) Duration(options ...any) core.ZodType[any, any] {
	opts := parseISODurationOptions(options...)

	// Build base String schema with coercion/error options
	base := String(core.SchemaParams{Error: opts.Error, Coerce: opts.Coerce})

	// Re-use existing duration validator from ZodString to leverage validate.ISODuration logic
	schema := base.Duration()

	return any(schema).(core.ZodType[any, any])
}

//////////////////////////////////////////
//////////   ISO COERCE API      ////////
//////////////////////////////////////////

// ZodISOCoerce provides coerced ISO validators
type ZodISOCoerce struct{}

// Date returns a coerced ISO date format validator
func (ZodISOCoerce) Date(options ...any) *ZodISODate {
	opts := parseISODateOptions(options...)
	opts.Coerce = true
	return ISO.Date(opts)
}

// Time returns a coerced ISO time format validator
func (ZodISOCoerce) Time(options ...any) core.ZodType[any, any] {
	opts := parseISOTimeOptions(options...)
	opts.Coerce = true
	return ISO.Time(opts)
}

// DateTime returns a coerced ISO datetime format validator
func (ZodISOCoerce) DateTime(options ...any) core.ZodType[any, any] {
	opts := parseISODateTimeOptions(options...)
	opts.Coerce = true
	return ISO.DateTime(opts)
}

// Duration returns a coerced ISO duration format validator
func (ZodISOCoerce) Duration(options ...any) core.ZodType[any, any] {
	opts := parseISODurationOptions(options...)
	opts.Coerce = true
	return ISO.Duration(opts)
}

// Coerce provides access to coerced ISO format validators
// Usage: ISO.Coerce().Date(), ISO.Coerce().Time(), etc.
func (ZodISO) Coerce() ZodISOCoerce {
	return ZodISOCoerce{}
}

//////////////////////////////////////////
//////////   ISO OPTIONAL WRAPPERS //////
//////////////////////////////////////////

// ZodISOOptional provides optional ISO validators
// Following the pattern from type_string.go and type_optional.go
type ZodISOOptional struct{}

// Date returns an optional ISO date format validator
func (ZodISOOptional) Date(options ...any) core.ZodType[any, any] {
	// Build an ISO Date schema and wrap it with Optional so that nil values are allowed.
	// This reuses the existing Optional helper which turns any schema into an optional one.
	return Optional(ISO.Date(options...))
}

// Time returns an optional ISO time format validator
func (ZodISOOptional) Time(options ...any) core.ZodType[any, any] {
	// Build an ISO Time schema and wrap it with Optional to allow nil values.
	return Optional(ISO.Time(options...))
}

// DateTime returns an optional ISO datetime format validator
func (ZodISOOptional) DateTime(options ...any) core.ZodType[any, any] {
	// Build an ISO DateTime schema and wrap it with Optional to allow nil values.
	return Optional(ISO.DateTime(options...))
}

// Duration returns an optional ISO duration format validator
func (ZodISOOptional) Duration(options ...any) core.ZodType[any, any] {
	// Build an ISO Duration schema and wrap it with Optional to allow nil values.
	return Optional(ISO.Duration(options...))
}

// Optional provides access to optional ISO format validators
// Usage: ISO.Optional().Date(), ISO.Optional().Time(), etc.
func (ZodISO) Optional() ZodISOOptional {
	return ZodISOOptional{}
}

//////////////////////////////////////////
//////////   Helper Functions    ////////
//////////////////////////////////////////

// parseISODateOptions parses various option types for ISO date
func parseISODateOptions(options ...any) ISODateOptions {
	if len(options) == 0 {
		return ISODateOptions{}
	}

	switch opt := options[0].(type) {
	case string:
		return ISODateOptions{Error: opt}
	case ISODateOptions:
		return opt
	case core.SchemaParams:
		return ISODateOptions{Error: opt.Error, Coerce: opt.Coerce}
	default:
		return ISODateOptions{}
	}
}

// parseISOTimeOptions parses various option types for ISO time
func parseISOTimeOptions(options ...any) ISOTimeOptions {
	if len(options) == 0 {
		return ISOTimeOptions{}
	}

	switch opt := options[0].(type) {
	case string:
		return ISOTimeOptions{Error: opt}
	case ISOTimeOptions:
		return opt
	case core.SchemaParams:
		return ISOTimeOptions{Error: opt.Error, Coerce: opt.Coerce}
	default:
		return ISOTimeOptions{}
	}
}

// parseISODateTimeOptions parses various option types for ISO datetime
func parseISODateTimeOptions(options ...any) ISODateTimeOptions {
	if len(options) == 0 {
		return ISODateTimeOptions{}
	}

	switch opt := options[0].(type) {
	case string:
		return ISODateTimeOptions{Error: opt}
	case ISODateTimeOptions:
		return opt
	case core.SchemaParams:
		return ISODateTimeOptions{Error: opt.Error, Coerce: opt.Coerce}
	default:
		return ISODateTimeOptions{}
	}
}

// parseISODurationOptions parses various option types for ISO duration
func parseISODurationOptions(options ...any) ISODurationOptions {
	if len(options) == 0 {
		return ISODurationOptions{}
	}

	switch opt := options[0].(type) {
	case string:
		return ISODurationOptions{Error: opt}
	case ISODurationOptions:
		return opt
	case core.SchemaParams:
		return ISODurationOptions{Error: opt.Error, Coerce: opt.Coerce}
	default:
		return ISODurationOptions{}
	}
}

// getErrorMessageFromOptions extracts error message as string with fallback
func getErrorMessageFromOptions(errorOpt any, defaultMsg string) string {
	if errorOpt != nil {
		if msg, ok := errorOpt.(string); ok {
			return msg
		}
	}
	return defaultMsg
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodISODate) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

////////////////////////////
////   CONSTRUCTOR      ////
////////////////////////////

// createZodISODateFromDef creates a ZodISODate from definition using unified patterns
func createZodISODateFromDef(def *ZodISODateDef) *ZodISODate {
	internals := &ZodISODateInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           make([]core.ZodCheck, 0),
		Isst:             issues.ZodIssueInvalidType{Expected: "string"},
		Pattern:          nil,
		Values:           make(map[string]struct{}),
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		isoDateDef := &ZodISODateDef{
			ZodTypeDef: *newDef,
			Type:       "string",
			Checks:     newDef.Checks,
		}
		return any(createZodISODateFromDef(isoDateDef)).(core.ZodType[any, any])
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		result, err := engine.ParseType[string](
			payload.Value,
			&internals.ZodTypeInternals,
			"string",
			reflectx.ExtractString,
			func(v any) (*string, bool) {
				if reflectx.IsPointer(v) {
					if deref, ok := reflectx.Deref(v); ok {
						if str, ok := deref.(string); ok {
							return &str, true
						}
					}
				}
				return nil, false
			},
			func(value string, checks []core.ZodCheck, ctx *core.ParseContext) error {
				// Run all checks including date format validation
				if len(checks) > 0 {
					checkPayload := &core.ParsePayload{
						Value:  value,
						Issues: make([]core.ZodRawIssue, 0),
					}
					engine.RunChecksOnValue(value, checks, checkPayload, ctx)
					if len(checkPayload.Issues) > 0 {
						return &issues.ZodError{Issues: issues.ConvertRawIssuesToIssues(checkPayload.Issues, ctx)}
					}
				}
				return nil
			},
			func(v any) (string, bool) {
				// Use pkg/coerce for type coercion
				if result, err := coerce.ToISODate(v); err == nil {
					return result, true
				}
				return "", false
			},
			ctx,
		)

		if err != nil {
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := core.ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}

		payload.Value = result
		return payload
	}

	zodSchema := &ZodISODate{internals: internals}

	// Initialize the schema using the unified initZodType from type.go
	// Note: initZodType will copy def.Checks to internals.Checks and execute OnAttach callback
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

//nolint:unused // helper for future error customization
func getErrorFromOptions(errorOpt any, defaultMsg string) any {
	if errorOpt != nil {
		return errorOpt
	}
	return defaultMsg
}
