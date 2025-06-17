package gozod

import (
	"errors"
	"fmt"
	"reflect"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodArrayDef defines the configuration for fixed-length array schemas
type ZodArrayDef struct {
	ZodTypeDef
	Type   string              // "array"
	Items  []ZodType[any, any] // Fixed array element schemas
	Checks []ZodCheck          // Array-specific validation checks
}

// ZodArrayInternals represents the internal state of an array schema
type ZodArrayInternals struct {
	ZodTypeInternals
	Def    *ZodArrayDef           // Array definition
	Items  []ZodType[any, any]    // Fixed array element schemas
	Checks []ZodCheck             // Validation checks
	Bag    map[string]interface{} // Runtime configuration
}

// ZodArray represents a fixed-length array schema
type ZodArray struct {
	internals *ZodArrayInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodArray) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for array type conversion
func (z *ZodArray) Coerce(input interface{}) (interface{}, bool) {
	return coerceToArrayFunc(input)
}

// Parse validates and parses input with array element validation
func (z *ZodArray) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// create validator with element validation
	validator := func(value []interface{}, checks []ZodCheck, ctx *ParseContext) error {
		return validateArrayWithElements(value, checks, z, ctx)
	}

	// create type checker with reflection conversion
	typeChecker := func(v any) ([]interface{}, bool) {
		switch val := v.(type) {
		case []interface{}:
			return val, true
		default:
			// handle other slice/array types, check with reflection
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
				// convert to []interface{}
				length := rv.Len()
				slice := make([]interface{}, length)
				for i := 0; i < length; i++ {
					slice[i] = rv.Index(i).Interface()
				}
				return slice, true
			}
		}
		return nil, false
	}

	// create coercer
	coercer := func(v any) ([]interface{}, bool) {
		coerced := tryCoerceToArray(v)
		if result, ok := coerced.([]interface{}); ok {
			return result, true
		}
		return nil, false
	}

	return parseType[[]interface{}](
		input,
		&z.internals.ZodTypeInternals,
		"array",
		typeChecker,
		func(v any) (*[]interface{}, bool) { ptr, ok := v.(*[]interface{}); return ptr, ok },
		validator,
		coercer,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodArray) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Length adds exact length validation
func (z *ZodArray) Length(length int, params ...SchemaParams) *ZodArray {
	check := NewZodCheckLengthEquals(length, params...)
	result := AddCheck(z, check)
	return result.(*ZodArray)
}

// Min adds minimum length validation
func (z *ZodArray) Min(minimum int, params ...SchemaParams) *ZodArray {
	check := NewZodCheckMinLength(minimum, params...)
	result := AddCheck(z, check)
	return result.(*ZodArray)
}

// Max adds maximum length validation
func (z *ZodArray) Max(maximum int, params ...SchemaParams) *ZodArray {
	check := NewZodCheckMaxLength(maximum, params...)
	result := AddCheck(z, check)
	return result.(*ZodArray)
}

// NonEmpty adds non-empty validation
func (z *ZodArray) NonEmpty(params ...SchemaParams) *ZodArray {
	check := NewZodCheckMinLength(1, params...)
	result := AddCheck(z, check)
	return result.(*ZodArray)
}

// Element returns the schema for the element at the given index
func (z *ZodArray) Element(index int) ZodType[any, any] {
	if index >= 0 && index < len(z.internals.Items) {
		return z.internals.Items[index]
	}
	return nil
}

// Items returns all element schemas
func (z *ZodArray) Items() []ZodType[any, any] {
	result := make([]ZodType[any, any], len(z.internals.Items))
	copy(result, z.internals.Items)
	return result
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform provides type-safe array transformation
func (z *ZodArray) Transform(fn func([]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		arr, isNil, err := extractArrayValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilArray
		}
		return fn(arr, ctx)
	})
}

