package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

///////////////////////////
////   CUSTOM TYPE DEFINITIONS ////
///////////////////////////

// ZodCustomDef defines the configuration for custom validation schema
type ZodCustomDef struct {
	core.ZodTypeDef
	Type   core.ZodTypeCode // "custom"
	Check  string           // "custom"
	Path   []any            // PropertyKey[]
	Params map[string]any   // Custom parameters
	Fn     any              // Custom validation function (RefineFn or CheckFn)
	FnType string           // Function type: "refine" or "check"
}

// ZodCustomInternals contains custom validator internal state
type ZodCustomInternals struct {
	core.ZodTypeInternals
	Def  *ZodCustomDef        // Custom validation definition
	Issc *issues.ZodIssueBase // The set of issues this check might throw
	Bag  map[string]any       // Contains Class and other metadata
}

// ZodCustom represents a custom validation schema
type ZodCustom struct {
	internals *ZodCustomInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodCustom) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the custom-specific internals (type-safe access)
func (z *ZodCustom) GetZod() *ZodCustomInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodCustom) CloneFrom(source any) {
	if src, ok := source.(*ZodCustom); ok {
		// Deep copy all custom-specific state
		z.internals = &ZodCustomInternals{
			ZodTypeInternals: src.internals.ZodTypeInternals,
			Def:              src.internals.Def,
			Issc:             src.internals.Issc,
			Bag:              make(map[string]any),
		}
		// Use mapx.Copy for safer bag copying
		if src.internals.Bag != nil {
			z.internals.Bag = mapx.Copy(src.internals.Bag)
		}
	}
}

// Parse validates input with smart type inference using reflectx package
func (z *ZodCustom) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// 1. Unified nil handling using reflectx.IsNil - the only effect of Nilable
	if reflectx.IsNil(input) {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("custom", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return (*any)(nil), nil
	}

	// 2. Execute custom validation - using unified payload pattern
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	payload := &core.ParsePayload{
		Value:  input,
		Issues: make([]core.ZodRawIssue, 0),
		Path:   make([]any, 0),
	}

	// 3. Run custom check function
	result := z.internals.Parse(payload, parseCtx)
	if len(result.Issues) > 0 {
		finalizedIssues := make([]core.ZodIssue, len(result.Issues))
		for i, rawIssue := range result.Issues {
			finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
		}
		return nil, issues.NewZodError(finalizedIssues)
	}

	// 4. Smart type inference: custom validation preserves original input type
	return result.Value, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodCustom) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Transform provides type-safe custom transformation with smart dereferencing support
// Custom validation can handle any type, so Transform accepts any type
func (z *ZodCustom) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Reuse the generic implementation
	return z.TransformAny(fn)
}

// TransformAny flexible version of transformation
// Implements ZodType[any, any] interface: TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any]
func (z *ZodCustom) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Create a pure transform schema
	transform := Transform[any, any](fn)

	// Compose the current schema and the transform schema via pipe
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Refine provides type-safe custom validation
// Supports smart dereferencing: custom validation can handle any type
func (z *ZodCustom) Refine(fn func(any) bool, params ...any) *ZodCustom {
	result := z.RefineAny(fn, params...)
	return result.(*ZodCustom)
}

// RefineAny flexible version of validation that accepts any type
// Implements ZodType[any, any] interface
func (z *ZodCustom) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	// Create a new custom validator and compose it via pipe
	newCustom := Custom(core.RefineFn[any](fn), params...)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(newCustom).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Check adds modern check API
