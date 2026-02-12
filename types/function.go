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

// ZodFunctionDef defines the configuration for function validation.
type ZodFunctionDef struct {
	core.ZodTypeDef
	Input  core.ZodType[any]
	Output core.ZodType[any]
}

// ZodFunctionInternals contains function validator internal state.
type ZodFunctionInternals struct {
	core.ZodTypeInternals
	Def    *ZodFunctionDef
	Input  core.ZodType[any]
	Output core.ZodType[any]
}

// ZodFunction represents a function validation schema with type safety.
type ZodFunction[T FunctionConstraint] struct {
	internals *ZodFunctionInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema.
func (z *ZodFunction[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodFunction[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodFunction[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// withCheck clones internals, adds a check, and returns a new instance.
// This eliminates the repeated clone→addCheck→withInternals pattern.
func (z *ZodFunction[T]) withCheck(check core.ZodCheck) *ZodFunction[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// newInvalidTypeError creates a ZodError for invalid type inputs.
func newInvalidTypeError(value any, ctx *core.ParseContext) error {
	rawIssue := issues.NewRawIssue(core.InvalidType, value, issues.WithExpected(string(core.ZodTypeFunction)))
	finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
	return issues.NewZodError([]core.ZodIssue{finalIssue})
}

// convertFunctionResult converts the engine result to the constraint type T.
// Handles both direct function (any) and pointer to function (*any) based on T.
func (z *ZodFunction[T]) convertFunctionResult(result any) T {
	var zero T
	zeroType := reflect.TypeOf((*T)(nil)).Elem()

	// Check if T is *any (pointer type)
	if zeroType.Kind() == reflect.Ptr && zeroType.Elem() == reflect.TypeOf((*any)(nil)).Elem() {
		// T is *any - return pointer to function
		if result == nil {
			return any((*any)(nil)).(T)
		}
		fnPtr := &result
		return any(fnPtr).(T)
	}

	// T is any - return the function directly
	if result == nil {
		return zero
	}
	return any(result).(T) //nolint:unconvert
}

// Parse validates input and returns a function value matching the constraint type T.
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
	return z.convertFunctionResult(result), nil
}

// MustParse validates input and panics on error.
func (z *ZodFunction[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (z *ZodFunction[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse provides compile-time type safety by requiring exact type matching.
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
	return z.convertFunctionResult(result), nil
}

// MustStrictParse validates input with strict type matching and panics on error.
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

// Optional returns a schema that accepts the function type or nil, with constraint type *any.
func (z *ZodFunction[T]) Optional() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts the function type or nil, with constraint type *any.
func (z *ZodFunction[T]) Nilable() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility.
func (z *ZodFunction[T]) Nullish() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value returned when input is nil, bypassing validation.
func (z *ZodFunction[T]) Default(v any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory function that provides the default value when input is nil.
func (z *ZodFunction[T]) DefaultFunc(fn func() any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(fn)
	return z.withInternals(in)
}

// Prefault sets a prefault value that goes through the full parsing pipeline when input is nil.
func (z *ZodFunction[T]) Prefault(v any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a factory function that provides the prefault value through the full parsing pipeline.
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

// Transform creates a type-safe transformation pipeline.
func (z *ZodFunction[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(any(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this does not change the inferred type.
func (z *ZodFunction[T]) Overwrite(transform func(T) T, params ...any) *ZodFunction[T] {
	transformAny := func(input any) any {
		converted, ok := convertToFunctionType[T](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	return z.withCheck(checks.NewZodCheckOverwrite(transformAny, params...))
}

// Pipe creates a validation pipeline to another schema.
func (z *ZodFunction[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(any(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a custom validation function.
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
		errorMessage = schemaParams.Error
	}

	return z.withCheck(checks.NewCustom[any](wrapper, errorMessage))
}

// RefineAny adds a custom validation function that accepts any type.
func (z *ZodFunction[T]) RefineAny(fn func(any) bool, params ...any) *ZodFunction[T] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	return z.withCheck(checks.NewCustom[any](fn, errorMessage))
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Input sets the input schema for validating function arguments.
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

// Output sets the output schema for validating function return values.
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

// Implement wraps a function with input/output validation.
//
// It returns a new function with the same signature that validates
// arguments and return values against the configured Input/Output schemas.
func (z *ZodFunction[T]) Implement(fn any) (any, error) {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return nil, newInvalidTypeError(fn, &core.ParseContext{})
	}

	return z.createValidatedFunction(fnValue)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// extractFunction extracts a function from input value.
func (z *ZodFunction[T]) extractFunction(value any) (any, bool) {
	if value == nil {
		return nil, false
	}
	if reflect.ValueOf(value).Kind() == reflect.Func {
		return value, true
	}
	return nil, false
}

// extractFunctionPtr extracts a pointer to function from input value.
func (z *ZodFunction[T]) extractFunctionPtr(value any) (*any, bool) {
	if value == nil {
		return nil, true
	}
	if ptr, ok := value.(*any); ok {
		if ptr == nil {
			return nil, true
		}
		if ptrValue := *ptr; ptrValue != nil {
			if reflect.ValueOf(ptrValue).Kind() == reflect.Func {
				return ptr, true
			}
		}
	}
	return nil, false
}

// validateFunction validates that input is a function.
func (z *ZodFunction[T]) validateFunction(value any, checks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	if value == nil {
		internals := z.GetInternals()
		if internals.Optional || internals.Nilable {
			return value, nil
		}
		return nil, newInvalidTypeError(value, ctx)
	}

	if reflect.ValueOf(value).Kind() != reflect.Func {
		return nil, newInvalidTypeError(value, ctx)
	}

	return engine.ApplyChecks[any](value, checks, ctx)
}

// createValidatedFunction creates a wrapper function with input/output validation.
func (z *ZodFunction[T]) createValidatedFunction(fnValue reflect.Value) (any, error) {
	wrapper := reflect.MakeFunc(fnValue.Type(), func(args []reflect.Value) []reflect.Value {
		if z.internals.Input != nil {
			if err := z.validateInputArguments(args); err != nil {
				panic(err)
			}
		}

		results := fnValue.Call(args)

		if z.internals.Output != nil {
			if err := z.validateOutputResults(results); err != nil {
				panic(err)
			}
		}

		return results
	})

	return wrapper.Interface(), nil
}

// validateInputArguments validates function input arguments against the Input schema.
func (z *ZodFunction[T]) validateInputArguments(args []reflect.Value) error {
	if z.internals.Input == nil {
		return nil
	}

	inputs := make([]any, len(args))
	for i, arg := range args {
		inputs[i] = arg.Interface()
	}

	_, err := z.internals.Input.Parse(inputs)
	return err
}

// validateOutputResults validates function output results against the Output schema.
func (z *ZodFunction[T]) validateOutputResults(results []reflect.Value) error {
	if z.internals.Output == nil {
		return nil
	}

	var output any
	if len(results) == 1 {
		output = results[0].Interface()
	} else {
		outputs := make([]any, len(results))
		for i, result := range results {
			outputs[i] = result.Interface()
		}
		output = outputs
	}

	_, err := z.internals.Output.Parse(output)
	return err
}

// withPtrInternals creates a new ZodFunction instance with pointer constraint type *any.
func (z *ZodFunction[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodFunction[*any] {
	return &ZodFunction[*any]{internals: &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            z.internals.Input,
		Output:           z.internals.Output,
	}}
}

// withInternals creates a new ZodFunction instance preserving generic type T.
func (z *ZodFunction[T]) withInternals(in *core.ZodTypeInternals) *ZodFunction[T] {
	return &ZodFunction[T]{internals: &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            z.internals.Input,
		Output:           z.internals.Output,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodFunction[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodFunction[T]); ok {
		z.internals.ZodTypeInternals = src.internals.ZodTypeInternals
		z.internals.Def = src.internals.Def
		z.internals.Input = src.internals.Input
		z.internals.Output = src.internals.Output
	}
}

// convertToFunctionType converts any value to the function constraint type T.
func convertToFunctionType[T FunctionConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		zeroType := reflect.TypeOf((*T)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true
		}
		return zero, false
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Func {
		return zero, false
	}

	if converted, ok := any(v).(T); ok { //nolint:unconvert
		return converted, true
	}

	zeroType := reflect.TypeOf((*T)(nil)).Elem()
	if zeroType.Kind() == reflect.Ptr {
		if converted, ok := any(&v).(T); ok {
			return converted, true
		}
	}

	return zero, false
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// FunctionParams defines parameters for function schema creation.
type FunctionParams struct {
	Input  core.ZodType[any] `json:"input,omitempty"`
	Output core.ZodType[any] `json:"output,omitempty"`
}

// Function creates a function validation schema with constraint type any.
func Function(params ...any) *ZodFunction[any] {
	return FunctionTyped[any](params...)
}

// FunctionPtr creates a function validation schema with constraint type *any.
func FunctionPtr(params ...any) *ZodFunction[*any] {
	return FunctionTyped[*any](params...)
}

// FunctionTyped creates a function validation schema with the specified constraint type.
func FunctionTyped[T FunctionConstraint](params ...any) *ZodFunction[T] {
	schemaParams := utils.NormalizeParams(params...)

	var input, output core.ZodType[any]
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

	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodFunctionFromDef[T](def)
}

// =============================================================================
// INTERNAL CONSTRUCTORS
// =============================================================================

// newZodFunctionFromDef constructs a new ZodFunction from the given definition.
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
