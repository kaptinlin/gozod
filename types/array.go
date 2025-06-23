package types

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodArrayDef defines the configuration for fixed-length array schemas
type ZodArrayDef struct {
	core.ZodTypeDef
	Type   core.ZodTypeCode         // "array"
	Items  []core.ZodType[any, any] // Fixed array element schemas
	Checks []core.ZodCheck          // Array-specific validation checks
}

// ZodArrayInternals represents the internal state of an array schema
type ZodArrayInternals struct {
	core.ZodTypeInternals
	Def    *ZodArrayDef             // Array definition
	Items  []core.ZodType[any, any] // Fixed array element schemas
	Checks []core.ZodCheck          // Validation checks
	Bag    map[string]any           // Runtime configuration
}

// ZodArray represents a fixed-length array schema
type ZodArray struct {
	internals *ZodArrayInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodArray) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse validates and parses input with array element validation
func (z *ZodArray) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// create validator with element validation
	validator := func(value []any, checks []core.ZodCheck, ctx *core.ParseContext) error {
		return validateArray(value, checks, z, ctx)
	}

	// use reflectx extractor directly as type checker
	typeChecker := reflectx.ExtractSlice

	return engine.ParseType[[]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeArray,
		typeChecker,
		func(v any) (*[]any, bool) { ptr, ok := v.(*[]any); return ptr, ok },
		validator,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodArray) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Length adds exact length validation using checks package
func (z *ZodArray) Length(length int, params ...any) *ZodArray {
	clone := z.clone()
	check := checks.Length(length, params...)
	clone.internals.Checks = append(clone.internals.Checks, check)
	clone.internals.ZodTypeInternals.Checks = clone.internals.Checks
	return clone
}

// Min adds minimum length validation using checks package
func (z *ZodArray) Min(minimum int, params ...any) *ZodArray {
	clone := z.clone()
	check := checks.MinLength(minimum, params...)
	clone.internals.Checks = append(clone.internals.Checks, check)
	clone.internals.ZodTypeInternals.Checks = clone.internals.Checks
	return clone
}

// Max adds maximum length validation using checks package
func (z *ZodArray) Max(maximum int, params ...any) *ZodArray {
	clone := z.clone()
	check := checks.MaxLength(maximum, params...)
	clone.internals.Checks = append(clone.internals.Checks, check)
	clone.internals.ZodTypeInternals.Checks = clone.internals.Checks
	return clone
}

// NonEmpty adds non-empty validation using checks package
func (z *ZodArray) NonEmpty(params ...any) *ZodArray {
	return z.Min(1, params...)
}

// Element returns the schema for the element at the given index
func (z *ZodArray) Element(index int) core.ZodType[any, any] {
	if index >= 0 && index < len(z.internals.Items) {
		return z.internals.Items[index]
	}
	return nil
}

// Items returns all element schemas
func (z *ZodArray) Items() []core.ZodType[any, any] {
	result := make([]core.ZodType[any, any], len(z.internals.Items))
	copy(result, z.internals.Items)
	return result
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe array transformation
func (z *ZodArray) Transform(fn func([]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		arr, ok := reflectx.ExtractSlice(input)
		if !ok {
			return nil, errors.New("cannot transform non-array value")
		}
		return fn(arr, ctx)
	})
}

// TransformAny provides flexible transformation with any input/output types
func (z *ZodArray) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodArray) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the array schema optional
func (z *ZodArray) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable makes the array schema nilable
func (z *ZodArray) Nilable() core.ZodType[any, any] {
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {})
	cloned.(*ZodArray).internals.SetNilable()
	return cloned
}

// Nullable is an alias for Nilable (for TypeScript compatibility)
func (z *ZodArray) Nullable() core.ZodType[any, any] {
	return z.Nilable()
}

// Nullish makes the array schema nullish (optional AND nilable)
func (z *ZodArray) Nullish() core.ZodType[any, any] {
	return z.Optional().Nilable()
}

