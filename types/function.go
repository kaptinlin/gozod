package types

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for function validations
var (
	ErrExpectedFunction           = errors.New("expected function type")
	ErrFunctionSignatureMismatch  = errors.New("function signature mismatch")
	ErrFunctionParameterMismatch  = errors.New("function parameter type mismatch")
	ErrFunctionReturnTypeMismatch = errors.New("function return type mismatch")
)

// =============================================================================
// FUNCTION TYPE DEFINITION
// =============================================================================

// ZodFunctionDef defines the configuration for function validation
type ZodFunctionDef struct {
	core.ZodTypeDef
	Input  core.ZodType[any, any] // Schema for validating input arguments
	Output core.ZodType[any, any] // Schema for validating output result
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
	core.ZodTypeInternals
	Def *ZodFunctionDef // Schema definition
	Bag map[string]any  // Additional metadata
}

// createZodFunctionFromDef creates a ZodFunction from a definition
func createZodFunctionFromDef(def *ZodFunctionDef) *ZodFunction {
	internals := &ZodFunctionInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:     "function",
			Checks:   make([]core.ZodCheck, 0),
			Nilable:  false,
			Optional: false,
			Values:   make(map[any]struct{}),
		},
		Def: def,
		Bag: make(map[string]any),
	}

	// Set up constructor for cloning support on the embedded ZodTypeInternals
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		functionDef := &ZodFunctionDef{
			ZodTypeDef: *newDef,
			Input:      def.Input,  // Preserve the original input schema
			Output:     def.Output, // Preserve the original output schema
		}
		return createZodFunctionFromDef(functionDef)
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		return parseZodFunction(payload, def, &internals.ZodTypeInternals, ctx)
	}

	return &ZodFunction{
		internals: internals,
		def:       def,
	}
}

// FunctionParams represents configuration for function validation
type FunctionParams struct {
	Description string                 // Optional description
	Input       any                    // Can be Schema or []Schema
	Output      core.ZodType[any, any] // Return value schema
}

// Function creates a new function schema with input and output validation
func Function(params ...FunctionParams) *ZodFunction {
	var input, output core.ZodType[any, any]

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
		ZodTypeDef: core.ZodTypeDef{Type: "function"},
		Input:      input,
		Output:     output,
	}

	return createZodFunctionFromDef(def)
}

// normalizeInputSchema converts various input formats to a unified Schema
func normalizeInputSchema(input any) core.ZodType[any, any] {
	switch inputVal := input.(type) {
	case []core.ZodType[any, any]:
		// Array of schemas: create tuple-like validation
		inputArgs := make([]any, len(inputVal))
		for i, schema := range inputVal {
			inputArgs[i] = schema
		}
		return Array(inputArgs...)
	case core.ZodType[any, any]:
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
func (z *ZodFunction) Input(inputSchema any) *ZodFunction {
	input := normalizeInputSchema(inputSchema)

	newDef := &ZodFunctionDef{
		ZodTypeDef: z.def.ZodTypeDef,
		Input:      input,
		Output:     z.def.Output,
	}

	return createZodFunctionFromDef(newDef)
}

// Output creates a new function schema with specified output validation
func (z *ZodFunction) Output(outputSchema core.ZodType[any, any]) *ZodFunction {
	newDef := &ZodFunctionDef{
		ZodTypeDef: z.def.ZodTypeDef,
		Input:      z.def.Input,
		Output:     outputSchema,
	}

	return createZodFunctionFromDef(newDef)
}

// Implement validates a function and returns a wrapped version with validation
func (z *ZodFunction) Implement(fn any) (any, error) {
	if fn == nil {
		return nil, fmt.Errorf("function cannot be nil")
	}

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		receivedType := string(reflectx.ParsedType(fn))
		return nil, fmt.Errorf("expected function, got %s", receivedType)
	}

	fnType := fnValue.Type()

	// Validate function signature against input schema (parameter count only)
	if z.def.Input != nil {
		if err := z.validateFunctionSignature(fnType); err != nil {
			return nil, err
		}
	}

	// Create a new function that validates inputs and outputs
	wrappedFn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		// Validate input arguments if input schema is defined
		if z.def.Input != nil {
			if inputIssues := z.validateInputArguments(args); len(inputIssues) > 0 {
				zodErr := issues.NewZodError(issues.ConvertRawIssuesToIssues(inputIssues, nil))
				panic(zodErr)
			}
		}

		// Call the original function
		results := fnValue.Call(args)

		// Validate output if output schema is defined and there are results
		if z.def.Output != nil && len(results) > 0 {
			if outputIssues := z.validateOutputResults(results); len(outputIssues) > 0 {
				zodErr := issues.NewZodError(issues.ConvertRawIssuesToIssues(outputIssues, nil))
				panic(zodErr)
			}
		}

		return results
	})

	return wrappedFn.Interface(), nil
}

