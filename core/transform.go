package core

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrInvalidTransformType is a sentinel error returned when a type assertion
// fails during transformation.
var ErrInvalidTransformType = errors.New("invalid type for transform")

func isNilInput(input any) bool {
	if input == nil {
		return true
	}
	v := reflect.ValueOf(input)
	if !v.IsValid() {
		return true
	}
	//nolint:exhaustive // Only these kinds can be nil
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

// ZodTransform represents a schema that applies a function to a validated value.
//
// Generic Parameters:
//   - In: The input type that the source schema validates.
//   - Out: The output type that the transformation function produces.
type ZodTransform[In, Out any] struct {
	source    ZodType[In]
	transform func(In, *RefinementContext) (Out, error)
	internals *ZodTypeInternals
}

// Parse validates input with the source schema, then applies the transformation.
// Default/DefaultFunc values skip transformation completely (Zod v4 semantics).
func (z *ZodTransform[In, Out]) Parse(input any, ctx ...*ParseContext) (Out, error) {
	sourceInternals := z.source.Internals()
	isDefaultShortCircuit := isNilInput(input) &&
		(sourceInternals.DefaultValue != nil || sourceInternals.DefaultFunc != nil)

	if isDefaultShortCircuit {
		validated, err := z.source.Parse(input, ctx...)
		if err != nil {
			var zero Out
			return zero, err
		}
		if result, ok := any(validated).(Out); ok {
			return result, nil
		}
		var zero Out
		return zero, fmt.Errorf(
			"default value type %T is not compatible with output type %T: %w",
			validated, zero, ErrInvalidTransformType)
	}

	validated, err := z.source.Parse(input, ctx...)
	if err != nil {
		var zero Out
		return zero, err
	}

	parseCtx := getOrCreateContext(ctx...)
	refCtx := &RefinementContext{ParseContext: parseCtx}
	return z.transform(validated, refCtx)
}

// MustParse validates and transforms input, panicking on error.
func (z *ZodTransform[In, Out]) MustParse(input any, ctx ...*ParseContext) Out {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates and transforms input, returning an untyped result.
func (z *ZodTransform[In, Out]) ParseAny(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Internals returns the schema's internal state.
func (z *ZodTransform[In, Out]) Internals() *ZodTypeInternals {
	return z.internals
}

// IsOptional reports whether this schema accepts missing values.
func (z *ZodTransform[In, Out]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodTransform[In, Out]) IsNilable() bool {
	return z.internals.IsNilable()
}

// ZodPipe represents a pipeline that validates with a source schema
// then channels the result into a target schema via a target function.
//
// Generic Parameters:
//   - In: The input type for the source schema.
//   - Out: The output type from the target function.
type ZodPipe[In, Out any] struct {
	source    ZodType[In]
	target    ZodType[any]
	targetFn  func(In, *ParseContext) (Out, error)
	internals *ZodTypeInternals
}

// Parse validates input: source schema -> target function.
func (z *ZodPipe[In, Out]) Parse(input any, ctx ...*ParseContext) (Out, error) {
	intermediate, err := z.source.Parse(input, ctx...)
	if err != nil {
		var zero Out
		return zero, err
	}
	parseCtx := getOrCreateContext(ctx...)
	return z.targetFn(intermediate, parseCtx)
}

// MustParse validates through the pipeline, panicking on error.
func (z *ZodPipe[In, Out]) MustParse(input any, ctx ...*ParseContext) Out {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Internals returns the schema's internal state.
func (z *ZodPipe[In, Out]) Internals() *ZodTypeInternals {
	return z.internals
}

// IsOptional reports whether this schema accepts missing values.
func (z *ZodPipe[In, Out]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodPipe[In, Out]) IsNilable() bool {
	return z.internals.IsNilable()
}

// NewZodTransform creates a new transformation schema.
func NewZodTransform[In, Out any](source ZodType[In], wrapperFn func(In, *RefinementContext) (Out, error)) *ZodTransform[In, Out] {
	internals := source.Internals().Clone()
	internals.Type = ZodTypeTransform
	internals.SetTransform(func(data any, ctx *RefinementContext) (any, error) {
		if typedData, ok := data.(In); ok {
			return wrapperFn(typedData, ctx)
		}
		var zero In
		return nil, fmt.Errorf("expected %T: %w", zero, ErrInvalidTransformType)
	})
	return &ZodTransform[In, Out]{
		source:    source,
		transform: wrapperFn,
		internals: internals,
	}
}

// Inner returns the input schema for the transformation.
func (z *ZodTransform[In, Out]) Inner() ZodSchema {
	if schema, ok := z.source.(ZodSchema); ok {
		return schema
	}
	return nil
}

// NewZodPipe creates a new pipeline schema.
func NewZodPipe[In, Out any](source ZodType[In], target ZodType[any], targetFn func(In, *ParseContext) (Out, error)) *ZodPipe[In, Out] {
	internals := source.Internals().Clone()
	internals.Type = ZodTypePipe
	return &ZodPipe[In, Out]{
		source:    source,
		target:    target,
		targetFn:  targetFn,
		internals: internals,
	}
}

// Inner returns the input (source) schema.
func (z *ZodPipe[In, Out]) Inner() ZodSchema {
	if s, ok := z.source.(ZodSchema); ok {
		return s
	}
	return nil
}

// Output returns the output (target) schema.
func (z *ZodPipe[In, Out]) Output() ZodSchema {
	if t, ok := z.target.(ZodSchema); ok {
		return t
	}
	return nil
}

// Transform allows chaining transformations on ZodTransform.
func (z *ZodTransform[In, Middle]) Transform(fn func(Middle, *RefinementContext) (any, error)) *ZodTransform[Middle, any] {
	return &ZodTransform[Middle, any]{
		source:    z,
		transform: fn,
		internals: z.internals.Clone(),
	}
}

// Pipe allows chaining ZodTransform into ZodPipe.
func (z *ZodTransform[In, Middle]) Pipe(target ZodType[any]) *ZodPipe[Middle, any] {
	targetFn := func(input Middle, ctx *ParseContext) (any, error) {
		return target.Parse(input, ctx)
	}
	return NewZodPipe(z, target, targetFn)
}

// Transform allows chaining transformations on ZodPipe.
func (z *ZodPipe[In, Middle]) Transform(fn func(Middle, *RefinementContext) (any, error)) *ZodTransform[Middle, any] {
	return &ZodTransform[Middle, any]{
		source:    z,
		transform: fn,
		internals: z.internals.Clone(),
	}
}

// Pipe allows chaining ZodPipe into another ZodPipe.
func (z *ZodPipe[In, Middle]) Pipe(target ZodType[any]) *ZodPipe[Middle, any] {
	targetFn := func(input Middle, ctx *ParseContext) (any, error) {
		return target.Parse(input, ctx)
	}
	return NewZodPipe(z, target, targetFn)
}
