package gozod

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodPipe implements pipeline composition
// 1. Sequential execution: input schema first, then output schema on success
// 2. Issue accumulation: issues array is passed and accumulated between schemas
// 3. Early termination: right side not executed if left side fails
// 4. Property inheritance: smart inheritance of properties from both schemas
type ZodPipe[In, Out any] struct {
	in  ZodType[In, any]  // Input schema
	out ZodType[any, Out] // Output schema
	def ZodTypeDef
}

// GetInternals returns the internal state (for framework use)
func (zp *ZodPipe[In, Out]) GetInternals() *ZodTypeInternals {
	inInternals := zp.in.GetInternals()
	outInternals := zp.out.GetInternals()

	return &ZodTypeInternals{
		Type:    "pipe",
		Version: Version,
		Checks:  make([]ZodCheck, 0),
		OptIn:   inInternals.OptIn,
		OptOut:  outInternals.OptOut,
		Parse: func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
			// 1. Execute input schema
			inInternals := zp.in.GetInternals()
			leftPayload := inInternals.Parse(payload, ctx)

			// Check if left side has errors
			if len(leftPayload.Issues) > 0 {
				return leftPayload // Early termination
			}

			// Special handling: if left result is nil (usually from Optional), return nil directly
			if leftPayload.Value == nil {
				return leftPayload
			}

			// 2. Execute output schema
			outInternals := zp.out.GetInternals()
			rightPayload := &ParsePayload{
				Value:  leftPayload.Value,
				Issues: leftPayload.Issues, // Maintain issue accumulation
			}

			return outInternals.Parse(rightPayload, ctx)
		},
	}
}

// GetZod returns the pipe-specific internals (type-safe access)
func (zp *ZodPipe[In, Out]) GetZod() *ZodTypeInternals {
	return zp.GetInternals()
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (zp *ZodPipe[In, Out]) CloneFrom(source any) {
	if src, ok := source.(*ZodPipe[In, Out]); ok {
		// Copy all pipe-specific state
		zp.in = src.in
		zp.out = src.out
		zp.def = src.def
	}
}

// Coerce attempts to coerce input (applies to pipeline behavior)
func (zp *ZodPipe[In, Out]) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return zp.Parse(input, ctx...)
}

// Parse executes pipeline validation with smart type inference
// - Pipe().Parse(value) → right.Parse(left.Parse(value))
// - Maintains two-stage smart type inference: input → intermediate → output
// Core: pipeline ensures type conversion integrity, maintaining smart inference chain
// Special handling: Optional's nil result should skip subsequent pipeline, return nil directly
func (zp *ZodPipe[In, Out]) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := &ParseContext{}
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// 1. Execute input schema, maintain smart type inference
	leftResult, err := zp.in.Parse(input, parseCtx)
	if err != nil {
		// Early termination: don't execute right side if left side fails
		return nil, err
	}

	// Special handling: if left result is nil (usually from Optional), return nil directly
	if leftResult == nil {
		return nil, nil
	}

	// 2. Use left result as right input, continue smart type inference
	rightResult, err := zp.out.Parse(leftResult, parseCtx)
	if err != nil {
		return nil, err
	}

	return rightResult, nil
}

