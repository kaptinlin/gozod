package types

import (
	"errors"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// Error definitions for stringbool transformations
var (
	ErrExpectedStringBool = errors.New("expected stringbool type")
)

//////////////////////////////////////////
//////////////////////////////////////////
//////////                      //////////
//////////   ZodStringBool      //////////
//////////                      //////////
//////////////////////////////////////////
//////////////////////////////////////////

// ZodStringBoolDef defines the configuration for stringbool validation
type ZodStringBoolDef struct {
	core.ZodTypeDef
	Type   string   // "stringbool"
	Truthy []string // Truthy string values
	Falsy  []string // Falsy string values
	Case   string   // "sensitive" or "insensitive"
}

// ZodStringBoolInternals contains stringbool validator internal state
type ZodStringBoolInternals struct {
	core.ZodTypeInternals
	Def    *ZodStringBoolDef          // Schema definition
	Isst   issues.ZodIssueInvalidType // Invalid type issue template
	Truthy map[string]struct{}        // Truthy values set for fast lookup
	Falsy  map[string]struct{}        // Falsy values set for fast lookup
	Bag    map[string]any             // Runtime configuration
}

// ZodStringBool represents a string-to-boolean validation schema
type ZodStringBool struct {
	internals *ZodStringBoolInternals
}

// StringBoolOptions provides configuration for stringbool schema creation
type StringBoolOptions struct {
	Truthy []string // Values that evaluate to true
	Falsy  []string // Values that evaluate to false
	Case   string   // "sensitive" or "insensitive"
}

//////////////////////////////////////////
// Core Interface Implementation
//////////////////////////////////////////

// GetInternals returns the internal state for framework use
func (z *ZodStringBool) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce attempts to coerce input to string for stringbool validation
func (z *ZodStringBool) Coerce(input any) (any, bool) {
	if str, err := coerce.ToString(input); err == nil {
		return str, true
	}
	return nil, false
}

// Parse validates and converts string input to boolean output using smart type inference
func (z *ZodStringBool) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Use engine.ParseType with custom checkers/coercers that adhere to zod behaviour
	result, err := engine.ParseType[bool](
		input,
		&z.internals.ZodTypeInternals,
		"stringbool",
		// =========================
		// Type Checker
		// =========================
		func(v any) (bool, bool) {
			// Accept raw string
			if str, ok := v.(string); ok {
				if res, ok2 := z.tryStringToBool(str); ok2 {
					return res, true
				}
				// Un-recognised string – signal failure and let validator phase turn into value error.
				return false, false
			}
			// Accept *string (smart inference – return bool, not pointer)
			if ptr, ok := v.(*string); ok && ptr != nil {
				if res, ok2 := z.tryStringToBool(*ptr); ok2 {
					return res, true
				}
				return false, false
			}
			return false, false
		},
		// =========================
		// Pointer Checker
		// (keep pointer to bool only; do NOT match *string to avoid pointer return)
		// =========================
		func(v any) (*bool, bool) {
			if ptr, ok := v.(*bool); ok {
				return ptr, true
			}
			return nil, false
		},
		// =========================
		// Validator – run attached checks (e.g., Refine)
		// =========================
		func(value bool, checks []core.ZodCheck, ctx *core.ParseContext) error {
			if len(checks) == 0 {
				return nil
			}
			payload := &core.ParsePayload{
				Value:  value,
				Issues: make([]core.ZodRawIssue, 0),
			}
			engine.RunChecksOnValue(value, checks, payload, ctx)
			if len(payload.Issues) > 0 {
				finalized := make([]core.ZodIssue, len(payload.Issues))
				for i, raw := range payload.Issues {
					finalized[i] = issues.FinalizeIssue(raw, ctx, core.GetConfig())
				}
				return issues.NewZodError(finalized)
			}
			return nil
		},
		// =========================
		// Coercer – enabled only for non-bool values
		// =========================
		func(v any) (bool, bool) {
			// Exclude bools from coercion behaviour (see tests)
			switch v.(type) {
			case bool, *bool:
				return false, false
			}
			if str, err := coerce.ToString(v); err == nil {
				return z.tryStringToBool(str)
			}
			return false, false
		},
		parseCtx,
	)

	// Post-processing error mapping (convert invalid_type to invalid_value when input was string/*string)
	if err != nil {
		var zErr *issues.ZodError
		if errors.As(err, &zErr) {
			for i, iss := range zErr.Issues {
				if iss.Code == core.InvalidType {
					switch input.(type) {
					case string, *string:
						zErr.Issues[i].Code = core.InvalidValue
						// Re-evaluate message with schema error map if available
						if z.internals.Error != nil {
							msg := (*z.internals.Error)(core.ZodRawIssue{Code: core.InvalidValue})
							if msg != "" {
								zErr.Issues[i].Message = msg
							}
						}
					}
				}
			}
		}
		return nil, err
	}

	return result, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodStringBool) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////////////////////
