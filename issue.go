package gozod

import (
	"fmt"
)

///////////////////////////
////     BASE TYPES    ////
///////////////////////////

// ParseParams represents parameters for parse-time configuration
type ParseParams struct {
	Error       ZodErrorMap // Per-parse error customization
	ReportInput bool        // Whether to include input in errors
}

// ZodErrorMap represents a function that maps raw issues to error messages
type ZodErrorMap func(ZodRawIssue) string

// ZodIssueBase represents the base structure for all validation issues
type ZodIssueBase struct {
	Code    string        `json:"code,omitempty"`
	Input   interface{}   `json:"input,omitempty"`
	Path    []interface{} `json:"path"`
	Message string        `json:"message"`
}

////////////////////////////////
////     ISSUE SUBTYPES     ////
////////////////////////////////

// ZodIssueInvalidType represents a type validation error
type ZodIssueInvalidType struct {
	ZodIssueBase
	Expected string `json:"expected"`
	Received string `json:"received"`
}

// ZodIssueTooBig represents a value exceeding maximum constraint error
type ZodIssueTooBig struct {
	ZodIssueBase
	Origin    string      `json:"origin"`
	Maximum   interface{} `json:"maximum"`
	Inclusive bool        `json:"inclusive,omitempty"`
}

// ZodIssueTooSmall represents a value below minimum constraint error
type ZodIssueTooSmall struct {
	ZodIssueBase
	Origin    string      `json:"origin"`
	Minimum   interface{} `json:"minimum"`
	Inclusive bool        `json:"inclusive,omitempty"`
}

// ZodIssueInvalidStringFormat represents an invalid string format error
type ZodIssueInvalidStringFormat struct {
	ZodIssueBase
	Format  string `json:"format"`
	Pattern string `json:"pattern,omitempty"`
}

// ZodIssueNotMultipleOf represents a value not being a multiple of expected divisor
type ZodIssueNotMultipleOf struct {
	ZodIssueBase
	Divisor interface{} `json:"divisor"`
}

// ZodIssueUnrecognizedKeys represents unrecognized object keys error
type ZodIssueUnrecognizedKeys struct {
	ZodIssueBase
	Keys []string `json:"keys"`
}

// ZodIssueInvalidUnion represents failure to match any union schemas
type ZodIssueInvalidUnion struct {
	ZodIssueBase
	Errors [][]ZodIssue `json:"errors"`
}

// ZodIssueInvalidKey represents invalid key in a map or record
type ZodIssueInvalidKey struct {
	ZodIssueBase
	Origin string     `json:"origin"`
	Issues []ZodIssue `json:"issues"`
}

// ZodIssueInvalidElement represents invalid element in a collection
type ZodIssueInvalidElement struct {
	ZodIssueBase
	Origin string      `json:"origin"`
	Key    interface{} `json:"key"`
	Issues []ZodIssue  `json:"issues"`
}

// ZodIssueInvalidValue represents a value not matching expected values
type ZodIssueInvalidValue struct {
	ZodIssueBase
	Values []interface{} `json:"values"`
}

// ZodIssueCustom represents a custom validation error
type ZodIssueCustom struct {
	ZodIssueBase
	Params map[string]interface{} `json:"params,omitempty"`
}

////////////////////////////////////////////
////     FIRST-PARTY STRING FORMATS     ////
////////////////////////////////////////////

// ZodIssueStringCommonFormats represents common string format validation errors
type ZodIssueStringCommonFormats struct {
	ZodIssueInvalidStringFormat
}

// ZodIssueStringInvalidRegex represents regex pattern validation error
type ZodIssueStringInvalidRegex struct {
	ZodIssueInvalidStringFormat
	Pattern string `json:"pattern"`
}

// ZodIssueStringInvalidJWT represents JWT validation error
type ZodIssueStringInvalidJWT struct {
	ZodIssueInvalidStringFormat
	Algorithm string `json:"algorithm,omitempty"`
}

// ZodIssueStringStartsWith represents string prefix validation error
type ZodIssueStringStartsWith struct {
	ZodIssueInvalidStringFormat
	Prefix string `json:"prefix"`
}

// ZodIssueStringEndsWith represents string suffix validation error
type ZodIssueStringEndsWith struct {
	ZodIssueInvalidStringFormat
	Suffix string `json:"suffix"`
}

