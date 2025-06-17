package gozod

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

//////////////////////////////////////////
//////////////////////////////////////////
//////////                      //////////
//////////   ZodStringBool      //////////
//////////                      //////////
//////////////////////////////////////////
//////////////////////////////////////////

// ZodStringBoolDef defines the configuration for stringbool validation
type ZodStringBoolDef struct {
	ZodTypeDef
	Type   string   // "stringbool"
	Truthy []string // Truthy string values
	Falsy  []string // Falsy string values
	Case   string   // "sensitive" or "insensitive"
}

// ZodStringBoolInternals contains stringbool validator internal state
type ZodStringBoolInternals struct {
	ZodTypeInternals
	Def    *ZodStringBoolDef      // Schema definition
	Isst   ZodIssueInvalidType    // Invalid type issue template
	Truthy map[string]struct{}    // Truthy values set for fast lookup
	Falsy  map[string]struct{}    // Falsy values set for fast lookup
	Bag    map[string]interface{} // Runtime configuration
}

// ZodStringBool represents a string-to-boolean validation schema
type ZodStringBool struct {
	internals *ZodStringBoolInternals
}

// StringBoolOptions provides configuration for stringbool schema creation
type StringBoolOptions struct {
	Truthy []string // Values that evaluate to true
	Falsy  []string // Values that evaluate to false
	Case   string   // "sensitive" or "insensitive"
}

//////////////////////////////////////////
// Validation Methods
//////////////////////////////////////////

// GetInternals returns the internal state for framework use
func (z *ZodStringBool) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns stringbool-specific internal state
func (z *ZodStringBool) GetZod() *ZodStringBoolInternals {
	return z.internals
}

// Coerce attempts to coerce input to string for stringbool validation
func (z *ZodStringBool) Coerce(input interface{}) (interface{}, bool) {
	if str, ok := coerceToString(input); ok {
		return str, true
	}
	return nil, false
}

// CloneFrom implements Cloneable interface, copying stringbool-specific state
func (z *ZodStringBool) CloneFrom(source any) {
	if src, ok := source.(interface {
		GetZod() *ZodStringBoolInternals
	}); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		// Copy truthy/falsy maps
		if len(srcState.Truthy) > 0 {
			if tgtState.Truthy == nil {
				tgtState.Truthy = make(map[string]struct{})
			}
			for k, v := range srcState.Truthy {
				tgtState.Truthy[k] = v
			}
		}

		if len(srcState.Falsy) > 0 {
			if tgtState.Falsy == nil {
				tgtState.Falsy = make(map[string]struct{})
			}
			for k, v := range srcState.Falsy {
				tgtState.Falsy[k] = v
			}
		}

		// Copy bag state
		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]interface{})
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}
	}
}

// Parse validates and converts string input to boolean output using smart type inference
func (z *ZodStringBool) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// 1. Use parseType[string] to handle input validation and type inference
	stringResult, err := parseType[string](
		input,
		&z.internals.ZodTypeInternals,
		"string",
		func(v any) (string, bool) { str, ok := v.(string); return str, ok },
		func(v any) (*string, bool) { ptr, ok := v.(*string); return ptr, ok },
		z.validateStringInput,         // Only validates string, no conversion
		z.coerceToStringForStringBool, // StringBool-specific coercion
		parseCtx,
	)
	if err != nil {
		return nil, err
	}

	// 2. Handle string → bool conversion while preserving type inference
	return z.convertStringResultToBool(stringResult)
}

// MustParse validates the input value and panics on failure
func (z *ZodStringBool) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// validateStringInput validates string input without conversion (for parseType usage)
func (z *ZodStringBool) validateStringInput(value string, checks []ZodCheck, ctx *ParseContext) error {
	// Only basic string validation, no bool conversion
	// StringBool checks should apply to final bool result, not intermediate string
	return nil
}

// coerceToStringForStringBool provides stricter coercion logic for StringBool
func (z *ZodStringBool) coerceToStringForStringBool(value interface{}) (string, bool) {
	if value == nil {
		return "", false
	}

	switch v := value.(type) {
	case string:
		return v, true
	case *string:
		if v == nil {
			return "", false
		}
		return *v, true
	case int:
		// Only allow 0 and 1
		if v == 0 || v == 1 {
			return fmt.Sprintf("%d", v), true
		}
		return "", false
	case int8, int16, int32, int64:
		intVal := reflect.ValueOf(v).Int()
		if intVal == 0 || intVal == 1 {
			return fmt.Sprintf("%d", intVal), true
		}
		return "", false
	case uint, uint8, uint16, uint32, uint64:
		uintVal := reflect.ValueOf(v).Uint()
		if uintVal == 0 || uintVal == 1 {
			return fmt.Sprintf("%d", uintVal), true
		}
		return "", false
	case float32, float64:
		floatVal := reflect.ValueOf(v).Float()
		if floatVal == 0.0 || floatVal == 1.0 {
			return fmt.Sprintf("%.0f", floatVal), true
		}
		return "", false
	default:
		// Other types not allowed (including bool)
		return "", false
	}
}

