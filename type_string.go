package gozod

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod/regexes"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodStringDef defines the configuration for string validation
type ZodStringDef struct {
	ZodTypeDef
	Type   string     // "string"
	Checks []ZodCheck // String-specific validation checks
}

// ZodStringInternals contains string validator internal state
type ZodStringInternals struct {
	ZodTypeInternals
	Def     *ZodStringDef          // Schema definition
	Checks  []ZodCheck             // Validation checks
	Isst    ZodIssueInvalidType    // Invalid type issue template
	Pattern *regexp.Regexp         // Regex pattern (if any)
	Values  map[string]struct{}    // Allowed string values set
	Bag     map[string]interface{} // Additional metadata (formats, etc.)
}

// ZodString represents a string validation schema with type safety
type ZodString struct {
	internals *ZodStringInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodString) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for string type conversion
func (z *ZodString) Coerce(input interface{}) (interface{}, bool) {
	return coerceToString(input)
}

// Parse validates and parses input with smart type inference
func (z *ZodString) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return parseType[string](
		input,
		&z.internals.ZodTypeInternals,
		"string",
		func(v any) (string, bool) { str, ok := v.(string); return str, ok },
		func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
		validateString,
		coerceToString,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodString) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Basic Length Validation

// Min adds minimum length validation
func (z *ZodString) Min(minLen int, params ...SchemaParams) *ZodString {
	check := NewZodCheckMinLength(minLen, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Max adds maximum length validation
func (z *ZodString) Max(maxLen int, params ...SchemaParams) *ZodString {
	check := NewZodCheckMaxLength(maxLen, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Length adds exact length validation
func (z *ZodString) Length(length int, params ...SchemaParams) *ZodString {
	check := NewZodCheckLengthEquals(length, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Format Validation

// Email adds email format validation
func (z *ZodString) Email(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatEmail, regexes.Email, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// URL adds URL format validation
func (z *ZodString) URL(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatURL, regexes.URL, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// UUID adds UUID format validation
func (z *ZodString) UUID(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatUUID, regexes.UUID, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Base64 adds base64 format validation
func (z *ZodString) Base64(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatBase64, regexes.Base64, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Emoji adds emoji format validation
func (z *ZodString) Emoji(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatEmoji, regexes.Emoji, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// CUID adds CUID format validation
func (z *ZodString) CUID(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatCUID, regexes.CUID, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// CUID2 adds CUID2 format validation
func (z *ZodString) CUID2(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatCUID2, regexes.CUID2, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// ULID adds ULID format validation
func (z *ZodString) ULID(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatULID, regexes.ULID, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// NanoID adds NanoID format validation
func (z *ZodString) NanoID(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatNanoID, regexes.NanoID, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// XID adds XID format validation
func (z *ZodString) XID(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatXID, regexes.XID, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// KSUID adds KSUID format validation
func (z *ZodString) KSUID(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatKSUID, regexes.KSUID, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Base64URL adds base64url format validation
func (z *ZodString) Base64URL(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatBase64URL, regexes.Base64URL, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// E164 adds E164 phone number format validation
func (z *ZodString) E164(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatE164, regexes.E164, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// JWT adds JWT format validation
func (z *ZodString) JWT(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatJWT, regexes.JWT, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Pattern Validation

// Regex adds custom regex validation
func (z *ZodString) Regex(pattern *regexp.Regexp, params ...SchemaParams) *ZodString {
	check := NewZodCheckRegex(pattern, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// StartsWith adds prefix validation
func (z *ZodString) StartsWith(prefix string, params ...SchemaParams) *ZodString {
	check := NewZodCheckStartsWith(prefix, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// EndsWith adds suffix validation
func (z *ZodString) EndsWith(suffix string, params ...SchemaParams) *ZodString {
	check := NewZodCheckEndsWith(suffix, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Includes adds substring validation
func (z *ZodString) Includes(substring string, params ...SchemaParams) *ZodString {
	check := NewZodCheckIncludes(substring, nil, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// LowerCase adds lowercase validation
func (z *ZodString) LowerCase(params ...SchemaParams) *ZodString {
	check := NewZodCheckLowerCase(params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// UpperCase adds uppercase validation
func (z *ZodString) UpperCase(params ...SchemaParams) *ZodString {
	check := NewZodCheckUpperCase(params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Date/Time Validation

// DateTime adds ISO 8601 datetime validation
func (z *ZodString) DateTime(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatDatetime, regexes.DefaultDatetime, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Date adds ISO 8601 date validation (YYYY-MM-DD)
func (z *ZodString) Date(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatDate, regexes.Date, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Time adds ISO 8601 time validation (HH:MM:SS with optional milliseconds)
func (z *ZodString) Time(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatTime, regexes.DefaultTime, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Duration adds ISO 8601 duration validation (P1Y2M3DT4H5M6S format)
func (z *ZodString) Duration(params ...SchemaParams) *ZodString {
	check := NewZodCheckStringFormat(StringFormatDuration, regexes.Duration, params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

// Content Validation

// JSON adds JSON string validation
func (z *ZodString) JSON(params ...SchemaParams) *ZodString {
	check := NewZodCheckJSONString(params...)
	result := AddCheck(z, check)
	return result.(*ZodString)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe string transformation with smart dereferencing
func (z *ZodString) Transform(fn func(string, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		str, isNil, err := extractStringValue(input)

		if err != nil {
			return nil, err
		}

		if isNil {
			return nil, ErrTransformNilString
		}

		return fn(str, ctx)
	})
}

// TransformAny flexible version of transformation
func (z *ZodString) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// String Transformations

// Trim creates a transform that trims whitespace
func (z *ZodString) Trim() ZodType[any, any] {
	return z.createStringTransform(strings.TrimSpace)
}

// ToLowerCase creates a transform to lowercase
func (z *ZodString) ToLowerCase() ZodType[any, any] {
	return z.createStringTransform(strings.ToLower)
}

// ToUpperCase creates a transform to uppercase
func (z *ZodString) ToUpperCase() ZodType[any, any] {
	return z.createStringTransform(strings.ToUpper)
}

// Pipe operation for pipeline chaining
func (z *ZodString) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  z,
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the string optional
func (z *ZodString) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable makes the string nilable while preserving type inference
func (z *ZodString) Nilable() ZodType[any, any] {
	return Clone(z, func(def *ZodTypeDef) {
		// Nilable is a runtime flag
	}).(*ZodString).setNilable()
}

// Nullish makes the string both optional and nilable
func (z *ZodString) Nullish() ZodType[any, any] {
	return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Refine adds type-safe custom validation logic
func (z *ZodString) Refine(fn func(string) bool, params ...SchemaParams) *ZodString {
	result := z.RefineAny(func(v any) bool {
		str, isNil, err := extractStringValue(v)

		if err != nil {
			return false
		}

		if isNil {
			return true // Let Nilable flag handle nil validation
		}

		return fn(str)
	}, params...)
	return result.(*ZodString)
}

// RefineAny adds flexible custom validation logic
func (z *ZodString) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(z, check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodString) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodStringDefault is a default value wrapper for string type
type ZodStringDefault struct {
	*ZodDefault[*ZodString]
}

// DEFAULT METHODS

// Default creates a default wrapper with type safety
func (z *ZodString) Default(value string) ZodStringDefault {
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a default wrapper with function
func (z *ZodString) DefaultFunc(fn func() string) ZodStringDefault {
	genericFn := func() any { return fn() }
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// ZodStringDefault chainable validation methods

func (s ZodStringDefault) Min(minLen int, params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Min(minLen, params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Max(maxLen int, params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Max(maxLen, params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Length(length int, params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Length(length, params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Email(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Email(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) URL(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.URL(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) UUID(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.UUID(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Regex(pattern *regexp.Regexp, params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Regex(pattern, params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) StartsWith(prefix string, params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.StartsWith(prefix, params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) EndsWith(suffix string, params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.EndsWith(suffix, params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Includes(substring string, params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Includes(substring, params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) LowerCase(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.LowerCase(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) UpperCase(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.UpperCase(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) DateTime(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.DateTime(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Date(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Date(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Time(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Time(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Duration(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Duration(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) JSON(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.JSON(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Base64(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Base64(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Emoji(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Emoji(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) CUID(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.CUID(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) CUID2(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.CUID2(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) ULID(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.ULID(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) NanoID(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.NanoID(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) XID(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.XID(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) KSUID(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.KSUID(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Base64URL(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Base64URL(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) E164(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.E164(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) JWT(params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.JWT(params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Refine(fn func(string) bool, params ...SchemaParams) ZodStringDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodStringDefault{
		&ZodDefault[*ZodString]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodStringDefault) Transform(fn func(string, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		str, isNil, err := extractStringValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilString
		}
		return fn(str, ctx)
	})
}

func (s ZodStringDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodStringDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ZodStringPrefault is a prefault value wrapper for string type
type ZodStringPrefault struct {
	*ZodPrefault[*ZodString]
}

// PREFAULT METHODS

// Prefault creates a prefault wrapper with type safety
func (z *ZodString) Prefault(value string) ZodStringPrefault {
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodStringPrefault{
		&ZodPrefault[*ZodString]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a prefault wrapper with function
func (z *ZodString) PrefaultFunc(fn func() string) ZodStringPrefault {
	genericFn := func() any { return fn() }

	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodStringPrefault{
		&ZodPrefault[*ZodString]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// ZodStringPrefault chainable validation methods

func (s ZodStringPrefault) Min(minLen int, params ...SchemaParams) ZodStringPrefault {
	newInner := s.innerType.Min(minLen, params...)

	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodStringPrefault{
		&ZodPrefault[*ZodString]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodStringPrefault) Max(maxLen int, params ...SchemaParams) ZodStringPrefault {
	newInner := s.innerType.Max(maxLen, params...)

	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodStringPrefault{
		&ZodPrefault[*ZodString]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodStringPrefault) Email(params ...SchemaParams) ZodStringPrefault {
	newInner := s.innerType.Email(params...)

	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodStringPrefault{
		&ZodPrefault[*ZodString]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodStringPrefault) Refine(fn func(string) bool, params ...SchemaParams) ZodStringPrefault {
	newInner := s.innerType.Refine(fn, params...)

	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodStringPrefault{
		&ZodPrefault[*ZodString]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodStringPrefault) Transform(fn func(string, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		str, isNil, err := extractStringValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilString
		}
		return fn(str, ctx)
	})
}

func (s ZodStringPrefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodStringPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodStringFromDef creates a ZodString from definition using unified patterns
func createZodStringFromDef(def *ZodStringDef) *ZodString {
	internals := &ZodStringInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           def.Checks,
		Isst:             ZodIssueInvalidType{Expected: "string"},
		Pattern:          nil,
		Values:           make(map[string]struct{}),
		Bag:              make(map[string]interface{}),
	}

	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		stringDef := &ZodStringDef{
			ZodTypeDef: *newDef,
			Type:       ZodTypeString,
			Checks:     newDef.Checks,
		}
		return any(createZodStringFromDef(stringDef)).(ZodType[any, any])
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		result, err := parseType[string](
			payload.Value,
			&internals.ZodTypeInternals,
			"string",
			func(v any) (string, bool) { str, ok := v.(string); return str, ok },
			func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
			validateString,
			coerceToString,
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

	schema := &ZodString{internals: internals}
	initZodType(schema, &def.ZodTypeDef)
	return schema
}

// NewZodString creates a new string schema with unified parameter handling
func NewZodString(params ...SchemaParams) *ZodString {
	def := &ZodStringDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeString,
			Checks: make([]ZodCheck, 0),
		},
		Type:   ZodTypeString,
		Checks: make([]ZodCheck, 0),
	}

	schema := createZodStringFromDef(def)

	if len(params) > 0 {
		param := params[0]

		if param.Coerce {
			schema.internals.Bag["coerce"] = true
			schema.internals.ZodTypeInternals.Bag["coerce"] = true
		}

		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}

		if param.Description != "" {
			schema.internals.Bag["description"] = param.Description
		}
		if param.Abort {
			schema.internals.Bag["abort"] = true
		}
		if len(param.Path) > 0 {
			schema.internals.Bag["path"] = param.Path
		}
	}

	return schema
}

// String creates a new string schema (package-level constructor)
func String(params ...SchemaParams) *ZodString {
	return NewZodString(params...)
}

// CoercedString creates a new string schema with coercion enabled
func CoercedString(params ...SchemaParams) *ZodString {
	var coerceParams SchemaParams
	if len(params) > 0 {
		coerceParams = params[0]
	}
	coerceParams.Coerce = true

	return NewZodString(coerceParams)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// setNilable sets the Nilable flag internally
func (z *ZodString) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// GetZod returns the string-specific internals
func (z *ZodString) GetZod() *ZodStringInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface
func (z *ZodString) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodStringInternals }); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]interface{})
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}

		if len(srcState.ZodTypeInternals.Bag) > 0 {
			if tgtState.ZodTypeInternals.Bag == nil {
				tgtState.ZodTypeInternals.Bag = make(map[string]interface{})
			}
			for key, value := range srcState.ZodTypeInternals.Bag {
				tgtState.ZodTypeInternals.Bag[key] = value
			}
		}

		if len(srcState.Values) > 0 {
			if tgtState.Values == nil {
				tgtState.Values = make(map[string]struct{})
			}
			for key, value := range srcState.Values {
				tgtState.Values[key] = value
			}
		}

		if srcState.Pattern != nil {
			tgtState.Pattern = srcState.Pattern
		}
	}
}

// createStringTransform creates string transformation helper
func (z *ZodString) createStringTransform(transformFn func(string) string) ZodType[any, any] {
	transform := NewZodTransform[any, any](func(input any, _ctx *RefinementContext) (any, error) {
		str, isNil, err := extractStringValue(input)

		if err != nil {
			return nil, err
		}

		if isNil {
			return nil, ErrTransformNilString
		}

		return transformFn(str), nil
	})
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// extractStringValue extracts string value from input with smart handling
func extractStringValue(input any) (string, bool, error) {
	switch v := input.(type) {
	case string:
		return v, false, nil
	case *string:
		if v == nil {
			return "", true, nil
		}
		return *v, false, nil
	default:
		return "", false, fmt.Errorf("%w, got %T", ErrExpectedString, input)
	}
}
