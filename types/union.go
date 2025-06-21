package types

import (
	"errors"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
)

// Error definitions for union transformations
var (
	ErrTransformNilUnion = errors.New("cannot transform nil union value")
)

// =============================================================================
// CORE TYPE DEFINITIONS
// =============================================================================

// ZodUnionDef defines the configuration for union validation
type ZodUnionDef struct {
	core.ZodTypeDef
	Type    string                   // "union"
	Options []core.ZodType[any, any] // Union options to validate against
}

// ZodUnionInternals contains union validator internal state
type ZodUnionInternals struct {
	core.ZodTypeInternals
	Def     *ZodUnionDef                // Schema definition
	Options []core.ZodType[any, any]    // Union options
	Isst    issues.ZodIssueInvalidUnion // Invalid union issue template
	Bag     map[string]any              // Additional metadata
}

// ZodUnion represents a union validation schema for alternative types
type ZodUnion struct {
	internals *ZodUnionInternals
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodUnion) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse validates input with smart type inference
func (z *ZodUnion) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// Handle Optional/Nilable first
	if input == nil {
		if z.internals.Nilable || z.internals.Optional {
			return nil, nil
		}
	}
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Try to match each option first
	for _, option := range z.internals.Options {
		result, err := option.Parse(input, parseCtx)
		if err == nil {
			// Found matching option, perform Refine validation
			if len(z.internals.Checks) > 0 {
				payload := &core.ParsePayload{
					Value:  result,
					Issues: make([]core.ZodRawIssue, 0),
				}
				engine.RunChecksOnValue(result, z.internals.Checks, payload, parseCtx)
				if len(payload.Issues) > 0 {
					return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.Issues, parseCtx))
				}
			}
			return result, nil
		}
	}

	// All options failed to match, create invalid_union error
	// Collect error information from all options
	var allErrors []string
	for _, option := range z.internals.Options {
		_, err := option.Parse(input, parseCtx)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}

	// Create invalid_union error
	issue := CreateInvalidUnionIssue(input, nil)
	if len(allErrors) > 0 {
		issue.Properties = map[string]any{
			"unionErrors": allErrors,
		}
	}

	finalIssue := issues.FinalizeIssue(issue, parseCtx, core.GetConfig())
	return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// MustParse validates the input value and panics on failure
func (z *ZodUnion) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// TRANSFORM METHODS
// =============================================================================

// Transform provides type-safe union transformation
func (z *ZodUnion) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny flexible version of transformation - same implementation as Transform
func (z *ZodUnion) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Optional makes the union optional
func (z *ZodUnion) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable modifier: only changes nil handling, not type inference logic
func (z *ZodUnion) Nilable() core.ZodType[any, any] {
	return engine.Clone(z, func(def *core.ZodTypeDef) {}).(*ZodUnion).setNilable()
}

// setNilable sets the Nilable flag - internal method
func (z *ZodUnion) setNilable() core.ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Pipe creates a validation pipeline
func (z *ZodUnion) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Options provides access to the union's option schemas
func (z *ZodUnion) Options() []core.ZodType[any, any] {
	return z.internals.Options
}

// GetZod returns the internal state for framework usage
func (z *ZodUnion) GetZod() *ZodUnionInternals {
	return z.internals
}

// Refine adds type-safe custom validation logic to the union schema
func (z *ZodUnion) Refine(fn func(any) bool, params ...any) *ZodUnion {
	result := z.RefineAny(func(v any) bool {
		val, isNil, err := extractUnionValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true
		}
		return fn(val)
	}, params...)
	return result.(*ZodUnion)
}

// RefineAny adds flexible custom validation logic to the union schema
func (z *ZodUnion) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Check adds modern validation using direct payload access
func (z *ZodUnion) Check(fn core.CheckFn) *ZodUnion {
	check := checks.NewCustom[any](fn, core.SchemaParams{})
	result := engine.AddCheck(z, check)
	return result.(*ZodUnion)
}

//////////////////////////////////
////   TYPE-SAFE DEFAULT WRAPPER ////
//////////////////////////////////

// ZodUnionDefault is the Default wrapper for Union type
type ZodUnionDefault struct {
	*ZodDefault[*ZodUnion] // Embed concrete pointer for method promotion
}