// Implements ZodType[any, any] interface
func (z *ZodCustom) Check(fn core.CheckFn) core.ZodType[any, any] {
	newCustom := Custom(fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(newCustom).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline from this schema to another
func (z *ZodCustom) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

////////////////////////////
////   WRAPPER METHODS  ////
////////////////////////////

// Optional creates an optional version of this schema
func (z *ZodCustom) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable implements ZodType[any, any] interface: Nilable() core.ZodType[any, any]
func (z *ZodCustom) Nilable() core.ZodType[any, any] {
	// Create a new ZodCustom with Nilable flag set instead of using Clone
	newInternals := &ZodCustomInternals{
		ZodTypeInternals: z.internals.ZodTypeInternals,
		Def:              z.internals.Def,
		Bag:              make(map[string]any),
	}

	// Use mapx.Copy for safer bag copying
	if z.internals.Bag != nil {
		newInternals.Bag = mapx.Copy(z.internals.Bag)
	}

	// Set nilable flag using convenience method
	newInternals.SetNilable()

	// Copy the parse function
	newInternals.Parse = z.internals.Parse
	newInternals.Constructor = z.internals.Constructor

	return &ZodCustom{internals: newInternals}
}

// Nullish creates a nullish (optional and nilable) version of this schema
func (z *ZodCustom) Nullish() core.ZodType[any, any] {
	return any(Optional(Nilable(any(z).(core.ZodType[any, any])))).(core.ZodType[any, any])
}

////////////////////////////
////   CONSTRUCTOR FUNCTIONS ////
////////////////////////////

// createZodCustomFromDef creates a custom schema from definition
func createZodCustomFromDef(def *ZodCustomDef) *ZodCustom {
	internals := &ZodCustomInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals("custom"),
		Def:              def,
		Bag:              make(map[string]any),
	}

	// Set up constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		customDef := &ZodCustomDef{
			ZodTypeDef: *newDef,
			Type:       def.Type,
			Check:      def.Check,
			Path:       def.Path,
			Params:     def.Params,
			Fn:         def.Fn,
			FnType:     def.FnType,
		}
		return createZodCustomFromDef(customDef)
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		input := payload.Value
		switch def.FnType {
		case "refine":
			// Try both func(any) bool and core.RefineFn[any] types
			var refineFn func(any) bool
			var ok bool

			if fn, isFunc := def.Fn.(func(any) bool); isFunc {
				refineFn, ok = fn, true
			} else if fn, isRefineFn := def.Fn.(core.RefineFn[any]); isRefineFn {
				refineFn, ok = (func(any) bool)(fn), true
			}

			if ok {
				if !refineFn(input) {
					path := make([]any, len(payload.Path))
					copy(path, payload.Path)
					if def.Path != nil {
						path = append(path, def.Path...)
					}
					options := []func(*issues.ZodRawIssue){
						issues.WithOrigin("custom"),
						issues.WithPath(path),
					}
					if len(def.Params) > 0 {
						options = append(options, issues.WithParams(def.Params))
					}
					issue := issues.NewRawIssue("custom", input, options...)
					issue.Inst = internals

					// Apply custom error message if provided
					if def.Error != nil {
						errorMap := *def.Error
						customMessage := errorMap(issue)
						if customMessage != "" {
							issue.Message = customMessage
						}
					}

					payload.Issues = append(payload.Issues, issue)
				}
			} else {
				// Type assertion failed: treat as validation failure
				path := make([]any, len(payload.Path))
				copy(path, payload.Path)
				if def.Path != nil {
					path = append(path, def.Path...)
				}
				options := []func(*issues.ZodRawIssue){
					issues.WithOrigin("custom"),
					issues.WithPath(path),
				}
				issue := issues.NewRawIssue("custom", input, options...)
				issue.Message = "Invalid function type for refinement"
				issue.Inst = internals
				payload.Issues = append(payload.Issues, issue)
			}
		case "check":
			if checkFn, ok := def.Fn.(core.CheckFn); ok {
				checkFn(payload)
			} else {
				// Invalid check function type: add validation error
				path := make([]any, len(payload.Path))
				copy(path, payload.Path)
				if def.Path != nil {
					path = append(path, def.Path...)
				}
				options := []func(*issues.ZodRawIssue){
					issues.WithOrigin("custom"),
					issues.WithPath(path),
				}
				issue := issues.NewRawIssue("custom", input, options...)
				issue.Message = "Invalid check function type"
				payload.Issues = append(payload.Issues, issue)
			}
		}
		return payload
	}

	zodSchema := &ZodCustom{internals: internals}
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)
	return zodSchema
}

