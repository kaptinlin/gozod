package types

import (
	"errors"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodAnyDef defines the configuration for any validation
type ZodAnyDef struct {
	core.ZodTypeDef
	Type string // "any"
}

// ZodAnyInternals contains any validator internal state
type ZodAnyInternals struct {
	core.ZodTypeInternals
	Def  *ZodAnyDef                 // Schema definition
	Isst issues.ZodIssueInvalidType // Invalid type issue template (never used for Any)
	Bag  map[string]any             // Additional metadata
}

// ZodAny represents the concrete implementation of any validation schema
type ZodAny struct {
	internals *ZodAnyInternals
}

// GetInternals returns the internal state of the any schema
func (z *ZodAny) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the any-specific internals
func (z *ZodAny) GetZod() *ZodAnyInternals {
	return z.internals
}

// CloneFrom implements the Cloneable interface
func (z *ZodAny) CloneFrom(source any) {
	if src, ok := source.(*ZodAny); ok {
		if src.internals.Bag != nil {
			z.internals.Bag = make(map[string]any)
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// Parse validates that input can be any value with smart type inference
// Any vs Unknown distinction:
// - Any: explicitly accepts "any type", used for intentionally generic scenarios
// - Unknown: represents "type unknown, needs runtime checking", used for type inference scenarios
// - In Go, both have same validation behavior but different semantic intent
func (z *ZodAny) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Handle nil input with Nilable logic
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("any", input)
			finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return (*any)(nil), nil // Return typed nil pointer for Any
	}

	// Any type accepts any value and preserves smart type inference
	// Run validation checks if any exist
	if len(z.internals.Checks) > 0 {
		payload := &core.ParsePayload{
			Value:  input,
			Issues: make([]core.ZodRawIssue, 0),
		}
		engine.RunChecksOnValue(input, z.internals.Checks, payload, parseCtx)
		if len(payload.Issues) > 0 {
			finalizedIssues := make([]core.ZodIssue, len(payload.Issues))
			for i, rawIssue := range payload.Issues {
				finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
			}
			return nil, issues.NewZodError(finalizedIssues)
		}
	}

	// Return original value to maintain smart type inference
	return input, nil
}

// MustParse validates the input and panics on failure
func (z *ZodAny) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Check adds a validation check to the any type
func (z *ZodAny) Check(fn core.CheckFn) core.ZodType[any, any] {
	check := checks.NewCustom[any](func(v any) bool {
		payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
		fn(payload)
		return len(payload.Issues) == 0
	}, core.SchemaParams{})
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform adds data transformation logic to the any schema
func (z *ZodAny) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny adds flexible data transformation logic
func (z *ZodAny) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodAny) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the any schema optional
func (z *ZodAny) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the any schema nilable
func (z *ZodAny) Nilable() core.ZodType[any, any] {
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Nullish makes the any schema both optional and nilable
func (z *ZodAny) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Refine adds type-safe custom validation logic to the any schema
func (z *ZodAny) Refine(fn func(any) bool, params ...any) *ZodAny {
	result := z.RefineAny(fn, params...)
	return result.(*ZodAny)
}

// RefineAny adds flexible custom validation logic to the any schema
func (z *ZodAny) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodAny) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodAnyDefault represents an any schema with default value
type ZodAnyDefault struct {
	*ZodDefault[*ZodAny]
}

////////////////////////////
////   DEFAULT METHODS   ////
////////////////////////////

// Default adds a default value to the any schema, returns ZodAnyDefault support chain call
func (z *ZodAny) Default(value any) ZodAnyDefault {
	return ZodAnyDefault{
		&ZodDefault[*ZodAny]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the any schema, returns ZodAnyDefault support chain call
func (z *ZodAny) DefaultFunc(fn func() any) ZodAnyDefault {
	return ZodAnyDefault{
		&ZodDefault[*ZodAny]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

////////////////////////////
////   ANYDEFAULT CHAINING METHODS ////
////////////////////////////

// Refine adds a flexible validation function to the any schema, returns ZodAnyDefault support chain call
func (s ZodAnyDefault) Refine(fn func(any) bool, params ...any) ZodAnyDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodAnyDefault{
		&ZodDefault[*ZodAny]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodAnyDefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(fn)
}

// Optional adds an optional check to the any schema, returns ZodType support chain call
func (s ZodAnyDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the any schema, returns ZodType support chain call
func (s ZodAnyDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

////////////////////////////
////   ANY PREFAULT WRAPPER ////
////////////////////////////

// ZodAnyPrefault is a Prefault wrapper for Any type, returns ZodAnyPrefault support chain call
type ZodAnyPrefault struct {
	*ZodPrefault[*ZodAny] // Embed specific pointer, allow method promotion
}

////////////////////////////
////   PREFAULT METHODS   ////
////////////////////////////

// Prefault adds a prefault value to the any schema, returns ZodAnyPrefault support chain call
func (z *ZodAny) Prefault(value any) ZodAnyPrefault {
	// Construct Prefault's internals, Type = "prefault", copy internal type's checks/optional/nilable
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodAnyPrefault{
		&ZodPrefault[*ZodAny]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the any schema, returns ZodAnyPrefault support chain call
func (z *ZodAny) PrefaultFunc(fn func() any) ZodAnyPrefault {
	// Construct Prefault's internals, Type = "prefault", copy internal type's checks/optional/nilable
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodAnyPrefault{
		&ZodPrefault[*ZodAny]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  fn,
			isFunction:    true,
		},
	}
}

////////////////////////////
////   ANYPREFAULT chaining methods ////
////////////////////////////

// Refine adds a flexible validation function to the any schema, returns ZodAnyPrefault support chain call
func (a ZodAnyPrefault) Refine(fn func(any) bool, params ...any) ZodAnyPrefault {
	newInner := a.innerType.Refine(fn, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodAnyPrefault{
		&ZodPrefault[*ZodAny]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: a.prefaultValue,
			prefaultFunc:  a.prefaultFunc,
			isFunction:    a.isFunction,
		},
	}
}

// Transform adds data conversion, returns a generic ZodType support conversion pipeline
func (a ZodAnyPrefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use embedded ZodPrefault's TransformAny method
	return a.TransformAny(fn)
}

// Optional adds an optional check to the any schema, returns ZodType support chain call
func (a ZodAnyPrefault) Optional() core.ZodType[any, any] {
	// Wrap current ZodAnyPrefault instance, keep Prefault logic
	return Optional(any(a).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the any schema, returns ZodType support chain call
func (a ZodAnyPrefault) Nilable() core.ZodType[any, any] {
	// Wrap current ZodAnyPrefault instance, keep Prefault logic
	return Nilable(any(a).(core.ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodAnyFromDef creates a ZodAny from definition
func createZodAnyFromDef(def *ZodAnyDef) *ZodAny {
	internals := &ZodAnyInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{ZodIssueBase: issues.ZodIssueBase{}, Expected: "any", Received: ""}, // Never used but consistent
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		anyDef := &ZodAnyDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeAny,
		}
		return createZodAnyFromDef(anyDef)
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		zodInstance := &ZodAny{internals: internals}
		result, err := zodInstance.Parse(payload.Value, ctx)
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

	zodSchema := &ZodAny{internals: internals}
	engine.InitZodType(zodSchema, &def.ZodTypeDef)
	return zodSchema
}

// Any creates an any validation schema
func Any(params ...any) *ZodAny {
	def := &ZodAnyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeAny,
			Checks: make([]core.ZodCheck, 0),
		},
		Type: core.ZodTypeAny,
	}

	zodSchema := createZodAnyFromDef(def)

	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := issues.CreateErrorMap(p)
			if errorMap != nil {
				def.Error = errorMap
				zodSchema.internals.Error = errorMap
			}
		case core.SchemaParams:
			// Handle core.SchemaParams
			if p.Error != nil {
				errorMap := issues.CreateErrorMap(p.Error)
				if errorMap != nil {
					def.Error = errorMap
					zodSchema.internals.Error = errorMap
				}
			}

			// Handle description
			if p.Description != "" {
				if zodSchema.internals.Bag == nil {
					zodSchema.internals.Bag = make(map[string]any)
				}
				zodSchema.internals.Bag["description"] = p.Description
			}

			// Handle custom parameters
			if p.Params != nil {
				if zodSchema.internals.Bag == nil {
					zodSchema.internals.Bag = make(map[string]any)
				}
				for k, v := range p.Params {
					zodSchema.internals.Bag[k] = v
				}
			}
		}
	}

	return zodSchema
}
