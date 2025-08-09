package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/engine"
)

// ZodTransformDef represents the definition of a transformation schema.
type ZodTransformDef struct {
	core.ZodTypeDef
	transform   func(arg any, ctx *core.RefinementContext) (any, error)
	innerSchema core.ZodSchema
}

func (d *ZodTransformDef) Transform() func(any, *core.RefinementContext) (any, error) {
	return d.transform
}

// ZodTransform represents a schema with a custom transformation.
type ZodTransform[In any, Out any] struct {
	internals *ZodTransformInternals
}

// ZodTransformInternals contains internal state for ZodTransform.
type ZodTransformInternals struct {
	core.ZodTypeInternals
	Def *ZodTransformDef
}

// GetInternals returns the internal state of the ZodTransform schema.
func (z *ZodTransform[In, Out]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse applies the transformation to the input value.
func (z *ZodTransform[In, Out]) Parse(input any) (any, error) {
	// The actual transformation logic is handled by the engine,
	// which will call the transform function from internals.
	// This Parse method is here to satisfy the ZodType interface.
	return z.internals.Def.transform(input, nil)
}

// StrictParse validates the input using strict parsing rules and returns the transformed result
func (z *ZodTransform[In, Out]) StrictParse(input Out, ctx ...*core.ParseContext) (Out, error) {
	// For transform types, we need to work backwards - the input should already be the transformed type
	// We can't easily reverse the transformation, so we validate the input as the output type
	// and apply any additional checks from the internals
	if len(ctx) == 0 {
		ctx = []*core.ParseContext{core.NewParseContext()}
	}

	// Apply any checks defined on the transform schema itself
	validatedValue, err := engine.ApplyChecks[Out](input, z.internals.Checks, ctx[0])
	if err != nil {
		var zero Out
		return zero, err
	}

	return validatedValue, nil
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodTransform[In, Out]) MustStrictParse(input Out, ctx ...*core.ParseContext) Out {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// GetInner returns the input schema for the transformation.
func (z *ZodTransform[In, Out]) GetInner() core.ZodSchema {
	return z.internals.Def.innerSchema
}

// Type returns the type of the schema
func (z *ZodTransform[In, Out]) Type() core.ZodTypeCode {
	return core.ZodTypeTransform
}
