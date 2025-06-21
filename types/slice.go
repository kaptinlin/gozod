package types

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for slice transformations
var (
	ErrTransformNilSlice = errors.New("cannot transform nil slice value")
	ErrExpectedSlice     = errors.New("expected slice type")
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodSliceDef defines the configuration for slice validation
type ZodSliceDef struct {
	core.ZodTypeDef
	Type    string // "slice"
	Element any    // The element schema (type-erased)
}

// ZodSliceInternals contains slice validator internal state
type ZodSliceInternals struct {
	core.ZodTypeInternals
	Def     *ZodSliceDef
	Element any            // The element schema
	Bag     map[string]any // Runtime configuration
}

// ZodSlice represents a slice validation schema with type safety
type ZodSlice struct {
	internals *ZodSliceInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodSlice) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the slice-specific internals for framework usage
func (z *ZodSlice) GetZod() *ZodSliceInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodSlice) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodSliceInternals }); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		// Copy Element schema
		if srcState.Element != nil {
			tgtState.Element = srcState.Element
		}

		// Copy Bag state
		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]any)
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}
	}
}

// Coerce attempts to coerce input to slice format using existing utilities
func (z *ZodSlice) Coerce(input any) (any, bool) {
	result, err := coerce.ToSlice(input)
	return result, err == nil
}

// Parse validates input with smart type inference
func (z *ZodSlice) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Create element-only validator (elements handled separately now)
	validator := func(value []any, checks []core.ZodCheck, parsectx *core.ParseContext) error {
		// Element validation is now handled separately to collect all issues
		return nil
	}

	// Create coercer for element validation using existing utilities
	coercer := func(v any) ([]any, bool) {
		result, err := coerce.ToSlice(v)
		return result, err == nil
	}

	// Collect all issues from both type/element validation and constraint validation
	var allRawIssues []core.ZodRawIssue

	// Type validation and element validation first
	result, err := engine.ParseType[[]any](
		input,
		&z.internals.ZodTypeInternals,
		"slice",
		func(v any) ([]any, bool) {
			if slice, ok := v.([]any); ok {
				return slice, true
			}
			// Use reflectx.ExtractSlice instead of custom logic
			return reflectx.ExtractSlice(v)
		},
		func(v any) (*[]any, bool) { ptr, ok := v.(*[]any); return ptr, ok },
		validator,
		coercer,
		parseCtx,
	)

	// Extract element validation issues even if type validation succeeded
	var valueToCheck []any
	if result != nil {
		switch v := result.(type) {
		case []any:
			valueToCheck = v
		case *[]any:
			if v != nil {
				valueToCheck = *v
			} else {
				valueToCheck = []any{} // Empty slice for validation
			}
		default:
			valueToCheck = []any{}
		}
	} else if err != nil {
		// If result is nil due to type conversion failure, try to extract slice from input
		if slice, ok := reflectx.ExtractSlice(input); ok {
			valueToCheck = slice
		}
	}

	// Collect element validation issues using existing function from array.go
	if z.internals.Element != nil && len(valueToCheck) > 0 {
		_, elementIssues := validateSliceElements(valueToCheck, z.internals.Element.(core.ZodType[any, any]), []any{}, parseCtx, input)
		allRawIssues = append(allRawIssues, elementIssues...)
	}

	// Run checks
	if len(z.internals.Checks) > 0 {
		payload := &core.ParsePayload{Value: valueToCheck, Issues: make([]core.ZodRawIssue, 0)}
		engine.RunChecksOnValue(valueToCheck, z.internals.Checks, payload, parseCtx)
		allRawIssues = append(allRawIssues, payload.Issues...)
	}

	// Also include any type conversion errors
	if err != nil {
		var zodErr *issues.ZodError
		if errors.As(err, &zodErr) {
			for _, issue := range zodErr.Issues {
				properties := make(map[string]any)
				if issue.Expected != "" {
					properties["expected"] = issue.Expected
				}
				if issue.Received != "" {
					properties["received"] = issue.Received
				}
				rawIssue := core.ZodRawIssue{
					Code:       issue.Code,
					Input:      issue.Input,
					Properties: properties,
					Path:       issue.Path,
					Message:    issue.Message,
				}
				allRawIssues = append(allRawIssues, rawIssue)
			}
		} else if zodErrorInterface, ok := err.(interface {
			GetIssues() []core.ZodIssue
		}); ok {
			// Convert ZodIssue to ZodRawIssue
			for _, issue := range zodErrorInterface.GetIssues() {
				properties := make(map[string]any)
				if issue.Expected != "" {
					properties["expected"] = issue.Expected
				}
				if issue.Received != "" {
					properties["received"] = issue.Received
				}
				rawIssue := core.ZodRawIssue{
					Code:       issue.Code,
					Input:      issue.Input,
					Properties: properties,
					Path:       issue.Path,
					Message:    issue.Message,
				}
				allRawIssues = append(allRawIssues, rawIssue)
			}
		}
	}

	// If we have any issues, return them
	if len(allRawIssues) > 0 {
		finalizedIssues := make([]core.ZodIssue, len(allRawIssues))
		for i, rawIssue := range allRawIssues {
			finalizedIssues[i] = issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
		}
		return nil, issues.NewZodError(finalizedIssues)
	}

	// If there were no issues but parseType failed, return the original error
	if err != nil {
		return nil, err
	}

	return result, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodSlice) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Element returns the element schema for direct validation