// TransformAny provides flexible transformation with any input/output types
func (z *ZodArray) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodArray) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  z,
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the array optional
func (z *ZodArray) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable makes the array nilable
// Nilable allows the value to be nil, returns a new schema without modifying the original
func (z *ZodArray) Nilable() ZodType[any, any] {
	// Create a deep copy to maintain immutability
	clone := &ZodArray{
		internals: &ZodArrayInternals{
			ZodTypeInternals: ZodTypeInternals{
				Nilable: true, // Set the Nilable flag
				Coerce:  z.internals.Coerce,
				Parse:   z.internals.Parse,
			},
			Items:  make([]ZodType[any, any], len(z.internals.Items)),
			Checks: make([]ZodCheck, len(z.internals.Checks)),
			Bag:    make(map[string]interface{}),
		},
	}

	// Deep copy Items
	copy(clone.internals.Items, z.internals.Items)

	// Deep copy Checks
	copy(clone.internals.Checks, z.internals.Checks)

	// Copy Bag
	for k, v := range z.internals.Bag {
		clone.internals.Bag[k] = v
	}

	return clone
}

// Nullish makes the array both optional and nilable
func (z *ZodArray) Nullish() ZodType[any, any] {
	return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Refine adds type-safe custom validation logic to the array schema
func (z *ZodArray) Refine(fn func([]interface{}) bool, params ...SchemaParams) *ZodArray {
	result := z.RefineAny(func(v any) bool {
		arr, isNil, err := extractArrayValue(v)
		if err != nil {
			return false
		}
		if isNil {
			// nil *[]interface{} handling: return true to let upper logic (Nilable flag) decide whether to allow
			return true
		}
		return fn(arr)
	}, params...)
	return result.(*ZodArray)
}

// RefineAny adds flexible custom validation logic
func (z *ZodArray) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodArray) Check(fn CheckFn) *ZodArray {
	check := NewCustom[[]interface{}](func(v []interface{}) bool {
		payload := &ParsePayload{
			Value:  v,
			Issues: make([]ZodRawIssue, 0),
			Path:   make([]interface{}, 0),
		}
		beforeIssueCount := len(payload.Issues)
		fn(payload) // CheckFn modifies payload.Issues directly
		// Return true if no new issues were added
		return len(payload.Issues) == beforeIssueCount
	}, SchemaParams{})
	result := AddCheck(z, check)
	return result.(*ZodArray)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodArray) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodArrayDefault is a default value wrapper for array type
type ZodArrayDefault struct {
	*ZodDefault[*ZodArray]
}

// Default creates a default wrapper with type safety
func (z *ZodArray) Default(value []interface{}) ZodArrayDefault {
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a default wrapper with function
func (z *ZodArray) DefaultFunc(fn func() []interface{}) ZodArrayDefault {
	genericFn := func() any { return fn() }
	return ZodArrayDefault{
		&ZodDefault[*ZodArray]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// ZodArrayDefault chainable validation methods

func (s ZodArrayDefault) Length(length int, params ...SchemaParams) ZodArrayDefault {
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

func (s ZodArrayDefault) Min(minimum int, params ...SchemaParams) ZodArrayDefault {
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

func (s ZodArrayDefault) Max(maximum int, params ...SchemaParams) ZodArrayDefault {
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

func (s ZodArrayDefault) Refine(fn func([]interface{}) bool, params ...SchemaParams) ZodArrayDefault {
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

func (s ZodArrayDefault) Transform(fn func([]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		arr, isNil, err := extractArrayValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilArray
		}
		return fn(arr, ctx)
	})
}

func (s ZodArrayDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodArrayDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ZodArrayPrefault is a prefault value wrapper for array type
type ZodArrayPrefault struct {
	*ZodPrefault[*ZodArray]
}

// Prefault creates a prefault wrapper with type safety
func (z *ZodArray) Prefault(value []interface{}) ZodArrayPrefault {
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			innerType:     z,
			prefaultValue: value,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a prefault wrapper with function
func (z *ZodArray) PrefaultFunc(fn func() []interface{}) ZodArrayPrefault {
	genericFn := func() any { return fn() }
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			innerType:    z,
			prefaultFunc: genericFn,
			isFunction:   true,
		},
	}
}

// ZodArrayPrefault chainable validation methods

func (a ZodArrayPrefault) Min(minimum int, params ...SchemaParams) ZodArrayPrefault {
	newInner := a.innerType.Min(minimum, params...)
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			innerType:     newInner,
			prefaultValue: a.prefaultValue,
			prefaultFunc:  a.prefaultFunc,
			isFunction:    a.isFunction,
		},
	}
}

func (a ZodArrayPrefault) Max(maximum int, params ...SchemaParams) ZodArrayPrefault {
	newInner := a.innerType.Max(maximum, params...)
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			innerType:     newInner,
			prefaultValue: a.prefaultValue,
			prefaultFunc:  a.prefaultFunc,
			isFunction:    a.isFunction,
		},
	}
}

func (a ZodArrayPrefault) Length(length int, params ...SchemaParams) ZodArrayPrefault {
	newInner := a.innerType.Length(length, params...)
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			innerType:     newInner,
			prefaultValue: a.prefaultValue,
			prefaultFunc:  a.prefaultFunc,
			isFunction:    a.isFunction,
		},
	}
}