// convertStringResultToBool converts parseType[string] result to bool, preserving StringBool semantics
func (z *ZodStringBool) convertStringResultToBool(stringResult any) (any, error) {
	switch v := stringResult.(type) {
	case string:
		// string → bool conversion
		return z.stringToBool(v, z.internals.Checks, nil)
	case *string:
		if v == nil {
			// *string(nil) → *bool(nil) (only in Nilable context)
			return (*bool)(nil), nil
		}
		// StringBool semantics: regardless of input being string or *string, convert to bool
		return z.stringToBool(*v, z.internals.Checks, nil)
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnexpectedStringResultType, stringResult)
	}
}

// stringToBool performs the core string-to-bool conversion and validation
func (z *ZodStringBool) stringToBool(value string, checks []ZodCheck, ctx *ParseContext) (bool, error) {
	// Normalize case if insensitive
	normalizedValue := value
	if z.internals.Def.Case == "insensitive" {
		normalizedValue = strings.ToLower(value)
	}

	// Check truthy values
	if _, exists := z.internals.Truthy[normalizedValue]; exists {
		// Run additional checks on the boolean result
		if len(checks) > 0 {
			payload := &ParsePayload{
				Value:  true,
				Issues: make([]ZodRawIssue, 0),
			}
			runChecksOnValue(true, checks, payload, ctx)
			if len(payload.Issues) > 0 {
				return false, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}
		}
		return true, nil
	}

	// Check falsy values
	if _, exists := z.internals.Falsy[normalizedValue]; exists {
		// Run additional checks on the boolean result
		if len(checks) > 0 {
			payload := &ParsePayload{
				Value:  false,
				Issues: make([]ZodRawIssue, 0),
			}
			runChecksOnValue(false, checks, payload, ctx)
			if len(payload.Issues) > 0 {
				return false, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
			}
		}
		return false, nil
	}

	// Invalid value - create error
	allValues := make([]string, 0, len(z.internals.Def.Truthy)+len(z.internals.Def.Falsy))
	allValues = append(allValues, z.internals.Def.Truthy...)
	allValues = append(allValues, z.internals.Def.Falsy...)
	rawIssue := CreateInvalidValueIssue(value, []interface{}{allValues})

	// Use schema's custom error message if available
	config := GetConfig()
	if z.internals.Error != nil {
		config = &ZodConfig{CustomError: *z.internals.Error}
	}
	finalIssue := FinalizeIssue(rawIssue, ctx, config)
	return false, NewZodError([]ZodIssue{finalIssue})
}

//////////////////////////////////////////
// Transform Methods
//////////////////////////////////////////

// Transform creates a type-safe transformation of stringbool values
func (z *ZodStringBool) Transform(fn func(bool, *RefinementContext) (any, error)) ZodType[any, any] {
	// Wrapper function to handle type extraction
	wrappedFn := func(input any, ctx *RefinementContext) (any, error) {
		boolVal, isNil, err := extractBoolValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return (*bool)(nil), nil
		}
		return fn(boolVal, ctx)
	}
	return z.TransformAny(wrappedFn)
}

// TransformAny adds data transformation with any input type
func (z *ZodStringBool) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe connects this schema to another schema
func (z *ZodStringBool) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////////////////////
// Modifier Methods
//////////////////////////////////////////

// Optional makes the stringbool optional
func (z *ZodStringBool) Optional() ZodType[any, any] {
	return Optional(z)
}

// Nilable creates a nilable stringbool schema
func (z *ZodStringBool) Nilable() ZodType[any, any] {
	return Nilable(z)
}

// Nullish creates a nullish (optional + nilable) stringbool schema
func (z *ZodStringBool) Nullish() ZodType[any, any] {
	return Nullish(z)
}

// Refine adds type-safe custom validation logic to the stringbool schema
func (z *ZodStringBool) Refine(fn func(bool) bool, params ...SchemaParams) *ZodStringBool {
	result := z.RefineAny(func(v any) bool {
		boolVal, isNil, err := extractBoolValue(v)
		if err != nil {
			return false
		}
		if isNil {
			// nil *bool handling: return true to let upper logic (Nilable flag) decide
			return true
		}
		return fn(boolVal)
	}, params...)
	return result.(*ZodStringBool)
}

