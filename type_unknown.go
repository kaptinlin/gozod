package gozod

import (
	"errors"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodUnknownDef defines the configuration for unknown validation
type ZodUnknownDef struct {
	ZodTypeDef
	Type string // "unknown"
}

// ZodUnknownInternals contains unknown validator internal state
type ZodUnknownInternals struct {
	ZodTypeInternals
	Def  *ZodUnknownDef         // Schema definition
	Isst ZodIssueInvalidType    // Invalid type issue template (never used for Unknown)
	Bag  map[string]interface{} // Additional metadata
}

// ZodUnknown represents the concrete implementation of unknown validation schema
type ZodUnknown struct {
	internals *ZodUnknownInternals
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// GetInternals returns the internal state of the unknown schema
func (z *ZodUnknown) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the unknown-specific internals
func (z *ZodUnknown) GetZod() *ZodUnknownInternals {
	return z.internals
}

// Coerce attempts to coerce input (unknown accepts any value)
func (z *ZodUnknown) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Parse validates and accepts any input value
func (z *ZodUnknown) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Handle nil input with Nilable logic
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "unknown", "null")
			finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*interface{})(nil), nil
	}

	// Unknown type accepts any value and preserves smart type inference
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

	return input, nil
}

// MustParse validates the input and panics on failure
func (z *ZodUnknown) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// CloneFrom implements the Cloneable interface
func (z *ZodUnknown) CloneFrom(source any) {
	if src, ok := source.(*ZodUnknown); ok {
		if src.internals.Bag != nil {
			z.internals.Bag = make(map[string]interface{})
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Check adds a validation check
func (z *ZodUnknown) Check(fn CheckFn) ZodType[any, any] {
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

// Transform adds data transformation logic to the unknown schema
func (z *ZodUnknown) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny adds flexible data transformation logic
func (z *ZodUnknown) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodUnknown) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the unknown schema optional
func (z *ZodUnknown) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable makes the unknown schema nilable
func (z *ZodUnknown) Nilable() ZodType[any, any] {
	return Nilable(any(z).(ZodType[any, any]))
}

// Nullish makes the unknown schema both optional and nilable
func (z *ZodUnknown) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Refine adds type-safe custom validation logic
func (z *ZodUnknown) Refine(fn func(any) bool, params ...SchemaParams) *ZodUnknown {
	result := z.RefineAny(fn, params...)
	return result.(*ZodUnknown)
}

// RefineAny adds flexible custom validation logic
func (z *ZodUnknown) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// Unwrap returns the inner type
func (z *ZodUnknown) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodUnknownDefault represents an unknown schema with default value
// Provides perfect type safety and chainable method support
type ZodUnknownDefault struct {
	*ZodDefault[*ZodUnknown] // Embed concrete pointer to enable method promotion
}

////////////////////////////
////   DEFAULT METHODS   ////
////////////////////////////

// Default creates an unknown schema with default value
func (z *ZodUnknown) Default(value any) ZodUnknownDefault {
	return ZodUnknownDefault{
		&ZodDefault[*ZodUnknown]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates an unknown schema with default function
func (z *ZodUnknown) DefaultFunc(fn func() any) ZodUnknownDefault {
	return ZodUnknownDefault{
		&ZodDefault[*ZodUnknown]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

////////////////////////////
////   UNKNOWNDEFAULT CHAINING METHODS ////
////////////////////////////

// Refine adds a flexible validation function to the unknown schema, returns ZodUnknownDefault
func (s ZodUnknownDefault) Refine(fn func(any) bool, params ...SchemaParams) ZodUnknownDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodUnknownDefault{
		&ZodDefault[*ZodUnknown]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns generic ZodType to support transformation pipeline
func (s ZodUnknownDefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(fn)
}

// Optional makes the unknown schema optional
func (s ZodUnknownDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable makes the unknown schema nilable
func (s ZodUnknownDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

////////////////////////////
////   UNKNOWN PREFAULT WRAPPER ////
////////////////////////////

// ZodUnknownPrefault represents an unknown schema with prefault value
// Provides perfect type safety and chainable method support
type ZodUnknownPrefault struct {
	*ZodPrefault[*ZodUnknown] // Embed concrete pointer to enable method promotion
}

////////////////////////////
////   PREFAULT METHODS   ////
////////////////////////////

// Prefault creates an unknown schema with prefault value
func (z *ZodUnknown) Prefault(value any) ZodUnknownPrefault {
	// Construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
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

	return ZodUnknownPrefault{
		&ZodPrefault[*ZodUnknown]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates an unknown schema with prefault function
func (z *ZodUnknown) PrefaultFunc(fn func() any) ZodUnknownPrefault {
	// Construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
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

	return ZodUnknownPrefault{
		&ZodPrefault[*ZodUnknown]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  fn,
			isFunction:    true,
		},
	}
}

////////////////////////////
////   UNKNOWNPREFAULT CHAINING METHODS ////
////////////////////////////

// Refine adds a flexible validation function to the unknown schema, returns ZodUnknownPrefault
func (u ZodUnknownPrefault) Refine(fn func(any) bool, params ...SchemaParams) ZodUnknownPrefault {
	newInner := u.innerType.Refine(fn, params...)

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

	return ZodUnknownPrefault{
		&ZodPrefault[*ZodUnknown]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: u.prefaultValue,
			prefaultFunc:  u.prefaultFunc,
			isFunction:    u.isFunction,
		},
	}
}

// Transform adds data transformation, returns generic ZodType to support transformation pipeline
func (u ZodUnknownPrefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodPrefault's TransformAny method
	return u.TransformAny(fn)
}

// Optional makes the unknown schema optional
func (u ZodUnknownPrefault) Optional() ZodType[any, any] {
	// Wrap current ZodUnknownPrefault instance, preserve Prefault logic
	return Optional(any(u).(ZodType[any, any]))
}

// Nilable makes the unknown schema nilable
func (u ZodUnknownPrefault) Nilable() ZodType[any, any] {
	// Wrap current ZodUnknownPrefault instance, preserve Prefault logic
	return Nilable(any(u).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodUnknownFromDef creates a ZodUnknown from definition
func createZodUnknownFromDef(def *ZodUnknownDef) *ZodUnknown {
	internals := &ZodUnknownInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             ZodIssueInvalidType{Expected: "unknown"},
		Bag:              make(map[string]interface{}),
	}

	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		unknownDef := &ZodUnknownDef{
			ZodTypeDef: *newDef,
			Type:       ZodTypeUnknown,
		}
		return createZodUnknownFromDef(unknownDef)
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodUnknown{internals: internals}
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

	schema := &ZodUnknown{internals: internals}
	initZodType(schema, &def.ZodTypeDef)
	return schema
}

// NewZodUnknown creates a new unknown validation schema
func NewZodUnknown(params ...SchemaParams) *ZodUnknown {
	def := &ZodUnknownDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeUnknown,
			Checks: make([]ZodCheck, 0),
		},
		Type: ZodTypeUnknown,
	}

	schema := createZodUnknownFromDef(def)

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

// Unknown creates an unknown validation schema
func Unknown(params ...SchemaParams) *ZodUnknown {
	return NewZodUnknown(params...)
}
