package gozod

import (
	"errors"
	"fmt"
	"regexp"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodBoolDef defines the configuration for boolean validation
type ZodBoolDef struct {
	ZodTypeDef
	Type     string     // "boolean"
	Checks   []ZodCheck // Boolean-specific validation checks
	AllowNil bool       // Nilable modifier flag
}

// ZodBoolInternals contains boolean validator internal state
type ZodBoolInternals struct {
	ZodTypeInternals
	Def     *ZodBoolDef            // Schema definition
	Checks  []ZodCheck             // Validation checks
	Isst    ZodIssueInvalidType    // Invalid type issue template
	Pattern *regexp.Regexp         // Boolean pattern (if any)
	Bag     map[string]interface{} // Additional metadata (coerce flag, etc.)
}

// ZodBool represents a boolean validation schema with type safety
type ZodBool struct {
	internals *ZodBoolInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodBool) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for boolean type conversion
func (z *ZodBool) Coerce(input interface{}) (interface{}, bool) {
	return coerceToBool(input)
}

// Parse validates and parses input with smart type inference
func (z *ZodBool) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return parseType[bool](
		input,
		&z.internals.ZodTypeInternals,
		"bool",
		func(v any) (bool, bool) { b, ok := v.(bool); return b, ok },
		func(v any) (*bool, bool) { ptr, ok := v.(*bool); return ptr, ok },
		validateBool,
		func(v any) (bool, bool) {
			if shouldCoerce(z.internals.ZodTypeInternals.Bag) {
				return coerceToBool(v)
			}
			return false, false
		},
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodBool) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a type-safe transformation of boolean values
func (z *ZodBool) Transform(fn func(bool, *RefinementContext) (any, error)) ZodType[any, any] {
	wrappedFn := func(input any, ctx *RefinementContext) (any, error) {
		boolVal, isNil, err := extractBoolValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return (*bool)(nil), nil
		}
		return fn(boolVal, ctx)
	}
	return z.TransformAny(wrappedFn)
}

// TransformAny provides flexible transformation with any input/output types
func (z *ZodBool) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodBool) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the boolean optional
func (z *ZodBool) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable makes the boolean nilable
func (z *ZodBool) Nilable() ZodType[any, any] {
	return Nilable(z)
}

