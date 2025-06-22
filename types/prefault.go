package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
)

///////////////////////////
////   PREFAULT TYPE   ////
///////////////////////////

// ZodPrefault represents a validation pattern with fallback values
// Core design: contains an inner type, and forwards all methods through it
type ZodPrefault[T core.ZodType[any, any]] struct {
	internals     *core.ZodTypeInternals // Prefault's own internals, Type = "prefault"
	innerType     T                      // inner type
	prefaultValue any                    // fallback value
	prefaultFunc  func() any             // fallback function
	isFunction    bool                   // whether to use function to provide fallback value
}

///////////////////////////
////   PREFAULT CORE   ////
///////////////////////////

// GetInternals returns the state of the inner type
func (z *ZodPrefault[T]) GetInternals() *core.ZodTypeInternals {
	return z.internals
}

// GetZod returns the prefault-specific internals (type-safe access)
func (z *ZodPrefault[T]) GetZod() *core.ZodTypeInternals {
	// for prefault wrapper, return its own internals
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodPrefault[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodPrefault[T]); ok {
		// copy all prefault-specific state
		z.innerType = src.innerType
		z.prefaultValue = src.prefaultValue
		z.prefaultFunc = src.prefaultFunc
		z.isFunction = src.isFunction
		z.internals = src.internals
	}
}

// Parse with smart type inference
// try to validate, and use fallback value if failed
// unlike Default (use default value when nil), Prefault always tries to validate first
// important: for nil input, Prefault should reject and try fallback value
func (z *ZodPrefault[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	var contextToUse *core.ParseContext
	if len(ctx) > 0 {
		contextToUse = ctx[0]
	}

	// 1) If there is an override Parse (e.g. provided by Default().Prefault()), always use it first so that
	//    custom behaviour defined in that wrapper takes precedence – particularly important for correctly
	//    distinguishing Default vs Prefault semantics on nil input.
	if z.internals != nil && z.internals.Parse != nil {
		payload := &core.ParsePayload{
			Value:  input,
			Issues: make([]core.ZodRawIssue, 0),
			Path:   make([]any, 0),
		}
		result := z.internals.Parse(payload, contextToUse)
		if len(result.Issues) > 0 {
			return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(result.Issues, contextToUse))
		}
		return result.Value, nil
	}

	// 2) Handle nil input specially: attempt normal validation first (important for Optional schemas).
	//    Only if validation fails should we apply the Prefault value.
	if input == nil {
		if result, err := z.innerType.Parse(input, contextToUse); err == nil {
			return result, nil
		} else {
			// For stringbool schema, nil should not trigger prefault fallback – propagate the error.
			if innerInternals := z.innerType.GetInternals(); innerInternals != nil && innerInternals.Type == "stringbool" {
				return nil, err
			}

			var fallbackValue any
			if z.isFunction && z.prefaultFunc != nil {
				fallbackValue = z.prefaultFunc()
			} else {
				fallbackValue = z.prefaultValue
			}

			// Special-case: stringbool's fallback value is bool and should be returned without re-validation
			if innerInternals := z.innerType.GetInternals(); innerInternals != nil && innerInternals.Type == "stringbool" {
				if _, ok := fallbackValue.(bool); ok {
					return fallbackValue, nil
				}
			}

			return z.innerType.Parse(fallbackValue, contextToUse)
		}
	}

	// 3) Non-nil input: try normal validation first; on failure fall back.
	result, err := z.innerType.Parse(input, contextToUse)
	if err == nil {
		return result, nil
	}

	var fallbackValue any
	if z.isFunction && z.prefaultFunc != nil {
		fallbackValue = z.prefaultFunc()
	} else {
		fallbackValue = z.prefaultValue
	}

	if innerInternals := z.innerType.GetInternals(); innerInternals != nil && innerInternals.Type == "stringbool" {
		if _, ok := fallbackValue.(bool); ok {
			return fallbackValue, nil
		}
	}

	fallbackResult, fallbackErr := z.innerType.Parse(fallbackValue, contextToUse)
	if fallbackErr != nil {
		return nil, fallbackErr
	}
	return fallbackResult, nil
}

// MustParse execute validation, and panic if failed
func (z *ZodPrefault[T]) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

////////////////////////////
////   PREFAULT CHAINS  ////
////////////////////////////

