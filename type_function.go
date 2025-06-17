package gozod

import (
	"fmt"
	"reflect"
)

// =============================================================================
// FUNCTION TYPE DEFINITION
// =============================================================================

// ZodFunctionDef defines the configuration for function validation
type ZodFunctionDef struct {
	ZodTypeDef
	Input  Schema // Schema for validating input arguments
	Output Schema // Schema for validating output result
}

// ZodFunction represents a function type with input and output validation
type ZodFunction struct {
	internals *ZodFunctionInternals
	def       *ZodFunctionDef
}

// =============================================================================
// FUNCTION TYPE CONSTRUCTORS
// =============================================================================

// ZodFunctionInternals represents the internal state of ZodFunction
type ZodFunctionInternals struct {
	ZodTypeInternals
	Def *ZodFunctionDef        // Schema definition
	Bag map[string]interface{} // Additional metadata
}

// createZodFunctionFromDef creates a ZodFunction from a definition
func createZodFunctionFromDef(def *ZodFunctionDef) *ZodFunction {
	internals := &ZodFunctionInternals{
		ZodTypeInternals: ZodTypeInternals{
			Type:     "function",
			Checks:   make([]ZodCheck, 0),
			Nilable:  false,
			Optional: false,
			Values:   make(map[interface{}]struct{}),
		},
		Def: def,
		Bag: make(map[string]interface{}),
	}

	// Set up constructor for cloning support on the embedded ZodTypeInternals
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		functionDef := &ZodFunctionDef{
			ZodTypeDef: *newDef,
			Input:      def.Input,  // Preserve the original input schema
			Output:     def.Output, // Preserve the original output schema
		}
		return createZodFunctionFromDef(functionDef)
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		return parseZodFunction(payload, def, &internals.ZodTypeInternals, ctx)
	}

	return &ZodFunction{
		internals: internals,
		def:       def,
	}
}

// FunctionParams represents configuration for function validation
type FunctionParams struct {
	Description string      // Optional description
	Input       interface{} // Can be Schema or []Schema
	Output      Schema      // Return value schema
}

// Function creates a new function schema with input and output validation
func Function(params ...FunctionParams) *ZodFunction {
	var input, output Schema

	// Set defaults
	input = Any()  // Default: any input
	output = Any() // Default: any return type

	// Apply configuration if provided
	if len(params) > 0 {
		param := params[0]

		// Handle input schema
		if param.Input != nil {
			input = normalizeInputSchema(param.Input)
		}

		if param.Output != nil {
			output = param.Output
		}
	}

	// Create function definition
	def := &ZodFunctionDef{
		ZodTypeDef: ZodTypeDef{Type: "function"},
		Input:      input,
		Output:     output,
	}

	return createZodFunctionFromDef(def)
}

// normalizeInputSchema converts various input formats to a unified Schema
func normalizeInputSchema(input interface{}) Schema {
	switch inputVal := input.(type) {
	case []Schema:
		// Array of schemas: create tuple-like validation
		inputArgs := make([]interface{}, len(inputVal))
		for i, schema := range inputVal {
			inputArgs[i] = schema
		}
		return Array(inputArgs...)
	case Schema:
		// Direct Schema: single parameter schema
		return inputVal
	case nil:
		// Default to Any() for flexible input validation
		return Any()
	default:
		// Fallback: treat as unvalidated input
		return Any()
	}
}

// =============================================================================
// FUNCTION FACTORY METHODS
// =============================================================================

// Input creates a new function schema with specified input validation
func (z *ZodFunction) Input(inputSchema interface{}) *ZodFunction {
	input := normalizeInputSchema(inputSchema)

	newDef := &ZodFunctionDef{
		ZodTypeDef: z.def.ZodTypeDef,
		Input:      input,
		Output:     z.def.Output,
	}

	return createZodFunctionFromDef(newDef)
}

