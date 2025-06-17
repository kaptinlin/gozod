package gozod

// ZodTransformDef defines the configuration for type conversion
type ZodTransformDef[Out, In any] struct {
	ZodTypeDef
	Type        string
	TransformFn func(any, *RefinementContext) (any, error)
}

// ZodTransformInternals contains the internal state of the Transform validator
type ZodTransformInternals[Out, In any] struct {
	ZodTypeInternals
	Def *ZodTransformDef[Out, In] // Transform definition
}

// ZodTransform implements type conversion
type ZodTransform[Out, In any] struct {
	internals *ZodTransformInternals[Out, In]
}

///////////////////////////
////   CORE METHODS     ////
///////////////////////////

// GetInternals returns the internal state
func (z *ZodTransform[Out, In]) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse executes conversion and validation
func (z *ZodTransform[Out, In]) Parse(input any, ctx ...*ParseContext) (any, error) {
	var parseCtx *ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	if parseCtx == nil {
		parseCtx = NewParseContext()
	}

	// Create parse payload
	payload := &ParsePayload{
		Value:  input,
		Issues: make([]ZodRawIssue, 0),
		Path:   []interface{}{},
	}

	// Create enhanced context, support error reporting
	refinementCtx := &RefinementContext{
		ParseContext: parseCtx,
		Value:        input,
		AddIssue: func(issue ZodIssue) {
			// Convert ZodIssue to ZodRawIssue and add to payload
			rawIssue := ZodRawIssue{
				Code:    issue.Code,
				Message: issue.Message,
				Path:    payload.Path,
			}
			payload.Issues = append(payload.Issues, rawIssue)
		},
	}

	// Transform core
	transformed, err := z.internals.Def.TransformFn(input, refinementCtx)
	if err != nil {
		// Conversion failed: add error, do not modify data
		issue := CreateCustomIssue(input, err.Error())
		payload.Issues = append(payload.Issues, issue)
	} else {
		// Conversion successful: directly modify payload.Value
		payload.Value = transformed
	}

	// Check if there are errors reported through AddIssue
	if len(payload.Issues) > 0 {
		finalIssues := make([]ZodIssue, len(payload.Issues))
		for i, rawIssue := range payload.Issues {
			finalIssues[i] = FinalizeIssue(rawIssue, parseCtx, GetConfig())
		}
		return nil, NewZodError(finalIssues)
	}

	return payload.Value, nil
}

// MustParse executes conversion, panics on failure
func (z *ZodTransform[Out, In]) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

////////////////////////////
////   INTERFACE METHODS ////
////////////////////////////

// Nilable modifier: Transform's Nilable behavior, maintaining smart type inference
func (z *ZodTransform[Out, In]) Nilable() ZodType[any, any] {
	// Transform's Nilable behavior: create a new Transform supporting nil
	newDef := &ZodTransformDef[any, any]{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		TransformFn: z.internals.Def.TransformFn,
	}

	return createZodTransformFromDef(newDef)
}

// Refine adds custom validation
func (z *ZodTransform[Out, In]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	// Transform does not directly support Refine, needs to be wrapped in Pipe
	var schemaParams SchemaParams
	if len(params) > 0 {
		schemaParams = params[0]
	}
	custom := NewZodCustom(RefineFn[interface{}](fn), schemaParams)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(custom).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Check adds check API
func (z *ZodTransform[Out, In]) Check(fn func(*ParsePayload) error) ZodType[any, any] {
	checkFn := func(payload *ParsePayload) {
		if err := fn(payload); err != nil {
			issue := NewRawIssue("custom", payload.Value, WithOrigin("transform"))
			issue.Message = err.Error()
			payload.Issues = append(payload.Issues, issue)
		}
	}

	custom := NewZodCustom(CheckFn(checkFn), SchemaParams{})

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(custom).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Transform chain conversion
func (z *ZodTransform[Out, In]) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	newTransform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(newTransform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// TransformAny flexible version of conversion
func (z *ZodTransform[Out, In]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	newTransform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(newTransform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe pipe combination
func (z *ZodTransform[Out, In]) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

////////////////////////////
////   CONSTRUCTOR      ////
////////////////////////////

// Transform global function, create independent converter
func Transform(transformFn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](transformFn)
	return any(transform).(ZodType[any, any])
}

// createZodTransformFromDef create ZodTransform from definition
func createZodTransformFromDef[Out, In any](def *ZodTransformDef[Out, In]) *ZodTransform[Out, In] {
	internals := &ZodTransformInternals[Out, In]{
		ZodTypeInternals: newBaseZodTypeInternals("transform"),
		Def:              def,
	}

	// Set parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		// Create enhanced context, support error reporting
		refinementCtx := &RefinementContext{
			ParseContext: ctx,
			Value:        payload.Value,
			AddIssue: func(issue ZodIssue) {
				rawIssue := ZodRawIssue{
					Code:    issue.Code,
					Message: issue.Message,
					Path:    payload.Path,
				}
				payload.Issues = append(payload.Issues, rawIssue)
			},
		}

		// 🔥 Transform core principle: can modify data (fundamental difference from Refine)
		// Execute conversion function
		result, err := def.TransformFn(payload.Value, refinementCtx)
		if err != nil {
			// Conversion failed: add error, keep original value
			issue := CreateCustomIssue(payload.Value, err.Error())
			payload.Issues = append(payload.Issues, issue)
		} else {
			// Conversion successful: directly modify payload.Value
			payload.Value = result
		}
		return payload
	}

	return &ZodTransform[Out, In]{internals: internals}
}

// NewZodTransform create new ZodTransform
func NewZodTransform[Out, In any](transformFn func(any, *RefinementContext) (any, error)) *ZodTransform[Out, In] {
	def := &ZodTransformDef[Out, In]{
		ZodTypeDef: ZodTypeDef{
			Type:   "transform",
			Checks: make([]ZodCheck, 0),
		},
		Type:        "transform",
		TransformFn: transformFn,
	}

	return createZodTransformFromDef(def)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodTransform[In, Out]) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}
