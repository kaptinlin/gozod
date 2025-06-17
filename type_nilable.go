package gozod

import (
	"reflect"
	"regexp"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodNilableDef defines the configuration for nilable validation
type ZodNilableDef[T ZodType[any, any]] struct {
	ZodTypeDef
	Type      string // "nilable"
	InnerType T      // The wrapped type - using generic parameter
}

// ZodNilableInternals contains nilable validator internal state
type ZodNilableInternals[T ZodType[any, any]] struct {
	ZodTypeInternals
	Def     *ZodNilableDef[T]        // Nilable definition with generic
	Values  map[interface{}]struct{} // Inherited from inner type values
	Pattern *regexp.Regexp           // Inherited from inner type pattern
}

// ZodNilable represents a nilable validation schema
// Core design: contains inner type, obtains all methods through method forwarding
type ZodNilable[T ZodType[any, any]] struct {
	innerType T                 // Inner type (cannot embed type parameters, use fields)
	internals *ZodTypeInternals // Nilable's own internals
}

// GetInternals returns the internal state of the schema
func (z *ZodNilable[T]) GetInternals() *ZodTypeInternals {
	// Nilable needs its own internals to correctly handle nil values
	if z.internals == nil {
		innerInternals := z.innerType.GetInternals()
		z.internals = &ZodTypeInternals{
			Type:   "nilable",
			OptIn:  innerInternals.OptIn,  // Preserve input optionality
			OptOut: innerInternals.OptOut, // Preserve output optionality
			Parse: func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
				if payload.Value == nil || (reflect.ValueOf(payload.Value).Kind() == reflect.Ptr && reflect.ValueOf(payload.Value).IsNil()) {
					innerTypeInternals := z.innerType.GetInternals()
					if innerTypeInternals != nil {
						switch innerTypeInternals.Type {
						case "nil":
							payload.Value = nil
							return payload
						case "string":
							payload.Value = (*string)(nil)
						case "bool", "boolean":
							payload.Value = (*bool)(nil)
						case "stringbool":
							payload.Value = (*bool)(nil) // StringBool outputs bool, so nil returns *bool(nil)
						case "int", "int8", "int16", "int32", "int64":
							payload.Value = (*int)(nil)
						case "uint", "uint8", "uint16", "uint32", "uint64":
							payload.Value = (*uint)(nil)
						case "float32", "float64", "number":
							payload.Value = (*float64)(nil)
						case "complex64":
							payload.Value = (*complex64)(nil)
						case "complex128":
							payload.Value = (*complex128)(nil)
						case "any":
							payload.Value = (*interface{})(nil)
						default:
							// For other types, return generic nil pointer
							payload.Value = (*interface{})(nil)
						}
						return payload
					}

					// Delegate to inner type's Parse
					return z.innerType.GetInternals().Parse(payload, ctx)
				}

				// Delegate to inner type's Parse
				return z.innerType.GetInternals().Parse(payload, ctx)
			},
		}
	}
	return z.internals
}

// GetZod returns the nilable-specific internals (type-safe access)
func (z *ZodNilable[T]) GetZod() *ZodNilableInternals[T] {
	// Return nilable-specific internals if available
	return &ZodNilableInternals[T]{
		ZodTypeInternals: *z.GetInternals(),
		Def: &ZodNilableDef[T]{
			ZodTypeDef: ZodTypeDef{Type: "nilable"},
			Type:       "nilable",
			InnerType:  z.innerType,
		},
	}
}

