package gozod

import (
	"errors"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodAnyDef defines the configuration for any validation
type ZodAnyDef struct {
	ZodTypeDef
	Type string // "any"
}

// ZodAnyInternals contains any validator internal state
type ZodAnyInternals struct {
	ZodTypeInternals
	Def  *ZodAnyDef             // Schema definition
	Isst ZodIssueInvalidType    // Invalid type issue template (never used for Any)
	Bag  map[string]interface{} // Additional metadata
}

// ZodAny represents the concrete implementation of any validation schema
type ZodAny struct {
	internals *ZodAnyInternals
}

// GetInternals returns the internal state of the any schema
func (z *ZodAny) GetInternals() *ZodTypeInternals {
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
			z.internals.Bag = make(map[string]interface{})
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// Coerce attempts to coerce input (any accepts all values without coercion)
func (z *ZodAny) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Parse validates that input can be any value with smart type inference
// Any vs Unknown distinction:
// - Any: explicitly accepts "any type", used for intentionally generic scenarios
// - Unknown: represents "type unknown, needs runtime checking", used for type inference scenarios
// - In Go, both have same validation behavior but different semantic intent
func (z *ZodAny) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Handle nil input with Nilable logic
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "any", "null")
			finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*interface{})(nil), nil // Return typed nil pointer for Any
	}

	// Any type accepts any value and preserves smart type inference
	// Run validation checks if any exist
	if len(z.internals.Checks) > 0 {
		payload := &ParsePayload{
			Value:  input,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(input, z.internals.Checks, payload, parseCtx)
		if len(payload.Issues) > 0 {
			return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
		}
	}

	// Return original value to maintain smart type inference
	return input, nil
}

// MustParse validates the input and panics on failure
func (z *ZodAny) MustParse(input any, ctx ...*ParseContext) any {
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
func (z *ZodAny) Check(fn CheckFn) ZodType[any, any] {
	check := NewCustom[any](func(v any) bool {
		payload := &ParsePayload{Value: v, Issues: make([]ZodRawIssue, 0)}
		fn(payload)
		return len(payload.Issues) == 0
	}, SchemaParams{})
	return AddCheck(any(z).(ZodType[any, any]), check)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform adds data transformation logic to the any schema
func (z *ZodAny) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny adds flexible data transformation logic
func (z *ZodAny) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodAny) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the any schema optional
func (z *ZodAny) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable makes the any schema nilable
func (z *ZodAny) Nilable() ZodType[any, any] {
	return Nilable(any(z).(ZodType[any, any]))
}

// Nullish makes the any schema both optional and nilable
func (z *ZodAny) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Refine adds type-safe custom validation logic to the any schema
func (z *ZodAny) Refine(fn func(any) bool, params ...SchemaParams) *ZodAny {
	result := z.RefineAny(fn, params...)
	return result.(*ZodAny)
}

// RefineAny adds flexible custom validation logic to the any schema
func (z *ZodAny) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodAny) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
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
func (s ZodAnyDefault) Refine(fn func(any) bool, params ...SchemaParams) ZodAnyDefault {
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
func (s ZodAnyDefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(fn)
}

// Optional adds an optional check to the any schema, returns ZodType support chain call
func (s ZodAnyDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the any schema, returns ZodType support chain call
func (s ZodAnyDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
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
	// Construct Prefault's internals, Type = "prefault", copy internal type's checks/coerce/optional/nilable
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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
	// Construct Prefault's internals, Type = "prefault", copy internal type's checks/coerce/optional/nilable
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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
func (a ZodAnyPrefault) Refine(fn func(any) bool, params ...SchemaParams) ZodAnyPrefault {
	newInner := a.innerType.Refine(fn, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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
func (a ZodAnyPrefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodPrefault's TransformAny method
	return a.TransformAny(fn)
}

// Optional adds an optional check to the any schema, returns ZodType support chain call
func (a ZodAnyPrefault) Optional() ZodType[any, any] {
	// Wrap current ZodAnyPrefault instance, keep Prefault logic
	return Optional(any(a).(ZodType[any, any]))
}

// Nilable adds a nilable check to the any schema, returns ZodType support chain call
func (a ZodAnyPrefault) Nilable() ZodType[any, any] {
	// Wrap current ZodAnyPrefault instance, keep Prefault logic
	return Nilable(any(a).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodAnyFromDef creates a ZodAny from definition
func createZodAnyFromDef(def *ZodAnyDef) *ZodAny {
	internals := &ZodAnyInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             ZodIssueInvalidType{Expected: "any"}, // Never used but consistent
		Bag:              make(map[string]interface{}),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		anyDef := &ZodAnyDef{
			ZodTypeDef: *newDef,
			Type:       ZodTypeAny,
		}
		return createZodAnyFromDef(anyDef)
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodAny{internals: internals}
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

	schema := &ZodAny{internals: internals}
	initZodType(schema, &def.ZodTypeDef)
	return schema
}

// NewZodAny creates a new any validation schema
func NewZodAny(params ...SchemaParams) *ZodAny {
	def := &ZodAnyDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeAny,
			Checks: make([]ZodCheck, 0),
		},
		Type: ZodTypeAny,
	}

	schema := createZodAnyFromDef(def)

	if len(params) > 0 {
		param := params[0]
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}

		// Handle description
		if param.Description != "" {
			if schema.internals.Bag == nil {
				schema.internals.Bag = make(map[string]interface{})
			}
			schema.internals.Bag["description"] = param.Description
		}

		// Handle coercion flag
		if param.Coerce {
			if schema.internals.Bag == nil {
				schema.internals.Bag = make(map[string]interface{})
			}
			schema.internals.Bag["coerce"] = true
		}

		// Handle custom parameters
		if param.Params != nil {
			if schema.internals.Bag == nil {
				schema.internals.Bag = make(map[string]interface{})
			}
			for k, v := range param.Params {
				schema.internals.Bag[k] = v
			}
		}
	}

	return schema
}

// Any creates an any validation schema
func Any(params ...SchemaParams) *ZodAny {
	return NewZodAny(params...)
}
