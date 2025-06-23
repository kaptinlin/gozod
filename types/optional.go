package types

import (
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodOptionalDef defines the configuration for optional validation
type ZodOptionalDef[T core.ZodType[any, any]] struct {
	core.ZodTypeDef
	Type      core.ZodTypeCode // "optional"
	InnerType T                // The wrapped type - using generic parameter
}

// ZodOptionalInternals contains optional validator internal state
type ZodOptionalInternals[T core.ZodType[any, any]] struct {
	core.ZodTypeInternals
	Def     *ZodOptionalDef[T] // Optional definition with generic
	OptIn   string             // "optional" - input optionality
	OptOut  string             // "optional" - output optionality
	Values  map[any]struct{}   // Inherited from inner type values
	Pattern *regexp.Regexp     // Inherited from inner type pattern
}

// ZodOptional represents an optional validation schema
// Core design: contains inner type, gets all methods through method forwarding
type ZodOptional[T core.ZodType[any, any]] struct {
	innerType T                      // Inner type (cannot embed type parameter, use field)
	internals *core.ZodTypeInternals // Optional's own internals
}

// GetInternals returns the internal state of the schema
func (z *ZodOptional[T]) GetInternals() *core.ZodTypeInternals {
	// Optional needs its own internals to properly handle nil values
	if z.internals == nil {
		z.internals = &core.ZodTypeInternals{
			Type:   core.ZodTypeOptional,
			OptIn:  "optional",
			OptOut: "optional",
			Parse: func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
				// Implement Optional's nil handling logic
				if reflectx.IsNil(payload.GetValue()) {
					// Optional allows missing values, return generic nil
					payload.SetValue(nil)
					return payload
				}

				// Delegate to inner type's Parse
				return z.innerType.GetInternals().Parse(payload, ctx)
			},
		}
	}
	return z.internals
}

// GetZod returns the optional-specific internals (type-safe access)
func (z *ZodOptional[T]) GetZod() *ZodOptionalInternals[T] {
	// Return optional-specific internals if available
	return &ZodOptionalInternals[T]{
		ZodTypeInternals: *z.GetInternals(),
		Def: &ZodOptionalDef[T]{
			ZodTypeDef: core.ZodTypeDef{Type: core.ZodTypeOptional},
			Type:       core.ZodTypeOptional,
			InnerType:  z.innerType,
		},
		OptIn:  "optional",
		OptOut: "optional",
	}
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodOptional[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodOptional[T]); ok {
		// Copy the inner type
		z.innerType = src.innerType
	}
}

