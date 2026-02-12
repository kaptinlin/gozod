package core

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrInvalidTransformType is returned when a type assertion fails during
// transformation.
var ErrInvalidTransformType = errors.New("invalid type for transform")

// isNilInput reports whether input is nil, including typed nil values
// for pointer, interface, slice, map, chan, and func kinds.
func isNilInput(input any) bool {
	if input == nil {
		return true
	}
	v := reflect.ValueOf(input)
	if !v.IsValid() {
		return true
	}
	//nolint:exhaustive // Only these kinds can be nil.
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

// ZodTransform applies a transformation function to a validated value,
// potentially changing its type.
//
// In is the input type validated by the source schema.
// Out is the output type produced by the transformation function.
type ZodTransform[In, Out any] struct {
	source    ZodType[In]
	transform func(In, *RefinementContext) (Out, error)
	internals *ZodTypeInternals
}

// Parse validates input with the source schema, then applies the
// transformation. When input is nil and the source has a default or
// default function, the default value is returned directly without
// running the transformation (Zod v4 semantics).
func (t *ZodTransform[In, Out]) Parse(input any, ctx ...*ParseContext) (out Out, _ error) {
	si := t.source.Internals()
	hasDefault := isNilInput(input) && (si.DefaultValue != nil || si.DefaultFunc != nil)

	if hasDefault {
		validated, err := t.source.Parse(input, ctx...)
		if err != nil {
			return out, err
		}
		result, ok := any(validated).(Out)
		if !ok {
			return out, fmt.Errorf(
				"default value type %T is not compatible with output type %T: %w",
				validated, out, ErrInvalidTransformType)
		}
		return result, nil
	}

	validated, err := t.source.Parse(input, ctx...)
	if err != nil {
		return out, err
	}
	return t.transform(validated, &RefinementContext{
		ParseContext: getOrCreateContext(ctx...),
	})
}

// MustParse validates and transforms input, panicking on error.
func (t *ZodTransform[In, Out]) MustParse(input any, ctx ...*ParseContext) Out {
	result, err := t.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates and transforms input, returning an untyped result.
func (t *ZodTransform[In, Out]) ParseAny(input any, ctx ...*ParseContext) (any, error) {
	return t.Parse(input, ctx...)
}

// Internals returns the schema's internal configuration.
func (t *ZodTransform[In, Out]) Internals() *ZodTypeInternals {
	return t.internals
}

// IsOptional reports whether this schema accepts missing values.
func (t *ZodTransform[In, Out]) IsOptional() bool {
	return t.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (t *ZodTransform[In, Out]) IsNilable() bool {
	return t.internals.IsNilable()
}

// Inner returns the source schema that feeds into this transformation.
func (t *ZodTransform[In, Out]) Inner() ZodSchema {
	s, _ := t.source.(ZodSchema)
	return s
}

// Transform chains an additional transformation onto this one.
func (t *ZodTransform[In, Middle]) Transform(fn func(Middle, *RefinementContext) (any, error)) *ZodTransform[Middle, any] {
	return &ZodTransform[Middle, any]{
		source:    t,
		transform: fn,
		internals: t.internals.Clone(),
	}
}

// Pipe chains this transformation into a target schema for further validation.
func (t *ZodTransform[In, Middle]) Pipe(target ZodType[any]) *ZodPipe[Middle, any] {
	fn := func(input Middle, pc *ParseContext) (any, error) {
		return target.Parse(input, pc)
	}
	return NewZodPipe(t, target, fn)
}

// ZodPipe validates input with a source schema, then passes the result
// through a target function for further processing.
//
// In is the input type for the source schema.
// Out is the output type from the target function.
type ZodPipe[In, Out any] struct {
	source    ZodType[In]
	target    ZodType[any]
	targetFn  func(In, *ParseContext) (Out, error)
	internals *ZodTypeInternals
}

// Parse validates input through the source schema, then applies the
// target function.
func (p *ZodPipe[In, Out]) Parse(input any, ctx ...*ParseContext) (out Out, _ error) {
	intermediate, err := p.source.Parse(input, ctx...)
	if err != nil {
		return out, err
	}
	return p.targetFn(intermediate, getOrCreateContext(ctx...))
}

// MustParse validates through the pipeline, panicking on error.
func (p *ZodPipe[In, Out]) MustParse(input any, ctx ...*ParseContext) Out {
	result, err := p.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Internals returns the schema's internal configuration.
func (p *ZodPipe[In, Out]) Internals() *ZodTypeInternals {
	return p.internals
}

// IsOptional reports whether this schema accepts missing values.
func (p *ZodPipe[In, Out]) IsOptional() bool {
	return p.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (p *ZodPipe[In, Out]) IsNilable() bool {
	return p.internals.IsNilable()
}

// Inner returns the source schema of this pipeline.
func (p *ZodPipe[In, Out]) Inner() ZodSchema {
	s, _ := p.source.(ZodSchema)
	return s
}

// Output returns the target schema of this pipeline.
func (p *ZodPipe[In, Out]) Output() ZodSchema {
	s, _ := p.target.(ZodSchema)
	return s
}

// Transform chains a transformation onto this pipeline's output.
func (p *ZodPipe[In, Middle]) Transform(fn func(Middle, *RefinementContext) (any, error)) *ZodTransform[Middle, any] {
	return &ZodTransform[Middle, any]{
		source:    p,
		transform: fn,
		internals: p.internals.Clone(),
	}
}

// Pipe chains this pipeline into another target schema.
func (p *ZodPipe[In, Middle]) Pipe(target ZodType[any]) *ZodPipe[Middle, any] {
	fn := func(input Middle, pc *ParseContext) (any, error) {
		return target.Parse(input, pc)
	}
	return NewZodPipe(p, target, fn)
}

// NewZodTransform creates a transformation schema that validates input with
// source, then applies fn to produce a potentially different output type.
func NewZodTransform[In, Out any](source ZodType[In], fn func(In, *RefinementContext) (Out, error)) *ZodTransform[In, Out] {
	internals := source.Internals().Clone()
	internals.Type = ZodTypeTransform
	internals.SetTransform(func(data any, ctx *RefinementContext) (any, error) {
		typed, ok := data.(In)
		if !ok {
			var zero In
			return nil, fmt.Errorf("expected %T: %w", zero, ErrInvalidTransformType)
		}
		return fn(typed, ctx)
	})
	return &ZodTransform[In, Out]{
		source:    source,
		transform: fn,
		internals: internals,
	}
}

// NewZodPipe creates a pipeline schema that validates input with source,
// then passes the result through fn for further processing.
func NewZodPipe[In, Out any](source ZodType[In], target ZodType[any], fn func(In, *ParseContext) (Out, error)) *ZodPipe[In, Out] {
	internals := source.Internals().Clone()
	internals.Type = ZodTypePipe
	return &ZodPipe[In, Out]{
		source:    source,
		target:    target,
		targetFn:  fn,
		internals: internals,
	}
}
