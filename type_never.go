package gozod

import (
	"errors"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodNeverDef defines the configuration for never validation that always fails
type ZodNeverDef struct {
	ZodTypeDef
	Type string // "never"
}

// ZodNeverInternals contains never validator internal state
type ZodNeverInternals struct {
	ZodTypeInternals
	Def  *ZodNeverDef           // Schema definition
	Isst ZodIssueInvalidType    // Invalid type issue template
	Bag  map[string]interface{} // Additional metadata
}

// ZodNever represents a never validation schema that always fails validation
type ZodNever struct {
	internals *ZodNeverInternals
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Check adds a validation check (though never will always fail first)
func (z *ZodNever) Check(fn CheckFn) ZodType[any, any] {
	check := NewCustom[any](func(v any) bool {
		payload := &ParsePayload{Value: v, Issues: make([]ZodRawIssue, 0)}
		fn(payload)
		return len(payload.Issues) == 0
	}, SchemaParams{})
	return AddCheck(any(z).(ZodType[any, any]), check)
}
func (z *ZodNever) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the never-specific internals for framework usage
func (z *ZodNever) GetZod() *ZodNeverInternals {
	return z.internals
}

// CloneFrom implements the Cloneable interface
func (z *ZodNever) CloneFrom(source any) {
	if src, ok := source.(*ZodNever); ok {
		if src.internals.Bag != nil {
			z.internals.Bag = make(map[string]interface{})
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// Coerce attempts to coerce input (always fails for never type)
func (z *ZodNever) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Parse validates input (always fails except for Nilable never with nil input)
func (z *ZodNever) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Handle Nilable Never type: only nil can pass
	if z.internals.Nilable && input == nil {
		return (*interface{})(nil), nil // Return typed nil pointer for Never
	}

	// Never type's core semantics: fail for any input (including non-Nilable nil)
	rawIssue := CreateInvalidTypeIssue(input, "never", string(GetParsedType(input)))
	rawIssue.Message = "never type should never receive any value"
	finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
	return nil, NewZodError([]ZodIssue{finalIssue})
}

// MustParse validates the input value and panics on failure
func (z *ZodNever) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform adds data transformation (though never will always fail first)
func (z *ZodNever) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a new transform with another transformation
func (z *ZodNever) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodNever) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the never optional
func (z *ZodNever) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable makes the never nilable
func (z *ZodNever) Nilable() ZodType[any, any] {
	return Nilable(any(z).(ZodType[any, any]))
}

// Nullish makes the never both optional and nilable
func (z *ZodNever) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Refine adds validation (though never will always fail first)
func (z *ZodNever) Refine(fn func(any) bool, params ...SchemaParams) *ZodNever {
	result := z.RefineAny(fn, params...)
	return result.(*ZodNever)
}

// RefineAny provides flexible validation with any type
func (z *ZodNever) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodNever) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodNeverFromDef creates a ZodNever from definition
func createZodNeverFromDef(def *ZodNeverDef) *ZodNever {
	internals := &ZodNeverInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             ZodIssueInvalidType{Expected: "never"},
		Bag:              make(map[string]interface{}),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		neverDef := &ZodNeverDef{
			ZodTypeDef: *newDef,
			Type:       ZodTypeNever,
		}
		return createZodNeverFromDef(neverDef)
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodNever{internals: internals}
		result, err := schema.Parse(payload.Value, ctx)
		if err != nil {
			var zodErr *ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}
		payload.Value = result
		return payload
	}

	schema := &ZodNever{internals: internals}

	// Initialize the schema with proper error handling support
	initZodType(schema, &def.ZodTypeDef)

	return schema
}

// NewZodNever creates a new never schema that always fails validation
func NewZodNever(params ...SchemaParams) *ZodNever {
	def := &ZodNeverDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeNever,
			Checks: make([]ZodCheck, 0),
		},
		Type: ZodTypeNever,
	}

	schema := createZodNeverFromDef(def)

	// Apply schema parameters using modern error handling
	if len(params) > 0 {
		param := params[0]

		// Handle schema-level error configuration using utility function
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}

		// Store additional parameters in bag
		if param.Params != nil {
			for key, value := range param.Params {
				schema.internals.Bag[key] = value
			}
		}
	}

	return schema
}

// Never creates a new never schema (public constructor)
func Never(params ...SchemaParams) *ZodNever {
	return NewZodNever(params...)
}