// clone creates a copy of the array
func (z *ZodArray) clone() *ZodArray {
	newInternals := &ZodArrayInternals{
		ZodTypeInternals: z.internals.ZodTypeInternals,
		Def:              z.internals.Def,
		Items:            make([]core.ZodType[any, any], len(z.internals.Items)),
		Checks:           make([]core.ZodCheck, len(z.internals.Checks)),
		Bag:              make(map[string]any),
	}

	// Deep copy Items
	copy(newInternals.Items, z.internals.Items)

	// Deep copy Checks
	copy(newInternals.Checks, z.internals.Checks)

	// Ensure embedded ZodTypeInternals.Checks points to the cloned slice
	newInternals.ZodTypeInternals.Checks = newInternals.Checks

	// Deep copy Bag
	for k, v := range z.internals.Bag {
		newInternals.Bag[k] = v
	}

	return &ZodArray{internals: newInternals}
}

// Refine adds custom validation logic to the array schema
func (z *ZodArray) Refine(fn func([]any) bool, params ...any) *ZodArray {
	// Create a new array with refinement check
	clone := z.clone()
	// Create a custom check using the checks package
	check := checks.NewCustom[[]any](fn, params...)
	clone.internals.Checks = append(clone.internals.Checks, check)
	clone.internals.ZodTypeInternals.Checks = clone.internals.Checks
	return clone
}

// RefineAny adds custom validation with any input type
func (z *ZodArray) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	clone := z.clone()
	check := checks.NewCustom[any](fn, params...)
	clone.internals.Checks = append(clone.internals.Checks, check)
	clone.internals.ZodTypeInternals.Checks = clone.internals.Checks
	return any(clone).(core.ZodType[any, any])
}

// Check adds a custom validation check to the array schema
func (z *ZodArray) Check(fn core.CheckFn) *ZodArray {
	clone := z.clone()
	check := checks.NewCustom[any](fn)
	clone.internals.Checks = append(clone.internals.Checks, check)
	clone.internals.ZodTypeInternals.Checks = clone.internals.Checks
	return clone
}

// GetDef returns the schema definition
func (z *ZodArray) GetDef() *ZodArrayDef {
	return z.internals.Def
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodArray) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

////////////////////////////
// WRAPPER TYPES (DEFAULT & PREFAULT)
////////////////////////////

// ZodArrayDefault embeds ZodDefault with concrete pointer for method promotion
type ZodArrayDefault struct {
	*ZodDefault[*ZodArray] // Embed concrete pointer, allows method promotion
}

// ZodArrayPrefault embeds ZodPrefault with concrete pointer for method promotion
type ZodArrayPrefault struct {
	*ZodPrefault[*ZodArray] // Embed concrete pointer, allows method promotion
}

////////////////////////////
// DEFAULT METHODS
////////////////////////////

// Default creates a default wrapper for array schema with static value
func (z *ZodArray) Default(value []any) ZodArrayDefault {
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a default wrapper for array schema with function-provided value
func (z *ZodArray) DefaultFunc(fn func() []any) ZodArrayDefault {
	genericFn := func() any { return fn() }
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

////////////////////////////
// PREFAULT METHODS
////////////////////////////

// Prefault creates a prefault wrapper for array schema with static fallback value
func (z *ZodArray) Prefault(value []any) ZodArrayPrefault {
	// Construct Prefault internals, copying base internals but changing Type
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

	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a prefault wrapper for array schema with function fallback value
func (z *ZodArray) PrefaultFunc(fn func() []any) ZodArrayPrefault {
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

	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			internals:     internals,
			innerType:     z,
			prefaultValue: []any{},
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

////////////////////////////
// ZodArrayDefault – chainable, type-safe methods
////////////////////////////

// Parse delegates to embedded ZodDefault.Parse for convenience
func (s ZodArrayDefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

func (s ZodArrayDefault) Min(minimum int, params ...any) ZodArrayDefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodArrayDefault) Max(maximum int, params ...any) ZodArrayDefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodArrayDefault) Length(length int, params ...any) ZodArrayDefault {
	newInner := s.innerType.Length(length, params...)
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodArrayDefault) NonEmpty(params ...any) ZodArrayDefault {
	newInner := s.innerType.NonEmpty(params...)
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodArrayDefault) Refine(fn func([]any) bool, params ...any) ZodArrayDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodArrayDefault) Transform(fn func([]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		slice, ok := reflectx.ExtractSlice(input)
		if !ok {
			return nil, fmt.Errorf("expected array type, got %T", input)
		}
		return fn(slice, ctx)
	})
}

func (s ZodArrayDefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodArrayDefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

////////////////////////////
// ZodArrayPrefault – chainable, type-safe methods
////////////////////////////

func (s ZodArrayPrefault) Min(minimum int, params ...any) ZodArrayPrefault {
	newInner := s.innerType.Min(minimum, params...)
	// Recreate internals to keep Prefault wrapper
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
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodArrayPrefault) Max(maximum int, params ...any) ZodArrayPrefault {
	newInner := s.innerType.Max(maximum, params...)
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
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodArrayPrefault) Length(length int, params ...any) ZodArrayPrefault {
	newInner := s.innerType.Length(length, params...)
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
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodArrayPrefault) NonEmpty(params ...any) ZodArrayPrefault {
	newInner := s.innerType.NonEmpty(params...)
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
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodArrayPrefault) Refine(fn func([]any) bool, params ...any) ZodArrayPrefault {
	newInner := s.innerType.Refine(fn, params...)
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
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodArrayPrefault) Transform(fn func([]any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		slice, ok := reflectx.ExtractSlice(input)
		if !ok {
			return nil, fmt.Errorf("expected array type, got %T", input)
		}
		return fn(slice, ctx)
	})
}

func (s ZodArrayPrefault) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodArrayPrefault) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

