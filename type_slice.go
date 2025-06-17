package gozod

import (
	"errors"
	"fmt"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodSliceDef defines the configuration for slice validation
type ZodSliceDef struct {
	ZodTypeDef
	Type    string      // "slice"
	Element interface{} // The element schema (type-erased)
}

// ZodSliceInternals contains slice validator internal state
type ZodSliceInternals struct {
	ZodTypeInternals
	Def     *ZodSliceDef
	Element interface{}            // The element schema
	Bag     map[string]interface{} // Runtime configuration
}

// ZodSlice represents a slice validation schema with type safety
type ZodSlice struct {
	internals *ZodSliceInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodSlice) GetInternals() *ZodTypeInternals {
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
				tgtState.Bag = make(map[string]interface{})
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}
	}
}

// Coerce attempts to coerce input to slice format
func (z *ZodSlice) Coerce(input interface{}) (interface{}, bool) {
	return convertToSlice(input)
}

// Parse validates input with smart type inference
func (z *ZodSlice) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Create element-only validator (constraints handled separately)
	validator := func(value []interface{}, checks []ZodCheck, parseCtx *ParseContext) error {
		// Only validate elements, not constraints
		if z.internals.Element != nil {
			_, elementIssues := validateSliceElements(value, z.internals.Element.(ZodType[any, any]), []interface{}{}, nil, input)
			if len(elementIssues) > 0 {
				return &ZodError{Issues: convertRawIssuesToIssues(elementIssues, parseCtx)}
			}
		}
		return nil
	}

	// Create coercer for element validation
	coercer := func(v any) ([]interface{}, bool) {
		if slice, ok := convertToSlice(v); ok {
			return slice, true
		}
		return nil, false
	}

	// Type validation and element validation first
	result, err := parseType[[]interface{}](
		input,
		&z.internals.ZodTypeInternals,
		"slice",
		func(v any) ([]interface{}, bool) {
			if slice, ok := v.([]interface{}); ok {
				return slice, true
			}
			if slice, ok := convertToSlice(v); ok {
				return slice, true
			}
			return nil, false
		},
		func(v any) (*[]interface{}, bool) { ptr, ok := v.(*[]interface{}); return ptr, ok },
		validator,
		coercer,
		parseCtx,
	)

	if err != nil {
		return nil, err
	}

	// Constraint validation (length checks, etc.)
	if len(z.internals.Checks) > 0 {
		var valueToCheck []interface{}
		switch v := result.(type) {
		case []interface{}:
			valueToCheck = v
		case *[]interface{}:
			if v != nil {
				valueToCheck = *v
			} else {
				valueToCheck = []interface{}{} // Empty slice for validation
			}
		default:
			valueToCheck = []interface{}{}
		}

		payload := &ParsePayload{
			Value:  valueToCheck,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(valueToCheck, z.internals.Checks, payload, parseCtx)
		if len(payload.Issues) > 0 {
			return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
		}
	}

	return result, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodSlice) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Element returns the element schema for direct validation
func (z *ZodSlice) Element() ZodType[any, any] {
	if z.internals.Element != nil {
		return z.internals.Element.(ZodType[any, any])
	}
	return Any() // Return generic type if no element schema defined
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Min adds minimum length validation
func (z *ZodSlice) Min(minimum int, params ...SchemaParams) *ZodSlice {
	check := NewZodCheckMinLength(minimum, params...)
	result := AddCheck(z, check)
	return result.(*ZodSlice)
}

// Max adds maximum length validation
func (z *ZodSlice) Max(maximum int, params ...SchemaParams) *ZodSlice {
	check := NewZodCheckMaxLength(maximum, params...)
	result := AddCheck(z, check)
	return result.(*ZodSlice)
}

// Length adds exact length validation
func (z *ZodSlice) Length(length int, params ...SchemaParams) *ZodSlice {
	check := NewZodCheckLengthEquals(length, params...)
	result := AddCheck(z, check)
	return result.(*ZodSlice)
}

// NonEmpty adds non-empty validation (minimum length 1)
func (z *ZodSlice) NonEmpty(params ...SchemaParams) *ZodSlice {
	return z.Min(1, params...)
}

// Refine adds a type-safe refinement check for slice types
func (z *ZodSlice) Refine(fn func([]interface{}) bool, params ...SchemaParams) *ZodSlice {
	result := z.RefineAny(func(v any) bool {
		slice, isNil, err := extractSliceValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true // Let upper logic decide
		}
		return fn(slice)
	}, params...)
	return result.(*ZodSlice)
}

// RefineAny adds custom validation to the slice schema
func (z *ZodSlice) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[interface{}](fn, params...)
	return AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodSlice) Check(fn CheckFn) *ZodSlice {
	check := NewCustom[interface{}](func(v any) bool {
		payload := &ParsePayload{
			Value:  v,
			Issues: make([]ZodRawIssue, 0),
			Path:   make([]interface{}, 0),
		}
		fn(payload)
		return len(payload.Issues) == 0
	})
	result := AddCheck(z, check)
	return result.(*ZodSlice)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a transformation pipeline for slice types
func (z *ZodSlice) Transform(fn func([]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		slice, isNil, err := extractSliceValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilSlice
		}
		return fn(slice, ctx)
	})
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodSlice) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodSlice) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the slice schema optional
func (z *ZodSlice) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable creates a new slice schema that accepts nil values
func (z *ZodSlice) Nilable() ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodSlice) setNilable() ZodType[any, any] {
	cloned := Clone(z, func(def *ZodTypeDef) {
		// Clone operates on ZodTypeDef level
	})
	cloned.(*ZodSlice).internals.Nilable = true
	return cloned
}

