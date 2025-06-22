package types

import (
	"errors"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodNeverDef defines the configuration for never validation that always fails
type ZodNeverDef struct {
	core.ZodTypeDef
	Type string // "never"
}

// ZodNeverInternals contains never validator internal state
type ZodNeverInternals struct {
	core.ZodTypeInternals
	Def  *ZodNeverDef               // Schema definition
	Isst issues.ZodIssueInvalidType // Invalid type issue template
	Bag  map[string]any             // Additional metadata
}

// ZodNever represents a never validation schema that always fails validation
type ZodNever struct {
	internals *ZodNeverInternals
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Check adds a validation check (though never will always fail first)
func (z *ZodNever) Check(fn core.CheckFn) core.ZodType[any, any] {
	check := checks.NewCustom[any](func(v any) bool {
		payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
		fn(payload)
		return len(payload.Issues) == 0
	}, core.SchemaParams{})
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}
func (z *ZodNever) GetInternals() *core.ZodTypeInternals {
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
			z.internals.Bag = make(map[string]any)
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// Parse validates input (always fails except for Nilable never with nil input)
func (z *ZodNever) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Handle Nilable Never type: only nil can pass
	// Use reflectx.IsNil for more robust nil checking
	if z.internals.Nilable && reflectx.IsNil(input) {
		return (*any)(nil), nil // Return typed nil pointer for Never
	}

	// Never type's core semantics: fail for any input (including non-Nilable nil)
	// Fix the parameter order: expected first, then input
	rawIssue := issues.CreateInvalidTypeIssue("never", input)
	rawIssue.Message = "never type should never receive any value"
	finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
	return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// MustParse validates the input value and panics on failure
func (z *ZodNever) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodNever) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a new transform with another transformation
func (z *ZodNever) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodNever) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the never optional
func (z *ZodNever) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the never nilable
func (z *ZodNever) Nilable() core.ZodType[any, any] {
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Nullish makes the never both optional and nilable
func (z *ZodNever) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Refine adds validation (though never will always fail first)
func (z *ZodNever) Refine(fn func(any) bool, params ...any) *ZodNever {
	result := z.RefineAny(fn, params...)
	return result.(*ZodNever)
}

// RefineAny provides flexible validation with any type
func (z *ZodNever) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodNever) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodNeverFromDef creates a ZodNever from definition
func createZodNeverFromDef(def *ZodNeverDef) *ZodNever {
	internals := &ZodNeverInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: "never"},
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		neverDef := &ZodNeverDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeNever,
		}
		return createZodNeverFromDef(neverDef)
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		schema := &ZodNever{internals: internals}
		result, err := schema.Parse(payload.Value, ctx)
		if err != nil {
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := core.ZodRawIssue{
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

	zodSchema := &ZodNever{internals: internals}

	// Initialize the schema with proper error handling support
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

// Never creates a new never schema that always fails validation
func Never(params ...any) *ZodNever {
	def := &ZodNeverDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeNever,
			Checks: make([]core.ZodCheck, 0),
		},
		Type: core.ZodTypeNever,
	}

	schema := createZodNeverFromDef(def)

	// Apply schema parameters using modern error handling
	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := issues.CreateErrorMap(p)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		case core.SchemaParams:
			// Handle core.SchemaParams
			if p.Error != nil {
				errorMap := issues.CreateErrorMap(p.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}

			// Handle description
			if p.Description != "" {
				if schema.internals.Bag == nil {
					schema.internals.Bag = make(map[string]any)
				}
				schema.internals.Bag["description"] = p.Description
			}

			// Handle custom parameters
			if p.Params != nil {
				if schema.internals.Bag == nil {
					schema.internals.Bag = make(map[string]any)
				}
				for k, v := range p.Params {
					schema.internals.Bag[k] = v
				}
			}
		}
	}

	return schema
}
