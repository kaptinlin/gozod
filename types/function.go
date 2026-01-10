package types

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// FunctionConstraint restricts values to function or *function.
type FunctionConstraint interface {
	any | *any
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodFunctionDef defines the configuration for function validation
type ZodFunctionDef struct {
	core.ZodTypeDef
	Input  core.ZodType[any] // Schema for validating input arguments
	Output core.ZodType[any] // Schema for validating output result
}

// ZodFunctionInternals contains function validator internal state
type ZodFunctionInternals struct {
	core.ZodTypeInternals
	Def    *ZodFunctionDef   // Schema definition
	Input  core.ZodType[any] // Input validation schema
	Output core.ZodType[any] // Output validation schema
}

// ZodFunction represents a function validation schema with type safety
type ZodFunction[T FunctionConstraint] struct {
	internals *ZodFunctionInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodFunction[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodFunction[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodFunction[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates and returns a function that performs input/output validation
func (z *ZodFunction[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	result, err := engine.ParseComplex[any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeFunction,
		z.extractFunction,
		z.extractFunctionPtr,
		z.validateFunction,
		ctx...,
	)
	if err != nil {
		var zero T
		return zero, err
	}

	// Handle both direct function and pointer to function based on constraint type T
	var zero T

	// Use reflection to determine the actual constraint type
	zeroType := reflect.TypeOf((*T)(nil)).Elem()

	// Check if T is *any (pointer type)
	if zeroType.Kind() == reflect.Ptr && zeroType.Elem() == reflect.TypeOf((*any)(nil)).Elem() {
		// T is *any - return pointer to function
		if result == nil {
			return any((*any)(nil)).(T), nil
		}
		// Create pointer to the function
		fnPtr := &result
		return any(fnPtr).(T), nil
	} else {
		// T is any - return the function directly
		if result == nil {
			return zero, nil
		}
		return any(result).(T), nil //nolint:unconvert
	}
}

// MustParse is the type-safe variant that panics on error
func (z *ZodFunction[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodFunction[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse validates the input using strict parsing rules
func (z *ZodFunction[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	result, err := engine.ParseComplexStrict[any](
		any(input),
		&z.internals.ZodTypeInternals,
		core.ZodTypeFunction,
		z.extractFunction,
		z.extractFunctionPtr,
		z.validateFunction,
		ctx...,
	)
	if err != nil {
		var zero T
		return zero, err
	}

	// Handle both direct function and pointer to function based on constraint type T
	var zero T

	// Use reflection to determine the actual constraint type
	zeroType := reflect.TypeOf((*T)(nil)).Elem()

	// Check if T is *any (pointer type)
	if zeroType.Kind() == reflect.Ptr && zeroType.Elem() == reflect.TypeOf((*any)(nil)).Elem() {
		// T is *any - return pointer to function
		if result == nil {
			return any((*any)(nil)).(T), nil
		}
		// Create pointer to the function
		fnPtr := &result
		return any(fnPtr).(T), nil
	} else {
		// T is any - return the function directly
		if result == nil {
			return zero, nil
		}
		return any(result).(T), nil //nolint:unconvert
	}
}

// MustStrictParse validates the input using strict parsing rules and panics on error
func (z *ZodFunction[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional allows the function to be nil
func (z *ZodFunction[T]) Optional() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows the function to be nil
func (z *ZodFunction[T]) Nilable() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodFunction[T]) Nullish() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves the current generic type T
func (z *ZodFunction[T]) Default(v any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves the current generic type T
func (z *ZodFunction[T]) DefaultFunc(fn func() any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(fn)
	return z.withInternals(in)
}

// Prefault preserves the current generic type T
func (z *ZodFunction[T]) Prefault(v any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc preserves the current generic type T
func (z *ZodFunction[T]) PrefaultFunc(fn func() any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(fn)
	return z.withInternals(in)
}

// Meta stores metadata for this function schema.
func (z *ZodFunction[T]) Meta(meta core.GlobalMeta) *ZodFunction[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodFunction[T]) Describe(description string) *ZodFunction[T] {
	newInternals := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := z.withInternals(newInternals)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodFunction[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(any(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodFunction[T]) Overwrite(transform func(T) T, params ...any) *ZodFunction[T] {
	// Create a transformation function that works with the exact type T
	transformAny := func(input any) any {
		// Try to convert input to type T
		converted, ok := convertToFunctionType[T](input)
		if !ok {
			// If conversion fails, return original value
			return input
		}

		// Apply transformation directly on type T
		return transform(converted)
	}

	check := checks.NewZodCheckOverwrite(transformAny, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodFunction[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(any(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies custom validation function
func (z *ZodFunction[T]) Refine(fn func(T) bool, params ...any) *ZodFunction[T] {
	wrapper := func(v any) bool {
		if typedVal, ok := v.(T); ok {
			return fn(typedVal)
		}
		return false
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodFunction[T]) RefineAny(fn func(any) bool, params ...any) *ZodFunction[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Input sets the input schema for function arguments
func (z *ZodFunction[T]) Input(inputSchema core.ZodType[any]) *ZodFunction[T] {
	in := z.internals.Clone()
	newInternals := &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            inputSchema,
		Output:           z.internals.Output,
	}
	return &ZodFunction[T]{internals: newInternals}
}

// Output sets the output schema for function return value
func (z *ZodFunction[T]) Output(outputSchema core.ZodType[any]) *ZodFunction[T] {
	in := z.internals.Clone()
	newInternals := &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            z.internals.Input,
		Output:           outputSchema,
	}
	return &ZodFunction[T]{internals: newInternals}
}

// Implement wraps a function with input/output validation
func (z *ZodFunction[T]) Implement(fn any) (any, error) {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		rawIssue := issues.NewRawIssue(core.InvalidType, fn, issues.WithExpected(string(core.ZodTypeFunction)))
		finalIssue := issues.FinalizeIssue(rawIssue, &core.ParseContext{}, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	fnType := fnValue.Type()

	// Create wrapper function that validates input and output
	return z.createValidatedFunction(fnValue, fnType)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// extractFunction extracts a function from input value
func (z *ZodFunction[T]) extractFunction(value any) (any, bool) {
	if value == nil {
		return nil, false
	}

	// Check if it's a function type using reflection
	fnValue := reflect.ValueOf(value)
	if fnValue.Kind() == reflect.Func {
		return value, true
	}

	return nil, false
}

// extractFunctionPtr extracts a pointer to function from input value
func (z *ZodFunction[T]) extractFunctionPtr(value any) (*any, bool) {
	if value == nil {
		return nil, true
	}

	// If it's already a pointer to function
	if ptr, ok := value.(*any); ok {
		if ptr == nil {
			return nil, true
		}
		// Check if the pointed value is a function
		if ptrValue := *ptr; ptrValue != nil {
			fnValue := reflect.ValueOf(ptrValue)
			if fnValue.Kind() == reflect.Func {
				return ptr, true
			}
		}
	}

	// If it's a direct function, we can't extract a pointer from it
	return nil, false
}

// validateFunction validates that input is a function (modified for ParseComplex compatibility)
func (z *ZodFunction[T]) validateFunction(value any, checks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	if value == nil {
		internals := z.GetInternals()
		if internals.Optional || internals.Nilable {
			return value, nil
		}
		rawIssue := issues.NewRawIssue(core.InvalidType, value, issues.WithExpected(string(core.ZodTypeFunction)))
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	fnValue := reflect.ValueOf(value)
	if fnValue.Kind() != reflect.Func {
		rawIssue := issues.NewRawIssue(core.InvalidType, value, issues.WithExpected(string(core.ZodTypeFunction)))
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Validate function signature if input/output schemas are provided
	if z.internals.Input != nil || z.internals.Output != nil {
		if err := z.validateFunctionSignature(value, ctx); err != nil {
			return nil, err
		}
	}

	// Run standard checks validation
	return engine.ApplyChecks[any](value, checks, ctx)
}

// validateFunctionSignature validates the function signature against the schema's constraints
func (z *ZodFunction[T]) validateFunctionSignature(input any, ctx *core.ParseContext) error {
	// Input and output validation is deferred to runtime (Implement() call and function execution)
	// since we need actual arguments and return values to validate against the schemas.
	// Pre-validation is not possible for function types without runtime values.
	_ = ctx // Prevent unused parameter warning
	return nil
}

// createValidatedFunction creates a wrapper function with validation
func (z *ZodFunction[T]) createValidatedFunction(fnValue reflect.Value, fnType reflect.Type) (any, error) {
	// Create a wrapper function that validates inputs and outputs
	wrapperType := fnType
	wrapper := reflect.MakeFunc(wrapperType, func(args []reflect.Value) []reflect.Value {
		// Validate inputs if input schema is provided
		if z.internals.Input != nil {
			if err := z.validateInputArguments(args); err != nil {
				panic(err)
			}
		}

		// Call the original function
		results := fnValue.Call(args)

		// Validate outputs if output schema is provided
		if z.internals.Output != nil {
			if err := z.validateOutputResults(results); err != nil {
				panic(err)
			}
		}

		return results
	})

	return wrapper.Interface(), nil
}

// validateInputArguments validates function input arguments
func (z *ZodFunction[T]) validateInputArguments(args []reflect.Value) error {
	if z.internals.Input == nil {
		return nil
	}

	// Convert reflect.Value slice to []any
	inputs := make([]any, len(args))
	for i, arg := range args {
		inputs[i] = arg.Interface()
	}

	// Validate against input schema
	_, err := z.internals.Input.Parse(inputs)
	return err
}

// validateOutputResults validates function output results
func (z *ZodFunction[T]) validateOutputResults(results []reflect.Value) error {
	if z.internals.Output == nil {
		return nil
	}

	// Handle single return value or multiple return values
	var output any
	if len(results) == 1 {
		output = results[0].Interface()
	} else {
		// Convert multiple return values to slice
		outputs := make([]any, len(results))
		for i, result := range results {
			outputs[i] = result.Interface()
		}
		output = outputs
	}

	// Validate against output schema
	_, err := z.internals.Output.Parse(output)
	return err
}

// withPtrInternals creates a new ZodFunction instance with pointer type
func (z *ZodFunction[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodFunction[*any] {
	return &ZodFunction[*any]{internals: &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            z.internals.Input,
		Output:           z.internals.Output,
	}}
}

// withInternals creates a new ZodFunction instance preserving generic type T
func (z *ZodFunction[T]) withInternals(in *core.ZodTypeInternals) *ZodFunction[T] {
	return &ZodFunction[T]{internals: &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            z.internals.Input,
		Output:           z.internals.Output,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodFunction[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodFunction[T]); ok {
		// Copy all state from source
		z.internals.ZodTypeInternals = src.internals.ZodTypeInternals
		z.internals.Def = src.internals.Def
		z.internals.Input = src.internals.Input
		z.internals.Output = src.internals.Output
	}
}

// convertToFunctionType converts any value to the function constraint type T with strict type checking
func convertToFunctionType[T FunctionConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeOf((*T)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Check if input is a function
	if !reflect.ValueOf(v).IsValid() || reflect.ValueOf(v).Kind() != reflect.Func {
		return zero, false // Reject all non-function types
	}

	// Try direct conversion first
	if converted, ok := any(v).(T); ok { //nolint:unconvert
		return converted, true
	}

	// Handle pointer conversion for different types
	zeroType := reflect.TypeOf((*T)(nil)).Elem()
	if zeroType.Kind() == reflect.Ptr {
		// T is *any - return pointer to the function
		if converted, ok := any(&v).(T); ok {
			return converted, true
		}
	}

	return zero, false
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// FunctionParams defines parameters for function schema creation
type FunctionParams struct {
	Input  core.ZodType[any] `json:"input,omitempty"`
	Output core.ZodType[any] `json:"output,omitempty"`
}

// Function creates a function schema
func Function(params ...any) *ZodFunction[any] {
	return FunctionTyped[any](params...)
}

// FunctionPtr creates a schema for *function
func FunctionPtr(params ...any) *ZodFunction[*any] {
	return FunctionTyped[*any](params...)
}

// FunctionTyped is the underlying generic function for creating function schemas
func FunctionTyped[T FunctionConstraint](params ...any) *ZodFunction[T] {
	schemaParams := utils.NormalizeParams(params...)

	var input, output core.ZodType[any]

	// Extract function-specific parameters
	for _, param := range params {
		if fp, ok := param.(FunctionParams); ok {
			input = fp.Input
			output = fp.Output
			break
		}
	}

	def := &ZodFunctionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeFunction,
			Checks: []core.ZodCheck{},
		},
		Input:  input,
		Output: output,
	}

	// Apply the normalized parameters to the schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodFunctionFromDef[T](def)
}

// =============================================================================
// INTERNAL CONSTRUCTORS
// =============================================================================

// newZodFunctionFromDef constructs a new ZodFunction from the given definition
func newZodFunctionFromDef[T FunctionConstraint](def *ZodFunctionDef) *ZodFunction[T] {
	internals := &ZodFunctionInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:    def,
		Input:  def.Input,
		Output: def.Output,
	}

	// Provide a constructor so that AddCheck can create new schema instances
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		functionDef := &ZodFunctionDef{
			ZodTypeDef: *newDef,
			Input:      def.Input,
			Output:     def.Output,
		}
		return any(newZodFunctionFromDef[T](functionDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodFunction[T]{internals: internals}
}
