package types

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"

	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for transformations
var (
	ErrTransformNilBool = errors.New("cannot transform nil bool value")
	ErrExpectedBool     = errors.New("expected bool type")
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodBoolDef defines the configuration for boolean validation
type ZodBoolDef struct {
	core.ZodTypeDef
	Type     core.ZodTypeCode // Type identifier using type-safe constants
	Checks   []core.ZodCheck  // Boolean-specific validation checks
	AllowNil bool             // Nilable modifier flag
}

// ZodBoolInternals contains boolean validator internal state
type ZodBoolInternals struct {
	core.ZodTypeInternals
	Def     *ZodBoolDef                // Schema definition
	Checks  []core.ZodCheck            // Validation checks
	Isst    issues.ZodIssueInvalidType // Invalid type issue template
	Pattern *regexp.Regexp             // Boolean pattern (if any)
	Bag     map[string]any             // Additional metadata (coerce flag, etc.)
}

// ZodBool represents a boolean validation schema with type safety
type ZodBool struct {
	internals *ZodBoolInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodBool) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for boolean type conversion
func (z *ZodBool) Coerce(input any) (any, bool) {
	result, err := coerce.ToBool(input)
	return result, err == nil
}

// Parse implements intelligent type inference and validation
func (z *ZodBool) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// Reuse the first context if supplied; otherwise allocate lazily inside ParsePrimitive.
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return engine.ParsePrimitive[bool](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeBool,
		validateBool,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodBool) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodBool) Transform(fn func(bool, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	wrappedFn := func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for smart nil checking
		if reflectx.IsNil(input) {
			return (*bool)(nil), nil
		}

		// Handle pointer dereferencing using reflectx
		value := input
		if reflectx.IsPointer(input) {
			if deref, ok := reflectx.Deref(input); ok {
				if reflectx.IsNil(deref) {
					return (*bool)(nil), nil
				}
				value = deref
			}
		}

		// Direct type assertion for bool
		if boolVal, ok := value.(bool); ok {
			return fn(boolVal, ctx)
		}

		return nil, fmt.Errorf("%w, received %T", ErrExpectedBool, input)
	}
	return z.TransformAny(wrappedFn)
}

// TransformAny provides flexible transformation with any input/output types
func (z *ZodBool) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodBool) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the boolean optional
func (z *ZodBool) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable makes the boolean nilable
func (z *ZodBool) Nilable() core.ZodType[any, any] {
	return Nilable(z)
}

// Nullish combines optional and nilable
func (z *ZodBool) Nullish() core.ZodType[any, any] {
	return any(Nullish(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Refine adds type-safe custom validation logic
func (z *ZodBool) Refine(fn func(bool) bool, params ...any) *ZodBool {
	result := z.RefineAny(func(v any) bool {
		// Use reflectx for smart nil checking
		if reflectx.IsNil(v) {
			return true // Let Nilable flag handle nil validation
		}

		// Handle pointer dereferencing using reflectx
		value := v
		if reflectx.IsPointer(v) {
			if deref, ok := reflectx.Deref(v); ok {
				if reflectx.IsNil(deref) {
					return true // Let Nilable flag handle nil validation
				}
				value = deref
			}
		}

		// Direct type assertion for bool
		if boolVal, ok := value.(bool); ok {
			return fn(boolVal)
		}

		return false // Invalid type fails refinement
	}, params...)
	return result.(*ZodBool)
}

// RefineAny adds flexible custom validation logic
func (z *ZodBool) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodBool) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// VALIDATION FUNCTIONS
//////////////////////////

// validateBool validates boolean values with checks
func validateBool(value bool, checks []core.ZodCheck, ctx *core.ParseContext) error {
	if len(checks) > 0 {
		payload := &core.ParsePayload{
			Value:  value,
			Issues: make([]core.ZodRawIssue, 0),
		}
		engine.RunChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.Issues, ctx))
		}
	}
	return nil
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

