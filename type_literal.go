package gozod

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodLiteralDef defines the configuration for literal value validation
type ZodLiteralDef struct {
	ZodTypeDef
	Type   string        // "literal"
	Values []interface{} // Allowed literal values
}

// ZodLiteralInternals represents the internal state for literal validation
type ZodLiteralInternals struct {
	ZodTypeInternals
	Def     *ZodLiteralDef           // Literal definition
	Values  map[interface{}]struct{} // Set of allowed values
	Pattern *regexp.Regexp           // Regex pattern for validation
	Isst    ZodIssueInvalidType      // Issue type for validation errors
	Bag     map[string]interface{}   // Runtime configuration bag
}

// ZodLiteral represents a literal value schema that validates against specific values
type ZodLiteral struct {
	internals *ZodLiteralInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodLiteral) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for literal type conversion
func (z *ZodLiteral) Coerce(input interface{}) (interface{}, bool) {
	return coerceToLiteralValue(input, z.internals.Def.Values)
}

// Parse validates input with smart type inference
func (z *ZodLiteral) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Use unified nil handling logic
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "literal", "null")
			finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return inferTypedNilFromLiterals(z.internals.Def.Values), nil
	}

	// Smart type inference: input type determines output type
	// Check direct value matching
	if isValidLiteral(input, z.internals) {
		// Run validation checks - use unified runChecksOnValue
		if len(z.internals.Checks) > 0 {
			payload := &ParsePayload{
				Value:  input,
				Issues: make([]ZodRawIssue, 0),
			}
			runChecksOnValue(input, z.internals.Checks, payload, parseCtx)
			if len(payload.Issues) > 0 {
				return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
			}
		}
		return input, nil // Preserve original type
	}

	// Handle pointer types
	if refVal := extractLiteralPointerValue(input); refVal != nil {
		derefValue := refVal.value
		if refVal.isNil {
			if !z.internals.Nilable {
				rawIssue := CreateInvalidTypeIssue(input, "literal", "null")
				finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
				return nil, NewZodError([]ZodIssue{finalIssue})
			}
			return inferTypedNilFromLiterals(z.internals.Def.Values), nil
		}

		// Check if dereferenced value is in literal values
		if isValidLiteral(derefValue, z.internals) {
			// Run validation checks
			if len(z.internals.Checks) > 0 {
				payload := &ParsePayload{
					Value:  derefValue,
					Issues: make([]ZodRawIssue, 0),
				}
				runChecksOnValue(derefValue, z.internals.Checks, payload, parseCtx)
				if len(payload.Issues) > 0 {
					return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
				}
			}
			return input, nil // Return original pointer (preserve type)
		}
	}

	// Try type coercion (if enabled) - use unified shouldCoerce
	if shouldCoerce(z.internals.Bag) {
		for _, value := range z.internals.Def.Values {
			if coerced, ok := CoerceToSpecificValue(input, value); ok {
				// Run validation checks
				if len(z.internals.Checks) > 0 {
					payload := &ParsePayload{
						Value:  coerced,
						Issues: make([]ZodRawIssue, 0),
					}
					runChecksOnValue(coerced, z.internals.Checks, payload, parseCtx)
					if len(payload.Issues) > 0 {
						return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
					}
				}
				return coerced, nil
			}
		}
	}

	// Use unified error creation
	rawIssue := CreateInvalidLiteralIssue(input, z.internals.Def.Values)
	finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
	return nil, NewZodError([]ZodIssue{finalIssue})
}