// Parse ensures correct validation of internal type
func (s ZodUnionDefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

//////////////////////////////////
////   DEFAULT METHODS ////
//////////////////////////////////

// Default adds a default value to the union schema, returns ZodUnionDefault
func (z *ZodUnion) Default(value any) ZodUnionDefault {
	return ZodUnionDefault{
		&ZodDefault[*ZodUnion]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the union schema, returns ZodUnionDefault
func (z *ZodUnion) DefaultFunc(fn func() any) ZodUnionDefault {
	return ZodUnionDefault{
		&ZodDefault[*ZodUnion]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

//////////////////////////////////
////   UNIONDEFAULT CHAIN METHODS ////
//////////////////////////////////

// Refine refine the union with given function, return ZodUnionDefault
func (s ZodUnionDefault) Refine(fn func(any) bool, params ...any) ZodUnionDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodUnionDefault{
		&ZodDefault[*ZodUnion]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform transform the union with given function, return ZodType
func (s ZodUnionDefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		unionVal, isNil, err := extractUnionValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilUnion
		}
		return fn(unionVal, ctx)
	})
}

// Optional make the union optional
func (s ZodUnionDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable make the union nilable
func (s ZodUnionDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

//////////////////////////////////
////   TYPE-SAFE PREFAULT WRAPPER ////
//////////////////////////////////

// ZodUnionPrefault is the Prefault wrapper for Union type
type ZodUnionPrefault struct {
	*ZodPrefault[*ZodUnion] // Embed generic wrapper
}

// Parse ensures correct validation of internal type
func (u ZodUnionPrefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return u.ZodPrefault.Parse(input, ctx...)
}

//////////////////////////////////
////   PREFAULT METHODS ////
//////////////////////////////////

// Prefault adds a prefault value to the union schema, returns ZodUnionPrefault
func (z *ZodUnion) Prefault(value any) ZodUnionPrefault {
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

	return ZodUnionPrefault{
		&ZodPrefault[*ZodUnion]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the union schema, returns ZodUnionPrefault
func (z *ZodUnion) PrefaultFunc(fn func() any) ZodUnionPrefault {
	genericFn := func() any { return fn() }

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

	return ZodUnionPrefault{
		&ZodPrefault[*ZodUnion]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

//////////////////////////////////
////   UNIONPREFAULT CHAIN METHODS ////
//////////////////////////////////

// Refine refine the union with given function, return ZodUnionPrefault
func (u ZodUnionPrefault) Refine(fn func(any) bool, params ...any) ZodUnionPrefault {
	newInner := u.innerType.Refine(fn, params...)

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

	return ZodUnionPrefault{
		&ZodPrefault[*ZodUnion]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: u.prefaultValue,
			prefaultFunc:  u.prefaultFunc,
			isFunction:    u.isFunction,
		},
	}
}

// Transform transform the union with given function, return ZodType
func (u ZodUnionPrefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return u.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Union type: directly pass value, let transform function handle type checking
		return fn(input, ctx)
	})
}

// Optional make the union optional
func (u ZodUnionPrefault) Optional() core.ZodType[any, any] {
	return Optional(any(u).(core.ZodType[any, any]))
}

// Nilable make the union nilable
func (u ZodUnionPrefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(u).(core.ZodType[any, any]))
}

////////////////////////////
////   INTERNAL METHODS ////
////////////////////////////

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodUnion) CloneFrom(source any) {
	if src, ok := source.(*ZodUnion); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]any)
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy union-specific fields
		if src.internals.Options != nil {
			z.internals.Options = make([]core.ZodType[any, any], len(src.internals.Options))
			copy(z.internals.Options, src.internals.Options)
		}
	}
}

// createZodUnionFromDef creates a ZodUnion from definition
func createZodUnionFromDef(def *ZodUnionDef) *ZodUnion {
	internals := &ZodUnionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Options:          def.Options, // Set options from definition
		Isst:             issues.ZodIssueInvalidUnion{ZodIssueBase: issues.ZodIssueBase{}, Errors: [][]core.ZodIssue{}},
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		unionDef := &ZodUnionDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeUnion,
			Options:    def.Options, // Preserve original options
		}
		return any(createZodUnionFromDef(unionDef)).(core.ZodType[any, any])
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		results := make([]*core.ParsePayload, 0, len(def.Options))

		for _, option := range def.Options {
			optionPayload := &core.ParsePayload{
				Value:  payload.Value,
				Issues: make([]core.ZodRawIssue, 0),
			}

			resultPayload, err := option.Parse(payload.Value, ctx)
			if err != nil {
				// Parse failed - add to results for error reporting
				var zodErr *issues.ZodError
				if errors.As(err, &zodErr) {
					// Convert ZodIssue to ZodRawIssue
					for _, issue := range zodErr.Issues {
						rawIssue := core.ZodRawIssue{
							Code:    issue.Code,
							Input:   issue.Input,
							Path:    issue.Path,
							Message: issue.Message,
						}
						optionPayload.Issues = append(optionPayload.Issues, rawIssue)
					}
				} else {
					// Generic error - convert to ZodRawIssue
					issue := issues.NewRawIssue(issues.InvalidType, payload.Value)
					optionPayload.Issues = append(optionPayload.Issues, issue)
				}
				results = append(results, optionPayload)
				continue
			}

			// Parse succeeded - return immediately
			payload.Value = resultPayload

			// Run custom checks if any exist
			if customChecks, exists := internals.Bag["customChecks"].([]core.ZodCheck); exists {
				engine.RunChecksOnValue(payload.Value, customChecks, payload, ctx)
			}

			return payload
		}

		// No option succeeded - handle union error
		return handleUnionResults(results, payload, internals)
	}

	zodSchema := &ZodUnion{internals: internals}

	// Initialize the schema with proper error handling support
	engine.InitZodType(any(zodSchema).(core.ZodType[any, any]), &def.ZodTypeDef)

	return zodSchema
}