// String to Bool Conversion
//////////////////////////////////////////

// stringToBool performs the core string-to-bool conversion
func (z *ZodStringBool) stringToBool(value string) bool {
	normalizedValue := value
	if z.internals.Def.Case == "insensitive" {
		normalizedValue = strings.ToLower(value)
	}

	// Check truthy values
	if _, exists := z.internals.Truthy[normalizedValue]; exists {
		return true
	}

	// Check falsy values (default to false for unrecognized values)
	return false
}

// tryStringToBool tries to convert string to bool, returns (result, success)
func (z *ZodStringBool) tryStringToBool(value string) (bool, bool) {
	normalizedValue := value
	if z.internals.Def.Case == "insensitive" {
		normalizedValue = strings.ToLower(value)
	}

	// Check truthy values
	if _, exists := z.internals.Truthy[normalizedValue]; exists {
		return true, true
	}

	// Check falsy values
	if _, exists := z.internals.Falsy[normalizedValue]; exists {
		return false, true
	}

	// Unrecognized value
	return false, false
}

//////////////////////////////////////////
// Validation Methods
//////////////////////////////////////////

// Refine adds type-safe custom validation logic to the stringbool schema
func (z *ZodStringBool) Refine(fn func(bool) bool, params ...any) *ZodStringBool {
	check := checks.NewCustom[bool](fn, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodStringBool)
}

// RefineAny adds flexible custom validation logic to the stringbool schema
func (z *ZodStringBool) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

//////////////////////////////////////////
// Transform Methods
//////////////////////////////////////////

// Transform creates a type-safe transformation of stringbool values
func (z *ZodStringBool) Transform(fn func(bool, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(v any, ctx *core.RefinementContext) (any, error) {
		if b, ok := v.(bool); ok {
			return fn(b, ctx)
		}
		// For non-bool values, convert to bool first using stringToBool
		if str, ok := v.(string); ok {
			boolVal := z.stringToBool(str)
			return fn(boolVal, ctx)
		}
		return nil, errors.New("cannot transform non-bool value")
	})
}

// TransformAny adds data transformation with any input type
func (z *ZodStringBool) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
	}
}

// Pipe connects this schema to another schema
func (z *ZodStringBool) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
	}
}

//////////////////////////////////////////
// Modifier Methods
//////////////////////////////////////////

// Optional makes the stringbool optional
func (z *ZodStringBool) Optional() core.ZodType[any, any] {
	return Optional(z)
}

// Nilable creates a nilable stringbool schema
func (z *ZodStringBool) Nilable() core.ZodType[any, any] {
	return Nilable(z)
}

// Nullish creates a nullish (optional + nilable) stringbool schema
func (z *ZodStringBool) Nullish() core.ZodType[any, any] {
	return Nullish(z)
}

//////////////////////////////////////////
// Wrapper Types (Simplified)
//////////////////////////////////////////

// Default adds a default value to the stringbool schema
func (z *ZodStringBool) Default(value bool) core.ZodType[any, any] {
	return Default(z, value)
}

