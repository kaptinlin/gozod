package types

import "github.com/kaptinlin/gozod/core"

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

// GetInner returns the input schema for the transformation.
func (z *ZodTransform[In, Out]) GetInner() core.ZodSchema {
	return z.internals.Def.innerSchema
}

// Type returns the type of the schema
func (z *ZodTransform[In, Out]) Type() core.ZodTypeCode {
	return core.ZodTypeTransform
}
