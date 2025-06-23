package types

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

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

// ZodLiteralDef defines the configuration for literal value validation
type ZodLiteralDef struct {
	core.ZodTypeDef
	Type   core.ZodTypeCode // "literal"
	Values []any            // Allowed literal values
}

// ZodLiteralInternals represents the internal state for literal validation
type ZodLiteralInternals struct {
	core.ZodTypeInternals
	Def     *ZodLiteralDef             // Literal definition
	Values  map[any]struct{}           // Set of allowed values
	Pattern *regexp.Regexp             // Regex pattern for validation
	Isst    issues.ZodIssueInvalidType // Issue type for validation errors
	Bag     map[string]any             // Runtime configuration bag
}

// ZodLiteral represents a literal value schema that validates against specific values
type ZodLiteral struct {
	internals *ZodLiteralInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodLiteral) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse validates input with smart type inference
func (z *ZodLiteral) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Use reflectx.IsNil for unified nil handling logic
	if reflectx.IsNil(input) {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("literal", input)
			rawIssue.Inst = z
			finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return inferTypedNilFromLiterals(z.internals.Def.Values), nil
	}

	// Smart type inference: input type determines output type
	// Check direct value matching
	if isValidLiteral(input, z.internals) {
		// Run validation checks - use unified runChecksOnValue
		if len(z.internals.Checks) > 0 {
			// Use constructor instead of direct struct literal to respect private fields
			payload := core.NewParsePayload(input)
			engine.RunChecksOnValue(input, z.internals.Checks, payload, parseCtx)
			if len(payload.GetIssues()) > 0 {
				return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.GetIssues(), parseCtx))
			}
		}
		return input, nil // Preserve original type
	}

	// Handle pointer types using reflectx.Deref
	if derefValue, ok := reflectx.Deref(input); ok {
		// Check if input is a nil pointer
		if reflectx.IsNilPointer(input) {
			if !z.internals.Nilable {
				rawIssue := issues.CreateInvalidTypeIssue("literal", input)
				rawIssue.Inst = z
				finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
				return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
			}
			return inferTypedNilFromLiterals(z.internals.Def.Values), nil
		}

		// Check if dereferenced value is in literal values
		if isValidLiteral(derefValue, z.internals) {
			// Run validation checks
			if len(z.internals.Checks) > 0 {
				// Use constructor instead of direct struct literal to respect private fields
				payload := core.NewParsePayload(derefValue)
				engine.RunChecksOnValue(derefValue, z.internals.Checks, payload, parseCtx)
				if len(payload.GetIssues()) > 0 {
					return nil, issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.GetIssues(), parseCtx))
				}
			}
			return input, nil // Return original pointer (preserve type)
		}
	}

	// No coercion for literal type (non-primitive)

	// Create invalid value issue instead of invalid literal issue
	rawIssue := issues.CreateInvalidValueIssue(z.internals.Def.Values, input)
	rawIssue.Inst = z
	finalIssue := issues.FinalizeIssue(rawIssue, parseCtx, core.GetConfig())
	return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// MustParse validates the input value and panics on failure
func (z *ZodLiteral) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe literal transformation with smart dereferencing support
func (z *ZodLiteral) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny flexible version of transformation - provides backward compatibility
func (z *ZodLiteral) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Create pure Transform
	transform := Transform[any, any](fn)

	// Return pipe(inst, transform)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),         // Type conversion
		out: any(transform).(core.ZodType[any, any]), // Type conversion
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe chains this literal schema with another schema
func (z *ZodLiteral) Pipe(next core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: next,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the literal schema optional
func (z *ZodLiteral) Optional() core.ZodType[any, any] {
	return any(Optional(any(z).(core.ZodType[any, any]))).(core.ZodType[any, any])
}

// Nilable makes the literal schema nilable
func (z *ZodLiteral) Nilable() core.ZodType[any, any] {
	// Use Clone method to create new instance, avoiding manual state copying
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// No need to modify def, as Nilable is a runtime flag
	}).(*ZodLiteral)
	cloned.internals.SetNilable()
	return cloned
}

// Nullish makes the literal schema optional and nilable
func (z *ZodLiteral) Nullish() core.ZodType[any, any] {
	return any(Optional(z.Nilable())).(core.ZodType[any, any])
}