func (s ZodBoolDefault) Refine(fn func(bool) bool, params ...any) ZodBoolDefault {
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

func (s ZodBoolDefault) Transform(fn func(bool, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for smart nil checking
		if reflectx.IsNil(input) {
			return nil, ErrTransformNilBool
		}

		// Handle pointer dereferencing using reflectx
		value := input
		if reflectx.IsPointer(input) {
			if deref, ok := reflectx.Deref(input); ok {
				if reflectx.IsNil(deref) {
					return nil, ErrTransformNilBool
				}
				value = deref
			}
		}

		// Direct type assertion for bool
		if boolVal, ok := value.(bool); ok {
			return fn(boolVal, ctx)
		}

		return nil, fmt.Errorf("%w, received %T", ErrExpectedBool, input)
	})
}

func (s ZodBoolDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodBoolDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

// ZodBoolPrefault is a prefault value wrapper for boolean type
type ZodBoolPrefault struct {
	*ZodPrefault[*ZodBool]
}

// PREFAULT METHODS

// Prefault creates a prefault wrapper with type safety
func (z *ZodBool) Prefault(value bool) ZodBoolPrefault {
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

func (b ZodBoolPrefault) Refine(fn func(bool) bool, params ...any) ZodBoolPrefault {
	newInner := b.innerType.Refine(fn, params...)

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

func (b ZodBoolPrefault) Transform(fn func(bool, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return b.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for smart nil checking
		if reflectx.IsNil(input) {
			return nil, ErrTransformNilBool
		}

		// Handle pointer dereferencing using reflectx
		value := input
		if reflectx.IsPointer(input) {
			if deref, ok := reflectx.Deref(input); ok {
				if reflectx.IsNil(deref) {
					return nil, ErrTransformNilBool
				}
				value = deref
			}
		}

		// Direct type assertion for bool
		if boolVal, ok := value.(bool); ok {
			return fn(boolVal, ctx)
		}

		return nil, fmt.Errorf("%w, received %T", ErrExpectedBool, input)
	})
}

func (b ZodBoolPrefault) Optional() core.ZodType[any, any] {
	return Optional(any(b).(core.ZodType[any, any]))
}

func (b ZodBoolPrefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(b).(core.ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodBoolFromDef creates a ZodBool from definition using unified patterns
func createZodBoolFromDef(def *ZodBoolDef) *ZodBool {
	internals := &ZodBoolInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           def.Checks,
		Isst:             issues.ZodIssueInvalidType{ZodIssueBase: issues.ZodIssueBase{}, Expected: core.ZodTypeBool},
		Bag:              make(map[string]any),
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		boolDef := &ZodBoolDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeBool,
			Checks:     newDef.Checks,
		}
		return any(createZodBoolFromDef(boolDef)).(core.ZodType[any, any])
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		result, err := engine.ParseType[bool](
			payload.Value,
			&internals.ZodTypeInternals,
			core.ZodTypeBool,
			func(v any) (bool, bool) { b, ok := v.(bool); return b, ok },
			func(v any) (*bool, bool) { ptr, ok := v.(*bool); return ptr, ok },
			validateBool,
			ctx,
		)

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

	zodSchema := &ZodBool{internals: internals}
	engine.InitZodType(zodSchema, &def.ZodTypeDef)
	return zodSchema
}

// Bool creates a new bool schema
func Bool(params ...any) *ZodBool {
	def := &ZodBoolDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeBool,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   core.ZodTypeBool,
		Checks: make([]core.ZodCheck, 0),
	}

	schema := createZodBoolFromDef(def)

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

			// Handle abort flag
			if p.Abort {
				if schema.internals.Bag == nil {
					schema.internals.Bag = make(map[string]any)
				}
				schema.internals.Bag["abort"] = true
			}

			// Handle path
			if len(p.Path) > 0 {
				if schema.internals.Bag == nil {
					schema.internals.Bag = make(map[string]any)
				}
				schema.internals.Bag["path"] = p.Path
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
				tgtState.Bag = make(map[string]any)
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
