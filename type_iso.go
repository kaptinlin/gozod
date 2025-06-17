package gozod

import (
	"errors"
	"regexp"
	"time"

	"github.com/kaptinlin/gozod/regexes"
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
	ZodTypeDef
	Type   string     // "string"
	Checks []ZodCheck // ISO date-specific validation checks
}

// ZodISODateInternals contains ISO date internal state
type ZodISODateInternals struct {
	ZodTypeInternals
	Def     *ZodISODateDef         // Schema definition
	Checks  []ZodCheck             // Validation checks
	Isst    ZodIssueInvalidType    // Invalid type issue template
	Pattern *regexp.Regexp         // Regex pattern (if any)
	Values  map[string]struct{}    // Allowed string values set
	Bag     map[string]interface{} // Additional metadata
}

// ZodISODate represents an ISO date schema with validation support
type ZodISODate struct {
	internals *ZodISODateInternals
}

// GetInternals implements ZodType interface
func (z *ZodISODate) GetInternals() *ZodTypeInternals {
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
				tgtState.Bag = make(map[string]interface{})
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}

		// Copy state from ZodTypeInternals.Bag (used by parseType)
		if len(srcState.ZodTypeInternals.Bag) > 0 {
			if tgtState.ZodTypeInternals.Bag == nil {
				tgtState.ZodTypeInternals.Bag = make(map[string]interface{})
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
func (z *ZodISODate) Parse(input any, ctx ...*ParseContext) (any, error) {
	var parseCtx *ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	} else {
		parseCtx = NewParseContext()
	}

	return parseType(
		input,
		&z.internals.ZodTypeInternals,
		"string",
		func(input any) (string, bool) {
			if s, ok := input.(string); ok {
				return s, true
			}
			return "", false
		},
		func(input any) (*string, bool) {
			if s, ok := input.(*string); ok {
				return s, true
			}
			return nil, false
		},
		validateString,
		func(input any) (string, bool) {
			if result, ok := coerceToISODate(input); ok {
				if str, isStr := result.(string); isStr {
					return str, true
				}
			}
			return "", false
		},
		parseCtx,
	)
}

// MustParse implements ZodType interface
func (z *ZodISODate) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Coerce implements Coercible interface for ISO date coercion
func (z *ZodISODate) Coerce(input interface{}) (interface{}, bool) {
	return coerceToISODate(input)
}

// Min adds minimum date validation
func (z *ZodISODate) Min(minDate string, params ...SchemaParams) *ZodISODate {
	// Parse the date to validate it's a valid ISO date
	if _, err := time.Parse("2006-01-02", minDate); err != nil {
		panic("Invalid ISO date format for Min: " + minDate)
	}

	// Create ISO date minimum check using the checks system
	check := NewZodCheckISODateMin(minDate, params...)

	// Use AddCheck to properly add the check and maintain state
	result := AddCheck(z, check)
	return result.(*ZodISODate)
}

// Max adds maximum date validation
func (z *ZodISODate) Max(maxDate string, params ...SchemaParams) *ZodISODate {
	// Parse the date to validate it's a valid ISO date
	if _, err := time.Parse("2006-01-02", maxDate); err != nil {
		panic("Invalid ISO date format for Max: " + maxDate)
	}

	// Create ISO date maximum check using the checks system
	check := NewZodCheckISODateMax(maxDate, params...)

	// Use AddCheck to properly add the check and maintain state
	result := AddCheck(z, check)
	return result.(*ZodISODate)
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
func (z *ZodISODate) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	result := AddCheck(z, check)
	return result
}

// TransformAny adds data transformation with any input/output types
func (z *ZodISODate) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  z,
		out: transform,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Nilable makes the schema accept nil values
func (z *ZodISODate) Nilable() ZodType[any, any] {
	newInternals := &ZodISODateInternals{
		ZodTypeInternals: z.internals.ZodTypeInternals,
	}
	newInternals.Nilable = true
	return &ZodISODate{internals: newInternals}
}

// Pipe implements ZodType interface
func (z *ZodISODate) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  z,
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////////////////////
//////////   ISO Date Options   //////////
//////////////////////////////////////////

// ISODateOptions defines parameters for ISO date validation
type ISODateOptions struct {
	Error  interface{} // Custom error message
	Coerce bool        // Enable coercion
}

//////////////////////////////////////////
//////////   ISO Time Options   //////////
//////////////////////////////////////////

// ISOTimeOptions defines parameters for ISO time validation
type ISOTimeOptions struct {
	// Precision specifies number of decimal places for seconds
	// If nil, matches any number of decimal places
	// If 0 or negative, no decimal places allowed
	Precision *int        `json:"precision,omitempty"`
	Error     interface{} `json:"error,omitempty"`  // Custom error message
	Coerce    bool        `json:"coerce,omitempty"` // Enable coercion
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

	Error  interface{} `json:"error,omitempty"`  // Custom error message
	Coerce bool        `json:"coerce,omitempty"` // Enable coercion
}

//////////////////////////////////////////
//////////  ISO Duration Options ////////
//////////////////////////////////////////

// ISODurationOptions defines parameters for ISO duration validation
type ISODurationOptions struct {
	Error  interface{} // Custom error message
	Coerce bool        // Enable coercion
}

//////////////////////////////////////////
//////////     ISO Date API      ////////
//////////////////////////////////////////

// Date returns an ISO date format validator (YYYY-MM-DD) with Min/Max support
func (ZodISO) Date(options ...interface{}) *ZodISODate {
	opts := parseISODateOptions(options...)

	// Create definition with regex validation check
	def := &ZodISODateDef{
		ZodTypeDef: ZodTypeDef{
			Type:   "string",
			Checks: make([]ZodCheck, 0),
		},
		Type:   "string",
		Checks: make([]ZodCheck, 0),
	}

	// Add regex validation check
	regexCheck := NewZodCheckStringFormat(StringFormatDate, regexes.Date, SchemaParams{
		Error: getErrorFromOptions(opts.Error, "Invalid ISO date format"),
	})
	def.Checks = append(def.Checks, regexCheck)
	def.ZodTypeDef.Checks = append(def.ZodTypeDef.Checks, regexCheck)

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
func (ZodISO) Time(options ...interface{}) ZodType[any, any] {
	opts := parseISOTimeOptions(options...)

	// Use regexes.Time with precision options
	var regex = regexes.DefaultTime
	if opts.Precision != nil {
		regex = regexes.Time(regexes.TimeOptions{Precision: opts.Precision})
	}

	params := SchemaParams{
		Error: getErrorFromOptions(opts.Error, "Invalid ISO time format"),
	}
	if opts.Coerce {
		params.Coerce = true
	}

	schema := String(params).Regex(regex, SchemaParams{
		Error: getErrorFromOptions(opts.Error, "Invalid ISO time format"),
	})

	// Return the schema directly (String() returns ZodType[any, any] interface)
	return schema
}

//////////////////////////////////////////
//////////   ISO DateTime API    ////////
//////////////////////////////////////////

// DateTime returns an ISO 8601 datetime format validator
func (ZodISO) DateTime(options ...interface{}) ZodType[any, any] {
	opts := parseISODateTimeOptions(options...)

	// Use regexes.Datetime with precision, offset, local options
	var regex = regexes.DefaultDatetime
	if opts.Precision != nil || opts.Offset || opts.Local {
		regex = regexes.Datetime(regexes.DatetimeOptions{
			Precision: opts.Precision,
			Offset:    opts.Offset,
			Local:     opts.Local,
		})
	}

	params := SchemaParams{
		Error: getErrorFromOptions(opts.Error, "Invalid ISO datetime format"),
	}
	if opts.Coerce {
		params.Coerce = true
	}

	schema := String(params).Regex(regex, SchemaParams{
		Error: getErrorFromOptions(opts.Error, "Invalid ISO datetime format"),
	})

	// Return the schema directly (String() returns ZodType[any, any] interface)
	return schema
}

//////////////////////////////////////////
//////////   ISO Duration API    ////////
//////////////////////////////////////////

// Duration returns an ISO 8601 duration format validator
func (ZodISO) Duration(options ...interface{}) ZodType[any, any] {
	opts := parseISODurationOptions(options...)

	params := SchemaParams{
		Error: getErrorFromOptions(opts.Error, "Invalid ISO duration format"),
	}
	if opts.Coerce {
		params.Coerce = true
	}

	// Use regexes.Duration from datetimes.go with additional validation
	schema := String(params).Regex(regexes.Duration, SchemaParams{
		Error: getErrorFromOptions(opts.Error, "Invalid ISO duration format"),
	}).RefineAny(func(v interface{}) bool {
		// Additional validation to reject empty patterns like "P" or "PT"
		if str, ok := v.(string); ok {
			return regexes.ValidateDuration(str)
		}
		return false
	}, SchemaParams{
		Error: getErrorFromOptions(opts.Error, "Invalid ISO duration format"),
	})

	// Return the schema directly (String() returns ZodType[any, any] interface)
	return schema
}

//////////////////////////////////////////
//////////   ISO COERCE API      ////////
//////////////////////////////////////////

// ZodISOCoerce provides coerced ISO validators
type ZodISOCoerce struct{}

// Date returns a coerced ISO date format validator
func (ZodISOCoerce) Date(options ...interface{}) *ZodISODate {
	opts := parseISODateOptions(options...)
	opts.Coerce = true
	return ISO.Date(opts)
}

// Time returns a coerced ISO time format validator
func (ZodISOCoerce) Time(options ...interface{}) ZodType[any, any] {
	opts := parseISOTimeOptions(options...)
	opts.Coerce = true
	return ISO.Time(opts)
}

// DateTime returns a coerced ISO datetime format validator
func (ZodISOCoerce) DateTime(options ...interface{}) ZodType[any, any] {
	opts := parseISODateTimeOptions(options...)
	opts.Coerce = true
	return ISO.DateTime(opts)
}

// Duration returns a coerced ISO duration format validator
func (ZodISOCoerce) Duration(options ...interface{}) ZodType[any, any] {
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
func (ZodISOOptional) Date(options ...interface{}) *ZodOptional[ZodType[any, any]] {
	schema := ISO.Date(options...)
	return Optional(schema).(*ZodOptional[ZodType[any, any]])
}

// Time returns an optional ISO time format validator
func (ZodISOOptional) Time(options ...interface{}) *ZodOptional[ZodType[any, any]] {
	schema := ISO.Time(options...)
	return Optional(schema).(*ZodOptional[ZodType[any, any]])
}

// DateTime returns an optional ISO datetime format validator
func (ZodISOOptional) DateTime(options ...interface{}) *ZodOptional[ZodType[any, any]] {
	schema := ISO.DateTime(options...)
	return Optional(schema).(*ZodOptional[ZodType[any, any]])
}

// Duration returns an optional ISO duration format validator
func (ZodISOOptional) Duration(options ...interface{}) *ZodOptional[ZodType[any, any]] {
	schema := ISO.Duration(options...)
	return Optional(schema).(*ZodOptional[ZodType[any, any]])
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
func parseISODateOptions(options ...interface{}) ISODateOptions {
	if len(options) == 0 {
		return ISODateOptions{}
	}

	switch opt := options[0].(type) {
	case string:
		return ISODateOptions{Error: opt}
	case ISODateOptions:
		return opt
	case SchemaParams:
		return ISODateOptions{Error: opt.Error, Coerce: opt.Coerce}
	default:
		return ISODateOptions{}
	}
}

// parseISOTimeOptions parses various option types for ISO time
func parseISOTimeOptions(options ...interface{}) ISOTimeOptions {
	if len(options) == 0 {
		return ISOTimeOptions{}
	}

	switch opt := options[0].(type) {
	case string:
		return ISOTimeOptions{Error: opt}
	case ISOTimeOptions:
		return opt
	case SchemaParams:
		return ISOTimeOptions{Error: opt.Error, Coerce: opt.Coerce}
	default:
		return ISOTimeOptions{}
	}
}

// parseISODateTimeOptions parses various option types for ISO datetime
func parseISODateTimeOptions(options ...interface{}) ISODateTimeOptions {
	if len(options) == 0 {
		return ISODateTimeOptions{}
	}

	switch opt := options[0].(type) {
	case string:
		return ISODateTimeOptions{Error: opt}
	case ISODateTimeOptions:
		return opt
	case SchemaParams:
		return ISODateTimeOptions{Error: opt.Error, Coerce: opt.Coerce}
	default:
		return ISODateTimeOptions{}
	}
}

// parseISODurationOptions parses various option types for ISO duration
func parseISODurationOptions(options ...interface{}) ISODurationOptions {
	if len(options) == 0 {
		return ISODurationOptions{}
	}

	switch opt := options[0].(type) {
	case string:
		return ISODurationOptions{Error: opt}
	case ISODurationOptions:
		return opt
	case SchemaParams:
		return ISODurationOptions{Error: opt.Error, Coerce: opt.Coerce}
	default:
		return ISODurationOptions{}
	}
}

// getErrorFromOptions extracts error message with fallback
func getErrorFromOptions(errorOpt interface{}, defaultMsg string) interface{} {
	if errorOpt != nil {
		return errorOpt
	}
	return defaultMsg
}

// coerceToISODate attempts to coerce a value to ISO date string format
func coerceToISODate(value interface{}) (interface{}, bool) {
	if value == nil {
		return "", false
	}

	// If already a string, try to parse and reformat
	if str, ok := value.(string); ok {
		// Try to parse as various date formats and convert to ISO date
		if t, err := time.Parse("2006-01-02", str); err == nil {
			return t.Format("2006-01-02"), true
		}
		if t, err := time.Parse("2006-01-02T15:04:05Z07:00", str); err == nil {
			return t.Format("2006-01-02"), true
		}
		if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
			return t.Format("2006-01-02"), true
		}
		if t, err := time.Parse("01/02/2006", str); err == nil {
			return t.Format("2006-01-02"), true
		}
		if t, err := time.Parse("02/01/2006", str); err == nil {
			return t.Format("2006-01-02"), true
		}
		return str, false // Return original if can't parse
	}

	// If it's a time.Time, format it
	if t, ok := value.(time.Time); ok {
		return t.Format("2006-01-02"), true
	}

	// If it's a pointer to time.Time, dereference and format
	if t, ok := value.(*time.Time); ok && t != nil {
		return t.Format("2006-01-02"), true
	}

	// Try to convert to string first, then parse
	if str, ok := coerceToString(value); ok {
		return coerceToISODate(str)
	}

	return value, false
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodISODate) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

////////////////////////////
////   CONSTRUCTOR      ////
////////////////////////////

// createZodISODateFromDef creates a ZodISODate from definition using unified patterns
func createZodISODateFromDef(def *ZodISODateDef) *ZodISODate {
	internals := &ZodISODateInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           make([]ZodCheck, 0),
		Isst:             ZodIssueInvalidType{Expected: "string"},
		Pattern:          nil,
		Values:           make(map[string]struct{}),
		Bag:              make(map[string]interface{}),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		isoDateDef := &ZodISODateDef{
			ZodTypeDef: *newDef,
			Type:       "string",
			Checks:     newDef.Checks,
		}
		return any(createZodISODateFromDef(isoDateDef)).(ZodType[any, any])
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		result, err := parseType[string](
			payload.Value,
			&internals.ZodTypeInternals,
			"string",
			func(v any) (string, bool) { str, ok := v.(string); return str, ok },
			func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
			validateString,
			func(input any) (string, bool) {
				if result, ok := coerceToISODate(input); ok {
					if str, isStr := result.(string); isStr {
						return str, true
					}
				}
				return "", false
			},
			ctx,
		)

		if err != nil {
			var zodErr *ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
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

	schema := &ZodISODate{internals: internals}

	// Initialize the schema using the unified initZodType from type.go
	// Note: initZodType will copy def.Checks to internals.Checks and execute OnAttach callback
	initZodType(schema, &def.ZodTypeDef)

	return schema
}
