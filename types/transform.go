package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
)

// ZodTransformDef defines the configuration for type conversion
type ZodTransformDef[Out, In any] struct {
	core.ZodTypeDef
	Type        string
	TransformFn func(In, *core.RefinementContext) (Out, error)
}

// ZodTransformInternals contains the internal state of the Transform validator
type ZodTransformInternals[Out, In any] struct {
	core.ZodTypeInternals
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
func (z *ZodTransform[Out, In]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse executes conversion and validation
func (z *ZodTransform[Out, In]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	if parseCtx == nil {
		parseCtx = core.NewParseContext()
	}

	// Create parse payload
	// Use constructor instead of direct struct literal to respect private fields
	payload := core.NewParsePayloadWithPath(input, []any{})

	// Create enhanced context, support error reporting
	refinementCtx := &core.RefinementContext{
		ParseContext: parseCtx,
		Value:        input,
		AddIssue: func(issue core.ZodIssue) {
			// Convert ZodIssue to RawIssue using standardized converter
			rawIssue := issues.ConvertZodIssueToRaw(issue)
			rawIssue.Path = payload.GetPath()
			payload.AddIssue(rawIssue)
		},
	}

	// Transform core
	var inVal In
	castOk := false
	if input != nil {
		if v, ok := input.(In); ok {
			inVal = v
			castOk = true
		}
	} else {
		// keep zero value for In (may be nil pointer) when input is nil
		castOk = true
	}
	if !castOk {
		issue := issues.CreateInvalidTypeIssue("transform input", input)
		payload.AddIssue(issue)
	}

	transformedGeneric, err := z.internals.Def.TransformFn(inVal, refinementCtx)
	if err != nil {
		// Conversion failed: add error, do not modify data
		issue := issues.CreateCustomIssue(err.Error(), nil, input)
		payload.AddIssue(issue)
	} else {
		// Conversion successful: directly modify payload.GetValue()
		payload.SetValue(transformedGeneric)
	}

	// Check if there are errors reported through AddIssue
	if len(payload.GetIssues()) > 0 {
		finalIssues := make([]core.ZodIssue, len(payload.GetIssues()))
		for i, rawIssue := range payload.GetIssues() {
			finalIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
		}
		return nil, issues.NewZodError(finalIssues)
	}

	return payload.GetValue(), nil
}

// MustParse executes conversion, panics on failure
func (z *ZodTransform[Out, In]) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodTransform[Out, In]) Nilable() core.ZodType[any, any] {
	// Transform's Nilable behavior: create a new Transform supporting nil
	newDef := &ZodTransformDef[any, any]{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		TransformFn: func(v any, ctx *core.RefinementContext) (any, error) {
			// Attempt to cast to the original expected input type
			inVal, _ := v.(In)
			return z.internals.Def.TransformFn(inVal, ctx)
		},
	}

	return createZodTransformFromDef(newDef)
}

// Refine adds custom validation
func (z *ZodTransform[Out, In]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	custom := Custom(fn, params)

	return Pipe[any, any](any(z).(core.ZodType[any, any]), any(custom).(core.ZodType[any, any]))
}

// Check adds check API
func (z *ZodTransform[Out, In]) Check(fn func(*core.ParsePayload) error) core.ZodType[any, any] {
	checkFn := func(payload *core.ParsePayload) {
		if err := fn(payload); err != nil {
			issue := issues.NewRawIssue("custom", payload.GetValue(), issues.WithOrigin("transform"))
			issue.Message = err.Error()
			payload.AddIssue(issue)
		}
	}

	custom := Custom(core.CheckFn(checkFn))

	return Pipe[any, any](any(z).(core.ZodType[any, any]), any(custom).(core.ZodType[any, any]))
}

// Transform chain conversion
func (z *ZodTransform[Out, In]) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// wrap fn to match type parameters In,Out=any,any
	newTransform := Transform[any, any](fn)
	return Pipe[any, any](any(z).(core.ZodType[any, any]), any(newTransform).(core.ZodType[any, any]))
}

// TransformAny flexible version of conversion
func (z *ZodTransform[Out, In]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	newTransform := Transform[any, any](fn)

	return Pipe[any, any](any(z).(core.ZodType[any, any]), any(newTransform).(core.ZodType[any, any]))
}

// Pipe pipe combination
func (z *ZodTransform[Out, In]) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return Pipe[any, any](any(z).(core.ZodType[any, any]), out)
}

////////////////////////////
////   CONSTRUCTOR      ////
////////////////////////////

// Transform create new ZodTransform
func Transform[Out, In any](transformFn func(In, *core.RefinementContext) (Out, error)) *ZodTransform[Out, In] {
	def := &ZodTransformDef[Out, In]{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "transform",
			Checks: make([]core.ZodCheck, 0),
		},
		Type:        "transform",
		TransformFn: transformFn,
	}

	return createZodTransformFromDef(def)
}

// createZodTransformFromDef create ZodTransform from definition
func createZodTransformFromDef[Out, In any](def *ZodTransformDef[Out, In]) *ZodTransform[Out, In] {
	internals := &ZodTransformInternals[Out, In]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals("transform"),
		Def:              def,
	}

	// Set parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		// Create enhanced context, support error reporting
		refinementCtx := &core.RefinementContext{
			ParseContext: ctx,
			Value:        payload.GetValue(),
			AddIssue: func(issue core.ZodIssue) {
				// Convert ZodIssue to RawIssue using standardized converter
				rawIssue := issues.ConvertZodIssueToRaw(issue)
				rawIssue.Path = payload.GetPath()
				payload.AddIssue(rawIssue)
			},
		}

		// 🔥 Transform core principle: can modify data (fundamental difference from Refine)
		// Attempt to cast to expected input type
		var inVal In
		castOk := false
		if payload.GetValue() != nil {
			if v, ok := payload.GetValue().(In); ok {
				inVal = v
				castOk = true
			}
		} else {
			// keep zero value for In (may be nil pointer) when input is nil
			castOk = true
		}
		if !castOk {
			issue := issues.CreateInvalidTypeIssue("transform input", payload.GetValue())
			payload.AddIssue(issue)
		}
		result, err := def.TransformFn(inVal, refinementCtx)
		if err != nil {
			// Conversion failed: add error, keep original value
			issue := issues.CreateCustomIssue(err.Error(), nil, payload.GetValue())
			payload.AddIssue(issue)
		} else {
			// Conversion successful: directly modify payload.GetValue()
			payload.SetValue(result)
		}
		return payload
	}

	return &ZodTransform[Out, In]{internals: internals}
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodTransform[In, Out]) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// Tg is a concise wrapper around Transform that avoids specifying generic type
// parameters explicitly. It simply returns a ZodType[any, any] by instantiating
// Transform with <any, any>. This satisfies the existing test expectations of
// calling a global helper without generics, while keeping the flexible generic
// variant available when needed.
func Tg(transformFn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	t := Transform[any, any](func(a any, ctx *core.RefinementContext) (any, error) {
		return transformFn(a, ctx)
	})
	return any(t).(core.ZodType[any, any])
}