//////////////////////////
// FACTORY FUNCTIONS
//////////////////////////

// createZodArrayFromDef creates a ZodArray instance from a definition
func createZodArrayFromDef(def *ZodArrayDef) *ZodArray {
	internals := &ZodArrayInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:     "array",
			Optional: false,
			Nilable:  false,
		},
		Def:    def,
		Items:  def.Items,
		Checks: make([]core.ZodCheck, 0),
		Bag:    make(map[string]any),
	}

	// Always ensure we have at least one check to trigger validation
	hasDefaultCheck := false

	// Copy existing checks first
	if len(def.Checks) > 0 {
		internals.Checks = make([]core.ZodCheck, len(def.Checks))
		copy(internals.Checks, def.Checks)

		// Check if we already have a default check (to avoid duplicates)
		for _, check := range def.Checks {
			if checkInternals, ok := check.(*core.ZodCheckInternals); ok {
				if checkInternals.Def != nil && checkInternals.Def.Check == "array_default" {
					hasDefaultCheck = true
					break
				}
			}
		}
	}

	// Add default check only if we don't have one
	if !hasDefaultCheck {
		defaultCheck := checks.NewCustom[any](func(v any) bool {
			return true // Always pass, actual validation happens in validateArray
		})
		// Convert to interface and append
		internals.Checks = append(internals.Checks, core.ZodCheck(defaultCheck))
	}

	// Set ZodTypeInternals checks to match
	internals.ZodTypeInternals.Checks = make([]core.ZodCheck, len(internals.Checks))
	copy(internals.ZodTypeInternals.Checks, internals.Checks)

	array := &ZodArray{
		internals: internals,
	}

	// Set constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		arrayDef := &ZodArrayDef{
			ZodTypeDef: *newDef,
			Type:       "array",
			Items:      def.Items,     // Preserve original items
			Checks:     newDef.Checks, // Use new checks from AddCheck
		}
		return any(createZodArrayFromDef(arrayDef)).(core.ZodType[any, any])
	}

	// Set parse function for validation
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		value, err := array.Parse(payload.GetValue(), ctx)
		if err != nil {
			// Add issues to payload on validation error
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					// Convert ZodError to RawIssue using standardized converter
					rawIssue := issues.ConvertZodIssueToRaw(issue)
					rawIssue.Path = issue.Path
					payload.AddIssue(rawIssue)
				}
			}
		} else {
			payload.SetValue(value)
		}
		return payload
	}

	return array
}