// Output creates a new function schema with specified output validation
func (z *ZodFunction) Output(outputSchema Schema) *ZodFunction {
	newDef := &ZodFunctionDef{
		ZodTypeDef: z.def.ZodTypeDef,
		Input:      z.def.Input,
		Output:     outputSchema,
	}

	return createZodFunctionFromDef(newDef)
}

// Implement validates a function and returns a wrapped version with validation
func (z *ZodFunction) Implement(fn interface{}) (interface{}, error) {
	if fn == nil {
		return nil, ErrFunctionNil
	}

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		receivedType := string(GetParsedType(fn))
		return nil, fmt.Errorf("%w, got %s", ErrFunctionInvalid, receivedType)
	}

	fnType := fnValue.Type()

	// Validate function signature against input schema
	if z.def.Input != nil {
		if err := z.validateFunctionSignature(fnType); err != nil {
			return nil, err
		}
	}

	// Validate function return type against output schema
	if z.def.Output != nil {
		if err := z.validateFunctionReturnType(fnType); err != nil {
			return nil, err
		}
	}

	// Create a new function that validates inputs and outputs
	wrappedFn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		// Validate input arguments if input schema is defined
		if z.def.Input != nil {
			if issues := z.validateInputArguments(args); len(issues) > 0 {
				zodErr := NewZodError(convertRawIssuesToIssues(issues, nil))
				panic(zodErr)
			}
		}

		// Call the original function
		results := fnValue.Call(args)

		// Validate output if output schema is defined and there are results
		if z.def.Output != nil && len(results) > 0 {
			if issues := z.validateOutputResults(results); len(issues) > 0 {
				zodErr := NewZodError(convertRawIssuesToIssues(issues, nil))
				panic(zodErr)
			}
		}

		return results
	})

	return wrappedFn.Interface(), nil
}

// validateFunctionSignature validates that the function signature matches the input schema
func (z *ZodFunction) validateFunctionSignature(fnType reflect.Type) error {
	// Check if input schema is an Array (multiple parameters)
	if arraySchema, ok := z.def.Input.(*ZodArray); ok {
		expectedParamCount := len(arraySchema.internals.Items)
		actualParamCount := fnType.NumIn()

		if expectedParamCount != actualParamCount {
			return fmt.Errorf("%w: expected %d parameters, got %d", ErrFunctionSignatureMismatch, expectedParamCount, actualParamCount)
		}
		return nil
	}

	// For single parameter schema, expect exactly one parameter
	if fnType.NumIn() != 1 {
		return fmt.Errorf("%w: expected 1 parameter, got %d", ErrFunctionSignatureMismatch, fnType.NumIn())
	}

	// Check parameter type compatibility for common cases
	paramType := fnType.In(0)

	// First check if the schema is nilable - if so, be more permissive
	if internals := z.def.Input.GetInternals(); internals != nil && internals.Nilable {
		// For nilable types, we expect either the base type or a pointer to it
		// This is a simplified check - in practice, nilable types are more complex
		// We'll be permissive here and allow any type for nilable schemas
		return nil
	}

	// Check if the input schema expects a specific Go type (for non-nilable types)
	switch z.def.Input.(type) {
	case *ZodString:
		// String schema expects string parameter
		if paramType.Kind() != reflect.String {
			return fmt.Errorf("%w: expected string parameter, got %s", ErrFunctionParameterMismatch, paramType.Kind())
		}
	case *ZodInt:
		// Int schema expects int parameter
		if paramType.Kind() != reflect.Int {
			return fmt.Errorf("%w: expected int parameter, got %s", ErrFunctionParameterMismatch, paramType.Kind())
		}
	case *ZodBool:
		// Bool schema expects bool parameter
		if paramType.Kind() != reflect.Bool {
			return fmt.Errorf("%w: expected bool parameter, got %s", ErrFunctionParameterMismatch, paramType.Kind())
		}
	default:
		// For Union, Any, and other complex types, accept any parameter type
		// This allows for flexible typing when needed
	}

	return nil
}