func (z *ZodSlice) Element() core.ZodType[any, any] {
	if z.internals.Element != nil {
		return z.internals.Element.(core.ZodType[any, any])
	}
	return Any() // Return generic type if no element schema defined
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Min adds minimum length validation using existing checks
func (z *ZodSlice) Min(minimum int, params ...any) *ZodSlice {
	check := checks.MinLength(minimum, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodSlice)
}

// Max adds maximum length validation using existing checks
func (z *ZodSlice) Max(maximum int, params ...any) *ZodSlice {
	check := checks.MaxLength(maximum, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodSlice)
}

// Length adds exact length validation using existing checks
func (z *ZodSlice) Length(length int, params ...any) *ZodSlice {
	check := checks.Length(length, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodSlice)
}

// NonEmpty adds non-empty validation (minimum length 1)
func (z *ZodSlice) NonEmpty(params ...any) *ZodSlice {
	return z.Min(1, params...)
}

// Refine adds a type-safe refinement check for slice types
func (z *ZodSlice) Refine(fn func([]any) bool, params ...any) *ZodSlice {
	result := z.RefineAny(func(v any) bool {
		// Use reflectx.ExtractSlice instead of private method
		slice, ok := reflectx.ExtractSlice(v)
		if !ok {
			return false
		}
		if slice == nil {
			return true // Let upper logic decide
		}
		return fn(slice)
	}, params...)
	return result.(*ZodSlice)
}

// RefineAny adds custom validation to the slice schema
func (z *ZodSlice) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodSlice) Check(fn core.CheckFn) *ZodSlice {
	check := checks.NewCustom[any](func(v any) bool {
		payload := &core.ParsePayload{
			Value:  v,
			Issues: make([]core.ZodRawIssue, 0),
			Path:   make([]any, 0),
		}
		fn(payload)
		return len(payload.Issues) == 0
	})
	result := engine.AddCheck(z, check)
	return result.(*ZodSlice)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a transformation pipeline for slice types
func (z *ZodSlice) Transform(fn func([]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx.ExtractSlice instead of private method
		slice, ok := reflectx.ExtractSlice(input)
		if !ok {
			return nil, fmt.Errorf("%w, got %T", ErrExpectedSlice, input)
		}
		if slice == nil {
			return nil, ErrTransformNilSlice
		}
		return fn(slice, ctx)
	})
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodSlice) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodSlice) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the slice schema optional
func (z *ZodSlice) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable creates a new slice schema that accepts nil values
func (z *ZodSlice) Nilable() core.ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodSlice) setNilable() core.ZodType[any, any] {
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// Clone operates on ZodTypeDef level
	})
	cloned.(*ZodSlice).internals.Nilable = true
	return cloned
}