// MustParse validates the input value and panics on failure
func (z *ZodLiteral) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe literal transformation with smart dereferencing support
func (z *ZodLiteral) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny flexible version of transformation - provides backward compatibility
func (z *ZodLiteral) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// Create pure Transform
	transform := NewZodTransform[any, any](fn)

	// Return pipe(inst, transform)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),         // Type conversion
		out: any(transform).(ZodType[any, any]), // Type conversion
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe chains this literal schema with another schema
func (z *ZodLiteral) Pipe(next ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: next,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the literal schema optional
func (z *ZodLiteral) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable makes the literal schema nilable
func (z *ZodLiteral) Nilable() ZodType[any, any] {
	// Use Clone method to create new instance, avoiding manual state copying
	return Clone(z, func(def *ZodTypeDef) {
		// No need to modify def, as Nilable is a runtime flag
	}).(*ZodLiteral).setNilable()
}

// setNilable internal method to set Nilable flag
func (z *ZodLiteral) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Nullish makes the literal schema optional and nilable
func (z *ZodLiteral) Nullish() ZodType[any, any] {
	return any(Optional(z.Nilable())).(ZodType[any, any])
}

// Refine adds type-safe custom validation logic to the literal schema
func (z *ZodLiteral) Refine(fn func(any) bool, params ...SchemaParams) *ZodLiteral {
	// Use existing RefineAny infrastructure with direct inline handling
	result := z.RefineAny(func(v any) bool {
		val, isNil, err := extractLiteralValue(v)
		if err != nil {
			return false
		}
		if isNil {
			// Handle nil: return true to let upper logic (Nilable flag) decide whether to allow
			return true
		}
		return fn(val)
	}, params...)
	// Return concrete ZodLiteral type to support method chaining
	return result.(*ZodLiteral)
}

// RefineAny adds flexible custom validation logic to the literal schema
func (z *ZodLiteral) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	// Use NewCustom from checks.go to create refine check, following unified pattern
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// Check adds modern validation using direct payload access
func (z *ZodLiteral) Check(fn CheckFn) ZodType[any, any] {
	check := NewCustom[interface{}](func(v any) bool {
		payload := &ParsePayload{
			Value:  v,
			Issues: make([]ZodRawIssue, 0),
			Path:   make([]interface{}, 0),
		}
		fn(payload)
		return len(payload.Issues) == 0
	})
	return AddCheck(z, check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodLiteral) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodLiteralDefault is a Default wrapper for literal type
// Provides perfect type safety and chainable method support
type ZodLiteralDefault struct {
	*ZodDefault[*ZodLiteral] // Embed concrete pointer to enable method promotion
}

// Parse ensures correct validation call to inner type
func (s ZodLiteralDefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// Default adds a default value to the literal schema, returns ZodLiteralDefault support chain call
// Compile-time type safety: Literal("hello").Default(123) will fail to compile
func (z *ZodLiteral) Default(value any) ZodLiteralDefault {
	return ZodLiteralDefault{
		&ZodDefault[*ZodLiteral]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the literal schema, returns ZodLiteralDefault support chain call
func (z *ZodLiteral) DefaultFunc(fn func() any) ZodLiteralDefault {
	return ZodLiteralDefault{
		&ZodDefault[*ZodLiteral]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

// Refine adds a flexible validation function to the literal schema, returns ZodLiteralDefault support chain call
func (s ZodLiteralDefault) Refine(fn func(any) bool, params ...SchemaParams) ZodLiteralDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodLiteralDefault{
		&ZodDefault[*ZodLiteral]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodLiteralDefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of literal values
		literalVal, isNil, err := extractLiteralValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilLiteral
		}
		return fn(literalVal, ctx)
	})
}

// Optional adds an optional check to the literal schema, returns ZodType support chain call
func (s ZodLiteralDefault) Optional() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the literal schema, returns ZodType support chain call
func (s ZodLiteralDefault) Nilable() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(s).(ZodType[any, any]))
}

// ZodLiteralPrefault is a Prefault wrapper for literal type
// Provides perfect type safety and chainable method support
type ZodLiteralPrefault struct {
	*ZodPrefault[*ZodLiteral] // Embed generic wrapper
}

// Prefault adds a prefault value to the literal schema, returns ZodLiteralPrefault support chain call
// Compile-time type safety: Literal("hello").Prefault(123) will fail to compile
func (z *ZodLiteral) Prefault(value interface{}) ZodLiteralPrefault {
	return ZodLiteralPrefault{
		&ZodPrefault[*ZodLiteral]{
			innerType:     z,
			prefaultValue: value,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the literal schema, returns ZodLiteralPrefault support chain call
func (z *ZodLiteral) PrefaultFunc(fn func() interface{}) ZodLiteralPrefault {
	genericFn := func() any { return fn() }
	return ZodLiteralPrefault{
		&ZodPrefault[*ZodLiteral]{
			innerType:    z,
			prefaultFunc: genericFn,
			isFunction:   true,
		},
	}
}

// Refine adds a flexible validation function to the literal schema, returns ZodLiteralPrefault support chain call
func (l ZodLiteralPrefault) Refine(fn func(interface{}) bool, params ...SchemaParams) ZodLiteralPrefault {
	newInner := l.innerType.Refine(fn, params...)
	return ZodLiteralPrefault{
		&ZodPrefault[*ZodLiteral]{
			innerType:     newInner,
			prefaultValue: l.prefaultValue,
			prefaultFunc:  l.prefaultFunc,
			isFunction:    l.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (l ZodLiteralPrefault) Transform(fn func(interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodPrefault's TransformAny method
	return l.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of literal values
		literalVal, isNil, err := extractLiteralValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilLiteral
		}
		return fn(literalVal, ctx)
	})
}

// Optional adds an optional check to the literal schema, returns ZodType support chain call
func (l ZodLiteralPrefault) Optional() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(l).(ZodType[any, any]))
}

// Nilable adds a nilable check to the literal schema, returns ZodType support chain call
func (l ZodLiteralPrefault) Nilable() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(l).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodLiteralFromDef creates a ZodLiteral from a definition
func createZodLiteralFromDef(def *ZodLiteralDef) *ZodLiteral {
	internals := &ZodLiteralInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Values:           make(map[interface{}]struct{}),
		Pattern:          nil,
		Isst:             ZodIssueInvalidType{Expected: "literal"},
		Bag:              make(map[string]interface{}),
	}

	// Build values set using utility pattern - only for hashable types
	for _, value := range def.Values {
		if isHashable(value) {
			internals.Values[value] = struct{}{}
			internals.ZodTypeInternals.Values[value] = struct{}{}
		}
		// For non-hashable types, we'll rely on deep comparison in parseLiteralCore
	}

	// Build regex pattern for string values
	var patterns []string
	for _, value := range def.Values {
		switch v := value.(type) {
		case string:
			// Use escapeRegex from utils.go for consistency
			escaped := escapeRegex(v)
			patterns = append(patterns, escaped)
		case nil:
			patterns = append(patterns, "null")
		default:
			// Convert to string representation and escape
			escaped := escapeRegex(fmt.Sprintf("%v", v))
			patterns = append(patterns, escaped)
		}
	}

	if len(patterns) > 0 {
		patternStr := fmt.Sprintf("^(%s)$", strings.Join(patterns, "|"))
		if compiled, err := regexp.Compile(patternStr); err == nil {
			internals.Pattern = compiled
		}
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		literalDef := &ZodLiteralDef{
			ZodTypeDef: *newDef,
			Type:       ZodTypeLiteral,
			Values:     def.Values, // Preserve values
		}
		return createZodLiteralFromDef(literalDef)
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodLiteral{internals: internals}
		result, err := schema.Parse(payload.Value, ctx)
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

	schema := &ZodLiteral{internals: internals}

	// Initialize the schema using standard pattern
	initZodType(schema, &def.ZodTypeDef)

	return schema
}

// Literal creates a new literal schema that validates against one or more values
func Literal(values ...interface{}) *ZodLiteral {
	// Handle case where last parameter might be SchemaParams
	var params []SchemaParams
	var literalValues []interface{}

	if len(values) == 0 {
		// No values provided, create empty literal (will never match)
		literalValues = []interface{}{}
	} else {
		switch len(values) {
		case 1:
			// Single parameter - could be value, slice of values, or SchemaParams
			if param, ok := values[0].(SchemaParams); ok {
				// Single SchemaParams without values (edge case)
				literalValues = []interface{}{}
				params = []SchemaParams{param}
			} else if slice, ok := values[0].([]interface{}); ok {
				// Single slice - expand to multiple literal values
				literalValues = slice
			} else {
				// Single literal value
				literalValues = values
			}
		case 2:
			// Two parameters - check special case: slice + SchemaParams
			if slice, ok := values[0].([]interface{}); ok {
				if param, ok := values[1].(SchemaParams); ok {
					// First param is slice, second is SchemaParams
					literalValues = slice
					params = []SchemaParams{param}
				} else {
					// First param is slice, second is literal value
					literalValues = make([]interface{}, 0, len(slice)+1)
					literalValues = append(literalValues, slice...)
					literalValues = append(literalValues, values[1])
				}
			} else {
				// Check if last parameter is SchemaParams
				if lastParam, ok := values[len(values)-1].(SchemaParams); ok {
					// Last parameter is SchemaParams
					literalValues = values[:len(values)-1]
					params = []SchemaParams{lastParam}
				} else {
					// All parameters are literal values
					literalValues = values
				}
			}
		default:
			// Multiple parameters (3+) - check if last parameter is SchemaParams
			if lastParam, ok := values[len(values)-1].(SchemaParams); ok {
				// Last parameter is SchemaParams
				literalValues = values[:len(values)-1]
				params = []SchemaParams{lastParam}
			} else {
				// All parameters are literal values
				literalValues = values
			}
		}
	}

	def := &ZodLiteralDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeLiteral,
			Checks: make([]ZodCheck, 0),
		},
		Type:   ZodTypeLiteral,
		Values: literalValues,
	}

	schema := createZodLiteralFromDef(def)

	// Apply schema parameters using standard pattern from type_string.go
	if len(params) > 0 {
		param := params[0]
		// Store coerce flag in bag for parseLiteralCore to access
		if param.Coerce {
			schema.internals.Bag["coerce"] = true
		}

		// Handle schema-level error mapping using utility function
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
	}

	return schema
}

// NewZodLiteral creates a new literal value schema with explicit configuration
func NewZodLiteral(value interface{}, params ...SchemaParams) *ZodLiteral {
	allParams := []interface{}{value}
	if len(params) > 0 {
		allParams = append(allParams, params[0])
	}
	return Literal(allParams...)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodLiteral) CloneFrom(source any) {
	if src, ok := source.(*ZodLiteral); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]interface{})
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy literal-specific fields
		if src.internals.Values != nil {
			z.internals.Values = make(map[interface{}]struct{})
			for k, v := range src.internals.Values {
				z.internals.Values[k] = v
			}
		}

		z.internals.Pattern = src.internals.Pattern
	}
}