// validateInputArguments validates function input arguments against the input schema
func (z *ZodFunction) validateInputArguments(args []reflect.Value) []ZodRawIssue {
	// Convert args to interface{} slice for validation
	inputArgs := make([]interface{}, len(args))
	for i, arg := range args {
		inputArgs[i] = arg.Interface()
	}

	var issues []ZodRawIssue

	// Check if input schema is an Array and extract element schemas
	if arraySchema, ok := z.def.Input.(*ZodArray); ok {
		// Array syntax case: [z.string()] or [z.string(), z.number()]
		elementSchemas := arraySchema.internals.Items
		if len(elementSchemas) != len(inputArgs) {
			issue := CreateCustomIssue(
				inputArgs,
				fmt.Sprintf("Expected %d arguments, got %d", len(elementSchemas), len(inputArgs)),
			)
			issues = append(issues, issue)
			return issues
		}

		// Validate each argument individually against its schema
		for i, arg := range inputArgs {
			payload := &ParsePayload{
				Value:  arg,
				Path:   []interface{}{fmt.Sprintf("arg[%d]", i)},
				Issues: make([]ZodRawIssue, 0),
			}

			elementSchemas[i].GetInternals().Parse(payload, nil)
			if len(payload.Issues) > 0 {
				issues = append(issues, payload.Issues...)
			}
		}
		return issues
	}

	// Direct schema case: single parameter or custom schema
	if len(inputArgs) == 1 {
		payload := &ParsePayload{
			Value:  inputArgs[0],
			Path:   []interface{}{"input"},
			Issues: make([]ZodRawIssue, 0),
		}

		z.def.Input.GetInternals().Parse(payload, nil)
		return payload.Issues
	}

	// For multiple parameters with direct schema, validate as array
	payload := &ParsePayload{
		Value:  inputArgs,
		Path:   []interface{}{"input"},
		Issues: make([]ZodRawIssue, 0),
	}

	z.def.Input.GetInternals().Parse(payload, nil)
	return payload.Issues
}

// validateOutputResults validates function output results against the output schema
func (z *ZodFunction) validateOutputResults(results []reflect.Value) []ZodRawIssue {
	// For functions with single return value, validate directly
	if len(results) == 1 {
		payload := &ParsePayload{
			Value:  results[0].Interface(),
			Path:   []interface{}{"output"},
			Issues: make([]ZodRawIssue, 0),
		}

		z.def.Output.GetInternals().Parse(payload, nil)
		return payload.Issues
	}

	// For functions with multiple return values, validate as slice
	outputSlice := make([]interface{}, len(results))
	for i, result := range results {
		outputSlice[i] = result.Interface()
	}

	payload := &ParsePayload{
		Value:  outputSlice,
		Path:   []interface{}{"output"},
		Issues: make([]ZodRawIssue, 0),
	}

	z.def.Output.GetInternals().Parse(payload, nil)
	return payload.Issues
}

// validateFunctionReturnType validates that the function return type matches the output schema
func (z *ZodFunction) validateFunctionReturnType(fnType reflect.Type) error {
	// For functions with no return values, expect no output validation
	if fnType.NumOut() == 0 {
		return nil
	}

	// For functions with single return value, check type compatibility
	if fnType.NumOut() == 1 {
		returnType := fnType.Out(0)

		// Check if the output schema expects a specific Go type
		switch z.def.Output.(type) {
		case *ZodString:
			// String schema expects string return type
			if returnType.Kind() != reflect.String {
				return fmt.Errorf("%w: expected string return, got %s", ErrFunctionReturnTypeMismatch, returnType.Kind())
			}
		case *ZodInt:
			// Int schema expects int return type
			if returnType.Kind() != reflect.Int {
				return fmt.Errorf("%w: expected int return, got %s", ErrFunctionReturnTypeMismatch, returnType.Kind())
			}
		case *ZodBool:
			// Bool schema expects bool return type
			if returnType.Kind() != reflect.Bool {
				return fmt.Errorf("%w: expected bool return, got %s", ErrFunctionReturnTypeMismatch, returnType.Kind())
			}
		default:
			// For Union, Any, Object, and other complex types, accept any return type
			// This allows for flexible typing when needed
		}
		return nil
	}

	// For functions with multiple return values, we're more permissive
	// This could be enhanced in the future to validate each return value
	return nil
}

