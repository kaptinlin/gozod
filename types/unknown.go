package types

import (
	"errors"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
)

// Error definitions for unknown transformations
var (
	ErrExpectedUnknown = errors.New("expected unknown type")
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodUnknownDef defines the configuration for unknown validation
type ZodUnknownDef struct {
	core.ZodTypeDef
	Type   core.ZodTypeCode // "unknown"
	Checks []core.ZodCheck  // Unknown-specific validation checks
}

// ZodUnknownInternals contains unknown validator internal state
type ZodUnknownInternals struct {
	core.ZodTypeInternals
	Def  *ZodUnknownDef             // Schema definition
	Isst issues.ZodIssueInvalidType // Invalid type issue template
	Bag  map[string]any             // Additional metadata
}

// ZodUnknown represents the concrete implementation of unknown validation schema
type ZodUnknown struct {
	internals *ZodUnknownInternals
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// GetInternals returns the internal state of the schema
func (z *ZodUnknown) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for unknown type (accepts any value)
func (z *ZodUnknown) Coerce(input any) (any, bool) {
	// Unknown type accepts any value without conversion
	return input, true
}

// Parse implements intelligent type inference and validation
func (z *ZodUnknown) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return engine.ParseType[any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnknown,
		func(v any) (any, bool) { return v, true },                       // Direct acceptance extractor
		func(v any) (*any, bool) { ptr, ok := v.(*any); return ptr, ok }, // Pointer extractor
		validateUnknown,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodUnknown) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodUnknown) Check(fn core.CheckFn) core.ZodType[any, any] {
	check := checks.NewCustom[any](func(v any) bool {
		payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
		fn(payload)
		return len(payload.Issues) == 0
	})
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform adds data transformation logic to the unknown schema
func (z *ZodUnknown) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny adds flexible data transformation logic
func (z *ZodUnknown) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodUnknown) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the unknown schema optional
func (z *ZodUnknown) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the unknown schema nilable
func (z *ZodUnknown) Nilable() core.ZodType[any, any] {
	cloned := engine.Clone(z, nil).(*ZodUnknown)
	cloned.internals.Nilable = true
	return any(cloned).(core.ZodType[any, any])
}

// Nullish makes the unknown schema both optional and nilable
func (z *ZodUnknown) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Refine adds type-safe custom validation logic
func (z *ZodUnknown) Refine(fn func(any) bool, params ...any) *ZodUnknown {
	result := z.RefineAny(fn, params...)
	return result.(*ZodUnknown)
}

// RefineAny adds flexible custom validation logic
func (z *ZodUnknown) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Unwrap returns the inner type
func (z *ZodUnknown) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodUnknownDefault represents an unknown schema with default value
type ZodUnknownDefault struct {
	*ZodDefault[*ZodUnknown]
}

// ZodUnknownPrefault represents an unknown schema with prefault value
type ZodUnknownPrefault struct {
	*ZodPrefault[*ZodUnknown]
}

//////////////////////////
// DEFAULT METHODS
//////////////////////////

// Default adds a default value to the unknown schema, returns ZodUnknownDefault support chain call
func (z *ZodUnknown) Default(value any) ZodUnknownDefault {
	return ZodUnknownDefault{
		&ZodDefault[*ZodUnknown]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the unknown schema, returns ZodUnknownDefault support chain call
func (z *ZodUnknown) DefaultFunc(fn func() any) ZodUnknownDefault {
	return ZodUnknownDefault{
		&ZodDefault[*ZodUnknown]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

//////////////////////////
// UNKNOWNDEFAULT CHAINING METHODS
//////////////////////////

// Refine adds a flexible validation function to the unknown schema, returns ZodUnknownDefault support chain call
func (s ZodUnknownDefault) Refine(fn func(any) bool, params ...any) ZodUnknownDefault {
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

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodUnknownDefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(fn)
}

// Optional adds an optional check to the unknown schema, returns ZodType support chain call
func (s ZodUnknownDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the unknown schema, returns ZodType support chain call
func (s ZodUnknownDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

//////////////////////////
// PREFAULT METHODS
//////////////////////////

// Prefault adds a prefault value to the unknown schema, returns ZodUnknownPrefault support chain call
func (z *ZodUnknown) Prefault(value any) ZodUnknownPrefault {
	// Construct Prefault's internals, Type = "prefault", copy internal type's checks/coerce/optional/nilable
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
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

// PrefaultFunc adds a prefault function to the unknown schema, returns ZodUnknownPrefault support chain call
func (z *ZodUnknown) PrefaultFunc(fn func() any) ZodUnknownPrefault {
	// Construct Prefault's internals, Type = "prefault", copy internal type's checks/coerce/optional/nilable
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
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

//////////////////////////
// UNKNOWNPREFAULT CHAINING METHODS
//////////////////////////

// Refine adds a flexible validation function to the unknown schema, returns ZodUnknownPrefault support chain call
func (u ZodUnknownPrefault) Refine(fn func(any) bool, params ...any) ZodUnknownPrefault {
	newInner := u.innerType.Refine(fn, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
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

// Transform adds data conversion, returns a generic ZodType support conversion pipeline
func (u ZodUnknownPrefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use embedded ZodPrefault's TransformAny method
	return u.TransformAny(fn)
}

// Optional adds an optional check to the unknown schema, returns ZodType support chain call
func (u ZodUnknownPrefault) Optional() core.ZodType[any, any] {
	// Wrap current ZodUnknownPrefault instance, keep Prefault logic
	return Optional(any(u).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the unknown schema, returns ZodType support chain call
func (u ZodUnknownPrefault) Nilable() core.ZodType[any, any] {
	// Wrap current ZodUnknownPrefault instance, keep Prefault logic
	return Nilable(any(u).(core.ZodType[any, any]))
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// GetZod returns the unknown-specific internals
func (z *ZodUnknown) GetZod() *ZodUnknownInternals {
	return z.internals
}

// CloneFrom implements the Cloneable interface
func (z *ZodUnknown) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodUnknownInternals }); ok {
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

		if len(srcState.ZodTypeInternals.Bag) > 0 {
			if tgtState.ZodTypeInternals.Bag == nil {
				tgtState.ZodTypeInternals.Bag = make(map[string]any)
			}
			for key, value := range srcState.ZodTypeInternals.Bag {
				tgtState.ZodTypeInternals.Bag[key] = value
			}
		}
	}
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodUnknownFromDef creates a ZodUnknown from definition using unified patterns
func createZodUnknownFromDef(def *ZodUnknownDef) *ZodUnknown {
	internals := &ZodUnknownInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeUnknown},
		Bag:              make(map[string]any),
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		unknownDef := &ZodUnknownDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeUnknown,
			Checks:     newDef.Checks,
		}
		return any(createZodUnknownFromDef(unknownDef)).(core.ZodType[any, any])
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		result, err := engine.ParseType[any](
			payload.Value,
			&internals.ZodTypeInternals,
			"unknown",
			func(v any) (any, bool) { return v, true },
			func(v any) (*any, bool) { ptr, ok := v.(*any); return ptr, ok },
			validateUnknown,
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

	zodSchema := &ZodUnknown{internals: internals}
	engine.InitZodType(zodSchema, &def.ZodTypeDef)
	return zodSchema
}

// Unknown creates an unknown validation schema
func Unknown(params ...any) *ZodUnknown {
	def := &ZodUnknownDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeUnknown,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   core.ZodTypeUnknown,
		Checks: make([]core.ZodCheck, 0),
	}

	schema := createZodUnknownFromDef(def)

	// Apply schema parameters using the same pattern as string.go
	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return p
			})
			def.Error = &errorMap
			schema.internals.Error = &errorMap
		case core.SchemaParams:
			// Handle core.SchemaParams

			if p.Error != nil {
				// Handle string error messages by converting to ZodErrorMap
				if errStr, ok := p.Error.(string); ok {
					errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
						return errStr
					})
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				} else if errorMap, ok := p.Error.(core.ZodErrorMap); ok {
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				}
			}

			if p.Description != "" {
				schema.internals.Bag["description"] = p.Description
			}
			if p.Abort {
				schema.internals.Bag["abort"] = true
			}
			if len(p.Path) > 0 {
				schema.internals.Bag["path"] = p.Path
			}
			// Ensure any custom Params are preserved for introspection.
			if len(p.Params) > 0 {
				// Lazily initialize inner Bag if it was somehow nil.
				if schema.internals.Bag == nil {
					schema.internals.Bag = make(map[string]any)
				}
				if schema.internals.ZodTypeInternals.Bag == nil {
					schema.internals.ZodTypeInternals.Bag = make(map[string]any)
				}

				for k, v := range p.Params {
					schema.internals.Bag[k] = v                  // Store on outer Bag for tests
					schema.internals.ZodTypeInternals.Bag[k] = v // Mirror on embedded internals for consistency
				}
			}
		}
	}

	return schema
}

//////////////////////////
// VALIDATION FUNCTIONS
//////////////////////////

// validateUnknown validates unknown values with checks
func validateUnknown(value any, checks []core.ZodCheck, ctx *core.ParseContext) error {
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