// validateFunctionSignature validates that the function signature matches the input schema
func (z *ZodFunction) validateFunctionSignature(fnType reflect.Type) error {
	// Only ensure the number of parameters matches the input schema definition.

	// Case 1: the input schema is an Array -> multiple positional parameters expected.
	if arraySchema, ok := z.def.Input.(*ZodArray); ok {
		expectedParamCount := len(arraySchema.internals.Items)
		if fnType.NumIn() != expectedParamCount {
			return fmt.Errorf("function signature mismatch: expected %d parameters, got %d", expectedParamCount, fnType.NumIn())
		}
		return nil
	}

	// Case 2: single parameter schema -> exactly one parameter expected.
	if fnType.NumIn() != 1 {
		return fmt.Errorf("%w: expected 1 parameter, got %d", ErrFunctionSignatureMismatch, fnType.NumIn())
	}

	// We intentionally DO NOT enforce parameter type matching here to stay consistent with
	// the relaxed behaviour of the reference implementation. Input type correctness will
	// be ensured at invocation time via schema parsing.
	return nil
}

// validateInputArguments validates function input arguments against the input schema
func (z *ZodFunction) validateInputArguments(args []reflect.Value) []core.ZodRawIssue {
	// Convert args to any slice for validation
	inputArgs := make([]any, len(args))
	for i, arg := range args {
		inputArgs[i] = arg.Interface()
	}

	var validationIssues []core.ZodRawIssue

	// Check if input schema is an Array and extract element schemas
	if arraySchema, ok := z.def.Input.(*ZodArray); ok {
		// Array syntax case: [z.string()] or [z.string(), z.number()]
		elementSchemas := arraySchema.internals.Items
		if len(elementSchemas) != len(inputArgs) {
			issue := issues.CreateCustomIssue(
				fmt.Sprintf("Expected %d arguments, got %d", len(elementSchemas), len(inputArgs)),
				map[string]any{"expected": len(elementSchemas), "received": len(inputArgs)},
				inputArgs,
			)
			validationIssues = append(validationIssues, issue)
			return validationIssues
		}

		// Validate each argument individually against its schema
		for i, arg := range inputArgs {
			payload := &core.ParsePayload{
				Value:  arg,
				Path:   []any{fmt.Sprintf("arg[%d]", i)},
				Issues: make([]core.ZodRawIssue, 0),
			}

			elementSchemas[i].GetInternals().Parse(payload, nil)
			if len(payload.Issues) > 0 {
				validationIssues = append(validationIssues, payload.Issues...)
			}
		}
		return validationIssues
	}

	// Direct schema case: single parameter or custom schema
	if len(inputArgs) == 1 {
		payload := &core.ParsePayload{
			Value:  inputArgs[0],
			Path:   []any{"input"},
			Issues: make([]core.ZodRawIssue, 0),
		}

		z.def.Input.GetInternals().Parse(payload, nil)
		return payload.Issues
	}

	// For multiple parameters with direct schema, validate as array
	payload := &core.ParsePayload{
		Value:  inputArgs,
		Path:   []any{"input"},
		Issues: make([]core.ZodRawIssue, 0),
	}

	z.def.Input.GetInternals().Parse(payload, nil)
	return payload.Issues
}