// RefineAny adds generic validation with fallback behavior
func (z *ZodPrefault[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	newInner := z.innerType.RefineAny(fn, params...)
	// need to cast ZodType[any, any] back to T
	if castedInner, ok := newInner.(T); ok {
		return any(prefaultInternal(castedInner, z.prefaultValue, z.prefaultFunc, z.isFunction)).(core.ZodType[any, any])
	}
	// if cast failed, return the refine result of the inner type
	return newInner
}

// Refine adds a flexible validation function to the prefault schema, return ZodPrefault
func (z *ZodPrefault[T]) Refine(fn func(string) bool, params ...any) *ZodPrefault[T] {
	if refineMethod, ok := any(z.innerType).(interface {
		Refine(func(string) bool, ...any) T
	}); ok {
		newInner := refineMethod.Refine(fn, params...)
		return prefaultInternal(newInner, z.prefaultValue, z.prefaultFunc, z.isFunction)
	}
	return z
}

// Nilable makes the prefault wrapper nilable
func (z *ZodPrefault[T]) Nilable() core.ZodType[any, any] {
	nilableInner := z.innerType.Nilable()
	if castedInner, ok := nilableInner.(T); ok {
		return any(prefaultInternal(castedInner, z.prefaultValue, z.prefaultFunc, z.isFunction)).(core.ZodType[any, any])
	}
	return nilableInner
}

// Optional makes the prefault wrapper optional
func (z *ZodPrefault[T]) Optional() core.ZodType[any, any] {
	// apply to the inner type
	if optMethod, ok := any(z.innerType).(interface{ Optional() core.ZodType[any, any] }); ok {
		optInner := optMethod.Optional()
		if castedInner, ok := optInner.(T); ok {
			return any(prefaultInternal(castedInner, z.prefaultValue, z.prefaultFunc, z.isFunction)).(core.ZodType[any, any])
		}
		return optInner
	}
	return z
}

// TransformAny applies data transformation with fallback behavior
func (z *ZodPrefault[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// create a wrapped transformation function, first handle prefault logic, then transform
	wrappedFn := func(input any, ctx *core.RefinementContext) (any, error) {
		// first apply prefault logic (same as Parse method)
		result, err := z.innerType.Parse(input)
		var processedInput any
		if err != nil {
			// validation failed, use fallback value
			if z.isFunction && z.prefaultFunc != nil {
				processedInput = z.prefaultFunc()
			} else {
				processedInput = z.prefaultValue
			}
			// validate fallback value
			result, err = z.innerType.Parse(processedInput)
			if err != nil {
				return nil, err
			}
		}
		processedInput = result

		// then apply the transformation function
		return fn(processedInput, ctx)
	}

	// create a new Transform, bypass the inner type and use the processed input directly
	transform := Transform[any, any](wrappedFn)
	return transform
}

// Pipe pipeline composition
func (z *ZodPrefault[T]) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return z.innerType.Pipe(out)
}

// Prefault fallback value when validation fails - create new Prefault wrapper
func (z *ZodPrefault[T]) Prefault(value any) core.ZodType[any, any] {
	return Prefault(any(z).(core.ZodType[any, any]), value)
}

// PrefaultFunc fallback value based on function - create new Prefault wrapper
func (z *ZodPrefault[T]) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	return PrefaultFunc(any(z).(core.ZodType[any, any]), fn)
}

////////////////////////////
////   PREFAULT PACKAGE FUNCTIONS ////
////////////////////////////

// prefaultInternal creates a Prefault wrapper with a static fallback value. It mirrors
// the Default wrapper API: only two arguments, the schema and the fallback
// value.
func prefaultInternal[T core.ZodType[any, any]](innerType T, value any, fn func() any, isFunc bool) *ZodPrefault[T] {
	// construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
	baseInternals := innerType.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:  core.Version,
		Type:     core.ZodTypePrefault,
		Checks:   baseInternals.Checks,
		Optional: baseInternals.Optional,
		Nilable:  baseInternals.Nilable,
		// Prefault doesn't change input optionality, only provides fallback on validation failure
		OptIn:       baseInternals.OptIn,  // Preserve input optionality
		OptOut:      baseInternals.OptOut, // Preserve output optionality
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}
	return &ZodPrefault[T]{
		internals:     internals,
		innerType:     innerType,
		prefaultValue: value,
		prefaultFunc:  fn,
		isFunction:    isFunc,
	}
}

// Prefault creates a Prefault wrapper with a static fallback value. It mirrors
// the Default wrapper API: only two arguments, the schema and the fallback
// value.
func Prefault[T core.ZodType[any, any]](innerType T, fallback any) core.ZodType[any, any] {
	return any(prefaultInternal(innerType, fallback, nil, false)).(core.ZodType[any, any])
}

// PrefaultFunc creates a Prefault wrapper where the fallback value is provided
// by a lazy function.
func PrefaultFunc[T core.ZodType[any, any]](innerType T, fn func() any) core.ZodType[any, any] {
	return any(prefaultInternal(innerType, nil, fn, true)).(core.ZodType[any, any])
}

// Unwrap returns the inner type (for wrapper types, returns the wrapped type)
func (z *ZodPrefault[T]) Unwrap() core.ZodType[any, any] {
	return any(z.innerType).(core.ZodType[any, any])
}
