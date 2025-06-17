package gozod

import (
	"errors"
	"regexp"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodNilDef defines the configuration for nil validation
type ZodNilDef struct {
	ZodTypeDef
	Type string // "nil"
}

// ZodNilInternals contains nil validator internal state
type ZodNilInternals struct {
	ZodTypeInternals
	Pattern *regexp.Regexp
	Def     *ZodNilDef
	Values  map[interface{}]struct{}
	Isst    ZodIssueInvalidType
	Bag     map[string]interface{}
}

// ZodNil represents a nil validation schema
type ZodNil struct {
	internals *ZodNilInternals
}

// GetInternals returns the internal state of the nil schema
func (z *ZodNil) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the nil-specific internals
func (z *ZodNil) GetZod() *ZodNilInternals {
	return z.internals
}

// CloneFrom implements the Cloneable interface
func (z *ZodNil) CloneFrom(source any) {
	if src, ok := source.(*ZodNil); ok {
		if src.internals.Bag != nil {
			z.internals.Bag = make(map[string]interface{})
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// Coerce attempts to coerce input to nil (always fails for non-nil values)
func (z *ZodNil) Coerce(input any, ctx ...*ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Parse validates that input is nil
func (z *ZodNil) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	if input != nil {
		rawIssue := CreateInvalidTypeIssue(input, "nil", string(GetParsedType(input)))
		finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}

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

	return nil, nil
}

// MustParse parses the input and panics on validation failure
func (z *ZodNil) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Check adds a validation check
func (z *ZodNil) Check(fn CheckFn) ZodType[any, any] {
	custom := NewZodCustom(fn, SchemaParams{})
	custom.GetInternals().Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		result, err := z.Parse(payload.Value, ctx)
		if err != nil {
			payload.Issues = append(payload.Issues, ZodRawIssue{
				Code:    "invalid_type",
				Message: err.Error(),
				Input:   payload.Value,
			})
			return payload
		}
		payload.Value = result
		fn(payload)
		return payload
	}
	return any(custom).(ZodType[any, any])
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform adds data transformation
func (z *ZodNil) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a new transform with another transformation
func (z *ZodNil) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodNil) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the nil schema optional
func (z *ZodNil) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable makes the nil schema nilable
func (z *ZodNil) Nilable() ZodType[any, any] {
	return Nilable(any(z).(ZodType[any, any]))
}

// Nullish makes the nil schema both optional and nilable
func (z *ZodNil) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Refine provides type-safe nil validation
func (z *ZodNil) Refine(fn func() bool, params ...SchemaParams) *ZodNil {
	result := z.RefineAny(func(v any) bool {
		if v != nil {
			return false
		}
		return fn()
	}, params...)
	return result.(*ZodNil)
}

// RefineAny provides flexible validation with any type
func (z *ZodNil) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	var schemaParams SchemaParams
	if len(params) > 0 {
		schemaParams = params[0]
	}
	return NewZodCustom(RefineFn[interface{}](fn), schemaParams)
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodNilDefault represents a nil schema with default value
type ZodNilDefault struct {
	*ZodDefault[*ZodNil]
}

// Default creates a nil schema with default value
func (z *ZodNil) Default(value any) ZodNilDefault {
	return ZodNilDefault{
		&ZodDefault[*ZodNil]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a nil schema with default function
func (z *ZodNil) DefaultFunc(fn func() any) ZodNilDefault {
	return ZodNilDefault{
		&ZodDefault[*ZodNil]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

// Refine adds validation to default wrapper
func (s ZodNilDefault) Refine(fn func() bool, params ...SchemaParams) ZodNilDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodNilDefault{
		&ZodDefault[*ZodNil]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds transformation to default wrapper
func (s ZodNilDefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(fn)
}

// Optional makes default wrapper optional
func (s ZodNilDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable makes default wrapper nilable
func (s ZodNilDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ZodNilPrefault represents a nil schema with prefault value
type ZodNilPrefault struct {
	*ZodPrefault[*ZodNil]
}

// Prefault creates a nil schema with prefault value
func (z *ZodNil) Prefault(value any) ZodNilPrefault {
	// construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
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

	return ZodNilPrefault{
		&ZodPrefault[*ZodNil]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a nil schema with prefault function
func (z *ZodNil) PrefaultFunc(fn func() any) ZodNilPrefault {
	// construct Prefault's internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
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

	return ZodNilPrefault{
		&ZodPrefault[*ZodNil]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  fn,
			isFunction:    true,
		},
	}
}

// Refine adds validation to prefault wrapper
func (n ZodNilPrefault) Refine(fn func() bool, params ...SchemaParams) ZodNilPrefault {
	newInner := n.innerType.Refine(fn, params...)

	// construct new internals
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

	return ZodNilPrefault{
		&ZodPrefault[*ZodNil]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: n.prefaultValue,
			prefaultFunc:  n.prefaultFunc,
			isFunction:    n.isFunction,
		},
	}
}

// Transform adds transformation to prefault wrapper
func (n ZodNilPrefault) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return n.TransformAny(fn)
}

// Optional makes prefault wrapper optional
func (n ZodNilPrefault) Optional() ZodType[any, any] {
	return Optional(any(n).(ZodType[any, any]))
}

// Nilable makes prefault wrapper nilable
func (n ZodNilPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(n).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodNilFromDef creates a ZodNil from definition
func createZodNilFromDef(def *ZodNilDef) *ZodNil {
	// Create internals with standardized pattern
	internals := &ZodNilInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Values:           make(map[interface{}]struct{}),
		Bag:              make(map[string]interface{}),
	}

	// Set up nil value in Values set
	internals.Values[nil] = struct{}{}

	// Set up constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		nilDef := &ZodNilDef{
			ZodTypeDef: *newDef,
			Type:       ZodTypeNil,
		}
		return any(createZodNilFromDef(nilDef)).(ZodType[any, any])
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		parseCtx := (*ParseContext)(nil)
		if ctx != nil {
			parseCtx = ctx
		}

		// use parseType template, including smart type inference and nil value validation
		typeChecker := func(v any) (any, bool) {
			// nil type only accepts nil value
			if v == nil {
				return nil, true
			}
			return nil, false
		}

		coercer := func(v any) (any, bool) {
			// nil type does not support coercion, only accepts nil
			return nil, false
		}

		validator := func(value any, checks []ZodCheck, ctx *ParseContext) error {
			// run basic checks
			if len(checks) > 0 {
				runChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
			}

			// nil type does not need additional validation, because typeChecker already ensures the value is nil
			return nil
		}

		result, err := parseType[any](
			payload.Value,
			&internals.ZodTypeInternals,
			"nil",
			typeChecker,
			func(v any) (*any, bool) { ptr, ok := v.(*any); return ptr, ok },
			validator,
			coercer,
			parseCtx,
		)

		if err != nil {
			// convert error to ParsePayload format
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

	schema := &ZodNil{internals: internals}

	// Initialize the schema using existing infrastructure
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

// NewZodNil creates a new nil schema with optional parameters using standardized pattern
func NewZodNil(params ...SchemaParams) *ZodNil {
	def := &ZodNilDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeNil,
			Checks: make([]ZodCheck, 0),
		},
		Type: ZodTypeNil,
	}

	schema := createZodNilFromDef(def)

	// Apply schema parameters with modern error handling using utility functions
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

		// Handle coercion flag (though nil typically doesn't need coercion)
		if param.Coerce {
			if schema.internals.Bag == nil {
				schema.internals.Bag = make(map[string]interface{})
			}
			schema.internals.Bag["coerce"] = true
		}
	}

	return schema
}

///////////////////////////
////   NIL CONSTRUCTORS (Public API) ////
///////////////////////////

// Nil creates a new nil schema that only accepts nil values
func Nil(params ...SchemaParams) ZodType[any, any] {
	return any(NewZodNil(params...)).(ZodType[any, any])
}

// Null creates a new nil schema (alias for compatibility)
func Null(params ...SchemaParams) ZodType[any, any] {
	return any(NewZodNil(params...)).(ZodType[any, any])
}

////////////////////////////
////   HELPER FUNCTIONS ////
////////////////////////////

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodNil) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}