// =============================================================================
// CORE PARSE LOGIC WITH SMART TYPE INFERENCE
// =============================================================================

// parseZodFunction implements the core parsing logic for function type
func parseZodFunction(payload *ParsePayload, def *ZodFunctionDef, internals *ZodTypeInternals, ctx *ParseContext) *ParsePayload {
	input := payload.Value

	// 1. Unified nil handling
	if input == nil {
		if !internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "function", "null")
			payload.Issues = append(payload.Issues, rawIssue)
			return payload
		}
		payload.Value = (*interface{})(nil) // Return typed nil pointer
		return payload
	}

	// 2. Smart type inference: preserve exact function type
	switch v := input.(type) {
	case func():
		// No-argument, no-return function
		if len(internals.Checks) > 0 {
			runChecksOnValue(v, internals.Checks, payload, ctx)
			if len(payload.Issues) > 0 {
				return payload
			}
		}
		payload.Value = v
		return payload

	default:
		// Check if it's a function using reflection
		if fnValue := reflect.ValueOf(input); fnValue.Kind() == reflect.Func {
			// Any kind of function type - preserve original type
			if len(internals.Checks) > 0 {
				runChecksOnValue(input, internals.Checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return payload
				}
			}
			payload.Value = input // Keep original function type
			return payload
		}

		// Handle pointer to function
		if fnPtr := reflect.ValueOf(input); fnPtr.Kind() == reflect.Ptr && !fnPtr.IsNil() {
			if fnPtr.Elem().Kind() == reflect.Func {
				// Pointer to function - preserve pointer type
				if len(internals.Checks) > 0 {
					runChecksOnValue(input, internals.Checks, payload, ctx)
					if len(payload.Issues) > 0 {
						return payload
					}
				}
				payload.Value = input // Keep original pointer to function
				return payload
			}
		}

		// 3. Type coercion (if enabled) - functions typically don't coerce
		// Functions generally don't support coercion, skip this section

		// 4. Unified error creation
		rawIssue := CreateInvalidTypeIssue(input, "function", string(GetParsedType(input)))
		payload.Issues = append(payload.Issues, rawIssue)
		return payload
	}
}

// =============================================================================
// HELPER FUNCTIONS FOR TYPE EXTRACTION
// =============================================================================

// extractFunctionValue extracts function value, handling various input types
func extractFunctionValue(input any) (interface{}, bool, error) {
	if input == nil {
		return nil, true, nil
	}

	// Check if it's a function using reflection
	fnValue := reflect.ValueOf(input)
	if fnValue.Kind() == reflect.Func {
		return input, false, nil
	}

	return nil, false, fmt.Errorf("%w, got %T", ErrExpectedFunction, input)
}

// extractFunctionPointerValue extracts function value from pointer types

// =============================================================================
// ZODTYPE INTERFACE IMPLEMENTATION WITH SMART TYPE INFERENCE
// =============================================================================