// Array creates a new fixed-length array schema with the given element schemas
func Array(args ...any) *ZodArray {
	if len(args) == 0 {
		// Empty array schema
		return createZodArrayFromDef(&ZodArrayDef{
			ZodTypeDef: core.ZodTypeDef{Type: "array"},
			Type:       "array",
			Items:      []core.ZodType[any, any]{},
			Checks:     []core.ZodCheck{},
		})
	}

	var items []core.ZodType[any, any]
	var params core.SchemaParams

	// Parse arguments to find items and optional parameters
	for _, arg := range args {
		switch v := arg.(type) {
		case core.ZodType[any, any]:
			items = append(items, v)
		case []core.ZodType[any, any]:
			items = append(items, v...)
		case core.SchemaParams:
			params = v
		default:
			// Try to convert to ZodType if possible
			if zodType, ok := v.(core.ZodType[any, any]); ok {
				items = append(items, zodType)
			}
		}
	}

	def := &ZodArrayDef{
		ZodTypeDef: core.ZodTypeDef{
			Type: "array",
		},
		Type:   "array",
		Items:  items,
		Checks: []core.ZodCheck{},
	}

	// Apply schema parameters if provided
	if params.Error != nil {
		if errorStr, ok := params.Error.(string); ok {
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return errorStr
			})
			def.Error = &errorMap
		}
	}

	zodSchema := createZodArrayFromDef(def)
	if params.Error != nil {
		if errorStr, ok := params.Error.(string); ok {
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return errorStr
			})
			def.Error = &errorMap
			zodSchema.internals.Error = &errorMap
		}
	}

	return zodSchema
}

// GetZod returns the internals for debugging purposes
func (z *ZodArray) GetZod() *ZodArrayInternals {
	return z.internals
}

// CloneFrom copies configuration from another schema
func (z *ZodArray) CloneFrom(source any) {
	if sourceArray, ok := source.(*ZodArray); ok {
		// Deep copy the internals
		z.internals = &ZodArrayInternals{
			ZodTypeInternals: sourceArray.internals.ZodTypeInternals,
			Def:              sourceArray.internals.Def,
			Items:            make([]core.ZodType[any, any], len(sourceArray.internals.Items)),
			Checks:           make([]core.ZodCheck, len(sourceArray.internals.Checks)),
			Bag:              make(map[string]any),
		}

		// Copy items and checks
		copy(z.internals.Items, sourceArray.internals.Items)
		copy(z.internals.Checks, sourceArray.internals.Checks)

		// Copy bag
		for k, v := range sourceArray.internals.Bag {
			z.internals.Bag[k] = v
		}
	}
}

// validateArray validates array elements against their schemas
func validateArray(value []any, checks []core.ZodCheck, z *ZodArray, ctx *core.ParseContext) error {
	// Create payload for nested validation
	payload := core.NewParsePayload(value)

	// First validate array length against expected number of items (always enforce fixed length semantics)
	{
		expectedLength := len(z.internals.Items)
		actualLength := len(value)

		if actualLength != expectedLength {
			if actualLength > expectedLength {
				rawIssue := issues.CreateTooBigIssue(expectedLength, true, utils.GetLengthableOrigin(value), value)
				payload.AddIssue(rawIssue)
			} else {
				rawIssue := issues.CreateTooSmallIssue(expectedLength, true, utils.GetLengthableOrigin(value), value)
				payload.AddIssue(rawIssue)
			}
		} else {
			// Validate each element against corresponding item schema (only if length matches)
			for i, element := range value {
				if i < len(z.internals.Items) {
					itemSchema := z.internals.Items[i]
					if itemSchema != nil {
						// Validate element directly
						_, err := itemSchema.Parse(element, ctx)
						if err != nil {
							// Convert error to issues and add to payload
							var zodErr *issues.ZodError
							if errors.As(err, &zodErr) {
								for _, issue := range zodErr.Issues {
									// Convert ZodError to RawIssue using standardized converter
									rawIssue := issues.ConvertZodIssueToRaw(issue)

									// Prepend element index to path for array element errors
									newPath := make([]any, 0, len(issue.Path)+1)
									newPath = append(newPath, i)
									newPath = append(newPath, issue.Path...)
									rawIssue.Path = newPath
									rawIssue.Input = value // Use original array as input

									payload.AddIssue(rawIssue)
								}
							} else {
								// Create generic error for element
								rawIssue := core.ZodRawIssue{
									Code:    issues.InvalidType,
									Message: err.Error(),
									Path:    []any{i},
									Input:   value,
								}
								payload.AddIssue(rawIssue)
							}
						}
					}
				}
			}
		}
	}

	// Validate using the checks system (for Min, Max, etc.)
	for _, check := range checks {
		// Get the check internals through the interface method
		checkInternals := check.GetZod()
		if checkInternals != nil && checkInternals.Check != nil {
			// Call the check function with payload
			checkInternals.Check(payload)
		}
	}

	if payload.HasIssues() {
		convertedIssues := issues.ConvertRawIssuesToIssues(payload.GetIssues(), ctx)
		return issues.NewZodError(convertedIssues)
	}

	return nil
}