// Refine adds type-safe custom validation logic to the literal schema
func (z *ZodLiteral) Refine(fn func(any) bool, params ...any) *ZodLiteral {
	// Use existing RefineAny infrastructure with direct inline handling
	result := z.RefineAny(func(v any) bool {
		// Use reflectx for value extraction instead of custom extractLiteralValue
		if reflectx.IsNil(v) {
			// Handle nil: return true to let upper logic (Nilable flag) decide whether to allow
			return true
		}

		// Try to get the actual value (handle pointers)
		actualValue := v
		if deref, ok := reflectx.Deref(v); ok {
			actualValue = deref
		}

		return fn(actualValue)
	}, params...)
	// Return concrete ZodLiteral type to support method chaining
	return result.(*ZodLiteral)
}

// RefineAny adds flexible custom validation logic to the literal schema
func (z *ZodLiteral) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	// Use NewCustom from checks.go to create refine check, following unified pattern
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Check adds modern validation using direct payload access
func (z *ZodLiteral) Check(fn core.CheckFn) core.ZodType[any, any] {
	check := checks.NewCustom[any](func(v any) bool {
		// Use constructor instead of direct struct literal to respect private fields
		payload := core.NewParsePayloadWithPath(v, make([]any, 0))
		fn(payload)
		return len(payload.GetIssues()) == 0
	})
	return engine.AddCheck(z, check)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodLiteral) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodLiteralDefault is a Default wrapper for literal type
// Provides perfect type safety and chainable method support
type ZodLiteralDefault struct {
	*ZodDefault[*ZodLiteral] // Embed concrete pointer to enable method promotion
}

// Parse ensures correct validation call to inner type
func (s ZodLiteralDefault) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

// Default adds a default value to the literal schema, returns ZodLiteralDefault support chain call
// Compile-time type safety: Literal("hello").Default(123) will fail to compile
func (z *ZodLiteral) Default(value any) ZodLiteralDefault {
	return ZodLiteralDefault{
		&ZodDefault[*ZodLiteral]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the literal schema, returns ZodLiteralDefault support chain call
func (z *ZodLiteral) DefaultFunc(fn func() any) ZodLiteralDefault {
	return ZodLiteralDefault{
		&ZodDefault[*ZodLiteral]{
			innerType:   z,
			defaultFunc: fn,
			isFunction:  true,
		},
	}
}

// Refine adds a flexible validation function to the literal schema, returns ZodLiteralDefault support chain call
func (s ZodLiteralDefault) Refine(fn func(any) bool, params ...any) ZodLiteralDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodLiteralDefault{
		&ZodDefault[*ZodLiteral]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodLiteralDefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for smart handling of literal values
		if reflectx.IsNil(input) {
			return nil, fmt.Errorf("cannot transform nil literal")
		}

		// Get the actual value (handle pointers)
		actualValue := input
		if deref, ok := reflectx.Deref(input); ok {
			actualValue = deref
		}

		return fn(actualValue, ctx)
	})
}

// Optional adds an optional check to the literal schema, returns ZodType support chain call
func (s ZodLiteralDefault) Optional() core.ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the literal schema, returns ZodType support chain call
func (s ZodLiteralDefault) Nilable() core.ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(s).(core.ZodType[any, any]))
}

// ZodLiteralPrefault is a Prefault wrapper for literal type
// Provides perfect type safety and chainable method support
type ZodLiteralPrefault struct {
	*ZodPrefault[*ZodLiteral] // Embed generic wrapper
}

