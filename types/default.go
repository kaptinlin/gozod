package types

import "github.com/kaptinlin/gozod/core"

///////////////////////////
////   DEFAULT TYPE DEFINITIONS ////
///////////////////////////

// ZodDefault represents a validation schema with default value
// Core design: wraps inner type, delegates methods to preserve all functionality
type ZodDefault[T core.ZodType[any, any]] struct {
	innerType    T          // Inner type (cannot embed type parameter, use field)
	defaultValue any        // Default value
	defaultFunc  func() any // Default function
	isFunction   bool       // Whether using function to provide default value
}

///////////////////////////
////   CORE METHODS     ////
///////////////////////////

// GetInternals returns the internal state of the schema with modified optionality
func (z *ZodDefault[T]) GetInternals() *core.ZodTypeInternals {
	innerInternals := z.innerType.GetInternals()

	// Create a copy with modified optionality
	internals := *innerInternals // Copy the struct
	internals.OptIn = "optional" // Default makes input optional
	internals.OptOut = ""        // But output is not optional (always has a value)

	return &internals
}

// GetZod returns the default-specific internals (type-safe access)
func (z *ZodDefault[T]) GetZod() *core.ZodTypeInternals {
	// For default wrapper, return modified internals
	return z.GetInternals()
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodDefault[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodDefault[T]); ok {
		// Copy all default-specific state
		z.innerType = src.innerType
		z.defaultValue = src.defaultValue
		z.defaultFunc = src.defaultFunc
		z.isFunction = src.isFunction
	}
}