// GetZod returns the literal-specific internals
func (z *ZodLiteral) GetZod() *ZodLiteralInternals {
	return z.internals
}

// Clone creates a copy of the literal schema
func (z *ZodLiteral) Clone() ZodType[any, any] {
	return createZodLiteralFromDef(z.internals.Def)
}

// isValidLiteral checks if a value matches any of the allowed literal values
func isValidLiteral(input interface{}, internals *ZodLiteralInternals) bool {
	// For hashable types, use direct map lookup for performance
	if isHashable(input) {
		_, exists := internals.Values[input]
		return exists
	}

	// For non-hashable types (slice, map, function), use deep comparison
	for _, allowedValue := range internals.Def.Values {
		if deepEqual(input, allowedValue) {
			return true
		}
	}
	return false
}

// inferTypedNilFromLiterals returns a typed nil pointer based on literal values
func inferTypedNilFromLiterals(values []interface{}) interface{} {
	if len(values) == 0 {
		return (*interface{})(nil)
	}

	// Infer type from first literal value
	switch values[0].(type) {
	case string:
		return (*string)(nil)
	case int:
		return (*int)(nil)
	case bool:
		return (*bool)(nil)
	case float64:
		return (*float64)(nil)
	default:
		return (*interface{})(nil)
	}
}