// MustParse executes pipeline validation and panics on failure
func (zp *ZodPipe[In, Out]) MustParse(input any, ctx ...*ParseContext) any {
	result, err := zp.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Check adds modern check API
func (zp *ZodPipe[In, Out]) Check(fn CheckFn) ZodType[any, any] {
	// Check if output type supports Check method
	if checkMethod, ok := any(zp.out).(interface {
		Check(CheckFn) ZodType[any, any]
	}); ok {
		newOut := checkMethod.Check(fn)
		return &ZodPipe[any, any]{
			in:  any(zp.in).(ZodType[any, any]),
			out: any(newOut).(ZodType[any, any]),
			def: ZodTypeDef{Type: "pipe"},
		}
	}

	// If output type doesn't support Check, return self
	return any(zp).(ZodType[any, any])
}

// Refine adds type-safe custom validation logic to the pipe schema
// Core principle: validation success returns original value, validation failure returns error, never modifies input
func (zp *ZodPipe[In, Out]) Refine(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	// Use existing RefineAny infrastructure with direct inline handling
	return zp.RefineAny(fn, params...)
}

// RefineAny adds flexible custom validation logic to the pipe schema
func (zp *ZodPipe[In, Out]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	// Pipe's Refine should apply to final output
	newOut := zp.out.RefineAny(fn, params...)
	return &ZodPipe[any, any]{
		in:  any(zp.in).(ZodType[any, any]),
		out: any(newOut).(ZodType[any, any]),
		def: zp.def,
	}
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a new transform with another transformation (corresponds to z.pipe().transform())
func (zp *ZodPipe[In, Out]) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return zp.TransformAny(fn)
}

// TransformAny corresponds to z.pipe().transform() - same implementation as Transform method
func (zp *ZodPipe[In, Out]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// Create new Transform
	transform := NewZodTransform[any, any](fn)

	// Return new Pipe using type conversion to simplify implementation
	return &ZodPipe[any, any]{
		in:  any(zp).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe corresponds to z.pipe().pipe() - pipeline chaining
func (zp *ZodPipe[In, Out]) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(zp).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional modifier: apply to input schema, maintain smart type inference
func (zp *ZodPipe[In, Out]) Optional() ZodType[any, any] {
	// Create true Optional wrapper
	return any(Optional(any(zp).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable modifier: apply to input schema, maintain smart type inference
func (zp *ZodPipe[In, Out]) Nilable() ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  zp.in.Nilable(),
		out: any(zp.out).(ZodType[any, any]),
		def: zp.def,
	}
}

// Nullish makes the schema both optional and nilable
func (zp *ZodPipe[In, Out]) Nullish() ZodType[any, any] {
	return zp.Optional().Nilable()
}

// Unwrap returns the output schema (the final result of the pipe)
func (zp *ZodPipe[In, Out]) Unwrap() ZodType[any, any] {
	return any(zp.out).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// Prefault corresponds to z.pipe().catch() - fallback value on validation failure
func (zp *ZodPipe[In, Out]) Prefault(value any) ZodType[any, any] {
	// Pipe's Prefault should apply to output end
	// Need type assertion because ZodType[any, Out] interface might not have Prefault method
	if prefaultable, ok := any(zp.out).(interface{ Prefault(any) ZodType[any, any] }); ok {
		newOut := prefaultable.Prefault(value)
		return &ZodPipe[any, any]{
			in:  any(zp.in).(ZodType[any, any]),
			out: newOut,
			def: zp.def,
		}
	}
	// If output type doesn't support Prefault, return original pipe
	return &ZodPipe[any, any]{
		in:  any(zp.in).(ZodType[any, any]),
		out: any(zp.out).(ZodType[any, any]),
		def: zp.def,
	}
}

// PrefaultFunc corresponds to z.pipe().catch(() => value)
func (zp *ZodPipe[In, Out]) PrefaultFunc(fn func() any) ZodType[any, any] {
	// Need type assertion because ZodType[any, Out] interface might not have PrefaultFunc method
	if prefaultable, ok := any(zp.out).(interface {
		PrefaultFunc(func() any) ZodType[any, any]
	}); ok {
		newOut := prefaultable.PrefaultFunc(fn)
		return &ZodPipe[any, any]{
			in:  any(zp.in).(ZodType[any, any]),
			out: newOut,
			def: zp.def,
		}
	}
	// If output type doesn't support PrefaultFunc, return original pipe
	return &ZodPipe[any, any]{
		in:  any(zp.in).(ZodType[any, any]),
		out: any(zp.out).(ZodType[any, any]),
		def: zp.def,
	}
}

// Default adds default value wrapper (applies to input end)
func (zp *ZodPipe[In, Out]) Default(value any) ZodType[any, any] {
	// Check if input type supports Default method
	if defaultMethod, ok := any(zp.in).(interface {
		Default(any) ZodType[any, any]
	}); ok {
		newIn := defaultMethod.Default(value)
		return &ZodPipe[any, any]{
			in:  newIn,
			out: any(zp.out).(ZodType[any, any]),
			def: zp.def,
		}
	}
	// If input type doesn't support Default, return original pipe
	return &ZodPipe[any, any]{
		in:  any(zp.in).(ZodType[any, any]),
		out: any(zp.out).(ZodType[any, any]),
		def: zp.def,
	}
}

// DefaultFunc adds function-based default value wrapper (applies to input end)
func (zp *ZodPipe[In, Out]) DefaultFunc(fn func() any) ZodType[any, any] {
	// Check if input type supports DefaultFunc method
	if defaultFuncMethod, ok := any(zp.in).(interface {
		DefaultFunc(func() any) ZodType[any, any]
	}); ok {
		newIn := defaultFuncMethod.DefaultFunc(fn)
		return &ZodPipe[any, any]{
			in:  newIn,
			out: any(zp.out).(ZodType[any, any]),
			def: zp.def,
		}
	}
	// If input type doesn't support DefaultFunc, return original pipe
	return &ZodPipe[any, any]{
		in:  any(zp.in).(ZodType[any, any]),
		out: any(zp.out).(ZodType[any, any]),
		def: zp.def,
	}
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// Pipe global function, creates pipeline composition
func Pipe[In, Out any](from ZodType[In, any], to ZodType[any, Out]) ZodType[any, any] {
	pipe := &ZodPipe[In, Out]{
		in:  from,
		out: to,
		def: ZodTypeDef{Type: "pipe"},
	}
	return any(pipe).(ZodType[any, any])
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// GetIn exposes internal schema for access
func (zp *ZodPipe[In, Out]) GetIn() ZodType[In, any] {
	return zp.in
}

// GetOut exposes internal schema for access
func (zp *ZodPipe[In, Out]) GetOut() ZodType[any, Out] {
	return zp.out
}
