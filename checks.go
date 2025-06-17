package gozod

import (
	"fmt"
	"math"
	"math/big"
	"mime/multipart"
	"os"
	"regexp"
	"strings"

	expjson "github.com/go-json-experiment/json"
)

// =============================================================================
// CORE VALIDATION CHECK TYPES
// =============================================================================

// ZodCheckDef defines validation check configuration
type ZodCheckDef struct {
	Check string       // Check type identifier
	Error *ZodErrorMap // Custom error mapping
	Abort bool         // Whether to abort on validation failure
}

// ZodCheckInternals contains check internal state and configuration
type ZodCheckInternals struct {
	Def      *ZodCheckDef                     // Check definition
	Issc     *ZodIssueBase                    // Issues this check might throw
	Check    ZodCheckFn                       // Validation function
	OnAttach []func(schema interface{})       // Schema attachment callbacks
	When     func(payload *ParsePayload) bool // Conditional execution
}

// ZodCheck represents validation constraint interface
type ZodCheck interface {
	GetZod() *ZodCheckInternals
}

// ZodWhenFn defines conditional function type
type ZodWhenFn func(payload *ParsePayload) bool

// ZodCheckFn defines validation execution function
type ZodCheckFn func(payload *ParsePayload)

// NewZodCheckDef creates a new check definition
func NewZodCheckDef(checkType string) *ZodCheckDef {
	return &ZodCheckDef{
		Check: checkType,
	}
}

// NewZodCheckInternals creates check internals with definition
func NewZodCheckInternals(def *ZodCheckDef) *ZodCheckInternals {
	return &ZodCheckInternals{
		Def:      def,
		OnAttach: []func(schema interface{}){},
	}
}

// =============================================================================
// NUMERIC COMPARISON CHECKS
// =============================================================================

// ZodCheckLessThanDef defines less than validation constraint
type ZodCheckLessThanDef struct {
	ZodCheckDef
	Value     interface{}
	Inclusive bool
}

type ZodCheckLessThanInternals struct {
	ZodCheckInternals
	Def *ZodCheckLessThanDef
}

// ZodCheckLessThan validates value < or <= threshold
type ZodCheckLessThan struct {
	Internals *ZodCheckLessThanInternals
}

func (c *ZodCheckLessThan) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckLessThan creates a less than check
func NewZodCheckLessThan(value interface{}, inclusive bool, params ...SchemaParams) *ZodCheckLessThan {
	def := &ZodCheckLessThanDef{
		ZodCheckDef: ZodCheckDef{Check: "less_than"},
		Value:       value,
		Inclusive:   inclusive,
	}

	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckLessThanInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	origin := getNumericOrigin(value)

	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// Bag handling for schema metadata
	})
	internals.Check = func(payload *ParsePayload) {
		var passes bool
		if def.Inclusive {
			passes = compareNumeric(payload.Value, def.Value) <= 0
		} else {
			passes = compareNumeric(payload.Value, def.Value) < 0
		}

		if passes {
			return
		}

		issue := NewRawIssue(
			string(TooBig),
			payload.Value,
			WithOrigin(origin),
			WithMaximum(def.Value),
			WithInclusive(def.Inclusive),
			WithContinue(!def.Abort),
		)
		issue.Inst = &ZodCheckLessThan{Internals: internals}
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckLessThan{Internals: internals}
}

type ZodCheckGreaterThanDef struct {
	ZodCheckDef
	Value     interface{}
	Inclusive bool
}

type ZodCheckGreaterThanInternals struct {
	ZodCheckInternals
	Def *ZodCheckGreaterThanDef
}

// ZodCheckGreaterThan validates value > or >= threshold
type ZodCheckGreaterThan struct {
	Internals *ZodCheckGreaterThanInternals
}