// RefineAny adds flexible custom validation logic to the stringbool schema
func (z *ZodStringBool) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(z, check)
}

//////////////////////////////////////////
// Wrapper Types
//////////////////////////////////////////

// ZodStringBoolDefault is the Default wrapper for stringbool type
type ZodStringBoolDefault struct {
	*ZodDefault[*ZodStringBool] // Embed concrete pointer for method promotion
}

// Default adds a default value to the stringbool schema, returns ZodStringBoolDefault
func (z *ZodStringBool) Default(value bool) ZodStringBoolDefault {
	return ZodStringBoolDefault{
		&ZodDefault[*ZodStringBool]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the stringbool schema, returns ZodStringBoolDefault
func (z *ZodStringBool) DefaultFunc(fn func() bool) ZodStringBoolDefault {
	genericFn := func() any { return fn() }
	return ZodStringBoolDefault{
		&ZodDefault[*ZodStringBool]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Refine adds a flexible validation function to the stringbool schema, returns ZodStringBoolDefault
func (s ZodStringBoolDefault) Refine(fn func(bool) bool, params ...SchemaParams) ZodStringBoolDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodStringBoolDefault{
		&ZodDefault[*ZodStringBool]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType for transformation pipeline
func (s ZodStringBoolDefault) Transform(fn func(bool, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		boolVal, isNil, err := extractBoolValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilBool
		}
		return fn(boolVal, ctx)
	})
}

// Parse implements ZodType interface, special handling for StringBool boolean default values
func (s ZodStringBoolDefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	// StringBool special logic: when nil, directly return boolean default without string validation
	if input == nil {
		var defaultValue any
		if s.isFunction && s.defaultFunc != nil {
			defaultValue = s.defaultFunc()
		} else {
			defaultValue = s.defaultValue
		}

		// Directly return boolean default value without StringBool string validation
		return defaultValue, nil
	}

	// Non-nil input delegates to inner type's Parse (preserves smart type inference)
	return s.innerType.Parse(input, ctx...)
}

// Prefault method, supports method chaining
func (s ZodStringBoolDefault) Prefault(value bool) ZodType[any, any] {
	// Create special Prefault wrapper, preserving StringBool default value handling logic
	return &ZodPrefault[ZodType[any, any]]{
		innerType:     any(s).(ZodType[any, any]),
		prefaultValue: value,
		isFunction:    false,
	}
}

// PrefaultFunc method, supports method chaining
func (s ZodStringBoolDefault) PrefaultFunc(fn func() bool) ZodType[any, any] {
	genericFn := func() any { return fn() }
	// Create special Prefault wrapper, preserving StringBool default value handling logic
	return &ZodPrefault[ZodType[any, any]]{
		innerType:    any(s).(ZodType[any, any]),
		prefaultFunc: genericFn,
		isFunction:   true,
	}
}

// Optional wraps the current ZodStringBoolDefault instance, preserving Default logic
func (s ZodStringBoolDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable wraps the current ZodStringBoolDefault instance, preserving Default logic
func (s ZodStringBoolDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ZodStringBoolPrefault is the Prefault wrapper for stringbool type
type ZodStringBoolPrefault struct {
	*ZodPrefault[*ZodStringBool] // Embed concrete pointer for method promotion
}

// Prefault adds a prefault value to the stringbool schema, returns ZodStringBoolPrefault
func (z *ZodStringBool) Prefault(value bool) ZodStringBoolPrefault {
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

	return ZodStringBoolPrefault{
		&ZodPrefault[*ZodStringBool]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the stringbool schema, returns ZodStringBoolPrefault
func (z *ZodStringBool) PrefaultFunc(fn func() bool) ZodStringBoolPrefault {
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

	return ZodStringBoolPrefault{
		&ZodPrefault[*ZodStringBool]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// Refine adds a flexible validation function to the stringbool schema, returns ZodStringBoolPrefault
func (b ZodStringBoolPrefault) Refine(fn func(bool) bool, params ...SchemaParams) ZodStringBoolPrefault {
	newInner := b.innerType.Refine(fn, params...)

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

	return ZodStringBoolPrefault{
		&ZodPrefault[*ZodStringBool]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: b.prefaultValue,
			prefaultFunc:  b.prefaultFunc,
			isFunction:    b.isFunction,
		},
	}
}

// Transform adds data transformation, returns generic ZodType for transformation pipeline
func (b ZodStringBoolPrefault) Transform(fn func(bool, *RefinementContext) (any, error)) ZodType[any, any] {
	return b.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		boolVal, isNil, err := extractBoolValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilBool
		}
		return fn(boolVal, ctx)
	})
}

// Optional wraps the current ZodStringBoolPrefault instance, preserving Prefault logic
func (b ZodStringBoolPrefault) Optional() ZodType[any, any] {
	return Optional(any(b).(ZodType[any, any]))
}

// Nilable wraps the current ZodStringBoolPrefault instance, preserving Prefault logic
func (b ZodStringBoolPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(b).(ZodType[any, any]))
}

//////////////////////////////////////////
// Constructor Functions
//////////////////////////////////////////

// createZodStringBoolFromDef creates a ZodStringBool from definition with unified patterns
func createZodStringBoolFromDef(def *ZodStringBoolDef) *ZodStringBool {
	internals := &ZodStringBoolInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             ZodIssueInvalidType{Expected: "string"},
		Truthy:           make(map[string]struct{}),
		Falsy:            make(map[string]struct{}),
		Bag:              make(map[string]interface{}),
	}

	// Build lookup maps for efficient validation
	for _, value := range def.Truthy {
		normalizedValue := value
		if def.Case == "insensitive" {
			normalizedValue = strings.ToLower(value)
		}
		internals.Truthy[normalizedValue] = struct{}{}
	}

	for _, value := range def.Falsy {
		normalizedValue := value
		if def.Case == "insensitive" {
			normalizedValue = strings.ToLower(value)
		}
		internals.Falsy[normalizedValue] = struct{}{}
	}

	// Set up constructor for AddCheck functionality
	internals.Constructor = func(def *ZodTypeDef) ZodType[any, any] {
		stringBoolDef := &ZodStringBoolDef{
			ZodTypeDef: *def,
			Type:       "stringbool",
			Truthy:     internals.Def.Truthy,
			Falsy:      internals.Def.Falsy,
			Case:       internals.Def.Case,
		}
		return createZodStringBoolFromDef(stringBoolDef)
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		// Create temporary schema for parsing
		tempSchema := &ZodStringBool{internals: internals}

		// Use the new parseType-based logic
		result, err := tempSchema.Parse(payload.Value, ctx)

		if err != nil {
			// Convert error to ParsePayload format
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

	schema := &ZodStringBool{internals: internals}

	// Initialize the schema using unified initZodType from type.go
	initZodType(schema, &def.ZodTypeDef)

	return schema
}

// NewZodStringBool creates a new stringbool schema with unified error handling and parameter processing
func NewZodStringBool(options *StringBoolOptions, params ...SchemaParams) *ZodStringBool {
	// Default values
	truthy := []string{"true", "1", "yes", "on", "y", "enabled"}
	falsy := []string{"false", "0", "no", "off", "n", "disabled"}
	caseMode := "insensitive"

	// Apply custom options
	if options != nil {
		if len(options.Truthy) > 0 {
			truthy = options.Truthy
		}
		if len(options.Falsy) > 0 {
			falsy = options.Falsy
		}
		if options.Case != "" {
			caseMode = options.Case
		}
	}

	def := &ZodStringBoolDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeStringBool,
			Checks: make([]ZodCheck, 0),
		},
		Type:   ZodTypeStringBool,
		Truthy: truthy,
		Falsy:  falsy,
		Case:   caseMode,
	}

	schema := createZodStringBoolFromDef(def)

	// Apply schema parameters with unified error handling pattern
	if len(params) > 0 {
		param := params[0]

		// Store coerce flag in bag for parsing to access
		if param.Coerce {
			schema.internals.Bag["coerce"] = true
			schema.internals.ZodTypeInternals.Bag["coerce"] = true
		}

		// Handle schema-level error configuration using utility function
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}

		// Handle additional parameters following unified pattern
		if param.Description != "" {
			schema.internals.Bag["description"] = param.Description
		}
		if param.Abort {
			schema.internals.Bag["abort"] = true
		}
		if len(param.Path) > 0 {
			schema.internals.Bag["path"] = param.Path
		}
		if len(param.Params) > 0 {
			schema.internals.Bag["params"] = param.Params
		}
	}

	return schema
}

// StringBool creates a stringbool schema with default or custom options
func StringBool(options ...*StringBoolOptions) *ZodStringBool {
	var opts *StringBoolOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return NewZodStringBool(opts)
}

// StringBoolWithError creates a stringbool schema with custom error message
func StringBoolWithError(errorMessage string, options ...*StringBoolOptions) *ZodStringBool {
	var opts *StringBoolOptions
	if len(options) > 0 {
		opts = options[0]
	}
	return NewZodStringBool(opts, SchemaParams{Error: errorMessage})
}

//////////////////////////////////////////
// Utility Functions
//////////////////////////////////////////

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodStringBool) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}