// Custom creates a new custom validation schema
func Custom(fn any, params ...any) *ZodCustom {
	def := &ZodCustomDef{
		ZodTypeDef: core.ZodTypeDef{Type: core.ZodTypeCustom, Checks: make([]core.ZodCheck, 0)},
		Type:       "custom",
		Check:      "custom",
		Params:     make(map[string]any),
	}

	// Handle function parameter
	if fn != nil {
		// Determine function type and store
		switch f := fn.(type) {
		case core.CheckFn:
			def.Fn = f
			def.FnType = "check"
		case func(*core.ParsePayload):
			def.Fn = core.CheckFn(f)
			def.FnType = "check"
		case core.RefineFn[any]:
			def.Fn = f
			def.FnType = "refine"
		case func(any) bool:
			def.Fn = core.RefineFn[any](f)
			def.FnType = "refine"
		default:
			// Try to convert to a generic refine function
			def.Fn = core.RefineFn[any](func(any) bool { return true })
			def.FnType = "refine"
		}
	} else {
		// Default function that always passes
		def.Fn = core.RefineFn[any](func(any) bool { return true })
		def.FnType = "refine"
	}

	schema := createZodCustomFromDef(def)

	// Apply parameters - handle flexible parameter types
	if len(params) > 0 {
		// Try to convert first parameter to SchemaParams
		var param core.SchemaParams
		switch p := params[0].(type) {
		case core.SchemaParams:
			param = p
		case string:
			// If it's a string, treat it as an error message
			param = core.SchemaParams{Error: p}
		default:
			// For other types, create empty SchemaParams
			param = core.SchemaParams{}
		}

		if param.Error != nil {
			// Create error map manually since CreateErrorMap doesn't exist
			var errorMap core.ZodErrorMap
			switch e := param.Error.(type) {
			case string:
				errorMap = func(issue core.ZodRawIssue) string {
					return e
				}
			case core.ZodErrorMap:
				errorMap = e
			case func(core.ZodRawIssue) string:
				// Convert bare function literal to named type for consistent handling
				errorMap = core.ZodErrorMap(e)
			}

			if errorMap != nil {
				def.Error = &errorMap
				schema.internals.Error = &errorMap
			}
		}
		// Handle path parameter
		if param.Path != nil {
			// Manual copy to handle type conversion from []string to []any
			def.Path = make([]any, len(param.Path))
			for i, p := range param.Path {
				def.Path[i] = p
			}
		}
		// Handle params parameter using mapx for safer copying
		if param.Params != nil {
			def.Params = mapx.Copy(param.Params)
		}
	}

	return schema
}

////////////////////////////
////   PUBLIC FUNCTIONS ////
////////////////////////////

// Refine creates a custom refinement validation
func Refine[T any](fn func(T) bool, params ...any) *ZodCustom {
	// Convert typed function to any function
	genericFn := func(input any) bool {
		if typed, ok := input.(T); ok {
			return fn(typed)
		}
		return false // Type mismatch fails validation
	}

	return Custom(core.RefineFn[any](genericFn), params...)
}

// InstanceOf creates validation for Go type checking
func InstanceOf[T any](params ...any) *ZodCustom {
	fn := func(input any) bool {
		_, ok := input.(T)
		return ok
	}

	// Set default error message if not provided
	var finalParams []any
	if len(params) == 0 {
		finalParams = []any{core.SchemaParams{
			Error: "Input is not of the expected type",
		}}
	} else {
		// Try to handle different parameter types
		switch p := params[0].(type) {
		case core.SchemaParams:
			// Check if Error is nil and set default
			if p.Error == nil {
				p.Error = "Input is not of the expected type"
			}
			finalParams = []any{p}
		case string:
			// If it's a string, create SchemaParams with it as error message
			finalParams = []any{core.SchemaParams{Error: p}}
		default:
			// For other types, use default error message
			finalParams = []any{core.SchemaParams{
				Error: "Input is not of the expected type",
			}}
		}
	}

	custom := Custom(core.RefineFn[any](fn), finalParams...)

	// Store simple type information in the bag for debugging
	custom.internals.Bag["instanceOf"] = true

	return custom
}