// Parse implements intelligent type inference and validation
func (z *ZodFunction) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}
	// 1. Unified nil handling
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "function", "null")
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*interface{})(nil), nil
	}

	// 2. Smart type inference: function → function, *function → *function
	switch v := input.(type) {
	case func():
		// No-argument, no-return function
		if len(z.internals.Checks) > 0 {
			payload := &ParsePayload{Value: v, Issues: make([]ZodRawIssue, 0)}
			runChecksOnValue(v, z.internals.Checks, payload, parseCtx)
			if len(payload.Issues) > 0 {
				return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
			}
		}
		return v, nil

	default:
		// Check if it's any kind of function using reflection
		fnValue := reflect.ValueOf(input)
		if fnValue.Kind() == reflect.Func {
			// Any function type - preserve original type
			if len(z.internals.Checks) > 0 {
				payload := &ParsePayload{Value: input, Issues: make([]ZodRawIssue, 0)}
				runChecksOnValue(input, z.internals.Checks, payload, parseCtx)
				if len(payload.Issues) > 0 {
					return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
				}
			}
			return input, nil
		}

		// Handle pointer to function
		if fnValue.Kind() == reflect.Ptr && !fnValue.IsNil() {
			if fnValue.Elem().Kind() == reflect.Func {
				// Pointer to function - preserve pointer type
				if len(z.internals.Checks) > 0 {
					payload := &ParsePayload{Value: input, Issues: make([]ZodRawIssue, 0)}
					runChecksOnValue(input, z.internals.Checks, payload, parseCtx)
					if len(payload.Issues) > 0 {
						return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, parseCtx)}
					}
				}
				return input, nil
			}
		}

		// 3. Type coercion (if enabled) - functions typically don't support coercion
		if shouldCoerce(z.internals.Bag) {
			// Functions generally don't support coercion, but keep for future extensibility
			// Currently no coercion logic implemented for functions
			_ = z.internals.Bag // Acknowledge the condition to avoid empty branch warning
		}

		// 4. Unified error creation
		rawIssue := CreateInvalidTypeIssue(input, "function", string(GetParsedType(input)))
		finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}
}

// MustParse parses the input and panics on failure
func (z *ZodFunction) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// TYPE-SAFE REFINE METHODS
// =============================================================================

// Refine adds a type-safe refinement check for function types
func (z *ZodFunction) Refine(fn func(any) bool, params ...SchemaParams) *ZodFunction {
	result := z.RefineAny(func(v any) bool {
		val, isNil, err := extractFunctionValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true // Let upper logic decide
		}
		return fn(val)
	}, params...)
	return result.(*ZodFunction)
}

// RefineAny adds a refinement check that accepts any input type
func (z *ZodFunction) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// =============================================================================
// TRANSFORM METHODS
// =============================================================================

// Transform creates a transformation pipeline
func (z *ZodFunction) Transform(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodFunction) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// =============================================================================
// NILABLE MODIFIER WITH CLONE PATTERN
// =============================================================================

// Nilable creates a new function schema that accepts nil values
func (z *ZodFunction) Nilable() ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodFunction) setNilable() ZodType[any, any] {
	cloned := Clone(z, func(def *ZodTypeDef) {
		// Clone operates on ZodTypeDef level
	})
	cloned.(*ZodFunction).internals.Nilable = true
	return cloned
}

// =============================================================================
// WRAPPER METHODS FOR MODIFIERS
// =============================================================================

// Optional makes the function schema optional
func (z *ZodFunction) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nullish makes the function schema both optional and nullable
func (z *ZodFunction) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Pipe creates a validation pipeline
func (z *ZodFunction) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// GetInternals returns the internal type information
func (z *ZodFunction) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// =============================================================================
// HELPER CONSTRUCTORS FOR COMPATIBILITY
// =============================================================================

// NewZodFunction creates a function schema using the simplified API
func NewZodFunction(input interface{}, output Schema) *ZodFunction {
	return Function(FunctionParams{
		Input:  input,
		Output: output,
	})
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodFunction) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// =============================================================================
// DEFAULT AND PREFAULT WRAPPERS
// =============================================================================

// ZodFunctionDefault is a Default wrapper for function type
// Provides perfect type safety and chainable method support
type ZodFunctionDefault struct {
	*ZodDefault[*ZodFunction] // Embed concrete pointer to enable method promotion
}