// Nullish makes the slice schema both optional and nullable
func (z *ZodSlice) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodSlice) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodSliceDefault embeds ZodDefault with concrete pointer for method promotion
type ZodSliceDefault struct {
	*ZodDefault[*ZodSlice] // Embed concrete pointer, allows method promotion
}

// ZodSlicePrefault embeds ZodPrefault with concrete pointer for method promotion
type ZodSlicePrefault struct {
	*ZodPrefault[*ZodSlice] // Embed concrete pointer, allows method promotion
}

// Default creates a default wrapper for slice schema
func (z *ZodSlice) Default(value []any) ZodSliceDefault {
	return ZodSliceDefault{
		&ZodDefault[*ZodSlice]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a function-based default wrapper for slice schema
func (z *ZodSlice) DefaultFunc(fn func() []any) ZodSliceDefault {
	genericFn := func() any { return fn() }
	return ZodSliceDefault{
		&ZodDefault[*ZodSlice]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Prefault creates a prefault wrapper for slice schema
func (z *ZodSlice) Prefault(value []any) ZodSlicePrefault {
	// Construct Prefault internals
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

	return ZodSlicePrefault{
		&ZodPrefault[*ZodSlice]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a function-based prefault wrapper for slice schema
func (z *ZodSlice) PrefaultFunc(fn func() []any) ZodSlicePrefault {
	genericFn := func() any { return fn() }

	// Construct Prefault internals
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

	return ZodSlicePrefault{
		&ZodPrefault[*ZodSlice]{
			internals:     internals,
			innerType:     z,
			prefaultValue: []any{},
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// Type-safe wrapper methods for ZodSliceDefault
func (s ZodSliceDefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

func (s ZodSliceDefault) Min(minimum int, params ...any) ZodSliceDefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodSliceDefault{
		&ZodDefault[*ZodSlice]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodSliceDefault) Max(maximum int, params ...any) ZodSliceDefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodSliceDefault{
		&ZodDefault[*ZodSlice]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodSliceDefault) Length(length int, params ...any) ZodSliceDefault {
	newInner := s.innerType.Length(length, params...)
	return ZodSliceDefault{
		&ZodDefault[*ZodSlice]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodSliceDefault) NonEmpty(params ...any) ZodSliceDefault {
	newInner := s.innerType.NonEmpty(params...)
	return ZodSliceDefault{
		&ZodDefault[*ZodSlice]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodSliceDefault) Refine(fn func([]any) bool, params ...any) ZodSliceDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodSliceDefault{
		&ZodDefault[*ZodSlice]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodSliceDefault) Transform(fn func([]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx.ExtractSlice instead of private method
		slice, ok := reflectx.ExtractSlice(input)
		if !ok {
			return nil, fmt.Errorf("%w, got %T", ErrExpectedSlice, input)
		}
		if slice == nil {
			return nil, ErrTransformNilSlice
		}
		return fn(slice, ctx)
	})
}

func (s ZodSliceDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodSliceDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

// Type-safe wrapper methods for ZodSlicePrefault
func (s ZodSlicePrefault) Min(minimum int, params ...any) ZodSlicePrefault {
	newInner := s.innerType.Min(minimum, params...)

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

	return ZodSlicePrefault{
		&ZodPrefault[*ZodSlice]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodSlicePrefault) Max(maximum int, params ...any) ZodSlicePrefault {
	newInner := s.innerType.Max(maximum, params...)

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

	return ZodSlicePrefault{
		&ZodPrefault[*ZodSlice]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodSlicePrefault) Length(length int, params ...any) ZodSlicePrefault {
	newInner := s.innerType.Length(length, params...)

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

	return ZodSlicePrefault{
		&ZodPrefault[*ZodSlice]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodSlicePrefault) NonEmpty(params ...any) ZodSlicePrefault {
	newInner := s.innerType.NonEmpty(params...)

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

	return ZodSlicePrefault{
		&ZodPrefault[*ZodSlice]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodSlicePrefault) Refine(fn func([]any) bool, params ...any) ZodSlicePrefault {
	newInner := s.innerType.Refine(fn, params...)

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

	return ZodSlicePrefault{
		&ZodPrefault[*ZodSlice]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodSlicePrefault) Transform(fn func([]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx.ExtractSlice instead of private method
		slice, ok := reflectx.ExtractSlice(input)
		if !ok {
			return nil, fmt.Errorf("%w, got %T", ErrExpectedSlice, input)
		}
		if slice == nil {
			return nil, ErrTransformNilSlice
		}
		return fn(slice, ctx)
	})
}

func (s ZodSlicePrefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodSlicePrefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodSliceFromDef creates a ZodSlice from definition
func createZodSliceFromDef(def *ZodSliceDef) *ZodSlice {
	internals := &ZodSliceInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Element:          def.Element,
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		sliceDef := &ZodSliceDef{
			ZodTypeDef: *newDef,
			Type:       def.Type,
			Element:    def.Element,
		}
		return createZodSliceFromDef(sliceDef)
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		schema := &ZodSlice{internals: internals}
		result, err := schema.Parse(payload.Value, ctx)
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

	zodSchema := &ZodSlice{internals: internals}

	// Initialize the schema with proper error handling support
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

// Slice creates a new slice schema with element validation
func Slice(elementSchema core.ZodType[any, any], params ...core.SchemaParams) *ZodSlice {
	def := &ZodSliceDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   "slice",
			Checks: make([]core.ZodCheck, 0),
		},
		Type:    "slice",
		Element: elementSchema,
	}

	schema := createZodSliceFromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		param := params[0]

		// Store coerce flag in bag
		if param.Coerce {
			schema.internals.Bag["coerce"] = true
			schema.internals.ZodTypeInternals.Bag["coerce"] = true
		}

		// Handle schema-level error mapping
		if param.Error != nil {
			if errorStr, ok := param.Error.(string); ok {
				errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
					return errorStr
				})
				def.Error = &errorMap
				schema.internals.Error = &errorMap
			}
		}
	}

	return schema
}

// validateSliceElements validates each element in a slice using the given element schema
func validateSliceElements(slice []any, elementSchema any, basePath []any, ctx any, originalInput any) ([]any, []core.ZodRawIssue) {
	if elementSchema == nil {
		return slice, nil
	}

	// Type assertion to get the ZodType interface
	zodType, ok := elementSchema.(interface {
		Parse(input any, ctx ...*core.ParseContext) (any, error)
	})
	if !ok {
		// If it's not a proper ZodType, return as-is
		return slice, nil
	}

	validatedElements := make([]any, len(slice))
	var allIssues []core.ZodRawIssue

	// Convert ctx to proper type if provided
	var parseCtx *core.ParseContext
	if ctx != nil {
		if pc, ok := ctx.(*core.ParseContext); ok {
			parseCtx = pc
		}
	}

	for i, element := range slice {
		elementPath := make([]any, 0, len(basePath)+1)
		elementPath = append(elementPath, basePath...)
		elementPath = append(elementPath, i)

		// Parse element with schema
		validatedElement, err := zodType.Parse(element, parseCtx)
		if err != nil {
			// Handle parsing errors
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := core.ZodRawIssue{
						Code:    issue.Code,
						Message: issue.Message,
						Path:    append(elementPath, issue.Path...),
						Input:   originalInput,
					}
					allIssues = append(allIssues, rawIssue)
				}
			} else {
				// Create a generic error
				rawIssue := core.ZodRawIssue{
					Code:    "invalid_type",
					Message: err.Error(),
					Path:    elementPath,
					Input:   originalInput,
				}
				allIssues = append(allIssues, rawIssue)
			}
			validatedElements[i] = element // Keep original value on error
		} else {
			validatedElements[i] = validatedElement
		}
	}

	return validatedElements, allIssues
}