// validateOutputResults validates function output results against the output schema
func (z *ZodFunction) validateOutputResults(results []reflect.Value) []core.ZodRawIssue {
	// For functions with single return value, validate directly
	if len(results) == 1 {
		payload := &core.ParsePayload{
			Value:  results[0].Interface(),
			Path:   []any{"output"},
			Issues: make([]core.ZodRawIssue, 0),
		}

		z.def.Output.GetInternals().Parse(payload, nil)
		return payload.Issues
	}

	// For functions with multiple return values, validate as slice
	outputSlice := make([]any, len(results))
	for i, result := range results {
		outputSlice[i] = result.Interface()
	}

	payload := &core.ParsePayload{
		Value:  outputSlice,
		Path:   []any{"output"},
		Issues: make([]core.ZodRawIssue, 0),
	}

	z.def.Output.GetInternals().Parse(payload, nil)
	return payload.Issues
}

// =============================================================================
// CORE PARSE LOGIC WITH SMART TYPE INFERENCE
// =============================================================================

// parseZodFunction implements the core parsing logic for function type
func parseZodFunction(payload *core.ParsePayload, def *ZodFunctionDef, internals *core.ZodTypeInternals, ctx *core.ParseContext) *core.ParsePayload {
	input := payload.Value

	// 1. Unified nil handling
	if input == nil {
		if !internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("function", input)
			payload.Issues = append(payload.Issues, rawIssue)
			return payload
		}
		payload.Value = (*any)(nil) // Return typed nil pointer
		return payload
	}

	// 2. Smart type inference: preserve exact function type
	switch v := input.(type) {
	case func():
		// No-argument, no-return function
		if len(internals.Checks) > 0 {
			engine.RunChecksOnValue(v, internals.Checks, payload, ctx)
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
				engine.RunChecksOnValue(input, internals.Checks, payload, ctx)
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
					engine.RunChecksOnValue(input, internals.Checks, payload, ctx)
					if len(payload.Issues) > 0 {
						return payload
					}
				}
				payload.Value = input // Keep original pointer to function
				return payload
			}
		}

		// 3. No coercion for function type (non-primitive)

		// 4. Unified error creation
		rawIssue := issues.CreateInvalidTypeIssue("function", input)
		payload.Issues = append(payload.Issues, rawIssue)
		return payload
	}
}

// =============================================================================
// HELPER FUNCTIONS FOR TYPE EXTRACTION
// =============================================================================

// extractFunctionValue extracts function value, handling various input types
func extractFunctionValue(input any) (any, bool, error) {
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
func (z *ZodFunction) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}
	// 1. Unified nil handling
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue("function", input)
			finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		return (*any)(nil), nil
	}

	// 2. Smart type inference: function → function, *function → *function
	switch v := input.(type) {
	case func():
		// No-argument, no-return function
		if len(z.internals.Checks) > 0 {
			payload := &core.ParsePayload{Value: v, Issues: make([]core.ZodRawIssue, 0)}
			engine.RunChecksOnValue(v, z.internals.Checks, payload, parseCtx)
			if len(payload.Issues) > 0 {
				return nil, &core.ZodError{Issues: payload.Issues}
			}
		}
		return v, nil

	default:
		// Check if it's any kind of function using reflection
		fnValue := reflect.ValueOf(input)
		if fnValue.Kind() == reflect.Func {
			// Any function type - preserve original type
			if len(z.internals.Checks) > 0 {
				payload := &core.ParsePayload{Value: input, Issues: make([]core.ZodRawIssue, 0)}
				engine.RunChecksOnValue(input, z.internals.Checks, payload, parseCtx)
				if len(payload.Issues) > 0 {
					return nil, &core.ZodError{Issues: payload.Issues}
				}
			}
			return input, nil
		}

		// Handle pointer to function
		if fnValue.Kind() == reflect.Ptr && !fnValue.IsNil() {
			if fnValue.Elem().Kind() == reflect.Func {
				// Pointer to function - preserve pointer type
				if len(z.internals.Checks) > 0 {
					payload := &core.ParsePayload{Value: input, Issues: make([]core.ZodRawIssue, 0)}
					engine.RunChecksOnValue(input, z.internals.Checks, payload, parseCtx)
					if len(payload.Issues) > 0 {
						return nil, &core.ZodError{Issues: payload.Issues}
					}
				}
				return input, nil
			}
		}

		// 3. No coercion for function type (non-primitive)

		// 4. Unified error creation
		rawIssue := issues.CreateInvalidTypeIssue("function", input)
		finalIssue := issues.FinalizeIssue(rawIssue, nil, core.GetConfig())
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}
}