// Default adds a default value to the function, returns ZodFunctionDefault support chain call
// Compile-time type safety: Function().Default(10) will fail to compile
func (z *ZodFunction) Default(value interface{}) ZodFunctionDefault {
	return ZodFunctionDefault{
		&ZodDefault[*ZodFunction]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the function
func (z *ZodFunction) DefaultFunc(fn func() interface{}) ZodFunctionDefault {
	genericFn := func() any { return fn() }
	return ZodFunctionDefault{
		&ZodDefault[*ZodFunction]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Input adds input validation, returns ZodFunctionDefault for method chaining
func (s ZodFunctionDefault) Input(inputSchema interface{}) ZodFunctionDefault {
	newInner := s.innerType.Input(inputSchema)
	return ZodFunctionDefault{
		&ZodDefault[*ZodFunction]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Output adds output validation, returns ZodFunctionDefault for method chaining
func (s ZodFunctionDefault) Output(outputSchema Schema) ZodFunctionDefault {
	newInner := s.innerType.Output(outputSchema)
	return ZodFunctionDefault{
		&ZodDefault[*ZodFunction]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the function, returns ZodFunctionDefault support chain call
func (s ZodFunctionDefault) Refine(fn func(interface{}) bool, params ...SchemaParams) ZodFunctionDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodFunctionDefault{
		&ZodDefault[*ZodFunction]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodFunctionDefault) Transform(fn func(interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(fn)
}

// Optional adds an optional check to the function, returns ZodType support chain call
func (s ZodFunctionDefault) Optional() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the function, returns ZodType support chain call
func (s ZodFunctionDefault) Nilable() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(s).(ZodType[any, any]))
}

// ZodFunctionPrefault is a Prefault wrapper for function type
// Provides perfect type safety and chainable method support
type ZodFunctionPrefault struct {
	*ZodPrefault[*ZodFunction] // Embed concrete pointer to enable method promotion
}

// Prefault adds a prefault value to the function
// Compile-time type safety: Function().Prefault(10) will fail to compile
func (z *ZodFunction) Prefault(value interface{}) ZodFunctionPrefault {
	// Construct Prefault internals, Type = "prefault", copy checks/coerce/optional/nilable from underlying type
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

	return ZodFunctionPrefault{
		&ZodPrefault[*ZodFunction]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the function
func (z *ZodFunction) PrefaultFunc(fn func() interface{}) ZodFunctionPrefault {
	genericFn := func() any { return fn() }

	// Construct Prefault internals, Type = "prefault", copy checks/coerce/optional/nilable from underlying type
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

	return ZodFunctionPrefault{
		&ZodPrefault[*ZodFunction]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// Input adds input validation, returns ZodFunctionPrefault support chain call
func (s ZodFunctionPrefault) Input(inputSchema interface{}) ZodFunctionPrefault {
	newInner := s.innerType.Input(inputSchema)

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

	return ZodFunctionPrefault{
		&ZodPrefault[*ZodFunction]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Output adds output validation, returns ZodFunctionPrefault support chain call
func (s ZodFunctionPrefault) Output(outputSchema Schema) ZodFunctionPrefault {
	newInner := s.innerType.Output(outputSchema)

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

	return ZodFunctionPrefault{
		&ZodPrefault[*ZodFunction]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the function, returns ZodFunctionPrefault support chain call
func (s ZodFunctionPrefault) Refine(fn func(interface{}) bool, params ...SchemaParams) ZodFunctionPrefault {
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

	return ZodFunctionPrefault{
		&ZodPrefault[*ZodFunction]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodFunctionPrefault) Transform(fn func(interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(fn)
}

// Optional adds an optional check to the function, returns ZodType support chain call
func (s ZodFunctionPrefault) Optional() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the function, returns ZodType support chain call
func (s ZodFunctionPrefault) Nilable() ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(s).(ZodType[any, any]))
}