// Nullish combines optional and nilable
func (z *ZodBool) Nullish() ZodType[any, any] {
	return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Refine adds type-safe custom validation logic
func (z *ZodBool) Refine(fn func(bool) bool, params ...SchemaParams) *ZodBool {
	result := z.RefineAny(func(v any) bool {
		boolVal, isNil, err := extractBoolValue(v)

		if err != nil {
			return false
		}

		if isNil {
			return true // Let Nilable flag handle nil validation
		}

		return fn(boolVal)
	}, params...)
	return result.(*ZodBool)
}

// RefineAny adds flexible custom validation logic
func (z *ZodBool) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(z, check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodBool) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodBoolDefault is a default value wrapper for boolean type
type ZodBoolDefault struct {
	*ZodDefault[*ZodBool]
}

// DEFAULT METHODS

// Default creates a default wrapper with type safety
func (z *ZodBool) Default(value bool) ZodBoolDefault {
	return ZodBoolDefault{
		&ZodDefault[*ZodBool]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a default wrapper with function
func (z *ZodBool) DefaultFunc(fn func() bool) ZodBoolDefault {
	genericFn := func() any { return fn() }
	return ZodBoolDefault{
		&ZodDefault[*ZodBool]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// ZodBoolDefault chainable validation methods

func (s ZodBoolDefault) Refine(fn func(bool) bool, params ...SchemaParams) ZodBoolDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodBoolDefault{
		&ZodDefault[*ZodBool]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBoolDefault) Transform(fn func(bool, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		boolVal, isNil, err := extractBoolValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilBool
		}
		return fn(boolVal, ctx)
	})
}

func (s ZodBoolDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodBoolDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ZodBoolPrefault is a prefault value wrapper for boolean type
type ZodBoolPrefault struct {
	*ZodPrefault[*ZodBool]
}

// PREFAULT METHODS

// Prefault creates a prefault wrapper with type safety
func (z *ZodBool) Prefault(value bool) ZodBoolPrefault {
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

	return ZodBoolPrefault{
		&ZodPrefault[*ZodBool]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a prefault wrapper with function
func (z *ZodBool) PrefaultFunc(fn func() bool) ZodBoolPrefault {
	genericFn := func() any { return fn() }

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

	return ZodBoolPrefault{
		&ZodPrefault[*ZodBool]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// ZodBoolPrefault chainable validation methods

func (b ZodBoolPrefault) Refine(fn func(bool) bool, params ...SchemaParams) ZodBoolPrefault {
	newInner := b.innerType.Refine(fn, params...)

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

	return ZodBoolPrefault{
		&ZodPrefault[*ZodBool]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: b.prefaultValue,
			prefaultFunc:  b.prefaultFunc,
			isFunction:    b.isFunction,
		},
	}
}

func (b ZodBoolPrefault) Transform(fn func(bool, *RefinementContext) (any, error)) ZodType[any, any] {
	return b.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		boolVal, isNil, err := extractBoolValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilBool
		}
		return fn(boolVal, ctx)
	})
}

func (b ZodBoolPrefault) Optional() ZodType[any, any] {
	return Optional(any(b).(ZodType[any, any]))
}

func (b ZodBoolPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(b).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodBoolFromDef creates a ZodBool from definition using unified patterns
func createZodBoolFromDef(def *ZodBoolDef) *ZodBool {
	internals := &ZodBoolInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           def.Checks,
		Isst:             ZodIssueInvalidType{Expected: "bool"},
		Pattern:          nil,
		Bag:              make(map[string]interface{}),
	}

	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		booleanDef := &ZodBoolDef{
			ZodTypeDef: *newDef,
			Type:       "bool",
			Checks:     newDef.Checks,
		}
		return createZodBoolFromDef(booleanDef)
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		return parseBooleanWithparseType(payload, internals, ctx)
	}

	schema := &ZodBool{internals: internals}
	initZodType(schema, &def.ZodTypeDef)
	return schema
}

// NewZodBool creates a new boolean schema with unified parameter handling
func NewZodBool(params ...SchemaParams) *ZodBool {
	def := &ZodBoolDef{
		ZodTypeDef: ZodTypeDef{
			Type:   "bool",
			Checks: make([]ZodCheck, 0),
		},
		Type:   "bool",
		Checks: make([]ZodCheck, 0),
	}

	schema := createZodBoolFromDef(def)

	if len(params) > 0 {
		param := params[0]

		if param.Coerce {
			schema.internals.Bag["coerce"] = true
			schema.internals.ZodTypeInternals.Bag["coerce"] = true
		}

		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}

		if param.Description != "" {
			schema.internals.Bag["description"] = param.Description
		}
		if param.Abort {
			schema.internals.Bag["abort"] = true
		}
		if len(param.Path) > 0 {
			schema.internals.Bag["path"] = param.Path
		}
	}

	return schema
}

// Bool creates a new boolean schema (package-level constructor)
func Bool(params ...SchemaParams) *ZodBool {
	return NewZodBool(params...)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// GetZod returns the boolean-specific internals
func (z *ZodBool) GetZod() *ZodBoolInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface
func (z *ZodBool) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodBoolInternals }); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]interface{})
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}

		if srcState.Pattern != nil {
			tgtState.Pattern = srcState.Pattern
		}
	}
}

// extractBoolValue extracts boolean value from input with smart handling
func extractBoolValue(input any) (bool, bool, error) {
	if input == nil {
		return false, true, nil
	}

	switch v := input.(type) {
	case bool:
		return v, false, nil
	case *bool:
		if v == nil {
			return false, true, nil
		}
		return *v, false, nil
	default:
		return false, false, fmt.Errorf("%w, received %T", ErrExpectedBool, input)
	}
}

// parseBooleanWithparseType handles the core boolean parsing logic using parseType
func parseBooleanWithparseType(payload *ParsePayload, internals *ZodBoolInternals, ctx *ParseContext) *ParsePayload {
	coerce := false
	if coerceVal, exists := internals.Bag["coerce"]; exists {
		if coerceBool, ok := coerceVal.(bool); ok {
			coerce = coerceBool
		}
	}

	result, err := parseType[bool](
		payload.Value,
		&internals.ZodTypeInternals,
		"bool",
		func(v any) (bool, bool) { b, ok := v.(bool); return b, ok },
		func(v any) (*bool, bool) { ptr, ok := v.(*bool); return ptr, ok },
		validateBool,
		func(v any) (bool, bool) {
			if coerce {
				return coerceToBool(v)
			}
			return false, false
		},
		ctx,
	)

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