// Parse validates and parses input with smart type inference
func (z *ZodDefault[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	if input == nil {
		var defaultValue any
		if z.isFunction && z.defaultFunc != nil {
			defaultValue = z.defaultFunc()
		} else {
			defaultValue = z.defaultValue
		}

		// Check if inner type is Never, Never type's default value doesn't need validation
		internals := z.innerType.GetInternals()
		if internals != nil {
			if internals.Type == "never" {
				// Never type's default value returns directly without validation
				return defaultValue, nil
			}

			// Special-case: stringbool accepts string inputs only. Its default value is always bool.
			if internals.Type == "stringbool" {
				if _, ok := defaultValue.(bool); ok {
					return defaultValue, nil
				}
			}
		}

		// Other type's default values need validation through inner type
		return z.innerType.Parse(defaultValue, ctx...)
	}

	// Delegate to inner type's Parse (preserving smart type inference)
	return z.innerType.Parse(input, ctx...)
}

// MustParse validates the input value and panics on failure
func (z *ZodDefault[T]) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Check adds custom validation logic (delegates to inner type)
func (z *ZodDefault[T]) Check(fn core.CheckFn) core.ZodType[any, any] {
	// Check if inner type supports Check method
	if checkMethod, ok := any(z.innerType).(interface {
		Check(core.CheckFn) core.ZodType[any, any]
	}); ok {
		newInner := checkMethod.Check(fn)
		// Try to cast ZodType[any, any] back to T
		if castedInner, ok := newInner.(T); ok {
			return &ZodDefault[T]{
				innerType:    castedInner,
				defaultValue: z.defaultValue,
				defaultFunc:  z.defaultFunc,
				isFunction:   z.isFunction,
			}
		}
		// If casting fails, return inner type's check result
		return newInner
	}

	// If inner type doesn't support Check, return self
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a new transform with another transformation
func (z *ZodDefault[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a new transform with flexible transformation
// Wraps transform function to handle default value logic first, then transform
func (z *ZodDefault[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Create a wrapped transform function that handles default value logic first, then transforms
	wrappedFn := func(input any, ctx *core.RefinementContext) (any, error) {
		// First apply default value logic (same as Parse method)
		var processedInput any
		if input == nil {
			if z.isFunction && z.defaultFunc != nil {
				processedInput = z.defaultFunc()
			} else {
				processedInput = z.defaultValue
			}
		} else {
			// Delegate to inner type for validation first
			result, err := z.innerType.Parse(input)
			if err != nil {
				return nil, err
			}
			processedInput = result
		}

		// Then apply transform function
		return fn(processedInput, ctx)
	}

	// Create new Transform that bypasses inner type and directly uses processed input
	transform := Transform[any, any](wrappedFn)
	return transform
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional adds an optional check to the default schema, returns ZodType support chain call
func (z *ZodDefault[T]) Optional() core.ZodType[any, any] {
	// Apply to inner type
	if optMethod, ok := any(z.innerType).(interface{ Optional() core.ZodType[any, any] }); ok {
		return optMethod.Optional()
	}
	return any(z).(core.ZodType[any, any])
}

// Nilable adds a nilable check to the default schema, returns ZodType support chain call
func (z *ZodDefault[T]) Nilable() core.ZodType[any, any] {
	nilableInner := z.innerType.Nilable()
	return nilableInner
}

// Nullish makes the schema both optional and nilable (already optional via default, so just nilable)
func (z *ZodDefault[T]) Nullish() core.ZodType[any, any] {
	return z.Nilable()
}

// Refine adds a flexible validation function to the default schema, returns ZodDefault support chain call
func (z *ZodDefault[T]) Refine(fn func(string) bool, params ...any) *ZodDefault[T] {
	if refineMethod, ok := any(z.innerType).(interface {
		Refine(func(string) bool, ...any) T
	}); ok {
		newInner := refineMethod.Refine(fn, params...)
		return &ZodDefault[T]{
			innerType:    newInner,
			defaultValue: z.defaultValue,
			defaultFunc:  z.defaultFunc,
			isFunction:   z.isFunction,
		}
	}
	return z
}

// RefineAny general validation method - returns ZodType[any, any] to implement interface
func (z *ZodDefault[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	newInner := z.innerType.RefineAny(fn, params...)
	// Try to cast ZodType[any, any] back to T
	if castedInner, ok := newInner.(T); ok {
		return &ZodDefault[T]{
			innerType:    castedInner,
			defaultValue: z.defaultValue,
			defaultFunc:  z.defaultFunc,
			isFunction:   z.isFunction,
		}
	}
	// If casting fails, return inner type's refine result
	return newInner
}

// Default creates a default schema wrapper with static value
func Default[T interface{ GetInternals() *core.ZodTypeInternals }](innerType T, defaultValue any) core.ZodType[any, any] {
	anyInnerType := any(innerType).(core.ZodType[any, any])
	return &ZodDefault[core.ZodType[any, any]]{
		innerType:    anyInnerType,
		defaultValue: defaultValue,
		defaultFunc:  nil,
		isFunction:   false,
	}
}

// DefaultFunc creates a default schema wrapper with function-provided value
func DefaultFunc[T interface{ GetInternals() *core.ZodTypeInternals }](innerType T, fn func() any) core.ZodType[any, any] {
	anyInnerType := any(innerType).(core.ZodType[any, any])
	return &ZodDefault[core.ZodType[any, any]]{
		innerType:    anyInnerType,
		defaultValue: nil,
		defaultFunc:  fn,
		isFunction:   true,
	}
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// Add any necessary utility functions here
func (z *ZodDefault[T]) Unwrap() core.ZodType[any, any] {
	return any(z.innerType).(core.ZodType[any, any])
}

// Pipe creates a validation pipeline
// Fix: Pass the Default wrapper itself, not the inner type
// This preserves the Default's optionality (OptIn="optional")
func (z *ZodDefault[T]) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]), // Use the Default wrapper
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// Prefault fallback value when validation fails - creates new Prefault wrapper
func (z *ZodDefault[T]) Prefault(value any) core.ZodType[any, any] {
	// Create a special Prefault wrapper that knows how to handle Default
	return &ZodPrefault[core.ZodType[any, any]]{
		internals: &core.ZodTypeInternals{
			Type:   "prefault",
			OptIn:  "optional", // Inherit Default's OptIn
			OptOut: "",         // Inherit Default's OptOut
			Parse: func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
				// Special logic: for nil input, directly use Default's logic, don't go through Prefault validation
				if payload.Value == nil {
					// Use Default's default value without validation
					var defaultValue any
					if z.isFunction && z.defaultFunc != nil {
						defaultValue = z.defaultFunc()
					} else {
						defaultValue = z.defaultValue
					}
					payload.Value = defaultValue
					return payload
				}

				// For non-nil input, try validation first, use Prefault value on failure
				result, err := z.innerType.Parse(payload.Value, ctx)
				if err == nil {
					payload.Value = result
					return payload
				}

				// Validation failed, use Prefault value
				payload.Value = value
				return payload
			},
		},
		innerType:     any(z).(core.ZodType[any, any]),
		prefaultValue: value,
		isFunction:    false,
	}
}

// PrefaultFunc function-based fallback value - creates new Prefault wrapper
func (z *ZodDefault[T]) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	// Create new Prefault wrapping current Default type with function
	return &ZodPrefault[core.ZodType[any, any]]{
		internals: &core.ZodTypeInternals{
			Type:   "prefault",
			OptIn:  "optional", // Inherit Default's OptIn
			OptOut: "",         // Inherit Default's OptOut
			Parse: func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
				// Special logic: for nil input, directly use Default's logic
				if payload.Value == nil {
					// Use Default's default value without validation
					var defaultValue any
					if z.isFunction && z.defaultFunc != nil {
						defaultValue = z.defaultFunc()
					} else {
						defaultValue = z.defaultValue
					}
					payload.Value = defaultValue
					return payload
				}

				// For non-nil input, try validation first, use Prefault function on failure
				result, err := z.innerType.Parse(payload.Value, ctx)
				if err == nil {
					payload.Value = result
					return payload
				}

				// Validation failed, use Prefault function value
				payload.Value = fn()
				return payload
			},
		},
		innerType:    any(z).(core.ZodType[any, any]),
		prefaultFunc: fn,
		isFunction:   true,
	}
}