// Nullish makes the slice schema both optional and nullable
func (z *ZodSlice) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodSlice) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
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
func (z *ZodSlice) Default(value []interface{}) ZodSliceDefault {
	return ZodSliceDefault{
		&ZodDefault[*ZodSlice]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a function-based default wrapper for slice schema
func (z *ZodSlice) DefaultFunc(fn func() []interface{}) ZodSliceDefault {
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
func (z *ZodSlice) Prefault(value []interface{}) ZodSlicePrefault {
	// Construct Prefault internals
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
func (z *ZodSlice) PrefaultFunc(fn func() []interface{}) ZodSlicePrefault {
	genericFn := func() any { return fn() }

	// Construct Prefault internals
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

	return ZodSlicePrefault{
		&ZodPrefault[*ZodSlice]{
			internals:     internals,
			innerType:     z,
			prefaultValue: []interface{}{},
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// Type-safe wrapper methods for ZodSliceDefault
func (s ZodSliceDefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

func (s ZodSliceDefault) Min(minimum int, params ...SchemaParams) ZodSliceDefault {
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

func (s ZodSliceDefault) Max(maximum int, params ...SchemaParams) ZodSliceDefault {
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

func (s ZodSliceDefault) Length(length int, params ...SchemaParams) ZodSliceDefault {
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

func (s ZodSliceDefault) NonEmpty(params ...SchemaParams) ZodSliceDefault {
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

func (s ZodSliceDefault) Refine(fn func([]interface{}) bool, params ...SchemaParams) ZodSliceDefault {
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

func (s ZodSliceDefault) Transform(fn func([]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		slice, isNil, err := extractSliceValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilSlice
		}
		return fn(slice, ctx)
	})
}

func (s ZodSliceDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodSliceDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// Type-safe wrapper methods for ZodSlicePrefault
func (s ZodSlicePrefault) Min(minimum int, params ...SchemaParams) ZodSlicePrefault {
	newInner := s.innerType.Min(minimum, params...)

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

func (s ZodSlicePrefault) Max(maximum int, params ...SchemaParams) ZodSlicePrefault {
	newInner := s.innerType.Max(maximum, params...)

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

func (s ZodSlicePrefault) Length(length int, params ...SchemaParams) ZodSlicePrefault {
	newInner := s.innerType.Length(length, params...)

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

func (s ZodSlicePrefault) NonEmpty(params ...SchemaParams) ZodSlicePrefault {
	newInner := s.innerType.NonEmpty(params...)

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

func (s ZodSlicePrefault) Refine(fn func([]interface{}) bool, params ...SchemaParams) ZodSlicePrefault {
	newInner := s.innerType.Refine(fn, params...)

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

func (s ZodSlicePrefault) Transform(fn func([]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		slice, isNil, err := extractSliceValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilSlice
		}
		return fn(slice, ctx)
	})
}

func (s ZodSlicePrefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodSlicePrefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodSliceFromDef creates a ZodSlice from definition
func createZodSliceFromDef(def *ZodSliceDef) *ZodSlice {
	internals := &ZodSliceInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Element:          def.Element,
		Bag:              make(map[string]interface{}),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		sliceDef := &ZodSliceDef{
			ZodTypeDef: *newDef,
			Type:       def.Type,
			Element:    def.Element,
		}
		return createZodSliceFromDef(sliceDef)
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodSlice{internals: internals}
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

	schema := &ZodSlice{internals: internals}

	// Initialize the schema with proper error handling support
	initZodType(schema, &def.ZodTypeDef)

	return schema
}

// NewZodSlice creates a new slice schema with element validation
func NewZodSlice(elementSchema ZodType[any, any], params ...SchemaParams) *ZodSlice {
	def := &ZodSliceDef{
		ZodTypeDef: ZodTypeDef{
			Type:   "slice",
			Checks: make([]ZodCheck, 0),
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
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
	}

	return schema
}

// Slice creates a new slice schema (main constructor)
func Slice(elementSchema ZodType[any, any], params ...SchemaParams) *ZodSlice {
	return NewZodSlice(elementSchema, params...)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// extractSliceValue extracts slice value, handling various input types
func extractSliceValue(input any) ([]interface{}, bool, error) {
	if input == nil {
		return nil, true, nil
	}

	switch v := input.(type) {
	case []interface{}:
		return v, false, nil
	case *[]interface{}:
		if v == nil {
			return nil, true, nil
		}
		return *v, false, nil
	default:
		// Try to convert other slice types
		if slice, ok := convertToSlice(input); ok {
			return slice, false, nil
		}
		return nil, false, fmt.Errorf("%w, got %T", ErrExpectedSlice, input)
	}
}