// Check creates a low-level custom check function
func Check(fn func(*core.ParsePayload), params ...any) *ZodCustom {
	return Custom(core.CheckFn(fn), params...)
}

////////////////////////////
////   HELPER FUNCTIONS ////
////////////////////////////

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodCustom) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

// ==================== CUSTOM DEFAULT/PREFAULT WRAPPER TYPES ====================

// ZodCustomDefault is a Default wrapper for Custom type
// Provides perfect type safety and chainable method support
type ZodCustomDefault struct {
	*ZodDefault[*ZodCustom] // Embed to reuse implementation
}

// ZodCustomPrefault is a Prefault wrapper for Custom type
// Provides perfect type safety and chainable method support
type ZodCustomPrefault struct {
	*ZodPrefault[*ZodCustom] // Embed to reuse implementation
}

// ==================== DEFAULT METHOD IMPLEMENTATIONS ====================

// Default adds a default value to the custom schema
func (z *ZodCustom) Default(value any) ZodCustomDefault {
	return ZodCustomDefault{
		&ZodDefault[*ZodCustom]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the custom schema
func (z *ZodCustom) DefaultFunc(fn func() any) ZodCustomDefault {
	genericFn := func() any { return fn() }
	return ZodCustomDefault{
		&ZodDefault[*ZodCustom]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Refine adds a flexible validation function, returns chainable ZodCustomDefault
func (s ZodCustomDefault) Refine(fn func(any) bool, params ...any) ZodCustomDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodCustomDefault{
		&ZodDefault[*ZodCustom]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation
func (s ZodCustomDefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(fn)
}

// Check converts error returning check into refine and chains
func (s ZodCustomDefault) Check(fn func(*core.ParsePayload) error) ZodCustomDefault {
	refineFn := func(input any) bool {
		payload := &core.ParsePayload{Value: input}
		err := fn(payload)
		return err == nil
	}
	newInner := s.innerType.Refine(refineFn)
	return ZodCustomDefault{
		&ZodDefault[*ZodCustom]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Optional wrapper
func (s ZodCustomDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable wrapper
func (s ZodCustomDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

////////////////////////////
////   PREFAULT METHOD IMPLEMENTATIONS ////
////////////////////////////

// Prefault adds a prefault value to the custom schema
func (z *ZodCustom) Prefault(value any) ZodCustomPrefault {
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     baseInternals.Version,
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

	return ZodCustomPrefault{
		&ZodPrefault[*ZodCustom]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the custom schema
func (z *ZodCustom) PrefaultFunc(fn func() any) ZodCustomPrefault {
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     baseInternals.Version,
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
	genericFn := func() any { return fn() }
	return ZodCustomPrefault{
		&ZodPrefault[*ZodCustom]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// Refine chaining for Prefault wrapper
func (s ZodCustomPrefault) Refine(fn func(any) bool, params ...any) ZodCustomPrefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodCustomPrefault{
		&ZodPrefault[*ZodCustom]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Transform chaining for Prefault wrapper
func (s ZodCustomPrefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(fn)
}

// Check chaining for Prefault wrapper
func (s ZodCustomPrefault) Check(fn func(*core.ParsePayload) error) ZodCustomPrefault {
	refineFn := func(input any) bool {
		payload := &core.ParsePayload{Value: input}
		err := fn(payload)
		return err == nil
	}
	newInner := s.innerType.Refine(refineFn)
	return ZodCustomPrefault{
		&ZodPrefault[*ZodCustom]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Optional wrapper for Prefault
func (s ZodCustomPrefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable wrapper for Prefault
func (s ZodCustomPrefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}