func (a ZodArrayPrefault) NonEmpty(params ...SchemaParams) ZodArrayPrefault {
	newInner := a.innerType.NonEmpty(params...)
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			innerType:     newInner,
			prefaultValue: a.prefaultValue,
			prefaultFunc:  a.prefaultFunc,
			isFunction:    a.isFunction,
		},
	}
}

func (a ZodArrayPrefault) Refine(fn func([]interface{}) bool, params ...SchemaParams) ZodArrayPrefault {
	newInner := a.innerType.Refine(fn, params...)
	return ZodArrayPrefault{
		&ZodPrefault[*ZodArray]{
			innerType:     newInner,
			prefaultValue: a.prefaultValue,
			prefaultFunc:  a.prefaultFunc,
			isFunction:    a.isFunction,
		},
	}
}

func (a ZodArrayPrefault) Transform(fn func([]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return a.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		arrayVal, isNil, err := extractArrayValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilArray
		}
		return fn(arrayVal, ctx)
	})
}

func (a ZodArrayPrefault) Optional() ZodType[any, any] {
	return Optional(any(a).(ZodType[any, any]))
}

func (a ZodArrayPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(a).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodArrayFromDef creates a ZodArray from definition
func createZodArrayFromDef(def *ZodArrayDef) *ZodArray {
	internals := &ZodArrayInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Items:            def.Items,
		Checks:           def.Checks,
		Bag:              make(map[string]interface{}),
	}

	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		arrayDef := &ZodArrayDef{
			ZodTypeDef: *newDef,
			Type:       "array",
			Items:      def.Items,
			Checks:     newDef.Checks,
		}
		return createZodArrayFromDef(arrayDef)
	}

	schema := &ZodArray{internals: internals}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
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

	initZodType(schema, &def.ZodTypeDef)
	return schema
}