func (c *ZodCheckGreaterThan) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckGreaterThan creates a greater than check
func NewZodCheckGreaterThan(value interface{}, inclusive bool, params ...SchemaParams) *ZodCheckGreaterThan {
	def := &ZodCheckGreaterThanDef{
		ZodCheckDef: ZodCheckDef{Check: "greater_than"},
		Value:       value,
		Inclusive:   inclusive,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckGreaterThanInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	origin := getNumericOrigin(value)

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		var passes bool
		if def.Inclusive {
			passes = compareNumeric(payload.Value, def.Value) >= 0
		} else {
			passes = compareNumeric(payload.Value, def.Value) > 0
		}

		if passes {
			return
		}

		issue := NewRawIssue(
			string(TooSmall),
			payload.Value,
			WithOrigin(origin),
			WithMinimum(def.Value),
			WithInclusive(def.Inclusive),
			WithContinue(!def.Abort),
		)
		issue.Inst = &ZodCheckGreaterThan{Internals: internals}
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckGreaterThan{Internals: internals}
}

type ZodCheckMultipleOfDef struct {
	ZodCheckDef
	Value interface{}
}

type ZodCheckMultipleOfInternals struct {
	ZodCheckInternals
	Def *ZodCheckMultipleOfDef
}

// ZodCheckMultipleOf validates value is multiple of divisor
type ZodCheckMultipleOf struct {
	Internals *ZodCheckMultipleOfInternals
}

func (c *ZodCheckMultipleOf) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckMultipleOf creates a new multiple of check
func NewZodCheckMultipleOf(value interface{}, params ...SchemaParams) *ZodCheckMultipleOf {
	def := &ZodCheckMultipleOfDef{
		ZodCheckDef: ZodCheckDef{Check: "multiple_of"},
		Value:       value,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckMultipleOfInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		isMultiple := false

		// Handle big.Int specially
		if payloadBigInt, ok := payload.Value.(*big.Int); ok {
			if defBigInt, ok := def.Value.(*big.Int); ok {
				// Both are big.Int
				remainder := new(big.Int)
				remainder.Mod(payloadBigInt, defBigInt)
				isMultiple = remainder.Cmp(big.NewInt(0)) == 0
			} else {
				// Payload is big.Int, def is regular number - convert def to big.Int
				if defFloat := toFloat64(def.Value); defFloat == float64(int64(defFloat)) {
					defBigInt := big.NewInt(int64(defFloat))
					remainder := new(big.Int)
					remainder.Mod(payloadBigInt, defBigInt)
					isMultiple = remainder.Cmp(big.NewInt(0)) == 0
				}
			}
		} else if defBigInt, ok := def.Value.(*big.Int); ok {
			// Def is big.Int, payload is regular number - convert payload to big.Int
			if payloadFloat := toFloat64(payload.Value); payloadFloat == float64(int64(payloadFloat)) {
				payloadBigInt := big.NewInt(int64(payloadFloat))
				remainder := new(big.Int)
				remainder.Mod(payloadBigInt, defBigInt)
				isMultiple = remainder.Cmp(big.NewInt(0)) == 0
			}
		} else {
			// Both are regular numeric types
			switch payloadVal := payload.Value.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				payloadFloat := toFloat64(payloadVal)
				defFloat := toFloat64(def.Value)
				if payloadFloat == float64(int64(payloadFloat)) && defFloat == float64(int64(defFloat)) {
					// Both are integers
					isMultiple = int64(payloadFloat)%int64(defFloat) == 0
				} else {
					// Use float safe remainder
					isMultiple = floatSafeRemainder(payloadFloat, defFloat) == 0
				}
			case float32, float64:
				payloadFloat := toFloat64(payloadVal)
				defFloat := toFloat64(def.Value)
				isMultiple = floatSafeRemainder(payloadFloat, defFloat) == 0
			}
		}

		if isMultiple {
			return
		}

		issue := NewRawIssue(
			string(NotMultipleOf),
			payload.Value,
			WithOrigin(getNumericOrigin(payload.Value)),
			WithDivisor(def.Value),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckMultipleOf{Internals: internals}
}

///////////////////////////////////////
/////    ZodCheckPositive         /////
///////////////////////////////////////

// ZodCheckPositiveDef defines positive validation constraint
type ZodCheckPositiveDef struct {
	ZodCheckGreaterThanDef
}

// ZodCheckPositiveInternals contains positive check internal state
type ZodCheckPositiveInternals struct {
	ZodCheckInternals
	Def *ZodCheckPositiveDef
}

// ZodCheckPositive represents positive validation check
type ZodCheckPositive struct {
	Internals *ZodCheckPositiveInternals
}

func (c *ZodCheckPositive) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckPositive creates a new positive check (value > 0)
func NewZodCheckPositive(params ...SchemaParams) *ZodCheckPositive {
	def := &ZodCheckPositiveDef{
		ZodCheckGreaterThanDef: ZodCheckGreaterThanDef{
			ZodCheckDef: ZodCheckDef{Check: "greater_than"},
			Value:       0,
			Inclusive:   false, // positive means > 0, not >= 0
		},
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckPositiveInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	origin := getNumericOrigin(def.Value)

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		passes := compareNumeric(payload.Value, 0) > 0

		if passes {
			return
		}

		issue := NewRawIssue(
			string(TooSmall),
			payload.Value,
			WithOrigin(origin),
			WithMinimum(0),
			WithInclusive(false),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckPositive{Internals: internals}
}

///////////////////////////////////////
/////    ZodCheckNonnegative      /////
///////////////////////////////////////

// ZodCheckNonnegativeDef defines nonnegative validation constraint
type ZodCheckNonnegativeDef struct {
	ZodCheckGreaterThanDef
}

// ZodCheckNonnegativeInternals contains nonnegative check internal state
type ZodCheckNonnegativeInternals struct {
	ZodCheckInternals
	Def *ZodCheckNonnegativeDef
}

// ZodCheckNonnegative represents nonnegative validation check
type ZodCheckNonnegative struct {
	Internals *ZodCheckNonnegativeInternals
}

func (c *ZodCheckNonnegative) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckNonnegative creates a new nonnegative check (value >= 0)
func NewZodCheckNonnegative(params ...SchemaParams) *ZodCheckNonnegative {
	def := &ZodCheckNonnegativeDef{
		ZodCheckGreaterThanDef: ZodCheckGreaterThanDef{
			ZodCheckDef: ZodCheckDef{Check: "greater_than"},
			Value:       0,
			Inclusive:   true, // nonnegative means >= 0
		},
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckNonnegativeInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	origin := getNumericOrigin(def.Value)

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		passes := compareNumeric(payload.Value, 0) >= 0

		if passes {
			return
		}

		issue := NewRawIssue(
			string(TooSmall),
			payload.Value,
			WithOrigin(origin),
			WithMinimum(0),
			WithInclusive(true),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckNonnegative{Internals: internals}
}

///////////////////////////////////////
/////    ZodCheckNegative         /////
///////////////////////////////////////

// ZodCheckNegativeDef defines negative validation constraint
type ZodCheckNegativeDef struct {
	ZodCheckLessThanDef
}

// ZodCheckNegativeInternals contains negative check internal state
type ZodCheckNegativeInternals struct {
	ZodCheckInternals
	Def *ZodCheckNegativeDef
}

// ZodCheckNegative represents negative validation check
type ZodCheckNegative struct {
	Internals *ZodCheckNegativeInternals
}

func (c *ZodCheckNegative) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckNegative creates a new negative check (value < 0)
func NewZodCheckNegative(params ...SchemaParams) *ZodCheckNegative {
	def := &ZodCheckNegativeDef{
		ZodCheckLessThanDef: ZodCheckLessThanDef{
			ZodCheckDef: ZodCheckDef{Check: "less_than"},
			Value:       0,
			Inclusive:   false, // negative means < 0, not <= 0
		},
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckNegativeInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	origin := getNumericOrigin(def.Value)

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		passes := compareNumeric(payload.Value, 0) < 0

		if passes {
			return
		}

		issue := NewRawIssue(
			string(TooBig),
			payload.Value,
			WithOrigin(origin),
			WithMaximum(0),
			WithInclusive(false),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckNegative{Internals: internals}
}

///////////////////////////////////////
/////    ZodCheckNonpositive      /////
///////////////////////////////////////

// ZodCheckNonpositiveDef defines nonpositive validation constraint
type ZodCheckNonpositiveDef struct {
	ZodCheckLessThanDef
}

// ZodCheckNonpositiveInternals contains nonpositive check internal state
type ZodCheckNonpositiveInternals struct {
	ZodCheckInternals
	Def *ZodCheckNonpositiveDef
}

// ZodCheckNonpositive represents nonpositive validation check
type ZodCheckNonpositive struct {
	Internals *ZodCheckNonpositiveInternals
}

func (c *ZodCheckNonpositive) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckNonpositive creates a new nonpositive check (value <= 0)
func NewZodCheckNonpositive(params ...SchemaParams) *ZodCheckNonpositive {
	def := &ZodCheckNonpositiveDef{
		ZodCheckLessThanDef: ZodCheckLessThanDef{
			ZodCheckDef: ZodCheckDef{Check: "less_than"},
			Value:       0,
			Inclusive:   true, // nonpositive means <= 0
		},
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckNonpositiveInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	origin := getNumericOrigin(def.Value)

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		passes := compareNumeric(payload.Value, 0) <= 0

		if passes {
			return
		}

		issue := NewRawIssue(
			string(TooBig),
			payload.Value,
			WithOrigin(origin),
			WithMaximum(0),
			WithInclusive(true),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckNonpositive{Internals: internals}
}

///////////////////////////////////////
/////    ZodCheckNumberFormat     /////
///////////////////////////////////////

// ZodNumberFormats represents number format types
type ZodNumberFormats string

const (
	NumberFormatInt32   ZodNumberFormats = "int32"
	NumberFormatUint32  ZodNumberFormats = "uint32"
	NumberFormatFloat32 ZodNumberFormats = "float32"
	NumberFormatFloat64 ZodNumberFormats = "float64"
	NumberFormatSafeint ZodNumberFormats = "safeint"
)

// ZodCheckNumberFormatDef defines number format validation constraint
type ZodCheckNumberFormatDef struct {
	ZodCheckDef
	Format ZodNumberFormats
}

// ZodCheckNumberFormatInternals contains number format check internal state
type ZodCheckNumberFormatInternals struct {
	ZodCheckInternals
	Def *ZodCheckNumberFormatDef
}

// ZodCheckNumberFormat represents number format validation check
type ZodCheckNumberFormat struct {
	Internals *ZodCheckNumberFormatInternals
}

func (c *ZodCheckNumberFormat) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckNumberFormat creates a new number format check
func NewZodCheckNumberFormat(format ZodNumberFormats, params ...SchemaParams) *ZodCheckNumberFormat {
	if format == "" {
		format = NumberFormatFloat64
	}

	def := &ZodCheckNumberFormatDef{
		ZodCheckDef: ZodCheckDef{Check: "number_format"},
		Format:      format,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckNumberFormatInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	isInt := strings.Contains(string(format), "int")
	origin := "number"
	if isInt {
		origin = "int"
	}

	ranges := NUMBER_FORMAT_RANGES[string(format)]
	minimum := ranges[0]
	maximum := ranges[1]

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(float64)
		if !ok {
			// Try to convert other numeric types
			switch v := payload.Value.(type) {
			case int:
				input = float64(v)
			case int32:
				input = float64(v)
			case int64:
				input = float64(v)
			case float32:
				input = float64(v)
			default:
				issue := CreateInvalidTypeIssue(
					payload.Value,
					origin,
					string(GetParsedType(payload.Value)),
					WithContinue(!def.Abort),
				)
				payload.Issues = append(payload.Issues, issue)
				return
			}
		}

		// Check for NaN and Infinity for finite number formats
		if format == NumberFormatFloat32 || format == NumberFormatFloat64 {
			if math.IsNaN(input) {
				issue := CreateInvalidTypeIssue(
					payload.Value,
					"number",
					"nan",
					WithContinue(!def.Abort),
				)
				payload.Issues = append(payload.Issues, issue)
				return
			}
			if math.IsInf(input, 0) {
				issue := CreateInvalidTypeIssue(
					payload.Value,
					"number",
					"infinity",
					WithContinue(!def.Abort),
				)
				payload.Issues = append(payload.Issues, issue)
				return
			}
		}

		if isInt {
			// Check if it's an integer
			truncatedInput := float64(int64(input))

			// Check if the input is actually a whole number
			if input != truncatedInput {
				// This is a float value when expecting an integer
				issue := CreateInvalidTypeIssue(
					payload.Value,
					origin,
					"number",
					WithContinue(!def.Abort),
				)
				payload.Issues = append(payload.Issues, issue)
				return
			}

			// Check safe integer range for safeint format
			if format == NumberFormatSafeint {
				const maxSafeInt = 9007199254740991  // Number.MAX_SAFE_INTEGER
				const minSafeInt = -9007199254740991 // Number.MIN_SAFE_INTEGER
				if input > maxSafeInt {
					issue := CreateTooBigIssue(
						payload.Value,
						origin,
						maxSafeInt,
						false,
						WithContinue(!def.Abort),
					)
					payload.Issues = append(payload.Issues, issue)
					return
				}
				if input < minSafeInt {
					issue := CreateTooSmallIssue(
						payload.Value,
						origin,
						minSafeInt,
						false,
						WithContinue(!def.Abort),
					)
					payload.Issues = append(payload.Issues, issue)
					return
				}
			}
		}

		// Check bounds
		if input < minimum {
			issue := CreateTooSmallIssue(
				payload.Value,
				"number",
				minimum,
				true,
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
		}

		if input > maximum {
			issue := CreateTooBigIssue(
				payload.Value,
				"number",
				maximum,
				true,
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
		}
	}

	return &ZodCheckNumberFormat{Internals: internals}
}

// =============================================================================
// STRING AND LENGTH CHECKS
// =============================================================================

type ZodCheckMaxLengthDef struct {
	ZodCheckDef
	Maximum int
}

type ZodCheckMaxLengthInternals struct {
	ZodCheckInternals
	Def *ZodCheckMaxLengthDef
}

// ZodCheckMaxLength validates maximum length constraint
type ZodCheckMaxLength struct {
	Internals *ZodCheckMaxLengthInternals
}

func (c *ZodCheckMaxLength) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckMaxLength creates a maximum length check
func NewZodCheckMaxLength(maximum int, params ...SchemaParams) *ZodCheckMaxLength {
	def := &ZodCheckMaxLengthDef{
		ZodCheckDef: ZodCheckDef{Check: "max_length"},
		Maximum:     maximum,
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckMaxLengthInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set when condition
	internals.When = func(payload *ParsePayload) bool {
		return !nullish(payload.Value) && hasLength(payload.Value)
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	check := &ZodCheckMaxLength{Internals: internals}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		length := getLength(payload.Value)
		if length <= def.Maximum {
			return
		}

		origin := getLengthableOrigin(payload.Value)
		issue := CreateTooBigIssue(
			payload.Value,
			origin,
			def.Maximum,
			true, // String length checks are inclusive in TypeScript Zod v4
			WithContinue(!def.Abort),
			WithInst(check),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return check
}

/////////////////////////////////////
/////    ZodCheckMinLength      /////
/////////////////////////////////////

// ZodCheckMinLengthDef defines min length validation constraint
type ZodCheckMinLengthDef struct {
	ZodCheckDef
	Minimum int
}

// ZodCheckMinLengthInternals contains min length check internal state
type ZodCheckMinLengthInternals struct {
	ZodCheckInternals
	Def *ZodCheckMinLengthDef
}

// ZodCheckMinLength represents min length validation check
type ZodCheckMinLength struct {
	Internals *ZodCheckMinLengthInternals
}

func (c *ZodCheckMinLength) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckMinLength creates a new min length check with optional parameters
func NewZodCheckMinLength(minimum int, params ...SchemaParams) *ZodCheckMinLength {
	def := &ZodCheckMinLengthDef{
		ZodCheckDef: ZodCheckDef{Check: "min_length"},
		Minimum:     minimum,
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckMinLengthInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set when condition
	internals.When = func(payload *ParsePayload) bool {
		return !nullish(payload.Value) && hasLength(payload.Value)
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	check := &ZodCheckMinLength{Internals: internals}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		length := getLength(payload.Value)
		if length >= def.Minimum {
			return
		}

		origin := getLengthableOrigin(payload.Value)
		issue := CreateTooSmallIssue(
			payload.Value,
			origin,
			def.Minimum,
			true, // String length checks are inclusive in TypeScript Zod v4
			WithContinue(!def.Abort),
			WithInst(check),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return check
}

/////////////////////////////////////
/////    ZodCheckStringFormat   /////
/////////////////////////////////////

// ZodStringFormats represents string format types
type ZodStringFormats string

const (
	StringFormatEmail      ZodStringFormats = "email"
	StringFormatURL        ZodStringFormats = "url"
	StringFormatEmoji      ZodStringFormats = "emoji"
	StringFormatUUID       ZodStringFormats = "uuid"
	StringFormatGUID       ZodStringFormats = "guid"
	StringFormatNanoID     ZodStringFormats = "nanoid"
	StringFormatCUID       ZodStringFormats = "cuid"
	StringFormatCUID2      ZodStringFormats = "cuid2"
	StringFormatULID       ZodStringFormats = "ulid"
	StringFormatXID        ZodStringFormats = "xid"
	StringFormatKSUID      ZodStringFormats = "ksuid"
	StringFormatDatetime   ZodStringFormats = "datetime"
	StringFormatDate       ZodStringFormats = "date"
	StringFormatTime       ZodStringFormats = "time"
	StringFormatDuration   ZodStringFormats = "duration"
	StringFormatIPv4       ZodStringFormats = "ipv4"
	StringFormatIPv6       ZodStringFormats = "ipv6"
	StringFormatCIDRv4     ZodStringFormats = "cidrv4"
	StringFormatCIDRv6     ZodStringFormats = "cidrv6"
	StringFormatBase64     ZodStringFormats = "base64"
	StringFormatBase64URL  ZodStringFormats = "base64url"
	StringFormatJSONString ZodStringFormats = "json_string"
	StringFormatE164       ZodStringFormats = "e164"
	StringFormatLowercase  ZodStringFormats = "lowercase"
	StringFormatUppercase  ZodStringFormats = "uppercase"
	StringFormatRegex      ZodStringFormats = "regex"
	StringFormatJWT        ZodStringFormats = "jwt"
	StringFormatStartsWith ZodStringFormats = "starts_with"
	StringFormatEndsWith   ZodStringFormats = "ends_with"
	StringFormatIncludes   ZodStringFormats = "includes"
)

// ZodCheckStringFormatDef defines string format validation constraint
type ZodCheckStringFormatDef struct {
	ZodCheckDef
	Format  ZodStringFormats
	Pattern *regexp.Regexp
}

// ZodCheckStringFormatInternals contains string format check internal state
type ZodCheckStringFormatInternals struct {
	ZodCheckInternals
	Def *ZodCheckStringFormatDef
}

// ZodCheckStringFormat represents string format validation check
type ZodCheckStringFormat struct {
	Internals *ZodCheckStringFormatInternals
}

func (c *ZodCheckStringFormat) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckStringFormat creates a new string format check
func NewZodCheckStringFormat(format ZodStringFormats, pattern *regexp.Regexp, params ...SchemaParams) *ZodCheckStringFormat {
	def := &ZodCheckStringFormatDef{
		ZodCheckDef: ZodCheckDef{Check: "string_format"},
		Format:      format,
		Pattern:     pattern,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckStringFormatInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	check := &ZodCheckStringFormat{Internals: internals}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := CreateInvalidTypeIssue(
				payload.Value,
				"string",
				string(GetParsedType(payload.Value)),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		if def.Pattern.MatchString(input) {
			return
		}

		issue := CreateInvalidFormatIssue(
			payload.Value,
			string(def.Format),
			WithPattern(def.Pattern.String()),
			WithContinue(!def.Abort),
			WithInst(check),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return check
}

////////////////////////////////
/////    ZodCheckRegex     /////
////////////////////////////////

// ZodCheckRegexDef defines regex validation constraint
type ZodCheckRegexDef struct {
	ZodCheckStringFormatDef
}

// ZodCheckRegexInternals contains regex check internal state
type ZodCheckRegexInternals struct {
	ZodCheckInternals
	Def *ZodCheckRegexDef
}

// ZodCheckRegex represents regex validation check
type ZodCheckRegex struct {
	Internals *ZodCheckRegexInternals
}

func (c *ZodCheckRegex) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckRegex creates a new regex check with optional parameters
func NewZodCheckRegex(pattern *regexp.Regexp, params ...SchemaParams) *ZodCheckRegex {
	def := &ZodCheckRegexDef{
		ZodCheckStringFormatDef: ZodCheckStringFormatDef{
			ZodCheckDef: ZodCheckDef{Check: "string_format"},
			Format:      StringFormatRegex,
			Pattern:     pattern,
		},
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckRegexInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := NewRawIssue(
				string(InvalidType),
				payload.Value,
				WithExpected("string"),
				WithReceived(string(GetParsedType(payload.Value))),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		if def.Pattern.MatchString(input) {
			return
		}

		issue := NewRawIssue(
			string(InvalidFormat),
			payload.Value,
			WithOrigin("string"),
			WithFormat("regex"),
			WithPattern(def.Pattern.String()),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckRegex{Internals: internals}
}

///////////////////////////////////
/////    ZodCheckIncludes     /////
///////////////////////////////////

// ZodCheckIncludesDef defines includes validation constraint
type ZodCheckIncludesDef struct {
	ZodCheckStringFormatDef
	Includes string
	Position *int
}

// ZodCheckIncludesInternals contains includes check internal state
type ZodCheckIncludesInternals struct {
	ZodCheckInternals
	Def *ZodCheckIncludesDef
}

// ZodCheckIncludes represents includes validation check
type ZodCheckIncludes struct {
	Internals *ZodCheckIncludesInternals
}

func (c *ZodCheckIncludes) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckIncludes creates a new includes check
func NewZodCheckIncludes(includes string, position *int, params ...SchemaParams) *ZodCheckIncludes {
	pattern := regexp.MustCompile(escapeRegex(includes))

	def := &ZodCheckIncludesDef{
		ZodCheckStringFormatDef: ZodCheckStringFormatDef{
			ZodCheckDef: ZodCheckDef{Check: "string_format"},
			Format:      StringFormatIncludes,
			Pattern:     pattern,
		},
		Includes: includes,
		Position: position,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckIncludesInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := NewRawIssue(
				string(InvalidType),
				payload.Value,
				WithExpected("string"),
				WithReceived(string(GetParsedType(payload.Value))),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		// Check if string includes the substring
		var found bool
		if def.Position == nil {
			found = strings.Contains(input, def.Includes)
		} else {
			// TypeScript: payload.value.includes(def.includes, def.position)
			// This means starting search from position, not substring of input[position:]
			switch {
			case *def.Position >= 0 && *def.Position < len(input):
				found = strings.Contains(input[*def.Position:], def.Includes)
			case *def.Position >= len(input):
				// Position is beyond string length, cannot find substring
				found = false
			default:
				// Position is negative, check from start
				found = strings.Contains(input, def.Includes)
			}
		}

		if found {
			return
		}

		issue := NewRawIssue(
			string(InvalidFormat),
			payload.Value,
			WithOrigin("string"),
			WithFormat("includes"),
			WithIncludes(def.Includes),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckIncludes{Internals: internals}
}

/////////////////////////////////////
/////    ZodCheckStartsWith     /////
/////////////////////////////////////

// ZodCheckStartsWithDef defines starts with validation constraint
type ZodCheckStartsWithDef struct {
	ZodCheckStringFormatDef
	Prefix string
}

// ZodCheckStartsWithInternals contains starts with check internal state
type ZodCheckStartsWithInternals struct {
	ZodCheckInternals
	Def *ZodCheckStartsWithDef
}

// ZodCheckStartsWith represents starts with validation check
type ZodCheckStartsWith struct {
	Internals *ZodCheckStartsWithInternals
}

func (c *ZodCheckStartsWith) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckStartsWith creates a new starts with check with optional parameters
func NewZodCheckStartsWith(prefix string, params ...SchemaParams) *ZodCheckStartsWith {
	pattern := regexp.MustCompile("^" + escapeRegex(prefix) + ".*")

	def := &ZodCheckStartsWithDef{
		ZodCheckStringFormatDef: ZodCheckStringFormatDef{
			ZodCheckDef: ZodCheckDef{Check: "string_format"},
			Format:      StringFormatStartsWith,
			Pattern:     pattern,
		},
		Prefix: prefix,
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckStartsWithInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := NewRawIssue(
				string(InvalidType),
				payload.Value,
				WithExpected("string"),
				WithReceived(string(GetParsedType(payload.Value))),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		if strings.HasPrefix(input, def.Prefix) {
			return
		}

		issue := NewRawIssue(
			string(InvalidFormat),
			payload.Value,
			WithOrigin("string"),
			WithFormat("starts_with"),
			WithPrefix(def.Prefix),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckStartsWith{Internals: internals}
}

//////////////////////////////////
/////   ZodCheckEndsWith     /////
//////////////////////////////////

// ZodCheckEndsWithDef defines ends with validation constraint
type ZodCheckEndsWithDef struct {
	ZodCheckStringFormatDef
	Suffix string
}

// ZodCheckEndsWithInternals contains ends with check internal state
type ZodCheckEndsWithInternals struct {
	ZodCheckInternals
	Def *ZodCheckEndsWithDef
}

// ZodCheckEndsWith represents ends with validation check
type ZodCheckEndsWith struct {
	Internals *ZodCheckEndsWithInternals
}

func (c *ZodCheckEndsWith) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckEndsWith creates a new ends with check with optional parameters
func NewZodCheckEndsWith(suffix string, params ...SchemaParams) *ZodCheckEndsWith {
	pattern := regexp.MustCompile(".*" + escapeRegex(suffix) + "$")

	def := &ZodCheckEndsWithDef{
		ZodCheckStringFormatDef: ZodCheckStringFormatDef{
			ZodCheckDef: ZodCheckDef{Check: "string_format"},
			Format:      StringFormatEndsWith,
			Pattern:     pattern,
		},
		Suffix: suffix,
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckEndsWithInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := NewRawIssue(
				string(InvalidType),
				payload.Value,
				WithExpected("string"),
				WithReceived(string(GetParsedType(payload.Value))),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		if strings.HasSuffix(input, def.Suffix) {
			return
		}

		issue := NewRawIssue(
			string(InvalidFormat),
			payload.Value,
			WithOrigin("string"),
			WithFormat("ends_with"),
			WithSuffix(def.Suffix),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckEndsWith{Internals: internals}
}

///////////////////////////////////
/////    ZodCheckOverwrite    /////
///////////////////////////////////

// ZodCheckOverwriteDef defines overwrite validation constraint
type ZodCheckOverwriteDef struct {
	ZodCheckDef
	Transform func(interface{}) interface{}
}

// ZodCheckOverwriteInternals contains overwrite check internal state
type ZodCheckOverwriteInternals struct {
	ZodCheckInternals
	Def *ZodCheckOverwriteDef
}

// ZodCheckOverwrite represents overwrite validation check
type ZodCheckOverwrite struct {
	Internals *ZodCheckOverwriteInternals
}

func (c *ZodCheckOverwrite) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckOverwrite creates a new overwrite check
func NewZodCheckOverwrite(transform func(interface{}) interface{}, params ...SchemaParams) *ZodCheckOverwrite {
	def := &ZodCheckOverwriteDef{
		ZodCheckDef: ZodCheckDef{Check: "overwrite"},
		Transform:   transform,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckOverwriteInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		payload.Value = def.Transform(payload.Value)
	}

	return &ZodCheckOverwrite{Internals: internals}
}

//////////////////////////////////////
/////    ZodCheckLowerCase       /////
//////////////////////////////////////

// ZodCheckLowerCaseDef defines lowercase validation constraint
type ZodCheckLowerCaseDef struct {
	ZodCheckStringFormatDef
}

// ZodCheckLowerCaseInternals contains lowercase check internal state
type ZodCheckLowerCaseInternals struct {
	ZodCheckInternals
	Def *ZodCheckLowerCaseDef
}

// ZodCheckLowerCase represents lowercase validation check
type ZodCheckLowerCase struct {
	Internals *ZodCheckLowerCaseInternals
}

func (c *ZodCheckLowerCase) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckLowerCase creates a new lowercase check
func NewZodCheckLowerCase(params ...SchemaParams) *ZodCheckLowerCase {
	// Use lowercase pattern from regexes package
	pattern := regexp.MustCompile(`^[a-z]*$`)

	def := &ZodCheckLowerCaseDef{
		ZodCheckStringFormatDef: ZodCheckStringFormatDef{
			ZodCheckDef: ZodCheckDef{Check: "string_format"},
			Format:      StringFormatLowercase,
			Pattern:     pattern,
		},
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckLowerCaseInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Use string format check logic
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := NewRawIssue(
				string(InvalidType),
				payload.Value,
				WithExpected("string"),
				WithReceived(string(GetParsedType(payload.Value))),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		if def.Pattern.MatchString(input) {
			return
		}

		issue := NewRawIssue(
			string(InvalidFormat),
			payload.Value,
			WithOrigin("string"),
			WithFormat("lowercase"),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckLowerCase{Internals: internals}
}

//////////////////////////////////////
/////    ZodCheckUpperCase       /////
//////////////////////////////////////

// ZodCheckUpperCaseDef defines uppercase validation constraint
type ZodCheckUpperCaseDef struct {
	ZodCheckStringFormatDef
}

// ZodCheckUpperCaseInternals contains uppercase check internal state
type ZodCheckUpperCaseInternals struct {
	ZodCheckInternals
	Def *ZodCheckUpperCaseDef
}

// ZodCheckUpperCase represents uppercase validation check
type ZodCheckUpperCase struct {
	Internals *ZodCheckUpperCaseInternals
}

func (c *ZodCheckUpperCase) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckUpperCase creates a new uppercase check
func NewZodCheckUpperCase(params ...SchemaParams) *ZodCheckUpperCase {
	// Use uppercase pattern from regexes package
	pattern := regexp.MustCompile(`^[A-Z]*$`)

	def := &ZodCheckUpperCaseDef{
		ZodCheckStringFormatDef: ZodCheckStringFormatDef{
			ZodCheckDef: ZodCheckDef{Check: "string_format"},
			Format:      StringFormatUppercase,
			Pattern:     pattern,
		},
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckUpperCaseInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Use string format check logic
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := NewRawIssue(
				string(InvalidType),
				payload.Value,
				WithExpected("string"),
				WithReceived(string(GetParsedType(payload.Value))),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		if def.Pattern.MatchString(input) {
			return
		}

		issue := NewRawIssue(
			string(InvalidFormat),
			payload.Value,
			WithOrigin("string"),
			WithFormat("uppercase"),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckUpperCase{Internals: internals}
}

//////////////////////////////////////
/////    ZodCheckLengthEquals   /////
//////////////////////////////////////

// ZodCheckLengthEqualsDef defines length equals validation constraint
type ZodCheckLengthEqualsDef struct {
	ZodCheckDef
	Length int
}

// ZodCheckLengthEqualsInternals contains length equals check internal state
type ZodCheckLengthEqualsInternals struct {
	ZodCheckInternals
	Def *ZodCheckLengthEqualsDef
}

// ZodCheckLengthEquals represents length equals validation check
type ZodCheckLengthEquals struct {
	Internals *ZodCheckLengthEqualsInternals
}

func (c *ZodCheckLengthEquals) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckLengthEquals creates a new length equals check with optional parameters
func NewZodCheckLengthEquals(length int, params ...SchemaParams) *ZodCheckLengthEquals {
	def := &ZodCheckLengthEqualsDef{
		ZodCheckDef: ZodCheckDef{Check: "length_equals"},
		Length:      length,
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckLengthEqualsInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set when condition
	internals.When = func(payload *ParsePayload) bool {
		return !nullish(payload.Value) && hasLength(payload.Value)
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		inputLength := getLength(payload.Value)
		if inputLength == def.Length {
			return
		}

		origin := getLengthableOrigin(payload.Value)
		tooBig := inputLength > def.Length

		var issue ZodRawIssue
		if tooBig {
			issue = NewRawIssue(
				string(TooBig),
				payload.Value,
				WithOrigin(origin),
				WithMaximum(def.Length),
				WithContinue(!def.Abort),
			)
		} else {
			issue = NewRawIssue(
				string(TooSmall),
				payload.Value,
				WithOrigin(origin),
				WithMinimum(def.Length),
				WithContinue(!def.Abort),
			)
		}
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckLengthEquals{Internals: internals}
}

//////////////////////////////////
/////    ZodCheckMaxSize     /////
//////////////////////////////////

// ZodCheckMaxSizeDef defines max size validation constraint
type ZodCheckMaxSizeDef struct {
	ZodCheckDef
	Maximum int
}

// ZodCheckMaxSizeInternals contains max size check internal state
type ZodCheckMaxSizeInternals struct {
	ZodCheckInternals
	Def *ZodCheckMaxSizeDef
}

// ZodCheckMaxSize represents max size validation check
type ZodCheckMaxSize struct {
	Internals *ZodCheckMaxSizeInternals
}

func (c *ZodCheckMaxSize) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckMaxSize creates a new max size check with optional parameters
func NewZodCheckMaxSize(maximum int, params ...SchemaParams) *ZodCheckMaxSize {
	def := &ZodCheckMaxSizeDef{
		ZodCheckDef: ZodCheckDef{Check: "max_size"},
		Maximum:     maximum,
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckMaxSizeInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set when condition
	internals.When = func(payload *ParsePayload) bool {
		return !nullish(payload.Value) && hasSize(payload.Value)
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		size := getSize(payload.Value)
		if size <= def.Maximum {
			return
		}

		origin := getSizableOrigin(payload.Value)
		issue := NewRawIssue(
			string(TooBig),
			payload.Value,
			WithOrigin(origin),
			WithMaximum(def.Maximum),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckMaxSize{Internals: internals}
}

//////////////////////////////////
/////    ZodCheckMinSize     /////
//////////////////////////////////

// ZodCheckMinSizeDef defines min size validation constraint
type ZodCheckMinSizeDef struct {
	ZodCheckDef
	Minimum int
}

// ZodCheckMinSizeInternals contains min size check internal state
type ZodCheckMinSizeInternals struct {
	ZodCheckInternals
	Def *ZodCheckMinSizeDef
}

// ZodCheckMinSize represents min size validation check
type ZodCheckMinSize struct {
	Internals *ZodCheckMinSizeInternals
}

func (c *ZodCheckMinSize) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckMinSize creates a new min size check with optional parameters
func NewZodCheckMinSize(minimum int, params ...SchemaParams) *ZodCheckMinSize {
	def := &ZodCheckMinSizeDef{
		ZodCheckDef: ZodCheckDef{Check: "min_size"},
		Minimum:     minimum,
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckMinSizeInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set when condition
	internals.When = func(payload *ParsePayload) bool {
		return !nullish(payload.Value) && hasSize(payload.Value)
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		size := getSize(payload.Value)
		if size >= def.Minimum {
			return
		}

		origin := getSizableOrigin(payload.Value)
		issue := NewRawIssue(
			string(TooSmall),
			payload.Value,
			WithOrigin(origin),
			WithMinimum(def.Minimum),
			WithContinue(!def.Abort),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckMinSize{Internals: internals}
}

//////////////////////////////////
/////  ZodCheckSizeEquals    /////
//////////////////////////////////

// ZodCheckSizeEqualsDef defines size equals validation constraint
type ZodCheckSizeEqualsDef struct {
	ZodCheckDef
	Size int
}

// ZodCheckSizeEqualsInternals contains size equals check internal state
type ZodCheckSizeEqualsInternals struct {
	ZodCheckInternals
	Def *ZodCheckSizeEqualsDef
}

// ZodCheckSizeEquals represents size equals validation check
type ZodCheckSizeEquals struct {
	Internals *ZodCheckSizeEqualsInternals
}

func (c *ZodCheckSizeEquals) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckSizeEquals creates a new size equals check with optional parameters
func NewZodCheckSizeEquals(size int, params ...SchemaParams) *ZodCheckSizeEquals {
	def := &ZodCheckSizeEqualsDef{
		ZodCheckDef: ZodCheckDef{Check: "size_equals"},
		Size:        size,
	}

	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckSizeEqualsInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set when condition
	internals.When = func(payload *ParsePayload) bool {
		return !nullish(payload.Value) && hasSize(payload.Value)
	}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		actualSize := getSize(payload.Value)
		if actualSize == def.Size {
			return
		}

		origin := getSizableOrigin(payload.Value)
		tooBig := actualSize > def.Size

		var issue ZodRawIssue
		if tooBig {
			issue = NewRawIssue(
				string(TooBig),
				payload.Value,
				WithOrigin(origin),
				WithMaximum(def.Size),
				WithContinue(!def.Abort),
			)
		} else {
			issue = NewRawIssue(
				string(TooSmall),
				payload.Value,
				WithOrigin(origin),
				WithMinimum(def.Size),
				WithContinue(!def.Abort),
			)
		}
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckSizeEquals{Internals: internals}
}

///////////////////////////////////
/////    ZodCheckProperty     /////
///////////////////////////////////

// ZodCheckPropertyDef defines property validation constraint
type ZodCheckPropertyDef struct {
	ZodCheckDef
	Property string
	Schema   interface{} // schemas.$ZodType equivalent
}

// ZodCheckPropertyInternals contains property check internal state
type ZodCheckPropertyInternals struct {
	ZodCheckInternals
	Def *ZodCheckPropertyDef
}

// ZodCheckProperty represents property validation check
type ZodCheckProperty struct {
	Internals *ZodCheckPropertyInternals
}

func (c *ZodCheckProperty) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// handleCheckPropertyResult handles the result of property validation
func handleCheckPropertyResult(result *ParsePayload, payload *ParsePayload, property string) {
	if len(result.Issues) > 0 {
		// Use PrefixIssues utility function
		prefixedIssues := prefixIssues(property, result.Issues)
		payload.Issues = append(payload.Issues, prefixedIssues...)
	}
}

// PrefixIssues adds path prefix to issues
func prefixIssues(path interface{}, issues []ZodRawIssue) []ZodRawIssue {
	for i := range issues {
		// Ensure path exists
		if issues[i].Path == nil {
			issues[i].Path = make([]interface{}, 0)
		}
		// Prepend the path
		newPath := make([]interface{}, len(issues[i].Path)+1)
		newPath[0] = path
		copy(newPath[1:], issues[i].Path)
		issues[i].Path = newPath
	}
	return issues
}

// NewZodCheckProperty creates a new property check with optional parameters
func NewZodCheckProperty(property string, schema interface{}, params ...SchemaParams) *ZodCheckProperty {
	def := &ZodCheckPropertyDef{
		ZodCheckDef: ZodCheckDef{Check: "property"},
		Property:    property,
		Schema:      schema,
	}
	// Apply normalized parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckPropertyInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set check function following TypeScript implementation exactly
	internals.Check = func(payload *ParsePayload) {
		// Extract property value: (payload.value as any)[def.property]
		var propertyValue interface{}
		if objMap, ok := payload.Value.(map[string]interface{}); ok {
			propertyValue = objMap[def.Property]
		} else {
			// Property doesn't exist on non-object
			issue := NewRawIssue(
				string(InvalidType),
				payload.Value,
				WithExpected("object"),
				WithReceived(string(GetParsedType(payload.Value))),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		// If no schema provided, just check property exists
		if def.Schema == nil {
			if propertyValue == nil {
				issue := NewRawIssue(
					string(InvalidType),
					payload.Value,
					WithExpected("defined"),
					WithReceived("undefined"),
					WithPath([]interface{}{def.Property}),
					WithContinue(!def.Abort),
				)
				payload.Issues = append(payload.Issues, issue)
			}
			return
		}

		// Create result payload: { value: ..., issues: [] }
		result := &ParsePayload{
			Value:  propertyValue,
			Issues: []ZodRawIssue{},
		}

		// Run schema validation: def.schema._zod.run(result, {})
		if Schema, ok := def.Schema.(ZodType[any, any]); ok {
			schemaInternals := Schema.GetInternals()
			if schemaInternals.Parse != nil {
				ctx := NewParseContext()
				result = schemaInternals.Parse(result, ctx)

				// Run checks if any exist
				if len(schemaInternals.Checks) > 0 {
					result = runChecks(result, schemaInternals.Checks, ctx)
				}
			}
		}

		// Handle result: handleCheckPropertyResult(result, payload, def.property)
		handleCheckPropertyResult(result, payload, def.Property)
	}

	return &ZodCheckProperty{Internals: internals}
}

// =============================================================================
// PARAMETER HANDLING UTILITIES
// =============================================================================

// ApplySchemaParams applies schema parameters to check definition
func ApplySchemaParams(def *ZodCheckDef, params ...SchemaParams) {
	if len(params) == 0 {
		return
	}

	param := params[0]

	if param.Error != nil {
		switch err := param.Error.(type) {
		case string:
			errorMsg := err
			errorMap := ZodErrorMap(func(ZodRawIssue) string {
				return errorMsg
			})
			def.Error = &errorMap
		case ZodErrorMap:
			def.Error = &err
		case *ZodErrorMap:
			def.Error = err
		case func(ZodRawIssue) string:
			errorMap := ZodErrorMap(err)
			def.Error = &errorMap
		}
	}

	if param.Abort {
		def.Abort = param.Abort
	}
}

///////////////////////////////////////
/////     MAP SIZE CHECKS         ////
///////////////////////////////////////

// NewZodCheckMapSize creates a size check specifically for maps
func NewZodCheckMapSize(size int, params ...SchemaParams) ZodCheck {
	def := &ZodCheckSizeEqualsDef{
		ZodCheckDef: ZodCheckDef{Check: "size_equals"},
		Size:        size,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckSizeEqualsInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set up check function with "map" origin
	internals.Check = func(payload *ParsePayload) {
		if !hasSize(payload.Value) {
			issue := CreateInvalidTypeIssue(payload.Value, "map", string(GetParsedType(payload.Value)), func(issue *ZodRawIssue) {
				issue.Inst = internals
			})
			payload.Issues = append(payload.Issues, issue)
			return
		}

		currentSize := getSize(payload.Value)
		if currentSize == def.Size {
			return
		}

		if currentSize < def.Size {
			issue := NewRawIssue(
				string(TooSmall),
				payload.Value,
				WithOrigin("map"),
				WithMinimum(def.Size),
				WithContinue(!def.Abort),
			)
			issue.Inst = internals
			payload.Issues = append(payload.Issues, issue)
		} else {
			issue := NewRawIssue(
				string(TooBig),
				payload.Value,
				WithOrigin("map"),
				WithMaximum(def.Size),
				WithContinue(!def.Abort),
			)
			issue.Inst = internals
			payload.Issues = append(payload.Issues, issue)
		}
	}

	return &ZodCheckSizeEquals{Internals: internals}
}

// NewZodCheckMapMinSize creates a minimum size check specifically for maps
func NewZodCheckMapMinSize(minimum int, params ...SchemaParams) ZodCheck {
	def := &ZodCheckMinSizeDef{
		ZodCheckDef: ZodCheckDef{Check: "min_size"},
		Minimum:     minimum,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckMinSizeInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set up check function with "map" origin
	internals.Check = func(payload *ParsePayload) {
		if !hasSize(payload.Value) {
			issue := CreateInvalidTypeIssue(payload.Value, "map", string(GetParsedType(payload.Value)), func(issue *ZodRawIssue) {
				issue.Inst = internals
			})
			payload.Issues = append(payload.Issues, issue)
			return
		}

		currentSize := getSize(payload.Value)
		if currentSize >= def.Minimum {
			return
		}

		issue := NewRawIssue(
			string(TooSmall),
			payload.Value,
			WithOrigin("map"),
			WithMinimum(def.Minimum),
			WithReceived(fmt.Sprintf("%d", currentSize)),
			WithContinue(!def.Abort),
		)
		issue.Inst = internals
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckMinSize{Internals: internals}
}

// NewZodCheckMapMaxSize creates a maximum size check specifically for maps
func NewZodCheckMapMaxSize(maximum int, params ...SchemaParams) ZodCheck {
	def := &ZodCheckMaxSizeDef{
		ZodCheckDef: ZodCheckDef{Check: "max_size"},
		Maximum:     maximum,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckMaxSizeInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set up check function with "map" origin
	internals.Check = func(payload *ParsePayload) {
		if !hasSize(payload.Value) {
			issue := CreateInvalidTypeIssue(payload.Value, "map", string(GetParsedType(payload.Value)), func(issue *ZodRawIssue) {
				issue.Inst = internals
			})
			payload.Issues = append(payload.Issues, issue)
			return
		}

		currentSize := getSize(payload.Value)
		if currentSize <= def.Maximum {
			return
		}

		issue := NewRawIssue(
			string(TooBig),
			payload.Value,
			WithOrigin("map"),
			WithMaximum(def.Maximum),
			WithReceived(fmt.Sprintf("%d", currentSize)),
			WithContinue(!def.Abort),
		)
		issue.Inst = internals
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckMaxSize{Internals: internals}
}

// =============================================================================
// CUSTOM VALIDATION CHECKS
// =============================================================================

// ZodCheckCustomDef defines custom validation check configuration
type ZodCheckCustomDef struct {
	ZodCheckDef
	Type   string
	Path   []interface{}
	Params map[string]interface{}
	Fn     interface{} // RefineFn or CheckFn
	FnType string      // "refine" or "check"
}

type ZodCheckCustomInternals struct {
	ZodCheckInternals
	Def  *ZodCheckCustomDef
	Issc *ZodIssueBase
	Bag  map[string]interface{}
}

// ZodCheckCustom represents custom validation check
type ZodCheckCustom struct {
	Internals *ZodCheckCustomInternals
}

func (z *ZodCheckCustom) GetZod() *ZodCheckInternals {
	return &z.Internals.ZodCheckInternals
}

// NewCustom creates a custom validation check
func NewCustom[T any](fn interface{}, params ...SchemaParams) *ZodCheckCustom {
	def := &ZodCheckCustomDef{
		ZodCheckDef: ZodCheckDef{
			Check: "custom",
		},
		Type:   "custom",
		Params: make(map[string]interface{}),
	}

	// Determine function type and store
	switch f := fn.(type) {
	case CheckFn:
		def.Fn = f
		def.FnType = "check"
	case func(*ParsePayload):
		def.Fn = CheckFn(f)
		def.FnType = "check"
	case RefineFn[string]:
		def.Fn = f
		def.FnType = "refine"
	case func(string) bool:
		def.Fn = RefineFn[string](f)
		def.FnType = "refine"
	case RefineFn[map[string]interface{}]:
		def.Fn = f
		def.FnType = "refine"
	case func(map[string]interface{}) bool:
		def.Fn = RefineFn[map[string]interface{}](f)
		def.FnType = "refine"
	case RefineFn[interface{}]:
		def.Fn = f
		def.FnType = "refine"
	case func(interface{}) bool:
		def.Fn = RefineFn[interface{}](f)
		def.FnType = "refine"
	default:
		panic("Invalid function type for NewCustom check")
	}

	// Apply parameters
	ApplySchemaParams(&def.ZodCheckDef, params...)

	// Handle path parameter
	if len(params) > 0 && params[0].Path != nil {
		def.Path = make([]interface{}, len(params[0].Path))
		for i, p := range params[0].Path {
			def.Path[i] = p
		}
	}

	// Handle params parameter
	if len(params) > 0 && params[0].Params != nil {
		def.Params = make(map[string]interface{})
		for k, v := range params[0].Params {
			def.Params[k] = v
		}
	}

	internals := &ZodCheckCustomInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
		Bag:               make(map[string]interface{}),
	}

	// Set up check function
	internals.Check = func(payload *ParsePayload) {
		executeCustomCheck(payload, internals)
	}

	return &ZodCheckCustom{Internals: internals}
}

// handleRefineResult handles refine function validation results with strong typing
func handleRefineResult(result bool, payload *ParsePayload, input interface{}, internals *ZodCheckCustomInternals) {
	// Check if result is false (validation failed)
	if !result {
		// Construct error path
		path := make([]interface{}, len(payload.Path))
		copy(path, payload.Path)

		// Add custom path
		if internals.Def.Path != nil {
			path = append(path, internals.Def.Path...)
		}

		// Create error options
		options := []func(*ZodRawIssue){
			WithOrigin("custom"),
			WithPath(path),
			WithContinue(!internals.Def.Abort),
		}

		// Add custom parameters if provided
		if len(internals.Def.Params) > 0 {
			options = append(options, WithParams(internals.Def.Params))
		}

		// Create error
		issue := NewRawIssue("custom", input, options...)

		// Set the Inst field so FinalizeIssue can access schema-level error mapping
		issue.Inst = internals

		payload.Issues = append(payload.Issues, issue)
	}
}

// Note: isTruthy function removed as we now use strong-typed boolean returns
// This eliminates runtime type checking and improves performance

// executeCustomCheck executes custom validation check with strong typing
func executeCustomCheck(payload *ParsePayload, internals *ZodCheckCustomInternals) {
	switch internals.Def.FnType {
	case "refine":
		// Execute refine function with strong typing
		switch fn := internals.Def.Fn.(type) {
		case RefineFn[string]:
			if str, ok := payload.Value.(string); ok {
				result := fn(str)
				handleRefineResult(result, payload, payload.Value, internals)
			} else {
				panic("Type mismatch: expected string")
			}
		case RefineFn[map[string]interface{}]:
			if mapData, ok := payload.Value.(map[string]interface{}); ok {
				result := fn(mapData)
				handleRefineResult(result, payload, payload.Value, internals)
			} else {
				panic("Type mismatch: expected map[string]interface{}")
			}
		case RefineFn[interface{}]:
			// Handle interface{} type - accepts any value
			result := fn(payload.Value)
			handleRefineResult(result, payload, payload.Value, internals)
		default:
			panic("Invalid refine function type")
		}

	case "check":
		// Execute check function
		// The CheckFn now has access to payload.AddIssue() method for ctx.issues.push() functionality
		checkFn := internals.Def.Fn.(CheckFn)
		checkFn(payload)

	default:
		panic("Unknown custom function type: " + internals.Def.FnType)
	}
}

///////////////////////////////////
/////    ZodCheckMimeType     /////
///////////////////////////////////

// ZodCheckMimeTypeDef defines MIME type validation constraint
type ZodCheckMimeTypeDef struct {
	ZodCheckDef
	Mime []string // util.MimeTypes[] equivalent
}

// ZodCheckMimeTypeInternals contains MIME type check internal state
type ZodCheckMimeTypeInternals struct {
	ZodCheckInternals
	Def *ZodCheckMimeTypeDef
}

// ZodCheckMimeType represents MIME type validation check
type ZodCheckMimeType struct {
	Internals *ZodCheckMimeTypeInternals
}

func (c *ZodCheckMimeType) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckMimeType creates a new MIME type check
func NewZodCheckMimeType(mimeTypes []string, params ...SchemaParams) *ZodCheckMimeType {
	def := &ZodCheckMimeTypeDef{
		ZodCheckDef: ZodCheckDef{Check: "mime_type"},
		Mime:        mimeTypes,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckMimeTypeInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Create MIME type set for efficient lookup
	mimeSet := make(map[string]struct{})
	for _, mime := range mimeTypes {
		mimeSet[mime] = struct{}{}
	}

	// Set onattach callback
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// TypeScript's bag equivalent would be handled in schema
	})

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		// Extract file type from payload
		var fileType string
		var hasContentType bool

		switch file := payload.Value.(type) {
		case *multipart.FileHeader:
			// Get Content-Type from multipart file header
			if file.Header != nil {
				if contentType := file.Header.Get("Content-Type"); contentType != "" {
					fileType = contentType
					hasContentType = true
				}
			}
			if !hasContentType {
				fileType = "application/octet-stream" // Default MIME type
			}
		case *os.File:
			// os.File has no built-in MIME type, use default value
			fileType = "application/octet-stream"
		default:
			// Not a file type, should have been caught by file type validation
			issue := CreateInvalidTypeIssue(
				payload.Value,
				"file",
				string(GetParsedType(payload.Value)),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		// Check if MIME type is allowed
		if _, allowed := mimeSet[fileType]; allowed {
			return
		}

		// Create invalid value issue with path ["type"]
		values := make([]interface{}, len(def.Mime))
		for i, mime := range def.Mime {
			values[i] = mime
		}

		issue := NewRawIssue(
			string(InvalidValue),
			fileType,
			WithOrigin("file"),
			WithPath([]interface{}{"type"}),
			WithValues(values),
			WithContinue(!def.Abort),
		)
		issue.Inst = internals
		payload.Issues = append(payload.Issues, issue)
	}

	return &ZodCheckMimeType{Internals: internals}
}

///////////////////////////////////
/////    ZodCheckISODateMin   /////
///////////////////////////////////

// ZodCheckISODateMinDef defines ISO date minimum validation constraint
type ZodCheckISODateMinDef struct {
	ZodCheckDef
	MinDate string // ISO date string (YYYY-MM-DD)
}

// ZodCheckISODateMinInternals contains ISO date min check internal state
type ZodCheckISODateMinInternals struct {
	ZodCheckInternals
	Def *ZodCheckISODateMinDef
}

// ZodCheckISODateMin represents ISO date minimum validation check
type ZodCheckISODateMin struct {
	Internals *ZodCheckISODateMinInternals
}

func (c *ZodCheckISODateMin) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckISODateMin creates a new ISO date minimum check
func NewZodCheckISODateMin(minDate string, params ...SchemaParams) *ZodCheckISODateMin {
	def := &ZodCheckISODateMinDef{
		ZodCheckDef: ZodCheckDef{Check: "iso_date_min"},
		MinDate:     minDate,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckISODateMinInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set onattach callback to update bag
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// Update bag.minimum for getter methods
		if zodType, ok := schema.(ZodType[any, any]); ok {
			internals := zodType.GetInternals()
			if internals.Bag == nil {
				internals.Bag = make(map[string]interface{})
			}
			internals.Bag["minimum"] = minDate
		}
	})

	check := &ZodCheckISODateMin{Internals: internals}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := CreateInvalidTypeIssue(
				payload.Value,
				"string",
				string(GetParsedType(payload.Value)),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		// Use string comparison for ISO dates (YYYY-MM-DD format ensures lexicographic order)
		if input >= def.MinDate {
			return
		}

		issue := CreateTooSmallIssue(
			payload.Value,
			"date",
			def.MinDate,
			true, // inclusive
			WithContinue(!def.Abort),
			WithInst(check),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return check
}

///////////////////////////////////
/////    ZodCheckISODateMax   /////
///////////////////////////////////

// ZodCheckISODateMaxDef defines ISO date maximum validation constraint
type ZodCheckISODateMaxDef struct {
	ZodCheckDef
	MaxDate string // ISO date string (YYYY-MM-DD)
}

// ZodCheckISODateMaxInternals contains ISO date max check internal state
type ZodCheckISODateMaxInternals struct {
	ZodCheckInternals
	Def *ZodCheckISODateMaxDef
}

// ZodCheckISODateMax represents ISO date maximum validation check
type ZodCheckISODateMax struct {
	Internals *ZodCheckISODateMaxInternals
}

func (c *ZodCheckISODateMax) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckISODateMax creates a new ISO date maximum check
func NewZodCheckISODateMax(maxDate string, params ...SchemaParams) *ZodCheckISODateMax {
	def := &ZodCheckISODateMaxDef{
		ZodCheckDef: ZodCheckDef{Check: "iso_date_max"},
		MaxDate:     maxDate,
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckISODateMaxInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set onattach callback to update bag
	internals.OnAttach = append(internals.OnAttach, func(schema interface{}) {
		// Update bag.maximum for getter methods
		if zodType, ok := schema.(ZodType[any, any]); ok {
			internals := zodType.GetInternals()
			if internals.Bag == nil {
				internals.Bag = make(map[string]interface{})
			}
			internals.Bag["maximum"] = maxDate
		}
	})

	check := &ZodCheckISODateMax{Internals: internals}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := CreateInvalidTypeIssue(
				payload.Value,
				"string",
				string(GetParsedType(payload.Value)),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		// Use string comparison for ISO dates (YYYY-MM-DD format ensures lexicographic order)
		if input <= def.MaxDate {
			return
		}

		issue := CreateTooBigIssue(
			payload.Value,
			"date",
			def.MaxDate,
			true, // inclusive
			WithContinue(!def.Abort),
			WithInst(check),
		)
		payload.Issues = append(payload.Issues, issue)
	}

	return check
}

///////////////////////////////////
/////    ZodCheckJSONString   /////
///////////////////////////////////

// ZodCheckJSONStringDef defines JSON string validation constraint
type ZodCheckJSONStringDef struct {
	ZodCheckDef
}

// ZodCheckJSONStringInternals contains JSON string check internal state
type ZodCheckJSONStringInternals struct {
	ZodCheckInternals
	Def *ZodCheckJSONStringDef
}

// ZodCheckJSONString represents JSON string validation check
type ZodCheckJSONString struct {
	Internals *ZodCheckJSONStringInternals
}

func (c *ZodCheckJSONString) GetZod() *ZodCheckInternals {
	return &c.Internals.ZodCheckInternals
}

// NewZodCheckJSONString creates a new JSON string validation check
func NewZodCheckJSONString(params ...SchemaParams) *ZodCheckJSONString {
	def := &ZodCheckJSONStringDef{
		ZodCheckDef: ZodCheckDef{Check: "json_string"},
	}

	// Apply schema parameters to definition
	ApplySchemaParams(&def.ZodCheckDef, params...)

	internals := &ZodCheckJSONStringInternals{
		ZodCheckInternals: *NewZodCheckInternals(&def.ZodCheckDef),
		Def:               def,
	}

	// Set check function
	internals.Check = func(payload *ParsePayload) {
		input, ok := payload.Value.(string)
		if !ok {
			issue := CreateInvalidTypeIssue(
				payload.Value,
				"string",
				string(GetParsedType(payload.Value)),
				WithContinue(!def.Abort),
			)
			payload.Issues = append(payload.Issues, issue)
			return
		}

		// Validate JSON by attempting to parse it using go-json-experiment
		var jsonData interface{}
		if err := expjson.Unmarshal([]byte(input), &jsonData); err != nil {
			// Create invalid format issue for invalid JSON
			issue := CreateInvalidFormatIssue(
				payload.Value,
				"json",
				WithContinue(!def.Abort),
			)

			payload.Issues = append(payload.Issues, issue)
		}
	}

	return &ZodCheckJSONString{Internals: internals}
}