// Prefault adds a prefault value to the literal schema, returns ZodLiteralPrefault support chain call
// Compile-time type safety: Literal("hello").Prefault(123) will fail to compile
func (z *ZodLiteral) Prefault(value any) ZodLiteralPrefault {
	return ZodLiteralPrefault{
		&ZodPrefault[*ZodLiteral]{
			innerType:     z,
			prefaultValue: value,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the literal schema, returns ZodLiteralPrefault support chain call
func (z *ZodLiteral) PrefaultFunc(fn func() any) ZodLiteralPrefault {
	genericFn := func() any { return fn() }
	return ZodLiteralPrefault{
		&ZodPrefault[*ZodLiteral]{
			innerType:    z,
			prefaultFunc: genericFn,
			isFunction:   true,
		},
	}
}

// Refine adds a flexible validation function to the literal schema, returns ZodLiteralPrefault support chain call
func (l ZodLiteralPrefault) Refine(fn func(any) bool, params ...any) ZodLiteralPrefault {
	newInner := l.innerType.Refine(fn, params...)
	return ZodLiteralPrefault{
		&ZodPrefault[*ZodLiteral]{
			innerType:     newInner,
			prefaultValue: l.prefaultValue,
			prefaultFunc:  l.prefaultFunc,
			isFunction:    l.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (l ZodLiteralPrefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use embedded ZodPrefault's TransformAny method
	return l.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for smart handling of literal values
		if reflectx.IsNil(input) {
			return nil, fmt.Errorf("cannot transform nil literal")
		}

		// Get the actual value (handle pointers)
		actualValue := input
		if deref, ok := reflectx.Deref(input); ok {
			actualValue = deref
		}

		return fn(actualValue, ctx)
	})
}

// Optional adds an optional check to the literal schema, returns ZodType support chain call
func (l ZodLiteralPrefault) Optional() core.ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(l).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the literal schema, returns ZodType support chain call
func (l ZodLiteralPrefault) Nilable() core.ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(l).(core.ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodLiteralFromDef creates a ZodLiteral from a definition
func createZodLiteralFromDef(def *ZodLiteralDef) *ZodLiteral {
	internals := &ZodLiteralInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Values:           make(map[any]struct{}),
		Pattern:          nil,
		Isst:             issues.ZodIssueInvalidType{Expected: core.ZodTypeLiteral},
		Bag:              make(map[string]any),
	}

	// Build values set using utility pattern - only for hashable types
	for _, value := range def.Values {
		if isHashable(value) {
			internals.Values[value] = struct{}{}
			internals.ZodTypeInternals.Values[value] = struct{}{}
		}
		// For non-hashable types, we'll rely on deep comparison in parseLiteralCore
	}

	// Build regex pattern for string values
	var patterns []string
	for _, value := range def.Values {
		switch v := value.(type) {
		case string:
			// Use escapeRegex from utils.go for consistency
			escaped := utils.EscapeRegex(v)
			patterns = append(patterns, escaped)
		case nil:
			patterns = append(patterns, "null")
		default:
			// Convert to string representation and escape
			escaped := utils.EscapeRegex(fmt.Sprintf("%v", v))
			patterns = append(patterns, escaped)
		}
	}

	if len(patterns) > 0 {
		patternStr := fmt.Sprintf("^(%s)$", strings.Join(patterns, "|"))
		if compiled, err := regexp.Compile(patternStr); err == nil {
			internals.Pattern = compiled
		}
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		literalDef := &ZodLiteralDef{
			ZodTypeDef: *newDef,
			Type:       core.ZodTypeLiteral,
			Values:     def.Values, // Preserve values
		}
		return createZodLiteralFromDef(literalDef)
	}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		schema := &ZodLiteral{internals: internals}
		result, err := schema.Parse(payload.GetValue(), ctx)
		if err != nil {
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := issues.ConvertZodIssueToRaw(issue)
					rawIssue.Path = issue.Path
					payload.AddIssue(rawIssue)
				}
			}
			return payload
		}
		payload.SetValue(result)
		return payload
	}

	zodSchema := &ZodLiteral{internals: internals}

	// Initialize the schema using standard pattern
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

// Literal creates a new literal schema
func Literal(values ...any) *ZodLiteral {
	// Handle case where last parameter might be SchemaParams
	var params []core.SchemaParams
	var literalValues []any

	if len(values) == 0 {
		// No values provided, create empty literal (will never match)
		literalValues = []any{}
	} else {
		switch len(values) {
		case 1:
			// Single parameter - could be value, slice of values, or SchemaParams
			if param, ok := values[0].(core.SchemaParams); ok {
				// Single SchemaParams without values (edge case)
				literalValues = []any{}
				params = []core.SchemaParams{param}
			} else if slice, ok := values[0].([]any); ok {
				// Single slice - expand to multiple literal values
				literalValues = slice
			} else {
				// Single literal value
				literalValues = values
			}
		case 2:
			// Two parameters - check special case: slice + SchemaParams
			if slice, ok := values[0].([]any); ok {
				if param, ok := values[1].(core.SchemaParams); ok {
					// First param is slice, second is core.SchemaParams
					literalValues = slice
					params = []core.SchemaParams{param}
				} else {
					// First param is slice, second is literal value
					literalValues = make([]any, 0, len(slice)+1)
					literalValues = append(literalValues, slice...)
					literalValues = append(literalValues, values[1])
				}
			} else {
				// Check if last parameter is core.SchemaParams
				if lastParam, ok := values[len(values)-1].(core.SchemaParams); ok {
					// Last parameter is core.SchemaParams
					literalValues = values[:len(values)-1]
					params = []core.SchemaParams{lastParam}
				} else {
					// All parameters are literal values
					literalValues = values
				}
			}
		default:
			// Multiple parameters (3+) - check if last parameter is core.SchemaParams
			if lastParam, ok := values[len(values)-1].(core.SchemaParams); ok {
				// Last parameter is core.SchemaParams
				literalValues = values[:len(values)-1]
				params = []core.SchemaParams{lastParam}
			} else {
				// All parameters are literal values
				literalValues = values
			}
		}
	}

	def := &ZodLiteralDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeLiteral,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:   core.ZodTypeLiteral,
		Values: literalValues,
	}

	schema := createZodLiteralFromDef(def)

	// Apply schema parameters using standard pattern from type_string.go
	if len(params) > 0 {
		param := params[0]

		// Handle schema-level error configuration
		if param.Error != nil {
			// Create error map manually since CreateErrorMap doesn't exist
			var errorMap core.ZodErrorMap
			if msg, ok := param.Error.(string); ok {
				errorMap = func(issue core.ZodRawIssue) string {
					return msg
				}
			} else if fn, ok := param.Error.(core.ZodErrorMap); ok {
				errorMap = fn
			}

			if errorMap != nil {
				def.Error = &errorMap
				schema.internals.Error = &errorMap
			}
		}
	}

	return schema
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodLiteral) CloneFrom(source any) {
	if src, ok := source.(*ZodLiteral); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]any)
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy literal-specific fields
		if src.internals.Values != nil {
			z.internals.Values = make(map[any]struct{})
			for k, v := range src.internals.Values {
				z.internals.Values[k] = v
			}
		}

		z.internals.Pattern = src.internals.Pattern
	}
}