// ZodIssueStringIncludes represents string inclusion validation error
type ZodIssueStringIncludes struct {
	ZodIssueInvalidStringFormat
	Includes string `json:"includes"`
}

////////////////////////
////     UNIONS     ////
////////////////////////

// ZodStringFormatIssues represents the union of all string format issues
type ZodStringFormatIssues interface{}

// ZodIssue represents the union of all possible validation issues, optimized for Go
type ZodIssue struct {
	ZodIssueBase

	// Invalid type fields
	Expected string `json:"expected,omitempty"`
	Received string `json:"received,omitempty"`

	// Size constraint fields
	Minimum   interface{} `json:"minimum,omitempty"`
	Maximum   interface{} `json:"maximum,omitempty"`
	Inclusive bool        `json:"inclusive,omitempty"`

	// String format fields
	Format    string `json:"format,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Suffix    string `json:"suffix,omitempty"`
	Includes  string `json:"includes,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`

	// Other constraint fields
	Divisor interface{}   `json:"divisor,omitempty"`
	Keys    []string      `json:"keys,omitempty"`
	Values  []interface{} `json:"values,omitempty"`
	Origin  string        `json:"origin,omitempty"`

	// Nested issue fields
	Errors [][]ZodIssue           `json:"errors,omitempty"`
	Issues []ZodIssue             `json:"issues,omitempty"`
	Key    interface{}            `json:"key,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// ZodRawIssue represents a raw issue before finalization, optimized to reduce duplication
type ZodRawIssue struct {
	// Core fields (required)
	Code  string      `json:"code"`
	Input interface{} `json:"input,omitempty"`

	// Optional fields (will be set during finalization)
	Path    []interface{} `json:"path,omitempty"`
	Message string        `json:"message,omitempty"`

	// Dynamic properties stored as key-value pairs to reduce struct size
	Properties map[string]interface{} `json:"properties,omitempty"`

	// Internal fields (not serialized)
	Continue bool        `json:"-"`
	Inst     interface{} `json:"-"`
}

/////////////////////////
////     CONSTANTS   ////
/////////////////////////

// IssueCode represents validation issue types
type IssueCode string

const (
	InvalidType      IssueCode = "invalid_type"
	TooBig           IssueCode = "too_big"
	TooSmall         IssueCode = "too_small"
	InvalidFormat    IssueCode = "invalid_format"
	NotMultipleOf    IssueCode = "not_multiple_of"
	UnrecognizedKeys IssueCode = "unrecognized_keys"
	InvalidUnion     IssueCode = "invalid_union"
	InvalidKey       IssueCode = "invalid_key"
	InvalidElement   IssueCode = "invalid_element"
	InvalidValue     IssueCode = "invalid_value"
	Custom           IssueCode = "custom"
)

//////////////////////////////
////     CORE METHODS     ////
//////////////////////////////

// Error implements the error interface for ZodIssue
func (i ZodIssue) Error() string {
	return i.Message
}

// String returns string representation of ZodIssue
func (i ZodIssue) String() string {
	return fmt.Sprintf("ZodIssue{Code: %s, Message: %s, Path: %v}", i.Code, i.Message, i.Path)
}

//////////////////////////////
////  CONSTRUCTOR FUNCTIONS ////
//////////////////////////////

// NewRawIssueFromMessage creates a ZodRawIssue with custom message
func NewRawIssueFromMessage(message string, input interface{}, inst interface{}) ZodRawIssue {
	return ZodRawIssue{
		Code:       string(Custom),
		Message:    message,
		Input:      input,
		Inst:       inst,
		Path:       []interface{}{},
		Properties: make(map[string]interface{}),
	}
}

// NewRawIssue creates a new raw issue with the given code and input using Go's options pattern
func NewRawIssue(code string, input interface{}, options ...func(*ZodRawIssue)) ZodRawIssue {
	issue := ZodRawIssue{
		Code:       code,
		Input:      input,
		Path:       []interface{}{},
		Properties: make(map[string]interface{}),
	}

	for _, option := range options {
		option(&issue)
	}

	return issue
}

//////////////////////////////
////   OPTION FUNCTIONS   ////
//////////////////////////////

// WithOrigin sets the origin field using properties map
func WithOrigin(origin string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["origin"] = origin
	}
}

// WithMessage sets the message field
func WithMessage(message string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		issue.Message = message
	}
}

// WithMinimum sets the minimum field using properties map
func WithMinimum(minimum interface{}) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["minimum"] = minimum
	}
}

// WithMaximum sets the maximum field using properties map
func WithMaximum(maximum interface{}) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["maximum"] = maximum
	}
}

// WithExpected sets the expected field using properties map
func WithExpected(expected string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["expected"] = expected
	}
}

// WithReceived sets the received field using properties map
func WithReceived(received string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["received"] = received
	}
}

// WithPath sets the path field
func WithPath(path []interface{}) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		issue.Path = path
	}
}

// WithInclusive sets the inclusive field using properties map
func WithInclusive(inclusive bool) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["inclusive"] = inclusive
	}
}

// WithFormat sets the format field using properties map
func WithFormat(format string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["format"] = format
	}
}

// WithContinue sets the continue field
func WithContinue(cont bool) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		issue.Continue = cont
	}
}

// WithPattern sets the pattern field using properties map
func WithPattern(pattern string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["pattern"] = pattern
	}
}

// WithPrefix sets the prefix field using properties map
func WithPrefix(prefix string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["prefix"] = prefix
	}
}

// WithSuffix sets the suffix field using properties map
func WithSuffix(suffix string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["suffix"] = suffix
	}
}

// WithIncludes sets the includes field using properties map
func WithIncludes(includes string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["includes"] = includes
	}
}

// WithDivisor sets the divisor field using properties map
func WithDivisor(divisor interface{}) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["divisor"] = divisor
	}
}

// WithKeys sets the keys field using properties map
func WithKeys(keys []string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["keys"] = keys
	}
}

// WithValues sets the values field using properties map
func WithValues(values []interface{}) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["values"] = values
	}
}

// WithAlgorithm sets the algorithm field using properties map
func WithAlgorithm(algorithm string) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["algorithm"] = algorithm
	}
}

// WithParams sets the params field using properties map
func WithParams(params map[string]interface{}) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]interface{})
		}
		issue.Properties["params"] = params
	}
}

// WithInst sets the inst field for error mapping resolution
func WithInst(inst interface{}) func(*ZodRawIssue) {
	return func(issue *ZodRawIssue) {
		issue.Inst = inst
	}
}

//////////////////////////////
////  TYPE SAFE ACCESSORS ////
//////////////////////////////

// getStringProperty safely gets a string property from the properties map
func getStringProperty(properties map[string]interface{}, key string) string {
	if properties == nil {
		return ""
	}
	if value, ok := properties[key].(string); ok {
		return value
	}
	return ""
}

// GetExpected returns the expected value from properties map
func (r ZodRawIssue) GetExpected() string {
	return getStringProperty(r.Properties, "expected")
}

// GetReceived returns the received value from properties map
func (r ZodRawIssue) GetReceived() string {
	return getStringProperty(r.Properties, "received")
}

// GetOrigin returns the origin value from properties map
func (r ZodRawIssue) GetOrigin() string {
	return getStringProperty(r.Properties, "origin")
}

// GetFormat returns the format value from properties map
func (r ZodRawIssue) GetFormat() string {
	return getStringProperty(r.Properties, "format")
}

// GetPattern returns the pattern value from properties map
func (r ZodRawIssue) GetPattern() string {
	return getStringProperty(r.Properties, "pattern")
}

// GetPrefix returns the prefix value from properties map
func (r ZodRawIssue) GetPrefix() string {
	return getStringProperty(r.Properties, "prefix")
}

// GetSuffix returns the suffix value from properties map
func (r ZodRawIssue) GetSuffix() string {
	return getStringProperty(r.Properties, "suffix")
}

// GetIncludes returns the includes value from properties map
func (r ZodRawIssue) GetIncludes() string {
	return getStringProperty(r.Properties, "includes")
}

// GetMinimum returns the minimum value from properties map
func (r ZodRawIssue) GetMinimum() interface{} {
	if r.Properties == nil {
		return nil
	}
	return r.Properties["minimum"]
}

// GetMaximum returns the maximum value from properties map
func (r ZodRawIssue) GetMaximum() interface{} {
	if r.Properties == nil {
		return nil
	}
	return r.Properties["maximum"]
}

// GetInclusive returns the inclusive value from properties map
func (r ZodRawIssue) GetInclusive() bool {
	if r.Properties == nil {
		return false
	}
	if val, ok := r.Properties["inclusive"].(bool); ok {
		return val
	}
	return false
}

// GetDivisor returns the divisor value from properties map
func (r ZodRawIssue) GetDivisor() interface{} {
	if r.Properties == nil {
		return nil
	}
	return r.Properties["divisor"]
}

// GetKeys returns the keys value from properties map
func (r ZodRawIssue) GetKeys() []string {
	if r.Properties == nil {
		return nil
	}
	if val, ok := r.Properties["keys"].([]string); ok {
		return val
	}
	return nil
}

// GetValues returns the values from properties map
func (r ZodRawIssue) GetValues() []interface{} {
	if r.Properties == nil {
		return nil
	}
	if val, ok := r.Properties["values"].([]interface{}); ok {
		return val
	}
	return nil
}

// GetExpected returns the expected type for invalid_type issues
func (i ZodIssue) GetExpected() (string, bool) {
	if i.Code != string(InvalidType) {
		return "", false
	}
	return i.Expected, i.Expected != ""
}

// GetReceived returns the received type for invalid_type issues
func (i ZodIssue) GetReceived() (string, bool) {
	if i.Code != string(InvalidType) {
		return "", false
	}
	return i.Received, i.Received != ""
}

// GetMinimum returns the minimum value for too_small issues
func (i ZodIssue) GetMinimum() (interface{}, bool) {
	if i.Code != string(TooSmall) {
		return nil, false
	}
	return i.Minimum, i.Minimum != nil
}

// GetMaximum returns the maximum value for too_big issues
func (i ZodIssue) GetMaximum() (interface{}, bool) {
	if i.Code != string(TooBig) {
		return nil, false
	}
	return i.Maximum, i.Maximum != nil
}

// GetFormat returns the format for invalid_format issues
func (i ZodIssue) GetFormat() (string, bool) {
	if i.Code != string(InvalidFormat) {
		return "", false
	}
	return i.Format, i.Format != ""
}

// GetDivisor returns the divisor for not_multiple_of issues
func (i ZodIssue) GetDivisor() (interface{}, bool) {
	if i.Code != string(NotMultipleOf) {
		return nil, false
	}
	return i.Divisor, i.Divisor != nil
}

//////////////////////////////
////   ISSUE FINALIZATION ////
//////////////////////////////

// FinalizeIssue creates a finalized ZodIssue from a ZodRawIssue
func FinalizeIssue(iss ZodRawIssue, ctx *ParseContext, config *ZodConfig) ZodIssue {
	// Ensure path is not nil (equivalent to `path: iss.path ?? []`)
	path := iss.Path
	if path == nil {
		path = []interface{}{}
	}

	// Generate message using error resolution chain
	message := iss.Message
	if message == "" {
		// Try to get message from various sources in priority order
		// 1. inst?.error (schema-level error)
		// 2. ctx?.error (context-level error)
		// 3. config.customError (global custom error)
		// 4. config.localeError (locale error)
		// 5. default message

		// Check schema-level error (inst?.error)
		if iss.Inst != nil {
			// Try to extract error mapping from different types of schema internals
			switch inst := iss.Inst.(type) {
			case *ZodCheckCustomInternals:
				if inst.Def.Error != nil {
					if instMsg := (*inst.Def.Error)(iss); instMsg != "" {
						message = instMsg
					}
				}
			case *ZodStringInternals:
				if inst.Error != nil {
					if instMsg := (*inst.Error)(iss); instMsg != "" {
						message = instMsg
					}
				}
			case *ZodStructInternals:
				if inst.Error != nil {
					if instMsg := (*inst.Error)(iss); instMsg != "" {
						message = instMsg
					}
				}
			case interface{ GetZod() *ZodCheckInternals }:
				// Handle ZodCheck types (like ZodCheckMinLength, ZodCheckMaxLength, etc.)
				if checkInternals := inst.GetZod(); checkInternals != nil && checkInternals.Def != nil && checkInternals.Def.Error != nil {
					if instMsg := (*checkInternals.Def.Error)(iss); instMsg != "" {
						message = instMsg
					}
				}
			case interface{ GetInternals() *ZodTypeInternals }:
				// Generic case for any schema that implements GetInternals()
				if typeInternals := inst.GetInternals(); typeInternals != nil && typeInternals.Error != nil {
					if instMsg := (*typeInternals.Error)(iss); instMsg != "" {
						message = instMsg
					}
				}
			}
		}

		// Check context-level error
		if message == "" && ctx != nil && ctx.Error != nil {
			if ctxMsg := ctx.Error(iss); ctxMsg != "" {
				message = ctxMsg
			}
		}

		// Then check config-level errors
		if message == "" && config != nil {
			if customError := config.GetCustomError(); customError != nil {
				if customMsg := customError(iss); customMsg != "" {
					message = customMsg
				}
			}

			if message == "" {
				if localeError := config.GetLocaleError(); localeError != nil {
					if localeMsg := localeError(iss); localeMsg != "" {
						message = localeMsg
					}
				}
			}
		}

		if message == "" {
			message = generateDefaultMessage(iss)
		}
	}

	// Create finalized issue
	issue := ZodIssue{
		ZodIssueBase: ZodIssueBase{
			Code:    iss.Code,
			Path:    path,
			Message: message,
		},
	}

	// Handle input field based on context ReportInput setting
	if ctx == nil || ctx.ReportInput {
		issue.Input = iss.Input
	}

	// Map properties from raw issue to typed fields
	if iss.Properties != nil {
		mapPropertiesToIssue(&issue, iss.Properties)
	}

	return issue
}

// generateDefaultMessage generates default error message based on issue code and properties
func generateDefaultMessage(raw ZodRawIssue) string {
	switch raw.Code {
	case string(InvalidType):
		expected := getStringProperty(raw.Properties, "expected")
		received := getStringProperty(raw.Properties, "received")
		return fmt.Sprintf("Invalid input: expected %s, received %s", expected, received)
	case string(TooSmall):
		minimum := raw.Properties["minimum"]
		origin := getStringProperty(raw.Properties, "origin")
		inclusive := raw.Properties["inclusive"]

		// Match TypeScript Zod v4 format: "Too small: expected array to have >=2 items"
		// Use "slice" for Go slice types to maintain Go terminology
		if origin == "slice" || origin == "array" {
			originName := "slice"
			if origin == "array" {
				originName = "array"
			}
			if inclusive != nil && inclusive.(bool) {
				return fmt.Sprintf("Too small: expected %s to have >=%v items", originName, minimum)
			} else {
				return fmt.Sprintf("Too small: expected %s to have >%v items", originName, minimum)
			}
		}
		// For other types, use the original format
		return fmt.Sprintf("%s must be at least %v", origin, minimum)
	case string(TooBig):
		maximum := raw.Properties["maximum"]
		origin := getStringProperty(raw.Properties, "origin")
		inclusive := raw.Properties["inclusive"]

		// Match TypeScript Zod v4 format: "Too big: expected array to have <=2 items"
		// Use "slice" for Go slice types to maintain Go terminology
		if origin == "slice" || origin == "array" {
			originName := "slice"
			if origin == "array" {
				originName = "array"
			}
			if inclusive != nil && inclusive.(bool) {
				return fmt.Sprintf("Too big: expected %s to have <=%v items", originName, maximum)
			} else {
				return fmt.Sprintf("Too big: expected %s to have <%v items", originName, maximum)
			}
		}
		// For other types, use the original format
		return fmt.Sprintf("%s must be at most %v", origin, maximum)
	case string(InvalidFormat):
		format := getStringProperty(raw.Properties, "format")
		return fmt.Sprintf("Invalid %s", format)
	case string(NotMultipleOf):
		divisor := raw.Properties["divisor"]
		return fmt.Sprintf("Number must be a multiple of %v", divisor)
	case string(UnrecognizedKeys):
		return "Unrecognized key(s) in object"
	case string(InvalidUnion):
		return "Invalid input"
	case string(InvalidValue):
		return "Invalid value"
	case string(Custom):
		return "Refinement failed"
	default:
		return "Invalid input"
	}
}

// mapPropertiesToIssue maps properties from ZodRawIssue to ZodIssue typed fields
func mapPropertiesToIssue(issue *ZodIssue, properties map[string]interface{}) {
	if properties == nil {
		return
	}

	// Map common properties
	if expected, ok := properties["expected"].(string); ok {
		issue.Expected = expected
	}
	if received, ok := properties["received"].(string); ok {
		issue.Received = received
	}
	if minimum := properties["minimum"]; minimum != nil {
		issue.Minimum = minimum
	}
	if maximum := properties["maximum"]; maximum != nil {
		issue.Maximum = maximum
	}
	if inclusive, ok := properties["inclusive"].(bool); ok {
		issue.Inclusive = inclusive
	}
	if format, ok := properties["format"].(string); ok {
		issue.Format = format
	}
	if pattern, ok := properties["pattern"].(string); ok {
		issue.Pattern = pattern
	}
	if prefix, ok := properties["prefix"].(string); ok {
		issue.Prefix = prefix
	}
	if suffix, ok := properties["suffix"].(string); ok {
		issue.Suffix = suffix
	}
	if includes, ok := properties["includes"].(string); ok {
		issue.Includes = includes
	}
	if algorithm, ok := properties["algorithm"].(string); ok {
		issue.Algorithm = algorithm
	}
	if divisor := properties["divisor"]; divisor != nil {
		issue.Divisor = divisor
	}
	if keys, ok := properties["keys"].([]string); ok {
		issue.Keys = keys
	}
	if values, ok := properties["values"].([]interface{}); ok {
		issue.Values = values
	}
	if origin, ok := properties["origin"].(string); ok {
		issue.Origin = origin
	}
	if key := properties["key"]; key != nil {
		issue.Key = key
	}
	if params, ok := properties["params"].(map[string]interface{}); ok {
		issue.Params = params
	}
}

//////////////////////////////
////  ERROR CREATION HELPERS ////
//////////////////////////////

// CreateInvalidTypeIssue creates a standardized invalid_type issue
func CreateInvalidTypeIssue(input interface{}, expected, received string, options ...func(*ZodRawIssue)) ZodRawIssue {
	return NewRawIssue(
		string(InvalidType),
		input,
		append([]func(*ZodRawIssue){
			WithExpected(expected),
			WithReceived(received),
		}, options...)...,
	)
}

// CreateTooBigIssue creates a standardized too_big issue
func CreateTooBigIssue(input interface{}, origin string, maximum interface{}, inclusive bool, options ...func(*ZodRawIssue)) ZodRawIssue {
	return NewRawIssue(
		string(TooBig),
		input,
		append([]func(*ZodRawIssue){
			WithOrigin(origin),
			WithMaximum(maximum),
			WithInclusive(inclusive),
		}, options...)...,
	)
}

// CreateTooSmallIssue creates a standardized too_small issue
func CreateTooSmallIssue(input interface{}, origin string, minimum interface{}, inclusive bool, options ...func(*ZodRawIssue)) ZodRawIssue {
	return NewRawIssue(
		string(TooSmall),
		input,
		append([]func(*ZodRawIssue){
			WithOrigin(origin),
			WithMinimum(minimum),
			WithInclusive(inclusive),
		}, options...)...,
	)
}

// CreateInvalidFormatIssue creates a standardized invalid_format issue
func CreateInvalidFormatIssue(input interface{}, format string, options ...func(*ZodRawIssue)) ZodRawIssue {
	return NewRawIssue(
		string(InvalidFormat),
		input,
		append([]func(*ZodRawIssue){
			WithOrigin("string"),
			WithFormat(format),
		}, options...)...,
	)
}

// CreateNotMultipleOfIssue creates a standardized not_multiple_of issue
func CreateNotMultipleOfIssue(input interface{}, divisor interface{}, options ...func(*ZodRawIssue)) ZodRawIssue {
	return NewRawIssue(
		string(NotMultipleOf),
		input,
		append([]func(*ZodRawIssue){
			WithDivisor(divisor),
		}, options...)...,
	)
}

// CreateCustomIssue creates a standardized custom issue
func CreateCustomIssue(input interface{}, message string, options ...func(*ZodRawIssue)) ZodRawIssue {
	return NewRawIssue(
		string(Custom),
		input,
		append([]func(*ZodRawIssue){
			WithMessage(message),
		}, options...)...,
	)
}

// CreateInvalidValueIssue creates a standardized invalid_value issue
func CreateInvalidValueIssue(input interface{}, validValues []interface{}, options ...func(*ZodRawIssue)) ZodRawIssue {
	return NewRawIssue(
		string(InvalidValue),
		input,
		append([]func(*ZodRawIssue){
			WithValues(validValues),
		}, options...)...,
	)
}

// CreateInvalidIntersectionIssue creates a standardized invalid_intersection issue
func CreateInvalidIntersectionIssue(input interface{}, mergeError string, options ...func(*ZodRawIssue)) ZodRawIssue {
	return NewRawIssue(
		string(Custom), // Use Custom code for intersection merge errors
		input,
		append([]func(*ZodRawIssue){
			WithMessage(fmt.Sprintf("Unmergable intersection. %s", mergeError)),
		}, options...)...,
	)
}