// Coerce attempts to coerce input (delegates to inner type)
func (z *ZodNilable[T]) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Parse validates and parses input with smart type inference
// Core: only handles nil, delegates everything else to inner type
func (z *ZodNilable[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	// Core: only handles nil, delegates everything else to inner type
	if input == nil || (reflect.ValueOf(input).Kind() == reflect.Ptr && reflect.ValueOf(input).IsNil()) {
		// Determine inner type and return corresponding typed nil pointer
		innerTypeInternals := z.innerType.GetInternals()
		if innerTypeInternals != nil {
			switch innerTypeInternals.Type {
			case "nil":
				// For nil type, return true nil
				return nil, nil
			case "string":
				return (*string)(nil), nil
			case "bool", "boolean":
				return (*bool)(nil), nil
			case "stringbool":
				return (*bool)(nil), nil // StringBool outputs bool, so nil returns *bool(nil)
			case "int", "int8", "int16", "int32", "int64":
				return (*int)(nil), nil
			case "uint", "uint8", "uint16", "uint32", "uint64":
				return (*uint)(nil), nil
			case "float32", "float64", "number":
				return (*float64)(nil), nil
			case "complex64":
				return (*complex64)(nil), nil
			case "complex128":
				return (*complex128)(nil), nil
			case "any":
				return (*interface{})(nil), nil
			default:
				// For other types, return generic nil pointer
				return (*interface{})(nil), nil
			}
		}
		// If inner type information cannot be obtained, return generic nil pointer
		return (*interface{})(nil), nil
	}

	// Fully delegate to inner type, maintain smart inference
	return z.innerType.Parse(input, ctx...)
}

// MustParse validates the input value and panics on failure
func (z *ZodNilable[T]) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Generic validation method - returns ZodType[any, any] to implement interface
func (z *ZodNilable[T]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	// For nilable values, special handling is required for nil values
	wrappedFn := func(value any) bool {
		// Nilable refine: nil values always pass validation (indicates value can be explicitly null)
		if value == nil {
			return true
		}
		// Validate existing values
		return fn(value)
	}

	newInner := z.innerType.RefineAny(wrappedFn, params...)
	// Need to convert ZodType[any, any] back to T
	if castedInner, ok := newInner.(T); ok {
		return &ZodNilable[T]{
			innerType: castedInner,
		}
	}
	// If conversion fails, return the refine result of the inner type
	return newInner
}

// Refine adds type-safe custom validation logic to the nilable schema
func (z *ZodNilable[T]) Refine(fn func(any) bool, params ...SchemaParams) *ZodNilable[T] {
	// For nilable values, special handling is required for nil values
	wrappedFn := func(value any) bool {
		// Nilable refine: nil values always pass validation (indicates value can be explicitly null)
		if value == nil {
			return true
		}
		// Validate existing values
		return fn(value)
	}

	if refineMethod, ok := any(z.innerType).(interface {
		Refine(func(any) bool, ...SchemaParams) T
	}); ok {
		newInner := refineMethod.Refine(wrappedFn, params...)
		return &ZodNilable[T]{
			innerType: newInner,
		}
	}
	return z
}

// Optional makes the nilable schema optional
func (z *ZodNilable[T]) Optional() ZodType[any, any] {
	optionalInner := Optional(any(z).(ZodType[any, any]))
	return optionalInner
}

// Nilable makes the nilable schema nilable
func (z *ZodNilable[T]) Nilable() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// TransformAny creates a new transform with another transformation
func (z *ZodNilable[T]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	wrappedFn := func(value any, ctx *RefinementContext) (any, error) {
		if value == nil {
			return fn(nil, ctx)
		}

		// For non-nil values, first delegate to inner type for validation
		result, err := z.innerType.Parse(value)
		if err != nil {
			return nil, err
		}

		// Execute conversion on validated value
		return fn(result, ctx)
	}

	// Create a new Transform, bypassing inner type and using processed input directly
	transform := NewZodTransform[any, any](wrappedFn)
	return transform
}

// Pipe creates a validation pipeline
func (z *ZodNilable[T]) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Prefault provides a fallback value when validation fails
func (z *ZodNilable[T]) Prefault(value any) ZodType[any, any] {
	// Create new Prefault wrapper for current Nilable type
	return &ZodPrefault[ZodType[any, any]]{
		innerType:     any(z).(ZodType[any, any]),
		prefaultValue: value,
		isFunction:    false,
	}
}

// PrefaultFunc provides a fallback value based on a function
func (z *ZodNilable[T]) PrefaultFunc(fn func() any) ZodType[any, any] {
	// Create new Prefault wrapper for current Nilable type
	return &ZodPrefault[ZodType[any, any]]{
		innerType:    any(z).(ZodType[any, any]),
		prefaultFunc: fn,
		isFunction:   true,
	}
}

// Check adds custom validation logic
func (z *ZodNilable[T]) Check(fn func(*ParsePayload) error) ZodType[any, any] {
	// For nilable values, special handling is required for nil values
	wrappedFn := func(payload *ParsePayload) error {
		// Nilable check: nil values skip check (indicates value can be explicitly null)
		if payload.Value == nil {
			return nil
		}
		// Validate existing values
		return fn(payload)
	}

	// Check if inner type supports Check method
	if checkMethod, ok := any(z.innerType).(interface {
		Check(func(*ParsePayload) error) ZodType[any, any]
	}); ok {
		newInner := checkMethod.Check(wrappedFn)
		// Need to convert ZodType[any, any] back to T
		if castedInner, ok := newInner.(T); ok {
			return &ZodNilable[T]{
				innerType: castedInner,
			}
		}
		// If conversion fails, return the check result of the inner type
		return newInner
	}

	// If inner type does not support Check, return
	return any(z).(ZodType[any, any])
}

// Unwrap returns the inner type
func (z *ZodNilable[T]) Unwrap() ZodType[any, any] {
	return any(z.innerType).(ZodType[any, any])
}

////////////////////////////
////   PACKAGE FUNCTIONS ////
////////////////////////////

// Nilable creates a nilable schema wrapper
func Nilable[T interface{ GetInternals() *ZodTypeInternals }](innerType T, params ...SchemaParams) ZodType[any, any] {
	// Directly use type constraints, avoid complex type conversions
	anyInnerType := any(innerType).(ZodType[any, any])
	return &ZodNilable[ZodType[any, any]]{
		innerType: anyInnerType,
	}
}

// Compatibility function - maintain backward compatibility
func NewZodNilable(innerType ZodType[any, any]) *ZodNilable[ZodType[any, any]] {
	return &ZodNilable[ZodType[any, any]]{
		innerType: innerType,
	}
}

////////////////////////////
////   DEFAULT COMBINED TYPE   ////
////////////////////////////

// ZodNilableDefault is the Default wrapper for nilable types
// Provides perfect type safety and chainable support
type ZodNilableDefault[T ZodType[any, any]] struct {
	*ZodDefault[*ZodNilable[T]] // Embed generic wrapper
}

// Parse method rewritten to correctly handle Nilable + Default combination logic
func (s *ZodNilableDefault[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	// Nilable + Default logic: use default value when nil, otherwise delegate to Nilable
	if input == nil {
		// Use Default logic
		return s.ZodDefault.Parse(input, ctx...)
	}
	// Delegate to inner Nilable
	return s.ZodDefault.innerType.Parse(input, ctx...)
}

// Default adds a default value to Nilable
func (z *ZodNilable[T]) Default(value any) *ZodNilableDefault[T] {
	defaultWrapper := &ZodDefault[*ZodNilable[T]]{
		innerType:    z,
		defaultValue: value,
		isFunction:   false,
	}
	return &ZodNilableDefault[T]{
		ZodDefault: defaultWrapper,
	}
}

// DefaultFunc adds a function default value to Nilable
func (z *ZodNilable[T]) DefaultFunc(fn func() any) *ZodNilableDefault[T] {
	defaultWrapper := &ZodDefault[*ZodNilable[T]]{
		innerType:   z,
		defaultFunc: fn,
		isFunction:  true,
	}
	return &ZodNilableDefault[T]{
		ZodDefault: defaultWrapper,
	}
}

// Refine adds type-safe custom validation logic to the nilable schema
func (s *ZodNilableDefault[T]) Refine(fn func(any) bool, params ...SchemaParams) *ZodNilableDefault[T] {
	newInner := s.ZodDefault.innerType.Refine(fn, params...)
	newDefault := &ZodDefault[*ZodNilable[T]]{
		innerType:    newInner,
		defaultValue: s.ZodDefault.defaultValue,
		defaultFunc:  s.ZodDefault.defaultFunc,
		isFunction:   s.ZodDefault.isFunction,
	}
	return &ZodNilableDefault[T]{
		ZodDefault: newDefault,
	}
}

// Transform adds data transformation, returns a generic ZodType support for transformation pipeline
func (s *ZodNilableDefault[T]) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// use the TransformAny method of the embedded ZodDefault
	return s.TransformAny(fn)
}
