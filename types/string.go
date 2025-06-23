package types

import (
	"errors"
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/jsonx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for string transformations
var (
	ErrExpectedString     = errors.New("expected string type")
	ErrTransformNilString = errors.New("cannot transform nil string")
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodStringDef defines the configuration for string validation
type ZodStringDef struct {
	core.ZodTypeDef
	Type   core.ZodTypeCode // Type identifier using type-safe constants
	Checks []core.ZodCheck  // String-specific validation checks
}

// ZodStringInternals contains string validator internal state
type ZodStringInternals struct {
	core.ZodTypeInternals
	Def     *ZodStringDef              // Schema definition
	Checks  []core.ZodCheck            // Validation checks
	Isst    issues.ZodIssueInvalidType // Invalid type issue template
	Pattern *regexp.Regexp             // Regex pattern (if any)
	Values  map[string]struct{}        // Allowed string values set
	Bag     map[string]any             // Additional metadata (formats, etc.)
}

// ZodString represents a string validation schema with type safety
type ZodString struct {
	internals *ZodStringInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodString) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for string type conversion
func (z *ZodString) Coerce(input any) (any, bool) {
	result, err := coerce.ToString(input)
	return result, err == nil
}

// Parse implements intelligent type inference and validation
func (z *ZodString) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return engine.ParsePrimitive[string](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeString,
		validateString,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodString) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodString) Min(minLen int, params ...any) *ZodString {
	check := checks.MinLength(minLen, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Max adds maximum length validation
func (z *ZodString) Max(maxLen int, params ...any) *ZodString {
	check := checks.MaxLength(maxLen, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Length adds exact length validation
func (z *ZodString) Length(length int, params ...any) *ZodString {
	check := checks.Length(length, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Format Validation

// Email adds email format validation
func (z *ZodString) Email(params ...any) *ZodString {
	check := checks.Email(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// URL adds URL format validation
func (z *ZodString) URL(params ...any) *ZodString {
	check := checks.URL(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// UUID adds UUID format validation
func (z *ZodString) UUID(params ...any) *ZodString {
	check := checks.UUID(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Base64 adds base64 format validation
func (z *ZodString) Base64(params ...any) *ZodString {
	check := checks.Base64(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// GUID adds GUID format validation
func (z *ZodString) GUID(params ...any) *ZodString {
	check := checks.GUID(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// CUID adds CUID format validation
func (z *ZodString) CUID(params ...any) *ZodString {
	check := checks.CUID(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// CUID2 adds CUID2 format validation
func (z *ZodString) CUID2(params ...any) *ZodString {
	check := checks.CUID2(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// ULID adds ULID format validation
func (z *ZodString) ULID(params ...any) *ZodString {
	check := checks.ULID(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// NanoID adds NanoID format validation
func (z *ZodString) NanoID(params ...any) *ZodString {
	check := checks.NanoID(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// XID adds XID format validation
func (z *ZodString) XID(params ...any) *ZodString {
	check := checks.XID(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// KSUID adds KSUID format validation
func (z *ZodString) KSUID(params ...any) *ZodString {
	check := checks.KSUID(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Base64URL adds base64url format validation
func (z *ZodString) Base64URL(params ...any) *ZodString {
	check := checks.Base64URL(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// E164 adds E164 phone number format validation
func (z *ZodString) E164(params ...any) *ZodString {
	check := checks.E164(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// JWT adds JWT format validation
func (z *ZodString) JWT(params ...any) *ZodString {
	check := checks.JWT(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Pattern Validation

// Regex adds custom regex validation
func (z *ZodString) Regex(pattern *regexp.Regexp, params ...any) *ZodString {
	check := checks.Regex(pattern, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// StartsWith adds prefix validation
func (z *ZodString) StartsWith(prefix string, params ...any) *ZodString {
	check := checks.StartsWith(prefix, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// EndsWith adds suffix validation
func (z *ZodString) EndsWith(suffix string, params ...any) *ZodString {
	check := checks.EndsWith(suffix, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Includes adds substring validation
func (z *ZodString) Includes(substring string, params ...any) *ZodString {
	check := checks.Includes(substring, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// LowerCase adds lowercase validation
func (z *ZodString) LowerCase(params ...any) *ZodString {
	check := checks.Lowercase(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// UpperCase adds uppercase validation
func (z *ZodString) UpperCase(params ...any) *ZodString {
	check := checks.Uppercase(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Date/Time Validation

// DateTime adds ISO 8601 datetime validation
func (z *ZodString) DateTime(params ...any) *ZodString {
	check := checks.ISODateTime(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Date adds ISO 8601 date validation (YYYY-MM-DD)
func (z *ZodString) Date(params ...any) *ZodString {
	check := checks.ISODate(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Time adds ISO 8601 time validation (HH:MM:SS with optional milliseconds)
func (z *ZodString) Time(params ...any) *ZodString {
	check := checks.ISOTime(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Duration adds ISO 8601 duration validation (P1Y2M3DT4H5M6S format)
func (z *ZodString) Duration(params ...any) *ZodString {
	check := checks.ISODuration(params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

// Content Validation

// JSON adds JSON string validation
func (z *ZodString) JSON(params ...any) *ZodString {
	// Create custom error message from params if provided
	var errorMap *core.ZodErrorMap
	if len(params) > 0 {
		if errStr, ok := params[0].(string); ok {
			em := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return errStr
			})
			errorMap = &em
		} else if schemaParams, ok := params[0].(core.SchemaParams); ok && schemaParams.Error != nil {
			if errStr, ok := schemaParams.Error.(string); ok {
				em := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
					return errStr
				})
				errorMap = &em
			}
		}
	}

	def := &core.ZodCheckDef{Check: "json"}
	if errorMap != nil {
		def.Error = errorMap
	}

	check := &core.ZodCheckInternals{
		Def: def,
		Check: func(payload *core.ParsePayload) {
			// Use pkg/jsonx JSON validation directly
			if !jsonx.IsValid(payload.Value) {
				rawIssue := issues.CreateInvalidFormatIssue("json", payload.Value, nil)
				payload.Issues = append(payload.Issues, rawIssue)
			}
		},
		OnAttach: []func(any){
			func(schema any) {
				// Set JSON format for JSON Schema
				if s, ok := schema.(interface{ GetInternals() *core.ZodTypeInternals }); ok {
					internals := s.GetInternals()
					if internals.Bag == nil {
						internals.Bag = make(map[string]any)
					}
					internals.Bag["format"] = "json"
					internals.Bag["type"] = "string"
				}
			},
		},
	}
	result := engine.AddCheck(z, check)
	return result.(*ZodString)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe string transformation with smart dereferencing
func (z *ZodString) Transform(fn func(string, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		str, ok := reflectx.ExtractString(input)
		if !ok {
			// Try pointer dereferencing
			if ptr, ptrOk := reflectx.Deref(input); ptrOk {
				if str, ok = reflectx.ExtractString(ptr); !ok {
					return nil, ErrExpectedString
				}
			} else {
				return nil, ErrExpectedString
			}
		}

		return fn(str, ctx)
	})
}

// TransformAny flexible version of transformation
func (z *ZodString) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// String Transformations

// Trim creates a transform that trims whitespace
func (z *ZodString) Trim() core.ZodType[any, any] {
	return z.createStringTransform(strings.TrimSpace)
}

// ToLowerCase creates a transform to lowercase
func (z *ZodString) ToLowerCase() core.ZodType[any, any] {
	return z.createStringTransform(strings.ToLower)
}

// ToUpperCase creates a transform to uppercase
func (z *ZodString) ToUpperCase() core.ZodType[any, any] {
	return z.createStringTransform(strings.ToUpper)
}

// Pipe operation for pipeline chaining
func (z *ZodString) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  z,
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the string optional
func (z *ZodString) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable makes the string nilable while preserving type inference
func (z *ZodString) Nilable() core.ZodType[any, any] {
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Nullish makes the string both optional and nilable
func (z *ZodString) Nullish() core.ZodType[any, any] {
	return any(Nullish(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Refine adds type-safe custom validation logic
func (z *ZodString) Refine(fn func(string) bool, params ...any) *ZodString {
	result := z.RefineAny(func(v any) bool {
		str, ok := reflectx.ExtractString(v)
		if !ok {
			// Try pointer dereferencing
			if ptr, ptrOk := reflectx.Deref(v); ptrOk {
				str, ok = reflectx.ExtractString(ptr)
			}
		}
		if !ok {
			return false // Let type validation handle wrong types
		}
		return fn(str)
	}, params...)
	return result.(*ZodString)
}

// RefineAny adds flexible custom validation logic
func (z *ZodString) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodString) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
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

func (s ZodStringDefault) Min(minLen int, params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Max(maxLen int, params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Length(length int, params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Email(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) URL(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) UUID(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Regex(pattern *regexp.Regexp, params ...any) ZodStringDefault {
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

func (s ZodStringDefault) StartsWith(prefix string, params ...any) ZodStringDefault {
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

func (s ZodStringDefault) EndsWith(suffix string, params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Includes(substring string, params ...any) ZodStringDefault {
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

func (s ZodStringDefault) LowerCase(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) UpperCase(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) DateTime(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Date(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Time(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Duration(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) JSON(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Base64(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) CUID(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) CUID2(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) ULID(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) NanoID(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) XID(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) KSUID(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Base64URL(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) E164(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) JWT(params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Refine(fn func(string) bool, params ...any) ZodStringDefault {
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

func (s ZodStringDefault) Transform(fn func(string, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		str, ok := reflectx.ExtractString(input)
		if !ok {
			// Try pointer dereferencing
			if ptr, ptrOk := reflectx.Deref(input); ptrOk {
				if str, ok = reflectx.ExtractString(ptr); !ok {
					return nil, ErrExpectedString
				}
			} else {
				return nil, ErrExpectedString
			}
		}
		return fn(str, ctx)
	})
}

func (s ZodStringDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodStringDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

// ZodStringPrefault is a prefault value wrapper for string type
type ZodStringPrefault struct {
	*ZodPrefault[*ZodString]
}

// PREFAULT METHODS

// Prefault creates a prefault wrapper with type safety
func (z *ZodString) Prefault(value string) ZodStringPrefault {
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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

func (s ZodStringPrefault) Min(minLen int, params ...any) ZodStringPrefault {
	newInner := s.innerType.Min(minLen, params...)

	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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

func (s ZodStringPrefault) Max(maxLen int, params ...any) ZodStringPrefault {
	newInner := s.innerType.Max(maxLen, params...)

	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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

func (s ZodStringPrefault) Email(params ...any) ZodStringPrefault {
	newInner := s.innerType.Email(params...)

	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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

func (s ZodStringPrefault) Refine(fn func(string) bool, params ...any) ZodStringPrefault {
	newInner := s.innerType.Refine(fn, params...)

	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
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

func (s ZodStringPrefault) Transform(fn func(string, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		str, ok := reflectx.ExtractString(input)
		if !ok {
			// Try pointer dereferencing
			if ptr, ptrOk := reflectx.Deref(input); ptrOk {
				if str, ok = reflectx.ExtractString(ptr); !ok {
					return nil, ErrExpectedString
				}
			} else {
				return nil, ErrExpectedString
			}
		}
		return fn(str, ctx)
	})
}

func (s ZodStringPrefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodStringPrefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodStringFromDef creates a ZodString from definition using unified patterns
func createZodStringFromDef(def *ZodStringDef) *ZodString {
	internals := &ZodStringInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           def.Checks,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeString},
		Pattern:          nil,
		Values:           make(map[string]struct{}),
		Bag:              make(map[string]any),
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		stringDef := &ZodStringDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeString,
			Checks:     newDef.Checks,
		}
		return any(createZodStringFromDef(stringDef)).(core.ZodType[any, any])
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		result, err := engine.ParsePrimitive[string](
			payload.Value,
			&internals.ZodTypeInternals,
			core.ZodTypeString,
			validateString,
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

	zodSchema := &ZodString{internals: internals}
	engine.InitZodType(zodSchema, &def.ZodTypeDef)
	return zodSchema
}

// String creates a new string schema with unified parameter handling
func String(params ...any) *ZodString {
	def := &ZodStringDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeString,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   core.ZodTypeString,
		Checks: make([]core.ZodCheck, 0),
	}

	schema := createZodStringFromDef(def)

	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return p
			})
			def.Error = &errorMap
			schema.internals.Error = &errorMap
		case core.SchemaParams:
			// Handle core.SchemaParams
			if p.Error != nil {
				// Handle string error messages by converting to ZodErrorMap
				if errStr, ok := p.Error.(string); ok {
					errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
						return errStr
					})
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				} else if errorMap, ok := p.Error.(core.ZodErrorMap); ok {
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				}
			}

			if p.Description != "" {
				schema.internals.Bag["description"] = p.Description
			}
			if p.Abort {
				schema.internals.Bag["abort"] = true
			}
			if len(p.Path) > 0 {
				schema.internals.Bag["path"] = p.Path
			}
		}
	}

	return schema
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

//nolint:unused // retained for API compatibility
func convertToSchemaParams(params ...any) []core.SchemaParams {
	if len(params) == 0 {
		return []core.SchemaParams{}
	}

	result := make([]core.SchemaParams, 0, len(params))
	for _, param := range params {
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			result = append(result, core.SchemaParams{Error: p})
		case core.SchemaParams:
			// Already a SchemaParams
			result = append(result, p)
		default:
			// For other types, try to use as Error
			result = append(result, core.SchemaParams{Error: p})
		}
	}
	return result
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
				tgtState.Bag = make(map[string]any)
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}

		if len(srcState.ZodTypeInternals.Bag) > 0 {
			if tgtState.ZodTypeInternals.Bag == nil {
				tgtState.ZodTypeInternals.Bag = make(map[string]any)
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
func (z *ZodString) createStringTransform(transformFn func(string) string) core.ZodType[any, any] {
	transform := Transform[any, any](func(input any, _ctx *core.RefinementContext) (any, error) {
		str, ok := reflectx.ExtractString(input)
		if !ok {
			// Try pointer dereferencing
			if ptr, ptrOk := reflectx.Deref(input); ptrOk {
				if str, ok = reflectx.ExtractString(ptr); !ok {
					return nil, ErrExpectedString
				}
			} else {
				return nil, ErrExpectedString
			}
		}

		return transformFn(str), nil
	})
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// VALIDATION FUNCTIONS
//////////////////////////

// validateString validates string values with checks
func validateString(value string, checks []core.ZodCheck, ctx *core.ParseContext) error {
	if len(checks) > 0 {
		payload := &core.ParsePayload{
			Value:  value,
			Issues: make([]core.ZodRawIssue, 0),
		}
		engine.RunChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.Issues, ctx))
		}
	}
	return nil
}