// CreateInvalidLiteralIssue creates an invalid literal value issue
func CreateInvalidLiteralIssue(input interface{}, expectedValues []interface{}, options ...func(*ZodRawIssue)) ZodRawIssue {
	message := fmt.Sprintf("Invalid literal value. Expected one of: %s, received: %s",
		JoinValues(expectedValues, " | "),
		StringifyPrimitive(input))

	issue := NewRawIssue(
		string(InvalidValue),
		input,
		WithValues(expectedValues),
		func(issue *ZodRawIssue) {
			issue.Message = message
		},
	)

	// Apply additional options
	for _, opt := range options {
		opt(&issue)
	}

	return issue
}

// LiteralPointerValue represents the result of extracting a pointer value for literals
type LiteralPointerValue struct {
	value any
	isNil bool
}

// extractLiteralPointerValue smart pointer value extraction helper method
// Returns: LiteralPointerValue{value, isNil} or nil (if not a pointer type)
func extractLiteralPointerValue(input any) *LiteralPointerValue {
	switch v := input.(type) {
	case *string:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *int:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *int8:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *int16:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *int32:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *int64:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *uint:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *uint8:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *uint16:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *uint32:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *uint64:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *float32:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *float64:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	case *bool:
		if v == nil {
			return &LiteralPointerValue{value: nil, isNil: true}
		}
		return &LiteralPointerValue{value: *v, isNil: false}
	default:
		return nil // Not a pointer type
	}
}

// CoerceToSpecificValue attempts to coerce input to match a specific literal value
func CoerceToSpecificValue(input any, targetValue any) (any, bool) {
	// Create a single-value set to use with CoerceToValueSet
	valueSet := map[interface{}]struct{}{
		targetValue: {},
	}
	return coerceToValueSet(input, valueSet)
}

// extractLiteralValue smart literal value extraction helper method
// Returns: (literal value, is nil pointer, error)
func extractLiteralValue(input any) (any, bool, error) {
	switch v := input.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return v, false, nil
	default:
		if refVal := extractLiteralPointerValue(input); refVal != nil {
			if refVal.isNil {
				return nil, true, nil
			}
			return refVal.value, false, nil
		}
		return nil, false, fmt.Errorf("%w, got %T", ErrExpectedLiteral, input)
	}
}

// isHashable checks if a value can be used as a map key
func isHashable(value interface{}) bool {
	if value == nil {
		return true
	}

	switch value.(type) {
	case []interface{}, []string, []int, []bool, []float64:
		return false // slices are not hashable
	case func():
		return false // functions are not hashable
	}

	// Use reflection to check for any map type or slice type
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Func:
		return false // maps, slices, and functions are not hashable
	case reflect.Invalid:
		return false // invalid values are not hashable
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.Chan, reflect.Interface, reflect.Pointer, reflect.String,
		reflect.Struct, reflect.UnsafePointer:
		return true // these types are hashable
	default:
		return true // most other types are hashable
	}
}