// Union creates a new union schema
func Union(options []core.ZodType[any, any], params ...any) *ZodUnion {
	def := &ZodUnionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeUnion,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:    core.ZodTypeUnion,
		Options: options,
	}

	schema := createZodUnionFromDef(def)

	// Apply schema parameters
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
			// Store coerce flag in bag for parseUnionCore to access
			if p.Coerce {
				schema.internals.Bag["coerce"] = true
			}

			// Handle schema-level error mapping using utility function
			if p.Error != nil {
				errorMap := issues.CreateErrorMap(p.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}
		}
	}

	return schema
}

// WithUnionErrors creates an option to set union errors
func WithUnionErrors(errors [][]core.ZodRawIssue) func(*core.ZodRawIssue) {
	return func(issue *core.ZodRawIssue) {
		if issue.Properties == nil {
			issue.Properties = make(map[string]any)
		}
		issue.Properties["errors"] = errors
	}
}

// handleUnionResults processes union validation results
func handleUnionResults(results []*core.ParsePayload, final *core.ParsePayload, internals *ZodUnionInternals) *core.ParsePayload {
	// Check if any result succeeded
	for _, result := range results {
		if len(result.Issues) == 0 {
			final.Value = result.Value
			return final
		}
	}

	// No result succeeded - create union error
	issue := CreateInvalidUnionIssue(final.Value, results, func(issue *core.ZodRawIssue) {
		issue.Inst = internals
	})
	final.Issues = append(final.Issues, issue)
	return final
}

// CreateInvalidUnionIssue creates an invalid union issue
func CreateInvalidUnionIssue(input any, results []*core.ParsePayload, options ...func(*core.ZodRawIssue)) core.ZodRawIssue {
	// Extract all error issues from failed results
	allErrors := make([][]core.ZodRawIssue, 0, len(results))
	for _, result := range results {
		if len(result.Issues) > 0 {
			allErrors = append(allErrors, result.Issues)
		}
	}

	issue := issues.NewRawIssue(
		issues.InvalidUnion,
		input,
		WithUnionErrors(allErrors),
	)

	// Apply additional options
	for _, opt := range options {
		opt(&issue)
	}

	return issue
}

// extractUnionValue extracts union value - simplified helper method
// Returns: (union value, is nil pointer, error)
func extractUnionValue(input any) (any, bool, error) {
	// Union types accept any value - the union logic will determine validity
	// We don't need to pre-filter here, just check for nil pointer cases
	if input == nil {
		return nil, true, nil
	}

	// For union types, we pass through the value as-is
	// The union parsing logic will handle type checking
	return input, false, nil
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodUnion) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}