// GetZod returns the literal-specific internals
func (z *ZodLiteral) GetZod() *ZodLiteralInternals {
	return z.internals
}

// Clone creates a copy of the literal schema
func (z *ZodLiteral) Clone() core.ZodType[any, any] {
	return createZodLiteralFromDef(z.internals.Def)
}

// isValidLiteral checks if a value matches any of the allowed literal values
func isValidLiteral(input any, internals *ZodLiteralInternals) bool {
	// For hashable types, use direct map lookup for performance
	if isHashable(input) {
		_, exists := internals.Values[input]
		return exists
	}

	// For non-hashable types (slice, map, function), use deep comparison
	for _, allowedValue := range internals.Def.Values {
		if reflect.DeepEqual(input, allowedValue) {
			return true
		}
	}
	return false
}

// inferTypedNilFromLiterals returns a typed nil pointer based on literal values
func inferTypedNilFromLiterals(values []any) any {
	if len(values) == 0 {
		return (*any)(nil)
	}

	// Infer type from first literal value
	switch values[0].(type) {
	case string:
		return (*string)(nil)
	case int:
		return (*int)(nil)
	case bool:
		return (*bool)(nil)
	case float64:
		return (*float64)(nil)
	default:
		return (*any)(nil)
	}
}

// isHashable checks if a value can be used as a map key
// Use reflectx for more robust type checking
func isHashable(value any) bool {
	if reflectx.IsNil(value) {
		return true
	}

	// Use reflectx type checking functions
	if reflectx.IsSlice(value) || reflectx.IsMap(value) || reflectx.IsFunc(value) {
		return false // slices, maps, and functions are not hashable
	}

	// Use reflection to check for any map type or slice type
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Func:
		return false // maps, slices, and functions are not hashable
	case reflect.Invalid:
		return false // invalid values are not hashable
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.Chan, reflect.Interface, reflect.Pointer, reflect.String,
		reflect.Struct, reflect.UnsafePointer:
		return true // these types are hashable
	default:
		return true // most other types are hashable
	}
}

//////////////////////////
//  ACCESSOR METHODS    //
//////////////////////////

// Value returns the single literal value. It panics if the schema was defined with multiple values.
// This mirrors the behavior of Zod v4's `.value` getter.
func (z *ZodLiteral) Value() any {
	if len(z.internals.Def.Values) != 1 {
		panic("literal.Value() is only available on single-value literal schemas")
	}
	return z.internals.Def.Values[0]
}
