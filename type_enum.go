package gozod

import (
	"fmt"
	"regexp"
	"strings"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodEnumDef defines the configuration for enum validation
type ZodEnumDef[T comparable] struct {
	ZodTypeDef
	Type    string       // "enum"
	Entries map[string]T // Enum entries (key-value pairs)
}

// ZodEnumInternals contains enum validator internal state
type ZodEnumInternals[T comparable] struct {
	ZodTypeInternals
	Def     *ZodEnumDef[T]         // Schema definition
	Entries map[string]T           // Enum entries
	Values  map[T]struct{}         // Set of valid values
	Pattern *regexp.Regexp         // Regex pattern for validation
	Isst    ZodIssueInvalidValue   // Invalid value issue template
	Bag     map[string]interface{} // Additional metadata
}

// ZodEnum represents a type-safe enum validation schema
type ZodEnum[T comparable] struct {
	internals *ZodEnumInternals[T]
}

// GetInternals returns the internal state of the schema
func (z *ZodEnum[T]) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for enum value conversion
func (z *ZodEnum[T]) Coerce(input interface{}) (interface{}, bool) {
	return coerceToValueSet(input, convertValueSetToInterface(z.internals.Values))
}

// Parse validates and parses input with smart type inference
func (z *ZodEnum[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Handle nil input
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "enum", "null")
			finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return nil, nil
	}

	// Type assertion check
	if value, ok := input.(T); ok {
		if _, exists := z.internals.Values[value]; exists {
			if len(z.internals.Checks) > 0 {
				payload := &ParsePayload{
					Value:  value,
					Issues: make([]ZodRawIssue, 0),
				}
				runChecksOnValue(value, z.internals.Checks, payload, parseCtx)
				if len(payload.Issues) > 0 {
					return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
				}
			}
			return value, nil
		}
	}

	// Handle pointer types
	if refVal := extractPointerValue(input); refVal != nil {
		if refVal.isNil {
			if !z.internals.Nilable {
				rawIssue := CreateInvalidTypeIssue(input, "enum", "null")
				finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
				return nil, NewZodError([]ZodIssue{finalIssue})
			}
			return nil, nil
		}

		if value, ok := refVal.value.(T); ok {
			if _, exists := z.internals.Values[value]; exists {
				if len(z.internals.Checks) > 0 {
					payload := &ParsePayload{
						Value:  value,
						Issues: make([]ZodRawIssue, 0),
					}
					runChecksOnValue(value, z.internals.Checks, payload, parseCtx)
					if len(payload.Issues) > 0 {
						return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
					}
				}
				return input, nil // Preserve pointer type
			}
		}
	}

	// Try type coercion (if enabled)
	if shouldCoerce(z.internals.Bag) {
		if coerced, ok := coerceToValueSet(input, convertValueSetToInterface(z.internals.Values)); ok {
			if value, ok := coerced.(T); ok {
				if len(z.internals.Checks) > 0 {
					payload := &ParsePayload{
						Value:  value,
						Issues: make([]ZodRawIssue, 0),
					}
					runChecksOnValue(value, z.internals.Checks, payload, parseCtx)
					if len(payload.Issues) > 0 {
						return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
					}
				}
				return value, nil
			}
		}
	}

	// Create error
	values := make([]interface{}, 0, len(z.internals.Values))
	for value := range z.internals.Values {
		values = append(values, value)
	}

	rawIssue := CreateInvalidLiteralIssue(input, values)
	finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
	return nil, NewZodError([]ZodIssue{finalIssue})
}

// MustParse validates the input value and panics on failure
func (z *ZodEnum[T]) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// ENUM SPECIFIC METHODS
//////////////////////////

// Enum returns the enum mapping
func (z *ZodEnum[T]) Enum() map[string]T {
	result := make(map[string]T)
	for k, v := range z.internals.Entries {
		result[k] = v
	}
	return result
}

// Options returns all possible values
func (z *ZodEnum[T]) Options() []T {
	values := make([]T, 0, len(z.internals.Values))
	for value := range z.internals.Values {
		values = append(values, value)
	}
	return values
}