// Array creates a new fixed-length array schema with the given element schemas
func Array(args ...interface{}) *ZodArray {
	var items []ZodType[any, any]
	var params *SchemaParams

	if len(args) == 0 {
		items = []ZodType[any, any]{}
	} else {
		lastArg := args[len(args)-1]
		if schemaParam, ok := lastArg.(SchemaParams); ok {
			params = &schemaParam
			for i := 0; i < len(args)-1; i++ {
				if schema, ok := args[i].(ZodType[any, any]); ok {
					items = append(items, schema)
				} else {
					panic(fmt.Sprintf("Array() argument %d must be ZodType, got %T", i, args[i]))
				}
			}
		} else {
			for i, arg := range args {
				if schema, ok := arg.(ZodType[any, any]); ok {
					items = append(items, schema)
				} else {
					panic(fmt.Sprintf("Array() argument %d must be ZodType, got %T", i, arg))
				}
			}
		}
	}

	def := &ZodArrayDef{
		ZodTypeDef: ZodTypeDef{Type: "array"},
		Type:       "array",
		Items:      items,
		Checks:     make([]ZodCheck, 0),
	}

	if params != nil {
		if params.Error != nil {
			errorMap := createErrorMap(params.Error)
			if errorMap != nil {
				def.Error = errorMap
			}
		}
	}

	schema := createZodArrayFromDef(def)

	if params != nil {
		if params.Error != nil {
			errorMap := createErrorMap(params.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
		if params.Coerce {
			schema.internals.Bag["coerce"] = true
		}
	}

	return schema
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// validateArrayWithElements validates array elements and runs checks following the original working pattern
func validateArrayWithElements(value []interface{}, checks []ZodCheck, z *ZodArray, ctx *ParseContext) error {
	// First run the additional checks (Min, Max, etc.)
	if len(checks) > 0 {
		payload := &ParsePayload{
			Value:  value,
			Issues: make([]ZodRawIssue, 0),
			Path:   []interface{}{},
		}
		runChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
		}
	}

	// Always check length matching - this creates the fixed-length constraint
	expectedLength := len(z.internals.Items)
	actualLength := len(value)

	if actualLength != expectedLength {
		var code string
		var message string
		if actualLength > expectedLength {
			code = string(TooBig)
			message = fmt.Sprintf("Array must have at most %d element(s), got %d", expectedLength, actualLength)
		} else {
			code = string(TooSmall)
			message = fmt.Sprintf("Array must have at least %d element(s), got %d", expectedLength, actualLength)
		}

		return &ZodError{
			Issues: []ZodIssue{
				{
					ZodIssueBase: ZodIssueBase{
						Code:    code,
						Message: message,
						Input:   value,
						Path:    []interface{}{},
					},
				},
			},
		}
	}

	// Validate each element against its schema if items are defined
	if len(z.internals.Items) > 0 {
		for i, itemSchema := range z.internals.Items {
			if i < len(value) {
				// If array has coercion enabled, enable it for element types too
				itemInternals := itemSchema.GetInternals()
				if shouldCoerce(z.internals.Bag) {
					if itemInternals.Bag == nil {
						itemInternals.Bag = make(map[string]interface{})
					}
					itemInternals.Bag["coerce"] = true
				}

				result, err := itemSchema.Parse(value[i], ctx)
				if err != nil {
					return err
				}

				// Update the element value (may have been coerced)
				value[i] = result
			}
		}
	}

	return nil
}

// GetZod returns the array-specific internals
func (z *ZodArray) GetZod() *ZodArrayInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface for array-specific state copying
func (z *ZodArray) CloneFrom(source any) {
	if src, ok := source.(*ZodArray); ok {
		z.internals.Items = make([]ZodType[any, any], len(src.internals.Items))
		copy(z.internals.Items, src.internals.Items)

		z.internals.Checks = make([]ZodCheck, len(src.internals.Checks))
		copy(z.internals.Checks, src.internals.Checks)

		if src.internals.Bag != nil {
			z.internals.Bag = make(map[string]interface{})
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
	}
}

// coerceToArrayFunc attempts to coerce a value to []interface{}
func coerceToArrayFunc(input interface{}) ([]interface{}, bool) {
	coerced := tryCoerceToArray(input)
	if result, ok := coerced.([]interface{}); ok {
		return result, true
	}
	return nil, false
}

// tryCoerceToArray attempts to coerce input to an array
func tryCoerceToArray(input interface{}) interface{} {
	if input == nil {
		return []interface{}{}
	}

	reflectValue := reflect.ValueOf(input)
	switch reflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		length := reflectValue.Len()
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			result[i] = reflectValue.Index(i).Interface()
		}
		return result
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr,
		reflect.String, reflect.Struct, reflect.UnsafePointer:
		return []interface{}{input}
	default:
		return []interface{}{input}
	}
}

// CreateInvalidArrayLengthIssue creates an issue for invalid array length
func CreateInvalidArrayLengthIssue(value interface{}, expected, actual int, modifier func(*ZodRawIssue)) ZodRawIssue {
	code := TooSmall
	if actual > expected {
		code = TooBig
	}

	return NewRawIssue(
		string(code),
		value,
		WithOrigin("array"),
		WithExpected(fmt.Sprintf("length %d", expected)),
		WithReceived(fmt.Sprintf("length %d", actual)),
		modifier,
	)
}

// extractArrayValue extracts array value from input with smart handling
func extractArrayValue(input any) ([]interface{}, bool, error) {
	switch v := input.(type) {
	case []interface{}:
		return v, false, nil
	case *[]interface{}:
		if v == nil {
			return nil, true, nil
		}
		return *v, false, nil
	default:
		rv := reflect.ValueOf(input)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			length := rv.Len()
			result := make([]interface{}, length)
			for i := 0; i < length; i++ {
				result[i] = rv.Index(i).Interface()
			}
			return result, false, nil
		}
		return nil, false, fmt.Errorf("%w, got %T", ErrExpectedArray, input)
	}
}
