package core

// =============================================================================
// TRANSFORMATION & PIPELINE PRIMITIVES
// =============================================================================
//
// This file implements the primitives for `gozod`'s `.transform()` and `.pipe()`
// functionalities. These allow chaining validation with post-processing steps,
// such as converting a value from one type to another or passing a validated
// value into another schema.
//
// Key Primitives:
//
//	1. ZodTransform - The result of a `.transform()` call. It wraps a source
//	   schema and a transformation function.
//
//	2. ZodPipe - The result of a `.pipe()` call. It validates with a source
//	   schema and then channels the result into a target schema.
//
// Note: The `RefinementContext` used by transformations is defined in `parsing.go`.

// -----------------------------------------------------------------------------
// ZODTRANSFORM IMPLEMENTATION
// -----------------------------------------------------------------------------

// ZodTransform represents a schema that applies a function to a validated value.
// It is returned by the `.transform()` method on any schema.
//
// Generic Parameters:
//   - In: The input type that the source schema validates.
//   - Out: The output type that the transformation function produces.
type ZodTransform[In, Out any] struct {
	source    ZodType[In]                               // Source schema for input validation
	transform func(In, *RefinementContext) (Out, error) // The user-defined transformation function
	internals *ZodTypeInternals                         // Internal state and metadata
}

// Parse validates input with the source schema, then applies the transformation.
func (z *ZodTransform[In, Out]) Parse(input any, ctx ...*ParseContext) (Out, error) {
	// Step 1: Validate with source schema
	validated, err := z.source.Parse(input, ctx...)
	if err != nil {
		var zero Out
		return zero, err
	}

	// Step 2: Apply transformation function (created by wrapfn)
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

// GetInternals returns the internal state of this transformation schema.
func (z *ZodTransform[In, Out]) GetInternals() *ZodTypeInternals {
	return z.internals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodTransform[In, Out]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodTransform[In, Out]) IsNilable() bool {
	return z.internals.IsNilable()
}

// =============================================================================
// ZODPIPE - SIMPLIFIED PIPELINE WITH WRAPFN PATTERN
// =============================================================================

// ZodPipe represents a pipeline using the WrapFn pattern.
// Instead of holding a target schema object, it uses a target function created by wrapfn.
//
// Design Philosophy:
//   - Direct function composition eliminates adapter overhead
//   - Each type implements its own wrapfn logic for type conversion
//   - Cleaner architecture with better performance
//
// Generic Parameters:
//   - In: The input type for the source schema
//   - Out: The output type from the target function
//
// Key Innovation:
//   - targetFn: A function that handles type conversion and target validation
//   - Created by each type's Pipe method using wrapfn pattern
//   - No need for adapter objects or converters
type ZodPipe[In, Out any] struct {
	source    ZodType[In]                          // Source schema for input validation
	targetFn  func(In, *ParseContext) (Out, error) // Target function (from wrapfn)
	internals *ZodTypeInternals                    // Internal state and metadata
}

// Parse validates input through the wrapfn pipeline: source schema -> target function.
func (z *ZodPipe[In, Out]) Parse(input any, ctx ...*ParseContext) (Out, error) {
	// Step 1: Validate with source schema
	intermediate, err := z.source.Parse(input, ctx...)
	if err != nil {
		var zero Out
		return zero, err
	}

	// Step 2: Apply target function (handles type conversion + target validation)
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

// GetInternals returns the internal state of this pipeline schema.
func (z *ZodPipe[In, Out]) GetInternals() *ZodTypeInternals {
	return z.internals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodPipe[In, Out]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodPipe[In, Out]) IsNilable() bool {
	return z.internals.IsNilable()
}

// =============================================================================
// WRAPFN PATTERN IMPLEMENTATION HELPERS
// =============================================================================

// NewZodTransform creates a new transformation schema.
// This is used by each type's Transform method with their own wrapfn logic.
//
// Parameters:
//   - source: The source schema
//   - wrapperFn: The wrapper function created by the type's Transform method
//
// Returns:
//   - *ZodTransform[In, Out]: A new transformation schema
func NewZodTransform[In, Out any](source ZodType[In], wrapperFn func(In, *RefinementContext) (Out, error)) *ZodTransform[In, Out] {
	return &ZodTransform[In, Out]{
		source:    source,
		transform: wrapperFn,
		internals: source.GetInternals().Clone(),
	}
}

// NewZodPipe creates a new pipeline schema.
// This is used by each type's Pipe method with their own wrapfn logic.
//
// Parameters:
//   - source: The source schema
//   - targetFn: The target function created by the type's Pipe method
//
// Returns:
//   - *ZodPipe[In, Out]: A new pipeline schema
func NewZodPipe[In, Out any](source ZodType[In], targetFn func(In, *ParseContext) (Out, error)) *ZodPipe[In, Out] {
	return &ZodPipe[In, Out]{
		source:    source,
		targetFn:  targetFn,
		internals: source.GetInternals().Clone(),
	}
}

// =============================================================================
// CHAINING METHODS - FOR TRANSFORM AND PIPE COMPOSITIONS
// =============================================================================

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
	return NewZodPipe[Middle, any](z, targetFn)
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
	return NewZodPipe[Middle, any](z, targetFn)
}