// Extract creates a sub-enum with specified keys
func (z *ZodEnum[T]) Extract(keys []string, params ...SchemaParams) *ZodEnum[T] {
	newEntries := make(map[string]T)
	for _, key := range keys {
		if value, exists := z.internals.Entries[key]; exists {
			newEntries[key] = value
		} else {
			panic(fmt.Sprintf("Key '%s' not found in enum", key))
		}
	}

	mergedParams := mergeEnumSchemaParams(z.internals.Bag, params...)
	return NewZodEnum(newEntries, mergedParams)
}

// Exclude creates a sub-enum excluding specified keys
func (z *ZodEnum[T]) Exclude(keys []string, params ...SchemaParams) *ZodEnum[T] {
	excludeSet := make(map[string]bool)
	for _, key := range keys {
		if _, exists := z.internals.Entries[key]; !exists {
			panic(fmt.Sprintf("Key '%s' not found in enum", key))
		}
		excludeSet[key] = true
	}

	newEntries := make(map[string]T)
	for key, value := range z.internals.Entries {
		if !excludeSet[key] {
			newEntries[key] = value
		}
	}

	mergedParams := mergeEnumSchemaParams(z.internals.Bag, params...)
	return NewZodEnum(newEntries, mergedParams)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe enum transformation
func (z *ZodEnum[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](func(input any, ctx *RefinementContext) (any, error) {
		if value, ok := input.(T); ok {
			return fn(value, ctx)
		}
		return nil, ErrInvalidTypeForTransform
	})

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: transform,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// TransformAny flexible version of transformation
func (z *ZodEnum[T]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: transform,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodEnum[T]) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the enum optional
func (z *ZodEnum[T]) Optional() ZodType[any, any] {
	return NewZodOptional(any(z).(ZodType[any, any]))
}

// Nilable makes the enum nilable
func (z *ZodEnum[T]) Nilable() ZodType[any, any] {
	clone := z.clone()
	clone.internals.Nilable = true
	return any(clone).(ZodType[any, any])
}

// Nullish makes the enum both optional and nilable
func (z *ZodEnum[T]) Nullish() ZodType[any, any] {
	return z.Optional().Nilable()
}

// Refine adds type-safe custom validation
func (z *ZodEnum[T]) Refine(fn func(T) bool, params ...SchemaParams) *ZodEnum[T] {
	refined := z.RefineAny(func(val any) bool {
		if v, ok := val.(T); ok {
			return fn(v)
		}
		return false
	}, params...)

	if enumType, ok := refined.(*ZodEnum[T]); ok {
		return enumType
	}

	// If type assertion fails, create a new enum
	clone := z.clone()
	check := createEnumRefinementCheck(func(val any) bool {
		if v, ok := val.(T); ok {
			return fn(v)
		}
		return false
	}, params...)
	clone.internals.Checks = append(clone.internals.Checks, check)
	return clone
}

// RefineAny adds flexible custom validation
func (z *ZodEnum[T]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	clone := z.clone()
	check := createEnumRefinementCheck(fn, params...)
	clone.internals.Checks = append(clone.internals.Checks, check)
	return any(clone).(ZodType[any, any])
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodEnum[T]) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodEnumDefault is a default value wrapper for enum type
type ZodEnumDefault[T comparable] struct {
	inner        *ZodEnum[T]
	defaultValue T
	defaultFunc  func() T
	isFunction   bool
}

// GetInternals returns the internal state
func (s ZodEnumDefault[T]) GetInternals() *ZodTypeInternals {
	return s.inner.GetInternals()
}

// Parse method with default value support
func (s ZodEnumDefault[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	if input == nil {
		if s.isFunction && s.defaultFunc != nil {
			return s.defaultFunc(), nil
		}
		return s.defaultValue, nil
	}
	return s.inner.Parse(input, ctx...)
}

// MustParse method
func (s ZodEnumDefault[T]) MustParse(input any, ctx ...*ParseContext) any {
	result, err := s.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// TransformAny method
func (s ZodEnumDefault[T]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.inner.TransformAny(fn)
}

// Pipe method
func (s ZodEnumDefault[T]) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return s.inner.Pipe(out)
}

// Optional method
func (s ZodEnumDefault[T]) Optional() ZodType[any, any] {
	return s.inner.Optional()
}

// Nilable method
func (s ZodEnumDefault[T]) Nilable() ZodType[any, any] {
	return s.inner.Nilable()
}

// Nullish method
func (s ZodEnumDefault[T]) Nullish() ZodType[any, any] {
	return s.inner.Nullish()
}

// RefineAny method
func (s ZodEnumDefault[T]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	return s.inner.RefineAny(fn, params...)
}

// Unwrap method
func (s ZodEnumDefault[T]) Unwrap() ZodType[any, any] {
	return s.inner.Unwrap()
}

// Default creates an enum with default value
func (z *ZodEnum[T]) Default(value T) ZodEnumDefault[T] {
	return ZodEnumDefault[T]{
		inner:        z,
		defaultValue: value,
		isFunction:   false,
	}
}

// DefaultFunc creates an enum with default function
func (z *ZodEnum[T]) DefaultFunc(fn func() T) ZodEnumDefault[T] {
	return ZodEnumDefault[T]{
		inner:       z,
		defaultFunc: fn,
		isFunction:  true,
	}
}

// Refine adds validation to default enum
func (s ZodEnumDefault[T]) Refine(fn func(T) bool, params ...SchemaParams) ZodEnumDefault[T] {
	refined := s.inner.Refine(fn, params...)
	return ZodEnumDefault[T]{
		inner:        refined,
		defaultValue: s.defaultValue,
		defaultFunc:  s.defaultFunc,
		isFunction:   s.isFunction,
	}
}

// Transform transforms default enum
func (s ZodEnumDefault[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.inner.Transform(fn)
}

// Enum returns enum mapping
func (s ZodEnumDefault[T]) Enum() map[string]T {
	return s.inner.Enum()
}

// Options returns all possible values
func (s ZodEnumDefault[T]) Options() []T {
	return s.inner.Options()
}

// Extract creates sub-enum with specified keys
func (s ZodEnumDefault[T]) Extract(keys []string, params ...SchemaParams) ZodEnumDefault[T] {
	extracted := s.inner.Extract(keys, params...)
	return ZodEnumDefault[T]{
		inner:        extracted,
		defaultValue: s.defaultValue,
		defaultFunc:  s.defaultFunc,
		isFunction:   s.isFunction,
	}
}

// Exclude creates sub-enum excluding specified keys
func (s ZodEnumDefault[T]) Exclude(keys []string, params ...SchemaParams) ZodEnumDefault[T] {
	excluded := s.inner.Exclude(keys, params...)
	return ZodEnumDefault[T]{
		inner:        excluded,
		defaultValue: s.defaultValue,
		defaultFunc:  s.defaultFunc,
		isFunction:   s.isFunction,
	}
}

// ZodEnumPrefault is a prefault value wrapper for enum type
type ZodEnumPrefault[T comparable] struct {
	inner         *ZodEnum[T]
	prefaultValue T
	prefaultFunc  func() T
	isFunction    bool
}

// GetInternals returns the internal state
func (e ZodEnumPrefault[T]) GetInternals() *ZodTypeInternals {
	return e.inner.GetInternals()
}

// Parse method with prefault value support
func (e ZodEnumPrefault[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	result, err := e.inner.Parse(input, ctx...)
	if err != nil {
		// Use prefault value on validation failure
		if e.isFunction && e.prefaultFunc != nil {
			return e.prefaultFunc(), nil
		}
		return e.prefaultValue, nil
	}
	return result, nil
}

// MustParse method
func (e ZodEnumPrefault[T]) MustParse(input any, ctx ...*ParseContext) any {
	result, err := e.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// TransformAny method
func (e ZodEnumPrefault[T]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return e.inner.TransformAny(fn)
}

// Pipe method
func (e ZodEnumPrefault[T]) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return e.inner.Pipe(out)
}

// Optional method
func (e ZodEnumPrefault[T]) Optional() ZodType[any, any] {
	return e.inner.Optional()
}

// Nilable method
func (e ZodEnumPrefault[T]) Nilable() ZodType[any, any] {
	return e.inner.Nilable()
}

// Nullish method
func (e ZodEnumPrefault[T]) Nullish() ZodType[any, any] {
	return e.inner.Nullish()
}

// RefineAny method
func (e ZodEnumPrefault[T]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	return e.inner.RefineAny(fn, params...)
}

// Unwrap method
func (e ZodEnumPrefault[T]) Unwrap() ZodType[any, any] {
	return e.inner.Unwrap()
}

// Prefault creates an enum with prefault value
func (z *ZodEnum[T]) Prefault(value T) ZodEnumPrefault[T] {
	return ZodEnumPrefault[T]{
		inner:         z,
		prefaultValue: value,
		isFunction:    false,
	}
}

// PrefaultFunc creates an enum with prefault function
func (z *ZodEnum[T]) PrefaultFunc(fn func() T) ZodEnumPrefault[T] {
	return ZodEnumPrefault[T]{
		inner:        z,
		prefaultFunc: fn,
		isFunction:   true,
	}
}

// Refine adds validation to prefault enum
func (e ZodEnumPrefault[T]) Refine(fn func(T) bool, params ...SchemaParams) ZodEnumPrefault[T] {
	refined := e.inner.Refine(fn, params...)
	return ZodEnumPrefault[T]{
		inner:         refined,
		prefaultValue: e.prefaultValue,
		prefaultFunc:  e.prefaultFunc,
		isFunction:    e.isFunction,
	}
}

// Transform transforms prefault enum
func (e ZodEnumPrefault[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	return e.inner.Transform(fn)
}

// Enum returns enum mapping
func (e ZodEnumPrefault[T]) Enum() map[string]T {
	return e.inner.Enum()
}

// Options returns all possible values
func (e ZodEnumPrefault[T]) Options() []T {
	return e.inner.Options()
}

// Extract creates sub-enum with specified keys
func (e ZodEnumPrefault[T]) Extract(keys []string, params ...SchemaParams) ZodEnumPrefault[T] {
	extracted := e.inner.Extract(keys, params...)
	return ZodEnumPrefault[T]{
		inner:         extracted,
		prefaultValue: e.prefaultValue,
		prefaultFunc:  e.prefaultFunc,
		isFunction:    e.isFunction,
	}
}

// Exclude creates sub-enum excluding specified keys
func (e ZodEnumPrefault[T]) Exclude(keys []string, params ...SchemaParams) ZodEnumPrefault[T] {
	excluded := e.inner.Exclude(keys, params...)
	return ZodEnumPrefault[T]{
		inner:         excluded,
		prefaultValue: e.prefaultValue,
		prefaultFunc:  e.prefaultFunc,
		isFunction:    e.isFunction,
	}
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodEnumFromDef creates enum from definition
func createZodEnumFromDef[T comparable](def *ZodEnumDef[T]) *ZodEnum[T] {
	values := make(map[T]struct{})
	for _, value := range def.Entries {
		values[value] = struct{}{}
	}

	valueSlice := make([]interface{}, 0, len(values))
	for value := range values {
		valueSlice = append(valueSlice, value)
	}

	pattern := createEnumPattern(valueSlice)

	internals := &ZodEnumInternals[T]{
		ZodTypeInternals: ZodTypeInternals{
			Version: Version,
			Type:    def.Type,
			Checks:  make([]ZodCheck, 0),
			Nilable: false,
			Bag:     make(map[string]interface{}),
		},
		Def:     def,
		Entries: def.Entries,
		Values:  values,
		Pattern: pattern,
		Isst: ZodIssueInvalidValue{
			ZodIssueBase: ZodIssueBase{
				Code:    "invalid_value",
				Message: "Invalid enum value",
			},
		},
		Bag: make(map[string]interface{}),
	}

	return &ZodEnum[T]{internals: internals}
}

// NewZodEnum creates a new type-safe enum
func NewZodEnum[T comparable](entries map[string]T, params ...SchemaParams) *ZodEnum[T] {
	mergedParams := mergeEnumSchemaParams(nil, params...)

	def := &ZodEnumDef[T]{
		ZodTypeDef: ZodTypeDef{
			Type: "enum",
		},
		Type:    "enum",
		Entries: entries,
	}

	enum := createZodEnumFromDef(def)

	// Apply parameters
	if mergedParams.Error != nil {
		if errorStr, ok := mergedParams.Error.(string); ok {
			enum.internals.Isst.Message = errorStr
		}
	}

	if mergedParams.Coerce {
		enum.internals.Bag["coerce"] = true
	}
	if mergedParams.Description != "" {
		enum.internals.Bag["description"] = mergedParams.Description
	}

	return enum
}

// Enum creates type-safe enum from values
func Enum[T comparable](values ...T) *ZodEnum[T] {
	if len(values) == 0 {
		return NewZodEnum(map[string]T{})
	}

	entries := make(map[string]T)
	for i, value := range values {
		key := fmt.Sprintf("%d", i)
		entries[key] = value
	}

	return NewZodEnum(entries)
}

// EnumMap creates type-safe enum from mapping
func EnumMap[T comparable](entries map[string]T) *ZodEnum[T] {
	return NewZodEnum(entries)
}

// EnumSlice creates type-safe enum from slice
func EnumSlice[T comparable](values []T) *ZodEnum[T] {
	entries := make(map[string]T)
	for i, value := range values {
		key := fmt.Sprintf("%d", i)
		entries[key] = value
	}
	return NewZodEnum(entries)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// GetZod returns internal state (backward compatibility)
func (z *ZodEnum[T]) GetZod() *ZodEnumInternals[T] {
	return z.internals
}

// clone creates a copy of the enum
func (z *ZodEnum[T]) clone() *ZodEnum[T] {
	newInternals := &ZodEnumInternals[T]{
		ZodTypeInternals: z.internals.ZodTypeInternals,
		Def:              z.internals.Def,
		Entries:          make(map[string]T),
		Values:           make(map[T]struct{}),
		Pattern:          z.internals.Pattern,
		Isst:             z.internals.Isst,
		Bag:              make(map[string]interface{}),
	}

	// Deep copy Entries
	for k, v := range z.internals.Entries {
		newInternals.Entries[k] = v
	}

	// Deep copy Values
	for v := range z.internals.Values {
		newInternals.Values[v] = struct{}{}
	}

	// Deep copy Bag
	for k, v := range z.internals.Bag {
		newInternals.Bag[k] = v
	}

	return &ZodEnum[T]{internals: newInternals}
}

// mergeEnumSchemaParams merges enum schema parameters
func mergeEnumSchemaParams(bag map[string]interface{}, params ...SchemaParams) SchemaParams {
	result := SchemaParams{}

	for _, param := range params {
		if param.Error != nil {
			result.Error = param.Error
		}
		if param.Description != "" {
			result.Description = param.Description
		}
		if param.Coerce {
			result.Coerce = true
		}
		if param.Abort {
			result.Abort = true
		}
		if len(param.Path) > 0 {
			result.Path = param.Path
		}
		if len(param.Params) > 0 {
			result.Params = param.Params
		}
	}

	return result
}

// createEnumRefinementCheck creates enum refinement check
func createEnumRefinementCheck(fn func(any) bool, params ...SchemaParams) ZodCheck {
	return NewCustom[any](fn, params...)
}

// convertValueSetToInterface converts typed value set to interface{} set
func convertValueSetToInterface[T comparable](values map[T]struct{}) map[interface{}]struct{} {
	result := make(map[interface{}]struct{})
	for value := range values {
		result[value] = struct{}{}
	}
	return result
}

// createEnumPattern creates regex pattern for enum values
func createEnumPattern(values []interface{}) *regexp.Regexp {
	if len(values) == 0 {
		return regexp.MustCompile("^$a") // Never matches
	}

	patterns := make([]string, len(values))
	for i, value := range values {
		patterns[i] = regexp.QuoteMeta(fmt.Sprintf("%v", value))
	}

	pattern := "^(" + strings.Join(patterns, "|") + ")$"
	return regexp.MustCompile(pattern)
}

// PointerValue structure for pointer value extraction
type PointerValue struct {
	value any
	isNil bool
}

// extractPointerValue extracts pointer value with nil checking
func extractPointerValue(input any) *PointerValue {
	switch v := input.(type) {
	case *string:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *int:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *int8:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *int16:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *int32:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *int64:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *uint:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *uint8:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *uint16:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *uint32:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *uint64:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *float32:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *float64:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	case *bool:
		if v == nil {
			return &PointerValue{value: nil, isNil: true}
		}
		return &PointerValue{value: *v, isNil: false}
	default:
		return nil
	}
}