// Parse validates and parses input with smart type inference
// - undefined/nil returns nil (field is missing)
// - otherwise delegates to inner type, preserving smart inference
func (z *ZodOptional[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	if reflectx.IsNil(input) {
		return nil, nil
	}

	return z.innerType.Parse(input, ctx...)
}

// MustParse validates and parses input with smart type inference
func (z *ZodOptional[T]) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

////////////////////////////
////   CHAIN METHODS    ////
////////////////////////////

// RefineAny is a generic refinement method that returns ZodType[any, any] to implement the interface
func (z *ZodOptional[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	wrappedFn := func(value any) bool {
		if reflectx.IsNil(value) {
			return true
		}
		return fn(value)
	}

	newInner := z.innerType.RefineAny(wrappedFn, params...)
	if castedInner, ok := newInner.(T); ok {
		return &ZodOptional[T]{
			innerType: castedInner,
		}
	}
	return newInner
}

// Refine is a type-safe refinement method
func (z *ZodOptional[T]) Refine(fn func(any) bool, params ...any) *ZodOptional[T] {
	wrappedFn := func(value any) bool {
		if reflectx.IsNil(value) {
			return true
		}
		return fn(value)
	}

	if refineMethod, ok := any(z.innerType).(interface {
		Refine(func(any) bool, ...any) T
	}); ok {
		newInner := refineMethod.Refine(wrappedFn, params...)
		return &ZodOptional[T]{
			innerType: newInner,
		}
	}
	return z
}

// Nilable is a modifier that returns a proper Nilable wrapper around the Optional
func (z *ZodOptional[T]) Nilable() core.ZodType[any, any] {
	// Create a proper Nilable wrapper around the Optional
	// This preserves the Optional's optionality information
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Optional is a modifier that returns itself (already optional)
func (z *ZodOptional[T]) Optional() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// TransformAny is a modifier that returns a proper Transform wrapper around the Optional
func (z *ZodOptional[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	wrappedFn := func(value any, ctx *core.RefinementContext) (any, error) {
		if reflectx.IsNil(value) {
			return fn(nil, ctx)
		}

		result, err := z.innerType.Parse(value)
		if err != nil {
			return nil, err
		}

		return fn(result, ctx)
	}

	transform := Transform[any, any](wrappedFn)
	return transform
}

// Pipe is a modifier that returns a proper Pipe wrapper around the Optional
func (z *ZodOptional[T]) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Prefault is a modifier that returns a proper Prefault wrapper around the Optional
func (z *ZodOptional[T]) Prefault(value any) core.ZodType[any, any] {
	return &ZodPrefault[core.ZodType[any, any]]{
		innerType:     any(z).(core.ZodType[any, any]),
		prefaultValue: value,
		isFunction:    false,
	}
}

// PrefaultFunc is a modifier that returns a proper Prefault wrapper around the Optional
func (z *ZodOptional[T]) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	return &ZodPrefault[core.ZodType[any, any]]{
		innerType:    any(z).(core.ZodType[any, any]),
		prefaultFunc: fn,
		isFunction:   true,
	}
}

// Check is a modifier that returns a proper Check wrapper around the Optional
func (z *ZodOptional[T]) Check(fn func(*core.ParsePayload) error) core.ZodType[any, any] {
	wrappedFn := func(payload *core.ParsePayload) error {
		if reflectx.IsNil(payload.GetValue()) {
			return nil
		}
		return fn(payload)
	}

	// Check if the inner type supports Check method
	if checkMethod, ok := any(z.innerType).(interface {
		Check(func(*core.ParsePayload) error) core.ZodType[any, any]
	}); ok {
		newInner := checkMethod.Check(wrappedFn)
		if castedInner, ok := newInner.(T); ok {
			return &ZodOptional[T]{
				innerType: castedInner,
			}
		}
		return newInner
	}

	return any(z).(core.ZodType[any, any])
}

// Unwrap returns the inner type
func (z *ZodOptional[T]) Unwrap() core.ZodType[any, any] {
	return any(z.innerType).(core.ZodType[any, any])
}

// IsOptional returns true if the type is optional
func (z *ZodOptional[T]) IsOptional() bool {
	return true
}

////////////////////////////
////   PACKAGE FUNCTIONS ////
////////////////////////////

// Optional creates an optional wrapper (improved version - automatic inference)
func Optional[T interface{ GetInternals() *core.ZodTypeInternals }](innerType T, params ...any) core.ZodType[any, any] {
	// Directly use type constraints, avoiding complex type conversions
	anyInnerType := any(innerType).(core.ZodType[any, any])
	return &ZodOptional[core.ZodType[any, any]]{
		innerType: anyInnerType,
	}
}

// Nullish creates an alias for the optional wrapper
func Nullish[T interface{ GetInternals() *core.ZodTypeInternals }](innerType T) core.ZodType[any, any] {
	return Optional(innerType)
}

////////////////////////////
////   DEFAULT COMBINED TYPE   ////
////////////////////////////

// ZodOptionalDefault is the Default wrapper for optional types
// Provides perfect type safety and chainable calls
type ZodOptionalDefault[T core.ZodType[any, any]] struct {
	*ZodDefault[*ZodOptional[T]] // embedded generic wrapper
}

// Parse method rewritten to correctly handle Optional + Default combination logic
func (s *ZodOptionalDefault[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// Optional + Default logic: use default value when nil, otherwise delegate to Optional
	if input == nil {
		// Use Default logic
		return s.ZodDefault.Parse(input, ctx...)
	}
	// Delegate to inner Optional
	return s.ZodDefault.innerType.Parse(input, ctx...)
}

// Default adds a default value to the Optional
func (z *ZodOptional[T]) Default(value any) *ZodOptionalDefault[T] {
	defaultWrapper := &ZodDefault[*ZodOptional[T]]{
		innerType:    z,
		defaultValue: value,
		isFunction:   false,
	}
	return &ZodOptionalDefault[T]{
		ZodDefault: defaultWrapper,
	}
}

// DefaultFunc adds a function default value to the Optional
func (z *ZodOptional[T]) DefaultFunc(fn func() any) *ZodOptionalDefault[T] {
	defaultWrapper := &ZodDefault[*ZodOptional[T]]{
		innerType:   z,
		defaultFunc: fn,
		isFunction:  true,
	}
	return &ZodOptionalDefault[T]{
		ZodDefault: defaultWrapper,
	}
}

// Refine adds a flexible validation function to the optional schema, return ZodOptionalDefault
func (s *ZodOptionalDefault[T]) Refine(fn func(any) bool, params ...any) *ZodOptionalDefault[T] {
	newInner := s.ZodDefault.innerType.Refine(fn, params...)
	newDefault := &ZodDefault[*ZodOptional[T]]{
		innerType:    newInner,
		defaultValue: s.ZodDefault.defaultValue,
		defaultFunc:  s.ZodDefault.defaultFunc,
		isFunction:   s.ZodDefault.isFunction,
	}
	return &ZodOptionalDefault[T]{
		ZodDefault: newDefault,
	}
}

// Transform adds data transformation, returns generic ZodType to support transformation pipeline
func (s *ZodOptionalDefault[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(fn)
}
