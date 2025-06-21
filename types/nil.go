package types

import (
	"errors"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for nil transformations
var (
	ErrExpectedNil = errors.New("expected nil type")
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodNilDef defines the configuration for nil validation
type ZodNilDef struct {
	core.ZodTypeDef
	Type   string          // "nil"
	Checks []core.ZodCheck // Nil-specific validation checks
}

// ZodNilInternals contains nil validator internal state
type ZodNilInternals struct {
	core.ZodTypeInternals
	Def  *ZodNilDef                 // Schema definition
	Isst issues.ZodIssueInvalidType // Invalid type issue template
	Bag  map[string]any             // Additional metadata
}

// ZodNil represents a nil validation schema
type ZodNil struct {
	internals *ZodNilInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodNil) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for nil type (only accepts nil)
func (z *ZodNil) Coerce(input any) (any, bool) {
	// Nil type only accepts nil values, no coercion
	return input, reflectx.IsNil(input)
}

// Parse implements intelligent type inference and validation
func (z *ZodNil) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Accept actual nil (including typed nils via reflectx) directly.
	if !reflectx.IsNil(input) {
		raw := issues.CreateInvalidTypeIssue("nil", input)
		final := issues.FinalizeIssue(raw, parseCtx, core.GetConfig())
		return nil, issues.NewZodError([]core.ZodIssue{final})
	}

	// Execute any attached checks (rare for nil).
	if err := validateNil(input, z.internals.Checks, parseCtx); err != nil {
		return nil, err
	}

	// Always return untyped nil interface to satisfy tests.
	return nil, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodNil) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodNil) Check(fn core.CheckFn) core.ZodType[any, any] {
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

// Transform adds data transformation logic to the nil schema
func (z *ZodNil) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny adds flexible data transformation logic
func (z *ZodNil) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
	}
}

// Pipe creates a validation pipeline
func (z *ZodNil) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the nil schema optional
func (z *ZodNil) Optional() core.ZodType[any, any] {
	return Optional(z)
}

// Nilable makes the nil schema nilable
func (z *ZodNil) Nilable() core.ZodType[any, any] {
	return engine.Clone(z, func(def *core.ZodTypeDef) {
		// Nilable is a runtime flag
	}).(*ZodNil)
}

// Nullish makes the nil schema both optional and nilable
func (z *ZodNil) Nullish() core.ZodType[any, any] {
	return Nullish(z)
}

// Refine adds type-safe custom validation logic
func (z *ZodNil) Refine(fn func() bool, params ...any) *ZodNil {
	result := z.RefineAny(func(v any) bool {
		// Use reflectx.IsNil for robust nil checking
		if !reflectx.IsNil(v) {
			return false
		}
		return fn()
	}, params...)
	return result.(*ZodNil)
}

// RefineAny adds flexible custom validation logic
func (z *ZodNil) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Unwrap returns the inner type
func (z *ZodNil) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// DEFAULT METHODS
//////////////////////////

// Default adds a default value to the nil schema, returns ZodType support chain call
func (z *ZodNil) Default(value any) core.ZodType[any, any] {
	return Default(z, value)
}

// DefaultFunc adds a default function to the nil schema, returns ZodType support chain call
func (z *ZodNil) DefaultFunc(fn func() any) core.ZodType[any, any] {
	return DefaultFunc(z, fn)
}

//////////////////////////
// PREFAULT METHODS
//////////////////////////

// Prefault adds a prefault value to the nil schema, returns ZodType support chain call
func (z *ZodNil) Prefault(value any) core.ZodType[any, any] {
	return Prefault(any(z).(core.ZodType[any, any]), value)
}

// PrefaultFunc adds a prefault function to the nil schema, returns ZodType support chain call
func (z *ZodNil) PrefaultFunc(fn func() any) core.ZodType[any, any] {
	return PrefaultFunc(any(z).(core.ZodType[any, any]), fn)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// GetZod returns the nil-specific internals
func (z *ZodNil) GetZod() *ZodNilInternals {
	return z.internals
}

// CloneFrom implements the Cloneable interface
func (z *ZodNil) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodNilInternals }); ok {
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

// createZodNilFromDef creates a ZodNil from definition using unified patterns
func createZodNilFromDef(def *ZodNilDef) *ZodNil {
	internals := &ZodNilInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: "nil"},
		Bag:              make(map[string]any),
	}

	// For nil schema, we inherently allow nil values without error.
	internals.ZodTypeInternals.Nilable = true

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		nilDef := &ZodNilDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeNil,
			Checks:     newDef.Checks,
		}
		return any(createZodNilFromDef(nilDef)).(core.ZodType[any, any])
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		result, err := engine.ParseType[any](
			payload.Value,
			&internals.ZodTypeInternals,
			"nil",
			func(v any) (any, bool) {
				if reflectx.IsNil(v) {
					return nil, true
				}
				return nil, false
			},
			func(v any) (*any, bool) { return nil, false },
			validateNil,
			func(v any) (any, bool) {
				return nil, reflectx.IsNil(v)
			},
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

	zodSchema := &ZodNil{internals: internals}
	engine.InitZodType(zodSchema, &def.ZodTypeDef)
	return zodSchema
}

// Nil creates a new nil schema with unified parameter handling
func Nil(params ...any) *ZodNil {
	def := &ZodNilDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeNil,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   core.ZodTypeNil,
		Checks: make([]core.ZodCheck, 0),
	}

	schema := createZodNilFromDef(def)

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
			if p.Coerce {
				schema.internals.Bag["coerce"] = true
				schema.internals.ZodTypeInternals.Bag["coerce"] = true
			}

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
		}
	}

	return schema
}

// Null is an alias for Nil to align with previous API semantics where both
// names were accepted by the tests. It simply forwards to Nil.
func Null(params ...any) *ZodNil {
	return Nil(params...)
}

//////////////////////////
// VALIDATION FUNCTIONS
//////////////////////////

// validateNil validates nil values with checks
func validateNil(value any, checks []core.ZodCheck, ctx *core.ParseContext) error {
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