// DefaultFunc adds a default function to the stringbool schema
func (z *ZodStringBool) DefaultFunc(fn func() bool) core.ZodType[any, any] {
	return DefaultFunc(z, func() any { return fn() })
}

// Prefault adds a prefault value to the stringbool schema
func (z *ZodStringBool) Prefault(value bool) core.ZodType[any, any] {
	return Prefault(any(z).(core.ZodType[any, any]), value)
}

// PrefaultFunc adds a prefault function to the stringbool schema
func (z *ZodStringBool) PrefaultFunc(fn func() bool) core.ZodType[any, any] {
	return PrefaultFunc(any(z).(core.ZodType[any, any]), func() any { return fn() })
}

//////////////////////////////////////////
// Constructor Functions
//////////////////////////////////////////

// createZodStringBoolFromDef creates a ZodStringBool from definition with unified patterns
func createZodStringBoolFromDef(def *ZodStringBoolDef) *ZodStringBool {
	internals := &ZodStringBoolInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Isst:             issues.ZodIssueInvalidType{Expected: "stringbool"},
		Truthy:           make(map[string]struct{}),
		Falsy:            make(map[string]struct{}),
		Bag:              make(map[string]any),
	}

	// Build lookup maps for efficient validation
	for _, value := range def.Truthy {
		normalizedValue := value
		if def.Case == "insensitive" {
			normalizedValue = strings.ToLower(value)
		}
		internals.Truthy[normalizedValue] = struct{}{}
	}

	for _, value := range def.Falsy {
		normalizedValue := value
		if def.Case == "insensitive" {
			normalizedValue = strings.ToLower(value)
		}
		internals.Falsy[normalizedValue] = struct{}{}
	}

	// Set up constructor for AddCheck functionality
	internals.Constructor = func(def *core.ZodTypeDef) core.ZodType[any, any] {
		stringBoolDef := &ZodStringBoolDef{
			ZodTypeDef: *def,
			Type:       "stringbool",
			Truthy:     internals.Def.Truthy,
			Falsy:      internals.Def.Falsy,
			Case:       internals.Def.Case,
		}
		return createZodStringBoolFromDef(stringBoolDef)
	}

	zodSchema := &ZodStringBool{internals: internals}

	// Initialize the schema using unified initZodType from engine
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

// StringBool creates a new stringbool schema. It can be called in three forms:
//
//	StringBool()                       -> use default options
//	StringBool(opts)                   -> custom truthy/falsy options
//	StringBool(opts, params) / StringBool(params...) for additional SchemaParams.
//
// The implementation inspects the first argument: if it is a *StringBoolOptions
// it is treated as the options struct; otherwise the first argument is assumed
// to be schema parameters.
func StringBool(args ...any) *ZodStringBool {
	var options *StringBoolOptions
	var params []any

	if len(args) > 0 {
		// Special-case nil placeholder for options (to align with StringBool(nil, core.SchemaParams{...}))
		if args[0] == nil {
			// No options provided, remaining args are treated as params
			if len(args) > 1 {
				params = args[1:]
			}
		} else if opt, ok := args[0].(*StringBoolOptions); ok {
			options = opt
			if len(args) > 1 {
				params = args[1:]
			}
		} else {
			params = args
		}
	}

	// Default values
	truthy := []string{"true", "1", "yes", "on", "y", "enabled"}
	falsy := []string{"false", "0", "no", "off", "n", "disabled"}
	caseMode := "insensitive"

	// Apply custom options
	if options != nil {
		if len(options.Truthy) > 0 {
			truthy = options.Truthy
		}
		if len(options.Falsy) > 0 {
			falsy = options.Falsy
		}
		if options.Case != "" {
			caseMode = options.Case
		}
	}

	def := &ZodStringBoolDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "stringbool",
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   "stringbool",
		Truthy: truthy,
		Falsy:  falsy,
		Case:   caseMode,
	}

	schema := createZodStringBoolFromDef(def)

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

//////////////////////////////////////////
// Utility Functions
//////////////////////////////////////////

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodStringBool) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}
