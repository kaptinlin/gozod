package gozod

///////////////////////////
////   PREFAULT TYPE   ////
///////////////////////////

// ZodPrefault represents a validation pattern with fallback values
// Core design: contains an inner type, and forwards all methods through it
type ZodPrefault[T ZodType[any, any]] struct {
	internals     *ZodTypeInternals // Prefault's own internals, Type = "prefault"
	innerType     T                 // inner type
	prefaultValue any               // fallback value
	prefaultFunc  func() any        // fallback function
	isFunction    bool              // whether to use function to provide fallback value
}

///////////////////////////
////   PREFAULT CORE   ////
///////////////////////////

// GetInternals returns the state of the inner type
func (z *ZodPrefault[T]) GetInternals() *ZodTypeInternals {
	return z.internals
}

// GetZod returns the prefault-specific internals (type-safe access)
func (z *ZodPrefault[T]) GetZod() *ZodTypeInternals {
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

// Coerce attempts to coerce input (with fallback behavior)
func (z *ZodPrefault[T]) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Parse with smart type inference
// try to validate, and use fallback value if failed
// unlike Default (use default value when nil), Prefault always tries to validate first
// important: for nil input, Prefault should reject and try fallback value
func (z *ZodPrefault[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	var contextToUse *ParseContext
	if len(ctx) > 0 {
		contextToUse = ctx[0]
	}

	// if there is a custom Parse function (e.g. from Default.Prefault), use it
	if z.internals != nil && z.internals.Parse != nil {
		payload := &ParsePayload{
			Value:  input,
			Issues: make([]ZodRawIssue, 0),
			Path:   make([]interface{}, 0),
		}
		result := z.internals.Parse(payload, contextToUse)
		if len(result.Issues) > 0 {
			return nil, &ZodError{Issues: convertRawIssuesToIssues(result.Issues, contextToUse)}
		}
		return result.Value, nil
	}

	// default Prefault logic
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
	fallbackResult, fallbackErr := z.innerType.Parse(fallbackValue, contextToUse)
	if fallbackErr != nil {
		return nil, fallbackErr
	}
	return fallbackResult, nil
}

// MustParse execute validation, and panic if failed
func (z *ZodPrefault[T]) MustParse(input any, ctx ...*ParseContext) any {
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
func (z *ZodPrefault[T]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	newInner := z.innerType.RefineAny(fn, params...)
	// need to cast ZodType[any, any] back to T
	if castedInner, ok := newInner.(T); ok {
		return newZodPrefault(castedInner, z.prefaultValue, z.prefaultFunc, z.isFunction)
	}
	// if cast failed, return the refine result of the inner type
	return newInner
}

// Refine adds a flexible validation function to the prefault schema, return ZodPrefault
func (z *ZodPrefault[T]) Refine(fn func(string) bool, params ...SchemaParams) *ZodPrefault[T] {
	if refineMethod, ok := any(z.innerType).(interface {
		Refine(func(string) bool, ...SchemaParams) T
	}); ok {
		newInner := refineMethod.Refine(fn, params...)
		return newZodPrefault(newInner, z.prefaultValue, z.prefaultFunc, z.isFunction)
	}
	return z
}

// Nilable makes the prefault wrapper nilable
func (z *ZodPrefault[T]) Nilable() ZodType[any, any] {
	nilableInner := z.innerType.Nilable()
	if castedInner, ok := nilableInner.(T); ok {
		return newZodPrefault(castedInner, z.prefaultValue, z.prefaultFunc, z.isFunction)
	}
	return nilableInner
}

// Optional makes the prefault wrapper optional
func (z *ZodPrefault[T]) Optional() ZodType[any, any] {
	// apply to the inner type
	if optMethod, ok := any(z.innerType).(interface{ Optional() ZodType[any, any] }); ok {
		optInner := optMethod.Optional()
		if castedInner, ok := optInner.(T); ok {
			return newZodPrefault(castedInner, z.prefaultValue, z.prefaultFunc, z.isFunction)
		}
		return optInner
	}
	return z
}

// TransformAny applies data transformation with fallback behavior
func (z *ZodPrefault[T]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// create a wrapped transformation function, first handle prefault logic, then transform
	wrappedFn := func(input any, ctx *RefinementContext) (any, error) {
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
	transform := NewZodTransform[any, any](wrappedFn)
	return transform
}

// Pipe pipeline composition
func (z *ZodPrefault[T]) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return z.innerType.Pipe(out)
}

// Prefault fallback value when validation fails - create new Prefault wrapper
func (z *ZodPrefault[T]) Prefault(value any) ZodType[any, any] {
	return newZodPrefault(any(z).(ZodType[any, any]), value, nil, false)
}

// PrefaultFunc fallback value based on function - create new Prefault wrapper
func (z *ZodPrefault[T]) PrefaultFunc(fn func() any) ZodType[any, any] {
	return newZodPrefault(any(z).(ZodType[any, any]), nil, fn, true)
}

////////////////////////////
////   PREFAULT PACKAGE FUNCTIONS ////
////////////////////////////

// Prefault create a pattern wrapper with fallback value
func Prefault[In, Out any](innerType ZodType[In, Out], prefaultValue any) ZodType[any, any] {
	return newZodPrefault(any(innerType).(ZodType[any, any]), prefaultValue, nil, false)
}

// PrefaultFunc create a pattern wrapper with function fallback value
func PrefaultFunc[In, Out any](innerType ZodType[In, Out], fn func() any) ZodType[any, any] {
	return newZodPrefault(any(innerType).(ZodType[any, any]), nil, fn, true)
}

////////////////////////////
////   CONSTRUCTOR      ////
////////////////////////////

func newZodPrefault[T ZodType[any, any]](innerType T, value any, fn func() any, isFunc bool) *ZodPrefault[T] {
	// construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
	baseInternals := innerType.GetInternals()
	internals := &ZodTypeInternals{
		Version:  baseInternals.Version,
		Type:     ZodTypePrefault,
		Checks:   baseInternals.Checks,
		Coerce:   baseInternals.Coerce,
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

// Unwrap returns the inner type (for wrapper types, returns the wrapped type)
func (z *ZodPrefault[T]) Unwrap() ZodType[any, any] {
	return any(z.innerType).(ZodType[any, any])
}