// MustParse parses the input and panics on failure
func (z *ZodFunction) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodFunction) Refine(fn func(any) bool, params ...any) *ZodFunction {
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
func (z *ZodFunction) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// =============================================================================
// TRANSFORM METHODS
// =============================================================================

// Transform creates a transformation pipeline
func (z *ZodFunction) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(fn)
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodFunction) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// =============================================================================
// NILABLE MODIFIER WITH CLONE PATTERN
// =============================================================================

// Nilable creates a new function schema that accepts nil values
func (z *ZodFunction) Nilable() core.ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodFunction) setNilable() core.ZodType[any, any] {
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// Clone operates on ZodTypeDef level
	})
	cloned.(*ZodFunction).internals.Nilable = true
	return cloned
}

// =============================================================================
// WRAPPER METHODS FOR MODIFIERS
// =============================================================================

// Optional makes the function schema optional
func (z *ZodFunction) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nullish makes the function schema both optional and nullable
func (z *ZodFunction) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Pipe creates a validation pipeline
func (z *ZodFunction) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// GetInternals returns the internal type information
func (z *ZodFunction) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodFunction) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
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
func (z *ZodFunction) Default(value any) ZodFunctionDefault {
	return ZodFunctionDefault{
		&ZodDefault[*ZodFunction]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the function
func (z *ZodFunction) DefaultFunc(fn func() any) ZodFunctionDefault {
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
func (s ZodFunctionDefault) Input(inputSchema any) ZodFunctionDefault {
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
func (s ZodFunctionDefault) Output(outputSchema core.ZodType[any, any]) ZodFunctionDefault {
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
func (s ZodFunctionDefault) Refine(fn func(any) bool, params ...any) ZodFunctionDefault {
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
func (s ZodFunctionDefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(fn)
}

// Optional adds an optional check to the function, returns ZodType support chain call
func (s ZodFunctionDefault) Optional() core.ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the function, returns ZodType support chain call
func (s ZodFunctionDefault) Nilable() core.ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(s).(core.ZodType[any, any]))
}

// ZodFunctionPrefault is a Prefault wrapper for function type
// Provides perfect type safety and chainable method support
type ZodFunctionPrefault struct {
	*ZodPrefault[*ZodFunction] // Embed concrete pointer to enable method promotion
}

// Prefault adds a prefault value to the function
// Compile-time type safety: Function().Prefault(10) will fail to compile
func (z *ZodFunction) Prefault(value any) ZodFunctionPrefault {
	// Construct Prefault internals, Type = "prefault", copy checks/coerce/optional/nilable from underlying type
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
func (z *ZodFunction) PrefaultFunc(fn func() any) ZodFunctionPrefault {
	genericFn := func() any { return fn() }

	// Construct Prefault internals, Type = "prefault", copy checks/coerce/optional/nilable from underlying type
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
func (s ZodFunctionPrefault) Input(inputSchema any) ZodFunctionPrefault {
	newInner := s.innerType.Input(inputSchema)

	// Construct new internals
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
func (s ZodFunctionPrefault) Output(outputSchema core.ZodType[any, any]) ZodFunctionPrefault {
	newInner := s.innerType.Output(outputSchema)

	// Construct new internals
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
func (s ZodFunctionPrefault) Refine(fn func(any) bool, params ...any) ZodFunctionPrefault {
	newInner := s.innerType.Refine(fn, params...)

	// Construct new internals
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
func (s ZodFunctionPrefault) Transform(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(fn)
}

// Optional adds an optional check to the function, returns ZodType support chain call
func (s ZodFunctionPrefault) Optional() core.ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the function, returns ZodType support chain call
func (s ZodFunctionPrefault) Nilable() core.ZodType[any, any] {
	// Wrap current wrapper instance instead of underlying type
	return Nilable(any(s).(core.ZodType[any, any]))
}
