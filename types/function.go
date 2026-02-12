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

// Internals returns the internal state of the schema.
func (z *ZodFunction[T]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined or missing values.
func (z *ZodFunction[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodFunction[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodFunction[T]) withCheck(check core.ZodCheck) *ZodFunction[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// newFuncTypeError creates a ZodError for invalid type inputs.
func newFuncTypeError(v any, ctx *core.ParseContext) error {
	raw := issues.NewRawIssue(core.InvalidType, v, issues.WithExpected(string(core.ZodTypeFunction)))
	return issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(raw, ctx, nil)})
}

// convertResult converts the engine result to the constraint type T.
func (z *ZodFunction[T]) convertResult(result any) T {
	typ := reflect.TypeOf((*T)(nil)).Elem()

	if typ.Kind() == reflect.Ptr && typ.Elem() == reflect.TypeOf((*any)(nil)).Elem() {
		if result == nil {
			return any((*any)(nil)).(T)
		}
		ptr := &result
		return any(ptr).(T)
	}

	if result == nil {
		var zero T
		return zero
	}
	return any(result).(T) //nolint:unconvert
}

// Parse validates input and returns a function value.
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
	return z.convertResult(result), nil
}

// MustParse validates input and panics on error.
func (z *ZodFunction[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodFunction[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse validates input with compile-time type safety.
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
	return z.convertResult(result), nil
}

// MustStrictParse validates with compile-time type safety and panics on error.
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

// Optional returns a schema that accepts nil with constraint type *any.
func (z *ZodFunction[T]) Optional() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil with constraint type *any.
func (z *ZodFunction[T]) Nilable() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodFunction[T]) Nullish() *ZodFunction[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value returned when input is nil.
func (z *ZodFunction[T]) Default(v any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory function for the default value.
func (z *ZodFunction[T]) DefaultFunc(fn func() any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(fn)
	return z.withInternals(in)
}

// Prefault sets a prefault value that goes through the full parsing pipeline.
func (z *ZodFunction[T]) Prefault(v any) *ZodFunction[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a factory function for the prefault value.
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
	in := z.internals.Clone()

	meta, ok := core.GlobalRegistry.Get(z)
	if !ok {
		meta = core.GlobalMeta{}
	}
	meta.Description = description

	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, meta)
	return clone
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a transformation pipeline.
func (z *ZodFunction[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrap := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(any(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrap)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodFunction[T]) Overwrite(transform func(T) T, params ...any) *ZodFunction[T] {
	wrap := func(input any) any {
		val, ok := convertToFuncType[T](input)
		if !ok {
			return input
		}
		return transform(val)
	}
	return z.withCheck(checks.NewZodCheckOverwrite(wrap, params...))
}

// Pipe creates a validation pipeline to another schema.
func (z *ZodFunction[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrap := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(any(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrap)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a custom validation function.
func (z *ZodFunction[T]) Refine(fn func(T) bool, params ...any) *ZodFunction[T] {
	wrap := func(v any) bool {
		if val, ok := v.(T); ok {
			return fn(val)
		}
		return false
	}

	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}

	return z.withCheck(checks.NewCustom[any](wrap, msg))
}

// RefineAny adds a custom validation function that accepts any type.
func (z *ZodFunction[T]) RefineAny(fn func(any) bool, params ...any) *ZodFunction[T] {
	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}

	return z.withCheck(checks.NewCustom[any](fn, msg))
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Input sets the input schema for function arguments.
func (z *ZodFunction[T]) Input(schema core.ZodType[any]) *ZodFunction[T] {
	in := z.internals.Clone()
	fi := &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            schema,
		Output:           z.internals.Output,
	}
	return &ZodFunction[T]{internals: fi}
}

// Output sets the output schema for function return values.
func (z *ZodFunction[T]) Output(schema core.ZodType[any]) *ZodFunction[T] {
	in := z.internals.Clone()
	fi := &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            z.internals.Input,
		Output:           schema,
	}
	return &ZodFunction[T]{internals: fi}
}

// Implement wraps a function with input/output validation.
func (z *ZodFunction[T]) Implement(fn any) (any, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return nil, newFuncTypeError(fn, &core.ParseContext{})
	}
	return z.makeValidated(v)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// extractFunction extracts a function from the input value.
func (z *ZodFunction[T]) extractFunction(v any) (any, bool) {
	if v == nil {
		return nil, false
	}
	if reflect.ValueOf(v).Kind() == reflect.Func {
		return v, true
	}
	return nil, false
}

// extractFunctionPtr extracts a pointer to a function from the input value.
func (z *ZodFunction[T]) extractFunctionPtr(v any) (*any, bool) {
	if v == nil {
		return nil, true
	}
	ptr, ok := v.(*any)
	if !ok {
		return nil, false
	}
	if ptr == nil {
		return nil, true
	}
	if val := *ptr; val != nil && reflect.ValueOf(val).Kind() == reflect.Func {
		return ptr, true
	}
	return nil, false
}

// validateFunction validates that the input value is a function.
func (z *ZodFunction[T]) validateFunction(v any, checks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
	if v == nil {
		in := z.Internals()
		if in.Optional || in.Nilable {
			return v, nil
		}
		return nil, newFuncTypeError(v, ctx)
	}

	if reflect.ValueOf(v).Kind() != reflect.Func {
		return nil, newFuncTypeError(v, ctx)
	}

	return engine.ApplyChecks[any](v, checks, ctx)
}

// makeValidated creates a wrapper function with input and output validation.
func (z *ZodFunction[T]) makeValidated(fn reflect.Value) (any, error) {
	wrapper := reflect.MakeFunc(fn.Type(), func(args []reflect.Value) []reflect.Value {
		if z.internals.Input != nil {
			if err := z.validateInput(args); err != nil {
				panic(err)
			}
		}

		results := fn.Call(args)

		if z.internals.Output != nil {
			if err := z.validateOutput(results); err != nil {
				panic(err)
			}
		}

		return results
	})

	return wrapper.Interface(), nil
}

// validateInput validates function arguments against the Input schema.
func (z *ZodFunction[T]) validateInput(args []reflect.Value) error {
	if z.internals.Input == nil {
		return nil
	}

	inputs := make([]any, len(args))
	for i, a := range args {
		inputs[i] = a.Interface()
	}

	_, err := z.internals.Input.Parse(inputs)
	return err
}

// validateOutput validates function return values against the Output schema.
func (z *ZodFunction[T]) validateOutput(results []reflect.Value) error {
	if z.internals.Output == nil {
		return nil
	}

	var out any
	if len(results) == 1 {
		out = results[0].Interface()
	} else {
		vals := make([]any, len(results))
		for i, r := range results {
			vals[i] = r.Interface()
		}
		out = vals
	}

	_, err := z.internals.Output.Parse(out)
	return err
}

// withPtrInternals creates a new instance with pointer constraint type *any.
func (z *ZodFunction[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodFunction[*any] {
	return &ZodFunction[*any]{internals: &ZodFunctionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Input:            z.internals.Input,
		Output:           z.internals.Output,
	}}
}

// withInternals creates a new instance preserving generic type T.
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

// convertToFuncType converts any value to the function constraint type T.
func convertToFuncType[T FunctionConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		typ := reflect.TypeOf((*T)(nil)).Elem()
		return zero, typ.Kind() == reflect.Ptr
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Func {
		return zero, false
	}

	if val, ok := any(v).(T); ok { //nolint:unconvert
		return val, true
	}

	typ := reflect.TypeOf((*T)(nil)).Elem()
	if typ.Kind() == reflect.Ptr {
		if val, ok := any(&v).(T); ok {
			return val, true
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
	sp := utils.NormalizeParams(params...)

	var input, output core.ZodType[any]
	for _, p := range params {
		fp, ok := p.(FunctionParams)
		if !ok {
			continue
		}
		input = fp.Input
		output = fp.Output
		break
	}

	def := &ZodFunctionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeFunction,
			Checks: []core.ZodCheck{},
		},
		Input:  input,
		Output: output,
	}

	utils.ApplySchemaParams(&def.ZodTypeDef, sp)

	return newFuncFromDef[T](def)
}

// =============================================================================
// INTERNAL CONSTRUCTORS
// =============================================================================

// newFuncFromDef constructs a new ZodFunction from the given definition.
func newFuncFromDef[T FunctionConstraint](def *ZodFunctionDef) *ZodFunction[T] {
	in := &ZodFunctionInternals{
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

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		fd := &ZodFunctionDef{
			ZodTypeDef: *d,
			Input:      def.Input,
			Output:     def.Output,
		}
		return any(newFuncFromDef[T](fd)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodFunction[T]{internals: in}
}
