package types

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodEnumDef defines the configuration for enum validation
type ZodEnumDef[T comparable] struct {
	core.ZodTypeDef
	Type    core.ZodTypeCode // Type identifier using type-safe constants
	Entries map[string]T     // Enum entries (key-value pairs)
}

// ZodEnumInternals contains enum validator internal state
type ZodEnumInternals[T comparable] struct {
	core.ZodTypeInternals
	Def     *ZodEnumDef[T]              // Schema definition
	Entries map[string]T                // Enum entries
	Values  map[T]struct{}              // Set of valid values
	Pattern *regexp.Regexp              // Regex pattern for validation
	Isst    issues.ZodIssueInvalidValue // Invalid value issue template
	Bag     map[string]any              // Additional metadata
}

// ZodEnum represents a type-safe enum validation schema
type ZodEnum[T comparable] struct {
	internals *ZodEnumInternals[T]
}

// GetInternals returns the internal state of the schema
func (z *ZodEnum[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse validates and parses input using the unified engine.ParseType template
func (z *ZodEnum[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Use engine.ParseType template for unified parsing
	result, err := engine.ParseType[T](
		input,
		&z.internals.ZodTypeInternals,
		"enum",
		// Type checker function
		func(v any) (T, bool) {
			if value, ok := v.(T); ok {
				if _, exists := z.internals.Values[value]; exists {
					return value, true
				}
			}
			var zero T
			return zero, false
		},
		// Pointer checker function
		func(v any) (*T, bool) {
			// Match only when the input is *T itself, preserving pointer identity
			ptr, ok := v.(*T)
			return ptr, ok
		},
		// Validator function
		func(value T, checks []core.ZodCheck, ctx *core.ParseContext) error {
			if len(checks) > 0 {
				payload := &core.ParsePayload{
					Value:  value,
					Issues: make([]core.ZodRawIssue, 0),
				}
				engine.RunChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					finalizedIssues := make([]core.ZodIssue, len(payload.Issues))
					for i, rawIssue := range payload.Issues {
						finalizedIssues[i] = issues.FinalizeIssue(rawIssue, ctx, core.GetConfig())
					}
					return issues.NewZodError(finalizedIssues)
				}
			}
			return nil
		},
		parseCtx,
	)

	if err != nil {
		var zErr *issues.ZodError
		if errors.As(err, &zErr) {
			for i, iss := range zErr.Issues {
				if iss.Code == "invalid_type" {
					zErr.Issues[i].Code = "invalid_value"
					zErr.Issues[i].Message = "invalid enum value"
				}
			}
		}
		return nil, err
	}
	return result, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodEnum[T]) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodEnum[T]) Extract(keys []string, params ...any) *ZodEnum[T] {
	newEntries := make(map[string]T)
	for _, key := range keys {
		if value, exists := z.internals.Entries[key]; exists {
			newEntries[key] = value
		} else {
			panic(fmt.Sprintf("Key '%s' not found in enum", key))
		}
	}

	return EnumMap(newEntries, params...)
}

// Exclude creates a sub-enum excluding specified keys
func (z *ZodEnum[T]) Exclude(keys []string, params ...any) *ZodEnum[T] {
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

	return EnumMap(newEntries, params...)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe enum transformation (value of enum -> any)
func (z *ZodEnum[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	wrapper := func(input any, ctx *core.RefinementContext) (any, error) {
		// Support both pointer and value inputs
		if val, ok := input.(T); ok {
			return fn(val, ctx)
		}
		if ptr, ok := input.(*T); ok && ptr != nil {
			return fn(*ptr, ctx)
		}
		return nil, fmt.Errorf("invalid type for transform")
	}
	return z.TransformAny(wrapper)
}

// TransformAny builds the actual transform pipeline (enum -> transform)
func (z *ZodEnum[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodEnum[T]) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the enum optional
func (z *ZodEnum[T]) Optional() core.ZodType[any, any] {
	// Use the generic Optional wrapper to ensure nil is returned when the value is absent
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the enum nilable
func (z *ZodEnum[T]) Nilable() core.ZodType[any, any] {
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Nullish makes the enum both optional and nilable
func (z *ZodEnum[T]) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Refine adds type-safe custom validation
func (z *ZodEnum[T]) Refine(fn func(T) bool, params ...any) *ZodEnum[T] {
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
	check := checks.NewCustom[any](func(val any) bool {
		if v, ok := val.(T); ok {
			return fn(v)
		}
		return false
	}, params...)
	clone.internals.Checks = append(clone.internals.Checks, core.ZodCheck(check))
	return clone
}

// RefineAny adds flexible custom validation
func (z *ZodEnum[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	clone := z.clone()
	check := checks.NewCustom[any](fn, params...)
	clone.internals.Checks = append(clone.internals.Checks, core.ZodCheck(check))
	return any(clone).(core.ZodType[any, any])
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodEnum[T]) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
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
func (s ZodEnumDefault[T]) GetInternals() *core.ZodTypeInternals {
	return s.inner.GetInternals()
}

// Parse method with default value support
func (s ZodEnumDefault[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	if input == nil {
		if s.isFunction && s.defaultFunc != nil {
			return s.defaultFunc(), nil
		}
		return s.defaultValue, nil
	}
	return s.inner.Parse(input, ctx...)
}

// MustParse method
func (s ZodEnumDefault[T]) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := s.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// TransformAny method
func (s ZodEnumDefault[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.inner.TransformAny(fn)
}

// Pipe method
func (s ZodEnumDefault[T]) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return s.inner.Pipe(out)
}

// Optional method
func (s ZodEnumDefault[T]) Optional() core.ZodType[any, any] {
	return s.inner.Optional()
}

// Nilable method
func (s ZodEnumDefault[T]) Nilable() core.ZodType[any, any] {
	return s.inner.Nilable()
}

// Nullish method
func (s ZodEnumDefault[T]) Nullish() core.ZodType[any, any] {
	return s.inner.Nullish()
}

// RefineAny method
func (s ZodEnumDefault[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return s.inner.RefineAny(fn, params...)
}

// Unwrap method
func (s ZodEnumDefault[T]) Unwrap() core.ZodType[any, any] {
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
func (s ZodEnumDefault[T]) Refine(fn func(T) bool, params ...any) ZodEnumDefault[T] {
	refined := s.inner.Refine(fn, params...)
	return ZodEnumDefault[T]{
		inner:        refined,
		defaultValue: s.defaultValue,
		defaultFunc:  s.defaultFunc,
		isFunction:   s.isFunction,
	}
}

// Transform transforms default enum
func (s ZodEnumDefault[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
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
func (s ZodEnumDefault[T]) Extract(keys []string, params ...any) ZodEnumDefault[T] {
	extracted := s.inner.Extract(keys, params...)
	return ZodEnumDefault[T]{
		inner:        extracted,
		defaultValue: s.defaultValue,
		defaultFunc:  s.defaultFunc,
		isFunction:   s.isFunction,
	}
}

// Exclude creates sub-enum excluding specified keys
func (s ZodEnumDefault[T]) Exclude(keys []string, params ...any) ZodEnumDefault[T] {
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
func (e ZodEnumPrefault[T]) GetInternals() *core.ZodTypeInternals {
	return e.inner.GetInternals()
}

// Parse method with prefault value support
func (e ZodEnumPrefault[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
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
func (e ZodEnumPrefault[T]) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := e.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// TransformAny method
func (e ZodEnumPrefault[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return e.inner.TransformAny(fn)
}

// Pipe method
func (e ZodEnumPrefault[T]) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return e.inner.Pipe(out)
}

// Optional method
func (e ZodEnumPrefault[T]) Optional() core.ZodType[any, any] {
	return e.inner.Optional()
}

// Nilable method
func (e ZodEnumPrefault[T]) Nilable() core.ZodType[any, any] {
	return e.inner.Nilable()
}

// Nullish method
func (e ZodEnumPrefault[T]) Nullish() core.ZodType[any, any] {
	return e.inner.Nullish()
}

// RefineAny method
func (e ZodEnumPrefault[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return e.inner.RefineAny(fn, params...)
}

// Unwrap method
func (e ZodEnumPrefault[T]) Unwrap() core.ZodType[any, any] {
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
func (e ZodEnumPrefault[T]) Refine(fn func(T) bool, params ...any) ZodEnumPrefault[T] {
	refined := e.inner.Refine(fn, params...)
	return ZodEnumPrefault[T]{
		inner:         refined,
		prefaultValue: e.prefaultValue,
		prefaultFunc:  e.prefaultFunc,
		isFunction:    e.isFunction,
	}
}

// Transform transforms prefault enum
func (e ZodEnumPrefault[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
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
func (e ZodEnumPrefault[T]) Extract(keys []string, params ...any) ZodEnumPrefault[T] {
	extracted := e.inner.Extract(keys, params...)
	return ZodEnumPrefault[T]{
		inner:         extracted,
		prefaultValue: e.prefaultValue,
		prefaultFunc:  e.prefaultFunc,
		isFunction:    e.isFunction,
	}
}

// Exclude creates sub-enum excluding specified keys
func (e ZodEnumPrefault[T]) Exclude(keys []string, params ...any) ZodEnumPrefault[T] {
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

	valueSlice := make([]any, 0, len(values))
	for value := range values {
		valueSlice = append(valueSlice, value)
	}

	// Create pattern using existing utility
	var pattern *regexp.Regexp
	if len(valueSlice) > 0 {
		patterns := make([]string, len(valueSlice))
		for i, value := range valueSlice {
			patterns[i] = regexp.QuoteMeta(fmt.Sprintf("%v", value))
		}
		patternStr := "^(" + strings.Join(patterns, "|") + ")$"
		pattern = regexp.MustCompile(patternStr)
	} else {
		pattern = regexp.MustCompile("^$a") // Never matches
	}

	internals := &ZodEnumInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Entries:          def.Entries,
		Values:           values,
		Pattern:          pattern,
		Isst: issues.ZodIssueInvalidValue{
			ZodIssueBase: issues.ZodIssueBase{
				Code:    "invalid_value",
				Message: "Invalid enum value",
			},
			Values: valueSlice,
		},
		Bag: make(map[string]any),
	}

	zodSchema := &ZodEnum[T]{internals: internals}

	// Initialize the schema using engine
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)

	return zodSchema
}

// Enum creates a new enum from a list of values.
// Example: Enum("Red", "Green", "Blue")
func Enum[T comparable](values ...T) *ZodEnum[T] {
	return EnumSlice(values)
}

// EnumMap creates type-safe enum from mapping
func EnumMap[T comparable](entries map[string]T, params ...any) *ZodEnum[T] {
	def := &ZodEnumDef[T]{
		ZodTypeDef: core.ZodTypeDef{
			Type: "enum",
		},
		Type:    "enum",
		Entries: entries,
	}

	enum := createZodEnumFromDef(def)

	// Apply schema parameters following unified pattern
	if len(params) > 0 {
		if param, ok := params[0].(core.SchemaParams); ok {
			engine.ApplySchemaParams(&def.ZodTypeDef, param)

			if param.Error != nil {
				errorMap := issues.CreateErrorMap(param.Error)
				if errorMap != nil {
					def.Error = errorMap
					enum.internals.Error = errorMap
				}
			}
			if param.Description != "" {
				enum.internals.Bag["description"] = param.Description
			}
		}
	}

	return enum
}

// EnumSlice creates type-safe enum from slice
func EnumSlice[T comparable](values []T) *ZodEnum[T] {
	entries := make(map[string]T)
	for i, value := range values {
		key := fmt.Sprintf("%d", i)
		entries[key] = value
	}
	return EnumMap(entries)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// GetZod returns internal state (backward compatibility)
func (z *ZodEnum[T]) GetZod() *ZodEnumInternals[T] {
	return z.internals
}

// clone creates a copy of the enum using pkg utilities
func (z *ZodEnum[T]) clone() *ZodEnum[T] {
	newInternals := &ZodEnumInternals[T]{
		ZodTypeInternals: z.internals.ZodTypeInternals,
		Def:              z.internals.Def,
		Entries:          make(map[string]T),
		Values:           make(map[T]struct{}),
		Pattern:          z.internals.Pattern,
		Isst:             z.internals.Isst,
		Bag:              make(map[string]any),
	}

	// Deep copy Entries
	for k, v := range z.internals.Entries {
		newInternals.Entries[k] = v
	}

	// Deep copy Values
	for v := range z.internals.Values {
		newInternals.Values[v] = struct{}{}
	}

	// Deep copy Bag using mapx
	if z.internals.Bag != nil {
		newInternals.Bag = mapx.Copy(z.internals.Bag)
	}

	return &ZodEnum[T]{internals: newInternals}
}
